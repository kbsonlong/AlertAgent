package v1

import (
	"net/http"

	"alert_agent/internal/config"
	"alert_agent/internal/model"
	"alert_agent/internal/pkg/database"

	"github.com/gin-gonic/gin"
)

// GetSettings 获取系统设置
func GetSettings(c *gin.Context) {
	var settings model.Settings
	result := database.DB.Order("updated_at desc").First(&settings)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"msg":  "获取系统设置失败",
			"data": nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "success",
		"data": settings,
	})
}

// GetSystemConfig 获取系统配置
func GetSystemConfig(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "success",
		"data": gin.H{
			"ollama_enabled": config.GetConfig().Ollama.Enabled,
		},
	})
}

// GetCurrentConfig 获取当前完整配置信息（用于调试和监控）
func GetCurrentConfig(c *gin.Context) {
	cfg := config.GetConfig()
	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "success",
		"data": gin.H{
			"server": gin.H{
				"port": cfg.Server.Port,
				"mode": cfg.Server.Mode,
			},
			"database": gin.H{
				"host":     cfg.Database.Host,
				"port":     cfg.Database.Port,
				"username": cfg.Database.Username,
				"dbname":   cfg.Database.DBName,
				"charset":  cfg.Database.Charset,
			},
			"redis": gin.H{
				"host":      cfg.Redis.Host,
				"port":      cfg.Redis.Port,
				"db":        cfg.Redis.DB,
				"pool_size": cfg.Redis.PoolSize,
			},
			"ollama": gin.H{
				"enabled":      cfg.Ollama.Enabled,
				"api_endpoint": cfg.Ollama.APIEndpoint,
				"model":        cfg.Ollama.Model,
				"timeout":      cfg.Ollama.Timeout,
				"max_retries":  cfg.Ollama.MaxRetries,
			},
		},
	})
}

// UpdateSettings 更新系统设置
func UpdateSettings(c *gin.Context) {
	var settings model.Settings
	if err := c.ShouldBindJSON(&settings); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "无效的请求参数",
			"data": nil,
		})
		return
	}

	result := database.DB.Save(&settings)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"msg":  "更新系统设置失败",
			"data": nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "success",
		"data": settings,
	})
}
