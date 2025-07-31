package performance

import (
	"context"
	"runtime"
	"sync"
	"time"
)

// ResourceMonitor 系统资源监控器
type ResourceMonitor struct {
	cpuSamples    []float64
	memorySamples []int64
	gcPauses      []time.Duration
	mutex         sync.RWMutex
	running       bool
}

// NewResourceMonitor 创建新的资源监控器
func NewResourceMonitor() *ResourceMonitor {
	return &ResourceMonitor{
		cpuSamples:    make([]float64, 0),
		memorySamples: make([]int64, 0),
		gcPauses:      make([]time.Duration, 0),
	}
}

// Start 开始监控系统资源
func (rm *ResourceMonitor) Start(ctx context.Context) {
	rm.mutex.Lock()
	rm.running = true
	rm.mutex.Unlock()

	ticker := time.NewTicker(100 * time.Millisecond) // 每100ms采样一次
	defer ticker.Stop()

	var lastGCPause uint64

	for {
		select {
		case <-ctx.Done():
			rm.mutex.Lock()
			rm.running = false
			rm.mutex.Unlock()
			return
		case <-ticker.C:
			rm.collectSample(&lastGCPause)
		}
	}
}

// collectSample 收集一次资源使用样本
func (rm *ResourceMonitor) collectSample(lastGCPause *uint64) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	rm.mutex.Lock()
	defer rm.mutex.Unlock()

	// 收集内存使用情况
	rm.memorySamples = append(rm.memorySamples, int64(m.Alloc))

	// 收集GC暂停时间
	if m.PauseTotalNs > *lastGCPause {
		gcPause := time.Duration(m.PauseTotalNs - *lastGCPause)
		rm.gcPauses = append(rm.gcPauses, gcPause)
		*lastGCPause = m.PauseTotalNs
	}

	// 简化的CPU使用率计算（基于goroutine数量）
	cpuUsage := float64(runtime.NumGoroutine()) / 1000.0
	if cpuUsage > 1.0 {
		cpuUsage = 1.0
	}
	rm.cpuSamples = append(rm.cpuSamples, cpuUsage)
}

// GetReport 获取资源使用报告
func (rm *ResourceMonitor) GetReport() *ResourceReport {
	rm.mutex.RLock()
	defer rm.mutex.RUnlock()

	report := &ResourceReport{
		GoroutineCount: runtime.NumGoroutine(),
	}

	// 计算CPU使用统计
	if len(rm.cpuSamples) > 0 {
		var totalCPU float64
		maxCPU := rm.cpuSamples[0]

		for _, cpu := range rm.cpuSamples {
			totalCPU += cpu
			if cpu > maxCPU {
				maxCPU = cpu
			}
		}

		report.AvgCPUUsage = totalCPU / float64(len(rm.cpuSamples))
		report.PeakCPUUsage = maxCPU
	}

	// 计算内存使用统计
	if len(rm.memorySamples) > 0 {
		var totalMemory int64
		maxMemory := rm.memorySamples[0]

		for _, memory := range rm.memorySamples {
			totalMemory += memory
			if memory > maxMemory {
				maxMemory = memory
			}
		}

		report.AvgMemoryUsage = totalMemory / int64(len(rm.memorySamples))
		report.PeakMemoryUsage = maxMemory
	}

	// 计算GC暂停时间
	if len(rm.gcPauses) > 0 {
		var totalGCPause time.Duration
		for _, pause := range rm.gcPauses {
			totalGCPause += pause
		}
		report.GCPauseTime = totalGCPause / time.Duration(len(rm.gcPauses))
	}

	return report
}

// Reset 重置监控数据
func (rm *ResourceMonitor) Reset() {
	rm.mutex.Lock()
	defer rm.mutex.Unlock()

	rm.cpuSamples = rm.cpuSamples[:0]
	rm.memorySamples = rm.memorySamples[:0]
	rm.gcPauses = rm.gcPauses[:0]
}

