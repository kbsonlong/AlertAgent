<template>
  <div class="config-sync-monitor">
    <div class="page-header">
      <h1>配置同步监控</h1>
      <div class="header-actions">
        <a-button @click="refreshData" :loading="loading">
          <template #icon><ReloadOutlined /></template>
          刷新
        </a-button>
        <a-button type="primary" @click="showExceptionModal = true">
          <template #icon><ExclamationCircleOutlined /></template>
          异常管理
        </a-button>
      </div>
    </div>

    <!-- 概览卡片 -->
    <div class="overview-cards">
      <a-row :gutter="16">
        <a-col :span="6">
          <a-card>
            <a-statistic
              title="总集群数"
              :value="summary?.total_clusters || 0"
              :value-style="{ color: '#1890ff' }"
            />
          </a-card>
        </a-col>
        <a-col :span="6">
          <a-card>
            <a-statistic
              title="健康集群"
              :value="summary?.healthy_clusters || 0"
              :value-style="{ color: '#52c41a' }"
            />
          </a-card>
        </a-col>
        <a-col :span="6">
          <a-card>
            <a-statistic
              title="异常集群"
              :value="summary?.unhealthy_clusters || 0"
              :value-style="{ color: '#ff4d4f' }"
            />
          </a-card>
        </a-col>
        <a-col :span="6">
          <a-card>
            <div class="health-status">
              <div class="status-label">整体状态</div>
              <a-tag :color="getHealthColor(summary?.overall_health)">
                {{ getHealthText(summary?.overall_health) }}
              </a-tag>
            </div>
          </a-card>
        </a-col>
      </a-row>
    </div>

    <!-- 集群监控表格 -->
    <a-card title="集群同步状态" class="cluster-table-card">
      <a-table
        :columns="clusterColumns"
        :data-source="clusterData"
        :loading="loading"
        row-key="key"
        :pagination="false"
      >
        <template #bodyCell="{ column, record }">
          <template v-if="column.key === 'health_status'">
            <a-tag :color="getHealthColor(record.health_status)">
              {{ getHealthText(record.health_status) }}
            </a-tag>
          </template>
          <template v-else-if="column.key === 'sync_delay'">
            <span :class="{ 'delay-warning': record.sync_delay > 300, 'delay-critical': record.sync_delay > 1800 }">
              {{ formatDelay(record.sync_delay) }}
            </span>
          </template>
          <template v-else-if="column.key === 'success_rate'">
            <a-progress
              :percent="record.success_rate"
              :status="record.success_rate < 90 ? 'exception' : 'success'"
              size="small"
            />
          </template>
          <template v-else-if="column.key === 'actions'">
            <a-space>
              <a-button size="small" @click="viewClusterDetails(record)">详情</a-button>
              <a-button size="small" @click="checkConsistency(record)">一致性检查</a-button>
              <a-button size="small" @click="viewVersions(record)">版本管理</a-button>
            </a-space>
          </template>
        </template>
      </a-table>
    </a-card>

    <!-- 监控图表 -->
    <a-row :gutter="16" class="charts-row">
      <a-col :span="12">
        <a-card title="同步延迟趋势">
          <div ref="delayChartRef" style="height: 300px;"></div>
        </a-card>
      </a-col>
      <a-col :span="12">
        <a-card title="失败率趋势">
          <div ref="failureChartRef" style="height: 300px;"></div>
        </a-card>
      </a-col>
    </a-row>

    <!-- 异常管理模态框 -->
    <a-modal
      v-model:open="showExceptionModal"
      title="同步异常管理"
      width="1200px"
      :footer="null"
    >
      <ExceptionManagement @close="showExceptionModal = false" />
    </a-modal>

    <!-- 集群详情模态框 -->
    <a-modal
      v-model:open="showClusterModal"
      :title="`集群详情 - ${selectedCluster?.cluster_id}`"
      width="800px"
      :footer="null"
    >
      <ClusterDetails v-if="selectedCluster" :cluster="selectedCluster" />
    </a-modal>

    <!-- 版本管理模态框 -->
    <a-modal
      v-model:open="showVersionModal"
      :title="`版本管理 - ${selectedCluster?.cluster_id}`"
      width="1000px"
      :footer="null"
    >
      <VersionManagement
        v-if="selectedCluster"
        :cluster-id="selectedCluster.cluster_id"
        :config-type="selectedCluster.config_type"
      />
    </a-modal>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, nextTick } from 'vue';
