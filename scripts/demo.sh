#!/bin/bash

# AlertAgent 演示脚本
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

log_demo() {
    echo -e "${YELLOW}[DEMO]${NC} $1"
}

# 等待用户输入
wait_for_user() {
    echo
    read -p "按 Enter 键继续..."
    echo
}

# 检查服务是否运行
check_service() {
    local url=$1
    local service_name=$2
    local max_attempts=10
    local attempt=1
    
    while [ $attempt -le $max_attempts ]; do
        if curl -s $url >/dev/null 2>&1; then
            log_success "$service_name 服务正在运行"
            return 0
        fi
        
        echo -n "."
        sleep 1
        attempt=$((attempt + 1))
    done
    
    echo
    log_error "$service_name 服务未运行，请先启动开发环境"
    return 1
}

# API 演示
demo_api_calls() {
    log_demo "演示 API 调用"
    echo "以下是一些常用的 API 调用示例："
    echo
    
    # 健康检查
    log_info "1. 健康检查"
    echo "curl http://localhost:8080/health"
    if curl -s http://localhost:8080/health; then
        echo
        log_success "健康检查成功"
    else
        log_error "健康检查失败"
    fi
    
    wait_for_user
    
    # 获取告警列表
    log_info "2. 获取告警列表"
    echo "curl http://localhost:8080/api/v1/alerts"
    if curl -s http://localhost:8080/api/v1/alerts | jq . 2>/dev/null || curl -s http://localhost:8080/api/v1/alerts; then
        echo
        log_success "获取告警列表成功"
    else
        log_error "获取告警列表失败"
    fi
    
    wait_for_user
    
    # 创建告警
    log_info "3. 创建示例告警"
    local alert_data='{
        "name": "演示告警",
        "level": "warning",
        "source": "demo",
        "content": "这是一个演示告警",
        "rule_id": 1,
        "title": "演示告警标题"
    }'
    
    echo "curl -X POST http://localhost:8080/api/v1/alerts \\"
    echo "  -H 'Content-Type: application/json' \\"
    echo "  -d '$alert_data'"
    
    if curl -s -X POST http://localhost:8080/api/v1/alerts \
        -H "Content-Type: application/json" \
        -d "$alert_data" | jq . 2>/dev/null || \
       curl -s -X POST http://localhost:8080/api/v1/alerts \
        -H "Content-Type: application/json" \
        -d "$alert_data"; then
        echo
        log_success "创建告警成功"
    else
        log_error "创建告警失败"
    fi
    
    wait_for_user
}

# 前端演示
demo_frontend() {
    log_demo "演示前端功能"
    echo "前端应用提供了以下功能："
    echo
    echo "📊 主要功能:"
    echo "  - 告警管理: 查看、创建、更新告警"
    echo "  - 规则管理: 配置告警规则"
    echo "  - 通知管理: 设置通知方式和模板"
    echo "  - 知识库: AI 智能分析和建议"
    echo
    echo "🌐 访问地址:"
    echo "  - 前端应用: http://localhost:5173"
    echo "  - API 文档: http://localhost:8080/swagger/index.html"
    echo
    
    if command -v open >/dev/null 2>&1; then
        read -p "是否打开前端应用？(y/N): " -n 1 -r
        echo
        if [[ $REPLY =~ ^[Yy]$ ]]; then
            log_info "正在打开前端应用..."
            open http://localhost:5173
        fi
    elif command -v xdg-open >/dev/null 2>&1; then
        read -p "是否打开前端应用？(y/N): " -n 1 -r
        echo
        if [[ $REPLY =~ ^[Yy]$ ]]; then
            log_info "正在打开前端应用..."
            xdg-open http://localhost:5173
        fi
    else
        log_info "请手动访问: http://localhost:5173"
    fi
    
    wait_for_user
}

# 数据库演示
demo_database() {
    log_demo "演示数据库操作"
    echo "数据库包含以下表："
    echo
    
    if mysql -u root -palong123 alert_agent -e "SHOW TABLES;" 2>/dev/null; then
        log_success "数据库连接成功"
        echo
        
        log_info "查看告警数据:"
        mysql -u root -palong123 alert_agent -e "SELECT id, name, level, status, created_at FROM alerts LIMIT 5;" 2>/dev/null || log_warning "暂无告警数据"
        echo
        
        log_info "查看规则数据:"
        mysql -u root -palong123 alert_agent -e "SELECT id, name, level, enabled, created_at FROM rules LIMIT 5;" 2>/dev/null || log_warning "暂无规则数据"
    else
        log_error "数据库连接失败，请检查 MySQL 服务和配置"
    fi
    
    wait_for_user
}

