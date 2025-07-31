<template>
  <div class="notification-template-form">
    <a-form
      ref="formRef"
      :model="formData"
      :rules="formRules"
      layout="vertical"
      @finish="handleSubmit"
    >
      <!-- 基本信息 -->
      <div class="form-section">
        <h3 class="section-title">
          <InfoCircleOutlined />
          基本信息
        </h3>
        
        <a-row :gutter="16">
          <a-col :span="12">
            <a-form-item label="模板名称" name="name">
              <a-input
                v-model:value="formData.name"
                placeholder="请输入模板名称"
                :maxlength="50"
                show-count
              />
            </a-form-item>
          </a-col>
          <a-col :span="12">
            <a-form-item label="模板类型" name="type">
              <a-select
                v-model:value="formData.type"
                placeholder="请选择模板类型"
                @change="handleTypeChange"
              >
                <a-select-option value="email">
                  <MailOutlined /> 邮件
                </a-select-option>
                <a-select-option value="webhook">
                  <ApiOutlined /> Webhook
                </a-select-option>
                <a-select-option value="dingtalk">
                  <MessageOutlined /> 钉钉
                </a-select-option>
                <a-select-option value="wechat">
                  <WechatOutlined /> 企业微信
                </a-select-option>
                <a-select-option value="slack">
                  <SlackOutlined /> Slack
                </a-select-option>
              </a-select>
            </a-form-item>
          </a-col>
        </a-row>
        
        <a-form-item label="描述" name="description">
          <a-textarea
            v-model:value="formData.description"
            placeholder="请输入模板描述"
            :rows="3"
            :maxlength="200"
            show-count
          />
        </a-form-item>
      </div>

      <!-- 模板配置 -->
      <div class="form-section">
        <h3 class="section-title">
          <FileTextOutlined />
          模板配置
        </h3>
        
        <!-- 邮件模板配置 -->
        <div v-if="formData.type === 'email'" class="template-config">
          <a-form-item label="邮件主题" name="template.subject">
            <a-input
              v-model:value="formData.template.subject"
              placeholder="请输入邮件主题模板"
            />
          </a-form-item>
          <a-form-item label="邮件内容" name="template.body">
            <a-textarea
              v-model:value="formData.template.body"
              placeholder="请输入邮件内容模板，支持变量：{{.GroupLabels.alertname}}, {{.CommonAnnotations.summary}} 等"
              :rows="8"
            />
          </a-form-item>
          <div class="template-help">
            <a-alert
              message="模板变量说明"
              description="可使用 Go template 语法，常用变量：{{.GroupLabels.alertname}} (告警名称)、{{.CommonAnnotations.summary}} (摘要)、{{.CommonAnnotations.description}} (描述)"
              type="info"
              show-icon
            />
          </div>
        </div>
        
        <!-- Webhook 模板配置 -->
        <div v-else-if="formData.type === 'webhook'" class="template-config">
          <a-form-item label="请求体格式" name="template.format">
            <a-select v-model:value="formData.template.format" placeholder="选择格式">
              <a-select-option value="json">JSON</a-select-option>
              <a-select-option value="form">Form Data</a-select-option>
            </a-select>
          </a-form-item>
          <a-form-item label="请求体内容" name="template.body">
            <a-textarea
              v-model:value="formData.template.body"
              placeholder="请输入 Webhook 请求体模板"
              :rows="10"
            />
          </a-form-item>
          <div class="template-help">
            <a-alert
              message="Webhook 模板说明"
              description="支持 JSON 格式的请求体，可使用模板变量进行动态替换"
              type="info"
              show-icon
            />
          </div>
        </div>
        
        <!-- 钉钉模板配置 -->
        <div v-else-if="formData.type === 'dingtalk'" class="template-config">
          <a-form-item label="消息类型" name="template.msgType">
            <a-select v-model:value="formData.template.msgType" placeholder="选择消息类型">
              <a-select-option value="text">文本消息</a-select-option>
              <a-select-option value="markdown">Markdown 消息</a-select-option>
            </a-select>
          </a-form-item>
          <a-form-item v-if="formData.template.msgType === 'markdown'" label="标题" name="template.title">
            <a-input
              v-model:value="formData.template.title"
              placeholder="请输入消息标题"
            />
          </a-form-item>
          <a-form-item label="消息内容" name="template.content">
            <a-textarea
              v-model:value="formData.template.content"
              :placeholder="formData.template.msgType === 'markdown' ? '请输入 Markdown 格式的消息内容' : '请输入文本消息内容'"
              :rows="8"
            />
          </a-form-item>
        </div>
        
        <!-- 企业微信模板配置 -->
        <div v-else-if="formData.type === 'wechat'" class="template-config">
          <a-form-item label="消息类型" name="template.msgType">
            <a-select v-model:value="formData.template.msgType" placeholder="选择消息类型">
              <a-select-option value="text">文本消息</a-select-option>
              <a-select-option value="markdown">Markdown 消息</a-select-option>
            </a-select>
          </a-form-item>
          <a-form-item label="消息内容" name="template.content">
            <a-textarea
              v-model:value="formData.template.content"
              :placeholder="formData.template.msgType === 'markdown' ? '请输入 Markdown 格式的消息内容' : '请输入文本消息内容'"
              :rows="8"
            />
          </a-form-item>
        </div>
        
        <!-- Slack 模板配置 -->
        <div v-else-if="formData.type === 'slack'" class="template-config">
          <a-form-item label="消息文本" name="template.text">
            <a-textarea
              v-model:value="formData.template.text"
              placeholder="请输入 Slack 消息文本"
              :rows="6"
            />
          </a-form-item>
          <a-form-item label="Blocks (可选)" name="template.blocks">
            <a-textarea
              v-model:value="formData.template.blocks"
              placeholder="请输入 Slack Blocks JSON 格式内容"
              :rows="8"
            />
          </a-form-item>
        </div>
      </div>

      <!-- 预览区域 -->
      <div v-if="formData.type" class="form-section">
        <h3 class="section-title">
          <EyeOutlined />
          模板预览
        </h3>
        
        <div class="template-preview">
          <div v-if="formData.type === 'email'" class="email-preview">
            <div class="preview-item">
              <strong>主题：</strong>
              <div class="preview-content">{{ formData.template.subject || '未设置' }}</div>
            </div>
            <div class="preview-item">
              <strong>内容：</strong>
              <div class="preview-content">{{ formData.template.body || '未设置' }}</div>
            </div>
          </div>
          
          <div v-else-if="formData.type === 'webhook'" class="webhook-preview">
            <div class="preview-item">
              <strong>格式：</strong> {{ formData.template.format || 'json' }}
            </div>
            <div class="preview-item">
              <strong>请求体：</strong>
              <pre class="preview-code">{{ formatJson(formData.template.body) }}</pre>
            </div>
          </div>
          
          <div v-else-if="formData.type === 'dingtalk'" class="dingtalk-preview">
            <div class="preview-item">
              <strong>类型：</strong> {{ formData.template.msgType === 'markdown' ? 'Markdown' : '文本' }}
            </div>
            <div v-if="formData.template.title" class="preview-item">
              <strong>标题：</strong> {{ formData.template.title }}
            </div>
            <div class="preview-item">
              <strong>内容：</strong>
              <div class="preview-content">{{ formData.template.content || '未设置' }}</div>
            </div>
          </div>
          
          <div v-else-if="formData.type === 'wechat'" class="wechat-preview">
            <div class="preview-item">
              <strong>类型：</strong> {{ formData.template.msgType === 'markdown' ? 'Markdown' : '文本' }}
            </div>
            <div class="preview-item">
              <strong>内容：</strong>
              <div class="preview-content">{{ formData.template.content || '未设置' }}</div>
            </div>
          </div>
          
          <div v-else-if="formData.type === 'slack'" class="slack-preview">
            <div class="preview-item">
              <strong>文本：</strong>
              <div class="preview-content">{{ formData.template.text || '未设置' }}</div>
            </div>
            <div v-if="formData.template.blocks" class="preview-item">
              <strong>Blocks：</strong>
              <pre class="preview-code">{{ formatJson(formData.template.blocks) }}</pre>
            </div>
          </div>
        </div>
      </div>

      <!-- 表单操作 -->
      <div class="form-actions">
        <a-space>
          <a-button @click="handleCancel">取消</a-button>
          <a-button type="primary" html-type="submit" :loading="submitting">
            {{ mode === 'create' ? '创建' : '更新' }}
          </a-button>
        </a-space>
      </div>
    </a-form>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, watch } from 'vue'
