<template>
  <div class="user-notification-manager">
    <div class="manager-header">
      <h3>
        <UserOutlined />
        用户通知管理
      </h3>
      <div class="header-actions">
        <a-button @click="exportPreferences" :loading="exportLoading">
          <DownloadOutlined /> 导出配置
        </a-button>
        <a-button @click="importPreferences">
          <UploadOutlined /> 导入配置
        </a-button>
        <a-button type="primary" @click="savePreferences" :loading="saveLoading">
          <SaveOutlined /> 保存设置
        </a-button>
      </div>
    </div>

    <div class="manager-content">
      <a-tabs v-model:activeKey="activeTab">
        <!-- 通知偏好设置 -->
        <a-tab-pane key="preferences" tab="通知偏好">
          <div class="preferences-section">
            <div class="global-settings">
              <h4>全局设置</h4>
              <a-row :gutter="16">
                <a-col :span="8">
                  <a-form-item label="启用通知">
                    <a-switch
                      v-model:checked="preferences.enabled"
                      checked-children="启用"
                      un-checked-children="禁用"
                    />
                  </a-form-item>
                </a-col>
                <a-col :span="8">
                  <a-form-item label="静默时间">
                    <a-time-range-picker
                      v-model:value="preferences.quietHours"
                      format="HH:mm"
                      placeholder="['开始时间', '结束时间']"
                    />
                  </a-form-item>
                </a-col>
                <a-col :span="8">
                  <a-form-item label="最大频率">
                    <a-select v-model:value="preferences.maxFrequency">
                      <a-select-option value="immediate">立即</a-select-option>
                      <a-select-option value="5min">5分钟</a-select-option>
                      <a-select-option value="15min">15分钟</a-select-option>
                      <a-select-option value="30min">30分钟</a-select-option>
                      <a-select-option value="1hour">1小时</a-select-option>
                    </a-select>
                  </a-form-item>
                </a-col>
              </a-row>
            </div>

            <div class="plugin-preferences">
              <h4>插件通知设置</h4>
              <div class="plugin-list">
                <div
                  v-for="plugin in availablePlugins"
                  :key="plugin.name"
                  class="plugin-preference-item"
                >
                  <div class="plugin-header">
                    <div class="plugin-info">
                      <component :is="getPluginIcon(plugin.name)" class="plugin-icon" />
                      <span class="plugin-name">{{ getPluginDisplayName(plugin.name) }}</span>
                    </div>
                    <a-switch
                      v-model:checked="preferences.plugins[plugin.name]?.enabled"
                      @change="(checked) => updatePluginPreference(plugin.name, 'enabled', checked)"
                    />
                  </div>
                  
                  <div v-if="preferences.plugins[plugin.name]?.enabled" class="plugin-settings">
                    <a-row :gutter="16">
                      <a-col :span="8">
                        <a-form-item label="优先级">
                          <a-select
                            :value="preferences.plugins[plugin.name]?.priority || 'normal'"
                            @update:value="(value) => updatePluginPreference(plugin.name, 'priority', value)"
                          >
                            <a-select-option value="high">高</a-select-option>
                            <a-select-option value="normal">普通</a-select-option>
                            <a-select-option value="low">低</a-select-option>
                          </a-select>
                        </a-form-item>
                      </a-col>
                      <a-col :span="8">
                        <a-form-item label="告警级别">
                          <a-select
                            :value="preferences.plugins[plugin.name]?.severityFilter || []"
                            mode="multiple"
                            placeholder="选择要接收的告警级别"
                            @update:value="(value) => updatePluginPreference(plugin.name, 'severityFilter', value)"
                          >
                            <a-select-option value="critical">严重</a-select-option>
                            <a-select-option value="error">错误</a-select-option>
                            <a-select-option value="warning">警告</a-select-option>
                            <a-select-option value="info">信息</a-select-option>
                          </a-select>
                        </a-form-item>
                      </a-col>
                      <a-col :span="8">
                        <a-form-item label="时间过滤">
                          <a-checkbox-group
                            :value="preferences.plugins[plugin.name]?.timeFilter || []"
                            @update:value="(value) => updatePluginPreference(plugin.name, 'timeFilter', value)"
                          >
                            <a-checkbox value="workdays">工作日</a-checkbox>
                            <a-checkbox value="weekends">周末</a-checkbox>
                            <a-checkbox value="holidays">节假日</a-checkbox>
                          </a-checkbox-group>
                        </a-form-item>
                      </a-col>
                    </a-row>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </a-tab-pane>

        <!-- 通知历史 -->
        <a-tab-pane key="history" tab="通知历史">
          <div class="history-section">
            <div class="history-filters">
              <a-row :gutter="16">
                <a-col :span="6">
                  <a-select
                    v-model:value="historyFilters.plugin"
                    placeholder="选择插件"
                    allow-clear
                  >
                    <a-select-option
                      v-for="plugin in availablePlugins"
                      :key="plugin.name"
                      :value="plugin.name"
                    >
                      {{ getPluginDisplayName(plugin.name) }}
                    </a-select-option>
                  </a-select>
                </a-col>
                <a-col :span="6">
                  <a-select
                    v-model:value="historyFilters.status"
                    placeholder="选择状态"
                    allow-clear
                  >
                    <a-select-option value="success">成功</a-select-option>
                    <a-select-option value="failed">失败</a-select-option>
                  </a-select>
                </a-col>
                <a-col :span="6">
                  <a-range-picker
                    v-model:value="historyFilters.dateRange"
                    placeholder="['开始日期', '结束日期']"
                  />
                </a-col>
                <a-col :span="6">
                  <a-button type="primary" @click="loadNotificationHistory">
                    <SearchOutlined /> 查询
                  </a-button>
                </a-col>
              </a-row>
            </div>

            <div class="history-table">
              <a-table
                :columns="historyColumns"
                :data-source="notificationHistory"
                :loading="historyLoading"
                :pagination="historyPagination"
                @change="handleHistoryTableChange"
              >
                <template #bodyCell="{ column, record }">
                  <template v-if="column.key === 'plugin'">
                    <a-tag :color="getPluginTagColor(record.plugin)">
                      {{ getPluginDisplayName(record.plugin) }}
                    </a-tag>
                  </template>
                  <template v-else-if="column.key === 'status'">
                    <a-tag :color="record.status === 'success' ? 'green' : 'red'">
                      {{ record.status === 'success' ? '成功' : '失败' }}
                    </a-tag>
                  </template>
                  <template v-else-if="column.key === 'sent_at'">
                    {{ formatTime(record.sent_at) }}
                  </template>
                  <template v-else-if="column.key === 'actions'">
                    <a-space>
                      <a-button size="small" @click="viewNotificationDetail(record)">
                        <EyeOutlined /> 详情
                      </a-button>
                      <a-button
                        v-if="record.status === 'failed'"
                        size="small"
                        type="primary"
                        @click="retryNotification(record)"
                      >
                        <RedoOutlined /> 重试
                      </a-button>
                    </a-space>
                  </template>
                </template>
              </a-table>
            </div>
          </div>
        </a-tab-pane>

        <!-- 统计分析 -->
        <a-tab-pane key="statistics" tab="统计分析">
          <div class="statistics-section">
            <div class="stats-overview">
              <a-row :gutter="16">
                <a-col :span="6">
                  <a-statistic
                    title="总通知数"
                    :value="statistics.total"
                    :value-style="{ color: '#1890ff' }"
                  />
                </a-col>
                <a-col :span="6">
                  <a-statistic
                    title="成功数"
                    :value="statistics.success"
                    :value-style="{ color: '#52c41a' }"
                  />
                </a-col>
                <a-col :span="6">
                  <a-statistic
                    title="失败数"
                    :value="statistics.failed"
                    :value-style="{ color: '#ff4d4f' }"
                  />
                </a-col>
                <a-col :span="6">
                  <a-statistic
                    title="成功率"
                    :value="statistics.successRate"
                    suffix="%"
                    :precision="1"
                    :value-style="{ color: getSuccessRateColor() }"
                  />
                </a-col>
              </a-row>
            </div>

            <div class="stats-charts">
              <a-row :gutter="16">
                <a-col :span="12">
                  <div class="chart-container">
                    <h4>插件使用分布</h4>
                    <div class="chart-placeholder">
                      <!-- 这里可以集成图表库如ECharts -->
                      <a-empty description="图表功能开发中" />
                    </div>
                  </div>
                </a-col>
                <a-col :span="12">
                  <div class="chart-container">
                    <h4>通知趋势</h4>
                    <div class="chart-placeholder">
                      <a-empty description="图表功能开发中" />
                    </div>
                  </div>
                </a-col>
              </a-row>
            </div>
          </div>
        </a-tab-pane>
      </a-tabs>
    </div>

    <!-- 通知详情模态框 -->
    <a-modal
      v-model:open="detailModalVisible"
      title="通知详情"
      width="600px"
      :footer="null"
    >
      <div v-if="selectedNotification" class="notification-detail">
        <a-descriptions :column="1" bordered>
          <a-descriptions-item label="通知ID">
            {{ selectedNotification.id }}
          </a-descriptions-item>
          <a-descriptions-item label="插件">
            <a-tag :color="getPluginTagColor(selectedNotification.plugin)">
              {{ getPluginDisplayName(selectedNotification.plugin) }}
            </a-tag>
          </a-descriptions-item>
          <a-descriptions-item label="状态">
            <a-tag :color="selectedNotification.status === 'success' ? 'green' : 'red'">
              {{ selectedNotification.status === 'success' ? '成功' : '失败' }}
            </a-tag>
          </a-descriptions-item>
          <a-descriptions-item label="发送时间">
            {{ formatTime(selectedNotification.sent_at) }}
          </a-descriptions-item>
          <a-descriptions-item label="消息标题">
            {{ selectedNotification.title }}
          </a-descriptions-item>
          <a-descriptions-item label="消息内容">
            <div class="message-content">{{ selectedNotification.content }}</div>
          </a-descriptions-item>
          <a-descriptions-item v-if="selectedNotification.error" label="错误信息">
            <a-typography-text type="danger">
              {{ selectedNotification.error }}
            </a-typography-text>
          </a-descriptions-item>
        </a-descriptions>
      </div>
    </a-modal>

    <!-- 导入配置模态框 -->
    <a-modal
      v-model:open="importModalVisible"
      title="导入通知配置"
      width="600px"
      @ok="handleImport"
      @cancel="importModalVisible = false"
    >
      <div class="import-content">
        <a-upload
          :before-upload="handleImportFile"
          :show-upload-list="false"
          accept=".json"
        >
          <a-button>
            <UploadOutlined /> 选择配置文件
          </a-button>
        </a-upload>
        <div class="import-textarea-section">
          <p>或者直接粘贴配置内容：</p>
          <a-textarea
            v-model:value="importContent"
            placeholder="请粘贴JSON格式的配置内容"
            :rows="8"
          />
        </div>
      </div>
    </a-modal>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { message } from 'ant-design-vue'
