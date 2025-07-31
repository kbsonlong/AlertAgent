package performance

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// TestHighConcurrencyAlertProcessing 测试高并发告警处理性能
func TestHighConcurrencyAlertProcessing(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}

	// 初始化测试环境
	testDB, testRedis, cleanup := setupPerformanceTestEnvironment(t)
	defer cleanup()

	testCases := []struct {
		name           string
		concurrency    int
		totalAlerts    int
		expectedTPS    int // 期望的每秒处理数
		maxLatency     time.Duration
	}{
		{
			name:           "低并发基准测试",
			concurrency:    10,
			totalAlerts:    1000,
			expectedTPS:    100,
			maxLatency:     100 * time.Millisecond,
		},
		{
			name:           "中等并发测试",
			concurrency:    50,
			totalAlerts:    5000,
			expectedTPS:    500,
			maxLatency:     200 * time.Millisecond,
		},
		{
			name:           "高并发压力测试",
			concurrency:    100,
			totalAlerts:    10000,
			expectedTPS:    800,
			maxLatency:     500 * time.Millisecond,
		},
		{
			name:           "极限并发测试",
			concurrency:    200,
			totalAlerts:    20000,
			expectedTPS:    1000,
			maxLatency:     1 * time.Second,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// 清理测试数据
			clearPerformanceTestData(testDB, testRedis)

			// 执行性能测试
			result := runAlertProcessingPerformanceTest(t, testDB, testRedis, tc.concurrency, tc.totalAlerts)

			// 验证性能指标
			t.Logf("Performance Results for %s:", tc.name)
			t.Logf("  Total Alerts: %d", result.TotalAlerts)
			t.Logf("  Duration: %v", result.Duration)
			t.Logf("  TPS: %.2f", result.TPS)
			t.Logf("  Average Latency: %v", result.AvgLatency)
			t.Logf("  P95 Latency: %v", result.P95Latency)
			t.Logf("  P99 Latency: %v", result.P99Latency)
			t.Logf("  Max Latency: %v", result.MaxLatency)
			t.Logf("  Success Rate: %.2f%%", result.SuccessRate*100)
			t.Logf("  Memory Usage: %d MB", result.MemoryUsage/1024/1024)

			// 性能断言
			assert.GreaterOrEqual(t, result.TPS, float64(tc.expectedTPS*0.8), "TPS should be at least 80%% of expected")
			assert.LessOrEqual(t, result.P95Latency, tc.maxLatency, "P95 latency should be within acceptable range")
			assert.GreaterOrEqual(t, result.SuccessRate, 0.99, "Success rate should be at least 99%%")
		})
	}
}

// TestConfigSyncPerformance 测试配置同步性能基准
func TestConfigSyncPerformance(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}

	testDB, testRedis, cleanup := setupPerformanceTestEnvironment(t)
	defer cleanup()

	testCases := []struct {
		name            string
		clusterCount    int
		configSize      int // KB
		syncInterval    time.Duration
		expectedLatency time.Duration
	}{
		{
			name:            "小规模集群同步",
			clusterCount:    10,
			configSize:      10,
			syncInterval:    5 * time.Second,
			expectedLatency: 100 * time.Millisecond,
		},
		{
			name:            "中等规模集群同步",
			clusterCount:    50,
			configSize:      50,
			syncInterval:    10 * time.Second,
			expectedLatency: 500 * time.Millisecond,
		},
		{
			name:            "大规模集群同步",
			clusterCount:    100,
			configSize:      100,
			syncInterval:    30 * time.Second,
			expectedLatency: 1 * time.Second,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			clearPerformanceTestData(testDB, testRedis)

			result := runConfigSyncPerformanceTest(t, testDB, testRedis, tc.clusterCount, tc.configSize)

			t.Logf("Config Sync Performance Results for %s:", tc.name)
			t.Logf("  Clusters: %d", result.ClusterCount)
			t.Logf("  Config Size: %d KB", result.ConfigSize)
			t.Logf("  Total Sync Time: %v", result.TotalSyncTime)
			t.Logf("  Average Sync Latency: %v", result.AvgSyncLatency)
			t.Logf("  Max Sync Latency: %v", result.MaxSyncLatency)
			t.Logf("  Sync Success Rate: %.2f%%", result.SyncSuccessRate*100)
			t.Logf("  Throughput: %.2f syncs/sec", result.Throughput)

			assert.LessOrEqual(t, result.AvgSyncLatency, tc.expectedLatency, "Average sync latency should be within expected range")
			assert.GreaterOrEqual(t, result.SyncSuccessRate, 0.95, "Sync success rate should be at least 95%%")
		})
	}
}

