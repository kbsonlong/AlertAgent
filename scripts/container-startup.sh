#!/bin/bash

# AlertAgent 容器启动脚本
# 用于容器内部的启动前检查和初始化

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# 日志函数
log_info() {
    echo -e "${BLUE}[$(date '+%Y-%m-%d %H:%M:%S')] [INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[$(date '+%Y-%m-%d %H:%M:%S')] [SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[$(date '+%Y-%m-%d %H:%M:%S')] [WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[$(date '+%Y-%m-%d %H:%M:%S')] [ERROR]${NC} $1"
}

# 等待服务可用
wait_for_service() {
    local host=$1
    local port=$2
    local service_name=$3
    local max_attempts=${4:-30}
    local attempt=1
    
    log_info "等待 $service_name 服务可用 ($host:$port)..."
    
    while [ $attempt -le $max_attempts ]; do
        if nc -z "$host" "$port" 2>/dev/null; then
            log_success "$service_name 服务已可用"
            return 0
        else
            log_info "等待 $service_name 服务... ($attempt/$max_attempts)"
            sleep 2
            ((attempt++))
        fi
    done
    
    log_error "$service_name 服务等待超时"
    return 1
}

# 检查环境变量
check_env_vars() {
    local required_vars=("$@")
    local missing_vars=()
    
    for var in "${required_vars[@]}"; do
        if [ -z "${!var}" ]; then
            missing_vars+=("$var")
        fi
    done
    
    if [ ${#missing_vars[@]} -ne 0 ]; then
        log_error "缺少必需的环境变量:"
        for var in "${missing_vars[@]}"; do
            echo "  - $var"
        done
        exit 1
    fi
    
    log_success "环境变量检查通过"
}

# 创建必要的目录
create_directories() {
    local dirs=("logs" "tmp" "data")
    
    for dir in "${dirs[@]}"; do
        if [ ! -d "$dir" ]; then
            mkdir -p "$dir"
            log_info "创建目录: $dir"
        fi
    done
    
    log_success "目录检查完成"
}

# 检查配置文件
check_config_file() {
    local config_file=${CONFIG_PATH:-"/app/config/config.yaml"}
    
    if [ ! -f "$config_file" ]; then
        log_error "配置文件不存在: $config_file"
        exit 1
    fi
    
    log_success "配置文件检查通过: $config_file"
}

# 数据库连接检查
check_database() {
    if [ -n "$DB_HOST" ] && [ -n "$DB_PORT" ]; then
        wait_for_service "$DB_HOST" "$DB_PORT" "MySQL数据库" 60
    else
        log_warning "未配置数据库连接参数，跳过数据库检查"
    fi
}

# Redis连接检查
check_redis() {
    if [ -n "$REDIS_HOST" ] && [ -n "$REDIS_PORT" ]; then
        wait_for_service "$REDIS_HOST" "$REDIS_PORT" "Redis缓存" 30
    else
        log_warning "未配置Redis连接参数，跳过Redis检查"
    fi
}

# 启动前检查
pre_start_checks() {
    log_info "开始启动前检查..."
    
    # 创建必要目录
    create_directories
    
    # 检查配置文件
    check_config_file
    
    # 检查依赖服务
    check_database
    check_redis
    
    log_success "启动前检查完成"
}

# 启动AlertAgent Core
start_core() {
    log_info "启动AlertAgent Core服务..."
    
    # 检查必需的环境变量
    check_env_vars "DB_HOST" "DB_USER" "DB_PASSWORD" "DB_NAME" "REDIS_HOST"
    
    # 启动前检查
    pre_start_checks
    
    # 启动服务
    log_info "执行: ./alertagent-core"
    exec ./alertagent-core
}

# 启动Worker
start_worker() {
    log_info "启动AlertAgent Worker服务..."
    
    # 检查必需的环境变量
    check_env_vars "DB_HOST" "DB_USER" "DB_PASSWORD" "DB_NAME" "REDIS_HOST"
    
    # 启动前检查
    pre_start_checks
    
    # 构建启动命令
    local cmd="./alertagent-worker"
    
    # 添加命令行参数
    if [ -n "$WORKER_NAME" ]; then
        cmd="$cmd -name $WORKER_NAME"
    fi
    
    if [ -n "$WORKER_TYPE" ]; then
        cmd="$cmd -type $WORKER_TYPE"
    fi
    
    if [ -n "$WORKER_CONCURRENCY" ]; then
        cmd="$cmd -concurrency $WORKER_CONCURRENCY"
    fi
    
    if [ -n "$WORKER_QUEUES" ]; then
        cmd="$cmd -queues $WORKER_QUEUES"
    fi
    
    if [ -n "$HEALTH_PORT" ]; then
        cmd="$cmd -health-port $HEALTH_PORT"
    fi
    
    # 启动服务
    log_info "执行: $cmd"
    exec $cmd
}

# 启动Sidecar
start_sidecar() {
    log_info "启动AlertAgent Sidecar服务..."
    
    # 检查必需的环境变量
    check_env_vars "ALERTAGENT_ENDPOINT" "CLUSTER_ID" "CONFIG_TYPE" "CONFIG_PATH" "RELOAD_URL"
    
    # 构建启动命令
    local cmd="./alertagent-sidecar"
    cmd="$cmd -endpoint $ALERTAGENT_ENDPOINT"
    cmd="$cmd -cluster-id $CLUSTER_ID"
    cmd="$cmd -type $CONFIG_TYPE"
    cmd="$cmd -config-path $CONFIG_PATH"
    cmd="$cmd -reload-url $RELOAD_URL"
    
    if [ -n "$SYNC_INTERVAL" ]; then
        cmd="$cmd -sync-interval $SYNC_INTERVAL"
    fi
    
    if [ -n "$HEALTH_PORT" ]; then
        cmd="$cmd -health-port $HEALTH_PORT"
    fi
    
    if [ -n "$LOG_LEVEL" ]; then
        cmd="$cmd -log-level $LOG_LEVEL"
    fi
    
    # 启动服务
    log_info "执行: $cmd"
    exec $cmd
}

# 主函数
main() {
    local service_type=${1:-"core"}
    
    log_info "AlertAgent容器启动脚本"
    log_info "服务类型: $service_type"
    log_info "当前用户: $(whoami)"
    log_info "工作目录: $(pwd)"
    
    case $service_type in
        "core")
            start_core
            ;;
        "worker")
            start_worker
            ;;
        "sidecar")
            start_sidecar
            ;;
        *)
            log_error "不支持的服务类型: $service_type"
            echo "支持的服务类型: core, worker, sidecar"
            exit 1
            ;;
    esac
}

# 信号处理
trap 'log_info "收到终止信号，正在关闭..."; exit 0' SIGTERM SIGINT

# 执行主函数
main "$@"