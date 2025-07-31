/**
 * 系统配置管理 API 服务
 * 提供系统设置相关的所有API调用方法
 */

import ApiService, { ApiResponse } from './api'

// 系统配置数据类型定义
export interface SystemConfig {
  id: number
  category: string
  key: string
  value: any
  description?: string
  type: 'string' | 'number' | 'boolean' | 'json' | 'array'
  is_public: boolean
  is_editable: boolean
  validation_rule?: string
  default_value?: any
  created_at: string
  updated_at: string
  updated_by: string
}

// 系统配置分组
export interface SystemConfigGroup {
  category: string
  name: string
  description?: string
  configs: SystemConfig[]
}

// 更新系统配置请求参数
export interface UpdateSystemConfigRequest {
  value: any
  description?: string
}

// 批量更新系统配置请求参数
export interface BatchUpdateSystemConfigRequest {
  configs: Array<{
    key: string
    value: any
    category?: string
  }>
}

// 系统信息
export interface SystemInfo {
  version: string
  build_time: string
  git_commit: string
  go_version: string
  os: string
  arch: string
  uptime: string
  start_time: string
  timezone: string
  database: {
    type: string
    version: string
    connection_status: 'connected' | 'disconnected'
    max_connections: number
    active_connections: number
  }
  redis?: {
    version: string
    connection_status: 'connected' | 'disconnected'
    memory_usage: string
    connected_clients: number
  }
  storage: {
    total_space: string
    used_space: string
    free_space: string
    usage_percentage: number
  }
  memory: {
    total: string
    used: string
    free: string
    usage_percentage: number
  }
  cpu: {
    cores: number
    usage_percentage: number
  }
}

// 系统统计信息
export interface SystemStats {
  total_users: number
  total_groups: number
  total_rules: number
  active_rules: number
  total_alerts: number
  unhandled_alerts: number
  total_knowledge: number
  total_providers: number
  active_providers: number
  system_health: 'healthy' | 'warning' | 'critical'
  last_updated: string
}

// 系统健康检查结果
export interface SystemHealthCheck {
  overall_status: 'healthy' | 'warning' | 'critical'
  checks: Array<{
    name: string
    status: 'healthy' | 'warning' | 'critical'
    message?: string
    details?: Record<string, any>
    checked_at: string
  }>
  last_check_at: string
}

// 系统日志查询参数
export interface SystemLogQueryParams {
  level?: 'debug' | 'info' | 'warn' | 'error' | 'fatal'
  module?: string
  start_time?: string
  end_time?: string
  search?: string
  page?: number
  pageSize?: number
  sortOrder?: 'asc' | 'desc'
}

// 系统日志条目
export interface SystemLogEntry {
  id: string
  timestamp: string
  level: 'debug' | 'info' | 'warn' | 'error' | 'fatal'
  module: string
  message: string
  details?: Record<string, any>
  user_id?: number
  user_name?: string
  ip_address?: string
  request_id?: string
}

// 系统备份信息
export interface SystemBackup {
  id: string
  name: string
  description?: string
  type: 'full' | 'config' | 'data'
  status: 'pending' | 'running' | 'completed' | 'failed'
  file_size?: number
  file_path?: string
  download_url?: string
  created_at: string
  completed_at?: string
  error_message?: string
  created_by: string
}

// 创建备份请求参数
export interface CreateBackupRequest {
  name: string
  description?: string
  type: 'full' | 'config' | 'data'
  include_files?: boolean
  compress?: boolean
}

// 系统恢复请求参数
export interface RestoreBackupRequest {
  backup_id: string
  restore_type: 'full' | 'config' | 'data'
  overwrite_existing?: boolean
}

/**
 * 系统服务类
 */
export class SystemService {
  private static readonly BASE_URL = '/api/v1/system'

  /**
   * 获取系统信息
   */
  static async getSystemInfo(): Promise<ApiResponse<SystemInfo>> {
    return ApiService.get(`${this.BASE_URL}/info`)
  }

  /**
   * 获取系统统计信息
   */
  static async getSystemStats(): Promise<ApiResponse<SystemStats>> {
    return ApiService.get(`${this.BASE_URL}/stats`)
  }

  /**
   * 获取系统健康检查结果
   */
  static async getSystemHealth(): Promise<ApiResponse<SystemHealthCheck>> {
    return ApiService.get(`${this.BASE_URL}/health`)
  }

  /**
   * 执行系统健康检查
   */
  static async performHealthCheck(): Promise<ApiResponse<SystemHealthCheck>> {
    return ApiService.post(`${this.BASE_URL}/health/check`)
  }

  /**
   * 获取所有系统配置
   */
  static async getAllConfigs(): Promise<ApiResponse<SystemConfigGroup[]>> {
    return ApiService.get(`${this.BASE_URL}/configs`)
  }

  /**
   * 获取指定分类的系统配置
   */
  static async getConfigsByCategory(category: string): Promise<ApiResponse<SystemConfig[]>> {
    return ApiService.get(`${this.BASE_URL}/configs/${category}`)
  }

  /**
   * 获取单个系统配置
   */
  static async getConfig(category: string, key: string): Promise<ApiResponse<SystemConfig>> {
    return ApiService.get(`${this.BASE_URL}/configs/${category}/${key}`)
  }

