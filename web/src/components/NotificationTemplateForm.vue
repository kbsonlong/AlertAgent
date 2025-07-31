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
            <a-form-item label="名称" name="name">
              <a-input
                v-model:value="formData.name"
                placeholder="请输入模板名称"
              />
            </a-form-item>
          </a-col>
          <a-col :span="12">
            <a-form-item label="类型" name="type">
              <a-select
                v-model:value="formData.type"
                placeholder="请选择模板类型"
              >
                <a-select-option value="email">
                  <MailOutlined /> 邮件模板
                </a-select-option>
                <a-select-option value="webhook">
                  <ApiOutlined /> Webhook模板
                </a-select-option>
                <a-select-option value="dingtalk">
                  <MessageOutlined /> 钉钉模板
                </a-select-option>
                <a-select-option value="wechat">
                  <WechatOutlined /> 企业微信模板
                </a-select-option>
                <a-select-option value="slack">
                  <SlackOutlined /> Slack模板
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
          />
        </a-form-item>
        
        <a-form-item label="状态" name="enabled">
          <a-switch
            v-model:checked="formData.enabled"
            checked-children="启用"
            un-checked-children="禁用"
          />
        </a-form-item>
      </div>

      <!-- 邮件模板 -->
      <div v-if="formData.type === 'email'" class="form-section">
        <h3 class="section-title">
          <MailOutlined />
          邮件模板
        </h3>
        
        <a-form-item label="邮件主题" name="['template', 'subject']">
          <a-input
            v-model:value="formData.template.subject"
            placeholder="请输入邮件主题模板"
          />
          <div class="template-help">
            <a-typography-text type="secondary">
              支持变量：{{ '{{' }} .GroupLabels.alertname {{ '}}' }}、{{ '{{' }} .CommonLabels.severity {{ '}}' }}、{{ '{{' }} .Alerts | len {{ '}}' }} 等
            </a-typography-text>
          </div>
        </a-form-item>
        
        <a-form-item label="邮件内容" name="['template', 'body']">
          <div class="template-editor">
            <a-tabs v-model:activeKey="emailActiveTab">
              <a-tab-pane key="edit" tab="编辑">
                <a-textarea
                  v-model:value="formData.template.body"
                  placeholder="请输入邮件内容模板"
                  :rows="15"
                  class="template-textarea"
                />
              </a-tab-pane>
              <a-tab-pane key="preview" tab="预览">
                <div class="template-preview" v-html="emailPreview"></div>
              </a-tab-pane>
            </a-tabs>
          </div>
          <div class="template-help">
            <a-typography-text type="secondary">
              支持HTML格式和Go模板语法
            </a-typography-text>
          </div>
        </a-form-item>
      </div>

      <!-- Webhook 模板 -->
      <div v-else-if="formData.type === 'webhook'" class="form-section">
        <h3 class="section-title">
          <ApiOutlined />
          Webhook 模板
        </h3>
        
        <a-form-item label="请求体格式" name="['template', 'format']">
          <a-radio-group v-model:value="formData.template.format">
            <a-radio value="json">JSON</a-radio>
            <a-radio value="form">Form Data</a-radio>
            <a-radio value="text">纯文本</a-radio>
          </a-radio-group>
        </a-form-item>
        
        <a-form-item label="请求体模板" name="['template', 'body']">
          <div class="template-editor">
            <a-tabs v-model:activeKey="webhookActiveTab">
              <a-tab-pane key="edit" tab="编辑">
                <a-textarea
                  v-model:value="formData.template.body"
                  placeholder="请输入请求体模板"
                  :rows="15"
                  class="template-textarea"
                />
              </a-tab-pane>
              <a-tab-pane key="preview" tab="预览">
                <pre class="template-preview">{{ webhookPreview }}</pre>
              </a-tab-pane>
            </a-tabs>
          </div>
          <div class="template-help">
            <a-typography-text type="secondary">
              支持Go模板语法，JSON格式请确保语法正确
            </a-typography-text>
          </div>
        </a-form-item>
      </div>

      <!-- 钉钉模板 -->
      <div v-else-if="formData.type === 'dingtalk'" class="form-section">
        <h3 class="section-title">
          <MessageOutlined />
          钉钉模板
        </h3>
        
        <a-form-item label="消息类型" name="['template', 'msgType']">
          <a-radio-group v-model:value="formData.template.msgType">
            <a-radio value="text">文本消息</a-radio>
            <a-radio value="markdown">Markdown消息</a-radio>
          </a-radio-group>
        </a-form-item>
        
        <a-form-item v-if="formData.template.msgType === 'markdown'" label="标题" name="['template', 'title']">
          <a-input
            v-model:value="formData.template.title"
            placeholder="请输入消息标题"
          />
        </a-form-item>
        
        <a-form-item label="消息内容" name="['template', 'content']">
          <div class="template-editor">
            <a-tabs v-model:activeKey="dingtalkActiveTab">
              <a-tab-pane key="edit" tab="编辑">
                <a-textarea
                  v-model:value="formData.template.content"
                  placeholder="请输入消息内容模板"
                  :rows="12"
                  class="template-textarea"
                />
              </a-tab-pane>
              <a-tab-pane key="preview" tab="预览">
                <div class="template-preview" v-html="dingtalkPreview"></div>
              </a-tab-pane>
            </a-tabs>
          </div>
          <div class="template-help">
            <a-typography-text type="secondary">
              {{ formData.template.msgType === 'markdown' ? '支持Markdown格式和Go模板语法' : '支持纯文本和Go模板语法' }}
            </a-typography-text>
          </div>
        </a-form-item>
      </div>

      <!-- 企业微信模板 -->
      <div v-else-if="formData.type === 'wechat'" class="form-section">
        <h3 class="section-title">
          <WechatOutlined />
          企业微信模板
        </h3>
        
        <a-form-item label="消息类型" name="['template', 'msgType']">
          <a-radio-group v-model:value="formData.template.msgType">
            <a-radio value="text">文本消息</a-radio>
            <a-radio value="markdown">Markdown消息</a-radio>
          </a-radio-group>
        </a-form-item>
        
        <a-form-item label="消息内容" name="['template', 'content']">
          <div class="template-editor">
            <a-tabs v-model:activeKey="wechatActiveTab">
              <a-tab-pane key="edit" tab="编辑">
                <a-textarea
                  v-model:value="formData.template.content"
                  placeholder="请输入消息内容模板"
                  :rows="12"
                  class="template-textarea"
                />
              </a-tab-pane>
              <a-tab-pane key="preview" tab="预览">
                <div class="template-preview" v-html="wechatPreview"></div>
              </a-tab-pane>
            </a-tabs>
          </div>
          <div class="template-help">
            <a-typography-text type="secondary">
              {{ formData.template.msgType === 'markdown' ? '支持Markdown格式和Go模板语法' : '支持纯文本和Go模板语法' }}
            </a-typography-text>
          </div>
        </a-form-item>
      </div>

      <!-- Slack 模板 -->
      <div v-else-if="formData.type === 'slack'" class="form-section">
        <h3 class="section-title">
          <SlackOutlined />
          Slack 模板
        </h3>
        
        <a-form-item label="消息格式" name="['template', 'format']">
          <a-radio-group v-model:value="formData.template.format">
            <a-radio value="text">纯文本</a-radio>
            <a-radio value="blocks">Block Kit</a-radio>
          </a-radio-group>
        </a-form-item>
        
        <a-form-item v-if="formData.template.format === 'text'" label="消息内容" name="['template', 'text']">
          <a-textarea
            v-model:value="formData.template.text"
            placeholder="请输入消息内容模板"
            :rows="8"
          />
        </a-form-item>
        
        <a-form-item v-else label="Block Kit JSON" name="['template', 'blocks']">
          <div class="template-editor">
            <a-tabs v-model:activeKey="slackActiveTab">
              <a-tab-pane key="edit" tab="编辑">
                <a-textarea
                  v-model:value="formData.template.blocks"
                  placeholder="请输入Block Kit JSON模板"
                  :rows="15"
                  class="template-textarea"
                />
              </a-tab-pane>
              <a-tab-pane key="preview" tab="预览">
                <pre class="template-preview">{{ slackPreview }}</pre>
              </a-tab-pane>
            </a-tabs>
          </div>
        </a-form-item>
        
        <div class="template-help">
          <a-typography-text type="secondary">
            支持Go模板语法，Block Kit格式请参考Slack官方文档
          </a-typography-text>
        </div>
      </div>

      <!-- 模板变量说明 -->
      <div class="form-section">
        <h3 class="section-title">
          <QuestionCircleOutlined />
          模板变量说明
        </h3>
        
        <a-collapse>
          <a-collapse-panel key="common" header="通用变量">
            <div class="variable-list">
              <div class="variable-item">
                <code>{{ '{{' }} .Status {{ '}}' }}</code>
                <span>告警状态：firing 或 resolved</span>
              </div>
              <div class="variable-item">
                <code>{{ '{{' }} .Alerts | len {{ '}}' }}</code>
                <span>告警数量</span>
              </div>
              <div class="variable-item">
                <code>{{ '{{' }} .GroupLabels.alertname {{ '}}' }}</code>
                <span>告警名称</span>
              </div>
              <div class="variable-item">
                <code>{{ '{{' }} .CommonLabels.severity {{ '}}' }}</code>
                <span>告警级别</span>
              </div>
              <div class="variable-item">
                <code>{{ '{{' }} .CommonLabels.instance {{ '}}' }}</code>
                <span>实例地址</span>
              </div>
            </div>
          </a-collapse-panel>
          
          <a-collapse-panel key="alerts" header="告警列表">
            <div class="variable-list">
              <div class="variable-item">
                <code>{{ '{{' }} range .Alerts {{ '}}' }}...{{ '{{' }} end {{ '}}' }}</code>
                <span>遍历所有告警</span>
              </div>
              <div class="variable-item">
                <code>{{ '{{' }} .Labels.alertname {{ '}}' }}</code>
                <span>单个告警的名称</span>
              </div>
              <div class="variable-item">
                <code>{{ '{{' }} .Annotations.summary {{ '}}' }}</code>
                <span>单个告警的摘要</span>
              </div>
              <div class="variable-item">
                <code>{{ '{{' }} .Annotations.description {{ '}}' }}</code>
                <span>单个告警的描述</span>
              </div>
              <div class="variable-item">
                <code>{{ '{{' }} .StartsAt {{ '}}' }}</code>
                <span>告警开始时间</span>
              </div>
            </div>
          </a-collapse-panel>
          
          <a-collapse-panel key="functions" header="模板函数">
            <div class="variable-list">
              <div class="variable-item">
                <code>{{ '{{' }} .StartsAt | date "2006-01-02 15:04:05" {{ '}}' }}</code>
                <span>格式化时间</span>
              </div>
              <div class="variable-item">
                <code>{{ '{{' }} .Labels.severity | upper {{ '}}' }}</code>
                <span>转换为大写</span>
              </div>
              <div class="variable-item">
                <code>{{ '{{' }} .Annotations.description | truncate 100 {{ '}}' }}</code>
                <span>截断文本</span>
              </div>
              <div class="variable-item">
                <code>{{ '{{' }} if eq .Status "firing" {{ '}}' }}...{{ '{{' }} end {{ '}}' }}</code>
                <span>条件判断</span>
              </div>
            </div>
          </a-collapse-panel>
        </a-collapse>
      </div>

      <!-- 表单操作 -->
      <div class="form-actions">
        <a-space>
          <a-button @click="handleCancel">取消</a-button>
          <a-button @click="handlePreview" :loading="previewLoading">
            <EyeOutlined /> 预览测试
          </a-button>
          <a-button type="primary" html-type="submit" :loading="submitLoading">
            {{ mode === 'create' ? '创建' : '更新' }}
          </a-button>
        </a-space>
      </div>
    </a-form>

    <!-- 预览模态框 -->
    <a-modal
      v-model:open="previewVisible"
      title="模板预览"
      width="800px"
      :footer="null"
    >
      <div class="preview-content">
        <div v-if="previewData" class="preview-result">
          <h4>渲染结果：</h4>
          <div v-if="formData.type === 'email'" class="email-preview">
            <div class="email-subject">
              <strong>主题：</strong>{{ previewData.subject }}
            </div>
            <div class="email-body" v-html="previewData.body"></div>
          </div>
          <div v-else class="text-preview">
            <pre>{{ previewData.content || previewData.text || previewData.body }}</pre>
          </div>
        </div>
        <div v-else class="preview-empty">
          <a-empty description="暂无预览数据" />
        </div>
      </div>
    </a-modal>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, watch, onMounted } from 'vue'
