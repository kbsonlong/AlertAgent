<template>
  <div class="permission-list">
    <!-- 页面标题 -->
    <div class="page-header">
      <h1>权限管理</h1>
      <p>管理系统权限，配置角色权限</p>
    </div>

    <!-- 操作栏 -->
    <div class="action-bar">
      <a-space>
        <a-button type="primary" @click="handleCreate">
          <template #icon><PlusOutlined /></template>
          新建权限
        </a-button>
        <a-button @click="handleExport">
          <template #icon><ExportOutlined /></template>
          导出权限
        </a-button>
        <a-button @click="fetchPermissionList">
          <template #icon><ReloadOutlined /></template>
          刷新
        </a-button>
      </a-space>
    </div>

    <!-- 统计卡片 -->
    <div class="stats-cards">
      <a-row :gutter="16">
        <a-col :span="6">
          <a-card>
            <a-statistic title="总权限数" :value="stats.total" />
          </a-card>
        </a-col>
        <a-col :span="6">
          <a-card>
            <a-statistic title="系统权限" :value="stats.system" />
          </a-card>
        </a-col>
        <a-col :span="6">
          <a-card>
            <a-statistic title="自定义权限" :value="stats.custom" />
          </a-card>
        </a-col>
        <a-col :span="6">
          <a-card>
            <a-statistic title="启用权限" :value="stats.active" />
          </a-card>
        </a-col>
      </a-row>
    </div>

    <!-- 搜索和筛选 -->
    <a-card class="search-card">
      <a-form layout="inline" :model="searchForm">
        <a-form-item label="权限名称">
          <a-input v-model:value="searchForm.name" placeholder="请输入权限名称" allowClear />
        </a-form-item>
        <a-form-item label="权限代码">
          <a-input v-model:value="searchForm.code" placeholder="请输入权限代码" allowClear />
        </a-form-item>
        <a-form-item label="权限类型">
          <a-select v-model:value="searchForm.type" placeholder="请选择权限类型" allowClear style="width: 120px">
            <a-select-option value="system">系统权限</a-select-option>
            <a-select-option value="custom">自定义权限</a-select-option>
          </a-select>
        </a-form-item>
        <a-form-item label="状态">
          <a-select v-model:value="searchForm.status" placeholder="请选择状态" allowClear style="width: 100px">
            <a-select-option value="active">启用</a-select-option>
            <a-select-option value="inactive">禁用</a-select-option>
          </a-select>
        </a-form-item>
        <a-form-item>
          <a-space>
            <a-button type="primary" @click="handleSearch">
              <template #icon><SearchOutlined /></template>
              搜索
            </a-button>
            <a-button @click="handleReset">
              重置
            </a-button>
          </a-space>
        </a-form-item>
      </a-form>
    </a-card>

    <!-- 批量操作 -->
    <div v-if="selectedRowKeys.length > 0" class="batch-actions">
      <a-alert
        :message="`已选择 ${selectedRowKeys.length} 项`"
        type="info"
        show-icon
        closable
        @close="selectedRowKeys = []"
      >
        <template #action>
          <a-space>
            <a-button size="small" @click="handleBatchEnable">批量启用</a-button>
            <a-button size="small" @click="handleBatchDisable">批量禁用</a-button>
            <a-button size="small" danger @click="handleBatchDelete">批量删除</a-button>
          </a-space>
        </template>
      </a-alert>
    </div>

    <!-- 权限表格 -->
    <a-card>
      <a-table
        :columns="columns"
        :data-source="filteredPermissionList"
        :loading="loading"
        :row-selection="{ selectedRowKeys, onChange: onSelectChange }"
        :pagination="{
          current: pagination.current,
          pageSize: pagination.pageSize,
          total: pagination.total,
          showSizeChanger: true,
          showQuickJumper: true,
          showTotal: (total, range) => `第 ${range[0]}-${range[1]} 条/共 ${total} 条`,
          onChange: handleTableChange,
          onShowSizeChange: handleTableChange
        }"
        row-key="id"
      >
        <template #bodyCell="{ column, record }">
          <template v-if="column.key === 'type'">
            <a-tag :color="record.type === 'system' ? 'blue' : 'green'">
              {{ record.type === 'system' ? '系统权限' : '自定义权限' }}
            </a-tag>
          </template>
          <template v-else-if="column.key === 'status'">
            <a-tag :color="record.status === 'active' ? 'success' : 'default'">
              {{ record.status === 'active' ? '启用' : '禁用' }}
            </a-tag>
          </template>
          <template v-else-if="column.key === 'actions'">
            <a-space>
              <a-button type="link" size="small" @click="handleEdit(record)">编辑</a-button>
              <a-button 
                type="link" 
                size="small" 
                :disabled="record.type === 'system'"
                danger 
                @click="handleDelete(record)"
              >
                删除
              </a-button>
              <a-dropdown>
                <template #overlay>
                  <a-menu>
                    <a-menu-item key="toggle" @click="handleToggleStatus(record)">
                      {{ record.status === 'active' ? '禁用' : '启用' }}
                    </a-menu-item>
                    <a-menu-item key="roles" @click="handleViewRoles(record)">
                      查看关联角色
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

    <!-- 权限编辑弹窗 -->
    <a-modal
      v-model:open="modalVisible"
      :title="isEdit ? '编辑权限' : '新建权限'"
      :confirm-loading="submitLoading"
      @ok="handleSubmit"
      @cancel="handleCancel"
      width="600px"
    >
      <a-form
        ref="formRef"
        :model="formData"
        :rules="formRules"
        :label-col="{ span: 6 }"
        :wrapper-col="{ span: 18 }"
      >
        <a-form-item label="权限名称" name="name">
          <a-input v-model:value="formData.name" placeholder="请输入权限名称" />
        </a-form-item>
        <a-form-item label="权限代码" name="code">
          <a-input v-model:value="formData.code" placeholder="请输入权限代码" />
        </a-form-item>
        <a-form-item label="权限描述" name="description">
          <a-textarea v-model:value="formData.description" placeholder="请输入权限描述" :rows="3" />
        </a-form-item>
        <a-form-item label="权限类型" name="type">
          <a-select v-model:value="formData.type" placeholder="请选择权限类型">
            <a-select-option value="system">系统权限</a-select-option>
            <a-select-option value="custom">自定义权限</a-select-option>
          </a-select>
        </a-form-item>
        <a-form-item label="状态" name="status">
          <a-radio-group v-model:value="formData.status">
            <a-radio value="active">启用</a-radio>
            <a-radio value="inactive">禁用</a-radio>
          </a-radio-group>
        </a-form-item>
        <a-form-item label="资源路径" name="resource">
          <a-input v-model:value="formData.resource" placeholder="请输入资源路径，如：/api/users" />
        </a-form-item>
        <a-form-item label="操作类型" name="action">
          <a-select v-model:value="formData.action" placeholder="请选择操作类型">
            <a-select-option value="create">创建</a-select-option>
            <a-select-option value="read">读取</a-select-option>
            <a-select-option value="update">更新</a-select-option>
            <a-select-option value="delete">删除</a-select-option>
            <a-select-option value="*">全部</a-select-option>
          </a-select>
        </a-form-item>
      </a-form>
    </a-modal>

    <!-- 关联角色弹窗 -->
    <a-modal
      v-model:open="rolesModalVisible"
      title="关联角色"
      :footer="null"
      width="800px"
    >
      <div v-if="selectedPermission">
        <h4>权限：{{ selectedPermission.name }}</h4>
        <a-table
          :columns="roleColumns"
          :data-source="permissionRoles"
          :loading="roleLoading"
          :pagination="false"
          row-key="id"
        >
          <template #bodyCell="{ column, record }">
            <template v-if="column.key === 'type'">
              <a-tag :color="record.type === 'system' ? 'blue' : 'green'">
                {{ record.type === 'system' ? '系统角色' : '自定义角色' }}
              </a-tag>
            </template>
            <template v-else-if="column.key === 'status'">
              <a-tag :color="record.status === 'active' ? 'success' : 'default'">
                {{ record.status === 'active' ? '启用' : '禁用' }}
              </a-tag>
            </template>
          </template>
        </a-table>
      </div>
    </a-modal>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { message, Modal } from 'ant-design-vue'
