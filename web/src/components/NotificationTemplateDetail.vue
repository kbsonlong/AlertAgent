<template>
  <div class="notification-template-detail">
    <div v-if="template" class="detail-content">
      <!-- 基本信息 -->
      <div class="detail-section">
        <h3 class="section-title">
          <InfoCircleOutlined />
          基本信息
        </h3>
        
        <a-descriptions :column="2" bordered>
          <a-descriptions-item label="名称">
            {{ template.name }}
          </a-descriptions-item>
          <a-descriptions-item label="类型">
            <a-tag :color="getTypeColor(template.type)">
              <component :is="getTypeIcon(template.type)" />
              {{ getTypeName(template.type) }}
            </a-tag>
          </a-descriptions-item>
          <a-descriptions-item label="状态">
            <a-tag :color="template.enabled ? 'success' : 'default'">
              {{ template.enabled ? '启用' : '禁用' }}
            </a-tag>
          </a-descriptions-item>
          <a-descriptions-item label="创建时间">
            {{ formatDateTime(template.created_at) }}
          </a-descriptions-item>
          <a-descriptions-item label="更新时间">
            {{ formatDateTime(template.updated_at) }}
          </a-descriptions-item>
          <a-descriptions-item label="描述" :span="2">
            {{ template.description || '暂无描述' }}
          </a-descriptions-item>
        </a-descriptions>
      </div>

      <!-- 模板内容 -->
      <div class="detail-section">
        <h3 class="section-title">
          <FileTextOutlined />
          模板内容
        </h3>
        
        <!-- 邮件模板 -->
        <div v-if="template.type === 'email'" class="template-content">
          <div class="template-item">
            <h4>邮件主题</h4>
            <div class="template-code">
              <pre>{{ template.template?.subject || '未设置' }}</pre>
            </div>
          </div>
          <div class="template-item">
            <h4>邮件内容</h4>
            <div class="template-code">
              <pre>{{ template.template?.body || '未设置' }}</pre>
            </div>
          </div>
        </div>
        
        <!-- Webhook 模板 -->
        <div v-else-if="template.type === 'webhook'" class="template-content">
          <div class="template-item">
            <h4>请求体格式：{{ template.template?.format || 'json' }}</h4>
            <div class="template-code">
              <pre>{{ formatJson(template.template?.body) }}</pre>
            </div>
          </div>
        </div>
        
        <!-- 钉钉模板 -->
        <div v-else-if="template.type === 'dingtalk'" class="template-content">
          <div class="template-item">
            <h4>消息类型：{{ template.template?.msgType === 'markdown' ? 'Markdown' : '文本' }}</h4>
            <div v-if="template.template?.title" class="template-item">
              <h5>标题</h5>
              <div class="template-code">
                <pre>{{ template.template.title }}</pre>
              </div>
            </div>
            <div class="template-item">
              <h5>内容</h5>
              <div class="template-code">
                <pre>{{ template.template?.content || '未设置' }}</pre>
              </div>
            </div>
          </div>
        </div>
        
        <!-- 企业微信模板 -->
        <div v-else-if="template.type === 'wechat'" class="template-content">
          <div class="template-item">
            <h4>消息类型：{{ template.template?.msgType === 'markdown' ? 'Markdown' : '文本' }}</h4>
            <div class="template-code">
              <pre>{{ template.template?.content || '未设置' }}</pre>
            </div>
          </div>
        </div>
        
        <!-- Slack 模板 -->
        <div v-else-if="template.type === 'slack'" class="template-content">
          <div class="template-item">
            <h4>消息格式：{{ template.template?.format === 'blocks' ? 'Block Kit' : '纯文本' }}</h4>
            <div class="template-code">
              <pre>{{ template.template?.format === 'blocks' ? formatJson(template.template?.blocks) : template.template?.text || '未设置' }}</pre>
            </div>
          </div>
        </div>
      </div>

      <!-- 使用统计 -->
      <div class="detail-section">
        <h3 class="section-title">
          <BarChartOutlined />
          使用统计
        </h3>
        
        <a-row :gutter="16">
          <a-col :span="6">
            <a-statistic
              title="总发送次数"
              :value="template.stats?.total_sent || 0"
              :value-style="{ color: '#1890ff' }"
            >
              <template #prefix>
                <SendOutlined />
              </template>
            </a-statistic>
          </a-col>
          <a-col :span="6">
            <a-statistic
              title="成功次数"
              :value="template.stats?.success_count || 0"
              :value-style="{ color: '#52c41a' }"
            >
              <template #prefix>
                <CheckCircleOutlined />
              </template>
            </a-statistic>
          </a-col>
          <a-col :span="6">
            <a-statistic
              title="失败次数"
              :value="template.stats?.error_count || 0"
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
              :value="getSuccessRate(template.stats)"
              suffix="%"
              :value-style="{ color: '#52c41a' }"
            >
              <template #prefix>
                <TrophyOutlined />
              </template>
            </a-statistic>
          </a-col>
        </a-row>
      </div>

      <!-- 关联通知组 -->
      <div class="detail-section">
        <h3 class="section-title">
          <TeamOutlined />
          关联通知组
        </h3>
        
        <div v-if="template.groups && template.groups.length > 0" class="groups-list">
          <a-tag
            v-for="group in template.groups"
            :key="group.id"
            color="blue"
            class="group-tag"
          >
            {{ group.name }}
          </a-tag>
        </div>
        <a-empty v-else description="暂无关联通知组" />
      </div>

      <!-- 操作历史 -->
      <div class="detail-section">
        <h3 class="section-title">
          <HistoryOutlined />
          操作历史
        </h3>
        
        <a-timeline>
          <a-timeline-item
            v-for="(log, index) in template.logs"
            :key="index"
            :color="getLogColor(log.action)"
          >
            <template #dot>
              <component :is="getLogIcon(log.action)" />
            </template>
            <div class="log-content">
              <div class="log-header">
                <span class="log-action">{{ getLogActionName(log.action) }}</span>
                <span class="log-time">{{ formatDateTime(log.created_at) }}</span>
              </div>
              <div class="log-user">操作人：{{ log.user || '系统' }}</div>
              <div v-if="log.description" class="log-description">{{ log.description }}</div>
            </div>
          </a-timeline-item>
        </a-timeline>
      </div>
    </div>
    
    <div v-else class="detail-empty">
      <a-empty description="暂无模板详情" />
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import {
  Descriptions,
  Tag,
  Statistic,
  Row,
  Col,
  Timeline,
  Empty
} from 'ant-design-vue'
import {
  InfoCircleOutlined,
  FileTextOutlined,
  BarChartOutlined,
  TeamOutlined,
  HistoryOutlined,
  MailOutlined,
  ApiOutlined,
  MessageOutlined,
  WechatOutlined,
  SlackOutlined,
  SendOutlined,
  CheckCircleOutlined,
  CloseCircleOutlined,
  TrophyOutlined,
  EditOutlined,
  DeleteOutlined,
  CopyOutlined
} from '@ant-design/icons-vue'
import { formatDateTime } from '@/utils/datetime'
import type { NotificationTemplate } from '@/types'

