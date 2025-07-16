package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"alert_agent/internal/model"

	goredis "github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

var (
	ErrProviderNotFound = errors.New("provider not found")
	ErrInvalidProvider  = errors.New("invalid provider data")
)

const (
	ProviderCacheKeyPrefix = "provider:"
	ProviderCacheTTL       = 24 * time.Hour
)

// ProviderService 数据源服务
type ProviderService struct {
	db    *gorm.DB
	cache *goredis.Client
}

// NewProviderService 创建数据源服务实例
func NewProviderService(db *gorm.DB, cache *goredis.Client) *ProviderService {
	return &ProviderService{
		db:    db,
		cache: cache,
	}
}

// CreateProvider 创建数据源
func (s *ProviderService) CreateProvider(ctx context.Context, provider *model.Provider) error {
	// 验证数据源配置
	if err := provider.Validate(); err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidProvider, err)
	}

	// 设置初始状态
	provider.Status = model.ProviderStatusActive

	// 开启事务
	tx := s.db.Begin()
	if err := tx.Create(provider).Error; err != nil {
		tx.Rollback()
		return err
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		return err
	}

	// 更新缓存
	cacheKey := fmt.Sprintf("%s%d", ProviderCacheKeyPrefix, provider.ID)
	if err := s.cache.Set(ctx, cacheKey, provider, ProviderCacheTTL).Err(); err != nil {
		return err
	}

	return nil
}

// GetProvider 获取数据源
func (s *ProviderService) GetProvider(ctx context.Context, id uint) (*model.Provider, error) {
	// 尝试从缓存获取
	cacheKey := fmt.Sprintf("%s%d", ProviderCacheKeyPrefix, id)
	var provider model.Provider
	if err := s.cache.Get(ctx, cacheKey).Scan(&provider); err == nil {
		return &provider, nil
	}

	// 从数据库获取
	if err := s.db.First(&provider, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrProviderNotFound
		}
		return nil, err
	}

	// 更新缓存
	if err := s.cache.Set(ctx, cacheKey, &provider, ProviderCacheTTL).Err(); err != nil {
		return nil, err
	}

	return &provider, nil
}

// ListProviders 获取数据源列表
func (s *ProviderService) ListProviders(ctx context.Context) ([]*model.Provider, error) {
	var providers []*model.Provider
	if err := s.db.Find(&providers).Error; err != nil {
		return nil, err
	}
	return providers, nil
}

// UpdateProvider 更新数据源
func (s *ProviderService) UpdateProvider(ctx context.Context, provider *model.Provider) error {
	// 验证数据源配置
	if err := provider.Validate(); err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidProvider, err)
	}

	// 开启事务
	tx := s.db.Begin()
	if err := tx.Save(provider).Error; err != nil {
		tx.Rollback()
		return err
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		return err
	}

	// 更新缓存
	cacheKey := fmt.Sprintf("%s%d", ProviderCacheKeyPrefix, provider.ID)
	if err := s.cache.Set(ctx, cacheKey, provider, ProviderCacheTTL).Err(); err != nil {
		return err
	}

	return nil
}

// DeleteProvider 删除数据源
func (s *ProviderService) DeleteProvider(ctx context.Context, id uint) error {
	// 开启事务
	tx := s.db.Begin()
	if err := tx.Delete(&model.Provider{}, id).Error; err != nil {
		tx.Rollback()
		return err
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		return err
	}

	// 删除缓存
	cacheKey := fmt.Sprintf("%s%d", ProviderCacheKeyPrefix, id)
	if err := s.cache.Del(ctx, cacheKey).Err(); err != nil {
		return err
	}

	return nil
}
