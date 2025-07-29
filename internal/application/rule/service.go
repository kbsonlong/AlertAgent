package rule

import (
	"context"
	"fmt"
	"strings"
	"time"

	"alert_agent/internal/domain/rule"
	"alert_agent/pkg/types"
)

// Service 规则服务实现
type Service struct {
	repo rule.Repository
}

// NewService 创建规则服务
func NewService(repo rule.Repository) rule.Service {
	return &Service{
		repo: repo,
	}
}

// CreateRule 创建规则
func (s *Service) CreateRule(ctx context.Context, r *rule.PrometheusRule) (*rule.PrometheusRule, error) {
	if err := s.validateRule(r); err != nil {
		return nil, fmt.Errorf("validate rule failed: %w", err)
	}

	// 检查规则名称是否已存在
	existing, err := s.repo.GetRuleByName(ctx, r.Name, r.ClusterID)
	if err != nil && err.Error() != "record not found" {
		return nil, fmt.Errorf("check existing rule failed: %w", err)
	}
	if existing != nil {
		return nil, fmt.Errorf("rule with name %s already exists", r.Name)
	}

	r.Enabled = true
	r.Version = 1

	if err := s.repo.CreateRule(ctx, r); err != nil {
		return nil, err
	}

	return r, nil
}

// UpdateRule 更新规则
func (s *Service) UpdateRule(ctx context.Context, id uint, r *rule.PrometheusRule) (*rule.PrometheusRule, error) {
	if err := s.validateRule(r); err != nil {
		return nil, fmt.Errorf("validate rule failed: %w", err)
	}

	// 检查规则是否存在
	existing, err := s.repo.GetRule(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get existing rule failed: %w", err)
	}

	// 检查名称冲突（排除自身）
	if r.Name != existing.Name {
		nameExists, err := s.repo.GetRuleByName(ctx, r.Name, r.ClusterID)
		if err != nil && err.Error() != "record not found" {
			return nil, fmt.Errorf("check rule name conflict failed: %w", err)
		}
		if nameExists != nil && nameExists.ID != id {
			return nil, fmt.Errorf("rule with name %s already exists", r.Name)
		}
	}

	r.ID = id
	r.CreatedAt = existing.CreatedAt
	r.Version = existing.Version + 1

	if err := s.repo.UpdateRule(ctx, r); err != nil {
		return nil, err
	}

	return r, nil
}

// DeleteRule 删除规则
func (s *Service) DeleteRule(ctx context.Context, id uint) error {
	// 检查规则是否存在
	_, err := s.repo.GetRule(ctx, id)
	if err != nil {
		return fmt.Errorf("get rule failed: %w", err)
	}

	// 检查是否有关联的分发记录
	distributions, err := s.repo.ListDistributionsByRule(ctx, id)
	if err != nil {
		return fmt.Errorf("check rule distributions failed: %w", err)
	}
	if len(distributions) > 0 {
		return fmt.Errorf("cannot delete rule with active distributions")
	}

	return s.repo.DeleteRule(ctx, id)
}

// GetRule 获取规则
func (s *Service) GetRule(ctx context.Context, id uint) (*rule.PrometheusRule, error) {
	return s.repo.GetRule(ctx, id)
}

// ListRules 列出规则
func (s *Service) ListRules(ctx context.Context, query *types.Query) (*types.PageResult, error) {
	rules, total, err := s.repo.ListRules(ctx, query)
	if err != nil {
		return nil, err
	}

	page := 1
	size := query.Limit
	if query.Offset > 0 && query.Limit > 0 {
		page = (query.Offset / query.Limit) + 1
	}

	return &types.PageResult{
		Data:  rules,
		Total: total,
		Page:  page,
		Size:  size,
	}, nil
}

// ListRulesByCluster 按集群列出规则
func (s *Service) ListRulesByCluster(ctx context.Context, clusterID string) ([]*rule.PrometheusRule, error) {
	return s.repo.ListRulesByCluster(ctx, clusterID)
}

