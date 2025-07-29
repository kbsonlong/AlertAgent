package alert

import (
	"context"
	"strings"
	"time"

	"alert_agent/internal/domain/alert"
	"alert_agent/internal/model"

	"gorm.io/gorm"
)

// GORMAlertRepository GORM 实现的告警仓储
type GORMAlertRepository struct {
	db *gorm.DB
}

// NewGORMAlertRepository 创建新的 GORM 告警仓储
func NewGORMAlertRepository(db *gorm.DB) alert.AlertRepository {
	return &GORMAlertRepository{
		db: db,
	}
}

// Create 创建告警
func (r *GORMAlertRepository) Create(ctx context.Context, alertModel *model.Alert) error {
	return r.db.WithContext(ctx).Create(alertModel).Error
}

// Update 更新告警
func (r *GORMAlertRepository) Update(ctx context.Context, alertModel *model.Alert) error {
	return r.db.WithContext(ctx).Save(alertModel).Error
}

// UpdateByID 根据ID更新告警字段
func (r *GORMAlertRepository) UpdateByID(ctx context.Context, id uint, updates map[string]interface{}) error {
	return r.db.WithContext(ctx).Model(&model.Alert{}).Where("id = ?", id).Updates(updates).Error
}

// GetByID 根据ID获取告警
func (r *GORMAlertRepository) GetByID(ctx context.Context, id uint) (*model.Alert, error) {
	var alertModel model.Alert
	err := r.db.WithContext(ctx).First(&alertModel, id).Error
	if err != nil {
		return nil, err
	}
	return &alertModel, nil
}

// List 获取告警列表
func (r *GORMAlertRepository) List(ctx context.Context, filter alert.AlertFilter) ([]*model.Alert, error) {
	query := r.db.WithContext(ctx).Model(&model.Alert{})

	// 应用过滤条件
	query = r.applyFilter(query, filter)

	// 应用分页
	if filter.Limit > 0 {
		query = query.Limit(filter.Limit)
	}
	if filter.Offset > 0 {
		query = query.Offset(filter.Offset)
	}

	// 按创建时间倒序排列
	query = query.Order("created_at DESC")

	var alerts []*model.Alert
	err := query.Find(&alerts).Error
	return alerts, err
}

// Count 统计告警数量
func (r *GORMAlertRepository) Count(ctx context.Context, filter alert.AlertFilter) (int64, error) {
	query := r.db.WithContext(ctx).Model(&model.Alert{})
	query = r.applyFilter(query, filter)

	var count int64
	err := query.Count(&count).Error
	return count, err
}

// Delete 删除告警
func (r *GORMAlertRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&model.Alert{}, id).Error
}

// BatchUpdate 批量更新告警
func (r *GORMAlertRepository) BatchUpdate(ctx context.Context, ids []uint, updates map[string]interface{}) error {
	return r.db.WithContext(ctx).Model(&model.Alert{}).Where("id IN ?", ids).Updates(updates).Error
}

// GetSimilarAlerts 获取相似告警
func (r *GORMAlertRepository) GetSimilarAlerts(ctx context.Context, alertModel *model.Alert, limit int) ([]*model.Alert, error) {
	query := r.db.WithContext(ctx).Model(&model.Alert{}).Where("id != ?", alertModel.ID)

	// 基于标题和内容的相似性查找
	if alertModel.Title != "" {
		query = query.Where("title LIKE ?", "%"+alertModel.Title+"%")
	}
	if alertModel.Source != "" {
		query = query.Where("source = ?", alertModel.Source)
	}
	if alertModel.Level != "" {
		query = query.Where("level = ?", alertModel.Level)
	}

	query = query.Order("created_at DESC").Limit(limit)

	var alerts []*model.Alert
	err := query.Find(&alerts).Error
	return alerts, err
}

