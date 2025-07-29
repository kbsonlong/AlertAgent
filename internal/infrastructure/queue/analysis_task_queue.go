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

// Redis键名常量
const (
	AnalysisTaskQueueKey     = "analysis:queue:tasks"          // 普通任务队列
	AnalysisPriorityQueueKey = "analysis:queue:priority"       // 优先级任务队列
	AnalysisProcessingSetKey = "analysis:processing"            // 处理中任务集合
	AnalysisTaskDataKey      = "analysis:task:%s"              // 任务数据键模板
	AnalysisQueueStatsKey    = "analysis:queue:stats"          // 队列统计信息
)

// AnalysisTaskQueueImpl Redis分析任务队列实现
type AnalysisTaskQueueImpl struct {
	client *redis.Client
	logger *zap.Logger
}

// NewAnalysisTaskQueue 创建分析任务队列实例
func NewAnalysisTaskQueue(client *redis.Client) analysis.AnalysisTaskQueue {
	return &AnalysisTaskQueueImpl{
		client: client,
		logger: logger.L.Named("analysis-task-queue"),
	}
}

// Push 推送任务到队列
func (q *AnalysisTaskQueueImpl) Push(ctx context.Context, task *analysis.AnalysisTask) error {
	if task == nil {
		return fmt.Errorf("task cannot be nil")
	}

	// 序列化任务数据
	taskData, err := json.Marshal(task)
	if err != nil {
		q.logger.Error("Failed to marshal task", 
			zap.String("task_id", task.ID),
			zap.Error(err))
		return fmt.Errorf("failed to marshal task: %w", err)
	}

	// 使用事务确保原子性
	pipe := q.client.TxPipeline()

	// 保存任务数据
	taskKey := fmt.Sprintf(AnalysisTaskDataKey, task.ID)
	pipe.Set(ctx, taskKey, taskData, 24*time.Hour) // 24小时过期

	// 根据优先级决定队列
	if task.Priority > 0 {
		// 高优先级任务放入优先级队列
		pipe.ZAdd(ctx, AnalysisPriorityQueueKey, redis.Z{
			Score:  float64(task.Priority),
			Member: task.ID,
		})
	} else {
		// 普通任务放入普通队列
		pipe.LPush(ctx, AnalysisTaskQueueKey, task.ID)
	}

	// 更新统计信息
	pipe.HIncrBy(ctx, AnalysisQueueStatsKey, "total_pushed", 1)
	pipe.HSet(ctx, AnalysisQueueStatsKey, "last_push_time", time.Now().Unix())

	if _, err := pipe.Exec(ctx); err != nil {
		q.logger.Error("Failed to push task", 
			zap.String("task_id", task.ID),
			zap.Error(err))
		return fmt.Errorf("failed to push task: %w", err)
	}

	q.logger.Debug("Task pushed to queue", 
		zap.String("task_id", task.ID),
		zap.Int("priority", task.Priority))
	return nil
}

// Pop 从队列中获取任务
func (q *AnalysisTaskQueueImpl) Pop(ctx context.Context) (*analysis.AnalysisTask, error) {
	return q.PopWithTimeout(ctx, 0)
}

