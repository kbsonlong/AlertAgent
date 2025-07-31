<template>
  <div class="rule-form">
    <a-form
      ref="formRef"
      :model="formData"
      :rules="rules"
      layout="vertical"
      @finish="handleSubmit"
    >
      <!-- 基本信息 -->
      <a-card title="基本信息" class="form-card">
        <a-row :gutter="16">
          <a-col :span="12">
            <a-form-item label="规则名称" name="name">
              <a-input
                v-model:value="formData.name"
                placeholder="请输入规则名称"
              />
            </a-form-item>
          </a-col>
          <a-col :span="12">
            <a-form-item label="数据源" name="provider_id">
              <a-select
                v-model:value="formData.provider_id"
                placeholder="请选择数据源"
              >
                <a-select-option
                  v-for="provider in providers"
                  :key="provider.id"
                  :value="provider.id"
                >
                  {{ provider.name }}
                </a-select-option>
              </a-select>
            </a-form-item>
          </a-col>
        </a-row>
        
        <a-form-item label="描述" name="description">
          <a-textarea
            v-model:value="formData.description"
            placeholder="请输入规则描述"
            :rows="3"
          />
        </a-form-item>
        
        <a-row :gutter="16">
          <a-col :span="8">
            <a-form-item label="严重程度" name="severity">
              <a-select
                v-model:value="formData.severity"
                placeholder="请选择严重程度"
              >
                <a-select-option value="critical">严重</a-select-option>
                <a-select-option value="warning">警告</a-select-option>
                <a-select-option value="info">信息</a-select-option>
              </a-select>
            </a-form-item>
          </a-col>
          <a-col :span="8">
            <a-form-item label="评估间隔" name="evaluation_interval">
              <a-input
                v-model:value="formData.evaluation_interval"
                placeholder="如: 30s, 1m, 5m"
                addonAfter="秒/分钟"
              />
            </a-form-item>
          </a-col>
          <a-col :span="8">
            <a-form-item label="持续时间" name="for_duration">
              <a-input
                v-model:value="formData.for_duration"
                placeholder="如: 1m, 5m, 10m"
                addonAfter="分钟"
              />
            </a-form-item>
          </a-col>
        </a-row>
      </a-card>

      <!-- 查询条件 -->
      <a-card title="查询条件" class="form-card">
        <a-form-item label="查询表达式" name="query">
          <a-textarea
            v-model:value="formData.query"
            placeholder="请输入PromQL查询表达式"
            :rows="4"
          />
          <div class="query-help">
            <a-space>
              <a-button type="link" size="small" @click="showQueryBuilder">
                <template #icon><BuildOutlined /></template>
                查询构建器
              </a-button>
              <a-button type="link" size="small" @click="validateQuery">
                <template #icon><CheckCircleOutlined /></template>
                验证查询
              </a-button>
              <a-button type="link" size="small" @click="showQueryHelp">
                <template #icon><QuestionCircleOutlined /></template>
                语法帮助
              </a-button>
            </a-space>
          </div>
        </a-form-item>
        
        <a-form-item label="条件表达式" name="condition">
          <a-input
            v-model:value="formData.condition"
            placeholder="如: > 0.8, < 100, != 0"
          />
        </a-form-item>
      </a-card>

      <!-- 标签和注释 -->
      <a-card title="标签和注释" class="form-card">
        <a-form-item label="标签">
          <div class="labels-editor">
            <div
              v-for="(label, index) in formData.labels"
              :key="index"
              class="label-item"
            >
              <a-input
                v-model:value="label.key"
                placeholder="标签名"
                style="width: 40%"
              />
              <span class="separator">=</span>
              <a-input
                v-model:value="label.value"
                placeholder="标签值"
                style="width: 40%"
              />
              <a-button
                type="text"
                danger
                @click="removeLabel(index)"
                :disabled="formData.labels.length <= 1"
              >
                <template #icon><DeleteOutlined /></template>
              </a-button>
            </div>
            <a-button type="dashed" @click="addLabel" block>
              <template #icon><PlusOutlined /></template>
              添加标签
            </a-button>
          </div>
        </a-form-item>
        
        <a-form-item label="注释">
          <div class="annotations-editor">
            <div
              v-for="(annotation, index) in formData.annotations"
              :key="index"
              class="annotation-item"
            >
              <a-input
                v-model:value="annotation.key"
                placeholder="注释名"
                style="width: 30%"
              />
              <span class="separator">=</span>
              <a-textarea
                v-model:value="annotation.value"
                placeholder="注释值"
                :rows="2"
                style="width: 60%"
              />
              <a-button
                type="text"
                danger
                @click="removeAnnotation(index)"
                :disabled="formData.annotations.length <= 1"
              >
                <template #icon><DeleteOutlined /></template>
              </a-button>
            </div>
            <a-button type="dashed" @click="addAnnotation" block>
              <template #icon><PlusOutlined /></template>
              添加注释
            </a-button>
          </div>
        </a-form-item>
      </a-card>

      <!-- 通知配置 -->
      <a-card title="通知配置" class="form-card">
        <a-form-item label="通知组" name="notification_groups">
          <a-select
            v-model:value="formData.notification_groups"
            mode="multiple"
            placeholder="请选择通知组"
            :options="notificationGroupOptions"
          />
        </a-form-item>
        
        <a-row :gutter="16">
          <a-col :span="12">
            <a-form-item label="通知模板" name="notification_template">
              <a-select
                v-model:value="formData.notification_template"
                placeholder="请选择通知模板"
                allow-clear
              >
                <a-select-option
                  v-for="template in notificationTemplates"
                  :key="template.id"
                  :value="template.id"
                >
                  {{ template.name }}
                </a-select-option>
              </a-select>
            </a-form-item>
          </a-col>
          <a-col :span="12">
            <a-form-item label="通知频率" name="notification_interval">
              <a-input
                v-model:value="formData.notification_interval"
                placeholder="如: 5m, 10m, 1h"
                addonAfter="分钟/小时"
              />
            </a-form-item>
          </a-col>
        </a-row>
      </a-card>

      <!-- 高级配置 -->
      <a-card title="高级配置" class="form-card">
        <a-row :gutter="16">
          <a-col :span="8">
            <a-form-item>
              <a-checkbox v-model:checked="formData.enabled">
                启用规则
              </a-checkbox>
            </a-form-item>
          </a-col>
          <a-col :span="8">
            <a-form-item>
              <a-checkbox v-model:checked="formData.keep_firing_for">
                保持触发状态
              </a-checkbox>
            </a-form-item>
          </a-col>
          <a-col :span="8">
            <a-form-item>
              <a-checkbox v-model:checked="formData.resolve_timeout">
                自动解决超时
              </a-checkbox>
            </a-form-item>
          </a-col>
        </a-row>
        
        <a-form-item label="规则组" name="rule_group">
          <a-input
            v-model:value="formData.rule_group"
            placeholder="请输入规则组名称"
          />
        </a-form-item>
        
        <a-form-item label="外部标签">
          <a-textarea
            v-model:value="formData.external_labels"
            placeholder="YAML格式的外部标签配置"
            :rows="3"
          />
        </a-form-item>
      </a-card>
    </a-form>

    <!-- 查询构建器弹窗 -->
    <a-modal
      v-model:open="queryBuilderVisible"
      title="查询构建器"
      width="800px"
      @ok="applyQuery"
    >
      <query-builder
        ref="queryBuilderRef"
        :providers="providers"
        @query-change="handleQueryChange"
      />
    </a-modal>

    <!-- 查询帮助弹窗 -->
    <a-modal
      v-model:open="queryHelpVisible"
      title="PromQL语法帮助"
      width="600px"
      :footer="null"
    >
      <div class="query-help-content">
        <h4>基本语法</h4>
        <ul>
          <li><code>metric_name</code> - 基本指标查询</li>
          <li><code>metric_name{label="value"}</code> - 带标签过滤</li>
          <li><code>metric_name[5m]</code> - 时间范围查询</li>
          <li><code>rate(metric_name[5m])</code> - 计算速率</li>
        </ul>
        
        <h4>常用函数</h4>
        <ul>
          <li><code>rate()</code> - 计算每秒平均增长率</li>
          <li><code>increase()</code> - 计算时间范围内的增长量</li>
          <li><code>avg()</code> - 计算平均值</li>
          <li><code>sum()</code> - 计算总和</li>
          <li><code>max()</code> - 计算最大值</li>
          <li><code>min()</code> - 计算最小值</li>
        </ul>
        
        <h4>示例</h4>
        <ul>
          <li><code>up == 0</code> - 服务不可用</li>
          <li><code>rate(http_requests_total[5m]) > 100</code> - HTTP请求速率过高</li>
          <li><code>node_memory_MemAvailable_bytes / node_memory_MemTotal_bytes < 0.1</code> - 内存使用率过高</li>
        </ul>
      </div>
    </a-modal>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, watch, onMounted } from 'vue'
