package di

import (
	"fmt"

	"alert_agent/internal/middleware"
	"alert_agent/internal/security/audit"
	"alert_agent/internal/security/auth"
	"alert_agent/internal/security/config"
	"alert_agent/internal/security/crypto"
	"alert_agent/internal/security/handler"
	"alert_agent/internal/security/rbac"
	"alert_agent/internal/security/user"
	"alert_agent/internal/security/validator"
)

// Container 依赖注入容器
type Container struct {
	// 配置
	SecurityConfig *config.SecurityConfig

	// 核心组件
	JWTManager       *auth.JWTManager
	EncryptionManager *crypto.EncryptionManager
	Validator        *validator.Validator
	AuditLogger      *audit.AuditLogger
	RBACManager      *rbac.RBACManager

	// 服务层
	UserService *user.UserService

	// 处理器
	AuthHandler *handler.AuthHandler

	// 中间件配置
	MiddlewareConfig *middleware.SecurityConfig
}

// NewContainer 创建新的依赖注入容器
func NewContainer() (*Container, error) {
	container := &Container{}

	// 初始化安全配置
	if err := container.initSecurityConfig(); err != nil {
		return nil, fmt.Errorf("failed to init security config: %w", err)
	}

	// 初始化核心组件
	if err := container.initCoreComponents(); err != nil {
		return nil, fmt.Errorf("failed to init core components: %w", err)
	}

	// 初始化服务层
	if err := container.initServices(); err != nil {
		return nil, fmt.Errorf("failed to init services: %w", err)
	}

	// 初始化处理器
	if err := container.initHandlers(); err != nil {
		return nil, fmt.Errorf("failed to init handlers: %w", err)
	}

	// 初始化中间件配置
	if err := container.initMiddlewareConfig(); err != nil {
		return nil, fmt.Errorf("failed to init middleware config: %w", err)
	}

	return container, nil
}

// initSecurityConfig 初始化安全配置
func (c *Container) initSecurityConfig() error {
	c.SecurityConfig = config.DefaultSecurityConfig()

	// 可以从环境变量或配置文件中覆盖默认配置
	// 这里使用默认配置

	return c.SecurityConfig.Validate()
}

// initCoreComponents 初始化核心组件
func (c *Container) initCoreComponents() error {
	// 初始化JWT管理器
	c.JWTManager = auth.NewJWTManager(
		c.SecurityConfig.JWT.Secret,
		c.SecurityConfig.JWT.Expiration,
	)

	// 初始化加密管理器
	var err error
	c.EncryptionManager, err = crypto.NewEncryptionManager(c.SecurityConfig.Encryption.Key, []byte(c.SecurityConfig.Encryption.Salt))
	if err != nil {
		return fmt.Errorf("failed to create encryption manager: %w", err)
	}

	// 初始化验证器
	c.Validator = validator.NewValidator()

	// 初始化审计日志器
	c.AuditLogger, err = audit.NewAuditLogger(c.SecurityConfig.Audit.LogFile, c.SecurityConfig.Audit.Enabled)
	if err != nil {
		return fmt.Errorf("failed to create audit logger: %w", err)
	}

	// 初始化RBAC管理器
	c.RBACManager = rbac.NewRBACManager()

	// 设置默认角色和权限
	if err := c.setupDefaultRBAC(); err != nil {
		return fmt.Errorf("failed to setup default RBAC: %w", err)
	}

	return nil
}

// initServices 初始化服务层
func (c *Container) initServices() error {
	// 这里需要实际的用户仓库实现
	// 暂时使用nil，在实际项目中需要注入真实的仓库实现
	var userRepo user.UserRepository = nil

	c.UserService = user.NewUserService(
		userRepo,
		c.JWTManager,
		c.RBACManager,
		c.EncryptionManager,
		c.AuditLogger,
		c.SecurityConfig.Auth.MaxLoginAttempts,
		c.SecurityConfig.Auth.LockoutDuration,
	)

	return nil
}

// initHandlers 初始化处理器
func (c *Container) initHandlers() error {
	c.AuthHandler = handler.NewAuthHandler(c.UserService)
	return nil
}

// initMiddlewareConfig 初始化中间件配置
func (c *Container) initMiddlewareConfig() error {
	c.MiddlewareConfig = &middleware.SecurityConfig{
		JWTManager:  c.JWTManager,
		RBACManager: c.RBACManager,
		AuditLogger: c.AuditLogger,
		SkipPaths: []string{
			"/health",
			"/api/v1/auth/register",
			"/api/v1/auth/login",
		},
	}

	return nil
}

// setupDefaultRBAC 设置默认的RBAC角色和权限
func (c *Container) setupDefaultRBAC() error {
	// RBAC管理器已经在初始化时设置了默认角色
	// 这里可以添加额外的自定义角色
	
	// 定义额外的角色
	userRole := &rbac.Role{
		Name:        "user",
		Description: "Regular user",
		Permissions: []rbac.Permission{rbac.PermissionUserRead},
	}

	moderatorRole := &rbac.Role{
		Name:        "moderator",
		Description: "Content moderator",
		Permissions: []rbac.Permission{
			rbac.PermissionUserRead,
			rbac.PermissionUserWrite,
		},
	}

	// 尝试添加角色（如果不存在的话）
	if _, err := c.RBACManager.GetRole("user"); err != nil {
		if err := c.RBACManager.AddRole(userRole); err != nil {
			return fmt.Errorf("failed to add user role: %w", err)
		}
	}

	if _, err := c.RBACManager.GetRole("moderator"); err != nil {
		if err := c.RBACManager.AddRole(moderatorRole); err != nil {
			return fmt.Errorf("failed to add moderator role: %w", err)
		}
	}

	return nil
}

// GetSecurityConfig 获取安全配置
func (c *Container) GetSecurityConfig() *config.SecurityConfig {
	return c.SecurityConfig
}

// GetJWTManager 获取JWT管理器
func (c *Container) GetJWTManager() *auth.JWTManager {
	return c.JWTManager
}

// GetEncryptionManager 获取加密管理器
func (c *Container) GetEncryptionManager() *crypto.EncryptionManager {
	return c.EncryptionManager
}

// GetValidator 获取验证器
func (c *Container) GetValidator() *validator.Validator {
	return c.Validator
}

// GetAuditLogger 获取审计日志器
func (c *Container) GetAuditLogger() *audit.AuditLogger {
	return c.AuditLogger
}

// GetRBACManager 获取RBAC管理器
func (c *Container) GetRBACManager() *rbac.RBACManager {
	return c.RBACManager
}

// GetUserService 获取用户服务
func (c *Container) GetUserService() *user.UserService {
	return c.UserService
}

// GetAuthHandler 获取认证处理器
func (c *Container) GetAuthHandler() *handler.AuthHandler {
	return c.AuthHandler
}

// GetMiddlewareConfig 获取中间件配置
func (c *Container) GetMiddlewareConfig() *middleware.SecurityConfig {
	return c.MiddlewareConfig
}

// Cleanup 清理资源
func (c *Container) Cleanup() error {
	// 这里可以添加资源清理逻辑
	// 比如关闭数据库连接、停止后台任务等
	return nil
}