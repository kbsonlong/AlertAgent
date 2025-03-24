import { request } from '../utils/requests';

// 类型定义
export interface NotificationGroup {
  id?: number;
  name: string;
  description?: string;
  members: string[];
  channels: string[];
}

export interface NotificationTemplate {
  id?: number;
  name: string;
  description?: string;
  content: string;
  type: 'email' | 'sms' | 'webhook';
  variables?: string[];
}

// 通知相关API服务

/**
 * 获取通知组列表
 */
export const getGroups = async () => {
  return request<NotificationGroup[]>('/api/v1/groups', {
    method: 'GET'
  });
};

/**
 * 创建通知组
 * @param data 通知组数据
 */
export const createGroup = async (data: NotificationGroup) => {
  return request<NotificationGroup>('/api/v1/groups', {
    method: 'POST',
    data
  });
};

/**
 * 更新通知组
 * @param id 通知组ID
 * @param data 通知组数据
 */
export const updateGroup = async (id: number, data: NotificationGroup) => {
  return request<NotificationGroup>(`/api/v1/groups/${id}`, {
    method: 'PUT',
    data
  });
};

/**
 * 删除通知组
 * @param id 通知组ID
 */
export const deleteGroup = async (id: number) => {
  return request<void>(`/api/v1/groups/${id}`, {
    method: 'DELETE'
  });
};

/**
 * 获取通知模板列表
 */
export const getTemplates = async () => {
  return request<NotificationTemplate[]>('/api/v1/templates', {
    method: 'GET'
  });
};

/**
 * 创建通知模板
 * @param data 模板数据
 */
export const createTemplate = async (data: NotificationTemplate) => {
  return request<NotificationTemplate>('/api/v1/templates', {
    method: 'POST',
    data
  });
};

/**
 * 更新通知模板
 * @param id 模板ID
 * @param data 模板数据
 */
export const updateTemplate = async (id: number, data: NotificationTemplate) => {
  return request<NotificationTemplate>(`/api/v1/templates/${id}`, {
    method: 'PUT',
    data
  });
};

/**
 * 删除通知模板
 * @param id 模板ID
 */
export const deleteTemplate = async (id: number) => {
  return request<void>(`/api/v1/templates/${id}`, {
    method: 'DELETE'
  });
};