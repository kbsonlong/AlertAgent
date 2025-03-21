package v1

import (
	"net/http"
	"strconv"
	"time"

	"alert_agent/internal/model"
	"alert_agent/internal/pkg/database"
	"alert_agent/internal/pkg/queue"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// AsyncAnalyzeAlert 异步分析告警
// @Summary Asynchronously analyze alert using AI
// @Description Submit alert for asynchronous analysis using AI service
// @Tags alerts
// @Accept json
// @Produce json
// @Param id path int true "Alert ID"
// @Success 200 {object} response.Response{data=map[string]string}
// @Router /api/v1/alerts/{id}/async-analyze [post]
func AsyncAnalyzeAlert(c *gin.Context) {
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

	// 将分析任务加入队列
	if err := queue.EnqueueAnalysisTask(c.Request.Context(), alert.ID); err != nil {
		logger.Error("enqueue analysis task failed",
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

// GetAnalysisStatus 获取分析状态
// @Summary Get alert analysis status
// @Description Get the status of an asynchronous alert analysis task
// @Tags alerts
// @Accept json
// @Produce json
// @Param id path int true "Alert ID"
// @Success 200 {object} response.Response{data=map[string]interface{}}
// @Router /api/v1/alerts/{id}/analysis-status [get]
func GetAnalysisStatus(c *gin.Context) {
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

	// 检查数据库中是否已有分析结果
	if alert.Analysis != "" && alert.Analysis != "暂无分析" {
		c.JSON(http.StatusOK, gin.H{
			"code": 200,
			"msg":  "分析已完成",
			"data": gin.H{
				"status":   "completed",
				"analysis": alert.Analysis,
			},
		})
		return
	}

	// 从Redis获取分析结果
	result, err := queue.GetAnalysisResult(c.Request.Context(), alert.ID)
	if err != nil {
		logger.Error("get analysis result failed",
			zap.Error(err),
			zap.Uint("alert_id", alert.ID),
		)

		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"msg":  "获取分析状态失败",
			"data": nil,
		})
		return
	}

	// 如果结果不存在，说明任务还在队列中或处理中
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

	// 如果有错误信息，说明分析失败
	if result.Error != "" {
		c.JSON(http.StatusOK, gin.H{
			"code": 200,
			"msg":  "分析失败",
			"data": gin.H{
				"status": "failed",
				"error":  result.Error,
			},
		})
		return
	}

	// 分析成功，返回结果
	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "分析已完成",
		"data": gin.H{
			"status":   "completed",
			"analysis": result.Analysis,
		},
	})
}
