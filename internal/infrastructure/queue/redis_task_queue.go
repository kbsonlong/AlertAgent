package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"alert_agent/internal/domain/analysis"
	"alert_agent/internal/pkg/logger"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

const (
	// Redis 键前缀
	TaskQueueKey    = "analysis:queue:tasks"
	PriorityQueueKey = "analysis:queue:priority"
	ProcessingSetKey = "analysis:processing"
	TaskDataKey     = "analysis:task:%s"
	QueueStatsKey   = "analysis:stats"
)

// RedisTaskQueue Redis 任务队列实现
type RedisTaskQueue struct {
	client *redis.Client
	logger *zap.Logger
}

// NewRedisTaskQueue 创建 Redis 任务队列
func NewRedisTaskQueue(client *redis.Client) analysis.AnalysisTaskQueue {
	return &RedisTaskQueue{
		client: client,
		logger: logger.L.Named("redis-task-queue"),
	}
}

// Push 推送任务到队列
func (q *RedisTaskQueue) Push(ctx context.Context, task *analysis.AnalysisTask) error {
	// 序列化任务数据
	taskData, err := json.Marshal(task)
	if err != nil {
		q.logger.Error("Failed to marshal task", 
			zap.String("task_id", task.ID),
			zap.Error(err))
		return fmt.Errorf("failed to marshal task: %w", err)
	}

	// 使用 Redis 事务确保原子性
	pipe := q.client.TxPipeline()

	// 存储任务数据
	taskKey := fmt.Sprintf(TaskDataKey, task.ID)
	pipe.Set(ctx, taskKey, taskData, 24*time.Hour) // 24小时过期

	// 添加到优先级队列
	pipe.ZAdd(ctx, PriorityQueueKey, redis.Z{
		Score:  float64(task.Priority),
		Member: task.ID,
	})

	// 添加到普通队列（FIFO）
	pipe.LPush(ctx, TaskQueueKey, task.ID)

	// 更新统计信息
	pipe.HIncrBy(ctx, QueueStatsKey, "total_pushed", 1)
	pipe.HIncrBy(ctx, QueueStatsKey, fmt.Sprintf("type_%s", task.Type), 1)

	// 执行事务
	if _, err := pipe.Exec(ctx); err != nil {
		q.logger.Error("Failed to push task to queue", 
			zap.String("task_id", task.ID),
			zap.Error(err))
		return fmt.Errorf("failed to push task to queue: %w", err)
	}

	q.logger.Info("Task pushed to queue", 
		zap.String("task_id", task.ID),
		zap.String("type", string(task.Type)),
		zap.Int("priority", task.Priority))

	return nil
}

// Pop 从队列中弹出任务（优先级优先）
func (q *RedisTaskQueue) Pop(ctx context.Context) (*analysis.AnalysisTask, error) {
	return q.PopWithTimeout(ctx, 1*time.Second)
}

// PopWithTimeout 带超时的获取任务
func (q *RedisTaskQueue) PopWithTimeout(ctx context.Context, timeout time.Duration) (*analysis.AnalysisTask, error) {
	// 首先尝试从优先级队列获取高优先级任务
	result := q.client.ZPopMax(ctx, PriorityQueueKey, 1)
	if result.Err() != nil && result.Err() != redis.Nil {
		q.logger.Error("Failed to pop from priority queue", zap.Error(result.Err()))
		return nil, fmt.Errorf("failed to pop from priority queue: %w", result.Err())
	}

	var taskID string
	if len(result.Val()) > 0 {
		// 从优先级队列获取到任务
		taskID = result.Val()[0].Member.(string)
		// 从普通队列中移除该任务
		q.client.LRem(ctx, TaskQueueKey, 1, taskID)
	} else {
		// 优先级队列为空，从普通队列获取
		result := q.client.BRPop(ctx, timeout, TaskQueueKey)
		if result.Err() != nil {
			if result.Err() == redis.Nil {
				return nil, nil // 队列为空
			}
			q.logger.Error("Failed to pop from task queue", zap.Error(result.Err()))
			return nil, fmt.Errorf("failed to pop from task queue: %w", result.Err())
		}
		taskID = result.Val()[1]
	}

	if taskID == "" {
		return nil, nil // 队列为空
	}

	// 获取任务数据
	task, err := q.getTaskData(ctx, taskID)
	if err != nil {
		return nil, err
	}

	// 添加到处理集合
	q.client.SAdd(ctx, ProcessingSetKey, taskID)

	// 更新统计信息
	q.client.HIncrBy(ctx, QueueStatsKey, "total_popped", 1)

	q.logger.Info("Task popped from queue", 
		zap.String("task_id", taskID),
		zap.String("type", string(task.Type)))

	return task, nil
}

