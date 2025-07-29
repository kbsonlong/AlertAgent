package http

import (
	"net/http"
	"strconv"

	"alert_agent/internal/domain/channel"
	"alert_agent/internal/shared/errors"
	"alert_agent/pkg/types"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// ChannelHandler 通道HTTP处理器
type ChannelHandler struct {
	service channel.Service
	logger  *zap.Logger
}

// NewChannelHandler 创建通道处理器
func NewChannelHandler(service channel.Service, logger *zap.Logger) *ChannelHandler {
	return &ChannelHandler{
		service: service,
		logger:  logger,
	}
}

// CreateChannel 创建通道
// @Summary 创建通道
// @Description 创建新的通知通道
// @Tags channels
// @Accept json
// @Produce json
// @Param channel body channel.CreateChannelRequest true "通道信息"
// @Success 201 {object} types.APIResponse{data=channel.Channel}
// @Failure 400 {object} types.APIResponse
// @Failure 500 {object} types.APIResponse
// @Router /api/v1/channels [post]
func (h *ChannelHandler) CreateChannel(c *gin.Context) {
	var req channel.CreateChannelRequest
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

	channelEntity, err := h.service.CreateChannel(c.Request.Context(), &req)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, types.APIResponse{
		Status:  "success",
		Message: "Channel created successfully",
		Data:    channelEntity,
	})
}

// GetChannel 获取通道详情
// @Summary 获取通道详情
// @Description 根据ID获取通道详情
// @Tags channels
// @Produce json
// @Param id path string true "通道ID"
// @Success 200 {object} types.APIResponse{data=channel.Channel}
// @Failure 404 {object} types.APIResponse
// @Failure 500 {object} types.APIResponse
// @Router /api/v1/channels/{id} [get]
func (h *ChannelHandler) GetChannel(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, types.APIResponse{
			Status:  "error",
			Message: "Channel ID is required",
			Error: &types.ErrorInfo{
				Type:    "validation",
				Code:    "INVALID_ID",
				Message: "Channel ID is required",
			},
		})
		return
	}

	channelEntity, err := h.service.GetChannel(c.Request.Context(), id)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, types.APIResponse{
		Status:  "success",
		Message: "Channel retrieved successfully",
		Data:    channelEntity,
	})
}

// ListChannels 获取通道列表
// @Summary 获取通道列表
// @Description 获取通道列表，支持分页和过滤
// @Tags channels
// @Produce json
// @Param limit query int false "每页数量" default(10)
// @Param offset query int false "偏移量" default(0)
// @Param search query string false "搜索关键词"
// @Param type query string false "通道类型"
// @Param status query string false "通道状态"
// @Success 200 {object} types.APIResponse{data=types.PageResult}
// @Failure 500 {object} types.APIResponse
// @Router /api/v1/channels [get]
func (h *ChannelHandler) ListChannels(c *gin.Context) {
	// 解析查询参数
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	search := c.Query("search")
	channelType := c.Query("type")
	status := c.Query("status")

	query := types.Query{
		Limit:  limit,
		Offset: offset,
		Search: search,
		Filter: make(map[string]interface{}),
	}

	if channelType != "" {
		query.Filter["type"] = channelType
	}
	if status != "" {
		query.Filter["status"] = status
	}

	result, err := h.service.ListChannels(c.Request.Context(), query)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, types.APIResponse{
		Status:  "success",
		Message: "Channels retrieved successfully",
		Data:    result,
	})
}

// UpdateChannel 更新通道
// @Summary 更新通道
// @Description 更新通道信息
// @Tags channels
// @Accept json
// @Produce json
// @Param id path string true "通道ID"
// @Param channel body channel.UpdateChannelRequest true "更新信息"
// @Success 200 {object} types.APIResponse{data=channel.Channel}
// @Failure 400 {object} types.APIResponse
// @Failure 404 {object} types.APIResponse
// @Failure 500 {object} types.APIResponse
// @Router /api/v1/channels/{id} [put]
func (h *ChannelHandler) UpdateChannel(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, types.APIResponse{
			Status:  "error",
			Message: "Channel ID is required",
			Error: &types.ErrorInfo{
				Type:    "validation",
				Code:    "INVALID_ID",
				Message: "Channel ID is required",
			},
		})
		return
	}

	var req channel.UpdateChannelRequest
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

	channelEntity, err := h.service.UpdateChannel(c.Request.Context(), id, &req)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, types.APIResponse{
		Status:  "success",
		Message: "Channel updated successfully",
		Data:    channelEntity,
	})
}

// DeleteChannel 删除通道
// @Summary 删除通道
// @Description 删除指定的通道
// @Tags channels
// @Produce json
// @Param id path string true "通道ID"
// @Success 200 {object} types.APIResponse
// @Failure 404 {object} types.APIResponse
// @Failure 500 {object} types.APIResponse
// @Router /api/v1/channels/{id} [delete]
func (h *ChannelHandler) DeleteChannel(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, types.APIResponse{
			Status:  "error",
			Message: "Channel ID is required",
			Error: &types.ErrorInfo{
				Type:    "validation",
				Code:    "INVALID_ID",
				Message: "Channel ID is required",
			},
		})
		return
	}

	err := h.service.DeleteChannel(c.Request.Context(), id)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, types.APIResponse{
		Status:  "success",
		Message: "Channel deleted successfully",
	})
}

