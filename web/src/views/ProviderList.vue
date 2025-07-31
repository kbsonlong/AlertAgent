<template>
  <div class="provider-list">
    <!-- 页面标题和操作 -->
    <div class="page-header">
      <div class="header-content">
        <h2>数据源管理</h2>
        <p>管理监控数据源和连接配置</p>
      </div>
      <div class="header-actions">
        <a-space>
          <a-button @click="handleRefresh">
            <template #icon><ReloadOutlined /></template>
            刷新
          </a-button>
          <a-button type="primary" @click="handleCreate">
            <template #icon><PlusOutlined /></template>
            新建数据源
          </a-button>
        </a-space>
      </div>
    </div>

    <!-- 数据源统计 -->
    <div class="stats-cards">
      <a-row :gutter="16">
        <a-col :span="6">
          <a-card>
            <a-statistic
              title="总数据源"
              :value="stats.total"
              :value-style="{ color: '#1890ff' }"
            >
              <template #prefix>
                <DatabaseOutlined />
              </template>
            </a-statistic>
          </a-card>
        </a-col>
        <a-col :span="6">
          <a-card>
            <a-statistic
              title="在线"
              :value="stats.online"
              :value-style="{ color: '#52c41a' }"
            >
              <template #prefix>
                <CheckCircleOutlined />
              </template>
            </a-statistic>
          </a-card>
        </a-col>
        <a-col :span="6">
          <a-card>
            <a-statistic
              title="离线"
              :value="stats.offline"
              :value-style="{ color: '#ff4d4f' }"
            >
              <template #prefix>
                <CloseCircleOutlined />
              </template>
            </a-statistic>
          </a-card>
        </a-col>
        <a-col :span="6">
          <a-card>
            <a-statistic
              title="告警中"
              :value="stats.alerting"
              :value-style="{ color: '#faad14' }"
            >
              <template #prefix>
                <WarningOutlined />
              </template>
            </a-statistic>
          </a-card>
        </a-col>
      </a-row>
    </div>

    <!-- 搜索和筛选 -->
    <a-card class="search-card">
      <a-form layout="inline" :model="searchForm">
        <a-form-item label="名称">
          <a-input
            v-model:value="searchForm.name"
            placeholder="搜索数据源名称"
            style="width: 200px"
            @press-enter="handleSearch"
          >
            <template #prefix><SearchOutlined /></template>
          </a-input>
        </a-form-item>
        <a-form-item label="类型">
          <a-select
            v-model:value="searchForm.type"
            placeholder="请选择类型"
            style="width: 150px"
            allow-clear
          >
            <a-select-option value="prometheus">Prometheus</a-select-option>
            <a-select-option value="grafana">Grafana</a-select-option>
            <a-select-option value="alertmanager">AlertManager</a-select-option>
            <a-select-option value="elasticsearch">Elasticsearch</a-select-option>
            <a-select-option value="influxdb">InfluxDB</a-select-option>
          </a-select>
        </a-form-item>
        <a-form-item label="状态">
          <a-select
            v-model:value="searchForm.status"
            placeholder="请选择状态"
            style="width: 120px"
            allow-clear
          >
            <a-select-option value="online">在线</a-select-option>
            <a-select-option value="offline">离线</a-select-option>
            <a-select-option value="error">错误</a-select-option>
          </a-select>
        </a-form-item>
        <a-form-item>
          <a-space>
            <a-button type="primary" @click="handleSearch">
              <template #icon><SearchOutlined /></template>
              搜索
            </a-button>
            <a-button @click="handleReset">
              <template #icon><ReloadOutlined /></template>
              重置
            </a-button>
          </a-space>
        </a-form-item>
      </a-form>
    </a-card>

    <!-- 批量操作 -->
    <div class="batch-actions" v-if="selectedRowKeys.length > 0">
      <a-alert
        :message="`已选择 ${selectedRowKeys.length} 项`"
        type="info"
        show-icon
      >
        <template #action>
          <a-space>
            <a-button size="small" @click="handleBatchTest">
              批量测试
            </a-button>
            <a-button size="small" @click="handleBatchSync">
              批量同步
            </a-button>
            <a-button size="small" danger @click="handleBatchDelete">
              批量删除
            </a-button>
          </a-space>
        </template>
      </a-alert>
    </div>

    <!-- 数据源列表 -->
    <a-card class="table-card">
      <a-table
        :columns="columns"
        :data-source="providerList"
        :loading="loading"
        :pagination="pagination"
        :row-selection="rowSelection"
        @change="handleTableChange"
        row-key="id"
      >
        <template #bodyCell="{ column, record }">
          <template v-if="column.key === 'name'">
            <div class="provider-name">
              <div class="name-content">
                <component :is="getTypeIcon(record.type)" class="type-icon" />
                <div class="name-info">
                  <a @click="handleView(record)" class="name-link">
                    {{ record.name }}
                  </a>
                  <div class="name-meta">
                    {{ record.url }}
                  </div>
                </div>
              </div>
            </div>
          </template>
          
          <template v-else-if="column.key === 'type'">
            <a-tag :color="getTypeColor(record.type)">
              {{ getTypeText(record.type) }}
            </a-tag>
          </template>
          
          <template v-else-if="column.key === 'status'">
            <div class="status-info">
              <a-badge
                :status="getStatusBadge(record.status)"
                :text="getStatusText(record.status)"
              />
              <div class="status-meta" v-if="record.lastCheck">
                {{ getRelativeTime(record.lastCheck) }}
              </div>
            </div>
          </template>
          
          <template v-else-if="column.key === 'health'">
            <div class="health-info">
              <a-progress
                :percent="record.health || 0"
                size="small"
                :stroke-color="getHealthColor(record.health)"
              />
              <div class="health-meta">
                {{ record.health || 0 }}%
              </div>
            </div>
          </template>
          
          <template v-else-if="column.key === 'metrics'">
            <div class="metrics-info">
              <div class="metric-item">
                <span class="metric-label">指标:</span>
                <span class="metric-value">{{ record.metricsCount || 0 }}</span>
              </div>
              <div class="metric-item">
                <span class="metric-label">规则:</span>
                <span class="metric-value">{{ record.rulesCount || 0 }}</span>
              </div>
            </div>
          </template>
          
          <template v-else-if="column.key === 'updatedAt'">
            <div class="time-info">
              <div>{{ formatDateTime(record.updatedAt) }}</div>
              <div class="time-relative">{{ getRelativeTime(record.updatedAt) }}</div>
            </div>
          </template>
          
          <template v-else-if="column.key === 'action'">
            <a-space>
              <a-button type="link" size="small" @click="handleView(record)">
                查看
              </a-button>
              <a-button type="link" size="small" @click="handleTest(record)">
                测试
              </a-button>
              <a-dropdown>
                <template #overlay>
                  <a-menu>
                    <a-menu-item @click="handleEdit(record)">
                      <EditOutlined /> 编辑
                    </a-menu-item>
                    <a-menu-item @click="handleSync(record)">
                      <SyncOutlined /> 同步
                    </a-menu-item>
                    <a-menu-item @click="handleViewMetrics(record)">
                      <BarChartOutlined /> 查看指标
                    </a-menu-item>
                    <a-menu-item @click="handleViewHealth(record)">
                      <HeartOutlined /> 健康状态
                    </a-menu-item>
                    <a-menu-divider />
                    <a-menu-item @click="handleDelete(record)" danger>
                      <DeleteOutlined /> 删除
                    </a-menu-item>
                  </a-menu>
                </template>
                <a-button type="link" size="small">
                  更多 <DownOutlined />
                </a-button>
              </a-dropdown>
            </a-space>
          </template>
        </template>
      </a-table>
    </a-card>

    <!-- 数据源详情抽屉 -->
    <a-drawer
      v-model:open="detailVisible"
      title="数据源详情"
      width="800"
      :footer-style="{ textAlign: 'right' }"
    >
      <ProviderDetail
        v-if="currentProvider"
        :provider="currentProvider"
        @edit="handleEdit"
        @test="handleTest"
        @delete="handleDelete"
        @close="detailVisible = false"
      />
      <template #footer>
        <a-space>
          <a-button @click="detailVisible = false">关闭</a-button>
          <a-button @click="handleTest(currentProvider)">
            测试连接
          </a-button>
          <a-button type="primary" @click="handleEdit(currentProvider)">
            编辑
          </a-button>
        </a-space>
      </template>
    </a-drawer>

    <!-- 数据源表单模态框 -->
    <a-modal
      v-model:open="formVisible"
      :title="formMode === 'create' ? '新建数据源' : '编辑数据源'"
      width="800"
      :footer="null"
      :destroy-on-close="true"
    >
      <ProviderForm
        v-if="formVisible"
        :provider="currentProvider"
        :mode="formMode"
        @submit="handleFormSubmit"
        @cancel="formVisible = false"
      />
    </a-modal>

    <!-- 测试结果模态框 -->
    <a-modal
      v-model:open="testVisible"
      title="连接测试结果"
      :footer="null"
      width="600"
    >
      <div class="test-result">
        <div class="test-header">
          <a-result
            :status="testResult.success ? 'success' : 'error'"
            :title="testResult.success ? '连接成功' : '连接失败'"
            :sub-title="testResult.message"
          />
        </div>
        
        <div class="test-details" v-if="testResult.details">
          <h4>详细信息</h4>
          <a-descriptions :column="1" bordered size="small">
            <a-descriptions-item
              v-for="(value, key) in testResult.details"
              :key="key"
              :label="key"
            >
              {{ value }}
            </a-descriptions-item>
          </a-descriptions>
        </div>
        
        <div class="test-actions">
          <a-space>
            <a-button @click="testVisible = false">关闭</a-button>
            <a-button type="primary" @click="handleRetryTest" v-if="!testResult.success">
              重新测试
            </a-button>
          </a-space>
        </div>
      </div>
    </a-modal>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import {
  Card,
  Row,
  Col,
  Statistic,
  Form,
  Input,
  Select,
  Button,
  Space,
  Table,
  Tag,
  Badge,
  Progress,
  Alert,
  Drawer,
  Modal,
  Dropdown,
  Menu,
  Result,
  Descriptions,
  message
} from 'ant-design-vue'
import {
  DatabaseOutlined,
  CheckCircleOutlined,
  CloseCircleOutlined,
  WarningOutlined,
  SearchOutlined,
  ReloadOutlined,
  PlusOutlined,
  EditOutlined,
  SyncOutlined,
  BarChartOutlined,
  HeartOutlined,
  DeleteOutlined,
  DownOutlined
} from '@ant-design/icons-vue'
import { formatDateTime, getRelativeTime } from '@/utils/datetime'
import {
  getProviders,
  getProvider,
  createProvider,
  updateProvider,
  deleteProvider,
  testProvider,
  getProviderHealth,
  syncProvider,
  batchDeleteProviders
} from '@/services/provider'
import type { Provider } from '@/types'
import ProviderDetail from '@/components/ProviderDetail.vue'
import ProviderForm from '@/components/ProviderForm.vue'

