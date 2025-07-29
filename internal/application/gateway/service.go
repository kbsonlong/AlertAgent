package gateway

import (
	"context"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"alert_agent/internal/application/channel"
	"alert_agent/internal/domain/gateway"
	"alert_agent/internal/shared/errors"
)

// GatewayService 网关服务接口
type GatewayService interface {
	// 告警处理
	ReceiveAlert(ctx context.Context, alert *gateway.Alert) error
	ProcessAlert(ctx context.Context, alert *gateway.Alert) (*gateway.ProcessingResult, error)

	// 处理记录管理
	GetProcessingRecord(ctx context.Context, alertID string) (*gateway.ProcessingRecord, error)
	ListProcessingRecords(ctx context.Context, query *gateway.ProcessingQuery) ([]*gateway.ProcessingRecord, int64, error)

	// 收敛管理
	CreateConvergenceRecord(ctx context.Context, record *gateway.ConvergenceRecord) error
	GetConvergenceWindow(ctx context.Context, key string) (*gateway.ConvergenceWindow, error)
	SetConvergenceWindow(ctx context.Context, window *gateway.ConvergenceWindow, ttl time.Duration) error

	// 抑制规则管理
	CreateSuppressionRule(ctx context.Context, rule *gateway.SuppressionRule) (*gateway.SuppressionRule, error)
	UpdateSuppressionRule(ctx context.Context, id string, rule *gateway.SuppressionRule) (*gateway.SuppressionRule, error)
	DeleteSuppressionRule(ctx context.Context, id string) error
	GetSuppressionRule(ctx context.Context, id string) (*gateway.SuppressionRule, error)
	ListSuppressionRules(ctx context.Context, enabled *bool) ([]*gateway.SuppressionRule, error)

	// 路由规则管理
	CreateRoutingRule(ctx context.Context, rule *gateway.RoutingRule) (*gateway.RoutingRule, error)
	UpdateRoutingRule(ctx context.Context, id string, rule *gateway.RoutingRule) (*gateway.RoutingRule, error)
	DeleteRoutingRule(ctx context.Context, id string) error
	GetRoutingRule(ctx context.Context, id string) (*gateway.RoutingRule, error)
	ListRoutingRules(ctx context.Context, enabled *bool) ([]*gateway.RoutingRule, error)
}

// gatewayServiceImpl 网关服务实现
type gatewayServiceImpl struct {
	repo           gateway.GatewayRepository
	channelService channel.ChannelService
	logger         *zap.Logger
}

// NewGatewayService 创建网关服务
func NewGatewayService(
	repo gateway.GatewayRepository,
	channelService channel.ChannelService,
	logger *zap.Logger,
) GatewayService {
	return &gatewayServiceImpl{
		repo:           repo,
		channelService: channelService,
		logger:         logger,
	}
}

