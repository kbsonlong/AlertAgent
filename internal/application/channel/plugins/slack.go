package plugins

import (
	"alert_agent/internal/domain/channel"
	"alert_agent/pkg/types"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// SlackPlugin Slack插件实现
type SlackPlugin struct {
	client *http.Client
	status channel.PluginStatus
}

// SlackMessage Slack消息结构
type SlackMessage struct {
	Channel     string            `json:"channel,omitempty"`
	Text        string            `json:"text"`
	Username    string            `json:"username,omitempty"`
	IconEmoji   string            `json:"icon_emoji,omitempty"`
	IconURL     string            `json:"icon_url,omitempty"`
	Attachments []SlackAttachment `json:"attachments,omitempty"`
	Blocks      []SlackBlock      `json:"blocks,omitempty"`
}

// SlackAttachment Slack附件
type SlackAttachment struct {
	Color      string       `json:"color,omitempty"`
	Title      string       `json:"title,omitempty"`
	TitleLink  string       `json:"title_link,omitempty"`
	Text       string       `json:"text,omitempty"`
	Fields     []SlackField `json:"fields,omitempty"`
	Footer     string       `json:"footer,omitempty"`
	Timestamp  int64        `json:"ts,omitempty"`
	MarkdownIn []string     `json:"mrkdwn_in,omitempty"`
}

// SlackField Slack字段
type SlackField struct {
	Title string `json:"title"`
	Value string `json:"value"`
	Short bool   `json:"short"`
}

// SlackBlock Slack块
type SlackBlock struct {
	Type string      `json:"type"`
	Text *SlackText  `json:"text,omitempty"`
}

// SlackText Slack文本
type SlackText struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// NewSlackPlugin 创建新的Slack插件实例
func NewSlackPlugin() *SlackPlugin {
	return &SlackPlugin{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		status: channel.PluginStatusLoaded,
	}
}

// GetType 获取插件类型
func (p *SlackPlugin) GetType() channel.ChannelType {
	return channel.ChannelTypeSlack
}

// GetName 获取插件名称
func (p *SlackPlugin) GetName() string {
	return "Slack"
}

// GetVersion 获取插件版本
func (p *SlackPlugin) GetVersion() string {
	return "1.0.0"
}

// GetDescription 获取插件描述
func (p *SlackPlugin) GetDescription() string {
	return "Slack通知插件，支持发送消息到Slack频道"
}

// GetConfigSchema 获取配置模式
func (p *SlackPlugin) GetConfigSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"webhook_url": map[string]interface{}{
				"type":        "string",
				"description": "Slack Webhook URL",
				"format":      "uri",
			},
			"token": map[string]interface{}{
				"type":        "string",
				"description": "Slack Bot Token (以xoxb-开头)",
			},
			"channel": map[string]interface{}{
				"type":        "string",
				"description": "默认频道名称或ID",
			},
			"username": map[string]interface{}{
				"type":        "string",
				"description": "机器人用户名",
				"default":     "AlertAgent",
			},
			"icon_emoji": map[string]interface{}{
				"type":        "string",
				"description": "机器人图标emoji",
				"default":     ":warning:",
			},
			"icon_url": map[string]interface{}{
				"type":        "string",
				"description": "机器人图标URL",
				"format":      "uri",
			},
			"use_blocks": map[string]interface{}{
				"type":        "boolean",
				"description": "使用Slack Blocks格式",
				"default":     false,
			},
			"mention_users": map[string]interface{}{
				"type":        "array",
				"description": "需要@的用户ID列表",
				"items": map[string]interface{}{
					"type": "string",
				},
			},
			"mention_channel": map[string]interface{}{
				"type":        "boolean",
				"description": "是否@channel",
				"default":     false,
			},
			"timeout": map[string]interface{}{
				"type":        "integer",
				"description": "请求超时时间（秒）",
				"minimum":     1,
				"maximum":     300,
				"default":     30,
			},
		},
		"anyOf": []map[string]interface{}{
			{"required": []string{"webhook_url"}},
			{"required": []string{"token", "channel"}},
		},
	}
}

