<template>
  <div class="notification-config-validator">
    <div class="validator-header">
      <h3>
        <SafetyCertificateOutlined />
        配置验证
      </h3>
      <a-button 
        type="primary" 
        size="small" 
        @click="validateConfig" 
        :loading="validating"
      >
        <CheckOutlined /> 验证配置
      </a-button>
    </div>

    <div v-if="validationResult" class="validation-result">
      <!-- 验证成功 -->
      <div v-if="validationResult.valid" class="validation-success">
        <a-result
          status="success"
          title="配置验证通过"
          sub-title="所有配置项都符合要求，插件可以正常工作"
        >
          <template #extra>
            <a-descriptions :column="1" size="small">
              <a-descriptions-item label="验证时间">
                {{ formatTime(validationResult.timestamp) }}
              </a-descriptions-item>
              <a-descriptions-item label="验证项目">
                {{ validationResult.checkedItems }} 项
              </a-descriptions-item>
            </a-descriptions>
          </template>
        </a-result>
      </div>

      <!-- 验证失败 -->
      <div v-else class="validation-failure">
        <a-result
          status="error"
          title="配置验证失败"
          sub-title="发现配置问题，请根据以下建议进行修复"
        >
          <template #extra>
            <div class="validation-errors">
              <a-alert
                v-for="(error, index) in validationResult.errors"
                :key="index"
                :message="error.message"
                :description="error.suggestion"
                type="error"
                show-icon
                :closable="false"
                class="error-item"
              >
                <template #action>
                  <a-button 
                    v-if="error.fixable" 
                    size="small" 
                    type="primary"
                    @click="applyFix(error)"
                  >
                    自动修复
                  </a-button>
                </template>
              </a-alert>
            </div>
          </template>
        </a-result>
      </div>
    </div>

    <!-- 配置建议 -->
    <div v-if="suggestions.length > 0" class="config-suggestions">
      <h4>
        <BulbOutlined />
        配置建议
      </h4>
      <a-list
        :data-source="suggestions"
        size="small"
      >
        <template #renderItem="{ item }">
          <a-list-item>
            <a-list-item-meta>
              <template #title>
                <span>{{ item.title }}</span>
                <a-tag :color="getSuggestionColor(item.level)" size="small">
                  {{ getSuggestionLevelText(item.level) }}
                </a-tag>
              </template>
              <template #description>
                {{ item.description }}
              </template>
            </a-list-item-meta>
            <template #actions>
              <a-button 
                v-if="item.actionable" 
                size="small" 
                type="link"
                @click="applySuggestion(item)"
              >
                应用建议
              </a-button>
            </template>
          </a-list-item>
        </template>
      </a-list>
    </div>

    <!-- 配置检查清单 -->
    <div class="config-checklist">
      <h4>
        <CheckSquareOutlined />
        配置检查清单
      </h4>
      <a-checkbox-group v-model:value="checkedItems" class="checklist">
        <div v-for="item in checklistItems" :key="item.key" class="checklist-item">
          <a-checkbox :value="item.key" :disabled="item.disabled">
            {{ item.label }}
          </a-checkbox>
          <div v-if="item.description" class="checklist-description">
            {{ item.description }}
          </div>
        </div>
      </a-checkbox-group>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, watch } from 'vue'
import { message } from 'ant-design-vue'
import {
  Button,
  Result,
  Descriptions,
  Alert,
  List,
  Tag,
  Checkbox
} from 'ant-design-vue'
import {
  SafetyCertificateOutlined,
  CheckOutlined,
  BulbOutlined,
  CheckSquareOutlined
} from '@ant-design/icons-vue'
import { formatTime } from '@/utils/datetime'

const AButton = Button
const AResult = Result
const ADescriptions = Descriptions
const ADescriptionsItem = Descriptions.Item
const AAlert = Alert
const AList = List
const AListItem = List.Item
const AListItemMeta = List.Item.Meta
const ATag = Tag
const ACheckbox = Checkbox
const ACheckboxGroup = Checkbox.Group

