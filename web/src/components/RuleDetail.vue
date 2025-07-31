<template>
  <div class="rule-detail">
    <!-- 基本信息 -->
    <a-card title="基本信息" class="detail-card">
      <a-descriptions :column="2" bordered>
        <a-descriptions-item label="规则名称">
          {{ rule.name }}
        </a-descriptions-item>
        <a-descriptions-item label="状态">
          <a-tag :color="rule.enabled ? 'green' : 'red'">
            {{ rule.enabled ? '启用' : '禁用' }}
          </a-tag>
        </a-descriptions-item>
        <a-descriptions-item label="严重程度">
          <a-tag :color="getSeverityColor(rule.severity)">
            {{ getSeverityText(rule.severity) }}
          </a-tag>
        </a-descriptions-item>
        <a-descriptions-item label="数据源">
          {{ getProviderName(rule.provider_id) }}
        </a-descriptions-item>
        <a-descriptions-item label="评估间隔">
          {{ rule.evaluation_interval }}
        </a-descriptions-item>
        <a-descriptions-item label="持续时间">
          {{ rule.for_duration }}
        </a-descriptions-item>
        <a-descriptions-item label="规则组">
          {{ rule.rule_group || 'default' }}
        </a-descriptions-item>
        <a-descriptions-item label="触发次数">
          <a-badge
            :count="rule.firing_count"
            :number-style="{ backgroundColor: rule.firing_count > 0 ? '#ff4d4f' : '#52c41a' }"
          >
            <span>{{ rule.firing_count || 0 }}</span>
          </a-badge>
        </a-descriptions-item>
        <a-descriptions-item label="创建时间" :span="2">
          {{ formatDateTime(rule.created_at) }}
        </a-descriptions-item>
        <a-descriptions-item label="更新时间" :span="2">
          {{ formatDateTime(rule.updated_at) }}
        </a-descriptions-item>
        <a-descriptions-item label="描述" :span="2">
          {{ rule.description || '暂无描述' }}
        </a-descriptions-item>
      </a-descriptions>
    </a-card>

    <!-- 查询配置 -->
    <a-card title="查询配置" class="detail-card">
      <div class="query-section">
        <h4>查询表达式</h4>
        <div class="code-block">
          <pre>{{ rule.query }}</pre>
          <a-button type="link" size="small" @click="copyToClipboard(rule.query)">
            <template #icon><CopyOutlined /></template>
            复制
          </a-button>
        </div>
      </div>
      
      <div class="query-section">
        <h4>条件表达式</h4>
        <div class="condition-block">
          <a-tag>{{ rule.condition }}</a-tag>
        </div>
      </div>
    </a-card>

    <!-- 标签和注释 -->
    <a-card title="标签和注释" class="detail-card">
      <a-row :gutter="16">
        <a-col :span="12">
          <h4>标签</h4>
          <div class="labels-section">
            <a-tag
              v-for="(value, key) in rule.labels"
              :key="key"
              class="label-tag"
            >
              {{ key }}={{ value }}
            </a-tag>
            <div v-if="!rule.labels || Object.keys(rule.labels).length === 0" class="no-data">
              暂无标签
            </div>
          </div>
        </a-col>
        <a-col :span="12">
          <h4>注释</h4>
          <div class="annotations-section">
            <div
              v-for="(value, key) in rule.annotations"
              :key="key"
              class="annotation-item"
            >
              <strong>{{ key }}:</strong>
              <span>{{ value }}</span>
            </div>
            <div v-if="!rule.annotations || Object.keys(rule.annotations).length === 0" class="no-data">
              暂无注释
            </div>
          </div>
        </a-col>
      </a-row>
    </a-card>

    <!-- 通知配置 -->
    <a-card title="通知配置" class="detail-card">
      <a-descriptions :column="2" bordered>
        <a-descriptions-item label="通知组">
          <a-space v-if="rule.notification_groups && rule.notification_groups.length > 0">
            <a-tag v-for="group in rule.notification_groups" :key="group">
              {{ group }}
            </a-tag>
          </a-space>
          <span v-else class="no-data">未配置</span>
        </a-descriptions-item>
        <a-descriptions-item label="通知模板">
          {{ rule.notification_template || '默认模板' }}
        </a-descriptions-item>
        <a-descriptions-item label="通知间隔">
          {{ rule.notification_interval || '5m' }}
        </a-descriptions-item>
        <a-descriptions-item label="保持触发">
          <a-tag :color="rule.keep_firing_for ? 'green' : 'default'">
            {{ rule.keep_firing_for ? '是' : '否' }}
          </a-tag>
        </a-descriptions-item>
      </a-descriptions>
    </a-card>

    <!-- 外部标签 -->
    <a-card title="外部标签" class="detail-card" v-if="rule.external_labels">
      <div class="external-labels">
        <pre>{{ JSON.stringify(rule.external_labels, null, 2) }}</pre>
      </div>
    </a-card>

    <!-- 规则历史 -->
    <a-card title="规则历史" class="detail-card">
      <a-timeline>
        <a-timeline-item color="green">
          <div class="timeline-content">
            <div class="timeline-title">规则创建</div>
            <div class="timeline-time">{{ formatDateTime(rule.created_at) }}</div>
          </div>
        </a-timeline-item>
        <a-timeline-item color="blue" v-if="rule.updated_at !== rule.created_at">
          <div class="timeline-content">
            <div class="timeline-title">规则更新</div>
            <div class="timeline-time">{{ formatDateTime(rule.updated_at) }}</div>
          </div>
        </a-timeline-item>
        <a-timeline-item color="orange" v-if="rule.last_evaluation">
          <div class="timeline-content">
            <div class="timeline-title">最后评估</div>
            <div class="timeline-time">{{ formatDateTime(rule.last_evaluation) }}</div>
          </div>
        </a-timeline-item>
        <a-timeline-item color="red" v-if="rule.last_firing">
          <div class="timeline-content">
            <div class="timeline-title">最后触发</div>
            <div class="timeline-time">{{ formatDateTime(rule.last_firing) }}</div>
          </div>
        </a-timeline-item>
      </a-timeline>
    </a-card>

    <!-- 操作按钮 -->
    <div class="action-buttons">
      <a-space>
        <a-button type="primary" @click="$emit('edit', rule)">
          <template #icon><EditOutlined /></template>
          编辑
        </a-button>
        <a-button @click="$emit('test', rule)">
          <template #icon><PlayCircleOutlined /></template>
          测试
        </a-button>
        <a-button @click="duplicateRule">
          <template #icon><CopyOutlined /></template>
          复制
        </a-button>
        <a-button @click="exportRule">
          <template #icon><ExportOutlined /></template>
          导出
        </a-button>
        <a-button danger @click="$emit('delete', rule)">
          <template #icon><DeleteOutlined /></template>
          删除
        </a-button>
      </a-space>
    </div>
  </div>
