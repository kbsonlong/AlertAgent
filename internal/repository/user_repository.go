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

// UserRepository 用户仓储接口
type UserRepository interface {
	// 基础CRUD操作
	Create(ctx context.Context, user *model.User) error
	GetByID(ctx context.Context, id string) (*model.User, error)
	GetByUsername(ctx context.Context, username string) (*model.User, error)
	GetByEmail(ctx context.Context, email string) (*model.User, error)
	Update(ctx context.Context, user *model.User) error
	Delete(ctx context.Context, id string) error
	
	// 查询操作
	List(ctx context.Context, filters map[string]interface{}, offset, limit int) ([]*model.User, int64, error)
	GetByRole(ctx context.Context, role string) ([]*model.User, error)
	GetByStatus(ctx context.Context, status string) ([]*model.User, error)
	GetByDepartment(ctx context.Context, department string) ([]*model.User, error)
	
	// 批量操作
	BatchCreate(ctx context.Context, users []*model.User) error
	BatchUpdate(ctx context.Context, ids []string, updates map[string]interface{}) error
	BatchDelete(ctx context.Context, ids []string) error
	
	// 角色关联操作
	AssignRoles(ctx context.Context, userID string, roleIDs []string) error
	RemoveRoles(ctx context.Context, userID string, roleIDs []string) error
	SyncRoles(ctx context.Context, userID string, roleIDs []string) error
	GetUserRoles(ctx context.Context, userID string) ([]*model.Role, error)
	GetRoleUsers(ctx context.Context, roleID string) ([]*model.User, error)
	
	// 权限查询
	GetUserPermissions(ctx context.Context, userID string) ([]*model.Permission, error)
	CheckUserPermission(ctx context.Context, userID, permissionCode string) (bool, error)
	
	// 统计操作
	GetStats(ctx context.Context) (*model.UserStats, error)
	Count(ctx context.Context, filters map[string]interface{}) (int64, error)
	
	// 验证操作
	CheckUsernameExists(ctx context.Context, username string, excludeID string) (bool, error)
	CheckEmailExists(ctx context.Context, email string, excludeID string) (bool, error)
}

// userRepository 用户仓储实现
type userRepository struct {
	db *gorm.DB
}

// NewUserRepository 创建用户仓储实例
func NewUserRepository() UserRepository {
	return &userRepository{
		db: database.DB,
	}
}

// Create 创建用户
func (r *userRepository) Create(ctx context.Context, user *model.User) error {
	logger.L.Debug("Creating user", zap.String("username", user.Username))
	
	if err := r.db.WithContext(ctx).Create(user).Error; err != nil {
		logger.L.Error("Failed to create user", zap.Error(err))
		return fmt.Errorf("failed to create user: %w", err)
	}
	
	logger.L.Info("User created successfully", zap.String("id", user.ID))
	return nil
}

// GetByID 根据ID获取用户
func (r *userRepository) GetByID(ctx context.Context, id string) (*model.User, error) {
	var user model.User
	
	if err := r.db.WithContext(ctx).Preload("Roles").Where("id = ?", id).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("user not found")
		}
		logger.L.Error("Failed to get user by ID", zap.Error(err))
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	
	return &user, nil
}

// GetByUsername 根据用户名获取用户
func (r *userRepository) GetByUsername(ctx context.Context, username string) (*model.User, error) {
	var user model.User
	
	if err := r.db.WithContext(ctx).Preload("Roles").Where("username = ?", username).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("user not found")
		}
		logger.L.Error("Failed to get user by username", zap.Error(err))
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	
	return &user, nil
}

// GetByEmail 根据邮箱获取用户
func (r *userRepository) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	var user model.User
	
	if err := r.db.WithContext(ctx).Preload("Roles").Where("email = ?", email).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("user not found")
		}
		logger.L.Error("Failed to get user by email", zap.Error(err))
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	
	return &user, nil
}

// Update 更新用户
func (r *userRepository) Update(ctx context.Context, user *model.User) error {
	logger.L.Debug("Updating user", zap.String("id", user.ID))
	
	if err := r.db.WithContext(ctx).Save(user).Error; err != nil {
		logger.L.Error("Failed to update user", zap.Error(err))
		return fmt.Errorf("failed to update user: %w", err)
	}
	
	logger.L.Info("User updated successfully", zap.String("id", user.ID))
	return nil
}

