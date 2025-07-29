package performance

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// LoadTestConfig 负载测试配置
type LoadTestConfig struct {
	Concurrency int           // 并发数
	Duration    time.Duration // 测试持续时间
	RampUp      time.Duration // 预热时间
	TargetRPS   int           // 目标RPS
}

// LoadTestResult 负载测试结果
type LoadTestResult struct {
	TotalRequests    int64         `json:"total_requests"`
	SuccessRequests  int64         `json:"success_requests"`
	FailedRequests   int64         `json:"failed_requests"`
	AverageLatency   time.Duration `json:"average_latency"`
	P95Latency       time.Duration `json:"p95_latency"`
	P99Latency       time.Duration `json:"p99_latency"`
	MaxLatency       time.Duration `json:"max_latency"`
	MinLatency       time.Duration `json:"min_latency"`
	Throughput       float64       `json:"throughput"` // RPS
	ErrorRate        float64       `json:"error_rate"`
	Duration         time.Duration `json:"duration"`
}

// RequestMetric 单个请求的指标
type RequestMetric struct {
	Latency    time.Duration
	StatusCode int
	Error      error
	Timestamp  time.Time
}

// LoadTester 负载测试器
type LoadTester struct {
	server   *httptest.Server
	client   *http.Client
	baseURL  string
	token    string
	metrics  []RequestMetric
	mu       sync.Mutex
	config   LoadTestConfig
}

// NewLoadTester 创建负载测试器
func NewLoadTester(config LoadTestConfig) *LoadTester {
	gin.SetMode(gin.TestMode)
	router := setupTestRouter()

	server := httptest.NewServer(router)
	client := &http.Client{
		Timeout: 30 * time.Second,
		Transport: &http.Transport{
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 100,
			IdleConnTimeout:     90 * time.Second,
		},
	}

	return &LoadTester{
		server:  server,
		client:  client,
		baseURL: server.URL,
		token:   "mock-jwt-token",
		metrics: make([]RequestMetric, 0),
		config:  config,
	}
}

// setupTestRouter 设置测试路由
func setupTestRouter() *gin.Engine {
	router := gin.New()
	router.Use(gin.Recovery())

	// 健康检查
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"message": "Alert Agent is running",
		})
	})

	// API路由
	v1 := router.Group("/api/v1")
	{
		// 认证
		v1.POST("/auth/login", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"status": "success",
				"data":   gin.H{"token": "mock-jwt-token"},
			})
		})

		// 分析API
		v1.POST("/analysis/submit", func(c *gin.Context) {
			// 模拟处理时间
			time.Sleep(time.Millisecond * 10)
			c.JSON(http.StatusOK, gin.H{
				"status": "success",
				"data":   gin.H{"task_id": "task-123", "status": "queued"},
			})
		})

		v1.GET("/analysis/result/:task_id", func(c *gin.Context) {
			// 模拟处理时间
			time.Sleep(time.Millisecond * 5)
			c.JSON(http.StatusOK, gin.H{
				"status": "success",
				"data":   gin.H{"task_id": c.Param("task_id"), "status": "completed"},
			})
		})

		// 通道API
		v1.GET("/channels", func(c *gin.Context) {
			// 模拟数据库查询时间
			time.Sleep(time.Millisecond * 15)
			c.JSON(http.StatusOK, gin.H{
				"status": "success",
				"data":   gin.H{"items": []gin.H{}, "total": 0},
			})
		})

		v1.POST("/channels", func(c *gin.Context) {
			// 模拟创建操作时间
			time.Sleep(time.Millisecond * 20)
			c.JSON(http.StatusCreated, gin.H{
				"status": "success",
				"data":   gin.H{"id": "channel-123", "name": "Test Channel"},
			})
		})
	}

	return router
}

// Close 关闭测试器
func (lt *LoadTester) Close() {
	if lt.server != nil {
		lt.server.Close()
	}
}

// makeRequest 发送HTTP请求并记录指标
func (lt *LoadTester) makeRequest(method, path string, data interface{}) {
	start := time.Now()
	metric := RequestMetric{
		Timestamp: start,
	}

	var body *bytes.Buffer
	if data != nil {
		jsonData, err := json.Marshal(data)
		if err != nil {
			metric.Error = err
			metric.Latency = time.Since(start)
			lt.recordMetric(metric)
			return
		}
		body = bytes.NewBuffer(jsonData)
	} else {
		body = bytes.NewBuffer(nil)
	}

	req, err := http.NewRequest(method, lt.baseURL+path, body)
	if err != nil {
		metric.Error = err
		metric.Latency = time.Since(start)
		lt.recordMetric(metric)
		return
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+lt.token)

	resp, err := lt.client.Do(req)
	if err != nil {
		metric.Error = err
		metric.Latency = time.Since(start)
		lt.recordMetric(metric)
		return
	}
	defer resp.Body.Close()

	metric.StatusCode = resp.StatusCode
	metric.Latency = time.Since(start)
	lt.recordMetric(metric)
}

