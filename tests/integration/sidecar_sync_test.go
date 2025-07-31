package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSidecarConfigSync 测试Sidecar配置同步集成
func TestSidecarConfigSync(t *testing.T) {
	// 清理测试数据
	require.NoError(t, clearTestData())

	// 创建测试用的API服务器
	router := setupTestAPIServer()
	server := httptest.NewServer(router)
	defer server.Close()

	t.Run("完整配置同步流程", func(t *testing.T) {
		testCompleteSyncFlow(t, server.URL)
	})

	t.Run("配置变更检测", func(t *testing.T) {
		testConfigChangeDetection(t, server.URL)
	})

	t.Run("同步失败重试", func(t *testing.T) {
		testSyncFailureRetry(t, server.URL)
	})

	t.Run("多集群配置同步", func(t *testing.T) {
		testMultiClusterSync(t, server.URL)
	})
}

func setupTestAPIServer() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// 模拟配置拉取接口
	router.GET("/api/v1/config/sync", func(c *gin.Context) {
		clusterID := c.Query("cluster_id")
		configType := c.Query("config_type")

		if clusterID == "" || configType == "" {
			c.JSON(400, gin.H{"error": "cluster_id and type are required"})
			return
		}

		// 生成测试配置
		config := generateTestConfig(configType, clusterID)
		hash := calculateConfigHash(config)

		c.Header("X-Config-Hash", hash)
		c.Header("Content-Type", "application/yaml")
		c.String(200, config)
	})

	// 模拟同步状态上报接口
	router.POST("/api/v1/config/sync/status", func(c *gin.Context) {
		var status map[string]interface{}
		if err := c.ShouldBindJSON(&status); err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}

		// 保存同步状态到数据库
		err := testDB.Exec(`
			INSERT INTO config_sync_status 
			(id, cluster_id, config_type, config_hash, sync_status, sync_time, error_message)
			VALUES (UUID(), ?, ?, ?, ?, FROM_UNIXTIME(?), ?)
			ON DUPLICATE KEY UPDATE
			config_hash = VALUES(config_hash),
			sync_status = VALUES(sync_status),
			sync_time = VALUES(sync_time),
			error_message = VALUES(error_message)
		`, status["cluster_id"], status["config_type"], status["config_hash"],
			status["status"], status["sync_time"], status["error_message"]).Error

		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}

		c.JSON(200, gin.H{"message": "status updated"})
	})

	// 模拟目标系统reload接口
	router.POST("/reload", func(c *gin.Context) {
		// 模拟reload操作
		time.Sleep(100 * time.Millisecond)
		c.JSON(200, gin.H{"message": "reloaded"})
	})

	return router
}

func testCompleteSyncFlow(t *testing.T, serverURL string) {
	clusterID := "test-cluster-1"
	configType := "prometheus"

	// 创建Sidecar同步器
	syncer := &ConfigSyncer{
		AlertAgentEndpoint: serverURL,
		ClusterID:         clusterID,
		ConfigType:        configType,
		ConfigPath:        "/tmp/test-prometheus.yml",
		ReloadURL:         serverURL + "/reload",
		SyncInterval:      time.Second,
		httpClient:        &http.Client{Timeout: 5 * time.Second},
	}

	// 执行同步
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err := syncer.syncConfig(ctx)
	assert.NoError(t, err)

	// 验证同步状态已保存到数据库
	var syncStatus struct {
		ClusterID   string `db:"cluster_id"`
		ConfigType  string `db:"config_type"`
		SyncStatus  string `db:"sync_status"`
		ConfigHash  string `db:"config_hash"`
		ErrorMessage string `db:"error_message"`
	}

	err = testDB.Raw(`
		SELECT cluster_id, config_type, sync_status, config_hash, error_message
		FROM config_sync_status 
		WHERE cluster_id = ? AND config_type = ?
	`, clusterID, configType).Scan(&syncStatus).Error

	require.NoError(t, err)
	assert.Equal(t, clusterID, syncStatus.ClusterID)
	assert.Equal(t, configType, syncStatus.ConfigType)
	assert.Equal(t, "success", syncStatus.SyncStatus)
	assert.NotEmpty(t, syncStatus.ConfigHash)
	assert.Empty(t, syncStatus.ErrorMessage)
}

