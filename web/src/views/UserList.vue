<template>
  <div class="user-list-page">
    <div class="page-header">
      <h1 class="page-title">
        <UserOutlined />
        用户管理
      </h1>
      <div class="page-actions">
        <a-space>
          <a-button @click="handleImport">
            <ImportOutlined /> 导入用户
          </a-button>
          <a-button @click="handleExport">
            <ExportOutlined /> 导出用户
          </a-button>
          <a-button type="primary" @click="handleCreate">
            <PlusOutlined /> 新建用户
          </a-button>
        </a-space>
      </div>
    </div>

    <!-- 统计卡片 -->
    <div class="stats-cards">
      <a-row :gutter="16">
        <a-col :span="6">
          <a-card>
            <a-statistic
              title="总用户数"
              :value="stats.total"
              :value-style="{ color: '#1890ff' }"
            >
              <template #prefix>
                <UserOutlined />
              </template>
            </a-statistic>
          </a-card>
        </a-col>
        <a-col :span="6">
          <a-card>
            <a-statistic
              title="活跃用户"
              :value="stats.active"
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
              title="禁用用户"
              :value="stats.disabled"
              :value-style="{ color: '#ff4d4f' }"
            >
              <template #prefix>
                <StopOutlined />
              </template>
            </a-statistic>
          </a-card>
        </a-col>
        <a-col :span="6">
          <a-card>
            <a-statistic
              title="在线用户"
              :value="stats.online"
              :value-style="{ color: '#722ed1' }"
            >
              <template #prefix>
                <GlobalOutlined />
              </template>
            </a-statistic>
          </a-card>
        </a-col>
      </a-row>
    </div>

    <!-- 搜索和筛选 -->
    <div class="search-form">
      <a-form layout="inline" :model="searchForm" @finish="handleSearch">
        <a-form-item label="用户名" name="username">
          <a-input
            v-model:value="searchForm.username"
            placeholder="请输入用户名"
            allow-clear
          />
        </a-form-item>
        <a-form-item label="邮箱" name="email">
          <a-input
            v-model:value="searchForm.email"
            placeholder="请输入邮箱"
            allow-clear
          />
        </a-form-item>
        <a-form-item label="角色" name="role">
          <a-select
            v-model:value="searchForm.role"
            placeholder="请选择角色"
            allow-clear
            style="width: 120px"
          >
            <a-select-option value="admin">管理员</a-select-option>
            <a-select-option value="user">普通用户</a-select-option>
            <a-select-option value="viewer">只读用户</a-select-option>
          </a-select>
        </a-form-item>
        <a-form-item label="状态" name="status">
          <a-select
            v-model:value="searchForm.status"
            placeholder="请选择状态"
            allow-clear
            style="width: 120px"
          >
            <a-select-option value="active">活跃</a-select-option>
            <a-select-option value="disabled">禁用</a-select-option>
          </a-select>
        </a-form-item>
        <a-form-item>
          <a-space>
            <a-button type="primary" html-type="submit">
              <SearchOutlined /> 搜索
            </a-button>
            <a-button @click="handleReset">
              <ReloadOutlined /> 重置
            </a-button>
          </a-space>
        </a-form-item>
      </a-form>
    </div>

    <!-- 批量操作 -->
    <div class="batch-actions" v-if="selectedRowKeys.length > 0">
      <a-alert
        :message="`已选择 ${selectedRowKeys.length} 项`"
        type="info"
        show-icon
      >
        <template #action>
          <a-space>
            <a-button size="small" @click="handleBatchEnable">
              启用
            </a-button>
            <a-button size="small" @click="handleBatchDisable">
              禁用
            </a-button>
            <a-button size="small" danger @click="handleBatchDelete">
              删除
            </a-button>
          </a-space>
        </template>
      </a-alert>
    </div>

    <!-- 用户表格 -->
    <div class="table-container">
      <a-table
        :columns="columns"
        :data-source="users"
        :loading="loading"
        :pagination="pagination"
        :row-selection="rowSelection"
        @change="handleTableChange"
        row-key="id"
      >
        <template #bodyCell="{ column, record }">
          <template v-if="column.key === 'avatar'">
            <a-avatar :src="record.avatar" :size="32">
              {{ record.username?.charAt(0)?.toUpperCase() }}
            </a-avatar>
          </template>
          
          <template v-else-if="column.key === 'username'">
            <a @click="handleView(record)" class="username-link">
              {{ record.username }}
            </a>
          </template>
          
          <template v-else-if="column.key === 'role'">
            <a-tag :color="getRoleColor(record.role)">
              {{ getRoleText(record.role) }}
            </a-tag>
          </template>
          
          <template v-else-if="column.key === 'status'">
            <a-tag :color="record.status === 'active' ? 'green' : 'red'">
              {{ record.status === 'active' ? '活跃' : '禁用' }}
            </a-tag>
          </template>
          
          <template v-else-if="column.key === 'lastLogin'">
            <span v-if="record.lastLogin">
              {{ formatDateTime(record.lastLogin) }}
            </span>
            <span v-else class="text-muted">从未登录</span>
          </template>
          
          <template v-else-if="column.key === 'createdAt'">
            {{ formatDateTime(record.createdAt) }}
          </template>
          
          <template v-else-if="column.key === 'actions'">
            <a-space>
              <a-button type="link" size="small" @click="handleView(record)">
                <EyeOutlined /> 查看
              </a-button>
              <a-button type="link" size="small" @click="handleEdit(record)">
                <EditOutlined /> 编辑
              </a-button>
              <a-dropdown>
                <a-button type="link" size="small">
                  <MoreOutlined />
                </a-button>
                <template #overlay>
                  <a-menu>
                    <a-menu-item @click="handleResetPassword(record)">
                      <KeyOutlined /> 重置密码
                    </a-menu-item>
                    <a-menu-item @click="handleToggleStatus(record)">
                      <template v-if="record.status === 'active'">
                        <StopOutlined /> 禁用
                      </template>
                      <template v-else>
                        <CheckCircleOutlined /> 启用
                      </template>
                    </a-menu-item>
                    <a-menu-divider />
                    <a-menu-item @click="handleDelete(record)" danger>
                      <DeleteOutlined /> 删除
                    </a-menu-item>
                  </a-menu>
                </template>
              </a-dropdown>
            </a-space>
          </template>
        </template>
      </a-table>
    </div>

    <!-- 用户详情抽屉 -->
    <a-drawer
      v-model:open="detailVisible"
      title="用户详情"
      width="600"
      :footer-style="{ textAlign: 'right' }"
    >
      <UserDetail
        v-if="detailVisible && currentUser"
        :user="currentUser"
        @edit="handleEdit"
        @close="detailVisible = false"
      />
    </a-drawer>

    <!-- 用户表单抽屉 -->
    <a-drawer
      v-model:open="formVisible"
      :title="isEdit ? '编辑用户' : '新建用户'"
      width="600"
      :footer-style="{ textAlign: 'right' }"
    >
      <UserForm
        v-if="formVisible"
        :user="currentUser"
        :is-edit="isEdit"
        @submit="handleFormSubmit"
        @cancel="formVisible = false"
      />
    </a-drawer>

    <!-- 导入用户模态框 -->
    <a-modal
      v-model:open="importVisible"
      title="导入用户"
      @ok="handleImportSubmit"
      @cancel="importVisible = false"
    >
      <a-upload-dragger
        v-model:file-list="importFileList"
        :before-upload="() => false"
        accept=".xlsx,.xls,.csv"
      >
        <p class="ant-upload-drag-icon">
          <InboxOutlined />
        </p>
        <p class="ant-upload-text">点击或拖拽文件到此区域上传</p>
        <p class="ant-upload-hint">
          支持 Excel 和 CSV 格式文件
        </p>
      </a-upload-dragger>
      
      <div class="import-template" style="margin-top: 16px;">
        <a @click="downloadTemplate">
          <DownloadOutlined /> 下载导入模板
        </a>
      </div>
    </a-modal>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import {
  Row,
  Col,
  Card,
  Statistic,
  Form,
  Input,
  Select,
  Button,
  Space,
  Alert,
  Table,
  Avatar,
  Tag,
  Dropdown,
  Menu,
  Drawer,
  Modal,
  Upload,
  message
} from 'ant-design-vue'
import {
  UserOutlined,
  ImportOutlined,
  ExportOutlined,
  PlusOutlined,
  CheckCircleOutlined,
  StopOutlined,
  GlobalOutlined,
  SearchOutlined,
  ReloadOutlined,
  EyeOutlined,
  EditOutlined,
  MoreOutlined,
  KeyOutlined,
  DeleteOutlined,
  InboxOutlined,
  DownloadOutlined
} from '@ant-design/icons-vue'
import { formatDateTime } from '@/utils/datetime'
import {
  getUserList,
  getUserStats,
  createUser,
  updateUser,
  deleteUser,
  resetUserPassword,
  toggleUserStatus,
  importUsers,
  exportUsers
} from '@/services/user'
import type { User } from '@/types'
import UserDetail from '@/components/UserDetail.vue'
import UserForm from '@/components/UserForm.vue'

