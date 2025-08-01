package service

import (
	"context"
	"fmt"
	"strings"

	"alert_agent/internal/model"
	"alert_agent/internal/pkg/logger"
	"alert_agent/internal/repository"

	"go.uber.org/zap"
)

// RoleService 角色服务接口
type RoleService interface {
	// 基础CRUD操作
	CreateRole(ctx context.Context, req *CreateRoleRequest) (*model.Role, error)
	GetRole(ctx context.Context, id string) (*model.Role, error)
	GetRoleByCode(ctx context.Context, code string) (*model.Role, error)
	UpdateRole(ctx context.Context, id string, req *UpdateRoleRequest) (*model.Role, error)
	DeleteRole(ctx context.Context, id string) error
	
	// 查询操作
	ListRoles(ctx context.Context, req *ListRolesRequest) (*ListRolesResponse, error)
	GetRolesByType(ctx context.Context, roleType string) ([]*model.Role, error)
	GetSystemRoles(ctx context.Context) ([]*model.Role, error)
	
	// 批量操作
	BatchCreateRoles(ctx context.Context, req *BatchCreateRolesRequest) ([]*model.Role, error)
	BatchUpdateRoles(ctx context.Context, req *BatchUpdateRolesRequest) error
	BatchDeleteRoles(ctx context.Context, ids []string) error
	
	// 权限管理
	AssignPermissions(ctx context.Context, roleID string, permissionIDs []string) error
	RemovePermissions(ctx context.Context, roleID string, permissionIDs []string) error
	SyncPermissions(ctx context.Context, roleID string, permissionIDs []string) error
	GetRolePermissions(ctx context.Context, roleID string) ([]*model.Permission, error)
	
	// 用户管理
	GetRoleUsers(ctx context.Context, roleID string) ([]*model.User, error)
	GetUserRoles(ctx context.Context, userID string) ([]*model.Role, error)
	
	// 统计操作
	GetRoleStats(ctx context.Context) (*model.RoleStats, error)
	
	// 初始化系统角色
	InitializeSystemRoles(ctx context.Context) error
}

// roleService 角色服务实现
type roleService struct {
	roleRepo       repository.RoleRepository
	permissionRepo repository.PermissionRepository
}

// NewRoleService 创建角色服务实例
func NewRoleService(roleRepo repository.RoleRepository, permissionRepo repository.PermissionRepository) RoleService {
	return &roleService{
		roleRepo:       roleRepo,
		permissionRepo: permissionRepo,
	}
}

// CreateRoleRequest 创建角色请求
type CreateRoleRequest struct {
	Name        string   `json:"name" validate:"required,min=2,max=100"`
	Code        string   `json:"code" validate:"required,min=2,max=100"`
	Description string   `json:"description" validate:"max=500"`
	Type        string   `json:"type" validate:"oneof=system custom"`
	Status      string   `json:"status" validate:"oneof=active inactive"`
	Permissions []string `json:"permissions"` // 权限ID列表
	CreatedBy   string   `json:"created_by"`
}

// UpdateRoleRequest 更新角色请求
type UpdateRoleRequest struct {
	Name        *string `json:"name" validate:"omitempty,min=2,max=100"`
	Description *string `json:"description" validate:"omitempty,max=500"`
	Status      *string `json:"status" validate:"omitempty,oneof=active inactive"`
	UpdatedBy   *string `json:"updated_by"`
}

// ListRolesRequest 角色列表请求
type ListRolesRequest struct {
	Page     int    `json:"page" validate:"min=1"`
	PageSize int    `json:"page_size" validate:"min=1,max=100"`
	Name     string `json:"name"`
	Code     string `json:"code"`
	Type     string `json:"type"`
	Status   string `json:"status"`
	IsSystem *bool  `json:"is_system"`
}

// ListRolesResponse 角色列表响应
type ListRolesResponse struct {
	Roles    []*model.Role `json:"roles"`
	Total    int64         `json:"total"`
	Page     int           `json:"page"`
	PageSize int           `json:"page_size"`
}

// BatchCreateRolesRequest 批量创建角色请求
type BatchCreateRolesRequest struct {
	Roles     []*CreateRoleRequest `json:"roles" validate:"required,min=1,max=100"`
	CreatedBy string               `json:"created_by"`
}

