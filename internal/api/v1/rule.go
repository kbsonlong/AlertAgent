package v1

import (
	"encoding/json"
	"net/http"
	"strconv"

	"alert_agent/internal/model"
	"alert_agent/internal/pkg/database"

	"github.com/gin-gonic/gin"
)

// ListRules 获取告警规则列表
func ListRules(c *gin.Context) {
	var rules []model.Rule
	result := database.DB.Find(&rules)
	if result.Error != nil {
		c.Header("Content-Type", "application/json; charset=utf-8")
		c.Data(http.StatusInternalServerError, "application/json; charset=utf-8", []byte(`{"code":500,"msg":"获取规则列表失败","data":null}`))
		return
	}

	data, err := json.Marshal(gin.H{
		"code": 200,
		"msg":  "success",
		"data": rules,
	})
	if err != nil {
		c.Header("Content-Type", "application/json; charset=utf-8")
		c.Data(http.StatusInternalServerError, "application/json; charset=utf-8", []byte(`{"code":500,"msg":"序列化数据失败","data":null}`))
		return
	}

	c.Header("Content-Type", "application/json; charset=utf-8")
	c.Data(http.StatusOK, "application/json; charset=utf-8", data)
}

// CreateRule 创建告警规则
func CreateRule(c *gin.Context) {
	var rule model.Rule
	if err := c.ShouldBindJSON(&rule); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "Invalid request body",
			"data": nil,
		})
		return
	}

	result := database.DB.Create(&rule)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"msg":  "Failed to create rule",
			"data": nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "success",
		"data": rule,
	})
}

// GetRule 获取单个告警规则
func GetRule(c *gin.Context) {
	id := c.Param("id")
	ruleID, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "Invalid rule ID",
			"data": nil,
		})
		return
	}

	var rule model.Rule
	result := database.DB.First(&rule, ruleID)
	if result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code": 404,
			"msg":  "Rule not found",
			"data": nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "success",
		"data": rule,
	})
}

// UpdateRule 更新告警规则
func UpdateRule(c *gin.Context) {
	id := c.Param("id")
	ruleID, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "Invalid rule ID",
			"data": nil,
		})
		return
	}

	var rule model.Rule
	if err := c.ShouldBindJSON(&rule); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "Invalid request body",
			"data": nil,
		})
		return
	}

	result := database.DB.Model(&model.Rule{}).Where("id = ?", ruleID).Updates(rule)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"msg":  "Failed to update rule",
			"data": nil,
		})
		return
	}

	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"code": 404,
			"msg":  "Rule not found",
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

// DeleteRule 删除告警规则
func DeleteRule(c *gin.Context) {
	id := c.Param("id")
	ruleID, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "Invalid rule ID",
			"data": nil,
		})
		return
	}

	result := database.DB.Delete(&model.Rule{}, ruleID)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"msg":  "Failed to delete rule",
			"data": nil,
		})
		return
	}

	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"code": 404,
			"msg":  "Rule not found",
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
