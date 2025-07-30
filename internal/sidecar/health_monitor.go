package sidecar

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"alert_agent/internal/pkg/logger"

	"go.uber.org/zap"
)

// HealthStatus 健康状态
type HealthStatus struct {
	Status        string            `json:"status"`
	LastSync      time.Time         `json:"last_sync"`
	LastError     string            `json:"last_error,omitempty"`
	SyncCount     int64             `json:"sync_count"`
	ErrorCount    int64             `json:"error_count"`
	Uptime        time.Duration     `json:"uptime"`
	ConfigHash    string            `json:"config_hash"`
	ConfigSize    int64             `json:"config_size"`
	Metrics       map[string]interface{} `json:"metrics"`
	StartTime     time.Time         `json:"start_time"`
}

// HealthMonitor 健康监控器
type HealthMonitor struct {
	syncer      *ConfigSyncer
	status      *HealthStatus
	mutex       sync.RWMutex
	httpServer  *http.Server
	port        int
	startTime   time.Time
	retryPolicy *RetryPolicy
}

// RetryPolicy 重试策略
type RetryPolicy struct {
	MaxRetries      int           `json:"max_retries"`
	InitialDelay    time.Duration `json:"initial_delay"`
	MaxDelay        time.Duration `json:"max_delay"`
	BackoffFactor   float64       `json:"backoff_factor"`
	RetryableErrors []string      `json:"retryable_errors"`
}

// NewHealthMonitor 创建健康监控器
func NewHealthMonitor(syncer *ConfigSyncer, port int) *HealthMonitor {
	startTime := time.Now()
	
	return &HealthMonitor{
		syncer:    syncer,
		port:      port,
		startTime: startTime,
		status: &HealthStatus{
			Status:     "starting",
			StartTime:  startTime,
			SyncCount:  0,
			ErrorCount: 0,
			Metrics:    make(map[string]interface{}),
		},
		retryPolicy: &RetryPolicy{
			MaxRetries:    5,
			InitialDelay:  1 * time.Second,
			MaxDelay:      30 * time.Second,
			BackoffFactor: 2.0,
			RetryableErrors: []string{
				"connection refused",
				"timeout",
				"temporary failure",
				"network unreachable",
			},
		},
	}
}

// Start 启动健康监控器
func (hm *HealthMonitor) Start(ctx context.Context) error {
	// 启动HTTP健康检查服务器
	if err := hm.startHealthServer(ctx); err != nil {
		return fmt.Errorf("failed to start health server: %w", err)
	}

	// 启动状态上报协程
	go hm.startStatusReporting(ctx)

	// 启动指标收集协程
	go hm.startMetricsCollection(ctx)

	logger.L.Info("Health monitor started",
		zap.Int("port", hm.port),
	)

	return nil
}

// startHealthServer 启动健康检查HTTP服务器
func (hm *HealthMonitor) startHealthServer(ctx context.Context) error {
	mux := http.NewServeMux()
	
	// 健康检查端点
	mux.HandleFunc("/health", hm.handleHealth)
	mux.HandleFunc("/health/ready", hm.handleReady)
	mux.HandleFunc("/health/live", hm.handleLive)
	mux.HandleFunc("/metrics", hm.handleMetrics)
	mux.HandleFunc("/status", hm.handleStatus)

	hm.httpServer = &http.Server{
		Addr:    fmt.Sprintf(":%d", hm.port),
		Handler: mux,
	}

	go func() {
		if err := hm.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.L.Error("Health server failed", zap.Error(err))
		}
	}()

	// 等待服务器启动
	time.Sleep(100 * time.Millisecond)

	return nil
}

// handleHealth 处理健康检查请求
func (hm *HealthMonitor) handleHealth(w http.ResponseWriter, r *http.Request) {
	hm.mutex.RLock()
	status := *hm.status
	hm.mutex.RUnlock()

	status.Uptime = time.Since(hm.startTime)

	w.Header().Set("Content-Type", "application/json")
	
	// 根据状态设置HTTP状态码
	if status.Status == "healthy" {
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusServiceUnavailable)
	}

	json.NewEncoder(w).Encode(status)
}

// handleReady 处理就绪检查请求
func (hm *HealthMonitor) handleReady(w http.ResponseWriter, r *http.Request) {
	hm.mutex.RLock()
	status := hm.status.Status
	hm.mutex.RUnlock()

	if status == "healthy" || status == "syncing" {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ready"))
	} else {
		w.WriteHeader(http.StatusServiceUnavailable)
		w.Write([]byte("not ready"))
	}
}

// handleLive 处理存活检查请求
func (hm *HealthMonitor) handleLive(w http.ResponseWriter, r *http.Request) {
	// 只要进程在运行就认为是存活的
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("alive"))
}

// handleMetrics 处理指标请求
func (hm *HealthMonitor) handleMetrics(w http.ResponseWriter, r *http.Request) {
	hm.mutex.RLock()
	metrics := make(map[string]interface{})
	for k, v := range hm.status.Metrics {
		metrics[k] = v
	}
	hm.mutex.RUnlock()

	// 添加基本指标
	metrics["sync_count_total"] = hm.status.SyncCount
	metrics["error_count_total"] = hm.status.ErrorCount
	metrics["uptime_seconds"] = time.Since(hm.startTime).Seconds()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(metrics)
}