import {
  PlusOutlined,
  ExportOutlined,
  ReloadOutlined,
  SearchOutlined,
  DownOutlined
} from '@ant-design/icons-vue'
import { UserService } from '@/services/userService'
import type { Permission, Role } from '@/services/userService'

// 响应式数据
const loading = ref(false)
const submitLoading = ref(false)
const roleLoading = ref(false)
const modalVisible = ref(false)
const rolesModalVisible = ref(false)
const isEdit = ref(false)
const selectedRowKeys = ref<number[]>([])
const formRef = ref()

// 权限列表
const permissionList = ref<Permission[]>([])
const selectedPermission = ref<Permission | null>(null)
const permissionRoles = ref<Role[]>([])

// 统计数据
const stats = reactive({
  total: 0,
  system: 0,
  custom: 0,
  active: 0
})

// 搜索表单
const searchForm = reactive({
  name: '',
  code: '',
  type: '',
  status: ''
})

// 分页配置
const pagination = reactive({
  current: 1,
  pageSize: 10,
  total: 0
})

// 表单数据
const formData = reactive({
  name: '',
  code: '',
  description: '',
  type: 'custom',
  status: 'active',
  resource: '',
  action: ''
})

// 表单验证规则
const formRules = {
  name: [{ required: true, message: '请输入权限名称', trigger: 'blur' }],
  code: [{ required: true, message: '请输入权限代码', trigger: 'blur' }],
  type: [{ required: true, message: '请选择权限类型', trigger: 'change' }],
  status: [{ required: true, message: '请选择状态', trigger: 'change' }],
  resource: [{ required: true, message: '请输入资源路径', trigger: 'blur' }],
  action: [{ required: true, message: '请选择操作类型', trigger: 'change' }]
}

