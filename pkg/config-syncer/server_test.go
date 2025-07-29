package configsyncer

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"go.uber.org/zap"
)

// TestNewHTTPServer 测试创建HTTP服务器
func TestNewHTTPServer(t *testing.T) {
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
	server := NewHTTPServer(syncer, 8080)

	if server == nil {
		t.Fatal("Expected server to be created")
	}

	if server.port != 8080 {
		t.Errorf("Expected port 8080, got %d", server.port)
	}

	if server.syncer != syncer {
		t.Error("Expected syncer to be set correctly")
	}

	if server.logger != logger {
		t.Error("Expected logger to be set correctly")
	}
}

// TestHealthHandler 测试健康检查端点
func TestHealthHandler(t *testing.T) {
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
	server := NewHTTPServer(syncer, 8080)

	tests := []struct {
		name           string
		path           string
		healthy        bool
		expectedStatus int
	}{
		{
			name:           "health endpoint - healthy",
			path:           "/health",
			healthy:        true,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "health endpoint - unhealthy",
			path:           "/health",
			healthy:        false,
			expectedStatus: http.StatusServiceUnavailable,
		},
		{
			name:           "healthz endpoint - healthy",
			path:           "/healthz",
			healthy:        true,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "healthz endpoint - unhealthy",
			path:           "/healthz",
			healthy:        false,
			expectedStatus: http.StatusServiceUnavailable,
		},
		{
			name:           "ready endpoint - healthy",
			path:           "/ready",
			healthy:        true,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "ready endpoint - unhealthy",
			path:           "/ready",
			healthy:        false,
			expectedStatus: http.StatusServiceUnavailable,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 设置同步器健康状态
			syncer.healthy = tt.healthy

			req := httptest.NewRequest("GET", tt.path, nil)
			w := httptest.NewRecorder()

			server.healthHandler(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			// 验证响应内容
			var response HealthStatus
			err := json.Unmarshal(w.Body.Bytes(), &response)
			if err != nil {
				t.Fatalf("Failed to unmarshal response: %v", err)
			}

			expectedStatus := "healthy"
			if !tt.healthy {
				expectedStatus = "unhealthy"
			}

			if response.Status != expectedStatus {
				t.Errorf("Expected status %s, got %s", expectedStatus, response.Status)
			}

			if response.ClusterID != "test-cluster" {
				t.Errorf("Expected cluster ID 'test-cluster', got %s", response.ClusterID)
			}

			if response.ConfigType != "prometheus" {
				t.Errorf("Expected config type 'prometheus', got %s", response.ConfigType)
			}
		})
	}
}

// TestMetricsHandler 测试指标端点
func TestMetricsHandler(t *testing.T) {
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
	server := NewHTTPServer(syncer, 8080)

	// 设置一些测试指标
	syncer.syncCount = 5
	syncer.successCount = 4
	syncer.failureCount = 1
	syncer.lastConfigHash = "test-hash"
	syncer.configVersion = "v1.0.0"
	syncer.lastError = "test error"

	req := httptest.NewRequest("GET", "/metrics", nil)
	w := httptest.NewRecorder()

	server.metricsHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	// 验证响应内容
	var response SyncMetrics
	err := json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response.SyncCount != 5 {
		t.Errorf("Expected sync count 5, got %d", response.SyncCount)
	}

	if response.SuccessCount != 4 {
		t.Errorf("Expected success count 4, got %d", response.SuccessCount)
	}

	if response.FailureCount != 1 {
		t.Errorf("Expected failure count 1, got %d", response.FailureCount)
	}

	if response.ConfigHash != "test-hash" {
		t.Errorf("Expected config hash 'test-hash', got %s", response.ConfigHash)
	}

	if response.ConfigVersion != "v1.0.0" {
		t.Errorf("Expected config version 'v1.0.0', got %s", response.ConfigVersion)
	}

	if response.LastError != "test error" {
		t.Errorf("Expected last error 'test error', got %s", response.LastError)
	}
}

// TestStatusHandler 测试状态端点
func TestStatusHandler(t *testing.T) {
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
	server := NewHTTPServer(syncer, 8080)

	req := httptest.NewRequest("GET", "/status", nil)
	w := httptest.NewRecorder()

	server.statusHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	// 验证响应内容
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	// 检查是否包含健康状态和指标
	if _, ok := response["health"]; !ok {
		t.Error("Expected health field in status response")
	}

	if _, ok := response["metrics"]; !ok {
		t.Error("Expected metrics field in status response")
	}
}

// TestRootHandler 测试根端点
func TestRootHandler(t *testing.T) {
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
	server := NewHTTPServer(syncer, 8080)

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	server.rootHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	// 验证响应内容
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response["service"] != "config-syncer" {
		t.Errorf("Expected service 'config-syncer', got %v", response["service"])
	}

	if response["cluster_id"] != "test-cluster" {
		t.Errorf("Expected cluster_id 'test-cluster', got %v", response["cluster_id"])
	}

	if response["config_type"] != "prometheus" {
		t.Errorf("Expected config_type 'prometheus', got %v", response["config_type"])
	}

	// 检查是否包含端点列表
	if _, ok := response["endpoints"]; !ok {
		t.Error("Expected endpoints field in root response")
	}
}

// TestHTTPServerIntegration 测试HTTP服务器集成
func TestHTTPServerIntegration(t *testing.T) {
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
	server := NewHTTPServer(syncer, 0) // 使用端口0让系统自动分配

	// 创建测试路由
	mux := http.NewServeMux()
	mux.HandleFunc("/", server.rootHandler)
	mux.HandleFunc("/health", server.healthHandler)
	mux.HandleFunc("/healthz", server.healthHandler)
	mux.HandleFunc("/ready", server.readinessHandler)
	mux.HandleFunc("/metrics", server.metricsHandler)
	mux.HandleFunc("/status", server.statusHandler)

	// 创建测试服务器
	testServer := httptest.NewServer(mux)
	defer testServer.Close()

	// 测试所有端点
	endpoints := []struct {
		path           string
		expectedStatus int
	}{
		{"/", http.StatusOK},
		{"/health", http.StatusOK},
		{"/healthz", http.StatusOK},
		{"/ready", http.StatusOK},
		{"/metrics", http.StatusOK},
		{"/status", http.StatusOK},
	}

	for _, endpoint := range endpoints {
		t.Run(endpoint.path, func(t *testing.T) {
			resp, err := http.Get(testServer.URL + endpoint.path)
			if err != nil {
				t.Fatalf("Failed to make request to %s: %v", endpoint.path, err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != endpoint.expectedStatus {
				t.Errorf("Expected status %d for %s, got %d", endpoint.expectedStatus, endpoint.path, resp.StatusCode)
			}

			// 验证Content-Type
			contentType := resp.Header.Get("Content-Type")
			if contentType != "application/json" {
				t.Errorf("Expected Content-Type 'application/json' for %s, got %s", endpoint.path, contentType)
			}
		})
	}
}

// TestResponseWriter 测试响应写入器
func TestResponseWriter(t *testing.T) {
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
	server := NewHTTPServer(syncer, 8080)

	// 创建测试处理器
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte("test response"))
	})

	// 包装处理器
	wrappedHandler := server.loggingMiddleware(handler)

	req := httptest.NewRequest("POST", "/test", nil)
	w := httptest.NewRecorder()

	wrappedHandler.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status %d, got %d", http.StatusCreated, w.Code)
	}

	if w.Body.String() != "test response" {
		t.Errorf("Expected body 'test response', got %s", w.Body.String())
	}
}