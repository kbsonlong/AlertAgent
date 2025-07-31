<template>
  <div class="user-detail-component">
    <!-- 用户基本信息 -->
    <div class="user-info-section">
      <div class="user-avatar-section">
        <a-avatar 
          :size="64" 
          class="user-avatar"
        >
          {{ user?.username?.charAt(0)?.toUpperCase() }}
        </a-avatar>
        <div class="user-basic-info">
          <h3>{{ user?.username }}</h3>
          <p class="user-email">{{ user?.email }}</p>
          <a-tag 
            color="green"
            class="user-status"
          >
            活跃
          </a-tag>
        </div>
      </div>
    </div>

    <a-divider />

    <!-- 详细信息 -->
    <div class="user-details-section">
      <a-descriptions :column="1" size="small">
        <a-descriptions-item label="角色">
          <a-tag :color="getRoleColor(user?.role)">
            {{ getRoleLabel(user?.role) }}
          </a-tag>
        </a-descriptions-item>
        <a-descriptions-item label="创建时间">
          {{ user?.created_at ? formatDateTime(user.created_at) : '-' }}
        </a-descriptions-item>
        <a-descriptions-item label="更新时间">
          {{ user?.updated_at ? formatDateTime(user.updated_at) : '-' }}
        </a-descriptions-item>
        <a-descriptions-item label="最后登录">
          从未登录
        </a-descriptions-item>
      </a-descriptions>
    </div>

    <!-- 个人资料 -->
    <div class="user-profile-section">
      <a-divider orientation="left">个人资料</a-divider>
      <a-descriptions :column="1" size="small">
        <a-descriptions-item label="姓名">
          -
        </a-descriptions-item>
        <a-descriptions-item label="电话">
          -
        </a-descriptions-item>
        <a-descriptions-item label="部门">
          -
        </a-descriptions-item>
        <a-descriptions-item label="职位">
          -
        </a-descriptions-item>
      </a-descriptions>
    </div>

    <!-- 权限信息 -->
    <div class="user-permissions-section">
      <a-divider orientation="left">权限信息</a-divider>
      <div class="permissions-list">
        <a-tag class="permission-tag">
          基础权限
        </a-tag>
      </div>
    </div>

    <!-- 操作按钮 -->
    <div class="user-actions">
      <a-space>
        <a-button type="primary" @click="handleEdit">
          <template #icon><EditOutlined /></template>
          编辑
        </a-button>
        <a-button 
          danger 
          @click="handleDelete"
          :disabled="user?.role === 'admin'"
        >
          <template #icon><DeleteOutlined /></template>
          删除
        </a-button>
        <a-button @click="handleClose">
          关闭
        </a-button>
      </a-space>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import {
  Avatar,
  Tag,
  Divider,
  Descriptions,
  DescriptionsItem,
  Button,
  Space,
  Modal,
  message
} from 'ant-design-vue'
import {
  EditOutlined,
  DeleteOutlined
} from '@ant-design/icons-vue'
import type { User } from '@/types'
import { formatDateTime } from '@/utils/datetime'

// 组件属性
interface Props {
  user: User | null
}

const props = defineProps<Props>()

// 组件事件
interface Emits {
  edit: [user: User]
  delete: [user: User]
  close: []
}

const emit = defineEmits<Emits>()

// 获取角色颜色
const getRoleColor = (role?: string) => {
  const colorMap: Record<string, string> = {
    admin: 'red',
    operator: 'blue',
    viewer: 'green'
  }
  return colorMap[role || ''] || 'default'
}

// 获取角色标签
const getRoleLabel = (role?: string) => {
  const labelMap: Record<string, string> = {
    admin: '管理员',
    operator: '操作员',
    viewer: '查看者'
  }
  return labelMap[role || ''] || role || '-'
}

// 获取用户全名 (暂时不使用)
// const getUserFullName = (profile?: any) => {
//   if (!profile) return ''
//   const { firstName, lastName, fullName } = profile
//   return fullName || `${firstName || ''} ${lastName || ''}`.trim() || ''
// }

// 处理编辑
const handleEdit = () => {
  if (props.user) {
    emit('edit', props.user)
  }
}

// 处理删除
const handleDelete = () => {
  if (!props.user) return
  
  Modal.confirm({
    title: '确认删除',
    content: `确定要删除用户 "${props.user.username}" 吗？此操作不可恢复。`,
    okText: '确定',
    cancelText: '取消',
    okType: 'danger',
    onOk() {
      if (props.user) {
        emit('delete', props.user)
      }
    }
  })
}

// 处理关闭
const handleClose = () => {
  emit('close')
}
</script>

<style scoped>
.user-detail-component {
  padding: 16px;
}

.user-info-section {
  margin-bottom: 16px;
}

.user-avatar-section {
  display: flex;
  align-items: center;
  gap: 16px;
}

.user-basic-info h3 {
  margin: 0 0 4px 0;
  font-size: 18px;
  font-weight: 600;
  color: #262626;
}

.user-email {
  margin: 0 0 8px 0;
  color: #8c8c8c;
  font-size: 14px;
}

.user-status {
  font-size: 12px;
}

.user-details-section,
.user-profile-section,
.user-permissions-section {
  margin-bottom: 16px;
}

.permissions-list {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.permission-tag {
  margin: 0;
}

.user-actions {
  padding-top: 16px;
  border-top: 1px solid #f0f0f0;
  text-align: right;
}

:deep(.ant-descriptions-item-label) {
  font-weight: 500;
  color: #595959;
}

:deep(.ant-descriptions-item-content) {
  color: #262626;
}

:deep(.ant-divider-horizontal.ant-divider-with-text) {
  margin: 16px 0;
}

:deep(.ant-divider-horizontal.ant-divider-with-text-left::before) {
  width: 5%;
}

:deep(.ant-divider-horizontal.ant-divider-with-text-left::after) {
  width: 95%;
}
</style>