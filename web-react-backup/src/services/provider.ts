import { request } from '../utils/requests';

export interface Provider {
  id: number;
  name: string;
  type: string;
  status: string;
  description?: string;
  endpoint: string;
  auth_type?: string;
  auth_config?: string;
  labels?: string;
  last_check?: string;
  last_error?: string;
}

export interface ProviderListParams {
  page?: number;
  pageSize?: number;
  keyword?: string;
  type?: string;
  status?: string;
}

export interface CreateProviderParams {
  name: string;
  type: string;
  description?: string;
  endpoint: string;
  auth_type?: string;
  auth_config?: string;
  labels?: string;
}

export interface UpdateProviderParams extends CreateProviderParams {
  id: number;
  status?: string;
}

export interface TestProviderParams {
  type: string;
  endpoint: string;
  auth_type?: string;
  auth_config?: string;
}

export interface TestProviderResult {
  status: string;
  message: string;
}

/**
 * 获取数据源列表
 */
export async function getProviderList(params?: ProviderListParams) {
  return request('/api/v1/providers', {
    method: 'GET',
    params,
  });
}

/**
 * 获取单个数据源
 */
export async function getProviderById(id: number) {
  return request(`/api/v1/providers/${id}`, {
    method: 'GET',
  });
}

/**
 * 创建数据源
 */
export async function createProvider(data: CreateProviderParams) {
  return request('/api/v1/providers', {
    method: 'POST',
    data,
  });
}

/**
 * 更新数据源
 */
export async function updateProvider(id: number, data: UpdateProviderParams) {
  return request(`/api/v1/providers/${id}`, {
    method: 'PUT',
    data,
  });
}

/**
 * 删除数据源
 */
export async function deleteProvider(id: number) {
  return request(`/api/v1/providers/${id}`, {
    method: 'DELETE',
  });
}

/**
 * 测试数据源连接
 */
export async function testProvider(data: TestProviderParams) {
  return request('/api/v1/providers/test', {
    method: 'POST',
    data,
  });
}

/**
 * 数据源类型选项
 */
export const PROVIDER_TYPES = [
  { label: 'Prometheus', value: 'prometheus' },
  { label: 'VictoriaMetrics', value: 'victoriametrics' },
];

/**
 * 数据源状态选项
 */
export const PROVIDER_STATUS = [
  { label: '活跃', value: 'active' },
  { label: '非活跃', value: 'inactive' },
];

/**
 * 认证类型选项
 */
export const AUTH_TYPES = [
  { label: '无认证', value: 'none' },
  { label: 'Basic Auth', value: 'basic' },
  { label: 'Bearer Token', value: 'bearer' },
  { label: 'API Key', value: 'apikey' },
];