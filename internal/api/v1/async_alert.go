package v1

import (
	"fmt"
	"net/http"
	"strconv"

	"alert_agent/internal/model"
	"alert_agent/internal/pkg/logger"
	"alert_agent/internal/pkg/queue"
	"alert_agent/internal/service"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// AsyncAlertAPI 异步告警API
type AsyncAlertAPI struct {
	alertService *service.AlertService
	taskService  *service.TaskService
}

// NewAsyncAlertAPI 创建异步告警API实例
func NewAsyncAlertAPI(alertService *service.AlertService, taskService *service.TaskService) *AsyncAlertAPI {
	return &AsyncAlertAPI{
		alertService: alertService,
		taskService:  taskService,
	}
}

// CreateAsyncAlert 创建异步告警
func (api *AsyncAlertAPI) CreateAsyncAlert(c *gin.Context) {
	var alert model.Alert
	if err := c.ShouldBindJSON(&alert); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "无效的请求参数: " + err.Error(),
			"data": nil,
		})
		return
	}

	// 创建告警（会自动触发异步分析和通知）
	if err := api.alertService.CreateAlert(c.Request.Context(), &alert); err != nil {
		logger.L.Error("Failed to create async alert",
			zap.Error(err),
			zap.String("alert_name", alert.Name),
		)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"msg":  "创建告警失败: " + err.Error(),
			"data": nil,
		})
		return
	}

	// 立即返回告警创建结果，分析将异步进行
	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "告警创建成功，正在异步分析中",
		"data": gin.H{
			"alert_id": alert.ID,
			"status":   "created",
			"analysis_status": "pending",
		},
	})
}

// GetAlertAnalysisStatus 获取告警分析状态
func (api *AsyncAlertAPI) GetAlertAnalysisStatus(c *gin.Context) {
	alertIDStr := c.Param("id")
	alertID, err := strconv.ParseUint(alertIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "无效的告警ID",
			"data": nil,
		})
		return
	}

	// 获取告警信息
	alert, err := api.alertService.GetAlert(c.Request.Context(), uint(alertID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code": 404,
			"msg":  "告警不存在",
			"data": nil,
		})
		return
	}

	// 构建响应数据
	response := gin.H{
		"alert_id": alert.ID,
		"status":   alert.Status,
		"analysis": alert.Analysis,
	}

	// 判断分析状态
	if alert.Analysis == "" {
		response["analysis_status"] = "pending"
	} else {
		response["analysis_status"] = "completed"
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "success",
		"data": response,
	})
}

// TriggerManualAnalysis 手动触发告警分析
func (api *AsyncAlertAPI) TriggerManualAnalysis(c *gin.Context) {
	alertIDStr := c.Param("id")
	alertID, err := strconv.ParseUint(alertIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "无效的告警ID",
			"data": nil,
		})
		return
	}

	// 触发手动分析
	if err := api.alertService.TriggerManualAnalysis(c.Request.Context(), uint(alertID)); err != nil {
		logger.L.Error("Failed to trigger manual analysis",
			zap.Error(err),
			zap.Uint64("alert_id", alertID),
		)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"msg":  "触发分析失败: " + err.Error(),
			"data": nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "分析任务已提交",
		"data": gin.H{
			"alert_id": alertID,
			"status":   "analysis_triggered",
		},
	})
}

// GetTaskStatus 获取任务状态
func (api *AsyncAlertAPI) GetTaskStatus(c *gin.Context) {
	taskID := c.Param("task_id")
	if taskID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "任务ID不能为空",
			"data": nil,
		})
		return
	}

	// 获取任务状态
	task, err := api.taskService.GetTaskStatus(c.Request.Context(), taskID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code": 404,
			"msg":  "任务不存在",
			"data": err.Error(),
		})
		return
	}

	if task == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code": 404,
			"msg":  "任务不存在",
			"data": nil,
		})
		return
	}

	// 构建响应数据
	response := gin.H{
		"task_id":     task.ID,
		"type":        string(task.Type),
		"status":      string(task.Status),
		"priority":    int(task.Priority),
		"retry":       task.Retry,
		"max_retry":   task.MaxRetry,
		"created_at":  task.CreatedAt,
		"updated_at":  task.UpdatedAt,
		"error_msg":   task.ErrorMsg,
	}

	if task.StartedAt != nil {
		response["started_at"] = *task.StartedAt
	}
	if task.CompletedAt != nil {
		response["completed_at"] = *task.CompletedAt
		response["duration"] = task.GetDuration().String()
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "success",
		"data": response,
	})
}

