<template>
  <div class="provider-form">
    <a-form
      ref="formRef"
      :model="formData"
      :rules="rules"
      layout="vertical"
      @finish="handleSubmit"
    >
      <!-- 基本信息 -->
      <div class="form-section">
        <h3>基本信息</h3>
        <a-row :gutter="16">
          <a-col :span="12">
            <a-form-item label="数据源名称" name="name">
              <a-input
                v-model:value="formData.name"
                placeholder="请输入数据源名称"
              />
            </a-form-item>
          </a-col>
          <a-col :span="12">
            <a-form-item label="数据源类型" name="type">
              <a-select
                v-model:value="formData.type"
                placeholder="请选择数据源类型"
                @change="handleTypeChange"
              >
                <a-select-option value="prometheus">
                  <div class="type-option">
                    <DatabaseOutlined class="type-icon" />
                    Prometheus
                  </div>
                </a-select-option>
                <a-select-option value="grafana">
                  <div class="type-option">
                    <BarChartOutlined class="type-icon" />
                    Grafana
                  </div>
                </a-select-option>
                <a-select-option value="alertmanager">
                  <div class="type-option">
                    <WarningOutlined class="type-icon" />
                    AlertManager
                  </div>
                </a-select-option>
                <a-select-option value="elasticsearch">
                  <div class="type-option">
                    <SearchOutlined class="type-icon" />
                    Elasticsearch
                  </div>
                </a-select-option>
                <a-select-option value="influxdb">
                  <div class="type-option">
                    <DatabaseOutlined class="type-icon" />
                    InfluxDB
                  </div>
                </a-select-option>
              </a-select>
            </a-form-item>
          </a-col>
        </a-row>
        
        <a-form-item label="连接URL" name="url">
          <a-input
            v-model:value="formData.url"
            placeholder="请输入数据源连接URL"
            addon-before="https://"
          >
            <template #suffix>
              <a-button
                type="link"
                size="small"
                @click="handleTestConnection"
                :loading="testing"
              >
                测试连接
              </a-button>
            </template>
          </a-input>
        </a-form-item>
        
        <a-form-item label="描述">
          <a-textarea
            v-model:value="formData.description"
            placeholder="请输入数据源描述"
            :rows="3"
          />
        </a-form-item>
      </div>

      <!-- 认证配置 -->
      <div class="form-section">
        <h3>认证配置</h3>
        <a-form-item label="认证类型" name="authType">
          <a-radio-group v-model:value="formData.authType" @change="handleAuthTypeChange">
            <a-radio value="none">无认证</a-radio>
            <a-radio value="basic">基础认证</a-radio>
            <a-radio value="bearer">Bearer Token</a-radio>
            <a-radio value="oauth2">OAuth2</a-radio>
            <a-radio value="apikey">API Key</a-radio>
          </a-radio-group>
        </a-form-item>
        
        <!-- 基础认证 -->
        <div v-if="formData.authType === 'basic'" class="auth-config">
          <a-row :gutter="16">
            <a-col :span="12">
              <a-form-item label="用户名" name="username">
                <a-input
                  v-model:value="formData.authConfig.username"
                  placeholder="请输入用户名"
                />
              </a-form-item>
            </a-col>
            <a-col :span="12">
              <a-form-item label="密码" name="password">
                <a-input-password
                  v-model:value="formData.authConfig.password"
                  placeholder="请输入密码"
                />
              </a-form-item>
            </a-col>
          </a-row>
        </div>
        
        <!-- Bearer Token -->
        <div v-if="formData.authType === 'bearer'" class="auth-config">
          <a-form-item label="Token" name="token">
            <a-input
              v-model:value="formData.authConfig.token"
              placeholder="请输入Bearer Token"
              type="password"
            />
          </a-form-item>
        </div>
        
        <!-- OAuth2 -->
        <div v-if="formData.authType === 'oauth2'" class="auth-config">
          <a-row :gutter="16">
            <a-col :span="12">
              <a-form-item label="Client ID" name="clientId">
                <a-input
                  v-model:value="formData.authConfig.clientId"
                  placeholder="请输入Client ID"
                />
              </a-form-item>
            </a-col>
            <a-col :span="12">
              <a-form-item label="Client Secret" name="clientSecret">
                <a-input-password
                  v-model:value="formData.authConfig.clientSecret"
                  placeholder="请输入Client Secret"
                />
              </a-form-item>
            </a-col>
          </a-row>
          <a-form-item label="Token URL" name="tokenUrl">
            <a-input
              v-model:value="formData.authConfig.tokenUrl"
              placeholder="请输入Token获取URL"
            />
          </a-form-item>
        </div>
        
        <!-- API Key -->
        <div v-if="formData.authType === 'apikey'" class="auth-config">
          <a-row :gutter="16">
            <a-col :span="12">
              <a-form-item label="Key名称" name="keyName">
                <a-input
                  v-model:value="formData.authConfig.keyName"
                  placeholder="请输入Key名称"
                />
              </a-form-item>
            </a-col>
            <a-col :span="12">
              <a-form-item label="Key值" name="keyValue">
                <a-input
                  v-model:value="formData.authConfig.keyValue"
                  placeholder="请输入Key值"
                  type="password"
                />
              </a-form-item>
            </a-col>
          </a-row>
          <a-form-item label="传递方式" name="keyLocation">
            <a-radio-group v-model:value="formData.authConfig.keyLocation">
              <a-radio value="header">请求头</a-radio>
              <a-radio value="query">查询参数</a-radio>
            </a-radio-group>
          </a-form-item>
        </div>
      </div>

      <!-- 连接配置 -->
      <div class="form-section">
        <h3>连接配置</h3>
        <a-row :gutter="16">
          <a-col :span="8">
            <a-form-item label="超时时间(秒)" name="timeout">
              <a-input-number
                v-model:value="formData.timeout"
                :min="1"
                :max="300"
                placeholder="30"
                style="width: 100%"
              />
            </a-form-item>
          </a-col>
          <a-col :span="8">
            <a-form-item label="重试次数" name="retryCount">
              <a-input-number
                v-model:value="formData.retryCount"
                :min="0"
                :max="10"
                placeholder="3"
                style="width: 100%"
              />
            </a-form-item>
          </a-col>
          <a-col :span="8">
            <a-form-item label="检查间隔(秒)" name="checkInterval">
              <a-input-number
                v-model:value="formData.checkInterval"
                :min="10"
                :max="3600"
                placeholder="60"
                style="width: 100%"
              />
            </a-form-item>
          </a-col>
        </a-row>
        
        <a-row :gutter="16">
          <a-col :span="12">
            <a-form-item label="TLS验证">
              <a-switch
                v-model:checked="formData.tlsVerify"
                checked-children="启用"
                un-checked-children="禁用"
              />
            </a-form-item>
          </a-col>
          <a-col :span="12">
            <a-form-item label="启用状态">
              <a-switch
                v-model:checked="formData.enabled"
                checked-children="启用"
                un-checked-children="禁用"
              />
            </a-form-item>
          </a-col>
        </a-row>
        
        <a-form-item label="代理设置">
          <a-input
            v-model:value="formData.proxy"
            placeholder="http://proxy.example.com:8080"
          />
        </a-form-item>
      </div>

      <!-- 标签和注解 -->
      <div class="form-section">
        <h3>标签和注解</h3>
        <a-row :gutter="16">
          <a-col :span="12">
            <a-form-item label="标签">
              <div class="key-value-list">
                <div
                  v-for="(item, index) in formData.labels"
                  :key="index"
                  class="key-value-item"
                >
                  <a-input
                    v-model:value="item.key"
                    placeholder="标签名"
                    style="width: 40%"
                  />
                  <span class="separator">=</span>
                  <a-input
                    v-model:value="item.value"
                    placeholder="标签值"
                    style="width: 40%"
                  />
                  <a-button
                    type="text"
                    danger
                    @click="removeLabel(index)"
                    :disabled="formData.labels.length <= 1"
                  >
                    <DeleteOutlined />
                  </a-button>
                </div>
                <a-button
                  type="dashed"
                  @click="addLabel"
                  style="width: 100%; margin-top: 8px"
                >
                  <PlusOutlined /> 添加标签
                </a-button>
              </div>
            </a-form-item>
          </a-col>
          <a-col :span="12">
            <a-form-item label="注解">
              <div class="key-value-list">
                <div
                  v-for="(item, index) in formData.annotations"
                  :key="index"
                  class="key-value-item"
                >
                  <a-input
                    v-model:value="item.key"
                    placeholder="注解名"
                    style="width: 40%"
                  />
                  <span class="separator">=</span>
                  <a-input
                    v-model:value="item.value"
                    placeholder="注解值"
                    style="width: 40%"
                  />
                  <a-button
                    type="text"
                    danger
                    @click="removeAnnotation(index)"
                    :disabled="formData.annotations.length <= 1"
                  >
                    <DeleteOutlined />
                  </a-button>
                </div>
                <a-button
                  type="dashed"
                  @click="addAnnotation"
                  style="width: 100%; margin-top: 8px"
                >
                  <PlusOutlined /> 添加注解
                </a-button>
              </div>
            </a-form-item>
          </a-col>
        </a-row>
      </div>

      <!-- 表单操作 -->
      <div class="form-actions">
        <a-space>
          <a-button @click="handleCancel">取消</a-button>
          <a-button @click="handleReset">重置</a-button>
          <a-button
            type="primary"
            html-type="submit"
            :loading="submitting"
          >
            {{ mode === 'create' ? '创建' : '更新' }}
          </a-button>
        </a-space>
      </div>
    </a-form>

    <!-- 测试连接结果 -->
    <a-modal
      v-model:open="testVisible"
      title="连接测试结果"
      :footer="null"
      width="500"
    >
      <a-result
        :status="testResult.success ? 'success' : 'error'"
        :title="testResult.success ? '连接成功' : '连接失败'"
        :sub-title="testResult.message"
      >
        <template #extra>
          <a-space>
            <a-button @click="testVisible = false">关闭</a-button>
            <a-button type="primary" @click="handleTestConnection" v-if="!testResult.success">
              重新测试
            </a-button>
          </a-space>
        </template>
      </a-result>
    </a-modal>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, watch, onMounted } from 'vue'
