package metrics

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

// MetricsConfig Prometheus指标配置
type MetricsConfig struct {
	Enabled bool   `yaml:"enabled" json:"enabled"`
	Port    int    `yaml:"port" json:"port"`
	Path    string `yaml:"path" json:"path"`
}

// DefaultMetricsConfig 默认指标配置
func DefaultMetricsConfig() *MetricsConfig {
	return &MetricsConfig{
		Enabled: true,
		Port:    9090,
		Path:    "/metrics",
	}
}

// PrometheusMetrics Prometheus指标收集器
type PrometheusMetrics struct {
	logger   *zap.Logger
	registry *prometheus.Registry
	config   *MetricsConfig
	server   *http.Server

	// HTTP请求指标
	httpRequestsTotal    *prometheus.CounterVec
	httpRequestDuration  *prometheus.HistogramVec
	httpRequestsInFlight *prometheus.GaugeVec

	// 告警处理指标
	alertProcessingTotal    *prometheus.CounterVec
	alertProcessingDuration *prometheus.HistogramVec
	alertProcessingErrors   *prometheus.CounterVec

	// 规则分发指标
	ruleDistributionTotal    *prometheus.CounterVec
	ruleDistributionDuration *prometheus.HistogramVec
	ruleDistributionErrors   *prometheus.CounterVec

	// 系统资源指标
	goroutinesCount     *prometheus.GaugeVec
	memoryUsage         *prometheus.GaugeVec
	cpuUsage            *prometheus.GaugeVec
	databaseConnections *prometheus.GaugeVec

	// 业务指标
	activeAlerts        *prometheus.GaugeVec
	clusterHealth       *prometheus.GaugeVec
	workflowExecutions  *prometheus.CounterVec
	aiAnalysisMetrics   *prometheus.HistogramVec
}

// NewPrometheusMetrics 创建Prometheus指标收集器
func NewPrometheusMetrics(config *MetricsConfig, logger *zap.Logger) *PrometheusMetrics {
	if config == nil {
		config = DefaultMetricsConfig()
	}
	registry := prometheus.NewRegistry()
	factory := promauto.With(registry)

	metrics := &PrometheusMetrics{
		logger:   logger,
		registry: registry,
		config:   config,

		// HTTP请求指标
		httpRequestsTotal: factory.NewCounterVec(
			prometheus.CounterOpts{
				Name: "alertagent_http_requests_total",
				Help: "Total number of HTTP requests",
			},
			[]string{"method", "path", "status_code"},
		),

		httpRequestDuration: factory.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "alertagent_http_request_duration_seconds",
				Help:    "HTTP request duration in seconds",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"method", "path"},
		),

		httpRequestsInFlight: factory.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "alertagent_http_requests_in_flight",
				Help: "Number of HTTP requests currently being processed",
			},
			[]string{"method", "path"},
		),

		// 告警处理指标
		alertProcessingTotal: factory.NewCounterVec(
			prometheus.CounterOpts{
				Name: "alertagent_alert_processing_total",
				Help: "Total number of alerts processed",
			},
			[]string{"cluster_id", "alert_type", "status"},
		),

		alertProcessingDuration: factory.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "alertagent_alert_processing_duration_seconds",
				Help:    "Alert processing duration in seconds",
				Buckets: []float64{0.1, 0.5, 1.0, 2.5, 5.0, 10.0},
			},
			[]string{"cluster_id", "alert_type"},
		),

		alertProcessingErrors: factory.NewCounterVec(
			prometheus.CounterOpts{
				Name: "alertagent_alert_processing_errors_total",
				Help: "Total number of alert processing errors",
			},
			[]string{"cluster_id", "alert_type", "error_type"},
		),

		// 规则分发指标
		ruleDistributionTotal: factory.NewCounterVec(
			prometheus.CounterOpts{
				Name: "alertagent_rule_distribution_total",
				Help: "Total number of rule distributions",
			},
			[]string{"cluster_id", "rule_type", "status"},
		),

		ruleDistributionDuration: factory.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "alertagent_rule_distribution_duration_seconds",
				Help:    "Rule distribution duration in seconds",
				Buckets: []float64{0.5, 1.0, 2.0, 5.0, 10.0, 30.0},
			},
			[]string{"cluster_id", "rule_type"},
		),

		ruleDistributionErrors: factory.NewCounterVec(
			prometheus.CounterOpts{
				Name: "alertagent_rule_distribution_errors_total",
				Help: "Total number of rule distribution errors",
			},
			[]string{"cluster_id", "rule_type", "error_type"},
		),

		// 系统资源指标
		goroutinesCount: factory.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "alertagent_goroutines_count",
				Help: "Number of goroutines currently running",
			},
			[]string{"component"},
		),

		memoryUsage: factory.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "alertagent_memory_usage_bytes",
				Help: "Memory usage in bytes",
			},
			[]string{"type"},
		),

		cpuUsage: factory.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "alertagent_cpu_usage_percent",
				Help: "CPU usage percentage",
			},
			[]string{"core"},
		),

		databaseConnections: factory.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "alertagent_database_connections",
				Help: "Number of database connections",
			},
			[]string{"state"},
		),

		// 业务指标
		activeAlerts: factory.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "alertagent_active_alerts",
				Help: "Number of active alerts",
			},
			[]string{"cluster_id", "severity"},
		),

		clusterHealth: factory.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "alertagent_cluster_health",
				Help: "Cluster health status (1=healthy, 0=unhealthy)",
			},
			[]string{"cluster_id"},
		),

		workflowExecutions: factory.NewCounterVec(
			prometheus.CounterOpts{
				Name: "alertagent_workflow_executions_total",
				Help: "Total number of workflow executions",
			},
			[]string{"workflow_type", "status"},
		),

		aiAnalysisMetrics: factory.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "alertagent_ai_analysis_duration_seconds",
				Help:    "AI analysis duration in seconds",
				Buckets: []float64{0.1, 0.5, 1.0, 2.0, 5.0, 10.0, 30.0},
			},
			[]string{"model_type", "analysis_type"},
		),
	}

	// 注册Go运行时指标
	registry.MustRegister(prometheus.NewGoCollector())
	registry.MustRegister(prometheus.NewProcessCollector(prometheus.ProcessCollectorOpts{}))

	return metrics
}

