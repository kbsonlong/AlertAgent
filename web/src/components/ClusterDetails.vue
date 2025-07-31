<template>
  <div class="cluster-details">
    <!-- 基本信息 -->
    <div class="basic-info">
      <a-descriptions title="基本信息" :column="2" bordered>
        <a-descriptions-item label="集群ID">
          {{ cluster?.cluster_id }}
        </a-descriptions-item>
        <a-descriptions-item label="集群名称">
          {{ cluster?.cluster_name || '-' }}
        </a-descriptions-item>
        <a-descriptions-item label="配置类型">
          <a-tag :color="getConfigTypeColor(cluster?.config_type)">
            {{ getConfigTypeText(cluster?.config_type) }}
          </a-tag>
        </a-descriptions-item>
        <a-descriptions-item label="同步状态">
          <a-tag :color="getStatusColor(cluster?.sync_status)">
            {{ getStatusText(cluster?.sync_status) }}
          </a-tag>
        </a-descriptions-item>
        <a-descriptions-item label="最后同步时间">
          {{ cluster?.last_sync_time ? formatDateTime(cluster.last_sync_time) : '从未同步' }}
        </a-descriptions-item>
        <a-descriptions-item label="下次同步时间">
          {{ cluster?.next_sync_time ? formatDateTime(cluster.next_sync_time) : '-' }}
        </a-descriptions-item>
        <a-descriptions-item label="同步延迟">
          <span :class="getDelayClass(cluster?.sync_delay)">
            {{ formatDelay(cluster?.sync_delay) }}
          </span>
        </a-descriptions-item>
        <a-descriptions-item label="失败次数">
          <a-badge 
            :count="cluster?.failure_count || 0" 
            :number-style="{ backgroundColor: cluster?.failure_count > 0 ? '#ff4d4f' : '#52c41a' }"
          />
        </a-descriptions-item>
      </a-descriptions>
    </div>

    <!-- 配置信息 -->
    <div class="config-info">
      <a-card title="配置信息" size="small" style="margin-top: 16px;">
        <template #extra>
          <a-space>
            <a-button size="small" @click="handleRefreshConfig">
              <template #icon><ReloadOutlined /></template>
              刷新
            </a-button>
            <a-button size="small" @click="handleEditConfig">
              <template #icon><EditOutlined /></template>
              编辑
            </a-button>
          </a-space>
        </template>
        
        <a-tabs v-model:activeKey="activeTab" size="small">
          <a-tab-pane key="current" tab="当前配置">
            <div class="config-content">
              <pre v-if="cluster?.current_config">{{ formatConfig(cluster.current_config) }}</pre>
              <a-empty v-else description="暂无配置" size="small" />
            </div>
          </a-tab-pane>
          <a-tab-pane key="target" tab="目标配置">
            <div class="config-content">
              <pre v-if="cluster?.target_config">{{ formatConfig(cluster.target_config) }}</pre>
              <a-empty v-else description="暂无配置" size="small" />
            </div>
          </a-tab-pane>
          <a-tab-pane key="diff" tab="配置差异">
            <div class="config-diff">
              <div v-if="configDiff.length > 0">
                <div v-for="(diff, index) in configDiff" :key="index" class="diff-item">
                  <div class="diff-path">{{ diff.path }}</div>
                  <div class="diff-content">
                    <div class="diff-old">
                      <span class="diff-label">当前:</span>
                      <code>{{ diff.oldValue }}</code>
                    </div>
                    <div class="diff-new">
                      <span class="diff-label">目标:</span>
                      <code>{{ diff.newValue }}</code>
                    </div>
                  </div>
                </div>
              </div>
              <a-empty v-else description="无配置差异" size="small" />
            </div>
          </a-tab-pane>
        </a-tabs>
      </a-card>
    </div>

    <!-- 同步历史 -->
    <div class="sync-history">
      <a-card title="同步历史" size="small" style="margin-top: 16px;">
        <template #extra>
          <a-button size="small" @click="loadSyncHistory">
            <template #icon><ReloadOutlined /></template>
            刷新
          </a-button>
        </template>
        
        <a-table
          :columns="historyColumns"
          :data-source="syncHistory"
          :loading="historyLoading"
          :pagination="{
            pageSize: 5,
            showSizeChanger: false,
            showQuickJumper: false
          }"
          size="small"
        >
          <template #bodyCell="{ column, record }">
            <template v-if="column.key === 'status'">
              <a-tag :color="getStatusColor(record.status)">
                {{ getStatusText(record.status) }}
              </a-tag>
            </template>
            <template v-else-if="column.key === 'sync_time'">
              {{ formatDateTime(record.sync_time) }}
            </template>
            <template v-else-if="column.key === 'duration'">
              {{ formatDuration(record.duration) }}
            </template>
            <template v-else-if="column.key === 'action'">
              <a-button
                type="link"
                size="small"
                @click="handleViewSyncDetail(record)"
              >
                详情
              </a-button>
            </template>
          </template>
        </a-table>
      </a-card>
    </div>

    <!-- 操作按钮 -->
    <div class="actions" style="margin-top: 16px; text-align: right;">
      <a-space>
        <a-button @click="handleManualSync" :loading="syncLoading">
          <template #icon><SyncOutlined /></template>
          手动同步
        </a-button>
        <a-button @click="handleResetConfig" danger>
          <template #icon><UndoOutlined /></template>
          重置配置
        </a-button>
      </a-space>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { message } from 'ant-design-vue'
