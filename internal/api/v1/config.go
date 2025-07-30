package v1

import (
	"crypto/sha256"
	"fmt"
	"net/http"
	"time"

	"alert_agent/internal/model"
	"alert_agent/internal/pkg/database"
	"alert_agent/internal/pkg/logger"
	"alert_agent/internal/pkg/response"
	"alert_agent/internal/service"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// ConfigAPI 配置API处理器
type ConfigAPI struct {
	configService *service.ConfigService
}

// NewConfigAPI 创建配置API处理器
func NewConfigAPI() *ConfigAPI {
	return &ConfigAPI{
		configService: service.NewConfigService(),
	}
}

// GetSyncConfig Sidecar配置拉取接口
// GET /api/v1/config/sync
func (c *ConfigAPI) GetSyncConfig(ctx *gin.Context) {
	clusterID := ctx.Query("cluster_id")
	configType := ctx.Query("type") // prometheus, alertmanager, vmalert

	if clusterID == "" || configType == "" {
		response.Error(ctx, http.StatusBadRequest, "cluster_id and type are required")
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
		response.Error(ctx, http.StatusInternalServerError, "Failed to get config")
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
// POST /api/v1/config/sync/status
func (c *ConfigAPI) UpdateSyncStatus(ctx *gin.Context) {
	var req struct {
		ClusterID    string `json:"cluster_id" binding:"required"`
		ConfigType   string `json:"config_type" binding:"required"`
		Status       string `json:"status" binding:"required"`
		SyncTime     int64  `json:"sync_time" binding:"required"`
		ErrorMessage string `json:"error_message"`
		ConfigHash   string `json:"config_hash"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Error(ctx, http.StatusBadRequest, "Invalid request format")
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
		response.Error(ctx, http.StatusInternalServerError, "Failed to update sync status")
		return
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
// GET /api/v1/config/sync/status
func (c *ConfigAPI) GetSyncStatus(ctx *gin.Context) {
	clusterID := ctx.Query("cluster_id")
	configType := ctx.Query("type")

	if clusterID == "" {
		response.Error(ctx, http.StatusBadRequest, "cluster_id is required")
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
		response.Error(ctx, http.StatusInternalServerError, "Failed to get sync status")
		return
	}

	response.Success(ctx, gin.H{
		"statuses": statuses,
	})
}

// ListClusters 列出所有集群
// GET /api/v1/config/clusters
func (c *ConfigAPI) ListClusters(ctx *gin.Context) {
	clusters, err := c.configService.ListClusters(ctx.Request.Context())
	if err != nil {
		logger.L.Error("Failed to list clusters", zap.Error(err))
		response.Error(ctx, http.StatusInternalServerError, "Failed to list clusters")
		return
	}

	response.Success(ctx, gin.H{
		"clusters": clusters,
	})
}

// TriggerSync 触发配置同步
// POST /api/v1/config/sync/trigger
func (c *ConfigAPI) TriggerSync(ctx *gin.Context) {
	var req struct {
		ClusterID  string `json:"cluster_id" binding:"required"`
		ConfigType string `json:"config_type"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Error(ctx, http.StatusBadRequest, "Invalid request format")
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
		response.Error(ctx, http.StatusInternalServerError, "Failed to trigger sync")
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