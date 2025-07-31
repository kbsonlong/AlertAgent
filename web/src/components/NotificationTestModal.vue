<template>
  <a-modal
    v-model:open="visible"
    title="通知测试"
    width="800px"
    :footer="null"
    :destroyOnClose="true"
  >
    <div class="notification-test-modal">
      <a-form
        ref="formRef"
        :model="testForm"
        :rules="formRules"
        layout="vertical"
        @finish="handleTest"
      >
        <!-- 测试配置 -->
        <div class="test-section">
          <h3 class="section-title">
            <ExperimentOutlined />
            测试配置
          </h3>
          
          <a-row :gutter="16">
            <a-col :span="12">
              <a-form-item label="插件名称">
                <a-input :value="pluginName" disabled />
              </a-form-item>
            </a-col>
            <a-col :span="12">
              <a-form-item label="测试类型" name="testType">
                <a-select v-model:value="testForm.testType">
                  <a-select-option value="config">配置测试</a-select-option>
                  <a-select-option value="message">消息发送测试</a-select-option>
                  <a-select-option value="health">健康检查</a-select-option>
                </a-select>
              </a-form-item>
            </a-col>
          </a-row>
        </div>

        <!-- 测试消息配置 -->
        <div v-if="testForm.testType === 'message'" class="test-section">
          <h3 class="section-title">
            <MessageOutlined />
            测试消息
          </h3>
          
          <a-form-item label="消息标题" name="title">
            <a-input
              v-model:value="testForm.title"
              placeholder="请输入测试消息标题"
            />
          </a-form-item>
          
          <a-form-item label="消息内容" name="content">
            <a-textarea
              v-model:value="testForm.content"
              placeholder="请输入测试消息内容"
              :rows="4"
            />
          </a-form-item>
          
          <a-row :gutter="16">
            <a-col :span="12">
              <a-form-item label="消息级别" name="severity">
                <a-select v-model:value="testForm.severity">
                  <a-select-option value="info">信息</a-select-option>
                  <a-select-option value="warning">警告</a-select-option>
                  <a-select-option value="error">错误</a-select-option>
                  <a-select-option value="critical">严重</a-select-option>
                </a-select>
              </a-form-item>
            </a-col>
            <a-col :span="12">
              <a-form-item label="测试场景" name="scenario">
                <a-select v-model:value="testForm.scenario">
                  <a-select-option value="simple">简单测试</a-select-option>
                  <a-select-option value="alert">告警模拟</a-select-option>
                  <a-select-option value="batch">批量测试</a-select-option>
                </a-select>
              </a-form-item>
            </a-col>
          </a-row>
          
          <!-- 高级选项 -->
          <a-collapse>
            <a-collapse-panel key="advanced" header="高级选项">
              <a-form-item label="自定义标签">
                <div class="labels-config">
                  <div
                    v-for="(label, index) in testForm.labels"
                    :key="index"
                    class="label-item"
                  >
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
              
              <a-form-item label="自定义注解">
                <div class="annotations-config">
                  <div
                    v-for="(annotation, index) in testForm.annotations"
                    :key="index"
                    class="annotation-item"
                  >
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
            </a-collapse-panel>
          </a-collapse>
        </div>

        <!-- 测试结果 -->
        <div v-if="testResult" class="test-section">
          <h3 class="section-title">
            <CheckCircleOutlined v-if="testResult.success" style="color: #52c41a" />
            <CloseCircleOutlined v-else style="color: #ff4d4f" />
            测试结果
          </h3>
          
          <div class="test-result">
            <a-result
              :status="testResult.success ? 'success' : 'error'"
              :title="testResult.success ? '测试成功' : '测试失败'"
              :sub-title="getTestResultSubTitle()"
            >
              <template #extra>
                <div class="result-details">
                  <a-descriptions :column="2" bordered size="small">
                    <a-descriptions-item label="响应时间">
                      {{ formatDuration(testResult.duration) }}
                    </a-descriptions-item>
                    <a-descriptions-item label="测试时间">
                      {{ formatTime(testResult.timestamp) }}
                    </a-descriptions-item>
                    <a-descriptions-item v-if="testResult.error" label="错误信息" :span="2">
                      <a-typography-text type="danger" copyable>
                        {{ testResult.error }}
                      </a-typography-text>
                    </a-descriptions-item>
                  </a-descriptions>
                </div>
              </template>
            </a-result>
          </div>
        </div>

        <!-- 操作按钮 -->
        <div class="form-actions">
          <a-space>
            <a-button @click="handleCancel">取消</a-button>
            <a-button @click="resetTest" v-if="testResult">重置</a-button>
            <a-button 
              type="primary" 
              html-type="submit" 
              :loading="testing"
              :disabled="!isFormValid"
            >
              <ExperimentOutlined /> 开始测试
            </a-button>
          </a-space>
        </div>
      </a-form>
    </div>
  </a-modal>
