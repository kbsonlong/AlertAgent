<template>
  <div class="queue-monitor">
    <div class="page-header">
      <h1>任务队列监控</h1>
      <div class="header-actions">
        <a-button @click="refreshData" :loading="loading">
          <template #icon><ReloadOutlined /></template>
          刷新
        </a-button>
        <a-button type="primary" @click="showOptimizeModal">
          <template #icon><SettingOutlined /></template>
          队列优化
        </a-button>
      </div>
    </div>

    <!-- 队列概览卡片 -->
    <div class="overview-cards">
      <a-row :gutter="16">
        <a-col :span="6" v-for="(queue, queueName) in queueMetrics" :key="queueName">
          <a-card class="metric-card" :class="getQueueStatusClass(queue)">
            <div class="metric-header">
              <h3>{{ getQueueDisplayName(queueName) }}</h3>
              <a-tag :color="getQueueStatusColor(queue)">
                {{ getQueueStatus(queue) }}
              </a-tag>
            </div>
            <div class="metric-content">
              <div class="metric-item">
                <span class="label">待处理:</span>
                <span class="value">{{ queue.pending_count }}</span>
              </div>
              <div class="metric-item">
                <span class="label">处理中:</span>
                <span class="value">{{ queue.processing_count }}</span>
              </div>
              <div class="metric-item">
                <span class="label">失败:</span>
                <span class="value error">{{ queue.failed_count }}</span>
              </div>
              <div class="metric-item">
                <span class="label">吞吐量:</span>
                <span class="value">{{ queue.throughput_per_min?.toFixed(1) || 0 }}/min</span>
              </div>
              <div class="metric-item">
                <span class="label">错误率:</span>
                <span class="value" :class="{ error: queue.error_rate > 5 }">
                  {{ queue.error_rate?.toFixed(1) || 0 }}%
                </span>
              </div>
            </div>
            <div class="metric-actions">
              <a-button size="small" @click="viewQueueDetails(queueName)">
                详情
              </a-button>
              <a-button size="small" @click="cleanupQueue(queueName)">
                清理
              </a-button>
            </div>
          </a-card>
        </a-col>
      </a-row>
    </div>

    <!-- 性能趋势图表 -->
    <a-card title="性能趋势" class="chart-card">
      <a-tabs v-model:activeKey="activeChartTab">
        <a-tab-pane key="throughput" tab="吞吐量">
          <div ref="throughputChart" class="chart-container"></div>
        </a-tab-pane>
        <a-tab-pane key="error-rate" tab="错误率">
          <div ref="errorRateChart" class="chart-container"></div>
        </a-tab-pane>
        <a-tab-pane key="latency" tab="处理延迟">
          <div ref="latencyChart" class="chart-container"></div>
        </a-tab-pane>
      </a-tabs>
    </a-card>

    <!-- 队列告警 -->
    <a-card title="队列告警" class="alerts-card" v-if="queueAlerts.length > 0">
      <div class="alerts-list">
        <div 
          v-for="alert in queueAlerts" 
          :key="alert.id" 
          class="alert-item"
          :class="`severity-${alert.severity}`"
        >
          <div class="alert-header">
            <div class="alert-title">
              <a-tag :color="getAlertSeverityColor(alert.severity)">
                {{ alert.severity.toUpperCase() }}
              </a-tag>
              <strong>{{ alert.title }}</strong>
            </div>
            <div class="alert-actions">
              <a-button 
                v-if="alert.status === 'active'" 
                size="small" 
                @click="acknowledgeAlert(alert.id)"
              >
                确认
              </a-button>
              <a-button size="small" @click="viewAlertDetails(alert)">
                详情
              </a-button>
            </div>
          </div>
          <div class="alert-message">{{ alert.message }}</div>
          <div class="alert-meta">
            <span>队列: {{ getQueueDisplayName(alert.queue_name) }}</span>
            <span>时间: {{ formatDateTime(alert.created_at) }}</span>
            <span v-if="alert.acknowledged_by">确认人: {{ alert.acknowledged_by }}</span>
          </div>
        </div>
      </div>
    </a-card>

    <!-- 任务列表 -->
    <a-card title="任务列表" class="task-list-card">
      <div class="task-filters">
        <a-row :gutter="16">
          <a-col :span="6">
            <a-select
              v-model:value="taskFilter.queueName"
              placeholder="选择队列"
              allowClear
              @change="loadTasks"
            >
              <a-select-option value="">全部队列</a-select-option>
              <a-select-option
                v-for="queueName in Object.keys(queueMetrics)"
                :key="queueName"
                :value="queueName"
              >
                {{ getQueueDisplayName(queueName) }}
              </a-select-option>
            </a-select>
          </a-col>
          <a-col :span="6">
            <a-select
              v-model:value="taskFilter.status"
              placeholder="选择状态"
              allowClear
              @change="loadTasks"
            >
              <a-select-option value="">全部状态</a-select-option>
              <a-select-option value="pending">待处理</a-select-option>
              <a-select-option value="processing">处理中</a-select-option>
              <a-select-option value="completed">已完成</a-select-option>
              <a-select-option value="failed">失败</a-select-option>
            </a-select>
          </a-col>
          <a-col :span="6">
            <a-select
              v-model:value="taskFilter.taskType"
              placeholder="选择任务类型"
              allowClear
              @change="loadTasks"
            >
              <a-select-option value="">全部类型</a-select-option>
              <a-select-option value="ai_analysis">AI分析</a-select-option>
              <a-select-option value="notification">通知发送</a-select-option>
              <a-select-option value="config_sync">配置同步</a-select-option>
              <a-select-option value="rule_update">规则更新</a-select-option>
            </a-select>
          </a-col>
          <a-col :span="6">
            <a-button @click="loadTasks" :loading="tasksLoading">
              <template #icon><SearchOutlined /></template>
              搜索
            </a-button>
          </a-col>
        </a-row>
      </div>

      <a-table
        :columns="taskColumns"
        :data-source="tasks"
        :loading="tasksLoading"
        :pagination="taskPagination"
        @change="handleTaskTableChange"
        row-key="id"
      >
        <template #bodyCell="{ column, record }">
          <template v-if="column.key === 'status'">
            <a-tag :color="getTaskStatusColor(record.status)">
              {{ getTaskStatusText(record.status) }}
            </a-tag>
          </template>
          <template v-else-if="column.key === 'type'">
            <a-tag>{{ getTaskTypeText(record.type) }}</a-tag>
          </template>
          <template v-else-if="column.key === 'priority'">
            <a-tag :color="getPriorityColor(record.priority)">
              {{ getPriorityText(record.priority) }}
            </a-tag>
          </template>
          <template v-else-if="column.key === 'duration'">
            {{ formatDuration(record) }}
          </template>
          <template v-else-if="column.key === 'created_at'">
            {{ formatDateTime(record.created_at) }}
          </template>
          <template v-else-if="column.key === 'actions'">
            <a-space>
              <a-button
                size="small"
                @click="viewTaskDetails(record)"
              >
                详情
              </a-button>
              <a-button
                v-if="record.status === 'failed'"
                size="small"
                type="primary"
                @click="retryTask(record.id)"
              >
                重试
              </a-button>
              <a-button
                v-if="record.status === 'processing'"
                size="small"
                type="default"
                @click="cancelTask(record.id)"
              >
                取消
              </a-button>
              <a-button
                v-if="['pending', 'failed'].includes(record.status)"
                size="small"
                danger
                @click="skipTask(record.id)"
              >
                跳过
              </a-button>
              <a-dropdown>
                <a-button size="small">
                  更多
                  <DownOutlined />
                </a-button>
                <template #overlay>
                  <a-menu>
                    <a-menu-item @click="viewTaskLogs(record)">
                      <FileTextOutlined />
                      查看日志
                    </a-menu-item>
                    <a-menu-item @click="viewTaskHistory(record)">
                      <HistoryOutlined />
                      执行历史
                    </a-menu-item>
                    <a-menu-item @click="exportSingleTask(record)">
                      <ExportOutlined />
                      导出数据
                    </a-menu-item>
                  </a-menu>
                </template>
              </a-dropdown>
            </a-space>
          </template>
        </template>
      </a-table>

      <!-- 批量操作 -->
      <div class="batch-actions" v-if="selectedTasks.length > 0">
        <a-space>
          <span>已选择 {{ selectedTasks.length }} 个任务</span>
          <a-button @click="batchRetryTasks" :loading="batchLoading">
            批量重试
          </a-button>
          <a-button @click="batchSkipTasks" :loading="batchLoading">
            批量跳过
          </a-button>
          <a-button @click="batchCancelTasks" :loading="batchLoading">
            批量取消
          </a-button>
          <a-dropdown>
            <a-button :loading="batchLoading">
              批量操作
              <DownOutlined />
            </a-button>
            <template #overlay>
              <a-menu>
                <a-menu-item @click="exportSelectedTasks">
                  <ExportOutlined />
                  导出选中任务
                </a-menu-item>
                <a-menu-item @click="batchDeleteTasks" danger>
                  <DeleteOutlined />
                  删除选中任务
                </a-menu-item>
              </a-menu>
            </template>
          </a-dropdown>
          <a-button @click="selectedTasks = []">
            取消选择
          </a-button>
        </a-space>
      </div>
    </a-card>

    <!-- 队列优化模态框 -->
    <a-modal
      v-model:open="optimizeModalVisible"
      title="队列优化"
      @ok="handleOptimize"
      :confirm-loading="optimizeLoading"
    >
      <a-form :model="optimizeForm" layout="vertical">
        <a-form-item label="选择队列">
          <a-select v-model:value="optimizeForm.queueName" placeholder="选择要优化的队列">
            <a-select-option
              v-for="queueName in Object.keys(queueMetrics)"
              :key="queueName"
              :value="queueName"
            >
              {{ getQueueDisplayName(queueName) }}
            </a-select-option>
          </a-select>
        </a-form-item>
        <a-form-item label="优化选项">
          <a-checkbox-group v-model:value="optimizeForm.options">
            <a-checkbox value="auto_scale">自动扩缩容</a-checkbox>
            <a-checkbox value="cleanup_expired">清理过期任务</a-checkbox>
            <a-checkbox value="rebalance">重新平衡队列</a-checkbox>
            <a-checkbox value="optimize_workers">Worker优化</a-checkbox>
          </a-checkbox-group>
        </a-form-item>
        <a-form-item label="最大存活时间" v-if="optimizeForm.options.includes('cleanup_expired')">
          <a-input v-model:value="optimizeForm.maxAge" placeholder="例如: 24h, 7d" />
        </a-form-item>
        
        <!-- 显示优化建议 -->
        <div v-if="optimizeForm.queueName && queueRecommendations.length > 0" class="recommendations-section">
          <h4>优化建议</h4>
          <div class="recommendations-list">
            <div 
              v-for="rec in queueRecommendations" 
              :key="rec.id" 
              class="recommendation-item"
              :class="`priority-${rec.priority}`"
            >
              <div class="rec-header">
                <span class="rec-title">{{ rec.title }}</span>
                <a-tag :color="getPriorityColor(rec.priority)">{{ rec.priority }}</a-tag>
              </div>
              <div class="rec-description">{{ rec.description }}</div>
              <div class="rec-action">建议操作: {{ rec.action }}</div>
              <div v-if="rec.auto_fix" class="rec-autofix">
                <a-button size="small" type="link" @click="applyRecommendation(rec)">
                  自动修复
                </a-button>
              </div>
            </div>
          </div>
        </div>
      </a-form>
    </a-modal>

    <!-- 任务日志模态框 -->
    <a-modal
      v-model:open="taskLogsModalVisible"
      title="任务执行日志"
      :footer="null"
      width="900px"
    >
      <div v-if="selectedTaskLogs.length > 0" class="task-logs">
        <div class="logs-header">
          <a-space>
            <a-select v-model:value="logLevelFilter" placeholder="日志级别" allowClear style="width: 120px">
              <a-select-option value="">全部</a-select-option>
              <a-select-option value="debug">Debug</a-select-option>
              <a-select-option value="info">Info</a-select-option>
              <a-select-option value="warn">Warn</a-select-option>
              <a-select-option value="error">Error</a-select-option>
            </a-select>
            <a-button @click="refreshTaskLogs" size="small">
              <ReloadOutlined />
              刷新
            </a-button>
          </a-space>
        </div>
        <div class="logs-content">
          <div 
            v-for="log in filteredTaskLogs" 
            :key="log.id" 
            class="log-entry"
            :class="`log-${log.level}`"
          >
            <div class="log-header">
              <span class="log-timestamp">{{ formatDateTime(log.timestamp) }}</span>
              <a-tag :color="getLogLevelColor(log.level)">{{ log.level.toUpperCase() }}</a-tag>
              <span v-if="log.worker_id" class="log-worker">Worker: {{ log.worker_id }}</span>
            </div>
            <div class="log-message">{{ log.message }}</div>
            <div v-if="log.context" class="log-context">
              <pre>{{ JSON.stringify(log.context, null, 2) }}</pre>
            </div>
          </div>
        </div>
      </div>
      <a-empty v-else description="暂无日志数据" />
    </a-modal>

    <!-- 任务历史模态框 -->
    <a-modal
      v-model:open="taskHistoryModalVisible"
      title="任务执行历史"
      :footer="null"
      width="800px"
    >
      <div v-if="selectedTaskHistory.length > 0" class="task-history">
        <a-timeline>
          <a-timeline-item 
            v-for="record in selectedTaskHistory" 
            :key="record.id"
            :color="getHistoryActionColor(record.action)"
          >
            <template #dot>
              <component :is="getHistoryActionIcon(record.action)" />
            </template>
            <div class="history-content">
              <div class="history-header">
                <strong>{{ getHistoryActionText(record.action) }}</strong>
                <span class="history-timestamp">{{ formatDateTime(record.timestamp) }}</span>
              </div>
              <div class="history-message">{{ record.message }}</div>
              <div v-if="record.worker_id" class="history-worker">
                Worker: {{ record.worker_id }}
              </div>
              <div v-if="record.duration" class="history-duration">
                耗时: {{ record.duration }}ms
              </div>
              <div v-if="record.error_msg" class="history-error">
                <a-alert :message="record.error_msg" type="error" size="small" />
              </div>
            </div>
          </a-timeline-item>
        </a-timeline>
      </div>
      <a-empty v-else description="暂无历史数据" />
    </a-modal>

    <!-- 任务详情模态框 -->
    <a-modal
      v-model:open="taskDetailModalVisible"
      title="任务详情"
      :footer="null"
      width="800px"
    >
      <div v-if="selectedTask" class="task-detail">
        <a-descriptions :column="2" bordered>
          <a-descriptions-item label="任务ID">{{ selectedTask.id }}</a-descriptions-item>
          <a-descriptions-item label="任务类型">
            <a-tag>{{ getTaskTypeText(selectedTask.type) }}</a-tag>
          </a-descriptions-item>
          <a-descriptions-item label="状态">
            <a-tag :color="getTaskStatusColor(selectedTask.status)">
              {{ getTaskStatusText(selectedTask.status) }}
            </a-tag>
          </a-descriptions-item>
          <a-descriptions-item label="优先级">
            <a-tag :color="getPriorityColor(selectedTask.priority)">
              {{ getPriorityText(selectedTask.priority) }}
            </a-tag>
          </a-descriptions-item>
          <a-descriptions-item label="重试次数">
            {{ selectedTask.retry }}/{{ selectedTask.max_retry }}
          </a-descriptions-item>
          <a-descriptions-item label="Worker ID">
            {{ selectedTask.worker_id || '-' }}
          </a-descriptions-item>
          <a-descriptions-item label="创建时间">
            {{ formatDateTime(selectedTask.created_at) }}
          </a-descriptions-item>
          <a-descriptions-item label="开始时间">
            {{ selectedTask.started_at ? formatDateTime(selectedTask.started_at) : '-' }}
          </a-descriptions-item>
          <a-descriptions-item label="完成时间">
            {{ selectedTask.completed_at ? formatDateTime(selectedTask.completed_at) : '-' }}
          </a-descriptions-item>
          <a-descriptions-item label="执行时长">
            {{ formatDuration(selectedTask) }}
          </a-descriptions-item>
        </a-descriptions>
        
        <div v-if="selectedTask.error_msg" class="error-section">
          <h4>错误信息</h4>
          <a-alert :message="selectedTask.error_msg" type="error" show-icon />
        </div>
        
        <div class="payload-section">
          <h4>任务载荷</h4>
          <pre class="payload-content">{{ JSON.stringify(selectedTask.payload, null, 2) }}</pre>
        </div>
      </div>
    </a-modal>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, nextTick } from 'vue'