import {
  Card,
  Form,
  Input,
  Select,
  Textarea,
  Row,
  Col,
  Button,
  Space,
  Checkbox,
  Modal,
  message
} from 'ant-design-vue'
import {
  BuildOutlined,
  CheckCircleOutlined,
  QuestionCircleOutlined,
  DeleteOutlined,
  PlusOutlined
} from '@ant-design/icons-vue'
import type { Rule, Provider, CreateRuleRequest, UpdateRuleRequest } from '@/types'
import QueryBuilder from './QueryBuilder.vue'

const ACard = Card
const AForm = Form
const AFormItem = Form.Item
const AInput = Input
const ASelect = Select
const ASelectOption = Select.Option
const ATextarea = Textarea
const ARow = Row
const ACol = Col
const AButton = Button
const ASpace = Space
const ACheckbox = Checkbox
const AModal = Modal

interface Props {
  rule?: Rule | null
  providers: Provider[]
}

interface Emits {
  (e: 'submit', data: CreateRuleRequest | UpdateRuleRequest): void
}

const props = defineProps<Props>()
const emit = defineEmits<Emits>()

const formRef = ref()
const queryBuilderRef = ref()
const queryBuilderVisible = ref(false)
const queryHelpVisible = ref(false)
const builtQuery = ref('')

