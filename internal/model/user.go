package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// User 用户模型
type User struct {
	ID           string         `json:"id" gorm:"type:varchar(36);primaryKey"`
	Username     string         `json:"username" gorm:"type:varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;not null;uniqueIndex"`
	Email        string         `json:"email" gorm:"type:varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;not null;uniqueIndex"`
	FullName     string         `json:"full_name" gorm:"type:varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;not null"`
	Password     string         `json:"-" gorm:"type:varchar(255);not null"`
	Phone        string         `json:"phone" gorm:"type:varchar(20) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci"`
	Department   string         `json:"department" gorm:"type:varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci"`
	Position     string         `json:"position" gorm:"type:varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci"`
	Role         string         `json:"role" gorm:"type:varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;not null;index"`
	Status       string         `json:"status" gorm:"type:varchar(20) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;not null;default:'active';index"`
	LastLoginAt  *time.Time     `json:"last_login_at" gorm:"index"`
	LastLoginIP  string         `json:"last_login_ip" gorm:"type:varchar(45)"`
	LoginCount   int            `json:"login_count" gorm:"default:0"`
	CreatedBy    string         `json:"created_by" gorm:"type:varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci"`
	UpdatedBy    string         `json:"updated_by" gorm:"type:varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `json:"-" gorm:"index"`

	// 关联关系
	Roles []Role `json:"roles,omitempty" gorm:"many2many:user_roles;"`
}

// 用户状态常量
const (
	UserStatusActive   = "active"
	UserStatusInactive = "inactive"
	UserStatusLocked   = "locked"
)

// 用户角色常量
const (
	UserRoleAdmin    = "admin"
	UserRoleOperator = "operator"
	UserRoleViewer   = "viewer"
)

// TableName 指定表名
func (User) TableName() string {
	return "users"
}

// BeforeCreate GORM钩子：创建前生成ID
func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.ID == "" {
		u.ID = uuid.New().String()
	}
	return nil
}

// UserStats 用户统计
type UserStats struct {
	Total    int64 `json:"total"`
	Active   int64 `json:"active"`
	Inactive int64 `json:"inactive"`
	Locked   int64 `json:"locked"`
}