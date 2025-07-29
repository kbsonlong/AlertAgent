package n8n

import (
	"context"
	"encoding/json"
	"time"

	"alert_agent/internal/domain/analysis"
	"gorm.io/gorm"
)

// N8NWorkflowExecutionModel GORM 模型
type N8NWorkflowExecutionModel struct {
	ID          string    `gorm:"primaryKey;type:varchar(255)" json:"id"`
	WorkflowID  string    `gorm:"type:varchar(255);index" json:"workflow_id"`
	Status      string    `gorm:"type:varchar(50);index" json:"status"`
	StartedAt   time.Time `gorm:"index" json:"started_at"`
	FinishedAt  *time.Time `json:"finished_at"`
	InputData   string    `gorm:"type:text" json:"input_data"`
	OutputData  string    `gorm:"type:text" json:"output_data"`
	ErrorData   *string   `gorm:"type:text" json:"error_data"`
	Metadata    string    `gorm:"type:text" json:"metadata"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// TableName 指定表名
func (N8NWorkflowExecutionModel) TableName() string {
	return "n8n_workflow_executions"
}

// ToEntity 转换为领域实体
func (m *N8NWorkflowExecutionModel) ToEntity() (*analysis.N8NWorkflowExecution, error) {
	execution := &analysis.N8NWorkflowExecution{
		ID:         m.ID,
		WorkflowID: m.WorkflowID,
		Status:     analysis.N8NWorkflowStatus(m.Status),
		StartedAt:  m.StartedAt,
		FinishedAt: m.FinishedAt,
		ErrorData:  m.ErrorData,
	}

	// 解析 JSON 字段
	if m.InputData != "" {
		if err := json.Unmarshal([]byte(m.InputData), &execution.InputData); err != nil {
			return nil, err
		}
	}

	if m.OutputData != "" {
		if err := json.Unmarshal([]byte(m.OutputData), &execution.OutputData); err != nil {
			return nil, err
		}
	}

	if m.Metadata != "" {
		if err := json.Unmarshal([]byte(m.Metadata), &execution.Metadata); err != nil {
			return nil, err
		}
	}

	return execution, nil
}

// FromEntity 从领域实体创建模型
func (m *N8NWorkflowExecutionModel) FromEntity(execution *analysis.N8NWorkflowExecution) error {
	m.ID = execution.ID
	m.WorkflowID = execution.WorkflowID
	m.Status = string(execution.Status)
	m.StartedAt = execution.StartedAt
	m.FinishedAt = execution.FinishedAt
	m.ErrorData = execution.ErrorData

	// 序列化 JSON 字段
	if execution.InputData != nil {
		inputData, err := json.Marshal(execution.InputData)
		if err != nil {
			return err
		}
		m.InputData = string(inputData)
	}

	if execution.OutputData != nil {
		outputData, err := json.Marshal(execution.OutputData)
		if err != nil {
			return err
		}
		m.OutputData = string(outputData)
	}

	if execution.Metadata != nil {
		metadata, err := json.Marshal(execution.Metadata)
		if err != nil {
			return err
		}
		m.Metadata = string(metadata)
	}

	return nil
}

// GORMExecutionRepository GORM 实现的工作流执行记录存储
type GORMExecutionRepository struct {
	db *gorm.DB
}

// NewGORMExecutionRepository 创建新的 GORM 执行记录存储
func NewGORMExecutionRepository(db *gorm.DB) *GORMExecutionRepository {
	return &GORMExecutionRepository{db: db}
}

// Create 创建工作流执行记录
func (r *GORMExecutionRepository) Create(ctx context.Context, execution *analysis.N8NWorkflowExecution) error {
	var model N8NWorkflowExecutionModel
	if err := model.FromEntity(execution); err != nil {
		return err
	}

	return r.db.WithContext(ctx).Create(&model).Error
}

// GetByID 根据ID获取工作流执行记录
func (r *GORMExecutionRepository) GetByID(ctx context.Context, id string) (*analysis.N8NWorkflowExecution, error) {
	var model N8NWorkflowExecutionModel
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&model).Error; err != nil {
		return nil, err
	}

	return model.ToEntity()
}

// Update 更新工作流执行记录
func (r *GORMExecutionRepository) Update(ctx context.Context, execution *analysis.N8NWorkflowExecution) error {
	var model N8NWorkflowExecutionModel
	if err := model.FromEntity(execution); err != nil {
		return err
	}

	return r.db.WithContext(ctx).Where("id = ?", execution.ID).Updates(&model).Error
}

// Delete 删除工作流执行记录
func (r *GORMExecutionRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Where("id = ?", id).Delete(&N8NWorkflowExecutionModel{}).Error
}

// ListByWorkflowID 根据工作流ID列出执行记录
func (r *GORMExecutionRepository) ListByWorkflowID(ctx context.Context, workflowID string, limit, offset int) ([]*analysis.N8NWorkflowExecution, error) {
	var models []N8NWorkflowExecutionModel
	if err := r.db.WithContext(ctx).
		Where("workflow_id = ?", workflowID).
		Order("started_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&models).Error; err != nil {
		return nil, err
	}

	executions := make([]*analysis.N8NWorkflowExecution, len(models))
	for i, model := range models {
		execution, err := model.ToEntity()
		if err != nil {
			return nil, err
		}
		executions[i] = execution
	}

	return executions, nil
}

// ListByStatus 根据状态列出执行记录
func (r *GORMExecutionRepository) ListByStatus(ctx context.Context, status analysis.N8NWorkflowStatus, limit, offset int) ([]*analysis.N8NWorkflowExecution, error) {
	var models []N8NWorkflowExecutionModel
	if err := r.db.WithContext(ctx).
		Where("status = ?", string(status)).
		Order("started_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&models).Error; err != nil {
		return nil, err
	}

	executions := make([]*analysis.N8NWorkflowExecution, len(models))
	for i, model := range models {
		execution, err := model.ToEntity()
		if err != nil {
			return nil, err
		}
		executions[i] = execution
	}

	return executions, nil
}

// ListByDateRange 根据日期范围列出执行记录
func (r *GORMExecutionRepository) ListByDateRange(ctx context.Context, startTime, endTime time.Time, limit, offset int) ([]*analysis.N8NWorkflowExecution, error) {
	var models []N8NWorkflowExecutionModel
	if err := r.db.WithContext(ctx).
		Where("started_at BETWEEN ? AND ?", startTime, endTime).
		Order("started_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&models).Error; err != nil {
		return nil, err
	}

	executions := make([]*analysis.N8NWorkflowExecution, len(models))
	for i, model := range models {
		execution, err := model.ToEntity()
		if err != nil {
			return nil, err
		}
		executions[i] = execution
	}

	return executions, nil
}

// GetStatistics 获取执行统计信息
func (r *GORMExecutionRepository) GetStatistics(ctx context.Context, workflowID string, startTime, endTime time.Time) (map[string]interface{}, error) {
	type StatusCount struct {
		Status string `json:"status"`
		Count  int64  `json:"count"`
	}

	var statusCounts []StatusCount
	query := r.db.WithContext(ctx).
		Model(&N8NWorkflowExecutionModel{}).
		Select("status, COUNT(*) as count").
		Where("started_at BETWEEN ? AND ?", startTime, endTime).
		Group("status")

	if workflowID != "" {
		query = query.Where("workflow_id = ?", workflowID)
	}

	if err := query.Find(&statusCounts).Error; err != nil {
		return nil, err
	}

	// 计算总执行次数
	var totalCount int64
	totalQuery := r.db.WithContext(ctx).
		Model(&N8NWorkflowExecutionModel{}).
		Where("started_at BETWEEN ? AND ?", startTime, endTime)

	if workflowID != "" {
		totalQuery = totalQuery.Where("workflow_id = ?", workflowID)
	}

	if err := totalQuery.Count(&totalCount).Error; err != nil {
		return nil, err
	}

	// 计算平均执行时间
	var avgDuration float64
	durationQuery := r.db.WithContext(ctx).
		Model(&N8NWorkflowExecutionModel{}).
		Select("AVG(EXTRACT(EPOCH FROM (finished_at - started_at))) as avg_duration").
		Where("started_at BETWEEN ? AND ? AND finished_at IS NOT NULL", startTime, endTime)

	if workflowID != "" {
		durationQuery = durationQuery.Where("workflow_id = ?", workflowID)
	}

	if err := durationQuery.Scan(&avgDuration).Error; err != nil {
		return nil, err
	}

	// 构建统计结果
	stats := map[string]interface{}{
		"total_count":    totalCount,
		"avg_duration":   avgDuration,
		"status_counts":  statusCounts,
		"start_time":     startTime,
		"end_time":       endTime,
		"workflow_id":    workflowID,
	}

	return stats, nil
}