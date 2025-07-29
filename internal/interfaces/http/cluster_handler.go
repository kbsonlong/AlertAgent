package http

import (
	"net/http"
	"strconv"

	"alert_agent/internal/domain/cluster"
	"alert_agent/internal/shared/errors"
	"alert_agent/pkg/types"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// ClusterHandler 集群HTTP处理器
type ClusterHandler struct {
	service cluster.Service
	logger  *zap.Logger
}

// NewClusterHandler 创建集群处理器
func NewClusterHandler(service cluster.Service, logger *zap.Logger) *ClusterHandler {
	return &ClusterHandler{
		service: service,
		logger:  logger,
	}
}

// CreateCluster 创建集群
// @Summary 创建集群
// @Description 创建新的集群
// @Tags clusters
// @Accept json
// @Produce json
// @Param cluster body cluster.CreateClusterRequest true "集群信息"
// @Success 201 {object} types.APIResponse{data=cluster.Cluster}
// @Failure 400 {object} types.APIResponse
// @Failure 500 {object} types.APIResponse
// @Router /api/v1/clusters [post]
func (h *ClusterHandler) CreateCluster(c *gin.Context) {
	var req cluster.CreateClusterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("invalid request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, types.APIResponse{
			Status:  "error",
			Message: "Invalid request body",
			Error: &types.ErrorInfo{
				Type:    "validation",
				Code:    "INVALID_REQUEST",
				Message: err.Error(),
			},
		})
		return
	}

	clusterEntity, err := h.service.CreateCluster(c.Request.Context(), &req)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, types.APIResponse{
		Status:  "success",
		Message: "Cluster created successfully",
		Data:    clusterEntity,
	})
}

// GetCluster 获取集群详情
// @Summary 获取集群详情
// @Description 根据ID获取集群详情
// @Tags clusters
// @Produce json
// @Param id path string true "集群ID"
// @Success 200 {object} types.APIResponse{data=cluster.Cluster}
// @Failure 404 {object} types.APIResponse
// @Failure 500 {object} types.APIResponse
// @Router /api/v1/clusters/{id} [get]
func (h *ClusterHandler) GetCluster(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, types.APIResponse{
			Status:  "error",
			Message: "Cluster ID is required",
			Error: &types.ErrorInfo{
				Type:    "validation",
				Code:    "INVALID_ID",
				Message: "Cluster ID is required",
			},
		})
		return
	}

	clusterEntity, err := h.service.GetCluster(c.Request.Context(), id)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, types.APIResponse{
		Status:  "success",
		Message: "Cluster retrieved successfully",
		Data:    clusterEntity,
	})
}

// ListClusters 获取集群列表
// @Summary 获取集群列表
// @Description 获取集群列表，支持分页和过滤
// @Tags clusters
// @Produce json
// @Param limit query int false "每页数量" default(10)
// @Param offset query int false "偏移量" default(0)
// @Param search query string false "搜索关键词"
// @Param type query string false "集群类型"
// @Param status query string false "集群状态"
// @Success 200 {object} types.APIResponse{data=types.PageResult}
// @Failure 500 {object} types.APIResponse
// @Router /api/v1/clusters [get]
func (h *ClusterHandler) ListClusters(c *gin.Context) {
	// 解析查询参数
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	search := c.Query("search")
	clusterType := c.Query("type")
	status := c.Query("status")

	query := types.Query{
		Limit:  limit,
		Offset: offset,
		Search: search,
		Filter: make(map[string]interface{}),
	}

	if clusterType != "" {
		query.Filter["type"] = clusterType
	}
	if status != "" {
		query.Filter["status"] = status
	}

	result, err := h.service.ListClusters(c.Request.Context(), query)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, types.APIResponse{
		Status:  "success",
		Message: "Clusters retrieved successfully",
		Data:    result,
	})
}

// UpdateCluster 更新集群
// @Summary 更新集群
// @Description 更新集群信息
// @Tags clusters
// @Accept json
// @Produce json
// @Param id path string true "集群ID"
// @Param cluster body cluster.UpdateClusterRequest true "更新信息"
// @Success 200 {object} types.APIResponse{data=cluster.Cluster}
// @Failure 400 {object} types.APIResponse
// @Failure 404 {object} types.APIResponse
// @Failure 500 {object} types.APIResponse
// @Router /api/v1/clusters/{id} [put]
func (h *ClusterHandler) UpdateCluster(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, types.APIResponse{
			Status:  "error",
			Message: "Cluster ID is required",
			Error: &types.ErrorInfo{
				Type:    "validation",
				Code:    "INVALID_ID",
				Message: "Cluster ID is required",
			},
		})
		return
	}

	var req cluster.UpdateClusterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("invalid request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, types.APIResponse{
			Status:  "error",
			Message: "Invalid request body",
			Error: &types.ErrorInfo{
				Type:    "validation",
				Code:    "INVALID_REQUEST",
				Message: err.Error(),
			},
		})
		return
	}

	clusterEntity, err := h.service.UpdateCluster(c.Request.Context(), id, &req)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, types.APIResponse{
		Status:  "success",
		Message: "Cluster updated successfully",
		Data:    clusterEntity,
	})
}