import {
  Form,
  Input,
  Select,
  Switch,
  Button,
  Space,
  Row,
  Col,
  Radio,
  Tabs,
  Typography,
  Collapse,
  Modal,
  Empty,
  message
} from 'ant-design-vue'
import {
  InfoCircleOutlined,
  MailOutlined,
  ApiOutlined,
  MessageOutlined,
  WechatOutlined,
  SlackOutlined,
  QuestionCircleOutlined,
  EyeOutlined
} from '@ant-design/icons-vue'
import { previewNotificationTemplate } from '@/services/notification'
import type { NotificationTemplate } from '@/types'

const AForm = Form
const AFormItem = Form.Item
const AInput = Input
const ATextarea = Input.TextArea
const ASelect = Select
const ASelectOption = Select.Option
const ASwitch = Switch
const AButton = Button
const ASpace = Space
const ARow = Row
const ACol = Col
const ARadio = Radio
const ARadioGroup = Radio.Group
const ATabs = Tabs
const ATabPane = Tabs.TabPane
const ATypographyText = Typography.Text
const ACollapse = Collapse
const ACollapsePanel = Collapse.Panel
const AModal = Modal
const AEmpty = Empty

interface Props {
  template?: NotificationTemplate | null
  mode: 'create' | 'edit'
}

