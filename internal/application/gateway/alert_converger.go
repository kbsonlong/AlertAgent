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

// AlertConvergerService 告警收敛服务实现
type AlertConvergerService struct {
	featureToggle feature.ToggleManager
	metricsCollector gateway.MetricsCollector
	processingRepo gateway.AlertProcessingRepository
	convergenceWindows map[string]*ConvergenceWindow // 收敛窗口
}

// ConvergenceWindow 收敛窗口
type ConvergenceWindow struct {
	Alerts    []*model.Alert
	StartTime time.Time
	EndTime   time.Time
	GroupKey  string
	Count     int
}

// NewAlertConvergerService 创建新的告警收敛服务
func NewAlertConvergerService(
	featureToggle feature.ToggleManager,
	metricsCollector gateway.MetricsCollector,
	processingRepo gateway.AlertProcessingRepository,
) gateway.AlertConverger {
	return &AlertConvergerService{
		featureToggle: featureToggle,
		metricsCollector: metricsCollector,
		processingRepo: processingRepo,
		convergenceWindows: make(map[string]*ConvergenceWindow),
	}
}

// ShouldConverge 判断是否应该收敛告警
func (acs *AlertConvergerService) ShouldConverge(ctx context.Context, alertCtx *gateway.AlertContext) (bool, *gateway.ConvergenceResult, error) {
	alert := alertCtx.Alert
	
	// 检查收敛功能是否启用
	if !acs.featureToggle.IsEnabled(ctx, feature.FeatureBasicConvergence) {
		return false, nil, nil
	}

	// 生成收敛分组键
	groupKey := acs.generateGroupKey(alert)
	
	// 检查是否存在活跃的收敛窗口
	window, exists := acs.convergenceWindows[groupKey]
	if !exists {
		// 创建新的收敛窗口
		window = &ConvergenceWindow{
			Alerts:    []*model.Alert{alert},
			StartTime: time.Now(),
			EndTime:   time.Now().Add(5 * time.Minute), // 默认5分钟收敛窗口
			GroupKey:  groupKey,
			Count:     1,
		}
		acs.convergenceWindows[groupKey] = window
		return false, nil, nil // 第一个告警不收敛
	}
	
	// 检查收敛窗口是否仍然有效
	if time.Now().After(window.EndTime) {
		// 窗口已过期，清理并创建新窗口
		delete(acs.convergenceWindows, groupKey)
		window = &ConvergenceWindow{
			Alerts:    []*model.Alert{alert},
			StartTime: time.Now(),
			EndTime:   time.Now().Add(5 * time.Minute),
			GroupKey:  groupKey,
			Count:     1,
		}
		acs.convergenceWindows[groupKey] = window
		return false, nil, nil
	}
	
	// 添加到现有窗口
	window.Alerts = append(window.Alerts, alert)
	window.Count++
	
	// 创建收敛结果
	result := &gateway.ConvergenceResult{
		Converged:         true,
		GroupID:           groupKey,
		Representative:    alert,
		SimilarAlerts:     window.Alerts[:len(window.Alerts)-1], // 排除当前告警
		ConvergenceRule:   fmt.Sprintf("Converged %d similar alerts in time window", window.Count),
		Metadata: map[string]interface{}{
			"window_start": window.StartTime.Unix(),
			"window_end":   window.EndTime.Unix(),
			"group_key":    groupKey,
		},
	}
	

	
	return true, result, nil
}

// Converge 执行告警收敛
func (acs *AlertConvergerService) Converge(ctx context.Context, alerts []*model.Alert) (*gateway.ConvergenceResult, error) {
	if len(alerts) == 0 {
		return nil, fmt.Errorf("no alerts to converge")
	}
	
	// 按相似性分组
	groups := acs.groupAlertsBySimilarityForAlerts(alerts)
	
	// 选择最大的组进行收敛
	var largestGroup []*model.Alert
	var largestGroupKey string
	for groupKey, group := range groups {
		if len(group) > len(largestGroup) {
			largestGroup = group
			largestGroupKey = groupKey
		}
	}
	
	if len(largestGroup) < 2 {
		return nil, fmt.Errorf("insufficient alerts for convergence")
	}
	
	// 创建收敛结果
	result := &gateway.ConvergenceResult{
		Converged:         true,
		GroupID:           largestGroupKey,
		Representative:    largestGroup[0],
		SimilarAlerts:     largestGroup[1:],
		ConvergenceRule:   fmt.Sprintf("Converged %d similar alerts", len(largestGroup)),
		Metadata: map[string]interface{}{
			"convergence_time": time.Now().Unix(),
			"group_key":        largestGroupKey,
			"algorithm":        "similarity_based",
		},
	}
	

	
	return result, nil
}