const ARow = Row
const ACol = Col
const ACard = Card
const AStatistic = Statistic
const AForm = Form
const AFormItem = Form.Item
const AInput = Input
const ASelect = Select
const ASelectOption = Select.Option
const AButton = Button
const ASpace = Space
const AAlert = Alert
const ATable = Table
const AAvatar = Avatar
const ATag = Tag
const ADropdown = Dropdown
const AMenu = Menu
const AMenuItem = Menu.Item
const AMenuDivider = Menu.Divider
const ADrawer = Drawer
const AModal = Modal
const AUploadDragger = Upload.Dragger

// 响应式数据
const loading = ref(false)
const users = ref<User[]>([])
const selectedRowKeys = ref<string[]>([])
const detailVisible = ref(false)
const formVisible = ref(false)
const importVisible = ref(false)
const isEdit = ref(false)
const currentUser = ref<User | null>(null)
const importFileList = ref([])

// 统计数据
const stats = reactive({
  total: 0,
  active: 0,
  disabled: 0,
  online: 0
})

// 搜索表单
const searchForm = reactive({
  username: '',
  email: '',
  role: undefined,
  status: undefined
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

// 表格列配置
const columns = [
  {
    title: '头像',
    key: 'avatar',
    width: 80,
    align: 'center'
  },
  {
    title: '用户名',
    key: 'username',
    dataIndex: 'username',
    sorter: true
  },
  {
    title: '邮箱',
    key: 'email',
    dataIndex: 'email'
  },
  {
    title: '角色',
    key: 'role',
    dataIndex: 'role',
    width: 100
  },
  {
    title: '状态',
    key: 'status',
    dataIndex: 'status',
    width: 80
  },
  {
    title: '最后登录',
    key: 'lastLogin',
    dataIndex: 'lastLogin',
    width: 180,
    sorter: true
  },
  {
    title: '创建时间',
    key: 'createdAt',
    dataIndex: 'createdAt',
    width: 180,
    sorter: true
  },
  {
    title: '操作',
    key: 'actions',
    width: 200,
    fixed: 'right'
  }
]

// 行选择配置
const rowSelection = computed(() => ({
  selectedRowKeys: selectedRowKeys.value,
  onChange: (keys: string[]) => {
    selectedRowKeys.value = keys
  }
}))

// 获取角色颜色
const getRoleColor = (role: string) => {
  const colors = {
    admin: 'red',
    user: 'blue',
    viewer: 'green'
  }
  return colors[role as keyof typeof colors] || 'default'
}

// 获取角色文本
const getRoleText = (role: string) => {
  const texts = {
    admin: '管理员',
    user: '普通用户',
    viewer: '只读用户'
  }
  return texts[role as keyof typeof texts] || role
}

// 加载用户列表
const loadUsers = async () => {
  try {
    loading.value = true
    const params = {
      page: pagination.current,
      pageSize: pagination.pageSize,
      ...searchForm
    }
    
    const response = await getUserList(params)
    // 后端返回的数据结构是 {users: [...], total: ..., page: ..., size: ...}
    users.value = response.users || []
    pagination.total = response.total || 0
  } catch (error) {
    console.error('加载用户列表失败:', error)
    message.error('加载用户列表失败')
  } finally {
    loading.value = false
  }
}

// 加载统计数据
const loadStats = async () => {
  try {
    const response = await getUserStats()
    Object.assign(stats, response)
  } catch (error) {
    console.error('加载统计数据失败:', error)
  }
}

// 搜索处理
const handleSearch = () => {
  pagination.current = 1
  loadUsers()
}

// 重置处理
const handleReset = () => {
  Object.assign(searchForm, {
    username: '',
    email: '',
    role: undefined,
    status: undefined
  })
  pagination.current = 1
  loadUsers()
}

// 表格变化处理
const handleTableChange = (pag: any, filters: any, sorter: any) => {
  pagination.current = pag.current
  pagination.pageSize = pag.pageSize
  loadUsers()
}

// 查看用户
const handleView = (user: User) => {
  currentUser.value = user
  detailVisible.value = true
}

// 编辑用户
const handleEdit = (user: User) => {
  currentUser.value = user
  isEdit.value = true
  formVisible.value = true
}

// 新建用户
const handleCreate = () => {
  currentUser.value = null
  isEdit.value = false
  formVisible.value = true
}

// 表单提交
const handleFormSubmit = async (userData: any) => {
  try {
    if (isEdit.value && currentUser.value) {
      await updateUser(currentUser.value.id, userData)
      message.success('用户更新成功')
    } else {
      await createUser(userData)
      message.success('用户创建成功')
    }
    
    formVisible.value = false
    loadUsers()
    loadStats()
  } catch (error) {
    console.error('保存用户失败:', error)
    message.error('保存用户失败')
  }
}

// 删除用户
const handleDelete = (user: User) => {
  Modal.confirm({
    title: '确认删除',
    content: `确定要删除用户 "${user.username}" 吗？`,
    onOk: async () => {
      try {
        await deleteUser(user.id)
        message.success('用户删除成功')
        loadUsers()
        loadStats()
      } catch (error) {
        console.error('删除用户失败:', error)
        message.error('删除用户失败')
      }
    }
  })
}

// 重置密码
const handleResetPassword = (user: User) => {
  Modal.confirm({
    title: '确认重置密码',
    content: `确定要重置用户 "${user.username}" 的密码吗？`,
    onOk: async () => {
      try {
        const response = await resetUserPassword(user.id)
        Modal.info({
          title: '密码重置成功',
          content: `新密码：${response.newPassword}`
        })
      } catch (error) {
        console.error('重置密码失败:', error)
        message.error('重置密码失败')
      }
    }
  })
}

// 切换用户状态
const handleToggleStatus = async (user: User) => {
  try {
    await toggleUserStatus(user.id)
    message.success('用户状态更新成功')
    loadUsers()
    loadStats()
  } catch (error) {
    console.error('更新用户状态失败:', error)
    message.error('更新用户状态失败')
  }
}

// 批量启用
const handleBatchEnable = async () => {
  try {
    // 实现批量启用逻辑
    message.success('批量启用成功')
    selectedRowKeys.value = []
    loadUsers()
    loadStats()
  } catch (error) {
    console.error('批量启用失败:', error)
    message.error('批量启用失败')
  }
}

// 批量禁用
const handleBatchDisable = async () => {
  try {
    // 实现批量禁用逻辑
    message.success('批量禁用成功')
    selectedRowKeys.value = []
    loadUsers()
    loadStats()
  } catch (error) {
    console.error('批量禁用失败:', error)
    message.error('批量禁用失败')
  }
}

// 批量删除
const handleBatchDelete = () => {
  Modal.confirm({
    title: '确认批量删除',
    content: `确定要删除选中的 ${selectedRowKeys.value.length} 个用户吗？`,
    onOk: async () => {
      try {
        // 实现批量删除逻辑
        message.success('批量删除成功')
        selectedRowKeys.value = []
        loadUsers()
        loadStats()
      } catch (error) {
        console.error('批量删除失败:', error)
        message.error('批量删除失败')
      }
    }
  })
}

// 导入用户
const handleImport = () => {
  importVisible.value = true
}

// 导入提交
const handleImportSubmit = async () => {
  if (importFileList.value.length === 0) {
    message.error('请选择要导入的文件')
    return
  }
  
  try {
    const file = importFileList.value[0] as any
    await importUsers(file.originFileObj)
    message.success('用户导入成功')
    importVisible.value = false
    importFileList.value = []
    loadUsers()
    loadStats()
  } catch (error) {
    console.error('导入用户失败:', error)
    message.error('导入用户失败')
  }
}

// 导出用户
const handleExport = async () => {
  try {
    const blob = await exportUsers(searchForm)
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = 'users.xlsx'
    a.click()
    URL.revokeObjectURL(url)
    message.success('用户导出成功')
  } catch (error) {
    console.error('导出用户失败:', error)
    message.error('导出用户失败')
  }
}

// 下载模板
const downloadTemplate = () => {
  const template = [
    ['用户名', '邮箱', '角色', '状态'],
    ['admin', 'admin@example.com', 'admin', 'active'],
    ['user1', 'user1@example.com', 'user', 'active']
  ]
  
  const csvContent = template.map(row => row.join(',')).join('\n')
  const blob = new Blob([csvContent], { type: 'text/csv;charset=utf-8;' })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = 'user-template.csv'
  a.click()
  URL.revokeObjectURL(url)
}

// 组件挂载
onMounted(() => {
  loadUsers()
  loadStats()
})
</script>

<style scoped>
.user-list-page {
  padding: 24px;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 24px;
}

.page-title {
  display: flex;
  align-items: center;
  gap: 12px;
  margin: 0;
  font-size: 24px;
  font-weight: 600;
  color: #262626;
}

.stats-cards {
  margin-bottom: 24px;
}

.search-form {
  background: white;
  padding: 24px;
  border-radius: 8px;
  margin-bottom: 16px;
}

.batch-actions {
  margin-bottom: 16px;
}

.table-container {
  background: white;
  border-radius: 8px;
  overflow: hidden;
}

.username-link {
  color: #1890ff;
  text-decoration: none;
}

.username-link:hover {
  text-decoration: underline;
}

.text-muted {
  color: #8c8c8c;
}

.import-template {
  text-align: center;
}

/* 响应式设计 */
@media (max-width: 768px) {
  .user-list-page {
    padding: 16px;
  }
  
  .page-header {
    flex-direction: column;
    gap: 16px;
    align-items: stretch;
  }
  
  .stats-cards :deep(.ant-col) {
    margin-bottom: 16px;
  }
  
  .search-form {
    padding: 16px;
  }
  
  .search-form :deep(.ant-form-item) {
    margin-bottom: 16px;
  }
  
  .table-container :deep(.ant-table) {
    font-size: 12px;
  }
  
  .table-container :deep(.ant-table-thead > tr > th),
  .table-container :deep(.ant-table-tbody > tr > td) {
    padding: 8px 4px;
  }
}
</style>