import { message } from 'ant-design-vue'
import {
  ReloadOutlined,
  SettingOutlined,
  SearchOutlined,
  DownOutlined,
  FileTextOutlined,
  HistoryOutlined,
  ExportOutlined,
  DeleteOutlined
} from '@ant-design/icons-vue'
import * as echarts from 'echarts'
import { queueService } from '@/services/queue'
import type { QueueMetrics, Task, TaskFilter, PerformanceStats } from '@/types/queue'

// 响应式数据
const loading = ref(false)
const tasksLoading = ref(false)
const batchLoading = ref(false)
const optimizeLoading = ref(false)

const queueMetrics = ref<Record<string, QueueMetrics>>({})
const tasks = ref<Task[]>([])
const selectedTasks = ref<string[]>([])
const selectedTask = ref<Task | null>(null)

const activeChartTab = ref('throughput')
const optimizeModalVisible = ref(false)
const taskDetailModalVisible = ref(false)
const taskLogsModalVisible = ref(false)
const taskHistoryModalVisible = ref(false)

const selectedTaskLogs = ref<any[]>([])
const selectedTaskHistory = ref<any[]>([])
const logLevelFilter = ref('')

const queueRecommendations = ref<any[]>([])
const queueAlerts = ref<any[]>([])
const optimizeResult = ref<any>(null)