// BatchUpdateRolesRequest 批量更新角色请求
type BatchUpdateRolesRequest struct {
	IDs       []string           `json:"ids" validate:"required,min=1"`
	Updates   *UpdateRoleRequest `json:"updates" validate:"required"`
	UpdatedBy string             `json:"updated_by"`
}

// CreateRole 创建角色
func (s *roleService) CreateRole(ctx context.Context, req *CreateRoleRequest) (*model.Role, error) {
	logger.L.Debug("Creating role", zap.String("code", req.Code))
	
	// 检查角色代码是否已存在
	existingRole, err := s.roleRepo.GetByCode(ctx, req.Code)
	if err == nil && existingRole != nil {
		return nil, fmt.Errorf("role with code '%s' already exists", req.Code)
	}
	
	// 创建角色对象
	role := &model.Role{
		Name:        req.Name,
		Code:        req.Code,
		Description: req.Description,
		Type:        req.Type,
		Status:      req.Status,
		IsSystem:    req.Type == model.RoleTypeSystem,
		CreatedBy:   req.CreatedBy,
	}
	
	// 设置默认值
	if role.Type == "" {
		role.Type = model.RoleTypeCustom
	}
	if role.Status == "" {
		role.Status = model.RoleStatusActive
	}
	
	// 保存到数据库
	if err := s.roleRepo.Create(ctx, role); err != nil {
		return nil, fmt.Errorf("failed to create role: %w", err)
	}
	
	// 分配权限
	if len(req.Permissions) > 0 {
		if err := s.roleRepo.AssignPermissions(ctx, role.ID, req.Permissions); err != nil {
			logger.L.Error("Failed to assign permissions to role", zap.Error(err))
			// 不回滚角色创建，只记录错误
		}
	}
	
	logger.L.Info("Role created successfully", zap.String("id", role.ID))
	return role, nil
}

// GetRole 获取角色
func (s *roleService) GetRole(ctx context.Context, id string) (*model.Role, error) {
	if id == "" {
		return nil, fmt.Errorf("role ID is required")
	}
	
	role, err := s.roleRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get role: %w", err)
	}
	
	return role, nil
}

// GetRoleByCode 根据代码获取角色
func (s *roleService) GetRoleByCode(ctx context.Context, code string) (*model.Role, error) {
	if code == "" {
		return nil, fmt.Errorf("role code is required")
	}
	
	role, err := s.roleRepo.GetByCode(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("failed to get role by code: %w", err)
	}
	
	return role, nil
}

// UpdateRole 更新角色
func (s *roleService) UpdateRole(ctx context.Context, id string, req *UpdateRoleRequest) (*model.Role, error) {
	logger.L.Debug("Updating role", zap.String("id", id))
	
	// 获取现有角色
	role, err := s.roleRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get role: %w", err)
	}
	
	// 检查是否为系统角色
	if role.IsSystemRole() {
		return nil, fmt.Errorf("cannot update system role")
	}
	
	// 更新字段
	if req.Name != nil {
		role.Name = *req.Name
	}
	if req.Description != nil {
		role.Description = *req.Description
	}
	if req.Status != nil {
		role.Status = *req.Status
	}
	if req.UpdatedBy != nil {
		role.UpdatedBy = *req.UpdatedBy
	}
	
	// 保存更新
	if err := s.roleRepo.Update(ctx, role); err != nil {
		return nil, fmt.Errorf("failed to update role: %w", err)
	}
	
	logger.L.Info("Role updated successfully", zap.String("id", id))
	return role, nil
}

// DeleteRole 删除角色
func (s *roleService) DeleteRole(ctx context.Context, id string) error {
	logger.L.Debug("Deleting role", zap.String("id", id))
	
	if id == "" {
		return fmt.Errorf("role ID is required")
	}
	
	if err := s.roleRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete role: %w", err)
	}
	
	logger.L.Info("Role deleted successfully", zap.String("id", id))
	return nil
}

