package queue

import (
	"context"
	"fmt"
	"time"

	"alert_agent/internal/pkg/logger"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// TaskProducer 任务生产者接口
type TaskProducer interface {
	// PublishTask 发布任务
	PublishTask(ctx context.Context, queueName string, task *Task) error
	// PublishDelayedTask 发布延迟任务
	PublishDelayedTask(ctx context.Context, queueName string, task *Task, delay time.Duration) error
	// PublishAIAnalysisTask 发布AI分析任务
	PublishAIAnalysisTask(ctx context.Context, alertID string, alertData map[string]interface{}) error
	// PublishNotificationTask 发布通知任务
	PublishNotificationTask(ctx context.Context, alertID string, channels []string, message map[string]interface{}) error
	// PublishConfigSyncTask 发布配置同步任务
	PublishConfigSyncTask(ctx context.Context, syncType, ruleID string, targets []string) error
	// GetTaskStatus 获取任务状态
	GetTaskStatus(ctx context.Context, taskID string) (*Task, error)
}

// DefaultTaskProducer 默认任务生产者实现
type DefaultTaskProducer struct {
	queue MessageQueue
}

// NewTaskProducer 创建任务生产者
func NewTaskProducer(queue MessageQueue) TaskProducer {
	return &DefaultTaskProducer{
		queue: queue,
	}
}

// PublishTask 发布任务
func (p *DefaultTaskProducer) PublishTask(ctx context.Context, queueName string, task *Task) error {
	if task.ID == "" {
		task.ID = uuid.New().String()
	}
	
	if task.CreatedAt.IsZero() {
		task.CreatedAt = time.Now()
	}
	
	if task.MaxRetry == 0 {
		task.MaxRetry = 3
	}

	err := p.queue.Publish(ctx, queueName, task)
	if err != nil {
		logger.L.Error("Failed to publish task",
			zap.String("queue", queueName),
			zap.String("task_id", task.ID),
			zap.String("type", string(task.Type)),
			zap.Error(err),
		)
		return fmt.Errorf("failed to publish task: %w", err)
	}

	logger.L.Info("Task published successfully",
		zap.String("queue", queueName),
		zap.String("task_id", task.ID),
		zap.String("type", string(task.Type)),
		zap.Int("priority", int(task.Priority)),
	)

	return nil
}

// PublishDelayedTask 发布延迟任务
func (p *DefaultTaskProducer) PublishDelayedTask(ctx context.Context, queueName string, task *Task, delay time.Duration) error {
	if task.ID == "" {
		task.ID = uuid.New().String()
	}
	
	if task.CreatedAt.IsZero() {
		task.CreatedAt = time.Now()
	}
	
	if task.MaxRetry == 0 {
		task.MaxRetry = 3
	}

	err := p.queue.PublishDelayed(ctx, queueName, task, delay)
	if err != nil {
		logger.L.Error("Failed to publish delayed task",
			zap.String("queue", queueName),
			zap.String("task_id", task.ID),
			zap.String("type", string(task.Type)),
			zap.Duration("delay", delay),
			zap.Error(err),
		)
		return fmt.Errorf("failed to publish delayed task: %w", err)
	}

	logger.L.Info("Delayed task published successfully",
		zap.String("queue", queueName),
		zap.String("task_id", task.ID),
		zap.String("type", string(task.Type)),
		zap.Duration("delay", delay),
	)

	return nil
}

// PublishAIAnalysisTask 发布AI分析任务
func (p *DefaultTaskProducer) PublishAIAnalysisTask(ctx context.Context, alertID string, alertData map[string]interface{}) error {
	payload := AIAnalysisPayload{
		AlertID:      alertID,
		AlertData:    alertData,
		AnalysisType: "root_cause",
		Options: map[string]interface{}{
			"include_suggestions": true,
			"include_similar":     true,
		},
	}

	task := &Task{
		Type:     TaskTypeAIAnalysis,
		Priority: PriorityHigh,
		MaxRetry: 3,
		Payload: map[string]interface{}{
			"alert_id":      payload.AlertID,
			"alert_data":    payload.AlertData,
			"analysis_type": payload.AnalysisType,
			"options":       payload.Options,
		},
	}

	return p.PublishTask(ctx, string(TaskTypeAIAnalysis), task)
}

// PublishNotificationTask 发布通知任务
func (p *DefaultTaskProducer) PublishNotificationTask(ctx context.Context, alertID string, channels []string, message map[string]interface{}) error {
	payload := NotificationPayload{
		AlertID:  alertID,
		Channels: channels,
		Message:  message,
		Template: "default_alert",
		Variables: map[string]interface{}{
			"alert_id": alertID,
			"timestamp": time.Now().Format("2006-01-02 15:04:05"),
		},
	}

	task := &Task{
		Type:     TaskTypeNotification,
		Priority: PriorityNormal,
		MaxRetry: 5,
		Payload: map[string]interface{}{
			"alert_id":  payload.AlertID,
			"channels":  payload.Channels,
			"message":   payload.Message,
			"template":  payload.Template,
			"variables": payload.Variables,
		},
	}

	return p.PublishTask(ctx, string(TaskTypeNotification), task)
}

// PublishConfigSyncTask 发布配置同步任务
func (p *DefaultTaskProducer) PublishConfigSyncTask(ctx context.Context, syncType, ruleID string, targets []string) error {
	payload := ConfigSyncPayload{
		Type:    syncType,
		RuleID:  ruleID,
		Targets: targets,
		ConfigData: map[string]interface{}{
			"sync_time": time.Now().Unix(),
		},
	}

	task := &Task{
		Type:     TaskTypeConfigSync,
		Priority: PriorityNormal,
		MaxRetry: 3,
		Payload: map[string]interface{}{
			"type":        payload.Type,
			"rule_id":     payload.RuleID,
			"targets":     payload.Targets,
			"config_data": payload.ConfigData,
		},
	}

	return p.PublishTask(ctx, string(TaskTypeConfigSync), task)
}

// GetTaskStatus 获取任务状态
func (p *DefaultTaskProducer) GetTaskStatus(ctx context.Context, taskID string) (*Task, error) {
	task, err := p.queue.GetTaskStatus(ctx, taskID)
	if err != nil {
		return nil, fmt.Errorf("failed to get task status: %w", err)
	}
	return task, nil
}

// BatchPublishTasks 批量发布任务
func (p *DefaultTaskProducer) BatchPublishTasks(ctx context.Context, queueName string, tasks []*Task) error {
	for _, task := range tasks {
		if err := p.PublishTask(ctx, queueName, task); err != nil {
			return fmt.Errorf("failed to publish task %s: %w", task.ID, err)
		}
	}
	return nil
}

// PublishHighPriorityTask 发布高优先级任务
func (p *DefaultTaskProducer) PublishHighPriorityTask(ctx context.Context, queueName string, taskType TaskType, payload map[string]interface{}) error {
	task := &Task{
		Type:     taskType,
		Priority: PriorityHigh,
		MaxRetry: 3,
		Payload:  payload,
	}

	return p.PublishTask(ctx, queueName, task)
}

// PublishCriticalTask 发布紧急任务
func (p *DefaultTaskProducer) PublishCriticalTask(ctx context.Context, queueName string, taskType TaskType, payload map[string]interface{}) error {
	task := &Task{
		Type:     taskType,
		Priority: PriorityCritical,
		MaxRetry: 5,
		Payload:  payload,
	}

	return p.PublishTask(ctx, queueName, task)
}

// ScheduleRecurringTask 调度周期性任务
func (p *DefaultTaskProducer) ScheduleRecurringTask(ctx context.Context, queueName string, task *Task, interval time.Duration) error {
	// 首次发布
	if err := p.PublishTask(ctx, queueName, task); err != nil {
		return err
	}

	// 调度下次执行
	nextTask := *task
	nextTask.ID = uuid.New().String()
	nextTask.CreatedAt = time.Time{}
	
	return p.PublishDelayedTask(ctx, queueName, &nextTask, interval)
}