package model

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// 任务状态常量
const (
	TaskStatusPending    = "pending"
	TaskStatusProcessing = "processing"
	TaskStatusCompleted  = "completed"
	TaskStatusFailed     = "failed"
)

// 任务类型常量
const (
	TaskTypeAIAnalysis     = "ai_analysis"
	TaskTypeNotification   = "notification"
	TaskTypeConfigSync     = "config_sync"
	TaskTypeRuleDistribute = "rule_distribute"
)

// TaskQueue 任务队列
type TaskQueue struct {
	ID          string         `json:"id" gorm:"type:varchar(36);primaryKey"`
	QueueName   string         `json:"queue_name" gorm:"type:varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;not null;index"`
	TaskType    string         `json:"task_type" gorm:"type:varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;not null;index"`
	Payload     string         `json:"payload" gorm:"type:json;not null"`
	Priority    int            `json:"priority" gorm:"default:0;index"`
	RetryCount  int            `json:"retry_count" gorm:"default:0"`
	MaxRetry    int            `json:"max_retry" gorm:"default:3"`
	Status      string         `json:"status" gorm:"type:varchar(20) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;not null;default:'pending';index"`
	ScheduledAt time.Time      `json:"scheduled_at" gorm:"index"`
	StartedAt   *time.Time     `json:"started_at"`
	CompletedAt *time.Time     `json:"completed_at"`
	ErrorMessage string        `json:"error_message" gorm:"type:text CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci"`
	WorkerID    string         `json:"worker_id" gorm:"type:varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;index"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
}

// TableName 指定表名
func (TaskQueue) TableName() string {
	return "task_queue"
}

// BeforeCreate GORM钩子：创建前生成ID
func (t *TaskQueue) BeforeCreate(tx *gorm.DB) error {
	if t.ID == "" {
		t.ID = uuid.New().String()
	}
	if t.ScheduledAt.IsZero() {
		t.ScheduledAt = time.Now()
	}
	return nil
}

// GetPayloadMap 获取载荷映射
func (t *TaskQueue) GetPayloadMap() (map[string]interface{}, error) {
	var payload map[string]interface{}
	err := json.Unmarshal([]byte(t.Payload), &payload)
	return payload, err
}

// SetPayloadMap 设置载荷映射
func (t *TaskQueue) SetPayloadMap(payload map[string]interface{}) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	t.Payload = string(data)
	return nil
}

// GetDuration 获取执行时长（毫秒）
func (t *TaskQueue) GetDuration() int64 {
	if t.StartedAt == nil || t.CompletedAt == nil {
		return 0
	}
	return t.CompletedAt.Sub(*t.StartedAt).Milliseconds()
}

// IsExpired 检查任务是否过期
func (t *TaskQueue) IsExpired(timeoutMinutes int) bool {
	if t.Status != TaskStatusProcessing {
		return false
	}
	if t.StartedAt == nil {
		return false
	}
	return time.Since(*t.StartedAt).Minutes() > float64(timeoutMinutes)
}

// WorkerInstance Worker实例
type WorkerInstance struct {
	ID             string         `json:"id" gorm:"type:varchar(36);primaryKey"`
	WorkerID       string         `json:"worker_id" gorm:"type:varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;not null;uniqueIndex"`
	QueueName      string         `json:"queue_name" gorm:"type:varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;not null;index"`
	HostName       string         `json:"host_name" gorm:"type:varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;not null"`
	ProcessID      int            `json:"process_id" gorm:"not null"`
	Concurrency    int            `json:"concurrency" gorm:"not null;default:1"`
	Status         string         `json:"status" gorm:"type:varchar(20) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;not null;default:'active';index"`
	LastHeartbeat  time.Time      `json:"last_heartbeat" gorm:"index"`
	TasksProcessed int64          `json:"tasks_processed" gorm:"default:0"`
	TasksFailed    int64          `json:"tasks_failed" gorm:"default:0"`
	StartedAt      time.Time      `json:"started_at"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
}

// TableName 指定表名
func (WorkerInstance) TableName() string {
	return "worker_instances"
}

// BeforeCreate GORM钩子：创建前生成ID
func (w *WorkerInstance) BeforeCreate(tx *gorm.DB) error {
	if w.ID == "" {
		w.ID = uuid.New().String()
	}
	if w.LastHeartbeat.IsZero() {
		w.LastHeartbeat = time.Now()
	}
	if w.StartedAt.IsZero() {
		w.StartedAt = time.Now()
	}
	return nil
}

// GetFailureRate 获取失败率
func (w *WorkerInstance) GetFailureRate() float64 {
	if w.TasksProcessed == 0 {
		return 0
	}
	return float64(w.TasksFailed) / float64(w.TasksProcessed) * 100
}

// GetUptimeHours 获取运行时长（小时）
func (w *WorkerInstance) GetUptimeHours() float64 {
	return time.Since(w.StartedAt).Hours()
}

