package repository

import (
	"context"
	"fmt"
	"strings"

	"gorm.io/gorm"
	"go.uber.org/zap"

	"alert_agent/internal/domain/channel"
	"alert_agent/internal/shared/logger"
	"alert_agent/pkg/types"
)

// ChannelRepository 通道仓储实现
type ChannelRepository struct {
	db     *gorm.DB
	logger *zap.Logger
}

// NewChannelRepository 创建通道仓储
func NewChannelRepository(db *gorm.DB) channel.Repository {
	return &ChannelRepository{
		db:     db,
		logger: logger.WithComponent("channel-repository"),
	}
}

// Create 创建通道
func (r *ChannelRepository) Create(ctx context.Context, ch *channel.Channel) error {
	r.logger.Debug("creating channel", zap.String("id", ch.ID), zap.String("name", ch.Name))

	if err := r.db.WithContext(ctx).Create(ch).Error; err != nil {
		r.logger.Error("failed to create channel", zap.Error(err))
		return fmt.Errorf("failed to create channel: %w", err)
	}

	return nil
}

// GetByID 根据ID获取通道
func (r *ChannelRepository) GetByID(ctx context.Context, id string) (*channel.Channel, error) {
	r.logger.Debug("getting channel by id", zap.String("id", id))

	var ch channel.Channel
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&ch).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("channel not found")
		}
		r.logger.Error("failed to get channel by id", zap.Error(err))
		return nil, fmt.Errorf("failed to get channel: %w", err)
	}

	return &ch, nil
}

// GetByName 根据名称获取通道
func (r *ChannelRepository) GetByName(ctx context.Context, name string) (*channel.Channel, error) {
	r.logger.Debug("getting channel by name", zap.String("name", name))

	var ch channel.Channel
	if err := r.db.WithContext(ctx).Where("name = ?", name).First(&ch).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("channel not found")
		}
		r.logger.Error("failed to get channel by name", zap.Error(err))
		return nil, fmt.Errorf("failed to get channel: %w", err)
	}

	return &ch, nil
}