// Initialize 初始化插件
func (p *SlackPlugin) Initialize(ctx context.Context, config map[string]interface{}) error {
	// 这里可以进行插件级别的初始化
	return nil
}

// Start 启动插件
func (p *SlackPlugin) Start(ctx context.Context) error {
	p.status = channel.PluginStatusActive
	return nil
}

// Stop 停止插件
func (p *SlackPlugin) Stop(ctx context.Context) error {
	p.status = channel.PluginStatusInactive
	return nil
}

// HealthCheck 健康检查
func (p *SlackPlugin) HealthCheck(ctx context.Context) error {
	return nil
}

// SendMessage 发送消息
func (p *SlackPlugin) SendMessage(ctx context.Context, config channel.ChannelConfig, message *types.Message) (*channel.SendResult, error) {
	start := time.Now()
	result := &channel.SendResult{
		ChannelID: "", // 由调用方设置
		Success:   false,
		Timestamp: start,
	}
	
	// 构建Slack消息
	slackMsg, err := p.buildSlackMessage(&config, message)
	if err != nil {
		result.Error = fmt.Sprintf("构建消息失败: %v", err)
		result.Latency = time.Since(start)
		return result, err
	}
	
	// 发送消息
	err = p.sendSlackMessage(ctx, &config, slackMsg)
	if err != nil {
		result.Error = fmt.Sprintf("发送失败: %v", err)
		result.Latency = time.Since(start)
		return result, err
	}
	
	result.Success = true
	result.Latency = time.Since(start)
	return result, nil
}