</template>

<script setup lang="ts">
import { ref, reactive, computed, watch } from 'vue'
import { message } from 'ant-design-vue'
import {
  Modal,
  Form,
  Input,
  Select,
  Button,
  Space,
  Row,
  Col,
  Textarea,
  Collapse,
  Result,
  Descriptions,
  Typography
} from 'ant-design-vue'
import {
  ExperimentOutlined,
  MessageOutlined,
  CheckCircleOutlined,
  CloseCircleOutlined,
  PlusOutlined,
  DeleteOutlined
} from '@ant-design/icons-vue'
import { testPluginConfig, type PluginTestResult } from '@/services/plugin'
import { formatTime } from '@/utils/datetime'

const AModal = Modal
const AForm = Form
const AFormItem = Form.Item
const AInput = Input
const ATextarea = Textarea
const ASelect = Select
const ASelectOption = Select.Option
const AButton = Button
const ASpace = Space
const ARow = Row
const ACol = Col
const ACollapse = Collapse
const ACollapsePanel = Collapse.Panel
const AResult = Result
const ADescriptions = Descriptions
const ADescriptionsItem = Descriptions.Item
const ATypographyText = Typography.Text

interface Props {
  visible: boolean
  pluginName: string
  pluginConfig: Record<string, any>
}

const props = defineProps<Props>()

const emit = defineEmits<{
  'update:visible': [value: boolean]
  'test-completed': [result: PluginTestResult]
}>()

const formRef = ref()
const testing = ref(false)
const testResult = ref<PluginTestResult | null>(null)

// 测试表单数据
const testForm = reactive({
  testType: 'config',
  title: '测试通知消息',
  content: '这是一条测试消息，用于验证通知插件配置是否正确。如果您收到此消息，说明配置已成功。',
  severity: 'info',
  scenario: 'simple',
  labels: [] as Array<{ key: string; value: string }>,
  annotations: [] as Array<{ key: string; value: string }>
})

// 表单验证规则
const formRules = {
  testType: [
    { required: true, message: '请选择测试类型', trigger: 'change' }
  ],
  title: [
    { required: true, message: '请输入消息标题', trigger: 'blur' }
  ],
  content: [
    { required: true, message: '请输入消息内容', trigger: 'blur' }
  ],
  severity: [
    { required: true, message: '请选择消息级别', trigger: 'change' }
  ],
  scenario: [
    { required: true, message: '请选择测试场景', trigger: 'change' }
  ]
}

// 表单有效性检查
const isFormValid = computed(() => {
  if (testForm.testType === 'message') {
    return testForm.title && testForm.content && testForm.severity && testForm.scenario
  }
  return true
})

// 获取测试结果副标题
const getTestResultSubTitle = () => {
  if (!testResult.value) return ''
  
  if (testResult.value.success) {
    return `插件 ${props.pluginName} 测试通过，配置正确且功能正常`
  } else {
    return `插件 ${props.pluginName} 测试失败，请检查配置或网络连接`
  }
}

