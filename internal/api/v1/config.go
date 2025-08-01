package v1

import (
	"crypto/sha256"
	"fmt"
	"net/http"
	"time"

	"alert_agent/internal/pkg/logger"
	"alert_agent/internal/pkg/response"
	"alert_agent/internal/service"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// ConfigAPI 配置API处理器
type ConfigAPI struct {
	configService  *service.ConfigService
	monitorService *service.ConfigSyncMonitor
}

// NewConfigAPI 创建配置API处理器
func NewConfigAPI() *ConfigAPI {
	return &ConfigAPI{
		configService:  service.NewConfigService(),
		monitorService: service.NewConfigSyncMonitor(),
	}
}

// GetSyncConfig Sidecar配置拉取接口
// @Summary 获取同步配置
// @Description 获取指定集群和类型的配置内容，用于Sidecar配置同步
// @Tags 配置管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param cluster_id query string true "集群ID"
// @Param type query string true "配置类型" Enums(prometheus,alertmanager,vmalert)
// @Success 200 {object} response.Response{data=object}
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/config/sync [get]
func (c *ConfigAPI) GetSyncConfig(ctx *gin.Context) {
	clusterID := ctx.Query("cluster_id")
	configType := ctx.Query("type") // prometheus, alertmanager, vmalert

	if clusterID == "" || configType == "" {
		response.Error(ctx, http.StatusBadRequest, "cluster_id and type are required", nil)
		return
	}

	logger.L.Debug("Fetching sync config",
		zap.String("cluster_id", clusterID),
		zap.String("config_type", configType),
	)

	// 获取配置内容
	config, err := c.configService.GetConfig(ctx.Request.Context(), clusterID, configType)
	if err != nil {
		logger.L.Error("Failed to get config",
			zap.String("cluster_id", clusterID),
			zap.String("config_type", configType),
		zap.Error(err),
	)
	response.Error(ctx, http.StatusInternalServerError, "Failed to get config", err)
	return
	}

	// 计算配置hash
	hash := sha256.Sum256([]byte(config))
	configHash := fmt.Sprintf("%x", hash)

	// 检查If-None-Match头，支持条件请求
	if ifNoneMatch := ctx.GetHeader("If-None-Match"); ifNoneMatch == configHash {
		ctx.Status(http.StatusNotModified)
		return
	}

	// 设置响应头
	ctx.Header("X-Config-Hash", configHash)
	ctx.Header("Content-Type", "application/yaml")
	ctx.Header("Cache-Control", "no-cache")

	// 返回配置内容
	ctx.String(http.StatusOK, config)

	logger.L.Debug("Config sent successfully",
		zap.String("cluster_id", clusterID),
		zap.String("config_type", configType),
		zap.String("hash", configHash),
		zap.Int("size", len(config)),
	)
}

// UpdateSyncStatus 更新同步状态接口
// @Summary 更新同步状态
// @Description 更新配置同步状态信息，由Sidecar调用
// @Tags 配置管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body object{cluster_id=string,config_type=string,status=string,sync_time=int64,error_message=string,config_hash=string,config_size=int64,sync_duration_ms=int64} true "同步状态信息"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/config/sync/status [post]
func (c *ConfigAPI) UpdateSyncStatus(ctx *gin.Context) {
	var req struct {
		ClusterID    string `json:"cluster_id" binding:"required"`
		ConfigType   string `json:"config_type" binding:"required"`
		Status       string `json:"status" binding:"required"`
		SyncTime     int64  `json:"sync_time" binding:"required"`
		ErrorMessage string `json:"error_message"`
		ConfigHash   string `json:"config_hash"`
		ConfigSize   int64  `json:"config_size"`
		SyncDuration int64  `json:"sync_duration_ms"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Error(ctx, http.StatusBadRequest, "Invalid request format", err)
		return
	}

	logger.L.Debug("Updating sync status",
		zap.String("cluster_id", req.ClusterID),
		zap.String("config_type", req.ConfigType),
		zap.String("status", req.Status),
	)

	// 更新同步状态
	err := c.configService.UpdateSyncStatus(ctx.Request.Context(), &service.SyncStatusUpdate{
		ClusterID:    req.ClusterID,
		ConfigType:   req.ConfigType,
		Status:       req.Status,
		SyncTime:     time.Unix(req.SyncTime, 0),
		ErrorMessage: req.ErrorMessage,
		ConfigHash:   req.ConfigHash,
	})

	if err != nil {
		logger.L.Error("Failed to update sync status",
			zap.String("cluster_id", req.ClusterID),
			zap.String("config_type", req.ConfigType),
			zap.Error(err),
		)
		response.Error(ctx, http.StatusInternalServerError, "Failed to update sync status", err)
		return
	}

	// 记录同步历史用于监控
	if req.SyncDuration > 0 {
		err = c.monitorService.RecordSyncHistory(
			ctx.Request.Context(),
			req.ClusterID,
			req.ConfigType,
			req.ConfigHash,
			req.ConfigSize,
			req.Status,
			req.SyncDuration,
			req.ErrorMessage,
		)
		if err != nil {
			logger.L.Error("Failed to record sync history",
				zap.String("cluster_id", req.ClusterID),
				zap.String("config_type", req.ConfigType),
				zap.Error(err),
			)
			// 不影响主流程，只记录错误
		}
	}

	response.Success(ctx, gin.H{
		"message": "Sync status updated successfully",
	})

	logger.L.Info("Sync status updated",
		zap.String("cluster_id", req.ClusterID),
		zap.String("config_type", req.ConfigType),
		zap.String("status", req.Status),
	)
}

// GetSyncStatus 获取同步状态接口
// @Summary 获取同步状态
// @Description 获取指定集群的配置同步状态信息
// @Tags 配置管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param cluster_id query string true "集群ID"
// @Param type query string false "配置类型" Enums(prometheus,alertmanager,vmalert)
// @Success 200 {object} response.Response{data=object}
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/config/sync/status [get]
func (c *ConfigAPI) GetSyncStatus(ctx *gin.Context) {
	clusterID := ctx.Query("cluster_id")
	configType := ctx.Query("type")

	if clusterID == "" {
		response.Error(ctx, http.StatusBadRequest, "cluster_id is required", nil)
		return
	}

	// 获取同步状态
	statuses, err := c.configService.GetSyncStatus(ctx.Request.Context(), clusterID, configType)
	if err != nil {
		logger.L.Error("Failed to get sync status",
			zap.String("cluster_id", clusterID),
			zap.String("config_type", configType),
			zap.Error(err),
		)
		response.Error(ctx, http.StatusInternalServerError, "Failed to get sync status", err)
		return
	}

	response.Success(ctx, gin.H{
		"statuses": statuses,
	})
}

// ListClusters 列出所有集群
// @Summary 获取集群列表
// @Description 获取系统中所有已注册的集群列表
// @Tags 配置管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.Response{data=[]object}
// @Failure 401 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/config/clusters [get]
func (c *ConfigAPI) ListClusters(ctx *gin.Context) {
	clusters, err := c.configService.ListClusters(ctx.Request.Context())
	if err != nil {
		logger.L.Error("Failed to list clusters", zap.Error(err))
		response.Error(ctx, http.StatusInternalServerError, "Failed to list clusters", err)
		return
	}

	response.Success(ctx, gin.H{
		"clusters": clusters,
	})
}

// TriggerSync 触发配置同步
// @Summary 触发配置同步
// @Description 手动触发指定集群的配置同步操作
// @Tags 配置管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body object{cluster_id=string,config_type=string} true "同步触发请求"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/config/sync/trigger [post]
func (c *ConfigAPI) TriggerSync(ctx *gin.Context) {
	var req struct {
		ClusterID  string `json:"cluster_id" binding:"required"`
		ConfigType string `json:"config_type"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Error(ctx, http.StatusBadRequest, "Invalid request format", err)
		return
	}

	// 触发同步（这里可以通过消息队列或其他方式通知Sidecar）
	err := c.configService.TriggerSync(ctx.Request.Context(), req.ClusterID, req.ConfigType)
	if err != nil {
		logger.L.Error("Failed to trigger sync",
			zap.String("cluster_id", req.ClusterID),
			zap.String("config_type", req.ConfigType),
			zap.Error(err),
		)
		response.Error(ctx, http.StatusInternalServerError, "Failed to trigger sync", err)
		return
	}

	response.Success(ctx, gin.H{
		"message": "Sync triggered successfully",
	})

	logger.L.Info("Sync triggered",
		zap.String("cluster_id", req.ClusterID),
		zap.String("config_type", req.ConfigType),
	)
}

// RegisterConfigRoutes 注册配置相关路由
func RegisterConfigRoutes(r *gin.RouterGroup, authMiddleware gin.HandlerFunc) {
	configAPI := NewConfigAPI()
	
	config := r.Group("/config")
	config.Use(authMiddleware)
	{
		// Sidecar配置同步接口
		config.GET("/sync", configAPI.GetSyncConfig)
		config.POST("/sync/status", configAPI.UpdateSyncStatus)
		config.GET("/sync/status", configAPI.GetSyncStatus)
		config.POST("/sync/trigger", configAPI.TriggerSync)
		
		// 集群管理
		config.GET("/clusters", configAPI.ListClusters)
	}
}