// CreateRuleGroup 创建规则组
func (s *Service) CreateRuleGroup(ctx context.Context, group *rule.RuleGroup) (*rule.RuleGroup, error) {
	if err := s.validateRuleGroup(group); err != nil {
		return nil, fmt.Errorf("validate rule group failed: %w", err)
	}

	// 检查规则组名称是否已存在
	existing, err := s.repo.GetRuleGroupByName(ctx, group.Name, group.ClusterID)
	if err != nil && err.Error() != "record not found" {
		return nil, fmt.Errorf("check existing rule group failed: %w", err)
	}
	if existing != nil {
		return nil, fmt.Errorf("rule group with name %s already exists", group.Name)
	}

	group.Enabled = true
	group.Version = 1

	if err := s.repo.CreateRuleGroup(ctx, group); err != nil {
		return nil, err
	}

	return group, nil
}

// UpdateRuleGroup 更新规则组
func (s *Service) UpdateRuleGroup(ctx context.Context, id uint, group *rule.RuleGroup) (*rule.RuleGroup, error) {
	if err := s.validateRuleGroup(group); err != nil {
		return nil, fmt.Errorf("validate rule group failed: %w", err)
	}

	// 检查规则组是否存在
	existing, err := s.repo.GetRuleGroup(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get existing rule group failed: %w", err)
	}

	// 检查名称冲突（排除自身）
	if group.Name != existing.Name {
		nameExists, err := s.repo.GetRuleGroupByName(ctx, group.Name, group.ClusterID)
		if err != nil && err.Error() != "record not found" {
			return nil, fmt.Errorf("check rule group name conflict failed: %w", err)
		}
		if nameExists != nil && nameExists.ID != id {
			return nil, fmt.Errorf("rule group with name %s already exists", group.Name)
		}
	}

	group.ID = id
	group.CreatedAt = existing.CreatedAt
	group.Version = existing.Version + 1

	if err := s.repo.UpdateRuleGroup(ctx, group); err != nil {
		return nil, err
	}

	return group, nil
}

// DeleteRuleGroup 删除规则组
func (s *Service) DeleteRuleGroup(ctx context.Context, id uint) error {
	// 检查规则组是否存在
	_, err := s.repo.GetRuleGroup(ctx, id)
	if err != nil {
		return fmt.Errorf("get rule group failed: %w", err)
	}

	// 检查是否有关联的规则
	rules, err := s.repo.ListRulesByGroup(ctx, "", "")
	if err != nil {
		return fmt.Errorf("check group rules failed: %w", err)
	}
	if len(rules) > 0 {
		return fmt.Errorf("cannot delete rule group with existing rules")
	}

	return s.repo.DeleteRuleGroup(ctx, id)
}

// GetRuleGroup 获取规则组
func (s *Service) GetRuleGroup(ctx context.Context, id uint) (*rule.RuleGroup, error) {
	return s.repo.GetRuleGroup(ctx, id)
}

// ListRuleGroups 列出规则组
func (s *Service) ListRuleGroups(ctx context.Context, query *types.Query) (*types.PageResult, error) {
	groups, total, err := s.repo.ListRuleGroups(ctx, query)
	if err != nil {
		return nil, err
	}

	page := 1
	size := query.Limit
	if query.Offset > 0 && query.Limit > 0 {
		page = (query.Offset / query.Limit) + 1
	}

	return &types.PageResult{
		Data:  groups,
		Total: total,
		Page:  page,
		Size:  size,
	}, nil
}

// ListRuleGroupsByCluster 按集群列出规则组
func (s *Service) ListRuleGroupsByCluster(ctx context.Context, clusterID string) ([]*rule.RuleGroup, error) {
	return s.repo.ListRuleGroupsByCluster(ctx, clusterID)
}

// ValidateRule 验证规则
func (s *Service) ValidateRule(ctx context.Context, r *rule.PrometheusRule) error {
	return s.validateRule(r)
}

// ValidateRuleGroup 验证规则组
func (s *Service) ValidateRuleGroup(ctx context.Context, group *rule.RuleGroup) error {
	return s.validateRuleGroup(group)
}

// ValidateRuleExpression 验证规则表达式
func (s *Service) ValidateRuleExpression(ctx context.Context, expression string) error {
	if expression == "" {
		return fmt.Errorf("expression is required")
	}
	// 简单的Prometheus表达式验证
	if !strings.Contains(expression, "{") || !strings.Contains(expression, "}") {
		return fmt.Errorf("invalid prometheus expression format")
	}
	return nil
}

