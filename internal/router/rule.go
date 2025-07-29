package router

import (
	"github.com/gin-gonic/gin"
	"alert_agent/internal/handler"
)

// RegisterRuleRoutes 注册规则相关路由
func RegisterRuleRoutes(r *gin.Engine, ruleHandler *handler.RuleHandler) {
	api := r.Group("/api/v1")
	SetupRuleRoutes(api, ruleHandler)
}

// SetupRuleRoutes 设置规则相关路由
func SetupRuleRoutes(r *gin.RouterGroup, ruleHandler *handler.RuleHandler) {
	// 规则管理路由
	rules := r.Group("/rules")
	{
		rules.POST("", ruleHandler.CreateRule)                    // 创建规则
		rules.GET("", ruleHandler.ListRules)                     // 列出规则
		rules.GET("/:id", ruleHandler.GetRule)                   // 获取规则
		rules.PUT("/:id", ruleHandler.UpdateRule)                // 更新规则
		rules.DELETE("/:id", ruleHandler.DeleteRule)             // 删除规则
		rules.POST("/validate", ruleHandler.ValidateRule)        // 验证规则
		rules.POST("/:id/distribute", ruleHandler.DistributeRule) // 分发规则
		rules.GET("/:id/distribution", ruleHandler.GetDistributionStatus) // 获取分发状态
	}

	// 规则组管理路由
	ruleGroups := r.Group("/rule-groups")
	{
		ruleGroups.POST("", ruleHandler.CreateRuleGroup)         // 创建规则组
		ruleGroups.GET("", ruleHandler.ListRuleGroups)          // 列出规则组
		ruleGroups.GET("/:id", ruleHandler.GetRuleGroup)        // 获取规则组
		ruleGroups.PUT("/:id", ruleHandler.UpdateRuleGroup)     // 更新规则组
		ruleGroups.DELETE("/:id", ruleHandler.DeleteRuleGroup)  // 删除规则组
	}

	// 规则分发管理路由
	distributions := r.Group("/distributions")
	{
		distributions.GET("", ruleHandler.ListDistributions)     // 列出分发记录
	}

	// 集群同步路由
	clusters := r.Group("/clusters")
	{
		clusters.POST("/:cluster_id/sync", ruleHandler.SyncRulesToCluster) // 同步规则到集群
		clusters.GET("/:cluster_id/sync-status", ruleHandler.GetSyncStatus) // 获取同步状态
	}

	// 冲突管理路由
	conflicts := r.Group("/conflicts")
	{
		conflicts.GET("", ruleHandler.ListConflicts)            // 列出冲突
		conflicts.GET("/detect", ruleHandler.DetectConflicts)   // 检测冲突
		conflicts.PUT("/:id/resolve", ruleHandler.ResolveConflict) // 解决冲突
	}

	// 统计信息路由
	stats := r.Group("/stats")
	{
		stats.GET("/rules", ruleHandler.GetRuleStats)           // 获取规则统计
	}
}