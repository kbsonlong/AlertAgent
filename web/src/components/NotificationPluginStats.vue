<template>
  <div class="notification-plugin-stats">
    <a-spin :spinning="loading">
      <div v-if="stats" class="stats-content">
        <!-- 统计概览 -->
        <div class="stats-overview">
          <a-row :gutter="16">
            <a-col :span="6">
              <a-statistic
                title="总发送数"
                :value="stats.total_sent"
                :value-style="{ color: '#1890ff' }"
              >
                <template #prefix>
                  <SendOutlined />
                </template>
              </a-statistic>
            </a-col>
            <a-col :span="6">
              <a-statistic
                title="成功数"
                :value="stats.success_count"
                :value-style="{ color: '#52c41a' }"
              >
                <template #prefix>
                  <CheckCircleOutlined />
                </template>
              </a-statistic>
            </a-col>
            <a-col :span="6">
              <a-statistic
                title="失败数"
                :value="stats.failure_count"
                :value-style="{ color: '#ff4d4f' }"
              >
                <template #prefix>
                  <CloseCircleOutlined />
                </template>
              </a-statistic>
            </a-col>
            <a-col :span="6">
              <a-statistic
                title="成功率"
                :value="successRate"
                suffix="%"
                :precision="1"
                :value-style="{ color: getSuccessRateColor() }"
              >
                <template #prefix>
                  <PercentageOutlined />
                </template>
              </a-statistic>
            </a-col>
          </a-row>
        </div>

        <!-- 性能指标 -->
        <div class="performance-metrics">
          <h4>性能指标</h4>
          <a-descriptions :column="2" bordered>
            <a-descriptions-item label="平均响应时间">
              {{ formatDuration(stats.avg_duration) }}
            </a-descriptions-item>
            <a-descriptions-item label="最后发送时间">
              {{ formatTime(stats.last_sent) }}
            </a-descriptions-item>
          </a-descriptions>
        </div>

        <!-- 错误信息 -->
        <div v-if="stats.last_error" class="error-info">
          <h4>最近错误</h4>
          <a-alert
            type="error"
            :message="stats.last_error"
            show-icon
            banner
          />
        </div>

        <!-- 健康状态 -->
        <div class="health-status">
          <h4>健康状态</h4>
          <a-tag
            :color="getHealthStatusColor()"
            class="health-tag"
          >
            <component :is="getHealthStatusIcon()" />
            {{ getHealthStatusText() }}
          </a-tag>
        </div>
      </div>

      <div v-else class="no-stats">
        <a-empty description="暂无统计数据" />
      </div>
    </a-spin>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import {
  Spin,
  Row,
  Col,
  Statistic,
  Descriptions,
  Alert,
  Tag,
  Empty
} from 'ant-design-vue'
import {
  SendOutlined,
  CheckCircleOutlined,
  CloseCircleOutlined,
  PercentageOutlined,
  HeartOutlined,
  WarningOutlined,
  StopOutlined
} from '@ant-design/icons-vue'
import { getPluginStats, type PluginStats } from '@/services/plugin'
import { formatTime } from '@/utils/datetime'

const ASpin = Spin
const ARow = Row
const ACol = Col
const AStatistic = Statistic
const ADescriptions = Descriptions
const ADescriptionsItem = Descriptions.Item
const AAlert = Alert
const ATag = Tag
const AEmpty = Empty

interface Props {
  pluginName: string
  stats?: PluginStats | null
}

const props = defineProps<Props>()

const loading = ref(false)
const stats = ref<PluginStats | null>(props.stats || null)

// 成功率计算
const successRate = computed(() => {
  if (!stats.value || stats.value.total_sent === 0) {
    return 0
  }
  return (stats.value.success_count / stats.value.total_sent) * 100
})

// 获取成功率颜色
const getSuccessRateColor = () => {
  const rate = successRate.value
  if (rate >= 95) return '#52c41a'
  if (rate >= 80) return '#faad14'
  return '#ff4d4f'
}

// 获取健康状态颜色
const getHealthStatusColor = () => {
  const rate = successRate.value
  if (rate >= 95) return 'green'
  if (rate >= 80) return 'orange'
  return 'red'
}

// 获取健康状态图标
const getHealthStatusIcon = () => {
  const rate = successRate.value
  if (rate >= 95) return HeartOutlined
  if (rate >= 80) return WarningOutlined
  return StopOutlined
}

// 获取健康状态文本
const getHealthStatusText = () => {
  const rate = successRate.value
  if (rate >= 95) return '健康'
  if (rate >= 80) return '警告'
  return '异常'
}

// 格式化持续时间
const formatDuration = (duration: number) => {
  if (duration < 1000) {
    return `${duration.toFixed(0)}ms`
  } else if (duration < 60000) {
    return `${(duration / 1000).toFixed(2)}s`
  } else {
    return `${(duration / 60000).toFixed(2)}min`
  }
}

// 加载统计数据
const loadStats = async () => {
  if (props.stats) {
    stats.value = props.stats
    return
  }

  try {
    loading.value = true
    const result = await getPluginStats(props.pluginName)
    stats.value = result
  } catch (error) {
    console.error('加载统计数据失败:', error)
    stats.value = null
  } finally {
    loading.value = false
  }
}

// 组件挂载时加载数据
onMounted(() => {
  loadStats()
})
</script>

<style scoped>
.notification-plugin-stats {
  padding: 0;
}

.stats-content {
  display: flex;
  flex-direction: column;
  gap: 24px;
}

.stats-overview {
  padding: 20px;
  background: #fafafa;
  border-radius: 8px;
}

.performance-metrics,
.error-info,
.health-status {
  padding: 20px;
  background: #fafafa;
  border-radius: 8px;
}

.performance-metrics h4,
.error-info h4,
.health-status h4 {
  margin: 0 0 16px 0;
  font-size: 16px;
  font-weight: 600;
  color: #262626;
}

.health-tag {
  font-size: 14px;
  padding: 4px 12px;
  display: inline-flex;
  align-items: center;
  gap: 6px;
}

.no-stats {
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: 200px;
}

/* 响应式设计 */
@media (max-width: 768px) {
  .stats-overview :deep(.ant-col) {
    margin-bottom: 16px;
  }
  
  .performance-metrics,
  .error-info,
  .health-status {
    padding: 16px;
  }
}
</style>