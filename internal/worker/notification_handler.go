package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"alert_agent/internal/model"
	"alert_agent/internal/pkg/database"
	"alert_agent/internal/pkg/logger"
	"alert_agent/internal/pkg/queue"

	"go.uber.org/zap"
)

// NotificationHandler 通知任务处理器
type NotificationHandler struct {
	channels map[string]NotificationChannel
}

// NotificationChannel 通知渠道接口
type NotificationChannel interface {
	Name() string
	Send(ctx context.Context, config map[string]interface{}, message *NotificationMessage) error
	ValidateConfig(config map[string]interface{}) error
	HealthCheck(ctx context.Context) error
}

// NotificationMessage 通知消息
type NotificationMessage struct {
	AlertID     string                 `json:"alert_id"`
	Title       string                 `json:"title"`
	Content     string                 `json:"content"`
	Level       string                 `json:"level"`
	Source      string                 `json:"source"`
	Timestamp   time.Time              `json:"timestamp"`
	Labels      map[string]string      `json:"labels"`
	Annotations map[string]string      `json:"annotations"`
	Extra       map[string]interface{} `json:"extra"`
}

// NotificationResult 通知结果
type NotificationResult struct {
	Channel   string    `json:"channel"`
	Success   bool      `json:"success"`
	Error     string    `json:"error,omitempty"`
	Duration  time.Duration `json:"duration"`
	Timestamp time.Time `json:"timestamp"`
}

// NewNotificationHandler 创建通知处理器
func NewNotificationHandler() *NotificationHandler {
	handler := &NotificationHandler{
		channels: make(map[string]NotificationChannel),
	}

	// 注册默认通知渠道
	handler.registerDefaultChannels()

	return handler
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

	messageData, ok := task.Payload["message"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("invalid message in task payload")
	}

	logger.L.Info("Processing notification task",
		zap.String("task_id", task.ID),
		zap.String("alert_id", alertID),
		zap.Int("channel_count", len(channels)),
	)

	// 获取告警信息
	var alert model.Alert
	if err := database.DB.Where("id = ?", alertID).First(&alert).Error; err != nil {
		return fmt.Errorf("failed to get alert %s: %w", alertID, err)
	}

	// 构建通知消息
	message := h.buildNotificationMessage(&alert, messageData)

	// 发送通知到各个渠道
	results := make([]*NotificationResult, 0, len(channels))
	var lastError error

	for _, channelInterface := range channels {
		channelConfig, ok := channelInterface.(map[string]interface{})
		if !ok {
			logger.L.Warn("Invalid channel config", zap.Any("channel", channelInterface))
			continue
		}

		channelName, ok := channelConfig["name"].(string)
		if !ok {
			logger.L.Warn("Missing channel name", zap.Any("config", channelConfig))
			continue
		}

		result := h.sendToChannel(ctx, channelName, channelConfig, message)
		results = append(results, result)

		if !result.Success {
			lastError = fmt.Errorf("channel %s failed: %s", channelName, result.Error)
		}
	}

	// 记录通知结果
	if err := h.recordNotificationResults(ctx, &alert, results); err != nil {
		logger.L.Error("Failed to record notification results", zap.Error(err))
	}

	// 检查是否有成功的通知
	successCount := 0
	for _, result := range results {
		if result.Success {
			successCount++
		}
	}

	if successCount == 0 && len(results) > 0 {
		return fmt.Errorf("all notification channels failed, last error: %v", lastError)
	}

	logger.L.Info("Notification task completed",
		zap.String("task_id", task.ID),
		zap.String("alert_id", alertID),
		zap.Int("total_channels", len(results)),
		zap.Int("success_count", successCount),
	)

	return nil
}

// buildNotificationMessage 构建通知消息
func (h *NotificationHandler) buildNotificationMessage(alert *model.Alert, messageData map[string]interface{}) *NotificationMessage {
	message := &NotificationMessage{
		AlertID:     fmt.Sprintf("%d", alert.ID),
		Title:       alert.Title,
		Content:     alert.Content,
		Level:       alert.Level,
		Source:      alert.Source,
		Timestamp:   alert.CreatedAt,
		Labels:      make(map[string]string),
		Annotations: make(map[string]string),
		Extra:       make(map[string]interface{}),
	}

	// 解析标签和注释
	if alert.Labels != "" {
		json.Unmarshal([]byte(alert.Labels), &message.Labels)
	}

	// 从messageData中获取额外信息
	if title, ok := messageData["title"].(string); ok && title != "" {
		message.Title = title
	}

	if content, ok := messageData["content"].(string); ok && content != "" {
		message.Content = content
	}

	if template, ok := messageData["template"].(string); ok && template != "" {
		message.Content = h.renderTemplate(template, alert, messageData)
	}

	if extra, ok := messageData["extra"].(map[string]interface{}); ok {
		message.Extra = extra
	}

	return message
}

