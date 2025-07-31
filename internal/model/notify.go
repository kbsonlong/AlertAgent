package model

import (
	"encoding/json"
	"errors"
	"strings"

	"gorm.io/gorm"
)

// 通知类型常量
const (
	NotifyTypeEmail   = "email"
	NotifyTypeSMS     = "sms"
	NotifyTypeWebhook = "webhook"
)

// 通知状态常量
const (
	NotifyStatusPending  = "pending"
	NotifyStatusSent     = "sent"
	NotifyStatusFailed   = "failed"
	NotifyStatusRetrying = "retrying"
)

// NotifyTemplate 通知模板
type NotifyTemplate struct {
	gorm.Model
	Name        string `gorm:"size:100;not null;uniqueIndex" json:"name"`
	Type        string `gorm:"size:50;not null" json:"type"`
	Content     string `gorm:"type:text;not null" json:"content"`
	Description string `gorm:"type:text" json:"description,omitempty"`
	Variables   string `gorm:"type:text" json:"variables,omitempty"`
	Enabled     bool   `gorm:"default:true" json:"enabled"`
}

// Validate 验证通知模板
func (t *NotifyTemplate) Validate() error {
	if t.Name == "" {
		return errors.New("template name is required")
	}
	if t.Content == "" {
		return errors.New("template content is required")
	}
	if !isValidNotifyType(t.Type) {
		return errors.New("invalid notify type")
	}
	return nil
}

// GetVariables 获取模板变量列表
func (t *NotifyTemplate) GetVariables() []string {
	if t.Variables == "" {
		return nil
	}
	var vars []string
	_ = json.Unmarshal([]byte(t.Variables), &vars)
	return vars
}

// SetVariables 设置模板变量列表
func (t *NotifyTemplate) SetVariables(vars []string) {
	if len(vars) == 0 {
		t.Variables = ""
		return
	}
	data, _ := json.Marshal(vars)
	t.Variables = string(data)
}

// NotifyGroup 通知组
type NotifyGroup struct {
	gorm.Model
	Name        string `gorm:"size:100;not null;uniqueIndex" json:"name"`
	Description string `gorm:"type:text;not null" json:"description"`
	Contacts    string `gorm:"type:json;not null" json:"contacts"`
	Members     string `gorm:"type:text" json:"members"`
	Channels    string `gorm:"type:text" json:"channels"`
	Enabled     bool   `gorm:"default:true" json:"enabled"`
}

// Validate 验证通知组
func (g *NotifyGroup) Validate() error {
	if g.Name == "" {
		return errors.New("group name is required")
	}
	if g.Members == "" {
		return errors.New("group members is required")
	}
	return nil
}

// GetMembers 获取组成员列表
func (g *NotifyGroup) GetMembers() []string {
	if g.Members == "" {
		return nil
	}
	return strings.Split(g.Members, ",")
}

// SetMembers 设置组成员列表
func (g *NotifyGroup) SetMembers(members []string) {
	g.Members = strings.Join(members, ",")
}

// GetChannels 获取通知渠道列表
func (g *NotifyGroup) GetChannels() []string {
	if g.Channels == "" {
		return nil
	}
	return strings.Split(g.Channels, ",")
}

// SetChannels 设置通知渠道列表
func (g *NotifyGroup) SetChannels(channels []string) {
	g.Channels = strings.Join(channels, ",")
}

// NotifyRecord 通知记录
type NotifyRecord struct {
	gorm.Model
	AlertID    uint   `gorm:"index" json:"alert_id"`
	Type       string `gorm:"size:50;not null" json:"type"`
	Target     string `gorm:"size:255;not null" json:"target"`
	Content    string `gorm:"type:text" json:"content"`
	Status     string `gorm:"size:20;not null;default:'pending'" json:"status"`
	Response   string `gorm:"type:text" json:"response,omitempty"`
	RetryCount int    `gorm:"default:0" json:"retry_count"`
	Error      string `gorm:"type:text" json:"error,omitempty"`
}

// Validate 验证通知记录
func (r *NotifyRecord) Validate() error {
	if r.AlertID == 0 {
		return errors.New("alert id is required")
	}
	if !isValidNotifyType(r.Type) {
		return errors.New("invalid notify type")
	}
	if r.Target == "" {
		return errors.New("notify target is required")
	}
	if r.Content == "" {
		return errors.New("notify content is required")
	}
	if !isValidNotifyStatus(r.Status) {
		return errors.New("invalid notify status")
	}
	return nil
}

// isValidNotifyType 验证通知类型
func isValidNotifyType(typ string) bool {
	switch typ {
	case NotifyTypeEmail, NotifyTypeSMS, NotifyTypeWebhook:
		return true
	}
	return false
}

// NotificationPlugin 通知插件配置
type NotificationPlugin struct {
	gorm.Model
	Name        string `gorm:"size:100;not null;uniqueIndex" json:"name"`
	DisplayName string `gorm:"size:200;not null" json:"display_name"`
	Version     string `gorm:"size:50;not null" json:"version"`
	Config      string `gorm:"type:json;not null" json:"config"` // 插件配置JSON
	Enabled     bool   `gorm:"default:false" json:"enabled"`
	Priority    int    `gorm:"default:0" json:"priority"` // 发送优先级
}

// Validate 验证通知插件配置
func (p *NotificationPlugin) Validate() error {
	if p.Name == "" {
		return errors.New("plugin name is required")
	}
	if p.DisplayName == "" {
		return errors.New("plugin display name is required")
	}
	if p.Version == "" {
		return errors.New("plugin version is required")
	}
	if p.Config == "" {
		return errors.New("plugin config is required")
	}
	return nil
}

// GetConfig 获取插件配置
func (p *NotificationPlugin) GetConfig() (map[string]interface{}, error) {
	var config map[string]interface{}
	if err := json.Unmarshal([]byte(p.Config), &config); err != nil {
		return nil, err
	}
	return config, nil
}

// SetConfig 设置插件配置
func (p *NotificationPlugin) SetConfig(config map[string]interface{}) error {
	data, err := json.Marshal(config)
	if err != nil {
		return err
	}
	p.Config = string(data)
	return nil
}

// isValidNotifyStatus 验证通知状态
func isValidNotifyStatus(status string) bool {
	switch status {
	case NotifyStatusPending, NotifyStatusSent, NotifyStatusFailed, NotifyStatusRetrying:
		return true
	}
	return false
}
