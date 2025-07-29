package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"alert_agent/internal/domain/analysis"
	"gorm.io/gorm"
	"go.uber.org/zap"
)

// DifyAnalysisRepositoryImpl Dify 分析仓储实现
type DifyAnalysisRepositoryImpl struct {
	db     *gorm.DB
	logger *zap.Logger
}

// NewDifyAnalysisRepository 创建新的 Dify 分析仓储
func NewDifyAnalysisRepository(db *gorm.DB, logger *zap.Logger) analysis.DifyAnalysisRepository {
	return &DifyAnalysisRepositoryImpl{
		db:     db,
		logger: logger,
	}
}

// DeleteAnalysisResult 删除分析结果
func (r *DifyAnalysisRepositoryImpl) DeleteAnalysisResult(ctx context.Context, id string) error {
	if err := r.db.WithContext(ctx).Where("id = ?", id).Delete(&DifyAnalysisResultModel{}).Error; err != nil {
		r.logger.Error("Failed to delete analysis result",
			zap.String("task_id", id),
			zap.Error(err),
		)
		return fmt.Errorf("failed to delete analysis result: %w", err)
	}
	
	r.logger.Info("Analysis result deleted", zap.String("task_id", id))
	return nil
}

// GetAnalysisResultsByAlertID 根据告警ID获取分析结果
func (r *DifyAnalysisRepositoryImpl) GetAnalysisResultsByAlertID(ctx context.Context, alertID uint) ([]*analysis.DifyAnalysisResult, error) {
	var models []DifyAnalysisResultModel
	
	if err := r.db.WithContext(ctx).Where("alert_id = ?", alertID).Order("created_at DESC").Find(&models).Error; err != nil {
		return nil, fmt.Errorf("failed to get analysis results by alert ID: %w", err)
	}
	
	// 转换为领域对象
	results := make([]*analysis.DifyAnalysisResult, len(models))
	for i, model := range models {
		result, err := r.modelToResult(&model)
		if err != nil {
			r.logger.Error("Failed to convert model to result",
				zap.String("task_id", model.TaskID),
				zap.Error(err),
			)
			continue
		}
		results[i] = result
	}
	
	return results, nil
}

// UpdateAnalysisResult 更新分析结果
func (r *DifyAnalysisRepositoryImpl) UpdateAnalysisResult(ctx context.Context, result *analysis.DifyAnalysisResult) error {
	// 转换为数据模型
	model := &DifyAnalysisResultModel{
		ID:             result.ID,
		AlertID:        result.AlertID,
		AnalysisType:   result.AnalysisType,
		Confidence:     result.Confidence,
		ConversationID: result.ConversationID,
		MessageID:      result.MessageID,
		RawResponse:    result.RawResponse,
		ProcessingTime: result.ProcessingTime,
		Status:         result.Status,
		Error:          result.Error,
		UpdatedAt:      time.Now(),
	}
	
	// 序列化结果
	if result.Result != nil {
		resultJSON, err := json.Marshal(result.Result)
		if err != nil {
			return fmt.Errorf("failed to marshal result: %w", err)
		}
		model.ResultJSON = string(resultJSON)
	}
	
	// 序列化使用量
	if result.Usage != nil {
		usageJSON, err := json.Marshal(result.Usage)
		if err != nil {
			return fmt.Errorf("failed to marshal usage: %w", err)
		}
		model.UsageJSON = string(usageJSON)
	}
	
	// 更新数据库
	if err := r.db.WithContext(ctx).Where("id = ?", result.ID).Updates(model).Error; err != nil {
		r.logger.Error("Failed to update analysis result",
			zap.String("task_id", result.ID),
			zap.Error(err),
		)
		return fmt.Errorf("failed to update analysis result: %w", err)
	}
	
	r.logger.Info("Analysis result updated", zap.String("task_id", result.ID))
	return nil
}

