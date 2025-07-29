package cluster

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	clusterDomain "alert_agent/internal/domain/cluster"
	"go.uber.org/zap"
)

// ConfigSyncer 配置同步器
type ConfigSyncer struct {
	mu           sync.RWMutex
	clusters     map[string]*clusterDomain.Cluster
	syncStatus   map[string]*clusterDomain.SyncStatus
	templates    map[string]*clusterDomain.ConfigTemplate
	logger       *zap.Logger
	syncInterval time.Duration
	stopCh       chan struct{}
	running      bool
}

// NewConfigSyncer 创建新的配置同步器
func NewConfigSyncer(logger *zap.Logger, syncInterval time.Duration) *ConfigSyncer {
	return &ConfigSyncer{
		clusters:     make(map[string]*clusterDomain.Cluster),
		syncStatus:   make(map[string]*clusterDomain.SyncStatus),
		templates:    make(map[string]*clusterDomain.ConfigTemplate),
		logger:       logger,
		syncInterval: syncInterval,
		stopCh:       make(chan struct{}),
		running:      false,
	}
}

// Start 启动配置同步器
func (cs *ConfigSyncer) Start(ctx context.Context) error {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	
	if cs.running {
		return fmt.Errorf("config syncer is already running")
	}
	
	cs.running = true
	cs.stopCh = make(chan struct{})
	
	go cs.syncLoop(ctx)
	
	cs.logger.Info("Config syncer started", zap.Duration("interval", cs.syncInterval))
	return nil
}

// Stop 停止配置同步器
func (cs *ConfigSyncer) Stop() error {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	
	if !cs.running {
		return fmt.Errorf("config syncer is not running")
	}
	
	close(cs.stopCh)
	cs.running = false
	
	cs.logger.Info("Config syncer stopped")
	return nil
}

// AddCluster 添加集群到同步器
func (cs *ConfigSyncer) AddCluster(cluster *clusterDomain.Cluster) {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	
	cs.clusters[cluster.ID] = cluster
	cs.syncStatus[cluster.ID] = &clusterDomain.SyncStatus{
		ClusterID:    cluster.ID,
		Status:       clusterDomain.SyncStatusPending,
		LastSync:     time.Time{},
		NextSync:     time.Now().Add(cs.syncInterval),
		Version:      "0",
		ConfigHash:   "",
		ErrorMessage: "",
		RetryCount:   0,
		SyncDetails:  make(map[string]interface{}),
	}
	
	cs.logger.Info("Cluster added to config syncer", zap.String("cluster_id", cluster.ID))
}

// RemoveCluster 从同步器移除集群
func (cs *ConfigSyncer) RemoveCluster(clusterID string) {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	
	delete(cs.clusters, clusterID)
	delete(cs.syncStatus, clusterID)
	
	cs.logger.Info("Cluster removed from config syncer", zap.String("cluster_id", clusterID))
}

// SyncConfig 同步配置到指定集群
func (cs *ConfigSyncer) SyncConfig(ctx context.Context, clusterID string, config interface{}) error {
	cs.mu.Lock()
	cluster, exists := cs.clusters[clusterID]
	if !exists {
		cs.mu.Unlock()
		return fmt.Errorf("cluster not found: %s", clusterID)
	}
	
	status := cs.syncStatus[clusterID]
	status.Status = clusterDomain.SyncStatusInProgress
	status.RetryCount++
	cs.mu.Unlock()
	
	// 执行同步逻辑
	err := cs.performSync(ctx, cluster, config)
	
	cs.mu.Lock()
	defer cs.mu.Unlock()
	
	if err != nil {
		status.Status = clusterDomain.SyncStatusFailed
		status.ErrorMessage = err.Error()
		cs.logger.Error("Config sync failed", 
			zap.String("cluster_id", clusterID),
			zap.Error(err))
		return err
	}
	
	status.Status = clusterDomain.SyncStatusSuccess
	status.LastSync = time.Now()
	status.NextSync = time.Now().Add(cs.syncInterval)
	status.ErrorMessage = ""
	
	// 计算配置哈希
	configBytes, _ := json.Marshal(config)
	status.ConfigHash = fmt.Sprintf("%x", configBytes)
	
	cs.logger.Info("Config sync completed", zap.String("cluster_id", clusterID))
	return nil
}

// BatchSyncConfig 批量同步配置
func (cs *ConfigSyncer) BatchSyncConfig(ctx context.Context, clusterIDs []string, config interface{}) error {
	errors := make([]error, 0)
	
	for _, clusterID := range clusterIDs {
		if err := cs.SyncConfig(ctx, clusterID, config); err != nil {
			errors = append(errors, fmt.Errorf("cluster %s: %w", clusterID, err))
		}
	}
	
	if len(errors) > 0 {
		return fmt.Errorf("batch sync failed for %d clusters: %v", len(errors), errors)
	}
	
	return nil
}

// GetSyncStatus 获取同步状态
func (cs *ConfigSyncer) GetSyncStatus(clusterID string) (*clusterDomain.SyncStatus, error) {
	cs.mu.RLock()
	defer cs.mu.RUnlock()
	
	status, exists := cs.syncStatus[clusterID]
	if !exists {
		return nil, fmt.Errorf("cluster not found: %s", clusterID)
	}
	
	// 返回状态副本
	return &clusterDomain.SyncStatus{
		ClusterID:    status.ClusterID,
		Status:       status.Status,
		LastSync:     status.LastSync,
		NextSync:     status.NextSync,
		Version:      status.Version,
		ConfigHash:   status.ConfigHash,
		ErrorMessage: status.ErrorMessage,
		RetryCount:   status.RetryCount,
		SyncDetails:  status.SyncDetails,
	}, nil
}