// ValidateConfig 验证配置
func (p *SlackPlugin) ValidateConfig(config channel.ChannelConfig) error {

	
	// 检查是否有webhook_url或token+channel
	webhookURL, hasWebhook := config.Settings["webhook_url"].(string)
	token, hasToken := config.Settings["token"].(string)
	channel, hasChannel := config.Settings["channel"].(string)
	
	if hasWebhook && webhookURL != "" {
		// 使用Webhook模式
		if !strings.HasPrefix(webhookURL, "https://hooks.slack.com/") {
			return fmt.Errorf("无效的Slack Webhook URL")
		}
	} else if hasToken && hasChannel && token != "" && channel != "" {
		// 使用Bot Token模式
		if !strings.HasPrefix(token, "xoxb-") {
			return fmt.Errorf("无效的Slack Bot Token，必须以xoxb-开头")
		}
	} else {
		return fmt.Errorf("必须配置webhook_url或者token+channel")
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
	
	return nil
}

// TestConnection 测试连接
func (p *SlackPlugin) TestConnection(ctx context.Context, config channel.ChannelConfig) (*channel.TestResult, error) {
	start := time.Now()
	result := &channel.TestResult{
		Success:   false,
		Timestamp: start.Unix(),
	}
	
	// 发送测试消息
	testMessage := &types.Message{
		Title:     "AlertAgent连接测试",
		Content:   "这是一条来自AlertAgent的测试消息",
		Priority:  types.PriorityLow,
		CreatedAt: time.Now(),
	}
	
	slackMsg, err := p.buildSlackMessage(&config, testMessage)
	if err != nil {
		result.Message = fmt.Sprintf("构建测试消息失败: %v", err)
		result.Latency = time.Since(start).Milliseconds()
		return result, err
	}
	
	err = p.sendSlackMessage(ctx, &config, slackMsg)
	if err != nil {
		result.Message = fmt.Sprintf("发送测试消息失败: %v", err)
		result.Latency = time.Since(start).Milliseconds()
		return result, err
	}
	
	result.Success = true
	result.Message = "连接测试成功"
	result.Latency = time.Since(start).Milliseconds()
	return result, nil
}

// GetCapabilities 获取插件支持的功能
func (p *SlackPlugin) GetCapabilities() []channel.PluginCapability {
	return []channel.PluginCapability{
		channel.CapabilityTextMessage,
		channel.CapabilityMarkdownMessage,
		channel.CapabilityAttachments,
		channel.CapabilityTemplating,
		channel.CapabilityHealthCheck,
	}
}

// SupportsFeature 检查是否支持特定功能
func (p *SlackPlugin) SupportsFeature(feature channel.PluginCapability) bool {
	capabilities := p.GetCapabilities()
	for _, cap := range capabilities {
		if cap == feature {
			return true
		}
	}
	return false
}

// buildSlackMessage 构建Slack消息
func (p *SlackPlugin) buildSlackMessage(config *channel.ChannelConfig, message *types.Message) (*SlackMessage, error) {
	slackMsg := &SlackMessage{}
	
	// 设置频道
	if channel, exists := config.Settings["channel"].(string); exists && channel != "" {
		slackMsg.Channel = channel
	}
	
	// 设置用户名
	if username, exists := config.Settings["username"].(string); exists && username != "" {
		slackMsg.Username = username
	} else {
		slackMsg.Username = "AlertAgent"
	}
	
	// 设置图标
	if iconEmoji, exists := config.Settings["icon_emoji"].(string); exists && iconEmoji != "" {
		slackMsg.IconEmoji = iconEmoji
	} else if iconURL, exists := config.Settings["icon_url"].(string); exists && iconURL != "" {
		slackMsg.IconURL = iconURL
	} else {
		slackMsg.IconEmoji = ":warning:"
	}
	
	// 检查是否使用Blocks格式
	useBlocks := false
	if useBlocksInterface, exists := config.Settings["use_blocks"]; exists {
		if useBlocksBool, ok := useBlocksInterface.(bool); ok {
			useBlocks = useBlocksBool
		}
	}
	
	if useBlocks {
		// 使用Blocks格式
		slackMsg.Blocks = p.buildSlackBlocks(message)
	} else {
		// 使用Attachments格式
		slackMsg.Text = message.Title
		slackMsg.Attachments = p.buildSlackAttachments(message)
	}
	
	// 添加@mentions
	if mentionUsers, exists := config.Settings["mention_users"]; exists {
		if users, ok := mentionUsers.([]interface{}); ok {
			mentions := ""
			for _, user := range users {
				if userStr, ok := user.(string); ok {
					mentions += fmt.Sprintf("<@%s> ", userStr)
				}
			}
			if mentions != "" {
				slackMsg.Text = mentions + slackMsg.Text
			}
		}
	}
	
	// 添加@channel
	if mentionChannel, exists := config.Settings["mention_channel"]; exists {
		if mentionChannelBool, ok := mentionChannel.(bool); ok && mentionChannelBool {
			slackMsg.Text = "<!channel> " + slackMsg.Text
		}
	}
	
	return slackMsg, nil
}

// buildSlackAttachments 构建Slack附件
func (p *SlackPlugin) buildSlackAttachments(message *types.Message) []SlackAttachment {
	attachment := SlackAttachment{
		Title:      message.Title,
		Text:       message.Content,
		Footer:     "AlertAgent",
		Timestamp:  message.CreatedAt.Unix(),
		MarkdownIn: []string{"text", "fields"},
	}
	
	// 根据优先级设置颜色
	switch message.Priority {
	case types.PriorityCritical:
		attachment.Color = "danger"
	case types.PriorityHigh:
		attachment.Color = "warning"
	case types.PriorityMedium:
		attachment.Color = "#439FE0"
	case types.PriorityLow:
		attachment.Color = "good"
	default:
		attachment.Color = "#439FE0"
	}
	
	// 添加字段
	fields := []SlackField{
		{
			Title: "优先级",
			Value: p.getPriorityText(message.Priority),
			Short: true,
		},
	}
	
	if message.Type != "" {
		fields = append(fields, SlackField{
			Title: "类型",
			Value: message.Type,
			Short: true,
		})
	}
	
	// 添加额外数据
	if len(message.Data) > 0 {
		for key, value := range message.Data {
			fields = append(fields, SlackField{
				Title: key,
				Value: fmt.Sprintf("%v", value),
				Short: true,
			})
		}
	}
	
	attachment.Fields = fields
	return []SlackAttachment{attachment}
}

// buildSlackBlocks 构建Slack块
func (p *SlackPlugin) buildSlackBlocks(message *types.Message) []SlackBlock {
	blocks := []SlackBlock{
		{
			Type: "header",
			Text: &SlackText{
				Type: "plain_text",
				Text: message.Title,
			},
		},
		{
			Type: "section",
			Text: &SlackText{
				Type: "mrkdwn",
				Text: message.Content,
			},
		},
	}
	
	// 添加优先级和其他信息
	infoText := fmt.Sprintf("*优先级:* %s", p.getPriorityText(message.Priority))
	if message.Type != "" {
		infoText += fmt.Sprintf("\n*类型:* %s", message.Type)
	}
	if !message.CreatedAt.IsZero() {
		infoText += fmt.Sprintf("\n*时间:* %s", message.CreatedAt.Format("2006-01-02 15:04:05"))
	}
	
	blocks = append(blocks, SlackBlock{
		Type: "section",
		Text: &SlackText{
			Type: "mrkdwn",
			Text: infoText,
		},
	})
	
	return blocks
}

// getPriorityText 获取优先级文本
func (p *SlackPlugin) getPriorityText(priority types.Priority) string {
	switch priority {
	case types.PriorityCritical:
		return "🔴 严重"
	case types.PriorityHigh:
		return "🟡 高"
	case types.PriorityMedium:
		return "🔵 中等"
	case types.PriorityLow:
		return "🟢 低"
	default:
		return "ℹ️ 通知"
	}
}

// sendSlackMessage 发送Slack消息
func (p *SlackPlugin) sendSlackMessage(ctx context.Context, config *channel.ChannelConfig, message *SlackMessage) error {
	// 序列化消息
	payload, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("序列化消息失败: %w", err)
	}
	
	// 检查使用哪种发送方式
	if webhookURL, exists := config.Settings["webhook_url"].(string); exists && webhookURL != "" {
		// 使用Webhook发送
		return p.sendViaWebhook(ctx, webhookURL, payload)
	} else if token, exists := config.Settings["token"].(string); exists && token != "" {
		// 使用API发送
		return p.sendViaAPI(ctx, token, payload)
	}
	
	return fmt.Errorf("未配置有效的发送方式")
}

