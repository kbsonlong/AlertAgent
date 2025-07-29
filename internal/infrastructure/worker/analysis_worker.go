package worker

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"alert_agent/internal/domain/analysis"
	"alert_agent/internal/pkg/logger"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// AnalysisWorkerImpl 分析工作器实现
type AnalysisWorkerImpl struct {
	id              string
	status          atomic.Value // string: running, stopped, error
	currentTask     atomic.Value // string: 当前任务ID
	processedCount  int64
	errorCount      int64
	startTime       time.Time
	lastActiveTime  atomic.Value // time.Time
	metadata        map[string]interface{}

	taskQueue       analysis.AnalysisTaskQueue
	taskRepo        analysis.AnalysisTaskRepository
	resultRepo      analysis.AnalysisResultRepository
	progressTracker analysis.AnalysisProgressTracker
	analysisEngine  analysis.AnalysisEngine
	metricsCollector analysis.AnalysisMetricsCollector

	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
	logger *zap.Logger
}

// NewAnalysisWorker 创建分析工作器
func NewAnalysisWorker(
	taskQueue analysis.AnalysisTaskQueue,
	taskRepo analysis.AnalysisTaskRepository,
	resultRepo analysis.AnalysisResultRepository,
	progressTracker analysis.AnalysisProgressTracker,
	analysisEngine analysis.AnalysisEngine,
	metricsCollector analysis.AnalysisMetricsCollector,
) analysis.AnalysisWorker {
	workerID := uuid.New().String()
	worker := &AnalysisWorkerImpl{
		id:              workerID,
		startTime:       time.Now(),
		metadata:        make(map[string]interface{}),
		taskQueue:       taskQueue,
		taskRepo:        taskRepo,
		resultRepo:      resultRepo,
		progressTracker: progressTracker,
		analysisEngine:  analysisEngine,
		metricsCollector: metricsCollector,
		logger:          logger.L.Named(fmt.Sprintf("analysis-worker-%s", workerID[:8])),
	}

	worker.status.Store("stopped")
	worker.currentTask.Store("")
	worker.lastActiveTime.Store(time.Now())

	return worker
}

// Start 启动工作器
func (w *AnalysisWorkerImpl) Start(ctx context.Context) error {
	if w.status.Load().(string) == "running" {
		return fmt.Errorf("worker %s is already running", w.id)
	}

	w.ctx, w.cancel = context.WithCancel(ctx)
	w.status.Store("running")
	w.lastActiveTime.Store(time.Now())

	w.wg.Add(1)
	go w.workerLoop()

	w.logger.Info("Analysis worker started", zap.String("worker_id", w.id))
	return nil
}

// Stop 停止工作器
func (w *AnalysisWorkerImpl) Stop(ctx context.Context) error {
	if w.status.Load().(string) != "running" {
		return fmt.Errorf("worker %s is not running", w.id)
	}

	w.status.Store("stopping")
	w.cancel()

	// 等待工作器停止，带超时
	done := make(chan struct{})
	go func() {
		w.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		w.status.Store("stopped")
		w.logger.Info("Analysis worker stopped", zap.String("worker_id", w.id))
		return nil
	case <-time.After(30 * time.Second):
		w.status.Store("error")
		w.logger.Error("Worker stop timeout", zap.String("worker_id", w.id))
		return fmt.Errorf("worker stop timeout")
	case <-ctx.Done():
		w.status.Store("error")
		w.logger.Error("Worker stop cancelled", zap.String("worker_id", w.id))
		return ctx.Err()
	}
}

// GetStatus 获取工作器状态
func (w *AnalysisWorkerImpl) GetStatus() *analysis.WorkerStatus {
	return &analysis.WorkerStatus{
		ID:              w.id,
		Status:          w.status.Load().(string),
		CurrentTask:     w.currentTask.Load().(string),
		ProcessedCount:  atomic.LoadInt64(&w.processedCount),
		ErrorCount:      atomic.LoadInt64(&w.errorCount),
		LastActiveTime:  w.lastActiveTime.Load().(time.Time),
		StartTime:       w.startTime,
		Metadata:        w.metadata,
	}
}

