package router

import (
	v1 "alert_agent/internal/api/v1"
	"alert_agent/internal/config"
	"alert_agent/internal/container"
	"alert_agent/internal/middleware"

	"github.com/gin-gonic/gin"
	ginSwagger "github.com/swaggo/gin-swagger"
	swaggerFiles "github.com/swaggo/files"
)

// RegisterRoutes 注册所有路由
func RegisterRoutes(r *gin.Engine) {
	cfg := config.GetConfig()
	
	// 创建依赖注入容器
	container := container.NewContainer()

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

		// 需要认证的认证路由
		v1.RegisterProtectedAuthRoutes(authenticated)

		// 用户管理 - 需要认证
		users := authenticated.Group("/users")
		{
			users.GET("", v1.ListUsers)
			users.POST("", middleware.RequireRole("admin"), v1.CreateUser)
			users.GET("/stats", v1.GetUserStats)
			users.PUT("/batch", middleware.RequireRole("admin"), v1.BatchUpdateUsers)
			users.GET("/:id", v1.GetUser)
			users.PUT("/:id", middleware.RequireRole("admin"), v1.UpdateUser)
			users.DELETE("/:id", middleware.RequireRole("admin"), v1.DeleteUser)
			users.GET("/:id/permissions", v1.GetUserPermissions)
			users.PUT("/:id/permissions", middleware.RequireRole("admin"), v1.UpdateUserPermissions)
		}

		// 告警规则管理 - 需要认证 (使用新的RuleAPI)
		rules := authenticated.Group("/rules")
		{
			rules.GET("", container.RuleAPI.ListRules)
			rules.GET("/stats", v1.GetRuleStats)
			rules.POST("", middleware.RequireRole("admin", "operator"), container.RuleAPI.CreateRule)
			rules.GET("/:id", container.RuleAPI.GetRule)
			rules.PUT("/:id", middleware.RequireRole("admin", "operator"), container.RuleAPI.UpdateRule)
			rules.DELETE("/:id", middleware.RequireRole("admin"), container.RuleAPI.DeleteRule)
			rules.POST("/validate", middleware.RequireRole("admin", "operator"), container.RuleAPI.ValidateRule)
			
			// 规则分发相关路由
			rules.GET("/:id/distribution", container.RuleAPI.GetRuleDistribution)
			rules.GET("/:id/distribution/:target", container.RuleAPI.GetTargetDistribution)
			rules.POST("/distribution/summary", container.RuleAPI.GetDistributionSummary)
			rules.POST("/distribution/retry", middleware.RequireRole("admin", "operator"), container.RuleAPI.RetryDistribution)
			rules.PUT("/distribution/status", middleware.RequireRole("admin", "operator"), container.RuleAPI.UpdateDistributionStatus)
			rules.GET("/distribution/retryable", middleware.RequireRole("admin", "operator"), container.RuleAPI.GetRetryableDistributions)
			rules.POST("/distribution/process-retry", middleware.RequireRole("admin", "operator"), container.RuleAPI.ProcessRetryableDistributions)
			
			// 批量操作
			rules.POST("/batch", middleware.RequireRole("admin", "operator"), container.RuleAPI.BatchRuleOperation)
			
			// 版本控制相关路由
			rules.GET("/:id/versions", container.RuleVersionAPI.GetRuleVersions)
			rules.GET("/:id/versions/:version", container.RuleVersionAPI.GetRuleVersion)
			rules.POST("/:id/rollback", middleware.RequireRole("admin", "operator"), container.RuleVersionAPI.RollbackRule)
			rules.GET("/:id/audit-logs", container.RuleVersionAPI.GetRuleAuditLogs)
			
			// 版本对比
			rules.POST("/versions/compare", container.RuleVersionAPI.CompareRuleVersions)
			
			// 带审计的规则操作
			rules.POST("/audit", middleware.RequireRole("admin", "operator"), container.RuleVersionAPI.CreateRuleWithAudit)
			rules.PUT("/:id/audit", middleware.RequireRole("admin", "operator"), container.RuleVersionAPI.UpdateRuleWithAudit)
			rules.DELETE("/:id/audit", middleware.RequireRole("admin"), container.RuleVersionAPI.DeleteRuleWithAudit)
			
			// 全局审计日志
			rules.GET("/audit-logs", middleware.RequireRole("admin"), container.RuleVersionAPI.GetAllAuditLogs)
		}

		// 告警记录管理 - 需要认证
		alerts := authenticated.Group("/alerts")
		{
			alerts.GET("", v1.ListAlerts)
			alerts.GET("/stats", v1.GetAlertStats)
			alerts.POST("", middleware.RequireRole("admin", "operator"), v1.CreateAlert)
			alerts.GET("/:id", v1.GetAlert)
			alerts.PUT("/:id", middleware.RequireRole("admin", "operator"), v1.UpdateAlert)
			alerts.POST("/:id/handle", middleware.RequireRole("admin", "operator"), v1.HandleAlert)
			alerts.GET("/:id/similar", v1.FindSimilarAlerts)
			alerts.POST("/:id/analyze", v1.AnalyzeAlert)
			alerts.POST("/:id/convert-to-knowledge", middleware.RequireRole("admin", "operator"), v1.ConvertAlertToKnowledge)

			// 异步分析告警路由已集成到告警路由中
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

		// 权限管理 - 需要管理员权限
		permissions := authenticated.Group("/permissions")
		permissions.Use(middleware.RequireRole("admin"))
		{
			permissions.GET("", container.PermissionController.ListPermissions)
			permissions.POST("", container.PermissionController.CreatePermission)
			permissions.GET("/stats", container.PermissionController.GetPermissionStats)
			permissions.POST("/init", container.PermissionController.InitializeSystemPermissions)
			permissions.GET("/:id", container.PermissionController.GetPermission)
			permissions.PUT("/:id", container.PermissionController.UpdatePermission)
			permissions.DELETE("/:id", container.PermissionController.DeletePermission)
		}

		// 角色管理 - 需要管理员权限
		roles := authenticated.Group("/roles")
		roles.Use(middleware.RequireRole("admin"))
		{
			roles.GET("", container.RoleController.ListRoles)
			roles.POST("", container.RoleController.CreateRole)
			roles.GET("/stats", container.RoleController.GetRoleStats)
			roles.POST("/init", container.RoleController.InitializeSystemRoles)
			roles.GET("/:id", container.RoleController.GetRole)
			roles.PUT("/:id", container.RoleController.UpdateRole)
			roles.DELETE("/:id", container.RoleController.DeleteRole)
			
			// 角色权限管理
			roles.GET("/:id/permissions", container.RoleController.GetRolePermissions)
			roles.POST("/:id/permissions", container.RoleController.AssignPermissions)
			roles.DELETE("/:id/permissions", container.RoleController.RemovePermissions)
			roles.PUT("/:id/permissions", container.RoleController.SyncPermissions)
		}

		// 知识库管理 - 需要认证
		knowledge := authenticated.Group("/knowledge")
		{
			knowledge.GET("", v1.ListKnowledge)
			knowledge.POST("", middleware.RequireRole("admin", "operator"), v1.CreateKnowledge)
			knowledge.GET("/categories", v1.GetKnowledgeCategories)
			knowledge.GET("/tags", v1.GetKnowledgeTags)
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

		// 队列管理 - 需要认证
		queues := authenticated.Group("/queues")
		{
			// 队列指标
			queues.GET("/metrics", container.QueueAPI.GetAllQueueMetrics)
			queues.GET("/:queue_name/metrics", container.QueueAPI.GetQueueMetrics)
			queues.GET("/task-metrics/:task_type", container.QueueAPI.GetTaskMetrics)
			
			// 任务管理
			queues.GET("/tasks/:task_id", container.QueueAPI.GetTaskStatus)
			queues.GET("/tasks", container.QueueAPI.ListTasks)
			queues.POST("/tasks/:task_id/retry", middleware.RequireRole("admin", "operator"), container.QueueAPI.RetryTask)
			queues.POST("/tasks/:task_id/skip", middleware.RequireRole("admin", "operator"), container.QueueAPI.SkipTask)
			queues.POST("/tasks/:task_id/cancel", middleware.RequireRole("admin", "operator"), container.QueueAPI.CancelTask)
			
			// 批量任务操作
			queues.POST("/tasks/batch/retry", middleware.RequireRole("admin", "operator"), container.QueueAPI.BatchRetryTasks)
			queues.POST("/tasks/batch/skip", middleware.RequireRole("admin", "operator"), container.QueueAPI.BatchSkipTasks)
			
			// 任务日志和历史
			queues.GET("/tasks/:task_id/logs", container.QueueAPI.GetTaskLogs)
			queues.GET("/tasks/:task_id/history", container.QueueAPI.GetTaskHistory)
			
			// 任务导出
			queues.GET("/tasks/export", container.QueueAPI.ExportTasks)
			
			// 队列健康和维护
			queues.GET("/health", container.QueueAPI.GetQueueHealth)
			queues.POST("/:queue_name/cleanup", middleware.RequireRole("admin"), container.QueueAPI.CleanupExpiredTasks)
			
			// 队列优化和故障处理
			queues.POST("/:queue_name/optimize", middleware.RequireRole("admin"), container.QueueAPI.OptimizeQueue)
			queues.GET("/:queue_name/recommendations", container.QueueAPI.GetQueueRecommendations)
			queues.POST("/:queue_name/scale", middleware.RequireRole("admin"), container.QueueAPI.ScaleQueue)
			queues.POST("/:queue_name/pause", middleware.RequireRole("admin"), container.QueueAPI.PauseQueue)
			queues.POST("/:queue_name/resume", middleware.RequireRole("admin"), container.QueueAPI.ResumeQueue)
			
			// 队列告警
			queues.GET("/alerts", container.QueueAPI.GetQueueAlerts)
			queues.POST("/alerts/:alert_id/acknowledge", middleware.RequireRole("admin", "operator"), container.QueueAPI.AcknowledgeAlert)
		}
	}

	// Swagger文档路由
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
}
