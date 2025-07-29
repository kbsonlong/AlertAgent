package repository

import (
	"context"
	"fmt"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"alert_agent/internal/domain/channel"
	"alert_agent/internal/infrastructure/database"
)

// ChannelRepositoryImpl 渠道仓储实现
type ChannelRepositoryImpl struct {
	db     *gorm.DB
	logger *zap.Logger
}

// NewChannelRepository 创建渠道仓储
func NewChannelRepository(db *gorm.DB, logger *zap.Logger) channel.ChannelRepository {
	return &ChannelRepositoryImpl{
		db:     db,
		logger: logger,
	}
}

// Create 创建渠道
func (r *ChannelRepositoryImpl) Create(ctx context.Context, ch *channel.Channel) error {
	dbChannel := r.toDBModel(ch)
	if err := r.db.WithContext(ctx).Create(dbChannel).Error; err != nil {
		return fmt.Errorf("failed to create channel: %w", err)
	}
	return nil
}

// Update 更新渠道
func (r *ChannelRepositoryImpl) Update(ctx context.Context, ch *channel.Channel) error {
	dbChannel := r.toDBModel(ch)
	if err := r.db.WithContext(ctx).Save(dbChannel).Error; err != nil {
		return fmt.Errorf("failed to update channel: %w", err)
	}
	return nil
}

// Delete 删除渠道
func (r *ChannelRepositoryImpl) Delete(ctx context.Context, id string) error {
	if err := r.db.WithContext(ctx).Delete(&database.Channel{}, "id = ?", id).Error; err != nil {
		return fmt.Errorf("failed to delete channel: %w", err)
	}
	return nil
}

// GetByID 根据ID获取渠道
func (r *ChannelRepositoryImpl) GetByID(ctx context.Context, id string) (*channel.Channel, error) {
	var dbChannel database.Channel
	if err := r.db.WithContext(ctx).First(&dbChannel, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("channel not found")
		}
		return nil, fmt.Errorf("failed to get channel: %w", err)
	}
	return r.toDomainModel(&dbChannel), nil
}

// List 列出渠道
func (r *ChannelRepositoryImpl) List(ctx context.Context, query *channel.ChannelQuery) ([]*channel.Channel, int64, error) {
	db := r.db.WithContext(ctx).Model(&database.Channel{})

	// 应用过滤条件
	if query.Name != "" {
		db = db.Where("name LIKE ?", "%"+query.Name+"%")
	}
	if query.Type != "" {
		db = db.Where("type = ?", query.Type)
	}
	if query.GroupID != "" {
		db = db.Where("group_id = ?", query.GroupID)
	}
	if query.Status != "" {
		db = db.Where("status = ?", query.Status)
	}

	// 计算总数
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count channels: %w", err)
	}

	// 应用分页
	if query.Page > 0 && query.PageSize > 0 {
		offset := (query.Page - 1) * query.PageSize
		db = db.Offset(offset).Limit(query.PageSize)
	}

	// 应用排序
	if query.SortBy != "" {
		order := query.SortBy
		if query.SortDesc {
			order += " DESC"
		}
		db = db.Order(order)
	} else {
		db = db.Order("created_at DESC")
	}

	// 查询数据
	var dbChannels []database.Channel
	if err := db.Find(&dbChannels).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list channels: %w", err)
	}

	// 转换为领域模型
	channels := make([]*channel.Channel, len(dbChannels))
	for i, dbChannel := range dbChannels {
		channels[i] = r.toDomainModel(&dbChannel)
	}

	return channels, total, nil
}

// GetByType 根据类型获取渠道
func (r *ChannelRepositoryImpl) GetByType(ctx context.Context, channelType string) ([]*channel.Channel, error) {
	var dbChannels []database.Channel
	if err := r.db.WithContext(ctx).Where("type = ?", channelType).Find(&dbChannels).Error; err != nil {
		return nil, fmt.Errorf("failed to get channels by type: %w", err)
	}

	channels := make([]*channel.Channel, len(dbChannels))
	for i, dbChannel := range dbChannels {
		channels[i] = r.toDomainModel(&dbChannel)
	}

	return channels, nil
}

