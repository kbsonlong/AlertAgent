package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"alert_agent/internal/shared/errors"
	"alert_agent/internal/shared/logger"
)

// ErrorHandler 错误处理中间件
func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// 处理错误
		if len(c.Errors) > 0 {
			err := c.Errors.Last().Err
			handleError(c, err)
		}
	}
}

// handleError 处理错误
func handleError(c *gin.Context, err error) {
	log := logger.WithContext(c.Request.Context())

	// 检查是否为应用错误
	if appErr, ok := err.(*errors.AppError); ok {
		handleAppError(c, appErr, log)
	} else {
		handleUnknownError(c, err, log)
	}
}

// handleAppError 处理应用错误
func handleAppError(c *gin.Context, err *errors.AppError, log *zap.Logger) {
	statusCode := errors.GetHTTPStatusCode(err)

	response := gin.H{
		"success": false,
		"code":    statusCode,
		"type":    err.Type,
		"message": err.Message,
	}

	if err.Details != "" {
		response["details"] = err.Details
	}

	// 记录错误日志
	logFields := []zap.Field{
		zap.String("type", string(err.Type)),
		zap.String("code", err.Code),
		zap.String("message", err.Message),
		zap.String("path", c.Request.URL.Path),
		zap.String("method", c.Request.Method),
	}

	if err.Cause != nil {
		logFields = append(logFields, zap.Error(err.Cause))
	}

	// 根据错误类型选择日志级别
	switch err.Type {
	case errors.ErrorTypeValidation, errors.ErrorTypeNotFound, errors.ErrorTypeConflict:
		log.Warn("Application error", logFields...)
	case errors.ErrorTypeTimeout, errors.ErrorTypeRateLimit:
		log.Warn("Application error", logFields...)
	default:
		log.Error("Application error", logFields...)
	}

	c.JSON(statusCode, response)
}

// handleUnknownError 处理未知错误
func handleUnknownError(c *gin.Context, err error, log *zap.Logger) {
	response := gin.H{
		"success": false,
		"code":    http.StatusInternalServerError,
		"type":    "internal",
		"message": "Internal server error",
	}

	log.Error("Unknown error",
		zap.Error(err),
		zap.String("path", c.Request.URL.Path),
		zap.String("method", c.Request.Method),
	)

	c.JSON(http.StatusInternalServerError, response)
}

// RecoveryHandler 恢复处理中间件
func RecoveryHandler() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		log := logger.WithContext(c.Request.Context())

		log.Error("Panic recovered",
			zap.Any("panic", recovered),
			zap.String("path", c.Request.URL.Path),
			zap.String("method", c.Request.Method),
		)

		response := gin.H{
			"success": false,
			"code":    http.StatusInternalServerError,
			"type":    "internal",
			"message": "Internal server error",
		}

		c.JSON(http.StatusInternalServerError, response)
	})
}