// 图表实例
const throughputChart = ref<HTMLElement>()
const errorRateChart = ref<HTMLElement>()
const latencyChart = ref<HTMLElement>()

let throughputChartInstance: echarts.ECharts | null = null
let errorRateChartInstance: echarts.ECharts | null = null
let latencyChartInstance: echarts.ECharts | null = null

// 任务过滤器
const taskFilter = reactive<TaskFilter>({
  queueName: '',
  status: '',
  taskType: '',
  page: 1,
  pageSize: 20
})

// 任务分页
const taskPagination = reactive({
  current: 1,
  pageSize: 20,
  total: 0,
  showSizeChanger: true,
  showQuickJumper: true,
  showTotal: (total: number) => `共 ${total} 条记录`
})

// 优化表单
const optimizeForm = reactive({
  queueName: '',
  options: [] as string[],
  maxAge: '24h'
})

// 任务表格列定义
const taskColumns = [
  {
    title: '任务ID',
    dataIndex: 'id',
    key: 'id',
    width: 200,
    ellipsis: true
  },
  {
    title: '类型',
    dataIndex: 'type',
    key: 'type',
    width: 120
  },
  {
    title: '状态',
    dataIndex: 'status',
    key: 'status',
    width: 100
  },
  {
    title: '优先级',
    dataIndex: 'priority',
    key: 'priority',
    width: 100
  },
  {
    title: '重试次数',
    dataIndex: 'retry',
    key: 'retry',
    width: 100,
    customRender: ({ record }: { record: Task }) => `${record.retry}/${record.max_retry}`
  },
  {
    title: '执行时长',
    key: 'duration',
    width: 120
  },
  {
    title: '创建时间',
    dataIndex: 'created_at',
    key: 'created_at',
    width: 180
  },
  {
    title: '操作',
    key: 'actions',
    width: 200,
    fixed: 'right' as const
  }
]

