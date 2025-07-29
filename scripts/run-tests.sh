#!/bin/bash

# AlertAgent 测试运行脚本
# 用于运行各种类型的测试：单元测试、集成测试、性能测试、兼容性测试

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

# 显示帮助信息
show_help() {
    cat << EOF
AlertAgent 测试运行脚本

用法: $0 [选项] [测试类型]

测试类型:
    unit            运行单元测试
    integration     运行集成测试
    performance     运行性能测试
    compatibility   运行兼容性测试
    load           运行负载测试
    all            运行所有测试
    coverage       运行测试覆盖率分析

选项:
    -h, --help     显示此帮助信息
    -v, --verbose  详细输出
    -c, --clean    清理测试环境
    -p, --parallel 并行运行测试
    --timeout      设置测试超时时间（默认：10m）
    --race         启用竞态检测
    --bench        运行基准测试

示例:
    $0 unit                    # 运行单元测试
    $0 integration -v          # 详细运行集成测试
    $0 all --race             # 运行所有测试并启用竞态检测
    $0 performance --bench     # 运行性能测试和基准测试
    $0 coverage               # 生成测试覆盖率报告

EOF
}

# 默认参数
VERBOSE=false
CLEAN=false
PARALLEL=false
TIMEOUT="10m"
RACE=false
BENCH=false
TEST_TYPE=""

# 解析命令行参数
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
        -c|--clean)
            CLEAN=true
            shift
            ;;
        -p|--parallel)
            PARALLEL=true
            shift
            ;;
        --timeout)
            TIMEOUT="$2"
            shift 2
            ;;
        --race)
            RACE=true
            shift
            ;;
        --bench)
            BENCH=true
            shift
            ;;
        unit|integration|performance|compatibility|load|all|coverage)
            TEST_TYPE="$1"
            shift
            ;;
        *)
            log_error "未知参数: $1"
            show_help
            exit 1
            ;;
    esac
done

# 如果没有指定测试类型，显示帮助
if [[ -z "$TEST_TYPE" ]]; then
    log_error "请指定测试类型"
    show_help
    exit 1
fi

# 构建测试参数
TEST_ARGS=""
if [[ "$VERBOSE" == "true" ]]; then
    TEST_ARGS="$TEST_ARGS -v"
fi

if [[ "$PARALLEL" == "true" ]]; then
    TEST_ARGS="$TEST_ARGS -parallel 4"
fi

if [[ "$RACE" == "true" ]]; then
    TEST_ARGS="$TEST_ARGS -race"
fi

TEST_ARGS="$TEST_ARGS -timeout $TIMEOUT"

# 清理函数
cleanup() {
    log_info "清理测试环境..."
    
    # 停止可能运行的测试服务
    pkill -f "go test" || true
    
    # 清理测试数据库
    if [[ -f "test.db" ]]; then
        rm -f test.db
        log_info "已删除测试数据库"
    fi
    
    # 清理测试日志
    if [[ -d "test-logs" ]]; then
        rm -rf test-logs
        log_info "已删除测试日志目录"
    fi
    
    # 清理覆盖率文件
    find . -name "*.out" -type f -delete 2>/dev/null || true
    find . -name "coverage.html" -type f -delete 2>/dev/null || true
    
    log_success "测试环境清理完成"
}

# 如果指定了清理选项，执行清理并退出
if [[ "$CLEAN" == "true" ]]; then
    cleanup
    exit 0
fi

# 检查Go环境
if ! command -v go &> /dev/null; then
    log_error "Go 未安装或不在 PATH 中"
    exit 1
fi

log_info "Go 版本: $(go version)"

# 检查项目根目录
if [[ ! -f "go.mod" ]]; then
    log_error "请在项目根目录运行此脚本"
    exit 1
fi

# 创建测试日志目录
mkdir -p test-logs

# 运行单元测试
run_unit_tests() {
    log_info "运行单元测试..."
    
    local unit_args="$TEST_ARGS"
    if [[ "$BENCH" == "true" ]]; then
        unit_args="$unit_args -bench=."
    fi
    
    # 排除集成测试、性能测试和兼容性测试目录
    local test_packages=$(go list ./... | grep -v "/test/integration" | grep -v "/test/performance" | grep -v "/test/compatibility")
    
    if go test $unit_args $test_packages 2>&1 | tee test-logs/unit-tests.log; then
        log_success "单元测试通过"
        return 0
    else
        log_error "单元测试失败"
        return 1
    fi
}

# 运行集成测试
run_integration_tests() {
    log_info "运行集成测试..."
    
    # 设置集成测试环境变量
    export INTEGRATION_TEST=1
    
    local integration_args="$TEST_ARGS"
    if [[ "$BENCH" == "true" ]]; then
        integration_args="$integration_args -bench=."
    fi
    
    if go test $integration_args ./test/integration/... 2>&1 | tee test-logs/integration-tests.log; then
        log_success "集成测试通过"
        return 0
    else
        log_error "集成测试失败"
        return 1
    fi
}

