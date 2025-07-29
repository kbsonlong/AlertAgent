package plugins

import (
	"alert_agent/internal/domain/channel"
	"alert_agent/pkg/types"
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// WebhookPlugin Webhook插件实现
type WebhookPlugin struct {
	client *http.Client
}

// WebhookPayload Webhook负载结构
type WebhookPayload struct {
	ID        string                 `json:"id"`
	Title     string                 `json:"title"`
	Content   string                 `json:"content"`
	Priority  string                 `json:"priority"`
	Type      string                 `json:"type"`
	Timestamp int64                  `json:"timestamp"`
	Data      map[string]interface{} `json:"data,omitempty"`
}

// NewWebhookPlugin 创建新的Webhook插件实例
func NewWebhookPlugin() *WebhookPlugin {
	return &WebhookPlugin{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// GetType 获取插件类型
func (p *WebhookPlugin) GetType() channel.ChannelType {
	return channel.ChannelTypeWebhook
}

// GetName 获取插件名称
func (p *WebhookPlugin) GetName() string {
	return "Webhook"
}

// GetVersion 获取插件版本
func (p *WebhookPlugin) GetVersion() string {
	return "1.0.0"
}

// GetDescription 获取插件描述
func (p *WebhookPlugin) GetDescription() string {
	return "Webhook通知插件，支持HTTP POST请求发送告警消息"
}

// GetConfigSchema 获取配置模式
func (p *WebhookPlugin) GetConfigSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"url": map[string]interface{}{
				"type":        "string",
				"description": "Webhook URL地址",
				"format":      "uri",
			},
			"method": map[string]interface{}{
				"type":        "string",
				"description": "HTTP请求方法",
				"enum":        []string{"POST", "PUT", "PATCH"},
				"default":     "POST",
			},
			"headers": map[string]interface{}{
				"type":        "object",
				"description": "自定义HTTP头",
				"additionalProperties": map[string]interface{}{
					"type": "string",
				},
			},
			"timeout": map[string]interface{}{
				"type":        "integer",
				"description": "请求超时时间（秒）",
				"minimum":     1,
				"maximum":     300,
				"default":     30,
			},
			"secret": map[string]interface{}{
				"type":        "string",
				"description": "用于HMAC签名的密钥",
			},
			"content_type": map[string]interface{}{
				"type":        "string",
				"description": "Content-Type头",
				"enum":        []string{"application/json", "application/x-www-form-urlencoded"},
				"default":     "application/json",
			},
			"retry_count": map[string]interface{}{
				"type":        "integer",
				"description": "重试次数",
				"minimum":     0,
				"maximum":     5,
				"default":     3,
			},
			"retry_interval": map[string]interface{}{
				"type":        "integer",
				"description": "重试间隔（秒）",
				"minimum":     1,
				"maximum":     60,
				"default":     5,
			},
			"custom_payload": map[string]interface{}{
				"type":        "object",
				"description": "自定义负载模板",
			},
		},
		"required": []string{"url"},
	}
}

// Initialize 初始化插件
func (p *WebhookPlugin) Initialize(ctx context.Context, config map[string]interface{}) error {
	// 这里可以进行插件级别的初始化
	return nil
}

// Start 启动插件
func (p *WebhookPlugin) Start(ctx context.Context) error {
	return nil
}

// Stop 停止插件
func (p *WebhookPlugin) Stop(ctx context.Context) error {
	return nil
}

// HealthCheck 健康检查
func (p *WebhookPlugin) HealthCheck(ctx context.Context) error {
	return nil
}

