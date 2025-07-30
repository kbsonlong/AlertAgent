package v1

import (
	"net/http"
	"strconv"

	"alert_agent/internal/application/channel/plugins"
	"alert_agent/internal/pkg/response"
	"alert_agent/internal/service"

	"github.com/gin-gonic/gin"
)

// PluginAPI 插件管理API
type PluginAPI struct {
	pluginService *service.PluginManagerService
}

// NewPluginAPI 创建插件API实例
func NewPluginAPI() *PluginAPI {
	return &PluginAPI{
		pluginService: service.NewPluginManagerService(),
	}
}

// GetAvailablePlugins 获取所有可用插件
// @Summary 获取所有可用插件
// @Description 获取系统中所有已注册的通知插件信息
// @Tags 插件管理
// @Accept json
// @Produce json
// @Success 200 {object} response.Response{data=[]plugins.PluginInfo}
// @Router /api/v1/plugins [get]
func (api *PluginAPI) GetAvailablePlugins(c *gin.Context) {
	plugins := api.pluginService.GetAvailablePlugins()
	response.Success(c, plugins)
}

// GetPluginConfig 获取插件配置
// @Summary 获取插件配置
// @Description 获取指定插件的配置信息
// @Tags 插件管理
// @Accept json
// @Produce json
// @Param name path string true "插件名称"
// @Success 200 {object} response.Response{data=plugins.PluginConfig}
// @Failure 404 {object} response.Response
// @Router /api/v1/plugins/{name}/config [get]
func (api *PluginAPI) GetPluginConfig(c *gin.Context) {
	pluginName := c.Param("name")
	
	config, err := api.pluginService.GetPluginConfig(pluginName)
	if err != nil {
		response.Error(c, http.StatusNotFound, err.Error())
		return
	}
	
	response.Success(c, config)
}

// SetPluginConfig 设置插件配置
// @Summary 设置插件配置
// @Description 设置指定插件的配置信息
// @Tags 插件管理
// @Accept json
// @Produce json
// @Param name path string true "插件名称"
// @Param config body plugins.PluginConfig true "插件配置"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Router /api/v1/plugins/{name}/config [post]
func (api *PluginAPI) SetPluginConfig(c *gin.Context) {
	pluginName := c.Param("name")
	
	var config plugins.PluginConfig
	if err := c.ShouldBindJSON(&config); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request body: "+err.Error())
		return
	}
	
	config.Name = pluginName
	
	if err := api.pluginService.ConfigurePlugin(c.Request.Context(), pluginName, &config); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	
	response.Success(c, "Plugin configured successfully")
}

// TestPluginConfig 测试插件配置
// @Summary 测试插件配置
// @Description 使用指定配置测试插件是否工作正常
// @Tags 插件管理
// @Accept json
// @Produce json
// @Param name path string true "插件名称"
// @Param config body map[string]interface{} true "插件配置"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Router /api/v1/plugins/{name}/test [post]
func (api *PluginAPI) TestPluginConfig(c *gin.Context) {
	pluginName := c.Param("name")
	
	var config map[string]interface{}
	if err := c.ShouldBindJSON(&config); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request body: "+err.Error())
		return
	}
	
	if err := api.pluginService.TestPluginConfig(c.Request.Context(), pluginName, config); err != nil {
		response.Error(c, http.StatusBadRequest, "Plugin test failed: "+err.Error())
		return
	}
	
	response.Success(c, "Plugin test successful")
}

// EnablePlugin 启用插件
// @Summary 启用插件
// @Description 启用指定的通知插件
// @Tags 插件管理
// @Accept json
// @Produce json
// @Param name path string true "插件名称"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Router /api/v1/plugins/{name}/enable [post]
func (api *PluginAPI) EnablePlugin(c *gin.Context) {
	pluginName := c.Param("name")
	
	if err := api.pluginService.EnablePlugin(c.Request.Context(), pluginName); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	
	response.Success(c, "Plugin enabled successfully")
}

// DisablePlugin 禁用插件
// @Summary 禁用插件
// @Description 禁用指定的通知插件
// @Tags 插件管理
// @Accept json
// @Produce json
// @Param name path string true "插件名称"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Router /api/v1/plugins/{name}/disable [post]
func (api *PluginAPI) DisablePlugin(c *gin.Context) {
	pluginName := c.Param("name")
	
	if err := api.pluginService.DisablePlugin(c.Request.Context(), pluginName); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	
	response.Success(c, "Plugin disabled successfully")
}