  /**
   * 更新系统配置
   */
  static async updateConfig(
    category: string,
    key: string,
    data: UpdateSystemConfigRequest
  ): Promise<ApiResponse<SystemConfig>> {
    return ApiService.put(`${this.BASE_URL}/configs/${category}/${key}`, data)
  }

  /**
   * 批量更新系统配置
   */
  static async batchUpdateConfigs(data: BatchUpdateSystemConfigRequest): Promise<ApiResponse<SystemConfig[]>> {
    return ApiService.put(`${this.BASE_URL}/configs/batch`, data)
  }

  /**
   * 重置系统配置为默认值
   */
  static async resetConfig(category: string, key: string): Promise<ApiResponse<SystemConfig>> {
    return ApiService.post(`${this.BASE_URL}/configs/${category}/${key}/reset`)
  }

  /**
   * 重置分类下所有配置为默认值
   */
  static async resetCategoryConfigs(category: string): Promise<ApiResponse<SystemConfig[]>> {
    return ApiService.post(`${this.BASE_URL}/configs/${category}/reset`)
  }

  /**
   * 获取系统日志
   */
  static async getSystemLogs(params?: SystemLogQueryParams): Promise<ApiResponse<{
    logs: SystemLogEntry[]
    total: number
    page: number
    pageSize: number
  }>> {
    return ApiService.get(`${this.BASE_URL}/logs`, params)
  }

  /**
   * 清理系统日志
   */
  static async clearSystemLogs(params?: {
    before_date?: string
    level?: string
    module?: string
  }): Promise<ApiResponse<{ deleted_count: number }>> {
    return ApiService.delete(`${this.BASE_URL}/logs`)
  }

  /**
   * 导出系统日志
   */
  static async exportSystemLogs(params?: SystemLogQueryParams): Promise<ApiResponse<{ download_url: string }>> {
    return ApiService.post(`${this.BASE_URL}/logs/export`, params)
  }

  /**
   * 获取系统备份列表
   */
  static async getBackups(): Promise<ApiResponse<SystemBackup[]>> {
    return ApiService.get(`${this.BASE_URL}/backups`)
  }

  /**
   * 创建系统备份
   */
  static async createBackup(data: CreateBackupRequest): Promise<ApiResponse<SystemBackup>> {
    return ApiService.post(`${this.BASE_URL}/backups`, data)
  }

  /**
   * 获取备份详情
   */
  static async getBackup(id: string): Promise<ApiResponse<SystemBackup>> {
    return ApiService.get(`${this.BASE_URL}/backups/${id}`)
  }

  /**
   * 删除备份
   */
  static async deleteBackup(id: string): Promise<ApiResponse<void>> {
    return ApiService.delete(`${this.BASE_URL}/backups/${id}`)
  }

  /**
   * 下载备份文件
   */
  static async downloadBackup(id: string): Promise<ApiResponse<{ download_url: string }>> {
    return ApiService.get(`${this.BASE_URL}/backups/${id}/download`)
  }

  /**
   * 恢复系统备份
   */
  static async restoreBackup(data: RestoreBackupRequest): Promise<ApiResponse<{ task_id: string }>> {
    return ApiService.post(`${this.BASE_URL}/restore`, data)
  }

  /**
   * 获取恢复任务状态
   */
  static async getRestoreStatus(taskId: string): Promise<ApiResponse<{
    status: 'pending' | 'running' | 'completed' | 'failed'
    progress: number
    message?: string
    error?: string
    started_at: string
    completed_at?: string
  }>> {
    return ApiService.get(`${this.BASE_URL}/restore/${taskId}/status`)
  }

  /**
   * 重启系统服务
   */
  static async restartSystem(): Promise<ApiResponse<{ message: string }>> {
    return ApiService.post(`${this.BASE_URL}/restart`)
  }

  /**
   * 获取系统维护状态
   */
  static async getMaintenanceStatus(): Promise<ApiResponse<{
    enabled: boolean
    message?: string
    start_time?: string
    end_time?: string
    created_by?: string
  }>> {
    return ApiService.get(`${this.BASE_URL}/maintenance`)
  }

  /**
   * 启用系统维护模式
   */
  static async enableMaintenance(data: {
    message?: string
    duration?: number // 分钟
  }): Promise<ApiResponse<void>> {
    return ApiService.post(`${this.BASE_URL}/maintenance/enable`, data)
  }

  /**
   * 禁用系统维护模式
   */
  static async disableMaintenance(): Promise<ApiResponse<void>> {
    return ApiService.post(`${this.BASE_URL}/maintenance/disable`)
  }

  /**
   * 获取系统配置分类列表
   */
  static async getConfigCategories(): Promise<ApiResponse<Array<{
    category: string
    name: string
    description?: string
    config_count: number
  }>>> {
    return ApiService.get(`${this.BASE_URL}/configs/categories`)
  }

  /**
   * 验证系统配置
   */
  static async validateConfig(category: string, key: string, value: any): Promise<ApiResponse<{
    valid: boolean
    error?: string
    suggestions?: string[]
  }>> {
    return ApiService.post(`${this.BASE_URL}/configs/${category}/${key}/validate`, { value })
  }
}

export default SystemService