// DifyAnalysisResultModel 分析结果数据模型
type DifyAnalysisResultModel struct {
	ID             string    `gorm:"primaryKey;size:255" json:"id"`
	AlertID        uint      `gorm:"index;not null" json:"alert_id"`
	AnalysisType   string    `gorm:"size:50;not null" json:"analysis_type"`
	ConversationID string    `gorm:"size:255" json:"conversation_id"`
	MessageID      string    `gorm:"size:255" json:"message_id"`
	ResultJSON     string    `gorm:"type:text" json:"result_json"`
	RawResponse    string    `gorm:"type:text" json:"raw_response"`
	Confidence     float64   `gorm:"type:decimal(5,4)" json:"confidence"`
	ProcessingTime int64     `gorm:"default:0" json:"processing_time"`
	UsageJSON      string    `gorm:"type:text" json:"usage_json"`
	Status         string    `gorm:"size:50;not null" json:"status"`
	Error          string    `gorm:"type:text" json:"error"`
	CreatedAt      time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt      time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

// TableName 指定表名
func (DifyAnalysisResultModel) TableName() string {
	return "dify_analysis_results"
}

// SaveAnalysisResult 保存分析结果
func (r *DifyAnalysisRepositoryImpl) SaveAnalysisResult(ctx context.Context, result *analysis.DifyAnalysisResult) error {
	// 转换为数据模型
	model := &DifyAnalysisResultModel{
		ID:             result.ID,
		AlertID:        result.AlertID,
		AnalysisType:   result.AnalysisType,
		ConversationID: result.ConversationID,
		MessageID:      result.MessageID,
		RawResponse:    result.RawResponse,
		Confidence:     result.Confidence,
		ProcessingTime: result.ProcessingTime,
		Status:         result.Status,
		Error:          result.Error,
		CreatedAt:      result.CreatedAt,
		UpdatedAt:      result.UpdatedAt,
	}
	
	// 序列化结果
	if result.Result != nil {
		resultJSON, err := json.Marshal(result.Result)
		if err != nil {
			return fmt.Errorf("failed to marshal result: %w", err)
		}
		model.ResultJSON = string(resultJSON)
	}
	
	// 序列化使用量
	if result.Usage != nil {
		usageJSON, err := json.Marshal(result.Usage)
		if err != nil {
			return fmt.Errorf("failed to marshal usage: %w", err)
		}
		model.UsageJSON = string(usageJSON)
	}
	
	// 保存到数据库
	if err := r.db.WithContext(ctx).Create(model).Error; err != nil {
		r.logger.Error("Failed to save analysis result",
			zap.String("task_id", result.ID),
			zap.Uint("alert_id", result.AlertID),
			zap.Error(err),
		)
		return fmt.Errorf("failed to save analysis result: %w", err)
	}
	
	r.logger.Info("Analysis result saved",
		zap.String("task_id", result.ID),
		zap.Uint("alert_id", result.AlertID),
		zap.String("analysis_type", result.AnalysisType),
	)
	
	return nil
}

// GetAnalysisResult 获取分析结果
func (r *DifyAnalysisRepositoryImpl) GetAnalysisResult(ctx context.Context, taskID string) (*analysis.DifyAnalysisResult, error) {
	var model DifyAnalysisResultModel
	
	if err := r.db.WithContext(ctx).Where("id = ?", taskID).First(&model).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("analysis result not found for task: %s", taskID)
		}
		return nil, fmt.Errorf("failed to get analysis result: %w", err)
	}
	
	return r.modelToResult(&model)
}

// GetAnalysisHistory 获取分析历史
func (r *DifyAnalysisRepositoryImpl) GetAnalysisHistory(ctx context.Context, filter *analysis.DifyAnalysisFilter) ([]*analysis.DifyAnalysisResult, error) {
	query := r.db.WithContext(ctx).Model(&DifyAnalysisResultModel{})
	
	// 应用过滤条件
	if filter != nil {
		if len(filter.AlertIDs) > 0 {
			query = query.Where("alert_id IN ?", filter.AlertIDs)
		}
		if len(filter.AnalysisTypes) > 0 {
			query = query.Where("analysis_type IN ?", filter.AnalysisTypes)
		}
		if len(filter.Statuses) > 0 {
			query = query.Where("status IN ?", filter.Statuses)
		}
		if filter.StartTime != nil {
			query = query.Where("created_at >= ?", *filter.StartTime)
		}
		if filter.EndTime != nil {
			query = query.Where("created_at <= ?", *filter.EndTime)
		}

		
		// 排序和分页
		sortBy := "created_at"
		if filter.SortBy != "" {
			sortBy = filter.SortBy
		}
		
		sortOrder := "DESC"
		if filter.SortOrder != "" {
			sortOrder = filter.SortOrder
		}
		
		query = query.Order(fmt.Sprintf("%s %s", sortBy, sortOrder))
		
		if filter.Limit > 0 {
			query = query.Limit(filter.Limit)
		} else {
			query = query.Limit(100) // 默认限制
		}
		if filter.Offset > 0 {
			query = query.Offset(filter.Offset)
		}
	}
	
	var models []DifyAnalysisResultModel
	if err := query.Find(&models).Error; err != nil {
		return nil, fmt.Errorf("failed to get analysis history: %w", err)
	}
	
	// 转换为领域对象
	results := make([]*analysis.DifyAnalysisResult, len(models))
	for i, model := range models {
		result, err := r.modelToResult(&model)
		if err != nil {
			r.logger.Error("Failed to convert model to result",
				zap.String("task_id", model.ID),
				zap.Error(err),
			)
			continue
		}
		results[i] = result
	}
	
	return results, nil
}

