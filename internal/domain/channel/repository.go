package channel

import (
	"context"

	"alert_agent/pkg/types"
)

// Repository 通道仓储接口
type Repository interface {
	// Create 创建通道
	Create(ctx context.Context, channel *Channel) error

	// GetByID 根据ID获取通道
	GetByID(ctx context.Context, id string) (*Channel, error)

	// GetByName 根据名称获取通道
	GetByName(ctx context.Context, name string) (*Channel, error)

	// List 获取通道列表
	List(ctx context.Context, query types.Query) ([]*Channel, int64, error)

	// Update 更新通道
	Update(ctx context.Context, channel *Channel) error

	// Delete 删除通道
	Delete(ctx context.Context, id string) error

	// GetByType 根据类型获取通道列表
	GetByType(ctx context.Context, channelType ChannelType) ([]*Channel, error)

	// GetByStatus 根据状态获取通道列表
	GetByStatus(ctx context.Context, status ChannelStatus) ([]*Channel, error)

	// GetActiveChannels 获取激活的通道列表
	GetActiveChannels(ctx context.Context) ([]*Channel, error)

	// GetByTags 根据标签获取通道列表
	GetByTags(ctx context.Context, tags []string) ([]*Channel, error)

	// GetByLabels 根据标签获取通道列表
	GetByLabels(ctx context.Context, labels map[string]string) ([]*Channel, error)

	// UpdateStatus 更新通道状态
	UpdateStatus(ctx context.Context, id string, status ChannelStatus) error

	// BatchUpdate 批量更新通道
	BatchUpdate(ctx context.Context, channels []*Channel) error

	// Count 获取通道总数
	Count(ctx context.Context, filter map[string]interface{}) (int64, error)

	// Exists 检查通道是否存在
	Exists(ctx context.Context, id string) (bool, error)

	// ExistsByName 检查通道名称是否存在
	ExistsByName(ctx context.Context, name string) (bool, error)
}