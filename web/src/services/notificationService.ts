/**
 * 通知模板管理 API 服务
 * 提供通知模板相关的所有API调用方法
 */

import ApiService, { ApiResponse, PaginatedResponse } from './api'

// 通知模板数据类型定义
export interface NotificationTemplate {
  id: number
  name: string
  type: 'email' | 'sms' | 'webhook' | 'dingtalk' | 'wechat' | 'slack' | 'telegram'
  subject?: string
  content: string
  format: 'text' | 'html' | 'markdown' | 'json'
  variables: string[]
  description?: string
  is_default: boolean
  status: 'active' | 'inactive'
  config?: {
    webhook_url?: string
    headers?: Record<string, string>
    method?: 'GET' | 'POST' | 'PUT'
    timeout?: number
    retry_count?: number
    retry_interval?: number
  }
  usage_count: number
  last_used_at?: string
  created_at: string
  updated_at: string
  created_by: string
  updated_by: string
}

// 创建通知模板请求参数
export interface CreateNotificationTemplateRequest {
  name: string
  type: 'email' | 'sms' | 'webhook' | 'dingtalk' | 'wechat' | 'slack' | 'telegram'
  subject?: string
  content: string
  format?: 'text' | 'html' | 'markdown' | 'json'
  description?: string
  is_default?: boolean
  status?: 'active' | 'inactive'
  config?: {
    webhook_url?: string
    headers?: Record<string, string>
    method?: 'GET' | 'POST' | 'PUT'
    timeout?: number
    retry_count?: number
    retry_interval?: number
  }
}

// 更新通知模板请求参数
export interface UpdateNotificationTemplateRequest {
  name?: string
  subject?: string
  content?: string
  format?: 'text' | 'html' | 'markdown' | 'json'
  description?: string
  is_default?: boolean
  status?: 'active' | 'inactive'
  config?: {
    webhook_url?: string
    headers?: Record<string, string>
    method?: 'GET' | 'POST' | 'PUT'
    timeout?: number
    retry_count?: number
    retry_interval?: number
  }
}

// 通知模板查询参数
export interface NotificationTemplateQueryParams {
  page?: number
  pageSize?: number
  type?: string
  status?: string
  search?: string
  is_default?: boolean
  sortBy?: 'created_at' | 'updated_at' | 'name' | 'usage_count' | 'last_used_at'
  sortOrder?: 'asc' | 'desc'
}

// 通知发送请求参数
export interface SendNotificationRequest {
  template_id: number
  recipients: string[]
  variables?: Record<string, any>
  priority?: 'low' | 'normal' | 'high' | 'urgent'
  scheduled_at?: string
  retry_config?: {
    max_retries: number
    retry_interval: number
  }
}

// 通知发送结果
export interface NotificationSendResult {
  id: string
  template_id: number
  status: 'pending' | 'sending' | 'sent' | 'failed' | 'cancelled'
  recipients: Array<{
    address: string
    status: 'pending' | 'sent' | 'failed'
    error_message?: string
    sent_at?: string
  }>
  variables: Record<string, any>
  priority: 'low' | 'normal' | 'high' | 'urgent'
  scheduled_at?: string
  sent_at?: string
  error_message?: string
  retry_count: number
  created_at: string
}

// 通知历史记录
export interface NotificationHistory {
  id: string
  template_id: number
  template_name: string
  type: 'email' | 'sms' | 'webhook' | 'dingtalk' | 'wechat' | 'slack' | 'telegram'
  recipient: string
  subject?: string
  content: string
  status: 'sent' | 'failed' | 'cancelled'
  error_message?: string
  response_data?: any
  sent_at?: string
  created_at: string
  variables?: Record<string, any>
  priority: 'low' | 'normal' | 'high' | 'urgent'
  retry_count: number
  cost?: number
}

// 通知统计信息
export interface NotificationStats {
  total_sent: number
  total_failed: number
  success_rate: number
  avg_response_time: number
  stats_by_type: Array<{
    type: string
    sent_count: number
    failed_count: number
    success_rate: number
  }>
  stats_by_template: Array<{
    template_id: number
    template_name: string
    sent_count: number
    failed_count: number
    success_rate: number
  }>
  daily_stats: Array<{
    date: string
    sent_count: number
    failed_count: number
  }>
}

