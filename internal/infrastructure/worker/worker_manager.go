package worker

import (
	"context"
	"fmt"
	"sync"
	"time"

	"alert_agent/internal/domain/analysis"
	"alert_agent/internal/pkg/logger"

	"go.uber.org/zap"
)

// AnalysisWorkerManagerImpl 分析工作器管理器实现
type AnalysisWorkerManagerImpl struct {
	workers         map[string]analysis.AnalysisWorker
	workerFactory   WorkerFactory
	targetCount     int
	mu              sync.RWMutex
	ctx             context.Context
	cancel          context.CancelFunc
	logger          *zap.Logger
	metrics         map[string]interface{}
	metricsLock     sync.RWMutex
}

// WorkerFactory 工作器工厂接口
type WorkerFactory interface {
	CreateWorker() analysis.AnalysisWorker
}

// DefaultWorkerFactory 默认工作器工厂
type DefaultWorkerFactory struct {
	taskQueue       analysis.AnalysisTaskQueue
	taskRepo        analysis.AnalysisTaskRepository
	resultRepo      analysis.AnalysisResultRepository
	progressTracker analysis.AnalysisProgressTracker
	analysisEngine  analysis.AnalysisEngine
	metricsCollector analysis.AnalysisMetricsCollector
}

// CreateWorker 创建工作器
func (f *DefaultWorkerFactory) CreateWorker() analysis.AnalysisWorker {
	return NewAnalysisWorker(
		f.taskQueue,
		f.taskRepo,
		f.resultRepo,
		f.progressTracker,
		f.analysisEngine,
		f.metricsCollector,
	)
}

// NewDefaultWorkerFactory 创建默认工作器工厂
func NewDefaultWorkerFactory(
	taskQueue analysis.AnalysisTaskQueue,
	taskRepo analysis.AnalysisTaskRepository,
	resultRepo analysis.AnalysisResultRepository,
	progressTracker analysis.AnalysisProgressTracker,
	analysisEngine analysis.AnalysisEngine,
	metricsCollector analysis.AnalysisMetricsCollector,
) WorkerFactory {
	return &DefaultWorkerFactory{
		taskQueue:       taskQueue,
		taskRepo:        taskRepo,
		resultRepo:      resultRepo,
		progressTracker: progressTracker,
		analysisEngine:  analysisEngine,
		metricsCollector: metricsCollector,
	}
}

// NewAnalysisWorkerManager 创建分析工作器管理器
func NewAnalysisWorkerManager(workerFactory WorkerFactory) analysis.AnalysisWorkerManager {
	ctx, cancel := context.WithCancel(context.Background())
	return &AnalysisWorkerManagerImpl{
		workers:       make(map[string]analysis.AnalysisWorker),
		workerFactory: workerFactory,
		targetCount:   0,
		ctx:           ctx,
		cancel:        cancel,
		logger:        logger.L.Named("worker-manager"),
		metrics:       make(map[string]interface{}),
	}
}

// StartWorkers 启动工作器
func (m *AnalysisWorkerManagerImpl) StartWorkers(ctx context.Context, count int) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if count <= 0 {
		return fmt.Errorf("worker count must be positive, got: %d", count)
	}

	m.targetCount = count
	currentCount := len(m.workers)

	if currentCount >= count {
		m.logger.Info("Sufficient workers already running", 
			zap.Int("current", currentCount),
			zap.Int("target", count))
		return nil
	}

	// 启动新的工作器
	neededCount := count - currentCount
	for i := 0; i < neededCount; i++ {
		worker := m.workerFactory.CreateWorker()
		if err := worker.Start(ctx); err != nil {
			m.logger.Error("Failed to start worker", 
				zap.String("worker_id", worker.GetID()),
				zap.Error(err))
			continue
		}

		m.workers[worker.GetID()] = worker
		m.logger.Info("Worker started", 
			zap.String("worker_id", worker.GetID()))
	}

	m.updateMetrics()
	m.logger.Info("Workers started", 
		zap.Int("total_workers", len(m.workers)),
		zap.Int("target_count", count))

	return nil
}

