package channel

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"alert_agent/internal/domain/channel"
	"alert_agent/internal/shared/errors"
)

// ChannelService 渠道服务接口
type ChannelService interface {
	// 渠道管理
	CreateChannel(ctx context.Context, req *channel.CreateChannelRequest) (*channel.Channel, error)
	UpdateChannel(ctx context.Context, id string, req *channel.UpdateChannelRequest) (*channel.Channel, error)
	DeleteChannel(ctx context.Context, id string) error
	GetChannel(ctx context.Context, id string) (*channel.Channel, error)
	ListChannels(ctx context.Context, query *channel.ChannelQuery) ([]*channel.Channel, int64, error)

	// 消息发送
	SendMessage(ctx context.Context, channelID string, message *channel.Message) error
	BroadcastMessage(ctx context.Context, channelIDs []string, message *channel.Message) error

	// 健康检查和测试
	TestChannel(ctx context.Context, id string) (*channel.TestResult, error)
	GetChannelHealth(ctx context.Context, id string) (*channel.HealthStatus, error)

	// 插件管理
	RegisterPlugin(plugin channel.ChannelPlugin) error
	GetAvailablePlugins() []channel.PluginInfo
}

// channelServiceImpl 渠道服务实现
type channelServiceImpl struct {
	repo    channel.ChannelRepository
	plugins map[string]channel.ChannelPlugin
	logger  *zap.Logger
}

// NewChannelService 创建渠道服务
func NewChannelService(repo channel.ChannelRepository, logger *zap.Logger) ChannelService {
	return &channelServiceImpl{
		repo:    repo,
		plugins: make(map[string]channel.ChannelPlugin),
		logger:  logger,
	}
}

// CreateChannel 创建渠道
func (s *channelServiceImpl) CreateChannel(ctx context.Context, req *channel.CreateChannelRequest) (*channel.Channel, error) {
	// 验证插件是否存在
	plugin, exists := s.plugins[req.Type]
	if !exists {
		return nil, errors.NewValidationError("INVALID_CHANNEL_TYPE", fmt.Sprintf("Channel type %s is not supported", req.Type))
	}

	// 验证配置
	if err := plugin.ValidateConfig(req.Config); err != nil {
		return nil, errors.NewValidationErrorWithDetails("INVALID_CONFIG", "Channel configuration is invalid", err.Error())
	}

	// 创建渠道对象
	ch := &channel.Channel{
		ID:           uuid.New().String(),
		Name:         req.Name,
		Type:         req.Type,
		Description:  req.Description,
		Config:       req.Config,
		GroupID:      req.GroupID,
		Tags:         req.Tags,
		Status:       channel.ChannelStatusActive,
		HealthStatus: channel.HealthStatusUnknown,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	// 保存到数据库
	if err := s.repo.Create(ctx, ch); err != nil {
		s.logger.Error("Failed to create channel", zap.Error(err), zap.String("name", req.Name))
		return nil, errors.NewInternalError("Failed to create channel", err)
	}

	s.logger.Info("Channel created successfully", zap.String("id", ch.ID), zap.String("name", ch.Name), zap.String("type", ch.Type))
	return ch, nil
}

// UpdateChannel 更新渠道
func (s *channelServiceImpl) UpdateChannel(ctx context.Context, id string, req *channel.UpdateChannelRequest) (*channel.Channel, error) {
	// 获取现有渠道
	ch, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, errors.NewNotFoundError("Channel")
	}

	// 如果更新了配置，需要验证
	if req.Config != nil {
		plugin, exists := s.plugins[ch.Type]
		if !exists {
			return nil, errors.NewInternalError("Channel plugin not found", nil)
		}

		if err := plugin.ValidateConfig(req.Config); err != nil {
			return nil, errors.NewValidationErrorWithDetails("INVALID_CONFIG", "Channel configuration is invalid", err.Error())
		}
		ch.Config = req.Config
	}

	// 更新字段
	if req.Name != "" {
		ch.Name = req.Name
	}
	if req.Description != "" {
		ch.Description = req.Description
	}
	if req.GroupID != "" {
		ch.GroupID = req.GroupID
	}
	if req.Tags != nil {
		ch.Tags = req.Tags
	}
	if req.Status != "" {
		ch.Status = req.Status
	}
	ch.UpdatedAt = time.Now()

	// 保存更新
	if err := s.repo.Update(ctx, ch); err != nil {
		s.logger.Error("Failed to update channel", zap.Error(err), zap.String("id", id))
		return nil, errors.NewInternalError("Failed to update channel", err)
	}

	s.logger.Info("Channel updated successfully", zap.String("id", id), zap.String("name", ch.Name))
	return ch, nil
}