// List 获取通道列表
func (r *ChannelRepository) List(ctx context.Context, query types.Query) ([]*channel.Channel, int64, error) {
	r.logger.Debug("listing channels", zap.Int("limit", query.Limit), zap.Int("offset", query.Offset))

	var channels []*channel.Channel
	var total int64

	// 构建查询
	db := r.db.WithContext(ctx).Model(&channel.Channel{})

	// 应用过滤器
	db = r.applyFilters(db, query.Filter)

	// 应用搜索
	if query.Search != "" {
		db = db.Where("name LIKE ? OR description LIKE ?", "%"+query.Search+"%", "%"+query.Search+"%")
	}

	// 获取总数
	if err := db.Count(&total).Error; err != nil {
		r.logger.Error("failed to count channels", zap.Error(err))
		return nil, 0, fmt.Errorf("failed to count channels: %w", err)
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
	if err := db.Find(&channels).Error; err != nil {
		r.logger.Error("failed to list channels", zap.Error(err))
		return nil, 0, fmt.Errorf("failed to list channels: %w", err)
	}

	return channels, total, nil
}

// Update 更新通道
func (r *ChannelRepository) Update(ctx context.Context, ch *channel.Channel) error {
	r.logger.Debug("updating channel", zap.String("id", ch.ID), zap.String("name", ch.Name))

	if err := r.db.WithContext(ctx).Save(ch).Error; err != nil {
		r.logger.Error("failed to update channel", zap.Error(err))
		return fmt.Errorf("failed to update channel: %w", err)
	}

	return nil
}

// Delete 删除通道
func (r *ChannelRepository) Delete(ctx context.Context, id string) error {
	r.logger.Debug("deleting channel", zap.String("id", id))

	if err := r.db.WithContext(ctx).Where("id = ?", id).Delete(&channel.Channel{}).Error; err != nil {
		r.logger.Error("failed to delete channel", zap.Error(err))
		return fmt.Errorf("failed to delete channel: %w", err)
	}

	return nil
}

// GetByType 根据类型获取通道列表
func (r *ChannelRepository) GetByType(ctx context.Context, channelType channel.ChannelType) ([]*channel.Channel, error) {
	r.logger.Debug("getting channels by type", zap.String("type", string(channelType)))

	var channels []*channel.Channel
	if err := r.db.WithContext(ctx).Where("type = ?", channelType).Find(&channels).Error; err != nil {
		r.logger.Error("failed to get channels by type", zap.Error(err))
		return nil, fmt.Errorf("failed to get channels by type: %w", err)
	}

	return channels, nil
}

// GetByStatus 根据状态获取通道列表
func (r *ChannelRepository) GetByStatus(ctx context.Context, status channel.ChannelStatus) ([]*channel.Channel, error) {
	r.logger.Debug("getting channels by status", zap.String("status", string(status)))

	var channels []*channel.Channel
	if err := r.db.WithContext(ctx).Where("status = ?", status).Find(&channels).Error; err != nil {
		r.logger.Error("failed to get channels by status", zap.Error(err))
		return nil, fmt.Errorf("failed to get channels by status: %w", err)
	}

	return channels, nil
}

// GetActiveChannels 获取激活的通道列表
func (r *ChannelRepository) GetActiveChannels(ctx context.Context) ([]*channel.Channel, error) {
	r.logger.Debug("getting active channels")

	return r.GetByStatus(ctx, channel.ChannelStatusActive)
}

// GetByTags 根据标签获取通道列表
func (r *ChannelRepository) GetByTags(ctx context.Context, tags []string) ([]*channel.Channel, error) {
	r.logger.Debug("getting channels by tags", zap.Strings("tags", tags))

	var channels []*channel.Channel
	db := r.db.WithContext(ctx)

	// 构建标签查询
	for _, tag := range tags {
		db = db.Where("JSON_CONTAINS(tags, ?)", fmt.Sprintf(`"%s"`, tag))
	}

	if err := db.Find(&channels).Error; err != nil {
		r.logger.Error("failed to get channels by tags", zap.Error(err))
		return nil, fmt.Errorf("failed to get channels by tags: %w", err)
	}

	return channels, nil
}

// GetByLabels 根据标签获取通道列表
func (r *ChannelRepository) GetByLabels(ctx context.Context, labels map[string]string) ([]*channel.Channel, error) {
	r.logger.Debug("getting channels by labels", zap.Any("labels", labels))

	var channels []*channel.Channel
	db := r.db.WithContext(ctx)

	// 构建标签查询
	for key, value := range labels {
		db = db.Where("JSON_EXTRACT(labels, ?) = ?", fmt.Sprintf("$.%s", key), value)
	}

	if err := db.Find(&channels).Error; err != nil {
		r.logger.Error("failed to get channels by labels", zap.Error(err))
		return nil, fmt.Errorf("failed to get channels by labels: %w", err)
	}

	return channels, nil
}

// UpdateStatus 更新通道状态
func (r *ChannelRepository) UpdateStatus(ctx context.Context, id string, status channel.ChannelStatus) error {
	r.logger.Debug("updating channel status", zap.String("id", id), zap.String("status", string(status)))

	if err := r.db.WithContext(ctx).Model(&channel.Channel{}).Where("id = ?", id).Update("status", status).Error; err != nil {
		r.logger.Error("failed to update channel status", zap.Error(err))
		return fmt.Errorf("failed to update channel status: %w", err)
	}

	return nil
}

// BatchUpdate 批量更新通道
func (r *ChannelRepository) BatchUpdate(ctx context.Context, channels []*channel.Channel) error {
	r.logger.Debug("batch updating channels", zap.Int("count", len(channels)))

	tx := r.db.WithContext(ctx).Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	for _, ch := range channels {
		if err := tx.Save(ch).Error; err != nil {
			r.logger.Error("failed to update channel in batch", zap.String("id", ch.ID), zap.Error(err))
			tx.Rollback()
			return fmt.Errorf("failed to update channel %s: %w", ch.ID, err)
		}
	}

	if err := tx.Commit().Error; err != nil {
		r.logger.Error("failed to commit batch update", zap.Error(err))
		return fmt.Errorf("failed to commit batch update: %w", err)
	}

	return nil
}

// Count 获取通道总数
func (r *ChannelRepository) Count(ctx context.Context, filter map[string]interface{}) (int64, error) {
	r.logger.Debug("counting channels")

	var count int64
	db := r.db.WithContext(ctx).Model(&channel.Channel{})

	// 应用过滤器
	db = r.applyFilters(db, filter)

	if err := db.Count(&count).Error; err != nil {
		r.logger.Error("failed to count channels", zap.Error(err))
		return 0, fmt.Errorf("failed to count channels: %w", err)
	}

	return count, nil
}

// Exists 检查通道是否存在
func (r *ChannelRepository) Exists(ctx context.Context, id string) (bool, error) {
	r.logger.Debug("checking channel existence", zap.String("id", id))

	var count int64
	if err := r.db.WithContext(ctx).Model(&channel.Channel{}).Where("id = ?", id).Count(&count).Error; err != nil {
		r.logger.Error("failed to check channel existence", zap.Error(err))
		return false, fmt.Errorf("failed to check channel existence: %w", err)
	}

	return count > 0, nil
}

// ExistsByName 检查通道名称是否存在
func (r *ChannelRepository) ExistsByName(ctx context.Context, name string) (bool, error) {
	r.logger.Debug("checking channel name existence", zap.String("name", name))

	var count int64
	if err := r.db.WithContext(ctx).Model(&channel.Channel{}).Where("name = ?", name).Count(&count).Error; err != nil {
		r.logger.Error("failed to check channel name existence", zap.Error(err))
		return false, fmt.Errorf("failed to check channel name existence: %w", err)
	}

	return count > 0, nil
}

// applyFilters 应用过滤器
func (r *ChannelRepository) applyFilters(db *gorm.DB, filter map[string]interface{}) *gorm.DB {
	if filter == nil {
		return db
	}

	for key, value := range filter {
		switch key {
		case "type":
			db = db.Where("type = ?", value)
		case "status":
			db = db.Where("status = ?", value)
		case "priority":
			db = db.Where("priority = ?", value)
		case "priority_gte":
			db = db.Where("priority >= ?", value)
		case "priority_lte":
			db = db.Where("priority <= ?", value)
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