// GetQueueStats 获取队列统计信息
func (api *AsyncAlertAPI) GetQueueStats(c *gin.Context) {
	queueName := c.Query("queue")
	if queueName == "" {
		// 获取所有队列指标
		metrics, err := api.taskService.GetAllQueueMetrics(c.Request.Context())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"code": 500,
				"msg":  "获取队列指标失败: " + err.Error(),
				"data": nil,
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"code": 200,
			"msg":  "success",
			"data": metrics,
		})
		return
	}

	// 获取指定队列统计
	stats, err := api.taskService.GetQueueStats(c.Request.Context(), queueName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"msg":  "获取队列统计失败: " + err.Error(),
			"data": nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "success",
		"data": stats,
	})
}

// RetryFailedTasks 重试失败的任务
func (api *AsyncAlertAPI) RetryFailedTasks(c *gin.Context) {
	alertIDStr := c.Param("id")
	alertID, err := strconv.ParseUint(alertIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "无效的告警ID",
			"data": nil,
		})
		return
	}

	// 重试失败的任务
	if err := api.alertService.RetryFailedTasks(c.Request.Context(), uint(alertID)); err != nil {
		logger.L.Error("Failed to retry failed tasks",
			zap.Error(err),
			zap.Uint64("alert_id", alertID),
		)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"msg":  "重试任务失败: " + err.Error(),
			"data": nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "重试任务已提交",
		"data": gin.H{
			"alert_id": alertID,
			"status":   "retry_triggered",
		},
	})
}

// PublishCustomTask 发布自定义任务
func (api *AsyncAlertAPI) PublishCustomTask(c *gin.Context) {
	var req struct {
		QueueName string                 `json:"queue_name" binding:"required"`
		Type      string                 `json:"type" binding:"required"`
		Priority  int                    `json:"priority"`
		MaxRetry  int                    `json:"max_retry"`
		Payload   map[string]interface{} `json:"payload" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "无效的请求参数: " + err.Error(),
			"data": nil,
		})
		return
	}

	// 创建任务
	task := &queue.Task{
		Type:     queue.TaskType(req.Type),
		Priority: queue.TaskPriority(req.Priority),
		MaxRetry: req.MaxRetry,
		Payload:  req.Payload,
	}

	if task.MaxRetry == 0 {
		task.MaxRetry = 3
	}

	// 发布任务
	if err := api.taskService.PublishTask(c.Request.Context(), req.QueueName, task); err != nil {
		logger.L.Error("Failed to publish custom task",
			zap.Error(err),
			zap.String("queue", req.QueueName),
			zap.String("type", req.Type),
		)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"msg":  "发布任务失败: " + err.Error(),
			"data": nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "任务发布成功",
		"data": gin.H{
			"task_id":    task.ID,
			"queue_name": req.QueueName,
			"type":       req.Type,
			"status":     "published",
		},
	})
}

// GetSystemHealth 获取系统健康状态
func (api *AsyncAlertAPI) GetSystemHealth(c *gin.Context) {
	health, err := api.taskService.GetHealthStatus(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"msg":  "获取系统健康状态失败: " + err.Error(),
			"data": nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "success",
		"data": health,
	})
}

// BatchCreateAlerts 批量创建告警
func (api *AsyncAlertAPI) BatchCreateAlerts(c *gin.Context) {
	var req struct {
		Alerts []model.Alert `json:"alerts" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "无效的请求参数: " + err.Error(),
			"data": nil,
		})
		return
	}

	if len(req.Alerts) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "告警列表不能为空",
			"data": nil,
		})
		return
	}

	if len(req.Alerts) > 100 {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "批量创建告警数量不能超过100个",
			"data": nil,
		})
		return
	}

	// 批量创建告警
	var results []gin.H
	var successCount, failCount int

	for i, alert := range req.Alerts {
		if err := api.alertService.CreateAlert(c.Request.Context(), &alert); err != nil {
			logger.L.Error("Failed to create alert in batch",
				zap.Error(err),
				zap.Int("index", i),
				zap.String("alert_name", alert.Name),
			)
			results = append(results, gin.H{
				"index":    i,
				"alert_id": nil,
				"status":   "failed",
				"error":    err.Error(),
			})
			failCount++
		} else {
			results = append(results, gin.H{
				"index":    i,
				"alert_id": alert.ID,
				"status":   "created",
				"error":    nil,
			})
			successCount++
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  fmt.Sprintf("批量创建完成，成功: %d, 失败: %d", successCount, failCount),
		"data": gin.H{
			"total":        len(req.Alerts),
			"success_count": successCount,
			"fail_count":   failCount,
			"results":      results,
		},
	})
}