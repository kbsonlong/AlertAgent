<template>
  <div class="exception-management">
    <!-- 搜索和筛选 -->
    <div class="search-section">
      <a-row :gutter="[16, 16]">
        <a-col :span="6">
          <a-input
            v-model:value="searchForm.keyword"
            placeholder="搜索异常信息"
            allow-clear
            @change="handleSearch"
          >
            <template #prefix>
              <SearchOutlined />
            </template>
          </a-input>
        </a-col>
        <a-col :span="4">
          <a-select
            v-model:value="searchForm.status"
            placeholder="状态"
            allow-clear
            @change="handleSearch"
          >
            <a-select-option value="pending">待处理</a-select-option>
            <a-select-option value="processing">处理中</a-select-option>
            <a-select-option value="resolved">已解决</a-select-option>
            <a-select-option value="ignored">已忽略</a-select-option>
          </a-select>
        </a-col>
        <a-col :span="4">
          <a-select
            v-model:value="searchForm.severity"
            placeholder="严重程度"
            allow-clear
            @change="handleSearch"
          >
            <a-select-option value="critical">严重</a-select-option>
            <a-select-option value="warning">警告</a-select-option>
            <a-select-option value="info">信息</a-select-option>
          </a-select>
        </a-col>
        <a-col :span="6">
          <a-range-picker
            v-model:value="searchForm.dateRange"
            @change="handleSearch"
            size="default"
          />
        </a-col>
        <a-col :span="4">
          <a-space>
            <a-button type="primary" @click="handleSearch">
              <template #icon><SearchOutlined /></template>
              搜索
            </a-button>
            <a-button @click="handleReset">
              重置
            </a-button>
          </a-space>
        </a-col>
      </a-row>
    </div>

    <!-- 异常列表 -->
    <div class="exception-list">
      <a-table
        :columns="columns"
        :data-source="exceptionList"
        :loading="loading"
        :pagination="{
          current: pagination.current,
          pageSize: pagination.pageSize,
          total: pagination.total,
          showSizeChanger: true,
          showQuickJumper: true,
          showTotal: (total, range) => `第 ${range[0]}-${range[1]} 条，共 ${total} 条`
        }"
        @change="handleTableChange"
        size="small"
      >
        <template #bodyCell="{ column, record }">
          <template v-if="column.key === 'severity'">
            <a-tag :color="getSeverityColor(record.severity)">
              {{ getSeverityText(record.severity) }}
            </a-tag>
          </template>
          <template v-else-if="column.key === 'status'">
            <a-tag :color="getStatusColor(record.status)">
              {{ getStatusText(record.status) }}
            </a-tag>
          </template>
          <template v-else-if="column.key === 'created_at'">
            {{ formatDateTime(record.created_at) }}
          </template>
          <template v-else-if="column.key === 'action'">
            <a-space>
              <a-button
                type="link"
                size="small"
                @click="handleView(record)"
              >
                查看
              </a-button>
              <a-button
                v-if="record.status === 'pending'"
                type="link"
                size="small"
                @click="handleProcess(record)"
              >
                处理
              </a-button>
              <a-button
                v-if="record.status !== 'resolved'"
                type="link"
                size="small"
                @click="handleResolve(record)"
              >
                标记解决
              </a-button>
              <a-button
                v-if="record.status !== 'ignored'"
                type="link"
                size="small"
                danger
                @click="handleIgnore(record)"
              >
                忽略
              </a-button>
            </a-space>
          </template>
        </template>
      </a-table>
    </div>

    <!-- 异常详情抽屉 -->
    <a-drawer
      v-model:open="detailVisible"
      title="异常详情"
      width="600"
      @close="handleDetailClose"
    >
      <div v-if="currentException" class="exception-detail">
        <a-descriptions :column="1" bordered>
          <a-descriptions-item label="异常ID">
            {{ currentException.id }}
          </a-descriptions-item>
          <a-descriptions-item label="集群ID">
            {{ currentException.cluster_id }}
          </a-descriptions-item>
          <a-descriptions-item label="配置类型">
            {{ currentException.config_type }}
          </a-descriptions-item>
          <a-descriptions-item label="严重程度">
            <a-tag :color="getSeverityColor(currentException.severity)">
              {{ getSeverityText(currentException.severity) }}
            </a-tag>
          </a-descriptions-item>
          <a-descriptions-item label="状态">
            <a-tag :color="getStatusColor(currentException.status)">
              {{ getStatusText(currentException.status) }}
            </a-tag>
          </a-descriptions-item>
          <a-descriptions-item label="异常信息">
            <pre>{{ currentException.error_message }}</pre>
          </a-descriptions-item>
          <a-descriptions-item label="堆栈信息" v-if="currentException.stack_trace">
            <pre>{{ currentException.stack_trace }}</pre>
          </a-descriptions-item>
          <a-descriptions-item label="发生时间">
            {{ formatDateTime(currentException.created_at) }}
          </a-descriptions-item>
          <a-descriptions-item label="处理时间" v-if="currentException.processed_at">
            {{ formatDateTime(currentException.processed_at) }}
          </a-descriptions-item>
          <a-descriptions-item label="处理备注" v-if="currentException.process_note">
            {{ currentException.process_note }}
          </a-descriptions-item>
        </a-descriptions>
      </div>
    </a-drawer>

    <!-- 处理异常模态框 -->
    <a-modal
      v-model:open="processVisible"
      title="处理异常"
      @ok="handleProcessSubmit"
      @cancel="handleProcessCancel"
    >
      <a-form :model="processForm" layout="vertical">
        <a-form-item label="处理备注" required>
          <a-textarea
            v-model:value="processForm.note"
            placeholder="请输入处理备注"
            :rows="4"
          />
        </a-form-item>
        <a-form-item label="处理方式">
          <a-radio-group v-model:value="processForm.action">
            <a-radio value="resolve">标记为已解决</a-radio>
            <a-radio value="ignore">忽略此异常</a-radio>
            <a-radio value="retry">重试同步</a-radio>
          </a-radio-group>
        </a-form-item>
      </a-form>
    </a-modal>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { message } from 'ant-design-vue'
