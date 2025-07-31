package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"alert_agent/internal/model"
	"alert_agent/internal/pkg/database"
	"alert_agent/internal/pkg/logger"
	"alert_agent/internal/pkg/queue"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// generateID 生成UUID
func generateID() string {
	return uuid.New().String()
}

// ConfigSyncExceptionHandler 配置同步异常处理器
type ConfigSyncExceptionHandler struct {
	monitorService *ConfigSyncMonitor
	alertService   *AlertService
	queueProducer  queue.TaskProducer
}

// NewConfigSyncExceptionHandler 创建配置同步异常处理器
func NewConfigSyncExceptionHandler() *ConfigSyncExceptionHandler {
	return &ConfigSyncExceptionHandler{
		monitorService: NewConfigSyncMonitor(),
		alertService:   nil, // TODO: Initialize with proper dependencies
		queueProducer:  nil, // TODO: Initialize with proper dependencies
	}
}

// SyncException 同步异常
type SyncException struct {
	ID           string    `json:"id" gorm:"type:varchar(36);primaryKey"`
	ClusterID    string    `json:"cluster_id" gorm:"type:varchar(100);not null;index"`
	ConfigType   string    `json:"config_type" gorm:"type:varchar(50);not null;index"`
	ExceptionType string   `json:"exception_type" gorm:"type:varchar(50);not null;index"` // timeout, connection_error, validation_error, etc.
	ErrorMessage string    `json:"error_message" gorm:"type:text"`
	Severity     string    `json:"severity" gorm:"type:varchar(20);not null;default:'medium'"` // low, medium, high, critical
	Status       string    `json:"status" gorm:"type:varchar(20);not null;default:'open'"` // open, investigating, resolved
	FirstOccurred time.Time `json:"first_occurred"`
	LastOccurred  time.Time `json:"last_occurred"`
	OccurrenceCount int64   `json:"occurrence_count" gorm:"default:1"`
	AutoRetryCount  int     `json:"auto_retry_count" gorm:"default:0"`
	MaxAutoRetry    int     `json:"max_auto_retry" gorm:"default:3"`
	NextRetryAt     *time.Time `json:"next_retry_at"`
	ResolvedAt      *time.Time `json:"resolved_at"`
	ResolvedBy      string    `json:"resolved_by"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// TableName 指定表名
func (SyncException) TableName() string {
	return "config_sync_exceptions"
}

// BeforeCreate GORM钩子：创建前生成ID
func (se *SyncException) BeforeCreate(tx *gorm.DB) error {
	if se.ID == "" {
		se.ID = generateID()
	}
	return nil
}

// ExceptionAnalysis 异常分析结果
type ExceptionAnalysis struct {
	ExceptionID     string            `json:"exception_id"`
	RootCause       string            `json:"root_cause"`
	PossibleCauses  []string          `json:"possible_causes"`
	SuggestedActions []string         `json:"suggested_actions"`
	RelatedExceptions []string        `json:"related_exceptions"`
	Confidence      float64           `json:"confidence"`
	AnalysisTime    time.Time         `json:"analysis_time"`
	Metadata        map[string]interface{} `json:"metadata"`
}

// DetectSyncExceptions 检测同步异常
func (cseh *ConfigSyncExceptionHandler) DetectSyncExceptions(ctx context.Context) error {
	logger.L.Debug("Detecting sync exceptions")

	// 获取最近的同步状态
	var statuses []model.ConfigSyncStatus
	err := database.DB.WithContext(ctx).
		Where("updated_at >= ?", time.Now().Add(-1*time.Hour)).
		Find(&statuses).Error
	if err != nil {
		return fmt.Errorf("failed to get sync statuses: %w", err)
	}

	for _, status := range statuses {
		if err := cseh.analyzeStatus(ctx, status); err != nil {
			logger.L.Error("Failed to analyze sync status",
				zap.String("cluster_id", status.ClusterID),
				zap.String("config_type", status.ConfigType),
				zap.Error(err),
			)
		}
	}

	// 检查长时间未同步的情况
	if err := cseh.detectStaleSyncs(ctx); err != nil {
		logger.L.Error("Failed to detect stale syncs", zap.Error(err))
	}

	return nil
}

// analyzeStatus 分析单个同步状态
func (cseh *ConfigSyncExceptionHandler) analyzeStatus(ctx context.Context, status model.ConfigSyncStatus) error {
	// 检查是否为失败状态
	if status.SyncStatus == "failed" && status.ErrorMessage != "" {
		return cseh.handleSyncFailure(ctx, status)
	}

	// 检查同步延迟
	if status.SyncTime != nil {
		delay := time.Since(*status.SyncTime)
		if delay > 10*time.Minute { // 超过10分钟未同步
			return cseh.handleSyncDelay(ctx, status, delay)
		}
	}

	return nil
}

// handleSyncFailure 处理同步失败
func (cseh *ConfigSyncExceptionHandler) handleSyncFailure(ctx context.Context, status model.ConfigSyncStatus) error {
	exceptionType := cseh.classifyError(status.ErrorMessage)
	severity := cseh.calculateSeverity(exceptionType, status.ErrorMessage)

	// 查找是否已存在相同异常
	var existingException SyncException
	err := database.DB.WithContext(ctx).
		Where("cluster_id = ? AND config_type = ? AND exception_type = ? AND status != 'resolved'",
			status.ClusterID, status.ConfigType, exceptionType).
		First(&existingException).Error

	if err == gorm.ErrRecordNotFound {
		// 创建新异常
		exception := SyncException{
			ClusterID:      status.ClusterID,
			ConfigType:     status.ConfigType,
			ExceptionType:  exceptionType,
			ErrorMessage:   status.ErrorMessage,
			Severity:       severity,
			Status:         "open",
			FirstOccurred:  time.Now(),
			LastOccurred:   time.Now(),
			OccurrenceCount: 1,
		}

		if err := database.DB.WithContext(ctx).Create(&exception).Error; err != nil {
			return fmt.Errorf("failed to create sync exception: %w", err)
		}

		// 发送告警
		if err := cseh.sendExceptionAlert(ctx, &exception); err != nil {
			logger.L.Error("Failed to send exception alert", zap.Error(err))
		}

		// 触发自动重试
		if cseh.shouldAutoRetry(exceptionType) {
			if err := cseh.scheduleAutoRetry(ctx, &exception); err != nil {
				logger.L.Error("Failed to schedule auto retry", zap.Error(err))
			}
		}

		logger.L.Info("New sync exception created",
			zap.String("exception_id", exception.ID),
			zap.String("cluster_id", status.ClusterID),
			zap.String("config_type", status.ConfigType),
			zap.String("exception_type", exceptionType),
		)
	} else if err != nil {
		return fmt.Errorf("failed to query existing exception: %w", err)
	} else {
		// 更新现有异常
		existingException.LastOccurred = time.Now()
		existingException.OccurrenceCount++
		existingException.ErrorMessage = status.ErrorMessage

		// 如果异常频繁发生，提升严重程度
		if existingException.OccurrenceCount >= 5 {
			existingException.Severity = cseh.escalateSeverity(existingException.Severity)
		}

		if err := database.DB.WithContext(ctx).Save(&existingException).Error; err != nil {
			return fmt.Errorf("failed to update sync exception: %w", err)
		}

		logger.L.Info("Sync exception updated",
			zap.String("exception_id", existingException.ID),
			zap.Int64("occurrence_count", existingException.OccurrenceCount),
		)
	}

	return nil
}

// handleSyncDelay 处理同步延迟
func (cseh *ConfigSyncExceptionHandler) handleSyncDelay(ctx context.Context, status model.ConfigSyncStatus, delay time.Duration) error {
	severity := "medium"
	if delay > 30*time.Minute {
		severity = "high"
	}
	if delay > 1*time.Hour {
		severity = "critical"
	}

	// 查找是否已存在延迟异常
	var existingException SyncException
	err := database.DB.WithContext(ctx).
		Where("cluster_id = ? AND config_type = ? AND exception_type = 'sync_delay' AND status != 'resolved'",
			status.ClusterID, status.ConfigType).
		First(&existingException).Error

	if err == gorm.ErrRecordNotFound {
		// 创建新的延迟异常
		exception := SyncException{
			ClusterID:      status.ClusterID,
			ConfigType:     status.ConfigType,
			ExceptionType:  "sync_delay",
			ErrorMessage:   fmt.Sprintf("Sync delayed for %v", delay),
			Severity:       severity,
			Status:         "open",
			FirstOccurred:  time.Now(),
			LastOccurred:   time.Now(),
			OccurrenceCount: 1,
		}

		if err := database.DB.WithContext(ctx).Create(&exception).Error; err != nil {
			return fmt.Errorf("failed to create sync delay exception: %w", err)
		}

		// 发送告警
		if err := cseh.sendExceptionAlert(ctx, &exception); err != nil {
			logger.L.Error("Failed to send delay alert", zap.Error(err))
		}
	} else if err != nil {
		return fmt.Errorf("failed to query existing delay exception: %w", err)
	} else {
		// 更新现有延迟异常
		existingException.LastOccurred = time.Now()
		existingException.ErrorMessage = fmt.Sprintf("Sync delayed for %v", delay)
		existingException.Severity = severity

		if err := database.DB.WithContext(ctx).Save(&existingException).Error; err != nil {
			return fmt.Errorf("failed to update sync delay exception: %w", err)
		}
	}

	return nil
}

// detectStaleSyncs 检测长时间未同步的情况
func (cseh *ConfigSyncExceptionHandler) detectStaleSyncs(ctx context.Context) error {
	// 查找超过1小时未同步的配置
	var staleStatuses []model.ConfigSyncStatus
	err := database.DB.WithContext(ctx).
		Where("sync_time IS NULL OR sync_time < ?", time.Now().Add(-1*time.Hour)).
		Find(&staleStatuses).Error
	if err != nil {
		return fmt.Errorf("failed to get stale sync statuses: %w", err)
	}

	for _, status := range staleStatuses {
		var timeSinceSync time.Duration
		if status.SyncTime == nil {
			timeSinceSync = time.Since(status.CreatedAt)
		} else {
			timeSinceSync = time.Since(*status.SyncTime)
		}

		if timeSinceSync > 1*time.Hour {
			if err := cseh.handleSyncDelay(ctx, status, timeSinceSync); err != nil {
				logger.L.Error("Failed to handle stale sync",
					zap.String("cluster_id", status.ClusterID),
					zap.String("config_type", status.ConfigType),
					zap.Error(err),
				)
			}
		}
	}

	return nil
}

// classifyError 分类错误类型
func (cseh *ConfigSyncExceptionHandler) classifyError(errorMessage string) string {
	errorLower := strings.ToLower(errorMessage)

	if strings.Contains(errorLower, "timeout") || strings.Contains(errorLower, "deadline") {
		return "timeout"
	}
	if strings.Contains(errorLower, "connection") || strings.Contains(errorLower, "network") {
		return "connection_error"
	}
	if strings.Contains(errorLower, "validation") || strings.Contains(errorLower, "invalid") {
		return "validation_error"
	}
	if strings.Contains(errorLower, "permission") || strings.Contains(errorLower, "unauthorized") {
		return "permission_error"
	}
	if strings.Contains(errorLower, "not found") || strings.Contains(errorLower, "404") {
		return "not_found"
	}
	if strings.Contains(errorLower, "server error") || strings.Contains(errorLower, "500") {
		return "server_error"
	}

	return "unknown_error"
}

// calculateSeverity 计算严重程度
func (cseh *ConfigSyncExceptionHandler) calculateSeverity(exceptionType, errorMessage string) string {
	switch exceptionType {
	case "timeout", "connection_error":
		return "high"
	case "server_error":
		return "critical"
	case "permission_error":
		return "high"
	case "validation_error":
		return "medium"
	case "not_found":
		return "medium"
	default:
		return "low"
	}
}

// escalateSeverity 提升严重程度
func (cseh *ConfigSyncExceptionHandler) escalateSeverity(currentSeverity string) string {
	switch currentSeverity {
	case "low":
		return "medium"
	case "medium":
		return "high"
	case "high":
		return "critical"
	default:
		return currentSeverity
	}
}

// shouldAutoRetry 判断是否应该自动重试
func (cseh *ConfigSyncExceptionHandler) shouldAutoRetry(exceptionType string) bool {
	switch exceptionType {
	case "timeout", "connection_error", "server_error":
		return true
	default:
		return false
	}
}

// scheduleAutoRetry 安排自动重试
func (cseh *ConfigSyncExceptionHandler) scheduleAutoRetry(ctx context.Context, exception *SyncException) error {
	if exception.AutoRetryCount >= exception.MaxAutoRetry {
		return nil // 已达到最大重试次数
	}

	// 计算下次重试时间（指数退避）
	retryDelay := time.Duration(1<<exception.AutoRetryCount) * time.Minute
	nextRetryAt := time.Now().Add(retryDelay)

	exception.NextRetryAt = &nextRetryAt
	exception.AutoRetryCount++

	if err := database.DB.WithContext(ctx).Save(exception).Error; err != nil {
		return fmt.Errorf("failed to update exception for retry: %w", err)
	}

	// 发送重试任务到队列
	// retryTask := map[string]interface{}{
	//	"exception_id": exception.ID,
	//	"cluster_id":   exception.ClusterID,
	//	"config_type":  exception.ConfigType,
	//	"retry_count":  exception.AutoRetryCount,
	// }

	// TODO: Fix PublishDelayed method
	// if err := cseh.queueProducer.PublishDelayed(ctx, "config_sync_retry", retryTask, retryDelay); err != nil {
	//	return fmt.Errorf("failed to schedule retry task: %w", err)
	// }

	logger.L.Info("Auto retry scheduled",
		zap.String("exception_id", exception.ID),
		zap.Int("retry_count", exception.AutoRetryCount),
		zap.Time("next_retry_at", nextRetryAt),
	)

	return nil
}

// sendExceptionAlert 发送异常告警
func (cseh *ConfigSyncExceptionHandler) sendExceptionAlert(ctx context.Context, exception *SyncException) error {
	// 创建告警
	alert := &model.Alert{
		Title:       fmt.Sprintf("配置同步异常: %s", exception.ExceptionType),
		Content:     fmt.Sprintf("集群 %s 的 %s 配置同步出现异常: %s", exception.ClusterID, exception.ConfigType, exception.ErrorMessage),
		Severity:    exception.Severity,
		Status:      "firing",
		Source:      "config_sync_monitor",
		Labels:      fmt.Sprintf(`{"cluster_id":"%s","config_type":"%s","exception_type":"%s"}`, exception.ClusterID, exception.ConfigType, exception.ExceptionType),
		// Annotations field doesn't exist, using Labels instead
		// Annotations: fmt.Sprintf(`{"exception_id":"%s","first_occurred":"%s"}`, exception.ID, exception.FirstOccurred.Format(time.RFC3339)),
	}

	if err := cseh.alertService.CreateAlert(ctx, alert); err != nil {
		return fmt.Errorf("failed to create alert: %w", err)
	}

	logger.L.Info("Exception alert sent",
		zap.String("exception_id", exception.ID),
		zap.Uint("alert_id", alert.ID),
	)

	return nil
}

// AnalyzeException 分析异常根因
func (cseh *ConfigSyncExceptionHandler) AnalyzeException(ctx context.Context, exceptionID string) (*ExceptionAnalysis, error) {
	// 获取异常详情
	var exception SyncException
	err := database.DB.WithContext(ctx).Where("id = ?", exceptionID).First(&exception).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get exception: %w", err)
	}

	// 获取相关的历史记录
	var histories []model.ConfigSyncHistory
	err = database.DB.WithContext(ctx).
		Where("cluster_id = ? AND config_type = ? AND created_at >= ?",
			exception.ClusterID, exception.ConfigType, exception.FirstOccurred.Add(-1*time.Hour)).
		Order("created_at DESC").
		Limit(50).
		Find(&histories).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get sync history: %w", err)
	}

	// 分析根因
	analysis := &ExceptionAnalysis{
		ExceptionID:  exceptionID,
		AnalysisTime: time.Now(),
		Metadata:     make(map[string]interface{}),
	}

	// 基于异常类型和历史记录进行分析
	switch exception.ExceptionType {
	case "timeout":
		analysis.RootCause = "网络延迟或目标系统响应缓慢"
		analysis.PossibleCauses = []string{
			"网络连接不稳定",
			"目标系统负载过高",
			"配置文件过大",
			"DNS解析问题",
		}
		analysis.SuggestedActions = []string{
			"检查网络连接状态",
			"增加超时时间配置",
			"检查目标系统资源使用情况",
			"考虑分批同步大配置文件",
		}
	case "connection_error":
		analysis.RootCause = "无法建立与目标系统的连接"
		analysis.PossibleCauses = []string{
			"目标系统服务未启动",
			"防火墙阻止连接",
			"网络路由问题",
			"端口配置错误",
		}
		analysis.SuggestedActions = []string{
			"检查目标系统服务状态",
			"验证网络连通性",
			"检查防火墙规则",
			"确认端口配置正确",
		}
	case "validation_error":
		analysis.RootCause = "配置内容格式或语法错误"
		analysis.PossibleCauses = []string{
			"配置模板错误",
			"数据源数据异常",
			"配置生成逻辑错误",
		}
		analysis.SuggestedActions = []string{
			"检查配置模板语法",
			"验证数据源数据完整性",
			"测试配置生成逻辑",
			"启用配置验证功能",
		}
	default:
		analysis.RootCause = "未知原因导致的同步失败"
		analysis.PossibleCauses = []string{
			"系统内部错误",
			"资源不足",
			"权限问题",
		}
		analysis.SuggestedActions = []string{
			"查看详细错误日志",
			"检查系统资源使用情况",
			"验证权限配置",
		}
	}

	// 计算置信度
	analysis.Confidence = cseh.calculateConfidence(exception, histories)

	// 查找相关异常
	var relatedExceptions []SyncException
	err = database.DB.WithContext(ctx).
		Where("cluster_id = ? AND exception_type = ? AND id != ? AND status != 'resolved'",
			exception.ClusterID, exception.ExceptionType, exceptionID).
		Find(&relatedExceptions).Error
	if err == nil {
		for _, related := range relatedExceptions {
			analysis.RelatedExceptions = append(analysis.RelatedExceptions, related.ID)
		}
	}

	// 添加统计信息到元数据
	analysis.Metadata["occurrence_count"] = exception.OccurrenceCount
	analysis.Metadata["duration"] = time.Since(exception.FirstOccurred).String()
	analysis.Metadata["auto_retry_count"] = exception.AutoRetryCount
	analysis.Metadata["history_count"] = len(histories)

	return analysis, nil
}

// calculateConfidence 计算分析置信度
func (cseh *ConfigSyncExceptionHandler) calculateConfidence(exception SyncException, histories []model.ConfigSyncHistory) float64 {
	confidence := 0.5 // 基础置信度

	// 基于异常类型调整置信度
	switch exception.ExceptionType {
	case "timeout", "connection_error":
		confidence += 0.3 // 网络相关问题比较容易判断
	case "validation_error":
		confidence += 0.2 // 配置错误需要更多分析
	}

	// 基于历史记录调整置信度
	if len(histories) > 10 {
		confidence += 0.1 // 有足够的历史数据
	}

	// 基于发生频率调整置信度
	if exception.OccurrenceCount > 5 {
		confidence += 0.1 // 频繁发生的问题更容易分析
	}

	// 确保置信度在合理范围内
	if confidence > 1.0 {
		confidence = 1.0
	}
	if confidence < 0.1 {
		confidence = 0.1
	}

	return confidence
}

// ResolveException 解决异常
func (cseh *ConfigSyncExceptionHandler) ResolveException(ctx context.Context, exceptionID, resolvedBy, resolution string) error {
	var exception SyncException
	err := database.DB.WithContext(ctx).Where("id = ?", exceptionID).First(&exception).Error
	if err != nil {
		return fmt.Errorf("failed to get exception: %w", err)
	}

	now := time.Now()
	exception.Status = "resolved"
	exception.ResolvedAt = &now
	exception.ResolvedBy = resolvedBy

	if err := database.DB.WithContext(ctx).Save(&exception).Error; err != nil {
		return fmt.Errorf("failed to resolve exception: %w", err)
	}

	logger.L.Info("Exception resolved",
		zap.String("exception_id", exceptionID),
		zap.String("resolved_by", resolvedBy),
	)

	return nil
}

// GetActiveExceptions 获取活跃异常
func (cseh *ConfigSyncExceptionHandler) GetActiveExceptions(ctx context.Context, clusterID, configType string) ([]SyncException, error) {
	query := database.DB.WithContext(ctx).Where("status != 'resolved'")
	
	if clusterID != "" {
		query = query.Where("cluster_id = ?", clusterID)
	}
	if configType != "" {
		query = query.Where("config_type = ?", configType)
	}

	var exceptions []SyncException
	err := query.Order("severity DESC, created_at DESC").Find(&exceptions).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get active exceptions: %w", err)
	}

	return exceptions, nil
}