package rbac

import (
	"context"
	"fmt"
	"strings"
	"sync"
)

// Permission 权限定义
type Permission string

const (
	// 告警相关权限
	PermissionAlertRead   Permission = "alert:read"
	PermissionAlertWrite  Permission = "alert:write"
	PermissionAlertDelete Permission = "alert:delete"
	PermissionAlertHandle Permission = "alert:handle"

	// 规则相关权限
	PermissionRuleRead   Permission = "rule:read"
	PermissionRuleWrite  Permission = "rule:write"
	PermissionRuleDelete Permission = "rule:delete"

	// 集群相关权限
	PermissionClusterRead   Permission = "cluster:read"
	PermissionClusterWrite  Permission = "cluster:write"
	PermissionClusterDelete Permission = "cluster:delete"

	// 用户管理权限
	PermissionUserRead   Permission = "user:read"
	PermissionUserWrite  Permission = "user:write"
	PermissionUserDelete Permission = "user:delete"

	// 系统管理权限
	PermissionSystemConfig Permission = "system:config"
	PermissionSystemAudit  Permission = "system:audit"
	PermissionSystemAdmin  Permission = "system:admin"
)

// Role 角色定义
type Role struct {
	Name        string       `json:"name"`
	Description string       `json:"description"`
	Permissions []Permission `json:"permissions"`
}

// User 用户定义
type User struct {
	ID       string   `json:"id"`
	Username string   `json:"username"`
	Roles    []string `json:"roles"`
}

// RBACManager RBAC管理器
type RBACManager struct {
	mu          sync.RWMutex
	roles       map[string]*Role
	userRoles   map[string][]string
	roleCache   map[string][]Permission // 角色权限缓存
}

// NewRBACManager 创建RBAC管理器
func NewRBACManager() *RBACManager {
	manager := &RBACManager{
		roles:     make(map[string]*Role),
		userRoles: make(map[string][]string),
		roleCache: make(map[string][]Permission),
	}

	// 初始化默认角色
	manager.initDefaultRoles()
	return manager
}

// initDefaultRoles 初始化默认角色
func (rm *RBACManager) initDefaultRoles() {
	// 管理员角色
	adminRole := &Role{
		Name:        "admin",
		Description: "系统管理员",
		Permissions: []Permission{
			PermissionAlertRead, PermissionAlertWrite, PermissionAlertDelete, PermissionAlertHandle,
			PermissionRuleRead, PermissionRuleWrite, PermissionRuleDelete,
			PermissionClusterRead, PermissionClusterWrite, PermissionClusterDelete,
			PermissionUserRead, PermissionUserWrite, PermissionUserDelete,
			PermissionSystemConfig, PermissionSystemAudit, PermissionSystemAdmin,
		},
	}

	// 运维角色
	operatorRole := &Role{
		Name:        "operator",
		Description: "运维人员",
		Permissions: []Permission{
			PermissionAlertRead, PermissionAlertWrite, PermissionAlertHandle,
			PermissionRuleRead, PermissionRuleWrite,
			PermissionClusterRead, PermissionClusterWrite,
		},
	}

	// 只读角色
	viewerRole := &Role{
		Name:        "viewer",
		Description: "只读用户",
		Permissions: []Permission{
			PermissionAlertRead,
			PermissionRuleRead,
			PermissionClusterRead,
		},
	}

	rm.roles["admin"] = adminRole
	rm.roles["operator"] = operatorRole
	rm.roles["viewer"] = viewerRole

	// 更新缓存
	rm.updateRoleCache("admin")
	rm.updateRoleCache("operator")
	rm.updateRoleCache("viewer")
}

// AddRole 添加角色
func (rm *RBACManager) AddRole(role *Role) error {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	if _, exists := rm.roles[role.Name]; exists {
		return fmt.Errorf("role %s already exists", role.Name)
	}

	rm.roles[role.Name] = role
	rm.updateRoleCache(role.Name)
	return nil
}

