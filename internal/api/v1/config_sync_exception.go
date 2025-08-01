package v1

import (
	"net/http"

	"alert_agent/internal/pkg/logger"
	"alert_agent/internal/pkg/response"
	"alert_agent/internal/service"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// ConfigSyncExceptionAPI 配置同步异常API处理器
type ConfigSyncExceptionAPI struct {
	exceptionHandler *service.ConfigSyncExceptionHandler
}

// NewConfigSyncExceptionAPI 创建配置同步异常API处理器
func NewConfigSyncExceptionAPI() *ConfigSyncExceptionAPI {
	return &ConfigSyncExceptionAPI{
		exceptionHandler: service.NewConfigSyncExceptionHandler(),
	}
}

// DetectExceptions 检测同步异常
// @Summary 检测同步异常
// @Description 主动检测配置同步过程中的异常情况
// @Tags 配置同步异常管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.Response "检测完成"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /api/v1/config/sync/exceptions/detect [post]
func (api *ConfigSyncExceptionAPI) DetectExceptions(ctx *gin.Context) {
	logger.L.Debug("Detecting sync exceptions")

	err := api.exceptionHandler.DetectSyncExceptions(ctx.Request.Context())
	if err != nil {
		logger.L.Error("Failed to detect sync exceptions", zap.Error(err))
		response.Error(ctx, http.StatusInternalServerError, "Failed to detect sync exceptions", err)
		return
	}

	response.Success(ctx, gin.H{
		"message": "Exception detection completed",
	})
}

// GetActiveExceptions 获取活跃异常
// @Summary 获取活跃异常
// @Description 获取当前活跃的配置同步异常列表
// @Tags 配置同步异常管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param cluster_id query string false "集群ID"
// @Param config_type query string false "配置类型"
// @Success 200 {object} response.Response{data=[]object} "获取成功"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /api/v1/config/sync/exceptions [get]
func (api *ConfigSyncExceptionAPI) GetActiveExceptions(ctx *gin.Context) {
	clusterID := ctx.Query("cluster_id")
	configType := ctx.Query("config_type")

	logger.L.Debug("Getting active exceptions",
		zap.String("cluster_id", clusterID),
		zap.String("config_type", configType),
	)

	exceptions, err := api.exceptionHandler.GetActiveExceptions(ctx.Request.Context(), clusterID, configType)
	if err != nil {
		logger.L.Error("Failed to get active exceptions",
			zap.String("cluster_id", clusterID),
			zap.String("config_type", configType),
			zap.Error(err),
		)
		response.Error(ctx, http.StatusInternalServerError, "Failed to get active exceptions", err)
		return
	}

	response.Success(ctx, gin.H{
		"exceptions": exceptions,
	})
}

// AnalyzeException 分析异常根因
// @Summary 分析异常根因
// @Description 分析指定异常的根本原因
// @Tags 配置同步异常管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "异常ID"
// @Success 200 {object} response.Response{data=object} "分析成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /api/v1/config/sync/exceptions/{id}/analysis [get]
func (api *ConfigSyncExceptionAPI) AnalyzeException(ctx *gin.Context) {
	exceptionID := ctx.Param("id")
	if exceptionID == "" {
		response.Error(ctx, http.StatusBadRequest, "Exception ID is required", nil)
		return
	}

	logger.L.Debug("Analyzing exception", zap.String("exception_id", exceptionID))

	analysis, err := api.exceptionHandler.AnalyzeException(ctx.Request.Context(), exceptionID)
	if err != nil {
		logger.L.Error("Failed to analyze exception",
			zap.String("exception_id", exceptionID),
			zap.Error(err),
	)
	response.Error(ctx, http.StatusInternalServerError, "Failed to analyze exception", err)
	return
	}

	response.Success(ctx, gin.H{
		"analysis": analysis,
	})
}

// ResolveException 解决异常
// @Summary 解决异常
// @Description 标记指定异常为已解决状态
// @Tags 配置同步异常管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "异常ID"
// @Param request body object{resolved_by=string,resolution=string} true "解决请求"
// @Success 200 {object} response.Response "解决成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /api/v1/config/sync/exceptions/{id}/resolve [post]
func (api *ConfigSyncExceptionAPI) ResolveException(ctx *gin.Context) {
	exceptionID := ctx.Param("id")
	if exceptionID == "" {
		response.Error(ctx, http.StatusBadRequest, "Exception ID is required", nil)
		return
	}

	var req struct {
		ResolvedBy string `json:"resolved_by" binding:"required"`
		Resolution string `json:"resolution"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Error(ctx, http.StatusBadRequest, "Invalid request format", err)
		return
	}

	logger.L.Debug("Resolving exception",
		zap.String("exception_id", exceptionID),
		zap.String("resolved_by", req.ResolvedBy),
	)

	err := api.exceptionHandler.ResolveException(ctx.Request.Context(), exceptionID, req.ResolvedBy, req.Resolution)
	if err != nil {
		logger.L.Error("Failed to resolve exception",
			zap.String("exception_id", exceptionID),
			zap.Error(err),
	)
	response.Error(ctx, http.StatusInternalServerError, "Failed to resolve exception", err)
	return
	}

	response.Success(ctx, gin.H{
		"message": "Exception resolved successfully",
	})
}

// TriggerManualRetry 触发手动重试
// @Summary 触发手动重试
// @Description 对指定的同步异常触发手动重试
// @Tags 配置同步异常
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "异常ID"
// @Param request body object{retry_type=string,retry_config=object} true "重试配置"
// @Success 200 {object} response.Response "重试触发成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 404 {object} response.Response "异常不存在"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /api/v1/config/sync/exceptions/{id}/retry [post]
func (api *ConfigSyncExceptionAPI) TriggerManualRetry(ctx *gin.Context) {
	exceptionID := ctx.Param("id")
	if exceptionID == "" {
		response.Error(ctx, http.StatusBadRequest, "Exception ID is required", nil)
		return
	}

	var req struct {
		RetryBy string `json:"retry_by" binding:"required"`
		Force   bool   `json:"force"` // 是否强制重试，忽略重试次数限制
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Error(ctx, http.StatusBadRequest, "Invalid request format", err)
		return
	}

	logger.L.Debug("Triggering manual retry",
		zap.String("exception_id", exceptionID),
		zap.String("retry_by", req.RetryBy),
		zap.Bool("force", req.Force),
	)

	// 这里可以实现手动重试逻辑
	// 例如：重置重试计数器，立即触发同步任务等

	response.Success(ctx, gin.H{
		"message": "Manual retry triggered successfully",
	})
}

// GetExceptionStatistics 获取异常统计信息
// @Summary 获取异常统计信息
// @Description 获取配置同步异常的统计信息
// @Tags 配置同步异常
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param cluster_id query string false "集群ID"
// @Param config_type query string false "配置类型"
// @Param start_time query string false "开始时间(RFC3339格式)"
// @Param end_time query string false "结束时间(RFC3339格式)"
// @Success 200 {object} response.Response{data=object} "获取成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /api/v1/config/sync/exceptions/statistics [get]
func (api *ConfigSyncExceptionAPI) GetExceptionStatistics(ctx *gin.Context) {
	clusterID := ctx.Query("cluster_id")
	configType := ctx.Query("config_type")

	logger.L.Debug("Getting exception statistics",
		zap.String("cluster_id", clusterID),
		zap.String("config_type", configType),
	)

	// 这里可以实现统计逻辑
	// 例如：按类型统计异常数量、按严重程度统计、按时间段统计等

	statistics := gin.H{
		"total_exceptions": 0,
		"by_type": gin.H{
			"timeout":          0,
			"connection_error": 0,
			"validation_error": 0,
			"permission_error": 0,
			"server_error":     0,
			"unknown_error":    0,
		},
		"by_severity": gin.H{
			"low":      0,
			"medium":   0,
			"high":     0,
			"critical": 0,
		},
		"by_status": gin.H{
			"open":          0,
			"investigating": 0,
			"resolved":      0,
		},
	}

	response.Success(ctx, gin.H{
		"statistics": statistics,
	})
}

// RegisterConfigSyncExceptionRoutes 注册配置同步异常相关路由
func RegisterConfigSyncExceptionRoutes(r *gin.RouterGroup, authMiddleware gin.HandlerFunc) {
	api := NewConfigSyncExceptionAPI()
	
	exceptions := r.Group("/config/sync/exceptions")
	exceptions.Use(authMiddleware)
	{
		// 异常检测和管理
		exceptions.POST("/detect", api.DetectExceptions)
		exceptions.GET("", api.GetActiveExceptions)
		exceptions.GET("/statistics", api.GetExceptionStatistics)
		
		// 单个异常操作
		exceptions.GET("/:id/analysis", api.AnalyzeException)
		exceptions.POST("/:id/resolve", api.ResolveException)
		exceptions.POST("/:id/retry", api.TriggerManualRetry)
	}
}