// ListRoles 获取角色列表
func (s *roleService) ListRoles(ctx context.Context, req *ListRolesRequest) (*ListRolesResponse, error) {
	// 设置默认分页参数
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 20
	}
	
	// 构建过滤条件
	filters := make(map[string]interface{})
	if req.Name != "" {
		filters["name"] = req.Name
	}
	if req.Code != "" {
		filters["code"] = req.Code
	}
	if req.Type != "" {
		filters["type"] = req.Type
	}
	if req.Status != "" {
		filters["status"] = req.Status
	}
	if req.IsSystem != nil {
		filters["is_system"] = *req.IsSystem
	}
	
	// 计算偏移量
	offset := (req.Page - 1) * req.PageSize
	
	// 获取角色列表
	roles, total, err := s.roleRepo.List(ctx, filters, offset, req.PageSize)
	if err != nil {
		return nil, fmt.Errorf("failed to list roles: %w", err)
	}
	
	return &ListRolesResponse{
		Roles:    roles,
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
	}, nil
}

// GetRolesByType 根据类型获取角色
func (s *roleService) GetRolesByType(ctx context.Context, roleType string) ([]*model.Role, error) {
	if roleType == "" {
		return nil, fmt.Errorf("role type is required")
	}
	
	roles, err := s.roleRepo.GetByType(ctx, roleType)
	if err != nil {
		return nil, fmt.Errorf("failed to get roles by type: %w", err)
	}
	
	return roles, nil
}

// GetSystemRoles 获取系统角色
func (s *roleService) GetSystemRoles(ctx context.Context) ([]*model.Role, error) {
	roles, err := s.roleRepo.GetSystemRoles(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get system roles: %w", err)
	}
	
	return roles, nil
}

// BatchCreateRoles 批量创建角色
func (s *roleService) BatchCreateRoles(ctx context.Context, req *BatchCreateRolesRequest) ([]*model.Role, error) {
	logger.L.Debug("Batch creating roles", zap.Int("count", len(req.Roles)))
	
	var roles []*model.Role
	var codes []string
	
	// 验证并构建角色对象
	for _, roleReq := range req.Roles {
		// 检查代码重复
		for _, code := range codes {
			if code == roleReq.Code {
				return nil, fmt.Errorf("duplicate role code in request: %s", roleReq.Code)
			}
		}
		codes = append(codes, roleReq.Code)
		
		// 检查数据库中是否已存在
		existingRole, err := s.roleRepo.GetByCode(ctx, roleReq.Code)
		if err == nil && existingRole != nil {
			return nil, fmt.Errorf("role with code '%s' already exists", roleReq.Code)
		}
		
		// 创建角色对象
		role := &model.Role{
			Name:        roleReq.Name,
			Code:        roleReq.Code,
			Description: roleReq.Description,
			Type:        roleReq.Type,
			Status:      roleReq.Status,
			IsSystem:    roleReq.Type == model.RoleTypeSystem,
			CreatedBy:   req.CreatedBy,
		}
		
		// 设置默认值
		if role.Type == "" {
			role.Type = model.RoleTypeCustom
		}
		if role.Status == "" {
			role.Status = model.RoleStatusActive
		}
		
		roles = append(roles, role)
	}
	
	// 批量保存
	if err := s.roleRepo.BatchCreate(ctx, roles); err != nil {
		return nil, fmt.Errorf("failed to batch create roles: %w", err)
	}
	
	// 分配权限
	for i, roleReq := range req.Roles {
		if len(roleReq.Permissions) > 0 {
			if err := s.roleRepo.AssignPermissions(ctx, roles[i].ID, roleReq.Permissions); err != nil {
				logger.L.Error("Failed to assign permissions to role", zap.String("role_id", roles[i].ID), zap.Error(err))
			}
		}
	}
	
	logger.L.Info("Roles batch created successfully", zap.Int("count", len(roles)))
	return roles, nil
}

