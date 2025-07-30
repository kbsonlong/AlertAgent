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

// RuleAPI 规则API处理器
type RuleAPI struct {
	ruleService         service.RuleService
	distributionService service.RuleDistributionService
}

// NewRuleAPI 创建规则API处理器实例
func NewRuleAPI(ruleService service.RuleService, distributionService service.RuleDistributionService) *RuleAPI {
	return &RuleAPI{
		ruleService:         ruleService,
		distributionService: distributionService,
	}
}

// CreateRule 创建告警规则
// POST /api/v1/rules
func (r *RuleAPI) CreateRule(c *gin.Context) {
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

	rule, err := r.ruleService.CreateRule(c.Request.Context(), &req)
	if err != nil {
		logger.L.Error("Failed to create rule", zap.Error(err))
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

// GetRule 获取单个告警规则
// GET /api/v1/rules/{id}
func (r *RuleAPI) GetRule(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "Rule ID is required",
			"data": nil,
		})
		return
	}

	rule, err := r.ruleService.GetRule(c.Request.Context(), id)
	if err != nil {
		logger.L.Error("Failed to get rule", zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{
			"code": 404,
			"msg":  "Rule not found",
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

// UpdateRule 更新告警规则
// PUT /api/v1/rules/{id}
func (r *RuleAPI) UpdateRule(c *gin.Context) {
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

	rule, err := r.ruleService.UpdateRule(c.Request.Context(), id, &req)
	if err != nil {
		logger.L.Error("Failed to update rule", zap.Error(err))
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

// DeleteRule 删除告警规则
// DELETE /api/v1/rules/{id}
func (r *RuleAPI) DeleteRule(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "Rule ID is required",
			"data": nil,
		})
		return
	}

	err := r.ruleService.DeleteRule(c.Request.Context(), id)
	if err != nil {
		logger.L.Error("Failed to delete rule", zap.Error(err))
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

// ListRules 获取告警规则列表
// GET /api/v1/rules
func (r *RuleAPI) ListRules(c *gin.Context) {
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

	rules, total, err := r.ruleService.ListRules(c.Request.Context(), page, pageSize)
	if err != nil {
		logger.L.Error("Failed to list rules", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"msg":  "Failed to list rules",
			"data": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "success",
		"data": gin.H{
			"rules":     rules,
			"total":     total,
			"page":      page,
			"page_size": pageSize,
		},
	})
}

// GetRuleDistribution 获取规则分发状态
// GET /api/v1/rules/{id}/distribution
func (r *RuleAPI) GetRuleDistribution(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "Rule ID is required",
			"data": nil,
		})
		return
	}

	status, err := r.distributionService.GetDistributionStatus(c.Request.Context(), id)
	if err != nil {
		logger.L.Error("Failed to get rule distribution status", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"msg":  "Failed to get distribution status",
			"data": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "success",
		"data": status,
	})
}

// ValidateRule 验证规则语法
// POST /api/v1/rules/validate
func (r *RuleAPI) ValidateRule(c *gin.Context) {
	var req struct {
		Expression string `json:"expression" binding:"required"`
		Duration   string `json:"duration" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		logger.L.Error("Failed to bind request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "Invalid request body",
			"data": err.Error(),
		})
		return
	}

	err := r.ruleService.ValidateRule(c.Request.Context(), req.Expression, req.Duration)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "Rule validation failed",
			"data": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "Rule validation passed",
		"data": nil,
	})
}

// Legacy functions for backward compatibility
// These will be deprecated once the frontend is updated

// ListRules 获取告警规则列表 (Legacy)
func ListRules(c *gin.Context) {
	// This is a placeholder for backward compatibility
	// The actual implementation should use the new RuleAPI
	c.JSON(http.StatusNotImplemented, gin.H{
		"code": 501,
		"msg":  "This endpoint is deprecated, please use the new RuleAPI",
		"data": nil,
	})
}

// CreateRule 创建告警规则 (Legacy)
func CreateRule(c *gin.Context) {
	// This is a placeholder for backward compatibility
	// The actual implementation should use the new RuleAPI
	c.JSON(http.StatusNotImplemented, gin.H{
		"code": 501,
		"msg":  "This endpoint is deprecated, please use the new RuleAPI",
		"data": nil,
	})
}

// GetRule 获取单个告警规则 (Legacy)
func GetRule(c *gin.Context) {
	// This is a placeholder for backward compatibility
	// The actual implementation should use the new RuleAPI
	c.JSON(http.StatusNotImplemented, gin.H{
		"code": 501,
		"msg":  "This endpoint is deprecated, please use the new RuleAPI",
		"data": nil,
	})
}

// UpdateRule 更新告警规则 (Legacy)
func UpdateRule(c *gin.Context) {
	// This is a placeholder for backward compatibility
	// The actual implementation should use the new RuleAPI
	c.JSON(http.StatusNotImplemented, gin.H{
		"code": 501,
		"msg":  "This endpoint is deprecated, please use the new RuleAPI",
		"data": nil,
	})
}

// DeleteRule 删除告警规则 (Legacy)
func DeleteRule(c *gin.Context) {
	// This is a placeholder for backward compatibility
	// The actual implementation should use the new RuleAPI
	c.JSON(http.StatusNotImplemented, gin.H{
		"code": 501,
		"msg":  "This endpoint is deprecated, please use the new RuleAPI",
		"data": nil,
	})
}
// GetDistributionSummary 获取多个规则的分发汇总
// POST /api/v1/rules/distribution/summary
func (r *RuleAPI) GetDistributionSummary(c *gin.Context) {
	var req struct {
		RuleIDs []string `json:"rule_ids"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		logger.L.Error("Failed to bind request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "Invalid request body",
			"data": err.Error(),
		})
		return
	}

	summaries, err := r.distributionService.GetDistributionSummary(c.Request.Context(), req.RuleIDs)
	if err != nil {
		logger.L.Error("Failed to get distribution summary", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"msg":  "Failed to get distribution summary",
			"data": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "success",
		"data": summaries,
	})
}

