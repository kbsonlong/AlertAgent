package feature

import (
	"context"
	"fmt"
	"math"
	"sync"
	"time"

	"go.uber.org/zap"
)

// AIMetrics AI模型指标
type AIMetrics struct {
	Accuracy     float64   `json:"accuracy"`      // 准确率
	Confidence   float64   `json:"confidence"`    // 置信度
	Latency      int       `json:"latency"`       // 延迟(ms)
	SuccessRate  float64   `json:"success_rate"`  // 成功率
	ErrorRate    float64   `json:"error_rate"`    // 错误率
	Timestamp    time.Time `json:"timestamp"`     // 时间戳
	SampleCount  int       `json:"sample_count"`  // 样本数量
}

// MaturityLevel AI模型成熟度等级
type MaturityLevel string

const (
	MaturityLevelLow      MaturityLevel = "low"      // 低成熟度
	MaturityLevelMedium   MaturityLevel = "medium"   // 中等成熟度
	MaturityLevelHigh     MaturityLevel = "high"     // 高成熟度
	MaturityLevelCritical MaturityLevel = "critical" // 关键成熟度
)

// MaturityAssessment 成熟度评估结果
type MaturityAssessment struct {
	FeatureName     FeatureName   `json:"feature_name"`
	Level           MaturityLevel `json:"level"`
	Score           float64       `json:"score"`           // 综合评分 (0-1)
	Metrics         AIMetrics     `json:"metrics"`         // 当前指标
	Requirements    AIMaturityRequirement `json:"requirements"` // 要求
	Recommendations []string      `json:"recommendations"` // 改进建议
	ShouldDegrade   bool          `json:"should_degrade"`  // 是否应该降级
	AssessedAt      time.Time     `json:"assessed_at"`     // 评估时间
}

// AIMaturityEvaluator AI模型成熟度评估器
type AIMaturityEvaluator struct {
	logger         *zap.Logger
	metricsStore   map[FeatureName][]AIMetrics // 指标存储
	assessments    map[FeatureName]*MaturityAssessment // 评估结果缓存
	mutex          sync.RWMutex
	degradeCallbacks map[FeatureName][]func(MaturityAssessment) // 降级回调
}

// NewAIMaturityEvaluator 创建AI模型成熟度评估器
func NewAIMaturityEvaluator(logger *zap.Logger) *AIMaturityEvaluator {
	return &AIMaturityEvaluator{
		logger:           logger,
		metricsStore:     make(map[FeatureName][]AIMetrics),
		assessments:      make(map[FeatureName]*MaturityAssessment),
		degradeCallbacks: make(map[FeatureName][]func(MaturityAssessment)),
	}
}

// RecordMetrics 记录AI模型指标
func (e *AIMaturityEvaluator) RecordMetrics(featureName FeatureName, metrics AIMetrics) {
	e.mutex.Lock()
	defer e.mutex.Unlock()
	
	if e.metricsStore[featureName] == nil {
		e.metricsStore[featureName] = make([]AIMetrics, 0)
	}
	
	metrics.Timestamp = time.Now()
	e.metricsStore[featureName] = append(e.metricsStore[featureName], metrics)
	
	// 保持最近1000条记录
	if len(e.metricsStore[featureName]) > 1000 {
		e.metricsStore[featureName] = e.metricsStore[featureName][len(e.metricsStore[featureName])-1000:]
	}
	
	e.logger.Debug("AI metrics recorded",
		zap.String("feature", string(featureName)),
		zap.Float64("accuracy", metrics.Accuracy),
		zap.Float64("confidence", metrics.Confidence),
		zap.Int("latency", metrics.Latency))
}

