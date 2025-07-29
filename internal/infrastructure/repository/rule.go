package repository

import (
	"context"
	"alert_agent/internal/domain/rule"
	"alert_agent/pkg/types"
	"gorm.io/gorm"
)

// ruleRepository 规则仓储实现
type ruleRepository struct {
	db *gorm.DB
}

// NewRuleRepository 创建规则仓储实例
func NewRuleRepository(db *gorm.DB) rule.Repository {
	return &ruleRepository{
		db: db,
	}
}

// CreateRule 创建规则
func (r *ruleRepository) CreateRule(ctx context.Context, ruleEntity *rule.PrometheusRule) error {
	return r.db.WithContext(ctx).Create(ruleEntity).Error
}

// UpdateRule 更新规则
func (r *ruleRepository) UpdateRule(ctx context.Context, ruleEntity *rule.PrometheusRule) error {
	return r.db.WithContext(ctx).Save(ruleEntity).Error
}

// DeleteRule 删除规则
func (r *ruleRepository) DeleteRule(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&rule.PrometheusRule{}, id).Error
}

// GetRule 获取规则
func (r *ruleRepository) GetRule(ctx context.Context, id uint) (*rule.PrometheusRule, error) {
	var ruleEntity rule.PrometheusRule
	err := r.db.WithContext(ctx).First(&ruleEntity, id).Error
	if err != nil {
		return nil, err
	}
	return &ruleEntity, nil
}

// GetRuleByName 根据名称和集群ID获取规则
func (r *ruleRepository) GetRuleByName(ctx context.Context, name, clusterID string) (*rule.PrometheusRule, error) {
	var ruleEntity rule.PrometheusRule
	err := r.db.WithContext(ctx).Where("name = ? AND cluster_id = ?", name, clusterID).First(&ruleEntity).Error
	if err != nil {
		return nil, err
	}
	return &ruleEntity, nil
}

// ListRules 列出规则
func (r *ruleRepository) ListRules(ctx context.Context, query *types.Query) ([]*rule.PrometheusRule, int64, error) {
	var rules []*rule.PrometheusRule
	var total int64

	db := r.db.WithContext(ctx).Model(&rule.PrometheusRule{})

	// 应用搜索条件
	if query.Search != "" {
		db = db.Where("name LIKE ? OR description LIKE ?", "%"+query.Search+"%", "%"+query.Search+"%")
	}

	// 应用过滤条件
	if query.Filter != nil {
		for key, value := range query.Filter {
			db = db.Where(key+" = ?", value)
		}
	}

	// 计算总数
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 应用排序
	if query.OrderBy != "" {
		db = db.Order(query.OrderBy)
	} else {
		db = db.Order("created_at DESC")
	}

	// 应用分页
	if query.Limit > 0 {
		db = db.Limit(query.Limit)
	}
	if query.Offset > 0 {
		db = db.Offset(query.Offset)
	}

	err := db.Find(&rules).Error
	return rules, total, err
}

// ListRulesByCluster 根据集群ID列出规则
func (r *ruleRepository) ListRulesByCluster(ctx context.Context, clusterID string) ([]*rule.PrometheusRule, error) {
	var rules []*rule.PrometheusRule
	err := r.db.WithContext(ctx).Where("cluster_id = ?", clusterID).Find(&rules).Error
	return rules, err
}

// ListRulesByGroup 根据规则组列出规则
func (r *ruleRepository) ListRulesByGroup(ctx context.Context, groupName, clusterID string) ([]*rule.PrometheusRule, error) {
	var rules []*rule.PrometheusRule
	err := r.db.WithContext(ctx).Where("group_name = ? AND cluster_id = ?", groupName, clusterID).Find(&rules).Error
	return rules, err
}

// CreateRuleGroup 创建规则组
func (r *ruleRepository) CreateRuleGroup(ctx context.Context, group *rule.RuleGroup) error {
	return r.db.WithContext(ctx).Create(group).Error
}

// UpdateRuleGroup 更新规则组
func (r *ruleRepository) UpdateRuleGroup(ctx context.Context, group *rule.RuleGroup) error {
	return r.db.WithContext(ctx).Save(group).Error
}

// DeleteRuleGroup 删除规则组
func (r *ruleRepository) DeleteRuleGroup(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&rule.RuleGroup{}, id).Error
}

// GetRuleGroup 获取规则组
func (r *ruleRepository) GetRuleGroup(ctx context.Context, id uint) (*rule.RuleGroup, error) {
	var group rule.RuleGroup
	err := r.db.WithContext(ctx).Preload("Rules").First(&group, id).Error
	if err != nil {
		return nil, err
	}
	return &group, nil
}

