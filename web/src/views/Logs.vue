<template>
  <div class="logs-page">
    <!-- 页面头部 -->
    <div class="page-header">
      <div class="header-content">
        <div class="header-left">
          <h1 class="page-title">系统日志</h1>
          <p class="page-description">查看和管理系统运行日志</p>
        </div>
        <div class="header-actions">
          <a-space>
            <a-button @click="handleRefresh" :loading="loading">
              <template #icon><ReloadOutlined /></template>
              刷新
            </a-button>
            <a-button @click="handleExport" :loading="exporting">
              <template #icon><DownloadOutlined /></template>
              导出日志
            </a-button>
            <a-button 
              type="primary" 
              danger 
              @click="handleClearLogs"
              :loading="clearing"
            >
              <template #icon><DeleteOutlined /></template>
              清空日志
            </a-button>
          </a-space>
        </div>
      </div>
    </div>

    <!-- 统计卡片 -->
    <a-row :gutter="[16, 16]" class="stats-section">
      <a-col :span="6">
        <a-card>
          <a-statistic
            title="总日志数"
            :value="stats.total"
            :value-style="{ color: '#1890ff' }"
          >
            <template #prefix>
              <FileTextOutlined />
            </template>
          </a-statistic>
        </a-card>
      </a-col>
      <a-col :span="6">
        <a-card>
          <a-statistic
            title="错误日志"
            :value="stats.error"
            :value-style="{ color: '#f5222d' }"
          >
            <template #prefix>
              <ExclamationCircleOutlined />
            </template>
          </a-statistic>
        </a-card>
      </a-col>
      <a-col :span="6">
        <a-card>
          <a-statistic
            title="警告日志"
            :value="stats.warning"
            :value-style="{ color: '#fa8c16' }"
          >
            <template #prefix>
              <WarningOutlined />
            </template>
          </a-statistic>
        </a-card>
      </a-col>
      <a-col :span="6">
        <a-card>
          <a-statistic
            title="今日日志"
            :value="stats.today"
            :value-style="{ color: '#52c41a' }"
          >
            <template #prefix>
              <CalendarOutlined />
            </template>
          </a-statistic>
        </a-card>
      </a-col>
    </a-row>

    <!-- 搜索和筛选 -->
    <a-card class="search-section">
      <a-form layout="inline" :model="searchForm" @finish="handleSearch">
        <a-form-item label="关键词">
          <a-input
            v-model:value="searchForm.keyword"
            placeholder="搜索日志内容"
            style="width: 200px;"
            allowClear
          >
            <template #prefix><SearchOutlined /></template>
          </a-input>
        </a-form-item>
        
        <a-form-item label="日志级别">
          <a-select
            v-model:value="searchForm.level"
            placeholder="选择日志级别"
            style="width: 120px;"
            allowClear
          >
            <a-select-option value="debug">Debug</a-select-option>
            <a-select-option value="info">Info</a-select-option>
            <a-select-option value="warning">Warning</a-select-option>
            <a-select-option value="error">Error</a-select-option>
            <a-select-option value="fatal">Fatal</a-select-option>
          </a-select>
        </a-form-item>
        
        <a-form-item label="来源">
          <a-select
            v-model:value="searchForm.source"
            placeholder="选择日志来源"
            style="width: 150px;"
            allowClear
          >
            <a-select-option value="api">API</a-select-option>
            <a-select-option value="scheduler">调度器</a-select-option>
            <a-select-option value="alertmanager">告警管理</a-select-option>
            <a-select-option value="notification">通知</a-select-option>
            <a-select-option value="provider">数据源</a-select-option>
            <a-select-option value="auth">认证</a-select-option>
          </a-select>
        </a-form-item>
        
        <a-form-item label="时间范围">
          <a-range-picker
            v-model:value="searchForm.timeRange"
            show-time
            format="YYYY-MM-DD HH:mm:ss"
            style="width: 300px;"
          />
        </a-form-item>
        
        <a-form-item>
          <a-space>
            <a-button type="primary" html-type="submit" :loading="loading">
              搜索
            </a-button>
            <a-button @click="handleReset">
              重置
            </a-button>
          </a-space>
        </a-form-item>
      </a-form>
    </a-card>

    <!-- 日志列表 -->
    <a-card class="logs-table">
      <template #title>
        <div class="table-header">
          <span>日志列表</span>
          <div class="table-actions">
            <a-space>
              <a-switch
                v-model:checked="autoRefresh"
                checked-children="自动刷新"
                un-checked-children="手动刷新"
                @change="handleAutoRefreshChange"
              />
              <a-select
                v-model:value="refreshInterval"
                style="width: 100px;"
                size="small"
                :disabled="!autoRefresh"
              >
                <a-select-option :value="5">5秒</a-select-option>
                <a-select-option :value="10">10秒</a-select-option>
                <a-select-option :value="30">30秒</a-select-option>
                <a-select-option :value="60">1分钟</a-select-option>
              </a-select>
            </a-space>
          </div>
        </div>
      </template>
      
      <a-table
        :columns="columns"
        :data-source="logs.list"
        :pagination="{
          current: pagination.current,
          pageSize: pagination.pageSize,
          total: logs.total,
          showSizeChanger: true,
          showQuickJumper: true,
          showTotal: (total, range) => `第 ${range[0]}-${range[1]} 条，共 ${total} 条`
        }"
        :loading="loading"
        @change="handleTableChange"
        size="small"
        :scroll="{ x: 1200 }"
      >
        <template #bodyCell="{ column, record }">
          <template v-if="column.key === 'level'">
            <a-tag :color="getLevelColor(record.level)">
              {{ getLevelLabel(record.level) }}
            </a-tag>
          </template>
          
          <template v-else-if="column.key === 'message'">
            <div class="log-message">
              <a-tooltip :title="record.message">
                <span class="message-text">{{ record.message }}</span>
              </a-tooltip>
              <a-button 
                type="link" 
                size="small" 
                @click="showLogDetail(record)"
              >
                详情
              </a-button>
            </div>
          </template>
          
          <template v-else-if="column.key === 'source'">
            <a-tag>{{ getSourceLabel(record.source) }}</a-tag>
          </template>
          
          <template v-else-if="column.key === 'timestamp'">
            {{ formatDateTime(record.timestamp) }}
          </template>
        </template>
      </a-table>
    </a-card>

    <!-- 日志详情模态框 -->
    <a-modal
      v-model:open="detailModalVisible"
      title="日志详情"
      width="800px"
      :footer="null"
    >
      <div v-if="selectedLog" class="log-detail">
        <a-descriptions :column="2" bordered size="small">
          <a-descriptions-item label="ID">
            {{ selectedLog.id }}
          </a-descriptions-item>
          <a-descriptions-item label="级别">
            <a-tag :color="getLevelColor(selectedLog.level)">
              {{ getLevelLabel(selectedLog.level) }}
            </a-tag>
          </a-descriptions-item>
          <a-descriptions-item label="来源">
            <a-tag>{{ getSourceLabel(selectedLog.source) }}</a-tag>
          </a-descriptions-item>
          <a-descriptions-item label="时间">
            {{ formatDateTime(selectedLog.timestamp) }}
          </a-descriptions-item>
          <a-descriptions-item label="消息" :span="2">
            <pre class="log-message-detail">{{ selectedLog.message }}</pre>
          </a-descriptions-item>
          <a-descriptions-item label="堆栈信息" :span="2" v-if="selectedLog.stack">
            <pre class="log-stack">{{ selectedLog.stack }}</pre>
          </a-descriptions-item>
          <a-descriptions-item label="额外信息" :span="2" v-if="selectedLog.extra">
            <pre class="log-extra">{{ JSON.stringify(selectedLog.extra, null, 2) }}</pre>
          </a-descriptions-item>
        </a-descriptions>
      </div>
    </a-modal>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, onUnmounted } from 'vue'
