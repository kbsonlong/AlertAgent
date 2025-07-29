package channel

import (
	"context"
	"time"
)

// Channel 渠道领域模型
type Channel struct {
	ID              string                 `json:"id"`
	Name            string                 `json:"name"`
	Type            string                 `json:"type"`
	Description     string                 `json:"description"`
	Config          map[string]interface{} `json:"config"`
	GroupID         string                 `json:"group_id"`
	Tags            []string               `json:"tags"`
	Status          ChannelStatus          `json:"status"`
	HealthStatus    HealthStatus           `json:"health_status"`
	LastHealthCheck *time.Time             `json:"last_health_check"`
	ResponseTime    int                    `json:"response_time"`
	SuccessRate     float64                `json:"success_rate"`
	CreatedBy       string                 `json:"created_by"`
	UpdatedBy       string                 `json:"updated_by"`
	CreatedAt       time.Time              `json:"created_at"`
	UpdatedAt       time.Time              `json:"updated_at"`
}

// ChannelGroup 渠道分组
type ChannelGroup struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	ParentID    string    `json:"parent_id"`
	Path        string    `json:"path"`
	Level       int       `json:"level"`
	SortOrder   int       `json:"sort_order"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Message 消息
type Message struct {
	Title       string                 `json:"title"`
	Content     string                 `json:"content"`
	Severity    string                 `json:"severity"`
	Labels      map[string]string      `json:"labels"`
	Annotations map[string]string      `json:"annotations"`
	Metadata    map[string]interface{} `json:"metadata"`
	Timestamp   time.Time              `json:"timestamp"`
}

// TestResult 测试结果
type TestResult struct {
	Success      bool      `json:"success"`
	ResponseTime int       `json:"response_time"` // 毫秒
	Message      string    `json:"message"`
	Error        string    `json:"error,omitempty"`
	Timestamp    time.Time `json:"timestamp"`
}

// ChannelStatus 渠道状态
type ChannelStatus string

const (
	ChannelStatusActive   ChannelStatus = "active"
	ChannelStatusInactive ChannelStatus = "inactive"
	ChannelStatusDisabled ChannelStatus = "disabled"
)

// HealthStatus 健康状态
type HealthStatus string

const (
	HealthStatusHealthy   HealthStatus = "healthy"
	HealthStatusUnhealthy HealthStatus = "unhealthy"
	HealthStatusUnknown   HealthStatus = "unknown"
)

// CreateChannelRequest 创建渠道请求
type CreateChannelRequest struct {
	Name        string                 `json:"name" binding:"required"`
	Type        string                 `json:"type" binding:"required"`
	Description string                 `json:"description"`
	Config      map[string]interface{} `json:"config" binding:"required"`
	GroupID     string                 `json:"group_id"`
	Tags        []string               `json:"tags"`
}

// UpdateChannelRequest 更新渠道请求
type UpdateChannelRequest struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Config      map[string]interface{} `json:"config"`
	GroupID     string                 `json:"group_id"`
	Tags        []string               `json:"tags"`
	Status      ChannelStatus          `json:"status"`
}

// ChannelQuery 渠道查询
type ChannelQuery struct {
	Name     string        `json:"name"`
	Type     string        `json:"type"`
	GroupID  string        `json:"group_id"`
	Status   ChannelStatus `json:"status"`
	Tags     []string      `json:"tags"`
	Page     int           `json:"page"`
	PageSize int           `json:"page_size"`
	SortBy   string        `json:"sort_by"`
	SortDesc bool          `json:"sort_desc"`
}

// ChannelRepository 渠道仓储接口
type ChannelRepository interface {
	// 基本CRUD操作
	Create(ctx context.Context, channel *Channel) error
	Update(ctx context.Context, channel *Channel) error
	Delete(ctx context.Context, id string) error
	GetByID(ctx context.Context, id string) (*Channel, error)
	List(ctx context.Context, query *ChannelQuery) ([]*Channel, int64, error)

	// 查询操作
	GetByType(ctx context.Context, channelType string) ([]*Channel, error)
	GetByGroupID(ctx context.Context, groupID string) ([]*Channel, error)
	GetByTags(ctx context.Context, tags []string) ([]*Channel, error)

	// 健康状态操作
	UpdateHealthStatus(ctx context.Context, id string, status HealthStatus, responseTime int) error
	GetUnhealthyChannels(ctx context.Context) ([]*Channel, error)

	// 分组操作
	CreateGroup(ctx context.Context, group *ChannelGroup) error
	UpdateGroup(ctx context.Context, group *ChannelGroup) error
	DeleteGroup(ctx context.Context, id string) error
	GetGroupByID(ctx context.Context, id string) (*ChannelGroup, error)
	ListGroups(ctx context.Context) ([]*ChannelGroup, error)
}

// ChannelPlugin 渠道插件接口
type ChannelPlugin interface {
	GetType() string
	GetName() string
	GetConfigSchema() *ConfigSchema
	ValidateConfig(config map[string]interface{}) error
	TestConnection(config map[string]interface{}) (*TestResult, error)
	SendMessage(config map[string]interface{}, message *Message) error
	GetHealthStatus(config map[string]interface{}) (*HealthStatus, error)
}

// ConfigSchema 配置模式
type ConfigSchema struct {
	Fields []ConfigField `json:"fields"`
}

// ConfigField 配置字段
type ConfigField struct {
	Name        string           `json:"name"`
	Type        string           `json:"type"`
	Label       string           `json:"label"`
	Description string           `json:"description"`
	Required    bool             `json:"required"`
	Default     interface{}      `json:"default,omitempty"`
	Options     []ConfigOption   `json:"options,omitempty"`
	Validation  *ValidationRule  `json:"validation,omitempty"`
}

// ConfigOption 配置选项
type ConfigOption struct {
	Label string      `json:"label"`
	Value interface{} `json:"value"`
}

// ValidationRule 验证规则
type ValidationRule struct {
	Pattern string `json:"pattern,omitempty"`
	Min     int    `json:"min,omitempty"`
	Max     int    `json:"max,omitempty"`
	Message string `json:"message,omitempty"`
}

// PluginInfo 插件信息
type PluginInfo struct {
	Type        string        `json:"type"`
	Name        string        `json:"name"`
	Description string        `json:"description"`
	Version     string        `json:"version"`
	Schema      *ConfigSchema `json:"schema"`
}