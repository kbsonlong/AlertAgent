<template>
  <div class="notification-group-detail">
    <!-- 基本信息 -->
    <div class="detail-section">
      <h3 class="section-title">
        <InfoCircleOutlined />
        基本信息
      </h3>
      <div class="info-grid">
        <div class="info-item">
          <label>名称</label>
          <div class="info-value">{{ group.name }}</div>
        </div>
        <div class="info-item">
          <label>类型</label>
          <div class="info-value">
            <a-tag :color="getTypeColor(group.type)">
              <component :is="getTypeIcon(group.type)" />
              {{ getTypeText(group.type) }}
            </a-tag>
          </div>
        </div>
        <div class="info-item">
          <label>状态</label>
          <div class="info-value">
            <a-badge
              :status="group.enabled ? 'success' : 'default'"
              :text="group.enabled ? '启用' : '禁用'"
            />
          </div>
        </div>
        <div class="info-item">
          <label>描述</label>
          <div class="info-value">{{ group.description || '-' }}</div>
        </div>
        <div class="info-item">
          <label>创建时间</label>
          <div class="info-value">{{ formatDateTime(group.createdAt) }}</div>
        </div>
        <div class="info-item">
          <label>更新时间</label>
          <div class="info-value">{{ formatDateTime(group.updatedAt) }}</div>
        </div>
      </div>
    </div>

    <!-- 配置信息 -->
    <div class="detail-section">
      <h3 class="section-title">
        <SettingOutlined />
        配置信息
      </h3>
      
      <!-- 邮件配置 -->
      <div v-if="group.type === 'email'" class="config-content">
        <div class="config-group">
          <h4>SMTP 配置</h4>
          <div class="info-grid">
            <div class="info-item">
              <label>SMTP 服务器</label>
              <div class="info-value">{{ group.config?.smtp?.host || '-' }}</div>
            </div>
            <div class="info-item">
              <label>端口</label>
              <div class="info-value">{{ group.config?.smtp?.port || '-' }}</div>
            </div>
            <div class="info-item">
              <label>用户名</label>
              <div class="info-value">{{ group.config?.smtp?.username || '-' }}</div>
            </div>
            <div class="info-item">
              <label>TLS</label>
              <div class="info-value">
                <a-badge
                  :status="group.config?.smtp?.tls ? 'success' : 'default'"
                  :text="group.config?.smtp?.tls ? '启用' : '禁用'"
                />
              </div>
            </div>
          </div>
        </div>
        
        <div class="config-group">
          <h4>发送配置</h4>
          <div class="info-grid">
            <div class="info-item">
              <label>发件人</label>
              <div class="info-value">{{ group.config?.from || '-' }}</div>
            </div>
            <div class="info-item">
              <label>收件人</label>
              <div class="info-value">
                <a-tag v-for="to in group.config?.to" :key="to" class="recipient-tag">
                  {{ to }}
                </a-tag>
                <span v-if="!group.config?.to?.length">-</span>
              </div>
            </div>
            <div class="info-item">
              <label>抄送</label>
              <div class="info-value">
                <a-tag v-for="cc in group.config?.cc" :key="cc" class="recipient-tag">
                  {{ cc }}
                </a-tag>
                <span v-if="!group.config?.cc?.length">-</span>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- Webhook 配置 -->
      <div v-else-if="group.type === 'webhook'" class="config-content">
        <div class="config-group">
          <h4>Webhook 配置</h4>
          <div class="info-grid">
            <div class="info-item">
              <label>URL</label>
              <div class="info-value">
                <a-typography-text copyable>{{ group.config?.url || '-' }}</a-typography-text>
              </div>
            </div>
            <div class="info-item">
              <label>方法</label>
              <div class="info-value">
                <a-tag color="blue">{{ group.config?.method || 'POST' }}</a-tag>
              </div>
            </div>
            <div class="info-item">
              <label>超时时间</label>
              <div class="info-value">{{ group.config?.timeout || 30 }}s</div>
            </div>
            <div class="info-item">
              <label>重试次数</label>
              <div class="info-value">{{ group.config?.retries || 3 }}</div>
            </div>
          </div>
        </div>
        
        <div class="config-group" v-if="group.config?.headers">
          <h4>请求头</h4>
          <div class="headers-list">
            <div v-for="(value, key) in group.config.headers" :key="key" class="header-item">
              <span class="header-key">{{ key }}:</span>
              <span class="header-value">{{ value }}</span>
            </div>
          </div>
        </div>
      </div>

      <!-- 钉钉配置 -->
      <div v-else-if="group.type === 'dingtalk'" class="config-content">
        <div class="config-group">
          <h4>钉钉配置</h4>
          <div class="info-grid">
            <div class="info-item">
              <label>Webhook URL</label>
              <div class="info-value">
                <a-typography-text copyable>{{ group.config?.webhook || '-' }}</a-typography-text>
              </div>
            </div>
            <div class="info-item">
              <label>密钥</label>
              <div class="info-value">
                <span v-if="group.config?.secret">已配置</span>
                <span v-else>-</span>
              </div>
            </div>
            <div class="info-item">
              <label>@所有人</label>
              <div class="info-value">
                <a-badge
                  :status="group.config?.atAll ? 'success' : 'default'"
                  :text="group.config?.atAll ? '是' : '否'"
                />
              </div>
            </div>
            <div class="info-item" v-if="group.config?.atMobiles?.length">
              <label>@手机号</label>
              <div class="info-value">
                <a-tag v-for="mobile in group.config.atMobiles" :key="mobile" class="mobile-tag">
                  {{ mobile }}
                </a-tag>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- 企业微信配置 -->
      <div v-else-if="group.type === 'wechat'" class="config-content">
        <div class="config-group">
          <h4>企业微信配置</h4>
          <div class="info-grid">
            <div class="info-item">
              <label>Webhook URL</label>
              <div class="info-value">
                <a-typography-text copyable>{{ group.config?.webhook || '-' }}</a-typography-text>
              </div>
            </div>
            <div class="info-item">
              <label>提及用户</label>
              <div class="info-value">
                <a-tag v-for="user in group.config?.mentionedList" :key="user" class="mention-tag">
                  {{ user }}
                </a-tag>
                <span v-if="!group.config?.mentionedList?.length">-</span>
              </div>
            </div>
            <div class="info-item">
              <label>提及手机号</label>
              <div class="info-value">
                <a-tag v-for="mobile in group.config?.mentionedMobileList" :key="mobile" class="mobile-tag">
                  {{ mobile }}
                </a-tag>
                <span v-if="!group.config?.mentionedMobileList?.length">-</span>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- Slack 配置 -->
      <div v-else-if="group.type === 'slack'" class="config-content">
        <div class="config-group">
          <h4>Slack 配置</h4>
          <div class="info-grid">
            <div class="info-item">
              <label>Webhook URL</label>
              <div class="info-value">
                <a-typography-text copyable>{{ group.config?.webhook || '-' }}</a-typography-text>
              </div>
            </div>
            <div class="info-item">
              <label>频道</label>
              <div class="info-value">{{ group.config?.channel || '-' }}</div>
            </div>
            <div class="info-item">
              <label>用户名</label>
              <div class="info-value">{{ group.config?.username || '-' }}</div>
            </div>
            <div class="info-item">
              <label>图标</label>
              <div class="info-value">{{ group.config?.iconEmoji || group.config?.iconUrl || '-' }}</div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- 发送统计 -->
    <div class="detail-section">
      <h3 class="section-title">
        <BarChartOutlined />
        发送统计
      </h3>
      <div class="stats-content">
        <a-row :gutter="16">
          <a-col :span="6">
            <a-statistic
              title="总发送数"
              :value="group.stats?.totalSent || 0"
              :value-style="{ color: '#1890ff' }"
            />
          </a-col>
          <a-col :span="6">
            <a-statistic
              title="成功数"
              :value="group.stats?.successCount || 0"
              :value-style="{ color: '#52c41a' }"
            />
          </a-col>
          <a-col :span="6">
            <a-statistic
              title="失败数"
              :value="group.stats?.failedCount || 0"
              :value-style="{ color: '#ff4d4f' }"
            />
          </a-col>
          <a-col :span="6">
            <a-statistic
              title="成功率"
              :value="getSuccessRate(group.stats)"
              suffix="%"
              :value-style="{ color: getSuccessRateColor(group.stats) }"
            />
          </a-col>
        </a-row>
        
        <div class="stats-chart" v-if="group.stats?.recentSends?.length">
          <h4>最近发送记录</h4>
          <div class="recent-sends">
            <div v-for="send in group.stats.recentSends" :key="send.id" class="send-item">
              <div class="send-info">
                <div class="send-time">{{ formatDateTime(send.sentAt) }}</div>
                <div class="send-status">
                  <a-badge
                    :status="send.success ? 'success' : 'error'"
                    :text="send.success ? '成功' : '失败'"
                  />
                </div>
              </div>
              <div class="send-details">
                <div class="send-subject">{{ send.subject || send.title || '-' }}</div>
                <div class="send-error" v-if="!send.success && send.error">
                  {{ send.error }}
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- 标签和注解 -->
    <div class="detail-section" v-if="group.labels || group.annotations">
      <h3 class="section-title">
        <TagsOutlined />
        标签和注解
      </h3>
      
      <div class="labels-section" v-if="group.labels && Object.keys(group.labels).length">
        <h4>标签</h4>
        <div class="labels-list">
          <a-tag v-for="(value, key) in group.labels" :key="key" class="label-tag">
            {{ key }}: {{ value }}
          </a-tag>
        </div>
      </div>
      
      <div class="annotations-section" v-if="group.annotations && Object.keys(group.annotations).length">
        <h4>注解</h4>
        <div class="annotations-list">
          <div v-for="(value, key) in group.annotations" :key="key" class="annotation-item">
            <span class="annotation-key">{{ key }}:</span>
            <span class="annotation-value">{{ value }}</span>
          </div>
        </div>
      </div>
    </div>

    <!-- 操作历史 -->
    <div class="detail-section" v-if="group.history?.length">
      <h3 class="section-title">
        <HistoryOutlined />
        操作历史
      </h3>
      <div class="history-content">
        <a-timeline>
          <a-timeline-item
            v-for="item in group.history"
            :key="item.id"
            :color="getHistoryColor(item.action)"
          >
            <template #dot>
              <component :is="getHistoryIcon(item.action)" />
            </template>
            <div class="history-item">
              <div class="history-header">
                <span class="history-action">{{ getHistoryText(item.action) }}</span>
                <span class="history-time">{{ formatDateTime(item.createdAt) }}</span>
              </div>
              <div class="history-details">
                <div class="history-user">操作人: {{ item.user || 'System' }}</div>
                <div class="history-description" v-if="item.description">
                  {{ item.description }}
                </div>
              </div>
            </div>
          </a-timeline-item>
        </a-timeline>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import {
  Badge,
  Tag,
  Typography,
  Row,
  Col,
  Statistic,
  Timeline
} from 'ant-design-vue'
import {
  InfoCircleOutlined,
  SettingOutlined,
  BarChartOutlined,
  TagsOutlined,
  HistoryOutlined,
  MailOutlined,
  ApiOutlined,
  MessageOutlined,
  WechatOutlined,
  SlackOutlined,
  EditOutlined,
  DeleteOutlined,
  CheckCircleOutlined,
  ExclamationCircleOutlined
} from '@ant-design/icons-vue'
import { formatDateTime } from '@/utils/datetime'
import type { NotificationGroup } from '@/types'

