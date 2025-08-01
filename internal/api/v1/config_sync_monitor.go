package v1

import (
	"net/http"
	"strconv"

	"alert_agent/internal/pkg/logger"
	"alert_agent/internal/pkg/response"
	"alert_agent/internal/service"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// ConfigSyncMonitorAPI 配置同步监控API处理器
type ConfigSyncMonitorAPI struct {
	monitorService *service.ConfigSyncMonitor
}

// NewConfigSyncMonitorAPI 创建配置同步监控API处理器
func NewConfigSyncMonitorAPI() *ConfigSyncMonitorAPI {
	return &ConfigSyncMonitorAPI{
		monitorService: service.NewConfigSyncMonitor(),
	}
}

// GetSyncMetrics 获取同步指标
// @Summary 获取同步指标
// @Description 获取配置同步的各项性能指标
// @Tags 配置同步监控
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.Response{data=object} "获取成功"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /api/v1/config/sync/metrics [get]
func (api *ConfigSyncMonitorAPI) GetSyncMetrics(ctx *gin.Context) {
	logger.L.Debug("Getting sync metrics")

	metrics, err := api.monitorService.CollectSyncMetrics(ctx.Request.Context())
	if err != nil {
		logger.L.Error("Failed to collect sync metrics", zap.Error(err))
	response.Error(ctx, http.StatusInternalServerError, "Failed to collect sync metrics", err)
	return
	}

	response.Success(ctx, gin.H{
		"metrics": metrics,
	})
}

// GetSyncDelayMetrics 获取同步延迟指标
// @Summary 获取同步延迟指标
// @Description 获取配置同步的延迟相关指标
// @Tags 配置同步监控
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param cluster_id query string false "集群ID"
// @Param config_type query string false "配置类型"
// @Param hours query int false "时间范围(小时)" default(24)
// @Success 200 {object} response.Response{data=object} "获取成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /api/v1/config/sync/metrics/delay [get]
func (api *ConfigSyncMonitorAPI) GetSyncDelayMetrics(ctx *gin.Context) {
	clusterID := ctx.Query("cluster_id")
	configType := ctx.Query("config_type")
	hoursStr := ctx.DefaultQuery("hours", "24")

	hours, err := strconv.Atoi(hoursStr)
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, "Invalid hours parameter", err)
		return
	}

	logger.L.Debug("Getting sync delay metrics",
		zap.String("cluster_id", clusterID),
		zap.String("config_type", configType),
		zap.Int("hours", hours),
	)

	metrics, err := api.monitorService.GetSyncDelayMetrics(ctx.Request.Context(), clusterID, configType, hours)
	if err != nil {
		logger.L.Error("Failed to get sync delay metrics",
			zap.String("cluster_id", clusterID),
			zap.String("config_type", configType),
			zap.Error(err),
		)
		response.Error(ctx, http.StatusInternalServerError, "Failed to get sync delay metrics", err)
		return
	}

	response.Success(ctx, gin.H{
		"delay_metrics": metrics,
	})
}

// GetFailureRateMetrics 获取失败率指标
// @Summary 获取失败率指标
// @Description 获取配置同步的失败率相关指标
// @Tags 配置同步监控
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param cluster_id query string false "集群ID"
// @Param config_type query string false "配置类型"
// @Param hours query int false "时间范围(小时)" default(24)
// @Success 200 {object} response.Response{data=object} "获取成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /api/v1/config/sync/metrics/failure-rate [get]
func (api *ConfigSyncMonitorAPI) GetFailureRateMetrics(ctx *gin.Context) {
	clusterID := ctx.Query("cluster_id")
	configType := ctx.Query("config_type")
	hoursStr := ctx.DefaultQuery("hours", "24")

	hours, err := strconv.Atoi(hoursStr)
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, "Invalid hours parameter", err)
		return
	}

	logger.L.Debug("Getting failure rate metrics",
		zap.String("cluster_id", clusterID),
		zap.String("config_type", configType),
		zap.Int("hours", hours),
	)

	metrics, err := api.monitorService.GetFailureRateMetrics(ctx.Request.Context(), clusterID, configType, hours)
	if err != nil {
		logger.L.Error("Failed to get failure rate metrics",
			zap.String("cluster_id", clusterID),
			zap.String("config_type", configType),
			zap.Error(err),
		)
		response.Error(ctx, http.StatusInternalServerError, "Failed to get failure rate metrics", err)
		return
	}

	response.Success(ctx, gin.H{
		"failure_rate_metrics": metrics,
	})
}