// GetAnalysisTrends 获取分析趋势
func (r *DifyAnalysisRepositoryImpl) GetAnalysisTrends(ctx context.Context, request *analysis.DifyTrendRequest) (*analysis.DifyTrendResponse, error) {
	if request.TimeRange == nil {
		return nil, fmt.Errorf("time range is required")
	}
	
	// 构建基础查询
	query := r.db.WithContext(ctx).Model(&DifyAnalysisResultModel{}).
		Where("created_at >= ? AND created_at <= ?", request.TimeRange.Start, request.TimeRange.End)
	
	// 应用过滤条件
	if request.Filters != nil {
		for key, value := range request.Filters {
			switch key {
			case "analysis_type":
				query = query.Where("analysis_type = ?", value)
			case "alert_id":
				query = query.Where("alert_id = ?", value)
			}
		}
	}
	
	// 根据聚合方式构建查询
	aggregation := request.Aggregation
	if aggregation == "" {
		aggregation = "daily" // 默认按天聚合
	}
	
	var trends []analysis.TrendData
	var statistics map[string]interface{}
	
	switch aggregation {
	case "hourly":
		trends, statistics = r.getHourlyTrends(ctx, query, request)
	case "daily":
		trends, statistics = r.getDailyTrends(ctx, query, request)
	case "weekly":
		trends, statistics = r.getWeeklyTrends(ctx, query, request)
	default:
		return nil, fmt.Errorf("unsupported aggregation: %s", aggregation)
	}
	
	// 生成洞察和建议
	insights := r.generateInsights(trends, statistics)
	recommendations := r.generateRecommendations(trends, statistics)
	
	return &analysis.DifyTrendResponse{
		Trends:          trends,
		Statistics:      statistics,
		Insights:        insights,
		Recommendations: recommendations,
	}, nil
}