// GetRuleGroupByName 根据名称获取规则组
func (r *ruleRepository) GetRuleGroupByName(ctx context.Context, name, clusterID string) (*rule.RuleGroup, error) {
	var group rule.RuleGroup
	err := r.db.WithContext(ctx).Preload("Rules").Where("name = ? AND cluster_id = ?", name, clusterID).First(&group).Error
	if err != nil {
		return nil, err
	}
	return &group, nil
}

// ListRuleGroups 列出规则组
func (r *ruleRepository) ListRuleGroups(ctx context.Context, query *types.Query) ([]*rule.RuleGroup, int64, error) {
	var groups []*rule.RuleGroup
	var total int64

	db := r.db.WithContext(ctx).Model(&rule.RuleGroup{})

	// 应用搜索条件
	if query.Search != "" {
		db = db.Where("name LIKE ?", "%"+query.Search+"%")
	}

	// 应用过滤条件
	if query.Filter != nil {
		for key, value := range query.Filter {
			db = db.Where(key+" = ?", value)
		}
	}

	// 计算总数
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 应用排序
	if query.OrderBy != "" {
		db = db.Order(query.OrderBy)
	} else {
		db = db.Order("created_at DESC")
	}

	// 应用分页
	if query.Limit > 0 {
		db = db.Limit(query.Limit)
	}
	if query.Offset > 0 {
		db = db.Offset(query.Offset)
	}

	err := db.Preload("Rules").Find(&groups).Error
	return groups, total, err
}

// ListRuleGroupsByCluster 根据集群ID列出规则组
func (r *ruleRepository) ListRuleGroupsByCluster(ctx context.Context, clusterID string) ([]*rule.RuleGroup, error) {
	var groups []*rule.RuleGroup
	err := r.db.WithContext(ctx).Preload("Rules").Where("cluster_id = ?", clusterID).Find(&groups).Error
	return groups, err
}

// CreateDistribution 创建分发记录
func (r *ruleRepository) CreateDistribution(ctx context.Context, distribution *rule.RuleDistribution) error {
	return r.db.WithContext(ctx).Create(distribution).Error
}

// UpdateDistribution 更新分发记录
func (r *ruleRepository) UpdateDistribution(ctx context.Context, distribution *rule.RuleDistribution) error {
	return r.db.WithContext(ctx).Save(distribution).Error
}

// GetDistribution 获取分发记录
func (r *ruleRepository) GetDistribution(ctx context.Context, id uint) (*rule.RuleDistribution, error) {
	var distribution rule.RuleDistribution
	err := r.db.WithContext(ctx).First(&distribution, id).Error
	if err != nil {
		return nil, err
	}
	return &distribution, nil
}

// ListDistributions 列出分发记录
func (r *ruleRepository) ListDistributions(ctx context.Context, query *types.Query) ([]*rule.RuleDistribution, int64, error) {
	var distributions []*rule.RuleDistribution
	var total int64

	db := r.db.WithContext(ctx).Model(&rule.RuleDistribution{})

	// 应用搜索条件
	if query.Search != "" {
		db = db.Where("cluster_id LIKE ? OR status LIKE ?", "%"+query.Search+"%", "%"+query.Search+"%")
	}

	// 应用过滤条件
	if query.Filter != nil {
		for key, value := range query.Filter {
			db = db.Where(key+" = ?", value)
		}
	}

	// 计算总数
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 应用排序
	if query.OrderBy != "" {
		db = db.Order(query.OrderBy)
	} else {
		db = db.Order("created_at DESC")
	}

	// 应用分页
	if query.Limit > 0 {
		db = db.Limit(query.Limit)
	}
	if query.Offset > 0 {
		db = db.Offset(query.Offset)
	}

	err := db.Find(&distributions).Error
	return distributions, total, err
}

// ListDistributionsByCluster 根据集群ID列出分发记录
func (r *ruleRepository) ListDistributionsByCluster(ctx context.Context, clusterID string) ([]*rule.RuleDistribution, error) {
	var distributions []*rule.RuleDistribution
	err := r.db.WithContext(ctx).Where("cluster_id = ?", clusterID).Find(&distributions).Error
	return distributions, err
}

// ListDistributionsByRule 根据规则ID列出分发记录
func (r *ruleRepository) ListDistributionsByRule(ctx context.Context, ruleID uint) ([]*rule.RuleDistribution, error) {
	var distributions []*rule.RuleDistribution
	err := r.db.WithContext(ctx).Where("rule_id = ?", ruleID).Find(&distributions).Error
	return distributions, err
}

