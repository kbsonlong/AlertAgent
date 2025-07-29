package domain

import (
	"gorm.io/gorm"
)

// Role 角色实体
type Role struct {
	gorm.Model
	Name        string       `gorm:"uniqueIndex;not null;size:50" json:"name"`
	Description string       `gorm:"size:255" json:"description"`
	IsActive    bool         `gorm:"default:true" json:"is_active"`
	Permissions []Permission `gorm:"many2many:role_permissions;" json:"permissions"`
}

// TableName 指定表名
func (Role) TableName() string {
	return "roles"
}

// HasPermission 检查角色是否具有指定权限
func (r *Role) HasPermission(permissionName string) bool {
	for _, permission := range r.Permissions {
		if permission.Name == permissionName {
			return true
		}
	}
	return false
}