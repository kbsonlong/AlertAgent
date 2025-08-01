import { get, post, put } from '../utils/request'
import type { User, ApiResponse } from '@/types'

export interface UpdateProfileRequest {
  email: string
}

export interface ChangePasswordRequest {
  currentPassword: string
  newPassword: string
}

export const userApi = {
  // 获取当前用户信息
  getCurrentUser(): Promise<ApiResponse<User>> {
    return get('/api/v1/auth/profile')
  },

  // 更新个人资料
  updateProfile(data: UpdateProfileRequest): Promise<ApiResponse<User>> {
    return put('/api/v1/auth/profile', data)
  },

  // 修改密码
  changePassword(data: ChangePasswordRequest): Promise<ApiResponse<void>> {
    return put('/api/v1/auth/password', data)
  },

  // 发送邮箱验证邮件
  sendVerificationEmail(): Promise<ApiResponse<void>> {
    return post('/api/v1/auth/verify-email')
  },

  // 上传头像
  uploadAvatar(file: File): Promise<ApiResponse<{ url: string }>> {
    const formData = new FormData()
    formData.append('avatar', file)
    return post('/api/v1/auth/avatar', formData, {
      headers: {
        'Content-Type': 'multipart/form-data'
      }
    })
  }
}