package domain

import (
	"gorm.io/gorm"
)

// Permission 权限实体
type Permission struct {
	gorm.Model
	Name        string `gorm:"uniqueIndex;not null;size:100" json:"name"`
	Description string `gorm:"size:255" json:"description"`
	Resource    string `gorm:"not null;size:50" json:"resource"`
	Action      string `gorm:"not null;size:50" json:"action"`
	IsActive    bool   `gorm:"default:true" json:"is_active"`
}

// TableName 指定表名
func (Permission) TableName() string {
	return "permissions"
}

// GetFullName 获取权限的完整名称
func (p *Permission) GetFullName() string {
	return p.Resource + ":" + p.Action
}