import {
  Form,
  FormItem,
  Input,
  Textarea,
  Select,
  SelectOption,
  Button,
  Space,
  Row,
  Col,
  Alert
} from 'ant-design-vue'
import {
  InfoCircleOutlined,
  FileTextOutlined,
  EyeOutlined,
  MailOutlined,
  ApiOutlined,
  MessageOutlined,
  WechatOutlined,
  SlackOutlined
} from '@ant-design/icons-vue'
import type { NotificationTemplate } from '@/types'
import type { FormInstance } from 'ant-design-vue'

// 组件属性
interface Props {
  template?: NotificationTemplate | null
  mode: 'create' | 'edit'
}

const props = withDefaults(defineProps<Props>(), {
  template: null,
  mode: 'create'
})

// 组件事件
interface Emits {
  submit: [data: any]
  cancel: []
}

const emit = defineEmits<Emits>()

// 表单引用
const formRef = ref<FormInstance>()
const submitting = ref(false)

// 表单数据
const formData = reactive({
  name: '',
  type: 'email' as 'email' | 'webhook' | 'dingtalk' | 'wechat' | 'slack',
  description: '',
  enabled: true,
  template: {
    subject: '',
    body: '',
    content: '',
    text: '',
    format: 'json',
    msgType: 'text',
    title: '',
    blocks: ''
  }
})

