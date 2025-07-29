package analysis

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"alert_agent/internal/domain/analysis"
	"alert_agent/internal/domain/alert"
	"go.uber.org/zap"
	"github.com/google/uuid"
)

// DifyAnalysisServiceImpl Dify 分析服务实现
type DifyAnalysisServiceImpl struct {
	difyClient         analysis.DifyClient
	analysisRepository analysis.DifyAnalysisRepository
	alertRepository    alert.AlertRepository
	logger             *zap.Logger
	
	// 任务管理
	tasks    map[string]*analysis.DifyAnalysisTask
	tasksMux sync.RWMutex
	
	// 配置
	config *DifyAnalysisConfig
}

// DifyAnalysisConfig Dify 分析配置
type DifyAnalysisConfig struct {
	// 默认超时时间
	DefaultTimeout time.Duration
	
	// 最大重试次数
	MaxRetries int
	
	// 重试间隔
	RetryInterval time.Duration
	
	// 默认 Agent ID
	DefaultAgentID string
	
	// 默认工作流 ID
	DefaultWorkflowID string
	
	// 并发限制
	ConcurrencyLimit int
	
	// 任务清理间隔
	TaskCleanupInterval time.Duration
	
	// 任务保留时间
	TaskRetentionTime time.Duration
}

// NewDifyAnalysisService 创建新的 Dify 分析服务
func NewDifyAnalysisService(
	difyClient analysis.DifyClient,
	analysisRepository analysis.DifyAnalysisRepository,
	alertRepository alert.AlertRepository,
	logger *zap.Logger,
	config *DifyAnalysisConfig,
) analysis.DifyAnalysisService {
	if config == nil {
		config = &DifyAnalysisConfig{
			DefaultTimeout:      5 * time.Minute,
			MaxRetries:          3,
			RetryInterval:       30 * time.Second,
			ConcurrencyLimit:    10,
			TaskCleanupInterval: 1 * time.Hour,
			TaskRetentionTime:   24 * time.Hour,
		}
	}
	
	service := &DifyAnalysisServiceImpl{
		difyClient:         difyClient,
		analysisRepository: analysisRepository,
		alertRepository:    alertRepository,
		logger:             logger,
		tasks:              make(map[string]*analysis.DifyAnalysisTask),
		config:             config,
	}
	
	// 启动任务清理协程
	go service.startTaskCleanup()
	
	return service
}

// AnalyzeAlert 分析告警（异步模式）
func (s *DifyAnalysisServiceImpl) AnalyzeAlert(ctx context.Context, request *analysis.DifyAnalysisRequest) (*analysis.DifyAnalysisTask, error) {
	// 生成任务ID
	taskID := uuid.New().String()
	
	// 创建分析任务
	task := &analysis.DifyAnalysisTask{
		ID:           taskID,
		AlertID:      request.AlertID,
		AnalysisType: request.AnalysisType,
		Status:       analysis.DifyAnalysisStatusPending,
		Progress:     0,
		CreatedAt:    time.Now(),
		RetryCount:   0,
		MaxRetries:   s.config.MaxRetries,
	}
	
	// 存储任务
	s.tasksMux.Lock()
	s.tasks[taskID] = task
	s.tasksMux.Unlock()
	
	// 异步执行分析
	go s.executeAnalysis(context.Background(), task, request)
	
	s.logger.Info("Dify analysis task created",
		zap.String("task_id", taskID),
		zap.Uint("alert_id", request.AlertID),
		zap.String("analysis_type", request.AnalysisType),
	)
	
	return task, nil
}

// GetAnalysisResult 获取分析结果
func (s *DifyAnalysisServiceImpl) GetAnalysisResult(ctx context.Context, taskID string) (*analysis.DifyAnalysisResult, error) {
	// 先从内存中查找
	s.tasksMux.RLock()
	task, exists := s.tasks[taskID]
	s.tasksMux.RUnlock()
	
	if !exists {
		// 从数据库中查找
		return s.analysisRepository.GetAnalysisResult(ctx, taskID)
	}
	
	if task.Status != analysis.DifyAnalysisStatusCompleted {
		return nil, fmt.Errorf("analysis not completed, current status: %s", task.Status)
	}
	
	// 从数据库获取完整结果
	return s.analysisRepository.GetAnalysisResult(ctx, taskID)
}

