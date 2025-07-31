package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/google/uuid"
)

// 迁移配置
type MigrationConfig struct {
	DSN        string
	DryRun     bool
	BatchSize  int
	Verbose    bool
	SkipBackup bool
}

// 迁移统计
type MigrationStats struct {
	RulesProcessed    int
	RulesMigrated     int
	AlertsProcessed   int
	AlertsUpdated     int
	ConfigsCreated    int
	ErrorsEncountered int
	StartTime         time.Time
	EndTime           time.Time
}

// 旧的规则结构
type OldRule struct {
	ID            uint   `gorm:"primarykey"`
	Name          string `gorm:"size:255;not null"`
	Description   string `gorm:"type:text;not null"`
	Level         string `gorm:"size:50;not null"`
	Enabled       bool   `gorm:"not null;default:true"`
	ConditionExpr string `gorm:"type:text;not null"`
	NotifyType    string `gorm:"size:50;not null"`
	NotifyGroup   string `gorm:"size:255;not null"`
	Template      string `gorm:"size:255;not null"`
	CreatedAt     time.Time
	UpdatedAt     time.Time
	DeletedAt     *time.Time `gorm:"index"`
}

func (OldRule) TableName() string {
	return "rules"
}

// 新的规则结构
type NewRule struct {
	ID          string `gorm:"type:varchar(36);primaryKey"`
	Name        string `gorm:"type:varchar(255);not null"`
	Expression  string `gorm:"type:text;not null"`
	Duration    string `gorm:"type:varchar(50);not null"`
	Severity    string `gorm:"type:varchar(20);not null"`
	Labels      string `gorm:"type:json"`
	Annotations string `gorm:"type:json"`
	Targets     string `gorm:"type:json"`
	Version     string `gorm:"type:varchar(50);not null;default:'v1.0.0'"`
	Status      string `gorm:"type:varchar(20);not null;default:'pending'"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   *time.Time `gorm:"index"`
}

func (NewRule) TableName() string {
	return "alert_rules"
}

// 旧的告警结构
type OldAlert struct {
	ID        uint   `gorm:"primarykey"`
	Name      string `gorm:"size:255;not null"`
	Title     string `gorm:"size:255;not null"`
	Level     string `gorm:"type:varchar(50);not null"`
	Status    string `gorm:"type:varchar(50);not null;default:'new'"`
	Source    string `gorm:"type:varchar(255);not null"`
	Content   string `gorm:"type:text;not null"`
	Labels    string `gorm:"type:text"`
	Analysis  string `gorm:"type:text"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time `gorm:"index"`
}

func (OldAlert) TableName() string {
	return "alerts"
}

// 数据迁移器
type DataMigrator struct {
	db     *gorm.DB
	config *MigrationConfig
	stats  *MigrationStats
}

