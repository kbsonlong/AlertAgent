package service

import (
	"context"
	"fmt"
	"time"

	"alert_agent/internal/pkg/logger"
	"alert_agent/internal/pkg/queue"
	"alert_agent/internal/pkg/redis"

	"go.uber.org/zap"
)

// TaskService 任务服务
type TaskService struct {
	messageQueue MessageQueue
	taskProducer queue.TaskProducer
	monitor      *queue.QueueMonitor
	workers      map[string]*queue.Worker
}

// MessageQueue 消息队列接口别名
type MessageQueue = queue.MessageQueue

// NewTaskService 创建任务服务
func NewTaskService() *TaskService {
	// 创建Redis消息队列
	messageQueue := queue.NewRedisMessageQueue(redis.Client, "alertagent")
	
	// 创建任务生产者
	taskProducer := queue.NewTaskProducer(messageQueue)
	
	// 创建队列监控器
	monitor := queue.NewQueueMonitor(messageQueue, redis.Client, "alertagent")

	return &TaskService{
		messageQueue: messageQueue,
		taskProducer: taskProducer,
		monitor:      monitor,
		workers:      make(map[string]*queue.Worker),
	}
}

// GetTaskProducer 获取任务生产者
func (s *TaskService) GetTaskProducer() queue.TaskProducer {
	return s.taskProducer
}

// GetMessageQueue 获取消息队列
func (s *TaskService) GetMessageQueue() MessageQueue {
	return s.messageQueue
}

// GetMonitor 获取队列监控器
func (s *TaskService) GetMonitor() *queue.QueueMonitor {
	return s.monitor
}

// StartWorker 启动工作器
func (s *TaskService) StartWorker(ctx context.Context, workerName string, concurrency int, queueNames []string) error {
	if _, exists := s.workers[workerName]; exists {
		return fmt.Errorf("worker %s already exists", workerName)
	}

	// 创建工作器
	worker := queue.NewWorker(s.messageQueue, s.monitor, concurrency)
	
	// 注册任务处理器
	s.registerHandlers(worker)
	
	// 启动工作器
	if err := worker.Start(ctx, queueNames); err != nil {
		return fmt.Errorf("failed to start worker %s: %w", workerName, err)
	}

	s.workers[workerName] = worker
	
	logger.L.Info("Worker started successfully",
		zap.String("worker_name", workerName),
		zap.Int("concurrency", concurrency),
		zap.Strings("queues", queueNames),
	)

	return nil
}

// StopWorker 停止工作器
func (s *TaskService) StopWorker(workerName string) error {
	worker, exists := s.workers[workerName]
	if !exists {
		return fmt.Errorf("worker %s not found", workerName)
	}

	worker.Stop()
	delete(s.workers, workerName)
	
	logger.L.Info("Worker stopped", zap.String("worker_name", workerName))
	return nil
}

// registerHandlers 注册任务处理器
func (s *TaskService) registerHandlers(worker *queue.Worker) {
	// 注册AI分析处理器
	aiHandler := queue.NewAIAnalysisHandler(NewOllamaService())
	worker.RegisterHandler(aiHandler)
	
	// 注册通知处理器
	notificationHandler := queue.NewNotificationHandler()
	worker.RegisterHandler(notificationHandler)
	
	// 注册配置同步处理器
	configSyncHandler := queue.NewConfigSyncHandler()
	worker.RegisterHandler(configSyncHandler)
}

// PublishTask 发布任务
func (s *TaskService) PublishTask(ctx context.Context, queueName string, task *queue.Task) error {
	return s.taskProducer.PublishTask(ctx, queueName, task)
}

// PublishDelayedTask 发布延迟任务
func (s *TaskService) PublishDelayedTask(ctx context.Context, queueName string, task *queue.Task, delay time.Duration) error {
	return s.taskProducer.PublishDelayedTask(ctx, queueName, task, delay)
}

// GetQueueStats 获取队列统计信息
func (s *TaskService) GetQueueStats(ctx context.Context, queueName string) (*queue.QueueStats, error) {
	return s.messageQueue.GetQueueStats(ctx, queueName)
}

// GetAllQueueMetrics 获取所有队列指标
func (s *TaskService) GetAllQueueMetrics(ctx context.Context) (map[string]*queue.QueueMetrics, error) {
	return s.monitor.GetAllQueueMetrics(ctx)
}

