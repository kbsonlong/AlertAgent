<template>
  <div class="notification-group-form">
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
                placeholder="请输入通知组名称"
              />
            </a-form-item>
          </a-col>
          <a-col :span="12">
            <a-form-item label="类型" name="type">
              <a-select
                v-model:value="formData.type"
                placeholder="请选择通知类型"
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
            placeholder="请输入通知组描述"
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

      <!-- 邮件配置 -->
      <div v-if="formData.type === 'email'" class="form-section">
        <h3 class="section-title">
          <MailOutlined />
          邮件配置
        </h3>
        
        <div class="config-group">
          <h4>SMTP 服务器</h4>
          <a-row :gutter="16">
            <a-col :span="12">
              <a-form-item label="服务器地址" name="['config', 'smtp', 'host']">
                <a-input
                  v-model:value="formData.config.smtp.host"
                  placeholder="smtp.example.com"
                />
              </a-form-item>
            </a-col>
            <a-col :span="6">
              <a-form-item label="端口" name="['config', 'smtp', 'port']">
                <a-input-number
                  v-model:value="formData.config.smtp.port"
                  placeholder="587"
                  :min="1"
                  :max="65535"
                  style="width: 100%"
                />
              </a-form-item>
            </a-col>
            <a-col :span="6">
              <a-form-item label="TLS" name="['config', 'smtp', 'tls']">
                <a-switch
                  v-model:checked="formData.config.smtp.tls"
                  checked-children="启用"
                  un-checked-children="禁用"
                />
              </a-form-item>
            </a-col>
          </a-row>
          
          <a-row :gutter="16">
            <a-col :span="12">
              <a-form-item label="用户名" name="['config', 'smtp', 'username']">
                <a-input
                  v-model:value="formData.config.smtp.username"
                  placeholder="请输入SMTP用户名"
                />
              </a-form-item>
            </a-col>
            <a-col :span="12">
              <a-form-item label="密码" name="['config', 'smtp', 'password']">
                <a-input-password
                  v-model:value="formData.config.smtp.password"
                  placeholder="请输入SMTP密码"
                />
              </a-form-item>
            </a-col>
          </a-row>
        </div>
        
        <div class="config-group">
          <h4>发送配置</h4>
          <a-form-item label="发件人" name="['config', 'from']">
            <a-input
              v-model:value="formData.config.from"
              placeholder="sender@example.com"
            />
          </a-form-item>
          
          <a-form-item label="收件人" name="['config', 'to']">
            <a-select
              v-model:value="formData.config.to"
              mode="tags"
              placeholder="请输入收件人邮箱，支持多个"
              style="width: 100%"
            >
            </a-select>
          </a-form-item>
          
          <a-form-item label="抄送" name="['config', 'cc']">
            <a-select
              v-model:value="formData.config.cc"
              mode="tags"
              placeholder="请输入抄送邮箱，支持多个"
              style="width: 100%"
            >
            </a-select>
          </a-form-item>
        </div>
      </div>

      <!-- Webhook 配置 -->
      <div v-else-if="formData.type === 'webhook'" class="form-section">
        <h3 class="section-title">
          <ApiOutlined />
          Webhook 配置
        </h3>
        
        <a-row :gutter="16">
          <a-col :span="18">
            <a-form-item label="URL" name="['config', 'url']">
              <a-input
                v-model:value="formData.config.url"
                placeholder="https://example.com/webhook"
              />
            </a-form-item>
          </a-col>
          <a-col :span="6">
            <a-form-item label="方法" name="['config', 'method']">
              <a-select v-model:value="formData.config.method">
                <a-select-option value="POST">POST</a-select-option>
                <a-select-option value="PUT">PUT</a-select-option>
                <a-select-option value="PATCH">PATCH</a-select-option>
              </a-select>
            </a-form-item>
          </a-col>
        </a-row>
        
        <a-row :gutter="16">
          <a-col :span="12">
            <a-form-item label="超时时间(秒)" name="['config', 'timeout']">
              <a-input-number
                v-model:value="formData.config.timeout"
                :min="1"
                :max="300"
                style="width: 100%"
              />
            </a-form-item>
          </a-col>
          <a-col :span="12">
            <a-form-item label="重试次数" name="['config', 'retries']">
              <a-input-number
                v-model:value="formData.config.retries"
                :min="0"
                :max="10"
                style="width: 100%"
              />
            </a-form-item>
          </a-col>
        </a-row>
        
        <a-form-item label="请求头">
          <div class="headers-config">
            <div v-for="(header, index) in formData.config.headers" :key="index" class="header-item">
              <a-input
                v-model:value="header.key"
                placeholder="Header名称"
                style="width: 200px; margin-right: 8px"
              />
              <a-input
                v-model:value="header.value"
                placeholder="Header值"
                style="flex: 1; margin-right: 8px"
              />
              <a-button
                type="text"
                danger
                @click="removeHeader(index)"
              >
                <DeleteOutlined />
              </a-button>
            </div>
            <a-button type="dashed" @click="addHeader" style="width: 100%">
              <PlusOutlined /> 添加请求头
            </a-button>
          </div>
        </a-form-item>
      </div>

      <!-- 钉钉配置 -->
      <div v-else-if="formData.type === 'dingtalk'" class="form-section">
        <h3 class="section-title">
          <MessageOutlined />
          钉钉配置
        </h3>
        
        <a-form-item label="Webhook URL" name="['config', 'webhook']">
          <a-input
            v-model:value="formData.config.webhook"
            placeholder="https://oapi.dingtalk.com/robot/send?access_token=xxx"
          />
        </a-form-item>
        
        <a-form-item label="密钥" name="['config', 'secret']">
          <a-input-password
            v-model:value="formData.config.secret"
            placeholder="请输入钉钉机器人密钥（可选）"
          />
        </a-form-item>
        
        <a-form-item label="@所有人" name="['config', 'atAll']">
          <a-switch
            v-model:checked="formData.config.atAll"
            checked-children="是"
            un-checked-children="否"
          />
        </a-form-item>
        
        <a-form-item label="@手机号" name="['config', 'atMobiles']">
          <a-select
            v-model:value="formData.config.atMobiles"
            mode="tags"
            placeholder="请输入要@的手机号，支持多个"
            style="width: 100%"
          >
          </a-select>
        </a-form-item>
      </div>

      <!-- 企业微信配置 -->
      <div v-else-if="formData.type === 'wechat'" class="form-section">
        <h3 class="section-title">
          <WechatOutlined />
          企业微信配置
        </h3>
        
        <a-form-item label="Webhook URL" name="['config', 'webhook']">
          <a-input
            v-model:value="formData.config.webhook"
            placeholder="https://qyapi.weixin.qq.com/cgi-bin/webhook/send?key=xxx"
          />
        </a-form-item>
        
        <a-form-item label="提及用户" name="['config', 'mentionedList']">
          <a-select
            v-model:value="formData.config.mentionedList"
            mode="tags"
            placeholder="请输入要提及的用户ID，支持多个"
            style="width: 100%"
          >
          </a-select>
        </a-form-item>
        
        <a-form-item label="提及手机号" name="['config', 'mentionedMobileList']">
          <a-select
            v-model:value="formData.config.mentionedMobileList"
            mode="tags"
            placeholder="请输入要提及的手机号，支持多个"
            style="width: 100%"
          >
          </a-select>
        </a-form-item>
      </div>

      <!-- Slack 配置 -->
      <div v-else-if="formData.type === 'slack'" class="form-section">
        <h3 class="section-title">
          <SlackOutlined />
          Slack 配置
        </h3>
        
        <a-form-item label="Webhook URL" name="['config', 'webhook']">
          <a-input
            v-model:value="formData.config.webhook"
            placeholder="https://hooks.slack.com/services/xxx/xxx/xxx"
          />
        </a-form-item>
        
        <a-row :gutter="16">
          <a-col :span="12">
            <a-form-item label="频道" name="['config', 'channel']">
              <a-input
                v-model:value="formData.config.channel"
                placeholder="#general"
              />
            </a-form-item>
          </a-col>
          <a-col :span="12">
            <a-form-item label="用户名" name="['config', 'username']">
              <a-input
                v-model:value="formData.config.username"
                placeholder="AlertBot"
              />
            </a-form-item>
          </a-col>
        </a-row>
        
        <a-row :gutter="16">
          <a-col :span="12">
            <a-form-item label="图标表情" name="['config', 'iconEmoji']">
              <a-input
                v-model:value="formData.config.iconEmoji"
                placeholder=":warning:"
              />
            </a-form-item>
          </a-col>
          <a-col :span="12">
            <a-form-item label="图标URL" name="['config', 'iconUrl']">
              <a-input
                v-model:value="formData.config.iconUrl"
                placeholder="https://example.com/icon.png"
              />
            </a-form-item>
          </a-col>
        </a-row>
      </div>

      <!-- 标签和注解 -->
      <div class="form-section">
        <h3 class="section-title">
          <TagsOutlined />
          标签和注解
        </h3>
        
        <a-form-item label="标签">
          <div class="labels-config">
            <div v-for="(label, index) in formData.labels" :key="index" class="label-item">
              <a-input
                v-model:value="label.key"
                placeholder="标签名"
                style="width: 200px; margin-right: 8px"
              />
              <a-input
                v-model:value="label.value"
                placeholder="标签值"
                style="flex: 1; margin-right: 8px"
              />
              <a-button
                type="text"
                danger
                @click="removeLabel(index)"
              >
                <DeleteOutlined />
              </a-button>
            </div>
            <a-button type="dashed" @click="addLabel" style="width: 100%">
              <PlusOutlined /> 添加标签
            </a-button>
          </div>
        </a-form-item>
        
        <a-form-item label="注解">
          <div class="annotations-config">
            <div v-for="(annotation, index) in formData.annotations" :key="index" class="annotation-item">
              <a-input
                v-model:value="annotation.key"
                placeholder="注解名"
                style="width: 200px; margin-right: 8px"
              />
              <a-textarea
                v-model:value="annotation.value"
                placeholder="注解值"
                :rows="2"
                style="flex: 1; margin-right: 8px"
              />
              <a-button
                type="text"
                danger
                @click="removeAnnotation(index)"
              >
                <DeleteOutlined />
              </a-button>
            </div>
            <a-button type="dashed" @click="addAnnotation" style="width: 100%">
              <PlusOutlined /> 添加注解
            </a-button>
          </div>
        </a-form-item>
      </div>

      <!-- 表单操作 -->
      <div class="form-actions">
        <a-space>
          <a-button @click="handleCancel">取消</a-button>
          <a-button @click="handleTest" :loading="testLoading">
            <ExperimentOutlined /> 测试连接
          </a-button>
          <a-button type="primary" html-type="submit" :loading="submitLoading">
            {{ mode === 'create' ? '创建' : '更新' }}
          </a-button>
        </a-space>
      </div>
    </a-form>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, watch, onMounted } from 'vue'
