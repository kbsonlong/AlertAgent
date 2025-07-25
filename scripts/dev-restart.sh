#!/bin/bash

# AlertAgent 开发环境重启脚本
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

# 主函数
main() {
    echo "🔄 AlertAgent 开发环境重启脚本"
    echo "=================================="
    echo
    
    log_info "正在重启开发环境..."
    
    # 停止现有服务
    log_info "停止现有服务..."
    ./scripts/dev-stop.sh
    
    # 等待一下确保进程完全停止
    sleep 2
    
    # 重新启动服务
    log_info "重新启动服务..."
    ./scripts/dev-setup.sh
}

# 检查脚本是否存在
if [ ! -f "scripts/dev-stop.sh" ] || [ ! -f "scripts/dev-setup.sh" ]; then
    log_error "找不到必要的脚本文件，请确保在项目根目录下运行"
    exit 1
fi

# 执行主函数
main "$@"