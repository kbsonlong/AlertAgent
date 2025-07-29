package configsyncer

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"go.uber.org/zap"
)

// TestNewConfigSyncer 测试创建配置同步器
func TestNewConfigSyncer(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	config := &Config{
		AlertAgentEndpoint: "http://test",
		ClusterID:         "test-cluster",
		ConfigType:        "prometheus",
		ConfigPath:        "/tmp/test.yml",
		ReloadURL:         "http://test/reload",
		SyncInterval:      30 * time.Second,
		Logger:            logger,
	}

	syncer := NewConfigSyncer(config)

	if syncer == nil {
		t.Fatal("Expected syncer to be created")
	}

	if syncer.config.HTTPTimeout != 30*time.Second {
		t.Errorf("Expected default HTTP timeout to be 30s, got %v", syncer.config.HTTPTimeout)
	}

	if syncer.config.MaxRetries != 3 {
		t.Errorf("Expected default max retries to be 3, got %d", syncer.config.MaxRetries)
	}

	if syncer.config.RetryBackoff != 5*time.Second {
		t.Errorf("Expected default retry backoff to be 5s, got %v", syncer.config.RetryBackoff)
	}

	if !syncer.healthy {
		t.Error("Expected syncer to be healthy initially")
	}
}

// TestValidateConfig 测试配置验证
func TestValidateConfig(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	tests := []struct {
		name        string
		config      *Config
		expectError bool
	}{
		{
			name: "valid config",
			config: &Config{
				AlertAgentEndpoint: "http://test",
				ClusterID:         "test-cluster",
				ConfigType:        "prometheus",
				ConfigPath:        "/tmp/test.yml",
				ReloadURL:         "http://test/reload",
				SyncInterval:      30 * time.Second,
				Logger:            logger,
			},
			expectError: false,
		},
		{
			name: "missing endpoint",
			config: &Config{
				ClusterID:    "test-cluster",
				ConfigType:   "prometheus",
				ConfigPath:   "/tmp/test.yml",
				ReloadURL:    "http://test/reload",
				SyncInterval: 30 * time.Second,
				Logger:       logger,
			},
			expectError: true,
		},
		{
			name: "missing cluster ID",
			config: &Config{
				AlertAgentEndpoint: "http://test",
				ConfigType:        "prometheus",
				ConfigPath:        "/tmp/test.yml",
				ReloadURL:         "http://test/reload",
				SyncInterval:      30 * time.Second,
				Logger:            logger,
			},
			expectError: true,
		},
		{
			name: "invalid sync interval",
			config: &Config{
				AlertAgentEndpoint: "http://test",
				ClusterID:         "test-cluster",
				ConfigType:        "prometheus",
				ConfigPath:        "/tmp/test.yml",
				ReloadURL:         "http://test/reload",
				SyncInterval:      0,
				Logger:            logger,
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			syncer := NewConfigSyncer(tt.config)
			err := syncer.ValidateConfig()

			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}

			if !tt.expectError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
		})
	}
}

// TestHasConfigChanged 测试配置变化检测
func TestHasConfigChanged(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	config := &Config{
		AlertAgentEndpoint: "http://test",
		ClusterID:         "test-cluster",
		ConfigType:        "prometheus",
		ConfigPath:        "/tmp/test.yml",
		ReloadURL:         "http://test/reload",
		SyncInterval:      30 * time.Second,
		Logger:            logger,
	}

	syncer := NewConfigSyncer(config)

	// 初始状态应该检测到变化
	if !syncer.hasConfigChanged("new-hash") {
		t.Error("Expected config change to be detected initially")
	}

	// 设置相同的哈希值
	syncer.lastConfigHash = "same-hash"
	if syncer.hasConfigChanged("same-hash") {
		t.Error("Expected no config change for same hash")
	}

	// 设置不同的哈希值
	if !syncer.hasConfigChanged("different-hash") {
		t.Error("Expected config change for different hash")
	}
}

// TestWriteConfigFileAtomic 测试原子性配置文件写入
func TestWriteConfigFileAtomic(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	// 创建临时目录
	tempDir, err := os.MkdirTemp("", "config-syncer-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	configPath := filepath.Join(tempDir, "test.yml")
	config := &Config{
		AlertAgentEndpoint: "http://test",
		ClusterID:         "test-cluster",
		ConfigType:        "prometheus",
		ConfigPath:        configPath,
		ReloadURL:         "http://test/reload",
		SyncInterval:      30 * time.Second,
		Logger:            logger,
	}

	syncer := NewConfigSyncer(config)

	// 测试写入配置
	testConfig := "test: configuration\nversion: 1.0"
	err = syncer.writeConfigFileAtomic(testConfig)
	if err != nil {
		t.Fatalf("Failed to write config: %v", err)
	}

	// 验证文件内容
	content, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("Failed to read config file: %v", err)
	}

	if string(content) != testConfig {
		t.Errorf("Expected config content %q, got %q", testConfig, string(content))
	}
}

// TestCalculateHash 测试哈希计算
func TestCalculateHash(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	config := &Config{
		AlertAgentEndpoint: "http://test",
		ClusterID:         "test-cluster",
		ConfigType:        "prometheus",
		ConfigPath:        "/tmp/test.yml",
		ReloadURL:         "http://test/reload",
		SyncInterval:      30 * time.Second,
		Logger:            logger,
	}

	syncer := NewConfigSyncer(config)

	// 测试相同内容产生相同哈希
	data := []byte("test data")
	hash1 := syncer.calculateHash(data)
	hash2 := syncer.calculateHash(data)

	if hash1 != hash2 {
		t.Errorf("Expected same hash for same data, got %s and %s", hash1, hash2)
	}

	// 测试不同内容产生不同哈希
	differentData := []byte("different test data")
	hash3 := syncer.calculateHash(differentData)

	if hash1 == hash3 {
		t.Errorf("Expected different hash for different data, got same hash %s", hash1)
	}

	// 验证哈希长度（SHA256应该是64个字符）
	if len(hash1) != 64 {
		t.Errorf("Expected hash length 64, got %d", len(hash1))
	}
}