import { message, Modal } from 'ant-design-vue'
import {
  ReloadOutlined,
  DownloadOutlined,
  DeleteOutlined,
  FileTextOutlined,
  ExclamationCircleOutlined,
  WarningOutlined,
  CalendarOutlined,
  SearchOutlined
} from '@ant-design/icons-vue'
import { getSystemLogs, clearSystemLogs } from '@/services/system'
import { formatDateTime } from '@/utils/datetime'
import type { TableColumnsType } from 'ant-design-vue'

// 日志接口
interface LogEntry {
  id: string
  level: string
  message: string
  timestamp: string
  source: string
  stack?: string
  extra?: any
}

// 响应式数据
const loading = ref(false)
const exporting = ref(false)
const clearing = ref(false)
const autoRefresh = ref(false)
const refreshInterval = ref(10)
const detailModalVisible = ref(false)
const selectedLog = ref<LogEntry | null>(null)
const refreshTimer = ref<NodeJS.Timeout | null>(null)

// 统计数据
const stats = reactive({
  total: 0,
  error: 0,
  warning: 0,
  today: 0
})

// 日志列表
const logs = reactive({
  list: [] as LogEntry[],
  total: 0
})

// 分页
const pagination = reactive({
  current: 1,
  pageSize: 20
})

// 搜索表单
const searchForm = reactive({
  keyword: '',
  level: '',
  source: '',
  timeRange: null as any
})

