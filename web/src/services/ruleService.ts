/**
 * 规则管理 API 服务
 * 提供告警规则相关的所有API调用方法
 */

import ApiService, { ApiResponse, PaginatedResponse } from './api'

// 规则数据类型定义
export interface Rule {
  id: number
  name: string
  description: string
  expr: string // PromQL表达式
  for_duration: string // 持续时间
  severity: 'critical' | 'warning' | 'info'
  labels: Record<string, string>
  annotations: Record<string, string>
  status: 'active' | 'inactive' | 'testing'
  group_name: string
  provider_id: number
  provider_name?: string
  created_at: string
  updated_at: string
  created_by: string
  updated_by: string
  version: number
  is_distributed: boolean
  distribution_targets?: string[]
}

// 创建规则请求参数
export interface CreateRuleRequest {
  name: string
  description: string
  expr: string
  for_duration?: string
  severity: 'critical' | 'warning' | 'info'
  labels?: Record<string, string>
  annotations?: Record<string, string>
  group_name: string
  provider_id: number
  status?: 'active' | 'inactive' | 'testing'
  distribution_targets?: string[]
}

// 更新规则请求参数
export interface UpdateRuleRequest {
  name?: string
  description?: string
  expr?: string
  for_duration?: string
  severity?: 'critical' | 'warning' | 'info'
  labels?: Record<string, string>
  annotations?: Record<string, string>
  group_name?: string
  provider_id?: number
  status?: 'active' | 'inactive' | 'testing'
  distribution_targets?: string[]
}

// 规则查询参数
export interface RuleQueryParams {
  page?: number
  pageSize?: number
  name?: string
  group_name?: string
  severity?: string
  status?: string
  provider_id?: number
  search?: string
  sortBy?: 'created_at' | 'updated_at' | 'name' | 'severity'
  sortOrder?: 'asc' | 'desc'
}

// 规则版本信息
export interface RuleVersion {
  id: number
  rule_id: number
  version: number
  name: string
  description: string
  expr: string
  for_duration: string
  severity: string
  labels: Record<string, string>
  annotations: Record<string, string>
  group_name: string
  provider_id: number
  status: string
  created_at: string
  created_by: string
  change_log?: string
  is_current: boolean
}

// 规则审计日志
export interface RuleAuditLog {
  id: number
  rule_id: number
  action: 'create' | 'update' | 'delete' | 'activate' | 'deactivate' | 'rollback'
  old_version?: number
  new_version?: number
  changes: Record<string, any>
  change_log?: string
  created_at: string
  created_by: string
  ip_address?: string
}

// 规则验证结果
export interface RuleValidationResult {
  is_valid: boolean
  errors: string[]
  warnings: string[]
  suggestions: string[]
  estimated_cost?: number
}

// 规则测试结果
export interface RuleTestResult {
  is_valid: boolean
  result_count: number
  sample_results: any[]
  execution_time: number
  errors: string[]
}

/**
 * 规则服务类
 */
export class RuleService {
  private static readonly BASE_URL = '/api/v1/rules'

  /**
   * 获取规则列表
   */
  static async getRuleList(params?: RuleQueryParams): Promise<ApiResponse<PaginatedResponse<Rule>>> {
    return ApiService.get(this.BASE_URL, params)
  }

  /**
   * 获取规则详情
   */
  static async getRule(id: number): Promise<ApiResponse<Rule>> {
    return ApiService.get(`${this.BASE_URL}/${id}`)
  }

  /**
   * 创建规则
   */
  static async createRule(data: CreateRuleRequest): Promise<ApiResponse<Rule>> {
    return ApiService.post(this.BASE_URL, data)
  }

  /**
   * 更新规则
   */
  static async updateRule(id: number, data: UpdateRuleRequest): Promise<ApiResponse<Rule>> {
    return ApiService.put(`${this.BASE_URL}/${id}`, data)
  }

  /**
   * 删除规则
   */
  static async deleteRule(id: number): Promise<ApiResponse<void>> {
    return ApiService.delete(`${this.BASE_URL}/${id}`)
  }

  /**
   * 批量删除规则
   */
  static async batchDeleteRules(ids: number[]): Promise<ApiResponse<void>> {
    return ApiService.post(`${this.BASE_URL}/batch-delete`, { ids })
  }

  /**
   * 验证规则
   */
  static async validateRule(data: { expr: string; provider_id: number }): Promise<ApiResponse<RuleValidationResult>> {
    return ApiService.post(`${this.BASE_URL}/validate`, data)
  }

