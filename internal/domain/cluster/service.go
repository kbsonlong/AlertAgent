package cluster

import (
	"context"
	"time"

	"alert_agent/pkg/types"
)

// Service 集群服务接口
type Service interface {
	// CreateCluster 创建集群
	CreateCluster(ctx context.Context, req *CreateClusterRequest) (*Cluster, error)

	// GetCluster 获取集群
	GetCluster(ctx context.Context, id string) (*Cluster, error)

	// UpdateCluster 更新集群
	UpdateCluster(ctx context.Context, id string, req *UpdateClusterRequest) (*Cluster, error)

	// DeleteCluster 删除集群
	DeleteCluster(ctx context.Context, id string) error

	// ListClusters 获取集群列表
	ListClusters(ctx context.Context, query types.Query) (*types.PageResult, error)

	// TestClusterConnection 测试集群连接
	TestClusterConnection(ctx context.Context, id string) (*ConnectionTestResult, error)

	// GetClusterHealth 获取集群健康状态
	GetClusterHealth(ctx context.Context, id string) (*ClusterHealth, error)

	// SyncClusterConfig 同步集群配置
	SyncClusterConfig(ctx context.Context, id string) error

	// GetClustersByType 根据类型获取集群
	GetClustersByType(ctx context.Context, clusterType ClusterType) ([]*Cluster, error)

	// GetActiveClusters 获取激活的集群
	GetActiveClusters(ctx context.Context) ([]*Cluster, error)

	// UpdateClusterStatus 更新集群状态
	UpdateClusterStatus(ctx context.Context, id string, status ClusterStatus) error

	// ValidateClusterConfig 验证集群配置
	ValidateClusterConfig(ctx context.Context, clusterType ClusterType, config ClusterConfig) error

	// GetClusterMetrics 获取集群指标
	GetClusterMetrics(ctx context.Context, id string) (*ClusterMetrics, error)

	// BulkUpdateClusters 批量更新集群
	BulkUpdateClusters(ctx context.Context, updates []*BulkUpdateClusterRequest) error

	// ImportClusters 导入集群
	ImportClusters(ctx context.Context, clusters []*Cluster) (*ImportClusterResult, error)

	// ExportClusters 导出集群
	ExportClusters(ctx context.Context, filter map[string]interface{}) ([]*Cluster, error)

	// DiscoverClusters 发现集群
	DiscoverClusters(ctx context.Context, req *DiscoverRequest) ([]*Cluster, error)

	// MonitorCluster 监控集群
	MonitorCluster(ctx context.Context, id string) (*MonitorResult, error)
}

// CreateClusterRequest 创建集群请求
type CreateClusterRequest struct {
	Name        string            `json:"name" validate:"required,min=1,max=255"`
	Type        ClusterType       `json:"type" validate:"required"`
	Description string            `json:"description" validate:"max=1000"`
	Config      ClusterConfig     `json:"config" validate:"required"`
	Version     string            `json:"version"`
	Endpoints   []string          `json:"endpoints" validate:"required,min=1"`
	Tags        []string          `json:"tags"`
	Labels      map[string]string `json:"labels"`
}

// UpdateClusterRequest 更新集群请求
type UpdateClusterRequest struct {
	Name        *string            `json:"name,omitempty" validate:"omitempty,min=1,max=255"`
	Description *string            `json:"description,omitempty" validate:"omitempty,max=1000"`
	Config      *ClusterConfig     `json:"config,omitempty"`
	Status      *ClusterStatus     `json:"status,omitempty"`
	Version     *string            `json:"version,omitempty"`
	Endpoints   []string           `json:"endpoints,omitempty"`
	Tags        []string           `json:"tags,omitempty"`
	Labels      map[string]string  `json:"labels,omitempty"`
}

// ConnectionTestResult 连接测试结果
type ConnectionTestResult struct {
	Success     bool                       `json:"success"`
	Message     string                     `json:"message"`
	Latency     time.Duration              `json:"latency"`
	Endpoints   map[string]*EndpointResult `json:"endpoints"`
	Timestamp   time.Time                  `json:"timestamp"`
	Version     string                     `json:"version"`
	Features    []string                   `json:"features"`
}