import {
  ReloadOutlined,
  EditOutlined,
  SyncOutlined,
  UndoOutlined
} from '@ant-design/icons-vue'
import { formatDateTime } from '@/utils/datetime'

// 组件属性
interface Props {
  cluster: any
}

const props = defineProps<Props>()

// 响应式数据
const activeTab = ref('current')
const syncHistory = ref<any[]>([])
const historyLoading = ref(false)
const syncLoading = ref(false)

// 同步历史表格列
const historyColumns = [
  {
    title: '同步时间',
    dataIndex: 'sync_time',
    key: 'sync_time',
    width: 160
  },
  {
    title: '状态',
    dataIndex: 'status',
    key: 'status',
    width: 100
  },
  {
    title: '耗时',
    dataIndex: 'duration',
    key: 'duration',
    width: 80
  },
  {
    title: '变更数量',
    dataIndex: 'changes_count',
    key: 'changes_count',
    width: 100
  },
  {
    title: '错误信息',
    dataIndex: 'error_message',
    key: 'error_message',
    ellipsis: true
  },
  {
    title: '操作',
    key: 'action',
    width: 80
  }
]

// 计算配置差异
const configDiff = computed(() => {
  if (!props.cluster?.current_config || !props.cluster?.target_config) {
    return []
  }
  
  // 这里应该实现真正的配置差异比较逻辑
  // 现在返回模拟数据
  return [
    {
      path: 'global.scrape_interval',
      oldValue: '15s',
      newValue: '30s'
    },
    {
      path: 'rule_files[0]',
      oldValue: '/etc/prometheus/rules/*.yml',
      newValue: '/etc/prometheus/rules/**/*.yml'
    }
  ]
})

// 获取配置类型颜色
const getConfigTypeColor = (type: string) => {
  const colorMap: Record<string, string> = {
    prometheus: 'blue',
    alertmanager: 'orange',
    grafana: 'green'
  }
  return colorMap[type] || 'default'
}

// 获取配置类型文本
const getConfigTypeText = (type: string) => {
  const textMap: Record<string, string> = {
    prometheus: 'Prometheus',
    alertmanager: 'Alertmanager',
    grafana: 'Grafana'
  }
  return textMap[type] || type
}

// 获取状态颜色
const getStatusColor = (status: string) => {
  const colorMap: Record<string, string> = {
    success: 'green',
    failed: 'red',
    syncing: 'blue',
    pending: 'orange'
  }
  return colorMap[status] || 'default'
}

// 获取状态文本
const getStatusText = (status: string) => {
  const textMap: Record<string, string> = {
    success: '成功',
    failed: '失败',
    syncing: '同步中',
    pending: '待同步'
  }
  return textMap[status] || status
}

