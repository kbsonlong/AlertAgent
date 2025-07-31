package worker

import (
	"context"
	"fmt"
	"sync"
	"time"

	"alert_agent/internal/pkg/logger"
	"alert_agent/internal/pkg/queue"

	"go.uber.org/zap"
)

// WorkerScaler Worker扩缩容管理器
type WorkerScaler struct {
	manager         *WorkerManager
	scalingPolicies map[string]*ScalingPolicy
	monitor         *WorkerMonitor
	mutex           sync.RWMutex
	isRunning       bool
	stopChan        chan struct{}
}

// ScalingPolicy 扩缩容策略
type ScalingPolicy struct {
	WorkerType          string        `json:"worker_type"`
	MinInstances        int           `json:"min_instances"`
	MaxInstances        int           `json:"max_instances"`
	TargetCPUPercent    float64       `json:"target_cpu_percent"`
	TargetQueueLength   int64         `json:"target_queue_length"`
	ScaleUpThreshold    float64       `json:"scale_up_threshold"`
	ScaleDownThreshold  float64       `json:"scale_down_threshold"`
	ScaleUpCooldown     time.Duration `json:"scale_up_cooldown"`
	ScaleDownCooldown   time.Duration `json:"scale_down_cooldown"`
	CheckInterval       time.Duration `json:"check_interval"`
	LastScaleUp         time.Time     `json:"last_scale_up"`
	LastScaleDown       time.Time     `json:"last_scale_down"`
}

// WorkerMonitor Worker监控器
type WorkerMonitor struct {
	messageQueue queue.MessageQueue
	metrics      map[string]*WorkerMetrics
	mutex        sync.RWMutex
}

// WorkerMetrics Worker指标
type WorkerMetrics struct {
	WorkerName         string            `json:"worker_name"`
	WorkerType         string            `json:"worker_type"`
	CPUUsage           float64           `json:"cpu_usage"`
	MemoryUsage        float64           `json:"memory_usage"`
	TasksPerSecond     float64           `json:"tasks_per_second"`
	AverageLatency     time.Duration     `json:"average_latency"`
	QueueLengths       map[string]int64  `json:"queue_lengths"`
	ErrorRate          float64           `json:"error_rate"`
	LastUpdate         time.Time         `json:"last_update"`
	HealthStatus       string            `json:"health_status"`
}

// ScalingDecision 扩缩容决策
type ScalingDecision struct {
	WorkerType    string    `json:"worker_type"`
	Action        string    `json:"action"` // scale_up, scale_down, no_action
	CurrentCount  int       `json:"current_count"`
	TargetCount   int       `json:"target_count"`
	Reason        string    `json:"reason"`
	Timestamp     time.Time `json:"timestamp"`
}

// NewWorkerScaler 创建Worker扩缩容管理器
func NewWorkerScaler(manager *WorkerManager, messageQueue queue.MessageQueue) *WorkerScaler {
	monitor := &WorkerMonitor{
		messageQueue: messageQueue,
		metrics:      make(map[string]*WorkerMetrics),
	}

	scaler := &WorkerScaler{
		manager:         manager,
		scalingPolicies: make(map[string]*ScalingPolicy),
		monitor:         monitor,
		stopChan:        make(chan struct{}),
	}

	// 设置默认扩缩容策略
	scaler.setDefaultPolicies()

	return scaler
}

// Start 启动扩缩容管理器
func (s *WorkerScaler) Start(ctx context.Context) error {
	s.mutex.Lock()
	if s.isRunning {
		s.mutex.Unlock()
		return fmt.Errorf("worker scaler is already running")
	}
	s.isRunning = true
	s.mutex.Unlock()

	logger.L.Info("Starting worker scaler")

	// 启动监控循环
	go s.monitorLoop(ctx)

	// 启动扩缩容决策循环
	go s.scalingLoop(ctx)

	return nil
}

