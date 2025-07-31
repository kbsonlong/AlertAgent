import { request } from "../utils/requests";

export interface Knowledge {
  id: number;
  title: string;
  content: string;
  source: string;
  source_id: number;
  created_at: string;
  updated_at: string;
}

export interface KnowledgeListParams {
  category?: string;
  tags?: string[];
  keyword?: string;
  page?: number;
  pageSize?: number;
}

// API响应类型
export interface ApiResponse<T> {
  code: number;
  msg?: string;
  data: T;
}

export async function getKnowledgeList(params: KnowledgeListParams) {
  return request('/api/v1/knowledge', {
    method: 'GET',
    params,
  }) as Promise<ApiResponse<{
    list: Knowledge[];
    total: number;
    page: number;
    pageSize: number;
  }>>;
}

export async function getKnowledgeById(id: number) {
  return request(`/api/v1/knowledge/${id}`, {
    method: 'GET',
  });
}

export async function findSimilarKnowledge(content: string, limit: number = 5) {
  return request('/api/v1/knowledge/similar', {
    method: 'POST',
    data: {
      content,
      limit,
    },
  });
}