import {
  Button,
  Tabs,
  Form,
  Switch,
  Select,
  Checkbox,
  TimePicker,
  DatePicker,
  Table,
  Tag,
  Space,
  Statistic,
  Row,
  Col,
  Modal,
  Descriptions,
  Typography,
  Upload,
  Textarea,
  Empty
} from 'ant-design-vue'
import {
  UserOutlined,
  DownloadOutlined,
  UploadOutlined,
  SaveOutlined,
  SearchOutlined,
  EyeOutlined,
  RedoOutlined,
  MailOutlined,
  ApiOutlined,
  MessageOutlined,
  WechatOutlined,
  SlackOutlined,
  BellOutlined
} from '@ant-design/icons-vue'
import { getAvailablePlugins, type PluginInfo } from '@/services/plugin'
import { formatTime } from '@/utils/datetime'

const AButton = Button
const ATabs = Tabs
const ATabPane = Tabs.TabPane
const AFormItem = Form.Item
const ASwitch = Switch
const ASelect = Select
const ASelectOption = Select.Option
const ACheckbox = Checkbox
const ACheckboxGroup = Checkbox.Group
const ATimeRangePicker = TimePicker.RangePicker
const ARangePicker = DatePicker.RangePicker
const ATable = Table
const ATag = Tag
const ASpace = Space
const AStatistic = Statistic
const ARow = Row
const ACol = Col
const AModal = Modal
const ADescriptions = Descriptions
const ADescriptionsItem = Descriptions.Item
const ATypographyText = Typography.Text
const AUpload = Upload
const ATextarea = Textarea
const AEmpty = Empty

