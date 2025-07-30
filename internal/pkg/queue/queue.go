package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"alert_agent/internal/pkg/logger"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// MessageQueue 消息队列接口
type MessageQueue interface {
	// Publish 发布任务到队列
	Publish(ctx context.Context, queueName string, task *Task) error
	// PublishDelayed 延迟发布任务
	PublishDelayed(ctx context.Context, queueName string, task *Task, delay time.Duration) error
	// Consume 从队列消费任务
	Consume(ctx context.Context, queueName string) (*Task, error)
	// ConsumeBlocking 阻塞式消费任务
	ConsumeBlocking(ctx context.Context, queueName string, timeout time.Duration) (*Task, error)
	// Ack 确认任务完成
	Ack(ctx context.Context, task *Task, result *TaskResult) error
	// Nack 拒绝任务（重新入队或进入死信队列）
	Nack(ctx context.Context, task *Task, requeue bool) error
	// GetQueueStats 获取队列统计信息
	GetQueueStats(ctx context.Context, queueName string) (*QueueStats, error)
	// GetTaskStatus 获取任务状态
	GetTaskStatus(ctx context.Context, taskID string) (*Task, error)
	// Close 关闭队列连接
	Close() error
}

// QueueStats 队列统计信息
type QueueStats struct {
	QueueName     string `json:"queue_name"`
	PendingCount  int64  `json:"pending_count"`
	ProcessingCount int64  `json:"processing_count"`
	CompletedCount int64  `json:"completed_count"`
	FailedCount   int64  `json:"failed_count"`
	DeadLetterCount int64  `json:"dead_letter_count"`
}

// RedisMessageQueue Redis消息队列实现
type RedisMessageQueue struct {
	client       *redis.Client
	keyPrefix    string
	maxRetries   int
	retryDelay   time.Duration
	taskTimeout  time.Duration
}

// NewRedisMessageQueue 创建Redis消息队列
func NewRedisMessageQueue(client *redis.Client, keyPrefix string) *RedisMessageQueue {
	return &RedisMessageQueue{
		client:      client,
		keyPrefix:   keyPrefix,
		maxRetries:  3,
		retryDelay:  time.Minute,
		taskTimeout: 10 * time.Minute,
	}
}

// SetRetryConfig 设置重试配置
func (q *RedisMessageQueue) SetRetryConfig(maxRetries int, retryDelay time.Duration) {
	q.maxRetries = maxRetries
	q.retryDelay = retryDelay
}

// SetTaskTimeout 设置任务超时时间
func (q *RedisMessageQueue) SetTaskTimeout(timeout time.Duration) {
	q.taskTimeout = timeout
}

// getQueueKey 获取队列键名
func (q *RedisMessageQueue) getQueueKey(queueName string) string {
	return fmt.Sprintf("%s:queue:%s", q.keyPrefix, queueName)
}

// getProcessingKey 获取处理中队列键名
func (q *RedisMessageQueue) getProcessingKey(queueName string) string {
	return fmt.Sprintf("%s:processing:%s", q.keyPrefix, queueName)
}

// getTaskKey 获取任务键名
func (q *RedisMessageQueue) getTaskKey(taskID string) string {
	return fmt.Sprintf("%s:task:%s", q.keyPrefix, taskID)
}

// getResultKey 获取结果键名
func (q *RedisMessageQueue) getResultKey(taskID string) string {
	return fmt.Sprintf("%s:result:%s", q.keyPrefix, taskID)
}

// getDelayedKey 获取延迟队列键名
func (q *RedisMessageQueue) getDelayedKey(queueName string) string {
	return fmt.Sprintf("%s:delayed:%s", q.keyPrefix, queueName)
}

// getDeadLetterKey 获取死信队列键名
func (q *RedisMessageQueue) getDeadLetterKey(queueName string) string {
	return fmt.Sprintf("%s:dead:%s", q.keyPrefix, queueName)
}

