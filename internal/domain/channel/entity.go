package channel

import (
	"fmt"
	"time"

	"alert_agent/pkg/types"
)

// Channel 通道实体
type Channel struct {
	ID          string            `json:"id" gorm:"primaryKey;type:varchar(36)"`
	Name        string            `json:"name" gorm:"type:varchar(255);not null;uniqueIndex"`
	Type        ChannelType       `json:"type" gorm:"type:varchar(50);not null"`
	Description string            `json:"description" gorm:"type:text"`
	Config      ChannelConfig     `json:"config" gorm:"type:json"`
	Status      ChannelStatus     `json:"status" gorm:"type:varchar(20);default:'active'"`
	Priority    int               `json:"priority" gorm:"default:0"`
	Tags        []string          `json:"tags" gorm:"type:json"`
	Labels      map[string]string `json:"labels" gorm:"type:json"`
	Metadata    types.Metadata    `json:"metadata" gorm:"embedded"`
	CreatedAt   time.Time         `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time         `json:"updated_at" gorm:"autoUpdateTime"`
}

// ChannelType 通道类型
type ChannelType string

const (
	ChannelTypeEmail     ChannelType = "email"
	ChannelTypeSMS       ChannelType = "sms"
	ChannelTypeWebhook   ChannelType = "webhook"
	ChannelTypeSlack     ChannelType = "slack"
	ChannelTypeDingTalk  ChannelType = "dingtalk"
	ChannelTypeWeChat    ChannelType = "wechat"
	ChannelTypeTelegram  ChannelType = "telegram"
	ChannelTypePagerDuty ChannelType = "pagerduty"
	ChannelTypeCustom    ChannelType = "custom"
)

// ChannelStatus 通道状态
type ChannelStatus string

const (
	ChannelStatusActive   ChannelStatus = "active"
	ChannelStatusInactive ChannelStatus = "inactive"
	ChannelStatusError    ChannelStatus = "error"
	ChannelStatusTesting  ChannelStatus = "testing"
)

// ChannelConfig 通道配置
type ChannelConfig struct {
	// 通用配置
	Enabled     bool                   `json:"enabled"`
	Timeout     time.Duration          `json:"timeout"`
	RetryConfig types.RetryConfig      `json:"retry_config"`
	RateLimit   RateLimitConfig        `json:"rate_limit"`
	Filters     []FilterConfig         `json:"filters"`
	Template    TemplateConfig         `json:"template"`
	Settings    map[string]interface{} `json:"settings"`
}

// RateLimitConfig 限流配置
type RateLimitConfig struct {
	Enabled  bool          `json:"enabled"`
	Rate     int           `json:"rate"`     // 每秒请求数
	Burst    int           `json:"burst"`    // 突发请求数
	Window   time.Duration `json:"window"`   // 时间窗口
	MaxDaily int           `json:"max_daily"` // 每日最大发送数
}

// FilterConfig 过滤器配置
type FilterConfig struct {
	Type      string                 `json:"type"`      // severity, label, time
	Condition string                 `json:"condition"` // eq, ne, gt, lt, in, not_in
	Value     interface{}            `json:"value"`
	Metadata  map[string]interface{} `json:"metadata"`
}

// TemplateConfig 模板配置
type TemplateConfig struct {
	Subject string            `json:"subject"`
	Body    string            `json:"body"`
	Format  string            `json:"format"` // text, html, markdown
	Vars    map[string]string `json:"vars"`
}

// Validate 验证通道配置
func (c *Channel) Validate() error {
	if c.Name == "" {
		return fmt.Errorf("name is required")
	}
	if c.Type == "" {
		return fmt.Errorf("type is required")
	}
	return nil
}

// IsActive 检查通道是否激活
func (c *Channel) IsActive() bool {
	return c.Status == ChannelStatusActive
}

// CanSend 检查是否可以发送
func (c *Channel) CanSend() bool {
	return c.IsActive() && c.Config.Enabled
}

// GetSetting 获取设置值
func (c *Channel) GetSetting(key string) interface{} {
	if c.Config.Settings == nil {
		return nil
	}
	return c.Config.Settings[key]
}

// SetSetting 设置配置值
func (c *Channel) SetSetting(key string, value interface{}) {
	if c.Config.Settings == nil {
		c.Config.Settings = make(map[string]interface{})
	}
	c.Config.Settings[key] = value
}

// AddTag 添加标签
func (c *Channel) AddTag(tag string) {
	for _, t := range c.Tags {
		if t == tag {
			return
		}
	}
	c.Tags = append(c.Tags, tag)
}

// RemoveTag 移除标签
func (c *Channel) RemoveTag(tag string) {
	for i, t := range c.Tags {
		if t == tag {
			c.Tags = append(c.Tags[:i], c.Tags[i+1:]...)
			return
		}
	}
}

// HasTag 检查是否有标签
func (c *Channel) HasTag(tag string) bool {
	for _, t := range c.Tags {
		if t == tag {
			return true
		}
	}
	return false
}

// SetLabel 设置标签
func (c *Channel) SetLabel(key, value string) {
	if c.Labels == nil {
		c.Labels = make(map[string]string)
	}
	c.Labels[key] = value
}

// GetLabel 获取标签值
func (c *Channel) GetLabel(key string) string {
	if c.Labels == nil {
		return ""
	}
	return c.Labels[key]
}

// HasLabel 检查是否有标签
func (c *Channel) HasLabel(key string) bool {
	if c.Labels == nil {
		return false
	}
	_, exists := c.Labels[key]
	return exists
}