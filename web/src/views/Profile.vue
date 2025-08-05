<template>
  <div class="profile">
    <!-- 页面头部 -->
    <div class="page-header">
      <a-page-header title="个人资料">
        <template #extra>
          <a-space>
            <a-button @click="handleEdit" :loading="loading">
              <template #icon><EditOutlined /></template>
              编辑资料
            </a-button>
            <a-button @click="handleChangePassword">
              <template #icon><KeyOutlined /></template>
              修改密码
            </a-button>
          </a-space>
        </template>
      </a-page-header>
    </div>

    <!-- 个人信息卡片 -->
    <a-row :gutter="[16, 16]">
      <!-- 基本信息 -->
      <a-col :span="24" :lg="12">
        <a-card title="基本信息" :loading="loading">
          <div class="profile-info">
            <div class="avatar-section">
              <a-avatar 
                :size="80" 
                :src="userInfo?.avatar"
                class="user-avatar"
              >
                {{ userInfo?.username?.charAt(0)?.toUpperCase() }}
              </a-avatar>
              <a-upload
                :show-upload-list="false"
                :before-upload="beforeUpload"
                @change="handleAvatarChange"
                class="avatar-upload"
              >
                <a-button size="small" type="link">
                  <template #icon><CameraOutlined /></template>
                  更换头像
                </a-button>
              </a-upload>
            </div>
            
            <a-descriptions :column="1" bordered>
              <a-descriptions-item label="用户名">
                {{ userInfo?.username }}
              </a-descriptions-item>
              <a-descriptions-item label="邮箱">
                {{ userInfo?.email }}
              </a-descriptions-item>
              <a-descriptions-item label="角色">
                <a-tag :color="getRoleColor(userInfo?.role)">
                  {{ getRoleText(userInfo?.role) }}
                </a-tag>
              </a-descriptions-item>
              <a-descriptions-item label="状态">
                <a-tag :color="userInfo?.status === 'active' ? 'green' : 'red'">
                  {{ userInfo?.status === 'active' ? '活跃' : '禁用' }}
                </a-tag>
              </a-descriptions-item>
              <a-descriptions-item label="创建时间">
                {{ formatDate(userInfo?.created_at) }}
              </a-descriptions-item>
              <a-descriptions-item label="最后登录">
                {{ formatDate(userInfo?.last_login_at) }}
              </a-descriptions-item>
            </a-descriptions>
          </div>
        </a-card>
      </a-col>

      <!-- 安全设置 -->
      <a-col :span="24" :lg="12">
        <a-card title="安全设置">
          <a-list>
            <a-list-item>
              <a-list-item-meta
                title="登录密码"
                description="定期更换密码可以提高账户安全性"
              />
              <template #actions>
                <a @click="handleChangePassword">修改</a>
              </template>
            </a-list-item>
            <a-list-item>
              <a-list-item-meta
                title="邮箱验证"
                :description="userInfo?.email_verified ? '已验证' : '未验证'"
              />
              <template #actions>
                <a v-if="!userInfo?.email_verified" @click="handleVerifyEmail">验证</a>
                <span v-else style="color: #52c41a;">✓</span>
              </template>
            </a-list-item>
          </a-list>
        </a-card>
      </a-col>
    </a-row>

    <!-- 编辑资料模态框 -->
    <a-modal
      v-model:open="editModalVisible"
      title="编辑个人资料"
      @ok="handleSaveProfile"
      :confirm-loading="saving"
    >
      <a-form
        ref="formRef"
        :model="editForm"
        :rules="rules"
        layout="vertical"
      >
        <a-form-item label="用户名" name="username">
          <a-input v-model:value="editForm.username" disabled />
        </a-form-item>
        <a-form-item label="邮箱" name="email">
          <a-input v-model:value="editForm.email" />
        </a-form-item>
      </a-form>
    </a-modal>

    <!-- 修改密码模态框 -->
    <a-modal
      v-model:open="passwordModalVisible"
      title="修改密码"
      @ok="handleSavePassword"
      :confirm-loading="saving"
    >
      <a-form
        ref="passwordFormRef"
        :model="passwordForm"
        :rules="passwordRules"
        layout="vertical"
      >
        <a-form-item label="当前密码" name="currentPassword">
          <a-input-password v-model:value="passwordForm.currentPassword" />
        </a-form-item>
        <a-form-item label="新密码" name="newPassword">
          <a-input-password v-model:value="passwordForm.newPassword" />
        </a-form-item>
        <a-form-item label="确认新密码" name="confirmPassword">
          <a-input-password v-model:value="passwordForm.confirmPassword" />
        </a-form-item>
      </a-form>
    </a-modal>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { message } from 'ant-design-vue'
import { EditOutlined, KeyOutlined, CameraOutlined } from '@ant-design/icons-vue'
import { useUserStore } from '@/stores'
import type { User } from '@/types'
import { userApi } from '@/api/user'
import { formatDate } from '@/utils/date'

const userStore = useUserStore()
const loading = ref(false)
const saving = ref(false)
const editModalVisible = ref(false)
const passwordModalVisible = ref(false)
const userInfo = ref<User | null>(null)
const formRef = ref()
const passwordFormRef = ref()

