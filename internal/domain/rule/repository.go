package rule

import (
	"context"
	"alert_agent/pkg/types"
)

// Repository 规则仓储接口
type Repository interface {
	// 规则管理
	CreateRule(ctx context.Context, rule *PrometheusRule) error
	UpdateRule(ctx context.Context, rule *PrometheusRule) error
	DeleteRule(ctx context.Context, id uint) error
	GetRule(ctx context.Context, id uint) (*PrometheusRule, error)
	GetRuleByName(ctx context.Context, name, clusterID string) (*PrometheusRule, error)
	ListRules(ctx context.Context, query *types.Query) ([]*PrometheusRule, int64, error)
	ListRulesByCluster(ctx context.Context, clusterID string) ([]*PrometheusRule, error)
	ListRulesByGroup(ctx context.Context, groupName, clusterID string) ([]*PrometheusRule, error)

	// 规则组管理
	CreateRuleGroup(ctx context.Context, group *RuleGroup) error
	UpdateRuleGroup(ctx context.Context, group *RuleGroup) error
	DeleteRuleGroup(ctx context.Context, id uint) error
	GetRuleGroup(ctx context.Context, id uint) (*RuleGroup, error)
	GetRuleGroupByName(ctx context.Context, name, clusterID string) (*RuleGroup, error)
	ListRuleGroups(ctx context.Context, query *types.Query) ([]*RuleGroup, int64, error)
	ListRuleGroupsByCluster(ctx context.Context, clusterID string) ([]*RuleGroup, error)

	// 规则分发管理
	CreateDistribution(ctx context.Context, distribution *RuleDistribution) error
	UpdateDistribution(ctx context.Context, distribution *RuleDistribution) error
	GetDistribution(ctx context.Context, id uint) (*RuleDistribution, error)
	ListDistributions(ctx context.Context, query *types.Query) ([]*RuleDistribution, int64, error)
	ListDistributionsByCluster(ctx context.Context, clusterID string) ([]*RuleDistribution, error)
	ListDistributionsByRule(ctx context.Context, ruleID uint) ([]*RuleDistribution, error)
	ListPendingDistributions(ctx context.Context) ([]*RuleDistribution, error)
	ListFailedDistributions(ctx context.Context) ([]*RuleDistribution, error)

	// 规则冲突管理
	CreateConflict(ctx context.Context, conflict *RuleConflict) error
	UpdateConflict(ctx context.Context, conflict *RuleConflict) error
	GetConflict(ctx context.Context, id uint) (*RuleConflict, error)
	ListConflicts(ctx context.Context, query *types.Query) ([]*RuleConflict, int64, error)
	ListConflictsByCluster(ctx context.Context, clusterID string) ([]*RuleConflict, error)
	ListUnresolvedConflicts(ctx context.Context) ([]*RuleConflict, error)
	DeleteConflict(ctx context.Context, id uint) error

	// 规则版本管理
	CreateVersion(ctx context.Context, version *RuleVersion) error
	GetVersion(ctx context.Context, id uint) (*RuleVersion, error)
	ListVersionsByRule(ctx context.Context, ruleID uint) ([]*RuleVersion, error)
	GetLatestVersion(ctx context.Context, ruleID uint) (*RuleVersion, error)
	DeleteOldVersions(ctx context.Context, ruleID uint, keepCount int) error

	// 批量操作
	BatchCreateRules(ctx context.Context, rules []*PrometheusRule) error
	BatchUpdateRules(ctx context.Context, rules []*PrometheusRule) error
	BatchDeleteRules(ctx context.Context, ids []uint) error
	BatchCreateDistributions(ctx context.Context, distributions []*RuleDistribution) error

	// 统计查询
	CountRulesByCluster(ctx context.Context, clusterID string) (int64, error)
	CountRulesByStatus(ctx context.Context, clusterID string) (map[string]int64, error)
	CountDistributionsByStatus(ctx context.Context, clusterID string) (map[string]int64, error)
	CountConflictsByType(ctx context.Context, clusterID string) (map[string]int64, error)
}