import {
  Form,
  Input,
  Select,
  Switch,
  Button,
  Space,
  Row,
  Col,
  InputNumber,
  message
} from 'ant-design-vue'
import {
  InfoCircleOutlined,
  MailOutlined,
  ApiOutlined,
  MessageOutlined,
  WechatOutlined,
  SlackOutlined,
  TagsOutlined,
  PlusOutlined,
  DeleteOutlined,
  ExperimentOutlined
} from '@ant-design/icons-vue'
import { testNotificationGroup } from '@/services/notification'
import type { NotificationGroup } from '@/types'

const AForm = Form
const AFormItem = Form.Item
const AInput = Input
const AInputPassword = Input.Password
const AInputNumber = InputNumber
const ATextarea = Input.TextArea
const ASelect = Select
const ASelectOption = Select.Option
const ASwitch = Switch
const AButton = Button
const ASpace = Space
const ARow = Row
const ACol = Col

interface Props {
  group?: NotificationGroup | null
  mode: 'create' | 'edit'
}

const props = withDefaults(defineProps<Props>(), {
  group: null,
  mode: 'create'
})

const emit = defineEmits<{
  submit: [data: any]
  cancel: []
}>()

const formRef = ref()
const submitLoading = ref(false)
const testLoading = ref(false)

// 表单数据
const formData = reactive({
  name: '',
  type: 'email',
  description: '',
  enabled: true,
  config: {
    smtp: {
      host: '',
      port: 587,
      username: '',
      password: '',
      tls: true
    },
    from: '',
    to: [],
    cc: [],
    url: '',
    method: 'POST',
    timeout: 30,
    retries: 3,
    headers: [],
    webhook: '',
    secret: '',
    atAll: false,
    atMobiles: [],
    mentionedList: [],
    mentionedMobileList: [],
    channel: '',
    username: '',
    iconEmoji: '',
    iconUrl: ''
  },
  labels: [],
  annotations: []
})

