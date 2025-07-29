package repository

import (
	"context"
	"fmt"
	"strings"

	"gorm.io/gorm"
	"go.uber.org/zap"

	"alert_agent/internal/domain/cluster"
	"alert_agent/internal/shared/logger"
	"alert_agent/pkg/types"
)

// ClusterRepository 集群仓储实现
type ClusterRepository struct {
	db     *gorm.DB
	logger *zap.Logger
}

// NewClusterRepository 创建集群仓储
func NewClusterRepository(db *gorm.DB) cluster.Repository {
	return &ClusterRepository{
		db:     db,
		logger: logger.WithComponent("cluster-repository"),
	}
}

// Create 创建集群
func (r *ClusterRepository) Create(ctx context.Context, c *cluster.Cluster) error {
	r.logger.Debug("creating cluster", zap.String("id", c.ID), zap.String("name", c.Name))

	if err := r.db.WithContext(ctx).Create(c).Error; err != nil {
		r.logger.Error("failed to create cluster", zap.Error(err))
		return fmt.Errorf("failed to create cluster: %w", err)
	}

	return nil
}

// GetByID 根据ID获取集群
func (r *ClusterRepository) GetByID(ctx context.Context, id string) (*cluster.Cluster, error) {
	r.logger.Debug("getting cluster by id", zap.String("id", id))

	var c cluster.Cluster
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&c).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("cluster not found")
		}
		r.logger.Error("failed to get cluster by id", zap.Error(err))
		return nil, fmt.Errorf("failed to get cluster: %w", err)
	}

	return &c, nil
}

// GetByName 根据名称获取集群
func (r *ClusterRepository) GetByName(ctx context.Context, name string) (*cluster.Cluster, error) {
	r.logger.Debug("getting cluster by name", zap.String("name", name))

	var c cluster.Cluster
	if err := r.db.WithContext(ctx).Where("name = ?", name).First(&c).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("cluster not found")
		}
		r.logger.Error("failed to get cluster by name", zap.Error(err))
		return nil, fmt.Errorf("failed to get cluster: %w", err)
	}

	return &c, nil
}

// List 获取集群列表
func (r *ClusterRepository) List(ctx context.Context, query types.Query) ([]*cluster.Cluster, int64, error) {
	r.logger.Debug("listing clusters", zap.Int("limit", query.Limit), zap.Int("offset", query.Offset))

	var clusters []*cluster.Cluster
	var total int64

	// 构建查询
	db := r.db.WithContext(ctx).Model(&cluster.Cluster{})

	// 应用过滤器
	db = r.applyFilters(db, query.Filter)

	// 应用搜索
	if query.Search != "" {
		db = db.Where("name LIKE ? OR description LIKE ?", "%"+query.Search+"%", "%"+query.Search+"%")
	}

	// 获取总数
	if err := db.Count(&total).Error; err != nil {
		r.logger.Error("failed to count clusters", zap.Error(err))
		return nil, 0, fmt.Errorf("failed to count clusters: %w", err)
	}

	// 应用排序
	if query.OrderBy != "" {
		db = db.Order(query.OrderBy)
	} else {
		db = db.Order("created_at DESC")
	}

	// 应用分页
	if query.Limit > 0 {
		db = db.Limit(query.Limit)
	}
	if query.Offset > 0 {
		db = db.Offset(query.Offset)
	}

	// 执行查询
	if err := db.Find(&clusters).Error; err != nil {
		r.logger.Error("failed to list clusters", zap.Error(err))
		return nil, 0, fmt.Errorf("failed to list clusters: %w", err)
	}

	return clusters, total, nil
}

// Update 更新集群
func (r *ClusterRepository) Update(ctx context.Context, c *cluster.Cluster) error {
	r.logger.Debug("updating cluster", zap.String("id", c.ID), zap.String("name", c.Name))

	if err := r.db.WithContext(ctx).Save(c).Error; err != nil {
		r.logger.Error("failed to update cluster", zap.Error(err))
		return fmt.Errorf("failed to update cluster: %w", err)
	}

	return nil
}

