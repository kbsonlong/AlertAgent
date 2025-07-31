<template>
  <div class="version-management">
    <!-- 当前版本信息 -->
    <div class="current-version">
      <a-card title="当前版本" size="small">
        <a-descriptions :column="2" size="small">
          <a-descriptions-item label="版本号">
            <a-tag color="blue">{{ currentVersion?.version || '-' }}</a-tag>
          </a-descriptions-item>
          <a-descriptions-item label="发布时间">
            {{ currentVersion?.release_time ? formatDateTime(currentVersion.release_time) : '-' }}
          </a-descriptions-item>
          <a-descriptions-item label="配置哈希">
            <code>{{ currentVersion?.config_hash || '-' }}</code>
          </a-descriptions-item>
          <a-descriptions-item label="状态">
            <a-tag :color="getVersionStatusColor(currentVersion?.status)">
              {{ getVersionStatusText(currentVersion?.status) }}
            </a-tag>
          </a-descriptions-item>
        </a-descriptions>
      </a-card>
    </div>

    <!-- 版本列表 -->
    <div class="version-list">
      <a-card title="版本历史" size="small" style="margin-top: 16px;">
        <template #extra>
          <a-space>
            <a-button size="small" @click="loadVersions">
              <template #icon><ReloadOutlined /></template>
              刷新
            </a-button>
            <a-button size="small" type="primary" @click="handleCreateVersion">
              <template #icon><PlusOutlined /></template>
              创建版本
            </a-button>
          </a-space>
        </template>
        
        <a-table
          :columns="columns"
          :data-source="versionList"
          :loading="loading"
          :pagination="{
            current: pagination.current,
            pageSize: pagination.pageSize,
            total: pagination.total,
            showSizeChanger: true,
            showQuickJumper: true,
            showTotal: (total, range) => `第 ${range[0]}-${range[1]} 条，共 ${total} 条`
          }"
          @change="handleTableChange"
          size="small"
        >
          <template #bodyCell="{ column, record }">
            <template v-if="column.key === 'version'">
              <a-tag 
                :color="record.is_current ? 'green' : 'default'"
                style="cursor: pointer;"
                @click="handleViewVersion(record)"
              >
                {{ record.version }}
                <CheckCircleOutlined v-if="record.is_current" style="margin-left: 4px;" />
              </a-tag>
            </template>
            <template v-else-if="column.key === 'status'">
              <a-tag :color="getVersionStatusColor(record.status)">
                {{ getVersionStatusText(record.status) }}
              </a-tag>
            </template>
            <template v-else-if="column.key === 'release_time'">
              {{ formatDateTime(record.release_time) }}
            </template>
            <template v-else-if="column.key === 'config_hash'">
              <code style="font-size: 12px;">{{ record.config_hash?.substring(0, 8) }}...</code>
            </template>
            <template v-else-if="column.key === 'action'">
              <a-space>
                <a-button
                  type="link"
                  size="small"
                  @click="handleViewVersion(record)"
                >
                  查看
                </a-button>
                <a-button
                  v-if="!record.is_current"
                  type="link"
                  size="small"
                  @click="handleRollback(record)"
                >
                  回滚
                </a-button>
                <a-button
                  type="link"
                  size="small"
                  @click="handleCompare(record)"
                >
                  对比
                </a-button>
                <a-button
                  v-if="!record.is_current && record.status !== 'active'"
                  type="link"
                  size="small"
                  danger
                  @click="handleDeleteVersion(record)"
                >
                  删除
                </a-button>
              </a-space>
            </template>
          </template>
        </a-table>
      </a-card>
    </div>

    <!-- 版本详情抽屉 -->
    <a-drawer
      v-model:open="detailVisible"
      :title="`版本详情 - ${selectedVersion?.version}`"
      width="800"
      @close="handleDetailClose"
    >
      <div v-if="selectedVersion" class="version-detail">
        <a-descriptions title="基本信息" :column="1" bordered size="small">
          <a-descriptions-item label="版本号">
            {{ selectedVersion.version }}
          </a-descriptions-item>
          <a-descriptions-item label="状态">
            <a-tag :color="getVersionStatusColor(selectedVersion.status)">
              {{ getVersionStatusText(selectedVersion.status) }}
            </a-tag>
          </a-descriptions-item>
          <a-descriptions-item label="发布时间">
            {{ formatDateTime(selectedVersion.release_time) }}
          </a-descriptions-item>
          <a-descriptions-item label="配置哈希">
            <code>{{ selectedVersion.config_hash }}</code>
          </a-descriptions-item>
          <a-descriptions-item label="变更说明">
            {{ selectedVersion.changelog || '-' }}
          </a-descriptions-item>
          <a-descriptions-item label="创建者">
            {{ selectedVersion.created_by || '-' }}
          </a-descriptions-item>
        </a-descriptions>
        
        <a-divider>配置内容</a-divider>
        <div class="config-content">
          <pre>{{ formatConfig(selectedVersion.config_content) }}</pre>
        </div>
      </div>
    </a-drawer>

    <!-- 版本对比抽屉 -->
    <a-drawer
      v-model:open="compareVisible"
      title="版本对比"
      width="1200"
      @close="handleCompareClose"
    >
      <div class="version-compare">
        <a-row :gutter="16">
          <a-col :span="12">
            <a-card title="选择基准版本" size="small">
              <a-select
                v-model:value="compareForm.baseVersion"
                style="width: 100%;"
                placeholder="选择基准版本"
                @change="handleCompareVersionChange"
              >
                <a-select-option
                  v-for="version in versionList"
                  :key="version.id"
                  :value="version.id"
                >
                  {{ version.version }}
                </a-select-option>
              </a-select>
            </a-card>
          </a-col>
          <a-col :span="12">
            <a-card title="对比版本" size="small">
              <a-select
                v-model:value="compareForm.targetVersion"
                style="width: 100%;"
                placeholder="选择对比版本"
                @change="handleCompareVersionChange"
              >
                <a-select-option
                  v-for="version in versionList"
                  :key="version.id"
                  :value="version.id"
                >
                  {{ version.version }}
                </a-select-option>
              </a-select>
            </a-card>
          </a-col>
        </a-row>
        
        <div v-if="compareResult" class="compare-result" style="margin-top: 16px;">
          <a-card title="差异对比" size="small">
            <div class="diff-content">
              <div v-for="(diff, index) in compareResult" :key="index" class="diff-item">
                <div class="diff-path">{{ diff.path }}</div>
                <div class="diff-changes">
                  <div class="diff-old">
                    <span class="diff-label">基准版本:</span>
                    <code>{{ diff.oldValue }}</code>
                  </div>
                  <div class="diff-new">
                    <span class="diff-label">对比版本:</span>
                    <code>{{ diff.newValue }}</code>
                  </div>
                </div>
              </div>
            </div>
          </a-card>
        </div>
      </div>
    </a-drawer>

    <!-- 创建版本模态框 -->
    <a-modal
      v-model:open="createVisible"
      title="创建新版本"
      @ok="handleCreateSubmit"
      @cancel="handleCreateCancel"
      :confirm-loading="createLoading"
    >
      <a-form :model="createForm" layout="vertical">
        <a-form-item label="版本号" required>
          <a-input
            v-model:value="createForm.version"
            placeholder="请输入版本号，如 v1.0.1"
          />
        </a-form-item>
        <a-form-item label="变更说明">
          <a-textarea
            v-model:value="createForm.changelog"
            placeholder="请输入版本变更说明"
            :rows="4"
          />
        </a-form-item>
        <a-form-item label="立即激活">
          <a-switch v-model:checked="createForm.activate" />
          <div style="color: #8c8c8c; font-size: 12px; margin-top: 4px;">
            开启后将立即将此版本设为当前活跃版本
          </div>
        </a-form-item>
      </a-form>
    </a-modal>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { message, Modal } from 'ant-design-vue'
