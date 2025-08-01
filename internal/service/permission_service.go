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

// PermissionService 权限服务接口
type PermissionService interface {
	// 基础CRUD操作
	CreatePermission(ctx context.Context, req *CreatePermissionRequest) (*model.Permission, error)
	GetPermission(ctx context.Context, id string) (*model.Permission, error)
	GetPermissionByCode(ctx context.Context, code string) (*model.Permission, error)
	UpdatePermission(ctx context.Context, id string, req *UpdatePermissionRequest) (*model.Permission, error)
	DeletePermission(ctx context.Context, id string) error
	
	// 查询操作
	ListPermissions(ctx context.Context, req *ListPermissionsRequest) (*ListPermissionsResponse, error)
	GetPermissionsByCategory(ctx context.Context, category string) ([]*model.Permission, error)
	GetPermissionsByType(ctx context.Context, permissionType string) ([]*model.Permission, error)
	
	// 批量操作
	BatchCreatePermissions(ctx context.Context, req *BatchCreatePermissionsRequest) ([]*model.Permission, error)
	BatchUpdatePermissions(ctx context.Context, req *BatchUpdatePermissionsRequest) error
	BatchDeletePermissions(ctx context.Context, ids []string) error
	
	// 统计操作
	GetPermissionStats(ctx context.Context) (*model.PermissionStats, error)
	
	// 初始化系统权限
	InitializeSystemPermissions(ctx context.Context) error
}

// permissionService 权限服务实现
type permissionService struct {
	permissionRepo repository.PermissionRepository
	roleRepo       repository.RoleRepository
}

// NewPermissionService 创建权限服务实例
func NewPermissionService(permissionRepo repository.PermissionRepository, roleRepo repository.RoleRepository) PermissionService {
	return &permissionService{
		permissionRepo: permissionRepo,
		roleRepo:       roleRepo,
	}
}

// CreatePermissionRequest 创建权限请求
type CreatePermissionRequest struct {
	Name        string `json:"name" validate:"required,min=2,max=100"`
	Code        string `json:"code" validate:"required,min=2,max=100"`
	Description string `json:"description" validate:"max=500"`
	Resource    string `json:"resource" validate:"required,min=1,max=255"`
	Action      string `json:"action" validate:"required,oneof=create read update delete list export import manage"`
	Category    string `json:"category" validate:"required,oneof=user role permission alert rule provider config system"`
	Type        string `json:"type" validate:"oneof=system custom"`
	Status      string `json:"status" validate:"oneof=active inactive"`
	CreatedBy   string `json:"created_by"`
}

// UpdatePermissionRequest 更新权限请求
type UpdatePermissionRequest struct {
	Name        *string `json:"name" validate:"omitempty,min=2,max=100"`
	Description *string `json:"description" validate:"omitempty,max=500"`
	Resource    *string `json:"resource" validate:"omitempty,min=1,max=255"`
	Action      *string `json:"action" validate:"omitempty,oneof=create read update delete list export import manage"`
	Category    *string `json:"category" validate:"omitempty,oneof=user role permission alert rule provider config system"`
	Status      *string `json:"status" validate:"omitempty,oneof=active inactive"`
	UpdatedBy   *string `json:"updated_by"`
}

// ListPermissionsRequest 权限列表请求
type ListPermissionsRequest struct {
	Page     int    `json:"page" validate:"min=1"`
	PageSize int    `json:"page_size" validate:"min=1,max=100"`
	Name     string `json:"name"`
	Code     string `json:"code"`
	Category string `json:"category"`
	Type     string `json:"type"`
	Status   string `json:"status"`
	Resource string `json:"resource"`
	Action   string `json:"action"`
}

// ListPermissionsResponse 权限列表响应
type ListPermissionsResponse struct {
	Permissions []*model.Permission `json:"permissions"`
	Total       int64              `json:"total"`
	Page        int                `json:"page"`
	PageSize    int                `json:"page_size"`
}

// BatchCreatePermissionsRequest 批量创建权限请求
type BatchCreatePermissionsRequest struct {
	Permissions []*CreatePermissionRequest `json:"permissions" validate:"required,min=1,max=100"`
	CreatedBy   string                     `json:"created_by"`
}

// BatchUpdatePermissionsRequest 批量更新权限请求
type BatchUpdatePermissionsRequest struct {
	IDs       []string                `json:"ids" validate:"required,min=1"`
	Updates   *UpdatePermissionRequest `json:"updates" validate:"required"`
	UpdatedBy string                  `json:"updated_by"`
}

