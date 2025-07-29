package configsyncer

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"go.uber.org/zap"
)

// HTTPServer HTTP服务器配置
type HTTPServer struct {
	syncer *ConfigSyncer
	server *http.Server
	logger *zap.Logger
	port   int
}

// NewHTTPServer 创建新的HTTP服务器
func NewHTTPServer(syncer *ConfigSyncer, port int) *HTTPServer {
	return &HTTPServer{
		syncer: syncer,
		logger: syncer.config.Logger,
		port:   port,
	}
}

// Start 启动HTTP服务器
func (s *HTTPServer) Start(ctx context.Context) error {
	mux := http.NewServeMux()

	// 健康检查端点
	mux.HandleFunc("/health", s.healthHandler)
	mux.HandleFunc("/healthz", s.healthHandler)
	mux.HandleFunc("/ready", s.readinessHandler)

	// 指标端点
	mux.HandleFunc("/metrics", s.metricsHandler)
	mux.HandleFunc("/status", s.statusHandler)

	// 根路径
	mux.HandleFunc("/", s.rootHandler)

	s.server = &http.Server{
		Addr:    fmt.Sprintf(":%d", s.port),
		Handler: s.loggingMiddleware(mux),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	s.logger.Info("Starting HTTP server", zap.Int("port", s.port))

	// 启动服务器
	go func() {
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			s.logger.Error("HTTP server failed", zap.Error(err))
		}
	}()

	// 等待上下文取消
	<-ctx.Done()

	// 优雅关闭
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	s.logger.Info("Shutting down HTTP server")
	return s.server.Shutdown(shutdownCtx)
}

// healthHandler 健康检查处理器
func (s *HTTPServer) healthHandler(w http.ResponseWriter, r *http.Request) {
	health := s.syncer.GetHealthStatus()

	w.Header().Set("Content-Type", "application/json")

	if health.Status == "healthy" {
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusServiceUnavailable)
	}

	if err := json.NewEncoder(w).Encode(health); err != nil {
		s.logger.Error("Failed to encode health response", zap.Error(err))
	}
}

// readinessHandler 就绪检查处理器
func (s *HTTPServer) readinessHandler(w http.ResponseWriter, r *http.Request) {
	// 检查是否已经进行过至少一次同步
	metrics := s.syncer.GetMetrics()
	isReady := metrics.SyncCount > 0 || time.Since(s.syncer.startTime) < 2*time.Minute

	w.Header().Set("Content-Type", "application/json")

	response := map[string]interface{}{
		"ready":      isReady,
		"timestamp":  time.Now(),
		"sync_count": metrics.SyncCount,
		"uptime":     metrics.Uptime,
	}

	if isReady {
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusServiceUnavailable)
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		s.logger.Error("Failed to encode readiness response", zap.Error(err))
	}
}

// metricsHandler 指标处理器
func (s *HTTPServer) metricsHandler(w http.ResponseWriter, r *http.Request) {
	metrics := s.syncer.GetMetrics()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(metrics); err != nil {
		s.logger.Error("Failed to encode metrics response", zap.Error(err))
	}
}

// statusHandler 状态处理器
func (s *HTTPServer) statusHandler(w http.ResponseWriter, r *http.Request) {
	health := s.syncer.GetHealthStatus()
	metrics := s.syncer.GetMetrics()

	status := map[string]interface{}{
		"health":  health,
		"metrics": metrics,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(status); err != nil {
		s.logger.Error("Failed to encode status response", zap.Error(err))
	}
}

// rootHandler 根路径处理器
func (s *HTTPServer) rootHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	info := map[string]interface{}{
		"service":     "config-syncer",
		"version":     "1.0.0",
		"cluster_id":  s.syncer.config.ClusterID,
		"config_type": s.syncer.config.ConfigType,
		"endpoints": map[string]string{
			"health":  "/health",
			"ready":   "/ready",
			"metrics": "/metrics",
			"status":  "/status",
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(info); err != nil {
		s.logger.Error("Failed to encode root response", zap.Error(err))
	}
}

// loggingMiddleware 日志中间件
func (s *HTTPServer) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// 创建响应写入器包装器来捕获状态码
		wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		next.ServeHTTP(wrapped, r)

		duration := time.Since(start)

		s.logger.Info("HTTP request",
			zap.String("method", r.Method),
			zap.String("path", r.URL.Path),
			zap.String("remote_addr", r.RemoteAddr),
			zap.String("user_agent", r.UserAgent()),
			zap.Int("status_code", wrapped.statusCode),
			zap.Duration("duration", duration),
		)
	})
}

// responseWriter 响应写入器包装器
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}