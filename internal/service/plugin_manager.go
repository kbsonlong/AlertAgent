package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"alert_agent/internal/application/channel/plugins"
	"alert_agent/internal/model"
	"alert_agent/internal/pkg/database"
	"alert_agent/internal/pkg/logger"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// PluginManagerService 插件管理服务
type PluginManagerService struct {
	pluginManager *plugins.PluginManager
	registry      *plugins.PluginRegistry
	db            *gorm.DB
	logger        *zap.Logger
}

// NewPluginManagerService 创建插件管理服务
func NewPluginManagerService() *PluginManagerService {
	pluginManager := plugins.NewPluginManager(logger.GetLogger())
	
	return &PluginManagerService{
		pluginManager: pluginManager,
		registry:      pluginManager.GetRegistry(),
		db:            database.GetDB(),
		logger:        logger.GetLogger(),
	}
}

// Start 启动插件管理服务
func (pms *PluginManagerService) Start() error {
	// 启动插件管理器
	if err := pms.pluginManager.Start(); err != nil {
		return fmt.Errorf("failed to start plugin manager: %w", err)
	}

	// 从数据库加载插件配置
	if err := pms.LoadPluginConfigs(context.Background()); err != nil {
		pms.logger.Error("Failed to load plugin configs from database", zap.Error(err))
	}

	return nil
}

// Stop 停止插件管理服务
func (pms *PluginManagerService) Stop() error {
	return pms.pluginManager.Stop()
}

// RegisterPlugin 注册插件
func (pms *PluginManagerService) RegisterPlugin(plugin plugins.NotificationPlugin) error {
	return pms.registry.RegisterPlugin(plugin)
}

// UnregisterPlugin 注销插件
func (pms *PluginManagerService) UnregisterPlugin(name string) error {
	return pms.registry.UnregisterPlugin(name)
}

// GetAvailablePlugins 获取所有可用插件
func (pms *PluginManagerService) GetAvailablePlugins() []plugins.PluginInfo {
	return pms.registry.GetAvailablePlugins()
}

// ConfigurePlugin 配置插件
func (pms *PluginManagerService) ConfigurePlugin(ctx context.Context, name string, config *plugins.PluginConfig) error {
	// 验证插件配置
	if err := pms.registry.ValidatePluginConfig(name, config.Config); err != nil {
		return fmt.Errorf("invalid plugin config: %w", err)
	}

	// 设置插件配置
	if err := pms.registry.SetPluginConfig(name, config); err != nil {
		return err
	}

	// 保存配置到数据库
	if err := pms.savePluginConfigToDB(ctx, name, config); err != nil {
		pms.logger.Error("Failed to save plugin config to database", 
			zap.String("plugin", name), zap.Error(err))
		// 不返回错误，因为内存中的配置已经设置成功
	}

	return nil
}

// GetPluginConfig 获取插件配置
func (pms *PluginManagerService) GetPluginConfig(name string) (*plugins.PluginConfig, error) {
	config, exists := pms.registry.GetPluginConfig(name)
	if !exists {
		return nil, fmt.Errorf("plugin %s not configured", name)
	}
	return config, nil
}

// SendNotification 发送通知
func (pms *PluginManagerService) SendNotification(ctx context.Context, pluginName string, message *plugins.NotificationMessage) (*plugins.SendResult, error) {
	return pms.registry.SendNotification(ctx, pluginName, message)
}

// SendNotificationToAll 向所有启用的插件发送通知
func (pms *PluginManagerService) SendNotificationToAll(ctx context.Context, message *plugins.NotificationMessage) map[string]*plugins.SendResult {
	return pms.registry.SendNotificationToAll(ctx, message)
}

// HealthCheckPlugin 检查插件健康状态
func (pms *PluginManagerService) HealthCheckPlugin(ctx context.Context, pluginName string) error {
	return pms.registry.HealthCheckPlugin(ctx, pluginName)
}

// HealthCheckAll 检查所有插件健康状态
func (pms *PluginManagerService) HealthCheckAll(ctx context.Context) map[string]error {
	return pms.registry.HealthCheckAll(ctx)
}

// GetPluginStats 获取插件统计信息
func (pms *PluginManagerService) GetPluginStats(pluginName string) (*plugins.PluginStats, error) {
	stats, exists := pms.registry.GetPluginStats(pluginName)
	if !exists {
		return nil, fmt.Errorf("plugin %s not found", pluginName)
	}
	return stats, nil
}

// GetAllPluginStats 获取所有插件统计信息
func (pms *PluginManagerService) GetAllPluginStats() map[string]*plugins.PluginStats {
	return pms.registry.GetAllPluginStats()
}

// LoadPluginConfigs 从数据库加载插件配置
func (pms *PluginManagerService) LoadPluginConfigs(ctx context.Context) error {
	var pluginConfigs []model.NotificationPlugin
	if err := pms.db.WithContext(ctx).Find(&pluginConfigs).Error; err != nil {
		return fmt.Errorf("failed to load plugin configs from database: %w", err)
	}

	for _, dbConfig := range pluginConfigs {
		var config map[string]interface{}
		if err := json.Unmarshal([]byte(dbConfig.Config), &config); err != nil {
			pms.logger.Error("Failed to unmarshal plugin config", 
				zap.String("plugin", dbConfig.Name), zap.Error(err))
			continue
		}

		pluginConfig := &plugins.PluginConfig{
			Name:     dbConfig.Name,
			Enabled:  dbConfig.Enabled,
			Config:   config,
			Priority: dbConfig.Priority,
		}

		if err := pms.registry.SetPluginConfig(dbConfig.Name, pluginConfig); err != nil {
			pms.logger.Error("Failed to set plugin config", 
				zap.String("plugin", dbConfig.Name), zap.Error(err))
		}
	}

	pms.logger.Info("Plugin configs loaded from database", zap.Int("count", len(pluginConfigs)))
	return nil
}

