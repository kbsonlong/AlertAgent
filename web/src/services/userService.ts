/**
 * 用户和组管理 API 服务
 * 提供用户认证、权限管理相关的所有API调用方法
 */

import ApiService, { ApiResponse, PaginatedResponse } from './api'

// 用户数据类型定义
export interface User {
  id: number
  username: string
  email: string
  full_name: string
  avatar?: string
  phone?: string
  department?: string
  position?: string
  status: 'active' | 'inactive' | 'locked'
  role: 'admin' | 'operator' | 'viewer'
  groups: Group[]
  permissions: Permission[]
  last_login_at?: string
  last_login_ip?: string
  login_count: number
  created_at: string
  updated_at: string
  created_by: string
  updated_by: string
  preferences?: {
    language: string
    timezone: string
    theme: 'light' | 'dark' | 'auto'
    notifications: {
      email: boolean
      sms: boolean
      browser: boolean
    }
  }
}

// 用户组数据类型定义
export interface Group {
  id: number
  name: string
  description?: string
  type: 'system' | 'custom'
  status: 'active' | 'inactive'
  permissions: Permission[]
  users: User[]
  user_count: number
  created_at: string
  updated_at: string
  created_by: string
  updated_by: string
}

// 权限数据类型定义
export interface Permission {
  id: number
  name: string
  code: string
  description?: string
  resource: string
  action: string
  category: string
  type: 'system' | 'custom'
  status: 'active' | 'inactive'
  is_system: boolean
  created_at: string
  updated_at: string
}

// 角色数据类型定义
export interface Role {
  id: number
  name: string
  code: string
  description?: string
  type: 'system' | 'custom'
  status: 'active' | 'inactive'
  permissions: Permission[]
  users: User[]
  user_count: number
  created_at: string
  updated_at: string
}

// 登录请求参数
export interface LoginRequest {
  username: string
  password: string
  remember_me?: boolean
  captcha?: string
  captcha_id?: string
}

// 登录响应
export interface LoginResponse {
  access_token: string
  refresh_token: string
  token_type: string
  expires_in: number
  user: User
}

// 注册请求参数
export interface RegisterRequest {
  username: string
  email: string
  password: string
  full_name: string
  phone?: string
  department?: string
  position?: string
  captcha?: string
  captcha_id?: string
}

// 创建用户请求参数
export interface CreateUserRequest {
  username: string
  email: string
  password: string
  full_name: string
  phone?: string
  department?: string
  position?: string
  role: 'admin' | 'operator' | 'viewer'
  status?: 'active' | 'inactive'
  group_ids?: number[]
}

// 更新用户请求参数
export interface UpdateUserRequest {
  email?: string
  full_name?: string
  phone?: string
  department?: string
  position?: string
  role?: 'admin' | 'operator' | 'viewer'
  status?: 'active' | 'inactive' | 'locked'
  group_ids?: number[]
}

// 修改密码请求参数
export interface ChangePasswordRequest {
  old_password: string
  new_password: string
  confirm_password: string
}

// 重置密码请求参数
export interface ResetPasswordRequest {
  email: string
  captcha?: string
  captcha_id?: string
}

// 用户查询参数
export interface UserQueryParams {
  page?: number
  pageSize?: number
  search?: string
  role?: string
  status?: string
  group_id?: number
  department?: string
  sortBy?: 'created_at' | 'updated_at' | 'username' | 'last_login_at'
  sortOrder?: 'asc' | 'desc'
}

// 创建用户组请求参数
export interface CreateGroupRequest {
  name: string
  description?: string
  permission_ids?: number[]
  user_ids?: number[]
}

// 更新用户组请求参数
export interface UpdateGroupRequest {
  name?: string
  description?: string
  status?: 'active' | 'inactive'
  permission_ids?: number[]
  user_ids?: number[]
}

// 用户组查询参数
export interface GroupQueryParams {
  page?: number
  pageSize?: number
  search?: string
  type?: string
  status?: string
  sortBy?: 'created_at' | 'updated_at' | 'name' | 'user_count'
  sortOrder?: 'asc' | 'desc'
}

// 用户偏好设置
export interface UserPreferences {
  language: string
  timezone: string
  theme: 'light' | 'dark' | 'auto'
  notifications: {
    email: boolean
    sms: boolean
    browser: boolean
  }
  dashboard_layout?: any
  default_page_size?: number
}

