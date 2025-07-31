<template>
  <div class="role-list-page">
    <div class="page-header">
      <h1 class="page-title">
        <SafetyCertificateOutlined />
        角色管理
      </h1>
      <div class="page-actions">
        <a-space>
          <a-button @click="handleExport">
            <ExportOutlined /> 导出角色
          </a-button>
          <a-button type="primary" @click="handleCreate">
            <PlusOutlined /> 新建角色
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
              title="总角色数"
              :value="stats.total"
              :value-style="{ color: '#1890ff' }"
            >
              <template #prefix>
                <SafetyCertificateOutlined />
              </template>
            </a-statistic>
          </a-card>
        </a-col>
        <a-col :span="6">
          <a-card>
            <a-statistic
              title="系统角色"
              :value="stats.system"
              :value-style="{ color: '#52c41a' }"
            >
              <template #prefix>
                <SettingOutlined />
              </template>
            </a-statistic>
          </a-card>
        </a-col>
        <a-col :span="6">
          <a-card>
            <a-statistic
              title="自定义角色"
              :value="stats.custom"
              :value-style="{ color: '#722ed1' }"
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
              title="活跃角色"
              :value="stats.active"
              :value-style="{ color: '#fa8c16' }"
            >
              <template #prefix>
                <CheckCircleOutlined />
              </template>
            </a-statistic>
          </a-card>
        </a-col>
      </a-row>
    </div>

    <!-- 搜索和筛选 -->
    <div class="search-form">
      <a-form layout="inline" :model="searchForm" @finish="handleSearch">
        <a-form-item label="角色名称" name="name">
          <a-input
            v-model:value="searchForm.name"
            placeholder="请输入角色名称"
            allow-clear
          />
        </a-form-item>
        <a-form-item label="角色类型" name="type">
          <a-select
            v-model:value="searchForm.type"
            placeholder="请选择类型"
            allow-clear
            style="width: 120px"
          >
            <a-select-option value="system">系统角色</a-select-option>
            <a-select-option value="custom">自定义角色</a-select-option>
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
            <a-select-option value="inactive">禁用</a-select-option>
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

    <!-- 角色列表 -->
    <div class="role-table">
      <a-table
        :columns="columns"
        :data-source="roleList"
        :loading="loading"
        :pagination="pagination"
        :row-selection="rowSelection"
        @change="handleTableChange"
        row-key="id"
      >
        <template #bodyCell="{ column, record }">
          <template v-if="column.key === 'name'">
            <div class="role-name-cell">
              <a-avatar :size="32" class="role-avatar">
                <template #icon>
                  <SafetyCertificateOutlined v-if="record.type === 'system'" />
                  <UserOutlined v-else />
                </template>
              </a-avatar>
              <div class="role-info">
                <div class="role-name">{{ record.name }}</div>
                <div class="role-code">{{ record.code }}</div>
              </div>
            </div>
          </template>
          
          <template v-else-if="column.key === 'type'">
            <a-tag :color="record.type === 'system' ? 'blue' : 'purple'">
              {{ record.type === 'system' ? '系统角色' : '自定义角色' }}
            </a-tag>
          </template>
          
          <template v-else-if="column.key === 'status'">
            <a-tag :color="record.status === 'active' ? 'green' : 'red'">
              {{ record.status === 'active' ? '活跃' : '禁用' }}
            </a-tag>
          </template>
          
          <template v-else-if="column.key === 'permissions'">
            <a-tooltip title="点击查看权限详情">
              <a-tag color="blue" style="cursor: pointer" @click="showPermissions(record)">
                {{ record.permissions?.length || 0 }} 个权限
              </a-tag>
            </a-tooltip>
          </template>
          
          <template v-else-if="column.key === 'user_count'">
            <a-tooltip title="点击查看用户列表">
              <a-tag color="orange" style="cursor: pointer" @click="showUsers(record)">
                {{ record.user_count || 0 }} 个用户
              </a-tag>
            </a-tooltip>
          </template>
          
          <template v-else-if="column.key === 'action'">
            <a-space>
              <a-button type="link" size="small" @click="handleView(record)">
                查看
              </a-button>
              <a-button 
                type="link" 
                size="small" 
                @click="handleEdit(record)"
                :disabled="record.type === 'system'"
              >
                编辑
              </a-button>
              <a-button 
                type="link" 
                size="small" 
                @click="handleCopy(record)"
              >
                复制
              </a-button>
              <a-popconfirm
                title="确定要删除这个角色吗？"
                @confirm="handleDelete(record)"
                :disabled="record.type === 'system'"
              >
                <a-button 
                  type="link" 
                  size="small" 
                  danger
                  :disabled="record.type === 'system'"
                >
                  删除
                </a-button>
              </a-popconfirm>
            </a-space>
          </template>
        </template>
      </a-table>
    </div>

    <!-- 角色表单弹窗 -->
    <a-modal
      v-model:open="modalVisible"
      :title="modalTitle"
      :width="800"
      @ok="handleSubmit"
      @cancel="handleCancel"
      :confirm-loading="submitLoading"
    >
      <a-form
        ref="formRef"
        :model="formData"
        :rules="formRules"
        layout="vertical"
      >
        <a-row :gutter="16">
          <a-col :span="12">
            <a-form-item label="角色名称" name="name">
              <a-input v-model:value="formData.name" placeholder="请输入角色名称" />
            </a-form-item>
          </a-col>
          <a-col :span="12">
            <a-form-item label="角色代码" name="code">
              <a-input v-model:value="formData.code" placeholder="请输入角色代码" />
            </a-form-item>
          </a-col>
        </a-row>
        
        <a-form-item label="角色描述" name="description">
          <a-textarea 
            v-model:value="formData.description" 
            placeholder="请输入角色描述"
            :rows="3"
          />
        </a-form-item>
        
        <a-form-item label="角色状态" name="status">
          <a-radio-group v-model:value="formData.status">
            <a-radio value="active">活跃</a-radio>
            <a-radio value="inactive">禁用</a-radio>
          </a-radio-group>
        </a-form-item>
        
        <a-form-item label="权限配置" name="permission_ids">
          <div class="permission-tree">
            <a-tree
              v-model:checkedKeys="formData.permission_ids"
              :tree-data="permissionTree"
              checkable
              :check-strictly="false"
              :expanded-keys="expandedKeys"
              @expand="onExpand"
            >
              <template #title="{ title, key, permission }">
                <div class="permission-node">
                  <span class="permission-name">{{ title }}</span>
                  <a-tag size="small" color="blue">{{ permission?.action }}</a-tag>
                </div>
              </template>
            </a-tree>
          </div>
        </a-form-item>
      </a-form>
    </a-modal>

    <!-- 权限详情弹窗 -->
    <a-modal
      v-model:open="permissionModalVisible"
      title="权限详情"
      :width="600"
      :footer="null"
    >
      <div class="permission-details">
        <a-descriptions :column="1" bordered>
          <a-descriptions-item label="角色名称">
            {{ selectedRole?.name }}
          </a-descriptions-item>
          <a-descriptions-item label="权限数量">
            {{ selectedRole?.permissions?.length || 0 }}
          </a-descriptions-item>
        </a-descriptions>
        
        <a-divider>权限列表</a-divider>
        
        <a-list
          :data-source="selectedRole?.permissions || []"
          size="small"
        >
          <template #renderItem="{ item }">
            <a-list-item>
              <a-list-item-meta>
                <template #title>
                  <span>{{ item.name }}</span>
                  <a-tag color="blue" style="margin-left: 8px">{{ item.action }}</a-tag>
                </template>
                <template #description>
                  <div>
                    <span>资源: {{ item.resource }}</span>
                    <a-divider type="vertical" />
                    <span>分类: {{ item.category }}</span>
                  </div>
                  <div style="margin-top: 4px; color: #666;">
                    {{ item.description }}
                  </div>
                </template>
              </a-list-item-meta>
            </a-list-item>
          </template>
        </a-list>
      </div>
    </a-modal>

    <!-- 用户列表弹窗 -->
    <a-modal
      v-model:open="userModalVisible"
      title="角色用户"
      :width="800"
      :footer="null"
    >
      <div class="role-users">
        <a-descriptions :column="1" bordered>
          <a-descriptions-item label="角色名称">
            {{ selectedRole?.name }}
          </a-descriptions-item>
          <a-descriptions-item label="用户数量">
            {{ selectedRole?.user_count || 0 }}
          </a-descriptions-item>
        </a-descriptions>
        
        <a-divider>用户列表</a-divider>
        
        <a-table
          :columns="userColumns"
          :data-source="roleUsers"
          :loading="userLoading"
          :pagination="false"
          size="small"
        >
          <template #bodyCell="{ column, record }">
            <template v-if="column.key === 'avatar'">
              <a-avatar :src="record.avatar" :size="32">
                {{ record.username?.charAt(0)?.toUpperCase() }}
              </a-avatar>
            </template>
            <template v-else-if="column.key === 'status'">
              <a-tag :color="record.status === 'active' ? 'green' : 'red'">
                {{ record.status === 'active' ? '活跃' : '禁用' }}
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
import { message } from 'ant-design-vue'
import {
  SafetyCertificateOutlined,
  ExportOutlined,
  PlusOutlined,
  SearchOutlined,
  ReloadOutlined,
  UserOutlined,
  SettingOutlined,
  CheckCircleOutlined
} from '@ant-design/icons-vue'
import type { TableColumnsType, TreeProps } from 'ant-design-vue'
import { UserService, type Role, type Permission, type User } from '@/services/userService'

