package observability

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// ObservabilityMiddleware 观测性中间件
type ObservabilityMiddleware struct {
	manager *Manager
}

// NewObservabilityMiddleware 创建观测性中间件
func NewObservabilityMiddleware(manager *Manager) *ObservabilityMiddleware {
	return &ObservabilityMiddleware{
		manager: manager,
	}
}

// HTTPMetricsMiddleware HTTP指标中间件
func (om *ObservabilityMiddleware) HTTPMetricsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		
		// 处理请求
		c.Next()
		
		// 记录指标
		duration := time.Since(start)
		om.manager.RecordHTTPRequest(
			c.Request.Method,
			c.FullPath(),
			c.Writer.Status(),
			duration,
		)
	}
}

// HTTPTracingMiddleware HTTP追踪中间件
func (om *ObservabilityMiddleware) HTTPTracingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 开始追踪
		ctx, span := om.manager.TraceHTTPRequest(
			c.Request.Context(),
			c.Request.Method,
			c.Request.URL.String(),
			c.Request.UserAgent(),
		)
		
		// 设置上下文
		c.Request = c.Request.WithContext(ctx)
		
		// 处理请求
		c.Next()
		
		// 结束追踪
		if span != nil {
			span.SetTag("http.status_code", strconv.Itoa(c.Writer.Status()))
			span.SetTag("http.response_size", strconv.Itoa(c.Writer.Size()))
			
			if c.Writer.Status() >= 400 {
				span.LogFields(map[string]interface{}{"error": true, "message": "HTTP error response"})
			} else {
				span.LogFields(map[string]interface{}{"error": false})
			}
			
			span.Finish()
		}
	}
}

// HTTPLoggingMiddleware HTTP日志中间件
func (om *ObservabilityMiddleware) HTTPLoggingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		
		// 处理请求
		c.Next()
		
		// 记录日志
		duration := time.Since(start)
		logger := om.manager.LogWithContext(c.Request.Context())
		
		if logger != nil {
			fields := map[string]interface{}{
				"method":        c.Request.Method,
				"path":          c.Request.URL.Path,
				"status":        c.Writer.Status(),
				"duration":      duration.Milliseconds(),
				"user_agent":    c.Request.UserAgent(),
				"remote_addr":   c.ClientIP(),
				"request_size":  c.Request.ContentLength,
				"response_size": c.Writer.Size(),
			}
			
			zapFields := mapToZapFields(fields)
			if c.Writer.Status() >= 500 {
				logger.Error("HTTP request completed with server error", zapFields...)
			} else if c.Writer.Status() >= 400 {
				logger.Warn("HTTP request completed with client error", zapFields...)
			} else {
				logger.Info("HTTP request completed", zapFields...)
			}
		}
	}
}

// ErrorHandlingMiddleware 错误处理中间件
func (om *ObservabilityMiddleware) ErrorHandlingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 处理请求
		c.Next()
		
		// 处理错误
		if len(c.Errors) > 0 {
			for _, err := range c.Errors {
				om.manager.LogError(
					c.Request.Context(),
					err.Err,
					"http_request",
					map[string]interface{}{
						"method": c.Request.Method,
						"path":   c.Request.URL.Path,
						"status": c.Writer.Status(),
					},
				)
			}
		}
	}
}

// RecoveryMiddleware 恢复中间件
func (om *ObservabilityMiddleware) RecoveryMiddleware() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		// 记录panic日志
		om.manager.LogError(
			c.Request.Context(),
			gin.Error{Err: gin.Error{Err: recovered.(error)}}.Err,
			"panic_recovery",
			map[string]interface{}{
				"method": c.Request.Method,
				"path":   c.Request.URL.Path,
				"panic":  recovered,
			},
		)
		
		// 返回500错误
		c.AbortWithStatus(500)
	})
}

// HealthCheckHandler 健康检查处理器
func (om *ObservabilityMiddleware) HealthCheckHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		report := om.manager.CheckHealth(c.Request.Context())
		if report == nil {
			c.JSON(500, gin.H{"error": "health check unavailable"})
			return
		}
		
		if report.Status == "healthy" {
			c.JSON(200, report)
		} else {
			c.JSON(503, report)
		}
	}
}

// ReadinessCheckHandler 就绪检查处理器
func (om *ObservabilityMiddleware) ReadinessCheckHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		report := om.manager.CheckReadiness(c.Request.Context())
		if report == nil {
			c.JSON(500, gin.H{"error": "readiness check unavailable"})
			return
		}
		
		if report.Ready {
			c.JSON(200, report)
		} else {
			c.JSON(503, report)
		}
	}
}

// MetricsHandler 指标处理器
func (om *ObservabilityMiddleware) MetricsHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		if om.manager.GetMetricsManager() != nil {
			registry := om.manager.GetMetricsManager().GetRegistry()
			if registry != nil {
				// 这里可以添加自定义的指标输出逻辑
				c.JSON(200, gin.H{"message": "metrics available at /metrics endpoint"})
				return
			}
		}
		c.JSON(500, gin.H{"error": "metrics unavailable"})
	}
}

