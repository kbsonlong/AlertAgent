package cluster

import (
	"context"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"alert_agent/internal/domain/cluster"
	"alert_agent/internal/shared/errors"
)

// ClusterService 集群服务接口
type ClusterService interface {
	// 集群管理
	RegisterCluster(ctx context.Context, config *cluster.ClusterConfig) (*cluster.Cluster, error)
	UpdateCluster(ctx context.Context, id string, config *cluster.ClusterConfig) (*cluster.Cluster, error)
	DeleteCluster(ctx context.Context, id string) error
	GetCluster(ctx context.Context, id string) (*cluster.Cluster, error)
	GetClusterByName(ctx context.Context, name string) (*cluster.Cluster, error)
	ListClusters(ctx context.Context) ([]*cluster.Cluster, error)

	// 健康检查
	HealthCheck(ctx context.Context) (map[string]*cluster.HealthStatus, error)
	GetClusterHealth(ctx context.Context, id string) (*cluster.HealthStatus, error)

	// 配置分发
	DistributeConfig(ctx context.Context, clusterID string, config *cluster.Config) error
	GetSyncStatus(ctx context.Context, clusterID string) (*cluster.SyncStatus, error)
	GetSyncRecords(ctx context.Context, clusterID string, limit int) ([]*cluster.SyncRecord, error)
}

// clusterServiceImpl 集群服务实现
type clusterServiceImpl struct {
	repo         cluster.ClusterRepository
	synchronizer cluster.ConfigSynchronizer
	healthChecker cluster.HealthChecker
	logger       *zap.Logger
}

// NewClusterService 创建集群服务
func NewClusterService(
	repo cluster.ClusterRepository,
	logger *zap.Logger,
) ClusterService {
	return &clusterServiceImpl{
		repo:   repo,
		logger: logger,
	}
}

// RegisterCluster 注册集群
func (s *clusterServiceImpl) RegisterCluster(ctx context.Context, config *cluster.ClusterConfig) (*cluster.Cluster, error) {
	// 检查名称是否已存在
	existing, err := s.repo.GetByName(ctx, config.Name)
	if err == nil && existing != nil {
		return nil, errors.NewConflictError("Cluster name already exists")
	}

	// 创建集群对象
	c := &cluster.Cluster{
		ID:                  uuid.New().String(),
		Name:                config.Name,
		Endpoint:            config.Endpoint,
		ConfigPath:          config.ConfigPath,
		RulesPath:           config.RulesPath,
		SyncInterval:        config.SyncInterval,
		HealthCheckInterval: config.HealthCheckInterval,
		Status:              cluster.ClusterStatusActive,
		HealthStatus:        cluster.HealthStatusUnknown,
		SyncStatus:          cluster.SyncStatusPending,
		Labels:              config.Labels,
		Metadata:            config.Metadata,
		CreatedAt:           time.Now(),
		UpdatedAt:           time.Now(),
	}

	// 设置默认值
	if c.SyncInterval == 0 {
		c.SyncInterval = 30
	}
	if c.HealthCheckInterval == 0 {
		c.HealthCheckInterval = 10
	}

	// 保存到数据库
	if err := s.repo.Create(ctx, c); err != nil {
		s.logger.Error("Failed to register cluster", zap.Error(err), zap.String("name", config.Name))
		return nil, errors.NewInternalError("Failed to register cluster", err)
	}

	s.logger.Info("Cluster registered successfully", 
		zap.String("id", c.ID), 
		zap.String("name", c.Name), 
		zap.String("endpoint", c.Endpoint))
	return c, nil
}

// UpdateCluster 更新集群
func (s *clusterServiceImpl) UpdateCluster(ctx context.Context, id string, config *cluster.ClusterConfig) (*cluster.Cluster, error) {
	// 获取现有集群
	c, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, errors.NewNotFoundError("Cluster")
	}

	// 如果名称发生变化，检查是否冲突
	if config.Name != c.Name {
		existing, err := s.repo.GetByName(ctx, config.Name)
		if err == nil && existing != nil && existing.ID != id {
			return nil, errors.NewConflictError("Cluster name already exists")
		}
		c.Name = config.Name
	}

	// 更新字段
	c.Endpoint = config.Endpoint
	c.ConfigPath = config.ConfigPath
	c.RulesPath = config.RulesPath
	if config.SyncInterval > 0 {
		c.SyncInterval = config.SyncInterval
	}
	if config.HealthCheckInterval > 0 {
		c.HealthCheckInterval = config.HealthCheckInterval
	}
	c.Labels = config.Labels
	c.Metadata = config.Metadata
	c.UpdatedAt = time.Now()

	// 保存更新
	if err := s.repo.Update(ctx, c); err != nil {
		s.logger.Error("Failed to update cluster", zap.Error(err), zap.String("id", id))
		return nil, errors.NewInternalError("Failed to update cluster", err)
	}

	s.logger.Info("Cluster updated successfully", zap.String("id", id), zap.String("name", c.Name))
	return c, nil
}

// DeleteCluster 删除集群
func (s *clusterServiceImpl) DeleteCluster(ctx context.Context, id string) error {
	// 检查集群是否存在
	_, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return errors.NewNotFoundError("Cluster")
	}

	// 删除集群
	if err := s.repo.Delete(ctx, id); err != nil {
		s.logger.Error("Failed to delete cluster", zap.Error(err), zap.String("id", id))
		return errors.NewInternalError("Failed to delete cluster", err)
	}

	s.logger.Info("Cluster deleted successfully", zap.String("id", id))
	return nil
}