// GetAnalysisProgress 获取分析进度
func (s *DifyAnalysisServiceImpl) GetAnalysisProgress(ctx context.Context, taskID string) (*analysis.DifyAnalysisProgress, error) {
	s.tasksMux.RLock()
	task, exists := s.tasks[taskID]
	s.tasksMux.RUnlock()
	
	if !exists {
		return nil, fmt.Errorf("task not found: %s", taskID)
	}
	
	progress := &analysis.DifyAnalysisProgress{
		TaskID:                 taskID,
		Status:                 task.Status,
		Progress:               task.Progress,
		CurrentStep:            s.getCurrentStep(task),
		TotalSteps:             s.getTotalSteps(task.AnalysisType),
		CompletedSteps:         s.getCompletedSteps(task),
		EstimatedRemainingTime: s.getEstimatedRemainingTime(task),
		UpdatedAt:              time.Now(),
	}
	
	return progress, nil
}

// CancelAnalysis 取消分析任务
func (s *DifyAnalysisServiceImpl) CancelAnalysis(ctx context.Context, taskID string) error {
	s.tasksMux.Lock()
	task, exists := s.tasks[taskID]
	if exists {
		task.Status = analysis.DifyAnalysisStatusCancelled
		
		// 如果有工作流运行ID，尝试取消工作流
		if task.WorkflowRunID != "" {
			go func() {
				if err := s.difyClient.CancelWorkflow(context.Background(), task.WorkflowRunID); err != nil {
					s.logger.Error("Failed to cancel workflow",
						zap.String("task_id", taskID),
						zap.String("workflow_run_id", task.WorkflowRunID),
						zap.Error(err),
					)
				}
			}()
		}
	}
	s.tasksMux.Unlock()
	
	if !exists {
		return fmt.Errorf("task not found: %s", taskID)
	}
	
	s.logger.Info("Dify analysis task cancelled",
		zap.String("task_id", taskID),
	)
	
	return nil
}

// RetryAnalysis 重试分析任务
func (s *DifyAnalysisServiceImpl) RetryAnalysis(ctx context.Context, taskID string) error {
	s.tasksMux.Lock()
	task, exists := s.tasks[taskID]
	if exists && task.Status == analysis.DifyAnalysisStatusFailed {
		if task.RetryCount >= task.MaxRetries {
			s.tasksMux.Unlock()
			return fmt.Errorf("maximum retry attempts exceeded")
		}
		
		task.Status = analysis.DifyAnalysisStatusPending
		task.Progress = 0
		task.Error = ""
		task.RetryCount++
	}
	s.tasksMux.Unlock()
	
	if !exists {
		return fmt.Errorf("task not found: %s", taskID)
	}
	
	// 重新执行分析
	// 这里需要重新构建请求，简化处理
	request := &analysis.DifyAnalysisRequest{
		AlertID:      task.AlertID,
		AnalysisType: task.AnalysisType,
		UserID:       "system", // 重试时使用系统用户
	}
	
	go s.executeAnalysis(context.Background(), task, request)
	
	s.logger.Info("Dify analysis task retried",
		zap.String("task_id", taskID),
		zap.Int("retry_count", task.RetryCount),
	)
	
	return nil
}

// GetAnalysisHistory 获取分析历史
func (s *DifyAnalysisServiceImpl) GetAnalysisHistory(ctx context.Context, filter *analysis.DifyAnalysisFilter) ([]*analysis.DifyAnalysisResult, error) {
	return s.analysisRepository.GetAnalysisHistory(ctx, filter)
}

// GetAnalysisTrends 获取分析趋势
func (s *DifyAnalysisServiceImpl) GetAnalysisTrends(ctx context.Context, request *analysis.DifyTrendRequest) (*analysis.DifyTrendResponse, error) {
	return s.analysisRepository.GetAnalysisTrends(ctx, request)
}

// SearchKnowledge 搜索知识库
func (s *DifyAnalysisServiceImpl) SearchKnowledge(ctx context.Context, query string, options *analysis.KnowledgeSearchOptions) (*analysis.KnowledgeSearchResult, error) {
	// 设置默认选项
	if options == nil {
		options = &analysis.KnowledgeSearchOptions{
			DatasetIDs: []string{"default"},
			Limit:      10,
			SimilarityThreshold: 0.7,
		}
	}
	
	return s.difyClient.SearchKnowledge(ctx, query, options)
}

