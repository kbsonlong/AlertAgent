package service

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	"alert_agent/internal/model"
	"alert_agent/internal/repository"

	"github.com/google/uuid"
)

// RuleService 规则服务接口
type RuleService interface {
	CreateRule(ctx context.Context, req *model.CreateRuleRequest) (*model.Rule, error)
	GetRule(ctx context.Context, id string) (*model.Rule, error)
	UpdateRule(ctx context.Context, id string, req *model.UpdateRuleRequest) (*model.Rule, error)
	DeleteRule(ctx context.Context, id string) error
	ListRules(ctx context.Context, page, pageSize int) ([]*model.Rule, int64, error)
	GetDistributionStatus(ctx context.Context, ruleID string) (*model.RuleDistributionStatus, error)
	ValidateRule(ctx context.Context, expression, duration string) error
	
	// 版本控制相关方法
	CreateRuleWithAudit(ctx context.Context, req *model.CreateRuleRequest, userID, userName, ipAddress, userAgent string) (*model.Rule, error)
	UpdateRuleWithAudit(ctx context.Context, id string, req *model.UpdateRuleRequest, userID, userName, ipAddress, userAgent string) (*model.Rule, error)
	DeleteRuleWithAudit(ctx context.Context, id string, userID, userName, ipAddress, userAgent string) error
}

// ruleService 规则服务实现
type ruleService struct {
	ruleRepo       repository.RuleRepository
	validator      RuleValidator
	versionService RuleVersionService
}

// NewRuleService 创建规则服务实例
func NewRuleService(ruleRepo repository.RuleRepository, validator RuleValidator, versionService RuleVersionService) RuleService {
	return &ruleService{
		ruleRepo:       ruleRepo,
		validator:      validator,
		versionService: versionService,
	}
}

// CreateRule 创建规则
func (s *ruleService) CreateRule(ctx context.Context, req *model.CreateRuleRequest) (*model.Rule, error) {
	// 1. 验证规则语法
	if err := s.validator.ValidateRule(req.Expression, req.Duration); err != nil {
		return nil, fmt.Errorf("rule validation failed: %w", err)
	}

	// 2. 验证严重程度
	if err := s.validator.ValidateSeverity(req.Severity); err != nil {
		return nil, fmt.Errorf("severity validation failed: %w", err)
	}

	// 3. 检查规则名称是否已存在
	if existingRule, _ := s.ruleRepo.GetByName(ctx, req.Name); existingRule != nil {
		return nil, fmt.Errorf("rule with name '%s' already exists", req.Name)
	}

	// 4. 创建规则对象
	rule := &model.Rule{
		ID:         uuid.New().String(),
		Name:       req.Name,
		Expression: req.Expression,
		Duration:   req.Duration,
		Severity:   req.Severity,
		Version:    "v1.0.0",
		Status:     "pending",
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	// 5. 设置标签、注释和目标
	if err := rule.SetLabelsMap(req.Labels); err != nil {
		return nil, fmt.Errorf("failed to set labels: %w", err)
	}

	if err := rule.SetAnnotationsMap(req.Annotations); err != nil {
		return nil, fmt.Errorf("failed to set annotations: %w", err)
	}

	if err := rule.SetTargetsList(req.Targets); err != nil {
		return nil, fmt.Errorf("failed to set targets: %w", err)
	}

	// 6. 保存到数据库
	if err := s.ruleRepo.Create(ctx, rule); err != nil {
		return nil, fmt.Errorf("failed to create rule: %w", err)
	}

	// TODO: 7. 发布配置同步任务（在后续任务中实现）
	// task := &ConfigSyncTask{
	//     Type:     "rule_create",
	//     RuleID:   rule.ID,
	//     Targets:  req.Targets,
	//     Priority: "normal",
	// }
	// if err := s.taskPublisher.PublishTask(ctx, "config_sync", task); err != nil {
	//     log.Errorf("failed to publish config sync task: %v", err)
	// }

	return rule, nil
}

// GetRule 获取规则
func (s *ruleService) GetRule(ctx context.Context, id string) (*model.Rule, error) {
	rule, err := s.ruleRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get rule: %w", err)
	}
	return rule, nil
}

