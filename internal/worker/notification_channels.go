package worker

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

	"alert_agent/internal/application/channel/plugins"
	"alert_agent/internal/pkg/logger"

	"go.uber.org/zap"
)

// EmailChannel 邮件通知渠道
type EmailChannel struct {
	client *http.Client
}

// NewEmailChannel 创建邮件通知渠道
func NewEmailChannel() *EmailChannel {
	return &EmailChannel{
		client: &http.Client{Timeout: 30 * time.Second},
	}
}

func (e *EmailChannel) Name() string {
	return "email"
}

func (e *EmailChannel) ValidateConfig(config map[string]interface{}) error {
	required := []string{"smtp_host", "smtp_port", "username", "password", "to"}
	for _, field := range required {
		if _, ok := config[field]; !ok {
			return fmt.Errorf("missing required field: %s", field)
		}
	}
	return nil
}

func (e *EmailChannel) Send(ctx context.Context, config map[string]interface{}, message *plugins.NotificationMessage) error {
	// 这里实现邮件发送逻辑
	// 暂时返回成功，实际实现需要配置SMTP服务器
	logger.L.Info("Email notification sent (mock)",
		zap.String("alert_id", message.AlertID),
		zap.String("title", message.Title),
	)
	return nil
}

func (e *EmailChannel) HealthCheck(ctx context.Context) error {
	return nil
}

// WebhookChannel Webhook通知渠道
type WebhookChannel struct {
	client *http.Client
}

// NewWebhookChannel 创建Webhook通知渠道
func NewWebhookChannel() *WebhookChannel {
	return &WebhookChannel{
		client: &http.Client{Timeout: 30 * time.Second},
	}
}

func (w *WebhookChannel) Name() string {
	return "webhook"
}

func (w *WebhookChannel) ValidateConfig(config map[string]interface{}) error {
	if _, ok := config["url"]; !ok {
		return fmt.Errorf("missing required field: url")
	}
	return nil
}

func (w *WebhookChannel) Send(ctx context.Context, config map[string]interface{}, message *plugins.NotificationMessage) error {
	// 从配置中获取URL和其他参数
	url, _ := config["url"].(string)
	method, _ := config["method"].(string)
	if method == "" {
		method = "POST"
	}

	// 构建请求体
	payload := map[string]interface{}{
		"alert_id":  message.AlertID,
		"title":     message.Title,
		"content":   message.Content,
		"level":     message.Severity,
		"timestamp": message.Timestamp.Unix(),
		"labels":    message.Labels,
		"extra":     message.Extra,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "AlertAgent-Worker/1.0")

	// 添加自定义头部
	if headers, ok := config["headers"].(map[string]interface{}); ok {
		for key, value := range headers {
			req.Header.Set(key, fmt.Sprintf("%v", value))
		}
	}

	resp, err := w.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send webhook: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("webhook returned status %d: %s", resp.StatusCode, string(body))
	}

	logger.L.Info("Webhook notification sent",
		zap.String("alert_id", message.AlertID),
		zap.String("url", url),
		zap.Int("status_code", resp.StatusCode),
	)

	return nil
}

func (w *WebhookChannel) HealthCheck(ctx context.Context) error {
	return nil
}

// DingTalkChannel 钉钉通知渠道
type DingTalkChannel struct {
	client *http.Client
}

// NewDingTalkChannel 创建钉钉通知渠道
func NewDingTalkChannel() *DingTalkChannel {
	return &DingTalkChannel{
		client: &http.Client{Timeout: 30 * time.Second},
	}
}

func (d *DingTalkChannel) Name() string {
	return "dingtalk"
}

func (d *DingTalkChannel) ValidateConfig(config map[string]interface{}) error {
	if _, ok := config["webhook_url"]; !ok {
		return fmt.Errorf("missing required field: webhook_url")
	}
	
	webhookURL, _ := config["webhook_url"].(string)
	if !strings.HasPrefix(webhookURL, "https://oapi.dingtalk.com/robot/send") {
		return fmt.Errorf("invalid dingtalk webhook URL")
	}
	
	return nil
}