// handleStatus 处理状态请求
func (hm *HealthMonitor) handleStatus(w http.ResponseWriter, r *http.Request) {
	hm.handleHealth(w, r)
}

// UpdateSyncSuccess 更新同步成功状态
func (hm *HealthMonitor) UpdateSyncSuccess(configHash string, configSize int64) {
	hm.mutex.Lock()
	defer hm.mutex.Unlock()

	hm.status.Status = "healthy"
	hm.status.LastSync = time.Now()
	hm.status.LastError = ""
	hm.status.SyncCount++
	hm.status.ConfigHash = configHash
	hm.status.ConfigSize = configSize

	logger.L.Debug("Sync success recorded",
		zap.String("config_hash", configHash),
		zap.Int64("config_size", configSize),
		zap.Int64("sync_count", hm.status.SyncCount),
	)
}

// UpdateSyncError 更新同步错误状态
func (hm *HealthMonitor) UpdateSyncError(err error) {
	hm.mutex.Lock()
	defer hm.mutex.Unlock()

	hm.status.Status = "error"
	hm.status.LastError = err.Error()
	hm.status.ErrorCount++

	logger.L.Error("Sync error recorded",
		zap.Error(err),
		zap.Int64("error_count", hm.status.ErrorCount),
	)
}

// UpdateSyncStatus 更新同步状态
func (hm *HealthMonitor) UpdateSyncStatus(status string) {
	hm.mutex.Lock()
	defer hm.mutex.Unlock()

	hm.status.Status = status

	logger.L.Debug("Sync status updated",
		zap.String("status", status),
	)
}

// startStatusReporting 启动状态上报
func (hm *HealthMonitor) startStatusReporting(ctx context.Context) {
	ticker := time.NewTicker(60 * time.Second) // 每分钟上报一次
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			hm.reportStatus(ctx)
		}
	}
}

// reportStatus 上报状态到AlertAgent
func (hm *HealthMonitor) reportStatus(ctx context.Context) {
	hm.mutex.RLock()
	status := *hm.status
	hm.mutex.RUnlock()

	// 构建状态报告
	report := map[string]interface{}{
		"cluster_id":    hm.syncer.config.ClusterID,
		"config_type":   hm.syncer.config.ConfigType,
		"status":        status.Status,
		"last_sync":     status.LastSync.Unix(),
		"sync_count":    status.SyncCount,
		"error_count":   status.ErrorCount,
		"uptime":        time.Since(hm.startTime).Seconds(),
		"config_hash":   status.ConfigHash,
		"config_size":   status.ConfigSize,
		"last_error":    status.LastError,
	}

	// 这里可以发送到AlertAgent的状态收集接口
	logger.L.Debug("Status report prepared",
		zap.Any("report", report),
	)
}

// startMetricsCollection 启动指标收集
func (hm *HealthMonitor) startMetricsCollection(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second) // 每30秒收集一次指标
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			hm.collectMetrics()
		}
	}
}

// collectMetrics 收集指标
func (hm *HealthMonitor) collectMetrics() {
	hm.mutex.Lock()
	defer hm.mutex.Unlock()

	// 收集系统指标
	hm.status.Metrics["uptime_seconds"] = time.Since(hm.startTime).Seconds()
	hm.status.Metrics["sync_success_rate"] = hm.calculateSuccessRate()
	hm.status.Metrics["last_sync_duration"] = hm.getLastSyncDuration()
	
	// 可以添加更多指标，如内存使用、CPU使用等
}

// calculateSuccessRate 计算成功率
func (hm *HealthMonitor) calculateSuccessRate() float64 {
	total := hm.status.SyncCount + hm.status.ErrorCount
	if total == 0 {
		return 0.0
	}
	return float64(hm.status.SyncCount) / float64(total)
}

// getLastSyncDuration 获取最后一次同步耗时
func (hm *HealthMonitor) getLastSyncDuration() float64 {
	if hm.status.LastSync.IsZero() {
		return 0.0
	}
	// 这里应该记录实际的同步耗时，暂时返回0
	return 0.0
}

// ShouldRetry 判断是否应该重试
func (hm *HealthMonitor) ShouldRetry(err error, attempt int) bool {
	if attempt >= hm.retryPolicy.MaxRetries {
		return false
	}

	errStr := err.Error()
	for _, retryableErr := range hm.retryPolicy.RetryableErrors {
		if contains(errStr, retryableErr) {
			return true
		}
	}

	return false
}

// GetRetryDelay 获取重试延迟
func (hm *HealthMonitor) GetRetryDelay(attempt int) time.Duration {
	delay := time.Duration(float64(hm.retryPolicy.InitialDelay) * 
		pow(hm.retryPolicy.BackoffFactor, float64(attempt)))
	
	if delay > hm.retryPolicy.MaxDelay {
		delay = hm.retryPolicy.MaxDelay
	}
	
	return delay
}

// Stop 停止健康监控器
func (hm *HealthMonitor) Stop(ctx context.Context) error {
	if hm.httpServer != nil {
		return hm.httpServer.Shutdown(ctx)
	}
	return nil
}

// 辅助函数
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || 
		(len(s) > len(substr) && 
			(s[:len(substr)] == substr || 
			 s[len(s)-len(substr):] == substr ||
			 indexOf(s, substr) >= 0)))
}

func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

func pow(base, exp float64) float64 {
	result := 1.0
	for i := 0; i < int(exp); i++ {
		result *= base
	}
	return result
}