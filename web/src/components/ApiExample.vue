<template>
  <div class="api-example">
    <a-card title="API 集成示例" class="mb-4">
      <a-space direction="vertical" size="large" style="width: 100%">
        <!-- 用户信息 -->
        <a-card size="small" title="用户信息">
          <a-button @click="getCurrentUser" :loading="userLoading" type="primary">
            获取当前用户信息
          </a-button>
          <div v-if="currentUser" class="mt-2">
            <a-descriptions :column="2" size="small">
              <a-descriptions-item label="用户名">{{ currentUser.username }}</a-descriptions-item>
              <a-descriptions-item label="邮箱">{{ currentUser.email }}</a-descriptions-item>
              <a-descriptions-item label="全名">{{ currentUser.full_name }}</a-descriptions-item>
              <a-descriptions-item label="角色">{{ currentUser.role }}</a-descriptions-item>
            </a-descriptions>
          </div>
        </a-card>

        <!-- 系统信息 -->
        <a-card size="small" title="系统信息">
          <a-button @click="getSystemInfo" :loading="systemLoading" type="primary">
            获取系统信息
          </a-button>
          <div v-if="systemInfo" class="mt-2">
            <a-descriptions :column="2" size="small">
              <a-descriptions-item label="版本">{{ systemInfo.version }}</a-descriptions-item>
              <a-descriptions-item label="构建时间">{{ systemInfo.build_time }}</a-descriptions-item>
              <a-descriptions-item label="运行时间">{{ systemInfo.uptime }}</a-descriptions-item>
              <a-descriptions-item label="数据库状态">{{ systemInfo.database.connection_status }}</a-descriptions-item>
            </a-descriptions>
          </div>
        </a-card>

        <!-- 知识库列表 -->
        <a-card size="small" title="知识库">
          <a-button @click="getKnowledgeList" :loading="knowledgeLoading" type="primary">
            获取知识库列表
          </a-button>
          <div v-if="knowledgeList.length > 0" class="mt-2">
            <a-table 
              :dataSource="knowledgeList" 
              :columns="knowledgeColumns" 
              :pagination="false"
              size="small"
            />
          </div>
        </a-card>

        <!-- 告警列表 -->
        <a-card size="small" title="告警">
          <a-button @click="getAlertList" :loading="alertLoading" type="primary">
            获取告警列表
          </a-button>
          <div v-if="alertList.length > 0" class="mt-2">
            <a-table 
              :dataSource="alertList" 
              :columns="alertColumns" 
              :pagination="false"
              size="small"
            />
          </div>
        </a-card>

        <!-- 数据源列表 -->
        <a-card size="small" title="数据源">
          <a-button @click="getProviderList" :loading="providerLoading" type="primary">
            获取数据源列表
          </a-button>
          <div v-if="providerList.length > 0" class="mt-2">
            <a-table 
              :dataSource="providerList" 
              :columns="providerColumns" 
              :pagination="false"
              size="small"
            />
          </div>
        </a-card>
      </a-space>
    </a-card>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { message } from 'ant-design-vue'
import {
  mockUserInfo,
  mockSystemInfo,
  mockKnowledgeList,
  mockAlertList,
  mockProviderList
} from '@/services/mockApi'

// 响应式数据
const userLoading = ref(false)
const systemLoading = ref(false)
const knowledgeLoading = ref(false)
const alertLoading = ref(false)
const providerLoading = ref(false)

const currentUser = ref<any>(null)
const systemInfo = ref<any>(null)
const knowledgeList = ref<any[]>([])
const alertList = ref<any[]>([])
const providerList = ref<any[]>([])

// 表格列定义
const knowledgeColumns = [
  { title: 'ID', dataIndex: 'id', key: 'id' },
  { title: '标题', dataIndex: 'title', key: 'title' },
  { title: '类型', dataIndex: 'type', key: 'type' },
  { title: '状态', dataIndex: 'status', key: 'status' },
  { title: '创建时间', dataIndex: 'created_at', key: 'created_at' }
]

const alertColumns = [
  { title: 'ID', dataIndex: 'id', key: 'id' },
  { title: '标题', dataIndex: 'title', key: 'title' },
  { title: '级别', dataIndex: 'severity', key: 'severity' },
  { title: '状态', dataIndex: 'status', key: 'status' },
  { title: '创建时间', dataIndex: 'created_at', key: 'created_at' }
]

const providerColumns = [
  { title: 'ID', dataIndex: 'id', key: 'id' },
  { title: '名称', dataIndex: 'name', key: 'name' },
  { title: '类型', dataIndex: 'type', key: 'type' },
  { title: '状态', dataIndex: 'status', key: 'status' },
  { title: '健康状态', dataIndex: 'health_status', key: 'health_status' }
]

// API 调用方法
const getCurrentUser = async () => {
  userLoading.value = true
  try {
    const response = await mockUserInfo()
    currentUser.value = response.data
    message.success('获取用户信息成功')
  } catch (error) {
    console.error('获取用户信息错误:', error)
    message.error('获取用户信息失败')
  } finally {
    userLoading.value = false
  }
}

const getSystemInfo = async () => {
  systemLoading.value = true
  try {
    const response = await mockSystemInfo()
    systemInfo.value = response.data
    message.success('获取系统信息成功')
  } catch (error) {
    console.error('获取系统信息错误:', error)
    message.error('获取系统信息失败')
  } finally {
    systemLoading.value = false
  }
}

const getKnowledgeList = async () => {
  knowledgeLoading.value = true
  try {
    const response = await mockKnowledgeList()
    knowledgeList.value = response.data.items || []
    message.success('获取知识库列表成功')
  } catch (error) {
    console.error('获取知识库列表错误:', error)
    message.error('获取知识库列表失败')
  } finally {
    knowledgeLoading.value = false
  }
}

const getAlertList = async () => {
  alertLoading.value = true
  try {
    const response = await mockAlertList()
    alertList.value = response.data.items || []
    message.success('获取告警列表成功')
  } catch (error) {
    console.error('获取告警列表错误:', error)
    message.error('获取告警列表失败')
  } finally {
    alertLoading.value = false
  }
}

const getProviderList = async () => {
  providerLoading.value = true
  try {
    const response = await mockProviderList()
    providerList.value = response.data.items || []
    message.success('获取数据源列表成功')
  } catch (error) {
    console.error('获取数据源列表错误:', error)
    message.error('获取数据源列表失败')
  } finally {
    providerLoading.value = false
  }
}

// 组件挂载时获取数据
onMounted(() => {
  getCurrentUser()
  getSystemInfo()
  getKnowledgeList()
  getAlertList()
  getProviderList()
})
</script>

<style scoped>
.api-example {
  padding: 20px;
}

.mb-4 {
  margin-bottom: 16px;
}

.mt-2 {
  margin-top: 8px;
}
</style>