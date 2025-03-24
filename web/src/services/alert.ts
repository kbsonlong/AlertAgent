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

// 告警相关API服务

/**
 * 获取告警列表
 */
export const getAlerts = async () => {
  return request<Alert[]>('/api/v1/alerts', {
    method: 'GET'
  });
};

/**
 * 更新告警状态
 * @param id 告警ID
 * @param status 状态：acknowledged(已确认) 或 resolved(已解决)
 */
export const updateAlertStatus = async (id: number, status: 'acknowledged' | 'resolved') => {
  return request<Alert>(`/api/v1/alerts/${id}`, {
    method: 'PUT',
    data: { status }
  });
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
  return request<{ taskId: string }>(`/api/v1/alerts/${id}/async-analyze`, {
    method: 'POST'
  });
};

/**
 * 获取分析状态
 * @param id 告警ID
 */
export const getAnalysisStatus = async (id: number) => {
  return request<AlertAnalysis>(`/api/v1/alerts/${id}/analysis-status`, {
    method: 'GET'
  });
};