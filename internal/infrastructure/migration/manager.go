package migration

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// Manager 迁移管理器
type Manager struct {
	migrator *Migrator
	logger   *zap.Logger
}

// NewManager 创建迁移管理器
func NewManager(db *gorm.DB, logger *zap.Logger) *Manager {
	if logger == nil {
		logger = zap.NewNop()
	}

	migrator := NewMigrator(db, logger)
	
	// 添加所有迁移步骤
	for _, step := range GetAllMigrationSteps() {
		migrator.AddStep(step)
	}

	return &Manager{
		migrator: migrator,
		logger:   logger,
	}
}

// MigrateToLatest 迁移到最新版本
func (m *Manager) MigrateToLatest(ctx context.Context) error {
	m.logger.Info("开始数据库迁移到最新版本")
	start := time.Now()

	if err := m.migrator.Migrate(ctx); err != nil {
		m.logger.Error("数据库迁移失败", zap.Error(err))
		return fmt.Errorf("迁移到最新版本失败: %w", err)
	}

	duration := time.Since(start)
	m.logger.Info("数据库迁移完成", zap.Duration("duration", duration))
	return nil
}

// MigrateToVersion 迁移到指定版本 (注意：当前实现迁移所有待执行步骤，不支持指定版本)
func (m *Manager) MigrateToVersion(ctx context.Context, version string) error {
	m.logger.Info("开始数据库迁移", zap.String("target_version", version))
	start := time.Now()

	// 当前实现会执行所有待执行的迁移步骤
	if err := m.migrator.Migrate(ctx); err != nil {
		m.logger.Error("数据库迁移失败", 
			zap.String("target_version", version),
			zap.Error(err))
		return fmt.Errorf("迁移失败: %w", err)
	}

	duration := time.Since(start)
	m.logger.Info("数据库迁移完成", 
		zap.String("target_version", version),
		zap.Duration("duration", duration))
	return nil
}

// RollbackToVersion 回滚到指定版本
func (m *Manager) RollbackToVersion(ctx context.Context, version string) error {
	m.logger.Info("开始数据库回滚到指定版本", zap.String("version", version))
	start := time.Now()

	if err := m.migrator.Rollback(ctx, version); err != nil {
		m.logger.Error("数据库回滚失败", 
			zap.String("version", version),
			zap.Error(err))
		return fmt.Errorf("回滚到版本 %s 失败: %w", version, err)
	}

	duration := time.Since(start)
	m.logger.Info("数据库回滚完成", 
		zap.String("version", version),
		zap.Duration("duration", duration))
	return nil
}

// GetMigrationStatus 获取迁移状态
func (m *Manager) GetMigrationStatus() ([]Migration, error) {
	m.logger.Debug("获取数据库迁移状态")

	var migrations []Migration
	if err := m.migrator.db.Order("executed_at DESC").Find(&migrations).Error; err != nil {
		m.logger.Error("获取迁移状态失败", zap.Error(err))
		return nil, fmt.Errorf("获取迁移状态失败: %w", err)
	}

	m.logger.Debug("获取迁移状态成功", zap.Int("count", len(migrations)))
	return migrations, nil
}

// GetPendingMigrations 获取待执行的迁移
func (m *Manager) GetPendingMigrations() ([]MigrationStep, error) {
	m.logger.Debug("获取待执行的迁移")

	// 获取已执行的迁移
	var executedVersions []string
	if err := m.migrator.db.Model(&Migration{}).Where("status = ?", "success").Pluck("version", &executedVersions).Error; err != nil {
		m.logger.Error("获取已执行迁移失败", zap.Error(err))
		return nil, fmt.Errorf("获取已执行迁移失败: %w", err)
	}

	// 创建已执行版本的映射
	executedMap := make(map[string]bool)
	for _, version := range executedVersions {
		executedMap[version] = true
	}

	// 筛选待执行的迁移
	var pending []MigrationStep
	for _, step := range m.migrator.steps {
		if !executedMap[step.Version] {
			pending = append(pending, step)
		}
	}

	m.logger.Debug("获取待执行迁移成功", zap.Int("count", len(pending)))
	return pending, nil
}

// ValidateDatabase 验证数据库状态
func (m *Manager) ValidateDatabase() error {
	m.logger.Info("开始验证数据库状态")

	// 检查迁移表是否存在
	if err := m.migrator.ensureMigrationTable(); err != nil {
		m.logger.Error("迁移表检查失败", zap.Error(err))
		return fmt.Errorf("迁移表检查失败: %w", err)
	}

	// 检查是否有失败的迁移
	var failedCount int64
	if err := m.migrator.db.Model(&Migration{}).Where("status = ?", "failed").Count(&failedCount).Error; err != nil {
		m.logger.Error("检查失败迁移失败", zap.Error(err))
		return fmt.Errorf("检查失败迁移失败: %w", err)
	}

	if failedCount > 0 {
		m.logger.Warn("发现失败的迁移", zap.Int64("count", failedCount))
		return fmt.Errorf("发现 %d 个失败的迁移，请检查并修复", failedCount)
	}

	// 验证关键表是否存在
	requiredTables := []string{
		"alert_channels",
		"channel_groups",
		"alertmanager_clusters",
		"alert_processing_records",
		"ai_analysis_records",
	}

	for _, table := range requiredTables {
		if !m.migrator.db.Migrator().HasTable(table) {
			m.logger.Error("缺少必需的表", zap.String("table", table))
			return fmt.Errorf("缺少必需的表: %s", table)
		}
	}

	m.logger.Info("数据库状态验证通过")
	return nil
}

