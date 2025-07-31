package worker

import (
	"context"
	"fmt"
	"sync"
	"time"

	"alert_agent/internal/pkg/logger"
	"alert_agent/internal/pkg/queue"
	"alert_agent/internal/service"

	"go.uber.org/zap"
)

// WorkerInstance Worker实例
type WorkerInstance struct {
	config       *WorkerConfig
	messageQueue queue.MessageQueue
	monitor      *queue.QueueMonitor
	handlers     map[queue.TaskType]queue.TaskHandler
	healthServer *HealthServer
	isRunning    bool
	stopChan     chan struct{}
	wg           sync.WaitGroup
	mutex        sync.RWMutex
	stats        *WorkerStats
}

// WorkerStats Worker统计信息
type WorkerStats struct {
	Name            string            `json:"name"`
	Type            string            `json:"type"`
	Status          string            `json:"status"`
	Concurrency     int               `json:"concurrency"`
	Queues          []string          `json:"queues"`
	TasksProcessed  int64             `json:"tasks_processed"`
	TasksSucceeded  int64             `json:"tasks_succeeded"`
	TasksFailed     int64             `json:"tasks_failed"`
	AverageLatency  time.Duration     `json:"average_latency"`
	LastActivity    time.Time         `json:"last_activity"`
	StartTime       time.Time         `json:"start_time"`
	QueueStats      map[string]*queue.QueueStats `json:"queue_stats"`
}

// Start 启动Worker
func (w *WorkerInstance) Start(ctx context.Context) error {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	if w.isRunning {
		return fmt.Errorf("worker %s is already running", w.config.Name)
	}

	w.isRunning = true
	w.stopChan = make(chan struct{})
	w.stats = &WorkerStats{
		Name:        w.config.Name,
		Type:        w.config.Type,
		Status:      "running",
		Concurrency: w.config.Concurrency,
		Queues:      w.config.Queues,
		StartTime:   time.Now(),
		QueueStats:  make(map[string]*queue.QueueStats),
	}

	// 启动健康检查服务器
	if err := w.healthServer.Start(); err != nil {
		w.isRunning = false
		return fmt.Errorf("failed to start health server: %w", err)
	}

	// 启动工作协程
	for i := 0; i < w.config.Concurrency; i++ {
		w.wg.Add(1)
		go w.workerLoop(ctx, i)
	}

	// 启动统计更新协程
	w.wg.Add(1)
	go w.statsUpdateLoop(ctx)

	logger.L.Info("Worker started",
		zap.String("name", w.config.Name),
		zap.String("type", w.config.Type),
		zap.Int("concurrency", w.config.Concurrency),
		zap.Strings("queues", w.config.Queues),
	)

	return nil
}

// Stop 停止Worker
func (w *WorkerInstance) Stop(ctx context.Context) error {
	w.mutex.Lock()
	if !w.isRunning {
		w.mutex.Unlock()
		return nil
	}

	logger.L.Info("Stopping worker", zap.String("name", w.config.Name))
	
	w.isRunning = false
	w.stats.Status = "stopping"
	close(w.stopChan)
	w.mutex.Unlock()

	// 等待所有工作协程结束
	done := make(chan struct{})
	go func() {
		w.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		logger.L.Info("All worker goroutines stopped", zap.String("name", w.config.Name))
	case <-ctx.Done():
		logger.L.Warn("Worker stop timeout", zap.String("name", w.config.Name))
	}

	// 停止健康检查服务器
	if err := w.healthServer.Stop(ctx); err != nil {
		logger.L.Error("Failed to stop health server", zap.Error(err))
	}

	w.mutex.Lock()
	w.stats.Status = "stopped"
	w.mutex.Unlock()

	logger.L.Info("Worker stopped", zap.String("name", w.config.Name))
	return nil
}