// HealthCheckWithConfig 带配置的健康检查
func (p *WebhookPlugin) HealthCheckWithConfig(ctx context.Context, channelID string, config *channel.ChannelConfig) *channel.HealthStatus {
	status := &channel.HealthStatus{
		ChannelID:    channelID,
		Status:       channel.HealthStatusUnhealthy,
		Message:      "",
		LastCheck:    time.Now(),
		ResponseTime: 0,
	}
	
	url, ok := config.Settings["url"].(string)
	if !ok || url == "" {
		status.Message = "Webhook URL未配置"
		return status
	}
	
	// 发送测试请求
	testPayload := map[string]interface{}{
		"test": true,
		"timestamp": time.Now().Unix(),
	}
	
	payloadBytes, _ := json.Marshal(testPayload)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(payloadBytes))
	if err != nil {
		status.Message = fmt.Sprintf("创建请求失败: %v", err)
		return status
	}
	
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "AlertAgent-Webhook/1.0")
	
	resp, err := p.client.Do(req)
	if err != nil {
		status.Message = fmt.Sprintf("请求失败: %v", err)
		return status
	}
	defer resp.Body.Close()
	
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		status.Status = channel.HealthStatusHealthy
		status.Message = "连接正常"
	} else {
		status.Message = fmt.Sprintf("HTTP状态码: %d", resp.StatusCode)
	}
	
	return status
}

// SendMessage 发送消息
func (p *WebhookPlugin) SendMessage(ctx context.Context, config channel.ChannelConfig, message *types.Message) (*channel.SendResult, error) {
	start := time.Now()
	result := &channel.SendResult{
		ChannelID: "", // 由调用方设置
		Success:   false,
		Timestamp: start,
	}
	
	url, ok := config.Settings["url"].(string)
	if !ok || url == "" {
		result.Error = "Webhook URL未配置"
		result.Latency = time.Since(start)
		return result, fmt.Errorf("Webhook URL未配置")
	}
	
	// 构建负载
	payload, err := p.buildPayload(&config, message)
	if err != nil {
		result.Error = fmt.Sprintf("构建负载失败: %v", err)
		result.Latency = time.Since(start)
		return result, err
	}
	
	// 获取重试配置
	retryCount := 3
	if retryCountInterface, exists := config.Settings["retry_count"]; exists {
		switch v := retryCountInterface.(type) {
		case int:
			retryCount = v
		case float64:
			retryCount = int(v)
		}
	}
	
	retryInterval := 5 * time.Second
	if retryIntervalInterface, exists := config.Settings["retry_interval"]; exists {
		switch v := retryIntervalInterface.(type) {
		case int:
			retryInterval = time.Duration(v) * time.Second
		case float64:
			retryInterval = time.Duration(int(v)) * time.Second
		}
	}
	
	// 执行发送（带重试）
	var lastErr error
	for i := 0; i <= retryCount; i++ {
		if i > 0 {
			select {
			case <-ctx.Done():
				result.Error = "请求被取消"
				result.Latency = time.Since(start)
				return result, fmt.Errorf("请求被取消")
			case <-time.After(retryInterval):
			}
		}
		
		err := p.sendWebhook(ctx, &config, payload)
		if err == nil {
			result.Success = true
			result.RetryCount = i
			result.Latency = time.Since(start)
			return result, nil
		}
		
		lastErr = err
	}
	
	result.Error = fmt.Sprintf("发送失败（重试%d次）: %v", retryCount, lastErr)
	result.RetryCount = retryCount
	result.Latency = time.Since(start)
	return result, lastErr
}

