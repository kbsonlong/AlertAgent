package gateway

import (
	"context"
	"time"

	"alert_agent/internal/model"
)

// AlertStatus 告警状态
type AlertStatus string

const (
	AlertStatusReceived   AlertStatus = "received"   // 已接收
	AlertStatusProcessing AlertStatus = "processing" // 处理中
	AlertStatusRouted     AlertStatus = "routed"     // 已路由
	AlertStatusSuppressed AlertStatus = "suppressed" // 已抑制
	AlertStatusConverged  AlertStatus = "converged"  // 已收敛
	AlertStatusFailed     AlertStatus = "failed"     // 处理失败
)

// ProcessingMode 处理模式
type ProcessingMode string

const (
	ModeDirectPassthrough ProcessingMode = "direct_passthrough" // 直通模式
	ModeBasicConvergence  ProcessingMode = "basic_convergence"   // 基础收敛
	ModeSmartRouting      ProcessingMode = "smart_routing"       // 智能路由
)

// AlertProcessingRecord 告警处理记录
type AlertProcessingRecord struct {
	ID              string                 `json:"id" gorm:"primaryKey;type:varchar(64)"`
	AlertID         uint                   `json:"alert_id" gorm:"not null;index"`
	OriginalAlert   *model.Alert           `json:"original_alert" gorm:"-"`
	ProcessingMode  ProcessingMode         `json:"processing_mode" gorm:"type:varchar(32);not null"`
	Status          AlertStatus            `json:"status" gorm:"type:varchar(32);not null"`
	ReceivedAt      time.Time              `json:"received_at" gorm:"not null"`
	ProcessedAt     *time.Time             `json:"processed_at,omitempty"`
	RoutedAt        *time.Time             `json:"routed_at,omitempty"`
	ProcessingSteps []ProcessingStep       `json:"processing_steps" gorm:"serializer:json"`
	Metadata        map[string]interface{} `json:"metadata" gorm:"serializer:json"`
	ErrorMessage    string                 `json:"error_message,omitempty" gorm:"type:text"`
	CreatedAt       time.Time              `json:"created_at"`
	UpdatedAt       time.Time              `json:"updated_at"`
}

// ProcessingStep 处理步骤
type ProcessingStep struct {
	Step        string                 `json:"step"`
	Status      string                 `json:"status"`
	StartTime   time.Time              `json:"start_time"`
	EndTime     *time.Time             `json:"end_time,omitempty"`
	Duration    time.Duration          `json:"duration"`
	Details     map[string]interface{} `json:"details,omitempty"`
	ErrorMsg    string                 `json:"error_msg,omitempty"`
}

// RoutingDecision 路由决策
type RoutingDecision struct {
	ChannelIDs   []string               `json:"channel_ids"`
	Priority     int                    `json:"priority"`
	Delay        time.Duration          `json:"delay"`
	Suppressed   bool                   `json:"suppressed"`
	Reason       string                 `json:"reason"`
	Confidence   float64                `json:"confidence"`
	Metadata     map[string]interface{} `json:"metadata"`
	DecisionTime time.Time              `json:"decision_time"`
}

// ConvergenceResult 收敛结果
type ConvergenceResult struct {
	Converged       bool                   `json:"converged"`
	GroupID         string                 `json:"group_id"`
	Representative  *model.Alert           `json:"representative"`
	SimilarAlerts   []*model.Alert         `json:"similar_alerts"`
	ConvergenceRule string                 `json:"convergence_rule"`
	Metadata        map[string]interface{} `json:"metadata"`
}

// AlertContext 告警上下文
type AlertContext struct {
	Alert           *model.Alert           `json:"alert"`
	ClusterID       string                 `json:"cluster_id"`
	Environment     string                 `json:"environment"`
	Team            string                 `json:"team"`
	HistoricalData  []HistoricalAlert      `json:"historical_data"`
	RelatedMetrics  map[string]interface{} `json:"related_metrics"`
	ProcessingHints map[string]interface{} `json:"processing_hints"`
}

// HistoricalAlert 历史告警
type HistoricalAlert struct {
	Alert      *model.Alert `json:"alert"`
	Similarity float64      `json:"similarity"`
	Timestamp  time.Time    `json:"timestamp"`
}

// GatewayStatistics 网关统计信息
type GatewayStatistics struct {
	TotalReceived     int64                      `json:"total_received"`
	TotalProcessed    int64                      `json:"total_processed"`
	TotalRouted       int64                      `json:"total_routed"`
	TotalSuppressed   int64                      `json:"total_suppressed"`
	TotalConverged    int64                      `json:"total_converged"`
	TotalFailed       int64                      `json:"total_failed"`
	AverageLatency    time.Duration              `json:"average_latency"`
	ProcessingModes   map[ProcessingMode]int64   `json:"processing_modes"`
	StatusDistribution map[AlertStatus]int64     `json:"status_distribution"`
	LastUpdated       time.Time                  `json:"last_updated"`
}

// Repository interfaces

// AlertProcessingRepository 告警处理记录仓储接口
type AlertProcessingRepository interface {
	Create(ctx context.Context, record *AlertProcessingRecord) error
	Update(ctx context.Context, record *AlertProcessingRecord) error
	GetByID(ctx context.Context, id string) (*AlertProcessingRecord, error)
	GetByAlertID(ctx context.Context, alertID uint) (*AlertProcessingRecord, error)
	List(ctx context.Context, filter AlertProcessingFilter) ([]*AlertProcessingRecord, error)
	GetStatistics(ctx context.Context, timeRange TimeRange) (*GatewayStatistics, error)
}

// AlertProcessingFilter 告警处理记录过滤器
type AlertProcessingFilter struct {
	Status         []AlertStatus    `json:"status,omitempty"`
	ProcessingMode []ProcessingMode `json:"processing_mode,omitempty"`
	TimeRange      *TimeRange       `json:"time_range,omitempty"`
	Limit          int              `json:"limit,omitempty"`
	Offset         int              `json:"offset,omitempty"`
}

// TimeRange 时间范围
type TimeRange struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}