// GetByGroupID 根据分组ID获取渠道
func (r *ChannelRepositoryImpl) GetByGroupID(ctx context.Context, groupID string) ([]*channel.Channel, error) {
	var dbChannels []database.Channel
	if err := r.db.WithContext(ctx).Where("group_id = ?", groupID).Find(&dbChannels).Error; err != nil {
		return nil, fmt.Errorf("failed to get channels by group: %w", err)
	}

	channels := make([]*channel.Channel, len(dbChannels))
	for i, dbChannel := range dbChannels {
		channels[i] = r.toDomainModel(&dbChannel)
	}

	return channels, nil
}

// GetByTags 根据标签获取渠道
func (r *ChannelRepositoryImpl) GetByTags(ctx context.Context, tags []string) ([]*channel.Channel, error) {
	// 这里需要根据具体的数据库实现来查询JSON数组
	// 简化实现，实际应该使用JSON查询
	var dbChannels []database.Channel
	if err := r.db.WithContext(ctx).Find(&dbChannels).Error; err != nil {
		return nil, fmt.Errorf("failed to get channels by tags: %w", err)
	}

	// 过滤包含指定标签的渠道
	var filteredChannels []*channel.Channel
	for _, dbChannel := range dbChannels {
		if r.containsTags(dbChannel.Tags, tags) {
			filteredChannels = append(filteredChannels, r.toDomainModel(&dbChannel))
		}
	}

	return filteredChannels, nil
}

// UpdateHealthStatus 更新健康状态
func (r *ChannelRepositoryImpl) UpdateHealthStatus(ctx context.Context, id string, status channel.HealthStatus, responseTime int) error {
	updates := map[string]interface{}{
		"health_status":      string(status),
		"response_time":      responseTime,
		"last_health_check":  gorm.Expr("NOW()"),
	}

	if err := r.db.WithContext(ctx).Model(&database.Channel{}).Where("id = ?", id).Updates(updates).Error; err != nil {
		return fmt.Errorf("failed to update health status: %w", err)
	}

	return nil
}

// GetUnhealthyChannels 获取不健康的渠道
func (r *ChannelRepositoryImpl) GetUnhealthyChannels(ctx context.Context) ([]*channel.Channel, error) {
	var dbChannels []database.Channel
	if err := r.db.WithContext(ctx).Where("health_status = ?", "unhealthy").Find(&dbChannels).Error; err != nil {
		return nil, fmt.Errorf("failed to get unhealthy channels: %w", err)
	}

	channels := make([]*channel.Channel, len(dbChannels))
	for i, dbChannel := range dbChannels {
		channels[i] = r.toDomainModel(&dbChannel)
	}

	return channels, nil
}

// CreateGroup 创建分组
func (r *ChannelRepositoryImpl) CreateGroup(ctx context.Context, group *channel.ChannelGroup) error {
	dbGroup := &database.ChannelGroup{
		ID:          group.ID,
		Name:        group.Name,
		Description: group.Description,
		ParentID:    group.ParentID,
		Path:        group.Path,
		Level:       group.Level,
		SortOrder:   group.SortOrder,
		CreatedAt:   group.CreatedAt,
		UpdatedAt:   group.UpdatedAt,
	}

	if err := r.db.WithContext(ctx).Create(dbGroup).Error; err != nil {
		return fmt.Errorf("failed to create channel group: %w", err)
	}

	return nil
}

// UpdateGroup 更新分组
func (r *ChannelRepositoryImpl) UpdateGroup(ctx context.Context, group *channel.ChannelGroup) error {
	dbGroup := &database.ChannelGroup{
		ID:          group.ID,
		Name:        group.Name,
		Description: group.Description,
		ParentID:    group.ParentID,
		Path:        group.Path,
		Level:       group.Level,
		SortOrder:   group.SortOrder,
		UpdatedAt:   group.UpdatedAt,
	}

	if err := r.db.WithContext(ctx).Save(dbGroup).Error; err != nil {
		return fmt.Errorf("failed to update channel group: %w", err)
	}

	return nil
}

// DeleteGroup 删除分组
func (r *ChannelRepositoryImpl) DeleteGroup(ctx context.Context, id string) error {
	if err := r.db.WithContext(ctx).Delete(&database.ChannelGroup{}, "id = ?", id).Error; err != nil {
		return fmt.Errorf("failed to delete channel group: %w", err)
	}
	return nil
}