// ValidateConfig 验证配置
func (p *WebhookPlugin) ValidateConfig(config channel.ChannelConfig) error {

	
	url, ok := config.Settings["url"].(string)
	if !ok || url == "" {
		return fmt.Errorf("Webhook URL不能为空")
	}
	
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		return fmt.Errorf("Webhook URL必须以http://或https://开头")
	}
	
	// 验证HTTP方法
	if methodInterface, exists := config.Settings["method"]; exists {
		method, ok := methodInterface.(string)
		if !ok {
			return fmt.Errorf("HTTP方法必须是字符串")
		}
		method = strings.ToUpper(method)
		if method != "POST" && method != "PUT" && method != "PATCH" {
			return fmt.Errorf("HTTP方法必须是POST、PUT或PATCH")
		}
	}
	
	// 验证超时时间
	if timeoutInterface, exists := config.Settings["timeout"]; exists {
		var timeout int
		switch v := timeoutInterface.(type) {
		case int:
			timeout = v
		case float64:
			timeout = int(v)
		default:
			return fmt.Errorf("超时时间必须是数字")
		}
		if timeout < 1 || timeout > 300 {
			return fmt.Errorf("超时时间必须在1-300秒之间")
		}
	}
	
	// 验证重试次数
	if retryCountInterface, exists := config.Settings["retry_count"]; exists {
		var retryCount int
		switch v := retryCountInterface.(type) {
		case int:
			retryCount = v
		case float64:
			retryCount = int(v)
		default:
			return fmt.Errorf("重试次数必须是数字")
		}
		if retryCount < 0 || retryCount > 5 {
			return fmt.Errorf("重试次数必须在0-5之间")
		}
	}
	
	return nil
}

// TestConnection 测试连接
func (p *WebhookPlugin) TestConnection(ctx context.Context, config channel.ChannelConfig) (*channel.TestResult, error) {
	start := time.Now()
	result := &channel.TestResult{
		Success:   false,
		Timestamp: start.Unix(),
	}
	
	url, ok := config.Settings["url"].(string)
	if !ok || url == "" {
		result.Message = "Webhook URL未配置"
		result.Latency = time.Since(start).Milliseconds()
		return result, fmt.Errorf("Webhook URL未配置")
	}
	
	// 发送测试请求
	testPayload := map[string]interface{}{
		"test":      true,
		"message":   "AlertAgent连接测试",
		"timestamp": time.Now().Unix(),
	}
	
	payloadBytes, _ := json.Marshal(testPayload)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(payloadBytes))
	if err != nil {
		result.Message = fmt.Sprintf("创建请求失败: %v", err)
		result.Latency = time.Since(start).Milliseconds()
		return result, err
	}
	
	// 设置请求头
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "AlertAgent-Webhook/1.0")
	
	// 添加自定义头
	if headersInterface, exists := config.Settings["headers"]; exists {
		if headers, ok := headersInterface.(map[string]interface{}); ok {
			for key, value := range headers {
				if strValue, ok := value.(string); ok {
					req.Header.Set(key, strValue)
				}
			}
		}
	}
	
	// 添加签名
	if secret, exists := config.Settings["secret"].(string); exists && secret != "" {
		signature := p.generateSignature(payloadBytes, secret)
		req.Header.Set("X-Signature", signature)
	}
	
	resp, err := p.client.Do(req)
	if err != nil {
		result.Message = fmt.Sprintf("请求失败: %v", err)
		result.Latency = time.Since(start).Milliseconds()
		return result, err
	}
	defer resp.Body.Close()
	
	body, _ := io.ReadAll(resp.Body)
	
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		result.Success = true
		result.Message = "连接测试成功"
		result.Details = map[string]interface{}{
			"status_code": resp.StatusCode,
			"response":    string(body),
		}
	} else {
		result.Message = fmt.Sprintf("HTTP状态码: %d", resp.StatusCode)
		result.Details = map[string]interface{}{
			"status_code": resp.StatusCode,
			"response":    string(body),
		}
	}
	
	result.Latency = time.Since(start).Milliseconds()
	return result, nil
}

// GetCapabilities 获取插件支持的功能
func (p *WebhookPlugin) GetCapabilities() []channel.PluginCapability {
	return []channel.PluginCapability{
		channel.CapabilityTextMessage,
		channel.CapabilityTemplating,
		channel.CapabilityHealthCheck,
	}
}