import {
  Form,
  Row,
  Col,
  Input,
  Select,
  Button,
  Space,
  Radio,
  InputNumber,
  Switch,
  Modal,
  Result,
  message
} from 'ant-design-vue'
import {
  DatabaseOutlined,
  BarChartOutlined,
  WarningOutlined,
  SearchOutlined,
  DeleteOutlined,
  PlusOutlined
} from '@ant-design/icons-vue'
import { testProvider } from '@/services/provider'
import type { Provider } from '@/types'

const AForm = Form
const AFormItem = Form.Item
const ARow = Row
const ACol = Col
const AInput = Input
const AInputPassword = Input.Password
const ATextarea = Input.TextArea
const AInputNumber = InputNumber
const ASelect = Select
const ASelectOption = Select.Option
const AButton = Button
const ASpace = Space
const ARadio = Radio
const ARadioGroup = Radio.Group
const ASwitch = Switch
const AModal = Modal
const AResult = Result

// 组件属性
interface Props {
  provider?: Provider | null
  mode: 'create' | 'edit'
}

const props = withDefaults(defineProps<Props>(), {
  provider: null,
  mode: 'create'
})

// 组件事件
interface Emits {
  submit: [data: any]
  cancel: []
}

const emit = defineEmits<Emits>()

// 表单引用
const formRef = ref()

