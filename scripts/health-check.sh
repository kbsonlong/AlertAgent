#!/bin/bash

# AlertAgent 健康检查脚本
# 用于Docker容器健康检查

set -e

# 配置
HEALTH_CHECK_URL=${HEALTH_CHECK_URL:-"http://localhost:8080/api/v1/health"}
TIMEOUT=${TIMEOUT:-10}
MAX_RETRIES=${MAX_RETRIES:-3}

# 日志函数
log_info() {
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] [INFO] $1"
}

log_error() {
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] [ERROR] $1" >&2
}

# 健康检查函数
check_health() {
    local url=$1
    local timeout=$2
    
    if curl -f -s --max-time "$timeout" "$url" > /dev/null 2>&1; then
        return 0
    else
        return 1
    fi
}

# 主函数
main() {
    local retry_count=0
    
    while [ $retry_count -lt $MAX_RETRIES ]; do
        if check_health "$HEALTH_CHECK_URL" "$TIMEOUT"; then
            log_info "健康检查通过: $HEALTH_CHECK_URL"
            exit 0
        else
            retry_count=$((retry_count + 1))
            log_error "健康检查失败 (尝试 $retry_count/$MAX_RETRIES): $HEALTH_CHECK_URL"
            
            if [ $retry_count -lt $MAX_RETRIES ]; then
                sleep 2
            fi
        fi
    done
    
    log_error "健康检查最终失败，已达到最大重试次数"
    exit 1
}

main "$@"