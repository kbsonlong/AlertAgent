package plugins

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// WeChatWorkPlugin 企业微信通知插件
type WeChatWorkPlugin struct {
	httpClient *http.Client
}

// NewWeChatWorkPlugin 创建企业微信插件实例
func NewWeChatWorkPlugin() *WeChatWorkPlugin {
	return &WeChatWorkPlugin{
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// Name 插件名称
func (w *WeChatWorkPlugin) Name() string {
	return "wechat_work"
}

// Version 插件版本
func (w *WeChatWorkPlugin) Version() string {
	return "1.0.0"
}

// Description 插件描述
func (w *WeChatWorkPlugin) Description() string {
	return "企业微信群机器人通知插件，支持应用消息推送"
}

// ConfigSchema 配置Schema
func (w *WeChatWorkPlugin) ConfigSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"webhook_url": map[string]interface{}{
				"type":        "string",
				"description": "企业微信机器人Webhook URL",
				"pattern":     "^https://qyapi\\.weixin\\.qq\\.com/cgi-bin/webhook/send",
				"required":    true,
			},
			"message_type": map[string]interface{}{
				"type":        "string",
				"description": "消息类型",
				"enum":        []string{"text", "markdown"},
				"default":     "markdown",
				"required":    false,
			},
			"mentioned_list": map[string]interface{}{
				"type":        "array",
				"description": "@指定用户列表（userid）",
				"items": map[string]interface{}{
					"type": "string",
				},
				"required": false,
			},
			"mentioned_mobile_list": map[string]interface{}{
				"type":        "array",
				"description": "@指定用户手机号列表",
				"items": map[string]interface{}{
					"type": "string",
				},
				"required": false,
			},
		},
		"required": []string{"webhook_url"},
	}
}

// ValidateConfig 验证配置
func (w *WeChatWorkPlugin) ValidateConfig(config map[string]interface{}) error {
	webhookURL, ok := config["webhook_url"].(string)
	if !ok || webhookURL == "" {
		return fmt.Errorf("webhook_url is required")
	}

	if !strings.HasPrefix(webhookURL, "https://qyapi.weixin.qq.com/cgi-bin/webhook/send") {
		return fmt.Errorf("invalid wechat work webhook URL, must start with https://qyapi.weixin.qq.com/cgi-bin/webhook/send")
	}

	// 验证消息类型
	if msgType, exists := config["message_type"]; exists {
		if msgTypeStr, ok := msgType.(string); ok {
			if msgTypeStr != "text" && msgTypeStr != "markdown" {
				return fmt.Errorf("message_type must be 'text' or 'markdown'")
			}
		}
	}

	// 验证@手机号格式
	if mentionedMobiles, exists := config["mentioned_mobile_list"]; exists {
		if mobiles, ok := mentionedMobiles.([]interface{}); ok {
			for _, mobile := range mobiles {
				if mobileStr, ok := mobile.(string); ok {
					if len(mobileStr) != 11 || !strings.HasPrefix(mobileStr, "1") {
						return fmt.Errorf("invalid mobile number: %s", mobileStr)
					}
				}
			}
		}
	}

	return nil
}

// Send 发送通知
func (w *WeChatWorkPlugin) Send(ctx context.Context, config map[string]interface{}, message *NotificationMessage) error {
	webhookURL := config["webhook_url"].(string)
	messageType, _ := config["message_type"].(string)
	if messageType == "" {
		messageType = "markdown"
	}

	mentionedList, _ := config["mentioned_list"].([]interface{})
	mentionedMobileList, _ := config["mentioned_mobile_list"].([]interface{})

	// 构建企业微信消息
	var wechatMessage map[string]interface{}
	if messageType == "text" {
		wechatMessage = w.buildTextMessage(message, mentionedList, mentionedMobileList)
	} else {
		wechatMessage = w.buildMarkdownMessage(message, mentionedList, mentionedMobileList)
	}

	// 发送请求
	return w.sendRequest(ctx, webhookURL, wechatMessage)
}

// buildTextMessage 构建文本消息
func (w *WeChatWorkPlugin) buildTextMessage(message *NotificationMessage, mentionedList, mentionedMobileList []interface{}) map[string]interface{} {
	content := w.formatTextContent(message)

	wechatMessage := map[string]interface{}{
		"msgtype": "text",
		"text": map[string]interface{}{
			"content": content,
		},
	}

	// 添加@信息
	if len(mentionedList) > 0 || len(mentionedMobileList) > 0 {
		if len(mentionedList) > 0 {
			mentioned := make([]string, len(mentionedList))
			for i, user := range mentionedList {
				mentioned[i] = user.(string)
			}
			wechatMessage["text"].(map[string]interface{})["mentioned_list"] = mentioned
		}

		if len(mentionedMobileList) > 0 {
			mentionedMobiles := make([]string, len(mentionedMobileList))
			for i, mobile := range mentionedMobileList {
				mentionedMobiles[i] = mobile.(string)
			}
			wechatMessage["text"].(map[string]interface{})["mentioned_mobile_list"] = mentionedMobiles
		}
	}

	return wechatMessage
}

// buildMarkdownMessage 构建Markdown消息
func (w *WeChatWorkPlugin) buildMarkdownMessage(message *NotificationMessage, mentionedList, mentionedMobileList []interface{}) map[string]interface{} {
	content := w.formatMarkdownContent(message)

	wechatMessage := map[string]interface{}{
		"msgtype": "markdown",
		"markdown": map[string]interface{}{
			"content": content,
		},
	}

	return wechatMessage
}

