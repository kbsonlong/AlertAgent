package gateway

import (
	"context"
	"fmt"
	"strings"
	"time"

	"alert_agent/internal/domain/gateway"
	"alert_agent/internal/model"
	"alert_agent/internal/pkg/feature"
)

// AlertSuppressorService 告警抑制服务实现
type AlertSuppressorService struct {
	featureToggle feature.ToggleManager
	metricsCollector gateway.MetricsCollector
	suppressionRules map[string]*gateway.SuppressionRule // 内存中的抑制规则
}

// NewAlertSuppressorService 创建新的告警抑制服务
func NewAlertSuppressorService(
	featureToggle feature.ToggleManager,
	metricsCollector gateway.MetricsCollector,
) gateway.AlertSuppressor {
	return &AlertSuppressorService{
		featureToggle: featureToggle,
		metricsCollector: metricsCollector,
		suppressionRules: make(map[string]*gateway.SuppressionRule),
	}
}

// ShouldSuppress 判断是否应该抑制告警
func (ass *AlertSuppressorService) ShouldSuppress(ctx context.Context, alertCtx *gateway.AlertContext) (bool, string, error) {
	alert := alertCtx.Alert
	
	// 检查自动抑制功能是否启用
	if !ass.featureToggle.IsEnabled(ctx, feature.FeatureAutoSuppression) {
		return false, "auto suppression disabled", nil
	}

	// 遍历所有抑制规则
	for ruleID, rule := range ass.suppressionRules {
		if !rule.Enabled {
			continue
		}
		
		// 检查规则是否匹配
		if ass.matchesRule(alert, rule) {
			// 检查抑制时长是否仍然有效
			if ass.isSuppressionActive(rule) {
				reason := fmt.Sprintf("Suppressed by rule: %s (%s)", rule.Name, ruleID)
				ass.metricsCollector.RecordError(ctx, "alert_suppressed", fmt.Errorf("alert suppressed: %s", reason))
				return true, reason, nil
			}
		}
	}

	// 检查基于历史数据的智能抑制
	if ass.featureToggle.IsEnabled(ctx, feature.FeatureSmartRouting) {
		if shouldSuppress, reason := ass.checkIntelligentSuppression(ctx, alertCtx); shouldSuppress {
			return true, reason, nil
		}
	}

	return false, "no suppression rules matched", nil
}

// AddSuppressionRule 添加抑制规则
func (ass *AlertSuppressorService) AddSuppressionRule(ctx context.Context, rule gateway.SuppressionRule) error {
	if rule.ID == "" {
		return fmt.Errorf("rule ID cannot be empty")
	}
	
	if rule.Name == "" {
		return fmt.Errorf("rule name cannot be empty")
	}
	
	// 设置创建时间
	rule.CreatedAt = time.Now().Unix()
	rule.UpdatedAt = rule.CreatedAt
	
	// 存储规则
	ass.suppressionRules[rule.ID] = &rule
	
	return nil
}

// RemoveSuppressionRule 移除抑制规则
func (ass *AlertSuppressorService) RemoveSuppressionRule(ctx context.Context, ruleID string) error {
	if ruleID == "" {
		return fmt.Errorf("rule ID cannot be empty")
	}
	
	if _, exists := ass.suppressionRules[ruleID]; !exists {
		return fmt.Errorf("rule with ID %s not found", ruleID)
	}
	
	delete(ass.suppressionRules, ruleID)
	return nil
}

