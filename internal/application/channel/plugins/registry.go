package plugins

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"
)

// PluginRegistry 插件注册表
type PluginRegistry struct {
	plugins map[string]NotificationPlugin
	configs map[string]*PluginConfig
	stats   map[string]*PluginStats
	mutex   sync.RWMutex
	logger  *zap.Logger
}

// NewPluginRegistry 创建插件注册表
func NewPluginRegistry(logger *zap.Logger) *PluginRegistry {
	return &PluginRegistry{
		plugins: make(map[string]NotificationPlugin),
		configs: make(map[string]*PluginConfig),
		stats:   make(map[string]*PluginStats),
		logger:  logger,
	}
}

// RegisterPlugin 注册插件
func (pr *PluginRegistry) RegisterPlugin(plugin NotificationPlugin) error {
	pr.mutex.Lock()
	defer pr.mutex.Unlock()

	name := plugin.Name()
	if _, exists := pr.plugins[name]; exists {
		return fmt.Errorf("plugin %s already registered", name)
	}

	// 初始化插件
	if err := plugin.Initialize(); err != nil {
		return fmt.Errorf("failed to initialize plugin %s: %w", name, err)
	}

	pr.plugins[name] = plugin
	pr.stats[name] = &PluginStats{
		Name: name,
	}

	pr.logger.Info("Plugin registered successfully",
		zap.String("name", name),
		zap.String("version", plugin.Version()),
		zap.String("description", plugin.Description()))

	return nil
}

// UnregisterPlugin 注销插件
func (pr *PluginRegistry) UnregisterPlugin(name string) error {
	pr.mutex.Lock()
	defer pr.mutex.Unlock()

	plugin, exists := pr.plugins[name]
	if !exists {
		return fmt.Errorf("plugin %s not found", name)
	}

	// 关闭插件
	if err := plugin.Shutdown(); err != nil {
		pr.logger.Error("Failed to shutdown plugin", zap.String("name", name), zap.Error(err))
	}

	delete(pr.plugins, name)
	delete(pr.configs, name)
	delete(pr.stats, name)

	pr.logger.Info("Plugin unregistered", zap.String("name", name))
	return nil
}

// GetPlugin 获取插件
func (pr *PluginRegistry) GetPlugin(name string) (NotificationPlugin, bool) {
	pr.mutex.RLock()
	defer pr.mutex.RUnlock()

	plugin, exists := pr.plugins[name]
	return plugin, exists
}

// GetAvailablePlugins 获取所有可用插件信息
func (pr *PluginRegistry) GetAvailablePlugins() []PluginInfo {
	pr.mutex.RLock()
	defer pr.mutex.RUnlock()

	var plugins []PluginInfo
	for name, plugin := range pr.plugins {
		config := pr.configs[name]
		status := "inactive"
		if config != nil && config.Enabled {
			status = "active"
		}

		plugins = append(plugins, PluginInfo{
			Name:        name,
			Version:     plugin.Version(),
			Description: plugin.Description(),
			Schema:      plugin.ConfigSchema(),
			Status:      status,
			LoadTime:    time.Now(), // TODO: 记录实际加载时间
		})
	}

	return plugins
}

// SetPluginConfig 设置插件配置
func (pr *PluginRegistry) SetPluginConfig(name string, config *PluginConfig) error {
	pr.mutex.Lock()
	defer pr.mutex.Unlock()

	plugin, exists := pr.plugins[name]
	if !exists {
		return fmt.Errorf("plugin %s not found", name)
	}

	// 验证配置
	if err := plugin.ValidateConfig(config.Config); err != nil {
		return fmt.Errorf("invalid config for plugin %s: %w", name, err)
	}

	pr.configs[name] = config
	pr.logger.Info("Plugin config updated", zap.String("name", name), zap.Bool("enabled", config.Enabled))
	return nil
}

// GetPluginConfig 获取插件配置
func (pr *PluginRegistry) GetPluginConfig(name string) (*PluginConfig, bool) {
	pr.mutex.RLock()
	defer pr.mutex.RUnlock()

	config, exists := pr.configs[name]
	return config, exists
}