// EvaluateMaturity 评估AI模型成熟度
func (e *AIMaturityEvaluator) EvaluateMaturity(ctx context.Context, featureName FeatureName, requirements AIMaturityRequirement) bool {
	e.mutex.Lock()
	defer e.mutex.Unlock()
	
	// 获取最近的指标数据
	recentMetrics := e.getRecentMetrics(featureName, requirements.EvaluationPeriod)
	if len(recentMetrics) == 0 {
		e.logger.Warn("No metrics available for maturity evaluation",
			zap.String("feature", string(featureName)))
		return false
	}
	
	// 计算平均指标
	avgMetrics := e.calculateAverageMetrics(recentMetrics)
	
	// 进行成熟度评估
	assessment := e.performAssessment(featureName, avgMetrics, requirements)
	
	// 缓存评估结果
	e.assessments[featureName] = assessment
	
	// 检查是否需要降级
	if assessment.ShouldDegrade {
		e.triggerDegradeCallbacks(featureName, *assessment)
	}
	
	e.logger.Info("AI maturity evaluated",
		zap.String("feature", string(featureName)),
		zap.String("level", string(assessment.Level)),
		zap.Float64("score", assessment.Score),
		zap.Bool("should_degrade", assessment.ShouldDegrade))
	
	// 返回是否满足成熟度要求
	return !assessment.ShouldDegrade && assessment.Score >= 0.7 // 70%以上认为合格
}

// getRecentMetrics 获取最近的指标数据
func (e *AIMaturityEvaluator) getRecentMetrics(featureName FeatureName, evaluationPeriodHours int) []AIMetrics {
	metrics, exists := e.metricsStore[featureName]
	if !exists {
		return nil
	}
	
	cutoffTime := time.Now().Add(-time.Duration(evaluationPeriodHours) * time.Hour)
	recentMetrics := make([]AIMetrics, 0)
	
	for _, metric := range metrics {
		if metric.Timestamp.After(cutoffTime) {
			recentMetrics = append(recentMetrics, metric)
		}
	}
	
	return recentMetrics
}

// calculateAverageMetrics 计算平均指标
func (e *AIMaturityEvaluator) calculateAverageMetrics(metrics []AIMetrics) AIMetrics {
	if len(metrics) == 0 {
		return AIMetrics{}
	}
	
	var totalAccuracy, totalConfidence, totalSuccessRate, totalErrorRate float64
	var totalLatency, totalSamples int
	
	for _, metric := range metrics {
		totalAccuracy += metric.Accuracy
		totalConfidence += metric.Confidence
		totalSuccessRate += metric.SuccessRate
		totalErrorRate += metric.ErrorRate
		totalLatency += metric.Latency
		totalSamples += metric.SampleCount
	}
	
	count := float64(len(metrics))
	return AIMetrics{
		Accuracy:    totalAccuracy / count,
		Confidence:  totalConfidence / count,
		Latency:     int(float64(totalLatency) / count),
		SuccessRate: totalSuccessRate / count,
		ErrorRate:   totalErrorRate / count,
		SampleCount: totalSamples,
		Timestamp:   time.Now(),
	}
}

// performAssessment 执行成熟度评估
func (e *AIMaturityEvaluator) performAssessment(featureName FeatureName, metrics AIMetrics, requirements AIMaturityRequirement) *MaturityAssessment {
	assessment := &MaturityAssessment{
		FeatureName:  featureName,
		Metrics:      metrics,
		Requirements: requirements,
		AssessedAt:   time.Now(),
	}
	
	// 计算各项指标的得分
	accuracyScore := e.calculateScore(metrics.Accuracy, requirements.MinAccuracy, 1.0)
	confidenceScore := e.calculateScore(metrics.Confidence, requirements.MinConfidence, 1.0)
	latencyScore := e.calculateLatencyScore(metrics.Latency, requirements.MaxLatency)
	successRateScore := e.calculateScore(metrics.SuccessRate, requirements.MinSuccessRate, 1.0)
	
	// 计算综合得分（加权平均）
	weights := map[string]float64{
		"accuracy":     0.3,
		"confidence":   0.25,
		"latency":      0.2,
		"success_rate": 0.25,
	}
	
	assessment.Score = accuracyScore*weights["accuracy"] +
		confidenceScore*weights["confidence"] +
		latencyScore*weights["latency"] +
		successRateScore*weights["success_rate"]
	
	// 确定成熟度等级
	assessment.Level = e.determineMaturityLevel(assessment.Score)
	
	// 检查是否应该降级
	assessment.ShouldDegrade = e.shouldDegrade(metrics, requirements)
	
	// 生成改进建议
	assessment.Recommendations = e.generateRecommendations(metrics, requirements)
	
	return assessment
}

