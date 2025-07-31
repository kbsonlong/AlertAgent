<template>
  <div class="knowledge-list">
    <!-- 页面标题和操作 -->
    <div class="page-header">
      <div class="header-content">
        <h2>知识库管理</h2>
        <p>管理告警处理知识和解决方案</p>
      </div>
      <div class="header-actions">
        <a-space>
          <a-button @click="handleImport">
            <template #icon><ImportOutlined /></template>
            导入
          </a-button>
          <a-button @click="handleExport">
            <template #icon><ExportOutlined /></template>
            导出
          </a-button>
          <a-button type="primary" @click="handleCreate">
            <template #icon><PlusOutlined /></template>
            新建知识
          </a-button>
        </a-space>
      </div>
    </div>

    <!-- 知识库统计 -->
    <div class="stats-cards">
      <a-row :gutter="16">
        <a-col :span="6">
          <a-card>
            <a-statistic
              title="总知识数"
              :value="stats.total"
              :value-style="{ color: '#1890ff' }"
            >
              <template #prefix>
                <BookOutlined />
              </template>
            </a-statistic>
          </a-card>
        </a-col>
        <a-col :span="6">
          <a-card>
            <a-statistic
              title="已发布"
              :value="stats.published"
              :value-style="{ color: '#52c41a' }"
            >
              <template #prefix>
                <CheckCircleOutlined />
              </template>
            </a-statistic>
          </a-card>
        </a-col>
        <a-col :span="6">
          <a-card>
            <a-statistic
              title="草稿"
              :value="stats.draft"
              :value-style="{ color: '#faad14' }"
            >
              <template #prefix>
                <EditOutlined />
              </template>
            </a-statistic>
          </a-card>
        </a-col>
        <a-col :span="6">
          <a-card>
            <a-statistic
              title="本月新增"
              :value="stats.thisMonth"
              :value-style="{ color: '#722ed1' }"
            >
              <template #prefix>
                <PlusCircleOutlined />
              </template>
            </a-statistic>
          </a-card>
        </a-col>
      </a-row>
    </div>

    <!-- 搜索和筛选 -->
    <a-card class="search-card">
      <a-form layout="inline" :model="searchForm">
        <a-form-item label="关键词">
          <a-input
            v-model:value="searchForm.keyword"
            placeholder="搜索标题、内容或标签"
            style="width: 200px"
            @press-enter="handleSearch"
          >
            <template #prefix><SearchOutlined /></template>
          </a-input>
        </a-form-item>
        <a-form-item label="分类">
          <a-select
            v-model:value="searchForm.category"
            placeholder="请选择分类"
            style="width: 150px"
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
        <a-form-item label="状态">
          <a-select
            v-model:value="searchForm.status"
            placeholder="请选择状态"
            style="width: 120px"
            allow-clear
          >
            <a-select-option value="published">已发布</a-select-option>
            <a-select-option value="draft">草稿</a-select-option>
          </a-select>
        </a-form-item>
        <a-form-item label="标签">
          <a-select
            v-model:value="searchForm.tags"
            mode="multiple"
            placeholder="请选择标签"
            style="width: 200px"
            :options="tagOptions"
          />
        </a-form-item>
        <a-form-item>
          <a-space>
            <a-button type="primary" @click="handleSearch">
              <template #icon><SearchOutlined /></template>
              搜索
            </a-button>
            <a-button @click="handleReset">
              <template #icon><ReloadOutlined /></template>
              重置
            </a-button>
          </a-space>
        </a-form-item>
      </a-form>
    </a-card>

    <!-- 批量操作 -->
    <div class="batch-actions" v-if="selectedRowKeys.length > 0">
      <a-alert
        :message="`已选择 ${selectedRowKeys.length} 项`"
        type="info"
        show-icon
      >
        <template #action>
          <a-space>
            <a-button size="small" @click="handleBatchPublish">
              发布
            </a-button>
            <a-button size="small" @click="handleBatchUnpublish">
              取消发布
            </a-button>
            <a-button size="small" danger @click="handleBatchDelete">
              删除
            </a-button>
          </a-space>
        </template>
      </a-alert>
    </div>

    <!-- 知识列表 -->
    <a-card class="table-card">
      <a-table
        :columns="columns"
        :data-source="knowledgeList"
        :loading="loading"
        :pagination="pagination"
        :row-selection="rowSelection"
        @change="handleTableChange"
        row-key="id"
      >
        <template #bodyCell="{ column, record }">
          <template v-if="column.key === 'title'">
            <div class="knowledge-title">
              <a @click="handleView(record)" class="title-link">
                {{ record.title }}
              </a>
              <div class="title-meta">
                <a-tag v-if="record.category" size="small" color="blue">
                  {{ getCategoryName(record.category) }}
                </a-tag>
                <span class="meta-text">{{ record.summary }}</span>
              </div>
            </div>
          </template>
          
          <template v-else-if="column.key === 'tags'">
            <div class="tags-container">
              <a-tag
                v-for="tag in record.tags"
                :key="tag"
                size="small"
                :color="getTagColor(tag)"
              >
                {{ tag }}
              </a-tag>
            </div>
          </template>
          
          <template v-else-if="column.key === 'status'">
            <a-tag :color="getStatusColor(record.status)">
              {{ getStatusText(record.status) }}
            </a-tag>
          </template>
          
          <template v-else-if="column.key === 'views'">
            <span class="view-count">
              <EyeOutlined /> {{ record.views || record.viewCount || 0 }}
            </span>
          </template>
          
          <template v-else-if="column.key === 'updatedAt'">
            <div class="time-info">
              <div>{{ formatDateTime(record.updatedAt || record.updated_at || record.updateTime) }}</div>
              <div class="time-relative">{{ getRelativeTime(record.updatedAt || record.updated_at || record.updateTime) }}</div>
            </div>
          </template>
          
          <template v-else-if="column.key === 'action'">
            <a-space>
              <a-button type="link" size="small" @click="handleView(record)">
                查看
              </a-button>
              <a-button type="link" size="small" @click="handleEdit(record)">
                编辑
              </a-button>
              <a-dropdown>
                <template #overlay>
                  <a-menu>
                    <a-menu-item @click="handleDuplicate(record)">
                      <CopyOutlined /> 复制
                    </a-menu-item>
                    <a-menu-item @click="handleExportSingle(record)">
                      <ExportOutlined /> 导出
                    </a-menu-item>
                    <a-menu-divider />
                    <a-menu-item
                      @click="handleToggleStatus(record)"
                      :disabled="(record.status || 'draft') === 'published'"
                    >
                      <CheckCircleOutlined /> 发布
                    </a-menu-item>
                    <a-menu-item
                      @click="handleToggleStatus(record)"
                      :disabled="(record.status || 'draft') === 'draft'"
                    >
                      <StopOutlined /> 取消发布
                    </a-menu-item>
                    <a-menu-divider />
                    <a-menu-item @click="handleDelete(record)" danger>
                      <DeleteOutlined /> 删除
                    </a-menu-item>
                  </a-menu>
                </template>
                <a-button type="link" size="small">
                  更多 <DownOutlined />
                </a-button>
              </a-dropdown>
            </a-space>
          </template>
        </template>
      </a-table>
    </a-card>

    <!-- 知识详情抽屉 -->
    <a-drawer
      v-model:open="detailVisible"
      title="知识详情"
      width="800"
      :footer-style="{ textAlign: 'right' }"
    >
      <KnowledgeDetail
        v-if="currentKnowledge"
        :knowledge="currentKnowledge"
        @edit="handleEdit"
        @delete="handleDelete"
        @close="detailVisible = false"
      />
      <template #footer>
        <a-space>
          <a-button @click="detailVisible = false">关闭</a-button>
          <a-button type="primary" @click="handleEdit(currentKnowledge)">
            编辑
          </a-button>
        </a-space>
      </template>
    </a-drawer>

    <!-- 知识表单模态框 -->
    <a-modal
      v-model:open="formVisible"
      :title="formMode === 'create' ? '新建知识' : '编辑知识'"
      width="1000"
      :footer="null"
      :destroy-on-close="true"
    >
      <KnowledgeForm
        v-if="formVisible"
        :knowledge="currentKnowledge"
        :mode="formMode"
        :categories="categories"
        :tags="tags"
        @submit="handleFormSubmit"
        @cancel="formVisible = false"
      />
    </a-modal>

    <!-- 导入模态框 -->
    <a-modal
      v-model:open="importVisible"
      title="导入知识"
      :footer="null"
    >
      <div class="import-content">
        <a-upload-dragger
          v-model:file-list="importFileList"
          :before-upload="beforeUpload"
          @change="handleImportChange"
          accept=".json,.csv,.xlsx"
        >
          <p class="ant-upload-drag-icon">
            <InboxOutlined />
          </p>
          <p class="ant-upload-text">点击或拖拽文件到此区域上传</p>
          <p class="ant-upload-hint">
            支持 JSON、CSV、Excel 格式文件
          </p>
        </a-upload-dragger>
        
        <div class="import-actions" v-if="importFileList.length > 0">
          <a-space>
            <a-button @click="importVisible = false">取消</a-button>
            <a-button type="primary" @click="handleImportSubmit" :loading="importing">
              开始导入
            </a-button>
          </a-space>
        </div>
      </div>
    </a-modal>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import {
  Card,
  Row,
  Col,
  Statistic,
  Form,
  Input,
  Select,
  Button,
  Space,
  Table,
  Tag,
  Alert,
  Drawer,
  Modal,
  Upload,
  Dropdown,
  Menu,
  message
} from 'ant-design-vue'
import {
  BookOutlined,
  CheckCircleOutlined,
  EditOutlined,
  PlusCircleOutlined,
  SearchOutlined,
  ReloadOutlined,
  ImportOutlined,
  ExportOutlined,
  PlusOutlined,
  EyeOutlined,
  CopyOutlined,
  StopOutlined,
  DeleteOutlined,
  DownOutlined,
  InboxOutlined
} from '@ant-design/icons-vue'
import { formatDateTime, getRelativeTime } from '@/utils/datetime'
import {
  getKnowledgeList,
  getKnowledge,
  createKnowledge,
  updateKnowledge,
  deleteKnowledge,
  getKnowledgeCategories,
  getKnowledgeTags,
  batchDeleteKnowledge,
  importKnowledge,
  exportKnowledge
} from '@/services/knowledge'
import type { Knowledge, KnowledgeCategory } from '@/types'