// 格式化持续时间
const formatDuration = (duration: number) => {
  if (duration < 1000) {
    return `${duration.toFixed(0)}ms`
  } else if (duration < 60000) {
    return `${(duration / 1000).toFixed(2)}s`
  } else {
    return `${(duration / 60000).toFixed(2)}min`
  }
}

// 添加标签
const addLabel = () => {
  testForm.labels.push({ key: '', value: '' })
}

// 删除标签
const removeLabel = (index: number) => {
  testForm.labels.splice(index, 1)
}

// 添加注解
const addAnnotation = () => {
  testForm.annotations.push({ key: '', value: '' })
}

// 删除注解
const removeAnnotation = (index: number) => {
  testForm.annotations.splice(index, 1)
}

// 执行测试
const handleTest = async () => {
  try {
    testing.value = true
    testResult.value = null
    
    let testConfig = { ...props.pluginConfig }
    
    // 根据测试类型构建测试数据
    if (testForm.testType === 'message') {
      // 构建测试消息
      const testMessage = {
        title: testForm.title,
        content: testForm.content,
        severity: testForm.severity,
        alert_id: `test-${Date.now()}`,
        timestamp: new Date().toISOString(),
        labels: testForm.labels.reduce((acc, item) => {
          if (item.key && item.value) {
            acc[item.key] = item.value
          }
          return acc
        }, {} as Record<string, string>),
        annotations: testForm.annotations.reduce((acc, item) => {
          if (item.key && item.value) {
            acc[item.key] = item.value
          }
          return acc
        }, {} as Record<string, string>)
      }
      
      // 将测试消息添加到配置中
      testConfig = {
        ...testConfig,
        test_message: testMessage,
        test_scenario: testForm.scenario
      }
    }
    
    const result = await testPluginConfig(props.pluginName, testConfig)
    testResult.value = result
    
    if (result.success) {
      message.success('测试成功')
    } else {
      message.error(`测试失败: ${result.error}`)
    }
    
    emit('test-completed', result)
  } catch (error) {
    console.error('测试失败:', error)
    message.error('测试执行失败')
    
    // 创建错误结果
    testResult.value = {
      success: false,
      error: error instanceof Error ? error.message : '未知错误',
      duration: 0,
      timestamp: new Date().toISOString()
    }
  } finally {
    testing.value = false
  }
}

// 重置测试
const resetTest = () => {
  testResult.value = null
  testForm.testType = 'config'
  testForm.title = '测试通知消息'
  testForm.content = '这是一条测试消息，用于验证通知插件配置是否正确。如果您收到此消息，说明配置已成功。'
  testForm.severity = 'info'
  testForm.scenario = 'simple'
  testForm.labels = []
  testForm.annotations = []
}

// 取消测试
const handleCancel = () => {
  emit('update:visible', false)
  resetTest()
}

// 监听visible变化，重置表单
watch(() => props.visible, (newVisible) => {
  if (newVisible) {
    resetTest()
  }
})
</script>

<style scoped>
.notification-test-modal {
  padding: 0;
}

.test-section {
  margin-bottom: 32px;
  padding: 24px;
  background: #fafafa;
  border-radius: 8px;
}

.test-section:last-of-type {
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

.labels-config,
.annotations-config {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.label-item,
.annotation-item {
  display: flex;
  align-items: flex-start;
  gap: 8px;
}

.test-result {
  background: white;
  border-radius: 8px;
  padding: 20px;
}

.result-details {
  margin-top: 16px;
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
  .test-section {
    padding: 16px;
  }
  
  .label-item,
  .annotation-item {
    flex-direction: column;
    align-items: stretch;
  }
  
  .label-item :deep(.ant-input),
  .annotation-item :deep(.ant-input),
  .annotation-item :deep(.ant-textarea) {
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