// ReceiveAlert 接收告警
func (s *gatewayServiceImpl) ReceiveAlert(ctx context.Context, alert *gateway.Alert) error {
	// 创建处理记录
	record := &gateway.ProcessingRecord{
		ID:               uuid.New().String(),
		AlertID:          alert.ID,
		AlertName:        alert.Name,
		Severity:         alert.Severity,
		ReceivedAt:       time.Now(),
		ProcessingStatus: "received",
		Labels:           alert.Labels,
		Annotations:      alert.Annotations,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	if err := s.repo.CreateProcessingRecord(ctx, record); err != nil {
		s.logger.Error("Failed to create processing record", zap.Error(err), zap.String("alert_id", alert.ID))
		return errors.NewInternalError("Failed to create processing record", err)
	}

	s.logger.Info("Alert received successfully", 
		zap.String("alert_id", alert.ID), 
		zap.String("alert_name", alert.Name),
		zap.String("severity", alert.Severity))

	return nil
}

// ProcessAlert 处理告警
func (s *gatewayServiceImpl) ProcessAlert(ctx context.Context, alert *gateway.Alert) (*gateway.ProcessingResult, error) {
	startTime := time.Now()

	// 获取处理记录
	record, err := s.repo.GetProcessingRecord(ctx, alert.ID)
	if err != nil {
		return nil, errors.NewNotFoundError("Processing record")
	}

	// 更新处理状态
	record.ProcessingStatus = "processing"
	record.UpdatedAt = time.Now()

	// 这里实现基础的直通模式处理逻辑
	// 在第一阶段，告警直接通过用户定义的渠道发送
	result := &gateway.ProcessingResult{
		AlertID:        alert.ID,
		Action:         gateway.ActionSent,
		ProcessingTime: int(time.Since(startTime).Milliseconds()),
		Metadata:       make(map[string]interface{}),
	}

	// 更新处理记录
	record.ProcessingStatus = "processed"
	record.ProcessedAt = &[]time.Time{time.Now()}[0]
	record.ActionTaken = string(result.Action)
	record.ResolutionTime = result.ProcessingTime
	record.UpdatedAt = time.Now()

	if err := s.repo.UpdateProcessingRecord(ctx, record); err != nil {
		s.logger.Error("Failed to update processing record", zap.Error(err), zap.String("alert_id", alert.ID))
		return nil, errors.NewInternalError("Failed to update processing record", err)
	}

	s.logger.Info("Alert processed successfully", 
		zap.String("alert_id", alert.ID),
		zap.String("action", string(result.Action)),
		zap.Int("processing_time", result.ProcessingTime))

	return result, nil
}

// GetProcessingRecord 获取处理记录
func (s *gatewayServiceImpl) GetProcessingRecord(ctx context.Context, alertID string) (*gateway.ProcessingRecord, error) {
	record, err := s.repo.GetProcessingRecord(ctx, alertID)
	if err != nil {
		return nil, errors.NewNotFoundError("Processing record")
	}
	return record, nil
}

// ListProcessingRecords 列出处理记录
func (s *gatewayServiceImpl) ListProcessingRecords(ctx context.Context, query *gateway.ProcessingQuery) ([]*gateway.ProcessingRecord, int64, error) {
	records, total, err := s.repo.ListProcessingRecords(ctx, query)
	if err != nil {
		s.logger.Error("Failed to list processing records", zap.Error(err))
		return nil, 0, errors.NewInternalError("Failed to list processing records", err)
	}
	return records, total, nil
}

// CreateConvergenceRecord 创建收敛记录
func (s *gatewayServiceImpl) CreateConvergenceRecord(ctx context.Context, record *gateway.ConvergenceRecord) error {
	if err := s.repo.CreateConvergenceRecord(ctx, record); err != nil {
		s.logger.Error("Failed to create convergence record", zap.Error(err))
		return errors.NewInternalError("Failed to create convergence record", err)
	}
	return nil
}

// GetConvergenceWindow 获取收敛窗口
func (s *gatewayServiceImpl) GetConvergenceWindow(ctx context.Context, key string) (*gateway.ConvergenceWindow, error) {
	window, err := s.repo.GetConvergenceWindow(ctx, key)
	if err != nil {
		return nil, errors.NewNotFoundError("Convergence window")
	}
	return window, nil
}

// SetConvergenceWindow 设置收敛窗口
func (s *gatewayServiceImpl) SetConvergenceWindow(ctx context.Context, window *gateway.ConvergenceWindow, ttl time.Duration) error {
	if err := s.repo.SetConvergenceWindow(ctx, window, ttl); err != nil {
		s.logger.Error("Failed to set convergence window", zap.Error(err), zap.String("key", window.Key))
		return errors.NewInternalError("Failed to set convergence window", err)
	}
	return nil
}

// CreateSuppressionRule 创建抑制规则
func (s *gatewayServiceImpl) CreateSuppressionRule(ctx context.Context, rule *gateway.SuppressionRule) (*gateway.SuppressionRule, error) {
	rule.ID = uuid.New().String()
	rule.CreatedAt = time.Now()
	rule.UpdatedAt = time.Now()

	if err := s.repo.CreateSuppressionRule(ctx, rule); err != nil {
		s.logger.Error("Failed to create suppression rule", zap.Error(err), zap.String("name", rule.Name))
		return nil, errors.NewInternalError("Failed to create suppression rule", err)
	}

	s.logger.Info("Suppression rule created successfully", zap.String("id", rule.ID), zap.String("name", rule.Name))
	return rule, nil
}

// UpdateSuppressionRule 更新抑制规则
func (s *gatewayServiceImpl) UpdateSuppressionRule(ctx context.Context, id string, rule *gateway.SuppressionRule) (*gateway.SuppressionRule, error) {
	// 检查规则是否存在
	existing, err := s.repo.GetSuppressionRule(ctx, id)
	if err != nil {
		return nil, errors.NewNotFoundError("Suppression rule")
	}

	// 更新字段
	existing.Name = rule.Name
	existing.Description = rule.Description
	existing.Enabled = rule.Enabled
	existing.Priority = rule.Priority
	existing.Conditions = rule.Conditions
	existing.Schedule = rule.Schedule
	existing.UpdatedBy = rule.UpdatedBy
	existing.UpdatedAt = time.Now()

	if err := s.repo.UpdateSuppressionRule(ctx, existing); err != nil {
		s.logger.Error("Failed to update suppression rule", zap.Error(err), zap.String("id", id))
		return nil, errors.NewInternalError("Failed to update suppression rule", err)
	}

	s.logger.Info("Suppression rule updated successfully", zap.String("id", id), zap.String("name", existing.Name))
	return existing, nil
}

// DeleteSuppressionRule 删除抑制规则
func (s *gatewayServiceImpl) DeleteSuppressionRule(ctx context.Context, id string) error {
	// 检查规则是否存在
	_, err := s.repo.GetSuppressionRule(ctx, id)
	if err != nil {
		return errors.NewNotFoundError("Suppression rule")
	}

	if err := s.repo.DeleteSuppressionRule(ctx, id); err != nil {
		s.logger.Error("Failed to delete suppression rule", zap.Error(err), zap.String("id", id))
		return errors.NewInternalError("Failed to delete suppression rule", err)
	}

	s.logger.Info("Suppression rule deleted successfully", zap.String("id", id))
	return nil
}

// GetSuppressionRule 获取抑制规则
func (s *gatewayServiceImpl) GetSuppressionRule(ctx context.Context, id string) (*gateway.SuppressionRule, error) {
	rule, err := s.repo.GetSuppressionRule(ctx, id)
	if err != nil {
		return nil, errors.NewNotFoundError("Suppression rule")
	}
	return rule, nil
}

// ListSuppressionRules 列出抑制规则
func (s *gatewayServiceImpl) ListSuppressionRules(ctx context.Context, enabled *bool) ([]*gateway.SuppressionRule, error) {
	rules, err := s.repo.ListSuppressionRules(ctx, enabled)
	if err != nil {
		s.logger.Error("Failed to list suppression rules", zap.Error(err))
		return nil, errors.NewInternalError("Failed to list suppression rules", err)
	}
	return rules, nil
}

// CreateRoutingRule 创建路由规则
func (s *gatewayServiceImpl) CreateRoutingRule(ctx context.Context, rule *gateway.RoutingRule) (*gateway.RoutingRule, error) {
	rule.ID = uuid.New().String()
	rule.CreatedAt = time.Now()
	rule.UpdatedAt = time.Now()

	if err := s.repo.CreateRoutingRule(ctx, rule); err != nil {
		s.logger.Error("Failed to create routing rule", zap.Error(err), zap.String("name", rule.Name))
		return nil, errors.NewInternalError("Failed to create routing rule", err)
	}

	s.logger.Info("Routing rule created successfully", zap.String("id", rule.ID), zap.String("name", rule.Name))
	return rule, nil
}

// UpdateRoutingRule 更新路由规则
func (s *gatewayServiceImpl) UpdateRoutingRule(ctx context.Context, id string, rule *gateway.RoutingRule) (*gateway.RoutingRule, error) {
	// 检查规则是否存在
	existing, err := s.repo.GetRoutingRule(ctx, id)
	if err != nil {
		return nil, errors.NewNotFoundError("Routing rule")
	}

	// 更新字段
	existing.Name = rule.Name
	existing.Description = rule.Description
	existing.Enabled = rule.Enabled
	existing.Priority = rule.Priority
	existing.Conditions = rule.Conditions
	existing.Actions = rule.Actions
	existing.ChannelIDs = rule.ChannelIDs
	existing.UpdatedBy = rule.UpdatedBy
	existing.UpdatedAt = time.Now()

	if err := s.repo.UpdateRoutingRule(ctx, existing); err != nil {
		s.logger.Error("Failed to update routing rule", zap.Error(err), zap.String("id", id))
		return nil, errors.NewInternalError("Failed to update routing rule", err)
	}

	s.logger.Info("Routing rule updated successfully", zap.String("id", id), zap.String("name", existing.Name))
	return existing, nil
}

// DeleteRoutingRule 删除路由规则
func (s *gatewayServiceImpl) DeleteRoutingRule(ctx context.Context, id string) error {
	// 检查规则是否存在
	_, err := s.repo.GetRoutingRule(ctx, id)
	if err != nil {
		return errors.NewNotFoundError("Routing rule")
	}

	if err := s.repo.DeleteRoutingRule(ctx, id); err != nil {
		s.logger.Error("Failed to delete routing rule", zap.Error(err), zap.String("id", id))
		return errors.NewInternalError("Failed to delete routing rule", err)
	}

	s.logger.Info("Routing rule deleted successfully", zap.String("id", id))
	return nil
}

// GetRoutingRule 获取路由规则
func (s *gatewayServiceImpl) GetRoutingRule(ctx context.Context, id string) (*gateway.RoutingRule, error) {
	rule, err := s.repo.GetRoutingRule(ctx, id)
	if err != nil {
		return nil, errors.NewNotFoundError("Routing rule")
	}
	return rule, nil
}

// ListRoutingRules 列出路由规则
func (s *gatewayServiceImpl) ListRoutingRules(ctx context.Context, enabled *bool) ([]*gateway.RoutingRule, error) {
	rules, err := s.repo.ListRoutingRules(ctx, enabled)
	if err != nil {
		s.logger.Error("Failed to list routing rules", zap.Error(err))
		return nil, errors.NewInternalError("Failed to list routing rules", err)
	}
	return rules, nil
}