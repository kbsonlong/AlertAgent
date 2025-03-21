// 通知相关API服务

/**
 * 获取通知组列表
 */
export const getGroups = async () => {
  const response = await fetch('/api/v1/groups');
  return await response.json();
};

/**
 * 创建通知组
 * @param data 通知组数据
 */
export const createGroup = async (data: any) => {
  const response = await fetch('/api/v1/groups', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify(data),
  });
  return await response.json();
};

/**
 * 更新通知组
 * @param id 通知组ID
 * @param data 通知组数据
 */
export const updateGroup = async (id: number, data: any) => {
  const response = await fetch(`/api/v1/groups/${id}`, {
    method: 'PUT',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify(data),
  });
  return await response.json();
};

/**
 * 删除通知组
 * @param id 通知组ID
 */
export const deleteGroup = async (id: number) => {
  const response = await fetch(`/api/v1/groups/${id}`, {
    method: 'DELETE',
  });
  return await response.json();
};

/**
 * 获取通知模板列表
 */
export const getTemplates = async () => {
  const response = await fetch('/api/v1/templates');
  return await response.json();
};

/**
 * 创建通知模板
 * @param data 模板数据
 */
export const createTemplate = async (data: any) => {
  const response = await fetch('/api/v1/templates', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify(data),
  });
  return await response.json();
};

/**
 * 更新通知模板
 * @param id 模板ID
 * @param data 模板数据
 */
export const updateTemplate = async (id: number, data: any) => {
  const response = await fetch(`/api/v1/templates/${id}`, {
    method: 'PUT',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify(data),
  });
  return await response.json();
};

/**
 * 删除通知模板
 * @param id 模板ID
 */
export const deleteTemplate = async (id: number) => {
  const response = await fetch(`/api/v1/templates/${id}`, {
    method: 'DELETE',
  });
  return await response.json();
};