// BatchUpdateRoles 批量更新角色
func (s *roleService) BatchUpdateRoles(ctx context.Context, req *BatchUpdateRolesRequest) error {
	logger.L.Debug("Batch updating roles", zap.Int("count", len(req.IDs)))
	
	// 构建更新字段
	updates := make(map[string]interface{})
	if req.Updates.Name != nil {
		updates["name"] = *req.Updates.Name
	}
	if req.Updates.Description != nil {
		updates["description"] = *req.Updates.Description
	}
	if req.Updates.Status != nil {
		updates["status"] = *req.Updates.Status
	}
	if req.UpdatedBy != "" {
		updates["updated_by"] = req.UpdatedBy
	}
	
	if len(updates) == 0 {
		return fmt.Errorf("no fields to update")
	}
	
	// 执行批量更新
	if err := s.roleRepo.BatchUpdate(ctx, req.IDs, updates); err != nil {
		return fmt.Errorf("failed to batch update roles: %w", err)
	}
	
	logger.L.Info("Roles batch updated successfully", zap.Int("count", len(req.IDs)))
	return nil
}

// BatchDeleteRoles 批量删除角色
func (s *roleService) BatchDeleteRoles(ctx context.Context, ids []string) error {
	logger.L.Debug("Batch deleting roles", zap.Int("count", len(ids)))
	
	if len(ids) == 0 {
		return fmt.Errorf("no role IDs provided")
	}
	
	if err := s.roleRepo.BatchDelete(ctx, ids); err != nil {
		return fmt.Errorf("failed to batch delete roles: %w", err)
	}
	
	logger.L.Info("Roles batch deleted successfully", zap.Int("count", len(ids)))
	return nil
}

// AssignPermissions 为角色分配权限
func (s *roleService) AssignPermissions(ctx context.Context, roleID string, permissionIDs []string) error {
	logger.L.Debug("Assigning permissions to role", zap.String("role_id", roleID), zap.Int("permission_count", len(permissionIDs)))
	
	if roleID == "" {
		return fmt.Errorf("role ID is required")
	}
	
	if len(permissionIDs) == 0 {
		return fmt.Errorf("no permission IDs provided")
	}
	
	if err := s.roleRepo.AssignPermissions(ctx, roleID, permissionIDs); err != nil {
		return fmt.Errorf("failed to assign permissions: %w", err)
	}
	
	logger.L.Info("Permissions assigned to role successfully", zap.String("role_id", roleID))
	return nil
}

// RemovePermissions 移除角色权限
func (s *roleService) RemovePermissions(ctx context.Context, roleID string, permissionIDs []string) error {
	logger.L.Debug("Removing permissions from role", zap.String("role_id", roleID), zap.Int("permission_count", len(permissionIDs)))
	
	if roleID == "" {
		return fmt.Errorf("role ID is required")
	}
	
	if len(permissionIDs) == 0 {
		return fmt.Errorf("no permission IDs provided")
	}
	
	if err := s.roleRepo.RemovePermissions(ctx, roleID, permissionIDs); err != nil {
		return fmt.Errorf("failed to remove permissions: %w", err)
	}
	
	logger.L.Info("Permissions removed from role successfully", zap.String("role_id", roleID))
	return nil
}

// SyncPermissions 同步角色权限
func (s *roleService) SyncPermissions(ctx context.Context, roleID string, permissionIDs []string) error {
	logger.L.Debug("Syncing permissions for role", zap.String("role_id", roleID), zap.Int("permission_count", len(permissionIDs)))
	
	if roleID == "" {
		return fmt.Errorf("role ID is required")
	}
	
	if err := s.roleRepo.SyncPermissions(ctx, roleID, permissionIDs); err != nil {
		return fmt.Errorf("failed to sync permissions: %w", err)
	}
	
	logger.L.Info("Permissions synced for role successfully", zap.String("role_id", roleID))
	return nil
}

// GetRolePermissions 获取角色权限
func (s *roleService) GetRolePermissions(ctx context.Context, roleID string) ([]*model.Permission, error) {
	if roleID == "" {
		return nil, fmt.Errorf("role ID is required")
	}
	
	permissions, err := s.roleRepo.GetRolePermissions(ctx, roleID)
	if err != nil {
		return nil, fmt.Errorf("failed to get role permissions: %w", err)
	}
	
	return permissions, nil
}

// GetRoleUsers 获取角色关联的用户
func (s *roleService) GetRoleUsers(ctx context.Context, roleID string) ([]*model.User, error) {
	if roleID == "" {
		return nil, fmt.Errorf("role ID is required")
	}
	
	users, err := s.roleRepo.GetRoleUsers(ctx, roleID)
	if err != nil {
		return nil, fmt.Errorf("failed to get role users: %w", err)
	}
	
	return users, nil
}