// 用户活动日志
export interface UserActivity {
  id: number
  user_id: number
  action: string
  resource: string
  resource_id?: number
  description: string
  ip_address: string
  user_agent: string
  created_at: string
  details?: Record<string, any>
}

/**
 * 用户服务类
 */
export class UserService {
  private static readonly BASE_URL = '/api/v1'

  /**
   * 用户登录
   */
  static async login(data: LoginRequest): Promise<ApiResponse<LoginResponse>> {
    return ApiService.post(`${this.BASE_URL}/auth/login`, data)
  }

  /**
   * 用户注册
   */
  static async register(data: RegisterRequest): Promise<ApiResponse<User>> {
    return ApiService.post(`${this.BASE_URL}/auth/register`, data)
  }

  /**
   * 用户登出
   */
  static async logout(): Promise<ApiResponse<void>> {
    return ApiService.post(`${this.BASE_URL}/auth/logout`)
  }

  /**
   * 刷新令牌
   */
  static async refreshToken(refreshToken: string): Promise<ApiResponse<LoginResponse>> {
    return ApiService.post(`${this.BASE_URL}/auth/refresh`, { refresh_token: refreshToken })
  }

  /**
   * 获取当前用户信息
   */
  static async getCurrentUser(): Promise<ApiResponse<User>> {
    return ApiService.get(`${this.BASE_URL}/auth/me`)
  }

  /**
   * 更新当前用户信息
   */
  static async updateCurrentUser(data: {
    full_name?: string
    email?: string
    phone?: string
    department?: string
    position?: string
  }): Promise<ApiResponse<User>> {
    return ApiService.put(`${this.BASE_URL}/auth/me`, data)
  }

  /**
   * 修改当前用户密码
   */
  static async changePassword(data: ChangePasswordRequest): Promise<ApiResponse<void>> {
    return ApiService.post(`${this.BASE_URL}/auth/change-password`, data)
  }

  /**
   * 重置密码
   */
  static async resetPassword(data: ResetPasswordRequest): Promise<ApiResponse<void>> {
    return ApiService.post(`${this.BASE_URL}/auth/reset-password`, data)
  }

  /**
   * 获取用户列表
   */
  static async getUserList(params?: UserQueryParams): Promise<ApiResponse<PaginatedResponse<User>>> {
    return ApiService.get(`${this.BASE_URL}/users`, params)
  }

  /**
   * 获取用户详情
   */
  static async getUser(id: number): Promise<ApiResponse<User>> {
    return ApiService.get(`${this.BASE_URL}/users/${id}`)
  }

  /**
   * 创建用户
   */
  static async createUser(data: CreateUserRequest): Promise<ApiResponse<User>> {
    return ApiService.post(`${this.BASE_URL}/users`, data)
  }

  /**
   * 更新用户
   */
  static async updateUser(id: number, data: UpdateUserRequest): Promise<ApiResponse<User>> {
    return ApiService.put(`${this.BASE_URL}/users/${id}`, data)
  }

  /**
   * 删除用户
   */
  static async deleteUser(id: number): Promise<ApiResponse<void>> {
    return ApiService.delete(`${this.BASE_URL}/users/${id}`)
  }

  /**
   * 锁定用户
   */
  static async lockUser(id: number): Promise<ApiResponse<void>> {
    return ApiService.post(`${this.BASE_URL}/users/${id}/lock`)
  }

  /**
   * 解锁用户
   */
  static async unlockUser(id: number): Promise<ApiResponse<void>> {
    return ApiService.post(`${this.BASE_URL}/users/${id}/unlock`)
  }

  /**
   * 重置用户密码
   */
  static async resetUserPassword(id: number, newPassword: string): Promise<ApiResponse<void>> {
    return ApiService.post(`${this.BASE_URL}/users/${id}/reset-password`, { new_password: newPassword })
  }

  /**
   * 获取用户组列表
   */
  static async getGroupList(params?: GroupQueryParams): Promise<ApiResponse<PaginatedResponse<Group>>> {
    return ApiService.get(`${this.BASE_URL}/groups`, params)
  }

  /**
   * 获取用户组详情
   */
  static async getGroup(id: number): Promise<ApiResponse<Group>> {
    return ApiService.get(`${this.BASE_URL}/groups/${id}`)
  }

  /**
   * 创建用户组
   */
  static async createGroup(data: CreateGroupRequest): Promise<ApiResponse<Group>> {
    return ApiService.post(`${this.BASE_URL}/groups`, data)
  }

