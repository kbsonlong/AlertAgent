package repository

import (
	"context"
	"fmt"
	"time"

	"alert_agent/internal/model"

	"gorm.io/gorm"
)

// RuleRepository 规则仓库接口
type RuleRepository interface {
	Create(ctx context.Context, rule *model.Rule) error
	GetByID(ctx context.Context, id string) (*model.Rule, error)
	GetByName(ctx context.Context, name string) (*model.Rule, error)
	Update(ctx context.Context, rule *model.Rule) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, offset, limit int) ([]*model.Rule, int64, error)
	ListByStatus(ctx context.Context, status string, offset, limit int) ([]*model.Rule, int64, error)
	UpdateStatus(ctx context.Context, id, status string) error
	GetByTargets(ctx context.Context, targets []string) ([]*model.Rule, error)
}

// RuleVersionRepository 规则版本仓库接口
type RuleVersionRepository interface {
	Create(ctx context.Context, version *model.RuleVersion) error
	GetByID(ctx context.Context, id string) (*model.RuleVersion, error)
	GetByRuleIDAndVersion(ctx context.Context, ruleID, version string) (*model.RuleVersion, error)
	ListByRuleID(ctx context.Context, ruleID string, offset, limit int) ([]*model.RuleVersion, int64, error)
	GetLatestByRuleID(ctx context.Context, ruleID string) (*model.RuleVersion, error)
	DeleteByRuleID(ctx context.Context, ruleID string) error
}

// RuleAuditLogRepository 规则审计日志仓库接口
type RuleAuditLogRepository interface {
	Create(ctx context.Context, log *model.RuleAuditLog) error
	GetByID(ctx context.Context, id string) (*model.RuleAuditLog, error)
	List(ctx context.Context, filter *RuleAuditLogFilter, offset, limit int) ([]*model.RuleAuditLog, int64, error)
	ListByRuleID(ctx context.Context, ruleID string, offset, limit int) ([]*model.RuleAuditLog, int64, error)
}

// RuleAuditLogFilter 规则审计日志过滤器
type RuleAuditLogFilter struct {
	RuleID string
	Action string
	UserID string
}

// RuleDistributionRepository 规则分发仓库接口
type RuleDistributionRepository interface {
	Create(ctx context.Context, record *model.RuleDistributionRecord) error
	GetByID(ctx context.Context, id string) (*model.RuleDistributionRecord, error)
	GetByRuleIDAndTarget(ctx context.Context, ruleID, target string) (*model.RuleDistributionRecord, error)
	Update(ctx context.Context, record *model.RuleDistributionRecord) error
	Delete(ctx context.Context, id string) error
	ListByRuleID(ctx context.Context, ruleID string) ([]*model.RuleDistributionRecord, error)
	ListByTarget(ctx context.Context, target string) ([]*model.RuleDistributionRecord, error)
	ListByStatus(ctx context.Context, status string, offset, limit int) ([]*model.RuleDistributionRecord, int64, error)
	ListRetryable(ctx context.Context, limit int) ([]*model.RuleDistributionRecord, error)
	BatchUpdateStatus(ctx context.Context, ruleIDs []string, targets []string, status string) error
	GetDistributionSummary(ctx context.Context, ruleIDs []string) ([]*model.RuleDistributionSummary, error)
	DeleteByRuleID(ctx context.Context, ruleID string) error
}

// ruleRepository 规则仓库实现
type ruleRepository struct {
	db *gorm.DB
}

// NewRuleRepository 创建规则仓库实例
func NewRuleRepository(db *gorm.DB) RuleRepository {
	return &ruleRepository{db: db}
}

// Create 创建规则
func (r *ruleRepository) Create(ctx context.Context, rule *model.Rule) error {
	if err := r.db.WithContext(ctx).Create(rule).Error; err != nil {
		return fmt.Errorf("failed to create rule: %w", err)
	}
	return nil
}

// GetByID 根据ID获取规则
func (r *ruleRepository) GetByID(ctx context.Context, id string) (*model.Rule, error) {
	var rule model.Rule
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&rule).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("rule not found: %s", id)
		}
		return nil, fmt.Errorf("failed to get rule: %w", err)
	}
	return &rule, nil
}

// GetByName 根据名称获取规则
func (r *ruleRepository) GetByName(ctx context.Context, name string) (*model.Rule, error) {
	var rule model.Rule
	if err := r.db.WithContext(ctx).Where("name = ?", name).First(&rule).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("rule not found: %s", name)
		}
		return nil, fmt.Errorf("failed to get rule: %w", err)
	}
	return &rule, nil
}

