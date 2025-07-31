<template>
  <div class="provider-detail">
    <!-- 基本信息 -->
    <div class="detail-section">
      <h3>基本信息</h3>
      <a-descriptions :column="2" bordered>
        <a-descriptions-item label="名称">
          <div class="name-with-icon">
            <component :is="getTypeIcon(provider.type)" class="type-icon" />
            {{ provider.name }}
          </div>
        </a-descriptions-item>
        <a-descriptions-item label="类型">
          <a-tag :color="getTypeColor(provider.type)">
            {{ getTypeText(provider.type) }}
          </a-tag>
        </a-descriptions-item>
        <a-descriptions-item label="状态">
          <a-badge
            :status="getStatusBadge(provider.status)"
            :text="getStatusText(provider.status)"
          />
        </a-descriptions-item>
        <a-descriptions-item label="健康度">
          <a-progress
            :percent="provider.health || 0"
            size="small"
            :stroke-color="getHealthColor(provider.health)"
          />
        </a-descriptions-item>
        <a-descriptions-item label="URL" :span="2">
          <a :href="provider.url" target="_blank" class="url-link">
            {{ provider.url }}
            <LinkOutlined />
          </a>
        </a-descriptions-item>
        <a-descriptions-item label="描述" :span="2">
          {{ provider.description || '暂无描述' }}
        </a-descriptions-item>
      </a-descriptions>
    </div>

    <!-- 连接配置 -->
    <div class="detail-section">
      <h3>连接配置</h3>
      <a-descriptions :column="2" bordered>
        <a-descriptions-item label="认证类型">
          {{ getAuthTypeText(provider.authType) }}
        </a-descriptions-item>
        <a-descriptions-item label="超时时间">
          {{ provider.timeout || 30 }}秒
        </a-descriptions-item>
        <a-descriptions-item label="重试次数">
          {{ provider.retryCount || 3 }}次
        </a-descriptions-item>
        <a-descriptions-item label="TLS验证">
          <a-tag :color="provider.tlsVerify ? 'green' : 'orange'">
            {{ provider.tlsVerify ? '启用' : '禁用' }}
          </a-tag>
        </a-descriptions-item>
        <a-descriptions-item label="代理设置" :span="2">
          {{ provider.proxy || '无' }}
        </a-descriptions-item>
      </a-descriptions>
    </div>

    <!-- 统计信息 -->
    <div class="detail-section">
      <h3>统计信息</h3>
      <a-row :gutter="16">
        <a-col :span="6">
          <a-card>
            <a-statistic
              title="指标数量"
              :value="provider.metricsCount || 0"
              :value-style="{ color: '#1890ff' }"
            >
              <template #prefix>
                <BarChartOutlined />
              </template>
            </a-statistic>
          </a-card>
        </a-col>
        <a-col :span="6">
          <a-card>
            <a-statistic
              title="规则数量"
              :value="provider.rulesCount || 0"
              :value-style="{ color: '#52c41a' }"
            >
              <template #prefix>
                <AlertOutlined />
              </template>
            </a-statistic>
          </a-card>
        </a-col>
        <a-col :span="6">
          <a-card>
            <a-statistic
              title="告警数量"
              :value="provider.alertsCount || 0"
              :value-style="{ color: '#faad14' }"
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
              title="响应时间"
              :value="provider.responseTime || 0"
              suffix="ms"
              :value-style="{ color: '#722ed1' }"
            >
              <template #prefix>
                <ClockCircleOutlined />
              </template>
            </a-statistic>
          </a-card>
        </a-col>
      </a-row>
    </div>

    <!-- 健康状态 -->
    <div class="detail-section">
      <h3>健康状态</h3>
      <a-card>
        <div class="health-status">
          <div class="health-header">
            <div class="health-score">
              <a-progress
                type="circle"
                :percent="provider.health || 0"
                :stroke-color="getHealthColor(provider.health)"
                :width="80"
              />
              <div class="health-text">
                <div class="score">{{ provider.health || 0 }}%</div>
                <div class="label">健康度</div>
              </div>
            </div>
            <div class="health-info">
              <div class="info-item">
                <span class="label">最后检查:</span>
                <span class="value">{{ formatDateTime(provider.lastCheck) }}</span>
              </div>
              <div class="info-item">
                <span class="label">检查间隔:</span>
                <span class="value">{{ provider.checkInterval || 60 }}秒</span>
              </div>
              <div class="info-item">
                <span class="label">连续失败:</span>
                <span class="value">{{ provider.failureCount || 0 }}次</span>
              </div>
            </div>
          </div>
          
          <div class="health-metrics" v-if="provider.healthMetrics">
            <h4>性能指标</h4>
            <a-row :gutter="16">
              <a-col :span="6">
                <div class="metric-item">
                  <div class="metric-label">CPU使用率</div>
                  <a-progress
                    :percent="provider.healthMetrics.cpuUsage || 0"
                    size="small"
                    :stroke-color="getMetricColor(provider.healthMetrics.cpuUsage)"
                  />
                </div>
              </a-col>
              <a-col :span="6">
                <div class="metric-item">
                  <div class="metric-label">内存使用率</div>
                  <a-progress
                    :percent="provider.healthMetrics.memoryUsage || 0"
                    size="small"
                    :stroke-color="getMetricColor(provider.healthMetrics.memoryUsage)"
                  />
                </div>
              </a-col>
              <a-col :span="6">
                <div class="metric-item">
                  <div class="metric-label">磁盘使用率</div>
                  <a-progress
                    :percent="provider.healthMetrics.diskUsage || 0"
                    size="small"
                    :stroke-color="getMetricColor(provider.healthMetrics.diskUsage)"
                  />
                </div>
              </a-col>
              <a-col :span="6">
                <div class="metric-item">
                  <div class="metric-label">运行时间</div>
                  <div class="metric-value">
                    {{ formatUptime(provider.healthMetrics.uptime) }}
                  </div>
                </div>
              </a-col>
            </a-row>
          </div>
        </div>
      </a-card>
    </div>

    <!-- 标签和注解 -->
    <div class="detail-section">
      <h3>标签和注解</h3>
      <a-row :gutter="16">
        <a-col :span="12">
          <a-card title="标签" size="small">
            <div class="tags-container">
              <a-tag
                v-for="(value, key) in provider.labels"
                :key="key"
                color="blue"
                class="tag-item"
              >
                {{ key }}: {{ value }}
              </a-tag>
              <div v-if="!provider.labels || Object.keys(provider.labels).length === 0" class="empty-text">
                暂无标签
              </div>
            </div>
          </a-card>
        </a-col>
        <a-col :span="12">
          <a-card title="注解" size="small">
            <div class="annotations-container">
              <div
                v-for="(value, key) in provider.annotations"
                :key="key"
                class="annotation-item"
              >
                <div class="annotation-key">{{ key }}:</div>
                <div class="annotation-value">{{ value }}</div>
              </div>
              <div v-if="!provider.annotations || Object.keys(provider.annotations).length === 0" class="empty-text">
                暂无注解
              </div>
            </div>
          </a-card>
        </a-col>
      </a-row>
    </div>

    <!-- 操作历史 -->
    <div class="detail-section">
      <h3>操作历史</h3>
      <a-card>
        <a-timeline>
          <a-timeline-item
            v-for="(item, index) in provider.history"
            :key="index"
            :color="getHistoryColor(item.type)"
          >
            <template #dot>
              <component :is="getHistoryIcon(item.type)" />
            </template>
            <div class="history-item">
              <div class="history-header">
                <span class="history-action">{{ item.action }}</span>
                <span class="history-time">{{ formatDateTime(item.timestamp) }}</span>
              </div>
              <div class="history-details" v-if="item.details">
                {{ item.details }}
              </div>
              <div class="history-user" v-if="item.user">
                操作人: {{ item.user }}
              </div>
            </div>
          </a-timeline-item>
          <a-timeline-item v-if="!provider.history || provider.history.length === 0">
            <div class="empty-text">暂无操作历史</div>
          </a-timeline-item>
        </a-timeline>
      </a-card>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import {
  Card,
  Row,
  Col,
  Descriptions,
  Tag,
  Badge,
  Progress,
  Statistic,
  Timeline
} from 'ant-design-vue'
import {
  DatabaseOutlined,
  BarChartOutlined,
  AlertOutlined,
  WarningOutlined,
  ClockCircleOutlined,
  LinkOutlined,
  CheckCircleOutlined,
  CloseCircleOutlined,
  EditOutlined,
  DeleteOutlined,
  SyncOutlined
} from '@ant-design/icons-vue'
import { formatDateTime } from '@/utils/datetime'
import type { Provider } from '@/types'