// calculateScore 计算单项指标得分
func (e *AIMaturityEvaluator) calculateScore(actual, min, max float64) float64 {
	if actual < min {
		return 0.0
	}
	if actual >= max {
		return 1.0
	}
	return (actual - min) / (max - min)
}

// calculateLatencyScore 计算延迟得分（越低越好）
func (e *AIMaturityEvaluator) calculateLatencyScore(actual, max int) float64 {
	if actual > max {
		return 0.0
	}
	if actual <= max/2 {
		return 1.0
	}
	return 1.0 - float64(actual-max/2)/float64(max/2)
}

// determineMaturityLevel 确定成熟度等级
func (e *AIMaturityEvaluator) determineMaturityLevel(score float64) MaturityLevel {
	switch {
	case score >= 0.9:
		return MaturityLevelCritical
	case score >= 0.8:
		return MaturityLevelHigh
	case score >= 0.6:
		return MaturityLevelMedium
	default:
		return MaturityLevelLow
	}
}

// shouldDegrade 判断是否应该降级
func (e *AIMaturityEvaluator) shouldDegrade(metrics AIMetrics, requirements AIMaturityRequirement) bool {
	// 任何一项关键指标不达标都应该降级
	if metrics.Accuracy < requirements.MinAccuracy*0.9 { // 允许10%的容差
		return true
	}
	if metrics.Confidence < requirements.MinConfidence*0.9 {
		return true
	}
	if metrics.Latency > int(float64(requirements.MaxLatency)*1.2) { // 允许20%的容差
		return true
	}
	if metrics.SuccessRate < requirements.MinSuccessRate*0.95 { // 允许5%的容差
		return true
	}
	
	return false
}

// generateRecommendations 生成改进建议
func (e *AIMaturityEvaluator) generateRecommendations(metrics AIMetrics, requirements AIMaturityRequirement) []string {
	recommendations := make([]string, 0)
	
	if metrics.Accuracy < requirements.MinAccuracy {
		recommendations = append(recommendations, 
			fmt.Sprintf("提高模型准确率：当前%.2f%%，要求%.2f%%", 
				metrics.Accuracy*100, requirements.MinAccuracy*100))
	}
	
	if metrics.Confidence < requirements.MinConfidence {
		recommendations = append(recommendations, 
			fmt.Sprintf("提高模型置信度：当前%.2f%%，要求%.2f%%", 
				metrics.Confidence*100, requirements.MinConfidence*100))
	}
	
	if metrics.Latency > requirements.MaxLatency {
		recommendations = append(recommendations, 
			fmt.Sprintf("优化响应延迟：当前%dms，要求<%dms", 
				metrics.Latency, requirements.MaxLatency))
	}
	
	if metrics.SuccessRate < requirements.MinSuccessRate {
		recommendations = append(recommendations, 
			fmt.Sprintf("提高成功率：当前%.2f%%，要求%.2f%%", 
				metrics.SuccessRate*100, requirements.MinSuccessRate*100))
	}
	
	if metrics.ErrorRate > 0.05 { // 错误率超过5%
		recommendations = append(recommendations, 
			fmt.Sprintf("降低错误率：当前%.2f%%", metrics.ErrorRate*100))
	}
	
	return recommendations
}

// GetAssessment 获取成熟度评估结果
func (e *AIMaturityEvaluator) GetAssessment(featureName FeatureName) (*MaturityAssessment, error) {
	e.mutex.RLock()
	defer e.mutex.RUnlock()
	
	assessment, exists := e.assessments[featureName]
	if !exists {
		return nil, fmt.Errorf("no assessment found for feature %s", featureName)
	}
	
	// 返回副本
	assessmentCopy := *assessment
	return &assessmentCopy, nil
}

// GetAllAssessments 获取所有成熟度评估结果
func (e *AIMaturityEvaluator) GetAllAssessments() map[FeatureName]*MaturityAssessment {
	e.mutex.RLock()
	defer e.mutex.RUnlock()
	
	result := make(map[FeatureName]*MaturityAssessment)
	for name, assessment := range e.assessments {
		assessmentCopy := *assessment
		result[name] = &assessmentCopy
	}
	
	return result
}