  /**
   * 测试规则
   */
  static async testRule(data: { expr: string; provider_id: number; duration?: string }): Promise<ApiResponse<RuleTestResult>> {
    return ApiService.post(`${this.BASE_URL}/test`, data)
  }

  /**
   * 激活规则
   */
  static async activateRule(id: number): Promise<ApiResponse<Rule>> {
    return ApiService.post(`${this.BASE_URL}/${id}/activate`)
  }

  /**
   * 停用规则
   */
  static async deactivateRule(id: number): Promise<ApiResponse<Rule>> {
    return ApiService.post(`${this.BASE_URL}/${id}/deactivate`)
  }

  /**
   * 批量激活规则
   */
  static async batchActivateRules(ids: number[]): Promise<ApiResponse<void>> {
    return ApiService.post(`${this.BASE_URL}/batch-activate`, { ids })
  }

  /**
   * 批量停用规则
   */
  static async batchDeactivateRules(ids: number[]): Promise<ApiResponse<void>> {
    return ApiService.post(`${this.BASE_URL}/batch-deactivate`, { ids })
  }

  /**
   * 分发规则
   */
  static async distributeRule(id: number, targets: string[]): Promise<ApiResponse<void>> {
    return ApiService.post(`${this.BASE_URL}/${id}/distribute`, { targets })
  }

  /**
   * 获取规则版本列表
   */
  static async getRuleVersions(ruleId: number): Promise<ApiResponse<RuleVersion[]>> {
    return ApiService.get(`${this.BASE_URL}/${ruleId}/versions`)
  }

  /**
   * 获取规则版本详情
   */
  static async getRuleVersion(ruleId: number, version: number): Promise<ApiResponse<RuleVersion>> {
    return ApiService.get(`${this.BASE_URL}/${ruleId}/versions/${version}`)
  }

  /**
   * 回滚规则到指定版本
   */
  static async rollbackRule(ruleId: number, version: number, changeLog?: string): Promise<ApiResponse<Rule>> {
    return ApiService.post(`${this.BASE_URL}/${ruleId}/rollback`, { version, change_log: changeLog })
  }

  /**
   * 比较规则版本
   */
  static async compareRuleVersions(data: {
    rule_id: number
    version1: number
    version2: number
  }): Promise<ApiResponse<any>> {
    return ApiService.post(`${this.BASE_URL}/versions/compare`, data)
  }

  /**
   * 获取规则审计日志
   */
  static async getRuleAuditLogs(ruleId: number): Promise<ApiResponse<RuleAuditLog[]>> {
    return ApiService.get(`${this.BASE_URL}/${ruleId}/audit-logs`)
  }

  /**
   * 获取全局审计日志
   */
  static async getAllAuditLogs(params?: {
    page?: number
    pageSize?: number
    action?: string
    user?: string
    start_time?: string
    end_time?: string
  }): Promise<ApiResponse<PaginatedResponse<RuleAuditLog>>> {
    return ApiService.get(`${this.BASE_URL}/audit-logs`, params)
  }

  /**
   * 带审计的创建规则
   */
  static async createRuleWithAudit(data: CreateRuleRequest & { change_log?: string }): Promise<ApiResponse<Rule>> {
    return ApiService.post(`${this.BASE_URL}/audit`, data)
  }

  /**
   * 带审计的更新规则
   */
  static async updateRuleWithAudit(id: number, data: UpdateRuleRequest & { change_log?: string }): Promise<ApiResponse<Rule>> {
    return ApiService.put(`${this.BASE_URL}/${id}/audit`, data)
  }

  /**
   * 带审计的删除规则
   */
  static async deleteRuleWithAudit(id: number, changeLog?: string): Promise<ApiResponse<void>> {
    return ApiService.delete(`${this.BASE_URL}/${id}/audit?change_log=${encodeURIComponent(changeLog || '')}`)
  }

  /**
   * 导出规则
   */
  static async exportRules(ids?: number[]): Promise<ApiResponse<{ download_url: string }>> {
    const data = ids ? { ids } : {}
    return ApiService.post(`${this.BASE_URL}/export`, data)
  }

  /**
   * 导入规则
   */
  static async importRules(file: File): Promise<ApiResponse<{ imported_count: number; failed_count: number }>> {
    const formData = new FormData()
    formData.append('file', file)
    
    return ApiService.post(`${this.BASE_URL}/import`, formData)
  }

  /**
   * 获取规则组列表
   */
  static async getRuleGroups(): Promise<ApiResponse<string[]>> {
    return ApiService.get(`${this.BASE_URL}/groups`)
  }
}

export default RuleService