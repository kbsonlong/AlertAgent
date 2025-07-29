#!/bin/bash

# AlertAgent Helm Chart 部署脚本
# 使用方法: ./deploy.sh [环境] [操作]
# 环境: dev, staging, prod
# 操作: install, upgrade, uninstall, test

set -e

# 默认配置
DEFAULT_ENVIRONMENT="dev"
DEFAULT_ACTION="install"
CHART_NAME="alertagent"
CHART_PATH="./alertagent"
NAMESPACE="alertagent"

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

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

# 显示帮助信息
show_help() {
    cat << EOF
AlertAgent Helm Chart 部署脚本

使用方法:
    $0 [环境] [操作] [选项]

环境:
    dev      - 开发环境 (默认)
    staging  - 测试环境
    prod     - 生产环境

操作:
    install    - 安装 (默认)
    upgrade    - 升级
    uninstall  - 卸载
    test       - 运行测试
    status     - 查看状态
    logs       - 查看日志

选项:
    --dry-run           - 模拟运行，不实际执行
    --debug             - 启用调试模式
    --wait              - 等待部署完成
    --timeout DURATION - 设置超时时间 (默认: 10m)
    --values FILE       - 指定自定义 values 文件
    --set KEY=VALUE     - 设置特定的值
    --namespace NS      - 指定命名空间 (默认: alertagent)
    --create-namespace  - 如果命名空间不存在则创建
    --help              - 显示此帮助信息

示例:
    $0 dev install --wait --create-namespace
    $0 prod upgrade --values values-prod.yaml
    $0 staging uninstall
    $0 dev test
    $0 prod status

EOF
}

# 检查依赖
check_dependencies() {
    log_info "检查依赖..."
    
    # 检查 helm
    if ! command -v helm &> /dev/null; then
        log_error "Helm 未安装，请先安装 Helm"
        exit 1
    fi
    
    # 检查 kubectl
    if ! command -v kubectl &> /dev/null; then
        log_error "kubectl 未安装，请先安装 kubectl"
        exit 1
    fi
    
    # 检查 Kubernetes 连接
    if ! kubectl cluster-info &> /dev/null; then
        log_error "无法连接到 Kubernetes 集群"
        exit 1
    fi
    
    log_success "依赖检查通过"
}

# 添加 Helm 仓库
add_helm_repos() {
    log_info "添加 Helm 仓库..."
    
    helm repo add bitnami https://charts.bitnami.com/bitnami
    helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
    helm repo add grafana https://grafana.github.io/helm-charts
    helm repo update
    
    log_success "Helm 仓库添加完成"
}

# 创建命名空间
create_namespace() {
    if [[ "$CREATE_NAMESPACE" == "true" ]]; then
        log_info "创建命名空间: $NAMESPACE"
        kubectl create namespace "$NAMESPACE" --dry-run=client -o yaml | kubectl apply -f -
        log_success "命名空间创建完成"
    fi
}

# 获取 values 文件
get_values_file() {
    local env=$1
    local values_file="values-${env}.yaml"
    
    if [[ -n "$CUSTOM_VALUES" ]]; then
        echo "$CUSTOM_VALUES"
    elif [[ -f "$values_file" ]]; then
        echo "$values_file"
    else
        echo "values.yaml"
    fi
}

# 构建 Helm 命令参数
build_helm_args() {
    local args=()
    
    # 基本参数
    args+=("--namespace" "$NAMESPACE")
    
    # Values 文件
    local values_file
    values_file=$(get_values_file "$ENVIRONMENT")
    if [[ -f "$CHART_PATH/$values_file" ]]; then
        args+=("--values" "$CHART_PATH/$values_file")
    fi
    
    # 其他选项
    [[ "$DRY_RUN" == "true" ]] && args+=("--dry-run")
    [[ "$DEBUG" == "true" ]] && args+=("--debug")
    [[ "$WAIT" == "true" ]] && args+=("--wait")
    [[ -n "$TIMEOUT" ]] && args+=("--timeout" "$TIMEOUT")
    [[ "$CREATE_NAMESPACE" == "true" ]] && args+=("--create-namespace")
    
    # 自定义设置
    for set_value in "${SET_VALUES[@]}"; do
        args+=("--set" "$set_value")
    done
    
    echo "${args[@]}"
}

# 安装
install_chart() {
    log_info "安装 AlertAgent ($ENVIRONMENT 环境)..."
    
    local helm_args
    helm_args=$(build_helm_args)
    
    # shellcheck disable=SC2086
    helm install "$CHART_NAME-$ENVIRONMENT" "$CHART_PATH" $helm_args
    
    log_success "AlertAgent 安装完成"
    
    if [[ "$DRY_RUN" != "true" ]]; then
        show_status
    fi
}

# 升级
upgrade_chart() {
    log_info "升级 AlertAgent ($ENVIRONMENT 环境)..."
    
    local helm_args
    helm_args=$(build_helm_args)
    
    # shellcheck disable=SC2086
    helm upgrade "$CHART_NAME-$ENVIRONMENT" "$CHART_PATH" $helm_args
    
    log_success "AlertAgent 升级完成"
    
    if [[ "$DRY_RUN" != "true" ]]; then
        show_status
    fi
}

