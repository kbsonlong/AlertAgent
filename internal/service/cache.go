package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"alert_agent/internal/pkg/redis"
	"alert_agent/internal/pkg/logger"

	"go.uber.org/zap"
)

// CacheService 缓存服务
type CacheService struct {
	defaultTTL time.Duration
}

// NewCacheService 创建缓存服务实例
func NewCacheService() *CacheService {
	return &CacheService{
		defaultTTL: 5 * time.Minute, // 默认缓存5分钟
	}
}

// CacheKey 缓存键常量
const (
	UserStatsKey  = "stats:users"
	AlertStatsKey = "stats:alerts"
	RuleStatsKey  = "stats:rules"
)

// GetOrSet 获取缓存数据，如果不存在则执行回调函数获取数据并缓存
func (c *CacheService) GetOrSet(ctx context.Context, key string, data interface{}, fetchFunc func() (interface{}, error)) error {
	// 尝试从缓存获取
	cachedData, err := redis.Get(ctx, key)
	if err != nil {
		logger.L.Error("Failed to get cache", zap.String("key", key), zap.Error(err))
	} else if cachedData != "" {
		// 缓存命中，反序列化数据
		if err := json.Unmarshal([]byte(cachedData), data); err != nil {
			logger.L.Error("Failed to unmarshal cached data", zap.String("key", key), zap.Error(err))
		} else {
			logger.L.Debug("Cache hit", zap.String("key", key))
			return nil
		}
	}

	// 缓存未命中或反序列化失败，执行回调函数获取数据
	logger.L.Debug("Cache miss, fetching data", zap.String("key", key))
	freshData, err := fetchFunc()
	if err != nil {
		return fmt.Errorf("failed to fetch data: %w", err)
	}

	// 将获取的数据赋值给传入的指针
	if err := c.copyData(freshData, data); err != nil {
		return fmt.Errorf("failed to copy data: %w", err)
	}

	// 序列化并缓存数据
	cachedBytes, err := json.Marshal(freshData)
	if err != nil {
		logger.L.Error("Failed to marshal data for cache", zap.String("key", key), zap.Error(err))
		return nil // 不影响主流程
	}

	if err := redis.Set(ctx, key, string(cachedBytes), c.defaultTTL); err != nil {
		logger.L.Error("Failed to set cache", zap.String("key", key), zap.Error(err))
	} else {
		logger.L.Debug("Data cached successfully", zap.String("key", key), zap.Duration("ttl", c.defaultTTL))
	}

	return nil
}

// copyData 将源数据复制到目标指针
func (c *CacheService) copyData(src, dst interface{}) error {
	// 通过JSON序列化和反序列化来复制数据
	data, err := json.Marshal(src)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, dst)
}

// InvalidateCache 使缓存失效
func (c *CacheService) InvalidateCache(ctx context.Context, key string) error {
	if err := redis.Del(ctx, key); err != nil {
		logger.L.Error("Failed to invalidate cache", zap.String("key", key), zap.Error(err))
		return err
	}
	logger.L.Debug("Cache invalidated", zap.String("key", key))
	return nil
}

// SetTTL 设置缓存TTL
func (c *CacheService) SetTTL(ttl time.Duration) {
	c.defaultTTL = ttl
}

// GetTTL 获取缓存剩余时间
func (c *CacheService) GetTTL(ctx context.Context, key string) (time.Duration, error) {
	return redis.TTL(ctx, key)
}

// RefreshCache 刷新缓存（删除后重新获取）
func (c *CacheService) RefreshCache(ctx context.Context, key string, data interface{}, fetchFunc func() (interface{}, error)) error {
	// 先删除缓存
	if err := c.InvalidateCache(ctx, key); err != nil {
		return err
	}
	// 重新获取并缓存
	return c.GetOrSet(ctx, key, data, fetchFunc)
}

// BatchInvalidate 批量使缓存失效
func (c *CacheService) BatchInvalidate(ctx context.Context, keys []string) error {
	for _, key := range keys {
		if err := c.InvalidateCache(ctx, key); err != nil {
			logger.L.Error("Failed to invalidate cache in batch", zap.String("key", key), zap.Error(err))
			// 继续处理其他键，不中断
		}
	}
	return nil
}