package repository

import (
	"context"
	"fmt"

	"alert_agent/internal/model"
	"alert_agent/internal/pkg/database"
	"alert_agent/internal/pkg/logger"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// RoleRepository 角色仓储接口
type RoleRepository interface {
	// 基础CRUD操作
	Create(ctx context.Context, role *model.Role) error
	GetByID(ctx context.Context, id string) (*model.Role, error)
	GetByCode(ctx context.Context, code string) (*model.Role, error)
	Update(ctx context.Context, role *model.Role) error
	Delete(ctx context.Context, id string) error
	
	// 查询操作
	List(ctx context.Context, filters map[string]interface{}, offset, limit int) ([]*model.Role, int64, error)
	GetByType(ctx context.Context, roleType string) ([]*model.Role, error)
	GetByStatus(ctx context.Context, status string) ([]*model.Role, error)
	GetSystemRoles(ctx context.Context) ([]*model.Role, error)
	
	// 批量操作
	BatchCreate(ctx context.Context, roles []*model.Role) error
	BatchUpdate(ctx context.Context, ids []string, updates map[string]interface{}) error
	BatchDelete(ctx context.Context, ids []string) error
	
	// 权限关联操作
	AssignPermissions(ctx context.Context, roleID string, permissionIDs []string) error
	RemovePermissions(ctx context.Context, roleID string, permissionIDs []string) error
	GetRolePermissions(ctx context.Context, roleID string) ([]*model.Permission, error)
	SyncPermissions(ctx context.Context, roleID string, permissionIDs []string) error
	
	// 用户关联操作
	GetRoleUsers(ctx context.Context, roleID string) ([]*model.User, error)
	GetUserRoles(ctx context.Context, userID string) ([]*model.Role, error)
	
	// 统计操作
	GetStats(ctx context.Context) (*model.RoleStats, error)
	Count(ctx context.Context, filters map[string]interface{}) (int64, error)
}

// roleRepository 角色仓储实现
type roleRepository struct {
	db *gorm.DB
}

// NewRoleRepository 创建角色仓储实例
func NewRoleRepository() RoleRepository {
	return &roleRepository{
		db: database.DB,
	}
}

// Create 创建角色
func (r *roleRepository) Create(ctx context.Context, role *model.Role) error {
	logger.L.Debug("Creating role", zap.String("code", role.Code))
	
	if err := r.db.WithContext(ctx).Create(role).Error; err != nil {
		logger.L.Error("Failed to create role", zap.Error(err))
		return fmt.Errorf("failed to create role: %w", err)
	}
	
	logger.L.Info("Role created successfully", zap.String("id", role.ID))
	return nil
}

// GetByID 根据ID获取角色
func (r *roleRepository) GetByID(ctx context.Context, id string) (*model.Role, error) {
	var role model.Role
	
	if err := r.db.WithContext(ctx).Preload("Permissions").Preload("Users").First(&role, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("role not found")
		}
		logger.L.Error("Failed to get role by ID", zap.Error(err))
		return nil, fmt.Errorf("failed to get role: %w", err)
	}
	
	return &role, nil
}

// GetByCode 根据代码获取角色
func (r *roleRepository) GetByCode(ctx context.Context, code string) (*model.Role, error) {
	var role model.Role
	
	if err := r.db.WithContext(ctx).Preload("Permissions").Preload("Users").First(&role, "code = ?", code).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("role not found")
		}
		logger.L.Error("Failed to get role by code", zap.Error(err))
		return nil, fmt.Errorf("failed to get role: %w", err)
	}
	
	return &role, nil
}

// Update 更新角色
func (r *roleRepository) Update(ctx context.Context, role *model.Role) error {
	logger.L.Debug("Updating role", zap.String("id", role.ID))
	
	if err := r.db.WithContext(ctx).Save(role).Error; err != nil {
		logger.L.Error("Failed to update role", zap.Error(err))
		return fmt.Errorf("failed to update role: %w", err)
	}
	
	logger.L.Info("Role updated successfully", zap.String("id", role.ID))
	return nil
}