// Delete 删除集群
func (r *ClusterRepository) Delete(ctx context.Context, id string) error {
	r.logger.Debug("deleting cluster", zap.String("id", id))

	if err := r.db.WithContext(ctx).Where("id = ?", id).Delete(&cluster.Cluster{}).Error; err != nil {
		r.logger.Error("failed to delete cluster", zap.Error(err))
		return fmt.Errorf("failed to delete cluster: %w", err)
	}

	return nil
}

// GetByType 根据类型获取集群列表
func (r *ClusterRepository) GetByType(ctx context.Context, clusterType cluster.ClusterType) ([]*cluster.Cluster, error) {
	r.logger.Debug("getting clusters by type", zap.String("type", string(clusterType)))

	var clusters []*cluster.Cluster
	if err := r.db.WithContext(ctx).Where("type = ?", clusterType).Find(&clusters).Error; err != nil {
		r.logger.Error("failed to get clusters by type", zap.Error(err))
		return nil, fmt.Errorf("failed to get clusters by type: %w", err)
	}

	return clusters, nil
}

// GetByStatus 根据状态获取集群列表
func (r *ClusterRepository) GetByStatus(ctx context.Context, status cluster.ClusterStatus) ([]*cluster.Cluster, error) {
	r.logger.Debug("getting clusters by status", zap.String("status", string(status)))

	var clusters []*cluster.Cluster
	if err := r.db.WithContext(ctx).Where("status = ?", status).Find(&clusters).Error; err != nil {
		r.logger.Error("failed to get clusters by status", zap.Error(err))
		return nil, fmt.Errorf("failed to get clusters by status: %w", err)
	}

	return clusters, nil
}

// GetActiveClusters 获取激活的集群列表
func (r *ClusterRepository) GetActiveClusters(ctx context.Context) ([]*cluster.Cluster, error) {
	r.logger.Debug("getting active clusters")

	return r.GetByStatus(ctx, cluster.ClusterStatusActive)
}

// GetByTags 根据标签获取集群列表
func (r *ClusterRepository) GetByTags(ctx context.Context, tags []string) ([]*cluster.Cluster, error) {
	r.logger.Debug("getting clusters by tags", zap.Strings("tags", tags))

	var clusters []*cluster.Cluster
	db := r.db.WithContext(ctx)

	// 构建标签查询
	for _, tag := range tags {
		db = db.Where("JSON_CONTAINS(tags, ?)", fmt.Sprintf(`"%s"`, tag))
	}

	if err := db.Find(&clusters).Error; err != nil {
		r.logger.Error("failed to get clusters by tags", zap.Error(err))
		return nil, fmt.Errorf("failed to get clusters by tags: %w", err)
	}

	return clusters, nil
}

// GetByLabels 根据标签获取集群列表
func (r *ClusterRepository) GetByLabels(ctx context.Context, labels map[string]string) ([]*cluster.Cluster, error) {
	r.logger.Debug("getting clusters by labels", zap.Any("labels", labels))

	var clusters []*cluster.Cluster
	db := r.db.WithContext(ctx)

	// 构建标签查询
	for key, value := range labels {
		db = db.Where("JSON_EXTRACT(labels, ?) = ?", fmt.Sprintf("$.%s", key), value)
	}

	if err := db.Find(&clusters).Error; err != nil {
		r.logger.Error("failed to get clusters by labels", zap.Error(err))
		return nil, fmt.Errorf("failed to get clusters by labels: %w", err)
	}

	return clusters, nil
}

