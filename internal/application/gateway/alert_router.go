package gateway

import (
	"context"
	"fmt"
	"time"

	"alert_agent/internal/domain/gateway"
	"alert_agent/internal/model"
	"alert_agent/internal/pkg/feature"
)

// AlertRouterService 告警路由服务实现
type AlertRouterService struct {
	featureToggle feature.ToggleManager
	metricsCollector gateway.MetricsCollector
}

// NewAlertRouterService 创建新的告警路由服务
func NewAlertRouterService(
	featureToggle feature.ToggleManager,
	metricsCollector gateway.MetricsCollector,
) gateway.AlertRouter {
	return &AlertRouterService{
		featureToggle: featureToggle,
		metricsCollector: metricsCollector,
	}
}

// Route 路由告警到适当的目标
func (ars *AlertRouterService) Route(ctx context.Context, alertCtx *gateway.AlertContext) (*gateway.RoutingDecision, error) {
	start := time.Now()
	defer func() {
		ars.metricsCollector.RecordProcessingLatency(ctx, gateway.ModeDirectPassthrough, time.Since(start).Milliseconds())
	}()

	// 检查直接路由功能是否启用
	if ars.featureToggle.IsEnabled(ctx, feature.FeatureDirectRouting) {
		return ars.performDirectRouting(ctx, alertCtx)
	}

	// 检查智能路由功能是否启用
	if ars.featureToggle.IsEnabled(ctx, feature.FeatureSmartRouting) {
		return ars.performSmartRouting(ctx, alertCtx)
	}

	// 默认路由策略
	return ars.performDefaultRouting(ctx, alertCtx)
}

// GetAvailableChannels 获取可用渠道
func (ars *AlertRouterService) GetAvailableChannels(ctx context.Context, alertCtx *gateway.AlertContext) ([]string, error) {
	return []string{"email", "sms", "slack", "webhook"}, nil
}

// ValidateRouting 验证路由决策
func (ars *AlertRouterService) ValidateRouting(ctx context.Context, decision *gateway.RoutingDecision) error {
	if len(decision.ChannelIDs) == 0 {
		return fmt.Errorf("no targets specified in routing decision")
	}
	return nil
}

// performDirectRouting 执行直接路由
func (ars *AlertRouterService) performDirectRouting(ctx context.Context, alertCtx *gateway.AlertContext) (*gateway.RoutingDecision, error) {
	alert := alertCtx.Alert
	// 基于告警级别的简单路由规则
	targets := ars.determineTargetsByLevel(alert.Level)
	
	decision := &gateway.RoutingDecision{
		ChannelIDs:   targets,
		Priority:     ars.calculatePriority(alert),
		Reason:       "Direct routing based on alert level",
		Confidence:   1.0,
		DecisionTime: time.Now(),
		Metadata: map[string]interface{}{
			"routing_type": "direct",
			"alert_level": alert.Level,
			"target_count": len(targets),
		},
	}

	ars.metricsCollector.RecordAlertRouted(ctx, decision)
	return decision, nil
}

// performSmartRouting 执行智能路由
func (ars *AlertRouterService) performSmartRouting(ctx context.Context, alertCtx *gateway.AlertContext) (*gateway.RoutingDecision, error) {
	alert := alertCtx.Alert
	// 智能路由逻辑（简化实现）
	// 在实际实现中，这里会使用AI/ML模型进行路由决策
	
	// 分析告警特征
	features := ars.extractAlertFeatures(alert)
	
	// 基于特征进行智能路由
	targets := ars.performIntelligentTargetSelection(features, alert)
	
	decision := &gateway.RoutingDecision{
		ChannelIDs:   targets,
		Priority:     ars.calculateSmartPriority(alert, features),
		Reason:       "Smart routing based on AI analysis",
		Confidence:   0.85, // 模拟置信度
		DecisionTime: time.Now(),
		Metadata: map[string]interface{}{
			"routing_type": "smart",
			"features": features,
		},
	}

	ars.metricsCollector.RecordAlertRouted(ctx, decision)
	return decision, nil
}

