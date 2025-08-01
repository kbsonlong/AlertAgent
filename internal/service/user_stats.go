package service

import (
	"context"
	"fmt"

	"alert_agent/internal/pkg/database"
	"alert_agent/internal/model"
	"alert_agent/internal/pkg/logger"

	"go.uber.org/zap"
)

// UserStatsService 用户统计服务
type UserStatsService struct {
	cacheService *CacheService
}

// NewUserStatsService 创建用户统计服务实例
func NewUserStatsService() *UserStatsService {
	return &UserStatsService{
		cacheService: NewCacheService(),
	}
}

// GetUserStats 获取用户统计数据（带缓存）
func (s *UserStatsService) GetUserStats(ctx context.Context) (*model.UserStats, error) {
	var stats model.UserStats

	err := s.cacheService.GetOrSet(ctx, UserStatsKey, &stats, func() (interface{}, error) {
		return s.fetchUserStatsFromDB(ctx)
	})

	if err != nil {
		return nil, fmt.Errorf("failed to get user stats: %w", err)
	}

	return &stats, nil
}

// fetchUserStatsFromDB 从数据库获取用户统计数据
func (s *UserStatsService) fetchUserStatsFromDB(ctx context.Context) (*model.UserStats, error) {
	logger.L.Debug("Fetching user stats from database")

	stats := &model.UserStats{}

	// 获取总用户数
	var totalCount int64
	if err := database.DB.Model(&model.User{}).Count(&totalCount).Error; err != nil {
		logger.L.Error("Failed to count total users", zap.Error(err))
		return nil, fmt.Errorf("failed to count total users: %w", err)
	}
	stats.Total = totalCount

	// 获取活跃用户数（状态为active）
	var activeCount int64
	if err := database.DB.Model(&model.User{}).Where("status = ?", model.UserStatusActive).Count(&activeCount).Error; err != nil {
		logger.L.Error("Failed to count active users", zap.Error(err))
		return nil, fmt.Errorf("failed to count active users: %w", err)
	}
	stats.Active = activeCount

	// 获取非活跃用户数（状态为inactive）
	var inactiveCount int64
	if err := database.DB.Model(&model.User{}).Where("status = ?", model.UserStatusInactive).Count(&inactiveCount).Error; err != nil {
		logger.L.Error("Failed to count inactive users", zap.Error(err))
		return nil, fmt.Errorf("failed to count inactive users: %w", err)
	}
	stats.Inactive = inactiveCount

	// 获取锁定用户数（状态为locked）
	var lockedCount int64
	if err := database.DB.Model(&model.User{}).Where("status = ?", model.UserStatusLocked).Count(&lockedCount).Error; err != nil {
		logger.L.Error("Failed to count locked users", zap.Error(err))
		return nil, fmt.Errorf("failed to count locked users: %w", err)
	}
	stats.Locked = lockedCount



	logger.L.Debug("User stats fetched from database",
		zap.Int64("total", stats.Total),
		zap.Int64("active", stats.Active),
		zap.Int64("inactive", stats.Inactive),
		zap.Int64("locked", stats.Locked),
	)

	return stats, nil
}

// RefreshUserStats 刷新用户统计缓存
func (s *UserStatsService) RefreshUserStats(ctx context.Context) error {
	logger.L.Info("Refreshing user stats cache")
	return s.cacheService.InvalidateCache(ctx, UserStatsKey)
}

// InvalidateUserStatsCache 使用户统计缓存失效
func (s *UserStatsService) InvalidateUserStatsCache(ctx context.Context) error {
	return s.cacheService.InvalidateCache(ctx, UserStatsKey)
}