const ADescriptions = Descriptions
const ADescriptionsItem = Descriptions.Item
const ATag = Tag
const AStatistic = Statistic
const ARow = Row
const ACol = Col
const ATimeline = Timeline
const ATimelineItem = Timeline.Item
const AEmpty = Empty

interface Props {
  template: NotificationTemplate | null
}

defineProps<Props>()

// 获取类型颜色
const getTypeColor = (type: string) => {
  const colors = {
    email: 'blue',
    webhook: 'green',
    dingtalk: 'orange',
    wechat: 'cyan',
    slack: 'purple'
  }
  return colors[type] || 'default'
}

// 获取类型图标
const getTypeIcon = (type: string) => {
  const icons = {
    email: MailOutlined,
    webhook: ApiOutlined,
    dingtalk: MessageOutlined,
    wechat: WechatOutlined,
    slack: SlackOutlined
  }
  return icons[type] || InfoCircleOutlined
}

// 获取类型名称
const getTypeName = (type: string) => {
  const names = {
    email: '邮件',
    webhook: 'Webhook',
    dingtalk: '钉钉',
    wechat: '企业微信',
    slack: 'Slack'
  }
  return names[type] || type
}

// 格式化JSON
const formatJson = (jsonStr: string) => {
  if (!jsonStr) return '未设置'
  try {
    return JSON.stringify(JSON.parse(jsonStr), null, 2)
  } catch {
    return jsonStr
  }
}

// 计算成功率
const getSuccessRate = (stats: any) => {
  if (!stats || !stats.total_sent) return 0
  return Math.round((stats.success_count / stats.total_sent) * 100)
}

// 获取日志颜色
const getLogColor = (action: string) => {
  const colors = {
    create: 'green',
    update: 'blue',
    delete: 'red',
    test: 'orange',
    send: 'purple'
  }
  return colors[action] || 'gray'
}

// 获取日志图标
const getLogIcon = (action: string) => {
  const icons = {
    create: CheckCircleOutlined,
    update: EditOutlined,
    delete: DeleteOutlined,
    test: SendOutlined,
    send: SendOutlined
  }
  return icons[action] || InfoCircleOutlined
}

// 获取日志操作名称
const getLogActionName = (action: string) => {
  const names = {
    create: '创建模板',
    update: '更新模板',
    delete: '删除模板',
    test: '测试模板',
    send: '发送通知'
  }
  return names[action] || action
}
</script>

<style scoped>
.notification-template-detail {
  padding: 0;
}

.detail-content {
  display: flex;
  flex-direction: column;
  gap: 24px;
}

.detail-section {
  padding: 24px;
  background: #fafafa;
  border-radius: 8px;
}

.section-title {
  display: flex;
  align-items: center;
  gap: 8px;
  margin: 0 0 20px 0;
  font-size: 16px;
  font-weight: 600;
  color: #262626;
}

.template-content {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.template-item h4,
.template-item h5 {
  margin: 0 0 12px 0;
  font-size: 14px;
  font-weight: 600;
  color: #595959;
}

.template-code {
  background: #f5f5f5;
  border: 1px solid #e8e8e8;
  border-radius: 6px;
  padding: 16px;
  overflow-x: auto;
}

.template-code pre {
  margin: 0;
  font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
  font-size: 12px;
  line-height: 1.5;
  white-space: pre-wrap;
  word-break: break-all;
}

.groups-list {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.group-tag {
  margin: 0;
}

.log-content {
  padding: 8px 0;
}

.log-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 4px;
}

.log-action {
  font-weight: 600;
  color: #262626;
}

.log-time {
  font-size: 12px;
  color: #8c8c8c;
}

.log-user {
  font-size: 13px;
  color: #595959;
  margin-bottom: 4px;
}

.log-description {
  font-size: 13px;
  color: #8c8c8c;
}

.detail-empty {
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: 300px;
}

/* 响应式设计 */
@media (max-width: 768px) {
  .detail-section {
    padding: 16px;
  }
  
  .template-code {
    padding: 12px;
  }
  
  .template-code pre {
    font-size: 11px;
  }
  
  .log-header {
    flex-direction: column;
    align-items: flex-start;
    gap: 4px;
  }
  
  .groups-list {
    gap: 6px;
  }
}
</style>