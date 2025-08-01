package service

import (
	"context"
	"fmt"

	"alert_agent/internal/model"
	"alert_agent/internal/pkg/database"
	"alert_agent/internal/pkg/logger"

	"go.uber.org/zap"
)

// RuleStatsService 规则统计服务
type RuleStatsService struct {
	cacheService *CacheService
}

// NewRuleStatsService 创建规则统计服务实例
func NewRuleStatsService() *RuleStatsService {
	return &RuleStatsService{
		cacheService: NewCacheService(),
	}
}

// RuleStats 规则统计数据结构
type RuleStats struct {
	Total    int64 `json:"total"`
	Enabled  int64 `json:"enabled"`
	Disabled int64 `json:"disabled"`
	ByLevel  map[string]int64 `json:"by_level"`
}

// GetRuleStats 获取规则统计信息
func (rss *RuleStatsService) GetRuleStats(ctx context.Context) (*RuleStats, error) {
	cacheKey := RuleStatsKey

	// 使用缓存服务获取或设置数据
	var stats RuleStats
	err := rss.cacheService.GetOrSet(ctx, cacheKey, &stats, func() (interface{}, error) {
		return rss.fetchRuleStatsFromDB(ctx)
	})

	if err != nil {
		return nil, err
	}

	return &stats, nil
}

// fetchRuleStatsFromDB 从数据库获取规则统计信息
func (rss *RuleStatsService) fetchRuleStatsFromDB(ctx context.Context) (RuleStats, error) {
	var stats RuleStats

	// 获取总规则数
	var totalCount int64
	if err := database.DB.Model(&model.Rule{}).Count(&totalCount).Error; err != nil {
		logger.L.Error("Failed to count total rules", zap.Error(err))
		return stats, fmt.Errorf("failed to count total rules: %w", err)
	}
	stats.Total = totalCount

	// 获取活跃规则数（状态为active）
	var activeCount int64
	if err := database.DB.Model(&model.Rule{}).Where("status = ?", "active").Count(&activeCount).Error; err != nil {
		logger.L.Error("Failed to count active rules", zap.Error(err))
		return stats, fmt.Errorf("failed to count active rules: %w", err)
	}
	stats.Enabled = activeCount

	// 获取非活跃规则数
	stats.Disabled = totalCount - activeCount

	// 按严重级别统计
	stats.ByLevel = make(map[string]int64)
	levelStats := []struct {
		Severity string
		Count    int64
	}{}

	if err := database.DB.Model(&model.Rule{}).
		Select("severity, COUNT(*) as count").
		Group("severity").
		Scan(&levelStats).Error; err != nil {
		logger.L.Error("Failed to get rule severity statistics", zap.Error(err))
		return stats, fmt.Errorf("failed to get rule severity statistics: %w", err)
	}

	for _, stat := range levelStats {
		stats.ByLevel[stat.Severity] = stat.Count
	}

	logger.L.Debug("Rule stats fetched from database",
		zap.Int64("total", stats.Total),
		zap.Int64("enabled", stats.Enabled),
		zap.Int64("disabled", stats.Disabled),
	)

	return stats, nil
}

// InvalidateRuleStatsCache 清除规则统计缓存
func (rss *RuleStatsService) InvalidateRuleStatsCache(ctx context.Context) error {
	cacheKey := RuleStatsKey
	return rss.cacheService.InvalidateCache(ctx, cacheKey)
}