package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"alert_agent/internal/pkg/redis"

	redisClient "github.com/go-redis/redis/v8"
	"go.uber.org/zap"
)

var (
	logger = zap.L()
)

// 队列名称常量
const (
	AlertAnalysisQueue      = "alert:analysis:queue"
	AlertAnalysisProcessing = "alert:analysis:processing"
	AlertAnalysisResult     = "alert:analysis:result:"
	AlertAnalysisTimeout    = 10 * time.Minute
)

// AnalysisTask 分析任务
type AnalysisTask struct {
	AlertID   uint      `json:"alert_id"`
	CreatedAt time.Time `json:"created_at"`
}

// AnalysisResult 分析结果
type AnalysisResult struct {
	AlertID   uint      `json:"alert_id"`
	Analysis  string    `json:"analysis"`
	CreatedAt time.Time `json:"created_at"`
	Error     string    `json:"error,omitempty"`
}

// EnqueueAnalysisTask 将分析任务加入队列
func EnqueueAnalysisTask(ctx context.Context, alertID uint) error {
	task := AnalysisTask{
		AlertID:   alertID,
		CreatedAt: time.Now(),
	}

	data, err := json.Marshal(task)
	if err != nil {
		return fmt.Errorf("failed to marshal analysis task: %w", err)
	}

	// 使用Redis List作为队列
	err = redis.Client.LPush(ctx, AlertAnalysisQueue, string(data)).Err()
	if err != nil {
		return fmt.Errorf("failed to enqueue analysis task: %w", err)
	}

	logger.Info("Analysis task enqueued",
		zap.Uint("alert_id", alertID),
	)

	return nil
}

// DequeueAnalysisTask 从队列获取分析任务
func DequeueAnalysisTask(ctx context.Context) (*AnalysisTask, error) {
	// 使用BRPOPLPUSH原子操作，将任务从队列移动到处理中列表
	result, err := redis.Client.BRPopLPush(ctx, AlertAnalysisQueue, AlertAnalysisProcessing, 0).Result()
	if err != nil {
		if err == redisClient.Nil {
			return nil, nil // 队列为空
		}
		return nil, fmt.Errorf("failed to dequeue analysis task: %w", err)
	}

	var task AnalysisTask
	if err := json.Unmarshal([]byte(result), &task); err != nil {
		return nil, fmt.Errorf("failed to unmarshal analysis task: %w", err)
	}

	logger.Info("Analysis task dequeued",
		zap.Uint("alert_id", task.AlertID),
	)

	return &task, nil
}

// CompleteAnalysisTask 完成分析任务
func CompleteAnalysisTask(ctx context.Context, result *AnalysisResult) error {
	// 将结果存储到Redis
	data, err := json.Marshal(result)
	if err != nil {
		return fmt.Errorf("failed to marshal analysis result: %w", err)
	}

	key := fmt.Sprintf("%s%d", AlertAnalysisResult, result.AlertID)
	err = redis.Client.Set(ctx, key, string(data), AlertAnalysisTimeout).Err()
	if err != nil {
		return fmt.Errorf("failed to store analysis result: %w", err)
	}

	// 从处理中列表移除任务
	task := AnalysisTask{
		AlertID: result.AlertID,
	}
	taskData, err := json.Marshal(task)
	if err != nil {
		return fmt.Errorf("failed to marshal analysis task: %w", err)
	}

	err = redis.Client.LRem(ctx, AlertAnalysisProcessing, 1, string(taskData)).Err()
	if err != nil {
		logger.Warn("Failed to remove task from processing list",
			zap.Error(err),
			zap.Uint("alert_id", result.AlertID),
		)
		// 继续执行，不返回错误
	}

	logger.Info("Analysis task completed",
		zap.Uint("alert_id", result.AlertID),
	)

	return nil
}

// GetAnalysisResult 获取分析结果
func GetAnalysisResult(ctx context.Context, alertID uint) (*AnalysisResult, error) {
	key := fmt.Sprintf("%s%d", AlertAnalysisResult, alertID)
	data, err := redis.Get(ctx, key)
	if err != nil {
		if err == redisClient.Nil {
			return nil, nil // 结果不存在
		}
		return nil, fmt.Errorf("failed to get analysis result: %w", err)
	}

	var result AnalysisResult
	if err := json.Unmarshal([]byte(data), &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal analysis result: %w", err)
	}

	return &result, nil
}
