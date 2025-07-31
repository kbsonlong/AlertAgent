<template>
  <div class="rule-list">
    <!-- 页面标题和操作 -->
    <div class="page-header">
      <div class="header-content">
        <h2>告警规则</h2>
        <p>管理和配置告警规则，监控系统状态</p>
      </div>
      <div class="header-actions">
        <a-space>
          <a-button @click="importRules">
            <template #icon><ImportOutlined /></template>
            导入规则
          </a-button>
          <a-button @click="exportRules">
            <template #icon><ExportOutlined /></template>
            导出规则
          </a-button>
          <a-button type="primary" @click="showCreateModal">
            <template #icon><PlusOutlined /></template>
            新建规则
          </a-button>
        </a-space>
      </div>
    </div>

    <!-- 规则统计 -->
    <a-row :gutter="16" class="stats-row">
      <a-col :span="6">
        <a-card>
          <a-statistic
            title="总规则数"
            :value="stats.total"
            :value-style="{ color: '#1890ff' }"
          >
            <template #prefix><FileTextOutlined /></template>
          </a-statistic>
        </a-card>
      </a-col>
      <a-col :span="6">
        <a-card>
          <a-statistic
            title="启用规则"
            :value="stats.enabled"
            :value-style="{ color: '#52c41a' }"
          >
            <template #prefix><CheckCircleOutlined /></template>
          </a-statistic>
        </a-card>
      </a-col>
      <a-col :span="6">
        <a-card>
          <a-statistic
            title="禁用规则"
            :value="stats.disabled"
            :value-style="{ color: '#faad14' }"
          >
            <template #prefix><StopOutlined /></template>
          </a-statistic>
        </a-card>
      </a-col>
      <a-col :span="6">
        <a-card>
          <a-statistic
            title="触发中"
            :value="stats.firing"
            :value-style="{ color: '#ff4d4f' }"
          >
            <template #prefix><ExclamationCircleOutlined /></template>
          </a-statistic>
        </a-card>
      </a-col>
    </a-row>

    <!-- 搜索和筛选 -->
    <a-card class="search-card">
      <a-form layout="inline" :model="searchForm" @submit="handleSearch">
        <a-form-item label="规则名称">
          <a-input
            v-model:value="searchForm.name"
            placeholder="请输入规则名称"
            allow-clear
            style="width: 200px"
          />
        </a-form-item>
        <a-form-item label="状态">
          <a-select
            v-model:value="searchForm.status"
            placeholder="请选择状态"
            allow-clear
            style="width: 120px"
          >
            <a-select-option value="enabled">启用</a-select-option>
            <a-select-option value="disabled">禁用</a-select-option>
          </a-select>
        </a-form-item>
        <a-form-item label="严重程度">
          <a-select
            v-model:value="searchForm.severity"
            placeholder="请选择严重程度"
            allow-clear
            style="width: 120px"
          >
            <a-select-option value="critical">严重</a-select-option>
            <a-select-option value="warning">警告</a-select-option>
            <a-select-option value="info">信息</a-select-option>
          </a-select>
        </a-form-item>
        <a-form-item label="数据源">
          <a-select
            v-model:value="searchForm.provider"
            placeholder="请选择数据源"
            allow-clear
            style="width: 150px"
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
        <a-form-item>
          <a-space>
            <a-button type="primary" html-type="submit">
              <template #icon><SearchOutlined /></template>
              搜索
            </a-button>
            <a-button @click="resetSearch">
              重置
            </a-button>
          </a-space>
        </a-form-item>
      </a-form>
    </a-card>

    <!-- 批量操作 -->
    <div class="batch-actions" v-if="selectedRowKeys.length > 0">
      <a-space>
        <span>已选择 {{ selectedRowKeys.length }} 项</span>
        <a-button @click="batchEnable">
          <template #icon><PlayCircleOutlined /></template>
          批量启用
        </a-button>
        <a-button @click="batchDisable">
          <template #icon><PauseCircleOutlined /></template>
          批量禁用
        </a-button>
        <a-button danger @click="batchDelete">
          <template #icon><DeleteOutlined /></template>
          批量删除
        </a-button>
      </a-space>
    </div>

    <!-- 规则表格 -->
    <a-card>
      <a-table
        :columns="columns"
        :data-source="rules"
        :loading="loading"
        :pagination="pagination"
        :row-selection="rowSelection"
        @change="handleTableChange"
        row-key="id"
      >
        <template #name="{ record }">
          <div class="rule-name">
            <a @click="viewRule(record)">{{ record.name }}</a>
            <div class="rule-description">{{ record.description }}</div>
          </div>
        </template>
        
        <template #status="{ record }">
          <a-switch
            :checked="record.enabled"
            @change="(checked) => toggleRule(record, checked)"
            :loading="record.toggling"
          >
            <template #checkedChildren>启用</template>
            <template #unCheckedChildren>禁用</template>
          </a-switch>
        </template>
        
        <template #severity="{ record }">
          <a-tag :color="getSeverityColor(record.severity)">
            {{ getSeverityText(record.severity) }}
          </a-tag>
        </template>
        
        <template #provider="{ record }">
          <a-tag>{{ getProviderName(record.provider_id) }}</a-tag>
        </template>
        
        <template #firing="{ record }">
          <a-badge
            :count="record.firing_count"
            :number-style="{ backgroundColor: record.firing_count > 0 ? '#ff4d4f' : '#52c41a' }"
          >
            <span>{{ record.firing_count > 0 ? '触发中' : '正常' }}</span>
          </a-badge>
        </template>
        
        <template #updated_at="{ record }">
          {{ formatDateTime(record.updated_at) }}
        </template>
        
        <template #action="{ record }">
          <a-space>
            <a-button type="link" size="small" @click="viewRule(record)">
              查看
            </a-button>
            <a-button type="link" size="small" @click="editRule(record)">
              编辑
            </a-button>
            <a-button type="link" size="small" @click="testRule(record)">
              测试
            </a-button>
            <a-dropdown>
              <template #overlay>
                <a-menu>
                  <a-menu-item @click="duplicateRule(record)">
                    <CopyOutlined /> 复制
                  </a-menu-item>
                  <a-menu-item @click="syncRule(record)">
                    <SyncOutlined /> 同步
                  </a-menu-item>
                  <a-menu-divider />
                  <a-menu-item @click="deleteRule(record)" class="danger-item">
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
      </a-table>
    </a-card>

    <!-- 创建/编辑规则弹窗 -->
    <a-modal
      v-model:open="modalVisible"
      :title="modalTitle"
      width="800px"
      :confirm-loading="modalLoading"
      @ok="handleModalOk"
      @cancel="handleModalCancel"
    >
      <rule-form
        ref="ruleFormRef"
        :rule="currentRule"
        :providers="providers"
        @submit="handleRuleSubmit"
      />
    </a-modal>

    <!-- 规则详情抽屉 -->
    <a-drawer
      v-model:open="drawerVisible"
      title="规则详情"
      width="600px"
      placement="right"
    >
      <rule-detail
        v-if="selectedRule"
        :rule="selectedRule"
        @edit="editRule"
        @delete="deleteRule"
        @test="testRule"
      />
    </a-drawer>

    <!-- 测试结果弹窗 -->
    <a-modal
      v-model:open="testModalVisible"
      title="规则测试结果"
      width="600px"
      :footer="null"
    >
      <div v-if="testResult">
        <a-result
          :status="testResult.success ? 'success' : 'error'"
          :title="testResult.success ? '测试通过' : '测试失败'"
          :sub-title="testResult.message"
        >
          <template #extra>
            <a-button type="primary" @click="testModalVisible = false">
              关闭
            </a-button>
          </template>
        </a-result>
        
        <div v-if="testResult.details" class="test-details">
          <h4>详细信息</h4>
          <pre>{{ JSON.stringify(testResult.details, null, 2) }}</pre>
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
  Switch,
  Badge,
  Dropdown,
  Menu,
  Modal,
  Drawer,
  Result,
  message
} from 'ant-design-vue'
import {
  PlusOutlined,
  ImportOutlined,
  ExportOutlined,
  FileTextOutlined,
  CheckCircleOutlined,
  StopOutlined,
  ExclamationCircleOutlined,
  SearchOutlined,
  PlayCircleOutlined,
  PauseCircleOutlined,
  DeleteOutlined,
  CopyOutlined,
  SyncOutlined,
  DownOutlined
} from '@ant-design/icons-vue'
import { formatDateTime } from '@/utils/datetime'
import { useRuleStore } from '@/stores'
import {
  getRules,
  createRule,
  updateRule,
  deleteRule as deleteRuleApi,
  toggleRule as toggleRuleApi,
  testRule as testRuleApi,
  syncRule as syncRuleApi,
  batchDeleteRules,
  importRules as importRulesApi,
  exportRules as exportRulesApi
} from '@/services/rule'
import { getProviders } from '@/services/provider'
import type { Rule, Provider, CreateRuleRequest, UpdateRuleRequest } from '@/types'
import RuleForm from '@/components/RuleForm.vue'
import RuleDetail from '@/components/RuleDetail.vue'

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
const ASwitch = Switch
const ABadge = Badge
const ADropdown = Dropdown
const AMenu = Menu
const AMenuItem = Menu.Item
const AMenuDivider = Menu.Divider
const AModal = Modal
const ADrawer = Drawer
const AResult = Result

