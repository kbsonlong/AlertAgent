package v1

import (
	"net/http"
	"strconv"

	"alert_agent/internal/model"
	"alert_agent/internal/pkg/database"

	"github.com/gin-gonic/gin"
)

// ListTemplates 获取通知模板列表
func ListTemplates(c *gin.Context) {
	var templates []model.NotifyTemplate
	result := database.DB.Find(&templates)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"msg":  "Failed to get templates",
			"data": nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "success",
		"data": templates,
	})
}

// CreateTemplate 创建通知模板
func CreateTemplate(c *gin.Context) {
	var template model.NotifyTemplate
	if err := c.ShouldBindJSON(&template); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "Invalid request body",
			"data": nil,
		})
		return
	}

	result := database.DB.Create(&template)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"msg":  "Failed to create template",
			"data": nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "success",
		"data": template,
	})
}

// GetTemplate 获取单个通知模板
func GetTemplate(c *gin.Context) {
	id := c.Param("id")
	templateID, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "Invalid template ID",
			"data": nil,
		})
		return
	}

	var template model.NotifyTemplate
	result := database.DB.First(&template, templateID)
	if result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code": 404,
			"msg":  "Template not found",
			"data": nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "success",
		"data": template,
	})
}

// UpdateTemplate 更新通知模板
func UpdateTemplate(c *gin.Context) {
	id := c.Param("id")
	templateID, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "Invalid template ID",
			"data": nil,
		})
		return
	}

	var template model.NotifyTemplate
	if err := c.ShouldBindJSON(&template); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "Invalid request body",
			"data": nil,
		})
		return
	}

	result := database.DB.Model(&model.NotifyTemplate{}).Where("id = ?", templateID).Updates(template)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"msg":  "Failed to update template",
			"data": nil,
		})
		return
	}

	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"code": 404,
			"msg":  "Template not found",
			"data": nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "success",
		"data": nil,
	})
}

// DeleteTemplate 删除通知模板
func DeleteTemplate(c *gin.Context) {
	id := c.Param("id")
	templateID, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "Invalid template ID",
			"data": nil,
		})
		return
	}

	result := database.DB.Delete(&model.NotifyTemplate{}, templateID)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"msg":  "Failed to delete template",
			"data": nil,
		})
		return
	}

	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"code": 404,
			"msg":  "Template not found",
			"data": nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "success",
		"data": nil,
	})
}