// Delete 删除用户（软删除）
func (r *userRepository) Delete(ctx context.Context, id string) error {
	logger.L.Debug("Deleting user", zap.String("id", id))
	
	if err := r.db.WithContext(ctx).Delete(&model.User{}, "id = ?", id).Error; err != nil {
		logger.L.Error("Failed to delete user", zap.Error(err))
		return fmt.Errorf("failed to delete user: %w", err)
	}
	
	logger.L.Info("User deleted successfully", zap.String("id", id))
	return nil
}

// List 获取用户列表
func (r *userRepository) List(ctx context.Context, filters map[string]interface{}, offset, limit int) ([]*model.User, int64, error) {
	var users []*model.User
	var total int64
	
	query := r.db.WithContext(ctx).Model(&model.User{})
	
	// 应用筛选条件
	for key, value := range filters {
		switch key {
		case "username":
			if v, ok := value.(string); ok && v != "" {
				query = query.Where("username LIKE ?", "%"+v+"%")
			}
		case "email":
			if v, ok := value.(string); ok && v != "" {
				query = query.Where("email LIKE ?", "%"+v+"%")
			}
		case "role":
			if v, ok := value.(string); ok && v != "" {
				query = query.Where("role = ?", v)
			}
		case "status":
			if v, ok := value.(string); ok && v != "" {
				query = query.Where("status = ?", v)
			}
		case "department":
			if v, ok := value.(string); ok && v != "" {
				query = query.Where("department LIKE ?", "%"+v+"%")
			}
		case "full_name":
			if v, ok := value.(string); ok && v != "" {
				query = query.Where("full_name LIKE ?", "%"+v+"%")
			}
		}
	}
	
	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		logger.L.Error("Failed to count users", zap.Error(err))
		return nil, 0, fmt.Errorf("failed to count users: %w", err)
	}
	
	// 获取数据
	if err := query.Preload("Roles").Offset(offset).Limit(limit).Order("created_at DESC").Find(&users).Error; err != nil {
		logger.L.Error("Failed to list users", zap.Error(err))
		return nil, 0, fmt.Errorf("failed to list users: %w", err)
	}
	
	return users, total, nil
}

// GetByRole 根据角色获取用户
func (r *userRepository) GetByRole(ctx context.Context, role string) ([]*model.User, error) {
	var users []*model.User
	
	if err := r.db.WithContext(ctx).Preload("Roles").Where("role = ?", role).Find(&users).Error; err != nil {
		logger.L.Error("Failed to get users by role", zap.Error(err))
		return nil, fmt.Errorf("failed to get users by role: %w", err)
	}
	
	return users, nil
}

// GetByStatus 根据状态获取用户
func (r *userRepository) GetByStatus(ctx context.Context, status string) ([]*model.User, error) {
	var users []*model.User
	
	if err := r.db.WithContext(ctx).Preload("Roles").Where("status = ?", status).Find(&users).Error; err != nil {
		logger.L.Error("Failed to get users by status", zap.Error(err))
		return nil, fmt.Errorf("failed to get users by status: %w", err)
	}
	
	return users, nil
}

// GetByDepartment 根据部门获取用户
func (r *userRepository) GetByDepartment(ctx context.Context, department string) ([]*model.User, error) {
	var users []*model.User
	
	if err := r.db.WithContext(ctx).Preload("Roles").Where("department = ?", department).Find(&users).Error; err != nil {
		logger.L.Error("Failed to get users by department", zap.Error(err))
		return nil, fmt.Errorf("failed to get users by department: %w", err)
	}
	
	return users, nil
}

// BatchCreate 批量创建用户
func (r *userRepository) BatchCreate(ctx context.Context, users []*model.User) error {
	logger.L.Debug("Batch creating users", zap.Int("count", len(users)))
	
	if err := r.db.WithContext(ctx).CreateInBatches(users, 100).Error; err != nil {
		logger.L.Error("Failed to batch create users", zap.Error(err))
		return fmt.Errorf("failed to batch create users: %w", err)
	}
	
	logger.L.Info("Users batch created successfully", zap.Int("count", len(users)))
	return nil
}

// BatchUpdate 批量更新用户
func (r *userRepository) BatchUpdate(ctx context.Context, ids []string, updates map[string]interface{}) error {
	logger.L.Debug("Batch updating users", zap.Strings("ids", ids))
	
	if err := r.db.WithContext(ctx).Model(&model.User{}).Where("id IN ?", ids).Updates(updates).Error; err != nil {
		logger.L.Error("Failed to batch update users", zap.Error(err))
		return fmt.Errorf("failed to batch update users: %w", err)
	}
	
	logger.L.Info("Users batch updated successfully", zap.Strings("ids", ids))
	return nil
}

