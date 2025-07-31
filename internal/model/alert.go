package model

import (
	"encoding/json"
	"errors"
	"time"

	"gorm.io/gorm"
)

// Alert 状态常量
const (
	AlertStatusNew          = "new"
	AlertStatusAcknowledged = "acknowledged"
	AlertStatusResolved     = "resolved"
)

// Alert 级别常量
const (
	AlertLevelCritical = "critical"
	AlertLevelHigh     = "high"
	AlertLevelMedium   = "medium"
	AlertLevelLow      = "low"
)

// Alert 告警记录
type Alert struct {
	ID          uint           `json:"id" gorm:"primarykey"`
	CreatedAt   time.Time      `json:"-"`
	UpdatedAt   time.Time      `json:"-"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`
	Name        string         `json:"name" gorm:"size:255;not null"`
	Title       string         `json:"title" gorm:"size:255;not null"`
	Level       string         `json:"level" gorm:"type:varchar(50);not null"`
	Status      string         `json:"status" gorm:"type:varchar(50);not null;default:'new'"`
	Source      string         `json:"source" gorm:"type:varchar(255);not null"`
	Content     string         `json:"content" gorm:"type:text;not null"`
	Labels      string         `json:"labels,omitempty" gorm:"type:text"`
	RuleID      uint           `json:"rule_id" gorm:"not null"`
	TemplateID  uint           `json:"template_id,omitempty"`
	GroupID     uint           `json:"group_id,omitempty"`
	Handler     string         `json:"handler,omitempty" gorm:"type:varchar(100)"`
	HandleTime  *time.Time     `json:"-"`
	HandleNote  string         `json:"handle_note,omitempty" gorm:"type:text"`
	Analysis           string         `json:"analysis,omitempty" gorm:"type:text"`
	AnalysisStatus     string         `json:"analysis_status" gorm:"type:varchar(20) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;default:'pending'"`
	AnalysisResult     string         `json:"analysis_result" gorm:"type:json"`
	AISummary          string         `json:"ai_summary" gorm:"type:text CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci"`
	SimilarAlerts      string         `json:"similar_alerts" gorm:"type:json"`
	ResolutionSuggestion string       `json:"resolution_suggestion" gorm:"type:text CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci"`
	Fingerprint        string         `json:"fingerprint" gorm:"type:varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci"`
	NotifyTime         *time.Time     `json:"-"`
	NotifyCount        int            `json:"notify_count,omitempty" gorm:"default:0"`
	Severity           string         `json:"severity" gorm:"type:varchar(20);not null;default:'medium'"`
}

// Validate 验证告警数据
func (a *Alert) Validate() error {
	if a.Name == "" {
		return errors.New("name is required")
	}
	if a.Title == "" {
		return errors.New("title is required")
	}
	if a.Content == "" {
		return errors.New("content is required")
	}
	if a.Source == "" {
		return errors.New("source is required")
	}
	if !isValidLevel(a.Level) {
		return errors.New("invalid alert level")
	}
	if !isValidStatus(a.Status) {
		return errors.New("invalid alert status")
	}
	return nil
}

// isValidLevel 验证告警级别
func isValidLevel(level string) bool {
	switch level {
	case AlertLevelCritical, AlertLevelHigh, AlertLevelMedium, AlertLevelLow:
		return true
	}
	return false
}

// isValidStatus 验证告警状态
func isValidStatus(status string) bool {
	switch status {
	case AlertStatusNew, AlertStatusAcknowledged, AlertStatusResolved:
		return true
	}
	return false
}

