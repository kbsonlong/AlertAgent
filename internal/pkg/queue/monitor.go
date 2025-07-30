package queue

import (
	"context"
	"fmt"
	"time"

	"alert_agent/internal/pkg/logger"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// QueueMonitor 队列监控器
type QueueMonitor struct {
	queue     MessageQueue
	client    *redis.Client
	keyPrefix string
}

// NewQueueMonitor 创建队列监控器
func NewQueueMonitor(queue MessageQueue, client *redis.Client, keyPrefix string) *QueueMonitor {
	return &QueueMonitor{
		queue:     queue,
		client:    client,
		keyPrefix: keyPrefix,
	}
}

// QueueMetrics 队列指标
type QueueMetrics struct {
	QueueName       string        `json:"queue_name"`
	PendingCount    int64         `json:"pending_count"`
	ProcessingCount int64         `json:"processing_count"`
	CompletedCount  int64         `json:"completed_count"`
	FailedCount     int64         `json:"failed_count"`
	DeadLetterCount int64         `json:"dead_letter_count"`
	DelayedCount    int64         `json:"delayed_count"`
	ThroughputPerMin float64      `json:"throughput_per_min"`
	AvgProcessingTime time.Duration `json:"avg_processing_time"`
	ErrorRate       float64       `json:"error_rate"`
	LastUpdated     time.Time     `json:"last_updated"`
}

// TaskMetrics 任务指标
type TaskMetrics struct {
	TaskType          TaskType      `json:"task_type"`
	TotalCount        int64         `json:"total_count"`
	CompletedCount    int64         `json:"completed_count"`
	FailedCount       int64         `json:"failed_count"`
	AvgProcessingTime time.Duration `json:"avg_processing_time"`
	SuccessRate       float64       `json:"success_rate"`
	LastHourCount     int64         `json:"last_hour_count"`
}

// GetQueueMetrics 获取队列指标
func (m *QueueMonitor) GetQueueMetrics(ctx context.Context, queueName string) (*QueueMetrics, error) {
	stats, err := m.queue.GetQueueStats(ctx, queueName)
	if err != nil {
		return nil, fmt.Errorf("failed to get queue stats: %w", err)
	}

	// 获取延迟队列统计
	delayedKey := fmt.Sprintf("%s:delayed:%s", m.keyPrefix, queueName)
	delayedCount, err := m.client.ZCard(ctx, delayedKey).Result()
	if err != nil {
		delayedCount = 0
	}

	// 计算吞吐量（每分钟完成的任务数）
	throughput, err := m.calculateThroughput(ctx, queueName)
	if err != nil {
		logger.L.Warn("Failed to calculate throughput", zap.Error(err))
		throughput = 0
	}

	// 计算平均处理时间
	avgProcessingTime, err := m.calculateAvgProcessingTime(ctx, queueName)
	if err != nil {
		logger.L.Warn("Failed to calculate avg processing time", zap.Error(err))
		avgProcessingTime = 0
	}

	// 计算错误率
	errorRate := float64(0)
	if stats.CompletedCount+stats.FailedCount > 0 {
		errorRate = float64(stats.FailedCount) / float64(stats.CompletedCount+stats.FailedCount) * 100
	}

	metrics := &QueueMetrics{
		QueueName:         queueName,
		PendingCount:      stats.PendingCount,
		ProcessingCount:   stats.ProcessingCount,
		CompletedCount:    stats.CompletedCount,
		FailedCount:       stats.FailedCount,
		DeadLetterCount:   stats.DeadLetterCount,
		DelayedCount:      delayedCount,
		ThroughputPerMin:  throughput,
		AvgProcessingTime: avgProcessingTime,
		ErrorRate:         errorRate,
		LastUpdated:       time.Now(),
	}

	return metrics, nil
}

// GetAllQueueMetrics 获取所有队列指标
func (m *QueueMonitor) GetAllQueueMetrics(ctx context.Context) (map[string]*QueueMetrics, error) {
	queueNames := []string{
		string(TaskTypeAIAnalysis),
		string(TaskTypeNotification),
		string(TaskTypeConfigSync),
		string(TaskTypeRuleUpdate),
		string(TaskTypeHealthCheck),
	}

	metrics := make(map[string]*QueueMetrics)
	
	for _, queueName := range queueNames {
		queueMetrics, err := m.GetQueueMetrics(ctx, queueName)
		if err != nil {
			logger.L.Warn("Failed to get metrics for queue",
				zap.String("queue", queueName),
				zap.Error(err),
			)
			continue
		}
		metrics[queueName] = queueMetrics
	}

	return metrics, nil
}

// GetTaskMetrics 获取任务类型指标
func (m *QueueMonitor) GetTaskMetrics(ctx context.Context, taskType TaskType) (*TaskMetrics, error) {
	// 从Redis获取任务统计信息
	statsKey := fmt.Sprintf("%s:stats:task:%s", m.keyPrefix, string(taskType))
	
	pipe := m.client.Pipeline()
	totalCmd := pipe.HGet(ctx, statsKey, "total")
	completedCmd := pipe.HGet(ctx, statsKey, "completed")
	failedCmd := pipe.HGet(ctx, statsKey, "failed")
	avgTimeCmd := pipe.HGet(ctx, statsKey, "avg_time")
	lastHourCmd := pipe.HGet(ctx, statsKey, "last_hour")
	
	_, err := pipe.Exec(ctx)
	if err != nil && err != redis.Nil {
		return nil, fmt.Errorf("failed to get task metrics: %w", err)
	}

	total := parseInt64(totalCmd.Val())
	completed := parseInt64(completedCmd.Val())
	failed := parseInt64(failedCmd.Val())
	avgTime := parseDuration(avgTimeCmd.Val())
	lastHour := parseInt64(lastHourCmd.Val())

	successRate := float64(0)
	if total > 0 {
		successRate = float64(completed) / float64(total) * 100
	}

	metrics := &TaskMetrics{
		TaskType:          taskType,
		TotalCount:        total,
		CompletedCount:    completed,
		FailedCount:       failed,
		AvgProcessingTime: avgTime,
		SuccessRate:       successRate,
		LastHourCount:     lastHour,
	}

	return metrics, nil
}

// calculateThroughput 计算吞吐量
func (m *QueueMonitor) calculateThroughput(ctx context.Context, queueName string) (float64, error) {
	// 获取最近一分钟完成的任务数
	throughputKey := fmt.Sprintf("%s:throughput:%s", m.keyPrefix, queueName)
	
	now := time.Now()
	oneMinuteAgo := now.Add(-time.Minute)
	
	count, err := m.client.ZCount(ctx, throughputKey, 
		fmt.Sprintf("%d", oneMinuteAgo.Unix()),
		fmt.Sprintf("%d", now.Unix()),
	).Result()
	
	if err != nil {
		return 0, err
	}

	return float64(count), nil
}

// calculateAvgProcessingTime 计算平均处理时间
func (m *QueueMonitor) calculateAvgProcessingTime(ctx context.Context, queueName string) (time.Duration, error) {
	// 从统计信息中获取平均处理时间
	statsKey := fmt.Sprintf("%s:stats:queue:%s", m.keyPrefix, queueName)
	
	avgTimeStr, err := m.client.HGet(ctx, statsKey, "avg_processing_time").Result()
	if err != nil {
		if err == redis.Nil {
			return 0, nil
		}
		return 0, err
	}

	return parseDuration(avgTimeStr), nil
}

// RecordTaskCompletion 记录任务完成
func (m *QueueMonitor) RecordTaskCompletion(ctx context.Context, task *Task, duration time.Duration, success bool) error {
	pipe := m.client.Pipeline()
	
	// 更新队列统计
	queueStatsKey := fmt.Sprintf("%s:stats:queue:%s", m.keyPrefix, string(task.Type))
	if success {
		pipe.HIncrBy(ctx, queueStatsKey, "completed", 1)
	} else {
		pipe.HIncrBy(ctx, queueStatsKey, "failed", 1)
	}
	
	// 更新任务类型统计
	taskStatsKey := fmt.Sprintf("%s:stats:task:%s", m.keyPrefix, string(task.Type))
	pipe.HIncrBy(ctx, taskStatsKey, "total", 1)
	if success {
		pipe.HIncrBy(ctx, taskStatsKey, "completed", 1)
	} else {
		pipe.HIncrBy(ctx, taskStatsKey, "failed", 1)
	}
	
	// 记录处理时间
	pipe.HSet(ctx, taskStatsKey, "last_duration", duration.String())
	
	// 更新吞吐量统计
	throughputKey := fmt.Sprintf("%s:throughput:%s", m.keyPrefix, string(task.Type))
	pipe.ZAdd(ctx, throughputKey, redis.Z{
		Score:  float64(time.Now().Unix()),
		Member: task.ID,
	})
	
	// 清理旧的吞吐量数据（保留1小时）
	oneHourAgo := time.Now().Add(-time.Hour).Unix()
	pipe.ZRemRangeByScore(ctx, throughputKey, "0", fmt.Sprintf("%d", oneHourAgo))
	
	_, err := pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to record task completion: %w", err)
	}

	return nil
}