// CreatePermission 创建权限
func (s *permissionService) CreatePermission(ctx context.Context, req *CreatePermissionRequest) (*model.Permission, error) {
	logger.L.Debug("Creating permission", zap.String("code", req.Code))
	
	// 检查权限代码是否已存在
	existingPermission, err := s.permissionRepo.GetByCode(ctx, req.Code)
	if err == nil && existingPermission != nil {
		return nil, fmt.Errorf("permission with code '%s' already exists", req.Code)
	}
	
	// 创建权限对象
	permission := &model.Permission{
		Name:        req.Name,
		Code:        req.Code,
		Description: req.Description,
		Resource:    req.Resource,
		Action:      req.Action,
		Category:    req.Category,
		Type:        req.Type,
		Status:      req.Status,
		IsSystem:    req.Type == model.PermissionTypeSystem,
		CreatedBy:   req.CreatedBy,
	}
	
	// 设置默认值
	if permission.Type == "" {
		permission.Type = model.PermissionTypeCustom
	}
	if permission.Status == "" {
		permission.Status = model.PermissionStatusActive
	}
	
	// 保存到数据库
	if err := s.permissionRepo.Create(ctx, permission); err != nil {
		return nil, fmt.Errorf("failed to create permission: %w", err)
	}
	
	logger.L.Info("Permission created successfully", zap.String("id", permission.ID))
	return permission, nil
}

// GetPermission 获取权限
func (s *permissionService) GetPermission(ctx context.Context, id string) (*model.Permission, error) {
	if id == "" {
		return nil, fmt.Errorf("permission ID is required")
	}
	
	permission, err := s.permissionRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get permission: %w", err)
	}
	
	return permission, nil
}

// GetPermissionByCode 根据代码获取权限
func (s *permissionService) GetPermissionByCode(ctx context.Context, code string) (*model.Permission, error) {
	if code == "" {
		return nil, fmt.Errorf("permission code is required")
	}
	
	permission, err := s.permissionRepo.GetByCode(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("failed to get permission by code: %w", err)
	}
	
	return permission, nil
}

// UpdatePermission 更新权限
func (s *permissionService) UpdatePermission(ctx context.Context, id string, req *UpdatePermissionRequest) (*model.Permission, error) {
	logger.L.Debug("Updating permission", zap.String("id", id))
	
	// 获取现有权限
	permission, err := s.permissionRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get permission: %w", err)
	}
	
	// 检查是否为系统权限
	if permission.IsSystemPermission() {
		return nil, fmt.Errorf("cannot update system permission")
	}
	
	// 更新字段
	if req.Name != nil {
		permission.Name = *req.Name
	}
	if req.Description != nil {
		permission.Description = *req.Description
	}
	if req.Resource != nil {
		permission.Resource = *req.Resource
	}
	if req.Action != nil {
		permission.Action = *req.Action
	}
	if req.Category != nil {
		permission.Category = *req.Category
	}
	if req.Status != nil {
		permission.Status = *req.Status
	}
	if req.UpdatedBy != nil {
		permission.UpdatedBy = *req.UpdatedBy
	}
	
	// 保存更新
	if err := s.permissionRepo.Update(ctx, permission); err != nil {
		return nil, fmt.Errorf("failed to update permission: %w", err)
	}
	
	logger.L.Info("Permission updated successfully", zap.String("id", id))
	return permission, nil
}

// DeletePermission 删除权限
func (s *permissionService) DeletePermission(ctx context.Context, id string) error {
	logger.L.Debug("Deleting permission", zap.String("id", id))
	
	if id == "" {
		return fmt.Errorf("permission ID is required")
	}
	
	if err := s.permissionRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete permission: %w", err)
	}
	
	logger.L.Info("Permission deleted successfully", zap.String("id", id))
	return nil
}

// ListPermissions 获取权限列表
func (s *permissionService) ListPermissions(ctx context.Context, req *ListPermissionsRequest) (*ListPermissionsResponse, error) {
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
	if req.Category != "" {
		filters["category"] = req.Category
	}
	if req.Type != "" {
		filters["type"] = req.Type
	}
	if req.Status != "" {
		filters["status"] = req.Status
	}
	if req.Resource != "" {
		filters["resource"] = req.Resource
	}
	if req.Action != "" {
		filters["action"] = req.Action
	}
	
	// 计算偏移量
	offset := (req.Page - 1) * req.PageSize
	
	// 获取权限列表
	permissions, total, err := s.permissionRepo.List(ctx, filters, offset, req.PageSize)
	if err != nil {
		return nil, fmt.Errorf("failed to list permissions: %w", err)
	}
	
	return &ListPermissionsResponse{
		Permissions: permissions,
		Total:       total,
		Page:        req.Page,
		PageSize:    req.PageSize,
	}, nil
}

