#!/bin/bash

# AlertAgent 测试自动化脚本
# 用于本地开发和CI/CD环境中的测试执行

set -e

# 配置变量
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
TEST_RESULTS_DIR="$PROJECT_ROOT/test-results"
COVERAGE_DIR="$PROJECT_ROOT/coverage"

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

# 显示帮助信息
show_help() {
    cat << EOF
AlertAgent 测试自动化脚本

用法: $0 [选项] [测试类型]

测试类型:
    unit            运行单元测试
    integration     运行集成测试
    performance     运行性能测试
    frontend        运行前端测试
    all             运行所有测试 (默认)

选项:
    -h, --help      显示此帮助信息
    -v, --verbose   详细输出
    -c, --coverage  生成覆盖率报告
    -r, --report    生成测试报告
    -p, --parallel  并行运行测试
    --clean         清理测试环境
    --setup         设置测试环境
    --ci            CI模式运行

示例:
    $0 unit -c              # 运行单元测试并生成覆盖率报告
    $0 integration -v       # 详细模式运行集成测试
    $0 all -r -p           # 并行运行所有测试并生成报告
    $0 --setup             # 仅设置测试环境
    $0 --clean             # 仅清理测试环境

EOF
}

# 解析命令行参数
VERBOSE=false
COVERAGE=false
REPORT=false
PARALLEL=false
CLEAN_ONLY=false
SETUP_ONLY=false
CI_MODE=false
TEST_TYPE="all"

while [[ $# -gt 0 ]]; do
    case $1 in
        -h|--help)
            show_help
            exit 0
            ;;
        -v|--verbose)
            VERBOSE=true
            shift
            ;;
        -c|--coverage)
            COVERAGE=true
            shift
            ;;
        -r|--report)
            REPORT=true
            shift
            ;;
        -p|--parallel)
            PARALLEL=true
            shift
            ;;
        --clean)
            CLEAN_ONLY=true
            shift
            ;;
        --setup)
            SETUP_ONLY=true
            shift
            ;;
        --ci)
            CI_MODE=true
            shift
            ;;
        unit|integration|performance|frontend|all)
            TEST_TYPE=$1
            shift
            ;;
        *)
            log_error "未知选项: $1"
            show_help
            exit 1
            ;;
    esac
done

# 检查依赖
check_dependencies() {
    log_info "检查依赖..."
    
    # 检查Go
    if ! command -v go &> /dev/null; then
        log_error "Go未安装或不在PATH中"
        exit 1
    fi
    
    # 检查Node.js (如果需要前端测试)
    if [[ "$TEST_TYPE" == "frontend" || "$TEST_TYPE" == "all" ]]; then
        if ! command -v node &> /dev/null; then
            log_error "Node.js未安装或不在PATH中"
            exit 1
        fi
        if ! command -v npm &> /dev/null; then
            log_error "npm未安装或不在PATH中"
            exit 1
        fi
    fi
    
    # 检查Docker (如果需要集成测试)
    if [[ "$TEST_TYPE" == "integration" || "$TEST_TYPE" == "performance" || "$TEST_TYPE" == "all" ]]; then
        if ! command -v docker &> /dev/null; then
            log_warning "Docker未安装，将跳过需要Docker的测试"
        fi
    fi
    
    log_success "依赖检查完成"
}

# 设置测试环境
setup_test_environment() {
    log_info "设置测试环境..."
    
    # 创建测试结果目录
    mkdir -p "$TEST_RESULTS_DIR"
    mkdir -p "$COVERAGE_DIR"
    
    # 设置Go模块
    cd "$PROJECT_ROOT"
    go mod download
    
    # 启动测试服务 (MySQL, Redis)
    if [[ "$TEST_TYPE" == "integration" || "$TEST_TYPE" == "performance" || "$TEST_TYPE" == "all" ]]; then
        if command -v docker-compose &> /dev/null; then
            log_info "启动测试服务..."
            docker-compose -f docker-compose.test.yml up -d
            
            # 等待服务启动
            log_info "等待服务启动..."
            sleep 10
            
            # 检查服务状态
            if ! docker-compose -f docker-compose.test.yml ps | grep -q "Up"; then
                log_error "测试服务启动失败"
                exit 1
            fi
        else
            log_warning "docker-compose未安装，请手动启动MySQL和Redis服务"
        fi
    fi
    
    # 设置前端环境
    if [[ "$TEST_TYPE" == "frontend" || "$TEST_TYPE" == "all" ]]; then
        if [ -d "$PROJECT_ROOT/web" ]; then
            log_info "安装前端依赖..."
            cd "$PROJECT_ROOT/web"
            npm ci
        fi
    fi
    
    log_success "测试环境设置完成"
}

