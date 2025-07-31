package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sort"
	"time"
)

// PerformanceResult 性能测试结果结构
type PerformanceResult struct {
	Name           string        `json:"name"`
	Operations     int           `json:"operations"`
	Duration       time.Duration `json:"duration"`
	OpsPerSecond   float64       `json:"ops_per_second"`
	AvgLatency     time.Duration `json:"avg_latency"`
	P95Latency     time.Duration `json:"p95_latency"`
	P99Latency     time.Duration `json:"p99_latency"`
	ErrorRate      float64       `json:"error_rate"`
	MemoryUsage    int64         `json:"memory_usage"`
	AllocatedBytes int64         `json:"allocated_bytes"`
}

// LoadTestResult 负载测试结果结构
type LoadTestResult struct {
	TestName        string        `json:"test_name"`
	Duration        time.Duration `json:"duration"`
	Concurrency     int           `json:"concurrency"`
	TotalRequests   int64         `json:"total_requests"`
	SuccessRequests int64         `json:"success_requests"`
	FailedRequests  int64         `json:"failed_requests"`
	AvgRPS          float64       `json:"avg_rps"`
	PeakRPS         float64       `json:"peak_rps"`
	AvgLatency      time.Duration `json:"avg_latency"`
	P95Latency      time.Duration `json:"p95_latency"`
	P99Latency      time.Duration `json:"p99_latency"`
	ErrorRate       float64       `json:"error_rate"`
}

// ResourceUsage 资源使用情况
type ResourceUsage struct {
	PeakCPUUsage    float64 `json:"peak_cpu_usage"`
	PeakMemoryUsage int64   `json:"peak_memory_usage"`
	AvgCPUUsage     float64 `json:"avg_cpu_usage"`
	AvgMemoryUsage  int64   `json:"avg_memory_usage"`
	GoroutineCount  int     `json:"goroutine_count"`
}

// PerformanceReport 性能报告
type PerformanceReport struct {
	GeneratedAt      time.Time           `json:"generated_at"`
	TestEnvironment  string              `json:"test_environment"`
	GoVersion        string              `json:"go_version"`
	BenchmarkResults []PerformanceResult `json:"benchmark_results"`
	LoadTestResults  []LoadTestResult    `json:"load_test_results"`
	ResourceUsage    ResourceUsage       `json:"resource_usage"`
	Summary          PerformanceSummary  `json:"summary"`
	Recommendations  []string            `json:"recommendations"`
}

// PerformanceSummary 性能摘要
type PerformanceSummary struct {
	TotalTests      int     `json:"total_tests"`
	PassedTests     int     `json:"passed_tests"`
	FailedTests     int     `json:"failed_tests"`
	AverageRPS      float64 `json:"average_rps"`
	AverageLatency  string  `json:"average_latency"`
	PeakMemoryUsage string  `json:"peak_memory_usage"`
	OverallScore    string  `json:"overall_score"`
}

func main() {
	// 读取性能测试结果文件
	resultsDir := getResultsDirectory()

	report := &PerformanceReport{
		GeneratedAt:     time.Now(),
		TestEnvironment: getTestEnvironment(),
		GoVersion:       getGoVersion(),
	}

	// 读取基准测试结果
	benchmarkResults, err := readBenchmarkResults(resultsDir)
	if err != nil {
		log.Printf("Warning: Failed to read benchmark results: %v", err)
	} else {
		report.BenchmarkResults = benchmarkResults
	}

	// 读取负载测试结果
	loadTestResults, err := readLoadTestResults(resultsDir)
	if err != nil {
		log.Printf("Warning: Failed to read load test results: %v", err)
	} else {
		report.LoadTestResults = loadTestResults
	}

	// 读取资源使用情况
	resourceUsage, err := readResourceUsage(resultsDir)
	if err != nil {
		log.Printf("Warning: Failed to read resource usage: %v", err)
	} else {
		report.ResourceUsage = resourceUsage
	}

	// 生成摘要
	report.Summary = generateSummary(report)

	// 生成建议
	report.Recommendations = generateRecommendations(report)

	// 输出Markdown格式的报告
	generateMarkdownReport(report)
}