// Update 更新规则
func (r *ruleRepository) Update(ctx context.Context, rule *model.Rule) error {
	if err := r.db.WithContext(ctx).Save(rule).Error; err != nil {
		return fmt.Errorf("failed to update rule: %w", err)
	}
	return nil
}

// Delete 删除规则
func (r *ruleRepository) Delete(ctx context.Context, id string) error {
	result := r.db.WithContext(ctx).Delete(&model.Rule{}, "id = ?", id)
	if result.Error != nil {
		return fmt.Errorf("failed to delete rule: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("rule not found: %s", id)
	}
	return nil
}

// List 获取规则列表
func (r *ruleRepository) List(ctx context.Context, offset, limit int) ([]*model.Rule, int64, error) {
	var rules []*model.Rule
	var total int64

	// 获取总数
	if err := r.db.WithContext(ctx).Model(&model.Rule{}).Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count rules: %w", err)
	}

	// 获取分页数据
	if err := r.db.WithContext(ctx).Offset(offset).Limit(limit).Order("created_at DESC").Find(&rules).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list rules: %w", err)
	}

	return rules, total, nil
}

// ListByStatus 根据状态获取规则列表
func (r *ruleRepository) ListByStatus(ctx context.Context, status string, offset, limit int) ([]*model.Rule, int64, error) {
	var rules []*model.Rule
	var total int64

	query := r.db.WithContext(ctx).Model(&model.Rule{}).Where("status = ?", status)

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count rules by status: %w", err)
	}

	// 获取分页数据
	if err := query.Offset(offset).Limit(limit).Order("created_at DESC").Find(&rules).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list rules by status: %w", err)
	}

	return rules, total, nil
}

// UpdateStatus 更新规则状态
func (r *ruleRepository) UpdateStatus(ctx context.Context, id, status string) error {
	result := r.db.WithContext(ctx).Model(&model.Rule{}).Where("id = ?", id).Update("status", status)
	if result.Error != nil {
		return fmt.Errorf("failed to update rule status: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("rule not found: %s", id)
	}
	return nil
}

// GetByTargets 根据目标获取规则列表
func (r *ruleRepository) GetByTargets(ctx context.Context, targets []string) ([]*model.Rule, error) {
	var rules []*model.Rule
	
	// 构建查询条件，查找包含任一目标的规则
	query := r.db.WithContext(ctx)
	for i, target := range targets {
		if i == 0 {
			query = query.Where("JSON_CONTAINS(targets, ?)", fmt.Sprintf(`"%s"`, target))
		} else {
			query = query.Or("JSON_CONTAINS(targets, ?)", fmt.Sprintf(`"%s"`, target))
		}
	}

	if err := query.Find(&rules).Error; err != nil {
		return nil, fmt.Errorf("failed to get rules by targets: %w", err)
	}

	return rules, nil
}

// ruleVersionRepository 规则版本仓库实现
type ruleVersionRepository struct {
	db *gorm.DB
}

// NewRuleVersionRepository 创建规则版本仓库实例
func NewRuleVersionRepository(db *gorm.DB) RuleVersionRepository {
	return &ruleVersionRepository{db: db}
}

// Create 创建规则版本
func (r *ruleVersionRepository) Create(ctx context.Context, version *model.RuleVersion) error {
	if err := r.db.WithContext(ctx).Create(version).Error; err != nil {
		return fmt.Errorf("failed to create rule version: %w", err)
	}
	return nil
}

// GetByID 根据ID获取规则版本
func (r *ruleVersionRepository) GetByID(ctx context.Context, id string) (*model.RuleVersion, error) {
	var version model.RuleVersion
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&version).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("rule version not found: %s", id)
		}
		return nil, fmt.Errorf("failed to get rule version: %w", err)
	}
	return &version, nil
}

// GetByRuleIDAndVersion 根据规则ID和版本号获取规则版本
func (r *ruleVersionRepository) GetByRuleIDAndVersion(ctx context.Context, ruleID, version string) (*model.RuleVersion, error) {
	var ruleVersion model.RuleVersion
	if err := r.db.WithContext(ctx).Where("rule_id = ? AND version = ?", ruleID, version).First(&ruleVersion).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("rule version not found: %s@%s", ruleID, version)
		}
		return nil, fmt.Errorf("failed to get rule version: %w", err)
	}
	return &ruleVersion, nil
}

