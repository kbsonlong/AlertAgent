package observability

import (
	"context"
	"fmt"
	"time"

	"alert_agent/internal/observability/health"
	"alert_agent/internal/observability/logging"
	"alert_agent/internal/observability/metrics"
	"alert_agent/internal/observability/tracing"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// Config 观测性配置
type Config struct {
	Metrics *metrics.MetricsConfig   `yaml:"metrics" json:"metrics"`
	Tracing *tracing.TracingConfig   `yaml:"tracing" json:"tracing"`
	Logging *logging.LoggingConfig   `yaml:"logging" json:"logging"`
	Health  *health.HealthConfig     `yaml:"health" json:"health"`
}

// DefaultConfig 默认观测性配置
func DefaultConfig() *Config {
	return &Config{
		Metrics: metrics.DefaultMetricsConfig(),
		Tracing: tracing.DefaultTracingConfig(),
		Logging: logging.DefaultLoggingConfig(),
		Health:  health.DefaultHealthConfig(),
	}
}

// Manager 观测性管理器
type Manager struct {
	config         *Config
	metricsManager *metrics.PrometheusMetrics
	tracingManager *tracing.TracingManager
	loggingManager *logging.LoggerManager
	healthManager  *health.HealthManager
	logger         *zap.Logger
}

// NewManager 创建观测性管理器
func NewManager(config *Config) *Manager {
	return &Manager{
		config: config,
	}
}

// Initialize 初始化观测性组件
func (m *Manager) Initialize(ctx context.Context, db *gorm.DB) error {
	// 初始化日志管理器
	m.loggingManager = logging.NewLoggerManager(m.config.Logging)
	if err := m.loggingManager.Initialize(); err != nil {
		return fmt.Errorf("failed to initialize logging: %w", err)
	}
	m.logger = m.loggingManager.GetLogger()

	// 初始化指标管理器
	m.metricsManager = metrics.NewPrometheusMetrics(m.config.Metrics, m.logger)
	if err := m.metricsManager.Initialize(); err != nil {
		return fmt.Errorf("failed to initialize metrics: %w", err)
	}

	// 初始化追踪管理器
	m.tracingManager = tracing.NewTracingManager(m.config.Tracing, m.logger)
	if err := m.tracingManager.Initialize(ctx); err != nil {
		return fmt.Errorf("failed to initialize tracing: %w", err)
	}

	// 初始化健康检查管理器
	m.healthManager = health.NewHealthManager(m.logger, m.config.Health.Version)

	// 注册默认健康检查器
	if db != nil {
		dbChecker := health.NewDatabaseHealthChecker(db)
		m.healthManager.RegisterHealthChecker(dbChecker)
		m.healthManager.RegisterReadinessChecker(dbChecker)
	}

	// 注册Redis健康检查器（如果配置了Redis）
	redisChecker := health.NewRedisHealthChecker()
	m.healthManager.RegisterHealthChecker(redisChecker)

	// 注册外部服务健康检查器
	for _, service := range m.config.Health.ExternalServices {
		externalChecker := health.NewExternalServiceHealthChecker(
			service.Name,
			service.URL,
			time.Duration(service.Timeout)*time.Second,
		)
		m.healthManager.RegisterHealthChecker(externalChecker)
	}

	m.logger.Info("Observability components initialized successfully")
	return nil
}

// GetMetricsManager 获取指标管理器
func (m *Manager) GetMetricsManager() *metrics.PrometheusMetrics {
	return m.metricsManager
}

// GetTracingManager 获取追踪管理器
func (m *Manager) GetTracingManager() *tracing.TracingManager {
	return m.tracingManager
}

// GetLoggingManager 获取日志管理器
func (m *Manager) GetLoggingManager() *logging.LoggerManager {
	return m.loggingManager
}

// GetHealthManager 获取健康检查管理器
func (m *Manager) GetHealthManager() *health.HealthManager {
	return m.healthManager
}

// GetLogger 获取日志记录器
func (m *Manager) GetLogger() *zap.Logger {
	return m.logger
}

// StartMetricsServer 启动指标服务器
func (m *Manager) StartMetricsServer() error {
	if m.metricsManager != nil && m.config.Metrics.Enabled {
		return m.metricsManager.StartServer()
	}
	return nil
}

// RecordHTTPRequest 记录HTTP请求指标
func (m *Manager) RecordHTTPRequest(method, path string, statusCode int, duration time.Duration) {
	if m.metricsManager != nil {
		m.metricsManager.RecordHTTPRequest(method, path, statusCode, duration)
	}
}

// RecordAlertProcessed 记录告警处理指标
func (m *Manager) RecordAlertProcessed(clusterID, alertType, status string, duration time.Duration) {
	if m.metricsManager != nil {
		m.metricsManager.RecordAlertProcessing(clusterID, alertType, status, duration)
	}
}

// RecordRuleDistribution 记录规则分发指标
func (m *Manager) RecordRuleDistribution(clusterID, ruleType, status string, duration time.Duration) {
	if m.metricsManager != nil {
		m.metricsManager.RecordRuleDistribution(clusterID, ruleType, status, duration)
	}
}

// UpdateActiveAlerts 更新活跃告警数量
func (m *Manager) UpdateActiveAlerts(clusterID, severity string, count float64) {
	if m.metricsManager != nil {
		m.metricsManager.SetActiveAlerts(clusterID, severity, count)
	}
}

// UpdateClusterHealth 更新集群健康状态
func (m *Manager) UpdateClusterHealth(clusterID string, healthy bool) {
	if m.metricsManager != nil {
		m.metricsManager.SetClusterHealth(clusterID, healthy)
	}
}

// TraceHTTPRequest 追踪HTTP请求
func (m *Manager) TraceHTTPRequest(ctx context.Context, method, url, userAgent string) (context.Context, *tracing.Span) {
	if m.tracingManager != nil {
		return m.tracingManager.TraceHTTPRequest(ctx, method, url, userAgent)
	}
	return ctx, nil
}

// TraceDBOperation 追踪数据库操作
func (m *Manager) TraceDBOperation(ctx context.Context, operation, table string) (context.Context, *tracing.Span) {
	if m.tracingManager != nil {
		return m.tracingManager.TraceDBOperation(ctx, operation, table)
	}
	return ctx, nil
}

// TraceBusinessOperation 追踪业务操作
func (m *Manager) TraceBusinessOperation(ctx context.Context, operation string, tags map[string]interface{}) (context.Context, *tracing.Span) {
	if m.tracingManager != nil {
		return m.tracingManager.TraceBusinessOperation(ctx, operation, tags)
	}
	return ctx, nil
}

// LogWithContext 带上下文的日志记录
func (m *Manager) LogWithContext(ctx context.Context) *logging.ContextLogger {
	if m.loggingManager != nil {
		return m.loggingManager.WithContext(ctx)
	}
	return nil
}

// LogError 记录错误
func (m *Manager) LogError(ctx context.Context, err error, operation string, details map[string]interface{}) {
	if m.loggingManager != nil {
		middleware := logging.NewLoggerMiddleware(m.loggingManager)
		middleware.LogError(ctx, err, operation, details)
	}
}

// LogBusinessEvent 记录业务事件
func (m *Manager) LogBusinessEvent(ctx context.Context, event string, details map[string]interface{}) {
	if m.loggingManager != nil {
		middleware := logging.NewLoggerMiddleware(m.loggingManager)
		middleware.LogBusinessEvent(ctx, event, details)
	}
}

// LogSecurityEvent 记录安全事件
func (m *Manager) LogSecurityEvent(ctx context.Context, event string, severity string, details map[string]interface{}) {
	if m.loggingManager != nil {
		middleware := logging.NewLoggerMiddleware(m.loggingManager)
		middleware.LogSecurityEvent(ctx, event, severity, details)
	}
}

// CheckHealth 执行健康检查
func (m *Manager) CheckHealth(ctx context.Context) *health.HealthReport {
	if m.healthManager != nil {
		return m.healthManager.CheckHealth(ctx)
	}
	return nil
}

// CheckReadiness 执行就绪检查
func (m *Manager) CheckReadiness(ctx context.Context) *health.ReadinessReport {
	if m.healthManager != nil {
		return m.healthManager.CheckReadiness(ctx)
	}
	return nil
}

// Shutdown 关闭观测性组件
func (m *Manager) Shutdown(ctx context.Context) error {
	var errors []error

	// 关闭追踪管理器
	if m.tracingManager != nil {
		if err := m.tracingManager.Shutdown(ctx); err != nil {
			errors = append(errors, fmt.Errorf("failed to shutdown tracing: %w", err))
		}
	}

	// 关闭指标管理器
	if m.metricsManager != nil {
		if err := m.metricsManager.Shutdown(ctx); err != nil {
			errors = append(errors, fmt.Errorf("failed to shutdown metrics: %w", err))
		}
	}

	// 关闭健康检查管理器
	if m.healthManager != nil {
		if err := m.healthManager.Shutdown(ctx); err != nil {
			errors = append(errors, fmt.Errorf("failed to shutdown health: %w", err))
		}
	}

	// 同步日志
	if m.loggingManager != nil {
		if err := m.loggingManager.Close(); err != nil {
			errors = append(errors, fmt.Errorf("failed to close logging: %w", err))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("shutdown errors: %v", errors)
	}

	return nil
}