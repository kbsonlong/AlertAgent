<template>
  <div class="settings-page">
    <div class="page-header">
      <h1 class="page-title">
        <SettingOutlined />
        系统设置
      </h1>
      <p class="page-description">管理系统配置和参数设置</p>
    </div>

    <div class="settings-content">
      <a-row :gutter="24">
        <!-- 设置菜单 -->
        <a-col :span="6">
          <div class="settings-menu">
            <a-menu
              v-model:selectedKeys="selectedKeys"
              mode="vertical"
              @click="handleMenuClick"
            >
              <a-menu-item key="general">
                <GlobalOutlined />
                通用设置
              </a-menu-item>
              <a-menu-item key="alert">
                <BellOutlined />
                告警设置
              </a-menu-item>
              <a-menu-item key="notification">
                <MessageOutlined />
                通知设置
              </a-menu-item>
              <a-menu-item key="storage">
                <DatabaseOutlined />
                存储设置
              </a-menu-item>
              <a-menu-item key="security">
                <SafetyOutlined />
                安全设置
              </a-menu-item>
              <a-menu-item key="system">
                <DesktopOutlined />
                系统信息
              </a-menu-item>
            </a-menu>
          </div>
        </a-col>

        <!-- 设置内容 -->
        <a-col :span="18">
          <div class="settings-panel">
            <!-- 通用设置 -->
            <div v-if="activeTab === 'general'" class="setting-section">
              <h2 class="section-title">通用设置</h2>
              
              <a-form
                ref="generalFormRef"
                :model="generalSettings"
                layout="vertical"
                @finish="handleGeneralSubmit"
              >
                <a-row :gutter="16">
                  <a-col :span="12">
                    <a-form-item label="系统名称" name="systemName">
                      <a-input
                        v-model:value="generalSettings.systemName"
                        placeholder="请输入系统名称"
                      />
                    </a-form-item>
                  </a-col>
                  <a-col :span="12">
                    <a-form-item label="系统版本" name="version">
                      <a-input
                        v-model:value="generalSettings.version"
                        placeholder="请输入系统版本"
                        disabled
                      />
                    </a-form-item>
                  </a-col>
                </a-row>
                
                <a-form-item label="系统描述" name="description">
                  <a-textarea
                    v-model:value="generalSettings.description"
                    placeholder="请输入系统描述"
                    :rows="3"
                  />
                </a-form-item>
                
                <a-row :gutter="16">
                  <a-col :span="12">
                    <a-form-item label="时区" name="timezone">
                      <a-select
                        v-model:value="generalSettings.timezone"
                        placeholder="请选择时区"
                      >
                        <a-select-option value="Asia/Shanghai">Asia/Shanghai</a-select-option>
                        <a-select-option value="UTC">UTC</a-select-option>
                        <a-select-option value="America/New_York">America/New_York</a-select-option>
                        <a-select-option value="Europe/London">Europe/London</a-select-option>
                      </a-select>
                    </a-form-item>
                  </a-col>
                  <a-col :span="12">
                    <a-form-item label="语言" name="language">
                      <a-select
                        v-model:value="generalSettings.language"
                        placeholder="请选择语言"
                      >
                        <a-select-option value="zh-CN">简体中文</a-select-option>
                        <a-select-option value="en-US">English</a-select-option>
                      </a-select>
                    </a-form-item>
                  </a-col>
                </a-row>
                
                <a-form-item>
                  <a-button type="primary" html-type="submit" :loading="generalLoading">
                    保存设置
                  </a-button>
                </a-form-item>
              </a-form>
            </div>

            <!-- 告警设置 -->
            <div v-else-if="activeTab === 'alert'" class="setting-section">
              <h2 class="section-title">告警设置</h2>
              
              <a-form
                ref="alertFormRef"
                :model="alertSettings"
                layout="vertical"
                @finish="handleAlertSubmit"
              >
                <a-row :gutter="16">
                  <a-col :span="12">
                    <a-form-item label="默认评估间隔(秒)" name="defaultEvaluationInterval">
                      <a-input-number
                        v-model:value="alertSettings.defaultEvaluationInterval"
                        :min="1"
                        :max="3600"
                        style="width: 100%"
                      />
                    </a-form-item>
                  </a-col>
                  <a-col :span="12">
                    <a-form-item label="默认持续时间(秒)" name="defaultDuration">
                      <a-input-number
                        v-model:value="alertSettings.defaultDuration"
                        :min="0"
                        :max="86400"
                        style="width: 100%"
                      />
                    </a-form-item>
                  </a-col>
                </a-row>
                
                <a-row :gutter="16">
                  <a-col :span="12">
                    <a-form-item label="最大告警数量" name="maxAlerts">
                      <a-input-number
                        v-model:value="alertSettings.maxAlerts"
                        :min="1"
                        :max="10000"
                        style="width: 100%"
                      />
                    </a-form-item>
                  </a-col>
                  <a-col :span="12">
                    <a-form-item label="告警保留天数" name="retentionDays">
                      <a-input-number
                        v-model:value="alertSettings.retentionDays"
                        :min="1"
                        :max="365"
                        style="width: 100%"
                      />
                    </a-form-item>
                  </a-col>
                </a-row>
                
                <a-form-item label="自动解决" name="autoResolve">
                  <a-switch
                    v-model:checked="alertSettings.autoResolve"
                    checked-children="启用"
                    un-checked-children="禁用"
                  />
                  <div class="form-help">
                    启用后，当告警条件不再满足时自动解决告警
                  </div>
                </a-form-item>
                
                <a-form-item label="静默模式" name="silenceMode">
                  <a-switch
                    v-model:checked="alertSettings.silenceMode"
                    checked-children="启用"
                    un-checked-children="禁用"
                  />
                  <div class="form-help">
                    启用后，系统将不会发送任何告警通知
                  </div>
                </a-form-item>
                
                <a-form-item>
                  <a-button type="primary" html-type="submit" :loading="alertLoading">
                    保存设置
                  </a-button>
                </a-form-item>
              </a-form>
            </div>

            <!-- 通知设置 -->
            <div v-else-if="activeTab === 'notification'" class="setting-section">
              <h2 class="section-title">通知设置</h2>
              
              <a-form
                ref="notificationFormRef"
                :model="notificationSettings"
                layout="vertical"
                @finish="handleNotificationSubmit"
              >
                <a-row :gutter="16">
                  <a-col :span="12">
                    <a-form-item label="默认发送频率(分钟)" name="defaultFrequency">
                      <a-input-number
                        v-model:value="notificationSettings.defaultFrequency"
                        :min="1"
                        :max="1440"
                        style="width: 100%"
                      />
                    </a-form-item>
                  </a-col>
                  <a-col :span="12">
                    <a-form-item label="最大重试次数" name="maxRetries">
                      <a-input-number
                        v-model:value="notificationSettings.maxRetries"
                        :min="0"
                        :max="10"
                        style="width: 100%"
                      />
                    </a-form-item>
                  </a-col>
                </a-row>
                
                <a-row :gutter="16">
                  <a-col :span="12">
                    <a-form-item label="发送超时(秒)" name="sendTimeout">
                      <a-input-number
                        v-model:value="notificationSettings.sendTimeout"
                        :min="1"
                        :max="300"
                        style="width: 100%"
                      />
                    </a-form-item>
                  </a-col>
                  <a-col :span="12">
                    <a-form-item label="批量发送大小" name="batchSize">
                      <a-input-number
                        v-model:value="notificationSettings.batchSize"
                        :min="1"
                        :max="1000"
                        style="width: 100%"
                      />
                    </a-form-item>
                  </a-col>
                </a-row>
                
                <a-form-item label="启用队列" name="enableQueue">
                  <a-switch
                    v-model:checked="notificationSettings.enableQueue"
                    checked-children="启用"
                    un-checked-children="禁用"
                  />
                  <div class="form-help">
                    启用后，通知将通过队列异步发送
                  </div>
                </a-form-item>
                
                <a-form-item>
                  <a-button type="primary" html-type="submit" :loading="notificationLoading">
                    保存设置
                  </a-button>
                </a-form-item>
              </a-form>
            </div>

            <!-- 存储设置 -->
            <div v-else-if="activeTab === 'storage'" class="setting-section">
              <h2 class="section-title">存储设置</h2>
              
              <a-form
                ref="storageFormRef"
                :model="storageSettings"
                layout="vertical"
                @finish="handleStorageSubmit"
              >
                <a-form-item label="数据库类型" name="dbType">
                  <a-select
                    v-model:value="storageSettings.dbType"
                    placeholder="请选择数据库类型"
                    disabled
                  >
                    <a-select-option value="mysql">MySQL</a-select-option>
                    <a-select-option value="postgresql">PostgreSQL</a-select-option>
                    <a-select-option value="sqlite">SQLite</a-select-option>
                  </a-select>
                </a-form-item>
                
                <a-row :gutter="16">
                  <a-col :span="12">
                    <a-form-item label="最大连接数" name="maxConnections">
                      <a-input-number
                        v-model:value="storageSettings.maxConnections"
                        :min="1"
                        :max="1000"
                        style="width: 100%"
                      />
                    </a-form-item>
                  </a-col>
                  <a-col :span="12">
                    <a-form-item label="连接超时(秒)" name="connectionTimeout">
                      <a-input-number
                        v-model:value="storageSettings.connectionTimeout"
                        :min="1"
                        :max="300"
                        style="width: 100%"
                      />
                    </a-form-item>
                  </a-col>
                </a-row>
                
                <a-row :gutter="16">
                  <a-col :span="12">
                    <a-form-item label="数据保留天数" name="dataRetentionDays">
                      <a-input-number
                        v-model:value="storageSettings.dataRetentionDays"
                        :min="1"
                        :max="3650"
                        style="width: 100%"
                      />
                    </a-form-item>
                  </a-col>
                  <a-col :span="12">
                    <a-form-item label="备份间隔(小时)" name="backupInterval">
                      <a-input-number
                        v-model:value="storageSettings.backupInterval"
                        :min="1"
                        :max="168"
                        style="width: 100%"
                      />
                    </a-form-item>
                  </a-col>
                </a-row>
                
                <a-form-item label="自动清理" name="autoCleanup">
                  <a-switch
                    v-model:checked="storageSettings.autoCleanup"
                    checked-children="启用"
                    un-checked-children="禁用"
                  />
                  <div class="form-help">
                    启用后，系统将自动清理过期数据
                  </div>
                </a-form-item>
                
                <a-form-item>
                  <a-button type="primary" html-type="submit" :loading="storageLoading">
                    保存设置
                  </a-button>
                </a-form-item>
              </a-form>
            </div>

            <!-- 安全设置 -->
            <div v-else-if="activeTab === 'security'" class="setting-section">
              <h2 class="section-title">安全设置</h2>
              
              <a-form
                ref="securityFormRef"
                :model="securitySettings"
                layout="vertical"
                @finish="handleSecuritySubmit"
              >
                <a-row :gutter="16">
                  <a-col :span="12">
                    <a-form-item label="会话超时(分钟)" name="sessionTimeout">
                      <a-input-number
                        v-model:value="securitySettings.sessionTimeout"
                        :min="5"
                        :max="1440"
                        style="width: 100%"
                      />
                    </a-form-item>
                  </a-col>
                  <a-col :span="12">
                    <a-form-item label="最大登录尝试" name="maxLoginAttempts">
                      <a-input-number
                        v-model:value="securitySettings.maxLoginAttempts"
                        :min="1"
                        :max="10"
                        style="width: 100%"
                      />
                    </a-form-item>
                  </a-col>
                </a-row>
                
                <a-row :gutter="16">
                  <a-col :span="12">
                    <a-form-item label="锁定时间(分钟)" name="lockoutDuration">
                      <a-input-number
                        v-model:value="securitySettings.lockoutDuration"
                        :min="1"
                        :max="1440"
                        style="width: 100%"
                      />
                    </a-form-item>
                  </a-col>
                  <a-col :span="12">
                    <a-form-item label="密码最小长度" name="minPasswordLength">
                      <a-input-number
                        v-model:value="securitySettings.minPasswordLength"
                        :min="6"
                        :max="32"
                        style="width: 100%"
                      />
                    </a-form-item>
                  </a-col>
                </a-row>
                
                <a-form-item label="启用HTTPS" name="enableHttps">
                  <a-switch
                    v-model:checked="securitySettings.enableHttps"
                    checked-children="启用"
                    un-checked-children="禁用"
                  />
                  <div class="form-help">
                    启用后，系统将强制使用HTTPS协议
                  </div>
                </a-form-item>
                
                <a-form-item label="启用审计日志" name="enableAuditLog">
                  <a-switch
                    v-model:checked="securitySettings.enableAuditLog"
                    checked-children="启用"
                    un-checked-children="禁用"
                  />
                  <div class="form-help">
                    启用后，系统将记录所有用户操作
                  </div>
                </a-form-item>
                
                <a-form-item>
                  <a-button type="primary" html-type="submit" :loading="securityLoading">
                    保存设置
                  </a-button>
                </a-form-item>
              </a-form>
            </div>

            <!-- 系统信息 -->
            <div v-else-if="activeTab === 'system'" class="setting-section">
              <h2 class="section-title">系统信息</h2>
              
              <div class="system-info">
                <a-descriptions :column="2" bordered>
                  <a-descriptions-item label="系统版本">
                    {{ systemInfo.version }}
                  </a-descriptions-item>
                  <a-descriptions-item label="构建时间">
                    {{ systemInfo.buildTime }}
                  </a-descriptions-item>
                  <a-descriptions-item label="Go版本">
                    {{ systemInfo.goVersion }}
                  </a-descriptions-item>
                  <a-descriptions-item label="运行时间">
                    {{ systemInfo.uptime }}
                  </a-descriptions-item>
                  <a-descriptions-item label="CPU使用率">
                    <a-progress :percent="systemInfo.cpuUsage" size="small" />
                  </a-descriptions-item>
                  <a-descriptions-item label="内存使用率">
                    <a-progress :percent="systemInfo.memoryUsage" size="small" />
                  </a-descriptions-item>
                  <a-descriptions-item label="磁盘使用率">
                    <a-progress :percent="systemInfo.diskUsage" size="small" />
                  </a-descriptions-item>
                  <a-descriptions-item label="网络连接数">
                    {{ systemInfo.connections }}
                  </a-descriptions-item>
                </a-descriptions>
                
                <div class="system-actions">
                  <a-space>
                    <a-button @click="refreshSystemInfo" :loading="systemInfoLoading">
                      <ReloadOutlined /> 刷新信息
                    </a-button>
                    <a-button @click="exportSystemInfo">
                      <DownloadOutlined /> 导出信息
                    </a-button>
                  </a-space>
                </div>
              </div>
            </div>
          </div>
        </a-col>
      </a-row>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import {
  Row,
  Col,
  Menu,
  Form,
  Input,
  Select,
  Switch,
  Button,
  InputNumber,
  Descriptions,
  Progress,
  Space,
  message
} from 'ant-design-vue'
import {
  SettingOutlined,
  GlobalOutlined,
  BellOutlined,
  MessageOutlined,
  DatabaseOutlined,
  SafetyOutlined,
  DesktopOutlined,
  ReloadOutlined,
  DownloadOutlined
} from '@ant-design/icons-vue'
import {
  getSystemSettings,
  updateSystemSettings,
  getSystemInfo
} from '@/services/system'