// GetAnalysisMetrics 获取分析指标
func (r *DifyAnalysisRepositoryImpl) GetAnalysisMetrics(ctx context.Context, timeRange *analysis.TimeRange) (*analysis.DifyAnalysisMetrics, error) {
	if timeRange == nil {
		return nil, fmt.Errorf("time range is required")
	}
	
	query := r.db.WithContext(ctx).Model(&DifyAnalysisResultModel{}).
		Where("created_at >= ? AND created_at <= ?", timeRange.Start, timeRange.End)
	
	// 获取基础统计
	var totalAnalyses int64
	var successfulAnalyses int64
	var totalTokenUsage int64
	var totalCost float64
	var avgConfidence float64
	
	query.Count(&totalAnalyses)
	query.Where("result != '' AND result IS NOT NULL").Count(&successfulAnalyses)
	query.Select("COALESCE(SUM(token_usage), 0)").Scan(&totalTokenUsage)
	query.Select("COALESCE(SUM(cost), 0)").Scan(&totalCost)
	query.Select("COALESCE(AVG(confidence), 0)").Scan(&avgConfidence)
	
	failedAnalyses := totalAnalyses - successfulAnalyses
	
	// 获取平均处理时间（这里简化处理，实际应该从任务表获取）
	avgProcessingTime := 120000.0 // 2分钟，单位毫秒
	
	// 按分析类型统计
	byAnalysisType := make(map[string]*analysis.AnalysisTypeMetrics)
	var typeStats []struct {
		AnalysisType string
		Total        int64
		Successful   int64
		AvgConfidence float64
	}
	
	r.db.WithContext(ctx).Model(&DifyAnalysisResultModel{}).
		Where("created_at >= ? AND created_at <= ?", timeRange.Start, timeRange.End).
		Select("analysis_type, COUNT(*) as total, SUM(CASE WHEN result != '' AND result IS NOT NULL THEN 1 ELSE 0 END) as successful, AVG(confidence) as avg_confidence").
		Group("analysis_type").
		Scan(&typeStats)
	
	for _, stat := range typeStats {
		byAnalysisType[stat.AnalysisType] = &analysis.AnalysisTypeMetrics{
			Type:              stat.AnalysisType,
			Total:             stat.Total,
			Successful:        stat.Successful,
			Failed:            stat.Total - stat.Successful,
			AverageTime:       avgProcessingTime,
			AverageConfidence: stat.AvgConfidence,
		}
	}
	
	// 按时间统计（按天）
	var timeStats []struct {
		Date       time.Time
		Count      int64
		Successful int64
	}
	
	r.db.WithContext(ctx).Model(&DifyAnalysisResultModel{}).
		Where("created_at >= ? AND created_at <= ?", timeRange.Start, timeRange.End).
		Select("DATE(created_at) as date, COUNT(*) as count, SUM(CASE WHEN result != '' AND result IS NOT NULL THEN 1 ELSE 0 END) as successful").
		Group("DATE(created_at)").
		Order("date").
		Scan(&timeStats)
	
	byTime := make([]analysis.TimeMetrics, len(timeStats))
	for i, stat := range timeStats {
		byTime[i] = analysis.TimeMetrics{
			Timestamp:   stat.Date,
			Count:       stat.Count,
			Successful:  stat.Successful,
			Failed:      stat.Count - stat.Successful,
			AverageTime: avgProcessingTime,
		}
	}
	
	// 错误统计（简化处理）
	errorStats := map[string]int64{
		"timeout":           failedAnalyses / 3,
		"api_error":         failedAnalyses / 3,
		"context_build_error": failedAnalyses / 3,
	}
	
	return &analysis.DifyAnalysisMetrics{
		TotalAnalyses:         totalAnalyses,
		SuccessfulAnalyses:    successfulAnalyses,
		FailedAnalyses:        failedAnalyses,
		AverageProcessingTime: avgProcessingTime,
		AverageConfidence:     avgConfidence,
		TotalTokenUsage:       totalTokenUsage,
		TotalCost:             totalCost,
		ByAnalysisType:        byAnalysisType,
		ByTime:                byTime,
		ErrorStats:            errorStats,
	}, nil
}

// 辅助方法

// modelToResult 将数据模型转换为领域对象
func (r *DifyAnalysisRepositoryImpl) modelToResult(model *DifyAnalysisResultModel) (*analysis.DifyAnalysisResult, error) {
	result := &analysis.DifyAnalysisResult{
		ID:             model.ID,
		AlertID:        model.AlertID,
		AnalysisType:   model.AnalysisType,
		ConversationID: model.ConversationID,
		MessageID:      model.MessageID,
		RawResponse:    model.RawResponse,
		Confidence:     model.Confidence,
		ProcessingTime: model.ProcessingTime,
		Status:         model.Status,
		Error:          model.Error,
		CreatedAt:      model.CreatedAt,
		UpdatedAt:      model.UpdatedAt,
	}
	
	// 反序列化结果
	if model.ResultJSON != "" {
		var resultData map[string]interface{}
		if err := json.Unmarshal([]byte(model.ResultJSON), &resultData); err != nil {
			r.logger.Error("Failed to unmarshal result",
				zap.String("task_id", model.ID),
				zap.Error(err),
			)
		} else {
			// 需要将 resultData 转换为 *analysis.AnalysisResult
			// 这里暂时设置为 nil，实际使用时需要根据具体结构进行转换
			result.Result = nil
		}
	}
	
	// 反序列化使用量
	if model.UsageJSON != "" {
		var usage map[string]interface{}
		if err := json.Unmarshal([]byte(model.UsageJSON), &usage); err != nil {
			r.logger.Error("Failed to unmarshal usage",
				zap.String("task_id", model.ID),
				zap.Error(err),
			)
		} else {
			// 需要将 usage 转换为 *analysis.DifyUsage
			// 这里暂时设置为 nil，实际使用时需要根据具体结构进行转换
			result.Usage = nil
		}
	}
	
	return result, nil
}