// ListByRuleID 根据规则ID获取版本列表
func (r *ruleVersionRepository) ListByRuleID(ctx context.Context, ruleID string, offset, limit int) ([]*model.RuleVersion, int64, error) {
	var versions []*model.RuleVersion
	var total int64

	query := r.db.WithContext(ctx).Model(&model.RuleVersion{}).Where("rule_id = ?", ruleID)

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count rule versions: %w", err)
	}

	// 获取分页数据，按创建时间倒序
	if err := query.Offset(offset).Limit(limit).Order("created_at DESC").Find(&versions).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list rule versions: %w", err)
	}

	return versions, total, nil
}

// GetLatestByRuleID 获取规则的最新版本
func (r *ruleVersionRepository) GetLatestByRuleID(ctx context.Context, ruleID string) (*model.RuleVersion, error) {
	var version model.RuleVersion
	if err := r.db.WithContext(ctx).Where("rule_id = ?", ruleID).Order("created_at DESC").First(&version).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("no versions found for rule: %s", ruleID)
		}
		return nil, fmt.Errorf("failed to get latest rule version: %w", err)
	}
	return &version, nil
}

// DeleteByRuleID 删除规则的所有版本
func (r *ruleVersionRepository) DeleteByRuleID(ctx context.Context, ruleID string) error {
	if err := r.db.WithContext(ctx).Where("rule_id = ?", ruleID).Delete(&model.RuleVersion{}).Error; err != nil {
		return fmt.Errorf("failed to delete rule versions: %w", err)
	}
	return nil
}

// ruleAuditLogRepository 规则审计日志仓库实现
type ruleAuditLogRepository struct {
	db *gorm.DB
}

// NewRuleAuditLogRepository 创建规则审计日志仓库实例
func NewRuleAuditLogRepository(db *gorm.DB) RuleAuditLogRepository {
	return &ruleAuditLogRepository{db: db}
}

// Create 创建审计日志
func (r *ruleAuditLogRepository) Create(ctx context.Context, log *model.RuleAuditLog) error {
	if err := r.db.WithContext(ctx).Create(log).Error; err != nil {
		return fmt.Errorf("failed to create rule audit log: %w", err)
	}
	return nil
}

// GetByID 根据ID获取审计日志
func (r *ruleAuditLogRepository) GetByID(ctx context.Context, id string) (*model.RuleAuditLog, error) {
	var log model.RuleAuditLog
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&log).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("rule audit log not found: %s", id)
		}
		return nil, fmt.Errorf("failed to get rule audit log: %w", err)
	}
	return &log, nil
}

// List 获取审计日志列表
func (r *ruleAuditLogRepository) List(ctx context.Context, filter *RuleAuditLogFilter, offset, limit int) ([]*model.RuleAuditLog, int64, error) {
	var logs []*model.RuleAuditLog
	var total int64

	query := r.db.WithContext(ctx).Model(&model.RuleAuditLog{})

	// 应用过滤条件
	if filter != nil {
		if filter.RuleID != "" {
			query = query.Where("rule_id = ?", filter.RuleID)
		}
		if filter.Action != "" {
			query = query.Where("action = ?", filter.Action)
		}
		if filter.UserID != "" {
			query = query.Where("user_id = ?", filter.UserID)
		}
	}

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count rule audit logs: %w", err)
	}

	// 获取分页数据，按创建时间倒序
	if err := query.Offset(offset).Limit(limit).Order("created_at DESC").Find(&logs).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list rule audit logs: %w", err)
	}

	return logs, total, nil
}

// ListByRuleID 根据规则ID获取审计日志列表
func (r *ruleAuditLogRepository) ListByRuleID(ctx context.Context, ruleID string, offset, limit int) ([]*model.RuleAuditLog, int64, error) {
	filter := &RuleAuditLogFilter{RuleID: ruleID}
	return r.List(ctx, filter, offset, limit)
}

// ruleDistributionRepository 规则分发仓库实现
type ruleDistributionRepository struct {
	db *gorm.DB
}

// NewRuleDistributionRepository 创建规则分发仓库实例
func NewRuleDistributionRepository(db *gorm.DB) RuleDistributionRepository {
	return &ruleDistributionRepository{db: db}
}