// renderTemplate 渲染消息模板
func (h *NotificationHandler) renderTemplate(template string, alert *model.Alert, variables map[string]interface{}) string {
	content := template

	// 替换告警相关变量
	replacements := map[string]string{
		"{{.AlertID}}":    fmt.Sprintf("%d", alert.ID),
		"{{.Title}}":      alert.Title,
		"{{.Content}}":    alert.Content,
		"{{.Level}}":      alert.Level,
		"{{.Source}}":     alert.Source,
		"{{.Status}}":     alert.Status,
		"{{.CreatedAt}}":  alert.CreatedAt.Format("2006-01-02 15:04:05"),
		"{{.UpdatedAt}}":  alert.UpdatedAt.Format("2006-01-02 15:04:05"),
	}

	// 添加分析结果
	if alert.Analysis != "" {
		replacements["{{.Analysis}}"] = alert.Analysis
	}

	// 替换自定义变量
	if vars, ok := variables["variables"].(map[string]interface{}); ok {
		for key, value := range vars {
			placeholder := fmt.Sprintf("{{.%s}}", key)
			replacements[placeholder] = fmt.Sprintf("%v", value)
		}
	}

	// 执行替换
	for placeholder, value := range replacements {
		content = strings.ReplaceAll(content, placeholder, value)
	}

	return content
}

// sendToChannel 发送通知到指定渠道
func (h *NotificationHandler) sendToChannel(ctx context.Context, channelName string, config map[string]interface{}, message *NotificationMessage) *NotificationResult {
	startTime := time.Now()
	result := &NotificationResult{
		Channel:   channelName,
		Timestamp: startTime,
	}

	// 查找通知渠道
	channel, exists := h.channels[channelName]
	if !exists {
		result.Success = false
		result.Error = fmt.Sprintf("notification channel %s not found", channelName)
		result.Duration = time.Since(startTime)
		return result
	}

	// 验证配置
	if err := channel.ValidateConfig(config); err != nil {
		result.Success = false
		result.Error = fmt.Sprintf("invalid config: %v", err)
		result.Duration = time.Since(startTime)
		return result
	}

	// 发送通知
	if err := channel.Send(ctx, config, message); err != nil {
		result.Success = false
		result.Error = err.Error()
		result.Duration = time.Since(startTime)
		
		logger.L.Error("Failed to send notification",
			zap.String("channel", channelName),
			zap.String("alert_id", message.AlertID),
			zap.Error(err),
		)
		return result
	}

	result.Success = true
	result.Duration = time.Since(startTime)

	logger.L.Info("Notification sent successfully",
		zap.String("channel", channelName),
		zap.String("alert_id", message.AlertID),
		zap.Duration("duration", result.Duration),
	)

	return result
}

// recordNotificationResults 记录通知结果
func (h *NotificationHandler) recordNotificationResults(ctx context.Context, alert *model.Alert, results []*NotificationResult) error {
	// 这里可以将通知结果保存到数据库或日志中
	// 暂时只记录到日志

	for _, result := range results {
		logger.L.Info("Notification result",
			zap.String("alert_id", fmt.Sprintf("%d", alert.ID)),
			zap.String("channel", result.Channel),
			zap.Bool("success", result.Success),
			zap.String("error", result.Error),
			zap.Duration("duration", result.Duration),
		)
	}

	// 可以考虑将结果保存到单独的通知历史表中
	// 这里暂时跳过数据库操作

	return nil
}

// registerDefaultChannels 注册默认通知渠道
func (h *NotificationHandler) registerDefaultChannels() {
	// 注册邮件通知渠道
	h.channels["email"] = NewEmailChannel()
	
	// 注册Webhook通知渠道
	h.channels["webhook"] = NewWebhookChannel()
	
	// 注册钉钉通知渠道
	h.channels["dingtalk"] = NewDingTalkChannel()
	
	// 注册企业微信通知渠道
	h.channels["wechat"] = NewWeChatChannel()

	logger.L.Info("Default notification channels registered",
		zap.Int("channel_count", len(h.channels)),
		zap.Strings("channels", h.getChannelNames()),
	)
}

// RegisterChannel 注册通知渠道
func (h *NotificationHandler) RegisterChannel(channel NotificationChannel) {
	h.channels[channel.Name()] = channel
	logger.L.Info("Notification channel registered", zap.String("channel", channel.Name()))
}

// GetChannels 获取所有通知渠道
func (h *NotificationHandler) GetChannels() map[string]NotificationChannel {
	return h.channels
}

// getChannelNames 获取渠道名称列表
func (h *NotificationHandler) getChannelNames() []string {
	names := make([]string, 0, len(h.channels))
	for name := range h.channels {
		names = append(names, name)
	}
	return names
}

// HealthCheck 健康检查
func (h *NotificationHandler) HealthCheck(ctx context.Context) map[string]error {
	results := make(map[string]error)
	
	for name, channel := range h.channels {
		if err := channel.HealthCheck(ctx); err != nil {
			results[name] = err
		}
	}
	
	return results
}