package analysis

import (
	"context"
	"fmt"
	"time"

	"alert_agent/internal/domain/alert"
	"alert_agent/internal/domain/analysis"
	"alert_agent/internal/model"
	"alert_agent/internal/shared/logger"

	"go.uber.org/zap"
)

// N8NAnalysisService n8n 异步分析服务
type N8NAnalysisService struct {
	workflowManager analysis.N8NWorkflowManager
	alertRepo       alert.AlertRepository
	executionRepo   analysis.N8NWorkflowExecutionRepository
	logger          *zap.Logger
}

// N8NAnalysisConfig n8n 分析服务配置
type N8NAnalysisConfig struct {
	// 默认工作流模板ID
	DefaultWorkflowTemplateID string `json:"default_workflow_template_id"`
	// 批处理大小
	BatchSize int `json:"batch_size"`
	// 处理间隔
	ProcessInterval time.Duration `json:"process_interval"`
	// 重试次数
	MaxRetries int `json:"max_retries"`
	// 超时时间
	Timeout time.Duration `json:"timeout"`
	// 是否启用自动分析
	AutoAnalysisEnabled bool `json:"auto_analysis_enabled"`
}

// NewN8NAnalysisService 创建新的 n8n 分析服务
func NewN8NAnalysisService(
	workflowManager analysis.N8NWorkflowManager,
	alertRepo alert.AlertRepository,
	executionRepo analysis.N8NWorkflowExecutionRepository,
) *N8NAnalysisService {
	return &N8NAnalysisService{
		workflowManager: workflowManager,
		alertRepo:       alertRepo,
		executionRepo:   executionRepo,
		logger:          logger.GetLogger(),
	}
}

// AnalyzeAlert 分析单个告警
func (s *N8NAnalysisService) AnalyzeAlert(ctx context.Context, alertID uint, workflowTemplateID string) (*analysis.N8NWorkflowExecution, error) {
	// 获取告警信息
	_, err := s.alertRepo.GetByID(ctx, alertID)
	if err != nil {
		return nil, fmt.Errorf("failed to get alert: %w", err)
	}

	// 检查告警是否已经在分析中
	executions, err := s.executionRepo.ListByWorkflowID(ctx, workflowTemplateID, 1, 0)
	if err != nil {
		s.logger.Warn("Failed to check existing executions", zap.Error(err))
	}

	// 检查是否有正在运行的分析
	for _, exec := range executions {
		if exec.Status == analysis.N8NWorkflowStatusRunning {
			if exec.Metadata != nil {
				if alertIDStr, ok := exec.Metadata["alert_id"]; ok && alertIDStr == fmt.Sprintf("%d", alertID) {
					return exec, nil // 返回正在运行的执行
				}
			}
		}
	}

	// 触发工作流分析
	metadata := map[string]interface{}{
		"alert_id": fmt.Sprintf("%d", alertID),
		"type":     "alert_analysis",
	}
	execution, err := s.workflowManager.TriggerAnalysisWorkflow(ctx, fmt.Sprintf("%d", alertID), workflowTemplateID, metadata)
	if err != nil {
		return nil, fmt.Errorf("failed to trigger analysis workflow: %w", err)
	}

	s.logger.Info("Alert analysis triggered",
		zap.Uint("alert_id", alertID),
		zap.String("execution_id", execution.ID),
		zap.String("workflow_template_id", workflowTemplateID))

	return execution, nil
}

