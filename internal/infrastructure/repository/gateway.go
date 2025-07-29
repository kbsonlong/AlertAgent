package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"alert_agent/internal/domain/gateway"
	"alert_agent/internal/infrastructure/database"
)

// GatewayRepositoryImpl 网关仓储实现
type GatewayRepositoryImpl struct {
	db     *gorm.DB
	redis  *redis.Client
	logger *zap.Logger
}

// NewGatewayRepository 创建网关仓储
func NewGatewayRepository(db *gorm.DB, redis *redis.Client, logger *zap.Logger) gateway.GatewayRepository {
	return &GatewayRepositoryImpl{
		db:     db,
		redis:  redis,
		logger: logger,
	}
}

// CreateProcessingRecord 创建处理记录
func (r *GatewayRepositoryImpl) CreateProcessingRecord(ctx context.Context, record *gateway.ProcessingRecord) error {
	dbRecord := r.toProcessingDBModel(record)
	if err := r.db.WithContext(ctx).Create(dbRecord).Error; err != nil {
		return fmt.Errorf("failed to create processing record: %w", err)
	}
	return nil
}

// UpdateProcessingRecord 更新处理记录
func (r *GatewayRepositoryImpl) UpdateProcessingRecord(ctx context.Context, record *gateway.ProcessingRecord) error {
	dbRecord := r.toProcessingDBModel(record)
	if err := r.db.WithContext(ctx).Save(dbRecord).Error; err != nil {
		return fmt.Errorf("failed to update processing record: %w", err)
	}
	return nil
}

// GetProcessingRecord 获取处理记录
func (r *GatewayRepositoryImpl) GetProcessingRecord(ctx context.Context, alertID string) (*gateway.ProcessingRecord, error) {
	var dbRecord database.AlertProcessingRecord
	if err := r.db.WithContext(ctx).First(&dbRecord, "alert_id = ?", alertID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("processing record not found")
		}
		return nil, fmt.Errorf("failed to get processing record: %w", err)
	}
	return r.toProcessingDomainModel(&dbRecord), nil
}

// ListProcessingRecords 列出处理记录
func (r *GatewayRepositoryImpl) ListProcessingRecords(ctx context.Context, query *gateway.ProcessingQuery) ([]*gateway.ProcessingRecord, int64, error) {
	db := r.db.WithContext(ctx).Model(&database.AlertProcessingRecord{})

	// 应用过滤条件
	if query.AlertName != "" {
		db = db.Where("alert_name LIKE ?", "%"+query.AlertName+"%")
	}
	if query.Severity != "" {
		db = db.Where("severity = ?", query.Severity)
	}
	if query.ClusterID != "" {
		db = db.Where("cluster_id = ?", query.ClusterID)
	}
	if query.Status != "" {
		db = db.Where("processing_status = ?", query.Status)
	}
	if !query.StartTime.IsZero() {
		db = db.Where("received_at >= ?", query.StartTime)
	}
	if !query.EndTime.IsZero() {
		db = db.Where("received_at <= ?", query.EndTime)
	}

	// 计算总数
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count processing records: %w", err)
	}

	// 应用分页
	if query.Page > 0 && query.PageSize > 0 {
		offset := (query.Page - 1) * query.PageSize
		db = db.Offset(offset).Limit(query.PageSize)
	}

	// 应用排序
	if query.SortBy != "" {
		order := query.SortBy
		if query.SortDesc {
			order += " DESC"
		}
		db = db.Order(order)
	} else {
		db = db.Order("received_at DESC")
	}

	// 查询数据
	var dbRecords []database.AlertProcessingRecord
	if err := db.Find(&dbRecords).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list processing records: %w", err)
	}

	// 转换为领域模型
	records := make([]*gateway.ProcessingRecord, len(dbRecords))
	for i, dbRecord := range dbRecords {
		records[i] = r.toProcessingDomainModel(&dbRecord)
	}

	return records, total, nil
}

// CreateConvergenceRecord 创建收敛记录
func (r *GatewayRepositoryImpl) CreateConvergenceRecord(ctx context.Context, record *gateway.ConvergenceRecord) error {
	dbRecord := r.toConvergenceDBModel(record)
	if err := r.db.WithContext(ctx).Create(dbRecord).Error; err != nil {
		return fmt.Errorf("failed to create convergence record: %w", err)
	}
	return nil
}