import {
  ReloadOutlined,
  PlusOutlined,
  CheckCircleOutlined
} from '@ant-design/icons-vue'
import { formatDateTime } from '@/utils/datetime'

// 组件属性
interface Props {
  clusterId: string
  configType: string
}

const props = defineProps<Props>()

// 响应式数据
const loading = ref(false)
const createLoading = ref(false)
const versionList = ref<any[]>([])
const currentVersion = ref<any>(null)
const selectedVersion = ref<any>(null)
const detailVisible = ref(false)
const compareVisible = ref(false)
const createVisible = ref(false)
const compareResult = ref<any[]>([])

// 分页配置
const pagination = reactive({
  current: 1,
  pageSize: 10,
  total: 0
})

// 对比表单
const compareForm = reactive({
  baseVersion: undefined,
  targetVersion: undefined
})

// 创建表单
const createForm = reactive({
  version: '',
  changelog: '',
  activate: false
})

// 表格列定义
const columns = [
  {
    title: '版本号',
    dataIndex: 'version',
    key: 'version',
    width: 120
  },
  {
    title: '状态',
    dataIndex: 'status',
    key: 'status',
    width: 100
  },
  {
    title: '发布时间',
    dataIndex: 'release_time',
    key: 'release_time',
    width: 160
  },
  {
    title: '配置哈希',
    dataIndex: 'config_hash',
    key: 'config_hash',
    width: 120
  },
  {
    title: '变更说明',
    dataIndex: 'changelog',
    key: 'changelog',
    ellipsis: true
  },
  {
    title: '创建者',
    dataIndex: 'created_by',
    key: 'created_by',
    width: 100
  },
  {
    title: '操作',
    key: 'action',
    width: 200,
    fixed: 'right'
  }
]