// BatchAnalyzeAlerts 批量分析告警
func (s *N8NAnalysisService) BatchAnalyzeAlerts(ctx context.Context, config N8NAnalysisConfig) error {
	// 获取需要分析的告警
	alerts, err := s.alertRepo.GetAlertsForAnalysis(ctx, config.BatchSize)
	if err != nil {
		return fmt.Errorf("failed to get alerts for analysis: %w", err)
	}

	if len(alerts) == 0 {
		s.logger.Debug("No alerts need analysis")
		return nil
	}

	s.logger.Info("Starting batch analysis", zap.Int("alert_count", len(alerts)))

	// 并发分析告警
	semaphore := make(chan struct{}, 5) // 限制并发数
	errorChan := make(chan error, len(alerts))

	for _, alert := range alerts {
		go func(alert *model.Alert) {
			semaphore <- struct{}{} // 获取信号量
			defer func() { <-semaphore }() // 释放信号量

			_, err := s.AnalyzeAlert(ctx, alert.ID, config.DefaultWorkflowTemplateID)
			if err != nil {
				s.logger.Error("Failed to analyze alert",
					zap.Uint("alert_id", alert.ID),
					zap.Error(err))
				errorChan <- err
			} else {
				errorChan <- nil
			}
		}(alert)
	}

	// 等待所有分析完成
	var errors []error
	for i := 0; i < len(alerts); i++ {
		if err := <-errorChan; err != nil {
			errors = append(errors, err)
		}
	}

	if len(errors) > 0 {
		s.logger.Warn("Some alerts failed to analyze", zap.Int("error_count", len(errors)))
		return fmt.Errorf("batch analysis completed with %d errors", len(errors))
	}

	s.logger.Info("Batch analysis completed successfully", zap.Int("alert_count", len(alerts)))
	return nil
}

// GetAnalysisStatus 获取分析状态
func (s *N8NAnalysisService) GetAnalysisStatus(ctx context.Context, executionID string) (*analysis.N8NWorkflowExecution, error) {
	execution, err := s.executionRepo.GetByID(ctx, executionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get execution: %w", err)
	}

	// 如果执行还在运行中，尝试从工作流管理器获取最新状态
	if execution.Status == analysis.N8NWorkflowStatusRunning {
		latestExecution, err := s.workflowManager.GetExecutionLogs(ctx, executionID)
		if err != nil {
			s.logger.Warn("Failed to get latest execution status", zap.Error(err))
			return execution, nil
		}

		// 更新执行状态
		if len(latestExecution) > 0 {
			// 这里可以根据日志判断状态变化
			if err := s.executionRepo.Update(ctx, execution); err != nil {
				s.logger.Error("Failed to update execution status", zap.Error(err))
			}
		}
	}

	return execution, nil
}

// CancelAnalysis 取消分析
func (s *N8NAnalysisService) CancelAnalysis(ctx context.Context, executionID string) error {
	// 更新执行状态为取消
	execution, err := s.executionRepo.GetByID(ctx, executionID)
	if err != nil {
		s.logger.Warn("Failed to get execution for status update", zap.Error(err))
		return fmt.Errorf("failed to get execution: %w", err)
	}

	execution.Status = analysis.N8NWorkflowStatusFailed

	if err := s.executionRepo.Update(ctx, execution); err != nil {
		s.logger.Error("Failed to update cancelled execution status", zap.Error(err))
		return fmt.Errorf("failed to update execution status: %w", err)
	}

	s.logger.Info("Analysis cancelled", zap.String("execution_id", executionID))
	return nil
}

// RetryAnalysis 重试分析
func (s *N8NAnalysisService) RetryAnalysis(ctx context.Context, executionID string) (*analysis.N8NWorkflowExecution, error) {
	// 获取原始执行信息
	originalExecution, err := s.executionRepo.GetByID(ctx, executionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get original execution: %w", err)
	}

	// 检查是否可以重试
	if originalExecution.Status == analysis.N8NWorkflowStatusRunning {
		return nil, fmt.Errorf("cannot retry running execution")
	}

	// 从元数据中获取告警ID
	alertIDInterface, ok := originalExecution.Metadata["alert_id"]
	if !ok {
		return nil, fmt.Errorf("alert_id not found in execution metadata")
	}

	alertIDStr, ok := alertIDInterface.(string)
	if !ok {
		return nil, fmt.Errorf("alert_id is not a string")
	}

	// 重新触发分析
	metadata := map[string]interface{}{
		"alert_id": alertIDStr,
		"type":     "alert_analysis_retry",
		"original_execution_id": executionID,
	}
	newExecution, err := s.workflowManager.TriggerAnalysisWorkflow(ctx, alertIDStr, originalExecution.WorkflowID, metadata)
	if err != nil {
		return nil, fmt.Errorf("failed to retry analysis workflow: %w", err)
	}

	s.logger.Info("Analysis retried",
		zap.String("original_execution_id", executionID),
		zap.String("new_execution_id", newExecution.ID),
		zap.String("alert_id", alertIDStr))

	return newExecution, nil
}

