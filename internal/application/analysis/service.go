package analysis

import (
	"context"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"alert_agent/internal/domain/analysis"
	"alert_agent/internal/shared/errors"
)

// AnalysisService 分析服务接口
type AnalysisService interface {
	// 分析管理
	CreateAnalysis(ctx context.Context, request *analysis.AnalysisRequest) (*analysis.AnalysisRecord, error)
	GetAnalysis(ctx context.Context, id string) (*analysis.AnalysisRecord, error)
	GetAnalysisByAlertID(ctx context.Context, alertID string) ([]*analysis.AnalysisRecord, error)
	ListAnalysis(ctx context.Context, query *analysis.AnalysisQuery) ([]*analysis.AnalysisRecord, int64, error)
	UpdateAnalysisStatus(ctx context.Context, id string, status analysis.AnalysisStatus) error

	// 统计信息
	GetAnalysisStats(ctx context.Context, startTime, endTime time.Time) (*analysis.AnalysisStats, error)
	GetPendingAnalysis(ctx context.Context, limit int) ([]*analysis.AnalysisRecord, error)
}

// analysisServiceImpl 分析服务实现
type analysisServiceImpl struct {
	repo   analysis.AnalysisRepository
	logger *zap.Logger
}

// NewAnalysisService 创建分析服务
func NewAnalysisService(repo analysis.AnalysisRepository, logger *zap.Logger) AnalysisService {
	return &analysisServiceImpl{
		repo:   repo,
		logger: logger,
	}
}

// CreateAnalysis 创建分析
func (s *analysisServiceImpl) CreateAnalysis(ctx context.Context, request *analysis.AnalysisRequest) (*analysis.AnalysisRecord, error) {
	// 验证请求
	if request.AlertID == "" {
		return nil, errors.NewValidationError("MISSING_ALERT_ID", "Alert ID is required")
	}
	if request.AnalysisType == "" {
		request.AnalysisType = "root_cause_analysis"
	}

	// 创建分析记录
	record := &analysis.AnalysisRecord{
		ID:           uuid.New().String(),
		AlertID:      request.AlertID,
		AnalysisType: request.AnalysisType,
		RequestData: map[string]interface{}{
			"alert_data": request.AlertData,
			"context":    request.Context,
			"priority":   request.Priority,
			"timeout":    request.Timeout,
		},
		Status:    analysis.AnalysisStatusPending,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// 保存到数据库
	if err := s.repo.Create(ctx, record); err != nil {
		s.logger.Error("Failed to create analysis record", zap.Error(err), zap.String("alert_id", request.AlertID))
		return nil, errors.NewInternalError("Failed to create analysis record", err)
	}

	s.logger.Info("Analysis record created successfully", 
		zap.String("id", record.ID), 
		zap.String("alert_id", request.AlertID),
		zap.String("analysis_type", request.AnalysisType))
	
	return record, nil
}

// GetAnalysis 获取分析
func (s *analysisServiceImpl) GetAnalysis(ctx context.Context, id string) (*analysis.AnalysisRecord, error) {
	record, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, errors.NewNotFoundError("Analysis record")
	}
	return record, nil
}

// GetAnalysisByAlertID 根据告警ID获取分析
func (s *analysisServiceImpl) GetAnalysisByAlertID(ctx context.Context, alertID string) ([]*analysis.AnalysisRecord, error) {
	records, err := s.repo.GetByAlertID(ctx, alertID)
	if err != nil {
		s.logger.Error("Failed to get analysis by alert ID", zap.Error(err), zap.String("alert_id", alertID))
		return nil, errors.NewInternalError("Failed to get analysis records", err)
	}
	return records, nil
}

// ListAnalysis 列出分析
func (s *analysisServiceImpl) ListAnalysis(ctx context.Context, query *analysis.AnalysisQuery) ([]*analysis.AnalysisRecord, int64, error) {
	records, total, err := s.repo.List(ctx, query)
	if err != nil {
		s.logger.Error("Failed to list analysis records", zap.Error(err))
		return nil, 0, errors.NewInternalError("Failed to list analysis records", err)
	}
	return records, total, nil
}

// UpdateAnalysisStatus 更新分析状态
func (s *analysisServiceImpl) UpdateAnalysisStatus(ctx context.Context, id string, status analysis.AnalysisStatus) error {
	// 检查记录是否存在
	_, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return errors.NewNotFoundError("Analysis record")
	}

	// 更新状态
	if err := s.repo.UpdateStatus(ctx, id, status); err != nil {
		s.logger.Error("Failed to update analysis status", zap.Error(err), zap.String("id", id))
		return errors.NewInternalError("Failed to update analysis status", err)
	}

	s.logger.Info("Analysis status updated successfully", 
		zap.String("id", id), 
		zap.String("status", string(status)))
	
	return nil
}

// GetAnalysisStats 获取分析统计
func (s *analysisServiceImpl) GetAnalysisStats(ctx context.Context, startTime, endTime time.Time) (*analysis.AnalysisStats, error) {
	stats, err := s.repo.GetAnalysisStats(ctx, startTime, endTime)
	if err != nil {
		s.logger.Error("Failed to get analysis stats", zap.Error(err))
		return nil, errors.NewInternalError("Failed to get analysis statistics", err)
	}
	return stats, nil
}

// GetPendingAnalysis 获取待处理的分析
func (s *analysisServiceImpl) GetPendingAnalysis(ctx context.Context, limit int) ([]*analysis.AnalysisRecord, error) {
	records, err := s.repo.GetPendingAnalysis(ctx, limit)
	if err != nil {
		s.logger.Error("Failed to get pending analysis", zap.Error(err))
		return nil, errors.NewInternalError("Failed to get pending analysis", err)
	}
	return records, nil
}