import {
  SearchOutlined
} from '@ant-design/icons-vue'
import { formatDateTime } from '@/utils/datetime'

// 组件事件
interface Emits {
  close: []
}

const emit = defineEmits<Emits>()

// 响应式数据
const loading = ref(false)
const exceptionList = ref<any[]>([])
const detailVisible = ref(false)
const processVisible = ref(false)
const currentException = ref<any>(null)

// 搜索表单
const searchForm = reactive({
  keyword: '',
  status: undefined,
  severity: undefined,
  dateRange: undefined
})

// 处理表单
const processForm = reactive({
  note: '',
  action: 'resolve'
})

// 分页配置
const pagination = reactive({
  current: 1,
  pageSize: 10,
  total: 0
})

// 表格列定义
const columns = [
  {
    title: 'ID',
    dataIndex: 'id',
    key: 'id',
    width: 80
  },
  {
    title: '集群ID',
    dataIndex: 'cluster_id',
    key: 'cluster_id',
    width: 120
  },
  {
    title: '配置类型',
    dataIndex: 'config_type',
    key: 'config_type',
    width: 100
  },
  {
    title: '严重程度',
    dataIndex: 'severity',
    key: 'severity',
    width: 100
  },
  {
    title: '状态',
    dataIndex: 'status',
    key: 'status',
    width: 100
  },
  {
    title: '异常信息',
    dataIndex: 'error_message',
    key: 'error_message',
    ellipsis: true
  },
  {
    title: '发生时间',
    dataIndex: 'created_at',
    key: 'created_at',
    width: 160
  },
  {
    title: '操作',
    key: 'action',
    width: 200,
    fixed: 'right'
  }
]

// 获取严重程度颜色
const getSeverityColor = (severity: string) => {
  const colorMap: Record<string, string> = {
    critical: 'red',
    warning: 'orange',
    info: 'blue'
  }
  return colorMap[severity] || 'default'
}

// 获取严重程度文本
const getSeverityText = (severity: string) => {
  const textMap: Record<string, string> = {
    critical: '严重',
    warning: '警告',
    info: '信息'
  }
  return textMap[severity] || severity
}

