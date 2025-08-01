package v1

import (
	"github.com/gin-gonic/gin"
)

// UserRequest 用户请求结构
type UserRequest struct {
	Username   string `json:"username" binding:"required" example:"admin"`
	Email      string `json:"email" binding:"required,email" example:"admin@example.com"`
	FullName   string `json:"full_name" binding:"required" example:"管理员"`
	Password   string `json:"password,omitempty" example:"password123"`
	Phone      string `json:"phone,omitempty" example:"+86 138 0013 8000"`
	Department string `json:"department,omitempty" example:"技术部"`
	Position   string `json:"position,omitempty" example:"系统管理员"`
	Role       string `json:"role" binding:"required" example:"admin"`
	Status     string `json:"status" binding:"required" example:"active"`
}

// UserResponse 用户响应结构
type UserResponse struct {
	ID           uint     `json:"id" example:"1"`
	Username     string   `json:"username" example:"admin"`
	Email        string   `json:"email" example:"admin@example.com"`
	FullName     string   `json:"full_name" example:"管理员"`
	Phone        string   `json:"phone,omitempty" example:"+86 138 0013 8000"`
	Department   string   `json:"department,omitempty" example:"技术部"`
	Position     string   `json:"position,omitempty" example:"系统管理员"`
	Role         string   `json:"role" example:"admin"`
	Status       string   `json:"status" example:"active"`
	LastLoginAt  *string  `json:"last_login_at,omitempty" example:"2024-01-01T12:00:00Z"`
	LastLoginIP  string   `json:"last_login_ip,omitempty" example:"192.168.1.100"`
	LoginCount   int      `json:"login_count" example:"10"`
	CreatedAt    string   `json:"created_at" example:"2024-01-01T12:00:00Z"`
	UpdatedAt    string   `json:"updated_at" example:"2024-01-01T12:00:00Z"`
	CreatedBy    string   `json:"created_by" example:"admin"`
	UpdatedBy    string   `json:"updated_by" example:"admin"`
	Permissions  []string `json:"permissions,omitempty"`
}

// UserListResponse 用户列表响应结构
type UserListResponse struct {
	Users []UserResponse `json:"users"`
	Total int            `json:"total" example:"100"`
	Page  int            `json:"page" example:"1"`
	Size  int            `json:"size" example:"10"`
}

// UserStatsResponse 用户统计响应结构
type UserStatsResponse struct {
	Total    int64 `json:"total" example:"100"`
	Active   int64 `json:"active" example:"85"`
	Inactive int64 `json:"inactive" example:"10"`
	Locked   int64 `json:"locked" example:"5"`
}

// 用户控制器实例
var userController *UserController

// InitUserController 初始化用户控制器
func InitUserController(uc *UserController) {
	userController = uc
}

// ListUsers 获取用户列表
// @Summary 获取用户列表
// @Description 获取系统中所有用户的列表，支持分页和筛选
// @Tags 用户管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "页码" default(1)
// @Param size query int false "每页数量" default(10)
// @Param username query string false "用户名筛选"
// @Param email query string false "邮箱筛选"
// @Param role query string false "角色筛选"
// @Param status query string false "状态筛选"
// @Success 200 {object} response.Response{data=UserListDetailResponse}
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/users [get]
func ListUsers(c *gin.Context) {
	userController.ListUsers(c)
}

// GetUser 获取用户详情
// @Summary 获取用户详情
// @Description 根据用户ID获取用户的详细信息
// @Tags 用户管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "用户ID"
// @Success 200 {object} response.Response{data=UserDetailResponse}
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/users/{id} [get]
func GetUser(c *gin.Context) {
	userController.GetUser(c)
}

// CreateUser 创建用户
// @Summary 创建用户
// @Description 创建新的系统用户
// @Tags 用户管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body CreateUserRequest true "用户信息"
// @Success 201 {object} response.Response{data=UserDetailResponse}
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 409 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/users [post]
func CreateUser(c *gin.Context) {
	userController.CreateUser(c)
}

// UpdateUser 更新用户
// @Summary 更新用户
// @Description 更新用户信息
// @Tags 用户管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "用户ID"
// @Param request body UpdateUserRequest true "用户信息"
// @Success 200 {object} response.Response{data=UserDetailResponse}
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/users/{id} [put]
func UpdateUser(c *gin.Context) {
	userController.UpdateUser(c)
}

// DeleteUser 删除用户
// @Summary 删除用户
// @Description 删除指定的用户
// @Tags 用户管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "用户ID"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/users/{id} [delete]
func DeleteUser(c *gin.Context) {
	userController.DeleteUser(c)
}