// DistributeRule 分发规则
func (s *Service) DistributeRule(ctx context.Context, ruleID uint, clusterIDs []string) error {
	// 检查规则是否存在
	_, err := s.repo.GetRule(ctx, ruleID)
	if err != nil {
		return fmt.Errorf("get rule failed: %w", err)
	}

	// 创建分发记录
	var distributions []*rule.RuleDistribution
	for _, clusterID := range clusterIDs {
		distribution := &rule.RuleDistribution{
			RuleID:    ruleID,
			ClusterID: clusterID,
			Version:   1,
			Status:    "pending",
		}
		distributions = append(distributions, distribution)
	}

	return s.repo.BatchCreateDistributions(ctx, distributions)
}

// DistributeRuleGroup 分发规则组
func (s *Service) DistributeRuleGroup(ctx context.Context, groupID uint, clusterIDs []string) error {
	// 检查规则组是否存在
	_, err := s.repo.GetRuleGroup(ctx, groupID)
	if err != nil {
		return fmt.Errorf("get rule group failed: %w", err)
	}

	// 获取规则组下的所有规则
	rules, err := s.repo.ListRulesByGroup(ctx, "", "")
	if err != nil {
		return fmt.Errorf("get group rules failed: %w", err)
	}

	// 为每个规则创建分发记录
	var distributions []*rule.RuleDistribution
	for _, r := range rules {
		for _, clusterID := range clusterIDs {
			distribution := &rule.RuleDistribution{
				RuleID:    r.ID,
				ClusterID: clusterID,
				Version:   r.Version,
				Status:    "pending",
			}
			distributions = append(distributions, distribution)
		}
	}

	return s.repo.BatchCreateDistributions(ctx, distributions)
}

// RetryFailedDistributions 重试失败的分发
func (s *Service) RetryFailedDistributions(ctx context.Context, clusterID string) error {
	failedDistributions, err := s.repo.ListFailedDistributions(ctx)
	if err != nil {
		return fmt.Errorf("get failed distributions failed: %w", err)
	}

	// 更新状态为pending以便重试
	for _, dist := range failedDistributions {
		if dist.ClusterID == clusterID {
			dist.Status = "pending"
			dist.RetryCount++
			now := time.Now()
			dist.LastRetryAt = &now
			if err := s.repo.UpdateDistribution(ctx, dist); err != nil {
				return fmt.Errorf("update distribution failed: %w", err)
			}
		}
	}

	return nil
}

// GetDistributionStatus 获取分发状态
func (s *Service) GetDistributionStatus(ctx context.Context, ruleID uint, clusterID string) (*rule.RuleDistribution, error) {
	distributions, err := s.repo.ListDistributionsByRule(ctx, ruleID)
	if err != nil {
		return nil, fmt.Errorf("get distributions failed: %w", err)
	}

	for _, dist := range distributions {
		if dist.ClusterID == clusterID {
			return dist, nil
		}
	}

	return nil, fmt.Errorf("distribution not found")
}

// ListDistributions 列出分发记录
func (s *Service) ListDistributions(ctx context.Context, query *types.Query) (*types.PageResult, error) {
	distributions, total, err := s.repo.ListDistributions(ctx, query)
	if err != nil {
		return nil, err
	}

	page := 1
	size := query.Limit
	if query.Offset > 0 && query.Limit > 0 {
		page = (query.Offset / query.Limit) + 1
	}

	return &types.PageResult{
		Data:  distributions,
		Total: total,
		Page:  page,
		Size:  size,
	}, nil
}

