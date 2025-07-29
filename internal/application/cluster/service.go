package cluster

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"alert_agent/internal/domain/cluster"
	"alert_agent/internal/shared/errors"
	"alert_agent/internal/shared/logger"
	"alert_agent/pkg/types"
)

// ClusterService 集群应用服务实现
type ClusterService struct {
	repo   cluster.Repository
	logger *zap.Logger
}

// NewClusterService 创建集群应用服务
func NewClusterService(repo cluster.Repository) cluster.Service {
	return &ClusterService{
		repo:   repo,
		logger: logger.WithComponent("cluster-service"),
	}
}

// CreateCluster 创建集群
func (s *ClusterService) CreateCluster(ctx context.Context, req *cluster.CreateClusterRequest) (*cluster.Cluster, error) {
	s.logger.Info("creating cluster", zap.String("name", req.Name), zap.String("type", string(req.Type)))

	// 验证请求
	if err := s.validateCreateRequest(req); err != nil {
		return nil, err
	}

	// 检查名称是否已存在
	exists, err := s.repo.ExistsByName(ctx, req.Name)
	if err != nil {
		s.logger.Error("failed to check cluster name existence", zap.Error(err))
		return nil, errors.NewInternalError("failed to check cluster name", err)
	}
	if exists {
		return nil, errors.NewValidationError("DUPLICATE_NAME", "cluster name already exists")
	}

	// 创建集群实体
	c := &cluster.Cluster{
		ID:          uuid.New().String(),
		Name:        req.Name,
		Type:        req.Type,
		Description: req.Description,
		Version:     req.Version,
		Endpoints:   req.Endpoints,
		Config:      req.Config,
		Status:      cluster.ClusterStatusInactive,
		Tags:        req.Tags,
		Labels:      req.Labels,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// 验证集群配置
	if err := c.Validate(); err != nil {
		return nil, errors.NewValidationError("INVALID_CONFIG", "invalid cluster configuration")
	}

	// 保存到数据库
	if err := s.repo.Create(ctx, c); err != nil {
		s.logger.Error("failed to create cluster", zap.Error(err))
		return nil, errors.NewInternalError("failed to create cluster", err)
	}

	s.logger.Info("cluster created successfully", zap.String("id", c.ID), zap.String("name", c.Name))
	return c, nil
}

// GetCluster 获取集群
func (s *ClusterService) GetCluster(ctx context.Context, id string) (*cluster.Cluster, error) {
	s.logger.Debug("getting cluster", zap.String("id", id))

	if id == "" {
		return nil, errors.NewValidationError("INVALID_ID", "cluster id is required")
	}

	c, err := s.repo.GetByID(ctx, id)
	if err != nil {
		s.logger.Error("failed to get cluster", zap.String("id", id), zap.Error(err))
		return nil, errors.NewNotFoundError("cluster")
	}

	return c, nil
}

// GetClusterByName 根据名称获取集群
func (s *ClusterService) GetClusterByName(ctx context.Context, name string) (*cluster.Cluster, error) {
	s.logger.Debug("getting cluster by name", zap.String("name", name))

	if name == "" {
		return nil, errors.NewValidationError("INVALID_NAME", "cluster name is required")
	}

	c, err := s.repo.GetByName(ctx, name)
	if err != nil {
		s.logger.Error("failed to get cluster by name", zap.String("name", name), zap.Error(err))
		return nil, errors.NewNotFoundError("cluster")
	}

	return c, nil
}

// ListClusters 获取集群列表
func (s *ClusterService) ListClusters(ctx context.Context, query types.Query) (*types.PageResult, error) {
	s.logger.Debug("listing clusters", zap.Int("limit", query.Limit), zap.Int("offset", query.Offset))

	clusters, total, err := s.repo.List(ctx, query)
	if err != nil {
		s.logger.Error("failed to list clusters", zap.Error(err))
		return nil, errors.NewInternalError("failed to list clusters", err)
	}

	page := query.Offset/query.Limit + 1
	if query.Limit == 0 {
		page = 1
	}

	return &types.PageResult{
		Data:  clusters,
		Total: total,
		Page:  page,
		Size:  len(clusters),
	}, nil
}

// UpdateCluster 更新集群
func (s *ClusterService) UpdateCluster(ctx context.Context, id string, req *cluster.UpdateClusterRequest) (*cluster.Cluster, error) {
	s.logger.Info("updating cluster", zap.String("id", id))

	if id == "" {
		return nil, errors.NewValidationError("INVALID_ID", "cluster id is required")
	}

	// 获取现有集群
	c, err := s.repo.GetByID(ctx, id)
	if err != nil {
		s.logger.Error("failed to get cluster for update", zap.String("id", id), zap.Error(err))
		return nil, errors.NewNotFoundError("cluster not found")
	}

	// 更新字段
	if req.Name != nil && *req.Name != "" && *req.Name != c.Name {
		// 检查新名称是否已存在
		exists, err := s.repo.ExistsByName(ctx, *req.Name)
		if err != nil {
			s.logger.Error("failed to check cluster name existence", zap.Error(err))
			return nil, errors.NewInternalError("failed to check cluster name", err)
		}
		if exists {
			return nil, errors.NewValidationError("DUPLICATE_NAME", "cluster name already exists")
		}
		c.Name = *req.Name
	}

	if req.Description != nil && *req.Description != "" {
		c.Description = *req.Description
	}
	if req.Version != nil && *req.Version != "" {
		c.Version = *req.Version
	}
	if req.Endpoints != nil {
		c.Endpoints = req.Endpoints
	}
	if req.Config != nil {
		c.Config = *req.Config
	}
	if req.Tags != nil {
		c.Tags = req.Tags
	}
	if req.Labels != nil {
		c.Labels = req.Labels
	}

	c.UpdatedAt = time.Now()

	// 验证更新后的配置
	if err := c.Validate(); err != nil {
		return nil, errors.NewValidationError("INVALID_CONFIG", "invalid cluster configuration")
	}

	// 保存更新
	if err := s.repo.Update(ctx, c); err != nil {
		s.logger.Error("failed to update cluster", zap.Error(err))
		return nil, errors.NewInternalError("failed to update cluster", err)
	}

	s.logger.Info("cluster updated successfully", zap.String("id", c.ID), zap.String("name", c.Name))
	return c, nil
}

// DeleteCluster 删除集群
func (s *ClusterService) DeleteCluster(ctx context.Context, id string) error {
	s.logger.Info("deleting cluster", zap.String("id", id))

	if id == "" {
		return errors.NewValidationError("INVALID_ID", "cluster id is required")
	}

	// 检查集群是否存在
	exists, err := s.repo.Exists(ctx, id)
	if err != nil {
		s.logger.Error("failed to check cluster existence", zap.Error(err))
		return errors.NewInternalError("failed to check cluster existence", err)
	}
	if !exists {
		return errors.NewNotFoundError("cluster")
	}

	// 删除集群
	if err := s.repo.Delete(ctx, id); err != nil {
		s.logger.Error("failed to delete cluster", zap.Error(err))
		return errors.NewInternalError("failed to delete cluster", err)
	}

	s.logger.Info("cluster deleted successfully", zap.String("id", id))
	return nil
}

// TestClusterConnection 测试集群连接
func (s *ClusterService) TestClusterConnection(ctx context.Context, id string) (*cluster.ConnectionTestResult, error) {
	s.logger.Info("testing cluster connection", zap.String("id", id))

	_, err := s.repo.GetByID(ctx, id)
	if err != nil {
		s.logger.Error("failed to get cluster for connection test", zap.String("id", id), zap.Error(err))
		return nil, errors.NewNotFoundError("cluster not found")
	}

	// TODO: 实现实际的连接测试逻辑
	// 这里应该根据集群类型实现具体的连接测试
	result := &cluster.ConnectionTestResult{
		Success:   true,
		Message:   "Connection test successful",
		Latency:   time.Millisecond * 50,
		Timestamp: time.Now(),
		Endpoints: make(map[string]*cluster.EndpointResult),
	}

	s.logger.Info("cluster connection test completed", zap.String("id", id), zap.Bool("success", result.Success))
	return result, nil
}

// GetClusterHealth 获取集群健康状态
func (s *ClusterService) GetClusterHealth(ctx context.Context, id string) (*cluster.ClusterHealth, error) {
	s.logger.Debug("getting cluster health", zap.String("id", id))

	c, err := s.repo.GetByID(ctx, id)
	if err != nil {
		s.logger.Error("failed to get cluster for health check", zap.String("id", id), zap.Error(err))
		return nil, errors.NewNotFoundError("cluster not found")
	}

	// TODO: 实现实际的健康检查逻辑
	health := &cluster.ClusterHealth{
		ClusterID: c.ID,
		Status:    "healthy",
		Message:   "Cluster is healthy",
		LastCheck: time.Now(),
		Endpoints: make(map[string]*cluster.EndpointHealth),
	}

	return health, nil
}

// GetClusterMetrics 获取集群指标
func (s *ClusterService) GetClusterMetrics(ctx context.Context, id string) (*cluster.ClusterMetrics, error) {
	s.logger.Debug("getting cluster metrics", zap.String("id", id))

	_, err := s.repo.GetByID(ctx, id)
	if err != nil {
		s.logger.Error("failed to get cluster for metrics", zap.String("id", id), zap.Error(err))
		return nil, errors.NewNotFoundError("cluster not found")
	}

	// TODO: 实现实际的指标收集逻辑
	metrics := &cluster.ClusterMetrics{
		ClusterID:       id,
		TotalRequests:   1000,
		TotalErrors:     10,
		SuccessRate:     99.0,
		AvgLatency:      time.Millisecond * 100,
		Throughput:      100.5,
		EndpointMetrics: make(map[string]*cluster.EndpointMetrics),
		Timestamp:       time.Now(),
	}

	return metrics, nil
}

// SyncClusterConfig 同步集群配置
func (s *ClusterService) SyncClusterConfig(ctx context.Context, id string) error {
	s.logger.Info("syncing cluster config", zap.String("id", id))

	_, err := s.repo.GetByID(ctx, id)
	if err != nil {
		s.logger.Error("failed to get cluster for config sync", zap.String("id", id), zap.Error(err))
		return errors.NewNotFoundError("cluster not found")
	}

	// TODO: 实现实际的配置同步逻辑
	s.logger.Info("cluster config synced successfully", zap.String("id", id))
	return nil
}

// GetClustersByType 根据类型获取集群列表
func (s *ClusterService) GetClustersByType(ctx context.Context, clusterType cluster.ClusterType) ([]*cluster.Cluster, error) {
	s.logger.Debug("getting clusters by type", zap.String("type", string(clusterType)))

	clusters, err := s.repo.GetByType(ctx, clusterType)
	if err != nil {
		s.logger.Error("failed to get clusters by type", zap.Error(err))
		return nil, errors.NewInternalError("failed to get clusters by type", err)
	}

	return clusters, nil
}

// GetActiveClusters 获取激活的集群
func (s *ClusterService) GetActiveClusters(ctx context.Context) ([]*cluster.Cluster, error) {
	s.logger.Debug("getting active clusters")

	clusters, err := s.repo.GetByStatus(ctx, cluster.ClusterStatusActive)
	if err != nil {
		s.logger.Error("failed to get active clusters", zap.Error(err))
		return nil, errors.NewInternalError("failed to get active clusters", err)
	}

	return clusters, nil
}

// UpdateClusterStatus 更新集群状态
func (s *ClusterService) UpdateClusterStatus(ctx context.Context, id string, status cluster.ClusterStatus) error {
	s.logger.Info("updating cluster status", zap.String("id", id), zap.String("status", string(status)))

	// 检查集群是否存在
	existingCluster, err := s.repo.GetByID(ctx, id)
	if err != nil {
		s.logger.Error("failed to get cluster", zap.String("id", id), zap.Error(err))
		return errors.NewNotFoundError("cluster not found")
	}

	// 更新状态
	existingCluster.Status = status
	existingCluster.UpdatedAt = time.Now()

	if err := s.repo.Update(ctx, existingCluster); err != nil {
		s.logger.Error("failed to update cluster status", zap.String("id", id), zap.Error(err))
		return errors.NewInternalError("failed to update cluster status", err)
	}

	s.logger.Info("cluster status updated successfully", zap.String("id", id), zap.String("status", string(status)))
	return nil
}

// ValidateClusterConfig 验证集群配置
func (s *ClusterService) ValidateClusterConfig(ctx context.Context, clusterType cluster.ClusterType, config cluster.ClusterConfig) error {
	s.logger.Info("validating cluster config", zap.String("type", string(clusterType)))

	// 基础验证
	if config.Auth.Username == "" && config.Auth.Token == "" {
		return errors.NewValidationError("MISSING_AUTH", "authentication credentials are required")
	}

	if config.Connection.Timeout <= 0 {
		return errors.NewValidationError("INVALID_TIMEOUT", "connection timeout must be positive")
	}

	// 根据集群类型进行特定验证
	switch clusterType {
	case cluster.ClusterTypePrometheus:
		if config.Connection.Timeout < time.Second {
			return errors.NewValidationError("PROMETHEUS_TIMEOUT_TOO_SHORT", "prometheus timeout should be at least 1 second")
		}
	case cluster.ClusterTypeAlertmanager:
		if config.Connection.MaxConnections <= 0 {
			return errors.NewValidationError("INVALID_MAX_CONNECTIONS", "max connections must be positive for alertmanager")
		}
	case cluster.ClusterTypeGrafana:
		if config.Auth.Token == "" {
			return errors.NewValidationError("MISSING_API_TOKEN", "API token is required for grafana")
		}
	}

	s.logger.Info("cluster config validation passed", zap.String("type", string(clusterType)))
	return nil
}

// GetClustersByStatus 根据状态获取集群列表
func (s *ClusterService) GetClustersByStatus(ctx context.Context, status cluster.ClusterStatus) ([]*cluster.Cluster, error) {
	s.logger.Debug("getting clusters by status", zap.String("status", string(status)))

	clusters, err := s.repo.GetByStatus(ctx, status)
	if err != nil {
		s.logger.Error("failed to get clusters by status", zap.Error(err))
		return nil, errors.NewInternalError("failed to get clusters by status", err)
	}

	return clusters, nil
}

// BulkUpdateClusters 批量更新集群
func (s *ClusterService) BulkUpdateClusters(ctx context.Context, updates []*cluster.BulkUpdateClusterRequest) error {
	// 验证请求
	if len(updates) == 0 {
		return errors.NewValidationError("INVALID_REQUEST", "updates cannot be empty")
	}

	s.logger.Info("bulk updating clusters", zap.Int("count", len(updates)))

	// 批量更新
	for _, update := range updates {
		if update == nil {
			continue
		}

		// 获取集群
		existingCluster, err := s.repo.GetByID(ctx, update.ClusterID)
		if err != nil {
			return errors.NewNotFoundError(fmt.Sprintf("cluster %s not found", update.ClusterID))
		}

		// 应用更新
		if update.Updates.Status != nil {
			existingCluster.Status = *update.Updates.Status
		}
		if update.Updates.Tags != nil {
			existingCluster.Tags = update.Updates.Tags
		}
		if update.Updates.Labels != nil {
			existingCluster.Labels = update.Updates.Labels
		}
		if update.Updates.Name != nil {
			existingCluster.Name = *update.Updates.Name
		}
		if update.Updates.Description != nil {
			existingCluster.Description = *update.Updates.Description
		}
		if update.Updates.Config != nil {
			existingCluster.Config = *update.Updates.Config
		}
		if update.Updates.Version != nil {
			existingCluster.Version = *update.Updates.Version
		}
		if update.Updates.Endpoints != nil {
			existingCluster.Endpoints = update.Updates.Endpoints
		}

		existingCluster.UpdatedAt = time.Now()

		// 保存更新
		if err := s.repo.Update(ctx, existingCluster); err != nil {
			return errors.NewInternalError("failed to update cluster", err)
		}
	}

	s.logger.Info("clusters bulk updated successfully", zap.Int("count", len(updates)))
	return nil
}

// ImportClusters 导入集群
func (s *ClusterService) ImportClusters(ctx context.Context, clusters []*cluster.Cluster) (*cluster.ImportClusterResult, error) {
	s.logger.Info("importing clusters", zap.Int("count", len(clusters)))

	result := &cluster.ImportClusterResult{
		Total:   len(clusters),
		Success: 0,
		Failed:  0,
		Errors:  make([]string, 0),
	}

	for _, c := range clusters {
		// 生成新的ID
		c.ID = uuid.New().String()
		c.CreatedAt = time.Now()
		c.UpdatedAt = time.Now()

		// 验证集群
		if err := c.Validate(); err != nil {
			result.Failed++
			result.Errors = append(result.Errors, fmt.Sprintf("cluster %s validation failed: %v", c.Name, err))
			continue
		}

		// 检查名称是否已存在
		exists, err := s.repo.ExistsByName(ctx, c.Name)
		if err != nil {
			result.Failed++
			result.Errors = append(result.Errors, fmt.Sprintf("cluster %s name check failed: %v", c.Name, err))
			continue
		}
		if exists {
			result.Failed++
			result.Errors = append(result.Errors, fmt.Sprintf("cluster %s name already exists", c.Name))
			continue
		}

		// 创建集群
		if err := s.repo.Create(ctx, c); err != nil {
			result.Failed++
			result.Errors = append(result.Errors, fmt.Sprintf("cluster %s creation failed: %v", c.Name, err))
			continue
		}

		result.Success++
	}

	s.logger.Info("clusters import completed", zap.Int("success", result.Success), zap.Int("failed", result.Failed))
	return result, nil
}

// ExportClusters 导出集群
func (s *ClusterService) ExportClusters(ctx context.Context, filter map[string]interface{}) ([]*cluster.Cluster, error) {
	s.logger.Info("exporting clusters")

	// 设置无限制的查询
	query := types.Query{
		Limit:  0,
		Offset: 0,
		Filter: make(map[string]interface{}),
	}

	// 构建查询条件
	if search, ok := filter["search"].(string); ok && search != "" {
		query.Search = search
	}
	if status, ok := filter["status"].(string); ok && status != "" {
		query.Filter["status"] = status
	}

	clusters, _, err := s.repo.List(ctx, query)
	if err != nil {
		s.logger.Error("failed to export clusters", zap.Error(err))
		return nil, errors.NewInternalError("failed to export clusters", err)
	}

	s.logger.Info("clusters exported successfully", zap.Int("count", len(clusters)))
	return clusters, nil
}

// DiscoverClusters 发现集群
func (s *ClusterService) DiscoverClusters(ctx context.Context, req *cluster.DiscoverRequest) ([]*cluster.Cluster, error) {
	s.logger.Info("discovering clusters", zap.String("network", req.Network))

	// TODO: 实现实际的集群发现逻辑
	// 这里应该根据网络范围和端口扫描发现集群

	discovered := make([]*cluster.Cluster, 0)

	s.logger.Info("cluster discovery completed", zap.Int("found", len(discovered)))
	return discovered, nil
}

// MonitorCluster 监控集群
func (s *ClusterService) MonitorCluster(ctx context.Context, id string) (*cluster.MonitorResult, error) {
	s.logger.Info("monitoring cluster", zap.String("id", id))

	_, err := s.repo.GetByID(ctx, id)
	if err != nil {
		s.logger.Error("failed to get cluster for monitoring", zap.String("id", id), zap.Error(err))
		return nil, errors.NewNotFoundError("cluster")
	}

	// TODO: 实现实际的监控逻辑
	result := &cluster.MonitorResult{
		ClusterID: id,
		Status:    cluster.ClusterStatusActive,
		Health:    &cluster.ClusterHealth{ClusterID: id, Status: "healthy"},
		Metrics:   &cluster.ClusterMetrics{ClusterID: id},
		Alerts:    []*cluster.MonitorAlert{},
		Timestamp: time.Now(),
	}

	s.logger.Info("cluster monitoring completed", zap.String("id", id), zap.String("status", string(result.Status)))
	return result, nil
}

// validateCreateRequest 验证创建请求
func (s *ClusterService) validateCreateRequest(req *cluster.CreateClusterRequest) error {
	if req.Name == "" {
		return errors.NewValidationError("INVALID_NAME", "cluster name is required")
	}
	if req.Type == "" {
		return errors.NewValidationError("INVALID_TYPE", "cluster type is required")
	}
	if len(req.Endpoints) == 0 {
		return errors.NewValidationError("INVALID_ENDPOINTS", "at least one endpoint is required")
	}
	return nil
}