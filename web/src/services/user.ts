import { request } from '@/utils/request'
import type { PaginatedResponse } from '@/types'

// 用户接口
export interface User {
  id: string
  username: string
  email: string
  avatar?: string
  role: 'admin' | 'user' | 'viewer'
  status: 'active' | 'disabled'
  lastLogin?: string
  createdAt: string
  updatedAt: string
  profile?: {
    firstName?: string
    lastName?: string
    phone?: string
    department?: string
    position?: string
  }
  permissions?: string[]
}

// 用户统计接口
export interface UserStats {
  total: number
  active: number
  disabled: number
  online: number
}

// 用户创建/更新数据
export interface UserFormData {
  username: string
  email: string
  password?: string
  role: string
  status: string
  avatar?: string
  profile?: {
    firstName?: string
    lastName?: string
    phone?: string
    department?: string
    position?: string
  }
}

// 获取用户列表
export const getUserList = async (params: {
  page?: number
  pageSize?: number
  username?: string
  email?: string
  role?: string
  status?: string
  sortBy?: string
  sortOrder?: string
}): Promise<PaginatedResponse<User>> => {
  const response = await request({
    url: '/api/v1/users',
    method: 'GET',
    params
  })
  return response.data
}

// 获取用户详情
export const getUserDetail = async (id: string): Promise<User> => {
  const response = await request({
    url: `/api/v1/users/${id}`,
    method: 'GET'
  })
  return response.data
}

// 创建用户
export const createUser = async (userData: UserFormData): Promise<User> => {
  const response = await request({
    url: '/api/v1/users',
    method: 'POST',
    data: userData
  })
  return response.data
}

// 更新用户
export const updateUser = async (id: string, userData: Partial<UserFormData>): Promise<User> => {
  const response = await request({
    url: `/api/v1/users/${id}`,
    method: 'PUT',
    data: userData
  })
  return response.data
}

// 删除用户
export const deleteUser = async (id: string): Promise<void> => {
  await request({
    url: `/api/v1/users/${id}`,
    method: 'DELETE'
  })
}

// 批量删除用户
export const batchDeleteUsers = async (ids: string[]): Promise<void> => {
  await request({
    url: '/api/v1/users/batch',
    method: 'DELETE',
    data: { ids }
  })
}

// 重置用户密码
export const resetUserPassword = async (id: string): Promise<{ newPassword: string }> => {
  const response = await request({
    url: `/api/v1/users/${id}/reset-password`,
    method: 'POST'
  })
  return response.data
}

// 切换用户状态
export const toggleUserStatus = async (id: string): Promise<void> => {
  await request({
    url: `/api/v1/users/${id}/toggle-status`,
    method: 'POST'
  })
}

// 批量更新用户状态
export const batchUpdateUserStatus = async (ids: string[], status: string): Promise<void> => {
  await request({
    url: '/api/v1/users/batch/status',
    method: 'PUT',
    data: { ids, status }
  })
}

// 获取用户统计
export const getUserStats = async (): Promise<UserStats> => {
  const response = await request({
    url: '/api/v1/users/stats',
    method: 'GET'
  })
  return response.data
}

// 导入用户
export const importUsers = async (file: File): Promise<{
  success: number
  failed: number
  errors?: Array<{ row: number; message: string }>
}> => {
  const formData = new FormData()
  formData.append('file', file)
  
  const response = await request({
    url: '/api/v1/users/import',
    method: 'POST',
    data: formData,
    headers: {
      'Content-Type': 'multipart/form-data'
    }
  })
  return response.data
}

// 导出用户
export const exportUsers = async (params?: {
  username?: string
  email?: string
  role?: string
  status?: string
}): Promise<Blob> => {
  const response = await request({
    url: '/api/v1/users/export',
    method: 'GET',
    params,
    responseType: 'blob'
  })
  return response.data
}

// 获取用户权限
export const getUserPermissions = async (id: string): Promise<string[]> => {
  const response = await request({
    url: `/api/v1/users/${id}/permissions`,
    method: 'GET'
  })
  return response.data
}

// 更新用户权限
export const updateUserPermissions = async (id: string, permissions: string[]): Promise<void> => {
  await request({
    url: `/api/v1/users/${id}/permissions`,
    method: 'PUT',
    data: { permissions }
  })
}

// 获取用户角色列表
export const getUserRoles = async (): Promise<Array<{
  value: string
  label: string
  description?: string
  permissions: string[]
}>> => {
  const response = await request({
    url: '/api/v1/users/roles',
    method: 'GET'
  })
  return response.data
}

// 检查用户名是否可用
export const checkUsernameAvailable = async (username: string, excludeId?: string): Promise<boolean> => {
  const response = await request({
    url: '/api/v1/users/check-username',
    method: 'GET',
    params: { username, excludeId }
  })
  return response.data.available
}

// 检查邮箱是否可用
export const checkEmailAvailable = async (email: string, excludeId?: string): Promise<boolean> => {
  const response = await request({
    url: '/api/v1/users/check-email',
    method: 'GET',
    params: { email, excludeId }
  })
  return response.data.available
}

// 获取用户活动日志
export const getUserActivityLogs = async (id: string, params?: {
  page?: number
  pageSize?: number
  startTime?: string
  endTime?: string
  action?: string
}): Promise<PaginatedResponse<{
  id: string
  action: string
  description: string
  ip: string
  userAgent: string
  createdAt: string
}>> => {
  const response = await request({
    url: `/api/v1/users/${id}/activity-logs`,
    method: 'GET',
    params
  })
  return response.data
}

// 发送用户邀请
export const sendUserInvitation = async (email: string, role: string): Promise<void> => {
  await request({
    url: '/api/v1/users/invite',
    method: 'POST',
    data: { email, role }
  })
}

// 重新发送用户邀请
export const resendUserInvitation = async (id: string): Promise<void> => {
  await request({
    url: `/api/v1/users/${id}/resend-invitation`,
    method: 'POST'
  })
}

// 获取在线用户列表
export const getOnlineUsers = async (): Promise<Array<{
  id: string
  username: string
  avatar?: string
  lastActivity: string
  ip: string
}>> => {
  const response = await request({
    url: '/api/v1/users/online',
    method: 'GET'
  })
  return response.data
}

// 强制用户下线
export const forceUserOffline = async (id: string): Promise<void> => {
  await request({
    url: `/api/v1/users/${id}/force-offline`,
    method: 'POST'
  })
}