// IsRunning 检查监控器是否正在运行
func (rm *ResourceMonitor) IsRunning() bool {
	rm.mutex.RLock()
	defer rm.mutex.RUnlock()
	return rm.running
}

// runHighLoadWorker 运行高负载Worker用于测试
func runHighLoadWorker(ctx context.Context, db *gorm.DB, redisClient *redis.Client, taskCount, workerID int) {
	for i := 0; i < taskCount; i++ {
		select {
		case <-ctx.Done():
			return
		default:
			// 模拟CPU密集型任务
			performCPUIntensiveTask()

			// 模拟内存分配
			performMemoryIntensiveTask()

			// 模拟数据库操作
			performDatabaseOperation(db, workerID, i)

			// 模拟Redis操作
			performRedisOperation(ctx, redisClient, workerID, i)

			// 短暂休息避免过度消耗资源
			time.Sleep(time.Millisecond)
		}
	}
}

// performCPUIntensiveTask 执行CPU密集型任务
func performCPUIntensiveTask() {
	// 简单的CPU密集型计算
	sum := 0
	for i := 0; i < 10000; i++ {
		sum += i * i
	}
	_ = sum
}

// performMemoryIntensiveTask 执行内存密集型任务
func performMemoryIntensiveTask() {
	// 分配和释放内存
	data := make([]byte, 1024*10) // 10KB
	for i := range data {
		data[i] = byte(i % 256)
	}
	_ = data
}

// performRedisOperation 执行Redis操作
func performRedisOperation(ctx context.Context, redisClient *redis.Client, workerID, opID int) {
	key := fmt.Sprintf("perf:worker:%d:op:%d", workerID, opID)
	value := fmt.Sprintf("data-%d-%d", workerID, opID)

	// 设置值
	redisClient.Set(ctx, key, value, time.Minute)

	// 获取值
	redisClient.Get(ctx, key)

	// 删除值
	redisClient.Del(ctx, key)
}

// BenchmarkResult 基准测试结果
type BenchmarkResult struct {
	Name           string        `json:"name"`
	Operations     int           `json:"operations"`
	Duration       time.Duration `json:"duration"`
	OpsPerSecond   float64       `json:"ops_per_second"`
	AvgLatency     time.Duration `json:"avg_latency"`
	MinLatency     time.Duration `json:"min_latency"`
	MaxLatency     time.Duration `json:"max_latency"`
	P50Latency     time.Duration `json:"p50_latency"`
	P95Latency     time.Duration `json:"p95_latency"`
	P99Latency     time.Duration `json:"p99_latency"`
	MemoryUsage    int64         `json:"memory_usage"`
	AllocatedBytes int64         `json:"allocated_bytes"`
	ErrorRate      float64       `json:"error_rate"`
}

// Benchmark 基准测试框架
type Benchmark struct {
	name      string
	operation func() error
	setup     func() error
	teardown  func() error
}

// NewBenchmark 创建新的基准测试
func NewBenchmark(name string, operation func() error) *Benchmark {
	return &Benchmark{
		name:      name,
		operation: operation,
	}
}

// WithSetup 设置初始化函数
func (b *Benchmark) WithSetup(setup func() error) *Benchmark {
	b.setup = setup
	return b
}

// WithTeardown 设置清理函数
func (b *Benchmark) WithTeardown(teardown func() error) *Benchmark {
	b.teardown = teardown
	return b
}

