package feature

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"go.uber.org/zap"
)

// FeatureEvent 功能事件类型
type FeatureEvent string

const (
	EventFeatureEnabled   FeatureEvent = "enabled"
	EventFeatureDisabled  FeatureEvent = "disabled"
	EventFeatureToggled   FeatureEvent = "toggled"
	EventFeatureDegraded  FeatureEvent = "degraded"
	EventFeatureRestored  FeatureEvent = "restored"
	EventFeatureError     FeatureEvent = "error"
)

// FeatureMetrics 功能监控指标
type FeatureMetrics struct {
	// 功能状态指标
	FeatureState *prometheus.GaugeVec
	
	// 功能使用指标
	FeatureUsage *prometheus.CounterVec
	
	// 功能切换指标
	FeatureToggles *prometheus.CounterVec
	
	// 功能错误指标
	FeatureErrors *prometheus.CounterVec
	
	// AI成熟度指标
	AIMaturityScore *prometheus.GaugeVec
	
	// 功能延迟指标
	FeatureLatency *prometheus.HistogramVec
	
	// 功能降级指标
	FeatureDegradations *prometheus.CounterVec
}

// AlertRule 告警规则
type AlertRule struct {
	Name        string            `json:"name" yaml:"name"`
	Feature     FeatureName       `json:"feature" yaml:"feature"`
	Metric      string            `json:"metric" yaml:"metric"`
	Condition   string            `json:"condition" yaml:"condition"`   // >, <, ==, !=
	Threshold   float64           `json:"threshold" yaml:"threshold"`
	Duration    time.Duration     `json:"duration" yaml:"duration"`
	Labels      map[string]string `json:"labels" yaml:"labels"`
	Annotations map[string]string `json:"annotations" yaml:"annotations"`
	Enabled     bool              `json:"enabled" yaml:"enabled"`
}

// Alert 告警
type Alert struct {
	Rule        AlertRule         `json:"rule"`
	Value       float64           `json:"value"`
	FiredAt     time.Time         `json:"fired_at"`
	Labels      map[string]string `json:"labels"`
	Annotations map[string]string `json:"annotations"`
}

// FeatureMonitor 功能监控器
type FeatureMonitor struct {
	logger      *zap.Logger
	metrics     *FeatureMetrics
	alertRules  map[string]*AlertRule
	activeAlerts map[string]*Alert
	mutex       sync.RWMutex
	alertHandlers []func(Alert)
}

// NewFeatureMonitor 创建功能监控器
func NewFeatureMonitor(logger *zap.Logger) *FeatureMonitor {
	return NewFeatureMonitorWithRegistry(logger, prometheus.DefaultRegisterer)
}

// NewFeatureMonitorWithRegistry 使用指定注册器创建功能监控器
func NewFeatureMonitorWithRegistry(logger *zap.Logger, registerer prometheus.Registerer) *FeatureMonitor {
	factory := promauto.With(registerer)
	
	metrics := &FeatureMetrics{
		FeatureState: factory.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "alertagent_feature_state",
				Help: "Current state of features (0=disabled, 1=enabled, 2=canary, 3=gradual)",
			},
			[]string{"feature", "phase"},
		),
		
		FeatureUsage: factory.NewCounterVec(
			prometheus.CounterOpts{
				Name: "alertagent_feature_usage_total",
				Help: "Total number of feature usage",
			},
			[]string{"feature", "phase", "result"},
		),
		
		FeatureToggles: factory.NewCounterVec(
			prometheus.CounterOpts{
				Name: "alertagent_feature_toggles_total",
				Help: "Total number of feature toggles",
			},
			[]string{"feature", "from_state", "to_state"},
		),
		
		FeatureErrors: factory.NewCounterVec(
			prometheus.CounterOpts{
				Name: "alertagent_feature_errors_total",
				Help: "Total number of feature errors",
			},
			[]string{"feature", "error_type"},
		),
		
		AIMaturityScore: factory.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "alertagent_ai_maturity_score",
				Help: "AI model maturity score for features",
			},
			[]string{"feature", "metric_type"},
		),
		
		FeatureLatency: factory.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "alertagent_feature_latency_seconds",
				Help:    "Feature execution latency in seconds",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"feature", "operation"},
		),
		
		FeatureDegradations: factory.NewCounterVec(
			prometheus.CounterOpts{
				Name: "alertagent_feature_degradations_total",
				Help: "Total number of feature degradations",
			},
			[]string{"feature", "reason"},
		),
	}
	
	monitor := &FeatureMonitor{
		logger:       logger,
		metrics:      metrics,
		alertRules:   make(map[string]*AlertRule),
		activeAlerts: make(map[string]*Alert),
		alertHandlers: make([]func(Alert), 0),
	}
	
	// 初始化默认告警规则
	monitor.initializeDefaultAlertRules()
	
	return monitor
}