// RecordSyncHistory 记录同步历史
// @Summary 记录同步历史
// @Description 记录配置同步的历史信息
// @Tags 配置同步监控
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body object{cluster_id=string,config_type=string,config_hash=string,config_size=int64,sync_status=string,sync_duration_ms=int64,error_message=string} true "同步历史记录"
// @Success 200 {object} response.Response "记录成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /api/v1/config/sync/history [post]
func (api *ConfigSyncMonitorAPI) RecordSyncHistory(ctx *gin.Context) {
	var req struct {
		ClusterID    string `json:"cluster_id" binding:"required"`
		ConfigType   string `json:"config_type" binding:"required"`
		ConfigHash   string `json:"config_hash" binding:"required"`
		ConfigSize   int64  `json:"config_size"`
		SyncStatus   string `json:"sync_status" binding:"required"`
		SyncDuration int64  `json:"sync_duration_ms"`
		ErrorMessage string `json:"error_message"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Error(ctx, http.StatusBadRequest, "Invalid request format", err)
		return
	}

	logger.L.Debug("Recording sync history",
		zap.String("cluster_id", req.ClusterID),
		zap.String("config_type", req.ConfigType),
		zap.String("status", req.SyncStatus),
	)

	err := api.monitorService.RecordSyncHistory(
		ctx.Request.Context(),
		req.ClusterID,
		req.ConfigType,
		req.ConfigHash,
		req.ConfigSize,
		req.SyncStatus,
		req.SyncDuration,
		req.ErrorMessage,
	)

	if err != nil {
		logger.L.Error("Failed to record sync history",
			zap.String("cluster_id", req.ClusterID),
			zap.String("config_type", req.ConfigType),
			zap.Error(err),
		)
		response.Error(ctx, http.StatusInternalServerError, "Failed to record sync history", err)
		return
	}

	response.Success(ctx, gin.H{
		"message": "Sync history recorded successfully",
	})
}

// CleanupOldHistory 清理旧历史记录
// @Summary 清理旧历史记录
// @Description 清理过期的配置同步历史记录
// @Tags 配置同步监控
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param days query int false "保留天数" default(30)
// @Success 200 {object} response.Response{data=object} "清理成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /api/v1/config/sync/history/cleanup [delete]
func (api *ConfigSyncMonitorAPI) CleanupOldHistory(ctx *gin.Context) {
	retentionDaysStr := ctx.DefaultQuery("retention_days", "30")
	retentionDays, err := strconv.Atoi(retentionDaysStr)
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, "Invalid retention_days parameter", err)
		return
	}

	logger.L.Debug("Cleaning up old sync history",
		zap.Int("retention_days", retentionDays),
	)

	err = api.monitorService.CleanupOldHistory(ctx.Request.Context(), retentionDays)
	if err != nil {
		logger.L.Error("Failed to cleanup old history",
			zap.Int("retention_days", retentionDays),
			zap.Error(err),
		)
		response.Error(ctx, http.StatusInternalServerError, "Failed to cleanup old history", err)
		return
	}

	response.Success(ctx, gin.H{
		"message": "Old history cleaned up successfully",
	})
}

// RegisterConfigSyncMonitorRoutes 注册配置同步监控相关路由
func RegisterConfigSyncMonitorRoutes(r *gin.RouterGroup, authMiddleware gin.HandlerFunc) {
	api := NewConfigSyncMonitorAPI()
	
	monitor := r.Group("/config/sync")
	monitor.Use(authMiddleware)
	{
		// 同步指标
		monitor.GET("/metrics", api.GetSyncMetrics)
		monitor.GET("/metrics/delay", api.GetSyncDelayMetrics)
		monitor.GET("/metrics/failure-rate", api.GetFailureRateMetrics)
		
		// 同步历史
		monitor.POST("/history", api.RecordSyncHistory)
		monitor.DELETE("/history/cleanup", api.CleanupOldHistory)
	}
}