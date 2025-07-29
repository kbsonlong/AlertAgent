package analysis

import (
	"context"
	"time"
)

// DifyAnalysisService Dify AI 分析服务接口
type DifyAnalysisService interface {
	// AnalyzeAlert 分析告警（异步模式）
	AnalyzeAlert(ctx context.Context, request *DifyAnalysisRequest) (*DifyAnalysisTask, error)
	
	// GetAnalysisResult 获取分析结果
	GetAnalysisResult(ctx context.Context, taskID string) (*DifyAnalysisResult, error)
	
	// GetAnalysisProgress 获取分析进度
	GetAnalysisProgress(ctx context.Context, taskID string) (*DifyAnalysisProgress, error)
	
	// CancelAnalysis 取消分析任务
	CancelAnalysis(ctx context.Context, taskID string) error
	
	// RetryAnalysis 重试分析任务
	RetryAnalysis(ctx context.Context, taskID string) error
	
	// GetAnalysisHistory 获取分析历史
	GetAnalysisHistory(ctx context.Context, filter *DifyAnalysisFilter) ([]*DifyAnalysisResult, error)
	
	// GetAnalysisTrends 获取分析趋势
	GetAnalysisTrends(ctx context.Context, request *DifyTrendRequest) (*DifyTrendResponse, error)
	
	// SearchKnowledge 搜索知识库
	SearchKnowledge(ctx context.Context, query string, options *KnowledgeSearchOptions) (*KnowledgeSearchResult, error)
	
	// BuildAlertContext 构建告警上下文
	BuildAlertContext(ctx context.Context, alertID uint, options *ContextBuildOptions) (*AlertContext, error)
	
	// UpdateAlertWithAnalysis 将分析结果回写到告警记录
	UpdateAlertWithAnalysis(ctx context.Context, alertID uint, analysisResult *DifyAnalysisResult) error
	
	// GetAnalysisMetrics 获取分析指标
	GetAnalysisMetrics(ctx context.Context, timeRange *TimeRange) (*DifyAnalysisMetrics, error)
	
	// HealthCheck 健康检查
	HealthCheck(ctx context.Context) error
}

// DifyAnalysisRequest Dify 分析请求
type DifyAnalysisRequest struct {
	// 告警ID
	AlertID uint `json:"alert_id"`
	
	// 分析类型
	AnalysisType string `json:"analysis_type"`
	
	// 告警数据
	AlertData map[string]interface{} `json:"alert_data"`
	
	// 上下文数据
	Context *AlertContext `json:"context,omitempty"`
	
	// 分析选项
	Options *DifyAnalysisOptions `json:"options,omitempty"`
	
	// 用户ID
	UserID string `json:"user_id"`
	
	// 优先级
	Priority int `json:"priority,omitempty"`
	
	// 超时时间
	Timeout time.Duration `json:"timeout,omitempty"`
}

// DifyAnalysisTask Dify 分析任务
type DifyAnalysisTask struct {
	// 任务ID
	ID string `json:"id"`
	
	// 告警ID
	AlertID uint `json:"alert_id"`
	
	// 分析类型
	AnalysisType string `json:"analysis_type"`
	
	// 状态
	Status string `json:"status"`
	
	// 进度
	Progress int `json:"progress"`
	
	// 对话ID
	ConversationID string `json:"conversation_id,omitempty"`
	
	// 消息ID
	MessageID string `json:"message_id,omitempty"`
	
	// 工作流运行ID
	WorkflowRunID string `json:"workflow_run_id,omitempty"`
	
	// 创建时间
	CreatedAt time.Time `json:"created_at"`
	
	// 开始时间
	StartedAt *time.Time `json:"started_at,omitempty"`
	
	// 完成时间
	CompletedAt *time.Time `json:"completed_at,omitempty"`
	
	// 错误信息
	Error string `json:"error,omitempty"`
	
	// 重试次数
	RetryCount int `json:"retry_count"`
	
	// 最大重试次数
	MaxRetries int `json:"max_retries"`
}

// DifyAnalysisProgress Dify 分析进度
type DifyAnalysisProgress struct {
	// 任务ID
	TaskID string `json:"task_id"`
	
	// 状态
	Status string `json:"status"`
	
	// 进度百分比
	Progress int `json:"progress"`
	
	// 当前步骤
	CurrentStep string `json:"current_step"`
	
	// 总步骤数
	TotalSteps int `json:"total_steps"`
	
	// 已完成步骤数
	CompletedSteps int `json:"completed_steps"`
	
	// 预估剩余时间（秒）
	EstimatedRemainingTime int `json:"estimated_remaining_time"`
	
	// 详细信息
	Details map[string]interface{} `json:"details,omitempty"`
	
	// 更新时间
	UpdatedAt time.Time `json:"updated_at"`
}