// Initialize 初始化指标收集器
func (pm *PrometheusMetrics) Initialize() error {
	// 注册所有指标
	pm.registry.MustRegister(
		pm.httpRequestsTotal,
		pm.httpRequestDuration,
		pm.httpRequestsInFlight,
		pm.alertProcessingTotal,
		pm.alertProcessingDuration,
		pm.alertProcessingErrors,
		pm.ruleDistributionTotal,
		pm.ruleDistributionDuration,
		pm.ruleDistributionErrors,
		pm.goroutinesCount,
		pm.memoryUsage,
		pm.cpuUsage,
		pm.databaseConnections,
		pm.activeAlerts,
		pm.clusterHealth,
		pm.workflowExecutions,
		pm.aiAnalysisMetrics,
	)
	pm.logger.Info("Prometheus metrics initialized")
	return nil
}

// StartServer 启动指标服务器
func (pm *PrometheusMetrics) StartServer() error {
	if !pm.config.Enabled {
		return nil
	}

	mux := http.NewServeMux()
	mux.Handle(pm.config.Path, promhttp.HandlerFor(pm.registry, promhttp.HandlerOpts{}))

	pm.server = &http.Server{
		Addr:    fmt.Sprintf(":%d", pm.config.Port),
		Handler: mux,
	}

	go func() {
		pm.logger.Info("Starting metrics server", zap.Int("port", pm.config.Port))
		if err := pm.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			pm.logger.Error("Metrics server error", zap.Error(err))
		}
	}()

	return nil
}

// Shutdown 关闭指标服务器
func (pm *PrometheusMetrics) Shutdown(ctx context.Context) error {
	if pm.server != nil {
		return pm.server.Shutdown(ctx)
	}
	return nil
}

// RecordHTTPRequest 记录HTTP请求指标
func (pm *PrometheusMetrics) RecordHTTPRequest(method, path string, statusCode int, duration time.Duration) {
	status := fmt.Sprintf("%d", statusCode)
	pm.httpRequestsTotal.WithLabelValues(method, path, status).Inc()
	pm.httpRequestDuration.WithLabelValues(method, path).Observe(duration.Seconds())
}

// IncHTTPRequestsInFlight 增加正在处理的HTTP请求数
func (pm *PrometheusMetrics) IncHTTPRequestsInFlight(method, path string) {
	pm.httpRequestsInFlight.WithLabelValues(method, path).Inc()
}

