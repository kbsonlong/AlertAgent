package tracing

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"
)

// TracingConfig 追踪配置
type TracingConfig struct {
	Enabled     bool   `yaml:"enabled" json:"enabled"`
	ServiceName string `yaml:"service_name" json:"service_name"`
	Version     string `yaml:"version" json:"version"`
	Environment string `yaml:"environment" json:"environment"`
	SampleRate  float64 `yaml:"sample_rate" json:"sample_rate"`
	MaxSpans    int    `yaml:"max_spans" json:"max_spans"`
}

// DefaultTracingConfig 默认追踪配置
func DefaultTracingConfig() *TracingConfig {
	return &TracingConfig{
		Enabled:     false,
		ServiceName: "alertagent",
		Version:     "1.0.0",
		Environment: "development",
		SampleRate:  1.0,
		MaxSpans:    1000,
	}
}

// SpanContext Span上下文
type SpanContext struct {
	TraceID string
	SpanID  string
}

// Span 追踪Span
type Span struct {
	TraceID     string                 `json:"trace_id"`
	SpanID      string                 `json:"span_id"`
	ParentID    string                 `json:"parent_id,omitempty"`
	OperationName string               `json:"operation_name"`
	StartTime   time.Time              `json:"start_time"`
	EndTime     *time.Time             `json:"end_time,omitempty"`
	Duration    time.Duration          `json:"duration"`
	Tags        map[string]interface{} `json:"tags"`
	Logs        []SpanLog              `json:"logs"`
	Status      SpanStatus             `json:"status"`
	mu          sync.RWMutex
}

// SpanLog Span日志
type SpanLog struct {
	Timestamp time.Time              `json:"timestamp"`
	Fields    map[string]interface{} `json:"fields"`
}

// SpanStatus Span状态
type SpanStatus struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// TracingManager 追踪管理器
type TracingManager struct {
	config      *TracingConfig
	logger      *zap.Logger
	activeSpans map[string]*Span
	mu          sync.RWMutex
}

// NewTracingManager 创建追踪管理器
func NewTracingManager(config *TracingConfig, logger *zap.Logger) *TracingManager {
	return &TracingManager{
		config:      config,
		logger:      logger,
		activeSpans: make(map[string]*Span),
	}
}

// Initialize 初始化追踪
func (tm *TracingManager) Initialize(ctx context.Context) error {
	if !tm.config.Enabled {
		tm.logger.Info("Tracing is disabled")
		return nil
	}

	tm.logger.Info("Tracing initialized successfully",
		zap.String("service_name", tm.config.ServiceName),
		zap.String("version", tm.config.Version),
		zap.String("environment", tm.config.Environment),
		zap.Float64("sample_rate", tm.config.SampleRate),
	)

	return nil
}

