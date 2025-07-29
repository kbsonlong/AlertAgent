package analysis

import (
	"time"

	"alert_agent/internal/model"
)

// AnalysisStatus 分析状态
type AnalysisStatus string

const (
	AnalysisStatusPending    AnalysisStatus = "pending"     // 待处理
	AnalysisStatusProcessing AnalysisStatus = "processing"  // 处理中
	AnalysisStatusCompleted  AnalysisStatus = "completed"   // 已完成
	AnalysisStatusFailed     AnalysisStatus = "failed"      // 失败
	AnalysisStatusCancelled  AnalysisStatus = "cancelled"   // 已取消
)

// AnalysisType 分析类型
type AnalysisType string

const (
	AnalysisTypeRootCause     AnalysisType = "root_cause_analysis"     // 根因分析
	AnalysisTypeImpactAssess  AnalysisType = "impact_assessment"       // 影响评估
	AnalysisTypeSolution      AnalysisType = "solution_recommendation"  // 解决方案推荐
	AnalysisTypeClassification AnalysisType = "classification"          // 告警分类
	AnalysisTypePriority      AnalysisType = "priority_assessment"     // 优先级评估
)

// AnalysisTask 分析任务
type AnalysisTask struct {
	ID          string        `json:"id"`           // 任务ID
	AlertID     string        `json:"alert_id"`     // 告警ID
	Type        AnalysisType  `json:"type"`         // 分析类型
	Status      AnalysisStatus `json:"status"`       // 任务状态
	Priority    int           `json:"priority"`     // 任务优先级 (1-10, 10最高)
	RetryCount  int           `json:"retry_count"`  // 重试次数
	MaxRetries  int           `json:"max_retries"`  // 最大重试次数
	Timeout     time.Duration `json:"timeout"`      // 超时时间
	CreatedAt   time.Time     `json:"created_at"`   // 创建时间
	UpdatedAt   time.Time     `json:"updated_at"`   // 更新时间
	StartedAt   *time.Time    `json:"started_at"`   // 开始处理时间
	CompletedAt *time.Time    `json:"completed_at"` // 完成时间
	Metadata    map[string]interface{} `json:"metadata"` // 任务元数据
}

// AnalysisResult 分析结果
type AnalysisResult struct {
	ID              string                 `json:"id"`               // 结果ID
	TaskID          string                 `json:"task_id"`          // 任务ID
	AlertID         string                 `json:"alert_id"`         // 告警ID
	Type            AnalysisType           `json:"type"`             // 分析类型
	Status          AnalysisStatus         `json:"status"`           // 分析状态
	ConfidenceScore float64                `json:"confidence_score"` // 置信度分数 (0-1)
	ProcessingTime  time.Duration          `json:"processing_time"`  // 处理耗时
	Result          map[string]interface{} `json:"result"`           // 分析结果数据
	Summary         string                 `json:"summary"`          // 结果摘要
	Recommendations []string               `json:"recommendations"`  // 推荐操作
	ErrorMessage    string                 `json:"error_message"`    // 错误信息
	CreatedAt       time.Time              `json:"created_at"`       // 创建时间
	UpdatedAt       time.Time              `json:"updated_at"`       // 更新时间
	Metadata        map[string]interface{} `json:"metadata"`         // 结果元数据
}

// AnalysisProgress 分析进度
type AnalysisProgress struct {
	TaskID      string    `json:"task_id"`      // 任务ID
	Stage       string    `json:"stage"`        // 当前阶段
	Progress    float64   `json:"progress"`     // 进度百分比 (0-100)
	Message     string    `json:"message"`      // 进度消息
	UpdatedAt   time.Time `json:"updated_at"`   // 更新时间
}

// AnalysisRequest 分析请求
type AnalysisRequest struct {
	Alert       *model.Alert           `json:"alert"`        // 告警信息
	Type        AnalysisType           `json:"type"`         // 分析类型
	Priority    int                    `json:"priority"`     // 优先级
	Timeout     time.Duration          `json:"timeout"`      // 超时时间
	Options     map[string]interface{} `json:"options"`      // 分析选项
	Callback    string                 `json:"callback"`     // 回调URL
}

// AnalysisFilter 分析查询过滤器
type AnalysisFilter struct {
	TaskIDs    []string        `json:"task_ids"`    // 任务ID列表
	AlertIDs   []string        `json:"alert_ids"`   // 告警ID列表
	Types      []AnalysisType  `json:"types"`       // 分析类型列表
	Statuses   []AnalysisStatus `json:"statuses"`    // 状态列表
	StartTime  *time.Time      `json:"start_time"`  // 开始时间
	EndTime    *time.Time      `json:"end_time"`    // 结束时间
	Limit      int             `json:"limit"`       // 限制数量
	Offset     int             `json:"offset"`      // 偏移量
}

// AnalysisStatistics 分析统计信息
type AnalysisStatistics struct {
	TotalTasks      int64                        `json:"total_tasks"`       // 总任务数
	PendingTasks    int64                        `json:"pending_tasks"`     // 待处理任务数
	ProcessingTasks int64                        `json:"processing_tasks"`  // 处理中任务数
	CompletedTasks  int64                        `json:"completed_tasks"`   // 已完成任务数
	FailedTasks     int64                        `json:"failed_tasks"`      // 失败任务数
	AverageTime     time.Duration                `json:"average_time"`      // 平均处理时间
	SuccessRate     float64                      `json:"success_rate"`      // 成功率
	TypeDistribution map[AnalysisType]int64      `json:"type_distribution"` // 类型分布
	StatusDistribution map[AnalysisStatus]int64  `json:"status_distribution"` // 状态分布
	LastUpdated     time.Time                    `json:"last_updated"`      // 最后更新时间
}

// WorkerStatus 工作器状态
type WorkerStatus struct {
	ID              string    `json:"id"`               // 工作器ID
	Status          string    `json:"status"`           // 状态 (running, stopped, error)
	CurrentTask     string    `json:"current_task"`     // 当前处理任务ID
	ProcessedCount  int64     `json:"processed_count"`  // 已处理任务数
	ErrorCount      int64     `json:"error_count"`      // 错误次数
	LastActiveTime  time.Time `json:"last_active_time"` // 最后活跃时间
	StartTime       time.Time `json:"start_time"`       // 启动时间
	Metadata        map[string]interface{} `json:"metadata"` // 工作器元数据
}

// QueueStatus 队列状态
type QueueStatus struct {
	PendingCount    int64     `json:"pending_count"`    // 待处理任务数
	ProcessingCount int64     `json:"processing_count"` // 处理中任务数
	CompletedCount  int64     `json:"completed_count"`  // 已完成任务数
	FailedCount     int64     `json:"failed_count"`     // 失败任务数
	TotalCount      int64     `json:"total_count"`      // 总任务数
	OldestTask      *time.Time `json:"oldest_task"`      // 最老任务时间
	LastUpdated     time.Time `json:"last_updated"`     // 最后更新时间
}