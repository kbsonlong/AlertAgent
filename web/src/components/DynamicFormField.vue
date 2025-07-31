<template>
  <a-form-item
    :label="fieldLabel"
    :required="required"
    :help="fieldSchema.description"
  >
    <!-- 字符串输入 -->
    <a-input
      v-if="fieldType === 'string' && !fieldSchema.enum"
      :value="value"
      :placeholder="fieldSchema.placeholder || `请输入${fieldLabel}`"
      @update:value="handleUpdate"
    />
    
    <!-- 密码输入 -->
    <a-input-password
      v-else-if="fieldType === 'string' && fieldSchema.format === 'password'"
      :value="value"
      :placeholder="fieldSchema.placeholder || `请输入${fieldLabel}`"
      @update:value="handleUpdate"
    />
    
    <!-- URL输入 -->
    <a-input
      v-else-if="fieldType === 'string' && fieldSchema.format === 'url'"
      :value="value"
      :placeholder="fieldSchema.placeholder || 'https://example.com'"
      @update:value="handleUpdate"
    />
    
    <!-- 邮箱输入 -->
    <a-input
      v-else-if="fieldType === 'string' && fieldSchema.format === 'email'"
      :value="value"
      :placeholder="fieldSchema.placeholder || 'user@example.com'"
      @update:value="handleUpdate"
    />
    
    <!-- 多行文本 -->
    <a-textarea
      v-else-if="fieldType === 'string' && fieldSchema.format === 'textarea'"
      :value="value"
      :placeholder="fieldSchema.placeholder || `请输入${fieldLabel}`"
      :rows="fieldSchema.rows || 3"
      @update:value="handleUpdate"
    />
    
    <!-- 枚举选择 -->
    <a-select
      v-else-if="fieldType === 'string' && fieldSchema.enum"
      :value="value"
      :placeholder="fieldSchema.placeholder || `请选择${fieldLabel}`"
      @update:value="handleUpdate"
    >
      <a-select-option
        v-for="option in fieldSchema.enum"
        :key="option"
        :value="option"
      >
        {{ getEnumLabel(option) }}
      </a-select-option>
    </a-select>
    
    <!-- 数字输入 -->
    <a-input-number
      v-else-if="fieldType === 'number' || fieldType === 'integer'"
      :value="value"
      :placeholder="fieldSchema.placeholder || `请输入${fieldLabel}`"
      :min="fieldSchema.minimum"
      :max="fieldSchema.maximum"
      :step="fieldType === 'integer' ? 1 : 0.1"
      style="width: 100%"
      @update:value="handleUpdate"
    />
    
    <!-- 布尔值开关 -->
    <a-switch
      v-else-if="fieldType === 'boolean'"
      :checked="value"
      :checked-children="fieldSchema.checkedChildren || '是'"
      :un-checked-children="fieldSchema.unCheckedChildren || '否'"
      @update:checked="handleUpdate"
    />
    
    <!-- 数组输入 -->
    <div v-else-if="fieldType === 'array'" class="array-field">
      <div v-if="fieldSchema.items?.type === 'string'" class="string-array">
        <a-select
          :value="value || []"
          mode="tags"
          :placeholder="fieldSchema.placeholder || `请输入${fieldLabel}`"
          style="width: 100%"
          @update:value="handleUpdate"
        >
        </a-select>
      </div>
      <div v-else class="complex-array">
        <div
          v-for="(item, index) in (value || [])"
          :key="index"
          class="array-item"
        >
          <DynamicFormField
            :field-key="`${fieldKey}[${index}]`"
            :field-schema="fieldSchema.items"
            :value="item"
            @update:value="updateArrayItem(index, $event)"
          />
          <a-button
            type="text"
            danger
            size="small"
            @click="removeArrayItem(index)"
          >
            <DeleteOutlined />
          </a-button>
        </div>
        <a-button
          type="dashed"
          @click="addArrayItem"
          style="width: 100%; margin-top: 8px"
        >
          <PlusOutlined /> 添加项目
        </a-button>
      </div>
    </div>
    
    <!-- 对象输入 -->
    <div v-else-if="fieldType === 'object'" class="object-field">
      <div v-if="fieldSchema.additionalProperties" class="key-value-pairs">
        <div
          v-for="(objValue, objKey, index) in (value || {})"
          :key="index"
          class="key-value-pair"
        >
          <a-input
            :value="objKey"
            placeholder="键名"
            style="width: 200px; margin-right: 8px"
            @update:value="updateObjectKey(objKey, $event)"
          />
          <a-input
            :value="objValue"
            placeholder="键值"
            style="flex: 1; margin-right: 8px"
            @update:value="updateObjectValue(objKey, $event)"
          />
          <a-button
            type="text"
            danger
            size="small"
            @click="removeObjectKey(objKey)"
          >
            <DeleteOutlined />
          </a-button>
        </div>
        <a-button
          type="dashed"
          @click="addObjectKey"
          style="width: 100%; margin-top: 8px"
        >
          <PlusOutlined /> 添加键值对
        </a-button>
      </div>
      <div v-else class="structured-object">
        <DynamicFormRenderer
          :schema="fieldSchema"
          :model="value || {}"
          @update:model="handleUpdate"
        />
      </div>
    </div>
    
    <!-- 未知类型 -->
    <a-input
      v-else
      :value="typeof value === 'string' ? value : JSON.stringify(value)"
      :placeholder="`请输入${fieldLabel}`"
      @update:value="handleUpdate"
    />
  </a-form-item>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import {
  FormItem,
  Input,
  InputNumber,
  Select,
  Switch,
  Button,
  Textarea
} from 'ant-design-vue'
import {
  PlusOutlined,
  DeleteOutlined
} from '@ant-design/icons-vue'