// Run 运行基准测试
func (b *Benchmark) Run(operations int, concurrency int) (*BenchmarkResult, error) {
	// 执行初始化
	if b.setup != nil {
		if err := b.setup(); err != nil {
			return nil, fmt.Errorf("setup failed: %w", err)
		}
	}

	// 执行清理
	defer func() {
		if b.teardown != nil {
			b.teardown()
		}
	}()

	// 记录初始内存状态
	var startMemStats, endMemStats runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&startMemStats)

	// 准备结果收集
	latencies := make([]time.Duration, 0, operations)
	var latencyMutex sync.Mutex
	var errorCount int64

	// 开始基准测试
	startTime := time.Now()

	// 并发执行操作
	var wg sync.WaitGroup
	opsPerWorker := operations / concurrency
	if opsPerWorker == 0 {
		opsPerWorker = 1
		concurrency = operations
	}

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			for j := 0; j < opsPerWorker; j++ {
				opStart := time.Now()
				err := b.operation()
				opDuration := time.Since(opStart)

				latencyMutex.Lock()
				latencies = append(latencies, opDuration)
				latencyMutex.Unlock()

				if err != nil {
					atomic.AddInt64(&errorCount, 1)
				}
			}
		}()
	}

	wg.Wait()
	totalDuration := time.Since(startTime)

	// 记录结束内存状态
	runtime.GC()
	runtime.ReadMemStats(&endMemStats)

	// 计算结果
	result := &BenchmarkResult{
		Name:           b.name,
		Operations:     len(latencies),
		Duration:       totalDuration,
		OpsPerSecond:   float64(len(latencies)) / totalDuration.Seconds(),
		MemoryUsage:    int64(endMemStats.Alloc),
		AllocatedBytes: int64(endMemStats.TotalAlloc - startMemStats.TotalAlloc),
		ErrorRate:      float64(errorCount) / float64(len(latencies)),
	}

	// 计算延迟统计
	if len(latencies) > 0 {
		result.AvgLatency = calculateAverageLatency(latencies)
		result.MinLatency = calculateMinLatency(latencies)
		result.MaxLatency = calculateMaxLatency(latencies)
		result.P50Latency = calculatePercentileLatency(latencies, 0.50)
		result.P95Latency = calculatePercentileLatency(latencies, 0.95)
		result.P99Latency = calculatePercentileLatency(latencies, 0.99)
	}

	return result, nil
}

// calculateMinLatency 计算最小延迟
func calculateMinLatency(latencies []time.Duration) time.Duration {
	if len(latencies) == 0 {
		return 0
	}

	min := latencies[0]
	for _, latency := range latencies {
		if latency < min {
			min = latency
		}
	}

	return min
}

// LoadTestConfig 负载测试配置
type LoadTestConfig struct {
	Duration    time.Duration `json:"duration"`
	Concurrency int           `json:"concurrency"`
	RampUpTime  time.Duration `json:"ramp_up_time"`
	RampDownTime time.Duration `json:"ramp_down_time"`
	TargetRPS   int           `json:"target_rps"`
}

// LoadTestResult 负载测试结果
type LoadTestResult struct {
	Config          LoadTestConfig `json:"config"`
	TotalRequests   int64          `json:"total_requests"`
	SuccessRequests int64          `json:"success_requests"`
	FailedRequests  int64          `json:"failed_requests"`
	AvgRPS          float64        `json:"avg_rps"`
	PeakRPS         float64        `json:"peak_rps"`
	AvgLatency      time.Duration  `json:"avg_latency"`
	P95Latency      time.Duration  `json:"p95_latency"`
	P99Latency      time.Duration  `json:"p99_latency"`
	ErrorRate       float64        `json:"error_rate"`
	ResourceUsage   ResourceReport `json:"resource_usage"`
}

// LoadTester 负载测试器
type LoadTester struct {
	config    LoadTestConfig
	operation func() error
	monitor   *ResourceMonitor
}

// NewLoadTester 创建负载测试器
func NewLoadTester(config LoadTestConfig, operation func() error) *LoadTester {
	return &LoadTester{
		config:    config,
		operation: operation,
		monitor:   NewResourceMonitor(),
	}
}