// 编辑表单
const editForm = reactive({
  username: '',
  email: ''
})

// 密码表单
const passwordForm = reactive({
  currentPassword: '',
  newPassword: '',
  confirmPassword: ''
})

// 表单验证规则
const rules = {
  email: [
    { required: true, message: '请输入邮箱', trigger: 'blur' },
    { type: 'email', message: '请输入正确的邮箱格式', trigger: 'blur' }
  ]
}

const passwordRules = {
  currentPassword: [
    { required: true, message: '请输入当前密码', trigger: 'blur' }
  ],
  newPassword: [
    { required: true, message: '请输入新密码', trigger: 'blur' },
    { min: 6, message: '密码长度至少6位', trigger: 'blur' }
  ],
  confirmPassword: [
    { required: true, message: '请确认新密码', trigger: 'blur' },
    {
      validator: (rule: any, value: string) => {
        if (value !== passwordForm.newPassword) {
          return Promise.reject('两次输入的密码不一致')
        }
        return Promise.resolve()
      },
      trigger: 'blur'
    }
  ]
}

// 获取角色颜色
const getRoleColor = (role?: string) => {
  const colors: Record<string, string> = {
    admin: 'red',
    operator: 'blue',
    viewer: 'green'
  }
  return colors[role || ''] || 'default'
}

// 获取角色文本
const getRoleText = (role?: string) => {
  const texts: Record<string, string> = {
    admin: '管理员',
    operator: '操作员',
    viewer: '查看者'
  }
  return texts[role || ''] || role
}

// 获取用户信息
const fetchUserInfo = async () => {
  try {
    loading.value = true
    await userStore.fetchCurrentUser()
    userInfo.value = userStore.user
    
    // 更新编辑表单
    editForm.username = userInfo.value?.username || ''
    editForm.email = userInfo.value?.email || ''
  } catch (error) {
    console.error('获取用户信息失败:', error)
    message.error('获取用户信息失败')
  } finally {
    loading.value = false
  }
}

// 编辑资料
const handleEdit = () => {
  editModalVisible.value = true
}

// 保存资料
const handleSaveProfile = async () => {
  try {
    await formRef.value.validate()
    saving.value = true
    
    await userApi.updateProfile({
      email: editForm.email
    })
    
    message.success('个人资料更新成功')
    editModalVisible.value = false
    await fetchUserInfo()
  } catch (error) {
    console.error('更新个人资料失败:', error)
    message.error('更新个人资料失败')
  } finally {
    saving.value = false
  }
}

// 修改密码
const handleChangePassword = () => {
  passwordModalVisible.value = true
  // 重置表单
  passwordForm.currentPassword = ''
  passwordForm.newPassword = ''
  passwordForm.confirmPassword = ''
}

// 保存密码
const handleSavePassword = async () => {
  try {
    await passwordFormRef.value.validate()
    saving.value = true
    
    if (!userInfo.value?.id) {
      message.error('用户信息获取失败，请刷新页面重试')
      return
    }
    
    await userApi.changePassword(userInfo.value.id.toString(), {
      currentPassword: passwordForm.currentPassword,
      newPassword: passwordForm.newPassword
    })
    
    message.success('密码修改成功')
    passwordModalVisible.value = false
  } catch (error) {
    console.error('修改密码失败:', error)
    message.error('修改密码失败')
  } finally {
    saving.value = false
  }
}

// 验证邮箱
const handleVerifyEmail = async () => {
  try {
    await userApi.sendVerificationEmail()
    message.success('验证邮件已发送，请查收')
  } catch (error) {
    console.error('发送验证邮件失败:', error)
    message.error('发送验证邮件失败')
  }
}

// 头像上传前验证
const beforeUpload = (file: File) => {
  const isJpgOrPng = file.type === 'image/jpeg' || file.type === 'image/png'
  if (!isJpgOrPng) {
    message.error('只能上传 JPG/PNG 格式的图片!')
    return false
  }
  const isLt2M = file.size / 1024 / 1024 < 2
  if (!isLt2M) {
    message.error('图片大小不能超过 2MB!')
    return false
  }
  return true
}

// 头像上传
const handleAvatarChange = async (info: any) => {
  if (info.file.status === 'uploading') {
    loading.value = true
    return
  }
  if (info.file.status === 'done') {
    // 这里应该调用上传头像的API
    message.success('头像上传成功')
    loading.value = false
    await fetchUserInfo()
  }
  if (info.file.status === 'error') {
    message.error('头像上传失败')
    loading.value = false
  }
}

onMounted(() => {
  fetchUserInfo()
})
</script>

<style scoped>
.profile {
  padding: 24px;
}

.page-header {
  margin-bottom: 24px;
}

.profile-info {
  text-align: center;
}

.avatar-section {
  margin-bottom: 24px;
}

.user-avatar {
  margin-bottom: 8px;
}

.avatar-upload {
  display: block;
}

.ant-descriptions {
  text-align: left;
}
</style>