const ACard = Card
const ARow = Row
const ACol = Col
const AStatistic = Statistic
const AForm = Form
const AFormItem = Form.Item
const AInput = Input
const ASelect = Select
const ASelectOption = Select.Option
const AButton = Button
const ASpace = Space
const ATable = Table
const ATag = Tag
const ABadge = Badge
const AProgress = Progress
const AAlert = Alert
const ADrawer = Drawer
const AModal = Modal
const ADropdown = Dropdown
const AMenu = Menu
const AMenuItem = Menu.Item
const AMenuDivider = Menu.Divider
const AResult = Result
const ADescriptions = Descriptions
const ADescriptionsItem = Descriptions.Item

// 响应式数据
const loading = ref(false)
const providerList = ref<Provider[]>([])
const selectedRowKeys = ref<string[]>([])

// 统计数据
const stats = reactive({
  total: 0,
  online: 0,
  offline: 0,
  alerting: 0
})

// 搜索表单
const searchForm = reactive({
  name: '',
  type: undefined,
  status: undefined
})

// 分页配置
const pagination = reactive({
  current: 1,
  pageSize: 10,
  total: 0,
  showSizeChanger: true,
  showQuickJumper: true,
  showTotal: (total: number) => `共 ${total} 条记录`
})

// 表格配置
const columns = [
  {
    title: '名称',
    key: 'name',
    width: 250,
    ellipsis: true
  },
  {
    title: '类型',
    key: 'type',
    width: 120
  },
  {
    title: '状态',
    key: 'status',
    width: 120
  },
  {
    title: '健康度',
    key: 'health',
    width: 120
  },
  {
    title: '指标/规则',
    key: 'metrics',
    width: 120
  },
  {
    title: '更新时间',
    key: 'updatedAt',
    width: 180,
    sorter: true
  },
  {
    title: '操作',
    key: 'action',
    width: 200,
    fixed: 'right'
  }
]

