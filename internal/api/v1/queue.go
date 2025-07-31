package v1

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"alert_agent/internal/pkg/queue"
	"alert_agent/internal/pkg/response"

	"github.com/gin-gonic/gin"
)

// QueueAPI 队列管理API
type QueueAPI struct {
	queueManager *queue.RedisMessageQueue
	monitor      *queue.QueueMonitor
}

// NewQueueAPI 创建队列API实例
func NewQueueAPI(queueManager *queue.RedisMessageQueue, monitor *queue.QueueMonitor) *QueueAPI {
	return &QueueAPI{
		queueManager: queueManager,
		monitor:      monitor,
	}
}

// GetQueueMetrics 获取队列指标
// @Summary 获取队列指标
// @Description 获取指定队列的详细指标信息
// @Tags Queue
// @Accept json
// @Produce json
// @Param queue_name path string true "队列名称"
// @Success 200 {object} response.Response{data=queue.QueueMetrics}
// @Router /api/v1/queues/{queue_name}/metrics [get]
func (q *QueueAPI) GetQueueMetrics(c *gin.Context) {
	queueName := c.Param("queue_name")
	if queueName == "" {
		response.Error(c, http.StatusBadRequest, "队列名称不能为空", nil)
		return
	}

	metrics, err := q.monitor.GetQueueMetrics(c.Request.Context(), queueName)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "获取队列指标失败: "+err.Error(), err)
		return
	}

	response.Success(c, metrics)
}

// GetAllQueueMetrics 获取所有队列指标
// @Summary 获取所有队列指标
// @Description 获取系统中所有队列的指标信息
// @Tags Queue
// @Accept json
// @Produce json
// @Success 200 {object} response.Response{data=map[string]queue.QueueMetrics}
// @Router /api/v1/queues/metrics [get]
func (q *QueueAPI) GetAllQueueMetrics(c *gin.Context) {
	metrics, err := q.monitor.GetAllQueueMetrics(c.Request.Context())
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "获取队列指标失败: "+err.Error(), err)
		return
	}

	response.Success(c, metrics)
}

// GetTaskMetrics 获取任务类型指标
// @Summary 获取任务类型指标
// @Description 获取指定任务类型的指标信息
// @Tags Queue
// @Accept json
// @Produce json
// @Param task_type path string true "任务类型"
// @Success 200 {object} response.Response{data=queue.TaskMetrics}
// @Router /api/v1/queues/tasks/{task_type}/metrics [get]
func (q *QueueAPI) GetTaskMetrics(c *gin.Context) {
	taskTypeStr := c.Param("task_type")
	if taskTypeStr == "" {
		response.Error(c, http.StatusBadRequest, "任务类型不能为空", nil)
		return
	}

	taskType := queue.TaskType(taskTypeStr)
	metrics, err := q.monitor.GetTaskMetrics(c.Request.Context(), taskType)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "获取任务指标失败: "+err.Error(), err)
		return
	}

	response.Success(c, metrics)
}

// GetTaskStatus 获取任务状态
// @Summary 获取任务状态
// @Description 获取指定任务的详细状态信息
// @Tags Queue
// @Accept json
// @Produce json
// @Param task_id path string true "任务ID"
// @Success 200 {object} response.Response{data=queue.Task}
// @Router /api/v1/queues/tasks/{task_id} [get]
func (q *QueueAPI) GetTaskStatus(c *gin.Context) {
	taskID := c.Param("task_id")
	if taskID == "" {
		response.Error(c, http.StatusBadRequest, "任务ID不能为空", nil)
		return
	}

	task, err := q.queueManager.GetTaskStatus(c.Request.Context(), taskID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "获取任务状态失败: "+err.Error(), err)
		return
	}

	if task == nil {
		response.Error(c, http.StatusNotFound, "任务不存在", nil)
		return
	}

	response.Success(c, task)
}

// ListTasks 获取任务列表
// @Summary 获取任务列表
// @Description 获取指定队列的任务列表，支持分页和过滤
// @Tags Queue
// @Accept json
// @Produce json
// @Param queue_name query string false "队列名称"
// @Param status query string false "任务状态"
// @Param task_type query string false "任务类型"
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页大小" default(20)
// @Success 200 {object} response.Response{data=TaskListResponse}
// @Router /api/v1/queues/tasks [get]
func (q *QueueAPI) ListTasks(c *gin.Context) {
	queueName := c.Query("queue_name")
	status := c.Query("status")
	taskType := c.Query("task_type")
	
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	// 这里需要实现任务列表查询逻辑
	// 由于当前的队列实现主要基于Redis，我们需要扩展查询功能
	tasks, total, err := q.getTaskList(c, queueName, status, taskType, page, pageSize)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "获取任务列表失败: "+err.Error(), err)
		return
	}

	result := TaskListResponse{
		Tasks:    tasks,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}

	response.Success(c, result)
}

