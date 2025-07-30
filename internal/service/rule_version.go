package service

import (
	"context"
	"fmt"
	"reflect"
	"time"

	"alert_agent/internal/model"
	"alert_agent/internal/repository"

	"github.com/google/uuid"
)

// RuleVersionService 规则版本服务接口
type RuleVersionService interface {
	// 版本管理
	CreateVersion(ctx context.Context, rule *model.Rule, changeType, changedBy, changeNote string) (*model.RuleVersion, error)
	GetVersionsByRuleID(ctx context.Context, ruleID string, page, pageSize int) ([]*model.RuleVersion, int64, error)
	GetVersionByRuleIDAndVersion(ctx context.Context, ruleID, version string) (*model.RuleVersion, error)
	
	// 版本对比
	CompareVersions(ctx context.Context, req *model.RuleVersionCompareRequest) (*model.RuleVersionCompareResponse, error)
	
	// 规则回滚
	RollbackRule(ctx context.Context, req *model.RuleRollbackRequest, userID, userName, ipAddress, userAgent string) (*model.Rule, error)
	
	// 审计日志
	CreateAuditLog(ctx context.Context, log *model.RuleAuditLog) error
	GetAuditLogs(ctx context.Context, req *model.RuleAuditLogListRequest) ([]*model.RuleAuditLog, int64, error)
}

// ruleVersionService 规则版本服务实现
type ruleVersionService struct {
	ruleRepo        repository.RuleRepository
	versionRepo     repository.RuleVersionRepository
	auditLogRepo    repository.RuleAuditLogRepository
}

// NewRuleVersionService 创建规则版本服务实例
func NewRuleVersionService(
	ruleRepo repository.RuleRepository,
	versionRepo repository.RuleVersionRepository,
	auditLogRepo repository.RuleAuditLogRepository,
) RuleVersionService {
	return &ruleVersionService{
		ruleRepo:     ruleRepo,
		versionRepo:  versionRepo,
		auditLogRepo: auditLogRepo,
	}
}

// CreateVersion 创建规则版本
func (s *ruleVersionService) CreateVersion(ctx context.Context, rule *model.Rule, changeType, changedBy, changeNote string) (*model.RuleVersion, error) {
	version := &model.RuleVersion{
		ID:          uuid.New().String(),
		RuleID:      rule.ID,
		Version:     rule.Version,
		Name:        rule.Name,
		Expression:  rule.Expression,
		Duration:    rule.Duration,
		Severity:    rule.Severity,
		Labels:      rule.Labels,
		Annotations: rule.Annotations,
		Targets:     rule.Targets,
		Status:      rule.Status,
		ChangeType:  changeType,
		ChangedBy:   changedBy,
		ChangeNote:  changeNote,
		CreatedAt:   time.Now(),
	}

	if err := s.versionRepo.Create(ctx, version); err != nil {
		return nil, fmt.Errorf("failed to create rule version: %w", err)
	}

	return version, nil
}

// GetVersionsByRuleID 获取规则的版本列表
func (s *ruleVersionService) GetVersionsByRuleID(ctx context.Context, ruleID string, page, pageSize int) ([]*model.RuleVersion, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100
	}

	offset := (page - 1) * pageSize
	versions, total, err := s.versionRepo.ListByRuleID(ctx, ruleID, offset, pageSize)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get rule versions: %w", err)
	}

	return versions, total, nil
}

// GetVersionByRuleIDAndVersion 获取指定版本的规则
func (s *ruleVersionService) GetVersionByRuleIDAndVersion(ctx context.Context, ruleID, version string) (*model.RuleVersion, error) {
	ruleVersion, err := s.versionRepo.GetByRuleIDAndVersion(ctx, ruleID, version)
	if err != nil {
		return nil, fmt.Errorf("failed to get rule version: %w", err)
	}

	return ruleVersion, nil
}

// CompareVersions 对比两个版本的差异
func (s *ruleVersionService) CompareVersions(ctx context.Context, req *model.RuleVersionCompareRequest) (*model.RuleVersionCompareResponse, error) {
	// 获取两个版本
	oldVersion, err := s.versionRepo.GetByRuleIDAndVersion(ctx, req.RuleID, req.OldVersion)
	if err != nil {
		return nil, fmt.Errorf("failed to get old version: %w", err)
	}

	newVersion, err := s.versionRepo.GetByRuleIDAndVersion(ctx, req.RuleID, req.NewVersion)
	if err != nil {
		return nil, fmt.Errorf("failed to get new version: %w", err)
	}

	// 对比差异
	differences := s.compareRuleVersions(oldVersion, newVersion)

	response := &model.RuleVersionCompareResponse{
		RuleID:      req.RuleID,
		OldVersion:  oldVersion,
		NewVersion:  newVersion,
		Differences: differences,
	}

	return response, nil
}

