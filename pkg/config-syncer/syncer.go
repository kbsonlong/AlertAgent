package configsyncer

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"go.uber.org/zap"
)

// Config 配置同步器配置
type Config struct {
	AlertAgentEndpoint string        // AlertAgent API端点
	ClusterID         string        // 集群ID
	ConfigType        string        // 配置类型: "prometheus" 或 "alertmanager"
	ConfigPath        string        // 配置文件路径
	ReloadURL         string        // 重载URL
	SyncInterval      time.Duration // 同步间隔
	Logger            *zap.Logger   // 日志器
	HTTPTimeout       time.Duration // HTTP超时时间
	MaxRetries        int           // 最大重试次数
	RetryBackoff      time.Duration // 重试退避时间
}

// ConfigSyncer 配置同步器
type ConfigSyncer struct {
	config         *Config
	lastConfigHash string
	httpClient     *http.Client
	retryCount     int
	startTime      time.Time
	lastSyncTime   time.Time
	syncCount      int64
	successCount   int64
	failureCount   int64
	lastError      string
	configVersion  string
	healthy        bool
}

// ConfigResponse AlertAgent配置响应
type ConfigResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	Data    struct {
		Config     string `json:"config"`
		ConfigHash string `json:"config_hash"`
		Version    string `json:"version"`
	} `json:"data"`
}

// SyncMetrics 同步指标
type SyncMetrics struct {
	LastSyncTime    time.Time `json:"last_sync_time"`
	SyncCount       int64     `json:"sync_count"`
	SuccessCount    int64     `json:"success_count"`
	FailureCount    int64     `json:"failure_count"`
	LastError       string    `json:"last_error,omitempty"`
	ConfigHash      string    `json:"config_hash"`
	ConfigVersion   string    `json:"config_version"`
	RetryCount      int       `json:"retry_count"`
	NextSyncTime    time.Time `json:"next_sync_time"`
	Healthy         bool      `json:"healthy"`
	Uptime          time.Duration `json:"uptime"`
}

// HealthStatus 健康状态
type HealthStatus struct {
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
	Uptime    time.Duration `json:"uptime"`
	Version   string    `json:"version"`
	ClusterID string    `json:"cluster_id"`
	ConfigType string   `json:"config_type"`
	LastSync  time.Time `json:"last_sync,omitempty"`
	Error     string    `json:"error,omitempty"`
}

// NewConfigSyncer 创建新的配置同步器
func NewConfigSyncer(config *Config) *ConfigSyncer {
	// 设置默认值
	if config.HTTPTimeout == 0 {
		config.HTTPTimeout = 30 * time.Second
	}
	if config.MaxRetries == 0 {
		config.MaxRetries = 3
	}
	if config.RetryBackoff == 0 {
		config.RetryBackoff = 5 * time.Second
	}

	return &ConfigSyncer{
		config: config,
		httpClient: &http.Client{
			Timeout: config.HTTPTimeout,
		},
		startTime: time.Now(),
		healthy:   true,
	}
}

// Start 启动配置同步器
func (cs *ConfigSyncer) Start(ctx context.Context) error {
	cs.config.Logger.Info("Starting config syncer",
		zap.String("cluster_id", cs.config.ClusterID),
		zap.String("config_type", cs.config.ConfigType),
		zap.Duration("sync_interval", cs.config.SyncInterval))

	// 启动时立即同步一次
	if err := cs.syncConfig(ctx); err != nil {
		cs.config.Logger.Error("Initial config sync failed", zap.Error(err))
		// 初始同步失败不应该阻止启动，继续定期尝试
	}

	ticker := time.NewTicker(cs.config.SyncInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			cs.config.Logger.Info("Context cancelled, stopping config syncer")
			return ctx.Err()
		case <-ticker.C:
			if err := cs.syncConfig(ctx); err != nil {
				cs.config.Logger.Error("Config sync failed", zap.Error(err))
			}
		}
	}
}

// syncConfig 同步配置
func (cs *ConfigSyncer) syncConfig(ctx context.Context) error {
	cs.config.Logger.Debug("Starting config sync")
	cs.syncCount++

	// 1. 从AlertAgent拉取配置
	configResp, err := cs.fetchConfigWithRetry(ctx)
	if err != nil {
		cs.failureCount++
		cs.lastError = err.Error()
		cs.healthy = false
		return fmt.Errorf("failed to fetch config: %w", err)
	}

	// 2. 检查配置是否有变化
	if !cs.hasConfigChanged(configResp.Data.ConfigHash) {
		cs.config.Logger.Debug("Config unchanged, skipping sync")
		cs.successCount++
		cs.lastSyncTime = time.Now()
		cs.healthy = true
		cs.lastError = ""
		return nil
	}

	// 3. 原子性写入配置文件
	if err := cs.writeConfigFileAtomic(configResp.Data.Config); err != nil {
		cs.failureCount++
		cs.lastError = err.Error()
		cs.healthy = false
		return fmt.Errorf("failed to write config: %w", err)
	}

	// 4. 触发热重载
	if err := cs.triggerReload(ctx); err != nil {
		cs.failureCount++
		cs.lastError = err.Error()
		cs.healthy = false
		return fmt.Errorf("failed to trigger reload: %w", err)
	}

	// 5. 更新本地状态
	cs.lastConfigHash = configResp.Data.ConfigHash
	cs.configVersion = configResp.Data.Version
	cs.retryCount = 0
	cs.successCount++
	cs.lastSyncTime = time.Now()
	cs.healthy = true
	cs.lastError = ""

	cs.config.Logger.Info("Config sync completed successfully",
		zap.String("config_hash", configResp.Data.ConfigHash),
		zap.String("version", configResp.Data.Version))

	return nil
}

