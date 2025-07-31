package service

import (
	"context"
	"fmt"
	"time"

	"alert_agent/internal/model"
	"alert_agent/internal/pkg/database"
	"alert_agent/internal/pkg/logger"

	"go.uber.org/zap"
)

// ConfigSyncMonitor 配置同步监控服务
type ConfigSyncMonitor struct {
	configService *ConfigService
}

// NewConfigSyncMonitor 创建配置同步监控服务
func NewConfigSyncMonitor() *ConfigSyncMonitor {
	return &ConfigSyncMonitor{
		configService: NewConfigService(),
	}
}

// SyncMetrics 同步指标
type SyncMetrics struct {
	ClusterID       string            `json:"cluster_id"`
	ConfigType      string            `json:"config_type"`
	LastSyncTime    *time.Time        `json:"last_sync_time"`
	SyncStatus      string            `json:"sync_status"`
	SyncDelay       int64             `json:"sync_delay_seconds"`
	SuccessRate     float64           `json:"success_rate"`
	FailureCount    int64             `json:"failure_count"`
	AverageDuration int64             `json:"average_duration_ms"`
	ErrorMessage    string            `json:"error_message,omitempty"`
	IsHealthy       bool              `json:"is_healthy"`
	ConfigHash      string            `json:"config_hash"`
	ConfigSize      int64             `json:"config_size"`
}

// SyncSummary 同步状态汇总
type SyncSummary struct {
	TotalClusters    int                        `json:"total_clusters"`
	HealthyClusters  int                        `json:"healthy_clusters"`
	UnhealthyClusters int                       `json:"unhealthy_clusters"`
	ConfigTypes      []string                   `json:"config_types"`
	ClusterMetrics   map[string][]SyncMetrics   `json:"cluster_metrics"`
	OverallHealth    string                     `json:"overall_health"`
	LastUpdateTime   time.Time                  `json:"last_update_time"`
}

// CollectSyncMetrics 收集同步指标
func (csm *ConfigSyncMonitor) CollectSyncMetrics(ctx context.Context) (*SyncSummary, error) {
	logger.L.Debug("Collecting sync metrics")

	// 获取所有同步状态
	var statuses []model.ConfigSyncStatus
	err := database.DB.WithContext(ctx).Find(&statuses).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get sync statuses: %w", err)
	}

	// 按集群分组计算指标
	clusterMetrics := make(map[string][]SyncMetrics)
	configTypesSet := make(map[string]bool)
	healthyClusters := 0
	totalClusters := 0

	clusterHealthMap := make(map[string]bool)

	for _, status := range statuses {
		metrics, err := csm.calculateMetrics(ctx, status)
		if err != nil {
			logger.L.Error("Failed to calculate metrics",
				zap.String("cluster_id", status.ClusterID),
				zap.String("config_type", status.ConfigType),
				zap.Error(err),
			)
			continue
		}

		clusterMetrics[status.ClusterID] = append(clusterMetrics[status.ClusterID], *metrics)
		configTypesSet[status.ConfigType] = true

		// 记录集群健康状态
		if _, exists := clusterHealthMap[status.ClusterID]; !exists {
			clusterHealthMap[status.ClusterID] = true
			totalClusters++
		}
		if !metrics.IsHealthy {
			clusterHealthMap[status.ClusterID] = false
		}
	}

	// 统计健康集群数量
	for _, isHealthy := range clusterHealthMap {
		if isHealthy {
			healthyClusters++
		}
	}

	// 提取配置类型列表
	configTypes := make([]string, 0, len(configTypesSet))
	for configType := range configTypesSet {
		configTypes = append(configTypes, configType)
	}

	// 计算整体健康状态
	overallHealth := "healthy"
	if healthyClusters == 0 {
		overallHealth = "critical"
	} else if healthyClusters < totalClusters {
		overallHealth = "warning"
	}

	summary := &SyncSummary{
		TotalClusters:     totalClusters,
		HealthyClusters:   healthyClusters,
		UnhealthyClusters: totalClusters - healthyClusters,
		ConfigTypes:       configTypes,
		ClusterMetrics:    clusterMetrics,
		OverallHealth:     overallHealth,
		LastUpdateTime:    time.Now(),
	}

	logger.L.Info("Sync metrics collected",
		zap.Int("total_clusters", totalClusters),
		zap.Int("healthy_clusters", healthyClusters),
		zap.String("overall_health", overallHealth),
	)

	return summary, nil
}

