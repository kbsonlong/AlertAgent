<template>
  <div class="notification-template-preview">
    <div v-if="template" class="preview-content">
      <!-- 邮件模板预览 -->
      <div v-if="template.type === 'email'" class="email-preview">
        <div class="preview-section">
          <h4 class="preview-title">
            <MailOutlined />
            邮件预览
          </h4>
          <div class="email-container">
            <div class="email-header">
              <div class="email-field">
                <label>主题：</label>
                <div class="email-subject">{{ renderedSubject || template.template?.subject || '未设置主题' }}</div>
              </div>
            </div>
            <div class="email-body">
              <div class="email-content" v-html="renderedContent || template.template?.body || '未设置内容'"></div>
            </div>
          </div>
        </div>
      </div>

      <!-- Webhook 模板预览 -->
      <div v-else-if="template.type === 'webhook'" class="webhook-preview">
        <div class="preview-section">
          <h4 class="preview-title">
            <ApiOutlined />
            Webhook 预览
          </h4>
          <div class="webhook-container">
            <div class="webhook-info">
              <div class="info-item">
                <label>格式：</label>
                <a-tag>{{ template.template?.format || 'json' }}</a-tag>
              </div>
            </div>
            <div class="webhook-body">
              <label>请求体：</label>
              <pre class="code-block">{{ formatWebhookBody() }}</pre>
            </div>
          </div>
        </div>
      </div>

      <!-- 钉钉模板预览 -->
      <div v-else-if="template.type === 'dingtalk'" class="dingtalk-preview">
        <div class="preview-section">
          <h4 class="preview-title">
            <MessageOutlined />
            钉钉消息预览
          </h4>
          <div class="dingtalk-container">
            <div class="message-info">
              <div class="info-item">
                <label>消息类型：</label>
                <a-tag :color="template.template?.msgType === 'markdown' ? 'blue' : 'default'">
                  {{ template.template?.msgType === 'markdown' ? 'Markdown' : '文本' }}
                </a-tag>
              </div>
            </div>
            <div v-if="template.template?.title" class="message-title">
              <label>标题：</label>
              <div class="title-content">{{ template.template.title }}</div>
            </div>
            <div class="message-content">
              <label>内容：</label>
              <div class="content-body" :class="{ 'markdown-content': template.template?.msgType === 'markdown' }">
                <div v-if="template.template?.msgType === 'markdown'" v-html="renderMarkdown(renderedContent || template.template?.content)"></div>
                <div v-else class="text-content">{{ renderedContent || template.template?.content || '未设置内容' }}</div>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- 企业微信模板预览 -->
      <div v-else-if="template.type === 'wechat'" class="wechat-preview">
        <div class="preview-section">
          <h4 class="preview-title">
            <WechatOutlined />
            企业微信消息预览
          </h4>
          <div class="wechat-container">
            <div class="message-info">
              <div class="info-item">
                <label>消息类型：</label>
                <a-tag :color="template.template?.msgType === 'markdown' ? 'blue' : 'default'">
                  {{ template.template?.msgType === 'markdown' ? 'Markdown' : '文本' }}
                </a-tag>
              </div>
            </div>
            <div class="message-content">
              <label>内容：</label>
              <div class="content-body" :class="{ 'markdown-content': template.template?.msgType === 'markdown' }">
                <div v-if="template.template?.msgType === 'markdown'" v-html="renderMarkdown(renderedContent || template.template?.content)"></div>
                <div v-else class="text-content">{{ renderedContent || template.template?.content || '未设置内容' }}</div>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- Slack 模板预览 -->
      <div v-else-if="template.type === 'slack'" class="slack-preview">
        <div class="preview-section">
          <h4 class="preview-title">
            <SlackOutlined />
            Slack 消息预览
          </h4>
          <div class="slack-container">
            <div class="message-info">
              <div class="info-item">
                <label>消息格式：</label>
                <a-tag :color="template.template?.format === 'blocks' ? 'blue' : 'default'">
                  {{ template.template?.format === 'blocks' ? 'Block Kit' : '纯文本' }}
                </a-tag>
              </div>
            </div>
            <div class="message-content">
              <div v-if="template.template?.format === 'blocks'" class="blocks-content">
                <label>Blocks：</label>
                <pre class="code-block">{{ formatJson(template.template?.blocks) }}</pre>
              </div>
              <div v-else class="text-content">
                <label>文本：</label>
                <div class="content-body">{{ renderedContent || template.template?.text || '未设置内容' }}</div>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- 预览变量信息 -->
      <div v-if="previewData" class="variables-section">
        <h4 class="preview-title">
          <SettingOutlined />
          变量信息
        </h4>
        <div class="variables-info">
          <div v-if="previewData.variables_used?.length" class="variable-group">
            <label>已使用变量：</label>
            <div class="variable-tags">
              <a-tag v-for="variable in previewData.variables_used" :key="variable" color="success">
                {{ variable }}
              </a-tag>
            </div>
          </div>
          <div v-if="previewData.variables_missing?.length" class="variable-group">
            <label>缺失变量：</label>
            <div class="variable-tags">
              <a-tag v-for="variable in previewData.variables_missing" :key="variable" color="error">
                {{ variable }}
              </a-tag>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- 加载状态 -->
    <div v-else-if="loading" class="loading-container">
      <a-spin size="large" tip="正在生成预览..." />
    </div>

    <!-- 空状态 -->
    <div v-else class="empty-container">
      <a-empty description="暂无预览内容" />
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { Tag, Spin, Empty } from 'ant-design-vue'
import {
  MailOutlined,
  ApiOutlined,
  MessageOutlined,
  WechatOutlined,
  SlackOutlined,
  SettingOutlined
} from '@ant-design/icons-vue'
import type { NotificationTemplate } from '@/types'
import { previewNotificationTemplate } from '@/services/notification'
import { message } from 'ant-design-vue'

