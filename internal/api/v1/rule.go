package v1

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"alert_agent/internal/model"
	"alert_agent/internal/pkg/database"
	"alert_agent/internal/pkg/logger"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
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
		logger.L.Error("Failed to get rule", zap.Error(result.Error))
		c.JSON(http.StatusNotFound, gin.H{
			"code": 404,
			"msg":  "Rule not found",
			"data": result.Error.Error(),
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

	// 直接解析请求体到map，避免gorm.Model字段的干扰
	var requestData map[string]interface{}
	if err := c.ShouldBindJSON(&requestData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "Invalid request body",
			"data": nil,
		})
		return
	}

	// 构建更新数据，只包含请求中提供的字段
	updateData := make(map[string]interface{})

	// 检查并添加各个字段
	if name, exists := requestData["name"]; exists {
		updateData["name"] = name
	}
	if description, exists := requestData["description"]; exists {
		updateData["description"] = description
	}
	if level, exists := requestData["level"]; exists {
		updateData["level"] = level
	}
	if enabled, exists := requestData["enabled"]; exists {
		updateData["enabled"] = enabled
	}
	if providerID, exists := requestData["provider_id"]; exists {
		updateData["provider_id"] = providerID
	}
	if queryExpr, exists := requestData["query_expr"]; exists {
		updateData["query_expr"] = queryExpr
	}
	if conditionExpr, exists := requestData["condition_expr"]; exists {
		updateData["condition_expr"] = conditionExpr
	}
	if notifyType, exists := requestData["notify_type"]; exists {
		updateData["notify_type"] = notifyType
	}
	if notifyGroup, exists := requestData["notify_group"]; exists {
		updateData["notify_group"] = notifyGroup
	}
	if template, exists := requestData["template"]; exists {
		updateData["template"] = template
	}

	// 使用原生SQL更新，避免GORM的自动字段处理
	if len(updateData) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "No fields to update",
			"data": nil,
		})
		return
	}
	logger.L.Debug("updateData", zap.Any("updateData", updateData))

	// 构建SET子句
	setClauses := make([]string, 0, len(updateData))
	values := make([]interface{}, 0, len(updateData)+1)

	for field, value := range updateData {
		setClauses = append(setClauses, field+" = ?")
		values = append(values, value)
	}
	values = append(values, ruleID) // 添加WHERE条件的参数

	sqlQuery := "UPDATE rules SET " + strings.Join(setClauses, ", ") + " WHERE id = ?"
	logger.L.Debug("sqlQuery", zap.String("sqlQuery", sqlQuery))
	result := database.DB.Exec(sqlQuery, values...)
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
