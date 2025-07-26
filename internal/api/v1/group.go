package v1

import (
	"net/http"
	"strconv"

	"alert_agent/internal/model"
	"alert_agent/internal/pkg/database"

	"github.com/gin-gonic/gin"
)

// ListGroups 获取通知组列表
func ListGroups(c *gin.Context) {
	var groups []model.NotifyGroup
	result := database.DB.Find(&groups)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"msg":  "获取通知组列表失败",
			"data": result.Error.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "success",
		"data": groups,
	})
}

// CreateGroup 创建通知组
func CreateGroup(c *gin.Context) {
	var group model.NotifyGroup
	if err := c.ShouldBindJSON(&group); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "Invalid request body",
			"data": err.Error(),
		})
		return
	}

	result := database.DB.Create(&group)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"msg":  "Failed to create group",
			"data": result.Error.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "success",
		"data": group,
	})
}

// GetGroup 获取单个通知组
func GetGroup(c *gin.Context) {
	id := c.Param("id")
	groupID, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "Invalid group ID",
			"data": err.Error(),
		})
		return
	}

	var group model.NotifyGroup
	result := database.DB.First(&group, groupID)
	if result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code": 404,
			"msg":  "Group not found",
			"data": result.Error.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "success",
		"data": group,
	})
}

// UpdateGroup 更新通知组
func UpdateGroup(c *gin.Context) {
	id := c.Param("id")
	groupID, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "Invalid group ID",
			"data": nil,
		})
		return
	}

	var group model.NotifyGroup
	if err := c.ShouldBindJSON(&group); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "Invalid request body",
			"data": err.Error(),
		})
		return
	}

	result := database.DB.Model(&model.NotifyGroup{}).Where("id = ?", groupID).Updates(group)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"msg":  "Failed to update group",
			"data": result.Error.Error(),
		})
		return
	}

	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"code": 404,
			"msg":  "Group not found",
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

// DeleteGroup 删除通知组
func DeleteGroup(c *gin.Context) {
	id := c.Param("id")
	groupID, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "Invalid group ID",
			"data": nil,
		})
		return
	}

	result := database.DB.Delete(&model.NotifyGroup{}, groupID)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"msg":  "Failed to delete group",
			"data": result.Error.Error(),
		})
		return
	}

	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"code": 404,
			"msg":  "Group not found",
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