// 行选择配置
const rowSelection = {
  selectedRowKeys,
  onChange: (keys: string[]) => {
    selectedRowKeys.value = keys
  }
}

// 详情抽屉
const detailVisible = ref(false)
const currentProvider = ref<Provider | null>(null)

// 表单模态框
const formVisible = ref(false)
const formMode = ref<'create' | 'edit'>('create')

// 测试结果模态框
const testVisible = ref(false)
const testResult = ref({
  success: false,
  message: '',
  details: null
})

// 获取类型图标
const getTypeIcon = (type: string) => {
  const iconMap: Record<string, any> = {
    prometheus: DatabaseOutlined,
    grafana: BarChartOutlined,
    alertmanager: WarningOutlined,
    elasticsearch: SearchOutlined,
    influxdb: DatabaseOutlined
  }
  return iconMap[type] || DatabaseOutlined
}

// 获取类型颜色
const getTypeColor = (type: string) => {
  const colorMap: Record<string, string> = {
    prometheus: 'orange',
    grafana: 'blue',
    alertmanager: 'red',
    elasticsearch: 'green',
    influxdb: 'purple'
  }
  return colorMap[type] || 'default'
}

// 获取类型文本
const getTypeText = (type: string) => {
  const textMap: Record<string, string> = {
    prometheus: 'Prometheus',
    grafana: 'Grafana',
    alertmanager: 'AlertManager',
    elasticsearch: 'Elasticsearch',
    influxdb: 'InfluxDB'
  }
  return textMap[type] || type
}

