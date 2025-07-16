package model

import (
	"encoding/json"
	"errors"
)

// PrometheusConfig Prometheus数据源配置
type PrometheusConfig struct {
	BasicAuth *BasicAuth `json:"basic_auth,omitempty"`
	TLSConfig *TLSConfig `json:"tls_config,omitempty"`
}

// BasicAuth 基础认证配置
type BasicAuth struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// TLSConfig TLS配置
type TLSConfig struct {
	CACert             string `json:"ca_cert,omitempty"`
	ServerName         string `json:"server_name,omitempty"`
	InsecureSkipVerify bool   `json:"insecure_skip_verify,omitempty"`
}

// PrometheusProvider Prometheus数据源
type PrometheusProvider struct {
	Provider
	Config PrometheusConfig
}

// NewPrometheusProvider 创建Prometheus数据源
func NewPrometheusProvider() *PrometheusProvider {
	return &PrometheusProvider{
		Provider: Provider{
			Type:   ProviderTypePrometheus,
			Status: ProviderStatusActive,
		},
	}
}

// ValidateConfig 验证Prometheus配置
func (p *PrometheusProvider) ValidateConfig() error {
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