// 模板测试请求参数
export interface TestTemplateRequest {
  template_id?: number
  type: 'email' | 'sms' | 'webhook' | 'dingtalk' | 'wechat' | 'slack' | 'telegram'
  subject?: string
  content: string
  format: 'text' | 'html' | 'markdown' | 'json'
  config?: {
    webhook_url?: string
    headers?: Record<string, string>
    method?: 'GET' | 'POST' | 'PUT'
    timeout?: number
  }
  recipients: string[]
  variables?: Record<string, any>
}

// 模板测试结果
export interface TestTemplateResult {
  success: boolean
  rendered_content: string
  rendered_subject?: string
  response_time: number
  error_message?: string
  response_data?: any
  validation_errors?: string[]
}

// 通知渠道配置
export interface NotificationChannel {
  id: number
  name: string
  type: 'email' | 'sms' | 'webhook' | 'dingtalk' | 'wechat' | 'slack' | 'telegram'
  config: Record<string, any>
  status: 'active' | 'inactive'
  is_default: boolean
  description?: string
  created_at: string
  updated_at: string
}

/**
 * 通知服务类
 */
export class NotificationService {
  private static readonly BASE_URL = '/api/v1/notifications'

  /**
   * 获取通知模板列表
   */
  static async getTemplateList(params?: NotificationTemplateQueryParams): Promise<ApiResponse<PaginatedResponse<NotificationTemplate>>> {
    return ApiService.get(`${this.BASE_URL}/templates`, params)
  }

  /**
   * 获取通知模板详情
   */
  static async getTemplate(id: number): Promise<ApiResponse<NotificationTemplate>> {
    return ApiService.get(`${this.BASE_URL}/templates/${id}`)
  }

  /**
   * 创建通知模板
   */
  static async createTemplate(data: CreateNotificationTemplateRequest): Promise<ApiResponse<NotificationTemplate>> {
    return ApiService.post(`${this.BASE_URL}/templates`, data)
  }

  /**
   * 更新通知模板
   */
  static async updateTemplate(id: number, data: UpdateNotificationTemplateRequest): Promise<ApiResponse<NotificationTemplate>> {
    return ApiService.put(`${this.BASE_URL}/templates/${id}`, data)
  }

  /**
   * 删除通知模板
   */
  static async deleteTemplate(id: number): Promise<ApiResponse<void>> {
    return ApiService.delete(`${this.BASE_URL}/templates/${id}`)
  }

  /**
   * 复制通知模板
   */
  static async copyTemplate(id: number, name: string): Promise<ApiResponse<NotificationTemplate>> {
    return ApiService.post(`${this.BASE_URL}/templates/${id}/copy`, { name })
  }

  /**
   * 设置默认模板
   */
  static async setDefaultTemplate(id: number, type: string): Promise<ApiResponse<void>> {
    return ApiService.post(`${this.BASE_URL}/templates/${id}/set-default`, { type })
  }

  /**
   * 测试通知模板
   */
  static async testTemplate(data: TestTemplateRequest): Promise<ApiResponse<TestTemplateResult>> {
    return ApiService.post(`${this.BASE_URL}/templates/test`, data)
  }

  /**
   * 预览模板渲染结果
   */
  static async previewTemplate(id: number, variables?: Record<string, any>): Promise<ApiResponse<{
    rendered_content: string
    rendered_subject?: string
    variables_used: string[]
    variables_missing: string[]
  }>> {
    return ApiService.post(`${this.BASE_URL}/templates/${id}/preview`, { variables })
  }

  /**
   * 发送通知
   */
  static async sendNotification(data: SendNotificationRequest): Promise<ApiResponse<NotificationSendResult>> {
    return ApiService.post(`${this.BASE_URL}/send`, data)
  }

  /**
   * 获取通知发送状态
   */
  static async getNotificationStatus(id: string): Promise<ApiResponse<NotificationSendResult>> {
    return ApiService.get(`${this.BASE_URL}/send/${id}`)
  }

  /**
   * 取消通知发送
   */
  static async cancelNotification(id: string): Promise<ApiResponse<void>> {
    return ApiService.post(`${this.BASE_URL}/send/${id}/cancel`)
  }

  /**
   * 重试失败的通知
   */
  static async retryNotification(id: string): Promise<ApiResponse<NotificationSendResult>> {
    return ApiService.post(`${this.BASE_URL}/send/${id}/retry`)
  }

