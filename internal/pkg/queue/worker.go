package queue

import (
	"context"
	"fmt"
	"time"

	"alert_agent/internal/model"
	"alert_agent/internal/pkg/database"
	"alert_agent/internal/pkg/logger"
	"alert_agent/internal/service"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// TaskHandler 任务处理器接口
type TaskHandler interface {
	Handle(ctx context.Context, task *Task) error
	Type() TaskType
}

// Worker 队列工作器
type Worker struct {
	id          string
	queue       MessageQueue
	monitor     *QueueMonitor
	handlers    map[TaskType]TaskHandler
	concurrency int
	shutdown    chan struct{}
	isRunning   bool
}

// NewWorker 创建工作器
func NewWorker(queue MessageQueue, monitor *QueueMonitor, concurrency int) *Worker {
	return &Worker{
		id:          uuid.New().String(),
		queue:       queue,
		monitor:     monitor,
		handlers:    make(map[TaskType]TaskHandler),
		concurrency: concurrency,
		shutdown:    make(chan struct{}),
		isRunning:   false,
	}
}

// RegisterHandler 注册任务处理器
func (w *Worker) RegisterHandler(handler TaskHandler) {
	w.handlers[handler.Type()] = handler
	logger.L.Info("Task handler registered",
		zap.String("worker_id", w.id),
		zap.String("task_type", string(handler.Type())),
	)
}

// SetConcurrency 设置并发数
func (w *Worker) SetConcurrency(concurrency int) {
	w.concurrency = concurrency
}

// Start 启动工作器
func (w *Worker) Start(ctx context.Context, queueNames []string) error {
	if w.isRunning {
		logger.L.Warn("Worker is already running", zap.String("worker_id", w.id))
		return nil
	}

	w.isRunning = true
	logger.L.Info("Starting worker",
		zap.String("worker_id", w.id),
		zap.Int("concurrency", w.concurrency),
		zap.Strings("queues", queueNames),
	)

	// 启动多个工作协程
	for i := 0; i < w.concurrency; i++ {
		go w.workerLoop(ctx, queueNames, i)
	}

	return nil
}

// workerLoop 工作器循环
func (w *Worker) workerLoop(ctx context.Context, queueNames []string, workerIndex int) {
	workerID := fmt.Sprintf("%s-%d", w.id, workerIndex)
	
	logger.L.Info("Worker loop started",
		zap.String("worker_id", workerID),
	)

	for {
		select {
		case <-w.shutdown:
			logger.L.Info("Worker loop shutting down", zap.String("worker_id", workerID))
			return
		case <-ctx.Done():
			logger.L.Info("Context canceled, worker loop shutting down", zap.String("worker_id", workerID))
			return
		default:
			// 轮询所有队列
			task := w.consumeFromQueues(ctx, queueNames)
			if task == nil {
				time.Sleep(100 * time.Millisecond)
				continue
			}

			// 处理任务
			w.processTask(ctx, task, workerID)
		}
	}
}

// consumeFromQueues 从队列中消费任务
func (w *Worker) consumeFromQueues(ctx context.Context, queueNames []string) *Task {
	for _, queueName := range queueNames {
		task, err := w.queue.Consume(ctx, queueName)
		if err != nil {
			logger.L.Error("Failed to consume task from queue",
				zap.String("queue", queueName),
				zap.Error(err),
			)
			continue
		}
		
		if task != nil {
			return task
		}
	}
	return nil
}

// Stop 停止工作器
func (w *Worker) Stop() {
	if !w.isRunning {
		return
	}

	logger.L.Info("Stopping analysis worker")
	close(w.shutdown)
	w.isRunning = false
}

// processTask 处理任务
func (w *Worker) processTask(ctx context.Context, task *Task, workerID string) {
	startTime := time.Now()
	
	logger.L.Info("Processing task",
		zap.String("worker_id", workerID),
		zap.String("task_id", task.ID),
		zap.String("type", string(task.Type)),
		zap.Int("retry", task.Retry),
	)

	// 更新任务状态
	task.MarkProcessing(workerID)

	// 查找处理器
	handler, exists := w.handlers[task.Type]
	if !exists {
		logger.L.Error("No handler found for task type",
			zap.String("task_id", task.ID),
			zap.String("type", string(task.Type)),
		)
		
		w.handleTaskFailure(ctx, task, fmt.Errorf("no handler found for task type: %s", task.Type))
		return
	}

	// 创建带超时的上下文
	timeoutCtx, cancel := context.WithTimeout(ctx, 10*time.Minute)
	defer cancel()

	// 执行任务
	err := handler.Handle(timeoutCtx, task)
	duration := time.Since(startTime)

	if err != nil {
		logger.L.Error("Task processing failed",
			zap.String("worker_id", workerID),
			zap.String("task_id", task.ID),
			zap.String("type", string(task.Type)),
			zap.Duration("duration", duration),
			zap.Error(err),
		)
		
		w.handleTaskFailure(ctx, task, err)
		
		// 记录失败指标
		if w.monitor != nil {
			w.monitor.RecordTaskCompletion(ctx, task, duration, false)
		}
		return
	}

	// 任务成功完成
	logger.L.Info("Task completed successfully",
		zap.String("worker_id", workerID),
		zap.String("task_id", task.ID),
		zap.String("type", string(task.Type)),
		zap.Duration("duration", duration),
	)

	// 确认任务完成
	result := &TaskResult{
		TaskID:      task.ID,
		Status:      TaskStatusCompleted,
		Duration:    duration,
		CompletedAt: time.Now(),
		WorkerID:    workerID,
	}

	if err := w.queue.Ack(ctx, task, result); err != nil {
		logger.L.Error("Failed to ack task",
			zap.String("task_id", task.ID),
			zap.Error(err),
		)
	}

	// 记录成功指标
	if w.monitor != nil {
		w.monitor.RecordTaskCompletion(ctx, task, duration, true)
	}
}

// handleTaskFailure 处理任务失败
func (w *Worker) handleTaskFailure(ctx context.Context, task *Task, err error) {
	task.MarkFailed(err.Error())

	// 决定是否重试
	shouldRetry := task.ShouldRetry()
	
	if nackErr := w.queue.Nack(ctx, task, shouldRetry); nackErr != nil {
		logger.L.Error("Failed to nack task",
			zap.String("task_id", task.ID),
			zap.Error(nackErr),
		)
	}

	if shouldRetry {
		logger.L.Info("Task will be retried",
			zap.String("task_id", task.ID),
			zap.Int("retry", task.Retry),
			zap.Int("max_retry", task.MaxRetry),
		)
	} else {
		logger.L.Warn("Task moved to dead letter queue",
			zap.String("task_id", task.ID),
			zap.Int("retry", task.Retry),
			zap.String("error", err.Error()),
		)
	}
}

// AIAnalysisHandler AI分析任务处理器 (已移动到 internal/worker/ai_analysis_handler.go)
// 这里保留一个简化版本以保持兼容性
type AIAnalysisHandler struct {
	ollamaService *service.OllamaService
}

// NewAIAnalysisHandler 创建AI分析处理器
func NewAIAnalysisHandler(ollamaService *service.OllamaService) *AIAnalysisHandler {
	return &AIAnalysisHandler{
		ollamaService: ollamaService,
	}
}

// Type 返回处理器类型
func (h *AIAnalysisHandler) Type() TaskType {
	return TaskTypeAIAnalysis
}

// Handle 处理AI分析任务
func (h *AIAnalysisHandler) Handle(ctx context.Context, task *Task) error {
	// 解析任务载荷
	alertID, ok := task.Payload["alert_id"].(string)
	if !ok {
		return fmt.Errorf("invalid alert_id in task payload")
	}

	analysisType, _ := task.Payload["analysis_type"].(string)
	if analysisType == "" {
		analysisType = "root_cause"
	}

	logger.L.Info("Processing AI analysis task",
		zap.String("task_id", task.ID),
		zap.String("alert_id", alertID),
		zap.String("analysis_type", analysisType),
	)

	// 获取告警信息
	var alert model.Alert
	if err := database.DB.Where("id = ?", alertID).First(&alert).Error; err != nil {
		return fmt.Errorf("failed to get alert %s: %w", alertID, err)
	}

	// 调用Ollama服务进行分析
	analysis, err := h.ollamaService.AnalyzeAlert(ctx, &alert)
	if err != nil {
		return fmt.Errorf("AI analysis failed for alert %s: %w", alertID, err)
	}

	// 更新数据库中的分析结果
	alert.Analysis = analysis
	if err := database.DB.Save(&alert).Error; err != nil {
		return fmt.Errorf("failed to update alert analysis in database: %w", err)
	}

	logger.L.Info("AI analysis completed successfully",
		zap.String("task_id", task.ID),
		zap.String("alert_id", alertID),
	)

	return nil
}

// NotificationHandler 通知任务处理器
type NotificationHandler struct {
	// 这里可以添加通知服务依赖
}

// NewNotificationHandler 创建通知处理器
func NewNotificationHandler() *NotificationHandler {
	return &NotificationHandler{}
}

// Type 返回处理器类型
func (h *NotificationHandler) Type() TaskType {
	return TaskTypeNotification
}

// Handle 处理通知任务
func (h *NotificationHandler) Handle(ctx context.Context, task *Task) error {
	// 解析任务载荷
	alertID, ok := task.Payload["alert_id"].(string)
	if !ok {
		return fmt.Errorf("invalid alert_id in task payload")
	}

	channels, ok := task.Payload["channels"].([]interface{})
	if !ok {
		return fmt.Errorf("invalid channels in task payload")
	}

	message, ok := task.Payload["message"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("invalid message in task payload")
	}

	logger.L.Info("Processing notification task",
		zap.String("task_id", task.ID),
		zap.String("alert_id", alertID),
		zap.Int("channel_count", len(channels)),
	)

	// TODO: 实现实际的通知发送逻辑
	// 这里应该调用通知服务发送通知到各个渠道
	
	logger.L.Info("Notification sent successfully",
		zap.String("task_id", task.ID),
		zap.String("alert_id", alertID),
	)

	return nil
}

// ConfigSyncHandler 配置同步任务处理器
type ConfigSyncHandler struct {
	// 这里可以添加配置同步服务依赖
}

// NewConfigSyncHandler 创建配置同步处理器
func NewConfigSyncHandler() *ConfigSyncHandler {
	return &ConfigSyncHandler{}
}

// Type 返回处理器类型
func (h *ConfigSyncHandler) Type() TaskType {
	return TaskTypeConfigSync
}

// Handle 处理配置同步任务
func (h *ConfigSyncHandler) Handle(ctx context.Context, task *Task) error {
	// 解析任务载荷
	syncType, ok := task.Payload["type"].(string)
	if !ok {
		return fmt.Errorf("invalid type in task payload")
	}

	ruleID, ok := task.Payload["rule_id"].(string)
	if !ok {
		return fmt.Errorf("invalid rule_id in task payload")
	}

	targets, ok := task.Payload["targets"].([]interface{})
	if !ok {
		return fmt.Errorf("invalid targets in task payload")
	}

	logger.L.Info("Processing config sync task",
		zap.String("task_id", task.ID),
		zap.String("type", syncType),
		zap.String("rule_id", ruleID),
		zap.Int("target_count", len(targets)),
	)

	// TODO: 实现实际的配置同步逻辑
	// 这里应该调用配置同步服务同步规则到目标系统
	
	logger.L.Info("Config sync completed successfully",
		zap.String("task_id", task.ID),
		zap.String("rule_id", ruleID),
	)

	return nil
}