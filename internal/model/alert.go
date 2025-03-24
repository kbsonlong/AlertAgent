package model

import (
	"errors"
	"time"

	"gorm.io/gorm"
)

// Alert 状态常量
const (
	AlertStatusNew          = "new"
	AlertStatusAcknowledged = "acknowledged"
	AlertStatusResolved     = "resolved"
)

// Alert 级别常量
const (
	AlertLevelCritical = "critical"
	AlertLevelHigh     = "high"
	AlertLevelMedium   = "medium"
	AlertLevelLow      = "low"
)

// Alert 告警记录
type Alert struct {
	ID          uint           `json:"id" gorm:"primarykey"`
	CreatedAt   time.Time      `json:"-"`
	UpdatedAt   time.Time      `json:"-"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`
	Title       string         `json:"title" gorm:"size:255;not null"`
	Level       string         `json:"level" gorm:"type:varchar(50);not null"`
	Status      string         `json:"status" gorm:"type:varchar(50);not null;default:'new'"`
	Source      string         `json:"source" gorm:"type:varchar(255);not null"`
	Content     string         `json:"content" gorm:"type:text;not null"`
	Labels      string         `json:"labels,omitempty" gorm:"type:text"`
	RuleID      uint           `json:"rule_id" gorm:"not null"`
	TemplateID  uint           `json:"template_id,omitempty"`
	GroupID     uint           `json:"group_id,omitempty"`
	Handler     string         `json:"handler,omitempty" gorm:"type:varchar(100)"`
	HandleTime  *time.Time     `json:"-"`
	HandleNote  string         `json:"handle_note,omitempty" gorm:"type:text"`
	Analysis    string         `json:"analysis,omitempty" gorm:"type:text"`
	NotifyTime  *time.Time     `json:"-"`
	NotifyCount int            `json:"notify_count,omitempty" gorm:"default:0"`
	Severity    string         `json:"severity" gorm:"type:varchar(20);not null;default:'medium'"`
}

// Validate 验证告警数据
func (a *Alert) Validate() error {
	if a.Title == "" {
		return errors.New("title is required")
	}
	if a.Content == "" {
		return errors.New("content is required")
	}
	if a.Source == "" {
		return errors.New("source is required")
	}
	if !isValidLevel(a.Level) {
		return errors.New("invalid alert level")
	}
	if !isValidStatus(a.Status) {
		return errors.New("invalid alert status")
	}
	return nil
}

// isValidLevel 验证告警级别
func isValidLevel(level string) bool {
	switch level {
	case AlertLevelCritical, AlertLevelHigh, AlertLevelMedium, AlertLevelLow:
		return true
	}
	return false
}

// isValidStatus 验证告警状态
func isValidStatus(status string) bool {
	switch status {
	case AlertStatusNew, AlertStatusAcknowledged, AlertStatusResolved:
		return true
	}
	return false
}

// AlertResponse 告警响应
type AlertResponse struct {
	ID          uint   `json:"id"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
	Title       string `json:"title"`
	Level       string `json:"level"`
	Status      string `json:"status"`
	Source      string `json:"source"`
	Content     string `json:"content"`
	Labels      string `json:"labels,omitempty"`
	RuleID      uint   `json:"rule_id"`
	TemplateID  uint   `json:"template_id,omitempty"`
	GroupID     uint   `json:"group_id,omitempty"`
	Handler     string `json:"handler,omitempty"`
	HandleTime  string `json:"handle_time,omitempty"`
	HandleNote  string `json:"handle_note,omitempty"`
	Analysis    string `json:"analysis,omitempty"`
	NotifyTime  string `json:"notify_time,omitempty"`
	NotifyCount int    `json:"notify_count,omitempty"`
	Severity    string `json:"severity"`
}

// ToResponse 转换为响应格式
func (a *Alert) ToResponse() *AlertResponse {
	resp := &AlertResponse{
		ID:          a.ID,
		CreatedAt:   a.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   a.UpdatedAt.Format(time.RFC3339),
		Title:       a.Title,
		Level:       a.Level,
		Status:      a.Status,
		Source:      a.Source,
		Content:     a.Content,
		Labels:      a.Labels,
		RuleID:      a.RuleID,
		TemplateID:  a.TemplateID,
		GroupID:     a.GroupID,
		Handler:     a.Handler,
		HandleNote:  a.HandleNote,
		Analysis:    a.Analysis,
		NotifyCount: a.NotifyCount,
		Severity:    a.Severity,
	}
	if a.HandleTime != nil {
		resp.HandleTime = a.HandleTime.Format(time.RFC3339)
	}
	if a.NotifyTime != nil {
		resp.NotifyTime = a.NotifyTime.Format(time.RFC3339)
	}
	return resp
}

// SimilarAlert 相似告警
type SimilarAlert struct {
	Alert      Alert   `json:"alert"`
	Similarity float64 `json:"similarity"`
}