const ruleStore = useRuleStore()

// 响应式数据
const loading = ref(false)
const rules = ref<Rule[]>([])
const providers = ref<Provider[]>([])
const selectedRowKeys = ref<string[]>([])
const modalVisible = ref(false)
const modalLoading = ref(false)
const drawerVisible = ref(false)
const testModalVisible = ref(false)
const currentRule = ref<Rule | null>(null)
const selectedRule = ref<Rule | null>(null)
const testResult = ref<any>(null)
const ruleFormRef = ref()

// 搜索表单
const searchForm = reactive({
  name: '',
  status: undefined,
  severity: undefined,
  provider: undefined
})

// 分页配置
const pagination = reactive({
  current: 1,
  pageSize: 20,
  total: 0,
  showSizeChanger: true,
  showQuickJumper: true,
  showTotal: (total: number) => `共 ${total} 条记录`
})

// 统计数据
const stats = computed(() => {
  const total = rules.value.length
  const enabled = rules.value.filter(rule => rule.enabled).length
  const disabled = total - enabled
  const firing = rules.value.filter(rule => rule.firing_count > 0).length
  
  return {
    total,
    enabled,
    disabled,
    firing
  }
})

// 弹窗标题
const modalTitle = computed(() => {
  return currentRule.value ? '编辑规则' : '新建规则'
})

