<template>
  <div class="query-builder">
    <!-- 指标选择 -->
    <a-card title="指标选择" size="small" class="builder-card">
      <a-row :gutter="16">
        <a-col :span="12">
          <a-form-item label="数据源">
            <a-select
              v-model:value="queryConfig.provider"
              placeholder="请选择数据源"
              @change="handleProviderChange"
            >
              <a-select-option
                v-for="provider in providers"
                :key="provider.id"
                :value="provider.id"
              >
                {{ provider.name }}
              </a-select-option>
            </a-select>
          </a-form-item>
        </a-col>
        <a-col :span="12">
          <a-form-item label="指标名称">
            <a-select
              v-model:value="queryConfig.metric"
              placeholder="请选择指标"
              show-search
              :filter-option="filterMetric"
              @change="handleMetricChange"
            >
              <a-select-option
                v-for="metric in metrics"
                :key="metric"
                :value="metric"
              >
                {{ metric }}
              </a-select-option>
            </a-select>
          </a-form-item>
        </a-col>
      </a-row>
    </a-card>

    <!-- 标签过滤 -->
    <a-card title="标签过滤" size="small" class="builder-card">
      <div class="label-filters">
        <div
          v-for="(filter, index) in queryConfig.labelFilters"
          :key="index"
          class="filter-item"
        >
          <a-select
            v-model:value="filter.label"
            placeholder="标签名"
            style="width: 30%"
            @change="handleLabelChange(index)"
          >
            <a-select-option
              v-for="label in labels"
              :key="label"
              :value="label"
            >
              {{ label }}
            </a-select-option>
          </a-select>
          
          <a-select
            v-model:value="filter.operator"
            style="width: 20%"
          >
            <a-select-option value="=">=</a-select-option>
            <a-select-option value="!=">!=</a-select-option>
            <a-select-option value="=~">=~</a-select-option>
            <a-select-option value="!~">!~</a-select-option>
          </a-select>
          
          <a-select
            v-model:value="filter.value"
            placeholder="标签值"
            style="width: 35%"
            mode="tags"
            :options="getLabelValues(filter.label)"
          />
          
          <a-button
            type="text"
            danger
            @click="removeLabelFilter(index)"
            :disabled="queryConfig.labelFilters.length <= 1"
          >
            <template #icon><DeleteOutlined /></template>
          </a-button>
        </div>
        
        <a-button type="dashed" @click="addLabelFilter" block>
          <template #icon><PlusOutlined /></template>
          添加标签过滤
        </a-button>
      </div>
    </a-card>

    <!-- 时间范围 -->
    <a-card title="时间范围" size="small" class="builder-card">
      <a-row :gutter="16">
        <a-col :span="12">
          <a-form-item label="范围类型">
            <a-radio-group v-model:value="queryConfig.rangeType">
              <a-radio value="instant">即时查询</a-radio>
              <a-radio value="range">范围查询</a-radio>
            </a-radio-group>
          </a-form-item>
        </a-col>
        <a-col :span="12" v-if="queryConfig.rangeType === 'range'">
          <a-form-item label="时间范围">
            <a-select v-model:value="queryConfig.timeRange">
              <a-select-option value="1m">1分钟</a-select-option>
              <a-select-option value="5m">5分钟</a-select-option>
              <a-select-option value="15m">15分钟</a-select-option>
              <a-select-option value="30m">30分钟</a-select-option>
              <a-select-option value="1h">1小时</a-select-option>
              <a-select-option value="6h">6小时</a-select-option>
              <a-select-option value="12h">12小时</a-select-option>
              <a-select-option value="1d">1天</a-select-option>
            </a-select>
          </a-form-item>
        </a-col>
      </a-row>
    </a-card>

    <!-- 函数应用 -->
    <a-card title="函数应用" size="small" class="builder-card">
      <a-row :gutter="16">
        <a-col :span="12">
          <a-form-item label="聚合函数">
            <a-select
              v-model:value="queryConfig.aggregateFunction"
              placeholder="请选择聚合函数"
              allow-clear
            >
              <a-select-option value="sum">sum - 求和</a-select-option>
              <a-select-option value="avg">avg - 平均值</a-select-option>
              <a-select-option value="max">max - 最大值</a-select-option>
              <a-select-option value="min">min - 最小值</a-select-option>
              <a-select-option value="count">count - 计数</a-select-option>
              <a-select-option value="stddev">stddev - 标准差</a-select-option>
            </a-select>
          </a-form-item>
        </a-col>
        <a-col :span="12">
          <a-form-item label="时间函数">
            <a-select
              v-model:value="queryConfig.timeFunction"
              placeholder="请选择时间函数"
              allow-clear
            >
              <a-select-option value="rate">rate - 速率</a-select-option>
              <a-select-option value="irate">irate - 瞬时速率</a-select-option>
              <a-select-option value="increase">increase - 增长量</a-select-option>
              <a-select-option value="delta">delta - 变化量</a-select-option>
              <a-select-option value="deriv">deriv - 导数</a-select-option>
            </a-select>
          </a-form-item>
        </a-col>
      </a-row>
      
      <a-form-item label="分组标签" v-if="queryConfig.aggregateFunction">
        <a-select
          v-model:value="queryConfig.groupBy"
          mode="multiple"
          placeholder="请选择分组标签"
          :options="groupByOptions"
        />
      </a-form-item>
    </a-card>

    <!-- 查询预览 -->
    <a-card title="查询预览" size="small" class="builder-card">
      <div class="query-preview">
        <div class="query-text">
          <pre>{{ generatedQuery }}</pre>
        </div>
        <div class="query-actions">
          <a-space>
            <a-button type="primary" @click="applyQuery">
              <template #icon><CheckOutlined /></template>
              应用查询
            </a-button>
            <a-button @click="copyQuery">
              <template #icon><CopyOutlined /></template>
              复制
            </a-button>
            <a-button @click="resetQuery">
              <template #icon><ReloadOutlined /></template>
              重置
            </a-button>
          </a-space>
        </div>
      </div>
    </a-card>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, watch, onMounted } from 'vue'
