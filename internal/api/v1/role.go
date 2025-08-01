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

// RoleController 角色控制器
type RoleController struct {
	roleService service.RoleService
}

// NewRoleController 创建角色控制器
func NewRoleController(roleService service.RoleService) *RoleController {
	return &RoleController{
		roleService: roleService,
	}
}

// RoleRequest 角色请求结构
type RoleRequest struct {
	Name        string   `json:"name" binding:"required,min=2,max=100"`
	Code        string   `json:"code" binding:"required,min=2,max=100"`
	Description string   `json:"description" binding:"max=500"`
	Type        string   `json:"type" binding:"oneof=system custom"`
	Status      string   `json:"status" binding:"oneof=active inactive"`
	Permissions []string `json:"permissions"` // 权限ID列表
}

// RoleResponse 角色响应结构
type RoleResponse struct {
	ID          string                    `json:"id"`
	Name        string                    `json:"name"`
	Code        string                    `json:"code"`
	Description string                    `json:"description"`
	Type        string                    `json:"type"`
	Status      string                    `json:"status"`
	IsSystem    bool                      `json:"is_system"`
	Permissions []*PermissionSimpleResponse `json:"permissions,omitempty"`
	CreatedAt   string                    `json:"created_at"`
	UpdatedAt   string                    `json:"updated_at"`
	CreatedBy   string                    `json:"created_by"`
	UpdatedBy   string                    `json:"updated_by"`
}

// RoleListResponse 角色列表响应
type RoleListResponse struct {
	Roles    []*RoleResponse `json:"roles"`
	Total    int64           `json:"total"`
	Page     int             `json:"page"`
	PageSize int             `json:"page_size"`
}

// RoleStatsResponse 角色统计响应
type RoleStatsResponse struct {
	Total  int64 `json:"total"`
	System int64 `json:"system"`
	Custom int64 `json:"custom"`
	Active int64 `json:"active"`
}

// PermissionSimpleResponse 权限简单响应结构
type PermissionSimpleResponse struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Code     string `json:"code"`
	Resource string `json:"resource"`
	Action   string `json:"action"`
}

// ListRoles 获取角色列表
// ListRoles 获取角色列表
// @Summary 获取角色列表
// @Description 获取系统中所有角色的列表，支持分页和筛选
// @Tags 角色管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(20)
// @Param name query string false "角色名称筛选"
// @Param code query string false "角色代码筛选"
// @Param type query string false "角色类型筛选" Enums(system,custom)
// @Param status query string false "角色状态筛选" Enums(active,inactive)
// @Param is_system query bool false "是否系统角色"
// @Success 200 {object} response.Response{data=RoleListResponse}
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/roles [get]
func (c *RoleController) ListRoles(ctx *gin.Context) {
	// 解析查询参数
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(ctx.DefaultQuery("page_size", "20"))
	
	var isSystem *bool
	if isSystemStr := ctx.Query("is_system"); isSystemStr != "" {
		if val, err := strconv.ParseBool(isSystemStr); err == nil {
			isSystem = &val
		}
	}
	
	req := &service.ListRolesRequest{
		Page:     page,
		PageSize: pageSize,
		Name:     ctx.Query("name"),
		Code:     ctx.Query("code"),
		Type:     ctx.Query("type"),
		Status:   ctx.Query("status"),
		IsSystem: isSystem,
	}
	
	// 调用服务
	resp, err := c.roleService.ListRoles(ctx.Request.Context(), req)
	if err != nil {
		logger.L.Error("Failed to list roles", zap.Error(err))
		response.InternalServerError(ctx, "获取角色列表失败", err)
		return
	}
	
	// 转换响应
	roles := make([]*RoleResponse, len(resp.Roles))
	for i, role := range resp.Roles {
		roles[i] = c.convertToRoleResponse(role)
	}
	
	response.Pagination(ctx, RoleListResponse{
		Roles:    roles,
		Total:    resp.Total,
		Page:     resp.Page,
		PageSize: resp.PageSize,
	}, resp.Total, resp.Page, resp.PageSize)
}

// GetRole 获取角色详情
// @Summary 获取角色详情
// @Description 根据角色ID获取角色的详细信息
// @Tags 角色管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "角色ID"
// @Success 200 {object} response.Response{data=RoleResponse}
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/roles/{id} [get]
func (c *RoleController) GetRole(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" {
		response.BadRequest(ctx, "角色ID不能为空", nil)
		return
	}
	
	role, err := c.roleService.GetRole(ctx.Request.Context(), id)
	if err != nil {
		logger.L.Error("Failed to get role", zap.String("id", id), zap.Error(err))
		response.NotFound(ctx, "角色不存在", err)
		return
	}
	
	response.Success(ctx, c.convertToRoleResponse(role))
}