// BuildAlertContext 构建告警上下文
func (s *DifyAnalysisServiceImpl) BuildAlertContext(ctx context.Context, alertID uint, options *analysis.ContextBuildOptions) (*analysis.AlertContext, error) {
	// 获取告警基本信息
	alertInfo, err := s.alertRepository.GetByID(ctx, alertID)
	if err != nil {
		return nil, fmt.Errorf("failed to get alert: %w", err)
	}
	
	context := &analysis.AlertContext{
		Alert: s.convertAlertToMap(alertInfo),
	}
	
	if options == nil {
		return context, nil
	}
	
	// 构建历史告警上下文
	if options.IncludeHistory {
		historyDays := options.HistoryDays
		if historyDays <= 0 {
			historyDays = 7 // 默认7天
		}
		
		// 这里应该调用实际的历史告警查询
		// historicalAlerts, _ := s.alertRepository.GetHistoricalAlerts(ctx, alertID, historyDays)
		// context.HistoricalAlerts = s.convertAlertsToMaps(historicalAlerts)
	}
	
	// 构建相关告警上下文
	if options.IncludeRelated {
		// 这里应该调用实际的相关告警查询
		// relatedAlerts, _ := s.alertRepository.GetRelatedAlerts(ctx, alertID)
		// context.RelatedAlerts = s.convertAlertsToMaps(relatedAlerts)
	}
	
	// 构建指标上下文
	if options.IncludeMetrics {
		// 这里应该调用实际的指标查询
		// metrics, _ := s.metricsRepository.GetMetrics(ctx, alertID, options.MetricsTimeRange)
		// context.Metrics = metrics
	}
	
	// 构建日志上下文
	if options.IncludeLogs {
		// 这里应该调用实际的日志查询
		// logs, _ := s.logsRepository.GetLogs(ctx, alertID, options.LogsTimeRange)
		// context.Logs = logs
	}
	
	return context, nil
}

// UpdateAlertWithAnalysis 将分析结果回写到告警记录
func (s *DifyAnalysisServiceImpl) UpdateAlertWithAnalysis(ctx context.Context, alertID uint, analysisResult *analysis.DifyAnalysisResult) error {
	// 构建更新数据
	updateData := map[string]interface{}{
		"analysis_result": analysisResult.Result,
		"analysis_confidence": analysisResult.Confidence,
		"analysis_completed_at": time.Now(),
	}
	
	if analysisResult.RootCause != "" {
		updateData["root_cause"] = analysisResult.RootCause
	}
	
	if analysisResult.Impact != "" {
		updateData["impact_assessment"] = analysisResult.Impact
	}
	
	if len(analysisResult.Recommendations) > 0 {
		recommendationsJSON, _ := json.Marshal(analysisResult.Recommendations)
		updateData["recommendations"] = string(recommendationsJSON)
	}
	
	// 更新告警记录
	return s.alertRepository.UpdateByID(ctx, alertID, updateData)
}

// GetAnalysisMetrics 获取分析指标
func (s *DifyAnalysisServiceImpl) GetAnalysisMetrics(ctx context.Context, timeRange *analysis.TimeRange) (*analysis.DifyAnalysisMetrics, error) {
	return s.analysisRepository.GetAnalysisMetrics(ctx, timeRange)
}

// HealthCheck 健康检查
func (s *DifyAnalysisServiceImpl) HealthCheck(ctx context.Context) error {
	return s.difyClient.HealthCheck(ctx)
}