// Stop 停止扩缩容管理器
func (s *WorkerScaler) Stop() error {
	s.mutex.Lock()
	if !s.isRunning {
		s.mutex.Unlock()
		return nil
	}
	s.isRunning = false
	s.mutex.Unlock()

	logger.L.Info("Stopping worker scaler")
	close(s.stopChan)

	return nil
}

// monitorLoop 监控循环
func (s *WorkerScaler) monitorLoop(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-s.stopChan:
			return
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.collectMetrics(ctx)
		}
	}
}

// scalingLoop 扩缩容决策循环
func (s *WorkerScaler) scalingLoop(ctx context.Context) {
	ticker := time.NewTicker(60 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-s.stopChan:
			return
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.makeScalingDecisions(ctx)
		}
	}
}

// collectMetrics 收集指标
func (s *WorkerScaler) collectMetrics(ctx context.Context) {
	workers := s.manager.ListWorkers()
	
	s.monitor.mutex.Lock()
	defer s.monitor.mutex.Unlock()

	for name, worker := range workers {
		stats, err := worker.GetStats(ctx)
		if err != nil {
			logger.L.Error("Failed to get worker stats",
				zap.String("worker", name),
				zap.Error(err),
			)
			continue
		}

		// 计算队列长度
		queueLengths := make(map[string]int64)
		for queueName, queueStats := range stats.QueueStats {
			queueLengths[queueName] = queueStats.PendingCount
		}

		// 计算错误率
		errorRate := 0.0
		if stats.TasksProcessed > 0 {
			errorRate = float64(stats.TasksFailed) / float64(stats.TasksProcessed)
		}

		// 计算任务处理速率
		tasksPerSecond := 0.0
		if !stats.StartTime.IsZero() {
			uptime := time.Since(stats.StartTime)
			if uptime.Seconds() > 0 {
				tasksPerSecond = float64(stats.TasksProcessed) / uptime.Seconds()
			}
		}

		metrics := &WorkerMetrics{
			WorkerName:     name,
			WorkerType:     stats.Type,
			CPUUsage:       s.getCPUUsage(name),       // 需要实现CPU使用率获取
			MemoryUsage:    s.getMemoryUsage(name),    // 需要实现内存使用率获取
			TasksPerSecond: tasksPerSecond,
			AverageLatency: stats.AverageLatency,
			QueueLengths:   queueLengths,
			ErrorRate:      errorRate,
			LastUpdate:     time.Now(),
			HealthStatus:   s.getHealthStatus(stats),
		}

		s.monitor.metrics[name] = metrics
	}

	logger.L.Debug("Metrics collected",
		zap.Int("worker_count", len(workers)),
	)
}

// makeScalingDecisions 做出扩缩容决策
func (s *WorkerScaler) makeScalingDecisions(ctx context.Context) {
	s.mutex.RLock()
	policies := make(map[string]*ScalingPolicy)
	for k, v := range s.scalingPolicies {
		policies[k] = v
	}
	s.mutex.RUnlock()

	for workerType, policy := range policies {
		decision := s.evaluateScalingPolicy(ctx, workerType, policy)
		
		if decision.Action != "no_action" {
			logger.L.Info("Scaling decision made",
				zap.String("worker_type", decision.WorkerType),
				zap.String("action", decision.Action),
				zap.Int("current_count", decision.CurrentCount),
				zap.Int("target_count", decision.TargetCount),
				zap.String("reason", decision.Reason),
			)

			if err := s.executeScalingDecision(ctx, decision); err != nil {
				logger.L.Error("Failed to execute scaling decision",
					zap.String("worker_type", decision.WorkerType),
					zap.String("action", decision.Action),
					zap.Error(err),
				)
			}
		}
	}
}