// savePluginConfigToDB 保存插件配置到数据库
func (pms *PluginManagerService) savePluginConfigToDB(ctx context.Context, name string, config *plugins.PluginConfig) error {
	configJSON, err := json.Marshal(config.Config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	dbConfig := model.NotificationPlugin{
		Name:        name,
		DisplayName: name, // TODO: 从插件获取显示名称
		Version:     "1.0.0", // TODO: 从插件获取版本
		Config:      string(configJSON),
		Enabled:     config.Enabled,
		Priority:    config.Priority,
		UpdatedAt:   time.Now(),
	}

	// 使用 UPSERT 操作
	if err := pms.db.WithContext(ctx).Save(&dbConfig).Error; err != nil {
		return fmt.Errorf("failed to save plugin config: %w", err)
	}

	return nil
}

// TestPluginConfig 测试插件配置
func (pms *PluginManagerService) TestPluginConfig(ctx context.Context, pluginName string, config map[string]interface{}) error {
	// 验证配置
	if err := pms.registry.ValidatePluginConfig(pluginName, config); err != nil {
		return fmt.Errorf("config validation failed: %w", err)
	}

	// 创建测试消息
	testMessage := &plugins.NotificationMessage{
		Title:     "配置测试",
		Content:   "这是一条测试消息，用于验证通知插件配置是否正确。",
		Severity:  "info",
		AlertID:   "test-" + fmt.Sprintf("%d", time.Now().Unix()),
		Timestamp: time.Now(),
		Labels: map[string]string{
			"test": "true",
		},
		Annotations: map[string]string{
			"description": "插件配置测试消息",
		},
	}

	// 获取插件
	plugin, exists := pms.registry.GetPlugin(pluginName)
	if !exists {
		return fmt.Errorf("plugin %s not found", pluginName)
	}

	// 发送测试消息
	return plugin.Send(ctx, config, testMessage)
}

// EnablePlugin 启用插件
func (pms *PluginManagerService) EnablePlugin(ctx context.Context, pluginName string) error {
	config, exists := pms.registry.GetPluginConfig(pluginName)
	if !exists {
		return fmt.Errorf("plugin %s not configured", pluginName)
	}

	config.Enabled = true
	return pms.ConfigurePlugin(ctx, pluginName, config)
}

// DisablePlugin 禁用插件
func (pms *PluginManagerService) DisablePlugin(ctx context.Context, pluginName string) error {
	config, exists := pms.registry.GetPluginConfig(pluginName)
	if !exists {
		return fmt.Errorf("plugin %s not configured", pluginName)
	}

	config.Enabled = false
	return pms.ConfigurePlugin(ctx, pluginName, config)
}

// GetPluginSchema 获取插件配置Schema
func (pms *PluginManagerService) GetPluginSchema(pluginName string) (map[string]interface{}, error) {
	plugin, exists := pms.registry.GetPlugin(pluginName)
	if !exists {
		return nil, fmt.Errorf("plugin %s not found", pluginName)
	}

	return plugin.ConfigSchema(), nil
}

// GetPluginHealthStatus 获取插件健康状态
func (pms *PluginManagerService) GetPluginHealthStatus(pluginName string) (*plugins.HealthStatus, error) {
	healthChecker := pms.pluginManager.GetHealthChecker()
	status, exists := healthChecker.GetHealthStatus(pluginName)
	if !exists {
		return nil, fmt.Errorf("health status not found for plugin %s", pluginName)
	}
	return status, nil
}

// GetAllPluginHealthStatus 获取所有插件健康状态
func (pms *PluginManagerService) GetAllPluginHealthStatus() map[string]*plugins.HealthStatus {
	healthChecker := pms.pluginManager.GetHealthChecker()
	return healthChecker.GetAllHealthStatus()
}

// GetPluginUsageStats 获取插件使用统计
func (pms *PluginManagerService) GetPluginUsageStats(pluginName string) (*plugins.UsageStats, error) {
	statsCollector := pms.pluginManager.GetStatsCollector()
	stats, exists := statsCollector.GetUsageStats(pluginName)
	if !exists {
		return nil, fmt.Errorf("usage stats not found for plugin %s", pluginName)
	}
	return stats, nil
}

// GetAllPluginUsageStats 获取所有插件使用统计
func (pms *PluginManagerService) GetAllPluginUsageStats() map[string]*plugins.UsageStats {
	statsCollector := pms.pluginManager.GetStatsCollector()
	return statsCollector.GetAllUsageStats()
}

// HotLoadPlugin 热加载插件
func (pms *PluginManagerService) HotLoadPlugin(pluginPath string) error {
	hotLoader := pms.pluginManager.GetHotLoader()
	return hotLoader.LoadPlugin(pluginPath)
}

// HotUnloadPlugin 热卸载插件
func (pms *PluginManagerService) HotUnloadPlugin(pluginName string) error {
	hotLoader := pms.pluginManager.GetHotLoader()
	return hotLoader.UnloadPlugin(pluginName)
}

// HotReloadPlugin 热重载插件
func (pms *PluginManagerService) HotReloadPlugin(pluginName string) error {
	hotLoader := pms.pluginManager.GetHotLoader()
	return hotLoader.ReloadPlugin(pluginName)
}