// 表单验证规则
const formRules = computed(() => ({
  name: [
    { required: true, message: '请输入模板名称' },
    { min: 2, max: 50, message: '模板名称长度为 2-50 个字符' }
  ],
  type: [
    { required: true, message: '请选择模板类型' }
  ],
  'template.subject': [
    { 
      required: formData.type === 'email', 
      message: '请输入邮件主题' 
    }
  ],
  'template.body': [
    { 
      required: formData.type === 'email' || formData.type === 'webhook', 
      message: '请输入模板内容' 
    }
  ],
  'template.content': [
    { 
      required: formData.type === 'dingtalk' || formData.type === 'wechat', 
      message: '请输入消息内容' 
    }
  ],
  'template.text': [
    { 
      required: formData.type === 'slack', 
      message: '请输入消息文本' 
    }
  ]
}))

// 格式化 JSON
const formatJson = (str: string) => {
  if (!str) return ''
  try {
    return JSON.stringify(JSON.parse(str), null, 2)
  } catch {
    return str
  }
}

// 处理类型变化
const handleTypeChange = (type: string) => {
  // 清空模板配置
  formData.template = {
    subject: '',
    body: '',
    content: '',
    text: '',
    format: 'json',
    msgType: 'text',
    title: '',
    blocks: ''
  }
}

// 处理表单提交
const handleSubmit = async () => {
  try {
    await formRef.value?.validate()
    submitting.value = true
    
    const submitData = {
      name: formData.name,
      type: formData.type,
      description: formData.description,
      enabled: formData.enabled,
      template: { ...formData.template }
    }
    
    // 清理不需要的字段
    if (formData.type !== 'email') {
      delete submitData.template.subject
      delete submitData.template.body
    }
    if (formData.type !== 'webhook') {
      delete submitData.template.format
    }
    if (formData.type !== 'dingtalk' && formData.type !== 'wechat') {
      delete submitData.template.content
      delete submitData.template.msgType
    }
    if (formData.type !== 'dingtalk') {
      delete submitData.template.title
    }
    if (formData.type !== 'slack') {
      delete submitData.template.text
      delete submitData.template.blocks
    }
    
    emit('submit', submitData)
  } catch (error) {
    console.error('表单验证失败:', error)
  } finally {
    submitting.value = false
  }
}

// 处理取消
const handleCancel = () => {
  emit('cancel')
}

// 初始化表单数据
const initFormData = () => {
  if (props.template) {
    formData.name = props.template.name
    formData.type = props.template.type
    formData.description = props.template.description || ''
    formData.enabled = props.template.enabled
    
    if (props.template.template) {
      formData.template = {
        subject: props.template.template.subject || '',
        body: props.template.template.body || '',
        content: props.template.template.content || '',
        text: props.template.template.text || '',
        format: props.template.template.format || 'json',
        msgType: props.template.template.msgType || 'text',
        title: props.template.template.title || '',
        blocks: props.template.template.blocks || ''
      }
    }
  }
}

// 监听模板变化
watch(
  () => props.template,
  () => {
    initFormData()
  },
  { immediate: true }
)
</script>

<style scoped>
.notification-template-form {
  padding: 20px;
}

.form-section {
  margin-bottom: 32px;
}

.section-title {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 16px;
  font-size: 16px;
  font-weight: 600;
  color: #262626;
}

.template-config {
  background: #fafafa;
  padding: 16px;
  border-radius: 6px;
  margin-bottom: 16px;
}

.template-help {
  margin-top: 16px;
}

.template-preview {
  background: #f5f5f5;
  padding: 16px;
  border-radius: 6px;
  border: 1px solid #d9d9d9;
}

.preview-item {
  margin-bottom: 12px;
}

.preview-item:last-child {
  margin-bottom: 0;
}

.preview-content {
  margin-top: 4px;
  padding: 8px;
  background: white;
  border-radius: 4px;
  border: 1px solid #e8e8e8;
  white-space: pre-wrap;
  word-break: break-word;
}

.preview-code {
  margin-top: 4px;
  padding: 12px;
  background: white;
  border-radius: 4px;
  border: 1px solid #e8e8e8;
  font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
  font-size: 12px;
  line-height: 1.4;
  overflow-x: auto;
}

.form-actions {
  padding-top: 24px;
  border-top: 1px solid #f0f0f0;
  text-align: right;
}

:deep(.ant-form-item-label) {
  font-weight: 500;
}

:deep(.ant-select-selector) {
  border-radius: 6px;
}

:deep(.ant-input) {
  border-radius: 6px;
}

:deep(.ant-input-affix-wrapper) {
  border-radius: 6px;
}
</style>