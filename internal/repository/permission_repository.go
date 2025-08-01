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

// PermissionRepository 权限仓储接口
type PermissionRepository interface {
	// 基础CRUD操作
	Create(ctx context.Context, permission *model.Permission) error
	GetByID(ctx context.Context, id string) (*model.Permission, error)
	GetByCode(ctx context.Context, code string) (*model.Permission, error)
	Update(ctx context.Context, permission *model.Permission) error
	Delete(ctx context.Context, id string) error
	
	// 查询操作
	List(ctx context.Context, filters map[string]interface{}, offset, limit int) ([]*model.Permission, int64, error)
	GetByCategory(ctx context.Context, category string) ([]*model.Permission, error)
	GetByType(ctx context.Context, permissionType string) ([]*model.Permission, error)
	GetByStatus(ctx context.Context, status string) ([]*model.Permission, error)
	
	// 批量操作
	BatchCreate(ctx context.Context, permissions []*model.Permission) error
	BatchUpdate(ctx context.Context, ids []string, updates map[string]interface{}) error
	BatchDelete(ctx context.Context, ids []string) error
	
	// 关联查询
	GetPermissionRoles(ctx context.Context, permissionID string) ([]*model.Role, error)
	
	// 统计操作
	GetStats(ctx context.Context) (*model.PermissionStats, error)
	Count(ctx context.Context, filters map[string]interface{}) (int64, error)
}

// permissionRepository 权限仓储实现
type permissionRepository struct {
	db *gorm.DB
}

// NewPermissionRepository 创建权限仓储实例
func NewPermissionRepository() PermissionRepository {
	return &permissionRepository{
		db: database.DB,
	}
}

// Create 创建权限
func (r *permissionRepository) Create(ctx context.Context, permission *model.Permission) error {
	logger.L.Debug("Creating permission", zap.String("code", permission.Code))
	
	if err := r.db.WithContext(ctx).Create(permission).Error; err != nil {
		logger.L.Error("Failed to create permission", zap.Error(err))
		return fmt.Errorf("failed to create permission: %w", err)
	}
	
	logger.L.Info("Permission created successfully", zap.String("id", permission.ID))
	return nil
}

// GetByID 根据ID获取权限
func (r *permissionRepository) GetByID(ctx context.Context, id string) (*model.Permission, error) {
	var permission model.Permission
	
	if err := r.db.WithContext(ctx).Preload("Roles").First(&permission, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("permission not found")
		}
		logger.L.Error("Failed to get permission by ID", zap.Error(err))
		return nil, fmt.Errorf("failed to get permission: %w", err)
	}
	
	return &permission, nil
}

// GetByCode 根据代码获取权限
func (r *permissionRepository) GetByCode(ctx context.Context, code string) (*model.Permission, error) {
	var permission model.Permission
	
	if err := r.db.WithContext(ctx).Preload("Roles").First(&permission, "code = ?", code).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("permission not found")
		}
		logger.L.Error("Failed to get permission by code", zap.Error(err))
		return nil, fmt.Errorf("failed to get permission: %w", err)
	}
	
	return &permission, nil
}

// Update 更新权限
func (r *permissionRepository) Update(ctx context.Context, permission *model.Permission) error {
	logger.L.Debug("Updating permission", zap.String("id", permission.ID))
	
	if err := r.db.WithContext(ctx).Save(permission).Error; err != nil {
		logger.L.Error("Failed to update permission", zap.Error(err))
		return fmt.Errorf("failed to update permission: %w", err)
	}
	
	logger.L.Info("Permission updated successfully", zap.String("id", permission.ID))
	return nil
}

// Delete 删除权限
func (r *permissionRepository) Delete(ctx context.Context, id string) error {
	logger.L.Debug("Deleting permission", zap.String("id", id))
	
	// 检查是否为系统权限
	var permission model.Permission
	if err := r.db.WithContext(ctx).First(&permission, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("permission not found")
		}
		return fmt.Errorf("failed to get permission: %w", err)
	}
	
	if permission.IsSystemPermission() {
		return fmt.Errorf("cannot delete system permission")
	}
	
	// 开启事务删除权限及其关联关系
	tx := r.db.WithContext(ctx).Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()
	
	// 删除角色权限关联
	if err := tx.Exec("DELETE FROM role_permissions WHERE permission_id = ?", id).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to delete role permission associations: %w", err)
	}
	
	// 删除权限
	if err := tx.Delete(&permission).Error; err != nil {
		tx.Rollback()
		logger.L.Error("Failed to delete permission", zap.Error(err))
		return fmt.Errorf("failed to delete permission: %w", err)
	}
	
	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	
	logger.L.Info("Permission deleted successfully", zap.String("id", id))
	return nil
}

