package model

import (
	"encoding/json"
	"errors"
)

// VictoriaMetricsConfig VictoriaMetrics数据源配置
type VictoriaMetricsConfig struct {
	BasicAuth *BasicAuth `json:"basic_auth,omitempty"`
	TLSConfig *TLSConfig `json:"tls_config,omitempty"`
	TenantID  string     `json:"tenant_id,omitempty"`
}

// VictoriaMetricsProvider VictoriaMetrics数据源
type VictoriaMetricsProvider struct {
	Provider
	Config VictoriaMetricsConfig
}

// NewVictoriaMetricsProvider 创建VictoriaMetrics数据源
func NewVictoriaMetricsProvider() *VictoriaMetricsProvider {
	return &VictoriaMetricsProvider{
		Provider: Provider{
			Type:   ProviderTypeVictoriaMetrics,
			Status: ProviderStatusActive,
		},
	}
}

// ValidateConfig 验证VictoriaMetrics配置
func (p *VictoriaMetricsProvider) ValidateConfig() error {
	if err := p.Provider.Validate(); err != nil {
		return err
	}

	// 解析AuthConfig
	if p.AuthType == "basic" {
		if p.AuthConfig == "" {
			return errors.New("auth config is required for basic auth")
		}

		var basicAuth BasicAuth
		if err := json.Unmarshal([]byte(p.AuthConfig), &basicAuth); err != nil {
			return errors.New("invalid basic auth config")
		}

		if basicAuth.Username == "" || basicAuth.Password == "" {
			return errors.New("username and password are required for basic auth")
		}
	}

	return nil
}
