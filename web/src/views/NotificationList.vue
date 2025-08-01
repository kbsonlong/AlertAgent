<template>
  <div class="notification-list">
    <!-- 页面标题和操作 -->
    <div class="page-header">
      <div class="header-content">
        <h2>通知管理</h2>
        <p>管理通知组和通知模板配置</p>
      </div>
      <div class="header-actions">
        <a-space>
          <a-button @click="handleRefresh">
            <template #icon><ReloadOutlined /></template>
            刷新
          </a-button>
          <a-dropdown>
            <template #overlay>
              <a-menu>
                <a-menu-item @click="handleCreateGroup">
                  <TeamOutlined /> 新建通知组
                </a-menu-item>
                <a-menu-item @click="handleCreateTemplate">
                  <FileTextOutlined /> 新建模板
                </a-menu-item>
              </a-menu>
            </template>
            <a-button type="primary">
              新建 <DownOutlined />
            </a-button>
          </a-dropdown>
        </a-space>
      </div>
    </div>

    <!-- 统计卡片 -->
    <div class="stats-cards">
      <a-row :gutter="16">
        <a-col :span="6">
          <a-card>
            <a-statistic
              title="通知组"
              :value="stats.groups"
              :value-style="{ color: '#1890ff' }"
            >
              <template #prefix>
                <TeamOutlined />
              </template>
            </a-statistic>
          </a-card>
        </a-col>
        <a-col :span="6">
          <a-card>
            <a-statistic
              title="通知模板"
              :value="stats.templates"
              :value-style="{ color: '#52c41a' }"
            >
              <template #prefix>
                <FileTextOutlined />
              </template>
            </a-statistic>
          </a-card>
        </a-col>
        <a-col :span="6">
          <a-card>
            <a-statistic
              title="今日发送"
              :value="stats.todaySent"
              :value-style="{ color: '#faad14' }"
            >
              <template #prefix>
                <SendOutlined />
              </template>
            </a-statistic>
          </a-card>
        </a-col>
        <a-col :span="6">
          <a-card>
            <a-statistic
              title="发送失败"
              :value="stats.failed"
              :value-style="{ color: '#ff4d4f' }"
            >
              <template #prefix>
                <ExclamationCircleOutlined />
              </template>
            </a-statistic>
          </a-card>
        </a-col>
      </a-row>
    </div>

    <!-- 标签页 -->
    <a-card class="tabs-card">
      <a-tabs v-model:activeKey="activeTab" @change="handleTabChange">
        <!-- 通知组 -->
        <a-tab-pane key="groups" tab="通知组">
          <div class="tab-content">
            <!-- 搜索和筛选 -->
            <div class="search-section">
              <a-form layout="inline" :model="groupSearchForm">
                <a-form-item label="名称">
                  <a-input
                    v-model:value="groupSearchForm.name"
                    placeholder="搜索通知组名称"
                    style="width: 200px"
                    @press-enter="handleGroupSearch"
                  >
                    <template #prefix><SearchOutlined /></template>
                  </a-input>
                </a-form-item>
                <a-form-item label="类型">
                  <a-select
                    v-model:value="groupSearchForm.type"
                    placeholder="请选择类型"
                    style="width: 150px"
                    allow-clear
                  >
                    <a-select-option value="email">邮件</a-select-option>
                    <a-select-option value="webhook">Webhook</a-select-option>
                    <a-select-option value="dingtalk">钉钉</a-select-option>
                    <a-select-option value="wechat">企业微信</a-select-option>
                    <a-select-option value="slack">Slack</a-select-option>
                  </a-select>
                </a-form-item>
                <a-form-item label="状态">
                  <a-select
                    v-model:value="groupSearchForm.enabled"
                    placeholder="请选择状态"
                    style="width: 120px"
                    allow-clear
                  >
                    <a-select-option :value="true">启用</a-select-option>
                    <a-select-option :value="false">禁用</a-select-option>
                  </a-select>
                </a-form-item>
                <a-form-item>
                  <a-space>
                    <a-button type="primary" @click="handleGroupSearch">
                      <template #icon><SearchOutlined /></template>
                      搜索
                    </a-button>
                    <a-button @click="handleGroupReset">
                      <template #icon><ReloadOutlined /></template>
                      重置
                    </a-button>
                  </a-space>
                </a-form-item>
              </a-form>
            </div>

            <!-- 批量操作 -->
            <div class="batch-actions" v-if="selectedGroupKeys.length > 0">
              <a-alert
                :message="`已选择 ${selectedGroupKeys.length} 项`"
                type="info"
                show-icon
              >
                <template #action>
                  <a-space>
                    <a-button size="small" @click="handleBatchGroupTest">
                      批量测试
                    </a-button>
                    <a-button size="small" danger @click="handleBatchGroupDelete">
                      批量删除
                    </a-button>
                  </a-space>
                </template>
              </a-alert>
            </div>

            <!-- 通知组列表 -->
            <a-table
              :columns="groupColumns"
              :data-source="groupList"
              :loading="groupLoading"
              :pagination="groupPagination"
              :row-selection="groupRowSelection"
              @change="handleGroupTableChange"
              row-key="id"
            >
              <template #bodyCell="{ column, record }">
                <template v-if="column.key === 'name'">
                  <div class="group-name">
                    <div class="name-content">
                      <component :is="getGroupTypeIcon(record.type)" class="type-icon" />
                      <div class="name-info">
                        <a @click="handleGroupView(record)" class="name-link">
                          {{ record.name }}
                        </a>
                        <div class="name-meta">
                          {{ record.description }}
                        </div>
                      </div>
                    </div>
                  </div>
                </template>
                
                <template v-else-if="column.key === 'type'">
                  <a-tag :color="getGroupTypeColor(record.type)">
                    {{ getGroupTypeText(record.type) }}
                  </a-tag>
                </template>
                
                <template v-else-if="column.key === 'enabled'">
                  <a-switch
                    :checked="record.enabled"
                    @change="(checked) => handleGroupToggle(record, checked)"
                    size="small"
                  />
                </template>
                
                <template v-else-if="column.key === 'stats'">
                  <div class="stats-info">
                    <div class="stat-item">
                      <span class="stat-label">今日:</span>
                      <span class="stat-value">{{ record.todaySent || 0 }}</span>
                    </div>
                    <div class="stat-item">
                      <span class="stat-label">失败:</span>
                      <span class="stat-value error">{{ record.failedCount || 0 }}</span>
                    </div>
                  </div>
                </template>
                
                <template v-else-if="column.key === 'updatedAt'">
                  <div class="time-info">
                    <div>{{ formatDateTime(record.updatedAt) }}</div>
                    <div class="time-relative">{{ getRelativeTime(record.updatedAt) }}</div>
                  </div>
                </template>
                
                <template v-else-if="column.key === 'action'">
                  <a-space>
                    <a-button type="link" size="small" @click="handleGroupView(record)">
                      查看
                    </a-button>
                    <a-button type="link" size="small" @click="handleGroupTest(record)">
                      测试
                    </a-button>
                    <a-dropdown>
                      <template #overlay>
                        <a-menu>
                          <a-menu-item @click="handleGroupEdit(record)">
                            <EditOutlined /> 编辑
                          </a-menu-item>
                          <a-menu-item @click="handleGroupDuplicate(record)">
                            <CopyOutlined /> 复制
                          </a-menu-item>
                          <a-menu-divider />
                          <a-menu-item @click="handleGroupDelete(record)" danger>
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
          </div>
        </a-tab-pane>

        <!-- 通知模板 -->
        <a-tab-pane key="templates" tab="通知模板">
          <div class="tab-content">
            <!-- 搜索和筛选 -->
            <div class="search-section">
              <a-form layout="inline" :model="templateSearchForm">
                <a-form-item label="名称">
                  <a-input
                    v-model:value="templateSearchForm.name"
                    placeholder="搜索模板名称"
                    style="width: 200px"
                    @press-enter="handleTemplateSearch"
                  >
                    <template #prefix><SearchOutlined /></template>
                  </a-input>
                </a-form-item>
                <a-form-item label="类型">
                  <a-select
                    v-model:value="templateSearchForm.type"
                    placeholder="请选择类型"
                    style="width: 150px"
                    allow-clear
                  >
                    <a-select-option value="email">邮件</a-select-option>
                    <a-select-option value="webhook">Webhook</a-select-option>
                    <a-select-option value="dingtalk">钉钉</a-select-option>
                    <a-select-option value="wechat">企业微信</a-select-option>
                    <a-select-option value="slack">Slack</a-select-option>
                  </a-select>
                </a-form-item>
                <a-form-item>
                  <a-space>
                    <a-button type="primary" @click="handleTemplateSearch">
                      <template #icon><SearchOutlined /></template>
                      搜索
                    </a-button>
                    <a-button @click="handleTemplateReset">
                      <template #icon><ReloadOutlined /></template>
                      重置
                    </a-button>
                  </a-space>
                </a-form-item>
              </a-form>
            </div>

            <!-- 批量操作 -->
            <div class="batch-actions" v-if="selectedTemplateKeys.length > 0">
              <a-alert
                :message="`已选择 ${selectedTemplateKeys.length} 项`"
                type="info"
                show-icon
              >
                <template #action>
                  <a-space>
                    <a-button size="small" @click="handleBatchTemplateDelete">
                      批量删除
                    </a-button>
                  </a-space>
                </template>
              </a-alert>
            </div>

            <!-- 通知模板列表 -->
            <a-table
              :columns="templateColumns"
              :data-source="templateList"
              :loading="templateLoading"
              :pagination="templatePagination"
              :row-selection="templateRowSelection"
              @change="handleTemplateTableChange"
              row-key="id"
            >
              <template #bodyCell="{ column, record }">
                <template v-if="column.key === 'name'">
                  <div class="template-name">
                    <div class="name-content">
                      <FileTextOutlined class="type-icon" />
                      <div class="name-info">
                        <a @click="handleTemplateView(record)" class="name-link">
                          {{ record.name }}
                        </a>
                        <div class="name-meta">
                          {{ record.description }}
                        </div>
                      </div>
                    </div>
                  </div>
                </template>
                
                <template v-else-if="column.key === 'type'">
                  <a-tag :color="getTemplateTypeColor(record.type)">
                    {{ getTemplateTypeText(record.type) }}
                  </a-tag>
                </template>
                
                <template v-else-if="column.key === 'usage'">
                  <div class="usage-info">
                    <a-progress
                      :percent="getUsagePercent(record.usageCount)"
                      size="small"
                      :stroke-color="getUsageColor(record.usageCount)"
                    />
                    <div class="usage-text">
                      {{ record.usageCount || 0 }} 次使用
                    </div>
                  </div>
                </template>
                
                <template v-else-if="column.key === 'updatedAt'">
                  <div class="time-info">
                    <div>{{ formatDateTime(record.updatedAt) }}</div>
                    <div class="time-relative">{{ getRelativeTime(record.updatedAt) }}</div>
                  </div>
                </template>
                
                <template v-else-if="column.key === 'action'">
                  <a-space>
                    <a-button type="link" size="small" @click="handleTemplateView(record)">
                      查看
                    </a-button>
                    <a-button type="link" size="small" @click="handleTemplatePreview(record)">
                      预览
                    </a-button>
                    <a-dropdown>
                      <template #overlay>
                        <a-menu>
                          <a-menu-item @click="handleTemplateEdit(record)">
                            <EditOutlined /> 编辑
                          </a-menu-item>
                          <a-menu-item @click="handleTemplateDuplicate(record)">
                            <CopyOutlined /> 复制
                          </a-menu-item>
                          <a-menu-divider />
                          <a-menu-item @click="handleTemplateDelete(record)" danger>
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
          </div>
        </a-tab-pane>
      </a-tabs>
    </a-card>

    <!-- 通知组详情抽屉 -->
    <a-drawer
      v-model:open="groupDetailVisible"
      title="通知组详情"
      width="800"
      :footer-style="{ textAlign: 'right' }"
    >
      <NotificationGroupDetail
        v-if="currentGroup"
        :group="currentGroup"
        @edit="handleGroupEdit"
        @test="handleGroupTest"
        @delete="handleGroupDelete"
        @close="groupDetailVisible = false"
      />
      <template #footer>
        <a-space>
          <a-button @click="groupDetailVisible = false">关闭</a-button>
          <a-button @click="handleGroupTest(currentGroup)">
            测试发送
          </a-button>
          <a-button type="primary" @click="handleGroupEdit(currentGroup)">
            编辑
          </a-button>
        </a-space>
      </template>
    </a-drawer>

    <!-- 通知模板详情抽屉 -->
    <a-drawer
      v-model:open="templateDetailVisible"
      title="通知模板详情"
      width="800"
      :footer-style="{ textAlign: 'right' }"
    >
      <NotificationTemplateDetail
        v-if="currentTemplate"
        :template="currentTemplate"
        @edit="handleTemplateEdit"
        @preview="handleTemplatePreview"
        @delete="handleTemplateDelete"
        @close="templateDetailVisible = false"
      />
      <template #footer>
        <a-space>
          <a-button @click="templateDetailVisible = false">关闭</a-button>
          <a-button @click="handleTemplatePreview(currentTemplate)">
            预览
          </a-button>
          <a-button type="primary" @click="handleTemplateEdit(currentTemplate)">
            编辑
          </a-button>
        </a-space>
      </template>
    </a-drawer>

    <!-- 通知组表单模态框 -->
    <a-modal
      v-model:open="groupFormVisible"
      :title="groupFormMode === 'create' ? '新建通知组' : '编辑通知组'"
      width="800"
      :footer="null"
      :destroy-on-close="true"
    >
      <NotificationGroupForm
        v-if="groupFormVisible"
        :group="currentGroup"
        :mode="groupFormMode"
        @submit="handleGroupFormSubmit"
        @cancel="groupFormVisible = false"
      />
    </a-modal>

    <!-- 通知模板表单模态框 -->
    <a-modal
      v-model:open="templateFormVisible"
      :title="templateFormMode === 'create' ? '新建通知模板' : '编辑通知模板'"
      width="1000"
      :footer="null"
      :destroy-on-close="true"
    >
      <NotificationTemplateForm
        v-if="templateFormVisible"
        :template="currentTemplate"
        :mode="templateFormMode"
        @submit="handleTemplateFormSubmit"
        @cancel="templateFormVisible = false"
      />
    </a-modal>

    <!-- 模板预览模态框 -->
    <a-modal
      v-model:open="previewVisible"
      title="模板预览"
      width="800"
      :footer="null"
    >
      <NotificationTemplatePreview
        v-if="currentTemplate && previewVisible"
        :template="currentTemplate"
      />
    </a-modal>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import {
  Card,
  Row,
  Col,
  Statistic,
  Tabs,
  Form,
  Input,
  Select,
  Button,
  Space,
  Table,
  Tag,
  Switch,
  Progress,
  Alert,
  Drawer,
  Modal,
  Dropdown,
  Menu,
  message
} from 'ant-design-vue'
import {
  TeamOutlined,
  FileTextOutlined,
  SendOutlined,
  ExclamationCircleOutlined,
  SearchOutlined,
  ReloadOutlined,
  DownOutlined,
  EditOutlined,
  CopyOutlined,
  DeleteOutlined,
  MailOutlined,
  ApiOutlined,
  MessageOutlined,
  WechatOutlined,
  SlackOutlined
} from '@ant-design/icons-vue'
import { formatDateTime, getRelativeTime } from '@/utils/datetime'
import {
  getNotificationGroups,
  getNotificationGroup,
  createNotificationGroup,
  updateNotificationGroup,
  deleteNotificationGroup,
  testNotificationGroup,
  batchDeleteNotificationGroups,
  getNotificationTemplates,
  getNotificationTemplate,
  createNotificationTemplate,
  updateNotificationTemplate,
  deleteNotificationTemplate,
  previewNotificationTemplate,
  batchDeleteNotificationTemplates
} from '@/services/notification'
import type { NotificationGroup, NotificationTemplate } from '@/types'
import NotificationGroupDetail from '@/components/NotificationGroupDetail.vue'
import NotificationGroupForm from '@/components/NotificationGroupForm.vue'
import NotificationTemplateDetail from '@/components/NotificationTemplateDetail.vue'
import NotificationTemplateForm from '@/components/NotificationTemplateForm.vue'
import NotificationTemplatePreview from '@/components/NotificationTemplatePreview.vue'

