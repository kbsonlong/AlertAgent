package model

import (
	"encoding/json"
	"time"

	"gorm.io/gorm"
)

// RuleVersion 规则版本记录
type RuleVersion struct {
	ID          string         `json:"id" gorm:"type:varchar(36);primaryKey"`
	RuleID      string         `json:"rule_id" gorm:"type:varchar(36);not null;index"`
	Version     string         `json:"version" gorm:"type:varchar(50);not null"`
	Name        string         `json:"name" gorm:"type:varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;not null"`
	Expression  string         `json:"expression" gorm:"type:text CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;not null"`
	Duration    string         `json:"duration" gorm:"type:varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;not null"`
	Severity    string         `json:"severity" gorm:"type:varchar(20) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;not null"`
	Labels      string         `json:"labels" gorm:"type:json"`
	Annotations string         `json:"annotations" gorm:"type:json"`
	Targets     string         `json:"targets" gorm:"type:json"`
	Status      string         `json:"status" gorm:"type:varchar(20) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;not null"`
	ChangeType  string         `json:"change_type" gorm:"type:varchar(20) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;not null"` // create, update, delete
	ChangedBy   string         `json:"changed_by" gorm:"type:varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci"`
	ChangeNote  string         `json:"change_note" gorm:"type:text CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci"`
	CreatedAt   time.Time      `json:"created_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`
}

// GetLabelsMap 获取标签映射
func (rv *RuleVersion) GetLabelsMap() (map[string]string, error) {
	if rv.Labels == "" {
		return make(map[string]string), nil
	}
	var labels map[string]string
	err := json.Unmarshal([]byte(rv.Labels), &labels)
	return labels, err
}

// SetLabelsMap 设置标签映射
func (rv *RuleVersion) SetLabelsMap(labels map[string]string) error {
	if labels == nil {
		rv.Labels = ""
		return nil
	}
	data, err := json.Marshal(labels)
	if err != nil {
		return err
	}
	rv.Labels = string(data)
	return nil
}

// GetAnnotationsMap 获取注释映射
func (rv *RuleVersion) GetAnnotationsMap() (map[string]string, error) {
	if rv.Annotations == "" {
		return make(map[string]string), nil
	}
	var annotations map[string]string
	err := json.Unmarshal([]byte(rv.Annotations), &annotations)
	return annotations, err
}

// SetAnnotationsMap 设置注释映射
func (rv *RuleVersion) SetAnnotationsMap(annotations map[string]string) error {
	if annotations == nil {
		rv.Annotations = ""
		return nil
	}
	data, err := json.Marshal(annotations)
	if err != nil {
		return err
	}
	rv.Annotations = string(data)
	return nil
}

// GetTargetsList 获取目标列表
func (rv *RuleVersion) GetTargetsList() ([]string, error) {
	if rv.Targets == "" {
		return make([]string, 0), nil
	}
	var targets []string
	err := json.Unmarshal([]byte(rv.Targets), &targets)
	return targets, err
}

// SetTargetsList 设置目标列表
func (rv *RuleVersion) SetTargetsList(targets []string) error {
	if targets == nil {
		rv.Targets = ""
		return nil
	}
	data, err := json.Marshal(targets)
	if err != nil {
		return err
	}
	rv.Targets = string(data)
	return nil
}

// TableName 指定表名
func (RuleVersion) TableName() string {
	return "rule_versions"
}

// ToRule 转换为规则对象
func (rv *RuleVersion) ToRule() *Rule {
	return &Rule{
		ID:          rv.RuleID,
		Name:        rv.Name,
		Expression:  rv.Expression,
		Duration:    rv.Duration,
		Severity:    rv.Severity,
		Labels:      rv.Labels,
		Annotations: rv.Annotations,
		Targets:     rv.Targets,
		Version:     rv.Version,
		Status:      rv.Status,
		CreatedAt:   rv.CreatedAt,
		UpdatedAt:   rv.CreatedAt,
	}
}

