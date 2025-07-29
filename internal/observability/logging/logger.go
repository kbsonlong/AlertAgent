package logging

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// LoggingConfig 日志配置
type LoggingConfig struct {
	Level      string `yaml:"level" json:"level"`           // 日志级别: debug, info, warn, error
	Format     string `yaml:"format" json:"format"`         // 日志格式: json, console
	Output     string `yaml:"output" json:"output"`         // 输出方式: stdout, file, both
	FilePath   string `yaml:"file_path" json:"file_path"`   // 日志文件路径
	MaxSize    int    `yaml:"max_size" json:"max_size"`     // 单个日志文件最大大小(MB)
	MaxBackups int    `yaml:"max_backups" json:"max_backups"` // 保留的日志文件数量
	MaxAge     int    `yaml:"max_age" json:"max_age"`       // 日志文件保留天数
	Compress   bool   `yaml:"compress" json:"compress"`     // 是否压缩旧日志文件
}

// DefaultLoggingConfig 默认日志配置
func DefaultLoggingConfig() *LoggingConfig {
	return &LoggingConfig{
		Level:      "info",
		Format:     "json",
		Output:     "stdout",
		FilePath:   "logs/app.log",
		MaxSize:    100,
		MaxBackups: 5,
		MaxAge:     30,
		Compress:   true,
	}
}

// LoggerManager 日志管理器
type LoggerManager struct {
	config *LoggingConfig
	logger *zap.Logger
	file   *os.File
}

// NewLoggerManager 创建日志管理器
func NewLoggerManager(config *LoggingConfig) *LoggerManager {
	return &LoggerManager{
		config: config,
	}
}

// Initialize 初始化日志
func (lm *LoggerManager) Initialize() error {
	// 解析日志级别
	level, err := zapcore.ParseLevel(lm.config.Level)
	if err != nil {
		return fmt.Errorf("invalid log level: %w", err)
	}

	// 创建编码器配置
	encoderConfig := lm.createEncoderConfig()

	// 创建编码器
	var encoder zapcore.Encoder
	switch lm.config.Format {
	case "json":
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	case "console":
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	default:
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	}

	// 创建写入器
	writeSyncer, err := lm.createWriteSyncer()
	if err != nil {
		return fmt.Errorf("failed to create write syncer: %w", err)
	}

	// 创建核心
	core := zapcore.NewCore(encoder, writeSyncer, level)

	// 创建日志选项
	opts := []zap.Option{
		zap.AddCallerSkip(1), // 跳过包装函数
		zap.AddCaller(),
		zap.AddStacktrace(zapcore.ErrorLevel),
	}

	// 创建logger
	lm.logger = zap.New(core, opts...)

	return nil
}

// createEncoderConfig 创建编码器配置
func (lm *LoggerManager) createEncoderConfig() zapcore.EncoderConfig {
	config := zapcore.EncoderConfig{
		TimeKey:        "timestamp",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		FunctionKey:    zapcore.OmitKey,
		MessageKey:     "message",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     lm.timeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   lm.callerEncoder,
	}

	if lm.config.Format == "console" {
		config.EncodeLevel = zapcore.CapitalColorLevelEncoder
		config.EncodeCaller = zapcore.ShortCallerEncoder
	}

	return config
}

// timeEncoder 时间编码器
func (lm *LoggerManager) timeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(t.Format("2006-01-02T15:04:05.000Z07:00"))
}

// callerEncoder 调用者编码器
func (lm *LoggerManager) callerEncoder(caller zapcore.EntryCaller, enc zapcore.PrimitiveArrayEncoder) {
	// 获取相对路径
	if caller.Defined {
		file := caller.File
		if wd, err := os.Getwd(); err == nil {
			if rel, err := filepath.Rel(wd, file); err == nil {
				file = rel
			}
		}
		enc.AppendString(fmt.Sprintf("%s:%d", file, caller.Line))
	}
}