// DecHTTPRequestsInFlight 减少正在处理的HTTP请求数
func (pm *PrometheusMetrics) DecHTTPRequestsInFlight(method, path string) {
	pm.httpRequestsInFlight.WithLabelValues(method, path).Dec()
}

// RecordAlertProcessing 记录告警处理指标
func (pm *PrometheusMetrics) RecordAlertProcessing(clusterID, alertType, status string, duration time.Duration) {
	pm.alertProcessingTotal.WithLabelValues(clusterID, alertType, status).Inc()
	pm.alertProcessingDuration.WithLabelValues(clusterID, alertType).Observe(duration.Seconds())
}

// RecordAlertProcessingError 记录告警处理错误
func (pm *PrometheusMetrics) RecordAlertProcessingError(clusterID, alertType, errorType string) {
	pm.alertProcessingErrors.WithLabelValues(clusterID, alertType, errorType).Inc()
}

// RecordRuleDistribution 记录规则分发指标
func (pm *PrometheusMetrics) RecordRuleDistribution(clusterID, ruleType, status string, duration time.Duration) {
	pm.ruleDistributionTotal.WithLabelValues(clusterID, ruleType, status).Inc()
	pm.ruleDistributionDuration.WithLabelValues(clusterID, ruleType).Observe(duration.Seconds())
}

// RecordRuleDistributionError 记录规则分发错误
func (pm *PrometheusMetrics) RecordRuleDistributionError(clusterID, ruleType, errorType string) {
	pm.ruleDistributionErrors.WithLabelValues(clusterID, ruleType, errorType).Inc()
}

// SetGoroutinesCount 设置协程数量
func (pm *PrometheusMetrics) SetGoroutinesCount(component string, count float64) {
	pm.goroutinesCount.WithLabelValues(component).Set(count)
}

// SetMemoryUsage 设置内存使用量
func (pm *PrometheusMetrics) SetMemoryUsage(memType string, bytes float64) {
	pm.memoryUsage.WithLabelValues(memType).Set(bytes)
}

// SetCPUUsage 设置CPU使用率
func (pm *PrometheusMetrics) SetCPUUsage(core string, percent float64) {
	pm.cpuUsage.WithLabelValues(core).Set(percent)
}

// SetDatabaseConnections 设置数据库连接数
func (pm *PrometheusMetrics) SetDatabaseConnections(state string, count float64) {
	pm.databaseConnections.WithLabelValues(state).Set(count)
}

// SetActiveAlerts 设置活跃告警数
func (pm *PrometheusMetrics) SetActiveAlerts(clusterID, severity string, count float64) {
	pm.activeAlerts.WithLabelValues(clusterID, severity).Set(count)
}

// SetClusterHealth 设置集群健康状态
func (pm *PrometheusMetrics) SetClusterHealth(clusterID string, healthy bool) {
	value := 0.0
	if healthy {
		value = 1.0
	}
	pm.clusterHealth.WithLabelValues(clusterID).Set(value)
}

// RecordWorkflowExecution 记录工作流执行
func (pm *PrometheusMetrics) RecordWorkflowExecution(workflowType, status string) {
	pm.workflowExecutions.WithLabelValues(workflowType, status).Inc()
}

// RecordAIAnalysis 记录AI分析指标
func (pm *PrometheusMetrics) RecordAIAnalysis(modelType, analysisType string, duration time.Duration) {
	pm.aiAnalysisMetrics.WithLabelValues(modelType, analysisType).Observe(duration.Seconds())
}

// GetRegistry 获取Prometheus注册器
func (pm *PrometheusMetrics) GetRegistry() *prometheus.Registry {
	return pm.registry
}

// Handler 返回Prometheus HTTP处理器
func (pm *PrometheusMetrics) Handler() http.Handler {
	return promhttp.HandlerFor(pm.registry, promhttp.HandlerOpts{
		EnableOpenMetrics: true,
	})
}

// StartMetricsServer 启动指标服务器
func (pm *PrometheusMetrics) StartMetricsServer(ctx context.Context, addr string) error {
	mux := http.NewServeMux()
	mux.Handle("/metrics", pm.Handler())

	server := &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			pm.logger.Error("Failed to shutdown metrics server", zap.Error(err))
		}
	}()

	pm.logger.Info("Starting metrics server", zap.String("addr", addr))
	return server.ListenAndServe()
}