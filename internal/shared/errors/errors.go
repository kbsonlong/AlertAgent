package errors

import (
	"fmt"
	"net/http"
)

// ErrorType 错误类型
type ErrorType string

const (
	ErrorTypeValidation   ErrorType = "validation"
	ErrorTypeNotFound     ErrorType = "not_found"
	ErrorTypeConflict     ErrorType = "conflict"
	ErrorTypeInternal     ErrorType = "internal"
	ErrorTypeExternal     ErrorType = "external"
	ErrorTypeTimeout      ErrorType = "timeout"
	ErrorTypeRateLimit    ErrorType = "rate_limit"
	ErrorTypeUnauthorized ErrorType = "unauthorized"
	ErrorTypeForbidden    ErrorType = "forbidden"
)

// AppError 应用错误
type AppError struct {
	Type    ErrorType `json:"type"`
	Code    string    `json:"code"`
	Message string    `json:"message"`
	Details string    `json:"details,omitempty"`
	Cause   error     `json:"-"`
}

func (e *AppError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("[%s] %s: %s (caused by: %v)", e.Type, e.Code, e.Message, e.Cause)
	}
	return fmt.Sprintf("[%s] %s: %s", e.Type, e.Code, e.Message)
}

// Unwrap 支持errors.Unwrap
func (e *AppError) Unwrap() error {
	return e.Cause
}

// NewValidationError 创建验证错误
func NewValidationError(code, message string) *AppError {
	return &AppError{
		Type:    ErrorTypeValidation,
		Code:    code,
		Message: message,
	}
}

// NewValidationErrorWithDetails 创建带详情的验证错误
func NewValidationErrorWithDetails(code, message, details string) *AppError {
	return &AppError{
		Type:    ErrorTypeValidation,
		Code:    code,
		Message: message,
		Details: details,
	}
}

// NewNotFoundError 创建未找到错误
func NewNotFoundError(resource string) *AppError {
	return &AppError{
		Type:    ErrorTypeNotFound,
		Code:    "RESOURCE_NOT_FOUND",
		Message: fmt.Sprintf("%s not found", resource),
	}
}

// NewConflictError 创建冲突错误
func NewConflictError(message string) *AppError {
	return &AppError{
		Type:    ErrorTypeConflict,
		Code:    "RESOURCE_CONFLICT",
		Message: message,
	}
}

// NewInternalError 创建内部错误
func NewInternalError(message string, cause error) *AppError {
	return &AppError{
		Type:    ErrorTypeInternal,
		Code:    "INTERNAL_ERROR",
		Message: message,
		Cause:   cause,
	}
}

// NewExternalError 创建外部服务错误
func NewExternalError(service, message string, cause error) *AppError {
	return &AppError{
		Type:    ErrorTypeExternal,
		Code:    "EXTERNAL_SERVICE_ERROR",
		Message: fmt.Sprintf("%s: %s", service, message),
		Cause:   cause,
	}
}

// NewTimeoutError 创建超时错误
func NewTimeoutError(operation string) *AppError {
	return &AppError{
		Type:    ErrorTypeTimeout,
		Code:    "OPERATION_TIMEOUT",
		Message: fmt.Sprintf("%s operation timed out", operation),
	}
}

// NewRateLimitError 创建限流错误
func NewRateLimitError(message string) *AppError {
	return &AppError{
		Type:    ErrorTypeRateLimit,
		Code:    "RATE_LIMIT_EXCEEDED",
		Message: message,
	}
}

// NewUnauthorizedError 创建未授权错误
func NewUnauthorizedError(message string) *AppError {
	return &AppError{
		Type:    ErrorTypeUnauthorized,
		Code:    "UNAUTHORIZED",
		Message: message,
	}
}

// NewForbiddenError 创建禁止访问错误
func NewForbiddenError(message string) *AppError {
	return &AppError{
		Type:    ErrorTypeForbidden,
		Code:    "FORBIDDEN",
		Message: message,
	}
}

// GetHTTPStatusCode 获取HTTP状态码
func GetHTTPStatusCode(err error) int {
	if appErr, ok := err.(*AppError); ok {
		switch appErr.Type {
		case ErrorTypeValidation:
			return http.StatusBadRequest
		case ErrorTypeNotFound:
			return http.StatusNotFound
		case ErrorTypeConflict:
			return http.StatusConflict
		case ErrorTypeUnauthorized:
			return http.StatusUnauthorized
		case ErrorTypeForbidden:
			return http.StatusForbidden
		case ErrorTypeTimeout:
			return http.StatusRequestTimeout
		case ErrorTypeRateLimit:
			return http.StatusTooManyRequests
		case ErrorTypeExternal:
			return http.StatusBadGateway
		default:
			return http.StatusInternalServerError
		}
	}
	return http.StatusInternalServerError
}

// IsAppError 检查是否为应用错误
func IsAppError(err error) bool {
	_, ok := err.(*AppError)
	return ok
}

// IsErrorType 检查错误类型
func IsErrorType(err error, errorType ErrorType) bool {
	if appErr, ok := err.(*AppError); ok {
		return appErr.Type == errorType
	}
	return false
}

// WrapError 包装错误
func WrapError(err error, message string) *AppError {
	if appErr, ok := err.(*AppError); ok {
		return &AppError{
			Type:    appErr.Type,
			Code:    appErr.Code,
			Message: message,
			Details: appErr.Message,
			Cause:   appErr.Cause,
		}
	}
	return NewInternalError(message, err)
}