// executeAnalysis 执行分析任务
func (s *DifyAnalysisServiceImpl) executeAnalysis(ctx context.Context, task *analysis.DifyAnalysisTask, request *analysis.DifyAnalysisRequest) {
	defer func() {
		if r := recover(); r != nil {
			s.logger.Error("Analysis task panicked",
				zap.String("task_id", task.ID),
				zap.Any("panic", r),
			)
			
			s.tasksMux.Lock()
			task.Status = analysis.DifyAnalysisStatusFailed
			task.Error = fmt.Sprintf("panic: %v", r)
			s.tasksMux.Unlock()
		}
	}()
	
	// 更新任务状态
	s.updateTaskStatus(task, analysis.DifyAnalysisStatusRunning, 10)
	
	// 构建告警上下文
	contextOptions := &analysis.ContextBuildOptions{
		IncludeHistory:        true,
		HistoryDays:          7,
		IncludeRelated:       true,
		IncludeMetrics:       true,
		MetricsTimeRange:     60,
		IncludeLogs:          true,
		LogsTimeRange:        30,
		IncludeServices:      true,
		IncludeInfrastructure: true,
	}
	
	if request.Options != nil {
		if request.Options.IncludeHistory {
			contextOptions.IncludeHistory = request.Options.IncludeHistory
		}
		if request.Options.HistoryDays > 0 {
			contextOptions.HistoryDays = request.Options.HistoryDays
		}
		if request.Options.IncludeRelatedAlerts {
			contextOptions.IncludeRelated = request.Options.IncludeRelatedAlerts
		}
		if request.Options.IncludeMetrics {
			contextOptions.IncludeMetrics = request.Options.IncludeMetrics
		}
		if request.Options.IncludeLogs {
			contextOptions.IncludeLogs = request.Options.IncludeLogs
		}
	}
	
	alertContext, err := s.BuildAlertContext(ctx, request.AlertID, contextOptions)
	if err != nil {
		s.handleAnalysisError(task, fmt.Errorf("failed to build alert context: %w", err))
		return
	}
	
	s.updateTaskStatus(task, analysis.DifyAnalysisStatusRunning, 30)
	
	// 根据分析类型选择处理方式
	var result *analysis.DifyAnalysisResult
	switch request.AnalysisType {
	case analysis.DifyAnalysisTypeRootCause:
		result, err = s.performRootCauseAnalysis(ctx, task, request, alertContext)
	case analysis.DifyAnalysisTypeImpact:
		result, err = s.performImpactAnalysis(ctx, task, request, alertContext)
	case analysis.DifyAnalysisTypeRecommendation:
		result, err = s.performRecommendationAnalysis(ctx, task, request, alertContext)
	case analysis.DifyAnalysisTypeTrend:
		result, err = s.performTrendAnalysis(ctx, task, request, alertContext)
	case analysis.DifyAnalysisTypeClassification:
		result, err = s.performClassificationAnalysis(ctx, task, request, alertContext)
	default:
		err = fmt.Errorf("unsupported analysis type: %s", request.AnalysisType)
	}
	
	if err != nil {
		s.handleAnalysisError(task, err)
		return
	}
	
	s.updateTaskStatus(task, analysis.DifyAnalysisStatusRunning, 90)
	
	// 保存分析结果
	if err := s.analysisRepository.SaveAnalysisResult(ctx, result); err != nil {
		s.handleAnalysisError(task, fmt.Errorf("failed to save analysis result: %w", err))
		return
	}
	
	// 回写分析结果到告警记录
	if err := s.UpdateAlertWithAnalysis(ctx, request.AlertID, result); err != nil {
		s.logger.Error("Failed to update alert with analysis result",
			zap.String("task_id", task.ID),
			zap.Uint("alert_id", request.AlertID),
			zap.Error(err),
		)
		// 不因为回写失败而标记整个任务失败
	}
	
	// 标记任务完成
	s.updateTaskStatus(task, analysis.DifyAnalysisStatusCompleted, 100)
	task.CompletedAt = &[]time.Time{time.Now()}[0]
	
	s.logger.Info("Dify analysis task completed",
		zap.String("task_id", task.ID),
		zap.Uint("alert_id", request.AlertID),
		zap.String("analysis_type", request.AnalysisType),
		zap.Duration("duration", time.Since(task.CreatedAt)),
	)
}

// performRootCauseAnalysis 执行根因分析
func (s *DifyAnalysisServiceImpl) performRootCauseAnalysis(ctx context.Context, task *analysis.DifyAnalysisTask, request *analysis.DifyAnalysisRequest, alertContext *analysis.AlertContext) (*analysis.DifyAnalysisResult, error) {
	// 构建分析提示
	prompt := s.buildRootCausePrompt(alertContext)
	
	// 调用 Dify API
	chatRequest := &analysis.DifyChatRequest{
		Inputs: map[string]interface{}{
			"alert_data": alertContext.Alert,
			"context": alertContext,
			"analysis_type": "root_cause",
		},
		Query:        prompt,
		ResponseMode: analysis.DifyResponseModeBlocking,
		User:         request.UserID,
	}
	
	chatResponse, err := s.difyClient.ChatMessage(ctx, chatRequest)
	if err != nil {
		return nil, fmt.Errorf("chat message failed: %w", err)
	}
	
	// 解析分析结果
	result := &analysis.DifyAnalysisResult{
		TaskID:        task.ID,
		AlertID:       request.AlertID,
		AnalysisType:  request.AnalysisType,
		Result:        chatResponse.Answer,
		RootCause:     s.extractRootCause(chatResponse.Answer),
		Confidence:    s.calculateConfidence(chatResponse),
		TokenUsage:    chatResponse.Metadata.Usage.TotalTokens,
		Cost:          s.calculateCost(chatResponse.Metadata.Usage),
		CreatedAt:     time.Now(),
		ConversationID: chatResponse.ConversationID,
		MessageID:     chatResponse.MessageID,
	}
	
	return result, nil
}