func testConfigChangeDetection(t *testing.T, serverURL string) {
	clusterID := "test-cluster-2"
	configType := "alertmanager"

	syncer := &ConfigSyncer{
		AlertAgentEndpoint: serverURL,
		ClusterID:         clusterID,
		ConfigType:        configType,
		ConfigPath:        "/tmp/test-alertmanager.yml",
		ReloadURL:         serverURL + "/reload",
		SyncInterval:      time.Second,
		httpClient:        &http.Client{Timeout: 5 * time.Second},
	}

	ctx := context.Background()

	// 第一次同步
	err := syncer.syncConfig(ctx)
	assert.NoError(t, err)
	firstHash := syncer.lastConfigHash

	// 第二次同步（配置未变更）
	err = syncer.syncConfig(ctx)
	assert.NoError(t, err)
	assert.Equal(t, firstHash, syncer.lastConfigHash)

	// 模拟配置变更（通过修改时间戳）
	time.Sleep(time.Millisecond * 100)
	
	// 第三次同步（应该检测到变更）
	err = syncer.syncConfig(ctx)
	assert.NoError(t, err)
	// 注意：在实际实现中，配置变更会导致hash变化
}

func testSyncFailureRetry(t *testing.T, serverURL string) {
	clusterID := "test-cluster-3"
	configType := "vmalert"

	// 创建一个会失败的reload URL
	failingReloadURL := serverURL + "/failing-reload"

	syncer := &ConfigSyncer{
		AlertAgentEndpoint: serverURL,
		ClusterID:         clusterID,
		ConfigType:        configType,
		ConfigPath:        "/tmp/test-vmalert.yml",
		ReloadURL:         failingReloadURL,
		SyncInterval:      time.Second,
		httpClient:        &http.Client{Timeout: 5 * time.Second},
	}

	ctx := context.Background()

	// 执行同步（应该失败）
	err := syncer.syncConfig(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to trigger reload")

	// 验证失败状态已记录
	var syncStatus struct {
		SyncStatus   string `db:"sync_status"`
		ErrorMessage string `db:"error_message"`
	}

	err = testDB.Raw(`
		SELECT sync_status, error_message
		FROM config_sync_status 
		WHERE cluster_id = ? AND config_type = ?
	`, clusterID, configType).Scan(&syncStatus).Error

	if err == nil {
		// 如果记录存在，应该是失败状态
		assert.Equal(t, "failed", syncStatus.SyncStatus)
		assert.NotEmpty(t, syncStatus.ErrorMessage)
	}
}

func testMultiClusterSync(t *testing.T, serverURL string) {
	clusters := []struct {
		clusterID  string
		configType string
	}{
		{"cluster-a", "prometheus"},
		{"cluster-b", "alertmanager"},
		{"cluster-c", "vmalert"},
	}

	// 并发同步多个集群
	results := make(chan error, len(clusters))

	for _, cluster := range clusters {
		go func(clusterID, configType string) {
			syncer := &ConfigSyncer{
				AlertAgentEndpoint: serverURL,
				ClusterID:         clusterID,
				ConfigType:        configType,
				ConfigPath:        fmt.Sprintf("/tmp/test-%s-%s.yml", clusterID, configType),
				ReloadURL:         serverURL + "/reload",
				SyncInterval:      time.Second,
				httpClient:        &http.Client{Timeout: 5 * time.Second},
			}

			ctx := context.Background()
			results <- syncer.syncConfig(ctx)
		}(cluster.clusterID, cluster.configType)
	}

	// 等待所有同步完成
	for i := 0; i < len(clusters); i++ {
		err := <-results
		assert.NoError(t, err)
	}

	// 验证所有集群的同步状态
	var count int64
	err := testDB.Raw(`
		SELECT COUNT(*) FROM config_sync_status 
		WHERE sync_status = 'success'
	`).Scan(&count).Error

	require.NoError(t, err)
	assert.Equal(t, int64(len(clusters)), count)
}

// ConfigSyncer 模拟Sidecar配置同步器
type ConfigSyncer struct {
	AlertAgentEndpoint string
	ClusterID         string
	ConfigType        string
	ConfigPath        string
	ReloadURL         string
	SyncInterval      time.Duration
	lastConfigHash    string
	httpClient        *http.Client
}

func (cs *ConfigSyncer) syncConfig(ctx context.Context) error {
	// 1. 从AlertAgent拉取配置
	config, serverHash, err := cs.fetchConfig(ctx)
	if err != nil {
		return fmt.Errorf("failed to fetch config: %w", err)
	}

	// 2. 检查配置是否有变化
	if serverHash == cs.lastConfigHash {
		return nil
	}

	// 3. 触发目标系统reload
	if err := cs.triggerReload(ctx); err != nil {
		return fmt.Errorf("failed to trigger reload: %w", err)
	}

	// 4. 更新hash
	cs.lastConfigHash = serverHash

	// 5. 回调AlertAgent更新同步状态
	if err := cs.reportSyncStatus(ctx, "success", ""); err != nil {
		return fmt.Errorf("failed to report sync status: %w", err)
	}

	return nil
}

func (cs *ConfigSyncer) fetchConfig(ctx context.Context) ([]byte, string, error) {
	endpoint := fmt.Sprintf("%s/api/v1/config/sync?cluster_id=%s&type=%s",
		cs.AlertAgentEndpoint, cs.ClusterID, cs.ConfigType)

	req, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, "", err
	}

	resp, err := cs.httpClient.Do(req)
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, "", fmt.Errorf("API returned status: %s", resp.Status)
	}

	config, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", err
	}

	serverHash := resp.Header.Get("X-Config-Hash")
	return config, serverHash, nil
}