const ABadge = Badge
const ATag = Tag
const ATypographyText = Typography.Text
const ARow = Row
const ACol = Col
const AStatistic = Statistic
const ATimeline = Timeline
const ATimelineItem = Timeline.Item

interface Props {
  group: NotificationGroup
}

const props = defineProps<Props>()

const emit = defineEmits<{
  edit: [group: NotificationGroup]
  test: [group: NotificationGroup]
  delete: [group: NotificationGroup]
  close: []
}>()

// 获取类型图标
const getTypeIcon = (type: string) => {
  const iconMap: Record<string, any> = {
    email: MailOutlined,
    webhook: ApiOutlined,
    dingtalk: MessageOutlined,
    wechat: WechatOutlined,
    slack: SlackOutlined
  }
  return iconMap[type] || MessageOutlined
}

// 获取类型颜色
const getTypeColor = (type: string) => {
  const colorMap: Record<string, string> = {
    email: 'blue',
    webhook: 'green',
    dingtalk: 'orange',
    wechat: 'cyan',
    slack: 'purple'
  }
  return colorMap[type] || 'default'
}

// 获取类型文本
const getTypeText = (type: string) => {
  const textMap: Record<string, string> = {
    email: '邮件',
    webhook: 'Webhook',
    dingtalk: '钉钉',
    wechat: '企业微信',
    slack: 'Slack'
  }
  return textMap[type] || type
}