// performImpactAnalysis 执行影响分析
func (s *DifyAnalysisServiceImpl) performImpactAnalysis(ctx context.Context, task *analysis.DifyAnalysisTask, request *analysis.DifyAnalysisRequest, alertContext *analysis.AlertContext) (*analysis.DifyAnalysisResult, error) {
	// 类似根因分析的实现
	prompt := s.buildImpactPrompt(alertContext)
	
	chatRequest := &analysis.DifyChatRequest{
		Inputs: map[string]interface{}{
			"alert_data": alertContext.Alert,
			"context": alertContext,
			"analysis_type": "impact",
		},
		Query:        prompt,
		ResponseMode: analysis.DifyResponseModeBlocking,
		User:         request.UserID,
	}
	
	chatResponse, err := s.difyClient.ChatMessage(ctx, chatRequest)
	if err != nil {
		return nil, fmt.Errorf("chat message failed: %w", err)
	}
	
	result := &analysis.DifyAnalysisResult{
		TaskID:        task.ID,
		AlertID:       request.AlertID,
		AnalysisType:  request.AnalysisType,
		Result:        chatResponse.Answer,
		Impact:        s.extractImpact(chatResponse.Answer),
		Confidence:    s.calculateConfidence(chatResponse),
		TokenUsage:    chatResponse.Metadata.Usage.TotalTokens,
		Cost:          s.calculateCost(chatResponse.Metadata.Usage),
		CreatedAt:     time.Now(),
		ConversationID: chatResponse.ConversationID,
		MessageID:     chatResponse.MessageID,
	}
	
	return result, nil
}

// performRecommendationAnalysis 执行建议分析
func (s *DifyAnalysisServiceImpl) performRecommendationAnalysis(ctx context.Context, task *analysis.DifyAnalysisTask, request *analysis.DifyAnalysisRequest, alertContext *analysis.AlertContext) (*analysis.DifyAnalysisResult, error) {
	prompt := s.buildRecommendationPrompt(alertContext)
	
	chatRequest := &analysis.DifyChatRequest{
		Inputs: map[string]interface{}{
			"alert_data": alertContext.Alert,
			"context": alertContext,
			"analysis_type": "recommendation",
		},
		Query:        prompt,
		ResponseMode: analysis.DifyResponseModeBlocking,
		User:         request.UserID,
	}
	
	chatResponse, err := s.difyClient.ChatMessage(ctx, chatRequest)
	if err != nil {
		return nil, fmt.Errorf("chat message failed: %w", err)
	}
	
	result := &analysis.DifyAnalysisResult{
		TaskID:         task.ID,
		AlertID:        request.AlertID,
		AnalysisType:   request.AnalysisType,
		Result:         chatResponse.Answer,
		Recommendations: s.extractRecommendations(chatResponse.Answer),
		Confidence:     s.calculateConfidence(chatResponse),
		TokenUsage:     chatResponse.Metadata.Usage.TotalTokens,
		Cost:           s.calculateCost(chatResponse.Metadata.Usage),
		CreatedAt:      time.Now(),
		ConversationID: chatResponse.ConversationID,
		MessageID:      chatResponse.MessageID,
	}
	
	return result, nil
}

