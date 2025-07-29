package gateway

import (
	"context"
	"fmt"
	"time"

	"alert_agent/internal/domain/gateway"
	"alert_agent/internal/model"
	"alert_agent/internal/pkg/feature"
)

// SmartGatewayImpl 智能网关服务实现
type SmartGatewayImpl struct {
	alertReceiver    gateway.AlertReceiver
	alertProcessor   gateway.AlertProcessor
	alertRouter      gateway.AlertRouter
	alertSuppressor  gateway.AlertSuppressor
	alertConverger   gateway.AlertConverger
	processingRepo   gateway.AlertProcessingRepository
	featureToggle    feature.ToggleManager
	metricsCollector gateway.MetricsCollector
}

// NewSmartGatewayImpl 创建新的智能网关实现
func NewSmartGatewayImpl(
	alertReceiver gateway.AlertReceiver,
	alertProcessor gateway.AlertProcessor,
	alertRouter gateway.AlertRouter,
	alertSuppressor gateway.AlertSuppressor,
	alertConverger gateway.AlertConverger,
	processingRepo gateway.AlertProcessingRepository,
	featureToggle feature.ToggleManager,
	metricsCollector gateway.MetricsCollector,
) gateway.SmartGateway {
	return &SmartGatewayImpl{
		alertReceiver:    alertReceiver,
		alertProcessor:   alertProcessor,
		alertRouter:      alertRouter,
		alertSuppressor:  alertSuppressor,
		alertConverger:   alertConverger,
		processingRepo:   processingRepo,
		featureToggle:    featureToggle,
		metricsCollector: metricsCollector,
	}
}

// ReceiveAlert 接收告警
func (sg *SmartGatewayImpl) ReceiveAlert(ctx context.Context, alert *model.Alert) (*gateway.AlertProcessingRecord, error) {
	startTime := time.Now()
	defer func() {
		sg.metricsCollector.RecordProcessingLatency(ctx, gateway.ModeDirectPassthrough, int64(time.Since(startTime)))
	}()

	// 接收和验证告警
	record, err := sg.alertReceiver.Receive(ctx, alert)
	if err != nil {
		sg.metricsCollector.RecordError(ctx, "receive_alert_failed", err)
		return nil, fmt.Errorf("failed to receive alert: %w", err)
	}

	// 记录接收指标
	sg.metricsCollector.RecordAlertReceived(ctx, alert)

	return record, nil
}

// ProcessAlert 处理告警
func (sg *SmartGatewayImpl) ProcessAlert(ctx context.Context, alertCtx *gateway.AlertContext) (*gateway.AlertProcessingRecord, error) {
	startTime := time.Now()
	
	// 处理告警
	record, err := sg.alertProcessor.Process(ctx, alertCtx)
	if err != nil {
		sg.metricsCollector.RecordError(ctx, "process_alert_failed", err)
		return nil, fmt.Errorf("failed to process alert: %w", err)
	}

	// 记录处理指标
	sg.metricsCollector.RecordProcessingLatency(ctx, record.ProcessingMode, int64(time.Since(startTime)))
	sg.metricsCollector.RecordAlertProcessed(ctx, record)

	return record, nil
}

// RouteAlert 路由告警
func (sg *SmartGatewayImpl) RouteAlert(ctx context.Context, alertCtx *gateway.AlertContext) (*gateway.RoutingDecision, error) {
	startTime := time.Now()
	defer func() {
		sg.metricsCollector.RecordProcessingLatency(ctx, gateway.ModeDirectPassthrough, int64(time.Since(startTime)))
	}()

	// 路由告警
	decision, err := sg.alertRouter.Route(ctx, alertCtx)
	if err != nil {
		sg.metricsCollector.RecordError(ctx, "route_alert_failed", err)
		return nil, fmt.Errorf("failed to route alert: %w", err)
	}

	// 记录路由指标
	sg.metricsCollector.RecordAlertRouted(ctx, decision)

	return decision, nil
}

// ConvergeAlerts 收敛告警
func (sg *SmartGatewayImpl) ConvergeAlerts(ctx context.Context, alerts []*model.Alert) (*gateway.ConvergenceResult, error) {
	startTime := time.Now()
	defer func() {
		sg.metricsCollector.RecordProcessingLatency(ctx, gateway.ModeBasicConvergence, int64(time.Since(startTime)))
	}()

	// 检查收敛功能是否启用
	if !sg.featureToggle.IsEnabled(ctx, feature.FeatureBasicConvergence) {
		return nil, fmt.Errorf("convergence feature is disabled")
	}

	// 收敛告警
	result, err := sg.alertConverger.Converge(ctx, alerts)
	if err != nil {
		sg.metricsCollector.RecordError(ctx, "converge_alerts_failed", err)
		return nil, fmt.Errorf("failed to converge alerts: %w", err)
	}

	return result, nil
}

