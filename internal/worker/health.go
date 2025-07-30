package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"alert_agent/internal/pkg/logger"

	"go.uber.org/zap"
)

// HealthServer 健康检查服务器
type HealthServer struct {
	worker *WorkerInstance
	server *http.Server
	port   int
}

// HealthResponse 健康检查响应
type HealthResponse struct {
	Status    string                 `json:"status"`
	Timestamp time.Time              `json:"timestamp"`
	Worker    *WorkerHealthInfo      `json:"worker"`
	Queues    map[string]interface{} `json:"queues,omitempty"`
}

// WorkerHealthInfo Worker健康信息
type WorkerHealthInfo struct {
	Name        string        `json:"name"`
	Type        string        `json:"type"`
	Status      string        `json:"status"`
	Uptime      time.Duration `json:"uptime"`
	Concurrency int           `json:"concurrency"`
	Queues      []string      `json:"queues"`
}

// NewHealthServer 创建健康检查服务器
func NewHealthServer(worker *WorkerInstance, port int) *HealthServer {
	return &HealthServer{
		worker: worker,
		port:   port,
	}
}

// Start 启动健康检查服务器
func (h *HealthServer) Start() error {
	mux := http.NewServeMux()
	
	// 健康检查端点
	mux.HandleFunc("/health", h.healthHandler)
	mux.HandleFunc("/health/live", h.livenessHandler)
	mux.HandleFunc("/health/ready", h.readinessHandler)
	mux.HandleFunc("/stats", h.statsHandler)
	mux.HandleFunc("/metrics", h.metricsHandler)

	h.server = &http.Server{
		Addr:    fmt.Sprintf(":%d", h.port),
		Handler: mux,
	}

	go func() {
		logger.L.Info("Health server starting",
			zap.Int("port", h.port),
			zap.String("worker", h.worker.config.Name),
		)
		
		if err := h.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.L.Error("Health server failed", zap.Error(err))
		}
	}()

	return nil
}

// Stop 停止健康检查服务器
func (h *HealthServer) Stop(ctx context.Context) error {
	if h.server == nil {
		return nil
	}

	logger.L.Info("Stopping health server", zap.Int("port", h.port))
	return h.server.Shutdown(ctx)
}

// healthHandler 通用健康检查处理器
func (h *HealthServer) healthHandler(w http.ResponseWriter, r *http.Request) {
	stats, err := h.worker.GetStats(r.Context())
	if err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, "Failed to get worker stats", err)
		return
	}

	response := &HealthResponse{
		Status:    "healthy",
		Timestamp: time.Now(),
		Worker: &WorkerHealthInfo{
			Name:        stats.Name,
			Type:        stats.Type,
			Status:      stats.Status,
			Uptime:      time.Since(stats.StartTime),
			Concurrency: stats.Concurrency,
			Queues:      stats.Queues,
		},
		Queues: make(map[string]interface{}),
	}

	// 添加队列信息
	for queueName, queueStats := range stats.QueueStats {
		response.Queues[queueName] = map[string]interface{}{
			"pending":     queueStats.PendingCount,
			"processing":  queueStats.ProcessingCount,
			"completed":   queueStats.CompletedCount,
			"failed":      queueStats.FailedCount,
			"dead_letter": queueStats.DeadLetterCount,
		}
	}

	// 检查Worker状态
	if !h.worker.IsRunning() {
		response.Status = "unhealthy"
		w.WriteHeader(http.StatusServiceUnavailable)
	}

	h.writeJSONResponse(w, response)
}

// livenessHandler 存活检查处理器
func (h *HealthServer) livenessHandler(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"status":    "alive",
		"timestamp": time.Now(),
		"worker":    h.worker.config.Name,
	}

	h.writeJSONResponse(w, response)
}

// readinessHandler 就绪检查处理器
func (h *HealthServer) readinessHandler(w http.ResponseWriter, r *http.Request) {
	isReady := h.worker.IsRunning()
	
	status := "ready"
	statusCode := http.StatusOK
	
	if !isReady {
		status = "not_ready"
		statusCode = http.StatusServiceUnavailable
	}

	response := map[string]interface{}{
		"status":    status,
		"timestamp": time.Now(),
		"worker":    h.worker.config.Name,
		"ready":     isReady,
	}

	w.WriteHeader(statusCode)
	h.writeJSONResponse(w, response)
}

// statsHandler 统计信息处理器
func (h *HealthServer) statsHandler(w http.ResponseWriter, r *http.Request) {
	stats, err := h.worker.GetStats(r.Context())
	if err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, "Failed to get worker stats", err)
		return
	}

	h.writeJSONResponse(w, stats)
}

