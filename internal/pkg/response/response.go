package response

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// Response 标准响应结构
type Response struct {
	Code      int         `json:"code"`
	Message   string      `json:"message"`
	Data      interface{} `json:"data,omitempty"`
	RequestID string      `json:"request_id,omitempty"`
	Timestamp int64       `json:"timestamp"`
}

// ErrorResponse 错误响应结构
type ErrorResponse struct {
	Code      int         `json:"code"`
	Message   string      `json:"message"`
	Error     string      `json:"error,omitempty"`
	RequestID string      `json:"request_id,omitempty"`
	Timestamp int64       `json:"timestamp"`
}

// PaginationResponse 分页响应结构
type PaginationResponse struct {
	Code      int         `json:"code"`
	Message   string      `json:"message"`
	Data      interface{} `json:"data"`
	Total     int64       `json:"total"`
	Page      int         `json:"page"`
	PageSize  int         `json:"page_size"`
	RequestID string      `json:"request_id,omitempty"`
	Timestamp int64       `json:"timestamp"`
}

// Success 成功响应
func Success(c *gin.Context, data interface{}) {
	requestID := c.GetString("request_id")
	
	c.JSON(http.StatusOK, Response{
		Code:      http.StatusOK,
		Message:   "Success",
		Data:      data,
		RequestID: requestID,
		Timestamp: time.Now().Unix(),
	})
}

// SuccessWithMessage 带消息的成功响应
func SuccessWithMessage(c *gin.Context, message string, data interface{}) {
	requestID := c.GetString("request_id")
	
	c.JSON(http.StatusOK, Response{
		Code:      http.StatusOK,
		Message:   message,
		Data:      data,
		RequestID: requestID,
		Timestamp: time.Now().Unix(),
	})
}

// Created 创建成功响应
func Created(c *gin.Context, data interface{}) {
	requestID := c.GetString("request_id")
	
	c.JSON(http.StatusCreated, Response{
		Code:      http.StatusCreated,
		Message:   "Created successfully",
		Data:      data,
		RequestID: requestID,
		Timestamp: time.Now().Unix(),
	})
}

// Updated 更新成功响应
func Updated(c *gin.Context, data interface{}) {
	requestID := c.GetString("request_id")
	
	c.JSON(http.StatusOK, Response{
		Code:      http.StatusOK,
		Message:   "Updated successfully",
		Data:      data,
		RequestID: requestID,
		Timestamp: time.Now().Unix(),
	})
}

// Deleted 删除成功响应
func Deleted(c *gin.Context) {
	requestID := c.GetString("request_id")
	
	c.JSON(http.StatusOK, Response{
		Code:      http.StatusOK,
		Message:   "Deleted successfully",
		RequestID: requestID,
		Timestamp: time.Now().Unix(),
	})
}

// Pagination 分页响应
func Pagination(c *gin.Context, data interface{}, total int64, page, pageSize int) {
	requestID := c.GetString("request_id")
	
	c.JSON(http.StatusOK, PaginationResponse{
		Code:      http.StatusOK,
		Message:   "Success",
		Data:      data,
		Total:     total,
		Page:      page,
		PageSize:  pageSize,
		RequestID: requestID,
		Timestamp: time.Now().Unix(),
	})
}

// Error 错误响应
func Error(c *gin.Context, code int, message string, err error) {
	requestID := c.GetString("request_id")
	
	errorMsg := ""
	if err != nil {
		errorMsg = err.Error()
	}
	
	c.JSON(code, ErrorResponse{
		Code:      code,
		Message:   message,
		Error:     errorMsg,
		RequestID: requestID,
		Timestamp: time.Now().Unix(),
	})
}

// BadRequest 400错误响应
func BadRequest(c *gin.Context, message string, err error) {
	Error(c, http.StatusBadRequest, message, err)
}

// Unauthorized 401错误响应
func Unauthorized(c *gin.Context, message string, err error) {
	Error(c, http.StatusUnauthorized, message, err)
}

// Forbidden 403错误响应
func Forbidden(c *gin.Context, message string, err error) {
	Error(c, http.StatusForbidden, message, err)
}

// NotFound 404错误响应
func NotFound(c *gin.Context, message string, err error) {
	Error(c, http.StatusNotFound, message, err)
}

// Conflict 409错误响应
func Conflict(c *gin.Context, message string, err error) {
	Error(c, http.StatusConflict, message, err)
}

// InternalServerError 500错误响应
func InternalServerError(c *gin.Context, message string, err error) {
	Error(c, http.StatusInternalServerError, message, err)
}

// ValidationError 参数验证错误响应
func ValidationError(c *gin.Context, err error) {
	BadRequest(c, "Validation failed", err)
}