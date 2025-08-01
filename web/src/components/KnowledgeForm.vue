<template>
  <div class="knowledge-form">
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
          <a-col :span="24">
            <a-form-item label="标题" name="title">
              <a-input
                v-model:value="formData.title"
                placeholder="请输入知识标题"
                :maxlength="100"
                show-count
              />
            </a-form-item>
          </a-col>
        </a-row>
        
        <a-row :gutter="16">
          <a-col :span="12">
            <a-form-item label="分类" name="category">
              <a-select
                v-model:value="formData.category"
                placeholder="请选择分类"
                allow-clear
              >
                <a-select-option
                  v-for="category in categories"
                  :key="category.id"
                  :value="category.id"
                >
                  {{ category.name }}
                </a-select-option>
              </a-select>
            </a-form-item>
          </a-col>
          <a-col :span="12">
            <a-form-item label="状态" name="status">
              <a-radio-group v-model:value="formData.status">
                <a-radio value="draft">草稿</a-radio>
                <a-radio value="published">发布</a-radio>
              </a-radio-group>
            </a-form-item>
          </a-col>
        </a-row>
        
        <a-row :gutter="16">
          <a-col :span="24">
            <a-form-item label="标签" name="tags">
              <a-select
                v-model:value="formData.tags"
                mode="tags"
                placeholder="请输入或选择标签"
                :options="tagOptions"
                :max-tag-count="10"
              />
            </a-form-item>
          </a-col>
        </a-row>
        
        <a-row :gutter="16">
          <a-col :span="24">
            <a-form-item label="摘要" name="summary">
              <a-textarea
                v-model:value="formData.summary"
                placeholder="请输入知识摘要（可选）"
                :rows="3"
                :maxlength="500"
                show-count
              />
            </a-form-item>
          </a-col>
        </a-row>
      </a-card>

      <!-- 内容编辑 -->
      <a-card title="内容编辑" class="form-card">
        <a-form-item name="content">
          <div class="editor-container">
            <!-- 工具栏 -->
            <div class="editor-toolbar">
              <a-space>
                <a-button size="small" @click="insertMarkdown('**', '**')">
                  <template #icon><BoldOutlined /></template>
                  粗体
                </a-button>
                <a-button size="small" @click="insertMarkdown('*', '*')">
                  <template #icon><ItalicOutlined /></template>
                  斜体
                </a-button>
                <a-button size="small" @click="insertMarkdown('`', '`')">
                  <template #icon><CodeOutlined /></template>
                  代码
                </a-button>
                <a-divider type="vertical" />
                <a-button size="small" @click="insertMarkdown('# ', '')">
                  H1
                </a-button>
                <a-button size="small" @click="insertMarkdown('## ', '')">
                  H2
                </a-button>
                <a-button size="small" @click="insertMarkdown('### ', '')">
                  H3
                </a-button>
                <a-divider type="vertical" />
                <a-button size="small" @click="insertMarkdown('- ', '')">
                  <template #icon><UnorderedListOutlined /></template>
                  列表
                </a-button>
                <a-button size="small" @click="insertMarkdown('> ', '')">
                  <template #icon><MessageOutlined /></template>
                  引用
                </a-button>
                <a-divider type="vertical" />
                <a-button size="small" @click="showPreview = !showPreview">
                  <template #icon><EyeOutlined /></template>
                  {{ showPreview ? '隐藏预览' : '显示预览' }}
                </a-button>
              </a-space>
            </div>
            
            <!-- 编辑器 -->
            <div class="editor-content" :class="{ 'split-view': showPreview }">
              <div class="editor-input">
                <a-textarea
                  ref="editorRef"
                  v-model:value="formData.content"
                  placeholder="请输入知识内容，支持 Markdown 格式"
                  :rows="20"
                  class="content-editor"
                />
              </div>
              
              <div v-if="showPreview" class="editor-preview">
                <div class="preview-content" v-html="renderMarkdown(formData.content)"></div>
              </div>
            </div>
          </div>
        </a-form-item>
      </a-card>

      <!-- 来源信息 -->
      <a-card title="来源信息" class="form-card" v-if="mode === 'edit'">
        <a-row :gutter="16">
          <a-col :span="12">
            <a-form-item label="来源类型">
              <a-input :value="getSourceText(formData.source)" disabled />
            </a-form-item>
          </a-col>
          <a-col :span="12">
            <a-form-item label="来源ID">
              <a-input :value="formData.source_id" disabled />
            </a-form-item>
          </a-col>
        </a-row>
        
        <a-row :gutter="16" v-if="formData.similarity">
          <a-col :span="12">
            <a-form-item label="相似度">
              <a-progress
                :percent="Math.round(formData.similarity * 100)"
                :stroke-color="getSimilarityColor(formData.similarity)"
              />
            </a-form-item>
          </a-col>
        </a-row>
      </a-card>

      <!-- 操作按钮 -->
      <div class="form-actions">
        <a-space>
          <a-button @click="handleCancel">取消</a-button>
          <a-button @click="handleSaveDraft" :loading="saving">
            保存草稿
          </a-button>
          <a-button type="primary" html-type="submit" :loading="saving">
            {{ mode === 'create' ? '创建' : '更新' }}
          </a-button>
        </a-space>
      </div>
    </a-form>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, watch, nextTick } from 'vue'