// 响应式数据
const submitting = ref(false)
const testing = ref(false)
const testVisible = ref(false)
const testResult = ref({
  success: false,
  message: ''
})

// 表单数据
const formData = reactive({
  name: '',
  type: 'prometheus',
  url: '',
  description: '',
  authType: 'none',
  authConfig: {
    username: '',
    password: '',
    token: '',
    clientId: '',
    clientSecret: '',
    tokenUrl: '',
    keyName: '',
    keyValue: '',
    keyLocation: 'header'
  },
  timeout: 30,
  retryCount: 3,
  checkInterval: 60,
  tlsVerify: true,
  enabled: true,
  proxy: '',
  labels: [{ key: '', value: '' }],
  annotations: [{ key: '', value: '' }]
})

// 表单验证规则
const rules = {
  name: [
    { required: true, message: '请输入数据源名称', trigger: 'blur' },
    { min: 2, max: 50, message: '名称长度在2-50个字符', trigger: 'blur' }
  ],
  type: [
    { required: true, message: '请选择数据源类型', trigger: 'change' }
  ],
  url: [
    { required: true, message: '请输入连接URL', trigger: 'blur' },
    { type: 'url', message: '请输入有效的URL', trigger: 'blur' }
  ],
  username: [
    { required: true, message: '请输入用户名', trigger: 'blur', when: () => formData.authType === 'basic' }
  ],
  password: [
    { required: true, message: '请输入密码', trigger: 'blur', when: () => formData.authType === 'basic' }
  ],
  token: [
    { required: true, message: '请输入Token', trigger: 'blur', when: () => formData.authType === 'bearer' }
  ],
  clientId: [
    { required: true, message: '请输入Client ID', trigger: 'blur', when: () => formData.authType === 'oauth2' }
  ],
  clientSecret: [
    { required: true, message: '请输入Client Secret', trigger: 'blur', when: () => formData.authType === 'oauth2' }
  ],
  tokenUrl: [
    { required: true, message: '请输入Token URL', trigger: 'blur', when: () => formData.authType === 'oauth2' }
  ],
  keyName: [
    { required: true, message: '请输入Key名称', trigger: 'blur', when: () => formData.authType === 'apikey' }
  ],
  keyValue: [
    { required: true, message: '请输入Key值', trigger: 'blur', when: () => formData.authType === 'apikey' }
  ]
}

