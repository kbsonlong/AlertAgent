<template>
  <div class="user-detail">
    <!-- 页面头部 -->
    <div class="page-header">
      <a-page-header
        :title="user?.username || '用户详情'"
        @back="$router.go(-1)"
      >
        <template #extra>
          <a-space>
            <a-button @click="handleEdit">
              <template #icon><EditOutlined /></template>
              编辑
            </a-button>
            <a-button 
              type="primary" 
              danger 
              @click="handleDelete"
              :disabled="user?.role === 'admin'"
            >
              <template #icon><DeleteOutlined /></template>
              删除
            </a-button>
          </a-space>
        </template>
      </a-page-header>
    </div>

    <!-- 用户信息卡片 -->
    <a-row :gutter="[16, 16]">
      <!-- 基本信息 -->
      <a-col :span="24" :lg="8">
        <a-card title="基本信息" :loading="loading">
          <div class="user-avatar-section">
            <a-avatar 
              :size="80" 
              :src="user?.avatar"
              class="user-avatar"
            >
              {{ user?.username?.charAt(0)?.toUpperCase() }}
            </a-avatar>
            <div class="user-basic-info">
              <h3>{{ user?.username }}</h3>
              <p class="user-email">{{ user?.email }}</p>
              <a-tag 
                :color="user?.status === 'active' ? 'green' : 'red'"
                class="user-status"
              >
                {{ user?.status === 'active' ? '活跃' : '禁用' }}
              </a-tag>
            </div>
          </div>
          
          <a-divider />
          
          <a-descriptions :column="1" size="small">
            <a-descriptions-item label="角色">
              <a-tag :color="getRoleColor(user?.role)">
                {{ getRoleLabel(user?.role) }}
              </a-tag>
            </a-descriptions-item>
            <a-descriptions-item label="最后登录">
              {{ user?.lastLogin ? formatDateTime(user.lastLogin) : '从未登录' }}
            </a-descriptions-item>
            <a-descriptions-item label="创建时间">
              {{ user?.createdAt ? formatDateTime(user.createdAt) : '-' }}
            </a-descriptions-item>
            <a-descriptions-item label="更新时间">
              {{ user?.updatedAt ? formatDateTime(user.updatedAt) : '-' }}
            </a-descriptions-item>
          </a-descriptions>
        </a-card>
      </a-col>

      <!-- 个人资料 -->
      <a-col :span="24" :lg="8">
        <a-card title="个人资料" :loading="loading">
          <a-descriptions :column="1" size="small">
            <a-descriptions-item label="姓名">
              {{ getUserFullName(user?.profile) || '-' }}
            </a-descriptions-item>
            <a-descriptions-item label="电话">
              {{ user?.profile?.phone || '-' }}
            </a-descriptions-item>
            <a-descriptions-item label="部门">
              {{ user?.profile?.department || '-' }}
            </a-descriptions-item>
            <a-descriptions-item label="职位">
              {{ user?.profile?.position || '-' }}
            </a-descriptions-item>
          </a-descriptions>
        </a-card>
      </a-col>

      <!-- 权限信息 -->
      <a-col :span="24" :lg="8">
        <a-card title="权限信息" :loading="loading">
          <div class="permissions-section">
            <h4>用户权限</h4>
            <div class="permissions-list">
              <a-tag 
                v-for="permission in user?.permissions" 
                :key="permission"
                class="permission-tag"
              >
                {{ permission }}
              </a-tag>
              <a-empty 
                v-if="!user?.permissions?.length" 
                :image="false" 
                description="暂无权限"
                size="small"
              />
            </div>
          </div>
        </a-card>
      </a-col>
    </a-row>

    <!-- 活动日志 -->
    <a-row :gutter="[16, 16]" style="margin-top: 16px;">
      <a-col :span="24">
        <a-card title="活动日志" :loading="logsLoading">
          <template #extra>
            <a-space>
              <a-range-picker 
                v-model:value="logDateRange"
                @change="handleLogDateChange"
                size="small"
              />
              <a-select 
                v-model:value="logAction"
                placeholder="选择操作类型"
                style="width: 120px;"
                size="small"
                allowClear
                @change="loadActivityLogs"
              >
                <a-select-option value="login">登录</a-select-option>
                <a-select-option value="logout">登出</a-select-option>
                <a-select-option value="create">创建</a-select-option>
                <a-select-option value="update">更新</a-select-option>
                <a-select-option value="delete">删除</a-select-option>
              </a-select>
            </a-space>
          </template>
          
          <a-table
            :columns="logColumns"
            :data-source="activityLogs.list"
            :pagination="{
              current: logPagination.current,
              pageSize: logPagination.pageSize,
              total: activityLogs.total,
              showSizeChanger: true,
              showQuickJumper: true,
              showTotal: (total, range) => `第 ${range[0]}-${range[1]} 条，共 ${total} 条`
            }"
            @change="handleLogTableChange"
            size="small"
          >
            <template #bodyCell="{ column, record }">
              <template v-if="column.key === 'action'">
                <a-tag :color="getActionColor(record.action)">
                  {{ getActionLabel(record.action) }}
                </a-tag>
              </template>
              <template v-else-if="column.key === 'createdAt'">
                {{ formatDateTime(record.createdAt) }}
              </template>
            </template>
          </a-table>
        </a-card>
      </a-col>
    </a-row>

    <!-- 编辑用户抽屉 -->
    <a-drawer
      v-model:open="editDrawerVisible"
      title="编辑用户"
      width="600"
      @close="handleEditCancel"
    >
      <UserForm
        v-if="editDrawerVisible"
        :user="user"
        @submit="handleEditSubmit"
        @cancel="handleEditCancel"
      />
    </a-drawer>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { message, Modal } from 'ant-design-vue'
