import { get, post, put, del } from '@/utils/request'
import type { Knowledge, CreateKnowledgeRequest, UpdateKnowledgeRequest, PaginatedResponse } from '@/types'

/**
 * 获取知识库列表
 * @param params 查询参数
 */
export const getKnowledgeList = async (params?: {
  page?: number
  page_size?: number
  category?: string
  search?: string
  tags?: string
}) => {
  return get<PaginatedResponse<Knowledge>>('/api/v1/knowledge', params)
}

/**
 * 获取单个知识库条目
 * @param id 知识库ID
 */
export const getKnowledge = async (id: number) => {
  return get<Knowledge>(`/api/v1/knowledge/${id}`)
}

/**
 * 创建知识库条目
 * @param data 知识库数据
 */
export const createKnowledge = async (data: CreateKnowledgeRequest) => {
  return post<Knowledge>('/api/v1/knowledge', data)
}

/**
 * 更新知识库条目
 * @param id 知识库ID
 * @param data 更新数据
 */
export const updateKnowledge = async (id: number, data: UpdateKnowledgeRequest) => {
  return put<Knowledge>(`/api/v1/knowledge/${id}`, data)
}

/**
 * 删除知识库条目
 * @param id 知识库ID
 */
export const deleteKnowledge = async (id: number) => {
  return del(`/api/v1/knowledge/${id}`)
}

/**
 * 搜索相关知识
 * @param query 搜索关键词
 * @param limit 返回数量限制
 */
export const searchKnowledge = async (query: string, limit = 10) => {
  return get<Knowledge[]>('/api/v1/knowledge/search', {
    query,
    limit
  })
}

/**
 * 获取知识库分类列表
 */
export const getKnowledgeCategories = async () => {
  return get<Array<{
    category: string
    count: number
  }>>('/api/v1/knowledge/categories')
}

/**
 * 获取知识库标签列表
 */
export const getKnowledgeTags = async () => {
  return get<Array<{
    tag: string
    count: number
  }>>('/api/v1/knowledge/tags')
}

/**
 * 批量删除知识库条目
 * @param ids 知识库ID数组
 */
export const batchDeleteKnowledge = async (ids: number[]) => {
  return del('/api/v1/knowledge/batch', {
    data: { ids }
  })
}

/**
 * 导入知识库
 * @param file 知识库文件
 */
export const importKnowledge = async (file: File) => {
  const formData = new FormData()
  formData.append('file', file)
  
  return post<{
    imported_count: number
    failed_count: number
    errors?: string[]
  }>('/api/v1/knowledge/import', formData, {
    headers: {
      'Content-Type': 'multipart/form-data'
    }
  })
}

/**
 * 导出知识库
 * @param ids 知识库ID数组，为空则导出所有
 */
export const exportKnowledge = async (ids?: number[]) => {
  return get('/api/v1/knowledge/export', { ids }, {
    responseType: 'blob'
  })
}

/**
 * 重建知识库向量索引
 */
export const rebuildKnowledgeIndex = async () => {
  return post('/api/v1/knowledge/rebuild-index')
}