#!/bin/bash

# 配置同步器构建脚本
# 用于构建配置同步器的Docker镜像

set -e

# 脚本配置
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
IMAGE_NAME="alertagent/config-syncer"
TAG="${TAG:-latest}"
DOCKERFILE_PATH="$PROJECT_ROOT/build/config-syncer/Dockerfile"

# 颜色输出
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
    log_info "Checking dependencies..."
    
    if ! command -v docker &> /dev/null; then
        log_error "Docker is not installed or not in PATH"
        exit 1
    fi
    
    if ! command -v go &> /dev/null; then
        log_error "Go is not installed or not in PATH"
        exit 1
    fi
    
    log_success "All dependencies are available"
}

# 检查项目结构
check_project_structure() {
    log_info "Checking project structure..."
    
    if [[ ! -f "$PROJECT_ROOT/go.mod" ]]; then
        log_error "go.mod not found in project root: $PROJECT_ROOT"
        exit 1
    fi
    
    if [[ ! -f "$DOCKERFILE_PATH" ]]; then
        log_error "Dockerfile not found: $DOCKERFILE_PATH"
        exit 1
    fi
    
    if [[ ! -f "$PROJECT_ROOT/cmd/config-syncer/main.go" ]]; then
        log_error "main.go not found: $PROJECT_ROOT/cmd/config-syncer/main.go"
        exit 1
    fi
    
    log_success "Project structure is valid"
}

# 运行测试
run_tests() {
    log_info "Running tests..."
    
    cd "$PROJECT_ROOT"
    
    # 运行单元测试
    if go test ./pkg/config-syncer/... -v; then
        log_success "Tests passed"
    else
        log_warning "Some tests failed, but continuing with build"
    fi
}

# 构建二进制文件
build_binary() {
    log_info "Building binary..."
    
    cd "$PROJECT_ROOT"
    
    # 设置构建变量
    VERSION=$(git describe --tags --always --dirty 2>/dev/null || echo "dev")
    BUILD_TIME=$(date -u '+%Y-%m-%d_%H:%M:%S')
    GIT_COMMIT=$(git rev-parse HEAD 2>/dev/null || echo "unknown")
    
    # 构建二进制文件
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
        -ldflags="-w -s -X main.Version=$VERSION -X main.BuildTime=$BUILD_TIME -X main.GitCommit=$GIT_COMMIT" \
        -a -installsuffix cgo \
        -o "$PROJECT_ROOT/bin/config-syncer" \
        "$PROJECT_ROOT/cmd/config-syncer"
    
    if [[ $? -eq 0 ]]; then
        log_success "Binary built successfully: $PROJECT_ROOT/bin/config-syncer"
    else
        log_error "Failed to build binary"
        exit 1
    fi
}

# 构建Docker镜像
build_docker_image() {
    log_info "Building Docker image: $IMAGE_NAME:$TAG"
    
    cd "$PROJECT_ROOT"
    
    # 构建镜像
    docker build \
        --file "$DOCKERFILE_PATH" \
        --tag "$IMAGE_NAME:$TAG" \
        --build-arg VERSION="$(git describe --tags --always --dirty 2>/dev/null || echo 'dev')" \
        --build-arg BUILD_TIME="$(date -u '+%Y-%m-%d_%H:%M:%S')" \
        --build-arg GIT_COMMIT="$(git rev-parse HEAD 2>/dev/null || echo 'unknown')" \
        .
    
    if [[ $? -eq 0 ]]; then
        log_success "Docker image built successfully: $IMAGE_NAME:$TAG"
    else
        log_error "Failed to build Docker image"
        exit 1
    fi
}