// calculateMetrics 计算单个配置类型的指标
func (csm *ConfigSyncMonitor) calculateMetrics(ctx context.Context, status model.ConfigSyncStatus) (*SyncMetrics, error) {
	// 获取历史记录计算成功率和平均耗时
	var histories []model.ConfigSyncHistory
	err := database.DB.WithContext(ctx).
		Where("cluster_id = ? AND config_type = ?", status.ClusterID, status.ConfigType).
		Where("created_at >= ?", time.Now().Add(-24*time.Hour)). // 最近24小时
		Order("created_at DESC").
		Limit(100).
		Find(&histories).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get sync history: %w", err)
	}

	// 计算成功率
	successCount := int64(0)
	totalCount := int64(len(histories))
	totalDuration := int64(0)
	failureCount := int64(0)

	for _, history := range histories {
		if history.SyncStatus == "success" {
			successCount++
		} else {
			failureCount++
		}
		totalDuration += history.SyncDuration
	}

	successRate := float64(0)
	if totalCount > 0 {
		successRate = float64(successCount) / float64(totalCount) * 100
	}

	averageDuration := int64(0)
	if totalCount > 0 {
		averageDuration = totalDuration / totalCount
	}

	// 计算同步延迟
	syncDelay := status.GetSyncDelay()

	// 判断健康状态（成功率>90%，延迟<300秒，最近状态为成功）
	isHealthy := status.IsHealthy(300) && successRate >= 90.0

	metrics := &SyncMetrics{
		ClusterID:       status.ClusterID,
		ConfigType:      status.ConfigType,
		LastSyncTime:    status.SyncTime,
		SyncStatus:      status.SyncStatus,
		SyncDelay:       syncDelay,
		SuccessRate:     successRate,
		FailureCount:    failureCount,
		AverageDuration: averageDuration,
		ErrorMessage:    status.ErrorMessage,
		IsHealthy:       isHealthy,
		ConfigHash:      status.ConfigHash,
	}

	// 获取配置大小
	if len(histories) > 0 {
		metrics.ConfigSize = histories[0].ConfigSize
	}

	return metrics, nil
}

// GetSyncDelayMetrics 获取同步延迟指标
func (csm *ConfigSyncMonitor) GetSyncDelayMetrics(ctx context.Context, clusterID, configType string, hours int) ([]SyncDelayPoint, error) {
	if hours <= 0 {
		hours = 24
	}

	var histories []model.ConfigSyncHistory
	query := database.DB.WithContext(ctx).
		Where("created_at >= ?", time.Now().Add(-time.Duration(hours)*time.Hour)).
		Order("created_at ASC")

	if clusterID != "" {
		query = query.Where("cluster_id = ?", clusterID)
	}
	if configType != "" {
		query = query.Where("config_type = ?", configType)
	}

	err := query.Find(&histories).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get sync history: %w", err)
	}

	// 按时间间隔聚合数据
	points := make([]SyncDelayPoint, 0)
	intervalMinutes := 30 // 30分钟间隔

	if len(histories) == 0 {
		return points, nil
	}

	startTime := histories[0].CreatedAt.Truncate(time.Duration(intervalMinutes) * time.Minute)
	endTime := time.Now().Truncate(time.Duration(intervalMinutes) * time.Minute)

	for t := startTime; t.Before(endTime) || t.Equal(endTime); t = t.Add(time.Duration(intervalMinutes) * time.Minute) {
		nextT := t.Add(time.Duration(intervalMinutes) * time.Minute)
		
		var intervalHistories []model.ConfigSyncHistory
		for _, h := range histories {
			if (h.CreatedAt.Equal(t) || h.CreatedAt.After(t)) && h.CreatedAt.Before(nextT) {
				intervalHistories = append(intervalHistories, h)
			}
		}

		if len(intervalHistories) > 0 {
			// 计算该时间段的平均延迟
			totalDuration := int64(0)
			successCount := 0
			for _, h := range intervalHistories {
				totalDuration += h.SyncDuration
				if h.SyncStatus == "success" {
					successCount++
				}
			}

			avgDuration := totalDuration / int64(len(intervalHistories))
			successRate := float64(successCount) / float64(len(intervalHistories)) * 100

			points = append(points, SyncDelayPoint{
				Timestamp:    t,
				Duration:     avgDuration,
				SuccessRate:  successRate,
				SampleCount:  len(intervalHistories),
			})
		}
	}

	return points, nil
}

