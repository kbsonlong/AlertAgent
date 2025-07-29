# AlertAgent Helm Chart

这是 AlertAgent 的官方 Helm Chart，用于在 Kubernetes 集群中部署和管理 AlertAgent 系统。

## 概述

AlertAgent 是一个智能告警管理系统，提供以下核心功能：

- **告警聚合与去重**：智能聚合相似告警，减少告警噪音
- **AI 驱动分析**：使用 AI 分析告警模式和根因
- **多渠道通知**：支持钉钉、邮件、Webhook 等多种通知方式
- **规则引擎**：灵活的告警规则配置和管理
- **集群健康监控**：实时监控 Kubernetes 集群健康状态

## 架构组件

- **API 服务**：提供 RESTful API 接口
- **Worker 服务**：处理告警队列和后台任务
- **Rule Server**：规则引擎服务
- **PostgreSQL**：主数据库
- **Redis**：缓存和消息队列
- **MySQL**：规则引擎专用数据库
- **Prometheus**：监控指标收集
- **Grafana**：监控仪表板

## 快速开始

### 前置条件

- Kubernetes 1.20+
- Helm 3.8+
- 至少 4GB 可用内存
- 至少 2 CPU 核心
- 支持 ReadWriteOnce 的存储类

### 安装

1. **添加 Helm 仓库**

```bash
helm repo add bitnami https://charts.bitnami.com/bitnami
helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
helm repo add grafana https://grafana.github.io/helm-charts
helm repo update
```

2. **创建命名空间**

```bash
kubectl create namespace alertagent
```

3. **安装 AlertAgent**

```bash
# 开发环境
helm install alertagent ./alertagent \
  --namespace alertagent \
  --values ./alertagent/values-dev.yaml \
  --create-namespace \
  --wait

# 生产环境
helm install alertagent ./alertagent \
  --namespace alertagent \
  --values ./alertagent/values-prod.yaml \
  --create-namespace \
  --wait
```

4. **使用部署脚本（推荐）**

```bash
# 给脚本执行权限
chmod +x deploy.sh

# 部署开发环境
./deploy.sh dev install --wait --create-namespace

# 部署生产环境
./deploy.sh prod install --values values-prod.yaml --wait
```

### 验证安装

```bash
# 检查 Pod 状态
kubectl get pods -n alertagent

# 检查服务状态
kubectl get services -n alertagent

# 运行测试
helm test alertagent -n alertagent
```

## 配置说明

### 环境配置文件

- `values.yaml` - 默认配置
- `values-dev.yaml` - 开发环境配置
- `values-staging.yaml` - 测试环境配置
- `values-prod.yaml` - 生产环境配置

### 主要配置项

#### 全局配置

```yaml
global:
  imageRegistry: ""                    # 镜像仓库地址
  imagePullPolicy: IfNotPresent        # 镜像拉取策略
  storageClass: ""                     # 存储类
  namespace: "alertagent"              # 命名空间
```

#### API 服务配置

```yaml
api:
  enabled: true                        # 是否启用
  replicaCount: 3                      # 副本数
  image:
    repository: alertagent/api         # 镜像仓库
    tag: "v1.0.0"                      # 镜像标签
  service:
    type: ClusterIP                    # 服务类型
    port: 8080                         # 服务端口
  ingress:
    enabled: true                      # 是否启用 Ingress
    hosts:
      - host: api.alertagent.example.com
  resources:
    limits:
      cpu: 2000m
      memory: 2Gi
    requests:
      cpu: 500m
      memory: 512Mi
  autoscaling:
    enabled: true                      # 是否启用自动扩缩容
    minReplicas: 3
    maxReplicas: 10
```

#### 数据库配置

```yaml
postgresql:
  enabled: true                        # 是否启用内置 PostgreSQL
  auth:
    postgresPassword: "password"       # 管理员密码
    username: "alertagent"             # 应用用户名
    password: "password"               # 应用密码
    database: "alertagent"             # 数据库名
  primary:
    persistence:
      enabled: true                    # 是否启用持久化
      size: 100Gi                      # 存储大小
```

#### Redis 配置

```yaml
redis:
  enabled: true                        # 是否启用内置 Redis
  auth:
    enabled: true                      # 是否启用认证
    password: "password"               # Redis 密码
  architecture: replication            # 架构模式：standalone/replication
```

#### 监控配置