// GetAnalysisHistory 获取分析历史
func (s *N8NAnalysisService) GetAnalysisHistory(ctx context.Context, alertID uint, limit int) ([]*analysis.N8NWorkflowExecution, error) {
	// 通过元数据查找相关的执行记录
	// 注意：这里需要根据实际的存储结构来实现查询逻辑
	executions, err := s.executionRepo.ListByDateRange(ctx, time.Time{}, time.Now(), limit, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to get executions: %w", err)
	}

	// 过滤出与指定告警相关的执行
	var result []*analysis.N8NWorkflowExecution
	alertIDStr := fmt.Sprintf("%d", alertID)
	for _, exec := range executions {
		if exec.Metadata != nil {
			if execAlertID, ok := exec.Metadata["alert_id"]; ok && execAlertID == alertIDStr {
				result = append(result, exec)
			}
		}
	}

	return result, nil
}

// GetAnalysisMetrics 获取分析指标
func (s *N8NAnalysisService) GetAnalysisMetrics(ctx context.Context, timeRange alert.TimeRange) (*AnalysisMetrics, error) {
	// 获取指定时间范围内的执行统计
	stats, err := s.executionRepo.GetStatistics(ctx, "all", *timeRange.Start, *timeRange.End)
	if err != nil {
		return nil, fmt.Errorf("failed to get execution statistics: %w", err)
	}

	// 从统计结果中提取数据
	totalExecutions, _ := stats["total_executions"].(int64)
	successfulExecutions, _ := stats["successful_executions"].(int64)
	failedExecutions, _ := stats["failed_executions"].(int64)
	runningExecutions, _ := stats["running_executions"].(int64)
	averageExecutionTime, _ := stats["average_execution_time"].(time.Duration)

	return &AnalysisMetrics{
		TotalExecutions:       totalExecutions,
		SuccessfulExecutions:  successfulExecutions,
		FailedExecutions:      failedExecutions,
		RunningExecutions:     runningExecutions,
		AverageExecutionTime:  averageExecutionTime,
		TimeRange:             timeRange,
	}, nil
}

// AnalysisMetrics 分析指标
type AnalysisMetrics struct {
	TotalExecutions       int64             `json:"total_executions"`
	SuccessfulExecutions  int64             `json:"successful_executions"`
	FailedExecutions      int64             `json:"failed_executions"`
	RunningExecutions     int64             `json:"running_executions"`
	AverageExecutionTime  time.Duration     `json:"average_execution_time"`
	TimeRange             alert.TimeRange   `json:"time_range"`
}

// StartAutoAnalysis 启动自动分析
func (s *N8NAnalysisService) StartAutoAnalysis(ctx context.Context, config N8NAnalysisConfig) {
	if !config.AutoAnalysisEnabled {
		s.logger.Info("Auto analysis is disabled")
		return
	}

	s.logger.Info("Starting auto analysis",
		zap.Duration("interval", config.ProcessInterval),
		zap.Int("batch_size", config.BatchSize))

	ticker := time.NewTicker(config.ProcessInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			s.logger.Info("Auto analysis stopped")
			return
		case <-ticker.C:
			if err := s.BatchAnalyzeAlerts(ctx, config); err != nil {
				s.logger.Error("Auto analysis failed", zap.Error(err))
			}
		}
	}
}