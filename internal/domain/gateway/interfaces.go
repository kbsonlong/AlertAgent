package gateway

import (
	"context"

	"alert_agent/internal/model"
)

// SmartGateway 智能告警网关接口
type SmartGateway interface {
	// ReceiveAlert 接收告警
	ReceiveAlert(ctx context.Context, alert *model.Alert) (*AlertProcessingRecord, error)
	
	// RouteAlert 路由告警
	RouteAlert(ctx context.Context, alertCtx *AlertContext) (*RoutingDecision, error)
	
	// ConvergeAlerts 收敛告警
	ConvergeAlerts(ctx context.Context, alerts []*model.Alert) (*ConvergenceResult, error)
	
	// GetProcessingRecord 获取处理记录
	GetProcessingRecord(ctx context.Context, recordID string) (*AlertProcessingRecord, error)
	
	// GetStatistics 获取统计信息
	GetStatistics(ctx context.Context, timeRange TimeRange) (*GatewayStatistics, error)
}

// AlertReceiver 告警接收器接口
type AlertReceiver interface {
	// Receive 接收告警
	Receive(ctx context.Context, alert *model.Alert) (*AlertProcessingRecord, error)
	
	// ValidateAlert 验证告警
	ValidateAlert(ctx context.Context, alert *model.Alert) error
	
	// EnrichAlert 丰富告警信息
	EnrichAlert(ctx context.Context, alert *model.Alert) (*AlertContext, error)
}

// AlertProcessor 告警处理器接口
type AlertProcessor interface {
	// Process 处理告警
	Process(ctx context.Context, alertCtx *AlertContext) (*AlertProcessingRecord, error)
	
	// GetProcessingMode 获取处理模式
	GetProcessingMode(ctx context.Context, alertCtx *AlertContext) ProcessingMode
	
	// UpdateProcessingRecord 更新处理记录
	UpdateProcessingRecord(ctx context.Context, record *AlertProcessingRecord, step ProcessingStep) error
}

// AlertRouter 告警路由器接口
type AlertRouter interface {
	// Route 路由告警
	Route(ctx context.Context, alertCtx *AlertContext) (*RoutingDecision, error)
	
	// GetAvailableChannels 获取可用渠道
	GetAvailableChannels(ctx context.Context, alertCtx *AlertContext) ([]string, error)
	
	// ValidateRouting 验证路由决策
	ValidateRouting(ctx context.Context, decision *RoutingDecision) error
}

// AlertSuppressor 告警抑制器接口
type AlertSuppressor interface {
	// ShouldSuppress 判断是否应该抑制告警
	ShouldSuppress(ctx context.Context, alertCtx *AlertContext) (bool, string, error)
	
	// AddSuppressionRule 添加抑制规则
	AddSuppressionRule(ctx context.Context, rule SuppressionRule) error
	
	// RemoveSuppressionRule 移除抑制规则
	RemoveSuppressionRule(ctx context.Context, ruleID string) error
}

// AlertConverger 告警收敛器接口
type AlertConverger interface {
	// Converge 收敛告警
	Converge(ctx context.Context, alerts []*model.Alert) (*ConvergenceResult, error)
	
	// FindSimilarAlerts 查找相似告警
	FindSimilarAlerts(ctx context.Context, alert *model.Alert) ([]*model.Alert, error)
	
	// CalculateSimilarity 计算相似度
	CalculateSimilarity(ctx context.Context, alert1, alert2 *model.Alert) (float64, error)
}

// SuppressionRule 抑制规则
type SuppressionRule struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Conditions  map[string]interface{} `json:"conditions"`
	Duration    int64                  `json:"duration"` // 抑制时长（秒）
	Enabled     bool                   `json:"enabled"`
	CreatedAt   int64                  `json:"created_at"`
	UpdatedAt   int64                  `json:"updated_at"`
}

// ProcessingStrategy 处理策略接口
type ProcessingStrategy interface {
	// CanHandle 判断是否可以处理该告警
	CanHandle(ctx context.Context, alertCtx *AlertContext) bool
	
	// Process 处理告警
	Process(ctx context.Context, alertCtx *AlertContext) (*AlertProcessingRecord, error)
	
	// GetMode 获取处理模式
	GetMode() ProcessingMode
}

// DirectPassthroughStrategy 直通策略接口
type DirectPassthroughStrategy interface {
	ProcessingStrategy
	
	// RouteDirectly 直接路由
	RouteDirectly(ctx context.Context, alertCtx *AlertContext) (*RoutingDecision, error)
}

// BasicConvergenceStrategy 基础收敛策略接口
type BasicConvergenceStrategy interface {
	ProcessingStrategy
	
	// CheckConvergence 检查收敛
	CheckConvergence(ctx context.Context, alertCtx *AlertContext) (*ConvergenceResult, error)
}

// SmartRoutingStrategy 智能路由策略接口
type SmartRoutingStrategy interface {
	ProcessingStrategy
	
	// AnalyzeAlert 分析告警
	AnalyzeAlert(ctx context.Context, alertCtx *AlertContext) (*AnalysisResult, error)
	
	// MakeRoutingDecision 做出路由决策
	MakeRoutingDecision(ctx context.Context, analysis *AnalysisResult) (*RoutingDecision, error)
}

// AnalysisResult 分析结果
type AnalysisResult struct {
	Severity    string                 `json:"severity"`
	Category    string                 `json:"category"`
	RootCause   string                 `json:"root_cause"`
	Impact      string                 `json:"impact"`
	Recommendations []string           `json:"recommendations"`
	Confidence  float64                `json:"confidence"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// FeatureToggleService 功能开关服务接口
type FeatureToggleService interface {
	// IsEnabled 检查功能是否启用
	IsEnabled(ctx context.Context, feature string) bool
	
	// GetProcessingMode 获取当前处理模式
	GetProcessingMode(ctx context.Context) ProcessingMode
	
	// CanUseSmartFeatures 是否可以使用智能功能
	CanUseSmartFeatures(ctx context.Context) bool
}

// MetricsCollector 指标收集器接口
type MetricsCollector interface {
	// RecordAlertReceived 记录告警接收
	RecordAlertReceived(ctx context.Context, alert *model.Alert)
	
	// RecordAlertProcessed 记录告警处理
	RecordAlertProcessed(ctx context.Context, record *AlertProcessingRecord)
	
	// RecordAlertRouted 记录告警路由
	RecordAlertRouted(ctx context.Context, decision *RoutingDecision)
	
	// RecordProcessingLatency 记录处理延迟
	RecordProcessingLatency(ctx context.Context, mode ProcessingMode, latency int64)
	
	// RecordError 记录错误
	RecordError(ctx context.Context, operation string, err error)
}