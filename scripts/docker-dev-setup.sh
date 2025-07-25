#!/bin/bash

# AlertAgent Docker 开发环境启动脚本
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

# 检查 Docker 和 Docker Compose
check_docker() {
    log_info "检查 Docker 环境..."
    
    if ! command -v docker &> /dev/null; then
        log_error "Docker 未安装，请先安装 Docker"
        log_info "安装地址: https://docs.docker.com/get-docker/"
        exit 1
    fi
    
    if ! command -v docker-compose &> /dev/null && ! docker compose version &> /dev/null; then
        log_error "Docker Compose 未安装，请先安装 Docker Compose"
        exit 1
    fi
    
    # 检查 Docker 是否运行
    if ! docker info &> /dev/null; then
        log_error "Docker 未运行，请启动 Docker"
        exit 1
    fi
    
    log_success "Docker 环境检查完成"
}

# 启动 Docker 服务
start_docker_services() {
    log_info "启动 Docker 服务..."
    
    # 使用新版本的 docker compose 或旧版本的 docker-compose
    if docker compose version &> /dev/null; then
        COMPOSE_CMD="docker compose"
    else
        COMPOSE_CMD="docker-compose"
    fi
    
    # 启动服务
    $COMPOSE_CMD -f docker-compose.dev.yml up -d
    
    log_success "Docker 服务启动完成"
}

# 等待服务就绪
wait_for_services() {
    log_info "等待服务就绪..."
    
    # 等待 MySQL
    log_info "等待 MySQL 启动..."
    timeout=60
    while [ $timeout -gt 0 ]; do
        if docker exec alertagent-mysql mysqladmin ping -h localhost --silent 2>/dev/null; then
            log_success "MySQL 已就绪"
            break
        fi
        echo -n "."
        sleep 2
        timeout=$((timeout - 2))
    done
    
    if [ $timeout -le 0 ]; then
        log_error "MySQL 启动超时"
        exit 1
    fi
    
    # 等待 Redis
    log_info "等待 Redis 启动..."
    timeout=30
    while [ $timeout -gt 0 ]; do
        if docker exec alertagent-redis redis-cli ping 2>/dev/null | grep -q PONG; then
            log_success "Redis 已就绪"
            break
        fi
        echo -n "."
        sleep 1
        timeout=$((timeout - 1))
    done
    
    if [ $timeout -le 0 ]; then
        log_error "Redis 启动超时"
        exit 1
    fi
    
    # 等待 Ollama
    log_info "等待 Ollama 启动..."
    timeout=60
    while [ $timeout -gt 0 ]; do
        if curl -s http://localhost:11434/api/tags >/dev/null 2>&1; then
            log_success "Ollama 已就绪"
            break
        fi
        echo -n "."
        sleep 2
        timeout=$((timeout - 2))
    done
    
    if [ $timeout -le 0 ]; then
        log_warning "Ollama 启动超时，但可以继续"
    fi
}

# 安装 Ollama 模型
setup_ollama_model() {
    log_info "设置 Ollama 模型..."
    
    # 检查模型是否已安装
    if docker exec alertagent-ollama ollama list | grep -q "deepseek-r1:32b"; then
        log_success "模型 deepseek-r1:32b 已安装"
        return 0
    fi
    
    log_info "正在下载模型 deepseek-r1:32b (这可能需要一些时间)..."
    if docker exec alertagent-ollama ollama pull deepseek-r1:32b; then
        log_success "模型下载完成"
    else
        log_warning "模型下载失败，请手动执行: docker exec alertagent-ollama ollama pull deepseek-r1:32b"
    fi
}

# 安装依赖
install_dependencies() {
    log_info "安装项目依赖..."
    
    # 安装 Go 依赖
    if go mod download && go mod tidy; then
        log_success "Go 依赖安装完成"
    else
        log_error "Go 依赖安装失败"
        exit 1
    fi
    
    # 安装前端依赖
    cd web
    if npm install; then
        log_success "前端依赖安装完成"
    else
        log_error "前端依赖安装失败"
        exit 1
    fi
    cd ..
}

