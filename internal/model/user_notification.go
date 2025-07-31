package model

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// UserNotificationConfig 用户通知配置
type UserNotificationConfig struct {
	ID           string         `json:"id" gorm:"type:varchar(36);primaryKey"`
	UserID       string         `json:"user_id" gorm:"type:varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;not null;index"`
	PluginName   string         `json:"plugin_name" gorm:"type:varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;not null;index"`
	Config       string         `json:"config" gorm:"type:json;not null"`
	Enabled      bool           `json:"enabled" gorm:"default:true;index"`
	AlertLevels  string         `json:"alert_levels" gorm:"type:json"`
	TimeWindows  string         `json:"time_windows" gorm:"type:json"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `json:"-" gorm:"index"`
}

// TableName 指定表名
func (UserNotificationConfig) TableName() string {
	return "user_notification_configs"
}

// BeforeCreate GORM钩子：创建前生成ID
func (u *UserNotificationConfig) BeforeCreate(tx *gorm.DB) error {
	if u.ID == "" {
		u.ID = uuid.New().String()
	}
	return nil
}

// GetConfig 获取配置映射
func (u *UserNotificationConfig) GetConfig() (map[string]interface{}, error) {
	var config map[string]interface{}
	err := json.Unmarshal([]byte(u.Config), &config)
	return config, err
}

// SetConfig 设置配置映射
func (u *UserNotificationConfig) SetConfig(config map[string]interface{}) error {
	data, err := json.Marshal(config)
	if err != nil {
		return err
	}
	u.Config = string(data)
	return nil
}

// GetAlertLevels 获取告警级别过滤
func (u *UserNotificationConfig) GetAlertLevels() ([]string, error) {
	if u.AlertLevels == "" {
		return []string{"critical", "high", "medium", "low"}, nil // 默认所有级别
	}
	var levels []string
	err := json.Unmarshal([]byte(u.AlertLevels), &levels)
	return levels, err
}

// SetAlertLevels 设置告警级别过滤
func (u *UserNotificationConfig) SetAlertLevels(levels []string) error {
	if levels == nil {
		u.AlertLevels = ""
		return nil
	}
	data, err := json.Marshal(levels)
	if err != nil {
		return err
	}
	u.AlertLevels = string(data)
	return nil
}

// TimeWindow 时间窗口配置
type TimeWindow struct {
	Start   string   `json:"start"`   // 格式: "09:00"
	End     string   `json:"end"`     // 格式: "18:00"
	Days    []string `json:"days"`    // 格式: ["monday", "tuesday", ...]
	Enabled bool     `json:"enabled"` // 是否启用
}

// GetTimeWindows 获取时间窗口配置
func (u *UserNotificationConfig) GetTimeWindows() ([]TimeWindow, error) {
	if u.TimeWindows == "" {
		return []TimeWindow{}, nil
	}
	var windows []TimeWindow
	err := json.Unmarshal([]byte(u.TimeWindows), &windows)
	return windows, err
}

// SetTimeWindows 设置时间窗口配置
func (u *UserNotificationConfig) SetTimeWindows(windows []TimeWindow) error {
	if windows == nil {
		u.TimeWindows = ""
		return nil
	}
	data, err := json.Marshal(windows)
	if err != nil {
		return err
	}
	u.TimeWindows = string(data)
	return nil
}

// IsInTimeWindow 检查当前时间是否在时间窗口内
func (u *UserNotificationConfig) IsInTimeWindow() (bool, error) {
	windows, err := u.GetTimeWindows()
	if err != nil {
		return false, err
	}
	
	if len(windows) == 0 {
		return true, nil // 没有配置时间窗口，默认允许
	}
	
	now := time.Now()
	currentDay := now.Weekday().String()
	currentTime := now.Format("15:04")
	
	for _, window := range windows {
		if !window.Enabled {
			continue
		}
		
		// 检查是否在指定的天
		dayMatch := false
		for _, day := range window.Days {
			if day == currentDay {
				dayMatch = true
				break
			}
		}
		
		if !dayMatch {
			continue
		}
		
		// 检查是否在时间范围内
		if currentTime >= window.Start && currentTime <= window.End {
			return true, nil
		}
	}
	
	return false, nil
}

// ShouldNotify 检查是否应该发送通知
func (u *UserNotificationConfig) ShouldNotify(alertLevel string) (bool, error) {
	if !u.Enabled {
		return false, nil
	}
	
	// 检查告警级别过滤
	levels, err := u.GetAlertLevels()
	if err != nil {
		return false, err
	}
	
	levelMatch := false
	for _, level := range levels {
		if level == alertLevel {
			levelMatch = true
			break
		}
	}
	
	if !levelMatch {
		return false, nil
	}
	
	// 检查时间窗口
	return u.IsInTimeWindow()
}

// NotificationRecord 通知发送记录
type NotificationRecord struct {
	ID             string         `json:"id" gorm:"type:varchar(36);primaryKey"`
	AlertID        *uint          `json:"alert_id" gorm:"index"`
	UserID         string         `json:"user_id" gorm:"type:varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;index"`
	PluginName     string         `json:"plugin_name" gorm:"type:varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;not null;index"`
	Channel        string         `json:"channel" gorm:"type:varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;not null"`
	MessageTitle   string         `json:"message_title" gorm:"type:varchar(500) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci"`
	MessageContent string         `json:"message_content" gorm:"type:text CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci"`
	Status         string         `json:"status" gorm:"type:varchar(20) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;not null;default:'pending';index"`
	Response       string         `json:"response" gorm:"type:text CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci"`
	RetryCount     int            `json:"retry_count" gorm:"default:0"`
	ErrorMessage   string         `json:"error_message" gorm:"type:text CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci"`
	SentAt         *time.Time     `json:"sent_at" gorm:"index"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
}