// GetRole 获取角色
func (rm *RBACManager) GetRole(roleName string) (*Role, error) {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	role, exists := rm.roles[roleName]
	if !exists {
		return nil, fmt.Errorf("role %s not found", roleName)
	}

	return role, nil
}

// AssignRole 为用户分配角色
func (rm *RBACManager) AssignRole(userID, roleName string) error {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	if _, exists := rm.roles[roleName]; !exists {
		return fmt.Errorf("role %s not found", roleName)
	}

	userRoles := rm.userRoles[userID]
	for _, role := range userRoles {
		if role == roleName {
			return nil // 角色已存在
		}
	}

	rm.userRoles[userID] = append(userRoles, roleName)
	return nil
}

// RevokeRole 撤销用户角色
func (rm *RBACManager) RevokeRole(userID, roleName string) error {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	userRoles := rm.userRoles[userID]
	for i, role := range userRoles {
		if role == roleName {
			rm.userRoles[userID] = append(userRoles[:i], userRoles[i+1:]...)
			return nil
		}
	}

	return fmt.Errorf("user %s does not have role %s", userID, roleName)
}

// HasPermission 检查用户是否有指定权限
func (rm *RBACManager) HasPermission(userID string, permission Permission) bool {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	userRoles := rm.userRoles[userID]
	for _, roleName := range userRoles {
		if permissions, exists := rm.roleCache[roleName]; exists {
			for _, perm := range permissions {
				if perm == permission {
					return true
				}
			}
		}
	}

	return false
}

// HasAnyPermission 检查用户是否有任意一个权限
func (rm *RBACManager) HasAnyPermission(userID string, permissions []Permission) bool {
	for _, permission := range permissions {
		if rm.HasPermission(userID, permission) {
			return true
		}
	}
	return false
}

// HasAllPermissions 检查用户是否有所有权限
func (rm *RBACManager) HasAllPermissions(userID string, permissions []Permission) bool {
	for _, permission := range permissions {
		if !rm.HasPermission(userID, permission) {
			return false
		}
	}
	return true
}

// GetUserRoles 获取用户角色
func (rm *RBACManager) GetUserRoles(userID string) []string {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	return rm.userRoles[userID]
}

// GetUserPermissions 获取用户所有权限
func (rm *RBACManager) GetUserPermissions(userID string) []Permission {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	permissionSet := make(map[Permission]bool)
	userRoles := rm.userRoles[userID]

	for _, roleName := range userRoles {
		if permissions, exists := rm.roleCache[roleName]; exists {
			for _, perm := range permissions {
				permissionSet[perm] = true
			}
		}
	}

	var permissions []Permission
	for perm := range permissionSet {
		permissions = append(permissions, perm)
	}

	return permissions
}

// CheckResourcePermission 检查资源权限（支持通配符）
func (rm *RBACManager) CheckResourcePermission(userID, resource, action string) bool {
	permission := Permission(fmt.Sprintf("%s:%s", resource, action))
	return rm.HasPermission(userID, permission)
}

// updateRoleCache 更新角色权限缓存
func (rm *RBACManager) updateRoleCache(roleName string) {
	if role, exists := rm.roles[roleName]; exists {
		rm.roleCache[roleName] = role.Permissions
	}
}

// ValidatePermission 验证权限格式
func ValidatePermission(permission string) bool {
	parts := strings.Split(permission, ":")
	return len(parts) == 2 && parts[0] != "" && parts[1] != ""
}

// PermissionContext 权限上下文
type PermissionContext struct {
	UserID     string
	Resource   string
	Action     string
	Attributes map[string]interface{}
}

// CheckPermissionWithContext 基于上下文检查权限
func (rm *RBACManager) CheckPermissionWithContext(ctx context.Context, permCtx *PermissionContext) bool {
	// 基础权限检查
	if !rm.CheckResourcePermission(permCtx.UserID, permCtx.Resource, permCtx.Action) {
		return false
	}

	// 可以在这里添加更复杂的上下文权限检查逻辑
	// 例如：基于资源属性、时间、IP地址等的权限检查

	return true
}