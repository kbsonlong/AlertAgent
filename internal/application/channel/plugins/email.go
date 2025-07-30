package plugins

import (
	"context"
	"crypto/tls"
	"fmt"
	"html/template"
	"net/smtp"
	"strconv"
	"strings"
	"time"
)

// EmailPlugin 邮件通知插件
type EmailPlugin struct{}

// NewEmailPlugin 创建邮件插件实例
func NewEmailPlugin() *EmailPlugin {
	return &EmailPlugin{}
}

// Name 插件名称
func (e *EmailPlugin) Name() string {
	return "email"
}

// Version 插件版本
func (e *EmailPlugin) Version() string {
	return "1.0.0"
}

// Description 插件描述
func (e *EmailPlugin) Description() string {
	return "邮件通知插件，支持SMTP发送和HTML模板渲染"
}

// ConfigSchema 配置Schema
func (e *EmailPlugin) ConfigSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"smtp_host": map[string]interface{}{
				"type":        "string",
				"description": "SMTP服务器地址",
				"required":    true,
			},
			"smtp_port": map[string]interface{}{
				"type":        "integer",
				"description": "SMTP服务器端口",
				"default":     587,
				"required":    true,
			},
			"username": map[string]interface{}{
				"type":        "string",
				"description": "SMTP用户名",
				"required":    true,
			},
			"password": map[string]interface{}{
				"type":        "string",
				"description": "SMTP密码",
				"required":    true,
			},
			"from_email": map[string]interface{}{
				"type":        "string",
				"description": "发件人邮箱地址",
				"required":    true,
			},
			"from_name": map[string]interface{}{
				"type":        "string",
				"description": "发件人姓名",
				"required":    false,
			},
			"to_emails": map[string]interface{}{
				"type":        "array",
				"description": "收件人邮箱地址列表",
				"items": map[string]interface{}{
					"type": "string",
				},
				"required": true,
			},
			"cc_emails": map[string]interface{}{
				"type":        "array",
				"description": "抄送邮箱地址列表",
				"items": map[string]interface{}{
					"type": "string",
				},
				"required": false,
			},
			"use_tls": map[string]interface{}{
				"type":        "boolean",
				"description": "是否使用TLS加密",
				"default":     true,
				"required":    false,
			},
			"template_type": map[string]interface{}{
				"type":        "string",
				"description": "邮件模板类型",
				"enum":        []string{"plain", "html"},
				"default":     "html",
				"required":    false,
			},
		},
		"required": []string{"smtp_host", "smtp_port", "username", "password", "from_email", "to_emails"},
	}
}

// ValidateConfig 验证配置
func (e *EmailPlugin) ValidateConfig(config map[string]interface{}) error {
	// 验证必填字段
	requiredFields := []string{"smtp_host", "smtp_port", "username", "password", "from_email", "to_emails"}
	for _, field := range requiredFields {
		if _, exists := config[field]; !exists {
			return fmt.Errorf("required field '%s' is missing", field)
		}
	}

	// 验证SMTP端口
	if port, ok := config["smtp_port"]; ok {
		if portFloat, ok := port.(float64); ok {
			if portFloat < 1 || portFloat > 65535 {
				return fmt.Errorf("smtp_port must be between 1 and 65535")
			}
		} else {
			return fmt.Errorf("smtp_port must be a number")
		}
	}

	// 验证邮箱地址格式
	fromEmail, _ := config["from_email"].(string)
	if !e.isValidEmail(fromEmail) {
		return fmt.Errorf("invalid from_email format")
	}

	// 验证收件人邮箱
	if toEmails, ok := config["to_emails"].([]interface{}); ok {
		if len(toEmails) == 0 {
			return fmt.Errorf("to_emails cannot be empty")
		}
		for _, email := range toEmails {
			if emailStr, ok := email.(string); ok {
				if !e.isValidEmail(emailStr) {
					return fmt.Errorf("invalid email format: %s", emailStr)
				}
			}
		}
	}

	// 验证抄送邮箱
	if ccEmails, exists := config["cc_emails"]; exists {
		if ccEmailList, ok := ccEmails.([]interface{}); ok {
			for _, email := range ccEmailList {
				if emailStr, ok := email.(string); ok {
					if !e.isValidEmail(emailStr) {
						return fmt.Errorf("invalid cc email format: %s", emailStr)
					}
				}
			}
		}
	}

	// 验证模板类型
	if templateType, exists := config["template_type"]; exists {
		if templateTypeStr, ok := templateType.(string); ok {
			if templateTypeStr != "plain" && templateTypeStr != "html" {
				return fmt.Errorf("template_type must be 'plain' or 'html'")
			}
		}
	}

	return nil
}

