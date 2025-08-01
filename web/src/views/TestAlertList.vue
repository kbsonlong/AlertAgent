<template>
  <div class="test-alert-list">
    <h2>告警列表AI分析测试页面</h2>
    <p>此页面用于测试告警列表中AI分析按钮的二次确认功能</p>
    
    <div class="test-controls">
      <a-space>
        <a-button @click="toggleAnalysisResult" type="primary">
          {{ hasAnalysisResult ? '移除分析结果' : '添加分析结果' }}
        </a-button>
        <a-button @click="resetAlerts">
          重置告警列表
        </a-button>
      </a-space>
    </div>
    
    <div class="alert-list-container">
      <!-- 告警列表 -->
      <a-card>
        <a-table
          :columns="columns"
          :data-source="testAlerts"
          :pagination="false"
          :scroll="{ x: 1200 }"
        >
          <!-- 状态列 -->
          <template #bodyCell="{ column, record }">
            <template v-if="column.key === 'status'">
              <a-tag :color="getStatusColor(record.status)" class="status-tag">
                {{ getStatusText(record.status) }}
              </a-tag>
            </template>

            <!-- 严重程度列 -->
            <template v-else-if="column.key === 'severity'">
              <a-tag :color="getSeverityColor(record.severity)">
                {{ getSeverityText(record.severity) }}
              </a-tag>
            </template>

            <!-- 时间列 -->
            <template v-else-if="column.key === 'created_at'">
              <a-tooltip :title="formatDateTime(record.created_at)">
                {{ getFriendlyTime(record.created_at) }}
              </a-tooltip>
            </template>

            <!-- 操作列 -->
            <template v-else-if="column.key === 'action'">
              <a-space>
                <a-button type="link" size="small">
                  查看
                </a-button>
                <a-button
                  v-if="record.status === 'firing'"
                  type="link"
                  size="small"
                >
                  确认
                </a-button>
                <a-button
                  v-if="record.status !== 'resolved'"
                  type="link"
                  size="small"
                >
                  解决
                </a-button>
                <a-dropdown>
                  <a-button type="link" size="small">
                    更多 <DownOutlined />
                  </a-button>
                  <template #overlay>
                    <a-menu>
                      <a-menu-item @click="analyzeAlert(record as Alert)">
                        <BulbOutlined /> AI分析
                      </a-menu-item>
                      <a-menu-item>
                        <BookOutlined /> 转为知识
                      </a-menu-item>
                      <a-menu-divider />
                      <a-menu-item danger>
                        <DeleteOutlined /> 删除
                      </a-menu-item>
                    </a-menu>
                  </template>
                </a-dropdown>
              </a-space>
            </template>
          </template>
        </a-table>
      </a-card>
    </div>

    <!-- AI分析结果模态框 -->
    <a-modal
      v-model:open="analysisModalVisible"
      title="AI分析结果"
      width="800"
      :footer="null"
    >
      <div v-if="analysisResult">
        <h4>分析结果</h4>
        <p>{{ analysisResult.analysis }}</p>
      </div>
    </a-modal>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, h } from 'vue'
import {
  Card,
  Table,
  Button,
  Space,
  Tag,
  Tooltip,
  Modal,
  Dropdown,
  Menu,
  message
} from 'ant-design-vue'
import {
  DownOutlined,
  BulbOutlined,
  BookOutlined,
  DeleteOutlined,
  ExclamationCircleOutlined
} from '@ant-design/icons-vue'
import { formatDateTime, getFriendlyTime } from '@/utils/datetime'
import type { Alert, AlertAnalysis } from '@/types'

const ACard = Card
const ATable = Table
const AButton = Button
const ASpace = Space
const ATag = Tag
const ATooltip = Tooltip
const AModal = Modal
const ADropdown = Dropdown
const AMenu = Menu
const AMenuItem = Menu.Item
const AMenuDivider = Menu.Divider

const hasAnalysisResult = ref(false)
const analysisModalVisible = ref(false)
const analysisResult = ref<AlertAnalysis | null>(null)