// 响应式数据
const loading = ref(false)
const submitLoading = ref(false)
const userLoading = ref(false)
const modalVisible = ref(false)
const permissionModalVisible = ref(false)
const userModalVisible = ref(false)
const selectedRowKeys = ref<number[]>([])
const expandedKeys = ref<number[]>([])

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
  type: undefined,
  status: undefined
})

// 角色列表
const roleList = ref<Role[]>([])
const selectedRole = ref<Role | null>(null)
const roleUsers = ref<User[]>([])

// 权限树
const permissionTree = ref<any[]>([])
const allPermissions = ref<Permission[]>([])

// 表单数据
const formData = reactive({
  name: '',
  code: '',
  description: '',
  status: 'active',
  permission_ids: [] as number[]
})

// 表单引用
const formRef = ref()

// 表单验证规则
const formRules = {
  name: [{ required: true, message: '请输入角色名称' }],
  code: [{ required: true, message: '请输入角色代码' }]
}

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
const columns: TableColumnsType = [
  {
    title: '角色名称',
    key: 'name',
    width: 200
  },
  {
    title: '类型',
    key: 'type',
    width: 100
  },
  {
    title: '状态',
    key: 'status',
    width: 100
  },
  {
    title: '权限',
    key: 'permissions',
    width: 120
  },
  {
    title: '用户数',
    key: 'user_count',
    width: 100
  },
  {
    title: '创建时间',
    dataIndex: 'created_at',
    width: 180
  },
  {
    title: '操作',
    key: 'action',
    width: 200,
    fixed: 'right'
  }
]

