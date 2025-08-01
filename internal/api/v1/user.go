package v1

import (
	"net/http"
	"strconv"

	"alert_agent/internal/pkg/logger"
	"alert_agent/internal/pkg/response"
	"alert_agent/internal/service"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
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
// @Success 200 {object} response.Response{data=UserListResponse}
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/users [get]
func ListUsers(c *gin.Context) {
	// 获取查询参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "10"))
	username := c.Query("username")
	email := c.Query("email")
	role := c.Query("role")
	status := c.Query("status")

	logger.L.Info("Listing users",
		zap.Int("page", page),
		zap.Int("size", size),
		zap.String("username", username),
		zap.String("email", email),
		zap.String("role", role),
		zap.String("status", status),
	)

	// 模拟用户数据
	users := []UserResponse{
		{
			ID:          1,
			Username:    "admin",
			Email:       "admin@example.com",
			FullName:    "系统管理员",
			Phone:       "+86 138 0013 8000",
			Department:  "技术部",
			Position:    "系统管理员",
			Role:        "admin",
			Status:      "active",
			LoginCount:  25,
			CreatedAt:   "2024-01-01T12:00:00Z",
			UpdatedAt:   "2024-01-01T12:00:00Z",
			CreatedBy:   "system",
			UpdatedBy:   "admin",
			Permissions: []string{"user:read", "user:write", "alert:read", "alert:write", "rule:read", "rule:write"},
		},
		{
			ID:          2,
			Username:    "operator",
			Email:       "operator@example.com",
			FullName:    "运维人员",
			Phone:       "+86 138 0013 8001",
			Department:  "运维部",
			Position:    "运维工程师",
			Role:        "operator",
			Status:      "active",
			LoginCount:  15,
			CreatedAt:   "2024-01-02T12:00:00Z",
			UpdatedAt:   "2024-01-02T12:00:00Z",
			CreatedBy:   "admin",
			UpdatedBy:   "admin",
			Permissions: []string{"alert:read", "alert:write", "rule:read"},
		},
		{
			ID:          3,
			Username:    "viewer",
			Email:       "viewer@example.com",
			FullName:    "观察者",
			Phone:       "+86 138 0013 8002",
			Department:  "业务部",
			Position:    "业务分析师",
			Role:        "viewer",
			Status:      "active",
			LoginCount:  8,
			CreatedAt:   "2024-01-03T12:00:00Z",
			UpdatedAt:   "2024-01-03T12:00:00Z",
			CreatedBy:   "admin",
			UpdatedBy:   "admin",
			Permissions: []string{"alert:read", "rule:read"},
		},
	}

	// 应用筛选
	filteredUsers := make([]UserResponse, 0)
	for _, user := range users {
		if username != "" && user.Username != username {
			continue
		}
		if email != "" && user.Email != email {
			continue
		}
		if role != "" && user.Role != role {
			continue
		}
		if status != "" && user.Status != status {
			continue
		}
		filteredUsers = append(filteredUsers, user)
	}

	// 分页处理
	total := len(filteredUsers)
	start := (page - 1) * size
	end := start + size
	if start > total {
		start = total
	}
	if end > total {
		end = total
	}

	pagedUsers := filteredUsers[start:end]

	response.Success(c, UserListResponse{
		Users: pagedUsers,
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
// @Param id path int true "用户ID"
// @Success 200 {object} response.Response{data=UserResponse}
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/users/{id} [get]
func GetUser(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的用户ID", err)
		return
	}

	logger.L.Info("Getting user", zap.Uint64("user_id", id))

	// 模拟获取用户数据
	if id == 1 {
		user := UserResponse{
			ID:          1,
			Username:    "admin",
			Email:       "admin@example.com",
			FullName:    "系统管理员",
			Phone:       "+86 138 0013 8000",
			Department:  "技术部",
			Position:    "系统管理员",
			Role:        "admin",
			Status:      "active",
			LoginCount:  25,
			CreatedAt:   "2024-01-01T12:00:00Z",
			UpdatedAt:   "2024-01-01T12:00:00Z",
			CreatedBy:   "system",
			UpdatedBy:   "admin",
			Permissions: []string{"user:read", "user:write", "alert:read", "alert:write", "rule:read", "rule:write"},
		}
		response.Success(c, user)
		return
	}

	response.NotFound(c, "用户不存在", nil)
}

// CreateUser 创建用户
// @Summary 创建用户
// @Description 创建新的系统用户
// @Tags 用户管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body UserRequest true "用户信息"
// @Success 201 {object} response.Response{data=UserResponse}
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 409 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/users [post]
func CreateUser(c *gin.Context) {
	var req UserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	logger.L.Info("Creating user",
		zap.String("username", req.Username),
		zap.String("email", req.Email),
		zap.String("role", req.Role),
	)

	// 模拟创建用户
	user := UserResponse{
		ID:          4,
		Username:    req.Username,
		Email:       req.Email,
		FullName:    req.FullName,
		Phone:       req.Phone,
		Department:  req.Department,
		Position:    req.Position,
		Role:        req.Role,
		Status:      req.Status,
		LoginCount:  0,
		CreatedAt:   "2024-01-04T12:00:00Z",
		UpdatedAt:   "2024-01-04T12:00:00Z",
		CreatedBy:   "admin",
		UpdatedBy:   "admin",
		Permissions: []string{"alert:read"},
	}

	c.JSON(http.StatusCreated, response.Response{
		Code:    201,
		Message: "用户创建成功",
		Data:    user,
	})
}

// UpdateUser 更新用户
// @Summary 更新用户
// @Description 更新用户信息
// @Tags 用户管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "用户ID"
// @Param request body UserRequest true "用户信息"
// @Success 200 {object} response.Response{data=UserResponse}
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/users/{id} [put]
func UpdateUser(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的用户ID", err)
		return
	}

	var req UserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	logger.L.Info("Updating user",
		zap.Uint64("user_id", id),
		zap.String("username", req.Username),
		zap.String("email", req.Email),
	)

	// 模拟更新用户
	user := UserResponse{
		ID:          uint(id),
		Username:    req.Username,
		Email:       req.Email,
		FullName:    req.FullName,
		Phone:       req.Phone,
		Department:  req.Department,
		Position:    req.Position,
		Role:        req.Role,
		Status:      req.Status,
		LoginCount:  10,
		CreatedAt:   "2024-01-01T12:00:00Z",
		UpdatedAt:   "2024-01-04T12:00:00Z",
		CreatedBy:   "admin",
		UpdatedBy:   "admin",
		Permissions: []string{"alert:read", "rule:read"},
	}

	response.SuccessWithMessage(c, "用户更新成功", user)
}

// DeleteUser 删除用户
// @Summary 删除用户
// @Description 删除指定的用户
// @Tags 用户管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "用户ID"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/users/{id} [delete]
func DeleteUser(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的用户ID", err)
		return
	}

	logger.L.Info("Deleting user", zap.Uint64("user_id", id))

	// 模拟删除用户
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
// GetUserStats 获取用户统计信息
// @Summary 获取用户统计信息
// @Description 获取系统中用户的统计信息，包括总数、活跃数、非活跃数和锁定数
// @Tags 用户管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.Response{data=UserStatsResponse}
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/users/stats [get]
func GetUserStats(c *gin.Context) {
	logger.L.Info("Getting user stats")

	// 创建用户统计服务
	userStatsService := service.NewUserStatsService()

	// 获取用户统计数据
	stats, err := userStatsService.GetUserStats(c.Request.Context())
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
// @Description 批量更新多个用户的状态或角色
// @Tags 用户管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body object{ids=[]int,status=string,role=string} true "批量更新请求"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/users/batch [put]
func BatchUpdateUsers(c *gin.Context) {
	var req struct {
		IDs    []uint `json:"ids" binding:"required"`
		Status string `json:"status,omitempty"`
		Role   string `json:"role,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	logger.L.Info("Batch updating users",
		zap.Any("user_ids", req.IDs),
		zap.String("status", req.Status),
		zap.String("role", req.Role),
	)

	// 模拟批量更新
	response.SuccessWithMessage(c, "批量更新成功", nil)
}

// GetUserPermissions 获取用户权限
// @Summary 获取用户权限
// @Description 获取指定用户的权限列表
// @Tags 用户管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "用户ID"
// @Success 200 {object} response.Response{data=[]string}
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/users/{id}/permissions [get]
func GetUserPermissions(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的用户ID", err)
		return
	}

	logger.L.Info("Getting user permissions", zap.Uint64("user_id", id))

	// 模拟权限数据
	permissions := []string{
		"user:read",
		"user:write",
		"alert:read",
		"alert:write",
		"rule:read",
		"rule:write",
	}

	response.Success(c, permissions)
}

// UpdateUserPermissions 更新用户权限
// @Summary 更新用户权限
// @Description 更新指定用户的权限列表
// @Tags 用户管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "用户ID"
// @Param request body object{permissions=[]string} true "权限列表"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/users/{id}/permissions [put]
func UpdateUserPermissions(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的用户ID", err)
		return
	}

	var req struct {
		Permissions []string `json:"permissions" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	logger.L.Info("Updating user permissions",
		zap.Uint64("user_id", id),
		zap.Strings("permissions", req.Permissions),
	)

	// 模拟更新权限
	response.SuccessWithMessage(c, "权限更新成功", nil)
}