// 获取状态徽章
const getStatusBadge = (status: string) => {
  const badgeMap: Record<string, string> = {
    online: 'success',
    offline: 'error',
    error: 'warning'
  }
  return badgeMap[status] || 'default'
}

// 获取状态文本
const getStatusText = (status: string) => {
  const textMap: Record<string, string> = {
    online: '在线',
    offline: '离线',
    error: '错误'
  }
  return textMap[status] || status
}

// 获取健康度颜色
const getHealthColor = (health: number) => {
  if (health >= 80) return '#52c41a'
  if (health >= 60) return '#faad14'
  return '#ff4d4f'
}

// 加载数据
const loadData = async () => {
  loading.value = true
  try {
    const params = {
      page: pagination.current,
      pageSize: pagination.pageSize,
      name: searchForm.name,
      type: searchForm.type,
      status: searchForm.status
    }
    
    const response = await getProviders(params)
    providerList.value = response.data.list
    pagination.total = response.data.total
    
    // 更新统计数据
    stats.total = response.data.stats.total
    stats.online = response.data.stats.online
    stats.offline = response.data.stats.offline
    stats.alerting = response.data.stats.alerting
  } catch (error) {
    console.error('加载数据源列表失败:', error)
    message.error('加载数据源列表失败')
  } finally {
    loading.value = false
  }
}

// 搜索
const handleSearch = () => {
  pagination.current = 1
  loadData()
}

// 重置
const handleReset = () => {
  Object.assign(searchForm, {
    name: '',
    type: undefined,
    status: undefined
  })
  pagination.current = 1
  loadData()
}

// 刷新
const handleRefresh = () => {
  loadData()
}

// 表格变化
const handleTableChange = (pag: any, filters: any, sorter: any) => {
  pagination.current = pag.current
  pagination.pageSize = pag.pageSize
  loadData()
}

// 查看详情
const handleView = async (provider: Provider) => {
  try {
    const response = await getProvider(provider.id)
    currentProvider.value = response.data
    detailVisible.value = true
  } catch (error) {
    console.error('获取数据源详情失败:', error)
    message.error('获取数据源详情失败')
  }
}

// 新建
const handleCreate = () => {
  currentProvider.value = null
  formMode.value = 'create'
  formVisible.value = true
}

// 编辑
const handleEdit = (provider: Provider) => {
  currentProvider.value = provider
  formMode.value = 'edit'
  formVisible.value = true
  detailVisible.value = false
}

// 表单提交
const handleFormSubmit = async (data: any) => {
  try {
    if (formMode.value === 'create') {
      await createProvider(data)
      message.success('创建成功')
    } else {
      await updateProvider(currentProvider.value!.id, data)
      message.success('更新成功')
    }
    
    formVisible.value = false
    loadData()
  } catch (error) {
    console.error('保存失败:', error)
    message.error('保存失败')
  }
}

// 测试连接
const handleTest = async (provider: Provider) => {
  try {
    const response = await testProvider(provider.id)
    testResult.value = {
      success: response.data.success,
      message: response.data.message,
      details: response.data.details
    }
    testVisible.value = true
  } catch (error) {
    console.error('测试连接失败:', error)
    testResult.value = {
      success: false,
      message: '测试连接失败',
      details: null
    }
    testVisible.value = true
  }
}

