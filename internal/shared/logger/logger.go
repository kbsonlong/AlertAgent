package logger

import (
	"context"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	L *zap.Logger
	S *zap.SugaredLogger
)

// Config 日志配置
type Config struct {
	Level            string   `yaml:"level"`
	Format           string   `yaml:"format"` // json, console
	OutputPaths      []string `yaml:"output_paths"`
	ErrorOutputPaths []string `yaml:"error_output_paths"`
	EnableCaller     bool     `yaml:"enable_caller"`
	EnableStacktrace bool     `yaml:"enable_stacktrace"`
}

// Init 初始化日志
func Init(config Config) error {
	level, err := zapcore.ParseLevel(config.Level)
	if err != nil {
		return err
	}

	var zapConfig zap.Config

	if config.Format == "json" {
		zapConfig = zap.NewProductionConfig()
	} else {
		zapConfig = zap.NewDevelopmentConfig()
		zapConfig.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}

	zapConfig.Level = zap.NewAtomicLevelAt(level)
	zapConfig.DisableCaller = !config.EnableCaller
	zapConfig.DisableStacktrace = !config.EnableStacktrace

	if len(config.OutputPaths) > 0 {
		zapConfig.OutputPaths = config.OutputPaths
	}

	if len(config.ErrorOutputPaths) > 0 {
		zapConfig.ErrorOutputPaths = config.ErrorOutputPaths
	}

	logger, err := zapConfig.Build()
	if err != nil {
		return err
	}

	L = logger
	S = logger.Sugar()

	return nil
}

// InitDefault 使用默认配置初始化日志
func InitDefault() error {
	config := Config{
		Level:            "info",
		Format:           "console",
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stderr"},
		EnableCaller:     true,
		EnableStacktrace: false,
	}
	return Init(config)
}

// WithContext 添加上下文信息
func WithContext(ctx context.Context) *zap.Logger {
	if L == nil {
		// 如果日志未初始化，使用默认配置
		InitDefault()
	}

	logger := L

	// 添加请求ID
	if requestID := ctx.Value("request_id"); requestID != nil {
		logger = logger.With(zap.String("request_id", requestID.(string)))
	}

	// 添加用户ID
	if userID := ctx.Value("user_id"); userID != nil {
		logger = logger.With(zap.String("user_id", userID.(string)))
	}

	// 添加追踪ID
	if traceID := ctx.Value("trace_id"); traceID != nil {
		logger = logger.With(zap.String("trace_id", traceID.(string)))
	}

	return logger
}

// WithFields 添加字段
func WithFields(fields ...zap.Field) *zap.Logger {
	if L == nil {
		InitDefault()
	}
	return L.With(fields...)
}

// WithComponent 添加组件名称
func WithComponent(component string) *zap.Logger {
	if L == nil {
		InitDefault()
	}
	return L.Named(component)
}

// Sync 同步日志
func Sync() {
	if L != nil {
		L.Sync()
	}
}

// GetLogger 获取日志实例
func GetLogger() *zap.Logger {
	if L == nil {
		InitDefault()
	}
	return L
}

// GetSugaredLogger 获取Sugar日志实例
func GetSugaredLogger() *zap.SugaredLogger {
	if S == nil {
		InitDefault()
	}
	return S
}

// Info 记录信息日志
func Info(msg string, fields ...zap.Field) {
	if L == nil {
		InitDefault()
	}
	L.Info(msg, fields...)
}

// Error 记录错误日志
func Error(msg string, fields ...zap.Field) {
	if L == nil {
		InitDefault()
	}
	L.Error(msg, fields...)
}

// Warn 记录警告日志
func Warn(msg string, fields ...zap.Field) {
	if L == nil {
		InitDefault()
	}
	L.Warn(msg, fields...)
}

// Debug 记录调试日志
func Debug(msg string, fields ...zap.Field) {
	if L == nil {
		InitDefault()
	}
	L.Debug(msg, fields...)
}

// Fatal 记录致命错误日志
func Fatal(msg string, fields ...zap.Field) {
	if L == nil {
		InitDefault()
	}
	L.Fatal(msg, fields...)
}