// matchesRule 检查告警是否匹配抑制规则
func (ass *AlertSuppressorService) matchesRule(alert *model.Alert, rule *gateway.SuppressionRule) bool {
	// 检查告警级别
	if level, exists := rule.Conditions["level"]; exists {
		if levelStr, ok := level.(string); ok && levelStr != alert.Level {
			return false
		}
	}
	
	// 检查告警来源
	if source, exists := rule.Conditions["source"]; exists {
		if sourceStr, ok := source.(string); ok && sourceStr != alert.Source {
			return false
		}
	}
	
	// 检查告警名称模式
	if namePattern, exists := rule.Conditions["name_pattern"]; exists {
		if pattern, ok := namePattern.(string); ok {
			if !strings.Contains(alert.Name, pattern) {
				return false
			}
		}
	}
	
	// 检查告警标签
	if labels, exists := rule.Conditions["labels"]; exists {
		if labelMap, ok := labels.(map[string]interface{}); ok {
			for key, expectedValue := range labelMap {
				if expectedStr, ok := expectedValue.(string); ok {
					// 简化的标签匹配，实际应该解析JSON格式的Labels字段
					if !strings.Contains(alert.Labels, fmt.Sprintf(`"%s":"%s"`, key, expectedStr)) {
						return false
					}
				}
			}
		}
	}
	
	return true
}

// isSuppressionActive 检查抑制是否仍然有效
func (ass *AlertSuppressorService) isSuppressionActive(rule *gateway.SuppressionRule) bool {
	if rule.Duration <= 0 {
		return true // 永久抑制
	}
	
	// 检查是否在抑制时间窗口内
	elapsedTime := time.Now().Unix() - rule.UpdatedAt
	return elapsedTime < rule.Duration
}

// checkIntelligentSuppression 检查基于智能分析的抑制
func (ass *AlertSuppressorService) checkIntelligentSuppression(ctx context.Context, alertCtx *gateway.AlertContext) (bool, string) {
	alert := alertCtx.Alert
	
	// 检查是否为重复告警
	if ass.isDuplicateAlert(alertCtx) {
		return true, "Intelligent suppression: duplicate alert detected"
	}
	
	// 检查是否为噪音告警
	if ass.isNoiseAlert(alert) {
		return true, "Intelligent suppression: noise alert detected"
	}
	
	// 检查是否在维护窗口
	if ass.isInMaintenanceWindow(alert) {
		return true, "Intelligent suppression: maintenance window active"
	}
	
	return false, ""
}

// isDuplicateAlert 检查是否为重复告警
func (ass *AlertSuppressorService) isDuplicateAlert(alertCtx *gateway.AlertContext) bool {
	// 简化的重复检测逻辑
	// 在实际实现中，这里会检查历史数据
	for _, historical := range alertCtx.HistoricalData {
		if historical.Similarity > 0.95 { // 95%相似度阈值
			// 检查时间间隔
			timeDiff := time.Since(historical.Timestamp)
			if timeDiff < 5*time.Minute { // 5分钟内的高相似度告警视为重复
				return true
			}
		}
	}
	return false
}

// isNoiseAlert 检查是否为噪音告警
func (ass *AlertSuppressorService) isNoiseAlert(alert *model.Alert) bool {
	// 简化的噪音检测逻辑
	// 检查告警级别和内容
	if alert.Level == "info" && strings.Contains(strings.ToLower(alert.Content), "test") {
		return true
	}
	
	// 检查是否为已知的噪音模式
	noisePatterns := []string{"heartbeat", "ping", "health_check"}
	for _, pattern := range noisePatterns {
		if strings.Contains(strings.ToLower(alert.Name), pattern) {
			return true
		}
	}
	
	return false
}

// isInMaintenanceWindow 检查是否在维护窗口
func (ass *AlertSuppressorService) isInMaintenanceWindow(alert *model.Alert) bool {
	// 简化的维护窗口检测
	// 在实际实现中，这里会查询维护计划数据库
	
	// 检查标签中是否有维护标记
	if strings.Contains(alert.Labels, `"maintenance":"true"`) || strings.Contains(alert.Labels, `"maintenance":"1"`) {
		return true
	}
	
	// 检查是否在预定义的维护时间窗口（例如：每天凌晨2-4点）
	now := time.Now()
	hour := now.Hour()
	if hour >= 2 && hour < 4 {
		// 在维护窗口内，检查是否为基础设施相关告警
	infraKeywords := []string{"disk", "memory", "cpu", "network"}
		for _, keyword := range infraKeywords {
			if strings.Contains(strings.ToLower(alert.Name), keyword) {
				return true
			}
		}
	}
	
	return false
}