package http

import (
	"alert_agent/internal/application/analysis"
	"alert_agent/internal/domain/channel"
	"alert_agent/internal/domain/cluster"
	domainAnalysis "alert_agent/internal/domain/analysis"
	"alert_agent/internal/security/di"
	"alert_agent/internal/security/routes"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Router HTTP路由器
type Router struct {
	clusterHandler    *ClusterHandler
	channelHandler    *ChannelHandler
	pluginHandler     *PluginHandler
	analysisHandler   *AnalysisHandler
	n8nService        *analysis.N8NAnalysisService
	workflowManager   domainAnalysis.N8NWorkflowManager
	securityContainer *di.Container
	logger            *zap.Logger
}

// NewRouter 创建路由器
func NewRouter(
	clusterService cluster.Service,
	channelService channel.Service,
	channelManager channel.ChannelManager,
	analysisService domainAnalysis.AnalysisService,
	n8nService *analysis.N8NAnalysisService,
	workflowManager domainAnalysis.N8NWorkflowManager,
	securityContainer *di.Container,
	logger *zap.Logger,
) *Router {
	return &Router{
		clusterHandler:    NewClusterHandler(clusterService, logger),
		channelHandler:    NewChannelHandler(channelService, logger),
		pluginHandler:     NewPluginHandler(channelManager, logger),
		analysisHandler:   NewAnalysisHandler(analysisService),
		n8nService:        n8nService,
		workflowManager:   workflowManager,
		securityContainer: securityContainer,
		logger:            logger,
	}
}

// SetupRoutes 设置路由
func (r *Router) SetupRoutes(engine *gin.Engine) {
	// 设置全局安全中间件
	routes.SetupSecurityMiddleware(engine)
	
	// 健康检查
	engine.GET("/health", r.healthCheck)
	
	// 设置安全相关路由（认证、用户管理等）
	routes.SetupAuthRoutes(engine, r.securityContainer.GetAuthHandler(), r.securityContainer.GetMiddlewareConfig())
	
	// 设置健康检查路由
	routes.SetupHealthRoutes(engine)

	// API v1 路由组
	v1 := engine.Group("/api/v1")
	{
		// 集群管理路由
		clusters := v1.Group("/clusters")
		{
			clusters.POST("", r.clusterHandler.CreateCluster)
			clusters.GET("", r.clusterHandler.ListClusters)
			clusters.GET("/:id", r.clusterHandler.GetCluster)
			clusters.PUT("/:id", r.clusterHandler.UpdateCluster)
			clusters.DELETE("/:id", r.clusterHandler.DeleteCluster)
			clusters.POST("/:id/test", r.clusterHandler.TestClusterConnection)
			clusters.GET("/:id/health", r.clusterHandler.GetClusterHealth)
			clusters.GET("/:id/metrics", r.clusterHandler.GetClusterMetrics)
		}

		// 通道管理路由
		channels := v1.Group("/channels")
		{
			// 基础CRUD操作
			channels.POST("", r.channelHandler.CreateChannel)
			channels.GET("", r.channelHandler.ListChannels)
			channels.GET("/:id", r.channelHandler.GetChannel)
			channels.PUT("/:id", r.channelHandler.UpdateChannel)
			channels.DELETE("/:id", r.channelHandler.DeleteChannel)
			
			// 通道操作
			channels.POST("/:id/test", r.channelHandler.TestChannel)
			channels.POST("/:id/send", r.channelHandler.SendMessage)
			
			// 健康检查和监控
			channels.GET("/:id/health", r.channelHandler.GetChannelHealth)
			channels.GET("/:id/stats", r.channelHandler.GetChannelStats)
			channels.POST("/health/batch", r.channelHandler.BatchHealthCheck)
		}
		
		// 插件管理路由
		plugins := v1.Group("/plugins")
		{
			plugins.GET("", r.pluginHandler.ListPlugins)
			plugins.GET("/:type", r.pluginHandler.GetPlugin)
			plugins.POST("/:type/validate", r.pluginHandler.ValidatePluginConfig)
			plugins.POST("/:type/test", r.pluginHandler.TestPluginConnection)
		}

		// 分析管理路由
		analysis := v1.Group("/analysis")
		{
			// 任务管理
			analysis.POST("/submit", r.analysisHandler.SubmitAnalysis)
			analysis.GET("/result/:id", r.analysisHandler.GetAnalysisResult)
			analysis.GET("/progress/:id", r.analysisHandler.GetAnalysisProgress)
			analysis.DELETE("/cancel/:id", r.analysisHandler.CancelAnalysis)
			analysis.POST("/retry/:id", r.analysisHandler.RetryAnalysis)
			
			// 任务列表和统计
			analysis.GET("/tasks", r.analysisHandler.ListAnalysisTasks)
			analysis.GET("/statistics", r.analysisHandler.GetAnalysisStatistics)
			
			// 队列和工作器状态
			analysis.GET("/queue/status", r.analysisHandler.GetQueueStatus)
			analysis.GET("/workers/status", r.analysisHandler.GetWorkerStatuses)
			
			// 健康检查
			analysis.GET("/health", r.analysisHandler.HealthCheck)
		}

		// n8n 分析路由
		n8n := v1.Group("/n8n")
		{
			// 注册 n8n 分析路由
			RegisterN8NRoutes(n8n, r.n8nService)
		}

		// n8n 回调路由
		callbacks := v1.Group("/callbacks")
		{
			// 注册 n8n 回调路由
			RegisterN8NCallbackRoutes(callbacks, r.workflowManager)
		}
	}
}

// healthCheck 健康检查
func (r *Router) healthCheck(c *gin.Context) {
	c.JSON(200, gin.H{
		"status":  "ok",
		"message": "Alert Agent is running",
	})
}