// Publish 发布任务到队列
func (q *RedisMessageQueue) Publish(ctx context.Context, queueName string, task *Task) error {
	// 生成任务ID
	if task.ID == "" {
		task.ID = uuid.New().String()
	}
	
	// 设置默认值
	if task.CreatedAt.IsZero() {
		task.CreatedAt = time.Now()
	}
	if task.ScheduledAt.IsZero() {
		task.ScheduledAt = time.Now()
	}
	if task.MaxRetry == 0 {
		task.MaxRetry = q.maxRetries
	}
	task.Status = TaskStatusPending
	task.UpdatedAt = time.Now()

	// 序列化任务
	taskData, err := task.ToJSON()
	if err != nil {
		return fmt.Errorf("failed to marshal task: %w", err)
	}

	pipe := q.client.Pipeline()
	
	// 保存任务详情
	taskKey := q.getTaskKey(task.ID)
	pipe.Set(ctx, taskKey, taskData, 24*time.Hour)
	
	// 根据优先级添加到队列
	queueKey := q.getQueueKey(queueName)
	priority := float64(task.Priority)
	pipe.ZAdd(ctx, queueKey, redis.Z{
		Score:  priority,
		Member: task.ID,
	})

	_, err = pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to publish task: %w", err)
	}

	logger.L.Debug("Task published to queue",
		zap.String("queue", queueName),
		zap.String("task_id", task.ID),
		zap.String("type", string(task.Type)),
		zap.Int("priority", int(task.Priority)),
	)

	return nil
}

// PublishDelayed 延迟发布任务
func (q *RedisMessageQueue) PublishDelayed(ctx context.Context, queueName string, task *Task, delay time.Duration) error {
	// 生成任务ID
	if task.ID == "" {
		task.ID = uuid.New().String()
	}
	
	// 设置调度时间
	task.ScheduledAt = time.Now().Add(delay)
	task.CreatedAt = time.Now()
	task.Status = TaskStatusPending
	task.UpdatedAt = time.Now()
	
	if task.MaxRetry == 0 {
		task.MaxRetry = q.maxRetries
	}

	// 序列化任务
	taskData, err := task.ToJSON()
	if err != nil {
		return fmt.Errorf("failed to marshal task: %w", err)
	}

	pipe := q.client.Pipeline()
	
	// 保存任务详情
	taskKey := q.getTaskKey(task.ID)
	pipe.Set(ctx, taskKey, taskData, 24*time.Hour)
	
	// 添加到延迟队列
	delayedKey := q.getDelayedKey(queueName)
	pipe.ZAdd(ctx, delayedKey, redis.Z{
		Score:  float64(task.ScheduledAt.Unix()),
		Member: task.ID,
	})

	_, err = pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to publish delayed task: %w", err)
	}

	logger.L.Debug("Delayed task published",
		zap.String("queue", queueName),
		zap.String("task_id", task.ID),
		zap.Duration("delay", delay),
	)

	return nil
}

// Consume 从队列消费任务
func (q *RedisMessageQueue) Consume(ctx context.Context, queueName string) (*Task, error) {
	// 首先处理延迟任务
	if err := q.processDelayedTasks(ctx, queueName); err != nil {
		logger.L.Warn("Failed to process delayed tasks", zap.Error(err))
	}

	queueKey := q.getQueueKey(queueName)
	processingKey := q.getProcessingKey(queueName)

	// 使用ZPOPMAX获取最高优先级的任务
	result, err := q.client.ZPopMax(ctx, queueKey, 1).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to pop task: %w", err)
	}

	if len(result) == 0 {
		return nil, nil
	}

	taskID := result[0].Member.(string)
	
	// 获取任务详情
	taskKey := q.getTaskKey(taskID)
	taskData, err := q.client.Get(ctx, taskKey).Bytes()
	if err != nil {
		if err == redis.Nil {
			logger.L.Warn("Task data not found", zap.String("task_id", taskID))
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get task data: %w", err)
	}

	var task Task
	if err := task.FromJSON(taskData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal task: %w", err)
	}

	// 移动到处理中队列
	pipe := q.client.Pipeline()
	pipe.ZAdd(ctx, processingKey, redis.Z{
		Score:  float64(time.Now().Unix()),
		Member: taskID,
	})
	
	// 更新任务状态
	task.MarkProcessing("")
	updatedData, _ := task.ToJSON()
	pipe.Set(ctx, taskKey, updatedData, 24*time.Hour)
	
	_, err = pipe.Exec(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to move task to processing: %w", err)
	}

	logger.L.Debug("Task consumed from queue",
		zap.String("queue", queueName),
		zap.String("task_id", task.ID),
		zap.String("type", string(task.Type)),
	)

	return &task, nil
}

