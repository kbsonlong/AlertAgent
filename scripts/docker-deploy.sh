#!/bin/bash

# AlertAgent Docker部署脚本
# 支持开发、测试、生产环境部署

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

# 检查环境
check_environment() {
    # 检查Docker和Docker Compose
    if ! command -v docker &> /dev/null; then
        log_error "Docker未安装，请先安装Docker"
        exit 1
    fi
    
    if ! command -v docker-compose &> /dev/null && ! docker compose version &> /dev/null; then
        log_error "Docker Compose未安装，请先安装Docker Compose"
        exit 1
    fi
    
    # 检查配置文件
    if [ ! -f "docker-compose.yml" ]; then
        log_error "docker-compose.yml文件不存在"
        exit 1
    fi
    
    log_success "环境检查通过"
}

# 创建环境配置文件
create_env_file() {
    local env=$1
    local env_file=".env.${env}"
    
    if [ ! -f "$env_file" ]; then
        log_info "创建环境配置文件: $env_file"
        
        case $env in
            "dev")
                cat > "$env_file" << EOF
# AlertAgent开发环境配置
COMPOSE_PROJECT_NAME=alertagent-dev
MYSQL_ROOT_PASSWORD=along123
MYSQL_DATABASE=alert_agent
MYSQL_USER=alertagent
MYSQL_PASSWORD=alertagent123
REDIS_PASSWORD=
OLLAMA_ENDPOINT=http://ollama:11434
JWT_SECRET=dev-jwt-secret-key-change-in-production
SERVER_MODE=debug
AI_WORKER_CONCURRENCY=2
NOTIFICATION_WORKER_CONCURRENCY=3
CONFIG_WORKER_CONCURRENCY=2
AI_WORKER_REPLICAS=1
NOTIFICATION_WORKER_REPLICAS=1
CONFIG_WORKER_REPLICAS=1
EOF
                ;;
            "test")
                cat > "$env_file" << EOF
# AlertAgent测试环境配置
COMPOSE_PROJECT_NAME=alertagent-test
MYSQL_ROOT_PASSWORD=test123456
MYSQL_DATABASE=alert_agent_test
MYSQL_USER=alertagent
MYSQL_PASSWORD=test123456
REDIS_PASSWORD=test123456
OLLAMA_ENDPOINT=http://ollama:11434
JWT_SECRET=test-jwt-secret-key-change-in-production
SERVER_MODE=release
AI_WORKER_CONCURRENCY=2
NOTIFICATION_WORKER_CONCURRENCY=3
CONFIG_WORKER_CONCURRENCY=2
AI_WORKER_REPLICAS=1
NOTIFICATION_WORKER_REPLICAS=1
CONFIG_WORKER_REPLICAS=1
EOF
                ;;
            "prod")
                cat > "$env_file" << EOF
# AlertAgent生产环境配置
COMPOSE_PROJECT_NAME=alertagent-prod
MYSQL_ROOT_PASSWORD=CHANGE_ME_STRONG_PASSWORD
MYSQL_DATABASE=alert_agent
MYSQL_USER=alertagent
MYSQL_PASSWORD=CHANGE_ME_STRONG_PASSWORD
REDIS_PASSWORD=CHANGE_ME_STRONG_PASSWORD
OLLAMA_ENDPOINT=http://ollama:11434
JWT_SECRET=CHANGE_ME_STRONG_JWT_SECRET
SERVER_MODE=release
AI_WORKER_CONCURRENCY=4
NOTIFICATION_WORKER_CONCURRENCY=5
CONFIG_WORKER_CONCURRENCY=3
AI_WORKER_REPLICAS=2
NOTIFICATION_WORKER_REPLICAS=2
CONFIG_WORKER_REPLICAS=1
EOF
                log_warning "请修改 $env_file 中的密码和密钥配置！"
                ;;
        esac
        
        log_success "环境配置文件创建完成: $env_file"
    else
        log_info "环境配置文件已存在: $env_file"
    fi
}