// AlertResponse 告警响应
type AlertResponse struct {
	ID          uint   `json:"id"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
	Name        string `json:"name"`
	Title       string `json:"title"`
	Level       string `json:"level"`
	Status      string `json:"status"`
	Source      string `json:"source"`
	Content     string `json:"content"`
	Labels      string `json:"labels,omitempty"`
	RuleID      uint   `json:"rule_id"`
	TemplateID  uint   `json:"template_id,omitempty"`
	GroupID     uint   `json:"group_id,omitempty"`
	Handler     string `json:"handler,omitempty"`
	HandleTime  string `json:"handle_time,omitempty"`
	HandleNote  string `json:"handle_note,omitempty"`
	Analysis             string `json:"analysis,omitempty"`
	AnalysisStatus       string `json:"analysis_status"`
	AnalysisResult       string `json:"analysis_result,omitempty"`
	AISummary            string `json:"ai_summary,omitempty"`
	SimilarAlerts        string `json:"similar_alerts,omitempty"`
	ResolutionSuggestion string `json:"resolution_suggestion,omitempty"`
	Fingerprint          string `json:"fingerprint,omitempty"`
	NotifyTime           string `json:"notify_time,omitempty"`
	NotifyCount          int    `json:"notify_count,omitempty"`
	Severity             string `json:"severity"`
}

// ToResponse 转换为响应格式
func (a *Alert) ToResponse() *AlertResponse {
	resp := &AlertResponse{
		ID:          a.ID,
		CreatedAt:   a.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   a.UpdatedAt.Format(time.RFC3339),
		Name:        a.Name,
		Title:       a.Title,
		Level:       a.Level,
		Status:      a.Status,
		Source:      a.Source,
		Content:     a.Content,
		Labels:      a.Labels,
		RuleID:      a.RuleID,
		TemplateID:  a.TemplateID,
		GroupID:     a.GroupID,
		Handler:     a.Handler,
		HandleNote:  a.HandleNote,
		Analysis:             a.Analysis,
		AnalysisStatus:       a.AnalysisStatus,
		AnalysisResult:       a.AnalysisResult,
		AISummary:            a.AISummary,
		SimilarAlerts:        a.SimilarAlerts,
		ResolutionSuggestion: a.ResolutionSuggestion,
		Fingerprint:          a.Fingerprint,
		NotifyCount:          a.NotifyCount,
		Severity:             a.Severity,
	}
	if a.HandleTime != nil {
		resp.HandleTime = a.HandleTime.Format(time.RFC3339)
	}
	if a.NotifyTime != nil {
		resp.NotifyTime = a.NotifyTime.Format(time.RFC3339)
	}
	return resp
}

// MarshalBinary 实现 encoding.BinaryMarshaler 接口
func (a *Alert) MarshalBinary() ([]byte, error) {
	return json.Marshal(a)
}

// UnmarshalBinary 实现 encoding.BinaryUnmarshaler 接口
func (a *Alert) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, a)
}

// GetAnalysisResultMap 获取分析结果映射
func (a *Alert) GetAnalysisResultMap() (map[string]interface{}, error) {
	if a.AnalysisResult == "" {
		return make(map[string]interface{}), nil
	}
	var result map[string]interface{}
	err := json.Unmarshal([]byte(a.AnalysisResult), &result)
	return result, err
}

// SetAnalysisResultMap 设置分析结果映射
func (a *Alert) SetAnalysisResultMap(result map[string]interface{}) error {
	if result == nil {
		a.AnalysisResult = ""
		return nil
	}
	data, err := json.Marshal(result)
	if err != nil {
		return err
	}
	a.AnalysisResult = string(data)
	return nil
}

// GetSimilarAlertsList 获取相似告警列表
func (a *Alert) GetSimilarAlertsList() ([]SimilarAlert, error) {
	if a.SimilarAlerts == "" {
		return make([]SimilarAlert, 0), nil
	}
	var alerts []SimilarAlert
	err := json.Unmarshal([]byte(a.SimilarAlerts), &alerts)
	return alerts, err
}

// SetSimilarAlertsList 设置相似告警列表
func (a *Alert) SetSimilarAlertsList(alerts []SimilarAlert) error {
	if alerts == nil {
		a.SimilarAlerts = ""
		return nil
	}
	data, err := json.Marshal(alerts)
	if err != nil {
		return err
	}
	a.SimilarAlerts = string(data)
	return nil
}

// IsAnalyzed 检查是否已分析
func (a *Alert) IsAnalyzed() bool {
	return a.AnalysisStatus == "completed" || a.AnalysisStatus == "analyzed"
}

// IsAnalyzing 检查是否正在分析
func (a *Alert) IsAnalyzing() bool {
	return a.AnalysisStatus == "analyzing" || a.AnalysisStatus == "processing"
}

// AnalysisFailed 检查分析是否失败
func (a *Alert) AnalysisFailed() bool {
	return a.AnalysisStatus == "failed" || a.AnalysisStatus == "error"
}

// SimilarAlert 相似告警
type SimilarAlert struct {
	Alert      Alert   `json:"alert"`
	Similarity float64 `json:"similarity"`
}
