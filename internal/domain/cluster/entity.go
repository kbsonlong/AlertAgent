package cluster

import (
	"fmt"
	"time"

	"alert_agent/pkg/types"
)

// Cluster 集群实体
type Cluster struct {
	ID          string            `json:"id" gorm:"primaryKey;type:varchar(36)"`
	Name        string            `json:"name" gorm:"type:varchar(255);not null;uniqueIndex"`
	Type        ClusterType       `json:"type" gorm:"type:varchar(50);not null"`
	Description string            `json:"description" gorm:"type:text"`
	Config      ClusterConfig     `json:"config" gorm:"type:json"`
	Status      ClusterStatus     `json:"status" gorm:"type:varchar(20);default:'active'"`
	Version     string            `json:"version" gorm:"type:varchar(50)"`
	Endpoints   []string          `json:"endpoints" gorm:"type:json"`
	Tags        []string          `json:"tags" gorm:"type:json"`
	Labels      map[string]string `json:"labels" gorm:"type:json"`
	Metadata    types.Metadata    `json:"metadata" gorm:"embedded"`
	CreatedAt   time.Time         `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time         `json:"updated_at" gorm:"autoUpdateTime"`
}

// ClusterType 集群类型
type ClusterType string

const (
	ClusterTypeAlertmanager ClusterType = "alertmanager"
	ClusterTypePrometheus   ClusterType = "prometheus"
	ClusterTypeGrafana      ClusterType = "grafana"
	ClusterTypeCustom       ClusterType = "custom"
)

// ClusterStatus 集群状态
type ClusterStatus string

const (
	ClusterStatusActive      ClusterStatus = "active"
	ClusterStatusInactive    ClusterStatus = "inactive"
	ClusterStatusMaintenance ClusterStatus = "maintenance"
	ClusterStatusError       ClusterStatus = "error"
	ClusterStatusUnknown     ClusterStatus = "unknown"
)

// ClusterConfig 集群配置
type ClusterConfig struct {
	// 认证配置
	Auth AuthConfig `json:"auth"`

	// 连接配置
	Connection ConnectionConfig `json:"connection"`

	// 健康检查配置
	HealthCheck HealthCheckConfig `json:"health_check"`

	// 同步配置
	Sync SyncConfig `json:"sync"`

	// 高可用配置
	HA HAConfig `json:"ha"`

	// 自定义设置
	Settings map[string]interface{} `json:"settings"`
}

// AuthConfig 认证配置
type AuthConfig struct {
	Type     string `json:"type"`     // basic, bearer, oauth2, cert
	Username string `json:"username"`
	Password string `json:"password"`
	Token    string `json:"token"`
	CertFile string `json:"cert_file"`
	KeyFile  string `json:"key_file"`
	CAFile   string `json:"ca_file"`
}

// ConnectionConfig 连接配置
type ConnectionConfig struct {
	Timeout         time.Duration     `json:"timeout"`
	RetryConfig     types.RetryConfig `json:"retry_config"`
	MaxConnections  int               `json:"max_connections"`
	KeepAlive       time.Duration     `json:"keep_alive"`
	TLSConfig       TLSConfig         `json:"tls_config"`
	ProxyURL        string            `json:"proxy_url"`
}

// TLSConfig TLS配置
type TLSConfig struct {
	Enabled            bool   `json:"enabled"`
	InsecureSkipVerify bool   `json:"insecure_skip_verify"`
	ServerName         string `json:"server_name"`
	CertFile           string `json:"cert_file"`
	KeyFile            string `json:"key_file"`
	CAFile             string `json:"ca_file"`
}

// HealthCheckConfig 健康检查配置
type HealthCheckConfig struct {
	Enabled  bool          `json:"enabled"`
	Interval time.Duration `json:"interval"`
	Timeout  time.Duration `json:"timeout"`
	Path     string        `json:"path"`
	Method   string        `json:"method"`
}

// SyncConfig 同步配置
type SyncConfig struct {
	Enabled   bool          `json:"enabled"`
	Interval  time.Duration `json:"interval"`
	BatchSize int           `json:"batch_size"`
	Parallel  int           `json:"parallel"`
}

// HAConfig 高可用配置
type HAConfig struct {
	Enabled         bool     `json:"enabled"`
	Primary         string   `json:"primary"`
	Secondaries     []string `json:"secondaries"`
	FailoverTimeout time.Duration `json:"failover_timeout"`
	LoadBalance     string   `json:"load_balance"` // round_robin, least_conn, random
}

// Validate 验证集群配置
func (c *Cluster) Validate() error {
	if c.Name == "" {
		return fmt.Errorf("name is required")
	}
	if c.Type == "" {
		return fmt.Errorf("type is required")
	}
	if len(c.Endpoints) == 0 {
		return fmt.Errorf("at least one endpoint is required")
	}
	return nil
}

// IsActive 检查集群是否激活
func (c *Cluster) IsActive() bool {
	return c.Status == ClusterStatusActive
}

// GetPrimaryEndpoint 获取主要端点
func (c *Cluster) GetPrimaryEndpoint() string {
	if len(c.Endpoints) == 0 {
		return ""
	}
	return c.Endpoints[0]
}

// GetSecondaryEndpoints 获取备用端点
func (c *Cluster) GetSecondaryEndpoints() []string {
	if len(c.Endpoints) <= 1 {
		return []string{}
	}
	return c.Endpoints[1:]
}

// AddEndpoint 添加端点
func (c *Cluster) AddEndpoint(endpoint string) {
	for _, ep := range c.Endpoints {
		if ep == endpoint {
			return
		}
	}
	c.Endpoints = append(c.Endpoints, endpoint)
}

// RemoveEndpoint 移除端点
func (c *Cluster) RemoveEndpoint(endpoint string) {
	for i, ep := range c.Endpoints {
		if ep == endpoint {
			c.Endpoints = append(c.Endpoints[:i], c.Endpoints[i+1:]...)
			return
		}
	}
}

// HasEndpoint 检查是否有端点
func (c *Cluster) HasEndpoint(endpoint string) bool {
	for _, ep := range c.Endpoints {
		if ep == endpoint {
			return true
		}
	}
	return false
}

// GetSetting 获取设置值
func (c *Cluster) GetSetting(key string) interface{} {
	if c.Config.Settings == nil {
		return nil
	}
	return c.Config.Settings[key]
}

// SetSetting 设置配置值
func (c *Cluster) SetSetting(key string, value interface{}) {
	if c.Config.Settings == nil {
		c.Config.Settings = make(map[string]interface{})
	}
	c.Config.Settings[key] = value
}

// AddTag 添加标签
func (c *Cluster) AddTag(tag string) {
	for _, t := range c.Tags {
		if t == tag {
			return
		}
	}
	c.Tags = append(c.Tags, tag)
}

// RemoveTag 移除标签
func (c *Cluster) RemoveTag(tag string) {
	for i, t := range c.Tags {
		if t == tag {
			c.Tags = append(c.Tags[:i], c.Tags[i+1:]...)
			return
		}
	}
}

// HasTag 检查是否有标签
func (c *Cluster) HasTag(tag string) bool {
	for _, t := range c.Tags {
		if t == tag {
			return true
		}
	}
	return false
}

// SetLabel 设置标签
func (c *Cluster) SetLabel(key, value string) {
	if c.Labels == nil {
		c.Labels = make(map[string]string)
	}
	c.Labels[key] = value
}

// GetLabel 获取标签值
func (c *Cluster) GetLabel(key string) string {
	if c.Labels == nil {
		return ""
	}
	return c.Labels[key]
}

// HasLabel 检查是否有标签
func (c *Cluster) HasLabel(key string) bool {
	if c.Labels == nil {
		return false
	}
	_, exists := c.Labels[key]
	return exists
}