// UpdateConvergenceRecord 更新收敛记录
func (r *GatewayRepositoryImpl) UpdateConvergenceRecord(ctx context.Context, record *gateway.ConvergenceRecord) error {
	dbRecord := r.toConvergenceDBModel(record)
	if err := r.db.WithContext(ctx).Save(dbRecord).Error; err != nil {
		return fmt.Errorf("failed to update convergence record: %w", err)
	}
	return nil
}

// GetConvergenceWindow 获取收敛窗口
func (r *GatewayRepositoryImpl) GetConvergenceWindow(ctx context.Context, key string) (*gateway.ConvergenceWindow, error) {
	redisKey := fmt.Sprintf("convergence_window:%s", key)
	data, err := r.redis.Get(ctx, redisKey).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, fmt.Errorf("convergence window not found")
		}
		return nil, fmt.Errorf("failed to get convergence window: %w", err)
	}

	var window gateway.ConvergenceWindow
	if err := json.Unmarshal([]byte(data), &window); err != nil {
		return nil, fmt.Errorf("failed to unmarshal convergence window: %w", err)
	}

	return &window, nil
}

// SetConvergenceWindow 设置收敛窗口
func (r *GatewayRepositoryImpl) SetConvergenceWindow(ctx context.Context, window *gateway.ConvergenceWindow, ttl time.Duration) error {
	redisKey := fmt.Sprintf("convergence_window:%s", window.Key)
	data, err := json.Marshal(window)
	if err != nil {
		return fmt.Errorf("failed to marshal convergence window: %w", err)
	}

	if err := r.redis.Set(ctx, redisKey, data, ttl).Err(); err != nil {
		return fmt.Errorf("failed to set convergence window: %w", err)
	}

	return nil
}

// CreateSuppressionRule 创建抑制规则
func (r *GatewayRepositoryImpl) CreateSuppressionRule(ctx context.Context, rule *gateway.SuppressionRule) error {
	dbRule := r.toSuppressionDBModel(rule)
	if err := r.db.WithContext(ctx).Create(dbRule).Error; err != nil {
		return fmt.Errorf("failed to create suppression rule: %w", err)
	}
	return nil
}

// UpdateSuppressionRule 更新抑制规则
func (r *GatewayRepositoryImpl) UpdateSuppressionRule(ctx context.Context, rule *gateway.SuppressionRule) error {
	dbRule := r.toSuppressionDBModel(rule)
	if err := r.db.WithContext(ctx).Save(dbRule).Error; err != nil {
		return fmt.Errorf("failed to update suppression rule: %w", err)
	}
	return nil
}

// DeleteSuppressionRule 删除抑制规则
func (r *GatewayRepositoryImpl) DeleteSuppressionRule(ctx context.Context, id string) error {
	if err := r.db.WithContext(ctx).Delete(&database.SuppressionRule{}, "id = ?", id).Error; err != nil {
		return fmt.Errorf("failed to delete suppression rule: %w", err)
	}
	return nil
}

// GetSuppressionRule 获取抑制规则
func (r *GatewayRepositoryImpl) GetSuppressionRule(ctx context.Context, id string) (*gateway.SuppressionRule, error) {
	var dbRule database.SuppressionRule
	if err := r.db.WithContext(ctx).First(&dbRule, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("suppression rule not found")
		}
		return nil, fmt.Errorf("failed to get suppression rule: %w", err)
	}
	return r.toSuppressionDomainModel(&dbRule), nil
}

// ListSuppressionRules 列出抑制规则
func (r *GatewayRepositoryImpl) ListSuppressionRules(ctx context.Context, enabled *bool) ([]*gateway.SuppressionRule, error) {
	db := r.db.WithContext(ctx).Model(&database.SuppressionRule{})
	
	if enabled != nil {
		db = db.Where("enabled = ?", *enabled)
	}

	var dbRules []database.SuppressionRule
	if err := db.Order("priority DESC, created_at DESC").Find(&dbRules).Error; err != nil {
		return nil, fmt.Errorf("failed to list suppression rules: %w", err)
	}

	rules := make([]*gateway.SuppressionRule, len(dbRules))
	for i, dbRule := range dbRules {
		rules[i] = r.toSuppressionDomainModel(&dbRule)
	}

	return rules, nil
}

