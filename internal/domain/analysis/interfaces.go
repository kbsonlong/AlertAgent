package analysis

import (
	"context"
	"time"
)

// AnalysisTaskQueue 分析任务队列接口
type AnalysisTaskQueue interface {
	// Push 推送任务到队列
	Push(ctx context.Context, task *AnalysisTask) error
	
	// Pop 从队列中获取任务
	Pop(ctx context.Context) (*AnalysisTask, error)
	
	// PopWithTimeout 带超时的获取任务
	PopWithTimeout(ctx context.Context, timeout time.Duration) (*AnalysisTask, error)
	
	// Peek 查看队列头部任务但不移除
	Peek(ctx context.Context) (*AnalysisTask, error)
	
	// Size 获取队列大小
	Size(ctx context.Context) (int64, error)
	
	// Clear 清空队列
	Clear(ctx context.Context) error
	
	// GetStatus 获取队列状态
	GetStatus(ctx context.Context) (*QueueStatus, error)
	
	// Remove 移除指定任务
	Remove(ctx context.Context, taskID string) error
	
	// UpdatePriority 更新任务优先级
	UpdatePriority(ctx context.Context, taskID string, priority int) error
}

// AnalysisTaskRepository 分析任务存储接口
type AnalysisTaskRepository interface {
	// Create 创建任务
	Create(ctx context.Context, task *AnalysisTask) error
	
	// GetByID 根据ID获取任务
	GetByID(ctx context.Context, taskID string) (*AnalysisTask, error)
	
	// Update 更新任务
	Update(ctx context.Context, task *AnalysisTask) error
	
	// Delete 删除任务
	Delete(ctx context.Context, taskID string) error
	
	// List 获取任务列表
	List(ctx context.Context, filter AnalysisFilter) ([]*AnalysisTask, error)
	
	// Count 统计任务数量
	Count(ctx context.Context, filter AnalysisFilter) (int64, error)
	
	// GetByAlertID 根据告警ID获取任务
	GetByAlertID(ctx context.Context, alertID string) ([]*AnalysisTask, error)
	
	// GetByStatus 根据状态获取任务
	GetByStatus(ctx context.Context, status AnalysisStatus) ([]*AnalysisTask, error)
	
	// UpdateStatus 更新任务状态
	UpdateStatus(ctx context.Context, taskID string, status AnalysisStatus) error
	
	// GetExpiredTasks 获取超时任务
	GetExpiredTasks(ctx context.Context) ([]*AnalysisTask, error)
}

// AnalysisResultRepository 分析结果存储接口
type AnalysisResultRepository interface {
	// Create 创建结果
	Create(ctx context.Context, result *AnalysisResult) error
	
	// GetByID 根据ID获取结果
	GetByID(ctx context.Context, resultID string) (*AnalysisResult, error)
	
	// GetByTaskID 根据任务ID获取结果
	GetByTaskID(ctx context.Context, taskID string) (*AnalysisResult, error)
	
	// GetByAlertID 根据告警ID获取结果
	GetByAlertID(ctx context.Context, alertID string) ([]*AnalysisResult, error)
	
	// Update 更新结果
	Update(ctx context.Context, result *AnalysisResult) error
	
	// Delete 删除结果
	Delete(ctx context.Context, resultID string) error
	
	// List 获取结果列表
	List(ctx context.Context, filter AnalysisFilter) ([]*AnalysisResult, error)
	
	// Count 统计结果数量
	Count(ctx context.Context, filter AnalysisFilter) (int64, error)
	
	// GetLatestByAlertID 获取告警的最新分析结果
	GetLatestByAlertID(ctx context.Context, alertID string, analysisType AnalysisType) (*AnalysisResult, error)
}

// AnalysisProgressTracker 分析进度跟踪接口
type AnalysisProgressTracker interface {
	// UpdateProgress 更新进度
	UpdateProgress(ctx context.Context, taskID string, progress *AnalysisProgress) error
	
	// GetProgress 获取进度
	GetProgress(ctx context.Context, taskID string) (*AnalysisProgress, error)
	
	// DeleteProgress 删除进度记录
	DeleteProgress(ctx context.Context, taskID string) error
	
	// GetProgressByTasks 批量获取任务进度
	GetProgressByTasks(ctx context.Context, taskIDs []string) (map[string]*AnalysisProgress, error)
}

// AnalysisEngine 分析引擎接口
type AnalysisEngine interface {
	// Analyze 执行分析
	Analyze(ctx context.Context, request *AnalysisRequest) (*AnalysisResult, error)
	
	// GetSupportedTypes 获取支持的分析类型
	GetSupportedTypes() []AnalysisType
	
	// ValidateRequest 验证分析请求
	ValidateRequest(request *AnalysisRequest) error
	
	// EstimateProcessingTime 估算处理时间
	EstimateProcessingTime(request *AnalysisRequest) time.Duration
	
	// GetEngineInfo 获取引擎信息
	GetEngineInfo() map[string]interface{}
}

// AnalysisWorker 分析工作器接口
type AnalysisWorker interface {
	// Start 启动工作器
	Start(ctx context.Context) error
	
	// Stop 停止工作器
	Stop(ctx context.Context) error
	
	// GetStatus 获取工作器状态
	GetStatus() *WorkerStatus
	
	// ProcessTask 处理单个任务
	ProcessTask(ctx context.Context, task *AnalysisTask) (*AnalysisResult, error)
	
	// GetID 获取工作器ID
	GetID() string
	
	// IsHealthy 检查工作器健康状态
	IsHealthy() bool
}

