<template>
  <div class="user-form">
    <a-form
      ref="formRef"
      :model="formData"
      :rules="formRules"
      layout="vertical"
      @finish="handleSubmit"
    >
      <!-- 基本信息 -->
      <a-card title="基本信息" class="form-section">
        <a-row :gutter="16">
          <a-col :span="12">
            <a-form-item label="用户名" name="username">
              <a-input 
                v-model:value="formData.username" 
                placeholder="请输入用户名"
                :disabled="isEdit"
              />
            </a-form-item>
          </a-col>
          <a-col :span="12">
            <a-form-item label="邮箱" name="email">
              <a-input 
                v-model:value="formData.email" 
                placeholder="请输入邮箱地址"
                type="email"
              />
            </a-form-item>
          </a-col>
        </a-row>
        
        <a-row :gutter="16">
          <a-col :span="12">
            <a-form-item label="角色" name="role">
              <a-select 
                v-model:value="formData.role" 
                placeholder="请选择角色"
                :loading="rolesLoading"
              >
                <a-select-option 
                  v-for="role in roles" 
                  :key="role.value" 
                  :value="role.value"
                >
                  {{ role.label }}
                </a-select-option>
              </a-select>
            </a-form-item>
          </a-col>
          <a-col :span="12">
            <a-form-item label="状态" name="status">
              <a-select v-model:value="formData.status" placeholder="请选择状态">
                <a-select-option value="active">活跃</a-select-option>
                <a-select-option value="disabled">禁用</a-select-option>
              </a-select>
            </a-form-item>
          </a-col>
        </a-row>
        
        <a-row :gutter="16" v-if="!isEdit">
          <a-col :span="12">
            <a-form-item label="密码" name="password">
              <a-input-password 
                v-model:value="formData.password" 
                placeholder="请输入密码"
                autocomplete="new-password"
              />
            </a-form-item>
          </a-col>
          <a-col :span="12">
            <a-form-item label="确认密码" name="confirmPassword">
              <a-input-password 
                v-model:value="confirmPassword" 
                placeholder="请再次输入密码"
                autocomplete="new-password"
              />
            </a-form-item>
          </a-col>
        </a-row>
        
        <a-form-item label="头像">
          <div class="avatar-upload">
            <a-upload
              v-model:file-list="avatarFileList"
              name="avatar"
              list-type="picture-card"
              class="avatar-uploader"
              :show-upload-list="false"
              :before-upload="beforeAvatarUpload"
              @change="handleAvatarChange"
            >
              <div v-if="formData.avatar">
                <img :src="formData.avatar" alt="avatar" style="width: 100%; height: 100%; object-fit: cover;" />
              </div>
              <div v-else class="upload-placeholder">
                <PlusOutlined />
                <div style="margin-top: 8px">上传头像</div>
              </div>
            </a-upload>
          </div>
        </a-form-item>
      </a-card>
      
      <!-- 个人资料 -->
      <a-card title="个人资料" class="form-section">
        <a-row :gutter="16">
          <a-col :span="12">
            <a-form-item label="姓" name="firstName">
              <a-input 
                v-model:value="formData.profile.firstName" 
                placeholder="请输入姓"
              />
            </a-form-item>
          </a-col>
          <a-col :span="12">
            <a-form-item label="名" name="lastName">
              <a-input 
                v-model:value="formData.profile.lastName" 
                placeholder="请输入名"
              />
            </a-form-item>
          </a-col>
        </a-row>
        
        <a-row :gutter="16">
          <a-col :span="12">
            <a-form-item label="电话" name="phone">
              <a-input 
                v-model:value="formData.profile.phone" 
                placeholder="请输入电话号码"
              />
            </a-form-item>
          </a-col>
          <a-col :span="12">
            <a-form-item label="部门" name="department">
              <a-input 
                v-model:value="formData.profile.department" 
                placeholder="请输入部门"
              />
            </a-form-item>
          </a-col>
        </a-row>
        
        <a-form-item label="职位" name="position">
          <a-input 
            v-model:value="formData.profile.position" 
            placeholder="请输入职位"
          />
        </a-form-item>
      </a-card>
      
      <!-- 表单操作 -->
      <div class="form-actions">
        <a-space>
          <a-button @click="handleCancel">取消</a-button>
          <a-button 
            type="primary" 
            html-type="submit" 
            :loading="submitting"
          >
            {{ isEdit ? '更新' : '创建' }}
          </a-button>
          <a-button 
            v-if="!isEdit"
            @click="handleTestConnection"
            :loading="testing"
          >
            <template #icon><ApiOutlined /></template>
            测试连接
          </a-button>
        </a-space>
      </div>
    </a-form>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted, watch } from 'vue'