  /**
   * 更新用户组
   */
  static async updateGroup(id: number, data: UpdateGroupRequest): Promise<ApiResponse<Group>> {
    return ApiService.put(`${this.BASE_URL}/groups/${id}`, data)
  }

  /**
   * 删除用户组
   */
  static async deleteGroup(id: number): Promise<ApiResponse<void>> {
    return ApiService.delete(`${this.BASE_URL}/groups/${id}`)
  }

  /**
   * 添加用户到组
   */
  static async addUsersToGroup(groupId: number, userIds: number[]): Promise<ApiResponse<void>> {
    return ApiService.post(`${this.BASE_URL}/groups/${groupId}/users`, { user_ids: userIds })
  }

  /**
   * 从组中移除用户
   */
  static async removeUsersFromGroup(groupId: number, userIds: number[]): Promise<ApiResponse<void>> {
    return ApiService.delete(`${this.BASE_URL}/groups/${groupId}/users`)
  }

  /**
   * 获取权限列表
   */
  static async getPermissionList(): Promise<ApiResponse<Permission[]>> {
    return ApiService.get(`${this.BASE_URL}/permissions`)
  }

  /**
   * 获取角色列表
   */
  static async getRoleList(): Promise<ApiResponse<Role[]>> {
    return ApiService.get(`${this.BASE_URL}/roles`)
  }

  /**
   * 获取用户权限
   */
  static async getUserPermissions(userId: number): Promise<ApiResponse<Permission[]>> {
    return ApiService.get(`${this.BASE_URL}/users/${userId}/permissions`)
  }

  /**
   * 检查用户权限
   */
  static async checkUserPermission(userId: number, permission: string): Promise<ApiResponse<{ has_permission: boolean }>> {
    return ApiService.get(`${this.BASE_URL}/users/${userId}/permissions/${permission}`)
  }

  /**
   * 获取用户偏好设置
   */
  static async getUserPreferences(): Promise<ApiResponse<UserPreferences>> {
    return ApiService.get(`${this.BASE_URL}/auth/preferences`)
  }

  /**
   * 更新用户偏好设置
   */
  static async updateUserPreferences(data: Partial<UserPreferences>): Promise<ApiResponse<UserPreferences>> {
    return ApiService.put(`${this.BASE_URL}/auth/preferences`, data)
  }

  /**
   * 获取用户活动日志
   */
  static async getUserActivities(params?: {
    user_id?: number
    action?: string
    resource?: string
    start_time?: string
    end_time?: string
    page?: number
    pageSize?: number
  }): Promise<ApiResponse<PaginatedResponse<UserActivity>>> {
    return ApiService.get(`${this.BASE_URL}/users/activities`, params)
  }

  /**
   * 获取当前用户活动日志
   */
  static async getCurrentUserActivities(params?: {
    action?: string
    resource?: string
    start_time?: string
    end_time?: string
    page?: number
    pageSize?: number
  }): Promise<ApiResponse<PaginatedResponse<UserActivity>>> {
    return ApiService.get(`${this.BASE_URL}/auth/activities`, params)
  }

  /**
   * 上传用户头像
   */
  static async uploadAvatar(file: File): Promise<ApiResponse<{ avatar_url: string }>> {
    const formData = new FormData()
    formData.append('avatar', file)
    
    return ApiService.post(`${this.BASE_URL}/auth/avatar`, formData)
  }

  /**
   * 获取验证码
   */
  static async getCaptcha(): Promise<ApiResponse<{ captcha_id: string; captcha_image: string }>> {
    return ApiService.get(`${this.BASE_URL}/auth/captcha`)
  }

  /**
   * 批量删除用户
   */
  static async batchDeleteUsers(userIds: number[]): Promise<ApiResponse<{ deleted_count: number }>> {
    return ApiService.delete(`${this.BASE_URL}/users/batch`)
  }

  /**
   * 导出用户列表
   */
  static async exportUsers(params?: UserQueryParams): Promise<ApiResponse<{ download_url: string }>> {
    return ApiService.post(`${this.BASE_URL}/users/export`, params)
  }

  /**
   * 导入用户
   */
  static async importUsers(file: File): Promise<ApiResponse<{ imported_count: number; failed_count: number }>> {
    const formData = new FormData()
    formData.append('file', file)
    
    return ApiService.post(`${this.BASE_URL}/users/import`, formData)
  }