// evaluateScalingPolicy 评估扩缩容策略
func (s *WorkerScaler) evaluateScalingPolicy(ctx context.Context, workerType string, policy *ScalingPolicy) *ScalingDecision {
	decision := &ScalingDecision{
		WorkerType:   workerType,
		Action:       "no_action",
		Timestamp:    time.Now(),
	}

	// 获取当前该类型的Worker数量
	currentCount := s.getWorkerCountByType(workerType)
	decision.CurrentCount = currentCount

	// 获取该类型Worker的平均指标
	avgMetrics := s.getAverageMetricsByType(workerType)
	if avgMetrics == nil {
		return decision
	}

	// 计算队列压力
	totalQueueLength := int64(0)
	for _, length := range avgMetrics.QueueLengths {
		totalQueueLength += length
	}

	// 检查是否需要扩容
	shouldScaleUp := false
	scaleUpReason := ""

	if totalQueueLength > policy.TargetQueueLength {
		shouldScaleUp = true
		scaleUpReason = fmt.Sprintf("queue length %d exceeds target %d", totalQueueLength, policy.TargetQueueLength)
	} else if avgMetrics.CPUUsage > policy.ScaleUpThreshold {
		shouldScaleUp = true
		scaleUpReason = fmt.Sprintf("CPU usage %.2f%% exceeds threshold %.2f%%", avgMetrics.CPUUsage*100, policy.ScaleUpThreshold*100)
	} else if avgMetrics.ErrorRate > 0.1 { // 错误率超过10%
		shouldScaleUp = true
		scaleUpReason = fmt.Sprintf("error rate %.2f%% is too high", avgMetrics.ErrorRate*100)
	}

	// 检查是否需要缩容
	shouldScaleDown := false
	scaleDownReason := ""

	if totalQueueLength == 0 && avgMetrics.CPUUsage < policy.ScaleDownThreshold && avgMetrics.TasksPerSecond < 0.1 {
		shouldScaleDown = true
		scaleDownReason = fmt.Sprintf("low utilization: CPU %.2f%%, TPS %.2f", avgMetrics.CPUUsage*100, avgMetrics.TasksPerSecond)
	}

	// 应用扩缩容决策
	if shouldScaleUp && currentCount < policy.MaxInstances {
		// 检查冷却时间
		if time.Since(policy.LastScaleUp) >= policy.ScaleUpCooldown {
			decision.Action = "scale_up"
			decision.TargetCount = currentCount + 1
			decision.Reason = scaleUpReason
			policy.LastScaleUp = time.Now()
		}
	} else if shouldScaleDown && currentCount > policy.MinInstances {
		// 检查冷却时间
		if time.Since(policy.LastScaleDown) >= policy.ScaleDownCooldown {
			decision.Action = "scale_down"
			decision.TargetCount = currentCount - 1
			decision.Reason = scaleDownReason
			policy.LastScaleDown = time.Now()
		}
	}

	return decision
}

// executeScalingDecision 执行扩缩容决策
func (s *WorkerScaler) executeScalingDecision(ctx context.Context, decision *ScalingDecision) error {
	switch decision.Action {
	case "scale_up":
		return s.scaleUp(ctx, decision.WorkerType)
	case "scale_down":
		return s.scaleDown(ctx, decision.WorkerType)
	default:
		return nil
	}
}

// scaleUp 扩容
func (s *WorkerScaler) scaleUp(ctx context.Context, workerType string) error {
	// 生成新的Worker名称
	workerName := fmt.Sprintf("%s-auto-%d", workerType, time.Now().Unix())
	
	// 创建Worker配置
	config := &WorkerConfig{
		Name:        workerName,
		Type:        workerType,
		Concurrency: 2, // 默认并发数
		Queues:      getDefaultQueues(workerType),
		HealthPort:  8081 + len(s.manager.ListWorkers()), // 动态分配端口
	}

	// 创建并启动Worker
	worker, err := s.manager.CreateWorker(ctx, config)
	if err != nil {
		return fmt.Errorf("failed to create worker: %w", err)
	}

	if err := worker.Start(ctx); err != nil {
		return fmt.Errorf("failed to start worker: %w", err)
	}

	logger.L.Info("Worker scaled up successfully",
		zap.String("worker_name", workerName),
		zap.String("worker_type", workerType),
	)

	return nil
}

