// API响应基础类型
export interface ApiResponse<T = any> {
  code: number
  message: string
  data: T
}

// 分页响应类型
export interface PaginatedResponse<T> {
  items: T[]
  total: number
  page: number
  page_size: number
}

// 用户列表响应类型（匹配后端返回结构）
export interface UserListResponse {
  users: User[]
  total: number
  page: number
  size: number
}

// 告警相关类型
export interface Alert {
  id: number | string
  created_at: string
  updated_at: string
  name: string
  title: string
  level: string
  status: 'firing' | 'acknowledged' | 'resolved'
  source: string
  content: string
  description?: string
  labels: string | Record<string, string> // 支持字符串和对象两种格式
  annotations?: Record<string, string>
  metrics?: {
    current: number
    threshold: number
    unit: string
    status: string
  }
  history?: Array<{
    timestamp: string
    status: string
    value?: number
    note?: string
  }>
  rule_id?: number
  template_id?: number
  group_id?: number
  handler?: string
  handle_time?: string
  handle_note?: string
  analysis?: string
  notify_time?: string
  notify_count: number
  severity: string
}

// 告警分析结果
export interface AlertAnalysis {
  analysis: string
  similar_alerts?: SimilarAlert[]
  knowledge_references?: Knowledge[]
}

// 相似告警
export interface SimilarAlert {
  alert: Alert
  similarity: number
}

// 规则相关类型
export interface Rule {
  id: number
  created_at: string
  updated_at: string
  name: string
  expression: string
  duration: string
  severity: string
  labels: Record<string, string>
  annotations: Record<string, string>
  targets: string[]
  enabled: boolean
  version: string
}

export interface CreateRuleRequest {
  name: string
  expression: string
  duration: string
  severity: string
  labels?: Record<string, string>
  annotations?: Record<string, string>
  targets?: string[]
}

export interface UpdateRuleRequest {
  name?: string
  expression?: string
  duration?: string
  severity?: string
  labels?: Record<string, string>
  annotations?: Record<string, string>
  targets?: string[]
}

// 知识库相关类型
export interface Knowledge {
  id: number
  created_at: string
  updated_at: string
  title: string
  content: string
  category: string
  tags?: string | string[] // 支持字符串和数组两种格式
  source: string
  source_id: number
  summary?: string
  similarity?: number
}

export interface CreateKnowledgeRequest {
  title: string
  content: string
  category: string
  tags?: string
}

export interface UpdateKnowledgeRequest {
  title?: string
  content?: string
  category?: string
  tags?: string
}

// 数据源相关类型
export interface Provider {
  id: number
  name: string
  type: 'prometheus' | 'victoriametrics'
  status: 'active' | 'inactive'
  description?: string
  endpoint: string
  auth_type?: string
  auth_config?: string
  labels?: string
  last_check?: string
  last_error?: string
}

export interface CreateProviderRequest {
  name: string
  type: 'prometheus' | 'victoriametrics'
  description?: string
  endpoint: string
  auth_type?: string
  auth_config?: string
  labels?: string
}

export interface UpdateProviderRequest {
  name?: string
  type?: 'prometheus' | 'victoriametrics'
  description?: string
  endpoint?: string
  auth_type?: string
  auth_config?: string
  labels?: string
}

// 通知相关类型
export interface NotificationGroup {
  id: number
  created_at: string
  updated_at: string
  name: string
  description?: string
  channels: NotificationChannel[]
  enabled: boolean
}

export interface NotificationChannel {
  id: number
  type: 'email' | 'webhook' | 'slack' | 'dingtalk'
  config: Record<string, any>
  enabled: boolean
}

export interface NotificationTemplate {
  id: number
  created_at: string
  updated_at: string
  name: string
  type: 'email' | 'webhook' | 'slack' | 'dingtalk' | 'wechat'
  description?: string
  enabled: boolean
  template: {
    subject?: string
    body?: string
    content?: string
    text?: string
    format?: string
    msgType?: string
    title?: string
    blocks?: string
  }
}

// 系统设置类型
export interface Settings {
  id: number
  key: string
  value: string
  description?: string
  type: 'string' | 'number' | 'boolean' | 'json'
  category: string
}

// 插件相关类型
export interface PluginInfo {
  name: string
  version: string
  description: string
  schema: Record<string, any>
  status: 'active' | 'inactive' | 'error'
  load_time: string
  last_error?: string
}

export interface PluginConfig {
  name: string
  enabled: boolean
  config: Record<string, any>
  priority: number
}

export interface PluginTestResult {
  success: boolean
  error?: string
  duration: number
  timestamp: string
}

export interface PluginStats {
  name: string
  total_sent: number
  success_count: number
  failure_count: number
  avg_duration: number
  last_sent: string
  last_error?: string
}

export interface PluginHealthStatus {
  name: string
  status: 'healthy' | 'unhealthy' | 'unknown'
  last_check: string
  error?: string
}

// 用户认证相关类型
export interface User {
  id: number
  username: string
  email: string
  role: 'admin' | 'operator' | 'viewer'
  status: 'active' | 'inactive'
  avatar?: string
  email_verified?: boolean
  last_login_at?: string
  created_at: string
  updated_at: string
}

export interface LoginRequest {
  username: string
  password: string
}

export interface LoginResponse {
  token: string
  user: User
  expires_at: string
}

// 异步任务相关类型
export interface AsyncTask {
  task_id: number
  submit_time: string
}

export interface AsyncTaskResult {
  status: 'pending' | 'running' | 'completed' | 'failed'
  result?: string
  error?: string
  message?: string
}

// 菜单项类型
export interface MenuItem {
  key: string
  label: string
  icon?: string
  path?: string
  children?: MenuItem[]
}

// 表格列配置类型
export interface TableColumn {
  title: string
  dataIndex: string
  key: string
  width?: number
  fixed?: 'left' | 'right'
  sorter?: boolean
  filters?: Array<{ text: string; value: any }>
  render?: (value: any, record: any, index: number) => any
}