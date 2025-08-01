package middleware

import (
	"alert_agent/internal/config"
	"alert_agent/internal/pkg/logger"
	"fmt"
	"net/http"
	"runtime/debug"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"golang.org/x/time/rate"
)

// StandardResponse 标准响应格式
type StandardResponse struct {
	Code      int         `json:"code"`
	Message   string      `json:"message"`
	Data      interface{} `json:"data,omitempty"`
	RequestID string      `json:"request_id,omitempty"`
	Timestamp int64       `json:"timestamp"`
}

// ErrorResponse 错误响应格式
type ErrorResponse struct {
	Code      int         `json:"code"`
	Message   string      `json:"message"`
	Error     string      `json:"error,omitempty"`
	RequestID string      `json:"request_id,omitempty"`
	Timestamp int64       `json:"timestamp"`
}

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

		// 请求ID
		requestID := c.GetString("request_id")

		// 日志记录
		logger.L.Info("HTTP Request",
			zap.String("method", reqMethod),
			zap.String("uri", reqUri),
			zap.String("client_ip", clientIP),
			zap.Int("status_code", statusCode),
			zap.Duration("latency", latencyTime),
			zap.String("request_id", requestID),
		)
	}
}

// Recovery 异常恢复中间件
func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				requestID := c.GetString("request_id")
				
				logger.L.Error("Panic recovered",
					zap.Any("error", err),
					zap.String("stack", string(debug.Stack())),
					zap.String("request_id", requestID),
				)

				c.AbortWithStatusJSON(http.StatusInternalServerError, ErrorResponse{
					Code:      http.StatusInternalServerError,
					Message:   "Internal Server Error",
					Error:     "An unexpected error occurred",
					RequestID: requestID,
					Timestamp: time.Now().Unix(),
				})
			}
		}()
		c.Next()
	}
}

// Cors 跨域中间件
func Cors() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Headers", "Content-Type,AccessToken,X-CSRF-Token, Authorization, X-Request-ID")
		c.Header("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		c.Header("Access-Control-Expose-Headers", "Content-Length, Access-Control-Allow-Origin, Access-Control-Allow-Headers, Content-Type, X-Request-ID")
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Content-Type", "application/json; charset=utf-8")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

// RequestID 请求ID中间件
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
		}
		
		c.Set("request_id", requestID)
		c.Header("X-Request-ID", requestID)
		c.Next()
	}
}

// ResponseFormatter 响应格式化中间件
func ResponseFormatter() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		
		// 如果已经有响应内容，不再处理
		if c.Writer.Written() {
			return
		}
		
		// 检查是否有错误
		if len(c.Errors) > 0 {
			err := c.Errors.Last()
			requestID := c.GetString("request_id")
			
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Code:      http.StatusInternalServerError,
				Message:   "Request failed",
				Error:     err.Error(),
				RequestID: requestID,
				Timestamp: time.Now().Unix(),
			})
		}
	}
}

// RateLimiter 限流器结构
type RateLimiter struct {
	limiters map[string]*rate.Limiter
	mutex    sync.RWMutex
	rate     rate.Limit
	burst    int
}

// NewRateLimiter 创建新的限流器
func NewRateLimiter(r rate.Limit, b int) *RateLimiter {
	return &RateLimiter{
		limiters: make(map[string]*rate.Limiter),
		rate:     r,
		burst:    b,
	}
}

// GetLimiter 获取指定IP的限流器
func (rl *RateLimiter) GetLimiter(ip string) *rate.Limiter {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()
	
	limiter, exists := rl.limiters[ip]
	if !exists {
		limiter = rate.NewLimiter(rl.rate, rl.burst)
		rl.limiters[ip] = limiter
	}
	
	return limiter
}

// 全局限流器实例
var globalRateLimiter = NewRateLimiter(rate.Every(time.Second), 100)

// RateLimit 限流中间件
func RateLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		limiter := globalRateLimiter.GetLimiter(ip)
		
		if !limiter.Allow() {
			requestID := c.GetString("request_id")
			
			c.AbortWithStatusJSON(http.StatusTooManyRequests, ErrorResponse{
				Code:      http.StatusTooManyRequests,
				Message:   "Rate limit exceeded",
				Error:     "Too many requests from this IP",
				RequestID: requestID,
				Timestamp: time.Now().Unix(),
			})
			return
		}
		
		c.Next()
	}
}

// JWTClaims JWT声明结构
type JWTClaims struct {
	UserID   string   `json:"user_id"`
	Username string   `json:"username"`
	Roles    []string `json:"roles"`
	jwt.RegisteredClaims
}