const ARow = Row
const ACol = Col
const AMenu = Menu
const AMenuItem = Menu.Item
const AForm = Form
const AFormItem = Form.Item
const AInput = Input
const ATextarea = Input.TextArea
const ASelect = Select
const ASelectOption = Select.Option
const ASwitch = Switch
const AButton = Button
const AInputNumber = InputNumber
const ADescriptions = Descriptions
const ADescriptionsItem = Descriptions.Item
const AProgress = Progress
const ASpace = Space

// 响应式数据
const selectedKeys = ref(['general'])
const activeTab = ref('general')

// 表单引用
const generalFormRef = ref()
const alertFormRef = ref()
const notificationFormRef = ref()
const storageFormRef = ref()
const securityFormRef = ref()

// 加载状态
const generalLoading = ref(false)
const alertLoading = ref(false)
const notificationLoading = ref(false)
const storageLoading = ref(false)
const securityLoading = ref(false)
const systemInfoLoading = ref(false)

// 设置数据
const generalSettings = reactive({
  systemName: 'AlertAgent',
  version: '1.0.0',
  description: '智能告警管理系统',
  timezone: 'Asia/Shanghai',
  language: 'zh-CN'
})

const alertSettings = reactive({
  defaultEvaluationInterval: 60,
  defaultDuration: 300,
  maxAlerts: 1000,
  retentionDays: 30,
  autoResolve: true,
  silenceMode: false
})

