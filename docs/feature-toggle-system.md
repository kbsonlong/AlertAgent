# AlertAgent 功能开关系统

## 概述

AlertAgent 功能开关系统是一个支持分阶段实施和AI模型成熟度评估的智能功能管理系统。它允许系统在保证告警及时性的前提下，渐进式地启用智能功能。

## 核心特性

### 1. 分阶段实施策略

系统分为两个主要阶段：

#### 第一阶段：告警及时性优先
- **直通路由** (`direct_routing`): 告警直接通过用户定义渠道发送
- **基础收敛** (`basic_convergence`): 可选的基础告警收敛功能
- **异步分析** (`async_analysis`): 异步AI分析，不影响告警发送
- **渠道管理** (`channel_management`): 告警渠道管理系统
- **集群同步** (`cluster_sync`): Alertmanager集群配置同步

#### 第二阶段：智能功能
- **智能路由** (`smart_routing`): 基于AI的智能路由决策
- **高级收敛** (`advanced_convergence`): 基于机器学习的智能收敛
- **自动抑制** (`auto_suppression`): 自动告警抑制
- **AI决策** (`ai_decision_making`): AI驱动的自动化决策
- **自动修复** (`auto_remediation`): 自动修复功能

### 2. AI模型成熟度评估

系统会持续评估AI模型的成熟度，包括：
- **准确率** (Accuracy): 模型预测的准确程度
- **置信度** (Confidence): 模型对预测结果的信心
- **延迟** (Latency): 模型响应时间
- **成功率** (Success Rate): 操作成功的比例

### 3. 自动降级机制

当AI模型性能不达标时，系统会自动降级相关功能，确保系统稳定性。

## 配置文件

### 主配置 (config/config.yaml)

```yaml
features:
    enabled: true
    config_path: "config/features.yaml"
    monitoring_enabled: true
    ai_maturity_enabled: true
    default_phase: "phase_one"
    auto_degradation_enabled: true
    metrics_retention_hours: 168  # 7 days
    alerting:
        enabled: true
        webhook_url: ""
        slack_channel: "#alerts"
        email_recipients: []
        alert_thresholds:
            error_rate: 0.05
            latency_p99: 1000
            ai_accuracy_drop: 0.1
```

### 功能配置 (config/features.yaml)

详细的功能配置文件，包含每个功能的状态、依赖关系、AI成熟度要求等。

## API 接口

### 功能管理

- `GET /api/v1/features` - 列出所有功能
- `GET /api/v1/features/{name}` - 获取功能详情
- `PUT /api/v1/features/{name}` - 更新功能状态
- `GET /api/v1/features/{name}/check` - 检查功能是否启用

### 阶段管理

- `POST /api/v1/features/phases/{phase}/enable` - 启用阶段
- `POST /api/v1/features/phases/{phase}/disable` - 禁用阶段

### AI成熟度

- `GET /api/v1/features/{name}/ai-maturity` - 获取AI成熟度评估
- `POST /api/v1/features/{name}/ai-metrics` - 记录AI指标

### 监控和告警

- `GET /api/v1/features/{name}/monitoring` - 获取监控报告
- `GET /api/v1/features/alerts` - 获取活跃告警

### 配置管理

- `GET /api/v1/features/export` - 导出配置
- `POST /api/v1/features/import` - 导入配置

## 使用示例

### 1. 检查功能状态

```go
import "alert_agent/internal/pkg/feature"

// 检查功能是否启用
enabled := featureService.IsEnabled(ctx, feature.FeatureDirectRouting)
if enabled {
    // 执行直通路由逻辑
}
```

### 2. 带用户上下文的检查

```go
userContext := map[string]interface{}{
    "user_group": "beta_testers",
    "cluster":    "staging",
}

enabled := featureService.IsEnabled(ctx, feature.FeatureSmartRouting, userContext)
```

### 3. 记录AI指标

```go
metrics := feature.AIMetrics{
    Accuracy:    0.92,
    Confidence:  0.88,
    Latency:     250,
    SuccessRate: 0.96,
    ErrorRate:   0.04,
    SampleCount: 100,
}

featureService.RecordAIMetrics(feature.FeatureSmartRouting, metrics)
```

### 4. 更新功能状态

```go
config, _ := featureService.GetFeature(feature.FeatureBasicConvergence)
newConfig := *config
newConfig.State = feature.StateEnabled
featureService.UpdateFeature(feature.FeatureBasicConvergence, &newConfig)
```

## 监控指标

系统提供以下Prometheus指标：

- `alertagent_feature_state` - 功能状态
- `alertagent_feature_usage_total` - 功能使用次数
- `alertagent_feature_toggles_total` - 功能切换次数
- `alertagent_feature_errors_total` - 功能错误次数
- `alertagent_ai_maturity_score` - AI成熟度分数
- `alertagent_feature_latency_seconds` - 功能执行延迟
- `alertagent_feature_degradations_total` - 功能降级次数

## 告警规则

系统内置以下告警规则：

1. **功能高错误率** - 功能错误率超过阈值
2. **AI成熟度过低** - AI模型成熟度分数低于要求
3. **功能频繁降级** - 功能在短时间内多次降级

## 最佳实践

### 1. 渐进式启用

- 从第一阶段开始，确保基础功能稳定
- 逐步启用第二阶段功能，密切监控性能
- 使用金丝雀部署和A/B测试

### 2. AI模型管理

- 定期评估AI模型性能
- 设置合理的成熟度阈值
- 建立模型降级和恢复机制

### 3. 监控和告警

- 配置适当的告警阈值
- 建立功能性能基线
- 定期审查功能使用情况

### 4. 配置管理

- 使用版本控制管理配置文件
- 定期备份功能配置
- 建立配置变更审批流程

## 故障排除

### 常见问题

1. **功能无法启用**
   - 检查依赖功能是否已启用
   - 验证AI成熟度是否满足要求
   - 查看系统日志获取详细错误信息

2. **AI功能自动降级**
   - 检查AI模型性能指标
   - 验证数据质量和模型训练状态
   - 调整成熟度阈值或改进模型

3. **配置热重载失败**
   - 检查配置文件语法
   - 验证文件权限
   - 查看配置监听器日志

### 调试命令

```bash
# 查看功能状态
curl http://localhost:8080/api/v1/features

# 检查特定功能
curl http://localhost:8080/api/v1/features/smart_routing

# 获取AI成熟度评估
curl http://localhost:8080/api/v1/features/smart_routing/ai-maturity

# 查看监控报告
curl http://localhost:8080/api/v1/features/smart_routing/monitoring?hours=24
```

## 扩展开发

### 添加新功能

1. 在 `feature/toggle.go` 中定义新的功能名称
2. 在 `config/features.yaml` 中添加功能配置
3. 在业务代码中使用 `featureService.IsEnabled()` 检查功能状态
4. 添加相应的监控指标和告警规则

### 自定义AI评估器

实现 `AIMaturityEvaluator` 接口，提供自定义的成熟度评估逻辑。

### 扩展监控指标

在 `FeatureMonitor` 中添加新的Prometheus指标，用于监控特定的功能行为。