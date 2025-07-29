package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"alert_agent/internal/domain/rule"
	"alert_agent/pkg/types"
)

// RuleHandler 规则处理器
type RuleHandler struct {
	ruleService rule.Service
}

// NewRuleHandler 创建规则处理器
func NewRuleHandler(ruleService rule.Service) *RuleHandler {
	return &RuleHandler{
		ruleService: ruleService,
	}
}

// CreateRule 创建规则
func (h *RuleHandler) CreateRule(c *gin.Context) {
	var req rule.PrometheusRule
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	createdRule, err := h.ruleService.CreateRule(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": createdRule})
}

// UpdateRule 更新规则
func (h *RuleHandler) UpdateRule(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid rule id"})
		return
	}

	var req rule.PrometheusRule
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updatedRule, err := h.ruleService.UpdateRule(c.Request.Context(), uint(id), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": updatedRule})
}

// DeleteRule 删除规则
func (h *RuleHandler) DeleteRule(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid rule id"})
		return
	}

	if err := h.ruleService.DeleteRule(c.Request.Context(), uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "rule deleted successfully"})
}

// GetRule 获取规则
func (h *RuleHandler) GetRule(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid rule id"})
		return
	}

	rule, err := h.ruleService.GetRule(c.Request.Context(), uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": rule})
}

// ListRules 列出规则
func (h *RuleHandler) ListRules(c *gin.Context) {
	query := &types.Query{
		Limit:  20,
		Offset: 0,
		Filter: make(map[string]interface{}),
	}

	// 解析查询参数
	if limitStr := c.Query("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 {
			query.Limit = limit
		}
	}

	if offsetStr := c.Query("offset"); offsetStr != "" {
		if offset, err := strconv.Atoi(offsetStr); err == nil && offset >= 0 {
			query.Offset = offset
		}
	}

	if clusterID := c.Query("cluster_id"); clusterID != "" {
		query.Filter["cluster_id"] = clusterID
	}

	if groupName := c.Query("group_name"); groupName != "" {
		query.Filter["group_name"] = groupName
	}

	if search := c.Query("search"); search != "" {
		query.Search = search
	}

	result, err := h.ruleService.ListRules(c.Request.Context(), query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// ValidateRule 验证规则
func (h *RuleHandler) ValidateRule(c *gin.Context) {
	var req rule.PrometheusRule
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.ruleService.ValidateRule(c.Request.Context(), &req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "rule is valid"})
}

