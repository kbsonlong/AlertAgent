#!/bin/bash

# AlertAgent 质量门禁检查脚本
# 用于确保代码质量和性能标准

set -e

# 配置变量
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
QUALITY_REPORT_DIR="$PROJECT_ROOT/quality-reports"

# 质量标准配置
MIN_COVERAGE=80
MAX_CYCLOMATIC_COMPLEXITY=10
MAX_FUNCTION_LENGTH=50
MAX_ERROR_RATE=0.01
MIN_PERFORMANCE_SCORE=80

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

# 质量门禁结果
QUALITY_GATE_PASSED=true

# 创建报告目录
mkdir -p "$QUALITY_REPORT_DIR"

# 代码覆盖率检查
check_code_coverage() {
    log_info "检查代码覆盖率..."
    
    cd "$PROJECT_ROOT"
    
    # 运行测试并生成覆盖率报告
    go test -v -race -coverprofile="$QUALITY_REPORT_DIR/coverage.out" -covermode=atomic ./internal/... > "$QUALITY_REPORT_DIR/coverage.log" 2>&1
    
    if [ $? -ne 0 ]; then
        log_error "测试执行失败"
        QUALITY_GATE_PASSED=false
        return 1
    fi
    
    # 计算覆盖率
    coverage_result=$(go tool cover -func="$QUALITY_REPORT_DIR/coverage.out" | tail -1)
    coverage_percentage=$(echo "$coverage_result" | awk '{print $3}' | sed 's/%//')
    
    log_info "代码覆盖率: ${coverage_percentage}%"
    
    # 检查覆盖率是否达标
    if (( $(echo "$coverage_percentage >= $MIN_COVERAGE" | bc -l) )); then
        log_success "代码覆盖率检查通过 (${coverage_percentage}% >= ${MIN_COVERAGE}%)"
    else
        log_error "代码覆盖率不达标 (${coverage_percentage}% < ${MIN_COVERAGE}%)"
        QUALITY_GATE_PASSED=false
    fi
    
    # 生成HTML覆盖率报告
    go tool cover -html="$QUALITY_REPORT_DIR/coverage.out" -o "$QUALITY_REPORT_DIR/coverage.html"
}

# 代码复杂度检查
check_code_complexity() {
    log_info "检查代码复杂度..."
    
    cd "$PROJECT_ROOT"
    
    # 使用gocyclo检查圈复杂度
    if command -v gocyclo &> /dev/null; then
        gocyclo -over $MAX_CYCLOMATIC_COMPLEXITY ./internal/... > "$QUALITY_REPORT_DIR/complexity.txt" 2>&1
        
        if [ -s "$QUALITY_REPORT_DIR/complexity.txt" ]; then
            log_error "发现高复杂度函数:"
            cat "$QUALITY_REPORT_DIR/complexity.txt"
            QUALITY_GATE_PASSED=false
        else
            log_success "代码复杂度检查通过"
        fi
    else
        log_warning "gocyclo未安装，跳过复杂度检查"
    fi
}

# 代码质量检查
check_code_quality() {
    log_info "检查代码质量..."
    
    cd "$PROJECT_ROOT"
    
    # 运行go vet
    log_info "运行go vet..."
    if ! go vet ./... > "$QUALITY_REPORT_DIR/vet.log" 2>&1; then
        log_error "go vet检查失败:"
        cat "$QUALITY_REPORT_DIR/vet.log"
        QUALITY_GATE_PASSED=false
    else
        log_success "go vet检查通过"
    fi
    
    # 运行golangci-lint
    if command -v golangci-lint &> /dev/null; then
        log_info "运行golangci-lint..."
        if ! golangci-lint run --timeout=5m > "$QUALITY_REPORT_DIR/lint.log" 2>&1; then
            log_error "golangci-lint检查失败:"
            cat "$QUALITY_REPORT_DIR/lint.log"
            QUALITY_GATE_PASSED=false
        else
            log_success "golangci-lint检查通过"
        fi
    else
        log_warning "golangci-lint未安装，跳过lint检查"
    fi
}

# 安全检查
check_security() {
    log_info "检查安全问题..."
    
    cd "$PROJECT_ROOT"
    
    # 运行gosec
    if command -v gosec &> /dev/null; then
        log_info "运行gosec安全扫描..."
        if ! gosec -fmt json -out "$QUALITY_REPORT_DIR/security.json" ./... > "$QUALITY_REPORT_DIR/security.log" 2>&1; then
            log_error "发现安全问题:"
            cat "$QUALITY_REPORT_DIR/security.log"
            QUALITY_GATE_PASSED=false
        else
            log_success "安全检查通过"
        fi
    else
        log_warning "gosec未安装，跳过安全检查"
    fi
    
    # 检查依赖漏洞
    if command -v nancy &> /dev/null; then
        log_info "检查依赖漏洞..."
        if ! go list -json -m all | nancy sleuth > "$QUALITY_REPORT_DIR/vulnerabilities.log" 2>&1; then
            log_error "发现依赖漏洞:"
            cat "$QUALITY_REPORT_DIR/vulnerabilities.log"
            QUALITY_GATE_PASSED=false
        else
            log_success "依赖漏洞检查通过"
        fi
    else
        log_warning "nancy未安装，跳过依赖漏洞检查"
    fi
}

