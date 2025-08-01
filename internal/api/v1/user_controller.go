package v1

import (
	"net/http"
	"strconv"
	"time"

	"alert_agent/internal/model"
	"alert_agent/internal/pkg/logger"
	"alert_agent/internal/pkg/response"
	"alert_agent/internal/service"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// UserController 用户控制器
type UserController struct {
	userService service.UserService
}

// NewUserController 创建用户控制器实例
func NewUserController(userService service.UserService) *UserController {
	return &UserController{
		userService: userService,
	}
}

// CreateUserRequest 创建用户请求结构
type CreateUserRequest struct {
	Username   string   `json:"username" binding:"required" example:"admin"`
	Email      string   `json:"email" binding:"required,email" example:"admin@example.com"`
	FullName   string   `json:"full_name" binding:"required" example:"管理员"`
	Password   string   `json:"password" binding:"required" example:"password123"`
	Phone      string   `json:"phone,omitempty" example:"+86 138 0013 8000"`
	Department string   `json:"department,omitempty" example:"技术部"`
	Position   string   `json:"position,omitempty" example:"系统管理员"`
	Status     string   `json:"status" binding:"required" example:"active"`
	RoleIDs    []string `json:"role_ids,omitempty" example:"[\"role-id-1\", \"role-id-2\"]"`
}

// UpdateUserRequest 更新用户请求结构
type UpdateUserRequest struct {
	Username   string   `json:"username" binding:"required" example:"admin"`
	Email      string   `json:"email" binding:"required,email" example:"admin@example.com"`
	FullName   string   `json:"full_name" binding:"required" example:"管理员"`
	Phone      string   `json:"phone,omitempty" example:"+86 138 0013 8000"`
	Department string   `json:"department,omitempty" example:"技术部"`
	Position   string   `json:"position,omitempty" example:"系统管理员"`
	Status     string   `json:"status" binding:"required" example:"active"`
	RoleIDs    []string `json:"role_ids,omitempty" example:"[\"role-id-1\", \"role-id-2\"]"`
}

// UserDetailResponse 用户详情响应结构
type UserDetailResponse struct {
	ID           string                    `json:"id" example:"user-id-1"`
	Username     string                    `json:"username" example:"admin"`
	Email        string                    `json:"email" example:"admin@example.com"`
	FullName     string                    `json:"full_name" example:"管理员"`
	Phone        string                    `json:"phone,omitempty" example:"+86 138 0013 8000"`
	Department   string                    `json:"department,omitempty" example:"技术部"`
	Position     string                    `json:"position,omitempty" example:"系统管理员"`
	Status       string                    `json:"status" example:"active"`
	LastLoginAt  *time.Time                `json:"last_login_at,omitempty" example:"2024-01-01T12:00:00Z"`
	LastLoginIP  string                    `json:"last_login_ip,omitempty" example:"192.168.1.100"`
	LoginCount   int                       `json:"login_count" example:"10"`
	CreatedAt    time.Time                 `json:"created_at" example:"2024-01-01T12:00:00Z"`
	UpdatedAt    time.Time                 `json:"updated_at" example:"2024-01-01T12:00:00Z"`
	CreatedBy    string                    `json:"created_by" example:"admin"`
	UpdatedBy    string                    `json:"updated_by" example:"admin"`
	Roles        []UserRoleResponse        `json:"roles,omitempty"`
	Permissions  []UserPermissionResponse  `json:"permissions,omitempty"`
}

// UserRoleResponse 用户角色响应结构
type UserRoleResponse struct {
	ID          string `json:"id" example:"role-id-1"`
	Name        string `json:"name" example:"管理员"`
	Code        string `json:"code" example:"admin"`
	Description string `json:"description,omitempty" example:"系统管理员角色"`
	Type        string `json:"type" example:"system"`
	Status      string `json:"status" example:"active"`
}

// UserPermissionResponse 用户权限响应结构
type UserPermissionResponse struct {
	ID          string `json:"id" example:"permission-id-1"`
	Name        string `json:"name" example:"用户管理"`
	Code        string `json:"code" example:"user:manage"`
	Description string `json:"description,omitempty" example:"用户管理权限"`
	Resource    string `json:"resource" example:"user"`
	Action      string `json:"action" example:"manage"`
	Status      string `json:"status" example:"active"`
}

// UserListDetailResponse 用户列表详情响应结构
type UserListDetailResponse struct {
	Users []UserDetailResponse `json:"users"`
	Total int64               `json:"total" example:"100"`
	Page  int                 `json:"page" example:"1"`
	Size  int                 `json:"size" example:"10"`
}

// BatchUpdateUsersRequest 批量更新用户请求结构
type BatchUpdateUsersRequest struct {
	IDs     []string           `json:"ids" binding:"required"`
	Updates map[string]interface{} `json:"updates" binding:"required"`
}

// AssignRolesRequest 分配角色请求结构
type AssignRolesRequest struct {
	RoleIDs []string `json:"role_ids" binding:"required"`
}

// ChangePasswordRequest 修改密码请求结构
type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required"`
}