// GetProcessingRecord 获取处理记录
func (sg *SmartGatewayImpl) GetProcessingRecord(ctx context.Context, recordID string) (*gateway.AlertProcessingRecord, error) {
	return sg.processingRepo.GetByID(ctx, recordID)
}

// GetProcessingRecords 获取处理记录列表
func (sg *SmartGatewayImpl) GetProcessingRecords(ctx context.Context, filter gateway.AlertProcessingFilter) ([]*gateway.AlertProcessingRecord, error) {
	return sg.processingRepo.List(ctx, filter)
}

// GetStatistics 获取统计信息
func (sg *SmartGatewayImpl) GetStatistics(ctx context.Context, timeRange gateway.TimeRange) (*gateway.GatewayStatistics, error) {
	// 从处理记录中统计数据
	filter := gateway.AlertProcessingFilter{
		TimeRange: &timeRange,
	}
	
	records, err := sg.processingRepo.List(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get processing records: %w", err)
	}

	// 计算统计信息
	stats := &gateway.GatewayStatistics{
		ProcessingModes: make(map[gateway.ProcessingMode]int64),
		LastUpdated:     time.Now(),
	}

	for _, record := range records {
		// 统计处理模式
		stats.ProcessingModes[record.ProcessingMode]++
	}

	return stats, nil
}

// ProcessAlertPipeline 完整的告警处理流水线
func (sg *SmartGatewayImpl) ProcessAlertPipeline(ctx context.Context, alertCtx *gateway.AlertContext) (*gateway.AlertProcessingRecord, error) {
	startTime := time.Now()
	defer func() {
		sg.metricsCollector.RecordProcessingLatency(ctx, gateway.ModeSmartRouting, int64(time.Since(startTime)))
	}()

	// 1. 接收告警
	record, err := sg.ReceiveAlert(ctx, alertCtx.Alert)
	if err != nil {
		return nil, fmt.Errorf("receive alert failed: %w", err)
	}

	// 2. 检查是否应该抑制
	shouldSuppress, suppressReason, err := sg.alertSuppressor.ShouldSuppress(ctx, alertCtx)
	if err != nil {
		sg.metricsCollector.RecordError(ctx, "suppression_check_failed", err)
		return record, fmt.Errorf("suppression check failed: %w", err)
	}

	if shouldSuppress {
		// 更新记录状态为抑制
		record.Status = gateway.AlertStatusSuppressed
		record.Metadata["suppression_reason"] = suppressReason
		
		// 保存记录
		if err := sg.processingRepo.Update(ctx, record); err != nil {
			sg.metricsCollector.RecordError(ctx, "update_record_failed", err)
		}
		
		return record, nil
	}

	// 3. 检查是否应该收敛
	convergenceResult, err := sg.alertConverger.Converge(ctx, []*model.Alert{alertCtx.Alert})
	if err != nil {
		sg.metricsCollector.RecordError(ctx, "convergence_check_failed", err)
		return record, fmt.Errorf("convergence check failed: %w", err)
	}

	if convergenceResult.Converged {
		// 更新记录状态为收敛
		record.Status = gateway.AlertStatusConverged
		record.Metadata["convergence_result"] = convergenceResult
		
		// 保存记录
		if err := sg.processingRepo.Update(ctx, record); err != nil {
			sg.metricsCollector.RecordError(ctx, "update_record_failed", err)
		}
		
		return record, nil
	}

	// 4. 处理告警
	processedRecord, err := sg.ProcessAlert(ctx, alertCtx)
	if err != nil {
		return record, fmt.Errorf("process alert failed: %w", err)
	}

	// 5. 路由告警
	routingDecision, err := sg.RouteAlert(ctx, alertCtx)
	if err != nil {
		return processedRecord, fmt.Errorf("route alert failed: %w", err)
	}

	// 6. 更新记录状态为已路由
	processedRecord.Status = gateway.AlertStatusRouted
	processedRecord.Metadata["routing_decision"] = routingDecision
	
	// 保存最终记录
	if err := sg.processingRepo.Update(ctx, processedRecord); err != nil {
		sg.metricsCollector.RecordError(ctx, "update_final_record_failed", err)
	}

	return processedRecord, nil
}

// Health 健康检查
func (sg *SmartGatewayImpl) Health(ctx context.Context) error {
	// 检查各个组件的健康状态
	// 这里可以添加具体的健康检查逻辑
	// 例如检查数据库连接、外部服务状态等
	
	return nil
}