package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"gorm.io/gorm"

	"alert_agent/internal/domain/analysis"
)

// NewAnalysisResultRepository 创建分析结果存储实例
func NewAnalysisResultRepository(db *gorm.DB) analysis.AnalysisResultRepository {
	return &AnalysisResultRepositoryImpl{
		db: db,
	}
}

// AnalysisResultRepositoryImpl 分析结果存储实现
type AnalysisResultRepositoryImpl struct {
	db *gorm.DB
}

// analysisResultModel 分析结果数据库模型
type analysisResultModel struct {
	ID              string    `gorm:"primaryKey;type:varchar(255)" json:"id"`
	TaskID          string    `gorm:"index;not null;type:varchar(255)" json:"task_id"`
	AlertID         string    `gorm:"index;not null;type:varchar(255)" json:"alert_id"`
	Type            string    `gorm:"type:varchar(50);not null" json:"type"`
	Status          string    `gorm:"type:varchar(20);not null;index" json:"status"`
	ConfidenceScore float64   `gorm:"not null;default:0" json:"confidence_score"`
	ProcessingTime  int64     `gorm:"not null;default:0" json:"processing_time_ms"`
	ResultJSON      string    `gorm:"type:text" json:"result_json"`
	Summary         string    `gorm:"type:text" json:"summary"`
	Recommendations string    `gorm:"type:text" json:"recommendations_json"`
	ErrorMessage    string    `gorm:"type:text" json:"error_message"`
	CreatedAt       time.Time `gorm:"not null" json:"created_at"`
	UpdatedAt       time.Time `gorm:"not null" json:"updated_at"`
	MetadataJSON    string    `gorm:"type:text" json:"metadata_json"`
}

// TableName 设置表名
func (analysisResultModel) TableName() string {
	return "analysis_results"
}

// toEntity 转换为领域实体
func (m *analysisResultModel) toEntity() (*analysis.AnalysisResult, error) {
	result := &analysis.AnalysisResult{
		ID:              m.ID,
		TaskID:          m.TaskID,
		AlertID:         m.AlertID,
		Type:            analysis.AnalysisType(m.Type),
		Status:          analysis.AnalysisStatus(m.Status),
		ConfidenceScore: m.ConfidenceScore,
		ProcessingTime:  time.Duration(m.ProcessingTime) * time.Millisecond,
		Summary:         m.Summary,
		ErrorMessage:    m.ErrorMessage,
		CreatedAt:       m.CreatedAt,
		UpdatedAt:       m.UpdatedAt,
	}

	// 解析结果JSON
	if m.ResultJSON != "" {
		if err := json.Unmarshal([]byte(m.ResultJSON), &result.Result); err != nil {
			return nil, fmt.Errorf("failed to unmarshal result JSON: %w", err)
		}
	} else {
		result.Result = make(map[string]interface{})
	}

	// 解析推荐JSON
	if m.Recommendations != "" {
		if err := json.Unmarshal([]byte(m.Recommendations), &result.Recommendations); err != nil {
			return nil, fmt.Errorf("failed to unmarshal recommendations JSON: %w", err)
		}
	} else {
		result.Recommendations = []string{}
	}

	// 解析元数据JSON
	if m.MetadataJSON != "" {
		if err := json.Unmarshal([]byte(m.MetadataJSON), &result.Metadata); err != nil {
			return nil, fmt.Errorf("failed to unmarshal metadata JSON: %w", err)
		}
	} else {
		result.Metadata = make(map[string]interface{})
	}

	return result, nil
}

