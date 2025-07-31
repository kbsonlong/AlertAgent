package model

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// RuleVersion 规则版本历史
type RuleVersion struct {
	ID          string         `json:"id" gorm:"type:varchar(36);primaryKey"`
	RuleID      string         `json:"rule_id" gorm:"type:varchar(36);not null;index"`
	Version     string         `json:"version" gorm:"type:varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;not null"`
	Name        string         `json:"name" gorm:"type:varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;not null"`
	Expression  string         `json:"expression" gorm:"type:text CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;not null"`
	Duration    string         `json:"duration" gorm:"type:varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;not null"`
	Severity    string         `json:"severity" gorm:"type:varchar(20) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;not null"`
	Labels      string         `json:"labels" gorm:"type:json"`
	Annotations string         `json:"annotations" gorm:"type:json"`
	Targets     string         `json:"targets" gorm:"type:json"`
	ChangeLog   string         `json:"change_log" gorm:"type:text CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci"`
	CreatedBy   string         `json:"created_by" gorm:"type:varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci"`
	CreatedAt   time.Time      `json:"created_at"`
}

// TableName 指定表名
func (RuleVersion) TableName() string {
	return "rule_versions"
}

// BeforeCreate GORM钩子：创建前生成ID
func (r *RuleVersion) BeforeCreate(tx *gorm.DB) error {
	if r.ID == "" {
		r.ID = uuid.New().String()
	}
	return nil
}

// GetLabelsMap 获取标签映射
func (r *RuleVersion) GetLabelsMap() (map[string]string, error) {
	if r.Labels == "" {
		return make(map[string]string), nil
	}
	var labels map[string]string
	err := json.Unmarshal([]byte(r.Labels), &labels)
	return labels, err
}

// SetLabelsMap 设置标签映射
func (r *RuleVersion) SetLabelsMap(labels map[string]string) error {
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
func (r *RuleVersion) GetAnnotationsMap() (map[string]string, error) {
	if r.Annotations == "" {
		return make(map[string]string), nil
	}
	var annotations map[string]string
	err := json.Unmarshal([]byte(r.Annotations), &annotations)
	return annotations, err
}

// SetAnnotationsMap 设置注释映射
func (r *RuleVersion) SetAnnotationsMap(annotations map[string]string) error {
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
func (r *RuleVersion) GetTargetsList() ([]string, error) {
	if r.Targets == "" {
		return make([]string, 0), nil
	}
	var targets []string
	err := json.Unmarshal([]byte(r.Targets), &targets)
	return targets, err
}

// SetTargetsList 设置目标列表
func (r *RuleVersion) SetTargetsList(targets []string) error {
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

// ToRule 转换为Rule对象
func (r *RuleVersion) ToRule() *Rule {
	return &Rule{
		ID:          r.RuleID,
		Name:        r.Name,
		Expression:  r.Expression,
		Duration:    r.Duration,
		Severity:    r.Severity,
		Labels:      r.Labels,
		Annotations: r.Annotations,
		Targets:     r.Targets,
		Version:     r.Version,
		CreatedAt:   r.CreatedAt,
		UpdatedAt:   r.CreatedAt,
	}
}



// BeforeCreate GORM钩子：创建前生成ID
func (r *RuleDistributionRecord) BeforeCreate(tx *gorm.DB) error {
	if r.ID == "" {
		r.ID = uuid.New().String()
	}
	return nil
}

// CanRetry 检查是否可以重试
func (r *RuleDistributionRecord) CanRetry() bool {
	return r.RetryCount < r.MaxRetry && r.Status == "failed"
}

// ShouldRetry 检查是否应该重试
func (r *RuleDistributionRecord) ShouldRetry() bool {
	if !r.CanRetry() {
		return false
	}
	if r.NextRetry == nil {
		return true
	}
	return time.Now().After(*r.NextRetry)
}



// RuleVersionComparison 规则版本比较
type RuleVersionComparison struct {
	Field    string      `json:"field"`
	OldValue interface{} `json:"old_value"`
	NewValue interface{} `json:"new_value"`
	Changed  bool        `json:"changed"`
}

// CompareVersions 比较两个规则版本
func CompareVersions(oldVersion, newVersion *RuleVersion) []RuleVersionComparison {
	comparisons := []RuleVersionComparison{
		{
			Field:    "name",
			OldValue: oldVersion.Name,
			NewValue: newVersion.Name,
			Changed:  oldVersion.Name != newVersion.Name,
		},
		{
			Field:    "expression",
			OldValue: oldVersion.Expression,
			NewValue: newVersion.Expression,
			Changed:  oldVersion.Expression != newVersion.Expression,
		},
		{
			Field:    "duration",
			OldValue: oldVersion.Duration,
			NewValue: newVersion.Duration,
			Changed:  oldVersion.Duration != newVersion.Duration,
		},
		{
			Field:    "severity",
			OldValue: oldVersion.Severity,
			NewValue: newVersion.Severity,
			Changed:  oldVersion.Severity != newVersion.Severity,
		},
		{
			Field:    "labels",
			OldValue: oldVersion.Labels,
			NewValue: newVersion.Labels,
			Changed:  oldVersion.Labels != newVersion.Labels,
		},
		{
			Field:    "annotations",
			OldValue: oldVersion.Annotations,
			NewValue: newVersion.Annotations,
			Changed:  oldVersion.Annotations != newVersion.Annotations,
		},
		{
			Field:    "targets",
			OldValue: oldVersion.Targets,
			NewValue: newVersion.Targets,
			Changed:  oldVersion.Targets != newVersion.Targets,
		},
	}
	
	return comparisons
}

// CreateRuleVersionRequest 创建规则版本请求
type CreateRuleVersionRequest struct {
	RuleID      string            `json:"rule_id" binding:"required"`
	Version     string            `json:"version" binding:"required"`
	Name        string            `json:"name" binding:"required"`
	Expression  string            `json:"expression" binding:"required"`
	Duration    string            `json:"duration" binding:"required"`
	Severity    string            `json:"severity" binding:"required"`
	Labels      map[string]string `json:"labels"`
	Annotations map[string]string `json:"annotations"`
	Targets     []string          `json:"targets"`
	ChangeLog   string            `json:"change_log"`
	CreatedBy   string            `json:"created_by"`
}

// RuleVersionQueryRequest 规则版本查询请求
type RuleVersionQueryRequest struct {
	RuleID   string `form:"rule_id"`
	Version  string `form:"version"`
	Page     int    `form:"page" binding:"min=1"`
	PageSize int    `form:"page_size" binding:"min=1,max=100"`
}

// RuleDistributionQueryRequest 规则分发查询请求
type RuleDistributionQueryRequest struct {
	RuleID   string `form:"rule_id"`
	Target   string `form:"target"`
	Status   string `form:"status"`
	Page     int    `form:"page" binding:"min=1"`
	PageSize int    `form:"page_size" binding:"min=1,max=100"`
}