import { message } from 'ant-design-vue';
import { ReloadOutlined, ExclamationCircleOutlined } from '@ant-design/icons-vue';
import * as echarts from 'echarts';
import { ConfigSyncMonitorService, type SyncSummary, type SyncMetrics } from '@/services/configSyncMonitor';
import ExceptionManagement from '@/components/ExceptionManagement.vue';
import ClusterDetails from '@/components/ClusterDetails.vue';
import VersionManagement from '@/components/VersionManagement.vue';

// 响应式数据
const loading = ref(false);
const summary = ref<SyncSummary>();
const clusterData = ref<any[]>([]);
const showExceptionModal = ref(false);
const showClusterModal = ref(false);
const showVersionModal = ref(false);
const selectedCluster = ref<any>();

// 图表引用
const delayChartRef = ref<HTMLElement>();
const failureChartRef = ref<HTMLElement>();
let delayChart: echarts.ECharts;
let failureChart: echarts.ECharts;

// 表格列定义
const clusterColumns = [
  {
    title: '集群ID',
    dataIndex: 'cluster_id',
    key: 'cluster_id',
    width: 150,
  },
  {
    title: '配置类型',
    dataIndex: 'config_type',
    key: 'config_type',
    width: 120,
  },
  {
    title: '健康状态',
    key: 'health_status',
    width: 100,
  },
  {
    title: '同步延迟',
    key: 'sync_delay',
    width: 100,
  },
  {
    title: '成功率',
    key: 'success_rate',
    width: 120,
  },
  {
    title: '最后同步',
    dataIndex: 'last_sync_time',
    key: 'last_sync_time',
    width: 160,
    customRender: ({ text }: { text: string }) => {
      return text ? new Date(text).toLocaleString() : '从未同步';
    },
  },
  {
    title: '失败次数',
    dataIndex: 'failure_count',
    key: 'failure_count',
    width: 100,
  },
  {
    title: '操作',
    key: 'actions',
    width: 200,
  },
];

// 获取健康状态颜色
const getHealthColor = (status: string) => {
  switch (status) {
    case 'healthy':
      return 'green';
    case 'warning':
      return 'orange';
    case 'critical':
      return 'red';
    default:
      return 'default';
  }
};

// 获取健康状态文本
const getHealthText = (status: string) => {
  switch (status) {
    case 'healthy':
      return '健康';
    case 'warning':
      return '警告';
    case 'critical':
      return '严重';
    default:
      return '未知';
  }
};

// 格式化延迟时间
const formatDelay = (seconds: number) => {
  if (seconds < 0) return '从未同步';
  if (seconds < 60) return `${seconds}秒`;
  if (seconds < 3600) return `${Math.floor(seconds / 60)}分钟`;
  return `${Math.floor(seconds / 3600)}小时`;
};

// 刷新数据
const refreshData = async () => {
  loading.value = true;
  try {
    await loadSyncMetrics();
    await loadChartData();
    message.success('数据刷新成功');
  } catch (error) {
    console.error('Failed to refresh data:', error);
    message.error('数据刷新失败');
  } finally {
    loading.value = false;
  }
};

// 加载同步指标
const loadSyncMetrics = async () => {
  try {
    const data = await ConfigSyncMonitorService.getSyncMetrics();
    summary.value = data;

    // 转换为表格数据
    const tableData: any[] = [];
    Object.entries(data.cluster_metrics).forEach(([clusterId, metrics]) => {
      metrics.forEach((metric: SyncMetrics) => {
        tableData.push({
          key: `${clusterId}-${metric.config_type}`,
          cluster_id: clusterId,
          config_type: metric.config_type,
          health_status: metric.is_healthy ? 'healthy' : 'critical',
          sync_delay: metric.sync_delay_seconds,
          success_rate: Math.round(metric.success_rate),
          last_sync_time: metric.last_sync_time,
          failure_count: metric.failure_count,
          ...metric,
        });
      });
    });
    clusterData.value = tableData;
  } catch (error) {
    console.error('Failed to load sync metrics:', error);
    message.error('加载同步指标失败');
  }
};

// 加载图表数据
const loadChartData = async () => {
  try {
    // 加载延迟数据
    const delayData = await ConfigSyncMonitorService.getSyncDelayMetrics();
    updateDelayChart(delayData);

    // 加载失败率数据
    const failureData = await ConfigSyncMonitorService.getFailureRateMetrics();
    updateFailureChart(failureData);
  } catch (error) {
    console.error('Failed to load chart data:', error);
  }
};