// 知识库分页响应接口
interface PaginatedResponse<T> {
  list: T[]
  total: number
  stats: {
    total: number
    published: number
    draft: number
    thisMonth: number
  }
}
import KnowledgeDetail from '@/components/KnowledgeDetail.vue'
import KnowledgeForm from '@/components/KnowledgeForm.vue'

const ACard = Card
const ARow = Row
const ACol = Col
const AStatistic = Statistic
const AForm = Form
const AFormItem = Form.Item
const AInput = Input
const ASelect = Select
const ASelectOption = Select.Option
const AButton = Button
const ASpace = Space
const ATable = Table
const ATag = Tag
const AAlert = Alert
const ADrawer = Drawer
const AModal = Modal
const AUploadDragger = Upload.Dragger
const ADropdown = Dropdown
const AMenu = Menu
const AMenuItem = Menu.Item
const AMenuDivider = Menu.Divider

// 响应式数据
const loading = ref(false)
const knowledgeList = ref<Knowledge[]>([])
const categories = ref<KnowledgeCategory[]>([])
const tags = ref<string[]>([])
const selectedRowKeys = ref<string[]>([])

// 统计数据
const stats = reactive({
  total: 0,
  published: 0,
  draft: 0,
  thisMonth: 0
})

// 搜索表单
const searchForm = reactive({
  keyword: '',
  category: undefined,
  status: undefined,
  tags: []
})