  /**
   * 获取通知历史记录
   */
  static async getNotificationHistory(params?: {
    template_id?: number
    type?: string
    status?: string
    recipient?: string
    start_time?: string
    end_time?: string
    page?: number
    pageSize?: number
    sortBy?: 'created_at' | 'sent_at'
    sortOrder?: 'asc' | 'desc'
  }): Promise<ApiResponse<PaginatedResponse<NotificationHistory>>> {
    return ApiService.get(`${this.BASE_URL}/history`, params)
  }

  /**
   * 获取通知统计信息
   */
  static async getNotificationStats(params?: {
    start_time?: string
    end_time?: string
    type?: string
    template_id?: number
  }): Promise<ApiResponse<NotificationStats>> {
    return ApiService.get(`${this.BASE_URL}/stats`, params)
  }

  /**
   * 获取通知渠道列表
   */
  static async getChannelList(): Promise<ApiResponse<NotificationChannel[]>> {
    return ApiService.get(`${this.BASE_URL}/channels`)
  }

  /**
   * 获取通知渠道详情
   */
  static async getChannel(id: number): Promise<ApiResponse<NotificationChannel>> {
    return ApiService.get(`${this.BASE_URL}/channels/${id}`)
  }

  /**
   * 创建通知渠道
   */
  static async createChannel(data: {
    name: string
    type: 'email' | 'sms' | 'webhook' | 'dingtalk' | 'wechat' | 'slack' | 'telegram'
    config: Record<string, any>
    description?: string
    is_default?: boolean
  }): Promise<ApiResponse<NotificationChannel>> {
    return ApiService.post(`${this.BASE_URL}/channels`, data)
  }

  /**
   * 更新通知渠道
   */
  static async updateChannel(id: number, data: {
    name?: string
    config?: Record<string, any>
    description?: string
    status?: 'active' | 'inactive'
    is_default?: boolean
  }): Promise<ApiResponse<NotificationChannel>> {
    return ApiService.put(`${this.BASE_URL}/channels/${id}`, data)
  }

  /**
   * 删除通知渠道
   */
  static async deleteChannel(id: number): Promise<ApiResponse<void>> {
    return ApiService.delete(`${this.BASE_URL}/channels/${id}`)
  }

  /**
   * 测试通知渠道
   */
  static async testChannel(id: number, data: {
    recipients: string[]
    test_message?: string
  }): Promise<ApiResponse<TestTemplateResult>> {
    return ApiService.post(`${this.BASE_URL}/channels/${id}/test`, data)
  }

  /**
   * 获取模板变量列表
   */
  static async getTemplateVariables(): Promise<ApiResponse<Array<{
    name: string
    description: string
    type: 'string' | 'number' | 'boolean' | 'object' | 'array'
    example: any
    category: string
  }>>> {
    return ApiService.get(`${this.BASE_URL}/variables`)
  }

  /**
   * 验证模板语法
   */
  static async validateTemplate(data: {
    content: string
    format: 'text' | 'html' | 'markdown' | 'json'
    variables?: Record<string, any>
  }): Promise<ApiResponse<{
    valid: boolean
    errors: string[]
    warnings: string[]
    variables_used: string[]
    variables_missing: string[]
  }>> {
    return ApiService.post(`${this.BASE_URL}/templates/validate`, data)
  }

  /**
   * 导出通知模板
   */
  static async exportTemplates(ids?: number[]): Promise<ApiResponse<{ download_url: string }>> {
    const data = ids ? { ids } : {}
    return ApiService.post(`${this.BASE_URL}/templates/export`, data)
  }

  /**
   * 导入通知模板
   */
  static async importTemplates(file: File): Promise<ApiResponse<{ imported_count: number; failed_count: number }>> {
    const formData = new FormData()
    formData.append('file', file)
    
    return ApiService.post(`${this.BASE_URL}/templates/import`, formData)
  }

  /**
   * 批量删除通知模板
   */
  static async batchDeleteTemplates(ids: number[]): Promise<ApiResponse<{ deleted_count: number }>> {
    return ApiService.delete(`${this.BASE_URL}/templates/batch`)
  }

  /**
   * 获取通知配额信息
   */
  static async getNotificationQuota(): Promise<ApiResponse<{
    email: { used: number; limit: number; remaining: number }
    sms: { used: number; limit: number; remaining: number }
    webhook: { used: number; limit: number; remaining: number }
    reset_date: string
  }>> {
    return ApiService.get(`${this.BASE_URL}/quota`)
  }
}

export default NotificationService