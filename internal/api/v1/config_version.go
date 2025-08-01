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

// ConfigVersionAPI 配置版本API处理器
type ConfigVersionAPI struct {
	versionManager *service.ConfigVersionManager
}

// NewConfigVersionAPI 创建配置版本API处理器
func NewConfigVersionAPI() *ConfigVersionAPI {
	return &ConfigVersionAPI{
		versionManager: service.NewConfigVersionManager(),
	}
}

// CreateVersion 创建配置版本
// @Summary 创建配置版本
// @Description 为指定集群和配置类型创建新的配置版本
// @Tags 配置版本管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body object{cluster_id=string,config_type=string,description=string,created_by=string} true "创建版本请求"
// @Success 200 {object} response.Response{data=object} "创建成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /api/v1/config/versions [post]
func (api *ConfigVersionAPI) CreateVersion(ctx *gin.Context) {
	var req struct {
		ClusterID   string `json:"cluster_id" binding:"required"`
		ConfigType  string `json:"config_type" binding:"required"`
		Description string `json:"description"`
		CreatedBy   string `json:"created_by" binding:"required"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Error(ctx, http.StatusBadRequest, "Invalid request format", err)
		return
	}

	logger.L.Debug("Creating config version",
		zap.String("cluster_id", req.ClusterID),
		zap.String("config_type", req.ConfigType),
		zap.String("created_by", req.CreatedBy),
	)

	version, err := api.versionManager.CreateVersion(
		ctx.Request.Context(),
		req.ClusterID,
		req.ConfigType,
		req.Description,
		req.CreatedBy,
	)
	if err != nil {
		logger.L.Error("Failed to create config version",
			zap.String("cluster_id", req.ClusterID),
			zap.String("config_type", req.ConfigType),
			zap.Error(err),
		)
		response.Error(ctx, http.StatusInternalServerError, "Failed to create config version", err)
		return
	}

	response.Success(ctx, gin.H{
		"version": version,
	})
}

// GetVersions 获取配置版本列表
// @Summary 获取配置版本列表
// @Description 根据集群ID和配置类型获取配置版本列表
// @Tags 配置版本管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param cluster_id query string true "集群ID"
// @Param config_type query string true "配置类型"
// @Param limit query int false "限制数量" default(50)
// @Success 200 {object} response.Response{data=object} "获取成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /api/v1/config/versions [get]
func (api *ConfigVersionAPI) GetVersions(ctx *gin.Context) {
	clusterID := ctx.Query("cluster_id")
	configType := ctx.Query("config_type")
	limitStr := ctx.DefaultQuery("limit", "50")

	if clusterID == "" || configType == "" {
		response.Error(ctx, http.StatusBadRequest, "cluster_id and config_type are required", nil)
		return
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, "Invalid limit parameter", err)
		return
	}

	logger.L.Debug("Getting config versions",
		zap.String("cluster_id", clusterID),
		zap.String("config_type", configType),
		zap.Int("limit", limit),
	)

	versions, err := api.versionManager.GetVersions(ctx.Request.Context(), clusterID, configType, limit)
	if err != nil {
		logger.L.Error("Failed to get config versions",
			zap.String("cluster_id", clusterID),
			zap.String("config_type", configType),
			zap.Error(err),
		)
		response.Error(ctx, http.StatusInternalServerError, "Failed to get config versions", err)
		return
	}

	response.Success(ctx, gin.H{
		"versions": versions,
	})
}

// GetVersion 获取指定版本
// @Summary 获取指定版本
// @Description 根据版本ID获取指定的配置版本信息
// @Tags 配置版本管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "版本ID"
// @Success 200 {object} response.Response{data=object} "获取成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /api/v1/config/versions/{id} [get]
func (api *ConfigVersionAPI) GetVersion(ctx *gin.Context) {
	versionID := ctx.Param("id")

	if versionID == "" {
		response.Error(ctx, http.StatusBadRequest, "Version ID is required", nil)
		return
	}

	logger.L.Debug("Getting config version",
		zap.String("version_id", versionID),
	)

	version, err := api.versionManager.GetVersion(ctx.Request.Context(), versionID)
	if err != nil {
		logger.L.Error("Failed to get config version",
			zap.String("version_id", versionID),
			zap.Error(err),
		)
		response.Error(ctx, http.StatusInternalServerError, "Failed to get config version", err)
		return
	}

	response.Success(ctx, gin.H{
		"version": version,
	})
}

// CompareVersions 比较版本差异
// @Summary 比较版本差异
// @Description 比较两个配置版本之间的差异
// @Tags 配置版本管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param from query string true "源版本ID"
// @Param to query string true "目标版本ID"
// @Success 200 {object} response.Response{data=object} "比较成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /api/v1/config/versions/compare [get]
func (api *ConfigVersionAPI) CompareVersions(ctx *gin.Context) {
	fromVersionID := ctx.Query("from")
	toVersionID := ctx.Query("to")

	if fromVersionID == "" || toVersionID == "" {
		response.Error(ctx, http.StatusBadRequest, "from and to version IDs are required", nil)
		return
	}

	logger.L.Debug("Comparing config versions",
		zap.String("from_version_id", fromVersionID),
		zap.String("to_version_id", toVersionID),
	)

	diff, err := api.versionManager.CompareVersions(ctx.Request.Context(), fromVersionID, toVersionID)
	if err != nil {
		logger.L.Error("Failed to compare config versions",
			zap.String("from_version_id", fromVersionID),
			zap.String("to_version_id", toVersionID),
			zap.Error(err),
		)
		response.Error(ctx, http.StatusInternalServerError, "Failed to compare config versions", err)
		return
	}

	response.Success(ctx, gin.H{
		"diff": diff,
	})
}

// RollbackToVersion 回滚到指定版本
// @Summary 回滚到指定版本
// @Description 将配置回滚到指定的版本
// @Tags 配置版本管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "版本ID"
// @Param request body object{rollback_by=string} true "回滚请求"
// @Success 200 {object} response.Response "回滚成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /api/v1/config/versions/{id}/rollback [post]
func (api *ConfigVersionAPI) RollbackToVersion(ctx *gin.Context) {
	versionID := ctx.Param("id")
	if versionID == "" {
		response.Error(ctx, http.StatusBadRequest, "Version ID is required", nil)
		return
	}

	var req struct {
		RollbackBy string `json:"rollback_by" binding:"required"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Error(ctx, http.StatusBadRequest, "Invalid request format", err)
		return
	}

	logger.L.Debug("Rolling back to version",
		zap.String("version_id", versionID),
		zap.String("rollback_by", req.RollbackBy),
	)

	err := api.versionManager.RollbackToVersion(ctx.Request.Context(), versionID, req.RollbackBy)
	if err != nil {
		logger.L.Error("Failed to rollback to version",
			zap.String("version_id", versionID),
			zap.Error(err),
		)
		response.Error(ctx, http.StatusInternalServerError, "Failed to rollback to version", err)
		return
	}

	response.Success(ctx, gin.H{
		"message": "Rollback completed successfully",
	})
}