// fromEntity 从领域实体转换
func (m *analysisResultModel) fromEntity(result *analysis.AnalysisResult) error {
	m.ID = result.ID
	m.TaskID = result.TaskID
	m.AlertID = result.AlertID
	m.Type = string(result.Type)
	m.Status = string(result.Status)
	m.ConfidenceScore = result.ConfidenceScore
	m.ProcessingTime = int64(result.ProcessingTime / time.Millisecond)
	m.Summary = result.Summary
	m.ErrorMessage = result.ErrorMessage
	m.CreatedAt = result.CreatedAt
	m.UpdatedAt = result.UpdatedAt

	// 序列化结果JSON
	if result.Result != nil {
		resultJSON, err := json.Marshal(result.Result)
		if err != nil {
			return fmt.Errorf("failed to marshal result JSON: %w", err)
		}
		m.ResultJSON = string(resultJSON)
	}

	// 序列化推荐JSON
	if result.Recommendations != nil {
		recommendationsJSON, err := json.Marshal(result.Recommendations)
		if err != nil {
			return fmt.Errorf("failed to marshal recommendations JSON: %w", err)
		}
		m.Recommendations = string(recommendationsJSON)
	}

	// 序列化元数据JSON
	if result.Metadata != nil {
		metadataJSON, err := json.Marshal(result.Metadata)
		if err != nil {
			return fmt.Errorf("failed to marshal metadata JSON: %w", err)
		}
		m.MetadataJSON = string(metadataJSON)
	}

	return nil
}

// Create 创建结果
func (r *AnalysisResultRepositoryImpl) Create(ctx context.Context, result *analysis.AnalysisResult) error {
	model := &analysisResultModel{}
	if err := model.fromEntity(result); err != nil {
		return fmt.Errorf("failed to convert entity to model: %w", err)
	}

	if err := r.db.WithContext(ctx).Create(model).Error; err != nil {
		return fmt.Errorf("failed to create analysis result: %w", err)
	}

	return nil
}

// GetByID 根据ID获取结果
func (r *AnalysisResultRepositoryImpl) GetByID(ctx context.Context, resultID string) (*analysis.AnalysisResult, error) {
	var model analysisResultModel
	if err := r.db.WithContext(ctx).Where("id = ?", resultID).First(&model).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("analysis result not found: %s", resultID)
		}
		return nil, fmt.Errorf("failed to get analysis result: %w", err)
	}

	return model.toEntity()
}

// GetByTaskID 根据任务ID获取结果
func (r *AnalysisResultRepositoryImpl) GetByTaskID(ctx context.Context, taskID string) (*analysis.AnalysisResult, error) {
	var model analysisResultModel
	if err := r.db.WithContext(ctx).Where("task_id = ?", taskID).First(&model).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("analysis result not found for task: %s", taskID)
		}
		return nil, fmt.Errorf("failed to get analysis result by task ID: %w", err)
	}

	return model.toEntity()
}

// GetByAlertID 根据告警ID获取结果
func (r *AnalysisResultRepositoryImpl) GetByAlertID(ctx context.Context, alertID string) ([]*analysis.AnalysisResult, error) {
	var models []analysisResultModel
	if err := r.db.WithContext(ctx).Where("alert_id = ?", alertID).Order("created_at DESC").Find(&models).Error; err != nil {
		return nil, fmt.Errorf("failed to get analysis results by alert ID: %w", err)
	}

	results := make([]*analysis.AnalysisResult, 0, len(models))
	for _, model := range models {
		result, err := model.toEntity()
		if err != nil {
			return nil, fmt.Errorf("failed to convert model to entity: %w", err)
		}
		results = append(results, result)
	}

	return results, nil
}

// Update 更新结果
func (r *AnalysisResultRepositoryImpl) Update(ctx context.Context, result *analysis.AnalysisResult) error {
	model := &analysisResultModel{}
	if err := model.fromEntity(result); err != nil {
		return fmt.Errorf("failed to convert entity to model: %w", err)
	}

	dbResult := r.db.WithContext(ctx).Model(&analysisResultModel{}).Where("id = ?", model.ID).Updates(model)
	if dbResult.Error != nil {
		return fmt.Errorf("failed to update analysis result: %w", dbResult.Error)
	}

	if dbResult.RowsAffected == 0 {
		return fmt.Errorf("analysis result not found: %s", model.ID)
	}

	return nil
}

