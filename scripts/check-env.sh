#!/bin/bash

# AlertAgent 开发环境检查脚本
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
    echo -e "${GREEN}[✓]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[⚠]${NC} $1"
}

log_error() {
    echo -e "${RED}[✗]${NC} $1"
}

log_section() {
    echo
    echo -e "${BLUE}=== $1 ===${NC}"
}

# 检查命令是否存在
check_command() {
    local cmd=$1
    local name=$2
    local required=${3:-true}
    
    if command -v $cmd &> /dev/null; then
        local version=$($cmd --version 2>/dev/null | head -n1 || echo "未知版本")
        log_success "$name 已安装: $version"
        return 0
    else
        if [ "$required" = "true" ]; then
            log_error "$name 未安装 (必需)"
        else
            log_warning "$name 未安装 (可选)"
        fi
        return 1
    fi
}

# 检查端口是否可用
check_port() {
    local port=$1
    local service=$2
    
    if lsof -Pi :$port -sTCP:LISTEN -t >/dev/null 2>&1; then
        local pid=$(lsof -ti:$port)
        local process=$(ps -p $pid -o comm= 2>/dev/null || echo "未知进程")
        log_warning "端口 $port ($service) 被占用 - PID: $pid, 进程: $process"
        return 1
    else
        log_success "端口 $port ($service) 可用"
        return 0
    fi
}

# 检查服务连接
check_service() {
    local host=$1
    local port=$2
    local service=$3
    
    if nc -z $host $port 2>/dev/null; then
        log_success "$service 服务可连接 ($host:$port)"
        return 0
    else
        log_error "$service 服务不可连接 ($host:$port)"
        return 1
    fi
}

# 检查 Go 环境
check_go_env() {
    log_section "Go 开发环境"
    
    if check_command "go" "Go"; then
        echo "  Go 版本: $(go version)"
        echo "  GOPATH: $(go env GOPATH)"
        echo "  GOROOT: $(go env GOROOT)"
        echo "  GOPROXY: $(go env GOPROXY)"
        
        # 检查 Go 版本
        local go_version=$(go version | grep -oE 'go[0-9]+\.[0-9]+' | sed 's/go//')
        local major=$(echo $go_version | cut -d. -f1)
        local minor=$(echo $go_version | cut -d. -f2)
        
        if [ $major -gt 1 ] || ([ $major -eq 1 ] && [ $minor -ge 21 ]); then
            log_success "Go 版本满足要求 (>= 1.21)"
        else
            log_error "Go 版本过低，需要 1.21 或更高版本"
        fi
        
        # 检查项目依赖
        if [ -f "go.mod" ]; then
            log_info "检查 Go 模块依赖..."
            if go mod verify &>/dev/null; then
                log_success "Go 模块依赖完整"
            else
                log_warning "Go 模块依赖可能有问题，建议运行: go mod download"
            fi
        else
            log_error "未找到 go.mod 文件"
        fi
    fi
}

# 检查 Node.js 环境
check_node_env() {
    log_section "Node.js 开发环境"
    
    if check_command "node" "Node.js"; then
        echo "  Node.js 版本: $(node --version)"
        
        # 检查 Node.js 版本
        local node_version=$(node --version | sed 's/v//' | cut -d. -f1)
        if [ $node_version -ge 18 ]; then
            log_success "Node.js 版本满足要求 (>= 18)"
        else
            log_error "Node.js 版本过低，需要 18 或更高版本"
        fi
    fi
    
    if check_command "npm" "npm"; then
        echo "  npm 版本: $(npm --version)"
        
        # 检查前端依赖
        if [ -f "web/package.json" ]; then
            log_info "检查前端依赖..."
            if [ -d "web/node_modules" ]; then
                log_success "前端依赖已安装"
            else
                log_warning "前端依赖未安装，建议运行: cd web && npm install"
            fi
        else
            log_error "未找到 web/package.json 文件"
        fi
    fi
}

# 检查数据库服务
check_database_services() {
    log_section "数据库服务"
    
    # 检查 MySQL
    if check_command "mysql" "MySQL Client"; then
        if check_service "localhost" "3306" "MySQL"; then
            # 尝试连接数据库
            if mysql -u root -palong123 -e "SELECT 1;" &>/dev/null; then
                log_success "MySQL 数据库连接成功"
                
                # 检查数据库是否存在
                if mysql -u root -palong123 -e "USE alert_agent;" &>/dev/null; then
                    log_success "alert_agent 数据库存在"
                else
                    log_warning "alert_agent 数据库不存在，建议运行: mysql -u root -palong123 < scripts/init.sql"
                fi
            else
                log_error "MySQL 数据库连接失败，请检查用户名密码"
            fi
        fi
    fi
    
    # 检查 Redis
    if check_command "redis-server" "Redis Server" false; then
        if check_service "localhost" "6379" "Redis"; then
            if redis-cli ping &>/dev/null; then
                log_success "Redis 连接成功"
            else
                log_error "Redis 连接失败"
            fi
        fi
    fi
}