# Docker 环境演示
demo_docker() {
    log_demo "演示 Docker 环境"
    
    if command -v docker >/dev/null 2>&1; then
        echo "Docker 容器状态:"
        docker ps --filter "name=alertagent" --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}" 2>/dev/null || echo "没有运行的 AlertAgent 容器"
        echo
        
        echo "🔧 管理工具:"
        echo "  - phpMyAdmin: http://localhost:8081 (用户名: root, 密码: along123)"
        echo "  - Redis Commander: http://localhost:8082"
        echo
        
        if docker ps --filter "name=alertagent" --quiet | grep -q .; then
            log_info "Docker 服务正在运行"
            
            read -p "是否查看容器日志？(y/N): " -n 1 -r
            echo
            if [[ $REPLY =~ ^[Yy]$ ]]; then
                log_info "显示最近 20 行日志..."
                if docker compose version >/dev/null 2>&1; then
                    docker compose -f docker-compose.dev.yml logs --tail=20
                else
                    docker-compose -f docker-compose.dev.yml logs --tail=20
                fi
            fi
        else
            log_warning "Docker 服务未运行，请先启动: make docker-dev"
        fi
    else
        log_warning "Docker 未安装"
    fi
    
    wait_for_user
}

# 开发工具演示
demo_dev_tools() {
    log_demo "演示开发工具"
    echo "项目提供了以下开发工具："
    echo
    
    echo "📋 Makefile 命令:"
    echo "  make help         # 查看所有命令"
    echo "  make dev          # 启动本地开发环境"
    echo "  make docker-dev   # 启动 Docker 开发环境"
    echo "  make test         # 运行测试"
    echo "  make lint         # 代码检查"
    echo "  make build        # 构建项目"
    echo
    
    echo "🔧 脚本工具:"
    echo "  scripts/dev-setup.sh      # 本地环境启动"
    echo "  scripts/docker-dev-setup.sh # Docker 环境启动"
    echo "  scripts/check-env.sh       # 环境检查"
    echo "  scripts/demo.sh            # 演示脚本（当前）"
    echo
    
    read -p "是否运行环境检查？(y/N): " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        ./scripts/check-env.sh
    fi
    
    wait_for_user
}

# 显示帮助信息
show_help() {
    echo "AlertAgent 演示脚本"
    echo "=================="
    echo
    echo "用法: $0 [选项]"
    echo
    echo "选项:"
    echo "  --api         仅演示 API 功能"
    echo "  --frontend    仅演示前端功能"
    echo "  --database    仅演示数据库功能"
    echo "  --docker      仅演示 Docker 功能"
    echo "  --dev-tools   仅演示开发工具"
    echo "  --help        显示此帮助信息"
    echo
    echo "不带参数运行将进行完整演示。"
}

# 主函数
main() {
    local demo_type="all"
    
    # 解析命令行参数
    while [[ $# -gt 0 ]]; do
        case $1 in
            --api)
                demo_type="api"
                shift
                ;;
            --frontend)
                demo_type="frontend"
                shift
                ;;
            --database)
                demo_type="database"
                shift
                ;;
            --docker)
                demo_type="docker"
                shift
                ;;
            --dev-tools)
                demo_type="dev-tools"
                shift
                ;;
            --help)
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
    
    echo "🎯 AlertAgent 功能演示"
    echo "======================"
    echo
    
    # 检查服务状态
    if [ "$demo_type" = "all" ] || [ "$demo_type" = "api" ] || [ "$demo_type" = "frontend" ]; then
        log_info "检查服务状态..."
        if ! check_service "http://localhost:8080/health" "后端"; then
            log_error "请先启动开发环境: make dev 或 make docker-dev"
            exit 1
        fi
        
        if ! check_service "http://localhost:5173" "前端"; then
            log_warning "前端服务未运行，部分演示可能无法进行"
        fi
    fi
    
    # 根据参数执行相应演示
    case $demo_type in
        "api")
            demo_api_calls
            ;;
        "frontend")
            demo_frontend
            ;;
        "database")
            demo_database
            ;;
        "docker")
            demo_docker
            ;;
        "dev-tools")
            demo_dev_tools
            ;;
        "all")
            demo_api_calls
            demo_frontend
            demo_database
            demo_docker
            demo_dev_tools
            ;;
    esac
    
    echo
    log_success "演示完成！"
    echo
    echo "📚 更多信息:"
    echo "  - API 文档: http://localhost:8080/swagger/index.html"
    echo "  - 项目文档: docs/"
    echo "  - 快速开始: docs/quick-start.md"
    echo "  - 获取帮助: make help"
}

# 执行主函数
main "$@"