# 部署服务
deploy_services() {
    local env=$1
    local action=$2
    local compose_file="docker-compose.yml"
    local env_file=".env.${env}"
    
    # 选择compose文件
    case $env in
        "dev")
            compose_file="docker-compose.dev.yml"
            ;;
        "test")
            compose_file="docker-compose.test.yml"
            ;;
        "prod")
            compose_file="docker-compose.prod.yml"
            ;;
    esac
    
    if [ ! -f "$compose_file" ]; then
        log_error "Docker Compose文件不存在: $compose_file"
        exit 1
    fi
    
    log_info "使用配置文件: $compose_file"
    log_info "使用环境文件: $env_file"
    
    # 执行操作
    case $action in
        "up")
            log_info "启动 $env 环境服务..."
            docker-compose -f "$compose_file" --env-file "$env_file" up -d
            ;;
        "down")
            log_info "停止 $env 环境服务..."
            docker-compose -f "$compose_file" --env-file "$env_file" down
            ;;
        "restart")
            log_info "重启 $env 环境服务..."
            docker-compose -f "$compose_file" --env-file "$env_file" restart
            ;;
        "logs")
            log_info "查看 $env 环境日志..."
            docker-compose -f "$compose_file" --env-file "$env_file" logs -f
            ;;
        "ps")
            log_info "查看 $env 环境服务状态..."
            docker-compose -f "$compose_file" --env-file "$env_file" ps
            ;;
        "pull")
            log_info "拉取 $env 环境镜像..."
            docker-compose -f "$compose_file" --env-file "$env_file" pull
            ;;
        *)
            log_error "不支持的操作: $action"
            exit 1
            ;;
    esac
}

# 健康检查
health_check() {
    local env=$1
    
    log_info "执行健康检查..."
    
    # 检查核心服务
    local core_url="http://localhost:8080/api/v1/health"
    local max_attempts=30
    local attempt=1
    
    while [ $attempt -le $max_attempts ]; do
        if curl -f -s "$core_url" > /dev/null 2>&1; then
            log_success "AlertAgent Core服务健康检查通过"
            break
        else
            log_info "等待AlertAgent Core服务启动... ($attempt/$max_attempts)"
            sleep 5
            ((attempt++))
        fi
    done
    
    if [ $attempt -gt $max_attempts ]; then
        log_error "AlertAgent Core服务健康检查失败"
        return 1
    fi
    
    # 检查数据库连接
    log_info "检查数据库连接..."
    if curl -f -s "${core_url}" | grep -q "database.*ok"; then
        log_success "数据库连接正常"
    else
        log_warning "数据库连接可能存在问题"
    fi
    
    # 检查Redis连接
    log_info "检查Redis连接..."
    if curl -f -s "${core_url}" | grep -q "redis.*ok"; then
        log_success "Redis连接正常"
    else
        log_warning "Redis连接可能存在问题"
    fi
    
    log_success "健康检查完成"
}

# 显示服务状态
show_status() {
    local env=$1
    
    log_info "服务状态概览:"
    echo
    
    # 显示容器状态
    docker ps --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}" | grep alertagent
    
    echo
    log_info "服务访问地址:"
    echo "  - AlertAgent Web UI: http://localhost:8080"
    echo "  - API文档: http://localhost:8080/api/v1/docs"
    
    if [ "$env" = "dev" ]; then
        echo "  - phpMyAdmin: http://localhost:8081"
        echo "  - Redis Commander: http://localhost:8082"
    fi
}

# 显示帮助信息
show_help() {
    echo "AlertAgent Docker部署脚本"
    echo
    echo "用法: $0 <environment> <action> [options]"
    echo
    echo "环境:"
    echo "  dev     开发环境"
    echo "  test    测试环境"
    echo "  prod    生产环境"
    echo
    echo "操作:"
    echo "  up      启动服务"
    echo "  down    停止服务"
    echo "  restart 重启服务"
    echo "  logs    查看日志"
    echo "  ps      查看服务状态"
    echo "  pull    拉取镜像"
    echo "  status  显示服务状态"
    echo "  health  健康检查"
    echo
    echo "示例:"
    echo "  $0 dev up          # 启动开发环境"
    echo "  $0 prod down       # 停止生产环境"
    echo "  $0 test logs       # 查看测试环境日志"
    echo "  $0 dev status      # 显示开发环境状态"
    echo
    echo "选项:"
    echo "  -h, --help    显示此帮助信息"
}

# 主函数
main() {
    local env=$1
    local action=$2
    
    if [ -z "$env" ] || [ -z "$action" ]; then
        show_help
        exit 1
    fi
    
    # 验证环境参数
    case $env in
        "dev"|"test"|"prod")
            ;;
        *)
            log_error "不支持的环境: $env"
            show_help
            exit 1
            ;;
    esac
    
    # 检查环境
    check_environment
    
    # 创建环境配置文件
    create_env_file "$env"
    
    # 执行操作
    case $action in
        "up"|"down"|"restart"|"logs"|"ps"|"pull")
            deploy_services "$env" "$action"
            if [ "$action" = "up" ]; then
                sleep 10
                health_check "$env"
                show_status "$env"
            fi
            ;;
        "status")
            show_status "$env"
            ;;
        "health")
            health_check "$env"
            ;;
        *)
            log_error "不支持的操作: $action"
            show_help
            exit 1
            ;;
    esac
    
    log_success "操作完成: $env $action"
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