// StopWorkers 停止工作器
func (m *AnalysisWorkerManagerImpl) StopWorkers(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if len(m.workers) == 0 {
		m.logger.Info("No workers to stop")
		return nil
	}

	// 并发停止所有工作器
	var wg sync.WaitGroup
	errorChan := make(chan error, len(m.workers))

	for workerID, worker := range m.workers {
		wg.Add(1)
		go func(id string, w analysis.AnalysisWorker) {
			defer wg.Done()
			if err := w.Stop(ctx); err != nil {
				m.logger.Error("Failed to stop worker", 
					zap.String("worker_id", id),
					zap.Error(err))
				errorChan <- fmt.Errorf("failed to stop worker %s: %w", id, err)
			} else {
				m.logger.Info("Worker stopped", zap.String("worker_id", id))
			}
		}(workerID, worker)
	}

	// 等待所有工作器停止
	wg.Wait()
	close(errorChan)

	// 收集错误
	var errors []error
	for err := range errorChan {
		errors = append(errors, err)
	}

	// 清空工作器映射
	m.workers = make(map[string]analysis.AnalysisWorker)
	m.targetCount = 0
	m.updateMetrics()

	if len(errors) > 0 {
		m.logger.Error("Some workers failed to stop", zap.Int("error_count", len(errors)))
		return fmt.Errorf("failed to stop %d workers", len(errors))
	}

	m.logger.Info("All workers stopped successfully")
	return nil
}

// GetWorkerStatuses 获取所有工作器状态
func (m *AnalysisWorkerManagerImpl) GetWorkerStatuses() []*analysis.WorkerStatus {
	m.mu.RLock()
	defer m.mu.RUnlock()

	statuses := make([]*analysis.WorkerStatus, 0, len(m.workers))
	for _, worker := range m.workers {
		statuses = append(statuses, worker.GetStatus())
	}

	return statuses
}

// ScaleWorkers 动态调整工作器数量
func (m *AnalysisWorkerManagerImpl) ScaleWorkers(ctx context.Context, targetCount int) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if targetCount < 0 {
		return fmt.Errorf("target count cannot be negative: %d", targetCount)
	}

	currentCount := len(m.workers)
	m.targetCount = targetCount

	if targetCount == currentCount {
		m.logger.Info("Worker count already at target", 
			zap.Int("count", currentCount))
		return nil
	}

	if targetCount > currentCount {
		// 需要增加工作器
		neededCount := targetCount - currentCount
		for i := 0; i < neededCount; i++ {
			worker := m.workerFactory.CreateWorker()
			if err := worker.Start(ctx); err != nil {
				m.logger.Error("Failed to start worker during scaling", 
					zap.String("worker_id", worker.GetID()),
					zap.Error(err))
				continue
			}

			m.workers[worker.GetID()] = worker
			m.logger.Info("Worker added during scaling", 
				zap.String("worker_id", worker.GetID()))
		}
	} else {
		// 需要减少工作器
		removeCount := currentCount - targetCount
		var workersToRemove []string

		// 选择要移除的工作器（优先选择不健康的）
		for workerID, worker := range m.workers {
			if len(workersToRemove) >= removeCount {
				break
			}
			if !worker.IsHealthy() {
				workersToRemove = append(workersToRemove, workerID)
			}
		}

		// 如果不健康的工作器不够，随机选择健康的
		for workerID := range m.workers {
			if len(workersToRemove) >= removeCount {
				break
			}
			found := false
			for _, id := range workersToRemove {
				if id == workerID {
					found = true
					break
				}
			}
			if !found {
				workersToRemove = append(workersToRemove, workerID)
			}
		}

		// 停止选中的工作器
		for _, workerID := range workersToRemove {
			worker := m.workers[workerID]
			if err := worker.Stop(ctx); err != nil {
				m.logger.Error("Failed to stop worker during scaling", 
					zap.String("worker_id", workerID),
					zap.Error(err))
			} else {
				m.logger.Info("Worker removed during scaling", 
					zap.String("worker_id", workerID))
			}
			delete(m.workers, workerID)
		}
	}

	m.updateMetrics()
	m.logger.Info("Workers scaled", 
		zap.Int("from", currentCount),
		zap.Int("to", len(m.workers)),
		zap.Int("target", targetCount))

	return nil
}