// generateGroupKey 生成收敛分组键
func (acs *AlertConvergerService) generateGroupKey(alert *model.Alert) string {
	// 基于告警名称、级别和来源生成分组键
	key := fmt.Sprintf("%s:%s:%s", alert.Name, alert.Level, alert.Source)
	
	// 添加关键标签信息
	if strings.Contains(alert.Labels, "service") {
		// 简化的标签提取，实际应该解析JSON
		key += ":service"
	}
	
	if strings.Contains(alert.Labels, "instance") {
		key += ":instance"
	}
	
	return key
}

// groupAlertsBySimilarity 按相似性对告警分组
func (acs *AlertConvergerService) groupAlertsBySimilarity(alerts []*gateway.AlertContext) map[string][]*gateway.AlertContext {
	groups := make(map[string][]*gateway.AlertContext)
	
	for _, alertCtx := range alerts {
		groupKey := acs.generateGroupKey(alertCtx.Alert)
		groups[groupKey] = append(groups[groupKey], alertCtx)
	}
	
	return groups
}

// groupAlertsBySimilarityForAlerts 按相似性对告警分组（直接处理Alert）
func (acs *AlertConvergerService) groupAlertsBySimilarityForAlerts(alerts []*model.Alert) map[string][]*model.Alert {
	groups := make(map[string][]*model.Alert)
	
	for _, alert := range alerts {
		groupKey := acs.generateGroupKey(alert)
		groups[groupKey] = append(groups[groupKey], alert)
	}
	
	return groups
}

// convertToHistoricalAlerts 将告警转换为历史告警格式
func (acs *AlertConvergerService) convertToHistoricalAlerts(alerts []*model.Alert) []gateway.HistoricalAlert {
	historical := make([]gateway.HistoricalAlert, len(alerts))
	
	for i, alert := range alerts {
		historical[i] = gateway.HistoricalAlert{
			Alert:      alert,
			Similarity: 0.95, // 默认高相似度
			Timestamp:  alert.CreatedAt,
		}
	}
	
	return historical
}

// extractAlertsFromContexts 从告警上下文中提取告警
func (acs *AlertConvergerService) extractAlertsFromContexts(alertCtxs []*gateway.AlertContext) []*model.Alert {
	alerts := make([]*model.Alert, len(alertCtxs))
	for i, ctx := range alertCtxs {
		alerts[i] = ctx.Alert
	}
	return alerts
}

// FindSimilarAlerts 查找相似告警
func (acs *AlertConvergerService) FindSimilarAlerts(ctx context.Context, alert *model.Alert) ([]*model.Alert, error) {
	// 生成分组键
	groupKey := acs.generateGroupKey(alert)

	// 检查是否存在收敛窗口
	window, exists := acs.convergenceWindows[groupKey]

	if !exists {
		return []*model.Alert{}, nil
	}

	// 检查窗口是否过期
	if time.Now().After(window.EndTime) {
		// 清理过期窗口
		delete(acs.convergenceWindows, groupKey)
		return []*model.Alert{}, nil
	}

	return window.Alerts, nil
}

// CalculateSimilarity 计算相似度
func (acs *AlertConvergerService) CalculateSimilarity(ctx context.Context, alert1, alert2 *model.Alert) (float64, error) {
	// 简单的相似度计算逻辑
	similarity := 0.0

	// 比较告警名称
	if alert1.Name == alert2.Name {
		similarity += 0.4
	}

	// 比较告警级别
	if alert1.Level == alert2.Level {
		similarity += 0.3
	}

	// 比较告警来源
	if alert1.Source == alert2.Source {
		similarity += 0.3
	}

	return similarity, nil
}

// cleanupExpiredWindows 清理过期的收敛窗口
func (acs *AlertConvergerService) cleanupExpiredWindows() {
	now := time.Now()
	for groupKey, window := range acs.convergenceWindows {
		if now.After(window.EndTime) {
			delete(acs.convergenceWindows, groupKey)
		}
	}
}

// StartCleanupRoutine 启动清理例程
func (acs *AlertConvergerService) StartCleanupRoutine(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Minute) // 每分钟清理一次
	defer ticker.Stop()
	
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			acs.cleanupExpiredWindows()
		}
	}
}