const ACard = Card
const ARow = Row
const ACol = Col
const ADescriptions = Descriptions
const ADescriptionsItem = Descriptions.Item
const ATag = Tag
const ABadge = Badge
const AProgress = Progress
const AStatistic = Statistic
const ATimeline = Timeline
const ATimelineItem = Timeline.Item

// 组件属性
interface Props {
  provider: Provider
}

const props = defineProps<Props>()

// 组件事件
interface Emits {
  edit: [provider: Provider]
  test: [provider: Provider]
  delete: [provider: Provider]
  close: []
}

const emit = defineEmits<Emits>()

// 获取类型图标
const getTypeIcon = (type: string) => {
  const iconMap: Record<string, any> = {
    prometheus: DatabaseOutlined,
    grafana: BarChartOutlined,
    alertmanager: WarningOutlined,
    elasticsearch: DatabaseOutlined,
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

// 获取指标颜色
const getMetricColor = (value: number) => {
  if (value >= 80) return '#ff4d4f'
  if (value >= 60) return '#faad14'
  return '#52c41a'
}

// 获取认证类型文本
const getAuthTypeText = (authType: string) => {
  const textMap: Record<string, string> = {
    none: '无认证',
    basic: '基础认证',
    bearer: 'Bearer Token',
    oauth2: 'OAuth2',
    apikey: 'API Key'
  }
  return textMap[authType] || authType
}

// 格式化运行时间
const formatUptime = (uptime: number) => {
  if (!uptime) return '0秒'
  
  const days = Math.floor(uptime / (24 * 3600))
  const hours = Math.floor((uptime % (24 * 3600)) / 3600)
  const minutes = Math.floor((uptime % 3600) / 60)
  
  if (days > 0) {
    return `${days}天${hours}小时`
  } else if (hours > 0) {
    return `${hours}小时${minutes}分钟`
  } else {
    return `${minutes}分钟`
  }
}

// 获取历史记录颜色
const getHistoryColor = (type: string) => {
  const colorMap: Record<string, string> = {
    create: 'green',
    update: 'blue',
    delete: 'red',
    test: 'orange',
    sync: 'purple'
  }
  return colorMap[type] || 'gray'
}

// 获取历史记录图标
const getHistoryIcon = (type: string) => {
  const iconMap: Record<string, any> = {
    create: CheckCircleOutlined,
    update: EditOutlined,
    delete: DeleteOutlined,
    test: WarningOutlined,
    sync: SyncOutlined
  }
  return iconMap[type] || CheckCircleOutlined
}
</script>

<style scoped>
.provider-detail {
  padding: 0;
}

.detail-section {
  margin-bottom: 24px;
}

.detail-section:last-child {
  margin-bottom: 0;
}

.detail-section h3 {
  margin: 0 0 16px 0;
  font-size: 16px;
  font-weight: 600;
  color: #262626;
}

.name-with-icon {
  display: flex;
  align-items: center;
  gap: 8px;
}

.type-icon {
  font-size: 16px;
  color: #1890ff;
}

.url-link {
  color: #1890ff;
  text-decoration: none;
  display: flex;
  align-items: center;
  gap: 4px;
}

.url-link:hover {
  text-decoration: underline;
}

.health-status {
  padding: 16px 0;
}

.health-header {
  display: flex;
  align-items: center;
  gap: 32px;
  margin-bottom: 24px;
}

.health-score {
  display: flex;
  align-items: center;
  gap: 16px;
}

.health-text {
  text-align: center;
}

.health-text .score {
  font-size: 24px;
  font-weight: 600;
  color: #262626;
}

.health-text .label {
  font-size: 14px;
  color: #8c8c8c;
}

.health-info {
  flex: 1;
}

.info-item {
  display: flex;
  justify-content: space-between;
  margin-bottom: 8px;
  padding: 8px 0;
  border-bottom: 1px solid #f0f0f0;
}

.info-item:last-child {
  border-bottom: none;
  margin-bottom: 0;
}

.info-item .label {
  color: #8c8c8c;
  font-weight: 500;
}

.info-item .value {
  color: #262626;
}

.health-metrics {
  border-top: 1px solid #f0f0f0;
  padding-top: 24px;
}

.health-metrics h4 {
  margin: 0 0 16px 0;
  font-size: 14px;
  font-weight: 600;
  color: #262626;
}

.metric-item {
  text-align: center;
}

.metric-label {
  font-size: 12px;
  color: #8c8c8c;
  margin-bottom: 8px;
}

.metric-value {
  font-size: 14px;
  font-weight: 500;
  color: #262626;
}

.tags-container,
.annotations-container {
  min-height: 60px;
}

.tag-item {
  margin-bottom: 8px;
}

.annotation-item {
  display: flex;
  margin-bottom: 8px;
  padding: 8px;
  background: #fafafa;
  border-radius: 4px;
}

.annotation-key {
  font-weight: 500;
  color: #262626;
  margin-right: 8px;
  min-width: 80px;
}

.annotation-value {
  color: #595959;
  word-break: break-all;
}

.empty-text {
  color: #8c8c8c;
  font-style: italic;
  text-align: center;
  padding: 20px 0;
}

.history-item {
  padding: 8px 0;
}

.history-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 4px;
}

.history-action {
  font-weight: 500;
  color: #262626;
}

.history-time {
  font-size: 12px;
  color: #8c8c8c;
}

.history-details {
  font-size: 14px;
  color: #595959;
  margin-bottom: 4px;
}

.history-user {
  font-size: 12px;
  color: #8c8c8c;
}

/* 响应式设计 */
@media (max-width: 768px) {
  .health-header {
    flex-direction: column;
    gap: 16px;
  }
  
  .health-score {
    flex-direction: column;
    gap: 8px;
  }
  
  .info-item {
    flex-direction: column;
    gap: 4px;
  }
  
  .history-header {
    flex-direction: column;
    align-items: flex-start;
    gap: 4px;
  }
  
  .annotation-item {
    flex-direction: column;
    gap: 4px;
  }
  
  .annotation-key {
    min-width: auto;
  }
}
</style>