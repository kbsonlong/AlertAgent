package service

import (
	"context"
	"fmt"

	"alert_agent/internal/model"
	"alert_agent/internal/pkg/database"
	"alert_agent/internal/pkg/logger"

	"go.uber.org/zap"
	"gorm.io/gorm"
)



// AlertStats 告警统计结构
type AlertStats struct {
	Total        int64                  `json:"total"`
	Firing       int64                  `json:"firing"`
	Acknowledged int64                  `json:"acknowledged"`
	Resolved     int64                  `json:"resolved"`
	ByLevel      map[string]int64       `json:"by_level"`
	BySource     map[string]int64       `json:"by_source"`
	BySeverity   map[string]int64       `json:"by_severity"`
}

// AlertStatsService 告警统计服务
type AlertStatsService struct {
	db           *gorm.DB
	cacheService *CacheService
}

// NewAlertStatsService 创建告警统计服务实例
func NewAlertStatsService(cacheService *CacheService) *AlertStatsService {
	return &AlertStatsService{
		db:           database.DB,
		cacheService: cacheService,
	}
}

// GetAlertStats 获取告警统计信息
func (s *AlertStatsService) GetAlertStats(ctx context.Context) (*AlertStats, error) {
	var stats AlertStats

	// 尝试从缓存获取，如果不存在则从数据库查询并缓存
	err := s.cacheService.GetOrSet(ctx, AlertStatsKey, &stats, func() (interface{}, error) {
		return s.fetchAlertStatsFromDB(ctx)
	})

	if err != nil {
		logger.L.Error("Failed to get alert stats", zap.Error(err))
		return nil, fmt.Errorf("failed to get alert stats: %w", err)
	}

	return &stats, nil
}

// fetchAlertStatsFromDB 从数据库获取告警统计信息
func (s *AlertStatsService) fetchAlertStatsFromDB(ctx context.Context) (*AlertStats, error) {
	stats := &AlertStats{
		ByLevel:    make(map[string]int64),
		BySource:   make(map[string]int64),
		BySeverity: make(map[string]int64),
	}

	// 获取总数
	if err := s.db.WithContext(ctx).Model(&model.Alert{}).Count(&stats.Total).Error; err != nil {
		logger.L.Error("Failed to count total alerts", zap.Error(err))
		return nil, fmt.Errorf("failed to count total alerts: %w", err)
	}

	// 获取各状态统计
	if err := s.db.WithContext(ctx).Model(&model.Alert{}).Where("status = ?", model.AlertStatusNew).Count(&stats.Firing).Error; err != nil {
		logger.L.Error("Failed to count firing alerts", zap.Error(err))
		return nil, fmt.Errorf("failed to count firing alerts: %w", err)
	}

	if err := s.db.WithContext(ctx).Model(&model.Alert{}).Where("status = ?", model.AlertStatusAcknowledged).Count(&stats.Acknowledged).Error; err != nil {
		logger.L.Error("Failed to count acknowledged alerts", zap.Error(err))
		return nil, fmt.Errorf("failed to count acknowledged alerts: %w", err)
	}

	if err := s.db.WithContext(ctx).Model(&model.Alert{}).Where("status = ?", model.AlertStatusResolved).Count(&stats.Resolved).Error; err != nil {
		logger.L.Error("Failed to count resolved alerts", zap.Error(err))
		return nil, fmt.Errorf("failed to count resolved alerts: %w", err)
	}

	// 获取按级别统计
	var levelStats []struct {
		Level string `json:"level"`
		Count int64  `json:"count"`
	}
	if err := s.db.WithContext(ctx).Model(&model.Alert{}).Select("level, count(*) as count").Group("level").Scan(&levelStats).Error; err != nil {
		logger.L.Error("Failed to get level stats", zap.Error(err))
		return nil, fmt.Errorf("failed to get level stats: %w", err)
	}
	for _, stat := range levelStats {
		stats.ByLevel[stat.Level] = stat.Count
	}

	// 获取按来源统计
	var sourceStats []struct {
		Source string `json:"source"`
		Count  int64  `json:"count"`
	}
	if err := s.db.WithContext(ctx).Model(&model.Alert{}).Select("source, count(*) as count").Group("source").Scan(&sourceStats).Error; err != nil {
		logger.L.Error("Failed to get source stats", zap.Error(err))
		return nil, fmt.Errorf("failed to get source stats: %w", err)
	}
	for _, stat := range sourceStats {
		stats.BySource[stat.Source] = stat.Count
	}

	// 获取按严重程度统计
	var severityStats []struct {
		Severity string `json:"severity"`
		Count    int64  `json:"count"`
	}
	if err := s.db.WithContext(ctx).Model(&model.Alert{}).Select("severity, count(*) as count").Group("severity").Scan(&severityStats).Error; err != nil {
		logger.L.Error("Failed to get severity stats", zap.Error(err))
		return nil, fmt.Errorf("failed to get severity stats: %w", err)
	}
	for _, stat := range severityStats {
		stats.BySeverity[stat.Severity] = stat.Count
	}

	logger.L.Debug("Fetched alert stats from database",
		zap.Int64("total", stats.Total),
		zap.Int64("firing", stats.Firing),
		zap.Int64("acknowledged", stats.Acknowledged),
		zap.Int64("resolved", stats.Resolved),
	)

	return stats, nil
}

// RefreshAlertStats 刷新告警统计缓存
func (s *AlertStatsService) RefreshAlertStats(ctx context.Context) error {
	return s.cacheService.InvalidateCache(ctx, AlertStatsKey)
}

// InvalidateAlertStatsCache 清除告警统计缓存
func (s *AlertStatsService) InvalidateAlertStatsCache(ctx context.Context) error {
	return s.cacheService.InvalidateCache(ctx, AlertStatsKey)
}