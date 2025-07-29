package rule

import (
	"time"
	"gorm.io/gorm"
)

// PrometheusRule Prometheus告警规则实体
type PrometheusRule struct {
	gorm.Model
	Name        string    `json:"name" gorm:"type:varchar(255);not null;uniqueIndex:idx_rule_name_cluster"`
	ClusterID   string    `json:"cluster_id" gorm:"type:varchar(100);not null;uniqueIndex:idx_rule_name_cluster"`
	GroupName   string    `json:"group_name" gorm:"type:varchar(255);not null"`
	Expression  string    `json:"expression" gorm:"type:text;not null"`
	Duration    string    `json:"duration" gorm:"type:varchar(50);default:'5m'"`
	Severity    string    `json:"severity" gorm:"type:varchar(50);not null"`
	Summary     string    `json:"summary" gorm:"type:text"`
	Description string    `json:"description" gorm:"type:text"`
	Labels      string    `json:"labels" gorm:"type:json"` // JSON格式存储标签
	Annotations string    `json:"annotations" gorm:"type:json"` // JSON格式存储注解
	Enabled     bool      `json:"enabled" gorm:"default:true"`
	Version     int64     `json:"version" gorm:"default:1"` // 版本号
	Checksum    string    `json:"checksum" gorm:"type:varchar(64)"` // 规则内容校验和
	CreatedBy   string    `json:"created_by" gorm:"type:varchar(100)"`
	UpdatedBy   string    `json:"updated_by" gorm:"type:varchar(100)"`
}

// RuleGroup 规则组
type RuleGroup struct {
	gorm.Model
	Name      string    `json:"name" gorm:"type:varchar(255);not null;uniqueIndex:idx_group_name_cluster"`
	ClusterID string    `json:"cluster_id" gorm:"type:varchar(100);not null;uniqueIndex:idx_group_name_cluster"`
	Interval  string    `json:"interval" gorm:"type:varchar(50);default:'30s'"`
	Enabled   bool      `json:"enabled" gorm:"default:true"`
	Version   int64     `json:"version" gorm:"default:1"`
	Checksum  string    `json:"checksum" gorm:"type:varchar(64)"`
	CreatedBy string    `json:"created_by" gorm:"type:varchar(100)"`
	UpdatedBy string    `json:"updated_by" gorm:"type:varchar(100)"`
	Rules     []PrometheusRule `json:"rules" gorm:"foreignKey:GroupName;references:Name"`
}

// RuleDistribution 规则分发记录
type RuleDistribution struct {
	gorm.Model
	RuleID      uint      `json:"rule_id" gorm:"not null"`
	ClusterID   string    `json:"cluster_id" gorm:"type:varchar(100);not null"`
	Version     int64     `json:"version" gorm:"not null"`
	Status      string    `json:"status" gorm:"type:varchar(50);not null"` // pending, success, failed
	Error       string    `json:"error" gorm:"type:text"`
	DistributedAt *time.Time `json:"distributed_at"`
	RetryCount  int       `json:"retry_count" gorm:"default:0"`
	LastRetryAt *time.Time `json:"last_retry_at"`
}

// RuleConflict 规则冲突记录
type RuleConflict struct {
	gorm.Model
	RuleID1     uint   `json:"rule_id_1" gorm:"not null"`
	RuleID2     uint   `json:"rule_id_2" gorm:"not null"`
	ClusterID   string `json:"cluster_id" gorm:"type:varchar(100);not null"`
	ConflictType string `json:"conflict_type" gorm:"type:varchar(50);not null"` // name, expression, label
	Description string `json:"description" gorm:"type:text"`
	Resolved    bool   `json:"resolved" gorm:"default:false"`
	ResolvedBy  string `json:"resolved_by" gorm:"type:varchar(100)"`
	ResolvedAt  *time.Time `json:"resolved_at"`
}

// RuleVersion 规则版本历史
type RuleVersion struct {
	gorm.Model
	RuleID      uint   `json:"rule_id" gorm:"not null"`
	Version     int64  `json:"version" gorm:"not null"`
	Content     string `json:"content" gorm:"type:longtext;not null"` // 完整的规则内容
	Checksum    string `json:"checksum" gorm:"type:varchar(64);not null"`
	ChangeLog   string `json:"change_log" gorm:"type:text"`
	CreatedBy   string `json:"created_by" gorm:"type:varchar(100)"`
}

// 规则状态常量
const (
	RuleStatusPending = "pending"
	RuleStatusSuccess = "success"
	RuleStatusFailed  = "failed"
)

// 冲突类型常量
const (
	ConflictTypeName       = "name"
	ConflictTypeExpression = "expression"
	ConflictTypeLabel      = "label"
)

// 严重级别常量
const (
	SeverityCritical = "critical"
	SeverityWarning  = "warning"
	SeverityInfo     = "info"
)