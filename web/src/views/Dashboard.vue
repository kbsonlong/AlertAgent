<template>
  <div class="dashboard">
    <!-- 页面头部 -->
    <div class="page-header">
      <div class="header-content">
        <div class="header-left">
          <h1 class="page-title">监控面板</h1>
          <p class="page-description">系统运行状态总览和API集成示例</p>
        </div>
        <div class="header-actions">
          <a-space>
            <a-button @click="handleRefresh" :loading="loading">
              <template #icon><ReloadOutlined /></template>
              刷新
            </a-button>
            <a-range-picker
              v-model:value="timeRange"
              show-time
              format="YYYY-MM-DD HH:mm"
              @change="handleTimeRangeChange"
            />
          </a-space>
        </div>
      </div>
    </div>

    <!-- 系统状态卡片 -->
    <a-row :gutter="[16, 16]" class="status-section">
      <a-col :span="6">
        <a-card class="status-card">
          <div class="status-content">
            <div class="status-icon">
              <a-badge 
                :status="systemHealth === 'healthy' ? 'success' : systemHealth === 'warning' ? 'warning' : 'error'"
                :text="getHealthText(systemHealth)"
              />
            </div>
            <div class="status-info">
              <h3>系统状态</h3>
              <p class="status-value">{{ getHealthText(systemHealth) }}</p>
            </div>
          </div>
        </a-card>
      </a-col>
      
      <a-col :span="6">
        <a-card class="status-card">
          <a-statistic
            title="活跃告警"
            :value="stats.alerts.active"
            :value-style="{ color: stats.alerts.active > 0 ? '#f5222d' : '#52c41a' }"
          >
            <template #prefix>
              <ExclamationCircleOutlined />
            </template>
          </a-statistic>
        </a-card>
      </a-col>
      
      <a-col :span="6">
        <a-card class="status-card">
          <a-statistic
            title="告警规则"
            :value="stats.rules.enabled"
            :suffix="`/ ${stats.rules.total}`"
            :value-style="{ color: '#1890ff' }"
          >
            <template #prefix>
              <SettingOutlined />
            </template>
          </a-statistic>
        </a-card>
      </a-col>
      
      <a-col :span="6">
        <a-card class="status-card">
          <a-statistic
            title="数据源"
            :value="stats.providers.healthy"
            :suffix="`/ ${stats.providers.total}`"
            :value-style="{ color: '#52c41a' }"
          >
            <template #prefix>
              <DatabaseOutlined />
            </template>
          </a-statistic>
        </a-card>
      </a-col>
    </a-row>

    <!-- 图表区域 -->
    <a-row :gutter="[16, 16]" class="charts-section">
      <!-- 告警趋势图 -->
      <a-col :span="12">
        <a-card title="告警趋势" :loading="chartsLoading">
          <template #extra>
            <a-select v-model:value="alertTrendPeriod" size="small" style="width: 100px;">
              <a-select-option value="1h">1小时</a-select-option>
              <a-select-option value="6h">6小时</a-select-option>
              <a-select-option value="24h">24小时</a-select-option>
              <a-select-option value="7d">7天</a-select-option>
            </a-select>
          </template>
          <div ref="alertTrendChart" class="chart-container"></div>
        </a-card>
      </a-col>
      
      <!-- 系统性能图 -->
      <a-col :span="12">
        <a-card title="系统性能" :loading="chartsLoading">
          <template #extra>
            <a-select v-model:value="performancePeriod" size="small" style="width: 100px;">
              <a-select-option value="1h">1小时</a-select-option>
              <a-select-option value="6h">6小时</a-select-option>
              <a-select-option value="24h">24小时</a-select-option>
            </a-select>
          </template>
          <div ref="performanceChart" class="chart-container"></div>
        </a-card>
      </a-col>
      
      <!-- 数据源状态图 -->
      <a-col :span="12">
        <a-card title="数据源状态" :loading="chartsLoading">
          <div ref="providerStatusChart" class="chart-container"></div>
        </a-card>
      </a-col>
      
      <!-- 通知统计图 -->
      <a-col :span="12">
        <a-card title="通知统计" :loading="chartsLoading">
          <template #extra>
            <a-select v-model:value="notificationPeriod" size="small" style="width: 100px;">
              <a-select-option value="1h">1小时</a-select-option>
              <a-select-option value="6h">6小时</a-select-option>
              <a-select-option value="24h">24小时</a-select-option>
              <a-select-option value="7d">7天</a-select-option>
            </a-select>
          </template>
          <div ref="notificationChart" class="chart-container"></div>
        </a-card>
      </a-col>
    </a-row>

    <!-- 最近告警和活动 -->
    <a-row :gutter="[16, 16]" class="activity-section">
      <!-- 最近告警 -->
      <a-col :span="12">
        <a-card title="最近告警" :loading="loading">
          <template #extra>
            <a-button type="link" size="small" @click="$router.push('/alerts')">
              查看全部
            </a-button>
          </template>
          
          <a-list
            :data-source="recentAlerts"
            size="small"
          >
            <template #renderItem="{ item }">
              <a-list-item>
                <a-list-item-meta>
                  <template #title>
                    <div class="alert-item">
                      <a-tag :color="getAlertSeverityColor(item.severity)">
                        {{ item.severity }}
                      </a-tag>
                      <span class="alert-name">{{ item.name }}</span>
                      <span class="alert-time">{{ formatDateTime(item.createdAt) }}</span>
                    </div>
                  </template>
                  <template #description>
                    <div class="alert-description">{{ item.description }}</div>
                  </template>
                </a-list-item-meta>
              </a-list-item>
            </template>
          </a-list>
          
          <a-empty v-if="!recentAlerts.length" description="暂无告警" />
        </a-card>
      </a-col>
      
      <!-- 系统活动 -->
      <a-col :span="12">
        <a-card title="系统活动" :loading="loading">
          <template #extra>
            <a-button type="link" size="small" @click="$router.push('/logs')">
              查看全部
            </a-button>
          </template>
          
          <a-timeline size="small">
            <a-timeline-item
              v-for="activity in recentActivities"
              :key="activity.id"
              :color="getActivityColor(activity.type)"
            >
              <div class="activity-item">
                <div class="activity-header">
                  <span class="activity-type">{{ getActivityTypeLabel(activity.type) }}</span>
                  <span class="activity-time">{{ formatDateTime(activity.timestamp) }}</span>
                </div>
                <div class="activity-description">{{ activity.description }}</div>
              </div>
            </a-timeline-item>
          </a-timeline>
          
          <a-empty v-if="!recentActivities.length" description="暂无活动" />
        </a-card>
      </a-col>
    </a-row>

    <!-- API 集成示例 -->
    <a-row :gutter="[16, 16]" class="api-section">
      <a-col :span="24">
        <a-card title="API集成示例">
          <ApiExample />
        </a-card>
      </a-col>
    </a-row>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, onUnmounted, nextTick } from 'vue'