// scaleDown 缩容
func (s *WorkerScaler) scaleDown(ctx context.Context, workerType string) error {
	// 找到该类型的一个Worker进行缩容
	workers := s.manager.ListWorkers()
	
	var targetWorker *WorkerInstance
	var targetName string
	
	for name, worker := range workers {
		if worker.GetConfig().Type == workerType {
			// 优先选择负载较低的Worker
			stats, err := worker.GetStats(ctx)
			if err != nil {
				continue
			}
			
			// 检查是否有正在处理的任务
			if stats.TasksProcessed == stats.TasksSucceeded + stats.TasksFailed {
				targetWorker = worker
				targetName = name
				break
			}
		}
	}

	if targetWorker == nil {
		return fmt.Errorf("no suitable worker found for scale down")
	}

	// 停止Worker
	if err := targetWorker.Stop(ctx); err != nil {
		return fmt.Errorf("failed to stop worker: %w", err)
	}

	// 从管理器中移除
	if err := s.manager.RemoveWorker(targetName); err != nil {
		return fmt.Errorf("failed to remove worker: %w", err)
	}

	logger.L.Info("Worker scaled down successfully",
		zap.String("worker_name", targetName),
		zap.String("worker_type", workerType),
	)

	return nil
}

// getWorkerCountByType 获取指定类型的Worker数量
func (s *WorkerScaler) getWorkerCountByType(workerType string) int {
	workers := s.manager.ListWorkers()
	count := 0
	
	for _, worker := range workers {
		if worker.GetConfig().Type == workerType {
			count++
		}
	}
	
	return count
}

// getAverageMetricsByType 获取指定类型Worker的平均指标
func (s *WorkerScaler) getAverageMetricsByType(workerType string) *WorkerMetrics {
	s.monitor.mutex.RLock()
	defer s.monitor.mutex.RUnlock()

	var typeMetrics []*WorkerMetrics
	for _, metrics := range s.monitor.metrics {
		if metrics.WorkerType == workerType {
			typeMetrics = append(typeMetrics, metrics)
		}
	}

	if len(typeMetrics) == 0 {
		return nil
	}

	// 计算平均值
	avgMetrics := &WorkerMetrics{
		WorkerType:   workerType,
		QueueLengths: make(map[string]int64),
	}

	totalCPU := 0.0
	totalMemory := 0.0
	totalTPS := 0.0
	totalLatency := time.Duration(0)
	totalErrorRate := 0.0

	for _, metrics := range typeMetrics {
		totalCPU += metrics.CPUUsage
		totalMemory += metrics.MemoryUsage
		totalTPS += metrics.TasksPerSecond
		totalLatency += metrics.AverageLatency
		totalErrorRate += metrics.ErrorRate

		// 合并队列长度
		for queue, length := range metrics.QueueLengths {
			avgMetrics.QueueLengths[queue] += length
		}
	}

	count := float64(len(typeMetrics))
	avgMetrics.CPUUsage = totalCPU / count
	avgMetrics.MemoryUsage = totalMemory / count
	avgMetrics.TasksPerSecond = totalTPS / count
	avgMetrics.AverageLatency = time.Duration(int64(totalLatency) / int64(count))
	avgMetrics.ErrorRate = totalErrorRate / count

	return avgMetrics
}

