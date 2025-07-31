#!/bin/bash

# AlertAgent 监控系统部署脚本
# 部署Prometheus、Grafana和日志收集系统

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# 日志函数
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 检查依赖
check_dependencies() {
    local missing_deps=()
    
    if ! command -v kubectl &> /dev/null; then
        missing_deps+=("kubectl")
    fi
    
    if ! command -v helm &> /dev/null; then
        missing_deps+=("helm")
    fi
    
    if [ ${#missing_deps[@]} -ne 0 ]; then
        log_error "缺少必需的依赖:"
        for dep in "${missing_deps[@]}"; do
            echo "  - $dep"
        done
        exit 1
    fi
    
    log_success "依赖检查通过"
}

# 添加Helm仓库
add_helm_repos() {
    log_info "添加Helm仓库..."
    
    # Prometheus社区仓库
    helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
    
    # Grafana仓库
    helm repo add grafana https://grafana.github.io/helm-charts
    
    # Elastic仓库
    helm repo add elastic https://helm.elastic.co
    
    # 更新仓库
    helm repo update
    
    log_success "Helm仓库添加完成"
}

# 创建监控命名空间
create_monitoring_namespace() {
    log_info "创建监控命名空间..."
    
    kubectl create namespace monitoring --dry-run=client -o yaml | kubectl apply -f -
    kubectl create namespace logging --dry-run=client -o yaml | kubectl apply -f -
    
    log_success "监控命名空间创建完成"
}

# 部署Prometheus
deploy_prometheus() {
    log_info "部署Prometheus..."
    
    # 创建Prometheus配置
    kubectl create configmap prometheus-config \
        --from-file=deploy/monitoring/prometheus-config.yaml \
        -n monitoring \
        --dry-run=client -o yaml | kubectl apply -f -
    
    # 创建告警规则
    kubectl create configmap prometheus-rules \
        --from-file=deploy/monitoring/alerting-rules.yml \
        -n monitoring \
        --dry-run=client -o yaml | kubectl apply -f -
    
    # 使用Helm部署Prometheus
    helm upgrade --install prometheus prometheus-community/kube-prometheus-stack \
        --namespace monitoring \
        --set prometheus.prometheusSpec.configMaps[0]=prometheus-config \
        --set prometheus.prometheusSpec.ruleSelector.matchLabels.app=prometheus \
        --set prometheus.prometheusSpec.retention=30d \
        --set prometheus.prometheusSpec.storageSpec.volumeClaimTemplate.spec.resources.requests.storage=50Gi \
        --set alertmanager.alertmanagerSpec.storage.volumeClaimTemplate.spec.resources.requests.storage=10Gi \
        --set grafana.enabled=true \
        --set grafana.adminPassword=admin123 \
        --set grafana.persistence.enabled=true \
        --set grafana.persistence.size=10Gi \
        --wait --timeout=600s
    
    log_success "Prometheus部署完成"
}

# 部署Grafana仪表板
deploy_grafana_dashboards() {
    log_info "部署Grafana仪表板..."
    
    # 等待Grafana就绪
    kubectl wait --for=condition=ready pod -l app.kubernetes.io/name=grafana -n monitoring --timeout=300s
    
    # 创建仪表板ConfigMap
    kubectl create configmap grafana-dashboard-alertagent \
        --from-file=deploy/monitoring/grafana-dashboard-alertagent.json \
        -n monitoring \
        --dry-run=client -o yaml | kubectl apply -f -
    
    # 标记为Grafana仪表板
    kubectl label configmap grafana-dashboard-alertagent \
        grafana_dashboard=1 \
        -n monitoring
    
    log_success "Grafana仪表板部署完成"
}

# 部署日志收集系统
deploy_logging() {
    log_info "部署日志收集系统..."
    
    # 部署Elasticsearch
    helm upgrade --install elasticsearch elastic/elasticsearch \
        --namespace logging \
        --set replicas=1 \
        --set minimumMasterNodes=1 \
        --set volumeClaimTemplate.resources.requests.storage=30Gi \
        --set resources.requests.cpu=500m \
        --set resources.requests.memory=2Gi \
        --set resources.limits.cpu=1000m \
        --set resources.limits.memory=4Gi \
        --wait --timeout=600s
    
    # 部署Kibana
    helm upgrade --install kibana elastic/kibana \
        --namespace logging \
        --set service.type=ClusterIP \
        --set resources.requests.cpu=200m \
        --set resources.requests.memory=1Gi \
        --wait --timeout=300s
    
    # 部署Fluentd
    kubectl apply -f deploy/logging/fluentd-config.yaml
    
    log_success "日志收集系统部署完成"
}

# 配置服务监控
configure_service_monitoring() {
    log_info "配置服务监控..."
    
    # 创建ServiceMonitor
    cat <<EOF | kubectl apply -f -
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: alertagent-monitor
  namespace: monitoring
  labels:
    app: alertagent
spec:
  selector:
    matchLabels:
      app.kubernetes.io/name: alertagent
  namespaceSelector:
    matchNames:
    - alertagent
  endpoints:
  - port: http
    path: /api/v1/metrics
    interval: 30s
  - port: health
    path: /metrics
    interval: 30s
EOF
    
    log_success "服务监控配置完成"
}

# 部署Node Exporter
deploy_node_exporter() {
    log_info "部署Node Exporter..."
    
    helm upgrade --install node-exporter prometheus-community/prometheus-node-exporter \
        --namespace monitoring \
        --set service.type=ClusterIP \
        --set service.port=9100 \
        --set service.targetPort=9100 \
        --wait --timeout=300s
    
    log_success "Node Exporter部署完成"
}

# 部署数据库监控
deploy_database_monitoring() {
    log_info "部署数据库监控..."
    
    # MySQL Exporter
    cat <<EOF | kubectl apply -f -
apiVersion: apps/v1
kind: Deployment
metadata:
  name: mysqld-exporter
  namespace: monitoring
  labels:
    app: mysqld-exporter
spec:
  replicas: 1
  selector:
    matchLabels:
      app: mysqld-exporter
  template:
    metadata:
      labels:
        app: mysqld-exporter
    spec:
      containers:
      - name: mysqld-exporter
        image: prom/mysqld-exporter:latest
        ports:
        - containerPort: 9104
          name: metrics
        env:
        - name: DATA_SOURCE_NAME
          value: "alertagent:password@(alertagent-mysql.alertagent.svc.cluster.local:3306)/"
        resources:
          requests:
            memory: 64Mi
            cpu: 50m
          limits:
            memory: 128Mi
            cpu: 100m
---
apiVersion: v1
kind: Service
metadata:
  name: mysqld-exporter
  namespace: monitoring
  labels:
    app: mysqld-exporter
spec:
  ports:
  - port: 9104
    targetPort: 9104
    name: metrics
  selector:
    app: mysqld-exporter
EOF
    
    # Redis Exporter
    cat <<EOF | kubectl apply -f -
apiVersion: apps/v1
kind: Deployment
metadata:
  name: redis-exporter
  namespace: monitoring
  labels:
    app: redis-exporter
spec:
  replicas: 1
  selector:
    matchLabels:
      app: redis-exporter
  template:
    metadata:
      labels:
        app: redis-exporter
    spec:
      containers:
      - name: redis-exporter
        image: oliver006/redis_exporter:latest
        ports:
        - containerPort: 9121
          name: metrics
        env:
        - name: REDIS_ADDR
          value: "redis://alertagent-redis.alertagent.svc.cluster.local:6379"
        - name: REDIS_PASSWORD
          valueFrom:
            secretKeyRef:
              name: alertagent-secrets
              key: redis-password
        resources:
          requests:
            memory: 64Mi
            cpu: 50m
          limits:
            memory: 128Mi
            cpu: 100m
---
apiVersion: v1
kind: Service
metadata:
  name: redis-exporter
  namespace: monitoring
  labels:
    app: redis-exporter
spec:
  ports:
  - port: 9121
    targetPort: 9121
    name: metrics
  selector:
    app: redis-exporter
EOF
    
    log_success "数据库监控部署完成"
}

# 配置告警通知
configure_alerting() {
    log_info "配置告警通知..."
    
    # 创建Alertmanager配置
    cat <<EOF | kubectl create secret generic alertmanager-config -n monitoring --from-literal=alertmanager.yml="
global:
  smtp_smarthost: 'smtp.example.com:587'
  smtp_from: 'alertmanager@example.com'
  smtp_auth_username: 'alertmanager@example.com'
  smtp_auth_password: 'password'

route:
  group_by: ['alertname']
  group_wait: 10s
  group_interval: 10s
  repeat_interval: 1h
  receiver: 'web.hook'

receivers:
- name: 'web.hook'
  email_configs:
  - to: 'admin@example.com'
    subject: 'AlertAgent 告警: {{ .GroupLabels.alertname }}'
    body: |
      {{ range .Alerts }}
      告警: {{ .Annotations.summary }}
      描述: {{ .Annotations.description }}
      时间: {{ .StartsAt }}
      {{ end }}
  webhook_configs:
  - url: 'http://alertagent-core.alertagent.svc.cluster.local:8080/api/v1/webhooks/alertmanager'
    send_resolved: true

inhibit_rules:
  - source_match:
      severity: 'critical'
    target_match:
      severity: 'warning'
    equal: ['alertname', 'dev', 'instance']
" --dry-run=client -o yaml | kubectl apply -f -
    
    log_success "告警通知配置完成"
}

# 显示访问信息
show_access_info() {
    log_info "监控系统访问信息:"
    echo
    
    # Prometheus访问信息
    echo "Prometheus:"
    echo "  - 端口转发: kubectl port-forward svc/prometheus-kube-prometheus-prometheus 9090:9090 -n monitoring"
    echo "  - 访问地址: http://localhost:9090"
    echo
    
    # Grafana访问信息
    echo "Grafana:"
    echo "  - 端口转发: kubectl port-forward svc/prometheus-grafana 3000:80 -n monitoring"
    echo "  - 访问地址: http://localhost:3000"
    echo "  - 用户名: admin"
    echo "  - 密码: admin123"
    echo
    
    # Kibana访问信息
    echo "Kibana:"
    echo "  - 端口转发: kubectl port-forward svc/kibana-kibana 5601:5601 -n logging"
    echo "  - 访问地址: http://localhost:5601"
    echo
    
    # AlertManager访问信息
    echo "AlertManager:"
    echo "  - 端口转发: kubectl port-forward svc/prometheus-kube-prometheus-alertmanager 9093:9093 -n monitoring"
    echo "  - 访问地址: http://localhost:9093"
}

# 健康检查
health_check() {
    log_info "执行监控系统健康检查..."
    
    # 检查Prometheus
    if kubectl get pods -n monitoring -l app.kubernetes.io/name=prometheus | grep -q Running; then
        log_success "Prometheus运行正常"
    else
        log_error "Prometheus运行异常"
    fi
    
    # 检查Grafana
    if kubectl get pods -n monitoring -l app.kubernetes.io/name=grafana | grep -q Running; then
        log_success "Grafana运行正常"
    else
        log_error "Grafana运行异常"
    fi
    
    # 检查Elasticsearch
    if kubectl get pods -n logging -l app=elasticsearch-master | grep -q Running; then
        log_success "Elasticsearch运行正常"
    else
        log_error "Elasticsearch运行异常"
    fi
    
    # 检查Fluentd
    if kubectl get pods -n logging -l app=fluentd | grep -q Running; then
        log_success "Fluentd运行正常"
    else
        log_error "Fluentd运行异常"
    fi
}

# 清理监控系统
cleanup() {
    log_info "清理监控系统..."
    
    # 删除Helm releases
    helm uninstall prometheus -n monitoring || true
    helm uninstall node-exporter -n monitoring || true
    helm uninstall elasticsearch -n logging || true
    helm uninstall kibana -n logging || true
    
    # 删除其他资源
    kubectl delete -f deploy/logging/fluentd-config.yaml || true
    kubectl delete deployment mysqld-exporter -n monitoring || true
    kubectl delete deployment redis-exporter -n monitoring || true
    kubectl delete service mysqld-exporter -n monitoring || true
    kubectl delete service redis-exporter -n monitoring || true
    
    # 删除命名空间
    kubectl delete namespace monitoring || true
    kubectl delete namespace logging || true
    
    log_success "监控系统清理完成"
}

# 显示帮助信息
show_help() {
    echo "AlertAgent 监控系统部署脚本"
    echo
    echo "用法: $0 <action> [options]"
    echo
    echo "操作:"
    echo "  deploy      完整部署监控系统"
    echo "  cleanup     清理监控系统"
    echo "  health      健康检查"
    echo "  info        显示访问信息"
    echo
    echo "示例:"
    echo "  $0 deploy           # 完整部署"
    echo "  $0 health           # 健康检查"
    echo "  $0 info             # 显示访问信息"
    echo "  $0 cleanup          # 清理部署"
    echo
    echo "选项:"
    echo "  -h, --help    显示此帮助信息"
}

# 主函数
main() {
    local action=$1
    
    if [ -z "$action" ]; then
        show_help
        exit 1
    fi
    
    case $action in
        "deploy")
            log_info "开始部署监控系统..."
            check_dependencies
            add_helm_repos
            create_monitoring_namespace
            deploy_prometheus
            deploy_grafana_dashboards
            deploy_logging
            configure_service_monitoring
            deploy_node_exporter
            deploy_database_monitoring
            configure_alerting
            health_check
            show_access_info
            log_success "监控系统部署完成！"
            ;;
        "cleanup")
            cleanup
            ;;
        "health")
            health_check
            ;;
        "info")
            show_access_info
            ;;
        *)
            log_error "不支持的操作: $action"
            show_help
            exit 1
            ;;
    esac
}

# 解析命令行参数
case "${1:-}" in
    -h|--help)
        show_help
        exit 0
        ;;
    *)
        main "$@"
        ;;
esac