// GetCurrentVersion 获取当前数据库版本
func (m *Manager) GetCurrentVersion() (string, error) {
	m.logger.Debug("获取当前数据库版本")

	var migration Migration
	if err := m.migrator.db.Where("status = ?", "success").Order("executed_at DESC").First(&migration).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			m.logger.Debug("未找到已执行的迁移，数据库可能是全新的")
			return "", nil
		}
		m.logger.Error("获取当前版本失败", zap.Error(err))
		return "", fmt.Errorf("获取当前版本失败: %w", err)
	}

	m.logger.Debug("获取当前版本成功", zap.String("version", migration.Version))
	return migration.Version, nil
}

// RepairFailedMigration 修复失败的迁移
func (m *Manager) RepairFailedMigration(version string) error {
	m.logger.Info("开始修复失败的迁移", zap.String("version", version))

	// 查找失败的迁移记录
	var migration Migration
	if err := m.migrator.db.Where("version = ? AND status = ?", version, "failed").First(&migration).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			m.logger.Warn("未找到失败的迁移记录", zap.String("version", version))
			return fmt.Errorf("未找到版本 %s 的失败迁移记录", version)
		}
		m.logger.Error("查找失败迁移记录失败", zap.Error(err))
		return fmt.Errorf("查找失败迁移记录失败: %w", err)
	}

	// 将状态重置为待执行
	if err := m.migrator.db.Model(&migration).Updates(map[string]interface{}{
		"status":       "pending",
		"error_message": nil,
		"updated_at":   time.Now(),
	}).Error; err != nil {
		m.logger.Error("重置迁移状态失败", zap.Error(err))
		return fmt.Errorf("重置迁移状态失败: %w", err)
	}

	m.logger.Info("失败迁移修复完成", zap.String("version", version))
	return nil
}

// CleanupMigrationHistory 清理迁移历史
func (m *Manager) CleanupMigrationHistory(keepDays int) error {
	m.logger.Info("开始清理迁移历史", zap.Int("keep_days", keepDays))

	cutoffTime := time.Now().AddDate(0, 0, -keepDays)

	result := m.migrator.db.Where("executed_at < ? AND status = ?", cutoffTime, "success").Delete(&Migration{})
	if result.Error != nil {
		m.logger.Error("清理迁移历史失败", zap.Error(result.Error))
		return fmt.Errorf("清理迁移历史失败: %w", result.Error)
	}

	m.logger.Info("迁移历史清理完成", 
		zap.Int64("deleted_count", result.RowsAffected),
		zap.Int("keep_days", keepDays))
	return nil
}

// ExportMigrationHistory 导出迁移历史
func (m *Manager) ExportMigrationHistory() ([]Migration, error) {
	m.logger.Debug("导出迁移历史")

	var migrations []Migration
	if err := m.migrator.db.Order("executed_at ASC").Find(&migrations).Error; err != nil {
		m.logger.Error("导出迁移历史失败", zap.Error(err))
		return nil, fmt.Errorf("导出迁移历史失败: %w", err)
	}

	m.logger.Debug("迁移历史导出成功", zap.Int("count", len(migrations)))
	return migrations, nil
}

// GetMigrationInfo 获取迁移信息
func (m *Manager) GetMigrationInfo() (*MigrationInfo, error) {
	m.logger.Debug("获取迁移信息")

	currentVersion, err := m.GetCurrentVersion()
	if err != nil {
		return nil, err
	}

	pendingMigrations, err := m.GetPendingMigrations()
	if err != nil {
		return nil, err
	}

	migrationHistory, err := m.GetMigrationStatus()
	if err != nil {
		return nil, err
	}

	info := &MigrationInfo{
		CurrentVersion:    currentVersion,
		PendingCount:      len(pendingMigrations),
		TotalMigrations:   len(m.migrator.steps),
		ExecutedCount:     len(migrationHistory),
		PendingMigrations: pendingMigrations,
		MigrationHistory:  migrationHistory,
	}

	m.logger.Debug("获取迁移信息成功", 
		zap.String("current_version", currentVersion),
		zap.Int("pending_count", info.PendingCount),
		zap.Int("total_migrations", info.TotalMigrations))

	return info, nil
}

// MigrationInfo 迁移信息
type MigrationInfo struct {
	CurrentVersion    string           `json:"current_version"`
	PendingCount      int              `json:"pending_count"`
	TotalMigrations   int              `json:"total_migrations"`
	ExecutedCount     int              `json:"executed_count"`
	PendingMigrations []MigrationStep  `json:"pending_migrations"`
	MigrationHistory  []Migration      `json:"migration_history"`
}

// IsUpToDate 检查数据库是否为最新版本
func (m *Manager) IsUpToDate() (bool, error) {
	m.logger.Debug("检查数据库是否为最新版本")

	pendingMigrations, err := m.GetPendingMigrations()
	if err != nil {
		return false, err
	}

	isUpToDate := len(pendingMigrations) == 0
	m.logger.Debug("数据库版本检查完成", zap.Bool("is_up_to_date", isUpToDate))

	return isUpToDate, nil
}