// Peek 查看队列中的下一个任务（不移除）
func (q *RedisTaskQueue) Peek(ctx context.Context) (*analysis.AnalysisTask, error) {
	// 首先检查优先级队列
	result := q.client.ZRevRange(ctx, PriorityQueueKey, 0, 0)
	if result.Err() != nil {
		q.logger.Error("Failed to peek priority queue", zap.Error(result.Err()))
		return nil, fmt.Errorf("failed to peek priority queue: %w", result.Err())
	}

	var taskID string
	if len(result.Val()) > 0 {
		taskID = result.Val()[0]
	} else {
		// 优先级队列为空，检查普通队列
		result := q.client.LIndex(ctx, TaskQueueKey, -1)
		if result.Err() != nil {
			if result.Err() == redis.Nil {
				return nil, nil // 队列为空
			}
			q.logger.Error("Failed to peek task queue", zap.Error(result.Err()))
			return nil, fmt.Errorf("failed to peek task queue: %w", result.Err())
		}
		taskID = result.Val()
	}

	if taskID == "" {
		return nil, nil // 队列为空
	}

	return q.getTaskData(ctx, taskID)
}

// Remove 从队列中移除指定任务
func (q *RedisTaskQueue) Remove(ctx context.Context, taskID string) error {
	pipe := q.client.TxPipeline()

	// 从所有队列中移除
	pipe.ZRem(ctx, PriorityQueueKey, taskID)
	pipe.LRem(ctx, TaskQueueKey, 0, taskID)
	pipe.SRem(ctx, ProcessingSetKey, taskID)

	// 删除任务数据
	taskKey := fmt.Sprintf(TaskDataKey, taskID)
	pipe.Del(ctx, taskKey)

	// 更新统计信息
	pipe.HIncrBy(ctx, QueueStatsKey, "total_removed", 1)

	if _, err := pipe.Exec(ctx); err != nil {
		q.logger.Error("Failed to remove task from queue", 
			zap.String("task_id", taskID),
			zap.Error(err))
		return fmt.Errorf("failed to remove task from queue: %w", err)
	}

	q.logger.Info("Task removed from queue", zap.String("task_id", taskID))
	return nil
}

// Size 获取队列大小
func (q *RedisTaskQueue) Size(ctx context.Context) (int64, error) {
	// 获取普通队列大小
	queueSize := q.client.LLen(ctx, TaskQueueKey)
	if queueSize.Err() != nil {
		q.logger.Error("Failed to get queue size", zap.Error(queueSize.Err()))
		return 0, fmt.Errorf("failed to get queue size: %w", queueSize.Err())
	}

	// 获取优先级队列大小
	prioritySize := q.client.ZCard(ctx, PriorityQueueKey)
	if prioritySize.Err() != nil {
		q.logger.Error("Failed to get priority queue size", zap.Error(prioritySize.Err()))
		return 0, fmt.Errorf("failed to get priority queue size: %w", prioritySize.Err())
	}

	// 返回两个队列的总大小（去重）
	totalSize := queueSize.Val()
	if prioritySize.Val() > 0 {
		// 如果优先级队列有任务，需要检查重复
		// 这里简化处理，实际应该去重
		totalSize = queueSize.Val()
	}

	return totalSize, nil
}