// Send 发送通知
func (e *EmailPlugin) Send(ctx context.Context, config map[string]interface{}, message *NotificationMessage) error {
	// 解析配置
	smtpHost := config["smtp_host"].(string)
	smtpPort := int(config["smtp_port"].(float64))
	username := config["username"].(string)
	password := config["password"].(string)
	fromEmail := config["from_email"].(string)
	fromName, _ := config["from_name"].(string)
	toEmails := config["to_emails"].([]interface{})
	ccEmails, _ := config["cc_emails"].([]interface{})
	useTLS, _ := config["use_tls"].(bool)
	templateType, _ := config["template_type"].(string)

	if templateType == "" {
		templateType = "html"
	}

	// 构建邮件内容
	subject := fmt.Sprintf("[%s] %s", e.getSeverityText(message.Severity), message.Title)
	var body string
	var contentType string

	if templateType == "html" {
		body = e.buildHTMLContent(message)
		contentType = "text/html; charset=UTF-8"
	} else {
		body = e.buildPlainContent(message)
		contentType = "text/plain; charset=UTF-8"
	}

	// 构建收件人列表
	var recipients []string
	for _, email := range toEmails {
		recipients = append(recipients, email.(string))
	}
	for _, email := range ccEmails {
		recipients = append(recipients, email.(string))
	}

	// 构建邮件头
	from := fromEmail
	if fromName != "" {
		from = fmt.Sprintf("%s <%s>", fromName, fromEmail)
	}

	headers := make(map[string]string)
	headers["From"] = from
	headers["To"] = strings.Join(e.interfaceSliceToStringSlice(toEmails), ", ")
	if len(ccEmails) > 0 {
		headers["Cc"] = strings.Join(e.interfaceSliceToStringSlice(ccEmails), ", ")
	}
	headers["Subject"] = subject
	headers["Content-Type"] = contentType
	headers["Date"] = time.Now().Format(time.RFC1123Z)
	headers["Message-ID"] = fmt.Sprintf("<%s@%s>", message.AlertID, smtpHost)

	// 构建完整邮件内容
	var emailContent strings.Builder
	for key, value := range headers {
		emailContent.WriteString(fmt.Sprintf("%s: %s\r\n", key, value))
	}
	emailContent.WriteString("\r\n")
	emailContent.WriteString(body)

	// 发送邮件
	return e.sendEmail(smtpHost, smtpPort, username, password, fromEmail, recipients, emailContent.String(), useTLS)
}

// buildHTMLContent 构建HTML邮件内容
func (e *EmailPlugin) buildHTMLContent(message *NotificationMessage) string {
	htmlTemplate := `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>{{.Title}}</title>
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .header { background-color: {{.SeverityColor}}; color: white; padding: 20px; text-align: center; }
        .content { padding: 20px; }
        .info-table { width: 100%; border-collapse: collapse; margin: 20px 0; }
        .info-table th, .info-table td { border: 1px solid #ddd; padding: 12px; text-align: left; }
        .info-table th { background-color: #f2f2f2; }
        .footer { background-color: #f8f9fa; padding: 15px; text-align: center; font-size: 12px; color: #666; }
        .severity-badge { display: inline-block; padding: 4px 8px; border-radius: 4px; color: white; font-weight: bold; }
    </style>
</head>
<body>
    <div class="header">
        <h1>{{.SeverityEmoji}} {{.Title}}</h1>
        <span class="severity-badge" style="background-color: {{.SeverityColor}};">{{.SeverityText}}</span>
    </div>
    
    <div class="content">
        <h2>告警详情</h2>
        <table class="info-table">
            <tr><th>告警ID</th><td>{{.AlertID}}</td></tr>
            <tr><th>告警时间</th><td>{{.Timestamp}}</td></tr>
            <tr><th>告警级别</th><td>{{.SeverityText}}</td></tr>
            <tr><th>告警内容</th><td>{{.Content}}</td></tr>
        </table>
        
        {{if .Labels}}
        <h3>标签信息</h3>
        <table class="info-table">
            {{range $key, $value := .Labels}}
            <tr><th>{{$key}}</th><td>{{$value}}</td></tr>
            {{end}}
        </table>
        {{end}}
        
        {{if .Annotations}}
        <h3>注释信息</h3>
        <table class="info-table">
            {{range $key, $value := .Annotations}}
            <tr><th>{{$key}}</th><td>{{$value}}</td></tr>
            {{end}}
        </table>
        {{end}}
    </div>
    
    <div class="footer">
        <p>此邮件由AlertAgent系统自动发送，请勿回复。</p>
        <p>发送时间: {{.SendTime}}</p>
    </div>
</body>
</html>`

	tmpl, err := template.New("email").Parse(htmlTemplate)
	if err != nil {
		return e.buildPlainContent(message) // 降级到纯文本
	}

	data := map[string]interface{}{
		"Title":         message.Title,
		"AlertID":       message.AlertID,
		"Timestamp":     message.Timestamp.Format("2006-01-02 15:04:05"),
		"Content":       message.Content,
		"Labels":        message.Labels,
		"Annotations":   message.Annotations,
		"SeverityText":  e.getSeverityText(message.Severity),
		"SeverityEmoji": e.getSeverityEmoji(message.Severity),
		"SeverityColor": e.getSeverityColor(message.Severity),
		"SendTime":      time.Now().Format("2006-01-02 15:04:05"),
	}

	var result strings.Builder
	if err := tmpl.Execute(&result, data); err != nil {
		return e.buildPlainContent(message) // 降级到纯文本
	}

	return result.String()
}

