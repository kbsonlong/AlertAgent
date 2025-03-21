package redis

import (
	"context"
	"fmt"
	"time"

	"alert_agent/internal/config"

	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
)

var (
	Client *redis.Client
	logger = zap.L()
)

// Init 初始化Redis客户端
func Init() error {
	cfg := config.GlobalConfig.Redis

	options := &redis.Options{
		Addr:         fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Password:     cfg.Password,
		DB:           cfg.DB,
		PoolSize:     cfg.PoolSize,
		MinIdleConns: cfg.MinIdleConns,
		MaxRetries:   cfg.MaxRetries,
		DialTimeout:  time.Duration(cfg.DialTimeout) * time.Second,
	}

	Client = redis.NewClient(options)

	// 测试连接并重试
	ctx := context.Background()
	var err error
	for i := 0; i <= cfg.MaxRetries; i++ {
		if err = Client.Ping(ctx).Err(); err == nil {
			logger.Info("Successfully connected to Redis",
				zap.String("host", cfg.Host),
				zap.Int("port", cfg.Port),
			)
			return nil
		}

		if i < cfg.MaxRetries {
			retryDelay := time.Duration(i+1) * time.Second
			logger.Warn("Failed to connect to Redis, retrying...",
				zap.Error(err),
				zap.Int("attempt", i+1),
				zap.Int("maxRetries", cfg.MaxRetries),
				zap.Duration("retryDelay", retryDelay),
			)
			time.Sleep(retryDelay)
		}
	}

	return fmt.Errorf("failed to connect to Redis after %d retries: %w", cfg.MaxRetries, err)
}

// Set 设置缓存，带重试机制
func Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	var err error
	for i := 0; i <= config.GlobalConfig.Redis.MaxRetries; i++ {
		if err = Client.Set(ctx, key, value, ttl).Err(); err == nil {
			return nil
		}

		if i < config.GlobalConfig.Redis.MaxRetries {
			retryDelay := time.Duration(i+1) * time.Second
			logger.Warn("Failed to set Redis key, retrying...",
				zap.String("key", key),
				zap.Error(err),
				zap.Int("attempt", i+1),
			)
			time.Sleep(retryDelay)
		}
	}
	return fmt.Errorf("failed to set Redis key after %d retries: %w", config.GlobalConfig.Redis.MaxRetries, err)
}

// Get 获取缓存，带重试机制
func Get(ctx context.Context, key string) (string, error) {
	var (
		value string
		err   error
	)

	for i := 0; i <= config.GlobalConfig.Redis.MaxRetries; i++ {
		value, err = Client.Get(ctx, key).Result()
		if err == nil || err == redis.Nil {
			return value, err
		}

		if i < config.GlobalConfig.Redis.MaxRetries {
			retryDelay := time.Duration(i+1) * time.Second
			logger.Warn("Failed to get Redis key, retrying...",
				zap.String("key", key),
				zap.Error(err),
				zap.Int("attempt", i+1),
			)
			time.Sleep(retryDelay)
		}
	}
	return "", fmt.Errorf("failed to get Redis key after %d retries: %w", config.GlobalConfig.Redis.MaxRetries, err)
}

// Del 删除缓存，带重试机制
func Del(ctx context.Context, key string) error {
	var err error
	for i := 0; i <= config.GlobalConfig.Redis.MaxRetries; i++ {
		if err = Client.Del(ctx, key).Err(); err == nil {
			return nil
		}

		if i < config.GlobalConfig.Redis.MaxRetries {
			retryDelay := time.Duration(i+1) * time.Second
			logger.Warn("Failed to delete Redis key, retrying...",
				zap.String("key", key),
				zap.Error(err),
				zap.Int("attempt", i+1),
			)
			time.Sleep(retryDelay)
		}
	}
	return fmt.Errorf("failed to delete Redis key after %d retries: %w", config.GlobalConfig.Redis.MaxRetries, err)
}

// Exists 检查key是否存在，带重试机制
func Exists(ctx context.Context, key string) (bool, error) {
	var (
		n   int64
		err error
	)

	for i := 0; i <= config.GlobalConfig.Redis.MaxRetries; i++ {
		n, err = Client.Exists(ctx, key).Result()
		if err == nil {
			return n > 0, nil
		}

		if i < config.GlobalConfig.Redis.MaxRetries {
			retryDelay := time.Duration(i+1) * time.Second
			logger.Warn("Failed to check Redis key existence, retrying...",
				zap.String("key", key),
				zap.Error(err),
				zap.Int("attempt", i+1),
			)
			time.Sleep(retryDelay)
		}
	}
	return false, fmt.Errorf("failed to check Redis key existence after %d retries: %w", config.GlobalConfig.Redis.MaxRetries, err)
}

// Close 关闭Redis连接
func Close() error {
	if Client != nil {
		return Client.Close()
	}
	return nil
}