// fetchConfigWithRetry 带重试的配置拉取
func (cs *ConfigSyncer) fetchConfigWithRetry(ctx context.Context) (*ConfigResponse, error) {
	var lastErr error

	for attempt := 0; attempt <= cs.config.MaxRetries; attempt++ {
		if attempt > 0 {
			// 指数退避重试
			backoff := time.Duration(math.Pow(2, float64(attempt-1))) * cs.config.RetryBackoff
			cs.config.Logger.Warn("Retrying config fetch",
				zap.Int("attempt", attempt),
				zap.Duration("backoff", backoff))

			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(backoff):
			}
		}

		configResp, err := cs.fetchConfig(ctx)
		if err == nil {
			return configResp, nil
		}

		lastErr = err
		cs.config.Logger.Error("Config fetch attempt failed",
			zap.Int("attempt", attempt+1),
			zap.Error(err))
	}

	cs.retryCount++
	return nil, fmt.Errorf("failed to fetch config after %d attempts: %w", cs.config.MaxRetries+1, lastErr)
}

// fetchConfig 从AlertAgent拉取配置
func (cs *ConfigSyncer) fetchConfig(ctx context.Context) (*ConfigResponse, error) {
	url := fmt.Sprintf("%s/api/v1/configs?cluster_id=%s&type=%s",
		cs.config.AlertAgentEndpoint, cs.config.ClusterID, cs.config.ConfigType)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", fmt.Sprintf("config-syncer/%s", cs.config.ConfigType))

	resp, err := cs.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("HTTP request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var configResp ConfigResponse
	if err := json.NewDecoder(resp.Body).Decode(&configResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if configResp.Status != "success" {
		return nil, fmt.Errorf("API returned error: %s", configResp.Message)
	}

	return &configResp, nil
}

// hasConfigChanged 检查配置是否有变化
func (cs *ConfigSyncer) hasConfigChanged(newHash string) bool {
	return cs.lastConfigHash != newHash
}

// writeConfigFileAtomic 原子性写入配置文件
func (cs *ConfigSyncer) writeConfigFileAtomic(config string) error {
	// 创建临时文件
	tempFile := cs.config.ConfigPath + ".tmp"

	// 确保目录存在
	dir := filepath.Dir(cs.config.ConfigPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// 写入临时文件
	if err := os.WriteFile(tempFile, []byte(config), 0644); err != nil {
		return fmt.Errorf("failed to write temp file: %w", err)
	}

	// 原子性重命名
	if err := os.Rename(tempFile, cs.config.ConfigPath); err != nil {
		// 清理临时文件
		os.Remove(tempFile)
		return fmt.Errorf("failed to rename temp file: %w", err)
	}

	cs.config.Logger.Debug("Config file written atomically",
		zap.String("path", cs.config.ConfigPath),
		zap.Int("size", len(config)))

	return nil
}

// triggerReload 触发热重载
func (cs *ConfigSyncer) triggerReload(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, "POST", cs.config.ReloadURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create reload request: %w", err)
	}

	resp, err := cs.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("reload request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("reload failed with status %d: %s", resp.StatusCode, string(body))
	}

	cs.config.Logger.Info("Hot reload triggered successfully",
		zap.String("reload_url", cs.config.ReloadURL))

	return nil
}

// GetMetrics 获取同步指标
func (cs *ConfigSyncer) GetMetrics() *SyncMetrics {
	return &SyncMetrics{
		LastSyncTime:  cs.lastSyncTime,
		SyncCount:     cs.syncCount,
		SuccessCount:  cs.successCount,
		FailureCount:  cs.failureCount,
		LastError:     cs.lastError,
		ConfigHash:    cs.lastConfigHash,
		ConfigVersion: cs.configVersion,
		RetryCount:    cs.retryCount,
		NextSyncTime:  time.Now().Add(cs.config.SyncInterval),
		Healthy:       cs.healthy,
		Uptime:        time.Since(cs.startTime),
	}
}

// GetHealthStatus 获取健康状态
func (cs *ConfigSyncer) GetHealthStatus() *HealthStatus {
	status := "healthy"
	if !cs.healthy {
		status = "unhealthy"
	}

	return &HealthStatus{
		Status:     status,
		Timestamp:  time.Now(),
		Uptime:     time.Since(cs.startTime),
		Version:    "1.0.0", // 可以从构建时注入
		ClusterID:  cs.config.ClusterID,
		ConfigType: cs.config.ConfigType,
		LastSync:   cs.lastSyncTime,
		Error:      cs.lastError,
	}
}

// IsHealthy 检查是否健康
func (cs *ConfigSyncer) IsHealthy() bool {
	return cs.healthy
}

// calculateHash 计算配置哈希
func (cs *ConfigSyncer) calculateHash(data []byte) string {
	h := sha256.Sum256(data)
	return hex.EncodeToString(h[:])
}

// ValidateConfig 验证配置
func (cs *ConfigSyncer) ValidateConfig() error {
	if cs.config.AlertAgentEndpoint == "" {
		return fmt.Errorf("AlertAgent endpoint is required")
	}
	if cs.config.ClusterID == "" {
		return fmt.Errorf("cluster ID is required")
	}
	if cs.config.ConfigType == "" {
		return fmt.Errorf("config type is required")
	}
	if cs.config.ConfigPath == "" {
		return fmt.Errorf("config path is required")
	}
	if cs.config.ReloadURL == "" {
		return fmt.Errorf("reload URL is required")
	}
	if cs.config.SyncInterval <= 0 {
		return fmt.Errorf("sync interval must be positive")
	}

	return nil
}