// GetPluginSchema 获取插件配置Schema
// @Summary 获取插件配置Schema
// @Description 获取指定插件的配置Schema，用于动态生成配置表单
// @Tags 插件管理
// @Accept json
// @Produce json
// @Param name path string true "插件名称"
// @Success 200 {object} response.Response{data=map[string]interface{}}
// @Failure 404 {object} response.Response
// @Router /api/v1/plugins/{name}/schema [get]
func (api *PluginAPI) GetPluginSchema(c *gin.Context) {
	pluginName := c.Param("name")
	
	schema, err := api.pluginService.GetPluginSchema(pluginName)
	if err != nil {
		response.Error(c, http.StatusNotFound, err.Error())
		return
	}
	
	response.Success(c, schema)
}

// HealthCheckPlugin 检查插件健康状态
// @Summary 检查插件健康状态
// @Description 检查指定插件的健康状态
// @Tags 插件管理
// @Accept json
// @Produce json
// @Param name path string true "插件名称"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Router /api/v1/plugins/{name}/health [get]
func (api *PluginAPI) HealthCheckPlugin(c *gin.Context) {
	pluginName := c.Param("name")
	
	if err := api.pluginService.HealthCheckPlugin(c.Request.Context(), pluginName); err != nil {
		response.Error(c, http.StatusBadRequest, "Plugin health check failed: "+err.Error())
		return
	}
	
	response.Success(c, "Plugin is healthy")
}

// HealthCheckAll 检查所有插件健康状态
// @Summary 检查所有插件健康状态
// @Description 检查所有已配置插件的健康状态
// @Tags 插件管理
// @Accept json
// @Produce json
// @Success 200 {object} response.Response{data=map[string]string}
// @Router /api/v1/plugins/health [get]
func (api *PluginAPI) HealthCheckAll(c *gin.Context) {
	results := api.pluginService.HealthCheckAll(c.Request.Context())
	
	healthStatus := make(map[string]string)
	for pluginName, err := range results {
		if err != nil {
			healthStatus[pluginName] = "unhealthy: " + err.Error()
		} else {
			healthStatus[pluginName] = "healthy"
		}
	}
	
	response.Success(c, healthStatus)
}

// GetPluginStats 获取插件统计信息
// @Summary 获取插件统计信息
// @Description 获取指定插件的使用统计信息
// @Tags 插件管理
// @Accept json
// @Produce json
// @Param name path string true "插件名称"
// @Success 200 {object} response.Response{data=plugins.PluginStats}
// @Failure 404 {object} response.Response
// @Router /api/v1/plugins/{name}/stats [get]
func (api *PluginAPI) GetPluginStats(c *gin.Context) {
	pluginName := c.Param("name")
	
	stats, err := api.pluginService.GetPluginStats(pluginName)
	if err != nil {
		response.Error(c, http.StatusNotFound, err.Error())
		return
	}
	
	response.Success(c, stats)
}

// GetAllPluginStats 获取所有插件统计信息
// @Summary 获取所有插件统计信息
// @Description 获取所有插件的使用统计信息
// @Tags 插件管理
// @Accept json
// @Produce json
// @Success 200 {object} response.Response{data=map[string]plugins.PluginStats}
// @Router /api/v1/plugins/stats [get]
func (api *PluginAPI) GetAllPluginStats(c *gin.Context) {
	stats := api.pluginService.GetAllPluginStats()
	response.Success(c, stats)
}

// SendTestNotification 发送测试通知
// @Summary 发送测试通知
// @Description 使用指定插件发送测试通知
// @Tags 插件管理
// @Accept json
// @Produce json
// @Param name path string true "插件名称"
// @Param message body plugins.NotificationMessage true "通知消息"
// @Success 200 {object} response.Response{data=plugins.SendResult}
// @Failure 400 {object} response.Response
// @Router /api/v1/plugins/{name}/send [post]
func (api *PluginAPI) SendTestNotification(c *gin.Context) {
	pluginName := c.Param("name")
	
	var message plugins.NotificationMessage
	if err := c.ShouldBindJSON(&message); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request body: "+err.Error())
		return
	}
	
	result, err := api.pluginService.SendNotification(c.Request.Context(), pluginName, &message)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Failed to send notification: "+err.Error())
		return
	}
	
	response.Success(c, result)
}

// GetPluginHealthStatus 获取插件健康状态
// @Summary 获取插件健康状态详情
// @Description 获取指定插件的详细健康状态信息
// @Tags 插件管理
// @Accept json
// @Produce json
// @Param name path string true "插件名称"
// @Success 200 {object} response.Response{data=plugins.HealthStatus}
// @Failure 404 {object} response.Response
// @Router /api/v1/plugins/{name}/health/status [get]
func (api *PluginAPI) GetPluginHealthStatus(c *gin.Context) {
	pluginName := c.Param("name")
	
	status, err := api.pluginService.GetPluginHealthStatus(pluginName)
	if err != nil {
		response.Error(c, http.StatusNotFound, err.Error())
		return
	}
	
	response.Success(c, status)
}