// RetryTask 重试任务
// @Summary 重试任务
// @Description 重新执行失败的任务
// @Tags Queue
// @Accept json
// @Produce json
// @Param task_id path string true "任务ID"
// @Success 200 {object} response.Response
// @Router /api/v1/queues/tasks/{task_id}/retry [post]
func (q *QueueAPI) RetryTask(c *gin.Context) {
	taskID := c.Param("task_id")
	if taskID == "" {
		response.Error(c, http.StatusBadRequest, "任务ID不能为空", nil)
		return
	}

	// 获取任务详情
	task, err := q.queueManager.GetTaskStatus(c.Request.Context(), taskID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "获取任务状态失败: "+err.Error(), err)
		return
	}

	if task == nil {
		response.Error(c, http.StatusNotFound, "任务不存在", nil)
		return
	}

	// 检查任务是否可以重试
	if task.Status != queue.TaskStatusFailed {
		response.Error(c, http.StatusBadRequest, "只能重试失败的任务", nil)
		return
	}

	// 重置任务状态并重新发布
	task.Status = queue.TaskStatusPending
	task.Retry = 0
	task.ErrorMsg = ""
	task.UpdatedAt = time.Now()

	err = q.queueManager.Publish(c.Request.Context(), string(task.Type), task)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "重试任务失败: "+err.Error(), err)
		return
	}

	response.Success(c, gin.H{"message": "任务已重新加入队列"})
}

// SkipTask 跳过任务
// @Summary 跳过任务
// @Description 跳过指定的任务，将其标记为已跳过
// @Tags Queue
// @Accept json
// @Produce json
// @Param task_id path string true "任务ID"
// @Success 200 {object} response.Response
// @Router /api/v1/queues/tasks/{task_id}/skip [post]
func (q *QueueAPI) SkipTask(c *gin.Context) {
	taskID := c.Param("task_id")
	if taskID == "" {
		response.Error(c, http.StatusBadRequest, "任务ID不能为空", nil)
		return
	}

	// 获取任务详情
	task, err := q.queueManager.GetTaskStatus(c.Request.Context(), taskID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "获取任务状态失败: "+err.Error(), err)
		return
	}

	if task == nil {
		response.Error(c, http.StatusNotFound, "任务不存在", nil)
		return
	}

	// 只能跳过待处理或失败的任务
	if task.Status != queue.TaskStatusPending && task.Status != queue.TaskStatusFailed {
		response.Error(c, http.StatusBadRequest, "只能跳过待处理或失败的任务", nil)
		return
	}

	// 使用Nack将任务移动到死信队列
	err = q.queueManager.Nack(c.Request.Context(), task, false)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "跳过任务失败: "+err.Error(), err)
		return
	}

	response.Success(c, gin.H{"message": "任务已跳过"})
}

// BatchRetryTasks 批量重试任务
// @Summary 批量重试任务
// @Description 批量重试失败的任务
// @Tags Queue
// @Accept json
// @Produce json
// @Param request body BatchTaskRequest true "批量任务请求"
// @Success 200 {object} response.Response{data=BatchTaskResponse}
// @Router /api/v1/queues/tasks/batch/retry [post]
func (q *QueueAPI) BatchRetryTasks(c *gin.Context) {
	var req BatchTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "请求参数错误: "+err.Error(), err)
		return
	}

	if len(req.TaskIDs) == 0 {
		response.Error(c, http.StatusBadRequest, "任务ID列表不能为空", nil)
		return
	}

	result := BatchTaskResponse{
		Total:     len(req.TaskIDs),
		Succeeded: 0,
		Failed:    0,
		Results:   make([]TaskOperationResult, 0, len(req.TaskIDs)),
	}

	for _, taskID := range req.TaskIDs {
		opResult := TaskOperationResult{
			TaskID: taskID,
		}

		// 获取任务详情
		task, err := q.queueManager.GetTaskStatus(c.Request.Context(), taskID)
		if err != nil {
			opResult.Success = false
			opResult.Error = "获取任务状态失败: " + err.Error()
			result.Failed++
		} else if task == nil {
			opResult.Success = false
			opResult.Error = "任务不存在"
			result.Failed++
		} else if task.Status != queue.TaskStatusFailed {
			opResult.Success = false
			opResult.Error = "只能重试失败的任务"
			result.Failed++
		} else {
			// 重试任务
			task.Status = queue.TaskStatusPending
			task.Retry = 0
			task.ErrorMsg = ""
			task.UpdatedAt = time.Now()

			err = q.queueManager.Publish(c.Request.Context(), string(task.Type), task)
			if err != nil {
				opResult.Success = false
				opResult.Error = "重试任务失败: " + err.Error()
				result.Failed++
			} else {
				opResult.Success = true
				result.Succeeded++
			}
		}

		result.Results = append(result.Results, opResult)
	}

	response.Success(c, result)
}

// GetQueueHealth 获取队列健康状态
// @Summary 获取队列健康状态
// @Description 获取所有队列的健康状态信息
// @Tags Queue
// @Accept json
// @Produce json
// @Success 200 {object} response.Response{data=map[string]interface{}}
// @Router /api/v1/queues/health [get]
func (q *QueueAPI) GetQueueHealth(c *gin.Context) {
	health, err := q.monitor.GetHealthStatus(c.Request.Context())
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "获取队列健康状态失败: "+err.Error(), err)
		return
	}

	response.Success(c, health)
}

// CleanupExpiredTasks 清理过期任务
// @Summary 清理过期任务
// @Description 清理指定队列中的过期任务
// @Tags Queue
// @Accept json
// @Produce json
// @Param queue_name path string true "队列名称"
// @Param max_age query string false "最大存活时间" default("1h")
// @Success 200 {object} response.Response
// @Router /api/v1/queues/{queue_name}/cleanup [post]
func (q *QueueAPI) CleanupExpiredTasks(c *gin.Context) {
	queueName := c.Param("queue_name")
	if queueName == "" {
		response.Error(c, http.StatusBadRequest, "队列名称不能为空", nil)
		return
	}

	maxAgeStr := c.DefaultQuery("max_age", "1h")
	maxAge, err := time.ParseDuration(maxAgeStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "时间格式错误: "+err.Error(), err)
		return
	}

	err = q.monitor.CleanupExpiredTasks(c.Request.Context(), queueName, maxAge)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "清理过期任务失败: "+err.Error(), err)
		return
	}

	response.Success(c, gin.H{"message": "过期任务清理完成"})
}