// formatTextContent 格式化文本内容
func (w *WeChatWorkPlugin) formatTextContent(message *NotificationMessage) string {
	var builder strings.Builder

	builder.WriteString(fmt.Sprintf("【%s】%s\n", w.getSeverityText(message.Severity), message.Title))
	builder.WriteString(fmt.Sprintf("告警时间: %s\n", message.Timestamp.Format("2006-01-02 15:04:05")))
	builder.WriteString(fmt.Sprintf("告警内容: %s\n", message.Content))

	if message.AlertID != "" {
		builder.WriteString(fmt.Sprintf("告警ID: %s\n", message.AlertID))
	}

	if len(message.Labels) > 0 {
		builder.WriteString("标签信息:\n")
		for key, value := range message.Labels {
			builder.WriteString(fmt.Sprintf("  %s: %s\n", key, value))
		}
	}

	if len(message.Annotations) > 0 {
		builder.WriteString("注释信息:\n")
		for key, value := range message.Annotations {
			builder.WriteString(fmt.Sprintf("  %s: %s\n", key, value))
		}
	}

	return builder.String()
}

// formatMarkdownContent 格式化Markdown内容
func (w *WeChatWorkPlugin) formatMarkdownContent(message *NotificationMessage) string {
	var builder strings.Builder

	// 标题和基本信息
	builder.WriteString(fmt.Sprintf("## %s %s\n", w.getSeverityEmoji(message.Severity), message.Title))
	builder.WriteString(fmt.Sprintf("**告警级别**: <font color=\"%s\">%s</font>\n", w.getSeverityColor(message.Severity), w.getSeverityText(message.Severity)))
	builder.WriteString(fmt.Sprintf("**告警时间**: %s\n", message.Timestamp.Format("2006-01-02 15:04:05")))
	builder.WriteString(fmt.Sprintf("**告警内容**: %s\n", message.Content))

	if message.AlertID != "" {
		builder.WriteString(fmt.Sprintf("**告警ID**: `%s`\n", message.AlertID))
	}

	// 标签信息
	if len(message.Labels) > 0 {
		builder.WriteString("**标签信息**:\n")
		for key, value := range message.Labels {
			builder.WriteString(fmt.Sprintf("> %s: %s\n", key, value))
		}
	}

	// 注释信息
	if len(message.Annotations) > 0 {
		builder.WriteString("**注释信息**:\n")
		for key, value := range message.Annotations {
			builder.WriteString(fmt.Sprintf("> %s: %s\n", key, value))
		}
	}

	// 添加分割线和时间戳
	builder.WriteString("---\n")
	builder.WriteString(fmt.Sprintf("*发送时间: %s*", time.Now().Format("2006-01-02 15:04:05")))

	return builder.String()
}

// getSeverityText 获取告警级别文本
func (w *WeChatWorkPlugin) getSeverityText(severity string) string {
	switch strings.ToLower(severity) {
	case "critical":
		return "严重"
	case "high":
		return "高"
	case "medium":
		return "中"
	case "low":
		return "低"
	case "info":
		return "信息"
	case "warning":
		return "警告"
	default:
		return "未知"
	}
}

// getSeverityEmoji 获取告警级别表情
func (w *WeChatWorkPlugin) getSeverityEmoji(severity string) string {
	switch strings.ToLower(severity) {
	case "critical":
		return "🔴"
	case "high":
		return "🟠"
	case "medium":
		return "🟡"
	case "low":
		return "🟢"
	case "info":
		return "ℹ️"
	case "warning":
		return "⚠️"
	default:
		return "❓"
	}
}

// getSeverityColor 获取告警级别颜色
func (w *WeChatWorkPlugin) getSeverityColor(severity string) string {
	switch strings.ToLower(severity) {
	case "critical":
		return "warning" // 红色
	case "high":
		return "warning" // 橙色
	case "medium":
		return "comment" // 黄色
	case "low":
		return "info"    // 绿色
	case "info":
		return "info"    // 蓝色
	case "warning":
		return "comment" // 黄色
	default:
		return "comment"
	}
}

// sendRequest 发送HTTP请求
func (w *WeChatWorkPlugin) sendRequest(ctx context.Context, webhookURL string, message map[string]interface{}) error {
	jsonData, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", webhookURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := w.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("wechat work API returned status %d: %s", resp.StatusCode, string(body))
	}

	// 解析响应检查是否成功
	var response struct {
		ErrCode int    `json:"errcode"`
		ErrMsg  string `json:"errmsg"`
	}

	if err := json.Unmarshal(body, &response); err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}

	if response.ErrCode != 0 {
		return fmt.Errorf("wechat work API error: %s (code: %d)", response.ErrMsg, response.ErrCode)
	}

	return nil
}

// HealthCheck 健康检查
func (w *WeChatWorkPlugin) HealthCheck(ctx context.Context, config map[string]interface{}) error {
	// 创建测试消息
	testMessage := &NotificationMessage{
		Title:     "健康检查",
		Content:   "这是一条测试消息，用于验证企业微信通知插件配置是否正确。",
		Severity:  "info",
		AlertID:   fmt.Sprintf("health-check-%d", time.Now().Unix()),
		Timestamp: time.Now(),
		Labels: map[string]string{
			"test": "true",
			"type": "health_check",
		},
		Annotations: map[string]string{
			"description": "企业微信插件健康检查消息",
		},
	}

	return w.Send(ctx, config, testMessage)
}

// Initialize 初始化插件
func (w *WeChatWorkPlugin) Initialize() error {
	// 企业微信插件无需特殊初始化
	return nil
}

// Shutdown 关闭插件
func (w *WeChatWorkPlugin) Shutdown() error {
	// 企业微信插件无需特殊清理
	return nil
}