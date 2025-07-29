package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"alert_agent/internal/domain/analysis"
)

// NewAnalysisProgressTracker 创建分析进度跟踪器实例
func NewAnalysisProgressTracker(redisClient *redis.Client) analysis.AnalysisProgressTracker {
	return &AnalysisProgressTrackerImpl{
		redisClient: redisClient,
		keyPrefix:   "analysis:progress:",
		ttl:         24 * time.Hour, // 进度信息保留24小时
	}
}

// AnalysisProgressTrackerImpl 分析进度跟踪器实现
type AnalysisProgressTrackerImpl struct {
	redisClient *redis.Client
	keyPrefix   string
	ttl         time.Duration
}

// getProgressKey 获取进度存储键
func (t *AnalysisProgressTrackerImpl) getProgressKey(taskID string) string {
	return t.keyPrefix + taskID
}

// UpdateProgress 更新进度
func (t *AnalysisProgressTrackerImpl) UpdateProgress(ctx context.Context, taskID string, progress *analysis.AnalysisProgress) error {
	if progress == nil {
		return fmt.Errorf("progress cannot be nil")
	}

	// 设置任务ID
	progress.TaskID = taskID
	progress.UpdatedAt = time.Now()

	// 序列化进度数据
	progressJSON, err := json.Marshal(progress)
	if err != nil {
		return fmt.Errorf("failed to marshal progress: %w", err)
	}

	// 存储到Redis
	key := t.getProgressKey(taskID)
	if err := t.redisClient.Set(ctx, key, progressJSON, t.ttl).Err(); err != nil {
		return fmt.Errorf("failed to store progress in Redis: %w", err)
	}

	return nil
}

// GetProgress 获取进度
func (t *AnalysisProgressTrackerImpl) GetProgress(ctx context.Context, taskID string) (*analysis.AnalysisProgress, error) {
	key := t.getProgressKey(taskID)
	
	// 从Redis获取数据
	progressJSON, err := t.redisClient.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, fmt.Errorf("progress not found for task: %s", taskID)
		}
		return nil, fmt.Errorf("failed to get progress from Redis: %w", err)
	}

	// 反序列化进度数据
	var progress analysis.AnalysisProgress
	if err := json.Unmarshal([]byte(progressJSON), &progress); err != nil {
		return nil, fmt.Errorf("failed to unmarshal progress: %w", err)
	}

	return &progress, nil
}

// DeleteProgress 删除进度记录
func (t *AnalysisProgressTrackerImpl) DeleteProgress(ctx context.Context, taskID string) error {
	key := t.getProgressKey(taskID)
	
	result := t.redisClient.Del(ctx, key)
	if err := result.Err(); err != nil {
		return fmt.Errorf("failed to delete progress from Redis: %w", err)
	}

	if result.Val() == 0 {
		return fmt.Errorf("progress not found for task: %s", taskID)
	}

	return nil
}

// GetProgressByTasks 批量获取任务进度
func (t *AnalysisProgressTrackerImpl) GetProgressByTasks(ctx context.Context, taskIDs []string) (map[string]*analysis.AnalysisProgress, error) {
	if len(taskIDs) == 0 {
		return make(map[string]*analysis.AnalysisProgress), nil
	}

	// 构建所有键
	keys := make([]string, len(taskIDs))
	for i, taskID := range taskIDs {
		keys[i] = t.getProgressKey(taskID)
	}

	// 批量获取数据
	results, err := t.redisClient.MGet(ctx, keys...).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to batch get progress from Redis: %w", err)
	}

	// 解析结果
	progressMap := make(map[string]*analysis.AnalysisProgress)
	for i, result := range results {
		taskID := taskIDs[i]
		
		if result == nil {
			// 该任务没有进度记录
			continue
		}

		progressJSON, ok := result.(string)
		if !ok {
			continue
		}

		var progress analysis.AnalysisProgress
		if err := json.Unmarshal([]byte(progressJSON), &progress); err != nil {
			// 跳过无法解析的数据
			continue
		}

		progressMap[taskID] = &progress
	}

	return progressMap, nil
}

// CleanupExpiredProgress 清理过期的进度记录（可选的维护方法）
func (t *AnalysisProgressTrackerImpl) CleanupExpiredProgress(ctx context.Context) error {
	// Redis的TTL会自动清理过期数据，这里可以实现额外的清理逻辑
	// 例如扫描所有进度键并检查关联的任务状态
	
	pattern := t.keyPrefix + "*"
	iter := t.redisClient.Scan(ctx, 0, pattern, 100).Iterator()
	
	var expiredKeys []string
	for iter.Next(ctx) {
		key := iter.Val()
		
		// 检查键的TTL
		ttl := t.redisClient.TTL(ctx, key).Val()
		if ttl < 0 {
			// TTL为-1表示没有设置过期时间，TTL为-2表示键不存在
			expiredKeys = append(expiredKeys, key)
		}
	}
	
	if err := iter.Err(); err != nil {
		return fmt.Errorf("failed to scan progress keys: %w", err)
	}
	
	// 删除没有TTL的键（可能是旧数据）
	if len(expiredKeys) > 0 {
		if err := t.redisClient.Del(ctx, expiredKeys...).Err(); err != nil {
			return fmt.Errorf("failed to delete expired progress keys: %w", err)
		}
	}
	
	return nil
}

// GetAllActiveProgress 获取所有活跃的进度记录（可选的查询方法）
func (t *AnalysisProgressTrackerImpl) GetAllActiveProgress(ctx context.Context) (map[string]*analysis.AnalysisProgress, error) {
	pattern := t.keyPrefix + "*"
	keys, err := t.redisClient.Keys(ctx, pattern).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get progress keys: %w", err)
	}
	
	if len(keys) == 0 {
		return make(map[string]*analysis.AnalysisProgress), nil
	}
	
	// 批量获取数据
	results, err := t.redisClient.MGet(ctx, keys...).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to batch get all progress from Redis: %w", err)
	}
	
	// 解析结果
	progressMap := make(map[string]*analysis.AnalysisProgress)
	for i, result := range results {
		if result == nil {
			continue
		}
		
		progressJSON, ok := result.(string)
		if !ok {
			continue
		}
		
		var progress analysis.AnalysisProgress
		if err := json.Unmarshal([]byte(progressJSON), &progress); err != nil {
			continue
		}
		
		// 从键中提取任务ID
		key := keys[i]
		taskID := key[len(t.keyPrefix):]
		progressMap[taskID] = &progress
	}
	
	return progressMap, nil
}