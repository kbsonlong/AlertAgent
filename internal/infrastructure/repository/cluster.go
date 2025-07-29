package repository

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"alert_agent/internal/domain/cluster"
	"alert_agent/internal/infrastructure/database"
)

// ClusterRepositoryImpl 集群仓储实现
type ClusterRepositoryImpl struct {
	db     *gorm.DB
	logger *zap.Logger
}

// NewClusterRepository 创建集群仓储
func NewClusterRepository(db *gorm.DB, logger *zap.Logger) cluster.ClusterRepository {
	return &ClusterRepositoryImpl{
		db:     db,
		logger: logger,
	}
}

// Create 创建集群
func (r *ClusterRepositoryImpl) Create(ctx context.Context, c *cluster.Cluster) error {
	dbCluster := r.toDBModel(c)
	if err := r.db.WithContext(ctx).Create(dbCluster).Error; err != nil {
		return fmt.Errorf("failed to create cluster: %w", err)
	}
	return nil
}

// Update 更新集群
func (r *ClusterRepositoryImpl) Update(ctx context.Context, c *cluster.Cluster) error {
	dbCluster := r.toDBModel(c)
	if err := r.db.WithContext(ctx).Save(dbCluster).Error; err != nil {
		return fmt.Errorf("failed to update cluster: %w", err)
	}
	return nil
}

// Delete 删除集群
func (r *ClusterRepositoryImpl) Delete(ctx context.Context, id string) error {
	if err := r.db.WithContext(ctx).Delete(&database.Cluster{}, "id = ?", id).Error; err != nil {
		return fmt.Errorf("failed to delete cluster: %w", err)
	}
	return nil
}

// GetByID 根据ID获取集群
func (r *ClusterRepositoryImpl) GetByID(ctx context.Context, id string) (*cluster.Cluster, error) {
	var dbCluster database.Cluster
	if err := r.db.WithContext(ctx).First(&dbCluster, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("cluster not found")
		}
		return nil, fmt.Errorf("failed to get cluster: %w", err)
	}
	return r.toDomainModel(&dbCluster), nil
}

// GetByName 根据名称获取集群
func (r *ClusterRepositoryImpl) GetByName(ctx context.Context, name string) (*cluster.Cluster, error) {
	var dbCluster database.Cluster
	if err := r.db.WithContext(ctx).First(&dbCluster, "name = ?", name).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("cluster not found")
		}
		return nil, fmt.Errorf("failed to get cluster: %w", err)
	}
	return r.toDomainModel(&dbCluster), nil
}

// List 列出集群
func (r *ClusterRepositoryImpl) List(ctx context.Context) ([]*cluster.Cluster, error) {
	var dbClusters []database.Cluster
	if err := r.db.WithContext(ctx).Order("created_at DESC").Find(&dbClusters).Error; err != nil {
		return nil, fmt.Errorf("failed to list clusters: %w", err)
	}

	clusters := make([]*cluster.Cluster, len(dbClusters))
	for i, dbCluster := range dbClusters {
		clusters[i] = r.toDomainModel(&dbCluster)
	}

	return clusters, nil
}

// GetByLabels 根据标签获取集群
func (r *ClusterRepositoryImpl) GetByLabels(ctx context.Context, labels map[string]string) ([]*cluster.Cluster, error) {
	// 这里需要根据具体的数据库实现来查询JSON字段
	// 简化实现，实际应该使用JSON查询
	var dbClusters []database.Cluster
	if err := r.db.WithContext(ctx).Find(&dbClusters).Error; err != nil {
		return nil, fmt.Errorf("failed to get clusters by labels: %w", err)
	}

	// 过滤包含指定标签的集群
	var filteredClusters []*cluster.Cluster
	for _, dbCluster := range dbClusters {
		if r.matchesLabels(dbCluster.Labels, labels) {
			filteredClusters = append(filteredClusters, r.toDomainModel(&dbCluster))
		}
	}

	return filteredClusters, nil
}

// GetActiveCluster 获取活跃集群
func (r *ClusterRepositoryImpl) GetActiveCluster(ctx context.Context) ([]*cluster.Cluster, error) {
	var dbClusters []database.Cluster
	if err := r.db.WithContext(ctx).Where("status = ?", "active").Find(&dbClusters).Error; err != nil {
		return nil, fmt.Errorf("failed to get active clusters: %w", err)
	}

	clusters := make([]*cluster.Cluster, len(dbClusters))
	for i, dbCluster := range dbClusters {
		clusters[i] = r.toDomainModel(&dbCluster)
	}

	return clusters, nil
}

// UpdateHealthStatus 更新健康状态
func (r *ClusterRepositoryImpl) UpdateHealthStatus(ctx context.Context, id string, status cluster.HealthStatus) error {
	updates := map[string]interface{}{
		"health_status":      string(status),
		"last_health_check":  gorm.Expr("NOW()"),
	}

	if err := r.db.WithContext(ctx).Model(&database.Cluster{}).Where("id = ?", id).Updates(updates).Error; err != nil {
		return fmt.Errorf("failed to update health status: %w", err)
	}

	return nil
}

// GetUnhealthyClusters 获取不健康的集群
func (r *ClusterRepositoryImpl) GetUnhealthyClusters(ctx context.Context) ([]*cluster.Cluster, error) {
	var dbClusters []database.Cluster
	if err := r.db.WithContext(ctx).Where("health_status = ?", "unhealthy").Find(&dbClusters).Error; err != nil {
		return nil, fmt.Errorf("failed to get unhealthy clusters: %w", err)
	}

	clusters := make([]*cluster.Cluster, len(dbClusters))
	for i, dbCluster := range dbClusters {
		clusters[i] = r.toDomainModel(&dbCluster)
	}

	return clusters, nil
}