# 验证镜像
validate_image() {
    log_info "Validating Docker image..."
    
    # 检查镜像是否存在
    if docker images "$IMAGE_NAME:$TAG" --format "table {{.Repository}}:{{.Tag}}" | grep -q "$IMAGE_NAME:$TAG"; then
        log_success "Image exists: $IMAGE_NAME:$TAG"
    else
        log_error "Image not found: $IMAGE_NAME:$TAG"
        exit 1
    fi
    
    # 获取镜像信息
    IMAGE_SIZE=$(docker images "$IMAGE_NAME:$TAG" --format "{{.Size}}")
    log_info "Image size: $IMAGE_SIZE"
    
    # 运行基本健康检查
    log_info "Running basic health check..."
    CONTAINER_ID=$(docker run -d --rm \
        -e ALERTAGENT_ENDPOINT="http://test" \
        -e CLUSTER_ID="test-cluster" \
        -e CONFIG_TYPE="test" \
        -e CONFIG_PATH="/tmp/test.yml" \
        -e RELOAD_URL="http://test/reload" \
        -e SYNC_INTERVAL="60s" \
        "$IMAGE_NAME:$TAG")
    
    sleep 5
    
    if docker ps | grep -q "$CONTAINER_ID"; then
        log_success "Container is running"
        docker stop "$CONTAINER_ID" > /dev/null
    else
        log_error "Container failed to start"
        docker logs "$CONTAINER_ID" 2>/dev/null || true
        exit 1
    fi
}

# 推送镜像（可选）
push_image() {
    if [[ "$PUSH" == "true" ]]; then
        log_info "Pushing image to registry..."
        
        docker push "$IMAGE_NAME:$TAG"
        
        if [[ $? -eq 0 ]]; then
            log_success "Image pushed successfully: $IMAGE_NAME:$TAG"
        else
            log_error "Failed to push image"
            exit 1
        fi
    fi
}

# 清理
cleanup() {
    log_info "Cleaning up..."
    
    # 删除构建的二进制文件
    if [[ -f "$PROJECT_ROOT/bin/config-syncer" ]]; then
        rm "$PROJECT_ROOT/bin/config-syncer"
        log_info "Removed binary file"
    fi
}

# 显示帮助信息
show_help() {
    cat << EOF
配置同步器构建脚本

用法: $0 [选项]

选项:
  -t, --tag TAG        设置镜像标签 (默认: latest)
  -p, --push          构建后推送镜像到注册表
  --skip-tests        跳过测试
  --skip-validation   跳过镜像验证
  -h, --help          显示此帮助信息

环境变量:
  TAG                 镜像标签 (默认: latest)
  PUSH                是否推送镜像 (true/false)
  SKIP_TESTS          是否跳过测试 (true/false)
  SKIP_VALIDATION     是否跳过验证 (true/false)

示例:
  $0                           # 构建默认镜像
  $0 -t v1.0.0                # 构建指定标签的镜像
  $0 -t v1.0.0 --push         # 构建并推送镜像
  $0 --skip-tests              # 跳过测试直接构建

EOF
}

# 主函数
main() {
    # 解析命令行参数
    while [[ $# -gt 0 ]]; do
        case $1 in
            -t|--tag)
                TAG="$2"
                shift 2
                ;;
            -p|--push)
                PUSH="true"
                shift
                ;;
            --skip-tests)
                SKIP_TESTS="true"
                shift
                ;;
            --skip-validation)
                SKIP_VALIDATION="true"
                shift
                ;;
            -h|--help)
                show_help
                exit 0
                ;;
            *)
                log_error "Unknown option: $1"
                show_help
                exit 1
                ;;
        esac
    done
    
    log_info "Starting build process for config-syncer"
    log_info "Image: $IMAGE_NAME:$TAG"
    
    # 执行构建步骤
    check_dependencies
    check_project_structure
    
    if [[ "$SKIP_TESTS" != "true" ]]; then
        run_tests
    else
        log_warning "Skipping tests"
    fi
    
    build_binary
    build_docker_image
    
    if [[ "$SKIP_VALIDATION" != "true" ]]; then
        validate_image
    else
        log_warning "Skipping image validation"
    fi
    
    push_image
    cleanup
    
    log_success "Build completed successfully!"
    log_info "Image: $IMAGE_NAME:$TAG"
}

# 错误处理
trap 'log_error "Build failed at line $LINENO"' ERR

# 运行主函数
main "$@"