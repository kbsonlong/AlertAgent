<template>
  <div class="knowledge-detail">
    <!-- 知识基本信息 -->
    <div class="detail-header">
      <div class="header-content">
        <h2 class="knowledge-title">{{ knowledge.title }}</h2>
        <div class="knowledge-meta">
          <a-space>
            <a-tag v-if="knowledge.category" color="blue">
              {{ getCategoryName(knowledge.category) }}
            </a-tag>
            <span class="meta-item">
              <ClockCircleOutlined /> {{ formatDateTime(knowledge.updated_at) }}
            </span>
          </a-space>
        </div>
      </div>
      <div class="header-actions">
        <a-space>
          <a-button @click="handleEdit">
            <template #icon><EditOutlined /></template>
            编辑
          </a-button>
          <a-button @click="handleExport">
            <template #icon><ExportOutlined /></template>
            导出
          </a-button>
          <a-dropdown>
            <template #overlay>
              <a-menu>
                <a-menu-item @click="handleDuplicate">
                  <CopyOutlined /> 复制
                </a-menu-item>
                <a-menu-item @click="handleShare">
                  <ShareAltOutlined /> 分享
                </a-menu-item>
                <a-menu-divider />
                <a-menu-item @click="handleDelete" danger>
                  <DeleteOutlined /> 删除
                </a-menu-item>
              </a-menu>
            </template>
            <a-button>
              更多 <DownOutlined />
            </a-button>
          </a-dropdown>
        </a-space>
      </div>
    </div>

    <!-- 知识标签 -->
    <div class="knowledge-tags" v-if="knowledge.tags">
      <h4>标签</h4>
      <div class="tags-container">
        <a-tag
          v-for="tag in getTagList(knowledge.tags)"
          :key="tag"
          :color="getTagColor(tag)"
        >
          {{ tag }}
        </a-tag>
      </div>
    </div>

    <!-- 知识摘要 -->
    <div class="knowledge-summary" v-if="knowledge.summary">
      <h4>摘要</h4>
      <div class="summary-content">
        {{ knowledge.summary }}
      </div>
    </div>

    <!-- 知识内容 -->
    <div class="knowledge-content">
      <h4>内容</h4>
      <div class="content-body">
        <div v-html="renderMarkdown(knowledge.content)"></div>
      </div>
    </div>

    <!-- 来源信息 -->
    <div class="knowledge-source" v-if="knowledge.source">
      <h4>来源信息</h4>
      <a-descriptions :column="2" bordered size="small">
        <a-descriptions-item label="来源类型">
          <a-tag :color="getSourceColor(knowledge.source)">
            {{ getSourceText(knowledge.source) }}
          </a-tag>
        </a-descriptions-item>
        <a-descriptions-item label="来源ID">
          {{ knowledge.source_id }}
        </a-descriptions-item>
        <a-descriptions-item label="相似度" v-if="knowledge.similarity">
          <a-progress
            :percent="Math.round(knowledge.similarity * 100)"
            size="small"
            :stroke-color="getSimilarityColor(knowledge.similarity)"
          />
        </a-descriptions-item>
      </a-descriptions>
    </div>

    <!-- 相关知识推荐 -->
    <div class="related-knowledge" v-if="relatedKnowledge.length > 0">
      <h4>相关知识</h4>
      <div class="related-list">
        <div
          v-for="item in relatedKnowledge"
          :key="item.id"
          class="related-item"
          @click="handleViewRelated(item)"
        >
          <div class="related-title">{{ item.title }}</div>
          <div class="related-meta">
            <a-space>
              <span class="similarity">
                相似度: {{ Math.round(item.similarity * 100) }}%
              </span>
              <span class="category">{{ getCategoryName(item.category) }}</span>
            </a-space>
          </div>
        </div>
      </div>
    </div>

    <!-- 操作历史 -->
    <div class="knowledge-history">
      <h4>操作历史</h4>
      <a-timeline>
        <a-timeline-item
          v-for="(history, index) in knowledgeHistory"
          :key="index"
          :color="getHistoryColor(history.action)"
        >
          <template #dot>
            <component :is="getHistoryIcon(history.action)" />
          </template>
          <div class="history-content">
            <div class="history-action">
              {{ getHistoryText(history.action) }}
            </div>
            <div class="history-meta">
              <span class="history-user">{{ history.user }}</span>
              <span class="history-time">{{ formatDateTime(history.timestamp) }}</span>
            </div>
            <div class="history-comment" v-if="history.comment">
              {{ history.comment }}
            </div>
          </div>
        </a-timeline-item>
      </a-timeline>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import {
  Space,
  Tag,
  Button,
  Dropdown,
  Menu,
  Descriptions,
  Progress,
  Timeline,
  message
} from 'ant-design-vue'
import {
  EyeOutlined,
  ClockCircleOutlined,
  EditOutlined,
  ExportOutlined,
  CopyOutlined,
  ShareAltOutlined,
  DeleteOutlined,
  DownOutlined,
  FileTextOutlined,
  UserOutlined,
  CheckCircleOutlined
} from '@ant-design/icons-vue'
import { formatDateTime } from '@/utils/datetime'
import type { Knowledge } from '@/types'

