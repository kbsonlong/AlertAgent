package service

import (
	"context"
	"fmt"
	"time"

	"alert_agent/internal/model"
	"alert_agent/internal/repository"

	"github.com/google/uuid"
)

// RuleDistributionService 规则分发服务接口
type RuleDistributionService interface {
	// 分发状态查询
	GetDistributionStatus(ctx context.Context, ruleID string) (*model.RuleDistributionSummary, error)
	GetDistributionSummary(ctx context.Context, ruleIDs []string) ([]*model.RuleDistributionSummary, error)
	GetTargetDistributionInfo(ctx context.Context, ruleID, target string) (*model.TargetDistributionInfo, error)
	
	// 批量操作
	BatchDistributeRules(ctx context.Context, req *model.BatchRuleOperation) (*model.BatchRuleOperationResult, error)
	BatchUpdateDistributionStatus(ctx context.Context, ruleIDs []string, targets []string, status string) error
	
	// 重试机制
	RetryFailedDistributions(ctx context.Context, req *model.RetryDistributionRequest) (*model.RetryDistributionResult, error)
	GetRetryableDistributions(ctx context.Context, limit int) ([]*model.RuleDistributionRecord, error)
	ProcessRetryableDistributions(ctx context.Context) error
	
	// 分发记录管理
	CreateDistributionRecord(ctx context.Context, ruleID, target, version string) (*model.RuleDistributionRecord, error)
	UpdateDistributionRecord(ctx context.Context, recordID string, status string, err error) error
	DeleteDistributionRecords(ctx context.Context, ruleID string) error
}

// ruleDistributionService 规则分发服务实现
type ruleDistributionService struct {
	distributionRepo repository.RuleDistributionRepository
	ruleRepo         repository.RuleRepository
}

// NewRuleDistributionService 创建规则分发服务实例
func NewRuleDistributionService(
	distributionRepo repository.RuleDistributionRepository,
	ruleRepo repository.RuleRepository,
) RuleDistributionService {
	return &ruleDistributionService{
		distributionRepo: distributionRepo,
		ruleRepo:         ruleRepo,
	}
}

// GetDistributionStatus 获取单个规则的分发状态
func (s *ruleDistributionService) GetDistributionStatus(ctx context.Context, ruleID string) (*model.RuleDistributionSummary, error) {
	summaries, err := s.distributionRepo.GetDistributionSummary(ctx, []string{ruleID})
	if err != nil {
		return nil, fmt.Errorf("failed to get distribution status: %w", err)
	}
	
	if len(summaries) == 0 {
		// 如果没有分发记录，检查规则是否存在
		rule, err := s.ruleRepo.GetByID(ctx, ruleID)
		if err != nil {
			return nil, fmt.Errorf("rule not found: %w", err)
		}
		
		// 返回基本信息，没有分发记录
		return &model.RuleDistributionSummary{
			RuleID:       rule.ID,
			RuleName:     rule.Name,
			Version:      rule.Version,
			TotalTargets: 0,
			SuccessCount: 0,
			FailedCount:  0,
			PendingCount: 0,
			LastSync:     rule.UpdatedAt,
			Targets:      []model.TargetDistributionInfo{},
		}, nil
	}
	
	return summaries[0], nil
}

// GetDistributionSummary 获取多个规则的分发汇总
func (s *ruleDistributionService) GetDistributionSummary(ctx context.Context, ruleIDs []string) ([]*model.RuleDistributionSummary, error) {
	return s.distributionRepo.GetDistributionSummary(ctx, ruleIDs)
}

