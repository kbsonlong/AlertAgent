package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// RuleAuditLog 规则审计日志
type RuleAuditLog struct {
	ID        string         `json:"id" gorm:"type:varchar(36);primaryKey"`
	RuleID    string         `json:"rule_id" gorm:"type:varchar(36);not null;index"`
	Action    string         `json:"action" gorm:"type:varchar(50);not null;index"` // create, update, delete, activate, deactivate
	UserID    string         `json:"user_id" gorm:"type:varchar(36);index"`
	UserName  string         `json:"user_name" gorm:"type:varchar(100)"`
	Changes   string         `json:"changes" gorm:"type:text"` // JSON格式的变更详情
	Reason    string         `json:"reason" gorm:"type:text"`  // 变更原因
	IPAddress string         `json:"ip_address" gorm:"type:varchar(45)"`
	UserAgent string         `json:"user_agent" gorm:"type:text"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

// TableName 指定表名
func (RuleAuditLog) TableName() string {
	return "rule_audit_logs"
}

// BeforeCreate GORM钩子：创建前生成ID
func (r *RuleAuditLog) BeforeCreate(tx *gorm.DB) error {
	if r.ID == "" {
		r.ID = uuid.New().String()
	}
	return nil
}