func getResultsDirectory() string {
	if dir := os.Getenv("TEST_RESULTS_DIR"); dir != "" {
		return dir
	}
	return "test-results"
}

func getTestEnvironment() string {
	if env := os.Getenv("TEST_ENVIRONMENT"); env != "" {
		return env
	}
	if os.Getenv("CI") == "true" {
		return "CI"
	}
	return "Local"
}

func getGoVersion() string {
	if version := os.Getenv("GO_VERSION"); version != "" {
		return version
	}
	return "Unknown"
}

func readBenchmarkResults(resultsDir string) ([]PerformanceResult, error) {
	var results []PerformanceResult

	// 查找所有基准测试结果文件
	pattern := filepath.Join(resultsDir, "benchmark_*.json")
	files, err := filepath.Glob(pattern)
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		data, err := ioutil.ReadFile(file)
		if err != nil {
			log.Printf("Warning: Failed to read %s: %v", file, err)
			continue
		}

		var result PerformanceResult
		if err := json.Unmarshal(data, &result); err != nil {
			log.Printf("Warning: Failed to parse %s: %v", file, err)
			continue
		}

		results = append(results, result)
	}

	// 按名称排序
	sort.Slice(results, func(i, j int) bool {
		return results[i].Name < results[j].Name
	})

	return results, nil
}

func readLoadTestResults(resultsDir string) ([]LoadTestResult, error) {
	var results []LoadTestResult

	// 查找所有负载测试结果文件
	pattern := filepath.Join(resultsDir, "loadtest_*.json")
	files, err := filepath.Glob(pattern)
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		data, err := ioutil.ReadFile(file)
		if err != nil {
			log.Printf("Warning: Failed to read %s: %v", file, err)
			continue
		}

		var result LoadTestResult
		if err := json.Unmarshal(data, &result); err != nil {
			log.Printf("Warning: Failed to parse %s: %v", file, err)
			continue
		}

		results = append(results, result)
	}

	return results, nil
}

func readResourceUsage(resultsDir string) (ResourceUsage, error) {
	var usage ResourceUsage

	file := filepath.Join(resultsDir, "resource_usage.json")
	data, err := ioutil.ReadFile(file)
	if err != nil {
		return usage, err
	}

	err = json.Unmarshal(data, &usage)
	return usage, err
}

func generateSummary(report *PerformanceReport) PerformanceSummary {
	summary := PerformanceSummary{
		TotalTests: len(report.BenchmarkResults) + len(report.LoadTestResults),
	}

	// 计算通过/失败的测试
	for _, result := range report.BenchmarkResults {
		if result.ErrorRate < 0.01 { // 错误率小于1%认为通过
			summary.PassedTests++
		} else {
			summary.FailedTests++
		}
	}

	for _, result := range report.LoadTestResults {
		if result.ErrorRate < 0.01 {
			summary.PassedTests++
		} else {
			summary.FailedTests++
		}
	}

	// 计算平均RPS
	var totalRPS float64
	rpsCount := 0

	for _, result := range report.BenchmarkResults {
		totalRPS += result.OpsPerSecond
		rpsCount++
	}

	for _, result := range report.LoadTestResults {
		totalRPS += result.AvgRPS
		rpsCount++
	}

	if rpsCount > 0 {
		summary.AverageRPS = totalRPS / float64(rpsCount)
	}

	// 计算平均延迟
	var totalLatency time.Duration
	latencyCount := 0

	for _, result := range report.BenchmarkResults {
		totalLatency += result.AvgLatency
		latencyCount++
	}

	for _, result := range report.LoadTestResults {
		totalLatency += result.AvgLatency
		latencyCount++
	}

	if latencyCount > 0 {
		avgLatency := totalLatency / time.Duration(latencyCount)
		summary.AverageLatency = avgLatency.String()
	}

	// 格式化内存使用
	summary.PeakMemoryUsage = formatBytes(report.ResourceUsage.PeakMemoryUsage)

	// 计算总体评分
	summary.OverallScore = calculateOverallScore(report)

	return summary
}

