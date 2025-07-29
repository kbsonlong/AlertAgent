package domain

import (
	"time"

	"gorm.io/gorm"
)

// AuditLog 审计日志实体
type AuditLog struct {
	gorm.Model
	UserID      uint      `gorm:"index" json:"user_id"`
	User        *User     `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Action      string    `gorm:"not null;size:100" json:"action"`
	Resource    string    `gorm:"not null;size:100" json:"resource"`
	ResourceID  string    `gorm:"size:50" json:"resource_id"`
	IPAddress   string    `gorm:"size:45" json:"ip_address"`
	UserAgent   string    `gorm:"size:500" json:"user_agent"`
	Details     string    `gorm:"type:text" json:"details"`
	Status      string    `gorm:"size:20;default:'success'" json:"status"`
	Timestamp   time.Time `gorm:"index" json:"timestamp"`
}

// TableName 指定表名
func (AuditLog) TableName() string {
	return "audit_logs"
}

// BeforeCreate GORM钩子，在创建前设置时间戳
func (a *AuditLog) BeforeCreate(tx *gorm.DB) error {
	a.Timestamp = time.Now()
	return nil
}