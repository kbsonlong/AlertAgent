package rule

import (
	"context"
	"alert_agent/pkg/types"
)

// Service 规则服务接口
type Service interface {
	// 规则管理
	CreateRule(ctx context.Context, rule *PrometheusRule) (*PrometheusRule, error)
	UpdateRule(ctx context.Context, id uint, rule *PrometheusRule) (*PrometheusRule, error)
	DeleteRule(ctx context.Context, id uint) error
	GetRule(ctx context.Context, id uint) (*PrometheusRule, error)
	ListRules(ctx context.Context, query *types.Query) (*types.PageResult, error)
	ListRulesByCluster(ctx context.Context, clusterID string) ([]*PrometheusRule, error)

	// 规则组管理
	CreateRuleGroup(ctx context.Context, group *RuleGroup) (*RuleGroup, error)
	UpdateRuleGroup(ctx context.Context, id uint, group *RuleGroup) (*RuleGroup, error)
	DeleteRuleGroup(ctx context.Context, id uint) error
	GetRuleGroup(ctx context.Context, id uint) (*RuleGroup, error)
	ListRuleGroups(ctx context.Context, query *types.Query) (*types.PageResult, error)
	ListRuleGroupsByCluster(ctx context.Context, clusterID string) ([]*RuleGroup, error)

	// 规则验证
	ValidateRule(ctx context.Context, rule *PrometheusRule) error
	ValidateRuleGroup(ctx context.Context, group *RuleGroup) error
	ValidateRuleExpression(ctx context.Context, expression string) error

	// 规则分发
	DistributeRule(ctx context.Context, ruleID uint, clusterIDs []string) error
	DistributeRuleGroup(ctx context.Context, groupID uint, clusterIDs []string) error
	RetryFailedDistributions(ctx context.Context, clusterID string) error
	GetDistributionStatus(ctx context.Context, ruleID uint, clusterID string) (*RuleDistribution, error)
	ListDistributions(ctx context.Context, query *types.Query) (*types.PageResult, error)

	// 规则冲突检测
	DetectConflicts(ctx context.Context, clusterID string) ([]*RuleConflict, error)
	ResolveConflict(ctx context.Context, conflictID uint, resolvedBy string) error
	ListConflicts(ctx context.Context, query *types.Query) (*types.PageResult, error)

	// 规则版本控制
	CreateRuleVersion(ctx context.Context, ruleID uint, changeLog string, createdBy string) (*RuleVersion, error)
	GetRuleVersions(ctx context.Context, ruleID uint) ([]*RuleVersion, error)
	RollbackRule(ctx context.Context, ruleID uint, version int64) error
	CompareRuleVersions(ctx context.Context, ruleID uint, version1, version2 int64) (*VersionComparison, error)

	// 规则同步
	SyncRulesToCluster(ctx context.Context, clusterID string) error
	GetSyncStatus(ctx context.Context, clusterID string) (*SyncStatus, error)
	GeneratePrometheusConfig(ctx context.Context, clusterID string) (string, error)

	// 规则统计
	GetRuleStats(ctx context.Context, clusterID string) (*RuleStats, error)
	GetDistributionStats(ctx context.Context, clusterID string) (*DistributionStats, error)
	GetConflictStats(ctx context.Context, clusterID string) (*ConflictStats, error)
}

// VersionComparison 版本比较结果
type VersionComparison struct {
	RuleID   uint   `json:"rule_id"`
	Version1 int64  `json:"version_1"`
	Version2 int64  `json:"version_2"`
	Changes  []Change `json:"changes"`
}

// Change 变更记录
type Change struct {
	Field    string `json:"field"`
	OldValue string `json:"old_value"`
	NewValue string `json:"new_value"`
	Type     string `json:"type"` // added, modified, deleted
}

// SyncStatus 同步状态
type SyncStatus struct {
	ClusterID     string `json:"cluster_id"`
	LastSyncTime  string `json:"last_sync_time"`
	SyncStatus    string `json:"sync_status"` // success, failed, in_progress
	RuleCount     int64  `json:"rule_count"`
	SuccessCount  int64  `json:"success_count"`
	FailedCount   int64  `json:"failed_count"`
	PendingCount  int64  `json:"pending_count"`
	LastError     string `json:"last_error,omitempty"`
}

// RuleStats 规则统计
type RuleStats struct {
	ClusterID     string            `json:"cluster_id"`
	TotalRules    int64             `json:"total_rules"`
	EnabledRules  int64             `json:"enabled_rules"`
	DisabledRules int64             `json:"disabled_rules"`
	RulesByGroup  map[string]int64  `json:"rules_by_group"`
	RulesBySeverity map[string]int64 `json:"rules_by_severity"`
}

// DistributionStats 分发统计
type DistributionStats struct {
	ClusterID        string            `json:"cluster_id"`
	TotalDistributions int64           `json:"total_distributions"`
	SuccessCount     int64             `json:"success_count"`
	FailedCount      int64             `json:"failed_count"`
	PendingCount     int64             `json:"pending_count"`
	RetryCount       int64             `json:"retry_count"`
	DistributionsByStatus map[string]int64 `json:"distributions_by_status"`
}

// ConflictStats 冲突统计
type ConflictStats struct {
	ClusterID         string            `json:"cluster_id"`
	TotalConflicts    int64             `json:"total_conflicts"`
	ResolvedConflicts int64             `json:"resolved_conflicts"`
	UnresolvedConflicts int64           `json:"unresolved_conflicts"`
	ConflictsByType   map[string]int64  `json:"conflicts_by_type"`
}