package cluster

import (
	"context"
	"time"

	"alert_agent/pkg/types"
)

// ClusterManager 集群管理器接口
// 负责多Alertmanager集群的注册、管理、健康监控、负载均衡和故障转移
type ClusterManager interface {
	// 集群注册和管理
	RegisterCluster(ctx context.Context, config *ClusterConfig) (*Cluster, error)
	UnregisterCluster(ctx context.Context, clusterID string) error
	GetCluster(ctx context.Context, clusterID string) (*Cluster, error)
	ListClusters(ctx context.Context, query types.Query) (*types.PageResult, error)
	UpdateCluster(ctx context.Context, clusterID string, config *ClusterConfig) (*Cluster, error)

	// 集群健康检查和监控
	HealthCheck(ctx context.Context, clusterID string) (*ClusterHealth, error)
	BatchHealthCheck(ctx context.Context, clusterIDs []string) (map[string]*ClusterHealth, error)
	StartHealthMonitor(ctx context.Context, interval time.Duration) error
	StopHealthMonitor() error
	GetHealthStatus(ctx context.Context, clusterID string) (*ClusterHealth, error)

	// 负载均衡和故障转移
	SelectCluster(ctx context.Context, strategy LoadBalanceStrategy) (*Cluster, error)
	Failover(ctx context.Context, failedClusterID string) (*Cluster, error)
	GetLoadBalanceStatus(ctx context.Context) (*LoadBalanceStatus, error)
	SetLoadBalanceStrategy(strategy LoadBalanceStrategy) error

	// 配置同步管理
	SyncConfig(ctx context.Context, clusterID string, config interface{}) error
	BatchSyncConfig(ctx context.Context, clusterIDs []string, config interface{}) error
	GetSyncStatus(ctx context.Context, clusterID string) (*SyncStatus, error)
	ValidateConfig(ctx context.Context, clusterType ClusterType, config interface{}) error

	// 集群指标和统计
	GetClusterMetrics(ctx context.Context, clusterID string) (*ClusterMetrics, error)
	GetClusterStats(ctx context.Context, clusterID string) (*ClusterStats, error)
	GetOverallStats(ctx context.Context) (*OverallClusterStats, error)

	// 集群发现和自动注册
	DiscoverClusters(ctx context.Context, discoveryConfig *DiscoveryConfig) ([]*Cluster, error)
	EnableAutoDiscovery(ctx context.Context, config *AutoDiscoveryConfig) error
	DisableAutoDiscovery(ctx context.Context) error

	// 配置模板和渲染
	CreateConfigTemplate(ctx context.Context, template *ConfigTemplate) error
	RenderConfig(ctx context.Context, templateID string, variables map[string]interface{}) (string, error)
	ApplyTemplate(ctx context.Context, clusterID string, templateID string, variables map[string]interface{}) error

	// 生命周期管理
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
	Reload(ctx context.Context) error
	GetStatus(ctx context.Context) (*ManagerStatus, error)
}

// LoadBalanceStrategy 负载均衡策略
type LoadBalanceStrategy string

const (
	LoadBalanceRoundRobin LoadBalanceStrategy = "round_robin"
	LoadBalanceWeighted   LoadBalanceStrategy = "weighted"
	LoadBalanceLeastConn  LoadBalanceStrategy = "least_conn"
	LoadBalanceHealthy    LoadBalanceStrategy = "healthy_only"
	LoadBalanceRandom     LoadBalanceStrategy = "random"
)

// LoadBalanceStatus 负载均衡状态
type LoadBalanceStatus struct {
	Strategy      LoadBalanceStrategy       `json:"strategy"`
	ActiveClusters int                      `json:"active_clusters"`
	TotalClusters  int                      `json:"total_clusters"`
	Distribution   map[string]*ClusterLoad  `json:"distribution"`
	LastUpdate     time.Time                `json:"last_update"`
}

// ClusterLoad 集群负载信息
type ClusterLoad struct {
	ClusterID    string    `json:"cluster_id"`
	Weight       float64   `json:"weight"`
	Connections  int       `json:"connections"`
	ResponseTime float64   `json:"response_time"`
	HealthScore  float64   `json:"health_score"`
	LastUsed     time.Time `json:"last_used"`
}