// List 获取权限列表
func (r *permissionRepository) List(ctx context.Context, filters map[string]interface{}, offset, limit int) ([]*model.Permission, int64, error) {
	var permissions []*model.Permission
	var total int64
	
	query := r.db.WithContext(ctx).Model(&model.Permission{})
	
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
		case "category":
			if v, ok := value.(string); ok && v != "" {
				query = query.Where("category = ?", v)
			}
		case "type":
			if v, ok := value.(string); ok && v != "" {
				query = query.Where("type = ?", v)
			}
		case "status":
			if v, ok := value.(string); ok && v != "" {
				query = query.Where("status = ?", v)
			}
		case "resource":
			if v, ok := value.(string); ok && v != "" {
				query = query.Where("resource LIKE ?", "%"+v+"%")
			}
		case "action":
			if v, ok := value.(string); ok && v != "" {
				query = query.Where("action = ?", v)
			}
		}
	}
	
	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		logger.L.Error("Failed to count permissions", zap.Error(err))
		return nil, 0, fmt.Errorf("failed to count permissions: %w", err)
	}
	
	// 获取分页数据
	if err := query.Preload("Roles").Offset(offset).Limit(limit).Order("created_at DESC").Find(&permissions).Error; err != nil {
		logger.L.Error("Failed to list permissions", zap.Error(err))
		return nil, 0, fmt.Errorf("failed to list permissions: %w", err)
	}
	
	return permissions, total, nil
}

// GetByCategory 根据分类获取权限
func (r *permissionRepository) GetByCategory(ctx context.Context, category string) ([]*model.Permission, error) {
	var permissions []*model.Permission
	
	if err := r.db.WithContext(ctx).Where("category = ? AND status = ?", category, model.PermissionStatusActive).Find(&permissions).Error; err != nil {
		logger.L.Error("Failed to get permissions by category", zap.Error(err))
		return nil, fmt.Errorf("failed to get permissions by category: %w", err)
	}
	
	return permissions, nil
}

// GetByType 根据类型获取权限
func (r *permissionRepository) GetByType(ctx context.Context, permissionType string) ([]*model.Permission, error) {
	var permissions []*model.Permission
	
	if err := r.db.WithContext(ctx).Where("type = ?", permissionType).Find(&permissions).Error; err != nil {
		logger.L.Error("Failed to get permissions by type", zap.Error(err))
		return nil, fmt.Errorf("failed to get permissions by type: %w", err)
	}
	
	return permissions, nil
}

// GetByStatus 根据状态获取权限
func (r *permissionRepository) GetByStatus(ctx context.Context, status string) ([]*model.Permission, error) {
	var permissions []*model.Permission
	
	if err := r.db.WithContext(ctx).Where("status = ?", status).Find(&permissions).Error; err != nil {
		logger.L.Error("Failed to get permissions by status", zap.Error(err))
		return nil, fmt.Errorf("failed to get permissions by status: %w", err)
	}
	
	return permissions, nil
}

// BatchCreate 批量创建权限
func (r *permissionRepository) BatchCreate(ctx context.Context, permissions []*model.Permission) error {
	logger.L.Debug("Batch creating permissions", zap.Int("count", len(permissions)))
	
	if err := r.db.WithContext(ctx).CreateInBatches(permissions, 100).Error; err != nil {
		logger.L.Error("Failed to batch create permissions", zap.Error(err))
		return fmt.Errorf("failed to batch create permissions: %w", err)
	}
	
	logger.L.Info("Permissions batch created successfully", zap.Int("count", len(permissions)))
	return nil
}

// BatchUpdate 批量更新权限
func (r *permissionRepository) BatchUpdate(ctx context.Context, ids []string, updates map[string]interface{}) error {
	logger.L.Debug("Batch updating permissions", zap.Int("count", len(ids)))
	
	if err := r.db.WithContext(ctx).Model(&model.Permission{}).Where("id IN ?", ids).Updates(updates).Error; err != nil {
		logger.L.Error("Failed to batch update permissions", zap.Error(err))
		return fmt.Errorf("failed to batch update permissions: %w", err)
	}
	
	logger.L.Info("Permissions batch updated successfully", zap.Int("count", len(ids)))
	return nil
}