// RuleAuditLog 规则变更审计日志
type RuleAuditLog struct {
	ID          string         `json:"id" gorm:"type:varchar(36);primaryKey"`
	RuleID      string         `json:"rule_id" gorm:"type:varchar(36);not null;index"`
	Action      string         `json:"action" gorm:"type:varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;not null"` // create, update, delete, rollback
	OldVersion  string         `json:"old_version" gorm:"type:varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci"`
	NewVersion  string         `json:"new_version" gorm:"type:varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci"`
	Changes     string         `json:"changes" gorm:"type:json"` // 变更详情
	UserID      string         `json:"user_id" gorm:"type:varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci"`
	UserName    string         `json:"user_name" gorm:"type:varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci"`
	IPAddress   string         `json:"ip_address" gorm:"type:varchar(45) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci"`
	UserAgent   string         `json:"user_agent" gorm:"type:text CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci"`
	Note        string         `json:"note" gorm:"type:text CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci"`
	CreatedAt   time.Time      `json:"created_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`
}

// GetChangesMap 获取变更详情映射
func (ral *RuleAuditLog) GetChangesMap() (map[string]interface{}, error) {
	if ral.Changes == "" {
		return make(map[string]interface{}), nil
	}
	var changes map[string]interface{}
	err := json.Unmarshal([]byte(ral.Changes), &changes)
	return changes, err
}

// SetChangesMap 设置变更详情映射
func (ral *RuleAuditLog) SetChangesMap(changes map[string]interface{}) error {
	if changes == nil {
		ral.Changes = ""
		return nil
	}
	data, err := json.Marshal(changes)
	if err != nil {
		return err
	}
	ral.Changes = string(data)
	return nil
}

// TableName 指定表名
func (RuleAuditLog) TableName() string {
	return "rule_audit_logs"
}

// RuleVersionCompareRequest 规则版本对比请求
type RuleVersionCompareRequest struct {
	RuleID      string `json:"rule_id" binding:"required"`
	OldVersion  string `json:"old_version" binding:"required"`
	NewVersion  string `json:"new_version" binding:"required"`
}

// RuleVersionCompareResponse 规则版本对比响应
type RuleVersionCompareResponse struct {
	RuleID      string                    `json:"rule_id"`
	OldVersion  *RuleVersion              `json:"old_version"`
	NewVersion  *RuleVersion              `json:"new_version"`
	Differences []RuleVersionDifference   `json:"differences"`
}

// RuleVersionDifference 规则版本差异
type RuleVersionDifference struct {
	Field    string      `json:"field"`
	OldValue interface{} `json:"old_value"`
	NewValue interface{} `json:"new_value"`
	Type     string      `json:"type"` // added, removed, modified
}

// RuleRollbackRequest 规则回滚请求
type RuleRollbackRequest struct {
	RuleID     string `json:"rule_id" binding:"required"`
	ToVersion  string `json:"to_version" binding:"required"`
	Note       string `json:"note"`
}

// RuleVersionListRequest 规则版本列表请求
type RuleVersionListRequest struct {
	RuleID   string `json:"rule_id" binding:"required"`
	Page     int    `json:"page"`
	PageSize int    `json:"page_size"`
}

// RuleVersionListResponse 规则版本列表响应
type RuleVersionListResponse struct {
	Versions []*RuleVersion `json:"versions"`
	Total    int64          `json:"total"`
	Page     int            `json:"page"`
	PageSize int            `json:"page_size"`
}

// RuleAuditLogListRequest 规则审计日志列表请求
type RuleAuditLogListRequest struct {
	RuleID   string `json:"rule_id"`
	Action   string `json:"action"`
	UserID   string `json:"user_id"`
	Page     int    `json:"page"`
	PageSize int    `json:"page_size"`
}

// RuleAuditLogListResponse 规则审计日志列表响应
type RuleAuditLogListResponse struct {
	Logs     []*RuleAuditLog `json:"logs"`
	Total    int64           `json:"total"`
	Page     int             `json:"page"`
	PageSize int             `json:"page_size"`
}