<template>
  <div class="login-container">
    <div class="login-box">
      <div class="login-header">
        <h1>告警管理系统</h1>
        <p>请登录您的账户</p>
      </div>
      
      <a-form
        :model="loginForm"
        :rules="rules"
        @finish="handleLogin"
        @finishFailed="handleLoginFailed"
        layout="vertical"
        class="login-form"
      >
        <a-form-item label="用户名" name="username">
          <a-input
          v-model:value="loginForm.username"
          placeholder="请输入用户名"
          size="large"
          :prefix="h(UserOutlined)"
          autocomplete="username"
        />
        </a-form-item>
        
        <a-form-item label="密码" name="password">
          <a-input-password
            v-model:value="loginForm.password"
            placeholder="请输入密码"
            size="large"
            :prefix="h(LockOutlined)"
            autocomplete="current-password"
          />
        </a-form-item>
        
        <a-form-item>
          <a-checkbox v-model:checked="loginForm.remember">
            记住我
          </a-checkbox>
        </a-form-item>
        
        <a-form-item>
          <a-button
            type="primary"
            html-type="submit"
            size="large"
            :loading="loading"
            block
          >
            登录
          </a-button>
        </a-form-item>
      </a-form>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, h } from 'vue'
import { useRouter } from 'vue-router'
import { message } from 'ant-design-vue'
import { UserOutlined, LockOutlined } from '@ant-design/icons-vue'
import { useUserStore } from '@/stores'
import { UserService } from '@/services/userService'
import type { Rule } from 'ant-design-vue/es/form'

const router = useRouter()
const userStore = useUserStore()
const loading = ref(false)

// 登录表单数据
const loginForm = ref({
  username: '',
  password: '',
  remember: false
})

// 表单验证规则
const rules: Record<string, Rule[]> = {
  username: [
    { required: true, message: '请输入用户名', trigger: 'blur' },
    { min: 3, max: 20, message: '用户名长度应为3-20个字符', trigger: 'blur' }
  ],
  password: [
    { required: true, message: '请输入密码', trigger: 'blur' },
    { min: 6, max: 20, message: '密码长度应为6-20个字符', trigger: 'blur' }
  ]
}

// 处理登录
const handleLogin = async (values: any) => {
  loading.value = true
  try {
    // 调用真实的登录API
    const response = await UserService.login({
      username: values.username,
      password: values.password,
      remember_me: values.remember
    })
    
    if (response.code === 200 && response.data) {
      // 设置用户信息和token
      userStore.setToken(response.data.access_token)
      userStore.setUser(response.data.user)
      
      message.success('登录成功')
      
      // 跳转到首页
      router.push('/')
    } else {
      message.error(response.message || '登录失败')
    }
  } catch (error: any) {
    console.error('登录失败:', error)
    const errorMessage = error?.response?.data?.message || error?.message || '登录失败，请检查用户名和密码'
    message.error(errorMessage)
  } finally {
    loading.value = false
  }
}

// 处理登录失败
const handleLoginFailed = (errorInfo: any) => {
  console.log('登录表单验证失败:', errorInfo)
}
</script>

<style scoped>
.login-container {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  padding: 20px;
}

.login-box {
  width: 100%;
  max-width: 400px;
  background: white;
  border-radius: 12px;
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.1);
  padding: 40px;
}

.login-header {
  text-align: center;
  margin-bottom: 32px;
}

.login-header h1 {
  color: #1890ff;
  font-size: 28px;
  font-weight: 600;
  margin-bottom: 8px;
}

.login-header p {
  color: #666;
  font-size: 14px;
  margin: 0;
}

.login-form {
  margin-top: 24px;
}

:deep(.ant-form-item-label > label) {
  font-weight: 500;
}

:deep(.ant-input-affix-wrapper) {
  border-radius: 8px;
}

:deep(.ant-btn) {
  border-radius: 8px;
  height: 44px;
  font-weight: 500;
}

@media (max-width: 480px) {
  .login-box {
    padding: 24px;
  }
  
  .login-header h1 {
    font-size: 24px;
  }
}
</style>