// createWriteSyncer 创建写入器
func (lm *LoggerManager) createWriteSyncer() (zapcore.WriteSyncer, error) {
	var syncers []zapcore.WriteSyncer

	switch lm.config.Output {
	case "stdout":
		syncers = append(syncers, zapcore.AddSync(os.Stdout))
	case "file":
		fileSyncer, err := lm.createFileSyncer()
		if err != nil {
			return nil, err
		}
		syncers = append(syncers, fileSyncer)
	case "both":
		syncers = append(syncers, zapcore.AddSync(os.Stdout))
		fileSyncer, err := lm.createFileSyncer()
		if err != nil {
			return nil, err
		}
		syncers = append(syncers, fileSyncer)
	default:
		syncers = append(syncers, zapcore.AddSync(os.Stdout))
	}

	return zapcore.NewMultiWriteSyncer(syncers...), nil
}

// createFileSyncer 创建文件写入器
func (lm *LoggerManager) createFileSyncer() (zapcore.WriteSyncer, error) {
	// 确保日志目录存在
	logDir := filepath.Dir(lm.config.FilePath)
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create log directory: %w", err)
	}

	// 创建或打开日志文件
	file, err := os.OpenFile(lm.config.FilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %w", err)
	}

	return zapcore.AddSync(file), nil
}

// GetLogger 获取zap logger
func (lm *LoggerManager) GetLogger() *zap.Logger {
	return lm.logger
}

// GetSugar 获取sugar logger
func (lm *LoggerManager) GetSugar() *zap.SugaredLogger {
	return lm.logger.Sugar()
}

// Sync 同步日志
func (lm *LoggerManager) Sync() error {
	if lm.logger != nil {
		return lm.logger.Sync()
	}
	return nil
}

// Close 关闭日志
func (lm *LoggerManager) Close() error {
	return lm.Sync()
}

// WithContext 带上下文的日志记录
func (lm *LoggerManager) WithContext(ctx context.Context) *ContextLogger {
	return &ContextLogger{
		logger: lm.logger,
		ctx:    ctx,
	}
}

// ContextLogger 上下文日志记录器
type ContextLogger struct {
	logger *zap.Logger
	ctx    context.Context
}

// getContextFields 获取上下文字段
func (cl *ContextLogger) getContextFields() []zap.Field {
	var fields []zap.Field

	// 添加追踪信息
	if traceID := cl.getTraceID(); traceID != "" {
		fields = append(fields, zap.String("trace_id", traceID))
	}

	if spanID := cl.getSpanID(); spanID != "" {
		fields = append(fields, zap.String("span_id", spanID))
	}

	// 添加请求ID
	if requestID := cl.getRequestID(); requestID != "" {
		fields = append(fields, zap.String("request_id", requestID))
	}

	// 添加用户ID
	if userID := cl.getUserID(); userID != "" {
		fields = append(fields, zap.String("user_id", userID))
	}

	return fields
}

// getTraceID 获取追踪ID
func (cl *ContextLogger) getTraceID() string {
	if traceID, ok := cl.ctx.Value("trace_id").(string); ok {
		return traceID
	}
	// 尝试从span获取
	if span, ok := cl.ctx.Value("span").(*interface{}); ok && span != nil {
		// 这里需要根据实际的span类型进行类型断言
		// 暂时返回空字符串
	}
	return ""
}

// getSpanID 获取SpanID
func (cl *ContextLogger) getSpanID() string {
	if spanID, ok := cl.ctx.Value("span_id").(string); ok {
		return spanID
	}
	return ""
}

// getRequestID 获取请求ID
func (cl *ContextLogger) getRequestID() string {
	if requestID, ok := cl.ctx.Value("request_id").(string); ok {
		return requestID
	}
	return ""
}

// getUserID 获取用户ID
func (cl *ContextLogger) getUserID() string {
	if userID, ok := cl.ctx.Value("user_id").(string); ok {
		return userID
	}
	return ""
}

// Debug 调试日志
func (cl *ContextLogger) Debug(msg string, fields ...zap.Field) {
	fields = append(cl.getContextFields(), fields...)
	cl.logger.Debug(msg, fields...)
}

// Info 信息日志
func (cl *ContextLogger) Info(msg string, fields ...zap.Field) {
	fields = append(cl.getContextFields(), fields...)
	cl.logger.Info(msg, fields...)
}

// Warn 警告日志
func (cl *ContextLogger) Warn(msg string, fields ...zap.Field) {
	fields = append(cl.getContextFields(), fields...)
	cl.logger.Warn(msg, fields...)
}

// Error 错误日志
func (cl *ContextLogger) Error(msg string, fields ...zap.Field) {
	fields = append(cl.getContextFields(), fields...)
	cl.logger.Error(msg, fields...)
}

