package model

import (
	"gorm.io/gorm"
)

// Rule 告警规则
type Rule struct {
	gorm.Model
	Name          string `json:"name" gorm:"type:varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;not null"`
	Description   string `json:"description" gorm:"type:text CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;not null"`
	Level         string `json:"level" gorm:"type:varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;not null"`
	Enabled       bool   `json:"enabled" gorm:"not null;default:true"`
	ProviderID    uint   `json:"provider_id" gorm:"not null"`
	ProviderID    uint   `json:"provider_id" gorm:"not null"`
	QueryExpr     string `json:"query_expr" gorm:"type:text CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;not null"`
	ConditionExpr string `json:"condition_expr" gorm:"type:text CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;not null"`
	NotifyType    string `json:"notify_type" gorm:"type:varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;not null"`
	NotifyGroup   string `json:"notify_group" gorm:"type:varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;not null"`
	Template      string `json:"template" gorm:"type:varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;not null"`
}
