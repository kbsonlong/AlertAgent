package repository

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"alert_agent/internal/domain/analysis"
	"alert_agent/internal/infrastructure/database"
)

// AnalysisRepositoryImpl 分析仓储实现
type AnalysisRepositoryImpl struct {
	db     *gorm.DB
	logger *zap.Logger
}

// NewAnalysisRepository 创建分析仓储
func NewAnalysisRepository(db *gorm.DB, logger *zap.Logger) analysis.AnalysisRepository {
	return &AnalysisRepositoryImpl{
		db:     db,
		logger: logger,
	}
}

// Create 创建分析记录
func (r *AnalysisRepositoryImpl) Create(ctx context.Context, record *analysis.AnalysisRecord) error {
	dbRecord := r.toDBModel(record)
	if err := r.db.WithContext(ctx).Create(dbRecord).Error; err != nil {
		return fmt.Errorf("failed to create analysis record: %w", err)
	}
	return nil
}

// Update 更新分析记录
func (r *AnalysisRepositoryImpl) Update(ctx context.Context, record *analysis.AnalysisRecord) error {
	dbRecord := r.toDBModel(record)
	if err := r.db.WithContext(ctx).Save(dbRecord).Error; err != nil {
		return fmt.Errorf("failed to update analysis record: %w", err)
	}
	return nil
}

// GetByID 根据ID获取分析记录
func (r *AnalysisRepositoryImpl) GetByID(ctx context.Context, id string) (*analysis.AnalysisRecord, error) {
	var dbRecord database.AIAnalysisRecord
	if err := r.db.WithContext(ctx).First(&dbRecord, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("analysis record not found")
		}
		return nil, fmt.Errorf("failed to get analysis record: %w", err)
	}
	return r.toDomainModel(&dbRecord), nil
}

// GetByAlertID 根据告警ID获取分析记录
func (r *AnalysisRepositoryImpl) GetByAlertID(ctx context.Context, alertID string) ([]*analysis.AnalysisRecord, error) {
	var dbRecords []database.AIAnalysisRecord
	if err := r.db.WithContext(ctx).Where("alert_id = ?", alertID).Order("created_at DESC").Find(&dbRecords).Error; err != nil {
		return nil, fmt.Errorf("failed to get analysis records by alert ID: %w", err)
	}

	records := make([]*analysis.AnalysisRecord, len(dbRecords))
	for i, dbRecord := range dbRecords {
		records[i] = r.toDomainModel(&dbRecord)
	}

	return records, nil
}

// List 列出分析记录
func (r *AnalysisRepositoryImpl) List(ctx context.Context, query *analysis.AnalysisQuery) ([]*analysis.AnalysisRecord, int64, error) {
	db := r.db.WithContext(ctx).Model(&database.AIAnalysisRecord{})

	// 应用过滤条件
	if query.AlertID != "" {
		db = db.Where("alert_id = ?", query.AlertID)
	}
	if query.AnalysisType != "" {
		db = db.Where("analysis_type = ?", query.AnalysisType)
	}
	if query.Status != "" {
		db = db.Where("status = ?", query.Status)
	}
	if query.Provider != "" {
		db = db.Where("provider = ?", query.Provider)
	}
	if !query.StartTime.IsZero() {
		db = db.Where("created_at >= ?", query.StartTime)
	}
	if !query.EndTime.IsZero() {
		db = db.Where("created_at <= ?", query.EndTime)
	}

	// 计算总数
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count analysis records: %w", err)
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
		db = db.Order("created_at DESC")
	}

	// 查询数据
	var dbRecords []database.AIAnalysisRecord
	if err := db.Find(&dbRecords).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list analysis records: %w", err)
	}

	// 转换为领域模型
	records := make([]*analysis.AnalysisRecord, len(dbRecords))
	for i, dbRecord := range dbRecords {
		records[i] = r.toDomainModel(&dbRecord)
	}

	return records, total, nil
}

// UpdateStatus 更新状态
func (r *AnalysisRepositoryImpl) UpdateStatus(ctx context.Context, id string, status analysis.AnalysisStatus) error {
	updates := map[string]interface{}{
		"status":     string(status),
		"updated_at": time.Now(),
	}

	if err := r.db.WithContext(ctx).Model(&database.AIAnalysisRecord{}).Where("id = ?", id).Updates(updates).Error; err != nil {
		return fmt.Errorf("failed to update analysis status: %w", err)
	}

	return nil
}