func generateRecommendations(report *PerformanceReport) []string {
	var recommendations []string

	// 基于性能结果生成建议
	for _, result := range report.BenchmarkResults {
		if result.ErrorRate > 0.05 {
			recommendations = append(recommendations,
				fmt.Sprintf("测试 '%s' 错误率较高 (%.2f%%)，建议检查错误处理逻辑",
					result.Name, result.ErrorRate*100))
		}

		if result.OpsPerSecond < 100 {
			recommendations = append(recommendations,
				fmt.Sprintf("测试 '%s' TPS较低 (%.2f)，建议优化性能",
					result.Name, result.OpsPerSecond))
		}

		if result.P95Latency > time.Second {
			recommendations = append(recommendations,
				fmt.Sprintf("测试 '%s' P95延迟较高 (%v)，建议优化响应时间",
					result.Name, result.P95Latency))
		}
	}

	// 基于资源使用生成建议
	if report.ResourceUsage.PeakMemoryUsage > 1024*1024*1024 { // 1GB
		recommendations = append(recommendations,
			"峰值内存使用超过1GB，建议优化内存使用")
	}

	if report.ResourceUsage.PeakCPUUsage > 0.8 {
		recommendations = append(recommendations,
			"峰值CPU使用率超过80%，建议优化CPU密集型操作")
	}

	if len(recommendations) == 0 {
		recommendations = append(recommendations, "性能表现良好，无特殊建议")
	}

	return recommendations
}

func calculateOverallScore(report *PerformanceReport) string {
	if report.Summary.TotalTests == 0 {
		return "N/A"
	}

	passRate := float64(report.Summary.PassedTests) / float64(report.Summary.TotalTests)

	switch {
	case passRate >= 0.95:
		return "优秀"
	case passRate >= 0.85:
		return "良好"
	case passRate >= 0.70:
		return "一般"
	default:
		return "需要改进"
	}
}

func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

