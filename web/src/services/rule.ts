import { get, post, put, del } from '@/utils/request'
import type { Rule, CreateRuleRequest, UpdateRuleRequest, PaginatedResponse } from '@/types'

/**
 * 获取规则列表
 * @param params 查询参数
 */
export const getRules = async (params?: {
  page?: number
  page_size?: number
  search?: string
  enabled?: boolean
}) => {
  return get<PaginatedResponse<Rule>>('/api/v1/rules', params)
}

/**
 * 获取单个规则详情
 * @param id 规则ID
 */
export const getRule = async (id: number) => {
  return get<Rule>(`/api/v1/rules/${id}`)
}

/**
 * 创建规则
 * @param data 规则数据
 */
export const createRule = async (data: CreateRuleRequest) => {
  return post<Rule>('/api/v1/rules', data)
}

/**
 * 更新规则
 * @param id 规则ID
 * @param data 更新数据
 */
export const updateRule = async (id: number, data: UpdateRuleRequest) => {
  return put<Rule>(`/api/v1/rules/${id}`, data)
}

/**
 * 删除规则
 * @param id 规则ID
 */
export const deleteRule = async (id: number) => {
  return del(`/api/v1/rules/${id}`)
}

/**
 * 启用/禁用规则
 * @param id 规则ID
 * @param enabled 是否启用
 */
export const toggleRule = async (id: number, enabled: boolean) => {
  return put<Rule>(`/api/v1/rules/${id}/toggle`, { enabled })
}

/**
 * 测试规则
 * @param expression 规则表达式
 */
export const testRule = async (expression: string) => {
  return post<{
    valid: boolean
    error?: string
    sample_data?: any[]
  }>('/api/v1/rules/test', { expression })
}

/**
 * 获取规则分发状态
 * @param id 规则ID
 */
export const getRuleDistribution = async (id: number) => {
  return get<{
    rule_id: string
    rule_name: string
    version: string
    targets: string[]
    status: string
    last_sync: string
    target_status: Array<{
      target: string
      status: string
      last_sync: string
      error?: string
    }>
  }>(`/api/v1/rules/${id}/distribution`)
}

/**
 * 同步规则到目标
 * @param id 规则ID
 * @param targets 目标列表
 */
export const syncRule = async (id: number, targets?: string[]) => {
  return post(`/api/v1/rules/${id}/sync`, { targets })
}

/**
 * 批量删除规则
 * @param ids 规则ID数组
 */
export const batchDeleteRules = async (ids: number[]) => {
  return del('/api/v1/rules/batch', {
    data: { ids }
  })
}

/**
 * 导入规则
 * @param file 规则文件
 */
export const importRules = async (file: File) => {
  const formData = new FormData()
  formData.append('file', file)
  
  return post<{
    imported_count: number
    failed_count: number
    errors?: string[]
  }>('/api/v1/rules/import', formData, {
    headers: {
      'Content-Type': 'multipart/form-data'
    }
  })
}

/**
 * 导出规则
 * @param ids 规则ID数组，为空则导出所有规则
 */
export const exportRules = async (ids?: number[]) => {
  return get('/api/v1/rules/export', { ids }, {
    responseType: 'blob'
  })
}