const props = withDefaults(defineProps<Props>(), {
  template: null,
  mode: 'create'
})

const emit = defineEmits<{
  submit: [data: any]
  cancel: []
}>()

const formRef = ref()
const submitLoading = ref(false)
const previewLoading = ref(false)
const previewVisible = ref(false)
const previewData = ref(null)

// Tab状态
const emailActiveTab = ref('edit')
const webhookActiveTab = ref('edit')
const dingtalkActiveTab = ref('edit')
const wechatActiveTab = ref('edit')
const slackActiveTab = ref('edit')

// 表单数据
const formData = reactive({
  name: '',
  type: 'email',
  description: '',
  enabled: true,
  template: {
    subject: '',
    body: '',
    format: 'json',
    msgType: 'text',
    title: '',
    content: '',
    text: '',
    blocks: ''
  }
})

// 表单验证规则
const formRules = {
  name: [
    { required: true, message: '请输入模板名称', trigger: 'blur' }
  ],
  type: [
    { required: true, message: '请选择模板类型', trigger: 'change' }
  ],
  ['template.subject']: [
    { required: true, message: '请输入邮件主题模板', trigger: 'blur' }
  ],
  ['template.body']: [
    { required: true, message: '请输入模板内容', trigger: 'blur' }
  ],
  ['template.content']: [
    { required: true, message: '请输入消息内容模板', trigger: 'blur' }
  ],
  ['template.text']: [
    { required: true, message: '请输入消息内容模板', trigger: 'blur' }
  ]
}