// BatchDelete 批量删除用户
func (r *userRepository) BatchDelete(ctx context.Context, ids []string) error {
	logger.L.Debug("Batch deleting users", zap.Strings("ids", ids))
	
	if err := r.db.WithContext(ctx).Delete(&model.User{}, "id IN ?", ids).Error; err != nil {
		logger.L.Error("Failed to batch delete users", zap.Error(err))
		return fmt.Errorf("failed to batch delete users: %w", err)
	}
	
	logger.L.Info("Users batch deleted successfully", zap.Strings("ids", ids))
	return nil
}

// AssignRoles 为用户分配角色
func (r *userRepository) AssignRoles(ctx context.Context, userID string, roleIDs []string) error {
	logger.L.Debug("Assigning roles to user", zap.String("user_id", userID), zap.Strings("role_ids", roleIDs))
	
	// 获取用户
	var user model.User
	if err := r.db.WithContext(ctx).First(&user, "id = ?", userID).Error; err != nil {
		return fmt.Errorf("user not found: %w", err)
	}
	
	// 获取角色
	var roles []model.Role
	if err := r.db.WithContext(ctx).Find(&roles, "id IN ?", roleIDs).Error; err != nil {
		return fmt.Errorf("failed to find roles: %w", err)
	}
	
	// 分配角色
	if err := r.db.WithContext(ctx).Model(&user).Association("Roles").Append(&roles); err != nil {
		logger.L.Error("Failed to assign roles", zap.Error(err))
		return fmt.Errorf("failed to assign roles: %w", err)
	}
	
	logger.L.Info("Roles assigned successfully", zap.String("user_id", userID))
	return nil
}

// RemoveRoles 移除用户角色
func (r *userRepository) RemoveRoles(ctx context.Context, userID string, roleIDs []string) error {
	logger.L.Debug("Removing roles from user", zap.String("user_id", userID), zap.Strings("role_ids", roleIDs))
	
	// 获取用户
	var user model.User
	if err := r.db.WithContext(ctx).First(&user, "id = ?", userID).Error; err != nil {
		return fmt.Errorf("user not found: %w", err)
	}
	
	// 获取角色
	var roles []model.Role
	if err := r.db.WithContext(ctx).Find(&roles, "id IN ?", roleIDs).Error; err != nil {
		return fmt.Errorf("failed to find roles: %w", err)
	}
	
	// 移除角色
	if err := r.db.WithContext(ctx).Model(&user).Association("Roles").Delete(&roles); err != nil {
		logger.L.Error("Failed to remove roles", zap.Error(err))
		return fmt.Errorf("failed to remove roles: %w", err)
	}
	
	logger.L.Info("Roles removed successfully", zap.String("user_id", userID))
	return nil
}

// SyncRoles 同步用户角色
func (r *userRepository) SyncRoles(ctx context.Context, userID string, roleIDs []string) error {
	logger.L.Debug("Syncing roles for user", zap.String("user_id", userID), zap.Strings("role_ids", roleIDs))
	
	// 获取用户
	var user model.User
	if err := r.db.WithContext(ctx).First(&user, "id = ?", userID).Error; err != nil {
		return fmt.Errorf("user not found: %w", err)
	}
	
	// 获取角色
	var roles []model.Role
	if err := r.db.WithContext(ctx).Find(&roles, "id IN ?", roleIDs).Error; err != nil {
		return fmt.Errorf("failed to find roles: %w", err)
	}
	
	// 同步角色
	if err := r.db.WithContext(ctx).Model(&user).Association("Roles").Replace(&roles); err != nil {
		logger.L.Error("Failed to sync roles", zap.Error(err))
		return fmt.Errorf("failed to sync roles: %w", err)
	}
	
	logger.L.Info("Roles synced successfully", zap.String("user_id", userID))
	return nil
}