# 清理测试环境
cleanup_test_environment() {
    log_info "清理测试环境..."
    
    # 停止测试服务
    if [ -f "$PROJECT_ROOT/docker-compose.test.yml" ]; then
        cd "$PROJECT_ROOT"
        docker-compose -f docker-compose.test.yml down -v
    fi
    
    # 清理测试数据库
    if command -v mysql &> /dev/null; then
        mysql -h localhost -u root -ppassword -e "DROP DATABASE IF EXISTS alertagent_test;" 2>/dev/null || true
        mysql -h localhost -u root -ppassword -e "DROP DATABASE IF EXISTS alertagent_perf_test;" 2>/dev/null || true
    fi
    
    # 清理Redis测试数据
    if command -v redis-cli &> /dev/null; then
        redis-cli -h localhost -p 6379 -n 1 FLUSHDB 2>/dev/null || true
        redis-cli -h localhost -p 6379 -n 2 FLUSHDB 2>/dev/null || true
    fi
    
    log_success "测试环境清理完成"
}

# 运行单元测试
run_unit_tests() {
    log_info "运行单元测试..."
    
    cd "$PROJECT_ROOT"
    
    local test_args="-v"
    local coverage_args=""
    
    if [[ "$COVERAGE" == true ]]; then
        coverage_args="-coverprofile=$COVERAGE_DIR/unit_coverage.out -covermode=atomic"
    fi
    
    if [[ "$PARALLEL" == true ]]; then
        test_args="$test_args -parallel 4"
    fi
    
    if [[ "$VERBOSE" == true ]]; then
        test_args="$test_args -v"
    fi
    
    # 运行单元测试
    if go test $test_args $coverage_args -race ./internal/... > "$TEST_RESULTS_DIR/unit_test.log" 2>&1; then
        log_success "单元测试通过"
        
        # 生成覆盖率报告
        if [[ "$COVERAGE" == true ]]; then
            go tool cover -html="$COVERAGE_DIR/unit_coverage.out" -o "$COVERAGE_DIR/unit_coverage.html"
            go tool cover -func="$COVERAGE_DIR/unit_coverage.out" | tail -1
        fi
    else
        log_error "单元测试失败"
        if [[ "$VERBOSE" == true ]]; then
            cat "$TEST_RESULTS_DIR/unit_test.log"
        fi
        return 1
    fi
}

# 运行集成测试
run_integration_tests() {
    log_info "运行集成测试..."
    
    cd "$PROJECT_ROOT"
    
    # 设置环境变量
    export TEST_DB_HOST=localhost
    export TEST_DB_PORT=3306
    export TEST_DB_USER=root
    export TEST_DB_PASSWORD=password
    export TEST_DB_NAME=alertagent_test
    export TEST_REDIS_HOST=localhost
    export TEST_REDIS_PORT=6379
    
    local test_args="-v -tags=integration -timeout=30m"
    
    if [[ "$PARALLEL" == true ]]; then
        test_args="$test_args -parallel 2"
    fi
    
    # 运行集成测试
    if go test $test_args ./tests/integration/... > "$TEST_RESULTS_DIR/integration_test.log" 2>&1; then
        log_success "集成测试通过"
    else
        log_error "集成测试失败"
        if [[ "$VERBOSE" == true ]]; then
            cat "$TEST_RESULTS_DIR/integration_test.log"
        fi
        return 1
    fi
}

# 运行性能测试
run_performance_tests() {
    log_info "运行性能测试..."
    
    cd "$PROJECT_ROOT"
    
    # 设置环境变量
    export PERF_TEST_DB_HOST=localhost
    export PERF_TEST_DB_PORT=3306
    export PERF_TEST_DB_USER=root
    export PERF_TEST_DB_PASSWORD=password
    export PERF_TEST_DB_NAME=alertagent_perf_test
    export PERF_TEST_REDIS_HOST=localhost
    export PERF_TEST_REDIS_PORT=6379
    
    local test_args="-v -tags=performance -timeout=60m"
    
    # 运行性能测试
    if go test $test_args ./tests/performance/... > "$TEST_RESULTS_DIR/performance_test.log" 2>&1; then
        log_success "性能测试完成"
        
        # 生成性能报告
        if [[ "$REPORT" == true ]]; then
            go run scripts/generate_perf_report.go > "$TEST_RESULTS_DIR/performance_report.md"
        fi
    else
        log_error "性能测试失败"
        if [[ "$VERBOSE" == true ]]; then
            cat "$TEST_RESULTS_DIR/performance_test.log"
        fi
        return 1
    fi
}