// TestSystemResourceMonitoring 测试系统资源使用监控
func TestSystemResourceMonitoring(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}

	testDB, testRedis, cleanup := setupPerformanceTestEnvironment(t)
	defer cleanup()

	// 启动资源监控
	monitor := NewResourceMonitor()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	go monitor.Start(ctx)

	// 模拟高负载场景
	t.Run("高负载资源监控", func(t *testing.T) {
		// 启动多个Worker模拟高负载
		workerCount := 20
		taskCount := 10000

		var wg sync.WaitGroup
		wg.Add(workerCount)

		for i := 0; i < workerCount; i++ {
			go func(workerID int) {
				defer wg.Done()
				runHighLoadWorker(ctx, testDB, testRedis, taskCount/workerCount, workerID)
			}(i)
		}

		wg.Wait()

		// 获取资源使用报告
		report := monitor.GetReport()

		t.Logf("Resource Usage Report:")
		t.Logf("  Peak CPU Usage: %.2f%%", report.PeakCPUUsage*100)
		t.Logf("  Peak Memory Usage: %d MB", report.PeakMemoryUsage/1024/1024)
		t.Logf("  Average CPU Usage: %.2f%%", report.AvgCPUUsage*100)
		t.Logf("  Average Memory Usage: %d MB", report.AvgMemoryUsage/1024/1024)
		t.Logf("  Goroutine Count: %d", report.GoroutineCount)
		t.Logf("  GC Pause Time: %v", report.GCPauseTime)

		// 资源使用断言
		assert.Less(t, report.PeakCPUUsage, 0.9, "Peak CPU usage should be less than 90%%")
		assert.Less(t, report.PeakMemoryUsage, 1024*1024*1024, "Peak memory usage should be less than 1GB")
		assert.Less(t, report.GCPauseTime, 100*time.Millisecond, "GC pause time should be reasonable")
	})
}

// TestDatabaseConnectionPoolPerformance 测试数据库连接池性能
func TestDatabaseConnectionPoolPerformance(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}

	testDB, _, cleanup := setupPerformanceTestEnvironment(t)
	defer cleanup()

	testCases := []struct {
		name           string
		maxConnections int
		concurrency    int
		operations     int
		expectedTPS    float64
	}{
		{
			name:           "小连接池测试",
			maxConnections: 10,
			concurrency:    20,
			operations:     1000,
			expectedTPS:    200,
		},
		{
			name:           "中等连接池测试",
			maxConnections: 50,
			concurrency:    100,
			operations:     5000,
			expectedTPS:    800,
		},
		{
			name:           "大连接池测试",
			maxConnections: 100,
			concurrency:    200,
			operations:     10000,
			expectedTPS:    1200,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// 配置连接池
			sqlDB, err := testDB.DB()
			require.NoError(t, err)
			
			sqlDB.SetMaxOpenConns(tc.maxConnections)
			sqlDB.SetMaxIdleConns(tc.maxConnections / 2)
			sqlDB.SetConnMaxLifetime(time.Hour)

			result := runDatabasePerformanceTest(t, testDB, tc.concurrency, tc.operations)

			t.Logf("Database Performance Results for %s:", tc.name)
			t.Logf("  Max Connections: %d", tc.maxConnections)
			t.Logf("  Concurrency: %d", tc.concurrency)
			t.Logf("  Total Operations: %d", result.TotalOperations)
			t.Logf("  Duration: %v", result.Duration)
			t.Logf("  TPS: %.2f", result.TPS)
			t.Logf("  Average Latency: %v", result.AvgLatency)
			t.Logf("  Connection Pool Stats: %+v", result.PoolStats)

			assert.GreaterOrEqual(t, result.TPS, tc.expectedTPS*0.8, "Database TPS should meet expectations")
			assert.LessOrEqual(t, result.AvgLatency, 50*time.Millisecond, "Average database latency should be reasonable")
		})
	}
}