// SendNotification 发送通知
func (pr *PluginRegistry) SendNotification(ctx context.Context, pluginName string, message *NotificationMessage) (*SendResult, error) {
	pr.mutex.RLock()
	plugin, pluginExists := pr.plugins[pluginName]
	config, configExists := pr.configs[pluginName]
	stats := pr.stats[pluginName]
	pr.mutex.RUnlock()

	if !pluginExists {
		return nil, fmt.Errorf("plugin %s not found", pluginName)
	}

	if !configExists || !config.Enabled {
		return nil, fmt.Errorf("plugin %s is not configured or disabled", pluginName)
	}

	startTime := time.Now()
	result := &SendResult{
		Timestamp: startTime,
	}

	// 发送通知
	err := plugin.Send(ctx, config.Config, message)
	result.Duration = time.Since(startTime)

	// 更新统计信息
	pr.mutex.Lock()
	stats.TotalSent++
	stats.LastSent = startTime
	if err != nil {
		stats.FailureCount++
		stats.LastError = err.Error()
		result.Success = false
		result.Error = err.Error()
	} else {
		stats.SuccessCount++
		result.Success = true
	}
	// 更新平均耗时
	if stats.TotalSent > 0 {
		stats.AvgDuration = time.Duration((int64(stats.AvgDuration)*stats.TotalSent + int64(result.Duration)) / (stats.TotalSent + 1))
	}
	pr.mutex.Unlock()

	pr.logger.Info("Notification sent",
		zap.String("plugin", pluginName),
		zap.String("alert_id", message.AlertID),
		zap.Bool("success", result.Success),
		zap.Duration("duration", result.Duration),
		zap.Error(err))

	return result, err
}

// SendNotificationToAll 向所有启用的插件发送通知
func (pr *PluginRegistry) SendNotificationToAll(ctx context.Context, message *NotificationMessage) map[string]*SendResult {
	pr.mutex.RLock()
	enabledPlugins := make([]string, 0)
	for name, config := range pr.configs {
		if config.Enabled {
			enabledPlugins = append(enabledPlugins, name)
		}
	}
	pr.mutex.RUnlock()

	results := make(map[string]*SendResult)
	for _, pluginName := range enabledPlugins {
		result, err := pr.SendNotification(ctx, pluginName, message)
		if err != nil {
			result = &SendResult{
				Success:   false,
				Error:     err.Error(),
				Timestamp: time.Now(),
			}
		}
		results[pluginName] = result
	}

	return results
}

// HealthCheckPlugin 检查插件健康状态
func (pr *PluginRegistry) HealthCheckPlugin(ctx context.Context, pluginName string) error {
	pr.mutex.RLock()
	plugin, pluginExists := pr.plugins[pluginName]
	config, configExists := pr.configs[pluginName]
	pr.mutex.RUnlock()

	if !pluginExists {
		return fmt.Errorf("plugin %s not found", pluginName)
	}

	if !configExists {
		return fmt.Errorf("plugin %s is not configured", pluginName)
	}

	return plugin.HealthCheck(ctx, config.Config)
}

// HealthCheckAll 检查所有插件健康状态
func (pr *PluginRegistry) HealthCheckAll(ctx context.Context) map[string]error {
	pr.mutex.RLock()
	pluginNames := make([]string, 0, len(pr.plugins))
	for name := range pr.plugins {
		pluginNames = append(pluginNames, name)
	}
	pr.mutex.RUnlock()

	results := make(map[string]error)
	for _, name := range pluginNames {
		results[name] = pr.HealthCheckPlugin(ctx, name)
	}

	return results
}

// GetPluginStats 获取插件统计信息
func (pr *PluginRegistry) GetPluginStats(pluginName string) (*PluginStats, bool) {
	pr.mutex.RLock()
	defer pr.mutex.RUnlock()

	stats, exists := pr.stats[pluginName]
	if !exists {
		return nil, false
	}

	// 返回副本避免并发修改
	statsCopy := *stats
	return &statsCopy, true
}

// GetAllPluginStats 获取所有插件统计信息
func (pr *PluginRegistry) GetAllPluginStats() map[string]*PluginStats {
	pr.mutex.RLock()
	defer pr.mutex.RUnlock()

	results := make(map[string]*PluginStats)
	for name, stats := range pr.stats {
		statsCopy := *stats
		results[name] = &statsCopy
	}

	return results
}

// ValidatePluginConfig 验证插件配置
func (pr *PluginRegistry) ValidatePluginConfig(pluginName string, config map[string]interface{}) error {
	pr.mutex.RLock()
	plugin, exists := pr.plugins[pluginName]
	pr.mutex.RUnlock()

	if !exists {
		return fmt.Errorf("plugin %s not found", pluginName)
	}

	return plugin.ValidateConfig(config)
}

// LoadPluginConfigs 从JSON加载插件配置
func (pr *PluginRegistry) LoadPluginConfigs(configData []byte) error {
	var configs map[string]*PluginConfig
	if err := json.Unmarshal(configData, &configs); err != nil {
		return fmt.Errorf("failed to unmarshal plugin configs: %w", err)
	}

	for name, config := range configs {
		if err := pr.SetPluginConfig(name, config); err != nil {
			pr.logger.Error("Failed to set plugin config", zap.String("name", name), zap.Error(err))
			continue
		}
	}

	return nil
}

// SavePluginConfigs 保存插件配置为JSON
func (pr *PluginRegistry) SavePluginConfigs() ([]byte, error) {
	pr.mutex.RLock()
	defer pr.mutex.RUnlock()

	return json.MarshalIndent(pr.configs, "", "  ")
}