// GetTargetDistributionInfo 获取特定目标的分发信息
func (s *ruleDistributionService) GetTargetDistributionInfo(ctx context.Context, ruleID, target string) (*model.TargetDistributionInfo, error) {
	record, err := s.distributionRepo.GetByRuleIDAndTarget(ctx, ruleID, target)
	if err != nil {
		return nil, fmt.Errorf("failed to get target distribution info: %w", err)
	}
	
	return &model.TargetDistributionInfo{
		Target:     record.Target,
		Status:     record.Status,
		Version:    record.Version,
		ConfigHash: record.ConfigHash,
		LastSync:   record.LastSync,
		Error:      record.Error,
		RetryCount: record.RetryCount,
		NextRetry:  record.NextRetry,
	}, nil
}

// BatchDistributeRules 批量分发规则
func (s *ruleDistributionService) BatchDistributeRules(ctx context.Context, req *model.BatchRuleOperation) (*model.BatchRuleOperationResult, error) {
	result := &model.BatchRuleOperationResult{
		TotalCount: len(req.RuleIDs),
		Results:    make([]model.BatchRuleOperationItem, 0, len(req.RuleIDs)),
	}
	
	for _, ruleID := range req.RuleIDs {
		item := model.BatchRuleOperationItem{
			RuleID:  ruleID,
			Success: false,
		}
		
		switch req.Action {
		case "distribute":
			err := s.distributeRule(ctx, ruleID, req.Targets)
			if err != nil {
				item.Error = err.Error()
				result.FailedCount++
			} else {
				item.Success = true
				result.SuccessCount++
			}
			
		case "delete":
			err := s.DeleteDistributionRecords(ctx, ruleID)
			if err != nil {
				item.Error = err.Error()
				result.FailedCount++
			} else {
				item.Success = true
				result.SuccessCount++
			}
			
		default:
			item.Error = fmt.Sprintf("unsupported action: %s", req.Action)
			result.FailedCount++
		}
		
		result.Results = append(result.Results, item)
	}
	
	return result, nil
}

// distributeRule 分发单个规则
func (s *ruleDistributionService) distributeRule(ctx context.Context, ruleID string, targets []string) error {
	// 获取规则信息
	rule, err := s.ruleRepo.GetByID(ctx, ruleID)
	if err != nil {
		return fmt.Errorf("rule not found: %w", err)
	}
	
	// 如果没有指定目标，从规则中获取
	if len(targets) == 0 {
		ruleTargets, err := rule.GetTargetsList()
		if err != nil {
			return fmt.Errorf("failed to get rule targets: %w", err)
		}
		targets = ruleTargets
	}
	
	// 为每个目标创建或更新分发记录
	for _, target := range targets {
		_, err := s.CreateDistributionRecord(ctx, ruleID, target, rule.Version)
		if err != nil {
			return fmt.Errorf("failed to create distribution record for target %s: %w", target, err)
		}
	}
	
	return nil
}

// BatchUpdateDistributionStatus 批量更新分发状态
func (s *ruleDistributionService) BatchUpdateDistributionStatus(ctx context.Context, ruleIDs []string, targets []string, status string) error {
	return s.distributionRepo.BatchUpdateStatus(ctx, ruleIDs, targets, status)
}

// RetryFailedDistributions 重试失败的分发
func (s *ruleDistributionService) RetryFailedDistributions(ctx context.Context, req *model.RetryDistributionRequest) (*model.RetryDistributionResult, error) {
	result := &model.RetryDistributionResult{
		Results: make([]model.RetryDistributionItem, 0),
	}
	
	for _, ruleID := range req.RuleIDs {
		// 获取该规则的分发记录
		records, err := s.distributionRepo.ListByRuleID(ctx, ruleID)
		if err != nil {
			continue
		}
		
		for _, record := range records {
			// 检查是否匹配目标过滤条件
			if len(req.Targets) > 0 {
				found := false
				for _, target := range req.Targets {
					if record.Target == target {
						found = true
						break
					}
				}
				if !found {
					continue
				}
			}
			
			result.TotalCount++
			
			item := model.RetryDistributionItem{
				RuleID: ruleID,
				Target: record.Target,
			}
			
			// 检查是否可以重试
			if !req.Force && !record.IsRetryable() {
				item.Status = "skipped"
				if record.RetryCount >= record.MaxRetry {
					item.Reason = "exceeded max retry count"
				} else {
					item.Reason = "not in failed status"
				}
				result.SkippedCount++
			} else {
				// 重置重试状态
				record.Status = "pending"
				record.Error = ""
				if req.Force {
					record.RetryCount = 0
					record.NextRetry = nil
				} else {
					record.IncrementRetry()
				}
				
				if err := s.distributionRepo.Update(ctx, record); err != nil {
					item.Status = "failed"
					item.Reason = err.Error()
				} else {
					item.Status = "retried"
					result.RetryCount++
				}
			}
			
			result.Results = append(result.Results, item)
		}
	}
	
	return result, nil
}