// CheckConsistency 检查配置一致性
// GET /api/v1/config/versions/consistency
func (api *ConfigVersionAPI) CheckConsistency(ctx *gin.Context) {
	clusterID := ctx.Query("cluster_id")
	configType := ctx.Query("config_type")

	if clusterID == "" || configType == "" {
		response.Error(ctx, http.StatusBadRequest, "cluster_id and config_type are required", nil)
		return
	}

	logger.L.Debug("Checking config consistency",
		zap.String("cluster_id", clusterID),
		zap.String("config_type", configType),
	)

	check, err := api.versionManager.CheckConsistency(ctx.Request.Context(), clusterID, configType)
	if err != nil {
		logger.L.Error("Failed to check config consistency",
			zap.String("cluster_id", clusterID),
			zap.String("config_type", configType),
			zap.Error(err),
		)
		response.Error(ctx, http.StatusInternalServerError, "Failed to check config consistency", err)
		return
	}

	response.Success(ctx, gin.H{
		"consistency_check": check,
	})
}

// GetActiveVersion 获取活跃版本
// GET /api/v1/config/versions/active
func (api *ConfigVersionAPI) GetActiveVersion(ctx *gin.Context) {
	clusterID := ctx.Query("cluster_id")
	configType := ctx.Query("config_type")

	if clusterID == "" || configType == "" {
		response.Error(ctx, http.StatusBadRequest, "cluster_id and config_type are required", nil)
		return
	}

	logger.L.Debug("Getting active config version",
		zap.String("cluster_id", clusterID),
		zap.String("config_type", configType),
	)

	version, err := api.versionManager.GetActiveVersion(ctx.Request.Context(), clusterID, configType)
	if err != nil {
		logger.L.Error("Failed to get active config version",
			zap.String("cluster_id", clusterID),
			zap.String("config_type", configType),
			zap.Error(err),
		)
		response.Error(ctx, http.StatusInternalServerError, "Failed to get active config version", err)
		return
	}

	response.Success(ctx, gin.H{
		"active_version": version,
	})
}