// ValidateConfig 验证配置
func (cs *ConfigSyncer) ValidateConfig(ctx context.Context, clusterType clusterDomain.ClusterType, config interface{}) error {
	// 基本配置验证
	if config == nil {
		return fmt.Errorf("config cannot be nil")
	}
	
	// 根据集群类型进行特定验证
	switch clusterType {
	case clusterDomain.ClusterTypeAlertmanager:
		return cs.validateAlertmanagerConfig(config)
	case clusterDomain.ClusterTypePrometheus:
		return cs.validatePrometheusConfig(config)
	default:
		cs.logger.Warn("Unknown cluster type, skipping validation", 
			zap.String("type", string(clusterType)))
		return nil
	}
}

// CreateConfigTemplate 创建配置模板
func (cs *ConfigSyncer) CreateConfigTemplate(template *clusterDomain.ConfigTemplate) error {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	
	if template.ID == "" {
		return fmt.Errorf("template ID cannot be empty")
	}
	
	template.CreatedAt = time.Now()
	template.UpdatedAt = time.Now()
	
	cs.templates[template.ID] = template
	
	cs.logger.Info("Config template created", 
		zap.String("template_id", template.ID),
		zap.String("name", template.Name))
	
	return nil
}

// RenderConfig 渲染配置模板
func (cs *ConfigSyncer) RenderConfig(templateID string, variables map[string]interface{}) (string, error) {
	cs.mu.RLock()
	template, exists := cs.templates[templateID]
	cs.mu.RUnlock()
	
	if !exists {
		return "", fmt.Errorf("template not found: %s", templateID)
	}
	
	// 简单的模板渲染实现
	renderedConfig := template.Template
	
	// 替换变量
	for key, value := range variables {
		placeholder := fmt.Sprintf("{{.%s}}", key)
		valueStr := fmt.Sprintf("%v", value)
		renderedConfig = fmt.Sprintf(renderedConfig, placeholder, valueStr)
	}
	
	return renderedConfig, nil
}

// ApplyTemplate 应用模板到集群
func (cs *ConfigSyncer) ApplyTemplate(ctx context.Context, clusterID string, templateID string, variables map[string]interface{}) error {
	// 渲染配置
	renderedConfig, err := cs.RenderConfig(templateID, variables)
	if err != nil {
		return fmt.Errorf("failed to render template: %w", err)
	}
	
	// 同步配置
	return cs.SyncConfig(ctx, clusterID, renderedConfig)
}

// syncLoop 同步循环
func (cs *ConfigSyncer) syncLoop(ctx context.Context) {
	ticker := time.NewTicker(cs.syncInterval)
	defer ticker.Stop()
	
	for {
		select {
		case <-ctx.Done():
			return
		case <-cs.stopCh:
			return
		case <-ticker.C:
			cs.performScheduledSync(ctx)
		}
	}
}

// performScheduledSync 执行定期同步
func (cs *ConfigSyncer) performScheduledSync(ctx context.Context) {
	cs.mu.RLock()
	clustersToSync := make([]string, 0)
	now := time.Now()
	
	for clusterID, status := range cs.syncStatus {
		if status.NextSync.Before(now) && status.Status != clusterDomain.SyncStatusInProgress {
			clustersToSync = append(clustersToSync, clusterID)
		}
	}
	cs.mu.RUnlock()
	
	// 执行同步
	for _, clusterID := range clustersToSync {
		go func(id string) {
			if err := cs.SyncConfig(ctx, id, nil); err != nil {
				cs.logger.Error("Scheduled sync failed", 
					zap.String("cluster_id", id),
					zap.Error(err))
			}
		}(clusterID)
	}
}

// performSync 执行实际的同步操作
func (cs *ConfigSyncer) performSync(ctx context.Context, cluster *clusterDomain.Cluster, config interface{}) error {
	// 模拟同步操作
	cs.logger.Info("Performing config sync", 
		zap.String("cluster_id", cluster.ID),
		zap.String("cluster_name", cluster.Name))
	
	// 这里应该实现实际的配置同步逻辑
	// 例如：通过HTTP API推送配置到Alertmanager
	
	// 模拟网络延迟
	time.Sleep(100 * time.Millisecond)
	
	return nil
}

// validateAlertmanagerConfig 验证Alertmanager配置
func (cs *ConfigSyncer) validateAlertmanagerConfig(config interface{}) error {
	// 实现Alertmanager配置验证逻辑
	cs.logger.Debug("Validating Alertmanager config")
	return nil
}

// validatePrometheusConfig 验证Prometheus配置
func (cs *ConfigSyncer) validatePrometheusConfig(config interface{}) error {
	// 实现Prometheus配置验证逻辑
	cs.logger.Debug("Validating Prometheus config")
	return nil
}

// GetAllSyncStatus 获取所有集群的同步状态
func (cs *ConfigSyncer) GetAllSyncStatus() map[string]*clusterDomain.SyncStatus {
	cs.mu.RLock()
	defer cs.mu.RUnlock()
	
	status := make(map[string]*clusterDomain.SyncStatus)
	for id, s := range cs.syncStatus {
		status[id] = &clusterDomain.SyncStatus{
			ClusterID:    s.ClusterID,
			Status:       s.Status,
			LastSync:     s.LastSync,
			NextSync:     s.NextSync,
			Version:      s.Version,
			ConfigHash:   s.ConfigHash,
			ErrorMessage: s.ErrorMessage,
			RetryCount:   s.RetryCount,
			SyncDetails:  s.SyncDetails,
		}
	}
	
	return status
}

// SetSyncInterval 设置同步间隔
func (cs *ConfigSyncer) SetSyncInterval(interval time.Duration) {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	
	cs.syncInterval = interval
	cs.logger.Info("Sync interval updated", zap.Duration("interval", interval))
}