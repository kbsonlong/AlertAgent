package plugins

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/smtp"
	"strings"
	"time"

	"alert_agent/internal/domain/channel"
	"alert_agent/pkg/types"
)

// EmailPlugin 邮件插件
type EmailPlugin struct {
	auth smtp.Auth
}

// NewEmailPlugin 创建邮件插件实例
func NewEmailPlugin() *EmailPlugin {
	return &EmailPlugin{}
}

// GetType 获取插件类型
func (p *EmailPlugin) GetType() channel.ChannelType {
	return channel.ChannelTypeEmail
}

// GetName 获取插件名称
func (p *EmailPlugin) GetName() string {
	return "邮件"
}

// GetVersion 获取插件版本
func (p *EmailPlugin) GetVersion() string {
	return "1.0.0"
}

// GetDescription 获取插件描述
func (p *EmailPlugin) GetDescription() string {
	return "SMTP邮件发送插件，支持HTML和文本格式邮件"
}

// GetConfigSchema 获取配置模式
func (p *EmailPlugin) GetConfigSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"smtp_host": map[string]interface{}{
				"type":        "string",
				"description": "SMTP服务器地址",
			},
			"smtp_port": map[string]interface{}{
				"type":        "integer",
				"description": "SMTP服务器端口",
				"default":     587,
			},
			"username": map[string]interface{}{
				"type":        "string",
				"description": "SMTP用户名",
			},
			"password": map[string]interface{}{
				"type":        "string",
				"description": "SMTP密码",
				"format":      "password",
			},
			"from_email": map[string]interface{}{
				"type":        "string",
				"description": "发件人邮箱地址",
				"format":      "email",
			},
			"from_name": map[string]interface{}{
				"type":        "string",
				"description": "发件人姓名",
				"default":     "AlertAgent",
			},
			"to_emails": map[string]interface{}{
				"type":        "array",
				"description": "收件人邮箱列表",
				"items": map[string]interface{}{
					"type":   "string",
					"format": "email",
				},
			},
			"cc_emails": map[string]interface{}{
				"type":        "array",
				"description": "抄送邮箱列表",
				"items": map[string]interface{}{
					"type":   "string",
					"format": "email",
				},
			},
			"use_tls": map[string]interface{}{
				"type":        "boolean",
				"description": "是否使用TLS加密",
				"default":     true,
			},
			"use_html": map[string]interface{}{
				"type":        "boolean",
				"description": "是否使用HTML格式",
				"default":     true,
			},
		},
		"required": []string{"smtp_host", "smtp_port", "username", "password", "from_email", "to_emails"},
	}
}

// Initialize 初始化插件
func (p *EmailPlugin) Initialize(ctx context.Context, config map[string]interface{}) error {
	// 这里可以进行插件级别的初始化
	return nil
}

// Start 启动插件
func (p *EmailPlugin) Start(ctx context.Context) error {
	return nil
}

// Stop 停止插件
func (p *EmailPlugin) Stop(ctx context.Context) error {
	return nil
}

// HealthCheck 健康检查
func (p *EmailPlugin) HealthCheck(ctx context.Context) error {
	// 简单的健康检查，检查插件是否正常运行
	return nil
}

// SendMessage 发送消息
func (p *EmailPlugin) SendMessage(ctx context.Context, config channel.ChannelConfig, message *types.Message) (*channel.SendResult, error) {
	start := time.Now()
	
	// 构建邮件内容
	emailContent, err := p.buildEmailContent(&config, message)
	if err != nil {
		return nil, fmt.Errorf("构建邮件内容失败: %w", err)
	}
	
	// 发送邮件
	err = p.sendEmail(&config, emailContent)
	if err != nil {
		return &channel.SendResult{
			Success:   false,
			Error:     err.Error(),
			Latency:   time.Since(start),
			Timestamp: time.Now(),
		}, err
	}
	
	return &channel.SendResult{
		MessageID: fmt.Sprintf("email-%d", time.Now().UnixNano()),
		Success:   true,
		Latency:   time.Since(start),
		Timestamp: time.Now(),
		Metadata: map[string]interface{}{
			"subject": message.Title,
			"to":      config.Settings["to_emails"],
		},
	}, nil
}