interface ValidationError {
  field: string
  message: string
  suggestion: string
  fixable: boolean
  fix?: () => void
}

interface ValidationResult {
  valid: boolean
  errors: ValidationError[]
  timestamp: string
  checkedItems: number
}

interface ConfigSuggestion {
  title: string
  description: string
  level: 'info' | 'warning' | 'error'
  actionable: boolean
  action?: () => void
}

interface ChecklistItem {
  key: string
  label: string
  description?: string
  disabled?: boolean
}

interface Props {
  pluginName: string
  pluginSchema: Record<string, any>
  config: Record<string, any>
}

const props = defineProps<Props>()

const emit = defineEmits<{
  'config-updated': [config: Record<string, any>]
}>()

const validating = ref(false)
const validationResult = ref<ValidationResult | null>(null)
const checkedItems = ref<string[]>([])

// 配置建议
const suggestions = ref<ConfigSuggestion[]>([])

// 检查清单项目
const checklistItems = computed<ChecklistItem[]>(() => {
  const items: ChecklistItem[] = []
  
  // 基于插件类型生成检查清单
  switch (props.pluginName) {
    case 'email':
      items.push(
        { key: 'smtp_host', label: 'SMTP服务器地址已配置', description: '确保SMTP服务器地址正确' },
        { key: 'smtp_port', label: 'SMTP端口已配置', description: '常用端口：587(TLS), 465(SSL), 25(无加密)' },
        { key: 'smtp_auth', label: 'SMTP认证信息已配置', description: '用户名和密码必须正确' },
        { key: 'from_address', label: '发件人地址已配置', description: '发件人邮箱地址必须有效' },
        { key: 'to_addresses', label: '收件人地址已配置', description: '至少配置一个收件人地址' }
      )
      break
    case 'dingtalk':
      items.push(
        { key: 'webhook_url', label: 'Webhook URL已配置', description: '钉钉机器人的Webhook地址' },
        { key: 'secret', label: '安全设置已配置', description: '建议配置加签密钥提高安全性' },
        { key: 'at_settings', label: '@人设置已配置', description: '配置@所有人或指定手机号' }
      )
      break
    case 'wechat':
      items.push(
        { key: 'webhook_url', label: 'Webhook URL已配置', description: '企业微信机器人的Webhook地址' },
        { key: 'mention_settings', label: '提及设置已配置', description: '配置要提及的用户或手机号' }
      )
      break
    case 'slack':
      items.push(
        { key: 'webhook_url', label: 'Webhook URL已配置', description: 'Slack应用的Webhook地址' },
        { key: 'channel', label: '频道已配置', description: '指定消息发送的频道' },
        { key: 'bot_settings', label: '机器人设置已配置', description: '配置机器人名称和图标' }
      )
      break
    case 'webhook':
      items.push(
        { key: 'url', label: 'URL已配置', description: '目标服务的HTTP端点' },
        { key: 'method', label: 'HTTP方法已配置', description: '通常使用POST方法' },
        { key: 'headers', label: '请求头已配置', description: '配置必要的认证和内容类型头' },
        { key: 'timeout', label: '超时设置已配置', description: '设置合理的请求超时时间' }
      )
      break
  }
  
  return items
})

// 获取建议级别颜色
const getSuggestionColor = (level: string) => {
  const colorMap: Record<string, string> = {
    'info': 'blue',
    'warning': 'orange',
    'error': 'red'
  }
  return colorMap[level] || 'default'
}

// 获取建议级别文本
const getSuggestionLevelText = (level: string) => {
  const textMap: Record<string, string> = {
    'info': '建议',
    'warning': '警告',
    'error': '错误'
  }
  return textMap[level] || level
}