// 表格列配置
const columns = [
  {
    title: '规则名称',
    dataIndex: 'name',
    key: 'name',
    ellipsis: true,
    slots: { customRender: 'name' }
  },
  {
    title: '状态',
    dataIndex: 'enabled',
    key: 'status',
    width: 100,
    slots: { customRender: 'status' }
  },
  {
    title: '严重程度',
    dataIndex: 'severity',
    key: 'severity',
    width: 100,
    slots: { customRender: 'severity' }
  },
  {
    title: '数据源',
    dataIndex: 'provider_id',
    key: 'provider',
    width: 120,
    slots: { customRender: 'provider' }
  },
  {
    title: '触发状态',
    dataIndex: 'firing_count',
    key: 'firing',
    width: 100,
    slots: { customRender: 'firing' }
  },
  {
    title: '更新时间',
    dataIndex: 'updated_at',
    key: 'updated_at',
    width: 180,
    slots: { customRender: 'updated_at' }
  },
  {
    title: '操作',
    key: 'action',
    width: 200,
    slots: { customRender: 'action' }
  }
]

// 行选择配置
const rowSelection = {
  selectedRowKeys,
  onChange: (keys: string[]) => {
    selectedRowKeys.value = keys
  }
}

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

// 获取数据源名称
const getProviderName = (providerId: string) => {
  const provider = providers.value.find(p => p.id === providerId)
  return provider?.name || '未知'
}