// TestChannel 测试通道
// @Summary 测试通道
// @Description 测试指定通道的连接状态
// @Tags channels
// @Produce json
// @Param id path string true "通道ID"
// @Success 200 {object} types.APIResponse{data=channel.TestResult}
// @Failure 404 {object} types.APIResponse
// @Failure 500 {object} types.APIResponse
// @Router /api/v1/channels/{id}/test [post]
func (h *ChannelHandler) TestChannel(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, types.APIResponse{
			Status:  "error",
			Message: "Channel ID is required",
			Error: &types.ErrorInfo{
				Type:    "validation",
				Code:    "INVALID_ID",
				Message: "Channel ID is required",
			},
		})
		return
	}

	result, err := h.service.TestChannel(c.Request.Context(), id)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, types.APIResponse{
		Status:  "success",
		Message: "Channel test completed",
		Data:    result,
	})
}

// SendMessage 发送消息
// @Summary 发送消息
// @Description 通过指定通道发送消息
// @Tags channels
// @Accept json
// @Produce json
// @Param id path string true "通道ID"
// @Param message body types.Message true "消息内容"
// @Success 200 {object} types.APIResponse
// @Failure 400 {object} types.APIResponse
// @Failure 404 {object} types.APIResponse
// @Failure 500 {object} types.APIResponse
// @Router /api/v1/channels/{id}/send [post]
func (h *ChannelHandler) SendMessage(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, types.APIResponse{
			Status:  "error",
			Message: "Channel ID is required",
			Error: &types.ErrorInfo{
				Type:    "validation",
				Code:    "INVALID_ID",
				Message: "Channel ID is required",
			},
		})
		return
	}

	var message types.Message
	if err := c.ShouldBindJSON(&message); err != nil {
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

	err := h.service.SendMessage(c.Request.Context(), id, &message)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, types.APIResponse{
		Status:  "success",
		Message: "Message sent successfully",
	})
}

// GetChannelHealth 获取通道健康状态
// @Summary 获取通道健康状态
// @Description 获取指定通道的健康状态信息
// @Tags channels
// @Produce json
// @Param id path string true "通道ID"
// @Success 200 {object} types.APIResponse{data=channel.HealthStatus}
// @Failure 404 {object} types.APIResponse
// @Failure 500 {object} types.APIResponse
// @Router /api/v1/channels/{id}/health [get]
func (h *ChannelHandler) GetChannelHealth(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, types.APIResponse{
			Status:  "error",
			Message: "Channel ID is required",
			Error: &types.ErrorInfo{
				Type:    "validation",
				Code:    "INVALID_ID",
				Message: "Channel ID is required",
			},
		})
		return
	}

	// TODO: 实现健康检查逻辑
	// 这里需要调用 ChannelManager 的 HealthCheck 方法
	// health, err := h.manager.HealthCheck(c.Request.Context(), id)
	
	// 临时返回模拟数据
	c.JSON(http.StatusOK, types.APIResponse{
		Status:  "success",
		Message: "Channel health retrieved successfully",
		Data: map[string]interface{}{
			"channel_id": id,
			"status":     "healthy",
			"message":    "Channel is operating normally",
			"last_check": "2024-12-19T10:00:00Z",
		},
	})
}

// GetChannelStats 获取通道统计信息
// @Summary 获取通道统计信息
// @Description 获取指定通道的统计信息，包括消息发送量、成功率等
// @Tags channels
// @Produce json
// @Param id path string true "通道ID"
// @Param start_date query string false "开始日期 (YYYY-MM-DD)"
// @Param end_date query string false "结束日期 (YYYY-MM-DD)"
// @Success 200 {object} types.APIResponse{data=channel.ChannelStats}
// @Failure 404 {object} types.APIResponse
// @Failure 500 {object} types.APIResponse
// @Router /api/v1/channels/{id}/stats [get]
func (h *ChannelHandler) GetChannelStats(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, types.APIResponse{
			Status:  "error",
			Message: "Channel ID is required",
			Error: &types.ErrorInfo{
				Type:    "validation",
				Code:    "INVALID_ID",
				Message: "Channel ID is required",
			},
		})
		return
	}

	stats, err := h.service.GetChannelStats(c.Request.Context(), id)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, types.APIResponse{
		Status:  "success",
		Message: "Channel statistics retrieved successfully",
		Data:    stats,
	})
}



// BatchHealthCheck 批量健康检查
// @Summary 批量健康检查
// @Description 批量检查多个通道的健康状态
// @Tags channels
// @Accept json
// @Produce json
// @Param channel_ids body []string true "通道ID列表"
// @Success 200 {object} types.APIResponse{data=map[string]channel.HealthStatus}
// @Failure 400 {object} types.APIResponse
// @Failure 500 {object} types.APIResponse
// @Router /api/v1/channels/health/batch [post]
func (h *ChannelHandler) BatchHealthCheck(c *gin.Context) {
	var req struct {
		ChannelIDs []string `json:"channel_ids" binding:"required"`
	}

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

	// TODO: 实现批量健康检查逻辑
	// 这里需要调用 ChannelManager 的 BatchHealthCheck 方法
	// results, err := h.manager.BatchHealthCheck(c.Request.Context(), req.ChannelIDs)
	
	// 临时返回模拟数据
	results := make(map[string]interface{})
	for _, id := range req.ChannelIDs {
		results[id] = map[string]interface{}{
			"channel_id": id,
			"status":     "healthy",
			"message":    "Channel is operating normally",
			"last_check": "2024-12-19T10:00:00Z",
		}
	}

	c.JSON(http.StatusOK, types.APIResponse{
		Status:  "success",
		Message: "Batch health check completed",
		Data:    results,
	})
}

// handleError 统一错误处理
func (h *ChannelHandler) handleError(c *gin.Context, err error) {
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