// 初始化表单数据
const initFormData = () => {
  if (props.provider) {
    // 编辑模式，填充现有数据
    Object.assign(formData, {
      name: props.provider.name,
      type: props.provider.type,
      url: props.provider.url,
      description: props.provider.description || '',
      authType: props.provider.authType || 'none',
      authConfig: props.provider.authConfig || {
        username: '',
        password: '',
        token: '',
        clientId: '',
        clientSecret: '',
        tokenUrl: '',
        keyName: '',
        keyValue: '',
        keyLocation: 'header'
      },
      timeout: props.provider.timeout || 30,
      retryCount: props.provider.retryCount || 3,
      checkInterval: props.provider.checkInterval || 60,
      tlsVerify: props.provider.tlsVerify !== false,
      enabled: props.provider.enabled !== false,
      proxy: props.provider.proxy || '',
      labels: props.provider.labels ? Object.entries(props.provider.labels).map(([key, value]) => ({ key, value })) : [{ key: '', value: '' }],
      annotations: props.provider.annotations ? Object.entries(props.provider.annotations).map(([key, value]) => ({ key, value })) : [{ key: '', value: '' }]
    })
  }
}

// 处理类型变化
const handleTypeChange = (type: string) => {
  // 根据类型设置默认配置
  const defaultConfigs: Record<string, any> = {
    prometheus: {
      url: 'http://localhost:9090',
      authType: 'none'
    },
    grafana: {
      url: 'http://localhost:3000',
      authType: 'basic'
    },
    alertmanager: {
      url: 'http://localhost:9093',
      authType: 'none'
    },
    elasticsearch: {
      url: 'http://localhost:9200',
      authType: 'basic'
    },
    influxdb: {
      url: 'http://localhost:8086',
      authType: 'basic'
    }
  }
  
  const config = defaultConfigs[type]
  if (config && props.mode === 'create') {
    formData.url = config.url
    formData.authType = config.authType
  }
}

// 处理认证类型变化
const handleAuthTypeChange = () => {
  // 清空认证配置
  Object.assign(formData.authConfig, {
    username: '',
    password: '',
    token: '',
    clientId: '',
    clientSecret: '',
    tokenUrl: '',
    keyName: '',
    keyValue: '',
    keyLocation: 'header'
  })
}