// DifyAnalysisOptions Dify 分析选项
type DifyAnalysisOptions struct {
	// 使用的 Agent ID
	AgentID string `json:"agent_id,omitempty"`
	
	// 使用的工作流ID
	WorkflowID string `json:"workflow_id,omitempty"`
	
	// 是否包含历史上下文
	IncludeHistory bool `json:"include_history,omitempty"`
	
	// 历史上下文天数
	HistoryDays int `json:"history_days,omitempty"`
	
	// 是否包含相关告警
	IncludeRelatedAlerts bool `json:"include_related_alerts,omitempty"`
	
	// 是否包含系统指标
	IncludeMetrics bool `json:"include_metrics,omitempty"`
	
	// 是否包含日志
	IncludeLogs bool `json:"include_logs,omitempty"`
	
	// 自定义参数
	CustomParams map[string]interface{} `json:"custom_params,omitempty"`
	
	// 输出格式
	OutputFormat string `json:"output_format,omitempty"`
	
	// 语言
	Language string `json:"language,omitempty"`
}

// AlertContext 告警上下文
type AlertContext struct {
	// 告警基本信息
	Alert map[string]interface{} `json:"alert"`
	
	// 历史告警
	HistoricalAlerts []map[string]interface{} `json:"historical_alerts,omitempty"`
	
	// 相关告警
	RelatedAlerts []map[string]interface{} `json:"related_alerts,omitempty"`
	
	// 系统指标
	Metrics map[string]interface{} `json:"metrics,omitempty"`
	
	// 日志信息
	Logs []map[string]interface{} `json:"logs,omitempty"`
	
	// 服务信息
	Services []map[string]interface{} `json:"services,omitempty"`
	
	// 基础设施信息
	Infrastructure map[string]interface{} `json:"infrastructure,omitempty"`
	
	// 业务信息
	Business map[string]interface{} `json:"business,omitempty"`
	
	// 环境信息
	Environment map[string]interface{} `json:"environment,omitempty"`
	
	// 依赖关系
	Dependencies []map[string]interface{} `json:"dependencies,omitempty"`
}

// ContextBuildOptions 上下文构建选项
type ContextBuildOptions struct {
	// 包含历史告警
	IncludeHistory bool `json:"include_history,omitempty"`
	
	// 历史天数
	HistoryDays int `json:"history_days,omitempty"`
	
	// 包含相关告警
	IncludeRelated bool `json:"include_related,omitempty"`
	
	// 包含指标
	IncludeMetrics bool `json:"include_metrics,omitempty"`
	
	// 指标时间范围（分钟）
	MetricsTimeRange int `json:"metrics_time_range,omitempty"`
	
	// 包含日志
	IncludeLogs bool `json:"include_logs,omitempty"`
	
	// 日志时间范围（分钟）
	LogsTimeRange int `json:"logs_time_range,omitempty"`
	
	// 包含服务信息
	IncludeServices bool `json:"include_services,omitempty"`
	
	// 包含基础设施信息
	IncludeInfrastructure bool `json:"include_infrastructure,omitempty"`
	
	// 包含业务信息
	IncludeBusiness bool `json:"include_business,omitempty"`
	
	// 包含依赖关系
	IncludeDependencies bool `json:"include_dependencies,omitempty"`
}

// KnowledgeSearchOptions 知识库搜索选项
type KnowledgeSearchOptions struct {
	// 搜索类型
	SearchType string `json:"search_type,omitempty"`
	
	// 数据集ID列表
	DatasetIDs []string `json:"dataset_ids,omitempty"`
	
	// 最大结果数
	Limit int `json:"limit,omitempty"`
	
	// 相似度阈值
	SimilarityThreshold float64 `json:"similarity_threshold,omitempty"`
	
	// 过滤条件
	Filters map[string]interface{} `json:"filters,omitempty"`
}

// KnowledgeSearchResult 知识库搜索结果
type KnowledgeSearchResult struct {
	// 查询
	Query string `json:"query"`
	
	// 结果列表
	Results []KnowledgeItem `json:"results"`
	
	// 总数
	Total int `json:"total"`
	
	// 搜索时间（毫秒）
	SearchTime int64 `json:"search_time"`
}

