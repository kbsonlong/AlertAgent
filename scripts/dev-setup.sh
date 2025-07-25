#!/bin/bash

# AlertAgent 开发环境一键启动脚本
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

# 检查命令是否存在
check_command() {
    if ! command -v $1 &> /dev/null; then
        log_error "$1 未安装，请先安装 $1"
        return 1
    fi
    return 0
}

# 检查端口是否被占用
check_port() {
    local port=$1
    if lsof -Pi :$port -sTCP:LISTEN -t >/dev/null 2>&1; then
        log_warning "端口 $port 已被占用"
        return 1
    fi
    return 0
}

# 等待服务启动
wait_for_service() {
    local host=$1
    local port=$2
    local service_name=$3
    local max_attempts=30
    local attempt=1
    
    log_info "等待 $service_name 启动..."
    
    while [ $attempt -le $max_attempts ]; do
        if nc -z $host $port 2>/dev/null; then
            log_success "$service_name 已启动"
            return 0
        fi
        
        echo -n "."
        sleep 1
        attempt=$((attempt + 1))
    done
    
    echo
    log_error "$service_name 启动超时"
    return 1
}

# 检查必要的依赖
check_dependencies() {
    log_info "检查系统依赖..."
    
    # 检查 Go
    if ! check_command "go"; then
        log_error "请安装 Go 1.21+: https://golang.org/dl/"
        exit 1
    fi
    
    # 检查 Node.js
    if ! check_command "node"; then
        log_error "请安装 Node.js 18+: https://nodejs.org/"
        exit 1
    fi
    
    # 检查 npm
    if ! check_command "npm"; then
        log_error "请安装 npm"
        exit 1
    fi
    
    # 检查 MySQL
    if ! check_command "mysql"; then
        log_error "请安装 MySQL 8.0+"
        exit 1
    fi
    
    # 检查 Redis
    if ! check_command "redis-server"; then
        log_error "请安装 Redis 6.0+"
        exit 1
    fi
    
    log_success "系统依赖检查完成"
}

# 启动 MySQL
start_mysql() {
    log_info "启动 MySQL..."
    
    # 检查 MySQL 是否已经运行
    if pgrep -x "mysqld" > /dev/null; then
        log_success "MySQL 已在运行"
        return 0
    fi
    
    # 尝试启动 MySQL (macOS)
    if [[ "$OSTYPE" == "darwin"* ]]; then
        if command -v brew &> /dev/null; then
            brew services start mysql
        else
            sudo /usr/local/mysql/support-files/mysql.server start
        fi
    else
        # Linux
        sudo systemctl start mysql || sudo service mysql start
    fi
    
    # 等待 MySQL 启动
    wait_for_service "localhost" "3306" "MySQL"
}

# 启动 Redis
start_redis() {
    log_info "启动 Redis..."
    
    # 检查 Redis 是否已经运行
    if pgrep -x "redis-server" > /dev/null; then
        log_success "Redis 已在运行"
        return 0
    fi
    
    # 启动 Redis
    if [[ "$OSTYPE" == "darwin"* ]]; then
        if command -v brew &> /dev/null; then
            brew services start redis
        else
            redis-server --daemonize yes
        fi
    else
        # Linux
        sudo systemctl start redis || sudo service redis start
    fi
    
    # 等待 Redis 启动
    wait_for_service "localhost" "6379" "Redis"
}

# 初始化数据库
init_database() {
    log_info "初始化数据库..."
    
    # 检查数据库是否存在
    if mysql -u root -palong123 -e "USE alert_agent;" 2>/dev/null; then
        log_warning "数据库 alert_agent 已存在，跳过初始化"
        return 0
    fi
    
    # 执行初始化脚本
    if mysql -u root -palong123 < scripts/init.sql; then
        log_success "数据库初始化完成"
    else
        log_error "数据库初始化失败"
        exit 1
    fi
}

# 安装 Go 依赖
install_go_deps() {
    log_info "安装 Go 依赖..."
    
    if go mod download && go mod tidy; then
        log_success "Go 依赖安装完成"
    else
        log_error "Go 依赖安装失败"
        exit 1
    fi
}

# 安装前端依赖
install_frontend_deps() {
    log_info "安装前端依赖..."
    
    cd web
    if npm install; then
        log_success "前端依赖安装完成"
    else
        log_error "前端依赖安装失败"
        exit 1
    fi
    cd ..
}

# 启动后端服务
start_backend() {
    log_info "启动后端服务..."
    
    # 检查端口
    if ! check_port 8080; then
        log_error "后端端口 8080 被占用，请检查"
        return 1
    fi
    
    # 启动后端
    go run cmd/main.go &
    BACKEND_PID=$!
    
    # 等待后端启动
    wait_for_service "localhost" "8080" "后端服务"
    
    echo $BACKEND_PID > .backend.pid
    log_success "后端服务已启动 (PID: $BACKEND_PID)"
}

# 启动前端服务
start_frontend() {
    log_info "启动前端服务..."
    
    # 检查端口
    if ! check_port 5173; then
        log_error "前端端口 5173 被占用，请检查"
        return 1
    fi
    
    cd web
    npm run dev &
    FRONTEND_PID=$!
    cd ..
    
    # 等待前端启动
    wait_for_service "localhost" "5173" "前端服务"
    
    echo $FRONTEND_PID > .frontend.pid
    log_success "前端服务已启动 (PID: $FRONTEND_PID)"
}

# 显示服务信息
show_services() {
    echo
    log_success "=== 开发环境启动完成 ==="
    echo
    echo "📊 服务地址:"
    echo "   前端: http://localhost:5173"
    echo "   后端: http://localhost:8080"
    echo "   API文档: http://localhost:8080/swagger/index.html"
    echo
    echo "🗄️  数据库信息:"
    echo "   MySQL: localhost:3306 (数据库: alert_agent)"
    echo "   Redis: localhost:6379"
    echo
    echo "🔧 管理命令:"
    echo "   停止服务: ./scripts/dev-stop.sh"
    echo "   查看日志: tail -f logs/alert_agent.log"
    echo "   重启服务: ./scripts/dev-restart.sh"
    echo
    log_info "按 Ctrl+C 停止所有服务"
}

# 清理函数
cleanup() {
    log_info "正在停止服务..."
    
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
    
    log_success "开发环境已停止"
    exit 0
}

# 设置信号处理
trap cleanup SIGINT SIGTERM

# 主函数
main() {
    echo "🚀 AlertAgent 开发环境启动脚本"
    echo "================================="
    echo
    
    # 检查依赖
    check_dependencies
    
    # 启动基础服务
    start_mysql
    start_redis
    
    # 初始化数据库
    init_database
    
    # 安装依赖
    install_go_deps
    install_frontend_deps
    
    # 启动应用服务
    start_backend
    start_frontend
    
    # 显示服务信息
    show_services
    
    # 保持脚本运行
    wait
}

# 执行主函数
main "$@"