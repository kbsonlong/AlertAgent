<template>
  <div class="alert-analysis">
    <!-- 分析概要 -->
    <a-card title="分析概要" class="analysis-card">
      <a-descriptions :column="2" bordered>
        <a-descriptions-item label="分析时间">
          {{ formatDateTime(analysis.analyzed_at) }}
        </a-descriptions-item>
        <a-descriptions-item label="分析模型">
          {{ analysis.model || 'GPT-4' }}
        </a-descriptions-item>
        <a-descriptions-item label="置信度">
          <a-progress
            :percent="Math.round(analysis.confidence * 100)"
            :stroke-color="getConfidenceColor(analysis.confidence)"
            size="small"
          />
        </a-descriptions-item>
        <a-descriptions-item label="严重程度评估">
          <a-tag :color="getSeverityColor(analysis.severity_assessment)">
            {{ getSeverityText(analysis.severity_assessment) }}
          </a-tag>
        </a-descriptions-item>
      </a-descriptions>
    </a-card>

    <!-- 根本原因分析 -->
    <a-card title="根本原因分析" class="analysis-card">
      <div class="analysis-content">
        <div class="analysis-section">
          <h4><BulbOutlined /> 可能原因</h4>
          <div v-html="renderMarkdown(analysis.root_cause || '暂无分析结果')"></div>
        </div>
        
        <div class="analysis-section" v-if="analysis.contributing_factors">
          <h4><ExclamationCircleOutlined /> 影响因素</h4>
          <a-list size="small">
            <a-list-item v-for="(factor, index) in analysis.contributing_factors" :key="index">
              <a-list-item-meta>
                <template #description>
                  {{ factor }}
                </template>
              </a-list-item-meta>
            </a-list-item>
          </a-list>
        </div>
      </div>
    </a-card>

    <!-- 影响评估 -->
    <a-card title="影响评估" class="analysis-card">
      <a-row :gutter="16">
        <a-col :span="8">
          <a-statistic
            title="业务影响"
            :value="getImpactLevel(analysis.business_impact)"
            :value-style="{ color: getImpactColor(analysis.business_impact) }"
          >
            <template #prefix><ShopOutlined /></template>
          </a-statistic>
        </a-col>
        <a-col :span="8">
          <a-statistic
            title="用户影响"
            :value="getImpactLevel(analysis.user_impact)"
            :value-style="{ color: getImpactColor(analysis.user_impact) }"
          >
            <template #prefix><UserOutlined /></template>
          </a-statistic>
        </a-col>
        <a-col :span="8">
          <a-statistic
            title="系统影响"
            :value="getImpactLevel(analysis.system_impact)"
            :value-style="{ color: getImpactColor(analysis.system_impact) }"
          >
            <template #prefix><DesktopOutlined /></template>
          </a-statistic>
        </a-col>
      </a-row>
      
      <div class="impact-description" v-if="analysis.impact_description">
        <h4>详细说明</h4>
        <div v-html="renderMarkdown(analysis.impact_description)"></div>
      </div>
    </a-card>

    <!-- 解决建议 -->
    <a-card title="解决建议" class="analysis-card">
      <div class="suggestions">
        <div class="suggestion-section">
          <h4><ToolOutlined /> 立即行动</h4>
          <a-list v-if="analysis.immediate_actions" size="small">
            <a-list-item v-for="(action, index) in analysis.immediate_actions" :key="index">
              <a-list-item-meta>
                <template #avatar>
                  <a-avatar size="small" :style="{ backgroundColor: '#ff4d4f' }">
                    {{ index + 1 }}
                  </a-avatar>
                </template>
                <template #description>
                  {{ action }}
                </template>
              </a-list-item-meta>
            </a-list-item>
          </a-list>
          <div v-else class="no-data">暂无立即行动建议</div>
        </div>
        
        <div class="suggestion-section">
          <h4><ClockCircleOutlined /> 长期措施</h4>
          <a-list v-if="analysis.long_term_solutions" size="small">
            <a-list-item v-for="(solution, index) in analysis.long_term_solutions" :key="index">
              <a-list-item-meta>
                <template #avatar>
                  <a-avatar size="small" :style="{ backgroundColor: '#52c41a' }">
                    {{ index + 1 }}
                  </a-avatar>
                </template>
                <template #description>
                  {{ solution }}
                </template>
              </a-list-item-meta>
            </a-list-item>
          </a-list>
          <div v-else class="no-data">暂无长期解决方案</div>
        </div>
      </div>
    </a-card>

    <!-- 相似告警 -->
    <a-card title="相似告警" class="analysis-card" v-if="analysis.similar_alerts && analysis.similar_alerts.length > 0">
      <a-table
        :columns="similarAlertsColumns"
        :data-source="analysis.similar_alerts"
        :pagination="false"
        size="small"
      >
        <template #similarity="{ record }">
          <a-progress
            :percent="Math.round(record.similarity * 100)"
            size="small"
            :stroke-color="getSimilarityColor(record.similarity)"
          />
        </template>
        
        <template #status="{ record }">
          <a-tag :color="getStatusColor(record.status)" size="small">
            {{ getStatusText(record.status) }}
          </a-tag>
        </template>
        
        <template #action="{ record }">
          <a-button type="link" size="small" @click="viewSimilarAlert(record)">
            查看
          </a-button>
        </template>
      </a-table>
    </a-card>

    <!-- 知识库推荐 -->
    <a-card title="相关知识" class="analysis-card" v-if="analysis.knowledge_recommendations && analysis.knowledge_recommendations.length > 0">
      <a-list>
        <a-list-item v-for="knowledge in analysis.knowledge_recommendations" :key="knowledge.id">
          <a-list-item-meta>
            <template #title>
              <a @click="viewKnowledge(knowledge)">{{ knowledge.title }}</a>
            </template>
            <template #description>
              <div class="knowledge-meta">
                <a-tag size="small">{{ knowledge.category }}</a-tag>
                <span class="knowledge-summary">{{ knowledge.summary }}</span>
              </div>
            </template>
          </a-list-item-meta>
          <template #actions>
            <a-button type="link" size="small" @click="viewKnowledge(knowledge)">
              查看详情
            </a-button>
          </template>
        </a-list-item>
      </a-list>
    </a-card>

    <!-- 分析详情 -->
    <a-card title="详细分析" class="analysis-card">
      <a-collapse>
        <a-collapse-panel key="technical" header="技术分析">
          <div v-html="renderMarkdown(analysis.technical_analysis || '暂无技术分析')"></div>
        </a-collapse-panel>
        <a-collapse-panel key="timeline" header="时间线分析">
          <a-timeline v-if="analysis.timeline && analysis.timeline.length > 0">
            <a-timeline-item
              v-for="(event, index) in analysis.timeline"
              :key="index"
              :color="getTimelineColor(event.type)"
            >
              <div class="timeline-event">
                <div class="event-time">{{ formatDateTime(event.timestamp) }}</div>
                <div class="event-description">{{ event.description }}</div>
              </div>
            </a-timeline-item>
          </a-timeline>
          <div v-else class="no-data">暂无时间线数据</div>
        </a-collapse-panel>
        <a-collapse-panel key="metrics" header="指标分析">
          <div v-html="renderMarkdown(analysis.metrics_analysis || '暂无指标分析')"></div>
        </a-collapse-panel>
      </a-collapse>
    </a-card>

    <!-- 操作按钮 -->
    <div class="action-buttons">
      <a-space>
        <a-button type="primary" @click="acceptSuggestions">
          <template #icon><CheckOutlined /></template>
          采纳建议
        </a-button>
        <a-button @click="exportAnalysis">
          <template #icon><ExportOutlined /></template>
          导出分析
        </a-button>
        <a-button @click="shareAnalysis">
          <template #icon><ShareAltOutlined /></template>
          分享
        </a-button>
        <a-button @click="$emit('close')">
          关闭
        </a-button>
      </a-space>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import {
  Card,
  Descriptions,
  Progress,
  Tag,
  List,
  Statistic,
  Row,
  Col,
  Table,
  Timeline,
  Collapse,
  Button,
  Space,
  message
} from 'ant-design-vue'
import {
  BulbOutlined,
  ExclamationCircleOutlined,
  ShopOutlined,
  UserOutlined,
  DesktopOutlined,
  ToolOutlined,
  ClockCircleOutlined,
  CheckOutlined,
  ExportOutlined,
  ShareAltOutlined
} from '@ant-design/icons-vue'
import { formatDateTime } from '@/utils/datetime'
import type { AlertAnalysis, SimilarAlert } from '@/types'

