package channel

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"alert_agent/internal/domain/channel"
	"alert_agent/internal/shared/logger"
	"alert_agent/pkg/types"
)

// ChannelService 通道应用服务实现
type ChannelService struct {
	repo   channel.Repository
	logger *zap.Logger
}

// NewChannelService 创建通道服务
func NewChannelService(repo channel.Repository) channel.Service {
	return &ChannelService{
		repo:   repo,
		logger: logger.WithComponent("channel-service"),
	}
}

// CreateChannel 创建通道
func (s *ChannelService) CreateChannel(ctx context.Context, req *channel.CreateChannelRequest) (*channel.Channel, error) {
	s.logger.Info("creating channel", zap.String("name", req.Name), zap.String("type", string(req.Type)))

	// 验证请求
	if err := s.validateCreateRequest(req); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// 检查名称是否已存在
	exists, err := s.repo.ExistsByName(ctx, req.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to check name existence: %w", err)
	}
	if exists {
		return nil, fmt.Errorf("channel name '%s' already exists", req.Name)
	}

	// 创建通道实体
	now := time.Now()
	ch := &channel.Channel{
		ID:          uuid.New().String(),
		Name:        req.Name,
		Type:        req.Type,
		Description: req.Description,
		Config:      req.Config,
		Status:      channel.ChannelStatusActive,
		Priority:    req.Priority,
		Tags:        req.Tags,
		Labels:      req.Labels,
		Metadata: types.Metadata{
			CreatedAt: now,
			UpdatedAt: now,
			Version:   1,
		},
	}

	// 验证通道配置
	if err := ch.Validate(); err != nil {
		return nil, fmt.Errorf("channel validation failed: %w", err)
	}

	// 保存到数据库
	if err := s.repo.Create(ctx, ch); err != nil {
		return nil, fmt.Errorf("failed to create channel: %w", err)
	}

	s.logger.Info("channel created successfully", zap.String("id", ch.ID), zap.String("name", ch.Name))
	return ch, nil
}

// GetChannel 获取通道
func (s *ChannelService) GetChannel(ctx context.Context, id string) (*channel.Channel, error) {
	s.logger.Debug("getting channel", zap.String("id", id))

	ch, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get channel: %w", err)
	}

	return ch, nil
}

// UpdateChannel 更新通道
func (s *ChannelService) UpdateChannel(ctx context.Context, id string, req *channel.UpdateChannelRequest) (*channel.Channel, error) {
	s.logger.Info("updating channel", zap.String("id", id))

	// 获取现有通道
	ch, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get channel: %w", err)
	}

	// 更新字段
	if req.Name != nil {
		// 检查新名称是否已存在
		if *req.Name != ch.Name {
			exists, err := s.repo.ExistsByName(ctx, *req.Name)
			if err != nil {
				return nil, fmt.Errorf("failed to check name existence: %w", err)
			}
			if exists {
				return nil, fmt.Errorf("channel name '%s' already exists", *req.Name)
			}
		}
		ch.Name = *req.Name
	}

	if req.Description != nil {
		ch.Description = *req.Description
	}

	if req.Config != nil {
		ch.Config = *req.Config
	}

	if req.Status != nil {
		ch.Status = *req.Status
	}

	if req.Priority != nil {
		ch.Priority = *req.Priority
	}

	if req.Tags != nil {
		ch.Tags = req.Tags
	}

	if req.Labels != nil {
		ch.Labels = req.Labels
	}

	// 更新元数据
	ch.Metadata.UpdatedAt = time.Now()
	ch.Metadata.Version++

	// 验证更新后的通道
	if err := ch.Validate(); err != nil {
		return nil, fmt.Errorf("channel validation failed: %w", err)
	}

	// 保存更新
	if err := s.repo.Update(ctx, ch); err != nil {
		return nil, fmt.Errorf("failed to update channel: %w", err)
	}

	s.logger.Info("channel updated successfully", zap.String("id", ch.ID), zap.String("name", ch.Name))
	return ch, nil
}

// DeleteChannel 删除通道
func (s *ChannelService) DeleteChannel(ctx context.Context, id string) error {
	s.logger.Info("deleting channel", zap.String("id", id))

	// 检查通道是否存在
	exists, err := s.repo.Exists(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to check channel existence: %w", err)
	}
	if !exists {
		return fmt.Errorf("channel not found")
	}

	// 删除通道
	if err := s.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete channel: %w", err)
	}

	s.logger.Info("channel deleted successfully", zap.String("id", id))
	return nil
}

// ListChannels 获取通道列表
func (s *ChannelService) ListChannels(ctx context.Context, query types.Query) (*types.PageResult, error) {
	s.logger.Debug("listing channels", zap.Int("limit", query.Limit), zap.Int("offset", query.Offset))

	channels, total, err := s.repo.List(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list channels: %w", err)
	}

	page := query.Offset/query.Limit + 1
	if query.Limit == 0 {
		page = 1
	}

	return &types.PageResult{
		Data:  channels,
		Total: total,
		Page:  page,
		Size:  len(channels),
	}, nil
}