// ConsumeBlocking 阻塞式消费任务
func (q *RedisMessageQueue) ConsumeBlocking(ctx context.Context, queueName string, timeout time.Duration) (*Task, error) {
	// 首先尝试非阻塞消费
	task, err := q.Consume(ctx, queueName)
	if err != nil {
		return nil, err
	}
	if task != nil {
		return task, nil
	}

	// 如果没有任务，使用阻塞方式等待
	queueKey := q.getQueueKey(queueName)
	
	result, err := q.client.BZPopMax(ctx, timeout, queueKey).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to blocking pop task: %w", err)
	}

	taskID := result.Member.(string)
	
	// 获取任务详情并移动到处理中队列
	return q.getTaskAndMarkProcessing(ctx, queueName, taskID)
}

// processDelayedTasks 处理延迟任务
func (q *RedisMessageQueue) processDelayedTasks(ctx context.Context, queueName string) error {
	delayedKey := q.getDelayedKey(queueName)
	queueKey := q.getQueueKey(queueName)
	
	now := float64(time.Now().Unix())
	
	// 获取到期的延迟任务
	result, err := q.client.ZRangeByScoreWithScores(ctx, delayedKey, &redis.ZRangeBy{
		Min: "0",
		Max: strconv.FormatFloat(now, 'f', 0, 64),
	}).Result()
	
	if err != nil {
		return fmt.Errorf("failed to get delayed tasks: %w", err)
	}

	if len(result) == 0 {
		return nil
	}

	pipe := q.client.Pipeline()
	
	for _, z := range result {
		taskID := z.Member.(string)
		
		// 获取任务详情以确定优先级
		taskKey := q.getTaskKey(taskID)
		taskData, err := q.client.Get(ctx, taskKey).Result()
		if err != nil {
			continue
		}
		
		var task Task
		if err := json.Unmarshal([]byte(taskData), &task); err != nil {
			continue
		}
		
		// 移动到主队列
		pipe.ZAdd(ctx, queueKey, redis.Z{
			Score:  float64(task.Priority),
			Member: taskID,
		})
		
		// 从延迟队列移除
		pipe.ZRem(ctx, delayedKey, taskID)
	}

	_, err = pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to move delayed tasks: %w", err)
	}

	logger.L.Debug("Processed delayed tasks",
		zap.String("queue", queueName),
		zap.Int("count", len(result)),
	)

	return nil
}

// getTaskAndMarkProcessing 获取任务并标记为处理中
func (q *RedisMessageQueue) getTaskAndMarkProcessing(ctx context.Context, queueName, taskID string) (*Task, error) {
	taskKey := q.getTaskKey(taskID)
	processingKey := q.getProcessingKey(queueName)
	
	taskData, err := q.client.Get(ctx, taskKey).Bytes()
	if err != nil {
		return nil, fmt.Errorf("failed to get task data: %w", err)
	}

	var task Task
	if err := task.FromJSON(taskData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal task: %w", err)
	}

	// 移动到处理中队列并更新状态
	pipe := q.client.Pipeline()
	pipe.ZAdd(ctx, processingKey, redis.Z{
		Score:  float64(time.Now().Unix()),
		Member: taskID,
	})
	
	task.MarkProcessing("")
	updatedData, _ := task.ToJSON()
	pipe.Set(ctx, taskKey, updatedData, 24*time.Hour)
	
	_, err = pipe.Exec(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to mark task as processing: %w", err)
	}

	return &task, nil
}