// GetGroupByID 根据ID获取分组
func (r *ChannelRepositoryImpl) GetGroupByID(ctx context.Context, id string) (*channel.ChannelGroup, error) {
	var dbGroup database.ChannelGroup
	if err := r.db.WithContext(ctx).First(&dbGroup, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("channel group not found")
		}
		return nil, fmt.Errorf("failed to get channel group: %w", err)
	}

	return &channel.ChannelGroup{
		ID:          dbGroup.ID,
		Name:        dbGroup.Name,
		Description: dbGroup.Description,
		ParentID:    dbGroup.ParentID,
		Path:        dbGroup.Path,
		Level:       dbGroup.Level,
		SortOrder:   dbGroup.SortOrder,
		CreatedAt:   dbGroup.CreatedAt,
		UpdatedAt:   dbGroup.UpdatedAt,
	}, nil
}

// ListGroups 列出分组
func (r *ChannelRepositoryImpl) ListGroups(ctx context.Context) ([]*channel.ChannelGroup, error) {
	var dbGroups []database.ChannelGroup
	if err := r.db.WithContext(ctx).Order("path, sort_order").Find(&dbGroups).Error; err != nil {
		return nil, fmt.Errorf("failed to list channel groups: %w", err)
	}

	groups := make([]*channel.ChannelGroup, len(dbGroups))
	for i, dbGroup := range dbGroups {
		groups[i] = &channel.ChannelGroup{
			ID:          dbGroup.ID,
			Name:        dbGroup.Name,
			Description: dbGroup.Description,
			ParentID:    dbGroup.ParentID,
			Path:        dbGroup.Path,
			Level:       dbGroup.Level,
			SortOrder:   dbGroup.SortOrder,
			CreatedAt:   dbGroup.CreatedAt,
			UpdatedAt:   dbGroup.UpdatedAt,
		}
	}

	return groups, nil
}

// toDBModel 转换为数据库模型
func (r *ChannelRepositoryImpl) toDBModel(ch *channel.Channel) *database.Channel {
	return &database.Channel{
		ID:              ch.ID,
		Name:            ch.Name,
		Type:            ch.Type,
		Description:     ch.Description,
		Config:          ch.Config,
		GroupID:         ch.GroupID,
		Tags:            ch.Tags,
		Status:          string(ch.Status),
		HealthStatus:    string(ch.HealthStatus),
		LastHealthCheck: ch.LastHealthCheck,
		ResponseTime:    ch.ResponseTime,
		SuccessRate:     ch.SuccessRate,
		CreatedBy:       ch.CreatedBy,
		UpdatedBy:       ch.UpdatedBy,
		CreatedAt:       ch.CreatedAt,
		UpdatedAt:       ch.UpdatedAt,
	}
}

// toDomainModel 转换为领域模型
func (r *ChannelRepositoryImpl) toDomainModel(dbChannel *database.Channel) *channel.Channel {
	return &channel.Channel{
		ID:              dbChannel.ID,
		Name:            dbChannel.Name,
		Type:            dbChannel.Type,
		Description:     dbChannel.Description,
		Config:          dbChannel.Config,
		GroupID:         dbChannel.GroupID,
		Tags:            dbChannel.Tags,
		Status:          channel.ChannelStatus(dbChannel.Status),
		HealthStatus:    channel.HealthStatus(dbChannel.HealthStatus),
		LastHealthCheck: dbChannel.LastHealthCheck,
		ResponseTime:    dbChannel.ResponseTime,
		SuccessRate:     dbChannel.SuccessRate,
		CreatedBy:       dbChannel.CreatedBy,
		UpdatedBy:       dbChannel.UpdatedBy,
		CreatedAt:       dbChannel.CreatedAt,
		UpdatedAt:       dbChannel.UpdatedAt,
	}
}

// containsTags 检查是否包含指定标签
func (r *ChannelRepositoryImpl) containsTags(channelTags, searchTags []string) bool {
	if len(searchTags) == 0 {
		return true
	}

	tagMap := make(map[string]bool)
	for _, tag := range channelTags {
		tagMap[tag] = true
	}

	for _, searchTag := range searchTags {
		if !tagMap[searchTag] {
			return false
		}
	}

	return true
}