// generateID 生成随机ID
func (tm *TracingManager) generateID() string {
	bytes := make([]byte, 8)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

// Shutdown 关闭追踪
func (tm *TracingManager) Shutdown(ctx context.Context) error {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	// 结束所有活跃的span
	for _, span := range tm.activeSpans {
		span.Finish()
	}
	tm.activeSpans = make(map[string]*Span)

	tm.logger.Info("Tracing shutdown successfully")
	return nil
}

// shouldSample 判断是否应该采样
func (tm *TracingManager) shouldSample() bool {
	if !tm.config.Enabled {
		return false
	}
	return tm.config.SampleRate >= 1.0 // 简化的采样逻辑
}

// StartSpan 开始一个新的span
func (tm *TracingManager) StartSpan(ctx context.Context, operationName string) (context.Context, *Span) {
	if !tm.shouldSample() {
		return ctx, nil
	}

	span := &Span{
		TraceID:       tm.generateID(),
		SpanID:        tm.generateID(),
		OperationName: operationName,
		StartTime:     time.Now(),
		Tags:          make(map[string]interface{}),
		Logs:          make([]SpanLog, 0),
		Status:        SpanStatus{Code: 0, Message: "OK"},
	}

	// 检查父span
	if parentSpan := tm.SpanFromContext(ctx); parentSpan != nil {
		span.TraceID = parentSpan.TraceID
		span.ParentID = parentSpan.SpanID
	}

	tm.mu.Lock()
	if len(tm.activeSpans) >= tm.config.MaxSpans {
		// 清理一些旧的span
		tm.cleanupOldSpans()
	}
	tm.activeSpans[span.SpanID] = span
	tm.mu.Unlock()

	// 将span添加到context中
	ctx = context.WithValue(ctx, "span", span)
	return ctx, span
}

// SpanFromContext 从上下文获取span
func (tm *TracingManager) SpanFromContext(ctx context.Context) *Span {
	if span, ok := ctx.Value("span").(*Span); ok {
		return span
	}
	return nil
}

// cleanupOldSpans 清理旧的span
func (tm *TracingManager) cleanupOldSpans() {
	cutoff := time.Now().Add(-5 * time.Minute)
	for id, span := range tm.activeSpans {
		if span.StartTime.Before(cutoff) {
			delete(tm.activeSpans, id)
		}
	}
}

// SetTag 设置span标签
func (span *Span) SetTag(key string, value interface{}) {
	if span == nil {
		return
	}
	span.mu.Lock()
	defer span.mu.Unlock()
	span.Tags[key] = value
}

// LogFields 记录日志字段
func (span *Span) LogFields(fields map[string]interface{}) {
	if span == nil {
		return
	}
	span.mu.Lock()
	defer span.mu.Unlock()
	span.Logs = append(span.Logs, SpanLog{
		Timestamp: time.Now(),
		Fields:    fields,
	})
}

// SetStatus 设置span状态
func (span *Span) SetStatus(code int, message string) {
	if span == nil {
		return
	}
	span.mu.Lock()
	defer span.mu.Unlock()
	span.Status = SpanStatus{Code: code, Message: message}
}

// Finish 结束span
func (span *Span) Finish() {
	if span == nil {
		return
	}
	span.mu.Lock()
	defer span.mu.Unlock()
	if span.EndTime == nil {
		now := time.Now()
		span.EndTime = &now
		span.Duration = now.Sub(span.StartTime)
	}
}

// TraceHTTPRequest HTTP请求追踪辅助函数
func (tm *TracingManager) TraceHTTPRequest(ctx context.Context, method, url, userAgent string) (context.Context, *Span) {
	ctx, span := tm.StartSpan(ctx, fmt.Sprintf("HTTP %s", method))
	if span != nil {
		span.SetTag("http.method", method)
		span.SetTag("http.url", url)
		span.SetTag("http.user_agent", userAgent)
		span.SetTag("component", "http")
	}
	return ctx, span
}

// TraceDBOperation 数据库操作追踪辅助函数
func (tm *TracingManager) TraceDBOperation(ctx context.Context, operation, table string) (context.Context, *Span) {
	ctx, span := tm.StartSpan(ctx, fmt.Sprintf("DB %s %s", operation, table))
	if span != nil {
		span.SetTag("db.operation", operation)
		span.SetTag("db.table", table)
		span.SetTag("db.type", "postgresql")
		span.SetTag("component", "database")
	}
	return ctx, span
}

// TraceExternalCall 外部调用追踪辅助函数
func (tm *TracingManager) TraceExternalCall(ctx context.Context, service, operation string) (context.Context, *Span) {
	ctx, span := tm.StartSpan(ctx, fmt.Sprintf("%s %s", service, operation))
	if span != nil {
		span.SetTag("service.name", service)
		span.SetTag("operation", operation)
		span.SetTag("component", "external")
	}
	return ctx, span
}

// TraceBusinessOperation 业务操作追踪辅助函数
func (tm *TracingManager) TraceBusinessOperation(ctx context.Context, operation string, tags map[string]interface{}) (context.Context, *Span) {
	ctx, span := tm.StartSpan(ctx, operation)
	if span != nil {
		span.SetTag("component", "business")
		for key, value := range tags {
			span.SetTag(key, value)
		}
	}
	return ctx, span
}

// WithTracing 追踪装饰器
func (tm *TracingManager) WithTracing(name string, fn func(ctx context.Context) error) func(ctx context.Context) error {
	return func(ctx context.Context) error {
		ctx, span := tm.StartSpan(ctx, name)
		if span != nil {
			defer span.Finish()
		}

		err := fn(ctx)
		if err != nil && span != nil {
			span.SetStatus(1, err.Error())
			span.LogFields(map[string]interface{}{
				"error": err.Error(),
				"level": "error",
			})
		} else if span != nil {
			span.SetStatus(0, "OK")
		}

		return err
	}
}

// WithTracingInterface 带返回值的追踪装饰器
func (tm *TracingManager) WithTracingInterface(name string, fn func(ctx context.Context) (interface{}, error)) func(ctx context.Context) (interface{}, error) {
	return func(ctx context.Context) (interface{}, error) {
		ctx, span := tm.StartSpan(ctx, name)
		if span != nil {
			defer span.Finish()
		}

		result, err := fn(ctx)
		if err != nil && span != nil {
			span.SetStatus(1, err.Error())
			span.LogFields(map[string]interface{}{
				"error": err.Error(),
				"level": "error",
			})
		} else if span != nil {
			span.SetStatus(0, "OK")
		}

		return result, err
	}
}

// GetTraceID 获取当前追踪ID
func (tm *TracingManager) GetTraceID(ctx context.Context) string {
	span := tm.SpanFromContext(ctx)
	if span != nil {
		return span.TraceID
	}
	return ""
}

// GetSpanID 获取当前SpanID
func (tm *TracingManager) GetSpanID(ctx context.Context) string {
	span := tm.SpanFromContext(ctx)
	if span != nil {
		return span.SpanID
	}
	return ""
}

// GetActiveSpans 获取活跃的span列表
func (tm *TracingManager) GetActiveSpans() map[string]*Span {
	tm.mu.RLock()
	defer tm.mu.RUnlock()
	spans := make(map[string]*Span)
	for id, span := range tm.activeSpans {
		spans[id] = span
	}
	return spans
}

// RecordError 记录错误
func (tm *TracingManager) RecordError(span *Span, err error) {
	if span != nil && err != nil {
		span.SetStatus(1, err.Error())
		span.LogFields(map[string]interface{}{
			"error": err.Error(),
			"level": "error",
			"event": "error",
		})
	}
}