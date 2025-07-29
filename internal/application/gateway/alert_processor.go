package gateway

import (
	"context"
	"fmt"
	"time"

	"alert_agent/internal/domain/gateway"
	"alert_agent/internal/model"
	"alert_agent/internal/pkg/feature"

	"go.uber.org/zap"
)

// AlertProcessorService 告警处理器服务实现
type AlertProcessorService struct {
	repository         gateway.AlertProcessingRepository
	featureToggle      gateway.FeatureToggleService
	metricsCollector   gateway.MetricsCollector
	strategies         map[gateway.ProcessingMode]gateway.ProcessingStrategy
	logger             *zap.Logger
}

// NewAlertProcessorService 创建告警处理器服务
func NewAlertProcessorService(
	repository gateway.AlertProcessingRepository,
	featureToggle gateway.FeatureToggleService,
	metricsCollector gateway.MetricsCollector,
	logger *zap.Logger,
) *AlertProcessorService {
	aps := &AlertProcessorService{
		repository:       repository,
		featureToggle:    featureToggle,
		metricsCollector: metricsCollector,
		strategies:       make(map[gateway.ProcessingMode]gateway.ProcessingStrategy),
		logger:           logger,
	}

	// 注册处理策略
	aps.registerStrategies()

	return aps
}

// Process 处理告警
func (aps *AlertProcessorService) Process(ctx context.Context, alertCtx *gateway.AlertContext) (*gateway.AlertProcessingRecord, error) {
	start := time.Now()

	// 创建处理记录
	record := &gateway.AlertProcessingRecord{
		AlertID:        alertCtx.Alert.ID,
		OriginalAlert:  alertCtx.Alert,
		Status:         gateway.AlertStatusProcessing,
		ReceivedAt:     time.Now(),
		ProcessingSteps: []gateway.ProcessingStep{},
		Metadata:       make(map[string]interface{}),
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	// 确定处理模式
	mode := aps.GetProcessingMode(ctx, alertCtx)
	record.ProcessingMode = mode

	// 添加处理步骤
	step := gateway.ProcessingStep{
		Step:      "processing_started",
		Status:    "in_progress",
		StartTime: start,
		Details: map[string]interface{}{
			"processing_mode": string(mode),
		},
	}
	record.ProcessingSteps = append(record.ProcessingSteps, step)

	// 获取对应的处理策略
	strategy, exists := aps.strategies[mode]
	if !exists {
		aps.logger.Error("No strategy found for processing mode", zap.String("mode", string(mode)))
		return nil, fmt.Errorf("no strategy found for processing mode: %s", mode)
	}

	// 执行处理策略
	processedRecord, err := strategy.Process(ctx, alertCtx)
	if err != nil {
		aps.logger.Error("Strategy processing failed", zap.Error(err), zap.String("mode", string(mode)))
		
		// 更新失败状态
		record.Status = gateway.AlertStatusFailed
		step.Status = "failed"
		step.ErrorMsg = err.Error()
		endTime := time.Now()
		step.EndTime = &endTime
		step.Duration = time.Since(start)
		
		// 更新处理记录
		record.ProcessingSteps[len(record.ProcessingSteps)-1] = step
		record.UpdatedAt = time.Now()
		
		if updateErr := aps.repository.Update(ctx, record); updateErr != nil {
			aps.logger.Error("Failed to update processing record", zap.Error(updateErr))
		}
		
		aps.metricsCollector.RecordError(ctx, "alert_processing", err)
		return nil, fmt.Errorf("strategy processing failed: %w", err)
	}

	// 更新成功状态
	record.Status = gateway.AlertStatusRouted
	step.Status = "completed"
	endTime := time.Now()
	step.EndTime = &endTime
	step.Duration = time.Since(start)

	// 更新处理记录
	record.ProcessingSteps[len(record.ProcessingSteps)-1] = step
	record.ProcessedAt = &endTime
	record.UpdatedAt = endTime

	if err := aps.repository.Update(ctx, record); err != nil {
		aps.logger.Error("Failed to update processing record", zap.Error(err))
		return nil, fmt.Errorf("failed to update processing record: %w", err)
	}

	aps.logger.Info("Alert processed successfully", 
		zap.String("record_id", record.ID),
		zap.String("processing_mode", string(mode)),
		zap.Duration("duration", time.Since(start)))

	return processedRecord, nil
}

// GetProcessingMode 获取处理模式
func (aps *AlertProcessorService) GetProcessingMode(ctx context.Context, alertCtx *gateway.AlertContext) gateway.ProcessingMode {
	// 检查功能开关状态
	directRoutingEnabled := aps.featureToggle.IsEnabled(ctx, string(feature.FeatureDirectRouting))
	basicConvergenceEnabled := aps.featureToggle.IsEnabled(ctx, string(feature.FeatureBasicConvergence))
	smartRoutingEnabled := aps.featureToggle.IsEnabled(ctx, string(feature.FeatureSmartRouting))

	// 根据告警级别和功能开关确定处理模式
	alert := alertCtx.Alert

	// 关键告警优先使用直通模式
	if alert.Level == model.AlertLevelCritical && directRoutingEnabled {
		return gateway.ModeDirectPassthrough
	}

	// 如果启用了智能路由，优先使用智能路由
	if smartRoutingEnabled {
		return gateway.ModeSmartRouting
	}

	// 如果启用了基础聚合，使用基础聚合模式
	if basicConvergenceEnabled {
		return gateway.ModeBasicConvergence
	}

	// 默认使用直通模式
	return gateway.ModeDirectPassthrough
}

// UpdateProcessingRecord 更新处理记录
func (aps *AlertProcessorService) UpdateProcessingRecord(ctx context.Context, record *gateway.AlertProcessingRecord, step gateway.ProcessingStep) error {
	record.ProcessingSteps = append(record.ProcessingSteps, step)
	record.UpdatedAt = time.Now()
	return aps.repository.Update(ctx, record)
}

// registerStrategies 注册处理策略
func (aps *AlertProcessorService) registerStrategies() {
	// 注册直通策略
	aps.strategies[gateway.ModeDirectPassthrough] = NewDirectPassthroughStrategy(
		aps.repository,
		aps.metricsCollector,
		aps.logger,
	)

	// 注册基础聚合策略
	aps.strategies[gateway.ModeBasicConvergence] = NewBasicConvergenceStrategy(
		aps.repository,
		aps.metricsCollector,
		aps.logger,
	)

	// 注册智能路由策略
	aps.strategies[gateway.ModeSmartRouting] = NewSmartRoutingStrategy(
		aps.repository,
		aps.metricsCollector,
		aps.logger,
	)
}

// DirectPassthroughStrategy 直通策略实现
type DirectPassthroughStrategy struct {
	repository       gateway.AlertProcessingRepository
	metricsCollector gateway.MetricsCollector
	logger           *zap.Logger
}

// NewDirectPassthroughStrategy 创建直通策略
func NewDirectPassthroughStrategy(
	repository gateway.AlertProcessingRepository,
	metricsCollector gateway.MetricsCollector,
	logger *zap.Logger,
) *DirectPassthroughStrategy {
	return &DirectPassthroughStrategy{
		repository:       repository,
		metricsCollector: metricsCollector,
		logger:           logger,
	}
}

// CanHandle 判断是否可以处理该告警
func (dps *DirectPassthroughStrategy) CanHandle(ctx context.Context, alertCtx *gateway.AlertContext) bool {
	return true // 直通策略可以处理所有告警
}

// Process 直通处理
func (dps *DirectPassthroughStrategy) Process(ctx context.Context, alertCtx *gateway.AlertContext) (*gateway.AlertProcessingRecord, error) {
	start := time.Now()

	// 创建处理记录
	record := &gateway.AlertProcessingRecord{
		AlertID:        alertCtx.Alert.ID,
		OriginalAlert:  alertCtx.Alert,
		ProcessingMode: gateway.ModeDirectPassthrough,
		Status:         gateway.AlertStatusRouted,
		ReceivedAt:     time.Now(),
		ProcessingSteps: []gateway.ProcessingStep{},
		Metadata:       make(map[string]interface{}),
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	// 直通模式不做任何处理，直接标记为可路由
	step := gateway.ProcessingStep{
		Step:      "direct_passthrough",
		Status:    "completed",
		StartTime: start,
		EndTime:   &[]time.Time{time.Now()}[0],
		Duration:  time.Since(start),
		Details: map[string]interface{}{
			"strategy": "direct_passthrough",
			"action":   "no_processing_required",
		},
	}

	record.ProcessingSteps = append(record.ProcessingSteps, step)

	dps.logger.Debug("Direct passthrough processing completed", 
		zap.String("record_id", record.ID),
		zap.Duration("duration", time.Since(start)))

	return record, nil
}

// GetMode 获取处理模式
func (dps *DirectPassthroughStrategy) GetMode() gateway.ProcessingMode {
	return gateway.ModeDirectPassthrough
}

// BasicConvergenceStrategy 基础聚合策略实现
type BasicConvergenceStrategy struct {
	repository       gateway.AlertProcessingRepository
	metricsCollector gateway.MetricsCollector
	logger           *zap.Logger
}

// NewBasicConvergenceStrategy 创建基础聚合策略
func NewBasicConvergenceStrategy(
	repository gateway.AlertProcessingRepository,
	metricsCollector gateway.MetricsCollector,
	logger *zap.Logger,
) *BasicConvergenceStrategy {
	return &BasicConvergenceStrategy{
		repository:       repository,
		metricsCollector: metricsCollector,
		logger:           logger,
	}
}

// CanHandle 判断是否可以处理该告警
func (bcs *BasicConvergenceStrategy) CanHandle(ctx context.Context, alertCtx *gateway.AlertContext) bool {
	return true // 基础聚合策略可以处理所有告警
}

// Process 基础聚合处理
func (bcs *BasicConvergenceStrategy) Process(ctx context.Context, alertCtx *gateway.AlertContext) (*gateway.AlertProcessingRecord, error) {
	start := time.Now()

	// 创建处理记录
	record := &gateway.AlertProcessingRecord{
		AlertID:        alertCtx.Alert.ID,
		OriginalAlert:  alertCtx.Alert,
		ProcessingMode: gateway.ModeBasicConvergence,
		Status:         gateway.AlertStatusConverged,
		ReceivedAt:     time.Now(),
		ProcessingSteps: []gateway.ProcessingStep{},
		Metadata:       make(map[string]interface{}),
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	// 基础聚合逻辑：检查是否有相似的告警
	similarAlerts, err := bcs.findSimilarAlerts(ctx, alertCtx.Alert)
	if err != nil {
		bcs.logger.Error("Failed to find similar alerts", zap.Error(err))
		return nil, fmt.Errorf("failed to find similar alerts: %w", err)
	}

	step := gateway.ProcessingStep{
		Step:      "basic_convergence",
		Status:    "completed",
		StartTime: start,
		EndTime:   &[]time.Time{time.Now()}[0],
		Duration:  time.Since(start),
		Details: map[string]interface{}{
			"strategy":           "basic_convergence",
			"similar_alerts":     len(similarAlerts),
			"convergence_action": "group_similar_alerts",
		},
	}

	// 如果找到相似告警，进行聚合
	if len(similarAlerts) > 0 {
		step.Details["convergence_applied"] = true
		step.Details["grouped_with"] = similarAlerts
		
		// 更新记录的聚合信息
		record.Metadata["convergence_group"] = similarAlerts
		record.Metadata["is_grouped"] = true
	} else {
		step.Details["convergence_applied"] = false
		step.Details["reason"] = "no_similar_alerts_found"
	}

	record.ProcessingSteps = append(record.ProcessingSteps, step)

	bcs.logger.Debug("Basic convergence processing completed", 
		zap.String("record_id", record.ID),
		zap.Int("similar_alerts", len(similarAlerts)),
		zap.Duration("duration", time.Since(start)))

	return record, nil
}

// GetMode 获取处理模式
func (bcs *BasicConvergenceStrategy) GetMode() gateway.ProcessingMode {
	return gateway.ModeBasicConvergence
}

// findSimilarAlerts 查找相似告警
func (bcs *BasicConvergenceStrategy) findSimilarAlerts(ctx context.Context, alert *model.Alert) ([]string, error) {
	// 简化实现：基于告警名称和级别查找相似告警
	// 在实际实现中，这里会查询数据库
	
	// 模拟查找逻辑
	var similarAlerts []string
	
	// 这里可以实现更复杂的相似性算法
	// 例如：基于告警名称、标签、来源等进行匹配
	
	return similarAlerts, nil
}

// SmartRoutingStrategy 智能路由策略实现
type SmartRoutingStrategy struct {
	repository       gateway.AlertProcessingRepository
	metricsCollector gateway.MetricsCollector
	logger           *zap.Logger
}

// NewSmartRoutingStrategy 创建智能路由策略
func NewSmartRoutingStrategy(
	repository gateway.AlertProcessingRepository,
	metricsCollector gateway.MetricsCollector,
	logger *zap.Logger,
) *SmartRoutingStrategy {
	return &SmartRoutingStrategy{
		repository:       repository,
		metricsCollector: metricsCollector,
		logger:           logger,
	}
}

// CanHandle 判断是否可以处理该告警
func (srs *SmartRoutingStrategy) CanHandle(ctx context.Context, alertCtx *gateway.AlertContext) bool {
	return true // 智能路由策略可以处理所有告警
}

// Process 智能路由处理
func (srs *SmartRoutingStrategy) Process(ctx context.Context, alertCtx *gateway.AlertContext) (*gateway.AlertProcessingRecord, error) {
	start := time.Now()

	// 创建处理记录
	record := &gateway.AlertProcessingRecord{
		AlertID:        alertCtx.Alert.ID,
		OriginalAlert:  alertCtx.Alert,
		ProcessingMode: gateway.ModeSmartRouting,
		Status:         gateway.AlertStatusRouted,
		ReceivedAt:     time.Now(),
		ProcessingSteps: []gateway.ProcessingStep{},
		Metadata:       make(map[string]interface{}),
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	// 智能路由逻辑：基于AI分析结果进行处理
	analysisResult, err := srs.performIntelligentAnalysis(ctx, alertCtx.Alert)
	if err != nil {
		srs.logger.Error("Failed to perform intelligent analysis", zap.Error(err))
		return nil, fmt.Errorf("failed to perform intelligent analysis: %w", err)
	}

	step := gateway.ProcessingStep{
		Step:      "smart_routing",
		Status:    "completed",
		StartTime: start,
		EndTime:   &[]time.Time{time.Now()}[0],
		Duration:  time.Since(start),
		Details: map[string]interface{}{
			"strategy":        "smart_routing",
			"ai_confidence":   analysisResult.Confidence,
			"severity":        analysisResult.Severity,
			"category":        analysisResult.Category,
		},
	}

	// 更新记录的AI分析结果
	record.Metadata["ai_analysis"] = analysisResult
	record.Metadata["smart_routing_applied"] = true

	record.ProcessingSteps = append(record.ProcessingSteps, step)

	srs.logger.Debug("Smart routing processing completed", 
		zap.String("record_id", record.ID),
		zap.Float64("confidence", analysisResult.Confidence),
		zap.Duration("duration", time.Since(start)))

	return record, nil
}

// GetMode 获取处理模式
func (srs *SmartRoutingStrategy) GetMode() gateway.ProcessingMode {
	return gateway.ModeSmartRouting
}

// performIntelligentAnalysis 执行智能分析
func (srs *SmartRoutingStrategy) performIntelligentAnalysis(ctx context.Context, alert *model.Alert) (*gateway.AnalysisResult, error) {
	// 简化的AI分析实现
	// 在实际实现中，这里会调用AI服务进行分析
	
	result := &gateway.AnalysisResult{
		Severity:     alert.Level,
		Category:     categorizeAlertByName(alert),
		RootCause:    "unknown",
		Impact:       "medium",
		Recommendations: []string{"investigate", "monitor"},
		Confidence:   0.85, // 模拟置信度
		Metadata: map[string]interface{}{
			"alert_pattern": "known_pattern",
			"historical_match": true,
			"severity_assessment": "medium",
			"risk_score": calculateRiskScore(alert),
			"recommended_action": determineRecommendedAction(alert),
		},
	}

	return result, nil
}

// categorizeAlertByName 对告警进行分类
func categorizeAlertByName(alert *model.Alert) string {
	name := alert.Name
	if name == "" {
		name = alert.Title
	}

	// 简化的分类逻辑
	switch {
	case containsAny(name, "cpu", "CPU"):
		return "performance"
	case containsAny(name, "memory", "Memory", "RAM"):
		return "performance"
	case containsAny(name, "disk", "Disk", "storage"):
		return "storage"
	case containsAny(name, "network", "Network", "connection"):
		return "network"
	case containsAny(name, "service", "Service", "application"):
		return "application"
	case containsAny(name, "database", "Database", "DB"):
		return "database"
	default:
		return "general"
	}
}

// containsAny 检查字符串是否包含任意一个子字符串
func containsAny(s string, substrings ...string) bool {
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

// calculateRiskScore 计算风险评分
func calculateRiskScore(alert *model.Alert) float64 {
	// 基于告警级别计算风险评分
	switch alert.Level {
	case model.AlertLevelCritical:
		return 0.9
	case model.AlertLevelHigh:
		return 0.7
	case model.AlertLevelMedium:
		return 0.5
	case model.AlertLevelLow:
		return 0.3
	default:
		return 0.1
	}
}

// determineRecommendedAction 确定推荐操作
func determineRecommendedAction(alert *model.Alert) string {
	// 基于告警级别确定推荐操作
	switch alert.Level {
	case model.AlertLevelCritical:
		return "immediate_escalation"
	case model.AlertLevelHigh:
		return "priority_routing"
	case model.AlertLevelMedium:
		return "standard_routing"
	case model.AlertLevelLow:
		return "batch_processing"
	default:
		return "standard_routing"
	}
}