// initializeDefaultAlertRules 初始化默认告警规则
func (fm *FeatureMonitor) initializeDefaultAlertRules() {
	defaultRules := []*AlertRule{
		{
			Name:      "FeatureHighErrorRate",
			Feature:   "", // 适用于所有功能
			Metric:    "alertagent_feature_errors_total",
			Condition: ">",
			Threshold: 10,
			Duration:  5 * time.Minute,
			Labels: map[string]string{
				"severity": "warning",
				"team":     "platform",
			},
			Annotations: map[string]string{
				"summary":     "Feature {{ $labels.feature }} has high error rate",
				"description": "Feature {{ $labels.feature }} error rate is {{ $value }} errors in 5 minutes",
			},
			Enabled: true,
		},
		{
			Name:      "AIMaturityScoreLow",
			Feature:   "", // 适用于所有AI功能
			Metric:    "alertagent_ai_maturity_score",
			Condition: "<",
			Threshold: 0.7,
			Duration:  10 * time.Minute,
			Labels: map[string]string{
				"severity": "critical",
				"team":     "ai",
			},
			Annotations: map[string]string{
				"summary":     "AI maturity score too low for feature {{ $labels.feature }}",
				"description": "Feature {{ $labels.feature }} AI maturity score is {{ $value }}, below threshold 0.7",
			},
			Enabled: true,
		},
		{
			Name:      "FeatureFrequentDegradation",
			Feature:   "",
			Metric:    "alertagent_feature_degradations_total",
			Condition: ">",
			Threshold: 3,
			Duration:  1 * time.Hour,
			Labels: map[string]string{
				"severity": "warning",
				"team":     "platform",
			},
			Annotations: map[string]string{
				"summary":     "Feature {{ $labels.feature }} degraded frequently",
				"description": "Feature {{ $labels.feature }} has been degraded {{ $value }} times in 1 hour",
			},
			Enabled: true,
		},
	}
	
	for _, rule := range defaultRules {
		fm.alertRules[rule.Name] = rule
	}
}

// RecordFeatureUsage 记录功能使用
func (fm *FeatureMonitor) RecordFeatureUsage(feature FeatureName, phase Phase, result string) {
	fm.metrics.FeatureUsage.WithLabelValues(string(feature), string(phase), result).Inc()
	
	fm.logger.Debug("Feature usage recorded",
		zap.String("feature", string(feature)),
		zap.String("phase", string(phase)),
		zap.String("result", result))
}

// RecordFeatureChange 记录功能状态变更
func (fm *FeatureMonitor) RecordFeatureChange(feature FeatureName, fromState, toState FeatureState) {
	fm.metrics.FeatureToggles.WithLabelValues(string(feature), string(fromState), string(toState)).Inc()
	
	// 更新功能状态指标
	stateValue := fm.getStateValue(toState)
	fm.metrics.FeatureState.WithLabelValues(string(feature), "").Set(stateValue)
	
	fm.logger.Info("Feature state changed",
		zap.String("feature", string(feature)),
		zap.String("from_state", string(fromState)),
		zap.String("to_state", string(toState)))
}

// RecordFeatureError 记录功能错误
func (fm *FeatureMonitor) RecordFeatureError(feature FeatureName, errorType string) {
	fm.metrics.FeatureErrors.WithLabelValues(string(feature), errorType).Inc()
	
	fm.logger.Error("Feature error recorded",
		zap.String("feature", string(feature)),
		zap.String("error_type", errorType))
}

// RecordAIMaturityScore 记录AI成熟度分数
func (fm *FeatureMonitor) RecordAIMaturityScore(feature FeatureName, metrics AIMetrics) {
	fm.metrics.AIMaturityScore.WithLabelValues(string(feature), "accuracy").Set(metrics.Accuracy)
	fm.metrics.AIMaturityScore.WithLabelValues(string(feature), "confidence").Set(metrics.Confidence)
	fm.metrics.AIMaturityScore.WithLabelValues(string(feature), "success_rate").Set(metrics.SuccessRate)
	fm.metrics.AIMaturityScore.WithLabelValues(string(feature), "error_rate").Set(metrics.ErrorRate)
	
	fm.logger.Debug("AI maturity score recorded",
		zap.String("feature", string(feature)),
		zap.Float64("accuracy", metrics.Accuracy),
		zap.Float64("confidence", metrics.Confidence))
}