// UpdateSyncStatus 更新同步状态
func (r *ClusterRepositoryImpl) UpdateSyncStatus(ctx context.Context, id string, status cluster.SyncStatus, syncTime time.Time) error {
	updates := map[string]interface{}{
		"sync_status":    string(status),
		"last_sync_time": syncTime,
	}

	if err := r.db.WithContext(ctx).Model(&database.Cluster{}).Where("id = ?", id).Updates(updates).Error; err != nil {
		return fmt.Errorf("failed to update sync status: %w", err)
	}

	return nil
}

// CreateSyncRecord 创建同步记录
func (r *ClusterRepositoryImpl) CreateSyncRecord(ctx context.Context, record *cluster.SyncRecord) error {
	dbRecord := &database.ConfigSyncRecord{
		ID:           record.ID,
		ClusterID:    record.ClusterID,
		ConfigType:   record.ConfigType,
		ConfigHash:   record.ConfigHash,
		SyncStatus:   string(record.SyncStatus),
		StartTime:    record.StartTime,
		EndTime:      record.EndTime,
		Duration:     record.Duration,
		ErrorMessage: record.ErrorMessage,
		ConfigData:   record.ConfigData,
		CreatedAt:    record.CreatedAt,
	}

	if err := r.db.WithContext(ctx).Create(dbRecord).Error; err != nil {
		return fmt.Errorf("failed to create sync record: %w", err)
	}

	return nil
}

// GetSyncRecords 获取同步记录
func (r *ClusterRepositoryImpl) GetSyncRecords(ctx context.Context, clusterID string, limit int) ([]*cluster.SyncRecord, error) {
	var dbRecords []database.ConfigSyncRecord
	query := r.db.WithContext(ctx).Where("cluster_id = ?", clusterID).Order("created_at DESC")
	
	if limit > 0 {
		query = query.Limit(limit)
	}

	if err := query.Find(&dbRecords).Error; err != nil {
		return nil, fmt.Errorf("failed to get sync records: %w", err)
	}

	records := make([]*cluster.SyncRecord, len(dbRecords))
	for i, dbRecord := range dbRecords {
		records[i] = &cluster.SyncRecord{
			ID:           dbRecord.ID,
			ClusterID:    dbRecord.ClusterID,
			ConfigType:   dbRecord.ConfigType,
			ConfigHash:   dbRecord.ConfigHash,
			SyncStatus:   cluster.SyncStatus(dbRecord.SyncStatus),
			StartTime:    dbRecord.StartTime,
			EndTime:      dbRecord.EndTime,
			Duration:     dbRecord.Duration,
			ErrorMessage: dbRecord.ErrorMessage,
			ConfigData:   dbRecord.ConfigData,
			CreatedAt:    dbRecord.CreatedAt,
		}
	}

	return records, nil
}

// toDBModel 转换为数据库模型
func (r *ClusterRepositoryImpl) toDBModel(c *cluster.Cluster) *database.Cluster {
	return &database.Cluster{
		ID:                  c.ID,
		Name:                c.Name,
		Endpoint:            c.Endpoint,
		ConfigPath:          c.ConfigPath,
		RulesPath:           c.RulesPath,
		SyncInterval:        c.SyncInterval,
		HealthCheckInterval: c.HealthCheckInterval,
		Status:              string(c.Status),
		HealthStatus:        string(c.HealthStatus),
		LastHealthCheck:     c.LastHealthCheck,
		LastSyncTime:        c.LastSyncTime,
		SyncStatus:          string(c.SyncStatus),
		Labels:              c.Labels,
		Metadata:            c.Metadata,
		CreatedAt:           c.CreatedAt,
		UpdatedAt:           c.UpdatedAt,
	}
}

// toDomainModel 转换为领域模型
func (r *ClusterRepositoryImpl) toDomainModel(dbCluster *database.Cluster) *cluster.Cluster {
	return &cluster.Cluster{
		ID:                  dbCluster.ID,
		Name:                dbCluster.Name,
		Endpoint:            dbCluster.Endpoint,
		ConfigPath:          dbCluster.ConfigPath,
		RulesPath:           dbCluster.RulesPath,
		SyncInterval:        dbCluster.SyncInterval,
		HealthCheckInterval: dbCluster.HealthCheckInterval,
		Status:              cluster.ClusterStatus(dbCluster.Status),
		HealthStatus:        cluster.HealthStatus(dbCluster.HealthStatus),
		LastHealthCheck:     dbCluster.LastHealthCheck,
		LastSyncTime:        dbCluster.LastSyncTime,
		SyncStatus:          cluster.SyncStatus(dbCluster.SyncStatus),
		Labels:              dbCluster.Labels,
		Metadata:            dbCluster.Metadata,
		CreatedAt:           dbCluster.CreatedAt,
		UpdatedAt:           dbCluster.UpdatedAt,
	}
}

// matchesLabels 检查标签是否匹配
func (r *ClusterRepositoryImpl) matchesLabels(clusterLabels, searchLabels map[string]string) bool {
	if len(searchLabels) == 0 {
		return true
	}

	for key, value := range searchLabels {
		if clusterValue, exists := clusterLabels[key]; !exists || clusterValue != value {
			return false
		}
	}

	return true
}