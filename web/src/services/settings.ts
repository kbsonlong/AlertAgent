import { get, put } from '@/utils/request'
import type { Settings } from '@/types'

/**
 * 获取系统设置
 * @param category 设置分类
 */
export const getSettings = async (category?: string) => {
  return get<Settings[]>('/api/v1/settings', { category })
}

/**
 * 更新系统设置
 * @param settings 设置数据
 */
export const updateSettings = async (settings: Array<{
  key: string
  value: string
}>) => {
  return put<Settings[]>('/api/v1/settings', { settings })
}

/**
 * 获取系统信息
 */
export const getSystemInfo = async () => {
  return get<{
    version: string
    build_time: string
    go_version: string
    git_commit: string
    uptime: number
    system: {
      os: string
      arch: string
      cpu_count: number
      memory_total: number
      memory_used: number
      disk_total: number
      disk_used: number
    }
    database: {
      type: string
      version: string
      connected: boolean
      connection_count: number
    }
    redis: {
      version: string
      connected: boolean
      memory_used: number
      keys_count: number
    }
    ollama: {
      connected: boolean
      models: string[]
      version?: string
    }
  }>('/api/v1/system/info')
}

/**
 * 获取系统健康状态
 */
export const getSystemHealth = async () => {
  return get<{
    status: 'healthy' | 'degraded' | 'unhealthy'
    checks: Array<{
      name: string
      status: 'pass' | 'fail' | 'warn'
      message?: string
      duration?: number
    }>
    timestamp: string
  }>('/api/v1/system/health')
}

/**
 * 获取系统指标
 * @param timeRange 时间范围（分钟）
 */
export const getSystemMetrics = async (timeRange = 60) => {
  return get<{
    cpu_usage: Array<{ timestamp: number; value: number }>
    memory_usage: Array<{ timestamp: number; value: number }>
    disk_usage: Array<{ timestamp: number; value: number }>
    network_in: Array<{ timestamp: number; value: number }>
    network_out: Array<{ timestamp: number; value: number }>
    request_count: Array<{ timestamp: number; value: number }>
    response_time: Array<{ timestamp: number; value: number }>
    error_rate: Array<{ timestamp: number; value: number }>
  }>('/api/v1/system/metrics', { time_range: timeRange })
}

/**
 * 获取系统日志
 * @param params 查询参数
 */
export const getSystemLogs = async (params?: {
  level?: 'debug' | 'info' | 'warn' | 'error'
  start_time?: string
  end_time?: string
  limit?: number
  offset?: number
  search?: string
}) => {
  return get<{
    logs: Array<{
      timestamp: string
      level: string
      message: string
      fields?: Record<string, any>
    }>
    total: number
  }>('/api/v1/system/logs', params)
}

/**
 * 清理系统缓存
 */
export const clearSystemCache = async () => {
  return put('/api/v1/system/cache/clear')
}

/**
 * 重启系统服务
 * @param service 服务名称
 */
export const restartSystemService = async (service: string) => {
  return put(`/api/v1/system/services/${service}/restart`)
}

/**
 * 备份系统数据
 */
export const backupSystemData = async () => {
  return get('/api/v1/system/backup', {}, {
    responseType: 'blob'
  })
}

/**
 * 恢复系统数据
 * @param file 备份文件
 */
export const restoreSystemData = async (file: File) => {
  const formData = new FormData()
  formData.append('file', file)
  
  return put<{
    success: boolean
    message: string
  }>('/api/v1/system/restore', formData, {
    headers: {
      'Content-Type': 'multipart/form-data'
    }
  })
}