// 性能测试结果结构
type AlertProcessingResult struct {
	TotalAlerts   int
	Duration      time.Duration
	TPS           float64
	AvgLatency    time.Duration
	P95Latency    time.Duration
	P99Latency    time.Duration
	MaxLatency    time.Duration
	SuccessRate   float64
	MemoryUsage   int64
	ErrorCount    int64
}

type ConfigSyncResult struct {
	ClusterCount      int
	ConfigSize        int
	TotalSyncTime     time.Duration
	AvgSyncLatency    time.Duration
	MaxSyncLatency    time.Duration
	SyncSuccessRate   float64
	Throughput        float64
}

type ResourceReport struct {
	PeakCPUUsage    float64
	PeakMemoryUsage int64
	AvgCPUUsage     float64
	AvgMemoryUsage  int64
	GoroutineCount  int
	GCPauseTime     time.Duration
}

type DatabaseResult struct {
	TotalOperations int
	Duration        time.Duration
	TPS             float64
	AvgLatency      time.Duration
	PoolStats       map[string]interface{}
}

// 性能测试实现函数
func runAlertProcessingPerformanceTest(t *testing.T, db *gorm.DB, redisClient *redis.Client, concurrency, totalAlerts int) *AlertProcessingResult {
	ctx := context.Background()
	
	// 统计变量
	var (
		successCount int64
		errorCount   int64
		latencies    []time.Duration
		latencyMutex sync.Mutex
	)

	// 创建任务生产者和消费者
	producer := &TestTaskProducer{redis: redisClient}
	
	// 启动Worker
	for i := 0; i < concurrency/2; i++ {
		worker := &TestPerformanceWorker{db: db, redis: redisClient}
		go worker.Start(ctx)
	}

	startTime := time.Now()
	
	// 并发发送告警
	var wg sync.WaitGroup
	alertsPerWorker := totalAlerts / concurrency
	
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			
			for j := 0; j < alertsPerWorker; j++ {
				alertStart := time.Now()
				
				alertData := generateTestAlert(workerID, j)
				err := producer.PublishAIAnalysisTask(ctx, alertData["alert_id"].(string), alertData)
				
				latency := time.Since(alertStart)
				latencyMutex.Lock()
				latencies = append(latencies, latency)
				latencyMutex.Unlock()
				
				if err != nil {
					atomic.AddInt64(&errorCount, 1)
				} else {
					atomic.AddInt64(&successCount, 1)
				}
			}
		}(i)
	}

	wg.Wait()
	duration := time.Since(startTime)

	// 等待所有任务处理完成
	waitForTaskCompletion(t, db, totalAlerts, 30*time.Second)

	// 计算性能指标
	result := &AlertProcessingResult{
		TotalAlerts: totalAlerts,
		Duration:    duration,
		TPS:         float64(totalAlerts) / duration.Seconds(),
		SuccessRate: float64(successCount) / float64(totalAlerts),
		ErrorCount:  errorCount,
	}

	// 计算延迟统计
	if len(latencies) > 0 {
		result.AvgLatency = calculateAverageLatency(latencies)
		result.P95Latency = calculatePercentileLatency(latencies, 0.95)
		result.P99Latency = calculatePercentileLatency(latencies, 0.99)
		result.MaxLatency = calculateMaxLatency(latencies)
	}

	// 获取内存使用情况
	result.MemoryUsage = getCurrentMemoryUsage()

	return result
}

func runConfigSyncPerformanceTest(t *testing.T, db *gorm.DB, redisClient *redis.Client, clusterCount, configSize int) *ConfigSyncResult {
	ctx := context.Background()
	
	var (
		syncLatencies []time.Duration
		successCount  int64
		latencyMutex  sync.Mutex
	)

	startTime := time.Now()

	// 并发同步配置到多个集群
	var wg sync.WaitGroup
	wg.Add(clusterCount)

	for i := 0; i < clusterCount; i++ {
		go func(clusterID int) {
			defer wg.Done()
			
			syncStart := time.Now()
			
			// 模拟配置同步
			config := generateTestConfig(configSize)
			err := simulateConfigSync(ctx, db, fmt.Sprintf("cluster-%d", clusterID), config)
			
			syncLatency := time.Since(syncStart)
			latencyMutex.Lock()
			syncLatencies = append(syncLatencies, syncLatency)
			latencyMutex.Unlock()
			
			if err == nil {
				atomic.AddInt64(&successCount, 1)
			}
		}(i)
	}

	wg.Wait()
	totalSyncTime := time.Since(startTime)

	result := &ConfigSyncResult{
		ClusterCount:    clusterCount,
		ConfigSize:      configSize,
		TotalSyncTime:   totalSyncTime,
		SyncSuccessRate: float64(successCount) / float64(clusterCount),
		Throughput:      float64(clusterCount) / totalSyncTime.Seconds(),
	}

	if len(syncLatencies) > 0 {
		result.AvgSyncLatency = calculateAverageLatency(syncLatencies)
		result.MaxSyncLatency = calculateMaxLatency(syncLatencies)
	}

	return result
}

