package model

import "time"

// RuleVersionCompareRequest 规则版本比较请求
type RuleVersionCompareRequest struct {
	OldVersionID string `json:"old_version_id" binding:"required"`
	NewVersionID string `json:"new_version_id" binding:"required"`
}

// RuleVersionCompareResponse 规则版本比较响应
type RuleVersionCompareResponse struct {
	OldVersion  *RuleVersion            `json:"old_version"`
	NewVersion  *RuleVersion            `json:"new_version"`
	Differences []RuleVersionComparison `json:"differences"`
	Summary     struct {
		TotalChanges int `json:"total_changes"`
		HasChanges   bool `json:"has_changes"`
	} `json:"summary"`
}

// RuleRollbackRequest 规则回滚请求
type RuleRollbackRequest struct {
	RuleID    string `json:"rule_id" binding:"required"`
	VersionID string `json:"version_id" binding:"required"`
	ToVersion string `json:"to_version" binding:"required"`
	Reason    string `json:"reason"`
	Note      string `json:"note"`
}

// RuleAuditLogListRequest 规则审计日志列表请求
type RuleAuditLogListRequest struct {
	RuleID    string     `json:"rule_id"`
	Action    string     `json:"action"`
	UserID    string     `json:"user_id"`
	StartTime *time.Time `json:"start_time"`
	EndTime   *time.Time `json:"end_time"`
	Page      int        `json:"page"`
	PageSize  int        `json:"page_size"`
}

// RuleVersionComparison 规则版本比较 - 已在 rule_version.go 中定义

// RuleVersionDifference 规则版本差异 (别名，保持兼容性)
type RuleVersionDifference = RuleVersionComparison