// 表单验证规则
const formRules = {
  name: [
    { required: true, message: '请输入通知组名称', trigger: 'blur' }
  ],
  type: [
    { required: true, message: '请选择通知类型', trigger: 'change' }
  ],
  ['config.smtp.host']: [
    { required: true, message: '请输入SMTP服务器地址', trigger: 'blur' }
  ],
  ['config.smtp.port']: [
    { required: true, message: '请输入SMTP端口', trigger: 'blur' }
  ],
  ['config.smtp.username']: [
    { required: true, message: '请输入SMTP用户名', trigger: 'blur' }
  ],
  ['config.smtp.password']: [
    { required: true, message: '请输入SMTP密码', trigger: 'blur' }
  ],
  ['config.from']: [
    { required: true, message: '请输入发件人邮箱', trigger: 'blur' },
    { type: 'email', message: '请输入有效的邮箱地址', trigger: 'blur' }
  ],
  ['config.to']: [
    { required: true, message: '请输入收件人邮箱', trigger: 'change' }
  ],
  ['config.url']: [
    { required: true, message: '请输入Webhook URL', trigger: 'blur' },
    { type: 'url', message: '请输入有效的URL', trigger: 'blur' }
  ],
  ['config.webhook']: [
    { required: true, message: '请输入Webhook URL', trigger: 'blur' },
    { type: 'url', message: '请输入有效的URL', trigger: 'blur' }
  ]
}