</template>

<script setup lang="ts">
import {
  Card,
  Descriptions,
  Tag,
  Badge,
  Row,
  Col,
  Space,
  Timeline,
  Button,
  message
} from 'ant-design-vue'
import {
  CopyOutlined,
  EditOutlined,
  PlayCircleOutlined,
  ExportOutlined,
  DeleteOutlined
} from '@ant-design/icons-vue'
import { formatDateTime } from '@/utils/datetime'
import type { Rule } from '@/types'

const ACard = Card
const ADescriptions = Descriptions
const ADescriptionsItem = Descriptions.Item
const ATag = Tag
const ABadge = Badge
const ARow = Row
const ACol = Col
const ASpace = Space
const ATimeline = Timeline
const ATimelineItem = Timeline.Item
const AButton = Button

interface Props {
  rule: Rule
}

interface Emits {
  (e: 'edit', rule: Rule): void
  (e: 'test', rule: Rule): void
  (e: 'delete', rule: Rule): void
}

const props = defineProps<Props>()
const emit = defineEmits<Emits>()

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

// 获取数据源名称
const getProviderName = (providerId: string) => {
  // 这里应该从store或props中获取provider信息
  return providerId || '未知'
}

// 复制到剪贴板
const copyToClipboard = async (text: string) => {
  try {
    await navigator.clipboard.writeText(text)
    message.success('已复制到剪贴板')
  } catch (error) {
    // 降级方案
    const textArea = document.createElement('textarea')
    textArea.value = text
    document.body.appendChild(textArea)
    textArea.select()
    document.execCommand('copy')
    document.body.removeChild(textArea)
    message.success('已复制到剪贴板')
  }
}

// 复制规则
const duplicateRule = () => {
  const newRule = {
    ...props.rule,
    name: `${props.rule.name} (副本)`,
    id: undefined
  }
  emit('edit', newRule as Rule)
}

// 导出规则
const exportRule = () => {
  const data = JSON.stringify(props.rule, null, 2)
  const blob = new Blob([data], { type: 'application/json' })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = `rule-${props.rule.name}-${Date.now()}.json`
  a.click()
  URL.revokeObjectURL(url)
  message.success('规则已导出')
}
</script>

<style scoped>
.rule-detail {
  padding: 0;
}

.detail-card {
  margin-bottom: 16px;
}

.detail-card:last-of-type {
  margin-bottom: 24px;
}

.query-section {
  margin-bottom: 24px;
}

.query-section:last-child {
  margin-bottom: 0;
}

.query-section h4 {
  margin-bottom: 12px;
  color: #1890ff;
}

.code-block {
  position: relative;
  background: #f5f5f5;
  border: 1px solid #d9d9d9;
  border-radius: 6px;
  padding: 12px;
}

.code-block pre {
  margin: 0;
  font-family: 'Courier New', monospace;
  font-size: 12px;
  line-height: 1.6;
  white-space: pre-wrap;
  word-break: break-all;
}

.code-block .ant-btn {
  position: absolute;
  top: 8px;
  right: 8px;
}

.condition-block {
  padding: 8px 0;
}

.labels-section,
.annotations-section {
  min-height: 60px;
}

.labels-section h4,
.annotations-section h4 {
  margin-bottom: 12px;
  color: #1890ff;
}

.label-tag {
  margin-bottom: 8px;
  margin-right: 8px;
  font-family: 'Courier New', monospace;
}

.annotation-item {
  margin-bottom: 12px;
  padding: 8px;
  background: #f9f9f9;
  border-radius: 4px;
}

.annotation-item strong {
  display: block;
  margin-bottom: 4px;
  color: #1890ff;
}

.annotation-item span {
  color: #666;
  line-height: 1.6;
}

.external-labels {
  background: #f5f5f5;
  border: 1px solid #d9d9d9;
  border-radius: 6px;
  padding: 12px;
}

.external-labels pre {
  margin: 0;
  font-family: 'Courier New', monospace;
  font-size: 12px;
  line-height: 1.6;
}

.timeline-content {
  padding: 4px 0;
}

.timeline-title {
  font-weight: 500;
  margin-bottom: 4px;
}

.timeline-time {
  font-size: 12px;
  color: #999;
}

.no-data {
  color: #999;
  font-style: italic;
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
  
  .code-block .ant-btn {
    position: static;
    margin-top: 8px;
  }
  
  .labels-section,
  .annotations-section {
    margin-bottom: 16px;
  }
}
</style>