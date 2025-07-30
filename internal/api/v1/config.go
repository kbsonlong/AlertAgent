package v1

import (
	"alert_agent/internal/config"
	"alert_agent/internal/pkg/response"

	"github.com/gin-gonic/gin"
)

// ConfigHandler 配置处理器
type ConfigHandler struct{}

// NewConfigHandler 创建配置处理器实例
func NewConfigHandler() *ConfigHandler {
	return &ConfigHandler{}
}

// GetConfig 获取当前配置
// @Summary 获取当前配置
// @Description 获取系统当前配置信息
// @Tags 配置管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.Response{data=config.Config}
// @Failure 401 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/config [get]
func (h *ConfigHandler) GetConfig(c *gin.Context) {
	cfg := config.GetConfig()
	response.Success(c, cfg)
}

// GetConfigYAML 获取配置的YAML格式
// @Summary 获取配置YAML
// @Description 获取系统配置的YAML格式
// @Tags 配置管理
// @Accept json
// @Produce text/yaml
// @Security BearerAuth
// @Success 200 {string} string "YAML配置内容"
// @Failure 401 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/config/yaml [get]
func (h *ConfigHandler) GetConfigYAML(c *gin.Context) {
	yamlData, err := config.GetConfigAsYAML()
	if err != nil {
		response.InternalServerError(c, "Failed to get config as YAML", err)
		return
	}

	c.Header("Content-Type", "text/yaml")
	c.String(200, string(yamlData))
}

// UpdateConfigRequest 更新配置请求
type UpdateConfigRequest struct {
	Config config.Config `json:"config" binding:"required"`
}

// UpdateConfig 更新配置
// @Summary 更新配置
// @Description 更新系统配置
// @Tags 配置管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body UpdateConfigRequest true "配置更新请求"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/config [put]
func (h *ConfigHandler) UpdateConfig(c *gin.Context) {
	var req UpdateConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	if err := config.UpdateConfig(req.Config); err != nil {
		response.BadRequest(c, "Failed to update config", err)
		return
	}

	response.SuccessWithMessage(c, "Configuration updated successfully", nil)
}

// SaveConfig 保存配置到文件
// @Summary 保存配置
// @Description 将当前配置保存到配置文件
// @Tags 配置管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.Response
// @Failure 401 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/config/save [post]
func (h *ConfigHandler) SaveConfig(c *gin.Context) {
	if err := config.SaveConfig(); err != nil {
		response.InternalServerError(c, "Failed to save config", err)
		return
	}

	response.SuccessWithMessage(c, "Configuration saved successfully", nil)
}

// GetConfigValueRequest 获取配置值请求
type GetConfigValueRequest struct {
	Path string `json:"path" binding:"required"`
}

// GetConfigValue 获取指定路径的配置值
// @Summary 获取配置值
// @Description 获取指定路径的配置值
// @Tags 配置管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param path query string true "配置路径（点分隔）"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Router /api/v1/config/value [get]
func (h *ConfigHandler) GetConfigValue(c *gin.Context) {
	path := c.Query("path")
	if path == "" {
		response.BadRequest(c, "Path parameter is required", nil)
		return
	}

	value, err := config.GetConfigValue(path)
	if err != nil {
		response.NotFound(c, "Config value not found", err)
		return
	}

	response.Success(c, map[string]interface{}{
		"path":  path,
		"value": value,
	})
}

// SetConfigValueRequest 设置配置值请求
type SetConfigValueRequest struct {
	Path  string      `json:"path" binding:"required"`
	Value interface{} `json:"value" binding:"required"`
}

// SetConfigValue 设置指定路径的配置值
// @Summary 设置配置值
// @Description 设置指定路径的配置值
// @Tags 配置管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body SetConfigValueRequest true "设置配置值请求"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Router /api/v1/config/value [put]
func (h *ConfigHandler) SetConfigValue(c *gin.Context) {
	var req SetConfigValueRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	if err := config.SetConfigValue(req.Path, req.Value); err != nil {
		response.BadRequest(c, "Failed to set config value", err)
		return
	}

	response.SuccessWithMessage(c, "Configuration value updated successfully", nil)
}

// ResetConfig 重置配置为默认值
// @Summary 重置配置
// @Description 将配置重置为默认值
// @Tags 配置管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.Response
// @Failure 401 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/config/reset [post]
func (h *ConfigHandler) ResetConfig(c *gin.Context) {
	if err := config.ResetToDefaults(); err != nil {
		response.InternalServerError(c, "Failed to reset config", err)
		return
	}

	response.SuccessWithMessage(c, "Configuration reset to defaults successfully", nil)
}

// RegisterConfigRoutes 注册配置管理路由
func RegisterConfigRoutes(r *gin.RouterGroup, requireAdmin gin.HandlerFunc) {
	configHandler := NewConfigHandler()
	
	configGroup := r.Group("/config")
	{
		configGroup.GET("", requireAdmin, configHandler.GetConfig)
		configGroup.GET("/yaml", requireAdmin, configHandler.GetConfigYAML)
		configGroup.PUT("", requireAdmin, configHandler.UpdateConfig)
		configGroup.POST("/save", requireAdmin, configHandler.SaveConfig)
		configGroup.GET("/value", requireAdmin, configHandler.GetConfigValue)
		configGroup.PUT("/value", requireAdmin, configHandler.SetConfigValue)
		configGroup.POST("/reset", requireAdmin, configHandler.ResetConfig)
	}
}