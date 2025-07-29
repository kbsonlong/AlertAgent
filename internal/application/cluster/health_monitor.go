package cluster

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"

	clusterDomain "alert_agent/internal/domain/cluster"
)

// HealthMonitor 健康监控器
type HealthMonitor struct {
	repository     clusterDomain.Repository
	logger         *zap.Logger
	mu             sync.RWMutex
	running        bool
	monitoredClusters map[string]bool
	healthCache    map[string]*clusterDomain.ClusterHealth
	lastCheckTime  time.Time
	stopChan       chan struct{}
	interval       time.Duration
}

// NewHealthMonitor 创建新的健康监控器
func NewHealthMonitor(repository clusterDomain.Repository, logger *zap.Logger) *HealthMonitor {
	return &HealthMonitor{
		repository:        repository,
		logger:            logger,
		monitoredClusters: make(map[string]bool),
		healthCache:       make(map[string]*clusterDomain.ClusterHealth),
		stopChan:          make(chan struct{}),
		interval:          30 * time.Second, // 默认30秒检查间隔
	}
}

// Start 启动健康监控器
func (hm *HealthMonitor) Start(ctx context.Context) error {
	hm.mu.Lock()
	defer hm.mu.Unlock()

	if hm.running {
		return fmt.Errorf("health monitor is already running")
	}

	hm.running = true
	hm.logger.Info("Health monitor started")
	return nil
}

// Stop 停止健康监控器
func (hm *HealthMonitor) Stop() {
	hm.mu.Lock()
	defer hm.mu.Unlock()

	if !hm.running {
		return
	}

	close(hm.stopChan)
	hm.running = false
	hm.logger.Info("Health monitor stopped")
}

// StartMonitoring 开始监控
func (hm *HealthMonitor) StartMonitoring(ctx context.Context, interval time.Duration) error {
	hm.mu.Lock()
	hm.interval = interval
	hm.mu.Unlock()

	go hm.monitorLoop(ctx)
	return nil
}

// StopMonitoring 停止监控
func (hm *HealthMonitor) StopMonitoring() {
	hm.Stop()
}

// AddCluster 添加集群到监控列表
func (hm *HealthMonitor) AddCluster(clusterID string) {
	hm.mu.Lock()
	defer hm.mu.Unlock()

	hm.monitoredClusters[clusterID] = true
	hm.logger.Info("Added cluster to monitoring", zap.String("cluster_id", clusterID))
}

// RemoveCluster 从监控列表移除集群
func (hm *HealthMonitor) RemoveCluster(clusterID string) {
	hm.mu.Lock()
	defer hm.mu.Unlock()

	delete(hm.monitoredClusters, clusterID)
	delete(hm.healthCache, clusterID)
	hm.logger.Info("Removed cluster from monitoring", zap.String("cluster_id", clusterID))
}

// CheckCluster 检查单个集群健康状态
func (hm *HealthMonitor) CheckCluster(ctx context.Context, clusterID string) (*clusterDomain.ClusterHealth, error) {
	// 获取集群信息
	cluster, err := hm.repository.GetByID(ctx, clusterID)
	if err != nil {
		return nil, fmt.Errorf("failed to get cluster: %w", err)
	}

	// 执行健康检查
	health := &clusterDomain.ClusterHealth{
		ClusterID: clusterID,
		Status:    clusterDomain.ClusterStatusActive, // 默认活跃
		Healthy:   true,
		Message:   "Health check completed",
		Endpoints: make(map[string]*clusterDomain.EndpointHealth),
		Metrics: &clusterDomain.HealthMetrics{
			CPUUsage:    0.0,
			MemoryUsage: 0.0,
			DiskUsage:   0.0,
			Connections: 0,
			Requests:    0,
			Errors:      0,
		},
		LastCheck: time.Now(),
		Uptime:    time.Hour, // 模拟运行时间
	}

	// 检查每个端点
	for _, endpoint := range cluster.Endpoints {
		endpointHealth := hm.checkEndpoint(ctx, endpoint)
		health.Endpoints[endpoint] = endpointHealth
		
		// 如果有端点不健康，整体状态为不健康
		if !endpointHealth.Healthy {
			health.Healthy = false
			health.Status = clusterDomain.ClusterStatusError
			health.Message = "Some endpoints are unhealthy"
		}
	}

	// 更新缓存
	hm.mu.Lock()
	hm.healthCache[clusterID] = health
	hm.lastCheckTime = time.Now()
	hm.mu.Unlock()

	return health, nil
}

// BatchCheck 批量健康检查
func (hm *HealthMonitor) BatchCheck(ctx context.Context, clusterIDs []string) (map[string]*clusterDomain.ClusterHealth, error) {
	results := make(map[string]*clusterDomain.ClusterHealth)

	for _, clusterID := range clusterIDs {
		health, err := hm.CheckCluster(ctx, clusterID)
		if err != nil {
			hm.logger.Error("Failed to check cluster health",
				zap.String("cluster_id", clusterID),
				zap.Error(err))
			// 创建错误状态
			health = &clusterDomain.ClusterHealth{
				ClusterID: clusterID,
				Status:    clusterDomain.ClusterStatusUnknown,
				Healthy:   false,
				Message:   err.Error(),
				LastCheck: time.Now(),
				Endpoints: make(map[string]*clusterDomain.EndpointHealth),
			}
		}
		results[clusterID] = health
	}

	return results, nil
}

// GetActiveMonitorCount 获取活跃监控数量
func (hm *HealthMonitor) GetActiveMonitorCount() int {
	hm.mu.RLock()
	defer hm.mu.RUnlock()
	return len(hm.monitoredClusters)
}

// GetLastCheckTime 获取最后检查时间
func (hm *HealthMonitor) GetLastCheckTime() time.Time {
	hm.mu.RLock()
	defer hm.mu.RUnlock()
	return hm.lastCheckTime
}

// monitorLoop 监控循环
func (hm *HealthMonitor) monitorLoop(ctx context.Context) {
	ticker := time.NewTicker(hm.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-hm.stopChan:
			return
		case <-ticker.C:
			hm.performHealthChecks(ctx)
		}
	}
}

// performHealthChecks 执行健康检查
func (hm *HealthMonitor) performHealthChecks(ctx context.Context) {
	hm.mu.RLock()
	clusterIDs := make([]string, 0, len(hm.monitoredClusters))
	for clusterID := range hm.monitoredClusters {
		clusterIDs = append(clusterIDs, clusterID)
	}
	hm.mu.RUnlock()

	if len(clusterIDs) == 0 {
		return
	}

	hm.logger.Debug("Performing scheduled health checks",
		zap.Int("cluster_count", len(clusterIDs)))

	_, err := hm.BatchCheck(ctx, clusterIDs)
	if err != nil {
		hm.logger.Error("Failed to perform batch health check", zap.Error(err))
	}
}

// checkEndpoint 检查单个端点
func (hm *HealthMonitor) checkEndpoint(ctx context.Context, endpoint string) *clusterDomain.EndpointHealth {
	start := time.Now()
	
	// TODO: 实现实际的端点健康检查逻辑
	// 这里是模拟实现
	time.Sleep(10 * time.Millisecond) // 模拟网络延迟
	
	return &clusterDomain.EndpointHealth{
		URL:                 endpoint,
		Healthy:             true,
		Latency:             time.Since(start),
		LastCheck:           time.Now(),
		ConsecutiveFailures: 0,
	}
}