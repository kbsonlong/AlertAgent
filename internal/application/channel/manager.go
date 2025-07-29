package channel

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"

	"alert_agent/internal/domain/channel"
	"alert_agent/pkg/types"
)

// DefaultChannelManager 默认渠道管理器实现
type DefaultChannelManager struct {
	repo          channel.Repository
	service       channel.Service
	logger        *zap.Logger
	plugins       map[channel.ChannelType]channel.ChannelPlugin
	pluginsMutex  sync.RWMutex
	healthMonitor *HealthMonitor
	config        *channel.ManagerConfig
	metrics       *ChannelMetrics
	running       bool
	mutex         sync.RWMutex
}

// NewDefaultChannelManager 创建默认渠道管理器
func NewDefaultChannelManager(
	repo channel.Repository,
	service channel.Service,
	logger *zap.Logger,
) *DefaultChannelManager {
	return &DefaultChannelManager{
		repo:          repo,
		service:       service,
		logger:        logger,
		plugins:       make(map[channel.ChannelType]channel.ChannelPlugin),
		healthMonitor: NewHealthMonitor(logger),
		config: &channel.ManagerConfig{
			HealthCheckInterval: 5 * time.Minute,
			MaxRetries:          3,
			DefaultTimeout:      30 * time.Second,
			RateLimitEnabled:    true,
			MetricsEnabled:      true,
			PluginConfig:        make(map[string]interface{}),
		},
		metrics: NewChannelMetrics(),
	}
}

// RegisterPlugin 注册插件
func (m *DefaultChannelManager) RegisterPlugin(plugin channel.ChannelPlugin) error {
	m.pluginsMutex.Lock()
	defer m.pluginsMutex.Unlock()

	pluginType := plugin.GetType()
	if _, exists := m.plugins[pluginType]; exists {
		return fmt.Errorf("plugin for channel type %s already registered", pluginType)
	}

	// 初始化插件
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	pluginConfig := m.config.PluginConfig[string(pluginType)]
	if pluginConfig == nil {
		pluginConfig = make(map[string]interface{})
	}

	if err := plugin.Initialize(ctx, pluginConfig.(map[string]interface{})); err != nil {
		return fmt.Errorf("failed to initialize plugin %s: %w", pluginType, err)
	}

	if err := plugin.Start(ctx); err != nil {
		return fmt.Errorf("failed to start plugin %s: %w", pluginType, err)
	}

	m.plugins[pluginType] = plugin
	m.logger.Info("Plugin registered successfully",
		zap.String("type", string(pluginType)),
		zap.String("name", plugin.GetName()),
		zap.String("version", plugin.GetVersion()))

	return nil
}

