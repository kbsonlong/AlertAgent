<template>
  <div class="test-alert-detail">
    <h2>告警详情测试页面</h2>
    <p>此页面用于测试AI分析二次确认功能</p>
    
    <div class="test-controls">
      <a-space>
        <a-button @click="toggleAnalysisResult" type="primary">
          {{ hasAnalysisResult ? '移除分析结果' : '添加分析结果' }}
        </a-button>
        <a-button @click="resetAlert">
          重置告警
        </a-button>
      </a-space>
    </div>
    
    <div class="alert-detail-container">
      <AlertDetail
        :alert="testAlert"
        @update="handleAlertUpdate"
        @close="() => {}"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive } from 'vue'
import { Button, Space } from 'ant-design-vue'
import AlertDetail from '@/components/AlertDetail.vue'
import type { Alert, AlertAnalysis } from '@/types'

const AButton = Button
const ASpace = Space

const hasAnalysisResult = ref(false)

// 测试告警数据
const testAlert = reactive<Alert>({
  id: 1,
  created_at: '2025-07-31T23:41:43+08:00',
  updated_at: '2025-08-01T20:40:17+08:00',
  name: '服务器CPU使用率告警',
  title: '服务器CPU使用率过高',
  level: 'medium',
  status: 'resolved',
  source: 'prometheus',
  content: '服务器web-01的CPU使用率已达到85%，超过阈值80%',
  description: '这是一个测试告警，用于验证告警详情功能',
  labels: '{"instance": "web-01", "job": "node-exporter", "team": "ops"}',
  severity: 'medium',
  notify_count: 0,
  rule_id: 1,
   template_id: 3,
   analysis_result: undefined,
  analysis: `<think>\n嗯，我现在要分析这个关于服务器CPU使用率告警的问题。首先，看看告警标题是"服务器CPU使用率过高"，来源是从Prometheus来的，级别是medium，也就是说有点严重但不是最紧急的情况。内容显示web-01的CPU用了85%，超过了阈值80%。\n\n首先是严重程度和影响范围。因为85%已经接近饱和了，可能会影响服务器的性能，导致响应时间变长、服务稳定性问题，甚至可能导致业务中断。影响范围只是这台服务器，但如果是关键业务的话，后果就更大了。\n\n接下来分析原因，可能是运行着很多任务或进程，比如Java应用、Python框架或者后台进程太多。另外，网络延迟也可能是一个因素，如果Web-01连接有问题，响应慢会导致资源耗尽。还有可能系统本身配置高，没有进行优化，比如线程池设置过满或者是内存不足导致的。\n\n然后是处理方案，首先检查负载情况，看看有没有启动太多的进程或者任务，调整它们的数量。其次，监控网络延迟，确保网络没问题。优化系统配置，关闭不必要的服务和资源管理工具。最后，考虑升级硬件或使用分布式架构来提高效率。\n\n预防措施方面，定期维护服务器，清理旧进程；实施QoS策略，保障关键应用的网络性能；进行性能调优，比如减少多线程或者优化数据库查询；以及使用自动化监控和告警系统，及时发现潜在问题。\n\n总的来说，虽然这只是一个中等级别的告警，但如果不采取措施，可能会影响服务器性能和业务连续性。所以需要从多个方面入手，找到根本原因，并采取有效的解决方案。\n</think>\n\n### 分析与建议\n\n#### 1. 告警的严重程度和影响范围\n- **严重程度**：服务器CPU使用率达到了85%，接近阈值上限（80%）。虽然未达到100%，但已经提示可能存在性能压力，可能会影响服务器的整体性能和响应速度。\n- **影响范围**：主要涉及Web-01服务器，可能导致以下问题：\n  - 应用程序运行缓慢或卡顿。\n  - 服务响应时间增加，影响用户体验。\n  - 可能引发系统 instability，尤其是在高负载情况下。\n\n#### 2. 可能的原因分析\n根据提供的信息和常见配置，以下是可能的原因列表：\n\n- **资源密集型应用程序**：Web-01可能运行多个Java、Python或其他资源密集型任务（如JDBC连接、数据处理等）。\n  \n- **多线程或进程占用资源**：服务器上可能同时运行了多个线程或进程，导致CPU超负荷运转。\n\n- **网络延迟**：虽然当前告警没有直接提到网络问题，但网络延迟可能导致资源耗尽或响应变慢，间接影响CPU使用情况。\n\n- **系统配置不足**：\n  - 缺乏足够的内存分配给Web-01。\n  - 系统未进行优化的设置，如线程池大小过大。\n\n#### 3. 建议的处理方案\n为了缓解和解决当前的问题，并预防未来可能出现的高CPU使用情况，建议采取以下措施：\n\n- **优化应用性能**：\n  - 审查Web-01上运行的应用程序，识别并关闭不必要的资源占用（如未使用的线程、进程或数据库连接）。\n  \n  - 使用JDK Profiler等工具进行性能分析，优化应用逻辑和代码效率。\n\n- **监控与调整资源分配**：\n  - 在Linux系统中，通过调整\`nice\`值限制运行中的后台进程对CPU的占用。\n\n- **升级硬件资源**：\n  - 增加Web-01的内存容量，确保其能够满足当前负载需求。\n  \n  - 如果可能，增加Web-01的物理CPU核数以应对高负载情况。\n\n- **调整系统设置**：\n  - 使用\`htop\`或类似工具监控CPU使用情况，并在必要时调整线程池大小和进程数量。\n  \n  - 关闭不必要的后台服务和进程，确保资源被合理分配。\n\n#### 4. 预防措施建议\n为了预防未来出现的高CPU使用问题，可以采取以下预防性措施：\n\n- **定期清理旧进程**：\n  使用\`ls /var/log/foregroundLog | grep ^sh-\`等脚本清理已终止或不再需要的后台进程。\n\n- **实施QoS策略**：\n  在网络上为Web-01分配优先级较高的队列，确保关键应用的响应速度不受影响。\n\n- **定期进行性能调优**：\n  定期检查和优化Web-01服务器的应用程序，减少资源浪费，并根据负载波动调整资源使用策略。\n\n- **监控与告警系统**：\n  配置Prometheus或其他监控工具，持续跟踪Web-01的CPU使用情况、内存使用率以及其他关键指标。及时触发自动告警，提前采取应对措施。\n\n通过以上分析和建议，可以有效缓解当前的高CPU使用问题，并预防未来的潜在风险。`
})

