package n8n

import (
	"context"
	"fmt"
	"strconv"
	"sync"
	"time"

	"alert_agent/internal/domain/alert"
	"alert_agent/internal/domain/analysis"
	"go.uber.org/zap"
)

// WorkflowManagerConfig 工作流管理器配置
type WorkflowManagerConfig struct {
	MonitorInterval    time.Duration `json:"monitor_interval"`
	MaxRetryAttempts   int           `json:"max_retry_attempts"`
	RetryDelay         time.Duration `json:"retry_delay"`
	ExecutionTimeout   time.Duration `json:"execution_timeout"`
	CallbackTimeout    time.Duration `json:"callback_timeout"`
	MaxConcurrentJobs  int           `json:"max_concurrent_jobs"`
	CleanupInterval    time.Duration `json:"cleanup_interval"`
	RetentionPeriod    time.Duration `json:"retention_period"`
}

// WorkflowManager N8N 工作流管理器实现
type WorkflowManager struct {
	config              *WorkflowManagerConfig
	n8nClient           analysis.N8NClient
	executionRepo       analysis.N8NWorkflowExecutionRepository
	alertRepo           alert.AlertRepository
	logger              *zap.Logger
	monitors            map[string]context.CancelFunc
	monitorMu           sync.RWMutex
	workflowTemplates   map[string]string // analysisType -> workflowID
	templatesMu         sync.RWMutex
	running             bool
	shutdownCh          chan struct{}
	wg                  sync.WaitGroup
}

// NewWorkflowManager 创建新的工作流管理器
func NewWorkflowManager(
	config *WorkflowManagerConfig,
	n8nClient analysis.N8NClient,
	executionRepo analysis.N8NWorkflowExecutionRepository,
	alertRepo alert.AlertRepository,
	logger *zap.Logger,
) *WorkflowManager {
	return &WorkflowManager{
		config:            config,
		n8nClient:         n8nClient,
		executionRepo:     executionRepo,
		alertRepo:         alertRepo,
		logger:            logger,
		monitors:          make(map[string]context.CancelFunc),
		workflowTemplates: make(map[string]string),
		shutdownCh:        make(chan struct{}),
	}
}

// Start 启动工作流管理器
func (wm *WorkflowManager) Start(ctx context.Context) error {
	wm.running = true
	
	// 启动清理任务
	wm.wg.Add(1)
	go wm.cleanupWorker(ctx)
	
	// 启动监控任务
	wm.wg.Add(1)
	go wm.monitorWorker(ctx)
	
	wm.logger.Info("workflow manager started")
	return nil
}

// Stop 停止工作流管理器
func (wm *WorkflowManager) Stop() error {
	if !wm.running {
		return nil
	}
	
	wm.running = false
	close(wm.shutdownCh)
	
	// 取消所有监控
	wm.monitorMu.Lock()
	for executionID, cancel := range wm.monitors {
		cancel()
		wm.logger.Info("canceled execution monitor", zap.String("execution_id", executionID))
	}
	wm.monitors = make(map[string]context.CancelFunc)
	wm.monitorMu.Unlock()
	
	// 等待所有 goroutine 结束
	wm.wg.Wait()
	
	wm.logger.Info("workflow manager stopped")
	return nil
}

// RegisterWorkflowTemplate 注册工作流模板
func (wm *WorkflowManager) RegisterWorkflowTemplate(analysisType, workflowID string) {
	wm.templatesMu.Lock()
	defer wm.templatesMu.Unlock()
	
	wm.workflowTemplates[analysisType] = workflowID
	wm.logger.Info("workflow template registered",
		zap.String("analysis_type", analysisType),
		zap.String("workflow_id", workflowID))
}

