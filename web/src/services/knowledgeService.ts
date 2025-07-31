/**
 * 知识库 API 服务
 * 提供知识库相关的所有API调用方法
 */

import ApiService, { ApiResponse, PaginatedResponse } from './api'

// 知识库数据类型定义
export interface Knowledge {
  id: number
  title: string
  content: string
  category: string
  tags: string[]
  source: 'manual' | 'alert' | 'import'
  priority: 'low' | 'medium' | 'high'
  status: 'active' | 'inactive' | 'archived'
  created_at: string
  updated_at: string
  created_by: string
  updated_by: string
  view_count: number
  like_count: number
}

// 创建知识库请求参数
export interface CreateKnowledgeRequest {
  title: string
  content: string
  category: string
  tags?: string[]
  priority?: 'low' | 'medium' | 'high'
  status?: 'active' | 'inactive'
}

// 更新知识库请求参数
export interface UpdateKnowledgeRequest {
  title?: string
  content?: string
  category?: string
  tags?: string[]
  priority?: 'low' | 'medium' | 'high'
  status?: 'active' | 'inactive' | 'archived'
}

// 知识库查询参数
export interface KnowledgeQueryParams {
  page?: number
  pageSize?: number
  category?: string
  tags?: string
  priority?: string
  status?: string
  search?: string
  sortBy?: 'created_at' | 'updated_at' | 'title' | 'view_count' | 'like_count'
  sortOrder?: 'asc' | 'desc'
}

/**
 * 知识库服务类
 */
export class KnowledgeService {
  private static readonly BASE_URL = '/api/v1/knowledge'

  /**
   * 获取知识库列表
   */
  static async getKnowledgeList(params?: KnowledgeQueryParams): Promise<ApiResponse<PaginatedResponse<Knowledge>>> {
    return ApiService.get(this.BASE_URL, params)
  }

  /**
   * 获取知识库详情
   */
  static async getKnowledge(id: number): Promise<ApiResponse<Knowledge>> {
    return ApiService.get(`${this.BASE_URL}/${id}`)
  }

  /**
   * 创建知识库
   */
  static async createKnowledge(data: CreateKnowledgeRequest): Promise<ApiResponse<Knowledge>> {
    return ApiService.post(this.BASE_URL, data)
  }

  /**
   * 更新知识库
   */
  static async updateKnowledge(id: number, data: UpdateKnowledgeRequest): Promise<ApiResponse<Knowledge>> {
    return ApiService.put(`${this.BASE_URL}/${id}`, data)
  }

  /**
   * 删除知识库
   */
  static async deleteKnowledge(id: number): Promise<ApiResponse<void>> {
    return ApiService.delete(`${this.BASE_URL}/${id}`)
  }

  /**
   * 批量删除知识库
   */
  static async batchDeleteKnowledge(ids: number[]): Promise<ApiResponse<void>> {
    return ApiService.post(`${this.BASE_URL}/batch-delete`, { ids })
  }

  /**
   * 搜索知识库
   */
  static async searchKnowledge(query: string, params?: Omit<KnowledgeQueryParams, 'search'>): Promise<ApiResponse<PaginatedResponse<Knowledge>>> {
    return ApiService.get(`${this.BASE_URL}/search`, { ...params, search: query })
  }

  /**
   * 获取知识库分类列表
   */
  static async getCategories(): Promise<ApiResponse<string[]>> {
    return ApiService.get(`${this.BASE_URL}/categories`)
  }

  /**
   * 获取知识库标签列表
   */
  static async getTags(): Promise<ApiResponse<string[]>> {
    return ApiService.get(`${this.BASE_URL}/tags`)
  }

  /**
   * 点赞知识库
   */
  static async likeKnowledge(id: number): Promise<ApiResponse<void>> {
    return ApiService.post(`${this.BASE_URL}/${id}/like`)
  }

  /**
   * 取消点赞知识库
   */
  static async unlikeKnowledge(id: number): Promise<ApiResponse<void>> {
    return ApiService.delete(`${this.BASE_URL}/${id}/like`)
  }

  /**
   * 导出知识库
   */
  static async exportKnowledge(ids?: number[]): Promise<ApiResponse<{ download_url: string }>> {
    const data = ids ? { ids } : {}
    return ApiService.post(`${this.BASE_URL}/export`, data)
  }

  /**
   * 导入知识库
   */
  static async importKnowledge(file: File): Promise<ApiResponse<{ imported_count: number; failed_count: number }>> {
    const formData = new FormData()
    formData.append('file', file)
    
    return ApiService.post(`${this.BASE_URL}/import`, formData)
  }
}

export default KnowledgeService