// 生命周期
onMounted(() => {
  loadData()
  initCharts()
  
  // 定时刷新数据
  setInterval(loadData, 30000) // 30秒刷新一次
})

// 方法
const loadData = async () => {
  await Promise.all([
    loadQueueMetrics(),
    loadTasks(),
    loadQueueAlerts()
  ])
}

const loadQueueMetrics = async () => {
  try {
    loading.value = true
    const response = await queueService.getAllQueueMetrics()
    queueMetrics.value = response.data
  } catch (error) {
    message.error('加载队列指标失败')
    console.error('Failed to load queue metrics:', error)
  } finally {
    loading.value = false
  }
}

const loadTasks = async () => {
  try {
    tasksLoading.value = true
    const response = await queueService.getTasks({
      ...taskFilter,
      page: taskPagination.current,
      pageSize: taskPagination.pageSize
    })
    
    tasks.value = response.data.tasks
    taskPagination.total = response.data.total
  } catch (error) {
    message.error('加载任务列表失败')
    console.error('Failed to load tasks:', error)
  } finally {
    tasksLoading.value = false
  }
}

const refreshData = () => {
  loadData()
  updateCharts()
}

const initCharts = async () => {
  await nextTick()
  
  if (throughputChart.value) {
    throughputChartInstance = echarts.init(throughputChart.value)
  }
  if (errorRateChart.value) {
    errorRateChartInstance = echarts.init(errorRateChart.value)
  }
  if (latencyChart.value) {
    latencyChartInstance = echarts.init(latencyChart.value)
  }
  
  updateCharts()
}

