package audit

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
)

// AuditLevel 审计级别
type AuditLevel string

const (
	AuditLevelInfo    AuditLevel = "INFO"
	AuditLevelWarning AuditLevel = "WARNING"
	AuditLevelError   AuditLevel = "ERROR"
	AuditLevelCritical AuditLevel = "CRITICAL"
)

// AuditAction 审计动作
type AuditAction string

const (
	// 认证相关
	ActionLogin       AuditAction = "LOGIN"
	ActionLogout      AuditAction = "LOGOUT"
	ActionLoginFailed AuditAction = "LOGIN_FAILED"

	// 用户管理
	ActionUserCreate AuditAction = "USER_CREATE"
	ActionUserUpdate AuditAction = "USER_UPDATE"
	ActionUserDelete AuditAction = "USER_DELETE"
	ActionUserView   AuditAction = "USER_VIEW"

	// 角色权限
	ActionRoleAssign AuditAction = "ROLE_ASSIGN"
	ActionRoleRevoke AuditAction = "ROLE_REVOKE"
	ActionPermCheck  AuditAction = "PERMISSION_CHECK"

	// 告警管理
	ActionAlertCreate AuditAction = "ALERT_CREATE"
	ActionAlertUpdate AuditAction = "ALERT_UPDATE"
	ActionAlertDelete AuditAction = "ALERT_DELETE"
	ActionAlertHandle AuditAction = "ALERT_HANDLE"
	ActionAlertView   AuditAction = "ALERT_VIEW"

	// 规则管理
	ActionRuleCreate AuditAction = "RULE_CREATE"
	ActionRuleUpdate AuditAction = "RULE_UPDATE"
	ActionRuleDelete AuditAction = "RULE_DELETE"
	ActionRuleView   AuditAction = "RULE_VIEW"

	// 集群管理
	ActionClusterCreate AuditAction = "CLUSTER_CREATE"
	ActionClusterUpdate AuditAction = "CLUSTER_UPDATE"
	ActionClusterDelete AuditAction = "CLUSTER_DELETE"
	ActionClusterView   AuditAction = "CLUSTER_VIEW"

	// 系统配置
	ActionConfigUpdate AuditAction = "CONFIG_UPDATE"
	ActionConfigView   AuditAction = "CONFIG_VIEW"

	// 数据访问
	ActionDataExport AuditAction = "DATA_EXPORT"
	ActionDataImport AuditAction = "DATA_IMPORT"
	ActionDataBackup AuditAction = "DATA_BACKUP"
)

// AuditEvent 审计事件
type AuditEvent struct {
	ID          string                 `json:"id"`
	Timestamp   time.Time              `json:"timestamp"`
	Level       AuditLevel             `json:"level"`
	Action      AuditAction            `json:"action"`
	UserID      string                 `json:"user_id"`
	Username    string                 `json:"username"`
	Resource    string                 `json:"resource"`
	ResourceID  string                 `json:"resource_id"`
	IPAddress   string                 `json:"ip_address"`
	UserAgent   string                 `json:"user_agent"`
	Success     bool                   `json:"success"`
	Message     string                 `json:"message"`
	Details     map[string]interface{} `json:"details"`
	RequestID   string                 `json:"request_id"`
	SessionID   string                 `json:"session_id"`
	Duration    time.Duration          `json:"duration"`
	ErrorCode   string                 `json:"error_code,omitempty"`
	ErrorMsg    string                 `json:"error_msg,omitempty"`
}

// AuditLogger 审计日志记录器
type AuditLogger struct {
	logger   *logrus.Logger
	enabled  bool
	filePath string
}

// NewAuditLogger 创建审计日志记录器
func NewAuditLogger(filePath string, enabled bool) (*AuditLogger, error) {
	logger := logrus.New()
	
	// 设置JSON格式
	logger.SetFormatter(&logrus.JSONFormatter{
		TimestampFormat: time.RFC3339,
	})

	// 如果指定了文件路径，则写入文件
	if filePath != "" {
		// 这里可以添加文件输出配置
		// 为了简化，暂时使用标准输出
	}

	return &AuditLogger{
		logger:   logger,
		enabled:  enabled,
		filePath: filePath,
	}, nil
}

// Log 记录审计日志
func (al *AuditLogger) Log(event *AuditEvent) {
	if !al.enabled {
		return
	}

	// 生成事件ID
	if event.ID == "" {
		event.ID = generateEventID()
	}

	// 设置时间戳
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}

	// 序列化事件
	eventJSON, err := json.Marshal(event)
	if err != nil {
		al.logger.WithError(err).Error("Failed to marshal audit event")
		return
	}

	// 根据级别记录日志
	switch event.Level {
	case AuditLevelInfo:
		al.logger.Info(string(eventJSON))
	case AuditLevelWarning:
		al.logger.Warn(string(eventJSON))
	case AuditLevelError:
		al.logger.Error(string(eventJSON))
	case AuditLevelCritical:
		al.logger.Error(string(eventJSON)) // logrus没有Critical级别，使用Error
	default:
		al.logger.Info(string(eventJSON))
	}
}

// LogLogin 记录登录事件
func (al *AuditLogger) LogLogin(userID, username, ipAddress, userAgent string, success bool, message string) {
	level := AuditLevelInfo
	action := ActionLogin
	if !success {
		level = AuditLevelWarning
		action = ActionLoginFailed
	}

	event := &AuditEvent{
		Level:     level,
		Action:    action,
		UserID:    userID,
		Username:  username,
		IPAddress: ipAddress,
		UserAgent: userAgent,
		Success:   success,
		Message:   message,
	}

	al.Log(event)
}