// IsEmpty 检查队列是否为空
func (q *RedisTaskQueue) IsEmpty(ctx context.Context) (bool, error) {
	size, err := q.Size(ctx)
	if err != nil {
		return false, err
	}
	return size == 0, nil
}

// Clear 清空队列
func (q *RedisTaskQueue) Clear(ctx context.Context) error {
	pipe := q.client.TxPipeline()

	// 获取所有任务ID
	taskIDs := q.client.LRange(ctx, TaskQueueKey, 0, -1)
	if taskIDs.Err() == nil {
		for _, taskID := range taskIDs.Val() {
			taskKey := fmt.Sprintf(TaskDataKey, taskID)
			pipe.Del(ctx, taskKey)
		}
	}

	// 清空所有队列
	pipe.Del(ctx, TaskQueueKey)
	pipe.Del(ctx, PriorityQueueKey)
	pipe.Del(ctx, ProcessingSetKey)
	pipe.Del(ctx, QueueStatsKey)

	if _, err := pipe.Exec(ctx); err != nil {
		q.logger.Error("Failed to clear queue", zap.Error(err))
		return fmt.Errorf("failed to clear queue: %w", err)
	}

	q.logger.Info("Queue cleared")
	return nil
}

// GetStatus 获取队列状态
func (q *RedisTaskQueue) GetStatus(ctx context.Context) (*analysis.QueueStatus, error) {
	// 获取队列大小
	queueSize, err := q.Size(ctx)
	if err != nil {
		return nil, err
	}

	// 获取处理中的任务数量
	processingSize := q.client.SCard(ctx, ProcessingSetKey)
	if processingSize.Err() != nil {
		q.logger.Error("Failed to get processing set size", zap.Error(processingSize.Err()))
		return nil, fmt.Errorf("failed to get processing set size: %w", processingSize.Err())
	}

	// 获取统计信息
	stats := q.client.HGetAll(ctx, QueueStatsKey)
	if stats.Err() != nil && stats.Err() != redis.Nil {
		q.logger.Error("Failed to get queue stats", zap.Error(stats.Err()))
		return nil, fmt.Errorf("failed to get queue stats: %w", stats.Err())
	}

	statsMap := stats.Val()
	totalPushed, _ := strconv.ParseInt(statsMap["total_pushed"], 10, 64)
	totalPopped, _ := strconv.ParseInt(statsMap["total_popped"], 10, 64)
	totalRemoved, _ := strconv.ParseInt(statsMap["total_removed"], 10, 64)

	status := &analysis.QueueStatus{
		PendingCount:    queueSize,
		ProcessingCount: processingSize.Val(),
		CompletedCount:  totalPopped,
		FailedCount:     totalRemoved,
		TotalCount:      totalPushed,
		LastUpdated:     time.Now(),
	}

	return status, nil
}

// CompleteTask 标记任务完成（从处理集合中移除）
func (q *RedisTaskQueue) CompleteTask(ctx context.Context, taskID string) error {
	pipe := q.client.TxPipeline()

	// 从处理集合中移除
	pipe.SRem(ctx, ProcessingSetKey, taskID)

	// 删除任务数据
	taskKey := fmt.Sprintf(TaskDataKey, taskID)
	pipe.Del(ctx, taskKey)

	// 更新统计信息
	pipe.HIncrBy(ctx, QueueStatsKey, "total_completed", 1)

	if _, err := pipe.Exec(ctx); err != nil {
		q.logger.Error("Failed to complete task", 
			zap.String("task_id", taskID),
			zap.Error(err))
		return fmt.Errorf("failed to complete task: %w", err)
	}

	q.logger.Debug("Task completed", zap.String("task_id", taskID))
	return nil
}