// Delete 删除结果
func (r *AnalysisResultRepositoryImpl) Delete(ctx context.Context, resultID string) error {
	result := r.db.WithContext(ctx).Where("id = ?", resultID).Delete(&analysisResultModel{})
	if result.Error != nil {
		return fmt.Errorf("failed to delete analysis result: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("analysis result not found: %s", resultID)
	}

	return nil
}

// List 获取结果列表
func (r *AnalysisResultRepositoryImpl) List(ctx context.Context, filter analysis.AnalysisFilter) ([]*analysis.AnalysisResult, error) {
	query := r.db.WithContext(ctx).Model(&analysisResultModel{})

	// 应用过滤条件
	if len(filter.TaskIDs) > 0 {
		query = query.Where("task_id IN ?", filter.TaskIDs)
	}

	if len(filter.AlertIDs) > 0 {
		query = query.Where("alert_id IN ?", filter.AlertIDs)
	}

	if len(filter.Types) > 0 {
		types := make([]string, len(filter.Types))
		for i, t := range filter.Types {
			types[i] = string(t)
		}
		query = query.Where("type IN ?", types)
	}

	if len(filter.Statuses) > 0 {
		statuses := make([]string, len(filter.Statuses))
		for i, s := range filter.Statuses {
			statuses[i] = string(s)
		}
		query = query.Where("status IN ?", statuses)
	}

	if !filter.StartTime.IsZero() {
		query = query.Where("created_at >= ?", filter.StartTime)
	}

	if !filter.EndTime.IsZero() {
		query = query.Where("created_at <= ?", filter.EndTime)
	}

	// 排序
	query = query.Order("created_at DESC")

	// 分页
	if filter.Limit > 0 {
		query = query.Limit(filter.Limit)
	}

	if filter.Offset > 0 {
		query = query.Offset(filter.Offset)
	}

	var models []analysisResultModel
	if err := query.Find(&models).Error; err != nil {
		return nil, fmt.Errorf("failed to list analysis results: %w", err)
	}

	results := make([]*analysis.AnalysisResult, 0, len(models))
	for _, model := range models {
		result, err := model.toEntity()
		if err != nil {
			return nil, fmt.Errorf("failed to convert model to entity: %w", err)
		}
		results = append(results, result)
	}

	return results, nil
}

// Count 统计结果数量
func (r *AnalysisResultRepositoryImpl) Count(ctx context.Context, filter analysis.AnalysisFilter) (int64, error) {
	query := r.db.WithContext(ctx).Model(&analysisResultModel{})

	// 应用过滤条件
	if len(filter.TaskIDs) > 0 {
		query = query.Where("task_id IN ?", filter.TaskIDs)
	}

	if len(filter.AlertIDs) > 0 {
		query = query.Where("alert_id IN ?", filter.AlertIDs)
	}

	if len(filter.Types) > 0 {
		types := make([]string, len(filter.Types))
		for i, t := range filter.Types {
			types[i] = string(t)
		}
		query = query.Where("type IN ?", types)
	}

	if len(filter.Statuses) > 0 {
		statuses := make([]string, len(filter.Statuses))
		for i, s := range filter.Statuses {
			statuses[i] = string(s)
		}
		query = query.Where("status IN ?", statuses)
	}

	if !filter.StartTime.IsZero() {
		query = query.Where("created_at >= ?", filter.StartTime)
	}

	if !filter.EndTime.IsZero() {
		query = query.Where("created_at <= ?", filter.EndTime)
	}

	var count int64
	if err := query.Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to count analysis results: %w", err)
	}

	return count, nil
}

// GetLatestByAlertID 获取告警的最新分析结果
func (r *AnalysisResultRepositoryImpl) GetLatestByAlertID(ctx context.Context, alertID string, analysisType analysis.AnalysisType) (*analysis.AnalysisResult, error) {
	var model analysisResultModel
	query := r.db.WithContext(ctx).Where("alert_id = ?", alertID)
	
	if analysisType != "" {
		query = query.Where("type = ?", string(analysisType))
	}
	
	if err := query.Order("created_at DESC").First(&model).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("no analysis result found for alert: %s", alertID)
		}
		return nil, fmt.Errorf("failed to get latest analysis result: %w", err)
	}

	return model.toEntity()
}