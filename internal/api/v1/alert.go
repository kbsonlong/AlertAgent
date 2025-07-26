package v1

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"alert_agent/internal/model"
	"alert_agent/internal/pkg/database"
	"alert_agent/internal/pkg/logger"
	"alert_agent/internal/service"

	"github.com/gin-gonic/gin"
)

var (
	log = logger.L
)

// ListAlerts 获取告警列表
func ListAlerts(c *gin.Context) {
	var alerts []model.Alert
	result := database.DB.Find(&alerts)
	if result.Error != nil {
		c.Header("Content-Type", "application/json; charset=utf-8")
		c.Data(http.StatusInternalServerError, "application/json; charset=utf-8", []byte(fmt.Sprintf(`{"code":500,"msg":"获取告警列表失败","data":"%s"}`, result.Error.Error())))
		return
	}

	var resp []*model.AlertResponse
	for i := range alerts {
		resp = append(resp, alerts[i].ToResponse())
	}

	data, err := json.Marshal(gin.H{
		"code": 200,
		"msg":  "success",
		"data": resp,
	})
	if err != nil {
		c.Header("Content-Type", "application/json; charset=utf-8")
		c.Data(http.StatusInternalServerError, "application/json; charset=utf-8", []byte(fmt.Sprintf(`{"code":500,"msg":"序列化数据失败","data":"%s"}`, err.Error())))
		return
	}

	c.Header("Content-Type", "application/json; charset=utf-8")
	c.Data(http.StatusOK, "application/json; charset=utf-8", data)
}

// CreateAlert 创建告警
func CreateAlert(c *gin.Context) {
	var alert model.Alert
	if err := c.ShouldBindJSON(&alert); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "无效的请求参数: " + err.Error(),
			"data": nil,
		})
		return
	}

	// 暂时跳过 Ollama 分析
	alert.Analysis = ""
	alert.Status = "active"

	result := database.DB.Create(&alert)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"msg":  "创建告警失败: " + result.Error.Error(),
			"data": nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "success",
		"data": alert.ToResponse(),
	})
}

// GetAlert 获取单个告警
func GetAlert(c *gin.Context) {
	id := c.Param("id")
	var alert model.Alert
	result := database.DB.First(&alert, id)
	if result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code": 404,
			"msg":  "告警不存在",
			"data": result.Error.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "success",
		"data": alert.ToResponse(),
	})
}

// UpdateAlert 更新告警
func UpdateAlert(c *gin.Context) {
	id := c.Param("id")
	var alert model.Alert
	if err := c.ShouldBindJSON(&alert); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "无效的请求参数",
			"data": err.Error(),
		})
		return
	}

	result := database.DB.Model(&model.Alert{}).Where("id = ?", id).Updates(alert)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"msg":  "更新告警失败",
			"data": result.Error.Error(),
		})
		return
	}

	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"code": 404,
			"msg":  "告警不存在",
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

// HandleAlert 处理告警
func HandleAlert(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		Handler string `json:"handler" binding:"required"`
		Note    string `json:"note" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "无效的请求参数",
			"data": err.Error(),
		})
		return
	}

	now := time.Now()
	result := database.DB.Model(&model.Alert{}).Where("id = ?", id).Updates(map[string]interface{}{
		"handler":     req.Handler,
		"note":        req.Note,
		"status":      "handled",
		"handle_time": &now,
	})

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"msg":  "处理告警失败",
			"data": result.Error.Error(),
		})
		return
	}

	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"code": 404,
			"msg":  "告警不存在",
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

// AnalyzeAlert 分析告警
func AnalyzeAlert(c *gin.Context) {
	var alert model.Alert
	if err := c.ShouldBindJSON(&alert); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 动态创建ollama服务实例以获取最新配置
	ollamaService := service.NewOllamaService()
	analysis, err := ollamaService.AnalyzeAlert(c.Request.Context(), &alert)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"analysis": analysis,
	})
}

// FindSimilarAlerts 查找相似告警
func FindSimilarAlerts(c *gin.Context) {
	var alert model.Alert
	if err := c.ShouldBindJSON(&alert); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 动态创建ollama服务实例以获取最新配置
	ollamaService := service.NewOllamaService()
	similarAlerts, err := ollamaService.FindSimilarAlerts(c.Request.Context(), &alert)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"similar_alerts": similarAlerts,
	})
}