// 响应式数据
const activeTab = ref('preferences')
const saveLoading = ref(false)
const exportLoading = ref(false)
const historyLoading = ref(false)
const detailModalVisible = ref(false)
const importModalVisible = ref(false)
const importContent = ref('')
const availablePlugins = ref<PluginInfo[]>([])
const selectedNotification = ref<any>(null)

// 用户通知偏好
const preferences = reactive({
  enabled: true,
  quietHours: [] as any[],
  maxFrequency: 'immediate',
  plugins: {} as Record<string, {
    enabled: boolean
    priority: string
    severityFilter: string[]
    timeFilter: string[]
  }>
})

// 历史记录过滤器
const historyFilters = reactive({
  plugin: undefined,
  status: undefined,
  dateRange: [] as any[]
})

// 通知历史数据
const notificationHistory = ref<any[]>([])
const historyPagination = reactive({
  current: 1,
  pageSize: 10,
  total: 0
})

// 统计数据
const statistics = reactive({
  total: 0,
  success: 0,
  failed: 0,
  successRate: 0
})

// 历史记录表格列
const historyColumns = [
  {
    title: '通知ID',
    dataIndex: 'id',
    key: 'id',
    width: 120
  },
  {
    title: '插件',
    dataIndex: 'plugin',
    key: 'plugin',
    width: 100
  },
  {
    title: '状态',
    dataIndex: 'status',
    key: 'status',
    width: 80
  },
  {
    title: '标题',
    dataIndex: 'title',
    key: 'title',
    ellipsis: true
  },
  {
    title: '发送时间',
    dataIndex: 'sent_at',
    key: 'sent_at',
    width: 180
  },
  {
    title: '操作',
    key: 'actions',
    width: 120
  }
]

