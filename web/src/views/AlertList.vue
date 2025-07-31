<template>
  <div class="alert-list">
    <!-- 页面标题和操作栏 -->
    <div class="page-header">
      <div class="header-left">
        <h2>告警管理</h2>
        <p>管理和监控系统告警信息</p>
      </div>
      <div class="header-right">
        <a-space>
          <a-button type="primary" @click="refreshData">
            <template #icon><ReloadOutlined /></template>
            刷新
          </a-button>
          <a-button @click="showBatchActions = !showBatchActions">
            <template #icon><SettingOutlined /></template>
            批量操作
          </a-button>
        </a-space>
      </div>
    </div>

    <!-- 统计卡片 -->
    <div class="stats-cards">
      <a-row :gutter="16">
        <a-col :span="6">
          <a-card>
            <a-statistic
              title="总告警数"
              :value="alertStats.total"
              :value-style="{ color: '#1890ff' }"
            >
              <template #prefix><AlertOutlined /></template>
            </a-statistic>
          </a-card>
        </a-col>
        <a-col :span="6">
          <a-card>
            <a-statistic
              title="触发中"
              :value="alertStats.firing"
              :value-style="{ color: '#ff4d4f' }"
            >
              <template #prefix><ExclamationCircleOutlined /></template>
            </a-statistic>
          </a-card>
        </a-col>
        <a-col :span="6">
          <a-card>
            <a-statistic
              title="已确认"
              :value="alertStats.acknowledged"
              :value-style="{ color: '#1890ff' }"
            >
              <template #prefix><CheckCircleOutlined /></template>
            </a-statistic>
          </a-card>
        </a-col>
        <a-col :span="6">
          <a-card>
            <a-statistic
              title="已解决"
              :value="alertStats.resolved"
              :value-style="{ color: '#52c41a' }"
            >
              <template #prefix><CheckOutlined /></template>
            </a-statistic>
          </a-card>
        </a-col>
      </a-row>
    </div>

    <!-- 搜索和筛选 -->
    <a-card class="search-card">
      <a-form layout="inline" :model="searchForm" @finish="handleSearch">
        <a-form-item label="关键词">
          <a-input
            v-model:value="searchForm.search"
            placeholder="搜索告警名称、描述"
            allow-clear
            style="width: 200px"
          >
            <template #prefix><SearchOutlined /></template>
          </a-input>
        </a-form-item>
        <a-form-item label="状态">
          <a-select
            v-model:value="searchForm.status"
            placeholder="选择状态"
            allow-clear
            style="width: 120px"
          >
            <a-select-option value="firing">触发中</a-select-option>
            <a-select-option value="resolved">已解决</a-select-option>
            <a-select-option value="acknowledged">已确认</a-select-option>
          </a-select>
        </a-form-item>
        <a-form-item label="严重程度">
          <a-select
            v-model:value="searchForm.severity"
            placeholder="选择严重程度"
            allow-clear
            style="width: 120px"
          >
            <a-select-option value="critical">严重</a-select-option>
            <a-select-option value="warning">警告</a-select-option>
            <a-select-option value="info">信息</a-select-option>
          </a-select>
        </a-form-item>
        <a-form-item label="时间范围">
          <a-range-picker
            v-model:value="searchForm.timeRange"
            show-time
            format="YYYY-MM-DD HH:mm:ss"
            style="width: 300px"
          />
        </a-form-item>
        <a-form-item>
          <a-space>
            <a-button type="primary" html-type="submit">
              <template #icon><SearchOutlined /></template>
              搜索
            </a-button>
            <a-button @click="resetSearch">
              重置
            </a-button>
          </a-space>
        </a-form-item>
      </a-form>
    </a-card>

    <!-- 批量操作栏 -->
    <a-card v-if="showBatchActions" class="batch-actions">
      <a-space>
        <span>已选择 {{ selectedRowKeys.length }} 项</span>
        <a-button
          type="primary"
          :disabled="selectedRowKeys.length === 0"
          @click="batchAcknowledge"
        >
          批量确认
        </a-button>
        <a-button
          :disabled="selectedRowKeys.length === 0"
          @click="batchResolve"
        >
          批量解决
        </a-button>
        <a-button
          danger
          :disabled="selectedRowKeys.length === 0"
          @click="batchDelete"
        >
          批量删除
        </a-button>
      </a-space>
    </a-card>

    <!-- 告警列表 -->
    <a-card>
      <a-table
        :columns="columns"
        :data-source="alerts"
        :loading="loading"
        :pagination="pagination"
        :row-selection="rowSelection"
        :scroll="{ x: 1200 }"
        @change="handleTableChange"
      >
        <!-- 状态列 -->
        <template #status="{ record }">
          <a-tag :color="getStatusColor(record.status)" class="status-tag">
            {{ getStatusText(record.status) }}
          </a-tag>
        </template>

        <!-- 严重程度列 -->
        <template #severity="{ record }">
          <a-tag :color="getSeverityColor(record.severity)">
            {{ getSeverityText(record.severity) }}
          </a-tag>
        </template>

        <!-- 时间列 -->
        <template #created_at="{ record }">
          <a-tooltip :title="formatDateTime(record.created_at)">
            {{ getFriendlyTime(record.created_at) }}
          </a-tooltip>
        </template>

        <!-- 操作列 -->
        <template #action="{ record }">
          <a-space>
            <a-button type="link" size="small" @click="viewAlert(record)">
              查看
            </a-button>
            <a-button
              v-if="record.status === 'firing'"
              type="link"
              size="small"
              @click="acknowledgeAlert(record)"
            >
              确认
            </a-button>
            <a-button
              v-if="record.status !== 'resolved'"
              type="link"
              size="small"
              @click="resolveAlert(record)"
            >
              解决
            </a-button>
            <a-dropdown>
              <a-button type="link" size="small">
                更多 <DownOutlined />
              </a-button>
              <template #overlay>
                <a-menu>
                  <a-menu-item @click="analyzeAlert(record)">
                    <BulbOutlined /> AI分析
                  </a-menu-item>
                  <a-menu-item @click="convertToKnowledge(record)">
                    <BookOutlined /> 转为知识
                  </a-menu-item>
                  <a-menu-divider />
                  <a-menu-item @click="deleteAlert(record)" danger>
                    <DeleteOutlined /> 删除
                  </a-menu-item>
                </a-menu>
              </template>
            </a-dropdown>
          </a-space>
        </template>
      </a-table>
    </a-card>

    <!-- 告警详情抽屉 -->
    <a-drawer
      v-model:open="detailDrawerVisible"
      title="告警详情"
      width="600"
      placement="right"
    >
      <AlertDetail
        v-if="selectedAlert"
        :alert="selectedAlert"
        @update="handleAlertUpdate"
        @close="detailDrawerVisible = false"
      />
    </a-drawer>

    <!-- AI分析结果模态框 -->
    <a-modal
      v-model:open="analysisModalVisible"
      title="AI分析结果"
      width="800"
      :footer="null"
    >
      <AlertAnalysis
        v-if="analysisResult"
        :analysis="analysisResult"
        @close="analysisModalVisible = false"
      />
    </a-modal>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import {
  Card,
  Table,
  Button,
  Space,
  Tag,
  Form,
  Input,
  Select,
  DatePicker,
  Statistic,
  Row,
  Col,
  Tooltip,
  Drawer,
  Modal,
  Dropdown,
  Menu,
  message
} from 'ant-design-vue'
import {
  ReloadOutlined,
  SettingOutlined,
  AlertOutlined,
  ExclamationCircleOutlined,
  CheckCircleOutlined,
  CheckOutlined,
  SearchOutlined,
  DownOutlined,
  BulbOutlined,
  BookOutlined,
  DeleteOutlined
} from '@ant-design/icons-vue'
import { useAlertStore } from '@/stores'
import {
  getAlerts,
  getAlertStats,
  updateAlert,
  analyzeAlert as analyzeAlertApi,
  convertToKnowledge as convertToKnowledgeApi,
  batchUpdateAlerts
} from '@/services/alert'
import { formatDateTime, getFriendlyTime } from '@/utils/datetime'
import type { Alert, AlertAnalysis } from '@/types'
import AlertDetail from '@/components/AlertDetail.vue'
import AlertAnalysis from '@/components/AlertAnalysis.vue'

