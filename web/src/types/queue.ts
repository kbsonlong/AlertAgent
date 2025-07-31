// 队列指标
export interface QueueMetrics {
  queue_name: string
  pending_count: number
  processing_count: number
  completed_count: number
  failed_count: number
  dead_letter_count: number
  delayed_count: number
  throughput_per_min: number
  avg_processing_time: number // 毫秒
  error_rate: number // 百分比
  last_updated: string
}

// 任务指标
export interface TaskMetrics {
  task_type: string
  total_count: number
  completed_count: number
  failed_count: number
  avg_processing_time: number // 毫秒
  success_rate: number // 百分比
  last_hour_count: number
}

// 任务状态
export type TaskStatus = 'pending' | 'processing' | 'completed' | 'failed' | 'retrying'

// 任务类型
export type TaskType = 'ai_analysis' | 'notification' | 'config_sync' | 'rule_update' | 'health_check'

// 任务优先级
export type TaskPriority = 0 | 1 | 2 | 3 // 低、普通、高、紧急

// 任务
export interface Task {
  id: string
  type: TaskType
  payload: Record<string, any>
  priority: TaskPriority
  status: TaskStatus
  retry: number
  max_retry: number
  created_at: string
  updated_at: string
  scheduled_at: string
  started_at?: string
  completed_at?: string
  error_msg?: string
  worker_id?: string
}

// 任务过滤器
export interface TaskFilter {
  queueName?: string
  status?: string
  taskType?: string
  page?: number
  pageSize?: number
  startTime?: Date
  endTime?: Date
}

// 任务列表响应
export interface TaskListResponse {
  tasks: Task[]
  total: number
  page: number
  page_size: number
}

// 批量任务请求
export interface BatchTaskRequest {
  task_ids: string[]
}

// 任务操作结果
export interface TaskOperationResult {
  task_id: string
  success: boolean
  error?: string
}

// 批量任务响应
export interface BatchTaskResponse {
  total: number
  succeeded: number
  failed: number
  results: TaskOperationResult[]
}

// 性能数据点
export interface ThroughputPoint {
  timestamp: string
  value: number
}

export interface ErrorRatePoint {
  timestamp: string
  value: number
}

export interface LatencyPoint {
  timestamp: string
  value: number // 毫秒
}

// Worker性能统计
export interface WorkerPerformanceStats {
  worker_id: string
  tasks_handled: number
  success_rate: number
  avg_latency: number // 毫秒
  last_active: string
  status: 'active' | 'idle' | 'inactive'
}

// 性能优化建议
export interface PerformanceRecommendation {
  type: string
  priority: 'low' | 'medium' | 'high' | 'critical'
  title: string
  description: string
  action: string
}

// 队列性能统计
export interface PerformanceStats {
  queue_name: string
  throughput_history: ThroughputPoint[]
  error_rate_history: ErrorRatePoint[]
  latency_history: LatencyPoint[]
  worker_stats: WorkerPerformanceStats[]
  recommendations: PerformanceRecommendation[]
}

// 队列健康状态
export interface QueueHealth {
  status: 'healthy' | 'degraded' | 'critical'
  timestamp: string
  queues: Record<string, QueueHealthDetail>
}

export interface QueueHealthDetail {
  status: 'healthy' | 'warning' | 'critical'
  pending_count: number
  processing_count: number
  error_rate: number
  throughput: number
  warning?: string
  error?: string
}

// 队列配置
export interface QueueConfig {
  name: string
  max_workers: number
  max_retry: number
  retry_delay: string // 例如: "1m", "30s"
  task_timeout: string
  priority_enabled: boolean
  dead_letter_enabled: boolean
  metrics_enabled: boolean
  auto_scale: {
    enabled: boolean
    min_workers: number
    max_workers: number
    scale_up_threshold: number
    scale_down_threshold: number
  }
}

// Worker信息
export interface WorkerInfo {
  id: string
  queue_name: string
  status: 'active' | 'idle' | 'inactive' | 'error'
  current_task?: string
  tasks_processed: number
  success_rate: number
  avg_processing_time: number
  last_heartbeat: string
  started_at: string
  version: string
  host: string
  pid: number
}

// 队列拓扑
export interface QueueTopology {
  queues: QueueNode[]
  workers: WorkerNode[]
  connections: Connection[]
}

export interface QueueNode {
  id: string
  name: string
  type: 'queue'
  status: 'active' | 'paused' | 'error'
  metrics: QueueMetrics
}