import {
  Card,
  Row,
  Col,
  Form,
  Input,
  Select,
  Radio,
  Button,
  Space,
  Divider,
  Progress,
  message
} from 'ant-design-vue'
import {
  BoldOutlined,
  ItalicOutlined,
  CodeOutlined,
  UnorderedListOutlined,
  MessageOutlined,
  EyeOutlined
} from '@ant-design/icons-vue'
import type { Knowledge } from '@/types'

const ACard = Card
const ARow = Row
const ACol = Col
const AForm = Form
const AFormItem = Form.Item
const AInput = Input
const ATextarea = Input.TextArea
const ASelect = Select
const ASelectOption = Select.Option
const ARadioGroup = Radio.Group
const ARadio = Radio
const AButton = Button
const ASpace = Space
const ADivider = Divider
const AProgress = Progress

interface Props {
  knowledge?: Knowledge | null
  mode: 'create' | 'edit'
  categories: any[]
  tags: string[]
}

interface Emits {
  (e: 'submit', data: any): void
  (e: 'cancel'): void
}

const props = defineProps<Props>()
const emit = defineEmits<Emits>()

// 表单引用
const formRef = ref()
const editorRef = ref()

// 响应式数据
const saving = ref(false)
const showPreview = ref(false)

// 表单数据
const formData = reactive({
  title: '',
  content: '',
  summary: '',
  category: '',
  tags: [],
  status: 'draft',
  source: 'manual',
  source_id: 0,
  similarity: undefined
})

// 表单验证规则
const rules = {
  title: [
    { required: true, message: '请输入知识标题', trigger: 'blur' },
    { min: 2, max: 100, message: '标题长度应在2-100个字符之间', trigger: 'blur' }
  ],
  content: [
    { required: true, message: '请输入知识内容', trigger: 'blur' },
    { min: 10, message: '内容至少需要10个字符', trigger: 'blur' }
  ],
  category: [
    { required: true, message: '请选择分类', trigger: 'change' }
  ]
} as any

// 获取标签列表
const getTagList = (tags: string | string[] | undefined) => {
  if (!tags) return []
  try {
    if (Array.isArray(tags)) {
      return tags
    }
    return typeof tags === 'string' ? tags.split(',').filter(Boolean) : []
  } catch {
    return []
  }
}

// 计算属性
const tagOptions = computed(() => {
  return props.tags.map(tag => ({
    label: tag,
    value: tag
  }))
})

// 监听知识数据变化
watch(
  () => props.knowledge,
  (newKnowledge) => {
    if (newKnowledge && props.mode === 'edit') {
      Object.assign(formData, {
        title: newKnowledge.title,
        content: newKnowledge.content,
        summary: newKnowledge.summary || '',
        category: newKnowledge.category,
        tags: getTagList(newKnowledge.tags),
        source: newKnowledge.source,
        source_id: newKnowledge.source_id,
        similarity: newKnowledge.similarity
      })
    }
  },
  { immediate: true }
)

// 获取来源文本
const getSourceText = (source: string) => {
  const textMap: Record<string, string> = {
    manual: '手动创建',
    alert: '告警转换',
    ai: 'AI生成',
    import: '批量导入'
  }
  return textMap[source] || source
}

// 获取相似度颜色
const getSimilarityColor = (similarity: number) => {
  if (similarity >= 0.8) return '#52c41a'
  if (similarity >= 0.6) return '#faad14'
  return '#ff4d4f'
}