import { message } from 'ant-design-vue'
import {
  ReloadOutlined,
  ExclamationCircleOutlined,
  SettingOutlined,
  DatabaseOutlined
} from '@ant-design/icons-vue'
import { getSystemStats, checkSystemHealth } from '@/services/system'
import { formatDateTime } from '@/utils/datetime'
import ApiExample from '@/components/ApiExample.vue'

// 模拟 ECharts（实际项目中需要安装 echarts）
interface EChartsInstance {
  setOption: (option: any) => void
  resize: () => void
  dispose: () => void
}

// 响应式数据
const loading = ref(false)
const chartsLoading = ref(false)
const systemHealth = ref<'healthy' | 'warning' | 'error'>('healthy')
const timeRange = ref()
const alertTrendPeriod = ref('24h')
const performancePeriod = ref('1h')
const notificationPeriod = ref('24h')

// 图表引用
const alertTrendChart = ref()
const performanceChart = ref()
const providerStatusChart = ref()
const notificationChart = ref()

// 图表实例
const chartInstances = ref<Record<string, EChartsInstance>>({})

// 统计数据
const stats = reactive({
  alerts: {
    total: 0,
    active: 0,
    resolved: 0
  },
  rules: {
    total: 0,
    enabled: 0,
    disabled: 0
  },
  providers: {
    total: 0,
    healthy: 0,
    unhealthy: 0
  },
  notifications: {
    total: 0,
    sent: 0,
    failed: 0
  }
})