// Run 运行负载测试
func (lt *LoadTester) Run(ctx context.Context) (*LoadTestResult, error) {
	// 启动资源监控
	monitorCtx, cancelMonitor := context.WithCancel(ctx)
	go lt.monitor.Start(monitorCtx)
	defer cancelMonitor()

	// 准备统计变量
	var (
		totalRequests   int64
		successRequests int64
		failedRequests  int64
		latencies       []time.Duration
		latencyMutex    sync.Mutex
		rpsHistory      []float64
		rpsHistoryMutex sync.Mutex
	)

	// 创建测试上下文
	testCtx, cancelTest := context.WithTimeout(ctx, lt.config.Duration)
	defer cancelTest()

	startTime := time.Now()

	// RPS统计goroutine
	go func() {
		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()

		var lastRequests int64
		for {
			select {
			case <-testCtx.Done():
				return
			case <-ticker.C:
				currentRequests := atomic.LoadInt64(&totalRequests)
				rps := float64(currentRequests - lastRequests)
				lastRequests = currentRequests

				rpsHistoryMutex.Lock()
				rpsHistory = append(rpsHistory, rps)
				rpsHistoryMutex.Unlock()
			}
		}
	}()

	// 计算每个阶段的并发数
	rampUpSteps := int(lt.config.RampUpTime.Seconds())
	if rampUpSteps == 0 {
		rampUpSteps = 1
	}

	// 启动负载测试
	var wg sync.WaitGroup

	// Ramp-up阶段
	for step := 1; step <= rampUpSteps; step++ {
		select {
		case <-testCtx.Done():
			break
		default:
		}

		currentConcurrency := (lt.config.Concurrency * step) / rampUpSteps
		if currentConcurrency == 0 {
			currentConcurrency = 1
		}

		for i := 0; i < currentConcurrency; i++ {
			wg.Add(1)
			go lt.worker(testCtx, &wg, &totalRequests, &successRequests, &failedRequests, &latencies, &latencyMutex)
		}

		time.Sleep(lt.config.RampUpTime / time.Duration(rampUpSteps))
	}

	// 等待测试完成
	wg.Wait()

	totalDuration := time.Since(startTime)

	// 计算结果
	result := &LoadTestResult{
		Config:          lt.config,
		TotalRequests:   totalRequests,
		SuccessRequests: successRequests,
		FailedRequests:  failedRequests,
		AvgRPS:          float64(totalRequests) / totalDuration.Seconds(),
		ErrorRate:       float64(failedRequests) / float64(totalRequests),
		ResourceUsage:   *lt.monitor.GetReport(),
	}

	// 计算峰值RPS
	if len(rpsHistory) > 0 {
		peakRPS := rpsHistory[0]
		for _, rps := range rpsHistory {
			if rps > peakRPS {
				peakRPS = rps
			}
		}
		result.PeakRPS = peakRPS
	}

	// 计算延迟统计
	if len(latencies) > 0 {
		result.AvgLatency = calculateAverageLatency(latencies)
		result.P95Latency = calculatePercentileLatency(latencies, 0.95)
		result.P99Latency = calculatePercentileLatency(latencies, 0.99)
	}

	return result, nil
}

// worker 负载测试工作协程
func (lt *LoadTester) worker(ctx context.Context, wg *sync.WaitGroup, totalRequests, successRequests, failedRequests *int64, latencies *[]time.Duration, latencyMutex *sync.Mutex) {
	defer wg.Done()

	for {
		select {
		case <-ctx.Done():
			return
		default:
			startTime := time.Now()
			err := lt.operation()
			latency := time.Since(startTime)

			atomic.AddInt64(totalRequests, 1)

			latencyMutex.Lock()
			*latencies = append(*latencies, latency)
			latencyMutex.Unlock()

			if err != nil {
				atomic.AddInt64(failedRequests, 1)
			} else {
				atomic.AddInt64(successRequests, 1)
			}

			// 控制请求速率
			if lt.config.TargetRPS > 0 {
				expectedInterval := time.Second / time.Duration(lt.config.TargetRPS)
				if latency < expectedInterval {
					time.Sleep(expectedInterval - latency)
				}
			}
		}
	}
}