// GetAllPluginHealthStatus 获取所有插件健康状态
// @Summary 获取所有插件健康状态
// @Description 获取所有插件的详细健康状态信息
// @Tags 插件管理
// @Accept json
// @Produce json
// @Success 200 {object} response.Response{data=map[string]plugins.HealthStatus}
// @Router /api/v1/plugins/health/status [get]
func (api *PluginAPI) GetAllPluginHealthStatus(c *gin.Context) {
	statuses := api.pluginService.GetAllPluginHealthStatus()
	response.Success(c, statuses)
}

// GetPluginUsageStats 获取插件使用统计
// @Summary 获取插件使用统计
// @Description 获取指定插件的详细使用统计信息
// @Tags 插件管理
// @Accept json
// @Produce json
// @Param name path string true "插件名称"
// @Success 200 {object} response.Response{data=plugins.UsageStats}
// @Failure 404 {object} response.Response
// @Router /api/v1/plugins/{name}/usage [get]
func (api *PluginAPI) GetPluginUsageStats(c *gin.Context) {
	pluginName := c.Param("name")
	
	stats, err := api.pluginService.GetPluginUsageStats(pluginName)
	if err != nil {
		response.Error(c, http.StatusNotFound, err.Error())
		return
	}
	
	response.Success(c, stats)
}

// GetAllPluginUsageStats 获取所有插件使用统计
// @Summary 获取所有插件使用统计
// @Description 获取所有插件的详细使用统计信息
// @Tags 插件管理
// @Accept json
// @Produce json
// @Success 200 {object} response.Response{data=map[string]plugins.UsageStats}
// @Router /api/v1/plugins/usage [get]
func (api *PluginAPI) GetAllPluginUsageStats(c *gin.Context) {
	stats := api.pluginService.GetAllPluginUsageStats()
	response.Success(c, stats)
}

// HotLoadPlugin 热加载插件
// @Summary 热加载插件
// @Description 动态加载新的插件
// @Tags 插件管理
// @Accept json
// @Produce json
// @Param request body map[string]string true "插件路径"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Router /api/v1/plugins/hot-load [post]
func (api *PluginAPI) HotLoadPlugin(c *gin.Context) {
	var request struct {
		PluginPath string `json:"plugin_path" binding:"required"`
	}
	
	if err := c.ShouldBindJSON(&request); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request body: "+err.Error())
		return
	}
	
	if err := api.pluginService.HotLoadPlugin(request.PluginPath); err != nil {
		response.Error(c, http.StatusBadRequest, "Failed to hot load plugin: "+err.Error())
		return
	}
	
	response.Success(c, "Plugin hot loaded successfully")
}

// HotUnloadPlugin 热卸载插件
// @Summary 热卸载插件
// @Description 动态卸载指定插件
// @Tags 插件管理
// @Accept json
// @Produce json
// @Param name path string true "插件名称"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Router /api/v1/plugins/{name}/hot-unload [post]
func (api *PluginAPI) HotUnloadPlugin(c *gin.Context) {
	pluginName := c.Param("name")
	
	if err := api.pluginService.HotUnloadPlugin(pluginName); err != nil {
		response.Error(c, http.StatusBadRequest, "Failed to hot unload plugin: "+err.Error())
		return
	}
	
	response.Success(c, "Plugin hot unloaded successfully")
}

// HotReloadPlugin 热重载插件
// @Summary 热重载插件
// @Description 动态重新加载指定插件
// @Tags 插件管理
// @Accept json
// @Produce json
// @Param name path string true "插件名称"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Router /api/v1/plugins/{name}/hot-reload [post]
func (api *PluginAPI) HotReloadPlugin(c *gin.Context) {
	pluginName := c.Param("name")
	
	if err := api.pluginService.HotReloadPlugin(pluginName); err != nil {
		response.Error(c, http.StatusBadRequest, "Failed to hot reload plugin: "+err.Error())
		return
	}
	
	response.Success(c, "Plugin hot reloaded successfully")
}

// GetPluginConfigHistory 获取插件配置历史
// @Summary 获取插件配置历史
// @Description 获取指定插件的配置变更历史
// @Tags 插件管理
// @Accept json
// @Produce json
// @Param name path string true "插件名称"
// @Param page query int false "页码" default(1)
// @Param size query int false "每页大小" default(10)
// @Success 200 {object} response.Response
// @Router /api/v1/plugins/{name}/config/history [get]
func (api *PluginAPI) GetPluginConfigHistory(c *gin.Context) {
	pluginName := c.Param("name")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "10"))
	
	// TODO: 实现配置历史查询逻辑
	// 这里暂时返回空数据，实际实现需要在数据库中记录配置变更历史
	
	result := map[string]interface{}{
		"plugin": pluginName,
		"page":   page,
		"size":   size,
		"total":  0,
		"data":   []interface{}{},
	}
	
	response.Success(c, result)
}