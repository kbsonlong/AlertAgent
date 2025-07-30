package queue

import (
	"encoding/json"
	"time"
)

// TaskType 任务类型
type TaskType string

const (
	TaskTypeAIAnalysis    TaskType = "ai_analysis"
	TaskTypeNotification  TaskType = "notification"
	TaskTypeConfigSync    TaskType = "config_sync"
	TaskTypeRuleUpdate    TaskType = "rule_update"
	TaskTypeHealthCheck   TaskType = "health_check"
)

// TaskStatus 任务状态
type TaskStatus string

const (
	TaskStatusPending    TaskStatus = "pending"
	TaskStatusProcessing TaskStatus = "processing"
	TaskStatusCompleted  TaskStatus = "completed"
	TaskStatusFailed     TaskStatus = "failed"
	TaskStatusRetrying   TaskStatus = "retrying"
)

// TaskPriority 任务优先级
type TaskPriority int

const (
	PriorityLow    TaskPriority = 0
	PriorityNormal TaskPriority = 1
	PriorityHigh   TaskPriority = 2
	PriorityCritical TaskPriority = 3
)

// Task 通用任务结构
type Task struct {
	ID          string                 `json:"id"`
	Type        TaskType               `json:"type"`
	Payload     map[string]interface{} `json:"payload"`
	Priority    TaskPriority           `json:"priority"`
	Status      TaskStatus             `json:"status"`
	Retry       int                    `json:"retry"`
	MaxRetry    int                    `json:"max_retry"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
	ScheduledAt time.Time              `json:"scheduled_at"`
	StartedAt   *time.Time             `json:"started_at,omitempty"`
	CompletedAt *time.Time             `json:"completed_at,omitempty"`
	ErrorMsg    string                 `json:"error_msg,omitempty"`
	WorkerID    string                 `json:"worker_id,omitempty"`
}

// TaskResult 任务执行结果
type TaskResult struct {
	TaskID      string                 `json:"task_id"`
	Status      TaskStatus             `json:"status"`
	Result      map[string]interface{} `json:"result,omitempty"`
	ErrorMsg    string                 `json:"error_msg,omitempty"`
	Duration    time.Duration          `json:"duration"`
	CompletedAt time.Time              `json:"completed_at"`
	WorkerID    string                 `json:"worker_id"`
}

// AIAnalysisPayload AI分析任务载荷
type AIAnalysisPayload struct {
	AlertID      string                 `json:"alert_id"`
	AlertData    map[string]interface{} `json:"alert_data"`
	AnalysisType string                 `json:"analysis_type"`
	Options      map[string]interface{} `json:"options,omitempty"`
}

// NotificationPayload 通知任务载荷
type NotificationPayload struct {
	AlertID   string                 `json:"alert_id"`
	Channels  []string               `json:"channels"`
	Message   map[string]interface{} `json:"message"`
	Template  string                 `json:"template,omitempty"`
	Variables map[string]interface{} `json:"variables,omitempty"`
}

// ConfigSyncPayload 配置同步任务载荷
type ConfigSyncPayload struct {
	Type      string   `json:"type"`       // rule_create, rule_update, rule_delete
	RuleID    string   `json:"rule_id"`
	Targets   []string `json:"targets"`
	ConfigData map[string]interface{} `json:"config_data,omitempty"`
}

// ToJSON 将任务转换为JSON
func (t *Task) ToJSON() ([]byte, error) {
	return json.Marshal(t)
}

// FromJSON 从JSON创建任务
func (t *Task) FromJSON(data []byte) error {
	return json.Unmarshal(data, t)
}

// IsExpired 检查任务是否过期
func (t *Task) IsExpired(timeout time.Duration) bool {
	return time.Since(t.CreatedAt) > timeout
}

// ShouldRetry 检查是否应该重试
func (t *Task) ShouldRetry() bool {
	return t.Retry < t.MaxRetry && t.Status == TaskStatusFailed
}

// IncrementRetry 增加重试次数
func (t *Task) IncrementRetry() {
	t.Retry++
	t.Status = TaskStatusRetrying
	t.UpdatedAt = time.Now()
}

// MarkProcessing 标记为处理中
func (t *Task) MarkProcessing(workerID string) {
	now := time.Now()
	t.Status = TaskStatusProcessing
	t.WorkerID = workerID
	t.StartedAt = &now
	t.UpdatedAt = now
}

// MarkCompleted 标记为完成
func (t *Task) MarkCompleted() {
	now := time.Now()
	t.Status = TaskStatusCompleted
	t.CompletedAt = &now
	t.UpdatedAt = now
}

// MarkFailed 标记为失败
func (t *Task) MarkFailed(errorMsg string) {
	now := time.Now()
	t.Status = TaskStatusFailed
	t.ErrorMsg = errorMsg
	t.CompletedAt = &now
	t.UpdatedAt = now
}

// GetDuration 获取任务执行时长
func (t *Task) GetDuration() time.Duration {
	if t.StartedAt == nil {
		return 0
	}
	if t.CompletedAt == nil {
		return time.Since(*t.StartedAt)
	}
	return t.CompletedAt.Sub(*t.StartedAt)
}

// String 返回任务类型字符串
func (tt TaskType) String() string {
	return string(tt)
}