# 卸载
uninstall_chart() {
    log_warning "卸载 AlertAgent ($ENVIRONMENT 环境)..."
    
    read -p "确认要卸载吗? (y/N): " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        helm uninstall "$CHART_NAME-$ENVIRONMENT" --namespace "$NAMESPACE"
        log_success "AlertAgent 卸载完成"
    else
        log_info "取消卸载"
    fi
}

# 运行测试
run_tests() {
    log_info "运行 AlertAgent 测试..."
    
    helm test "$CHART_NAME-$ENVIRONMENT" --namespace "$NAMESPACE"
    
    log_success "测试完成"
}

# 显示状态
show_status() {
    log_info "AlertAgent 状态:"
    
    echo
    echo "=== Helm Release ==="
    helm status "$CHART_NAME-$ENVIRONMENT" --namespace "$NAMESPACE"
    
    echo
    echo "=== Pods ==="
    kubectl get pods -n "$NAMESPACE" -l app.kubernetes.io/instance="$CHART_NAME-$ENVIRONMENT"
    
    echo
    echo "=== Services ==="
    kubectl get services -n "$NAMESPACE" -l app.kubernetes.io/instance="$CHART_NAME-$ENVIRONMENT"
    
    echo
    echo "=== Ingress ==="
    kubectl get ingress -n "$NAMESPACE" -l app.kubernetes.io/instance="$CHART_NAME-$ENVIRONMENT"
    
    echo
    echo "=== HPA ==="
    kubectl get hpa -n "$NAMESPACE" -l app.kubernetes.io/instance="$CHART_NAME-$ENVIRONMENT"
}

# 显示日志
show_logs() {
    log_info "AlertAgent 日志:"
    
    echo "选择要查看的服务:"
    echo "1) API"
    echo "2) Worker"
    echo "3) Rule Server"
    echo "4) All"
    
    read -p "请选择 (1-4): " -n 1 -r
    echo
    
    case $REPLY in
        1)
            kubectl logs -n "$NAMESPACE" -l app.kubernetes.io/component=api -f --tail=100
            ;;
        2)
            kubectl logs -n "$NAMESPACE" -l app.kubernetes.io/component=worker -f --tail=100
            ;;
        3)
            kubectl logs -n "$NAMESPACE" -l app.kubernetes.io/component=rule-server -f --tail=100
            ;;
        4)
            kubectl logs -n "$NAMESPACE" -l app.kubernetes.io/instance="$CHART_NAME-$ENVIRONMENT" -f --tail=100
            ;;
        *)
            log_error "无效选择"
            exit 1
            ;;
    esac
}

# 解析命令行参数
parse_args() {
    ENVIRONMENT="$DEFAULT_ENVIRONMENT"
    ACTION="$DEFAULT_ACTION"
    DRY_RUN="false"
    DEBUG="false"
    WAIT="false"
    TIMEOUT="10m"
    CREATE_NAMESPACE="false"
    CUSTOM_VALUES=""
    SET_VALUES=()
    
    while [[ $# -gt 0 ]]; do
        case $1 in
            dev|staging|prod)
                ENVIRONMENT="$1"
                shift
                ;;
            install|upgrade|uninstall|test|status|logs)
                ACTION="$1"
                shift
                ;;
            --dry-run)
                DRY_RUN="true"
                shift
                ;;
            --debug)
                DEBUG="true"
                shift
                ;;
            --wait)
                WAIT="true"
                shift
                ;;
            --timeout)
                TIMEOUT="$2"
                shift 2
                ;;
            --values)
                CUSTOM_VALUES="$2"
                shift 2
                ;;
            --set)
                SET_VALUES+=("$2")
                shift 2
                ;;
            --namespace)
                NAMESPACE="$2"
                shift 2
                ;;
            --create-namespace)
                CREATE_NAMESPACE="true"
                shift
                ;;
            --help|-h)
                show_help
                exit 0
                ;;
            *)
                log_error "未知参数: $1"
                show_help
                exit 1
                ;;
        esac
    done
}

# 主函数
main() {
    parse_args "$@"
    
    log_info "AlertAgent Helm 部署脚本"
    log_info "环境: $ENVIRONMENT"
    log_info "操作: $ACTION"
    log_info "命名空间: $NAMESPACE"
    
    check_dependencies
    
    case $ACTION in
        install)
            add_helm_repos
            create_namespace
            install_chart
            ;;
        upgrade)
            add_helm_repos
            upgrade_chart
            ;;
        uninstall)
            uninstall_chart
            ;;
        test)
            run_tests
            ;;
        status)
            show_status
            ;;
        logs)
            show_logs
            ;;
        *)
            log_error "未知操作: $ACTION"
            show_help
            exit 1
            ;;
    esac
}

# 执行主函数
main "$@"