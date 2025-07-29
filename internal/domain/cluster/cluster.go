package cluster

import (
	"context"
	"time"
)

// Cluster 集群领域模型
type Cluster struct {
	ID                  string            `json:"id"`
	Name                string            `json:"name"`
	Endpoint            string            `json:"endpoint"`
	ConfigPath          string            `json:"config_path"`
	RulesPath           string            `json:"rules_path"`
	SyncInterval        int               `json:"sync_interval"`
	HealthCheckInterval int               `json:"health_check_interval"`
	Status              ClusterStatus     `json:"status"`
	HealthStatus        HealthStatus      `json:"health_status"`
	LastHealthCheck     *time.Time        `json:"last_health_check"`
	LastSyncTime        *time.Time        `json:"last_sync_time"`
	SyncStatus          SyncStatus        `json:"sync_status"`
	Labels              map[string]string `json:"labels"`
	Metadata            map[string]interface{} `json:"metadata"`
	CreatedAt           time.Time         `json:"created_at"`
	UpdatedAt           time.Time         `json:"updated_at"`
}

// Config 配置
type Config struct {
	Type    string                 `json:"type"`
	Content string                 `json:"content"`
	Hash    string                 `json:"hash"`
	Data    map[string]interface{} `json:"data"`
}

// SyncStatus 同步状态
type SyncStatus string

const (
	SyncStatusPending    SyncStatus = "pending"
	SyncStatusInProgress SyncStatus = "in_progress"
	SyncStatusSuccess    SyncStatus = "success"
	SyncStatusFailed     SyncStatus = "failed"
)

// ClusterStatus 集群状态
type ClusterStatus string

const (
	ClusterStatusActive   ClusterStatus = "active"
	ClusterStatusInactive ClusterStatus = "inactive"
	ClusterStatusDisabled ClusterStatus = "disabled"
)

// HealthStatus 健康状态
type HealthStatus string

const (
	HealthStatusHealthy   HealthStatus = "healthy"
	HealthStatusUnhealthy HealthStatus = "unhealthy"
	HealthStatusUnknown   HealthStatus = "unknown"
)

// ClusterConfig 集群配置
type ClusterConfig struct {
	Name                string            `json:"name" binding:"required"`
	Endpoint            string            `json:"endpoint" binding:"required"`
	ConfigPath          string            `json:"config_path"`
	RulesPath           string            `json:"rules_path"`
	SyncInterval        int               `json:"sync_interval"`
	HealthCheckInterval int               `json:"health_check_interval"`
	Labels              map[string]string `json:"labels"`
	Metadata            map[string]interface{} `json:"metadata"`
}

// SyncRecord 同步记录
type SyncRecord struct {
	ID           string                 `json:"id"`
	ClusterID    string                 `json:"cluster_id"`
	ConfigType   string                 `json:"config_type"`
	ConfigHash   string                 `json:"config_hash"`
	SyncStatus   SyncStatus             `json:"sync_status"`
	StartTime    time.Time              `json:"start_time"`
	EndTime      *time.Time             `json:"end_time"`
	Duration     int                    `json:"duration"`
	ErrorMessage string                 `json:"error_message"`
	ConfigData   map[string]interface{} `json:"config_data"`
	CreatedAt    time.Time              `json:"created_at"`
}

// ClusterRepository 集群仓储接口
type ClusterRepository interface {
	// 基本CRUD操作
	Create(ctx context.Context, cluster *Cluster) error
	Update(ctx context.Context, cluster *Cluster) error
	Delete(ctx context.Context, id string) error
	GetByID(ctx context.Context, id string) (*Cluster, error)
	GetByName(ctx context.Context, name string) (*Cluster, error)
	List(ctx context.Context) ([]*Cluster, error)

	// 查询操作
	GetByLabels(ctx context.Context, labels map[string]string) ([]*Cluster, error)
	GetActiveCluster(ctx context.Context) ([]*Cluster, error)

	// 健康状态操作
	UpdateHealthStatus(ctx context.Context, id string, status HealthStatus) error
	GetUnhealthyClusters(ctx context.Context) ([]*Cluster, error)

	// 同步状态操作
	UpdateSyncStatus(ctx context.Context, id string, status SyncStatus, syncTime time.Time) error
	CreateSyncRecord(ctx context.Context, record *SyncRecord) error
	GetSyncRecords(ctx context.Context, clusterID string, limit int) ([]*SyncRecord, error)
}

// ConfigSynchronizer 配置同步器接口
type ConfigSynchronizer interface {
	SyncConfig(ctx context.Context, cluster *Cluster, config *Config) error
	GetSyncStatus(ctx context.Context, clusterID string) (*SyncStatus, error)
	ValidateConfig(config *Config) error
}

// HealthChecker 健康检查器接口
type HealthChecker interface {
	CheckHealth(ctx context.Context, cluster *Cluster) (*HealthStatus, error)
	StartHealthCheck(ctx context.Context, clusters []*Cluster) error
	StopHealthCheck(ctx context.Context) error
}