// BusinessEventMiddleware 业务事件记录中间件
func (om *ObservabilityMiddleware) BusinessEventMiddleware(eventName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		
		// 处理请求
		c.Next()
		
		// 记录业务事件
		duration := time.Since(start)
		om.manager.LogBusinessEvent(
			c.Request.Context(),
			eventName,
			map[string]interface{}{
				"method":   c.Request.Method,
				"path":     c.Request.URL.Path,
				"status":   c.Writer.Status(),
				"duration": duration.Milliseconds(),
				"success":  c.Writer.Status() < 400,
			},
		)
	}
}

// SecurityEventMiddleware 安全事件记录中间件
func (om *ObservabilityMiddleware) SecurityEventMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 检查可疑活动
		if om.isSuspiciousRequest(c) {
			om.manager.LogSecurityEvent(
				c.Request.Context(),
				"suspicious_request",
				"warning",
				map[string]interface{}{
					"method":      c.Request.Method,
					"path":        c.Request.URL.Path,
					"remote_addr": c.ClientIP(),
					"user_agent":  c.Request.UserAgent(),
					"headers":     c.Request.Header,
				},
			)
		}
		
		// 处理请求
		c.Next()
		
		// 检查认证失败
		if c.Writer.Status() == 401 || c.Writer.Status() == 403 {
			om.manager.LogSecurityEvent(
				c.Request.Context(),
				"authentication_failure",
				"high",
				map[string]interface{}{
					"method":      c.Request.Method,
					"path":        c.Request.URL.Path,
					"status":      c.Writer.Status(),
					"remote_addr": c.ClientIP(),
					"user_agent":  c.Request.UserAgent(),
				},
			)
		}
	}
}

// isSuspiciousRequest 检查是否为可疑请求
func (om *ObservabilityMiddleware) isSuspiciousRequest(c *gin.Context) bool {
	// 检查SQL注入模式
	path := c.Request.URL.Path
	query := c.Request.URL.RawQuery
	
	suspiciousPatterns := []string{
		"union", "select", "insert", "update", "delete", "drop",
		"script", "javascript", "vbscript", "onload", "onerror",
		"../", "..\\", "/etc/passwd", "/etc/shadow",
	}
	
	for _, pattern := range suspiciousPatterns {
		if contains(path, pattern) || contains(query, pattern) {
			return true
		}
	}
	
	return false
}

// contains 检查字符串是否包含子字符串（不区分大小写）
func contains(s, substr string) bool {
	return len(s) >= len(substr) && 
		   (s == substr || 
			len(s) > len(substr) && 
			(s[:len(substr)] == substr || 
			 s[len(s)-len(substr):] == substr ||
			 indexOfSubstring(s, substr) >= 0))
}

// indexOfSubstring 查找子字符串位置
func indexOfSubstring(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

// SetupObservabilityRoutes 设置观测性路由
func (om *ObservabilityMiddleware) SetupObservabilityRoutes(router *gin.Engine) {
	// 健康检查路由
	router.GET("/health", om.HealthCheckHandler())
	router.GET("/ready", om.ReadinessCheckHandler())
	
	// 指标路由（如果需要自定义处理）
	router.GET("/metrics-info", om.MetricsHandler())
}

// mapToZapFields 将map转换为zap.Field切片
func mapToZapFields(fields map[string]interface{}) []zap.Field {
	zapFields := make([]zap.Field, 0, len(fields))
	for key, value := range fields {
		switch v := value.(type) {
		case string:
			zapFields = append(zapFields, zap.String(key, v))
		case int:
			zapFields = append(zapFields, zap.Int(key, v))
		case int64:
			zapFields = append(zapFields, zap.Int64(key, v))
		case float64:
			zapFields = append(zapFields, zap.Float64(key, v))
		case bool:
			zapFields = append(zapFields, zap.Bool(key, v))
		case time.Duration:
			zapFields = append(zapFields, zap.Duration(key, v))
		default:
			zapFields = append(zapFields, zap.Any(key, v))
		}
	}
	return zapFields
}

// ApplyAllMiddleware 应用所有观测性中间件
func (om *ObservabilityMiddleware) ApplyAllMiddleware(router *gin.Engine) {
	// 恢复中间件（最先执行）
	router.Use(om.RecoveryMiddleware())
	
	// 日志中间件
	router.Use(om.HTTPLoggingMiddleware())
	
	// 指标中间件
	router.Use(om.HTTPMetricsMiddleware())
	
	// 追踪中间件
	router.Use(om.HTTPTracingMiddleware())
	
	// 错误处理中间件
	router.Use(om.ErrorHandlingMiddleware())
	
	// 安全事件中间件
	router.Use(om.SecurityEventMiddleware())
	
	// 设置观测性路由
	om.SetupObservabilityRoutes(router)
}