// SupportsFeature 检查是否支持特定功能
func (p *WebhookPlugin) SupportsFeature(feature channel.PluginCapability) bool {
	capabilities := p.GetCapabilities()
	for _, cap := range capabilities {
		if cap == feature {
			return true
		}
	}
	return false
}

// buildPayload 构建负载
func (p *WebhookPlugin) buildPayload(config *channel.ChannelConfig, message *types.Message) ([]byte, error) {
	// 检查是否有自定义负载模板
	if customPayloadInterface, exists := config.Settings["custom_payload"]; exists {
		if customPayload, ok := customPayloadInterface.(map[string]interface{}); ok {
			// 使用自定义负载模板
			payload := make(map[string]interface{})
			for key, value := range customPayload {
				// 简单的模板变量替换
				if strValue, ok := value.(string); ok {
					strValue = strings.ReplaceAll(strValue, "{{.Title}}", message.Title)
					strValue = strings.ReplaceAll(strValue, "{{.Content}}", message.Content)
					strValue = strings.ReplaceAll(strValue, "{{.Priority}}", string(message.Priority))
					strValue = strings.ReplaceAll(strValue, "{{.Type}}", message.Type)
					strValue = strings.ReplaceAll(strValue, "{{.Timestamp}}", strconv.FormatInt(message.CreatedAt.Unix(), 10))
					payload[key] = strValue
				} else {
					payload[key] = value
				}
			}
			return json.Marshal(payload)
		}
	}
	
	// 使用默认负载格式
	payload := &WebhookPayload{
		ID:        message.ID,
		Title:     message.Title,
		Content:   message.Content,
		Priority:  string(message.Priority),
		Type:      message.Type,
		Timestamp: message.CreatedAt.Unix(),
		Data:      message.Data,
	}
	
	return json.Marshal(payload)
}

// sendWebhook 发送Webhook请求
func (p *WebhookPlugin) sendWebhook(ctx context.Context, config *channel.ChannelConfig, payload []byte) error {
	url, _ := config.Settings["url"].(string)
	
	// 获取HTTP方法
	method := "POST"
	if methodInterface, exists := config.Settings["method"]; exists {
		if methodStr, ok := methodInterface.(string); ok {
			method = strings.ToUpper(methodStr)
		}
	}
	
	// 创建请求
	req, err := http.NewRequestWithContext(ctx, method, url, bytes.NewBuffer(payload))
	if err != nil {
		return fmt.Errorf("创建请求失败: %w", err)
	}
	
	// 设置Content-Type
	contentType := "application/json"
	if contentTypeInterface, exists := config.Settings["content_type"]; exists {
		if contentTypeStr, ok := contentTypeInterface.(string); ok {
			contentType = contentTypeStr
		}
	}
	req.Header.Set("Content-Type", contentType)
	req.Header.Set("User-Agent", "AlertAgent-Webhook/1.0")
	
	// 添加自定义头
	if headersInterface, exists := config.Settings["headers"]; exists {
		if headers, ok := headersInterface.(map[string]interface{}); ok {
			for key, value := range headers {
				if strValue, ok := value.(string); ok {
					req.Header.Set(key, strValue)
				}
			}
		}
	}
	
	// 添加签名
	if secret, exists := config.Settings["secret"].(string); exists && secret != "" {
		signature := p.generateSignature(payload, secret)
		req.Header.Set("X-Signature", signature)
	}
	
	// 发送请求
	resp, err := p.client.Do(req)
	if err != nil {
		return fmt.Errorf("发送请求失败: %w", err)
	}
	defer resp.Body.Close()
	
	// 检查响应状态码
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("HTTP状态码: %d, 响应: %s", resp.StatusCode, string(body))
	}
	
	return nil
}

// generateSignature 生成HMAC签名
func (p *WebhookPlugin) generateSignature(payload []byte, secret string) string {
	h := hmac.New(sha256.New, []byte(secret))
	h.Write(payload)
	return "sha256=" + hex.EncodeToString(h.Sum(nil))
}