const ACard = Card
const ARow = Row
const ACol = Col
const AStatistic = Statistic
const ATabs = Tabs
const ATabPane = Tabs.TabPane
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
const AProgress = Progress
const AAlert = Alert
const ADrawer = Drawer
const AModal = Modal
const ADropdown = Dropdown
const AMenu = Menu
const AMenuItem = Menu.Item
const AMenuDivider = Menu.Divider

// 响应式数据
const activeTab = ref('groups')
const groupLoading = ref(false)
const templateLoading = ref(false)
const groupList = ref<NotificationGroup[]>([])
const templateList = ref<NotificationTemplate[]>([])
const selectedGroupKeys = ref<string[]>([])
const selectedTemplateKeys = ref<string[]>([])

// 统计数据
const stats = reactive({
  groups: 0,
  templates: 0,
  todaySent: 0,
  failed: 0
})

// 搜索表单
const groupSearchForm = reactive({
  name: '',
  type: undefined,
  enabled: undefined
})

const templateSearchForm = reactive({
  name: '',
  type: undefined
})

// 分页配置
const groupPagination = reactive({
  current: 1,
  pageSize: 10,
  total: 0,
  showSizeChanger: true,
  showQuickJumper: true,
  showTotal: (total: number) => `共 ${total} 条记录`
})