// Ack 确认任务完成
func (q *RedisMessageQueue) Ack(ctx context.Context, task *Task, result *TaskResult) error {
	processingKey := q.getProcessingKey(result.TaskID)
	taskKey := q.getTaskKey(task.ID)
	resultKey := q.getResultKey(task.ID)

	// 更新任务状态
	task.MarkCompleted()
	taskData, err := task.ToJSON()
	if err != nil {
		return fmt.Errorf("failed to marshal task: %w", err)
	}

	// 保存结果
	resultData, err := json.Marshal(result)
	if err != nil {
		return fmt.Errorf("failed to marshal result: %w", err)
	}

	pipe := q.client.Pipeline()
	
	// 从处理中队列移除
	pipe.ZRem(ctx, processingKey, task.ID)
	
	// 更新任务状态
	pipe.Set(ctx, taskKey, taskData, 24*time.Hour)
	
	// 保存结果
	pipe.Set(ctx, resultKey, resultData, 24*time.Hour)

	_, err = pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to ack task: %w", err)
	}

	logger.L.Debug("Task acknowledged",
		zap.String("task_id", task.ID),
		zap.String("status", string(result.Status)),
		zap.Duration("duration", result.Duration),
	)

	return nil
}

// Nack 拒绝任务
func (q *RedisMessageQueue) Nack(ctx context.Context, task *Task, requeue bool) error {
	processingKey := q.getProcessingKey(task.Type.String())
	taskKey := q.getTaskKey(task.ID)

	pipe := q.client.Pipeline()
	
	// 从处理中队列移除
	pipe.ZRem(ctx, processingKey, task.ID)

	if requeue && task.ShouldRetry() {
		// 重新入队
		task.IncrementRetry()
		
		// 延迟重试
		delayedKey := q.getDelayedKey(string(task.Type))
		retryTime := time.Now().Add(q.retryDelay * time.Duration(task.Retry))
		
		pipe.ZAdd(ctx, delayedKey, redis.Z{
			Score:  float64(retryTime.Unix()),
			Member: task.ID,
		})
		
		logger.L.Debug("Task requeued for retry",
			zap.String("task_id", task.ID),
			zap.Int("retry", task.Retry),
			zap.Time("retry_at", retryTime),
		)
	} else {
		// 移动到死信队列
		deadLetterKey := q.getDeadLetterKey(string(task.Type))
		pipe.ZAdd(ctx, deadLetterKey, redis.Z{
			Score:  float64(time.Now().Unix()),
			Member: task.ID,
		})
		
		task.MarkFailed("Max retries exceeded or requeue disabled")
		
		logger.L.Warn("Task moved to dead letter queue",
			zap.String("task_id", task.ID),
			zap.Int("retry", task.Retry),
		)
	}

	// 更新任务状态
	taskData, _ := task.ToJSON()
	pipe.Set(ctx, taskKey, taskData, 24*time.Hour)

	_, err := pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to nack task: %w", err)
	}

	return nil
}

// GetQueueStats 获取队列统计信息
func (q *RedisMessageQueue) GetQueueStats(ctx context.Context, queueName string) (*QueueStats, error) {
	queueKey := q.getQueueKey(queueName)
	processingKey := q.getProcessingKey(queueName)
	deadLetterKey := q.getDeadLetterKey(queueName)

	pipe := q.client.Pipeline()
	pendingCmd := pipe.ZCard(ctx, queueKey)
	processingCmd := pipe.ZCard(ctx, processingKey)
	deadLetterCmd := pipe.ZCard(ctx, deadLetterKey)

	_, err := pipe.Exec(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get queue stats: %w", err)
	}

	stats := &QueueStats{
		QueueName:       queueName,
		PendingCount:    pendingCmd.Val(),
		ProcessingCount: processingCmd.Val(),
		DeadLetterCount: deadLetterCmd.Val(),
	}

	return stats, nil
}

// GetTaskStatus 获取任务状态
func (q *RedisMessageQueue) GetTaskStatus(ctx context.Context, taskID string) (*Task, error) {
	taskKey := q.getTaskKey(taskID)
	
	taskData, err := q.client.Get(ctx, taskKey).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get task: %w", err)
	}

	var task Task
	if err := task.FromJSON(taskData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal task: %w", err)
	}

	return &task, nil
}

// Close 关闭队列连接
func (q *RedisMessageQueue) Close() error {
	return q.client.Close()
}