# 运行性能测试
run_performance_tests() {
    log_info "运行性能测试..."
    
    # 设置性能测试环境变量
    export LOAD_TEST=1
    
    local perf_args="$TEST_ARGS -bench=."
    
    if go test $perf_args ./test/performance/... 2>&1 | tee test-logs/performance-tests.log; then
        log_success "性能测试通过"
        return 0
    else
        log_error "性能测试失败"
        return 1
    fi
}

# 运行兼容性测试
run_compatibility_tests() {
    log_info "运行兼容性测试..."
    
    # 设置兼容性测试环境变量
    export COMPATIBILITY_TEST=1
    
    local compat_args="$TEST_ARGS"
    if [[ "$BENCH" == "true" ]]; then
        compat_args="$compat_args -bench=."
    fi
    
    if go test $compat_args ./test/compatibility/... 2>&1 | tee test-logs/compatibility-tests.log; then
        log_success "兼容性测试通过"
        return 0
    else
        log_error "兼容性测试失败"
        return 1
    fi
}

# 运行负载测试
run_load_tests() {
    log_info "运行负载测试..."
    
    # 设置负载测试环境变量
    export LOAD_TEST=1
    
    local load_args="$TEST_ARGS -timeout 5m"
    
    if go test $load_args -run "TestHealthCheckLoad|TestAnalysisAPILoad|TestChannelAPILoad" ./test/performance/... 2>&1 | tee test-logs/load-tests.log; then
        log_success "负载测试通过"
        return 0
    else
        log_error "负载测试失败"
        return 1
    fi
}

# 运行测试覆盖率分析
run_coverage_tests() {
    log_info "运行测试覆盖率分析..."
    
    # 排除测试目录
    local test_packages=$(go list ./... | grep -v "/test/")
    
    # 生成覆盖率报告
    if go test -coverprofile=coverage.out -covermode=atomic $test_packages 2>&1 | tee test-logs/coverage-tests.log; then
        # 生成HTML报告
        go tool cover -html=coverage.out -o coverage.html
        
        # 显示覆盖率统计
        local coverage_percent=$(go tool cover -func=coverage.out | grep total | awk '{print $3}')
        log_success "测试覆盖率: $coverage_percent"
        log_info "HTML覆盖率报告已生成: coverage.html"
        
        # 检查覆盖率阈值
        local coverage_num=$(echo $coverage_percent | sed 's/%//')
        if (( $(echo "$coverage_num >= 80" | bc -l) )); then
            log_success "覆盖率达到要求 (>= 80%)"
        else
            log_warning "覆盖率低于要求 (< 80%)"
        fi
        
        return 0
    else
        log_error "覆盖率测试失败"
        return 1
    fi
}

# 运行所有测试
run_all_tests() {
    log_info "运行所有测试..."
    
    local failed_tests=()
    
    # 运行单元测试
    if ! run_unit_tests; then
        failed_tests+=("单元测试")
    fi
    
    # 运行集成测试
    if ! run_integration_tests; then
        failed_tests+=("集成测试")
    fi
    
    # 运行性能测试
    if ! run_performance_tests; then
        failed_tests+=("性能测试")
    fi
    
    # 运行兼容性测试
    if ! run_compatibility_tests; then
        failed_tests+=("兼容性测试")
    fi
    
    # 检查结果
    if [[ ${#failed_tests[@]} -eq 0 ]]; then
        log_success "所有测试通过！"
        return 0
    else
        log_error "以下测试失败: ${failed_tests[*]}"
        return 1
    fi
}

# 主函数
main() {
    log_info "开始运行 $TEST_TYPE 测试..."
    log_info "测试参数: $TEST_ARGS"
    
    local start_time=$(date +%s)
    local exit_code=0
    
    # 根据测试类型运行相应测试
    case $TEST_TYPE in
        unit)
            run_unit_tests || exit_code=$?
            ;;
        integration)
            run_integration_tests || exit_code=$?
            ;;
        performance)
            run_performance_tests || exit_code=$?
            ;;
        compatibility)
            run_compatibility_tests || exit_code=$?
            ;;
        load)
            run_load_tests || exit_code=$?
            ;;
        coverage)
            run_coverage_tests || exit_code=$?
            ;;
        all)
            run_all_tests || exit_code=$?
            ;;
    esac
    
    local end_time=$(date +%s)
    local duration=$((end_time - start_time))
    
    log_info "测试完成，耗时: ${duration}秒"
    log_info "测试日志保存在: test-logs/"
    
    if [[ $exit_code -eq 0 ]]; then
        log_success "测试执行成功！"
    else
        log_error "测试执行失败！"
    fi
    
    exit $exit_code
}

# 设置信号处理
trap cleanup EXIT INT TERM

# 运行主函数
main "$@"