const ACard = Card
const ATable = Table
const AButton = Button
const ASpace = Space
const ATag = Tag
const AForm = Form
const AFormItem = Form.Item
const AInput = Input
const ASelect = Select
const ASelectOption = Select.Option
const ARangePicker = DatePicker.RangePicker
const AStatistic = Statistic
const ARow = Row
const ACol = Col
const ATooltip = Tooltip
const ADrawer = Drawer
const AModal = Modal
const ADropdown = Dropdown
const AMenu = Menu
const AMenuItem = Menu.Item
const AMenuDivider = Menu.Divider

const alertStore = useAlertStore()

// 响应式数据
const loading = ref(false)
const alerts = ref<Alert[]>([])
const alertStats = ref({
  total: 0,
  firing: 0,
  resolved: 0,
  acknowledged: 0
})
const selectedRowKeys = ref<number[]>([])
const showBatchActions = ref(false)
const detailDrawerVisible = ref(false)
const analysisModalVisible = ref(false)
const selectedAlert = ref<Alert | null>(null)
const analysisResult = ref<AlertAnalysis | null>(null)

// 搜索表单
const searchForm = reactive({
  search: '',
  status: undefined,
  severity: undefined,
  timeRange: undefined
})

// 分页配置
const pagination = reactive({
  current: 1,
  pageSize: 20,
  total: 0,
  showSizeChanger: true,
  showQuickJumper: true,
  showTotal: (total: number) => `共 ${total} 条记录`
})