// DeleteVersion 删除版本
// @Summary 删除版本
// @Description 删除指定的配置版本
// @Tags 配置版本管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "版本ID"
// @Param request body object{deleted_by=string} true "删除请求"
// @Success 200 {object} response.Response "删除成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /api/v1/config/versions/{id} [delete]
func (api *ConfigVersionAPI) DeleteVersion(ctx *gin.Context) {
	versionID := ctx.Param("id")
	if versionID == "" {
		response.Error(ctx, http.StatusBadRequest, "Version ID is required", nil)
		return
	}

	var req struct {
		DeletedBy string `json:"deleted_by" binding:"required"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Error(ctx, http.StatusBadRequest, "Invalid request format", err)
		return
	}

	logger.L.Debug("Deleting config version",
		zap.String("version_id", versionID),
		zap.String("deleted_by", req.DeletedBy),
	)

	err := api.versionManager.DeleteVersion(ctx.Request.Context(), versionID, req.DeletedBy)
	if err != nil {
		logger.L.Error("Failed to delete config version",
			zap.String("version_id", versionID),
			zap.Error(err),
		)
		response.Error(ctx, http.StatusInternalServerError, "Failed to delete config version", err)
		return
	}

	response.Success(ctx, gin.H{
		"message": "Version deleted successfully",
	})
}

// CleanupOldVersions 清理旧版本
// POST /api/v1/config/versions/cleanup
func (api *ConfigVersionAPI) CleanupOldVersions(ctx *gin.Context) {
	var req struct {
		ClusterID  string `json:"cluster_id" binding:"required"`
		ConfigType string `json:"config_type" binding:"required"`
		KeepCount  int    `json:"keep_count"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Error(ctx, http.StatusBadRequest, "Invalid request format", err)
		return
	}

	if req.KeepCount <= 0 {
		req.KeepCount = 10
	}

	logger.L.Debug("Cleaning up old config versions",
		zap.String("cluster_id", req.ClusterID),
		zap.String("config_type", req.ConfigType),
		zap.Int("keep_count", req.KeepCount),
	)

	err := api.versionManager.CleanupOldVersions(
		ctx.Request.Context(),
		req.ClusterID,
		req.ConfigType,
		req.KeepCount,
	)
	if err != nil {
		logger.L.Error("Failed to cleanup old config versions",
			zap.String("cluster_id", req.ClusterID),
			zap.String("config_type", req.ConfigType),
			zap.Error(err),
		)
		response.Error(ctx, http.StatusInternalServerError, "Failed to cleanup old config versions", err)
		return
	}

	response.Success(ctx, gin.H{
		"message": "Old versions cleaned up successfully",
	})
}

// RegisterConfigVersionRoutes 注册配置版本相关路由
func RegisterConfigVersionRoutes(r *gin.RouterGroup, authMiddleware gin.HandlerFunc) {
	api := NewConfigVersionAPI()
	
	versions := r.Group("/config/versions")
	versions.Use(authMiddleware)
	{
		// 版本管理
		versions.POST("", api.CreateVersion)
		versions.GET("", api.GetVersions)
		versions.GET("/active", api.GetActiveVersion)
		versions.GET("/compare", api.CompareVersions)
		versions.GET("/consistency", api.CheckConsistency)
		versions.POST("/cleanup", api.CleanupOldVersions)
		
		// 单个版本操作
		versions.GET("/:id", api.GetVersion)
		versions.POST("/:id/rollback", api.RollbackToVersion)
		versions.DELETE("/:id", api.DeleteVersion)
	}
}