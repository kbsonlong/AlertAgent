package router

import (
	v1 "alert_agent/internal/api/v1"
	"alert_agent/internal/config"
	"alert_agent/internal/middleware"
	"alert_agent/internal/pkg/queue"
	"alert_agent/internal/pkg/redis"

	"github.com/gin-gonic/gin"
)

// RegisterRoutes 注册所有路由
func RegisterRoutes(r *gin.Engine) {
	cfg := config.GetConfig()
	
	// 创建Redis队列
	redisQueue := queue.NewRedisQueue(redis.Client, "alert:queue")

	// 创建异步告警处理器
	asyncAlertHandler := v1.NewAsyncAlertHandler(redisQueue)

	// API v1
	apiV1 := r.Group("/api/v1")
	{
		// 健康检查 - 无需认证
		apiV1.GET("/health", v1.HealthCheck)

		// 认证相关路由 - 无需认证
		v1.RegisterAuthRoutes(apiV1)

		// 需要认证的路由组
		authenticated := apiV1.Group("")
		if cfg.Gateway.Auth.Enabled {
			authenticated.Use(middleware.JWTAuth())
		}

		// 告警规则管理 - 需要认证
		rules := authenticated.Group("/rules")
		{
			rules.GET("", v1.ListRules)
			rules.POST("", middleware.RequireRole("admin", "operator"), v1.CreateRule)
			rules.GET("/:id", v1.GetRule)
			rules.PUT("/:id", middleware.RequireRole("admin", "operator"), v1.UpdateRule)
			rules.DELETE("/:id", middleware.RequireRole("admin"), v1.DeleteRule)
		}

		// 告警记录管理 - 需要认证
		alerts := authenticated.Group("/alerts")
		{
			alerts.GET("", v1.ListAlerts)
			alerts.POST("", middleware.RequireRole("admin", "operator"), v1.CreateAlert)
			alerts.GET("/:id", v1.GetAlert)
			alerts.PUT("/:id", middleware.RequireRole("admin", "operator"), v1.UpdateAlert)
			alerts.POST("/:id/handle", middleware.RequireRole("admin", "operator"), v1.HandleAlert)
			alerts.GET("/:id/similar", v1.FindSimilarAlerts)
			alerts.POST("/:id/analyze", v1.AnalyzeAlert)
			alerts.POST("/:id/convert-to-knowledge", middleware.RequireRole("admin", "operator"), v1.ConvertAlertToKnowledge)

			// 异步分析告警
			asyncAlertHandler.RegisterRoutes(alerts)
		}

		// 通知模板管理 - 需要认证
		templates := authenticated.Group("/templates")
		{
			templates.GET("", v1.ListTemplates)
			templates.POST("", middleware.RequireRole("admin", "operator"), v1.CreateTemplate)
			templates.GET("/:id", v1.GetTemplate)
			templates.PUT("/:id", middleware.RequireRole("admin", "operator"), v1.UpdateTemplate)
			templates.DELETE("/:id", middleware.RequireRole("admin"), v1.DeleteTemplate)
		}

		// 通知组管理 - 需要认证
		groups := authenticated.Group("/groups")
		{
			groups.GET("", v1.ListGroups)
			groups.POST("", middleware.RequireRole("admin", "operator"), v1.CreateGroup)
			groups.GET("/:id", v1.GetGroup)
			groups.PUT("/:id", middleware.RequireRole("admin", "operator"), v1.UpdateGroup)
			groups.DELETE("/:id", middleware.RequireRole("admin"), v1.DeleteGroup)
		}

		// 系统设置 - 需要管理员权限
		settings := authenticated.Group("/settings")
		{
			settings.GET("", middleware.RequireRole("admin"), v1.GetSettings)
			settings.PUT("", middleware.RequireRole("admin"), v1.UpdateSettings)
		}

		// 配置管理 - 需要管理员权限
		v1.RegisterConfigRoutes(authenticated, middleware.RequireRole("admin"))

		// 知识库管理 - 需要认证
		knowledge := authenticated.Group("/knowledge")
		{
			knowledge.GET("", v1.ListKnowledge)
			knowledge.POST("", middleware.RequireRole("admin", "operator"), v1.CreateKnowledge)
			knowledge.GET("/:id", v1.GetKnowledge)
			knowledge.PUT("/:id", middleware.RequireRole("admin", "operator"), v1.UpdateKnowledge)
			knowledge.DELETE("/:id", middleware.RequireRole("admin"), v1.DeleteKnowledge)
		}

		// 数据源管理 - 需要认证
		providers := authenticated.Group("/providers")
		{
			providers.GET("", v1.ListProviders)
			providers.POST("", middleware.RequireRole("admin"), v1.CreateProvider)
			providers.GET("/:id", v1.GetProvider)
			providers.PUT("/:id", middleware.RequireRole("admin"), v1.UpdateProvider)
			providers.DELETE("/:id", middleware.RequireRole("admin"), v1.DeleteProvider)
			providers.POST("/test", middleware.RequireRole("admin", "operator"), v1.TestProvider)
		}
	}
}
