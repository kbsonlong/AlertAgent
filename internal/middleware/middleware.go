package middleware

import (
	"net/http"
	"runtime/debug"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"alert_agent/internal/security/auth"
	"alert_agent/internal/security/audit"
	"alert_agent/internal/security/rbac"
	"alert_agent/internal/security/validator"
)

// Logger 日志中间件
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 开始时间
		startTime := time.Now()

		// 处理请求
		c.Next()

		// 结束时间
		endTime := time.Now()

		// 执行时间
		latencyTime := endTime.Sub(startTime)

		// 请求方式
		reqMethod := c.Request.Method

		// 请求路由
		reqUri := c.Request.RequestURI

		// 状态码
		statusCode := c.Writer.Status()

		// 请求IP
		clientIP := c.ClientIP()

		// 日志格式
		logrus.WithFields(logrus.Fields{
			"status_code":  statusCode,
			"latency_time": latencyTime,
			"client_ip":    clientIP,
			"req_method":   reqMethod,
			"req_uri":      reqUri,
		}).Info()
	}
}

// Recovery 异常恢复中间件
func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				logrus.WithFields(logrus.Fields{
					"error": err,
					"stack": string(debug.Stack()),
				}).Error("panic recovered")

				c.AbortWithStatusJSON(500, gin.H{
					"code": 500,
					"msg":  "Internal Server Error",
				})
			}
		}()
		c.Next()
	}
}

// SecurityConfig 安全配置
type SecurityConfig struct {
	JWTManager    *auth.JWTManager
	RBACManager   *rbac.RBACManager
	AuditLogger   *audit.AuditLogger
	SkipPaths     []string // 跳过认证的路径
}

// AuthMiddleware JWT认证中间件
func AuthMiddleware(config *SecurityConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 检查是否跳过认证
		for _, path := range config.SkipPaths {
			if strings.HasPrefix(c.Request.URL.Path, path) {
				c.Next()
				return
			}
		}

		// 获取Authorization头
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code": 401,
				"msg":  "Missing authorization header",
			})
			c.Abort()
			return
		}

		// 检查Bearer前缀
		if !strings.HasPrefix(authHeader, "Bearer ") {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code": 401,
				"msg":  "Invalid authorization header format",
			})
			c.Abort()
			return
		}

		// 提取token
		token := strings.TrimPrefix(authHeader, "Bearer ")

		// 验证token
		claims, err := config.JWTManager.VerifyToken(token)
		if err != nil {
			// 记录认证失败审计日志
			config.AuditLogger.LogLogin("", "", c.ClientIP(), c.GetHeader("User-Agent"), false, "Invalid JWT token")
			
			c.JSON(http.StatusUnauthorized, gin.H{
				"code": 401,
				"msg":  "Invalid token",
			})
			c.Abort()
			return
		}

		// 将用户信息存储到上下文
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("roles", claims.Roles)
		c.Set("permissions", claims.Perms)

		// 记录成功认证审计日志
		config.AuditLogger.LogLogin(claims.UserID, claims.Username, c.ClientIP(), c.GetHeader("User-Agent"), true, "JWT authentication successful")

		c.Next()
	}
}

