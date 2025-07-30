package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ConfigSyncStatus 配置同步状态
type ConfigSyncStatus struct {
	ID           string         `json:"id" gorm:"type:varchar(36);primaryKey"`
	ClusterID    string         `json:"cluster_id" gorm:"type:varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;not null;index"`
	ConfigType   string         `json:"config_type" gorm:"type:varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;not null;index"`
	ConfigHash   string         `json:"config_hash" gorm:"type:varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci"`
	SyncStatus   string         `json:"sync_status" gorm:"type:varchar(20) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;not null;default:'pending';index"`
	SyncTime     *time.Time     `json:"sync_time"`
	ErrorMessage string         `json:"error_message" gorm:"type:text CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `json:"-" gorm:"index"`
}

// TableName 指定表名
func (ConfigSyncStatus) TableName() string {
	return "config_sync_status"
}

// BeforeCreate GORM钩子：创建前生成ID
func (c *ConfigSyncStatus) BeforeCreate(tx *gorm.DB) error {
	if c.ID == "" {
		c.ID = generateID()
	}
	return nil
}

// ConfigSyncTrigger 配置同步触发记录
type ConfigSyncTrigger struct {
	ID         string         `json:"id" gorm:"type:varchar(36);primaryKey"`
	ClusterID  string         `json:"cluster_id" gorm:"type:varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;not null;index"`
	ConfigType string         `json:"config_type" gorm:"type:varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;not null;index"`
	TriggerBy  string         `json:"trigger_by" gorm:"type:varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;not null"`
	Reason     string         `json:"reason" gorm:"type:varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci"`
	Status     string         `json:"status" gorm:"type:varchar(20) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;not null;default:'pending';index"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `json:"-" gorm:"index"`
}

// TableName 指定表名
func (ConfigSyncTrigger) TableName() string {
	return "config_sync_triggers"
}

// BeforeCreate GORM钩子：创建前生成ID
func (c *ConfigSyncTrigger) BeforeCreate(tx *gorm.DB) error {
	if c.ID == "" {
		c.ID = generateID()
	}
	return nil
}

// ConfigSyncHistory 配置同步历史
type ConfigSyncHistory struct {
	ID           string         `json:"id" gorm:"type:varchar(36);primaryKey"`
	ClusterID    string         `json:"cluster_id" gorm:"type:varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;not null;index"`
	ConfigType   string         `json:"config_type" gorm:"type:varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;not null;index"`
	ConfigHash   string         `json:"config_hash" gorm:"type:varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;not null"`
	ConfigSize   int64          `json:"config_size" gorm:"not null;default:0"`
	SyncStatus   string         `json:"sync_status" gorm:"type:varchar(20) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;not null;index"`
	SyncDuration int64          `json:"sync_duration" gorm:"not null;default:0"` // 同步耗时，毫秒
	ErrorMessage string         `json:"error_message" gorm:"type:text CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci"`
	CreatedAt    time.Time      `json:"created_at"`
	DeletedAt    gorm.DeletedAt `json:"-" gorm:"index"`
}

// TableName 指定表名
func (ConfigSyncHistory) TableName() string {
	return "config_sync_history"
}

// BeforeCreate GORM钩子：创建前生成ID
func (c *ConfigSyncHistory) BeforeCreate(tx *gorm.DB) error {
	if c.ID == "" {
		c.ID = generateID()
	}
	return nil
}

// SyncStatusSummary 同步状态汇总
type SyncStatusSummary struct {
	ClusterID     string            `json:"cluster_id"`
	ConfigTypes   []string          `json:"config_types"`
	LastSyncTime  map[string]time.Time `json:"last_sync_time"`
	SyncStatus    map[string]string `json:"sync_status"`
	ErrorMessages map[string]string `json:"error_messages"`
	ConfigHashes  map[string]string `json:"config_hashes"`
}

// GetSyncDelay 获取同步延迟（秒）
func (c *ConfigSyncStatus) GetSyncDelay() int64 {
	if c.SyncTime == nil {
		return -1 // 从未同步
	}
	return time.Since(*c.SyncTime).Milliseconds() / 1000
}

// IsHealthy 检查同步状态是否健康
func (c *ConfigSyncStatus) IsHealthy(maxDelaySeconds int64) bool {
	if c.SyncStatus != "success" {
		return false
	}
	
	delay := c.GetSyncDelay()
	if delay < 0 || delay > maxDelaySeconds {
		return false
	}
	
	return true
}

// generateID 生成UUID
func generateID() string {
	return uuid.New().String()
}