// performDefaultRouting 执行默认路由
func (ars *AlertRouterService) performDefaultRouting(ctx context.Context, alertCtx *gateway.AlertContext) (*gateway.RoutingDecision, error) {
	// 默认路由到通用目标
	targets := []string{"default_channel"}
	
	decision := &gateway.RoutingDecision{
		ChannelIDs:   targets,
		Priority:     1, // 默认优先级
		Reason:       "Default routing - no specific routing rules matched",
		Confidence:   1.0,
		DecisionTime: time.Now(),
		Metadata: map[string]interface{}{
			"routing_type": "default",
		},
	}

	ars.metricsCollector.RecordAlertRouted(ctx, decision)
	return decision, nil
}

// determineTargetsByLevel 根据告警级别确定目标
func (ars *AlertRouterService) determineTargetsByLevel(level string) []string {
	switch level {
	case "critical":
		return []string{"critical_channel", "sms_channel", "phone_channel"}
	case "warning":
		return []string{"warning_channel", "email_channel"}
	case "info":
		return []string{"info_channel"}
	default:
		return []string{"default_channel"}
	}
}

// calculatePriority 计算路由优先级
func (ars *AlertRouterService) calculatePriority(alert *model.Alert) int {
	switch alert.Level {
	case "critical":
		return 5
	case "warning":
		return 3
	case "info":
		return 1
	default:
		return 1
	}
}

// extractAlertFeatures 提取告警特征
func (ars *AlertRouterService) extractAlertFeatures(alert *model.Alert) map[string]interface{} {
	return map[string]interface{}{
		"level": alert.Level,
		"source": alert.Source,
		"category": categorizeAlertByName(alert),
		"has_labels": len(alert.Labels) > 0,
		"content_length": len(alert.Content),
		"time_of_day": time.Now().Hour(),
		"day_of_week": int(time.Now().Weekday()),
	}
}

// performIntelligentTargetSelection 执行智能目标选择
func (ars *AlertRouterService) performIntelligentTargetSelection(features map[string]interface{}, alert *model.Alert) []string {
	// 简化的智能选择逻辑
	// 在实际实现中，这里会使用机器学习模型
	
	category, _ := features["category"].(string)
	level := alert.Level
	
	var targets []string
	
	// 基于类别和级别的智能路由
	switch category {
	case "performance":
		if level == "critical" {
			targets = []string{"performance_team", "oncall_engineer", "sms_channel"}
		} else {
			targets = []string{"performance_team", "email_channel"}
		}
	case "network":
		if level == "critical" {
			targets = []string{"network_team", "infrastructure_team", "phone_channel"}
		} else {
			targets = []string{"network_team", "slack_channel"}
		}
	case "database":
		if level == "critical" {
			targets = []string{"dba_team", "backend_team", "phone_channel"}
		} else {
			targets = []string{"dba_team", "email_channel"}
		}
	default:
		targets = ars.determineTargetsByLevel(level)
	}
	
	return targets
}

// calculateSmartPriority 计算智能路由优先级
func (ars *AlertRouterService) calculateSmartPriority(alert *model.Alert, features map[string]interface{}) int {
	basePriority := ars.calculatePriority(alert)
	
	// 基于特征调整优先级
	category, _ := features["category"].(string)
	timeOfDay, _ := features["time_of_day"].(int)
	
	// 业务时间内的关键告警优先级更高
	if timeOfDay >= 9 && timeOfDay <= 18 {
		basePriority++
	}
	
	// 某些类别的告警优先级更高
	if category == "database" || category == "network" {
		basePriority++
	}
	
	// 确保优先级在合理范围内
	if basePriority > 5 {
		basePriority = 5
	}
	if basePriority < 1 {
		basePriority = 1
	}
	
	return basePriority
}