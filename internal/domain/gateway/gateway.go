package gateway

import (
	"context"
	"time"
)

// Alert 告警
type Alert struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Severity    string            `json:"severity"`
	Status      string            `json:"status"`
	StartsAt    time.Time         `json:"starts_at"`
	EndsAt      *time.Time        `json:"ends_at"`
	Labels      map[string]string `json:"labels"`
	Annotations map[string]string `json:"annotations"`
	GeneratorURL string           `json:"generator_url"`
	Fingerprint string            `json:"fingerprint"`
}

// ProcessingResult 处理结果
type ProcessingResult struct {
	AlertID         string                 `json:"alert_id"`
	Action          ProcessingAction       `json:"action"`
	ChannelsSent    []string               `json:"channels_sent"`
	ConvergenceKey  string                 `json:"convergence_key,omitempty"`
	SuppressionRule string                 `json:"suppression_rule,omitempty"`
	RoutingDecision *RoutingDecision       `json:"routing_decision,omitempty"`
	ProcessingTime  int                    `json:"processing_time"` // 毫秒
	Metadata        map[string]interface{} `json:"metadata"`
}

// ProcessingAction 处理动作
type ProcessingAction string

const (
	ActionSent       ProcessingAction = "sent"
	ActionConverged  ProcessingAction = "converged"
	ActionSuppressed ProcessingAction = "suppressed"
	ActionRouted     ProcessingAction = "routed"
	ActionFailed     ProcessingAction = "failed"
)

// RoutingDecision 路由决策
type RoutingDecision struct {
	RuleID      string   `json:"rule_id"`
	RuleName    string   `json:"rule_name"`
	ChannelIDs  []string `json:"channel_ids"`
	Priority    int      `json:"priority"`
	Confidence  float64  `json:"confidence"`
	Reason      string   `json:"reason"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// ConvergenceWindow 收敛窗口
type ConvergenceWindow struct {
	Key                 string    `json:"key"`
	WindowStart         time.Time `json:"window_start"`
	WindowEnd           time.Time `json:"window_end"`
	AlertCount          int       `json:"alert_count"`
	MaxAlerts           int       `json:"max_alerts"`
	TriggerCondition    string    `json:"trigger_condition"`
	RepresentativeAlert *Alert    `json:"representative_alert"`
	Alerts              []*Alert  `json:"alerts"`
}

// SuppressionRule 抑制规则
type SuppressionRule struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Enabled     bool                   `json:"enabled"`
	Priority    int                    `json:"priority"`
	Conditions  map[string]interface{} `json:"conditions"`
	Schedule    map[string]interface{} `json:"schedule"`
	CreatedBy   string                 `json:"created_by"`
	UpdatedBy   string                 `json:"updated_by"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

// RoutingRule 路由规则
type RoutingRule struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Enabled     bool                   `json:"enabled"`
	Priority    int                    `json:"priority"`
	Conditions  map[string]interface{} `json:"conditions"`
	Actions     map[string]interface{} `json:"actions"`
	ChannelIDs  []string               `json:"channel_ids"`
	CreatedBy   string                 `json:"created_by"`
	UpdatedBy   string                 `json:"updated_by"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

// GatewayRepository 网关仓储接口
type GatewayRepository interface {
	// 告警处理记录
	CreateProcessingRecord(ctx context.Context, record *ProcessingRecord) error
	UpdateProcessingRecord(ctx context.Context, record *ProcessingRecord) error
	GetProcessingRecord(ctx context.Context, alertID string) (*ProcessingRecord, error)
	ListProcessingRecords(ctx context.Context, query *ProcessingQuery) ([]*ProcessingRecord, int64, error)

	// 收敛记录
	CreateConvergenceRecord(ctx context.Context, record *ConvergenceRecord) error
	UpdateConvergenceRecord(ctx context.Context, record *ConvergenceRecord) error
	GetConvergenceWindow(ctx context.Context, key string) (*ConvergenceWindow, error)
	SetConvergenceWindow(ctx context.Context, window *ConvergenceWindow, ttl time.Duration) error

	// 抑制规则
	CreateSuppressionRule(ctx context.Context, rule *SuppressionRule) error
	UpdateSuppressionRule(ctx context.Context, rule *SuppressionRule) error
	DeleteSuppressionRule(ctx context.Context, id string) error
	GetSuppressionRule(ctx context.Context, id string) (*SuppressionRule, error)
	ListSuppressionRules(ctx context.Context, enabled *bool) ([]*SuppressionRule, error)

	// 路由规则
	CreateRoutingRule(ctx context.Context, rule *RoutingRule) error
	UpdateRoutingRule(ctx context.Context, rule *RoutingRule) error
	DeleteRoutingRule(ctx context.Context, id string) error
	GetRoutingRule(ctx context.Context, id string) (*RoutingRule, error)
	ListRoutingRules(ctx context.Context, enabled *bool) ([]*RoutingRule, error)
}

// ProcessingRecord 处理记录
type ProcessingRecord struct {
	ID               string                 `json:"id"`
	AlertID          string                 `json:"alert_id"`
	AlertName        string                 `json:"alert_name"`
	Severity         string                 `json:"severity"`
	ClusterID        string                 `json:"cluster_id"`
	ReceivedAt       time.Time              `json:"received_at"`
	ProcessedAt      *time.Time             `json:"processed_at"`
	ProcessingStatus string                 `json:"processing_status"`
	AnalysisID       string                 `json:"analysis_id"`
	Decision         map[string]interface{} `json:"decision"`
	ActionTaken      string                 `json:"action_taken"`
	ResolutionTime   int                    `json:"resolution_time"`
	FeedbackScore    float64                `json:"feedback_score"`
	Labels           map[string]string      `json:"labels"`
	Annotations      map[string]string      `json:"annotations"`
	ChannelsSent     []string               `json:"channels_sent"`
	ErrorMessage     string                 `json:"error_message"`
	CreatedAt        time.Time              `json:"created_at"`
	UpdatedAt        time.Time              `json:"updated_at"`
}

// ConvergenceRecord 收敛记录
type ConvergenceRecord struct {
	ID                  string                 `json:"id"`
	ConvergenceKey      string                 `json:"convergence_key"`
	WindowStart         time.Time              `json:"window_start"`
	WindowEnd           time.Time              `json:"window_end"`
	AlertCount          int                    `json:"alert_count"`
	Status              string                 `json:"status"`
	TriggerCondition    map[string]interface{} `json:"trigger_condition"`
	RepresentativeAlert map[string]interface{} `json:"representative_alert"`
	ConvergedAlerts     []string               `json:"converged_alerts"`
	TriggeredAt         *time.Time             `json:"triggered_at"`
	CreatedAt           time.Time              `json:"created_at"`
	UpdatedAt           time.Time              `json:"updated_at"`
}

// ProcessingQuery 处理查询
type ProcessingQuery struct {
	AlertName string    `json:"alert_name"`
	Severity  string    `json:"severity"`
	ClusterID string    `json:"cluster_id"`
	Status    string    `json:"status"`
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
	Page      int       `json:"page"`
	PageSize  int       `json:"page_size"`
	SortBy    string    `json:"sort_by"`
	SortDesc  bool      `json:"sort_desc"`
}