// ListPendingDistributions 列出待分发的记录
func (r *ruleRepository) ListPendingDistributions(ctx context.Context) ([]*rule.RuleDistribution, error) {
	var distributions []*rule.RuleDistribution
	err := r.db.WithContext(ctx).Where("status = ?", rule.RuleStatusPending).Find(&distributions).Error
	return distributions, err
}

// ListFailedDistributions 列出分发失败的记录
func (r *ruleRepository) ListFailedDistributions(ctx context.Context) ([]*rule.RuleDistribution, error) {
	var distributions []*rule.RuleDistribution
	err := r.db.WithContext(ctx).Where("status = ?", rule.RuleStatusFailed).Find(&distributions).Error
	return distributions, err
}

// CreateConflict 创建冲突记录
func (r *ruleRepository) CreateConflict(ctx context.Context, conflict *rule.RuleConflict) error {
	return r.db.WithContext(ctx).Create(conflict).Error
}

// UpdateConflict 更新冲突记录
func (r *ruleRepository) UpdateConflict(ctx context.Context, conflict *rule.RuleConflict) error {
	return r.db.WithContext(ctx).Save(conflict).Error
}

// GetConflict 获取冲突记录
func (r *ruleRepository) GetConflict(ctx context.Context, id uint) (*rule.RuleConflict, error) {
	var conflict rule.RuleConflict
	err := r.db.WithContext(ctx).First(&conflict, id).Error
	if err != nil {
		return nil, err
	}
	return &conflict, nil
}

// ListConflicts 列出冲突记录
func (r *ruleRepository) ListConflicts(ctx context.Context, query *types.Query) ([]*rule.RuleConflict, int64, error) {
	var conflicts []*rule.RuleConflict
	var total int64

	db := r.db.WithContext(ctx).Model(&rule.RuleConflict{})

	// 应用搜索条件
	if query.Search != "" {
		db = db.Where("cluster_id LIKE ? OR conflict_type LIKE ?", "%"+query.Search+"%", "%"+query.Search+"%")
	}

	// 应用过滤条件
	if query.Filter != nil {
		for key, value := range query.Filter {
			db = db.Where(key+" = ?", value)
		}
	}

	// 计算总数
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 应用排序
	if query.OrderBy != "" {
		db = db.Order(query.OrderBy)
	} else {
		db = db.Order("created_at DESC")
	}

	// 应用分页
	if query.Limit > 0 {
		db = db.Limit(query.Limit)
	}
	if query.Offset > 0 {
		db = db.Offset(query.Offset)
	}

	err := db.Find(&conflicts).Error
	return conflicts, total, err
}

// ListConflictsByCluster 根据集群ID列出冲突记录
func (r *ruleRepository) ListConflictsByCluster(ctx context.Context, clusterID string) ([]*rule.RuleConflict, error) {
	var conflicts []*rule.RuleConflict
	err := r.db.WithContext(ctx).Where("cluster_id = ?", clusterID).Find(&conflicts).Error
	return conflicts, err
}

// ListUnresolvedConflicts 列出未解决的冲突
func (r *ruleRepository) ListUnresolvedConflicts(ctx context.Context) ([]*rule.RuleConflict, error) {
	var conflicts []*rule.RuleConflict
	err := r.db.WithContext(ctx).Where("resolved = ?", false).Find(&conflicts).Error
	return conflicts, err
}

// DeleteConflict 删除冲突记录
func (r *ruleRepository) DeleteConflict(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&rule.RuleConflict{}, id).Error
}

// CreateVersion 创建版本记录
func (r *ruleRepository) CreateVersion(ctx context.Context, version *rule.RuleVersion) error {
	return r.db.WithContext(ctx).Create(version).Error
}

// GetVersion 获取版本记录
func (r *ruleRepository) GetVersion(ctx context.Context, id uint) (*rule.RuleVersion, error) {
	var version rule.RuleVersion
	err := r.db.WithContext(ctx).First(&version, id).Error
	if err != nil {
		return nil, err
	}
	return &version, nil
}

// ListVersionsByRule 根据规则ID列出版本记录
func (r *ruleRepository) ListVersionsByRule(ctx context.Context, ruleID uint) ([]*rule.RuleVersion, error) {
	var versions []*rule.RuleVersion
	err := r.db.WithContext(ctx).Where("rule_id = ?", ruleID).Order("version DESC").Find(&versions).Error
	return versions, err
}