// RequeueTask 重新入队任务（从处理集合移回队列）
func (q *RedisTaskQueue) RequeueTask(ctx context.Context, taskID string) error {
	// 获取任务数据
	task, err := q.getTaskData(ctx, taskID)
	if err != nil {
		return fmt.Errorf("failed to get task data for requeue: %w", err)
	}

	pipe := q.client.TxPipeline()

	// 从处理集合中移除
	pipe.SRem(ctx, ProcessingSetKey, taskID)

	// 重新添加到队列
	pipe.ZAdd(ctx, PriorityQueueKey, redis.Z{
		Score:  float64(task.Priority),
		Member: taskID,
	})
	pipe.LPush(ctx, TaskQueueKey, taskID)

	// 更新统计信息
	pipe.HIncrBy(ctx, QueueStatsKey, "total_requeued", 1)

	if _, err := pipe.Exec(ctx); err != nil {
		q.logger.Error("Failed to requeue task", 
			zap.String("task_id", taskID),
			zap.Error(err))
		return fmt.Errorf("failed to requeue task: %w", err)
	}

	q.logger.Info("Task requeued", zap.String("task_id", taskID))
	return nil
}

// getTaskData 获取任务数据
func (q *RedisTaskQueue) getTaskData(ctx context.Context, taskID string) (*analysis.AnalysisTask, error) {
	taskKey := fmt.Sprintf(TaskDataKey, taskID)
	result := q.client.Get(ctx, taskKey)
	if result.Err() != nil {
		if result.Err() == redis.Nil {
			q.logger.Warn("Task data not found", zap.String("task_id", taskID))
			return nil, fmt.Errorf("task data not found: %s", taskID)
		}
		q.logger.Error("Failed to get task data", 
			zap.String("task_id", taskID),
			zap.Error(result.Err()))
		return nil, fmt.Errorf("failed to get task data: %w", result.Err())
	}

	var task analysis.AnalysisTask
	if err := json.Unmarshal([]byte(result.Val()), &task); err != nil {
		q.logger.Error("Failed to unmarshal task data", 
			zap.String("task_id", taskID),
			zap.Error(err))
		return nil, fmt.Errorf("failed to unmarshal task data: %w", err)
	}

	return &task, nil
}

// UpdatePriority 更新任务优先级
func (q *RedisTaskQueue) UpdatePriority(ctx context.Context, taskID string, priority int) error {
	// 检查任务是否在优先级队列中
	score := q.client.ZScore(ctx, PriorityQueueKey, taskID)
	if score.Err() != nil {
		if score.Err() == redis.Nil {
			// 任务不在优先级队列中，可能在普通队列或已被处理
			q.logger.Warn("Task not found in priority queue", 
				zap.String("task_id", taskID))
			return fmt.Errorf("task not found in priority queue: %s", taskID)
		}
		return fmt.Errorf("failed to check task priority: %w", score.Err())
	}

	// 更新优先级
	pipe := q.client.TxPipeline()
	pipe.ZAdd(ctx, PriorityQueueKey, redis.Z{
		Score:  float64(priority),
		Member: taskID,
	})

	// 更新任务数据中的优先级
	task, err := q.getTaskData(ctx, taskID)
	if err == nil {
		task.Priority = priority
		task.UpdatedAt = time.Now()
		taskData, err := json.Marshal(task)
		if err == nil {
			taskKey := fmt.Sprintf(TaskDataKey, taskID)
			pipe.Set(ctx, taskKey, taskData, 24*time.Hour)
		}
	}

	if _, err := pipe.Exec(ctx); err != nil {
		q.logger.Error("Failed to update task priority", 
			zap.String("task_id", taskID),
			zap.Int("priority", priority),
			zap.Error(err))
		return fmt.Errorf("failed to update task priority: %w", err)
	}

	q.logger.Info("Task priority updated", 
		zap.String("task_id", taskID),
		zap.Int("priority", priority))

	return nil
}