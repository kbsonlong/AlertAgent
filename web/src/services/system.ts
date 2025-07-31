import { request } from '@/utils/request'

// 系统设置接口
export interface SystemSettings {
  general?: {
    systemName: string
    version: string
    description: string
    timezone: string
    language: string
  }
  alert?: {
    defaultEvaluationInterval: number
    defaultDuration: number
    maxAlerts: number
    retentionDays: number
    autoResolve: boolean
    silenceMode: boolean
  }
  notification?: {
    defaultFrequency: number
    maxRetries: number
    sendTimeout: number
    batchSize: number
    enableQueue: boolean
  }
  storage?: {
    dbType: string
    maxConnections: number
    connectionTimeout: number
    dataRetentionDays: number
    backupInterval: number
    autoCleanup: boolean
  }
  security?: {
    sessionTimeout: number
    maxLoginAttempts: number
    lockoutDuration: number
    minPasswordLength: number
    enableHttps: boolean
    enableAuditLog: boolean
  }
}

// 系统信息接口
export interface SystemInfo {
  version: string
  buildTime: string
  goVersion: string
  uptime: string
  cpuUsage: number
  memoryUsage: number
  diskUsage: number
  connections: number
}

// 获取系统设置
export const getSystemSettings = async (): Promise<SystemSettings> => {
  const response = await request({
    url: '/api/v1/system/settings',
    method: 'GET'
  })
  return response.data
}

// 更新系统设置
export const updateSystemSettings = async (category: string, settings: any): Promise<void> => {
  await request({
    url: `/api/v1/system/settings/${category}`,
    method: 'PUT',
    data: settings
  })
}

// 获取系统信息
export const getSystemInfo = async (): Promise<SystemInfo> => {
  const response = await request({
    url: '/api/v1/system/info',
    method: 'GET'
  })
  return response.data
}

// 重启系统
export const restartSystem = async (): Promise<void> => {
  await request({
    url: '/api/v1/system/restart',
    method: 'POST'
  })
}

// 备份系统
export const backupSystem = async (): Promise<{ downloadUrl: string }> => {
  const response = await request({
    url: '/api/v1/system/backup',
    method: 'POST'
  })
  return response.data
}

// 恢复系统
export const restoreSystem = async (file: File): Promise<void> => {
  const formData = new FormData()
  formData.append('file', file)
  
  await request({
    url: '/api/v1/system/restore',
    method: 'POST',
    data: formData,
    headers: {
      'Content-Type': 'multipart/form-data'
    }
  })
}

// 检查系统健康状态
export const checkSystemHealth = async (): Promise<{
  status: 'healthy' | 'warning' | 'error'
  checks: Array<{
    name: string
    status: 'pass' | 'fail'
    message?: string
  }>
}> => {
  const response = await request({
    url: '/api/v1/system/health',
    method: 'GET'
  })
  return response.data
}

// 获取系统日志
export const getSystemLogs = async (params: {
  level?: string
  startTime?: string
  endTime?: string
  keyword?: string
  page?: number
  pageSize?: number
}): Promise<{
  list: Array<{
    id: string
    level: string
    message: string
    timestamp: string
    source: string
  }>
  total: number
}> => {
  const response = await request({
    url: '/api/v1/system/logs',
    method: 'GET',
    params
  })
  return response.data
}

// 清理系统日志
export const clearSystemLogs = async (beforeDate?: string): Promise<void> => {
  await request({
    url: '/api/v1/system/logs',
    method: 'DELETE',
    params: beforeDate ? { beforeDate } : {}
  })
}

// 获取系统统计信息
export const getSystemStats = async (): Promise<{
  alerts: {
    total: number
    active: number
    resolved: number
  }
  rules: {
    total: number
    enabled: number
    disabled: number
  }
  providers: {
    total: number
    healthy: number
    unhealthy: number
  }
  notifications: {
    total: number
    sent: number
    failed: number
  }
}> => {
  const response = await request({
    url: '/api/v1/system/stats',
    method: 'GET'
  })
  return response.data
}