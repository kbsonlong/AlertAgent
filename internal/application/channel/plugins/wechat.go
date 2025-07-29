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

	"alert_agent/internal/domain/channel"
	"alert_agent/pkg/types"
)

// WeChatPlugin 企业微信插件
type WeChatPlugin struct {
	client *http.Client
}

// NewWeChatPlugin 创建企业微信插件实例
func NewWeChatPlugin() *WeChatPlugin {
	return &WeChatPlugin{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// GetType 获取插件类型
func (p *WeChatPlugin) GetType() channel.ChannelType {
	return channel.ChannelTypeWeChat
}

// GetName 获取插件名称
func (p *WeChatPlugin) GetName() string {
	return "企业微信"
}

// GetVersion 获取插件版本
func (p *WeChatPlugin) GetVersion() string {
	return "1.0.0"
}

// GetDescription 获取插件描述
func (p *WeChatPlugin) GetDescription() string {
	return "企业微信群机器人消息发送插件，支持文本、Markdown格式消息"
}

// GetConfigSchema 获取配置模式
func (p *WeChatPlugin) GetConfigSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"webhook_url": map[string]interface{}{
				"type":        "string",
				"description": "企业微信群机器人Webhook URL",
				"format":      "uri",
			},
			"mentioned_list": map[string]interface{}{
				"type":        "array",
				"description": "@成员列表（手机号或@all）",
				"items": map[string]interface{}{
					"type": "string",
				},
			},
			"mentioned_mobile_list": map[string]interface{}{
				"type":        "array",
				"description": "@成员手机号列表",
				"items": map[string]interface{}{
					"type": "string",
				},
			},
			"message_type": map[string]interface{}{
				"type":        "string",
				"description": "消息类型",
				"enum":        []string{"text", "markdown"},
				"default":     "markdown",
			},
		},
		"required": []string{"webhook_url"},
	}
}

// Initialize 初始化插件
func (p *WeChatPlugin) Initialize(ctx context.Context, config map[string]interface{}) error {
	// 这里可以进行插件级别的初始化
	return nil
}

// Start 启动插件
func (p *WeChatPlugin) Start(ctx context.Context) error {
	return nil
}

// Stop 停止插件
func (p *WeChatPlugin) Stop(ctx context.Context) error {
	return nil
}

// HealthCheck 健康检查
func (p *WeChatPlugin) HealthCheck(ctx context.Context) error {
	// 简单的健康检查，检查插件是否正常运行
	return nil
}

// SendMessage 发送消息
func (p *WeChatPlugin) SendMessage(ctx context.Context, config channel.ChannelConfig, message *types.Message) (*channel.SendResult, error) {
	start := time.Now()
	
	// 构建企业微信消息
	wechatMsg, err := p.buildWeChatMessage(&config, message)
	if err != nil {
		return nil, fmt.Errorf("构建企业微信消息失败: %w", err)
	}
	
	// 获取配置
	webhookURL, _ := config.Settings["webhook_url"].(string)
	resp, err := p.sendToWeChat(ctx, webhookURL, wechatMsg)
	if err != nil {
		return &channel.SendResult{
			Success:   false,
			Error:     err.Error(),
			Latency:   time.Since(start),
			Timestamp: time.Now(),
		}, err
	}
	
	return &channel.SendResult{
		MessageID: resp.MessageID,
		Success:   true,
		Latency:   time.Since(start),
		Timestamp: time.Now(),
		Metadata: map[string]interface{}{
			"response": resp,
		},
	}, nil
}

// ValidateConfig 验证配置
func (p *WeChatPlugin) ValidateConfig(config channel.ChannelConfig) error {
	webhookURL, ok := config.Settings["webhook_url"].(string)
	if !ok || webhookURL == "" {
		return fmt.Errorf("webhook_url 不能为空")
	}
	
	if !strings.HasPrefix(webhookURL, "https://qyapi.weixin.qq.com/cgi-bin/webhook/send?") {
		return fmt.Errorf("webhook_url 格式不正确，应为企业微信群机器人URL")
	}
	
	// 验证消息类型
	if msgType, exists := config.Settings["message_type"]; exists {
		if msgTypeStr, ok := msgType.(string); ok {
			if msgTypeStr != "text" && msgTypeStr != "markdown" {
				return fmt.Errorf("message_type 必须为 text 或 markdown")
			}
		}
	}
	
	return nil
}

// TestConnection 测试连接
func (p *WeChatPlugin) TestConnection(ctx context.Context, config channel.ChannelConfig) (*channel.TestResult, error) {
	start := time.Now()
	
	// 验证配置
	if err := p.ValidateConfig(config); err != nil {
		return &channel.TestResult{
			Success:   false,
			Message:   fmt.Sprintf("配置验证失败: %v", err),
			Latency:   time.Since(start).Milliseconds(),
			Timestamp: time.Now().Unix(),
		}, err
	}
	
	// 发送测试消息
	testMessage := &types.Message{
		Title:    "AlertAgent Test",
		Content:  "This is a test message from AlertAgent",
		Priority: types.PriorityLow,
	}
// 发送测试消息
	result, err := p.SendMessage(ctx, config, testMessage)
	if err != nil {
		return &channel.TestResult{
			Success:   false,
			Message:   fmt.Sprintf("发送测试消息失败: %v", err),
			Latency:   time.Since(start).Milliseconds(),
			Timestamp: time.Now().Unix(),
			Details: map[string]interface{}{
				"error": err.Error(),
			},
		}, err
	}
	
	return &channel.TestResult{
		Success:   true,
		Message:   "企业微信连接测试成功",
		Latency:   result.Latency.Milliseconds(),
		Timestamp: time.Now().Unix(),
		Details: map[string]interface{}{
			"message_id": result.MessageID,
			"latency_ms": result.Latency.Milliseconds(),
		},
	}, nil
}

