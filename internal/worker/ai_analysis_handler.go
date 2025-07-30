package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"alert_agent/internal/model"
	"alert_agent/internal/pkg/database"
	"alert_agent/internal/pkg/logger"
	"alert_agent/internal/pkg/queue"
	"alert_agent/internal/service"

	"go.uber.org/zap"
)

// AIAnalysisHandler AI分析任务处理器
type AIAnalysisHandler struct {
	difyService   *service.DifyService
	ollamaService *service.OllamaService
	useDify       bool
}

// AIAnalysisResult AI分析结果
type AIAnalysisResult struct {
	AlertID        uint                   `json:"alert_id"`
	AnalysisType   string                 `json:"analysis_type"`
	Result         string                 `json:"result"`
	ActionPlan     string                 `json:"action_plan,omitempty"`
	SimilarAlerts  []string               `json:"similar_alerts,omitempty"`
	Confidence     float64                `json:"confidence"`
	ProcessingTime time.Duration          `json:"processing_time"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt      time.Time              `json:"created_at"`
}

// NewAIAnalysisHandler 创建AI分析处理器
func NewAIAnalysisHandler(difyService *service.DifyService, ollamaService *service.OllamaService) *AIAnalysisHandler {
	return &AIAnalysisHandler{
		difyService:   difyService,
		ollamaService: ollamaService,
		useDify:       true, // 默认优先使用Dify
	}
}

// Type 返回处理器类型
func (h *AIAnalysisHandler) Type() queue.TaskType {
	return queue.TaskTypeAIAnalysis
}

// Handle 处理AI分析任务
func (h *AIAnalysisHandler) Handle(ctx context.Context, task *queue.Task) error {
	startTime := time.Now()
	
	// 解析任务载荷
	alertID, ok := task.Payload["alert_id"].(string)
	if !ok {
		return fmt.Errorf("invalid alert_id in task payload")
	}

	analysisType, _ := task.Payload["analysis_type"].(string)
	if analysisType == "" {
		analysisType = "comprehensive"
	}

	logger.L.Info("Processing AI analysis task",
		zap.String("task_id", task.ID),
		zap.String("alert_id", alertID),
		zap.String("analysis_type", analysisType),
	)

	// 获取告警信息
	var alert model.Alert
	if err := database.DB.Where("id = ?", alertID).First(&alert).Error; err != nil {
		return fmt.Errorf("failed to get alert %s: %w", alertID, err)
	}

	// 更新告警状态为分析中
	if err := h.updateAlertStatus(ctx, &alert, "analyzing"); err != nil {
		logger.L.Warn("Failed to update alert status to analyzing", zap.Error(err))
	}

	// 执行AI分析
	analysisResult, err := h.performAnalysis(ctx, &alert, analysisType)
	if err != nil {
		// 更新状态为分析失败
		h.updateAlertStatus(ctx, &alert, "analysis_failed")
		return fmt.Errorf("AI analysis failed for alert %s: %w", alertID, err)
	}

	// 保存分析结果
	analysisResult.ProcessingTime = time.Since(startTime)
	if err := h.saveAnalysisResult(ctx, &alert, analysisResult); err != nil {
		return fmt.Errorf("failed to save analysis result: %w", err)
	}

	// 更新告警状态为已分析
	if err := h.updateAlertStatus(ctx, &alert, "analyzed"); err != nil {
		logger.L.Warn("Failed to update alert status to analyzed", zap.Error(err))
	}

	logger.L.Info("AI analysis completed successfully",
		zap.String("task_id", task.ID),
		zap.String("alert_id", alertID),
		zap.Duration("processing_time", analysisResult.ProcessingTime),
		zap.Float64("confidence", analysisResult.Confidence),
	)

	return nil
}

// performAnalysis 执行AI分析
func (h *AIAnalysisHandler) performAnalysis(ctx context.Context, alert *model.Alert, analysisType string) (*AIAnalysisResult, error) {
	result := &AIAnalysisResult{
		AlertID:      alert.ID,
		AnalysisType: analysisType,
		CreatedAt:    time.Now(),
		Metadata:     make(map[string]interface{}),
	}

	switch analysisType {
	case "comprehensive":
		return h.performComprehensiveAnalysis(ctx, alert, result)
	case "quick":
		return h.performQuickAnalysis(ctx, alert, result)
	case "root_cause":
		return h.performRootCauseAnalysis(ctx, alert, result)
	case "similar_search":
		return h.performSimilarSearch(ctx, alert, result)
	default:
		return h.performComprehensiveAnalysis(ctx, alert, result)
	}
}

// performComprehensiveAnalysis 执行综合分析
func (h *AIAnalysisHandler) performComprehensiveAnalysis(ctx context.Context, alert *model.Alert, result *AIAnalysisResult) (*AIAnalysisResult, error) {
	var analysisText string
	var err error

	// 尝试使用Dify进行分析
	if h.useDify && h.difyService != nil {
		logger.L.Debug("Using Dify for comprehensive analysis")
		
		// 使用Dify工作流进行综合分析
		workflowResult, difyErr := h.difyService.AnalyzeAlertWithWorkflow(ctx, alert)
		if difyErr != nil {
			logger.L.Warn("Dify workflow analysis failed, falling back to chat API", zap.Error(difyErr))
			
			// 回退到Dify Chat API
			analysisText, err = h.difyService.AnalyzeAlert(ctx, alert)
			if err != nil {
				logger.L.Warn("Dify chat analysis failed, falling back to Ollama", zap.Error(err))
				analysisText, err = h.ollamaService.AnalyzeAlert(ctx, alert)
			}
		} else {
			// 解析工作流结果
			if analysis, ok := workflowResult["analysis"].(string); ok {
				analysisText = analysis
			}
			if actionPlan, ok := workflowResult["action_plan"].(string); ok {
				result.ActionPlan = actionPlan
			}
			if confidence, ok := workflowResult["confidence"].(float64); ok {
				result.Confidence = confidence
			} else {
				result.Confidence = 0.8 // 默认置信度
			}
			
			// 保存额外的元数据
			result.Metadata["dify_workflow_result"] = workflowResult
			result.Metadata["analysis_engine"] = "dify_workflow"
		}
	} else {
		// 使用Ollama进行分析
		logger.L.Debug("Using Ollama for comprehensive analysis")
		analysisText, err = h.ollamaService.AnalyzeAlert(ctx, alert)
		result.Metadata["analysis_engine"] = "ollama"
		result.Confidence = 0.7 // Ollama默认置信度
	}

	if err != nil {
		return nil, fmt.Errorf("comprehensive analysis failed: %w", err)
	}

	result.Result = analysisText

	// 如果没有生成处理方案，尝试生成
	if result.ActionPlan == "" && h.useDify && h.difyService != nil {
		actionPlan, planErr := h.difyService.GenerateActionPlan(ctx, alert, analysisText)
		if planErr != nil {
			logger.L.Warn("Failed to generate action plan", zap.Error(planErr))
		} else {
			result.ActionPlan = actionPlan
		}
	}

	// 查找相似告警
	similarAlerts, similarErr := h.findSimilarAlerts(ctx, alert)
	if similarErr != nil {
		logger.L.Warn("Failed to find similar alerts", zap.Error(similarErr))
	} else {
		result.SimilarAlerts = similarAlerts
	}

	return result, nil
}

// performQuickAnalysis 执行快速分析
func (h *AIAnalysisHandler) performQuickAnalysis(ctx context.Context, alert *model.Alert, result *AIAnalysisResult) (*AIAnalysisResult, error) {
	var analysisText string
	var err error

	// 快速分析优先使用Dify Chat API
	if h.useDify && h.difyService != nil {
		logger.L.Debug("Using Dify for quick analysis")
		analysisText, err = h.difyService.AnalyzeAlert(ctx, alert)
		result.Metadata["analysis_engine"] = "dify_chat"
		result.Confidence = 0.75
	} else {
		logger.L.Debug("Using Ollama for quick analysis")
		analysisText, err = h.ollamaService.AnalyzeAlert(ctx, alert)
		result.Metadata["analysis_engine"] = "ollama"
		result.Confidence = 0.7
	}

	if err != nil {
		return nil, fmt.Errorf("quick analysis failed: %w", err)
	}

	result.Result = analysisText
	return result, nil
}

// performRootCauseAnalysis 执行根因分析
func (h *AIAnalysisHandler) performRootCauseAnalysis(ctx context.Context, alert *model.Alert, result *AIAnalysisResult) (*AIAnalysisResult, error) {
	// 根因分析需要更深入的分析，优先使用Dify
	if h.useDify && h.difyService != nil {
		logger.L.Debug("Using Dify for root cause analysis")
		
		// 构建根因分析的特定输入
		inputs := map[string]interface{}{
			"alert_title":   alert.Title,
			"alert_level":   alert.Level,
			"alert_source":  alert.Source,
			"alert_content": alert.Content,
			"analysis_type": "root_cause",
		}

		// 使用特定的根因分析查询
		query := fmt.Sprintf(`请对以下告警进行深入的根因分析：

告警标题：%s
告警级别：%s
告警来源：%s
告警内容：%s

请重点分析：
1. 告警的直接触发原因
2. 可能的系统性根本原因
3. 相关的依赖关系和影响链
4. 历史类似问题的模式分析

请提供详细的技术分析和建议。`, alert.Title, alert.Level, alert.Source, alert.Content)

		// 这里应该调用Dify的特定根因分析API，暂时使用通用分析
		analysisText, err := h.difyService.AnalyzeAlert(ctx, alert)
		if err != nil {
			logger.L.Warn("Dify root cause analysis failed, falling back to Ollama", zap.Error(err))
			analysisText, err = h.ollamaService.AnalyzeAlert(ctx, alert)
			result.Metadata["analysis_engine"] = "ollama"
			result.Confidence = 0.7
		} else {
			result.Metadata["analysis_engine"] = "dify_root_cause"
			result.Confidence = 0.85
		}

		if err != nil {
			return nil, fmt.Errorf("root cause analysis failed: %w", err)
		}

		result.Result = analysisText
	} else {
		// 使用Ollama进行根因分析
		logger.L.Debug("Using Ollama for root cause analysis")
		analysisText, err := h.ollamaService.AnalyzeAlert(ctx, alert)
		if err != nil {
			return nil, fmt.Errorf("root cause analysis failed: %w", err)
		}

		result.Result = analysisText
		result.Metadata["analysis_engine"] = "ollama"
		result.Confidence = 0.7
	}

	return result, nil
}

// performSimilarSearch 执行相似告警搜索
func (h *AIAnalysisHandler) performSimilarSearch(ctx context.Context, alert *model.Alert, result *AIAnalysisResult) (*AIAnalysisResult, error) {
	similarAlerts, err := h.findSimilarAlerts(ctx, alert)
	if err != nil {
		return nil, fmt.Errorf("similar search failed: %w", err)
	}

	result.SimilarAlerts = similarAlerts
	result.Result = fmt.Sprintf("找到 %d 个相似告警", len(similarAlerts))
	result.Confidence = 0.9
	result.Metadata["analysis_engine"] = "similarity_search"

	return result, nil
}

// findSimilarAlerts 查找相似告警
func (h *AIAnalysisHandler) findSimilarAlerts(ctx context.Context, alert *model.Alert) ([]string, error) {
	if h.useDify && h.difyService != nil {
		return h.difyService.FindSimilarAlerts(ctx, alert)
	}

	// 如果没有Dify，使用Ollama查找相似告警
	similarAlerts, err := h.ollamaService.FindSimilarAlerts(ctx, alert)
	if err != nil {
		return nil, err
	}

	// 将Alert对象转换为ID字符串
	alertIDs := make([]string, len(similarAlerts))
	for i, similarAlert := range similarAlerts {
		alertIDs[i] = fmt.Sprintf("%d", similarAlert.ID)
	}

	return alertIDs, nil
}

// updateAlertStatus 更新告警状态
func (h *AIAnalysisHandler) updateAlertStatus(ctx context.Context, alert *model.Alert, status string) error {
	alert.Status = status
	return database.DB.Save(alert).Error
}

// saveAnalysisResult 保存分析结果
func (h *AIAnalysisHandler) saveAnalysisResult(ctx context.Context, alert *model.Alert, result *AIAnalysisResult) error {
	// 将分析结果保存到告警的Analysis字段
	alert.Analysis = result.Result

	// 如果有处理方案，也保存到告警中
	if result.ActionPlan != "" {
		// 这里可以扩展Alert模型来包含ActionPlan字段
		// 暂时将其添加到Analysis中
		alert.Analysis += "\n\n## 处理方案\n" + result.ActionPlan
	}

	// 保存相似告警信息
	if len(result.SimilarAlerts) > 0 {
		similarAlertsJSON, _ := json.Marshal(result.SimilarAlerts)
		alert.Analysis += "\n\n## 相似告警\n" + string(similarAlertsJSON)
	}

	// 保存到数据库
	if err := database.DB.Save(alert).Error; err != nil {
		return fmt.Errorf("failed to update alert analysis in database: %w", err)
	}

	// 可以考虑将完整的分析结果保存到单独的表中
	// 这里暂时只更新Alert表

	logger.L.Info("Analysis result saved successfully",
		zap.Uint("alert_id", alert.ID),
		zap.String("analysis_type", result.AnalysisType),
		zap.Float64("confidence", result.Confidence),
		zap.Duration("processing_time", result.ProcessingTime),
	)

	return nil
}

// SetUseDify 设置是否使用Dify
func (h *AIAnalysisHandler) SetUseDify(useDify bool) {
	h.useDify = useDify
}

// GetAnalysisEngines 获取可用的分析引擎
func (h *AIAnalysisHandler) GetAnalysisEngines() []string {
	engines := []string{}
	
	if h.difyService != nil {
		engines = append(engines, "dify")
	}
	
	if h.ollamaService != nil {
		engines = append(engines, "ollama")
	}
	
	return engines
}