package service

import (
	"context"
	"crypto/sha256"
	"fmt"
	"strings"
	"time"

	"alert_agent/internal/pkg/database"
	"alert_agent/internal/pkg/logger"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// ConfigVersionManager 配置版本管理器
type ConfigVersionManager struct {
	configService *ConfigService
}

// NewConfigVersionManager 创建配置版本管理器
func NewConfigVersionManager() *ConfigVersionManager {
	return &ConfigVersionManager{
		configService: NewConfigService(),
	}
}

// ConfigVersion 配置版本
type ConfigVersion struct {
	ID           string    `json:"id" gorm:"type:varchar(36);primaryKey"`
	ClusterID    string    `json:"cluster_id" gorm:"type:varchar(100);not null;index"`
	ConfigType   string    `json:"config_type" gorm:"type:varchar(50);not null;index"`
	Version      string    `json:"version" gorm:"type:varchar(50);not null"`
	ConfigHash   string    `json:"config_hash" gorm:"type:varchar(64);not null;index"`
	ConfigContent string   `json:"config_content" gorm:"type:longtext"`
	ConfigSize   int64     `json:"config_size" gorm:"not null;default:0"`
	Description  string    `json:"description" gorm:"type:text"`
	CreatedBy    string    `json:"created_by" gorm:"type:varchar(100)"`
	IsActive     bool      `json:"is_active" gorm:"default:false;index"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// TableName 指定表名
func (ConfigVersion) TableName() string {
	return "config_versions"
}

// BeforeCreate GORM钩子：创建前生成ID
func (cv *ConfigVersion) BeforeCreate(tx *gorm.DB) error {
	if cv.ID == "" {
		cv.ID = generateID()
	}
	return nil
}

// ConfigDiff 配置差异
type ConfigDiff struct {
	ClusterID     string      `json:"cluster_id"`
	ConfigType    string      `json:"config_type"`
	FromVersion   string      `json:"from_version"`
	ToVersion     string      `json:"to_version"`
	DiffType      string      `json:"diff_type"` // added, removed, modified
	Changes       []DiffLine  `json:"changes"`
	Summary       DiffSummary `json:"summary"`
	GeneratedAt   time.Time   `json:"generated_at"`
}

// DiffLine 差异行
type DiffLine struct {
	LineNumber int    `json:"line_number"`
	Type       string `json:"type"` // added, removed, modified, context
	OldContent string `json:"old_content,omitempty"`
	NewContent string `json:"new_content,omitempty"`
}

// DiffSummary 差异摘要
type DiffSummary struct {
	AddedLines    int `json:"added_lines"`
	RemovedLines  int `json:"removed_lines"`
	ModifiedLines int `json:"modified_lines"`
	TotalChanges  int `json:"total_changes"`
}

// ConsistencyCheck 一致性检查结果
type ConsistencyCheck struct {
	ClusterID       string                    `json:"cluster_id"`
	ConfigType      string                    `json:"config_type"`
	ExpectedHash    string                    `json:"expected_hash"`
	ActualHash      string                    `json:"actual_hash"`
	IsConsistent    bool                      `json:"is_consistent"`
	Inconsistencies []InconsistencyDetail     `json:"inconsistencies,omitempty"`
	CheckTime       time.Time                 `json:"check_time"`
	Recommendations []string                  `json:"recommendations,omitempty"`
}

// InconsistencyDetail 不一致详情
type InconsistencyDetail struct {
	Type        string `json:"type"` // hash_mismatch, content_diff, missing_config
	Description string `json:"description"`
	Severity    string `json:"severity"` // low, medium, high, critical
	Impact      string `json:"impact"`
}

// CreateVersion 创建配置版本
func (cvm *ConfigVersionManager) CreateVersion(ctx context.Context, clusterID, configType, description, createdBy string) (*ConfigVersion, error) {
	logger.L.Debug("Creating config version",
		zap.String("cluster_id", clusterID),
		zap.String("config_type", configType),
	)

	// 获取当前配置内容
	configContent, err := cvm.configService.GetConfig(ctx, clusterID, configType)
	if err != nil {
		return nil, fmt.Errorf("failed to get config content: %w", err)
	}

	// 计算配置hash
	hash := sha256.Sum256([]byte(configContent))
	configHash := fmt.Sprintf("%x", hash)

	// 检查是否已存在相同hash的版本
	var existingVersion ConfigVersion
	err = database.DB.WithContext(ctx).
		Where("cluster_id = ? AND config_type = ? AND config_hash = ?", clusterID, configType, configHash).
		First(&existingVersion).Error
	
	if err == nil {
		// 已存在相同版本，直接返回
		logger.L.Info("Config version already exists",
			zap.String("version_id", existingVersion.ID),
			zap.String("config_hash", configHash),
		)
		return &existingVersion, nil
	} else if err != gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("failed to check existing version: %w", err)
	}

	// 生成版本号
	version, err := cvm.generateVersionNumber(ctx, clusterID, configType)
	if err != nil {
		return nil, fmt.Errorf("failed to generate version number: %w", err)
	}

	// 创建新版本
	configVersion := &ConfigVersion{
		ClusterID:     clusterID,
		ConfigType:    configType,
		Version:       version,
		ConfigHash:    configHash,
		ConfigContent: configContent,
		ConfigSize:    int64(len(configContent)),
		Description:   description,
		CreatedBy:     createdBy,
		IsActive:      true, // 新版本默认为活跃状态
	}

	// 开始事务
	tx := database.DB.WithContext(ctx).Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 将之前的活跃版本设为非活跃
	if err := tx.Model(&ConfigVersion{}).
		Where("cluster_id = ? AND config_type = ? AND is_active = ?", clusterID, configType, true).
		Update("is_active", false).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to deactivate previous versions: %w", err)
	}

	// 创建新版本
	if err := tx.Create(configVersion).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to create config version: %w", err)
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	logger.L.Info("Config version created",
		zap.String("version_id", configVersion.ID),
		zap.String("version", version),
		zap.String("config_hash", configHash),
	)

	return configVersion, nil
}

// generateVersionNumber 生成版本号
func (cvm *ConfigVersionManager) generateVersionNumber(ctx context.Context, clusterID, configType string) (string, error) {
	// 获取最新版本号
	var latestVersion ConfigVersion
	err := database.DB.WithContext(ctx).
		Where("cluster_id = ? AND config_type = ?", clusterID, configType).
		Order("created_at DESC").
		First(&latestVersion).Error

	if err == gorm.ErrRecordNotFound {
		// 第一个版本
		return "v1.0.0", nil
	} else if err != nil {
		return "", fmt.Errorf("failed to get latest version: %w", err)
	}

	// 解析版本号并递增
	version := latestVersion.Version
	if strings.HasPrefix(version, "v") {
		version = version[1:]
	}

	parts := strings.Split(version, ".")
	if len(parts) != 3 {
		// 如果版本格式不正确，重新开始
		return "v1.0.0", nil
	}

	// 递增补丁版本号
	var major, minor, patch int
	fmt.Sscanf(version, "%d.%d.%d", &major, &minor, &patch)
	patch++

	return fmt.Sprintf("v%d.%d.%d", major, minor, patch), nil
}

// GetVersions 获取配置版本列表
func (cvm *ConfigVersionManager) GetVersions(ctx context.Context, clusterID, configType string, limit int) ([]ConfigVersion, error) {
	if limit <= 0 {
		limit = 50
	}

	var versions []ConfigVersion
	query := database.DB.WithContext(ctx).
		Where("cluster_id = ? AND config_type = ?", clusterID, configType).
		Order("created_at DESC").
		Limit(limit)

	err := query.Find(&versions).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get config versions: %w", err)
	}

	return versions, nil
}

// GetVersion 获取指定版本
func (cvm *ConfigVersionManager) GetVersion(ctx context.Context, versionID string) (*ConfigVersion, error) {
	var version ConfigVersion
	err := database.DB.WithContext(ctx).Where("id = ?", versionID).First(&version).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get config version: %w", err)
	}

	return &version, nil
}

// CompareVersions 比较两个版本的差异
func (cvm *ConfigVersionManager) CompareVersions(ctx context.Context, fromVersionID, toVersionID string) (*ConfigDiff, error) {
	// 获取两个版本
	fromVersion, err := cvm.GetVersion(ctx, fromVersionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get from version: %w", err)
	}

	toVersion, err := cvm.GetVersion(ctx, toVersionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get to version: %w", err)
	}

	// 检查是否为同一集群和配置类型
	if fromVersion.ClusterID != toVersion.ClusterID || fromVersion.ConfigType != toVersion.ConfigType {
		return nil, fmt.Errorf("versions must be from the same cluster and config type")
	}

	logger.L.Debug("Comparing config versions",
		zap.String("from_version", fromVersion.Version),
		zap.String("to_version", toVersion.Version),
	)

	// 生成差异
	diff := &ConfigDiff{
		ClusterID:   fromVersion.ClusterID,
		ConfigType:  fromVersion.ConfigType,
		FromVersion: fromVersion.Version,
		ToVersion:   toVersion.Version,
		GeneratedAt: time.Now(),
	}

	// 如果hash相同，则无差异
	if fromVersion.ConfigHash == toVersion.ConfigHash {
		diff.DiffType = "no_change"
		diff.Summary = DiffSummary{}
		return diff, nil
	}

	// 计算差异
	changes, summary := cvm.calculateDiff(fromVersion.ConfigContent, toVersion.ConfigContent)
	diff.Changes = changes
	diff.Summary = summary

	if summary.TotalChanges > 0 {
		diff.DiffType = "modified"
	}

	return diff, nil
}

// calculateDiff 计算配置内容差异
func (cvm *ConfigVersionManager) calculateDiff(oldContent, newContent string) ([]DiffLine, DiffSummary) {
	oldLines := strings.Split(oldContent, "\n")
	newLines := strings.Split(newContent, "\n")

	var changes []DiffLine
	summary := DiffSummary{}

	// 简单的行级差异算法
	maxLen := len(oldLines)
	if len(newLines) > maxLen {
		maxLen = len(newLines)
	}

	for i := 0; i < maxLen; i++ {
		var oldLine, newLine string
		
		if i < len(oldLines) {
			oldLine = oldLines[i]
		}
		if i < len(newLines) {
			newLine = newLines[i]
		}

		if i >= len(oldLines) {
			// 新增行
			changes = append(changes, DiffLine{
				LineNumber: i + 1,
				Type:       "added",
				NewContent: newLine,
			})
			summary.AddedLines++
		} else if i >= len(newLines) {
			// 删除行
			changes = append(changes, DiffLine{
				LineNumber: i + 1,
				Type:       "removed",
				OldContent: oldLine,
			})
			summary.RemovedLines++
		} else if oldLine != newLine {
			// 修改行
			changes = append(changes, DiffLine{
				LineNumber: i + 1,
				Type:       "modified",
				OldContent: oldLine,
				NewContent: newLine,
			})
			summary.ModifiedLines++
		}
	}

	summary.TotalChanges = summary.AddedLines + summary.RemovedLines + summary.ModifiedLines

	return changes, summary
}

// RollbackToVersion 回滚到指定版本
func (cvm *ConfigVersionManager) RollbackToVersion(ctx context.Context, versionID, rollbackBy string) error {
	// 获取目标版本
	targetVersion, err := cvm.GetVersion(ctx, versionID)
	if err != nil {
		return fmt.Errorf("failed to get target version: %w", err)
	}

	logger.L.Info("Rolling back to version",
		zap.String("version_id", versionID),
		zap.String("version", targetVersion.Version),
		zap.String("rollback_by", rollbackBy),
	)

	// 开始事务
	tx := database.DB.WithContext(ctx).Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 将当前活跃版本设为非活跃
	if err := tx.Model(&ConfigVersion{}).
		Where("cluster_id = ? AND config_type = ? AND is_active = ?", 
			targetVersion.ClusterID, targetVersion.ConfigType, true).
		Update("is_active", false).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to deactivate current version: %w", err)
	}

	// 将目标版本设为活跃
	if err := tx.Model(&ConfigVersion{}).
		Where("id = ?", versionID).
		Update("is_active", true).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to activate target version: %w", err)
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit rollback transaction: %w", err)
	}

	// 触发配置同步
	if err := cvm.configService.TriggerSync(ctx, targetVersion.ClusterID, targetVersion.ConfigType); err != nil {
		logger.L.Error("Failed to trigger sync after rollback", zap.Error(err))
		// 不返回错误，因为回滚已经成功
	}

	logger.L.Info("Version rollback completed",
		zap.String("version_id", versionID),
		zap.String("rollback_by", rollbackBy),
	)

	return nil
}

// CheckConsistency 检查配置一致性
func (cvm *ConfigVersionManager) CheckConsistency(ctx context.Context, clusterID, configType string) (*ConsistencyCheck, error) {
	logger.L.Debug("Checking config consistency",
		zap.String("cluster_id", clusterID),
		zap.String("config_type", configType),
	)

	// 获取当前活跃版本
	var activeVersion ConfigVersion
	err := database.DB.WithContext(ctx).
		Where("cluster_id = ? AND config_type = ? AND is_active = ?", clusterID, configType, true).
		First(&activeVersion).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get active version: %w", err)
	}

	// 获取当前实际配置
	actualConfig, err := cvm.configService.GetConfig(ctx, clusterID, configType)
	if err != nil {
		return nil, fmt.Errorf("failed to get actual config: %w", err)
	}

	// 计算实际配置hash
	actualHash := fmt.Sprintf("%x", sha256.Sum256([]byte(actualConfig)))

	check := &ConsistencyCheck{
		ClusterID:    clusterID,
		ConfigType:   configType,
		ExpectedHash: activeVersion.ConfigHash,
		ActualHash:   actualHash,
		IsConsistent: activeVersion.ConfigHash == actualHash,
		CheckTime:    time.Now(),
	}

	if !check.IsConsistent {
		// 分析不一致的原因
		inconsistency := InconsistencyDetail{
			Type:        "hash_mismatch",
			Description: "配置内容与活跃版本不匹配",
			Severity:    "high",
			Impact:      "可能导致配置同步异常或系统行为不一致",
		}
		check.Inconsistencies = append(check.Inconsistencies, inconsistency)

		// 生成修复建议
		check.Recommendations = []string{
			"检查配置生成逻辑是否正确",
			"验证数据源数据完整性",
			"考虑重新生成配置版本",
			"检查是否有手动修改配置的情况",
		}

		logger.L.Warn("Config consistency check failed",
			zap.String("cluster_id", clusterID),
			zap.String("config_type", configType),
			zap.String("expected_hash", activeVersion.ConfigHash),
			zap.String("actual_hash", actualHash),
		)
	} else {
		logger.L.Debug("Config consistency check passed",
			zap.String("cluster_id", clusterID),
			zap.String("config_type", configType),
		)
	}

	return check, nil
}

// GetActiveVersion 获取活跃版本
func (cvm *ConfigVersionManager) GetActiveVersion(ctx context.Context, clusterID, configType string) (*ConfigVersion, error) {
	var version ConfigVersion
	err := database.DB.WithContext(ctx).
		Where("cluster_id = ? AND config_type = ? AND is_active = ?", clusterID, configType, true).
		First(&version).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get active version: %w", err)
	}

	return &version, nil
}

// DeleteVersion 删除版本（软删除）
func (cvm *ConfigVersionManager) DeleteVersion(ctx context.Context, versionID, deletedBy string) error {
	// 检查是否为活跃版本
	var version ConfigVersion
	err := database.DB.WithContext(ctx).Where("id = ?", versionID).First(&version).Error
	if err != nil {
		return fmt.Errorf("failed to get version: %w", err)
	}

	if version.IsActive {
		return fmt.Errorf("cannot delete active version")
	}

	// 软删除版本
	if err := database.DB.WithContext(ctx).Delete(&version).Error; err != nil {
		return fmt.Errorf("failed to delete version: %w", err)
	}

	logger.L.Info("Config version deleted",
		zap.String("version_id", versionID),
		zap.String("deleted_by", deletedBy),
	)

	return nil
}

// CleanupOldVersions 清理旧版本
func (cvm *ConfigVersionManager) CleanupOldVersions(ctx context.Context, clusterID, configType string, keepCount int) error {
	if keepCount <= 0 {
		keepCount = 10 // 默认保留10个版本
	}

	// 获取版本列表（按创建时间倒序）
	var versions []ConfigVersion
	err := database.DB.WithContext(ctx).
		Where("cluster_id = ? AND config_type = ?", clusterID, configType).
		Order("created_at DESC").
		Find(&versions).Error
	if err != nil {
		return fmt.Errorf("failed to get versions: %w", err)
	}

	if len(versions) <= keepCount {
		return nil // 版本数量未超过保留数量
	}

	// 删除多余的版本（保留活跃版本）
	var toDelete []string
	activeVersionFound := false
	keptCount := 0

	for _, version := range versions {
		if version.IsActive {
			activeVersionFound = true
			continue
		}

		if keptCount < keepCount {
			keptCount++
			continue
		}

		toDelete = append(toDelete, version.ID)
	}

	// 如果没有找到活跃版本，保留最新的版本
	if !activeVersionFound && len(versions) > 0 && len(toDelete) > 0 {
		// 从删除列表中移除最新的版本
		toDelete = toDelete[1:]
	}

	// 执行删除
	if len(toDelete) > 0 {
		result := database.DB.WithContext(ctx).
			Where("id IN ?", toDelete).
			Delete(&ConfigVersion{})
		
		if result.Error != nil {
			return fmt.Errorf("failed to cleanup old versions: %w", result.Error)
		}

		logger.L.Info("Old config versions cleaned up",
			zap.String("cluster_id", clusterID),
			zap.String("config_type", configType),
			zap.Int64("deleted_count", result.RowsAffected),
		)
	}

	return nil
}