// 模拟分析结果
const mockAnalysisResult: AlertAnalysis = {
  id: 1,
  alert_id: 1,
  analysis: '根据监控数据分析，服务器CPU使用率过高可能由以下原因导致：\n\n1. **进程异常**：某个进程占用大量CPU资源\n2. **负载突增**：业务请求量突然增加\n3. **资源不足**：服务器配置无法满足当前负载需求\n\n建议立即检查进程状态，必要时重启异常进程或扩容服务器资源。',
  analyzed_at: '2024-01-01T10:15:00Z',
  model: 'Ollama-Qwen',
  confidence: 0.85,
  severity_assessment: 'critical',
  root_cause: '进程异常导致CPU占用过高',
  contributing_factors: [
    '某个Java进程占用CPU超过80%',
    '内存不足导致频繁GC',
    '磁盘I/O瓶颈影响系统性能'
  ],
  business_impact: 'high',
  user_impact: 'medium',
  system_impact: 'high',
  impact_description: '可能导致系统响应缓慢，影响用户体验',
  immediate_actions: [
    '检查异常进程并重启',
    '监控内存使用情况',
    '检查磁盘空间'
  ],
  long_term_actions: [
    '优化应用程序性能',
    '考虑服务器扩容',
    '建立更完善的监控体系'
  ],
  prevention_measures: [
    '设置更细粒度的监控阈值',
    '定期进行性能测试',
    '建立自动扩容机制'
  ],
  created_at: '2024-01-01T10:15:00Z',
  updated_at: '2024-01-01T10:15:00Z'
}

// 切换分析结果
const toggleAnalysisResult = () => {
  if (hasAnalysisResult.value) {
    testAlert.analysis_result = undefined
    hasAnalysisResult.value = false
  } else {
    testAlert.analysis_result = mockAnalysisResult
    hasAnalysisResult.value = true
  }
}

// 重置告警
const resetAlert = () => {
  testAlert.analysis_result = undefined
  hasAnalysisResult.value = false
  testAlert.status = 'firing'
}

// 处理告警更新
const handleAlertUpdate = (updatedAlert: Alert) => {
  Object.assign(testAlert, updatedAlert)
}
</script>

<style scoped>
.test-alert-detail {
  padding: 24px;
  max-width: 1200px;
  margin: 0 auto;
}

.test-controls {
  margin: 24px 0;
  padding: 16px;
  background: #f5f5f5;
  border-radius: 8px;
}

.alert-detail-container {
  background: white;
  border-radius: 8px;
  padding: 24px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}
</style>