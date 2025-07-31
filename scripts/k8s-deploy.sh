#!/bin/bash

# AlertAgent Kubernetes部署脚本
# 支持多环境部署和管理

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
    
    if ! command -v docker &> /dev/null; then
        missing_deps+=("docker")
    fi
    
    if [ ${#missing_deps[@]} -ne 0 ]; then
        log_error "缺少必需的依赖:"
        for dep in "${missing_deps[@]}"; do
            echo "  - $dep"
        done
        exit 1
    fi
    
    # 检查kubectl连接
    if ! kubectl cluster-info &> /dev/null; then
        log_error "无法连接到Kubernetes集群，请检查kubeconfig配置"
        exit 1
    fi
    
    log_success "依赖检查通过"
}

# 创建命名空间
create_namespace() {
    log_info "创建命名空间..."
    
    if kubectl get namespace alertagent &> /dev/null; then
        log_info "命名空间 alertagent 已存在"
    else
        kubectl apply -f deploy/k8s/namespace.yaml
        log_success "命名空间创建完成"
    fi
}

# 创建密钥
create_secrets() {
    log_info "创建密钥配置..."
    
    # 检查是否已存在密钥
    if kubectl get secret alertagent-secrets -n alertagent &> /dev/null; then
        log_warning "密钥已存在，跳过创建"
        return
    fi
    
    # 生成随机密码
    local mysql_password=$(openssl rand -base64 32 | tr -d "=+/" | cut -c1-25)
    local redis_password=$(openssl rand -base64 32 | tr -d "=+/" | cut -c1-25)
    local jwt_secret=$(openssl rand -base64 64 | tr -d "=+/" | cut -c1-50)
    
    # 创建密钥
    kubectl create secret generic alertagent-secrets \
        --from-literal=mysql-root-password="$mysql_password" \
        --from-literal=mysql-password="$mysql_password" \
        --from-literal=redis-password="$redis_password" \
        --from-literal=jwt-secret="$jwt_secret" \
        -n alertagent
    
    log_success "密钥创建完成"
    log_warning "请妥善保存以下密码信息:"
    echo "  MySQL密码: $mysql_password"
    echo "  Redis密码: $redis_password"
    echo "  JWT密钥: $jwt_secret"
}

# 部署基础服务
deploy_infrastructure() {
    log_info "部署基础设施服务..."
    
    # 部署ConfigMap
    kubectl apply -f deploy/k8s/configmap.yaml
    log_info "ConfigMap部署完成"
    
    # 部署MySQL
    kubectl apply -f deploy/k8s/mysql.yaml
    log_info "MySQL部署完成"
    
    # 部署Redis
    kubectl apply -f deploy/k8s/redis.yaml
    log_info "Redis部署完成"
    
    log_success "基础设施服务部署完成"
}

# 等待服务就绪
wait_for_services() {
    log_info "等待基础服务就绪..."
    
    # 等待MySQL就绪
    log_info "等待MySQL服务就绪..."
    kubectl wait --for=condition=ready pod -l app.kubernetes.io/component=mysql -n alertagent --timeout=300s
    
    # 等待Redis就绪
    log_info "等待Redis服务就绪..."
    kubectl wait --for=condition=ready pod -l app.kubernetes.io/component=redis -n alertagent --timeout=300s
    
    log_success "基础服务已就绪"
}

# 部署应用服务
deploy_application() {
    log_info "部署应用服务..."
    
    # 部署AlertAgent Core
    kubectl apply -f deploy/k8s/alertagent-core.yaml
    log_info "AlertAgent Core部署完成"
    
    # 部署Worker
    kubectl apply -f deploy/k8s/alertagent-worker.yaml
    log_info "AlertAgent Worker部署完成"
    
    log_success "应用服务部署完成"
}

# 部署网络配置
deploy_networking() {
    log_info "部署网络配置..."
    
    # 检查是否需要部署Ingress
    if kubectl get ingressclass nginx &> /dev/null; then
        kubectl apply -f deploy/k8s/ingress.yaml
        log_success "Ingress配置部署完成"
    else
        log_warning "未检测到Nginx Ingress Controller，跳过Ingress部署"
    fi
}

# 等待应用就绪
wait_for_application() {
    log_info "等待应用服务就绪..."
    
    # 等待Core服务就绪
    log_info "等待AlertAgent Core服务就绪..."
    kubectl wait --for=condition=ready pod -l app.kubernetes.io/component=core -n alertagent --timeout=300s
    
    # 等待Worker服务就绪
    log_info "等待Worker服务就绪..."
    kubectl wait --for=condition=ready pod -l app.kubernetes.io/component=worker -n alertagent --timeout=300s
    
    log_success "应用服务已就绪"
}

# 健康检查
health_check() {
    log_info "执行健康检查..."
    
    # 获取服务端点
    local service_ip=$(kubectl get service alertagent-core -n alertagent -o jsonpath='{.spec.clusterIP}')
    
    # 端口转发进行健康检查
    kubectl port-forward service/alertagent-core 8080:8080 -n alertagent &
    local port_forward_pid=$!
    
    sleep 5
    
    # 执行健康检查
    if curl -f -s http://localhost:8080/api/v1/health > /dev/null; then
        log_success "健康检查通过"
    else
        log_error "健康检查失败"
        kill $port_forward_pid 2>/dev/null || true
        return 1
    fi
    
    kill $port_forward_pid 2>/dev/null || true
    log_success "健康检查完成"
}

# 显示部署状态
show_status() {
    log_info "部署状态概览:"
    echo
    
    # 显示Pod状态
    echo "Pod状态:"
    kubectl get pods -n alertagent -o wide
    echo
    
    # 显示Service状态
    echo "Service状态:"
    kubectl get services -n alertagent
    echo
    
    # 显示Ingress状态
    if kubectl get ingress -n alertagent &> /dev/null; then
        echo "Ingress状态:"
        kubectl get ingress -n alertagent
        echo
    fi
    
    # 显示HPA状态
    echo "HPA状态:"
    kubectl get hpa -n alertagent
    echo
    
    # 显示访问信息
    log_info "访问信息:"
    local external_ip=$(kubectl get service alertagent-lb -n alertagent -o jsonpath='{.status.loadBalancer.ingress[0].ip}' 2>/dev/null || echo "pending")
    
    if [ "$external_ip" != "pending" ] && [ -n "$external_ip" ]; then
        echo "  - 外部访问: http://$external_ip"
    else
        echo "  - 端口转发访问: kubectl port-forward service/alertagent-core 8080:8080 -n alertagent"
        echo "  - 然后访问: http://localhost:8080"
    fi
    
    echo "  - 集群内访问: http://alertagent-core.alertagent.svc.cluster.local:8080"
}

# 清理部署
cleanup() {
    log_info "清理部署资源..."
    
    # 删除应用资源
    kubectl delete -f deploy/k8s/alertagent-worker.yaml --ignore-not-found=true
    kubectl delete -f deploy/k8s/alertagent-core.yaml --ignore-not-found=true
    kubectl delete -f deploy/k8s/ingress.yaml --ignore-not-found=true
    
    # 删除基础设施
    kubectl delete -f deploy/k8s/redis.yaml --ignore-not-found=true
    kubectl delete -f deploy/k8s/mysql.yaml --ignore-not-found=true
    kubectl delete -f deploy/k8s/configmap.yaml --ignore-not-found=true
    kubectl delete -f deploy/k8s/secrets.yaml --ignore-not-found=true
    
    # 删除命名空间
    kubectl delete -f deploy/k8s/namespace.yaml --ignore-not-found=true
    
    log_success "清理完成"
}

# 更新部署
update_deployment() {
    local component=$1
    
    log_info "更新部署: $component"
    
    case $component in
        "core")
            kubectl rollout restart deployment/alertagent-core -n alertagent
            kubectl rollout status deployment/alertagent-core -n alertagent
            ;;
        "worker")
            kubectl rollout restart deployment/alertagent-worker-ai -n alertagent
            kubectl rollout restart deployment/alertagent-worker-notification -n alertagent
            kubectl rollout restart deployment/alertagent-worker-config -n alertagent
            kubectl rollout status deployment/alertagent-worker-ai -n alertagent
            kubectl rollout status deployment/alertagent-worker-notification -n alertagent
            kubectl rollout status deployment/alertagent-worker-config -n alertagent
            ;;
        "all")
            kubectl rollout restart deployment -n alertagent
            kubectl rollout status deployment/alertagent-core -n alertagent
            kubectl rollout status deployment/alertagent-worker-ai -n alertagent
            kubectl rollout status deployment/alertagent-worker-notification -n alertagent
            kubectl rollout status deployment/alertagent-worker-config -n alertagent
            ;;
        *)
            log_error "不支持的组件: $component"
            exit 1
            ;;
    esac
    
    log_success "更新完成: $component"
}

