#!/bin/bash

# AlertAgent 开发环境停止脚本
# 作者: AlertAgent Team
# 版本: 1.0.0

set -e

# 颜色定义
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

# 停止进程
stop_process() {
    local pid_file=$1
    local service_name=$2
    
    if [ -f $pid_file ]; then
        local pid=$(cat $pid_file)
        if kill -0 $pid 2>/dev/null; then
            kill $pid
            log_success "$service_name 已停止 (PID: $pid)"
        else
            log_warning "$service_name 进程不存在"
        fi
        rm -f $pid_file
    else
        log_warning "$service_name PID 文件不存在"
    fi
}

# 停止端口上的进程
stop_port_process() {
    local port=$1
    local service_name=$2
    
    local pid=$(lsof -ti:$port 2>/dev/null || true)
    if [ -n "$pid" ]; then
        kill $pid 2>/dev/null || true
        log_success "停止了端口 $port 上的 $service_name 进程 (PID: $pid)"
    else
        log_info "端口 $port 上没有运行的进程"
    fi
}

# 主函数
main() {
    echo "🛑 AlertAgent 开发环境停止脚本"
    echo "================================="
    echo
    
    log_info "正在停止开发环境..."
    
    # 停止后端服务
    log_info "停止后端服务..."
    stop_process ".backend.pid" "后端服务"
    stop_port_process "8080" "后端服务"
    
    # 停止前端服务
    log_info "停止前端服务..."
    stop_process ".frontend.pid" "前端服务"
    stop_port_process "5173" "前端服务"
    
    # 可选：停止数据库服务（注释掉，因为可能影响其他项目）
    # log_info "停止数据库服务..."
    # if [[ "$OSTYPE" == "darwin"* ]]; then
    #     if command -v brew &> /dev/null; then
    #         brew services stop mysql
    #         brew services stop redis
    #     fi
    # else
    #     sudo systemctl stop mysql
    #     sudo systemctl stop redis
    # fi
    
    echo
    log_success "开发环境已停止"
    echo
    log_info "如需重新启动，请运行: ./scripts/dev-setup.sh"
}

# 执行主函数
main "$@"