// 获取版本状态颜色
const getVersionStatusColor = (status: string) => {
  const colorMap: Record<string, string> = {
    active: 'green',
    inactive: 'default',
    deprecated: 'orange',
    archived: 'red'
  }
  return colorMap[status] || 'default'
}

// 获取版本状态文本
const getVersionStatusText = (status: string) => {
  const textMap: Record<string, string> = {
    active: '活跃',
    inactive: '非活跃',
    deprecated: '已弃用',
    archived: '已归档'
  }
  return textMap[status] || status
}

// 格式化配置
const formatConfig = (config: string) => {
  try {
    if (typeof config === 'string') {
      const parsed = JSON.parse(config)
      return JSON.stringify(parsed, null, 2)
    }
    return JSON.stringify(config, null, 2)
  } catch {
    return config
  }
}

// 加载版本列表
const loadVersions = async () => {
  loading.value = true
  try {
    // 模拟数据
    const mockVersions = [
      {
        id: 1,
        version: 'v1.2.0',
        status: 'active',
        is_current: true,
        release_time: '2024-01-01 10:00:00',
        config_hash: 'a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6',
        changelog: '添加新的告警规则，优化性能',
        created_by: 'admin',
        config_content: JSON.stringify({
          global: {
            scrape_interval: '15s',
            evaluation_interval: '15s'
          },
          rule_files: [
            '/etc/prometheus/rules/*.yml'
          ]
        }, null, 2)
      },
      {
        id: 2,
        version: 'v1.1.0',
        status: 'inactive',
        is_current: false,
        release_time: '2023-12-15 14:30:00',
        config_hash: 'b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6q7',
        changelog: '修复配置同步问题',
        created_by: 'operator',
        config_content: JSON.stringify({
          global: {
            scrape_interval: '30s',
            evaluation_interval: '30s'
          },
          rule_files: [
            '/etc/prometheus/rules/basic.yml'
          ]
        }, null, 2)
      },
      {
        id: 3,
        version: 'v1.0.0',
        status: 'archived',
        is_current: false,
        release_time: '2023-12-01 09:00:00',
        config_hash: 'c3d4e5f6g7h8i9j0k1l2m3n4o5p6q7r8',
        changelog: '初始版本',
        created_by: 'admin',
        config_content: JSON.stringify({
          global: {
            scrape_interval: '60s'
          }
        }, null, 2)
      }
    ]
    
    versionList.value = mockVersions
    currentVersion.value = mockVersions.find(v => v.is_current)
    pagination.total = mockVersions.length
  } catch (error) {
    message.error('加载版本列表失败')
  } finally {
    loading.value = false
  }
}

// 表格变化处理
const handleTableChange = (pag: any) => {
  pagination.current = pag.current
  pagination.pageSize = pag.pageSize
  loadVersions()
}

// 查看版本详情
const handleViewVersion = (record: any) => {
  selectedVersion.value = record
  detailVisible.value = true
}