// BatchDelete 批量删除权限
func (r *permissionRepository) BatchDelete(ctx context.Context, ids []string) error {
	logger.L.Debug("Batch deleting permissions", zap.Int("count", len(ids)))
	
	// 检查是否包含系统权限
	var systemCount int64
	if err := r.db.WithContext(ctx).Model(&model.Permission{}).Where("id IN ? AND (type = ? OR is_system = ?)", ids, model.PermissionTypeSystem, true).Count(&systemCount).Error; err != nil {
		return fmt.Errorf("failed to check system permissions: %w", err)
	}
	
	if systemCount > 0 {
		return fmt.Errorf("cannot delete system permissions")
	}
	
	// 开启事务
	tx := r.db.WithContext(ctx).Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()
	
	// 删除角色权限关联
	if err := tx.Exec("DELETE FROM role_permissions WHERE permission_id IN ?", ids).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to delete role permission associations: %w", err)
	}
	
	// 删除权限
	if err := tx.Where("id IN ?", ids).Delete(&model.Permission{}).Error; err != nil {
		tx.Rollback()
		logger.L.Error("Failed to batch delete permissions", zap.Error(err))
		return fmt.Errorf("failed to batch delete permissions: %w", err)
	}
	
	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	
	logger.L.Info("Permissions batch deleted successfully", zap.Int("count", len(ids)))
	return nil
}

// GetPermissionRoles 获取权限关联的角色
func (r *permissionRepository) GetPermissionRoles(ctx context.Context, permissionID string) ([]*model.Role, error) {
	var roles []*model.Role
	
	if err := r.db.WithContext(ctx).Table("roles").
		Joins("JOIN role_permissions ON roles.id = role_permissions.role_id").
		Where("role_permissions.permission_id = ?", permissionID).
		Find(&roles).Error; err != nil {
		logger.L.Error("Failed to get permission roles", zap.Error(err))
		return nil, fmt.Errorf("failed to get permission roles: %w", err)
	}
	
	return roles, nil
}

// GetStats 获取权限统计
func (r *permissionRepository) GetStats(ctx context.Context) (*model.PermissionStats, error) {
	stats := &model.PermissionStats{}
	
	// 总权限数
	if err := r.db.WithContext(ctx).Model(&model.Permission{}).Count(&stats.Total).Error; err != nil {
		return nil, fmt.Errorf("failed to count total permissions: %w", err)
	}
	
	// 系统权限数
	if err := r.db.WithContext(ctx).Model(&model.Permission{}).Where("type = ? OR is_system = ?", model.PermissionTypeSystem, true).Count(&stats.System).Error; err != nil {
		return nil, fmt.Errorf("failed to count system permissions: %w", err)
	}
	
	// 自定义权限数
	if err := r.db.WithContext(ctx).Model(&model.Permission{}).Where("type = ? AND is_system = ?", model.PermissionTypeCustom, false).Count(&stats.Custom).Error; err != nil {
		return nil, fmt.Errorf("failed to count custom permissions: %w", err)
	}
	
	// 激活权限数
	if err := r.db.WithContext(ctx).Model(&model.Permission{}).Where("status = ?", model.PermissionStatusActive).Count(&stats.Active).Error; err != nil {
		return nil, fmt.Errorf("failed to count active permissions: %w", err)
	}
	
	return stats, nil
}

// Count 统计权限数量
func (r *permissionRepository) Count(ctx context.Context, filters map[string]interface{}) (int64, error) {
	var count int64
	query := r.db.WithContext(ctx).Model(&model.Permission{})
	
	// 应用过滤条件
	for key, value := range filters {
		switch key {
		case "category":
			if v, ok := value.(string); ok && v != "" {
				query = query.Where("category = ?", v)
			}
		case "type":
			if v, ok := value.(string); ok && v != "" {
				query = query.Where("type = ?", v)
			}
		case "status":
			if v, ok := value.(string); ok && v != "" {
				query = query.Where("status = ?", v)
			}
		}
	}
	
	if err := query.Count(&count).Error; err != nil {
		logger.L.Error("Failed to count permissions", zap.Error(err))
		return 0, fmt.Errorf("failed to count permissions: %w", err)
	}
	
	return count, nil
}