// CreateRole 创建角色
// @Summary 创建角色
// @Description 创建新的系统角色
// @Tags 角色管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body RoleRequest true "角色信息"
// @Success 201 {object} response.Response{data=RoleResponse}
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 409 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/roles [post]
func (c *RoleController) CreateRole(ctx *gin.Context) {
	var req RoleRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.BadRequest(ctx, "请求参数无效", err)
		return
	}
	
	// 获取当前用户ID
	createdBy := ctx.GetString("user_id")
	if createdBy == "" {
		createdBy = "system"
	}
	
	createReq := &service.CreateRoleRequest{
		Name:        req.Name,
		Code:        req.Code,
		Description: req.Description,
		Type:        req.Type,
		Status:      req.Status,
		Permissions: req.Permissions,
		CreatedBy:   createdBy,
	}
	
	role, err := c.roleService.CreateRole(ctx.Request.Context(), createReq)
	if err != nil {
		logger.L.Error("Failed to create role", zap.Error(err))
		response.InternalServerError(ctx, "创建角色失败", err)
		return
	}
	
	response.Created(ctx, c.convertToRoleResponse(role))
}

// UpdateRole 更新角色
// @Summary 更新角色
// @Description 更新角色信息
// @Tags 角色管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "角色ID"
// @Param request body RoleRequest true "角色信息"
// @Success 200 {object} response.Response{data=RoleResponse}
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/roles/{id} [put]
func (c *RoleController) UpdateRole(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" {
		response.BadRequest(ctx, "角色ID不能为空", nil)
		return
	}
	
	var req RoleRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.BadRequest(ctx, "请求参数无效", err)
		return
	}
	
	// 获取当前用户ID
	updatedBy := ctx.GetString("user_id")
	if updatedBy == "" {
		updatedBy = "system"
	}
	
	updateReq := &service.UpdateRoleRequest{
		Name:        &req.Name,
		Description: &req.Description,
		Status:      &req.Status,
		UpdatedBy:   &updatedBy,
	}
	
	role, err := c.roleService.UpdateRole(ctx.Request.Context(), id, updateReq)
	if err != nil {
		logger.L.Error("Failed to update role", zap.String("id", id), zap.Error(err))
		response.InternalServerError(ctx, "更新角色失败", err)
		return
	}
	
	response.Updated(ctx, c.convertToRoleResponse(role))
}

// DeleteRole 删除角色
// @Summary 删除角色
// @Description 删除指定角色
// @Tags 角色管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "角色ID"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/roles/{id} [delete]
func (c *RoleController) DeleteRole(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" {
		response.BadRequest(ctx, "角色ID不能为空", nil)
		return
	}
	
	if err := c.roleService.DeleteRole(ctx.Request.Context(), id); err != nil {
		logger.L.Error("Failed to delete role", zap.String("id", id), zap.Error(err))
		response.InternalServerError(ctx, "删除角色失败", err)
		return
	}
	
	response.Deleted(ctx)
}

// GetRoleStats 获取角色统计
// @Summary 获取角色统计信息
// @Description 获取角色统计数据，包括总数、系统角色数、自定义角色数、活跃角色数
// @Tags 角色管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.Response{data=RoleStatsResponse}
// @Failure 401 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/roles/stats [get]
func (c *RoleController) GetRoleStats(ctx *gin.Context) {
	stats, err := c.roleService.GetRoleStats(ctx.Request.Context())
	if err != nil {
		logger.L.Error("Failed to get role stats", zap.Error(err))
		response.InternalServerError(ctx, "获取角色统计失败", err)
		return
	}
	
	response.Success(ctx, RoleStatsResponse{
		Total:  stats.Total,
		System: stats.System,
		Custom: stats.Custom,
		Active: stats.Active,
	})
}

// GetRolePermissions 获取角色权限
// @Summary 获取角色权限列表
// @Description 获取指定角色的权限列表
// @Tags 角色管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "角色ID"
// @Success 200 {object} response.Response{data=[]PermissionSimpleResponse}
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/roles/{id}/permissions [get]
func (c *RoleController) GetRolePermissions(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" {
		response.BadRequest(ctx, "角色ID不能为空", nil)
		return
	}
	
	permissions, err := c.roleService.GetRolePermissions(ctx.Request.Context(), id)
	if err != nil {
		logger.L.Error("Failed to get role permissions", zap.String("id", id), zap.Error(err))
		response.InternalServerError(ctx, "获取角色权限失败", err)
		return
	}
	
	// 转换响应
	responses := make([]*PermissionSimpleResponse, len(permissions))
	for i, permission := range permissions {
		responses[i] = &PermissionSimpleResponse{
			ID:       permission.ID,
			Name:     permission.Name,
			Code:     permission.Code,
			Resource: permission.Resource,
			Action:   permission.Action,
		}
	}
	
	response.Success(ctx, responses)
}