// DetectConflicts 检测冲突
func (s *Service) DetectConflicts(ctx context.Context, clusterID string) ([]*rule.RuleConflict, error) {
	// 获取集群的所有规则
	rules, err := s.repo.ListRulesByCluster(ctx, clusterID)
	if err != nil {
		return nil, fmt.Errorf("get cluster rules failed: %w", err)
	}

	var conflicts []*rule.RuleConflict
	nameMap := make(map[string][]*rule.PrometheusRule)

	// 按名称分组
	for _, r := range rules {
		nameMap[r.Name] = append(nameMap[r.Name], r)
	}

	// 检测名称冲突
	for name, ruleList := range nameMap {
		if len(ruleList) > 1 {
			for i := 1; i < len(ruleList); i++ {
				conflict := &rule.RuleConflict{
					RuleID1:     ruleList[0].ID,
					RuleID2:     ruleList[i].ID,
					ClusterID:   clusterID,
					ConflictType: rule.ConflictTypeName,
					Description: fmt.Sprintf("Rule name '%s' conflicts between rules %d and %d", name, ruleList[0].ID, ruleList[i].ID),
					Resolved:    false,
				}
				conflicts = append(conflicts, conflict)
			}
		}
	}

	return conflicts, nil
}

// ResolveConflict 解决冲突
func (s *Service) ResolveConflict(ctx context.Context, conflictID uint, resolvedBy string) error {
	// 获取冲突记录
	conflict, err := s.repo.GetConflict(ctx, conflictID)
	if err != nil {
		return fmt.Errorf("get conflict failed: %w", err)
	}

	// 更新冲突状态
	conflict.Resolved = true
	conflict.ResolvedBy = resolvedBy
	now := time.Now()
	conflict.ResolvedAt = &now

	return s.repo.UpdateConflict(ctx, conflict)
}

// ListConflicts 列出冲突
func (s *Service) ListConflicts(ctx context.Context, query *types.Query) (*types.PageResult, error) {
	conflicts, total, err := s.repo.ListConflicts(ctx, query)
	if err != nil {
		return nil, err
	}

	page := 1
	size := query.Limit
	if query.Offset > 0 && query.Limit > 0 {
		page = (query.Offset / query.Limit) + 1
	}

	return &types.PageResult{
		Data:  conflicts,
		Total: total,
		Page:  page,
		Size:  size,
	}, nil
}

// CreateRuleVersion 创建规则版本
func (s *Service) CreateRuleVersion(ctx context.Context, ruleID uint, changeLog string, createdBy string) (*rule.RuleVersion, error) {
	// 检查规则是否存在
	ruleObj, err := s.repo.GetRule(ctx, ruleID)
	if err != nil {
		return nil, fmt.Errorf("get rule failed: %w", err)
	}

	// 获取最新版本号
	latestVersion, err := s.repo.GetLatestVersion(ctx, ruleID)
	if err != nil && err.Error() != "record not found" {
		return nil, fmt.Errorf("get latest version failed: %w", err)
	}

	newVersionNumber := int64(1)
	if latestVersion != nil {
		newVersionNumber = latestVersion.Version + 1
	}

	version := &rule.RuleVersion{
		RuleID:    ruleID,
		Version:   newVersionNumber,
		Content:   fmt.Sprintf("%+v", ruleObj), // 简化的内容序列化
		Checksum:  fmt.Sprintf("%d", time.Now().Unix()), // 简化的校验和
		ChangeLog: changeLog,
		CreatedBy: createdBy,
	}

	if err := s.repo.CreateVersion(ctx, version); err != nil {
		return nil, err
	}

	return version, nil
}

// GetRuleVersions 获取规则版本
func (s *Service) GetRuleVersions(ctx context.Context, ruleID uint) ([]*rule.RuleVersion, error) {
	return s.repo.ListVersionsByRule(ctx, ruleID)
}

// RollbackRule 回滚规则
func (s *Service) RollbackRule(ctx context.Context, ruleID uint, version int64) error {
	// 获取指定版本
	versions, err := s.repo.ListVersionsByRule(ctx, ruleID)
	if err != nil {
		return fmt.Errorf("get rule versions failed: %w", err)
	}

	var targetVersion *rule.RuleVersion
	for _, v := range versions {
		if v.Version == version {
			targetVersion = v
			break
		}
	}

	if targetVersion == nil {
		return fmt.Errorf("version %d not found", version)
	}

	// 这里应该实现实际的回滚逻辑
	// 目前只是创建一个新版本记录
	_, err = s.CreateRuleVersion(ctx, ruleID, fmt.Sprintf("Rollback to version %d", version), "system")
	return err
}

