<template>
  <div class="dynamic-form-renderer">
    <a-form layout="vertical">
      <template v-for="(property, key) in schema.properties" :key="key">
        <DynamicFormField
          :field-key="key"
          :field-schema="property"
          :value="model[key]"
          :required="isRequired(key)"
          @update:value="updateValue(key, $event)"
        />
      </template>
    </a-form>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { Form } from 'ant-design-vue'
import DynamicFormField from './DynamicFormField.vue'

const AForm = Form

interface Props {
  schema: Record<string, any>
  model: Record<string, any>
}

const props = defineProps<Props>()

const emit = defineEmits<{
  'update:model': [value: Record<string, any>]
}>()

// 检查字段是否必填
const isRequired = (fieldKey: string) => {
  return props.schema.required?.includes(fieldKey) || false
}

// 更新字段值
const updateValue = (fieldKey: string, value: any) => {
  const newModel = { ...props.model }
  newModel[fieldKey] = value
  emit('update:model', newModel)
}
</script>

<style scoped>
.dynamic-form-renderer {
  width: 100%;
}
</style>