// metricsHandler Prometheus指标处理器
func (h *HealthServer) metricsHandler(w http.ResponseWriter, r *http.Request) {
	stats, err := h.worker.GetStats(r.Context())
	if err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, "Failed to get worker stats", err)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	
	// 输出Prometheus格式的指标
	fmt.Fprintf(w, "# HELP alertagent_worker_tasks_processed_total Total number of tasks processed\n")
	fmt.Fprintf(w, "# TYPE alertagent_worker_tasks_processed_total counter\n")
	fmt.Fprintf(w, "alertagent_worker_tasks_processed_total{worker=\"%s\",type=\"%s\"} %d\n", 
		stats.Name, stats.Type, stats.TasksProcessed)

	fmt.Fprintf(w, "# HELP alertagent_worker_tasks_succeeded_total Total number of tasks succeeded\n")
	fmt.Fprintf(w, "# TYPE alertagent_worker_tasks_succeeded_total counter\n")
	fmt.Fprintf(w, "alertagent_worker_tasks_succeeded_total{worker=\"%s\",type=\"%s\"} %d\n", 
		stats.Name, stats.Type, stats.TasksSucceeded)

	fmt.Fprintf(w, "# HELP alertagent_worker_tasks_failed_total Total number of tasks failed\n")
	fmt.Fprintf(w, "# TYPE alertagent_worker_tasks_failed_total counter\n")
	fmt.Fprintf(w, "alertagent_worker_tasks_failed_total{worker=\"%s\",type=\"%s\"} %d\n", 
		stats.Name, stats.Type, stats.TasksFailed)

	fmt.Fprintf(w, "# HELP alertagent_worker_average_latency_seconds Average task processing latency\n")
	fmt.Fprintf(w, "# TYPE alertagent_worker_average_latency_seconds gauge\n")
	fmt.Fprintf(w, "alertagent_worker_average_latency_seconds{worker=\"%s\",type=\"%s\"} %f\n", 
		stats.Name, stats.Type, stats.AverageLatency.Seconds())

	fmt.Fprintf(w, "# HELP alertagent_worker_uptime_seconds Worker uptime in seconds\n")
	fmt.Fprintf(w, "# TYPE alertagent_worker_uptime_seconds gauge\n")
	fmt.Fprintf(w, "alertagent_worker_uptime_seconds{worker=\"%s\",type=\"%s\"} %f\n", 
		stats.Name, stats.Type, time.Since(stats.StartTime).Seconds())

	// 队列指标
	for queueName, queueStats := range stats.QueueStats {
		fmt.Fprintf(w, "# HELP alertagent_queue_pending_tasks Number of pending tasks in queue\n")
		fmt.Fprintf(w, "# TYPE alertagent_queue_pending_tasks gauge\n")
		fmt.Fprintf(w, "alertagent_queue_pending_tasks{worker=\"%s\",queue=\"%s\"} %d\n", 
			stats.Name, queueName, queueStats.PendingCount)

		fmt.Fprintf(w, "# HELP alertagent_queue_processing_tasks Number of processing tasks in queue\n")
		fmt.Fprintf(w, "# TYPE alertagent_queue_processing_tasks gauge\n")
		fmt.Fprintf(w, "alertagent_queue_processing_tasks{worker=\"%s\",queue=\"%s\"} %d\n", 
			stats.Name, queueName, queueStats.ProcessingCount)

		fmt.Fprintf(w, "# HELP alertagent_queue_dead_letter_tasks Number of dead letter tasks in queue\n")
		fmt.Fprintf(w, "# TYPE alertagent_queue_dead_letter_tasks gauge\n")
		fmt.Fprintf(w, "alertagent_queue_dead_letter_tasks{worker=\"%s\",queue=\"%s\"} %d\n", 
			stats.Name, queueName, queueStats.DeadLetterCount)
	}
}

// writeJSONResponse 写入JSON响应
func (h *HealthServer) writeJSONResponse(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	
	if err := json.NewEncoder(w).Encode(data); err != nil {
		logger.L.Error("Failed to encode JSON response", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

// writeErrorResponse 写入错误响应
func (h *HealthServer) writeErrorResponse(w http.ResponseWriter, statusCode int, message string, err error) {
	logger.L.Error(message, zap.Error(err))
	
	response := map[string]interface{}{
		"status":    "error",
		"message":   message,
		"timestamp": time.Now(),
	}
	
	if err != nil {
		response["error"] = err.Error()
	}

	w.WriteHeader(statusCode)
	h.writeJSONResponse(w, response)
}