// AssignPermissions 为角色分配权限
// @Summary 为角色分配权限
// @Description 为指定角色分配权限列表
// @Tags 角色管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "角色ID"
// @Param request body object{permission_ids=[]string} true "权限ID列表"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/roles/{id}/permissions [post]
func (c *RoleController) AssignPermissions(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" {
		response.BadRequest(ctx, "角色ID不能为空", nil)
		return
	}
	
	var req struct {
		PermissionIDs []string `json:"permission_ids" binding:"required"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.BadRequest(ctx, "请求参数无效", err)
		return
	}
	
	if err := c.roleService.AssignPermissions(ctx.Request.Context(), id, req.PermissionIDs); err != nil {
		logger.L.Error("Failed to assign permissions", zap.String("role_id", id), zap.Error(err))
		response.InternalServerError(ctx, "分配权限失败", err)
		return
	}
	
	response.SuccessWithMessage(ctx, "权限分配成功", nil)
}

// RemovePermissions 移除角色权限
// @Summary 移除角色权限
// @Description 移除指定角色的权限列表
// @Tags 角色管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "角色ID"
// @Param request body object{permission_ids=[]string} true "权限ID列表"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/roles/{id}/permissions [delete]
func (c *RoleController) RemovePermissions(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" {
		response.BadRequest(ctx, "角色ID不能为空", nil)
		return
	}
	
	var req struct {
		PermissionIDs []string `json:"permission_ids" binding:"required"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.BadRequest(ctx, "请求参数无效", err)
		return
	}
	
	if err := c.roleService.RemovePermissions(ctx.Request.Context(), id, req.PermissionIDs); err != nil {
		logger.L.Error("Failed to remove permissions", zap.String("role_id", id), zap.Error(err))
		response.InternalServerError(ctx, "移除权限失败", err)
		return
	}
	
	response.SuccessWithMessage(ctx, "权限移除成功", nil)
}

// SyncPermissions 同步角色权限
// @Summary 同步角色权限
// @Description 同步指定角色的权限列表，替换现有权限
// @Tags 角色管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "角色ID"
// @Param request body object{permission_ids=[]string} true "权限ID列表"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/roles/{id}/permissions/sync [put]
func (c *RoleController) SyncPermissions(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" {
		response.BadRequest(ctx, "角色ID不能为空", nil)
		return
	}
	
	var req struct {
		PermissionIDs []string `json:"permission_ids"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.BadRequest(ctx, "请求参数无效", err)
		return
	}
	
	if err := c.roleService.SyncPermissions(ctx.Request.Context(), id, req.PermissionIDs); err != nil {
		logger.L.Error("Failed to sync permissions", zap.String("role_id", id), zap.Error(err))
		response.InternalServerError(ctx, "同步权限失败", err)
		return
	}
	
	response.SuccessWithMessage(ctx, "权限同步成功", nil)
}

// InitializeSystemRoles 初始化系统角色
// @Summary 初始化系统角色
// @Description 初始化系统预定义角色和权限
// @Tags 角色管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.Response
// @Failure 401 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/roles/initialize [post]
func (c *RoleController) InitializeSystemRoles(ctx *gin.Context) {
	if err := c.roleService.InitializeSystemRoles(ctx.Request.Context()); err != nil {
		logger.L.Error("Failed to initialize system roles", zap.Error(err))
		response.InternalServerError(ctx, "初始化系统角色失败", err)
		return
	}
	
	response.SuccessWithMessage(ctx, "系统角色初始化成功", nil)
}

// convertToRoleResponse 转换为角色响应结构
func (c *RoleController) convertToRoleResponse(role *model.Role) *RoleResponse {
	resp := &RoleResponse{
		ID:          role.ID,
		Name:        role.Name,
		Code:        role.Code,
		Description: role.Description,
		Type:        role.Type,
		Status:      role.Status,
		IsSystem:    role.IsSystem,
		CreatedAt:   role.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:   role.UpdatedAt.Format("2006-01-02 15:04:05"),
		CreatedBy:   role.CreatedBy,
		UpdatedBy:   role.UpdatedBy,
	}
	
	// 转换关联的权限
	if len(role.Permissions) > 0 {
		resp.Permissions = make([]*PermissionSimpleResponse, len(role.Permissions))
		for i, permission := range role.Permissions {
			resp.Permissions[i] = &PermissionSimpleResponse{
				ID:       permission.ID,
				Name:     permission.Name,
				Code:     permission.Code,
				Resource: permission.Resource,
				Action:   permission.Action,
			}
		}
	}
	
	return resp
}