// 用户列表列配置
const userColumns: TableColumnsType = [
  {
    title: '头像',
    key: 'avatar',
    width: 60
  },
  {
    title: '用户名',
    dataIndex: 'username',
    width: 120
  },
  {
    title: '邮箱',
    dataIndex: 'email',
    width: 200
  },
  {
    title: '状态',
    key: 'status',
    width: 80
  },
  {
    title: '最后登录',
    dataIndex: 'last_login_at',
    width: 150
  }
]

// 行选择配置
const rowSelection = {
  selectedRowKeys,
  onChange: (keys: number[]) => {
    selectedRowKeys.value = keys
  }
}

// 计算属性
const modalTitle = computed(() => {
  return formData.name ? '编辑角色' : '新建角色'
})

// 获取角色列表
const fetchRoleList = async () => {
  try {
    loading.value = true
    const response = await UserService.getRoleList()
    if (response.data) {
      roleList.value = response.data || []
      
      // 更新统计数据
      stats.total = roleList.value.length
      stats.system = roleList.value.filter(role => role.type === 'system').length
      stats.custom = roleList.value.filter(role => role.type === 'custom').length
      stats.active = roleList.value.filter(role => role.status === 'active').length
    }
  } catch (error) {
    console.error('获取角色列表失败:', error)
    message.error('获取角色列表失败')
  } finally {
    loading.value = false
  }
}

// 获取权限列表
const fetchPermissions = async () => {
  try {
    const response = await UserService.getPermissionList()
    if (response.data) {
      allPermissions.value = response.data || []
      buildPermissionTree()
    }
  } catch (error) {
    console.error('获取权限列表失败:', error)
  }
}

// 构建权限树
const buildPermissionTree = () => {
  const tree: any[] = []
  const categoryMap = new Map()
  
  allPermissions.value.forEach(permission => {
    if (!categoryMap.has(permission.category)) {
      const categoryNode = {
        title: permission.category,
        key: `category-${permission.category}`,
        children: []
      }
      categoryMap.set(permission.category, categoryNode)
      tree.push(categoryNode)
    }
    
    const categoryNode = categoryMap.get(permission.category)
    categoryNode.children.push({
      title: permission.name,
      key: permission.id,
      permission,
      isLeaf: true
    })
  })
  
  permissionTree.value = tree
  expandedKeys.value = tree.map(node => node.key)
}

// 搜索处理
const handleSearch = () => {
  pagination.current = 1
  fetchRoleList()
}

// 重置搜索
const handleReset = () => {
  Object.assign(searchForm, {
    name: '',
    type: undefined,
    status: undefined
  })
  handleSearch()
}

