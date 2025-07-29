package channel

import (
	"context"

	"alert_agent/pkg/types"
)

// Service 通道服务接口
type Service interface {
	// CreateChannel 创建通道
	CreateChannel(ctx context.Context, req *CreateChannelRequest) (*Channel, error)

	// GetChannel 获取通道
	GetChannel(ctx context.Context, id string) (*Channel, error)

	// UpdateChannel 更新通道
	UpdateChannel(ctx context.Context, id string, req *UpdateChannelRequest) (*Channel, error)

	// DeleteChannel 删除通道
	DeleteChannel(ctx context.Context, id string) error

	// ListChannels 获取通道列表
	ListChannels(ctx context.Context, query types.Query) (*types.PageResult, error)

	// TestChannel 测试通道连接
	TestChannel(ctx context.Context, id string) (*TestResult, error)

	// SendMessage 发送消息
	SendMessage(ctx context.Context, channelID string, message *types.Message) error

	// GetChannelsByType 根据类型获取通道
	GetChannelsByType(ctx context.Context, channelType ChannelType) ([]*Channel, error)

	// GetActiveChannels 获取激活的通道
	GetActiveChannels(ctx context.Context) ([]*Channel, error)

	// UpdateChannelStatus 更新通道状态
	UpdateChannelStatus(ctx context.Context, id string, status ChannelStatus) error

	// ValidateChannelConfig 验证通道配置
	ValidateChannelConfig(ctx context.Context, channelType ChannelType, config ChannelConfig) error

	// GetChannelStats 获取通道统计信息
	GetChannelStats(ctx context.Context, id string) (*ChannelStats, error)

	// BulkUpdateChannels 批量更新通道
	BulkUpdateChannels(ctx context.Context, updates []*BulkUpdateRequest) error

	// ImportChannels 导入通道
	ImportChannels(ctx context.Context, channels []*Channel) (*ImportResult, error)

	// ExportChannels 导出通道
	ExportChannels(ctx context.Context, filter map[string]interface{}) ([]*Channel, error)
}

// CreateChannelRequest 创建通道请求
type CreateChannelRequest struct {
	Name        string            `json:"name" validate:"required,min=1,max=255"`
	Type        ChannelType       `json:"type" validate:"required"`
	Description string            `json:"description" validate:"max=1000"`
	Config      ChannelConfig     `json:"config" validate:"required"`
	Priority    int               `json:"priority" validate:"min=0,max=100"`
	Tags        []string          `json:"tags"`
	Labels      map[string]string `json:"labels"`
}

// UpdateChannelRequest 更新通道请求
type UpdateChannelRequest struct {
	Name        *string            `json:"name,omitempty" validate:"omitempty,min=1,max=255"`
	Description *string            `json:"description,omitempty" validate:"omitempty,max=1000"`
	Config      *ChannelConfig     `json:"config,omitempty"`
	Status      *ChannelStatus     `json:"status,omitempty"`
	Priority    *int               `json:"priority,omitempty" validate:"omitempty,min=0,max=100"`
	Tags        []string           `json:"tags,omitempty"`
	Labels      map[string]string  `json:"labels,omitempty"`
}

// TestResult 测试结果
type TestResult struct {
	Success   bool              `json:"success"`
	Message   string            `json:"message"`
	Latency   int64             `json:"latency"` // 毫秒
	Details   map[string]interface{} `json:"details"`
	Timestamp int64             `json:"timestamp"`
}

// ChannelStats 通道统计信息
type ChannelStats struct {
	ChannelID     string  `json:"channel_id"`
	TotalSent     int64   `json:"total_sent"`
	TotalFailed   int64   `json:"total_failed"`
	SuccessRate   float64 `json:"success_rate"`
	AvgLatency    float64 `json:"avg_latency"`
	LastSentAt    int64   `json:"last_sent_at"`
	LastErrorAt   int64   `json:"last_error_at"`
	LastError     string  `json:"last_error"`
	DailySent     int64   `json:"daily_sent"`
	DailyFailed   int64   `json:"daily_failed"`
}

// BulkUpdateRequest 批量更新请求
type BulkUpdateRequest struct {
	ChannelID string             `json:"channel_id"`
	Updates   UpdateChannelRequest `json:"updates"`
}

// ImportResult 导入结果
type ImportResult struct {
	Total     int      `json:"total"`
	Success   int      `json:"success"`
	Failed    int      `json:"failed"`
	Errors    []string `json:"errors"`
	Created   []string `json:"created"`
	Updated   []string `json:"updated"`
	Skipped   []string `json:"skipped"`
}