package worker

import (
	"context"
	"fmt"

	"alert_agent/internal/model"
	"alert_agent/internal/pkg/database"
	"alert_agent/internal/pkg/logger"
	"alert_agent/internal/pkg/queue"
	"alert_agent/internal/service"

	"go.uber.org/zap"
)

// AIAnalysisHandler AI分析任务处理器
type AIAnalysisHandler struct {
	ollamaService *service.OllamaService
}

// NewAIAnalysisHandler 创建AI分析处理器
func NewAIAnalysisHandler(ollamaService *service.OllamaService) *AIAnalysisHandler {
	return &AIAnalysisHandler{
		ollamaService: ollamaService,
	}
}

// Type 返回处理器类型
func (h *AIAnalysisHandler) Type() queue.TaskType {
	return queue.TaskTypeAIAnalysis
}

// Handle 处理AI分析任务
func (h *AIAnalysisHandler) Handle(ctx context.Context, task *queue.Task) error {
	// 解析任务载荷
	alertID, ok := task.Payload["alert_id"].(string)
	if !ok {
		return fmt.Errorf("invalid alert_id in task payload")
	}

	analysisType, _ := task.Payload["analysis_type"].(string)
	if analysisType == "" {
		analysisType = "root_cause"
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

	// 调用Ollama服务进行分析
	analysis, err := h.ollamaService.AnalyzeAlert(ctx, &alert)
	if err != nil {
		return fmt.Errorf("AI analysis failed for alert %s: %w", alertID, err)
	}

	// 更新数据库中的分析结果
	alert.Analysis = analysis
	if err := database.DB.Save(&alert).Error; err != nil {
		return fmt.Errorf("failed to update alert analysis in database: %w", err)
	}

	logger.L.Info("AI analysis completed successfully",
		zap.String("task_id", task.ID),
		zap.String("alert_id", alertID),
	)

	return nil
}
