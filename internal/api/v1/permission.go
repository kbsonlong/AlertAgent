package v1

import (
	"strconv"

	"alert_agent/internal/model"
	"alert_agent/internal/pkg/logger"
	"alert_agent/internal/pkg/response"
	"alert_agent/internal/service"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// PermissionController 权限控制器
type PermissionController struct {
	permissionService service.PermissionService
}

// NewPermissionController 创建权限控制器
func NewPermissionController(permissionService service.PermissionService) *PermissionController {
	return &PermissionController{
		permissionService: permissionService,
	}
}

// PermissionRequest 权限请求结构
type PermissionRequest struct {
	Name        string `json:"name" binding:"required,min=2,max=100"`
	Code        string `json:"code" binding:"required,min=2,max=100"`
	Resource    string `json:"resource" binding:"required,min=2,max=100"`
	Action      string `json:"action" binding:"required,min=2,max=50"`
	Description string `json:"description" binding:"max=500"`
	Category    string `json:"category" binding:"required,oneof=user role permission alert system"`
	Type        string `json:"type" binding:"oneof=system custom"`
	Status      string `json:"status" binding:"oneof=active inactive"`
}

// PermissionResponse 权限响应结构
type PermissionResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Code        string `json:"code"`
	Resource    string `json:"resource"`
	Action      string `json:"action"`
	Description string `json:"description"`
	Category    string `json:"category"`
	Type        string `json:"type"`
	Status      string `json:"status"`
	IsSystem    bool   `json:"is_system"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
	CreatedBy   string `json:"created_by"`
	UpdatedBy   string `json:"updated_by"`
}

// PermissionListResponse 权限列表响应
type PermissionListResponse struct {
	Permissions []*PermissionResponse `json:"permissions"`
	Total       int64                 `json:"total"`
	Page        int                   `json:"page"`
	PageSize    int                   `json:"page_size"`
}

// PermissionStatsResponse 权限统计响应
type PermissionStatsResponse struct {
	Total  int64 `json:"total"`
	System int64 `json:"system"`
	Custom int64 `json:"custom"`
	Active int64 `json:"active"`
}

// ListPermissions 获取权限列表
// ListPermissions 获取权限列表
// @Summary 获取权限列表
// @Description 获取权限列表，支持分页和筛选
// @Tags 权限管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(20)
// @Param name query string false "权限名称"
// @Param code query string false "权限代码"
// @Param resource query string false "资源"
// @Param action query string false "操作"
// @Param category query string false "分类" Enums(user,role,permission,alert,system)
// @Param type query string false "类型" Enums(system,custom)
// @Param status query string false "状态" Enums(active,inactive)
// @Success 200 {object} response.Response{data=PermissionListResponse}
// @Failure 401 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/permissions [get]
func (c *PermissionController) ListPermissions(ctx *gin.Context) {
	// 解析查询参数
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(ctx.DefaultQuery("page_size", "20"))
	
	req := &service.ListPermissionsRequest{
		Page:     page,
		PageSize: pageSize,
		Name:     ctx.Query("name"),
		Code:     ctx.Query("code"),
		Resource: ctx.Query("resource"),
		Action:   ctx.Query("action"),
		Category: ctx.Query("category"),
		Type:     ctx.Query("type"),
		Status:   ctx.Query("status"),
	}
	
	// 调用服务
	resp, err := c.permissionService.ListPermissions(ctx.Request.Context(), req)
	if err != nil {
		logger.L.Error("Failed to list permissions", zap.Error(err))
		response.InternalServerError(ctx, "获取权限列表失败", err)
		return
	}
	
	// 转换响应
	permissions := make([]*PermissionResponse, len(resp.Permissions))
	for i, permission := range resp.Permissions {
		permissions[i] = c.convertToPermissionResponse(permission)
	}
	
	response.Pagination(ctx, PermissionListResponse{
		Permissions: permissions,
		Total:       resp.Total,
		Page:        resp.Page,
		PageSize:    resp.PageSize,
	}, resp.Total, resp.Page, resp.PageSize)
}

// GetPermission 获取权限详情
// @Summary 获取权限详情
// @Description 根据ID获取权限详细信息
// @Tags 权限管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "权限ID"
// @Success 200 {object} response.Response{data=PermissionResponse}
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/permissions/{id} [get]
func (c *PermissionController) GetPermission(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" {
		response.BadRequest(ctx, "权限ID不能为空", nil)
		return
	}
	
	permission, err := c.permissionService.GetPermission(ctx.Request.Context(), id)
	if err != nil {
		logger.L.Error("Failed to get permission", zap.String("id", id), zap.Error(err))
		response.NotFound(ctx, "权限不存在", err)
		return
	}
	
	response.Success(ctx, c.convertToPermissionResponse(permission))
}

// CreatePermission 创建权限
// @Summary 创建权限
// @Description 创建新的权限
// @Tags 权限管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body PermissionRequest true "权限信息"
// @Success 201 {object} response.Response{data=PermissionResponse}
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/permissions [post]
func (c *PermissionController) CreatePermission(ctx *gin.Context) {
	var req PermissionRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.BadRequest(ctx, "请求参数无效", err)
		return
	}
	
	// 获取当前用户ID
	createdBy := ctx.GetString("user_id")
	if createdBy == "" {
		createdBy = "system"
	}
	
	createReq := &service.CreatePermissionRequest{
		Name:        req.Name,
		Code:        req.Code,
		Resource:    req.Resource,
		Action:      req.Action,
		Description: req.Description,
		Category:    req.Category,
		Type:        req.Type,
		Status:      req.Status,
		CreatedBy:   createdBy,
	}
	
	permission, err := c.permissionService.CreatePermission(ctx.Request.Context(), createReq)
	if err != nil {
		logger.L.Error("Failed to create permission", zap.Error(err))
		response.InternalServerError(ctx, "创建权限失败", err)
		return
	}
	
	response.Created(ctx, c.convertToPermissionResponse(permission))
}

// UpdatePermission 更新权限
// @Summary 更新权限
// @Description 更新权限信息
// @Tags 权限管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "权限ID"
// @Param request body PermissionRequest true "权限信息"
// @Success 200 {object} response.Response{data=PermissionResponse}
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/permissions/{id} [put]
func (c *PermissionController) UpdatePermission(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" {
		response.BadRequest(ctx, "权限ID不能为空", nil)
		return
	}
	
	var req PermissionRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.BadRequest(ctx, "请求参数无效", err)
		return
	}
	
	// 获取当前用户ID
	updatedBy := ctx.GetString("user_id")
	if updatedBy == "" {
		updatedBy = "system"
	}
	
	updateReq := &service.UpdatePermissionRequest{
		Name:        &req.Name,
		Description: &req.Description,
		Status:      &req.Status,
		UpdatedBy:   &updatedBy,
	}
	
	permission, err := c.permissionService.UpdatePermission(ctx.Request.Context(), id, updateReq)
	if err != nil {
		logger.L.Error("Failed to update permission", zap.String("id", id), zap.Error(err))
		response.InternalServerError(ctx, "更新权限失败", err)
		return
	}
	
	response.Updated(ctx, c.convertToPermissionResponse(permission))
}

// DeletePermission 删除权限
// @Summary 删除权限
// @Description 删除指定权限
// @Tags 权限管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "权限ID"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/permissions/{id} [delete]
func (c *PermissionController) DeletePermission(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" {
		response.BadRequest(ctx, "权限ID不能为空", nil)
		return
	}
	
	if err := c.permissionService.DeletePermission(ctx.Request.Context(), id); err != nil {
		logger.L.Error("Failed to delete permission", zap.String("id", id), zap.Error(err))
		response.InternalServerError(ctx, "删除权限失败", err)
		return
	}
	
	response.Deleted(ctx)
}

// GetPermissionStats 获取权限统计
// @Summary 获取权限统计信息
// @Description 获取权限统计数据，包括总数、系统权限数、自定义权限数、活跃权限数
// @Tags 权限管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.Response{data=PermissionStatsResponse}
// @Failure 401 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/permissions/stats [get]
func (c *PermissionController) GetPermissionStats(ctx *gin.Context) {
	stats, err := c.permissionService.GetPermissionStats(ctx.Request.Context())
	if err != nil {
		logger.L.Error("Failed to get permission stats", zap.Error(err))
		response.InternalServerError(ctx, "获取权限统计失败", err)
		return
	}
	
	response.Success(ctx, PermissionStatsResponse{
		Total:  stats.Total,
		System: stats.System,
		Custom: stats.Custom,
		Active: stats.Active,
	})
}

// InitializeSystemPermissions 初始化系统权限
// @Summary 初始化系统权限
// @Description 初始化系统预定义权限
// @Tags 权限管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.Response
// @Failure 401 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/permissions/initialize [post]
func (c *PermissionController) InitializeSystemPermissions(ctx *gin.Context) {
	if err := c.permissionService.InitializeSystemPermissions(ctx.Request.Context()); err != nil {
		logger.L.Error("Failed to initialize system permissions", zap.Error(err))
		response.InternalServerError(ctx, "初始化系统权限失败", err)
		return
	}
	
	response.SuccessWithMessage(ctx, "系统权限初始化成功", nil)
}

// convertToPermissionResponse 转换为权限响应结构
func (c *PermissionController) convertToPermissionResponse(permission *model.Permission) *PermissionResponse {
	return &PermissionResponse{
		ID:          permission.ID,
		Name:        permission.Name,
		Code:        permission.Code,
		Resource:    permission.Resource,
		Action:      permission.Action,
		Description: permission.Description,
		Category:    permission.Category,
		Type:        permission.Type,
		Status:      permission.Status,
		IsSystem:    permission.IsSystem,
		CreatedAt:   permission.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:   permission.UpdatedAt.Format("2006-01-02 15:04:05"),
		CreatedBy:   permission.CreatedBy,
		UpdatedBy:   permission.UpdatedBy,
	}
}