// Delete 删除角色
func (r *roleRepository) Delete(ctx context.Context, id string) error {
	logger.L.Debug("Deleting role", zap.String("id", id))
	
	// 检查是否为系统角色
	var role model.Role
	if err := r.db.WithContext(ctx).First(&role, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("role not found")
		}
		return fmt.Errorf("failed to get role: %w", err)
	}
	
	if role.IsSystemRole() {
		return fmt.Errorf("cannot delete system role")
	}
	
	// 检查是否有用户关联
	var userCount int64
	if err := r.db.WithContext(ctx).Table("user_roles").Where("role_id = ?", id).Count(&userCount).Error; err != nil {
		return fmt.Errorf("failed to check user associations: %w", err)
	}
	
	if userCount > 0 {
		return fmt.Errorf("cannot delete role with associated users")
	}
	
	// 开启事务删除角色及其关联关系
	tx := r.db.WithContext(ctx).Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()
	
	// 删除角色权限关联
	if err := tx.Exec("DELETE FROM role_permissions WHERE role_id = ?", id).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to delete role permission associations: %w", err)
	}
	
	// 删除角色
	if err := tx.Delete(&role).Error; err != nil {
		tx.Rollback()
		logger.L.Error("Failed to delete role", zap.Error(err))
		return fmt.Errorf("failed to delete role: %w", err)
	}
	
	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	
	logger.L.Info("Role deleted successfully", zap.String("id", id))
	return nil
}

// List 获取角色列表
func (r *roleRepository) List(ctx context.Context, filters map[string]interface{}, offset, limit int) ([]*model.Role, int64, error) {
	var roles []*model.Role
	var total int64
	
	query := r.db.WithContext(ctx).Model(&model.Role{})
	
	// 应用过滤条件
	for key, value := range filters {
		switch key {
		case "name":
			if v, ok := value.(string); ok && v != "" {
				query = query.Where("name LIKE ?", "%"+v+"%")
			}
		case "code":
			if v, ok := value.(string); ok && v != "" {
				query = query.Where("code LIKE ?", "%"+v+"%")
			}
		case "type":
			if v, ok := value.(string); ok && v != "" {
				query = query.Where("type = ?", v)
			}
		case "status":
			if v, ok := value.(string); ok && v != "" {
				query = query.Where("status = ?", v)
			}
		case "is_system":
			if v, ok := value.(bool); ok {
				query = query.Where("is_system = ?", v)
			}
		}
	}
	
	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		logger.L.Error("Failed to count roles", zap.Error(err))
		return nil, 0, fmt.Errorf("failed to count roles: %w", err)
	}
	
	// 获取分页数据
	if err := query.Preload("Permissions").Offset(offset).Limit(limit).Order("created_at DESC").Find(&roles).Error; err != nil {
		logger.L.Error("Failed to list roles", zap.Error(err))
		return nil, 0, fmt.Errorf("failed to list roles: %w", err)
	}
	
	return roles, total, nil
}

// GetByType 根据类型获取角色
func (r *roleRepository) GetByType(ctx context.Context, roleType string) ([]*model.Role, error) {
	var roles []*model.Role
	
	if err := r.db.WithContext(ctx).Where("type = ?", roleType).Find(&roles).Error; err != nil {
		logger.L.Error("Failed to get roles by type", zap.Error(err))
		return nil, fmt.Errorf("failed to get roles by type: %w", err)
	}
	
	return roles, nil
}

// GetByStatus 根据状态获取角色
func (r *roleRepository) GetByStatus(ctx context.Context, status string) ([]*model.Role, error) {
	var roles []*model.Role
	
	if err := r.db.WithContext(ctx).Where("status = ?", status).Find(&roles).Error; err != nil {
		logger.L.Error("Failed to get roles by status", zap.Error(err))
		return nil, fmt.Errorf("failed to get roles by status: %w", err)
	}
	
	return roles, nil
}

// GetSystemRoles 获取系统角色
func (r *roleRepository) GetSystemRoles(ctx context.Context) ([]*model.Role, error) {
	var roles []*model.Role
	
	if err := r.db.WithContext(ctx).Where("is_system = ?", true).Find(&roles).Error; err != nil {
		logger.L.Error("Failed to get system roles", zap.Error(err))
		return nil, fmt.Errorf("failed to get system roles: %w", err)
	}
	
	return roles, nil
}

// BatchCreate 批量创建角色
func (r *roleRepository) BatchCreate(ctx context.Context, roles []*model.Role) error {
	logger.L.Debug("Batch creating roles", zap.Int("count", len(roles)))
	
	if err := r.db.WithContext(ctx).CreateInBatches(roles, 100).Error; err != nil {
		logger.L.Error("Failed to batch create roles", zap.Error(err))
		return fmt.Errorf("failed to batch create roles: %w", err)
	}
	
	logger.L.Info("Roles batch created successfully", zap.Int("count", len(roles)))
	return nil
}

