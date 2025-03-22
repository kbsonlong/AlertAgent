package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"alert_agent/internal/pkg/types"

	"github.com/redis/go-redis/v9"
)

// Queue 队列接口
type Queue interface {
	// Push 推送任务到队列
	Push(ctx context.Context, task *types.AlertTask) error
	// PushBatch 批量推送任务到队列
	PushBatch(ctx context.Context, tasks []*types.AlertTask) error
	// Pop 从队列中获取任务
	Pop(ctx context.Context) (*types.AlertTask, error)
	// Complete 标记任务完成
	Complete(ctx context.Context, result *types.AlertResult) error
	// GetResult 获取任务结果
	GetResult(ctx context.Context, taskID uint) (*types.AlertResult, error)
	// Close 关闭队列
	Close() error
}

// RedisQueue Redis队列实现
type RedisQueue struct {
	client *redis.Client
	key    string
}

// NewRedisQueue 创建Redis队列
func NewRedisQueue(client *redis.Client, key string) *RedisQueue {
	return &RedisQueue{
		client: client,
		key:    key,
	}
}

// Push 推送任务到队列
func (q *RedisQueue) Push(ctx context.Context, task *types.AlertTask) error {
	data, err := json.Marshal(task)
	if err != nil {
		return fmt.Errorf("failed to marshal task: %w", err)
	}

	if err := q.client.RPush(ctx, q.key, string(data)).Err(); err != nil {
		return fmt.Errorf("failed to push task: %w", err)
	}

	return nil
}

// PushBatch 批量推送任务到队列
func (q *RedisQueue) PushBatch(ctx context.Context, tasks []*types.AlertTask) error {
	pipe := q.client.Pipeline()
	for _, task := range tasks {
		data, err := json.Marshal(task)
		if err != nil {
			return fmt.Errorf("failed to marshal task: %w", err)
		}
		fmt.Println("data", string(q.key))
		pipe.RPush(ctx, q.key, string(data))
	}
	if _, err := pipe.Exec(ctx); err != nil {
		return fmt.Errorf("failed to push tasks: %w", err)
	}
	return nil
}

// Pop 从队列中获取任务
func (q *RedisQueue) Pop(ctx context.Context) (*types.AlertTask, error) {
	data, err := q.client.LPop(ctx, q.key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to pop task: %w", err)
	}

	var task types.AlertTask
	if err := json.Unmarshal(data, &task); err != nil {
		return nil, fmt.Errorf("failed to unmarshal task: %w", err)
	}

	return &task, nil
}

// Complete 标记任务完成
func (q *RedisQueue) Complete(ctx context.Context, result *types.AlertResult) error {
	data, err := json.Marshal(result)
	if err != nil {
		return fmt.Errorf("failed to marshal result: %w", err)
	}

	key := fmt.Sprintf("alert:result:%d", result.TaskID)
	if err := q.client.Set(ctx, key, string(data), 24*time.Hour).Err(); err != nil {
		return fmt.Errorf("failed to save result: %w", err)
	}

	return nil
}

// GetResult 获取任务结果
func (q *RedisQueue) GetResult(ctx context.Context, taskID uint) (*types.AlertResult, error) {
	key := fmt.Sprintf("alert:result:%d", taskID)
	data, err := q.client.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get result: %w", err)
	}

	var result types.AlertResult
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal result: %w", err)
	}

	return &result, nil
}

// Close 关闭队列
func (q *RedisQueue) Close() error {
	return q.client.Close()
}
