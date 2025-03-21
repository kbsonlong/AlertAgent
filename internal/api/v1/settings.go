package v1

import (
	"net/http"

	"alert_agent/internal/model"
	"alert_agent/internal/pkg/database"

	"github.com/gin-gonic/gin"
)

// GetSettings 获取系统设置
func GetSettings(c *gin.Context) {
	var settings model.Settings
	result := database.DB.First(&settings)
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
