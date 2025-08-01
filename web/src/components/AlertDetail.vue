<template>
  <div class="alert-detail">
    <!-- 告警基本信息 -->
    <a-card title="基本信息" class="detail-card">
      <a-descriptions :column="2" bordered>
        <a-descriptions-item label="告警名称">
          {{ alert.name }}
        </a-descriptions-item>
        <a-descriptions-item label="状态">
          <a-tag :color="getStatusColor(alert.status)">
            {{ getStatusText(alert.status) }}
          </a-tag>
        </a-descriptions-item>
        <a-descriptions-item label="严重程度">
          <a-tag :color="getSeverityColor(alert.severity)">
            {{ getSeverityText(alert.severity) }}
          </a-tag>
        </a-descriptions-item>
        <a-descriptions-item label="数据源">
          {{ alert.source }}
        </a-descriptions-item>
        <a-descriptions-item label="创建时间">
          {{ formatDateTime(alert.created_at) }}
        </a-descriptions-item>
        <a-descriptions-item label="更新时间">
          {{ formatDateTime(alert.updated_at) }}
        </a-descriptions-item>
        <a-descriptions-item label="描述" :span="2">
          {{ alert.description || '无描述' }}
        </a-descriptions-item>
      </a-descriptions>
    </a-card>

    <!-- 告警标签 -->
    <a-card title="标签" class="detail-card" v-if="parsedLabels && Object.keys(parsedLabels).length > 0">
      <a-space wrap>
        <a-tag v-for="(value, key) in parsedLabels" :key="key" color="blue">
          {{ key }}: {{ value }}
        </a-tag>
      </a-space>
    </a-card>

    <!-- 告警注解 -->
    <a-card title="注解" class="detail-card" v-if="alert.annotations && Object.keys(alert.annotations).length > 0">
      <a-descriptions :column="1" bordered>
        <a-descriptions-item
          v-for="(value, key) in alert.annotations"
          :key="key"
          :label="key"
        >
          <div v-if="isUrl(value)">
            <a :href="value" target="_blank">{{ value }}</a>
          </div>
          <div v-else-if="isMarkdown(key)" v-html="renderMarkdown(value)"></div>
          <div v-else>{{ value }}</div>
        </a-descriptions-item>
      </a-descriptions>
    </a-card>

    <!-- 告警指标 -->
    <a-card title="指标数据" class="detail-card" v-if="alert.metrics">
      <a-table
        :columns="metricColumns"
        :data-source="metricData"
        :pagination="false"
        size="small"
      >
        <template #value="{ record }">
          <a-tag :color="getMetricColor(record.status)">
            {{ record.value }}
          </a-tag>
        </template>
      </a-table>
    </a-card>

    <!-- 相关告警 -->
    <a-card title="相关告警" class="detail-card" v-if="relatedAlerts.length > 0">
      <a-list
        :data-source="relatedAlerts"
        size="small"
      >
        <template #renderItem="{ item }">
          <a-list-item>
            <a-list-item-meta>
              <template #title>
                <a @click="$emit('viewRelated', item)">{{ item.name }}</a>
              </template>
              <template #description>
                <a-space>
                  <a-tag :color="getStatusColor(item.status)" size="small">
                    {{ getStatusText(item.status) }}
                  </a-tag>
                  <span>{{ getFriendlyTime(item.created_at) }}</span>
                </a-space>
              </template>
            </a-list-item-meta>
          </a-list-item>
        </template>
      </a-list>
    </a-card>

    <!-- 操作历史 -->
    <a-card title="操作历史" class="detail-card">
      <a-timeline>
        <a-timeline-item
          v-for="(history, index) in alert.history || []"
          :key="index"
          :color="getHistoryColor(history.action)"
        >
          <template #dot>
            <component :is="getHistoryIcon(history.action)" />
          </template>
          <div class="history-item">
            <div class="history-action">
              <strong>{{ getHistoryText(history.action) }}</strong>
              <span class="history-time">{{ formatDateTime(history.timestamp) }}</span>
            </div>
            <div class="history-user" v-if="history.user">
              操作人: {{ history.user }}
            </div>
            <div class="history-comment" v-if="history.comment">
              备注: {{ history.comment }}
            </div>
          </div>
        </a-timeline-item>
        <a-timeline-item color="blue">
          <template #dot>
            <PlusOutlined />
          </template>
          <div class="history-item">
            <div class="history-action">
              <strong>告警创建</strong>
              <span class="history-time">{{ formatDateTime(alert.created_at) }}</span>
            </div>
          </div>
        </a-timeline-item>
      </a-timeline>
    </a-card>

    <!-- 操作按钮 -->
    <div class="action-buttons">
      <a-space>
        <a-button
          v-if="alert.status === 'firing'"
          type="primary"
          @click="acknowledgeAlert"
          :loading="actionLoading"
        >
          <template #icon><CheckCircleOutlined /></template>
          确认告警
        </a-button>
        <a-button
          v-if="alert.status !== 'resolved'"
          type="primary"
          @click="resolveAlert"
          :loading="actionLoading"
        >
          <template #icon><CheckOutlined /></template>
          解决告警
        </a-button>
        <a-button @click="analyzeAlert" :loading="analysisLoading">
          <template #icon><BulbOutlined /></template>
          AI分析
        </a-button>
        <a-button @click="convertToKnowledge" :loading="convertLoading">
          <template #icon><BookOutlined /></template>
          转为知识
        </a-button>
        <a-button @click="exportAlert">
          <template #icon><ExportOutlined /></template>
          导出
        </a-button>
      </a-space>
    </div>

    <!-- 添加备注模态框 -->
    <a-modal
      v-model:open="commentModalVisible"
      title="添加备注"
      @ok="submitComment"
      :confirm-loading="commentLoading"
    >
      <a-form :model="commentForm" layout="vertical">
        <a-form-item label="备注内容" required>
          <a-textarea
            v-model:value="commentForm.comment"
            placeholder="请输入备注内容"
            :rows="4"
          />
        </a-form-item>
      </a-form>
    </a-modal>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, reactive } from 'vue'
