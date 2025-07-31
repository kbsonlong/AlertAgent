/**
 * API 服务基础配置
 * 提供统一的HTTP客户端和请求拦截器
 */

import axios, { AxiosInstance, InternalAxiosRequestConfig, AxiosResponse } from 'axios'

// API 基础配置
const API_BASE_URL = (import.meta as any).env?.VITE_API_BASE_URL || 'http://localhost:8080'
const API_TIMEOUT = 30000

// 创建 axios 实例
const apiClient: AxiosInstance = axios.create({
  baseURL: API_BASE_URL,
  timeout: API_TIMEOUT,
  headers: {
    'Content-Type': 'application/json',
  },
})

// 请求拦截器 - 添加认证token
apiClient.interceptors.request.use(
  (config: InternalAxiosRequestConfig) => {
    // 从localStorage获取token
    const token = localStorage.getItem('auth_token')
    if (token && config.headers) {
      config.headers.Authorization = `Bearer ${token}`
    }
    return config
  },
  (error) => {
    return Promise.reject(error)
  }
)

// 响应拦截器 - 处理通用错误
apiClient.interceptors.response.use(
  (response: AxiosResponse) => {
    return response
  },
  (error) => {
    // 处理401未授权错误
    if (error.response?.status === 401) {
      localStorage.removeItem('auth_token')
      localStorage.removeItem('user_info')
      // 重定向到登录页
      window.location.href = '/login'
    }
    
    // 处理其他HTTP错误
    const errorMessage = error.response?.data?.message || error.message || '请求失败'
    console.error('API Error:', errorMessage)
    
    return Promise.reject({
      status: error.response?.status,
      message: errorMessage,
      data: error.response?.data
    })
  }
)

// API 响应类型定义
export interface ApiResponse<T = any> {
  code: number
  message: string
  data: T
  timestamp?: string
}

// 分页响应类型
export interface PaginatedResponse<T = any> {
  items: T[]
  total: number
  page: number
  pageSize: number
  totalPages: number
}

// 通用请求方法
export class ApiService {
  /**
   * GET 请求
   */
  static async get<T = any>(url: string, params?: any): Promise<ApiResponse<T>> {
    const response = await apiClient.get(url, { params })
    return response.data
  }

  /**
   * POST 请求
   */
  static async post<T = any>(url: string, data?: any): Promise<ApiResponse<T>> {
    const response = await apiClient.post(url, data)
    return response.data
  }

  /**
   * PUT 请求
   */
  static async put<T = any>(url: string, data?: any): Promise<ApiResponse<T>> {
    const response = await apiClient.put(url, data)
    return response.data
  }

  /**
   * DELETE 请求
   */
  static async delete<T = any>(url: string): Promise<ApiResponse<T>> {
    const response = await apiClient.delete(url)
    return response.data
  }

  /**
   * PATCH 请求
   */
  static async patch<T = any>(url: string, data?: any): Promise<ApiResponse<T>> {
    const response = await apiClient.patch(url, data)
    return response.data
  }
}

// 导出 axios 实例供特殊需求使用
export { apiClient }
export default ApiService