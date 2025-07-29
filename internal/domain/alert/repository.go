package alert

import (
	"context"
	"time"

	"alert_agent/internal/model"
)

// AlertFilter 告警过滤条件
type AlertFilter struct {
	Status     []string  `json:"status,omitempty"`
	Severity   []string  `json:"severity,omitempty"`
	Source     []string  `json:"source,omitempty"`
	RuleID     *uint     `json:"rule_id,omitempty"`
	CreatedAt  *TimeRange `json:"created_at,omitempty"`
	UpdatedAt  *TimeRange `json:"updated_at,omitempty"`
	Keywords   string    `json:"keywords,omitempty"`
	Limit      int       `json:"limit,omitempty"`
	Offset     int       `json:"offset,omitempty"`
}

// TimeRange 时间范围
type TimeRange struct {
	Start *time.Time `json:"start,omitempty"`
	End   *time.Time `json:"end,omitempty"`
}

// AlertStatistics 告警统计信息
type AlertStatistics struct {
	Total      int64            `json:"total"`
	ByStatus   map[string]int64 `json:"by_status"`
	BySeverity map[string]int64 `json:"by_severity"`
	BySource   map[string]int64 `json:"by_source"`
	Trend      []TrendPoint     `json:"trend"`
}

// TrendPoint 趋势数据点
type TrendPoint struct {
	Timestamp time.Time `json:"timestamp"`
	Count     int64     `json:"count"`
}

// AlertRepository 告警数据仓储接口
type AlertRepository interface {
	// Create 创建告警
	Create(ctx context.Context, alert *model.Alert) error
	
	// Update 更新告警
	Update(ctx context.Context, alert *model.Alert) error
	
	// UpdateByID 根据ID更新告警字段
	UpdateByID(ctx context.Context, id uint, updates map[string]interface{}) error
	
	// GetByID 根据ID获取告警
	GetByID(ctx context.Context, id uint) (*model.Alert, error)
	
	// List 获取告警列表
	List(ctx context.Context, filter AlertFilter) ([]*model.Alert, error)
	
	// Count 统计告警数量
	Count(ctx context.Context, filter AlertFilter) (int64, error)
	
	// Delete 删除告警
	Delete(ctx context.Context, id uint) error
	
	// BatchUpdate 批量更新告警
	BatchUpdate(ctx context.Context, ids []uint, updates map[string]interface{}) error
	
	// GetSimilarAlerts 获取相似告警
	GetSimilarAlerts(ctx context.Context, alert *model.Alert, limit int) ([]*model.Alert, error)
	
	// GetStatistics 获取告警统计信息
	GetStatistics(ctx context.Context, filter AlertFilter) (*AlertStatistics, error)
	
	// GetRecentAlerts 获取最近的告警
	GetRecentAlerts(ctx context.Context, duration time.Duration, limit int) ([]*model.Alert, error)
	
	// UpdateAnalysisResult 更新告警分析结果
	UpdateAnalysisResult(ctx context.Context, alertID uint, analysis string) error
	
	// GetAlertsForAnalysis 获取需要分析的告警
	GetAlertsForAnalysis(ctx context.Context, limit int) ([]*model.Alert, error)
	
	// MarkAsAnalyzed 标记告警为已分析
	MarkAsAnalyzed(ctx context.Context, alertID uint) error
}