// getHourlyTrends 获取小时级趋势
func (r *DifyAnalysisRepositoryImpl) getHourlyTrends(ctx context.Context, query *gorm.DB, request *analysis.DifyTrendRequest) ([]analysis.TrendData, map[string]interface{}) {
	var trends []analysis.TrendData
	var hourlyStats []struct {
		Hour  time.Time
		Count int64
	}
	
	query.Select("DATE_TRUNC('hour', created_at) as hour, COUNT(*) as count").
		Group("DATE_TRUNC('hour', created_at)").
		Order("hour").
		Scan(&hourlyStats)
	
	for _, stat := range hourlyStats {
		trends = append(trends, analysis.TrendData{
			Timestamp: stat.Hour,
			Value:     float64(stat.Count),
		})
	}
	
	statistics := map[string]interface{}{
		"total_points": len(trends),
		"aggregation":  "hourly",
	}
	
	return trends, statistics
}

// getDailyTrends 获取日级趋势
func (r *DifyAnalysisRepositoryImpl) getDailyTrends(ctx context.Context, query *gorm.DB, request *analysis.DifyTrendRequest) ([]analysis.TrendData, map[string]interface{}) {
	var trends []analysis.TrendData
	var dailyStats []struct {
		Day   time.Time
		Count int64
	}
	
	query.Select("DATE(created_at) as day, COUNT(*) as count").
		Group("DATE(created_at)").
		Order("day").
		Scan(&dailyStats)
	
	for _, stat := range dailyStats {
		trends = append(trends, analysis.TrendData{
			Timestamp: stat.Day,
			Value:     float64(stat.Count),
		})
	}
	
	statistics := map[string]interface{}{
		"total_points": len(trends),
		"aggregation":  "daily",
	}
	
	return trends, statistics
}

// getWeeklyTrends 获取周级趋势
func (r *DifyAnalysisRepositoryImpl) getWeeklyTrends(ctx context.Context, query *gorm.DB, request *analysis.DifyTrendRequest) ([]analysis.TrendData, map[string]interface{}) {
	var trends []analysis.TrendData
	var weeklyStats []struct {
		Week  time.Time
		Count int64
	}
	
	query.Select("DATE_TRUNC('week', created_at) as week, COUNT(*) as count").
		Group("DATE_TRUNC('week', created_at)").
		Order("week").
		Scan(&weeklyStats)
	
	for _, stat := range weeklyStats {
		trends = append(trends, analysis.TrendData{
			Timestamp: stat.Week,
			Value:     float64(stat.Count),
		})
	}
	
	statistics := map[string]interface{}{
		"total_points": len(trends),
		"aggregation":  "weekly",
	}
	
	return trends, statistics
}

// generateInsights 生成洞察
func (r *DifyAnalysisRepositoryImpl) generateInsights(trends []analysis.TrendData, statistics map[string]interface{}) []string {
	insights := []string{}
	
	if len(trends) == 0 {
		return insights
	}
	
	// 计算趋势
	if len(trends) >= 2 {
		first := trends[0].Value
		last := trends[len(trends)-1].Value
		
		if last > first {
			insights = append(insights, "分析请求量呈上升趋势")
		} else if last < first {
			insights = append(insights, "分析请求量呈下降趋势")
		} else {
			insights = append(insights, "分析请求量保持稳定")
		}
	}
	
	// 计算平均值
	var total float64
	for _, trend := range trends {
		total += trend.Value
	}
	avg := total / float64(len(trends))
	insights = append(insights, fmt.Sprintf("平均每日分析请求量: %.1f", avg))
	
	return insights
}

// generateRecommendations 生成建议
func (r *DifyAnalysisRepositoryImpl) generateRecommendations(trends []analysis.TrendData, statistics map[string]interface{}) []string {
	recommendations := []string{}
	
	if len(trends) == 0 {
		return recommendations
	}
	
	// 基于趋势生成建议
	if len(trends) >= 2 {
		first := trends[0].Value
		last := trends[len(trends)-1].Value
		
		if last > first*1.5 {
			recommendations = append(recommendations, "分析请求量显著增加，建议考虑扩容")
		} else if last < first*0.5 {
			recommendations = append(recommendations, "分析请求量显著减少，可以考虑优化资源配置")
		}
	}
	
	// 通用建议
	recommendations = append(recommendations, "定期监控分析质量和性能指标")
	recommendations = append(recommendations, "优化分析提示词以提高分析准确性")
	
	return recommendations
}