export interface WorkerNode {
  id: string
  name: string
  type: 'worker'
  status: 'active' | 'idle' | 'inactive'
  queue_name: string
  info: WorkerInfo
}

export interface Connection {
  from: string
  to: string
  type: 'consumes' | 'produces'
  status: 'active' | 'inactive'
}

// 任务执行日志
export interface TaskLog {
  id: string
  task_id: string
  level: 'debug' | 'info' | 'warn' | 'error'
  message: string
  timestamp: string
  worker_id?: string
  context?: Record<string, any>
}

// 队列统计历史
export interface QueueStatsHistory {
  queue_name: string
  period: string
  data_points: QueueStatsPoint[]
}

export interface QueueStatsPoint {
  timestamp: string
  pending_count: number
  processing_count: number
  completed_count: number
  failed_count: number
  throughput: number
  error_rate: number
  avg_latency: number
}

// WebSocket消息类型
export interface QueueWebSocketMessage {
  type: 'queue_metrics' | 'task_update' | 'worker_status' | 'alert'
  data: any
  timestamp: string
}

// 队列告警
export interface QueueAlert {
  id: string
  queue_name: string
  type: 'high_error_rate' | 'queue_backlog' | 'worker_down' | 'task_timeout'
  severity: 'warning' | 'critical'
  message: string
  details: Record<string, any>
  created_at: string
  resolved_at?: string
  status: 'active' | 'resolved' | 'acknowledged'
}

// 队列操作历史
export interface QueueOperation {
  id: string
  queue_name: string
  operation: 'pause' | 'resume' | 'clear' | 'optimize' | 'scale'
  operator: string
  parameters?: Record<string, any>
  result: 'success' | 'failed'
  error_message?: string
  created_at: string
  duration: number // 毫秒
}

// 导出选项
export interface ExportOptions {
  format: 'csv' | 'json' | 'xlsx'
  fields: string[]
  filter: TaskFilter
  include_payload: boolean
  include_logs: boolean
}

// 队列监控配置
export interface MonitorConfig {
  refresh_interval: number // 秒
  alert_thresholds: {
    error_rate: number
    queue_backlog: number
    processing_time: number
  }
  auto_cleanup: {
    enabled: boolean
    max_age: string
    batch_size: number
  }
  notifications: {
    enabled: boolean
    channels: string[]
    conditions: string[]
  }
}

// 队列优化选项
export interface QueueOptimizeOptions {
  auto_scale: boolean
  cleanup_expired: boolean
  rebalance: boolean
  max_age?: string
  optimize_workers: boolean
}

// 队列优化响应
export interface QueueOptimizeResponse {
  queue_name: string
  operations: OptimizationOperation[]
  summary: OptimizationSummary
  duration: number // 毫秒
  completed_at: string
}

// 优化操作
export interface OptimizationOperation {
  type: string
  status: 'success' | 'failed' | 'skipped'
  message: string
  details?: Record<string, any>
  duration: number // 毫秒
  error?: string
}

// 优化摘要
export interface OptimizationSummary {
  total_operations: number
  successful_operations: number
  failed_operations: number
  skipped_operations: number
  success_rate: number
  improvement_score: number
}

// 队列优化建议
export interface QueueRecommendation {
  id: string
  type: string
  priority: 'low' | 'medium' | 'high' | 'critical'
  title: string
  description: string
  action: string
  impact: string
  metrics: Record<string, any>
  auto_fix: boolean
  created_at: string
}

// 队列扩缩容请求
export interface QueueScaleRequest {
  target_workers: number
  scale_reason?: string
  force?: boolean
}

// 队列扩缩容响应
export interface QueueScaleResponse {
  queue_name: string
  current_workers: number
  target_workers: number
  scaled_workers: number
  status: string
  message: string
  estimated_time: number // 秒
  completed_at: string
}

// 队列告警
export interface QueueAlert {
  id: string
  queue_name: string
  type: string
  severity: 'info' | 'warning' | 'critical'
  title: string
  message: string
  details: Record<string, any>
  status: 'active' | 'resolved' | 'acknowledged'
  created_at: string
  updated_at: string
  resolved_at?: string
  acknowledged_at?: string
  acknowledged_by?: string
}

// 队列告警响应
export interface QueueAlertsResponse {
  alerts: QueueAlert[]
  total: number
  page: number
  page_size: number
}