// 表格列配置
const columns = [
  {
    title: '告警名称',
    dataIndex: 'name',
    key: 'name',
    width: 200,
    ellipsis: true
  },
  {
    title: '状态',
    dataIndex: 'status',
    key: 'status',
    width: 100,
    slots: { customRender: 'status' }
  },
  {
    title: '严重程度',
    dataIndex: 'severity',
    key: 'severity',
    width: 100,
    slots: { customRender: 'severity' }
  },
  {
    title: '描述',
    dataIndex: 'description',
    key: 'description',
    ellipsis: true
  },
  {
    title: '数据源',
    dataIndex: 'source',
    key: 'source',
    width: 120
  },
  {
    title: '创建时间',
    dataIndex: 'created_at',
    key: 'created_at',
    width: 150,
    slots: { customRender: 'created_at' }
  },
  {
    title: '操作',
    key: 'action',
    width: 200,
    fixed: 'right',
    slots: { customRender: 'action' }
  }
]

// 行选择配置
const rowSelection = computed(() => ({
  selectedRowKeys: selectedRowKeys.value,
  onChange: (keys: number[]) => {
    selectedRowKeys.value = keys
  }
}))

// 获取状态颜色
const getStatusColor = (status: string) => {
  const colorMap: Record<string, string> = {
    firing: 'red',
    resolved: 'green',
    acknowledged: 'blue'
  }
  return colorMap[status] || 'default'
}

// 获取状态文本
const getStatusText = (status: string) => {
  const textMap: Record<string, string> = {
    firing: '触发中',
    resolved: '已解决',
    acknowledged: '已确认'
  }
  return textMap[status] || status
}

// 获取严重程度颜色
const getSeverityColor = (severity: string) => {
  const colorMap: Record<string, string> = {
    critical: 'red',
    warning: 'orange',
    info: 'blue'
  }
  return colorMap[severity] || 'default'
}

// 获取严重程度文本
const getSeverityText = (severity: string) => {
  const textMap: Record<string, string> = {
    critical: '严重',
    warning: '警告',
    info: '信息'
  }
  return textMap[severity] || severity
}

// 加载数据
const loadData = async () => {
  loading.value = true
  try {
    const params = {
      page: pagination.current,
      page_size: pagination.pageSize,
      ...searchForm
    }
    
    if (searchForm.timeRange && searchForm.timeRange.length === 2) {
      params.start_time = searchForm.timeRange[0].format('YYYY-MM-DD HH:mm:ss')
      params.end_time = searchForm.timeRange[1].format('YYYY-MM-DD HH:mm:ss')
    }
    
    const response = await getAlerts(params)
    alerts.value = response.data.items
    pagination.total = response.data.total
    
    // 更新store
    alertStore.setAlerts(response.data.items)
  } catch (error) {
    message.error('加载告警数据失败')
  } finally {
    loading.value = false
  }
}

// 加载统计数据
const loadStats = async () => {
  try {
    const response = await getAlertStats()
    alertStats.value = response.data
    alertStore.setAlertStats(response.data)
  } catch (error) {
    message.error('加载统计数据失败')
  }
}

// 刷新数据
const refreshData = () => {
  loadData()
  loadStats()
}

