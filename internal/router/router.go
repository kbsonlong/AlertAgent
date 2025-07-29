package router

import (
	v1 "alert_agent/internal/api/v1"
	"alert_agent/internal/middleware"
	"alert_agent/internal/pkg/logger"
	"alert_agent/internal/pkg/queue"
	"alert_agent/internal/pkg/redis"
	"alert_agent/internal/service"

	"github.com/gin-gonic/gin"
)

// RegisterRoutes 注册所有路由
func RegisterRoutes(r *gin.Engine, featureService *service.FeatureService) {
	// 中间件
	r.Use(middleware.Logger())
	r.Use(middleware.Recovery())
	r.Use(middleware.Cors())

	// 创建Redis队列
	redisQueue := queue.NewRedisQueue(redis.Client, "alert:queue")

	// 创建异步告警处理器
	asyncAlertHandler := v1.NewAsyncAlertHandler(redisQueue)

	// API v1
	apiV1 := r.Group("/api/v1")
	{
		// 健康检查
		// apiV1.GET("/health", v1.HealthCheck)

		// 告警规则管理
		rules := apiV1.Group("/rules")
		{
			rules.GET("", v1.ListRules)
			rules.POST("", v1.CreateRule)
			rules.GET("/:id", v1.GetRule)
			rules.PUT("/:id", v1.UpdateRule)
			rules.DELETE("/:id", v1.DeleteRule)
		}

		// 告警记录管理
		alerts := apiV1.Group("/alerts")
		{
			alerts.GET("", v1.ListAlerts)
			alerts.POST("", v1.CreateAlert)
			alerts.GET("/:id", v1.GetAlert)
			alerts.PUT("/:id", v1.UpdateAlert)
			alerts.POST("/:id/handle", v1.HandleAlert)
			alerts.GET("/:id/similar", v1.FindSimilarAlerts)
			alerts.POST("/:id/analyze", v1.AnalyzeAlert)
			alerts.POST("/:id/convert-to-knowledge", v1.ConvertAlertToKnowledge)

			// 通知模板管理
			templates := apiV1.Group("/templates")
			{
				templates.GET("", v1.ListTemplates)
				templates.POST("", v1.CreateTemplate)
				templates.GET("/:id", v1.GetTemplate)
				templates.PUT("/:id", v1.UpdateTemplate)
				templates.DELETE("/:id", v1.DeleteTemplate)
			}

			// 通知组管理
			groups := apiV1.Group("/groups")
			{
				groups.GET("", v1.ListGroups)
				groups.POST("", v1.CreateGroup)
				groups.GET("/:id", v1.GetGroup)
				groups.PUT("/:id", v1.UpdateGroup)
				groups.DELETE("/:id", v1.DeleteGroup)
			}

			// 系统设置
		settings := apiV1.Group("/settings")
		{
			settings.GET("", v1.GetSettings)
			settings.PUT("", v1.UpdateSettings)
		}

		// 系统配置
		apiV1.GET("/config", v1.GetSystemConfig)
		apiV1.GET("/config/current", v1.GetCurrentConfig)

			// 异步分析告警
			asyncAlertHandler.RegisterRoutes(alerts)
		}

		// 知识库管理
		knowledge := apiV1.Group("/knowledge")
		{
			knowledge.GET("", v1.ListKnowledge)
			knowledge.POST("", v1.CreateKnowledge)
			knowledge.GET("/:id", v1.GetKnowledge)
			knowledge.PUT("/:id", v1.UpdateKnowledge)
			knowledge.DELETE("/:id", v1.DeleteKnowledge)
		}

		// 数据源管理
		providers := apiV1.Group("/providers")
		{
			providers.GET("", v1.ListProviders)
			providers.POST("", v1.CreateProvider)
			providers.GET("/:id", v1.GetProvider)
			providers.PUT("/:id", v1.UpdateProvider)
			providers.DELETE("/:id", v1.DeleteProvider)
			providers.POST("/test", v1.TestProvider)
		}

		// 功能开关管理
		if featureService != nil {
			featureHandler := v1.NewFeatureHandler(featureService, logger.L)
			features := apiV1.Group("/features")
			{
				features.GET("", featureHandler.ListFeatures)
				features.GET("/:name", featureHandler.GetFeature)
				features.PUT("/:name", featureHandler.UpdateFeature)
				features.GET("/:name/check", featureHandler.CheckFeature)
				features.GET("/:name/ai-maturity", featureHandler.GetAIMaturity)
				features.POST("/:name/ai-metrics", featureHandler.RecordAIMetrics)
				features.GET("/:name/monitoring", featureHandler.GetMonitoringReport)
				features.POST("/phases/:phase/enable", featureHandler.EnablePhase)
				features.POST("/phases/:phase/disable", featureHandler.DisablePhase)
				features.GET("/alerts", featureHandler.GetActiveAlerts)
				features.GET("/export", featureHandler.ExportConfig)
				features.POST("/import", featureHandler.ImportConfig)
			}
		}
	}
}