// 计算预览内容
const emailPreview = computed(() => {
  if (!formData.template.body) return ''
  // 简单的HTML渲染预览
  return formData.template.body.replace(/\n/g, '<br>')
})

const webhookPreview = computed(() => {
  if (!formData.template.body) return ''
  try {
    if (formData.template.format === 'json') {
      return JSON.stringify(JSON.parse(formData.template.body), null, 2)
    }
    return formData.template.body
  } catch {
    return formData.template.body
  }
})

const dingtalkPreview = computed(() => {
  if (!formData.template.content) return ''
  if (formData.template.msgType === 'markdown') {
    // 简单的Markdown渲染预览
    return formData.template.content
      .replace(/\*\*(.*?)\*\*/g, '<strong>$1</strong>')
      .replace(/\*(.*?)\*/g, '<em>$1</em>')
      .replace(/\n/g, '<br>')
  }
  return formData.template.content.replace(/\n/g, '<br>')
})

const wechatPreview = computed(() => {
  if (!formData.template.content) return ''
  if (formData.template.msgType === 'markdown') {
    return formData.template.content
      .replace(/\*\*(.*?)\*\*/g, '<strong>$1</strong>')
      .replace(/\*(.*?)\*/g, '<em>$1</em>')
      .replace(/\n/g, '<br>')
  }
  return formData.template.content.replace(/\n/g, '<br>')
})