// RollbackRule 回滚规则到指定版本
func (s *ruleVersionService) RollbackRule(ctx context.Context, req *model.RuleRollbackRequest, userID, userName, ipAddress, userAgent string) (*model.Rule, error) {
	// 1. 获取当前规则
	currentRule, err := s.ruleRepo.GetByID(ctx, req.RuleID)
	if err != nil {
		return nil, fmt.Errorf("failed to get current rule: %w", err)
	}

	// 2. 获取目标版本
	targetVersion, err := s.versionRepo.GetByRuleIDAndVersion(ctx, req.RuleID, req.ToVersion)
	if err != nil {
		return nil, fmt.Errorf("failed to get target version: %w", err)
	}

	// 3. 记录当前版本（用于审计）
	oldVersion := currentRule.Version

	// 4. 创建当前版本的备份（如果还没有的话）
	if _, err := s.versionRepo.GetByRuleIDAndVersion(ctx, req.RuleID, currentRule.Version); err != nil {
		if _, createErr := s.CreateVersion(ctx, currentRule, "backup", userID, "Backup before rollback"); createErr != nil {
			return nil, fmt.Errorf("failed to backup current version: %w", createErr)
		}
	}

	// 5. 恢复规则到目标版本
	restoredRule := targetVersion.ToRule()
	restoredRule.Version = s.generateNextVersion(currentRule.Version)
	restoredRule.Status = "pending" // 回滚后需要重新分发
	restoredRule.UpdatedAt = time.Now()

	if err := s.ruleRepo.Update(ctx, restoredRule); err != nil {
		return nil, fmt.Errorf("failed to rollback rule: %w", err)
	}

	// 6. 创建回滚后的版本记录
	if _, err := s.CreateVersion(ctx, restoredRule, "rollback", userID, fmt.Sprintf("Rollback to version %s: %s", req.ToVersion, req.Note)); err != nil {
		return nil, fmt.Errorf("failed to create rollback version: %w", err)
	}

	// 7. 记录审计日志
	changes := map[string]interface{}{
		"rollback_from": oldVersion,
		"rollback_to":   req.ToVersion,
		"new_version":   restoredRule.Version,
	}

	auditLog := &model.RuleAuditLog{
		ID:         uuid.New().String(),
		RuleID:     req.RuleID,
		Action:     "rollback",
		OldVersion: oldVersion,
		NewVersion: restoredRule.Version,
		UserID:     userID,
		UserName:   userName,
		IPAddress:  ipAddress,
		UserAgent:  userAgent,
		Note:       req.Note,
		CreatedAt:  time.Now(),
	}

	if err := auditLog.SetChangesMap(changes); err != nil {
		return nil, fmt.Errorf("failed to set audit log changes: %w", err)
	}

	if err := s.auditLogRepo.Create(ctx, auditLog); err != nil {
		return nil, fmt.Errorf("failed to create audit log: %w", err)
	}

	return restoredRule, nil
}

// CreateAuditLog 创建审计日志
func (s *ruleVersionService) CreateAuditLog(ctx context.Context, log *model.RuleAuditLog) error {
	if err := s.auditLogRepo.Create(ctx, log); err != nil {
		return fmt.Errorf("failed to create audit log: %w", err)
	}
	return nil
}

// GetAuditLogs 获取审计日志列表
func (s *ruleVersionService) GetAuditLogs(ctx context.Context, req *model.RuleAuditLogListRequest) ([]*model.RuleAuditLog, int64, error) {
	if req.Page < 1 {
		req.Page = 1
	}
	if req.PageSize < 1 {
		req.PageSize = 10
	}
	if req.PageSize > 100 {
		req.PageSize = 100
	}

	offset := (req.Page - 1) * req.PageSize
	filter := &repository.RuleAuditLogFilter{
		RuleID: req.RuleID,
		Action: req.Action,
		UserID: req.UserID,
	}

	logs, total, err := s.auditLogRepo.List(ctx, filter, offset, req.PageSize)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get audit logs: %w", err)
	}

	return logs, total, nil
}