// BatchUpdate 批量更新角色
func (r *roleRepository) BatchUpdate(ctx context.Context, ids []string, updates map[string]interface{}) error {
	logger.L.Debug("Batch updating roles", zap.Int("count", len(ids)))
	
	if err := r.db.WithContext(ctx).Model(&model.Role{}).Where("id IN ?", ids).Updates(updates).Error; err != nil {
		logger.L.Error("Failed to batch update roles", zap.Error(err))
		return fmt.Errorf("failed to batch update roles: %w", err)
	}
	
	logger.L.Info("Roles batch updated successfully", zap.Int("count", len(ids)))
	return nil
}

// BatchDelete 批量删除角色
func (r *roleRepository) BatchDelete(ctx context.Context, ids []string) error {
	logger.L.Debug("Batch deleting roles", zap.Int("count", len(ids)))
	
	// 检查是否包含系统角色
	var systemCount int64
	if err := r.db.WithContext(ctx).Model(&model.Role{}).Where("id IN ? AND is_system = ?", ids, true).Count(&systemCount).Error; err != nil {
		return fmt.Errorf("failed to check system roles: %w", err)
	}
	
	if systemCount > 0 {
		return fmt.Errorf("cannot delete system roles")
	}
	
	// 检查是否有用户关联
	var userCount int64
	if err := r.db.WithContext(ctx).Table("user_roles").Where("role_id IN ?", ids).Count(&userCount).Error; err != nil {
		return fmt.Errorf("failed to check user associations: %w", err)
	}
	
	if userCount > 0 {
		return fmt.Errorf("cannot delete roles with associated users")
	}
	
	// 开启事务
	tx := r.db.WithContext(ctx).Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()
	
	// 删除角色权限关联
	if err := tx.Exec("DELETE FROM role_permissions WHERE role_id IN ?", ids).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to delete role permission associations: %w", err)
	}
	
	// 删除角色
	if err := tx.Where("id IN ?", ids).Delete(&model.Role{}).Error; err != nil {
		tx.Rollback()
		logger.L.Error("Failed to batch delete roles", zap.Error(err))
		return fmt.Errorf("failed to batch delete roles: %w", err)
	}
	
	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	
	logger.L.Info("Roles batch deleted successfully", zap.Int("count", len(ids)))
	return nil
}

// AssignPermissions 为角色分配权限
func (r *roleRepository) AssignPermissions(ctx context.Context, roleID string, permissionIDs []string) error {
	logger.L.Debug("Assigning permissions to role", zap.String("role_id", roleID), zap.Int("permission_count", len(permissionIDs)))
	
	// 检查角色是否存在
	var role model.Role
	if err := r.db.WithContext(ctx).First(&role, "id = ?", roleID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("role not found")
		}
		return fmt.Errorf("failed to get role: %w", err)
	}
	
	// 检查权限是否存在
	var existingPermissions []string
	if err := r.db.WithContext(ctx).Model(&model.Permission{}).Where("id IN ?", permissionIDs).Pluck("id", &existingPermissions).Error; err != nil {
		return fmt.Errorf("failed to check permissions: %w", err)
	}
	
	if len(existingPermissions) != len(permissionIDs) {
		return fmt.Errorf("some permissions not found")
	}
	
	// 批量插入角色权限关联
	var associations []map[string]interface{}
	for _, permissionID := range permissionIDs {
		associations = append(associations, map[string]interface{}{
			"role_id":       roleID,
			"permission_id": permissionID,
		})
	}
	
	if err := r.db.WithContext(ctx).Table("role_permissions").Create(associations).Error; err != nil {
		logger.L.Error("Failed to assign permissions to role", zap.Error(err))
		return fmt.Errorf("failed to assign permissions: %w", err)
	}
	
	logger.L.Info("Permissions assigned to role successfully", zap.String("role_id", roleID))
	return nil
}

// RemovePermissions 移除角色权限
func (r *roleRepository) RemovePermissions(ctx context.Context, roleID string, permissionIDs []string) error {
	logger.L.Debug("Removing permissions from role", zap.String("role_id", roleID), zap.Int("permission_count", len(permissionIDs)))
	
	if err := r.db.WithContext(ctx).Exec("DELETE FROM role_permissions WHERE role_id = ? AND permission_id IN ?", roleID, permissionIDs).Error; err != nil {
		logger.L.Error("Failed to remove permissions from role", zap.Error(err))
		return fmt.Errorf("failed to remove permissions: %w", err)
	}
	
	logger.L.Info("Permissions removed from role successfully", zap.String("role_id", roleID))
	return nil
}