```yaml
prometheus:
  enabled: true                        # 是否启用 Prometheus
  server:
    retention: "30d"                   # 数据保留时间
    persistentVolume:
      size: 100Gi                      # 存储大小

grafana:
  enabled: true                        # 是否启用 Grafana
  adminUser: "admin"                   # 管理员用户名
  adminPassword: "password"            # 管理员密码
```

## 部署脚本使用

### 基本用法

```bash
./deploy.sh [环境] [操作] [选项]
```

### 环境选项

- `dev` - 开发环境（默认）
- `staging` - 测试环境
- `prod` - 生产环境

### 操作选项

- `install` - 安装（默认）
- `upgrade` - 升级
- `uninstall` - 卸载
- `test` - 运行测试
- `status` - 查看状态
- `logs` - 查看日志

### 常用命令示例

```bash
# 安装开发环境
./deploy.sh dev install --wait --create-namespace

# 升级生产环境
./deploy.sh prod upgrade --values values-prod.yaml

# 查看状态
./deploy.sh prod status

# 查看日志
./deploy.sh prod logs

# 运行测试
./deploy.sh dev test

# 卸载
./deploy.sh dev uninstall
```

### 高级选项

```bash
# 模拟运行
./deploy.sh prod install --dry-run

# 调试模式
./deploy.sh dev install --debug

# 自定义超时
./deploy.sh prod upgrade --timeout 15m

# 设置特定值
./deploy.sh prod install --set api.replicaCount=5

# 使用自定义 values 文件
./deploy.sh prod install --values my-values.yaml
```

## 升级指南

### 升级步骤

1. **备份数据**

```bash
# 备份 PostgreSQL
kubectl exec -n alertagent alertagent-postgresql-0 -- pg_dump -U alertagent alertagent > backup.sql

# 备份 Redis
kubectl exec -n alertagent alertagent-redis-master-0 -- redis-cli BGSAVE
```

2. **升级 Chart**

```bash
# 使用脚本升级
./deploy.sh prod upgrade --values values-prod.yaml --wait

# 或使用 Helm 命令
helm upgrade alertagent ./alertagent \
  --namespace alertagent \
  --values ./alertagent/values-prod.yaml \
  --wait
```

3. **验证升级**

```bash
# 检查状态
./deploy.sh prod status

# 运行测试
./deploy.sh prod test
```

### 回滚

```bash
# 查看历史版本
helm history alertagent -n alertagent

# 回滚到上一版本
helm rollback alertagent -n alertagent

# 回滚到指定版本
helm rollback alertagent 2 -n alertagent
```

## 监控和告警

### 访问监控界面

1. **Grafana**

```bash
# 获取 Grafana URL
echo "http://$(kubectl get ingress alertagent-grafana -n alertagent -o jsonpath='{.spec.rules[0].host}')"

# 获取管理员密码
kubectl get secret alertagent-grafana -n alertagent -o jsonpath="{.data.admin-password}" | base64 -d
```

2. **Prometheus**

```bash
# 获取 Prometheus URL
echo "http://$(kubectl get ingress alertagent-prometheus -n alertagent -o jsonpath='{.spec.rules[0].host}')"
```

### 内置仪表板

- **AlertAgent Overview** - 系统概览
- **AlertAgent Business** - 业务指标
- **AlertAgent Database** - 数据库监控
- **AlertAgent Infrastructure** - 基础设施监控

### 告警规则

系统包含以下预定义告警规则：

- 服务健康检查
- HTTP 错误率和延迟
- 资源使用率
- 数据库连接问题
- 告警队列积压
- Pod 重启频繁

## 故障排除

### 常见问题

1. **Pod 启动失败**

```bash
# 查看 Pod 状态
kubectl get pods -n alertagent

# 查看 Pod 详情
kubectl describe pod <pod-name> -n alertagent

# 查看日志
kubectl logs <pod-name> -n alertagent
```

2. **数据库连接问题**

```bash
# 检查数据库 Pod
kubectl get pods -n alertagent -l app.kubernetes.io/component=primary

# 测试数据库连接
kubectl exec -it alertagent-postgresql-0 -n alertagent -- psql -U alertagent -d alertagent -c "SELECT version();"
```

3. **Redis 连接问题**

```bash
# 检查 Redis Pod
kubectl get pods -n alertagent -l app.kubernetes.io/component=master

# 测试 Redis 连接
kubectl exec -it alertagent-redis-master-0 -n alertagent -- redis-cli ping
```

4. **Ingress 访问问题**

