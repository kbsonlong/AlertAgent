package user

import (
	"context"
	"errors"
	"fmt"
	"time"

	"alert_agent/internal/security/auth"
	"alert_agent/internal/security/audit"
	"alert_agent/internal/security/crypto"
	"alert_agent/internal/security/rbac"
	"alert_agent/internal/security/validator"
)

// User 用户实体
type User struct {
	ID          string    `json:"id" gorm:"primaryKey"`
	Username    string    `json:"username" gorm:"unique;not null"`
	Email       string    `json:"email" gorm:"unique;not null"`
	PasswordHash string   `json:"-" gorm:"not null"`
	Salt        string    `json:"-" gorm:"not null"`
	Roles       []string  `json:"roles" gorm:"type:text"`
	IsActive    bool      `json:"is_active" gorm:"default:true"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	LastLoginAt *time.Time `json:"last_login_at"`
	LoginAttempts int     `json:"-" gorm:"default:0"`
	LockedUntil   *time.Time `json:"-"`
}

// UserRepository 用户仓储接口
type UserRepository interface {
	Create(ctx context.Context, user *User) error
	GetByID(ctx context.Context, id string) (*User, error)
	GetByUsername(ctx context.Context, username string) (*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
	Update(ctx context.Context, user *User) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, offset, limit int) ([]*User, int64, error)
	IncrementLoginAttempts(ctx context.Context, userID string) error
	ResetLoginAttempts(ctx context.Context, userID string) error
	LockUser(ctx context.Context, userID string, until time.Time) error
}

// UserService 用户服务
type UserService struct {
	repo         UserRepository
	jwtManager   *auth.JWTManager
	rbacManager  *rbac.RBACManager
	cryptoManager *crypto.EncryptionManager
	auditLogger  *audit.AuditLogger
	validator    *validator.Validator
	maxLoginAttempts int
	lockoutDuration  time.Duration
}

// NewUserService 创建用户服务
func NewUserService(
	repo UserRepository,
	jwtManager *auth.JWTManager,
	rbacManager *rbac.RBACManager,
	cryptoManager *crypto.EncryptionManager,
	auditLogger *audit.AuditLogger,
	maxLoginAttempts int,
	lockoutDuration time.Duration,
) *UserService {
	return &UserService{
		repo:             repo,
		jwtManager:       jwtManager,
		rbacManager:      rbacManager,
		cryptoManager:    cryptoManager,
		auditLogger:      auditLogger,
		validator:        validator.NewValidator(),
		maxLoginAttempts: maxLoginAttempts,
		lockoutDuration:  lockoutDuration,
	}
}

// CreateUserRequest 创建用户请求
type CreateUserRequest struct {
	Username string   `json:"username"`
	Email    string   `json:"email"`
	Password string   `json:"password"`
	Roles    []string `json:"roles"`
}

// Validate 验证创建用户请求
func (r *CreateUserRequest) Validate() error {
	v := validator.NewValidator()
	v.Required("username", r.Username)
	v.MinLength("username", r.Username, 3)
	v.MaxLength("username", r.Username, 50)
	v.Required("email", r.Email)
	v.Email("email", r.Email)
	v.Required("password", r.Password)
	v.Password("password", r.Password)
	
	if v.HasErrors() {
		return v.GetErrors()
	}
	return nil
}

// LoginRequest 登录请求
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// Validate 验证登录请求
func (r *LoginRequest) Validate() error {
	v := validator.NewValidator()
	v.Required("username", r.Username)
	v.Required("password", r.Password)
	
	if v.HasErrors() {
		return v.GetErrors()
	}
	return nil
}

// LoginResponse 登录响应
type LoginResponse struct {
	User         *User  `json:"user"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
}

// UpdateUserRequest 更新用户请求
type UpdateUserRequest struct {
	Email    *string  `json:"email,omitempty"`
	Roles    []string `json:"roles,omitempty"`
	IsActive *bool    `json:"is_active,omitempty"`
}

// Validate 验证更新用户请求
func (r *UpdateUserRequest) Validate() error {
	v := validator.NewValidator()
	if r.Email != nil {
		v.Email("email", *r.Email)
	}
	
	if v.HasErrors() {
		return v.GetErrors()
	}
	return nil
}