// 计算成功率
const getSuccessRate = (stats: any) => {
  if (!stats || !stats.totalSent) return 0
  return Math.round((stats.successCount / stats.totalSent) * 100)
}

// 获取成功率颜色
const getSuccessRateColor = (stats: any) => {
  const rate = getSuccessRate(stats)
  if (rate >= 95) return '#52c41a'
  if (rate >= 80) return '#faad14'
  return '#ff4d4f'
}

// 获取历史操作颜色
const getHistoryColor = (action: string) => {
  const colorMap: Record<string, string> = {
    create: 'green',
    update: 'blue',
    delete: 'red',
    test: 'orange',
    enable: 'green',
    disable: 'gray'
  }
  return colorMap[action] || 'blue'
}

// 获取历史操作图标
const getHistoryIcon = (action: string) => {
  const iconMap: Record<string, any> = {
    create: CheckCircleOutlined,
    update: EditOutlined,
    delete: DeleteOutlined,
    test: ExclamationCircleOutlined,
    enable: CheckCircleOutlined,
    disable: ExclamationCircleOutlined
  }
  return iconMap[action] || EditOutlined
}

// 获取历史操作文本
const getHistoryText = (action: string) => {
  const textMap: Record<string, string> = {
    create: '创建',
    update: '更新',
    delete: '删除',
    test: '测试',
    enable: '启用',
    disable: '禁用'
  }
  return textMap[action] || action
}
</script>