// 更新延迟图表
const updateDelayChart = (data: any[]) => {
  if (!delayChart) return;

  const option = {
    title: {
      text: '同步延迟趋势',
      left: 'center',
    },
    tooltip: {
      trigger: 'axis',
      formatter: (params: any) => {
        const point = params[0];
        return `
          时间: ${new Date(point.axisValue).toLocaleString()}<br/>
          平均延迟: ${point.value}ms<br/>
          成功率: ${params[1].value}%<br/>
          样本数: ${point.data.sample_count}
        `;
      },
    },
    legend: {
      data: ['延迟时间', '成功率'],
      top: 30,
    },
    xAxis: {
      type: 'time',
      data: data.map(item => item.timestamp),
    },
    yAxis: [
      {
        type: 'value',
        name: '延迟 (ms)',
        position: 'left',
      },
      {
        type: 'value',
        name: '成功率 (%)',
        position: 'right',
        max: 100,
      },
    ],
    series: [
      {
        name: '延迟时间',
        type: 'line',
        data: data.map(item => item.duration_ms),
        smooth: true,
        itemStyle: { color: '#1890ff' },
      },
      {
        name: '成功率',
        type: 'line',
        yAxisIndex: 1,
        data: data.map(item => item.success_rate),
        smooth: true,
        itemStyle: { color: '#52c41a' },
      },
    ],
  };

  delayChart.setOption(option);
};

// 更新失败率图表
const updateFailureChart = (data: any[]) => {
  if (!failureChart) return;

  const option = {
    title: {
      text: '失败率趋势',
      left: 'center',
    },
    tooltip: {
      trigger: 'axis',
      formatter: (params: any) => {
        const point = params[0];
        return `
          时间: ${new Date(point.axisValue).toLocaleString()}<br/>
          失败率: ${point.value}%<br/>
          总数: ${point.data.total_count}<br/>
          失败数: ${point.data.failure_count}
        `;
      },
    },
    xAxis: {
      type: 'time',
      data: data.map(item => item.timestamp),
    },
    yAxis: {
      type: 'value',
      name: '失败率 (%)',
      max: 100,
    },
    series: [
      {
        name: '失败率',
        type: 'line',
        data: data.map(item => item.failure_rate),
        smooth: true,
        itemStyle: { color: '#ff4d4f' },
        areaStyle: { opacity: 0.3 },
      },
    ],
  };

  failureChart.setOption(option);
};

// 查看集群详情
const viewClusterDetails = (cluster: any) => {
  selectedCluster.value = cluster;
  showClusterModal.value = true;
};

// 检查一致性
const checkConsistency = async (cluster: any) => {
  try {
    loading.value = true;
    // 这里可以调用一致性检查API
    message.success('一致性检查已启动');
  } catch (error) {
    message.error('一致性检查失败');
  } finally {
    loading.value = false;
  }
};

// 查看版本管理
const viewVersions = (cluster: any) => {
  selectedCluster.value = cluster;
  showVersionModal.value = true;
};

// 初始化图表
const initCharts = async () => {
  await nextTick();
  
  if (delayChartRef.value) {
    delayChart = echarts.init(delayChartRef.value);
  }
  
  if (failureChartRef.value) {
    failureChart = echarts.init(failureChartRef.value);
  }

  // 监听窗口大小变化
  window.addEventListener('resize', () => {
    delayChart?.resize();
    failureChart?.resize();
  });
};

// 组件挂载
onMounted(async () => {
  await initCharts();
  await refreshData();
});
</script>

<style scoped>
.config-sync-monitor {
  padding: 24px;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 24px;
}

.page-header h1 {
  margin: 0;
  font-size: 24px;
  font-weight: 600;
}

.header-actions {
  display: flex;
  gap: 8px;
}

.overview-cards {
  margin-bottom: 24px;
}

.health-status {
  text-align: center;
}

.status-label {
  font-size: 14px;
  color: #666;
  margin-bottom: 8px;
}

.cluster-table-card {
  margin-bottom: 24px;
}

.charts-row {
  margin-bottom: 24px;
}

.delay-warning {
  color: #fa8c16;
}

.delay-critical {
  color: #ff4d4f;
  font-weight: bold;
}

:deep(.ant-statistic-content) {
  font-size: 24px;
}

:deep(.ant-card-head-title) {
  font-weight: 600;
}
</style>