// TestFetchConfig 测试配置拉取
func TestFetchConfig(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	// 创建模拟服务器
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("cluster_id") != "test-cluster" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if r.URL.Query().Get("type") != "prometheus" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"status": "success",
			"message": "Config retrieved successfully",
			"data": {
				"config": "test: config\nversion: 1.0",
				"config_hash": "abc123",
				"version": "v1.0.0"
			}
		}`))
	}))
	defer server.Close()

	config := &Config{
		AlertAgentEndpoint: server.URL,
		ClusterID:         "test-cluster",
		ConfigType:        "prometheus",
		ConfigPath:        "/tmp/test.yml",
		ReloadURL:         "http://test/reload",
		SyncInterval:      30 * time.Second,
		Logger:            logger,
	}

	syncer := NewConfigSyncer(config)

	ctx := context.Background()
	configResp, err := syncer.fetchConfig(ctx)
	if err != nil {
		t.Fatalf("Failed to fetch config: %v", err)
	}

	if configResp.Status != "success" {
		t.Errorf("Expected status 'success', got %s", configResp.Status)
	}

	if configResp.Data.Config != "test: config\nversion: 1.0" {
		t.Errorf("Expected config content, got %s", configResp.Data.Config)
	}

	if configResp.Data.ConfigHash != "abc123" {
		t.Errorf("Expected config hash 'abc123', got %s", configResp.Data.ConfigHash)
	}

	if configResp.Data.Version != "v1.0.0" {
		t.Errorf("Expected version 'v1.0.0', got %s", configResp.Data.Version)
	}
}

// TestGetMetrics 测试指标获取
func TestGetMetrics(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	config := &Config{
		AlertAgentEndpoint: "http://test",
		ClusterID:         "test-cluster",
		ConfigType:        "prometheus",
		ConfigPath:        "/tmp/test.yml",
		ReloadURL:         "http://test/reload",
		SyncInterval:      30 * time.Second,
		Logger:            logger,
	}

	syncer := NewConfigSyncer(config)

	// 设置一些测试数据
	syncer.syncCount = 10
	syncer.successCount = 8
	syncer.failureCount = 2
	syncer.lastConfigHash = "test-hash"
	syncer.configVersion = "v1.0.0"
	syncer.lastError = "test error"

	metrics := syncer.GetMetrics()

	if metrics.SyncCount != 10 {
		t.Errorf("Expected sync count 10, got %d", metrics.SyncCount)
	}

	if metrics.SuccessCount != 8 {
		t.Errorf("Expected success count 8, got %d", metrics.SuccessCount)
	}

	if metrics.FailureCount != 2 {
		t.Errorf("Expected failure count 2, got %d", metrics.FailureCount)
	}

	if metrics.ConfigHash != "test-hash" {
		t.Errorf("Expected config hash 'test-hash', got %s", metrics.ConfigHash)
	}

	if metrics.ConfigVersion != "v1.0.0" {
		t.Errorf("Expected config version 'v1.0.0', got %s", metrics.ConfigVersion)
	}

	if metrics.LastError != "test error" {
		t.Errorf("Expected last error 'test error', got %s", metrics.LastError)
	}
}

// TestGetHealthStatus 测试健康状态获取
func TestGetHealthStatus(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	config := &Config{
		AlertAgentEndpoint: "http://test",
		ClusterID:         "test-cluster",
		ConfigType:        "prometheus",
		ConfigPath:        "/tmp/test.yml",
		ReloadURL:         "http://test/reload",
		SyncInterval:      30 * time.Second,
		Logger:            logger,
	}

	syncer := NewConfigSyncer(config)

	// 测试健康状态
	syncer.healthy = true
	health := syncer.GetHealthStatus()

	if health.Status != "healthy" {
		t.Errorf("Expected status 'healthy', got %s", health.Status)
	}

	if health.ClusterID != "test-cluster" {
		t.Errorf("Expected cluster ID 'test-cluster', got %s", health.ClusterID)
	}

	if health.ConfigType != "prometheus" {
		t.Errorf("Expected config type 'prometheus', got %s", health.ConfigType)
	}

	// 测试不健康状态
	syncer.healthy = false
	syncer.lastError = "connection failed"
	health = syncer.GetHealthStatus()

	if health.Status != "unhealthy" {
		t.Errorf("Expected status 'unhealthy', got %s", health.Status)
	}

	if health.Error != "connection failed" {
		t.Errorf("Expected error 'connection failed', got %s", health.Error)
	}
}

// TestIsHealthy 测试健康检查
func TestIsHealthy(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	config := &Config{
		AlertAgentEndpoint: "http://test",
		ClusterID:         "test-cluster",
		ConfigType:        "prometheus",
		ConfigPath:        "/tmp/test.yml",
		ReloadURL:         "http://test/reload",
		SyncInterval:      30 * time.Second,
		Logger:            logger,
	}

	syncer := NewConfigSyncer(config)

	// 初始状态应该是健康的
	if !syncer.IsHealthy() {
		t.Error("Expected syncer to be healthy initially")
	}

	// 设置为不健康
	syncer.healthy = false
	if syncer.IsHealthy() {
		t.Error("Expected syncer to be unhealthy")
	}

	// 恢复健康
	syncer.healthy = true
	if !syncer.IsHealthy() {
		t.Error("Expected syncer to be healthy again")
	}
}