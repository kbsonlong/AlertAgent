<template>
  <div class="notification-plugin-config">
    <div class="config-header">
      <h2>通知插件配置</h2>
      <p>配置和管理通知渠道插件，支持多种通知方式</p>
    </div>

    <a-spin :spinning="loading">
      <div class="plugin-list">
        <a-row :gutter="[16, 16]">
          <a-col 
            v-for="plugin in availablePlugins" 
            :key="plugin.name"
            :xs="24" 
            :sm="12" 
            :lg="8"
          >
            <a-card 
              :class="['plugin-card', { 'plugin-configured': isPluginConfigured(plugin.name) }]"
              hoverable
            >
              <template #title>
                <div class="plugin-title">
                  <component 
                    :is="getPluginIcon(plugin.name)" 
                    class="plugin-icon"
                  />
                  <span>{{ getPluginDisplayName(plugin.name) }}</span>
                  <a-tag 
                    :color="getPluginStatusColor(plugin.status)"
                    class="plugin-status"
                  >
                    {{ getPluginStatusText(plugin.status) }}
                  </a-tag>
                </div>
              </template>

              <template #extra>
                <a-switch
                  :checked="isPluginEnabled(plugin.name)"
                  :loading="switchLoading[plugin.name]"
                  @change="(checked) => handlePluginToggle(plugin.name, checked)"
                />
              </template>

              <div class="plugin-content">
                <p class="plugin-description">{{ plugin.description }}</p>
                <div class="plugin-meta">
                  <span class="plugin-version">版本: {{ plugin.version }}</span>
                  <span class="plugin-load-time">
                    加载时间: {{ formatTime(plugin.load_time) }}
                  </span>
                </div>
                
                <div v-if="plugin.last_error" class="plugin-error">
                  <a-alert
                    type="error"
                    :message="plugin.last_error"
                    show-icon
                    banner
                  />
                </div>
              </div>

              <template #actions>
                <a-button 
                  type="primary" 
                  size="small"
                  @click="openConfigModal(plugin)"
                >
                  <SettingOutlined /> 配置
                </a-button>
                <a-button 
                  size="small"
                  :disabled="!isPluginConfigured(plugin.name)"
                  @click="openTestModal(plugin.name)"
                >
                  <ExperimentOutlined /> 测试
                </a-button>
                <a-button 
                  size="small"
                  @click="viewPluginStats(plugin.name)"
                >
                  <BarChartOutlined /> 统计
                </a-button>
              </template>
            </a-card>
          </a-col>
        </a-row>
      </div>
    </a-spin>

    <!-- 插件配置模态框 -->
    <a-modal
      v-model:open="configModalVisible"
      :title="`配置 ${currentPlugin?.name || ''} 插件`"
      width="800px"
      :footer="null"
      :destroyOnClose="true"
    >
      <NotificationPluginConfigForm
        v-if="currentPlugin"
        :plugin="currentPlugin"
        :config="currentPluginConfig"
        @submit="handleConfigSubmit"
        @cancel="closeConfigModal"
      />
    </a-modal>

    <!-- 插件统计模态框 -->
    <a-modal
      v-model:open="statsModalVisible"
      :title="`${currentStatsPlugin} 插件统计`"
      width="600px"
      :footer="null"
    >
      <NotificationPluginStats
        v-if="currentStatsPlugin"
        :plugin-name="currentStatsPlugin"
        :stats="currentStats"
      />
    </a-modal>

    <!-- 插件测试模态框 -->
    <NotificationTestModal
      v-model:visible="testModalVisible"
      :plugin-name="currentTestPlugin"
      :plugin-config="currentTestConfig"
      @test-completed="handleTestCompleted"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { message } from 'ant-design-vue'