// 获取插件图标
const getPluginIcon = (pluginName: string) => {
  const iconMap: Record<string, any> = {
    'email': MailOutlined,
    'dingtalk': MessageOutlined,
    'wechat': WechatOutlined,
    'slack': SlackOutlined,
    'webhook': ApiOutlined
  }
  return iconMap[pluginName] || BellOutlined
}

// 获取插件显示名称
const getPluginDisplayName = (pluginName: string) => {
  const nameMap: Record<string, string> = {
    'email': '邮件通知',
    'dingtalk': '钉钉通知',
    'wechat': '企业微信',
    'slack': 'Slack',
    'webhook': 'Webhook'
  }
  return nameMap[pluginName] || pluginName
}

// 获取插件标签颜色
const getPluginTagColor = (pluginName: string) => {
  const colorMap: Record<string, string> = {
    'email': 'blue',
    'dingtalk': 'cyan',
    'wechat': 'green',
    'slack': 'purple',
    'webhook': 'orange'
  }
  return colorMap[pluginName] || 'default'
}

// 获取成功率颜色
const getSuccessRateColor = () => {
  if (statistics.successRate >= 95) return '#52c41a'
  if (statistics.successRate >= 80) return '#faad14'
  return '#ff4d4f'
}

// 更新插件偏好
const updatePluginPreference = (pluginName: string, key: string, value: any) => {
  if (!preferences.plugins[pluginName]) {
    preferences.plugins[pluginName] = {
      enabled: false,
      priority: 'normal',
      severityFilter: [],
      timeFilter: []
    }
  }
  preferences.plugins[pluginName][key as keyof typeof preferences.plugins[string]] = value
}

