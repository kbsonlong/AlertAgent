package v1

import (
	"alert_agent/internal/pkg/response"
	"alert_agent/internal/service"

	"github.com/gin-gonic/gin"
)

// AuthHandler 认证处理器
type AuthHandler struct {
	authService *service.AuthService
}

// NewAuthHandler 创建认证处理器实例
func NewAuthHandler() *AuthHandler {
	return &AuthHandler{
		authService: service.NewAuthService(),
	}
}

// Login 用户登录
// @Summary 用户登录
// @Description 用户登录获取访问令牌
// @Tags 认证
// @Accept json
// @Produce json
// @Param request body service.LoginRequest true "登录请求"
// @Success 200 {object} response.Response{data=service.LoginResponse}
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Router /api/v1/auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req service.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	loginResp, err := h.authService.Login(req)
	if err != nil {
		response.Unauthorized(c, "Login failed", err)
		return
	}

	response.SuccessWithMessage(c, "Login successful", loginResp)
}

// RefreshToken 刷新访问令牌
// @Summary 刷新访问令牌
// @Description 使用刷新令牌获取新的访问令牌
// @Tags 认证
// @Accept json
// @Produce json
// @Param request body service.RefreshTokenRequest true "刷新令牌请求"
// @Success 200 {object} response.Response{data=service.LoginResponse}
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Router /api/v1/auth/refresh [post]
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req service.RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	loginResp, err := h.authService.RefreshTokens(req.RefreshToken)
	if err != nil {
		response.Unauthorized(c, "Token refresh failed", err)
		return
	}

	response.SuccessWithMessage(c, "Token refreshed successfully", loginResp)
}

// Logout 用户登出
// @Summary 用户登出
// @Description 用户登出，使令牌失效
// @Tags 认证
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.Response
// @Failure 401 {object} response.ErrorResponse
// @Router /api/v1/auth/logout [post]
func (h *AuthHandler) Logout(c *gin.Context) {
	// 从请求头获取令牌
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		response.Unauthorized(c, "Authorization header required", nil)
		return
	}

	tokenString := authHeader[7:] // 移除 "Bearer " 前缀
	
	if err := h.authService.Logout(tokenString); err != nil {
		response.InternalServerError(c, "Logout failed", err)
		return
	}

	response.SuccessWithMessage(c, "Logout successful", nil)
}

// GetProfile 获取用户信息
// @Summary 获取用户信息
// @Description 获取当前登录用户的信息
// @Tags 认证
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.Response{data=service.User}
// @Failure 401 {object} response.ErrorResponse
// @Router /api/v1/auth/profile [get]
func (h *AuthHandler) GetProfile(c *gin.Context) {
	userID := c.GetString("user_id")
	username := c.GetString("username")
	roles, _ := c.Get("roles")

	user := service.User{
		ID:       userID,
		Username: username,
		Roles:    roles.([]string),
	}

	response.Success(c, user)
}

// RegisterAuthRoutes 注册认证路由
func RegisterAuthRoutes(r *gin.RouterGroup) {
	authHandler := NewAuthHandler()
	
	auth := r.Group("/auth")
	{
		auth.POST("/login", authHandler.Login)
		auth.POST("/refresh", authHandler.RefreshToken)
		auth.POST("/logout", authHandler.Logout)
		auth.GET("/profile", authHandler.GetProfile)
	}
}