import { get, post, put, del } from '@/utils/request'
import type { Provider, CreateProviderRequest, UpdateProviderRequest, PaginatedResponse } from '@/types'

/**
 * 获取数据源列表
 * @param params 查询参数
 */
export const getProviders = async (params?: {
  page?: number
  page_size?: number
  type?: string
  status?: string
  search?: string
}) => {
  return get<PaginatedResponse<Provider>>('/api/v1/providers', params)
}

/**
 * 获取单个数据源详情
 * @param id 数据源ID
 */
export const getProvider = async (id: number) => {
  return get<Provider>(`/api/v1/providers/${id}`)
}

/**
 * 创建数据源
 * @param data 数据源数据
 */
export const createProvider = async (data: CreateProviderRequest) => {
  return post<Provider>('/api/v1/providers', data)
}

/**
 * 更新数据源
 * @param id 数据源ID
 * @param data 更新数据
 */
export const updateProvider = async (id: number, data: UpdateProviderRequest) => {
  return put<Provider>(`/api/v1/providers/${id}`, data)
}

/**
 * 删除数据源
 * @param id 数据源ID
 */
export const deleteProvider = async (id: number) => {
  return del(`/api/v1/providers/${id}`)
}

/**
 * 测试数据源连接
 * @param data 数据源配置
 */
export const testProvider = async (data: {
  type: string
  endpoint: string
  auth_type?: string
  auth_config?: string
}) => {
  return post<{
    success: boolean
    message: string
    response_time?: number
    version?: string
  }>('/api/v1/providers/test', data)
}

/**
 * 获取数据源健康状态
 * @param id 数据源ID
 */
export const getProviderHealth = async (id: number) => {
  return get<{
    status: 'healthy' | 'unhealthy' | 'unknown'
    last_check: string
    response_time?: number
    error?: string
    metrics?: {
      cpu_usage?: number
      memory_usage?: number
      disk_usage?: number
      uptime?: number
    }
  }>(`/api/v1/providers/${id}/health`)
}

/**
 * 获取数据源指标
 * @param id 数据源ID
 * @param query 查询表达式
 * @param start 开始时间
 * @param end 结束时间
 * @param step 步长
 */
export const getProviderMetrics = async (id: number, params: {
  query: string
  start?: string
  end?: string
  step?: string
}) => {
  return get<{
    status: string
    data: {
      resultType: string
      result: Array<{
        metric: Record<string, string>
        values?: Array<[number, string]>
        value?: [number, string]
      }>
    }
  }>(`/api/v1/providers/${id}/query`, params)
}

/**
 * 获取数据源标签
 * @param id 数据源ID
 * @param label 标签名称
 */
export const getProviderLabels = async (id: number, label?: string) => {
  return get<string[]>(`/api/v1/providers/${id}/labels`, { label })
}

/**
 * 获取数据源指标名称列表
 * @param id 数据源ID
 */
export const getProviderMetricNames = async (id: number) => {
  return get<string[]>(`/api/v1/providers/${id}/metrics`)
}

/**
 * 批量删除数据源
 * @param ids 数据源ID数组
 */
export const batchDeleteProviders = async (ids: number[]) => {
  return del('/api/v1/providers/batch', {
    data: { ids }
  })
}

/**
 * 同步数据源配置
 * @param id 数据源ID
 */
export const syncProvider = async (id: number) => {
  return post(`/api/v1/providers/${id}/sync`)
}