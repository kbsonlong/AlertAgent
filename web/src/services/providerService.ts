/**
 * 数据源管理 API 服务
 * 提供数据源相关的所有API调用方法
 */

import ApiService, { ApiResponse, PaginatedResponse } from './api'

// 数据源数据类型定义
export interface Provider {
  id: number
  name: string
  type: 'prometheus' | 'victoriametrics' | 'grafana' | 'elasticsearch' | 'influxdb'
  url: string
  description?: string
  config: Record<string, any>
  auth_config?: {
    type: 'none' | 'basic' | 'bearer' | 'api_key'
    username?: string
    password?: string
    token?: string
    api_key?: string
    headers?: Record<string, string>
  }
  status: 'active' | 'inactive' | 'error'
  health_status: 'healthy' | 'unhealthy' | 'unknown'
  last_check_at?: string
  error_message?: string
  created_at: string
  updated_at: string
  created_by: string
  updated_by: string
  metrics?: {
    total_queries: number
    avg_response_time: number
    success_rate: number
    last_query_at?: string
  }
}

// 创建数据源请求参数
export interface CreateProviderRequest {
  name: string
  type: 'prometheus' | 'victoriametrics' | 'grafana' | 'elasticsearch' | 'influxdb'
  url: string
  description?: string
  config?: Record<string, any>
  auth_config?: {
    type: 'none' | 'basic' | 'bearer' | 'api_key'
    username?: string
    password?: string
    token?: string
    api_key?: string
    headers?: Record<string, string>
  }
  status?: 'active' | 'inactive'
}

// 更新数据源请求参数
export interface UpdateProviderRequest {
  name?: string
  url?: string
  description?: string
  config?: Record<string, any>
  auth_config?: {
    type: 'none' | 'basic' | 'bearer' | 'api_key'
    username?: string
    password?: string
    token?: string
    api_key?: string
    headers?: Record<string, string>
  }
  status?: 'active' | 'inactive'
}

// 数据源查询参数
export interface ProviderQueryParams {
  page?: number
  pageSize?: number
  type?: string
  status?: string
  health_status?: string
  search?: string
  sortBy?: 'created_at' | 'updated_at' | 'name' | 'last_check_at'
  sortOrder?: 'asc' | 'desc'
}

// 数据源测试请求参数
export interface TestProviderRequest {
  name?: string
  type: 'prometheus' | 'victoriametrics' | 'grafana' | 'elasticsearch' | 'influxdb'
  url: string
  config?: Record<string, any>
  auth_config?: {
    type: 'none' | 'basic' | 'bearer' | 'api_key'
    username?: string
    password?: string
    token?: string
    api_key?: string
    headers?: Record<string, string>
  }
  test_query?: string
}

// 数据源测试结果
export interface ProviderTestResult {
  success: boolean
  response_time: number
  error_message?: string
  version?: string
  build_info?: Record<string, any>
  metrics?: {
    total_series: number
    total_samples: number
    uptime?: string
  }
  test_query_result?: {
    success: boolean
    result_count: number
    sample_data: any[]
    execution_time: number
    error?: string
  }
}

// 数据源健康检查结果
export interface ProviderHealthCheck {
  provider_id: number
  status: 'healthy' | 'unhealthy'
  response_time: number
  error_message?: string
  checked_at: string
  details?: Record<string, any>
}

// 数据源指标查询参数
export interface MetricQueryRequest {
  query: string
  start?: string
  end?: string
  step?: string
  timeout?: string
}

// 数据源指标查询结果
export interface MetricQueryResult {
  status: 'success' | 'error'
  data?: {
    resultType: 'matrix' | 'vector' | 'scalar' | 'string'
    result: any[]
  }
  error?: string
  warnings?: string[]
  execution_time: number
}

/**
 * 数据源服务类
 */
export class ProviderService {
  private static readonly BASE_URL = '/api/v1/providers'

  /**
   * 获取数据源列表
   */
  static async getProviderList(params?: ProviderQueryParams): Promise<ApiResponse<PaginatedResponse<Provider>>> {
    return ApiService.get(this.BASE_URL, params)
  }

  /**
   * 获取数据源详情
   */
  static async getProvider(id: number): Promise<ApiResponse<Provider>> {
    return ApiService.get(`${this.BASE_URL}/${id}`)
  }