func generateMarkdownReport(report *PerformanceReport) {
	fmt.Printf("# AlertAgent 性能测试报告\n\n")

	// 基本信息
	fmt.Printf("**生成时间:** %s\n", report.GeneratedAt.Format("2006-01-02 15:04:05"))
	fmt.Printf("**测试环境:** %s\n", report.TestEnvironment)
	fmt.Printf("**Go版本:** %s\n\n", report.GoVersion)

	// 测试摘要
	fmt.Printf("## 测试摘要\n\n")
	fmt.Printf("| 指标 | 值 |\n")
	fmt.Printf("|------|----|\n")
	fmt.Printf("| 总测试数 | %d |\n", report.Summary.TotalTests)
	fmt.Printf("| 通过测试 | %d |\n", report.Summary.PassedTests)
	fmt.Printf("| 失败测试 | %d |\n", report.Summary.FailedTests)
	fmt.Printf("| 平均RPS | %.2f |\n", report.Summary.AverageRPS)
	fmt.Printf("| 平均延迟 | %s |\n", report.Summary.AverageLatency)
	fmt.Printf("| 峰值内存使用 | %s |\n", report.Summary.PeakMemoryUsage)
	fmt.Printf("| 总体评分 | %s |\n\n", report.Summary.OverallScore)

	// 基准测试结果
	if len(report.BenchmarkResults) > 0 {
		fmt.Printf("## 基准测试结果\n\n")
		fmt.Printf("| 测试名称 | 操作数 | 持续时间 | TPS | 平均延迟 | P95延迟 | P99延迟 | 错误率 | 内存使用 |\n")
		fmt.Printf("|----------|--------|----------|-----|----------|---------|---------|--------|----------|\n")

		for _, result := range report.BenchmarkResults {
			fmt.Printf("| %s | %d | %v | %.2f | %v | %v | %v | %.2f%% | %s |\n",
				result.Name,
				result.Operations,
				result.Duration,
				result.OpsPerSecond,
				result.AvgLatency,
				result.P95Latency,
				result.P99Latency,
				result.ErrorRate*100,
				formatBytes(result.MemoryUsage))
		}
		fmt.Printf("\n")
	}

	// 负载测试结果
	if len(report.LoadTestResults) > 0 {
		fmt.Printf("## 负载测试结果\n\n")
		fmt.Printf("| 测试名称 | 并发数 | 总请求数 | 成功请求 | 失败请求 | 平均RPS | 峰值RPS | 平均延迟 | P95延迟 | 错误率 |\n")
		fmt.Printf("|----------|--------|----------|----------|----------|---------|---------|----------|---------|--------|\n")

		for _, result := range report.LoadTestResults {
			fmt.Printf("| %s | %d | %d | %d | %d | %.2f | %.2f | %v | %v | %.2f%% |\n",
				result.TestName,
				result.Concurrency,
				result.TotalRequests,
				result.SuccessRequests,
				result.FailedRequests,
				result.AvgRPS,
				result.PeakRPS,
				result.AvgLatency,
				result.P95Latency,
				result.ErrorRate*100)
		}
		fmt.Printf("\n")
	}

	// 资源使用情况
	fmt.Printf("## 资源使用情况\n\n")
	fmt.Printf("| 指标 | 值 |\n")
	fmt.Printf("|------|----|\n")
	fmt.Printf("| 峰值CPU使用率 | %.2f%% |\n", report.ResourceUsage.PeakCPUUsage*100)
	fmt.Printf("| 平均CPU使用率 | %.2f%% |\n", report.ResourceUsage.AvgCPUUsage*100)
	fmt.Printf("| 峰值内存使用 | %s |\n", formatBytes(report.ResourceUsage.PeakMemoryUsage))
	fmt.Printf("| 平均内存使用 | %s |\n", formatBytes(report.ResourceUsage.AvgMemoryUsage))
	fmt.Printf("| Goroutine数量 | %d |\n\n", report.ResourceUsage.GoroutineCount)

	// 建议
	fmt.Printf("## 优化建议\n\n")
	for i, recommendation := range report.Recommendations {
		fmt.Printf("%d. %s\n", i+1, recommendation)
	}
	fmt.Printf("\n")

	// 趋势分析 (如果有历史数据)
	fmt.Printf("## 趋势分析\n\n")
	fmt.Printf("*注意: 趋势分析需要多次测试运行的历史数据*\n\n")

	// 结论
	fmt.Printf("## 结论\n\n")
	if report.Summary.PassedTests == report.Summary.TotalTests {
		fmt.Printf("✅ 所有性能测试通过，系统性能表现良好。\n\n")
	} else {
		fmt.Printf("⚠️ 部分性能测试未达到预期，建议根据上述建议进行优化。\n\n")
	}

	// 附录
	fmt.Printf("## 附录\n\n")
	fmt.Printf("### 测试环境配置\n\n")
	fmt.Printf("- 操作系统: %s\n", getOSInfo())
	fmt.Printf("- Go版本: %s\n", report.GoVersion)
	fmt.Printf("- 测试时间: %s\n", report.GeneratedAt.Format("2006-01-02 15:04:05"))
	fmt.Printf("- 测试环境: %s\n\n", report.TestEnvironment)

	fmt.Printf("### 测试数据说明\n\n")
	fmt.Printf("- **TPS**: 每秒事务数 (Transactions Per Second)\n")
	fmt.Printf("- **P95延迟**: 95%的请求延迟时间\n")
	fmt.Printf("- **P99延迟**: 99%的请求延迟时间\n")
	fmt.Printf("- **错误率**: 失败请求占总请求的百分比\n")
	fmt.Printf("- **内存使用**: 测试期间的内存分配情况\n\n")
}

func getOSInfo() string {
	if osInfo := os.Getenv("RUNNER_OS"); osInfo != "" {
		return osInfo
	}
	return "Unknown"
}
