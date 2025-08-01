#!/bin/bash

# AlertAgent 模型测试运行脚本
# 用于运行所有生成的测试脚本

set -e  # 遇到错误时退出

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 打印带颜色的消息
print_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 检查是否在正确的目录
check_directory() {
    if [ ! -d "../internal/model" ]; then
        print_error "请在 AlertAgent/test 目录下运行此脚本"
        exit 1
    fi
    
    if [ ! -f "go.mod" ] && [ ! -f "../go.mod" ]; then
        print_warning "未找到 go.mod 文件，可能需要在项目根目录运行"
    fi
}

# 检查Go环境
check_go_env() {
    if ! command -v go &> /dev/null; then
        print_error "Go 未安装或不在 PATH 中"
        exit 1
    fi
    
    print_info "Go 版本: $(go version)"
}

# 运行集成测试
run_integration_tests() {
    print_info "运行集成测试..."
    
    if [ -f "integration_test.go" ]; then
        echo "======================================"
        echo "集成测试结果:"
        echo "======================================"
        
        if go test -v integration_test.go; then
            print_success "集成测试通过"
        else
            print_error "集成测试失败"
            return 1
        fi
    else
        print_warning "未找到 integration_test.go 文件"
    fi
}

# 运行性能测试
run_benchmark_tests() {
    print_info "运行性能测试..."
    
    if [ -f "benchmark_test.go" ]; then
        echo "======================================"
        echo "性能测试结果:"
        echo "======================================"
        
        if go test -bench=. -benchmem benchmark_test.go; then
            print_success "性能测试完成"
        else
            print_error "性能测试失败"
            return 1
        fi
    else
        print_warning "未找到 benchmark_test.go 文件"
    fi
}

# 验证模拟数据生成器
validate_mock_generator() {
    print_info "验证模拟数据生成器..."
    
    if [ -f "mock_data_generator.go" ]; then
        echo "======================================"
        echo "模拟数据生成器验证:"
        echo "======================================"
        
        # 检查语法
        if go build -o /dev/null mock_data_generator.go; then
            print_success "模拟数据生成器语法正确"
        else
            print_error "模拟数据生成器语法错误"
            return 1
        fi
    else
        print_warning "未找到 mock_data_generator.go 文件"
    fi
}

# 运行所有测试
run_all_tests() {
    local start_time=$(date +%s)
    local failed_tests=0
    
    print_info "开始运行 AlertAgent 模型测试套件"
    print_info "测试目录: $(pwd)"
    print_info "开始时间: $(date)"
    
    echo ""
    echo "=========================================="
    echo "AlertAgent 数据模型测试套件"
    echo "=========================================="
    echo ""
    
    # 运行各项测试
    if ! run_integration_tests; then
        ((failed_tests++))
    fi
    
    echo ""
    
    if ! run_benchmark_tests; then
        ((failed_tests++))
    fi
    
    echo ""
    
    if ! validate_mock_generator; then
        ((failed_tests++))
    fi
    
    # 计算总耗时
    local end_time=$(date +%s)
    local duration=$((end_time - start_time))
    
    echo ""
    echo "=========================================="
    echo "测试总结"
    echo "=========================================="
    print_info "总耗时: ${duration} 秒"
    
    if [ $failed_tests -eq 0 ]; then
        print_success "所有测试通过! ✅"
        echo ""
        echo "生成的测试文件:"
        echo "  - integration_test.go    (集成测试)"
        echo "  - benchmark_test.go      (性能测试)"
        echo "  - mock_data_generator.go (数据生成器)"
        echo "  - README.md              (使用说明)"
        echo ""
        echo "使用方法:"
        echo "  go test -v integration_test.go"
        echo "  go test -bench=. benchmark_test.go"
        return 0
    else
        print_error "有 $failed_tests 项测试失败 ❌"
        return 1
    fi
}

# 显示帮助信息
show_help() {
    echo "AlertAgent 模型测试运行脚本"
    echo ""
    echo "用法: $0 [选项]"
    echo ""
    echo "选项:"
    echo "  -h, --help        显示此帮助信息"
    echo "  -i, --integration 仅运行集成测试"
    echo "  -b, --benchmark   仅运行性能测试"
    echo "  -m, --mock        仅验证模拟数据生成器"
    echo "  -a, --all         运行所有测试 (默认)"
    echo ""
    echo "示例:"
    echo "  $0                # 运行所有测试"
    echo "  $0 -i             # 仅运行集成测试"
    echo "  $0 -b             # 仅运行性能测试"
}

# 主函数
main() {
    # 解析命令行参数
    case "${1:-}" in
        -h|--help)
            show_help
            exit 0
            ;;
        -i|--integration)
            check_directory
            check_go_env
            run_integration_tests
            ;;
        -b|--benchmark)
            check_directory
            check_go_env
            run_benchmark_tests
            ;;
        -m|--mock)
            check_directory
            check_go_env
            validate_mock_generator
            ;;
        -a|--all|"")
            check_directory
            check_go_env
            run_all_tests
            ;;
        *)
            print_error "未知选项: $1"
            show_help
            exit 1
            ;;
    esac
}

# 运行主函数
main "$@"