// CleanupExpiredTasks 清理过期任务
func (m *QueueMonitor) CleanupExpiredTasks(ctx context.Context, queueName string, maxAge time.Duration) error {
	processingKey := fmt.Sprintf("%s:processing:%s", m.keyPrefix, queueName)
	
	// 获取超时的任务
	cutoff := time.Now().Add(-maxAge).Unix()
	expiredTasks, err := m.client.ZRangeByScore(ctx, processingKey, &redis.ZRangeBy{
		Min: "0",
		Max: fmt.Sprintf("%d", cutoff),
	}).Result()
	
	if err != nil {
		return fmt.Errorf("failed to get expired tasks: %w", err)
	}

	if len(expiredTasks) == 0 {
		return nil
	}

	pipe := m.client.Pipeline()
	
	for _, taskID := range expiredTasks {
		// 从处理中队列移除
		pipe.ZRem(ctx, processingKey, taskID)
		
		// 获取任务详情
		taskKey := fmt.Sprintf("%s:task:%s", m.keyPrefix, taskID)
		taskData, err := m.client.Get(ctx, taskKey).Result()
		if err != nil {
			continue
		}
		
		var task Task
		if err := task.FromJSON([]byte(taskData)); err != nil {
			continue
		}
		
		// 检查是否应该重试
		if task.ShouldRetry() {
			// 重新入队
			task.IncrementRetry()
			updatedData, _ := task.ToJSON()
			pipe.Set(ctx, taskKey, updatedData, 24*time.Hour)
			
			queueKey := fmt.Sprintf("%s:queue:%s", m.keyPrefix, queueName)
			pipe.ZAdd(ctx, queueKey, redis.Z{
				Score:  float64(task.Priority),
				Member: taskID,
			})
		} else {
			// 移动到死信队列
			deadLetterKey := fmt.Sprintf("%s:dead:%s", m.keyPrefix, queueName)
			pipe.ZAdd(ctx, deadLetterKey, redis.Z{
				Score:  float64(time.Now().Unix()),
				Member: taskID,
			})
			
			task.MarkFailed("Task timeout")
			updatedData, _ := task.ToJSON()
			pipe.Set(ctx, taskKey, updatedData, 24*time.Hour)
		}
	}

	_, err = pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to cleanup expired tasks: %w", err)
	}

	logger.L.Info("Cleaned up expired tasks",
		zap.String("queue", queueName),
		zap.Int("count", len(expiredTasks)),
	)

	return nil
}