// 组件属性
interface Props {
  template: NotificationTemplate
  variables?: Record<string, any>
}

const props = withDefaults(defineProps<Props>(), {
  variables: () => ({})
})

// 响应式数据
const loading = ref(false)
const renderedContent = ref('')
const renderedSubject = ref('')
const previewData = ref<any>(null)

// 计算属性
const templateVariables = computed(() => {
  return props.variables || {}
})

// 方法
const loadPreview = async () => {
  if (!props.template?.id) return
  
  loading.value = true
  try {
    const response = await previewNotificationTemplate(props.template.id, templateVariables.value)
    if (response.data) {
      renderedContent.value = response.data.content || ''
      renderedSubject.value = response.data.subject || ''
      previewData.value = response.data
    }
  } catch (error) {
    console.error('预览模板失败:', error)
    message.error('预览模板失败')
  } finally {
    loading.value = false
  }
}

const formatJson = (data: any) => {
  if (!data) return ''
  try {
    if (typeof data === 'string') {
      return JSON.stringify(JSON.parse(data), null, 2)
    }
    return JSON.stringify(data, null, 2)
  } catch {
    return data
  }
}

const formatWebhookBody = () => {
  const body = renderedContent.value || props.template.template?.body
  if (!body) return '未设置请求体'
  
  try {
    if (typeof body === 'string') {
      return JSON.stringify(JSON.parse(body), null, 2)
    }
    return JSON.stringify(body, null, 2)
  } catch {
    return body
  }
}

const renderMarkdown = (content: string) => {
  if (!content) return ''
  // 简单的 Markdown 渲染，可以根据需要使用更完整的 Markdown 解析器
  return content
    .replace(/\*\*(.*?)\*\*/g, '<strong>$1</strong>')
    .replace(/\*(.*?)\*/g, '<em>$1</em>')
    .replace(/`(.*?)`/g, '<code>$1</code>')
    .replace(/\n/g, '<br>')
}

// 生命周期
onMounted(() => {
  loadPreview()
})

// 监听模板变化
watch(
  () => [props.template, props.variables],
  () => {
    loadPreview()
  },
  { deep: true }
)
</script>

<style scoped>
.notification-template-preview {
  padding: 20px;
}

.preview-content {
  max-height: 600px;
  overflow-y: auto;
}

.preview-section {
  margin-bottom: 24px;
}

.preview-title {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 16px;
  font-size: 16px;
  font-weight: 600;
  color: #262626;
}

