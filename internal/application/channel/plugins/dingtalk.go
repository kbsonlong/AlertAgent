package plugins

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// DingTalkPlugin 钉钉通知插件
type DingTalkPlugin struct {
	httpClient *http.Client
}

// NewDingTalkPlugin 创建钉钉插件实例
func NewDingTalkPlugin() *DingTalkPlugin {
	return &DingTalkPlugin{
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// Name 插件名称
func (d *DingTalkPlugin) Name() string {
	return "dingtalk"
}

// Version 插件版本
func (d *DingTalkPlugin) Version() string {
	return "1.0.0"
}

// Description 插件描述
func (d *DingTalkPlugin) Description() string {
	return "钉钉群机器人通知插件，支持群机器人和个人消息推送"
}

// ConfigSchema 配置Schema
func (d *DingTalkPlugin) ConfigSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"webhook_url": map[string]interface{}{
				"type":        "string",
				"description": "钉钉机器人Webhook URL",
				"pattern":     "^https://oapi\\.dingtalk\\.com/robot/send",
				"required":    true,
			},
			"secret": map[string]interface{}{
				"type":        "string",
				"description": "钉钉机器人密钥（可选，用于签名验证）",
				"required":    false,
			},
			"at_mobiles": map[string]interface{}{
				"type":        "array",
				"description": "@指定手机号列表",
				"items": map[string]interface{}{
					"type": "string",
				},
				"required": false,
			},
			"at_all": map[string]interface{}{
				"type":        "boolean",
				"description": "是否@所有人",
				"default":     false,
				"required":    false,
			},
			"message_type": map[string]interface{}{
				"type":        "string",
				"description": "消息类型",
				"enum":        []string{"text", "markdown"},
				"default":     "markdown",
				"required":    false,
			},
		},
		"required": []string{"webhook_url"},
	}
}