// PopWithTimeout 带超时的获取任务
func (q *AnalysisTaskQueueImpl) PopWithTimeout(ctx context.Context, timeout time.Duration) (*analysis.AnalysisTask, error) {
	var taskID string
	var err error

	// 首先尝试从优先级队列获取
	result := q.client.ZPopMax(ctx, AnalysisPriorityQueueKey, 1)
	if result.Err() == nil && len(result.Val()) > 0 {
		taskID = result.Val()[0].Member.(string)
	} else {
		// 从普通队列获取
		if timeout > 0 {
			// 阻塞式获取
			result := q.client.BRPop(ctx, timeout, AnalysisTaskQueueKey)
			if result.Err() != nil {
				if result.Err() == redis.Nil {
					return nil, nil // 超时，无任务
				}
				return nil, fmt.Errorf("failed to pop task: %w", result.Err())
			}
			taskID = result.Val()[1] // BRPop返回[key, value]
		} else {
			// 非阻塞式获取
			result := q.client.RPop(ctx, AnalysisTaskQueueKey)
			if result.Err() != nil {
				if result.Err() == redis.Nil {
					return nil, nil // 队列为空
				}
				return nil, fmt.Errorf("failed to pop task: %w", result.Err())
			}
			taskID = result.Val()
		}
	}

	if taskID == "" {
		return nil, nil // 无任务
	}

	// 获取任务数据
	task, err := q.getTaskData(ctx, taskID)
	if err != nil {
		return nil, err
	}

	// 将任务添加到处理集合
	pipe := q.client.TxPipeline()
	pipe.SAdd(ctx, AnalysisProcessingSetKey, taskID)
	pipe.HIncrBy(ctx, AnalysisQueueStatsKey, "total_popped", 1)
	pipe.HSet(ctx, AnalysisQueueStatsKey, "last_pop_time", time.Now().Unix())

	if _, err := pipe.Exec(ctx); err != nil {
		q.logger.Error("Failed to update processing state", 
			zap.String("task_id", taskID),
			zap.Error(err))
		// 不返回错误，因为任务已经获取
	}

	q.logger.Debug("Task popped from queue", zap.String("task_id", taskID))
	return task, nil
}

// Peek 查看队列头部任务但不移除
func (q *AnalysisTaskQueueImpl) Peek(ctx context.Context) (*analysis.AnalysisTask, error) {
	var taskID string

	// 首先查看优先级队列
	result := q.client.ZRevRange(ctx, AnalysisPriorityQueueKey, 0, 0)
	if result.Err() == nil && len(result.Val()) > 0 {
		taskID = result.Val()[0]
	} else {
		// 查看普通队列
		result := q.client.LIndex(ctx, AnalysisTaskQueueKey, -1)
		if result.Err() != nil {
			if result.Err() == redis.Nil {
				return nil, nil // 队列为空
			}
			return nil, fmt.Errorf("failed to peek task: %w", result.Err())
		}
		taskID = result.Val()
	}

	if taskID == "" {
		return nil, nil // 队列为空
	}

	return q.getTaskData(ctx, taskID)
}

// Size 获取队列大小
func (q *AnalysisTaskQueueImpl) Size(ctx context.Context) (int64, error) {
	// 获取普通队列大小
	queueSize := q.client.LLen(ctx, AnalysisTaskQueueKey)
	if queueSize.Err() != nil {
		q.logger.Error("Failed to get queue size", zap.Error(queueSize.Err()))
		return 0, fmt.Errorf("failed to get queue size: %w", queueSize.Err())
	}

	// 获取优先级队列大小
	prioritySize := q.client.ZCard(ctx, AnalysisPriorityQueueKey)
	if prioritySize.Err() != nil {
		q.logger.Error("Failed to get priority queue size", zap.Error(prioritySize.Err()))
		return 0, fmt.Errorf("failed to get priority queue size: %w", prioritySize.Err())
	}

	// 返回两个队列的总大小
	totalSize := queueSize.Val() + prioritySize.Val()
	return totalSize, nil
}

// Clear 清空队列
func (q *AnalysisTaskQueueImpl) Clear(ctx context.Context) error {
	pipe := q.client.TxPipeline()

	// 清空所有相关键
	pipe.Del(ctx, AnalysisTaskQueueKey)
	pipe.Del(ctx, AnalysisPriorityQueueKey)
	pipe.Del(ctx, AnalysisProcessingSetKey)
	pipe.Del(ctx, AnalysisQueueStatsKey)

	// 清空所有任务数据（通过模式匹配）
	keys := q.client.Keys(ctx, fmt.Sprintf(AnalysisTaskDataKey, "*"))
	if keys.Err() == nil && len(keys.Val()) > 0 {
		pipe.Del(ctx, keys.Val()...)
	}

	if _, err := pipe.Exec(ctx); err != nil {
		q.logger.Error("Failed to clear queue", zap.Error(err))
		return fmt.Errorf("failed to clear queue: %w", err)
	}

	q.logger.Info("Queue cleared")
	return nil
}