// 类型变化处理
const handleTypeChange = (type: string) => {
  // 重置配置
  formData.config = {
    smtp: {
      host: '',
      port: 587,
      username: '',
      password: '',
      tls: true
    },
    from: '',
    to: [],
    cc: [],
    url: '',
    method: 'POST',
    timeout: 30,
    retries: 3,
    headers: [],
    webhook: '',
    secret: '',
    atAll: false,
    atMobiles: [],
    mentionedList: [],
    mentionedMobileList: [],
    channel: '',
    username: '',
    iconEmoji: '',
    iconUrl: ''
  }
}

// 添加请求头
const addHeader = () => {
  formData.config.headers.push({ key: '', value: '' })
}

// 删除请求头
const removeHeader = (index: number) => {
  formData.config.headers.splice(index, 1)
}

// 添加标签
const addLabel = () => {
  formData.labels.push({ key: '', value: '' })
}

// 删除标签
const removeLabel = (index: number) => {
  formData.labels.splice(index, 1)
}

// 添加注解
const addAnnotation = () => {
  formData.annotations.push({ key: '', value: '' })
}

// 删除注解
const removeAnnotation = (index: number) => {
  formData.annotations.splice(index, 1)
}

// 测试连接
const handleTest = async () => {
  try {
    await formRef.value.validateFields()
    testLoading.value = true
    
    // 构建测试数据
    const testData = {
      ...formData,
      labels: formData.labels.reduce((acc, item) => {
        if (item.key && item.value) {
          acc[item.key] = item.value
        }
        return acc
      }, {}),
      annotations: formData.annotations.reduce((acc, item) => {
        if (item.key && item.value) {
          acc[item.key] = item.value
        }
        return acc
      }, {}),
      config: {
        ...formData.config,
        headers: formData.config.headers.reduce((acc, item) => {
          if (item.key && item.value) {
            acc[item.key] = item.value
          }
          return acc
        }, {})
      }
    }
    
    await testNotificationGroup(testData)
    message.success('连接测试成功')
  } catch (error) {
    console.error('连接测试失败:', error)
    message.error('连接测试失败')
  } finally {
    testLoading.value = false
  }
}

// 提交表单
const handleSubmit = async () => {
  try {
    submitLoading.value = true
    
    // 构建提交数据
    const submitData = {
      ...formData,
      labels: formData.labels.reduce((acc, item) => {
        if (item.key && item.value) {
          acc[item.key] = item.value
        }
        return acc
      }, {}),
      annotations: formData.annotations.reduce((acc, item) => {
        if (item.key && item.value) {
          acc[item.key] = item.value
        }
        return acc
      }, {}),
      config: {
        ...formData.config,
        headers: formData.config.headers.reduce((acc, item) => {
          if (item.key && item.value) {
            acc[item.key] = item.value
          }
          return acc
        }, {})
      }
    }
    
    emit('submit', submitData)
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
  if (props.group) {
    Object.assign(formData, {
      name: props.group.name || '',
      type: props.group.type || 'email',
      description: props.group.description || '',
      enabled: props.group.enabled !== false,
      config: {
        ...formData.config,
        ...props.group.config
      },
      labels: Object.entries(props.group.labels || {}).map(([key, value]) => ({ key, value })),
      annotations: Object.entries(props.group.annotations || {}).map(([key, value]) => ({ key, value }))
    })
    
    // 处理headers
    if (props.group.config?.headers) {
      formData.config.headers = Object.entries(props.group.config.headers).map(([key, value]) => ({ key, value }))
    }
  }
}

// 监听props变化
watch(() => props.group, initFormData, { immediate: true })

// 组件挂载
onMounted(() => {
  initFormData()
})
</script>

<style scoped>
.notification-group-form {
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

.config-group {
  margin-bottom: 24px;
}

.config-group:last-child {
  margin-bottom: 0;
}

.config-group h4 {
  margin: 0 0 16px 0;
  font-size: 14px;
  font-weight: 600;
  color: #595959;
  border-bottom: 1px solid #e8e8e8;
  padding-bottom: 8px;
}

.headers-config,
.labels-config,
.annotations-config {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.header-item,
.label-item,
.annotation-item {
  display: flex;
  align-items: flex-start;
  gap: 8px;
}

.form-actions {
  display: flex;
  justify-content: flex-end;
  padding: 16px 24px;
  background: #fafafa;
  border-radius: 8px;
  margin-top: 24px;
}

/* 响应式设计 */
@media (max-width: 768px) {
  .form-section {
    padding: 16px;
  }
  
  .header-item,
  .label-item,
  .annotation-item {
    flex-direction: column;
    align-items: stretch;
  }
  
  .header-item :deep(.ant-input),
  .label-item :deep(.ant-input) {
    width: 100% !important;
    margin-right: 0 !important;
    margin-bottom: 8px;
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