// LogLogout 记录登出事件
func (al *AuditLogger) LogLogout(userID, username, ipAddress string) {
	event := &AuditEvent{
		Level:     AuditLevelInfo,
		Action:    ActionLogout,
		UserID:    userID,
		Username:  username,
		IPAddress: ipAddress,
		Success:   true,
		Message:   "User logged out",
	}

	al.Log(event)
}

// LogResourceAccess 记录资源访问事件
func (al *AuditLogger) LogResourceAccess(userID, username, resource, resourceID, ipAddress string, action AuditAction, success bool, message string, details map[string]interface{}) {
	level := AuditLevelInfo
	if !success {
		level = AuditLevelWarning
	}

	event := &AuditEvent{
		Level:      level,
		Action:     action,
		UserID:     userID,
		Username:   username,
		Resource:   resource,
		ResourceID: resourceID,
		IPAddress:  ipAddress,
		Success:    success,
		Message:    message,
		Details:    details,
	}

	al.Log(event)
}

// LogPermissionCheck 记录权限检查事件
func (al *AuditLogger) LogPermissionCheck(userID, username, resource, permission, ipAddress string, granted bool) {
	level := AuditLevelInfo
	if !granted {
		level = AuditLevelWarning
	}

	message := fmt.Sprintf("Permission check: %s on %s", permission, resource)
	if granted {
		message += " - GRANTED"
	} else {
		message += " - DENIED"
	}

	event := &AuditEvent{
		Level:     level,
		Action:    ActionPermCheck,
		UserID:    userID,
		Username:  username,
		Resource:  resource,
		IPAddress: ipAddress,
		Success:   granted,
		Message:   message,
		Details: map[string]interface{}{
			"permission": permission,
			"granted":    granted,
		},
	}

	al.Log(event)
}

// LogSecurityEvent 记录安全事件
func (al *AuditLogger) LogSecurityEvent(userID, username, ipAddress, eventType, message string, details map[string]interface{}) {
	event := &AuditEvent{
		Level:     AuditLevelCritical,
		Action:    AuditAction(eventType),
		UserID:    userID,
		Username:  username,
		IPAddress: ipAddress,
		Success:   false,
		Message:   message,
		Details:   details,
	}

	al.Log(event)
}

// LogDataOperation 记录数据操作事件
func (al *AuditLogger) LogDataOperation(userID, username, operation, tableName, recordID, ipAddress string, success bool, changes map[string]interface{}) {
	level := AuditLevelInfo
	if !success {
		level = AuditLevelError
	}

	event := &AuditEvent{
		Level:      level,
		Action:     AuditAction(operation),
		UserID:     userID,
		Username:   username,
		Resource:   tableName,
		ResourceID: recordID,
		IPAddress:  ipAddress,
		Success:    success,
		Message:    fmt.Sprintf("%s operation on %s", operation, tableName),
		Details: map[string]interface{}{
			"changes": changes,
		},
	}

	al.Log(event)
}

// AuditContext 审计上下文
type AuditContext struct {
	UserID    string
	Username  string
	IPAddress string
	UserAgent string
	RequestID string
	SessionID string
	StartTime time.Time
}

// NewAuditContext 创建审计上下文
func NewAuditContext(userID, username, ipAddress, userAgent, requestID, sessionID string) *AuditContext {
	return &AuditContext{
		UserID:    userID,
		Username:  username,
		IPAddress: ipAddress,
		UserAgent: userAgent,
		RequestID: requestID,
		SessionID: sessionID,
		StartTime: time.Now(),
	}
}

// LogWithContext 使用上下文记录审计日志
func (al *AuditLogger) LogWithContext(ctx *AuditContext, action AuditAction, resource, resourceID string, success bool, message string, details map[string]interface{}) {
	level := AuditLevelInfo
	if !success {
		level = AuditLevelWarning
	}

	event := &AuditEvent{
		Level:      level,
		Action:     action,
		UserID:     ctx.UserID,
		Username:   ctx.Username,
		Resource:   resource,
		ResourceID: resourceID,
		IPAddress:  ctx.IPAddress,
		UserAgent:  ctx.UserAgent,
		RequestID:  ctx.RequestID,
		SessionID:  ctx.SessionID,
		Success:    success,
		Message:    message,
		Details:    details,
		Duration:   time.Since(ctx.StartTime),
	}

	al.Log(event)
}

// generateEventID 生成事件ID
func generateEventID() string {
	return fmt.Sprintf("audit_%d", time.Now().UnixNano())
}

// AuditMiddleware 审计中间件接口
type AuditMiddleware interface {
	AuditRequest(ctx context.Context, auditCtx *AuditContext, action AuditAction, resource string) func(success bool, message string, details map[string]interface{})
}

// DefaultAuditMiddleware 默认审计中间件
type DefaultAuditMiddleware struct {
	logger *AuditLogger
}

// NewDefaultAuditMiddleware 创建默认审计中间件
func NewDefaultAuditMiddleware(logger *AuditLogger) *DefaultAuditMiddleware {
	return &DefaultAuditMiddleware{
		logger: logger,
	}
}

// AuditRequest 审计请求
func (dam *DefaultAuditMiddleware) AuditRequest(ctx context.Context, auditCtx *AuditContext, action AuditAction, resource string) func(success bool, message string, details map[string]interface{}) {
	return func(success bool, message string, details map[string]interface{}) {
		dam.logger.LogWithContext(auditCtx, action, resource, "", success, message, details)
	}
}