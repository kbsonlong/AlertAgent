package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"alert_agent/internal/model"

	goredis "github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

var (
	ErrAlertNotFound = errors.New("alert not found")
	ErrInvalidAlert  = errors.New("invalid alert data")
)

const (
	AlertCacheKeyPrefix = "alert:"
	AlertCacheTTL       = 24 * time.Hour
)

// AlertService 告警服务
type AlertService struct {
	db    *gorm.DB
	cache *goredis.Client
}

// NewAlertService 创建告警服务实例
func NewAlertService(db *gorm.DB, cache *goredis.Client) *AlertService {
	return &AlertService{
		db:    db,
		cache: cache,
	}
}

// CreateAlert 创建告警
func (s *AlertService) CreateAlert(ctx context.Context, alert *model.Alert) error {
	// 验证告警数据
	if err := alert.Validate(); err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidAlert, err)
	}

	// 设置初始状态
	alert.Status = model.AlertStatusNew

	// 开启事务
	tx := s.db.WithContext(ctx).Begin()
	if err := tx.Error; err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// 创建告警记录
	if err := tx.Create(alert).Error; err != nil {
		return fmt.Errorf("failed to create alert: %w", err)
	}

	// 缓存告警信息
	if err := s.cacheAlert(ctx, alert); err != nil {
		return fmt.Errorf("failed to cache alert: %w", err)
	}

	// 触发告警通知
	if err := s.triggerNotification(ctx, tx, alert); err != nil {
		return fmt.Errorf("failed to trigger notification: %w", err)
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
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
	if err := s.db.WithContext(ctx).First(alert, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrAlertNotFound
		}
		return nil, fmt.Errorf("failed to get alert: %w", err)
	}

	// 更新缓存
	if err := s.cacheAlert(ctx, alert); err != nil {
		return nil, fmt.Errorf("failed to cache alert: %w", err)
	}

	return alert, nil
}

// UpdateAlertStatus 更新告警状态
func (s *AlertService) UpdateAlertStatus(ctx context.Context, id uint, status string, handler string, note string) error {
	// 验证状态
	if !isValidAlertStatus(status) {
		return fmt.Errorf("%w: invalid status", ErrInvalidAlert)
	}

	// 开启事务
	tx := s.db.WithContext(ctx).Begin()
	if err := tx.Error; err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// 获取告警
	alert := &model.Alert{}
	if err := tx.First(alert, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrAlertNotFound
		}
		return fmt.Errorf("failed to get alert: %w", err)
	}

	// 更新状态
	now := time.Now()
	alert.Status = status
	alert.Handler = handler
	alert.HandleNote = note
	alert.HandleTime = &now

	// 保存更新
	if err := tx.Save(alert).Error; err != nil {
		return fmt.Errorf("failed to update alert: %w", err)
	}

	// 更新缓存
	if err := s.cacheAlert(ctx, alert); err != nil {
		return fmt.Errorf("failed to cache alert: %w", err)
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// ListAlerts 获取告警列表
func (s *AlertService) ListAlerts(ctx context.Context, query *AlertQuery) ([]*model.Alert, int64, error) {
	db := s.db.WithContext(ctx).Model(&model.Alert{})

	// 应用查询条件
	if query != nil {
		if query.Status != "" {
			if !isValidAlertStatus(query.Status) {
				return nil, 0, fmt.Errorf("%w: invalid status", ErrInvalidAlert)
			}
			db = db.Where("status = ?", query.Status)
		}
		if query.Level != "" {
			if !isValidAlertLevel(query.Level) {
				return nil, 0, fmt.Errorf("%w: invalid level", ErrInvalidAlert)
			}
			db = db.Where("level = ?", query.Level)
		}
		if query.Source != "" {
			db = db.Where("source = ?", query.Source)
		}
		if query.StartTime != nil {
			db = db.Where("created_at >= ?", query.StartTime)
		}
		if query.EndTime != nil {
			db = db.Where("created_at <= ?", query.EndTime)
		}
	}

	// 获取总数
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count alerts: %w", err)
	}

	// 获取分页数据
	var alerts []*model.Alert
	if err := db.Offset(query.GetOffset()).Limit(query.GetLimit()).Order("id DESC").Find(&alerts).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list alerts: %w", err)
	}

	return alerts, total, nil
}