// GetCapabilities 获取插件支持的功能
func (p *WeChatPlugin) GetCapabilities() []channel.PluginCapability {
	return []channel.PluginCapability{
		channel.CapabilityTextMessage,
		channel.CapabilityMarkdownMessage,
		channel.CapabilityTemplating,
		channel.CapabilityHealthCheck,
	}
}

// SupportsFeature 检查是否支持特定功能
func (p *WeChatPlugin) SupportsFeature(feature channel.PluginCapability) bool {
	capabilities := p.GetCapabilities()
	for _, cap := range capabilities {
		if cap == feature {
			return true
		}
	}
	return false
}

// WeChatMessage 企业微信消息结构
type WeChatMessage struct {
	MsgType string `json:"msgtype"`
	Text    *struct {
		Content             string   `json:"content"`
		MentionedList       []string `json:"mentioned_list,omitempty"`
		MentionedMobileList []string `json:"mentioned_mobile_list,omitempty"`
	} `json:"text,omitempty"`
	Markdown *struct {
		Content string `json:"content"`
	} `json:"markdown,omitempty"`
}

// WeChatResponse 企业微信响应
type WeChatResponse struct {
	ErrCode   int    `json:"errcode"`
	ErrMsg    string `json:"errmsg"`
	MessageID string `json:"msgid,omitempty"`
}

// buildWeChatMessage 构建企业微信消息
func (p *WeChatPlugin) buildWeChatMessage(config *channel.ChannelConfig, message *types.Message) (*WeChatMessage, error) {
	msgType := "markdown"
	if mt, exists := config.Settings["message_type"]; exists {
		if mtStr, ok := mt.(string); ok {
			msgType = mtStr
		}
	}
	
	wechatMsg := &WeChatMessage{
		MsgType: msgType,
	}
	
	switch msgType {
	case "text":
		content := fmt.Sprintf("%s\n\n%s", message.Title, message.Content)
		wechatMsg.Text = &struct {
			Content             string   `json:"content"`
			MentionedList       []string `json:"mentioned_list,omitempty"`
			MentionedMobileList []string `json:"mentioned_mobile_list,omitempty"`
		}{
			Content: content,
		}
		
		// 添加@成员
		if mentionedList, exists := config.Settings["mentioned_list"]; exists {
			if list, ok := mentionedList.([]interface{}); ok {
				for _, item := range list {
					if str, ok := item.(string); ok {
						wechatMsg.Text.MentionedList = append(wechatMsg.Text.MentionedList, str)
					}
				}
			}
		}
		
		if mentionedMobileList, exists := config.Settings["mentioned_mobile_list"]; exists {
			if list, ok := mentionedMobileList.([]interface{}); ok {
				for _, item := range list {
					if str, ok := item.(string); ok {
						wechatMsg.Text.MentionedMobileList = append(wechatMsg.Text.MentionedMobileList, str)
					}
				}
			}
		}
		
	case "markdown":
		content := p.formatMarkdownContent(message)
		wechatMsg.Markdown = &struct {
			Content string `json:"content"`
		}{
			Content: content,
		}
		
	default:
		return nil, fmt.Errorf("不支持的消息类型: %s", msgType)
	}
	
	return wechatMsg, nil
}

// formatMarkdownContent 格式化Markdown内容
func (p *WeChatPlugin) formatMarkdownContent(message *types.Message) string {
	content := fmt.Sprintf("## %s\n\n", message.Title)
	
	// 添加优先级标识
	switch message.Priority {
	case types.PriorityCritical:
		content += "<font color=\"warning\">🔴 **严重告警**</font>\n\n"
	case types.PriorityHigh:
		content += "<font color=\"warning\">🟡 **高优先级**</font>\n\n"
	case types.PriorityMedium:
		content += "<font color=\"info\">🔵 **中等优先级**</font>\n\n"
	case types.PriorityLow:
		content += "<font color=\"comment\">ℹ️ **低优先级**</font>\n\n"
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
			content += fmt.Sprintf("> **%s**: %v\n", key, value)
		}
	}
	
	return content
}

// sendToWeChat 发送消息到企业微信
func (p *WeChatPlugin) sendToWeChat(ctx context.Context, webhookURL string, message *WeChatMessage) (*WeChatResponse, error) {
	// 序列化消息
	payload, err := json.Marshal(message)
	if err != nil {
		return nil, fmt.Errorf("序列化消息失败: %w", err)
	}
	
	// 创建请求
	req, err := http.NewRequestWithContext(ctx, "POST", webhookURL, bytes.NewBuffer(payload))
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}
	
	req.Header.Set("Content-Type", "application/json")
	
	// 发送请求
	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("发送请求失败: %w", err)
	}
	defer resp.Body.Close()
	
	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}
	
	// 解析响应
	var wechatResp WeChatResponse
	if err := json.Unmarshal(body, &wechatResp); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}
	
	// 检查错误码
	if wechatResp.ErrCode != 0 {
		return &wechatResp, fmt.Errorf("企业微信API错误: %s (错误码: %d)", wechatResp.ErrMsg, wechatResp.ErrCode)
	}
	
	return &wechatResp, nil
}