// KnowledgeItem 知识条目
type KnowledgeItem struct {
	// ID
	ID string `json:"id"`
	
	// 标题
	Title string `json:"title"`
	
	// 内容
	Content string `json:"content"`
	
	// 相似度分数
	Score float64 `json:"score"`
	
	// 数据集ID
	DatasetID string `json:"dataset_id"`
	
	// 数据集名称
	DatasetName string `json:"dataset_name"`
	
	// 文档ID
	DocumentID string `json:"document_id"`
	
	// 文档名称
	DocumentName string `json:"document_name"`
	
	// 段ID
	SegmentID string `json:"segment_id"`
	
	// 元数据
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// DifyTrendRequest Dify 趋势分析请求
type DifyTrendRequest struct {
	// 时间范围
	TimeRange *TimeRange `json:"time_range"`
	
	// 分析维度
	Dimensions []string `json:"dimensions,omitempty"`
	
	// 过滤条件
	Filters map[string]interface{} `json:"filters,omitempty"`
	
	// 聚合方式
	Aggregation string `json:"aggregation,omitempty"`
	
	// 分组字段
	GroupBy []string `json:"group_by,omitempty"`
}

// DifyTrendResponse Dify 趋势分析响应
type DifyTrendResponse struct {
	// 趋势数据
	Trends []TrendData `json:"trends"`
	
	// 统计信息
	Statistics map[string]interface{} `json:"statistics"`
	
	// 洞察
	Insights []string `json:"insights,omitempty"`
	
	// 建议
	Recommendations []string `json:"recommendations,omitempty"`
}

// TrendData 趋势数据
type TrendData struct {
	// 时间戳
	Timestamp time.Time `json:"timestamp"`
	
	// 值
	Value float64 `json:"value"`
	
	// 标签
	Labels map[string]string `json:"labels,omitempty"`
	
	// 元数据
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// DifyAnalysisMetrics Dify 分析指标
type DifyAnalysisMetrics struct {
	// 总分析次数
	TotalAnalyses int64 `json:"total_analyses"`
	
	// 成功分析次数
	SuccessfulAnalyses int64 `json:"successful_analyses"`
	
	// 失败分析次数
	FailedAnalyses int64 `json:"failed_analyses"`
	
	// 平均处理时间（毫秒）
	AverageProcessingTime float64 `json:"average_processing_time"`
	
	// 平均置信度
	AverageConfidence float64 `json:"average_confidence"`
	
	// 总令牌使用量
	TotalTokenUsage int64 `json:"total_token_usage"`
	
	// 总成本
	TotalCost float64 `json:"total_cost"`
	
	// 按分析类型统计
	ByAnalysisType map[string]*AnalysisTypeMetrics `json:"by_analysis_type"`
	
	// 按时间统计
	ByTime []TimeMetrics `json:"by_time"`
	
	// 错误统计
	ErrorStats map[string]int64 `json:"error_stats"`
}

// AnalysisTypeMetrics 分析类型指标
type AnalysisTypeMetrics struct {
	// 分析类型
	Type string `json:"type"`
	
	// 总次数
	Total int64 `json:"total"`
	
	// 成功次数
	Successful int64 `json:"successful"`
	
	// 失败次数
	Failed int64 `json:"failed"`
	
	// 平均处理时间
	AverageTime float64 `json:"average_time"`
	
	// 平均置信度
	AverageConfidence float64 `json:"average_confidence"`
}

// TimeMetrics 时间指标
type TimeMetrics struct {
	// 时间戳
	Timestamp time.Time `json:"timestamp"`
	
	// 分析次数
	Count int64 `json:"count"`
	
	// 成功次数
	Successful int64 `json:"successful"`
	
	// 失败次数
	Failed int64 `json:"failed"`
	
	// 平均处理时间
	AverageTime float64 `json:"average_time"`
}

// 常量定义
const (
	// 分析状态
	DifyAnalysisStatusPending    = "pending"
	DifyAnalysisStatusRunning    = "running"
	DifyAnalysisStatusCompleted  = "completed"
	DifyAnalysisStatusFailed     = "failed"
	DifyAnalysisStatusCancelled  = "cancelled"
	
	// 分析类型
	DifyAnalysisTypeRootCause    = "root_cause"
	DifyAnalysisTypeImpact       = "impact"
	DifyAnalysisTypeRecommendation = "recommendation"
	DifyAnalysisTypeTrend        = "trend"
	DifyAnalysisTypeClassification = "classification"
	
	// 响应模式
	DifyResponseModeBlocking     = "blocking"
	DifyResponseModeStreaming    = "streaming"
	
	// 搜索类型
	KnowledgeSearchTypeSemantic  = "semantic"
	KnowledgeSearchTypeKeyword   = "keyword"
	KnowledgeSearchTypeHybrid    = "hybrid"
)