package service

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"

	"alert_agent/internal/model"
	"alert_agent/internal/pkg/logger"
	"alert_agent/internal/repository"

	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

// UserService 用户服务接口
type UserService interface {
	// 基础CRUD操作
	Create(ctx context.Context, user *model.User) error
	GetByID(ctx context.Context, id string) (*model.User, error)
	GetByUsername(ctx context.Context, username string) (*model.User, error)
	Update(ctx context.Context, user *model.User) error
	Delete(ctx context.Context, id string) error
	
	// 查询操作
	List(ctx context.Context, filters map[string]interface{}, page, pageSize int) ([]*model.User, int64, error)
	GetByRole(ctx context.Context, role string) ([]*model.User, error)
	GetByStatus(ctx context.Context, status string) ([]*model.User, error)
	GetByDepartment(ctx context.Context, department string) ([]*model.User, error)
	
	// 批量操作
	BatchCreate(ctx context.Context, users []*model.User) error
	BatchUpdate(ctx context.Context, ids []string, updates map[string]interface{}) error
	BatchDelete(ctx context.Context, ids []string) error
	
	// 角色管理
	AssignRoles(ctx context.Context, userID string, roleIDs []string) error
	RemoveRoles(ctx context.Context, userID string, roleIDs []string) error
	SyncRoles(ctx context.Context, userID string, roleIDs []string) error
	GetUserRoles(ctx context.Context, userID string) ([]*model.Role, error)
	
	// 权限管理
	GetUserPermissions(ctx context.Context, userID string) ([]*model.Permission, error)
	CheckUserPermission(ctx context.Context, userID, permissionCode string) (bool, error)
	
	// 密码管理
	ChangePassword(ctx context.Context, userID, oldPassword, newPassword string) error
	ResetPassword(ctx context.Context, userID string) (string, error)
	
	// 状态管理
	Activate(ctx context.Context, userID string) error
	Deactivate(ctx context.Context, userID string) error
	Lock(ctx context.Context, userID string) error
	Unlock(ctx context.Context, userID string) error
	
	// 统计操作
	GetStats(ctx context.Context) (*model.UserStats, error)
	
	// 验证操作
	CheckUsernameExists(ctx context.Context, username string, excludeID string) (bool, error)
	CheckEmailExists(ctx context.Context, email string, excludeID string) (bool, error)
	VerifyPassword(ctx context.Context, userID, password string) (bool, error)
}

// userService 用户服务实现
type userService struct {
	userRepo repository.UserRepository
	roleRepo repository.RoleRepository
}

// NewUserService 创建用户服务实例
func NewUserService(userRepo repository.UserRepository, roleRepo repository.RoleRepository) UserService {
	return &userService{
		userRepo: userRepo,
		roleRepo: roleRepo,
	}
}

// Create 创建用户
func (s *userService) Create(ctx context.Context, user *model.User) error {
	logger.L.Debug("Creating user", zap.String("username", user.Username))
	
	// 检查用户名是否存在
	exists, err := s.userRepo.CheckUsernameExists(ctx, user.Username, "")
	if err != nil {
		return fmt.Errorf("failed to check username: %w", err)
	}
	if exists {
		return errors.New("username already exists")
	}
	
	// 检查邮箱是否存在
	if user.Email != "" {
		exists, err = s.userRepo.CheckEmailExists(ctx, user.Email, "")
		if err != nil {
			return fmt.Errorf("failed to check email: %w", err)
		}
		if exists {
			return errors.New("email already exists")
		}
	}
	
	// 设置默认值
	if user.Status == "" {
		user.Status = model.UserStatusActive
	}
	
	// 加密密码
	if user.Password != "" {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
		if err != nil {
			return fmt.Errorf("failed to hash password: %w", err)
		}
		user.Password = string(hashedPassword)
	}
	
	// 创建用户
	if err := s.userRepo.Create(ctx, user); err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}
	
	// 分配默认角色
	if len(user.Roles) > 0 {
		roleIDs := make([]string, 0, len(user.Roles))
		for _, role := range user.Roles {
			roleIDs = append(roleIDs, role.ID)
		}
		
		if err := s.userRepo.AssignRoles(ctx, user.ID, roleIDs); err != nil {
			logger.L.Error("Failed to assign default roles", zap.Error(err))
			// 不返回错误，继续执行
		}
	}
	
	logger.L.Info("User created successfully", zap.String("id", user.ID))
	return nil
}

// GetByID 根据ID获取用户
func (s *userService) GetByID(ctx context.Context, id string) (*model.User, error) {
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	
	// 清除敏感信息
	user.Password = ""
	
	return user, nil
}