// 表格列配置
const columns = [
  {
    title: '权限名称',
    dataIndex: 'name',
    key: 'name',
    sorter: true
  },
  {
    title: '权限代码',
    dataIndex: 'code',
    key: 'code'
  },
  {
    title: '描述',
    dataIndex: 'description',
    key: 'description',
    ellipsis: true
  },
  {
    title: '类型',
    dataIndex: 'type',
    key: 'type',
    filters: [
      { text: '系统权限', value: 'system' },
      { text: '自定义权限', value: 'custom' }
    ]
  },
  {
    title: '状态',
    dataIndex: 'status',
    key: 'status',
    filters: [
      { text: '启用', value: 'active' },
      { text: '禁用', value: 'inactive' }
    ]
  },
  {
    title: '资源路径',
    dataIndex: 'resource',
    key: 'resource'
  },
  {
    title: '操作类型',
    dataIndex: 'action',
    key: 'action'
  },
  {
    title: '创建时间',
    dataIndex: 'created_at',
    key: 'created_at',
    sorter: true
  },
  {
    title: '操作',
    key: 'actions',
    width: 200
  }
]

// 角色表格列配置
const roleColumns = [
  {
    title: '角色名称',
    dataIndex: 'name',
    key: 'name'
  },
  {
    title: '角色代码',
    dataIndex: 'code',
    key: 'code'
  },
  {
    title: '类型',
    dataIndex: 'type',
    key: 'type'
  },
  {
    title: '状态',
    dataIndex: 'status',
    key: 'status'
  },
  {
    title: '描述',
    dataIndex: 'description',
    key: 'description',
    ellipsis: true
  }
]

// 计算属性 - 过滤后的权限列表
const filteredPermissionList = computed(() => {
  let filtered = permissionList.value

  if (searchForm.name) {
    filtered = filtered.filter(item => 
      item.name.toLowerCase().includes(searchForm.name.toLowerCase())
    )
  }

  if (searchForm.code) {
    filtered = filtered.filter(item => 
      item.code.toLowerCase().includes(searchForm.code.toLowerCase())
    )
  }

  if (searchForm.type) {
    filtered = filtered.filter(item => item.type === searchForm.type)
  }

  if (searchForm.status) {
    filtered = filtered.filter(item => item.status === searchForm.status)
  }

  // 更新分页总数
  pagination.total = filtered.length

  // 分页处理
  const start = (pagination.current - 1) * pagination.pageSize
  const end = start + pagination.pageSize
  return filtered.slice(start, end)
})

// 获取权限列表
const fetchPermissionList = async () => {
  try {
    loading.value = true
    const response = await UserService.getPermissionList()
    if (response.data) {
      permissionList.value = response.data || []
      
      // 更新统计数据
      stats.total = permissionList.value.length
      stats.system = permissionList.value.filter(permission => permission.type === 'system').length
      stats.custom = permissionList.value.filter(permission => permission.type === 'custom').length
      stats.active = permissionList.value.filter(permission => permission.status === 'active').length
    }
  } catch (error) {
    console.error('获取权限列表失败:', error)
    message.error('获取权限列表失败')
  } finally {
    loading.value = false
  }
}

// 获取权限关联角色
const fetchPermissionRoles = async (permissionId: number) => {
  try {
    roleLoading.value = true
    const response = await UserService.getPermissionRoles(permissionId)
    if (response.data) {
      permissionRoles.value = response.data || []
    }
  } catch (error) {
    console.error('获取权限角色失败:', error)
    message.error('获取权限角色失败')
  } finally {
    roleLoading.value = false
  }
}

// 表格选择变化
const onSelectChange = (newSelectedRowKeys: number[]) => {
  selectedRowKeys.value = newSelectedRowKeys
}