// BatchRuleOperation 批量规则操作
// POST /api/v1/rules/batch
func (r *RuleAPI) BatchRuleOperation(c *gin.Context) {
	var req model.BatchRuleOperation
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.L.Error("Failed to bind request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "Invalid request body",
			"data": err.Error(),
		})
		return
	}

	result, err := r.distributionService.BatchDistributeRules(c.Request.Context(), &req)
	if err != nil {
		logger.L.Error("Failed to execute batch rule operation", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"msg":  "Failed to execute batch operation",
			"data": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "success",
		"data": result,
	})
}

// RetryDistribution 重试分发
// POST /api/v1/rules/distribution/retry
func (r *RuleAPI) RetryDistribution(c *gin.Context) {
	var req model.RetryDistributionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.L.Error("Failed to bind request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "Invalid request body",
			"data": err.Error(),
		})
		return
	}

	result, err := r.distributionService.RetryFailedDistributions(c.Request.Context(), &req)
	if err != nil {
		logger.L.Error("Failed to retry distribution", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"msg":  "Failed to retry distribution",
			"data": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "success",
		"data": result,
	})
}

// GetTargetDistribution 获取特定目标的分发信息
// GET /api/v1/rules/{id}/distribution/{target}
func (r *RuleAPI) GetTargetDistribution(c *gin.Context) {
	ruleID := c.Param("id")
	target := c.Param("target")
	
	if ruleID == "" || target == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "Rule ID and target are required",
			"data": nil,
		})
		return
	}

	info, err := r.distributionService.GetTargetDistributionInfo(c.Request.Context(), ruleID, target)
	if err != nil {
		logger.L.Error("Failed to get target distribution info", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"msg":  "Failed to get target distribution info",
			"data": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "success",
		"data": info,
	})
}

// UpdateDistributionStatus 更新分发状态
// PUT /api/v1/rules/distribution/status
func (r *RuleAPI) UpdateDistributionStatus(c *gin.Context) {
	var req struct {
		RuleIDs []string `json:"rule_ids" binding:"required"`
		Targets []string `json:"targets"`
		Status  string   `json:"status" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		logger.L.Error("Failed to bind request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "Invalid request body",
			"data": err.Error(),
		})
		return
	}

	// 验证状态值
	validStatuses := map[string]bool{
		"pending": true,
		"success": true,
		"failed":  true,
	}
	if !validStatuses[req.Status] {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "Invalid status value",
			"data": "Status must be one of: pending, success, failed",
		})
		return
	}

	err := r.distributionService.BatchUpdateDistributionStatus(c.Request.Context(), req.RuleIDs, req.Targets, req.Status)
	if err != nil {
		logger.L.Error("Failed to update distribution status", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"msg":  "Failed to update distribution status",
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

// GetRetryableDistributions 获取可重试的分发记录
// GET /api/v1/rules/distribution/retryable
func (r *RuleAPI) GetRetryableDistributions(c *gin.Context) {
	limit := 50 // 默认限制
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 200 {
			limit = l
		}
	}

	records, err := r.distributionService.GetRetryableDistributions(c.Request.Context(), limit)
	if err != nil {
		logger.L.Error("Failed to get retryable distributions", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"msg":  "Failed to get retryable distributions",
			"data": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "success",
		"data": gin.H{
			"records": records,
			"count":   len(records),
		},
	})
}

// ProcessRetryableDistributions 处理可重试的分发记录
// POST /api/v1/rules/distribution/process-retry
func (r *RuleAPI) ProcessRetryableDistributions(c *gin.Context) {
	err := r.distributionService.ProcessRetryableDistributions(c.Request.Context())
	if err != nil {
		logger.L.Error("Failed to process retryable distributions", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"msg":  "Failed to process retryable distributions",
			"data": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "success",
		"data": "Retryable distributions processed successfully",
	})
}