// 表格列定义
const columns: TableColumnsType = [
  {
    title: 'ID',
    dataIndex: 'id',
    key: 'id',
    width: 100,
    ellipsis: true
  },
  {
    title: '级别',
    dataIndex: 'level',
    key: 'level',
    width: 80,
    filters: [
      { text: 'Debug', value: 'debug' },
      { text: 'Info', value: 'info' },
      { text: 'Warning', value: 'warning' },
      { text: 'Error', value: 'error' },
      { text: 'Fatal', value: 'fatal' }
    ]
  },
  {
    title: '消息',
    dataIndex: 'message',
    key: 'message',
    ellipsis: true
  },
  {
    title: '来源',
    dataIndex: 'source',
    key: 'source',
    width: 120,
    filters: [
      { text: 'API', value: 'api' },
      { text: '调度器', value: 'scheduler' },
      { text: '告警管理', value: 'alertmanager' },
      { text: '通知', value: 'notification' },
      { text: '数据源', value: 'provider' },
      { text: '认证', value: 'auth' }
    ]
  },
  {
    title: '时间',
    dataIndex: 'timestamp',
    key: 'timestamp',
    width: 180,
    sorter: true
  }
]

// 方法
const loadLogs = async () => {
  loading.value = true
  try {
    const params = {
      page: pagination.current,
      pageSize: pagination.pageSize,
      keyword: searchForm.keyword,
      level: searchForm.level,
      source: searchForm.source,
      startTime: searchForm.timeRange?.[0]?.format('YYYY-MM-DD HH:mm:ss'),
      endTime: searchForm.timeRange?.[1]?.format('YYYY-MM-DD HH:mm:ss')
    }
    
    const response = await getSystemLogs(params)
    logs.list = response.list
    logs.total = response.total
    
    // 更新统计数据
    updateStats(response.list)
  } catch (error) {
    message.error('获取日志失败')
    console.error('获取日志失败:', error)
  } finally {
    loading.value = false
  }
}

const updateStats = (logList: LogEntry[]) => {
  stats.total = logs.total
  stats.error = logList.filter(log => log.level === 'error').length
  stats.warning = logList.filter(log => log.level === 'warning').length
  
  const today = new Date().toDateString()
  stats.today = logList.filter(log => 
    new Date(log.timestamp).toDateString() === today
  ).length
}

const handleRefresh = () => {
  loadLogs()
}

const handleSearch = () => {
  pagination.current = 1
  loadLogs()
}

const handleReset = () => {
  searchForm.keyword = ''
  searchForm.level = ''
  searchForm.source = ''
  searchForm.timeRange = null
  pagination.current = 1
  loadLogs()
}