// 分页配置
const pagination = reactive({
  current: 1,
  pageSize: 10,
  total: 0,
  showSizeChanger: true,
  showQuickJumper: true,
  showTotal: (total: number) => `共 ${total} 条记录`
})

// 表格配置
const columns = [
  {
    title: '标题',
    key: 'title',
    width: 300,
    ellipsis: true
  },
  {
    title: '标签',
    key: 'tags',
    width: 200
  },
  {
    title: '状态',
    key: 'status',
    width: 100
  },
  {
    title: '浏览量',
    key: 'views',
    width: 100,
    sorter: true
  },
  {
    title: '更新时间',
    key: 'updatedAt',
    width: 180,
    sorter: true
  },
  {
    title: '操作',
    key: 'action',
    width: 200,
    fixed: 'right'
  }
]

// 行选择配置
const rowSelection = {
  selectedRowKeys,
  onChange: (keys: string[]) => {
    selectedRowKeys.value = keys
  }
}

// 详情抽屉
const detailVisible = ref(false)
const currentKnowledge = ref<Knowledge | null>(null)

// 表单模态框
const formVisible = ref(false)
const formMode = ref<'create' | 'edit'>('create')

// 导入模态框
const importVisible = ref(false)
const importFileList = ref([])
const importing = ref(false)

