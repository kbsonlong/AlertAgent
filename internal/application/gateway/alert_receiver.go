package gateway

import (
	"context"
	"fmt"
	"time"

	"alert_agent/internal/domain/gateway"
	"alert_agent/internal/model"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// AlertReceiverService 告警接收器服务实现
type AlertReceiverService struct {
	repository       gateway.AlertProcessingRepository
	metricsCollector gateway.MetricsCollector
	logger           *zap.Logger
}

// NewAlertReceiverService 创建告警接收器服务
func NewAlertReceiverService(
	repository gateway.AlertProcessingRepository,
	metricsCollector gateway.MetricsCollector,
	logger *zap.Logger,
) *AlertReceiverService {
	return &AlertReceiverService{
		repository:       repository,
		metricsCollector: metricsCollector,
		logger:           logger,
	}
}

// Receive 接收告警
func (ars *AlertReceiverService) Receive(ctx context.Context, alert *model.Alert) (*gateway.AlertProcessingRecord, error) {
	start := time.Now()

	// 验证告警
	if err := ars.ValidateAlert(ctx, alert); err != nil {
		ars.logger.Error("Alert validation failed", zap.Error(err), zap.Uint("alert_id", alert.ID))
		return nil, fmt.Errorf("alert validation failed: %w", err)
	}

	// 创建处理记录
	record := &gateway.AlertProcessingRecord{
		ID:             uuid.New().String(),
		AlertID:        alert.ID,
		OriginalAlert:  alert,
		ProcessingMode: gateway.ModeDirectPassthrough, // 默认直通模式
		Status:         gateway.AlertStatusReceived,
		ReceivedAt:     time.Now(),
		ProcessingSteps: []gateway.ProcessingStep{
			{
				Step:      "alert_received",
				Status:    "completed",
				StartTime: start,
				EndTime:   &[]time.Time{time.Now()}[0],
				Duration:  time.Since(start),
				Details: map[string]interface{}{
					"alert_name":     alert.Name,
					"alert_level":    alert.Level,
					"alert_severity": alert.Severity,
					"alert_source":   alert.Source,
				},
			},
		},
		Metadata: map[string]interface{}{
			"received_at": time.Now().Unix(),
			"source":      alert.Source,
			"level":       alert.Level,
			"severity":    alert.Severity,
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// 保存处理记录
	if err := ars.repository.Create(ctx, record); err != nil {
		ars.logger.Error("Failed to create processing record", zap.Error(err))
		ars.metricsCollector.RecordError(ctx, "create_processing_record", err)
		return nil, fmt.Errorf("failed to create processing record: %w", err)
	}

	ars.logger.Info("Alert received successfully", 
		zap.String("record_id", record.ID),
		zap.Uint("alert_id", alert.ID),
		zap.String("alert_name", alert.Name))

	return record, nil
}

// ValidateAlert 验证告警
func (ars *AlertReceiverService) ValidateAlert(ctx context.Context, alert *model.Alert) error {
	if alert == nil {
		return fmt.Errorf("alert cannot be nil")
	}

	// 使用模型自带的验证方法
	if err := alert.Validate(); err != nil {
		return fmt.Errorf("alert validation failed: %w", err)
	}

	// 额外的业务验证
	if alert.ID == 0 {
		return fmt.Errorf("alert ID is required")
	}

	if alert.RuleID == 0 {
		return fmt.Errorf("rule ID is required")
	}

	// 验证告警级别
	if !isValidAlertLevel(alert.Level) {
		return fmt.Errorf("invalid alert level: %s", alert.Level)
	}

	// 验证告警状态
	if !isValidAlertStatus(alert.Status) {
		return fmt.Errorf("invalid alert status: %s", alert.Status)
	}

	return nil
}

// EnrichAlert 丰富告警信息
func (ars *AlertReceiverService) EnrichAlert(ctx context.Context, alert *model.Alert) (*gateway.AlertContext, error) {
	// 创建基础告警上下文
	alertCtx := &gateway.AlertContext{
		Alert:           alert,
		HistoricalData:  []gateway.HistoricalAlert{},
		RelatedMetrics:  make(map[string]interface{}),
		ProcessingHints: make(map[string]interface{}),
	}

	// 从告警标签中提取环境信息
	if alert.Labels != "" {
		// 这里可以解析JSON格式的标签
		// 简化实现，直接设置一些默认值
		alertCtx.Environment = "production" // 可以从标签中解析
		alertCtx.ClusterID = "default"     // 可以从标签中解析
		alertCtx.Team = "platform"         // 可以从标签中解析
	}

	// 添加处理提示
	alertCtx.ProcessingHints["priority"] = ars.calculatePriority(alert)
	alertCtx.ProcessingHints["urgency"] = ars.calculateUrgency(alert)
	alertCtx.ProcessingHints["category"] = ars.categorizeAlert(alert)

	// 查找历史相似告警（简化实现）
	// 在实际实现中，这里会查询数据库获取历史告警
	historicalAlerts := ars.findHistoricalAlerts(ctx, alert)
	alertCtx.HistoricalData = historicalAlerts

	// 添加相关指标信息
	alertCtx.RelatedMetrics["alert_frequency"] = len(historicalAlerts)
	alertCtx.RelatedMetrics["last_occurrence"] = time.Now().Add(-time.Hour).Unix()

	ars.logger.Debug("Alert context enriched", 
		zap.Uint("alert_id", alert.ID),
		zap.String("environment", alertCtx.Environment),
		zap.String("cluster_id", alertCtx.ClusterID),
		zap.Int("historical_count", len(historicalAlerts)))

	return alertCtx, nil
}

// calculatePriority 计算告警优先级
func (ars *AlertReceiverService) calculatePriority(alert *model.Alert) int {
	switch alert.Level {
	case model.AlertLevelCritical:
		return 1 // 最高优先级
	case model.AlertLevelHigh:
		return 2
	case model.AlertLevelMedium:
		return 3
	case model.AlertLevelLow:
		return 4
	default:
		return 5 // 最低优先级
	}
}

// calculateUrgency 计算告警紧急程度
func (ars *AlertReceiverService) calculateUrgency(alert *model.Alert) string {
	// 基于告警级别和严重程度计算紧急程度
	if alert.Level == model.AlertLevelCritical {
		return "immediate"
	}
	if alert.Level == model.AlertLevelHigh {
		return "high"
	}
	if alert.Level == model.AlertLevelMedium {
		return "medium"
	}
	return "low"
}

// categorizeAlert 对告警进行分类
func (ars *AlertReceiverService) categorizeAlert(alert *model.Alert) string {
	// 基于告警名称和内容进行简单分类
	name := alert.Name
	if name == "" {
		name = alert.Title
	}

	// 简化的分类逻辑
	switch {
	case contains(name, "cpu", "CPU"):
		return "performance"
	case contains(name, "memory", "Memory", "RAM"):
		return "performance"
	case contains(name, "disk", "Disk", "storage"):
		return "storage"
	case contains(name, "network", "Network", "connection"):
		return "network"
	case contains(name, "service", "Service", "application"):
		return "application"
	case contains(name, "database", "Database", "DB"):
		return "database"
	default:
		return "general"
	}
}

// findHistoricalAlerts 查找历史相似告警
func (ars *AlertReceiverService) findHistoricalAlerts(ctx context.Context, alert *model.Alert) []gateway.HistoricalAlert {
	// 简化实现，返回空数组
	// 在实际实现中，这里会查询数据库获取相似的历史告警
	return []gateway.HistoricalAlert{}
}

// contains 检查字符串是否包含任意一个子字符串
func contains(s string, substrings ...string) bool {
	for _, substr := range substrings {
		if len(s) >= len(substr) {
			for i := 0; i <= len(s)-len(substr); i++ {
				if s[i:i+len(substr)] == substr {
					return true
				}
			}
		}
	}
	return false
}

// isValidAlertLevel 验证告警级别
func isValidAlertLevel(level string) bool {
	switch level {
	case model.AlertLevelCritical, model.AlertLevelHigh, model.AlertLevelMedium, model.AlertLevelLow:
		return true
	default:
		return false
	}
}

// isValidAlertStatus 验证告警状态
func isValidAlertStatus(status string) bool {
	switch status {
	case model.AlertStatusNew, model.AlertStatusAcknowledged, model.AlertStatusResolved:
		return true
	default:
		return false
	}
}