// 加载数据
const loadData = async () => {
  try {
    loading.value = true
    const params = {
      page: pagination.current,
      page_size: pagination.pageSize,
      ...searchForm
    }
    
    const response = await getRules(params)
    rules.value = response.data.items
    pagination.total = response.data.total
  } catch (error) {
    message.error('加载规则列表失败')
  } finally {
    loading.value = false
  }
}

// 加载数据源
const loadProviders = async () => {
  try {
    const response = await getProviders()
    providers.value = response.data.items
  } catch (error) {
    console.error('加载数据源失败:', error)
  }
}

// 搜索
const handleSearch = () => {
  pagination.current = 1
  loadData()
}

// 重置搜索
const resetSearch = () => {
  Object.assign(searchForm, {
    name: '',
    status: undefined,
    severity: undefined,
    provider: undefined
  })
  pagination.current = 1
  loadData()
}

// 表格变化
const handleTableChange = (pag: any) => {
  pagination.current = pag.current
  pagination.pageSize = pag.pageSize
  loadData()
}

// 显示创建弹窗
const showCreateModal = () => {
  currentRule.value = null
  modalVisible.value = true
}

// 查看规则
const viewRule = (rule: Rule) => {
  selectedRule.value = rule
  drawerVisible.value = true
}

// 编辑规则
const editRule = (rule: Rule) => {
  currentRule.value = rule
  modalVisible.value = true
}

// 切换规则状态
const toggleRule = async (rule: Rule, enabled: boolean) => {
  try {
    rule.toggling = true
    await toggleRuleApi(rule.id, enabled)
    rule.enabled = enabled
    message.success(`规则已${enabled ? '启用' : '禁用'}`)
  } catch (error) {
    message.error(`${enabled ? '启用' : '禁用'}规则失败`)
  } finally {
    rule.toggling = false
  }
}

// 测试规则
const testRule = async (rule: Rule) => {
  try {
    const response = await testRuleApi(rule.id)
    testResult.value = response.data
    testModalVisible.value = true
  } catch (error) {
    message.error('测试规则失败')
  }
}

// 复制规则
const duplicateRule = (rule: Rule) => {
  const newRule = {
    ...rule,
    name: `${rule.name} (副本)`,
    id: undefined
  }
  currentRule.value = newRule as Rule
  modalVisible.value = true
}

// 同步规则
const syncRule = async (rule: Rule) => {
  try {
    await syncRuleApi(rule.id)
    message.success('规则同步成功')
    loadData()
  } catch (error) {
    message.error('规则同步失败')
  }
}

// 删除规则
const deleteRule = (rule: Rule) => {
  Modal.confirm({
    title: '确认删除',
    content: `确定要删除规则 "${rule.name}" 吗？`,
    okText: '确定',
    cancelText: '取消',
    onOk: async () => {
      try {
        await deleteRuleApi(rule.id)
        message.success('删除成功')
        loadData()
      } catch (error) {
        message.error('删除失败')
      }
    }
  })
}

// 批量启用
const batchEnable = async () => {
  try {
    for (const id of selectedRowKeys.value) {
      await toggleRuleApi(id, true)
    }
    message.success('批量启用成功')
    selectedRowKeys.value = []
    loadData()
  } catch (error) {
    message.error('批量启用失败')
  }
}

// 批量禁用
const batchDisable = async () => {
  try {
    for (const id of selectedRowKeys.value) {
      await toggleRuleApi(id, false)
    }
    message.success('批量禁用成功')
    selectedRowKeys.value = []
    loadData()
  } catch (error) {
    message.error('批量禁用失败')
  }
}