// 搜索处理
const handleSearch = () => {
  pagination.current = 1
  loadData()
}

// 重置搜索
const resetSearch = () => {
  Object.assign(searchForm, {
    search: '',
    status: undefined,
    severity: undefined,
    timeRange: undefined
  })
  handleSearch()
}

// 表格变化处理
const handleTableChange = (pag: any) => {
  pagination.current = pag.current
  pagination.pageSize = pag.pageSize
  loadData()
}

// 查看告警详情
const viewAlert = (alert: Alert) => {
  selectedAlert.value = alert
  detailDrawerVisible.value = true
}

// 确认告警
const acknowledgeAlert = async (alert: Alert) => {
  try {
    await updateAlert(alert.id, { status: 'acknowledged' })
    message.success('告警已确认')
    refreshData()
  } catch (error) {
    message.error('确认告警失败')
  }
}

// 解决告警
const resolveAlert = async (alert: Alert) => {
  try {
    await updateAlert(alert.id, { status: 'resolved' })
    message.success('告警已解决')
    refreshData()
  } catch (error) {
    message.error('解决告警失败')
  }
}

// AI分析告警
const analyzeAlert = async (alert: Alert) => {
  try {
    loading.value = true
    const response = await analyzeAlertApi(alert.id)
    analysisResult.value = response.data
    analysisModalVisible.value = true
  } catch (error) {
    message.error('AI分析失败')
  } finally {
    loading.value = false
  }
}

// 转为知识
const convertToKnowledge = async (alert: Alert) => {
  try {
    await convertToKnowledgeApi(alert.id)
    message.success('已转为知识库条目')
  } catch (error) {
    message.error('转换失败')
  }
}

// 删除告警
const deleteAlert = async (alert: Alert) => {
  try {
    await updateAlert(alert.id, { deleted: true })
    message.success('告警已删除')
    refreshData()
  } catch (error) {
    message.error('删除告警失败')
  }
}

// 批量确认
const batchAcknowledge = async () => {
  try {
    await batchUpdateAlerts({
      ids: selectedRowKeys.value,
      status: 'acknowledged'
    })
    message.success(`已确认 ${selectedRowKeys.value.length} 个告警`)
    selectedRowKeys.value = []
    refreshData()
  } catch (error) {
    message.error('批量确认失败')
  }
}

// 批量解决
const batchResolve = async () => {
  try {
    await batchUpdateAlerts({
      ids: selectedRowKeys.value,
      status: 'resolved'
    })
    message.success(`已解决 ${selectedRowKeys.value.length} 个告警`)
    selectedRowKeys.value = []
    refreshData()
  } catch (error) {
    message.error('批量解决失败')
  }
}

// 批量删除
const batchDelete = async () => {
  try {
    await batchUpdateAlerts({
      ids: selectedRowKeys.value,
      deleted: true
    })
    message.success(`已删除 ${selectedRowKeys.value.length} 个告警`)
    selectedRowKeys.value = []
    refreshData()
  } catch (error) {
    message.error('批量删除失败')
  }
}

// 告警更新处理
const handleAlertUpdate = (updatedAlert: Alert) => {
  const index = alerts.value.findIndex(alert => alert.id === updatedAlert.id)
  if (index !== -1) {
    alerts.value[index] = updatedAlert
  }
  alertStore.updateAlert(updatedAlert.id, updatedAlert)
}

// 组件挂载时加载数据
onMounted(() => {
  refreshData()
})
</script>

<style scoped>
.alert-list {
  padding: 0;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 24px;
}

.header-left h2 {
  margin: 0 0 8px 0;
  font-size: 24px;
  font-weight: 600;
}

.header-left p {
  margin: 0;
  color: #666;
}

.stats-cards {
  margin-bottom: 24px;
}

.search-card {
  margin-bottom: 16px;
}

.batch-actions {
  margin-bottom: 16px;
  background-color: #f0f2f5;
}

.status-tag {
  font-weight: 500;
}

/* 响应式设计 */
@media (max-width: 768px) {
  .page-header {
    flex-direction: column;
    gap: 16px;
  }
  
  .stats-cards :deep(.ant-col) {
    margin-bottom: 16px;
  }
  
  .search-card :deep(.ant-form) {
    flex-direction: column;
  }
  
  .search-card :deep(.ant-form-item) {
    margin-bottom: 16px;
  }
}
</style>