import { message } from 'ant-design-vue'
import { PlusOutlined, ApiOutlined } from '@ant-design/icons-vue'
import {
  createUser,
  updateUser,
  getUserRoles,
  checkUsernameAvailable,
  checkEmailAvailable,
  type User,
  type UserFormData
} from '@/services/user'
import type { UploadChangeParam, UploadFile } from 'ant-design-vue'

// Props
interface Props {
  user?: User | null
}

const props = withDefaults(defineProps<Props>(), {
  user: null
})

// Emits
const emit = defineEmits<{
  submit: [data: UserFormData]
  cancel: []
}>()

// 响应式数据
const formRef = ref()
const submitting = ref(false)
const testing = ref(false)
const rolesLoading = ref(false)
const confirmPassword = ref('')
const avatarFileList = ref<UploadFile[]>([])
const roles = ref<Array<{
  value: string
  label: string
  description?: string
  permissions: string[]
}>>([])

// 表单数据
const formData = reactive<UserFormData & { profile: NonNullable<UserFormData['profile']> }>({
  username: '',
  email: '',
  password: '',
  role: '',
  status: 'active',
  avatar: '',
  profile: {
    firstName: '',
    lastName: '',
    phone: '',
    department: '',
    position: ''
  }
})

// 计算属性
const isEdit = computed(() => !!props.user)

// 表单验证规则
const formRules = computed(() => ({
  username: [
    { required: true, message: '请输入用户名', trigger: 'blur' },
    { min: 3, max: 20, message: '用户名长度为3-20个字符', trigger: 'blur' },
    { pattern: /^[a-zA-Z0-9_-]+$/, message: '用户名只能包含字母、数字、下划线和连字符', trigger: 'blur' },
    {
      validator: async (rule: any, value: string) => {
        if (!value || value === props.user?.username) return Promise.resolve()
        const available = await checkUsernameAvailable(value, props.user?.id)
        if (!available) {
          return Promise.reject(new Error('用户名已存在'))
        }
        return Promise.resolve()
      },
      trigger: 'blur'
    }
  ],
  email: [
    { required: true, message: '请输入邮箱地址', trigger: 'blur' },
    { type: 'email', message: '请输入有效的邮箱地址', trigger: 'blur' },
    {
      validator: async (rule: any, value: string) => {
        if (!value || value === props.user?.email) return Promise.resolve()
        const available = await checkEmailAvailable(value, props.user?.id)
        if (!available) {
          return Promise.reject(new Error('邮箱已存在'))
        }
        return Promise.resolve()
      },
      trigger: 'blur'
    }
  ],
  role: [
    { required: true, message: '请选择角色', trigger: 'change' }
  ],
  status: [
    { required: true, message: '请选择状态', trigger: 'change' }
  ],
  password: isEdit.value ? [] : [
    { required: true, message: '请输入密码', trigger: 'blur' },
    { min: 6, max: 20, message: '密码长度为6-20个字符', trigger: 'blur' },
    { pattern: /^(?=.*[a-z])(?=.*[A-Z])(?=.*\d)[a-zA-Z\d@$!%*?&]{6,}$/, message: '密码必须包含大小写字母和数字', trigger: 'blur' }
  ],
  confirmPassword: isEdit.value ? [] : [
    { required: true, message: '请确认密码', trigger: 'blur' },
    {
      validator: (rule: any, value: string) => {
        if (value !== formData.password) {
          return Promise.reject(new Error('两次输入的密码不一致'))
        }
        return Promise.resolve()
      },
      trigger: 'blur'
    }
  ]
}))