// Create 创建分发记录
func (r *ruleDistributionRepository) Create(ctx context.Context, record *model.RuleDistributionRecord) error {
	if err := r.db.WithContext(ctx).Create(record).Error; err != nil {
		return fmt.Errorf("failed to create distribution record: %w", err)
	}
	return nil
}

// GetByID 根据ID获取分发记录
func (r *ruleDistributionRepository) GetByID(ctx context.Context, id string) (*model.RuleDistributionRecord, error) {
	var record model.RuleDistributionRecord
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&record).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("distribution record not found: %s", id)
		}
		return nil, fmt.Errorf("failed to get distribution record: %w", err)
	}
	return &record, nil
}

// GetByRuleIDAndTarget 根据规则ID和目标获取分发记录
func (r *ruleDistributionRepository) GetByRuleIDAndTarget(ctx context.Context, ruleID, target string) (*model.RuleDistributionRecord, error) {
	var record model.RuleDistributionRecord
	if err := r.db.WithContext(ctx).Where("rule_id = ? AND target = ?", ruleID, target).First(&record).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("distribution record not found: %s@%s", ruleID, target)
		}
		return nil, fmt.Errorf("failed to get distribution record: %w", err)
	}
	return &record, nil
}

// Update 更新分发记录
func (r *ruleDistributionRepository) Update(ctx context.Context, record *model.RuleDistributionRecord) error {
	if err := r.db.WithContext(ctx).Save(record).Error; err != nil {
		return fmt.Errorf("failed to update distribution record: %w", err)
	}
	return nil
}

// Delete 删除分发记录
func (r *ruleDistributionRepository) Delete(ctx context.Context, id string) error {
	result := r.db.WithContext(ctx).Delete(&model.RuleDistributionRecord{}, "id = ?", id)
	if result.Error != nil {
		return fmt.Errorf("failed to delete distribution record: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("distribution record not found: %s", id)
	}
	return nil
}

// ListByRuleID 根据规则ID获取分发记录列表
func (r *ruleDistributionRepository) ListByRuleID(ctx context.Context, ruleID string) ([]*model.RuleDistributionRecord, error) {
	var records []*model.RuleDistributionRecord
	if err := r.db.WithContext(ctx).Where("rule_id = ?", ruleID).Order("created_at DESC").Find(&records).Error; err != nil {
		return nil, fmt.Errorf("failed to list distribution records by rule ID: %w", err)
	}
	return records, nil
}

// ListByTarget 根据目标获取分发记录列表
func (r *ruleDistributionRepository) ListByTarget(ctx context.Context, target string) ([]*model.RuleDistributionRecord, error) {
	var records []*model.RuleDistributionRecord
	if err := r.db.WithContext(ctx).Where("target = ?", target).Order("created_at DESC").Find(&records).Error; err != nil {
		return nil, fmt.Errorf("failed to list distribution records by target: %w", err)
	}
	return records, nil
}

// ListByStatus 根据状态获取分发记录列表
func (r *ruleDistributionRepository) ListByStatus(ctx context.Context, status string, offset, limit int) ([]*model.RuleDistributionRecord, int64, error) {
	var records []*model.RuleDistributionRecord
	var total int64

	query := r.db.WithContext(ctx).Model(&model.RuleDistributionRecord{}).Where("status = ?", status)

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count distribution records by status: %w", err)
	}

	// 获取分页数据
	if err := query.Offset(offset).Limit(limit).Order("created_at DESC").Find(&records).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list distribution records by status: %w", err)
	}

	return records, total, nil
}

// ListRetryable 获取可重试的分发记录
func (r *ruleDistributionRepository) ListRetryable(ctx context.Context, limit int) ([]*model.RuleDistributionRecord, error) {
	var records []*model.RuleDistributionRecord
	
	query := r.db.WithContext(ctx).Where("status = ? AND retry_count < max_retry", "failed")
	
	// 添加时间条件：next_retry 为空或者已经到了重试时间
	query = query.Where("next_retry IS NULL OR next_retry <= ?", time.Now())
	
	if limit > 0 {
		query = query.Limit(limit)
	}
	
	if err := query.Order("created_at ASC").Find(&records).Error; err != nil {
		return nil, fmt.Errorf("failed to list retryable distribution records: %w", err)
	}
	
	return records, nil
}