// 测试告警数据
const testAlerts = reactive<Alert[]>([
  {
    id: 1,
    created_at: '2025-01-01T10:00:00+08:00',
    updated_at: '2025-01-01T10:00:00+08:00',
    name: '服务器CPU使用率告警',
    title: '服务器CPU使用率过高',
    level: 'medium',
    status: 'firing',
    source: 'prometheus',
    content: '服务器web-01的CPU使用率已达到85%，超过阈值80%',
    description: '这是一个测试告警，用于验证AI分析二次确认功能',
    labels: '{"instance": "web-01", "job": "node-exporter", "team": "ops"}',
    annotations: undefined,
    metrics: {
      current: 85,
      threshold: 80,
      unit: '%',
      status: 'critical'
    },
    history: [
      {
        timestamp: '2025-01-01T10:00:00+08:00',
        status: 'firing',
        value: 85,
        note: '告警触发'
      }
    ],
    rule_id: 1,
    template_id: undefined,
    group_id: undefined,
    handler: undefined,
    handle_time: undefined,
    handle_note: undefined,
    analysis: undefined,
    notify_time: undefined,
    notify_count: 0,
    severity: 'high',
    analysis_result: undefined
  },
  {
    id: 2,
    created_at: '2025-01-01T09:30:00+08:00',
    updated_at: '2025-01-01T09:30:00+08:00',
    name: '内存使用率告警',
    title: '服务器内存使用率过高',
    level: 'high',
    status: 'acknowledged',
    source: 'prometheus',
    content: '服务器web-02的内存使用率已达到90%，超过阈值85%',
    description: '这是另一个测试告警',
    labels: '{"instance": "web-02", "job": "node-exporter", "team": "ops"}',
    annotations: undefined,
    metrics: {
      current: 90,
      threshold: 85,
      unit: '%',
      status: 'critical'
    },
    history: [
      {
        timestamp: '2025-01-01T09:30:00+08:00',
        status: 'firing',
        value: 90,
        note: '告警触发'
      },
      {
        timestamp: '2025-01-01T09:35:00+08:00',
        status: 'acknowledged',
        value: 90,
        note: '告警已确认'
      }
    ],
    rule_id: 2,
    template_id: undefined,
    group_id: undefined,
    handler: undefined,
    handle_time: undefined,
    handle_note: undefined,
    analysis: undefined,
    notify_time: undefined,
    notify_count: 1,
    severity: 'critical',
    analysis_result: undefined
  }
])

// 表格列配置
const columns = [
  {
    title: '告警名称',
    dataIndex: 'name',
    key: 'name',
    width: 200,
    ellipsis: true
  },
  {
    title: '状态',
    dataIndex: 'status',
    key: 'status',
    width: 100
  },
  {
    title: '严重程度',
    dataIndex: 'severity',
    key: 'severity',
    width: 100
  },
  {
    title: '描述',
    dataIndex: 'description',
    key: 'description',
    ellipsis: true
  },
  {
    title: '数据源',
    dataIndex: 'source',
    key: 'source',
    width: 120
  },
  {
    title: '创建时间',
    dataIndex: 'created_at',
    key: 'created_at',
    width: 150
  },
  {
    title: '操作',
    key: 'action',
    width: 200,
    fixed: 'right' as const
  }
]

// 获取状态颜色
const getStatusColor = (status: string) => {
  const colorMap: Record<string, string> = {
    firing: 'red',
    resolved: 'green',
    acknowledged: 'blue'
  }
  return colorMap[status] || 'default'
}

// 获取状态文本
const getStatusText = (status: string) => {
  const textMap: Record<string, string> = {
    firing: '触发中',
    resolved: '已解决',
    acknowledged: '已确认'
  }
  return textMap[status] || status
}

// 获取严重程度颜色
const getSeverityColor = (severity: string) => {
  const colorMap: Record<string, string> = {
    low: 'green',
    medium: 'orange',
    high: 'red',
    critical: 'purple'
  }
  return colorMap[severity] || 'default'
}

// 获取严重程度文本
const getSeverityText = (severity: string) => {
  const textMap: Record<string, string> = {
    low: '低',
    medium: '中',
    high: '高',
    critical: '严重'
  }
  return textMap[severity] || severity
}

