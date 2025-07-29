package http

import (
	"net/http"
	"strconv"
	"time"

	analysisDomain "alert_agent/internal/domain/analysis"
	"alert_agent/internal/model"
	"alert_agent/internal/pkg/logger"
	"alert_agent/pkg/types"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// APIResponse 使用统一的API响应格式
type APIResponse = types.APIResponse

// AnalysisHandler 分析处理器
type AnalysisHandler struct {
	analysisService analysisDomain.AnalysisService
	logger          *zap.Logger
}

// NewAnalysisHandler 创建分析处理器
func NewAnalysisHandler(analysisService analysisDomain.AnalysisService) *AnalysisHandler {
	return &AnalysisHandler{
		analysisService: analysisService,
		logger:          logger.L.Named("analysis-handler"),
	}
}

// SubmitAnalysisRequest 提交分析请求
type SubmitAnalysisRequest struct {
	AlertID     uint                   `json:"alert_id" binding:"required"`
	Type        analysisDomain.AnalysisType `json:"type" binding:"required"`
	Priority    int                    `json:"priority"`
	Timeout     int                    `json:"timeout"` // 秒
	Options     map[string]interface{} `json:"options"`
	Callback    string                 `json:"callback"`
}

// SubmitAnalysisResponse 提交分析响应
type SubmitAnalysisResponse struct {
	TaskID    string `json:"task_id"`
	Status    string `json:"status"`
	Message   string `json:"message"`
	CreatedAt string `json:"created_at"`
}

// AnalysisResultResponse 分析结果响应
type AnalysisResultResponse struct {
	ID              string                 `json:"id"`
	TaskID          string                 `json:"task_id"`
	AlertID         string                 `json:"alert_id"`
	Type            string                 `json:"type"`
	Status          string                 `json:"status"`
	ConfidenceScore float64                `json:"confidence_score"`
	ProcessingTime  string                 `json:"processing_time"`
	Result          map[string]interface{} `json:"result"`
	Summary         string                 `json:"summary"`
	Recommendations []string               `json:"recommendations"`
	ErrorMessage    string                 `json:"error_message"`
	CreatedAt       string                 `json:"created_at"`
	UpdatedAt       string                 `json:"updated_at"`
}

// AnalysisProgressResponse 分析进度响应
type AnalysisProgressResponse struct {
	TaskID    string  `json:"task_id"`
	Stage     string  `json:"stage"`
	Progress  float64 `json:"progress"`
	Message   string  `json:"message"`
	UpdatedAt string  `json:"updated_at"`
}

// AnalysisTaskResponse 分析任务响应
type AnalysisTaskResponse struct {
	ID          string                 `json:"id"`
	AlertID     string                 `json:"alert_id"`
	Type        string                 `json:"type"`
	Status      string                 `json:"status"`
	Priority    int                    `json:"priority"`
	RetryCount  int                    `json:"retry_count"`
	MaxRetries  int                    `json:"max_retries"`
	Timeout     string                 `json:"timeout"`
	CreatedAt   string                 `json:"created_at"`
	UpdatedAt   string                 `json:"updated_at"`
	StartedAt   *string                `json:"started_at"`
	CompletedAt *string                `json:"completed_at"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// SubmitAnalysis 提交分析任务
// @Summary 提交分析任务
// @Description 提交一个新的告警分析任务
// @Tags analysis
// @Accept json
// @Produce json
// @Param request body SubmitAnalysisRequest true "分析请求"
// @Success 200 {object} APIResponse{data=SubmitAnalysisResponse}
// @Failure 400 {object} APIResponse
// @Failure 500 {object} APIResponse
// @Router /api/v1/analysis/submit [post]
func (h *AnalysisHandler) SubmitAnalysis(c *gin.Context) {
	var req SubmitAnalysisRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid request", zap.Error(err))
		c.JSON(http.StatusBadRequest, types.NewErrorResponse("Invalid request: "+err.Error()))
		return
	}

	// 构建分析请求
	analysisReq := &analysisDomain.AnalysisRequest{
		Alert: &model.Alert{
			ID: req.AlertID,
		},
		Type:     req.Type,
		Priority: req.Priority,
		Timeout:  time.Duration(req.Timeout) * time.Second,
		Options:  req.Options,
		Callback: req.Callback,
	}

	// 设置默认值
	if analysisReq.Priority == 0 {
		analysisReq.Priority = 5 // 默认优先级
	}
	if analysisReq.Timeout == 0 {
		analysisReq.Timeout = 5 * time.Minute // 默认超时时间
	}

	// 提交分析任务
	task, err := h.analysisService.SubmitAnalysis(c.Request.Context(), analysisReq)
	if err != nil {
		h.logger.Error("Failed to submit analysis", zap.Error(err))
		c.JSON(http.StatusInternalServerError, types.NewErrorResponse("Failed to submit analysis: "+err.Error()))
		return
	}

	response := SubmitAnalysisResponse{
		TaskID:    task.ID,
		Status:    string(task.Status),
		Message:   "Analysis task submitted successfully",
		CreatedAt: task.CreatedAt.Format(time.RFC3339),
	}

	c.JSON(http.StatusOK, types.NewSuccessResponse("Analysis task submitted", response))
}

// GetAnalysisResult 获取分析结果
// @Summary 获取分析结果
// @Description 根据任务ID获取分析结果
// @Tags analysis
// @Accept json
// @Produce json
// @Param task_id path string true "任务ID"
// @Success 200 {object} APIResponse{data=AnalysisResultResponse}
// @Failure 400 {object} APIResponse
// @Failure 404 {object} APIResponse
// @Failure 500 {object} APIResponse
// @Router /api/v1/analysis/result/{task_id} [get]
func (h *AnalysisHandler) GetAnalysisResult(c *gin.Context) {
	taskID := c.Param("task_id")
	if taskID == "" {
		c.JSON(http.StatusBadRequest, types.NewErrorResponse("Task ID is required"))
		return
	}

	result, err := h.analysisService.GetAnalysisResult(c.Request.Context(), taskID)
	if err != nil {
		h.logger.Error("Failed to get analysis result", zap.String("task_id", taskID), zap.Error(err))
		c.JSON(http.StatusNotFound, types.NewErrorResponse("Analysis result not found"))
		return
	}

	response := h.convertAnalysisResult(result)
	c.JSON(http.StatusOK, types.NewSuccessResponse("Analysis result retrieved", response))
}

// GetAnalysisProgress 获取分析进度
// @Summary 获取分析进度
// @Description 根据任务ID获取分析进度
// @Tags analysis
// @Accept json
// @Produce json
// @Param task_id path string true "任务ID"
// @Success 200 {object} APIResponse{data=AnalysisProgressResponse}
// @Failure 400 {object} APIResponse
// @Failure 404 {object} APIResponse
// @Failure 500 {object} APIResponse
// @Router /api/v1/analysis/progress/{task_id} [get]
func (h *AnalysisHandler) GetAnalysisProgress(c *gin.Context) {
	taskID := c.Param("task_id")
	if taskID == "" {
		c.JSON(http.StatusBadRequest, types.NewErrorResponse("Task ID is required"))
		return
	}

	progress, err := h.analysisService.GetAnalysisProgress(c.Request.Context(), taskID)
	if err != nil {
		h.logger.Error("Failed to get analysis progress", zap.String("task_id", taskID), zap.Error(err))
		c.JSON(http.StatusNotFound, types.NewErrorResponse("Analysis progress not found"))
		return
	}

	response := AnalysisProgressResponse{
		TaskID:    progress.TaskID,
		Stage:     progress.Stage,
		Progress:  progress.Progress,
		Message:   progress.Message,
		UpdatedAt: progress.UpdatedAt.Format(time.RFC3339),
	}

	c.JSON(http.StatusOK, types.NewSuccessResponse("Analysis progress retrieved", response))
}

// CancelAnalysis 取消分析任务
// @Summary 取消分析任务
// @Description 根据任务ID取消分析任务
// @Tags analysis
// @Accept json
// @Produce json
// @Param task_id path string true "任务ID"
// @Success 200 {object} APIResponse
// @Failure 400 {object} APIResponse
// @Failure 404 {object} APIResponse
// @Failure 500 {object} APIResponse
// @Router /api/v1/analysis/cancel/{task_id} [post]
func (h *AnalysisHandler) CancelAnalysis(c *gin.Context) {
	taskID := c.Param("task_id")
	if taskID == "" {
		c.JSON(http.StatusBadRequest, types.NewErrorResponse("Task ID is required"))
		return
	}

	err := h.analysisService.CancelAnalysis(c.Request.Context(), taskID)
	if err != nil {
		h.logger.Error("Failed to cancel analysis", zap.String("task_id", taskID), zap.Error(err))
		c.JSON(http.StatusInternalServerError, types.NewErrorResponse("Failed to cancel analysis: "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, types.NewSuccessResponse("Analysis task cancelled", nil))
}

// RetryAnalysis 重试分析任务
// @Summary 重试分析任务
// @Description 根据任务ID重试分析任务
// @Tags analysis
// @Accept json
// @Produce json
// @Param task_id path string true "任务ID"
// @Success 200 {object} APIResponse
// @Failure 400 {object} APIResponse
// @Failure 404 {object} APIResponse
// @Failure 500 {object} APIResponse
// @Router /api/v1/analysis/retry/{task_id} [post]
func (h *AnalysisHandler) RetryAnalysis(c *gin.Context) {
	taskID := c.Param("task_id")
	if taskID == "" {
		c.JSON(http.StatusBadRequest, types.NewErrorResponse("Task ID is required"))
		return
	}

	err := h.analysisService.RetryAnalysis(c.Request.Context(), taskID)
	if err != nil {
		h.logger.Error("Failed to retry analysis", zap.String("task_id", taskID), zap.Error(err))
		c.JSON(http.StatusInternalServerError, types.NewErrorResponse("Failed to retry analysis: "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, types.NewSuccessResponse("Analysis task retried", nil))
}

// ListAnalysisTasks 列出分析任务
// @Summary 列出分析任务
// @Description 根据过滤条件列出分析任务
// @Tags analysis
// @Accept json
// @Produce json
// @Param alert_id query string false "告警ID"
// @Param type query string false "分析类型"
// @Param status query string false "任务状态"
// @Param limit query int false "限制数量" default(20)
// @Param offset query int false "偏移量" default(0)
// @Success 200 {object} APIResponse{data=[]AnalysisTaskResponse}
// @Failure 400 {object} APIResponse
// @Failure 500 {object} APIResponse
// @Router /api/v1/analysis/tasks [get]
func (h *AnalysisHandler) ListAnalysisTasks(c *gin.Context) {
	// 解析查询参数
	filter := analysisDomain.AnalysisFilter{
		Limit:  20,
		Offset: 0,
	}

	if alertID := c.Query("alert_id"); alertID != "" {
		filter.AlertIDs = []string{alertID}
	}

	if analysisType := c.Query("type"); analysisType != "" {
		filter.Types = []analysisDomain.AnalysisType{analysisDomain.AnalysisType(analysisType)}
	}

	if status := c.Query("status"); status != "" {
		filter.Statuses = []analysisDomain.AnalysisStatus{analysisDomain.AnalysisStatus(status)}
	}

	if limitStr := c.Query("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 {
			filter.Limit = limit
		}
	}

	if offsetStr := c.Query("offset"); offsetStr != "" {
		if offset, err := strconv.Atoi(offsetStr); err == nil && offset >= 0 {
			filter.Offset = offset
		}
	}

	tasks, err := h.analysisService.GetAnalysisTasks(c.Request.Context(), filter)
	if err != nil {
		h.logger.Error("Failed to list analysis tasks", zap.Error(err))
		c.JSON(http.StatusInternalServerError, types.NewErrorResponse("Failed to list analysis tasks: "+err.Error()))
		return
	}

	response := make([]AnalysisTaskResponse, len(tasks))
	for i, task := range tasks {
		response[i] = h.convertAnalysisTask(task)
	}

	c.JSON(http.StatusOK, types.NewSuccessResponse("Analysis tasks retrieved", response))
}

// GetAnalysisStatistics 获取分析统计信息
// @Summary 获取分析统计信息
// @Description 获取分析任务的统计信息
// @Tags analysis
// @Accept json
// @Produce json
// @Param start_time query string false "开始时间 (RFC3339格式)"
// @Param end_time query string false "结束时间 (RFC3339格式)"
// @Success 200 {object} APIResponse{data=AnalysisStatistics}
// @Failure 400 {object} APIResponse
// @Failure 500 {object} APIResponse
// @Router /api/v1/analysis/statistics [get]
func (h *AnalysisHandler) GetAnalysisStatistics(c *gin.Context) {
	var timeRange *analysisDomain.TimeRange

	startTimeStr := c.Query("start_time")
	endTimeStr := c.Query("end_time")

	if startTimeStr != "" || endTimeStr != "" {
		timeRange = &analysisDomain.TimeRange{}

		if startTimeStr != "" {
			if startTime, err := time.Parse(time.RFC3339, startTimeStr); err == nil {
				timeRange.Start = startTime
			} else {
				c.JSON(http.StatusBadRequest, types.NewErrorResponse("Invalid start_time format"))
				return
			}
		}

		if endTimeStr != "" {
			if endTime, err := time.Parse(time.RFC3339, endTimeStr); err == nil {
				timeRange.End = endTime
			} else {
				c.JSON(http.StatusBadRequest, types.NewErrorResponse("Invalid end_time format"))
				return
			}
		}
	}

	stats, err := h.analysisService.GetAnalysisStatistics(c.Request.Context(), timeRange)
	if err != nil {
		h.logger.Error("Failed to get analysis statistics", zap.Error(err))
		c.JSON(http.StatusInternalServerError, types.NewErrorResponse("Failed to get analysis statistics: "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, types.NewSuccessResponse("Analysis statistics retrieved", stats))
}

// GetQueueStatus 获取队列状态
// @Summary 获取队列状态
// @Description 获取分析任务队列的状态信息
// @Tags analysis
// @Accept json
// @Produce json
// @Success 200 {object} APIResponse{data=QueueStatus}
// @Failure 500 {object} APIResponse
// @Router /api/v1/analysis/queue/status [get]
func (h *AnalysisHandler) GetQueueStatus(c *gin.Context) {
	status, err := h.analysisService.GetQueueStatus(c.Request.Context())
	if err != nil {
		h.logger.Error("Failed to get queue status", zap.Error(err))
		c.JSON(http.StatusInternalServerError, types.NewErrorResponse("Failed to get queue status: "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, types.NewSuccessResponse("Queue status retrieved", status))
}

// GetWorkerStatuses 获取工作器状态
// @Summary 获取工作器状态
// @Description 获取所有分析工作器的状态信息
// @Tags analysis
// @Accept json
// @Produce json
// @Success 200 {object} APIResponse{data=[]WorkerStatus}
// @Failure 500 {object} APIResponse
// @Router /api/v1/analysis/workers/status [get]
func (h *AnalysisHandler) GetWorkerStatuses(c *gin.Context) {
	statuses, err := h.analysisService.GetWorkerStatuses(c.Request.Context())
	if err != nil {
		h.logger.Error("Failed to get worker statuses", zap.Error(err))
		c.JSON(http.StatusInternalServerError, types.NewErrorResponse("Failed to get worker statuses: "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, types.NewSuccessResponse("Worker statuses retrieved", statuses))
}

// HealthCheck 健康检查
// @Summary 健康检查
// @Description 检查分析服务的健康状态
// @Tags analysis
// @Accept json
// @Produce json
// @Success 200 {object} APIResponse
// @Failure 500 {object} APIResponse
// @Router /api/v1/analysis/health [get]
func (h *AnalysisHandler) HealthCheck(c *gin.Context) {
	err := h.analysisService.HealthCheck(c.Request.Context())
	if err != nil {
		h.logger.Error("Health check failed", zap.Error(err))
		c.JSON(http.StatusInternalServerError, types.NewErrorResponse("Health check failed: "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, types.NewSuccessResponse("Analysis service is healthy", nil))
}

// convertAnalysisResult 转换分析结果
func (h *AnalysisHandler) convertAnalysisResult(result *analysisDomain.AnalysisResult) AnalysisResultResponse {
	return AnalysisResultResponse{
		ID:              result.ID,
		TaskID:          result.TaskID,
		AlertID:         result.AlertID,
		Type:            string(result.Type),
		Status:          string(result.Status),
		ConfidenceScore: result.ConfidenceScore,
		ProcessingTime:  result.ProcessingTime.String(),
		Result:          result.Result,
		Summary:         result.Summary,
		Recommendations: result.Recommendations,
		ErrorMessage:    result.ErrorMessage,
		CreatedAt:       result.CreatedAt.Format(time.RFC3339),
		UpdatedAt:       result.UpdatedAt.Format(time.RFC3339),
	}
}

// convertAnalysisTask 转换分析任务
func (h *AnalysisHandler) convertAnalysisTask(task *analysisDomain.AnalysisTask) AnalysisTaskResponse {
	var startedAt, completedAt *string
	if task.StartedAt != nil {
		formatted := task.StartedAt.Format(time.RFC3339)
		startedAt = &formatted
	}
	if task.CompletedAt != nil {
		formatted := task.CompletedAt.Format(time.RFC3339)
		completedAt = &formatted
	}

	return AnalysisTaskResponse{
		ID:          task.ID,
		AlertID:     task.AlertID,
		Type:        string(task.Type),
		Status:      string(task.Status),
		Priority:    task.Priority,
		RetryCount:  task.RetryCount,
		MaxRetries:  task.MaxRetries,
		Timeout:     task.Timeout.String(),
		CreatedAt:   task.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   task.UpdatedAt.Format(time.RFC3339),
		StartedAt:   startedAt,
		CompletedAt: completedAt,
		Metadata:    task.Metadata,
	}
}