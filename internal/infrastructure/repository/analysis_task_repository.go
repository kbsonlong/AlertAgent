package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"gorm.io/gorm"

	"alert_agent/internal/domain/analysis"
)

// NewAnalysisTaskRepository 创建分析任务存储实例
func NewAnalysisTaskRepository(db *gorm.DB) analysis.AnalysisTaskRepository {
	return &AnalysisTaskRepositoryImpl{
		db: db,
	}
}

// AnalysisTaskRepositoryImpl 分析任务存储实现
type AnalysisTaskRepositoryImpl struct {
	db *gorm.DB
}

// analysisTaskModel 分析任务数据库模型
type analysisTaskModel struct {
	ID             string    `gorm:"primaryKey;type:varchar(255)" json:"id"`
	AlertID        uint      `gorm:"index;not null" json:"alert_id"`
	Type           string    `gorm:"type:varchar(50);not null" json:"type"`
	Status         string    `gorm:"type:varchar(20);not null;index" json:"status"`
	Priority       int       `gorm:"not null;default:5" json:"priority"`
	RetryCount     int       `gorm:"not null;default:0" json:"retry_count"`
	MaxRetries     int       `gorm:"not null;default:3" json:"max_retries"`
	TimeoutSeconds int64     `gorm:"not null;default:300" json:"timeout_seconds"`
	CreatedAt      time.Time `gorm:"not null" json:"created_at"`
	UpdatedAt      time.Time `gorm:"not null" json:"updated_at"`
	StartedAt      *time.Time `json:"started_at"`
	CompletedAt    *time.Time `json:"completed_at"`
	MetadataJSON   string    `gorm:"type:text" json:"metadata_json"`
}

// TableName 指定表名
func (analysisTaskModel) TableName() string {
	return "analysis_tasks"
}

// toEntity 转换为领域实体
func (m *analysisTaskModel) toEntity() (*analysis.AnalysisTask, error) {
	var metadata map[string]interface{}
	if m.MetadataJSON != "" {
		if err := json.Unmarshal([]byte(m.MetadataJSON), &metadata); err != nil {
			return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
		}
	}

	return &analysis.AnalysisTask{
		ID:          m.ID,
		AlertID:     strconv.FormatUint(uint64(m.AlertID), 10),
		Type:        analysis.AnalysisType(m.Type),
		Status:      analysis.AnalysisStatus(m.Status),
		Priority:    m.Priority,
		RetryCount:  m.RetryCount,
		MaxRetries:  m.MaxRetries,
		Timeout:     time.Duration(m.TimeoutSeconds) * time.Second,
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
		StartedAt:   m.StartedAt,
		CompletedAt: m.CompletedAt,
		Metadata:    metadata,
	}, nil
}

// fromEntity 从领域实体转换
func (m *analysisTaskModel) fromEntity(task *analysis.AnalysisTask) error {
	alertID, err := strconv.ParseUint(task.AlertID, 10, 32)
	if err != nil {
		return fmt.Errorf("invalid alert ID: %w", err)
	}

	var metadataJSON string
	if task.Metadata != nil {
		data, err := json.Marshal(task.Metadata)
		if err != nil {
			return fmt.Errorf("failed to marshal metadata: %w", err)
		}
		metadataJSON = string(data)
	}

	m.ID = task.ID
	m.AlertID = uint(alertID)
	m.Type = string(task.Type)
	m.Status = string(task.Status)
	m.Priority = task.Priority
	m.RetryCount = task.RetryCount
	m.MaxRetries = task.MaxRetries
	m.TimeoutSeconds = int64(task.Timeout.Seconds())
	m.CreatedAt = task.CreatedAt
	m.UpdatedAt = task.UpdatedAt
	m.StartedAt = task.StartedAt
	m.CompletedAt = task.CompletedAt
	m.MetadataJSON = metadataJSON

	return nil
}

// Create 创建分析任务
func (r *AnalysisTaskRepositoryImpl) Create(ctx context.Context, task *analysis.AnalysisTask) error {
	var model analysisTaskModel
	if err := model.fromEntity(task); err != nil {
		return err
	}

	if err := r.db.WithContext(ctx).Create(&model).Error; err != nil {
		return fmt.Errorf("failed to create analysis task: %w", err)
	}

	return nil
}

