package database

import (
	"time"

	"gorm.io/gorm"
)

// Channel 告警渠道
type Channel struct {
	ID          string                 `json:"id" gorm:"primarykey;size:36"`
	Name        string                 `json:"name" gorm:"size:100;not null;index"`
	Type        string                 `json:"type" gorm:"size:50;not null;index"`
	Description string                 `json:"description" gorm:"type:text"`
	Config      map[string]interface{} `json:"config" gorm:"type:json"`
	GroupID     string                 `json:"group_id" gorm:"size:36;index"`
	Tags        []string               `json:"tags" gorm:"type:json"`
	Status      string                 `json:"status" gorm:"type:varchar(20);default:'active';index"`
	HealthStatus string                `json:"health_status" gorm:"type:varchar(20);default:'unknown'"`
	LastHealthCheck *time.Time         `json:"last_health_check"`
	ResponseTime    int                `json:"response_time"` // 毫秒
	SuccessRate     float64            `json:"success_rate" gorm:"type:decimal(5,2);default:0"`
	CreatedBy       string             `json:"created_by" gorm:"size:100"`
	UpdatedBy       string             `json:"updated_by" gorm:"size:100"`
	CreatedAt       time.Time          `json:"created_at"`
	UpdatedAt       time.Time          `json:"updated_at"`
	DeletedAt       gorm.DeletedAt     `json:"deleted_at" gorm:"index"`

	// 关联
	Group    *ChannelGroup `json:"group,omitempty" gorm:"foreignKey:GroupID"`
}

// ChannelGroup 渠道分组
type ChannelGroup struct {
	ID          string         `json:"id" gorm:"primarykey;size:36"`
	Name        string         `json:"name" gorm:"size:100;not null;uniqueIndex"`
	Description string         `json:"description" gorm:"type:text"`
	ParentID    string         `json:"parent_id" gorm:"size:36;index"`
	Path        string         `json:"path" gorm:"size:500;index"` // 层级路径，如 /root/group1/subgroup1
	Level       int            `json:"level" gorm:"default:0"`
	SortOrder   int            `json:"sort_order" gorm:"default:0"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"deleted_at" gorm:"index"`

	// 关联
	Parent   *ChannelGroup   `json:"parent,omitempty" gorm:"foreignKey:ParentID"`
	Children []ChannelGroup  `json:"children,omitempty" gorm:"foreignKey:ParentID"`
	Channels []Channel       `json:"channels,omitempty" gorm:"foreignKey:GroupID"`
}

// Cluster Alertmanager集群
type Cluster struct {
	ID                  string            `json:"id" gorm:"primarykey;size:36"`
	Name                string            `json:"name" gorm:"size:100;not null;uniqueIndex"`
	Endpoint            string            `json:"endpoint" gorm:"size:255;not null"`
	ConfigPath          string            `json:"config_path" gorm:"size:255"`
	RulesPath           string            `json:"rules_path" gorm:"size:255"`
	SyncInterval        int               `json:"sync_interval" gorm:"default:30"`
	HealthCheckInterval int               `json:"health_check_interval" gorm:"default:10"`
	Status              string            `json:"status" gorm:"type:varchar(20);default:'active';index"`
	HealthStatus        string            `json:"health_status" gorm:"type:varchar(20);default:'unknown'"`
	LastHealthCheck     *time.Time        `json:"last_health_check"`
	LastSyncTime        *time.Time        `json:"last_sync_time"`
	SyncStatus          string            `json:"sync_status" gorm:"type:varchar(20);default:'pending'"`
	Labels              map[string]string `json:"labels" gorm:"type:json"`
	Metadata            map[string]interface{} `json:"metadata" gorm:"type:json"`
	CreatedAt           time.Time         `json:"created_at"`
	UpdatedAt           time.Time         `json:"updated_at"`
	DeletedAt           gorm.DeletedAt    `json:"deleted_at" gorm:"index"`

	// 关联
	SyncRecords []ConfigSyncRecord `json:"sync_records,omitempty" gorm:"foreignKey:ClusterID"`
}