// ChangePasswordRequest 修改密码请求
type ChangePasswordRequest struct {
	OldPassword string `json:"old_password"`
	NewPassword string `json:"new_password"`
}

// Validate 验证修改密码请求
func (r *ChangePasswordRequest) Validate() error {
	v := validator.NewValidator()
	v.Required("old_password", r.OldPassword)
	v.Required("new_password", r.NewPassword)
	v.Password("new_password", r.NewPassword)
	
	if v.HasErrors() {
		return v.GetErrors()
	}
	return nil
}

// CreateUser 创建用户
func (s *UserService) CreateUser(ctx context.Context, req *CreateUserRequest, operatorID string) (*User, error) {
	// 验证输入
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// 检查用户名是否已存在
	existingUser, _ := s.repo.GetByUsername(ctx, req.Username)
	if existingUser != nil {
		return nil, errors.New("username already exists")
	}

	// 检查邮箱是否已存在
	existingUser, _ = s.repo.GetByEmail(ctx, req.Email)
	if existingUser != nil {
		return nil, errors.New("email already exists")
	}

	// 生成盐值
	salt, err := crypto.GenerateSalt(16)
	if err != nil {
		return nil, fmt.Errorf("failed to generate salt: %w", err)
	}

	// 加密密码
	passwordHash, err := crypto.HashPassword(req.Password, salt)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// 创建用户
	user := &User{
		ID:           generateUserID(),
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: passwordHash,
		Salt:         string(salt),
		Roles:        req.Roles,
		IsActive:     true,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	// 保存用户
	if err := s.repo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// 分配角色
	for _, role := range req.Roles {
		s.rbacManager.AssignRole(user.ID, role)
	}

	// 记录审计日志
	s.auditLogger.LogDataOperation(operatorID, "", "", "users", user.ID, "USER_CREATE", true, map[string]interface{}{
		"user_id": user.ID,
		"username": user.Username,
		"roles": req.Roles,
	})

	return user, nil
}

// Login 用户登录
func (s *UserService) Login(ctx context.Context, req *LoginRequest, clientIP, userAgent string) (*LoginResponse, error) {
	// 验证输入
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// 获取用户
	user, err := s.repo.GetByUsername(ctx, req.Username)
	if err != nil {
		// 记录登录失败审计日志
		s.auditLogger.LogLogin("", req.Username, clientIP, userAgent, false, "User not found")
		return nil, errors.New("invalid credentials")
	}

	// 检查用户是否被锁定
	if user.LockedUntil != nil && time.Now().Before(*user.LockedUntil) {
		s.auditLogger.LogLogin(user.ID, user.Username, clientIP, userAgent, false, "Account locked")
		return nil, errors.New("account is locked")
	}

	// 检查用户是否激活
	if !user.IsActive {
		s.auditLogger.LogLogin(user.ID, user.Username, clientIP, userAgent, false, "Account inactive")
		return nil, errors.New("account is inactive")
	}

	// 验证密码
	valid, err := crypto.VerifyPassword(req.Password, user.PasswordHash, []byte(user.Salt))
	if err != nil || !valid {
		// 增加登录尝试次数
		s.repo.IncrementLoginAttempts(ctx, user.ID)
		
		// 检查是否需要锁定账户
		if user.LoginAttempts+1 >= s.maxLoginAttempts {
			lockUntil := time.Now().Add(s.lockoutDuration)
			s.repo.LockUser(ctx, user.ID, lockUntil)
			s.auditLogger.LogSecurityEvent(user.ID, user.Username, clientIP, "ACCOUNT_LOCKED", "Too many failed login attempts", map[string]interface{}{
				"attempts": user.LoginAttempts + 1,
				"locked_until": lockUntil,
			})
		}

		s.auditLogger.LogLogin(user.ID, user.Username, clientIP, userAgent, false, "Invalid password")
		return nil, errors.New("invalid credentials")
	}

	// 重置登录尝试次数
	s.repo.ResetLoginAttempts(ctx, user.ID)

	// 更新最后登录时间
	now := time.Now()
	user.LastLoginAt = &now
	user.UpdatedAt = now
	s.repo.Update(ctx, user)

	// 获取用户权限
	permissions := s.rbacManager.GetUserPermissions(user.ID)
	permissionStrings := make([]string, len(permissions))
	for i, perm := range permissions {
		permissionStrings[i] = string(perm)
	}

	// 生成JWT token
	accessToken, err := s.jwtManager.GenerateToken(user.ID, user.Username, user.Roles, permissionStrings)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	// 使用RefreshToken方法生成刷新token
	refreshToken, err := s.jwtManager.RefreshToken(accessToken)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	// 记录成功登录审计日志
	s.auditLogger.LogLogin(user.ID, user.Username, clientIP, userAgent, true, "Login successful")

	return &LoginResponse{
		User:         user,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int64(time.Hour * 24 / time.Second), // 24小时
	}, nil
}

// GetUser 获取用户信息
func (s *UserService) GetUser(ctx context.Context, userID string) (*User, error) {
	return s.repo.GetByID(ctx, userID)
}

// UpdateUser 更新用户信息
func (s *UserService) UpdateUser(ctx context.Context, userID string, req *UpdateUserRequest, operatorID string) (*User, error) {
	// 验证输入
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// 获取用户
	user, err := s.repo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	// 更新字段
	if req.Email != nil {
		// 检查邮箱是否已被其他用户使用
		existingUser, _ := s.repo.GetByEmail(ctx, *req.Email)
		if existingUser != nil && existingUser.ID != userID {
			return nil, errors.New("email already exists")
		}
		user.Email = *req.Email
	}

	if req.IsActive != nil {
		user.IsActive = *req.IsActive
	}

	if req.Roles != nil {
		// 移除旧角色
		for _, role := range user.Roles {
			s.rbacManager.RevokeRole(userID, role)
		}
		// 分配新角色
		for _, role := range req.Roles {
			s.rbacManager.AssignRole(userID, role)
		}
		user.Roles = req.Roles
	}

	user.UpdatedAt = time.Now()

	// 保存更新
	if err := s.repo.Update(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	// 记录审计日志
	s.auditLogger.LogDataOperation(operatorID, "", "", "users", userID, "USER_UPDATE", true, map[string]interface{}{
		"updated_fields": req,
	})

	return user, nil
}

// ChangePassword 修改密码
func (s *UserService) ChangePassword(ctx context.Context, userID string, req *ChangePasswordRequest, clientIP string) error {
	// 验证输入
	if err := req.Validate(); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	// 获取用户
	user, err := s.repo.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	// 验证旧密码
	valid, err := crypto.VerifyPassword(req.OldPassword, user.PasswordHash, []byte(user.Salt))
	if err != nil || !valid {
		s.auditLogger.LogSecurityEvent(userID, user.Username, clientIP, "PASSWORD_CHANGE_FAILED", "Invalid old password", nil)
		return errors.New("invalid old password")
	}

	// 生成新盐值
	newSalt, err := crypto.GenerateSalt(16)
	if err != nil {
		return fmt.Errorf("failed to generate salt: %w", err)
	}

	// 加密新密码
	newPasswordHash, err := crypto.HashPassword(req.NewPassword, newSalt)
	if err != nil {
		return fmt.Errorf("failed to hash new password: %w", err)
	}

	// 更新密码
	user.PasswordHash = newPasswordHash
	user.Salt = string(newSalt)
	user.UpdatedAt = time.Now()

	if err := s.repo.Update(ctx, user); err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	// 记录审计日志
	s.auditLogger.LogSecurityEvent(userID, user.Username, clientIP, "PASSWORD_CHANGED", "Password changed successfully", nil)

	return nil
}

// DeleteUser 删除用户
func (s *UserService) DeleteUser(ctx context.Context, userID string, operatorID string) error {
	// 获取用户信息用于审计
	user, err := s.repo.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	// 移除所有角色
	for _, role := range user.Roles {
		s.rbacManager.RevokeRole(userID, role)
	}

	// 删除用户
	if err := s.repo.Delete(ctx, userID); err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	// 记录审计日志
	s.auditLogger.LogDataOperation(operatorID, "", "", "users", userID, "USER_DELETE", true, map[string]interface{}{
		"username": user.Username,
	})

	return nil
}

// ListUsers 获取用户列表
func (s *UserService) ListUsers(ctx context.Context, offset, limit int) ([]*User, int64, error) {
	return s.repo.List(ctx, offset, limit)
}

// 辅助函数
func generateUserID() string {
	return fmt.Sprintf("user_%d", time.Now().UnixNano())
}