// GetStatus 获取队列状态
func (q *AnalysisTaskQueueImpl) GetStatus(ctx context.Context) (*analysis.QueueStatus, error) {
	// 获取队列大小
	queueSize, err := q.Size(ctx)
	if err != nil {
		return nil, err
	}

	// 获取处理中的任务数量
	processingSize := q.client.SCard(ctx, AnalysisProcessingSetKey)
	if processingSize.Err() != nil {
		q.logger.Error("Failed to get processing set size", zap.Error(processingSize.Err()))
		return nil, fmt.Errorf("failed to get processing set size: %w", processingSize.Err())
	}

	// 获取统计信息
	stats := q.client.HGetAll(ctx, AnalysisQueueStatsKey)
	if stats.Err() != nil && stats.Err() != redis.Nil {
		q.logger.Error("Failed to get queue stats", zap.Error(stats.Err()))
		return nil, fmt.Errorf("failed to get queue stats: %w", stats.Err())
	}

	statsMap := stats.Val()
	totalPushed, _ := strconv.ParseInt(statsMap["total_pushed"], 10, 64)
	totalCompleted, _ := strconv.ParseInt(statsMap["total_completed"], 10, 64)
	totalFailed, _ := strconv.ParseInt(statsMap["total_failed"], 10, 64)

	// 获取最老任务时间
	var oldestTask *time.Time
	if queueSize > 0 {
		// 从普通队列获取最老任务
		oldestTaskID := q.client.LIndex(ctx, AnalysisTaskQueueKey, -1)
		if oldestTaskID.Err() == nil && oldestTaskID.Val() != "" {
			if task, err := q.getTaskData(ctx, oldestTaskID.Val()); err == nil {
				oldestTask = &task.CreatedAt
			}
		}
	}

	status := &analysis.QueueStatus{
		PendingCount:    queueSize,
		ProcessingCount: processingSize.Val(),
		CompletedCount:  totalCompleted,
		FailedCount:     totalFailed,
		TotalCount:      totalPushed,
		OldestTask:      oldestTask,
		LastUpdated:     time.Now(),
	}

	return status, nil
}

// Remove 移除指定任务
func (q *AnalysisTaskQueueImpl) Remove(ctx context.Context, taskID string) error {
	pipe := q.client.TxPipeline()

	// 从所有可能的位置移除任务
	pipe.LRem(ctx, AnalysisTaskQueueKey, 0, taskID)           // 从普通队列移除
	pipe.ZRem(ctx, AnalysisPriorityQueueKey, taskID)         // 从优先级队列移除
	pipe.SRem(ctx, AnalysisProcessingSetKey, taskID)         // 从处理集合移除

	// 删除任务数据
	taskKey := fmt.Sprintf(AnalysisTaskDataKey, taskID)
	pipe.Del(ctx, taskKey)

	// 更新统计信息
	pipe.HIncrBy(ctx, AnalysisQueueStatsKey, "total_removed", 1)

	if _, err := pipe.Exec(ctx); err != nil {
		q.logger.Error("Failed to remove task", 
			zap.String("task_id", taskID),
			zap.Error(err))
		return fmt.Errorf("failed to remove task: %w", err)
	}

	q.logger.Debug("Task removed", zap.String("task_id", taskID))
	return nil
}

