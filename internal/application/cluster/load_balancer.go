package cluster

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"sync"
	"time"

	clusterDomain "alert_agent/internal/domain/cluster"
	"go.uber.org/zap"
)

// LoadBalancer 负载均衡器
type LoadBalancer struct {
	mu       sync.RWMutex
	strategy clusterDomain.LoadBalanceStrategy
	clusters map[string]*clusterDomain.Cluster
	stats    map[string]*clusterDomain.ClusterLoad
	logger   *zap.Logger
}

// NewLoadBalancer 创建新的负载均衡器
func NewLoadBalancer(strategy clusterDomain.LoadBalanceStrategy, logger *zap.Logger) *LoadBalancer {
	return &LoadBalancer{
		strategy: strategy,
		clusters: make(map[string]*clusterDomain.Cluster),
		stats:    make(map[string]*clusterDomain.ClusterLoad),
		logger:   logger,
	}
}

// AddCluster 添加集群到负载均衡器
func (lb *LoadBalancer) AddCluster(cluster *clusterDomain.Cluster) {
	lb.mu.Lock()
	defer lb.mu.Unlock()
	
	lb.clusters[cluster.ID] = cluster
	lb.stats[cluster.ID] = &clusterDomain.ClusterLoad{
		ClusterID:    cluster.ID,
		Weight:       1.0,
		Connections:  0,
		ResponseTime: 0,
		HealthScore:  1.0,
		LastUsed:     time.Now(),
	}
	
	lb.logger.Info("Cluster added to load balancer", zap.String("cluster_id", cluster.ID))
}

// RemoveCluster 从负载均衡器移除集群
func (lb *LoadBalancer) RemoveCluster(clusterID string) {
	lb.mu.Lock()
	defer lb.mu.Unlock()
	
	delete(lb.clusters, clusterID)
	delete(lb.stats, clusterID)
	
	lb.logger.Info("Cluster removed from load balancer", zap.String("cluster_id", clusterID))
}

// SelectCluster 根据负载均衡策略选择集群
func (lb *LoadBalancer) SelectCluster(ctx context.Context) (*clusterDomain.Cluster, error) {
	lb.mu.RLock()
	defer lb.mu.RUnlock()
	
	// 获取健康的集群
	healthyClusters := make([]*clusterDomain.Cluster, 0)
	for _, cluster := range lb.clusters {
		if stats, exists := lb.stats[cluster.ID]; exists && stats.HealthScore > 0.5 {
			healthyClusters = append(healthyClusters, cluster)
		}
	}
	
	if len(healthyClusters) == 0 {
		return nil, errors.New("no healthy clusters available")
	}
	
	// 根据策略选择集群
	switch lb.strategy {
	case clusterDomain.LoadBalanceRoundRobin:
		return lb.selectRoundRobin(healthyClusters), nil
	case clusterDomain.LoadBalanceWeighted:
		return lb.selectWeighted(healthyClusters), nil
	case clusterDomain.LoadBalanceLeastConn:
		return lb.selectLeastConnections(healthyClusters), nil
	case clusterDomain.LoadBalanceRandom:
		return lb.selectRandom(healthyClusters), nil
	default:
		return lb.selectRoundRobin(healthyClusters), nil
	}
}

// selectRoundRobin 轮询选择
func (lb *LoadBalancer) selectRoundRobin(clusters []*clusterDomain.Cluster) *clusterDomain.Cluster {
	// 简单实现：基于时间戳选择
	index := int(time.Now().UnixNano()) % len(clusters)
	return clusters[index]
}

// selectWeighted 加权选择
func (lb *LoadBalancer) selectWeighted(clusters []*clusterDomain.Cluster) *clusterDomain.Cluster {
	// 简单实现：根据权重随机选择
	totalWeight := 0.0
	for _, cluster := range clusters {
		if stats, exists := lb.stats[cluster.ID]; exists {
			totalWeight += stats.Weight
		}
	}
	
	if totalWeight == 0 {
		return clusters[0]
	}
	
	random := rand.Float64() * totalWeight
	currentWeight := 0.0
	
	for _, cluster := range clusters {
		if stats, exists := lb.stats[cluster.ID]; exists {
			currentWeight += stats.Weight
			if random <= currentWeight {
				return cluster
			}
		}
	}
	
	return clusters[len(clusters)-1]
}

