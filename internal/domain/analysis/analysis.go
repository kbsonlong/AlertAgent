package analysis

import (
	"context"
	"time"
)

// AnalysisRecord AI分析记录
type AnalysisRecord struct {
	ID              string                 `json:"id"`
	AlertID         string                 `json:"alert_id"`
	AnalysisType    string                 `json:"analysis_type"`
	RequestData     map[string]interface{} `json:"request_data"`
	ResponseData    map[string]interface{} `json:"response_data"`
	AnalysisResult  map[string]interface{} `json:"analysis_result"`
	ConfidenceScore float64                `json:"confidence_score"`
	ProcessingTime  int                    `json:"processing_time"`
	Status          AnalysisStatus         `json:"status"`
	ErrorMessage    string                 `json:"error_message"`
	Provider        string                 `json:"provider"`
	ModelVersion    string                 `json:"model_version"`
	CreatedAt       time.Time              `json:"created_at"`
	UpdatedAt       time.Time              `json:"updated_at"`
}

// AnalysisStatus 分析状态
type AnalysisStatus string

const (
	AnalysisStatusPending    AnalysisStatus = "pending"
	AnalysisStatusInProgress AnalysisStatus = "in_progress"
	AnalysisStatusCompleted  AnalysisStatus = "completed"
	AnalysisStatusFailed     AnalysisStatus = "failed"
	AnalysisStatusTimeout    AnalysisStatus = "timeout"
)

// AnalysisRequest 分析请求
type AnalysisRequest struct {
	AlertID      string                 `json:"alert_id"`
	AnalysisType string                 `json:"analysis_type"`
	AlertData    map[string]interface{} `json:"alert_data"`
	Context      map[string]interface{} `json:"context"`
	Priority     int                    `json:"priority"`
	Timeout      int                    `json:"timeout"` // 秒
}

// AnalysisResult 分析结果
type AnalysisResult struct {
	RootCause       string                 `json:"root_cause"`
	ConfidenceScore float64                `json:"confidence_score"`
	Recommendations []Recommendation       `json:"recommendations"`
	RelatedAlerts   []string               `json:"related_alerts"`
	Metadata        map[string]interface{} `json:"metadata"`
	ProcessingTime  int                    `json:"processing_time"`
}

// Recommendation 建议
type Recommendation struct {
	Action      string                 `json:"action"`
	Description string                 `json:"description"`
	Priority    int                    `json:"priority"`
	Confidence  float64                `json:"confidence"`
	Automated   bool                   `json:"automated"`
	Parameters  map[string]interface{} `json:"parameters"`
}

// AnalysisRepository 分析仓储接口
type AnalysisRepository interface {
	// 基本CRUD操作
	Create(ctx context.Context, record *AnalysisRecord) error
	Update(ctx context.Context, record *AnalysisRecord) error
	GetByID(ctx context.Context, id string) (*AnalysisRecord, error)
	GetByAlertID(ctx context.Context, alertID string) ([]*AnalysisRecord, error)
	List(ctx context.Context, query *AnalysisQuery) ([]*AnalysisRecord, int64, error)

	// 状态操作
	UpdateStatus(ctx context.Context, id string, status AnalysisStatus) error
	GetPendingAnalysis(ctx context.Context, limit int) ([]*AnalysisRecord, error)

	// 统计操作
	GetAnalysisStats(ctx context.Context, startTime, endTime time.Time) (*AnalysisStats, error)
}

// AnalysisQuery 分析查询
type AnalysisQuery struct {
	AlertID      string         `json:"alert_id"`
	AnalysisType string         `json:"analysis_type"`
	Status       AnalysisStatus `json:"status"`
	Provider     string         `json:"provider"`
	StartTime    time.Time      `json:"start_time"`
	EndTime      time.Time      `json:"end_time"`
	Page         int            `json:"page"`
	PageSize     int            `json:"page_size"`
	SortBy       string         `json:"sort_by"`
	SortDesc     bool           `json:"sort_desc"`
}

// AnalysisStats 分析统计
type AnalysisStats struct {
	TotalAnalysis     int64   `json:"total_analysis"`
	CompletedAnalysis int64   `json:"completed_analysis"`
	FailedAnalysis    int64   `json:"failed_analysis"`
	AvgProcessingTime float64 `json:"avg_processing_time"`
	AvgConfidence     float64 `json:"avg_confidence"`
	SuccessRate       float64 `json:"success_rate"`
}