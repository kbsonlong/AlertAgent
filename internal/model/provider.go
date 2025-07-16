package model

import (
	"errors"
	"time"

	"gorm.io/gorm"
)

// Provider 类型常量
const (
	ProviderTypePrometheus      = "prometheus"
	ProviderTypeVictoriaMetrics = "victoriametrics"
)

// Provider 状态常量
const (
	ProviderStatusActive   = "active"
	ProviderStatusInactive = "inactive"
)

// Provider 数据源配置
type Provider struct {
	ID          uint           `json:"id" gorm:"primarykey"`
	CreatedAt   time.Time      `json:"-"`
	UpdatedAt   time.Time      `json:"-"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`
	Name        string         `json:"name" gorm:"size:255;not null;uniqueIndex"`
	Type        string         `json:"type" gorm:"type:varchar(50);not null"`
	Status      string         `json:"status" gorm:"type:varchar(50);not null;default:'active'"`
	Description string         `json:"description,omitempty" gorm:"type:text"`
	Endpoint    string         `json:"endpoint" gorm:"type:varchar(255);not null"`
	AuthType    string         `json:"auth_type,omitempty" gorm:"type:varchar(50);default:'none'"`
	AuthConfig  string         `json:"auth_config,omitempty" gorm:"type:text"`
	Labels      string         `json:"labels,omitempty" gorm:"type:text"`
	LastCheck   *time.Time     `json:"last_check,omitempty"`
	LastError   string         `json:"last_error,omitempty" gorm:"type:text"`
}

// Validate 验证数据源配置
func (p *Provider) Validate() error {
	if p.Name == "" {
		return errors.New("name is required")
	}

	if p.Type == "" {
		return errors.New("type is required")
	}

	if p.Type != ProviderTypePrometheus && p.Type != ProviderTypeVictoriaMetrics {
		return errors.New("invalid provider type")
	}

	if p.Endpoint == "" {
		return errors.New("endpoint is required")
	}

	return nil
}