// ProcessTask 处理单个任务
func (w *AnalysisWorkerImpl) ProcessTask(ctx context.Context, task *analysis.AnalysisTask) (*analysis.AnalysisResult, error) {
	if task == nil {
		return nil, fmt.Errorf("task cannot be nil")
	}

	w.currentTask.Store(task.ID)
	w.lastActiveTime.Store(time.Now())

	defer func() {
		w.currentTask.Store("")
		w.lastActiveTime.Store(time.Now())
	}()

	w.logger.Info("Processing task", 
		zap.String("task_id", task.ID),
		zap.String("alert_id", task.AlertID),
		zap.String("type", string(task.Type)))

	// 更新任务状态为处理中
	startTime := time.Now()
	task.Status = analysis.AnalysisStatusProcessing
	task.StartedAt = &startTime
	task.UpdatedAt = time.Now()

	if err := w.taskRepo.Update(ctx, task); err != nil {
		w.logger.Error("Failed to update task status to processing", 
			zap.String("task_id", task.ID),
			zap.Error(err))
	}

	// 更新进度
	progress := &analysis.AnalysisProgress{
		TaskID:    task.ID,
		Stage:     "initializing",
		Progress:  0,
		Message:   "开始分析任务",
		UpdatedAt: time.Now(),
	}
	w.progressTracker.UpdateProgress(ctx, task.ID, progress)

	// 创建分析请求
	request := &analysis.AnalysisRequest{
		Type:     task.Type,
		Priority: task.Priority,
		Timeout:  task.Timeout,
		Options:  make(map[string]interface{}),
	}

	// 从任务元数据中提取选项
	if options, ok := task.Metadata["options"].(map[string]interface{}); ok {
		request.Options = options
	}
	if callback, ok := task.Metadata["callback"].(string); ok {
		request.Callback = callback
	}

	// 创建带超时的上下文
	taskCtx, cancel := context.WithTimeout(ctx, task.Timeout)
	defer cancel()

	// 执行分析
	result, err := w.executeAnalysis(taskCtx, task, request)
	if err != nil {
		w.handleTaskError(ctx, task, err)
		atomic.AddInt64(&w.errorCount, 1)
		return nil, err
	}

	// 保存结果
	if err := w.resultRepo.Create(ctx, result); err != nil {
		w.logger.Error("Failed to save analysis result", 
			zap.String("task_id", task.ID),
			zap.Error(err))
		w.handleTaskError(ctx, task, fmt.Errorf("failed to save result: %w", err))
		atomic.AddInt64(&w.errorCount, 1)
		return nil, err
	}

	// 更新任务状态为完成
	completedTime := time.Now()
	task.Status = analysis.AnalysisStatusCompleted
	task.CompletedAt = &completedTime
	task.UpdatedAt = time.Now()

	if err := w.taskRepo.Update(ctx, task); err != nil {
		w.logger.Error("Failed to update task status to completed", 
			zap.String("task_id", task.ID),
			zap.Error(err))
	}

	// 更新最终进度
	finalProgress := &analysis.AnalysisProgress{
		TaskID:    task.ID,
		Stage:     "completed",
		Progress:  100,
		Message:   "分析完成",
		UpdatedAt: time.Now(),
	}
	w.progressTracker.UpdateProgress(ctx, task.ID, finalProgress)

	// 记录指标
	w.metricsCollector.RecordTaskCompleted(ctx, w.id, task.Type, time.Since(startTime))
	atomic.AddInt64(&w.processedCount, 1)

	w.logger.Info("Task completed successfully", 
		zap.String("task_id", task.ID),
		zap.Duration("processing_time", time.Since(startTime)))

	return result, nil
}

// GetID 获取工作器ID
func (w *AnalysisWorkerImpl) GetID() string {
	return w.id
}

// IsHealthy 检查工作器健康状态
func (w *AnalysisWorkerImpl) IsHealthy() bool {
	status := w.status.Load().(string)
	if status != "running" {
		return false
	}

	// 检查最后活跃时间
	lastActive := w.lastActiveTime.Load().(time.Time)
	if time.Since(lastActive) > 5*time.Minute {
		return false
	}

	return true
}

// workerLoop 工作器主循环
func (w *AnalysisWorkerImpl) workerLoop() {
	defer w.wg.Done()

	w.logger.Info("Worker loop started", zap.String("worker_id", w.id))

	for {
		select {
		case <-w.ctx.Done():
			w.logger.Info("Worker loop stopped", zap.String("worker_id", w.id))
			return
		default:
			// 从队列获取任务
			task, err := w.taskQueue.PopWithTimeout(w.ctx, 5*time.Second)
			if err != nil {
				w.logger.Error("Failed to pop task from queue", zap.Error(err))
				time.Sleep(1 * time.Second)
				continue
			}

			if task == nil {
				// 队列为空，继续等待
				w.lastActiveTime.Store(time.Now())
				continue
			}

			// 处理任务
			_, err = w.ProcessTask(w.ctx, task)
			if err != nil {
				w.logger.Error("Failed to process task", 
					zap.String("task_id", task.ID),
					zap.Error(err))
			}
		}
	}
}