// DeleteChannel 删除渠道
func (s *channelServiceImpl) DeleteChannel(ctx context.Context, id string) error {
	// 检查渠道是否存在
	_, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return errors.NewNotFoundError("Channel")
	}

	// 删除渠道
	if err := s.repo.Delete(ctx, id); err != nil {
		s.logger.Error("Failed to delete channel", zap.Error(err), zap.String("id", id))
		return errors.NewInternalError("Failed to delete channel", err)
	}

	s.logger.Info("Channel deleted successfully", zap.String("id", id))
	return nil
}

// GetChannel 获取渠道
func (s *channelServiceImpl) GetChannel(ctx context.Context, id string) (*channel.Channel, error) {
	ch, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, errors.NewNotFoundError("Channel")
	}
	return ch, nil
}

// ListChannels 列出渠道
func (s *channelServiceImpl) ListChannels(ctx context.Context, query *channel.ChannelQuery) ([]*channel.Channel, int64, error) {
	channels, total, err := s.repo.List(ctx, query)
	if err != nil {
		s.logger.Error("Failed to list channels", zap.Error(err))
		return nil, 0, errors.NewInternalError("Failed to list channels", err)
	}
	return channels, total, nil
}

// SendMessage 发送消息
func (s *channelServiceImpl) SendMessage(ctx context.Context, channelID string, message *channel.Message) error {
	// 获取渠道
	ch, err := s.repo.GetByID(ctx, channelID)
	if err != nil {
		return errors.NewNotFoundError("Channel")
	}

	// 检查渠道状态
	if ch.Status != channel.ChannelStatusActive {
		return errors.NewValidationError("CHANNEL_INACTIVE", "Channel is not active")
	}

	// 获取插件
	plugin, exists := s.plugins[ch.Type]
	if !exists {
		return errors.NewInternalError("Channel plugin not found", nil)
	}

	// 发送消息
	startTime := time.Now()
	err = plugin.SendMessage(ch.Config, message)
	responseTime := int(time.Since(startTime).Milliseconds())

	// 更新健康状态
	if err != nil {
		s.repo.UpdateHealthStatus(ctx, channelID, channel.HealthStatusUnhealthy, responseTime)
		s.logger.Error("Failed to send message", zap.Error(err), zap.String("channel_id", channelID))
		return errors.NewExternalError("Channel", "Failed to send message", err)
	}

	s.repo.UpdateHealthStatus(ctx, channelID, channel.HealthStatusHealthy, responseTime)
	s.logger.Info("Message sent successfully", zap.String("channel_id", channelID), zap.Int("response_time", responseTime))
	return nil
}

// BroadcastMessage 广播消息
func (s *channelServiceImpl) BroadcastMessage(ctx context.Context, channelIDs []string, message *channel.Message) error {
	var errors []error
	for _, channelID := range channelIDs {
		if err := s.SendMessage(ctx, channelID, message); err != nil {
			errors = append(errors, err)
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("failed to send to %d channels: %v", len(errors), errors)
	}

	return nil
}

// TestChannel 测试渠道
func (s *channelServiceImpl) TestChannel(ctx context.Context, id string) (*channel.TestResult, error) {
	// 获取渠道
	ch, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, errors.NewNotFoundError("Channel")
	}

	// 获取插件
	plugin, exists := s.plugins[ch.Type]
	if !exists {
		return nil, errors.NewInternalError("Channel plugin not found", nil)
	}

	// 执行测试
	startTime := time.Now()
	result, err := plugin.TestConnection(ch.Config)
	if err != nil {
		return &channel.TestResult{
			Success:      false,
			ResponseTime: int(time.Since(startTime).Milliseconds()),
			Error:        err.Error(),
			Timestamp:    time.Now(),
		}, nil
	}

	return result, nil
}

// GetChannelHealth 获取渠道健康状态
func (s *channelServiceImpl) GetChannelHealth(ctx context.Context, id string) (*channel.HealthStatus, error) {
	ch, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, errors.NewNotFoundError("Channel")
	}

	return &ch.HealthStatus, nil
}

// RegisterPlugin 注册插件
func (s *channelServiceImpl) RegisterPlugin(plugin channel.ChannelPlugin) error {
	pluginType := plugin.GetType()
	if _, exists := s.plugins[pluginType]; exists {
		return fmt.Errorf("plugin type %s already registered", pluginType)
	}

	s.plugins[pluginType] = plugin
	s.logger.Info("Plugin registered", zap.String("type", pluginType), zap.String("name", plugin.GetName()))
	return nil
}

// GetAvailablePlugins 获取可用插件
func (s *channelServiceImpl) GetAvailablePlugins() []channel.PluginInfo {
	var plugins []channel.PluginInfo
	for _, plugin := range s.plugins {
		plugins = append(plugins, channel.PluginInfo{
			Type:        plugin.GetType(),
			Name:        plugin.GetName(),
			Description: fmt.Sprintf("%s notification channel", plugin.GetName()),
			Version:     "1.0.0",
			Schema:      plugin.GetConfigSchema(),
		})
	}
	return plugins
}