// 渲染Markdown
const renderMarkdown = (content: string) => {
  if (!content) return ''
  
  return content
    .replace(/\*\*(.*?)\*\*/g, '<strong>$1</strong>')
    .replace(/\*(.*?)\*/g, '<em>$1</em>')
    .replace(/`(.*?)`/g, '<code>$1</code>')
    .replace(/\n/g, '<br>')
    .replace(/^#{3}\s(.*)$/gm, '<h3>$1</h3>')
    .replace(/^#{2}\s(.*)$/gm, '<h2>$1</h2>')
    .replace(/^#{1}\s(.*)$/gm, '<h1>$1</h1>')
    .replace(/^-\s(.*)$/gm, '<li>$1</li>')
    .replace(/^>\s(.*)$/gm, '<blockquote>$1</blockquote>')
}

// 插入Markdown语法
const insertMarkdown = async (before: string, after: string) => {
  await nextTick()
  const textarea = editorRef.value?.$el || editorRef.value
  if (!textarea) return
  
  const start = textarea.selectionStart
  const end = textarea.selectionEnd
  const selectedText = formData.content.substring(start, end)
  
  const newText = before + selectedText + after
  const newContent = 
    formData.content.substring(0, start) + 
    newText + 
    formData.content.substring(end)
  
  formData.content = newContent
  
  // 重新设置光标位置
  await nextTick()
  const newCursorPos = start + before.length + selectedText.length + after.length
  textarea.setSelectionRange(newCursorPos, newCursorPos)
  textarea.focus()
}

// 提交表单
const handleSubmit = async () => {
  try {
    await formRef.value.validate()
    
    const submitData = {
      ...formData,
      tags: formData.tags.join(',')
    }
    
    emit('submit', submitData)
  } catch (error) {
    console.error('表单验证失败:', error)
  }
}

// 保存草稿
const handleSaveDraft = async () => {
  try {
    saving.value = true
    
    const draftData = {
      ...formData,
      status: 'draft',
      tags: formData.tags.join(',')
    }
    
    emit('submit', draftData)
  } catch (error) {
    console.error('保存草稿失败:', error)
    message.error('保存草稿失败')
  } finally {
    saving.value = false
  }
}

// 取消
const handleCancel = () => {
  emit('cancel')
}
</script>

<style scoped>
.knowledge-form {
  padding: 0;
}

.form-card {
  margin-bottom: 24px;
}

.form-card:last-of-type {
  margin-bottom: 0;
}

.editor-container {
  border: 1px solid #d9d9d9;
  border-radius: 6px;
  overflow: hidden;
}

.editor-toolbar {
  padding: 12px 16px;
  background: #fafafa;
  border-bottom: 1px solid #f0f0f0;
}

.editor-content {
  display: flex;
  min-height: 400px;
}

.editor-content.split-view {
  display: grid;
  grid-template-columns: 1fr 1fr;
}

.editor-input {
  flex: 1;
}

.content-editor {
  border: none;
  border-radius: 0;
  resize: none;
  font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
  font-size: 14px;
  line-height: 1.6;
}

.content-editor:focus {
  box-shadow: none;
}

.editor-preview {
  flex: 1;
  border-left: 1px solid #f0f0f0;
  background: #fafafa;
}

.preview-content {
  padding: 16px;
  height: 100%;
  overflow-y: auto;
  font-size: 14px;
  line-height: 1.8;
}

.preview-content :deep(h1),
.preview-content :deep(h2),
.preview-content :deep(h3) {
  margin: 24px 0 16px 0;
  font-weight: 600;
}

.preview-content :deep(h1) {
  font-size: 20px;
}

.preview-content :deep(h2) {
  font-size: 18px;
}

.preview-content :deep(h3) {
  font-size: 16px;
}

.preview-content :deep(code) {
  padding: 2px 6px;
  background: #f5f5f5;
  border-radius: 3px;
  font-family: 'Courier New', monospace;
  font-size: 13px;
}

.preview-content :deep(strong) {
  font-weight: 600;
}

.preview-content :deep(em) {
  font-style: italic;
}

.preview-content :deep(li) {
  margin: 4px 0;
  padding-left: 8px;
}

.preview-content :deep(blockquote) {
  margin: 16px 0;
  padding: 12px 16px;
  background: #f0f9ff;
  border-left: 4px solid #1890ff;
  border-radius: 0 6px 6px 0;
}

.form-actions {
  margin-top: 32px;
  padding-top: 24px;
  border-top: 1px solid #f0f0f0;
  text-align: right;
}

/* 响应式设计 */
@media (max-width: 768px) {
  .editor-content.split-view {
    grid-template-columns: 1fr;
    grid-template-rows: 1fr 1fr;
  }
  
  .editor-preview {
    border-left: none;
    border-top: 1px solid #f0f0f0;
  }
  
  .editor-toolbar {
    padding: 8px 12px;
  }
  
  .editor-toolbar :deep(.ant-space) {
    flex-wrap: wrap;
  }
  
  .form-actions {
    text-align: left;
  }
  
  .form-actions :deep(.ant-space) {
    flex-wrap: wrap;
  }
}
</style>