// 重新测试
const handleRetryTest = () => {
  testVisible.value = false
  if (currentProvider.value) {
    handleTest(currentProvider.value)
  }
}

// 同步
const handleSync = async (provider: Provider) => {
  try {
    await syncProvider(provider.id)
    message.success('同步成功')
    loadData()
  } catch (error) {
    console.error('同步失败:', error)
    message.error('同步失败')
  }
}

// 查看指标
const handleViewMetrics = (provider: Provider) => {
  message.info(`查看 ${provider.name} 的指标`)
}

// 查看健康状态
const handleViewHealth = async (provider: Provider) => {
  try {
    const response = await getProviderHealth(provider.id)
    message.info(`健康度: ${response.data.health}%`)
  } catch (error) {
    console.error('获取健康状态失败:', error)
    message.error('获取健康状态失败')
  }
}

// 删除
const handleDelete = async (provider: Provider) => {
  try {
    await deleteProvider(provider.id)
    message.success('删除成功')
    detailVisible.value = false
    loadData()
  } catch (error) {
    console.error('删除失败:', error)
    message.error('删除失败')
  }
}

// 批量测试
const handleBatchTest = async () => {
  try {
    // 这里应该有批量测试的API
    message.success('批量测试完成')
    selectedRowKeys.value = []
    loadData()
  } catch (error) {
    console.error('批量测试失败:', error)
    message.error('批量测试失败')
  }
}

// 批量同步
const handleBatchSync = async () => {
  try {
    // 这里应该有批量同步的API
    message.success('批量同步完成')
    selectedRowKeys.value = []
    loadData()
  } catch (error) {
    console.error('批量同步失败:', error)
    message.error('批量同步失败')
  }
}

// 批量删除
const handleBatchDelete = async () => {
  try {
    await batchDeleteProviders(selectedRowKeys.value)
    message.success('批量删除成功')
    selectedRowKeys.value = []
    loadData()
  } catch (error) {
    console.error('批量删除失败:', error)
    message.error('批量删除失败')
  }
}

// 组件挂载
onMounted(() => {
  loadData()
})
</script>

<style scoped>
.provider-list {
  padding: 24px;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 24px;
}

.header-content h2 {
  margin: 0 0 8px 0;
  font-size: 24px;
  font-weight: 600;
}

.header-content p {
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
}

.table-card {
  margin-bottom: 0;
}

.provider-name {
  display: flex;
  align-items: center;
}

.name-content {
  display: flex;
  align-items: center;
  gap: 12px;
}

.type-icon {
  font-size: 20px;
  color: #1890ff;
}

.name-info {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.name-link {
  font-weight: 500;
  color: #1890ff;
  text-decoration: none;
}

.name-link:hover {
  text-decoration: underline;
}

.name-meta {
  font-size: 12px;
  color: #999;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  max-width: 200px;
}

.status-info {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.status-meta {
  font-size: 12px;
  color: #999;
}

.health-info {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.health-meta {
  font-size: 12px;
  color: #666;
  text-align: center;
}

.metrics-info {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.metric-item {
  display: flex;
  justify-content: space-between;
  font-size: 12px;
}

.metric-label {
  color: #999;
}

.metric-value {
  color: #666;
  font-weight: 500;
}

.time-info {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.time-relative {
  font-size: 12px;
  color: #999;
}

.test-result {
  padding: 16px 0;
}

.test-header {
  text-align: center;
  margin-bottom: 24px;
}

.test-details {
  margin-bottom: 24px;
}

.test-details h4 {
  margin: 0 0 16px 0;
  font-size: 16px;
  font-weight: 600;
}

.test-actions {
  text-align: right;
}

/* 响应式设计 */
@media (max-width: 768px) {
  .provider-list {
    padding: 16px;
  }
  
  .page-header {
    flex-direction: column;
    gap: 16px;
  }
  
  .header-actions {
    width: 100%;
  }
  
  .stats-cards :deep(.ant-col) {
    margin-bottom: 16px;
  }
  
  .search-card :deep(.ant-form-item) {
    margin-bottom: 16px;
  }
  
  .table-card :deep(.ant-table) {
    font-size: 12px;
  }
  
  .name-meta {
    max-width: 100px;
  }
}
</style>