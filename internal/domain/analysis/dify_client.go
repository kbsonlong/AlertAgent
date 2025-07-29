package analysis

import (
	"context"
	"time"
)

// DifyClient Dify AI 客户端接口
type DifyClient interface {
	// ChatMessage 发送聊天消息到 Dify Agent
	ChatMessage(ctx context.Context, request *DifyChatRequest) (*DifyChatResponse, error)
	
	// RunWorkflow 执行 Dify 工作流
	RunWorkflow(ctx context.Context, request *DifyWorkflowRequest) (*DifyWorkflowResponse, error)
	
	// GetConversation 获取对话历史
	GetConversation(ctx context.Context, conversationID string) (*DifyConversation, error)
	
	// HealthCheck 健康检查
	HealthCheck(ctx context.Context) error
	
	// GetModelInfo 获取模型信息
	GetModelInfo(ctx context.Context) (*DifyModelInfo, error)
}

// DifyChatRequest Dify 聊天请求
type DifyChatRequest struct {
	// 输入参数
	Inputs map[string]interface{} `json:"inputs"`
	
	// 用户查询内容
	Query string `json:"query"`
	
	// 响应模式：blocking（同步）或 streaming（流式）
	ResponseMode string `json:"response_mode"`
	
	// 用户标识
	User string `json:"user"`
	
	// 对话ID（可选，用于多轮对话）
	ConversationID string `json:"conversation_id,omitempty"`
	
	// 文件列表（可选）
	Files []DifyFile `json:"files,omitempty"`
	
	// 自动生成名称
	AutoGenerateName bool `json:"auto_generate_name,omitempty"`
}

// DifyChatResponse Dify 聊天响应
type DifyChatResponse struct {
	// 消息ID
	MessageID string `json:"message_id"`
	
	// 对话ID
	ConversationID string `json:"conversation_id"`
	
	// 模式
	Mode string `json:"mode"`
	
	// 回答内容
	Answer string `json:"answer"`
	
	// 元数据
	Metadata *DifyMetadata `json:"metadata,omitempty"`
	
	// 创建时间
	CreatedAt time.Time `json:"created_at"`
	
	// 使用情况
	Usage *DifyUsage `json:"usage,omitempty"`
}

// DifyWorkflowRequest Dify 工作流请求
type DifyWorkflowRequest struct {
	// 输入参数
	Inputs map[string]interface{} `json:"inputs"`
	
	// 响应模式
	ResponseMode string `json:"response_mode"`
	
	// 用户标识
	User string `json:"user"`
	
	// 文件列表（可选）
	Files []DifyFile `json:"files,omitempty"`
}

// DifyWorkflowResponse Dify 工作流响应
type DifyWorkflowResponse struct {
	// 工作流运行ID
	WorkflowRunID string `json:"workflow_run_id"`
	
	// 任务ID
	TaskID string `json:"task_id"`
	
	// 数据
	Data map[string]interface{} `json:"data"`
	
	// 创建时间
	CreatedAt time.Time `json:"created_at"`
	
	// 完成时间
	FinishedAt *time.Time `json:"finished_at,omitempty"`
	
	// 状态
	Status string `json:"status"`
	
	// 错误信息
	Error string `json:"error,omitempty"`
	
	// 使用情况
	Usage *DifyUsage `json:"usage,omitempty"`
}

// DifyConversation Dify 对话
type DifyConversation struct {
	// 对话ID
	ID string `json:"id"`
	
	// 名称
	Name string `json:"name"`
	
	// 输入参数
	Inputs map[string]interface{} `json:"inputs"`
	
	// 状态
	Status string `json:"status"`
	
	// 创建时间
	CreatedAt time.Time `json:"created_at"`
	
	// 更新时间
	UpdatedAt time.Time `json:"updated_at"`
}

// DifyFile Dify 文件
type DifyFile struct {
	// 文件类型
	Type string `json:"type"`
	
	// 传输方法
	TransferMethod string `json:"transfer_method"`
	
	// 文件名
	Name string `json:"name,omitempty"`
	
	// 文件URL
	URL string `json:"url,omitempty"`
	
	// 上传文件ID
	UploadFileID string `json:"upload_file_id,omitempty"`
}

// DifyMetadata Dify 元数据
type DifyMetadata struct {
	// 使用情况
	Usage *DifyUsage `json:"usage,omitempty"`
	
	// 检索结果
	RetrieverResources []DifyRetrieverResource `json:"retriever_resources,omitempty"`
	
	// 注释回复
	AnnotationReply *DifyAnnotationReply `json:"annotation_reply,omitempty"`
}