// RegisterDegradeCallback 注册降级回调
func (e *AIMaturityEvaluator) RegisterDegradeCallback(featureName FeatureName, callback func(MaturityAssessment)) {
	e.mutex.Lock()
	defer e.mutex.Unlock()
	
	if e.degradeCallbacks[featureName] == nil {
		e.degradeCallbacks[featureName] = make([]func(MaturityAssessment), 0)
	}
	
	e.degradeCallbacks[featureName] = append(e.degradeCallbacks[featureName], callback)
}

// triggerDegradeCallbacks 触发降级回调
func (e *AIMaturityEvaluator) triggerDegradeCallbacks(featureName FeatureName, assessment MaturityAssessment) {
	callbacks, exists := e.degradeCallbacks[featureName]
	if !exists {
		return
	}
	
	for _, callback := range callbacks {
		go func(cb func(MaturityAssessment)) {
			defer func() {
				if r := recover(); r != nil {
					e.logger.Error("Degrade callback panic",
						zap.String("feature", string(featureName)),
						zap.Any("panic", r))
				}
			}()
			cb(assessment)
		}(callback)
	}
}

// GetMetricsHistory 获取指标历史
func (e *AIMaturityEvaluator) GetMetricsHistory(featureName FeatureName, hours int) []AIMetrics {
	e.mutex.RLock()
	defer e.mutex.RUnlock()
	
	return e.getRecentMetrics(featureName, hours)
}

// ClearMetrics 清理指标数据
func (e *AIMaturityEvaluator) ClearMetrics(featureName FeatureName) {
	e.mutex.Lock()
	defer e.mutex.Unlock()
	
	delete(e.metricsStore, featureName)
	delete(e.assessments, featureName)
	
	e.logger.Info("Metrics cleared", zap.String("feature", string(featureName)))
}

// GetMaturityTrend 获取成熟度趋势
func (e *AIMaturityEvaluator) GetMaturityTrend(featureName FeatureName, hours int) []float64 {
	e.mutex.RLock()
	defer e.mutex.RUnlock()
	
	metrics := e.getRecentMetrics(featureName, hours)
	if len(metrics) == 0 {
		return nil
	}
	
	// 按小时分组计算趋势
	hourlyScores := make(map[int][]float64)
	for _, metric := range metrics {
		hour := metric.Timestamp.Hour()
		if hourlyScores[hour] == nil {
			hourlyScores[hour] = make([]float64, 0)
		}
		
		// 简化的评分计算
		score := (metric.Accuracy + metric.Confidence + metric.SuccessRate) / 3.0
		hourlyScores[hour] = append(hourlyScores[hour], score)
	}
	
	// 计算每小时平均分
	trend := make([]float64, 0)
	for hour := 0; hour < 24; hour++ {
		if scores, exists := hourlyScores[hour]; exists {
			sum := 0.0
			for _, score := range scores {
				sum += score
			}
			trend = append(trend, sum/float64(len(scores)))
		} else {
			trend = append(trend, 0.0)
		}
	}
	
	return trend
}

// PredictMaturityDegradation 预测成熟度降级风险
func (e *AIMaturityEvaluator) PredictMaturityDegradation(featureName FeatureName) (float64, error) {
	e.mutex.RLock()
	defer e.mutex.RUnlock()
	
	// 获取最近24小时的指标
	metrics := e.getRecentMetrics(featureName, 24)
	if len(metrics) < 10 { // 需要足够的数据点
		return 0.0, fmt.Errorf("insufficient data for prediction")
	}
	
	// 简单的线性回归预测趋势
	n := len(metrics)
	var sumX, sumY, sumXY, sumX2 float64
	
	for i, metric := range metrics {
		x := float64(i)
		y := (metric.Accuracy + metric.Confidence + metric.SuccessRate) / 3.0
		
		sumX += x
		sumY += y
		sumXY += x * y
		sumX2 += x * x
	}
	
	// 计算斜率（趋势）
	slope := (float64(n)*sumXY - sumX*sumY) / (float64(n)*sumX2 - sumX*sumX)
	
	// 如果斜率为负且绝对值较大，则降级风险较高
	if slope < 0 {
		risk := math.Min(math.Abs(slope)*100, 1.0) // 转换为0-1的风险值
		return risk, nil
	}
	
	return 0.0, nil
}