// TableName 指定表名
func (NotificationRecord) TableName() string {
	return "notification_records"
}

// BeforeCreate GORM钩子：创建前生成ID
func (n *NotificationRecord) BeforeCreate(tx *gorm.DB) error {
	if n.ID == "" {
		n.ID = uuid.New().String()
	}
	return nil
}

// IsSuccess 检查通知是否发送成功
func (n *NotificationRecord) IsSuccess() bool {
	return n.Status == "sent" || n.Status == "success"
}

// IsFailed 检查通知是否发送失败
func (n *NotificationRecord) IsFailed() bool {
	return n.Status == "failed"
}

// CanRetry 检查是否可以重试
func (n *NotificationRecord) CanRetry(maxRetry int) bool {
	return n.IsFailed() && n.RetryCount < maxRetry
}

// CreateUserNotificationConfigRequest 创建用户通知配置请求
type CreateUserNotificationConfigRequest struct {
	UserID      string                 `json:"user_id" binding:"required"`
	PluginName  string                 `json:"plugin_name" binding:"required"`
	Config      map[string]interface{} `json:"config" binding:"required"`
	Enabled     bool                   `json:"enabled"`
	AlertLevels []string               `json:"alert_levels"`
	TimeWindows []TimeWindow           `json:"time_windows"`
}

// UpdateUserNotificationConfigRequest 更新用户通知配置请求
type UpdateUserNotificationConfigRequest struct {
	Config      map[string]interface{} `json:"config"`
	Enabled     *bool                  `json:"enabled"`
	AlertLevels []string               `json:"alert_levels"`
	TimeWindows []TimeWindow           `json:"time_windows"`
}

// UserNotificationConfigQueryRequest 用户通知配置查询请求
type UserNotificationConfigQueryRequest struct {
	UserID     string `form:"user_id"`
	PluginName string `form:"plugin_name"`
	Enabled    *bool  `form:"enabled"`
	Page       int    `form:"page" binding:"min=1"`
	PageSize   int    `form:"page_size" binding:"min=1,max=100"`
}

// NotificationRecordQueryRequest 通知记录查询请求
type NotificationRecordQueryRequest struct {
	AlertID    *uint  `form:"alert_id"`
	UserID     string `form:"user_id"`
	PluginName string `form:"plugin_name"`
	Status     string `form:"status"`
	StartTime  string `form:"start_time"`
	EndTime    string `form:"end_time"`
	Page       int    `form:"page" binding:"min=1"`
	PageSize   int    `form:"page_size" binding:"min=1,max=100"`
}

// NotificationStats 通知统计
type NotificationStats struct {
	PluginName    string `json:"plugin_name"`
	TotalSent     int64  `json:"total_sent"`
	TotalFailed   int64  `json:"total_failed"`
	SuccessRate   float64 `json:"success_rate"`
	AvgRetryCount float64 `json:"avg_retry_count"`
}

// UserNotificationSummary 用户通知摘要
type UserNotificationSummary struct {
	UserID         string                    `json:"user_id"`
	TotalConfigs   int                       `json:"total_configs"`
	EnabledConfigs int                       `json:"enabled_configs"`
	Plugins        []string                  `json:"plugins"`
	RecentRecords  []NotificationRecord      `json:"recent_records"`
	Stats          []NotificationStats       `json:"stats"`
}