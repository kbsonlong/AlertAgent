import { get, post, put } from '@/utils/request'
import type { Alert, AlertAnalysis, AsyncTask, AsyncTaskResult, PaginatedResponse } from '@/types'

/**
 * 获取告警列表
 * @param params 查询参数
 */
export const getAlerts = async (params?: {
  page?: number
  page_size?: number
  status?: string
  level?: string
  search?: string
}) => {
  return get<PaginatedResponse<Alert>>('/api/v1/alerts', params)
}

/**
 * 获取单个告警详情
 * @param id 告警ID
 */
export const getAlert = async (id: number) => {
  return get<Alert>(`/api/v1/alerts/${id}`)
}

/**
 * 更新告警状态
 * @param id 告警ID
 * @param data 更新数据
 */
export const updateAlert = async (id: number, data: {
  status?: 'acknowledged' | 'resolved'
  handler?: string
  handle_note?: string
}) => {
  return put<Alert>(`/api/v1/alerts/${id}`, data)
}

/**
 * 同步分析告警
 * @param id 告警ID
 */
export const analyzeAlert = async (id: number) => {
  return post<AlertAnalysis>(`/api/v1/alerts/${id}/analyze`)
}

/**
 * 异步分析告警
 * @param id 告警ID
 */
export const asyncAnalyzeAlert = async (id: number) => {
  return post<AsyncTask>(`/api/v1/alerts/${id}/async/analyze`)
}

/**
 * 获取异步分析结果
 * @param taskId 任务ID
 */
export const getAnalysisResult = async (taskId: number) => {
  return get<AsyncTaskResult>(`/api/v1/alerts/async/result/${taskId}`)
}

/**
 * 转换告警为知识库
 * @param id 告警ID
 */
export const convertToKnowledge = async (id: number) => {
  return post<{ id: number }>(`/api/v1/alerts/${id}/convert-to-knowledge`)
}

/**
 * 批量更新告警状态
 * @param ids 告警ID数组
 * @param status 状态
 */
export const batchUpdateAlerts = async (ids: number[], status: 'acknowledged' | 'resolved') => {
  return put<{ updated_count: number }>('/api/v1/alerts/batch', {
    ids,
    status
  })
}

/**
 * 获取告警统计信息
 */
export const getAlertStats = async () => {
  return get<{
    total: number
    firing: number
    acknowledged: number
    resolved: number
    by_level: Record<string, number>
    by_source: Record<string, number>
  }>('/api/v1/alerts/stats')
}