package channel

import (
	"context"
	"time"

	"alert_agent/pkg/types"
)

// ChannelManager 渠道管理器接口
type ChannelManager interface {
	// 渠道生命周期管理
	RegisterPlugin(plugin ChannelPlugin) error
	UnregisterPlugin(channelType ChannelType) error
	GetPlugin(channelType ChannelType) (ChannelPlugin, error)
	ListPlugins() []ChannelPlugin

	// 渠道操作
	CreateChannel(ctx context.Context, req *CreateChannelRequest) (*Channel, error)
	UpdateChannel(ctx context.Context, id string, req *UpdateChannelRequest) (*Channel, error)
	DeleteChannel(ctx context.Context, id string) error
	GetChannel(ctx context.Context, id string) (*Channel, error)
	ListChannels(ctx context.Context, query types.Query) (*types.PageResult, error)

	// 消息发送
	SendMessage(ctx context.Context, channelID string, message *types.Message) (*SendResult, error)
	BroadcastMessage(ctx context.Context, channelIDs []string, message *types.Message) ([]*SendResult, error)

	// 渠道测试和验证
	TestChannel(ctx context.Context, id string) (*TestResult, error)
	ValidateConfig(ctx context.Context, channelType ChannelType, config ChannelConfig) error

	// 健康检查和监控
	HealthCheck(ctx context.Context, channelID string) (*HealthStatus, error)
	BatchHealthCheck(ctx context.Context, channelIDs []string) (map[string]*HealthStatus, error)
	StartHealthMonitor(ctx context.Context, interval time.Duration) error
	StopHealthMonitor() error

	// 统计和监控
	GetChannelStats(ctx context.Context, channelID string) (*ChannelStats, error)
	GetGlobalStats(ctx context.Context) (*GlobalStats, error)

	// 配置管理
	ReloadConfig(ctx context.Context) error
	GetConfig(ctx context.Context) (*ManagerConfig, error)
	UpdateConfig(ctx context.Context, config *ManagerConfig) error

	// 渠道查询
	GetActiveChannels(ctx context.Context) ([]*Channel, error)
}

// ChannelPlugin 渠道插件接口
type ChannelPlugin interface {
	// 插件信息
	GetType() ChannelType
	GetName() string
	GetVersion() string
	GetDescription() string
	GetConfigSchema() map[string]interface{}

	// 插件生命周期
	Initialize(ctx context.Context, config map[string]interface{}) error
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
	HealthCheck(ctx context.Context) error

	// 消息发送
	SendMessage(ctx context.Context, config ChannelConfig, message *types.Message) (*SendResult, error)

	// 配置验证
	ValidateConfig(config ChannelConfig) error

	// 测试连接
	TestConnection(ctx context.Context, config ChannelConfig) (*TestResult, error)

	// 支持的功能
	GetCapabilities() []PluginCapability
	SupportsFeature(feature PluginCapability) bool
}

// SendResult 发送结果
type SendResult struct {
	ChannelID   string                 `json:"channel_id"`
	MessageID   string                 `json:"message_id"`
	Success     bool                   `json:"success"`
	Error       string                 `json:"error,omitempty"`
	Latency     time.Duration          `json:"latency"`
	RetryCount  int                    `json:"retry_count"`
	Timestamp   time.Time              `json:"timestamp"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// HealthStatus 健康状态
type HealthStatus struct {
	ChannelID     string                 `json:"channel_id"`
	Status        HealthStatusType       `json:"status"`
	Message       string                 `json:"message"`
	LastCheck     time.Time              `json:"last_check"`
	ResponseTime  time.Duration          `json:"response_time"`
	ErrorCount    int                    `json:"error_count"`
	LastError     string                 `json:"last_error,omitempty"`
	LastErrorTime *time.Time             `json:"last_error_time,omitempty"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
}

// HealthStatusType 健康状态类型
type HealthStatusType string

const (
	HealthStatusHealthy   HealthStatusType = "healthy"
	HealthStatusUnhealthy HealthStatusType = "unhealthy"
	HealthStatusUnknown   HealthStatusType = "unknown"
	HealthStatusTesting   HealthStatusType = "testing"
)

// GlobalStats 全局统计信息
type GlobalStats struct {
	TotalChannels    int                        `json:"total_channels"`
	ActiveChannels   int                        `json:"active_channels"`
	HealthyChannels  int                        `json:"healthy_channels"`
	TotalSent        int64                      `json:"total_sent"`
	TotalFailed      int64                      `json:"total_failed"`
	SuccessRate      float64                    `json:"success_rate"`
	AvgLatency       float64                    `json:"avg_latency"`
	ChannelsByType   map[ChannelType]int        `json:"channels_by_type"`
	ChannelsByStatus map[ChannelStatus]int      `json:"channels_by_status"`
	DailyStats       *DailyStats                `json:"daily_stats"`
	LastUpdated      time.Time                  `json:"last_updated"`
}

// DailyStats 每日统计
type DailyStats struct {
	Date        string  `json:"date"`
	TotalSent   int64   `json:"total_sent"`
	TotalFailed int64   `json:"total_failed"`
	SuccessRate float64 `json:"success_rate"`
	AvgLatency  float64 `json:"avg_latency"`
}

// ManagerConfig 管理器配置
type ManagerConfig struct {
	HealthCheckInterval time.Duration          `json:"health_check_interval"`
	MaxRetries          int                    `json:"max_retries"`
	DefaultTimeout      time.Duration          `json:"default_timeout"`
	RateLimitEnabled    bool                   `json:"rate_limit_enabled"`
	MetricsEnabled      bool                   `json:"metrics_enabled"`
	PluginConfig        map[string]interface{} `json:"plugin_config"`
}

// PluginCapability 插件能力
type PluginCapability string

const (
	CapabilityTextMessage     PluginCapability = "text_message"
	CapabilityHTMLMessage     PluginCapability = "html_message"
	CapabilityMarkdownMessage PluginCapability = "markdown_message"
	CapabilityAttachments     PluginCapability = "attachments"
	CapabilityTemplating      PluginCapability = "templating"
	CapabilityRateLimit       PluginCapability = "rate_limit"
	CapabilityRetry           PluginCapability = "retry"
	CapabilityBatching        PluginCapability = "batching"
	CapabilityDeliveryStatus  PluginCapability = "delivery_status"
	CapabilityHealthCheck     PluginCapability = "health_check"
)

// PluginInfo 插件信息
type PluginInfo struct {
	Type         ChannelType        `json:"type"`
	Name         string             `json:"name"`
	Version      string             `json:"version"`
	Description  string             `json:"description"`
	Author       string             `json:"author"`
	Capabilities []PluginCapability `json:"capabilities"`
	ConfigSchema map[string]interface{} `json:"config_schema"`
	Status       PluginStatus       `json:"status"`
	LoadedAt     time.Time          `json:"loaded_at"`
}

// PluginStatus 插件状态
type PluginStatus string

const (
	PluginStatusLoaded    PluginStatus = "loaded"
	PluginStatusActive    PluginStatus = "active"
	PluginStatusInactive  PluginStatus = "inactive"
	PluginStatusError     PluginStatus = "error"
	PluginStatusUnloaded  PluginStatus = "unloaded"
)