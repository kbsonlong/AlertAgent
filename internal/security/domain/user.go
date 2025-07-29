package domain

import (
	"time"

	"gorm.io/gorm"
)

// User 用户实体
type User struct {
	gorm.Model
	Username    string    `gorm:"uniqueIndex;not null;size:50" json:"username"`
	Email       string    `gorm:"uniqueIndex;not null;size:100" json:"email"`
	PasswordHash string   `gorm:"not null;size:255" json:"-"`
	FullName    string    `gorm:"size:100" json:"full_name"`
	IsActive    bool      `gorm:"default:true" json:"is_active"`
	LastLoginAt *time.Time `json:"last_login_at"`
	Roles       []Role    `gorm:"many2many:user_roles;" json:"roles"`
}

// TableName 指定表名
func (User) TableName() string {
	return "users"
}

// HasRole 检查用户是否具有指定角色
func (u *User) HasRole(roleName string) bool {
	for _, role := range u.Roles {
		if role.Name == roleName {
			return true
		}
	}
	return false
}

// HasPermission 检查用户是否具有指定权限
func (u *User) HasPermission(permissionName string) bool {
	for _, role := range u.Roles {
		for _, permission := range role.Permissions {
			if permission.Name == permissionName {
				return true
			}
		}
	}
	return false
}