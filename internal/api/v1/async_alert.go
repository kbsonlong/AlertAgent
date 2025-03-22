package v1

import (
	"net/http"
	"strconv"
	"time"

	"alert_agent/internal/model"
	"alert_agent/internal/pkg/database"
	"alert_agent/internal/pkg/queue"
	"alert_agent/internal/pkg/types"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// AsyncAlertHandler 异步告警处理器
type AsyncAlertHandler struct {
	queue queue.Queue
}

// NewAsyncAlertHandler 创建异步告警处理器
func NewAsyncAlertHandler(queue queue.Queue) *AsyncAlertHandler {
	return &AsyncAlertHandler{
		queue: queue,
	}
}

// RegisterRoutes 注册路由
func (h *AsyncAlertHandler) RegisterRoutes(r *gin.RouterGroup) {
	alerts := r.Group("/alerts")
	{
		alerts.POST("/async/analyze", h.AsyncAnalyzeAlert)
		alerts.GET("/async/result/:task_id", h.GetAnalysisResult)
	}
}

// AsyncAnalyzeAlert 异步分析告警
// @Summary Asynchronously analyze alert using AI
// @Description Submit alert for asynchronous analysis using AI service
// @Tags alerts
// @Accept json
// @Produce json
// @Param id path int true "Alert ID"
// @Success 200 {object} response.Response{data=map[string]string}
// @Router /api/v1/alerts/{id}/async-analyze [post]
func (h *AsyncAlertHandler) AsyncAnalyzeAlert(c *gin.Context) {
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

	// 创建任务
	task := &types.AlertTask{
		ID:        uint(alertID),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name:      alert.Name,
		Level:     alert.Level,
		Source:    alert.Source,
		Content:   alert.Content,
		RuleID:    alert.RuleID,
		GroupID:   alert.GroupID,
	}

	// 将分析任务加入队列
	if err := h.queue.Push(c.Request.Context(), task); err != nil {
		zap.L().Error("enqueue analysis task failed",
			zap.Error(err),
			zap.Uint("alert_id", alert.ID),
		)

		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"msg":  "加入分析队列失败，请稍后重试",
			"data": nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "分析任务已加入队列，请稍后查看结果",
		"data": gin.H{
			"task_id":     alert.ID,
			"submit_time": time.Now().Format("2006-01-02 15:04:05"),
		},
	})
}

// GetAnalysisResult 获取分析结果
// @Summary Get alert analysis status
// @Description Get the status of an asynchronous alert analysis task
// @Tags alerts
// @Accept json
// @Produce json
// @Param id path int true "Alert ID"
// @Success 200 {object} response.Response{data=map[string]interface{}}
// @Router /api/v1/alerts/{id}/analysis-status [get]
func (h *AsyncAlertHandler) GetAnalysisResult(c *gin.Context) {
	taskID, err := strconv.ParseUint(c.Param("task_id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "无效的任务ID",
			"data": nil,
		})
		return
	}

	// 从队列获取结果
	result, err := h.queue.GetResult(c.Request.Context(), uint(taskID))
	if err != nil {
		zap.L().Error("get analysis result failed",
			zap.Error(err),
			zap.Uint("task_id", uint(taskID)),
		)

		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"msg":  "获取分析结果失败",
			"data": nil,
		})
		return
	}

	if result == nil {
		c.JSON(http.StatusOK, gin.H{
			"code": 200,
			"msg":  "分析任务处理中",
			"data": gin.H{
				"status": "processing",
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "success",
		"data": result,
	})
}