// GetRolePermissions 获取角色权限
func (r *roleRepository) GetRolePermissions(ctx context.Context, roleID string) ([]*model.Permission, error) {
	var permissions []*model.Permission
	
	if err := r.db.WithContext(ctx).Table("permissions").
		Joins("JOIN role_permissions ON permissions.id = role_permissions.permission_id").
		Where("role_permissions.role_id = ?", roleID).
		Find(&permissions).Error; err != nil {
		logger.L.Error("Failed to get role permissions", zap.Error(err))
		return nil, fmt.Errorf("failed to get role permissions: %w", err)
	}
	
	return permissions, nil
}

// SyncPermissions 同步角色权限
func (r *roleRepository) SyncPermissions(ctx context.Context, roleID string, permissionIDs []string) error {
	logger.L.Debug("Syncing permissions for role", zap.String("role_id", roleID), zap.Int("permission_count", len(permissionIDs)))
	
	// 开启事务
	tx := r.db.WithContext(ctx).Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()
	
	// 删除现有权限关联
	if err := tx.Exec("DELETE FROM role_permissions WHERE role_id = ?", roleID).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to delete existing permissions: %w", err)
	}
	
	// 添加新权限关联
	if len(permissionIDs) > 0 {
		var associations []map[string]interface{}
		for _, permissionID := range permissionIDs {
			associations = append(associations, map[string]interface{}{
				"role_id":       roleID,
				"permission_id": permissionID,
			})
		}
		
		if err := tx.Table("role_permissions").Create(associations).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to create new permissions: %w", err)
		}
	}
	
	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	
	logger.L.Info("Permissions synced for role successfully", zap.String("role_id", roleID))
	return nil
}

// GetRoleUsers 获取角色关联的用户
func (r *roleRepository) GetRoleUsers(ctx context.Context, roleID string) ([]*model.User, error) {
	var users []*model.User
	
	if err := r.db.WithContext(ctx).Table("users").
		Joins("JOIN user_roles ON users.id = user_roles.user_id").
		Where("user_roles.role_id = ?", roleID).
		Find(&users).Error; err != nil {
		logger.L.Error("Failed to get role users", zap.Error(err))
		return nil, fmt.Errorf("failed to get role users: %w", err)
	}
	
	return users, nil
}

// GetUserRoles 获取用户角色
func (r *roleRepository) GetUserRoles(ctx context.Context, userID string) ([]*model.Role, error) {
	var roles []*model.Role
	
	if err := r.db.WithContext(ctx).Table("roles").
		Joins("JOIN user_roles ON roles.id = user_roles.role_id").
		Where("user_roles.user_id = ?", userID).
		Find(&roles).Error; err != nil {
		logger.L.Error("Failed to get user roles", zap.Error(err))
		return nil, fmt.Errorf("failed to get user roles: %w", err)
	}
	
	return roles, nil
}

// GetStats 获取角色统计
func (r *roleRepository) GetStats(ctx context.Context) (*model.RoleStats, error) {
	stats := &model.RoleStats{}
	
	// 总角色数
	if err := r.db.WithContext(ctx).Model(&model.Role{}).Count(&stats.Total).Error; err != nil {
		return nil, fmt.Errorf("failed to count total roles: %w", err)
	}
	
	// 系统角色数
	if err := r.db.WithContext(ctx).Model(&model.Role{}).Where("is_system = ?", true).Count(&stats.System).Error; err != nil {
		return nil, fmt.Errorf("failed to count system roles: %w", err)
	}
	
	// 自定义角色数
	if err := r.db.WithContext(ctx).Model(&model.Role{}).Where("is_system = ?", false).Count(&stats.Custom).Error; err != nil {
		return nil, fmt.Errorf("failed to count custom roles: %w", err)
	}
	
	// 激活角色数
	if err := r.db.WithContext(ctx).Model(&model.Role{}).Where("status = ?", model.RoleStatusActive).Count(&stats.Active).Error; err != nil {
		return nil, fmt.Errorf("failed to count active roles: %w", err)
	}
	
	return stats, nil
}

// Count 统计角色数量
func (r *roleRepository) Count(ctx context.Context, filters map[string]interface{}) (int64, error) {
	var count int64
	query := r.db.WithContext(ctx).Model(&model.Role{})
	
	// 应用过滤条件
	for key, value := range filters {
		switch key {
		case "type":
			if v, ok := value.(string); ok && v != "" {
				query = query.Where("type = ?", v)
			}
		case "status":
			if v, ok := value.(string); ok && v != "" {
				query = query.Where("status = ?", v)
			}
		case "is_system":
			if v, ok := value.(bool); ok {
				query = query.Where("is_system = ?", v)
			}
		}
	}
	
	if err := query.Count(&count).Error; err != nil {
		logger.L.Error("Failed to count roles", zap.Error(err))
		return 0, fmt.Errorf("failed to count roles: %w", err)
	}
	
	return count, nil
}