package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Role 角色模型
type Role struct {
	ID          string         `json:"id" gorm:"type:varchar(36);primaryKey"`
	Name        string         `json:"name" gorm:"type:varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;not null;index"`
	Code        string         `json:"code" gorm:"type:varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;not null;uniqueIndex"`
	Description string         `json:"description" gorm:"type:text CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci"`
	Type        string         `json:"type" gorm:"type:varchar(20) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;not null;default:'custom';index"`
	Status      string         `json:"status" gorm:"type:varchar(20) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;not null;default:'active';index"`
	IsSystem    bool           `json:"is_system" gorm:"default:false;index"`
	CreatedBy   string         `json:"created_by" gorm:"type:varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci"`
	UpdatedBy   string         `json:"updated_by" gorm:"type:varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`

	// 关联关系
	Permissions []Permission `json:"permissions,omitempty" gorm:"many2many:role_permissions;"`
	Users       []User       `json:"users,omitempty" gorm:"many2many:user_roles;"`
}

// 角色类型常量
const (
	RoleTypeSystem = "system"
	RoleTypeCustom = "custom"
)

// 角色状态常量
const (
	RoleStatusActive   = "active"
	RoleStatusInactive = "inactive"
)

// 系统预定义角色常量
const (
	RoleCodeSuperAdmin = "super_admin"
	RoleCodeAdmin      = "admin"
	RoleCodeOperator   = "operator"
	RoleCodeViewer     = "viewer"
)

// TableName 指定表名
func (Role) TableName() string {
	return "roles"
}

// BeforeCreate GORM钩子：创建前生成ID
func (r *Role) BeforeCreate(tx *gorm.DB) error {
	if r.ID == "" {
		r.ID = uuid.New().String()
	}
	return nil
}

// RoleStats 角色统计
type RoleStats struct {
	Total  int64 `json:"total"`
	System int64 `json:"system"`
	Custom int64 `json:"custom"`
	Active int64 `json:"active"`
}

// IsSystemRole 判断是否为系统角色
func (r *Role) IsSystemRole() bool {
	return r.Type == RoleTypeSystem || r.IsSystem
}

// IsActive 判断角色是否激活
func (r *Role) IsActive() bool {
	return r.Status == RoleStatusActive
}

// HasPermission 检查角色是否拥有指定权限
func (r *Role) HasPermission(permissionCode string) bool {
	for _, permission := range r.Permissions {
		if permission.Code == permissionCode && permission.IsActive() {
			return true
		}
	}
	return false
}

// GetPermissionCodes 获取角色的所有权限代码
func (r *Role) GetPermissionCodes() []string {
	var codes []string
	for _, permission := range r.Permissions {
		codes = append(codes, permission.Code)
	}
	return codes
}