// UpdatePriority 更新任务优先级
func (q *AnalysisTaskQueueImpl) UpdatePriority(ctx context.Context, taskID string, priority int) error {
	// 检查任务是否存在
	taskKey := fmt.Sprintf(AnalysisTaskDataKey, taskID)
	exists := q.client.Exists(ctx, taskKey)
	if exists.Err() != nil {
		return fmt.Errorf("failed to check task existence: %w", exists.Err())
	}
	if exists.Val() == 0 {
		return fmt.Errorf("task not found: %s", taskID)
	}

	// 获取任务数据并更新优先级
	task, err := q.getTaskData(ctx, taskID)
	if err != nil {
		return err
	}

	oldPriority := task.Priority
	task.Priority = priority

	// 重新序列化任务数据
	taskData, err := json.Marshal(task)
	if err != nil {
		return fmt.Errorf("failed to marshal updated task: %w", err)
	}

	pipe := q.client.TxPipeline()

	// 更新任务数据
	pipe.Set(ctx, taskKey, taskData, 24*time.Hour)

	// 如果任务在队列中，需要重新调整位置
	if oldPriority > 0 {
		// 从优先级队列移除
		pipe.ZRem(ctx, AnalysisPriorityQueueKey, taskID)
	} else {
		// 从普通队列移除
		pipe.LRem(ctx, AnalysisTaskQueueKey, 0, taskID)
	}

	// 根据新优先级重新添加
	if priority > 0 {
		pipe.ZAdd(ctx, AnalysisPriorityQueueKey, redis.Z{
			Score:  float64(priority),
			Member: taskID,
		})
	} else {
		pipe.LPush(ctx, AnalysisTaskQueueKey, taskID)
	}

	if _, err := pipe.Exec(ctx); err != nil {
		q.logger.Error("Failed to update task priority", 
			zap.String("task_id", taskID),
			zap.Int("old_priority", oldPriority),
			zap.Int("new_priority", priority),
			zap.Error(err))
		return fmt.Errorf("failed to update task priority: %w", err)
	}

	q.logger.Debug("Task priority updated", 
		zap.String("task_id", taskID),
		zap.Int("old_priority", oldPriority),
		zap.Int("new_priority", priority))
	return nil
}

// CompleteTask 标记任务完成（从处理集合中移除）
func (q *AnalysisTaskQueueImpl) CompleteTask(ctx context.Context, taskID string) error {
	pipe := q.client.TxPipeline()

	// 从处理集合中移除
	pipe.SRem(ctx, AnalysisProcessingSetKey, taskID)

	// 删除任务数据
	taskKey := fmt.Sprintf(AnalysisTaskDataKey, taskID)
	pipe.Del(ctx, taskKey)

	// 更新统计信息
	pipe.HIncrBy(ctx, AnalysisQueueStatsKey, "total_completed", 1)

	if _, err := pipe.Exec(ctx); err != nil {
		q.logger.Error("Failed to complete task", 
			zap.String("task_id", taskID),
			zap.Error(err))
		return fmt.Errorf("failed to complete task: %w", err)
	}

	q.logger.Debug("Task completed", zap.String("task_id", taskID))
	return nil
}

// FailTask 标记任务失败
func (q *AnalysisTaskQueueImpl) FailTask(ctx context.Context, taskID string) error {
	pipe := q.client.TxPipeline()

	// 从处理集合中移除
	pipe.SRem(ctx, AnalysisProcessingSetKey, taskID)

	// 删除任务数据
	taskKey := fmt.Sprintf(AnalysisTaskDataKey, taskID)
	pipe.Del(ctx, taskKey)

	// 更新统计信息
	pipe.HIncrBy(ctx, AnalysisQueueStatsKey, "total_failed", 1)

	if _, err := pipe.Exec(ctx); err != nil {
		q.logger.Error("Failed to mark task as failed", 
			zap.String("task_id", taskID),
			zap.Error(err))
		return fmt.Errorf("failed to mark task as failed: %w", err)
	}

	q.logger.Debug("Task marked as failed", zap.String("task_id", taskID))
	return nil
}

// RequeueTask 重新入队任务
func (q *AnalysisTaskQueueImpl) RequeueTask(ctx context.Context, taskID string) error {
	// 获取任务数据
	task, err := q.getTaskData(ctx, taskID)
	if err != nil {
		return err
	}

	// 从处理集合中移除
	q.client.SRem(ctx, AnalysisProcessingSetKey, taskID)

	// 重新推送到队列
	return q.Push(ctx, task)
}

// GetProcessingTasks 获取正在处理的任务列表
func (q *AnalysisTaskQueueImpl) GetProcessingTasks(ctx context.Context) ([]string, error) {
	result := q.client.SMembers(ctx, AnalysisProcessingSetKey)
	if result.Err() != nil {
		return nil, fmt.Errorf("failed to get processing tasks: %w", result.Err())
	}
	return result.Val(), nil
}

// getTaskData 获取任务数据
func (q *AnalysisTaskQueueImpl) getTaskData(ctx context.Context, taskID string) (*analysis.AnalysisTask, error) {
	taskKey := fmt.Sprintf(AnalysisTaskDataKey, taskID)
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