// buildPlainContent 构建纯文本邮件内容
func (e *EmailPlugin) buildPlainContent(message *NotificationMessage) string {
	var builder strings.Builder

	builder.WriteString(fmt.Sprintf("告警标题: %s\n", message.Title))
	builder.WriteString(fmt.Sprintf("告警级别: %s\n", e.getSeverityText(message.Severity)))
	builder.WriteString(fmt.Sprintf("告警时间: %s\n", message.Timestamp.Format("2006-01-02 15:04:05")))
	builder.WriteString(fmt.Sprintf("告警ID: %s\n", message.AlertID))
	builder.WriteString(fmt.Sprintf("告警内容: %s\n\n", message.Content))

	if len(message.Labels) > 0 {
		builder.WriteString("标签信息:\n")
		for key, value := range message.Labels {
			builder.WriteString(fmt.Sprintf("  %s: %s\n", key, value))
		}
		builder.WriteString("\n")
	}

	if len(message.Annotations) > 0 {
		builder.WriteString("注释信息:\n")
		for key, value := range message.Annotations {
			builder.WriteString(fmt.Sprintf("  %s: %s\n", key, value))
		}
		builder.WriteString("\n")
	}

	builder.WriteString("---\n")
	builder.WriteString("此邮件由AlertAgent系统自动发送，请勿回复。\n")
	builder.WriteString(fmt.Sprintf("发送时间: %s\n", time.Now().Format("2006-01-02 15:04:05")))

	return builder.String()
}

// sendEmail 发送邮件
func (e *EmailPlugin) sendEmail(host string, port int, username, password, from string, to []string, message string, useTLS bool) error {
	addr := fmt.Sprintf("%s:%d", host, port)

	// 创建认证
	auth := smtp.PlainAuth("", username, password, host)

	if useTLS {
		// 使用TLS连接
		tlsConfig := &tls.Config{
			ServerName: host,
		}

		conn, err := tls.Dial("tcp", addr, tlsConfig)
		if err != nil {
			return fmt.Errorf("failed to connect to SMTP server with TLS: %w", err)
		}
		defer conn.Close()

		client, err := smtp.NewClient(conn, host)
		if err != nil {
			return fmt.Errorf("failed to create SMTP client: %w", err)
		}
		defer client.Quit()

		if err := client.Auth(auth); err != nil {
			return fmt.Errorf("SMTP authentication failed: %w", err)
		}

		if err := client.Mail(from); err != nil {
			return fmt.Errorf("failed to set sender: %w", err)
		}

		for _, recipient := range to {
			if err := client.Rcpt(recipient); err != nil {
				return fmt.Errorf("failed to set recipient %s: %w", recipient, err)
			}
		}

		writer, err := client.Data()
		if err != nil {
			return fmt.Errorf("failed to get data writer: %w", err)
		}

		if _, err := writer.Write([]byte(message)); err != nil {
			return fmt.Errorf("failed to write message: %w", err)
		}

		return writer.Close()
	} else {
		// 使用普通连接
		return smtp.SendMail(addr, auth, from, to, []byte(message))
	}
}

// 辅助方法
func (e *EmailPlugin) getSeverityText(severity string) string {
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

func (e *EmailPlugin) getSeverityEmoji(severity string) string {
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

func (e *EmailPlugin) getSeverityColor(severity string) string {
	switch strings.ToLower(severity) {
	case "critical":
		return "#dc3545"
	case "high":
		return "#fd7e14"
	case "medium":
		return "#ffc107"
	case "low":
		return "#28a745"
	case "info":
		return "#17a2b8"
	case "warning":
		return "#ffc107"
	default:
		return "#6c757d"
	}
}

func (e *EmailPlugin) isValidEmail(email string) bool {
	return strings.Contains(email, "@") && strings.Contains(email, ".")
}

func (e *EmailPlugin) interfaceSliceToStringSlice(slice []interface{}) []string {
	result := make([]string, len(slice))
	for i, v := range slice {
		result[i] = v.(string)
	}
	return result
}

// HealthCheck 健康检查
func (e *EmailPlugin) HealthCheck(ctx context.Context, config map[string]interface{}) error {
	// 创建测试消息
	testMessage := &NotificationMessage{
		Title:     "健康检查",
		Content:   "这是一条测试消息，用于验证邮件通知插件配置是否正确。",
		Severity:  "info",
		AlertID:   fmt.Sprintf("health-check-%d", time.Now().Unix()),
		Timestamp: time.Now(),
		Labels: map[string]string{
			"test": "true",
			"type": "health_check",
		},
		Annotations: map[string]string{
			"description": "邮件插件健康检查消息",
		},
	}

	return e.Send(ctx, config, testMessage)
}

// Initialize 初始化插件
func (e *EmailPlugin) Initialize() error {
	// 邮件插件无需特殊初始化
	return nil
}

// Shutdown 关闭插件
func (e *EmailPlugin) Shutdown() error {
	// 邮件插件无需特殊清理
	return nil
}