// GetCluster 获取集群
func (s *clusterServiceImpl) GetCluster(ctx context.Context, id string) (*cluster.Cluster, error) {
	c, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, errors.NewNotFoundError("Cluster")
	}
	return c, nil
}

// GetClusterByName 根据名称获取集群
func (s *clusterServiceImpl) GetClusterByName(ctx context.Context, name string) (*cluster.Cluster, error) {
	c, err := s.repo.GetByName(ctx, name)
	if err != nil {
		return nil, errors.NewNotFoundError("Cluster")
	}
	return c, nil
}

// ListClusters 列出集群
func (s *clusterServiceImpl) ListClusters(ctx context.Context) ([]*cluster.Cluster, error) {
	clusters, err := s.repo.List(ctx)
	if err != nil {
		s.logger.Error("Failed to list clusters", zap.Error(err))
		return nil, errors.NewInternalError("Failed to list clusters", err)
	}
	return clusters, nil
}

// HealthCheck 健康检查
func (s *clusterServiceImpl) HealthCheck(ctx context.Context) (map[string]*cluster.HealthStatus, error) {
	clusters, err := s.repo.GetActiveCluster(ctx)
	if err != nil {
		return nil, errors.NewInternalError("Failed to get active clusters", err)
	}

	result := make(map[string]*cluster.HealthStatus)
	for _, c := range clusters {
		if s.healthChecker != nil {
			status, err := s.healthChecker.CheckHealth(ctx, c)
			if err != nil {
				s.logger.Warn("Health check failed", zap.String("cluster_id", c.ID), zap.Error(err))
				unhealthy := cluster.HealthStatusUnhealthy
				result[c.ID] = &unhealthy
			} else {
				result[c.ID] = status
			}

			// 更新数据库中的健康状态
			s.repo.UpdateHealthStatus(ctx, c.ID, *result[c.ID])
		} else {
			unknown := cluster.HealthStatusUnknown
			result[c.ID] = &unknown
		}
	}

	return result, nil
}

// GetClusterHealth 获取集群健康状态
func (s *clusterServiceImpl) GetClusterHealth(ctx context.Context, id string) (*cluster.HealthStatus, error) {
	c, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, errors.NewNotFoundError("Cluster")
	}

	if s.healthChecker != nil {
		status, err := s.healthChecker.CheckHealth(ctx, c)
		if err != nil {
			s.logger.Warn("Health check failed", zap.String("cluster_id", id), zap.Error(err))
			unhealthy := cluster.HealthStatusUnhealthy
			return &unhealthy, nil
		}
		return status, nil
	}

	return &c.HealthStatus, nil
}

// DistributeConfig 分发配置
func (s *clusterServiceImpl) DistributeConfig(ctx context.Context, clusterID string, config *cluster.Config) error {
	// 获取集群
	c, err := s.repo.GetByID(ctx, clusterID)
	if err != nil {
		return errors.NewNotFoundError("Cluster")
	}

	// 检查集群状态
	if c.Status != cluster.ClusterStatusActive {
		return errors.NewValidationError("CLUSTER_INACTIVE", "Cluster is not active")
	}

	// 同步配置
	if s.synchronizer != nil {
		if err := s.synchronizer.SyncConfig(ctx, c, config); err != nil {
			s.logger.Error("Failed to sync config", zap.Error(err), zap.String("cluster_id", clusterID))
			
			// 更新同步状态
			s.repo.UpdateSyncStatus(ctx, clusterID, cluster.SyncStatusFailed, time.Now())
			
			return errors.NewExternalError("ConfigSync", "Failed to sync configuration", err)
		}

		// 更新同步状态
		s.repo.UpdateSyncStatus(ctx, clusterID, cluster.SyncStatusSuccess, time.Now())
	}

	s.logger.Info("Configuration distributed successfully", zap.String("cluster_id", clusterID))
	return nil
}

// GetSyncStatus 获取同步状态
func (s *clusterServiceImpl) GetSyncStatus(ctx context.Context, clusterID string) (*cluster.SyncStatus, error) {
	c, err := s.repo.GetByID(ctx, clusterID)
	if err != nil {
		return nil, errors.NewNotFoundError("Cluster")
	}

	if s.synchronizer != nil {
		status, err := s.synchronizer.GetSyncStatus(ctx, clusterID)
		if err != nil {
			s.logger.Warn("Failed to get sync status", zap.String("cluster_id", clusterID), zap.Error(err))
			return &c.SyncStatus, nil
		}
		return status, nil
	}

	return &c.SyncStatus, nil
}

// GetSyncRecords 获取同步记录
func (s *clusterServiceImpl) GetSyncRecords(ctx context.Context, clusterID string, limit int) ([]*cluster.SyncRecord, error) {
	// 检查集群是否存在
	_, err := s.repo.GetByID(ctx, clusterID)
	if err != nil {
		return nil, errors.NewNotFoundError("Cluster")
	}

	records, err := s.repo.GetSyncRecords(ctx, clusterID, limit)
	if err != nil {
		s.logger.Error("Failed to get sync records", zap.Error(err), zap.String("cluster_id", clusterID))
		return nil, errors.NewInternalError("Failed to get sync records", err)
	}

	return records, nil
}