const ASpace = Space
const ATag = Tag
const AButton = Button
const ADropdown = Dropdown
const AMenu = Menu
const AMenuItem = Menu.Item
const AMenuDivider = Menu.Divider
const ADescriptions = Descriptions
const ADescriptionsItem = Descriptions.Item
const AProgress = Progress
const ATimeline = Timeline
const ATimelineItem = Timeline.Item

interface Props {
  knowledge: Knowledge
}

interface Emits {
  (e: 'edit', knowledge: Knowledge): void
  (e: 'delete', knowledge: Knowledge): void
  (e: 'close'): void
}

const props = defineProps<Props>()
const emit = defineEmits<Emits>()

// 相关知识
const relatedKnowledge = ref<Knowledge[]>([])

// 操作历史
const knowledgeHistory = ref([
  {
    action: 'created',
    user: '系统管理员',
    timestamp: props.knowledge.created_at,
    comment: '知识创建'
  },
  {
    action: 'updated',
    user: '系统管理员',
    timestamp: props.knowledge.updated_at,
    comment: '内容更新'
  }
])

// 获取标签列表
const getTagList = (tags: string | string[] | undefined) => {
  if (!tags) return []
  try {
    return typeof tags === 'string' ? tags.split(',').filter(Boolean) : tags
  } catch {
    return []
  }
}

// 获取状态颜色
const getStatusColor = (status: string) => {
  const colorMap: Record<string, string> = {
    published: 'green',
    draft: 'orange'
  }
  return colorMap[status] || 'default'
}

// 获取状态文本
const getStatusText = (status: string) => {
  const textMap: Record<string, string> = {
    published: '已发布',
    draft: '草稿'
  }
  return textMap[status] || status
}

// 获取分类名称
const getCategoryName = (categoryId: string) => {
  // 这里应该从分类列表中获取名称
  const categoryMap: Record<string, string> = {
    'alert': '告警处理',
    'troubleshooting': '故障排查',
    'monitoring': '监控配置',
    'best-practice': '最佳实践'
  }
  return categoryMap[categoryId] || '未分类'
}

// 获取标签颜色
const getTagColor = (tag: string) => {
  const colors = ['blue', 'green', 'orange', 'red', 'purple', 'cyan']
  const index = tag.length % colors.length
  return colors[index]
}

// 获取来源颜色
const getSourceColor = (source: string) => {
  const colorMap: Record<string, string> = {
    manual: 'blue',
    alert: 'orange',
    ai: 'purple',
    import: 'green'
  }
  return colorMap[source] || 'default'
}

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

// 获取历史颜色
const getHistoryColor = (action: string) => {
  const colorMap: Record<string, string> = {
    created: 'green',
    updated: 'blue',
    published: 'purple',
    deleted: 'red'
  }
  return colorMap[action] || 'gray'
}

// 获取历史图标
const getHistoryIcon = (action: string) => {
  const iconMap: Record<string, any> = {
    created: FileTextOutlined,
    updated: EditOutlined,
    published: CheckCircleOutlined,
    deleted: DeleteOutlined
  }
  return iconMap[action] || UserOutlined
}

// 获取历史文本
const getHistoryText = (action: string) => {
  const textMap: Record<string, string> = {
    created: '创建知识',
    updated: '更新内容',
    published: '发布知识',
    deleted: '删除知识'
  }
  return textMap[action] || action
}