// GetTaskStatus 获取任务状态
func (s *TaskService) GetTaskStatus(ctx context.Context, taskID string) (*queue.Task, error) {
	return s.taskProducer.GetTaskStatus(ctx, taskID)
}

// CleanupExpiredTasks 清理过期任务
func (s *TaskService) CleanupExpiredTasks(ctx context.Context) error {
	queueNames := []string{
		string(queue.TaskTypeAIAnalysis),
		string(queue.TaskTypeNotification),
		string(queue.TaskTypeConfigSync),
	}

	for _, queueName := range queueNames {
		if err := s.monitor.CleanupExpiredTasks(ctx, queueName, 30*time.Minute); err != nil {
			logger.L.Error("Failed to cleanup expired tasks",
				zap.String("queue", queueName),
				zap.Error(err),
			)
		}
	}

	return nil
}

// GetHealthStatus 获取任务系统健康状态
func (s *TaskService) GetHealthStatus(ctx context.Context) (map[string]interface{}, error) {
	return s.monitor.GetHealthStatus(ctx)
}

// StartDefaultWorkers 启动默认工作器
func (s *TaskService) StartDefaultWorkers(ctx context.Context) error {
	// 启动AI分析工作器
	if err := s.StartWorker(ctx, "ai-analysis", 2, []string{string(queue.TaskTypeAIAnalysis)}); err != nil {
		return fmt.Errorf("failed to start AI analysis worker: %w", err)
	}

	// 启动通知工作器
	if err := s.StartWorker(ctx, "notification", 3, []string{string(queue.TaskTypeNotification)}); err != nil {
		return fmt.Errorf("failed to start notification worker: %w", err)
	}

	// 启动配置同步工作器
	if err := s.StartWorker(ctx, "config-sync", 1, []string{string(queue.TaskTypeConfigSync)}); err != nil {
		return fmt.Errorf("failed to start config sync worker: %w", err)
	}

	// 启动通用工作器处理其他任务
	if err := s.StartWorker(ctx, "general", 2, []string{
		string(queue.TaskTypeRuleUpdate),
		string(queue.TaskTypeHealthCheck),
	}); err != nil {
		return fmt.Errorf("failed to start general worker: %w", err)
	}

	logger.L.Info("All default workers started successfully")
	return nil
}

// StopAllWorkers 停止所有工作器
func (s *TaskService) StopAllWorkers() {
	for workerName := range s.workers {
		if err := s.StopWorker(workerName); err != nil {
			logger.L.Error("Failed to stop worker",
				zap.String("worker_name", workerName),
				zap.Error(err),
			)
		}
	}
	logger.L.Info("All workers stopped")
}

// SchedulePeriodicTasks 调度周期性任务
func (s *TaskService) SchedulePeriodicTasks(ctx context.Context) error {
	// 调度清理任务（每30分钟执行一次）
	cleanupTask := &queue.Task{
		Type:     queue.TaskTypeHealthCheck,
		Priority: queue.PriorityLow,
		MaxRetry: 1,
		Payload: map[string]interface{}{
			"task_type": "cleanup_expired_tasks",
		},
	}

	if err := s.PublishDelayedTask(ctx, string(queue.TaskTypeHealthCheck), cleanupTask, 30*time.Minute); err != nil {
		return fmt.Errorf("failed to schedule cleanup task: %w", err)
	}

	// 调度健康检查任务（每5分钟执行一次）
	healthCheckTask := &queue.Task{
		Type:     queue.TaskTypeHealthCheck,
		Priority: queue.PriorityLow,
		MaxRetry: 1,
		Payload: map[string]interface{}{
			"task_type": "health_check",
		},
	}

	if err := s.PublishDelayedTask(ctx, string(queue.TaskTypeHealthCheck), healthCheckTask, 5*time.Minute); err != nil {
		return fmt.Errorf("failed to schedule health check task: %w", err)
	}

	logger.L.Info("Periodic tasks scheduled successfully")
	return nil
}

// Close 关闭任务服务
func (s *TaskService) Close() error {
	// 停止所有工作器
	s.StopAllWorkers()
	
	// 关闭消息队列
	if err := s.messageQueue.Close(); err != nil {
		return fmt.Errorf("failed to close message queue: %w", err)
	}

	logger.L.Info("Task service closed successfully")
	return nil
}