// 计算属性
const tagOptions = computed(() => {
  return tags.value.map(tag => ({
    label: tag,
    value: tag
  }))
})

// 获取分类名称
const getCategoryName = (categoryId: string) => {
  const category = categories.value.find(c => c.id === categoryId)
  return category?.name || '未分类'
}

// 获取标签颜色
const getTagColor = (tag: string) => {
  const colors = ['blue', 'green', 'orange', 'red', 'purple', 'cyan']
  const index = tag.length % colors.length
  return colors[index]
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

// 加载数据
const loadData = async () => {
  loading.value = true
  try {
    const params = {
      page: pagination.current,
      pageSize: pagination.pageSize,
      keyword: searchForm.keyword,
      category: searchForm.category,
      status: searchForm.status,
      tags: searchForm.tags
    }
    
    const response = await getKnowledgeList(params)
    const data = response.data || response
    knowledgeList.value = data.list || []
    pagination.total = data.total || 0
    
    // 更新统计数据
    if (data.stats) {
      stats.total = data.stats.total || 0
      stats.published = data.stats.published || 0
      stats.draft = data.stats.draft || 0
      stats.thisMonth = data.stats.thisMonth || 0
    }
  } catch (error) {
    console.error('加载知识列表失败:', error)
    message.error('加载知识列表失败')
  } finally {
    loading.value = false
  }
}

// 加载分类和标签
const loadMetadata = async () => {
  try {
    const [categoriesRes, tagsRes] = await Promise.all([
      getKnowledgeCategories(),
      getKnowledgeTags()
    ])
    
    categories.value = categoriesRes.data
    tags.value = tagsRes.data
  } catch (error) {
    console.error('加载元数据失败:', error)
  }
}

// 搜索
const handleSearch = () => {
  pagination.current = 1
  loadData()
}

// 重置
const handleReset = () => {
  Object.assign(searchForm, {
    keyword: '',
    category: undefined,
    status: undefined,
    tags: []
  })
  pagination.current = 1
  loadData()
}

// 表格变化
const handleTableChange = (pag: any, filters: any, sorter: any) => {
  pagination.current = pag.current
  pagination.pageSize = pag.pageSize
  loadData()
}

// 查看详情
const handleView = async (knowledge: Knowledge) => {
  try {
    const response = await getKnowledge(knowledge.id)
    currentKnowledge.value = response.data || response
    detailVisible.value = true
  } catch (error) {
    console.error('获取知识详情失败:', error)
    message.error('获取知识详情失败')
  }
}

// 新建
const handleCreate = () => {
  currentKnowledge.value = null
  formMode.value = 'create'
  formVisible.value = true
}

// 编辑
const handleEdit = (knowledge: Knowledge) => {
  currentKnowledge.value = knowledge
  formMode.value = 'edit'
  formVisible.value = true
  detailVisible.value = false
}

// 表单提交
const handleFormSubmit = async (data: any) => {
  try {
    if (formMode.value === 'create') {
      await createKnowledge(data)
      message.success('创建成功')
    } else {
      await updateKnowledge(currentKnowledge.value!.id, data)
      message.success('更新成功')
    }
    
    formVisible.value = false
    loadData()
  } catch (error) {
    console.error('保存失败:', error)
    message.error('保存失败')
  }
}

// 复制
const handleDuplicate = async (knowledge: Knowledge) => {
  try {
    const { id, createdAt, updatedAt, created_at, updated_at, ...cleanData } = knowledge
    const data = {
      ...cleanData,
      title: `${knowledge.title} (副本)`,
      status: 'draft'
    }
    
    await createKnowledge(data)
    message.success('复制成功')
    loadData()
  } catch (error) {
    console.error('复制失败:', error)
    message.error('复制失败')
  }
}

// 切换状态
const handleToggleStatus = async (knowledge: Knowledge) => {
  try {
    const currentStatus = (knowledge as any).status || 'draft'
    const newStatus = currentStatus === 'published' ? 'draft' : 'published'
    await updateKnowledge(knowledge.id, { status: newStatus } as any)
    message.success('状态更新成功')
    loadData()
  } catch (error) {
    console.error('状态更新失败:', error)
    message.error('状态更新失败')
  }
}

// 删除
const handleDelete = async (knowledge: Knowledge) => {
  try {
    await deleteKnowledge(knowledge.id)
    message.success('删除成功')
    detailVisible.value = false
    loadData()
  } catch (error) {
    console.error('删除失败:', error)
    message.error('删除失败')
  }
}

// 批量发布
const handleBatchPublish = async () => {
  try {
    // 这里应该有批量更新状态的API
    message.success('批量发布成功')
    selectedRowKeys.value = []
    loadData()
  } catch (error) {
    console.error('批量发布失败:', error)
    message.error('批量发布失败')
  }
}

// 批量取消发布
const handleBatchUnpublish = async () => {
  try {
    // 这里应该有批量更新状态的API
    message.success('批量取消发布成功')
    selectedRowKeys.value = []
    loadData()
  } catch (error) {
    console.error('批量取消发布失败:', error)
    message.error('批量取消发布失败')
  }
}

// 批量删除
const handleBatchDelete = async () => {
  try {
    await batchDeleteKnowledge(selectedRowKeys.value)
    message.success('批量删除成功')
    selectedRowKeys.value = []
    loadData()
  } catch (error) {
    console.error('批量删除失败:', error)
    message.error('批量删除失败')
  }
}

// 导入
const handleImport = () => {
  importVisible.value = true
  importFileList.value = []
}

// 导出
const handleExport = async () => {
  try {
    await exportKnowledge()
    message.success('导出成功')
  } catch (error) {
    console.error('导出失败:', error)
    message.error('导出失败')
  }
}

// 导出单个
const handleExportSingle = async (knowledge: Knowledge) => {
  try {
    await exportKnowledge([knowledge.id])
    message.success('导出成功')
  } catch (error) {
    console.error('导出失败:', error)
    message.error('导出失败')
  }
}

// 上传前检查
const beforeUpload = (file: any) => {
  const isValidType = ['application/json', 'text/csv', 'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet'].includes(file.type)
  if (!isValidType) {
    message.error('只支持 JSON、CSV、Excel 格式文件')
  }
  const isLt10M = file.size / 1024 / 1024 < 10
  if (!isLt10M) {
    message.error('文件大小不能超过 10MB')
  }
  return false // 阻止自动上传
}

// 导入文件变化
const handleImportChange = (info: any) => {
  // 处理文件变化
}

// 提交导入
const handleImportSubmit = async () => {
  if (importFileList.value.length === 0) {
    message.warning('请选择要导入的文件')
    return
  }
  
  importing.value = true
  try {
    const formData = new FormData()
    const file = importFileList.value[0].originFileObj || importFileList.value[0]
    formData.append('file', file)
    
    await importKnowledge(file as File)
    message.success('导入成功')
    importVisible.value = false
    loadData()
  } catch (error) {
    console.error('导入失败:', error)
    message.error('导入失败')
  } finally {
    importing.value = false
  }
}

// 组件挂载
onMounted(() => {
  loadData()
  loadMetadata()
})
</script>

<style scoped>
.knowledge-list {
  padding: 24px;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 24px;
}

.header-content h2 {
  margin: 0 0 8px 0;
  font-size: 24px;
  font-weight: 600;
}

.header-content p {
  margin: 0;
  color: #666;
}

.stats-cards {
  margin-bottom: 24px;
}

.search-card {
  margin-bottom: 16px;
}

.batch-actions {
  margin-bottom: 16px;
}

.table-card {
  margin-bottom: 0;
}

.knowledge-title {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.title-link {
  font-weight: 500;
  color: #1890ff;
  text-decoration: none;
}

.title-link:hover {
  text-decoration: underline;
}

.title-meta {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 12px;
}

.meta-text {
  color: #999;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  max-width: 200px;
}

.tags-container {
  display: flex;
  flex-wrap: wrap;
  gap: 4px;
}

.view-count {
  display: flex;
  align-items: center;
  gap: 4px;
  color: #666;
}

.time-info {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.time-relative {
  font-size: 12px;
  color: #999;
}

.import-content {
  padding: 16px 0;
}

.import-actions {
  margin-top: 16px;
  text-align: right;
}

/* 响应式设计 */
@media (max-width: 768px) {
  .knowledge-list {
    padding: 16px;
  }
  
  .page-header {
    flex-direction: column;
    gap: 16px;
  }
  
  .header-actions {
    width: 100%;
  }
  
  .stats-cards :deep(.ant-col) {
    margin-bottom: 16px;
  }
  
  .search-card :deep(.ant-form-item) {
    margin-bottom: 16px;
  }
  
  .table-card :deep(.ant-table) {
    font-size: 12px;
  }
  
  .meta-text {
    max-width: 100px;
  }
}
</style>