package model

import (
	"encoding/json"
	"time"

	"gorm.io/gorm"
)

// Rule 告警规则 - 重构后的增强版本
type Rule struct {
	ID          string            `json:"id" gorm:"type:varchar(36);primaryKey"`
	Name        string            `json:"name" gorm:"type:varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;not null;index"`
	Expression  string            `json:"expression" gorm:"type:text CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;not null"`
	Duration    string            `json:"duration" gorm:"type:varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;not null"`
	Severity    string            `json:"severity" gorm:"type:varchar(20) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;not null"`
	Labels      string            `json:"labels" gorm:"type:json"`
	Annotations string            `json:"annotations" gorm:"type:json"`
	Targets     string            `json:"targets" gorm:"type:json"`
	Version     string            `json:"version" gorm:"type:varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;not null;default:'v1.0.0'"`
	Status      string            `json:"status" gorm:"type:varchar(20) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;not null;default:'pending';index"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
	DeletedAt   gorm.DeletedAt    `json:"-" gorm:"index"`
}

// GetLabelsMap 获取标签映射
func (r *Rule) GetLabelsMap() (map[string]string, error) {
	if r.Labels == "" {
		return make(map[string]string), nil
	}
	var labels map[string]string
	err := json.Unmarshal([]byte(r.Labels), &labels)
	return labels, err
}

// SetLabelsMap 设置标签映射
func (r *Rule) SetLabelsMap(labels map[string]string) error {
	if labels == nil {
		r.Labels = ""
		return nil
	}
	data, err := json.Marshal(labels)
	if err != nil {
		return err
	}
	r.Labels = string(data)
	return nil
}

// GetAnnotationsMap 获取注释映射
func (r *Rule) GetAnnotationsMap() (map[string]string, error) {
	if r.Annotations == "" {
		return make(map[string]string), nil
	}
	var annotations map[string]string
	err := json.Unmarshal([]byte(r.Annotations), &annotations)
	return annotations, err
}

// SetAnnotationsMap 设置注释映射
func (r *Rule) SetAnnotationsMap(annotations map[string]string) error {
	if annotations == nil {
		r.Annotations = ""
		return nil
	}
	data, err := json.Marshal(annotations)
	if err != nil {
		return err
	}
	r.Annotations = string(data)
	return nil
}

// GetTargetsList 获取目标列表
func (r *Rule) GetTargetsList() ([]string, error) {
	if r.Targets == "" {
		return make([]string, 0), nil
	}
	var targets []string
	err := json.Unmarshal([]byte(r.Targets), &targets)
	return targets, err
}

// SetTargetsList 设置目标列表
func (r *Rule) SetTargetsList(targets []string) error {
	if targets == nil {
		r.Targets = ""
		return nil
	}
	data, err := json.Marshal(targets)
	if err != nil {
		return err
	}
	r.Targets = string(data)
	return nil
}

// TableName 指定表名
func (Rule) TableName() string {
	return "alert_rules"
}

// CreateRuleRequest 创建规则请求
type CreateRuleRequest struct {
	Name        string            `json:"name" binding:"required"`
	Expression  string            `json:"expression" binding:"required"`
	Duration    string            `json:"duration" binding:"required"`
	Severity    string            `json:"severity" binding:"required"`
	Labels      map[string]string `json:"labels"`
	Annotations map[string]string `json:"annotations"`
	Targets     []string          `json:"targets"`
}

// UpdateRuleRequest 更新规则请求
type UpdateRuleRequest struct {
	Name        string            `json:"name"`
	Expression  string            `json:"expression"`
	Duration    string            `json:"duration"`
	Severity    string            `json:"severity"`
	Labels      map[string]string `json:"labels"`
	Annotations map[string]string `json:"annotations"`
	Targets     []string          `json:"targets"`
}

// RuleDistributionStatus 规则分发状态
type RuleDistributionStatus struct {
	RuleID      string                   `json:"rule_id"`
	RuleName    string                   `json:"rule_name"`
	Version     string                   `json:"version"`
	Targets     []string                 `json:"targets"`
	Status      string                   `json:"status"`
	LastSync    time.Time                `json:"last_sync"`
	TargetStatus []TargetDistributionStatus `json:"target_status"`
}

// TargetDistributionStatus 目标分发状态
type TargetDistributionStatus struct {
	Target    string    `json:"target"`
	Status    string    `json:"status"`
	LastSync  time.Time `json:"last_sync"`
	Error     string    `json:"error,omitempty"`
}