// GetByID 根据ID获取分析任务
func (r *AnalysisTaskRepositoryImpl) GetByID(ctx context.Context, taskID string) (*analysis.AnalysisTask, error) {
	var model analysisTaskModel
	if err := r.db.WithContext(ctx).Where("id = ?", taskID).First(&model).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("analysis task not found: %s", taskID)
		}
		return nil, fmt.Errorf("failed to get analysis task: %w", err)
	}

	return model.toEntity()
}

// Update 更新分析任务
func (r *AnalysisTaskRepositoryImpl) Update(ctx context.Context, task *analysis.AnalysisTask) error {
	var model analysisTaskModel
	if err := model.fromEntity(task); err != nil {
		return err
	}

	result := r.db.WithContext(ctx).Where("id = ?", task.ID).Updates(&model)
	if result.Error != nil {
		return fmt.Errorf("failed to update analysis task: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("analysis task not found: %s", task.ID)
	}

	return nil
}

// Delete 删除分析任务
func (r *AnalysisTaskRepositoryImpl) Delete(ctx context.Context, taskID string) error {
	result := r.db.WithContext(ctx).Where("id = ?", taskID).Delete(&analysisTaskModel{})
	if result.Error != nil {
		return fmt.Errorf("failed to delete analysis task: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("analysis task not found: %s", taskID)
	}

	return nil
}

// List 列出分析任务
func (r *AnalysisTaskRepositoryImpl) List(ctx context.Context, filter analysis.AnalysisFilter) ([]*analysis.AnalysisTask, error) {
	query := r.db.WithContext(ctx).Model(&analysisTaskModel{})

	// 应用过滤条件
	if len(filter.TaskIDs) > 0 {
		query = query.Where("id IN ?", filter.TaskIDs)
	}

	if len(filter.AlertIDs) > 0 {
		// 将AlertID字符串转换为uint进行查询
		alertIDs := make([]uint, 0, len(filter.AlertIDs))
		for _, alertID := range filter.AlertIDs {
			if id, err := strconv.ParseUint(alertID, 10, 32); err == nil {
				alertIDs = append(alertIDs, uint(id))
			}
		}
		if len(alertIDs) > 0 {
			query = query.Where("alert_id IN ?", alertIDs)
		}
	}

	if len(filter.Statuses) > 0 {
		statuses := make([]string, len(filter.Statuses))
		for i, status := range filter.Statuses {
			statuses[i] = string(status)
		}
		query = query.Where("status IN ?", statuses)
	}

	if len(filter.Types) > 0 {
		types := make([]string, len(filter.Types))
		for i, typ := range filter.Types {
			types[i] = string(typ)
		}
		query = query.Where("type IN ?", types)
	}

	if filter.StartTime != nil {
		query = query.Where("created_at >= ?", *filter.StartTime)
	}

	if filter.EndTime != nil {
		query = query.Where("created_at <= ?", *filter.EndTime)
	}

	if filter.Limit > 0 {
		query = query.Limit(filter.Limit)
	}

	if filter.Offset > 0 {
		query = query.Offset(filter.Offset)
	}

	// 默认按创建时间倒序排列
	query = query.Order("created_at DESC")

	var models []analysisTaskModel
	if err := query.Find(&models).Error; err != nil {
		return nil, fmt.Errorf("failed to list analysis tasks: %w", err)
	}

	tasks := make([]*analysis.AnalysisTask, len(models))
	for i, model := range models {
		task, err := model.toEntity()
		if err != nil {
			return nil, err
		}
		tasks[i] = task
	}

	return tasks, nil
}

// Count 统计分析任务数量
func (r *AnalysisTaskRepositoryImpl) Count(ctx context.Context, filter analysis.AnalysisFilter) (int64, error) {
	query := r.db.WithContext(ctx).Model(&analysisTaskModel{})

	// 应用过滤条件
	if len(filter.TaskIDs) > 0 {
		query = query.Where("id IN ?", filter.TaskIDs)
	}

	if len(filter.AlertIDs) > 0 {
		// 将AlertID字符串转换为uint进行查询
		alertIDs := make([]uint, 0, len(filter.AlertIDs))
		for _, alertID := range filter.AlertIDs {
			if id, err := strconv.ParseUint(alertID, 10, 32); err == nil {
				alertIDs = append(alertIDs, uint(id))
			}
		}
		if len(alertIDs) > 0 {
			query = query.Where("alert_id IN ?", alertIDs)
		}
	}

	if len(filter.Statuses) > 0 {
		statuses := make([]string, len(filter.Statuses))
		for i, status := range filter.Statuses {
			statuses[i] = string(status)
		}
		query = query.Where("status IN ?", statuses)
	}

	if len(filter.Types) > 0 {
		types := make([]string, len(filter.Types))
		for i, typ := range filter.Types {
			types[i] = string(typ)
		}
		query = query.Where("type IN ?", types)
	}

	if filter.StartTime != nil {
		query = query.Where("created_at >= ?", *filter.StartTime)
	}

	if filter.EndTime != nil {
		query = query.Where("created_at <= ?", *filter.EndTime)
	}

	var count int64
	if err := query.Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to count analysis tasks: %w", err)
	}

	return count, nil
}

// GetByAlertID 根据AlertID获取分析任务
func (r *AnalysisTaskRepositoryImpl) GetByAlertID(ctx context.Context, alertID string) ([]*analysis.AnalysisTask, error) {
	// 将AlertID字符串转换为uint进行查询
	alertIDUint := uint(0)
	if alertID != "" {
		if id, err := strconv.ParseUint(alertID, 10, 32); err == nil {
			alertIDUint = uint(id)
		}
	}

	var models []analysisTaskModel
	if err := r.db.WithContext(ctx).Where("alert_id = ?", alertIDUint).Find(&models).Error; err != nil {
		return nil, fmt.Errorf("failed to get analysis tasks by alert ID: %w", err)
	}

	tasks := make([]*analysis.AnalysisTask, len(models))
	for i, model := range models {
		task, err := model.toEntity()
		if err != nil {
			return nil, err
		}
		tasks[i] = task
	}

	return tasks, nil
}

// GetByStatus 根据状态获取分析任务
func (r *AnalysisTaskRepositoryImpl) GetByStatus(ctx context.Context, status analysis.AnalysisStatus) ([]*analysis.AnalysisTask, error) {
	var models []analysisTaskModel
	if err := r.db.WithContext(ctx).Where("status = ?", string(status)).Find(&models).Error; err != nil {
		return nil, fmt.Errorf("failed to get analysis tasks by status: %w", err)
	}

	tasks := make([]*analysis.AnalysisTask, len(models))
	for i, model := range models {
		task, err := model.toEntity()
		if err != nil {
			return nil, err
		}
		tasks[i] = task
	}

	return tasks, nil
}

// UpdateStatus 更新任务状态
func (r *AnalysisTaskRepositoryImpl) UpdateStatus(ctx context.Context, taskID string, status analysis.AnalysisStatus) error {
	updates := map[string]interface{}{
		"status":     string(status),
		"updated_at": time.Now(),
	}

	// 如果状态是开始处理，设置开始时间
	if status == analysis.AnalysisStatusProcessing {
		updates["started_at"] = time.Now()
	}

	// 如果状态是完成或失败，设置完成时间
	if status == analysis.AnalysisStatusCompleted || status == analysis.AnalysisStatusFailed {
		updates["completed_at"] = time.Now()
	}

	result := r.db.WithContext(ctx).Model(&analysisTaskModel{}).Where("id = ?", taskID).Updates(updates)
	if result.Error != nil {
		return fmt.Errorf("failed to update task status: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("task not found: %s", taskID)
	}

	return nil
}

// GetExpiredTasks 获取超时任务
func (r *AnalysisTaskRepositoryImpl) GetExpiredTasks(ctx context.Context) ([]*analysis.AnalysisTask, error) {
	now := time.Now()
	var models []analysisTaskModel

	// 查找处理中但已超时的任务
	query := r.db.WithContext(ctx).Where(
		"status = ? AND started_at IS NOT NULL AND started_at + INTERVAL timeout_seconds SECOND < ?",
		string(analysis.AnalysisStatusProcessing),
		now,
	)

	if err := query.Find(&models).Error; err != nil {
		return nil, fmt.Errorf("failed to get expired tasks: %w", err)
	}

	tasks := make([]*analysis.AnalysisTask, len(models))
	for i, model := range models {
		task, err := model.toEntity()
		if err != nil {
			return nil, err
		}
		tasks[i] = task
	}

	return tasks, nil
}

// GetStatistics 获取分析任务统计信息
func (r *AnalysisTaskRepositoryImpl) GetStatistics(ctx context.Context, filter analysis.AnalysisFilter) (*analysis.AnalysisStatistics, error) {
	var stats analysis.AnalysisStatistics

	// 获取总任务数
	total, err := r.Count(ctx, filter)
	if err != nil {
		return nil, err
	}
	stats.TotalTasks = total

	// 按状态统计
	stats.StatusDistribution = make(map[analysis.AnalysisStatus]int64)
	for _, status := range []analysis.AnalysisStatus{
		analysis.AnalysisStatusPending,
		analysis.AnalysisStatusProcessing,
		analysis.AnalysisStatusCompleted,
		analysis.AnalysisStatusFailed,
		analysis.AnalysisStatusCancelled,
	} {
		statusFilter := analysis.AnalysisFilter{
			Statuses: []analysis.AnalysisStatus{status},
			TaskIDs:  filter.TaskIDs,
			AlertIDs: filter.AlertIDs,
			StartTime: filter.StartTime,
			EndTime:   filter.EndTime,
		}

		count, err := r.Count(ctx, statusFilter)
		if err != nil {
			return nil, err
		}
		stats.StatusDistribution[status] = count
		
		// 设置具体状态计数
		switch status {
		case analysis.AnalysisStatusPending:
			stats.PendingTasks = count
		case analysis.AnalysisStatusProcessing:
			stats.ProcessingTasks = count
		case analysis.AnalysisStatusCompleted:
			stats.CompletedTasks = count
		case analysis.AnalysisStatusFailed:
			stats.FailedTasks = count
		}
	}

	// 按类型统计
	stats.TypeDistribution = make(map[analysis.AnalysisType]int64)
	for _, typ := range []analysis.AnalysisType{
		analysis.AnalysisTypeRootCause,
		analysis.AnalysisTypeImpactAssess,
		analysis.AnalysisTypeSolution,
		analysis.AnalysisTypeClassification,
		analysis.AnalysisTypePriority,
	} {
		typeFilter := analysis.AnalysisFilter{
			Types:     []analysis.AnalysisType{typ},
			TaskIDs:   filter.TaskIDs,
			AlertIDs:  filter.AlertIDs,
			StartTime: filter.StartTime,
			EndTime:   filter.EndTime,
		}

		count, err := r.Count(ctx, typeFilter)
		if err != nil {
			return nil, err
		}
		stats.TypeDistribution[typ] = count
	}

	// 平均处理时间（仅统计已完成的任务）
	var avgDuration time.Duration
	var avgQuery = `
		SELECT AVG(EXTRACT(EPOCH FROM (completed_at - started_at))) as avg_seconds
		FROM analysis_tasks 
		WHERE status = ? AND started_at IS NOT NULL AND completed_at IS NOT NULL
	`
	var avgSeconds float64
	if err := r.db.WithContext(ctx).Raw(avgQuery, string(analysis.AnalysisStatusCompleted)).Scan(&avgSeconds).Error; err != nil {
		return nil, fmt.Errorf("failed to calculate average processing time: %w", err)
	}
	avgDuration = time.Duration(avgSeconds) * time.Second
	stats.AverageTime = avgDuration

	// 计算成功率
	if stats.TotalTasks > 0 {
		stats.SuccessRate = float64(stats.CompletedTasks) / float64(stats.TotalTasks)
	}

	stats.LastUpdated = time.Now()
	return &stats, nil
}