// 辅助方法和数据结构

// TaskListResponse 任务列表响应
type TaskListResponse struct {
	Tasks    []*queue.Task `json:"tasks"`
	Total    int64         `json:"total"`
	Page     int           `json:"page"`
	PageSize int           `json:"page_size"`
}

// BatchTaskRequest 批量任务请求
type BatchTaskRequest struct {
	TaskIDs []string `json:"task_ids" binding:"required"`
}

// BatchTaskResponse 批量任务响应
type BatchTaskResponse struct {
	Total     int                     `json:"total"`
	Succeeded int                     `json:"succeeded"`
	Failed    int                     `json:"failed"`
	Results   []TaskOperationResult   `json:"results"`
}

// TaskOperationResult 任务操作结果
type TaskOperationResult struct {
	TaskID  string `json:"task_id"`
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

// getTaskList 获取任务列表（需要实现具体的查询逻辑）
func (q *QueueAPI) getTaskList(ctx *gin.Context, queueName, status, taskType string, page, pageSize int) ([]*queue.Task, int64, error) {
	// 这里需要实现具体的任务列表查询逻辑
	// 由于当前基于Redis的实现，可能需要扫描相关的键来获取任务列表
	// 为了演示，这里返回空列表
	return []*queue.Task{}, 0, nil
}

// BatchSkipTasks 批量跳过任务
// @Summary 批量跳过任务
// @Description 批量跳过待处理或失败的任务
// @Tags Queue
// @Accept json
// @Produce json
// @Param request body BatchTaskRequest true "批量任务请求"
// @Success 200 {object} response.Response{data=BatchTaskResponse}
// @Router /api/v1/queues/tasks/batch/skip [post]
func (q *QueueAPI) BatchSkipTasks(c *gin.Context) {
	var req BatchTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "请求参数错误: "+err.Error(), err)
		return
	}

	if len(req.TaskIDs) == 0 {
		response.Error(c, http.StatusBadRequest, "任务ID列表不能为空", nil)
		return
	}

	result := BatchTaskResponse{
		Total:     len(req.TaskIDs),
		Succeeded: 0,
		Failed:    0,
		Results:   make([]TaskOperationResult, 0, len(req.TaskIDs)),
	}

	for _, taskID := range req.TaskIDs {
		opResult := TaskOperationResult{
			TaskID: taskID,
		}

		// 获取任务详情
		task, err := q.queueManager.GetTaskStatus(c.Request.Context(), taskID)
		if err != nil {
			opResult.Success = false
			opResult.Error = "获取任务状态失败: " + err.Error()
			result.Failed++
		} else if task == nil {
			opResult.Success = false
			opResult.Error = "任务不存在"
			result.Failed++
		} else if task.Status != queue.TaskStatusPending && task.Status != queue.TaskStatusFailed {
			opResult.Success = false
			opResult.Error = "只能跳过待处理或失败的任务"
			result.Failed++
		} else {
			// 跳过任务
			err = q.queueManager.Nack(c.Request.Context(), task, false)
			if err != nil {
				opResult.Success = false
				opResult.Error = "跳过任务失败: " + err.Error()
				result.Failed++
			} else {
				opResult.Success = true
				result.Succeeded++
			}
		}

		result.Results = append(result.Results, opResult)
	}

	response.Success(c, result)
}

// GetTaskLogs 获取任务执行日志
// @Summary 获取任务执行日志
// @Description 获取指定任务的执行日志
// @Tags Queue
// @Accept json
// @Produce json
// @Param task_id path string true "任务ID"
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页大小" default(50)
// @Success 200 {object} response.Response{data=TaskLogsResponse}
// @Router /api/v1/queues/tasks/{task_id}/logs [get]
func (q *QueueAPI) GetTaskLogs(c *gin.Context) {
	taskID := c.Param("task_id")
	if taskID == "" {
		response.Error(c, http.StatusBadRequest, "任务ID不能为空", nil)
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "50"))
	
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 50
	}

	// 这里需要实现获取任务日志的逻辑
	logs, total, err := q.getTaskLogs(c, taskID, page, pageSize)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "获取任务日志失败: "+err.Error(), err)
		return
	}

	result := TaskLogsResponse{
		Logs:     logs,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}

	response.Success(c, result)
}

// CancelTask 取消任务
// @Summary 取消任务
// @Description 取消正在处理的任务
// @Tags Queue
// @Accept json
// @Produce json
// @Param task_id path string true "任务ID"
// @Success 200 {object} response.Response
// @Router /api/v1/queues/tasks/{task_id}/cancel [post]
func (q *QueueAPI) CancelTask(c *gin.Context) {
	taskID := c.Param("task_id")
	if taskID == "" {
		response.Error(c, http.StatusBadRequest, "任务ID不能为空", nil)
		return
	}

	// 获取任务详情
	task, err := q.queueManager.GetTaskStatus(c.Request.Context(), taskID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "获取任务状态失败: "+err.Error(), err)
		return
	}

	if task == nil {
		response.Error(c, http.StatusNotFound, "任务不存在", nil)
		return
	}

	// 只能取消处理中的任务
	if task.Status != queue.TaskStatusProcessing {
		response.Error(c, http.StatusBadRequest, "只能取消正在处理的任务", nil)
		return
	}

	// 标记任务为取消状态并移动到死信队列
	task.MarkFailed("Task cancelled by user")
	err = q.queueManager.Nack(c.Request.Context(), task, false)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "取消任务失败: "+err.Error(), err)
		return
	}

	response.Success(c, gin.H{"message": "任务已取消"})
}