// TriggerAnalysisWorkflow 触发分析工作流
func (wm *WorkflowManager) TriggerAnalysisWorkflow(ctx context.Context, alertID string, analysisType string, metadata map[string]interface{}) (*analysis.N8NWorkflowExecution, error) {
	// 获取工作流模板ID
	wm.templatesMu.RLock()
	workflowID, exists := wm.workflowTemplates[analysisType]
	wm.templatesMu.RUnlock()
	
	if !exists {
		return nil, fmt.Errorf("workflow template not found for analysis type: %s", analysisType)
	}
	
	// 转换 alertID 为 uint
	alertIDUint, err := strconv.ParseUint(alertID, 10, 32)
	if err != nil {
		return nil, fmt.Errorf("invalid alert ID: %w", err)
	}

	// 获取告警信息
	alertInfo, err := wm.alertRepo.GetByID(ctx, uint(alertIDUint))
	if err != nil {
		return nil, fmt.Errorf("get alert info: %w", err)
	}
	
	// 准备输入数据
	inputData := map[string]interface{}{
		"alert_id":      alertID,
		"analysis_type": analysisType,
		"alert_data":    alertInfo,
		"timestamp":     time.Now().Unix(),
	}
	
	// 合并元数据
	if metadata != nil {
		for k, v := range metadata {
			inputData[k] = v
		}
	}
	
	// 准备回调配置
	callbackConfig := &analysis.N8NCallbackConfig{
		URL:    fmt.Sprintf("%s/api/v1/n8n/callback", wm.getCallbackBaseURL()),
		Method: "POST",
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		Secret: wm.generateCallbackSecret(alertID),
	}
	
	// 触发工作流
	req := &analysis.N8NWorkflowTriggerRequest{
		WorkflowID: workflowID,
		InputData:  inputData,
		Metadata: map[string]interface{}{
			"alert_id":      alertID,
			"analysis_type": analysisType,
			"triggered_at":  time.Now(),
		},
		Callback: callbackConfig,
	}
	
	execution, err := wm.n8nClient.TriggerWorkflow(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("trigger workflow: %w", err)
	}
	
	// 保存执行记录
	if err := wm.executionRepo.Create(ctx, execution); err != nil {
		wm.logger.Error("failed to save execution record",
			zap.String("execution_id", execution.ID),
			zap.Error(err))
	}
	
	// 启动监控
	if err := wm.startMonitoring(ctx, execution.ID); err != nil {
		wm.logger.Error("failed to start monitoring",
			zap.String("execution_id", execution.ID),
			zap.Error(err))
	}
	
	wm.logger.Info("analysis workflow triggered",
		zap.String("alert_id", alertID),
		zap.String("analysis_type", analysisType),
		zap.String("workflow_id", workflowID),
		zap.String("execution_id", execution.ID))
	
	return execution, nil
}

// MonitorExecution 监控工作流执行
func (wm *WorkflowManager) MonitorExecution(ctx context.Context, executionID string) (<-chan *analysis.N8NWorkflowExecution, error) {
	ch := make(chan *analysis.N8NWorkflowExecution, 10)
	
	go func() {
		defer close(ch)
		
		ticker := time.NewTicker(wm.config.MonitorInterval)
		defer ticker.Stop()
		
		timeout := time.NewTimer(wm.config.ExecutionTimeout)
		defer timeout.Stop()
		
		for {
			select {
			case <-ctx.Done():
				return
			case <-timeout.C:
				wm.logger.Warn("execution monitoring timeout",
					zap.String("execution_id", executionID))
				return
			case <-ticker.C:
				execution, err := wm.n8nClient.GetWorkflowExecution(ctx, executionID)
				if err != nil {
					wm.logger.Error("failed to get execution status",
						zap.String("execution_id", executionID),
						zap.Error(err))
					continue
				}
				
				// 更新执行记录
				if err := wm.executionRepo.Update(ctx, execution); err != nil {
					wm.logger.Error("failed to update execution record",
						zap.String("execution_id", executionID),
						zap.Error(err))
				}
				
				select {
				case ch <- execution:
				case <-ctx.Done():
					return
				}
				
				// 如果执行完成，停止监控
				if execution.Status == analysis.N8NWorkflowStatusCompleted ||
					execution.Status == analysis.N8NWorkflowStatusFailed ||
					execution.Status == analysis.N8NWorkflowStatusCanceled {
					return
				}
			}
		}
	}()
	
	return ch, nil
}

// HandleCallback 处理工作流回调
func (wm *WorkflowManager) HandleCallback(ctx context.Context, executionID string, data map[string]interface{}) error {
	// 验证回调数据
	if err := wm.validateCallbackData(data); err != nil {
		return fmt.Errorf("invalid callback data: %w", err)
	}
	
	// 获取执行记录
	execution, err := wm.executionRepo.GetByID(ctx, executionID)
	if err != nil {
		return fmt.Errorf("get execution record: %w", err)
	}
	
	// 更新执行状态
	if status, ok := data["status"].(string); ok {
		execution.Status = analysis.N8NWorkflowStatus(status)
	}
	
	if outputData, ok := data["output"].(map[string]interface{}); ok {
		execution.OutputData = outputData
	}
	
	if errorData, ok := data["error"].(string); ok {
		execution.ErrorData = &errorData
	}
	
	if execution.Status == analysis.N8NWorkflowStatusCompleted ||
		execution.Status == analysis.N8NWorkflowStatusFailed ||
		execution.Status == analysis.N8NWorkflowStatusCanceled {
		now := time.Now()
		execution.FinishedAt = &now
	}
	
	// 保存更新
	if err := wm.executionRepo.Update(ctx, execution); err != nil {
		return fmt.Errorf("update execution record: %w", err)
	}
	
	// 处理分析结果
	if execution.Status == analysis.N8NWorkflowStatusCompleted {
		if err := wm.processAnalysisResult(ctx, execution); err != nil {
			wm.logger.Error("failed to process analysis result",
				zap.String("execution_id", executionID),
				zap.Error(err))
		}
	}
	
	// 停止监控
	wm.stopMonitoring(executionID)
	
	wm.logger.Info("workflow callback processed",
		zap.String("execution_id", executionID),
		zap.String("status", string(execution.Status)))
	
	return nil
}