// CompareRuleVersions 比较规则版本
func (s *Service) CompareRuleVersions(ctx context.Context, ruleID uint, version1, version2 int64) (*rule.VersionComparison, error) {
	versions, err := s.repo.ListVersionsByRule(ctx, ruleID)
	if err != nil {
		return nil, fmt.Errorf("get rule versions failed: %w", err)
	}

	var v1, v2 *rule.RuleVersion
	for _, v := range versions {
		if v.Version == version1 {
			v1 = v
		}
		if v.Version == version2 {
			v2 = v
		}
	}

	if v1 == nil {
		return nil, fmt.Errorf("version %d not found", version1)
	}
	if v2 == nil {
		return nil, fmt.Errorf("version %d not found", version2)
	}

	comparison := &rule.VersionComparison{
		RuleID:   ruleID,
		Version1: version1,
		Version2: version2,
		Changes:  []rule.Change{},
	}

	// 简单的内容比较
	if v1.Content != v2.Content {
		comparison.Changes = append(comparison.Changes, rule.Change{
			Field:    "content",
			OldValue: v1.Content,
			NewValue: v2.Content,
			Type:     "modified",
		})
	}

	return comparison, nil
}

// SyncRulesToCluster 同步规则到集群
func (s *Service) SyncRulesToCluster(ctx context.Context, clusterID string) error {
	// 获取待同步的分发记录
	pendingDistributions, err := s.repo.ListPendingDistributions(ctx)
	if err != nil {
		return fmt.Errorf("get pending distributions failed: %w", err)
	}

	// 过滤指定集群的分发记录
	for _, dist := range pendingDistributions {
		if dist.ClusterID == clusterID {
			// 这里应该实现实际的同步逻辑
			// 目前只是更新状态
			dist.Status = "success"
			now := time.Now()
			dist.DistributedAt = &now
			if err := s.repo.UpdateDistribution(ctx, dist); err != nil {
				return fmt.Errorf("update distribution failed: %w", err)
			}
		}
	}

	return nil
}

// GetSyncStatus 获取同步状态
func (s *Service) GetSyncStatus(ctx context.Context, clusterID string) (*rule.SyncStatus, error) {
	// 获取集群的分发统计
	stats, err := s.repo.CountDistributionsByStatus(ctx, clusterID)
	if err != nil {
		return nil, fmt.Errorf("get distribution stats failed: %w", err)
	}

	totalCount := int64(0)
	successCount := stats["success"]
	failedCount := stats["failed"]
	pendingCount := stats["pending"]

	for _, count := range stats {
		totalCount += count
	}

	status := "success"
	if failedCount > 0 {
		status = "failed"
	} else if pendingCount > 0 {
		status = "in_progress"
	}

	return &rule.SyncStatus{
		ClusterID:     clusterID,
		LastSyncTime:  time.Now().Format(time.RFC3339),
		SyncStatus:    status,
		RuleCount:     totalCount,
		SuccessCount:  successCount,
		FailedCount:   failedCount,
		PendingCount:  pendingCount,
	}, nil
}

// GeneratePrometheusConfig 生成Prometheus配置
func (s *Service) GeneratePrometheusConfig(ctx context.Context, clusterID string) (string, error) {
	// 获取集群的所有规则组
	groups, err := s.repo.ListRuleGroupsByCluster(ctx, clusterID)
	if err != nil {
		return "", fmt.Errorf("get rule groups failed: %w", err)
	}

	// 简化的配置生成
	config := "groups:\n"
	for _, group := range groups {
		config += fmt.Sprintf("- name: %s\n", group.Name)
		config += fmt.Sprintf("  interval: %s\n", group.Interval)
		config += "  rules:\n"

		// 获取组内规则
		rules, err := s.repo.ListRulesByGroup(ctx, group.Name, clusterID)
		if err != nil {
			continue
		}

		for _, r := range rules {
			config += fmt.Sprintf("  - alert: %s\n", r.Name)
			config += fmt.Sprintf("    expr: %s\n", r.Expression)
			config += fmt.Sprintf("    for: %s\n", r.Duration)
			config += "    annotations:\n"
			config += fmt.Sprintf("      summary: %s\n", r.Summary)
			config += fmt.Sprintf("      description: %s\n", r.Description)
		}
	}

	return config, nil
}