// UnregisterPlugin 注销插件
func (m *DefaultChannelManager) UnregisterPlugin(channelType channel.ChannelType) error {
	m.pluginsMutex.Lock()
	defer m.pluginsMutex.Unlock()

	plugin, exists := m.plugins[channelType]
	if !exists {
		return fmt.Errorf("plugin for channel type %s not found", channelType)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := plugin.Stop(ctx); err != nil {
		m.logger.Warn("Failed to stop plugin gracefully",
			zap.String("type", string(channelType)),
			zap.Error(err))
	}

	delete(m.plugins, channelType)
	m.logger.Info("Plugin unregistered", zap.String("type", string(channelType)))

	return nil
}

// GetPlugin 获取插件
func (m *DefaultChannelManager) GetPlugin(channelType channel.ChannelType) (channel.ChannelPlugin, error) {
	m.pluginsMutex.RLock()
	defer m.pluginsMutex.RUnlock()

	plugin, exists := m.plugins[channelType]
	if !exists {
		return nil, fmt.Errorf("plugin for channel type %s not found", channelType)
	}

	return plugin, nil
}

// ListPlugins 列出所有插件
func (m *DefaultChannelManager) ListPlugins() []channel.ChannelPlugin {
	m.pluginsMutex.RLock()
	defer m.pluginsMutex.RUnlock()

	plugins := make([]channel.ChannelPlugin, 0, len(m.plugins))
	for _, plugin := range m.plugins {
		plugins = append(plugins, plugin)
	}

	return plugins
}

// CreateChannel 创建渠道
func (m *DefaultChannelManager) CreateChannel(ctx context.Context, req *channel.CreateChannelRequest) (*channel.Channel, error) {
	// 验证插件是否存在
	plugin, err := m.GetPlugin(req.Type)
	if err != nil {
		return nil, fmt.Errorf("unsupported channel type %s: %w", req.Type, err)
	}

	// 验证配置
	if err := plugin.ValidateConfig(req.Config); err != nil {
		return nil, fmt.Errorf("invalid channel config: %w", err)
	}

	// 创建渠道
	ch, err := m.service.CreateChannel(ctx, req)
	if err != nil {
		return nil, err
	}

	// 更新指标
	m.metrics.IncChannelCreated(req.Type)

	m.logger.Info("Channel created",
		zap.String("id", ch.ID),
		zap.String("name", ch.Name),
		zap.String("type", string(ch.Type)))

	return ch, nil
}

// UpdateChannel 更新渠道
func (m *DefaultChannelManager) UpdateChannel(ctx context.Context, id string, req *channel.UpdateChannelRequest) (*channel.Channel, error) {
	// 获取现有渠道
	existingChannel, err := m.service.GetChannel(ctx, id)
	if err != nil {
		return nil, err
	}

	// 如果更新了配置，需要验证
	if req.Config != nil {
		plugin, err := m.GetPlugin(existingChannel.Type)
		if err != nil {
			return nil, err
		}

		if err := plugin.ValidateConfig(*req.Config); err != nil {
			return nil, fmt.Errorf("invalid channel config: %w", err)
		}
	}

	// 更新渠道
	ch, err := m.service.UpdateChannel(ctx, id, req)
	if err != nil {
		return nil, err
	}

	m.logger.Info("Channel updated",
		zap.String("id", ch.ID),
		zap.String("name", ch.Name))

	return ch, nil
}

// DeleteChannel 删除渠道
func (m *DefaultChannelManager) DeleteChannel(ctx context.Context, id string) error {
	// 获取渠道信息用于日志
	ch, err := m.service.GetChannel(ctx, id)
	if err != nil {
		return err
	}

	// 删除渠道
	if err := m.service.DeleteChannel(ctx, id); err != nil {
		return err
	}

	// 更新指标
	m.metrics.IncChannelDeleted(ch.Type)

	m.logger.Info("Channel deleted",
		zap.String("id", id),
		zap.String("name", ch.Name),
		zap.String("type", string(ch.Type)))

	return nil
}

// GetChannel 获取渠道
func (m *DefaultChannelManager) GetChannel(ctx context.Context, id string) (*channel.Channel, error) {
	return m.service.GetChannel(ctx, id)
}

// ListChannels 列出渠道
func (m *DefaultChannelManager) ListChannels(ctx context.Context, query types.Query) (*types.PageResult, error) {
	return m.service.ListChannels(ctx, query)
}

// SendMessage 发送消息
func (m *DefaultChannelManager) SendMessage(ctx context.Context, channelID string, message *types.Message) (*channel.SendResult, error) {
	start := time.Now()

	// 获取渠道信息
	ch, err := m.service.GetChannel(ctx, channelID)
	if err != nil {
		return nil, fmt.Errorf("failed to get channel: %w", err)
	}

	// 检查渠道是否可用
	if !ch.CanSend() {
		return &channel.SendResult{
			ChannelID: channelID,
			Success:   false,
			Error:     "channel is not active or disabled",
			Latency:   time.Since(start),
			Timestamp: time.Now(),
		}, nil
	}

	// 获取插件
	plugin, err := m.GetPlugin(ch.Type)
	if err != nil {
		return &channel.SendResult{
			ChannelID: channelID,
			Success:   false,
			Error:     fmt.Sprintf("plugin not found: %v", err),
			Latency:   time.Since(start),
			Timestamp: time.Now(),
		}, nil
	}

	// 发送消息
	result, err := plugin.SendMessage(ctx, ch.Config, message)
	if err != nil {
		result = &channel.SendResult{
			ChannelID: channelID,
			Success:   false,
			Error:     err.Error(),
			Latency:   time.Since(start),
			Timestamp: time.Now(),
		}
	} else {
		result.ChannelID = channelID
		result.Latency = time.Since(start)
		result.Timestamp = time.Now()
	}

	// 更新指标
	if result.Success {
		m.metrics.IncMessageSent(ch.Type)
		m.metrics.RecordLatency(ch.Type, result.Latency)
	} else {
		m.metrics.IncMessageFailed(ch.Type)
	}

	m.logger.Debug("Message sent",
		zap.String("channel_id", channelID),
		zap.String("channel_type", string(ch.Type)),
		zap.Bool("success", result.Success),
		zap.Duration("latency", result.Latency))

	return result, nil
}

// BroadcastMessage 广播消息
func (m *DefaultChannelManager) BroadcastMessage(ctx context.Context, channelIDs []string, message *types.Message) ([]*channel.SendResult, error) {
	results := make([]*channel.SendResult, len(channelIDs))

	// 并发发送消息
	var wg sync.WaitGroup
	for i, channelID := range channelIDs {
		wg.Add(1)
		go func(index int, id string) {
			defer wg.Done()
			result, err := m.SendMessage(ctx, id, message)
			if err != nil {
				results[index] = &channel.SendResult{
					ChannelID: id,
					Success:   false,
					Error:     err.Error(),
					Timestamp: time.Now(),
				}
			} else {
				results[index] = result
			}
		}(i, channelID)
	}

	wg.Wait()
	return results, nil
}

// TestChannel 测试渠道
func (m *DefaultChannelManager) TestChannel(ctx context.Context, id string) (*channel.TestResult, error) {
	// 获取渠道信息
	ch, err := m.service.GetChannel(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get channel: %w", err)
	}

	// 获取插件
	plugin, err := m.GetPlugin(ch.Type)
	if err != nil {
		return nil, fmt.Errorf("plugin not found: %w", err)
	}

	// 测试连接
	result, err := plugin.TestConnection(ctx, ch.Config)
	if err != nil {
		return &channel.TestResult{
			Success:   false,
			Message:   err.Error(),
			Timestamp: time.Now().Unix(),
		}, nil
	}

	m.logger.Info("Channel tested",
		zap.String("channel_id", id),
		zap.Bool("success", result.Success))

	return result, nil
}

// ValidateConfig 验证配置
func (m *DefaultChannelManager) ValidateConfig(ctx context.Context, channelType channel.ChannelType, config channel.ChannelConfig) error {
	plugin, err := m.GetPlugin(channelType)
	if err != nil {
		return fmt.Errorf("plugin not found: %w", err)
	}

	return plugin.ValidateConfig(config)
}

// HealthCheck 健康检查
func (m *DefaultChannelManager) HealthCheck(ctx context.Context, channelID string) (*channel.HealthStatus, error) {
	return m.healthMonitor.CheckChannel(ctx, channelID, m)
}

// BatchHealthCheck 批量健康检查
func (m *DefaultChannelManager) BatchHealthCheck(ctx context.Context, channelIDs []string) (map[string]*channel.HealthStatus, error) {
	return m.healthMonitor.BatchCheck(ctx, channelIDs, m)
}

// StartHealthMonitor 启动健康监控
func (m *DefaultChannelManager) StartHealthMonitor(ctx context.Context, interval time.Duration) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if m.running {
		return fmt.Errorf("health monitor is already running")
	}

	m.config.HealthCheckInterval = interval
	m.running = true

	go m.healthMonitor.Start(ctx, interval, m)

	m.logger.Info("Health monitor started", zap.Duration("interval", interval))
	return nil
}