  /**
   * 创建数据源
   */
  static async createProvider(data: CreateProviderRequest): Promise<ApiResponse<Provider>> {
    return ApiService.post(this.BASE_URL, data)
  }

  /**
   * 更新数据源
   */
  static async updateProvider(id: number, data: UpdateProviderRequest): Promise<ApiResponse<Provider>> {
    return ApiService.put(`${this.BASE_URL}/${id}`, data)
  }

  /**
   * 删除数据源
   */
  static async deleteProvider(id: number): Promise<ApiResponse<void>> {
    return ApiService.delete(`${this.BASE_URL}/${id}`)
  }

  /**
   * 测试数据源连接
   */
  static async testProvider(data: TestProviderRequest): Promise<ApiResponse<ProviderTestResult>> {
    return ApiService.post(`${this.BASE_URL}/test`, data)
  }

  /**
   * 测试现有数据源连接
   */
  static async testExistingProvider(id: number): Promise<ApiResponse<ProviderTestResult>> {
    return ApiService.post(`${this.BASE_URL}/${id}/test`)
  }

  /**
   * 获取数据源健康状态
   */
  static async getProviderHealth(id: number): Promise<ApiResponse<ProviderHealthCheck>> {
    return ApiService.get(`${this.BASE_URL}/${id}/health`)
  }

  /**
   * 批量检查数据源健康状态
   */
  static async batchCheckHealth(ids?: number[]): Promise<ApiResponse<ProviderHealthCheck[]>> {
    const data = ids ? { ids } : {}
    return ApiService.post(`${this.BASE_URL}/health-check`, data)
  }

  /**
   * 查询数据源指标
   */
  static async queryMetrics(id: number, data: MetricQueryRequest): Promise<ApiResponse<MetricQueryResult>> {
    return ApiService.post(`${this.BASE_URL}/${id}/query`, data)
  }

  /**
   * 获取数据源指标标签
   */
  static async getMetricLabels(id: number, metric?: string): Promise<ApiResponse<string[]>> {
    const params = metric ? { metric } : {}
    return ApiService.get(`${this.BASE_URL}/${id}/labels`, params)
  }

  /**
   * 获取数据源指标标签值
   */
  static async getMetricLabelValues(id: number, label: string, metric?: string): Promise<ApiResponse<string[]>> {
    const params = metric ? { metric } : {}
    return ApiService.get(`${this.BASE_URL}/${id}/label/${label}/values`, params)
  }

  /**
   * 获取数据源指标名称列表
   */
  static async getMetricNames(id: number, search?: string): Promise<ApiResponse<string[]>> {
    const params = search ? { search } : {}
    return ApiService.get(`${this.BASE_URL}/${id}/metrics`, params)
  }

  /**
   * 获取数据源统计信息
   */
  static async getProviderStats(id: number, params?: {
    start_time?: string
    end_time?: string
  }): Promise<ApiResponse<any>> {
    return ApiService.get(`${this.BASE_URL}/${id}/stats`, params)
  }

  /**
   * 获取所有数据源统计信息
   */
  static async getAllProvidersStats(): Promise<ApiResponse<any>> {
    return ApiService.get(`${this.BASE_URL}/stats`)
  }

  /**
   * 获取数据源类型配置模板
   */
  static async getProviderTypeConfig(type: string): Promise<ApiResponse<any>> {
    return ApiService.get(`${this.BASE_URL}/types/${type}/config`)
  }

  /**
   * 获取支持的数据源类型列表
   */
  static async getSupportedTypes(): Promise<ApiResponse<Array<{
    type: string
    name: string
    description: string
    features: string[]
    config_schema: any
  }>>> {
    return ApiService.get(`${this.BASE_URL}/types`)
  }

  /**
   * 导出数据源配置
   */
  static async exportProviders(ids?: number[]): Promise<ApiResponse<{ download_url: string }>> {
    const data = ids ? { ids } : {}
    return ApiService.post(`${this.BASE_URL}/export`, data)
  }

  /**
   * 导入数据源配置
   */
  static async importProviders(file: File): Promise<ApiResponse<{ imported_count: number; failed_count: number }>> {
    const formData = new FormData()
    formData.append('file', file)
    
    return ApiService.post(`${this.BASE_URL}/import`, formData)
  }
}

export default ProviderService