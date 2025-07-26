import { request } from '../utils/requests';

// 系统配置接口
export interface SystemConfig {
  ollama_enabled: boolean;
}

// API响应类型
export interface ApiResponse<T> {
  code: number;
  msg?: string;
  data: T;
}

/**
 * 获取系统配置
 */
export const getSystemConfig = async (): Promise<SystemConfig> => {
  const response = await request('/api/v1/config', {
    method: 'GET'
  }) as ApiResponse<SystemConfig>;
  return response.data;
};