import {
  Card,
  Row,
  Col,
  Form,
  Select,
  Radio,
  Button,
  Space,
  message
} from 'ant-design-vue'
import {
  DeleteOutlined,
  PlusOutlined,
  CheckOutlined,
  CopyOutlined,
  ReloadOutlined
} from '@ant-design/icons-vue'
import type { Provider } from '@/types'

const ACard = Card
const ARow = Row
const ACol = Col
const AFormItem = Form.Item
const ASelect = Select
const ASelectOption = Select.Option
const ARadioGroup = Radio.Group
const ARadio = Radio
const AButton = Button
const ASpace = Space

interface Props {
  providers: Provider[]
}

interface Emits {
  (e: 'query-change', query: string): void
}

const props = defineProps<Props>()
const emit = defineEmits<Emits>()

// 查询配置
const queryConfig = reactive({
  provider: '',
  metric: '',
  labelFilters: [
    { label: '', operator: '=', value: '' }
  ],
  rangeType: 'instant',
  timeRange: '5m',
  aggregateFunction: '',
  timeFunction: '',
  groupBy: []
})

// 数据选项
const metrics = ref<string[]>([])
const labels = ref<string[]>([])
const labelValues = ref<Record<string, string[]>>({})

// 分组选项
const groupByOptions = computed(() => {
  return labels.value.map(label => ({
    label,
    value: label
  }))
})

// 生成的查询
const generatedQuery = computed(() => {
  let query = ''
  
  if (!queryConfig.metric) {
    return '请选择指标'
  }
  
  // 基础指标
  query = queryConfig.metric
  
  // 添加标签过滤
  const filters = queryConfig.labelFilters
    .filter(f => f.label && f.value)
    .map(f => `${f.label}${f.operator}"${f.value}"`)
    .join(', ')
  
  if (filters) {
    query += `{${filters}}`
  }
  
  // 添加时间范围
  if (queryConfig.rangeType === 'range' && queryConfig.timeRange) {
    query += `[${queryConfig.timeRange}]`
  }
  
  // 应用时间函数
  if (queryConfig.timeFunction) {
    query = `${queryConfig.timeFunction}(${query})`
  }
  
  // 应用聚合函数
  if (queryConfig.aggregateFunction) {
    const groupBy = queryConfig.groupBy.length > 0 
      ? ` by (${queryConfig.groupBy.join(', ')})` 
      : ''
    query = `${queryConfig.aggregateFunction}${groupBy}(${query})`
  }
  
  return query
})

// 监听查询变化
watch(generatedQuery, (newQuery) => {
  emit('query-change', newQuery)
})

// 处理数据源变化
const handleProviderChange = (providerId: string) => {
  // 重置相关数据
  queryConfig.metric = ''
  queryConfig.labelFilters = [{ label: '', operator: '=', value: '' }]
  metrics.value = []
  labels.value = []
  labelValues.value = {}
  
  // 加载指标和标签
  loadMetrics(providerId)
}

// 处理指标变化
const handleMetricChange = (metric: string) => {
  // 加载该指标的标签
  loadLabels(queryConfig.provider, metric)
}

// 处理标签变化
const handleLabelChange = (index: number) => {
  const filter = queryConfig.labelFilters[index]
  if (filter.label) {
    loadLabelValues(queryConfig.provider, queryConfig.metric, filter.label)
  }
}

// 过滤指标
const filterMetric = (input: string, option: any) => {
  return option.value.toLowerCase().includes(input.toLowerCase())
}

// 获取标签值选项
const getLabelValues = (label: string) => {
  return (labelValues.value[label] || []).map(value => ({
    label: value,
    value
  }))
}

// 添加标签过滤
const addLabelFilter = () => {
  queryConfig.labelFilters.push({ label: '', operator: '=', value: '' })
}

// 删除标签过滤
const removeLabelFilter = (index: number) => {
  if (queryConfig.labelFilters.length > 1) {
    queryConfig.labelFilters.splice(index, 1)
  }
}