// 加载可用插件
const loadAvailablePlugins = async () => {
  try {
    const plugins = await getAvailablePlugins()
    availablePlugins.value = plugins
    
    // 初始化插件偏好
    plugins.forEach(plugin => {
      if (!preferences.plugins[plugin.name]) {
        preferences.plugins[plugin.name] = {
          enabled: true,
          priority: 'normal',
          severityFilter: ['critical', 'error', 'warning'],
          timeFilter: ['workdays']
        }
      }
    })
  } catch (error) {
    console.error('加载插件失败:', error)
    message.error('加载插件失败')
  }
}

// 保存偏好设置
const savePreferences = async () => {
  try {
    saveLoading.value = true
    
    // 这里应该调用API保存用户偏好
    // 暂时模拟保存过程
    await new Promise(resolve => setTimeout(resolve, 1000))
    
    message.success('通知偏好已保存')
  } catch (error) {
    console.error('保存偏好失败:', error)
    message.error('保存偏好失败')
  } finally {
    saveLoading.value = false
  }
}

// 导出偏好配置
const exportPreferences = async () => {
  try {
    exportLoading.value = true
    
    const config = {
      version: '1.0.0',
      timestamp: new Date().toISOString(),
      preferences: preferences
    }
    
    const blob = new Blob([JSON.stringify(config, null, 2)], { type: 'application/json' })
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = `notification-preferences-${new Date().toISOString().split('T')[0]}.json`
    document.body.appendChild(a)
    a.click()
    document.body.removeChild(a)
    URL.revokeObjectURL(url)
    
    message.success('配置已导出')
  } catch (error) {
    console.error('导出配置失败:', error)
    message.error('导出配置失败')
  } finally {
    exportLoading.value = false
  }
}

// 导入偏好配置
const importPreferences = () => {
  importModalVisible.value = true
}

// 处理导入文件
const handleImportFile = (file: File) => {
  const reader = new FileReader()
  reader.onload = (e) => {
    importContent.value = e.target?.result as string
  }
  reader.readAsText(file)
  return false
}

// 执行导入
const handleImport = async () => {
  try {
    const config = JSON.parse(importContent.value)
    
    if (config.preferences) {
      Object.assign(preferences, config.preferences)
      message.success('配置导入成功')
      importModalVisible.value = false
      importContent.value = ''
    } else {
      message.error('配置格式错误')
    }
  } catch (error) {
    console.error('导入配置失败:', error)
    message.error('配置格式错误')
  }
}

// 加载通知历史
const loadNotificationHistory = async () => {
  try {
    historyLoading.value = true
    
    // 这里应该调用API获取通知历史
    // 暂时使用模拟数据
    const mockHistory = [
      {
        id: 'notif-001',
        plugin: 'email',
        status: 'success',
        title: '系统告警通知',
        content: 'CPU使用率超过80%',
        sent_at: new Date().toISOString()
      },
      {
        id: 'notif-002',
        plugin: 'dingtalk',
        status: 'failed',
        title: '磁盘空间告警',
        content: '磁盘使用率超过90%',
        error: '网络连接超时',
        sent_at: new Date(Date.now() - 3600000).toISOString()
      }
    ]
    
    notificationHistory.value = mockHistory
    historyPagination.total = mockHistory.length
    
    // 更新统计数据
    statistics.total = mockHistory.length
    statistics.success = mockHistory.filter(item => item.status === 'success').length
    statistics.failed = mockHistory.filter(item => item.status === 'failed').length
    statistics.successRate = statistics.total > 0 ? (statistics.success / statistics.total) * 100 : 0
  } catch (error) {
    console.error('加载通知历史失败:', error)
    message.error('加载通知历史失败')
  } finally {
    historyLoading.value = false
  }
}

