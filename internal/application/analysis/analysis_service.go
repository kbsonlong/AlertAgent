package analysis

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"alert_agent/internal/domain/analysis"
	"alert_agent/internal/pkg/logger"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// AnalysisServiceImpl 分析服务实现
type AnalysisServiceImpl struct {
	taskQueue       analysis.AnalysisTaskQueue
	taskRepo        analysis.AnalysisTaskRepository
	resultRepo      analysis.AnalysisResultRepository
	progressTracker analysis.AnalysisProgressTracker
	workerManager   analysis.AnalysisWorkerManager
	notifier        analysis.AnalysisNotifier
	metricsCollector analysis.AnalysisMetricsCollector
	retryPolicy     analysis.AnalysisRetryPolicy
	logger          *zap.Logger
}

// NewAnalysisService 创建分析服务实例
func NewAnalysisService(
	taskQueue analysis.AnalysisTaskQueue,
	taskRepo analysis.AnalysisTaskRepository,
	resultRepo analysis.AnalysisResultRepository,
	progressTracker analysis.AnalysisProgressTracker,
	workerManager analysis.AnalysisWorkerManager,
	notifier analysis.AnalysisNotifier,
	metricsCollector analysis.AnalysisMetricsCollector,
	retryPolicy analysis.AnalysisRetryPolicy,
) analysis.AnalysisService {
	return &AnalysisServiceImpl{
		taskQueue:       taskQueue,
		taskRepo:        taskRepo,
		resultRepo:      resultRepo,
		progressTracker: progressTracker,
		workerManager:   workerManager,
		notifier:        notifier,
		metricsCollector: metricsCollector,
		retryPolicy:     retryPolicy,
		logger:          logger.L.Named("analysis-service"),
	}
}

// SubmitAnalysis 提交分析请求
func (s *AnalysisServiceImpl) SubmitAnalysis(ctx context.Context, request *analysis.AnalysisRequest) (*analysis.AnalysisTask, error) {
	// 验证请求
	if err := s.validateAnalysisRequest(request); err != nil {
		return nil, fmt.Errorf("invalid analysis request: %w", err)
	}

	// 检查是否已有相同的分析任务
	alertIDStr := strconv.FormatUint(uint64(request.Alert.ID), 10)
	existingTasks, err := s.taskRepo.GetByAlertID(ctx, alertIDStr)
	if err != nil {
		s.logger.Error("Failed to check existing tasks", zap.Error(err))
	} else {
		for _, task := range existingTasks {
			if task.Type == request.Type && 
			   (task.Status == analysis.AnalysisStatusPending || task.Status == analysis.AnalysisStatusProcessing) {
				s.logger.Info("Found existing task for alert", 
					zap.String("task_id", task.ID),
					zap.String("alert_id", alertIDStr),
					zap.String("type", string(request.Type)))
				return task, nil
			}
		}
	}

	// 创建新任务
	task := &analysis.AnalysisTask{
		ID:         uuid.New().String(),
		AlertID:    alertIDStr,
		Type:       request.Type,
		Status:     analysis.AnalysisStatusPending,
		Priority:   request.Priority,
		RetryCount: 0,
		MaxRetries: s.retryPolicy.GetMaxRetries(request.Type),
		Timeout:    request.Timeout,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
		Metadata: map[string]interface{}{
			"alert_name":   request.Alert.Name,
			"alert_level":  request.Alert.Level,
			"alert_source": request.Alert.Source,
			"options":      request.Options,
			"callback":     request.Callback,
		},
	}

	// 设置默认值
	if task.Priority <= 0 {
		task.Priority = 5 // 默认中等优先级
	}
	if task.Timeout <= 0 {
		task.Timeout = 5 * time.Minute // 默认5分钟超时
	}

	// 保存任务到数据库
	if err := s.taskRepo.Create(ctx, task); err != nil {
		s.logger.Error("Failed to create task", zap.Error(err))
		return nil, fmt.Errorf("failed to create task: %w", err)
	}

	// 推送到队列
	if err := s.taskQueue.Push(ctx, task); err != nil {
		s.logger.Error("Failed to push task to queue", zap.Error(err))
		// 更新任务状态为失败
		task.Status = analysis.AnalysisStatusFailed
		s.taskRepo.Update(ctx, task)
		return nil, fmt.Errorf("failed to push task to queue: %w", err)
	}

	// 记录指标
	s.metricsCollector.RecordTaskSubmitted(ctx, request.Type)

	s.logger.Info("Analysis task submitted", 
		zap.String("task_id", task.ID),
		zap.String("alert_id", alertIDStr),
		zap.String("type", string(request.Type)),
		zap.Int("priority", task.Priority))

	return task, nil
}