import {
  Card,
  Row,
  Col,
  Tag,
  Switch,
  Button,
  Modal,
  Spin,
  Alert
} from 'ant-design-vue'
import {
  SettingOutlined,
  ExperimentOutlined,
  BarChartOutlined,
  MailOutlined,
  ApiOutlined,
  MessageOutlined,
  WechatOutlined,
  SlackOutlined,
  BellOutlined
} from '@ant-design/icons-vue'
import {
  getAvailablePlugins,
  getPluginConfig,
  setPluginConfig,
  testPluginConfig,
  getPluginStats,
  enablePlugin,
  disablePlugin,
  type PluginInfo,
  type PluginConfig,
  type PluginStats
} from '@/services/plugin'
import NotificationPluginConfigForm from './NotificationPluginConfigForm.vue'
import NotificationPluginStats from './NotificationPluginStats.vue'
import NotificationTestModal from './NotificationTestModal.vue'
import { formatTime } from '@/utils/datetime'

const ACard = Card
const ARow = Row
const ACol = Col
const ATag = Tag
const ASwitch = Switch
const AButton = Button
const AModal = Modal
const ASpin = Spin
const AAlert = Alert

// 响应式数据
const loading = ref(false)
const availablePlugins = ref<PluginInfo[]>([])
const pluginConfigs = ref<Record<string, PluginConfig>>({})
const switchLoading = reactive<Record<string, boolean>>({})
const testLoading = reactive<Record<string, boolean>>({})

// 模态框状态
const configModalVisible = ref(false)
const statsModalVisible = ref(false)
const testModalVisible = ref(false)
const currentPlugin = ref<PluginInfo | null>(null)
const currentPluginConfig = ref<PluginConfig | null>(null)
const currentStatsPlugin = ref<string>('')
const currentStats = ref<PluginStats | null>(null)
const currentTestPlugin = ref<string>('')
const currentTestConfig = ref<Record<string, any>>({})

// 获取插件图标
const getPluginIcon = (pluginName: string) => {
  const iconMap: Record<string, any> = {
    'email': MailOutlined,
    'dingtalk': MessageOutlined,
    'wechat': WechatOutlined,
    'slack': SlackOutlined,
    'webhook': ApiOutlined
  }
  return iconMap[pluginName] || BellOutlined
}

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

// 获取插件状态颜色
const getPluginStatusColor = (status: string) => {
  const colorMap: Record<string, string> = {
    'active': 'green',
    'inactive': 'orange',
    'error': 'red'
  }
  return colorMap[status] || 'default'
}

// 获取插件状态文本
const getPluginStatusText = (status: string) => {
  const textMap: Record<string, string> = {
    'active': '正常',
    'inactive': '未激活',
    'error': '错误'
  }
  return textMap[status] || status
}

// 检查插件是否已配置
const isPluginConfigured = (pluginName: string) => {
  return !!pluginConfigs.value[pluginName]
}

// 检查插件是否已启用
const isPluginEnabled = (pluginName: string) => {
  const config = pluginConfigs.value[pluginName]
  return config?.enabled || false
}

// 加载可用插件
const loadAvailablePlugins = async () => {
  try {
    loading.value = true
    const plugins = await getAvailablePlugins()
    availablePlugins.value = plugins
    
    // 加载每个插件的配置
    await loadPluginConfigs()
  } catch (error) {
    console.error('加载插件失败:', error)
    message.error('加载插件失败')
  } finally {
    loading.value = false
  }
}

// 加载插件配置
const loadPluginConfigs = async () => {
  const configs: Record<string, PluginConfig> = {}
  
  for (const plugin of availablePlugins.value) {
    try {
      const config = await getPluginConfig(plugin.name)
      configs[plugin.name] = config
    } catch (error) {
      // 插件未配置时会返回404，这是正常的
      console.debug(`插件 ${plugin.name} 未配置`)
    }
  }
  
  pluginConfigs.value = configs
}