// 最近告警
const recentAlerts = ref([
  {
    id: '1',
    name: 'CPU使用率过高',
    severity: 'critical',
    description: '服务器CPU使用率超过90%',
    createdAt: new Date().toISOString()
  },
  {
    id: '2',
    name: '内存不足',
    severity: 'warning',
    description: '可用内存低于20%',
    createdAt: new Date(Date.now() - 300000).toISOString()
  }
])

// 最近活动
const recentActivities = ref([
  {
    id: '1',
    type: 'alert',
    description: '新增告警规则：CPU监控',
    timestamp: new Date().toISOString()
  },
  {
    id: '2',
    type: 'provider',
    description: '数据源连接恢复正常',
    timestamp: new Date(Date.now() - 600000).toISOString()
  },
  {
    id: '3',
    type: 'notification',
    description: '发送邮件通知成功',
    timestamp: new Date(Date.now() - 900000).toISOString()
  }
])

// 方法
const loadDashboardData = async () => {
  loading.value = true
  try {
    // 加载统计数据
    const statsData = await getSystemStats()
    Object.assign(stats, statsData)
    
    // 检查系统健康状态
    const healthData = await checkSystemHealth()
    systemHealth.value = healthData.status
    
  } catch (error) {
    message.error('加载面板数据失败')
    console.error('加载面板数据失败:', error)
  } finally {
    loading.value = false
  }
}

const loadCharts = async () => {
  chartsLoading.value = true
  try {
    await nextTick()
    
    // 初始化图表
    initAlertTrendChart()
    initPerformanceChart()
    initProviderStatusChart()
    initNotificationChart()
    
  } catch (error) {
    message.error('加载图表失败')
    console.error('加载图表失败:', error)
  } finally {
    chartsLoading.value = false
  }
}

const initAlertTrendChart = () => {
  if (!alertTrendChart.value) return
  
  // 模拟图表初始化（实际项目中使用 echarts.init）
  const mockChart = {
    setOption: (option: any) => {
      console.log('设置告警趋势图配置:', option)
    },
    resize: () => {},
    dispose: () => {}
  }
  
  chartInstances.value.alertTrend = mockChart
  
  // 设置图表配置
  const option = {
    title: {
      text: '告警趋势'
    },
    xAxis: {
      type: 'category',
      data: ['00:00', '04:00', '08:00', '12:00', '16:00', '20:00']
    },
    yAxis: {
      type: 'value'
    },
    series: [{
      data: [5, 8, 12, 6, 4, 7],
      type: 'line',
      smooth: true
    }]
  }
  
  mockChart.setOption(option)
}

const initPerformanceChart = () => {
  if (!performanceChart.value) return
  
  const mockChart = {
    setOption: (option: any) => {
      console.log('设置性能图配置:', option)
    },
    resize: () => {},
    dispose: () => {}
  }
  
  chartInstances.value.performance = mockChart
  
  const option = {
    title: {
      text: '系统性能'
    },
    legend: {
      data: ['CPU', '内存', '磁盘']
    },
    xAxis: {
      type: 'category',
      data: ['00:00', '04:00', '08:00', '12:00', '16:00', '20:00']
    },
    yAxis: {
      type: 'value',
      max: 100
    },
    series: [
      {
        name: 'CPU',
        data: [45, 52, 48, 65, 58, 62],
        type: 'line'
      },
      {
        name: '内存',
        data: [35, 42, 38, 55, 48, 52],
        type: 'line'
      },
      {
        name: '磁盘',
        data: [25, 28, 32, 35, 38, 42],
        type: 'line'
      }
    ]
  }
  
  mockChart.setOption(option)
}