// RetryFailedExecution 重试失败的执行
func (wm *WorkflowManager) RetryFailedExecution(ctx context.Context, executionID string) (*analysis.N8NWorkflowExecution, error) {
	// 获取原始执行记录
	originalExecution, err := wm.executionRepo.GetByID(ctx, executionID)
	if err != nil {
		return nil, fmt.Errorf("get original execution: %w", err)
	}
	
	if originalExecution.Status != analysis.N8NWorkflowStatusFailed {
		return nil, fmt.Errorf("execution is not in failed status: %s", originalExecution.Status)
	}
	
	// 重试工作流
	newExecution, err := wm.n8nClient.RetryWorkflowExecution(ctx, executionID)
	if err != nil {
		return nil, fmt.Errorf("retry workflow execution: %w", err)
	}
	
	// 保存新的执行记录
	if err := wm.executionRepo.Create(ctx, newExecution); err != nil {
		wm.logger.Error("failed to save retry execution record",
			zap.String("execution_id", newExecution.ID),
			zap.Error(err))
	}
	
	// 启动监控
	if err := wm.startMonitoring(ctx, newExecution.ID); err != nil {
		wm.logger.Error("failed to start monitoring for retry",
			zap.String("execution_id", newExecution.ID),
			zap.Error(err))
	}
	
	wm.logger.Info("workflow execution retried",
		zap.String("original_execution_id", executionID),
		zap.String("new_execution_id", newExecution.ID))
	
	return newExecution, nil
}

// GetExecutionLogs 获取执行日志
func (wm *WorkflowManager) GetExecutionLogs(ctx context.Context, executionID string) ([]string, error) {
	// 这里可以实现从 n8n 获取详细日志的逻辑
	// 目前返回基本信息
	execution, err := wm.executionRepo.GetByID(ctx, executionID)
	if err != nil {
		return nil, fmt.Errorf("get execution: %w", err)
	}
	
	logs := []string{
		fmt.Sprintf("Execution ID: %s", execution.ID),
		fmt.Sprintf("Workflow ID: %s", execution.WorkflowID),
		fmt.Sprintf("Status: %s", execution.Status),
		fmt.Sprintf("Started At: %s", execution.StartedAt.Format(time.RFC3339)),
	}
	
	if execution.FinishedAt != nil {
		logs = append(logs, fmt.Sprintf("Finished At: %s", execution.FinishedAt.Format(time.RFC3339)))
		duration := execution.FinishedAt.Sub(execution.StartedAt)
		logs = append(logs, fmt.Sprintf("Duration: %s", duration.String()))
	}
	
	if execution.ErrorData != nil {
		logs = append(logs, fmt.Sprintf("Error: %s", *execution.ErrorData))
	}
	
	return logs, nil
}

// GetWorkflowMetrics 获取工作流指标
func (wm *WorkflowManager) GetWorkflowMetrics(ctx context.Context, workflowID string, timeRange time.Duration) (map[string]interface{}, error) {
	endTime := time.Now()
	startTime := endTime.Add(-timeRange)
	
	return wm.executionRepo.GetStatistics(ctx, workflowID, startTime, endTime)
}

// 私有方法

// startMonitoring 启动执行监控
func (wm *WorkflowManager) startMonitoring(ctx context.Context, executionID string) error {
	wm.monitorMu.Lock()
	defer wm.monitorMu.Unlock()
	
	// 如果已经在监控，先取消
	if cancel, exists := wm.monitors[executionID]; exists {
		cancel()
	}
	
	monitorCtx, cancel := context.WithCancel(ctx)
	wm.monitors[executionID] = cancel
	
	go func() {
		defer func() {
			wm.monitorMu.Lock()
			delete(wm.monitors, executionID)
			wm.monitorMu.Unlock()
		}()
		
		ch, err := wm.MonitorExecution(monitorCtx, executionID)
		if err != nil {
			wm.logger.Error("failed to start execution monitoring",
				zap.String("execution_id", executionID),
				zap.Error(err))
			return
		}
		
		for execution := range ch {
			wm.logger.Debug("execution status update",
				zap.String("execution_id", execution.ID),
				zap.String("status", string(execution.Status)))
		}
	}()
	
	return nil
}