// recordMetric 记录请求指标
func (lt *LoadTester) recordMetric(metric RequestMetric) {
	lt.mu.Lock()
	defer lt.mu.Unlock()
	lt.metrics = append(lt.metrics, metric)
}

// RunLoadTest 运行负载测试
func (lt *LoadTester) RunLoadTest(testFunc func()) LoadTestResult {
	lt.metrics = make([]RequestMetric, 0)
	start := time.Now()

	// 创建工作协程
	var wg sync.WaitGroup
	stopCh := make(chan struct{})

	// 启动并发工作者
	for i := 0; i < lt.config.Concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-stopCh:
					return
				default:
					testFunc()
					// 控制请求速率
					if lt.config.TargetRPS > 0 {
						interval := time.Duration(int64(time.Second) / int64(lt.config.TargetRPS*lt.config.Concurrency))
						time.Sleep(interval)
					}
				}
			}
		}()
	}

	// 等待测试完成
	time.Sleep(lt.config.Duration)
	close(stopCh)
	wg.Wait()

	duration := time.Since(start)
	return lt.calculateResults(duration)
}

// calculateResults 计算测试结果
func (lt *LoadTester) calculateResults(duration time.Duration) LoadTestResult {
	lt.mu.Lock()
	defer lt.mu.Unlock()

	totalRequests := int64(len(lt.metrics))
	if totalRequests == 0 {
		return LoadTestResult{Duration: duration}
	}

	var successRequests, failedRequests int64
	var totalLatency time.Duration
	latencies := make([]time.Duration, 0, totalRequests)
	minLatency := time.Hour
	maxLatency := time.Duration(0)

	for _, metric := range lt.metrics {
		if metric.Error != nil || metric.StatusCode >= 400 {
			failedRequests++
		} else {
			successRequests++
		}

		totalLatency += metric.Latency
		latencies = append(latencies, metric.Latency)

		if metric.Latency < minLatency {
			minLatency = metric.Latency
		}
		if metric.Latency > maxLatency {
			maxLatency = metric.Latency
		}
	}

	// 计算百分位数
	p95Latency := calculatePercentile(latencies, 95)
	p99Latency := calculatePercentile(latencies, 99)

	averageLatency := totalLatency / time.Duration(totalRequests)
	throughput := float64(totalRequests) / duration.Seconds()
	errorRate := float64(failedRequests) / float64(totalRequests) * 100

	return LoadTestResult{
		TotalRequests:   totalRequests,
		SuccessRequests: successRequests,
		FailedRequests:  failedRequests,
		AverageLatency:  averageLatency,
		P95Latency:      p95Latency,
		P99Latency:      p99Latency,
		MaxLatency:      maxLatency,
		MinLatency:      minLatency,
		Throughput:      throughput,
		ErrorRate:       errorRate,
		Duration:        duration,
	}
}

// calculatePercentile 计算百分位数
func calculatePercentile(latencies []time.Duration, percentile int) time.Duration {
	if len(latencies) == 0 {
		return 0
	}

	// 简单排序
	for i := 0; i < len(latencies)-1; i++ {
		for j := 0; j < len(latencies)-i-1; j++ {
			if latencies[j] > latencies[j+1] {
				latencies[j], latencies[j+1] = latencies[j+1], latencies[j]
			}
		}
	}

	index := int(float64(len(latencies)) * float64(percentile) / 100.0)
	if index >= len(latencies) {
		index = len(latencies) - 1
	}
	return latencies[index]
}

// TestHealthCheckLoad 健康检查负载测试
func TestHealthCheckLoad(t *testing.T) {
	if os.Getenv("LOAD_TEST") == "" {
		t.Skip("Skipping load tests. Set LOAD_TEST=1 to run.")
	}

	config := LoadTestConfig{
		Concurrency: 10,
		Duration:    30 * time.Second,
		TargetRPS:   100,
	}

	loadTester := NewLoadTester(config)
	defer loadTester.Close()

	result := loadTester.RunLoadTest(func() {
		loadTester.makeRequest("GET", "/health", nil)
	})

	// 验证结果
	assert.Greater(t, result.TotalRequests, int64(0))
	assert.Less(t, result.ErrorRate, 5.0) // 错误率小于5%
	assert.Greater(t, result.Throughput, 50.0) // 吞吐量大于50 RPS
	assert.Less(t, result.AverageLatency, 100*time.Millisecond) // 平均延迟小于100ms

	t.Logf("Health Check Load Test Results:")
	t.Logf("Total Requests: %d", result.TotalRequests)
	t.Logf("Success Rate: %.2f%%", float64(result.SuccessRequests)/float64(result.TotalRequests)*100)
	t.Logf("Error Rate: %.2f%%", result.ErrorRate)
	t.Logf("Throughput: %.2f RPS", result.Throughput)
	t.Logf("Average Latency: %v", result.AverageLatency)
	t.Logf("P95 Latency: %v", result.P95Latency)
	t.Logf("P99 Latency: %v", result.P99Latency)
}

