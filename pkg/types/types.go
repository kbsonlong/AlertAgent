package types

import (
	"time"
)

// Query 查询参数
type Query struct {
	Limit   int                    `json:"limit"`
	Offset  int                    `json:"offset"`
	OrderBy string                 `json:"order_by"`
	Filter  map[string]interface{} `json:"filter"`
	Search  string                 `json:"search"`
}

// PageResult 分页结果
type PageResult struct {
	Data  interface{} `json:"data"`
	Total int64       `json:"total"`
	Page  int         `json:"page"`
	Size  int         `json:"size"`
}

// APIResponse 统一API响应格式
type APIResponse struct {
	Status  string      `json:"status"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Error   *ErrorInfo  `json:"error,omitempty"`
}

// ErrorInfo 错误信息
type ErrorInfo struct {
	Type    string `json:"type"`
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

// NewSuccessResponse 创建成功响应
func NewSuccessResponse(message string, data interface{}) APIResponse {
	return APIResponse{
		Status:  "success",
		Message: message,
		Data:    data,
	}
}

// NewErrorResponse 创建错误响应
func NewErrorResponse(message string) APIResponse {
	return APIResponse{
		Status:  "error",
		Message: message,
		Error: &ErrorInfo{
			Type:    "error",
			Code:    "GENERAL_ERROR",
			Message: message,
		},
	}
}

// NewErrorResponseWithCode 创建带错误码的错误响应
func NewErrorResponseWithCode(message, code, errorType string) APIResponse {
	return APIResponse{
		Status:  "error",
		Message: message,
		Error: &ErrorInfo{
			Type:    errorType,
			Code:    code,
			Message: message,
		},
	}
}

// Message 消息结构
type Message struct {
	ID        string                 `json:"id"`
	Type      string                 `json:"type"`
	Title     string                 `json:"title"`
	Content   string                 `json:"content"`
	Data      map[string]interface{} `json:"data"`
	Priority  Priority               `json:"priority"`
	CreatedAt time.Time              `json:"created_at"`
}

// Priority 优先级
type Priority string

const (
	PriorityLow      Priority = "low"
	PriorityMedium   Priority = "medium"
	PriorityHigh     Priority = "high"
	PriorityCritical Priority = "critical"
)

// Config 配置接口
type Config interface {
	Validate() error
	GetString(key string) string
	GetInt(key string) int
	GetBool(key string) bool
}

// HealthStatus 健康状态
type HealthStatus struct {
	Status       string            `json:"status"`
	Timestamp    time.Time         `json:"timestamp"`
	Version      string            `json:"version"`
	Uptime       time.Duration     `json:"uptime"`
	Dependencies map[string]string `json:"dependencies"`
}

// Metadata 元数据
type Metadata struct {
	CreatedAt time.Time         `json:"created_at"`
	UpdatedAt time.Time         `json:"updated_at"`
	CreatedBy string            `json:"created_by"`
	UpdatedBy string            `json:"updated_by"`
	Version   int64             `json:"version"`
	Tags      []string          `json:"tags"`
	Labels    map[string]string `json:"labels"`
}

// Event 事件
type Event struct {
	ID        string                 `json:"id"`
	Type      string                 `json:"type"`
	Source    string                 `json:"source"`
	Subject   string                 `json:"subject"`
	Data      map[string]interface{} `json:"data"`
	Timestamp time.Time              `json:"timestamp"`
}

// RetryConfig 重试配置
type RetryConfig struct {
	MaxRetries int           `json:"max_retries"`
	Delay      time.Duration `json:"delay"`
	MaxDelay   time.Duration `json:"max_delay"`
	Backoff    float64       `json:"backoff"`
}

// CircuitBreakerConfig 熔断器配置
type CircuitBreakerConfig struct {
	MaxRequests      uint32        `json:"max_requests"`
	Interval         time.Duration `json:"interval"`
	Timeout          time.Duration `json:"timeout"`
	ReadyToTrip      func(counts Counts) bool
	OnStateChange    func(name string, from State, to State)
	IsSuccessful     func(err error) bool
}

// Counts 计数器
type Counts struct {
	Requests             uint32
	TotalSuccesses       uint32
	TotalFailures        uint32
	ConsecutiveSuccesses uint32
	ConsecutiveFailures  uint32
}

// State 熔断器状态
type State int

const (
	StateClosed State = iota
	StateHalfOpen
	StateOpen
)

// String 返回状态字符串
func (s State) String() string {
	switch s {
	case StateClosed:
		return "closed"
	case StateHalfOpen:
		return "half-open"
	case StateOpen:
		return "open"
	default:
		return "unknown"
	}
}