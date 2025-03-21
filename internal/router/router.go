package router

import (
	v1 "alert_agent/internal/api/v1"
	"alert_agent/internal/middleware"

	"github.com/gin-gonic/gin"
)

// RegisterRoutes 注册所有路由
func RegisterRoutes(r *gin.Engine) {
	// 中间件
	r.Use(middleware.Logger())
	r.Use(middleware.Recovery())
	r.Use(middleware.Cors())

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
			// alerts.POST("/:id/analyze", v1.AnalyzeAlert)
			alerts.POST("/:id/async-analyze", v1.AsyncAnalyzeAlert)
			alerts.GET("/:id/analysis-status", v1.GetAnalysisStatus)
			alerts.GET("/:id/similar", v1.FindSimilarAlerts)

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
		}
	}
}