// 添加标签
const addLabel = () => {
  formData.labels.push({ key: '', value: '' })
}

// 删除标签
const removeLabel = (index: number) => {
  if (formData.labels.length > 1) {
    formData.labels.splice(index, 1)
  }
}

// 添加注解
const addAnnotation = () => {
  formData.annotations.push({ key: '', value: '' })
}

// 删除注解
const removeAnnotation = (index: number) => {
  if (formData.annotations.length > 1) {
    formData.annotations.splice(index, 1)
  }
}

// 测试连接
const handleTestConnection = async () => {
  try {
    testing.value = true
    
    // 构建测试数据
    const testData = {
      type: formData.type,
      url: formData.url,
      authType: formData.authType,
      authConfig: formData.authConfig,
      timeout: formData.timeout,
      tlsVerify: formData.tlsVerify
    }
    
    const response = await testProvider(testData)
    testResult.value = {
      success: response.data.success,
      message: response.data.message
    }
    testVisible.value = true
  } catch (error) {
    console.error('测试连接失败:', error)
    testResult.value = {
      success: false,
      message: '测试连接失败，请检查配置'
    }
    testVisible.value = true
  } finally {
    testing.value = false
  }
}

// 提交表单
const handleSubmit = async () => {
  try {
    submitting.value = true
    
    // 验证表单
    await formRef.value.validate()
    
    // 构建提交数据
    const submitData = {
      ...formData,
      labels: formData.labels.reduce((acc, item) => {
        if (item.key && item.value) {
          acc[item.key] = item.value
        }
        return acc
      }, {} as Record<string, string>),
      annotations: formData.annotations.reduce((acc, item) => {
        if (item.key && item.value) {
          acc[item.key] = item.value
        }
        return acc
      }, {} as Record<string, string>)
    }
    
    emit('submit', submitData)
  } catch (error) {
    console.error('表单验证失败:', error)
  } finally {
    submitting.value = false
  }
}

// 取消
const handleCancel = () => {
  emit('cancel')
}

// 重置表单
const handleReset = () => {
  formRef.value.resetFields()
  initFormData()
}

// 监听provider变化
watch(
  () => props.provider,
  () => {
    initFormData()
  },
  { immediate: true }
)

// 组件挂载
onMounted(() => {
  initFormData()
})
</script>

<style scoped>
.provider-form {
  padding: 0;
}

.form-section {
  margin-bottom: 32px;
  padding-bottom: 24px;
  border-bottom: 1px solid #f0f0f0;
}

.form-section:last-of-type {
  border-bottom: none;
  margin-bottom: 24px;
}

.form-section h3 {
  margin: 0 0 16px 0;
  font-size: 16px;
  font-weight: 600;
  color: #262626;
}

.type-option {
  display: flex;
  align-items: center;
  gap: 8px;
}

.type-icon {
  font-size: 16px;
}

.auth-config {
  margin-top: 16px;
  padding: 16px;
  background: #fafafa;
  border-radius: 6px;
}

.key-value-list {
  border: 1px solid #d9d9d9;
  border-radius: 6px;
  padding: 12px;
  background: #fafafa;
}

.key-value-item {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 8px;
}

.key-value-item:last-child {
  margin-bottom: 0;
}

.separator {
  color: #8c8c8c;
  font-weight: 500;
}

.form-actions {
  text-align: right;
  padding-top: 24px;
  border-top: 1px solid #f0f0f0;
}

/* 响应式设计 */
@media (max-width: 768px) {
  .form-section {
    margin-bottom: 24px;
    padding-bottom: 16px;
  }
  
  .key-value-item {
    flex-direction: column;
    align-items: stretch;
    gap: 8px;
  }
  
  .key-value-item .separator {
    display: none;
  }
  
  .form-actions {
    text-align: center;
  }
  
  .form-actions :deep(.ant-space) {
    width: 100%;
    justify-content: center;
  }
}
</style>