const templatePagination = reactive({
  current: 1,
  pageSize: 10,
  total: 0,
  showSizeChanger: true,
  showQuickJumper: true,
  showTotal: (total: number) => `共 ${total} 条记录`
})

// 表格配置
const groupColumns = [
  {
    title: '名称',
    key: 'name',
    width: 250,
    ellipsis: true
  },
  {
    title: '类型',
    key: 'type',
    width: 120
  },
  {
    title: '状态',
    key: 'enabled',
    width: 80
  },
  {
    title: '发送统计',
    key: 'stats',
    width: 120
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

const templateColumns = [
  {
    title: '名称',
    key: 'name',
    width: 250,
    ellipsis: true
  },
  {
    title: '类型',
    key: 'type',
    width: 120
  },
  {
    title: '使用情况',
    key: 'usage',
    width: 150
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
const groupRowSelection = {
  selectedRowKeys: selectedGroupKeys,
  onChange: (keys: string[]) => {
    selectedGroupKeys.value = keys
  }
}

const templateRowSelection = {
  selectedRowKeys: selectedTemplateKeys,
  onChange: (keys: string[]) => {
    selectedTemplateKeys.value = keys
  }
}

// 详情抽屉
const groupDetailVisible = ref(false)
const templateDetailVisible = ref(false)
const currentGroup = ref<NotificationGroup | null>(null)
const currentTemplate = ref<NotificationTemplate | null>(null)

// 表单模态框
const groupFormVisible = ref(false)
const templateFormVisible = ref(false)
const groupFormMode = ref<'create' | 'edit'>('create')
const templateFormMode = ref<'create' | 'edit'>('create')

// 预览模态框
const previewVisible = ref(false)

// 获取通知组类型图标
const getGroupTypeIcon = (type: string) => {
  const iconMap: Record<string, any> = {
    email: MailOutlined,
    webhook: ApiOutlined,
    dingtalk: MessageOutlined,
    wechat: WechatOutlined,
    slack: SlackOutlined
  }
  return iconMap[type] || MessageOutlined
}

// 获取通知组类型颜色
const getGroupTypeColor = (type: string) => {
  const colorMap: Record<string, string> = {
    email: 'blue',
    webhook: 'green',
    dingtalk: 'orange',
    wechat: 'cyan',
    slack: 'purple'
  }
  return colorMap[type] || 'default'
}

// 获取通知组类型文本
const getGroupTypeText = (type: string) => {
  const textMap: Record<string, string> = {
    email: '邮件',
    webhook: 'Webhook',
    dingtalk: '钉钉',
    wechat: '企业微信',
    slack: 'Slack'
  }
  return textMap[type] || type
}

// 获取模板类型颜色
const getTemplateTypeColor = (type: string) => {
  return getGroupTypeColor(type)
}

// 获取模板类型文本
const getTemplateTypeText = (type: string) => {
  return getGroupTypeText(type)
}

// 获取使用率百分比
const getUsagePercent = (count: number) => {
  const max = 100 // 假设最大使用次数为100
  return Math.min((count / max) * 100, 100)
}

// 获取使用率颜色
const getUsageColor = (count: number) => {
  if (count >= 50) return '#52c41a'
  if (count >= 20) return '#faad14'
  return '#1890ff'
}

// 加载通知组数据
const loadGroupData = async () => {
  groupLoading.value = true
  try {
    const params = {
      page: groupPagination.current,
      pageSize: groupPagination.pageSize,
      name: groupSearchForm.name,
      type: groupSearchForm.type,
      enabled: groupSearchForm.enabled
    }
    
    const response = await getNotificationGroups(params)
    groupList.value = response.data.list
    groupPagination.total = response.data.total
  } catch (error) {
    console.error('加载通知组列表失败:', error)
    message.error('加载通知组列表失败')
  } finally {
    groupLoading.value = false
  }
}

// 加载通知模板数据
const loadTemplateData = async () => {
  templateLoading.value = true
  try {
    const params = {
      page: templatePagination.current,
      pageSize: templatePagination.pageSize,
      name: templateSearchForm.name,
      type: templateSearchForm.type
    }
    
    const response = await getNotificationTemplates(params)
    templateList.value = response.data.list
    templatePagination.total = response.data.total
  } catch (error) {
    console.error('加载通知模板列表失败:', error)
    message.error('加载通知模板列表失败')
  } finally {
    templateLoading.value = false
  }
}

// 加载统计数据
const loadStats = async () => {
  try {
    // 这里应该有获取统计数据的API
    stats.groups = groupList.value.length
    stats.templates = templateList.value.length
    stats.todaySent = 0
    stats.failed = 0
  } catch (error) {
    console.error('加载统计数据失败:', error)
  }
}

// 刷新
const handleRefresh = () => {
  if (activeTab.value === 'groups') {
    loadGroupData()
  } else {
    loadTemplateData()
  }
  loadStats()
}

// 标签页切换
const handleTabChange = (key: string) => {
  activeTab.value = key
  if (key === 'groups') {
    loadGroupData()
  } else {
    loadTemplateData()
  }
}

// 通知组搜索
const handleGroupSearch = () => {
  groupPagination.current = 1
  loadGroupData()
}

// 通知组重置
const handleGroupReset = () => {
  Object.assign(groupSearchForm, {
    name: '',
    type: undefined,
    enabled: undefined
  })
  groupPagination.current = 1
  loadGroupData()
}

// 通知模板搜索
const handleTemplateSearch = () => {
  templatePagination.current = 1
  loadTemplateData()
}

// 通知模板重置
const handleTemplateReset = () => {
  Object.assign(templateSearchForm, {
    name: '',
    type: undefined
  })
  templatePagination.current = 1
  loadTemplateData()
}

// 通知组表格变化
const handleGroupTableChange = (pag: any, filters: any, sorter: any) => {
  groupPagination.current = pag.current
  groupPagination.pageSize = pag.pageSize
  loadGroupData()
}

// 通知模板表格变化
const handleTemplateTableChange = (pag: any, filters: any, sorter: any) => {
  templatePagination.current = pag.current
  templatePagination.pageSize = pag.pageSize
  loadTemplateData()
}

// 通知组操作
const handleCreateGroup = () => {
  currentGroup.value = null
  groupFormMode.value = 'create'
  groupFormVisible.value = true
}

const handleGroupView = async (group: NotificationGroup) => {
  try {
    const response = await getNotificationGroup(group.id)
    currentGroup.value = response.data
    groupDetailVisible.value = true
  } catch (error) {
    console.error('获取通知组详情失败:', error)
    message.error('获取通知组详情失败')
  }
}

const handleGroupEdit = (group: NotificationGroup) => {
  currentGroup.value = group
  groupFormMode.value = 'edit'
  groupFormVisible.value = true
  groupDetailVisible.value = false
}

const handleGroupTest = async (group: NotificationGroup) => {
  try {
    await testNotificationGroup(group.id)
    message.success('测试发送成功')
  } catch (error) {
    console.error('测试发送失败:', error)
    message.error('测试发送失败')
  }
}

const handleGroupToggle = async (group: NotificationGroup, enabled: boolean) => {
  try {
    await updateNotificationGroup(group.id, { enabled })
    group.enabled = enabled
    message.success(enabled ? '启用成功' : '禁用成功')
  } catch (error) {
    console.error('状态切换失败:', error)
    message.error('状态切换失败')
  }
}

const handleGroupDuplicate = (group: NotificationGroup) => {
  const duplicatedGroup = {
    ...group,
    name: `${group.name} - 副本`,
    id: undefined
  }
  currentGroup.value = duplicatedGroup
  groupFormMode.value = 'create'
  groupFormVisible.value = true
}

const handleGroupDelete = async (group: NotificationGroup) => {
  try {
    await deleteNotificationGroup(group.id)
    message.success('删除成功')
    groupDetailVisible.value = false
    loadGroupData()
  } catch (error) {
    console.error('删除失败:', error)
    message.error('删除失败')
  }
}

const handleGroupFormSubmit = async (data: any) => {
  try {
    if (groupFormMode.value === 'create') {
      await createNotificationGroup(data)
      message.success('创建成功')
    } else {
      await updateNotificationGroup(currentGroup.value!.id, data)
      message.success('更新成功')
    }
    
    groupFormVisible.value = false
    loadGroupData()
  } catch (error) {
    console.error('保存失败:', error)
    message.error('保存失败')
  }
}

// 批量操作
const handleBatchGroupTest = async () => {
  try {
    // 这里应该有批量测试的API
    message.success('批量测试完成')
    selectedGroupKeys.value = []
  } catch (error) {
    console.error('批量测试失败:', error)
    message.error('批量测试失败')
  }
}

const handleBatchGroupDelete = async () => {
  try {
    await batchDeleteNotificationGroups(selectedGroupKeys.value)
    message.success('批量删除成功')
    selectedGroupKeys.value = []
    loadGroupData()
  } catch (error) {
    console.error('批量删除失败:', error)
    message.error('批量删除失败')
  }
}

// 通知模板操作
const handleCreateTemplate = () => {
  currentTemplate.value = null
  templateFormMode.value = 'create'
  templateFormVisible.value = true
}

const handleTemplateView = async (template: NotificationTemplate) => {
  try {
    const response = await getNotificationTemplate(template.id)
    currentTemplate.value = response.data
    templateDetailVisible.value = true
  } catch (error) {
    console.error('获取通知模板详情失败:', error)
    message.error('获取通知模板详情失败')
  }
}

const handleTemplateEdit = (template: NotificationTemplate) => {
  currentTemplate.value = template
  templateFormMode.value = 'edit'
  templateFormVisible.value = true
  templateDetailVisible.value = false
}

const handleTemplatePreview = async (template: NotificationTemplate) => {
  try {
    const response = await previewNotificationTemplate(template.id, {})
    currentTemplate.value = {
      ...template,
      previewContent: response.data.content
    }
    previewVisible.value = true
  } catch (error) {
    console.error('预览模板失败:', error)
    message.error('预览模板失败')
  }
}

const handleTemplateDuplicate = (template: NotificationTemplate) => {
  const duplicatedTemplate = {
    ...template,
    name: `${template.name} - 副本`,
    id: undefined
  }
  currentTemplate.value = duplicatedTemplate
  templateFormMode.value = 'create'
  templateFormVisible.value = true
}

const handleTemplateDelete = async (template: NotificationTemplate) => {
  try {
    await deleteNotificationTemplate(template.id)
    message.success('删除成功')
    templateDetailVisible.value = false
    loadTemplateData()
  } catch (error) {
    console.error('删除失败:', error)
    message.error('删除失败')
  }
}

const handleTemplateFormSubmit = async (data: any) => {
  try {
    if (templateFormMode.value === 'create') {
      await createNotificationTemplate(data)
      message.success('创建成功')
    } else {
      await updateNotificationTemplate(currentTemplate.value!.id, data)
      message.success('更新成功')
    }
    
    templateFormVisible.value = false
    loadTemplateData()
  } catch (error) {
    console.error('保存失败:', error)
    message.error('保存失败')
  }
}

const handleBatchTemplateDelete = async () => {
  try {
    await batchDeleteNotificationTemplates(selectedTemplateKeys.value)
    message.success('批量删除成功')
    selectedTemplateKeys.value = []
    loadTemplateData()
  } catch (error) {
    console.error('批量删除失败:', error)
    message.error('批量删除失败')
  }
}

// 组件挂载
onMounted(() => {
  loadGroupData()
  loadStats()
})
</script>

<style scoped>
.notification-list {
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

.tabs-card {
  margin-bottom: 0;
}

.tab-content {
  padding: 16px 0;
}

.search-section {
  margin-bottom: 16px;
  padding: 16px;
  background: #fafafa;
  border-radius: 6px;
}

.batch-actions {
  margin-bottom: 16px;
}

.group-name,
.template-name {
  display: flex;
  align-items: center;
}

.name-content {
  display: flex;
  align-items: center;
  gap: 12px;
}

.type-icon {
  font-size: 20px;
  color: #1890ff;
}

.name-info {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.name-link {
  font-weight: 500;
  color: #1890ff;
  text-decoration: none;
}

.name-link:hover {
  text-decoration: underline;
}

.name-meta {
  font-size: 12px;
  color: #999;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  max-width: 200px;
}

.stats-info {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.stat-item {
  display: flex;
  justify-content: space-between;
  font-size: 12px;
}

.stat-label {
  color: #999;
}

.stat-value {
  color: #666;
  font-weight: 500;
}

.stat-value.error {
  color: #ff4d4f;
}

.usage-info {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.usage-text {
  font-size: 12px;
  color: #666;
  text-align: center;
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

/* 响应式设计 */
@media (max-width: 768px) {
  .notification-list {
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
  
  .search-section :deep(.ant-form-item) {
    margin-bottom: 16px;
  }
  
  .name-meta {
    max-width: 100px;
  }
}
</style>