// ValidateConfig 验证配置
func (p *EmailPlugin) ValidateConfig(config channel.ChannelConfig) error {
	// 验证必需字段
	requiredFields := []string{"smtp_host", "smtp_port", "username", "password", "from_email", "to_emails"}
	for _, field := range requiredFields {
		if _, exists := config.Settings[field]; !exists {
			return fmt.Errorf("缺少必需字段: %s", field)
		}
	}
	
	// 验证SMTP主机
	smtpHost, ok := config.Settings["smtp_host"].(string)
	if !ok || smtpHost == "" {
		return fmt.Errorf("smtp_host 不能为空")
	}
	
	// 验证SMTP端口
	smtpPort := config.Settings["smtp_port"]
	switch v := smtpPort.(type) {
	case int:
		if v <= 0 || v > 65535 {
			return fmt.Errorf("smtp_port 必须在1-65535范围内")
		}
	case float64:
		if v <= 0 || v > 65535 {
			return fmt.Errorf("smtp_port 必须在1-65535范围内")
		}
	default:
		return fmt.Errorf("smtp_port 必须为数字")
	}
	
	// 验证用户名和密码
	username, ok := config.Settings["username"].(string)
	if !ok || username == "" {
		return fmt.Errorf("username 不能为空")
	}
	
	password, ok := config.Settings["password"].(string)
	if !ok || password == "" {
		return fmt.Errorf("password 不能为空")
	}
	
	// 验证发件人邮箱
	fromEmail, ok := config.Settings["from_email"].(string)
	if !ok || fromEmail == "" {
		return fmt.Errorf("from_email 不能为空")
	}
	
	// 验证收件人邮箱列表
	toEmails, ok := config.Settings["to_emails"].([]interface{})
	if !ok || len(toEmails) == 0 {
		return fmt.Errorf("to_emails 不能为空")
	}
	
	for _, email := range toEmails {
		if emailStr, ok := email.(string); !ok || emailStr == "" {
			return fmt.Errorf("to_emails 包含无效的邮箱地址")
		}
	}
	
	return nil
}

// TestConnection 测试连接
func (p *EmailPlugin) TestConnection(ctx context.Context, config channel.ChannelConfig) (*channel.TestResult, error) {
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
	
	// 测试SMTP连接
	if err := p.testSMTPConnection(&config); err != nil {
		return &channel.TestResult{
			Success:   false,
			Message:   fmt.Sprintf("SMTP连接测试失败: %v", err),
			Latency:   time.Since(start).Milliseconds(),
			Timestamp: time.Now().Unix(),
			Details: map[string]interface{}{
				"error": err.Error(),
			},
		}, err
	}
	
	return &channel.TestResult{
		Success:   true,
		Message:   "邮件连接测试成功",
		Latency:   time.Since(start).Milliseconds(),
		Timestamp: time.Now().Unix(),
		Details: map[string]interface{}{
			"smtp_host": config.Settings["smtp_host"],
			"smtp_port": config.Settings["smtp_port"],
		},
	}, nil
}

// GetCapabilities 获取插件支持的功能
func (p *EmailPlugin) GetCapabilities() []channel.PluginCapability {
	return []channel.PluginCapability{
		channel.CapabilityTextMessage,
		channel.CapabilityHTMLMessage,
		channel.CapabilityAttachments,
		channel.CapabilityTemplating,
		channel.CapabilityHealthCheck,
	}
}

// SupportsFeature 检查是否支持特定功能
func (p *EmailPlugin) SupportsFeature(feature channel.PluginCapability) bool {
	capabilities := p.GetCapabilities()
	for _, cap := range capabilities {
		if cap == feature {
			return true
		}
	}
	return false
}

// EmailContent 邮件内容
type EmailContent struct {
	From    string
	To      []string
	CC      []string
	Subject string
	Body    string
	IsHTML  bool
}

// buildEmailContent 构建邮件内容
func (p *EmailPlugin) buildEmailContent(config *channel.ChannelConfig, message *types.Message) (*EmailContent, error) {
	// 获取配置
	fromEmail, _ := config.Settings["from_email"].(string)
	fromName, _ := config.Settings["from_name"].(string)
	if fromName == "" {
		fromName = "AlertAgent"
	}
	
	toEmailsInterface, _ := config.Settings["to_emails"].([]interface{})
	toEmails := make([]string, 0, len(toEmailsInterface))
	for _, email := range toEmailsInterface {
		if emailStr, ok := email.(string); ok {
			toEmails = append(toEmails, emailStr)
		}
	}
	
	var ccEmails []string
	if ccEmailsInterface, exists := config.Settings["cc_emails"]; exists {
		if ccList, ok := ccEmailsInterface.([]interface{}); ok {
			ccEmails = make([]string, 0, len(ccList))
			for _, email := range ccList {
				if emailStr, ok := email.(string); ok {
					ccEmails = append(ccEmails, emailStr)
				}
			}
		}
	}
	
	useHTML := true
	if useHTMLInterface, exists := config.Settings["use_html"]; exists {
		if useHTMLBool, ok := useHTMLInterface.(bool); ok {
			useHTML = useHTMLBool
		}
	}
	
	// 构建邮件内容
	body := p.formatEmailContent(message, useHTML)
	
	return &EmailContent{
		From:    fmt.Sprintf("%s <%s>", fromName, fromEmail),
		To:      toEmails,
		CC:      ccEmails,
		Subject: message.Title,
		Body:    body,
		IsHTML:  useHTML,
	}, nil
}