// GetByUsername 根据用户名获取用户
func (s *userService) GetByUsername(ctx context.Context, username string) (*model.User, error) {
	user, err := s.userRepo.GetByUsername(ctx, username)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	
	// 清除敏感信息
	user.Password = ""
	
	return user, nil
}

// Update 更新用户
func (s *userService) Update(ctx context.Context, user *model.User) error {
	logger.L.Debug("Updating user", zap.String("id", user.ID))
	
	// 检查用户是否存在
	existingUser, err := s.userRepo.GetByID(ctx, user.ID)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}
	
	// 检查用户名是否存在
	if user.Username != existingUser.Username {
		exists, err := s.userRepo.CheckUsernameExists(ctx, user.Username, user.ID)
		if err != nil {
			return fmt.Errorf("failed to check username: %w", err)
		}
		if exists {
			return errors.New("username already exists")
		}
	}
	
	// 检查邮箱是否存在
	if user.Email != existingUser.Email && user.Email != "" {
		exists, err := s.userRepo.CheckEmailExists(ctx, user.Email, user.ID)
		if err != nil {
			return fmt.Errorf("failed to check email: %w", err)
		}
		if exists {
			return errors.New("email already exists")
		}
	}
	
	// 保留原密码
	if user.Password == "" {
		user.Password = existingUser.Password
	} else {
		// 加密新密码
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
		if err != nil {
			return fmt.Errorf("failed to hash password: %w", err)
		}
		user.Password = string(hashedPassword)
	}
	
	// 更新用户
	if err := s.userRepo.Update(ctx, user); err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}
	
	// 同步角色
	if len(user.Roles) > 0 {
		roleIDs := make([]string, 0, len(user.Roles))
		for _, role := range user.Roles {
			roleIDs = append(roleIDs, role.ID)
		}
		
		if err := s.userRepo.SyncRoles(ctx, user.ID, roleIDs); err != nil {
			logger.L.Error("Failed to sync roles", zap.Error(err))
			// 不返回错误，继续执行
		}
	}
	
	logger.L.Info("User updated successfully", zap.String("id", user.ID))
	return nil
}

// Delete 删除用户
func (s *userService) Delete(ctx context.Context, id string) error {
	logger.L.Debug("Deleting user", zap.String("id", id))
	
	// 检查用户是否存在
	_, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}
	
	// 删除用户
	if err := s.userRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}
	
	logger.L.Info("User deleted successfully", zap.String("id", id))
	return nil
}

// List 获取用户列表
func (s *userService) List(ctx context.Context, filters map[string]interface{}, page, pageSize int) ([]*model.User, int64, error) {
	// 计算分页参数
	offset := (page - 1) * pageSize
	limit := pageSize
	
	// 获取用户列表
	users, total, err := s.userRepo.List(ctx, filters, offset, limit)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list users: %w", err)
	}
	
	// 清除敏感信息
	for _, user := range users {
		user.Password = ""
	}
	
	return users, total, nil
}

// GetByRole 根据角色获取用户
func (s *userService) GetByRole(ctx context.Context, role string) ([]*model.User, error) {
	users, err := s.userRepo.GetByRole(ctx, role)
	if err != nil {
		return nil, fmt.Errorf("failed to get users by role: %w", err)
	}
	
	// 清除敏感信息
	for _, user := range users {
		user.Password = ""
	}
	
	return users, nil
}

// GetByStatus 根据状态获取用户
func (s *userService) GetByStatus(ctx context.Context, status string) ([]*model.User, error) {
	users, err := s.userRepo.GetByStatus(ctx, status)
	if err != nil {
		return nil, fmt.Errorf("failed to get users by status: %w", err)
	}
	
	// 清除敏感信息
	for _, user := range users {
		user.Password = ""
	}
	
	return users, nil
}

// GetByDepartment 根据部门获取用户
func (s *userService) GetByDepartment(ctx context.Context, department string) ([]*model.User, error) {
	users, err := s.userRepo.GetByDepartment(ctx, department)
	if err != nil {
		return nil, fmt.Errorf("failed to get users by department: %w", err)
	}
	
	// 清除敏感信息
	for _, user := range users {
		user.Password = ""
	}
	
	return users, nil
}

// BatchCreate 批量创建用户
func (s *userService) BatchCreate(ctx context.Context, users []*model.User) error {
	logger.L.Debug("Batch creating users", zap.Int("count", len(users)))
	
	// 处理每个用户
	for _, user := range users {
		// 设置默认值
		if user.Status == "" {
			user.Status = model.UserStatusActive
		}
		
		// 加密密码
		if user.Password != "" {
			hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
			if err != nil {
				return fmt.Errorf("failed to hash password: %w", err)
			}
			user.Password = string(hashedPassword)
		}
	}
	
	// 批量创建用户
	if err := s.userRepo.BatchCreate(ctx, users); err != nil {
		return fmt.Errorf("failed to batch create users: %w", err)
	}
	
	logger.L.Info("Users batch created successfully", zap.Int("count", len(users)))
	return nil
}

