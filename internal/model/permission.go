package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Permission 权限模型
type Permission struct {
	ID          string         `json:"id" gorm:"type:varchar(36);primaryKey"`
	Name        string         `json:"name" gorm:"type:varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;not null;index"`
	Code        string         `json:"code" gorm:"type:varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;not null;uniqueIndex"`
	Description string         `json:"description" gorm:"type:text CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci"`
	Resource    string         `json:"resource" gorm:"type:varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;not null;index"`
	Action      string         `json:"action" gorm:"type:varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;not null;index"`
	Category    string         `json:"category" gorm:"type:varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;not null;index"`
	Type        string         `json:"type" gorm:"type:varchar(20) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;not null;default:'custom';index"`
	Status      string         `json:"status" gorm:"type:varchar(20) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;not null;default:'active';index"`
	IsSystem    bool           `json:"is_system" gorm:"default:false;index"`
	CreatedBy   string         `json:"created_by" gorm:"type:varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci"`
	UpdatedBy   string         `json:"updated_by" gorm:"type:varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`

	// 关联关系
	Roles []Role `json:"roles,omitempty" gorm:"many2many:role_permissions;"`
}

// 权限类型常量
const (
	PermissionTypeSystem = "system"
	PermissionTypeCustom = "custom"
)

// 权限状态常量
const (
	PermissionStatusActive   = "active"
	PermissionStatusInactive = "inactive"
)

// 权限操作常量
const (
	PermissionActionCreate = "create"
	PermissionActionRead   = "read"
	PermissionActionUpdate = "update"
	PermissionActionDelete = "delete"
	PermissionActionList   = "list"
	PermissionActionExport = "export"
	PermissionActionImport = "import"
	PermissionActionManage = "manage"
)

// 权限分类常量
const (
	PermissionCategoryUser       = "user"
	PermissionCategoryRole       = "role"
	PermissionCategoryPermission = "permission"
	PermissionCategoryAlert      = "alert"
	PermissionCategoryRule       = "rule"
	PermissionCategoryProvider   = "provider"
	PermissionCategoryConfig     = "config"
	PermissionCategorySystem     = "system"
)

// TableName 指定表名
func (Permission) TableName() string {
	return "permissions"
}

// BeforeCreate GORM钩子：创建前生成ID
func (p *Permission) BeforeCreate(tx *gorm.DB) error {
	if p.ID == "" {
		p.ID = uuid.New().String()
	}
	return nil
}

// PermissionStats 权限统计
type PermissionStats struct {
	Total  int64 `json:"total"`
	System int64 `json:"system"`
	Custom int64 `json:"custom"`
	Active int64 `json:"active"`
}

// GetPermissionKey 获取权限唯一标识
func (p *Permission) GetPermissionKey() string {
	return p.Resource + ":" + p.Action
}

// IsSystemPermission 判断是否为系统权限
func (p *Permission) IsSystemPermission() bool {
	return p.Type == PermissionTypeSystem || p.IsSystem
}

// IsActive 检查权限是否激活
func (p *Permission) IsActive() bool {
	return p.Status == PermissionStatusActive
}