// GetUserRoles 获取用户角色
func (r *userRepository) GetUserRoles(ctx context.Context, userID string) ([]*model.Role, error) {
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

// GetRoleUsers 获取角色用户
func (r *userRepository) GetRoleUsers(ctx context.Context, roleID string) ([]*model.User, error) {
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

// GetUserPermissions 获取用户权限
func (r *userRepository) GetUserPermissions(ctx context.Context, userID string) ([]*model.Permission, error) {
	var permissions []*model.Permission
	
	if err := r.db.WithContext(ctx).Table("permissions").
		Joins("JOIN role_permissions ON permissions.id = role_permissions.permission_id").
		Joins("JOIN user_roles ON role_permissions.role_id = user_roles.role_id").
		Where("user_roles.user_id = ?", userID).
		Distinct().Find(&permissions).Error; err != nil {
		logger.L.Error("Failed to get user permissions", zap.Error(err))
		return nil, fmt.Errorf("failed to get user permissions: %w", err)
	}
	
	return permissions, nil
}

// CheckUserPermission 检查用户权限
func (r *userRepository) CheckUserPermission(ctx context.Context, userID, permissionCode string) (bool, error) {
	var count int64
	
	if err := r.db.WithContext(ctx).Table("permissions").
		Joins("JOIN role_permissions ON permissions.id = role_permissions.permission_id").
		Joins("JOIN user_roles ON role_permissions.role_id = user_roles.role_id").
		Where("user_roles.user_id = ? AND permissions.code = ? AND permissions.status = ?", userID, permissionCode, "active").
		Count(&count).Error; err != nil {
		logger.L.Error("Failed to check user permission", zap.Error(err))
		return false, fmt.Errorf("failed to check user permission: %w", err)
	}
	
	return count > 0, nil
}

// GetStats 获取用户统计
func (r *userRepository) GetStats(ctx context.Context) (*model.UserStats, error) {
	stats := &model.UserStats{}
	
	// 总用户数
	if err := r.db.WithContext(ctx).Model(&model.User{}).Count(&stats.Total).Error; err != nil {
		return nil, fmt.Errorf("failed to count total users: %w", err)
	}
	
	// 活跃用户数
	if err := r.db.WithContext(ctx).Model(&model.User{}).Where("status = ?", model.UserStatusActive).Count(&stats.Active).Error; err != nil {
		return nil, fmt.Errorf("failed to count active users: %w", err)
	}
	
	// 非活跃用户数
	if err := r.db.WithContext(ctx).Model(&model.User{}).Where("status = ?", model.UserStatusInactive).Count(&stats.Inactive).Error; err != nil {
		return nil, fmt.Errorf("failed to count inactive users: %w", err)
	}
	
	// 锁定用户数
	if err := r.db.WithContext(ctx).Model(&model.User{}).Where("status = ?", model.UserStatusLocked).Count(&stats.Locked).Error; err != nil {
		return nil, fmt.Errorf("failed to count locked users: %w", err)
	}
	
	return stats, nil
}

// Count 统计用户数量
func (r *userRepository) Count(ctx context.Context, filters map[string]interface{}) (int64, error) {
	var count int64
	query := r.db.WithContext(ctx).Model(&model.User{})
	
	// 应用筛选条件
	for key, value := range filters {
		switch key {
		case "status":
			if v, ok := value.(string); ok && v != "" {
				query = query.Where("status = ?", v)
			}
		case "role":
			if v, ok := value.(string); ok && v != "" {
				query = query.Where("role = ?", v)
			}
		case "department":
			if v, ok := value.(string); ok && v != "" {
				query = query.Where("department = ?", v)
			}
		}
	}
	
	if err := query.Count(&count).Error; err != nil {
		logger.L.Error("Failed to count users", zap.Error(err))
		return 0, fmt.Errorf("failed to count users: %w", err)
	}
	
	return count, nil
}

// CheckUsernameExists 检查用户名是否存在
func (r *userRepository) CheckUsernameExists(ctx context.Context, username string, excludeID string) (bool, error) {
	var count int64
	query := r.db.WithContext(ctx).Model(&model.User{}).Where("username = ?", username)
	
	if excludeID != "" {
		query = query.Where("id != ?", excludeID)
	}
	
	if err := query.Count(&count).Error; err != nil {
		logger.L.Error("Failed to check username exists", zap.Error(err))
		return false, fmt.Errorf("failed to check username exists: %w", err)
	}
	
	return count > 0, nil
}

// CheckEmailExists 检查邮箱是否存在
func (r *userRepository) CheckEmailExists(ctx context.Context, email string, excludeID string) (bool, error) {
	var count int64
	query := r.db.WithContext(ctx).Model(&model.User{}).Where("email = ?", email)
	
	if excludeID != "" {
		query = query.Where("id != ?", excludeID)
	}
	
	if err := query.Count(&count).Error; err != nil {
		logger.L.Error("Failed to check email exists", zap.Error(err))
		return false, fmt.Errorf("failed to check email exists: %w", err)
	}
	
	return count > 0, nil
}