// RecordFeatureLatency 记录功能延迟
func (fm *FeatureMonitor) RecordFeatureLatency(feature FeatureName, operation string, duration time.Duration) {
	fm.metrics.FeatureLatency.WithLabelValues(string(feature), operation).Observe(duration.Seconds())
	
	fm.logger.Debug("Feature latency recorded",
		zap.String("feature", string(feature)),
		zap.String("operation", operation),
		zap.Duration("duration", duration))
}

// RecordFeatureDegradation 记录功能降级
func (fm *FeatureMonitor) RecordFeatureDegradation(feature FeatureName, reason string) {
	fm.metrics.FeatureDegradations.WithLabelValues(string(feature), reason).Inc()
	
	fm.logger.Warn("Feature degradation recorded",
		zap.String("feature", string(feature)),
		zap.String("reason", reason))
}

// getStateValue 获取状态数值
func (fm *FeatureMonitor) getStateValue(state FeatureState) float64 {
	switch state {
	case StateDisabled:
		return 0
	case StateEnabled:
		return 1
	case StateCanaryTest:
		return 2
	case StateGradualRoll:
		return 3
	default:
		return 0
	}
}

// AddAlertRule 添加告警规则
func (fm *FeatureMonitor) AddAlertRule(rule *AlertRule) error {
	fm.mutex.Lock()
	defer fm.mutex.Unlock()
	
	if _, exists := fm.alertRules[rule.Name]; exists {
		return fmt.Errorf("alert rule %s already exists", rule.Name)
	}
	
	fm.alertRules[rule.Name] = rule
	
	fm.logger.Info("Alert rule added",
		zap.String("rule_name", rule.Name),
		zap.String("feature", string(rule.Feature)),
		zap.String("metric", rule.Metric))
	
	return nil
}

// RemoveAlertRule 移除告警规则
func (fm *FeatureMonitor) RemoveAlertRule(ruleName string) error {
	fm.mutex.Lock()
	defer fm.mutex.Unlock()
	
	if _, exists := fm.alertRules[ruleName]; !exists {
		return fmt.Errorf("alert rule %s not found", ruleName)
	}
	
	delete(fm.alertRules, ruleName)
	
	fm.logger.Info("Alert rule removed", zap.String("rule_name", ruleName))
	
	return nil
}

// GetAlertRules 获取所有告警规则
func (fm *FeatureMonitor) GetAlertRules() map[string]*AlertRule {
	fm.mutex.RLock()
	defer fm.mutex.RUnlock()
	
	result := make(map[string]*AlertRule)
	for name, rule := range fm.alertRules {
		ruleCopy := *rule
		result[name] = &ruleCopy
	}
	
	return result
}

// CheckAlerts 检查告警条件
func (fm *FeatureMonitor) CheckAlerts(ctx context.Context) error {
	fm.mutex.Lock()
	defer fm.mutex.Unlock()
	
	for _, rule := range fm.alertRules {
		if !rule.Enabled {
			continue
		}
		
		// 这里应该查询Prometheus获取实际指标值
		// 为了演示，我们使用模拟值
		currentValue := fm.getMetricValue(rule.Metric, rule.Feature)
		
		shouldAlert := fm.evaluateCondition(currentValue, rule.Condition, rule.Threshold)
		
		alertKey := fmt.Sprintf("%s_%s", rule.Name, rule.Feature)
		existingAlert, hasActiveAlert := fm.activeAlerts[alertKey]
		
		if shouldAlert {
			if !hasActiveAlert {
				// 创建新告警
				alert := &Alert{
					Rule:    *rule,
					Value:   currentValue,
					FiredAt: time.Now(),
					Labels:  rule.Labels,
					Annotations: rule.Annotations,
				}
				
				fm.activeAlerts[alertKey] = alert
				fm.triggerAlert(*alert)
				
				fm.logger.Warn("Alert fired",
					zap.String("rule", rule.Name),
					zap.String("feature", string(rule.Feature)),
					zap.Float64("value", currentValue),
					zap.Float64("threshold", rule.Threshold))
			}
		} else {
			if hasActiveAlert {
				// 解除告警
				delete(fm.activeAlerts, alertKey)
				
				fm.logger.Info("Alert resolved",
					zap.String("rule", rule.Name),
					zap.String("feature", string(rule.Feature)),
					zap.Duration("duration", time.Since(existingAlert.FiredAt)))
			}
		}
	}
	
	return nil
}