// GetPendingAnalysis 获取待处理的分析
func (r *AnalysisRepositoryImpl) GetPendingAnalysis(ctx context.Context, limit int) ([]*analysis.AnalysisRecord, error) {
	var dbRecords []database.AIAnalysisRecord
	query := r.db.WithContext(ctx).Where("status = ?", "pending").Order("created_at ASC")
	
	if limit > 0 {
		query = query.Limit(limit)
	}

	if err := query.Find(&dbRecords).Error; err != nil {
		return nil, fmt.Errorf("failed to get pending analysis: %w", err)
	}

	records := make([]*analysis.AnalysisRecord, len(dbRecords))
	for i, dbRecord := range dbRecords {
		records[i] = r.toDomainModel(&dbRecord)
	}

	return records, nil
}

// GetAnalysisStats 获取分析统计
func (r *AnalysisRepositoryImpl) GetAnalysisStats(ctx context.Context, startTime, endTime time.Time) (*analysis.AnalysisStats, error) {
	var stats analysis.AnalysisStats

	// 总分析数
	if err := r.db.WithContext(ctx).Model(&database.AIAnalysisRecord{}).
		Where("created_at BETWEEN ? AND ?", startTime, endTime).
		Count(&stats.TotalAnalysis).Error; err != nil {
		return nil, fmt.Errorf("failed to count total analysis: %w", err)
	}

	// 完成的分析数
	if err := r.db.WithContext(ctx).Model(&database.AIAnalysisRecord{}).
		Where("created_at BETWEEN ? AND ? AND status = ?", startTime, endTime, "completed").
		Count(&stats.CompletedAnalysis).Error; err != nil {
		return nil, fmt.Errorf("failed to count completed analysis: %w", err)
	}

	// 失败的分析数
	if err := r.db.WithContext(ctx).Model(&database.AIAnalysisRecord{}).
		Where("created_at BETWEEN ? AND ? AND status = ?", startTime, endTime, "failed").
		Count(&stats.FailedAnalysis).Error; err != nil {
		return nil, fmt.Errorf("failed to count failed analysis: %w", err)
	}

	// 平均处理时间
	var avgProcessingTime float64
	if err := r.db.WithContext(ctx).Model(&database.AIAnalysisRecord{}).
		Where("created_at BETWEEN ? AND ? AND status = ?", startTime, endTime, "completed").
		Select("AVG(processing_time)").Scan(&avgProcessingTime).Error; err != nil {
		return nil, fmt.Errorf("failed to calculate avg processing time: %w", err)
	}
	stats.AvgProcessingTime = avgProcessingTime

	// 平均置信度
	var avgConfidence float64
	if err := r.db.WithContext(ctx).Model(&database.AIAnalysisRecord{}).
		Where("created_at BETWEEN ? AND ? AND status = ?", startTime, endTime, "completed").
		Select("AVG(confidence_score)").Scan(&avgConfidence).Error; err != nil {
		return nil, fmt.Errorf("failed to calculate avg confidence: %w", err)
	}
	stats.AvgConfidence = avgConfidence

	// 成功率
	if stats.TotalAnalysis > 0 {
		stats.SuccessRate = float64(stats.CompletedAnalysis) / float64(stats.TotalAnalysis) * 100
	}

	return &stats, nil
}

// toDBModel 转换为数据库模型
func (r *AnalysisRepositoryImpl) toDBModel(record *analysis.AnalysisRecord) *database.AIAnalysisRecord {
	return &database.AIAnalysisRecord{
		ID:              record.ID,
		AlertID:         record.AlertID,
		AnalysisType:    record.AnalysisType,
		RequestData:     record.RequestData,
		ResponseData:    record.ResponseData,
		AnalysisResult:  record.AnalysisResult,
		ConfidenceScore: record.ConfidenceScore,
		ProcessingTime:  record.ProcessingTime,
		Status:          string(record.Status),
		ErrorMessage:    record.ErrorMessage,
		Provider:        record.Provider,
		ModelVersion:    record.ModelVersion,
		CreatedAt:       record.CreatedAt,
		UpdatedAt:       record.UpdatedAt,
	}
}

// toDomainModel 转换为领域模型
func (r *AnalysisRepositoryImpl) toDomainModel(dbRecord *database.AIAnalysisRecord) *analysis.AnalysisRecord {
	return &analysis.AnalysisRecord{
		ID:              dbRecord.ID,
		AlertID:         dbRecord.AlertID,
		AnalysisType:    dbRecord.AnalysisType,
		RequestData:     dbRecord.RequestData,
		ResponseData:    dbRecord.ResponseData,
		AnalysisResult:  dbRecord.AnalysisResult,
		ConfidenceScore: dbRecord.ConfidenceScore,
		ProcessingTime:  dbRecord.ProcessingTime,
		Status:          analysis.AnalysisStatus(dbRecord.Status),
		ErrorMessage:    dbRecord.ErrorMessage,
		Provider:        dbRecord.Provider,
		ModelVersion:    dbRecord.ModelVersion,
		CreatedAt:       dbRecord.CreatedAt,
		UpdatedAt:       dbRecord.UpdatedAt,
	}
}