// 表格变化处理
const handleTableChange = (pag: any) => {
  pagination.current = pag.current
  pagination.pageSize = pag.pageSize
  fetchRoleList()
}

// 新建角色
const handleCreate = () => {
  resetForm()
  modalVisible.value = true
}

// 编辑角色
const handleEdit = (record: Role) => {
  Object.assign(formData, {
    ...record,
    permission_ids: record.permissions?.map(p => p.id) || []
  })
  modalVisible.value = true
}

// 查看角色
const handleView = (record: Role) => {
  selectedRole.value = record
  permissionModalVisible.value = true
}

// 复制角色
const handleCopy = (record: Role) => {
  Object.assign(formData, {
    name: `${record.name}_副本`,
    code: `${record.code}_copy`,
    description: record.description,
    status: 'active',
    permission_ids: record.permissions?.map(p => p.id) || []
  })
  modalVisible.value = true
}

// 删除角色
const handleDelete = async (record: Role) => {
  try {
    await UserService.deleteRole(record.id)
    message.success('删除成功')
    fetchRoleList()
  } catch (error) {
    console.error('删除角色失败:', error)
    message.error('删除失败')
  }
}

// 导出角色
const handleExport = () => {
  message.info('导出功能开发中')
}

// 显示权限详情
const showPermissions = (record: Role) => {
  selectedRole.value = record
  permissionModalVisible.value = true
}

// 显示用户列表
const showUsers = async (record: Role) => {
  selectedRole.value = record
  userModalVisible.value = true
  
  try {
    userLoading.value = true
    const response = await UserService.getRoleUsers(record.id)
    if (response.data) {
      roleUsers.value = response.data || []
    }
  } catch (error) {
    console.error('获取角色用户失败:', error)
    message.error('获取用户列表失败')
  } finally {
    userLoading.value = false
  }
}

// 提交表单
const handleSubmit = async () => {
  try {
    await formRef.value.validate()
    submitLoading.value = true
    
    if (formData.name && roleList.value.find(r => r.name === formData.name)) {
      // 编辑模式
      const role = roleList.value.find(r => r.name === formData.name)
      if (role) {
        const updateData = {
          ...formData,
          status: formData.status as 'active' | 'inactive'
        }
        await UserService.updateRole(role.id, updateData)
        message.success('更新成功')
      }
    } else {
      // 新建模式
      const roleData = {
        ...formData,
        type: 'custom' as const, // 新建的都是自定义角色
        status: formData.status as 'active' | 'inactive'
      }
      await UserService.createRole(roleData)
      message.success('创建成功')
    }
    
    modalVisible.value = false
    fetchRoleList()
  } catch (error) {
    console.error('保存角色失败:', error)
    message.error('保存失败')
  } finally {
    submitLoading.value = false
  }
}

// 取消表单
const handleCancel = () => {
  modalVisible.value = false
  resetForm()
}

// 重置表单
const resetForm = () => {
  Object.assign(formData, {
    name: '',
    code: '',
    description: '',
    status: 'active',
    permission_ids: []
  })
  formRef.value?.resetFields()
}

// 树展开处理
const onExpand: TreeProps['onExpand'] = (keys) => {
  expandedKeys.value = keys as number[]
}

// 组件挂载
onMounted(() => {
  fetchRoleList()
  fetchPermissions()
})
</script>

<style scoped>
.role-list-page {
  padding: 24px;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 24px;
}

.page-title {
  margin: 0;
  font-size: 24px;
  font-weight: 600;
  display: flex;
  align-items: center;
  gap: 8px;
}

.stats-cards {
  margin-bottom: 24px;
}

.search-form {
  background: #fff;
  padding: 24px;
  border-radius: 8px;
  margin-bottom: 16px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}

.role-table {
  background: #fff;
  border-radius: 8px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}

.role-name-cell {
  display: flex;
  align-items: center;
  gap: 12px;
}

.role-info {
  display: flex;
  flex-direction: column;
}

.role-name {
  font-weight: 500;
  color: #262626;
}

.role-code {
  font-size: 12px;
  color: #8c8c8c;
}

.role-avatar {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
}

.permission-tree {
  max-height: 300px;
  overflow-y: auto;
  border: 1px solid #d9d9d9;
  border-radius: 6px;
  padding: 8px;
}

.permission-node {
  display: flex;
  align-items: center;
  justify-content: space-between;
  width: 100%;
}

.permission-name {
  flex: 1;
}

.permission-details {
  max-height: 500px;
  overflow-y: auto;
}

.role-users {
  max-height: 500px;
  overflow-y: auto;
}
</style>