// 渲染Markdown
const renderMarkdown = (content: string) => {
  if (!content) return ''
  
  let html = content
  
  // 处理代码块（三个反引号）
  html = html.replace(/```([\s\S]*?)```/g, '<pre><code>$1</code></pre>')
  
  // 处理标题（需要在换行处理之前）
  html = html.replace(/^#{6}\s(.*)$/gm, '<h6>$1</h6>')
  html = html.replace(/^#{5}\s(.*)$/gm, '<h5>$1</h5>')
  html = html.replace(/^#{4}\s(.*)$/gm, '<h4>$1</h4>')
  html = html.replace(/^#{3}\s(.*)$/gm, '<h3>$1</h3>')
  html = html.replace(/^#{2}\s(.*)$/gm, '<h2>$1</h2>')
  html = html.replace(/^#{1}\s(.*)$/gm, '<h1>$1</h1>')
  
  // 处理链接
  html = html.replace(/\[([^\]]+)\]\(([^)]+)\)/g, '<a href="$2" target="_blank" rel="noopener noreferrer">$1</a>')
  
  // 处理图片
  html = html.replace(/!\[([^\]]*)\]\(([^)]+)\)/g, '<img src="$2" alt="$1" style="max-width: 100%; height: auto;" />')
  
  // 处理粗体和斜体
  html = html.replace(/\*\*\*(.*?)\*\*\*/g, '<strong><em>$1</em></strong>')
  html = html.replace(/\*\*(.*?)\*\*/g, '<strong>$1</strong>')
  html = html.replace(/\*(.*?)\*/g, '<em>$1</em>')
  
  // 处理删除线
  html = html.replace(/~~(.*?)~~/g, '<del>$1</del>')
  
  // 处理行内代码
  html = html.replace(/`([^`]+)`/g, '<code>$1</code>')
  
  // 处理无序列表
  html = html.replace(/^[\s]*[-*+]\s(.*)$/gm, '<li>$1</li>')
  html = html.replace(/(<li>.*<\/li>)/s, '<ul>$1</ul>')
  
  // 处理有序列表
  html = html.replace(/^[\s]*\d+\.\s(.*)$/gm, '<li>$1</li>')
  // 如果有连续的<li>但不在<ul>中，则包装为<ol>
  html = html.replace(/(<li>(?:(?!<\/ul>|<ul>)[\s\S])*?<\/li>)/g, (match) => {
    if (!match.includes('<ul>')) {
      return `<ol>${match}</ol>`
    }
    return match
  })
  
  // 处理引用
  html = html.replace(/^>\s(.*)$/gm, '<blockquote>$1</blockquote>')
  
  // 处理水平分割线
  html = html.replace(/^[-*_]{3,}$/gm, '<hr>')
  
  // 处理表格
  const tableRegex = /^\|(.+)\|\s*\n\|([\s\S]*?)\|\s*\n((?:\|.*\|\s*\n?)*)/gm
  html = html.replace(tableRegex, (match, header, separator, rows) => {
    const headerCells = header.split('|').map((cell: string) => `<th>${cell.trim()}</th>`).join('')
    const rowCells = rows.trim().split('\n').map((row: string) => {
      const cells = row.split('|').slice(1, -1).map((cell: string) => `<td>${cell.trim()}</td>`).join('')
      return `<tr>${cells}</tr>`
    }).join('')
    return `<table><thead><tr>${headerCells}</tr></thead><tbody>${rowCells}</tbody></table>`
  })
  
  // 处理段落（将连续的非HTML行包装为段落）
  html = html.replace(/^(?!<[h1-6]|<ul|<ol|<li|<blockquote|<pre|<hr|<table)([^\n]+)$/gm, '<p>$1</p>')
  
  // 处理换行
  html = html.replace(/\n/g, '<br>')
  
  // 清理多余的<br>标签
  html = html.replace(/<\/p><br>/g, '</p>')
  html = html.replace(/<br><p>/g, '<p>')
  html = html.replace(/<\/h[1-6]><br>/g, (match) => match.replace('<br>', ''))
  html = html.replace(/<br><h[1-6]>/g, (match) => match.replace('<br>', ''))
  html = html.replace(/<\/blockquote><br>/g, '</blockquote>')
  html = html.replace(/<br><blockquote>/g, '<blockquote>')
  html = html.replace(/<\/ul><br>/g, '</ul>')
  html = html.replace(/<\/ol><br>/g, '</ol>')
  html = html.replace(/<br><ul>/g, '<ul>')
  html = html.replace(/<br><ol>/g, '<ol>')
  
  return html
}

// 编辑
const handleEdit = () => {
  emit('edit', props.knowledge)
}

// 导出
const handleExport = () => {
  // 创建导出内容
  const exportData = {
    title: props.knowledge.title,
    content: props.knowledge.content,
    summary: props.knowledge.summary,
    tags: props.knowledge.tags,
    category: props.knowledge.category,
    created_at: props.knowledge.created_at,
    updated_at: props.knowledge.updated_at
  }
  
  // 创建下载链接
  const blob = new Blob([JSON.stringify(exportData, null, 2)], {
    type: 'application/json'
  })
  const url = URL.createObjectURL(blob)
  const link = document.createElement('a')
  link.href = url
  link.download = `knowledge-${props.knowledge.id}.json`
  document.body.appendChild(link)
  link.click()
  document.body.removeChild(link)
  URL.revokeObjectURL(url)
  
  message.success('导出成功')
}

// 复制
const handleDuplicate = () => {
  // 这里应该触发复制操作
  message.success('复制成功')
}

// 分享
const handleShare = async () => {
  try {
    const shareUrl = `${window.location.origin}/knowledge/${props.knowledge.id}`
    await navigator.clipboard.writeText(shareUrl)
    message.success('分享链接已复制到剪贴板')
  } catch (error) {
    message.error('复制分享链接失败')
  }
}

// 删除
const handleDelete = () => {
  emit('delete', props.knowledge)
}

// 查看相关知识
const handleViewRelated = (knowledge: Knowledge) => {
  // 这里应该跳转到相关知识详情
  message.info(`查看相关知识: ${knowledge.title}`)
}

// 加载相关知识
const loadRelatedKnowledge = async () => {
  try {
    // 这里应该调用API获取相关知识
    // const response = await getRelatedKnowledge(props.knowledge.id)
    // relatedKnowledge.value = response.data
    
    // 模拟数据
    relatedKnowledge.value = [
      {
        id: 1,
        title: '相关知识示例1',
        category: 'alert',
        similarity: 0.85,
        content: '',
        source: 'manual',
        source_id: 1,
        created_at: new Date().toISOString(),
        updated_at: new Date().toISOString()
      },
      {
        id: 2,
        title: '相关知识示例2',
        category: 'troubleshooting',
        similarity: 0.72,
        content: '',
        source: 'ai',
        source_id: 2,
        created_at: new Date().toISOString(),
        updated_at: new Date().toISOString()
      }
    ]
  } catch (error) {
    console.error('加载相关知识失败:', error)
  }
}

// 组件挂载
onMounted(() => {
  loadRelatedKnowledge()
})
</script>

<style scoped>
.knowledge-detail {
  padding: 0;
}

.detail-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 24px;
  padding-bottom: 16px;
  border-bottom: 1px solid #f0f0f0;
}

.knowledge-title {
  margin: 0 0 12px 0;
  font-size: 24px;
  font-weight: 600;
  line-height: 1.4;
}

.knowledge-meta {
  display: flex;
  align-items: center;
  gap: 16px;
}

.meta-item {
  display: flex;
  align-items: center;
  gap: 4px;
  color: #666;
  font-size: 14px;
}

.knowledge-tags,
.knowledge-summary,
.knowledge-content,
.knowledge-source,
.related-knowledge,
.knowledge-history {
  margin-bottom: 32px;
}

.knowledge-tags h4,
.knowledge-summary h4,
.knowledge-content h4,
.knowledge-source h4,
.related-knowledge h4,
.knowledge-history h4 {
  margin: 0 0 16px 0;
  font-size: 16px;
  font-weight: 600;
  color: #262626;
}

.tags-container {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.summary-content {
  padding: 16px;
  background: #f9f9f9;
  border-radius: 6px;
  border-left: 4px solid #1890ff;
  font-size: 14px;
  line-height: 1.6;
  color: #666;
}

.content-body {
  padding: 20px;
  background: white;
  border: 1px solid #f0f0f0;
  border-radius: 6px;
  font-size: 14px;
  line-height: 1.8;
}

.content-body :deep(h1),
.content-body :deep(h2),
.content-body :deep(h3) {
  margin: 24px 0 16px 0;
  font-weight: 600;
}

.content-body :deep(h1) {
  font-size: 20px;
}

.content-body :deep(h2) {
  font-size: 18px;
}

.content-body :deep(h3) {
  font-size: 16px;
}

.content-body :deep(code) {
  padding: 2px 6px;
  background: #f5f5f5;
  border-radius: 3px;
  font-family: 'Courier New', monospace;
  font-size: 13px;
}

.content-body :deep(strong) {
  font-weight: 600;
}

.content-body :deep(em) {
  font-style: italic;
}

.related-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.related-item {
  padding: 16px;
  background: #fafafa;
  border: 1px solid #f0f0f0;
  border-radius: 6px;
  cursor: pointer;
  transition: all 0.2s;
}

.related-item:hover {
  background: #f0f9ff;
  border-color: #1890ff;
}

.related-title {
  font-weight: 500;
  color: #1890ff;
  margin-bottom: 8px;
}

.related-meta {
  font-size: 12px;
  color: #999;
}

.similarity {
  font-weight: 500;
}

.history-content {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.history-action {
  font-weight: 500;
  color: #262626;
}

.history-meta {
  display: flex;
  gap: 12px;
  font-size: 12px;
  color: #999;
}

.history-comment {
  font-size: 13px;
  color: #666;
  font-style: italic;
}

/* 响应式设计 */
@media (max-width: 768px) {
  .detail-header {
    flex-direction: column;
    gap: 16px;
  }
  
  .header-actions {
    width: 100%;
  }
  
  .knowledge-meta {
    flex-direction: column;
    align-items: flex-start;
    gap: 8px;
  }
  
  .knowledge-title {
    font-size: 20px;
  }
  
  .content-body {
    padding: 16px;
  }
  
  .related-item {
    padding: 12px;
  }
}
</style>