// TestChannel 测试通道连接
func (s *ChannelService) TestChannel(ctx context.Context, id string) (*channel.TestResult, error) {
	s.logger.Info("testing channel", zap.String("id", id))

	// 获取通道
	ch, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get channel: %w", err)
	}

	// 执行测试
	start := time.Now()
	testResult := &channel.TestResult{
		Timestamp: start.Unix(),
		Details:   make(map[string]interface{}),
	}

	// 根据通道类型执行不同的测试逻辑
	err = s.performChannelTest(ctx, ch, testResult)
	testResult.Latency = time.Since(start).Milliseconds()

	if err != nil {
		testResult.Success = false
		testResult.Message = err.Error()
		s.logger.Warn("channel test failed", zap.String("id", id), zap.Error(err))
	} else {
		testResult.Success = true
		testResult.Message = "Channel test successful"
		s.logger.Info("channel test successful", zap.String("id", id))
	}

	return testResult, nil
}

// SendMessage 发送消息
func (s *ChannelService) SendMessage(ctx context.Context, channelID string, message *types.Message) error {
	s.logger.Info("sending message", zap.String("channel_id", channelID), zap.String("message_id", message.ID))

	// 获取通道
	ch, err := s.repo.GetByID(ctx, channelID)
	if err != nil {
		return fmt.Errorf("failed to get channel: %w", err)
	}

	// 检查通道是否可以发送
	if !ch.CanSend() {
		return fmt.Errorf("channel is not available for sending")
	}

	// 执行发送逻辑
	err = s.performMessageSend(ctx, ch, message)
	if err != nil {
		s.logger.Error("failed to send message", zap.String("channel_id", channelID), zap.Error(err))
		return fmt.Errorf("failed to send message: %w", err)
	}

	s.logger.Info("message sent successfully", zap.String("channel_id", channelID), zap.String("message_id", message.ID))
	return nil
}

// GetChannelsByType 根据类型获取通道
func (s *ChannelService) GetChannelsByType(ctx context.Context, channelType channel.ChannelType) ([]*channel.Channel, error) {
	s.logger.Debug("getting channels by type", zap.String("type", string(channelType)))

	channels, err := s.repo.GetByType(ctx, channelType)
	if err != nil {
		return nil, fmt.Errorf("failed to get channels by type: %w", err)
	}

	return channels, nil
}

// GetActiveChannels 获取激活的通道
func (s *ChannelService) GetActiveChannels(ctx context.Context) ([]*channel.Channel, error) {
	s.logger.Debug("getting active channels")

	channels, err := s.repo.GetActiveChannels(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get active channels: %w", err)
	}

	return channels, nil
}

// UpdateChannelStatus 更新通道状态
func (s *ChannelService) UpdateChannelStatus(ctx context.Context, id string, status channel.ChannelStatus) error {
	s.logger.Info("updating channel status", zap.String("id", id), zap.String("status", string(status)))

	err := s.repo.UpdateStatus(ctx, id, status)
	if err != nil {
		return fmt.Errorf("failed to update channel status: %w", err)
	}

	return nil
}

// ValidateChannelConfig 验证通道配置
func (s *ChannelService) ValidateChannelConfig(ctx context.Context, channelType channel.ChannelType, config channel.ChannelConfig) error {
	s.logger.Debug("validating channel config", zap.String("type", string(channelType)))

	// 根据通道类型执行不同的验证逻辑
	return s.performConfigValidation(channelType, config)
}

// GetChannelStats 获取通道统计信息
func (s *ChannelService) GetChannelStats(ctx context.Context, id string) (*channel.ChannelStats, error) {
	s.logger.Debug("getting channel stats", zap.String("id", id))

	// 这里应该从监控系统或统计数据库获取统计信息
	// 目前返回模拟数据
	stats := &channel.ChannelStats{
		ChannelID:   id,
		TotalSent:   0,
		TotalFailed: 0,
		SuccessRate: 0.0,
		AvgLatency:  0.0,
	}

	return stats, nil
}

// BulkUpdateChannels 批量更新通道
func (s *ChannelService) BulkUpdateChannels(ctx context.Context, updates []*channel.BulkUpdateRequest) error {
	s.logger.Info("bulk updating channels", zap.Int("count", len(updates)))

	for _, update := range updates {
		_, err := s.UpdateChannel(ctx, update.ChannelID, &update.Updates)
		if err != nil {
			s.logger.Error("failed to update channel in bulk", zap.String("id", update.ChannelID), zap.Error(err))
			return fmt.Errorf("failed to update channel %s: %w", update.ChannelID, err)
		}
	}

	return nil
}

