package performance

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// setupPerformanceTestEnvironment 设置性能测试环境
func setupPerformanceTestEnvironment(t *testing.T) (*gorm.DB, *redis.Client, func()) {
	// 设置测试数据库
	testDB := setupPerformanceTestDatabase(t)
	
	// 设置测试Redis
	testRedis := setupPerformanceTestRedis(t)
	
	// 初始化测试数据
	initPerformanceTestData(t, testDB)
	
	// 返回清理函数
	cleanup := func() {
		cleanupPerformanceTestData(testDB, testRedis)
	}
	
	return testDB, testRedis, cleanup
}

func setupPerformanceTestDatabase(t *testing.T) *gorm.DB {
	// 使用性能测试数据库配置
	dsn := getPerformanceTestDSN()
	
	// 创建测试数据库
	if err := createPerformanceTestDatabase(); err != nil {
		t.Fatalf("Failed to create performance test database: %v", err)
	}
	
	// 连接到测试数据库
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent), // 静默模式以提高性能
		PrepareStmt: true, // 启用预编译语句
		DisableForeignKeyConstraintWhenMigrating: true, // 禁用外键约束以提高性能
	})
	if err != nil {
		t.Fatalf("Failed to connect to performance test database: %v", err)
	}
	
	// 优化数据库连接池
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("Failed to get underlying sql.DB: %v", err)
	}
	
	// 性能优化配置
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetMaxIdleConns(50)
	sqlDB.SetConnMaxLifetime(time.Hour)
	sqlDB.SetConnMaxIdleTime(time.Minute * 30)
	
	// 运行性能测试迁移
	if err := runPerformanceTestMigrations(db); err != nil {
		t.Fatalf("Failed to run performance test migrations: %v", err)
	}
	
	return db
}

func setupPerformanceTestRedis(t *testing.T) *redis.Client {
	redisClient := redis.NewClient(&redis.Options{
		Addr:         getPerformanceTestRedisAddr(),
		Password:     "",
		DB:           2, // 使用专门的性能测试数据库
		PoolSize:     100, // 增加连接池大小
		MinIdleConns: 20,
		MaxRetries:   3,
		DialTimeout:  time.Second * 5,
		ReadTimeout:  time.Second * 3,
		WriteTimeout: time.Second * 3,
	})
	
	// 测试连接
	ctx := context.Background()
	_, err := redisClient.Ping(ctx).Result()
	if err != nil {
		t.Fatalf("Failed to connect to performance test redis: %v", err)
	}
	
	// 清空性能测试数据库
	if err := redisClient.FlushDB(ctx).Err(); err != nil {
		t.Fatalf("Failed to flush performance test redis: %v", err)
	}
	
	return redisClient
}

func getPerformanceTestDSN() string {
	host := getEnvOrDefault("PERF_TEST_DB_HOST", "localhost")
	port := getEnvOrDefault("PERF_TEST_DB_PORT", "3306")
	user := getEnvOrDefault("PERF_TEST_DB_USER", "root")
	password := getEnvOrDefault("PERF_TEST_DB_PASSWORD", "password")
	dbname := getEnvOrDefault("PERF_TEST_DB_NAME", "alertagent_perf_test")
	
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local&timeout=30s&readTimeout=30s&writeTimeout=30s",
		user, password, host, port, dbname)
}

func getPerformanceTestRedisAddr() string {
	host := getEnvOrDefault("PERF_TEST_REDIS_HOST", "localhost")
	port := getEnvOrDefault("PERF_TEST_REDIS_PORT", "6379")
	return fmt.Sprintf("%s:%s", host, port)
}

func createPerformanceTestDatabase() error {
	host := getEnvOrDefault("PERF_TEST_DB_HOST", "localhost")
	port := getEnvOrDefault("PERF_TEST_DB_PORT", "3306")
	user := getEnvOrDefault("PERF_TEST_DB_USER", "root")
	password := getEnvOrDefault("PERF_TEST_DB_PASSWORD", "password")
	dbname := getEnvOrDefault("PERF_TEST_DB_NAME", "alertagent_perf_test")
	
	// 连接到MySQL服务器（不指定数据库）
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/", user, password, host, port)
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return err
	}
	defer db.Close()
	
	// 创建性能测试数据库
	_, err = db.Exec(fmt.Sprintf("CREATE DATABASE IF NOT EXISTS %s", dbname))
	return err
}

