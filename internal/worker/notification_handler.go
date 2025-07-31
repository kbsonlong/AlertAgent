package worker

import (
	"context"
	"fmt"

	"alert_agent/internal/pkg/logger"
	"alert_agent/internal/pkg/queue"

	"go.uber.org/zap"
)

// NotificationHandler 通知任务处理器
type NotificationHandler struct {
	// 这里可以添加通知服务依赖
}

// NewNotificationHandler 创建通知处理器
func NewNotificationHandler() *NotificationHandler {
	return &NotificationHandler{}
}

// Type 返回处理器类型
func (h *NotificationHandler) Type() queue.TaskType {
	return queue.TaskTypeNotification
}

// Handle 处理通知任务
func (h *NotificationHandler) Handle(ctx context.Context, task *queue.Task) error {
	// 解析任务载荷
	alertID, ok := task.Payload["alert_id"].(string)
	if !ok {
		return fmt.Errorf("invalid alert_id in task payload")
	}

	channels, ok := task.Payload["channels"].([]interface{})
	if !ok {
		return fmt.Errorf("invalid channels in task payload")
	}

	_, ok = task.Payload["message"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("invalid message in task payload")
	}

	logger.L.Info("Processing notification task",
		zap.String("task_id", task.ID),
		zap.String("alert_id", alertID),
		zap.Int("channel_count", len(channels)),
	)

	// TODO: 实现实际的通知发送逻辑
	// 这里应该调用通知服务发送通知到各个渠道
	
	logger.L.Info("Notification sent successfully",
		zap.String("task_id", task.ID),
		zap.String("alert_id", alertID),
	)

	return nil
}