package routes

import (
	"github.com/gin-gonic/gin"

	"alert_agent/internal/middleware"
	"alert_agent/internal/security/handler"
)

// SetupAuthRoutes 设置认证相关路由
func SetupAuthRoutes(router *gin.Engine, authHandler *handler.AuthHandler, securityConfig *middleware.SecurityConfig) {
	// 公开路由（无需认证）
	public := router.Group("/api/v1")
	{
		// 用户注册和登录
		public.POST("/auth/register", authHandler.Register)
		public.POST("/auth/login", authHandler.Login)
	}

	// 需要认证的路由
	protected := router.Group("/api/v1")
	protected.Use(middleware.AuthMiddleware(securityConfig))
	protected.Use(middleware.AuditMiddleware(securityConfig))
	{
		// 用户个人操作
		protected.POST("/auth/logout", authHandler.Logout)
		protected.GET("/auth/profile", authHandler.GetProfile)
		protected.PUT("/auth/profile", authHandler.UpdateProfile)
		protected.POST("/auth/change-password", authHandler.ChangePassword)
	}

	// 管理员路由（需要管理员权限）
	admin := router.Group("/api/v1/admin")
	admin.Use(middleware.AuthMiddleware(securityConfig))
	admin.Use(middleware.RoleMiddleware(securityConfig, []string{"admin"}))
	admin.Use(middleware.AuditMiddleware(securityConfig))
	{
		// 用户管理
		admin.GET("/users", authHandler.ListUsers)
		admin.GET("/users/:id", authHandler.GetUser)
		admin.PUT("/users/:id", authHandler.UpdateUser)
		admin.DELETE("/users/:id", authHandler.DeleteUser)
	}
}

// SetupSecurityMiddleware 设置全局安全中间件
func SetupSecurityMiddleware(router *gin.Engine) {
	// 全局安全中间件
	router.Use(middleware.SecurityHeadersMiddleware())
	router.Use(middleware.InputValidationMiddleware())
	router.Use(middleware.RateLimitMiddleware(60)) // 每分钟60个请求

	// CORS 中间件（如果需要）
	// router.Use(cors.Default())

	// 恢复中间件
	router.Use(gin.Recovery())

	// 日志中间件
	router.Use(gin.Logger())
}

// SetupHealthRoutes 设置健康检查路由
func SetupHealthRoutes(router *gin.Engine) {
	health := router.Group("/health")
	{
		health.GET("/ping", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"status":  "ok",
				"message": "AlertAgent is running",
			})
		})

		health.GET("/ready", func(c *gin.Context) {
			// 这里可以添加数据库连接检查等
			c.JSON(200, gin.H{
				"status":  "ready",
				"message": "AlertAgent is ready to serve requests",
			})
		})

		health.GET("/live", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"status":  "alive",
				"message": "AlertAgent is alive",
			})
		})
	}
}