// JWTAuth JWT认证中间件
func JWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		requestID := c.GetString("request_id")
		
		// 添加调试日志
		logger.L.Debug("JWT Auth Debug",
			zap.String("auth_header", authHeader),
			zap.String("request_id", requestID),
		)
		
		if authHeader == "" {
			logger.L.Debug("Missing authorization header", zap.String("request_id", requestID))
			
			c.AbortWithStatusJSON(http.StatusUnauthorized, ErrorResponse{
				Code:      http.StatusUnauthorized,
				Message:   "Authorization header required",
				Error:     "Missing authorization token",
				RequestID: requestID,
				Timestamp: time.Now().Unix(),
			})
			return
		}

		// 检查Bearer前缀
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			logger.L.Debug("Invalid authorization format", 
				zap.String("auth_header", authHeader),
				zap.String("request_id", requestID),
			)
			
			c.AbortWithStatusJSON(http.StatusUnauthorized, ErrorResponse{
				Code:      http.StatusUnauthorized,
				Message:   "Invalid authorization format",
				Error:     "Authorization header must start with 'Bearer '",
				RequestID: requestID,
				Timestamp: time.Now().Unix(),
			})
			return
		}

		// 添加token调试日志
		logger.L.Debug("Parsing JWT token",
			zap.String("token_string", tokenString),
			zap.String("request_id", requestID),
		)

		// 解析JWT token
		cfg := config.GetConfig()
		token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
			// 添加签名方法调试日志
			logger.L.Debug("JWT signing method check",
				zap.Any("method", token.Method),
				zap.Any("header", token.Header),
				zap.String("request_id", requestID),
			)
			
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(cfg.Server.JWTSecret), nil
		})

		if err != nil {
			logger.L.Debug("JWT parsing failed",
				zap.Error(err),
				zap.String("token_string", tokenString),
				zap.String("request_id", requestID),
			)
			
			c.AbortWithStatusJSON(http.StatusUnauthorized, ErrorResponse{
				Code:      http.StatusUnauthorized,
				Message:   "Invalid token",
				Error:     err.Error(),
				RequestID: requestID,
				Timestamp: time.Now().Unix(),
			})
			return
		}

		// 添加claims调试日志
		logger.L.Debug("JWT token parsed successfully",
			zap.Bool("token_valid", token.Valid),
			zap.Any("claims", token.Claims),
			zap.String("request_id", requestID),
		)

		if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
			logger.L.Debug("JWT claims extracted",
				zap.String("user_id", claims.UserID),
				zap.String("username", claims.Username),
				zap.Strings("roles", claims.Roles),
				zap.String("request_id", requestID),
			)
			
			c.Set("user_id", claims.UserID)
			c.Set("username", claims.Username)
			c.Set("roles", claims.Roles)
		} else {
			logger.L.Debug("JWT claims validation failed",
				zap.Bool("claims_ok", ok),
				zap.Bool("token_valid", token.Valid),
				zap.Any("claims", token.Claims),
				zap.String("request_id", requestID),
			)
			
			c.AbortWithStatusJSON(http.StatusUnauthorized, ErrorResponse{
				Code:      http.StatusUnauthorized,
				Message:   "Invalid token claims",
				Error:     "Token validation failed",
				RequestID: requestID,
				Timestamp: time.Now().Unix(),
			})
			return
		}

		c.Next()
	}
}

// RequireRole 角色权限中间件
func RequireRole(requiredRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRoles, exists := c.Get("roles")
		if !exists {
			requestID := c.GetString("request_id")
			
			c.AbortWithStatusJSON(http.StatusForbidden, ErrorResponse{
				Code:      http.StatusForbidden,
				Message:   "Access denied",
				Error:     "User roles not found",
				RequestID: requestID,
				Timestamp: time.Now().Unix(),
			})
			return
		}

		roles, ok := userRoles.([]string)
		if !ok {
			requestID := c.GetString("request_id")
			
			c.AbortWithStatusJSON(http.StatusForbidden, ErrorResponse{
				Code:      http.StatusForbidden,
				Message:   "Access denied",
				Error:     "Invalid user roles format",
				RequestID: requestID,
				Timestamp: time.Now().Unix(),
			})
			return
		}

		// 检查用户是否具有所需角色
		hasRole := false
		for _, userRole := range roles {
			for _, requiredRole := range requiredRoles {
				if userRole == requiredRole || userRole == "admin" { // admin角色拥有所有权限
					hasRole = true
					break
				}
			}
			if hasRole {
				break
			}
		}

		if !hasRole {
			requestID := c.GetString("request_id")
			
			c.AbortWithStatusJSON(http.StatusForbidden, ErrorResponse{
				Code:      http.StatusForbidden,
				Message:   "Access denied",
				Error:     fmt.Sprintf("Required roles: %v", requiredRoles),
				RequestID: requestID,
				Timestamp: time.Now().Unix(),
			})
			return
		}

		c.Next()
	}
}
