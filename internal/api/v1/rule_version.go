package v1

import (
	"net/http"
	"strconv"

	"alert_agent/internal/model"
	"alert_agent/internal/pkg/logger"
	"alert_agent/internal/service"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// RuleVersionAPI 规则版本API处理器
type RuleVersionAPI struct {
	ruleService        service.RuleService
	ruleVersionService service.RuleVersionService
}

// NewRuleVersionAPI 创建规则版本API处理器实例
func NewRuleVersionAPI(ruleService service.RuleService, ruleVersionService service.RuleVersionService) *RuleVersionAPI {
	return &RuleVersionAPI{
		ruleService:        ruleService,
		ruleVersionService: ruleVersionService,
	}
}

// GetRuleVersions 获取规则版本列表
// GET /api/v1/rules/{id}/versions
func (r *RuleVersionAPI) GetRuleVersions(c *gin.Context) {
	ruleID := c.Param("id")
	if ruleID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "Rule ID is required",
			"data": nil,
		})
		return
	}

	// 解析分页参数
	page := 1
	pageSize := 10

	if pageStr := c.Query("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	if pageSizeStr := c.Query("page_size"); pageSizeStr != "" {
		if ps, err := strconv.Atoi(pageSizeStr); err == nil && ps > 0 {
			pageSize = ps
		}
	}

	versions, total, err := r.ruleVersionService.GetVersionsByRuleID(c.Request.Context(), ruleID, page, pageSize)
	if err != nil {
		logger.L.Error("Failed to get rule versions", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"msg":  "Failed to get rule versions",
			"data": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "success",
		"data": gin.H{
			"versions":  versions,
			"total":     total,
			"page":      page,
			"page_size": pageSize,
		},
	})
}

// GetRuleVersion 获取指定版本的规则
// GET /api/v1/rules/{id}/versions/{version}
func (r *RuleVersionAPI) GetRuleVersion(c *gin.Context) {
	ruleID := c.Param("id")
	version := c.Param("version")

	if ruleID == "" || version == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "Rule ID and version are required",
			"data": nil,
		})
		return
	}

	ruleVersion, err := r.ruleVersionService.GetVersionByRuleIDAndVersion(c.Request.Context(), ruleID, version)
	if err != nil {
		logger.L.Error("Failed to get rule version", zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{
			"code": 404,
			"msg":  "Rule version not found",
			"data": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "success",
		"data": ruleVersion,
	})
}

// CompareRuleVersions 对比规则版本
// POST /api/v1/rules/versions/compare
func (r *RuleVersionAPI) CompareRuleVersions(c *gin.Context) {
	var req model.RuleVersionCompareRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.L.Error("Failed to bind request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "Invalid request body",
			"data": err.Error(),
		})
		return
	}

	comparison, err := r.ruleVersionService.CompareVersions(c.Request.Context(), &req)
	if err != nil {
		logger.L.Error("Failed to compare rule versions", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"msg":  "Failed to compare rule versions",
			"data": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "success",
		"data": comparison,
	})
}

// RollbackRule 回滚规则到指定版本
// POST /api/v1/rules/{id}/rollback
func (r *RuleVersionAPI) RollbackRule(c *gin.Context) {
	ruleID := c.Param("id")
	if ruleID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "Rule ID is required",
			"data": nil,
		})
		return
	}

	var req model.RuleRollbackRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.L.Error("Failed to bind request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "Invalid request body",
			"data": err.Error(),
		})
		return
	}

	// 确保请求中的规则ID与URL中的一致
	req.RuleID = ruleID

	// 获取用户信息（在实际项目中，这些信息应该从认证中间件中获取）
	userID := c.GetString("user_id")
	userName := c.GetString("user_name")
	if userID == "" {
		userID = "system"
		userName = "System"
	}

	ipAddress := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")

	rule, err := r.ruleVersionService.RollbackRule(c.Request.Context(), &req, userID, userName, ipAddress, userAgent)
	if err != nil {
		logger.L.Error("Failed to rollback rule", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"msg":  "Failed to rollback rule",
			"data": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "success",
		"data": rule,
	})
}

