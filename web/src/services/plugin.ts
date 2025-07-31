import { get, post, put, del } from '@/utils/request'

// 插件信息类型
export interface PluginInfo {
  name: string
  version: string
  description: string
  schema: Record<string, any>
  status: 'active' | 'inactive' | 'error'
  load_time: string
  last_error?: string
}

// 插件配置类型
export interface PluginConfig {
  name: string
  enabled: boolean
  config: Record<string, any>
  priority: number
}

// 插件测试结果类型
export interface PluginTestResult {
  success: boolean
  error?: string
  duration: number
  timestamp: string
}

// 插件统计信息类型
export interface PluginStats {
  name: string
  total_sent: number
  success_count: number
  failure_count: number
  avg_duration: number
  last_sent: string
  last_error?: string
}

// 插件健康状态类型
export interface PluginHealthStatus {
  name: string
  status: 'healthy' | 'unhealthy' | 'unknown'
  last_check: string
  error?: string
}

/**
 * 获取所有可用插件
 */
export const getAvailablePlugins = async () => {
  return get<PluginInfo[]>('/api/v1/plugins')
}

/**
 * 获取插件配置
 * @param name 插件名称
 */
export const getPluginConfig = async (name: string) => {
  return get<PluginConfig>(`/api/v1/plugins/${name}/config`)
}

/**
 * 设置插件配置
 * @param name 插件名称
 * @param config 插件配置
 */
export const setPluginConfig = async (name: string, config: PluginConfig) => {
  return post<string>(`/api/v1/plugins/${name}/config`, config)
}

/**
 * 测试插件配置
 * @param name 插件名称
 * @param config 测试配置
 */
export const testPluginConfig = async (name: string, config: Record<string, any>) => {
  return post<PluginTestResult>(`/api/v1/plugins/${name}/test`, { config })
}

/**
 * 获取插件统计信息
 * @param name 插件名称
 */
export const getPluginStats = async (name: string) => {
  return get<PluginStats>(`/api/v1/plugins/${name}/stats`)
}

/**
 * 获取所有插件统计信息
 */
export const getAllPluginStats = async () => {
  return get<Record<string, PluginStats>>('/api/v1/plugins/stats')
}

/**
 * 获取插件健康状态
 * @param name 插件名称
 */
export const getPluginHealthStatus = async (name: string) => {
  return get<PluginHealthStatus>(`/api/v1/plugins/${name}/health`)
}

/**
 * 获取所有插件健康状态
 */
export const getAllPluginHealthStatus = async () => {
  return get<Record<string, PluginHealthStatus>>('/api/v1/plugins/health')
}

/**
 * 启用插件
 * @param name 插件名称
 */
export const enablePlugin = async (name: string) => {
  return post<string>(`/api/v1/plugins/${name}/enable`)
}

/**
 * 禁用插件
 * @param name 插件名称
 */
export const disablePlugin = async (name: string) => {
  return post<string>(`/api/v1/plugins/${name}/disable`)
}

/**
 * 获取插件配置Schema
 * @param name 插件名称
 */
export const getPluginSchema = async (name: string) => {
  return get<Record<string, any>>(`/api/v1/plugins/${name}/schema`)
}