// BatchUpdate 批量更新用户
func (s *userService) BatchUpdate(ctx context.Context, ids []string, updates map[string]interface{}) error {
	logger.L.Debug("Batch updating users", zap.Strings("ids", ids))
	
	// 检查是否包含密码更新
	if password, ok := updates["password"].(string); ok && password != "" {
		// 加密密码
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			return fmt.Errorf("failed to hash password: %w", err)
		}
		updates["password"] = string(hashedPassword)
	}
	
	// 批量更新用户
	if err := s.userRepo.BatchUpdate(ctx, ids, updates); err != nil {
		return fmt.Errorf("failed to batch update users: %w", err)
	}
	
	logger.L.Info("Users batch updated successfully", zap.Strings("ids", ids))
	return nil
}

// BatchDelete 批量删除用户
func (s *userService) BatchDelete(ctx context.Context, ids []string) error {
	logger.L.Debug("Batch deleting users", zap.Strings("ids", ids))
	
	// 批量删除用户
	if err := s.userRepo.BatchDelete(ctx, ids); err != nil {
		return fmt.Errorf("failed to batch delete users: %w", err)
	}
	
	logger.L.Info("Users batch deleted successfully", zap.Strings("ids", ids))
	return nil
}

// AssignRoles 为用户分配角色
func (s *userService) AssignRoles(ctx context.Context, userID string, roleIDs []string) error {
	logger.L.Debug("Assigning roles to user", zap.String("user_id", userID), zap.Strings("role_ids", roleIDs))
	
	// 检查用户是否存在
	_, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}
	
	// 检查角色是否存在
	for _, roleID := range roleIDs {
		_, err := s.roleRepo.GetByID(ctx, roleID)
		if err != nil {
			return fmt.Errorf("failed to get role: %w", err)
		}
	}
	
	// 分配角色
	if err := s.userRepo.AssignRoles(ctx, userID, roleIDs); err != nil {
		return fmt.Errorf("failed to assign roles: %w", err)
	}
	
	logger.L.Info("Roles assigned successfully", zap.String("user_id", userID))
	return nil
}

// RemoveRoles 移除用户角色
func (s *userService) RemoveRoles(ctx context.Context, userID string, roleIDs []string) error {
	logger.L.Debug("Removing roles from user", zap.String("user_id", userID), zap.Strings("role_ids", roleIDs))
	
	// 检查用户是否存在
	_, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}
	
	// 移除角色
	if err := s.userRepo.RemoveRoles(ctx, userID, roleIDs); err != nil {
		return fmt.Errorf("failed to remove roles: %w", err)
	}
	
	logger.L.Info("Roles removed successfully", zap.String("user_id", userID))
	return nil
}

// SyncRoles 同步用户角色
func (s *userService) SyncRoles(ctx context.Context, userID string, roleIDs []string) error {
	logger.L.Debug("Syncing roles for user", zap.String("user_id", userID), zap.Strings("role_ids", roleIDs))
	
	// 检查用户是否存在
	_, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}
	
	// 检查角色是否存在
	for _, roleID := range roleIDs {
		_, err := s.roleRepo.GetByID(ctx, roleID)
		if err != nil {
			return fmt.Errorf("failed to get role: %w", err)
		}
	}
	
	// 同步角色
	if err := s.userRepo.SyncRoles(ctx, userID, roleIDs); err != nil {
		return fmt.Errorf("failed to sync roles: %w", err)
	}
	
	logger.L.Info("Roles synced successfully", zap.String("user_id", userID))
	return nil
}

// GetUserRoles 获取用户角色
func (s *userService) GetUserRoles(ctx context.Context, userID string) ([]*model.Role, error) {
	// 检查用户是否存在
	_, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	
	// 获取用户角色
	roles, err := s.userRepo.GetUserRoles(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user roles: %w", err)
	}
	
	return roles, nil
}

// GetUserPermissions 获取用户权限
func (s *userService) GetUserPermissions(ctx context.Context, userID string) ([]*model.Permission, error) {
	// 检查用户是否存在
	_, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	
	// 获取用户权限
	permissions, err := s.userRepo.GetUserPermissions(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user permissions: %w", err)
	}
	
	return permissions, nil
}

// CheckUserPermission 检查用户权限
func (s *userService) CheckUserPermission(ctx context.Context, userID, permissionCode string) (bool, error) {
	// 检查用户是否存在
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return false, fmt.Errorf("failed to get user: %w", err)
	}
	
	// 系统管理员拥有所有权限
	for _, role := range user.Roles {
		if role.Code == "admin" {
			return true, nil
		}
	}
	
	// 检查用户权限
	hasPermission, err := s.userRepo.CheckUserPermission(ctx, userID, permissionCode)
	if err != nil {
		return false, fmt.Errorf("failed to check user permission: %w", err)
	}
	
	return hasPermission, nil
}