// ValidateConfig 验证配置
func (d *DingTalkPlugin) ValidateConfig(config map[string]interface{}) error {
	webhookURL, ok := config["webhook_url"].(string)
	if !ok || webhookURL == "" {
		return fmt.Errorf("webhook_url is required")
	}

	if !strings.HasPrefix(webhookURL, "https://oapi.dingtalk.com/robot/send") {
		return fmt.Errorf("invalid dingtalk webhook URL, must start with https://oapi.dingtalk.com/robot/send")
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
	if atMobiles, exists := config["at_mobiles"]; exists {
		if mobiles, ok := atMobiles.([]interface{}); ok {
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
func (d *DingTalkPlugin) Send(ctx context.Context, config map[string]interface{}, message *NotificationMessage) error {
	webhookURL := config["webhook_url"].(string)
	secret, _ := config["secret"].(string)
	atMobiles, _ := config["at_mobiles"].([]interface{})
	atAll, _ := config["at_all"].(bool)
	messageType, _ := config["message_type"].(string)
	if messageType == "" {
		messageType = "markdown"
	}

	// 构建钉钉消息
	var dingMessage map[string]interface{}
	if messageType == "text" {
		dingMessage = d.buildTextMessage(message, atMobiles, atAll)
	} else {
		dingMessage = d.buildMarkdownMessage(message, atMobiles, atAll)
	}

	// 计算签名
	if secret != "" {
		timestamp := time.Now().UnixMilli()
		sign := d.calculateSign(timestamp, secret)
		webhookURL = fmt.Sprintf("%s&timestamp=%d&sign=%s", webhookURL, timestamp, sign)
	}

	// 发送请求
	return d.sendRequest(ctx, webhookURL, dingMessage)
}

// buildTextMessage 构建文本消息
func (d *DingTalkPlugin) buildTextMessage(message *NotificationMessage, atMobiles []interface{}, atAll bool) map[string]interface{} {
	content := d.formatTextContent(message)

	dingMessage := map[string]interface{}{
		"msgtype": "text",
		"text": map[string]interface{}{
			"content": content,
		},
	}

	// 添加@信息
	if len(atMobiles) > 0 || atAll {
		at := map[string]interface{}{
			"isAtAll": atAll,
		}

		if len(atMobiles) > 0 {
			mobiles := make([]string, len(atMobiles))
			for i, mobile := range atMobiles {
				mobiles[i] = mobile.(string)
			}
			at["atMobiles"] = mobiles
		}

		dingMessage["at"] = at
	}

	return dingMessage
}

// buildMarkdownMessage 构建Markdown消息
func (d *DingTalkPlugin) buildMarkdownMessage(message *NotificationMessage, atMobiles []interface{}, atAll bool) map[string]interface{} {
	content := d.formatMarkdownContent(message)

	dingMessage := map[string]interface{}{
		"msgtype": "markdown",
		"markdown": map[string]interface{}{
			"title": message.Title,
			"text":  content,
		},
	}

	// 添加@信息
	if len(atMobiles) > 0 || atAll {
		at := map[string]interface{}{
			"isAtAll": atAll,
		}

		if len(atMobiles) > 0 {
			mobiles := make([]string, len(atMobiles))
			for i, mobile := range atMobiles {
				mobiles[i] = mobile.(string)
			}
			at["atMobiles"] = mobiles
		}

		dingMessage["at"] = at
	}

	return dingMessage
}

// formatTextContent 格式化文本内容
func (d *DingTalkPlugin) formatTextContent(message *NotificationMessage) string {
	var builder strings.Builder

	builder.WriteString(fmt.Sprintf("【%s】%s\n", d.getSeverityText(message.Severity), message.Title))
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
func (d *DingTalkPlugin) formatMarkdownContent(message *NotificationMessage) string {
	var builder strings.Builder

	// 标题和基本信息
	builder.WriteString(fmt.Sprintf("## %s %s\n\n", d.getSeverityEmoji(message.Severity), message.Title))
	builder.WriteString(fmt.Sprintf("**告警级别**: %s\n\n", d.getSeverityText(message.Severity)))
	builder.WriteString(fmt.Sprintf("**告警时间**: %s\n\n", message.Timestamp.Format("2006-01-02 15:04:05")))
	builder.WriteString(fmt.Sprintf("**告警内容**: %s\n\n", message.Content))

	if message.AlertID != "" {
		builder.WriteString(fmt.Sprintf("**告警ID**: `%s`\n\n", message.AlertID))
	}

	// 标签信息
	if len(message.Labels) > 0 {
		builder.WriteString("**标签信息**:\n\n")
		for key, value := range message.Labels {
			builder.WriteString(fmt.Sprintf("- **%s**: %s\n", key, value))
		}
		builder.WriteString("\n")
	}

	// 注释信息
	if len(message.Annotations) > 0 {
		builder.WriteString("**注释信息**:\n\n")
		for key, value := range message.Annotations {
			builder.WriteString(fmt.Sprintf("- **%s**: %s\n", key, value))
		}
		builder.WriteString("\n")
	}

	// 添加分割线
	builder.WriteString("---\n")
	builder.WriteString(fmt.Sprintf("*发送时间: %s*", time.Now().Format("2006-01-02 15:04:05")))

	return builder.String()
}

// getSeverityText 获取告警级别文本
func (d *DingTalkPlugin) getSeverityText(severity string) string {
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
func (d *DingTalkPlugin) getSeverityEmoji(severity string) string {
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

// calculateSign 计算签名
func (d *DingTalkPlugin) calculateSign(timestamp int64, secret string) string {
	stringToSign := fmt.Sprintf("%d\n%s", timestamp, secret)
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(stringToSign))
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

// sendRequest 发送HTTP请求
func (d *DingTalkPlugin) sendRequest(ctx context.Context, webhookURL string, message map[string]interface{}) error {
	jsonData, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", webhookURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := d.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("dingtalk API returned status %d: %s", resp.StatusCode, string(body))
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
		return fmt.Errorf("dingtalk API error: %s (code: %d)", response.ErrMsg, response.ErrCode)
	}

	return nil
}

// HealthCheck 健康检查
func (d *DingTalkPlugin) HealthCheck(ctx context.Context, config map[string]interface{}) error {
	// 创建测试消息
	testMessage := &NotificationMessage{
		Title:     "健康检查",
		Content:   "这是一条测试消息，用于验证钉钉通知插件配置是否正确。",
		Severity:  "info",
		AlertID:   fmt.Sprintf("health-check-%d", time.Now().Unix()),
		Timestamp: time.Now(),
		Labels: map[string]string{
			"test": "true",
			"type": "health_check",
		},
		Annotations: map[string]string{
			"description": "钉钉插件健康检查消息",
		},
	}

	return d.Send(ctx, config, testMessage)
}

// Initialize 初始化插件
func (d *DingTalkPlugin) Initialize() error {
	// 钉钉插件无需特殊初始化
	return nil
}

// Shutdown 关闭插件
func (d *DingTalkPlugin) Shutdown() error {
	// 钉钉插件无需特殊清理
	return nil
}