// getMetricValue 获取指标值（模拟实现）
func (fm *FeatureMonitor) getMetricValue(metric string, feature FeatureName) float64 {
	// 在实际实现中，这里应该查询Prometheus
	// 这里返回模拟值
	switch metric {
	case "alertagent_feature_errors_total":
		return 5.0
	case "alertagent_ai_maturity_score":
		return 0.8
	case "alertagent_feature_degradations_total":
		return 1.0
	default:
		return 0.0
	}
}

// evaluateCondition 评估告警条件
func (fm *FeatureMonitor) evaluateCondition(value float64, condition string, threshold float64) bool {
	switch condition {
	case ">":
		return value > threshold
	case "<":
		return value < threshold
	case "==":
		return value == threshold
	case "!=":
		return value != threshold
	case ">=":
		return value >= threshold
	case "<=":
		return value <= threshold
	default:
		return false
	}
}

// triggerAlert 触发告警
func (fm *FeatureMonitor) triggerAlert(alert Alert) {
	for _, handler := range fm.alertHandlers {
		go func(h func(Alert)) {
			defer func() {
				if r := recover(); r != nil {
					fm.logger.Error("Alert handler panic",
						zap.String("rule", alert.Rule.Name),
						zap.Any("panic", r))
				}
			}()
			h(alert)
		}(handler)
	}
}

// RegisterAlertHandler 注册告警处理器
func (fm *FeatureMonitor) RegisterAlertHandler(handler func(Alert)) {
	fm.alertHandlers = append(fm.alertHandlers, handler)
}

// GetActiveAlerts 获取活跃告警
func (fm *FeatureMonitor) GetActiveAlerts() map[string]*Alert {
	fm.mutex.RLock()
	defer fm.mutex.RUnlock()
	
	result := make(map[string]*Alert)
	for key, alert := range fm.activeAlerts {
		alertCopy := *alert
		result[key] = &alertCopy
	}
	
	return result
}

// StartMonitoring 启动监控
func (fm *FeatureMonitor) StartMonitoring(ctx context.Context) error {
	ticker := time.NewTicker(30 * time.Second) // 每30秒检查一次
	defer ticker.Stop()
	
	fm.logger.Info("Feature monitoring started")
	
	for {
		select {
		case <-ctx.Done():
			fm.logger.Info("Feature monitoring stopped")
			return ctx.Err()
		case <-ticker.C:
			if err := fm.CheckAlerts(ctx); err != nil {
				fm.logger.Error("Failed to check alerts", zap.Error(err))
			}
		}
	}
}

// GetMetrics 获取监控指标
func (fm *FeatureMonitor) GetMetrics() *FeatureMetrics {
	return fm.metrics
}

// GenerateReport 生成监控报告
func (fm *FeatureMonitor) GenerateReport(feature FeatureName, duration time.Duration) (*MonitoringReport, error) {
	// 这里应该查询Prometheus获取历史数据
	// 为了演示，返回模拟报告
	report := &MonitoringReport{
		Feature:     feature,
		Period:      duration,
		GeneratedAt: time.Now(),
		Summary: ReportSummary{
			TotalUsage:      1000,
			ErrorCount:      5,
			ErrorRate:       0.005,
			AvgLatency:      150 * time.Millisecond,
			DegradationCount: 1,
		},
		Recommendations: []string{
			"考虑优化功能性能以降低延迟",
			"监控错误率趋势，必要时进行调优",
		},
	}
	
	return report, nil
}

// MonitoringReport 监控报告
type MonitoringReport struct {
	Feature         FeatureName   `json:"feature"`
	Period          time.Duration `json:"period"`
	GeneratedAt     time.Time     `json:"generated_at"`
	Summary         ReportSummary `json:"summary"`
	Recommendations []string      `json:"recommendations"`
}

// ReportSummary 报告摘要
type ReportSummary struct {
	TotalUsage       int           `json:"total_usage"`
	ErrorCount       int           `json:"error_count"`
	ErrorRate        float64       `json:"error_rate"`
	AvgLatency       time.Duration `json:"avg_latency"`
	DegradationCount int           `json:"degradation_count"`
}