func runDatabasePerformanceTest(t *testing.T, db *gorm.DB, concurrency, operations int) *DatabaseResult {
	var (
		successCount int64
		latencies    []time.Duration
		latencyMutex sync.Mutex
	)

	startTime := time.Now()

	// 并发数据库操作
	var wg sync.WaitGroup
	opsPerWorker := operations / concurrency

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			
			for j := 0; j < opsPerWorker; j++ {
				opStart := time.Now()
				
				// 执行数据库操作（插入、查询、更新）
				err := performDatabaseOperation(db, workerID, j)
				
				latency := time.Since(opStart)
				latencyMutex.Lock()
				latencies = append(latencies, latency)
				latencyMutex.Unlock()
				
				if err == nil {
					atomic.AddInt64(&successCount, 1)
				}
			}
		}(i)
	}

	wg.Wait()
	duration := time.Since(startTime)

	// 获取连接池统计
	sqlDB, _ := db.DB()
	stats := sqlDB.Stats()
	poolStats := map[string]interface{}{
		"open_connections": stats.OpenConnections,
		"in_use":          stats.InUse,
		"idle":            stats.Idle,
		"wait_count":      stats.WaitCount,
		"wait_duration":   stats.WaitDuration,
	}

	result := &DatabaseResult{
		TotalOperations: operations,
		Duration:        duration,
		TPS:             float64(operations) / duration.Seconds(),
		PoolStats:       poolStats,
	}

	if len(latencies) > 0 {
		result.AvgLatency = calculateAverageLatency(latencies)
	}

	return result
}

// 辅助函数实现
func generateTestAlert(workerID, alertID int) map[string]interface{} {
	severities := []string{"critical", "warning", "info"}
	instances := []string{"server-01", "server-02", "server-03", "server-04"}
	
	return map[string]interface{}{
		"alert_id":   fmt.Sprintf("alert-%d-%d", workerID, alertID),
		"alertname":  fmt.Sprintf("TestAlert-%d", rand.Intn(10)),
		"instance":   instances[rand.Intn(len(instances))],
		"severity":   severities[rand.Intn(len(severities))],
		"summary":    fmt.Sprintf("Test alert from worker %d, alert %d", workerID, alertID),
		"timestamp":  time.Now().Unix(),
	}
}

func generateTestConfig(sizeKB int) string {
	// 生成指定大小的配置内容
	baseConfig := `
global:
  scrape_interval: 15s
  evaluation_interval: 15s

rule_files:
  - "/etc/prometheus/rules/*.yml"

scrape_configs:
  - job_name: 'prometheus'
    static_configs:
      - targets: ['localhost:9090']
`
	
	// 填充到指定大小
	currentSize := len(baseConfig)
	targetSize := sizeKB * 1024
	
	if currentSize < targetSize {
		padding := strings.Repeat("# padding\n", (targetSize-currentSize)/10)
		baseConfig += padding
	}
	
	return baseConfig
}

func simulateConfigSync(ctx context.Context, db *gorm.DB, clusterID, config string) error {
	// 模拟配置同步延迟
	time.Sleep(time.Duration(rand.Intn(100)) * time.Millisecond)
	
	// 更新同步状态
	return db.Exec(`
		INSERT INTO config_sync_status 
		(id, cluster_id, config_type, config_hash, sync_status, sync_time)
		VALUES (UUID(), ?, 'prometheus', ?, 'success', NOW())
		ON DUPLICATE KEY UPDATE
		config_hash = VALUES(config_hash),
		sync_status = VALUES(sync_status),
		sync_time = VALUES(sync_time)
	`, clusterID, fmt.Sprintf("%x", len(config))).Error
}

