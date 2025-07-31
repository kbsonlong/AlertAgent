<template>
  <div class="notification-plugin-config-form">
    <a-form
      ref="formRef"
      :model="formData"
      :rules="formRules"
      layout="vertical"
      @finish="handleSubmit"
    >
      <!-- 基本配置 -->
      <div class="form-section">
        <h3 class="section-title">
          <SettingOutlined />
          基本配置
        </h3>
        
        <a-row :gutter="16">
          <a-col :span="12">
            <a-form-item label="插件名称">
              <a-input :value="plugin.name" disabled />
            </a-form-item>
          </a-col>
          <a-col :span="12">
            <a-form-item label="启用状态" name="enabled">
              <a-switch
                v-model:checked="formData.enabled"
                checked-children="启用"
                un-checked-children="禁用"
              />
            </a-form-item>
          </a-col>
        </a-row>
        
        <a-row :gutter="16">
          <a-col :span="12">
            <a-form-item label="优先级" name="priority">
              <a-input-number
                v-model:value="formData.priority"
                :min="0"
                :max="100"
                placeholder="数字越小优先级越高"
                style="width: 100%"
              />
            </a-form-item>
          </a-col>
          <a-col :span="12">
            <a-form-item label="插件版本">
              <a-input :value="plugin.version" disabled />
            </a-form-item>
          </a-col>
        </a-row>
      </div>

      <!-- 动态配置表单 -->
      <div class="form-section">
        <h3 class="section-title">
          <FormOutlined />
          插件配置
        </h3>
        
        <div class="dynamic-form">
          <DynamicFormRenderer
            v-if="plugin.schema"
            :schema="plugin.schema"
            :model="formData.config"
            @update:model="updateConfig"
          />
          <div v-else class="no-schema">
            <a-empty description="该插件暂无配置项" />
          </div>
        </div>
      </div>

      <!-- 配置验证 -->
      <div class="form-section">
        <NotificationConfigValidator
          :plugin-name="plugin.name"
          :plugin-schema="plugin.schema"
          :config="formData.config"
          @config-updated="updateConfig"
        />
      </div>

      <!-- 配置预览 -->
      <div class="form-section">
        <h3 class="section-title">
          <EyeOutlined />
          配置预览
        </h3>
        
        <a-tabs v-model:activeKey="previewTab">
          <a-tab-pane key="json" tab="JSON格式">
            <pre class="config-preview">{{ configPreview }}</pre>
          </a-tab-pane>
          <a-tab-pane key="schema" tab="Schema信息">
            <pre class="schema-preview">{{ schemaPreview }}</pre>
          </a-tab-pane>
        </a-tabs>
      </div>

      <!-- 表单操作 -->
      <div class="form-actions">
        <a-space>
          <a-button @click="handleCancel">取消</a-button>
          <a-button 
            @click="handleTest" 
            :loading="testLoading"
            :disabled="!formData.enabled"
          >
            <ExperimentOutlined /> 测试配置
          </a-button>
          <a-button type="primary" html-type="submit" :loading="submitLoading">
            <SaveOutlined /> 保存配置
          </a-button>
        </a-space>
      </div>
    </a-form>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, watch, onMounted } from 'vue'
import { message } from 'ant-design-vue'
import {
  Form,
  Input,
  InputNumber,
  Switch,
  Button,
  Space,
  Row,
  Col,
  Tabs,
  Empty
} from 'ant-design-vue'
import {
  SettingOutlined,
  FormOutlined,
  EyeOutlined,
  ExperimentOutlined,
  SaveOutlined
} from '@ant-design/icons-vue'
import { testPluginConfig, type PluginInfo, type PluginConfig } from '@/services/plugin'
import DynamicFormRenderer from './DynamicFormRenderer.vue'
import NotificationConfigValidator from './NotificationConfigValidator.vue'

const AForm = Form
const AFormItem = Form.Item
const AInput = Input
const AInputNumber = InputNumber
const ASwitch = Switch
const AButton = Button
const ASpace = Space
const ARow = Row
const ACol = Col
const ATabs = Tabs
const ATabPane = Tabs.TabPane
const AEmpty = Empty

interface Props {
  plugin: PluginInfo
  config?: PluginConfig | null
}

const props = withDefaults(defineProps<Props>(), {
  config: null
})

const emit = defineEmits<{
  submit: [config: PluginConfig]
  cancel: []
}>()

const formRef = ref()
const submitLoading = ref(false)
const testLoading = ref(false)
const previewTab = ref('json')

// 表单数据
const formData = reactive<PluginConfig>({
  name: '',
  enabled: false,
  config: {},
  priority: 0
})

// 表单验证规则
const formRules = {
  priority: [
    { required: true, message: '请输入优先级', trigger: 'blur' },
    { type: 'number', min: 0, max: 100, message: '优先级范围为0-100', trigger: 'blur' }
  ]
}

// 配置预览
const configPreview = computed(() => {
  return JSON.stringify(formData.config, null, 2)
})

// Schema预览
const schemaPreview = computed(() => {
  return JSON.stringify(props.plugin.schema, null, 2)
})

// 更新配置
const updateConfig = (newConfig: Record<string, any>) => {
  formData.config = { ...newConfig }
}

// 测试配置
const handleTest = async () => {
  try {
    await formRef.value.validateFields()
    testLoading.value = true
    
    const result = await testPluginConfig(props.plugin.name, formData.config)
    
    if (result.success) {
      message.success('配置测试成功')
    } else {
      message.error(`测试失败: ${result.error}`)
    }
  } catch (error) {
    console.error('测试配置失败:', error)
    message.error('测试失败')
  } finally {
    testLoading.value = false
  }
}

// 提交表单
const handleSubmit = async () => {
  try {
    submitLoading.value = true
    emit('submit', { ...formData })
  } catch (error) {
    console.error('提交配置失败:', error)
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
  if (props.config) {
    Object.assign(formData, {
      name: props.config.name,
      enabled: props.config.enabled,
      config: { ...props.config.config },
      priority: props.config.priority
    })
  } else {
    Object.assign(formData, {
      name: props.plugin.name,
      enabled: false,
      config: {},
      priority: 0
    })
  }
}

// 监听props变化
watch(() => props.config, initFormData, { immediate: true })

// 组件挂载
onMounted(() => {
  initFormData()
})
</script>

<style scoped>
.notification-plugin-config-form {
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

.dynamic-form {
  min-height: 200px;
}

.no-schema {
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: 200px;
}

.config-preview,
.schema-preview {
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
  margin: 0;
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
  
  .config-preview,
  .schema-preview {
    font-size: 11px;
    max-height: 300px;
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