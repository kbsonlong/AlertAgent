#!/bin/bash

# AlertAgent Docker 开发环境停止脚本
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

# 停止应用服务
stop_app_services() {
    log_info "停止应用服务..."
    
    # 停止后端
    if [ -f .backend.pid ]; then
        BACKEND_PID=$(cat .backend.pid)
        if kill -0 $BACKEND_PID 2>/dev/null; then
            kill $BACKEND_PID
            log_success "后端服务已停止 (PID: $BACKEND_PID)"
        else
            log_warning "后端服务进程不存在"
        fi
        rm -f .backend.pid
    else
        log_warning "后端服务 PID 文件不存在"
    fi
    
    # 停止前端
    if [ -f .frontend.pid ]; then
        FRONTEND_PID=$(cat .frontend.pid)
        if kill -0 $FRONTEND_PID 2>/dev/null; then
            kill $FRONTEND_PID
            log_success "前端服务已停止 (PID: $FRONTEND_PID)"
        else
            log_warning "前端服务进程不存在"
        fi
        rm -f .frontend.pid
    else
        log_warning "前端服务 PID 文件不存在"
    fi
    
    # 停止端口上的进程
    local backend_pid=$(lsof -ti:8080 2>/dev/null || true)
    if [ -n "$backend_pid" ]; then
        kill $backend_pid 2>/dev/null || true
        log_success "停止了端口 8080 上的进程 (PID: $backend_pid)"
    fi
    
    local frontend_pid=$(lsof -ti:5173 2>/dev/null || true)
    if [ -n "$frontend_pid" ]; then
        kill $frontend_pid 2>/dev/null || true
        log_success "停止了端口 5173 上的进程 (PID: $frontend_pid)"
    fi
}

# 停止 Docker 服务
stop_docker_services() {
    log_info "停止 Docker 服务..."
    
    # 检查 docker-compose.dev.yml 是否存在
    if [ ! -f "docker-compose.dev.yml" ]; then
        log_error "找不到 docker-compose.dev.yml 文件"
        exit 1
    fi
    
    # 使用新版本的 docker compose 或旧版本的 docker-compose
    if docker compose version &> /dev/null; then
        COMPOSE_CMD="docker compose"
    else
        COMPOSE_CMD="docker-compose"
    fi
    
    # 停止并移除容器
    $COMPOSE_CMD -f docker-compose.dev.yml down
    
    log_success "Docker 服务已停止"
}

# 清理 Docker 资源（可选）
cleanup_docker_resources() {
    local cleanup_volumes=$1
    
    if [ "$cleanup_volumes" = "--cleanup" ] || [ "$cleanup_volumes" = "-c" ]; then
        log_warning "清理 Docker 资源（包括数据卷）..."
        
        # 使用新版本的 docker compose 或旧版本的 docker-compose
        if docker compose version &> /dev/null; then
            COMPOSE_CMD="docker compose"
        else
            COMPOSE_CMD="docker-compose"
        fi
        
        # 停止并移除容器、网络、数据卷
        $COMPOSE_CMD -f docker-compose.dev.yml down -v --remove-orphans
        
        # 清理未使用的镜像
        docker image prune -f
        
        log_success "Docker 资源清理完成"
        log_warning "注意：所有数据已被删除，下次启动将重新初始化"
    fi
}

# 显示帮助信息
show_help() {
    echo "用法: $0 [选项]"
    echo
    echo "选项:"
    echo "  -c, --cleanup    停止服务并清理所有 Docker 资源（包括数据卷）"
    echo "  -h, --help       显示此帮助信息"
    echo
    echo "示例:"
    echo "  $0               # 仅停止服务"
    echo "  $0 --cleanup     # 停止服务并清理所有数据"
}

# 主函数
main() {
    local cleanup_flag=""
    
    # 解析命令行参数
    while [[ $# -gt 0 ]]; do
        case $1 in
            -c|--cleanup)
                cleanup_flag="--cleanup"
                shift
                ;;
            -h|--help)
                show_help
                exit 0
                ;;
            *)
                log_error "未知选项: $1"
                show_help
                exit 1
                ;;
        esac
    done
    
    echo "🐳 AlertAgent Docker 开发环境停止脚本"
    echo "========================================"
    echo
    
    # 停止应用服务
    stop_app_services
    
    # 停止 Docker 服务
    stop_docker_services
    
    # 可选：清理 Docker 资源
    cleanup_docker_resources "$cleanup_flag"
    
    echo
    log_success "开发环境已停止"
    echo
    
    if [ "$cleanup_flag" != "--cleanup" ]; then
        log_info "如需重新启动，请运行: ./scripts/docker-dev-setup.sh"
        log_info "如需清理所有数据，请运行: $0 --cleanup"
    else
        log_info "如需重新启动，请运行: ./scripts/docker-dev-setup.sh"
    fi
}

# 执行主函数
main "$@"