# 运行前端测试
run_frontend_tests() {
    log_info "运行前端测试..."
    
    if [ ! -d "$PROJECT_ROOT/web" ]; then
        log_warning "前端目录不存在，跳过前端测试"
        return 0
    fi
    
    cd "$PROJECT_ROOT/web"
    
    # 运行linting
    log_info "运行前端代码检查..."
    if ! npm run lint > "$TEST_RESULTS_DIR/frontend_lint.log" 2>&1; then
        log_error "前端代码检查失败"
        if [[ "$VERBOSE" == true ]]; then
            cat "$TEST_RESULTS_DIR/frontend_lint.log"
        fi
        return 1
    fi
    
    # 运行类型检查
    log_info "运行前端类型检查..."
    if ! npm run type-check > "$TEST_RESULTS_DIR/frontend_typecheck.log" 2>&1; then
        log_error "前端类型检查失败"
        if [[ "$VERBOSE" == true ]]; then
            cat "$TEST_RESULTS_DIR/frontend_typecheck.log"
        fi
        return 1
    fi
    
    # 运行单元测试
    log_info "运行前端单元测试..."
    if ! npm run test:unit > "$TEST_RESULTS_DIR/frontend_unit.log" 2>&1; then
        log_error "前端单元测试失败"
        if [[ "$VERBOSE" == true ]]; then
            cat "$TEST_RESULTS_DIR/frontend_unit.log"
        fi
        return 1
    fi
    
    # 构建检查
    log_info "检查前端构建..."
    if ! npm run build > "$TEST_RESULTS_DIR/frontend_build.log" 2>&1; then
        log_error "前端构建失败"
        if [[ "$VERBOSE" == true ]]; then
            cat "$TEST_RESULTS_DIR/frontend_build.log"
        fi
        return 1
    fi
    
    log_success "前端测试通过"
}

# 生成测试报告
generate_test_report() {
    log_info "生成测试报告..."
    
    local report_file="$TEST_RESULTS_DIR/test_report.md"
    
    cat > "$report_file" << EOF
# AlertAgent 测试报告

生成时间: $(date)

## 测试概览

EOF
    
    # 单元测试结果
    if [ -f "$TEST_RESULTS_DIR/unit_test.log" ]; then
        echo "### 单元测试" >> "$report_file"
        if grep -q "PASS" "$TEST_RESULTS_DIR/unit_test.log"; then
            echo "✅ 通过" >> "$report_file"
        else
            echo "❌ 失败" >> "$report_file"
        fi
        echo "" >> "$report_file"
    fi
    
    # 集成测试结果
    if [ -f "$TEST_RESULTS_DIR/integration_test.log" ]; then
        echo "### 集成测试" >> "$report_file"
        if grep -q "PASS" "$TEST_RESULTS_DIR/integration_test.log"; then
            echo "✅ 通过" >> "$report_file"
        else
            echo "❌ 失败" >> "$report_file"
        fi
        echo "" >> "$report_file"
    fi
    
    # 性能测试结果
    if [ -f "$TEST_RESULTS_DIR/performance_report.md" ]; then
        echo "### 性能测试" >> "$report_file"
        cat "$TEST_RESULTS_DIR/performance_report.md" >> "$report_file"
        echo "" >> "$report_file"
    fi
    
    # 前端测试结果
    if [ -f "$TEST_RESULTS_DIR/frontend_unit.log" ]; then
        echo "### 前端测试" >> "$report_file"
        if grep -q "All tests passed" "$TEST_RESULTS_DIR/frontend_unit.log"; then
            echo "✅ 通过" >> "$report_file"
        else
            echo "❌ 失败" >> "$report_file"
        fi
        echo "" >> "$report_file"
    fi
    
    # 覆盖率信息
    if [ -f "$COVERAGE_DIR/unit_coverage.out" ]; then
        echo "### 代码覆盖率" >> "$report_file"
        go tool cover -func="$COVERAGE_DIR/unit_coverage.out" | tail -1 >> "$report_file"
        echo "" >> "$report_file"
    fi
    
    log_success "测试报告已生成: $report_file"
}

# 主函数
main() {
    log_info "开始运行AlertAgent测试自动化脚本"
    
    # 检查依赖
    check_dependencies
    
    # 仅清理模式
    if [[ "$CLEAN_ONLY" == true ]]; then
        cleanup_test_environment
        exit 0
    fi
    
    # 设置测试环境
    if [[ "$SETUP_ONLY" == true ]]; then
        setup_test_environment
        exit 0
    fi
    
    # 设置测试环境
    setup_test_environment
    
    # 运行测试
    local test_failed=false
    
    case "$TEST_TYPE" in
        unit)
            run_unit_tests || test_failed=true
            ;;
        integration)
            run_integration_tests || test_failed=true
            ;;
        performance)
            run_performance_tests || test_failed=true
            ;;
        frontend)
            run_frontend_tests || test_failed=true
            ;;
        all)
            run_unit_tests || test_failed=true
            run_integration_tests || test_failed=true
            run_frontend_tests || test_failed=true
            
            # 性能测试仅在非CI模式或明确要求时运行
            if [[ "$CI_MODE" != true ]] || [[ "$GITHUB_EVENT_NAME" == "schedule" ]]; then
                run_performance_tests || test_failed=true
            fi
            ;;
    esac
    
    # 生成测试报告
    if [[ "$REPORT" == true ]]; then
        generate_test_report
    fi
    
    # 清理测试环境 (除非是CI模式)
    if [[ "$CI_MODE" != true ]]; then
        cleanup_test_environment
    fi
    
    # 检查测试结果
    if [[ "$test_failed" == true ]]; then
        log_error "测试失败"
        exit 1
    else
        log_success "所有测试通过"
        exit 0
    fi
}

# 信号处理
trap cleanup_test_environment EXIT

# 运行主函数
main "$@"