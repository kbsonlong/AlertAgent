# 配置同步Sidecar容器

配置同步Sidecar容器是AlertAgent架构中的核心组件，负责从AlertAgent中央服务拉取配置并同步到Prometheus和Alertmanager实例。

## 功能特性

### 核心功能
- **统一配置拉取**: 从AlertAgent API拉取最新的配置文件
- **哈希比较**: 通过SHA256哈希值检测配置变化，避免不必要的重载
- **原子性写入**: 使用临时文件确保配置写入的原子性
- **热重载**: 自动触发Prometheus/Alertmanager的配置重载
- **错误处理**: 指数退避重试机制，确保高可用性
- **状态报告**: 提供详细的同步状态和指标

### 监控和健康检查
- **健康检查端点**: `/health`, `/healthz`, `/ready`
- **指标端点**: `/metrics`, `/status`
- **结构化日志**: 使用zap提供详细的操作日志
- **Prometheus指标**: 同步次数、成功率、错误统计等

## 架构设计

```
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│   AlertAgent    │    │  Config Syncer   │    │ Prometheus/     │
│   (Central)     │◄───┤   (Sidecar)      │───►│ Alertmanager    │
│                 │    │                  │    │                 │
└─────────────────┘    └──────────────────┘    └─────────────────┘
        │                        │                        │
        │                        │                        │
        ▼                        ▼                        ▼
   配置管理API              配置同步逻辑              配置热重载
```

## 配置参数

### 必需环境变量

| 变量名 | 描述 | 示例 |
|--------|------|------|
| `ALERTAGENT_ENDPOINT` | AlertAgent API端点 | `http://alertagent-service:8080` |
| `CLUSTER_ID` | 集群标识符 | `prometheus-cluster-1` |
| `CONFIG_TYPE` | 配置类型 | `prometheus`, `alertmanager`, `prometheus-rules` |
| `CONFIG_PATH` | 配置文件路径 | `/etc/prometheus/prometheus.yml` |
| `RELOAD_URL` | 重载端点URL | `http://localhost:9090/-/reload` |

### 可选环境变量

| 变量名 | 默认值 | 描述 |
|--------|--------|------|
| `SYNC_INTERVAL` | `30s` | 同步间隔 |
| `HTTP_TIMEOUT` | `30s` | HTTP请求超时 |
| `HTTP_PORT` | `8080` | 健康检查服务端口 |
| `MAX_RETRIES` | `3` | 最大重试次数 |
| `RETRY_BACKOFF` | `5s` | 重试退避基础时间 |
| `APP_ENV` | `production` | 应用环境 |

## 部署方式

### 1. Kubernetes Sidecar模式

#### Prometheus + Config Syncer

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: prometheus
spec:
  template:
    spec:
      containers:
      # Prometheus主容器
      - name: prometheus
        image: prom/prometheus:v2.45.0
        # ... prometheus配置
      
      # 配置同步Sidecar
      - name: config-syncer
        image: alertagent/config-syncer:latest
        env:
        - name: ALERTAGENT_ENDPOINT
          value: "http://alertagent-service:8080"
        - name: CLUSTER_ID
          value: "prometheus-cluster-1"
        - name: CONFIG_TYPE
          value: "prometheus"
        - name: CONFIG_PATH
          value: "/etc/prometheus/prometheus.yml"
        - name: RELOAD_URL
          value: "http://localhost:9090/-/reload"
        volumeMounts:
        - name: config-volume
          mountPath: /etc/prometheus
```

#### Alertmanager + Config Syncer

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: alertmanager
spec:
  template:
    spec:
      containers:
      # Alertmanager主容器
      - name: alertmanager
        image: prom/alertmanager:v0.26.0
        # ... alertmanager配置
      
      # 配置同步Sidecar
      - name: config-syncer
        image: alertagent/config-syncer:latest
        env:
        - name: ALERTAGENT_ENDPOINT
          value: "http://alertagent-service:8080"
        - name: CLUSTER_ID
          value: "alertmanager-cluster-1"
        - name: CONFIG_TYPE
          value: "alertmanager"
        - name: CONFIG_PATH
          value: "/etc/alertmanager/alertmanager.yml"
        - name: RELOAD_URL
          value: "http://localhost:9093/-/reload"
        volumeMounts:
        - name: config-volume
          mountPath: /etc/alertmanager
```

### 2. Docker Compose

```yaml
version: '3.8'
services:
  prometheus:
    image: prom/prometheus:v2.45.0
    volumes:
      - prometheus-config:/etc/prometheus
    
  config-syncer:
    image: alertagent/config-syncer:latest
    environment:
      - ALERTAGENT_ENDPOINT=http://alertagent:8080
      - CLUSTER_ID=prometheus-cluster-1
      - CONFIG_TYPE=prometheus
      - CONFIG_PATH=/etc/prometheus/prometheus.yml
      - RELOAD_URL=http://prometheus:9090/-/reload
      - SYNC_INTERVAL=30s
    volumes:
      - prometheus-config:/etc/prometheus
    depends_on:
      - alertagent

volumes:
  prometheus-config:
```

## API端点

### 健康检查

#### GET /health
返回服务健康状态