// ChangePassword 修改密码
func (s *userService) ChangePassword(ctx context.Context, userID, oldPassword, newPassword string) error {
	logger.L.Debug("Changing password for user", zap.String("user_id", userID))
	
	// 获取用户
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}
	
	// 验证旧密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(oldPassword)); err != nil {
		return errors.New("invalid old password")
	}
	
	// 加密新密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}
	
	// 更新密码
	user.Password = string(hashedPassword)
	if err := s.userRepo.Update(ctx, user); err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}
	
	logger.L.Info("Password changed successfully", zap.String("user_id", userID))
	return nil
}

// ResetPassword 重置密码
func (s *userService) ResetPassword(ctx context.Context, userID string) (string, error) {
	logger.L.Debug("Resetting password for user", zap.String("user_id", userID))
	
	// 获取用户
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return "", fmt.Errorf("failed to get user: %w", err)
	}
	
	// 生成随机密码
	newPassword := generateRandomPassword(12)
	
	// 加密新密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}
	
	// 更新密码
	user.Password = string(hashedPassword)
	if err := s.userRepo.Update(ctx, user); err != nil {
		return "", fmt.Errorf("failed to update user: %w", err)
	}
	
	logger.L.Info("Password reset successfully", zap.String("user_id", userID))
	return newPassword, nil
}

// Activate 激活用户
func (s *userService) Activate(ctx context.Context, userID string) error {
	logger.L.Debug("Activating user", zap.String("user_id", userID))
	
	// 获取用户
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}
	
	// 更新状态
	user.Status = model.UserStatusActive
	if err := s.userRepo.Update(ctx, user); err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}
	
	logger.L.Info("User activated successfully", zap.String("user_id", userID))
	return nil
}

// Deactivate 停用用户
func (s *userService) Deactivate(ctx context.Context, userID string) error {
	logger.L.Debug("Deactivating user", zap.String("user_id", userID))
	
	// 获取用户
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}
	
	// 更新状态
	user.Status = model.UserStatusInactive
	if err := s.userRepo.Update(ctx, user); err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}
	
	logger.L.Info("User deactivated successfully", zap.String("user_id", userID))
	return nil
}

// Lock 锁定用户
func (s *userService) Lock(ctx context.Context, userID string) error {
	logger.L.Debug("Locking user", zap.String("user_id", userID))
	
	// 获取用户
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}
	
	// 更新状态
	user.Status = model.UserStatusLocked
	if err := s.userRepo.Update(ctx, user); err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}
	
	logger.L.Info("User locked successfully", zap.String("user_id", userID))
	return nil
}

// Unlock 解锁用户
func (s *userService) Unlock(ctx context.Context, userID string) error {
	logger.L.Debug("Unlocking user", zap.String("user_id", userID))
	
	// 获取用户
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}
	
	// 更新状态
	user.Status = model.UserStatusActive
	if err := s.userRepo.Update(ctx, user); err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}
	
	logger.L.Info("User unlocked successfully", zap.String("user_id", userID))
	return nil
}

// GetStats 获取用户统计
func (s *userService) GetStats(ctx context.Context) (*model.UserStats, error) {
	stats, err := s.userRepo.GetStats(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get user stats: %w", err)
	}
	
	return stats, nil
}

// CheckUsernameExists 检查用户名是否存在
func (s *userService) CheckUsernameExists(ctx context.Context, username string, excludeID string) (bool, error) {
	exists, err := s.userRepo.CheckUsernameExists(ctx, username, excludeID)
	if err != nil {
		return false, fmt.Errorf("failed to check username exists: %w", err)
	}
	
	return exists, nil
}

// CheckEmailExists 检查邮箱是否存在
func (s *userService) CheckEmailExists(ctx context.Context, email string, excludeID string) (bool, error) {
	exists, err := s.userRepo.CheckEmailExists(ctx, email, excludeID)
	if err != nil {
		return false, fmt.Errorf("failed to check email exists: %w", err)
	}
	
	return exists, nil
}

// VerifyPassword 验证密码
func (s *userService) VerifyPassword(ctx context.Context, userID, password string) (bool, error) {
	// 获取用户
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return false, fmt.Errorf("failed to get user: %w", err)
	}
	
	// 验证密码
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	return err == nil, nil
}

// generateRandomPassword 生成随机密码
func generateRandomPassword(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*"
	password := make([]byte, length)
	
	for i := range password {
		n, _ := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		password[i] = charset[n.Int64()]
	}
	
	return string(password)
}