// stopMonitoring 停止执行监控
func (wm *WorkflowManager) stopMonitoring(executionID string) {
	wm.monitorMu.Lock()
	defer wm.monitorMu.Unlock()
	
	if cancel, exists := wm.monitors[executionID]; exists {
		cancel()
		delete(wm.monitors, executionID)
		wm.logger.Debug("execution monitoring stopped", zap.String("execution_id", executionID))
	}
}

// cleanupWorker 清理过期记录
func (wm *WorkflowManager) cleanupWorker(ctx context.Context) {
	defer wm.wg.Done()
	
	ticker := time.NewTicker(wm.config.CleanupInterval)
	defer ticker.Stop()
	
	for {
		select {
		case <-ctx.Done():
			return
		case <-wm.shutdownCh:
			return
		case <-ticker.C:
			wm.performCleanup(ctx)
		}
	}
}

// monitorWorker 监控工作器
func (wm *WorkflowManager) monitorWorker(ctx context.Context) {
	defer wm.wg.Done()
	
	ticker := time.NewTicker(wm.config.MonitorInterval)
	defer ticker.Stop()
	
	for {
		select {
		case <-ctx.Done():
			return
		case <-wm.shutdownCh:
			return
		case <-ticker.C:
			wm.checkRunningExecutions(ctx)
		}
	}
}

// performCleanup 执行清理
func (wm *WorkflowManager) performCleanup(ctx context.Context) {
	cutoffTime := time.Now().Add(-wm.config.RetentionPeriod)
	
	// 这里可以实现清理逻辑，比如删除过期的执行记录
	wm.logger.Debug("performing cleanup", zap.Time("cutoff_time", cutoffTime))
}

// checkRunningExecutions 检查运行中的执行
func (wm *WorkflowManager) checkRunningExecutions(ctx context.Context) {
	// 获取运行中的执行
	executions, err := wm.executionRepo.ListByStatus(ctx, analysis.N8NWorkflowStatusRunning, 100, 0)
	if err != nil {
		wm.logger.Error("failed to get running executions", zap.Error(err))
		return
	}
	
	for _, execution := range executions {
		// 检查是否超时
	if time.Since(execution.StartedAt) > wm.config.ExecutionTimeout {
			wm.logger.Warn("execution timeout detected",
				zap.String("execution_id", execution.ID),
				zap.Duration("duration", time.Since(execution.StartedAt)))
			
			// 尝试取消执行
			if err := wm.n8nClient.CancelWorkflowExecution(ctx, execution.ID); err != nil {
				wm.logger.Error("failed to cancel timeout execution",
					zap.String("execution_id", execution.ID),
					zap.Error(err))
			}
		}
	}
}

// validateCallbackData 验证回调数据
func (wm *WorkflowManager) validateCallbackData(data map[string]interface{}) error {
	if data == nil {
		return fmt.Errorf("callback data is nil")
	}
	
	if _, ok := data["status"]; !ok {
		return fmt.Errorf("missing status field")
	}
	
	return nil
}

// processAnalysisResult 处理分析结果
func (wm *WorkflowManager) processAnalysisResult(ctx context.Context, execution *analysis.N8NWorkflowExecution) error {
	// 从元数据中获取告警ID
	alertID, ok := execution.Metadata["alert_id"].(string)
	if !ok {
		return fmt.Errorf("alert_id not found in metadata")
	}
	
	// 这里可以实现将分析结果关联到告警记录的逻辑
	wm.logger.Info("processing analysis result",
		zap.String("alert_id", alertID),
		zap.String("execution_id", execution.ID))
	
	return nil
}

// getCallbackBaseURL 获取回调基础URL
func (wm *WorkflowManager) getCallbackBaseURL() string {
	// 这里应该从配置中获取
	return "http://localhost:8080"
}

// generateCallbackSecret 生成回调密钥
func (wm *WorkflowManager) generateCallbackSecret(alertID string) string {
	// 这里应该实现更安全的密钥生成逻辑
	return fmt.Sprintf("secret_%s_%d", alertID, time.Now().Unix())
}