// UpdateRule 更新规则
func (s *ruleService) UpdateRule(ctx context.Context, id string, req *model.UpdateRuleRequest) (*model.Rule, error) {
	// 1. 获取现有规则
	rule, err := s.ruleRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("rule not found: %w", err)
	}

	// 2. 验证新规则语法（如果提供了新的表达式或持续时间）
	expression := rule.Expression
	duration := rule.Duration

	if req.Expression != "" {
		expression = req.Expression
	}
	if req.Duration != "" {
		duration = req.Duration
	}

	if req.Expression != "" || req.Duration != "" {
		if err := s.validator.ValidateRule(expression, duration); err != nil {
			return nil, fmt.Errorf("rule validation failed: %w", err)
		}
	}

	// 3. 验证严重程度（如果提供了新的严重程度）
	if req.Severity != "" {
		if err := s.validator.ValidateSeverity(req.Severity); err != nil {
			return nil, fmt.Errorf("severity validation failed: %w", err)
		}
	}

	// 4. 检查规则名称是否与其他规则冲突
	if req.Name != "" && req.Name != rule.Name {
		if existingRule, _ := s.ruleRepo.GetByName(ctx, req.Name); existingRule != nil && existingRule.ID != id {
			return nil, fmt.Errorf("rule with name '%s' already exists", req.Name)
		}
	}

	// 5. 更新规则字段
	if req.Name != "" {
		rule.Name = req.Name
	}
	if req.Expression != "" {
		rule.Expression = req.Expression
	}
	if req.Duration != "" {
		rule.Duration = req.Duration
	}
	if req.Severity != "" {
		rule.Severity = req.Severity
	}

	// 6. 更新标签、注释和目标
	if req.Labels != nil {
		if err := rule.SetLabelsMap(req.Labels); err != nil {
			return nil, fmt.Errorf("failed to set labels: %w", err)
		}
	}

	if req.Annotations != nil {
		if err := rule.SetAnnotationsMap(req.Annotations); err != nil {
			return nil, fmt.Errorf("failed to set annotations: %w", err)
		}
	}

	if req.Targets != nil {
		if err := rule.SetTargetsList(req.Targets); err != nil {
			return nil, fmt.Errorf("failed to set targets: %w", err)
		}
	}

	// 7. 版本递增
	rule.Version = s.incrementVersion(rule.Version)
	rule.UpdatedAt = time.Now()
	rule.Status = "pending"

	// 8. 保存更新
	if err := s.ruleRepo.Update(ctx, rule); err != nil {
		return nil, fmt.Errorf("failed to update rule: %w", err)
	}

	// TODO: 9. 发布配置同步任务（在后续任务中实现）
	// task := &ConfigSyncTask{
	//     Type:     "rule_update",
	//     RuleID:   rule.ID,
	//     Targets:  targets,
	//     Priority: "normal",
	// }
	// if err := s.taskPublisher.PublishTask(ctx, "config_sync", task); err != nil {
	//     log.Errorf("failed to publish config sync task: %v", err)
	// }

	return rule, nil
}

// DeleteRule 删除规则
func (s *ruleService) DeleteRule(ctx context.Context, id string) error {
	// 1. 检查规则是否存在
	_, err := s.ruleRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("rule not found: %w", err)
	}

	// 2. 删除规则
	if err := s.ruleRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete rule: %w", err)
	}

	// TODO: 3. 发布配置同步任务（在后续任务中实现）
	// targets, _ := rule.GetTargetsList()
	// task := &ConfigSyncTask{
	//     Type:     "rule_delete",
	//     RuleID:   rule.ID,
	//     Targets:  targets,
	//     Priority: "normal",
	// }
	// if err := s.taskPublisher.PublishTask(ctx, "config_sync", task); err != nil {
	//     log.Errorf("failed to publish config sync task: %v", err)
	// }

	return nil
}

// ListRules 获取规则列表
func (s *ruleService) ListRules(ctx context.Context, page, pageSize int) ([]*model.Rule, int64, error) {
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
	rules, total, err := s.ruleRepo.List(ctx, offset, pageSize)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list rules: %w", err)
	}

	return rules, total, nil
}

// GetDistributionStatus 获取规则分发状态
func (s *ruleService) GetDistributionStatus(ctx context.Context, ruleID string) (*model.RuleDistributionStatus, error) {
	// 1. 获取规则信息
	rule, err := s.ruleRepo.GetByID(ctx, ruleID)
	if err != nil {
		return nil, fmt.Errorf("rule not found: %w", err)
	}

	// 2. 获取目标列表
	targets, err := rule.GetTargetsList()
	if err != nil {
		return nil, fmt.Errorf("failed to get targets: %w", err)
	}

	// 3. 构建分发状态（目前返回基本信息，具体的同步状态在后续任务中实现）
	status := &model.RuleDistributionStatus{
		RuleID:   rule.ID,
		RuleName: rule.Name,
		Version:  rule.Version,
		Targets:  targets,
		Status:   rule.Status,
		LastSync: rule.UpdatedAt,
		TargetStatus: make([]model.TargetDistributionStatus, len(targets)),
	}

	// 4. 填充目标状态（目前使用规则状态，具体实现在后续任务中完善）
	for i, target := range targets {
		status.TargetStatus[i] = model.TargetDistributionStatus{
			Target:   target,
			Status:   rule.Status,
			LastSync: rule.UpdatedAt,
		}
	}

	return status, nil
}