// compareRuleVersions 对比两个规则版本的差异
func (s *ruleVersionService) compareRuleVersions(oldVersion, newVersion *model.RuleVersion) []model.RuleVersionDifference {
	var differences []model.RuleVersionDifference

	// 对比基本字段
	if oldVersion.Name != newVersion.Name {
		differences = append(differences, model.RuleVersionDifference{
			Field:    "name",
			OldValue: oldVersion.Name,
			NewValue: newVersion.Name,
			Type:     "modified",
		})
	}

	if oldVersion.Expression != newVersion.Expression {
		differences = append(differences, model.RuleVersionDifference{
			Field:    "expression",
			OldValue: oldVersion.Expression,
			NewValue: newVersion.Expression,
			Type:     "modified",
		})
	}

	if oldVersion.Duration != newVersion.Duration {
		differences = append(differences, model.RuleVersionDifference{
			Field:    "duration",
			OldValue: oldVersion.Duration,
			NewValue: newVersion.Duration,
			Type:     "modified",
		})
	}

	if oldVersion.Severity != newVersion.Severity {
		differences = append(differences, model.RuleVersionDifference{
			Field:    "severity",
			OldValue: oldVersion.Severity,
			NewValue: newVersion.Severity,
			Type:     "modified",
		})
	}

	// 对比标签
	oldLabels, _ := oldVersion.GetLabelsMap()
	newLabels, _ := newVersion.GetLabelsMap()
	labelDiffs := s.compareStringMaps("labels", oldLabels, newLabels)
	differences = append(differences, labelDiffs...)

	// 对比注释
	oldAnnotations, _ := oldVersion.GetAnnotationsMap()
	newAnnotations, _ := newVersion.GetAnnotationsMap()
	annotationDiffs := s.compareStringMaps("annotations", oldAnnotations, newAnnotations)
	differences = append(differences, annotationDiffs...)

	// 对比目标
	oldTargets, _ := oldVersion.GetTargetsList()
	newTargets, _ := newVersion.GetTargetsList()
	targetDiffs := s.compareStringSlices("targets", oldTargets, newTargets)
	differences = append(differences, targetDiffs...)

	return differences
}

// compareStringMaps 对比字符串映射的差异
func (s *ruleVersionService) compareStringMaps(fieldPrefix string, oldMap, newMap map[string]string) []model.RuleVersionDifference {
	var differences []model.RuleVersionDifference

	// 检查删除和修改的项
	for key, oldValue := range oldMap {
		if newValue, exists := newMap[key]; exists {
			if oldValue != newValue {
				differences = append(differences, model.RuleVersionDifference{
					Field:    fmt.Sprintf("%s.%s", fieldPrefix, key),
					OldValue: oldValue,
					NewValue: newValue,
					Type:     "modified",
				})
			}
		} else {
			differences = append(differences, model.RuleVersionDifference{
				Field:    fmt.Sprintf("%s.%s", fieldPrefix, key),
				OldValue: oldValue,
				NewValue: nil,
				Type:     "removed",
			})
		}
	}

	// 检查新增的项
	for key, newValue := range newMap {
		if _, exists := oldMap[key]; !exists {
			differences = append(differences, model.RuleVersionDifference{
				Field:    fmt.Sprintf("%s.%s", fieldPrefix, key),
				OldValue: nil,
				NewValue: newValue,
				Type:     "added",
			})
		}
	}

	return differences
}

// compareStringSlices 对比字符串切片的差异
func (s *ruleVersionService) compareStringSlices(fieldName string, oldSlice, newSlice []string) []model.RuleVersionDifference {
	var differences []model.RuleVersionDifference

	if !reflect.DeepEqual(oldSlice, newSlice) {
		differences = append(differences, model.RuleVersionDifference{
			Field:    fieldName,
			OldValue: oldSlice,
			NewValue: newSlice,
			Type:     "modified",
		})
	}

	return differences
}

// generateNextVersion 生成下一个版本号
func (s *ruleVersionService) generateNextVersion(currentVersion string) string {
	// 这里使用简单的版本递增逻辑，实际项目中可能需要更复杂的版本管理
	// 例如：v1.0.0 -> v1.0.1
	return incrementVersion(currentVersion)
}

// incrementVersion 递增版本号（从 rule.go 复制过来的辅助函数）
func incrementVersion(currentVersion string) string {
	// 简单的版本递增逻辑：v1.0.0 -> v1.0.1
	if currentVersion == "" {
		return "v1.0.1"
	}

	// 这里可以实现更复杂的版本递增逻辑
	// 为了简化，我们直接在当前版本后面加上时间戳
	timestamp := time.Now().Unix()
	return fmt.Sprintf("%s.%d", currentVersion, timestamp)
}