const notificationSettings = reactive({
  defaultFrequency: 5,
  maxRetries: 3,
  sendTimeout: 30,
  batchSize: 100,
  enableQueue: true
})

const storageSettings = reactive({
  dbType: 'mysql',
  maxConnections: 100,
  connectionTimeout: 30,
  dataRetentionDays: 90,
  backupInterval: 24,
  autoCleanup: true
})

const securitySettings = reactive({
  sessionTimeout: 30,
  maxLoginAttempts: 5,
  lockoutDuration: 15,
  minPasswordLength: 8,
  enableHttps: true,
  enableAuditLog: true
})

const systemInfo = reactive({
  version: '1.0.0',
  buildTime: '2024-01-01 00:00:00',
  goVersion: 'go1.21.0',
  uptime: '0天0小时0分钟',
  cpuUsage: 0,
  memoryUsage: 0,
  diskUsage: 0,
  connections: 0
})

// 菜单点击处理
const handleMenuClick = ({ key }: { key: string }) => {
  activeTab.value = key
  selectedKeys.value = [key]
}

// 通用设置提交
const handleGeneralSubmit = async () => {
  try {
    generalLoading.value = true
    await updateSystemSettings('general', generalSettings)
    message.success('通用设置保存成功')
  } catch (error) {
    console.error('保存通用设置失败:', error)
    message.error('保存通用设置失败')
  } finally {
    generalLoading.value = false
  }
}

