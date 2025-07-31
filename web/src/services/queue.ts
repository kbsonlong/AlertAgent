import { request } from '@/utils/request'
import type { 
  QueueMetrics, 
  TaskMetrics, 
  Task, 
  TaskFilter, 
  TaskListResponse,
  BatchTaskRequest,
  BatchTaskResponse,
  PerformanceStats,
  QueueHealth
} from '@/types/queue'

export const queueService = {
  // 获取队列指标
  async getQueueMetrics(queueName: string) {
    return request<QueueMetrics>({
      url: `/api/v1/queues/${queueName}/metrics`,
      method: 'GET'
    })
  },

  // 获取所有队列指标
  async getAllQueueMetrics() {
    return request<Record<string, QueueMetrics>>({
      url: '/api/v1/queues/metrics',
      method: 'GET'
    })
  },

  // 获取任务类型指标
  async getTaskMetrics(taskType: string) {
    return request<TaskMetrics>({
      url: `/api/v1/queues/tasks/${taskType}/metrics`,
      method: 'GET'
    })
  },

  // 获取任务状态
  async getTaskStatus(taskId: string) {
    return request<Task>({
      url: `/api/v1/queues/tasks/${taskId}`,
      method: 'GET'
    })
  },

  // 获取任务列表
  async getTasks(filter: TaskFilter) {
    const params = new URLSearchParams()
    
    if (filter.queueName) params.append('queue_name', filter.queueName)
    if (filter.status) params.append('status', filter.status)
    if (filter.taskType) params.append('task_type', filter.taskType)
    if (filter.page) params.append('page', filter.page.toString())
    if (filter.pageSize) params.append('page_size', filter.pageSize.toString())
    if (filter.startTime) params.append('start_time', filter.startTime.toISOString())
    if (filter.endTime) params.append('end_time', filter.endTime.toISOString())

    return request<TaskListResponse>({
      url: `/api/v1/queues/tasks?${params.toString()}`,
      method: 'GET'
    })
  },

  // 重试任务
  async retryTask(taskId: string) {
    return request({
      url: `/api/v1/queues/tasks/${taskId}/retry`,
      method: 'POST'
    })
  },

  // 跳过任务
  async skipTask(taskId: string) {
    return request({
      url: `/api/v1/queues/tasks/${taskId}/skip`,
      method: 'POST'
    })
  },

  // 批量重试任务
  async batchRetryTasks(taskIds: string[]) {
    return request<BatchTaskResponse>({
      url: '/api/v1/queues/tasks/batch/retry',
      method: 'POST',
      data: { task_ids: taskIds } as BatchTaskRequest
    })
  },

  // 批量跳过任务
  async batchSkipTasks(taskIds: string[]) {
    return request<BatchTaskResponse>({
      url: '/api/v1/queues/tasks/batch/skip',
      method: 'POST',
      data: { task_ids: taskIds } as BatchTaskRequest
    })
  },

  // 取消任务
  async cancelTask(taskId: string) {
    return request({
      url: `/api/v1/queues/tasks/${taskId}/cancel`,
      method: 'POST'
    })
  },

  // 获取任务执行日志
  async getTaskLogs(taskId: string, page: number = 1, pageSize: number = 50) {
    return request({
      url: `/api/v1/queues/tasks/${taskId}/logs?page=${page}&page_size=${pageSize}`,
      method: 'GET'
    })
  },

  // 获取任务执行历史
  async getTaskHistory(taskId: string) {
    return request({
      url: `/api/v1/queues/tasks/${taskId}/history`,
      method: 'GET'
    })
  },

  // 获取队列健康状态
  async getQueueHealth() {
    return request<QueueHealth>({
      url: '/api/v1/queues/health',
      method: 'GET'
    })
  },

  // 清理过期任务
  async cleanupExpiredTasks(queueName: string, maxAge: string = '1h') {
    return request({
      url: `/api/v1/queues/${queueName}/cleanup?max_age=${maxAge}`,
      method: 'POST'
    })
  },

  // 队列优化
  async optimizeQueue(queueName: string, options: Record<string, any>) {
    return request({
      url: `/api/v1/queues/${queueName}/optimize`,
      method: 'POST',
      data: options
    })
  },

  // 获取队列性能统计
  async getQueuePerformanceStats(queueName: string, duration: string = '24h') {
    return request<PerformanceStats>({
      url: `/api/v1/queues/${queueName}/performance?duration=${duration}`,
      method: 'GET'
    })
  },

  // 获取Worker统计
  async getWorkerStats(queueName?: string) {
    const url = queueName 
      ? `/api/v1/queues/${queueName}/workers`
      : '/api/v1/queues/workers'
    
    return request({
      url,
      method: 'GET'
    })
  },

  // 获取队列配置
  async getQueueConfig(queueName: string) {
    return request({
      url: `/api/v1/queues/${queueName}/config`,
      method: 'GET'
    })
  },

  // 更新队列配置
  async updateQueueConfig(queueName: string, config: Record<string, any>) {
    return request({
      url: `/api/v1/queues/${queueName}/config`,
      method: 'PUT',
      data: config
    })
  },

  // 暂停队列
  async pauseQueue(queueName: string) {
    return request({
      url: `/api/v1/queues/${queueName}/pause`,
      method: 'POST'
    })
  },

  // 恢复队列
  async resumeQueue(queueName: string) {
    return request({
      url: `/api/v1/queues/${queueName}/resume`,
      method: 'POST'
    })
  },

  // 清空队列
  async clearQueue(queueName: string) {
    return request({
      url: `/api/v1/queues/${queueName}/clear`,
      method: 'POST'
    })
  },

  // 获取队列统计历史
  async getQueueStatsHistory(queueName: string, period: string = '24h') {
    return request({
      url: `/api/v1/queues/${queueName}/stats/history?period=${period}`,
      method: 'GET'
    })
  },

  // 导出任务数据
  async exportTasks(filter: TaskFilter, format: 'csv' | 'json' = 'csv') {
    const params = new URLSearchParams()
    
    if (filter.queueName) params.append('queue_name', filter.queueName)
    if (filter.status) params.append('status', filter.status)
    if (filter.taskType) params.append('task_type', filter.taskType)
    if (filter.startTime) params.append('start_time', filter.startTime.toISOString())
    if (filter.endTime) params.append('end_time', filter.endTime.toISOString())
    params.append('format', format)

    return request({
      url: `/api/v1/queues/tasks/export?${params.toString()}`,
      method: 'GET',
      responseType: 'blob'
    })
  },

  // 获取任务执行日志
  async getTaskLogs(taskId: string) {
    return request({
      url: `/api/v1/queues/tasks/${taskId}/logs`,
      method: 'GET'
    })
  },

  // 取消任务
  async cancelTask(taskId: string) {
    return request({
      url: `/api/v1/queues/tasks/${taskId}/cancel`,
      method: 'POST'
    })
  },

  // 获取队列拓扑
  async getQueueTopology() {
    return request({
      url: '/api/v1/queues/topology',
      method: 'GET'
    })
  },

  // 队列优化
  async optimizeQueue(queueName: string, options: QueueOptimizeOptions) {
    return request({
      url: `/api/v1/queues/${queueName}/optimize`,
      method: 'POST',
      data: options
    })
  },

  // 获取队列优化建议
  async getQueueRecommendations(queueName: string) {
    return request({
      url: `/api/v1/queues/${queueName}/recommendations`,
      method: 'GET'
    })
  },

  // 队列扩缩容
  async scaleQueue(queueName: string, targetWorkers: number, reason?: string) {
    return request({
      url: `/api/v1/queues/${queueName}/scale`,
      method: 'POST',
      data: {
        target_workers: targetWorkers,
        scale_reason: reason
      }
    })
  },

  // 获取队列告警
  async getQueueAlerts(queueName?: string, status?: string, page: number = 1, pageSize: number = 20) {
    const params = new URLSearchParams()
    if (queueName) params.append('queue_name', queueName)
    if (status) params.append('status', status)
    params.append('page', page.toString())
    params.append('page_size', pageSize.toString())

    return request({
      url: `/api/v1/queues/alerts?${params.toString()}`,
      method: 'GET'
    })
  },

  // 确认告警
  async acknowledgeAlert(alertId: string) {
    return request({
      url: `/api/v1/queues/alerts/${alertId}/acknowledge`,
      method: 'POST'
    })
  },

  // 获取实时队列状态（WebSocket连接辅助方法）
  createQueueStatusWebSocket(onMessage: (data: any) => void, onError?: (error: Event) => void) {
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
    const wsUrl = `${protocol}//${window.location.host}/api/v1/queues/ws`
    
    const ws = new WebSocket(wsUrl)
    
    ws.onmessage = (event) => {
      try {
        const data = JSON.parse(event.data)
        onMessage(data)
      } catch (error) {
        console.error('Failed to parse WebSocket message:', error)
      }
    }
    
    ws.onerror = (error) => {
      console.error('WebSocket error:', error)
      if (onError) {
        onError(error)
      }
    }
    
    ws.onclose = () => {
      console.log('WebSocket connection closed')
    }
    
    return ws
  }
}