// GetLatestVersion 获取最新版本
func (r *ruleRepository) GetLatestVersion(ctx context.Context, ruleID uint) (*rule.RuleVersion, error) {
	var version rule.RuleVersion
	err := r.db.WithContext(ctx).Where("rule_id = ?", ruleID).Order("version DESC").First(&version).Error
	if err != nil {
		return nil, err
	}
	return &version, nil
}

// DeleteOldVersions 删除旧版本
func (r *ruleRepository) DeleteOldVersions(ctx context.Context, ruleID uint, keepCount int) error {
	// 获取要保留的版本ID
	var versionIDs []uint
	err := r.db.WithContext(ctx).Model(&rule.RuleVersion{}).
		Where("rule_id = ?", ruleID).
		Order("version DESC").
		Limit(keepCount).
		Pluck("id", &versionIDs).Error
	if err != nil {
		return err
	}

	// 删除不在保留列表中的版本
	if len(versionIDs) > 0 {
		return r.db.WithContext(ctx).Where("rule_id = ? AND id NOT IN ?", ruleID, versionIDs).Delete(&rule.RuleVersion{}).Error
	}
	return nil
}

// BatchCreateRules 批量创建规则
func (r *ruleRepository) BatchCreateRules(ctx context.Context, rules []*rule.PrometheusRule) error {
	return r.db.WithContext(ctx).CreateInBatches(rules, 100).Error
}

// BatchUpdateRules 批量更新规则
func (r *ruleRepository) BatchUpdateRules(ctx context.Context, rules []*rule.PrometheusRule) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, ruleEntity := range rules {
			if err := tx.Save(ruleEntity).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// BatchDeleteRules 批量删除规则
func (r *ruleRepository) BatchDeleteRules(ctx context.Context, ids []uint) error {
	return r.db.WithContext(ctx).Delete(&rule.PrometheusRule{}, ids).Error
}

// BatchCreateDistributions 批量创建分发记录
func (r *ruleRepository) BatchCreateDistributions(ctx context.Context, distributions []*rule.RuleDistribution) error {
	return r.db.WithContext(ctx).CreateInBatches(distributions, 100).Error
}

// CountRulesByCluster 统计集群规则数量
func (r *ruleRepository) CountRulesByCluster(ctx context.Context, clusterID string) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&rule.PrometheusRule{}).Where("cluster_id = ?", clusterID).Count(&count).Error
	return count, err
}

// CountRulesByStatus 按状态统计规则数量
func (r *ruleRepository) CountRulesByStatus(ctx context.Context, clusterID string) (map[string]int64, error) {
	type StatusCount struct {
		Enabled bool  `json:"enabled"`
		Count   int64 `json:"count"`
	}

	var results []StatusCount
	err := r.db.WithContext(ctx).Model(&rule.PrometheusRule{}).
		Select("enabled, COUNT(*) as count").
		Where("cluster_id = ?", clusterID).
		Group("enabled").
		Find(&results).Error

	if err != nil {
		return nil, err
	}

	stats := make(map[string]int64)
	for _, result := range results {
		if result.Enabled {
			stats["enabled"] = result.Count
		} else {
			stats["disabled"] = result.Count
		}
	}

	return stats, nil
}

// CountDistributionsByStatus 按状态统计分发数量
func (r *ruleRepository) CountDistributionsByStatus(ctx context.Context, clusterID string) (map[string]int64, error) {
	type StatusCount struct {
		Status string `json:"status"`
		Count  int64  `json:"count"`
	}

	var results []StatusCount
	err := r.db.WithContext(ctx).Model(&rule.RuleDistribution{}).
		Select("status, COUNT(*) as count").
		Where("cluster_id = ?", clusterID).
		Group("status").
		Find(&results).Error

	if err != nil {
		return nil, err
	}

	stats := make(map[string]int64)
	for _, result := range results {
		stats[result.Status] = result.Count
	}

	return stats, nil
}

// CountConflictsByType 按类型统计冲突数量
func (r *ruleRepository) CountConflictsByType(ctx context.Context, clusterID string) (map[string]int64, error) {
	type TypeCount struct {
		ConflictType string `json:"conflict_type"`
		Count        int64  `json:"count"`
	}

	var results []TypeCount
	err := r.db.WithContext(ctx).Model(&rule.RuleConflict{}).
		Select("conflict_type, COUNT(*) as count").
		Where("cluster_id = ? AND resolved = ?", clusterID, false).
		Group("conflict_type").
		Find(&results).Error

	if err != nil {
		return nil, err
	}

	stats := make(map[string]int64)
	for _, result := range results {
		stats[result.ConflictType] = result.Count
	}

	return stats, nil
}