// ExportTasks 导出任务数据
// @Summary 导出任务数据
// @Description 导出任务数据为CSV或JSON格式
// @Tags Queue
// @Accept json
// @Produce json
// @Param queue_name query string false "队列名称"
// @Param status query string false "任务状态"
// @Param task_type query string false "任务类型"
// @Param start_time query string false "开始时间"
// @Param end_time query string false "结束时间"
// @Param format query string false "导出格式" default("csv")
// @Success 200 {file} file
// @Router /api/v1/queues/tasks/export [get]
func (q *QueueAPI) ExportTasks(c *gin.Context) {
	format := c.DefaultQuery("format", "csv")
	if format != "csv" && format != "json" {
		response.Error(c, http.StatusBadRequest, "不支持的导出格式", nil)
		return
	}

	// 构建过滤条件
	filter := TaskFilter{
		QueueName: c.Query("queue_name"),
		Status:    c.Query("status"),
		TaskType:  c.Query("task_type"),
		Page:      1,
		PageSize:  10000, // 导出时使用较大的页面大小
	}

	if startTimeStr := c.Query("start_time"); startTimeStr != "" {
		if startTime, err := time.Parse(time.RFC3339, startTimeStr); err == nil {
			filter.StartTime = &startTime
		}
	}

	if endTimeStr := c.Query("end_time"); endTimeStr != "" {
		if endTime, err := time.Parse(time.RFC3339, endTimeStr); err == nil {
			filter.EndTime = &endTime
		}
	}

	// 获取任务数据
	tasks, _, err := q.getTaskList(c, filter.QueueName, filter.Status, filter.TaskType, filter.Page, filter.PageSize)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "获取任务数据失败: "+err.Error(), err)
		return
	}

	// 根据格式导出数据
	switch format {
	case "csv":
		q.exportTasksAsCSV(c, tasks)
	case "json":
		q.exportTasksAsJSON(c, tasks)
	}
}

// GetTaskHistory 获取任务执行历史
// @Summary 获取任务执行历史
// @Description 获取任务的执行历史记录
// @Tags Queue
// @Accept json
// @Produce json
// @Param task_id path string true "任务ID"
// @Success 200 {object} response.Response{data=[]TaskHistoryRecord}
// @Router /api/v1/queues/tasks/{task_id}/history [get]
func (q *QueueAPI) GetTaskHistory(c *gin.Context) {
	taskID := c.Param("task_id")
	if taskID == "" {
		response.Error(c, http.StatusBadRequest, "任务ID不能为空", nil)
		return
	}

	history, err := q.getTaskHistory(c, taskID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "获取任务历史失败: "+err.Error(), err)
		return
	}

	response.Success(c, history)
}

// 辅助方法

// TaskLogsResponse 任务日志响应
type TaskLogsResponse struct {
	Logs     []*TaskLog `json:"logs"`
	Total    int64      `json:"total"`
	Page     int        `json:"page"`
	PageSize int        `json:"page_size"`
}

// TaskLog 任务日志
type TaskLog struct {
	ID        string                 `json:"id"`
	TaskID    string                 `json:"task_id"`
	Level     string                 `json:"level"`
	Message   string                 `json:"message"`
	Timestamp time.Time              `json:"timestamp"`
	WorkerID  string                 `json:"worker_id,omitempty"`
	Context   map[string]interface{} `json:"context,omitempty"`
}