// IsHealthy 检查Worker是否健康
func (w *WorkerInstance) IsHealthy(heartbeatTimeoutSeconds int) bool {
	if w.Status != "active" {
		return false
	}
	return time.Since(w.LastHeartbeat).Seconds() <= float64(heartbeatTimeoutSeconds)
}

// TaskExecutionHistory 任务执行历史
type TaskExecutionHistory struct {
	ID          string         `json:"id" gorm:"type:varchar(36);primaryKey"`
	TaskID      string         `json:"task_id" gorm:"type:varchar(36);not null;index"`
	WorkerID    string         `json:"worker_id" gorm:"type:varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;not null;index"`
	QueueName   string         `json:"queue_name" gorm:"type:varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;not null;index"`
	TaskType    string         `json:"task_type" gorm:"type:varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;not null;index"`
	Status      string         `json:"status" gorm:"type:varchar(20) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;not null;index"`
	StartedAt   time.Time      `json:"started_at" gorm:"not null;index"`
	CompletedAt *time.Time     `json:"completed_at"`
	DurationMs  int64          `json:"duration_ms" gorm:"default:0"`
	ErrorMessage string        `json:"error_message" gorm:"type:text CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci"`
	Result      string         `json:"result" gorm:"type:json"`
	CreatedAt   time.Time      `json:"created_at"`
}

// TableName 指定表名
func (TaskExecutionHistory) TableName() string {
	return "task_execution_history"
}

// BeforeCreate GORM钩子：创建前生成ID
func (t *TaskExecutionHistory) BeforeCreate(tx *gorm.DB) error {
	if t.ID == "" {
		t.ID = uuid.New().String()
	}
	return nil
}

// GetResultMap 获取结果映射
func (t *TaskExecutionHistory) GetResultMap() (map[string]interface{}, error) {
	if t.Result == "" {
		return make(map[string]interface{}), nil
	}
	var result map[string]interface{}
	err := json.Unmarshal([]byte(t.Result), &result)
	return result, err
}

// SetResultMap 设置结果映射
func (t *TaskExecutionHistory) SetResultMap(result map[string]interface{}) error {
	if result == nil {
		t.Result = ""
		return nil
	}
	data, err := json.Marshal(result)
	if err != nil {
		return err
	}
	t.Result = string(data)
	return nil
}

// CalculateDuration 计算执行时长
func (t *TaskExecutionHistory) CalculateDuration() {
	if t.CompletedAt != nil {
		t.DurationMs = t.CompletedAt.Sub(t.StartedAt).Milliseconds()
	}
}

// TaskQueueStats 任务队列统计
type TaskQueueStats struct {
	QueueName     string    `json:"queue_name"`
	TaskType      string    `json:"task_type"`
	Status        string    `json:"status"`
	TaskCount     int64     `json:"task_count"`
	AvgDurationMs float64   `json:"avg_duration_ms"`
	OldestTask    time.Time `json:"oldest_task"`
	NewestTask    time.Time `json:"newest_task"`
}

// WorkerPerformance Worker性能统计
type WorkerPerformance struct {
	WorkerID              string  `json:"worker_id"`
	QueueName             string  `json:"queue_name"`
	HostName              string  `json:"host_name"`
	Concurrency           int     `json:"concurrency"`
	Status                string  `json:"status"`
	TasksProcessed        int64   `json:"tasks_processed"`
	TasksFailed           int64   `json:"tasks_failed"`
	FailureRatePercent    float64 `json:"failure_rate_percent"`
	HeartbeatDelaySeconds int64   `json:"heartbeat_delay_seconds"`
	UptimeHours           float64 `json:"uptime_hours"`
}

// CreateTaskRequest 创建任务请求
type CreateTaskRequest struct {
	QueueName   string                 `json:"queue_name" binding:"required"`
	TaskType    string                 `json:"task_type" binding:"required"`
	Payload     map[string]interface{} `json:"payload" binding:"required"`
	Priority    int                    `json:"priority"`
	MaxRetry    int                    `json:"max_retry"`
	ScheduledAt *time.Time             `json:"scheduled_at"`
}

// UpdateTaskRequest 更新任务请求
type UpdateTaskRequest struct {
	Status       string                 `json:"status"`
	ErrorMessage string                 `json:"error_message"`
	Result       map[string]interface{} `json:"result"`
	WorkerID     string                 `json:"worker_id"`
}

// TaskQueryRequest 任务查询请求
type TaskQueryRequest struct {
	QueueName string `form:"queue_name"`
	TaskType  string `form:"task_type"`
	Status    string `form:"status"`
	WorkerID  string `form:"worker_id"`
	Page      int    `form:"page" binding:"min=1"`
	PageSize  int    `form:"page_size" binding:"min=1,max=100"`
}