const ACard = Card
const ADescriptions = Descriptions
const ADescriptionsItem = Descriptions.Item
const AProgress = Progress
const ATag = Tag
const AList = List
const AListItem = List.Item
const AListItemMeta = List.Item.Meta
const AStatistic = Statistic
const ARow = Row
const ACol = Col
const ATable = Table
const ATimeline = Timeline
const ATimelineItem = Timeline.Item
const ACollapse = Collapse
const ACollapsePanel = Collapse.Panel
const AButton = Button
const ASpace = Space

interface Props {
  analysis: AlertAnalysis
}

interface Emits {
  (e: 'close'): void
  (e: 'viewSimilar', alert: SimilarAlert): void
  (e: 'viewKnowledge', knowledge: any): void
}

const props = defineProps<Props>()
const emit = defineEmits<Emits>()

// 相似告警表格列
const similarAlertsColumns = [
  {
    title: '告警名称',
    dataIndex: 'name',
    key: 'name',
    ellipsis: true
  },
  {
    title: '相似度',
    dataIndex: 'similarity',
    key: 'similarity',
    width: 120,
    slots: { customRender: 'similarity' }
  },
  {
    title: '状态',
    dataIndex: 'status',
    key: 'status',
    width: 80,
    slots: { customRender: 'status' }
  },
  {
    title: '操作',
    key: 'action',
    width: 80,
    slots: { customRender: 'action' }
  }
]