// GetUserStats 获取用户统计
// @Summary 获取用户统计
// @Description 获取用户数量统计信息
// @Tags 用户管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.Response{data=UserStatsResponse}
// @Failure 401 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/users/stats [get]
func GetUserStats(c *gin.Context) {
	userController.GetUserStats(c)
}

// BatchUpdateUsers 批量更新用户
// @Summary 批量更新用户
// @Description 批量更新多个用户的状态或其他属性
// @Tags 用户管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body BatchUpdateUsersRequest true "批量更新请求"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/users/batch [put]
func BatchUpdateUsers(c *gin.Context) {
	userController.BatchUpdateUsers(c)
}

// BatchDeleteUsers 批量删除用户
// @Summary 批量删除用户
// @Description 批量删除多个用户
// @Tags 用户管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body object{ids=[]string} true "用户ID列表"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/users/batch [delete]
func BatchDeleteUsers(c *gin.Context) {
	userController.BatchDeleteUsers(c)
}

// GetUserPermissions 获取用户权限
// @Summary 获取用户权限
// @Description 获取指定用户的权限列表
// @Tags 用户管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "用户ID"
// @Success 200 {object} response.Response{data=[]UserPermissionResponse}
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/users/{id}/permissions [get]
func GetUserPermissions(c *gin.Context) {
	userController.GetUserPermissions(c)
}

// AssignRoles 分配角色
// @Summary 分配角色
// @Description 为用户分配角色
// @Tags 用户管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "用户ID"
// @Param request body AssignRolesRequest true "角色ID列表"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/users/{id}/roles [post]
func AssignRoles(c *gin.Context) {
	userController.AssignRoles(c)
}

// RemoveRoles 移除角色
// @Summary 移除角色
// @Description 移除用户的角色
// @Tags 用户管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "用户ID"
// @Param request body AssignRolesRequest true "角色ID列表"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/users/{id}/roles [delete]
func RemoveRoles(c *gin.Context) {
	userController.RemoveRoles(c)
}

// ChangePassword 修改密码
// @Summary 修改密码
// @Description 修改用户密码
// @Tags 用户管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "用户ID"
// @Param request body ChangePasswordRequest true "密码信息"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/users/{id}/password [put]
func ChangePassword(c *gin.Context) {
	userController.ChangePassword(c)
}

// ResetPassword 重置密码
// @Summary 重置密码
// @Description 重置用户密码
// @Tags 用户管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "用户ID"
// @Success 200 {object} response.Response{data=object{password=string}}
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/users/{id}/password/reset [post]
func ResetPassword(c *gin.Context) {
	userController.ResetPassword(c)
}

// ActivateUser 激活用户
// @Summary 激活用户
// @Description 激活指定用户
// @Tags 用户管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "用户ID"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/users/{id}/activate [post]
func ActivateUser(c *gin.Context) {
	userController.ActivateUser(c)
}

// DeactivateUser 停用用户
// @Summary 停用用户
// @Description 停用指定用户
// @Tags 用户管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "用户ID"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/users/{id}/deactivate [post]
func DeactivateUser(c *gin.Context) {
	userController.DeactivateUser(c)
}

// LockUser 锁定用户
// @Summary 锁定用户
// @Description 锁定指定用户
// @Tags 用户管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "用户ID"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/users/{id}/lock [post]
func LockUser(c *gin.Context) {
	userController.LockUser(c)
}

// UnlockUser 解锁用户
// @Summary 解锁用户
// @Description 解锁指定用户
// @Tags 用户管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "用户ID"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/users/{id}/unlock [post]
func UnlockUser(c *gin.Context) {
	userController.UnlockUser(c)
}

// CheckUsernameExists 检查用户名是否存在
// @Summary 检查用户名是否存在
// @Description 检查指定用户名是否已存在
// @Tags 用户管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param username query string true "用户名"
// @Param exclude_id query string false "排除的用户ID"
// @Success 200 {object} response.Response{data=object{exists=bool}}
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/users/check/username [get]
func CheckUsernameExists(c *gin.Context) {
	userController.CheckUsernameExists(c)
}

// CheckEmailExists 检查邮箱是否存在
// @Summary 检查邮箱是否存在
// @Description 检查指定邮箱是否已存在
// @Tags 用户管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param email query string true "邮箱"
// @Param exclude_id query string false "排除的用户ID"
// @Success 200 {object} response.Response{data=object{exists=bool}}
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/users/check/email [get]
func CheckEmailExists(c *gin.Context) {
	userController.CheckEmailExists(c)
}