// 获取状态颜色
const getStatusColor = (status: string) => {
  const colorMap: Record<string, string> = {
    pending: 'orange',
    processing: 'blue',
    resolved: 'green',
    ignored: 'gray'
  }
  return colorMap[status] || 'default'
}

// 获取状态文本
const getStatusText = (status: string) => {
  const textMap: Record<string, string> = {
    pending: '待处理',
    processing: '处理中',
    resolved: '已解决',
    ignored: '已忽略'
  }
  return textMap[status] || status
}

// 加载异常列表
const loadExceptions = async () => {
  loading.value = true
  try {
    // 模拟数据
    const mockData = [
      {
        id: 1,
        cluster_id: 'cluster-001',
        config_type: 'prometheus',
        severity: 'critical',
        status: 'pending',
        error_message: '配置同步失败：连接超时',
        stack_trace: 'Error: Connection timeout\n  at sync.js:123\n  at ...',
        created_at: '2024-01-01 10:00:00',
        processed_at: null,
        process_note: null
      },
      {
        id: 2,
        cluster_id: 'cluster-002',
        config_type: 'alertmanager',
        severity: 'warning',
        status: 'resolved',
        error_message: '配置验证失败：格式错误',
        stack_trace: null,
        created_at: '2024-01-01 09:30:00',
        processed_at: '2024-01-01 10:30:00',
        process_note: '已修复配置格式问题'
      }
    ]
    
    exceptionList.value = mockData
    pagination.total = mockData.length
  } catch (error) {
    message.error('加载异常列表失败')
  } finally {
    loading.value = false
  }
}

// 搜索处理
const handleSearch = () => {
  pagination.current = 1
  loadExceptions()
}

// 重置搜索
const handleReset = () => {
  Object.assign(searchForm, {
    keyword: '',
    status: undefined,
    severity: undefined,
    dateRange: undefined
  })
  handleSearch()
}

// 表格变化处理
const handleTableChange = (pag: any) => {
  pagination.current = pag.current
  pagination.pageSize = pag.pageSize
  loadExceptions()
}

// 查看异常详情
const handleView = (record: any) => {
  currentException.value = record
  detailVisible.value = true
}

// 处理异常
const handleProcess = (record: any) => {
  currentException.value = record
  processForm.note = ''
  processForm.action = 'resolve'
  processVisible.value = true
}

// 标记解决
const handleResolve = async (record: any) => {
  try {
    // 这里应该调用API
    message.success('已标记为解决')
    loadExceptions()
  } catch (error) {
    message.error('操作失败')
  }
}

// 忽略异常
const handleIgnore = async (record: any) => {
  try {
    // 这里应该调用API
    message.success('已忽略异常')
    loadExceptions()
  } catch (error) {
    message.error('操作失败')
  }
}

// 详情抽屉关闭
const handleDetailClose = () => {
  detailVisible.value = false
  currentException.value = null
}

// 处理提交
const handleProcessSubmit = async () => {
  if (!processForm.note.trim()) {
    message.error('请输入处理备注')
    return
  }
  
  try {
    // 这里应该调用API
    message.success('处理成功')
    processVisible.value = false
    loadExceptions()
  } catch (error) {
    message.error('处理失败')
  }
}

// 处理取消
const handleProcessCancel = () => {
  processVisible.value = false
  processForm.note = ''
  processForm.action = 'resolve'
}

// 组件挂载
onMounted(() => {
  loadExceptions()
})
</script>

<style scoped>
.exception-management {
  padding: 16px;
}

.search-section {
  margin-bottom: 16px;
  padding: 16px;
  background: #fafafa;
  border-radius: 6px;
}

.exception-list {
  background: white;
  border-radius: 6px;
}

.exception-detail pre {
  background: #f5f5f5;
  padding: 8px;
  border-radius: 4px;
  font-size: 12px;
  line-height: 1.4;
  max-height: 200px;
  overflow-y: auto;
}

:deep(.ant-table-tbody > tr > td) {
  padding: 8px 16px;
}

:deep(.ant-descriptions-item-content pre) {
  margin: 0;
  white-space: pre-wrap;
  word-break: break-all;
}
</style>