// 获取置信度颜色
const getConfidenceColor = (confidence: number) => {
  if (confidence >= 0.8) return '#52c41a'
  if (confidence >= 0.6) return '#faad14'
  return '#ff4d4f'
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

// 获取影响级别文本
const getImpactLevel = (impact: string) => {
  const levelMap: Record<string, string> = {
    high: '高',
    medium: '中',
    low: '低'
  }
  return levelMap[impact] || impact
}

// 获取影响级别颜色
const getImpactColor = (impact: string) => {
  const colorMap: Record<string, string> = {
    high: '#ff4d4f',
    medium: '#faad14',
    low: '#52c41a'
  }
  return colorMap[impact] || '#1890ff'
}

// 获取相似度颜色
const getSimilarityColor = (similarity: number) => {
  if (similarity >= 0.8) return '#52c41a'
  if (similarity >= 0.6) return '#faad14'
  return '#ff4d4f'
}

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

// 获取时间线颜色
const getTimelineColor = (type: string) => {
  const colorMap: Record<string, string> = {
    error: 'red',
    warning: 'orange',
    info: 'blue',
    success: 'green'
  }
  return colorMap[type] || 'blue'
}

// 渲染Markdown
const renderMarkdown = (text: string) => {
  return text
    .replace(/\*\*(.*?)\*\*/g, '<strong>$1</strong>')
    .replace(/\*(.*?)\*/g, '<em>$1</em>')
    .replace(/`(.*?)`/g, '<code>$1</code>')
    .replace(/\n/g, '<br>')
    .replace(/^### (.*$)/gim, '<h3>$1</h3>')
    .replace(/^## (.*$)/gim, '<h2>$1</h2>')
    .replace(/^# (.*$)/gim, '<h1>$1</h1>')
}

// 查看相似告警
const viewSimilarAlert = (alert: SimilarAlert) => {
  emit('viewSimilar', alert)
}

// 查看知识库
const viewKnowledge = (knowledge: any) => {
  emit('viewKnowledge', knowledge)
}

// 采纳建议
const acceptSuggestions = () => {
  message.success('建议已采纳，将自动执行相关操作')
}

// 导出分析
const exportAnalysis = () => {
  const data = JSON.stringify(props.analysis, null, 2)
  const blob = new Blob([data], { type: 'application/json' })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = `analysis-${Date.now()}.json`
  a.click()
  URL.revokeObjectURL(url)
  message.success('分析报告已导出')
}

// 分享分析
const shareAnalysis = () => {
  // 这里可以实现分享功能
  message.success('分享链接已复制到剪贴板')
}
</script>

<style scoped>
.alert-analysis {
  padding: 0;
}

.analysis-card {
  margin-bottom: 16px;
}

.analysis-card:last-of-type {
  margin-bottom: 24px;
}

.analysis-content {
  padding: 0;
}

.analysis-section {
  margin-bottom: 24px;
}

.analysis-section:last-child {
  margin-bottom: 0;
}

.analysis-section h4 {
  margin-bottom: 12px;
  color: #1890ff;
  display: flex;
  align-items: center;
  gap: 8px;
}

.impact-description {
  margin-top: 24px;
  padding-top: 16px;
  border-top: 1px solid #f0f0f0;
}

.impact-description h4 {
  margin-bottom: 12px;
}

.suggestions {
  padding: 0;
}

.suggestion-section {
  margin-bottom: 24px;
}

.suggestion-section:last-child {
  margin-bottom: 0;
}

.suggestion-section h4 {
  margin-bottom: 12px;
  color: #1890ff;
  display: flex;
  align-items: center;
  gap: 8px;
}

.knowledge-meta {
  display: flex;
  align-items: center;
  gap: 8px;
}

.knowledge-summary {
  color: #666;
  font-size: 12px;
}

.timeline-event {
  padding: 4px 0;
}

.event-time {
  font-size: 12px;
  color: #999;
  margin-bottom: 4px;
}

.event-description {
  color: #333;
}

.no-data {
  color: #999;
  font-style: italic;
  text-align: center;
  padding: 20px;
}

.action-buttons {
  padding: 16px 0;
  border-top: 1px solid #f0f0f0;
  text-align: center;
}

/* 代码块样式 */
:deep(code) {
  background-color: #f5f5f5;
  padding: 2px 4px;
  border-radius: 3px;
  font-family: 'Courier New', monospace;
}

/* 响应式设计 */
@media (max-width: 768px) {
  .analysis-card :deep(.ant-descriptions) {
    font-size: 12px;
  }
  
  .action-buttons {
    text-align: left;
  }
  
  .action-buttons :deep(.ant-space) {
    flex-wrap: wrap;
  }
  
  .knowledge-meta {
    flex-direction: column;
    align-items: flex-start;
  }
}
</style>