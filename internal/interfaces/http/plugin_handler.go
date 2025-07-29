package http

import (
	"net/http"

	"alert_agent/internal/domain/channel"
	"alert_agent/pkg/types"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// PluginHandler 插件HTTP处理器
type PluginHandler struct {
	manager channel.ChannelManager
	logger  *zap.Logger
}

// NewPluginHandler 创建插件处理器
func NewPluginHandler(manager channel.ChannelManager, logger *zap.Logger) *PluginHandler {
	return &PluginHandler{
		manager: manager,
		logger:  logger,
	}
}

// ListPlugins 获取插件列表
// @Summary 获取插件列表
// @Description 获取所有可用的通道插件列表及其配置模式
// @Tags plugins
// @Produce json
// @Success 200 {object} types.APIResponse{data=[]channel.PluginInfo}
// @Failure 500 {object} types.APIResponse
// @Router /api/v1/plugins [get]
func (h *PluginHandler) ListPlugins(c *gin.Context) {
	plugins := h.manager.ListPlugins()
	
	// 转换为API响应格式
	pluginInfos := make([]map[string]interface{}, 0, len(plugins))
	for _, plugin := range plugins {
		pluginInfo := map[string]interface{}{
			"type":        string(plugin.GetType()),
			"name":        plugin.GetName(),
			"version":     plugin.GetVersion(),
			"description": plugin.GetDescription(),
			"schema":      plugin.GetConfigSchema(),
			"capabilities": plugin.GetCapabilities(),
		}
		pluginInfos = append(pluginInfos, pluginInfo)
	}

	c.JSON(http.StatusOK, types.APIResponse{
		Status:  "success",
		Message: "Plugins retrieved successfully",
		Data: map[string]interface{}{
			"plugins": pluginInfos,
		},
	})
}

// GetPlugin 获取插件详情
// @Summary 获取插件详情
// @Description 根据类型获取插件详细信息
// @Tags plugins
// @Produce json
// @Param type path string true "插件类型"
// @Success 200 {object} types.APIResponse{data=channel.PluginInfo}
// @Failure 404 {object} types.APIResponse
// @Failure 500 {object} types.APIResponse
// @Router /api/v1/plugins/{type} [get]
func (h *PluginHandler) GetPlugin(c *gin.Context) {
	pluginType := c.Param("type")
	if pluginType == "" {
		c.JSON(http.StatusBadRequest, types.APIResponse{
			Status:  "error",
			Message: "Plugin type is required",
			Error: &types.ErrorInfo{
				Type:    "validation",
				Code:    "INVALID_TYPE",
				Message: "Plugin type is required",
			},
		})
		return
	}

	plugin, err := h.manager.GetPlugin(channel.ChannelType(pluginType))
	if err != nil {
		h.logger.Error("plugin not found", zap.String("type", pluginType), zap.Error(err))
		c.JSON(http.StatusNotFound, types.APIResponse{
			Status:  "error",
			Message: "Plugin not found",
			Error: &types.ErrorInfo{
				Type:    "not_found",
				Code:    "PLUGIN_NOT_FOUND",
				Message: "Plugin not found",
			},
		})
		return
	}

	pluginInfo := map[string]interface{}{
		"type":        string(plugin.GetType()),
		"name":        plugin.GetName(),
		"version":     plugin.GetVersion(),
		"description": plugin.GetDescription(),
		"schema":      plugin.GetConfigSchema(),
		"capabilities": plugin.GetCapabilities(),
	}

	c.JSON(http.StatusOK, types.APIResponse{
		Status:  "success",
		Message: "Plugin retrieved successfully",
		Data:    pluginInfo,
	})
}

// ValidatePluginConfig 验证插件配置
// @Summary 验证插件配置
// @Description 验证指定插件类型的配置是否有效
// @Tags plugins
// @Accept json
// @Produce json
// @Param type path string true "插件类型"
// @Param config body channel.ChannelConfig true "插件配置"
// @Success 200 {object} types.APIResponse
// @Failure 400 {object} types.APIResponse
// @Failure 404 {object} types.APIResponse
// @Failure 500 {object} types.APIResponse
// @Router /api/v1/plugins/{type}/validate [post]
func (h *PluginHandler) ValidatePluginConfig(c *gin.Context) {
	pluginType := c.Param("type")
	if pluginType == "" {
		c.JSON(http.StatusBadRequest, types.APIResponse{
			Status:  "error",
			Message: "Plugin type is required",
			Error: &types.ErrorInfo{
				Type:    "validation",
				Code:    "INVALID_TYPE",
				Message: "Plugin type is required",
			},
		})
		return
	}

	var config channel.ChannelConfig
	if err := c.ShouldBindJSON(&config); err != nil {
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

	err := h.manager.ValidateConfig(c.Request.Context(), channel.ChannelType(pluginType), config)
	if err != nil {
		h.logger.Error("config validation failed", zap.String("type", pluginType), zap.Error(err))
		c.JSON(http.StatusBadRequest, types.APIResponse{
			Status:  "error",
			Message: "Configuration validation failed",
			Error: &types.ErrorInfo{
				Type:    "validation",
				Code:    "CONFIG_INVALID",
				Message: err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, types.APIResponse{
		Status:  "success",
		Message: "Configuration is valid",
	})
}

// TestPluginConnection 测试插件连接
// @Summary 测试插件连接
// @Description 使用指定配置测试插件连接
// @Tags plugins
// @Accept json
// @Produce json
// @Param type path string true "插件类型"
// @Param config body channel.ChannelConfig true "插件配置"
// @Success 200 {object} types.APIResponse{data=channel.TestResult}
// @Failure 400 {object} types.APIResponse
// @Failure 404 {object} types.APIResponse
// @Failure 500 {object} types.APIResponse
// @Router /api/v1/plugins/{type}/test [post]
func (h *PluginHandler) TestPluginConnection(c *gin.Context) {
	pluginType := c.Param("type")
	if pluginType == "" {
		c.JSON(http.StatusBadRequest, types.APIResponse{
			Status:  "error",
			Message: "Plugin type is required",
			Error: &types.ErrorInfo{
				Type:    "validation",
				Code:    "INVALID_TYPE",
				Message: "Plugin type is required",
			},
		})
		return
	}

	var config channel.ChannelConfig
	if err := c.ShouldBindJSON(&config); err != nil {
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

	plugin, err := h.manager.GetPlugin(channel.ChannelType(pluginType))
	if err != nil {
		h.logger.Error("plugin not found", zap.String("type", pluginType), zap.Error(err))
		c.JSON(http.StatusNotFound, types.APIResponse{
			Status:  "error",
			Message: "Plugin not found",
			Error: &types.ErrorInfo{
				Type:    "not_found",
				Code:    "PLUGIN_NOT_FOUND",
				Message: "Plugin not found",
			},
		})
		return
	}

	result, err := plugin.TestConnection(c.Request.Context(), config)
	if err != nil {
		h.logger.Error("plugin test failed", zap.String("type", pluginType), zap.Error(err))
		c.JSON(http.StatusInternalServerError, types.APIResponse{
			Status:  "error",
			Message: "Plugin test failed",
			Error: &types.ErrorInfo{
				Type:    "internal",
				Code:    "TEST_FAILED",
				Message: err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, types.APIResponse{
		Status:  "success",
		Message: "Plugin test completed",
		Data:    result,
	})
}