// GetByEndpoint 根据端点获取集群
func (r *ClusterRepository) GetByEndpoint(ctx context.Context, endpoint string) (*cluster.Cluster, error) {
	r.logger.Debug("getting cluster by endpoint", zap.String("endpoint", endpoint))

	var c cluster.Cluster
	if err := r.db.WithContext(ctx).Where("JSON_CONTAINS(endpoints, ?)", fmt.Sprintf(`"%s"`, endpoint)).First(&c).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("cluster not found")
		}
		r.logger.Error("failed to get cluster by endpoint", zap.Error(err))
		return nil, fmt.Errorf("failed to get cluster: %w", err)
	}

	return &c, nil
}

// GetClustersByEndpoint 根据端点获取集群列表
func (r *ClusterRepository) GetClustersByEndpoint(ctx context.Context, endpoint string) ([]*cluster.Cluster, error) {
	r.logger.Debug("getting clusters by endpoint", zap.String("endpoint", endpoint))

	var clusters []*cluster.Cluster
	if err := r.db.WithContext(ctx).Where("JSON_CONTAINS(endpoints, ?)", fmt.Sprintf(`"%s"`, endpoint)).Find(&clusters).Error; err != nil {
		r.logger.Error("failed to get clusters by endpoint", zap.Error(err))
		return nil, fmt.Errorf("failed to get clusters by endpoint: %w", err)
	}

	return clusters, nil
}

// UpdateEndpoints 更新集群端点
func (r *ClusterRepository) UpdateEndpoints(ctx context.Context, id string, endpoints []string) error {
	r.logger.Debug("updating cluster endpoints", zap.String("id", id), zap.Strings("endpoints", endpoints))

	if err := r.db.WithContext(ctx).Model(&cluster.Cluster{}).Where("id = ?", id).Update("endpoints", endpoints).Error; err != nil {
		r.logger.Error("failed to update cluster endpoints", zap.Error(err))
		return fmt.Errorf("failed to update cluster endpoints: %w", err)
	}

	return nil
}

// GetHealthyClusters 获取健康的集群
func (r *ClusterRepository) GetHealthyClusters(ctx context.Context, clusterType cluster.ClusterType) ([]*cluster.Cluster, error) {
	r.logger.Debug("getting healthy clusters", zap.String("type", string(clusterType)))

	var clusters []*cluster.Cluster
	db := r.db.WithContext(ctx).Where("status = ?", cluster.ClusterStatusActive)

	if clusterType != "" {
		db = db.Where("type = ?", clusterType)
	}

	if err := db.Find(&clusters).Error; err != nil {
		r.logger.Error("failed to get healthy clusters", zap.Error(err))
		return nil, fmt.Errorf("failed to get healthy clusters: %w", err)
	}

	return clusters, nil
}

// UpdateStatus 更新集群状态
func (r *ClusterRepository) UpdateStatus(ctx context.Context, id string, status cluster.ClusterStatus) error {
	r.logger.Debug("updating cluster status", zap.String("id", id), zap.String("status", string(status)))

	if err := r.db.WithContext(ctx).Model(&cluster.Cluster{}).Where("id = ?", id).Update("status", status).Error; err != nil {
		r.logger.Error("failed to update cluster status", zap.Error(err))
		return fmt.Errorf("failed to update cluster status: %w", err)
	}

	return nil
}

// BatchUpdate 批量更新集群
func (r *ClusterRepository) BatchUpdate(ctx context.Context, clusters []*cluster.Cluster) error {
	r.logger.Debug("batch updating clusters", zap.Int("count", len(clusters)))

	tx := r.db.WithContext(ctx).Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	for _, c := range clusters {
		if err := tx.Save(c).Error; err != nil {
			r.logger.Error("failed to update cluster in batch", zap.String("id", c.ID), zap.Error(err))
			tx.Rollback()
			return fmt.Errorf("failed to update cluster %s: %w", c.ID, err)
		}
	}

	if err := tx.Commit().Error; err != nil {
		r.logger.Error("failed to commit batch update", zap.Error(err))
		return fmt.Errorf("failed to commit batch update: %w", err)
	}

	return nil
}

