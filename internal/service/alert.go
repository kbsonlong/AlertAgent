package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"alert_agent/internal/model"
	"alert_agent/internal/pkg/database"
	"alert_agent/internal/pkg/redis"
)

const (
	AlertCacheKeyPrefix = "alert:"
	AlertCacheTTL       = 24 * time.Hour
)

// AlertService 告警服务
type AlertService struct{}

// NewAlertService 创建告警服务实例
func NewAlertService() *AlertService {
	return &AlertService{}
}

// CreateAlert 创建告警
func (s *AlertService) CreateAlert(ctx context.Context, alert *model.Alert) error {
	// 创建告警记录
	if err := database.DB.Create(alert).Error; err != nil {
		return fmt.Errorf("failed to create alert: %w", err)
	}

	// 缓存告警信息
	if err := s.cacheAlert(ctx, alert); err != nil {
		return fmt.Errorf("failed to cache alert: %w", err)
	}

	// 触发告警通知
	if err := s.triggerNotification(ctx, alert); err != nil {
		return fmt.Errorf("failed to trigger notification: %w", err)
	}

	return nil
}

// GetAlert 获取告警信息
func (s *AlertService) GetAlert(ctx context.Context, id uint) (*model.Alert, error) {
	// 尝试从缓存获取
	alert, err := s.getAlertFromCache(ctx, id)
	if err == nil {
		return alert, nil
	}

	// 从数据库获取
	alert = &model.Alert{}
	if err := database.DB.First(alert, id).Error; err != nil {
		return nil, fmt.Errorf("failed to get alert: %w", err)
	}

	// 更新缓存
	if err := s.cacheAlert(ctx, alert); err != nil {
		return nil, fmt.Errorf("failed to cache alert: %w", err)
	}

	return alert, nil
}

// HandleAlert 处理告警
func (s *AlertService) HandleAlert(ctx context.Context, id uint, handler string, note string) error {
	alert := &model.Alert{}
	if err := database.DB.First(alert, id).Error; err != nil {
		return fmt.Errorf("failed to get alert: %w", err)
	}

	// 更新处理信息
	alert.Status = "handled"
	alert.Handler = handler
	alert.HandleNote = note
	now := time.Now()
	alert.HandleTime = &now

	if err := database.DB.Save(alert).Error; err != nil {
		return fmt.Errorf("failed to update alert: %w", err)
	}

	// 更新缓存
	if err := s.cacheAlert(ctx, alert); err != nil {
		return fmt.Errorf("failed to cache alert: %w", err)
	}

	return nil
}

// cacheAlert 缓存告警信息
func (s *AlertService) cacheAlert(ctx context.Context, alert *model.Alert) error {
	key := fmt.Sprintf("%s%d", AlertCacheKeyPrefix, alert.ID)
	data, err := json.Marshal(alert)
	if err != nil {
		return err
	}
	return redis.Set(ctx, key, string(data), AlertCacheTTL)
}

// getAlertFromCache 从缓存获取告警信息
func (s *AlertService) getAlertFromCache(ctx context.Context, id uint) (*model.Alert, error) {
	key := fmt.Sprintf("%s%d", AlertCacheKeyPrefix, id)
	data, err := redis.Get(ctx, key)
	if err != nil {
		return nil, err
	}

	alert := &model.Alert{}
	if err := json.Unmarshal([]byte(data), alert); err != nil {
		return nil, err
	}

	return alert, nil
}

// triggerNotification 触发告警通知
func (s *AlertService) triggerNotification(ctx context.Context, alert *model.Alert) error {
	// 获取告警规则
	rule := &model.Rule{}
	if err := database.DB.First(rule, alert.RuleID).Error; err != nil {
		return fmt.Errorf("failed to get rule: %w", err)
	}

	// 获取通知模板
	template := &model.NotifyTemplate{}
	if err := database.DB.First(template, "type = ?", rule.NotifyType).Error; err != nil {
		return fmt.Errorf("failed to get template: %w", err)
	}

	// 获取通知组
	group := &model.NotifyGroup{}
	if err := database.DB.First(group, "name = ?", rule.NotifyGroup).Error; err != nil {
		return fmt.Errorf("failed to get group: %w", err)
	}

	// 创建通知记录
	record := &model.NotifyRecord{
		AlertID: alert.ID,
		Type:    rule.NotifyType,
		Target:  group.Members,
		Content: template.Content, // TODO: 替换模板变量
		Status:  "pending",
	}

	if err := database.DB.Create(record).Error; err != nil {
		return fmt.Errorf("failed to create notify record: %w", err)
	}

	// TODO: 发送通知

	return nil
}
