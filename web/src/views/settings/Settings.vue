<template>
  <div class="settings-container">
    <a-card title="系统设置" :bordered="false">
      <a-tabs v-model:activeKey="activeTab" type="card">
        <a-tab-pane key="general" tab="常规设置">
          <a-form
            :model="generalSettings"
            :label-col="{ span: 6 }"
            :wrapper-col="{ span: 18 }"
            layout="horizontal"
          >
            <a-form-item label="系统名称">
              <a-input v-model:value="generalSettings.systemName" placeholder="请输入系统名称" />
            </a-form-item>
            <a-form-item label="系统描述">
              <a-textarea v-model:value="generalSettings.description" :rows="3" placeholder="请输入系统描述" />
            </a-form-item>
            <a-form-item label="时区">
              <a-select v-model:value="generalSettings.timezone" placeholder="请选择时区">
                <a-select-option value="Asia/Shanghai">Asia/Shanghai</a-select-option>
                <a-select-option value="UTC">UTC</a-select-option>
                <a-select-option value="America/New_York">America/New_York</a-select-option>
              </a-select>
            </a-form-item>
            <a-form-item label="语言">
              <a-select v-model:value="generalSettings.language" placeholder="请选择语言">
                <a-select-option value="zh-CN">中文</a-select-option>
                <a-select-option value="en-US">English</a-select-option>
              </a-select>
            </a-form-item>
          </a-form>
        </a-tab-pane>
        
        <a-tab-pane key="notification" tab="通知设置">
          <a-form
            :model="notificationSettings"
            :label-col="{ span: 6 }"
            :wrapper-col="{ span: 18 }"
            layout="horizontal"
          >
            <a-form-item label="邮件通知">
              <a-switch v-model:checked="notificationSettings.emailEnabled" />
            </a-form-item>
            <a-form-item label="短信通知">
              <a-switch v-model:checked="notificationSettings.smsEnabled" />
            </a-form-item>
            <a-form-item label="微信通知">
              <a-switch v-model:checked="notificationSettings.wechatEnabled" />
            </a-form-item>
            <a-form-item label="通知频率">
              <a-select v-model:value="notificationSettings.frequency" placeholder="请选择通知频率">
                <a-select-option value="immediate">立即</a-select-option>
                <a-select-option value="hourly">每小时</a-select-option>
                <a-select-option value="daily">每日</a-select-option>
              </a-select>
            </a-form-item>
          </a-form>
        </a-tab-pane>
        
        <a-tab-pane key="security" tab="安全设置">
          <a-form
            :model="securitySettings"
            :label-col="{ span: 6 }"
            :wrapper-col="{ span: 18 }"
            layout="horizontal"
          >
            <a-form-item label="会话超时">
              <a-input-number 
                v-model:value="securitySettings.sessionTimeout" 
                :min="1" 
                :max="1440" 
                addon-after="分钟"
              />
            </a-form-item>
            <a-form-item label="密码复杂度">
              <a-switch v-model:checked="securitySettings.passwordComplexity" />
            </a-form-item>
            <a-form-item label="双因子认证">
              <a-switch v-model:checked="securitySettings.twoFactorAuth" />
            </a-form-item>
            <a-form-item label="登录日志">
              <a-switch v-model:checked="securitySettings.loginLog" />
            </a-form-item>
          </a-form>
        </a-tab-pane>
      </a-tabs>
      
      <div class="settings-actions">
        <a-space>
          <a-button @click="handleReset">重置</a-button>
          <a-button type="primary" @click="handleSave" :loading="saving">
            保存设置
          </a-button>
        </a-space>
      </div>
    </a-card>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { message } from 'ant-design-vue'

// 响应式数据
const activeTab = ref('general')
const saving = ref(false)

// 设置数据
const generalSettings = reactive({
  systemName: 'AlertAgent',
  description: '智能告警监控系统',
  timezone: 'Asia/Shanghai',
  language: 'zh-CN'
})

const notificationSettings = reactive({
  emailEnabled: true,
  smsEnabled: false,
  wechatEnabled: true,
  frequency: 'immediate'
})

const securitySettings = reactive({
  sessionTimeout: 30,
  passwordComplexity: true,
  twoFactorAuth: false,
  loginLog: true
})

// 方法
const handleSave = async () => {
  try {
    saving.value = true
    
    // 模拟保存API调用
    await new Promise(resolve => setTimeout(resolve, 1000))
    
    message.success('设置保存成功')
  } catch (error) {
    console.error('保存设置失败:', error)
    message.error('保存设置失败')
  } finally {
    saving.value = false
  }
}

const handleReset = () => {
  // 重置为默认值
  Object.assign(generalSettings, {
    systemName: 'AlertAgent',
    description: '智能告警监控系统',
    timezone: 'Asia/Shanghai',
    language: 'zh-CN'
  })
  
  Object.assign(notificationSettings, {
    emailEnabled: true,
    smsEnabled: false,
    wechatEnabled: true,
    frequency: 'immediate'
  })
  
  Object.assign(securitySettings, {
    sessionTimeout: 30,
    passwordComplexity: true,
    twoFactorAuth: false,
    loginLog: true
  })
  
  message.info('设置已重置为默认值')
}

// 生命周期
onMounted(() => {
  // 加载设置数据
  console.log('加载系统设置')
})
</script>

<style scoped>
.settings-container {
  padding: 24px;
}

.settings-actions {
  margin-top: 24px;
  padding-top: 24px;
  border-top: 1px solid #f0f0f0;
  text-align: right;
}

:deep(.ant-tabs-content-holder) {
  padding: 24px 0;
}

:deep(.ant-form-item) {
  margin-bottom: 24px;
}

:deep(.ant-card-head-title) {
  font-size: 18px;
  font-weight: 600;
}
</style>