// GetRuleAuditLogs 获取规则审计日志
// GET /api/v1/rules/{id}/audit-logs
func (r *RuleVersionAPI) GetRuleAuditLogs(c *gin.Context) {
	ruleID := c.Param("id")
	if ruleID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "Rule ID is required",
			"data": nil,
		})
		return
	}

	// 解析查询参数
	req := &model.RuleAuditLogListRequest{
		RuleID:   ruleID,
		Action:   c.Query("action"),
		UserID:   c.Query("user_id"),
		Page:     1,
		PageSize: 10,
	}

	if pageStr := c.Query("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			req.Page = p
		}
	}

	if pageSizeStr := c.Query("page_size"); pageSizeStr != "" {
		if ps, err := strconv.Atoi(pageSizeStr); err == nil && ps > 0 {
			req.PageSize = ps
		}
	}

	logs, total, err := r.ruleVersionService.GetAuditLogs(c.Request.Context(), req)
	if err != nil {
		logger.L.Error("Failed to get rule audit logs", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"msg":  "Failed to get rule audit logs",
			"data": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "success",
		"data": gin.H{
			"logs":      logs,
			"total":     total,
			"page":      req.Page,
			"page_size": req.PageSize,
		},
	})
}

// GetAllAuditLogs 获取所有规则的审计日志
// GET /api/v1/rules/audit-logs
func (r *RuleVersionAPI) GetAllAuditLogs(c *gin.Context) {
	// 解析查询参数
	req := &model.RuleAuditLogListRequest{
		RuleID:   c.Query("rule_id"),
		Action:   c.Query("action"),
		UserID:   c.Query("user_id"),
		Page:     1,
		PageSize: 10,
	}

	if pageStr := c.Query("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			req.Page = p
		}
	}

	if pageSizeStr := c.Query("page_size"); pageSizeStr != "" {
		if ps, err := strconv.Atoi(pageSizeStr); err == nil && ps > 0 {
			req.PageSize = ps
		}
	}

	logs, total, err := r.ruleVersionService.GetAuditLogs(c.Request.Context(), req)
	if err != nil {
		logger.L.Error("Failed to get audit logs", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"msg":  "Failed to get audit logs",
			"data": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "success",
		"data": gin.H{
			"logs":      logs,
			"total":     total,
			"page":      req.Page,
			"page_size": req.PageSize,
		},
	})
}

// CreateRuleWithAudit 创建规则（带审计）
// POST /api/v1/rules/audit
func (r *RuleVersionAPI) CreateRuleWithAudit(c *gin.Context) {
	var req model.CreateRuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.L.Error("Failed to bind request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "Invalid request body",
			"data": err.Error(),
		})
		return
	}

	// 获取用户信息
	userID := c.GetString("user_id")
	userName := c.GetString("user_name")
	if userID == "" {
		userID = "system"
		userName = "System"
	}

	ipAddress := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")

	rule, err := r.ruleService.CreateRuleWithAudit(c.Request.Context(), &req, userID, userName, ipAddress, userAgent)
	if err != nil {
		logger.L.Error("Failed to create rule with audit", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"msg":  "Failed to create rule",
			"data": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"code": 201,
		"msg":  "success",
		"data": rule,
	})
}

// UpdateRuleWithAudit 更新规则（带审计）
// PUT /api/v1/rules/{id}/audit
func (r *RuleVersionAPI) UpdateRuleWithAudit(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "Rule ID is required",
			"data": nil,
		})
		return
	}

	var req model.UpdateRuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.L.Error("Failed to bind request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "Invalid request body",
			"data": err.Error(),
		})
		return
	}

	// 获取用户信息
	userID := c.GetString("user_id")
	userName := c.GetString("user_name")
	if userID == "" {
		userID = "system"
		userName = "System"
	}

	ipAddress := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")

	rule, err := r.ruleService.UpdateRuleWithAudit(c.Request.Context(), id, &req, userID, userName, ipAddress, userAgent)
	if err != nil {
		logger.L.Error("Failed to update rule with audit", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"msg":  "Failed to update rule",
			"data": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "success",
		"data": rule,
	})
}

// DeleteRuleWithAudit 删除规则（带审计）
// DELETE /api/v1/rules/{id}/audit
func (r *RuleVersionAPI) DeleteRuleWithAudit(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "Rule ID is required",
			"data": nil,
		})
		return
	}

	// 获取用户信息
	userID := c.GetString("user_id")
	userName := c.GetString("user_name")
	if userID == "" {
		userID = "system"
		userName = "System"
	}

	ipAddress := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")

	err := r.ruleService.DeleteRuleWithAudit(c.Request.Context(), id, userID, userName, ipAddress, userAgent)
	if err != nil {
		logger.L.Error("Failed to delete rule with audit", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"msg":  "Failed to delete rule",
			"data": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "success",
		"data": nil,
	})
}