func (d *DingTalkChannel) Send(ctx context.Context, config map[string]interface{}, message *plugins.NotificationMessage) error {
	webhookURL, _ := config["webhook_url"].(string)
	secret, _ := config["secret"].(string)
	atMobiles, _ := config["at_mobiles"].([]interface{})
	atAll, _ := config["at_all"].(bool)

	// 构建钉钉消息格式
	dingMessage := map[string]interface{}{
		"msgtype": "markdown",
		"markdown": map[string]interface{}{
			"title": message.Title,
			"text":  d.formatMessage(message),
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

	// 计算签名
	if secret != "" {
		timestamp := time.Now().UnixMilli()
		sign := d.calculateSign(timestamp, secret)
		webhookURL = fmt.Sprintf("%s&timestamp=%d&sign=%s", webhookURL, timestamp, sign)
	}

	// 发送请求
	jsonData, err := json.Marshal(dingMessage)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", webhookURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := d.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("dingtalk API returned status %d: %s", resp.StatusCode, string(body))
	}

	// 检查响应结果
	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err == nil {
		if errcode, ok := result["errcode"].(float64); ok && errcode != 0 {
			errmsg, _ := result["errmsg"].(string)
			return fmt.Errorf("dingtalk API error: %s (code: %.0f)", errmsg, errcode)
		}
	}

	logger.L.Info("DingTalk notification sent",
		zap.String("alert_id", message.AlertID),
		zap.String("title", message.Title),
	)

	return nil
}

func (d *DingTalkChannel) formatMessage(message *plugins.NotificationMessage) string {
	var builder strings.Builder

	builder.WriteString(fmt.Sprintf("## %s\n\n", message.Title))
	builder.WriteString(fmt.Sprintf("**告警级别**: %s\n\n", message.Severity))
	builder.WriteString(fmt.Sprintf("**告警时间**: %s\n\n", message.Timestamp.Format("2006-01-02 15:04:05")))
	builder.WriteString(fmt.Sprintf("**告警内容**: %s\n\n", message.Content))

	if len(message.Labels) > 0 {
		builder.WriteString("**标签信息**:\n\n")
		for key, value := range message.Labels {
			builder.WriteString(fmt.Sprintf("- %s: %s\n", key, value))
		}
		builder.WriteString("\n")
	}

	return builder.String()
}

func (d *DingTalkChannel) calculateSign(timestamp int64, secret string) string {
	stringToSign := fmt.Sprintf("%d\n%s", timestamp, secret)
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(stringToSign))
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

func (d *DingTalkChannel) HealthCheck(ctx context.Context) error {
	return nil
}

// WeChatChannel 企业微信通知渠道
type WeChatChannel struct {
	client *http.Client
}

// NewWeChatChannel 创建企业微信通知渠道
func NewWeChatChannel() *WeChatChannel {
	return &WeChatChannel{
		client: &http.Client{Timeout: 30 * time.Second},
	}
}

func (w *WeChatChannel) Name() string {
	return "wechat"
}

func (w *WeChatChannel) ValidateConfig(config map[string]interface{}) error {
	required := []string{"webhook_url"}
	for _, field := range required {
		if _, ok := config[field]; !ok {
			return fmt.Errorf("missing required field: %s", field)
		}
	}
	return nil
}

func (w *WeChatChannel) Send(ctx context.Context, config map[string]interface{}, message *plugins.NotificationMessage) error {
	webhookURL, _ := config["webhook_url"].(string)

	// 构建企业微信消息格式
	wechatMessage := map[string]interface{}{
		"msgtype": "markdown",
		"markdown": map[string]interface{}{
			"content": w.formatMessage(message),
		},
	}

	// 发送请求
	jsonData, err := json.Marshal(wechatMessage)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", webhookURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := w.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("wechat API returned status %d: %s", resp.StatusCode, string(body))
	}

	logger.L.Info("WeChat notification sent",
		zap.String("alert_id", message.AlertID),
		zap.String("title", message.Title),
	)

	return nil
}

func (w *WeChatChannel) formatMessage(message *plugins.NotificationMessage) string {
	var builder strings.Builder

	builder.WriteString(fmt.Sprintf("## %s\n", message.Title))
	builder.WriteString(fmt.Sprintf("**告警级别**: <font color=\"warning\">%s</font>\n", message.Severity))
	builder.WriteString(fmt.Sprintf("**告警时间**: %s\n", message.Timestamp.Format("2006-01-02 15:04:05")))
	builder.WriteString(fmt.Sprintf("**告警内容**: %s\n", message.Content))

	if len(message.Labels) > 0 {
		builder.WriteString("**标签信息**:\n")
		for key, value := range message.Labels {
			builder.WriteString(fmt.Sprintf("> %s: %s\n", key, value))
		}
	}

	return builder.String()
}

func (w *WeChatChannel) HealthCheck(ctx context.Context) error {
	return nil
}