```json
{
  "status": "healthy",
  "timestamp": "2024-01-15T10:30:00Z",
  "uptime": "2h30m15s",
  "version": "1.0.0",
  "cluster_id": "prometheus-cluster-1",
  "config_type": "prometheus",
  "last_sync": "2024-01-15T10:29:30Z",
  "error": ""
}
```

#### GET /ready
返回服务就绪状态

```json
{
  "ready": true,
  "timestamp": "2024-01-15T10:30:00Z",
  "sync_count": 120,
  "uptime": "2h30m15s"
}
```

### 指标监控

#### GET /metrics
返回详细的同步指标

```json
{
  "last_sync_time": "2024-01-15T10:29:30Z",
  "sync_count": 120,
  "success_count": 118,
  "failure_count": 2,
  "last_error": "",
  "config_hash": "a1b2c3d4e5f6...",
  "config_version": "v1.2.3",
  "retry_count": 0,
  "next_sync_time": "2024-01-15T10:30:30Z",
  "healthy": true,
  "uptime": "2h30m15s"
}
```

#### GET /status
返回完整的状态信息（健康状态 + 指标）

## 构建和部署

### 本地构建

```bash
# 构建二进制文件
go build -o config-syncer ./cmd/config-syncer

# 运行
./config-syncer
```

### Docker构建

```bash
# 使用构建脚本
./scripts/build-config-syncer.sh

# 或手动构建
docker build -f build/config-syncer/Dockerfile -t alertagent/config-syncer:latest .
```

### 构建脚本选项

```bash
# 基本构建
./scripts/build-config-syncer.sh

# 指定标签
./scripts/build-config-syncer.sh -t v1.0.0

# 构建并推送
./scripts/build-config-syncer.sh -t v1.0.0 --push

# 跳过测试
./scripts/build-config-syncer.sh --skip-tests
```

## 监控和告警

### Prometheus监控

可以通过以下指标监控配置同步器的状态：

```yaml
# prometheus.yml
scrape_configs:
- job_name: 'config-syncer'
  static_configs:
  - targets: ['config-syncer:8080']
  metrics_path: '/metrics'
  scrape_interval: 30s
```

### 告警规则

```yaml
# alerts.yml
groups:
- name: config-syncer
  rules:
  - alert: ConfigSyncerDown
    expr: up{job="config-syncer"} == 0
    for: 1m
    labels:
      severity: critical
    annotations:
      summary: "Config syncer is down"
      description: "Config syncer for {{ $labels.cluster_id }} has been down for more than 1 minute."
  
  - alert: ConfigSyncFailure
    expr: increase(config_sync_failures_total[5m]) > 3
    for: 2m
    labels:
      severity: warning
    annotations:
      summary: "Config sync failures detected"
      description: "Config syncer for {{ $labels.cluster_id }} has failed {{ $value }} times in the last 5 minutes."
```

## 故障排除

### 常见问题

#### 1. 配置同步失败

**症状**: 日志显示HTTP请求失败

**解决方案**:
- 检查`ALERTAGENT_ENDPOINT`是否正确
- 验证网络连接
- 检查AlertAgent服务状态

#### 2. 热重载失败

**症状**: 配置更新但服务未重载

**解决方案**:
- 检查`RELOAD_URL`是否正确
- 验证目标服务的重载端点
- 检查权限和网络访问

#### 3. 健康检查失败

**症状**: Kubernetes显示容器不健康

**解决方案**:
- 检查HTTP服务端口配置
- 验证健康检查端点响应
- 查看容器日志

### 日志分析

```bash
# 查看容器日志
kubectl logs -f deployment/prometheus -c config-syncer

# 查看特定错误
kubectl logs deployment/prometheus -c config-syncer | grep ERROR

# 查看同步状态
curl http://config-syncer:8080/status
```

## 安全考虑

### 网络安全
- 使用HTTPS连接AlertAgent API
- 限制网络访问权限
- 配置适当的防火墙规则

### 容器安全
- 使用非root用户运行
- 最小化容器权限
- 定期更新基础镜像

### 配置安全
- 避免在环境变量中存储敏感信息
- 使用Kubernetes Secrets管理凭据
- 启用配置加密传输

## 性能优化

### 资源配置

```yaml
resources:
  requests:
    memory: "64Mi"
    cpu: "50m"
  limits:
    memory: "128Mi"
    cpu: "100m"
```

### 同步优化
- 根据配置变更频率调整`SYNC_INTERVAL`
- 合理设置重试参数
- 监控网络延迟和响应时间

## 版本兼容性

| Config Syncer版本 | AlertAgent版本 | Prometheus版本 | Alertmanager版本 |
|-------------------|----------------|----------------|------------------|
| v1.0.x | v1.0.x+ | v2.40+ | v0.25+ |
| v1.1.x | v1.1.x+ | v2.45+ | v0.26+ |

## 贡献指南

### 开发环境设置

```bash
# 克隆项目
git clone https://github.com/your-org/alertagent.git
cd alertagent

# 安装依赖
go mod download

# 运行测试
go test ./pkg/config-syncer/...

# 本地运行
go run ./cmd/config-syncer
```

### 提交规范
- 遵循Go代码规范
- 添加适当的测试用例
- 更新相关文档
- 提交前运行完整测试套件

## 许可证

本项目采用MIT许可证，详见[LICENSE](../../LICENSE)文件。