import {
  Card,
  Descriptions,
  Tag,
  Space,
  Table,
  List,
  Timeline,
  Button,
  Modal,
  Form,
  Input,
  message
} from 'ant-design-vue'
import {
  CheckCircleOutlined,
  CheckOutlined,
  BulbOutlined,
  BookOutlined,
  ExportOutlined,
  PlusOutlined,
  EditOutlined,
  DeleteOutlined,
  ExclamationCircleOutlined
} from '@ant-design/icons-vue'
import { formatDateTime, getFriendlyTime } from '@/utils/datetime'
import { updateAlert, analyzeAlert as analyzeAlertApi, convertToKnowledge as convertToKnowledgeApi } from '@/services/alert'
import type { Alert } from '@/types'

const ACard = Card
const ADescriptions = Descriptions
const ADescriptionsItem = Descriptions.Item
const ATag = Tag
const ASpace = Space
const ATable = Table
const AList = List
const AListItem = List.Item
const AListItemMeta = List.Item.Meta
const ATimeline = Timeline
const ATimelineItem = Timeline.Item
const AButton = Button
const AModal = Modal
const AForm = Form
const AFormItem = Form.Item
const ATextarea = Input.TextArea

interface Props {
  alert: Alert
}

interface Emits {
  (e: 'update', alert: Alert): void
  (e: 'close'): void
  (e: 'viewRelated', alert: Alert): void
}

const props = defineProps<Props>()
const emit = defineEmits<Emits>()

// 响应式数据
const actionLoading = ref(false)
const analysisLoading = ref(false)
const convertLoading = ref(false)
const commentLoading = ref(false)
const commentModalVisible = ref(false)
const relatedAlerts = ref<Alert[]>([])

// 备注表单
const commentForm = reactive({
  comment: ''
})

// 指标表格列
const metricColumns = [
  {
    title: '指标名称',
    dataIndex: 'name',
    key: 'name'
  },
  {
    title: '当前值',
    dataIndex: 'value',
    key: 'value',
    slots: { customRender: 'value' }
  },
  {
    title: '阈值',
    dataIndex: 'threshold',
    key: 'threshold'
  },
  {
    title: '单位',
    dataIndex: 'unit',
    key: 'unit'
  }
]

// 解析标签数据
const parsedLabels = computed(() => {
  if (!props.alert.labels) return {}
  
  // 如果labels已经是对象，直接返回
  if (typeof props.alert.labels === 'object') {
    return props.alert.labels
  }
  
  // 如果labels是字符串，尝试解析JSON
  try {
    return JSON.parse(props.alert.labels)
  } catch (error) {
    console.error('解析labels失败:', error)
    return {}
  }
})

// 指标数据
const metricData = computed(() => {
  if (!props.alert.metrics) return []
  return Object.entries(props.alert.metrics).map(([key, value]) => ({
    name: key,
    value: value.current,
    threshold: value.threshold,
    unit: value.unit,
    status: value.status
  }))
})

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

