#!/bin/bash

# AlertAgent 数据库迁移快速设置脚本
# 用于快速启动开发环境并执行数据库迁移

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

# 检查依赖
check_dependencies() {
    log_info "检查依赖..."
    
    if ! command -v docker &> /dev/null; then
        log_error "Docker 未安装，请先安装 Docker"
        exit 1
    fi
    
    if ! command -v docker-compose &> /dev/null; then
        log_error "Docker Compose 未安装，请先安装 Docker Compose"
        exit 1
    fi
    
    if ! command -v go &> /dev/null; then
        log_error "Go 未安装，请先安装 Go 1.20+"
        exit 1
    fi
    
    log_success "依赖检查完成"
}

# 清理旧容器
cleanup_containers() {
    log_info "清理旧容器..."
    
    # 停止并删除相关容器
    docker-compose -f docker-compose.dev.yml down --remove-orphans 2>/dev/null || true
    
    # 删除迁移相关的容器
    docker rm -f alertagent-migrate 2>/dev/null || true
    
    log_success "容器清理完成"
}

# 启动数据库服务
start_database() {
    log_info "启动 PostgreSQL 数据库..."
    
    # 启动数据库服务
    docker-compose -f docker-compose.dev.yml up -d postgres
    
    # 等待数据库启动
    log_info "等待数据库启动..."
    timeout=60
    counter=0
    
    while [ $counter -lt $timeout ]; do
        if docker-compose -f docker-compose.dev.yml exec -T postgres pg_isready -U postgres -d alert_agent &>/dev/null; then
            log_success "数据库启动成功"
            return 0
        fi
        
        sleep 2
        counter=$((counter + 2))
        echo -n "."
    done
    
    log_error "数据库启动超时"
    exit 1
}

# 构建迁移工具
build_migrate_tool() {
    log_info "构建数据库迁移工具..."
    
    # 本地构建
    if make migrate-build; then
        log_success "本地迁移工具构建成功"
    else
        log_error "本地迁移工具构建失败"
        exit 1
    fi
    
    # Docker 构建
    if make migrate-docker-build; then
        log_success "Docker 迁移镜像构建成功"
    else
        log_error "Docker 迁移镜像构建失败"
        exit 1
    fi
}

# 执行数据库迁移
run_migration() {
    log_info "执行数据库迁移..."
    
    # 设置环境变量
    export DB_HOST=localhost
    export DB_PORT=5432
    export DB_USER=postgres
    export DB_PASSWORD=password
    export DB_NAME=alert_agent
    export LOG_LEVEL=info
    
    # 执行迁移
    if ./bin/migrate -action=migrate; then
        log_success "数据库迁移执行成功"
    else
        log_error "数据库迁移执行失败"
        exit 1
    fi
}

# 验证迁移结果
validate_migration() {
    log_info "验证迁移结果..."
    
    # 检查迁移状态
    if ./bin/migrate -action=status; then
        log_success "迁移状态检查通过"
    else
        log_warning "迁移状态检查失败"
    fi
    
    # 验证数据库
    if ./bin/migrate -action=validate; then
        log_success "数据库验证通过"
    else
        log_warning "数据库验证失败"
    fi
    
    # 显示详细信息
    log_info "迁移详细信息:"
    ./bin/migrate -action=info
}

# 显示连接信息
show_connection_info() {
    log_info "数据库连接信息:"
    echo "  主机: localhost"
    echo "  端口: 5432"
    echo "  数据库: alert_agent"
    echo "  用户名: postgres"
    echo "  密码: password"
    echo ""
    echo "连接字符串:"
    echo "  postgresql://postgres:password@localhost:5432/alert_agent?sslmode=disable"
    echo ""
    echo "Docker 网络连接:"
    echo "  postgresql://postgres:password@postgres:5432/alert_agent?sslmode=disable"
}

# 显示使用帮助
show_help() {
    echo "AlertAgent 数据库迁移设置脚本"
    echo ""
    echo "用法: $0 [选项]"
    echo ""
    echo "选项:"
    echo "  -h, --help     显示帮助信息"
    echo "  -c, --clean    清理并重新开始"
    echo "  -s, --status   仅检查迁移状态"
    echo "  -v, --validate 仅验证数据库"
    echo "  -r, --rollback 回滚到指定版本 (需要 -t 参数)"
    echo "  -t, --target   目标版本 (与 -r 一起使用)"
    echo ""
    echo "示例:"
    echo "  $0                    # 完整设置和迁移"
    echo "  $0 -c                 # 清理并重新设置"
    echo "  $0 -s                 # 检查迁移状态"
    echo "  $0 -r -t v2.0.0-001   # 回滚到指定版本"
}

# 主函数
main() {
    local clean_mode=false
    local status_only=false
    local validate_only=false
    local rollback_mode=false
    local target_version=""
    
    # 解析命令行参数
    while [[ $# -gt 0 ]]; do
        case $1 in
            -h|--help)
                show_help
                exit 0
                ;;
            -c|--clean)
                clean_mode=true
                shift
                ;;
            -s|--status)
                status_only=true
                shift
                ;;
            -v|--validate)
                validate_only=true
                shift
                ;;
            -r|--rollback)
                rollback_mode=true
                shift
                ;;
            -t|--target)
                target_version="$2"
                shift 2
                ;;
            *)
                log_error "未知参数: $1"
                show_help
                exit 1
                ;;
        esac
    done
    
    # 检查回滚参数
    if [ "$rollback_mode" = true ] && [ -z "$target_version" ]; then
        log_error "回滚模式需要指定目标版本 (-t 参数)"
        exit 1
    fi
    
    log_info "AlertAgent 数据库迁移设置开始..."
    
    # 检查依赖
    check_dependencies
    
    # 仅状态检查
    if [ "$status_only" = true ]; then
        export DB_HOST=localhost
        export DB_PORT=5432
        export DB_USER=postgres
        export DB_PASSWORD=password
        export DB_NAME=alert_agent
        
        if [ -f "./bin/migrate" ]; then
            ./bin/migrate -action=status
        else
            log_error "迁移工具未找到，请先运行完整设置"
            exit 1
        fi
        exit 0
    fi
    
    # 仅验证
    if [ "$validate_only" = true ]; then
        export DB_HOST=localhost
        export DB_PORT=5432
        export DB_USER=postgres
        export DB_PASSWORD=password
        export DB_NAME=alert_agent
        
        if [ -f "./bin/migrate" ]; then
            ./bin/migrate -action=validate
        else
            log_error "迁移工具未找到，请先运行完整设置"
            exit 1
        fi
        exit 0
    fi
    
    # 清理模式
    if [ "$clean_mode" = true ]; then
        cleanup_containers
    fi
    
    # 启动数据库
    start_database
    
    # 构建迁移工具
    build_migrate_tool
    
    # 回滚模式
    if [ "$rollback_mode" = true ]; then
        log_info "回滚到版本: $target_version"
        export DB_HOST=localhost
        export DB_PORT=5432
        export DB_USER=postgres
        export DB_PASSWORD=password
        export DB_NAME=alert_agent
        
        if ./bin/migrate -action=rollback -version="$target_version"; then
            log_success "回滚成功"
        else
            log_error "回滚失败"
            exit 1
        fi
    else
        # 执行迁移
        run_migration
    fi
    
    # 验证结果
    validate_migration
    
    # 显示连接信息
    show_connection_info
    
    log_success "数据库迁移设置完成！"
    log_info "你现在可以启动 AlertAgent 应用了"
}

# 执行主函数
main "$@"