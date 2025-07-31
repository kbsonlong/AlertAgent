// 模拟API响应，用于演示API集成效果
// 在实际项目中，这些数据将来自真实的后端API

export interface MockApiResponse<T = any> {
  code: number
  message: string
  data: T
  timestamp: string
}

// 模拟延迟
const delay = (ms: number) => new Promise(resolve => setTimeout(resolve, ms))

// 模拟系统信息
export const mockSystemInfo = async (): Promise<MockApiResponse> => {
  await delay(500)
  return {
    code: 200,
    message: 'success',
    data: {
      version: '1.0.0',
      uptime: '2天3小时45分钟',
      cpu_usage: 45.2,
      memory_usage: 68.7,
      disk_usage: 32.1,
      active_alerts: 12,
      total_rules: 156,
      last_backup: '2024-01-15 10:30:00'
    },
    timestamp: new Date().toISOString()
  }
}

// 模拟知识库列表
export const mockKnowledgeList = async (): Promise<MockApiResponse> => {
  await delay(300)
  return {
    code: 200,
    message: 'success',
    data: {
      items: [
        {
          id: 1,
          title: 'CPU使用率过高处理方案',
          content: '当CPU使用率超过80%时的处理步骤...',
          category: '性能优化',
          tags: ['CPU', '性能', '监控'],
          created_at: '2024-01-10 14:30:00',
          updated_at: '2024-01-12 09:15:00',
          views: 245,
          likes: 18
        },
        {
          id: 2,
          title: '内存泄漏排查指南',
          content: '如何快速定位和解决内存泄漏问题...',
          category: '故障排查',
          tags: ['内存', '调试', '性能'],
          created_at: '2024-01-08 16:20:00',
          updated_at: '2024-01-11 11:45:00',
          views: 189,
          likes: 23
        },
        {
          id: 3,
          title: '数据库连接池配置最佳实践',
          content: '合理配置数据库连接池参数的建议...',
          category: '数据库',
          tags: ['数据库', '连接池', '配置'],
          created_at: '2024-01-05 10:10:00',
          updated_at: '2024-01-09 15:30:00',
          views: 312,
          likes: 35
        }
      ],
      total: 3,
      page: 1,
      page_size: 10
    },
    timestamp: new Date().toISOString()
  }
}

// 模拟告警列表
export const mockAlertList = async (): Promise<MockApiResponse> => {
  await delay(400)
  return {
    code: 200,
    message: 'success',
    data: {
      items: [
        {
          id: 1,
          title: 'Web服务器CPU使用率过高',
          level: 'critical',
          status: 'active',
          source: 'prometheus',
          created_at: '2024-01-15 14:25:00',
          updated_at: '2024-01-15 14:25:00',
          description: 'web-server-01的CPU使用率已达到95%，持续时间超过5分钟'
        },
        {
          id: 2,
          title: '数据库连接数异常',
          level: 'warning',
          status: 'acknowledged',
          source: 'mysql',
          created_at: '2024-01-15 13:45:00',
          updated_at: '2024-01-15 14:10:00',
          description: '数据库连接数超过阈值80%，当前连接数：156/200'
        },
        {
          id: 3,
          title: '磁盘空间不足',
          level: 'warning',
          status: 'resolved',
          source: 'system',
          created_at: '2024-01-15 12:30:00',
          updated_at: '2024-01-15 13:15:00',
          description: '/var/log目录磁盘使用率达到85%'
        }
      ],
      total: 3,
      page: 1,
      page_size: 10
    },
    timestamp: new Date().toISOString()
  }
}

// 模拟数据源列表
export const mockProviderList = async (): Promise<MockApiResponse> => {
  await delay(350)
  return {
    code: 200,
    message: 'success',
    data: {
      items: [
        {
          id: 1,
          name: 'Prometheus监控',
          type: 'prometheus',
          url: 'http://prometheus:9090',
          status: 'connected',
          description: '主要的指标监控数据源',
          created_at: '2024-01-01 10:00:00',
          updated_at: '2024-01-15 14:20:00',
          last_check: '2024-01-15 14:20:00'
        },
        {
          id: 2,
          name: 'MySQL数据库',
          type: 'mysql',
          url: 'mysql://db:3306/alertagent',
          status: 'connected',
          description: '应用主数据库',
          created_at: '2024-01-01 10:00:00',
          updated_at: '2024-01-15 14:15:00',
          last_check: '2024-01-15 14:15:00'
        },
        {
          id: 3,
          name: 'ElasticSearch日志',
          type: 'elasticsearch',
          url: 'http://elasticsearch:9200',
          status: 'disconnected',
          description: '日志聚合和搜索',
          created_at: '2024-01-05 15:30:00',
          updated_at: '2024-01-15 14:00:00',
          last_check: '2024-01-15 14:00:00'
        }
      ],
      total: 3,
      page: 1,
      page_size: 10
    },
    timestamp: new Date().toISOString()
  }
}

// 模拟用户信息
export const mockUserInfo = async (): Promise<MockApiResponse> => {
  await delay(200)
  return {
    code: 200,
    message: 'success',
    data: {
      id: 1,
      username: 'admin',
      email: 'admin@alertagent.com',
      role: 'administrator',
      display_name: '系统管理员',
      avatar: '',
      last_login: '2024-01-15 14:00:00',
      created_at: '2024-01-01 10:00:00',
      permissions: [
        'system:read',
        'system:write',
        'user:read',
        'user:write',
        'alert:read',
        'alert:write',
        'knowledge:read',
        'knowledge:write'
      ]
    },
    timestamp: new Date().toISOString()
  }
}