// 验证配置
const validateConfig = async () => {
  try {
    validating.value = true
    
    const errors: ValidationError[] = []
    const newSuggestions: ConfigSuggestion[] = []
    
    // 基于Schema验证配置
    if (props.pluginSchema.properties) {
      for (const [key, schema] of Object.entries(props.pluginSchema.properties)) {
        const fieldSchema = schema as any
        const value = props.config[key]
        
        // 检查必填字段
        if (props.pluginSchema.required?.includes(key) && !value) {
          errors.push({
            field: key,
            message: `字段 "${fieldSchema.title || key}" 是必填的`,
            suggestion: `请为 "${fieldSchema.title || key}" 字段提供有效值`,
            fixable: false
          })
        }
        
        // 检查字段类型
        if (value !== undefined && fieldSchema.type) {
          const actualType = typeof value
          if (fieldSchema.type === 'number' && actualType !== 'number') {
            errors.push({
              field: key,
              message: `字段 "${fieldSchema.title || key}" 应该是数字类型`,
              suggestion: `请将 "${fieldSchema.title || key}" 的值改为数字`,
              fixable: true,
              fix: () => {
                const numValue = Number(value)
                if (!isNaN(numValue)) {
                  emit('config-updated', { ...props.config, [key]: numValue })
                }
              }
            })
          }
        }
        
        // 检查URL格式
        if (fieldSchema.format === 'url' && value) {
          try {
            new URL(value)
          } catch {
            errors.push({
              field: key,
              message: `字段 "${fieldSchema.title || key}" 不是有效的URL`,
              suggestion: `请提供有效的URL格式，如 https://example.com`,
              fixable: false
            })
          }
        }
        
        // 检查邮箱格式
        if (fieldSchema.format === 'email' && value) {
          const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/
          if (!emailRegex.test(value)) {
            errors.push({
              field: key,
              message: `字段 "${fieldSchema.title || key}" 不是有效的邮箱地址`,
              suggestion: `请提供有效的邮箱格式，如 user@example.com`,
              fixable: false
            })
          }
        }
      }
    }
    
    // 生成特定插件的建议
    generatePluginSpecificSuggestions(newSuggestions)
    
    // 更新验证结果
    validationResult.value = {
      valid: errors.length === 0,
      errors,
      timestamp: new Date().toISOString(),
      checkedItems: checklistItems.value.length
    }
    
    suggestions.value = newSuggestions
    
    if (errors.length === 0) {
      message.success('配置验证通过')
    } else {
      message.warning(`发现 ${errors.length} 个配置问题`)
    }
  } catch (error) {
    console.error('配置验证失败:', error)
    message.error('配置验证失败')
  } finally {
    validating.value = false
  }
}

// 生成插件特定建议
const generatePluginSpecificSuggestions = (suggestions: ConfigSuggestion[]) => {
  switch (props.pluginName) {
    case 'email':
      if (!props.config.smtp?.tls) {
        suggestions.push({
          title: '建议启用TLS加密',
          description: '启用TLS可以提高邮件传输的安全性',
          level: 'warning',
          actionable: true,
          action: () => {
            const newConfig = { ...props.config }
            if (!newConfig.smtp) newConfig.smtp = {}
            newConfig.smtp.tls = true
            emit('config-updated', newConfig)
          }
        })
      }
      break
    case 'dingtalk':
      if (!props.config.secret) {
        suggestions.push({
          title: '建议配置加签密钥',
          description: '配置加签密钥可以提高机器人的安全性',
          level: 'warning',
          actionable: false
        })
      }
      break
    case 'webhook':
      if (!props.config.timeout || props.config.timeout > 30) {
        suggestions.push({
          title: '建议设置合理的超时时间',
          description: '建议将超时时间设置为30秒以内',
          level: 'info',
          actionable: true,
          action: () => {
            emit('config-updated', { ...props.config, timeout: 30 })
          }
        })
      }
      break
  }
}

