#!/bin/bash

# AlertAgent Docker构建脚本
# 用于构建所有Docker镜像

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

# 检查Docker是否安装
check_docker() {
    if ! command -v docker &> /dev/null; then
        log_error "Docker未安装，请先安装Docker"
        exit 1
    fi
    
    if ! docker info &> /dev/null; then
        log_error "Docker服务未启动，请启动Docker服务"
        exit 1
    fi
    
    log_success "Docker环境检查通过"
}

# 构建镜像函数
build_image() {
    local dockerfile=$1
    local image_name=$2
    local context=${3:-.}
    
    log_info "构建镜像: $image_name"
    log_info "使用Dockerfile: $dockerfile"
    
    if docker build -f "$dockerfile" -t "$image_name" "$context"; then
        log_success "镜像构建成功: $image_name"
    else
        log_error "镜像构建失败: $image_name"
        exit 1
    fi
}

# 主函数
main() {
    log_info "开始构建AlertAgent Docker镜像..."
    
    # 检查Docker环境
    check_docker
    
    # 获取版本信息
    VERSION=${1:-latest}
    REGISTRY=${DOCKER_REGISTRY:-""}
    
    if [ -n "$REGISTRY" ]; then
        IMAGE_PREFIX="${REGISTRY}/"
    else
        IMAGE_PREFIX=""
    fi
    
    log_info "构建版本: $VERSION"
    log_info "镜像前缀: $IMAGE_PREFIX"
    
    # 构建AlertAgent Core镜像
    build_image "Dockerfile" "${IMAGE_PREFIX}alertagent-core:${VERSION}"
    
    # 构建Worker镜像
    build_image "Dockerfile.worker" "${IMAGE_PREFIX}alertagent-worker:${VERSION}"
    
    # 构建Sidecar镜像
    build_image "Dockerfile.sidecar" "${IMAGE_PREFIX}alertagent-sidecar:${VERSION}"
    
    log_success "所有镜像构建完成！"
    
    # 显示构建的镜像
    log_info "构建的镜像列表:"
    docker images | grep -E "(alertagent-core|alertagent-worker|alertagent-sidecar)" | grep "$VERSION"
    
    # 如果指定了registry，询问是否推送
    if [ -n "$REGISTRY" ]; then
        echo
        read -p "是否推送镜像到registry? (y/N): " -n 1 -r
        echo
        if [[ $REPLY =~ ^[Yy]$ ]]; then
            log_info "推送镜像到registry..."
            docker push "${IMAGE_PREFIX}alertagent-core:${VERSION}"
            docker push "${IMAGE_PREFIX}alertagent-worker:${VERSION}"
            docker push "${IMAGE_PREFIX}alertagent-sidecar:${VERSION}"
            log_success "镜像推送完成！"
        fi
    fi
}

# 显示帮助信息
show_help() {
    echo "AlertAgent Docker构建脚本"
    echo
    echo "用法: $0 [VERSION] [OPTIONS]"
    echo
    echo "参数:"
    echo "  VERSION     镜像版本标签 (默认: latest)"
    echo
    echo "环境变量:"
    echo "  DOCKER_REGISTRY  Docker镜像仓库地址"
    echo
    echo "示例:"
    echo "  $0                          # 构建latest版本"
    echo "  $0 v1.0.0                   # 构建v1.0.0版本"
    echo "  DOCKER_REGISTRY=registry.example.com $0 v1.0.0  # 构建并推送到私有仓库"
    echo
    echo "选项:"
    echo "  -h, --help    显示此帮助信息"
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