const AFormItem = FormItem
const AInput = Input
const AInputPassword = Input.Password
const AInputNumber = InputNumber
const ATextarea = Textarea
const ASelect = Select
const ASelectOption = Select.Option
const ASwitch = Switch
const AButton = Button

interface Props {
  fieldKey: string
  fieldSchema: Record<string, any>
  value: any
  required?: boolean
}

const props = defineProps<Props>()

const emit = defineEmits<{
  'update:value': [value: any]
}>()

// 字段类型
const fieldType = computed(() => {
  return props.fieldSchema.type || 'string'
})

// 字段标签
const fieldLabel = computed(() => {
  return props.fieldSchema.title || props.fieldKey
})

// 获取枚举选项标签
const getEnumLabel = (value: string) => {
  const enumLabels = props.fieldSchema.enumLabels
  if (enumLabels && enumLabels[value]) {
    return enumLabels[value]
  }
  return value
}

// 处理值更新
const handleUpdate = (newValue: any) => {
  emit('update:value', newValue)
}

// 数组操作
const updateArrayItem = (index: number, newValue: any) => {
  const newArray = [...(props.value || [])]
  newArray[index] = newValue
  emit('update:value', newArray)
}

const addArrayItem = () => {
  const newArray = [...(props.value || [])]
  const itemType = props.fieldSchema.items?.type || 'string'
  
  let defaultValue: any
  switch (itemType) {
    case 'string':
      defaultValue = ''
      break
    case 'number':
    case 'integer':
      defaultValue = 0
      break
    case 'boolean':
      defaultValue = false
      break
    case 'object':
      defaultValue = {}
      break
    case 'array':
      defaultValue = []
      break
    default:
      defaultValue = ''
  }
  
  newArray.push(defaultValue)
  emit('update:value', newArray)
}

const removeArrayItem = (index: number) => {
  const newArray = [...(props.value || [])]
  newArray.splice(index, 1)
  emit('update:value', newArray)
}

// 对象操作
const updateObjectKey = (oldKey: string, newKey: string) => {
  if (oldKey === newKey) return
  
  const newObject = { ...(props.value || {}) }
  if (newKey && !newObject.hasOwnProperty(newKey)) {
    newObject[newKey] = newObject[oldKey]
    delete newObject[oldKey]
    emit('update:value', newObject)
  }
}

const updateObjectValue = (key: string, newValue: any) => {
  const newObject = { ...(props.value || {}) }
  newObject[key] = newValue
  emit('update:value', newObject)
}

const addObjectKey = () => {
  const newObject = { ...(props.value || {}) }
  let keyIndex = 1
  let newKey = `key${keyIndex}`
  
  while (newObject.hasOwnProperty(newKey)) {
    keyIndex++
    newKey = `key${keyIndex}`
  }
  
  newObject[newKey] = ''
  emit('update:value', newObject)
}

const removeObjectKey = (key: string) => {
  const newObject = { ...(props.value || {}) }
  delete newObject[key]
  emit('update:value', newObject)
}
</script>

<style scoped>
.array-field {
  width: 100%;
}

.array-item {
  display: flex;
  align-items: flex-start;
  gap: 8px;
  margin-bottom: 12px;
  padding: 12px;
  background: #f9f9f9;
  border-radius: 6px;
}

.array-item :deep(.ant-form-item) {
  flex: 1;
  margin-bottom: 0;
}

.object-field {
  width: 100%;
}

.key-value-pair {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 12px;
  padding: 12px;
  background: #f9f9f9;
  border-radius: 6px;
}

.structured-object {
  padding: 16px;
  background: #f9f9f9;
  border-radius: 6px;
}

/* 响应式设计 */
@media (max-width: 768px) {
  .array-item,
  .key-value-pair {
    flex-direction: column;
    align-items: stretch;
  }
  
  .key-value-pair .ant-input {
    width: 100% !important;
    margin-right: 0 !important;
    margin-bottom: 8px;
  }
}
</style>