const initProviderStatusChart = () => {
  if (!providerStatusChart.value) return
  
  const mockChart = {
    setOption: (option: any) => {
      console.log('设置数据源状态图配置:', option)
    },
    resize: () => {},
    dispose: () => {}
  }
  
  chartInstances.value.providerStatus = mockChart
  
  const option = {
    title: {
      text: '数据源状态',
      left: 'center'
    },
    series: [{
      type: 'pie',
      radius: '50%',
      data: [
        { value: stats.providers.healthy, name: '健康' },
        { value: stats.providers.unhealthy, name: '异常' }
      ]
    }]
  }
  
  mockChart.setOption(option)
}

const initNotificationChart = () => {
  if (!notificationChart.value) return
  
  const mockChart = {
    setOption: (option: any) => {
      console.log('设置通知统计图配置:', option)
    },
    resize: () => {},
    dispose: () => {}
  }
  
  chartInstances.value.notification = mockChart
  
  const option = {
    title: {
      text: '通知统计'
    },
    xAxis: {
      type: 'category',
      data: ['邮件', 'Webhook', '钉钉', '微信', 'Slack']
    },
    yAxis: {
      type: 'value'
    },
    series: [{
      data: [120, 85, 95, 78, 65],
      type: 'bar'
    }]
  }
  
  mockChart.setOption(option)
}

const handleRefresh = () => {
  loadDashboardData()
  loadCharts()
}

const handleTimeRangeChange = () => {
  loadCharts()
}

// 辅助函数
const getHealthText = (health: string) => {
  const texts = {
    healthy: '健康',
    warning: '警告',
    error: '错误'
  }
  return texts[health as keyof typeof texts] || health
}

const getAlertSeverityColor = (severity: string) => {
  const colors = {
    critical: 'red',
    warning: 'orange',
    info: 'blue'
  }
  return colors[severity as keyof typeof colors] || 'default'
}

const getActivityColor = (type: string) => {
  const colors = {
    alert: 'red',
    provider: 'blue',
    notification: 'green',
    system: 'orange'
  }
  return colors[type as keyof typeof colors] || 'blue'
}

const getActivityTypeLabel = (type: string) => {
  const labels = {
    alert: '告警',
    provider: '数据源',
    notification: '通知',
    system: '系统'
  }
  return labels[type as keyof typeof labels] || type
}

// 清理图表实例
const disposeCharts = () => {
  Object.values(chartInstances.value).forEach(chart => {
    if (chart && chart.dispose) {
      chart.dispose()
    }
  })
  chartInstances.value = {}
}

// 生命周期
onMounted(() => {
  loadDashboardData()
  loadCharts()
})

onUnmounted(() => {
  disposeCharts()
})
</script>

<style scoped>
.dashboard {
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

.status-section,
.charts-section,
.activity-section,
.api-section {
  margin-bottom: 16px;
}

.status-card {
  height: 100%;
}

.status-content {
  display: flex;
  align-items: center;
}

.status-icon {
  margin-right: 16px;
}

.status-info h3 {
  margin: 0 0 4px 0;
  font-size: 14px;
  color: #8c8c8c;
}

.status-value {
  margin: 0;
  font-size: 20px;
  font-weight: 600;
  color: #262626;
}

.chart-container {
  height: 300px;
  width: 100%;
}

.alert-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  width: 100%;
}

.alert-name {
  flex: 1;
  margin: 0 8px;
  font-weight: 500;
}

.alert-time {
  font-size: 12px;
  color: #8c8c8c;
}

.alert-description {
  color: #8c8c8c;
  font-size: 12px;
}

.activity-item {
  width: 100%;
}

.activity-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 4px;
}

.activity-type {
  font-weight: 500;
  color: #262626;
}

.activity-time {
  font-size: 12px;
  color: #8c8c8c;
}

.activity-description {
  color: #8c8c8c;
  font-size: 12px;
}
</style>