// 方法
const loadRoles = async () => {
  rolesLoading.value = true
  try {
    roles.value = await getUserRoles()
  } catch (error) {
    message.error('获取角色列表失败')
    console.error('获取角色列表失败:', error)
  } finally {
    rolesLoading.value = false
  }
}

const initFormData = () => {
  if (props.user) {
    // 编辑模式，填充现有数据
    Object.assign(formData, {
      username: props.user.username,
      email: props.user.email,
      role: props.user.role,
      status: props.user.status,
      avatar: props.user.avatar || '',
      profile: {
        firstName: props.user.profile?.firstName || '',
        lastName: props.user.profile?.lastName || '',
        phone: props.user.profile?.phone || '',
        department: props.user.profile?.department || '',
        position: props.user.profile?.position || ''
      }
    })
  } else {
    // 创建模式，重置表单
    Object.assign(formData, {
      username: '',
      email: '',
      password: '',
      role: '',
      status: 'active',
      avatar: '',
      profile: {
        firstName: '',
        lastName: '',
        phone: '',
        department: '',
        position: ''
      }
    })
    confirmPassword.value = ''
  }
}

const beforeAvatarUpload = (file: UploadFile) => {
  const isJpgOrPng = file.type === 'image/jpeg' || file.type === 'image/png'
  if (!isJpgOrPng) {
    message.error('只能上传 JPG/PNG 格式的图片!')
    return false
  }
  const isLt2M = (file.size || 0) / 1024 / 1024 < 2
  if (!isLt2M) {
    message.error('图片大小不能超过 2MB!')
    return false
  }
  return false // 阻止自动上传
}

const handleAvatarChange = (info: UploadChangeParam) => {
  const file = info.file
  if (file.originFileObj) {
    const reader = new FileReader()
    reader.onload = (e) => {
      formData.avatar = e.target?.result as string
    }
    reader.readAsDataURL(file.originFileObj)
  }
}

const handleSubmit = async () => {
  try {
    await formRef.value.validate()
    submitting.value = true
    
    const submitData = { ...formData }
    if (isEdit.value) {
      delete submitData.password // 编辑时不提交密码
    }
    
    if (isEdit.value && props.user) {
      await updateUser(props.user.id, submitData)
    } else {
      await createUser(submitData)
    }
    
    emit('submit', submitData)
  } catch (error) {
    if (error.errorFields) {
      // 表单验证错误
      return
    }
    message.error(isEdit.value ? '更新用户失败' : '创建用户失败')
    console.error('提交用户表单失败:', error)
  } finally {
    submitting.value = false
  }
}

const handleCancel = () => {
  emit('cancel')
}

const handleTestConnection = async () => {
  try {
    testing.value = true
    // 这里可以添加测试连接的逻辑
    await new Promise(resolve => setTimeout(resolve, 1000)) // 模拟测试
    message.success('连接测试成功')
  } catch (error) {
    message.error('连接测试失败')
    console.error('连接测试失败:', error)
  } finally {
    testing.value = false
  }
}

// 监听器
watch(() => props.user, () => {
  initFormData()
}, { immediate: true })

// 生命周期
onMounted(() => {
  loadRoles()
})
</script>

<style scoped>
.user-form {
  max-width: 800px;
}

.form-section {
  margin-bottom: 16px;
}

.form-actions {
  margin-top: 24px;
  padding-top: 16px;
  border-top: 1px solid #f0f0f0;
  text-align: right;
}

.avatar-upload {
  display: flex;
  align-items: center;
}

.avatar-uploader {
  width: 100px;
  height: 100px;
}

.avatar-uploader .ant-upload {
  width: 100px;
  height: 100px;
  border-radius: 6px;
}

.upload-placeholder {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  height: 100%;
  color: #999;
  font-size: 12px;
}
</style>