# 性能检查
check_performance() {
    log_info "检查性能指标..."
    
    cd "$PROJECT_ROOT"
    
    # 运行基准测试
    log_info "运行基准测试..."
    if go test -bench=. -benchmem ./... > "$QUALITY_REPORT_DIR/benchmark.log" 2>&1; then
        log_success "基准测试完成"
        
        # 分析基准测试结果
        analyze_benchmark_results
    else
        log_error "基准测试失败"
        QUALITY_GATE_PASSED=false
    fi
    
    # 运行性能测试 (如果存在)
    if [ -d "tests/performance" ]; then
        log_info "运行性能测试..."
        if timeout 300 go test -v -tags=performance ./tests/performance/... > "$QUALITY_REPORT_DIR/performance.log" 2>&1; then
            log_success "性能测试完成"
            
            # 分析性能测试结果
            analyze_performance_results
        else
            log_warning "性能测试超时或失败"
        fi
    fi
}

# 分析基准测试结果
analyze_benchmark_results() {
    log_info "分析基准测试结果..."
    
    # 提取关键性能指标
    if grep -q "ns/op" "$QUALITY_REPORT_DIR/benchmark.log"; then
        log_success "基准测试结果可用"
        
        # 检查是否有性能回归
        check_performance_regression
    else
        log_warning "未找到基准测试结果"
    fi
}

# 分析性能测试结果
analyze_performance_results() {
    log_info "分析性能测试结果..."
    
    # 检查错误率
    error_rate=$(grep -o "错误率.*%" "$QUALITY_REPORT_DIR/performance.log" | head -1 | grep -o "[0-9.]*" || echo "0")
    
    if (( $(echo "$error_rate <= $MAX_ERROR_RATE * 100" | bc -l) )); then
        log_success "性能测试错误率检查通过 (${error_rate}%)"
    else
        log_error "性能测试错误率过高 (${error_rate}%)"
        QUALITY_GATE_PASSED=false
    fi
}

# 检查性能回归
check_performance_regression() {
    log_info "检查性能回归..."
    
    # 如果存在历史基准数据，进行对比
    if [ -f "$QUALITY_REPORT_DIR/benchmark_baseline.log" ]; then
        log_info "对比历史基准数据..."
        
        # 简单的性能回归检查
        current_avg=$(grep "ns/op" "$QUALITY_REPORT_DIR/benchmark.log" | awk '{sum+=$3; count++} END {print sum/count}' || echo "0")
        baseline_avg=$(grep "ns/op" "$QUALITY_REPORT_DIR/benchmark_baseline.log" | awk '{sum+=$3; count++} END {print sum/count}' || echo "0")
        
        if (( $(echo "$current_avg > $baseline_avg * 1.2" | bc -l) )); then
            log_error "检测到性能回归 (当前: ${current_avg}ns/op, 基准: ${baseline_avg}ns/op)"
            QUALITY_GATE_PASSED=false
        else
            log_success "未检测到性能回归"
        fi
    else
        log_info "未找到历史基准数据，保存当前结果作为基准"
        cp "$QUALITY_REPORT_DIR/benchmark.log" "$QUALITY_REPORT_DIR/benchmark_baseline.log"
    fi
}

# 检查技术债务
check_technical_debt() {
    log_info "检查技术债务..."
    
    cd "$PROJECT_ROOT"
    
    # 检查TODO和FIXME
    todo_count=$(find . -name "*.go" -not -path "./vendor/*" -exec grep -n "TODO\|FIXME\|HACK" {} \; | wc -l)
    
    log_info "发现 $todo_count 个技术债务标记"
    
    if [ "$todo_count" -gt 50 ]; then
        log_warning "技术债务标记过多 ($todo_count > 50)"
    else
        log_success "技术债务标记在可接受范围内"
    fi
    
    # 检查代码重复
    if command -v dupl &> /dev/null; then
        log_info "检查代码重复..."
        dupl -threshold 50 ./internal/... > "$QUALITY_REPORT_DIR/duplication.log" 2>&1
        
        if [ -s "$QUALITY_REPORT_DIR/duplication.log" ]; then
            log_warning "发现代码重复:"
            head -20 "$QUALITY_REPORT_DIR/duplication.log"
        else
            log_success "未发现显著代码重复"
        fi
    else
        log_warning "dupl未安装，跳过代码重复检查"
    fi
}