// 告警设置提交
const handleAlertSubmit = async () => {
  try {
    alertLoading.value = true
    await updateSystemSettings('alert', alertSettings)
    message.success('告警设置保存成功')
  } catch (error) {
    console.error('保存告警设置失败:', error)
    message.error('保存告警设置失败')
  } finally {
    alertLoading.value = false
  }
}

// 通知设置提交
const handleNotificationSubmit = async () => {
  try {
    notificationLoading.value = true
    await updateSystemSettings('notification', notificationSettings)
    message.success('通知设置保存成功')
  } catch (error) {
    console.error('保存通知设置失败:', error)
    message.error('保存通知设置失败')
  } finally {
    notificationLoading.value = false
  }
}

// 存储设置提交
const handleStorageSubmit = async () => {
  try {
    storageLoading.value = true
    await updateSystemSettings('storage', storageSettings)
    message.success('存储设置保存成功')
  } catch (error) {
    console.error('保存存储设置失败:', error)
    message.error('保存存储设置失败')
  } finally {
    storageLoading.value = false
  }
}

// 安全设置提交
const handleSecuritySubmit = async () => {
  try {
    securityLoading.value = true
    await updateSystemSettings('security', securitySettings)
    message.success('安全设置保存成功')
  } catch (error) {
    console.error('保存安全设置失败:', error)
    message.error('保存安全设置失败')
  } finally {
    securityLoading.value = false
  }
}