// EndpointResult 端点测试结果
type EndpointResult struct {
	URL       string        `json:"url"`
	Success   bool          `json:"success"`
	Latency   time.Duration `json:"latency"`
	Error     string        `json:"error,omitempty"`
	Status    int           `json:"status"`
	Timestamp time.Time     `json:"timestamp"`
}

// ClusterHealth 集群健康状态
type ClusterHealth struct {
	ClusterID   string                     `json:"cluster_id"`
	Status      ClusterStatus              `json:"status"`
	Healthy     bool                       `json:"healthy"`
	Message     string                     `json:"message"`
	Endpoints   map[string]*EndpointHealth `json:"endpoints"`
	Metrics     *HealthMetrics             `json:"metrics"`
	LastCheck   time.Time                  `json:"last_check"`
	Uptime      time.Duration              `json:"uptime"`
}

// EndpointHealth 端点健康状态
type EndpointHealth struct {
	URL         string        `json:"url"`
	Healthy     bool          `json:"healthy"`
	Latency     time.Duration `json:"latency"`
	Error       string        `json:"error,omitempty"`
	LastCheck   time.Time     `json:"last_check"`
	ConsecutiveFailures int   `json:"consecutive_failures"`
}

// HealthMetrics 健康指标
type HealthMetrics struct {
	CPUUsage    float64 `json:"cpu_usage"`
	MemoryUsage float64 `json:"memory_usage"`
	DiskUsage   float64 `json:"disk_usage"`
	Connections int     `json:"connections"`
	Requests    int64   `json:"requests"`
	Errors      int64   `json:"errors"`
}

// ClusterMetrics 集群指标
type ClusterMetrics struct {
	ClusterID       string                    `json:"cluster_id"`
	TotalRequests   int64                     `json:"total_requests"`
	TotalErrors     int64                     `json:"total_errors"`
	SuccessRate     float64                   `json:"success_rate"`
	AvgLatency      time.Duration             `json:"avg_latency"`
	Throughput      float64                   `json:"throughput"`
	EndpointMetrics map[string]*EndpointMetrics `json:"endpoint_metrics"`
	Timestamp       time.Time                 `json:"timestamp"`
}

// EndpointMetrics 端点指标
type EndpointMetrics struct {
	URL           string        `json:"url"`
	Requests      int64         `json:"requests"`
	Errors        int64         `json:"errors"`
	SuccessRate   float64       `json:"success_rate"`
	AvgLatency    time.Duration `json:"avg_latency"`
	LastRequest   time.Time     `json:"last_request"`
	LastError     time.Time     `json:"last_error"`
}

// BulkUpdateClusterRequest 批量更新集群请求
type BulkUpdateClusterRequest struct {
	ClusterID string               `json:"cluster_id"`
	Updates   UpdateClusterRequest `json:"updates"`
}

// ImportClusterResult 导入集群结果
type ImportClusterResult struct {
	Total     int      `json:"total"`
	Success   int      `json:"success"`
	Failed    int      `json:"failed"`
	Errors    []string `json:"errors"`
	Created   []string `json:"created"`
	Updated   []string `json:"updated"`
	Skipped   []string `json:"skipped"`
}

// DiscoverRequest 发现请求
type DiscoverRequest struct {
	Type      ClusterType `json:"type"`
	Network   string      `json:"network"`   // CIDR格式
	PortRange string      `json:"port_range"` // 如: "9090-9099"
	Timeout   time.Duration `json:"timeout"`
	Parallel  int         `json:"parallel"`
}

// MonitorResult 监控结果
type MonitorResult struct {
	ClusterID string                   `json:"cluster_id"`
	Status    ClusterStatus            `json:"status"`
	Health    *ClusterHealth           `json:"health"`
	Metrics   *ClusterMetrics          `json:"metrics"`
	Alerts    []*MonitorAlert          `json:"alerts"`
	Timestamp time.Time                `json:"timestamp"`
}

// MonitorAlert 监控告警
type MonitorAlert struct {
	Type        string                 `json:"type"`
	Severity    string                 `json:"severity"`
	Message     string                 `json:"message"`
	Details     map[string]interface{} `json:"details"`
	Timestamp   time.Time              `json:"timestamp"`
	Resolved    bool                   `json:"resolved"`
	ResolvedAt  *time.Time             `json:"resolved_at,omitempty"`
}