import {
  EditOutlined,
  DeleteOutlined
} from '@ant-design/icons-vue'
import { 
  getUserDetail, 
  deleteUser, 
  getUserActivityLogs,
  type User,
  type UserFormData
} from '@/services/user'
import { formatDateTime } from '@/utils/datetime'
import UserForm from '@/components/UserForm.vue'
import type { TableColumnsType } from 'ant-design-vue'

const route = useRoute()
const router = useRouter()

// 响应式数据
const loading = ref(false)
const logsLoading = ref(false)
const user = ref<User | null>(null)
const editDrawerVisible = ref(false)

// 活动日志相关
const activityLogs = reactive({
  list: [],
  total: 0
})
const logPagination = reactive({
  current: 1,
  pageSize: 10
})
const logDateRange = ref()
const logAction = ref()

// 表格列定义
const logColumns: TableColumnsType = [
  {
    title: '操作',
    dataIndex: 'action',
    key: 'action',
    width: 100
  },
  {
    title: '描述',
    dataIndex: 'description',
    key: 'description'
  },
  {
    title: 'IP地址',
    dataIndex: 'ip',
    key: 'ip',
    width: 120
  },
  {
    title: '用户代理',
    dataIndex: 'userAgent',
    key: 'userAgent',
    ellipsis: true
  },
  {
    title: '时间',
    dataIndex: 'createdAt',
    key: 'createdAt',
    width: 180
  }
]

// 计算属性
const userId = computed(() => route.params.id as string)

// 方法
const loadUserDetail = async () => {
  if (!userId.value) return
  
  loading.value = true
  try {
    user.value = await getUserDetail(userId.value)
  } catch (error) {
    message.error('获取用户详情失败')
    console.error('获取用户详情失败:', error)
  } finally {
    loading.value = false
  }
}

const loadActivityLogs = async () => {
  if (!userId.value) return
  
  logsLoading.value = true
  try {
    const params = {
      page: logPagination.current,
      pageSize: logPagination.pageSize,
      action: logAction.value,
      startTime: logDateRange.value?.[0]?.format('YYYY-MM-DD HH:mm:ss'),
      endTime: logDateRange.value?.[1]?.format('YYYY-MM-DD HH:mm:ss')
    }
    
    const response = await getUserActivityLogs(userId.value, params)
    activityLogs.list = response.list
    activityLogs.total = response.total
  } catch (error) {
    message.error('获取活动日志失败')
    console.error('获取活动日志失败:', error)
  } finally {
    logsLoading.value = false
  }
}

const handleEdit = () => {
  editDrawerVisible.value = true
}

const handleEditSubmit = async (formData: UserFormData) => {
  try {
    message.success('用户更新成功')
    editDrawerVisible.value = false
    await loadUserDetail()
  } catch (error) {
    message.error('用户更新失败')
    console.error('用户更新失败:', error)
  }
}

const handleEditCancel = () => {
  editDrawerVisible.value = false
}

const handleDelete = () => {
  Modal.confirm({
    title: '确认删除',
    content: `确定要删除用户 "${user.value?.username}" 吗？此操作不可恢复。`,
    okText: '确认',
    cancelText: '取消',
    okType: 'danger',
    onOk: async () => {
      try {
        await deleteUser(userId.value)
        message.success('用户删除成功')
        router.push('/users')
      } catch (error) {
        message.error('用户删除失败')
        console.error('用户删除失败:', error)
      }
    }
  })
}

const handleLogDateChange = () => {
  logPagination.current = 1
  loadActivityLogs()
}

const handleLogTableChange = (pagination: any) => {
  logPagination.current = pagination.current
  logPagination.pageSize = pagination.pageSize
  loadActivityLogs()
}

// 辅助函数
const getRoleColor = (role?: string) => {
  const colors = {
    admin: 'red',
    user: 'blue',
    viewer: 'green'
  }
  return colors[role as keyof typeof colors] || 'default'
}

const getRoleLabel = (role?: string) => {
  const labels = {
    admin: '管理员',
    user: '用户',
    viewer: '查看者'
  }
  return labels[role as keyof typeof labels] || role
}

const getUserFullName = (profile?: User['profile']) => {
  if (!profile) return ''
  const { firstName, lastName } = profile
  return [firstName, lastName].filter(Boolean).join(' ')
}

const getActionColor = (action: string) => {
  const colors = {
    login: 'green',
    logout: 'orange',
    create: 'blue',
    update: 'cyan',
    delete: 'red'
  }
  return colors[action as keyof typeof colors] || 'default'
}

const getActionLabel = (action: string) => {
  const labels = {
    login: '登录',
    logout: '登出',
    create: '创建',
    update: '更新',
    delete: '删除'
  }
  return labels[action as keyof typeof labels] || action
}

// 生命周期
onMounted(() => {
  loadUserDetail()
  loadActivityLogs()
})
</script>

<style scoped>
.user-detail {
  padding: 24px;
}

.page-header {
  background: #fff;
  margin-bottom: 16px;
  border-radius: 6px;
}

.user-avatar-section {
  display: flex;
  align-items: center;
  margin-bottom: 16px;
}

.user-avatar {
  margin-right: 16px;
  flex-shrink: 0;
}

.user-basic-info {
  flex: 1;
}

.user-basic-info h3 {
  margin: 0 0 8px 0;
  font-size: 18px;
  font-weight: 600;
}

.user-email {
  margin: 0 0 8px 0;
  color: #666;
  font-size: 14px;
}

.user-status {
  margin: 0;
}

.permissions-section h4 {
  margin: 0 0 12px 0;
  font-size: 14px;
  font-weight: 600;
}

.permissions-list {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.permission-tag {
  margin: 0;
}
</style>