# AlertAgent 监控和运维指南

本文档详细介绍了AlertAgent系统的监控、日志收集和运维管理。

## 目录

- [监控架构](#监控架构)
- [Prometheus监控](#prometheus监控)
- [Grafana仪表板](#grafana仪表板)
- [日志收集](#日志收集)
- [告警配置](#告警配置)
- [性能监控](#性能监控)
- [故障排查](#故障排查)
- [运维操作](#运维操作)

## 监控架构

AlertAgent采用现代化的监控架构，包含以下组件：

```
┌─────────────────────────────────────────────────────────────┐
│                    监控架构图                                │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐         │
│  │  Grafana    │  │ Prometheus  │  │AlertManager │         │
│  │  (可视化)   │  │  (监控)     │  │  (告警)     │         │
│  └─────────────┘  └─────────────┘  └─────────────┘         │
│                                                             │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐         │
│  │   Kibana    │  │Elasticsearch│  │   Fluentd   │         │
│  │  (日志查询) │  │  (日志存储) │  │  (日志收集) │         │
│  └─────────────┘  └─────────────┘  └─────────────┘         │
│                                                             │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐         │
│  │Node Exporter│  │MySQL Export │  │Redis Export │         │
│  │ (系统监控)  │  │ (数据库监控)│  │ (缓存监控)  │         │
│  └─────────────┘  └─────────────┘  └─────────────┘         │
└─────────────────────────────────────────────────────────────┘
```

## Prometheus监控

### 1. 部署Prometheus

```bash
# 使用脚本一键部署
./scripts/monitoring-setup.sh deploy

# 手动部署
helm upgrade --install prometheus prometheus-community/kube-prometheus-stack \
    --namespace monitoring \
    --create-namespace
```

### 2. 监控指标

#### 应用指标

- **HTTP请求指标**
  - `alertagent_http_requests_total`: HTTP请求总数
  - `alertagent_http_request_duration_seconds`: HTTP请求延迟
  - `alertagent_http_requests_in_flight`: 正在处理的HTTP请求数

- **业务指标**
  - `alertagent_alerts_processed_total`: 处理的告警总数
  - `alertagent_ai_analysis_total`: AI分析任务总数
  - `alertagent_ai_analysis_duration_seconds`: AI分析耗时
  - `alertagent_notification_total`: 通知发送总数
  - `alertagent_queue_size`: 队列长度

- **系统指标**
  - `alertagent_goroutines`: Goroutine数量
  - `alertagent_memory_usage_bytes`: 内存使用量
  - `alertagent_cpu_usage_seconds`: CPU使用时间

#### 基础设施指标

- **MySQL指标**
  - `mysql_global_status_threads_connected`: 连接数
  - `mysql_global_status_slow_queries`: 慢查询数
  - `mysql_global_status_queries`: 查询总数

- **Redis指标**
  - `redis_connected_clients`: 连接客户端数
  - `redis_memory_used_bytes`: 内存使用量
  - `redis_keyspace_hits_total`: 缓存命中数

- **Kubernetes指标**
  - `kube_pod_status_ready`: Pod就绪状态
  - `kube_deployment_status_replicas`: 副本数状态
  - `container_memory_working_set_bytes`: 容器内存使用

### 3. 查询示例

```promql
# HTTP请求速率
rate(alertagent_http_requests_total[5m])

# 错误率
rate(alertagent_http_requests_total{status=~"5.."}[5m]) / rate(alertagent_http_requests_total[5m])

# 95%分位延迟
histogram_quantile(0.95, rate(alertagent_http_request_duration_seconds_bucket[5m]))

# 队列积压
alertagent_queue_size > 100

# CPU使用率
rate(container_cpu_usage_seconds_total{pod=~"alertagent-.*"}[5m])
```

## Grafana仪表板

### 1. 访问Grafana

```bash
# 端口转发
kubectl port-forward svc/prometheus-grafana 3000:80 -n monitoring

# 访问地址: http://localhost:3000
# 用户名: admin
# 密码: admin123
```

### 2. 主要仪表板

#### AlertAgent系统监控仪表板

- **系统概览**: 服务状态、请求量、错误率
- **性能指标**: 响应时间、吞吐量、资源使用
- **业务指标**: 告警处理量、AI分析成功率
- **基础设施**: 数据库、缓存、Kubernetes集群状态

#### 关键面板配置

```json
{
  "title": "HTTP请求速率",
  "type": "graph",
  "targets": [
    {
      "expr": "rate(alertagent_http_requests_total[5m])",
      "legendFormat": "{{method}} {{path}}"
    }
  ]
}
```

### 3. 告警配置

```yaml
# Grafana告警规则
- alert: HighErrorRate
  expr: rate(alertagent_http_requests_total{status=~"5.."}[5m]) > 0.1
  for: 2m
  annotations:
    summary: "HTTP错误率过高"
    description: "错误率: {{ $value }}"
```

## 日志收集

### 1. 日志架构

- **Fluentd**: 日志收集和转发
- **Elasticsearch**: 日志存储和索引
- **Kibana**: 日志查询和分析

### 2. 日志格式

#### 结构化日志

```json
{
  "timestamp": "2024-01-15T10:30:00Z",
  "level": "INFO",
  "service": "alertagent-core",
  "message": "Alert processed successfully",
  "alert_id": "alert-123",
  "duration": 1.5,
  "user_id": "user-456"
}
```

#### 日志级别

- **DEBUG**: 调试信息
- **INFO**: 一般信息
- **WARN**: 警告信息
- **ERROR**: 错误信息
- **FATAL**: 致命错误

### 3. 日志查询

#### Kibana查询示例

```
# 查询错误日志
level:ERROR AND service:alertagent-core

# 查询特定时间范围的日志
@timestamp:[2024-01-15T00:00:00 TO 2024-01-15T23:59:59] AND level:ERROR

# 查询包含特定关键词的日志
message:"database connection failed"

# 聚合查询 - 按服务统计错误数
{
  "aggs": {
    "services": {
      "terms": {
        "field": "service.keyword"
      },
      "aggs": {
        "errors": {
          "filter": {
            "term": {
              "level": "ERROR"
            }
          }
        }
      }
    }
  }
}
```

### 4. 日志保留策略

```yaml
# Elasticsearch索引生命周期管理
PUT _ilm/policy/alertagent-policy
{
  "policy": {
    "phases": {
      "hot": {
        "actions": {
          "rollover": {
            "max_size": "10GB",
            "max_age": "7d"
          }
        }
      },
      "warm": {
        "min_age": "7d",
        "actions": {
          "allocate": {
            "number_of_replicas": 0
          }
        }
      },
      "cold": {
        "min_age": "30d",
        "actions": {
          "allocate": {
            "number_of_replicas": 0
          }
        }
      },
      "delete": {
        "min_age": "90d"
      }
    }
  }
}
```

## 告警配置

### 1. 告警规则

#### 服务可用性告警

```yaml
- alert: AlertAgentCoreDown
  expr: up{job="alertagent-core"} == 0
  for: 1m
  labels:
    severity: critical
  annotations:
    summary: "AlertAgent Core服务不可用"
    description: "服务已停止运行超过1分钟"
```

#### 性能告警

```yaml
- alert: HighLatency
  expr: histogram_quantile(0.95, rate(alertagent_http_request_duration_seconds_bucket[5m])) > 2
  for: 3m
  labels:
    severity: warning
  annotations:
    summary: "响应延迟过高"
    description: "95%分位延迟: {{ $value }}s"
```

#### 资源告警

```yaml
- alert: HighMemoryUsage
  expr: (container_memory_working_set_bytes / container_spec_memory_limit_bytes) > 0.8
  for: 5m
  labels:
    severity: warning
  annotations:
    summary: "内存使用率过高"
    description: "内存使用率: {{ $value | humanizePercentage }}"
```

### 2. 告警通知

#### 邮件通知

```yaml
receivers:
- name: 'email-alerts'
  email_configs:
  - to: 'admin@example.com'
    subject: 'AlertAgent 告警: {{ .GroupLabels.alertname }}'
    body: |
      {{ range .Alerts }}
      告警: {{ .Annotations.summary }}
      描述: {{ .Annotations.description }}
      时间: {{ .StartsAt }}
      {{ end }}
```

#### Webhook通知

```yaml
receivers:
- name: 'webhook-alerts'
  webhook_configs:
  - url: 'http://alertagent-core:8080/api/v1/webhooks/alertmanager'
    send_resolved: true
```

## 性能监控

### 1. 关键性能指标 (KPI)

#### 可用性指标

- **服务可用性**: 99.9%
- **平均故障恢复时间 (MTTR)**: < 15分钟
- **平均故障间隔时间 (MTBF)**: > 30天

#### 性能指标

- **响应时间**: 95%分位 < 2秒
- **吞吐量**: > 1000 RPS
- **错误率**: < 0.1%

#### 资源指标

- **CPU使用率**: < 70%
- **内存使用率**: < 80%
- **磁盘使用率**: < 80%

### 2. 性能基准测试

```bash
# HTTP性能测试
ab -n 10000 -c 100 http://alertagent-core:8080/api/v1/health

# 数据库性能测试
sysbench --test=oltp --mysql-host=mysql --mysql-user=root --mysql-password=password prepare
sysbench --test=oltp --mysql-host=mysql --mysql-user=root --mysql-password=password run

# Redis性能测试
redis-benchmark -h redis -p 6379 -n 100000 -c 50
```

### 3. 性能优化建议

#### 应用层优化

- 启用HTTP缓存
- 优化数据库查询
- 使用连接池
- 异步处理长时间任务

#### 基础设施优化

- 调整JVM参数
- 优化数据库配置
- 使用SSD存储
- 配置负载均衡

## 故障排查

### 1. 常见问题诊断

#### 服务无响应

```bash
# 检查Pod状态
kubectl get pods -n alertagent

# 查看Pod日志
kubectl logs -f deployment/alertagent-core -n alertagent

# 检查资源使用
kubectl top pods -n alertagent
```

#### 数据库连接问题

```bash
# 测试数据库连接
kubectl exec -it <mysql-pod> -n alertagent -- mysql -u root -p

# 检查连接数
kubectl exec -it <mysql-pod> -n alertagent -- mysql -u root -p -e "SHOW STATUS LIKE 'Threads_connected'"
```

#### 内存泄漏

```bash
# 查看内存使用趋势
kubectl top pods -n alertagent --sort-by=memory

# 生成内存dump
kubectl exec -it <pod-name> -n alertagent -- /usr/bin/pprof -http=:6060
```

### 2. 故障处理流程

1. **告警接收**: 通过监控系统接收告警
2. **问题确认**: 验证告警的真实性
3. **影响评估**: 评估故障影响范围
4. **应急处理**: 执行应急恢复措施
5. **根因分析**: 分析故障根本原因
6. **修复验证**: 验证修复效果
7. **总结改进**: 总结经验教训

### 3. 应急预案

#### 服务降级

```yaml
# 启用维护模式
apiVersion: v1
kind: ConfigMap
metadata:
  name: alertagent-config
data:
  maintenance_mode: "true"
  maintenance_message: "系统维护中，请稍后再试"
```

#### 数据备份恢复

```bash
# 数据库备份
kubectl exec -it <mysql-pod> -n alertagent -- mysqldump -u root -p alert_agent > backup.sql

# 数据库恢复
kubectl exec -i <mysql-pod> -n alertagent -- mysql -u root -p alert_agent < backup.sql
```

## 运维操作

### 1. 日常维护

#### 系统健康检查

```bash
# 执行健康检查
./scripts/monitoring-setup.sh health

# 检查磁盘空间
kubectl exec -it <pod-name> -n alertagent -- df -h

# 检查系统负载
kubectl top nodes
```

#### 日志清理

```bash
# 清理旧日志
kubectl exec -it <pod-name> -n alertagent -- find /app/logs -name "*.log" -mtime +7 -delete

# 清理Elasticsearch旧索引
curl -X DELETE "elasticsearch:9200/alertagent-*-$(date -d '30 days ago' +%Y.%m.%d)"
```

### 2. 容量规划

#### 资源监控

```promql
# CPU使用率趋势
avg(rate(container_cpu_usage_seconds_total{pod=~"alertagent-.*"}[5m])) by (pod)

# 内存使用率趋势
avg(container_memory_working_set_bytes{pod=~"alertagent-.*"}) by (pod)

# 存储使用率趋势
(1 - (node_filesystem_avail_bytes / node_filesystem_size_bytes)) * 100
```

#### 扩容建议

- **CPU使用率** > 70%: 考虑水平扩容
- **内存使用率** > 80%: 考虑垂直扩容
- **磁盘使用率** > 80%: 增加存储容量
- **网络带宽** > 80%: 优化网络配置

### 3. 安全运维

#### 安全检查

```bash
# 检查Pod安全上下文
kubectl get pods -n alertagent -o jsonpath='{.items[*].spec.securityContext}'

# 检查网络策略
kubectl get networkpolicy -n alertagent

# 检查RBAC权限
kubectl auth can-i --list --as=system:serviceaccount:alertagent:alertagent
```

#### 安全加固

- 定期更新镜像
- 启用Pod安全策略
- 配置网络策略
- 使用密钥管理
- 启用审计日志

### 4. 备份策略

#### 数据备份

```bash
# 自动备份脚本
#!/bin/bash
DATE=$(date +%Y%m%d_%H%M%S)
kubectl exec -it alertagent-mysql-0 -n alertagent -- mysqldump -u root -p alert_agent > backup_${DATE}.sql
aws s3 cp backup_${DATE}.sql s3://alertagent-backups/
```

#### 配置备份

```bash
# 备份Kubernetes配置
kubectl get all -n alertagent -o yaml > alertagent-backup.yaml
kubectl get configmaps -n alertagent -o yaml >> alertagent-backup.yaml
kubectl get secrets -n alertagent -o yaml >> alertagent-backup.yaml
```

## 最佳实践

### 1. 监控最佳实践

- 设置合理的告警阈值
- 避免告警风暴
- 定期审查告警规则
- 建立告警分级机制
- 实施告警静默策略

### 2. 日志最佳实践

- 使用结构化日志
- 设置合理的日志级别
- 避免敏感信息泄露
- 实施日志轮转策略
- 建立日志分析流程

### 3. 运维最佳实践

- 自动化运维操作
- 建立变更管理流程
- 实施蓝绿部署
- 定期进行灾难恢复演练
- 持续优化系统性能

## 参考资料

- [Prometheus官方文档](https://prometheus.io/docs/)
- [Grafana官方文档](https://grafana.com/docs/)
- [Elasticsearch官方文档](https://www.elastic.co/guide/)
- [Kubernetes监控最佳实践](https://kubernetes.io/docs/concepts/cluster-administration/monitoring/)
- [SRE实践指南](https://sre.google/books/)