// AlertProcessingRecord 告警处理记录
type AlertProcessingRecord struct {
	ID               string                 `json:"id" gorm:"primarykey;size:36"`
	AlertID          string                 `json:"alert_id" gorm:"size:100;not null;index"`
	AlertName        string                 `json:"alert_name" gorm:"size:100;index"`
	Severity         string                 `json:"severity" gorm:"size:20;index"`
	ClusterID        string                 `json:"cluster_id" gorm:"size:36;index"`
	ReceivedAt       time.Time              `json:"received_at" gorm:"index"`
	ProcessedAt      *time.Time             `json:"processed_at"`
	ProcessingStatus string                 `json:"processing_status" gorm:"type:varchar(20);default:'received';index"`
	AnalysisID       string                 `json:"analysis_id" gorm:"size:36;index"`
	Decision         map[string]interface{} `json:"decision" gorm:"type:json"`
	ActionTaken      string                 `json:"action_taken" gorm:"size:100"`
	ResolutionTime   int                    `json:"resolution_time"` // 秒
	FeedbackScore    float64                `json:"feedback_score" gorm:"type:decimal(3,2)"`
	Labels           map[string]string      `json:"labels" gorm:"type:json"`
	Annotations      map[string]string      `json:"annotations" gorm:"type:json"`
	ChannelsSent     []string               `json:"channels_sent" gorm:"type:json"`
	ErrorMessage     string                 `json:"error_message" gorm:"type:text"`
	CreatedAt        time.Time              `json:"created_at"`
	UpdatedAt        time.Time              `json:"updated_at"`

	// 关联
	Cluster         *Cluster          `json:"cluster,omitempty" gorm:"foreignKey:ClusterID"`
	AnalysisRecord  *AIAnalysisRecord `json:"analysis_record,omitempty" gorm:"foreignKey:AnalysisID"`
}

// AIAnalysisRecord AI分析记录
type AIAnalysisRecord struct {
	ID              string                 `json:"id" gorm:"primarykey;size:36"`
	AlertID         string                 `json:"alert_id" gorm:"size:100;not null;index"`
	AnalysisType    string                 `json:"analysis_type" gorm:"size:50;default:'root_cause_analysis';index"`
	RequestData     map[string]interface{} `json:"request_data" gorm:"type:json"`
	ResponseData    map[string]interface{} `json:"response_data" gorm:"type:json"`
	AnalysisResult  map[string]interface{} `json:"analysis_result" gorm:"type:json"`
	ConfidenceScore float64                `json:"confidence_score" gorm:"type:decimal(3,2)"`
	ProcessingTime  int                    `json:"processing_time"` // 毫秒
	Status          string                 `json:"status" gorm:"type:varchar(20);default:'pending';index"`
	ErrorMessage    string                 `json:"error_message" gorm:"type:text"`
	Provider        string                 `json:"provider" gorm:"size:50;index"` // dify, n8n, ollama
	ModelVersion    string                 `json:"model_version" gorm:"size:100"`
	CreatedAt       time.Time              `json:"created_at"`
	UpdatedAt       time.Time              `json:"updated_at"`

	// 关联
	ProcessingRecords []AlertProcessingRecord `json:"processing_records,omitempty" gorm:"foreignKey:AnalysisID"`
}

// ConvergenceRecord 告警收敛记录
type ConvergenceRecord struct {
	ID                  string                 `json:"id" gorm:"primarykey;size:36"`
	ConvergenceKey      string                 `json:"convergence_key" gorm:"size:255;not null;index"`
	WindowStart         time.Time              `json:"window_start" gorm:"index"`
	WindowEnd           time.Time              `json:"window_end" gorm:"index"`
	AlertCount          int                    `json:"alert_count" gorm:"default:0"`
	Status              string                 `json:"status" gorm:"type:varchar(20);default:'active';index"`
	TriggerCondition    map[string]interface{} `json:"trigger_condition" gorm:"type:json"`
	RepresentativeAlert map[string]interface{} `json:"representative_alert" gorm:"type:json"`
	ConvergedAlerts     []string               `json:"converged_alerts" gorm:"type:json"`
	TriggeredAt         *time.Time             `json:"triggered_at"`
	CreatedAt           time.Time              `json:"created_at"`
	UpdatedAt           time.Time              `json:"updated_at"`
}