<style scoped>
.notification-group-detail {
  padding: 0;
}

.detail-section {
  margin-bottom: 32px;
}

.detail-section:last-child {
  margin-bottom: 0;
}

.section-title {
  display: flex;
  align-items: center;
  gap: 8px;
  margin: 0 0 16px 0;
  font-size: 16px;
  font-weight: 600;
  color: #262626;
}

.info-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
  gap: 16px;
}

.info-item {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.info-item label {
  font-size: 12px;
  color: #8c8c8c;
  font-weight: 500;
}

.info-value {
  font-size: 14px;
  color: #262626;
  word-break: break-all;
}

.config-content {
  background: #fafafa;
  padding: 16px;
  border-radius: 6px;
}

.config-group {
  margin-bottom: 24px;
}

.config-group:last-child {
  margin-bottom: 0;
}

.config-group h4 {
  margin: 0 0 12px 0;
  font-size: 14px;
  font-weight: 600;
  color: #595959;
}

.recipient-tag,
.mobile-tag,
.mention-tag,
.label-tag {
  margin: 2px 4px 2px 0;
}

.headers-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.header-item {
  display: flex;
  gap: 8px;
  font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
  font-size: 12px;
}

.header-key {
  color: #1890ff;
  font-weight: 600;
}

.header-value {
  color: #262626;
}

.stats-content {
  background: #fafafa;
  padding: 16px;
  border-radius: 6px;
}

.stats-chart {
  margin-top: 24px;
}

.stats-chart h4 {
  margin: 0 0 16px 0;
  font-size: 14px;
  font-weight: 600;
  color: #595959;
}

.recent-sends {
  display: flex;
  flex-direction: column;
  gap: 12px;
  max-height: 300px;
  overflow-y: auto;
}

.send-item {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  padding: 12px;
  background: white;
  border-radius: 4px;
  border: 1px solid #f0f0f0;
}

.send-info {
  display: flex;
  flex-direction: column;
  gap: 4px;
  min-width: 120px;
}

.send-time {
  font-size: 12px;
  color: #8c8c8c;
}

.send-details {
  flex: 1;
  margin-left: 16px;
}

.send-subject {
  font-size: 14px;
  color: #262626;
  margin-bottom: 4px;
}

.send-error {
  font-size: 12px;
  color: #ff4d4f;
}

.labels-section,
.annotations-section {
  margin-bottom: 16px;
}

.labels-section:last-child,
.annotations-section:last-child {
  margin-bottom: 0;
}

.labels-section h4,
.annotations-section h4 {
  margin: 0 0 12px 0;
  font-size: 14px;
  font-weight: 600;
  color: #595959;
}

.labels-list {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.annotations-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.annotation-item {
  display: flex;
  gap: 8px;
  font-size: 14px;
}

.annotation-key {
  color: #1890ff;
  font-weight: 500;
  min-width: 120px;
}

.annotation-value {
  color: #262626;
  word-break: break-all;
}

.history-content {
  max-height: 400px;
  overflow-y: auto;
}

.history-item {
  padding-bottom: 8px;
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
  font-size: 12px;
  color: #595959;
}

.history-user {
  margin-bottom: 2px;
}

.history-description {
  color: #8c8c8c;
}

/* 响应式设计 */
@media (max-width: 768px) {
  .info-grid {
    grid-template-columns: 1fr;
  }
  
  .send-item {
    flex-direction: column;
    gap: 8px;
  }
  
  .send-details {
    margin-left: 0;
  }
  
  .history-header {
    flex-direction: column;
    align-items: flex-start;
    gap: 4px;
  }
}
</style>