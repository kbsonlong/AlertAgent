package plugins

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"alert_agent/internal/domain/channel"
	"alert_agent/pkg/types"
)

// DingTalkPlugin 钉钉插件
type DingTalkPlugin struct {
	name        string
	version     string
	description string
	status      channel.PluginStatus
	client      *http.Client
}

// NewDingTalkPlugin 创建钉钉插件
func NewDingTalkPlugin() *DingTalkPlugin {
	return &DingTalkPlugin{
		name:        "dingtalk",
		version:     "1.0.0",
		description: "DingTalk notification plugin",
		status:      channel.PluginStatusActive,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// GetType 获取插件类型
func (p *DingTalkPlugin) GetType() channel.ChannelType {
	return channel.ChannelTypeDingTalk
}

// GetName 获取插件名称
func (p *DingTalkPlugin) GetName() string {
	return p.name
}

// GetVersion 获取插件版本
func (p *DingTalkPlugin) GetVersion() string {
	return p.version
}

// GetDescription 获取插件描述
func (p *DingTalkPlugin) GetDescription() string {
	return p.description
}

// GetConfigSchema 获取配置模式
func (p *DingTalkPlugin) GetConfigSchema() map[string]interface{} {
	return p.getConfigSchema()
}

// Initialize 初始化插件
func (p *DingTalkPlugin) Initialize(ctx context.Context, config map[string]interface{}) error {
	p.status = channel.PluginStatusActive
	return nil
}

// Start 启动插件
func (p *DingTalkPlugin) Start(ctx context.Context) error {
	p.status = channel.PluginStatusActive
	return nil
}

// Stop 停止插件
func (p *DingTalkPlugin) Stop(ctx context.Context) error {
	p.status = channel.PluginStatusInactive
	return nil
}

// HealthCheck 健康检查
func (p *DingTalkPlugin) HealthCheck(ctx context.Context) error {
	return nil
}

// SendMessage 发送消息
func (p *DingTalkPlugin) SendMessage(ctx context.Context, config channel.ChannelConfig, message *types.Message) (*channel.SendResult, error) {
	start := time.Now()
	
	// 解析配置
	webhookURL, ok := config.Settings["webhook_url"].(string)
	if !ok || webhookURL == "" {
		return &channel.SendResult{
			Success:   false,
			Error:     "webhook_url is required",
			Timestamp: time.Now(),
			Latency:   time.Since(start),
		}, fmt.Errorf("webhook_url is required")
	}
	
	// 构建钉钉消息
	dingMsg, err := p.buildDingTalkMessage(config.Settings, message)
	if err != nil {
		return &channel.SendResult{
			Success:   false,
			Error:     err.Error(),
			Timestamp: time.Now(),
			Latency:   time.Since(start),
		}, err
	}
	
	// 添加签名（如果配置了密钥）
	if secret, ok := config.Settings["secret"].(string); ok && secret != "" {
		webhookURL, err = p.addSignature(webhookURL, secret)
		if err != nil {
			return &channel.SendResult{
				Success:   false,
				Error:     err.Error(),
				Timestamp: time.Now(),
				Latency:   time.Since(start),
			}, err
		}
	}
	
	// 发送请求
	body, err := json.Marshal(dingMsg)
	if err != nil {
		return &channel.SendResult{
			Success:   false,
			Error:     err.Error(),
			Timestamp: time.Now(),
			Latency:   time.Since(start),
		}, err
	}
	
	req, err := http.NewRequestWithContext(ctx, "POST", webhookURL, bytes.NewBuffer(body))
	if err != nil {
		return &channel.SendResult{
			Success:   false,
			Error:     err.Error(),
			Timestamp: time.Now(),
			Latency:   time.Since(start),
		}, err
	}
	
	req.Header.Set("Content-Type", "application/json")
	
	resp, err := p.client.Do(req)
	if err != nil {
		return &channel.SendResult{
			Success:   false,
			Error:     err.Error(),
			Timestamp: time.Now(),
			Latency:   time.Since(start),
		}, err
	}
	defer resp.Body.Close()
	
	// 解析响应
	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return &channel.SendResult{
			Success:   false,
			Error:     err.Error(),
			Timestamp: time.Now(),
			Latency:   time.Since(start),
		}, err
	}
	
	// 检查响应状态
	if errCode, ok := result["errcode"].(float64); ok && errCode != 0 {
		errMsg := "unknown error"
		if msg, ok := result["errmsg"].(string); ok {
			errMsg = msg
		}
		return &channel.SendResult{
			Success:   false,
			Error:     fmt.Sprintf("DingTalk API error: %s (code: %.0f)", errMsg, errCode),
			Timestamp: time.Now(),
			Latency:   time.Since(start),
		}, fmt.Errorf("DingTalk API error: %s", errMsg)
	}
	
	return &channel.SendResult{
		Success:   true,
		MessageID: fmt.Sprintf("dingtalk_%d", time.Now().Unix()),
		Timestamp: time.Now(),
		Latency:   time.Since(start),
	}, nil
}

// ValidateConfig 验证配置
func (p *DingTalkPlugin) ValidateConfig(config channel.ChannelConfig) error {
	// 检查必需字段
	webhookURL, ok := config.Settings["webhook_url"].(string)
	if !ok || webhookURL == "" {
		return fmt.Errorf("webhook_url is required")
	}
	
	// 验证URL格式
	if _, err := url.Parse(webhookURL); err != nil {
		return fmt.Errorf("invalid webhook_url: %w", err)
	}
	
	// 验证可选字段
	if secret, ok := config.Settings["secret"]; ok {
		if _, ok := secret.(string); !ok {
			return fmt.Errorf("secret must be a string")
		}
	}
	
	if atMobiles, ok := config.Settings["at_mobiles"]; ok {
		if _, ok := atMobiles.([]interface{}); !ok {
			return fmt.Errorf("at_mobiles must be an array")
		}
	}
	
	if atUserIds, ok := config.Settings["at_user_ids"]; ok {
		if _, ok := atUserIds.([]interface{}); !ok {
			return fmt.Errorf("at_user_ids must be an array")
		}
	}
	
	if isAtAll, ok := config.Settings["is_at_all"]; ok {
		if _, ok := isAtAll.(bool); !ok {
			return fmt.Errorf("is_at_all must be a boolean")
		}
	}
	
	return nil
}

// TestConnection 测试连接
func (p *DingTalkPlugin) TestConnection(ctx context.Context, config channel.ChannelConfig) (*channel.TestResult, error) {
	// 发送测试消息
	testMessage := &types.Message{
		Title:    "AlertAgent Test",
		Content:  "This is a test message from AlertAgent",
		Priority: types.PriorityLow,
	}
	
	result, err := p.SendMessage(ctx, config, testMessage)
	if err != nil {
		return &channel.TestResult{
			Success: false,
			Message: err.Error(),
			Latency: result.Latency.Milliseconds(),
		}, err
	}
	
	message := "测试连接成功"
	if result.Error != "" {
		message = result.Error
	}
	
	return &channel.TestResult{
		Success: result.Success,
		Message: message,
		Latency: result.Latency.Milliseconds(),
	}, nil
}

// GetCapabilities 获取插件能力
func (p *DingTalkPlugin) GetCapabilities() []channel.PluginCapability {
	return []channel.PluginCapability{
		channel.CapabilityTextMessage,
		channel.CapabilityMarkdownMessage,
		channel.CapabilityTemplating,
	}
}

// SupportsFeature 检查是否支持特定功能
func (p *DingTalkPlugin) SupportsFeature(feature channel.PluginCapability) bool {
	capabilities := p.GetCapabilities()
	for _, cap := range capabilities {
		if cap == feature {
			return true
		}
	}
	return false
}

// buildDingTalkMessage 构建钉钉消息
func (p *DingTalkPlugin) buildDingTalkMessage(config map[string]interface{}, message *types.Message) (map[string]interface{}, error) {
	// 基础消息结构
	dingMsg := map[string]interface{}{
		"msgtype": "markdown",
		"markdown": map[string]interface{}{
			"title": message.Title,
			"text":  p.formatMarkdownContent(message),
		},
	}
	
	// 添加@功能
	atInfo := map[string]interface{}{}
	
	// @手机号
	if atMobiles, ok := config["at_mobiles"].([]interface{}); ok && len(atMobiles) > 0 {
		mobiles := make([]string, 0, len(atMobiles))
		for _, mobile := range atMobiles {
			if mobileStr, ok := mobile.(string); ok {
				mobiles = append(mobiles, mobileStr)
			}
		}
		if len(mobiles) > 0 {
			atInfo["atMobiles"] = mobiles
		}
	}
	
	// @用户ID
	if atUserIds, ok := config["at_user_ids"].([]interface{}); ok && len(atUserIds) > 0 {
		userIds := make([]string, 0, len(atUserIds))
		for _, userId := range atUserIds {
			if userIdStr, ok := userId.(string); ok {
				userIds = append(userIds, userIdStr)
			}
		}
		if len(userIds) > 0 {
			atInfo["atUserIds"] = userIds
		}
	}
	
	// @所有人
	if isAtAll, ok := config["is_at_all"].(bool); ok {
		atInfo["isAtAll"] = isAtAll
	}
	
	if len(atInfo) > 0 {
		dingMsg["at"] = atInfo
	}
	
	return dingMsg, nil
}

// formatMarkdownContent 格式化Markdown内容
func (p *DingTalkPlugin) formatMarkdownContent(message *types.Message) string {
	content := fmt.Sprintf("## %s\n\n", message.Title)
	
	// 添加优先级标识
	switch message.Priority {
	case types.PriorityCritical:
		content += "🔴 **严重告警**\n\n"
	case types.PriorityHigh:
		content += "🟡 **高优先级**\n\n"
	case types.PriorityMedium:
		content += "🔵 **中等优先级**\n\n"
	case types.PriorityLow:
		content += "ℹ️ **低优先级**\n\n"
	default:
		content += "ℹ️ **通知**\n\n"
	}
	
	// 添加内容
	content += message.Content + "\n\n"
	
	// 添加时间戳
	if !message.CreatedAt.IsZero() {
		content += fmt.Sprintf("**时间**: %s\n\n", message.CreatedAt.Format("2006-01-02 15:04:05"))
	}
	
	// 添加类型
	if message.Type != "" {
		content += fmt.Sprintf("**类型**: `%s`\n\n", message.Type)
	}
	
	// 添加额外数据
	if len(message.Data) > 0 {
		content += "**详细信息**:\n\n"
		for key, value := range message.Data {
			content += fmt.Sprintf("- **%s**: %v\n", key, value)
		}
	}
	
	return content
}

// addSignature 添加签名
func (p *DingTalkPlugin) addSignature(webhookURL, secret string) (string, error) {
	timestamp := time.Now().UnixNano() / 1e6
	stringToSign := fmt.Sprintf("%d\n%s", timestamp, secret)
	
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(stringToSign))
	signature := base64.StdEncoding.EncodeToString(h.Sum(nil))
	
	u, err := url.Parse(webhookURL)
	if err != nil {
		return "", err
	}
	
	q := u.Query()
	q.Set("timestamp", strconv.FormatInt(timestamp, 10))
	q.Set("sign", signature)
	u.RawQuery = q.Encode()
	
	return u.String(), nil
}

// getConfigSchema 获取配置模式
func (p *DingTalkPlugin) getConfigSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"webhook_url": map[string]interface{}{
				"type":        "string",
				"description": "DingTalk webhook URL",
				"required":    true,
			},
			"secret": map[string]interface{}{
				"type":        "string",
				"description": "DingTalk webhook secret for signature",
				"required":    false,
			},
			"at_mobiles": map[string]interface{}{
				"type":        "array",
				"items":       map[string]string{"type": "string"},
				"description": "Mobile numbers to mention",
				"required":    false,
			},
			"at_user_ids": map[string]interface{}{
				"type":        "array",
				"items":       map[string]string{"type": "string"},
				"description": "User IDs to mention",
				"required":    false,
			},
			"is_at_all": map[string]interface{}{
				"type":        "boolean",
				"description": "Whether to mention all members",
				"required":    false,
				"default":     false,
			},
		},
		"required": []string{"webhook_url"},
	}
}