package gateway

import (
	"context"
	"fmt"
	"time"

	"alert_agent/internal/domain/gateway"
	"alert_agent/internal/model"
	"alert_agent/internal/pkg/feature"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// SmartGatewayService 智能告警网关服务实现
type SmartGatewayService struct {
	receiver         gateway.AlertReceiver
	processor        gateway.AlertProcessor
	router           gateway.AlertRouter
	suppressor       gateway.AlertSuppressor
	converger        gateway.AlertConverger
	repository       gateway.AlertProcessingRepository
	featureToggle    gateway.FeatureToggleService
	metricsCollector gateway.MetricsCollector
	logger           *zap.Logger

	// 处理策略
	strategies map[gateway.ProcessingMode]gateway.ProcessingStrategy
}

// NewSmartGatewayService 创建智能告警网关服务
func NewSmartGatewayService(
	receiver gateway.AlertReceiver,
	processor gateway.AlertProcessor,
	router gateway.AlertRouter,
	suppressor gateway.AlertSuppressor,
	converger gateway.AlertConverger,
	repository gateway.AlertProcessingRepository,
	featureToggle gateway.FeatureToggleService,
	metricsCollector gateway.MetricsCollector,
	logger *zap.Logger,
) *SmartGatewayService {
	sgs := &SmartGatewayService{
		receiver:         receiver,
		processor:        processor,
		router:           router,
		suppressor:       suppressor,
		converger:        converger,
		repository:       repository,
		featureToggle:    featureToggle,
		metricsCollector: metricsCollector,
		logger:           logger,
		strategies:       make(map[gateway.ProcessingMode]gateway.ProcessingStrategy),
	}

	return sgs
}

// RegisterStrategy 注册处理策略
func (sgs *SmartGatewayService) RegisterStrategy(strategy gateway.ProcessingStrategy) {
	sgs.strategies[strategy.GetMode()] = strategy
}

// ReceiveAlert 接收告警
func (sgs *SmartGatewayService) ReceiveAlert(ctx context.Context, alert *model.Alert) (*gateway.AlertProcessingRecord, error) {
	start := time.Now()
	defer func() {
		sgs.metricsCollector.RecordProcessingLatency(ctx, gateway.ModeDirectPassthrough, time.Since(start).Milliseconds())
	}()

	// 记录告警接收指标
	sgs.metricsCollector.RecordAlertReceived(ctx, alert)

	// 创建处理记录
	record := &gateway.AlertProcessingRecord{
		ID:             uuid.New().String(),
		AlertID:        alert.ID,
		OriginalAlert:  alert,
		Status:         gateway.AlertStatusReceived,
		ReceivedAt:     time.Now(),
		ProcessingSteps: []gateway.ProcessingStep{},
		Metadata:       make(map[string]interface{}),
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	// 保存初始记录
	if err := sgs.repository.Create(ctx, record); err != nil {
		sgs.logger.Error("Failed to create processing record", zap.Error(err), zap.Uint("alert_id", alert.ID))
		sgs.metricsCollector.RecordError(ctx, "create_processing_record", err)
		return nil, fmt.Errorf("failed to create processing record: %w", err)
	}

	// 接收告警
	processingRecord, err := sgs.receiver.Receive(ctx, alert)
	if err != nil {
		sgs.logger.Error("Failed to receive alert", zap.Error(err), zap.Uint("alert_id", alert.ID))
		sgs.metricsCollector.RecordError(ctx, "receive_alert", err)
		
		// 更新记录状态为失败
		record.Status = gateway.AlertStatusFailed
		record.ErrorMessage = err.Error()
		record.UpdatedAt = time.Now()
		sgs.repository.Update(ctx, record)
		
		return record, fmt.Errorf("failed to receive alert: %w", err)
	}

	// 异步处理告警
	go func() {
		processCtx := context.Background() // 使用新的context避免超时
		if err := sgs.processAlertAsync(processCtx, processingRecord); err != nil {
			sgs.logger.Error("Failed to process alert asynchronously", 
				zap.Error(err), 
				zap.String("record_id", processingRecord.ID))
		}
	}()

	return processingRecord, nil
}

// processAlertAsync 异步处理告警
func (sgs *SmartGatewayService) processAlertAsync(ctx context.Context, record *gateway.AlertProcessingRecord) error {
	// 丰富告警上下文
	alertCtx, err := sgs.receiver.EnrichAlert(ctx, record.OriginalAlert)
	if err != nil {
		sgs.logger.Warn("Failed to enrich alert context", zap.Error(err), zap.String("record_id", record.ID))
		// 创建基础上下文
		alertCtx = &gateway.AlertContext{
			Alert: record.OriginalAlert,
		}
	}

	// 获取处理模式
	processingMode := sgs.processor.GetProcessingMode(ctx, alertCtx)
	record.ProcessingMode = processingMode

	// 添加处理步骤
	step := gateway.ProcessingStep{
		Step:      "determine_processing_mode",
		Status:    "completed",
		StartTime: time.Now(),
		EndTime:   &[]time.Time{time.Now()}[0],
		Duration:  0,
		Details: map[string]interface{}{
			"mode": string(processingMode),
		},
	}
	record.ProcessingSteps = append(record.ProcessingSteps, step)

	// 更新记录状态
	record.Status = gateway.AlertStatusProcessing
	record.UpdatedAt = time.Now()
	if err := sgs.repository.Update(ctx, record); err != nil {
		sgs.logger.Error("Failed to update processing record", zap.Error(err))
	}

	// 根据处理模式选择策略
	strategy, exists := sgs.strategies[processingMode]
	if !exists {
		// 默认使用直通模式
		strategy = sgs.strategies[gateway.ModeDirectPassthrough]
		if strategy == nil {
			return fmt.Errorf("no processing strategy available")
		}
	}

	// 执行处理策略
	processedRecord, err := strategy.Process(ctx, alertCtx)
	if err != nil {
		sgs.logger.Error("Failed to process alert with strategy", 
			zap.Error(err), 
			zap.String("mode", string(processingMode)))
		
		// 更新记录状态为失败
		record.Status = gateway.AlertStatusFailed
		record.ErrorMessage = err.Error()
		record.UpdatedAt = time.Now()
		sgs.repository.Update(ctx, record)
		sgs.metricsCollector.RecordError(ctx, "process_alert", err)
		return err
	}

	// 合并处理结果
	if processedRecord != nil {
		record.ProcessingSteps = append(record.ProcessingSteps, processedRecord.ProcessingSteps...)
		record.Status = processedRecord.Status
		record.ProcessedAt = processedRecord.ProcessedAt
		record.RoutedAt = processedRecord.RoutedAt
		if processedRecord.ErrorMessage != "" {
			record.ErrorMessage = processedRecord.ErrorMessage
		}
		// 合并元数据
		for k, v := range processedRecord.Metadata {
			record.Metadata[k] = v
		}
	}

	// 最终更新记录
	record.UpdatedAt = time.Now()
	if err := sgs.repository.Update(ctx, record); err != nil {
		sgs.logger.Error("Failed to update final processing record", zap.Error(err))
	}

	// 记录处理完成指标
	sgs.metricsCollector.RecordAlertProcessed(ctx, record)

	return nil
}

// RouteAlert 路由告警
func (sgs *SmartGatewayService) RouteAlert(ctx context.Context, alertCtx *gateway.AlertContext) (*gateway.RoutingDecision, error) {
	// 检查是否应该抑制
	if sgs.featureToggle.IsEnabled(ctx, string(feature.FeatureAutoSuppression)) {
		suppressed, reason, err := sgs.suppressor.ShouldSuppress(ctx, alertCtx)
		if err != nil {
			sgs.logger.Warn("Failed to check suppression", zap.Error(err))
		} else if suppressed {
			return &gateway.RoutingDecision{
				Suppressed:   true,
				Reason:       reason,
				DecisionTime: time.Now(),
			}, nil
		}
	}

	// 执行路由决策
	decision, err := sgs.router.Route(ctx, alertCtx)
	if err != nil {
		sgs.logger.Error("Failed to route alert", zap.Error(err))
		sgs.metricsCollector.RecordError(ctx, "route_alert", err)
		return nil, fmt.Errorf("failed to route alert: %w", err)
	}

	// 记录路由指标
	sgs.metricsCollector.RecordAlertRouted(ctx, decision)

	return decision, nil
}

// ConvergeAlerts 收敛告警
func (sgs *SmartGatewayService) ConvergeAlerts(ctx context.Context, alerts []*model.Alert) (*gateway.ConvergenceResult, error) {
	if !sgs.featureToggle.IsEnabled(ctx, string(feature.FeatureBasicConvergence)) {
		return &gateway.ConvergenceResult{
			Converged: false,
			Metadata: map[string]interface{}{
				"reason": "convergence feature disabled",
			},
		}, nil
	}

	result, err := sgs.converger.Converge(ctx, alerts)
	if err != nil {
		sgs.logger.Error("Failed to converge alerts", zap.Error(err))
		sgs.metricsCollector.RecordError(ctx, "converge_alerts", err)
		return nil, fmt.Errorf("failed to converge alerts: %w", err)
	}

	return result, nil
}

// GetProcessingRecord 获取处理记录
func (sgs *SmartGatewayService) GetProcessingRecord(ctx context.Context, recordID string) (*gateway.AlertProcessingRecord, error) {
	record, err := sgs.repository.GetByID(ctx, recordID)
	if err != nil {
		sgs.logger.Error("Failed to get processing record", zap.Error(err), zap.String("record_id", recordID))
		return nil, fmt.Errorf("failed to get processing record: %w", err)
	}

	return record, nil
}

// GetStatistics 获取统计信息
func (sgs *SmartGatewayService) GetStatistics(ctx context.Context, timeRange gateway.TimeRange) (*gateway.GatewayStatistics, error) {
	stats, err := sgs.repository.GetStatistics(ctx, timeRange)
	if err != nil {
		sgs.logger.Error("Failed to get gateway statistics", zap.Error(err))
		return nil, fmt.Errorf("failed to get statistics: %w", err)
	}

	return stats, nil
}