// GetAnalysisResult 获取分析结果
func (s *AnalysisServiceImpl) GetAnalysisResult(ctx context.Context, taskID string) (*analysis.AnalysisResult, error) {
	result, err := s.resultRepo.GetByTaskID(ctx, taskID)
	if err != nil {
		s.logger.Error("Failed to get analysis result", 
			zap.String("task_id", taskID),
			zap.Error(err))
		return nil, fmt.Errorf("failed to get analysis result: %w", err)
	}

	return result, nil
}

// GetAnalysisProgress 获取分析进度
func (s *AnalysisServiceImpl) GetAnalysisProgress(ctx context.Context, taskID string) (*analysis.AnalysisProgress, error) {
	progress, err := s.progressTracker.GetProgress(ctx, taskID)
	if err != nil {
		s.logger.Error("Failed to get analysis progress", 
			zap.String("task_id", taskID),
			zap.Error(err))
		return nil, fmt.Errorf("failed to get analysis progress: %w", err)
	}

	return progress, nil
}

// CancelAnalysis 取消分析任务
func (s *AnalysisServiceImpl) CancelAnalysis(ctx context.Context, taskID string) error {
	// 获取任务
	task, err := s.taskRepo.GetByID(ctx, taskID)
	if err != nil {
		return fmt.Errorf("failed to get task: %w", err)
	}

	// 检查任务状态
	if task.Status == analysis.AnalysisStatusCompleted || task.Status == analysis.AnalysisStatusFailed {
		return fmt.Errorf("cannot cancel task in status: %s", task.Status)
	}

	// 更新任务状态
	task.Status = analysis.AnalysisStatusCancelled
	task.UpdatedAt = time.Now()

	if err := s.taskRepo.Update(ctx, task); err != nil {
		s.logger.Error("Failed to update task status", 
			zap.String("task_id", taskID),
			zap.Error(err))
		return fmt.Errorf("failed to update task status: %w", err)
	}

	// 从队列中移除任务
	if err := s.taskQueue.Remove(ctx, taskID); err != nil {
		s.logger.Warn("Failed to remove task from queue", 
			zap.String("task_id", taskID),
			zap.Error(err))
	}

	// 清理进度记录
	if err := s.progressTracker.DeleteProgress(ctx, taskID); err != nil {
		s.logger.Warn("Failed to delete progress", 
			zap.String("task_id", taskID),
			zap.Error(err))
	}

	s.logger.Info("Analysis task cancelled", zap.String("task_id", taskID))
	return nil
}

// RetryAnalysis 重试分析任务
func (s *AnalysisServiceImpl) RetryAnalysis(ctx context.Context, taskID string) error {
	// 获取任务
	task, err := s.taskRepo.GetByID(ctx, taskID)
	if err != nil {
		return fmt.Errorf("failed to get task: %w", err)
	}

	// 检查是否可以重试
	if task.Status != analysis.AnalysisStatusFailed {
		return fmt.Errorf("can only retry failed tasks, current status: %s", task.Status)
	}

	if task.RetryCount >= task.MaxRetries {
		return fmt.Errorf("task has reached maximum retry count: %d", task.MaxRetries)
	}

	// 重置任务状态
	task.Status = analysis.AnalysisStatusPending
	task.RetryCount++
	task.UpdatedAt = time.Now()
	task.StartedAt = nil
	task.CompletedAt = nil

	// 更新任务
	if err := s.taskRepo.Update(ctx, task); err != nil {
		s.logger.Error("Failed to update task for retry", 
			zap.String("task_id", taskID),
			zap.Error(err))
		return fmt.Errorf("failed to update task for retry: %w", err)
	}

	// 重新推送到队列
	if err := s.taskQueue.Push(ctx, task); err != nil {
		s.logger.Error("Failed to push retry task to queue", 
			zap.String("task_id", taskID),
			zap.Error(err))
		return fmt.Errorf("failed to push retry task to queue: %w", err)
	}

	s.logger.Info("Analysis task retried", 
		zap.String("task_id", taskID),
		zap.Int("retry_count", task.RetryCount))

	return nil
}

// GetAnalysisTasks 获取分析任务列表
func (s *AnalysisServiceImpl) GetAnalysisTasks(ctx context.Context, filter analysis.AnalysisFilter) ([]*analysis.AnalysisTask, error) {
	tasks, err := s.taskRepo.List(ctx, filter)
	if err != nil {
		s.logger.Error("Failed to get analysis tasks", zap.Error(err))
		return nil, fmt.Errorf("failed to get analysis tasks: %w", err)
	}

	return tasks, nil
}

// GetAnalysisResults 获取分析结果列表
func (s *AnalysisServiceImpl) GetAnalysisResults(ctx context.Context, filter analysis.AnalysisFilter) ([]*analysis.AnalysisResult, error) {
	results, err := s.resultRepo.List(ctx, filter)
	if err != nil {
		s.logger.Error("Failed to get analysis results", zap.Error(err))
		return nil, fmt.Errorf("failed to get analysis results: %w", err)
	}

	return results, nil
}

