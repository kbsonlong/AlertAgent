package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"alert_agent/internal/pkg/logger"
	"alert_agent/internal/pkg/queue"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// Helper functions
func parseInt64(s string) int64 {
	if s == "" {
		return 0
	}
	val, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0
	}
	return val
}

func parseDuration(s string) time.Duration {
	if s == "" {
		return 0
	}
	d, err := time.ParseDuration(s)
	if err != nil {
		return 0
	}
	return d
}

// QueueService 队列服务
type QueueService struct {
	queueManager *queue.RedisMessageQueue
	monitor      *queue.QueueMonitor
	client       *redis.Client
	keyPrefix    string
}

// NewQueueService 创建队列服务实例
func NewQueueService(queueManager *queue.RedisMessageQueue, monitor *queue.QueueMonitor, client *redis.Client, keyPrefix string) *QueueService {
	return &QueueService{
		queueManager: queueManager,
		monitor:      monitor,
		client:       client,
		keyPrefix:    keyPrefix,
	}
}

// TaskListFilter 任务列表过滤器
type TaskListFilter struct {
	QueueName string
	Status    string
	TaskType  string
	Page      int
	PageSize  int
	StartTime *time.Time
	EndTime   *time.Time
}

// TaskListResult 任务列表结果
type TaskListResult struct {
	Tasks    []*queue.Task `json:"tasks"`
	Total    int64         `json:"total"`
	Page     int           `json:"page"`
	PageSize int           `json:"page_size"`
}

// QueuePerformanceStats 队列性能统计
type QueuePerformanceStats struct {
	QueueName         string                    `json:"queue_name"`
	ThroughputHistory []ThroughputPoint         `json:"throughput_history"`
	ErrorRateHistory  []ErrorRatePoint          `json:"error_rate_history"`
	LatencyHistory    []LatencyPoint            `json:"latency_history"`
	WorkerStats       []WorkerPerformanceStats  `json:"worker_stats"`
	Recommendations   []PerformanceRecommendation `json:"recommendations"`
}

// ThroughputPoint 吞吐量数据点
type ThroughputPoint struct {
	Timestamp time.Time `json:"timestamp"`
	Value     float64   `json:"value"`
}

// ErrorRatePoint 错误率数据点
type ErrorRatePoint struct {
	Timestamp time.Time `json:"timestamp"`
	Value     float64   `json:"value"`
}

// LatencyPoint 延迟数据点
type LatencyPoint struct {
	Timestamp time.Time     `json:"timestamp"`
	Value     time.Duration `json:"value"`
}

// WorkerPerformanceStats Worker性能统计
type WorkerPerformanceStats struct {
	WorkerID      string        `json:"worker_id"`
	TasksHandled  int64         `json:"tasks_handled"`
	SuccessRate   float64       `json:"success_rate"`
	AvgLatency    time.Duration `json:"avg_latency"`
	LastActive    time.Time     `json:"last_active"`
	Status        string        `json:"status"`
}