// 获取指标颜色
const getMetricColor = (status: string) => {
  const colorMap: Record<string, string> = {
    normal: 'green',
    warning: 'orange',
    critical: 'red'
  }
  return colorMap[status] || 'default'
}

// 获取历史操作颜色
const getHistoryColor = (action: string) => {
  const colorMap: Record<string, string> = {
    acknowledge: 'blue',
    resolve: 'green',
    comment: 'gray',
    escalate: 'orange'
  }
  return colorMap[action] || 'blue'
}

// 获取历史操作图标
const getHistoryIcon = (action: string) => {
  const iconMap: Record<string, any> = {
    acknowledge: CheckCircleOutlined,
    resolve: CheckOutlined,
    comment: EditOutlined,
    escalate: ExclamationCircleOutlined
  }
  return iconMap[action] || EditOutlined
}

// 获取历史操作文本
const getHistoryText = (action: string) => {
  const textMap: Record<string, string> = {
    acknowledge: '确认告警',
    resolve: '解决告警',
    comment: '添加备注',
    escalate: '升级告警'
  }
  return textMap[action] || action
}

// 判断是否为URL
const isUrl = (text: string) => {
  try {
    new URL(text)
    return true
  } catch {
    return false
  }
}

// 判断是否为Markdown字段
const isMarkdown = (key: string) => {
  return key.toLowerCase().includes('markdown') || key.toLowerCase().includes('md')
}

// 渲染Markdown（简单实现）
const renderMarkdown = (text: string) => {
  return text
    .replace(/\*\*(.*?)\*\*/g, '<strong>$1</strong>')
    .replace(/\*(.*?)\*/g, '<em>$1</em>')
    .replace(/`(.*?)`/g, '<code>$1</code>')
    .replace(/\n/g, '<br>')
}

// 确认告警
const acknowledgeAlert = async () => {
  actionLoading.value = true
  try {
    const response = await updateAlert(props.alert.id, { status: 'acknowledged' })
    emit('update', response.data)
    message.success('告警已确认')
  } catch (error) {
    message.error('确认告警失败')
  } finally {
    actionLoading.value = false
  }
}

// 解决告警
const resolveAlert = async () => {
  actionLoading.value = true
  try {
    const response = await updateAlert(props.alert.id, { status: 'resolved' })
    emit('update', response.data)
    message.success('告警已解决')
  } catch (error) {
    message.error('解决告警失败')
  } finally {
    actionLoading.value = false
  }
}

// AI分析
const analyzeAlert = async () => {
  analysisLoading.value = true
  try {
    const response = await analyzeAlertApi(props.alert.id)
    // 这里可以显示分析结果或跳转到分析页面
    message.success('AI分析完成')
  } catch (error) {
    message.error('AI分析失败')
  } finally {
    analysisLoading.value = false
  }
}

// 转为知识
const convertToKnowledge = async () => {
  convertLoading.value = true
  try {
    await convertToKnowledgeApi(props.alert.id)
    message.success('已转为知识库条目')
  } catch (error) {
    message.error('转换失败')
  } finally {
    convertLoading.value = false
  }
}

// 导出告警
const exportAlert = () => {
  const data = JSON.stringify(props.alert, null, 2)
  const blob = new Blob([data], { type: 'application/json' })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = `alert-${props.alert.id}.json`
  a.click()
  URL.revokeObjectURL(url)
}

// 提交备注
const submitComment = async () => {
  if (!commentForm.comment.trim()) {
    message.error('请输入备注内容')
    return
  }
  
  commentLoading.value = true
  try {
    // 这里应该调用添加备注的API
    message.success('备注已添加')
    commentModalVisible.value = false
    commentForm.comment = ''
  } catch (error) {
    message.error('添加备注失败')
  } finally {
    commentLoading.value = false
  }
}
</script>

<style scoped>
.alert-detail {
  padding: 0;
}

.detail-card {
  margin-bottom: 16px;
}

.detail-card:last-of-type {
  margin-bottom: 24px;
}

.history-item {
  padding: 8px 0;
}

.history-action {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 4px;
}

.history-time {
  font-size: 12px;
  color: #999;
}

.history-user,
.history-comment {
  font-size: 12px;
  color: #666;
  margin-bottom: 2px;
}

.action-buttons {
  padding: 16px 0;
  border-top: 1px solid #f0f0f0;
  text-align: center;
}

/* 响应式设计 */
@media (max-width: 768px) {
  .detail-card :deep(.ant-descriptions) {
    font-size: 12px;
  }
  
  .action-buttons {
    text-align: left;
  }
  
  .action-buttons :deep(.ant-space) {
    flex-wrap: wrap;
  }
}
</style>