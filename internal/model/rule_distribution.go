package model

import (
	"time"

	"gorm.io/gorm"
)

// RuleDistributionRecord 规则分发记录
type RuleDistributionRecord struct {
	ID          string         `json:"id" gorm:"type:varchar(36);primaryKey"`
	RuleID      string         `json:"rule_id" gorm:"type:varchar(36);not null;index"`
	Target      string         `json:"target" gorm:"type:varchar(255);not null;index"`
	Status      string         `json:"status" gorm:"type:varchar(20);not null;default:'pending';index"`
	Version     string         `json:"version" gorm:"type:varchar(50);not null"`
	ConfigHash  string         `json:"config_hash" gorm:"type:varchar(64)"`
	LastSync    time.Time      `json:"last_sync"`
	Error       string         `json:"error" gorm:"type:text"`
	RetryCount  int            `json:"retry_count" gorm:"default:0"`
	MaxRetry    int            `json:"max_retry" gorm:"default:3"`
	NextRetry   *time.Time     `json:"next_retry"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`
}

// TableName 指定表名
func (RuleDistributionRecord) TableName() string {
	return "rule_distribution_records"
}

// BatchRuleOperation 批量规则操作
type BatchRuleOperation struct {
	Action  string   `json:"action" binding:"required"` // create, update, delete, distribute
	RuleIDs []string `json:"rule_ids" binding:"required"`
	Targets []string `json:"targets,omitempty"`
	Options map[string]interface{} `json:"options,omitempty"`
}

// BatchRuleOperationResult 批量规则操作结果
type BatchRuleOperationResult struct {
	TotalCount   int                           `json:"total_count"`
	SuccessCount int                           `json:"success_count"`
	FailedCount  int                           `json:"failed_count"`
	Results      []BatchRuleOperationItem      `json:"results"`
}

// BatchRuleOperationItem 批量规则操作项结果
type BatchRuleOperationItem struct {
	RuleID  string `json:"rule_id"`
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

// RuleDistributionSummary 规则分发汇总
type RuleDistributionSummary struct {
	RuleID       string                    `json:"rule_id"`
	RuleName     string                    `json:"rule_name"`
	Version      string                    `json:"version"`
	TotalTargets int                       `json:"total_targets"`
	SuccessCount int                       `json:"success_count"`
	FailedCount  int                       `json:"failed_count"`
	PendingCount int                       `json:"pending_count"`
	LastSync     time.Time                 `json:"last_sync"`
	Targets      []TargetDistributionInfo  `json:"targets"`
}

// TargetDistributionInfo 目标分发信息
type TargetDistributionInfo struct {
	Target      string     `json:"target"`
	Status      string     `json:"status"`
	Version     string     `json:"version"`
	ConfigHash  string     `json:"config_hash"`
	LastSync    time.Time  `json:"last_sync"`
	Error       string     `json:"error,omitempty"`
	RetryCount  int        `json:"retry_count"`
	NextRetry   *time.Time `json:"next_retry,omitempty"`
}

// RetryDistributionRequest 重试分发请求
type RetryDistributionRequest struct {
	RuleIDs []string `json:"rule_ids" binding:"required"`
	Targets []string `json:"targets,omitempty"` // 如果为空，重试所有失败的目标
	Force   bool     `json:"force"`             // 是否强制重试（忽略重试次数限制）
}

// RetryDistributionResult 重试分发结果
type RetryDistributionResult struct {
	TotalCount   int                      `json:"total_count"`
	RetryCount   int                      `json:"retry_count"`
	SkippedCount int                      `json:"skipped_count"`
	Results      []RetryDistributionItem  `json:"results"`
}

// RetryDistributionItem 重试分发项结果
type RetryDistributionItem struct {
	RuleID  string `json:"rule_id"`
	Target  string `json:"target"`
	Status  string `json:"status"` // retried, skipped, failed
	Reason  string `json:"reason,omitempty"`
}

// DistributionStatusFilter 分发状态过滤器
type DistributionStatusFilter struct {
	RuleIDs []string `json:"rule_ids,omitempty"`
	Targets []string `json:"targets,omitempty"`
	Status  []string `json:"status,omitempty"`
	Version string   `json:"version,omitempty"`
}

// IsRetryable 检查是否可以重试
func (r *RuleDistributionRecord) IsRetryable() bool {
	return r.Status == "failed" && r.RetryCount < r.MaxRetry
}

// ShouldRetryNow 检查是否应该立即重试
func (r *RuleDistributionRecord) ShouldRetryNow() bool {
	if !r.IsRetryable() {
		return false
	}
	if r.NextRetry == nil {
		return true
	}
	return time.Now().After(*r.NextRetry)
}

// CalculateNextRetry 计算下次重试时间
func (r *RuleDistributionRecord) CalculateNextRetry() time.Time {
	// 指数退避策略：2^retryCount 分钟
	backoffMinutes := 1 << r.RetryCount
	if backoffMinutes > 60 {
		backoffMinutes = 60 // 最大1小时
	}
	return time.Now().Add(time.Duration(backoffMinutes) * time.Minute)
}

// SetError 设置错误信息
func (r *RuleDistributionRecord) SetError(err error) {
	if err != nil {
		r.Error = err.Error()
		r.Status = "failed"
	} else {
		r.Error = ""
		r.Status = "success"
		r.LastSync = time.Now()
	}
	r.UpdatedAt = time.Now()
}

// IncrementRetry 增加重试次数
func (r *RuleDistributionRecord) IncrementRetry() {
	r.RetryCount++
	nextRetry := r.CalculateNextRetry()
	r.NextRetry = &nextRetry
	r.UpdatedAt = time.Now()
}