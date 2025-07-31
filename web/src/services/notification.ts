import { get, post, put, del } from '@/utils/request'
import type { NotificationGroup, NotificationTemplate, PaginatedResponse } from '@/types'

// 通知组相关API

/**
 * 获取通知组列表
 * @param params 查询参数
 */
export const getNotificationGroups = async (params?: {
  page?: number
  page_size?: number
  search?: string
  enabled?: boolean
}) => {
  return get<PaginatedResponse<NotificationGroup>>('/api/v1/groups', params)
}

/**
 * 获取单个通知组详情
 * @param id 通知组ID
 */
export const getNotificationGroup = async (id: number) => {
  return get<NotificationGroup>(`/api/v1/groups/${id}`)
}

/**
 * 创建通知组
 * @param data 通知组数据
 */
export const createNotificationGroup = async (data: {
  name: string
  description?: string
  channels: Array<{
    type: 'email' | 'webhook' | 'slack' | 'dingtalk'
    config: Record<string, any>
    enabled?: boolean
  }>
  enabled?: boolean
}) => {
  return post<NotificationGroup>('/api/v1/groups', data)
}

/**
 * 更新通知组
 * @param id 通知组ID
 * @param data 更新数据
 */
export const updateNotificationGroup = async (id: number, data: {
  name?: string
  description?: string
  channels?: Array<{
    id?: number
    type: 'email' | 'webhook' | 'slack' | 'dingtalk'
    config: Record<string, any>
    enabled?: boolean
  }>
  enabled?: boolean
}) => {
  return put<NotificationGroup>(`/api/v1/groups/${id}`, data)
}

/**
 * 删除通知组
 * @param id 通知组ID
 */
export const deleteNotificationGroup = async (id: number) => {
  return del(`/api/v1/groups/${id}`)
}

/**
 * 测试通知组
 * @param id 通知组ID
 * @param message 测试消息
 */
export const testNotificationGroup = async (id: number, message?: string) => {
  return post<{
    success: boolean
    results: Array<{
      channel_id: number
      channel_type: string
      success: boolean
      error?: string
    }>
  }>(`/api/v1/groups/${id}/test`, { message })
}

// 通知模板相关API

/**
 * 获取通知模板列表
 * @param params 查询参数
 */
export const getNotificationTemplates = async (params?: {
  page?: number
  page_size?: number
  type?: string
  search?: string
  enabled?: boolean
}) => {
  return get<PaginatedResponse<NotificationTemplate>>('/api/v1/templates', params)
}

/**
 * 获取单个通知模板详情
 * @param id 模板ID
 */
export const getNotificationTemplate = async (id: number) => {
  return get<NotificationTemplate>(`/api/v1/templates/${id}`)
}

/**
 * 创建通知模板
 * @param data 模板数据
 */
export const createNotificationTemplate = async (data: {
  name: string
  type: 'email' | 'webhook' | 'slack' | 'dingtalk'
  subject?: string
  content: string
  enabled?: boolean
}) => {
  return post<NotificationTemplate>('/api/v1/templates', data)
}

/**
 * 更新通知模板
 * @param id 模板ID
 * @param data 更新数据
 */
export const updateNotificationTemplate = async (id: number, data: {
  name?: string
  type?: 'email' | 'webhook' | 'slack' | 'dingtalk'
  subject?: string
  content?: string
  enabled?: boolean
}) => {
  return put<NotificationTemplate>(`/api/v1/templates/${id}`, data)
}

/**
 * 删除通知模板
 * @param id 模板ID
 */
export const deleteNotificationTemplate = async (id: number) => {
  return del(`/api/v1/templates/${id}`)
}

/**
 * 预览通知模板
 * @param id 模板ID
 * @param data 模板变量
 */
export const previewNotificationTemplate = async (id: number, data?: Record<string, any>) => {
  return post<{
    subject?: string
    content: string
  }>(`/api/v1/templates/${id}/preview`, data)
}

/**
 * 测试通知模板
 * @param id 模板ID
 * @param data 测试数据
 */
export const testNotificationTemplate = async (id: number, data: {
  channel_config: Record<string, any>
  template_data?: Record<string, any>
}) => {
  return post<{
    success: boolean
    error?: string
  }>(`/api/v1/templates/${id}/test`, data)
}

/**
 * 获取模板变量说明
 * @param type 模板类型
 */
export const getTemplateVariables = async (type: string) => {
  return get<Array<{
    name: string
    description: string
    type: string
    example?: any
  }>>(`/api/v1/templates/variables/${type}`)
}

/**
 * 批量删除通知组
 * @param ids 通知组ID数组
 */
export const batchDeleteNotificationGroups = async (ids: number[]) => {
  return del('/api/v1/groups/batch', {
    data: { ids }
  })
}

/**
 * 批量删除通知模板
 * @param ids 模板ID数组
 */
export const batchDeleteNotificationTemplates = async (ids: number[]) => {
  return del('/api/v1/templates/batch', {
    data: { ids }
  })
}