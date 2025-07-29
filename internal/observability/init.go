package observability

import (
	"context"
	"fmt"

	"alert_agent/internal/observability/health"
	"alert_agent/internal/observability/logging"
	"alert_agent/internal/observability/metrics"
	"alert_agent/internal/observability/tracing"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// InitializeObservability 初始化观测性组件
func InitializeObservability(ctx context.Context, config *Config, db *gorm.DB) (*Manager, error) {
	if config == nil {
		config = DefaultConfig()
	}

	// 创建管理器
	manager := NewManager(config)

	// 初始化组件
	if err := manager.Initialize(ctx, db); err != nil {
		return nil, fmt.Errorf("failed to initialize observability: %w", err)
	}

	return manager, nil
}

// SetupObservabilityWithGin 为Gin应用设置观测性
func SetupObservabilityWithGin(router *gin.Engine, manager *Manager) {
	// 创建中间件
	middleware := NewObservabilityMiddleware(manager)

	// 应用所有中间件
	middleware.ApplyAllMiddleware(router)
}

// QuickSetup 快速设置观测性（用于开发和测试）
func QuickSetup(ctx context.Context, router *gin.Engine, db *gorm.DB) (*Manager, error) {
	// 使用默认配置
	config := DefaultConfig()

	// 初始化观测性
	manager, err := InitializeObservability(ctx, config, db)
	if err != nil {
		return nil, err
	}

	// 设置Gin中间件
	SetupObservabilityWithGin(router, manager)

	// 启动指标服务器
	if err := manager.StartMetricsServer(); err != nil {
		manager.GetLogger().Warn("Failed to start metrics server", zap.Error(err))
	}

	return manager, nil
}

// CreateProductionConfig 创建生产环境配置
func CreateProductionConfig() *Config {
	return &Config{
		Metrics: &metrics.MetricsConfig{
			Enabled: true,
			Port:    9090,
			Path:    "/metrics",
		},
		Tracing: &tracing.TracingConfig{
			Enabled:     true,
			ServiceName: "alertagent",
			SampleRate:  0.1, // 10% 采样率
			MaxSpans:    10000,
		},
		Logging: &logging.LoggingConfig{
			Level:      "info",
			Format:     "json",
			Output:     "both",
			FilePath:   "/var/log/alertagent/app.log",
			MaxSize:    100,
			MaxBackups: 10,
			MaxAge:     30,
			Compress:   true,
		},
		Health: &health.HealthConfig{
			Version: "1.0.0",
			ExternalServices: []health.ExternalServiceConfig{
				{
					Name:    "prometheus",
					URL:     "http://prometheus:9090/api/v1/query?query=up",
					Timeout: 5,
				},
				{
					Name:    "alertmanager",
					URL:     "http://alertmanager:9093/api/v1/status",
					Timeout: 5,
				},
			},
		},
	}
}

// CreateDevelopmentConfig 创建开发环境配置
func CreateDevelopmentConfig() *Config {
	return &Config{
		Metrics: &metrics.MetricsConfig{
			Enabled: true,
			Port:    9090,
			Path:    "/metrics",
		},
		Tracing: &tracing.TracingConfig{
			Enabled:     true,
			ServiceName: "alertagent-dev",
			SampleRate:  1.0, // 100% 采样率用于开发
			MaxSpans:    1000,
		},
		Logging: &logging.LoggingConfig{
			Level:      "debug",
			Format:     "console",
			Output:     "stdout",
			FilePath:   "logs/app.log",
			MaxSize:    50,
			MaxBackups: 3,
			MaxAge:     7,
			Compress:   false,
		},
		Health: &health.HealthConfig{
			Version:          "1.0.0-dev",
			ExternalServices: []health.ExternalServiceConfig{},
		},
	}
}

// CreateTestConfig 创建测试环境配置
func CreateTestConfig() *Config {
	return &Config{
		Metrics: &metrics.MetricsConfig{
			Enabled: false, // 测试时禁用指标
			Port:    9090,
			Path:    "/metrics",
		},
		Tracing: &tracing.TracingConfig{
			Enabled:     false, // 测试时禁用追踪
			ServiceName: "alertagent-test",
			SampleRate:  0.0,
			MaxSpans:    100,
		},
		Logging: &logging.LoggingConfig{
			Level:      "warn", // 测试时只记录警告和错误
			Format:     "console",
			Output:     "stdout",
			FilePath:   "/tmp/test.log",
			MaxSize:    10,
			MaxBackups: 1,
			MaxAge:     1,
			Compress:   false,
		},
		Health: &health.HealthConfig{
			Version:          "1.0.0-test",
			ExternalServices: []health.ExternalServiceConfig{},
		},
	}
}

// ValidateConfig 验证配置
func ValidateConfig(config *Config) error {
	if config == nil {
		return fmt.Errorf("config cannot be nil")
	}

	// 验证指标配置
	if config.Metrics != nil {
		if config.Metrics.Port <= 0 || config.Metrics.Port > 65535 {
			return fmt.Errorf("invalid metrics port: %d", config.Metrics.Port)
		}
		if config.Metrics.Path == "" {
			return fmt.Errorf("metrics path cannot be empty")
		}
	}

	// 验证追踪配置
	if config.Tracing != nil {
		if config.Tracing.SampleRate < 0 || config.Tracing.SampleRate > 1 {
			return fmt.Errorf("invalid sample rate: %f", config.Tracing.SampleRate)
		}
		if config.Tracing.MaxSpans <= 0 {
			return fmt.Errorf("max spans must be positive: %d", config.Tracing.MaxSpans)
		}
	}

	// 验证日志配置
	if config.Logging != nil {
		validLevels := map[string]bool{
			"debug": true, "info": true, "warn": true, "error": true,
		}
		if !validLevels[config.Logging.Level] {
			return fmt.Errorf("invalid log level: %s", config.Logging.Level)
		}

		validFormats := map[string]bool{
			"json": true, "console": true,
		}
		if !validFormats[config.Logging.Format] {
			return fmt.Errorf("invalid log format: %s", config.Logging.Format)
		}

		validOutputs := map[string]bool{
			"stdout": true, "file": true, "both": true,
		}
		if !validOutputs[config.Logging.Output] {
			return fmt.Errorf("invalid log output: %s", config.Logging.Output)
		}
	}

	return nil
}

// GetConfigFromEnvironment 从环境变量获取配置
func GetConfigFromEnvironment() *Config {
	// 这里可以添加从环境变量读取配置的逻辑
	// 目前返回默认配置
	return DefaultConfig()
}