func runPerformanceTestMigrations(db *gorm.DB) error {
	// 创建性能测试所需的表结构
	migrations := []string{
		`CREATE TABLE IF NOT EXISTS alert_rules (
			id VARCHAR(36) PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			expression TEXT NOT NULL,
			duration VARCHAR(50) NOT NULL,
			severity VARCHAR(20) NOT NULL,
			labels JSON,
			annotations JSON,
			targets JSON,
			version VARCHAR(50) NOT NULL DEFAULT 'v1.0.0',
			status ENUM('pending', 'active', 'inactive', 'error') DEFAULT 'pending',
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			INDEX idx_name (name),
			INDEX idx_status (status),
			INDEX idx_created_at (created_at)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
		
		`CREATE TABLE IF NOT EXISTS config_sync_status (
			id VARCHAR(36) PRIMARY KEY,
			cluster_id VARCHAR(100) NOT NULL,
			config_type VARCHAR(50) NOT NULL,
			config_hash VARCHAR(64),
			sync_status ENUM('success', 'failed', 'pending') DEFAULT 'pending',
			sync_time TIMESTAMP,
			error_message TEXT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			UNIQUE KEY uk_cluster_type (cluster_id, config_type),
			INDEX idx_sync_status (sync_status),
			INDEX idx_sync_time (sync_time),
			INDEX idx_cluster_id (cluster_id)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
		
		`CREATE TABLE IF NOT EXISTS task_queue (
			id VARCHAR(36) PRIMARY KEY,
			queue_name VARCHAR(100) NOT NULL,
			task_type VARCHAR(50) NOT NULL,
			payload JSON NOT NULL,
			priority INT DEFAULT 0,
			retry_count INT DEFAULT 0,
			max_retry INT DEFAULT 3,
			status ENUM('pending', 'processing', 'completed', 'failed') DEFAULT 'pending',
			scheduled_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			started_at TIMESTAMP NULL,
			completed_at TIMESTAMP NULL,
			error_message TEXT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			INDEX idx_queue_status (queue_name, status),
			INDEX idx_scheduled_at (scheduled_at),
			INDEX idx_priority (priority),
			INDEX idx_task_type (task_type),
			INDEX idx_status (status)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
		
		`CREATE TABLE IF NOT EXISTS alerts (
			id VARCHAR(36) PRIMARY KEY,
			alertname VARCHAR(255) NOT NULL,
			instance VARCHAR(255),
			severity VARCHAR(20),
			status ENUM('firing', 'resolved', 'analyzing', 'analyzed') DEFAULT 'firing',
			labels JSON,
			annotations JSON,
			starts_at TIMESTAMP,
			ends_at TIMESTAMP NULL,
			analysis_result JSON,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			INDEX idx_alertname (alertname),
			INDEX idx_status (status),
			INDEX idx_severity (severity),
			INDEX idx_starts_at (starts_at),
			INDEX idx_created_at (created_at)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
		
		`CREATE TABLE IF NOT EXISTS notification_logs (
			id VARCHAR(36) PRIMARY KEY,
			alert_id VARCHAR(36),
			plugin_name VARCHAR(100),
			channel_config JSON,
			message JSON,
			status ENUM('pending', 'sent', 'failed') DEFAULT 'pending',
			error_message TEXT,
			sent_at TIMESTAMP NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			INDEX idx_alert_id (alert_id),
			INDEX idx_plugin_name (plugin_name),
			INDEX idx_status (status),
			INDEX idx_sent_at (sent_at)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
	}
	
	for _, migration := range migrations {
		if err := db.Exec(migration).Error; err != nil {
			return fmt.Errorf("failed to execute migration: %w", err)
		}
	}
	
	return nil
}

func initPerformanceTestData(t *testing.T, db *gorm.DB) {
	// 插入一些基础测试数据以提供更真实的测试环境
	
	// 插入测试告警规则
	rules := []map[string]interface{}{
		{
			"id":          "rule-cpu-high",
			"name":        "High CPU Usage",
			"expression":  "cpu_usage > 0.8",
			"duration":    "5m",
			"severity":    "critical",
			"labels":      `{"team": "infrastructure"}`,
			"annotations": `{"summary": "High CPU usage detected"}`,
			"targets":     `["cluster-1", "cluster-2"]`,
			"status":      "active",
		},
		{
			"id":          "rule-memory-high",
			"name":        "High Memory Usage",
			"expression":  "memory_usage > 0.9",
			"duration":    "3m",
			"severity":    "warning",
			"labels":      `{"team": "infrastructure"}`,
			"annotations": `{"summary": "High memory usage detected"}`,
			"targets":     `["cluster-1"]`,
			"status":      "active",
		},
	}
	
	for _, rule := range rules {
		db.Exec(`
			INSERT IGNORE INTO alert_rules 
			(id, name, expression, duration, severity, labels, annotations, targets, status)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
		`, rule["id"], rule["name"], rule["expression"], rule["duration"],
			rule["severity"], rule["labels"], rule["annotations"], rule["targets"], rule["status"])
	}
	
	// 插入一些历史告警数据
	for i := 0; i < 100; i++ {
		alertID := fmt.Sprintf("alert-history-%d", i)
		db.Exec(`
			INSERT IGNORE INTO alerts 
			(id, alertname, instance, severity, status, labels, starts_at)
			VALUES (?, ?, ?, ?, ?, ?, ?)
		`, alertID, "TestAlert", fmt.Sprintf("server-%02d", i%10), "warning", "resolved",
			`{"job": "prometheus"}`, time.Now().Add(-time.Duration(i)*time.Minute))
	}
}

func clearPerformanceTestData(db *gorm.DB, redisClient *redis.Client) {
	// 清理数据库表
	tables := []string{
		"notification_logs",
		"alerts", 
		"task_queue",
		"config_sync_status",
		"alert_rules",
	}
	
	for _, table := range tables {
		db.Exec(fmt.Sprintf("DELETE FROM %s WHERE created_at > DATE_SUB(NOW(), INTERVAL 1 HOUR)", table))
	}
	
	// 清理Redis
	ctx := context.Background()
	redisClient.FlushDB(ctx)
}

func cleanupPerformanceTestData(db *gorm.DB, redisClient *redis.Client) {
	if db != nil {
		// 删除性能测试数据库
		db.Exec("DROP DATABASE IF EXISTS alertagent_perf_test")
	}
	
	if redisClient != nil {
		// 清空Redis测试数据库
		ctx := context.Background()
		redisClient.FlushDB(ctx)
		redisClient.Close()
	}
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// PerformanceTestSuite 性能测试套件
type PerformanceTestSuite struct {
	DB    *gorm.DB
	Redis *redis.Client
	Monitor *ResourceMonitor
}

// NewPerformanceTestSuite 创建性能测试套件
func NewPerformanceTestSuite(t *testing.T) *PerformanceTestSuite {
	db, redis, _ := setupPerformanceTestEnvironment(t)
	
	return &PerformanceTestSuite{
		DB:      db,
		Redis:   redis,
		Monitor: NewResourceMonitor(),
	}
}

// RunBenchmark 运行基准测试
func (pts *PerformanceTestSuite) RunBenchmark(name string, operation func() error, config BenchmarkConfig) *BenchmarkResult {
	benchmark := NewBenchmark(name, operation)
	
	if config.Setup != nil {
		benchmark.WithSetup(config.Setup)
	}
	
	if config.Teardown != nil {
		benchmark.WithTeardown(config.Teardown)
	}
	
	result, err := benchmark.Run(config.Operations, config.Concurrency)
	if err != nil {
		log.Printf("Benchmark %s failed: %v", name, err)
		return nil
	}
	
	return result
}

// RunLoadTest 运行负载测试
func (pts *PerformanceTestSuite) RunLoadTest(operation func() error, config LoadTestConfig) *LoadTestResult {
	loadTester := NewLoadTester(config, operation)
	
	ctx := context.Background()
	result, err := loadTester.Run(ctx)
	if err != nil {
		log.Printf("Load test failed: %v", err)
		return nil
	}
	
	return result
}

// BenchmarkConfig 基准测试配置
type BenchmarkConfig struct {
	Operations  int
	Concurrency int
	Setup       func() error
	Teardown    func() error
}

// PerformanceMetrics 性能指标
type PerformanceMetrics struct {
	TPS           float64       `json:"tps"`
	AvgLatency    time.Duration `json:"avg_latency"`
	P95Latency    time.Duration `json:"p95_latency"`
	P99Latency    time.Duration `json:"p99_latency"`
	ErrorRate     float64       `json:"error_rate"`
	MemoryUsage   int64         `json:"memory_usage"`
	CPUUsage      float64       `json:"cpu_usage"`
	GoroutineCount int          `json:"goroutine_count"`
}

// CollectMetrics 收集性能指标
func (pts *PerformanceTestSuite) CollectMetrics() *PerformanceMetrics {
	report := pts.Monitor.GetReport()
	
	return &PerformanceMetrics{
		MemoryUsage:    report.PeakMemoryUsage,
		CPUUsage:       report.PeakCPUUsage,
		GoroutineCount: report.GoroutineCount,
	}
}

// ValidatePerformanceRequirements 验证性能要求
func ValidatePerformanceRequirements(t *testing.T, result *BenchmarkResult, requirements PerformanceRequirements) {
	if requirements.MinTPS > 0 && result.OpsPerSecond < requirements.MinTPS {
		t.Errorf("TPS %f is below minimum requirement %f", result.OpsPerSecond, requirements.MinTPS)
	}
	
	if requirements.MaxLatency > 0 && result.P95Latency > requirements.MaxLatency {
		t.Errorf("P95 latency %v exceeds maximum requirement %v", result.P95Latency, requirements.MaxLatency)
	}
	
	if requirements.MaxErrorRate > 0 && result.ErrorRate > requirements.MaxErrorRate {
		t.Errorf("Error rate %f exceeds maximum requirement %f", result.ErrorRate, requirements.MaxErrorRate)
	}
	
	if requirements.MaxMemoryUsage > 0 && result.MemoryUsage > requirements.MaxMemoryUsage {
		t.Errorf("Memory usage %d exceeds maximum requirement %d", result.MemoryUsage, requirements.MaxMemoryUsage)
	}
}

// PerformanceRequirements 性能要求
type PerformanceRequirements struct {
	MinTPS          float64       `json:"min_tps"`
	MaxLatency      time.Duration `json:"max_latency"`
	MaxErrorRate    float64       `json:"max_error_rate"`
	MaxMemoryUsage  int64         `json:"max_memory_usage"`
}

// GeneratePerformanceReport 生成性能报告
func GeneratePerformanceReport(results []*BenchmarkResult) string {
	var report strings.Builder
	
	report.WriteString("# Performance Test Report\n\n")
	report.WriteString("| Test Name | Operations | Duration | TPS | Avg Latency | P95 Latency | P99 Latency | Error Rate | Memory Usage |\n")
	report.WriteString("|-----------|------------|----------|-----|-------------|-------------|-------------|------------|-------------|\n")
	
	for _, result := range results {
		report.WriteString(fmt.Sprintf("| %s | %d | %v | %.2f | %v | %v | %v | %.2f%% | %d MB |\n",
			result.Name,
			result.Operations,
			result.Duration,
			result.OpsPerSecond,
			result.AvgLatency,
			result.P95Latency,
			result.P99Latency,
			result.ErrorRate*100,
			result.MemoryUsage/1024/1024,
		))
	}
	
	return report.String()
}

// SavePerformanceReport 保存性能报告
func SavePerformanceReport(report string, filename string) error {
	return os.WriteFile(filename, []byte(report), 0644)
}