// GetRuleStats 获取规则统计
func (s *Service) GetRuleStats(ctx context.Context, clusterID string) (*rule.RuleStats, error) {
	// 获取规则总数
	totalRules, err := s.repo.CountRulesByCluster(ctx, clusterID)
	if err != nil {
		return nil, fmt.Errorf("count rules failed: %w", err)
	}

	// 获取按状态分组的统计
	statsByStatus, err := s.repo.CountRulesByStatus(ctx, clusterID)
	if err != nil {
		return nil, fmt.Errorf("count rules by status failed: %w", err)
	}

	return &rule.RuleStats{
		ClusterID:       clusterID,
		TotalRules:      totalRules,
		EnabledRules:    statsByStatus["enabled"],
		DisabledRules:   statsByStatus["disabled"],
		RulesByGroup:    make(map[string]int64),
		RulesBySeverity: make(map[string]int64),
	}, nil
}

// GetDistributionStats 获取分发统计
func (s *Service) GetDistributionStats(ctx context.Context, clusterID string) (*rule.DistributionStats, error) {
	// 获取按状态分组的分发统计
	stats, err := s.repo.CountDistributionsByStatus(ctx, clusterID)
	if err != nil {
		return nil, fmt.Errorf("count distributions by status failed: %w", err)
	}

	totalDistributions := int64(0)
	for _, count := range stats {
		totalDistributions += count
	}

	return &rule.DistributionStats{
		ClusterID:             clusterID,
		TotalDistributions:    totalDistributions,
		SuccessCount:          stats["success"],
		FailedCount:           stats["failed"],
		PendingCount:          stats["pending"],
		RetryCount:            0, // 需要额外查询
		DistributionsByStatus: stats,
	}, nil
}

// GetConflictStats 获取冲突统计
func (s *Service) GetConflictStats(ctx context.Context, clusterID string) (*rule.ConflictStats, error) {
	// 获取按类型分组的冲突统计
	stats, err := s.repo.CountConflictsByType(ctx, clusterID)
	if err != nil {
		return nil, fmt.Errorf("count conflicts by type failed: %w", err)
	}

	totalConflicts := int64(0)
	for _, count := range stats {
		totalConflicts += count
	}

	// 获取未解决的冲突
	unresolvedConflicts, err := s.repo.ListUnresolvedConflicts(ctx)
	if err != nil {
		return nil, fmt.Errorf("get unresolved conflicts failed: %w", err)
	}

	unresolvedCount := int64(0)
	for _, conflict := range unresolvedConflicts {
		if conflict.ClusterID == clusterID {
			unresolvedCount++
		}
	}

	return &rule.ConflictStats{
		ClusterID:           clusterID,
		TotalConflicts:      totalConflicts,
		ResolvedConflicts:   totalConflicts - unresolvedCount,
		UnresolvedConflicts: unresolvedCount,
		ConflictsByType:     stats,
	}, nil
}

// validateRule 验证规则
func (s *Service) validateRule(r *rule.PrometheusRule) error {
	if r.Name == "" {
		return fmt.Errorf("rule name is required")
	}
	if r.Expression == "" {
		return fmt.Errorf("rule expression is required")
	}
	if r.Severity == "" {
		return fmt.Errorf("rule severity is required")
	}
	if r.ClusterID == "" {
		return fmt.Errorf("cluster ID is required")
	}

	// 验证严重级别
	validSeverities := []string{rule.SeverityCritical, rule.SeverityWarning, rule.SeverityInfo}
	valid := false
	for _, severity := range validSeverities {
		if r.Severity == severity {
			valid = true
			break
		}
	}
	if !valid {
		return fmt.Errorf("invalid severity: %s", r.Severity)
	}

	return nil
}

// validateRuleGroup 验证规则组
func (s *Service) validateRuleGroup(group *rule.RuleGroup) error {
	if group.Name == "" {
		return fmt.Errorf("rule group name is required")
	}
	if group.ClusterID == "" {
		return fmt.Errorf("cluster ID is required")
	}
	if group.Interval == "" {
		group.Interval = "30s" // 设置默认值
	}

	return nil
}