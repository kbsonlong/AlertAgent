package http

import (
	"net/http"
	"strconv"
	"time"

	"alert_agent/internal/application/analysis"
	"alert_agent/internal/domain/alert"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// N8NHandler n8n 分析处理器
type N8NHandler struct {
	analysisService *analysis.N8NAnalysisService
	logger          *zap.Logger
}

// NewN8NHandler 创建新的 n8n 处理器
func NewN8NHandler(analysisService *analysis.N8NAnalysisService, logger *zap.Logger) *N8NHandler {
	return &N8NHandler{
		analysisService: analysisService,
		logger:          logger,
	}
}

// AnalyzeAlertRequest 分析告警请求
type AnalyzeAlertRequest struct {
	WorkflowTemplateID string `json:"workflow_template_id" binding:"required"`
}

// AnalyzeAlertResponse 分析告警响应
type AnalyzeAlertResponse struct {
	ExecutionID string `json:"execution_id"`
	Status      string `json:"status"`
	Message     string `json:"message"`
}

// BatchAnalyzeRequest 批量分析请求
type BatchAnalyzeRequest struct {
	WorkflowTemplateID  string        `json:"workflow_template_id" binding:"required"`
	BatchSize           int           `json:"batch_size,omitempty"`
	ProcessInterval     time.Duration `json:"process_interval,omitempty"`
	MaxRetries          int           `json:"max_retries,omitempty"`
	Timeout             time.Duration `json:"timeout,omitempty"`
	AutoAnalysisEnabled bool          `json:"auto_analysis_enabled,omitempty"`
}

// AnalysisStatusResponse 分析状态响应
type AnalysisStatusResponse struct {
	ExecutionID   string                 `json:"execution_id"`
	WorkflowID    string                 `json:"workflow_id"`
	Status        string                 `json:"status"`
	StartedAt     time.Time              `json:"started_at"`
	FinishedAt    *time.Time             `json:"finished_at,omitempty"`
	Duration      *time.Duration         `json:"duration,omitempty"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
}

// AnalysisMetricsResponse 分析指标响应
type AnalysisMetricsResponse struct {
	TotalExecutions       int64         `json:"total_executions"`
	SuccessfulExecutions  int64         `json:"successful_executions"`
	FailedExecutions      int64         `json:"failed_executions"`
	RunningExecutions     int64         `json:"running_executions"`
	AverageExecutionTime  time.Duration `json:"average_execution_time"`
	TimeRange             TimeRange     `json:"time_range"`
}

// TimeRange 时间范围
type TimeRange struct {
	Start *time.Time `json:"start"`
	End   *time.Time `json:"end"`
}

// AnalyzeAlert 分析单个告警
// @Summary 分析告警
// @Description 使用 n8n 工作流分析指定的告警
// @Tags n8n
// @Accept json
// @Produce json
// @Param id path int true "告警ID"
// @Param request body AnalyzeAlertRequest true "分析请求"
// @Success 200 {object} AnalyzeAlertResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/alerts/{id}/analyze [post]
func (h *N8NHandler) AnalyzeAlert(c *gin.Context) {
	// 获取告警ID
	alertIDStr := c.Param("id")
	alertID, err := strconv.ParseUint(alertIDStr, 10, 32)
	if err != nil {
		h.logger.Error("Invalid alert ID", zap.String("alert_id", alertIDStr), zap.Error(err))
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_alert_id",
			Message: "Invalid alert ID format",
		})
		return
	}

	// 解析请求
	var req AnalyzeAlertRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_request",
			Message: err.Error(),
		})
		return
	}

	// 调用分析服务
	execution, err := h.analysisService.AnalyzeAlert(c.Request.Context(), uint(alertID), req.WorkflowTemplateID)
	if err != nil {
		h.logger.Error("Failed to analyze alert",
			zap.Uint("alert_id", uint(alertID)),
			zap.String("workflow_template_id", req.WorkflowTemplateID),
			zap.Error(err))
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "analysis_failed",
			Message: "Failed to start alert analysis",
		})
		return
	}

	h.logger.Info("Alert analysis started",
		zap.Uint("alert_id", uint(alertID)),
		zap.String("execution_id", execution.ID),
		zap.String("workflow_template_id", req.WorkflowTemplateID))

	c.JSON(http.StatusOK, AnalyzeAlertResponse{
		ExecutionID: execution.ID,
		Status:      string(execution.Status),
		Message:     "Alert analysis started successfully",
	})
}

// BatchAnalyzeAlerts 批量分析告警
// @Summary 批量分析告警
// @Description 批量分析需要处理的告警
// @Tags n8n
// @Accept json
// @Produce json
// @Param request body BatchAnalyzeRequest true "批量分析请求"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/alerts/batch-analyze [post]
func (h *N8NHandler) BatchAnalyzeAlerts(c *gin.Context) {
	// 解析请求
	var req BatchAnalyzeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_request",
			Message: err.Error(),
		})
		return
	}

	// 设置默认值
	config := analysis.N8NAnalysisConfig{
		DefaultWorkflowTemplateID: req.WorkflowTemplateID,
		BatchSize:                 req.BatchSize,
		ProcessInterval:           req.ProcessInterval,
		MaxRetries:                req.MaxRetries,
		Timeout:                   req.Timeout,
		AutoAnalysisEnabled:       req.AutoAnalysisEnabled,
	}

	// 设置默认值
	if config.BatchSize <= 0 {
		config.BatchSize = 10
	}
	if config.ProcessInterval <= 0 {
		config.ProcessInterval = 5 * time.Minute
	}
	if config.MaxRetries <= 0 {
		config.MaxRetries = 3
	}
	if config.Timeout <= 0 {
		config.Timeout = 30 * time.Minute
	}

	// 调用批量分析服务
	err := h.analysisService.BatchAnalyzeAlerts(c.Request.Context(), config)
	if err != nil {
		h.logger.Error("Failed to batch analyze alerts", zap.Error(err))
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "batch_analysis_failed",
			Message: "Failed to start batch analysis",
		})
		return
	}

	h.logger.Info("Batch analysis completed",
		zap.String("workflow_template_id", req.WorkflowTemplateID),
		zap.Int("batch_size", config.BatchSize))

	c.JSON(http.StatusOK, SuccessResponse{
		Message: "Batch analysis completed successfully",
	})
}

// GetAnalysisStatus 获取分析状态
// @Summary 获取分析状态
// @Description 获取指定执行的分析状态
// @Tags n8n
// @Produce json
// @Param execution_id path string true "执行ID"
// @Success 200 {object} AnalysisStatusResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/analysis/executions/{execution_id}/status [get]
func (h *N8NHandler) GetAnalysisStatus(c *gin.Context) {
	executionID := c.Param("execution_id")
	if executionID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "missing_execution_id",
			Message: "Execution ID is required",
		})
		return
	}

	// 获取分析状态
	execution, err := h.analysisService.GetAnalysisStatus(c.Request.Context(), executionID)
	if err != nil {
		h.logger.Error("Failed to get analysis status",
			zap.String("execution_id", executionID),
			zap.Error(err))
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "status_retrieval_failed",
			Message: "Failed to retrieve analysis status",
		})
		return
	}

	if execution == nil {
		c.JSON(http.StatusNotFound, ErrorResponse{
			Error:   "execution_not_found",
			Message: "Execution not found",
		})
		return
	}

	// 构建响应
	response := AnalysisStatusResponse{
		ExecutionID: execution.ID,
		WorkflowID:  execution.WorkflowID,
		Status:      string(execution.Status),
		StartedAt:   execution.StartedAt,
		Metadata:    execution.Metadata,
	}

	// 计算持续时间
	if execution.FinishedAt != nil {
		response.FinishedAt = execution.FinishedAt
		duration := execution.FinishedAt.Sub(execution.StartedAt)
		response.Duration = &duration
	}

	c.JSON(http.StatusOK, response)
}

// CancelAnalysis 取消分析
// @Summary 取消分析
// @Description 取消指定的分析执行
// @Tags n8n
// @Produce json
// @Param execution_id path string true "执行ID"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/analysis/executions/{execution_id}/cancel [post]
func (h *N8NHandler) CancelAnalysis(c *gin.Context) {
	executionID := c.Param("execution_id")
	if executionID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "missing_execution_id",
			Message: "Execution ID is required",
		})
		return
	}

	// 取消分析
	err := h.analysisService.CancelAnalysis(c.Request.Context(), executionID)
	if err != nil {
		h.logger.Error("Failed to cancel analysis",
			zap.String("execution_id", executionID),
			zap.Error(err))
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "cancellation_failed",
			Message: "Failed to cancel analysis",
		})
		return
	}

	h.logger.Info("Analysis cancelled", zap.String("execution_id", executionID))

	c.JSON(http.StatusOK, SuccessResponse{
		Message: "Analysis cancelled successfully",
	})
}

// RetryAnalysis 重试分析
// @Summary 重试分析
// @Description 重试失败的分析执行
// @Tags n8n
// @Produce json
// @Param execution_id path string true "执行ID"
// @Success 200 {object} AnalyzeAlertResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/analysis/executions/{execution_id}/retry [post]
func (h *N8NHandler) RetryAnalysis(c *gin.Context) {
	executionID := c.Param("execution_id")
	if executionID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "missing_execution_id",
			Message: "Execution ID is required",
		})
		return
	}

	// 重试分析
	newExecution, err := h.analysisService.RetryAnalysis(c.Request.Context(), executionID)
	if err != nil {
		h.logger.Error("Failed to retry analysis",
			zap.String("execution_id", executionID),
			zap.Error(err))
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "retry_failed",
			Message: "Failed to retry analysis",
		})
		return
	}

	h.logger.Info("Analysis retried",
		zap.String("original_execution_id", executionID),
		zap.String("new_execution_id", newExecution.ID))

	c.JSON(http.StatusOK, AnalyzeAlertResponse{
		ExecutionID: newExecution.ID,
		Status:      string(newExecution.Status),
		Message:     "Analysis retried successfully",
	})
}

// GetAnalysisHistory 获取分析历史
// @Summary 获取分析历史
// @Description 获取指定告警的分析历史
// @Tags n8n
// @Produce json
// @Param alert_id path int true "告警ID"
// @Param limit query int false "限制数量" default(10)
// @Success 200 {object} []AnalysisStatusResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/alerts/{alert_id}/analysis-history [get]
func (h *N8NHandler) GetAnalysisHistory(c *gin.Context) {
	// 获取告警ID
	alertIDStr := c.Param("alert_id")
	alertID, err := strconv.ParseUint(alertIDStr, 10, 32)
	if err != nil {
		h.logger.Error("Invalid alert ID", zap.String("alert_id", alertIDStr), zap.Error(err))
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_alert_id",
			Message: "Invalid alert ID format",
		})
		return
	}

	// 获取限制数量
	limitStr := c.DefaultQuery("limit", "10")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 10
	}

	// 获取分析历史
	executions, err := h.analysisService.GetAnalysisHistory(c.Request.Context(), uint(alertID), limit)
	if err != nil {
		h.logger.Error("Failed to get analysis history",
			zap.Uint("alert_id", uint(alertID)),
			zap.Error(err))
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "history_retrieval_failed",
			Message: "Failed to retrieve analysis history",
		})
		return
	}

	// 构建响应
	var responses []AnalysisStatusResponse
	for _, execution := range executions {
		response := AnalysisStatusResponse{
			ExecutionID: execution.ID,
			WorkflowID:  execution.WorkflowID,
			Status:      string(execution.Status),
			StartedAt:   execution.StartedAt,
			Metadata:    execution.Metadata,
		}

		// 计算持续时间
		if execution.FinishedAt != nil {
			response.FinishedAt = execution.FinishedAt
			duration := execution.FinishedAt.Sub(execution.StartedAt)
			response.Duration = &duration
		}

		responses = append(responses, response)
	}

	c.JSON(http.StatusOK, responses)
}

// GetAnalysisMetrics 获取分析指标
// @Summary 获取分析指标
// @Description 获取指定时间范围内的分析指标
// @Tags n8n
// @Produce json
// @Param start query string false "开始时间 (RFC3339格式)"
// @Param end query string false "结束时间 (RFC3339格式)"
// @Success 200 {object} AnalysisMetricsResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/analysis/metrics [get]
func (h *N8NHandler) GetAnalysisMetrics(c *gin.Context) {
	// 解析时间范围
	var timeRange alert.TimeRange

	startStr := c.Query("start")
	if startStr != "" {
		start, err := time.Parse(time.RFC3339, startStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse{
				Error:   "invalid_start_time",
				Message: "Invalid start time format, use RFC3339",
			})
			return
		}
		timeRange.Start = &start
	} else {
		// 默认为7天前
		start := time.Now().AddDate(0, 0, -7)
		timeRange.Start = &start
	}

	endStr := c.Query("end")
	if endStr != "" {
		end, err := time.Parse(time.RFC3339, endStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse{
				Error:   "invalid_end_time",
				Message: "Invalid end time format, use RFC3339",
			})
			return
		}
		timeRange.End = &end
	} else {
		// 默认为现在
		end := time.Now()
		timeRange.End = &end
	}

	// 获取分析指标
	metrics, err := h.analysisService.GetAnalysisMetrics(c.Request.Context(), timeRange)
	if err != nil {
		h.logger.Error("Failed to get analysis metrics", zap.Error(err))
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "metrics_retrieval_failed",
			Message: "Failed to retrieve analysis metrics",
		})
		return
	}

	// 构建响应
	response := AnalysisMetricsResponse{
		TotalExecutions:       metrics.TotalExecutions,
		SuccessfulExecutions:  metrics.SuccessfulExecutions,
		FailedExecutions:      metrics.FailedExecutions,
		RunningExecutions:     metrics.RunningExecutions,
		AverageExecutionTime:  metrics.AverageExecutionTime,
		TimeRange: TimeRange{
			Start: metrics.TimeRange.Start,
			End:   metrics.TimeRange.End,
		},
	}

	c.JSON(http.StatusOK, response)
}

// ErrorResponse 错误响应
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

// SuccessResponse 成功响应
type SuccessResponse struct {
	Message string `json:"message"`
}