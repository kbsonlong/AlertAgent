package sidecar

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"alert_agent/internal/pkg/logger"

	"go.uber.org/zap"
)

// Config Sidecar配置
type Config struct {
	AlertAgentEndpoint string        // AlertAgent API端点
	ClusterID         string        // 集群ID
	ConfigType        string        // 配置类型: prometheus, alertmanager, vmalert
	ConfigPath        string        // 配置文件路径
	ReloadURL         string        // 重载URL
	SyncInterval      time.Duration // 同步间隔
}

// ConfigSyncer 配置同步器
type ConfigSyncer struct {
	config         *Config
	lastConfigHash string
	httpClient     *http.Client
	integration    TargetIntegration
	healthMonitor  *HealthMonitor
}

// SyncStatus 同步状态
type SyncStatus struct {
	ClusterID    string `json:"cluster_id"`
	ConfigType   string `json:"config_type"`
	Status       string `json:"status"`
	SyncTime     int64  `json:"sync_time"`
	ErrorMessage string `json:"error_message,omitempty"`
	ConfigHash   string `json:"config_hash,omitempty"`
}

// NewConfigSyncer 创建新的配置同步器
func NewConfigSyncer(config *Config) *ConfigSyncer {
	// 创建目标系统集成
	factory := NewIntegrationFactory()
	integration, err := factory.CreateIntegration(config.ConfigType)
	if err != nil {
		logger.L.Fatal("Failed to create integration", 
			zap.String("config_type", config.ConfigType),
			zap.Error(err),
		)
	}

	syncer := &ConfigSyncer{
		config: config,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		integration: integration,
	}

	// 创建健康监控器
	syncer.healthMonitor = NewHealthMonitor(syncer, 8081)

	return syncer
}

// Start 启动配置同步器
func (cs *ConfigSyncer) Start(ctx context.Context) error {
	logger.L.Info("Starting config syncer",
		zap.String("cluster_id", cs.config.ClusterID),
		zap.String("config_type", cs.config.ConfigType),
		zap.Duration("sync_interval", cs.config.SyncInterval),
	)

	// 启动健康监控器
	if err := cs.healthMonitor.Start(ctx); err != nil {
		return fmt.Errorf("failed to start health monitor: %w", err)
	}

	// 启动时立即同步一次
	cs.healthMonitor.UpdateSyncStatus("syncing")
	if err := cs.syncConfigWithRetry(ctx); err != nil {
		logger.L.Error("Initial config sync failed", zap.Error(err))
		cs.healthMonitor.UpdateSyncError(err)
		// 不要因为初始同步失败就退出，继续定时同步
	}

	ticker := time.NewTicker(cs.config.SyncInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			logger.L.Info("Context cancelled, stopping config syncer")
			// 停止健康监控器
			cs.healthMonitor.Stop(ctx)
			return ctx.Err()
		case <-ticker.C:
			cs.healthMonitor.UpdateSyncStatus("syncing")
			if err := cs.syncConfigWithRetry(ctx); err != nil {
				logger.L.Error("Config sync failed", zap.Error(err))
				cs.healthMonitor.UpdateSyncError(err)
				// 同步失败时上报状态
				cs.reportSyncStatus(ctx, "failed", err.Error())
			}
		}
	}
}

// syncConfig 同步配置
func (cs *ConfigSyncer) syncConfig(ctx context.Context) error {
	logger.L.Debug("Starting config sync")

	// 1. 从AlertAgent拉取配置
	config, serverHash, err := cs.fetchConfig(ctx)
	if err != nil {
		return fmt.Errorf("failed to fetch config: %w", err)
	}

	// 2. 检查配置是否有变化
	if serverHash == cs.lastConfigHash {
		logger.L.Debug("Config unchanged, skipping sync",
			zap.String("hash", serverHash),
		)
		return nil
	}

	logger.L.Info("Config changed, starting sync",
		zap.String("old_hash", cs.lastConfigHash),
		zap.String("new_hash", serverHash),
	)

	// 3. 验证配置格式
	if err := cs.validateConfig(config); err != nil {
		return fmt.Errorf("config validation failed: %w", err)
	}

	// 4. 原子性写入配置文件
	if err := cs.writeConfigFile(config); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	// 5. 触发目标系统reload
	if err := cs.triggerReload(ctx); err != nil {
		return fmt.Errorf("failed to trigger reload: %w", err)
	}

	// 6. 更新hash并记录成功
	cs.lastConfigHash = serverHash
	logger.L.Info("Successfully synced config",
		zap.String("config_type", cs.config.ConfigType),
		zap.String("hash", serverHash),
	)

	// 7. 回调AlertAgent更新同步状态
	if err := cs.reportSyncStatus(ctx, "success", ""); err != nil {
		logger.L.Error("Failed to report sync status", zap.Error(err))
		// 不要因为状态上报失败而返回错误
	}

	return nil
}

