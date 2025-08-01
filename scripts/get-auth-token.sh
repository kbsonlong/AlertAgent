#!/bin/bash

# AlertAgent API 认证 Token 获取脚本
# 用于解决 401 未授权错误

set -e

# 配置
API_BASE_URL="http://localhost:8080"
USERNAME="admin"
PASSWORD="password"

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

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

# 检查服务是否运行
check_service() {
    log_info "检查 AlertAgent 服务状态..."
    
    if ! curl -s "$API_BASE_URL/api/v1/health" > /dev/null; then
        log_error "AlertAgent 服务未运行或无法访问"
        log_info "请先启动服务: make run"
        exit 1
    fi
    
    log_success "AlertAgent 服务正常运行"
}

# 获取认证 Token
get_token() {
    log_info "正在获取认证 Token..."
    
    local response
    response=$(curl -s -X POST "$API_BASE_URL/api/v1/auth/login" \
        -H "Content-Type: application/json" \
        -d "{\"username\":\"$USERNAME\",\"password\":\"$PASSWORD\"}")
    
    if [ $? -ne 0 ]; then
        log_error "登录请求失败"
        exit 1
    fi
    
    # 检查响应是否包含 token
    local token
    token=$(echo "$response" | jq -r '.data.access_token // empty' 2>/dev/null)
    
    if [ -z "$token" ] || [ "$token" = "null" ]; then
        log_error "登录失败，请检查用户名和密码"
        echo "响应: $response"
        exit 1
    fi
    
    log_success "Token 获取成功"
    echo "$token"
}

# 测试 Token
test_token() {
    local token="$1"
    
    log_info "测试 Token 有效性..."
    
    local response
    response=$(curl -s -H "Authorization: Bearer $token" \
        "$API_BASE_URL/api/v1/knowledge?page=1&pageSize=1")
    
    if [ $? -ne 0 ]; then
        log_error "Token 测试失败"
        return 1
    fi
    
    local code
    code=$(echo "$response" | jq -r '.code // empty' 2>/dev/null)
    
    if [ "$code" = "200" ]; then
        log_success "Token 验证通过"
        return 0
    else
        log_error "Token 验证失败"
        echo "响应: $response"
        return 1
    fi
}

# 显示使用示例
show_usage_examples() {
    local token="$1"
    
    echo ""
    log_info "使用示例:"
    echo ""
    echo "# 设置环境变量"
    echo "export AUTH_TOKEN='$token'"
    echo ""
    echo "# 使用 curl 访问 API"
    echo "curl -H \"Authorization: Bearer \$AUTH_TOKEN\" \"$API_BASE_URL/api/v1/knowledge\""
    echo ""
    echo "# 获取知识库列表"
    echo "curl -H \"Authorization: Bearer \$AUTH_TOKEN\" \"$API_BASE_URL/api/v1/knowledge?page=1&pageSize=10\""
    echo ""
    echo "# 获取健康状态（无需认证）"
    echo "curl \"$API_BASE_URL/api/v1/health\""
    echo ""
    echo "# 保存 Token 到文件"
    echo "echo '$token' > .auth_token"
    echo ""
    echo "# 从文件读取 Token"
    echo "TOKEN=\$(cat .auth_token)"
    echo "curl -H \"Authorization: Bearer \$TOKEN\" \"$API_BASE_URL/api/v1/knowledge\""
}

# 主函数
main() {
    echo "AlertAgent API 认证 Token 获取工具"
    echo "===================================="
    echo ""
    
    # 检查依赖
    if ! command -v curl &> /dev/null; then
        log_error "curl 未安装，请先安装 curl"
        exit 1
    fi
    
    if ! command -v jq &> /dev/null; then
        log_error "jq 未安装，请先安装 jq"
        exit 1
    fi
    
    # 检查服务
    check_service
    
    # 获取 Token
    local token
    token=$(get_token)
    
    # 测试 Token
    if test_token "$token"; then
        echo ""
        log_success "认证 Token:"
        echo "$token"
        
        # 显示使用示例
        show_usage_examples "$token"
        
        # 保存到文件
        echo "$token" > .auth_token
        log_success "Token 已保存到 .auth_token 文件"
    else
        log_error "Token 验证失败，请检查服务状态"
        exit 1
    fi
}

# 处理命令行参数
case "${1:-}" in
    --help|-h)
        echo "用法: $0 [选项]"
        echo ""
        echo "选项:"
        echo "  --help, -h     显示帮助信息"
        echo "  --token-only   仅输出 Token（用于脚本）"
        echo "  --test         测试现有 Token"
        echo ""
        echo "环境变量:"
        echo "  API_BASE_URL   API 基础 URL (默认: http://localhost:8080)"
        echo "  USERNAME       用户名 (默认: admin)"
        echo "  PASSWORD       密码 (默认: admin123)"
        exit 0
        ;;
    --token-only)
        check_service
        get_token
        exit 0
        ;;
    --test)
        if [ -f ".auth_token" ]; then
            token=$(cat .auth_token)
            test_token "$token"
        else
            log_error "未找到 .auth_token 文件"
            exit 1
        fi
        exit 0
        ;;
    "")
        main
        ;;
    *)
        log_error "未知参数: $1"
        echo "使用 --help 查看帮助信息"
        exit 1
        ;;
esac