// executeAnalysis 执行分析
func (w *AnalysisWorkerImpl) executeAnalysis(ctx context.Context, task *analysis.AnalysisTask, request *analysis.AnalysisRequest) (*analysis.AnalysisResult, error) {
	// 更新进度
	progress := &analysis.AnalysisProgress{
		TaskID:    task.ID,
		Stage:     "analyzing",
		Progress:  25,
		Message:   "正在执行分析",
		UpdatedAt: time.Now(),
	}
	w.progressTracker.UpdateProgress(ctx, task.ID, progress)

	// 验证分析请求
	if err := w.analysisEngine.ValidateRequest(request); err != nil {
		return nil, fmt.Errorf("invalid analysis request: %w", err)
	}

	// 更新进度
	progress.Progress = 50
	progress.Message = "分析引擎处理中"
	progress.UpdatedAt = time.Now()
	w.progressTracker.UpdateProgress(ctx, task.ID, progress)

	// 执行分析
	startTime := time.Now()
	result, err := w.analysisEngine.Analyze(ctx, request)
	if err != nil {
		return nil, fmt.Errorf("analysis engine failed: %w", err)
	}
	processingTime := time.Since(startTime)

	// 更新进度
	progress.Progress = 75
	progress.Message = "处理分析结果"
	progress.UpdatedAt = time.Now()
	w.progressTracker.UpdateProgress(ctx, task.ID, progress)

	// 设置结果基本信息
	result.ID = uuid.New().String()
	result.TaskID = task.ID
	result.AlertID = task.AlertID
	result.Type = task.Type
	result.Status = analysis.AnalysisStatusCompleted
	result.ProcessingTime = processingTime
	result.CreatedAt = time.Now()
	result.UpdatedAt = time.Now()

	// 设置元数据
	if result.Metadata == nil {
		result.Metadata = make(map[string]interface{})
	}
	result.Metadata["worker_id"] = w.id
	result.Metadata["processing_start"] = startTime
	result.Metadata["processing_end"] = time.Now()
	result.Metadata["engine_info"] = w.analysisEngine.GetEngineInfo()

	// 更新进度
	progress.Progress = 90
	progress.Message = "保存分析结果"
	progress.UpdatedAt = time.Now()
	w.progressTracker.UpdateProgress(ctx, task.ID, progress)

	return result, nil
}

// handleTaskError 处理任务错误
func (w *AnalysisWorkerImpl) handleTaskError(ctx context.Context, task *analysis.AnalysisTask, err error) {
	w.logger.Error("Task processing failed", 
		zap.String("task_id", task.ID),
		zap.String("alert_id", task.AlertID),
		zap.String("type", string(task.Type)),
		zap.Error(err))

	// 更新任务状态
	task.Status = analysis.AnalysisStatusFailed
	task.UpdatedAt = time.Now()

	if err := w.taskRepo.Update(ctx, task); err != nil {
		w.logger.Error("Failed to update task status to failed", 
			zap.String("task_id", task.ID),
			zap.Error(err))
	}

	// 创建错误结果
	errorResult := &analysis.AnalysisResult{
		ID:           uuid.New().String(),
		TaskID:       task.ID,
		AlertID:      task.AlertID,
		Type:         task.Type,
		Status:       analysis.AnalysisStatusFailed,
		ErrorMessage: err.Error(),
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		Metadata: map[string]interface{}{
			"worker_id": w.id,
			"error_time": time.Now(),
		},
	}

	// 保存错误结果
	if err := w.resultRepo.Create(ctx, errorResult); err != nil {
		w.logger.Error("Failed to save error result", 
			zap.String("task_id", task.ID),
			zap.Error(err))
	}

	// 更新错误进度
	errorProgress := &analysis.AnalysisProgress{
		TaskID:    task.ID,
		Stage:     "failed",
		Progress:  0,
		Message:   fmt.Sprintf("分析失败: %s", err.Error()),
		UpdatedAt: time.Now(),
	}
	w.progressTracker.UpdateProgress(ctx, task.ID, errorProgress)

	// 记录错误指标
	w.metricsCollector.RecordTaskFailed(ctx, w.id, task.Type, err)
}