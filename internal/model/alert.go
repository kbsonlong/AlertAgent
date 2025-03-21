package model

import (
	"time"

	"gorm.io/gorm"
)

// Alert 告警记录
type Alert struct {
	ID          uint           `json:"id" gorm:"primarykey"`
	CreatedAt   time.Time      `json:"-"`
	UpdatedAt   time.Time      `json:"-"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`
	Name        string         `json:"name" gorm:"type:varchar(255);not null"`
	Level       string         `json:"level" gorm:"type:varchar(50);not null"`
	Status      string         `json:"status" gorm:"type:varchar(50);not null;default:'active'"`
	Source      string         `json:"source" gorm:"type:varchar(255);not null"`
	Content     string         `json:"content" gorm:"type:text;not null"`
	Handler     string         `json:"handler" gorm:"type:varchar(100)"`
	Note        string         `json:"note" gorm:"type:text"`
	Analysis    string         `json:"analysis" gorm:"type:text"`
	RuleID      uint           `json:"rule_id" gorm:"not null"`
	TemplateID  uint           `json:"template_id"`
	GroupID     uint           `json:"group_id"`
	Title       string         `gorm:"size:255;not null" json:"title"`
	Labels      string         `gorm:"type:text" json:"labels"`
	HandleTime  *time.Time     `json:"-"`
	HandleNote  string         `gorm:"type:text" json:"handle_note"`
	NotifyTime  *time.Time     `json:"-"`
	NotifyCount int            `gorm:"default:0" json:"notify_count"`
}

// AlertResponse 告警响应
type AlertResponse struct {
	ID          uint   `json:"id"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
	Name        string `json:"name"`
	Level       string `json:"level"`
	Status      string `json:"status"`
	Source      string `json:"source"`
	Content     string `json:"content"`
	Handler     string `json:"handler,omitempty"`
	Note        string `json:"note,omitempty"`
	Analysis    string `json:"analysis,omitempty"`
	RuleID      uint   `json:"rule_id"`
	TemplateID  uint   `json:"template_id,omitempty"`
	GroupID     uint   `json:"group_id,omitempty"`
	Title       string `json:"title"`
	Labels      string `json:"labels,omitempty"`
	HandleTime  string `json:"handle_time,omitempty"`
	HandleNote  string `json:"handle_note,omitempty"`
	NotifyTime  string `json:"notify_time,omitempty"`
	NotifyCount int    `json:"notify_count,omitempty"`
}

// ToResponse 转换为响应格式
func (a *Alert) ToResponse() *AlertResponse {
	resp := &AlertResponse{
		ID:          a.ID,
		CreatedAt:   a.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:   a.UpdatedAt.Format("2006-01-02 15:04:05"),
		Name:        a.Name,
		Level:       a.Level,
		Status:      a.Status,
		Source:      a.Source,
		Content:     a.Content,
		Handler:     a.Handler,
		Note:        a.Note,
		Analysis:    a.Analysis,
		RuleID:      a.RuleID,
		TemplateID:  a.TemplateID,
		GroupID:     a.GroupID,
		Title:       a.Title,
		Labels:      a.Labels,
		HandleNote:  a.HandleNote,
		NotifyCount: a.NotifyCount,
	}
	if a.HandleTime != nil {
		resp.HandleTime = a.HandleTime.Format("2006-01-02 15:04:05")
	}
	if a.NotifyTime != nil {
		resp.NotifyTime = a.NotifyTime.Format("2006-01-02 15:04:05")
	}
	return resp
}

// SimilarAlert 相似告警
type SimilarAlert struct {
	Alert      Alert   `json:"alert"`
	Similarity float64 `json:"similarity"`
}