// formatEmailContent 格式化邮件内容
func (p *EmailPlugin) formatEmailContent(message *types.Message, useHTML bool) string {
	if useHTML {
		return p.formatHTMLContent(message)
	}
	return p.formatTextContent(message)
}

// formatHTMLContent 格式化HTML内容
func (p *EmailPlugin) formatHTMLContent(message *types.Message) string {
	html := `<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>` + message.Title + `</title>
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .header { background-color: #f4f4f4; padding: 20px; border-radius: 5px; margin-bottom: 20px; }
        .priority { padding: 5px 10px; border-radius: 3px; color: white; font-weight: bold; }
        .critical { background-color: #dc3545; }
        .high { background-color: #fd7e14; }
        .medium { background-color: #007bff; }
        .low { background-color: #28a745; }
        .content { margin: 20px 0; }
        .metadata { background-color: #f8f9fa; padding: 15px; border-radius: 5px; margin-top: 20px; }
        .footer { margin-top: 30px; padding-top: 20px; border-top: 1px solid #ddd; font-size: 12px; color: #666; }
    </style>
</head>
<body>
    <div class="header">
        <h2>` + message.Title + `</h2>
        <span class="priority `
	
	switch message.Priority {
	case types.PriorityCritical:
		html += `critical">🔴 严重告警`
	case types.PriorityHigh:
		html += `high">🟡 高优先级`
	case types.PriorityMedium:
		html += `medium">🔵 中等优先级`
	case types.PriorityLow:
		html += `low">🟢 低优先级`
	default:
		html += `medium">ℹ️ 通知`
	}
	
	html += `</span>
    </div>
    <div class="content">
        <p>` + strings.ReplaceAll(message.Content, "\n", "<br>") + `</p>
    </div>`
	
	if !message.CreatedAt.IsZero() || message.Type != "" || len(message.Data) > 0 {
		html += `
    <div class="metadata">
        <h4>详细信息</h4>`
		
		if !message.CreatedAt.IsZero() {
			html += `
        <p><strong>时间:</strong> ` + message.CreatedAt.Format("2006-01-02 15:04:05") + `</p>`
		}
		
		if message.Type != "" {
			html += `
        <p><strong>类型:</strong> ` + message.Type + `</p>`
		}
		
		if len(message.Data) > 0 {
			html += `
        <h5>额外数据:</h5>
        <ul>`
			for key, value := range message.Data {
				html += fmt.Sprintf(`
            <li><strong>%s:</strong> %v</li>`, key, value)
			}
			html += `
        </ul>`
		}
		
		html += `
    </div>`
	}
	
	html += `
    <div class="footer">
        <p>此邮件由 AlertAgent 自动发送，请勿回复。</p>
    </div>
</body>
</html>`
	
	return html
}

// formatTextContent 格式化文本内容
func (p *EmailPlugin) formatTextContent(message *types.Message) string {
	content := fmt.Sprintf("标题: %s\n\n", message.Title)
	
	// 添加优先级
	switch message.Priority {
	case types.PriorityCritical:
		content += "优先级: 🔴 严重告警\n\n"
	case types.PriorityHigh:
		content += "优先级: 🟡 高优先级\n\n"
	case types.PriorityMedium:
		content += "优先级: 🔵 中等优先级\n\n"
	case types.PriorityLow:
		content += "优先级: 🟢 低优先级\n\n"
	default:
		content += "优先级: ℹ️ 通知\n\n"
	}
	
	// 添加内容
	content += "内容:\n" + message.Content + "\n\n"
	
	// 添加时间戳
	if !message.CreatedAt.IsZero() {
		content += fmt.Sprintf("时间: %s\n\n", message.CreatedAt.Format("2006-01-02 15:04:05"))
	}
	
	// 添加类型
	if message.Type != "" {
		content += fmt.Sprintf("类型: %s\n\n", message.Type)
	}
	
	// 添加额外数据
	if len(message.Data) > 0 {
		content += "详细信息:\n"
		for key, value := range message.Data {
			content += fmt.Sprintf("  %s: %v\n", key, value)
		}
		content += "\n"
	}
	
	content += "---\n此邮件由 AlertAgent 自动发送，请勿回复。"
	
	return content
}