// GetUserRoles 获取用户角色
func (s *roleService) GetUserRoles(ctx context.Context, userID string) ([]*model.Role, error) {
	if userID == "" {
		return nil, fmt.Errorf("user ID is required")
	}
	
	roles, err := s.roleRepo.GetUserRoles(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user roles: %w", err)
	}
	
	return roles, nil
}

// GetRoleStats 获取角色统计
func (s *roleService) GetRoleStats(ctx context.Context) (*model.RoleStats, error) {
	stats, err := s.roleRepo.GetStats(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get role stats: %w", err)
	}
	
	return stats, nil
}

// InitializeSystemRoles 初始化系统角色
func (s *roleService) InitializeSystemRoles(ctx context.Context) error {
	logger.L.Info("Initializing system roles")
	
	// 定义系统角色
	systemRoles := []*model.Role{
		{
			Name:        "超级管理员",
			Code:        model.RoleCodeSuperAdmin,
			Description: "系统超级管理员，拥有所有权限",
			Type:        model.RoleTypeSystem,
			Status:      model.RoleStatusActive,
			IsSystem:    true,
		},
		{
			Name:        "管理员",
			Code:        model.RoleCodeAdmin,
			Description: "系统管理员，拥有大部分管理权限",
			Type:        model.RoleTypeSystem,
			Status:      model.RoleStatusActive,
			IsSystem:    true,
		},
		{
			Name:        "操作员",
			Code:        model.RoleCodeOperator,
			Description: "系统操作员，拥有基本操作权限",
			Type:        model.RoleTypeSystem,
			Status:      model.RoleStatusActive,
			IsSystem:    true,
		},
		{
			Name:        "查看者",
			Code:        model.RoleCodeViewer,
			Description: "系统查看者，只有查看权限",
			Type:        model.RoleTypeSystem,
			Status:      model.RoleStatusActive,
			IsSystem:    true,
		},
	}
	
	// 检查并创建不存在的角色
	var newRoles []*model.Role
	for _, role := range systemRoles {
		existing, err := s.roleRepo.GetByCode(ctx, role.Code)
		if err != nil && !strings.Contains(err.Error(), "not found") {
			return fmt.Errorf("failed to check role %s: %w", role.Code, err)
		}
		
		if existing == nil {
			newRoles = append(newRoles, role)
		}
	}
	
	if len(newRoles) > 0 {
		if err := s.roleRepo.BatchCreate(ctx, newRoles); err != nil {
			return fmt.Errorf("failed to create system roles: %w", err)
		}
		logger.L.Info("System roles initialized", zap.Int("count", len(newRoles)))
		
		// 为超级管理员分配所有权限
		if err := s.assignAllPermissionsToSuperAdmin(ctx); err != nil {
			logger.L.Error("Failed to assign permissions to super admin", zap.Error(err))
		}
	} else {
		logger.L.Info("All system roles already exist")
	}
	
	return nil
}

// assignAllPermissionsToSuperAdmin 为超级管理员分配所有权限
func (s *roleService) assignAllPermissionsToSuperAdmin(ctx context.Context) error {
	// 获取超级管理员角色
	superAdminRole, err := s.roleRepo.GetByCode(ctx, model.RoleCodeSuperAdmin)
	if err != nil {
		return fmt.Errorf("failed to get super admin role: %w", err)
	}
	
	// 获取所有权限
	allPermissions, _, err := s.permissionRepo.List(ctx, map[string]interface{}{"status": model.PermissionStatusActive}, 0, 1000)
	if err != nil {
		return fmt.Errorf("failed to get all permissions: %w", err)
	}
	
	// 提取权限ID
	var permissionIDs []string
	for _, permission := range allPermissions {
		permissionIDs = append(permissionIDs, permission.ID)
	}
	
	// 同步权限
	if len(permissionIDs) > 0 {
		if err := s.roleRepo.SyncPermissions(ctx, superAdminRole.ID, permissionIDs); err != nil {
			return fmt.Errorf("failed to sync permissions for super admin: %w", err)
		}
		logger.L.Info("All permissions assigned to super admin", zap.Int("permission_count", len(permissionIDs)))
	}
	
	return nil
}