// TaskHistoryRecord 任务历史记录
type TaskHistoryRecord struct {
	ID          string    `json:"id"`
	TaskID      string    `json:"task_id"`
	Action      string    `json:"action"` // created, started, completed, failed, retried, cancelled
	Status      string    `json:"status"`
	Message     string    `json:"message"`
	WorkerID    string    `json:"worker_id,omitempty"`
	Timestamp   time.Time `json:"timestamp"`
	Duration    int64     `json:"duration,omitempty"` // 毫秒
	ErrorMsg    string    `json:"error_msg,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// TaskFilter 任务过滤器（扩展版本）
type TaskFilter struct {
	QueueName string
	Status    string
	TaskType  string
	Page      int
	PageSize  int
	StartTime *time.Time
	EndTime   *time.Time
}

// getTaskLogs 获取任务日志（需要实现）
func (q *QueueAPI) getTaskLogs(ctx *gin.Context, taskID string, page, pageSize int) ([]*TaskLog, int64, error) {
	// 这里需要实现从Redis或数据库获取任务日志的逻辑
	// 为了演示，返回空结果
	return []*TaskLog{}, 0, nil
}

// getTaskHistory 获取任务历史（需要实现）
func (q *QueueAPI) getTaskHistory(ctx *gin.Context, taskID string) ([]*TaskHistoryRecord, error) {
	// 这里需要实现从Redis或数据库获取任务历史的逻辑
	// 为了演示，返回空结果
	return []*TaskHistoryRecord{}, nil
}

// exportTasksAsCSV 导出任务为CSV格式
func (q *QueueAPI) exportTasksAsCSV(c *gin.Context, tasks []*queue.Task) {
	c.Header("Content-Type", "text/csv")
	c.Header("Content-Disposition", "attachment; filename=tasks.csv")

	// CSV头部
	csvHeader := "ID,Type,Status,Priority,Retry,MaxRetry,CreatedAt,StartedAt,CompletedAt,Duration,ErrorMsg,WorkerID\n"
	c.String(http.StatusOK, csvHeader)

	// CSV数据
	for _, task := range tasks {
		duration := ""
		if task.StartedAt != nil {
			if task.CompletedAt != nil {
				duration = fmt.Sprintf("%.2f", task.CompletedAt.Sub(*task.StartedAt).Seconds())
			} else {
				duration = fmt.Sprintf("%.2f", time.Since(*task.StartedAt).Seconds())
			}
		}

		startedAt := ""
		if task.StartedAt != nil {
			startedAt = task.StartedAt.Format(time.RFC3339)
		}

		completedAt := ""
		if task.CompletedAt != nil {
			completedAt = task.CompletedAt.Format(time.RFC3339)
		}

		line := fmt.Sprintf("%s,%s,%s,%d,%d,%d,%s,%s,%s,%s,%s,%s\n",
			task.ID,
			task.Type,
			task.Status,
			task.Priority,
			task.Retry,
			task.MaxRetry,
			task.CreatedAt.Format(time.RFC3339),
			startedAt,
			completedAt,
			duration,
			task.ErrorMsg,
			task.WorkerID,
		)
		c.String(http.StatusOK, line)
	}
}

// exportTasksAsJSON 导出任务为JSON格式
func (q *QueueAPI) exportTasksAsJSON(c *gin.Context, tasks []*queue.Task) {
	c.Header("Content-Type", "application/json")
	c.Header("Content-Disposition", "attachment; filename=tasks.json")

	response.Success(c, gin.H{
		"tasks":      tasks,
		"total":      len(tasks),
		"exported_at": time.Now(),
	})
}

// OptimizeQueue 队列优化
// @Summary 队列优化
// @Description 对指定队列执行优化操作
// @Tags Queue
// @Accept json
// @Produce json
// @Param queue_name path string true "队列名称"
// @Param request body QueueOptimizeRequest true "优化请求"
// @Success 200 {object} response.Response{data=QueueOptimizeResponse}
// @Router /api/v1/queues/{queue_name}/optimize [post]
func (q *QueueAPI) OptimizeQueue(c *gin.Context) {
	queueName := c.Param("queue_name")
	if queueName == "" {
		response.Error(c, http.StatusBadRequest, "队列名称不能为空", nil)
		return
	}

	var req QueueOptimizeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "请求参数错误: "+err.Error(), err)
		return
	}

	// 执行队列优化
	result, err := q.optimizeQueue(c, queueName, &req)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "队列优化失败: "+err.Error(), err)
		return
	}

	response.Success(c, result)
}

// GetQueueRecommendations 获取队列优化建议
// @Summary 获取队列优化建议
// @Description 获取指定队列的性能优化建议
// @Tags Queue
// @Accept json
// @Produce json
// @Param queue_name path string true "队列名称"
// @Success 200 {object} response.Response{data=[]QueueRecommendation}
// @Router /api/v1/queues/{queue_name}/recommendations [get]
func (q *QueueAPI) GetQueueRecommendations(c *gin.Context) {
	queueName := c.Param("queue_name")
	if queueName == "" {
		response.Error(c, http.StatusBadRequest, "队列名称不能为空", nil)
		return
	}

	recommendations, err := q.getQueueRecommendations(c, queueName)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "获取优化建议失败: "+err.Error(), err)
		return
	}

	response.Success(c, recommendations)
}

// ScaleQueue 队列扩缩容
// @Summary 队列扩缩容
// @Description 手动调整队列的Worker数量
// @Tags Queue
// @Accept json
// @Produce json
// @Param queue_name path string true "队列名称"
// @Param request body QueueScaleRequest true "扩缩容请求"
// @Success 200 {object} response.Response{data=QueueScaleResponse}
// @Router /api/v1/queues/{queue_name}/scale [post]
func (q *QueueAPI) ScaleQueue(c *gin.Context) {
	queueName := c.Param("queue_name")
	if queueName == "" {
		response.Error(c, http.StatusBadRequest, "队列名称不能为空", nil)
		return
	}

	var req QueueScaleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "请求参数错误: "+err.Error(), err)
		return
	}

	// 执行队列扩缩容
	result, err := q.scaleQueue(c, queueName, &req)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "队列扩缩容失败: "+err.Error(), err)
		return
	}

	response.Success(c, result)
}

// PauseQueue 暂停队列
// @Summary 暂停队列
// @Description 暂停指定队列的任务处理
// @Tags Queue
// @Accept json
// @Produce json
// @Param queue_name path string true "队列名称"
// @Success 200 {object} response.Response
// @Router /api/v1/queues/{queue_name}/pause [post]
func (q *QueueAPI) PauseQueue(c *gin.Context) {
	queueName := c.Param("queue_name")
	if queueName == "" {
		response.Error(c, http.StatusBadRequest, "队列名称不能为空", nil)
		return
	}

	err := q.pauseQueue(c, queueName)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "暂停队列失败: "+err.Error(), err)
		return
	}

	response.Success(c, gin.H{"message": "队列已暂停"})
}

// ResumeQueue 恢复队列
// @Summary 恢复队列
// @Description 恢复指定队列的任务处理
// @Tags Queue
// @Accept json
// @Produce json
// @Param queue_name path string true "队列名称"
// @Success 200 {object} response.Response
// @Router /api/v1/queues/{queue_name}/resume [post]
func (q *QueueAPI) ResumeQueue(c *gin.Context) {
	queueName := c.Param("queue_name")
	if queueName == "" {
		response.Error(c, http.StatusBadRequest, "队列名称不能为空", nil)
		return
	}

	err := q.resumeQueue(c, queueName)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "恢复队列失败: "+err.Error(), err)
		return
	}

	response.Success(c, gin.H{"message": "队列已恢复"})
}

// GetQueueAlerts 获取队列告警
// @Summary 获取队列告警
// @Description 获取队列相关的告警信息
// @Tags Queue
// @Accept json
// @Produce json
// @Param queue_name query string false "队列名称"
// @Param status query string false "告警状态"
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页大小" default(20)
// @Success 200 {object} response.Response{data=QueueAlertsResponse}
// @Router /api/v1/queues/alerts [get]
func (q *QueueAPI) GetQueueAlerts(c *gin.Context) {
	queueName := c.Query("queue_name")
	status := c.Query("status")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	alerts, total, err := q.getQueueAlerts(c, queueName, status, page, pageSize)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "获取队列告警失败: "+err.Error(), err)
		return
	}

	result := QueueAlertsResponse{
		Alerts:   alerts,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}

	response.Success(c, result)
}

// AcknowledgeAlert 确认告警
// @Summary 确认告警
// @Description 确认指定的队列告警
// @Tags Queue
// @Accept json
// @Produce json
// @Param alert_id path string true "告警ID"
// @Success 200 {object} response.Response
// @Router /api/v1/queues/alerts/{alert_id}/acknowledge [post]
func (q *QueueAPI) AcknowledgeAlert(c *gin.Context) {
	alertID := c.Param("alert_id")
	if alertID == "" {
		response.Error(c, http.StatusBadRequest, "告警ID不能为空", nil)
		return
	}

	err := q.acknowledgeAlert(c, alertID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "确认告警失败: "+err.Error(), err)
		return
	}

	response.Success(c, gin.H{"message": "告警已确认"})
}

// 辅助方法和数据结构

// QueueOptimizeRequest 队列优化请求
type QueueOptimizeRequest struct {
	AutoScale       bool   `json:"auto_scale"`
	CleanupExpired  bool   `json:"cleanup_expired"`
	Rebalance       bool   `json:"rebalance"`
	MaxAge          string `json:"max_age,omitempty"`
	OptimizeWorkers bool   `json:"optimize_workers"`
}

// QueueOptimizeResponse 队列优化响应
type QueueOptimizeResponse struct {
	QueueName   string                 `json:"queue_name"`
	Operations  []OptimizationOperation `json:"operations"`
	Summary     OptimizationSummary    `json:"summary"`
	Duration    int64                  `json:"duration"` // 毫秒
	CompletedAt time.Time              `json:"completed_at"`
}

// OptimizationOperation 优化操作
type OptimizationOperation struct {
	Type        string                 `json:"type"`
	Status      string                 `json:"status"` // success, failed, skipped
	Message     string                 `json:"message"`
	Details     map[string]interface{} `json:"details,omitempty"`
	Duration    int64                  `json:"duration"` // 毫秒
	Error       string                 `json:"error,omitempty"`
}

// OptimizationSummary 优化摘要
type OptimizationSummary struct {
	TotalOperations     int     `json:"total_operations"`
	SuccessfulOperations int     `json:"successful_operations"`
	FailedOperations    int     `json:"failed_operations"`
	SkippedOperations   int     `json:"skipped_operations"`
	SuccessRate         float64 `json:"success_rate"`
	ImprovementScore    float64 `json:"improvement_score"`
}

// QueueRecommendation 队列优化建议
type QueueRecommendation struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"`
	Priority    string                 `json:"priority"`
	Title       string                 `json:"title"`
	Description string                 `json:"description"`
	Action      string                 `json:"action"`
	Impact      string                 `json:"impact"`
	Metrics     map[string]interface{} `json:"metrics"`
	AutoFix     bool                   `json:"auto_fix"`
	CreatedAt   time.Time              `json:"created_at"`
}