// sendEmail 发送邮件
func (p *EmailPlugin) sendEmail(config *channel.ChannelConfig, emailContent *EmailContent) error {
	// 获取配置
	smtpHost, _ := config.Settings["smtp_host"].(string)
	smtpPort := config.Settings["smtp_port"]
	username, _ := config.Settings["username"].(string)
	password, _ := config.Settings["password"].(string)
	useTLS := true
	if useTLSInterface, exists := config.Settings["use_tls"]; exists {
		if useTLSBool, ok := useTLSInterface.(bool); ok {
			useTLS = useTLSBool
		}
	}
	
	// 转换端口
	var port int
	switch v := smtpPort.(type) {
	case int:
		port = v
	case float64:
		port = int(v)
	default:
		port = 587
	}
	
	addr := fmt.Sprintf("%s:%d", smtpHost, port)
	
	// 构建邮件头
	headers := make(map[string]string)
	headers["From"] = emailContent.From
	headers["To"] = strings.Join(emailContent.To, ", ")
	if len(emailContent.CC) > 0 {
		headers["Cc"] = strings.Join(emailContent.CC, ", ")
	}
	headers["Subject"] = emailContent.Subject
	headers["MIME-Version"] = "1.0"
	if emailContent.IsHTML {
		headers["Content-Type"] = "text/html; charset=UTF-8"
	} else {
		headers["Content-Type"] = "text/plain; charset=UTF-8"
	}
	
	// 构建邮件消息
	message := ""
	for k, v := range headers {
		message += fmt.Sprintf("%s: %s\r\n", k, v)
	}
	message += "\r\n" + emailContent.Body
	
	// 准备收件人列表
	recipients := make([]string, 0, len(emailContent.To)+len(emailContent.CC))
	recipients = append(recipients, emailContent.To...)
	recipients = append(recipients, emailContent.CC...)
	
	// 初始化认证
	auth := smtp.PlainAuth("", username, password, smtpHost)
	
	if useTLS {
		// 使用TLS连接
		tlsConfig := &tls.Config{
			InsecureSkipVerify: false,
			ServerName:         smtpHost,
		}
		
		conn, err := tls.Dial("tcp", addr, tlsConfig)
		if err != nil {
			return fmt.Errorf("TLS连接失败: %w", err)
		}
		defer conn.Close()
		
		client, err := smtp.NewClient(conn, smtpHost)
		if err != nil {
			return fmt.Errorf("创建SMTP客户端失败: %w", err)
		}
		defer client.Quit()
		
		if err = client.Auth(auth); err != nil {
			return fmt.Errorf("SMTP认证失败: %w", err)
		}
		
		if err = client.Mail(emailContent.From); err != nil {
			return fmt.Errorf("设置发件人失败: %w", err)
		}
		
		for _, recipient := range recipients {
			if err = client.Rcpt(recipient); err != nil {
				return fmt.Errorf("设置收件人失败: %w", err)
			}
		}
		
		w, err := client.Data()
		if err != nil {
			return fmt.Errorf("开始数据传输失败: %w", err)
		}
		
		_, err = w.Write([]byte(message))
		if err != nil {
			return fmt.Errorf("写入邮件数据失败: %w", err)
		}
		
		err = w.Close()
		if err != nil {
			return fmt.Errorf("结束数据传输失败: %w", err)
		}
	} else {
		// 使用普通SMTP连接
		err := smtp.SendMail(addr, auth, emailContent.From, recipients, []byte(message))
		if err != nil {
			return fmt.Errorf("发送邮件失败: %w", err)
		}
	}
	
	return nil
}

// testSMTPConnection 测试SMTP连接
func (p *EmailPlugin) testSMTPConnection(config *channel.ChannelConfig) error {
	// 获取SMTP配置
	smtpHost, _ := config.Settings["smtp_host"].(string)
	smtpPort := config.Settings["smtp_port"]
	username, _ := config.Settings["username"].(string)
	password, _ := config.Settings["password"].(string)
	
	// 转换端口
	var port int
	switch v := smtpPort.(type) {
	case int:
		port = v
	case float64:
		port = int(v)
	default:
		port = 587
	}
	
	addr := fmt.Sprintf("%s:%d", smtpHost, port)
	
	// 测试连接
	auth := smtp.PlainAuth("", username, password, smtpHost)
	
	// 尝试连接并认证
	client, err := smtp.Dial(addr)
	if err != nil {
		return fmt.Errorf("连接SMTP服务器失败: %w", err)
	}
	defer client.Quit()
	
	if err = client.Auth(auth); err != nil {
		return fmt.Errorf("SMTP认证失败: %w", err)
	}
	
	return nil
}