// convertUserToResponse 转换用户模型为响应结构
func (uc *UserController) convertUserToResponse(user *model.User) UserDetailResponse {
	response := UserDetailResponse{
		ID:          user.ID,
		Username:    user.Username,
		Email:       user.Email,
		FullName:    user.FullName,
		Phone:       user.Phone,
		Department:  user.Department,
		Position:    user.Position,
		Status:      user.Status,
		LastLoginAt: user.LastLoginAt,
		LastLoginIP: user.LastLoginIP,
		LoginCount:  user.LoginCount,
		CreatedAt:   user.CreatedAt,
		UpdatedAt:   user.UpdatedAt,
		CreatedBy:   user.CreatedBy,
		UpdatedBy:   user.UpdatedBy,
	}
	
	// 转换角色
	if len(user.Roles) > 0 {
		response.Roles = make([]UserRoleResponse, 0, len(user.Roles))
		for _, role := range user.Roles {
			response.Roles = append(response.Roles, UserRoleResponse{
				ID:          role.ID,
				Name:        role.Name,
				Code:        role.Code,
				Description: role.Description,
				Type:        role.Type,
				Status:      role.Status,
			})
		}
	}
	
	return response
}

// convertPermissionsToResponse 转换权限模型为响应结构
func (uc *UserController) convertPermissionsToResponse(permissions []*model.Permission) []UserPermissionResponse {
	response := make([]UserPermissionResponse, 0, len(permissions))
	for _, permission := range permissions {
		response = append(response, UserPermissionResponse{
			ID:          permission.ID,
			Name:        permission.Name,
			Code:        permission.Code,
			Description: permission.Description,
			Resource:    permission.Resource,
			Action:      permission.Action,
			Status:      permission.Status,
		})
	}
	return response
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
// @Param department query string false "部门筛选"
// @Param full_name query string false "姓名筛选"
// @Success 200 {object} response.Response{data=UserListDetailResponse}
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/users [get]
func (uc *UserController) ListUsers(c *gin.Context) {
	// 获取查询参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "10"))
	
	// 构建筛选条件
	filters := make(map[string]interface{})
	if username := c.Query("username"); username != "" {
		filters["username"] = username
	}
	if email := c.Query("email"); email != "" {
		filters["email"] = email
	}
	if role := c.Query("role"); role != "" {
		filters["role"] = role
	}
	if status := c.Query("status"); status != "" {
		filters["status"] = status
	}
	if department := c.Query("department"); department != "" {
		filters["department"] = department
	}
	if fullName := c.Query("full_name"); fullName != "" {
		filters["full_name"] = fullName
	}
	
	logger.L.Info("Listing users",
		zap.Int("page", page),
		zap.Int("size", size),
		zap.Any("filters", filters),
	)
	
	// 获取用户列表
	users, total, err := uc.userService.List(c.Request.Context(), filters, page, size)
	if err != nil {
		logger.L.Error("Failed to list users", zap.Error(err))
		response.Error(c, http.StatusInternalServerError, "获取用户列表失败", err)
		return
	}
	
	// 转换响应格式
	userResponses := make([]UserDetailResponse, 0, len(users))
	for _, user := range users {
		userResponses = append(userResponses, uc.convertUserToResponse(user))
	}
	
	response.Success(c, UserListDetailResponse{
		Users: userResponses,
		Total: total,
		Page:  page,
		Size:  size,
	})
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
func (uc *UserController) GetUser(c *gin.Context) {
	userID := c.Param("id")
	if userID == "" {
		response.BadRequest(c, "用户ID不能为空", nil)
		return
	}
	
	logger.L.Info("Getting user", zap.String("user_id", userID))
	
	// 获取用户信息
	user, err := uc.userService.GetByID(c.Request.Context(), userID)
	if err != nil {
		logger.L.Error("Failed to get user", zap.Error(err))
		if err.Error() == "user not found" {
			response.NotFound(c, "用户不存在", err)
		} else {
			response.Error(c, http.StatusInternalServerError, "获取用户信息失败", err)
		}
		return
	}
	
	// 获取用户权限
	permissions, err := uc.userService.GetUserPermissions(c.Request.Context(), userID)
	if err != nil {
		logger.L.Error("Failed to get user permissions", zap.Error(err))
		// 不返回错误，继续执行
	}
	
	// 转换响应格式
	userResponse := uc.convertUserToResponse(user)
	if permissions != nil {
		userResponse.Permissions = uc.convertPermissionsToResponse(permissions)
	}
	
	response.Success(c, userResponse)
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
func (uc *UserController) CreateUser(c *gin.Context) {
	var req CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}
	
	logger.L.Info("Creating user",
		zap.String("username", req.Username),
		zap.String("email", req.Email),
	)
	
	// 构建用户模型
	user := &model.User{
		Username:   req.Username,
		Email:      req.Email,
		FullName:   req.FullName,
		Password:   req.Password,
		Phone:      req.Phone,
		Department: req.Department,
		Position:   req.Position,
		Status:     req.Status,
		CreatedBy:  "admin", // TODO: 从上下文获取当前用户
		UpdatedBy:  "admin", // TODO: 从上下文获取当前用户
	}
	
	// 创建用户
	if err := uc.userService.Create(c.Request.Context(), user); err != nil {
		logger.L.Error("Failed to create user", zap.Error(err))
		if err.Error() == "username already exists" || err.Error() == "email already exists" {
			response.Conflict(c, err.Error(), err)
		} else {
			response.Error(c, http.StatusInternalServerError, "创建用户失败", err)
		}
		return
	}
	
	// 分配角色
	if len(req.RoleIDs) > 0 {
		if err := uc.userService.AssignRoles(c.Request.Context(), user.ID, req.RoleIDs); err != nil {
			logger.L.Error("Failed to assign roles", zap.Error(err))
			// 不返回错误，继续执行
		}
	}
	
	// 获取创建后的用户信息
	createdUser, err := uc.userService.GetByID(c.Request.Context(), user.ID)
	if err != nil {
		logger.L.Error("Failed to get created user", zap.Error(err))
		response.Error(c, http.StatusInternalServerError, "获取创建的用户信息失败", err)
		return
	}
	
	c.JSON(http.StatusCreated, response.Response{
		Code:    201,
		Message: "用户创建成功",
		Data:    uc.convertUserToResponse(createdUser),
	})
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
func (uc *UserController) UpdateUser(c *gin.Context) {
	userID := c.Param("id")
	if userID == "" {
		response.BadRequest(c, "用户ID不能为空", nil)
		return
	}
	
	var req UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}
	
	logger.L.Info("Updating user",
		zap.String("user_id", userID),
		zap.String("username", req.Username),
		zap.String("email", req.Email),
	)
	
	// 获取现有用户信息
	existingUser, err := uc.userService.GetByID(c.Request.Context(), userID)
	if err != nil {
		logger.L.Error("Failed to get user", zap.Error(err))
		if err.Error() == "user not found" {
			response.NotFound(c, "用户不存在", err)
		} else {
			response.Error(c, http.StatusInternalServerError, "获取用户信息失败", err)
		}
		return
	}
	
	// 更新用户信息
	existingUser.Username = req.Username
	existingUser.Email = req.Email
	existingUser.FullName = req.FullName
	existingUser.Phone = req.Phone
	existingUser.Department = req.Department
	existingUser.Position = req.Position
	existingUser.Status = req.Status
	existingUser.UpdatedBy = "admin" // TODO: 从上下文获取当前用户
	
	// 更新用户
	if err := uc.userService.Update(c.Request.Context(), existingUser); err != nil {
		logger.L.Error("Failed to update user", zap.Error(err))
		if err.Error() == "username already exists" || err.Error() == "email already exists" {
			response.Conflict(c, err.Error(), err)
		} else {
			response.Error(c, http.StatusInternalServerError, "更新用户失败", err)
		}
		return
	}
	
	// 同步角色
	if len(req.RoleIDs) > 0 {
		if err := uc.userService.SyncRoles(c.Request.Context(), userID, req.RoleIDs); err != nil {
			logger.L.Error("Failed to sync roles", zap.Error(err))
			// 不返回错误，继续执行
		}
	}
	
	// 获取更新后的用户信息
	updatedUser, err := uc.userService.GetByID(c.Request.Context(), userID)
	if err != nil {
		logger.L.Error("Failed to get updated user", zap.Error(err))
		response.Error(c, http.StatusInternalServerError, "获取更新的用户信息失败", err)
		return
	}
	
	response.SuccessWithMessage(c, "用户更新成功", uc.convertUserToResponse(updatedUser))
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
func (uc *UserController) DeleteUser(c *gin.Context) {
	userID := c.Param("id")
	if userID == "" {
		response.BadRequest(c, "用户ID不能为空", nil)
		return
	}
	
	logger.L.Info("Deleting user", zap.String("user_id", userID))
	
	// 删除用户
	if err := uc.userService.Delete(c.Request.Context(), userID); err != nil {
		logger.L.Error("Failed to delete user", zap.Error(err))
		if err.Error() == "user not found" {
			response.NotFound(c, "用户不存在", err)
		} else {
			response.Error(c, http.StatusInternalServerError, "删除用户失败", err)
		}
		return
	}
	
	response.SuccessWithMessage(c, "用户删除成功", nil)
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
func (uc *UserController) GetUserStats(c *gin.Context) {
	logger.L.Info("Getting user stats")
	
	// 获取用户统计数据
	stats, err := uc.userService.GetStats(c.Request.Context())
	if err != nil {
		logger.L.Error("Failed to get user stats", zap.Error(err))
		response.Error(c, http.StatusInternalServerError, "获取用户统计信息失败", err)
		return
	}
	
	// 转换为响应格式
	responseStats := UserStatsResponse{
		Total:    stats.Total,
		Active:   stats.Active,
		Inactive: stats.Inactive,
		Locked:   stats.Locked,
	}
	
	response.Success(c, responseStats)
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
func (uc *UserController) BatchUpdateUsers(c *gin.Context) {
	var req BatchUpdateUsersRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}
	
	logger.L.Info("Batch updating users",
		zap.Strings("user_ids", req.IDs),
		zap.Any("updates", req.Updates),
	)
	
	// 批量更新用户
	if err := uc.userService.BatchUpdate(c.Request.Context(), req.IDs, req.Updates); err != nil {
		logger.L.Error("Failed to batch update users", zap.Error(err))
		response.Error(c, http.StatusInternalServerError, "批量更新用户失败", err)
		return
	}
	
	response.SuccessWithMessage(c, "批量更新成功", nil)
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
func (uc *UserController) BatchDeleteUsers(c *gin.Context) {
	var req struct {
		IDs []string `json:"ids" binding:"required"`
	}
	
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}
	
	logger.L.Info("Batch deleting users", zap.Strings("user_ids", req.IDs))
	
	// 批量删除用户
	if err := uc.userService.BatchDelete(c.Request.Context(), req.IDs); err != nil {
		logger.L.Error("Failed to batch delete users", zap.Error(err))
		response.Error(c, http.StatusInternalServerError, "批量删除用户失败", err)
		return
	}
	
	response.SuccessWithMessage(c, "批量删除成功", nil)
}

// GetUserPermissions 获取用户权限
// @Summary 获取用户权限
// @Description 获取指定用户的权限列表
// @Tags 用户管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "用户ID"
// @Success 200 {object} response.Response{data=[]PermissionResponse}
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/users/{id}/permissions [get]
func (uc *UserController) GetUserPermissions(c *gin.Context) {
	userID := c.Param("id")
	if userID == "" {
		response.BadRequest(c, "用户ID不能为空", nil)
		return
	}
	
	logger.L.Info("Getting user permissions", zap.String("user_id", userID))
	
	// 获取用户权限
	permissions, err := uc.userService.GetUserPermissions(c.Request.Context(), userID)
	if err != nil {
		logger.L.Error("Failed to get user permissions", zap.Error(err))
		if err.Error() == "user not found" {
			response.NotFound(c, "用户不存在", err)
		} else {
			response.Error(c, http.StatusInternalServerError, "获取用户权限失败", err)
		}
		return
	}
	
	response.Success(c, uc.convertPermissionsToResponse(permissions))
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
func (uc *UserController) AssignRoles(c *gin.Context) {
	userID := c.Param("id")
	if userID == "" {
		response.BadRequest(c, "用户ID不能为空", nil)
		return
	}
	
	var req AssignRolesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}
	
	logger.L.Info("Assigning roles to user",
		zap.String("user_id", userID),
		zap.Strings("role_ids", req.RoleIDs),
	)
	
	// 分配角色
	if err := uc.userService.AssignRoles(c.Request.Context(), userID, req.RoleIDs); err != nil {
		logger.L.Error("Failed to assign roles", zap.Error(err))
		if err.Error() == "user not found" {
			response.NotFound(c, "用户不存在", err)
		} else {
			response.Error(c, http.StatusInternalServerError, "分配角色失败", err)
		}
		return
	}
	
	response.SuccessWithMessage(c, "角色分配成功", nil)
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
func (uc *UserController) RemoveRoles(c *gin.Context) {
	userID := c.Param("id")
	if userID == "" {
		response.BadRequest(c, "用户ID不能为空", nil)
		return
	}
	
	var req AssignRolesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}
	
	logger.L.Info("Removing roles from user",
		zap.String("user_id", userID),
		zap.Strings("role_ids", req.RoleIDs),
	)
	
	// 移除角色
	if err := uc.userService.RemoveRoles(c.Request.Context(), userID, req.RoleIDs); err != nil {
		logger.L.Error("Failed to remove roles", zap.Error(err))
		if err.Error() == "user not found" {
			response.NotFound(c, "用户不存在", err)
		} else {
			response.Error(c, http.StatusInternalServerError, "移除角色失败", err)
		}
		return
	}
	
	response.SuccessWithMessage(c, "角色移除成功", nil)
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
func (uc *UserController) ChangePassword(c *gin.Context) {
	userID := c.Param("id")
	if userID == "" {
		response.BadRequest(c, "用户ID不能为空", nil)
		return
	}
	
	var req ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}
	
	logger.L.Info("Changing password for user", zap.String("user_id", userID))
	
	// 修改密码
	if err := uc.userService.ChangePassword(c.Request.Context(), userID, req.OldPassword, req.NewPassword); err != nil {
		logger.L.Error("Failed to change password", zap.Error(err))
		if err.Error() == "user not found" {
			response.NotFound(c, "用户不存在", err)
		} else if err.Error() == "invalid old password" {
			response.BadRequest(c, "旧密码错误", err)
		} else {
			response.Error(c, http.StatusInternalServerError, "修改密码失败", err)
		}
		return
	}
	
	response.SuccessWithMessage(c, "密码修改成功", nil)
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
func (uc *UserController) ResetPassword(c *gin.Context) {
	userID := c.Param("id")
	if userID == "" {
		response.BadRequest(c, "用户ID不能为空", nil)
		return
	}
	
	logger.L.Info("Resetting password for user", zap.String("user_id", userID))
	
	// 重置密码
	newPassword, err := uc.userService.ResetPassword(c.Request.Context(), userID)
	if err != nil {
		logger.L.Error("Failed to reset password", zap.Error(err))
		if err.Error() == "user not found" {
			response.NotFound(c, "用户不存在", err)
		} else {
			response.Error(c, http.StatusInternalServerError, "重置密码失败", err)
		}
		return
	}
	
	response.SuccessWithMessage(c, "密码重置成功", map[string]string{
		"password": newPassword,
	})
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
func (uc *UserController) ActivateUser(c *gin.Context) {
	userID := c.Param("id")
	if userID == "" {
		response.BadRequest(c, "用户ID不能为空", nil)
		return
	}
	
	logger.L.Info("Activating user", zap.String("user_id", userID))
	
	// 激活用户
	if err := uc.userService.Activate(c.Request.Context(), userID); err != nil {
		logger.L.Error("Failed to activate user", zap.Error(err))
		if err.Error() == "user not found" {
			response.NotFound(c, "用户不存在", err)
		} else {
			response.Error(c, http.StatusInternalServerError, "激活用户失败", err)
		}
		return
	}
	
	response.SuccessWithMessage(c, "用户激活成功", nil)
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
func (uc *UserController) DeactivateUser(c *gin.Context) {
	userID := c.Param("id")
	if userID == "" {
		response.BadRequest(c, "用户ID不能为空", nil)
		return
	}
	
	logger.L.Info("Deactivating user", zap.String("user_id", userID))
	
	// 停用用户
	if err := uc.userService.Deactivate(c.Request.Context(), userID); err != nil {
		logger.L.Error("Failed to deactivate user", zap.Error(err))
		if err.Error() == "user not found" {
			response.NotFound(c, "用户不存在", err)
		} else {
			response.Error(c, http.StatusInternalServerError, "停用用户失败", err)
		}
		return
	}
	
	response.SuccessWithMessage(c, "用户停用成功", nil)
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
func (uc *UserController) LockUser(c *gin.Context) {
	userID := c.Param("id")
	if userID == "" {
		response.BadRequest(c, "用户ID不能为空", nil)
		return
	}
	
	logger.L.Info("Locking user", zap.String("user_id", userID))
	
	// 锁定用户
	if err := uc.userService.Lock(c.Request.Context(), userID); err != nil {
		logger.L.Error("Failed to lock user", zap.Error(err))
		if err.Error() == "user not found" {
			response.NotFound(c, "用户不存在", err)
		} else {
			response.Error(c, http.StatusInternalServerError, "锁定用户失败", err)
		}
		return
	}
	
	response.SuccessWithMessage(c, "用户锁定成功", nil)
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
func (uc *UserController) UnlockUser(c *gin.Context) {
	userID := c.Param("id")
	if userID == "" {
		response.BadRequest(c, "用户ID不能为空", nil)
		return
	}
	
	logger.L.Info("Unlocking user", zap.String("user_id", userID))
	
	// 解锁用户
	if err := uc.userService.Unlock(c.Request.Context(), userID); err != nil {
		logger.L.Error("Failed to unlock user", zap.Error(err))
		if err.Error() == "user not found" {
			response.NotFound(c, "用户不存在", err)
		} else {
			response.Error(c, http.StatusInternalServerError, "解锁用户失败", err)
		}
		return
	}
	
	response.SuccessWithMessage(c, "用户解锁成功", nil)
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
func (uc *UserController) CheckUsernameExists(c *gin.Context) {
	username := c.Query("username")
	if username == "" {
		response.BadRequest(c, "用户名不能为空", nil)
		return
	}
	
	excludeID := c.Query("exclude_id")
	
	logger.L.Info("Checking username exists",
		zap.String("username", username),
		zap.String("exclude_id", excludeID),
	)
	
	// 检查用户名是否存在
	exists, err := uc.userService.CheckUsernameExists(c.Request.Context(), username, excludeID)
	if err != nil {
		logger.L.Error("Failed to check username exists", zap.Error(err))
		response.Error(c, http.StatusInternalServerError, "检查用户名失败", err)
		return
	}
	
	response.Success(c, map[string]bool{
		"exists": exists,
	})
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
func (uc *UserController) CheckEmailExists(c *gin.Context) {
	email := c.Query("email")
	if email == "" {
		response.BadRequest(c, "邮箱不能为空", nil)
		return
	}
	
	excludeID := c.Query("exclude_id")
	
	logger.L.Info("Checking email exists",
		zap.String("email", email),
		zap.String("exclude_id", excludeID),
	)
	
	// 检查邮箱是否存在
	exists, err := uc.userService.CheckEmailExists(c.Request.Context(), email, excludeID)
	if err != nil {
		logger.L.Error("Failed to check email exists", zap.Error(err))
		response.Error(c, http.StatusInternalServerError, "检查邮箱失败", err)
		return
	}
	
	response.Success(c, map[string]bool{
		"exists": exists,
	})
}