// GetAnalysisStatistics 获取分析统计信息
func (s *AnalysisServiceImpl) GetAnalysisStatistics(ctx context.Context, timeRange *analysis.TimeRange) (*analysis.AnalysisStatistics, error) {
	// 构建过滤器
	filter := analysis.AnalysisFilter{}
	if timeRange != nil {
		filter.StartTime = &timeRange.Start
		filter.EndTime = &timeRange.End
	}

	// 获取所有任务
	tasks, err := s.taskRepo.List(ctx, filter)
	if err != nil {
		s.logger.Error("Failed to get tasks for statistics", zap.Error(err))
		return nil, fmt.Errorf("failed to get tasks for statistics: %w", err)
	}

	// 计算统计信息
	stats := &analysis.AnalysisStatistics{
		TypeDistribution:   make(map[analysis.AnalysisType]int64),
		StatusDistribution: make(map[analysis.AnalysisStatus]int64),
		LastUpdated:        time.Now(),
	}

	var totalProcessingTime time.Duration
	var completedCount int64

	for _, task := range tasks {
		stats.TotalTasks++
		stats.TypeDistribution[task.Type]++
		stats.StatusDistribution[task.Status]++

		switch task.Status {
		case analysis.AnalysisStatusPending:
			stats.PendingTasks++
		case analysis.AnalysisStatusProcessing:
			stats.ProcessingTasks++
		case analysis.AnalysisStatusCompleted:
			stats.CompletedTasks++
			completedCount++
			if task.StartedAt != nil && task.CompletedAt != nil {
				totalProcessingTime += task.CompletedAt.Sub(*task.StartedAt)
			}
		case analysis.AnalysisStatusFailed:
			stats.FailedTasks++
		}
	}

	// 计算平均处理时间
	if completedCount > 0 {
		stats.AverageTime = totalProcessingTime / time.Duration(completedCount)
	}

	// 计算成功率
	if stats.TotalTasks > 0 {
		stats.SuccessRate = float64(stats.CompletedTasks) / float64(stats.TotalTasks) * 100
	}

	return stats, nil
}

// GetQueueStatus 获取队列状态
func (s *AnalysisServiceImpl) GetQueueStatus(ctx context.Context) (*analysis.QueueStatus, error) {
	status, err := s.taskQueue.GetStatus(ctx)
	if err != nil {
		s.logger.Error("Failed to get queue status", zap.Error(err))
		return nil, fmt.Errorf("failed to get queue status: %w", err)
	}

	return status, nil
}

// GetWorkerStatuses 获取工作器状态
func (s *AnalysisServiceImpl) GetWorkerStatuses(ctx context.Context) ([]*analysis.WorkerStatus, error) {
	statuses := s.workerManager.GetWorkerStatuses()
	return statuses, nil
}

// HealthCheck 健康检查
func (s *AnalysisServiceImpl) HealthCheck(ctx context.Context) error {
	// 检查队列连接
	if _, err := s.taskQueue.Size(ctx); err != nil {
		return fmt.Errorf("queue health check failed: %w", err)
	}

	// 检查工作器状态
	workerStatuses := s.workerManager.GetWorkerStatuses()
	healthyWorkers := 0
	for _, status := range workerStatuses {
		if status.Status == "running" {
			healthyWorkers++
		}
	}

	if healthyWorkers == 0 {
		return fmt.Errorf("no healthy workers available")
	}

	s.logger.Debug("Health check passed", 
		zap.Int("healthy_workers", healthyWorkers),
		zap.Int("total_workers", len(workerStatuses)))

	return nil
}

// validateAnalysisRequest 验证分析请求
func (s *AnalysisServiceImpl) validateAnalysisRequest(request *analysis.AnalysisRequest) error {
	if request == nil {
		return fmt.Errorf("request cannot be nil")
	}

	if request.Alert == nil {
		return fmt.Errorf("alert cannot be nil")
	}

	if request.Alert.ID == 0 {
		return fmt.Errorf("alert ID cannot be zero")
	}

	if request.Type == "" {
		return fmt.Errorf("analysis type cannot be empty")
	}

	// 验证分析类型
	validTypes := []analysis.AnalysisType{
		analysis.AnalysisTypeRootCause,
		analysis.AnalysisTypeImpactAssess,
		analysis.AnalysisTypeSolution,
		analysis.AnalysisTypeClassification,
		analysis.AnalysisTypePriority,
	}

	validType := false
	for _, validT := range validTypes {
		if request.Type == validT {
			validType = true
			break
		}
	}

	if !validType {
		return fmt.Errorf("invalid analysis type: %s", request.Type)
	}

	if request.Priority < 1 || request.Priority > 10 {
		request.Priority = 5 // 设置默认优先级
	}

	if request.Timeout <= 0 {
		request.Timeout = 5 * time.Minute // 设置默认超时
	}

	return nil
}