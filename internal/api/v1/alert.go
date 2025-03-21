package v1

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"alert_agent/internal/config"
	"alert_agent/internal/model"
	"alert_agent/internal/pkg/database"
	"alert_agent/internal/service"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

var (
	logger        = zap.L()
	openAIService = service.NewOpenAIService(&service.OpenAIConfig{
		Endpoint:   config.GlobalConfig.OpenAI.Endpoint,
		Model:      config.GlobalConfig.OpenAI.Model,
		Timeout:    config.GlobalConfig.OpenAI.Timeout,
		MaxRetries: config.GlobalConfig.OpenAI.MaxRetries,
	})
)

// ListAlerts 获取告警列表
func ListAlerts(c *gin.Context) {
	var alerts []model.Alert
	result := database.DB.Find(&alerts)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"msg":  "获取告警列表失败",
			"data": nil,
		})
		return
	}

	var resp []*model.AlertResponse
	for i := range alerts {
		resp = append(resp, alerts[i].ToResponse())
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "success",
		"data": resp,
	})
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
	alert.Analysis = "暂无分析"
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

// UpdateAlert 更新告警
func UpdateAlert(c *gin.Context) {
	id := c.Param("id")
	var alert model.Alert
	if err := c.ShouldBindJSON(&alert); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "无效的请求参数",
			"data": nil,
		})
		return
	}

	result := database.DB.Model(&model.Alert{}).Where("id = ?", id).Updates(alert)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"msg":  "更新告警失败",
			"data": nil,
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
			"data": nil,
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
			"data": nil,
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

// AnalyzeAlert godoc
// @Summary Analyze alert using AI
// @Description Analyze alert content using AI service
// @Tags alerts
// @Accept json
// @Produce json
// @Param id path int true "Alert ID"
// @Success 200 {object} response.Response{data=map[string]string}
// @Router /api/v1/alerts/{id}/analyze [post]
func AnalyzeAlert(c *gin.Context) {
	id := c.Param("id")
	alertID, err := strconv.Atoi(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "无效的告警ID",
			"data": nil,
		})
		return
	}

	// 获取告警信息
	var alert model.Alert
	if err := database.DB.First(&alert, alertID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code": 404,
			"msg":  "告警不存在",
			"data": nil,
		})
		return
	}

	// 调用OpenAI服务进行分析
	analysis, err := openAIService.AnalyzeAlert(c.Request.Context(), &alert)
	if err != nil {
		logger.Error("analyze alert failed",
			zap.Error(err),
			zap.Uint("alert_id", alert.ID),
			zap.String("alert_name", alert.Name),
		)

		// 根据错误类型提供不同的错误提示
		errorMsg := "分析服务暂时不可用，请稍后重试"
		if strings.Contains(err.Error(), "Ollama service is not available") {
			errorMsg = "AI分析服务当前不可用，系统正在尝试恢复连接，请稍后再试"
		} else if strings.Contains(err.Error(), "timed out") || strings.Contains(err.Error(), "deadline exceeded") {
			errorMsg = "AI分析服务响应超时，请稍后再试"
		} else if strings.Contains(err.Error(), "connection refused") || strings.Contains(err.Error(), "no route to host") {
			errorMsg = "无法连接到AI分析服务，请检查服务是否正常运行"
		}

		// 如果分析失败，返回空分析结果但不影响服务
		c.JSON(http.StatusOK, gin.H{
			"code": 200,
			"msg":  errorMsg,
			"data": gin.H{
				"analysis": "",
			},
		})
		return
	}

	// 更新告警的分析结果
	alert.Analysis = analysis
	if err := database.DB.Save(&alert).Error; err != nil {
		logger.Error("update alert analysis failed",
			zap.Error(err),
			zap.Uint("alert_id", alert.ID),
		)
		// 即使保存失败，也返回分析结果
		c.JSON(http.StatusOK, gin.H{
			"code": 200,
			"msg":  "分析结果获取成功，但更新数据库失败",
			"data": gin.H{
				"analysis": analysis,
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "success",
		"data": gin.H{
			"analysis": analysis,
		},
	})
}

// FindSimilarAlerts 查找相似告警
func FindSimilarAlerts(c *gin.Context) {
	id := c.Param("id")
	var alert model.Alert
	result := database.DB.First(&alert, id)
	if result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code": 404,
			"msg":  "告警不存在",
			"data": nil,
		})
		return
	}

	// 获取历史告警
	var history []model.Alert
	result = database.DB.Where("id != ?", id).Find(&history)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"msg":  "获取历史告警失败",
			"data": nil,
		})
		return
	}

	// 转换历史告警为指针切片
	historicalAlerts := make([]*model.Alert, len(history))
	for i := range history {
		historicalAlerts[i] = &history[i]
	}

	// 查找相似告警
	similarAlerts, err := openAIService.FindSimilarAlerts(c.Request.Context(), &alert, historicalAlerts)
	if err != nil {
		logger.Error("find similar alerts failed",
			zap.Error(err),
			zap.Uint("alert_id", alert.ID),
		)
		// 如果查找失败，返回空数组但不影响服务
		c.JSON(http.StatusOK, gin.H{
			"code": 200,
			"msg":  "success",
			"data": []*model.SimilarAlert{},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "success",
		"data": similarAlerts,
	})
}