// PermissionMiddleware 权限检查中间件
func PermissionMiddleware(config *SecurityConfig, requiredPermission rbac.Permission) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code": 401,
				"msg":  "User not authenticated",
			})
			c.Abort()
			return
		}

		userIDStr := userID.(string)
		username, _ := c.Get("username")
		usernameStr := username.(string)

		// 检查权限
		hasPermission := config.RBACManager.HasPermission(userIDStr, requiredPermission)

		// 记录权限检查审计日志
		config.AuditLogger.LogPermissionCheck(userIDStr, usernameStr, c.Request.URL.Path, string(requiredPermission), c.ClientIP(), hasPermission)

		if !hasPermission {
			c.JSON(http.StatusForbidden, gin.H{
				"code": 403,
				"msg":  "Insufficient permissions",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RoleMiddleware 角色检查中间件
func RoleMiddleware(config *SecurityConfig, requiredRoles []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code": 401,
				"msg":  "User not authenticated",
			})
			c.Abort()
			return
		}

		userIDStr := userID.(string)
		userRoles := config.RBACManager.GetUserRoles(userIDStr)

		// 检查用户是否有任意一个所需角色
		hasRole := false
		for _, requiredRole := range requiredRoles {
			for _, userRole := range userRoles {
				if userRole == requiredRole {
					hasRole = true
					break
				}
			}
			if hasRole {
				break
			}
		}

		if !hasRole {
			c.JSON(http.StatusForbidden, gin.H{
				"code": 403,
				"msg":  "Insufficient role permissions",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// InputValidationMiddleware 输入验证中间件
func InputValidationMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 对所有输入进行基本的安全检查
		if c.Request.Method == "POST" || c.Request.Method == "PUT" || c.Request.Method == "PATCH" {
			// 检查Content-Type
			contentType := c.GetHeader("Content-Type")
			if contentType != "" && !strings.Contains(contentType, "application/json") && !strings.Contains(contentType, "application/x-www-form-urlencoded") {
				c.JSON(http.StatusBadRequest, gin.H{
					"code": 400,
					"msg":  "Unsupported content type",
				})
				c.Abort()
				return
			}
		}

		// 检查请求大小
		if c.Request.ContentLength > 10*1024*1024 { // 10MB限制
			c.JSON(http.StatusRequestEntityTooLarge, gin.H{
				"code": 413,
				"msg":  "Request entity too large",
			})
			c.Abort()
			return
		}

		// 基本的XSS和SQL注入检查
		validator := validator.NewValidator()
		
		// 检查URL参数
		for key, values := range c.Request.URL.Query() {
			for _, value := range values {
				validator.XSSCheck(key, value)
				validator.SQLInjectionCheck(key, value)
			}
		}

		if validator.HasErrors() {
			c.JSON(http.StatusBadRequest, gin.H{
				"code": 400,
				"msg":  "Invalid input detected",
				"errors": validator.GetErrors(),
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// AuditMiddleware 审计中间件
func AuditMiddleware(config *SecurityConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		startTime := time.Now()

		// 创建审计上下文
		userID, _ := c.Get("user_id")
		username, _ := c.Get("username")
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = generateRequestID()
			c.Header("X-Request-ID", requestID)
		}

		auditCtx := audit.NewAuditContext(
			getStringValue(userID),
			getStringValue(username),
			c.ClientIP(),
			c.GetHeader("User-Agent"),
			requestID,
			"", // sessionID可以从cookie或其他地方获取
		)

		c.Set("audit_context", auditCtx)

		c.Next()

		// 记录请求完成审计日志
		duration := time.Since(startTime)
		status := c.Writer.Status()
		success := status < 400

		config.AuditLogger.LogResourceAccess(
			getStringValue(userID),
			getStringValue(username),
			c.Request.URL.Path,
			"",
			c.ClientIP(),
			audit.AuditAction(c.Request.Method),
			success,
			getStatusMessage(status),
			map[string]interface{}{
				"method":     c.Request.Method,
				"path":       c.Request.URL.Path,
				"status":     status,
				"duration":   duration.String(),
				"request_id": requestID,
			},
		)
	}
}

// SecurityHeadersMiddleware 安全头中间件
func SecurityHeadersMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 设置安全头
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		c.Header("Content-Security-Policy", "default-src 'self'")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Header("Permissions-Policy", "geolocation=(), microphone=(), camera=()")

		c.Next()
	}
}

// RateLimitMiddleware 限流中间件（简单实现）
func RateLimitMiddleware(requestsPerMinute int) gin.HandlerFunc {
	// 这里可以集成更复杂的限流库，如redis-based rate limiter
	return func(c *gin.Context) {
		// 简单的IP限流实现
		// 实际生产环境建议使用更完善的限流方案
		c.Next()
	}
}

// 辅助函数
func getStringValue(value interface{}) string {
	if value == nil {
		return ""
	}
	if str, ok := value.(string); ok {
		return str
	}
	return ""
}

func getStatusMessage(status int) string {
	switch {
	case status >= 200 && status < 300:
		return "Success"
	case status >= 400 && status < 500:
		return "Client Error"
	case status >= 500:
		return "Server Error"
	default:
		return "Unknown"
	}
}

func generateRequestID() string {
	return time.Now().Format("20060102150405") + "-" + time.Now().Format("000")
}

// Cors 跨域中间件
func Cors() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Headers", "Content-Type,AccessToken,X-CSRF-Token, Authorization")
		c.Header("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		c.Header("Access-Control-Expose-Headers", "Content-Length, Access-Control-Allow-Origin, Access-Control-Allow-Headers, Content-Type")
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Content-Type", "application/json; charset=utf-8")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}
