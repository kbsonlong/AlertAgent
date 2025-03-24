import { request } from "../utils/requests";

export interface Knowledge {
  id: number;
  title: string;
  content: string;
  category: string;
  tags?: string[];
  source: string;
  sourceId: number;
  summary: string;
  createdAt: string;
  updatedAt: string;
}

export interface KnowledgeListParams {
  category?: string;
  tags?: string[];
  keyword?: string;
  page?: number;
  pageSize?: number;
}

export async function getKnowledgeList(params: KnowledgeListParams): Promise<Knowledge[]> {
  return request<Knowledge[]>('/api/v1/knowledge', {
    method: 'GET',
    params,
  });
}

export async function getKnowledgeById(id: number): Promise<Knowledge> {
  return request<Knowledge>(`/api/v1/knowledge/${id}`, {
    method: 'GET',
  });
}

export async function findSimilarKnowledge(content: string, limit: number = 5): Promise<Knowledge[]> {
  return request<Knowledge[]>('/api/v1/knowledge/similar', {
    method: 'POST',
    data: {
      content,
      limit,
    },
  });
}