# 显示日志
show_logs() {
    local component=$1
    local follow=${2:-false}
    
    local follow_flag=""
    if [ "$follow" = "true" ]; then
        follow_flag="-f"
    fi
    
    case $component in
        "core")
            kubectl logs -l app.kubernetes.io/component=core -n alertagent $follow_flag
            ;;
        "worker")
            kubectl logs -l app.kubernetes.io/component=worker -n alertagent $follow_flag
            ;;
        "mysql")
            kubectl logs -l app.kubernetes.io/component=mysql -n alertagent $follow_flag
            ;;
        "redis")
            kubectl logs -l app.kubernetes.io/component=redis -n alertagent $follow_flag
            ;;
        "all")
            kubectl logs -l app.kubernetes.io/name=alertagent -n alertagent $follow_flag
            ;;
        *)
            log_error "不支持的组件: $component"
            exit 1
            ;;
    esac
}

# 显示帮助信息
show_help() {
    echo "AlertAgent Kubernetes部署脚本"
    echo
    echo "用法: $0 <action> [options]"
    echo
    echo "操作:"
    echo "  deploy      完整部署"
    echo "  cleanup     清理部署"
    echo "  status      显示状态"
    echo "  health      健康检查"
    echo "  update      更新部署 [core|worker|all]"
    echo "  logs        显示日志 [core|worker|mysql|redis|all] [follow]"
    echo "  restart     重启服务 [core|worker|all]"
    echo
    echo "示例:"
    echo "  $0 deploy           # 完整部署"
    echo "  $0 status           # 显示状态"
    echo "  $0 update core      # 更新Core服务"
    echo "  $0 logs core true   # 跟踪Core日志"
    echo "  $0 cleanup          # 清理部署"
    echo
    echo "选项:"
    echo "  -h, --help    显示此帮助信息"
}

# 主函数
main() {
    local action=$1
    local param1=$2
    local param2=$3
    
    if [ -z "$action" ]; then
        show_help
        exit 1
    fi
    
    # 检查依赖
    check_dependencies
    
    case $action in
        "deploy")
            log_info "开始完整部署..."
            create_namespace
            create_secrets
            deploy_infrastructure
            wait_for_services
            deploy_application
            deploy_networking
            wait_for_application
            health_check
            show_status
            log_success "部署完成！"
            ;;
        "cleanup")
            cleanup
            ;;
        "status")
            show_status
            ;;
        "health")
            health_check
            ;;
        "update")
            update_deployment "${param1:-all}"
            ;;
        "logs")
            show_logs "${param1:-all}" "${param2:-false}"
            ;;
        "restart")
            update_deployment "${param1:-all}"
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