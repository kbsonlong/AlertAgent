package v1

import (
	"net/http"
	"strconv"

	"alert_agent/internal/model"
	"alert_agent/internal/pkg/database"

	"github.com/gin-gonic/gin"
)

// ListGroups 获取通知组列表
// @Summary 获取通知组列表
// @Description 获取系统中所有通知组的列表信息
// @Tags 通知组管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.Response{data=[]model.NotifyGroup} "获取成功"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /api/v1/groups [get]
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
// @Summary 创建通知组
// @Description 创建新的通知组
// @Tags 通知组管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param group body model.NotifyGroup true "通知组信息"
// @Success 200 {object} response.Response{data=model.NotifyGroup} "创建成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /api/v1/groups [post]
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
// @Summary 获取单个通知组
// @Description 根据ID获取指定的通知组信息
// @Tags 通知组管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "通知组ID"
// @Success 200 {object} response.Response{data=model.NotifyGroup} "获取成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 404 {object} response.Response "通知组不存在"
// @Router /api/v1/groups/{id} [get]
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
// @Summary 更新通知组
// @Description 根据ID更新指定的通知组信息
// @Tags 通知组管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "通知组ID"
// @Param group body model.NotifyGroup true "通知组信息"
// @Success 200 {object} response.Response "更新成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 404 {object} response.Response "通知组不存在"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /api/v1/groups/{id} [put]
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
// @Summary 删除通知组
// @Description 根据ID删除指定的通知组
// @Tags 通知组管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "通知组ID"
// @Success 200 {object} response.Response "删除成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 404 {object} response.Response "通知组不存在"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /api/v1/groups/{id} [delete]
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
