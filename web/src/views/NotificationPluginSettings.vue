<template>
  <div class="notification-plugin-settings">
    <div class="page-header">
      <div class="header-content">
        <h1>通知插件设置</h1>
        <p>管理和配置系统通知插件，支持多种通知渠道</p>
      </div>
      <div class="header-actions">
        <a-button @click="refreshPlugins" :loading="refreshing">
          <ReloadOutlined /> 刷新插件
        </a-button>
        <a-button type="primary" @click="openBatchConfig">
          <SettingOutlined /> 批量配置
        </a-button>
      </div>
    </div>

    <div class="page-content">
      <a-tabs v-model:activeKey="activeTab" type="card">
        <a-tab-pane key="plugins" tab="插件配置">
          <NotificationPluginConfig ref="pluginConfigRef" />
        </a-tab-pane>
        <a-tab-pane key="user" tab="用户管理">
          <UserNotificationManager />
        </a-tab-pane>
      </a-tabs>
    </div>

    <!-- 批量配置模态框 -->
    <a-modal
      v-model:open="batchConfigVisible"
      title="批量插件配置"
      width="1000px"
      :footer="null"
    >
      <div class="batch-config-content">
        <a-tabs v-model:activeKey="batchConfigTab">
          <a-tab-pane key="export" tab="导出配置">
            <div class="export-section">
              <p>导出当前插件配置，可用于备份或迁移到其他环境</p>
              <a-button type="primary" @click="exportConfigs" :loading="exportLoading">
                <DownloadOutlined /> 导出配置
              </a-button>
              <div v-if="exportedConfig" class="export-result">
                <h4>导出结果：</h4>
                <a-textarea
                  :value="exportedConfig"
                  :rows="10"
                  readonly
                  class="export-textarea"
                />
                <div class="export-actions">
                  <a-button @click="copyToClipboard">
                    <CopyOutlined /> 复制到剪贴板
                  </a-button>
                  <a-button @click="downloadConfig">
                    <DownloadOutlined /> 下载文件
                  </a-button>
                </div>
              </div>
            </div>
          </a-tab-pane>
          
          <a-tab-pane key="import" tab="导入配置">
            <div class="import-section">
              <p>导入插件配置，支持JSON格式的配置文件</p>
              <a-upload
                :before-upload="handleImportFile"
                :show-upload-list="false"
                accept=".json"
              >
                <a-button>
                  <UploadOutlined /> 选择配置文件
                </a-button>
              </a-upload>
              <div class="import-textarea-section">
                <p>或者直接粘贴配置内容：</p>
                <a-textarea
                  v-model:value="importConfig"
                  placeholder="请粘贴JSON格式的配置内容"
                  :rows="8"
                />
                <div class="import-actions">
                  <a-button @click="validateImportConfig">
                    <CheckOutlined /> 验证配置
                  </a-button>
                  <a-button 
                    type="primary" 
                    @click="importConfigs" 
                    :loading="importLoading"
                    :disabled="!importConfig"
                  >
                    <ImportOutlined /> 导入配置
                  </a-button>
                </div>
              </div>
            </div>
          </a-tab-pane>
          
          <a-tab-pane key="health" tab="健康检查">
            <div class="health-section">
              <p>检查所有插件的健康状态</p>
              <a-button type="primary" @click="checkAllPluginsHealth" :loading="healthCheckLoading">
                <HeartOutlined /> 执行健康检查
              </a-button>
              <div v-if="healthCheckResults" class="health-results">
                <h4>检查结果：</h4>
                <a-list
                  :data-source="Object.entries(healthCheckResults)"
                  :split="false"
                >
                  <template #renderItem="{ item }">
                    <a-list-item>
                      <a-list-item-meta>
                        <template #title>
                          <span>{{ getPluginDisplayName(item[0]) }}</span>
                          <a-tag 
                            :color="item[1].status === 'healthy' ? 'green' : 'red'"
                            style="margin-left: 8px"
                          >
                            {{ item[1].status === 'healthy' ? '健康' : '异常' }}
                          </a-tag>
                        </template>
                        <template #description>
                          <div v-if="item[1].error" class="health-error">
                            {{ item[1].error }}
                          </div>
                          <div class="health-time">
                            最后检查: {{ formatTime(item[1].last_check) }}
                          </div>
                        </template>
                      </a-list-item-meta>
                    </a-list-item>
                  </template>
                </a-list>
              </div>
            </div>
          </a-tab-pane>
        </a-tabs>
      </div>
    </a-modal>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive } from 'vue'