// GetStatistics 获取告警统计信息
func (r *GORMAlertRepository) GetStatistics(ctx context.Context, filter alert.AlertFilter) (*alert.AlertStatistics, error) {
	stats := &alert.AlertStatistics{
		ByStatus:   make(map[string]int64),
		BySeverity: make(map[string]int64),
		BySource:   make(map[string]int64),
	}

	// 总数统计
	total, err := r.Count(ctx, filter)
	if err != nil {
		return nil, err
	}
	stats.Total = total

	// 按状态统计
	var statusStats []struct {
		Status string `json:"status"`
		Count  int64  `json:"count"`
	}
	query := r.db.WithContext(ctx).Model(&model.Alert{}).Select("status, COUNT(*) as count").Group("status")
	query = r.applyFilter(query, filter)
	if err := query.Find(&statusStats).Error; err != nil {
		return nil, err
	}
	for _, stat := range statusStats {
		stats.ByStatus[stat.Status] = stat.Count
	}

	// 按严重程度统计
	var severityStats []struct {
		Severity string `json:"severity"`
		Count    int64  `json:"count"`
	}
	query = r.db.WithContext(ctx).Model(&model.Alert{}).Select("severity, COUNT(*) as count").Group("severity")
	query = r.applyFilter(query, filter)
	if err := query.Find(&severityStats).Error; err != nil {
		return nil, err
	}
	for _, stat := range severityStats {
		stats.BySeverity[stat.Severity] = stat.Count
	}

	// 按来源统计
	var sourceStats []struct {
		Source string `json:"source"`
		Count  int64  `json:"count"`
	}
	query = r.db.WithContext(ctx).Model(&model.Alert{}).Select("source, COUNT(*) as count").Group("source")
	query = r.applyFilter(query, filter)
	if err := query.Find(&sourceStats).Error; err != nil {
		return nil, err
	}
	for _, stat := range sourceStats {
		stats.BySource[stat.Source] = stat.Count
	}

	return stats, nil
}

// GetRecentAlerts 获取最近的告警
func (r *GORMAlertRepository) GetRecentAlerts(ctx context.Context, duration time.Duration, limit int) ([]*model.Alert, error) {
	since := time.Now().Add(-duration)
	var alerts []*model.Alert
	err := r.db.WithContext(ctx).
		Where("created_at >= ?", since).
		Order("created_at DESC").
		Limit(limit).
		Find(&alerts).Error
	return alerts, err
}

// UpdateAnalysisResult 更新告警分析结果
func (r *GORMAlertRepository) UpdateAnalysisResult(ctx context.Context, alertID uint, analysis string) error {
	return r.db.WithContext(ctx).
		Model(&model.Alert{}).
		Where("id = ?", alertID).
		Update("analysis", analysis).Error
}

// GetAlertsForAnalysis 获取需要分析的告警
func (r *GORMAlertRepository) GetAlertsForAnalysis(ctx context.Context, limit int) ([]*model.Alert, error) {
	var alerts []*model.Alert
	err := r.db.WithContext(ctx).
		Where("analysis IS NULL OR analysis = ''").
		Where("status != ?", model.AlertStatusResolved).
		Order("created_at DESC").
		Limit(limit).
		Find(&alerts).Error
	return alerts, err
}

// MarkAsAnalyzed 标记告警为已分析
func (r *GORMAlertRepository) MarkAsAnalyzed(ctx context.Context, alertID uint) error {
	return r.db.WithContext(ctx).
		Model(&model.Alert{}).
		Where("id = ?", alertID).
		Update("updated_at", time.Now()).Error
}

// applyFilter 应用过滤条件
func (r *GORMAlertRepository) applyFilter(query *gorm.DB, filter alert.AlertFilter) *gorm.DB {
	// 状态过滤
	if len(filter.Status) > 0 {
		query = query.Where("status IN ?", filter.Status)
	}

	// 严重程度过滤
	if len(filter.Severity) > 0 {
		query = query.Where("severity IN ?", filter.Severity)
	}

	// 来源过滤
	if len(filter.Source) > 0 {
		query = query.Where("source IN ?", filter.Source)
	}

	// 规则ID过滤
	if filter.RuleID != nil {
		query = query.Where("rule_id = ?", *filter.RuleID)
	}

	// 创建时间过滤
	if filter.CreatedAt != nil {
		if filter.CreatedAt.Start != nil {
			query = query.Where("created_at >= ?", *filter.CreatedAt.Start)
		}
		if filter.CreatedAt.End != nil {
			query = query.Where("created_at <= ?", *filter.CreatedAt.End)
		}
	}

	// 更新时间过滤
	if filter.UpdatedAt != nil {
		if filter.UpdatedAt.Start != nil {
			query = query.Where("updated_at >= ?", *filter.UpdatedAt.Start)
		}
		if filter.UpdatedAt.End != nil {
			query = query.Where("updated_at <= ?", *filter.UpdatedAt.End)
		}
	}

	// 关键词搜索
	if filter.Keywords != "" {
		keywords := strings.TrimSpace(filter.Keywords)
		if keywords != "" {
			searchPattern := "%" + keywords + "%"
			query = query.Where("title LIKE ? OR content LIKE ? OR source LIKE ?", 
				searchPattern, searchPattern, searchPattern)
		}
	}

	return query
}