// DifyUsage Dify 使用情况
type DifyUsage struct {
	// 提示令牌数
	PromptTokens int `json:"prompt_tokens"`
	
	// 完成令牌数
	CompletionTokens int `json:"completion_tokens"`
	
	// 总令牌数
	TotalTokens int `json:"total_tokens"`
	
	// 单位价格
	UnitPrice string `json:"unit_price"`
	
	// 价格单位
	PriceUnit string `json:"price_unit"`
	
	// 总价格
	TotalPrice string `json:"total_price"`
	
	// 货币
	Currency string `json:"currency"`
	
	// 延迟（毫秒）
	Latency float64 `json:"latency"`
}

// DifyRetrieverResource Dify 检索资源
type DifyRetrieverResource struct {
	// 位置
	Position int `json:"position"`
	
	// 数据集ID
	DatasetID string `json:"dataset_id"`
	
	// 数据集名称
	DatasetName string `json:"dataset_name"`
	
	// 文档ID
	DocumentID string `json:"document_id"`
	
	// 文档名称
	DocumentName string `json:"document_name"`
	
	// 数据源类型
	DataSourceType string `json:"data_source_type"`
	
	// 段ID
	SegmentID string `json:"segment_id"`
	
	// 分数
	Score float64 `json:"score"`
	
	// 内容
	Content string `json:"content"`
}

// DifyAnnotationReply Dify 注释回复
type DifyAnnotationReply struct {
	// ID
	ID string `json:"id"`
	
	// 权威者ID
	AuthorizerID string `json:"authorizer_id"`
	
	// 创建时间
	CreatedAt time.Time `json:"created_at"`
	
	// 内容
	Content string `json:"content"`
}

// DifyModelInfo Dify 模型信息
type DifyModelInfo struct {
	// 模型名称
	Name string `json:"name"`
	
	// 提供商
	Provider string `json:"provider"`
	
	// 版本
	Version string `json:"version"`
	
	// 功能
	Features []string `json:"features"`
	
	// 参数
	Parameters map[string]interface{} `json:"parameters"`
}

// DifyAnalysisRepository Dify 分析结果仓库接口
type DifyAnalysisRepository interface {
	// SaveAnalysisResult 保存分析结果
	SaveAnalysisResult(ctx context.Context, result *DifyAnalysisResult) error
	
	// GetAnalysisResult 获取分析结果
	GetAnalysisResult(ctx context.Context, id string) (*DifyAnalysisResult, error)
	
	// GetAnalysisResultsByAlertID 根据告警ID获取分析结果
	GetAnalysisResultsByAlertID(ctx context.Context, alertID uint) ([]*DifyAnalysisResult, error)
	
	// GetAnalysisHistory 获取分析历史
	GetAnalysisHistory(ctx context.Context, filter *DifyAnalysisFilter) ([]*DifyAnalysisResult, error)
	
	// UpdateAnalysisResult 更新分析结果
	UpdateAnalysisResult(ctx context.Context, result *DifyAnalysisResult) error
	
	// DeleteAnalysisResult 删除分析结果
	DeleteAnalysisResult(ctx context.Context, id string) error
}

// DifyAnalysisResult Dify 分析结果
type DifyAnalysisResult struct {
	// ID
	ID string `json:"id"`
	
	// 告警ID
	AlertID uint `json:"alert_id"`
	
	// 分析类型
	AnalysisType string `json:"analysis_type"`
	
	// 对话ID
	ConversationID string `json:"conversation_id"`
	
	// 消息ID
	MessageID string `json:"message_id"`
	
	// 分析结果
	Result *AnalysisResult `json:"result"`
	
	// 原始响应
	RawResponse string `json:"raw_response"`
	
	// 置信度
	Confidence float64 `json:"confidence"`
	
	// 处理时间（毫秒）
	ProcessingTime int64 `json:"processing_time"`
	
	// 使用情况
	Usage *DifyUsage `json:"usage,omitempty"`
	
	// 状态
	Status string `json:"status"`
	
	// 错误信息
	Error string `json:"error,omitempty"`
	
	// 创建时间
	CreatedAt time.Time `json:"created_at"`
	
	// 更新时间
	UpdatedAt time.Time `json:"updated_at"`
}

// DifyAnalysisFilter Dify 分析过滤器
type DifyAnalysisFilter struct {
	// 告警ID列表
	AlertIDs []uint `json:"alert_ids,omitempty"`
	
	// 分析类型列表
	AnalysisTypes []string `json:"analysis_types,omitempty"`
	
	// 状态列表
	Statuses []string `json:"statuses,omitempty"`
	
	// 开始时间
	StartTime *time.Time `json:"start_time,omitempty"`
	
	// 结束时间
	EndTime *time.Time `json:"end_time,omitempty"`
	
	// 限制数量
	Limit int `json:"limit,omitempty"`
	
	// 偏移量
	Offset int `json:"offset,omitempty"`
	
	// 排序字段
	SortBy string `json:"sort_by,omitempty"`
	
	// 排序方向
	SortOrder string `json:"sort_order,omitempty"`
}