// CreateRoutingRule 创建路由规则
func (r *GatewayRepositoryImpl) CreateRoutingRule(ctx context.Context, rule *gateway.RoutingRule) error {
	dbRule := r.toRoutingDBModel(rule)
	if err := r.db.WithContext(ctx).Create(dbRule).Error; err != nil {
		return fmt.Errorf("failed to create routing rule: %w", err)
	}
	return nil
}

// UpdateRoutingRule 更新路由规则
func (r *GatewayRepositoryImpl) UpdateRoutingRule(ctx context.Context, rule *gateway.RoutingRule) error {
	dbRule := r.toRoutingDBModel(rule)
	if err := r.db.WithContext(ctx).Save(dbRule).Error; err != nil {
		return fmt.Errorf("failed to update routing rule: %w", err)
	}
	return nil
}

// DeleteRoutingRule 删除路由规则
func (r *GatewayRepositoryImpl) DeleteRoutingRule(ctx context.Context, id string) error {
	if err := r.db.WithContext(ctx).Delete(&database.RoutingRule{}, "id = ?", id).Error; err != nil {
		return fmt.Errorf("failed to delete routing rule: %w", err)
	}
	return nil
}

// GetRoutingRule 获取路由规则
func (r *GatewayRepositoryImpl) GetRoutingRule(ctx context.Context, id string) (*gateway.RoutingRule, error) {
	var dbRule database.RoutingRule
	if err := r.db.WithContext(ctx).First(&dbRule, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("routing rule not found")
		}
		return nil, fmt.Errorf("failed to get routing rule: %w", err)
	}
	return r.toRoutingDomainModel(&dbRule), nil
}

// ListRoutingRules 列出路由规则
func (r *GatewayRepositoryImpl) ListRoutingRules(ctx context.Context, enabled *bool) ([]*gateway.RoutingRule, error) {
	db := r.db.WithContext(ctx).Model(&database.RoutingRule{})
	
	if enabled != nil {
		db = db.Where("enabled = ?", *enabled)
	}

	var dbRules []database.RoutingRule
	if err := db.Order("priority DESC, created_at DESC").Find(&dbRules).Error; err != nil {
		return nil, fmt.Errorf("failed to list routing rules: %w", err)
	}

	rules := make([]*gateway.RoutingRule, len(dbRules))
	for i, dbRule := range dbRules {
		rules[i] = r.toRoutingDomainModel(&dbRule)
	}

	return rules, nil
}

// 转换方法
func (r *GatewayRepositoryImpl) toProcessingDBModel(record *gateway.ProcessingRecord) *database.AlertProcessingRecord {
	return &database.AlertProcessingRecord{
		ID:               record.ID,
		AlertID:          record.AlertID,
		AlertName:        record.AlertName,
		Severity:         record.Severity,
		ClusterID:        record.ClusterID,
		ReceivedAt:       record.ReceivedAt,
		ProcessedAt:      record.ProcessedAt,
		ProcessingStatus: record.ProcessingStatus,
		AnalysisID:       record.AnalysisID,
		Decision:         record.Decision,
		ActionTaken:      record.ActionTaken,
		ResolutionTime:   record.ResolutionTime,
		FeedbackScore:    record.FeedbackScore,
		Labels:           record.Labels,
		Annotations:      record.Annotations,
		ChannelsSent:     record.ChannelsSent,
		ErrorMessage:     record.ErrorMessage,
		CreatedAt:        record.CreatedAt,
		UpdatedAt:        record.UpdatedAt,
	}
}