# 检查 Docker 环境
check_docker_env() {
    log_section "Docker 环境"
    
    if check_command "docker" "Docker" false; then
        if docker info &>/dev/null; then
            log_success "Docker 服务运行正常"
            echo "  Docker 版本: $(docker --version)"
        else
            log_error "Docker 服务未运行"
        fi
    fi
    
    if check_command "docker-compose" "Docker Compose" false || docker compose version &>/dev/null; then
        if docker compose version &>/dev/null; then
            echo "  Docker Compose 版本: $(docker compose version)"
        else
            echo "  Docker Compose 版本: $(docker-compose --version)"
        fi
        
        # 检查 Docker Compose 文件
        if [ -f "docker-compose.dev.yml" ]; then
            log_success "Docker Compose 配置文件存在"
        else
            log_error "未找到 docker-compose.dev.yml 文件"
        fi
    fi
}

# 检查端口占用
check_ports() {
    log_section "端口检查"
    
    check_port "3306" "MySQL"
    check_port "6379" "Redis"
    check_port "8080" "后端服务"
    check_port "5173" "前端服务"
    check_port "11434" "Ollama"
    check_port "8081" "phpMyAdmin"
    check_port "8082" "Redis Commander"
}

# 检查项目文件
check_project_files() {
    log_section "项目文件检查"
    
    local required_files=(
        "go.mod"
        "cmd/main.go"
        "config/config.yaml"
        "scripts/init.sql"
        "web/package.json"
        "Makefile"
    )
    
    for file in "${required_files[@]}"; do
        if [ -f "$file" ]; then
            log_success "$file 存在"
        else
            log_error "$file 不存在"
        fi
    done
    
    local required_dirs=(
        "internal"
        "scripts"
        "web/src"
        "docs"
    )
    
    for dir in "${required_dirs[@]}"; do
        if [ -d "$dir" ]; then
            log_success "$dir/ 目录存在"
        else
            log_error "$dir/ 目录不存在"
        fi
    done
}

# 检查可选服务
check_optional_services() {
    log_section "可选服务"
    
    # 检查 Ollama
    if check_command "ollama" "Ollama" false; then
        if check_service "localhost" "11434" "Ollama"; then
            log_success "Ollama 服务可用"
        else
            log_warning "Ollama 服务不可用，AI 功能将无法使用"
        fi
    else
        log_warning "Ollama 未安装，AI 功能将无法使用"
    fi
    
    # 检查开发工具
    check_command "golangci-lint" "golangci-lint" false
    check_command "air" "air (热重载工具)" false
    check_command "git" "Git" false
    check_command "curl" "curl" false
    check_command "nc" "netcat" false
}

# 生成建议
generate_suggestions() {
    log_section "建议和下一步"
    
    echo "基于检查结果，建议您："
    echo
    
    if ! command -v go &> /dev/null; then
        echo "1. 安装 Go 1.21+: https://golang.org/dl/"
    fi
    
    if ! command -v node &> /dev/null; then
        echo "2. 安装 Node.js 18+: https://nodejs.org/"
    fi
    
    if ! command -v mysql &> /dev/null; then
        echo "3. 安装 MySQL 8.0+"
    fi
    
    if ! command -v redis-server &> /dev/null; then
        echo "4. 安装 Redis 6.0+"
    fi
    
    if [ ! -f "config/config.yaml" ]; then
        echo "5. 复制配置文件: cp config/config.yaml.example config/config.yaml"
    fi
    
    if [ ! -d "web/node_modules" ]; then
        echo "6. 安装前端依赖: cd web && npm install"
    fi
    
    if ! mysql -u root -palong123 -e "USE alert_agent;" &>/dev/null; then
        echo "7. 初始化数据库: mysql -u root -palong123 < scripts/init.sql"
    fi
    
    echo
    echo "准备就绪后，运行以下命令启动开发环境："
    echo "  make dev          # 本地环境"
    echo "  make docker-dev   # Docker 环境"
    echo
    echo "获取更多帮助："
    echo "  make help         # 查看所有命令"
    echo "  docs/quick-start.md # 详细安装指南"
}

# 主函数
main() {
    echo "🔍 AlertAgent 开发环境检查"
    echo "============================"
    echo
    
    # 检查各个组件
    check_go_env
    check_node_env
    check_database_services
    check_docker_env
    check_ports
    check_project_files
    check_optional_services
    
    # 生成建议
    generate_suggestions
    
    echo
    log_info "环境检查完成！"
}

# 执行主函数
main "$@"