// 应用修复
const applyFix = (error: ValidationError) => {
  if (error.fix) {
    error.fix()
    message.success('已应用自动修复')
    // 重新验证
    setTimeout(() => validateConfig(), 500)
  }
}

// 应用建议
const applySuggestion = (suggestion: ConfigSuggestion) => {
  if (suggestion.action) {
    suggestion.action()
    message.success('已应用建议')
    // 重新验证
    setTimeout(() => validateConfig(), 500)
  }
}

// 监听配置变化，自动更新检查清单
watch(() => props.config, () => {
  updateCheckedItems()
}, { deep: true, immediate: true })

// 更新检查清单状态
const updateCheckedItems = () => {
  const checked: string[] = []
  
  checklistItems.value.forEach(item => {
    let isChecked = false
    
    switch (item.key) {
      case 'smtp_host':
        isChecked = !!props.config.smtp?.host
        break
      case 'smtp_port':
        isChecked = !!props.config.smtp?.port
        break
      case 'smtp_auth':
        isChecked = !!(props.config.smtp?.username && props.config.smtp?.password)
        break
      case 'from_address':
        isChecked = !!props.config.from
        break
      case 'to_addresses':
        isChecked = !!(props.config.to && props.config.to.length > 0)
        break
      case 'webhook_url':
        isChecked = !!(props.config.webhook || props.config.url)
        break
      case 'secret':
        isChecked = !!props.config.secret
        break
      case 'at_settings':
        isChecked = !!(props.config.atAll || (props.config.atMobiles && props.config.atMobiles.length > 0))
        break
      case 'mention_settings':
        isChecked = !!(props.config.mentionedList?.length > 0 || props.config.mentionedMobileList?.length > 0)
        break
      case 'channel':
        isChecked = !!props.config.channel
        break
      case 'bot_settings':
        isChecked = !!(props.config.username || props.config.iconEmoji || props.config.iconUrl)
        break
      case 'url':
        isChecked = !!props.config.url
        break
      case 'method':
        isChecked = !!props.config.method
        break
      case 'headers':
        isChecked = !!(props.config.headers && Object.keys(props.config.headers).length > 0)
        break
      case 'timeout':
        isChecked = !!props.config.timeout
        break
    }
    
    if (isChecked) {
      checked.push(item.key)
    }
  })
  
  checkedItems.value = checked
}
</script>

<style scoped>
.notification-config-validator {
  padding: 0;
}

.validator-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
  padding-bottom: 12px;
  border-bottom: 1px solid #e8e8e8;
}

.validator-header h3 {
  display: flex;
  align-items: center;
  gap: 8px;
  margin: 0;
  font-size: 16px;
  font-weight: 600;
  color: #262626;
}

.validation-result {
  margin-bottom: 24px;
}

.validation-errors {
  display: flex;
  flex-direction: column;
  gap: 12px;
  margin-top: 16px;
}

.error-item {
  margin-bottom: 0;
}

.config-suggestions {
  margin-bottom: 24px;
  padding: 20px;
  background: #fafafa;
  border-radius: 8px;
}

.config-suggestions h4 {
  display: flex;
  align-items: center;
  gap: 8px;
  margin: 0 0 16px 0;
  font-size: 14px;
  font-weight: 600;
  color: #262626;
}

.config-checklist {
  padding: 20px;
  background: #fafafa;
  border-radius: 8px;
}

.config-checklist h4 {
  display: flex;
  align-items: center;
  gap: 8px;
  margin: 0 0 16px 0;
  font-size: 14px;
  font-weight: 600;
  color: #262626;
}

.checklist {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.checklist-item {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.checklist-description {
  margin-left: 24px;
  font-size: 12px;
  color: #8c8c8c;
}

/* 响应式设计 */
@media (max-width: 768px) {
  .validator-header {
    flex-direction: column;
    gap: 12px;
    align-items: stretch;
  }
  
  .config-suggestions,
  .config-checklist {
    padding: 16px;
  }
}
</style>