// GetActiveWorkerCount 获取活跃工作器数量
func (m *AnalysisWorkerManagerImpl) GetActiveWorkerCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	activeCount := 0
	for _, worker := range m.workers {
		if worker.IsHealthy() {
			activeCount++
		}
	}

	return activeCount
}

// RestartWorker 重启指定工作器
func (m *AnalysisWorkerManagerImpl) RestartWorker(ctx context.Context, workerID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	worker, exists := m.workers[workerID]
	if !exists {
		return fmt.Errorf("worker not found: %s", workerID)
	}

	m.logger.Info("Restarting worker", zap.String("worker_id", workerID))

	// 停止现有工作器
	if err := worker.Stop(ctx); err != nil {
		m.logger.Error("Failed to stop worker for restart", 
			zap.String("worker_id", workerID),
			zap.Error(err))
	}

	// 创建新工作器
	newWorker := m.workerFactory.CreateWorker()
	if err := newWorker.Start(ctx); err != nil {
		m.logger.Error("Failed to start new worker", 
			zap.String("new_worker_id", newWorker.GetID()),
			zap.Error(err))
		return fmt.Errorf("failed to start new worker: %w", err)
	}

	// 替换工作器
	delete(m.workers, workerID)
	m.workers[newWorker.GetID()] = newWorker

	m.updateMetrics()
	m.logger.Info("Worker restarted", 
		zap.String("old_worker_id", workerID),
		zap.String("new_worker_id", newWorker.GetID()))

	return nil
}

// GetWorkerMetrics 获取工作器指标
func (m *AnalysisWorkerManagerImpl) GetWorkerMetrics() map[string]interface{} {
	m.metricsLock.RLock()
	defer m.metricsLock.RUnlock()

	// 复制指标以避免并发修改
	metrics := make(map[string]interface{})
	for k, v := range m.metrics {
		metrics[k] = v
	}

	return metrics
}

// updateMetrics 更新指标
func (m *AnalysisWorkerManagerImpl) updateMetrics() {
	m.metricsLock.Lock()
	defer m.metricsLock.Unlock()

	totalWorkers := len(m.workers)
	activeWorkers := 0
	unhealthyWorkers := 0
	totalProcessed := int64(0)
	totalErrors := int64(0)

	for _, worker := range m.workers {
		status := worker.GetStatus()
		if worker.IsHealthy() {
			activeWorkers++
		} else {
			unhealthyWorkers++
		}
		totalProcessed += status.ProcessedCount
		totalErrors += status.ErrorCount
	}

	m.metrics["total_workers"] = totalWorkers
	m.metrics["active_workers"] = activeWorkers
	m.metrics["unhealthy_workers"] = unhealthyWorkers
	m.metrics["target_workers"] = m.targetCount
	m.metrics["total_processed"] = totalProcessed
	m.metrics["total_errors"] = totalErrors
	m.metrics["last_updated"] = time.Now()

	if totalProcessed > 0 {
		m.metrics["error_rate"] = float64(totalErrors) / float64(totalProcessed) * 100
	} else {
		m.metrics["error_rate"] = 0.0
	}
}

// HealthCheck 健康检查
func (m *AnalysisWorkerManagerImpl) HealthCheck() error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if len(m.workers) == 0 {
		return fmt.Errorf("no workers available")
	}

	activeCount := m.GetActiveWorkerCount()
	if activeCount == 0 {
		return fmt.Errorf("no healthy workers available")
	}

	// 检查是否有足够的健康工作器
	healthyRatio := float64(activeCount) / float64(len(m.workers))
	if healthyRatio < 0.5 {
		return fmt.Errorf("too many unhealthy workers: %d/%d healthy", activeCount, len(m.workers))
	}

	return nil
}

// Shutdown 关闭管理器
func (m *AnalysisWorkerManagerImpl) Shutdown(ctx context.Context) error {
	m.logger.Info("Shutting down worker manager")

	// 停止所有工作器
	if err := m.StopWorkers(ctx); err != nil {
		m.logger.Error("Failed to stop workers during shutdown", zap.Error(err))
		return err
	}

	// 取消上下文
	m.cancel()

	m.logger.Info("Worker manager shutdown completed")
	return nil
}