// DistributeRule 分发规则
func (h *RuleHandler) DistributeRule(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid rule id"})
		return
	}

	var req struct {
		ClusterIDs []string `json:"cluster_ids" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.ruleService.DistributeRule(c.Request.Context(), uint(id), req.ClusterIDs); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "rule distributed successfully"})
}

// GetDistributionStatus 获取分发状态
func (h *RuleHandler) GetDistributionStatus(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid rule id"})
		return
	}

	clusterID := c.Query("cluster_id")
	if clusterID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cluster_id is required"})
		return
	}

	status, err := h.ruleService.GetDistributionStatus(c.Request.Context(), uint(id), clusterID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": status})
}

// ListDistributions 列出分发记录
func (h *RuleHandler) ListDistributions(c *gin.Context) {
	query := &types.Query{
		Limit:  20,
		Offset: 0,
		Filter: make(map[string]interface{}),
	}

	// 解析查询参数
	if limitStr := c.Query("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 {
			query.Limit = limit
		}
	}

	if offsetStr := c.Query("offset"); offsetStr != "" {
		if offset, err := strconv.Atoi(offsetStr); err == nil && offset >= 0 {
			query.Offset = offset
		}
	}

	if clusterID := c.Query("cluster_id"); clusterID != "" {
		query.Filter["cluster_id"] = clusterID
	}

	if status := c.Query("status"); status != "" {
		query.Filter["status"] = status
	}

	result, err := h.ruleService.ListDistributions(c.Request.Context(), query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// SyncRulesToCluster 同步规则到集群
func (h *RuleHandler) SyncRulesToCluster(c *gin.Context) {
	clusterID := c.Param("cluster_id")
	if clusterID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cluster_id is required"})
		return
	}

	if err := h.ruleService.SyncRulesToCluster(c.Request.Context(), clusterID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "rules synced successfully"})
}

// GetSyncStatus 获取同步状态
func (h *RuleHandler) GetSyncStatus(c *gin.Context) {
	clusterID := c.Param("cluster_id")
	if clusterID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cluster_id is required"})
		return
	}

	status, err := h.ruleService.GetSyncStatus(c.Request.Context(), clusterID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": status})
}

// GetRuleStats 获取规则统计
func (h *RuleHandler) GetRuleStats(c *gin.Context) {
	clusterID := c.Query("cluster_id")
	if clusterID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cluster_id is required"})
		return
	}

	stats, err := h.ruleService.GetRuleStats(c.Request.Context(), clusterID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": stats})
}

// CreateRuleGroup 创建规则组
func (h *RuleHandler) CreateRuleGroup(c *gin.Context) {
	var req rule.RuleGroup
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	createdGroup, err := h.ruleService.CreateRuleGroup(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": createdGroup})
}

// UpdateRuleGroup 更新规则组
func (h *RuleHandler) UpdateRuleGroup(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid group id"})
		return
	}

	var req rule.RuleGroup
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updatedGroup, err := h.ruleService.UpdateRuleGroup(c.Request.Context(), uint(id), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": updatedGroup})
}

// DeleteRuleGroup 删除规则组
func (h *RuleHandler) DeleteRuleGroup(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid group id"})
		return
	}

	if err := h.ruleService.DeleteRuleGroup(c.Request.Context(), uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "rule group deleted successfully"})
}

// GetRuleGroup 获取规则组
func (h *RuleHandler) GetRuleGroup(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid group id"})
		return
	}

	group, err := h.ruleService.GetRuleGroup(c.Request.Context(), uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": group})
}

// ListRuleGroups 列出规则组
func (h *RuleHandler) ListRuleGroups(c *gin.Context) {
	query := &types.Query{
		Limit:  20,
		Offset: 0,
		Filter: make(map[string]interface{}),
	}

	// 解析查询参数
	if limitStr := c.Query("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 {
			query.Limit = limit
		}
	}

	if offsetStr := c.Query("offset"); offsetStr != "" {
		if offset, err := strconv.Atoi(offsetStr); err == nil && offset >= 0 {
			query.Offset = offset
		}
	}

	if clusterID := c.Query("cluster_id"); clusterID != "" {
		query.Filter["cluster_id"] = clusterID
	}

	if search := c.Query("search"); search != "" {
		query.Search = search
	}

	result, err := h.ruleService.ListRuleGroups(c.Request.Context(), query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// DetectConflicts 检测冲突
func (h *RuleHandler) DetectConflicts(c *gin.Context) {
	clusterID := c.Query("cluster_id")
	if clusterID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cluster_id is required"})
		return
	}

	conflicts, err := h.ruleService.DetectConflicts(c.Request.Context(), clusterID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": conflicts})
}

// ListConflicts 列出冲突
func (h *RuleHandler) ListConflicts(c *gin.Context) {
	query := &types.Query{
		Limit:  20,
		Offset: 0,
		Filter: make(map[string]interface{}),
	}

	// 解析查询参数
	if limitStr := c.Query("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 {
			query.Limit = limit
		}
	}

	if offsetStr := c.Query("offset"); offsetStr != "" {
		if offset, err := strconv.Atoi(offsetStr); err == nil && offset >= 0 {
			query.Offset = offset
		}
	}

	if clusterID := c.Query("cluster_id"); clusterID != "" {
		query.Filter["cluster_id"] = clusterID
	}

	if resolved := c.Query("resolved"); resolved != "" {
		query.Filter["resolved"] = resolved == "true"
	}

	result, err := h.ruleService.ListConflicts(c.Request.Context(), query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// ResolveConflict 解决冲突
func (h *RuleHandler) ResolveConflict(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid conflict id"})
		return
	}

	var req struct {
		ResolvedBy string `json:"resolved_by" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.ruleService.ResolveConflict(c.Request.Context(), uint(id), req.ResolvedBy); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "conflict resolved successfully"})
}