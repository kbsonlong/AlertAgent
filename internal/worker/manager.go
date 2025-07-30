package worker

import (
	"context"
	"fmt"
	"sync"

	"alert_agent/internal/pkg/logger"
	"alert_agent/internal/pkg/queue"
	"alert_agent/internal/pkg/redis"

	"go.uber.org/zap"
)

// WorkerManager Worker管理器
type WorkerManager struct {
	workers map[string]*WorkerInstance
	mutex   sync.RWMutex
}

// WorkerConfig Worker配置
type WorkerConfig struct {
	Name        string   `json:"name"`
	Type        string   `json:"type"`
	Concurrency int      `json:"concurrency"`
	Queues      []string `json:"queues"`
	HealthPort  int      `json:"health_port"`
}

// NewWorkerManager 创建Worker管理器
func NewWorkerManager() *WorkerManager {
	return &WorkerManager{
		workers: make(map[string]*WorkerInstance),
	}
}

// CreateWorker 创建Worker实例
func (m *WorkerManager) CreateWorker(ctx context.Context, config *WorkerConfig) (*WorkerInstance, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// 检查Worker是否已存在
	if _, exists := m.workers[config.Name]; exists {
		return nil, fmt.Errorf("worker %s already exists", config.Name)
	}

	// 创建消息队列
	messageQueue := queue.NewRedisMessageQueue(redis.Client, "alertagent")
	
	// 创建队列监控器
	monitor := queue.NewQueueMonitor(messageQueue, redis.Client, "alertagent")

	// 创建Worker实例
	worker := &WorkerInstance{
		config:       config,
		messageQueue: messageQueue,
		monitor:      monitor,
		handlers:     make(map[queue.TaskType]queue.TaskHandler),
		isRunning:    false,
	}

	// 注册任务处理器
	if err := worker.registerHandlers(); err != nil {
		return nil, fmt.Errorf("failed to register handlers: %w", err)
	}

	// 创建健康检查服务器
	healthServer := NewHealthServer(worker, config.HealthPort)
	worker.healthServer = healthServer

	m.workers[config.Name] = worker

	logger.L.Info("Worker created successfully",
		zap.String("name", config.Name),
		zap.String("type", config.Type),
		zap.Int("concurrency", config.Concurrency),
		zap.Strings("queues", config.Queues),
	)

	return worker, nil
}

// GetWorker 获取Worker实例
func (m *WorkerManager) GetWorker(name string) (*WorkerInstance, bool) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	
	worker, exists := m.workers[name]
	return worker, exists
}

// ListWorkers 列出所有Worker
func (m *WorkerManager) ListWorkers() map[string]*WorkerInstance {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	
	result := make(map[string]*WorkerInstance)
	for name, worker := range m.workers {
		result[name] = worker
	}
	return result
}

// RemoveWorker 移除Worker实例
func (m *WorkerManager) RemoveWorker(name string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	worker, exists := m.workers[name]
	if !exists {
		return fmt.Errorf("worker %s not found", name)
	}

	// 确保Worker已停止
	if worker.isRunning {
		return fmt.Errorf("worker %s is still running, stop it first", name)
	}

	delete(m.workers, name)
	
	logger.L.Info("Worker removed", zap.String("name", name))
	return nil
}

// StopAllWorkers 停止所有Worker
func (m *WorkerManager) StopAllWorkers(ctx context.Context) error {
	m.mutex.RLock()
	workers := make([]*WorkerInstance, 0, len(m.workers))
	for _, worker := range m.workers {
		workers = append(workers, worker)
	}
	m.mutex.RUnlock()

	var errors []error
	for _, worker := range workers {
		if err := worker.Stop(ctx); err != nil {
			errors = append(errors, fmt.Errorf("failed to stop worker %s: %w", worker.config.Name, err))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("errors stopping workers: %v", errors)
	}

	logger.L.Info("All workers stopped successfully")
	return nil
}

// GetWorkerStats 获取Worker统计信息
func (m *WorkerManager) GetWorkerStats(ctx context.Context) (map[string]*WorkerStats, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	stats := make(map[string]*WorkerStats)
	for name, worker := range m.workers {
		workerStats, err := worker.GetStats(ctx)
		if err != nil {
			logger.L.Error("Failed to get worker stats",
				zap.String("worker", name),
				zap.Error(err),
			)
			continue
		}
		stats[name] = workerStats
	}

	return stats, nil
}