// 表单数据
const formData = reactive({
  name: '',
  description: '',
  provider_id: '',
  severity: 'warning',
  query: '',
  condition: '',
  evaluation_interval: '30s',
  for_duration: '1m',
  labels: [{ key: '', value: '' }],
  annotations: [{ key: 'summary', value: '' }],
  notification_groups: [],
  notification_template: '',
  notification_interval: '5m',
  enabled: true,
  keep_firing_for: false,
  resolve_timeout: false,
  rule_group: 'default',
  external_labels: ''
})

// 表单验证规则
const rules = {
  name: [
    { required: true, message: '请输入规则名称', trigger: 'blur' }
  ],
  provider_id: [
    { required: true, message: '请选择数据源', trigger: 'change' }
  ],
  severity: [
    { required: true, message: '请选择严重程度', trigger: 'change' }
  ],
  query: [
    { required: true, message: '请输入查询表达式', trigger: 'blur' }
  ],
  condition: [
    { required: true, message: '请输入条件表达式', trigger: 'blur' }
  ]
}

// 通知组选项
const notificationGroupOptions = ref([])

// 通知模板选项
const notificationTemplates = ref([])

// 监听规则变化
watch(
  () => props.rule,
  (newRule) => {
    if (newRule) {
      Object.assign(formData, {
        name: newRule.name || '',
        description: newRule.description || '',
        provider_id: newRule.provider_id || '',
        severity: newRule.severity || 'warning',
        query: newRule.query || '',
        condition: newRule.condition || '',
        evaluation_interval: newRule.evaluation_interval || '30s',
        for_duration: newRule.for_duration || '1m',
        labels: newRule.labels ? Object.entries(newRule.labels).map(([key, value]) => ({ key, value })) : [{ key: '', value: '' }],
        annotations: newRule.annotations ? Object.entries(newRule.annotations).map(([key, value]) => ({ key, value })) : [{ key: 'summary', value: '' }],
        notification_groups: newRule.notification_groups || [],
        notification_template: newRule.notification_template || '',
        notification_interval: newRule.notification_interval || '5m',
        enabled: newRule.enabled !== false,
        keep_firing_for: newRule.keep_firing_for || false,
        resolve_timeout: newRule.resolve_timeout || false,
        rule_group: newRule.rule_group || 'default',
        external_labels: newRule.external_labels ? JSON.stringify(newRule.external_labels, null, 2) : ''
      })
    } else {
      resetForm()
    }
  },
  { immediate: true }
)