// 版本回滚
const handleRollback = (record: any) => {
  Modal.confirm({
    title: '确认回滚',
    content: `确定要回滚到版本 ${record.version} 吗？此操作将替换当前活跃版本。`,
    okText: '确定',
    cancelText: '取消',
    onOk: async () => {
      try {
        // 这里应该调用API
        message.success('版本回滚成功')
        loadVersions()
      } catch (error) {
        message.error('版本回滚失败')
      }
    }
  })
}

// 版本对比
const handleCompare = (record: any) => {
  compareForm.targetVersion = record.id
  if (currentVersion.value) {
    compareForm.baseVersion = currentVersion.value.id
  }
  compareVisible.value = true
  handleCompareVersionChange()
}

// 删除版本
const handleDeleteVersion = (record: any) => {
  Modal.confirm({
    title: '确认删除',
    content: `确定要删除版本 ${record.version} 吗？此操作不可恢复。`,
    okText: '确定',
    cancelText: '取消',
    okType: 'danger',
    onOk: async () => {
      try {
        // 这里应该调用API
        message.success('版本删除成功')
        loadVersions()
      } catch (error) {
        message.error('版本删除失败')
      }
    }
  })
}

// 创建版本
const handleCreateVersion = () => {
  createForm.version = ''
  createForm.changelog = ''
  createForm.activate = false
  createVisible.value = true
}

// 详情抽屉关闭
const handleDetailClose = () => {
  detailVisible.value = false
  selectedVersion.value = null
}

// 对比抽屉关闭
const handleCompareClose = () => {
  compareVisible.value = false
  compareForm.baseVersion = undefined
  compareForm.targetVersion = undefined
  compareResult.value = []
}

// 对比版本变化
const handleCompareVersionChange = () => {
  if (compareForm.baseVersion && compareForm.targetVersion) {
    // 模拟对比结果
    compareResult.value = [
      {
        path: 'global.scrape_interval',
        oldValue: '30s',
        newValue: '15s'
      },
      {
        path: 'rule_files[0]',
        oldValue: '/etc/prometheus/rules/basic.yml',
        newValue: '/etc/prometheus/rules/*.yml'
      }
    ]
  }
}

// 创建提交
const handleCreateSubmit = async () => {
  if (!createForm.version.trim()) {
    message.error('请输入版本号')
    return
  }
  
  createLoading.value = true
  try {
    // 这里应该调用API
    await new Promise(resolve => setTimeout(resolve, 1000))
    message.success('版本创建成功')
    createVisible.value = false
    loadVersions()
  } catch (error) {
    message.error('版本创建失败')
  } finally {
    createLoading.value = false
  }
}

// 创建取消
const handleCreateCancel = () => {
  createVisible.value = false
}

// 组件挂载
onMounted(() => {
  loadVersions()
})
</script>

<style scoped>
.version-management {
  padding: 16px;
}

.version-detail .config-content {
  max-height: 400px;
  overflow-y: auto;
  background: #f5f5f5;
  padding: 12px;
  border-radius: 4px;
}

.version-detail .config-content pre {
  margin: 0;
  font-size: 12px;
  line-height: 1.4;
}

.version-compare .compare-result {
  max-height: 500px;
  overflow-y: auto;
}

.diff-content {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.diff-item {
  padding: 12px;
  border: 1px solid #d9d9d9;
  border-radius: 4px;
  background: #fafafa;
}

.diff-path {
  font-weight: 600;
  color: #1890ff;
  margin-bottom: 8px;
}

.diff-changes {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.diff-old,
.diff-new {
  display: flex;
  align-items: center;
  gap: 8px;
}

.diff-label {
  font-weight: 500;
  min-width: 80px;
}

.diff-old .diff-label {
  color: #ff4d4f;
}

.diff-new .diff-label {
  color: #52c41a;
}

.diff-old code {
  background: #fff2f0;
  color: #ff4d4f;
  padding: 2px 6px;
  border-radius: 3px;
}

.diff-new code {
  background: #f6ffed;
  color: #52c41a;
  padding: 2px 6px;
  border-radius: 3px;
}

:deep(.ant-table-tbody > tr > td) {
  padding: 8px 12px;
}

:deep(.ant-descriptions-item-content) {
  word-break: break-all;
}
</style>