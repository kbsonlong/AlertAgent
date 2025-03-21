// 告警相关API服务

/**
 * 获取告警列表
 */
export const getAlerts = async () => {
  const response = await fetch('/api/v1/alerts');
  return await response.json();
};

/**
 * 更新告警状态
 * @param id 告警ID
 * @param status 状态：acknowledged(已确认) 或 resolved(已解决)
 */
export const updateAlertStatus = async (id: number, status: string) => {
  const response = await fetch(`/api/v1/alerts/${id}`, {
    method: 'PUT',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({ status }),
  });
  return await response.json();
};

/**
 * 同步分析告警
 * @param id 告警ID
 */
export const analyzeAlert = async (id: number) => {
  const response = await fetch(`/api/v1/alerts/${id}/analyze`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
  });
  return await response.json();
};

/**
 * 异步分析告警
 * @param id 告警ID
 */
export const asyncAnalyzeAlert = async (id: number) => {
  const response = await fetch(`/api/v1/alerts/${id}/async-analyze`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
  });
  return await response.json();
};

/**
 * 获取分析状态
 * @param id 告警ID
 */
export const getAnalysisStatus = async (id: number) => {
  const response = await fetch(`/api/v1/alerts/${id}/analysis-status`, {
    method: 'GET',
    headers: {
      'Accept': 'application/json',
    },
  });
  return await response.json();
};