func NewDataMigrator(config *MigrationConfig) (*DataMigrator, error) {
	// 配置日志级别
	logLevel := logger.Silent
	if config.Verbose {
		logLevel = logger.Info
	}

	db, err := gorm.Open(mysql.Open(config.DSN), &gorm.Config{
		Logger: logger.Default.LogMode(logLevel),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	return &DataMigrator{
		db:     db,
		config: config,
		stats: &MigrationStats{
			StartTime: time.Now(),
		},
	}, nil
}

// 执行迁移
func (m *DataMigrator) Migrate() error {
	log.Println("Starting data migration...")

	if !m.config.SkipBackup {
		if err := m.createBackup(); err != nil {
			return fmt.Errorf("failed to create backup: %w", err)
		}
	}

	// 迁移规则数据
	if err := m.migrateRules(); err != nil {
		return fmt.Errorf("failed to migrate rules: %w", err)
	}

	// 更新告警数据
	if err := m.updateAlerts(); err != nil {
		return fmt.Errorf("failed to update alerts: %w", err)
	}

	// 创建配置同步状态
	if err := m.createConfigSyncStatus(); err != nil {
		return fmt.Errorf("failed to create config sync status: %w", err)
	}

	// 验证迁移结果
	if err := m.validateMigration(); err != nil {
		return fmt.Errorf("migration validation failed: %w", err)
	}

	m.stats.EndTime = time.Now()
	m.printStats()

	log.Println("Data migration completed successfully!")
	return nil
}

// 创建备份
func (m *DataMigrator) createBackup() error {
	log.Println("Creating backup tables...")

	if m.config.DryRun {
		log.Println("[DRY RUN] Would create backup tables")
		return nil
	}

	// 创建备份表
	backupTables := []string{
		"CREATE TABLE IF NOT EXISTS rules_backup AS SELECT * FROM rules",
		"CREATE TABLE IF NOT EXISTS alerts_backup AS SELECT * FROM alerts",
		"CREATE TABLE IF NOT EXISTS providers_backup AS SELECT * FROM providers",
	}

	for _, sql := range backupTables {
		if err := m.db.Exec(sql).Error; err != nil {
			return fmt.Errorf("failed to create backup: %w", err)
		}
	}

	log.Println("Backup tables created successfully")
	return nil
}

// 迁移规则数据
func (m *DataMigrator) migrateRules() error {
	log.Println("Migrating rules data...")

	var oldRules []OldRule
	if err := m.db.Where("deleted_at IS NULL").Find(&oldRules).Error; err != nil {
		return fmt.Errorf("failed to fetch old rules: %w", err)
	}

	m.stats.RulesProcessed = len(oldRules)
	log.Printf("Found %d rules to migrate", len(oldRules))

	for i, oldRule := range oldRules {
		if m.config.Verbose {
			log.Printf("Migrating rule %d/%d: %s", i+1, len(oldRules), oldRule.Name)
		}

		newRule, err := m.convertRule(&oldRule)
		if err != nil {
			log.Printf("Failed to convert rule %s: %v", oldRule.Name, err)
			m.stats.ErrorsEncountered++
			continue
		}

		if m.config.DryRun {
			log.Printf("[DRY RUN] Would migrate rule: %s", newRule.Name)
			continue
		}

		// 检查是否已存在
		var existing NewRule
		if err := m.db.Where("name = ?", newRule.Name).First(&existing).Error; err == nil {
			// 更新现有记录
			if err := m.db.Model(&existing).Updates(newRule).Error; err != nil {
				log.Printf("Failed to update existing rule %s: %v", newRule.Name, err)
				m.stats.ErrorsEncountered++
				continue
			}
		} else {
			// 创建新记录
			if err := m.db.Create(newRule).Error; err != nil {
				log.Printf("Failed to create new rule %s: %v", newRule.Name, err)
				m.stats.ErrorsEncountered++
				continue
			}
		}

		m.stats.RulesMigrated++
	}

	log.Printf("Rules migration completed: %d processed, %d migrated, %d errors",
		m.stats.RulesProcessed, m.stats.RulesMigrated, m.stats.ErrorsEncountered)
	return nil
}

// 转换规则格式
func (m *DataMigrator) convertRule(oldRule *OldRule) (*NewRule, error) {
	// 转换严重程度
	severity := "medium"
	switch oldRule.Level {
	case "critical":
		severity = "critical"
	case "warning":
		severity = "warning"
	case "high":
		severity = "high"
	case "medium":
		severity = "medium"
	case "low":
		severity = "low"
	}

	// 创建注释
	annotations := map[string]string{
		"description":  oldRule.Description,
		"notify_type":  oldRule.NotifyType,
		"notify_group": oldRule.NotifyGroup,
		"template":     oldRule.Template,
	}
	annotationsJSON, _ := json.Marshal(annotations)

	// 创建目标列表
	targets := []string{"prometheus"}
	targetsJSON, _ := json.Marshal(targets)

	// 创建空标签
	labelsJSON, _ := json.Marshal(map[string]string{})

	status := "inactive"
	if oldRule.Enabled {
		status = "active"
	}

	return &NewRule{
		ID:          uuid.New().String(),
		Name:        oldRule.Name,
		Expression:  oldRule.ConditionExpr,
		Duration:    "5m", // 默认持续时间
		Severity:    severity,
		Labels:      string(labelsJSON),
		Annotations: string(annotationsJSON),
		Targets:     string(targetsJSON),
		Version:     "v1.0.0",
		Status:      status,
		CreatedAt:   oldRule.CreatedAt,
		UpdatedAt:   oldRule.UpdatedAt,
		DeletedAt:   oldRule.DeletedAt,
	}, nil
}

// 更新告警数据
func (m *DataMigrator) updateAlerts() error {
	log.Println("Updating alerts with new fields...")

	if m.config.DryRun {
		log.Println("[DRY RUN] Would update alerts with fingerprints and analysis status")
		return nil
	}

	// 添加新字段（如果不存在）
	alterSQL := `
		ALTER TABLE alerts 
		ADD COLUMN IF NOT EXISTS analysis_status VARCHAR(20) DEFAULT 'pending' AFTER analysis,
		ADD COLUMN IF NOT EXISTS analysis_result JSON AFTER analysis_status,
		ADD COLUMN IF NOT EXISTS ai_summary TEXT AFTER analysis_result,
		ADD COLUMN IF NOT EXISTS similar_alerts JSON AFTER ai_summary,
		ADD COLUMN IF NOT EXISTS resolution_suggestion TEXT AFTER similar_alerts,
		ADD COLUMN IF NOT EXISTS fingerprint VARCHAR(64) AFTER resolution_suggestion
	`
	if err := m.db.Exec(alterSQL).Error; err != nil {
		log.Printf("Warning: Failed to add new columns (they may already exist): %v", err)
	}

	// 批量更新告警指纹
	batchSize := m.config.BatchSize
	offset := 0

	for {
		var alerts []OldAlert
		if err := m.db.Limit(batchSize).Offset(offset).Where("deleted_at IS NULL").Find(&alerts).Error; err != nil {
			return fmt.Errorf("failed to fetch alerts: %w", err)
		}

		if len(alerts) == 0 {
			break
		}

		for _, alert := range alerts {
			fingerprint := m.generateFingerprint(&alert)
			analysisStatus := "pending"
			if alert.Analysis != "" {
				analysisStatus = "completed"
			}

			updateSQL := `
				UPDATE alerts 
				SET fingerprint = ?, analysis_status = ?
				WHERE id = ? AND (fingerprint IS NULL OR fingerprint = '')
			`
			if err := m.db.Exec(updateSQL, fingerprint, analysisStatus, alert.ID).Error; err != nil {
				log.Printf("Failed to update alert %d: %v", alert.ID, err)
				m.stats.ErrorsEncountered++
				continue
			}
		}

		m.stats.AlertsUpdated += len(alerts)
		offset += batchSize

		if m.config.Verbose {
			log.Printf("Updated %d alerts...", m.stats.AlertsUpdated)
		}
	}

	m.stats.AlertsProcessed = m.stats.AlertsUpdated
	log.Printf("Alerts update completed: %d processed, %d updated",
		m.stats.AlertsProcessed, m.stats.AlertsUpdated)
	return nil
}

// 生成告警指纹
func (m *DataMigrator) generateFingerprint(alert *OldAlert) string {
	content := fmt.Sprintf("%s|%s|%s", alert.Name, alert.Source, alert.Content)
	// 这里应该使用SHA256，为了简化使用简单的哈希
	return fmt.Sprintf("%x", []byte(content))[:64]
}

// 创建配置同步状态
func (m *DataMigrator) createConfigSyncStatus() error {
	log.Println("Creating config sync status records...")

	if m.config.DryRun {
		log.Println("[DRY RUN] Would create config sync status records")
		return nil
	}

	// 从providers表获取数据源信息
	var providers []struct {
		ID        uint   `gorm:"primarykey"`
		Type      string `gorm:"size:50;not null"`
		Status    string `gorm:"size:50;not null;default:'active'"`
		CreatedAt time.Time
		UpdatedAt time.Time
	}

	if err := m.db.Table("providers").Where("deleted_at IS NULL").Find(&providers).Error; err != nil {
		return fmt.Errorf("failed to fetch providers: %w", err)
	}

	for _, provider := range providers {
		configType := "prometheus"
		switch provider.Type {
		case "alertmanager":
			configType = "alertmanager"
		case "victoriametrics":
			configType = "vmalert"
		}

		syncStatus := "pending"
		if provider.Status == "active" {
			syncStatus = "success"
		}

		insertSQL := `
			INSERT IGNORE INTO config_sync_status 
			(id, cluster_id, config_type, sync_status, created_at, updated_at)
			VALUES (?, ?, ?, ?, ?, ?)
		`
		if err := m.db.Exec(insertSQL,
			uuid.New().String(),
			fmt.Sprintf("cluster-%d", provider.ID),
			configType,
			syncStatus,
			provider.CreatedAt,
			provider.UpdatedAt,
		).Error; err != nil {
			log.Printf("Failed to create config sync status for provider %d: %v", provider.ID, err)
			m.stats.ErrorsEncountered++
			continue
		}

		m.stats.ConfigsCreated++
	}

	log.Printf("Config sync status creation completed: %d created", m.stats.ConfigsCreated)
	return nil
}

// 验证迁移结果
func (m *DataMigrator) validateMigration() error {
	log.Println("Validating migration results...")

	// 检查规则迁移
	var oldRuleCount, newRuleCount int64
	m.db.Model(&OldRule{}).Where("deleted_at IS NULL").Count(&oldRuleCount)
	m.db.Model(&NewRule{}).Where("deleted_at IS NULL").Count(&newRuleCount)

	if oldRuleCount != newRuleCount {
		return fmt.Errorf("rule count mismatch: old=%d, new=%d", oldRuleCount, newRuleCount)
	}

	// 检查告警指纹
	var alertCount, fingerprintCount int64
	m.db.Model(&OldAlert{}).Where("deleted_at IS NULL").Count(&alertCount)
	m.db.Model(&OldAlert{}).Where("deleted_at IS NULL AND fingerprint IS NOT NULL AND fingerprint != ''").Count(&fingerprintCount)

	if alertCount != fingerprintCount {
		log.Printf("Warning: Not all alerts have fingerprints: total=%d, with_fingerprint=%d", alertCount, fingerprintCount)
	}

	log.Println("Migration validation completed successfully")
	return nil
}

// 打印统计信息
func (m *DataMigrator) printStats() {
	duration := m.stats.EndTime.Sub(m.stats.StartTime)

	// fmt.Println("\n" + "="*60)
	fmt.Println("MIGRATION STATISTICS")
	// fmt.Println("="*60)
	fmt.Printf("Duration: %v\n", duration)
	fmt.Printf("Rules Processed: %d\n", m.stats.RulesProcessed)
	fmt.Printf("Rules Migrated: %d\n", m.stats.RulesMigrated)
	fmt.Printf("Alerts Processed: %d\n", m.stats.AlertsProcessed)
	fmt.Printf("Alerts Updated: %d\n", m.stats.AlertsUpdated)
	fmt.Printf("Config Sync Records Created: %d\n", m.stats.ConfigsCreated)
	fmt.Printf("Errors Encountered: %d\n", m.stats.ErrorsEncountered)
	// fmt.Println("="*60)
}

func main() {
	var config MigrationConfig

	flag.StringVar(&config.DSN, "dsn", "", "Database DSN (required)")
	flag.BoolVar(&config.DryRun, "dry-run", false, "Perform a dry run without making changes")
	flag.IntVar(&config.BatchSize, "batch-size", 1000, "Batch size for processing")
	flag.BoolVar(&config.Verbose, "verbose", false, "Enable verbose logging")
	flag.BoolVar(&config.SkipBackup, "skip-backup", false, "Skip creating backup tables")
	flag.Parse()

	if config.DSN == "" {
		fmt.Println("Error: DSN is required")
		flag.Usage()
		os.Exit(1)
	}

	migrator, err := NewDataMigrator(&config)
	if err != nil {
		log.Fatalf("Failed to create migrator: %v", err)
	}

	if err := migrator.Migrate(); err != nil {
		log.Fatalf("Migration failed: %v", err)
	}
}
