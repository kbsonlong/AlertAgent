package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	"alert_agent/internal/security/di"
	"alert_agent/internal/security/routes"
)

func main() {
	// 初始化依赖注入容器
	container, err := di.NewContainer()
	if err != nil {
		log.Fatalf("Failed to initialize container: %v", err)
	}
	defer container.Cleanup()

	// 创建Gin引擎
	router := gin.New()

	// 设置全局安全中间件
	routes.SetupSecurityMiddleware(router)

	// 设置健康检查路由
	routes.SetupHealthRoutes(router)

	// 设置认证路由
	routes.SetupAuthRoutes(
		router,
		container.GetAuthHandler(),
		container.GetMiddlewareConfig(),
	)

	// 启动服务器
	log.Println("Starting security demo server on :8080")
	log.Println("Available endpoints:")
	log.Println("  Health: GET /health/ping")
	log.Println("  Register: POST /api/v1/auth/register")
	log.Println("  Login: POST /api/v1/auth/login")
	log.Println("  Profile: GET /api/v1/auth/profile (requires auth)")
	log.Println("  Admin Users: GET /api/v1/admin/users (requires admin role)")

	if err := http.ListenAndServe(":8080", router); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}