// StopHealthMonitor 停止健康监控
func (m *DefaultChannelManager) StopHealthMonitor() error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if !m.running {
		return fmt.Errorf("health monitor is not running")
	}

	m.healthMonitor.Stop()
	m.running = false

	m.logger.Info("Health monitor stopped")
	return nil
}

// GetChannelStats 获取渠道统计
func (m *DefaultChannelManager) GetChannelStats(ctx context.Context, channelID string) (*channel.ChannelStats, error) {
	return m.service.GetChannelStats(ctx, channelID)
}

// GetActiveChannels 获取激活的渠道
func (m *DefaultChannelManager) GetActiveChannels(ctx context.Context) ([]*channel.Channel, error) {
	return m.service.GetActiveChannels(ctx)
}

// GetGlobalStats 获取全局统计
func (m *DefaultChannelManager) GetGlobalStats(ctx context.Context) (*channel.GlobalStats, error) {
	// 获取所有渠道
	channels, err := m.service.GetActiveChannels(ctx)
	if err != nil {
		return nil, err
	}

	stats := &channel.GlobalStats{
		TotalChannels:    len(channels),
		ChannelsByType:   make(map[channel.ChannelType]int),
		ChannelsByStatus: make(map[channel.ChannelStatus]int),
		LastUpdated:      time.Now(),
	}

	// 统计各类型和状态的渠道数量
	for _, ch := range channels {
		stats.ChannelsByType[ch.Type]++
		stats.ChannelsByStatus[ch.Status]++
		if ch.Status == channel.ChannelStatusActive {
			stats.ActiveChannels++
		}
	}

	// 从指标系统获取统计数据
	metrics := m.metrics.GetGlobalMetrics()
	stats.TotalSent = metrics.TotalSent
	stats.TotalFailed = metrics.TotalFailed
	if stats.TotalSent > 0 {
		stats.SuccessRate = float64(stats.TotalSent-stats.TotalFailed) / float64(stats.TotalSent) * 100
	}
	stats.AvgLatency = metrics.AvgLatency

	return stats, nil
}

// ReloadConfig 重新加载配置
func (m *DefaultChannelManager) ReloadConfig(ctx context.Context) error {
	m.logger.Info("Reloading channel manager configuration")
	// 这里可以从配置文件或配置中心重新加载配置
	return nil
}

// GetConfig 获取配置
func (m *DefaultChannelManager) GetConfig(ctx context.Context) (*channel.ManagerConfig, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return m.config, nil
}

// UpdateConfig 更新配置
func (m *DefaultChannelManager) UpdateConfig(ctx context.Context, config *channel.ManagerConfig) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.config = config
	m.logger.Info("Channel manager configuration updated")
	return nil
}
