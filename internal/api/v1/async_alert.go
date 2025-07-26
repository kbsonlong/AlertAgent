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
	// 注册异步分析路由
	r.POST("/:id/async/analyze", h.AsyncAnalyzeAlert)
	r.GET("/async/result/:task_id", h.GetAnalysisResult)
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
			"data": err.Error(),
		})
		return
	}

	// 首先从MySQL查询告警信息和分析结果
	var alert model.Alert
	if err := database.DB.First(&alert, alertID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code": 404,
			"msg":  "告警不存在",
			"data": err.Error(),
		})
		return
	}

	// 检查是否已有分析结果
	if alert.Analysis != "" {
		// 已有分析结果，直接返回
		c.JSON(http.StatusOK, gin.H{
			"code": 200,
			"msg":  "分析结果已存在",
			"data": gin.H{
				"task_id": alert.ID,
				"status":  "completed",
				"result":  alert.Analysis,
				"message": "Analysis completed successfully",
			},
		})
		return
	}

	// 检查是否已在队列中处理
	queueResult, err := h.queue.GetResult(c.Request.Context(), uint(alertID))
	if err == nil && queueResult != nil {
		// 队列中已有结果，返回状态
		c.JSON(http.StatusOK, gin.H{
			"code": 200,
			"msg":  "分析任务已在处理中",
			"data": gin.H{
				"task_id": alert.ID,
				"status":  "processing",
			},
		})
		return
	}

	// 创建新的分析任务
	task := &types.AlertTask{
		ID:        uint(alertID),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name:      alert.Title,
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
			"status":      "processing",
		},
	})
}

// GetAnalysisResult 获取分析结果
// @Summary Get analysis result
// @Description Get the analysis result for a specific alert
// @Tags alerts
// @Accept json
// @Produce json
// @Param task_id path int true "Task ID"
// @Success 200 {object} response.Response{data=map[string]interface{}}
// @Router /api/v1/alerts/async/result/{task_id} [get]
func (h *AsyncAlertHandler) GetAnalysisResult(c *gin.Context) {
	taskID, err := strconv.ParseUint(c.Param("task_id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "无效的任务ID",
			"data": err.Error(),
		})
		return
	}

	// 首先从MySQL查询告警信息和分析结果
	var alert model.Alert
	if err := database.DB.First(&alert, taskID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code": 404,
			"msg":  "告警不存在",
			"data": err.Error(),
		})
		return
	}

	// 如果MySQL中已有分析结果，直接返回
	if alert.Analysis != "" {
		c.JSON(http.StatusOK, gin.H{
			"code": 200,
			"msg":  "获取分析结果成功",
			"data": gin.H{
				"status":  "completed",
				"result":  alert.Analysis,
				"message": "Analysis completed successfully",
			},
		})
		return
	}

	// 如果MySQL中没有结果，从Redis队列查询状态
	result, err := h.queue.GetResult(c.Request.Context(), uint(taskID))
	if err != nil {
		// 队列中也没有，说明任务未提交
		c.JSON(http.StatusNotFound, gin.H{
			"code": 404,
			"msg":  "分析任务不存在，请先提交分析任务",
			"data": gin.H{
				"status": "not_found",
			},
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

	// 返回队列中的处理状态
	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "获取分析状态成功",
		"data": result,
	})
}
