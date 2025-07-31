# AlertAgent Kubernetes 部署指南

本文档详细介绍了如何在Kubernetes集群中部署AlertAgent系统。

## 目录

- [系统架构](#系统架构)
- [前置要求](#前置要求)
- [快速部署](#快速部署)
- [配置说明](#配置说明)
- [服务发现和负载均衡](#服务发现和负载均衡)
- [自动扩缩容](#自动扩缩容)
- [监控和日志](#监控和日志)
- [故障排查](#故障排查)
- [升级和维护](#升级和维护)

## 系统架构

AlertAgent在Kubernetes中的部署架构：

```
┌─────────────────────────────────────────────────────────────┐
│                    Kubernetes 集群                          │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐         │
│  │   Ingress   │  │ LoadBalancer│  │ NetworkPolicy│        │
│  │ Controller  │  │   Service   │  │   (安全)    │         │
│  └─────────────┘  └─────────────┘  └─────────────┘         │
│                                                             │
│  ┌─────────────────────────────────────────────────────────┐ │
│  │                AlertAgent Namespace                     │ │
│  │                                                         │ │
│  │  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐     │ │
│  │  │AlertAgent   │  │   Worker    │  │   Worker    │     │ │
│  │  │    Core     │  │ AI-Analysis │  │Notification │     │ │
│  │  │ (2 replicas)│  │(2-10 replicas)│(2-8 replicas)│     │ │
│  │  └─────────────┘  └─────────────┘  └─────────────┘     │ │
│  │                                                         │ │
│  │  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐     │ │
│  │  │   Worker    │  │    MySQL    │  │    Redis    │     │ │
│  │  │Config-Sync  │  │ (StatefulSet)│ (StatefulSet) │     │ │
│  │  │(1 replica)  │  │             │  │             │     │ │
│  │  └─────────────┘  └─────────────┘  └─────────────┘     │ │
│  │                                                         │ │
│  │  ┌─────────────┐  ┌─────────────┐                      │ │
│  │  │   Ollama    │  │ ConfigMaps  │                      │ │
│  │  │ (Optional)  │  │ & Secrets   │                      │ │
│  │  └─────────────┘  └─────────────┘                      │ │
│  └─────────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────┘
```

## 前置要求

### 1. Kubernetes集群

- Kubernetes 1.20+
- 至少3个节点 (推荐)
- 支持PersistentVolume
- 网络插件 (如Calico, Flannel)

### 2. 必需组件

```bash
# 检查集群状态
kubectl cluster-info

# 检查节点状态
kubectl get nodes

# 检查存储类
kubectl get storageclass
```

### 3. 可选组件

- **Ingress Controller**: Nginx Ingress Controller
- **证书管理**: cert-manager
- **监控系统**: Prometheus + Grafana
- **日志收集**: ELK Stack 或 Loki

### 4. 客户端工具

- kubectl 1.20+
- Docker (用于构建镜像)
- Helm 3.0+ (可选)

## 快速部署

### 1. 准备镜像

```bash
# 构建Docker镜像
./scripts/docker-build.sh

# 推送到镜像仓库 (如果使用私有仓库)
docker tag alertagent-core:latest your-registry/alertagent-core:latest
docker push your-registry/alertagent-core:latest

docker tag alertagent-worker:latest your-registry/alertagent-worker:latest
docker push your-registry/alertagent-worker:latest

docker tag alertagent-sidecar:latest your-registry/alertagent-sidecar:latest
docker push your-registry/alertagent-sidecar:latest
```

### 2. 一键部署

```bash
# 执行完整部署
./scripts/k8s-deploy.sh deploy
```

### 3. 验证部署

```bash
# 检查部署状态
./scripts/k8s-deploy.sh status

# 执行健康检查
./scripts/k8s-deploy.sh health
```

## 配置说明

### 1. 命名空间配置

```yaml
# deploy/k8s/namespace.yaml
apiVersion: v1
kind: Namespace
metadata:
  name: alertagent
  labels:
    name: alertagent
```

### 2. 密钥管理

```bash
# 查看密钥
kubectl get secrets -n alertagent

# 更新密钥
kubectl create secret generic alertagent-secrets \
  --from-literal=mysql-password=new-password \
  --from-literal=redis-password=new-password \
  --from-literal=jwt-secret=new-jwt-secret \
  -n alertagent \
  --dry-run=client -o yaml | kubectl apply -f -
```

### 3. 配置文件管理

```bash
# 查看配置
kubectl get configmap alertagent-config -n alertagent -o yaml

# 更新配置
kubectl edit configmap alertagent-config -n alertagent

# 重启服务以应用新配置
kubectl rollout restart deployment -n alertagent
```

### 4. 持久化存储

```yaml
# PVC配置示例
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: mysql-pvc
  namespace: alertagent
spec:
  accessModes:
    - ReadWriteOnce
  resources:
    requests:
      storage: 20Gi
  storageClassName: fast-ssd  # 根据集群调整
```

## 服务发现和负载均衡

### 1. 内部服务发现

```yaml
# Headless Service用于服务发现
apiVersion: v1
kind: Service
metadata:
  name: alertagent-core-headless
  namespace: alertagent
spec:
  clusterIP: None
  selector:
    app.kubernetes.io/component: core
```

### 2. 负载均衡配置

```yaml
# LoadBalancer Service
apiVersion: v1
kind: Service
metadata:
  name: alertagent-lb
  namespace: alertagent
spec:
  type: LoadBalancer
  sessionAffinity: ClientIP
  selector:
    app.kubernetes.io/component: core
```

### 3. Ingress配置

```yaml
# Ingress配置
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: alertagent-ingress
  namespace: alertagent
  annotations:
    nginx.ingress.kubernetes.io/rate-limit: "100"
    nginx.ingress.kubernetes.io/ssl-redirect: "true"
spec:
  rules:
  - host: alertagent.example.com
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: alertagent-core
            port:
              number: 8080
```

## 自动扩缩容

### 1. Horizontal Pod Autoscaler (HPA)

```yaml
# AI Worker HPA配置
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: alertagent-worker-ai-hpa
  namespace: alertagent
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: alertagent-worker-ai
  minReplicas: 2
  maxReplicas: 10
  metrics:
  - type: Resource
    resource:
      name: cpu
      target:
        type: Utilization
        averageUtilization: 70
  - type: Resource
    resource:
      name: memory
      target:
        type: Utilization
        averageUtilization: 80
```

### 2. Vertical Pod Autoscaler (VPA)

```yaml
# VPA配置 (可选)
apiVersion: autoscaling.k8s.io/v1
kind: VerticalPodAutoscaler
metadata:
  name: alertagent-core-vpa
  namespace: alertagent
spec:
  targetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: alertagent-core
  updatePolicy:
    updateMode: "Auto"
```

### 3. 集群自动扩缩容

```yaml
# Cluster Autoscaler配置
apiVersion: v1
kind: ConfigMap
metadata:
  name: cluster-autoscaler-status
  namespace: kube-system
data:
  nodes.max: "10"
  nodes.min: "3"
```

## 监控和日志

### 1. Prometheus监控

```yaml
# ServiceMonitor配置
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: alertagent-monitor
  namespace: alertagent
spec:
  selector:
    matchLabels:
      app.kubernetes.io/name: alertagent
  endpoints:
  - port: http
    path: /api/v1/metrics
    interval: 30s
```

### 2. 日志收集

```bash
# 查看实时日志
kubectl logs -f deployment/alertagent-core -n alertagent

# 查看所有Pod日志
kubectl logs -l app.kubernetes.io/name=alertagent -n alertagent

# 使用脚本查看日志
./scripts/k8s-deploy.sh logs core true
```

### 3. 告警规则

```yaml
# PrometheusRule配置
apiVersion: monitoring.coreos.com/v1
kind: PrometheusRule
metadata:
  name: alertagent-rules
  namespace: alertagent
spec:
  groups:
  - name: alertagent
    rules:
    - alert: AlertAgentDown
      expr: up{job="alertagent"} == 0
      for: 1m
      labels:
        severity: critical
      annotations:
        summary: "AlertAgent服务不可用"
```

## 故障排查

### 1. 常见问题

#### Pod启动失败

```bash
# 查看Pod状态
kubectl get pods -n alertagent

# 查看Pod详细信息
kubectl describe pod <pod-name> -n alertagent

# 查看Pod日志
kubectl logs <pod-name> -n alertagent
```

#### 服务连接问题

```bash
# 测试服务连接
kubectl run test-pod --image=busybox -it --rm -- /bin/sh
# 在Pod内执行
nslookup alertagent-core.alertagent.svc.cluster.local
wget -qO- http://alertagent-core.alertagent.svc.cluster.local:8080/api/v1/health
```

#### 存储问题

```bash
# 查看PVC状态
kubectl get pvc -n alertagent

# 查看PV状态
kubectl get pv

# 查看存储类
kubectl get storageclass
```

### 2. 调试工具

```bash
# 进入Pod进行调试
kubectl exec -it <pod-name> -n alertagent -- /bin/sh

# 端口转发进行本地调试
kubectl port-forward service/alertagent-core 8080:8080 -n alertagent

# 查看资源使用情况
kubectl top pods -n alertagent
kubectl top nodes
```

### 3. 网络问题

```bash
# 查看网络策略
kubectl get networkpolicy -n alertagent

# 测试网络连通性
kubectl run netshoot --image=nicolaka/netshoot -it --rm -- /bin/bash
```

## 升级和维护

### 1. 滚动更新

```bash
# 更新镜像
kubectl set image deployment/alertagent-core alertagent-core=alertagent-core:v2.0.0 -n alertagent

# 查看更新状态
kubectl rollout status deployment/alertagent-core -n alertagent

# 回滚更新
kubectl rollout undo deployment/alertagent-core -n alertagent
```

### 2. 配置更新

```bash
# 更新ConfigMap
kubectl patch configmap alertagent-config -n alertagent --patch '{"data":{"key":"new-value"}}'

# 重启Deployment以应用新配置
kubectl rollout restart deployment/alertagent-core -n alertagent
```

### 3. 数据备份

```bash
# MySQL备份
kubectl exec -it <mysql-pod> -n alertagent -- mysqldump -u root -p alert_agent > backup.sql

# Redis备份
kubectl exec -it <redis-pod> -n alertagent -- redis-cli BGSAVE
```

### 4. 集群维护

```bash
# 驱逐节点上的Pod
kubectl drain <node-name> --ignore-daemonsets --delete-emptydir-data

# 标记节点为不可调度
kubectl cordon <node-name>

# 恢复节点调度
kubectl uncordon <node-name>
```

## 安全最佳实践

### 1. RBAC配置

```yaml
# 最小权限原则
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  namespace: alertagent
  name: alertagent-role
rules:
- apiGroups: [""]
  resources: ["configmaps", "secrets"]
  verbs: ["get", "list", "watch"]
```

### 2. 网络策略

```yaml
# 限制网络访问
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: alertagent-network-policy
  namespace: alertagent
spec:
  podSelector:
    matchLabels:
      app.kubernetes.io/name: alertagent
  policyTypes:
  - Ingress
  - Egress
```

### 3. Pod安全策略

```yaml
# Pod安全标准
apiVersion: v1
kind: Pod
spec:
  securityContext:
    runAsNonRoot: true
    runAsUser: 1001
    fsGroup: 1001
  containers:
  - name: alertagent
    securityContext:
      allowPrivilegeEscalation: false
      readOnlyRootFilesystem: true
      capabilities:
        drop:
        - ALL
```

## 性能优化

### 1. 资源配置

```yaml
# 合理的资源请求和限制
resources:
  requests:
    memory: "512Mi"
    cpu: "250m"
  limits:
    memory: "1Gi"
    cpu: "1000m"
```

### 2. 节点亲和性

```yaml
# 节点选择器
nodeSelector:
  node-type: compute

# 节点亲和性
affinity:
  nodeAffinity:
    requiredDuringSchedulingIgnoredDuringExecution:
      nodeSelectorTerms:
      - matchExpressions:
        - key: node-type
          operator: In
          values:
          - compute
```

### 3. Pod反亲和性

```yaml
# 避免单点故障
affinity:
  podAntiAffinity:
    preferredDuringSchedulingIgnoredDuringExecution:
    - weight: 100
      podAffinityTerm:
        labelSelector:
          matchExpressions:
          - key: app.kubernetes.io/component
            operator: In
            values:
            - core
        topologyKey: kubernetes.io/hostname
```

## 参考资料

- [Kubernetes官方文档](https://kubernetes.io/docs/)
- [Helm Charts最佳实践](https://helm.sh/docs/chart_best_practices/)
- [Prometheus Operator](https://github.com/prometheus-operator/prometheus-operator)
- [Nginx Ingress Controller](https://kubernetes.github.io/ingress-nginx/)
- [cert-manager](https://cert-manager.io/docs/)