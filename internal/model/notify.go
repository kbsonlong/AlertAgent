package model

import (
	"gorm.io/gorm"
)

// NotifyTemplate 通知模板
type NotifyTemplate struct {
	gorm.Model
	Name        string `gorm:"size:100;not null;uniqueIndex" json:"name"`
	Type        string `gorm:"size:50;not null" json:"type"`
	Content     string `gorm:"type:text;not null" json:"content"`
	Description string `gorm:"type:text" json:"description"`
	Enabled     bool   `gorm:"default:true" json:"enabled"`
}

// NotifyGroup 通知组
type NotifyGroup struct {
	gorm.Model
	Name        string `gorm:"size:100;not null;uniqueIndex" json:"name"`
	Description string `gorm:"type:text" json:"description"`
	Members     string `gorm:"type:text" json:"members"`
	Enabled     bool   `gorm:"default:true" json:"enabled"`
}

// NotifyRecord 通知记录
type NotifyRecord struct {
	gorm.Model
	AlertID    uint   `gorm:"index" json:"alert_id"`
	Type       string `gorm:"size:50;not null" json:"type"`
	Target     string `gorm:"size:255;not null" json:"target"`
	Content    string `gorm:"type:text" json:"content"`
	Status     string `gorm:"size:20;not null" json:"status"`
	Response   string `gorm:"type:text" json:"response"`
	RetryCount int    `gorm:"default:0" json:"retry_count"`
}