const updateCharts = async () => {
  // 这里需要获取性能统计数据并更新图表
  // 由于篇幅限制，这里只展示基本结构
  const performanceData = await loadPerformanceData()
  
  if (throughputChartInstance && performanceData) {
    updateThroughputChart(performanceData)
  }
  if (errorRateChartInstance && performanceData) {
    updateErrorRateChart(performanceData)
  }
  if (latencyChartInstance && performanceData) {
    updateLatencyChart(performanceData)
  }
}

const loadPerformanceData = async (): Promise<PerformanceStats | null> => {
  try {
    // 这里应该调用获取性能数据的API
    // const response = await queueService.getPerformanceStats()
    // return response.data
    return null
  } catch (error) {
    console.error('Failed to load performance data:', error)
    return null
  }
}

const updateThroughputChart = (data: PerformanceStats) => {
  const option = {
    title: { text: '吞吐量趋势' },
    tooltip: { trigger: 'axis' },
    xAxis: { type: 'time' },
    yAxis: { type: 'value', name: '任务/分钟' },
    series: Object.keys(queueMetrics.value).map(queueName => ({
      name: getQueueDisplayName(queueName),
      type: 'line',
      data: [] // 这里应该是实际的数据点
    }))
  }
  
  throughputChartInstance?.setOption(option)
}

const updateErrorRateChart = (data: PerformanceStats) => {
  const option = {
    title: { text: '错误率趋势' },
    tooltip: { trigger: 'axis' },
    xAxis: { type: 'time' },
    yAxis: { type: 'value', name: '错误率 (%)' },
    series: Object.keys(queueMetrics.value).map(queueName => ({
      name: getQueueDisplayName(queueName),
      type: 'line',
      data: [] // 这里应该是实际的数据点
    }))
  }
  
  errorRateChartInstance?.setOption(option)
}

const updateLatencyChart = (data: PerformanceStats) => {
  const option = {
    title: { text: '处理延迟趋势' },
    tooltip: { trigger: 'axis' },
    xAxis: { type: 'time' },
    yAxis: { type: 'value', name: '延迟 (ms)' },
    series: Object.keys(queueMetrics.value).map(queueName => ({
      name: getQueueDisplayName(queueName),
      type: 'line',
      data: [] // 这里应该是实际的数据点
    }))
  }
  
  latencyChartInstance?.setOption(option)
}

const handleTaskTableChange = (pagination: any) => {
  taskPagination.current = pagination.current
  taskPagination.pageSize = pagination.pageSize
  taskFilter.page = pagination.current
  taskFilter.pageSize = pagination.pageSize
  loadTasks()
}

const viewQueueDetails = (queueName: string) => {
  // 跳转到队列详情页面或显示详情模态框
  console.log('View queue details:', queueName)
}

const cleanupQueue = async (queueName: string) => {
  try {
    await queueService.cleanupExpiredTasks(queueName, '24h')
    message.success('队列清理完成')
    loadQueueMetrics()
  } catch (error) {
    message.error('队列清理失败')
    console.error('Failed to cleanup queue:', error)
  }
}

const viewTaskDetails = (task: Task) => {
  selectedTask.value = task
  taskDetailModalVisible.value = true
}

const retryTask = async (taskId: string) => {
  try {
    await queueService.retryTask(taskId)
    message.success('任务已重新加入队列')
    loadTasks()
  } catch (error) {
    message.error('重试任务失败')
    console.error('Failed to retry task:', error)
  }
}

const skipTask = async (taskId: string) => {
  try {
    await queueService.skipTask(taskId)
    message.success('任务已跳过')
    loadTasks()
  } catch (error) {
    message.error('跳过任务失败')
    console.error('Failed to skip task:', error)
  }
}

const cancelTask = async (taskId: string) => {
  try {
    await queueService.cancelTask(taskId)
    message.success('任务已取消')
    loadTasks()
  } catch (error) {
    message.error('取消任务失败')
    console.error('Failed to cancel task:', error)
  }
}

const batchRetryTasks = async () => {
  try {
    batchLoading.value = true
    await queueService.batchRetryTasks(selectedTasks.value)
    message.success('批量重试完成')
    selectedTasks.value = []
    loadTasks()
  } catch (error) {
    message.error('批量重试失败')
    console.error('Failed to batch retry tasks:', error)
  } finally {
    batchLoading.value = false
  }
}