// PerformanceRecommendation 性能优化建议
type PerformanceRecommendation struct {
	Type        string `json:"type"`
	Priority    string `json:"priority"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Action      string `json:"action"`
}

// GetTaskList 获取任务列表
func (s *QueueService) GetTaskList(ctx context.Context, filter *TaskListFilter) (*TaskListResult, error) {
	tasks := make([]*queue.Task, 0)
	var total int64 = 0

	// 根据过滤条件获取任务
	if filter.QueueName != "" {
		// 从指定队列获取任务
		queueTasks, queueTotal, err := s.getTasksFromQueue(ctx, filter)
		if err != nil {
			return nil, fmt.Errorf("failed to get tasks from queue: %w", err)
		}
		tasks = append(tasks, queueTasks...)
		total += queueTotal
	} else {
		// 从所有队列获取任务
		queueNames := []string{
			string(queue.TaskTypeAIAnalysis),
			string(queue.TaskTypeNotification),
			string(queue.TaskTypeConfigSync),
			string(queue.TaskTypeRuleUpdate),
			string(queue.TaskTypeHealthCheck),
		}

		for _, queueName := range queueNames {
			queueFilter := *filter
			queueFilter.QueueName = queueName
			
			queueTasks, queueTotal, err := s.getTasksFromQueue(ctx, &queueFilter)
			if err != nil {
				logger.L.Warn("Failed to get tasks from queue",
					zap.String("queue", queueName),
					zap.Error(err),
				)
				continue
			}
			tasks = append(tasks, queueTasks...)
			total += queueTotal
		}
	}

	// 应用分页
	start := (filter.Page - 1) * filter.PageSize
	end := start + filter.PageSize
	
	if start >= len(tasks) {
		tasks = []*queue.Task{}
	} else if end > len(tasks) {
		tasks = tasks[start:]
	} else {
		tasks = tasks[start:end]
	}

	return &TaskListResult{
		Tasks:    tasks,
		Total:    total,
		Page:     filter.Page,
		PageSize: filter.PageSize,
	}, nil
}

// getTasksFromQueue 从指定队列获取任务
func (s *QueueService) getTasksFromQueue(ctx context.Context, filter *TaskListFilter) ([]*queue.Task, int64, error) {
	tasks := make([]*queue.Task, 0)
	
	// 获取不同状态的任务
	statusQueues := map[string]string{
		"pending":    fmt.Sprintf("%s:queue:%s", s.keyPrefix, filter.QueueName),
		"processing": fmt.Sprintf("%s:processing:%s", s.keyPrefix, filter.QueueName),
		"failed":     fmt.Sprintf("%s:dead:%s", s.keyPrefix, filter.QueueName),
		"delayed":    fmt.Sprintf("%s:delayed:%s", s.keyPrefix, filter.QueueName),
	}

	for status, queueKey := range statusQueues {
		// 如果指定了状态过滤，跳过不匹配的状态
		if filter.Status != "" && filter.Status != status {
			continue
		}

		// 获取队列中的任务ID
		taskIDs, err := s.client.ZRevRange(ctx, queueKey, 0, -1).Result()
		if err != nil {
			if err != redis.Nil {
				logger.L.Warn("Failed to get task IDs from queue",
					zap.String("queue", queueKey),
					zap.Error(err),
				)
			}
			continue
		}

		// 获取任务详情
		for _, taskID := range taskIDs {
			task, err := s.queueManager.GetTaskStatus(ctx, taskID)
			if err != nil {
				logger.L.Warn("Failed to get task status",
					zap.String("task_id", taskID),
					zap.Error(err),
				)
				continue
			}

			if task == nil {
				continue
			}

			// 应用过滤条件
			if s.shouldIncludeTask(task, filter) {
				tasks = append(tasks, task)
			}
		}
	}

	return tasks, int64(len(tasks)), nil
}

// shouldIncludeTask 检查任务是否应该包含在结果中
func (s *QueueService) shouldIncludeTask(task *queue.Task, filter *TaskListFilter) bool {
	// 任务类型过滤
	if filter.TaskType != "" && string(task.Type) != filter.TaskType {
		return false
	}

	// 时间范围过滤
	if filter.StartTime != nil && task.CreatedAt.Before(*filter.StartTime) {
		return false
	}
	if filter.EndTime != nil && task.CreatedAt.After(*filter.EndTime) {
		return false
	}

	return true
}

// GetQueuePerformanceStats 获取队列性能统计
func (s *QueueService) GetQueuePerformanceStats(ctx context.Context, queueName string, duration time.Duration) (*QueuePerformanceStats, error) {
	stats := &QueuePerformanceStats{
		QueueName: queueName,
	}

	// 获取吞吐量历史
	throughputHistory, err := s.getThroughputHistory(ctx, queueName, duration)
	if err != nil {
		logger.L.Warn("Failed to get throughput history", zap.Error(err))
	} else {
		stats.ThroughputHistory = throughputHistory
	}

	// 获取错误率历史
	errorRateHistory, err := s.getErrorRateHistory(ctx, queueName, duration)
	if err != nil {
		logger.L.Warn("Failed to get error rate history", zap.Error(err))
	} else {
		stats.ErrorRateHistory = errorRateHistory
	}

	// 获取延迟历史
	latencyHistory, err := s.getLatencyHistory(ctx, queueName, duration)
	if err != nil {
		logger.L.Warn("Failed to get latency history", zap.Error(err))
	} else {
		stats.LatencyHistory = latencyHistory
	}

	// 获取Worker统计
	workerStats, err := s.getWorkerStats(ctx, queueName)
	if err != nil {
		logger.L.Warn("Failed to get worker stats", zap.Error(err))
	} else {
		stats.WorkerStats = workerStats
	}

	// 生成性能建议
	stats.Recommendations = s.generatePerformanceRecommendations(stats)

	return stats, nil
}

// getThroughputHistory 获取吞吐量历史
func (s *QueueService) getThroughputHistory(ctx context.Context, queueName string, duration time.Duration) ([]ThroughputPoint, error) {
	throughputKey := fmt.Sprintf("%s:throughput:%s", s.keyPrefix, queueName)
	
	now := time.Now()
	start := now.Add(-duration)
	
	// 按时间间隔分组统计
	interval := duration / 20 // 分成20个数据点
	points := make([]ThroughputPoint, 0, 20)
	
	for i := 0; i < 20; i++ {
		pointStart := start.Add(time.Duration(i) * interval)
		pointEnd := pointStart.Add(interval)
		
		count, err := s.client.ZCount(ctx, throughputKey,
			strconv.FormatInt(pointStart.Unix(), 10),
			strconv.FormatInt(pointEnd.Unix(), 10),
		).Result()
		
		if err != nil {
			logger.L.Warn("Failed to count throughput", zap.Error(err))
			count = 0
		}
		
		// 转换为每分钟的吞吐量
		throughputPerMin := float64(count) / interval.Minutes()
		
		points = append(points, ThroughputPoint{
			Timestamp: pointEnd,
			Value:     throughputPerMin,
		})
	}
	
	return points, nil
}

// getErrorRateHistory 获取错误率历史
func (s *QueueService) getErrorRateHistory(ctx context.Context, queueName string, duration time.Duration) ([]ErrorRatePoint, error) {
	// 从统计数据中获取错误率历史
	statsKey := fmt.Sprintf("%s:stats:queue:%s:history", s.keyPrefix, queueName)
	
	now := time.Now()
	start := now.Add(-duration)
	interval := duration / 20
	points := make([]ErrorRatePoint, 0, 20)
	
	for i := 0; i < 20; i++ {
		pointEnd := start.Add(time.Duration(i+1) * interval)
		
		// 获取该时间点的错误率（这里简化处理，实际应该从历史数据中获取）
		errorRate := 0.0
		
		// 从Redis获取历史错误率数据
		errorRateStr, err := s.client.HGet(ctx, statsKey, 
			fmt.Sprintf("error_rate_%d", pointEnd.Unix())).Result()
		if err == nil {
			if rate, parseErr := strconv.ParseFloat(errorRateStr, 64); parseErr == nil {
				errorRate = rate
			}
		}
		
		points = append(points, ErrorRatePoint{
			Timestamp: pointEnd,
			Value:     errorRate,
		})
	}
	
	return points, nil
}

// getLatencyHistory 获取延迟历史
func (s *QueueService) getLatencyHistory(ctx context.Context, queueName string, duration time.Duration) ([]LatencyPoint, error) {
	statsKey := fmt.Sprintf("%s:stats:queue:%s:history", s.keyPrefix, queueName)
	
	now := time.Now()
	start := now.Add(-duration)
	interval := duration / 20
	points := make([]LatencyPoint, 0, 20)
	
	for i := 0; i < 20; i++ {
		pointEnd := start.Add(time.Duration(i+1) * interval)
		
		// 获取该时间点的平均延迟
		var latency time.Duration = 0
		
		latencyStr, err := s.client.HGet(ctx, statsKey,
			fmt.Sprintf("avg_latency_%d", pointEnd.Unix())).Result()
		if err == nil {
			if d, parseErr := time.ParseDuration(latencyStr); parseErr == nil {
				latency = d
			}
		}
		
		points = append(points, LatencyPoint{
			Timestamp: pointEnd,
			Value:     latency,
		})
	}
	
	return points, nil
}

// getWorkerStats 获取Worker统计
func (s *QueueService) getWorkerStats(ctx context.Context, queueName string) ([]WorkerPerformanceStats, error) {
	workerStatsKey := fmt.Sprintf("%s:workers:%s", s.keyPrefix, queueName)
	
	// 获取所有Worker信息
	workerData, err := s.client.HGetAll(ctx, workerStatsKey).Result()
	if err != nil {
		if err == redis.Nil {
			return []WorkerPerformanceStats{}, nil
		}
		return nil, fmt.Errorf("failed to get worker stats: %w", err)
	}
	
	stats := make([]WorkerPerformanceStats, 0, len(workerData))
	
	for workerID, data := range workerData {
		var workerStat WorkerPerformanceStats
		
		// 解析Worker数据（假设数据以JSON格式存储）
		parts := strings.Split(data, ",")
		if len(parts) >= 4 {
			workerStat.WorkerID = workerID
			
			if tasksHandled, err := strconv.ParseInt(parts[0], 10, 64); err == nil {
				workerStat.TasksHandled = tasksHandled
			}
			
			if successRate, err := strconv.ParseFloat(parts[1], 64); err == nil {
				workerStat.SuccessRate = successRate
			}
			
			if avgLatency, err := time.ParseDuration(parts[2]); err == nil {
				workerStat.AvgLatency = avgLatency
			}
			
			if lastActive, err := time.Parse(time.RFC3339, parts[3]); err == nil {
				workerStat.LastActive = lastActive
				
				// 判断Worker状态
				if time.Since(lastActive) < 5*time.Minute {
					workerStat.Status = "active"
				} else if time.Since(lastActive) < 30*time.Minute {
					workerStat.Status = "idle"
				} else {
					workerStat.Status = "inactive"
				}
			}
		}
		
		stats = append(stats, workerStat)
	}
	
	return stats, nil
}

// generatePerformanceRecommendations 生成性能优化建议
func (s *QueueService) generatePerformanceRecommendations(stats *QueuePerformanceStats) []PerformanceRecommendation {
	recommendations := make([]PerformanceRecommendation, 0)
	
	// 分析吞吐量趋势
	if len(stats.ThroughputHistory) > 0 {
		avgThroughput := 0.0
		for _, point := range stats.ThroughputHistory {
			avgThroughput += point.Value
		}
		avgThroughput /= float64(len(stats.ThroughputHistory))
		
		if avgThroughput < 10 {
			recommendations = append(recommendations, PerformanceRecommendation{
				Type:        "throughput",
				Priority:    "medium",
				Title:       "吞吐量较低",
				Description: fmt.Sprintf("队列 %s 的平均吞吐量为 %.2f 任务/分钟，建议优化处理逻辑", stats.QueueName, avgThroughput),
				Action:      "考虑增加Worker数量或优化任务处理逻辑",
			})
		}
	}
	
	// 分析错误率
	if len(stats.ErrorRateHistory) > 0 {
		avgErrorRate := 0.0
		for _, point := range stats.ErrorRateHistory {
			avgErrorRate += point.Value
		}
		avgErrorRate /= float64(len(stats.ErrorRateHistory))
		
		if avgErrorRate > 5 {
			recommendations = append(recommendations, PerformanceRecommendation{
				Type:        "error_rate",
				Priority:    "high",
				Title:       "错误率过高",
				Description: fmt.Sprintf("队列 %s 的平均错误率为 %.2f%%，需要关注任务失败原因", stats.QueueName, avgErrorRate),
				Action:      "检查任务失败日志，优化错误处理逻辑",
			})
		}
	}
	
	// 分析Worker状态
	activeWorkers := 0
	totalWorkers := len(stats.WorkerStats)
	
	for _, worker := range stats.WorkerStats {
		if worker.Status == "active" {
			activeWorkers++
		}
	}
	
	if totalWorkers > 0 {
		activeRatio := float64(activeWorkers) / float64(totalWorkers)
		if activeRatio < 0.5 {
			recommendations = append(recommendations, PerformanceRecommendation{
				Type:        "worker",
				Priority:    "medium",
				Title:       "Worker活跃度低",
				Description: fmt.Sprintf("队列 %s 只有 %d/%d Worker处于活跃状态", stats.QueueName, activeWorkers, totalWorkers),
				Action:      "检查Worker健康状态，考虑重启不活跃的Worker",
			})
		}
	}
	
	// 分析延迟
	if len(stats.LatencyHistory) > 0 {
		avgLatency := time.Duration(0)
		for _, point := range stats.LatencyHistory {
			avgLatency += point.Value
		}
		avgLatency /= time.Duration(len(stats.LatencyHistory))
		
		if avgLatency > 30*time.Second {
			recommendations = append(recommendations, PerformanceRecommendation{
				Type:        "latency",
				Priority:    "medium",
				Title:       "处理延迟较高",
				Description: fmt.Sprintf("队列 %s 的平均处理延迟为 %v，可能影响系统响应性能", stats.QueueName, avgLatency),
				Action:      "优化任务处理逻辑，考虑并行处理或缓存机制",
			})
		}
	}
	
	return recommendations
}

// OptimizeQueue 队列优化
func (s *QueueService) OptimizeQueue(ctx context.Context, queueName string, options map[string]interface{}) error {
	// 获取当前队列指标
	metrics, err := s.monitor.GetQueueMetrics(ctx, queueName)
	if err != nil {
		return fmt.Errorf("failed to get queue metrics: %w", err)
	}
	
	// 根据指标和选项执行优化操作
	if autoScale, ok := options["auto_scale"].(bool); ok && autoScale {
		err = s.autoScaleWorkers(ctx, queueName, metrics)
		if err != nil {
			logger.L.Warn("Failed to auto scale workers", zap.Error(err))
		}
	}
	
	if cleanupExpired, ok := options["cleanup_expired"].(bool); ok && cleanupExpired {
		maxAge := 24 * time.Hour
		if maxAgeStr, ok := options["max_age"].(string); ok {
			if duration, parseErr := time.ParseDuration(maxAgeStr); parseErr == nil {
				maxAge = duration
			}
		}
		
		err = s.monitor.CleanupExpiredTasks(ctx, queueName, maxAge)
		if err != nil {
			logger.L.Warn("Failed to cleanup expired tasks", zap.Error(err))
		}
	}
	
	if rebalance, ok := options["rebalance"].(bool); ok && rebalance {
		err = s.rebalanceQueue(ctx, queueName)
		if err != nil {
			logger.L.Warn("Failed to rebalance queue", zap.Error(err))
		}
	}
	
	return nil
}

// autoScaleWorkers 自动扩缩容Worker
func (s *QueueService) autoScaleWorkers(ctx context.Context, queueName string, metrics *queue.QueueMetrics) error {
	// 简单的扩缩容逻辑
	targetWorkers := 1
	
	// 根据待处理任务数量决定Worker数量
	if metrics.PendingCount > 100 {
		targetWorkers = 5
	} else if metrics.PendingCount > 50 {
		targetWorkers = 3
	} else if metrics.PendingCount > 10 {
		targetWorkers = 2
	}
	
	// 记录扩缩容建议（实际的Worker扩缩容需要与容器编排系统集成）
	scaleKey := fmt.Sprintf("%s:scale:%s", s.keyPrefix, queueName)
	scaleInfo := map[string]interface{}{
		"target_workers": targetWorkers,
		"current_pending": metrics.PendingCount,
		"timestamp": time.Now().Unix(),
	}
	
	scaleData, _ := json.Marshal(scaleInfo)
	err := s.client.Set(ctx, scaleKey, scaleData, time.Hour).Err()
	if err != nil {
		return fmt.Errorf("failed to record scale info: %w", err)
	}
	
	logger.L.Info("Auto scale recommendation generated",
		zap.String("queue", queueName),
		zap.Int("target_workers", targetWorkers),
		zap.Int64("pending_count", metrics.PendingCount),
	)
	
	return nil
}

// rebalanceQueue 重新平衡队列
func (s *QueueService) rebalanceQueue(ctx context.Context, queueName string) error {
	// 重新平衡队列中的任务优先级
	queueKey := fmt.Sprintf("%s:queue:%s", s.keyPrefix, queueName)
	
	// 获取所有任务
	taskIDs, err := s.client.ZRange(ctx, queueKey, 0, -1).Result()
	if err != nil {
		return fmt.Errorf("failed to get tasks for rebalancing: %w", err)
	}
	
	// 重新计算优先级并更新
	pipe := s.client.Pipeline()
	
	for _, taskID := range taskIDs {
		task, err := s.queueManager.GetTaskStatus(ctx, taskID)
		if err != nil {
			continue
		}
		
		if task == nil {
			continue
		}
		
		// 根据任务年龄调整优先级
		age := time.Since(task.CreatedAt)
		newPriority := float64(task.Priority)
		
		// 老任务提高优先级
		if age > time.Hour {
			newPriority += 1
		}
		if age > 6*time.Hour {
			newPriority += 2
		}
		
		pipe.ZAdd(ctx, queueKey, redis.Z{
			Score:  newPriority,
			Member: taskID,
		})
	}
	
	_, err = pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to rebalance queue: %w", err)
	}
	
	logger.L.Info("Queue rebalanced",
		zap.String("queue", queueName),
		zap.Int("tasks_rebalanced", len(taskIDs)),
	)
	
	return nil
}
// TaskSearchOptions 任务搜索选项
type TaskSearchOptions struct {
	Filter      *TaskListFilter
	SortBy      string // created_at, updated_at, priority, status
	SortOrder   string // asc, desc
	IncludePayload bool
	IncludeLogs    bool
}

// TaskBatchOperation 批量任务操作
type TaskBatchOperation struct {
	TaskIDs   []string               `json:"task_ids"`
	Operation string                 `json:"operation"` // retry, skip, cancel, delete
	Options   map[string]interface{} `json:"options,omitempty"`
}

// TaskBatchResult 批量操作结果
type TaskBatchResult struct {
	Total     int                    `json:"total"`
	Succeeded int                    `json:"succeeded"`
	Failed    int                    `json:"failed"`
	Results   []TaskOperationResult  `json:"results"`
	Summary   map[string]interface{} `json:"summary"`
}

// TaskOperationResult 任务操作结果
type TaskOperationResult struct {
	TaskID    string                 `json:"task_id"`
	Success   bool                   `json:"success"`
	Error     string                 `json:"error,omitempty"`
	Details   map[string]interface{} `json:"details,omitempty"`
}

// SearchTasks 高级任务搜索
func (s *QueueService) SearchTasks(ctx context.Context, options *TaskSearchOptions) (*TaskListResult, error) {
	// 基础任务列表获取
	result, err := s.GetTaskList(ctx, options.Filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get task list: %w", err)
	}

	// 应用排序
	if options.SortBy != "" {
		s.sortTasks(result.Tasks, options.SortBy, options.SortOrder)
	}

	// 如果需要包含载荷和日志，进行额外处理
	if options.IncludePayload || options.IncludeLogs {
		for _, task := range result.Tasks {
			if options.IncludeLogs {
				// 获取任务日志
				logs, err := s.getTaskLogs(ctx, task.ID, 1, 10)
				if err != nil {
					logger.L.Warn("Failed to get task logs",
						zap.String("task_id", task.ID),
						zap.Error(err),
					)
				} else {
					// 将日志添加到任务的额外字段中
					if task.Payload == nil {
						task.Payload = make(map[string]interface{})
					}
					task.Payload["_logs"] = logs
				}
			}
		}
	}

	return result, nil
}

// BatchOperateTasks 批量操作任务
func (s *QueueService) BatchOperateTasks(ctx context.Context, operation *TaskBatchOperation) (*TaskBatchResult, error) {
	result := &TaskBatchResult{
		Total:     len(operation.TaskIDs),
		Succeeded: 0,
		Failed:    0,
		Results:   make([]TaskOperationResult, 0, len(operation.TaskIDs)),
		Summary:   make(map[string]interface{}),
	}

	// 统计各种状态的任务数量
	statusCount := make(map[string]int)

	for _, taskID := range operation.TaskIDs {
		opResult := TaskOperationResult{
			TaskID:  taskID,
			Details: make(map[string]interface{}),
		}

		// 获取任务详情
		task, err := s.queueManager.GetTaskStatus(ctx, taskID)
		if err != nil {
			opResult.Success = false
			opResult.Error = "获取任务状态失败: " + err.Error()
			result.Failed++
		} else if task == nil {
			opResult.Success = false
			opResult.Error = "任务不存在"
			result.Failed++
		} else {
			// 记录任务状态
			statusCount[string(task.Status)]++
			opResult.Details["original_status"] = task.Status
			opResult.Details["task_type"] = task.Type

			// 执行具体操作
			switch operation.Operation {
			case "retry":
				err = s.retryTask(ctx, task)
			case "skip":
				err = s.skipTask(ctx, task)
			case "cancel":
				err = s.cancelTask(ctx, task)
			case "delete":
				err = s.deleteTask(ctx, task)
			default:
				err = fmt.Errorf("unsupported operation: %s", operation.Operation)
			}

			if err != nil {
				opResult.Success = false
				opResult.Error = err.Error()
				result.Failed++
			} else {
				opResult.Success = true
				result.Succeeded++
			}
		}

		result.Results = append(result.Results, opResult)
	}

	// 生成操作摘要
	result.Summary["operation"] = operation.Operation
	result.Summary["status_distribution"] = statusCount
	result.Summary["success_rate"] = float64(result.Succeeded) / float64(result.Total) * 100

	return result, nil
}

// GetTaskAnalytics 获取任务分析数据
func (s *QueueService) GetTaskAnalytics(ctx context.Context, queueName string, period time.Duration) (*TaskAnalytics, error) {
	analytics := &TaskAnalytics{
		QueueName: queueName,
		Period:    period,
		GeneratedAt: time.Now(),
	}

	// 获取任务统计
	stats, err := s.getTaskStatistics(ctx, queueName, period)
	if err != nil {
		return nil, fmt.Errorf("failed to get task statistics: %w", err)
	}
	analytics.Statistics = stats

	// 获取性能趋势
	trends, err := s.getPerformanceTrends(ctx, queueName, period)
	if err != nil {
		return nil, fmt.Errorf("failed to get performance trends: %w", err)
	}
	analytics.Trends = trends

	// 获取错误分析
	errorAnalysis, err := s.getErrorAnalysis(ctx, queueName, period)
	if err != nil {
		return nil, fmt.Errorf("failed to get error analysis: %w", err)
	}
	analytics.ErrorAnalysis = errorAnalysis

	// 生成优化建议
	analytics.Recommendations = s.generateOptimizationRecommendations(analytics)

	return analytics, nil
}

// 辅助方法

// sortTasks 对任务进行排序
func (s *QueueService) sortTasks(tasks []*queue.Task, sortBy, sortOrder string) {
	if len(tasks) == 0 {
		return
	}

	// 这里应该实现具体的排序逻辑
	// 为了简化，这里只是一个示例
	logger.L.Debug("Sorting tasks",
		zap.String("sort_by", sortBy),
		zap.String("sort_order", sortOrder),
		zap.Int("task_count", len(tasks)),
	)
}

// retryTask 重试单个任务
func (s *QueueService) retryTask(ctx context.Context, task *queue.Task) error {
	if task.Status != queue.TaskStatusFailed {
		return fmt.Errorf("只能重试失败的任务")
	}

	// 重置任务状态
	task.Status = queue.TaskStatusPending
	task.Retry = 0
	task.ErrorMsg = ""
	task.UpdatedAt = time.Now()

	// 重新发布任务
	return s.queueManager.Publish(ctx, string(task.Type), task)
}

// skipTask 跳过单个任务
func (s *QueueService) skipTask(ctx context.Context, task *queue.Task) error {
	if task.Status != queue.TaskStatusPending && task.Status != queue.TaskStatusFailed {
		return fmt.Errorf("只能跳过待处理或失败的任务")
	}

	// 使用Nack将任务移动到死信队列
	return s.queueManager.Nack(ctx, task, false)
}

// cancelTask 取消单个任务
func (s *QueueService) cancelTask(ctx context.Context, task *queue.Task) error {
	if task.Status != queue.TaskStatusProcessing {
		return fmt.Errorf("只能取消正在处理的任务")
	}

	// 标记任务为取消状态
	task.MarkFailed("Task cancelled by user")
	return s.queueManager.Nack(ctx, task, false)
}

// deleteTask 删除单个任务
func (s *QueueService) deleteTask(ctx context.Context, task *queue.Task) error {
	// 从Redis中删除任务相关的所有键
	taskKey := fmt.Sprintf("%s:task:%s", s.keyPrefix, task.ID)
	resultKey := fmt.Sprintf("%s:result:%s", s.keyPrefix, task.ID)
	
	pipe := s.client.Pipeline()
	pipe.Del(ctx, taskKey)
	pipe.Del(ctx, resultKey)
	
	// 从各个队列中移除任务
	queueKeys := []string{
		fmt.Sprintf("%s:queue:%s", s.keyPrefix, string(task.Type)),
		fmt.Sprintf("%s:processing:%s", s.keyPrefix, string(task.Type)),
		fmt.Sprintf("%s:delayed:%s", s.keyPrefix, string(task.Type)),
		fmt.Sprintf("%s:dead:%s", s.keyPrefix, string(task.Type)),
	}
	
	for _, queueKey := range queueKeys {
		pipe.ZRem(ctx, queueKey, task.ID)
	}
	
	_, err := pipe.Exec(ctx)
	return err
}

// getTaskLogs 获取任务日志
func (s *QueueService) getTaskLogs(ctx context.Context, taskID string, page, pageSize int) ([]*TaskLog, error) {
	logsKey := fmt.Sprintf("%s:logs:%s", s.keyPrefix, taskID)
	
	// 从Redis获取日志数据
	start := int64((page - 1) * pageSize)
	end := start + int64(pageSize) - 1
	
	logEntries, err := s.client.LRange(ctx, logsKey, start, end).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get task logs: %w", err)
	}
	
	logs := make([]*TaskLog, 0, len(logEntries))
	for _, entry := range logEntries {
		var log TaskLog
		if err := json.Unmarshal([]byte(entry), &log); err != nil {
			logger.L.Warn("Failed to unmarshal log entry", zap.Error(err))
			continue
		}
		logs = append(logs, &log)
	}
	
	return logs, nil
}

// getTaskStatistics 获取任务统计信息
func (s *QueueService) getTaskStatistics(ctx context.Context, queueName string, period time.Duration) (*TaskStatistics, error) {
	stats := &TaskStatistics{
		QueueName: queueName,
		Period:    period,
	}

	// 从Redis获取统计数据
	statsKey := fmt.Sprintf("%s:stats:queue:%s", s.keyPrefix, queueName)
	
	pipe := s.client.Pipeline()
	totalCmd := pipe.HGet(ctx, statsKey, "total")
	completedCmd := pipe.HGet(ctx, statsKey, "completed")
	failedCmd := pipe.HGet(ctx, statsKey, "failed")
	avgTimeCmd := pipe.HGet(ctx, statsKey, "avg_time")
	
	_, err := pipe.Exec(ctx)
	if err != nil && err != redis.Nil {
		return nil, fmt.Errorf("failed to get task statistics: %w", err)
	}

	stats.TotalTasks = parseInt64(totalCmd.Val())
	stats.CompletedTasks = parseInt64(completedCmd.Val())
	stats.FailedTasks = parseInt64(failedCmd.Val())
	stats.AvgProcessingTime = parseDuration(avgTimeCmd.Val())

	if stats.TotalTasks > 0 {
		stats.SuccessRate = float64(stats.CompletedTasks) / float64(stats.TotalTasks) * 100
	}

	return stats, nil
}

// getPerformanceTrends 获取性能趋势
func (s *QueueService) getPerformanceTrends(ctx context.Context, queueName string, period time.Duration) (*PerformanceTrends, error) {
	trends := &PerformanceTrends{
		QueueName: queueName,
		Period:    period,
	}

	// 获取吞吐量趋势
	throughputTrend, err := s.getThroughputTrend(ctx, queueName, period)
	if err != nil {
		logger.L.Warn("Failed to get throughput trend", zap.Error(err))
	} else {
		trends.ThroughputTrend = throughputTrend
	}

	// 获取延迟趋势
	latencyTrend, err := s.getLatencyTrend(ctx, queueName, period)
	if err != nil {
		logger.L.Warn("Failed to get latency trend", zap.Error(err))
	} else {
		trends.LatencyTrend = latencyTrend
	}

	// 获取错误率趋势
	errorRateTrend, err := s.getErrorRateTrend(ctx, queueName, period)
	if err != nil {
		logger.L.Warn("Failed to get error rate trend", zap.Error(err))
	} else {
		trends.ErrorRateTrend = errorRateTrend
	}

	return trends, nil
}

// getErrorAnalysis 获取错误分析
func (s *QueueService) getErrorAnalysis(ctx context.Context, queueName string, period time.Duration) (*ErrorAnalysis, error) {
	analysis := &ErrorAnalysis{
		QueueName: queueName,
		Period:    period,
	}

	// 获取错误分布
	errorDistribution, err := s.getErrorDistribution(ctx, queueName, period)
	if err != nil {
		logger.L.Warn("Failed to get error distribution", zap.Error(err))
	} else {
		analysis.ErrorDistribution = errorDistribution
	}

	// 获取常见错误
	commonErrors, err := s.getCommonErrors(ctx, queueName, period)
	if err != nil {
		logger.L.Warn("Failed to get common errors", zap.Error(err))
	} else {
		analysis.CommonErrors = commonErrors
	}

	return analysis, nil
}

// generateOptimizationRecommendations 生成优化建议
func (s *QueueService) generateOptimizationRecommendations(analytics *TaskAnalytics) []*OptimizationRecommendation {
	recommendations := make([]*OptimizationRecommendation, 0)

	// 基于成功率的建议
	if analytics.Statistics.SuccessRate < 90 {
		recommendations = append(recommendations, &OptimizationRecommendation{
			Type:        "success_rate",
			Priority:    "high",
			Title:       "任务成功率偏低",
			Description: fmt.Sprintf("队列 %s 的任务成功率为 %.1f%%，低于推荐的90%%", analytics.QueueName, analytics.Statistics.SuccessRate),
			Action:      "检查任务失败原因，优化任务处理逻辑",
			Impact:      "提高任务成功率可以减少重试开销，提升系统稳定性",
		})
	}

	// 基于平均处理时间的建议
	if analytics.Statistics.AvgProcessingTime > 30*time.Second {
		recommendations = append(recommendations, &OptimizationRecommendation{
			Type:        "processing_time",
			Priority:    "medium",
			Title:       "任务处理时间较长",
			Description: fmt.Sprintf("队列 %s 的平均处理时间为 %v，建议优化处理逻辑", analytics.QueueName, analytics.Statistics.AvgProcessingTime),
			Action:      "分析任务处理瓶颈，考虑并行处理或缓存优化",
			Impact:      "减少处理时间可以提高系统吞吐量",
		})
	}

	return recommendations
}

// 辅助方法实现（简化版本）
func (s *QueueService) getThroughputTrend(ctx context.Context, queueName string, period time.Duration) ([]*TrendPoint, error) {
	// 实现吞吐量趋势获取逻辑
	return []*TrendPoint{}, nil
}

func (s *QueueService) getLatencyTrend(ctx context.Context, queueName string, period time.Duration) ([]*TrendPoint, error) {
	// 实现延迟趋势获取逻辑
	return []*TrendPoint{}, nil
}

func (s *QueueService) getErrorRateTrend(ctx context.Context, queueName string, period time.Duration) ([]*TrendPoint, error) {
	// 实现错误率趋势获取逻辑
	return []*TrendPoint{}, nil
}

func (s *QueueService) getErrorDistribution(ctx context.Context, queueName string, period time.Duration) (map[string]int64, error) {
	// 实现错误分布获取逻辑
	return make(map[string]int64), nil
}

func (s *QueueService) getCommonErrors(ctx context.Context, queueName string, period time.Duration) ([]*CommonError, error) {
	// 实现常见错误获取逻辑
	return []*CommonError{}, nil
}

// 数据结构定义

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

// TaskAnalytics 任务分析数据
type TaskAnalytics struct {
	QueueName       string                        `json:"queue_name"`
	Period          time.Duration                 `json:"period"`
	GeneratedAt     time.Time                     `json:"generated_at"`
	Statistics      *TaskStatistics               `json:"statistics"`
	Trends          *PerformanceTrends            `json:"trends"`
	ErrorAnalysis   *ErrorAnalysis                `json:"error_analysis"`
	Recommendations []*OptimizationRecommendation `json:"recommendations"`
}

// TaskStatistics 任务统计
type TaskStatistics struct {
	QueueName          string        `json:"queue_name"`
	Period             time.Duration `json:"period"`
	TotalTasks         int64         `json:"total_tasks"`
	CompletedTasks     int64         `json:"completed_tasks"`
	FailedTasks        int64         `json:"failed_tasks"`
	AvgProcessingTime  time.Duration `json:"avg_processing_time"`
	SuccessRate        float64       `json:"success_rate"`
}

// PerformanceTrends 性能趋势
type PerformanceTrends struct {
	QueueName       string        `json:"queue_name"`
	Period          time.Duration `json:"period"`
	ThroughputTrend []*TrendPoint `json:"throughput_trend"`
	LatencyTrend    []*TrendPoint `json:"latency_trend"`
	ErrorRateTrend  []*TrendPoint `json:"error_rate_trend"`
}

// TrendPoint 趋势数据点
type TrendPoint struct {
	Timestamp time.Time `json:"timestamp"`
	Value     float64   `json:"value"`
}

// ErrorAnalysis 错误分析
type ErrorAnalysis struct {
	QueueName         string                `json:"queue_name"`
	Period            time.Duration         `json:"period"`
	ErrorDistribution map[string]int64      `json:"error_distribution"`
	CommonErrors      []*CommonError        `json:"common_errors"`
}

// CommonError 常见错误
type CommonError struct {
	Message     string    `json:"message"`
	Count       int64     `json:"count"`
	FirstSeen   time.Time `json:"first_seen"`
	LastSeen    time.Time `json:"last_seen"`
	Percentage  float64   `json:"percentage"`
	Suggestions []string  `json:"suggestions"`
}

// OptimizationRecommendation 优化建议
type OptimizationRecommendation struct {
	Type        string `json:"type"`
	Priority    string `json:"priority"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Action      string `json:"action"`
	Impact      string `json:"impact"`
}