// 切换分析结果
const toggleAnalysisResult = () => {
  hasAnalysisResult.value = !hasAnalysisResult.value
  
  if (hasAnalysisResult.value) {
    // 为第一个告警添加分析结果
    testAlerts[0].analysis_result = {
      id: 1,
      alert_id: 1,
      analysis: '根据监控数据分析，服务器CPU使用率过高可能是由于以下原因：\n\n1. **应用程序负载增加**：可能有大量用户请求或批处理任务正在运行\n2. **资源配置不足**：服务器配置可能无法满足当前业务需求\n3. **代码性能问题**：应用程序可能存在性能瓶颈\n\n**建议处理措施：**\n- 立即检查当前运行的进程和服务\n- 考虑扩容或优化应用程序\n- 监控后续趋势，必要时进行资源调整',
      analyzed_at: '2025-01-01T10:05:00+08:00',
      model: 'gpt-4',
      confidence: 0.85,
      severity_assessment: 'high',
      root_cause: '应用程序负载增加导致CPU使用率过高',
      contributing_factors: ['用户请求量增加', '批处理任务运行', '资源配置不足'],
      business_impact: '可能影响用户体验和服务响应时间',
      user_impact: '用户可能遇到页面加载缓慢的问题',
      system_impact: '系统整体性能下降',
      impact_description: 'CPU使用率过高会导致系统响应变慢，影响用户体验',
      immediate_actions: ['检查当前进程', '监控系统负载', '准备扩容方案'],
      long_term_actions: ['优化应用程序性能', '调整资源配置', '建立监控告警'],
      prevention_measures: ['定期性能测试', '容量规划', '代码优化'],
      similar_alerts: [],
      knowledge_references: [],
      created_at: '2025-01-01T10:05:00+08:00',
      updated_at: '2025-01-01T10:05:00+08:00'
    }
    message.success('已为第一个告警添加分析结果')
  } else {
    // 移除分析结果
    testAlerts[0].analysis_result = undefined
    testAlerts[0].analysis = undefined
    message.success('已移除分析结果')
  }
}

// 重置告警列表
const resetAlerts = () => {
  hasAnalysisResult.value = false
  testAlerts.forEach(alert => {
    alert.analysis_result = undefined
    alert.analysis = undefined
  })
  message.success('告警列表已重置')
}

// AI分析告警
const analyzeAlert = async (alert: Alert) => {
  // 检查是否已有分析结果
  if (alert.analysis_result || alert.analysis) {
    // 弹出二次确认对话框
    Modal.confirm({
      title: '重新分析确认',
      content: '该告警已存在AI分析结果，是否要重新进行分析？重新分析将覆盖现有的分析结果。',
      icon: h(ExclamationCircleOutlined, { style: { color: '#faad14' } }),
      okText: '确定',
      cancelText: '取消',
      onOk: () => {
        performAnalysis(alert)
      }
    })
  } else {
    // 直接进行分析
    performAnalysis(alert)
  }
}

// 执行AI分析
const performAnalysis = async (alert: Alert) => {
  try {
    message.loading('正在进行AI分析...', 2)
    
    // 模拟API调用延迟
    await new Promise(resolve => setTimeout(resolve, 2000))
    
    // 模拟分析结果
    analysisResult.value = {
      id: Number(Date.now()),
      alert_id: alert.id,
      analysis: `针对告警"${alert.name}"的AI分析结果：\n\n这是一个${alert.severity}级别的告警，需要及时处理。建议立即检查相关系统状态并采取相应措施。`,
      analyzed_at: new Date().toISOString(),
      model: 'gpt-4',
      confidence: 0.88,
      severity_assessment: alert.severity,
      created_at: new Date().toISOString(),
      updated_at: new Date().toISOString()
    }
    
    analysisModalVisible.value = true
    message.success('AI分析完成')
  } catch (error) {
    message.error('AI分析失败')
  }
}
</script>

<style scoped>
.test-alert-list {
  padding: 24px;
}

.test-controls {
  margin-bottom: 24px;
  padding: 16px;
  background: #f5f5f5;
  border-radius: 6px;
}

.alert-list-container {
  margin-top: 16px;
}

.status-tag {
  font-weight: 500;
}

/* 响应式设计 */
@media (max-width: 768px) {
  .test-alert-list {
    padding: 16px;
  }
  
  .test-controls {
    padding: 12px;
  }
}
</style>