// GetRetryableDistributions 获取可重试的分发记录
func (s *ruleDistributionService) GetRetryableDistributions(ctx context.Context, limit int) ([]*model.RuleDistributionRecord, error) {
	return s.distributionRepo.ListRetryable(ctx, limit)
}

// ProcessRetryableDistributions 处理可重试的分发记录
func (s *ruleDistributionService) ProcessRetryableDistributions(ctx context.Context) error {
	// 获取可重试的记录
	records, err := s.GetRetryableDistributions(ctx, 100) // 限制每次处理100条
	if err != nil {
		return fmt.Errorf("failed to get retryable distributions: %w", err)
	}
	
	for _, record := range records {
		if !record.ShouldRetryNow() {
			continue
		}
		
		// 重置状态为pending，等待重新分发
		record.Status = "pending"
		record.IncrementRetry()
		
		if err := s.distributionRepo.Update(ctx, record); err != nil {
			// 记录错误但继续处理其他记录
			fmt.Printf("Failed to update distribution record %s: %v\n", record.ID, err)
		}
	}
	
	return nil
}

// CreateDistributionRecord 创建分发记录
func (s *ruleDistributionService) CreateDistributionRecord(ctx context.Context, ruleID, target, version string) (*model.RuleDistributionRecord, error) {
	// 检查是否已存在记录
	existingRecord, err := s.distributionRepo.GetByRuleIDAndTarget(ctx, ruleID, target)
	if err == nil {
		// 更新现有记录
		existingRecord.Version = version
		existingRecord.Status = "pending"
		existingRecord.Error = ""
		existingRecord.UpdatedAt = time.Now()
		
		if err := s.distributionRepo.Update(ctx, existingRecord); err != nil {
			return nil, fmt.Errorf("failed to update existing distribution record: %w", err)
		}
		return existingRecord, nil
	}
	
	// 创建新记录
	record := &model.RuleDistributionRecord{
		ID:         uuid.New().String(),
		RuleID:     ruleID,
		Target:     target,
		Status:     "pending",
		Version:    version,
		RetryCount: 0,
		MaxRetry:   3,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
	
	if err := s.distributionRepo.Create(ctx, record); err != nil {
		return nil, fmt.Errorf("failed to create distribution record: %w", err)
	}
	
	return record, nil
}

// UpdateDistributionRecord 更新分发记录
func (s *ruleDistributionService) UpdateDistributionRecord(ctx context.Context, recordID string, status string, err error) error {
	record, getErr := s.distributionRepo.GetByID(ctx, recordID)
	if getErr != nil {
		return fmt.Errorf("failed to get distribution record: %w", getErr)
	}
	
	record.Status = status
	record.SetError(err)
	
	if updateErr := s.distributionRepo.Update(ctx, record); updateErr != nil {
		return fmt.Errorf("failed to update distribution record: %w", updateErr)
	}
	
	return nil
}

// DeleteDistributionRecords 删除规则的所有分发记录
func (s *ruleDistributionService) DeleteDistributionRecords(ctx context.Context, ruleID string) error {
	return s.distributionRepo.DeleteByRuleID(ctx, ruleID)
}