// performTrendAnalysis 执行趋势分析
func (s *DifyAnalysisServiceImpl) performTrendAnalysis(ctx context.Context, task *analysis.DifyAnalysisTask, request *analysis.DifyAnalysisRequest, alertContext *analysis.AlertContext) (*analysis.DifyAnalysisResult, error) {
	// 趋势分析可能需要使用工作流
	workflowRequest := &analysis.DifyWorkflowRequest{
		Inputs: map[string]interface{}{
			"alert_data": alertContext.Alert,
			"historical_data": alertContext.HistoricalAlerts,
			"metrics": alertContext.Metrics,
		},
		ResponseMode: analysis.DifyResponseModeBlocking,
		User:         request.UserID,
	}
	
	workflowResponse, err := s.difyClient.RunWorkflow(ctx, workflowRequest)
	if err != nil {
		return nil, fmt.Errorf("workflow run failed: %w", err)
	}
	
	// 更新任务的工作流运行ID
	task.WorkflowRunID = workflowResponse.WorkflowRunID
	
	result := &analysis.DifyAnalysisResult{
		TaskID:         task.ID,
		AlertID:        request.AlertID,
		AnalysisType:   request.AnalysisType,
		Result:         s.formatWorkflowResult(workflowResponse),
		Confidence:     0.8, // 工作流分析的默认置信度
		TokenUsage:     0,    // 工作流可能不返回token使用量
		Cost:           0,    // 工作流可能不返回成本
		CreatedAt:      time.Now(),
		WorkflowRunID:  workflowResponse.WorkflowRunID,
	}
	
	return result, nil
}

// performClassificationAnalysis 执行分类分析
func (s *DifyAnalysisServiceImpl) performClassificationAnalysis(ctx context.Context, task *analysis.DifyAnalysisTask, request *analysis.DifyAnalysisRequest, alertContext *analysis.AlertContext) (*analysis.DifyAnalysisResult, error) {
	prompt := s.buildClassificationPrompt(alertContext)
	
	chatRequest := &analysis.DifyChatRequest{
		Inputs: map[string]interface{}{
			"alert_data": alertContext.Alert,
			"context": alertContext,
			"analysis_type": "classification",
		},
		Query:        prompt,
		ResponseMode: analysis.DifyResponseModeBlocking,
		User:         request.UserID,
	}
	
	chatResponse, err := s.difyClient.ChatMessage(ctx, chatRequest)
	if err != nil {
		return nil, fmt.Errorf("chat message failed: %w", err)
	}
	
	result := &analysis.DifyAnalysisResult{
		TaskID:         task.ID,
		AlertID:        request.AlertID,
		AnalysisType:   request.AnalysisType,
		Result:         chatResponse.Answer,
		Classification: s.extractClassification(chatResponse.Answer),
		Confidence:     s.calculateConfidence(chatResponse),
		TokenUsage:     chatResponse.Metadata.Usage.TotalTokens,
		Cost:           s.calculateCost(chatResponse.Metadata.Usage),
		CreatedAt:      time.Now(),
		ConversationID: chatResponse.ConversationID,
		MessageID:      chatResponse.MessageID,
	}
	
	return result, nil
}

// 辅助方法

func (s *DifyAnalysisServiceImpl) updateTaskStatus(task *analysis.DifyAnalysisTask, status string, progress int) {
	s.tasksMux.Lock()
	task.Status = status
	task.Progress = progress
	if status == analysis.DifyAnalysisStatusRunning && task.StartedAt == nil {
		task.StartedAt = &[]time.Time{time.Now()}[0]
	}
	s.tasksMux.Unlock()
}

func (s *DifyAnalysisServiceImpl) handleAnalysisError(task *analysis.DifyAnalysisTask, err error) {
	s.tasksMux.Lock()
	task.Status = analysis.DifyAnalysisStatusFailed
	task.Error = err.Error()
	s.tasksMux.Unlock()
	
	s.logger.Error("Dify analysis task failed",
		zap.String("task_id", task.ID),
		zap.Uint("alert_id", task.AlertID),
		zap.String("analysis_type", task.AnalysisType),
		zap.Error(err),
	)
}

func (s *DifyAnalysisServiceImpl) getCurrentStep(task *analysis.DifyAnalysisTask) string {
	switch task.Status {
	case analysis.DifyAnalysisStatusPending:
		return "等待开始"
	case analysis.DifyAnalysisStatusRunning:
		if task.Progress < 30 {
			return "构建上下文"
		} else if task.Progress < 90 {
			return "AI分析中"
		} else {
			return "保存结果"
		}
	case analysis.DifyAnalysisStatusCompleted:
		return "已完成"
	case analysis.DifyAnalysisStatusFailed:
		return "分析失败"
	case analysis.DifyAnalysisStatusCancelled:
		return "已取消"
	default:
		return "未知状态"
	}
}