/* 邮件预览样式 */
.email-container {
  border: 1px solid #d9d9d9;
  border-radius: 6px;
  background: white;
  overflow: hidden;
}

.email-header {
  padding: 16px;
  background: #fafafa;
  border-bottom: 1px solid #f0f0f0;
}

.email-field {
  display: flex;
  align-items: center;
  gap: 8px;
}

.email-field label {
  font-weight: 500;
  color: #666;
  min-width: 60px;
}

.email-subject {
  font-weight: 600;
  color: #262626;
}

.email-body {
  padding: 16px;
}

.email-content {
  line-height: 1.6;
  color: #262626;
  white-space: pre-wrap;
  word-break: break-word;
}

/* Webhook 预览样式 */
.webhook-container {
  border: 1px solid #d9d9d9;
  border-radius: 6px;
  background: white;
  padding: 16px;
}

.webhook-info {
  margin-bottom: 16px;
}

.info-item {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 8px;
}

.info-item label {
  font-weight: 500;
  color: #666;
  min-width: 60px;
}

.webhook-body label {
  display: block;
  font-weight: 500;
  color: #666;
  margin-bottom: 8px;
}

.code-block {
  background: #f5f5f5;
  border: 1px solid #d9d9d9;
  border-radius: 4px;
  padding: 12px;
  font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
  font-size: 12px;
  line-height: 1.4;
  overflow-x: auto;
  white-space: pre;
}

/* 钉钉/企业微信预览样式 */
.dingtalk-container,
.wechat-container {
  border: 1px solid #d9d9d9;
  border-radius: 6px;
  background: white;
  padding: 16px;
}

.message-info {
  margin-bottom: 16px;
}

.message-title {
  margin-bottom: 12px;
}

.message-title label {
  display: block;
  font-weight: 500;
  color: #666;
  margin-bottom: 4px;
}

.title-content {
  font-weight: 600;
  color: #262626;
  padding: 8px;
  background: #f5f5f5;
  border-radius: 4px;
}

.message-content label {
  display: block;
  font-weight: 500;
  color: #666;
  margin-bottom: 8px;
}

.content-body {
  padding: 12px;
  background: #f5f5f5;
  border-radius: 4px;
  border: 1px solid #e8e8e8;
}

.text-content {
  line-height: 1.6;
  color: #262626;
  white-space: pre-wrap;
  word-break: break-word;
}

.markdown-content {
  line-height: 1.6;
  color: #262626;
}

.markdown-content :deep(strong) {
  font-weight: 600;
}

.markdown-content :deep(em) {
  font-style: italic;
}

.markdown-content :deep(code) {
  background: #f0f0f0;
  padding: 2px 4px;
  border-radius: 2px;
  font-family: monospace;
}

/* Slack 预览样式 */
.slack-container {
  border: 1px solid #d9d9d9;
  border-radius: 6px;
  background: white;
  padding: 16px;
}

.blocks-content label,
.text-content label {
  display: block;
  font-weight: 500;
  color: #666;
  margin-bottom: 8px;
}

/* 变量信息样式 */
.variables-section {
  margin-top: 24px;
  padding-top: 24px;
  border-top: 1px solid #f0f0f0;
}

.variables-info {
  background: #fafafa;
  padding: 16px;
  border-radius: 6px;
  border: 1px solid #e8e8e8;
}

.variable-group {
  margin-bottom: 12px;
}

.variable-group:last-child {
  margin-bottom: 0;
}

.variable-group label {
  display: block;
  font-weight: 500;
  color: #666;
  margin-bottom: 8px;
}

.variable-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 4px;
}

/* 加载和空状态样式 */
.loading-container,
.empty-container {
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: 200px;
}

/* 响应式设计 */
@media (max-width: 768px) {
  .notification-template-preview {
    padding: 12px;
  }
  
  .email-header,
  .email-body,
  .webhook-container,
  .dingtalk-container,
  .wechat-container,
  .slack-container {
    padding: 12px;
  }
  
  .code-block {
    padding: 8px;
    font-size: 11px;
  }
  
  .variable-tags {
    gap: 2px;
  }
}
</style>