const slackPreview = computed(() => {
  if (formData.template.format === 'text') {
    return formData.template.text
  }
  if (!formData.template.blocks) return ''
  try {
    return JSON.stringify(JSON.parse(formData.template.blocks), null, 2)
  } catch {
    return formData.template.blocks
  }
})

// 预览测试
const handlePreview = async () => {
  try {
    await formRef.value.validateFields()
    previewLoading.value = true
    
    const result = await previewNotificationTemplate(formData)
    previewData.value = result
    previewVisible.value = true
  } catch (error) {
    console.error('预览失败:', error)
    message.error('预览失败')
  } finally {
    previewLoading.value = false
  }
}

// 提交表单
const handleSubmit = async () => {
  try {
    submitLoading.value = true
    emit('submit', formData)
  } catch (error) {
    console.error('表单提交失败:', error)
  } finally {
    submitLoading.value = false
  }
}

// 取消
const handleCancel = () => {
  emit('cancel')
}

// 初始化表单数据
const initFormData = () => {
  if (props.template) {
    Object.assign(formData, {
      name: props.template.name || '',
      type: props.template.type || 'email',
      description: props.template.description || '',
      enabled: props.template.enabled !== false,
      template: {
        ...formData.template,
        ...props.template.template
      }
    })
  }
}

// 监听props变化
watch(() => props.template, initFormData, { immediate: true })

// 组件挂载
onMounted(() => {
  initFormData()
})
</script>

<style scoped>
.notification-template-form {
  padding: 0;
}

.form-section {
  margin-bottom: 32px;
  padding: 24px;
  background: #fafafa;
  border-radius: 8px;
}

.form-section:last-of-type {
  margin-bottom: 24px;
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

.template-editor {
  border: 1px solid #d9d9d9;
  border-radius: 6px;
  overflow: hidden;
}

.template-textarea {
  border: none !important;
  box-shadow: none !important;
  resize: none;
}

.template-preview {
  padding: 12px;
  background: #f5f5f5;
  border: none;
  min-height: 300px;
  font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
  font-size: 12px;
  line-height: 1.5;
  white-space: pre-wrap;
  word-break: break-all;
}

.template-help {
  margin-top: 8px;
}

.variable-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.variable-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 8px;
  background: #f9f9f9;
  border-radius: 4px;
}

.variable-item code {
  background: #e6f7ff;
  color: #1890ff;
  padding: 2px 6px;
  border-radius: 3px;
  font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
  font-size: 12px;
  white-space: nowrap;
  min-width: 200px;
}

.variable-item span {
  color: #666;
  font-size: 13px;
}

.form-actions {
  display: flex;
  justify-content: flex-end;
  padding: 16px 24px;
  background: #fafafa;
  border-radius: 8px;
  margin-top: 24px;
}

.preview-content {
  max-height: 500px;
  overflow-y: auto;
}

.preview-result h4 {
  margin: 0 0 16px 0;
  color: #262626;
}

.email-preview {
  border: 1px solid #e8e8e8;
  border-radius: 6px;
  overflow: hidden;
}

.email-subject {
  padding: 12px 16px;
  background: #f5f5f5;
  border-bottom: 1px solid #e8e8e8;
  font-size: 14px;
}

.email-body {
  padding: 16px;
  background: white;
  min-height: 200px;
  line-height: 1.6;
}

.text-preview {
  background: #f5f5f5;
  padding: 16px;
  border-radius: 6px;
  font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
  font-size: 12px;
  line-height: 1.5;
  white-space: pre-wrap;
  word-break: break-all;
  max-height: 400px;
  overflow-y: auto;
}

.preview-empty {
  text-align: center;
  padding: 40px;
}

/* 响应式设计 */
@media (max-width: 768px) {
  .form-section {
    padding: 16px;
  }
  
  .template-preview {
    min-height: 200px;
    font-size: 11px;
  }
  
  .variable-item {
    flex-direction: column;
    align-items: flex-start;
    gap: 8px;
  }
  
  .variable-item code {
    min-width: auto;
    word-break: break-all;
  }
  
  .form-actions {
    padding: 16px;
  }
  
  .form-actions :deep(.ant-space) {
    width: 100%;
    justify-content: center;
  }
}
</style>