```bash
# 检查 Ingress 状态
kubectl get ingress -n alertagent

# 检查 Ingress Controller
kubectl get pods -n ingress-nginx
```

### 调试命令

```bash
# 查看所有资源
kubectl get all -n alertagent

# 查看事件
kubectl get events -n alertagent --sort-by='.lastTimestamp'

# 查看配置
kubectl get configmap alertagent-config -n alertagent -o yaml

# 查看密钥（小心！）
kubectl get secrets -n alertagent

# 端口转发调试
kubectl port-forward svc/alertagent-api 8080:8080 -n alertagent
```

## 安全考虑

### 生产环境安全清单

- [ ] 更改所有默认密码
- [ ] 启用 TLS/SSL
- [ ] 配置网络策略
- [ ] 启用 Pod 安全策略
- [ ] 配置 RBAC
- [ ] 启用审计日志
- [ ] 定期备份数据
- [ ] 监控安全事件

### 密钥管理

建议在生产环境中使用外部密钥管理系统：

- AWS Secrets Manager
- Azure Key Vault
- HashiCorp Vault
- Kubernetes External Secrets Operator

## 性能调优

### 资源配置建议

#### 小型部署（< 1000 告警/天）

```yaml
api:
  replicaCount: 2
  resources:
    requests: { cpu: 200m, memory: 256Mi }
    limits: { cpu: 500m, memory: 512Mi }

worker:
  replicaCount: 2
  resources:
    requests: { cpu: 100m, memory: 128Mi }
    limits: { cpu: 300m, memory: 256Mi }
```

#### 中型部署（1000-10000 告警/天）

```yaml
api:
  replicaCount: 3
  resources:
    requests: { cpu: 500m, memory: 512Mi }
    limits: { cpu: 1000m, memory: 1Gi }

worker:
  replicaCount: 5
  resources:
    requests: { cpu: 200m, memory: 256Mi }
    limits: { cpu: 500m, memory: 512Mi }
```

#### 大型部署（> 10000 告警/天）

```yaml
api:
  replicaCount: 5
  resources:
    requests: { cpu: 1000m, memory: 1Gi }
    limits: { cpu: 2000m, memory: 2Gi }

worker:
  replicaCount: 10
  resources:
    requests: { cpu: 500m, memory: 512Mi }
    limits: { cpu: 1000m, memory: 1Gi }
```

### 数据库优化

```yaml
postgresql:
  primary:
    resources:
      requests: { cpu: 500m, memory: 1Gi }
      limits: { cpu: 2000m, memory: 4Gi }
    initdb:
      scripts:
        tune.sql: |
          ALTER SYSTEM SET shared_buffers = '1GB';
          ALTER SYSTEM SET effective_cache_size = '3GB';
          ALTER SYSTEM SET maintenance_work_mem = '256MB';
          ALTER SYSTEM SET max_connections = 200;
```

## 贡献指南

### 开发环境设置

1. **克隆仓库**

```bash
git clone https://github.com/your-org/alertagent.git
cd alertagent/deployments/helm
```

2. **安装开发依赖**

```bash
# 安装 Helm
curl https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3 | bash

# 安装 helm-docs
go install github.com/norwoodj/helm-docs/cmd/helm-docs@latest

# 安装 kubeval
go install github.com/instrumenta/kubeval@latest
```

3. **验证 Chart**

```bash
# Lint Chart
helm lint ./alertagent

# 模板渲染测试
helm template alertagent ./alertagent --values ./alertagent/values-dev.yaml

# 验证 Kubernetes 资源
helm template alertagent ./alertagent | kubeval
```

### 提交更改

1. 更新 Chart 版本
2. 运行测试
3. 更新文档
4. 提交 Pull Request

## 许可证

MIT License - 详见 [LICENSE](../../LICENSE) 文件。

## 支持

- **文档**: https://alertagent.docs.example.com
- **Issues**: https://github.com/your-org/alertagent/issues
- **讨论**: https://github.com/your-org/alertagent/discussions
- **邮件**: support@example.com

## 更新日志

### v1.0.0

- 初始版本发布
- 支持 API、Worker、Rule Server 部署
- 集成 PostgreSQL、Redis、MySQL
- 内置 Prometheus 和 Grafana 监控
- 支持多环境配置
- 包含完整的 RBAC 和网络策略

---

**注意**: 在生产环境中部署前，请仔细阅读安全考虑部分，并根据实际需求调整配置。