// SuppressionRule 告警抑制规则
type SuppressionRule struct {
	ID          string                 `json:"id" gorm:"primarykey;size:36"`
	Name        string                 `json:"name" gorm:"size:100;not null;index"`
	Description string                 `json:"description" gorm:"type:text"`
	Enabled     bool                   `json:"enabled" gorm:"default:true;index"`
	Priority    int                    `json:"priority" gorm:"default:0;index"`
	Conditions  map[string]interface{} `json:"conditions" gorm:"type:json"`
	Schedule    map[string]interface{} `json:"schedule" gorm:"type:json"` // 时间窗口配置
	CreatedBy   string                 `json:"created_by" gorm:"size:100"`
	UpdatedBy   string                 `json:"updated_by" gorm:"size:100"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
	DeletedAt   gorm.DeletedAt         `json:"deleted_at" gorm:"index"`
}

// RoutingRule 路由规则
type RoutingRule struct {
	ID          string                 `json:"id" gorm:"primarykey;size:36"`
	Name        string                 `json:"name" gorm:"size:100;not null;index"`
	Description string                 `json:"description" gorm:"type:text"`
	Enabled     bool                   `json:"enabled" gorm:"default:true;index"`
	Priority    int                    `json:"priority" gorm:"default:0;index"`
	Conditions  map[string]interface{} `json:"conditions" gorm:"type:json"`
	Actions     map[string]interface{} `json:"actions" gorm:"type:json"`
	ChannelIDs  []string               `json:"channel_ids" gorm:"type:json"`
	CreatedBy   string                 `json:"created_by" gorm:"size:100"`
	UpdatedBy   string                 `json:"updated_by" gorm:"size:100"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
	DeletedAt   gorm.DeletedAt         `json:"deleted_at" gorm:"index"`
}

// ConfigSyncRecord 配置同步记录
type ConfigSyncRecord struct {
	ID           string                 `json:"id" gorm:"primarykey;size:36"`
	ClusterID    string                 `json:"cluster_id" gorm:"size:36;not null;index"`
	ConfigType   string                 `json:"config_type" gorm:"size:50;not null;index"` // alertmanager, prometheus_rules
	ConfigHash   string                 `json:"config_hash" gorm:"size:64;index"`
	SyncStatus   string                 `json:"sync_status" gorm:"type:varchar(20);default:'pending';index"`
	StartTime    time.Time              `json:"start_time"`
	EndTime      *time.Time             `json:"end_time"`
	Duration     int                    `json:"duration"` // 毫秒
	ErrorMessage string                 `json:"error_message" gorm:"type:text"`
	ConfigData   map[string]interface{} `json:"config_data" gorm:"type:json"`
	CreatedAt    time.Time              `json:"created_at"`

	// 关联
	Cluster *Cluster `json:"cluster,omitempty" gorm:"foreignKey:ClusterID"`
}

// AuditLog 审计日志
type AuditLog struct {
	ID         string                 `json:"id" gorm:"primarykey;size:36"`
	UserID     string                 `json:"user_id" gorm:"size:100;index"`
	UserName   string                 `json:"user_name" gorm:"size:100"`
	Action     string                 `json:"action" gorm:"size:100;not null;index"`
	Resource   string                 `json:"resource" gorm:"size:100;not null;index"`
	ResourceID string                 `json:"resource_id" gorm:"size:100;index"`
	Method     string                 `json:"method" gorm:"size:10;index"`
	Path       string                 `json:"path" gorm:"size:255"`
	IP         string                 `json:"ip" gorm:"size:45;index"`
	UserAgent  string                 `json:"user_agent" gorm:"size:500"`
	Status     int                    `json:"status" gorm:"index"`
	Duration   int                    `json:"duration"` // 毫秒
	Request    map[string]interface{} `json:"request" gorm:"type:json"`
	Response   map[string]interface{} `json:"response" gorm:"type:json"`
	Error      string                 `json:"error" gorm:"type:text"`
	CreatedAt  time.Time              `json:"created_at" gorm:"index"`
}

// TableName 方法用于指定表名
func (Channel) TableName() string              { return "channels" }
func (ChannelGroup) TableName() string         { return "channel_groups" }
func (Cluster) TableName() string              { return "clusters" }
func (AlertProcessingRecord) TableName() string { return "alert_processing_records" }
func (AIAnalysisRecord) TableName() string     { return "ai_analysis_records" }
func (ConvergenceRecord) TableName() string    { return "convergence_records" }
func (SuppressionRule) TableName() string      { return "suppression_rules" }
func (RoutingRule) TableName() string          { return "routing_rules" }
func (ConfigSyncRecord) TableName() string     { return "config_sync_records" }
func (AuditLog) TableName() string             { return "audit_logs" }