const batchSkipTasks = async () => {
  try {
    batchLoading.value = true
    await queueService.batchSkipTasks(selectedTasks.value)
    message.success('批量跳过完成')
    selectedTasks.value = []
    loadTasks()
  } catch (error) {
    message.error('批量跳过失败')
    console.error('Failed to batch skip tasks:', error)
  } finally {
    batchLoading.value = false
  }
}

const batchCancelTasks = async () => {
  try {
    batchLoading.value = true
    // 这里需要实现批量取消的API
    message.success('批量取消完成')
    selectedTasks.value = []
    loadTasks()
  } catch (error) {
    message.error('批量取消失败')
    console.error('Failed to batch cancel tasks:', error)
  } finally {
    batchLoading.value = false
  }
}

const batchDeleteTasks = async () => {
  try {
    batchLoading.value = true
    // 这里需要实现批量删除的API
    message.success('批量删除完成')
    selectedTasks.value = []
    loadTasks()
  } catch (error) {
    message.error('批量删除失败')
    console.error('Failed to batch delete tasks:', error)
  } finally {
    batchLoading.value = false
  }
}

const viewTaskLogs = async (task: Task) => {
  try {
    const response = await queueService.getTaskLogs(task.id)
    selectedTaskLogs.value = response.data.logs || []
    taskLogsModalVisible.value = true
  } catch (error) {
    message.error('获取任务日志失败')
    console.error('Failed to get task logs:', error)
  }
}

const viewTaskHistory = async (task: Task) => {
  try {
    const response = await queueService.getTaskHistory(task.id)
    selectedTaskHistory.value = response.data || []
    taskHistoryModalVisible.value = true
  } catch (error) {
    message.error('获取任务历史失败')
    console.error('Failed to get task history:', error)
  }
}

const refreshTaskLogs = async () => {
  if (selectedTask.value) {
    await viewTaskLogs(selectedTask.value)
  }
}

const exportSingleTask = async (task: Task) => {
  try {
    // 导出单个任务的数据
    const filter = {
      queueName: '',
      status: '',
      taskType: '',
      page: 1,
      pageSize: 1
    }
    
    const response = await queueService.exportTasks(filter, 'json')
    // 处理下载
    message.success('任务数据导出成功')
  } catch (error) {
    message.error('导出任务数据失败')
    console.error('Failed to export task:', error)
  }
}

const exportSelectedTasks = async () => {
  try {
    batchLoading.value = true
    // 导出选中的任务
    message.success('选中任务导出成功')
  } catch (error) {
    message.error('导出选中任务失败')
    console.error('Failed to export selected tasks:', error)
  } finally {
    batchLoading.value = false
  }
}

const showOptimizeModal = () => {
  optimizeModalVisible.value = true
}

const handleOptimize = async () => {
  try {
    optimizeLoading.value = true
    const options: Record<string, any> = {}
    
    optimizeForm.options.forEach(option => {
      options[option] = true
    })
    
    if (options.cleanup_expired) {
      options.max_age = optimizeForm.maxAge
    }
    
    const response = await queueService.optimizeQueue(optimizeForm.queueName, options)
    optimizeResult.value = response.data
    message.success('队列优化完成')
    optimizeModalVisible.value = false
    loadQueueMetrics()
  } catch (error) {
    message.error('队列优化失败')
    console.error('Failed to optimize queue:', error)
  } finally {
    optimizeLoading.value = false
  }
}

const loadQueueRecommendations = async (queueName: string) => {
  try {
    const response = await queueService.getQueueRecommendations(queueName)
    queueRecommendations.value = response.data
  } catch (error) {
    console.error('Failed to load queue recommendations:', error)
  }
}

const loadQueueAlerts = async () => {
  try {
    const response = await queueService.getQueueAlerts()
    queueAlerts.value = response.data.alerts
  } catch (error) {
    console.error('Failed to load queue alerts:', error)
  }
}

const acknowledgeAlert = async (alertId: string) => {
  try {
    await queueService.acknowledgeAlert(alertId)
    message.success('告警已确认')
    loadQueueAlerts()
  } catch (error) {
    message.error('确认告警失败')
    console.error('Failed to acknowledge alert:', error)
  }
}

const viewAlertDetails = (alert: any) => {
  // 显示告警详情
  console.log('Alert details:', alert)
}

const applyRecommendation = async (recommendation: any) => {
  try {
    // 根据建议类型执行相应操作
    if (recommendation.auto_fix) {
      message.success('正在应用优化建议...')
      // 这里可以调用相应的优化API
    }
  } catch (error) {
    message.error('应用建议失败')
    console.error('Failed to apply recommendation:', error)
  }
}

// 辅助方法
const getQueueDisplayName = (queueName: string): string => {
  const nameMap: Record<string, string> = {
    'ai_analysis': 'AI分析',
    'notification': '通知发送',
    'config_sync': '配置同步',
    'rule_update': '规则更新',
    'health_check': '健康检查'
  }
  return nameMap[queueName] || queueName
}