// 刷新系统信息
const refreshSystemInfo = async () => {
  try {
    systemInfoLoading.value = true
    const info = await getSystemInfo()
    Object.assign(systemInfo, info)
  } catch (error) {
    console.error('获取系统信息失败:', error)
    message.error('获取系统信息失败')
  } finally {
    systemInfoLoading.value = false
  }
}

// 导出系统信息
const exportSystemInfo = () => {
  const data = JSON.stringify(systemInfo, null, 2)
  const blob = new Blob([data], { type: 'application/json' })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = 'system-info.json'
  a.click()
  URL.revokeObjectURL(url)
}

// 加载设置数据
const loadSettings = async () => {
  try {
    const settings = await getSystemSettings()
    
    if (settings.general) {
      Object.assign(generalSettings, settings.general)
    }
    if (settings.alert) {
      Object.assign(alertSettings, settings.alert)
    }
    if (settings.notification) {
      Object.assign(notificationSettings, settings.notification)
    }
    if (settings.storage) {
      Object.assign(storageSettings, settings.storage)
    }
    if (settings.security) {
      Object.assign(securitySettings, settings.security)
    }
  } catch (error) {
    console.error('加载设置失败:', error)
    message.error('加载设置失败')
  }
}

// 组件挂载
onMounted(() => {
  loadSettings()
  refreshSystemInfo()
})
</script>