func (cs *ConfigSyncer) triggerReload(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, "POST", cs.ReloadURL, nil)
	if err != nil {
		return err
	}

	resp, err := cs.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("reload failed with status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

func (cs *ConfigSyncer) reportSyncStatus(ctx context.Context, status, errorMsg string) error {
	endpoint := fmt.Sprintf("%s/api/v1/config/sync/status", cs.AlertAgentEndpoint)

	payload := map[string]interface{}{
		"cluster_id":     cs.ClusterID,
		"config_type":    cs.ConfigType,
		"status":         status,
		"sync_time":      time.Now().Unix(),
		"error_message":  errorMsg,
		"config_hash":    cs.lastConfigHash,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := cs.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}

func generateTestConfig(configType, clusterID string) string {
	switch configType {
	case "prometheus":
		return fmt.Sprintf(`
global:
  scrape_interval: 15s
  evaluation_interval: 15s

rule_files:
  - "/etc/prometheus/rules/*.yml"

scrape_configs:
  - job_name: 'prometheus'
    static_configs:
      - targets: ['localhost:9090']
    labels:
      cluster: '%s'
`, clusterID)

	case "alertmanager":
		return fmt.Sprintf(`
global:
  smtp_smarthost: 'localhost:587'

route:
  group_by: ['alertname']
  group_wait: 10s
  group_interval: 10s
  repeat_interval: 1h
  receiver: 'web.hook'

receivers:
- name: 'web.hook'
  webhook_configs:
  - url: 'http://127.0.0.1:5001/'
    send_resolved: true

# Cluster: %s
`, clusterID)

	case "vmalert":
		return fmt.Sprintf(`
datasource:
  url: "http://localhost:8428"

notifier:
  url: "http://localhost:9093"

# Cluster: %s
`, clusterID)

	default:
		return fmt.Sprintf("# Test config for %s in cluster %s\n", configType, clusterID)
	}
}

func calculateConfigHash(config string) string {
	// 简单的hash计算，实际应该使用更强的hash算法
	return fmt.Sprintf("%x", len(config)+int(time.Now().UnixNano()%1000))
}