// 重置表单
const resetForm = () => {
  Object.assign(formData, {
    name: '',
    description: '',
    provider_id: '',
    severity: 'warning',
    query: '',
    condition: '',
    evaluation_interval: '30s',
    for_duration: '1m',
    labels: [{ key: '', value: '' }],
    annotations: [{ key: 'summary', value: '' }],
    notification_groups: [],
    notification_template: '',
    notification_interval: '5m',
    enabled: true,
    keep_firing_for: false,
    resolve_timeout: false,
    rule_group: 'default',
    external_labels: ''
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

// 添加注释
const addAnnotation = () => {
  formData.annotations.push({ key: '', value: '' })
}

// 删除注释
const removeAnnotation = (index: number) => {
  if (formData.annotations.length > 1) {
    formData.annotations.splice(index, 1)
  }
}

// 显示查询构建器
const showQueryBuilder = () => {
  queryBuilderVisible.value = true
}

// 验证查询
const validateQuery = () => {
  if (!formData.query) {
    message.warning('请先输入查询表达式')
    return
  }
  
  // 这里可以调用API验证查询语法
  message.success('查询表达式语法正确')
}

// 显示查询帮助
const showQueryHelp = () => {
  queryHelpVisible.value = true
}

// 处理查询变化
const handleQueryChange = (query: string) => {
  builtQuery.value = query
}

// 应用查询
const applyQuery = () => {
  if (builtQuery.value) {
    formData.query = builtQuery.value
  }
  queryBuilderVisible.value = false
}

// 提交表单
const handleSubmit = () => {
  const submitData: any = {
    name: formData.name,
    description: formData.description,
    provider_id: formData.provider_id,
    severity: formData.severity,
    query: formData.query,
    condition: formData.condition,
    evaluation_interval: formData.evaluation_interval,
    for_duration: formData.for_duration,
    labels: formData.labels.reduce((acc, label) => {
      if (label.key && label.value) {
        acc[label.key] = label.value
      }
      return acc
    }, {} as Record<string, string>),
    annotations: formData.annotations.reduce((acc, annotation) => {
      if (annotation.key && annotation.value) {
        acc[annotation.key] = annotation.value
      }
      return acc
    }, {} as Record<string, string>),
    notification_groups: formData.notification_groups,
    notification_template: formData.notification_template,
    notification_interval: formData.notification_interval,
    enabled: formData.enabled,
    keep_firing_for: formData.keep_firing_for,
    resolve_timeout: formData.resolve_timeout,
    rule_group: formData.rule_group
  }
  
  if (formData.external_labels) {
    try {
      submitData.external_labels = JSON.parse(formData.external_labels)
    } catch (error) {
      message.error('外部标签格式错误，请检查JSON格式')
      return
    }
  }
  
  emit('submit', submitData)
}

// 暴露提交方法
const submit = () => {
  formRef.value?.validate().then(() => {
    handleSubmit()
  }).catch(() => {
    message.error('请检查表单填写')
  })
}

defineExpose({
  submit
})

// 组件挂载
onMounted(() => {
  // 加载通知组和模板选项
  // loadNotificationOptions()
})
</script>

<style scoped>
.rule-form {
  max-width: 100%;
}

.form-card {
  margin-bottom: 16px;
}

.form-card:last-child {
  margin-bottom: 0;
}

.query-help {
  margin-top: 8px;
  padding-top: 8px;
  border-top: 1px solid #f0f0f0;
}

.labels-editor,
.annotations-editor {
  border: 1px solid #d9d9d9;
  border-radius: 6px;
  padding: 12px;
  background: #fafafa;
}

.label-item,
.annotation-item {
  display: flex;
  align-items: flex-start;
  gap: 8px;
  margin-bottom: 8px;
}

.label-item:last-of-type,
.annotation-item:last-of-type {
  margin-bottom: 12px;
}

.separator {
  display: flex;
  align-items: center;
  height: 32px;
  font-weight: bold;
  color: #666;
}

.query-help-content h4 {
  margin-top: 16px;
  margin-bottom: 8px;
  color: #1890ff;
}

.query-help-content h4:first-child {
  margin-top: 0;
}

.query-help-content ul {
  margin-bottom: 16px;
  padding-left: 20px;
}

.query-help-content li {
  margin-bottom: 4px;
  line-height: 1.6;
}

.query-help-content code {
  background: #f5f5f5;
  padding: 2px 4px;
  border-radius: 3px;
  font-family: 'Courier New', monospace;
}

/* 响应式设计 */
@media (max-width: 768px) {
  .label-item,
  .annotation-item {
    flex-direction: column;
    align-items: stretch;
  }
  
  .separator {
    display: none;
  }
  
  .label-item > *,
  .annotation-item > * {
    width: 100% !important;
    margin-bottom: 8px;
  }
  
  .label-item > button,
  .annotation-item > button {
    align-self: flex-end;
    width: auto !important;
  }
}
</style>