// 获取延迟样式类
const getDelayClass = (delay: number) => {
  if (!delay) return ''
  if (delay > 300) return 'delay-high'
  if (delay > 100) return 'delay-medium'
  return 'delay-low'
}

// 格式化延迟
const formatDelay = (delay: number) => {
  if (!delay) return '0ms'
  if (delay > 1000) {
    return `${(delay / 1000).toFixed(1)}s`
  }
  return `${delay}ms`
}

// 格式化持续时间
const formatDuration = (duration: number) => {
  if (!duration) return '-'
  if (duration > 60) {
    return `${Math.floor(duration / 60)}m ${duration % 60}s`
  }
  return `${duration}s`
}

// 格式化配置
const formatConfig = (config: string) => {
  try {
    if (typeof config === 'string') {
      // 尝试解析为JSON并格式化
      const parsed = JSON.parse(config)
      return JSON.stringify(parsed, null, 2)
    }
    return JSON.stringify(config, null, 2)
  } catch {
    // 如果不是JSON，直接返回
    return config
  }
}

// 加载同步历史
const loadSyncHistory = async () => {
  historyLoading.value = true
  try {
    // 模拟数据
    const mockHistory = [
      {
        id: 1,
        sync_time: '2024-01-01 10:30:00',
        status: 'success',
        duration: 45,
        changes_count: 3,
        error_message: null
      },
      {
        id: 2,
        sync_time: '2024-01-01 10:00:00',
        status: 'failed',
        duration: 120,
        changes_count: 0,
        error_message: '连接超时'
      },
      {
        id: 3,
        sync_time: '2024-01-01 09:30:00',
        status: 'success',
        duration: 30,
        changes_count: 1,
        error_message: null
      }
    ]
    
    syncHistory.value = mockHistory
  } catch (error) {
    message.error('加载同步历史失败')
  } finally {
    historyLoading.value = false
  }
}

// 刷新配置
const handleRefreshConfig = async () => {
  try {
    message.success('配置已刷新')
  } catch (error) {
    message.error('刷新配置失败')
  }
}

// 编辑配置
const handleEditConfig = () => {
  message.info('编辑配置功能开发中')
}

// 查看同步详情
const handleViewSyncDetail = (record: any) => {
  message.info('查看同步详情功能开发中')
}

// 手动同步
const handleManualSync = async () => {
  syncLoading.value = true
  try {
    // 模拟同步过程
    await new Promise(resolve => setTimeout(resolve, 2000))
    message.success('同步完成')
    loadSyncHistory()
  } catch (error) {
    message.error('同步失败')
  } finally {
    syncLoading.value = false
  }
}

// 重置配置
const handleResetConfig = () => {
  message.info('重置配置功能开发中')
}

// 组件挂载
onMounted(() => {
  loadSyncHistory()
})
</script>

<style scoped>
.cluster-details {
  padding: 16px;
}

.config-content {
  max-height: 400px;
  overflow-y: auto;
}

.config-content pre {
  background: #f5f5f5;
  padding: 12px;
  border-radius: 4px;
  font-size: 12px;
  line-height: 1.4;
  margin: 0;
}

.config-diff {
  max-height: 400px;
  overflow-y: auto;
}

.diff-item {
  margin-bottom: 16px;
  padding: 12px;
  border: 1px solid #d9d9d9;
  border-radius: 4px;
}

.diff-path {
  font-weight: 600;
  color: #1890ff;
  margin-bottom: 8px;
}

.diff-content {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.diff-old,
.diff-new {
  display: flex;
  align-items: center;
  gap: 8px;
}

.diff-label {
  font-weight: 500;
  min-width: 40px;
}

.diff-old .diff-label {
  color: #ff4d4f;
}

.diff-new .diff-label {
  color: #52c41a;
}

.diff-old code {
  background: #fff2f0;
  color: #ff4d4f;
}

.diff-new code {
  background: #f6ffed;
  color: #52c41a;
}

.delay-low {
  color: #52c41a;
}

.delay-medium {
  color: #faad14;
}

.delay-high {
  color: #ff4d4f;
}

:deep(.ant-descriptions-item-content) {
  word-break: break-all;
}

:deep(.ant-table-tbody > tr > td) {
  padding: 8px 12px;
}
</style>