// fetchConfig 从AlertAgent拉取配置
func (cs *ConfigSyncer) fetchConfig(ctx context.Context) ([]byte, string, error) {
	endpoint := fmt.Sprintf("%s/api/v1/config/sync?cluster_id=%s&type=%s",
		cs.config.AlertAgentEndpoint, cs.config.ClusterID, cs.config.ConfigType)

	req, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, "", fmt.Errorf("failed to create request: %w", err)
	}

	// 添加条件请求头，如果有上次的hash
	if cs.lastConfigHash != "" {
		req.Header.Set("If-None-Match", cs.lastConfigHash)
	}

	resp, err := cs.httpClient.Do(req)
	if err != nil {
		return nil, "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// 处理304 Not Modified
	if resp.StatusCode == http.StatusNotModified {
		return nil, cs.lastConfigHash, nil
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, "", fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	config, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", fmt.Errorf("failed to read response body: %w", err)
	}

	// 获取服务器返回的hash，如果没有则计算
	serverHash := resp.Header.Get("X-Config-Hash")
	if serverHash == "" {
		hash := sha256.Sum256(config)
		serverHash = fmt.Sprintf("%x", hash)
	}

	return config, serverHash, nil
}

// validateConfig 验证配置格式
func (cs *ConfigSyncer) validateConfig(config []byte) error {
	if len(config) == 0 {
		return fmt.Errorf("config is empty")
	}

	// 使用目标系统集成进行验证
	return cs.integration.ValidateConfig(config)
}

// writeConfigFile 原子性写入配置文件
func (cs *ConfigSyncer) writeConfigFile(config []byte) error {
	// 确保目录存在
	if err := os.MkdirAll(filepath.Dir(cs.config.ConfigPath), 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// 原子性写入：先写临时文件，再重命名
	tmpFile := cs.config.ConfigPath + ".tmp"
	if err := os.WriteFile(tmpFile, config, 0644); err != nil {
		return fmt.Errorf("failed to write temp file: %w", err)
	}

	if err := os.Rename(tmpFile, cs.config.ConfigPath); err != nil {
		// 清理临时文件
		os.Remove(tmpFile)
		return fmt.Errorf("failed to rename temp file: %w", err)
	}

	logger.L.Debug("Config file written successfully",
		zap.String("path", cs.config.ConfigPath),
		zap.Int("size", len(config)),
	)

	return nil
}

// triggerReload 触发目标系统重载
func (cs *ConfigSyncer) triggerReload(ctx context.Context) error {
	// 使用目标系统集成进行重载
	return cs.integration.TriggerReload(ctx, cs.config.ReloadURL)
}

// reportSyncStatus 上报同步状态
func (cs *ConfigSyncer) reportSyncStatus(ctx context.Context, status, errorMsg string) error {
	endpoint := fmt.Sprintf("%s/api/v1/config/sync/status", cs.config.AlertAgentEndpoint)

	syncStatus := SyncStatus{
		ClusterID:    cs.config.ClusterID,
		ConfigType:   cs.config.ConfigType,
		Status:       status,
		SyncTime:     time.Now().Unix(),
		ErrorMessage: errorMsg,
		ConfigHash:   cs.lastConfigHash,
	}

	jsonData, err := json.Marshal(syncStatus)
	if err != nil {
		return fmt.Errorf("failed to marshal sync status: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create status request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := cs.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send status request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("status report failed with status %d: %s", resp.StatusCode, string(body))
	}

	logger.L.Debug("Successfully reported sync status",
		zap.String("status", status),
		zap.String("cluster_id", cs.config.ClusterID),
		zap.String("config_type", cs.config.ConfigType),
	)

	return nil
}

// GetLastConfigHash 获取最后的配置hash
func (cs *ConfigSyncer) GetLastConfigHash() string {
	return cs.lastConfigHash
}

// ForceSync 强制同步配置
func (cs *ConfigSyncer) ForceSync(ctx context.Context) error {
	// 清空hash强制同步
	oldHash := cs.lastConfigHash
	cs.lastConfigHash = ""
	
	err := cs.syncConfig(ctx)
	if err != nil {
		// 恢复原hash
		cs.lastConfigHash = oldHash
	}
	
	return err
}

// syncConfigWithRetry 带重试的配置同步
func (cs *ConfigSyncer) syncConfigWithRetry(ctx context.Context) error {
	var lastErr error
	
	for attempt := 0; attempt < 5; attempt++ { // 最多重试5次
		err := cs.syncConfig(ctx)
		if err == nil {
			// 同步成功
			hash := cs.lastConfigHash
			size := int64(0) // 这里应该记录实际的配置大小
			cs.healthMonitor.UpdateSyncSuccess(hash, size)
			return nil
		}
		
		lastErr = err
		
		// 检查是否应该重试
		if !cs.healthMonitor.ShouldRetry(err, attempt) {
			break
		}
		
		// 计算重试延迟
		delay := cs.healthMonitor.GetRetryDelay(attempt)
		logger.L.Warn("Config sync failed, retrying",
			zap.Error(err),
			zap.Int("attempt", attempt+1),
			zap.Duration("delay", delay),
		)
		
		// 等待重试延迟
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(delay):
			// 继续重试
		}
	}
	
	// 所有重试都失败了
	cs.healthMonitor.UpdateSyncError(lastErr)
	return fmt.Errorf("sync failed after retries: %w", lastErr)
}