// 处理插件开关切换
const handlePluginToggle = async (pluginName: string, enabled: boolean) => {
  try {
    switchLoading[pluginName] = true
    
    if (enabled) {
      await enablePlugin(pluginName)
      message.success(`${getPluginDisplayName(pluginName)} 已启用`)
    } else {
      await disablePlugin(pluginName)
      message.success(`${getPluginDisplayName(pluginName)} 已禁用`)
    }
    
    // 重新加载配置
    await loadPluginConfigs()
  } catch (error) {
    console.error('切换插件状态失败:', error)
    message.error('操作失败')
  } finally {
    switchLoading[pluginName] = false
  }
}

// 打开配置模态框
const openConfigModal = async (plugin: PluginInfo) => {
  currentPlugin.value = plugin
  
  // 加载现有配置
  try {
    const config = await getPluginConfig(plugin.name)
    currentPluginConfig.value = config
  } catch (error) {
    // 插件未配置时创建默认配置
    currentPluginConfig.value = {
      name: plugin.name,
      enabled: false,
      config: {},
      priority: 0
    }
  }
  
  configModalVisible.value = true
}

// 关闭配置模态框
const closeConfigModal = () => {
  configModalVisible.value = false
  currentPlugin.value = null
  currentPluginConfig.value = null
}

// 处理配置提交
const handleConfigSubmit = async (config: PluginConfig) => {
  try {
    await setPluginConfig(config.name, config)
    message.success('配置保存成功')
    
    // 重新加载配置
    await loadPluginConfigs()
    closeConfigModal()
  } catch (error) {
    console.error('保存配置失败:', error)
    message.error('保存配置失败')
  }
}

// 打开测试模态框
const openTestModal = (pluginName: string) => {
  const config = pluginConfigs.value[pluginName]
  if (!config) {
    message.warning('请先配置插件')
    return
  }
  
  currentTestPlugin.value = pluginName
  currentTestConfig.value = config.config
  testModalVisible.value = true
}

// 处理测试完成
const handleTestCompleted = (result: any) => {
  if (result.success) {
    message.success(`${getPluginDisplayName(currentTestPlugin.value)} 测试成功`)
  } else {
    message.error(`测试失败: ${result.error}`)
  }
}

// 查看插件统计
const viewPluginStats = async (pluginName: string) => {
  try {
    const stats = await getPluginStats(pluginName)
    currentStatsPlugin.value = pluginName
    currentStats.value = stats
    statsModalVisible.value = true
  } catch (error) {
    console.error('获取统计信息失败:', error)
    message.error('获取统计信息失败')
  }
}

// 组件挂载时加载数据
onMounted(() => {
  loadAvailablePlugins()
})
</script>

<style scoped>
.notification-plugin-config {
  padding: 24px;
}

.config-header {
  margin-bottom: 24px;
}

.config-header h2 {
  margin: 0 0 8px 0;
  font-size: 24px;
  font-weight: 600;
  color: #262626;
}

.config-header p {
  margin: 0;
  color: #8c8c8c;
  font-size: 14px;
}

.plugin-card {
  height: 100%;
  transition: all 0.3s ease;
}

.plugin-card:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
}

.plugin-configured {
  border-color: #52c41a;
}

.plugin-title {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 16px;
  font-weight: 600;
}

.plugin-icon {
  font-size: 18px;
  color: #1890ff;
}

.plugin-status {
  margin-left: auto;
}

.plugin-content {
  min-height: 120px;
}

.plugin-description {
  margin: 0 0 16px 0;
  color: #595959;
  font-size: 14px;
  line-height: 1.5;
}

.plugin-meta {
  display: flex;
  flex-direction: column;
  gap: 4px;
  margin-bottom: 16px;
}

.plugin-version,
.plugin-load-time {
  font-size: 12px;
  color: #8c8c8c;
}

.plugin-error {
  margin-top: 12px;
}

.plugin-error :deep(.ant-alert) {
  font-size: 12px;
}

/* 响应式设计 */
@media (max-width: 768px) {
  .notification-plugin-config {
    padding: 16px;
  }
  
  .plugin-title {
    font-size: 14px;
  }
  
  .plugin-icon {
    font-size: 16px;
  }
  
  .plugin-content {
    min-height: auto;
  }
}
</style>