import { message } from 'ant-design-vue'
import {
  Button,
  Modal,
  Tabs,
  Textarea,
  Upload,
  List,
  Tag
} from 'ant-design-vue'
import {
  ReloadOutlined,
  SettingOutlined,
  DownloadOutlined,
  CopyOutlined,
  UploadOutlined,
  CheckOutlined,
  ImportOutlined,
  HeartOutlined
} from '@ant-design/icons-vue'
import {
  getAllPluginHealthStatus,
  type PluginHealthStatus
} from '@/services/plugin'
import NotificationPluginConfig from '@/components/NotificationPluginConfig.vue'
import UserNotificationManager from '@/components/UserNotificationManager.vue'
import { formatTime } from '@/utils/datetime'

const AButton = Button
const AModal = Modal
const ATabs = Tabs
const ATabPane = Tabs.TabPane
const ATextarea = Textarea
const AUpload = Upload
const AList = List
const AListItem = List.Item
const AListItemMeta = List.Item.Meta
const ATag = Tag

// 响应式数据
const activeTab = ref('plugins')
const refreshing = ref(false)
const batchConfigVisible = ref(false)
const batchConfigTab = ref('export')
const exportLoading = ref(false)
const importLoading = ref(false)
const healthCheckLoading = ref(false)
const exportedConfig = ref('')
const importConfig = ref('')
const healthCheckResults = ref<Record<string, PluginHealthStatus> | null>(null)

// 组件引用
const pluginConfigRef = ref()

// 获取插件显示名称
const getPluginDisplayName = (pluginName: string) => {
  const nameMap: Record<string, string> = {
    'email': '邮件通知',
    'dingtalk': '钉钉通知',
    'wechat': '企业微信',
    'slack': 'Slack',
    'webhook': 'Webhook'
  }
  return nameMap[pluginName] || pluginName
}

// 刷新插件
const refreshPlugins = async () => {
  try {
    refreshing.value = true
    if (pluginConfigRef.value) {
      await pluginConfigRef.value.loadAvailablePlugins()
    }
    message.success('插件列表已刷新')
  } catch (error) {
    console.error('刷新插件失败:', error)
    message.error('刷新插件失败')
  } finally {
    refreshing.value = false
  }
}

// 打开批量配置
const openBatchConfig = () => {
  batchConfigVisible.value = true
}

// 导出配置
const exportConfigs = async () => {
  try {
    exportLoading.value = true
    
    // 这里应该调用API获取所有插件配置
    // 暂时使用模拟数据
    const configs = {
      version: '1.0.0',
      timestamp: new Date().toISOString(),
      plugins: {
        email: {
          enabled: true,
          config: {
            smtp: {
              host: 'smtp.example.com',
              port: 587,
              username: 'user@example.com',
              password: '***'
            }
          },
          priority: 1
        }
      }
    }
    
    exportedConfig.value = JSON.stringify(configs, null, 2)
    message.success('配置导出成功')
  } catch (error) {
    console.error('导出配置失败:', error)
    message.error('导出配置失败')
  } finally {
    exportLoading.value = false
  }
}

// 复制到剪贴板
const copyToClipboard = async () => {
  try {
    await navigator.clipboard.writeText(exportedConfig.value)
    message.success('已复制到剪贴板')
  } catch (error) {
    console.error('复制失败:', error)
    message.error('复制失败')
  }
}