// selectLeastConnections 最少连接选择
func (lb *LoadBalancer) selectLeastConnections(clusters []*clusterDomain.Cluster) *clusterDomain.Cluster {
	var selected *clusterDomain.Cluster
	minConnections := int64(-1)
	
	for _, cluster := range clusters {
		if stats, exists := lb.stats[cluster.ID]; exists {
			if minConnections == -1 || int64(stats.Connections) < minConnections {
				minConnections = int64(stats.Connections)
				selected = cluster
			}
		}
	}
	
	if selected == nil {
		return clusters[0]
	}
	
	return selected
}

// selectRandom 随机选择
func (lb *LoadBalancer) selectRandom(clusters []*clusterDomain.Cluster) *clusterDomain.Cluster {
	index := rand.Intn(len(clusters))
	return clusters[index]
}

// UpdateClusterHealth 更新集群健康状态
func (lb *LoadBalancer) UpdateClusterHealth(clusterID string, healthy bool) {
	lb.mu.Lock()
	defer lb.mu.Unlock()
	
	if stats, exists := lb.stats[clusterID]; exists {
		if healthy {
			stats.HealthScore = 1.0
		} else {
			stats.HealthScore = 0.0
		}
		lb.logger.Info("Cluster health updated", 
			zap.String("cluster_id", clusterID),
			zap.Bool("healthy", healthy))
	}
}

// UpdateStats 更新集群统计信息
func (lb *LoadBalancer) UpdateStats(clusterID string, responseTime time.Duration, success bool) {
	lb.mu.Lock()
	defer lb.mu.Unlock()
	
	if stats, exists := lb.stats[clusterID]; exists {
		// 更新响应时间
		if stats.ResponseTime == 0 {
			stats.ResponseTime = float64(responseTime.Milliseconds())
		} else {
			stats.ResponseTime = (stats.ResponseTime + float64(responseTime.Milliseconds())) / 2
		}
		
		// 更新健康分数
		if success {
			stats.HealthScore = (stats.HealthScore + 1.0) / 2
		} else {
			stats.HealthScore = stats.HealthScore * 0.8
		}
		
		stats.LastUsed = time.Now()
	}
}

// GetStats 获取负载均衡统计信息
func (lb *LoadBalancer) GetStats() map[string]*clusterDomain.ClusterLoad {
	lb.mu.RLock()
	defer lb.mu.RUnlock()
	
	stats := make(map[string]*clusterDomain.ClusterLoad)
	for id, stat := range lb.stats {
		stats[id] = &clusterDomain.ClusterLoad{
			ClusterID:    stat.ClusterID,
			Weight:       stat.Weight,
			Connections:  stat.Connections,
			ResponseTime: stat.ResponseTime,
			HealthScore:  stat.HealthScore,
			LastUsed:     stat.LastUsed,
		}
	}
	
	return stats
}

// SetStrategy 设置负载均衡策略
func (lb *LoadBalancer) SetStrategy(strategy clusterDomain.LoadBalanceStrategy) {
	lb.mu.Lock()
	defer lb.mu.Unlock()
	
	lb.strategy = strategy
	lb.logger.Info("Load balance strategy updated", zap.String("strategy", string(strategy)))
}

// GetStrategy 获取当前负载均衡策略
func (lb *LoadBalancer) GetStrategy() clusterDomain.LoadBalanceStrategy {
	lb.mu.RLock()
	defer lb.mu.RUnlock()
	
	return lb.strategy
}

// IncrementActiveConnections 增加活跃连接数
func (lb *LoadBalancer) IncrementActiveConnections(clusterID string) {
	lb.mu.Lock()
	defer lb.mu.Unlock()
	
	if stats, exists := lb.stats[clusterID]; exists {
		stats.Connections++
	}
}

// DecrementActiveConnections 减少活跃连接数
func (lb *LoadBalancer) DecrementActiveConnections(clusterID string) {
	lb.mu.Lock()
	defer lb.mu.Unlock()
	
	if stats, exists := lb.stats[clusterID]; exists {
		if stats.Connections > 0 {
			stats.Connections--
		}
	}
}

// SetWeight 设置集群权重
func (lb *LoadBalancer) SetWeight(clusterID string, weight float64) error {
	lb.mu.Lock()
	defer lb.mu.Unlock()
	
	if weight < 0 {
		return fmt.Errorf("weight cannot be negative")
	}
	
	if stats, exists := lb.stats[clusterID]; exists {
		stats.Weight = weight
		lb.logger.Info("Cluster weight updated", 
			zap.String("cluster_id", clusterID),
			zap.Float64("weight", weight))
		return nil
	}
	
	return fmt.Errorf("cluster not found: %s", clusterID)
}