# 启动应用服务
start_app_services() {
    log_info "启动应用服务..."
    
    # 启动后端
    log_info "启动后端服务..."
    go run cmd/main.go &
    BACKEND_PID=$!
    echo $BACKEND_PID > .backend.pid
    
    # 等待后端启动
    timeout=30
    while [ $timeout -gt 0 ]; do
        if curl -s http://localhost:8080/health >/dev/null 2>&1; then
            log_success "后端服务已启动 (PID: $BACKEND_PID)"
            break
        fi
        echo -n "."
        sleep 1
        timeout=$((timeout - 1))
    done
    
    if [ $timeout -le 0 ]; then
        log_error "后端服务启动超时"
        exit 1
    fi
    
    # 启动前端
    log_info "启动前端服务..."
    cd web
    npm run dev &
    FRONTEND_PID=$!
    cd ..
    echo $FRONTEND_PID > .frontend.pid
    
    # 等待前端启动
    timeout=30
    while [ $timeout -gt 0 ]; do
        if curl -s http://localhost:5173 >/dev/null 2>&1; then
            log_success "前端服务已启动 (PID: $FRONTEND_PID)"
            break
        fi
        echo -n "."
        sleep 1
        timeout=$((timeout - 1))
    done
    
    if [ $timeout -le 0 ]; then
        log_error "前端服务启动超时"
        exit 1
    fi
}

# 显示服务信息
show_services() {
    echo
    log_success "=== Docker 开发环境启动完成 ==="
    echo
    echo "📊 应用服务:"
    echo "   前端: http://localhost:5173"
    echo "   后端: http://localhost:8080"
    echo "   API文档: http://localhost:8080/swagger/index.html"
    echo
    echo "🗄️  数据库服务:"
    echo "   MySQL: localhost:3306 (数据库: alert_agent)"
    echo "   Redis: localhost:6379"
    echo "   Ollama: http://localhost:11434"
    echo
    echo "🔧 管理工具:"
    echo "   phpMyAdmin: http://localhost:8081 (用户名: root, 密码: along123)"
    echo "   Redis Commander: http://localhost:8082"
    echo
    echo "🐳 Docker 管理:"
    echo "   查看容器状态: docker-compose -f docker-compose.dev.yml ps"
    echo "   查看日志: docker-compose -f docker-compose.dev.yml logs -f"
    echo "   停止服务: ./scripts/docker-dev-stop.sh"
    echo "   重启服务: ./scripts/docker-dev-restart.sh"
    echo
    log_info "按 Ctrl+C 停止应用服务 (Docker 服务将继续运行)"
}

# 清理函数
cleanup() {
    log_info "正在停止应用服务..."
    
    # 停止后端
    if [ -f .backend.pid ]; then
        BACKEND_PID=$(cat .backend.pid)
        if kill -0 $BACKEND_PID 2>/dev/null; then
            kill $BACKEND_PID
            log_success "后端服务已停止"
        fi
        rm -f .backend.pid
    fi
    
    # 停止前端
    if [ -f .frontend.pid ]; then
        FRONTEND_PID=$(cat .frontend.pid)
        if kill -0 $FRONTEND_PID 2>/dev/null; then
            kill $FRONTEND_PID
            log_success "前端服务已停止"
        fi
        rm -f .frontend.pid
    fi
    
    log_info "Docker 服务仍在运行，如需停止请执行: ./scripts/docker-dev-stop.sh"
    exit 0
}

# 设置信号处理
trap cleanup SIGINT SIGTERM

# 主函数
main() {
    echo "🐳 AlertAgent Docker 开发环境启动脚本"
    echo "========================================"
    echo
    
    # 检查 Docker
    check_docker
    
    # 启动 Docker 服务
    start_docker_services
    
    # 等待服务就绪
    wait_for_services
    
    # 设置 Ollama 模型
    setup_ollama_model
    
    # 安装依赖
    install_dependencies
    
    # 启动应用服务
    start_app_services
    
    # 显示服务信息
    show_services
    
    # 保持脚本运行
    wait
}

# 执行主函数
main "$@"