// cacheAlert 缓存告警信息
func (s *AlertService) cacheAlert(ctx context.Context, alert *model.Alert) error {
	key := fmt.Sprintf("%s%d", AlertCacheKeyPrefix, alert.ID)
	data, err := json.Marshal(alert)
	if err != nil {
		return fmt.Errorf("failed to marshal alert: %w", err)
	}
	return s.cache.Set(ctx, key, string(data), AlertCacheTTL).Err()
}

// getAlertFromCache 从缓存获取告警信息
func (s *AlertService) getAlertFromCache(ctx context.Context, id uint) (*model.Alert, error) {
	key := fmt.Sprintf("%s%d", AlertCacheKeyPrefix, id)
	data, err := s.cache.Get(ctx, key).Result()
	if err != nil {
		return nil, err
	}

	alert := &model.Alert{}
	if err := json.Unmarshal([]byte(data), alert); err != nil {
		return nil, fmt.Errorf("failed to unmarshal alert: %w", err)
	}

	return alert, nil
}

// triggerNotification 触发告警通知
func (s *AlertService) triggerNotification(ctx context.Context, tx *gorm.DB, alert *model.Alert) error {
	// 获取告警规则
	rule := &model.Rule{}
	if err := tx.First(rule, alert.RuleID).Error; err != nil {
		return fmt.Errorf("failed to get rule: %w", err)
	}

	// 获取通知模板
	template := &model.NotifyTemplate{}
	if err := tx.First(template, "type = ? AND enabled = ?", rule.NotifyType, true).Error; err != nil {
		return fmt.Errorf("failed to get template: %w", err)
	}

	// 获取通知组
	group := &model.NotifyGroup{}
	if err := tx.First(group, "name = ? AND enabled = ?", rule.NotifyGroup, true).Error; err != nil {
		return fmt.Errorf("failed to get group: %w", err)
	}

	// 创建通知记录
	record := &model.NotifyRecord{
		AlertID: alert.ID,
		Type:    template.Type,
		Target:  group.Members,
		Content: s.renderTemplate(template, alert),
		Status:  model.NotifyStatusPending,
	}

	if err := record.Validate(); err != nil {
		return fmt.Errorf("invalid notify record: %w", err)
	}

	if err := tx.Create(record).Error; err != nil {
		return fmt.Errorf("failed to create notify record: %w", err)
	}

	// 更新告警通知时间
	now := time.Now()
	alert.NotifyTime = &now
	alert.NotifyCount++
	if err := tx.Save(alert).Error; err != nil {
		return fmt.Errorf("failed to update alert notify info: %w", err)
	}

	return nil
}

// renderTemplate 渲染通知模板
func (s *AlertService) renderTemplate(template *model.NotifyTemplate, alert *model.Alert) string {
	// TODO: 实现模板渲染逻辑
	return template.Content
}

// AlertQuery 告警查询参数
type AlertQuery struct {
	Status    string     `json:"status"`
	Level     string     `json:"level"`
	Source    string     `json:"source"`
	StartTime *time.Time `json:"start_time"`
	EndTime   *time.Time `json:"end_time"`
	Page      int        `json:"page"`
	PageSize  int        `json:"page_size"`
}

// GetOffset 获取分页偏移量
func (q *AlertQuery) GetOffset() int {
	if q.Page <= 0 {
		q.Page = 1
	}
	return (q.Page - 1) * q.GetLimit()
}

// GetLimit 获取分页大小
func (q *AlertQuery) GetLimit() int {
	if q.PageSize <= 0 {
		q.PageSize = 20
	}
	if q.PageSize > 100 {
		q.PageSize = 100
	}
	return q.PageSize
}

// isValidAlertStatus 验证告警状态
func isValidAlertStatus(status string) bool {
	switch status {
	case model.AlertStatusNew, model.AlertStatusAcknowledged, model.AlertStatusResolved:
		return true
	}
	return false
}

// isValidAlertLevel 验证告警级别
func isValidAlertLevel(level string) bool {
	switch level {
	case model.AlertLevelCritical, model.AlertLevelHigh, model.AlertLevelMedium, model.AlertLevelLow:
		return true
	}
	return false
}