// setDefaultPolicies 设置默认扩缩容策略
func (s *WorkerScaler) setDefaultPolicies() {
	policies := map[string]*ScalingPolicy{
		"ai-analysis": {
			WorkerType:         "ai-analysis",
			MinInstances:       1,
			MaxInstances:       5,
			TargetQueueLength:  10,
			ScaleUpThreshold:   0.7,
			ScaleDownThreshold: 0.3,
			ScaleUpCooldown:    5 * time.Minute,
			ScaleDownCooldown:  10 * time.Minute,
			CheckInterval:      1 * time.Minute,
		},
		"notification": {
			WorkerType:         "notification",
			MinInstances:       1,
			MaxInstances:       3,
			TargetQueueLength:  20,
			ScaleUpThreshold:   0.8,
			ScaleDownThreshold: 0.2,
			ScaleUpCooldown:    3 * time.Minute,
			ScaleDownCooldown:  10 * time.Minute,
			CheckInterval:      1 * time.Minute,
		},
		"config-sync": {
			WorkerType:         "config-sync",
			MinInstances:       1,
			MaxInstances:       2,
			TargetQueueLength:  5,
			ScaleUpThreshold:   0.6,
			ScaleDownThreshold: 0.1,
			ScaleUpCooldown:    10 * time.Minute,
			ScaleDownCooldown:  15 * time.Minute,
			CheckInterval:      2 * time.Minute,
		},
	}

	s.mutex.Lock()
	s.scalingPolicies = policies
	s.mutex.Unlock()

	logger.L.Info("Default scaling policies set",
		zap.Int("policy_count", len(policies)),
	)
}

// getCPUUsage 获取CPU使用率（需要实现）
func (s *WorkerScaler) getCPUUsage(workerName string) float64 {
	// 这里应该实现实际的CPU使用率获取逻辑
	// 暂时返回模拟值
	return 0.5
}

// getMemoryUsage 获取内存使用率（需要实现）
func (s *WorkerScaler) getMemoryUsage(workerName string) float64 {
	// 这里应该实现实际的内存使用率获取逻辑
	// 暂时返回模拟值
	return 0.3
}

// getHealthStatus 获取健康状态
func (s *WorkerScaler) getHealthStatus(stats *WorkerStats) string {
	if stats.Status != "running" {
		return "unhealthy"
	}

	// 检查错误率
	if stats.TasksProcessed > 0 {
		errorRate := float64(stats.TasksFailed) / float64(stats.TasksProcessed)
		if errorRate > 0.1 {
			return "degraded"
		}
	}

	// 检查最后活动时间
	if time.Since(stats.LastActivity) > 5*time.Minute {
		return "idle"
	}

	return "healthy"
}

// GetMetrics 获取所有Worker指标
func (s *WorkerScaler) GetMetrics() map[string]*WorkerMetrics {
	s.monitor.mutex.RLock()
	defer s.monitor.mutex.RUnlock()

	result := make(map[string]*WorkerMetrics)
	for k, v := range s.monitor.metrics {
		result[k] = v
	}

	return result
}

// GetScalingPolicies 获取扩缩容策略
func (s *WorkerScaler) GetScalingPolicies() map[string]*ScalingPolicy {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	result := make(map[string]*ScalingPolicy)
	for k, v := range s.scalingPolicies {
		result[k] = v
	}

	return result
}

// UpdateScalingPolicy 更新扩缩容策略
func (s *WorkerScaler) UpdateScalingPolicy(workerType string, policy *ScalingPolicy) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.scalingPolicies[workerType] = policy
	
	logger.L.Info("Scaling policy updated",
		zap.String("worker_type", workerType),
		zap.Int("min_instances", policy.MinInstances),
		zap.Int("max_instances", policy.MaxInstances),
	)
}

// getDefaultQueues 根据worker类型返回默认队列
func getDefaultQueues(workerType string) []string {
	switch workerType {
	case "ai-analysis":
		return []string{string(queue.TaskTypeAIAnalysis)}
	case "notification":
		return []string{string(queue.TaskTypeNotification)}
	case "config-sync":
		return []string{string(queue.TaskTypeConfigSync)}
	case "general":
		return []string{
			string(queue.TaskTypeRuleUpdate),
			string(queue.TaskTypeHealthCheck),
		}
	default:
		return []string{
			string(queue.TaskTypeAIAnalysis),
			string(queue.TaskTypeNotification),
			string(queue.TaskTypeConfigSync),
		}
	}
}