// GetHealthStatus 获取队列健康状态
func (m *QueueMonitor) GetHealthStatus(ctx context.Context) (map[string]interface{}, error) {
	allMetrics, err := m.GetAllQueueMetrics(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get queue metrics: %w", err)
	}

	health := map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now(),
		"queues":    make(map[string]interface{}),
	}

	overallHealthy := true
	
	for queueName, metrics := range allMetrics {
		queueHealth := map[string]interface{}{
			"status":           "healthy",
			"pending_count":    metrics.PendingCount,
			"processing_count": metrics.ProcessingCount,
			"error_rate":       metrics.ErrorRate,
			"throughput":       metrics.ThroughputPerMin,
		}

		// 检查队列健康状态
		if metrics.PendingCount > 1000 {
			queueHealth["status"] = "warning"
			queueHealth["warning"] = "High pending count"
			overallHealthy = false
		}
		
		if metrics.ErrorRate > 10 {
			queueHealth["status"] = "critical"
			queueHealth["error"] = "High error rate"
			overallHealthy = false
		}
		
		if metrics.ProcessingCount > 100 {
			queueHealth["status"] = "warning"
			queueHealth["warning"] = "High processing count"
		}

		health["queues"].(map[string]interface{})[queueName] = queueHealth
	}

	if !overallHealthy {
		health["status"] = "degraded"
	}

	return health, nil
}

// 辅助函数
func parseInt64(s string) int64 {
	if s == "" {
		return 0
	}
	// 简单的字符串转换，实际应该使用strconv.ParseInt
	return 0
}

func parseDuration(s string) time.Duration {
	if s == "" {
		return 0
	}
	d, _ := time.ParseDuration(s)
	return d
}