// 表格变化处理
const handleTableChange = (page: number, pageSize: number) => {
  pagination.current = page
  pagination.pageSize = pageSize
}

// 搜索处理
const handleSearch = () => {
  pagination.current = 1
}

// 重置搜索
const handleReset = () => {
  Object.assign(searchForm, {
    name: '',
    code: '',
    type: '',
    status: ''
  })
  pagination.current = 1
}

// 新建权限
const handleCreate = () => {
  isEdit.value = false
  Object.assign(formData, {
    name: '',
    code: '',
    description: '',
    type: 'custom',
    status: 'active',
    resource: '',
    action: ''
  })
  modalVisible.value = true
}

// 编辑权限
const handleEdit = (record: Permission) => {
  isEdit.value = true
  Object.assign(formData, { ...record })
  modalVisible.value = true
}

// 删除权限
const handleDelete = (record: Permission) => {
  Modal.confirm({
    title: '确认删除',
    content: `确定要删除权限 "${record.name}" 吗？`,
    onOk: async () => {
      try {
        await UserService.deletePermission(record.id)
        message.success('删除成功')
        await fetchPermissionList()
      } catch (error) {
        console.error('删除权限失败:', error)
        message.error('删除权限失败')
      }
    }
  })
}

// 切换状态
const handleToggleStatus = async (record: Permission) => {
  try {
    const newStatus = record.status === 'active' ? 'inactive' : 'active'
    await UserService.updatePermission(record.id, { status: newStatus })
    message.success('状态更新成功')
    await fetchPermissionList()
  } catch (error) {
    console.error('更新状态失败:', error)
    message.error('更新状态失败')
  }
}

// 查看关联角色
const handleViewRoles = async (record: Permission) => {
  selectedPermission.value = record
  rolesModalVisible.value = true
  await fetchPermissionRoles(record.id)
}

// 批量启用
const handleBatchEnable = async () => {
  try {
    await UserService.batchUpdatePermissions(selectedRowKeys.value, { status: 'active' })
    message.success('批量启用成功')
    selectedRowKeys.value = []
    await fetchPermissionList()
  } catch (error) {
    console.error('批量启用失败:', error)
    message.error('批量启用失败')
  }
}

// 批量禁用
const handleBatchDisable = async () => {
  try {
    await UserService.batchUpdatePermissions(selectedRowKeys.value, { status: 'inactive' })
    message.success('批量禁用成功')
    selectedRowKeys.value = []
    await fetchPermissionList()
  } catch (error) {
    console.error('批量禁用失败:', error)
    message.error('批量禁用失败')
  }
}

// 批量删除
const handleBatchDelete = () => {
  Modal.confirm({
    title: '确认删除',
    content: `确定要删除选中的 ${selectedRowKeys.value.length} 个权限吗？`,
    onOk: async () => {
      try {
        await UserService.batchDeletePermissions(selectedRowKeys.value)
        message.success('批量删除成功')
        selectedRowKeys.value = []
        await fetchPermissionList()
      } catch (error) {
        console.error('批量删除失败:', error)
        message.error('批量删除失败')
      }
    }
  })
}

// 导出权限
const handleExport = async () => {
  try {
    await UserService.exportPermissions()
    message.success('导出成功')
  } catch (error) {
    console.error('导出失败:', error)
    message.error('导出失败')
  }
}

// 提交表单
const handleSubmit = async () => {
  try {
    await formRef.value?.validate()
    submitLoading.value = true
    
    const submitData = {
      ...formData,
      type: formData.type as 'system' | 'custom',
      status: formData.status as 'active' | 'inactive'
    }
    
    if (isEdit.value) {
      await UserService.updatePermission(selectedPermission.value!.id, submitData)
      message.success('权限更新成功')
    } else {
      await UserService.createPermission(submitData)
      message.success('权限创建成功')
    }
    
    modalVisible.value = false
    await fetchPermissionList()
  } catch (error) {
    console.error('保存权限失败:', error)
    message.error('保存权限失败')
  } finally {
    submitLoading.value = false
  }
}

// 取消编辑
const handleCancel = () => {
  modalVisible.value = false
  formRef.value?.resetFields()
}

// 组件挂载时获取数据
onMounted(() => {
  fetchPermissionList()
})
</script>

<style scoped>
.permission-list {
  padding: 24px;
}

.page-header {
  margin-bottom: 24px;
}

.page-header h1 {
  margin: 0;
  font-size: 24px;
  font-weight: 600;
}

.page-header p {
  margin: 8px 0 0 0;
  color: #666;
}

.action-bar {
  margin-bottom: 16px;
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
</style>