// 批量删除
const batchDelete = () => {
  Modal.confirm({
    title: '确认删除',
    content: `确定要删除选中的 ${selectedRowKeys.value.length} 个规则吗？`,
    okText: '确定',
    cancelText: '取消',
    onOk: async () => {
      try {
        await batchDeleteRules(selectedRowKeys.value)
        message.success('批量删除成功')
        selectedRowKeys.value = []
        loadData()
      } catch (error) {
        message.error('批量删除失败')
      }
    }
  })
}

// 导入规则
const importRules = () => {
  const input = document.createElement('input')
  input.type = 'file'
  input.accept = '.json,.yaml,.yml'
  input.onchange = async (e) => {
    const file = (e.target as HTMLInputElement).files?.[0]
    if (file) {
      try {
        const formData = new FormData()
        formData.append('file', file)
        await importRulesApi(formData)
        message.success('导入成功')
        loadData()
      } catch (error) {
        message.error('导入失败')
      }
    }
  }
  input.click()
}

// 导出规则
const exportRules = async () => {
  try {
    const response = await exportRulesApi()
    const blob = new Blob([response.data], { type: 'application/json' })
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = `rules-${Date.now()}.json`
    a.click()
    URL.revokeObjectURL(url)
    message.success('导出成功')
  } catch (error) {
    message.error('导出失败')
  }
}

// 弹窗确定
const handleModalOk = () => {
  ruleFormRef.value?.submit()
}

// 弹窗取消
const handleModalCancel = () => {
  modalVisible.value = false
  currentRule.value = null
}

// 规则提交
const handleRuleSubmit = async (formData: CreateRuleRequest | UpdateRuleRequest) => {
  try {
    modalLoading.value = true
    
    if (currentRule.value) {
      await updateRule(currentRule.value.id, formData as UpdateRuleRequest)
      message.success('更新成功')
    } else {
      await createRule(formData as CreateRuleRequest)
      message.success('创建成功')
    }
    
    modalVisible.value = false
    currentRule.value = null
    loadData()
  } catch (error) {
    message.error(currentRule.value ? '更新失败' : '创建失败')
  } finally {
    modalLoading.value = false
  }
}

// 组件挂载
onMounted(() => {
  loadData()
  loadProviders()
})
</script>

<style scoped>
.rule-list {
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

.stats-row {
  margin-bottom: 24px;
}

.search-card {
  margin-bottom: 16px;
}

.batch-actions {
  margin-bottom: 16px;
  padding: 12px 16px;
  background: #f5f5f5;
  border-radius: 6px;
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.rule-name {
  display: flex;
  flex-direction: column;
}

.rule-name a {
  font-weight: 500;
  margin-bottom: 4px;
}

.rule-description {
  font-size: 12px;
  color: #999;
  line-height: 1.4;
}

.danger-item {
  color: #ff4d4f !important;
}

.test-details {
  margin-top: 16px;
  padding-top: 16px;
  border-top: 1px solid #f0f0f0;
}

.test-details h4 {
  margin-bottom: 12px;
}

.test-details pre {
  background: #f5f5f5;
  padding: 12px;
  border-radius: 4px;
  font-size: 12px;
  max-height: 300px;
  overflow-y: auto;
}

/* 响应式设计 */
@media (max-width: 768px) {
  .rule-list {
    padding: 16px;
  }
  
  .page-header {
    flex-direction: column;
    gap: 16px;
  }
  
  .header-actions {
    width: 100%;
  }
  
  .stats-row :deep(.ant-col) {
    margin-bottom: 16px;
  }
  
  .search-card :deep(.ant-form-item) {
    margin-bottom: 16px;
  }
  
  .batch-actions {
    flex-direction: column;
    gap: 12px;
    align-items: flex-start;
  }
}
</style>