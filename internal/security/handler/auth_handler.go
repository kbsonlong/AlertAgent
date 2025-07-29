package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"alert_agent/internal/security/user"
)

// AuthHandler 认证处理器
type AuthHandler struct {
	userService *user.UserService
}

// NewAuthHandler 创建认证处理器
func NewAuthHandler(userService *user.UserService) *AuthHandler {
	return &AuthHandler{
		userService: userService,
	}
}

// APIResponse 统一API响应格式
type APIResponse struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// ErrorResponse 错误响应
type ErrorResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Error   string `json:"error,omitempty"`
}

// Register 用户注册
func (h *AuthHandler) Register(c *gin.Context) {
	var req user.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    400,
			Message: "Invalid request format",
			Error:   err.Error(),
		})
		return
	}

	// 获取操作者ID（从JWT中获取，如果是注册则为空）
	operatorID, _ := c.Get("user_id")
	operatorIDStr := ""
	if operatorID != nil {
		operatorIDStr = operatorID.(string)
	}

	createdUser, err := h.userService.CreateUser(c.Request.Context(), &req, operatorIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    400,
			Message: "Failed to create user",
			Error:   err.Error(),
		})
		return
	}

	// 不返回敏感信息
	createdUser.PasswordHash = ""
	createdUser.Salt = ""

	c.JSON(http.StatusCreated, APIResponse{
		Code:    201,
		Message: "User created successfully",
		Data:    createdUser,
	})
}

// Login 用户登录
func (h *AuthHandler) Login(c *gin.Context) {
	var req user.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    400,
			Message: "Invalid request format",
			Error:   err.Error(),
		})
		return
	}

	clientIP := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")

	loginResp, err := h.userService.Login(c.Request.Context(), &req, clientIP, userAgent)
	if err != nil {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Code:    401,
			Message: "Login failed",
			Error:   err.Error(),
		})
		return
	}

	// 不返回敏感信息
	loginResp.User.PasswordHash = ""
	loginResp.User.Salt = ""

	c.JSON(http.StatusOK, APIResponse{
		Code:    200,
		Message: "Login successful",
		Data:    loginResp,
	})
}

// Logout 用户登出
func (h *AuthHandler) Logout(c *gin.Context) {
	// 在实际实现中，可以将token加入黑名单
	// 这里简单返回成功响应
	c.JSON(http.StatusOK, APIResponse{
		Code:    200,
		Message: "Logout successful",
	})
}

// GetProfile 获取用户信息
func (h *AuthHandler) GetProfile(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Code:    401,
			Message: "User not authenticated",
		})
		return
	}

	userIDStr := userID.(string)
	user, err := h.userService.GetUser(c.Request.Context(), userIDStr)
	if err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{
			Code:    404,
			Message: "User not found",
			Error:   err.Error(),
		})
		return
	}

	// 不返回敏感信息
	user.PasswordHash = ""
	user.Salt = ""

	c.JSON(http.StatusOK, APIResponse{
		Code:    200,
		Message: "User profile retrieved successfully",
		Data:    user,
	})
}

// UpdateProfile 更新用户信息
func (h *AuthHandler) UpdateProfile(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Code:    401,
			Message: "User not authenticated",
		})
		return
	}

	var req user.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    400,
			Message: "Invalid request format",
			Error:   err.Error(),
		})
		return
	}

	userIDStr := userID.(string)
	updatedUser, err := h.userService.UpdateUser(c.Request.Context(), userIDStr, &req, userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    400,
			Message: "Failed to update user",
			Error:   err.Error(),
		})
		return
	}

	// 不返回敏感信息
	updatedUser.PasswordHash = ""
	updatedUser.Salt = ""

	c.JSON(http.StatusOK, APIResponse{
		Code:    200,
		Message: "User updated successfully",
		Data:    updatedUser,
	})
}

// ChangePassword 修改密码
func (h *AuthHandler) ChangePassword(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Code:    401,
			Message: "User not authenticated",
		})
		return
	}

	var req user.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    400,
			Message: "Invalid request format",
			Error:   err.Error(),
		})
		return
	}

	userIDStr := userID.(string)
	clientIP := c.ClientIP()

	if err := h.userService.ChangePassword(c.Request.Context(), userIDStr, &req, clientIP); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    400,
			Message: "Failed to change password",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, APIResponse{
		Code:    200,
		Message: "Password changed successfully",
	})
}

// ListUsers 获取用户列表（管理员功能）
func (h *AuthHandler) ListUsers(c *gin.Context) {
	// 解析分页参数
	offsetStr := c.DefaultQuery("offset", "0")
	limitStr := c.DefaultQuery("limit", "10")

	offset, err := strconv.Atoi(offsetStr)
	if err != nil {
		offset = 0
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit > 100 {
		limit = 10
	}

	users, total, err := h.userService.ListUsers(c.Request.Context(), offset, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    500,
			Message: "Failed to retrieve users",
			Error:   err.Error(),
		})
		return
	}

	// 不返回敏感信息
	for _, user := range users {
		user.PasswordHash = ""
		user.Salt = ""
	}

	c.JSON(http.StatusOK, APIResponse{
		Code:    200,
		Message: "Users retrieved successfully",
		Data: map[string]interface{}{
			"users":  users,
			"total":  total,
			"offset": offset,
			"limit":  limit,
		},
	})
}

// GetUser 获取指定用户信息（管理员功能）
func (h *AuthHandler) GetUser(c *gin.Context) {
	userID := c.Param("id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    400,
			Message: "User ID is required",
		})
		return
	}

	user, err := h.userService.GetUser(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{
			Code:    404,
			Message: "User not found",
			Error:   err.Error(),
		})
		return
	}

	// 不返回敏感信息
	user.PasswordHash = ""
	user.Salt = ""

	c.JSON(http.StatusOK, APIResponse{
		Code:    200,
		Message: "User retrieved successfully",
		Data:    user,
	})
}

// UpdateUser 更新指定用户信息（管理员功能）
func (h *AuthHandler) UpdateUser(c *gin.Context) {
	userID := c.Param("id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    400,
			Message: "User ID is required",
		})
		return
	}

	var req user.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    400,
			Message: "Invalid request format",
			Error:   err.Error(),
		})
		return
	}

	// 获取操作者ID
	operatorID, _ := c.Get("user_id")
	operatorIDStr := ""
	if operatorID != nil {
		operatorIDStr = operatorID.(string)
	}

	updatedUser, err := h.userService.UpdateUser(c.Request.Context(), userID, &req, operatorIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    400,
			Message: "Failed to update user",
			Error:   err.Error(),
		})
		return
	}

	// 不返回敏感信息
	updatedUser.PasswordHash = ""
	updatedUser.Salt = ""

	c.JSON(http.StatusOK, APIResponse{
		Code:    200,
		Message: "User updated successfully",
		Data:    updatedUser,
	})
}

// DeleteUser 删除用户（管理员功能）
func (h *AuthHandler) DeleteUser(c *gin.Context) {
	userID := c.Param("id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    400,
			Message: "User ID is required",
		})
		return
	}

	// 获取操作者ID
	operatorID, _ := c.Get("user_id")
	operatorIDStr := ""
	if operatorID != nil {
		operatorIDStr = operatorID.(string)
	}

	// 防止用户删除自己
	if operatorIDStr == userID {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    400,
			Message: "Cannot delete your own account",
		})
		return
	}

	if err := h.userService.DeleteUser(c.Request.Context(), userID, operatorIDStr); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    400,
			Message: "Failed to delete user",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, APIResponse{
		Code:    200,
		Message: "User deleted successfully",
	})
}