// SyncDelayPoint 同步延迟数据点
type SyncDelayPoint struct {
	Timestamp   time.Time `json:"timestamp"`
	Duration    int64     `json:"duration_ms"`
	SuccessRate float64   `json:"success_rate"`
	SampleCount int       `json:"sample_count"`
}

// GetFailureRateMetrics 获取失败率指标
func (csm *ConfigSyncMonitor) GetFailureRateMetrics(ctx context.Context, clusterID, configType string, hours int) ([]FailureRatePoint, error) {
	if hours <= 0 {
		hours = 24
	}

	var histories []model.ConfigSyncHistory
	query := database.DB.WithContext(ctx).
		Where("created_at >= ?", time.Now().Add(-time.Duration(hours)*time.Hour)).
		Order("created_at ASC")

	if clusterID != "" {
		query = query.Where("cluster_id = ?", clusterID)
	}
	if configType != "" {
		query = query.Where("config_type = ?", configType)
	}

	err := query.Find(&histories).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get sync history: %w", err)
	}

	// 按小时聚合失败率数据
	points := make([]FailureRatePoint, 0)
	
	if len(histories) == 0 {
		return points, nil
	}

	startTime := histories[0].CreatedAt.Truncate(time.Hour)
	endTime := time.Now().Truncate(time.Hour)

	for t := startTime; t.Before(endTime) || t.Equal(endTime); t = t.Add(time.Hour) {
		nextT := t.Add(time.Hour)
		
		totalCount := 0
		failureCount := 0
		
		for _, h := range histories {
			if (h.CreatedAt.Equal(t) || h.CreatedAt.After(t)) && h.CreatedAt.Before(nextT) {
				totalCount++
				if h.SyncStatus != "success" {
					failureCount++
				}
			}
		}

		if totalCount > 0 {
			failureRate := float64(failureCount) / float64(totalCount) * 100
			points = append(points, FailureRatePoint{
				Timestamp:   t,
				FailureRate: failureRate,
				TotalCount:  totalCount,
				FailureCount: failureCount,
			})
		}
	}

	return points, nil
}

// FailureRatePoint 失败率数据点
type FailureRatePoint struct {
	Timestamp    time.Time `json:"timestamp"`
	FailureRate  float64   `json:"failure_rate"`
	TotalCount   int       `json:"total_count"`
	FailureCount int       `json:"failure_count"`
}

// RecordSyncHistory 记录同步历史
func (csm *ConfigSyncMonitor) RecordSyncHistory(ctx context.Context, clusterID, configType, configHash string, 
	configSize int64, syncStatus string, syncDuration int64, errorMessage string) error {
	
	history := &model.ConfigSyncHistory{
		ClusterID:    clusterID,
		ConfigType:   configType,
		ConfigHash:   configHash,
		ConfigSize:   configSize,
		SyncStatus:   syncStatus,
		SyncDuration: syncDuration,
		ErrorMessage: errorMessage,
	}

	if err := database.DB.WithContext(ctx).Create(history).Error; err != nil {
		return fmt.Errorf("failed to record sync history: %w", err)
	}

	logger.L.Debug("Sync history recorded",
		zap.String("cluster_id", clusterID),
		zap.String("config_type", configType),
		zap.String("status", syncStatus),
		zap.Int64("duration", syncDuration),
	)

	return nil
}

// CleanupOldHistory 清理旧的历史记录
func (csm *ConfigSyncMonitor) CleanupOldHistory(ctx context.Context, retentionDays int) error {
	if retentionDays <= 0 {
		retentionDays = 30 // 默认保留30天
	}

	cutoffTime := time.Now().Add(-time.Duration(retentionDays) * 24 * time.Hour)
	
	result := database.DB.WithContext(ctx).
		Where("created_at < ?", cutoffTime).
		Delete(&model.ConfigSyncHistory{})

	if result.Error != nil {
		return fmt.Errorf("failed to cleanup old history: %w", result.Error)
	}

	logger.L.Info("Old sync history cleaned up",
		zap.Int64("deleted_records", result.RowsAffected),
		zap.Int("retention_days", retentionDays),
	)

	return nil
}