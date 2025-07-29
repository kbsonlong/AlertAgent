package migration

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"
	"go.uber.org/zap"
)

// Migration 数据库迁移记录
type Migration struct {
	ID          string    `json:"id" gorm:"primaryKey;type:varchar(36)"`
	Version     string    `json:"version" gorm:"type:varchar(50);not null;uniqueIndex"`
	Name        string    `json:"name" gorm:"type:varchar(255);not null"`
	Description string    `json:"description" gorm:"type:text"`
	ExecutedAt  time.Time `json:"executed_at" gorm:"autoCreateTime"`
	Checksum    string    `json:"checksum" gorm:"type:varchar(64);not null"`
	Success     bool      `json:"success" gorm:"default:false"`
	ErrorMsg    string    `json:"error_msg" gorm:"type:text"`
	Duration    int64     `json:"duration"` // 执行时间（毫秒）
}

// MigrationFunc 迁移函数类型
type MigrationFunc func(db *gorm.DB) error

// MigrationStep 迁移步骤
type MigrationStep struct {
	Version     string
	Name        string
	Description string
	Up          MigrationFunc
	Down        MigrationFunc
	Checksum    string
}

// Migrator 数据库迁移器
type Migrator struct {
	db     *gorm.DB
	logger *zap.Logger
	steps  []MigrationStep
}

// NewMigrator 创建迁移器
func NewMigrator(db *gorm.DB, logger *zap.Logger) *Migrator {
	return &Migrator{
		db:     db,
		logger: logger,
		steps:  make([]MigrationStep, 0),
	}
}

// AddStep 添加迁移步骤
func (m *Migrator) AddStep(step MigrationStep) {
	m.steps = append(m.steps, step)
}

// Migrate 执行迁移
func (m *Migrator) Migrate(ctx context.Context) error {
	// 确保迁移表存在
	if err := m.ensureMigrationTable(); err != nil {
		return fmt.Errorf("failed to ensure migration table: %w", err)
	}

	// 获取已执行的迁移
	executedMigrations, err := m.getExecutedMigrations()
	if err != nil {
		return fmt.Errorf("failed to get executed migrations: %w", err)
	}

	// 执行待迁移步骤
	for _, step := range m.steps {
		if _, exists := executedMigrations[step.Version]; exists {
			m.logger.Info("Migration already executed", zap.String("version", step.Version))
			continue
		}

		if err := m.executeStep(ctx, step); err != nil {
			return fmt.Errorf("failed to execute migration %s: %w", step.Version, err)
		}
	}

	return nil
}

// Rollback 回滚迁移
func (m *Migrator) Rollback(ctx context.Context, targetVersion string) error {
	// 获取已执行的迁移
	executedMigrations, err := m.getExecutedMigrations()
	if err != nil {
		return fmt.Errorf("failed to get executed migrations: %w", err)
	}

	// 找到需要回滚的迁移
	var rollbackSteps []MigrationStep
	for i := len(m.steps) - 1; i >= 0; i-- {
		step := m.steps[i]
		if step.Version == targetVersion {
			break
		}
		if _, exists := executedMigrations[step.Version]; exists {
			rollbackSteps = append(rollbackSteps, step)
		}
	}

	// 执行回滚
	for _, step := range rollbackSteps {
		if err := m.rollbackStep(ctx, step); err != nil {
			return fmt.Errorf("failed to rollback migration %s: %w", step.Version, err)
		}
	}

	return nil
}

// ensureMigrationTable 确保迁移表存在
func (m *Migrator) ensureMigrationTable() error {
	return m.db.AutoMigrate(&Migration{})
}

// getExecutedMigrations 获取已执行的迁移
func (m *Migrator) getExecutedMigrations() (map[string]*Migration, error) {
	var migrations []Migration
	if err := m.db.Where("success = ?", true).Find(&migrations).Error; err != nil {
		return nil, err
	}

	executed := make(map[string]*Migration)
	for i := range migrations {
		executed[migrations[i].Version] = &migrations[i]
	}

	return executed, nil
}

// executeStep 执行迁移步骤
func (m *Migrator) executeStep(ctx context.Context, step MigrationStep) error {
	m.logger.Info("Executing migration", 
		zap.String("version", step.Version),
		zap.String("name", step.Name))

	start := time.Now()
	migration := &Migration{
		ID:          generateID(),
		Version:     step.Version,
		Name:        step.Name,
		Description: step.Description,
		Checksum:    step.Checksum,
		Success:     false,
	}

	// 开始事务
	tx := m.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			migration.ErrorMsg = fmt.Sprintf("panic: %v", r)
			m.saveMigrationRecord(migration)
		}
	}()

	// 执行迁移
	if err := step.Up(tx); err != nil {
		tx.Rollback()
		migration.ErrorMsg = err.Error()
		migration.Duration = time.Since(start).Milliseconds()
		m.saveMigrationRecord(migration)
		return err
	}

	// 记录迁移成功
	migration.Success = true
	migration.Duration = time.Since(start).Milliseconds()
	if err := tx.Create(migration).Error; err != nil {
		tx.Rollback()
		return err
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		return err
	}

	m.logger.Info("Migration executed successfully",
		zap.String("version", step.Version),
		zap.Int64("duration_ms", migration.Duration))

	return nil
}

// rollbackStep 回滚迁移步骤
func (m *Migrator) rollbackStep(ctx context.Context, step MigrationStep) error {
	m.logger.Info("Rolling back migration",
		zap.String("version", step.Version),
		zap.String("name", step.Name))

	if step.Down == nil {
		return fmt.Errorf("no rollback function defined for migration %s", step.Version)
	}

	// 开始事务
	tx := m.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 执行回滚
	if err := step.Down(tx); err != nil {
		tx.Rollback()
		return err
	}

	// 删除迁移记录
	if err := tx.Where("version = ? AND success = ?", step.Version, true).Delete(&Migration{}).Error; err != nil {
		tx.Rollback()
		return err
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		return err
	}

	m.logger.Info("Migration rolled back successfully", zap.String("version", step.Version))
	return nil
}

// saveMigrationRecord 保存迁移记录（用于失败情况）
func (m *Migrator) saveMigrationRecord(migration *Migration) {
	if err := m.db.Create(migration).Error; err != nil {
		m.logger.Error("Failed to save migration record", zap.Error(err))
	}
}

// generateID 生成唯一ID
func generateID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

// GetMigrationStatus 获取迁移状态
func (m *Migrator) GetMigrationStatus() ([]Migration, error) {
	var migrations []Migration
	err := m.db.Order("executed_at DESC").Find(&migrations).Error
	return migrations, err
}

// ValidateMigrations 验证迁移完整性
func (m *Migrator) ValidateMigrations() error {
	executedMigrations, err := m.getExecutedMigrations()
	if err != nil {
		return err
	}

	for _, step := range m.steps {
		if migration, exists := executedMigrations[step.Version]; exists {
			if migration.Checksum != step.Checksum {
				return fmt.Errorf("checksum mismatch for migration %s", step.Version)
			}
		}
	}

	return nil
}