// 处理历史表格变化
const handleHistoryTableChange = (pagination: any) => {
  historyPagination.current = pagination.current
  historyPagination.pageSize = pagination.pageSize
  loadNotificationHistory()
}

// 查看通知详情
const viewNotificationDetail = (record: any) => {
  selectedNotification.value = record
  detailModalVisible.value = true
}

// 重试通知
const retryNotification = async (record: any) => {
  try {
    // 这里应该调用API重试通知
    message.success('通知重试已提交')
    
    // 重新加载历史记录
    await loadNotificationHistory()
  } catch (error) {
    console.error('重试通知失败:', error)
    message.error('重试通知失败')
  }
}

// 组件挂载时加载数据
onMounted(() => {
  loadAvailablePlugins()
  loadNotificationHistory()
})
</script>

<style scoped>
.user-notification-manager {
  padding: 0;
}

.manager-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 24px;
  padding-bottom: 16px;
  border-bottom: 1px solid #e8e8e8;
}

.manager-header h3 {
  display: flex;
  align-items: center;
  gap: 8px;
  margin: 0;
  font-size: 18px;
  font-weight: 600;
  color: #262626;
}

.header-actions {
  display: flex;
  gap: 12px;
}

.manager-content {
  background: white;
  border-radius: 8px;
}

.preferences-section {
  padding: 24px;
}

.global-settings,
.plugin-preferences {
  margin-bottom: 32px;
}

.global-settings h4,
.plugin-preferences h4 {
  margin: 0 0 16px 0;
  font-size: 16px;
  font-weight: 600;
  color: #262626;
}

.plugin-list {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.plugin-preference-item {
  padding: 16px;
  background: #fafafa;
  border-radius: 8px;
  border: 1px solid #e8e8e8;
}

.plugin-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
}

.plugin-info {
  display: flex;
  align-items: center;
  gap: 8px;
}

.plugin-icon {
  font-size: 16px;
  color: #1890ff;
}

.plugin-name {
  font-weight: 500;
}

.plugin-settings {
  padding-top: 16px;
  border-top: 1px solid #e8e8e8;
}

.history-section {
  padding: 24px;
}

.history-filters {
  margin-bottom: 16px;
  padding: 16px;
  background: #fafafa;
  border-radius: 8px;
}

.statistics-section {
  padding: 24px;
}

.stats-overview {
  margin-bottom: 32px;
  padding: 20px;
  background: #fafafa;
  border-radius: 8px;
}

.stats-charts {
  margin-top: 24px;
}

.chart-container {
  padding: 20px;
  background: #fafafa;
  border-radius: 8px;
  text-align: center;
}

.chart-container h4 {
  margin: 0 0 16px 0;
  font-size: 14px;
  font-weight: 600;
}

.chart-placeholder {
  height: 200px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.notification-detail {
  padding: 0;
}

.message-content {
  max-height: 100px;
  overflow-y: auto;
  word-break: break-all;
}

.import-content {
  padding: 0;
}

.import-textarea-section {
  margin-top: 16px;
}

.import-textarea-section p {
  margin: 0 0 8px 0;
  color: #595959;
}

/* 响应式设计 */
@media (max-width: 768px) {
  .manager-header {
    flex-direction: column;
    gap: 16px;
    align-items: stretch;
  }
  
  .header-actions {
    justify-content: center;
  }
  
  .preferences-section,
  .history-section,
  .statistics-section {
    padding: 16px;
  }
  
  .plugin-header {
    flex-direction: column;
    gap: 12px;
    align-items: stretch;
  }
  
  .history-filters .ant-row {
    flex-direction: column;
    gap: 12px;
  }
}
</style>