<style scoped>
.settings-page {
  padding: 24px;
}

.page-header {
  margin-bottom: 24px;
}

.page-title {
  display: flex;
  align-items: center;
  gap: 12px;
  margin: 0 0 8px 0;
  font-size: 24px;
  font-weight: 600;
  color: #262626;
}

.page-description {
  margin: 0;
  color: #8c8c8c;
  font-size: 14px;
}

.settings-content {
  background: white;
  border-radius: 8px;
  overflow: hidden;
}

.settings-menu {
  background: #fafafa;
  border-right: 1px solid #e8e8e8;
  min-height: 600px;
}

.settings-menu :deep(.ant-menu) {
  background: transparent;
  border: none;
}

.settings-menu :deep(.ant-menu-item) {
  display: flex;
  align-items: center;
  gap: 8px;
  margin: 0;
  border-radius: 0;
}

.settings-panel {
  padding: 24px;
  min-height: 600px;
}

.setting-section {
  max-width: 800px;
}

.section-title {
  margin: 0 0 24px 0;
  font-size: 20px;
  font-weight: 600;
  color: #262626;
  border-bottom: 1px solid #e8e8e8;
  padding-bottom: 12px;
}

.form-help {
  margin-top: 4px;
  font-size: 12px;
  color: #8c8c8c;
}

.system-info {
  display: flex;
  flex-direction: column;
  gap: 24px;
}

.system-actions {
  display: flex;
  justify-content: flex-end;
}

/* 响应式设计 */
@media (max-width: 768px) {
  .settings-page {
    padding: 16px;
  }
  
  .settings-content :deep(.ant-col) {
    width: 100% !important;
    flex: none !important;
  }
  
  .settings-menu {
    margin-bottom: 16px;
    min-height: auto;
  }
  
  .settings-menu :deep(.ant-menu) {
    display: flex;
    overflow-x: auto;
  }
  
  .settings-menu :deep(.ant-menu-item) {
    white-space: nowrap;
  }
  
  .settings-panel {
    padding: 16px;
    min-height: auto;
  }
}
</style>