  // ==================== 角色管理相关方法 ====================

  /**
   * 创建角色
   */
  static async createRole(data: {
    name: string
    code: string
    description?: string
    type: 'system' | 'custom'
    status?: 'active' | 'inactive'
    permission_ids?: number[]
  }): Promise<ApiResponse<Role>> {
    return ApiService.post(`${this.BASE_URL}/roles`, data)
  }

  /**
   * 更新角色
   */
  static async updateRole(id: number, data: {
    name?: string
    code?: string
    description?: string
    status?: 'active' | 'inactive'
    permission_ids?: number[]
  }): Promise<ApiResponse<Role>> {
    return ApiService.put(`${this.BASE_URL}/roles/${id}`, data)
  }

  /**
   * 删除角色
   */
  static async deleteRole(id: number): Promise<ApiResponse<void>> {
    return ApiService.delete(`${this.BASE_URL}/roles/${id}`)
  }

  /**
   * 获取角色详情
   */
  static async getRole(id: number): Promise<ApiResponse<Role>> {
    return ApiService.get(`${this.BASE_URL}/roles/${id}`)
  }

  /**
   * 获取角色用户列表
   */
  static async getRoleUsers(id: number): Promise<ApiResponse<User[]>> {
    return ApiService.get(`${this.BASE_URL}/roles/${id}/users`)
  }

  /**
   * 为角色分配用户
   */
  static async assignUsersToRole(roleId: number, userIds: number[]): Promise<ApiResponse<void>> {
    return ApiService.post(`${this.BASE_URL}/roles/${roleId}/users`, { user_ids: userIds })
  }

  /**
   * 从角色移除用户
   */
  static async removeUsersFromRole(roleId: number, userIds: number[]): Promise<ApiResponse<void>> {
    return ApiService.post(`${this.BASE_URL}/roles/${roleId}/users/remove`, { user_ids: userIds })
  }

  /**
   * 批量删除角色
   */
  static async batchDeleteRoles(roleIds: number[]): Promise<ApiResponse<{ deleted_count: number }>> {
    return ApiService.post(`${this.BASE_URL}/roles/batch/delete`, { role_ids: roleIds })
  }

  // ==================== 权限管理 API ====================

  /**
   * 创建权限
   */
  static async createPermission(data: {
    name: string
    code: string
    description?: string
    type: 'system' | 'custom'
    status?: 'active' | 'inactive'
    resource: string
    action: string
  }): Promise<ApiResponse<Permission>> {
    return ApiService.post(`${this.BASE_URL}/permissions`, data)
  }

  /**
   * 更新权限
   */
  static async updatePermission(id: number, data: {
    name?: string
    code?: string
    description?: string
    status?: 'active' | 'inactive'
    resource?: string
    action?: string
  }): Promise<ApiResponse<Permission>> {
    return ApiService.put(`${this.BASE_URL}/permissions/${id}`, data)
  }

  /**
   * 删除权限
   */
  static async deletePermission(id: number): Promise<ApiResponse<void>> {
    return ApiService.delete(`${this.BASE_URL}/permissions/${id}`)
  }

  /**
   * 获取权限详情
   */
  static async getPermission(id: number): Promise<ApiResponse<Permission>> {
    return ApiService.get(`${this.BASE_URL}/permissions/${id}`)
  }

  /**
   * 获取权限关联角色
   */
  static async getPermissionRoles(id: number): Promise<ApiResponse<Role[]>> {
    return ApiService.get(`${this.BASE_URL}/permissions/${id}/roles`)
  }

  /**
   * 批量更新权限
   */
  static async batchUpdatePermissions(permissionIds: number[], data: {
    status?: 'active' | 'inactive'
  }): Promise<ApiResponse<{ updated_count: number }>> {
    return ApiService.post(`${this.BASE_URL}/permissions/batch/update`, {
      permission_ids: permissionIds,
      ...data
    })
  }

  /**
   * 批量删除权限
   */
  static async batchDeletePermissions(permissionIds: number[]): Promise<ApiResponse<{ deleted_count: number }>> {
    return ApiService.post(`${this.BASE_URL}/permissions/batch/delete`, { permission_ids: permissionIds })
  }

  /**
   * 导出权限
   */
  static async exportPermissions(): Promise<ApiResponse<{ download_url: string }>> {
    return ApiService.get(`${this.BASE_URL}/permissions/export`)
  }
}

export default UserService