// Fatal 致命错误日志
func (cl *ContextLogger) Fatal(msg string, fields ...zap.Field) {
	fields = append(cl.getContextFields(), fields...)
	cl.logger.Fatal(msg, fields...)
}

// WithFields 添加字段
func (cl *ContextLogger) WithFields(fields ...zap.Field) *ContextLogger {
	return &ContextLogger{
		logger: cl.logger.With(fields...),
		ctx:    cl.ctx,
	}
}

// LoggerMiddleware HTTP日志中间件
type LoggerMiddleware struct {
	loggerManager *LoggerManager
}

// NewLoggerMiddleware 创建日志中间件
func NewLoggerMiddleware(loggerManager *LoggerManager) *LoggerMiddleware {
	return &LoggerMiddleware{
		loggerManager: loggerManager,
	}
}

// LogRequest 记录HTTP请求
func (lm *LoggerMiddleware) LogRequest(ctx context.Context, method, path, userAgent, clientIP string, statusCode int, duration time.Duration, bodySize int64) {
	logger := lm.loggerManager.WithContext(ctx)

	fields := []zap.Field{
		zap.String("method", method),
		zap.String("path", path),
		zap.String("user_agent", userAgent),
		zap.String("client_ip", clientIP),
		zap.Int("status_code", statusCode),
		zap.Duration("duration", duration),
		zap.Int64("body_size", bodySize),
	}

	if statusCode >= 500 {
		logger.Error("HTTP request failed", fields...)
	} else if statusCode >= 400 {
		logger.Warn("HTTP request error", fields...)
	} else {
		logger.Info("HTTP request", fields...)
	}
}

// LogError 记录错误
func (lm *LoggerMiddleware) LogError(ctx context.Context, err error, operation string, details map[string]interface{}) {
	logger := lm.loggerManager.WithContext(ctx)

	fields := []zap.Field{
		zap.Error(err),
		zap.String("operation", operation),
	}

	// 添加详细信息
	for key, value := range details {
		fields = append(fields, zap.Any(key, value))
	}

	// 添加调用栈信息
	if pc, file, line, ok := runtime.Caller(1); ok {
		if fn := runtime.FuncForPC(pc); fn != nil {
			fields = append(fields,
				zap.String("function", fn.Name()),
				zap.String("file", file),
				zap.Int("line", line),
			)
		}
	}

	logger.Error("Operation failed", fields...)
}

// LogBusinessEvent 记录业务事件
func (lm *LoggerMiddleware) LogBusinessEvent(ctx context.Context, event string, details map[string]interface{}) {
	logger := lm.loggerManager.WithContext(ctx)

	fields := []zap.Field{
		zap.String("event", event),
		zap.String("event_type", "business"),
	}

	// 添加详细信息
	for key, value := range details {
		fields = append(fields, zap.Any(key, value))
	}

	logger.Info("Business event", fields...)
}

// LogSecurityEvent 记录安全事件
func (lm *LoggerMiddleware) LogSecurityEvent(ctx context.Context, event string, severity string, details map[string]interface{}) {
	logger := lm.loggerManager.WithContext(ctx)

	fields := []zap.Field{
		zap.String("event", event),
		zap.String("event_type", "security"),
		zap.String("severity", severity),
	}

	// 添加详细信息
	for key, value := range details {
		fields = append(fields, zap.Any(key, value))
	}

	switch strings.ToLower(severity) {
	case "critical", "high":
		logger.Error("Security event", fields...)
	case "medium":
		logger.Warn("Security event", fields...)
	default:
		logger.Info("Security event", fields...)
	}
}

// LogPerformanceMetric 记录性能指标
func (lm *LoggerMiddleware) LogPerformanceMetric(ctx context.Context, metric string, value float64, unit string, tags map[string]string) {
	logger := lm.loggerManager.WithContext(ctx)

	fields := []zap.Field{
		zap.String("metric", metric),
		zap.Float64("value", value),
		zap.String("unit", unit),
		zap.String("event_type", "performance"),
	}

	// 添加标签
	for key, value := range tags {
		fields = append(fields, zap.String(fmt.Sprintf("tag_%s", key), value))
	}

	logger.Info("Performance metric", fields...)
}