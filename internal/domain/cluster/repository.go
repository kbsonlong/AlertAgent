package cluster

import (
	"context"

	"alert_agent/pkg/types"
)

// Repository 集群仓储接口
type Repository interface {
	// Create 创建集群
	Create(ctx context.Context, cluster *Cluster) error

	// GetByID 根据ID获取集群
	GetByID(ctx context.Context, id string) (*Cluster, error)

	// GetByName 根据名称获取集群
	GetByName(ctx context.Context, name string) (*Cluster, error)

	// List 获取集群列表
	List(ctx context.Context, query types.Query) ([]*Cluster, int64, error)

	// Update 更新集群
	Update(ctx context.Context, cluster *Cluster) error

	// Delete 删除集群
	Delete(ctx context.Context, id string) error

	// GetByType 根据类型获取集群列表
	GetByType(ctx context.Context, clusterType ClusterType) ([]*Cluster, error)

	// GetByStatus 根据状态获取集群列表
	GetByStatus(ctx context.Context, status ClusterStatus) ([]*Cluster, error)

	// GetActiveClusters 获取激活的集群列表
	GetActiveClusters(ctx context.Context) ([]*Cluster, error)

	// GetByTags 根据标签获取集群列表
	GetByTags(ctx context.Context, tags []string) ([]*Cluster, error)

	// GetByLabels 根据标签获取集群列表
	GetByLabels(ctx context.Context, labels map[string]string) ([]*Cluster, error)

	// UpdateStatus 更新集群状态
	UpdateStatus(ctx context.Context, id string, status ClusterStatus) error

	// BatchUpdate 批量更新集群
	BatchUpdate(ctx context.Context, clusters []*Cluster) error

	// Count 获取集群总数
	Count(ctx context.Context, filter map[string]interface{}) (int64, error)

	// Exists 检查集群是否存在
	Exists(ctx context.Context, id string) (bool, error)

	// ExistsByName 检查集群名称是否存在
	ExistsByName(ctx context.Context, name string) (bool, error)

	// GetClustersByEndpoint 根据端点获取集群
	GetClustersByEndpoint(ctx context.Context, endpoint string) ([]*Cluster, error)

	// UpdateEndpoints 更新集群端点
	UpdateEndpoints(ctx context.Context, id string, endpoints []string) error

	// GetHealthyClusters 获取健康的集群
	GetHealthyClusters(ctx context.Context, clusterType ClusterType) ([]*Cluster, error)
}