// ImportChannels 导入通道
func (s *ChannelService) ImportChannels(ctx context.Context, channels []*channel.Channel) (*channel.ImportResult, error) {
	s.logger.Info("importing channels", zap.Int("count", len(channels)))

	result := &channel.ImportResult{
		Total:   len(channels),
		Success: 0,
		Failed:  0,
		Errors:  []string{},
		Created: []string{},
		Updated: []string{},
		Skipped: []string{},
	}

	for _, ch := range channels {
		// 检查是否已存在
		exists, err := s.repo.ExistsByName(ctx, ch.Name)
		if err != nil {
			result.Failed++
			result.Errors = append(result.Errors, fmt.Sprintf("Failed to check existence for %s: %v", ch.Name, err))
			continue
		}

		if exists {
			result.Skipped = append(result.Skipped, ch.Name)
			continue
		}

		// 创建新通道
		ch.ID = uuid.New().String()
		ch.Metadata.CreatedAt = time.Now()
		ch.Metadata.UpdatedAt = time.Now()
		ch.Metadata.Version = 1

		if err := s.repo.Create(ctx, ch); err != nil {
			result.Failed++
			result.Errors = append(result.Errors, fmt.Sprintf("Failed to create %s: %v", ch.Name, err))
		} else {
			result.Success++
			result.Created = append(result.Created, ch.Name)
		}
	}

	return result, nil
}

// ExportChannels 导出通道
func (s *ChannelService) ExportChannels(ctx context.Context, filter map[string]interface{}) ([]*channel.Channel, error) {
	s.logger.Info("exporting channels")

	query := types.Query{
		Filter: filter,
		Limit:  0, // 获取所有
	}

	channels, _, err := s.repo.List(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to export channels: %w", err)
	}

	return channels, nil
}

// validateCreateRequest 验证创建请求
func (s *ChannelService) validateCreateRequest(req *channel.CreateChannelRequest) error {
	if req.Name == "" {
		return fmt.Errorf("name is required")
	}
	if req.Type == "" {
		return fmt.Errorf("type is required")
	}
	return nil
}

// performChannelTest 执行通道测试
func (s *ChannelService) performChannelTest(ctx context.Context, ch *channel.Channel, result *channel.TestResult) error {
	// 根据通道类型执行不同的测试逻辑
	switch ch.Type {
	case channel.ChannelTypeEmail:
		return s.testEmailChannel(ctx, ch, result)
	case channel.ChannelTypeWebhook:
		return s.testWebhookChannel(ctx, ch, result)
	case channel.ChannelTypeSlack:
		return s.testSlackChannel(ctx, ch, result)
	default:
		return fmt.Errorf("unsupported channel type: %s", ch.Type)
	}
}

// performMessageSend 执行消息发送
func (s *ChannelService) performMessageSend(ctx context.Context, ch *channel.Channel, message *types.Message) error {
	// 根据通道类型执行不同的发送逻辑
	switch ch.Type {
	case channel.ChannelTypeEmail:
		return s.sendEmailMessage(ctx, ch, message)
	case channel.ChannelTypeWebhook:
		return s.sendWebhookMessage(ctx, ch, message)
	case channel.ChannelTypeSlack:
		return s.sendSlackMessage(ctx, ch, message)
	default:
		return fmt.Errorf("unsupported channel type: %s", ch.Type)
	}
}

// performConfigValidation 执行配置验证
func (s *ChannelService) performConfigValidation(channelType channel.ChannelType, config channel.ChannelConfig) error {
	// 根据通道类型执行不同的验证逻辑
	switch channelType {
	case channel.ChannelTypeEmail:
		return s.validateEmailConfig(config)
	case channel.ChannelTypeWebhook:
		return s.validateWebhookConfig(config)
	case channel.ChannelTypeSlack:
		return s.validateSlackConfig(config)
	default:
		return fmt.Errorf("unsupported channel type: %s", channelType)
	}
}

// 以下是各种通道类型的具体实现方法（简化版本）

func (s *ChannelService) testEmailChannel(ctx context.Context, ch *channel.Channel, result *channel.TestResult) error {
	// TODO: 实现邮件通道测试逻辑
	result.Details["type"] = "email"
	return nil
}

func (s *ChannelService) testWebhookChannel(ctx context.Context, ch *channel.Channel, result *channel.TestResult) error {
	// TODO: 实现Webhook通道测试逻辑
	result.Details["type"] = "webhook"
	return nil
}

func (s *ChannelService) testSlackChannel(ctx context.Context, ch *channel.Channel, result *channel.TestResult) error {
	// TODO: 实现Slack通道测试逻辑
	result.Details["type"] = "slack"
	return nil
}

func (s *ChannelService) sendEmailMessage(ctx context.Context, ch *channel.Channel, message *types.Message) error {
	// TODO: 实现邮件发送逻辑
	return nil
}

func (s *ChannelService) sendWebhookMessage(ctx context.Context, ch *channel.Channel, message *types.Message) error {
	// TODO: 实现Webhook发送逻辑
	return nil
}

func (s *ChannelService) sendSlackMessage(ctx context.Context, ch *channel.Channel, message *types.Message) error {
	// TODO: 实现Slack发送逻辑
	return nil
}

func (s *ChannelService) validateEmailConfig(config channel.ChannelConfig) error {
	// TODO: 实现邮件配置验证逻辑
	return nil
}

func (s *ChannelService) validateWebhookConfig(config channel.ChannelConfig) error {
	// TODO: 实现Webhook配置验证逻辑
	return nil
}

func (s *ChannelService) validateSlackConfig(config channel.ChannelConfig) error {
	// TODO: 实现Slack配置验证逻辑
	return nil
}