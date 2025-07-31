/**
 * 告警管理 API 服务
 * 提供告警相关的所有API调用方法
 */

import ApiService, { ApiResponse, PaginatedResponse } from './api'

// 告警数据类型定义
export interface Alert {
  id: number
  title: string
  description: string
  severity: 'critical' | 'warning' | 'info'
  status: 'firing' | 'resolved' | 'silenced' | 'pending'
  source: string
  labels: Record<string, string>
  annotations: Record<string, string>
  starts_at: string
  ends_at?: string
  created_at: string
  updated_at: string
  handled_by?: string
  handled_at?: string
  resolution_note?: string
  fingerprint: string
  generator_url?: string
}

// 创建告警请求参数
export interface CreateAlertRequest {
  title: string
  description: string
  severity: 'critical' | 'warning' | 'info'
  source: string
  labels?: Record<string, string>
  annotations?: Record<string, string>
  starts_at?: string
  generator_url?: string
}

// 更新告警请求参数
export interface UpdateAlertRequest {
  title?: string
  description?: string
  severity?: 'critical' | 'warning' | 'info'
  status?: 'firing' | 'resolved' | 'silenced' | 'pending'
  labels?: Record<string, string>
  annotations?: Record<string, string>
  ends_at?: string
  resolution_note?: string
}

// 处理告警请求参数
export interface HandleAlertRequest {
  action: 'resolve' | 'silence' | 'acknowledge'
  note?: string
  silence_duration?: number // 静默时长（分钟）
}

// 告警查询参数
export interface AlertQueryParams {
  page?: number
  pageSize?: number
  severity?: string
  status?: string
  source?: string
  search?: string
  labels?: string // 标签过滤，格式：key1=value1,key2=value2
  start_time?: string
  end_time?: string
  sortBy?: 'created_at' | 'updated_at' | 'starts_at' | 'severity'
  sortOrder?: 'asc' | 'desc'
}

// 告警分析结果
export interface AlertAnalysis {
  summary: string
  root_cause: string
  impact_assessment: string
  recommended_actions: string[]
  similar_alerts: Alert[]
  knowledge_suggestions: any[]
}

// 相似告警查询结果
export interface SimilarAlert {
  alert: Alert
  similarity_score: number
  matching_fields: string[]
}

/**
 * 告警服务类
 */
export class AlertService {
  private static readonly BASE_URL = '/api/v1/alerts'

  /**
   * 获取告警列表
   */
  static async getAlertList(params?: AlertQueryParams): Promise<ApiResponse<PaginatedResponse<Alert>>> {
    return ApiService.get(this.BASE_URL, params)
  }

  /**
   * 获取告警详情
   */
  static async getAlert(id: number): Promise<ApiResponse<Alert>> {
    return ApiService.get(`${this.BASE_URL}/${id}`)
  }

  /**
   * 创建告警
   */
  static async createAlert(data: CreateAlertRequest): Promise<ApiResponse<Alert>> {
    return ApiService.post(this.BASE_URL, data)
  }

  /**
   * 更新告警
   */
  static async updateAlert(id: number, data: UpdateAlertRequest): Promise<ApiResponse<Alert>> {
    return ApiService.put(`${this.BASE_URL}/${id}`, data)
  }

  /**
   * 处理告警
   */
  static async handleAlert(id: number, data: HandleAlertRequest): Promise<ApiResponse<Alert>> {
    return ApiService.post(`${this.BASE_URL}/${id}/handle`, data)
  }

  /**
   * 批量处理告警
   */
  static async batchHandleAlerts(ids: number[], data: HandleAlertRequest): Promise<ApiResponse<void>> {
    return ApiService.post(`${this.BASE_URL}/batch-handle`, { ids, ...data })
  }

  /**
   * 查找相似告警
   */
  static async findSimilarAlerts(id: number): Promise<ApiResponse<SimilarAlert[]>> {
    return ApiService.get(`${this.BASE_URL}/${id}/similar`)
  }

  /**
   * 分析告警
   */
  static async analyzeAlert(id: number): Promise<ApiResponse<AlertAnalysis>> {
    return ApiService.post(`${this.BASE_URL}/${id}/analyze`)
  }

  /**
   * 将告警转换为知识库
   */
  static async convertAlertToKnowledge(id: number, data: {
    title: string
    category: string
    tags?: string[]
    additional_content?: string
  }): Promise<ApiResponse<{ knowledge_id: number }>> {
    return ApiService.post(`${this.BASE_URL}/${id}/convert-to-knowledge`, data)
  }

  /**
   * 获取告警统计信息
   */
  static async getAlertStats(params?: {
    start_time?: string
    end_time?: string
    group_by?: 'severity' | 'status' | 'source' | 'hour' | 'day'
  }): Promise<ApiResponse<any>> {
    return ApiService.get(`${this.BASE_URL}/stats`, params)
  }

  /**
   * 获取告警趋势数据
   */
  static async getAlertTrends(params?: {
    start_time?: string
    end_time?: string
    interval?: 'hour' | 'day' | 'week'
    severity?: string
  }): Promise<ApiResponse<any>> {
    return ApiService.get(`${this.BASE_URL}/trends`, params)
  }

  /**
   * 导出告警数据
   */
  static async exportAlerts(params?: AlertQueryParams): Promise<ApiResponse<{ download_url: string }>> {
    return ApiService.post(`${this.BASE_URL}/export`, params)
  }

  /**
   * 获取告警标签列表
   */
  static async getAlertLabels(): Promise<ApiResponse<Record<string, string[]>>> {
    return ApiService.get(`${this.BASE_URL}/labels`)
  }

  /**
   * 获取告警来源列表
   */
  static async getAlertSources(): Promise<ApiResponse<string[]>> {
    return ApiService.get(`${this.BASE_URL}/sources`)
  }
}

export default AlertService