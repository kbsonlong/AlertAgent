package sidecar

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"alert_agent/internal/pkg/logger"

	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
)

// TargetIntegration 目标系统集成接口
type TargetIntegration interface {
	// ValidateConfig 验证配置格式
	ValidateConfig(config []byte) error
	// TriggerReload 触发重载
	TriggerReload(ctx context.Context, reloadURL string) error
	// GetConfigType 获取配置类型
	GetConfigType() string
}

// PrometheusIntegration Prometheus集成
type PrometheusIntegration struct {
	httpClient *http.Client
}

// NewPrometheusIntegration 创建Prometheus集成
func NewPrometheusIntegration() *PrometheusIntegration {
	return &PrometheusIntegration{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (p *PrometheusIntegration) GetConfigType() string {
	return "prometheus"
}

func (p *PrometheusIntegration) ValidateConfig(config []byte) error {
	// 解析YAML格式
	var prometheusConfig struct {
		Groups []struct {
			Name  string `yaml:"name"`
			Rules []struct {
				Alert       string            `yaml:"alert,omitempty"`
				Expr        string            `yaml:"expr"`
				For         string            `yaml:"for,omitempty"`
				Labels      map[string]string `yaml:"labels,omitempty"`
				Annotations map[string]string `yaml:"annotations,omitempty"`
			} `yaml:"rules"`
		} `yaml:"groups"`
	}

	if err := yaml.Unmarshal(config, &prometheusConfig); err != nil {
		return fmt.Errorf("invalid YAML format: %w", err)
	}

	// 验证必需字段
	if len(prometheusConfig.Groups) == 0 {
		return fmt.Errorf("no rule groups found")
	}

	for i, group := range prometheusConfig.Groups {
		if group.Name == "" {
			return fmt.Errorf("group %d: name is required", i)
		}

		if len(group.Rules) == 0 {
			logger.L.Warn("Empty rule group found", zap.String("group", group.Name))
			continue
		}

		for j, rule := range group.Rules {
			if rule.Expr == "" {
				return fmt.Errorf("group %s, rule %d: expr is required", group.Name, j)
			}

			// 验证告警规则必须有alert字段
			if rule.Alert != "" && rule.Alert == "" {
				return fmt.Errorf("group %s, rule %d: alert name is required for alert rules", group.Name, j)
			}
		}
	}

	logger.L.Debug("Prometheus config validation passed",
		zap.Int("groups", len(prometheusConfig.Groups)),
	)

	return nil
}

func (p *PrometheusIntegration) TriggerReload(ctx context.Context, reloadURL string) error {
	// Prometheus reload API: POST /-/reload
	req, err := http.NewRequestWithContext(ctx, "POST", reloadURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create reload request: %w", err)
	}

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send reload request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("prometheus reload failed with status %d: %s", resp.StatusCode, string(body))
	}

	logger.L.Info("Prometheus reload triggered successfully")
	return nil
}

// AlertmanagerIntegration Alertmanager集成
type AlertmanagerIntegration struct {
	httpClient *http.Client
}

// NewAlertmanagerIntegration 创建Alertmanager集成
func NewAlertmanagerIntegration() *AlertmanagerIntegration {
	return &AlertmanagerIntegration{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (a *AlertmanagerIntegration) GetConfigType() string {
	return "alertmanager"
}

func (a *AlertmanagerIntegration) ValidateConfig(config []byte) error {
	// 解析Alertmanager配置
	var alertmanagerConfig struct {
		Global struct {
			SMTPSmarthost string `yaml:"smtp_smarthost,omitempty"`
			SMTPFrom      string `yaml:"smtp_from,omitempty"`
		} `yaml:"global,omitempty"`
		Route struct {
			GroupBy        []string `yaml:"group_by,omitempty"`
			GroupWait      string   `yaml:"group_wait,omitempty"`
			GroupInterval  string   `yaml:"group_interval,omitempty"`
			RepeatInterval string   `yaml:"repeat_interval,omitempty"`
			Receiver       string   `yaml:"receiver"`
			Routes         []interface{} `yaml:"routes,omitempty"`
		} `yaml:"route"`
		Receivers []struct {
			Name         string        `yaml:"name"`
			EmailConfigs []interface{} `yaml:"email_configs,omitempty"`
			WebhookConfigs []interface{} `yaml:"webhook_configs,omitempty"`
		} `yaml:"receivers"`
	}

	if err := yaml.Unmarshal(config, &alertmanagerConfig); err != nil {
		return fmt.Errorf("invalid YAML format: %w", err)
	}

	// 验证必需字段
	if alertmanagerConfig.Route.Receiver == "" {
		return fmt.Errorf("route.receiver is required")
	}

	if len(alertmanagerConfig.Receivers) == 0 {
		return fmt.Errorf("at least one receiver is required")
	}

	// 验证receiver是否存在
	receiverExists := false
	for _, receiver := range alertmanagerConfig.Receivers {
		if receiver.Name == alertmanagerConfig.Route.Receiver {
			receiverExists = true
			break
		}
	}

	if !receiverExists {
		return fmt.Errorf("receiver '%s' not found in receivers list", alertmanagerConfig.Route.Receiver)
	}

	logger.L.Debug("Alertmanager config validation passed",
		zap.Int("receivers", len(alertmanagerConfig.Receivers)),
	)

	return nil
}

func (a *AlertmanagerIntegration) TriggerReload(ctx context.Context, reloadURL string) error {
	// Alertmanager reload API: POST /-/reload
	req, err := http.NewRequestWithContext(ctx, "POST", reloadURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create reload request: %w", err)
	}

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send reload request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("alertmanager reload failed with status %d: %s", resp.StatusCode, string(body))
	}

	logger.L.Info("Alertmanager reload triggered successfully")
	return nil
}

// VMAlertIntegration VMAlert集成
type VMAlertIntegration struct {
	httpClient *http.Client
}

// NewVMAlertIntegration 创建VMAlert集成
func NewVMAlertIntegration() *VMAlertIntegration {
	return &VMAlertIntegration{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (v *VMAlertIntegration) GetConfigType() string {
	return "vmalert"
}

func (v *VMAlertIntegration) ValidateConfig(config []byte) error {
	// VMAlert使用类似Prometheus的规则格式，复用Prometheus验证逻辑
	prometheus := NewPrometheusIntegration()
	return prometheus.ValidateConfig(config)
}

func (v *VMAlertIntegration) TriggerReload(ctx context.Context, reloadURL string) error {
	// VMAlert reload API: POST /-/reload
	req, err := http.NewRequestWithContext(ctx, "POST", reloadURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create reload request: %w", err)
	}

	resp, err := v.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send reload request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("vmalert reload failed with status %d: %s", resp.StatusCode, string(body))
	}

	logger.L.Info("VMAlert reload triggered successfully")
	return nil
}

// IntegrationFactory 集成工厂
type IntegrationFactory struct{}

// NewIntegrationFactory 创建集成工厂
func NewIntegrationFactory() *IntegrationFactory {
	return &IntegrationFactory{}
}

// CreateIntegration 创建目标系统集成
func (f *IntegrationFactory) CreateIntegration(configType string) (TargetIntegration, error) {
	switch strings.ToLower(configType) {
	case "prometheus":
		return NewPrometheusIntegration(), nil
	case "alertmanager":
		return NewAlertmanagerIntegration(), nil
	case "vmalert":
		return NewVMAlertIntegration(), nil
	default:
		return nil, fmt.Errorf("unsupported config type: %s", configType)
	}
}

// GetSupportedTypes 获取支持的配置类型
func (f *IntegrationFactory) GetSupportedTypes() []string {
	return []string{"prometheus", "alertmanager", "vmalert"}
}

// ConfigTemplate 配置模板
type ConfigTemplate struct {
	Type        string `json:"type"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Template    string `json:"template"`
}

// GetConfigTemplates 获取配置模板
func (f *IntegrationFactory) GetConfigTemplates() []ConfigTemplate {
	return []ConfigTemplate{
		{
			Type:        "prometheus",
			Name:        "Basic Prometheus Rules",
			Description: "Basic Prometheus alerting rules template",
			Template: `groups:
  - name: basic_alerts
    rules:
      - alert: HighCPUUsage
        expr: cpu_usage_percent > 80
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "High CPU usage detected"
          description: "CPU usage is above 80% for more than 5 minutes"`,
		},
		{
			Type:        "alertmanager",
			Name:        "Basic Alertmanager Config",
			Description: "Basic Alertmanager configuration template",
			Template: `global:
  smtp_smarthost: 'localhost:587'
  smtp_from: 'alertmanager@example.com'

route:
  group_by: ['alertname']
  group_wait: 10s
  group_interval: 10s
  repeat_interval: 1h
  receiver: 'default'

receivers:
  - name: 'default'
    email_configs:
      - to: 'admin@example.com'
        subject: 'Alert: {{ .GroupLabels.alertname }}'
        body: |
          {{ range .Alerts }}
          Alert: {{ .Annotations.summary }}
          Description: {{ .Annotations.description }}
          {{ end }}`,
		},
		{
			Type:        "vmalert",
			Name:        "Basic VMAlert Rules",
			Description: "Basic VMAlert rules template",
			Template: `groups:
  - name: vmalert_rules
    rules:
      - alert: HighMemoryUsage
        expr: memory_usage_percent > 85
        for: 3m
        labels:
          severity: critical
        annotations:
          summary: "High memory usage detected"
          description: "Memory usage is above 85% for more than 3 minutes"`,
		},
	}
}