// workerLoop 工作循环
func (w *WorkerInstance) workerLoop(ctx context.Context, workerIndex int) {
	defer w.wg.Done()

	workerID := fmt.Sprintf("%s-%d", w.config.Name, workerIndex)
	
	logger.L.Info("Worker loop started", zap.String("worker_id", workerID))

	for {
		select {
		case <-w.stopChan:
			logger.L.Info("Worker loop stopping", zap.String("worker_id", workerID))
			return
		case <-ctx.Done():
			logger.L.Info("Context cancelled, worker loop stopping", zap.String("worker_id", workerID))
			return
		default:
			// 轮询所有队列
			task := w.consumeFromQueues(ctx)
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
func (w *WorkerInstance) consumeFromQueues(ctx context.Context) *queue.Task {
	for _, queueName := range w.config.Queues {
		task, err := w.messageQueue.Consume(ctx, queueName)
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

// processTask 处理任务
func (w *WorkerInstance) processTask(ctx context.Context, task *queue.Task, workerID string) {
	startTime := time.Now()
	
	logger.L.Info("Processing task",
		zap.String("worker_id", workerID),
		zap.String("task_id", task.ID),
		zap.String("type", string(task.Type)),
		zap.Int("retry", task.Retry),
	)

	// 更新统计信息
	w.updateStats(func(stats *WorkerStats) {
		stats.TasksProcessed++
		stats.LastActivity = time.Now()
	})

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
		
		// 更新失败统计
		w.updateStats(func(stats *WorkerStats) {
			stats.TasksFailed++
		})
		
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
	result := &queue.TaskResult{
		TaskID:      task.ID,
		Status:      queue.TaskStatusCompleted,
		Duration:    duration,
		CompletedAt: time.Now(),
		WorkerID:    workerID,
	}

	if err := w.messageQueue.Ack(ctx, task, result); err != nil {
		logger.L.Error("Failed to ack task",
			zap.String("task_id", task.ID),
			zap.Error(err),
		)
	}

	// 更新成功统计
	w.updateStats(func(stats *WorkerStats) {
		stats.TasksSucceeded++
		// 更新平均延迟
		totalTasks := stats.TasksSucceeded + stats.TasksFailed
		if totalTasks > 0 {
			stats.AverageLatency = time.Duration(
				(int64(stats.AverageLatency)*int64(totalTasks-1) + int64(duration)) / int64(totalTasks),
			)
		}
	})

	// 记录成功指标
	if w.monitor != nil {
		w.monitor.RecordTaskCompletion(ctx, task, duration, true)
	}
}

// handleTaskFailure 处理任务失败
func (w *WorkerInstance) handleTaskFailure(ctx context.Context, task *queue.Task, err error) {
	task.MarkFailed(err.Error())

	// 决定是否重试
	shouldRetry := task.ShouldRetry()
	
	if nackErr := w.messageQueue.Nack(ctx, task, shouldRetry); nackErr != nil {
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

// statsUpdateLoop 统计更新循环
func (w *WorkerInstance) statsUpdateLoop(ctx context.Context) {
	defer w.wg.Done()

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-w.stopChan:
			return
		case <-ctx.Done():
			return
		case <-ticker.C:
			w.updateQueueStats(ctx)
		}
	}
}

// updateQueueStats 更新队列统计信息
func (w *WorkerInstance) updateQueueStats(ctx context.Context) {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	for _, queueName := range w.config.Queues {
		stats, err := w.messageQueue.GetQueueStats(ctx, queueName)
		if err != nil {
			logger.L.Error("Failed to get queue stats",
				zap.String("queue", queueName),
				zap.Error(err),
			)
			continue
		}
		w.stats.QueueStats[queueName] = stats
	}
}

// updateStats 安全地更新统计信息
func (w *WorkerInstance) updateStats(updateFunc func(*WorkerStats)) {
	w.mutex.Lock()
	defer w.mutex.Unlock()
	
	if w.stats != nil {
		updateFunc(w.stats)
	}
}

// GetStats 获取Worker统计信息
func (w *WorkerInstance) GetStats(ctx context.Context) (*WorkerStats, error) {
	w.mutex.RLock()
	defer w.mutex.RUnlock()

	if w.stats == nil {
		return nil, fmt.Errorf("worker stats not available")
	}

	// 创建副本以避免并发访问问题
	statsCopy := *w.stats
	statsCopy.QueueStats = make(map[string]*queue.QueueStats)
	for k, v := range w.stats.QueueStats {
		statsCopy.QueueStats[k] = v
	}

	return &statsCopy, nil
}

// IsRunning 检查Worker是否正在运行
func (w *WorkerInstance) IsRunning() bool {
	w.mutex.RLock()
	defer w.mutex.RUnlock()
	return w.isRunning
}

// GetConfig 获取Worker配置
func (w *WorkerInstance) GetConfig() *WorkerConfig {
	return w.config
}

// registerHandlers 注册任务处理器
func (w *WorkerInstance) registerHandlers() error {
	// 根据Worker类型注册相应的处理器
	switch w.config.Type {
	case "ai-analysis":
		aiHandler := NewAIAnalysisHandler(service.NewOllamaService())
		w.handlers[aiHandler.Type()] = aiHandler
		
	case "notification":
		notificationHandler := NewNotificationHandler()
		w.handlers[notificationHandler.Type()] = notificationHandler
		
	case "config-sync":
		configSyncHandler := NewConfigSyncHandler()
		w.handlers[configSyncHandler.Type()] = configSyncHandler
		
	case "general":
		// 注册通用处理器
		aiHandler := NewAIAnalysisHandler(service.NewOllamaService())
		w.handlers[aiHandler.Type()] = aiHandler
		
		notificationHandler := NewNotificationHandler()
		w.handlers[notificationHandler.Type()] = notificationHandler
		
		configSyncHandler := NewConfigSyncHandler()
		w.handlers[configSyncHandler.Type()] = configSyncHandler
		
	default:
		// 默认注册所有处理器
		aiHandler := NewAIAnalysisHandler(service.NewOllamaService())
		w.handlers[aiHandler.Type()] = aiHandler
		
		notificationHandler := NewNotificationHandler()
		w.handlers[notificationHandler.Type()] = notificationHandler
		
		configSyncHandler := NewConfigSyncHandler()
		w.handlers[configSyncHandler.Type()] = configSyncHandler
	}

	logger.L.Info("Task handlers registered",
		zap.String("worker", w.config.Name),
		zap.Int("handler_count", len(w.handlers)),
	)

	return nil
}