const handleTableChange = (pag: any, filters: any, sorter: any) => {
  pagination.current = pag.current
  pagination.pageSize = pag.pageSize
  
  // 处理筛选
  if (filters.level) {
    searchForm.level = filters.level[0]
  }
  if (filters.source) {
    searchForm.source = filters.source[0]
  }
  
  loadLogs()
}

const handleExport = async () => {
  exporting.value = true
  try {
    // 这里应该调用导出API
    await new Promise(resolve => setTimeout(resolve, 1000)) // 模拟导出
    message.success('日志导出成功')
  } catch (error) {
    message.error('日志导出失败')
    console.error('日志导出失败:', error)
  } finally {
    exporting.value = false
  }
}

const handleClearLogs = () => {
  Modal.confirm({
    title: '确认清空日志',
    content: '确定要清空所有日志吗？此操作不可恢复。',
    okText: '确认',
    cancelText: '取消',
    okType: 'danger',
    onOk: async () => {
      clearing.value = true
      try {
        await clearSystemLogs()
        message.success('日志清空成功')
        loadLogs()
      } catch (error) {
        message.error('日志清空失败')
        console.error('日志清空失败:', error)
      } finally {
        clearing.value = false
      }
    }
  })
}

const handleAutoRefreshChange = (checked: boolean) => {
  if (checked) {
    startAutoRefresh()
  } else {
    stopAutoRefresh()
  }
}

const startAutoRefresh = () => {
  stopAutoRefresh()
  refreshTimer.value = setInterval(() => {
    loadLogs()
  }, refreshInterval.value * 1000)
}

const stopAutoRefresh = () => {
  if (refreshTimer.value) {
    clearInterval(refreshTimer.value)
    refreshTimer.value = null
  }
}

const showLogDetail = (log: LogEntry) => {
  selectedLog.value = log
  detailModalVisible.value = true
}

// 辅助函数
const getLevelColor = (level: string) => {
  const colors = {
    debug: 'default',
    info: 'blue',
    warning: 'orange',
    error: 'red',
    fatal: 'purple'
  }
  return colors[level as keyof typeof colors] || 'default'
}

const getLevelLabel = (level: string) => {
  const labels = {
    debug: 'Debug',
    info: 'Info',
    warning: 'Warning',
    error: 'Error',
    fatal: 'Fatal'
  }
  return labels[level as keyof typeof labels] || level
}

const getSourceLabel = (source: string) => {
  const labels = {
    api: 'API',
    scheduler: '调度器',
    alertmanager: '告警管理',
    notification: '通知',
    provider: '数据源',
    auth: '认证'
  }
  return labels[source as keyof typeof labels] || source
}

// 生命周期
onMounted(() => {
  loadLogs()
})

onUnmounted(() => {
  stopAutoRefresh()
})
</script>

<style scoped>
.logs-page {
  padding: 24px;
}

.page-header {
  background: #fff;
  padding: 24px;
  margin-bottom: 16px;
  border-radius: 6px;
}

.header-content {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.header-left {
  flex: 1;
}

.page-title {
  margin: 0 0 8px 0;
  font-size: 24px;
  font-weight: 600;
  color: #262626;
}

.page-description {
  margin: 0;
  color: #8c8c8c;
  font-size: 14px;
}

.stats-section {
  margin-bottom: 16px;
}

.search-section {
  margin-bottom: 16px;
}

.logs-table {
  margin-bottom: 16px;
}

.table-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  width: 100%;
}

.log-message {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.message-text {
  flex: 1;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  margin-right: 8px;
}

.log-detail {
  max-height: 600px;
  overflow-y: auto;
}

.log-message-detail,
.log-stack,
.log-extra {
  background: #f5f5f5;
  padding: 12px;
  border-radius: 4px;
  font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
  font-size: 12px;
  line-height: 1.5;
  white-space: pre-wrap;
  word-break: break-all;
  max-height: 200px;
  overflow-y: auto;
  margin: 0;
}

.log-stack {
  color: #d32f2f;
}

.log-extra {
  color: #1976d2;
}
</style>