// ValidateRule 验证规则
func (s *ruleService) ValidateRule(ctx context.Context, expression, duration string) error {
	return s.validator.ValidateRule(expression, duration)
}

// incrementVersion 递增版本号
func (s *ruleService) incrementVersion(currentVersion string) string {
	// 简单的版本递增逻辑：v1.0.0 -> v1.0.1
	if !strings.HasPrefix(currentVersion, "v") {
		return "v1.0.1"
	}

	// 移除 'v' 前缀
	version := strings.TrimPrefix(currentVersion, "v")
	parts := strings.Split(version, ".")

	if len(parts) != 3 {
		return "v1.0.1"
	}

	// 递增补丁版本号
	patch, err := strconv.Atoi(parts[2])
	if err != nil {
		return "v1.0.1"
	}

	patch++
	return fmt.Sprintf("v%s.%s.%d", parts[0], parts[1], patch)
}

// CreateRuleWithAudit 创建规则并记录审计日志
func (s *ruleService) CreateRuleWithAudit(ctx context.Context, req *model.CreateRuleRequest, userID, userName, ipAddress, userAgent string) (*model.Rule, error) {
	// 创建规则
	rule, err := s.CreateRule(ctx, req)
	if err != nil {
		return nil, err
	}

	// 如果版本服务可用，创建版本记录和审计日志
	if s.versionService != nil {
		// 创建版本记录
		if _, versionErr := s.versionService.CreateVersion(ctx, rule, "create", userID, "Initial version"); versionErr != nil {
			// 版本记录创建失败不影响规则创建，只记录日志
			fmt.Printf("Failed to create version record: %v\n", versionErr)
		}

		// 创建审计日志
		auditLog := &model.RuleAuditLog{
			ID:         uuid.New().String(),
			RuleID:     rule.ID,
			Action:     "create",
			Changes:    fmt.Sprintf(`{"action":"create","new_version":"%s"}`, rule.Version),
			UserID:     userID,
			UserName:   userName,
			IPAddress:  ipAddress,
			UserAgent:  userAgent,
			Reason:     "Rule created",
			CreatedAt:  time.Now(),
		}

		changes := map[string]interface{}{
			"name":        rule.Name,
			"expression":  rule.Expression,
			"duration":    rule.Duration,
			"severity":    rule.Severity,
		}

		// Set changes as JSON string
		if changesJSON, err := json.Marshal(changes); err == nil {
			auditLog.Changes = string(changesJSON)
			if auditErr := s.versionService.CreateAuditLog(ctx, auditLog); auditErr != nil {
				fmt.Printf("Failed to create audit log: %v\n", auditErr)
			}
		}
	}

	return rule, nil
}

// UpdateRuleWithAudit 更新规则并记录审计日志
func (s *ruleService) UpdateRuleWithAudit(ctx context.Context, id string, req *model.UpdateRuleRequest, userID, userName, ipAddress, userAgent string) (*model.Rule, error) {
	// 获取更新前的规则
	oldRule, err := s.ruleRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("rule not found: %w", err)
	}

	// 更新规则
	updatedRule, err := s.UpdateRule(ctx, id, req)
	if err != nil {
		return nil, err
	}

	// 如果版本服务可用，创建版本记录和审计日志
	if s.versionService != nil {
		// 创建版本记录
		if _, versionErr := s.versionService.CreateVersion(ctx, updatedRule, "update", userID, "Rule updated"); versionErr != nil {
			fmt.Printf("Failed to create version record: %v\n", versionErr)
		}

		// 创建审计日志
		auditLog := &model.RuleAuditLog{
			ID:         uuid.New().String(),
			RuleID:     updatedRule.ID,
			Action:     "update",
			Changes:    fmt.Sprintf(`{"action":"update","old_version":"%s","new_version":"%s"}`, oldRule.Version, updatedRule.Version),
			UserID:     userID,
			UserName:   userName,
			IPAddress:  ipAddress,
			UserAgent:  userAgent,
			Reason:     "Rule updated",
			CreatedAt:  time.Now(),
		}

		// 记录变更详情
		changes := s.buildChangeMap(oldRule, updatedRule, req)
		// Set changes as JSON string
		if changesJSON, err := json.Marshal(changes); err == nil {
			auditLog.Changes = string(changesJSON)
			if auditErr := s.versionService.CreateAuditLog(ctx, auditLog); auditErr != nil {
				fmt.Printf("Failed to create audit log: %v\n", auditErr)
			}
		}
	}

	return updatedRule, nil
}

