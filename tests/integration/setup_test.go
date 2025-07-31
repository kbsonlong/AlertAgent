package integration

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var (
	testDB    *gorm.DB
	testRedis *redis.Client
	ctx       = context.Background()
)

// TestMain 设置和清理测试环境
func TestMain(m *testing.M) {
	// 设置测试环境
	if err := setupTestEnvironment(); err != nil {
		log.Fatalf("Failed to setup test environment: %v", err)
	}

	// 运行测试
	code := m.Run()

	// 清理测试环境
	cleanupTestEnvironment()

	os.Exit(code)
}

func setupTestEnvironment() error {
	// 设置测试数据库
	if err := setupTestDatabase(); err != nil {
		return fmt.Errorf("failed to setup test database: %w", err)
	}

	// 设置测试Redis
	if err := setupTestRedis(); err != nil {
		return fmt.Errorf("failed to setup test redis: %w", err)
	}

	// 初始化测试数据
	if err := initTestData(); err != nil {
		return fmt.Errorf("failed to init test data: %w", err)
	}

	return nil
}

func setupTestDatabase() error {
	// 使用测试数据库配置
	dsn := getTestDSN()
	
	// 创建测试数据库
	if err := createTestDatabase(); err != nil {
		return err
	}

	// 连接到测试数据库
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return fmt.Errorf("failed to connect to test database: %w", err)
	}

	testDB = db

	// 运行迁移
	if err := runTestMigrations(); err != nil {
		return fmt.Errorf("failed to run test migrations: %w", err)
	}

	return nil
}

func setupTestRedis() error {
	testRedis = redis.NewClient(&redis.Options{
		Addr:     getTestRedisAddr(),
		Password: "",
		DB:       1, // 使用测试数据库
	})

	// 测试连接
	_, err := testRedis.Ping(ctx).Result()
	if err != nil {
		return fmt.Errorf("failed to connect to test redis: %w", err)
	}

	// 清空测试数据库
	if err := testRedis.FlushDB(ctx).Err(); err != nil {
		return fmt.Errorf("failed to flush test redis: %w", err)
	}

	return nil
}

func getTestDSN() string {
	host := getEnvOrDefault("TEST_DB_HOST", "localhost")
	port := getEnvOrDefault("TEST_DB_PORT", "3306")
	user := getEnvOrDefault("TEST_DB_USER", "root")
	password := getEnvOrDefault("TEST_DB_PASSWORD", "password")
	dbname := getEnvOrDefault("TEST_DB_NAME", "alertagent_test")

	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		user, password, host, port, dbname)
}

func getTestRedisAddr() string {
	host := getEnvOrDefault("TEST_REDIS_HOST", "localhost")
	port := getEnvOrDefault("TEST_REDIS_PORT", "6379")
	return fmt.Sprintf("%s:%s", host, port)
}

func createTestDatabase() error {
	host := getEnvOrDefault("TEST_DB_HOST", "localhost")
	port := getEnvOrDefault("TEST_DB_PORT", "3306")
	user := getEnvOrDefault("TEST_DB_USER", "root")
	password := getEnvOrDefault("TEST_DB_PASSWORD", "password")
	dbname := getEnvOrDefault("TEST_DB_NAME", "alertagent_test")

	// 连接到MySQL服务器（不指定数据库）
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/", user, password, host, port)
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return err
	}
	defer db.Close()

	// 创建测试数据库
	_, err = db.Exec(fmt.Sprintf("CREATE DATABASE IF NOT EXISTS %s", dbname))
	return err
}

func runTestMigrations() error {
	// 这里应该运行所有必要的数据库迁移
	// 为了简化，我们直接创建必要的表结构
	
	// 创建告警规则表
	if err := testDB.Exec(`
		CREATE TABLE IF NOT EXISTS alert_rules (
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
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
		)
	`).Error; err != nil {
		return err
	}

	// 创建配置同步状态表
	if err := testDB.Exec(`
		CREATE TABLE IF NOT EXISTS config_sync_status (
			id VARCHAR(36) PRIMARY KEY,
			cluster_id VARCHAR(100) NOT NULL,
			config_type VARCHAR(50) NOT NULL,
			config_hash VARCHAR(64),
			sync_status ENUM('success', 'failed', 'pending') DEFAULT 'pending',
			sync_time TIMESTAMP,
			error_message TEXT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			UNIQUE KEY uk_cluster_type (cluster_id, config_type)
		)
	`).Error; err != nil {
		return err
	}

	// 创建任务队列表
	if err := testDB.Exec(`
		CREATE TABLE IF NOT EXISTS task_queue (
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
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
		)
	`).Error; err != nil {
		return err
	}

	// 创建通知插件配置表
	if err := testDB.Exec(`
		CREATE TABLE IF NOT EXISTS notification_plugins (
			id VARCHAR(36) PRIMARY KEY,
			name VARCHAR(100) NOT NULL UNIQUE,
			display_name VARCHAR(255) NOT NULL,
			version VARCHAR(50) NOT NULL,
			config_schema JSON,
			enabled BOOLEAN DEFAULT TRUE,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
		)
	`).Error; err != nil {
		return err
	}

	return nil
}

func initTestData() error {
	// 插入测试用的通知插件配置
	plugins := []map[string]interface{}{
		{
			"id":           "dingtalk-plugin",
			"name":         "dingtalk",
			"display_name": "钉钉通知",
			"version":      "1.0.0",
			"config_schema": `{
				"type": "object",
				"properties": {
					"webhook_url": {"type": "string", "required": true},
					"secret": {"type": "string"},
					"at_all": {"type": "boolean", "default": false}
				}
			}`,
			"enabled": true,
		},
		{
			"id":           "email-plugin",
			"name":         "email",
			"display_name": "邮件通知",
			"version":      "1.0.0",
			"config_schema": `{
				"type": "object",
				"properties": {
					"smtp_host": {"type": "string", "required": true},
					"smtp_port": {"type": "integer", "required": true},
					"username": {"type": "string", "required": true},
					"password": {"type": "string", "required": true}
				}
			}`,
			"enabled": true,
		},
	}

	for _, plugin := range plugins {
		if err := testDB.Exec(`
			INSERT IGNORE INTO notification_plugins 
			(id, name, display_name, version, config_schema, enabled) 
			VALUES (?, ?, ?, ?, ?, ?)
		`, plugin["id"], plugin["name"], plugin["display_name"], 
			plugin["version"], plugin["config_schema"], plugin["enabled"]).Error; err != nil {
			return err
		}
	}

	return nil
}

func cleanupTestEnvironment() {
	if testDB != nil {
		// 清理测试数据
		testDB.Exec("DROP DATABASE IF EXISTS alertagent_test")
	}

	if testRedis != nil {
		// 清空Redis测试数据库
		testRedis.FlushDB(ctx)
		testRedis.Close()
	}
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// 测试辅助函数
func waitForCondition(condition func() bool, timeout time.Duration, interval time.Duration) bool {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if condition() {
			return true
		}
		time.Sleep(interval)
	}
	return false
}

func clearTestData() error {
	tables := []string{
		"alert_rules",
		"config_sync_status", 
		"task_queue",
	}

	for _, table := range tables {
		if err := testDB.Exec(fmt.Sprintf("DELETE FROM %s", table)).Error; err != nil {
			return err
		}
	}

	return testRedis.FlushDB(ctx).Err()
}