package http

import (
	"alert_agent/internal/application/analysis"
	domainAnalysis "alert_agent/internal/domain/analysis"
	"alert_agent/internal/shared/logger"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// RegisterN8NRoutes 注册 n8n 相关路由
func RegisterN8NRoutes(router *gin.RouterGroup, analysisService *analysis.N8NAnalysisService) {
	// 创建 n8n 处理器
	n8nHandler := NewN8NHandler(analysisService, logger.GetLogger())

	// n8n 分析相关路由
	n8nGroup := router.Group("/n8n")
	{
		// 告警分析
		n8nGroup.POST("/alerts/:id/analyze", n8nHandler.AnalyzeAlert)
		n8nGroup.POST("/alerts/batch-analyze", n8nHandler.BatchAnalyzeAlerts)
		n8nGroup.GET("/alerts/:alert_id/analysis-history", n8nHandler.GetAnalysisHistory)

		// 执行管理
		n8nGroup.GET("/executions/:execution_id/status", n8nHandler.GetAnalysisStatus)
		n8nGroup.POST("/executions/:execution_id/cancel", n8nHandler.CancelAnalysis)
		n8nGroup.POST("/executions/:execution_id/retry", n8nHandler.RetryAnalysis)

		// 分析指标
		n8nGroup.GET("/metrics", n8nHandler.GetAnalysisMetrics)
	}

	// 兼容性路由 - 保持与原有 API 路径一致
	analysisGroup := router.Group("/analysis")
	{
		// 执行状态和管理
		analysisGroup.GET("/executions/:execution_id/status", n8nHandler.GetAnalysisStatus)
		analysisGroup.POST("/executions/:execution_id/cancel", n8nHandler.CancelAnalysis)
		analysisGroup.POST("/executions/:execution_id/retry", n8nHandler.RetryAnalysis)
		analysisGroup.GET("/metrics", n8nHandler.GetAnalysisMetrics)
	}

	// 告警相关路由
	alertsGroup := router.Group("/alerts")
	{
		// 单个告警分析
		alertsGroup.POST("/:id/analyze", n8nHandler.AnalyzeAlert)
		// 批量分析
		alertsGroup.POST("/batch-analyze", n8nHandler.BatchAnalyzeAlerts)
		// 分析历史
		alertsGroup.GET("/:alert_id/analysis-history", n8nHandler.GetAnalysisHistory)
	}
}

// RegisterN8NCallbackRoutes 注册 n8n 回调路由
func RegisterN8NCallbackRoutes(router *gin.RouterGroup, workflowManager domainAnalysis.N8NWorkflowManager) {
	// n8n 回调处理器
	callbackHandler := &N8NCallbackHandler{
		workflowManager: workflowManager,
		logger:          logger.GetLogger(),
	}

	// 回调路由
	callbackGroup := router.Group("/callbacks")
	{
		callbackGroup.POST("/n8n/workflow/:execution_id", callbackHandler.HandleWorkflowCallback)
		callbackGroup.POST("/n8n/webhook/:webhook_id", callbackHandler.HandleWebhookCallback)
	}
}

// N8NCallbackHandler n8n 回调处理器
type N8NCallbackHandler struct {
	workflowManager domainAnalysis.N8NWorkflowManager
	logger          *zap.Logger
}

// WorkflowCallbackRequest 工作流回调请求
type WorkflowCallbackRequest struct {
	ExecutionID string                 `json:"execution_id"`
	Status      string                 `json:"status"`
	Data        map[string]interface{} `json:"data"`
	Error       *string                `json:"error,omitempty"`
	Timestamp   string                 `json:"timestamp"`
}

// WebhookCallbackRequest Webhook 回调请求
type WebhookCallbackRequest struct {
	WebhookID string                 `json:"webhook_id"`
	EventType string                 `json:"event_type"`
	Data      map[string]interface{} `json:"data"`
	Timestamp string                 `json:"timestamp"`
}

// HandleWorkflowCallback 处理工作流回调
// @Summary 处理 n8n 工作流回调
// @Description 接收 n8n 工作流执行完成后的回调通知
// @Tags callbacks
// @Accept json
// @Produce json
// @Param execution_id path string true "执行ID"
// @Param request body WorkflowCallbackRequest true "回调数据"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/callbacks/n8n/workflow/{execution_id} [post]
func (h *N8NCallbackHandler) HandleWorkflowCallback(c *gin.Context) {
	executionID := c.Param("execution_id")
	if executionID == "" {
		c.JSON(400, ErrorResponse{
			Error:   "missing_execution_id",
			Message: "Execution ID is required",
		})
		return
	}

	var req WorkflowCallbackRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid callback request", zap.Error(err))
		c.JSON(400, ErrorResponse{
			Error:   "invalid_request",
			Message: err.Error(),
		})
		return
	}

	// 验证执行ID匹配
	if req.ExecutionID != executionID {
		h.logger.Warn("Execution ID mismatch",
			zap.String("url_execution_id", executionID),
			zap.String("body_execution_id", req.ExecutionID))
		c.JSON(400, ErrorResponse{
			Error:   "execution_id_mismatch",
			Message: "Execution ID in URL and body do not match",
		})
		return
	}

	// 处理回调
	err := h.workflowManager.HandleCallback(c.Request.Context(), executionID, req.Data)
	if err != nil {
		h.logger.Error("Failed to handle workflow callback",
			zap.String("execution_id", executionID),
			zap.Error(err))
		c.JSON(500, ErrorResponse{
			Error:   "callback_processing_failed",
			Message: "Failed to process workflow callback",
		})
		return
	}

	h.logger.Info("Workflow callback processed successfully",
		zap.String("execution_id", executionID),
		zap.String("status", req.Status))

	c.JSON(200, SuccessResponse{
		Message: "Callback processed successfully",
	})
}

// HandleWebhookCallback 处理 Webhook 回调
// @Summary 处理 n8n Webhook 回调
// @Description 接收 n8n Webhook 事件通知
// @Tags callbacks
// @Accept json
// @Produce json
// @Param webhook_id path string true "Webhook ID"
// @Param request body WebhookCallbackRequest true "Webhook 数据"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/callbacks/n8n/webhook/{webhook_id} [post]
func (h *N8NCallbackHandler) HandleWebhookCallback(c *gin.Context) {
	webhookID := c.Param("webhook_id")
	if webhookID == "" {
		c.JSON(400, ErrorResponse{
			Error:   "missing_webhook_id",
			Message: "Webhook ID is required",
		})
		return
	}

	var req WebhookCallbackRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid webhook request", zap.Error(err))
		c.JSON(400, ErrorResponse{
			Error:   "invalid_request",
			Message: err.Error(),
		})
		return
	}

	// 验证 Webhook ID 匹配
	if req.WebhookID != webhookID {
		h.logger.Warn("Webhook ID mismatch",
			zap.String("url_webhook_id", webhookID),
			zap.String("body_webhook_id", req.WebhookID))
		c.JSON(400, ErrorResponse{
			Error:   "webhook_id_mismatch",
			Message: "Webhook ID in URL and body do not match",
		})
		return
	}

	// 记录 Webhook 事件
	h.logger.Info("Webhook callback received",
		zap.String("webhook_id", webhookID),
		zap.String("event_type", req.EventType))

	// 这里可以根据事件类型进行不同的处理
	// 目前只是记录日志

	c.JSON(200, SuccessResponse{
		Message: "Webhook processed successfully",
	})
}