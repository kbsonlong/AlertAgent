package logger

import (
	"os"

	"go.uber.org/zap"
)

// NewLogger 创建新的日志器
func NewLogger() *zap.Logger {
	// 根据环境变量决定日志配置
	env := os.Getenv("APP_ENV")
	if env == "production" {
		return newProductionLogger()
	}
	return newDevelopmentLogger()
}

// newProductionLogger 创建生产环境日志器
func newProductionLogger() *zap.Logger {
	config := zap.NewProductionConfig()
	config.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
	config.OutputPaths = []string{"stdout"}
	config.ErrorOutputPaths = []string{"stderr"}

	logger, err := config.Build()
	if err != nil {
		panic(err)
	}

	return logger
}

// newDevelopmentLogger 创建开发环境日志器
func newDevelopmentLogger() *zap.Logger {
	config := zap.NewDevelopmentConfig()
	config.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
	config.OutputPaths = []string{"stdout"}
	config.ErrorOutputPaths = []string{"stderr"}

	logger, err := config.Build()
	if err != nil {
		panic(err)
	}

	return logger
}

// NewStructuredLogger 创建结构化日志器
func NewStructuredLogger(level string, format string) *zap.Logger {
	var config zap.Config

	switch format {
	case "json":
		config = zap.NewProductionConfig()
	default:
		config = zap.NewDevelopmentConfig()
	}

	// 设置日志级别
	switch level {
	case "debug":
		config.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
	case "info":
		config.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
	case "warn":
		config.Level = zap.NewAtomicLevelAt(zap.WarnLevel)
	case "error":
		config.Level = zap.NewAtomicLevelAt(zap.ErrorLevel)
	default:
		config.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
	}

	logger, err := config.Build()
	if err != nil {
		panic(err)
	}

	return logger
}