// 应用查询
const applyQuery = () => {
  if (!generatedQuery.value || generatedQuery.value === '请选择指标') {
    message.warning('请先构建有效的查询')
    return
  }
  
  emit('query-change', generatedQuery.value)
  message.success('查询已应用')
}

// 复制查询
const copyQuery = async () => {
  if (!generatedQuery.value || generatedQuery.value === '请选择指标') {
    message.warning('没有可复制的查询')
    return
  }
  
  try {
    await navigator.clipboard.writeText(generatedQuery.value)
    message.success('查询已复制到剪贴板')
  } catch (error) {
    // 降级方案
    const textArea = document.createElement('textarea')
    textArea.value = generatedQuery.value
    document.body.appendChild(textArea)
    textArea.select()
    document.execCommand('copy')
    document.body.removeChild(textArea)
    message.success('查询已复制到剪贴板')
  }
}

// 重置查询
const resetQuery = () => {
  Object.assign(queryConfig, {
    provider: '',
    metric: '',
    labelFilters: [{ label: '', operator: '=', value: '' }],
    rangeType: 'instant',
    timeRange: '5m',
    aggregateFunction: '',
    timeFunction: '',
    groupBy: []
  })
  
  metrics.value = []
  labels.value = []
  labelValues.value = {}
}

// 加载指标
const loadMetrics = async (providerId: string) => {
  try {
    // 这里应该调用API获取指标列表
    // const response = await getProviderMetrics(providerId)
    // metrics.value = response.data
    
    // 模拟数据
    metrics.value = [
      'up',
      'http_requests_total',
      'http_request_duration_seconds',
      'node_cpu_seconds_total',
      'node_memory_MemTotal_bytes',
      'node_memory_MemAvailable_bytes',
      'node_disk_io_time_seconds_total',
      'node_network_receive_bytes_total',
      'node_network_transmit_bytes_total'
    ]
  } catch (error) {
    console.error('加载指标失败:', error)
    message.error('加载指标失败')
  }
}

// 加载标签
const loadLabels = async (providerId: string, metric: string) => {
  try {
    // 这里应该调用API获取标签列表
    // const response = await getProviderLabels(providerId, metric)
    // labels.value = response.data
    
    // 模拟数据
    labels.value = [
      'instance',
      'job',
      'method',
      'status',
      'handler',
      'device',
      'mode',
      'cpu'
    ]
  } catch (error) {
    console.error('加载标签失败:', error)
    message.error('加载标签失败')
  }
}

// 加载标签值
const loadLabelValues = async (providerId: string, metric: string, label: string) => {
  try {
    // 这里应该调用API获取标签值列表
    // const response = await getProviderLabelValues(providerId, metric, label)
    // labelValues.value[label] = response.data
    
    // 模拟数据
    const mockValues: Record<string, string[]> = {
      instance: ['localhost:9090', 'localhost:9100', 'localhost:8080'],
      job: ['prometheus', 'node-exporter', 'api-server'],
      method: ['GET', 'POST', 'PUT', 'DELETE'],
      status: ['200', '404', '500'],
      handler: ['/api/v1/query', '/api/v1/query_range', '/metrics'],
      device: ['sda', 'sdb', 'eth0', 'lo'],
      mode: ['user', 'system', 'idle', 'iowait'],
      cpu: ['0', '1', '2', '3']
    }
    
    labelValues.value[label] = mockValues[label] || []
  } catch (error) {
    console.error('加载标签值失败:', error)
    message.error('加载标签值失败')
  }
}

// 组件挂载
onMounted(() => {
  // 初始化
})
</script>

<style scoped>
.query-builder {
  padding: 0;
}

.builder-card {
  margin-bottom: 16px;
}

.builder-card:last-child {
  margin-bottom: 0;
}

.label-filters {
  border: 1px solid #d9d9d9;
  border-radius: 6px;
  padding: 12px;
  background: #fafafa;
}

.filter-item {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 8px;
}

.filter-item:last-of-type {
  margin-bottom: 12px;
}

.query-preview {
  border: 1px solid #d9d9d9;
  border-radius: 6px;
  overflow: hidden;
}

.query-text {
  background: #f5f5f5;
  padding: 12px;
  border-bottom: 1px solid #d9d9d9;
}

.query-text pre {
  margin: 0;
  font-family: 'Courier New', monospace;
  font-size: 14px;
  line-height: 1.6;
  white-space: pre-wrap;
  word-break: break-all;
}

.query-actions {
  padding: 12px;
  background: white;
  text-align: right;
}

/* 响应式设计 */
@media (max-width: 768px) {
  .filter-item {
    flex-direction: column;
    align-items: stretch;
  }
  
  .filter-item > * {
    width: 100% !important;
    margin-bottom: 8px;
  }
  
  .filter-item > button {
    align-self: flex-end;
    width: auto !important;
  }
  
  .query-actions {
    text-align: left;
  }
  
  .query-actions :deep(.ant-space) {
    flex-wrap: wrap;
  }
}
</style>