// GetPermissionsByCategory 根据分类获取权限
func (s *permissionService) GetPermissionsByCategory(ctx context.Context, category string) ([]*model.Permission, error) {
	if category == "" {
		return nil, fmt.Errorf("category is required")
	}
	
	permissions, err := s.permissionRepo.GetByCategory(ctx, category)
	if err != nil {
		return nil, fmt.Errorf("failed to get permissions by category: %w", err)
	}
	
	return permissions, nil
}

// GetPermissionsByType 根据类型获取权限
func (s *permissionService) GetPermissionsByType(ctx context.Context, permissionType string) ([]*model.Permission, error) {
	if permissionType == "" {
		return nil, fmt.Errorf("permission type is required")
	}
	
	permissions, err := s.permissionRepo.GetByType(ctx, permissionType)
	if err != nil {
		return nil, fmt.Errorf("failed to get permissions by type: %w", err)
	}
	
	return permissions, nil
}

// BatchCreatePermissions 批量创建权限
func (s *permissionService) BatchCreatePermissions(ctx context.Context, req *BatchCreatePermissionsRequest) ([]*model.Permission, error) {
	logger.L.Debug("Batch creating permissions", zap.Int("count", len(req.Permissions)))
	
	var permissions []*model.Permission
	var codes []string
	
	// 验证并构建权限对象
	for _, permReq := range req.Permissions {
		// 检查代码重复
		for _, code := range codes {
			if code == permReq.Code {
				return nil, fmt.Errorf("duplicate permission code in request: %s", permReq.Code)
			}
		}
		codes = append(codes, permReq.Code)
		
		// 检查数据库中是否已存在
		existingPermission, err := s.permissionRepo.GetByCode(ctx, permReq.Code)
		if err == nil && existingPermission != nil {
			return nil, fmt.Errorf("permission with code '%s' already exists", permReq.Code)
		}
		
		// 创建权限对象
		permission := &model.Permission{
			Name:        permReq.Name,
			Code:        permReq.Code,
			Description: permReq.Description,
			Resource:    permReq.Resource,
			Action:      permReq.Action,
			Category:    permReq.Category,
			Type:        permReq.Type,
			Status:      permReq.Status,
			IsSystem:    permReq.Type == model.PermissionTypeSystem,
			CreatedBy:   req.CreatedBy,
		}
		
		// 设置默认值
		if permission.Type == "" {
			permission.Type = model.PermissionTypeCustom
		}
		if permission.Status == "" {
			permission.Status = model.PermissionStatusActive
		}
		
		permissions = append(permissions, permission)
	}
	
	// 批量保存
	if err := s.permissionRepo.BatchCreate(ctx, permissions); err != nil {
		return nil, fmt.Errorf("failed to batch create permissions: %w", err)
	}
	
	logger.L.Info("Permissions batch created successfully", zap.Int("count", len(permissions)))
	return permissions, nil
}