// AnalysisWorkerManager 工作器管理器接口
type AnalysisWorkerManager interface {
	// StartWorkers 启动工作器
	StartWorkers(ctx context.Context, count int) error
	
	// StopWorkers 停止工作器
	StopWorkers(ctx context.Context) error
	
	// GetWorkerStatuses 获取所有工作器状态
	GetWorkerStatuses() []*WorkerStatus
	
	// ScaleWorkers 动态调整工作器数量
	ScaleWorkers(ctx context.Context, targetCount int) error
	
	// GetActiveWorkerCount 获取活跃工作器数量
	GetActiveWorkerCount() int
	
	// RestartWorker 重启指定工作器
	RestartWorker(ctx context.Context, workerID string) error
	
	// GetWorkerMetrics 获取工作器指标
	GetWorkerMetrics() map[string]interface{}
}

// AnalysisService 分析服务接口
type AnalysisService interface {
	// SubmitAnalysis 提交分析请求
	SubmitAnalysis(ctx context.Context, request *AnalysisRequest) (*AnalysisTask, error)
	
	// GetAnalysisResult 获取分析结果
	GetAnalysisResult(ctx context.Context, taskID string) (*AnalysisResult, error)
	
	// GetAnalysisProgress 获取分析进度
	GetAnalysisProgress(ctx context.Context, taskID string) (*AnalysisProgress, error)
	
	// CancelAnalysis 取消分析任务
	CancelAnalysis(ctx context.Context, taskID string) error
	
	// RetryAnalysis 重试分析任务
	RetryAnalysis(ctx context.Context, taskID string) error
	
	// GetAnalysisTasks 获取分析任务列表
	GetAnalysisTasks(ctx context.Context, filter AnalysisFilter) ([]*AnalysisTask, error)
	
	// GetAnalysisResults 获取分析结果列表
	GetAnalysisResults(ctx context.Context, filter AnalysisFilter) ([]*AnalysisResult, error)
	
	// GetAnalysisStatistics 获取分析统计信息
	GetAnalysisStatistics(ctx context.Context, timeRange *TimeRange) (*AnalysisStatistics, error)
	
	// GetQueueStatus 获取队列状态
	GetQueueStatus(ctx context.Context) (*QueueStatus, error)
	
	// GetWorkerStatuses 获取工作器状态
	GetWorkerStatuses(ctx context.Context) ([]*WorkerStatus, error)
	
	// HealthCheck 健康检查
	HealthCheck(ctx context.Context) error
}

// AnalysisNotifier 分析通知接口
type AnalysisNotifier interface {
	// NotifyTaskCompleted 通知任务完成
	NotifyTaskCompleted(ctx context.Context, task *AnalysisTask, result *AnalysisResult) error
	
	// NotifyTaskFailed 通知任务失败
	NotifyTaskFailed(ctx context.Context, task *AnalysisTask, err error) error
	
	// NotifyProgressUpdate 通知进度更新
	NotifyProgressUpdate(ctx context.Context, progress *AnalysisProgress) error
	
	// RegisterCallback 注册回调
	RegisterCallback(taskID string, callback func(*AnalysisResult, error)) error
	
	// UnregisterCallback 取消注册回调
	UnregisterCallback(taskID string) error
}

// AnalysisMetricsCollector 分析指标收集器接口
type AnalysisMetricsCollector interface {
	// RecordTaskSubmitted 记录任务提交
	RecordTaskSubmitted(ctx context.Context, analysisType AnalysisType)
	
	// RecordTaskStarted 记录任务开始
	RecordTaskStarted(ctx context.Context, taskID string, analysisType AnalysisType)
	
	// RecordTaskCompleted 记录任务完成
	RecordTaskCompleted(ctx context.Context, taskID string, analysisType AnalysisType, duration time.Duration)
	
	// RecordTaskFailed 记录任务失败
	RecordTaskFailed(ctx context.Context, taskID string, analysisType AnalysisType, err error)
	
	// RecordQueueSize 记录队列大小
	RecordQueueSize(ctx context.Context, size int64)
	
	// RecordWorkerStatus 记录工作器状态
	RecordWorkerStatus(ctx context.Context, workerID string, status string)
	
	// GetMetrics 获取指标数据
	GetMetrics(ctx context.Context) map[string]interface{}
}

// TimeRange 时间范围
type TimeRange struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}

// AnalysisCallback 分析回调函数类型
type AnalysisCallback func(result *AnalysisResult, err error)

// AnalysisMiddleware 分析中间件接口
type AnalysisMiddleware interface {
	// BeforeAnalysis 分析前处理
	BeforeAnalysis(ctx context.Context, request *AnalysisRequest) error
	
	// AfterAnalysis 分析后处理
	AfterAnalysis(ctx context.Context, request *AnalysisRequest, result *AnalysisResult, err error) error
}

// AnalysisRetryPolicy 重试策略接口
type AnalysisRetryPolicy interface {
	// ShouldRetry 是否应该重试
	ShouldRetry(task *AnalysisTask, err error) bool
	
	// GetRetryDelay 获取重试延迟
	GetRetryDelay(task *AnalysisTask) time.Duration
	
	// GetMaxRetries 获取最大重试次数
	GetMaxRetries(analysisType AnalysisType) int
}