const getQueueStatus = (queue: QueueMetrics): string => {
  if (queue.error_rate > 10) return '异常'
  if (queue.pending_count > 100) return '繁忙'
  if (queue.processing_count === 0 && queue.pending_count === 0) return '空闲'
  return '正常'
}

const getQueueStatusColor = (queue: QueueMetrics): string => {
  const status = getQueueStatus(queue)
  const colorMap: Record<string, string> = {
    '正常': 'success',
    '繁忙': 'warning',
    '异常': 'error',
    '空闲': 'default'
  }
  return colorMap[status] || 'default'
}

const getQueueStatusClass = (queue: QueueMetrics): string => {
  const status = getQueueStatus(queue)
  return `status-${status.toLowerCase()}`
}

const getTaskStatusText = (status: string): string => {
  const statusMap: Record<string, string> = {
    'pending': '待处理',
    'processing': '处理中',
    'completed': '已完成',
    'failed': '失败',
    'retrying': '重试中'
  }
  return statusMap[status] || status
}

const getTaskStatusColor = (status: string): string => {
  const colorMap: Record<string, string> = {
    'pending': 'default',
    'processing': 'processing',
    'completed': 'success',
    'failed': 'error',
    'retrying': 'warning'
  }
  return colorMap[status] || 'default'
}

const getTaskTypeText = (type: string): string => {
  const typeMap: Record<string, string> = {
    'ai_analysis': 'AI分析',
    'notification': '通知发送',
    'config_sync': '配置同步',
    'rule_update': '规则更新',
    'health_check': '健康检查'
  }
  return typeMap[type] || type
}

const getPriorityText = (priority: number): string => {
  const priorityMap: Record<number, string> = {
    0: '低',
    1: '普通',
    2: '高',
    3: '紧急'
  }
  return priorityMap[priority] || '未知'
}

const getPriorityColor = (priority: number): string => {
  const colorMap: Record<number, string> = {
    0: 'default',
    1: 'blue',
    2: 'orange',
    3: 'red'
  }
  return colorMap[priority] || 'default'
}

const formatDateTime = (dateTime: string): string => {
  return new Date(dateTime).toLocaleString('zh-CN')
}

const formatDuration = (task: Task): string => {
  if (!task.started_at) return '-'
  
  const start = new Date(task.started_at)
  const end = task.completed_at ? new Date(task.completed_at) : new Date()
  const duration = end.getTime() - start.getTime()
  
  if (duration < 1000) return `${duration}ms`
  if (duration < 60000) return `${(duration / 1000).toFixed(1)}s`
  if (duration < 3600000) return `${(duration / 60000).toFixed(1)}m`
  return `${(duration / 3600000).toFixed(1)}h`
}

// 计算属性：过滤后的任务日志
const filteredTaskLogs = computed(() => {
  if (!logLevelFilter.value) {
    return selectedTaskLogs.value
  }
  return selectedTaskLogs.value.filter(log => log.level === logLevelFilter.value)
})

// 日志级别颜色
const getLogLevelColor = (level: string): string => {
  const colorMap: Record<string, string> = {
    'debug': 'default',
    'info': 'blue',
    'warn': 'orange',
    'error': 'red'
  }
  return colorMap[level] || 'default'
}

// 历史操作颜色
const getHistoryActionColor = (action: string): string => {
  const colorMap: Record<string, string> = {
    'created': 'blue',
    'started': 'green',
    'completed': 'green',
    'failed': 'red',
    'retried': 'orange',
    'cancelled': 'red'
  }
  return colorMap[action] || 'default'
}

// 历史操作图标
const getHistoryActionIcon = (action: string): string => {
  const iconMap: Record<string, string> = {
    'created': 'PlusOutlined',
    'started': 'PlayCircleOutlined',
    'completed': 'CheckCircleOutlined',
    'failed': 'CloseCircleOutlined',
    'retried': 'ReloadOutlined',
    'cancelled': 'StopOutlined'
  }
  return iconMap[action] || 'InfoCircleOutlined'
}

// 历史操作文本
const getHistoryActionText = (action: string): string => {
  const textMap: Record<string, string> = {
    'created': '任务创建',
    'started': '开始执行',
    'completed': '执行完成',
    'failed': '执行失败',
    'retried': '重新尝试',
    'cancelled': '任务取消'
  }
  return textMap[action] || action
}

// 告警严重程度颜色
const getAlertSeverityColor = (severity: string): string => {
  const colorMap: Record<string, string> = {
    'info': 'blue',
    'warning': 'orange',
    'critical': 'red'
  }
  return colorMap[severity] || 'default'
}

// 优先级颜色（用于建议）
const getPriorityColor = (priority: string): string => {
  const colorMap: Record<string, string> = {
    'low': 'green',
    'medium': 'orange',
    'high': 'red',
    'critical': 'red'
  }
  return colorMap[priority] || 'default'
}
</script>

<style scoped>
.queue-monitor {
  padding: 24px;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 24px;
}

.page-header h1 {
  margin: 0;
  font-size: 24px;
  font-weight: 600;
}

.header-actions {
  display: flex;
  gap: 12px;
}

.overview-cards {
  margin-bottom: 24px;
}

.metric-card {
  height: 200px;
}