// BatchUpdateStatus 批量更新状态
func (r *ruleDistributionRepository) BatchUpdateStatus(ctx context.Context, ruleIDs []string, targets []string, status string) error {
	query := r.db.WithContext(ctx).Model(&model.RuleDistributionRecord{})
	
	if len(ruleIDs) > 0 {
		query = query.Where("rule_id IN ?", ruleIDs)
	}
	
	if len(targets) > 0 {
		query = query.Where("target IN ?", targets)
	}
	
	updates := map[string]interface{}{
		"status":     status,
		"updated_at": time.Now(),
	}
	
	if status == "success" {
		updates["last_sync"] = time.Now()
		updates["error"] = ""
	}
	
	if err := query.Updates(updates).Error; err != nil {
		return fmt.Errorf("failed to batch update distribution status: %w", err)
	}
	
	return nil
}

// GetDistributionSummary 获取分发汇总信息
func (r *ruleDistributionRepository) GetDistributionSummary(ctx context.Context, ruleIDs []string) ([]*model.RuleDistributionSummary, error) {
	var summaries []*model.RuleDistributionSummary
	
	// 构建查询
	query := `
		SELECT 
			r.id as rule_id,
			r.name as rule_name,
			r.version,
			COUNT(rd.id) as total_targets,
			SUM(CASE WHEN rd.status = 'success' THEN 1 ELSE 0 END) as success_count,
			SUM(CASE WHEN rd.status = 'failed' THEN 1 ELSE 0 END) as failed_count,
			SUM(CASE WHEN rd.status = 'pending' THEN 1 ELSE 0 END) as pending_count,
			MAX(rd.last_sync) as last_sync
		FROM alert_rules r
		LEFT JOIN rule_distribution_records rd ON r.id = rd.rule_id AND rd.deleted_at IS NULL
		WHERE r.deleted_at IS NULL
	`
	
	args := []interface{}{}
	if len(ruleIDs) > 0 {
		query += " AND r.id IN ?"
		args = append(args, ruleIDs)
	}
	
	query += " GROUP BY r.id, r.name, r.version ORDER BY r.created_at DESC"
	
	rows, err := r.db.WithContext(ctx).Raw(query, args...).Rows()
	if err != nil {
		return nil, fmt.Errorf("failed to get distribution summary: %w", err)
	}
	defer rows.Close()
	
	for rows.Next() {
		var summary model.RuleDistributionSummary
		var lastSync *time.Time
		
		if err := rows.Scan(
			&summary.RuleID,
			&summary.RuleName,
			&summary.Version,
			&summary.TotalTargets,
			&summary.SuccessCount,
			&summary.FailedCount,
			&summary.PendingCount,
			&lastSync,
		); err != nil {
			return nil, fmt.Errorf("failed to scan distribution summary: %w", err)
		}
		
		if lastSync != nil {
			summary.LastSync = *lastSync
		}
		
		// 获取目标详细信息
		targets, err := r.getTargetDistributionInfo(ctx, summary.RuleID)
		if err != nil {
			return nil, fmt.Errorf("failed to get target distribution info: %w", err)
		}
		summary.Targets = targets
		
		summaries = append(summaries, &summary)
	}
	
	return summaries, nil
}

// getTargetDistributionInfo 获取目标分发信息
func (r *ruleDistributionRepository) getTargetDistributionInfo(ctx context.Context, ruleID string) ([]model.TargetDistributionInfo, error) {
	var records []*model.RuleDistributionRecord
	if err := r.db.WithContext(ctx).Where("rule_id = ?", ruleID).Find(&records).Error; err != nil {
		return nil, fmt.Errorf("failed to get target distribution records: %w", err)
	}
	
	var targets []model.TargetDistributionInfo
	for _, record := range records {
		target := model.TargetDistributionInfo{
			Target:     record.Target,
			Status:     record.Status,
			Version:    record.Version,
			ConfigHash: record.ConfigHash,
			LastSync:   record.LastSync,
			Error:      record.Error,
			RetryCount: record.RetryCount,
			NextRetry:  record.NextRetry,
		}
		targets = append(targets, target)
	}
	
	return targets, nil
}

// DeleteByRuleID 删除规则的所有分发记录
func (r *ruleDistributionRepository) DeleteByRuleID(ctx context.Context, ruleID string) error {
	if err := r.db.WithContext(ctx).Where("rule_id = ?", ruleID).Delete(&model.RuleDistributionRecord{}).Error; err != nil {
		return fmt.Errorf("failed to delete distribution records by rule ID: %w", err)
	}
	return nil
}