// DeleteCluster 删除集群
// @Summary 删除集群
// @Description 删除指定的集群
// @Tags clusters
// @Produce json
// @Param id path string true "集群ID"
// @Success 200 {object} types.APIResponse
// @Failure 404 {object} types.APIResponse
// @Failure 500 {object} types.APIResponse
// @Router /api/v1/clusters/{id} [delete]
func (h *ClusterHandler) DeleteCluster(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, types.APIResponse{
			Status:  "error",
			Message: "Cluster ID is required",
			Error: &types.ErrorInfo{
				Type:    "validation",
				Code:    "INVALID_ID",
				Message: "Cluster ID is required",
			},
		})
		return
	}

	err := h.service.DeleteCluster(c.Request.Context(), id)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, types.APIResponse{
		Status:  "success",
		Message: "Cluster deleted successfully",
	})
}

// TestClusterConnection 测试集群连接
// @Summary 测试集群连接
// @Description 测试指定集群的连接状态
// @Tags clusters
// @Produce json
// @Param id path string true "集群ID"
// @Success 200 {object} types.APIResponse{data=cluster.ConnectionTestResult}
// @Failure 404 {object} types.APIResponse
// @Failure 500 {object} types.APIResponse
// @Router /api/v1/clusters/{id}/test [post]
func (h *ClusterHandler) TestClusterConnection(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, types.APIResponse{
			Status:  "error",
			Message: "Cluster ID is required",
			Error: &types.ErrorInfo{
				Type:    "validation",
				Code:    "INVALID_ID",
				Message: "Cluster ID is required",
			},
		})
		return
	}

	result, err := h.service.TestClusterConnection(c.Request.Context(), id)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, types.APIResponse{
		Status:  "success",
		Message: "Connection test completed",
		Data:    result,
	})
}

// GetClusterHealth 获取集群健康状态
// @Summary 获取集群健康状态
// @Description 获取指定集群的健康状态
// @Tags clusters
// @Produce json
// @Param id path string true "集群ID"
// @Success 200 {object} types.APIResponse{data=cluster.ClusterHealth}
// @Failure 404 {object} types.APIResponse
// @Failure 500 {object} types.APIResponse
// @Router /api/v1/clusters/{id}/health [get]
func (h *ClusterHandler) GetClusterHealth(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, types.APIResponse{
			Status:  "error",
			Message: "Cluster ID is required",
			Error: &types.ErrorInfo{
				Type:    "validation",
				Code:    "INVALID_ID",
				Message: "Cluster ID is required",
			},
		})
		return
	}

	health, err := h.service.GetClusterHealth(c.Request.Context(), id)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, types.APIResponse{
		Status:  "success",
		Message: "Cluster health retrieved successfully",
		Data:    health,
	})
}

// GetClusterMetrics 获取集群指标
// @Summary 获取集群指标
// @Description 获取指定集群的性能指标
// @Tags clusters
// @Produce json
// @Param id path string true "集群ID"
// @Success 200 {object} types.APIResponse{data=cluster.ClusterMetrics}
// @Failure 404 {object} types.APIResponse
// @Failure 500 {object} types.APIResponse
// @Router /api/v1/clusters/{id}/metrics [get]
func (h *ClusterHandler) GetClusterMetrics(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, types.APIResponse{
			Status:  "error",
			Message: "Cluster ID is required",
			Error: &types.ErrorInfo{
				Type:    "validation",
				Code:    "INVALID_ID",
				Message: "Cluster ID is required",
			},
		})
		return
	}

	metrics, err := h.service.GetClusterMetrics(c.Request.Context(), id)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, types.APIResponse{
		Status:  "success",
		Message: "Cluster metrics retrieved successfully",
		Data:    metrics,
	})
}

// handleError 统一错误处理
func (h *ClusterHandler) handleError(c *gin.Context, err error) {
	h.logger.Error("request failed", zap.Error(err))

	if appErr, ok := err.(*errors.AppError); ok {
		status := errors.GetHTTPStatusCode(appErr)
		c.JSON(status, types.APIResponse{
			Status:  "error",
			Message: appErr.Message,
			Error: &types.ErrorInfo{
				Type:    string(appErr.Type),
				Code:    appErr.Code,
				Message: appErr.Message,
				Details: appErr.Details,
			},
		})
		return
	}

	// 未知错误
	c.JSON(http.StatusInternalServerError, types.APIResponse{
		Status:  "error",
		Message: "Internal server error",
		Error: &types.ErrorInfo{
			Type:    "internal",
			Code:    "INTERNAL_ERROR",
			Message: "An unexpected error occurred",
		},
	})
}