// 下载配置文件
const downloadConfig = () => {
  const blob = new Blob([exportedConfig.value], { type: 'application/json' })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = `notification-plugins-config-${new Date().toISOString().split('T')[0]}.json`
  document.body.appendChild(a)
  a.click()
  document.body.removeChild(a)
  URL.revokeObjectURL(url)
  message.success('配置文件已下载')
}

// 处理导入文件
const handleImportFile = (file: File) => {
  const reader = new FileReader()
  reader.onload = (e) => {
    importConfig.value = e.target?.result as string
  }
  reader.readAsText(file)
  return false // 阻止自动上传
}

// 验证导入配置
const validateImportConfig = () => {
  try {
    JSON.parse(importConfig.value)
    message.success('配置格式验证通过')
  } catch (error) {
    message.error('配置格式错误，请检查JSON格式')
  }
}

// 导入配置
const importConfigs = async () => {
  try {
    importLoading.value = true
    
    // 验证配置格式
    const config = JSON.parse(importConfig.value)
    
    // 这里应该调用API导入配置
    // 暂时模拟导入过程
    await new Promise(resolve => setTimeout(resolve, 2000))
    
    message.success('配置导入成功')
    importConfig.value = ''
    batchConfigVisible.value = false
    
    // 刷新插件列表
    await refreshPlugins()
  } catch (error) {
    console.error('导入配置失败:', error)
    message.error('导入配置失败')
  } finally {
    importLoading.value = false
  }
}

// 检查所有插件健康状态
const checkAllPluginsHealth = async () => {
  try {
    healthCheckLoading.value = true
    const results = await getAllPluginHealthStatus()
    healthCheckResults.value = results
    message.success('健康检查完成')
  } catch (error) {
    console.error('健康检查失败:', error)
    message.error('健康检查失败')
  } finally {
    healthCheckLoading.value = false
  }
}
</script>

<style scoped>
.notification-plugin-settings {
  padding: 24px;
  min-height: 100vh;
  background: #f5f5f5;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 24px;
  padding: 24px;
  background: white;
  border-radius: 8px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}

.header-content h1 {
  margin: 0 0 8px 0;
  font-size: 28px;
  font-weight: 600;
  color: #262626;
}

.header-content p {
  margin: 0;
  color: #8c8c8c;
  font-size: 14px;
}

.header-actions {
  display: flex;
  gap: 12px;
}

.page-content {
  background: white;
  border-radius: 8px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}

.batch-config-content {
  padding: 0;
}

.export-section,
.import-section,
.health-section {
  padding: 20px 0;
}

.export-result {
  margin-top: 20px;
}

.export-result h4 {
  margin: 0 0 12px 0;
  font-size: 16px;
  font-weight: 600;
}

.export-textarea {
  margin-bottom: 12px;
  font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
  font-size: 12px;
}

.export-actions,
.import-actions {
  display: flex;
  gap: 12px;
}

.import-textarea-section {
  margin-top: 20px;
}

.import-textarea-section p {
  margin: 0 0 12px 0;
  color: #595959;
}

.import-actions {
  margin-top: 12px;
}

.health-results {
  margin-top: 20px;
}

.health-results h4 {
  margin: 0 0 16px 0;
  font-size: 16px;
  font-weight: 600;
}

.health-error {
  color: #ff4d4f;
  font-size: 12px;
  margin-bottom: 4px;
}

.health-time {
  color: #8c8c8c;
  font-size: 12px;
}

/* 响应式设计 */
@media (max-width: 768px) {
  .notification-plugin-settings {
    padding: 16px;
  }
  
  .page-header {
    flex-direction: column;
    gap: 16px;
    padding: 16px;
  }
  
  .header-actions {
    width: 100%;
    justify-content: center;
  }
  
  .export-actions,
  .import-actions {
    flex-direction: column;
  }
}
</style>