// SyncStatus 同步状态
type SyncStatus struct {
	ClusterID    string                 `json:"cluster_id"`
	Status       SyncStatusType         `json:"status"`
	LastSync     time.Time              `json:"last_sync"`
	NextSync     time.Time              `json:"next_sync"`
	Version      string                 `json:"version"`
	ConfigHash   string                 `json:"config_hash"`
	ErrorMessage string                 `json:"error_message,omitempty"`
	RetryCount   int                    `json:"retry_count"`
	SyncDetails  map[string]interface{} `json:"sync_details"`
}

// SyncStatusType 同步状态类型
type SyncStatusType string

const (
	SyncStatusPending    SyncStatusType = "pending"
	SyncStatusInProgress SyncStatusType = "in_progress"
	SyncStatusSuccess    SyncStatusType = "success"
	SyncStatusFailed     SyncStatusType = "failed"
	SyncStatusSkipped    SyncStatusType = "skipped"
)

// ClusterStats 集群统计信息
type ClusterStats struct {
	ClusterID        string            `json:"cluster_id"`
	Uptime           time.Duration     `json:"uptime"`
	TotalRequests    int64             `json:"total_requests"`
	SuccessRequests  int64             `json:"success_requests"`
	FailedRequests   int64             `json:"failed_requests"`
	AverageLatency   time.Duration     `json:"average_latency"`
	LastActivity     time.Time         `json:"last_activity"`
	ConfigVersion    string            `json:"config_version"`
	SyncCount        int64             `json:"sync_count"`
	FailoverCount    int64             `json:"failover_count"`
	CustomMetrics    map[string]interface{} `json:"custom_metrics"`
}

// OverallClusterStats 整体集群统计
type OverallClusterStats struct {
	TotalClusters    int                        `json:"total_clusters"`
	ActiveClusters   int                        `json:"active_clusters"`
	HealthyClusters  int                        `json:"healthy_clusters"`
	FailedClusters   int                        `json:"failed_clusters"`
	TotalRequests    int64                      `json:"total_requests"`
	TotalFailovers   int64                      `json:"total_failovers"`
	AverageLatency   time.Duration              `json:"average_latency"`
	ClusterStats     map[string]*ClusterStats   `json:"cluster_stats"`
	LastUpdate       time.Time                  `json:"last_update"`
}

// DiscoveryConfig 集群发现配置
type DiscoveryConfig struct {
	Method       DiscoveryMethod        `json:"method"`
	Interval     time.Duration          `json:"interval"`
	Targets      []string               `json:"targets"`
	Credentials  map[string]interface{} `json:"credentials"`
	Filters      map[string]interface{} `json:"filters"`
	AutoRegister bool                   `json:"auto_register"`
}

// DiscoveryMethod 发现方法
type DiscoveryMethod string

const (
	DiscoveryMethodKubernetes DiscoveryMethod = "kubernetes"
	DiscoveryMethodConsul     DiscoveryMethod = "consul"
	DiscoveryMethodEtcd       DiscoveryMethod = "etcd"
	DiscoveryMethodDNS        DiscoveryMethod = "dns"
	DiscoveryMethodStatic     DiscoveryMethod = "static"
)

// AutoDiscoveryConfig 自动发现配置
type AutoDiscoveryConfig struct {
	Enabled      bool                   `json:"enabled"`
	Interval     time.Duration          `json:"interval"`
	Discovery    *DiscoveryConfig       `json:"discovery"`
	Registration *RegistrationConfig    `json:"registration"`
	Filters      map[string]interface{} `json:"filters"`
}

// RegistrationConfig 注册配置
type RegistrationConfig struct {
	AutoApprove   bool                   `json:"auto_approve"`
	DefaultConfig map[string]interface{} `json:"default_config"`
	Tags          []string               `json:"tags"`
	Labels        map[string]string      `json:"labels"`
}

// ConfigTemplate 配置模板
type ConfigTemplate struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Type        ClusterType            `json:"type"`
	Template    string                 `json:"template"`
	Variables   map[string]interface{} `json:"variables"`
	Version     string                 `json:"version"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

// ManagerStatus 管理器状态
type ManagerStatus struct {
	Running          bool                   `json:"running"`
	StartTime        time.Time              `json:"start_time"`
	Uptime           time.Duration          `json:"uptime"`
	ManagedClusters  int                    `json:"managed_clusters"`
	HealthyMonitors  int                    `json:"healthy_monitors"`
	ActiveDiscovery  bool                   `json:"active_discovery"`
	LastHealthCheck  time.Time              `json:"last_health_check"`
	LastConfigSync   time.Time              `json:"last_config_sync"`
	Configuration    map[string]interface{} `json:"configuration"`
}