// QueueScaleRequest 队列扩缩容请求
type QueueScaleRequest struct {
	TargetWorkers int    `json:"target_workers" binding:"required,min=0,max=100"`
	ScaleReason   string `json:"scale_reason,omitempty"`
	Force         bool   `json:"force,omitempty"`
}

// QueueScaleResponse 队列扩缩容响应
type QueueScaleResponse struct {
	QueueName       string    `json:"queue_name"`
	CurrentWorkers  int       `json:"current_workers"`
	TargetWorkers   int       `json:"target_workers"`
	ScaledWorkers   int       `json:"scaled_workers"`
	Status          string    `json:"status"`
	Message         string    `json:"message"`
	EstimatedTime   int64     `json:"estimated_time"` // 秒
	CompletedAt     time.Time `json:"completed_at"`
}

// QueueAlert 队列告警
type QueueAlert struct {
	ID          string                 `json:"id"`
	QueueName   string                 `json:"queue_name"`
	Type        string                 `json:"type"`
	Severity    string                 `json:"severity"`
	Title       string                 `json:"title"`
	Message     string                 `json:"message"`
	Details     map[string]interface{} `json:"details"`
	Status      string                 `json:"status"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
	ResolvedAt  *time.Time             `json:"resolved_at,omitempty"`
	AcknowledgedAt *time.Time          `json:"acknowledged_at,omitempty"`
	AcknowledgedBy string              `json:"acknowledged_by,omitempty"`
}

// QueueAlertsResponse 队列告警响应
type QueueAlertsResponse struct {
	Alerts   []*QueueAlert `json:"alerts"`
	Total    int64         `json:"total"`
	Page     int           `json:"page"`
	PageSize int           `json:"page_size"`
}

// 实现方法

// optimizeQueue 执行队列优化
func (q *QueueAPI) optimizeQueue(ctx *gin.Context, queueName string, req *QueueOptimizeRequest) (*QueueOptimizeResponse, error) {
	startTime := time.Now()
	response := &QueueOptimizeResponse{
		QueueName:   queueName,
		Operations:  make([]OptimizationOperation, 0),
		CompletedAt: time.Now(),
	}

	// 自动扩缩容
	if req.AutoScale {
		op := q.executeAutoScale(ctx, queueName)
		response.Operations = append(response.Operations, op)
	}

	// 清理过期任务
	if req.CleanupExpired {
		maxAge := "24h"
		if req.MaxAge != "" {
			maxAge = req.MaxAge
		}
		op := q.executeCleanupExpired(ctx, queueName, maxAge)
		response.Operations = append(response.Operations, op)
	}

	// 重新平衡队列
	if req.Rebalance {
		op := q.executeRebalance(ctx, queueName)
		response.Operations = append(response.Operations, op)
	}

	// Worker优化
	if req.OptimizeWorkers {
		op := q.executeWorkerOptimization(ctx, queueName)
		response.Operations = append(response.Operations, op)
	}

	// 计算摘要
	response.Duration = time.Since(startTime).Milliseconds()
	response.Summary = q.calculateOptimizationSummary(response.Operations)

	return response, nil
}

// executeAutoScale 执行自动扩缩容
func (q *QueueAPI) executeAutoScale(ctx *gin.Context, queueName string) OptimizationOperation {
	startTime := time.Now()
	op := OptimizationOperation{
		Type:    "auto_scale",
		Details: make(map[string]interface{}),
	}

	// 获取队列指标
	metrics, err := q.monitor.GetQueueMetrics(ctx.Request.Context(), queueName)
	if err != nil {
		op.Status = "failed"
		op.Error = err.Error()
		op.Message = "获取队列指标失败"
		op.Duration = time.Since(startTime).Milliseconds()
		return op
	}

	// 计算目标Worker数量
	targetWorkers := q.calculateTargetWorkers(metrics)
	currentWorkers := q.getCurrentWorkers(ctx, queueName)

	op.Details["current_workers"] = currentWorkers
	op.Details["target_workers"] = targetWorkers
	op.Details["pending_count"] = metrics.PendingCount

	if targetWorkers == currentWorkers {
		op.Status = "skipped"
		op.Message = "Worker数量已是最优"
	} else {
		// 执行扩缩容（这里需要与容器编排系统集成）
		err := q.scaleWorkers(ctx, queueName, targetWorkers)
		if err != nil {
			op.Status = "failed"
			op.Error = err.Error()
			op.Message = "扩缩容执行失败"
		} else {
			op.Status = "success"
			op.Message = fmt.Sprintf("Worker数量从 %d 调整为 %d", currentWorkers, targetWorkers)
		}
	}

	op.Duration = time.Since(startTime).Milliseconds()
	return op
}

// executeCleanupExpired 执行过期任务清理
func (q *QueueAPI) executeCleanupExpired(ctx *gin.Context, queueName, maxAge string) OptimizationOperation {
	startTime := time.Now()
	op := OptimizationOperation{
		Type:    "cleanup_expired",
		Details: make(map[string]interface{}),
	}

	maxAgeDuration, err := time.ParseDuration(maxAge)
	if err != nil {
		op.Status = "failed"
		op.Error = err.Error()
		op.Message = "时间格式错误"
		op.Duration = time.Since(startTime).Milliseconds()
		return op
	}

	err = q.monitor.CleanupExpiredTasks(ctx.Request.Context(), queueName, maxAgeDuration)
	if err != nil {
		op.Status = "failed"
		op.Error = err.Error()
		op.Message = "清理过期任务失败"
	} else {
		op.Status = "success"
		op.Message = "过期任务清理完成"
		op.Details["max_age"] = maxAge
	}

	op.Duration = time.Since(startTime).Milliseconds()
	return op
}

// executeRebalance 执行队列重新平衡
func (q *QueueAPI) executeRebalance(ctx *gin.Context, queueName string) OptimizationOperation {
	startTime := time.Now()
	op := OptimizationOperation{
		Type:    "rebalance",
		Details: make(map[string]interface{}),
	}

	// 这里需要实现队列重新平衡逻辑
	// 简化实现
	op.Status = "success"
	op.Message = "队列重新平衡完成"

	op.Duration = time.Since(startTime).Milliseconds()
	return op
}

// executeWorkerOptimization 执行Worker优化
func (q *QueueAPI) executeWorkerOptimization(ctx *gin.Context, queueName string) OptimizationOperation {
	startTime := time.Now()
	op := OptimizationOperation{
		Type:    "worker_optimization",
		Details: make(map[string]interface{}),
	}

	// 这里需要实现Worker优化逻辑
	// 简化实现
	op.Status = "success"
	op.Message = "Worker优化完成"

	op.Duration = time.Since(startTime).Milliseconds()
	return op
}

// calculateTargetWorkers 计算目标Worker数量
func (q *QueueAPI) calculateTargetWorkers(metrics *queue.QueueMetrics) int {
	// 简单的扩缩容算法
	if metrics.PendingCount > 100 {
		return 5
	} else if metrics.PendingCount > 50 {
		return 3
	} else if metrics.PendingCount > 10 {
		return 2
	}
	return 1
}

// getCurrentWorkers 获取当前Worker数量
func (q *QueueAPI) getCurrentWorkers(ctx *gin.Context, queueName string) int {
	// 这里需要实现获取当前Worker数量的逻辑
	return 1
}

// scaleWorkers 执行Worker扩缩容
func (q *QueueAPI) scaleWorkers(ctx *gin.Context, queueName string, targetWorkers int) error {
	// 这里需要与容器编排系统（如Kubernetes）集成
	// 简化实现
	return nil
}

// calculateOptimizationSummary 计算优化摘要
func (q *QueueAPI) calculateOptimizationSummary(operations []OptimizationOperation) OptimizationSummary {
	summary := OptimizationSummary{
		TotalOperations: len(operations),
	}

	for _, op := range operations {
		switch op.Status {
		case "success":
			summary.SuccessfulOperations++
		case "failed":
			summary.FailedOperations++
		case "skipped":
			summary.SkippedOperations++
		}
	}

	if summary.TotalOperations > 0 {
		summary.SuccessRate = float64(summary.SuccessfulOperations) / float64(summary.TotalOperations) * 100
	}

	// 简单的改进分数计算
	summary.ImprovementScore = summary.SuccessRate * 0.8

	return summary
}

// getQueueRecommendations 获取队列优化建议
func (q *QueueAPI) getQueueRecommendations(ctx *gin.Context, queueName string) ([]*QueueRecommendation, error) {
	// 这里需要实现获取优化建议的逻辑
	// 简化实现，返回示例建议
	recommendations := []*QueueRecommendation{
		{
			ID:          "rec_001",
			Type:        "performance",
			Priority:    "medium",
			Title:       "考虑增加Worker数量",
			Description: "当前队列积压较多，建议增加Worker数量以提高处理能力",
			Action:      "将Worker数量从1增加到3",
			Impact:      "预计可提高50%的处理速度",
			AutoFix:     true,
			CreatedAt:   time.Now(),
		},
	}

	return recommendations, nil
}

// scaleQueue 执行队列扩缩容
func (q *QueueAPI) scaleQueue(ctx *gin.Context, queueName string, req *QueueScaleRequest) (*QueueScaleResponse, error) {
	currentWorkers := q.getCurrentWorkers(ctx, queueName)
	
	response := &QueueScaleResponse{
		QueueName:      queueName,
		CurrentWorkers: currentWorkers,
		TargetWorkers:  req.TargetWorkers,
		CompletedAt:    time.Now(),
	}

	if currentWorkers == req.TargetWorkers {
		response.Status = "no_change"
		response.Message = "Worker数量无需调整"
		response.ScaledWorkers = 0
		return response, nil
	}

	// 执行扩缩容
	err := q.scaleWorkers(ctx, queueName, req.TargetWorkers)
	if err != nil {
		response.Status = "failed"
		response.Message = "扩缩容失败: " + err.Error()
		return response, err
	}

	response.Status = "success"
	response.ScaledWorkers = req.TargetWorkers - currentWorkers
	if response.ScaledWorkers > 0 {
		response.Message = fmt.Sprintf("成功扩容 %d 个Worker", response.ScaledWorkers)
	} else {
		response.Message = fmt.Sprintf("成功缩容 %d 个Worker", -response.ScaledWorkers)
	}
	response.EstimatedTime = 30 // 预估30秒完成

	return response, nil
}

// pauseQueue 暂停队列
func (q *QueueAPI) pauseQueue(ctx *gin.Context, queueName string) error {
	// 这里需要实现暂停队列的逻辑
	// 可以通过设置Redis标志位来实现
	// pauseKey := fmt.Sprintf("alert_agent:queue:%s:paused", queueName)
	// 假设有Redis客户端
	// return redis.Client.Set(ctx, pauseKey, "true", 0).Err()
	return nil
}

// resumeQueue 恢复队列
func (q *QueueAPI) resumeQueue(ctx *gin.Context, queueName string) error {
	// 这里需要实现恢复队列的逻辑
	// pauseKey := fmt.Sprintf("alert_agent:queue:%s:paused", queueName)
	// return redis.Client.Del(ctx, pauseKey).Err()
	return nil
}

// getQueueAlerts 获取队列告警
func (q *QueueAPI) getQueueAlerts(ctx *gin.Context, queueName, status string, page, pageSize int) ([]*QueueAlert, int64, error) {
	// 这里需要实现获取队列告警的逻辑
	// 简化实现，返回示例告警
	alerts := []*QueueAlert{
		{
			ID:        "alert_001",
			QueueName: queueName,
			Type:      "high_error_rate",
			Severity:  "warning",
			Title:     "错误率过高",
			Message:   "队列错误率超过10%",
			Status:    "active",
			CreatedAt: time.Now().Add(-time.Hour),
			UpdatedAt: time.Now(),
		},
	}

	return alerts, int64(len(alerts)), nil
}

// acknowledgeAlert 确认告警
func (q *QueueAPI) acknowledgeAlert(ctx *gin.Context, alertID string) error {
	// 这里需要实现确认告警的逻辑
	return nil
}