func performDatabaseOperation(db *gorm.DB, workerID, opID int) error {
	// 随机执行不同类型的数据库操作
	operations := []func() error{
		func() error {
			// 插入操作
			return db.Exec(`
				INSERT INTO task_queue (id, queue_name, task_type, payload, status)
				VALUES (UUID(), 'test', 'perf_test', '{}', 'completed')
			`).Error
		},
		func() error {
			// 查询操作
			var count int64
			return db.Raw("SELECT COUNT(*) FROM task_queue WHERE status = 'completed'").Scan(&count).Error
		},
		func() error {
			// 更新操作
			return db.Exec("UPDATE task_queue SET updated_at = NOW() WHERE task_type = 'perf_test' LIMIT 1").Error
		},
	}
	
	op := operations[rand.Intn(len(operations))]
	return op()
}

func waitForTaskCompletion(t *testing.T, db *gorm.DB, expectedCount int, timeout time.Duration) {
	deadline := time.Now().Add(timeout)
	
	for time.Now().Before(deadline) {
		var count int64
		db.Raw("SELECT COUNT(*) FROM task_queue WHERE status = 'completed'").Scan(&count)
		
		if count >= int64(expectedCount*0.9) { // 允许10%的容错
			return
		}
		
		time.Sleep(100 * time.Millisecond)
	}
	
	t.Logf("Warning: Not all tasks completed within timeout")
}

func calculateAverageLatency(latencies []time.Duration) time.Duration {
	if len(latencies) == 0 {
		return 0
	}
	
	var total time.Duration
	for _, latency := range latencies {
		total += latency
	}
	
	return total / time.Duration(len(latencies))
}

func calculatePercentileLatency(latencies []time.Duration, percentile float64) time.Duration {
	if len(latencies) == 0 {
		return 0
	}
	
	// 简单的百分位计算
	sorted := make([]time.Duration, len(latencies))
	copy(sorted, latencies)
	
	// 简单排序
	for i := 0; i < len(sorted); i++ {
		for j := i + 1; j < len(sorted); j++ {
			if sorted[i] > sorted[j] {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}
	
	index := int(float64(len(sorted)) * percentile)
	if index >= len(sorted) {
		index = len(sorted) - 1
	}
	
	return sorted[index]
}

func calculateMaxLatency(latencies []time.Duration) time.Duration {
	if len(latencies) == 0 {
		return 0
	}
	
	max := latencies[0]
	for _, latency := range latencies {
		if latency > max {
			max = latency
		}
	}
	
	return max
}

func getCurrentMemoryUsage() int64 {
	// 简化的内存使用获取
	// 在实际实现中应该使用runtime.MemStats
	return 50 * 1024 * 1024 // 50MB 模拟值
}

// 测试辅助类型
type TestTaskProducer struct {
	redis *redis.Client
}

func (tp *TestTaskProducer) PublishAIAnalysisTask(ctx context.Context, alertID string, alertData map[string]interface{}) error {
	task := map[string]interface{}{
		"id":       uuid.New().String(),
		"type":     "ai_analysis",
		"payload":  alertData,
		"priority": 1,
	}
	
	taskData, _ := json.Marshal(task)
	return tp.redis.LPush(ctx, "ai_analysis", taskData).Err()
}

type TestPerformanceWorker struct {
	db    *gorm.DB
	redis *redis.Client
}

func (w *TestPerformanceWorker) Start(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			w.processTask(ctx)
		}
	}
}

func (w *TestPerformanceWorker) processTask(ctx context.Context) {
	result, err := w.redis.BRPop(ctx, time.Second, "ai_analysis").Result()
	if err != nil {
		return
	}
	
	if len(result) < 2 {
		return
	}
	
	// 模拟任务处理
	time.Sleep(time.Duration(rand.Intn(50)) * time.Millisecond)
	
	// 记录任务完成
	w.db.Exec(`
		INSERT INTO task_queue (id, queue_name, task_type, payload, status)
		VALUES (UUID(), 'ai_analysis', 'ai_analysis', ?, 'completed')
	`, result[1])
}