# 检查文档质量
check_documentation() {
    log_info "检查文档质量..."
    
    cd "$PROJECT_ROOT"
    
    # 检查README文件
    if [ -f "README.md" ]; then
        log_success "README.md存在"
    else
        log_error "缺少README.md文件"
        QUALITY_GATE_PASSED=false
    fi
    
    # 检查API文档
    if [ -f "docs/api.md" ]; then
        log_success "API文档存在"
    else
        log_warning "建议添加API文档"
    fi
    
    # 检查代码注释覆盖率
    total_functions=$(grep -r "^func " ./internal/ | wc -l)
    documented_functions=$(grep -B1 -r "^func " ./internal/ | grep "^//" | wc -l)
    
    if [ "$total_functions" -gt 0 ]; then
        doc_coverage=$((documented_functions * 100 / total_functions))
        log_info "函数文档覆盖率: ${doc_coverage}%"
        
        if [ "$doc_coverage" -lt 60 ]; then
            log_warning "函数文档覆盖率较低 (${doc_coverage}% < 60%)"
        else
            log_success "函数文档覆盖率良好"
        fi
    fi
}

# 检查依赖管理
check_dependencies() {
    log_info "检查依赖管理..."
    
    cd "$PROJECT_ROOT"
    
    # 检查go.mod和go.sum一致性
    if go mod verify > "$QUALITY_REPORT_DIR/mod_verify.log" 2>&1; then
        log_success "依赖验证通过"
    else
        log_error "依赖验证失败:"
        cat "$QUALITY_REPORT_DIR/mod_verify.log"
        QUALITY_GATE_PASSED=false
    fi
    
    # 检查未使用的依赖
    if command -v go-mod-outdated &> /dev/null; then
        log_info "检查过期依赖..."
        go list -u -m all | go-mod-outdated -update -direct > "$QUALITY_REPORT_DIR/outdated.log" 2>&1
        
        if [ -s "$QUALITY_REPORT_DIR/outdated.log" ]; then
            log_warning "发现过期依赖:"
            head -10 "$QUALITY_REPORT_DIR/outdated.log"
        else
            log_success "依赖都是最新的"
        fi
    fi
}

# 生成质量报告
generate_quality_report() {
    log_info "生成质量报告..."
    
    local report_file="$QUALITY_REPORT_DIR/quality_report.md"
    
    cat > "$report_file" << EOF
# AlertAgent 代码质量报告

生成时间: $(date)

## 质量门禁结果

EOF
    
    if [ "$QUALITY_GATE_PASSED" = true ]; then
        echo "✅ **质量门禁通过**" >> "$report_file"
    else
        echo "❌ **质量门禁失败**" >> "$report_file"
    fi
    
    echo "" >> "$report_file"
    
    # 添加各项检查结果
    echo "## 检查项目" >> "$report_file"
    echo "" >> "$report_file"
    
    # 代码覆盖率
    if [ -f "$QUALITY_REPORT_DIR/coverage.out" ]; then
        coverage_result=$(go tool cover -func="$QUALITY_REPORT_DIR/coverage.out" | tail -1)
        echo "### 代码覆盖率" >> "$report_file"
        echo "\`\`\`" >> "$report_file"
        echo "$coverage_result" >> "$report_file"
        echo "\`\`\`" >> "$report_file"
        echo "" >> "$report_file"
    fi
    
    # 安全检查结果
    if [ -f "$QUALITY_REPORT_DIR/security.json" ]; then
        echo "### 安全检查" >> "$report_file"
        security_issues=$(jq '.Issues | length' "$QUALITY_REPORT_DIR/security.json" 2>/dev/null || echo "0")
        echo "发现安全问题: $security_issues 个" >> "$report_file"
        echo "" >> "$report_file"
    fi
    
    # 性能测试结果
    if [ -f "$QUALITY_REPORT_DIR/performance.log" ]; then
        echo "### 性能测试" >> "$report_file"
        echo "详细结果请查看 performance.log" >> "$report_file"
        echo "" >> "$report_file"
    fi
    
    echo "## 建议" >> "$report_file"
    echo "" >> "$report_file"
    
    if [ "$QUALITY_GATE_PASSED" = true ]; then
        echo "- 代码质量良好，继续保持" >> "$report_file"
        echo "- 建议定期更新依赖" >> "$report_file"
        echo "- 持续关注性能指标" >> "$report_file"
    else
        echo "- 请修复质量门禁失败的问题" >> "$report_file"
        echo "- 提高代码覆盖率" >> "$report_file"
        echo "- 修复安全问题" >> "$report_file"
        echo "- 优化性能瓶颈" >> "$report_file"
    fi
    
    log_success "质量报告已生成: $report_file"
}

# 主函数
main() {
    log_info "开始运行质量门禁检查..."
    
    # 检查必要工具
    if ! command -v go &> /dev/null; then
        log_error "Go未安装"
        exit 1
    fi
    
    if ! command -v bc &> /dev/null; then
        log_error "bc计算器未安装"
        exit 1
    fi
    
    # 运行各项检查
    check_code_coverage
    check_code_quality
    check_code_complexity
    check_security
    check_performance
    check_technical_debt
    check_documentation
    check_dependencies
    
    # 生成质量报告
    generate_quality_report
    
    # 输出最终结果
    if [ "$QUALITY_GATE_PASSED" = true ]; then
        log_success "质量门禁检查通过"
        exit 0
    else
        log_error "质量门禁检查失败"
        exit 1
    fi
}

# 运行主函数
main "$@"