// sendViaWebhook 通过Webhook发送
func (p *SlackPlugin) sendViaWebhook(ctx context.Context, webhookURL string, payload []byte) error {
	req, err := http.NewRequestWithContext(ctx, "POST", webhookURL, bytes.NewBuffer(payload))
	if err != nil {
		return fmt.Errorf("创建请求失败: %w", err)
	}
	
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "AlertAgent-Slack/1.0")
	
	resp, err := p.client.Do(req)
	if err != nil {
		return fmt.Errorf("发送请求失败: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("HTTP状态码: %d, 响应: %s", resp.StatusCode, string(body))
	}
	
	return nil
}

// sendViaAPI 通过API发送
func (p *SlackPlugin) sendViaAPI(ctx context.Context, token string, payload []byte) error {
	apiURL := "https://slack.com/api/chat.postMessage"
	
	req, err := http.NewRequestWithContext(ctx, "POST", apiURL, bytes.NewBuffer(payload))
	if err != nil {
		return fmt.Errorf("创建请求失败: %w", err)
	}
	
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("User-Agent", "AlertAgent-Slack/1.0")
	
	resp, err := p.client.Do(req)
	if err != nil {
		return fmt.Errorf("发送请求失败: %w", err)
	}
	defer resp.Body.Close()
	
	body, _ := io.ReadAll(resp.Body)
	
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP状态码: %d, 响应: %s", resp.StatusCode, string(body))
	}
	
	// 解析响应检查是否成功
	var response map[string]interface{}
	if err := json.Unmarshal(body, &response); err != nil {
		return fmt.Errorf("解析响应失败: %w", err)
	}
	
	if ok, exists := response["ok"].(bool); !exists || !ok {
		errorMsg := "未知错误"
		if errStr, exists := response["error"].(string); exists {
			errorMsg = errStr
		}
		return fmt.Errorf("Slack API错误: %s", errorMsg)
	}
	
	return nil
}