// TestAnalysisAPILoad 分析API负载测试
func TestAnalysisAPILoad(t *testing.T) {
	if os.Getenv("LOAD_TEST") == "" {
		t.Skip("Skipping load tests. Set LOAD_TEST=1 to run.")
	}

	config := LoadTestConfig{
		Concurrency: 20,
		Duration:    60 * time.Second,
		TargetRPS:   50,
	}

	loadTester := NewLoadTester(config)
	defer loadTester.Close()

	result := loadTester.RunLoadTest(func() {
		submitData := map[string]interface{}{
			"alert_id": 12345,
			"type":     "root_cause",
			"priority": 8,
		}
		loadTester.makeRequest("POST", "/api/v1/analysis/submit", submitData)
	})

	// 验证结果
	assert.Greater(t, result.TotalRequests, int64(0))
	assert.Less(t, result.ErrorRate, 10.0) // 错误率小于10%
	assert.Greater(t, result.Throughput, 20.0) // 吞吐量大于20 RPS
	assert.Less(t, result.AverageLatency, 200*time.Millisecond) // 平均延迟小于200ms

	t.Logf("Analysis API Load Test Results:")
	t.Logf("Total Requests: %d", result.TotalRequests)
	t.Logf("Success Rate: %.2f%%", float64(result.SuccessRequests)/float64(result.TotalRequests)*100)
	t.Logf("Error Rate: %.2f%%", result.ErrorRate)
	t.Logf("Throughput: %.2f RPS", result.Throughput)
	t.Logf("Average Latency: %v", result.AverageLatency)
	t.Logf("P95 Latency: %v", result.P95Latency)
	t.Logf("P99 Latency: %v", result.P99Latency)
}

// TestChannelAPILoad 通道API负载测试
func TestChannelAPILoad(t *testing.T) {
	if os.Getenv("LOAD_TEST") == "" {
		t.Skip("Skipping load tests. Set LOAD_TEST=1 to run.")
	}

	config := LoadTestConfig{
		Concurrency: 15,
		Duration:    45 * time.Second,
		TargetRPS:   30,
	}

	loadTester := NewLoadTester(config)
	defer loadTester.Close()

	result := loadTester.RunLoadTest(func() {
		// 混合读写操作
		if time.Now().UnixNano()%2 == 0 {
			// 读操作
			loadTester.makeRequest("GET", "/api/v1/channels", nil)
		} else {
			// 写操作
			createData := map[string]interface{}{
				"name": "Test Channel",
				"type": "slack",
				"config": map[string]interface{}{
					"webhook_url": "https://hooks.slack.com/test",
				},
			}
			loadTester.makeRequest("POST", "/api/v1/channels", createData)
		}
	})

	// 验证结果
	assert.Greater(t, result.TotalRequests, int64(0))
	assert.Less(t, result.ErrorRate, 15.0) // 错误率小于15%
	assert.Greater(t, result.Throughput, 15.0) // 吞吐量大于15 RPS
	assert.Less(t, result.AverageLatency, 300*time.Millisecond) // 平均延迟小于300ms

	t.Logf("Channel API Load Test Results:")
	t.Logf("Total Requests: %d", result.TotalRequests)
	t.Logf("Success Rate: %.2f%%", float64(result.SuccessRequests)/float64(result.TotalRequests)*100)
	t.Logf("Error Rate: %.2f%%", result.ErrorRate)
	t.Logf("Throughput: %.2f RPS", result.Throughput)
	t.Logf("Average Latency: %v", result.AverageLatency)
	t.Logf("P95 Latency: %v", result.P95Latency)
	t.Logf("P99 Latency: %v", result.P99Latency)
}

// BenchmarkConcurrentRequests 并发请求基准测试
func BenchmarkConcurrentRequests(b *testing.B) {
	if os.Getenv("LOAD_TEST") == "" {
		b.Skip("Skipping load benchmarks. Set LOAD_TEST=1 to run.")
	}

	config := LoadTestConfig{
		Concurrency: 50,
		Duration:    10 * time.Second,
	}

	loadTester := NewLoadTester(config)
	defer loadTester.Close()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			loadTester.makeRequest("GET", "/health", nil)
		}
	})
}

// BenchmarkAPIThroughput API吞吐量基准测试
func BenchmarkAPIThroughput(b *testing.B) {
	if os.Getenv("LOAD_TEST") == "" {
		b.Skip("Skipping load benchmarks. Set LOAD_TEST=1 to run.")
	}

	config := LoadTestConfig{
		Concurrency: 1,
		Duration:    5 * time.Second,
	}

	loadTester := NewLoadTester(config)
	defer loadTester.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		loadTester.makeRequest("GET", "/health", nil)
	}
}