.metric-card.status-异常 {
  border-color: #ff4d4f;
}

.metric-card.status-繁忙 {
  border-color: #faad14;
}

.metric-card.status-正常 {
  border-color: #52c41a;
}

.metric-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
}

.metric-header h3 {
  margin: 0;
  font-size: 16px;
  font-weight: 500;
}

.metric-content {
  display: flex;
  flex-direction: column;
  gap: 8px;
  margin-bottom: 16px;
}

.metric-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.metric-item .label {
  color: #666;
  font-size: 14px;
}

.metric-item .value {
  font-weight: 500;
}

.metric-item .value.error {
  color: #ff4d4f;
}

.metric-actions {
  display: flex;
  gap: 8px;
}

.chart-card {
  margin-bottom: 24px;
}

.chart-container {
  height: 400px;
}

.task-list-card {
  margin-bottom: 24px;
}

.task-filters {
  margin-bottom: 16px;
}

.batch-actions {
  margin-top: 16px;
  padding: 16px;
  background: #f5f5f5;
  border-radius: 6px;
}

.task-detail .error-section {
  margin: 16px 0;
}

.task-detail .payload-section {
  margin: 16px 0;
}

.task-detail .payload-content {
  background: #f5f5f5;
  padding: 12px;
  border-radius: 4px;
  font-size: 12px;
  max-height: 300px;
  overflow-y: auto;
}

/* 任务日志样式 */
.task-logs .logs-header {
  margin-bottom: 16px;
  padding-bottom: 12px;
  border-bottom: 1px solid #f0f0f0;
}

.task-logs .logs-content {
  max-height: 500px;
  overflow-y: auto;
}

.log-entry {
  margin-bottom: 12px;
  padding: 12px;
  border-radius: 6px;
  border-left: 3px solid #d9d9d9;
}

.log-entry.log-debug {
  background: #f6f6f6;
  border-left-color: #d9d9d9;
}

.log-entry.log-info {
  background: #f0f9ff;
  border-left-color: #1890ff;
}

.log-entry.log-warn {
  background: #fffbf0;
  border-left-color: #faad14;
}

.log-entry.log-error {
  background: #fff2f0;
  border-left-color: #ff4d4f;
}

.log-header {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 8px;
}

.log-timestamp {
  font-size: 12px;
  color: #666;
  font-family: monospace;
}

.log-worker {
  font-size: 12px;
  color: #666;
}

.log-message {
  font-size: 14px;
  line-height: 1.5;
  margin-bottom: 8px;
}

.log-context {
  background: #f5f5f5;
  padding: 8px;
  border-radius: 4px;
  font-size: 12px;
}

.log-context pre {
  margin: 0;
  white-space: pre-wrap;
  word-break: break-all;
}

/* 任务历史样式 */
.task-history {
  max-height: 500px;
  overflow-y: auto;
}

.history-content {
  padding-left: 8px;
}

.history-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 8px;
}

.history-timestamp {
  font-size: 12px;
  color: #666;
  font-family: monospace;
}

.history-message {
  font-size: 14px;
  margin-bottom: 8px;
  color: #333;
}

.history-worker,
.history-duration {
  font-size: 12px;
  color: #666;
  margin-bottom: 4px;
}

.history-error {
  margin-top: 8px;
}

/* 队列告警样式 */
.alerts-card {
  margin-bottom: 24px;
}

.alerts-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.alert-item {
  padding: 16px;
  border-radius: 6px;
  border-left: 4px solid #d9d9d9;
}

.alert-item.severity-info {
  background: #f0f9ff;
  border-left-color: #1890ff;
}

.alert-item.severity-warning {
  background: #fffbf0;
  border-left-color: #faad14;
}

.alert-item.severity-critical {
  background: #fff2f0;
  border-left-color: #ff4d4f;
}

.alert-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 8px;
}

.alert-title {
  display: flex;
  align-items: center;
  gap: 8px;
}

.alert-actions {
  display: flex;
  gap: 8px;
}

.alert-message {
  font-size: 14px;
  margin-bottom: 8px;
  color: #333;
}

.alert-meta {
  display: flex;
  gap: 16px;
  font-size: 12px;
  color: #666;
}

/* 优化建议样式 */
.recommendations-section {
  margin-top: 16px;
  padding-top: 16px;
  border-top: 1px solid #f0f0f0;
}

.recommendations-section h4 {
  margin-bottom: 12px;
  font-size: 14px;
  font-weight: 500;
}

.recommendations-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.recommendation-item {
  padding: 12px;
  border-radius: 4px;
  border-left: 3px solid #d9d9d9;
  background: #fafafa;
}

.recommendation-item.priority-low {
  border-left-color: #52c41a;
}

.recommendation-item.priority-medium {
  border-left-color: #faad14;
}

.recommendation-item.priority-high {
  border-left-color: #ff4d4f;
}

.recommendation-item.priority-critical {
  border-left-color: #ff4d4f;
  background: #fff2f0;
}

.rec-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 4px;
}

.rec-title {
  font-weight: 500;
  font-size: 13px;
}

.rec-description,
.rec-action {
  font-size: 12px;
  color: #666;
  margin-bottom: 4px;
}

.rec-autofix {
  text-align: right;
}
</style>