// DeleteRuleWithAudit 删除规则并记录审计日志
func (s *ruleService) DeleteRuleWithAudit(ctx context.Context, id string, userID, userName, ipAddress, userAgent string) error {
	// 获取删除前的规则
	rule, err := s.ruleRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("rule not found: %w", err)
	}

	// 如果版本服务可用，创建版本记录和审计日志
	if s.versionService != nil {
		// 创建删除前的版本记录
		if _, versionErr := s.versionService.CreateVersion(ctx, rule, "delete", userID, "Rule deleted"); versionErr != nil {
			fmt.Printf("Failed to create version record: %v\n", versionErr)
		}

		// 创建审计日志
		auditLog := &model.RuleAuditLog{
			ID:         uuid.New().String(),
			RuleID:     rule.ID,
			Action:     "delete",
			Changes:    fmt.Sprintf(`{"action":"delete","old_version":"%s"}`, rule.Version),
			UserID:     userID,
			UserName:   userName,
			IPAddress:  ipAddress,
			UserAgent:  userAgent,
			Reason:     "Rule deleted",
			CreatedAt:  time.Now(),
		}

		changes := map[string]interface{}{
			"deleted_rule": map[string]interface{}{
				"name":       rule.Name,
				"expression": rule.Expression,
				"version":    rule.Version,
			},
		}

		// Set changes as JSON string
		if changesJSON, err := json.Marshal(changes); err == nil {
			auditLog.Changes = string(changesJSON)
			if auditErr := s.versionService.CreateAuditLog(ctx, auditLog); auditErr != nil {
				fmt.Printf("Failed to create audit log: %v\n", auditErr)
			}
		}
	}

	// 删除规则
	return s.DeleteRule(ctx, id)
}

// buildChangeMap 构建变更映射
func (s *ruleService) buildChangeMap(oldRule *model.Rule, newRule *model.Rule, req *model.UpdateRuleRequest) map[string]interface{} {
	changes := make(map[string]interface{})

	if req.Name != "" && oldRule.Name != newRule.Name {
		changes["name"] = map[string]interface{}{
			"old": oldRule.Name,
			"new": newRule.Name,
		}
	}

	if req.Expression != "" && oldRule.Expression != newRule.Expression {
		changes["expression"] = map[string]interface{}{
			"old": oldRule.Expression,
			"new": newRule.Expression,
		}
	}

	if req.Duration != "" && oldRule.Duration != newRule.Duration {
		changes["duration"] = map[string]interface{}{
			"old": oldRule.Duration,
			"new": newRule.Duration,
		}
	}

	if req.Severity != "" && oldRule.Severity != newRule.Severity {
		changes["severity"] = map[string]interface{}{
			"old": oldRule.Severity,
			"new": newRule.Severity,
		}
	}

	if req.Labels != nil {
		oldLabels, _ := oldRule.GetLabelsMap()
		newLabels, _ := newRule.GetLabelsMap()
		if !reflect.DeepEqual(oldLabels, newLabels) {
			changes["labels"] = map[string]interface{}{
				"old": oldLabels,
				"new": newLabels,
			}
		}
	}

	if req.Annotations != nil {
		oldAnnotations, _ := oldRule.GetAnnotationsMap()
		newAnnotations, _ := newRule.GetAnnotationsMap()
		if !reflect.DeepEqual(oldAnnotations, newAnnotations) {
			changes["annotations"] = map[string]interface{}{
				"old": oldAnnotations,
				"new": newAnnotations,
			}
		}
	}

	if req.Targets != nil {
		oldTargets, _ := oldRule.GetTargetsList()
		newTargets, _ := newRule.GetTargetsList()
		if !reflect.DeepEqual(oldTargets, newTargets) {
			changes["targets"] = map[string]interface{}{
				"old": oldTargets,
				"new": newTargets,
			}
		}
	}

	return changes
}