// Count 获取集群总数
func (r *ClusterRepository) Count(ctx context.Context, filter map[string]interface{}) (int64, error) {
	r.logger.Debug("counting clusters")

	var count int64
	db := r.db.WithContext(ctx).Model(&cluster.Cluster{})

	// 应用过滤器
	db = r.applyFilters(db, filter)

	if err := db.Count(&count).Error; err != nil {
		r.logger.Error("failed to count clusters", zap.Error(err))
		return 0, fmt.Errorf("failed to count clusters: %w", err)
	}

	return count, nil
}

// Exists 检查集群是否存在
func (r *ClusterRepository) Exists(ctx context.Context, id string) (bool, error) {
	r.logger.Debug("checking cluster existence", zap.String("id", id))

	var count int64
	if err := r.db.WithContext(ctx).Model(&cluster.Cluster{}).Where("id = ?", id).Count(&count).Error; err != nil {
		r.logger.Error("failed to check cluster existence", zap.Error(err))
		return false, fmt.Errorf("failed to check cluster existence: %w", err)
	}

	return count > 0, nil
}

// ExistsByName 检查集群名称是否存在
func (r *ClusterRepository) ExistsByName(ctx context.Context, name string) (bool, error) {
	r.logger.Debug("checking cluster name existence", zap.String("name", name))

	var count int64
	if err := r.db.WithContext(ctx).Model(&cluster.Cluster{}).Where("name = ?", name).Count(&count).Error; err != nil {
		r.logger.Error("failed to check cluster name existence", zap.Error(err))
		return false, fmt.Errorf("failed to check cluster name existence: %w", err)
	}

	return count > 0, nil
}

// applyFilters 应用过滤器
func (r *ClusterRepository) applyFilters(db *gorm.DB, filter map[string]interface{}) *gorm.DB {
	if filter == nil {
		return db
	}

	for key, value := range filter {
		switch key {
		case "type":
			db = db.Where("type = ?", value)
		case "status":
			db = db.Where("status = ?", value)
		case "version":
			db = db.Where("version = ?", value)
		case "version_gte":
			db = db.Where("version >= ?", value)
		case "version_lte":
			db = db.Where("version <= ?", value)
		case "created_after":
			db = db.Where("created_at > ?", value)
		case "created_before":
			db = db.Where("created_at < ?", value)
		case "updated_after":
			db = db.Where("updated_at > ?", value)
		case "updated_before":
			db = db.Where("updated_at < ?", value)
		case "tags":
			if tags, ok := value.([]string); ok {
				for _, tag := range tags {
					db = db.Where("JSON_CONTAINS(tags, ?)", fmt.Sprintf(`"%s"`, tag))
				}
			}
		case "labels":
			if labels, ok := value.(map[string]string); ok {
				for k, v := range labels {
					db = db.Where("JSON_EXTRACT(labels, ?) = ?", fmt.Sprintf("$.%s", k), v)
				}
			}
		case "name_like":
			db = db.Where("name LIKE ?", "%"+fmt.Sprintf("%v", value)+"%")
		case "description_like":
			db = db.Where("description LIKE ?", "%"+fmt.Sprintf("%v", value)+"%")
		case "endpoint":
			db = db.Where("JSON_CONTAINS(endpoints, ?)", fmt.Sprintf(`"%s"`, value))
		case "enabled":
			if enabled, ok := value.(bool); ok {
				db = db.Where("JSON_EXTRACT(config, '$.enabled') = ?", enabled)
			}
		default:
			// 对于未知的过滤器，尝试直接匹配字段
			if strings.Contains(key, ".") {
				// JSON字段查询
				parts := strings.SplitN(key, ".", 2)
				if len(parts) == 2 {
					db = db.Where(fmt.Sprintf("JSON_EXTRACT(%s, '$.%s') = ?", parts[0], parts[1]), value)
				}
			} else {
				// 普通字段查询
				db = db.Where(fmt.Sprintf("%s = ?", key), value)
			}
		}
	}

	return db
}