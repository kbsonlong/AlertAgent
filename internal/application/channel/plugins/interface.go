package plugins

import (
	"context"
	"time"
)

// NotificationPlugin 定义通知插件接口
type NotificationPlugin interface {
	// 插件基本信息
	Name() string
	Version() string
	Description() string
	
	// 配置相关
	ConfigSchema() map[string]interface{} // JSON Schema
	ValidateConfig(config map[string]interface{}) error
	
	// 通知发送
	Send(ctx context.Context, config map[string]interface{}, message *NotificationMessage) error
	
	// 健康检查
	HealthCheck(ctx context.Context, config map[string]interface{}) error
	
	// 插件生命周期
	Initialize() error
	Shutdown() error
}

// NotificationMessage 通知消息结构
type NotificationMessage struct {
	Title       string                 `json:"title"`
	Content     string                 `json:"content"`
	Severity    string                 `json:"severity"`
	AlertID     string                 `json:"alert_id"`
	Timestamp   time.Time              `json:"timestamp"`
	Labels      map[string]string      `json:"labels"`
	Annotations map[string]string      `json:"annotations"`
	Extra       map[string]interface{} `json:"extra"`
}

// PluginInfo 插件信息
type PluginInfo struct {
	Name        string                 `json:"name"`
	Version     string                 `json:"version"`
	Description string                 `json:"description"`
	Schema      map[string]interface{} `json:"schema"`
	Status      string                 `json:"status"` // active, inactive, error
	LoadTime    time.Time              `json:"load_time"`
	LastError   string                 `json:"last_error,omitempty"`
}

// PluginConfig 插件配置
type PluginConfig struct {
	Name     string                 `json:"name"`
	Enabled  bool                   `json:"enabled"`
	Config   map[string]interface{} `json:"config"`
	Priority int                    `json:"priority"` // 发送优先级，数字越小优先级越高
}

// SendResult 发送结果
type SendResult struct {
	Success   bool      `json:"success"`
	Error     string    `json:"error,omitempty"`
	Duration  time.Duration `json:"duration"`
	Timestamp time.Time `json:"timestamp"`
}

// PluginStats 插件统计信息
type PluginStats struct {
	Name         string    `json:"name"`
	TotalSent    int64     `json:"total_sent"`
	SuccessCount int64     `json:"success_count"`
	FailureCount int64     `json:"failure_count"`
	AvgDuration  time.Duration `json:"avg_duration"`
	LastSent     time.Time `json:"last_sent"`
	LastError    string    `json:"last_error,omitempty"`
}