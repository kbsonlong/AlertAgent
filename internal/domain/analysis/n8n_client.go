package analysis

import (
	"context"
	"time"
)

// N8NWorkflowStatus 工作流执行状态
type N8NWorkflowStatus string

const (
	N8NWorkflowStatusPending   N8NWorkflowStatus = "pending"
	N8NWorkflowStatusRunning   N8NWorkflowStatus = "running"
	N8NWorkflowStatusCompleted N8NWorkflowStatus = "completed"
	N8NWorkflowStatusFailed    N8NWorkflowStatus = "failed"
	N8NWorkflowStatusCanceled  N8NWorkflowStatus = "canceled"
)

// N8NWorkflowExecution 工作流执行信息
type N8NWorkflowExecution struct {
	ID          string            `json:"id"`
	WorkflowID  string            `json:"workflow_id"`
	Status      N8NWorkflowStatus `json:"status"`
	StartedAt   time.Time         `json:"started_at"`
	FinishedAt  *time.Time        `json:"finished_at,omitempty"`
	InputData   map[string]interface{} `json:"input_data"`
	OutputData  map[string]interface{} `json:"output_data,omitempty"`
	ErrorData   *string           `json:"error_data,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// N8NWorkflowTriggerRequest 工作流触发请求
type N8NWorkflowTriggerRequest struct {
	WorkflowID string                 `json:"workflow_id"`
	InputData  map[string]interface{} `json:"input_data"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
	Callback   *N8NCallbackConfig     `json:"callback,omitempty"`
}

// N8NCallbackConfig 回调配置
type N8NCallbackConfig struct {
	URL     string            `json:"url"`
	Method  string            `json:"method"`
	Headers map[string]string `json:"headers,omitempty"`
	Secret  string            `json:"secret,omitempty"`
}

// N8NWorkflowTemplate 工作流模板
type N8NWorkflowTemplate struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Version     string                 `json:"version"`
	Nodes       []N8NWorkflowNode      `json:"nodes"`
	Connections map[string]interface{} `json:"connections"`
	Settings    map[string]interface{} `json:"settings"`
	Active      bool                   `json:"active"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

// N8NWorkflowNode 工作流节点
type N8NWorkflowNode struct {
	ID         string                 `json:"id"`
	Name       string                 `json:"name"`
	Type       string                 `json:"type"`
	Parameters map[string]interface{} `json:"parameters"`
	Position   []int                  `json:"position"`
}

// N8NHealthStatus n8n 健康状态
type N8NHealthStatus struct {
	Status      string    `json:"status"`
	Version     string    `json:"version"`
	Uptime      int64     `json:"uptime"`
	Connections int       `json:"connections"`
	Executions  int       `json:"executions"`
	Timestamp   time.Time `json:"timestamp"`
}

// N8NClient n8n 客户端接口
type N8NClient interface {
	// TriggerWorkflow 触发工作流
	TriggerWorkflow(ctx context.Context, req *N8NWorkflowTriggerRequest) (*N8NWorkflowExecution, error)

	// GetWorkflowExecution 获取工作流执行状态
	GetWorkflowExecution(ctx context.Context, executionID string) (*N8NWorkflowExecution, error)

	// CancelWorkflowExecution 取消工作流执行
	CancelWorkflowExecution(ctx context.Context, executionID string) error

	// RetryWorkflowExecution 重试工作流执行
	RetryWorkflowExecution(ctx context.Context, executionID string) (*N8NWorkflowExecution, error)

	// ListWorkflowExecutions 列出工作流执行历史
	ListWorkflowExecutions(ctx context.Context, workflowID string, limit, offset int) ([]*N8NWorkflowExecution, error)

	// GetWorkflowTemplate 获取工作流模板
	GetWorkflowTemplate(ctx context.Context, workflowID string) (*N8NWorkflowTemplate, error)

	// CreateWorkflowTemplate 创建工作流模板
	CreateWorkflowTemplate(ctx context.Context, template *N8NWorkflowTemplate) (*N8NWorkflowTemplate, error)

	// UpdateWorkflowTemplate 更新工作流模板
	UpdateWorkflowTemplate(ctx context.Context, workflowID string, template *N8NWorkflowTemplate) (*N8NWorkflowTemplate, error)

	// DeleteWorkflowTemplate 删除工作流模板
	DeleteWorkflowTemplate(ctx context.Context, workflowID string) error

	// ListWorkflowTemplates 列出工作流模板
	ListWorkflowTemplates(ctx context.Context, limit, offset int) ([]*N8NWorkflowTemplate, error)

	// GetHealth 获取 n8n 健康状态
	GetHealth(ctx context.Context) (*N8NHealthStatus, error)

	// RegisterCallback 注册回调处理器
	RegisterCallback(ctx context.Context, callbackID string, handler func(ctx context.Context, data map[string]interface{}) error) error

	// UnregisterCallback 注销回调处理器
	UnregisterCallback(ctx context.Context, callbackID string) error
}

// N8NWorkflowExecutionRepository 工作流执行记录存储接口
type N8NWorkflowExecutionRepository interface {
	// Create 创建工作流执行记录
	Create(ctx context.Context, execution *N8NWorkflowExecution) error

	// GetByID 根据ID获取工作流执行记录
	GetByID(ctx context.Context, id string) (*N8NWorkflowExecution, error)

	// Update 更新工作流执行记录
	Update(ctx context.Context, execution *N8NWorkflowExecution) error

	// Delete 删除工作流执行记录
	Delete(ctx context.Context, id string) error

	// ListByWorkflowID 根据工作流ID列出执行记录
	ListByWorkflowID(ctx context.Context, workflowID string, limit, offset int) ([]*N8NWorkflowExecution, error)

	// ListByStatus 根据状态列出执行记录
	ListByStatus(ctx context.Context, status N8NWorkflowStatus, limit, offset int) ([]*N8NWorkflowExecution, error)

	// ListByDateRange 根据日期范围列出执行记录
	ListByDateRange(ctx context.Context, startTime, endTime time.Time, limit, offset int) ([]*N8NWorkflowExecution, error)

	// GetStatistics 获取执行统计信息
	GetStatistics(ctx context.Context, workflowID string, startTime, endTime time.Time) (map[string]interface{}, error)
}

// N8NWorkflowManager 工作流管理器接口
type N8NWorkflowManager interface {
	// TriggerAnalysisWorkflow 触发分析工作流
	TriggerAnalysisWorkflow(ctx context.Context, alertID string, analysisType string, metadata map[string]interface{}) (*N8NWorkflowExecution, error)

	// MonitorExecution 监控工作流执行
	MonitorExecution(ctx context.Context, executionID string) (<-chan *N8NWorkflowExecution, error)

	// HandleCallback 处理工作流回调
	HandleCallback(ctx context.Context, executionID string, data map[string]interface{}) error

	// RetryFailedExecution 重试失败的执行
	RetryFailedExecution(ctx context.Context, executionID string) (*N8NWorkflowExecution, error)

	// GetExecutionLogs 获取执行日志
	GetExecutionLogs(ctx context.Context, executionID string) ([]string, error)

	// GetWorkflowMetrics 获取工作流指标
	GetWorkflowMetrics(ctx context.Context, workflowID string, timeRange time.Duration) (map[string]interface{}, error)
}