import { request } from '../utils/requests';

// 类型定义
export interface Alert {
  id: number;
  title: string;
  content: string;
  status: 'new' | 'acknowledged' | 'resolved';
  severity: 'critical' | 'high' | 'medium' | 'low';
  source: string;
  createdAt: string;
  updatedAt: string;
  analysis?: string;
}

export interface AlertAnalysis {
  id: number;
  alertId: number;
  status: 'pending' | 'processing' | 'completed' | 'failed';
  result?: string;
  error?: string;
  createdAt: string;
  updatedAt: string;
}

export interface AnalysisTask {
  taskId: string;
}

// 告警相关API服务

/**
 * 获取告警列表
 */
export const getAlerts = async () => {
  const response = await request<Alert[]>('/api/v1/alerts', {
    method: 'GET'
  });
  return response.data;
};

/**
 * 更新告警状态
 * @param id 告警ID
 * @param status 状态：acknowledged(已确认) 或 resolved(已解决)
 */
export const updateAlertStatus = async (id: number, status: 'acknowledged' | 'resolved') => {
  const response = await request<Alert>(`/api/v1/alerts/${id}`, {
    method: 'PUT',
    data: { status }
  });
  return response.data;
};

/**
 * 同步分析告警
 * @param id 告警ID
 */
export const analyzeAlert = async (id: number) => {
  return request<AlertAnalysis>(`/api/v1/alerts/${id}/analyze`, {
    method: 'POST'
  });
};

/**
 * 异步分析告警
 * @param id 告警ID
 */
export const asyncAnalyzeAlert = async (id: number) => {
  const response = await request<{ task_id: number; submit_time: string }>(`/api/v1/alerts/${id}/async/analyze`, {
    method: 'POST'
  });
  return response.data;
};

/**
 * 获取分析结果
 * @param taskId 任务ID
 */
export const getAnalysisResult = async (taskId: number) => {
  const response = await request<{ status: string; result?: string; error?: string; message?: string }>(`/api/v1/alerts/async/result/${taskId}`, {
    method: 'GET'
  });
  return response.data;
};

/**
 * 获取分析状态（兼容旧接口）
 * @param id 告警ID
 */
export const getAnalysisStatus = async (id: number) => {
  return getAnalysisResult(id);
};

/**
 * 转换为知识库
 * @param id 告警ID
 */
export const convertToKnowledge = async (id: number) => {
  const response = await request<{ id: number }>(`/api/v1/alerts/${id}/convert-to-knowledge`, {
    method: 'POST'
  });
  return response.data;
};