func (s *DifyAnalysisServiceImpl) getTotalSteps(analysisType string) int {
	return 3 // 构建上下文、AI分析、保存结果
}

func (s *DifyAnalysisServiceImpl) getCompletedSteps(task *analysis.DifyAnalysisTask) int {
	if task.Progress >= 90 {
		return 3
	} else if task.Progress >= 30 {
		return 2
	} else if task.Progress >= 10 {
		return 1
	}
	return 0
}

func (s *DifyAnalysisServiceImpl) getEstimatedRemainingTime(task *analysis.DifyAnalysisTask) int {
	if task.Status == analysis.DifyAnalysisStatusCompleted ||
		task.Status == analysis.DifyAnalysisStatusFailed ||
		task.Status == analysis.DifyAnalysisStatusCancelled {
		return 0
	}
	
	elapsed := time.Since(task.CreatedAt).Seconds()
	if task.Progress <= 0 {
		return int(s.config.DefaultTimeout.Seconds())
	}
	
	totalEstimated := elapsed * 100 / float64(task.Progress)
	remaining := totalEstimated - elapsed
	if remaining < 0 {
		remaining = 0
	}
	
	return int(remaining)
}

func (s *DifyAnalysisServiceImpl) convertAlertToMap(alertInfo interface{}) map[string]interface{} {
	// 这里应该根据实际的告警结构进行转换
	// 简化处理，直接返回空map
	return map[string]interface{}{
		"id": "placeholder",
		"message": "placeholder alert",
	}
}

func (s *DifyAnalysisServiceImpl) buildRootCausePrompt(context *analysis.AlertContext) string {
	return "请分析这个告警的根本原因，并提供详细的分析过程和结论。"
}

func (s *DifyAnalysisServiceImpl) buildImpactPrompt(context *analysis.AlertContext) string {
	return "请分析这个告警的影响范围和严重程度，包括对业务和系统的潜在影响。"
}

func (s *DifyAnalysisServiceImpl) buildRecommendationPrompt(context *analysis.AlertContext) string {
	return "请基于告警信息和上下文，提供具体的解决建议和预防措施。"
}

func (s *DifyAnalysisServiceImpl) buildClassificationPrompt(context *analysis.AlertContext) string {
	return "请对这个告警进行分类，包括告警类型、严重级别、影响范围等。"
}

func (s *DifyAnalysisServiceImpl) extractRootCause(answer string) string {
	// 这里应该实现从AI回答中提取根因的逻辑
	// 简化处理
	return "根因分析结果"
}

func (s *DifyAnalysisServiceImpl) extractImpact(answer string) string {
	return "影响分析结果"
}

func (s *DifyAnalysisServiceImpl) extractRecommendations(answer string) []string {
	return []string{"建议1", "建议2"}
}

func (s *DifyAnalysisServiceImpl) extractClassification(answer string) string {
	return "分类结果"
}

func (s *DifyAnalysisServiceImpl) calculateConfidence(response *analysis.DifyChatResponse) float64 {
	// 这里应该根据AI回答的质量计算置信度
	// 简化处理
	return 0.85
}

func (s *DifyAnalysisServiceImpl) calculateCost(usage *analysis.DifyUsage) float64 {
	// 这里应该根据token使用量计算成本
	// 简化处理
	return float64(usage.TotalTokens) * 0.001
}

func (s *DifyAnalysisServiceImpl) formatWorkflowResult(response *analysis.DifyWorkflowResponse) string {
	// 格式化工作流结果
	resultJSON, _ := json.Marshal(response.Data)
	return string(resultJSON)
}

func (s *DifyAnalysisServiceImpl) startTaskCleanup() {
	ticker := time.NewTicker(s.config.TaskCleanupInterval)
	defer ticker.Stop()
	
	for range ticker.C {
		s.cleanupOldTasks()
	}
}

func (s *DifyAnalysisServiceImpl) cleanupOldTasks() {
	s.tasksMux.Lock()
	defer s.tasksMux.Unlock()
	
	now := time.Now()
	for taskID, task := range s.tasks {
		if now.Sub(task.CreatedAt) > s.config.TaskRetentionTime {
			delete(s.tasks, taskID)
			s.logger.Debug("Cleaned up old task",
				zap.String("task_id", taskID),
				zap.Duration("age", now.Sub(task.CreatedAt)),
			)
		}
	}
}