// BatchUpdatePermissions 批量更新权限
func (s *permissionService) BatchUpdatePermissions(ctx context.Context, req *BatchUpdatePermissionsRequest) error {
	logger.L.Debug("Batch updating permissions", zap.Int("count", len(req.IDs)))
	
	// 构建更新字段
	updates := make(map[string]interface{})
	if req.Updates.Name != nil {
		updates["name"] = *req.Updates.Name
	}
	if req.Updates.Description != nil {
		updates["description"] = *req.Updates.Description
	}
	if req.Updates.Resource != nil {
		updates["resource"] = *req.Updates.Resource
	}
	if req.Updates.Action != nil {
		updates["action"] = *req.Updates.Action
	}
	if req.Updates.Category != nil {
		updates["category"] = *req.Updates.Category
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
	if err := s.permissionRepo.BatchUpdate(ctx, req.IDs, updates); err != nil {
		return fmt.Errorf("failed to batch update permissions: %w", err)
	}
	
	logger.L.Info("Permissions batch updated successfully", zap.Int("count", len(req.IDs)))
	return nil
}

// BatchDeletePermissions 批量删除权限
func (s *permissionService) BatchDeletePermissions(ctx context.Context, ids []string) error {
	logger.L.Debug("Batch deleting permissions", zap.Int("count", len(ids)))
	
	if len(ids) == 0 {
		return fmt.Errorf("no permission IDs provided")
	}
	
	if err := s.permissionRepo.BatchDelete(ctx, ids); err != nil {
		return fmt.Errorf("failed to batch delete permissions: %w", err)
	}
	
	logger.L.Info("Permissions batch deleted successfully", zap.Int("count", len(ids)))
	return nil
}

// GetPermissionStats 获取权限统计
func (s *permissionService) GetPermissionStats(ctx context.Context) (*model.PermissionStats, error) {
	stats, err := s.permissionRepo.GetStats(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get permission stats: %w", err)
	}
	
	return stats, nil
}

// InitializeSystemPermissions 初始化系统权限
func (s *permissionService) InitializeSystemPermissions(ctx context.Context) error {
	logger.L.Info("Initializing system permissions")
	
	// 定义系统权限
	systemPermissions := []*model.Permission{
		// 用户管理权限
		{Name: "用户列表", Code: "user:list", Resource: "user", Action: "list", Category: "user", Type: model.PermissionTypeSystem, Status: model.PermissionStatusActive, IsSystem: true},
		{Name: "用户详情", Code: "user:read", Resource: "user", Action: "read", Category: "user", Type: model.PermissionTypeSystem, Status: model.PermissionStatusActive, IsSystem: true},
		{Name: "创建用户", Code: "user:create", Resource: "user", Action: "create", Category: "user", Type: model.PermissionTypeSystem, Status: model.PermissionStatusActive, IsSystem: true},
		{Name: "更新用户", Code: "user:update", Resource: "user", Action: "update", Category: "user", Type: model.PermissionTypeSystem, Status: model.PermissionStatusActive, IsSystem: true},
		{Name: "删除用户", Code: "user:delete", Resource: "user", Action: "delete", Category: "user", Type: model.PermissionTypeSystem, Status: model.PermissionStatusActive, IsSystem: true},
		{Name: "用户管理", Code: "user:manage", Resource: "user", Action: "manage", Category: "user", Type: model.PermissionTypeSystem, Status: model.PermissionStatusActive, IsSystem: true},
		
		// 角色管理权限
		{Name: "角色列表", Code: "role:list", Resource: "role", Action: "list", Category: "role", Type: model.PermissionTypeSystem, Status: model.PermissionStatusActive, IsSystem: true},
		{Name: "角色详情", Code: "role:read", Resource: "role", Action: "read", Category: "role", Type: model.PermissionTypeSystem, Status: model.PermissionStatusActive, IsSystem: true},
		{Name: "创建角色", Code: "role:create", Resource: "role", Action: "create", Category: "role", Type: model.PermissionTypeSystem, Status: model.PermissionStatusActive, IsSystem: true},
		{Name: "更新角色", Code: "role:update", Resource: "role", Action: "update", Category: "role", Type: model.PermissionTypeSystem, Status: model.PermissionStatusActive, IsSystem: true},
		{Name: "删除角色", Code: "role:delete", Resource: "role", Action: "delete", Category: "role", Type: model.PermissionTypeSystem, Status: model.PermissionStatusActive, IsSystem: true},
		{Name: "角色管理", Code: "role:manage", Resource: "role", Action: "manage", Category: "role", Type: model.PermissionTypeSystem, Status: model.PermissionStatusActive, IsSystem: true},
		
		// 权限管理权限
		{Name: "权限列表", Code: "permission:list", Resource: "permission", Action: "list", Category: "permission", Type: model.PermissionTypeSystem, Status: model.PermissionStatusActive, IsSystem: true},
		{Name: "权限详情", Code: "permission:read", Resource: "permission", Action: "read", Category: "permission", Type: model.PermissionTypeSystem, Status: model.PermissionStatusActive, IsSystem: true},
		{Name: "创建权限", Code: "permission:create", Resource: "permission", Action: "create", Category: "permission", Type: model.PermissionTypeSystem, Status: model.PermissionStatusActive, IsSystem: true},
		{Name: "更新权限", Code: "permission:update", Resource: "permission", Action: "update", Category: "permission", Type: model.PermissionTypeSystem, Status: model.PermissionStatusActive, IsSystem: true},
		{Name: "删除权限", Code: "permission:delete", Resource: "permission", Action: "delete", Category: "permission", Type: model.PermissionTypeSystem, Status: model.PermissionStatusActive, IsSystem: true},
		{Name: "权限管理", Code: "permission:manage", Resource: "permission", Action: "manage", Category: "permission", Type: model.PermissionTypeSystem, Status: model.PermissionStatusActive, IsSystem: true},
		
		// 告警管理权限
		{Name: "告警列表", Code: "alert:list", Resource: "alert", Action: "list", Category: "alert", Type: model.PermissionTypeSystem, Status: model.PermissionStatusActive, IsSystem: true},
		{Name: "告警详情", Code: "alert:read", Resource: "alert", Action: "read", Category: "alert", Type: model.PermissionTypeSystem, Status: model.PermissionStatusActive, IsSystem: true},
		{Name: "创建告警", Code: "alert:create", Resource: "alert", Action: "create", Category: "alert", Type: model.PermissionTypeSystem, Status: model.PermissionStatusActive, IsSystem: true},
		{Name: "更新告警", Code: "alert:update", Resource: "alert", Action: "update", Category: "alert", Type: model.PermissionTypeSystem, Status: model.PermissionStatusActive, IsSystem: true},
		{Name: "删除告警", Code: "alert:delete", Resource: "alert", Action: "delete", Category: "alert", Type: model.PermissionTypeSystem, Status: model.PermissionStatusActive, IsSystem: true},
		{Name: "告警管理", Code: "alert:manage", Resource: "alert", Action: "manage", Category: "alert", Type: model.PermissionTypeSystem, Status: model.PermissionStatusActive, IsSystem: true},
		
		// 规则管理权限
		{Name: "规则列表", Code: "rule:list", Resource: "rule", Action: "list", Category: "rule", Type: model.PermissionTypeSystem, Status: model.PermissionStatusActive, IsSystem: true},
		{Name: "规则详情", Code: "rule:read", Resource: "rule", Action: "read", Category: "rule", Type: model.PermissionTypeSystem, Status: model.PermissionStatusActive, IsSystem: true},
		{Name: "创建规则", Code: "rule:create", Resource: "rule", Action: "create", Category: "rule", Type: model.PermissionTypeSystem, Status: model.PermissionStatusActive, IsSystem: true},
		{Name: "更新规则", Code: "rule:update", Resource: "rule", Action: "update", Category: "rule", Type: model.PermissionTypeSystem, Status: model.PermissionStatusActive, IsSystem: true},
		{Name: "删除规则", Code: "rule:delete", Resource: "rule", Action: "delete", Category: "rule", Type: model.PermissionTypeSystem, Status: model.PermissionStatusActive, IsSystem: true},
		{Name: "规则管理", Code: "rule:manage", Resource: "rule", Action: "manage", Category: "rule", Type: model.PermissionTypeSystem, Status: model.PermissionStatusActive, IsSystem: true},
		
		// 系统管理权限
		{Name: "系统配置", Code: "system:config", Resource: "system", Action: "manage", Category: "system", Type: model.PermissionTypeSystem, Status: model.PermissionStatusActive, IsSystem: true},
		{Name: "系统监控", Code: "system:monitor", Resource: "system", Action: "read", Category: "system", Type: model.PermissionTypeSystem, Status: model.PermissionStatusActive, IsSystem: true},
		{Name: "系统管理", Code: "system:manage", Resource: "system", Action: "manage", Category: "system", Type: model.PermissionTypeSystem, Status: model.PermissionStatusActive, IsSystem: true},
	}
	
	// 检查并创建不存在的权限
	var newPermissions []*model.Permission
	for _, permission := range systemPermissions {
		existing, err := s.permissionRepo.GetByCode(ctx, permission.Code)
		if err != nil && !strings.Contains(err.Error(), "not found") {
			return fmt.Errorf("failed to check permission %s: %w", permission.Code, err)
		}
		
		if existing == nil {
			newPermissions = append(newPermissions, permission)
		}
	}
	
	if len(newPermissions) > 0 {
		if err := s.permissionRepo.BatchCreate(ctx, newPermissions); err != nil {
			return fmt.Errorf("failed to create system permissions: %w", err)
		}
		logger.L.Info("System permissions initialized", zap.Int("count", len(newPermissions)))
	} else {
		logger.L.Info("All system permissions already exist")
	}
	
	return nil
}