func (r *GatewayRepositoryImpl) toProcessingDomainModel(dbRecord *database.AlertProcessingRecord) *gateway.ProcessingRecord {
	return &gateway.ProcessingRecord{
		ID:               dbRecord.ID,
		AlertID:          dbRecord.AlertID,
		AlertName:        dbRecord.AlertName,
		Severity:         dbRecord.Severity,
		ClusterID:        dbRecord.ClusterID,
		ReceivedAt:       dbRecord.ReceivedAt,
		ProcessedAt:      dbRecord.ProcessedAt,
		ProcessingStatus: dbRecord.ProcessingStatus,
		AnalysisID:       dbRecord.AnalysisID,
		Decision:         dbRecord.Decision,
		ActionTaken:      dbRecord.ActionTaken,
		ResolutionTime:   dbRecord.ResolutionTime,
		FeedbackScore:    dbRecord.FeedbackScore,
		Labels:           dbRecord.Labels,
		Annotations:      dbRecord.Annotations,
		ChannelsSent:     dbRecord.ChannelsSent,
		ErrorMessage:     dbRecord.ErrorMessage,
		CreatedAt:        dbRecord.CreatedAt,
		UpdatedAt:        dbRecord.UpdatedAt,
	}
}

func (r *GatewayRepositoryImpl) toConvergenceDBModel(record *gateway.ConvergenceRecord) *database.ConvergenceRecord {
	return &database.ConvergenceRecord{
		ID:                  record.ID,
		ConvergenceKey:      record.ConvergenceKey,
		WindowStart:         record.WindowStart,
		WindowEnd:           record.WindowEnd,
		AlertCount:          record.AlertCount,
		Status:              record.Status,
		TriggerCondition:    record.TriggerCondition,
		RepresentativeAlert: record.RepresentativeAlert,
		ConvergedAlerts:     record.ConvergedAlerts,
		TriggeredAt:         record.TriggeredAt,
		CreatedAt:           record.CreatedAt,
		UpdatedAt:           record.UpdatedAt,
	}
}

func (r *GatewayRepositoryImpl) toSuppressionDBModel(rule *gateway.SuppressionRule) *database.SuppressionRule {
	return &database.SuppressionRule{
		ID:          rule.ID,
		Name:        rule.Name,
		Description: rule.Description,
		Enabled:     rule.Enabled,
		Priority:    rule.Priority,
		Conditions:  rule.Conditions,
		Schedule:    rule.Schedule,
		CreatedBy:   rule.CreatedBy,
		UpdatedBy:   rule.UpdatedBy,
		CreatedAt:   rule.CreatedAt,
		UpdatedAt:   rule.UpdatedAt,
	}
}

func (r *GatewayRepositoryImpl) toSuppressionDomainModel(dbRule *database.SuppressionRule) *gateway.SuppressionRule {
	return &gateway.SuppressionRule{
		ID:          dbRule.ID,
		Name:        dbRule.Name,
		Description: dbRule.Description,
		Enabled:     dbRule.Enabled,
		Priority:    dbRule.Priority,
		Conditions:  dbRule.Conditions,
		Schedule:    dbRule.Schedule,
		CreatedBy:   dbRule.CreatedBy,
		UpdatedBy:   dbRule.UpdatedBy,
		CreatedAt:   dbRule.CreatedAt,
		UpdatedAt:   dbRule.UpdatedAt,
	}
}

func (r *GatewayRepositoryImpl) toRoutingDBModel(rule *gateway.RoutingRule) *database.RoutingRule {
	return &database.RoutingRule{
		ID:          rule.ID,
		Name:        rule.Name,
		Description: rule.Description,
		Enabled:     rule.Enabled,
		Priority:    rule.Priority,
		Conditions:  rule.Conditions,
		Actions:     rule.Actions,
		ChannelIDs:  rule.ChannelIDs,
		CreatedBy:   rule.CreatedBy,
		UpdatedBy:   rule.UpdatedBy,
		CreatedAt:   rule.CreatedAt,
		UpdatedAt:   rule.UpdatedAt,
	}
}

func (r *GatewayRepositoryImpl) toRoutingDomainModel(dbRule *database.RoutingRule) *gateway.RoutingRule {
	return &gateway.RoutingRule{
		ID:          dbRule.ID,
		Name:        dbRule.Name,
		Description: dbRule.Description,
		Enabled:     dbRule.Enabled,
		Priority:    dbRule.Priority,
		Conditions:  dbRule.Conditions,
		Actions:     dbRule.Actions,
		ChannelIDs:  dbRule.ChannelIDs,
		CreatedBy:   dbRule.CreatedBy,
		UpdatedBy:   dbRule.UpdatedBy,
		CreatedAt:   dbRule.CreatedAt,
		UpdatedAt:   dbRule.UpdatedAt,
	}
}