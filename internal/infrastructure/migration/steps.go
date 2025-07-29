package migration

import (
	"crypto/md5"
	"fmt"

	"gorm.io/gorm"
)

// GetAllMigrationSteps 获取所有迁移步骤
func GetAllMigrationSteps() []MigrationStep {
	return []MigrationStep{
		// V1.0.0 -> V2.0.0: 基础架构升级
		{
			Version:     "v2.0.0-001",
			Name:        "Create Alert Channels Table",
			Description: "创建告警渠道管理表",
			Up:          createAlertChannelsTable,
			Down:        dropAlertChannelsTable,
			Checksum:    calculateChecksum("create_alert_channels_table"),
		},
		{
			Version:     "v2.0.0-002",
			Name:        "Create Channel Groups Table",
			Description: "创建渠道分组表",
			Up:          createChannelGroupsTable,
			Down:        dropChannelGroupsTable,
			Checksum:    calculateChecksum("create_channel_groups_table"),
		},
		{
			Version:     "v2.0.0-003",
			Name:        "Create Channel Templates Table",
			Description: "创建渠道模板表",
			Up:          createChannelTemplatesTable,
			Down:        dropChannelTemplatesTable,
			Checksum:    calculateChecksum("create_channel_templates_table"),
		},
		{
			Version:     "v2.0.0-004",
			Name:        "Create Channel Usage Stats Table",
			Description: "创建渠道使用统计表",
			Up:          createChannelUsageStatsTable,
			Down:        dropChannelUsageStatsTable,
			Checksum:    calculateChecksum("create_channel_usage_stats_table"),
		},
		{
			Version:     "v2.0.0-005",
			Name:        "Create Channel Permissions Table",
			Description: "创建渠道权限表",
			Up:          createChannelPermissionsTable,
			Down:        dropChannelPermissionsTable,
			Checksum:    calculateChecksum("create_channel_permissions_table"),
		},
		{
			Version:     "v2.0.0-006",
			Name:        "Create Alertmanager Clusters Table",
			Description: "创建Alertmanager集群管理表",
			Up:          createAlertmanagerClustersTable,
			Down:        dropAlertmanagerClustersTable,
			Checksum:    calculateChecksum("create_alertmanager_clusters_table"),
		},
		{
			Version:     "v2.0.0-007",
			Name:        "Create Rule Distributions Table",
			Description: "创建规则分发记录表",
			Up:          createRuleDistributionsTable,
			Down:        dropRuleDistributionsTable,
			Checksum:    calculateChecksum("create_rule_distributions_table"),
		},
		{
			Version:     "v2.0.0-008",
			Name:        "Create Alert Processing Records Table",
			Description: "创建告警处理记录表",
			Up:          createAlertProcessingRecordsTable,
			Down:        dropAlertProcessingRecordsTable,
			Checksum:    calculateChecksum("create_alert_processing_records_table"),
		},
		{
			Version:     "v2.0.0-009",
			Name:        "Create AI Analysis Records Table",
			Description: "创建AI分析记录表",
			Up:          createAIAnalysisRecordsTable,
			Down:        dropAIAnalysisRecordsTable,
			Checksum:    calculateChecksum("create_ai_analysis_records_table"),
		},
		{
			Version:     "v2.0.0-010",
			Name:        "Create Automation Actions Table",
			Description: "创建自动化操作记录表",
			Up:          createAutomationActionsTable,
			Down:        dropAutomationActionsTable,
			Checksum:    calculateChecksum("create_automation_actions_table"),
		},
		{
			Version:     "v2.0.0-011",
			Name:        "Create Alert Convergence Records Table",
			Description: "创建告警收敛记录表",
			Up:          createAlertConvergenceRecordsTable,
			Down:        dropAlertConvergenceRecordsTable,
			Checksum:    calculateChecksum("create_alert_convergence_records_table"),
		},
		{
			Version:     "v2.0.0-012",
			Name:        "Create Cluster Health Status Table",
			Description: "创建集群健康状态表",
			Up:          createClusterHealthStatusTable,
			Down:        dropClusterHealthStatusTable,
			Checksum:    calculateChecksum("create_cluster_health_status_table"),
		},
		{
			Version:     "v2.0.0-013",
			Name:        "Extend Existing Tables",
			Description: "扩展现有表结构以支持新功能",
			Up:          extendExistingTables,
			Down:        revertExistingTablesExtension,
			Checksum:    calculateChecksum("extend_existing_tables"),
		},
		{
			Version:     "v2.0.0-014",
			Name:        "Migrate Legacy Data",
			Description: "迁移V1版本的遗留数据",
			Up:          migrateLegacyData,
			Down:        revertLegacyDataMigration,
			Checksum:    calculateChecksum("migrate_legacy_data"),
		},
		{
			Version:     "v2.0.0-015",
			Name:        "Create Indexes and Constraints",
			Description: "创建索引和约束以优化性能",
			Up:          createIndexesAndConstraints,
			Down:        dropIndexesAndConstraints,
			Checksum:    calculateChecksum("create_indexes_and_constraints"),
		},
	}
}

// calculateChecksum 计算校验和
func calculateChecksum(content string) string {
	hash := md5.Sum([]byte(content))
	return fmt.Sprintf("%x", hash)
}

// createAlertChannelsTable 创建告警渠道表
func createAlertChannelsTable(db *gorm.DB) error {
	return db.Exec(`
		CREATE TABLE IF NOT EXISTS alert_channels (
			id VARCHAR(36) PRIMARY KEY,
			name VARCHAR(100) NOT NULL,
			type VARCHAR(50) NOT NULL,
			description TEXT,
			config JSON NOT NULL,
			group_id VARCHAR(36),
			tags JSON,
			status ENUM('active', 'inactive', 'error') DEFAULT 'active',
			health_status ENUM('healthy', 'unhealthy', 'unknown') DEFAULT 'unknown',
			last_health_check TIMESTAMP NULL,
			health_error_message TEXT,
			created_by VARCHAR(36),
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			INDEX idx_type (type),
			INDEX idx_status (status),
			INDEX idx_group_id (group_id),
			INDEX idx_created_by (created_by)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='告警渠道表';
	`).Error
}

func dropAlertChannelsTable(db *gorm.DB) error {
	return db.Exec("DROP TABLE IF EXISTS alert_channels").Error
}

// createChannelGroupsTable 创建渠道分组表
func createChannelGroupsTable(db *gorm.DB) error {
	return db.Exec(`
		CREATE TABLE IF NOT EXISTS channel_groups (
			id VARCHAR(36) PRIMARY KEY,
			name VARCHAR(100) NOT NULL UNIQUE,
			description TEXT,
			parent_id VARCHAR(36),
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			INDEX idx_parent_id (parent_id)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='渠道分组表';
	`).Error
}

func dropChannelGroupsTable(db *gorm.DB) error {
	return db.Exec("DROP TABLE IF EXISTS channel_groups").Error
}

// createChannelTemplatesTable 创建渠道模板表
func createChannelTemplatesTable(db *gorm.DB) error {
	return db.Exec(`
		CREATE TABLE IF NOT EXISTS channel_templates (
			id VARCHAR(36) PRIMARY KEY,
			name VARCHAR(100) NOT NULL,
			type VARCHAR(50) NOT NULL,
			description TEXT,
			config_template JSON NOT NULL,
			created_by VARCHAR(36),
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			INDEX idx_type (type),
			INDEX idx_created_by (created_by)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='渠道模板表';
	`).Error
}

func dropChannelTemplatesTable(db *gorm.DB) error {
	return db.Exec("DROP TABLE IF EXISTS channel_templates").Error
}

// createChannelUsageStatsTable 创建渠道使用统计表
func createChannelUsageStatsTable(db *gorm.DB) error {
	return db.Exec(`
		CREATE TABLE IF NOT EXISTS channel_usage_stats (
			id VARCHAR(36) PRIMARY KEY,
			channel_id VARCHAR(36) NOT NULL,
			date DATE NOT NULL,
			total_messages INT DEFAULT 0,
			success_messages INT DEFAULT 0,
			failed_messages INT DEFAULT 0,
			avg_response_time INT DEFAULT 0,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			UNIQUE KEY uk_channel_date (channel_id, date),
			INDEX idx_date (date)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='渠道使用统计表';
	`).Error
}

func dropChannelUsageStatsTable(db *gorm.DB) error {
	return db.Exec("DROP TABLE IF EXISTS channel_usage_stats").Error
}

// createChannelPermissionsTable 创建渠道权限表
func createChannelPermissionsTable(db *gorm.DB) error {
	return db.Exec(`
		CREATE TABLE IF NOT EXISTS channel_permissions (
			id VARCHAR(36) PRIMARY KEY,
			user_id VARCHAR(36) NOT NULL,
			channel_id VARCHAR(36) NOT NULL,
			permission ENUM('read', 'write', 'admin') NOT NULL,
			granted_by VARCHAR(36),
			granted_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			UNIQUE KEY uk_user_channel (user_id, channel_id),
			INDEX idx_user_id (user_id),
			INDEX idx_channel_id (channel_id)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='渠道权限表';
	`).Error
}

func dropChannelPermissionsTable(db *gorm.DB) error {
	return db.Exec("DROP TABLE IF EXISTS channel_permissions").Error
}

// createAlertmanagerClustersTable 创建Alertmanager集群管理表
func createAlertmanagerClustersTable(db *gorm.DB) error {
	return db.Exec(`
		CREATE TABLE IF NOT EXISTS alertmanager_clusters (
			id VARCHAR(36) PRIMARY KEY,
			name VARCHAR(100) NOT NULL UNIQUE,
			endpoint VARCHAR(255) NOT NULL,
			config_path VARCHAR(255),
			rules_path VARCHAR(255),
			sync_interval INT DEFAULT 30,
			health_check_interval INT DEFAULT 10,
			status ENUM('active', 'inactive', 'error') DEFAULT 'active',
			labels JSON,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			INDEX idx_name (name),
			INDEX idx_status (status)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='Alertmanager集群管理表';
	`).Error
}

func dropAlertmanagerClustersTable(db *gorm.DB) error {
	return db.Exec("DROP TABLE IF EXISTS alertmanager_clusters").Error
}

// createRuleDistributionsTable 创建规则分发记录表
func createRuleDistributionsTable(db *gorm.DB) error {
	return db.Exec(`
		CREATE TABLE IF NOT EXISTS rule_distributions (
			id VARCHAR(36) PRIMARY KEY,
			rule_id BIGINT UNSIGNED NOT NULL,
			cluster_id VARCHAR(36) NOT NULL,
			version VARCHAR(50),
			status ENUM('pending', 'synced', 'failed') DEFAULT 'pending',
			sync_time TIMESTAMP NULL,
			error_message TEXT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			UNIQUE KEY uk_rule_cluster (rule_id, cluster_id),
			INDEX idx_status (status),
			INDEX idx_sync_time (sync_time)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='规则分发记录表';
	`).Error
}

func dropRuleDistributionsTable(db *gorm.DB) error {
	return db.Exec("DROP TABLE IF EXISTS rule_distributions").Error
}

// createAlertProcessingRecordsTable 创建告警处理记录表
func createAlertProcessingRecordsTable(db *gorm.DB) error {
	return db.Exec(`
		CREATE TABLE IF NOT EXISTS alert_processing_records (
			id VARCHAR(36) PRIMARY KEY,
			alert_id VARCHAR(100) NOT NULL,
			alert_name VARCHAR(100),
			severity VARCHAR(20),
			cluster_id VARCHAR(36),
			received_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			processed_at TIMESTAMP NULL,
			processing_status ENUM('received', 'analyzing', 'processed', 'failed') DEFAULT 'received',
			analysis_id VARCHAR(36),
			decision JSON,
			action_taken VARCHAR(100),
			resolution_time INT,
			feedback_score DECIMAL(3,2),
			labels JSON,
			annotations JSON,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			INDEX idx_alert_id (alert_id),
			INDEX idx_status (processing_status),
			INDEX idx_received_at (received_at),
			INDEX idx_cluster_id (cluster_id)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='告警处理记录表';
	`).Error
}

func dropAlertProcessingRecordsTable(db *gorm.DB) error {
	return db.Exec("DROP TABLE IF EXISTS alert_processing_records").Error
}

// createAIAnalysisRecordsTable 创建AI分析记录表
func createAIAnalysisRecordsTable(db *gorm.DB) error {
	return db.Exec(`
		CREATE TABLE IF NOT EXISTS ai_analysis_records (
			id VARCHAR(36) PRIMARY KEY,
			alert_id VARCHAR(100) NOT NULL,
			analysis_type VARCHAR(50) DEFAULT 'root_cause_analysis',
			request_data JSON,
			response_data JSON,
			analysis_result JSON,
			confidence_score DECIMAL(3,2),
			processing_time INT,
			status ENUM('pending', 'processing', 'completed', 'failed') DEFAULT 'pending',
			error_message TEXT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			INDEX idx_alert_id (alert_id),
			INDEX idx_status (status),
			INDEX idx_analysis_type (analysis_type)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='AI分析记录表';
	`).Error
}

func dropAIAnalysisRecordsTable(db *gorm.DB) error {
	return db.Exec("DROP TABLE IF EXISTS ai_analysis_records").Error
}

// createAutomationActionsTable 创建自动化操作记录表
func createAutomationActionsTable(db *gorm.DB) error {
	return db.Exec(`
		CREATE TABLE IF NOT EXISTS automation_actions (
			id VARCHAR(36) PRIMARY KEY,
			alert_id VARCHAR(100) NOT NULL,
			action_type VARCHAR(100) NOT NULL,
			target_info JSON,
			parameters JSON,
			execution_status ENUM('pending', 'executing', 'completed', 'failed') DEFAULT 'pending',
			execution_result JSON,
			execution_time INT,
			error_message TEXT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			INDEX idx_alert_id (alert_id),
			INDEX idx_action_type (action_type),
			INDEX idx_status (execution_status)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='自动化操作记录表';
	`).Error
}

func dropAutomationActionsTable(db *gorm.DB) error {
	return db.Exec("DROP TABLE IF EXISTS automation_actions").Error
}

// createAlertConvergenceRecordsTable 创建告警收敛记录表
func createAlertConvergenceRecordsTable(db *gorm.DB) error {
	return db.Exec(`
		CREATE TABLE IF NOT EXISTS alert_convergence_records (
			id VARCHAR(36) PRIMARY KEY,
			convergence_key VARCHAR(255) NOT NULL,
			alert_count INT DEFAULT 1,
			first_alert_time TIMESTAMP,
			last_alert_time TIMESTAMP,
			convergence_window INT,
			status ENUM('active', 'expired', 'processed') DEFAULT 'active',
			representative_alert_id VARCHAR(100),
			converged_alerts JSON,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			INDEX idx_convergence_key (convergence_key),
			INDEX idx_status (status),
			INDEX idx_last_alert_time (last_alert_time)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='告警收敛记录表';
	`).Error
}

func dropAlertConvergenceRecordsTable(db *gorm.DB) error {
	return db.Exec("DROP TABLE IF EXISTS alert_convergence_records").Error
}

// createClusterHealthStatusTable 创建集群健康状态表
func createClusterHealthStatusTable(db *gorm.DB) error {
	return db.Exec(`
		CREATE TABLE IF NOT EXISTS cluster_health_status (
			id VARCHAR(36) PRIMARY KEY,
			cluster_id VARCHAR(36) NOT NULL,
			status ENUM('healthy', 'warning', 'critical', 'unknown') DEFAULT 'unknown',
			response_time INT,
			success_rate DECIMAL(5,2),
			last_check_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			error_message TEXT,
			metrics JSON,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			INDEX idx_cluster_id (cluster_id),
			INDEX idx_status (status),
			INDEX idx_last_check_time (last_check_time)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='集群健康状态表';
	`).Error
}

func dropClusterHealthStatusTable(db *gorm.DB) error {
	return db.Exec("DROP TABLE IF EXISTS cluster_health_status").Error
}

// extendExistingTables 扩展现有表结构
func extendExistingTables(db *gorm.DB) error {
	// 扩展 rules 表
	if err := db.Exec(`
		ALTER TABLE rules 
		ADD COLUMN IF NOT EXISTS target_clusters JSON AFTER level,
		ADD COLUMN IF NOT EXISTS distribution_status ENUM('pending', 'partial', 'complete', 'failed') DEFAULT 'pending' AFTER target_clusters,
		ADD COLUMN IF NOT EXISTS last_distribution_time TIMESTAMP NULL AFTER distribution_status,
		ADD COLUMN IF NOT EXISTS version VARCHAR(50) DEFAULT 'v1.0.0' AFTER last_distribution_time,
		ADD COLUMN IF NOT EXISTS notification_channels JSON AFTER version;
	`).Error; err != nil {
		return err
	}

	// 扩展 alerts 表
	if err := db.Exec(`
		ALTER TABLE alerts 
		ADD COLUMN IF NOT EXISTS cluster_id VARCHAR(36) AFTER rule_id,
		ADD COLUMN IF NOT EXISTS processing_record_id VARCHAR(36) AFTER cluster_id,
		ADD COLUMN IF NOT EXISTS ai_analysis_id VARCHAR(36) AFTER processing_record_id,
		ADD COLUMN IF NOT EXISTS convergence_id VARCHAR(36) AFTER ai_analysis_id,
		ADD COLUMN IF NOT EXISTS auto_resolved BOOLEAN DEFAULT FALSE AFTER convergence_id,
		ADD COLUMN IF NOT EXISTS resolution_method VARCHAR(100) AFTER auto_resolved,
		ADD COLUMN IF NOT EXISTS notification_channels JSON AFTER resolution_method;
	`).Error; err != nil {
		return err
	}

	return nil
}

func revertExistingTablesExtension(db *gorm.DB) error {
	// 回滚 rules 表扩展
	if err := db.Exec(`
		ALTER TABLE rules 
		DROP COLUMN IF EXISTS target_clusters,
		DROP COLUMN IF EXISTS distribution_status,
		DROP COLUMN IF EXISTS last_distribution_time,
		DROP COLUMN IF EXISTS version,
		DROP COLUMN IF EXISTS notification_channels;
	`).Error; err != nil {
		return err
	}

	// 回滚 alerts 表扩展
	if err := db.Exec(`
		ALTER TABLE alerts 
		DROP COLUMN IF EXISTS cluster_id,
		DROP COLUMN IF EXISTS processing_record_id,
		DROP COLUMN IF EXISTS ai_analysis_id,
		DROP COLUMN IF EXISTS convergence_id,
		DROP COLUMN IF EXISTS auto_resolved,
		DROP COLUMN IF EXISTS resolution_method,
		DROP COLUMN IF EXISTS notification_channels;
	`).Error; err != nil {
		return err
	}

	return nil
}

// migrateLegacyData 迁移V1版本的遗留数据
func migrateLegacyData(db *gorm.DB) error {
	// 迁移通知组到渠道系统
	if err := db.Exec(`
		INSERT INTO alert_channels (id, name, type, description, config, status, created_at, updated_at)
		SELECT 
			CONCAT('legacy-', id) as id,
			name,
			'legacy_group' as type,
			description,
			JSON_OBJECT('contacts', contacts, 'legacy_id', id) as config,
			IF(enabled = 1, 'active', 'inactive') as status,
			created_at,
			updated_at
		FROM notify_groups
		WHERE deleted_at IS NULL;
	`).Error; err != nil {
		return err
	}

	// 迁移数据源到集群管理
	if err := db.Exec(`
		INSERT INTO alertmanager_clusters (id, name, endpoint, status, labels, created_at, updated_at)
		SELECT 
			CONCAT('legacy-provider-', id) as id,
			name,
			endpoint,
			status,
			JSON_OBJECT('type', type, 'auth_type', auth_type, 'legacy_id', id) as labels,
			created_at,
			updated_at
		FROM providers
		WHERE deleted_at IS NULL AND type IN ('prometheus', 'victoriametrics');
	`).Error; err != nil {
		return err
	}

	return nil
}

func revertLegacyDataMigration(db *gorm.DB) error {
	// 删除迁移的数据
	if err := db.Exec("DELETE FROM alert_channels WHERE id LIKE 'legacy-%'").Error; err != nil {
		return err
	}

	if err := db.Exec("DELETE FROM alertmanager_clusters WHERE id LIKE 'legacy-provider-%'").Error; err != nil {
		return err
	}

	return nil
}

// createIndexesAndConstraints 创建索引和约束
func createIndexesAndConstraints(db *gorm.DB) error {
	// 添加外键约束
	constraints := []string{
		"ALTER TABLE channel_groups ADD CONSTRAINT fk_channel_groups_parent FOREIGN KEY (parent_id) REFERENCES channel_groups(id) ON DELETE SET NULL",
		"ALTER TABLE alert_channels ADD CONSTRAINT fk_alert_channels_group FOREIGN KEY (group_id) REFERENCES channel_groups(id) ON DELETE SET NULL",
		"ALTER TABLE channel_usage_stats ADD CONSTRAINT fk_channel_usage_stats_channel FOREIGN KEY (channel_id) REFERENCES alert_channels(id) ON DELETE CASCADE",
		"ALTER TABLE channel_permissions ADD CONSTRAINT fk_channel_permissions_channel FOREIGN KEY (channel_id) REFERENCES alert_channels(id) ON DELETE CASCADE",
		"ALTER TABLE rule_distributions ADD CONSTRAINT fk_rule_distributions_rule FOREIGN KEY (rule_id) REFERENCES rules(id) ON DELETE CASCADE",
		"ALTER TABLE rule_distributions ADD CONSTRAINT fk_rule_distributions_cluster FOREIGN KEY (cluster_id) REFERENCES alertmanager_clusters(id) ON DELETE CASCADE",
		"ALTER TABLE cluster_health_status ADD CONSTRAINT fk_cluster_health_status_cluster FOREIGN KEY (cluster_id) REFERENCES alertmanager_clusters(id) ON DELETE CASCADE",
	}

	for _, constraint := range constraints {
		if err := db.Exec(constraint).Error; err != nil {
			// 忽略已存在的约束错误
			continue
		}
	}

	// 创建复合索引
	indexes := []string{
		"CREATE INDEX IF NOT EXISTS idx_alert_processing_records_composite ON alert_processing_records(processing_status, received_at, cluster_id)",
		"CREATE INDEX IF NOT EXISTS idx_ai_analysis_records_composite ON ai_analysis_records(status, analysis_type, created_at)",
		"CREATE INDEX IF NOT EXISTS idx_automation_actions_composite ON automation_actions(execution_status, action_type, created_at)",
		"CREATE INDEX IF NOT EXISTS idx_alert_convergence_records_composite ON alert_convergence_records(status, convergence_key, last_alert_time)",
		"CREATE INDEX IF NOT EXISTS idx_channel_usage_stats_composite ON channel_usage_stats(channel_id, date, total_messages)",
	}

	for _, index := range indexes {
		if err := db.Exec(index).Error; err != nil {
			// 忽略已存在的索引错误
			continue
		}
	}

	return nil
}

func dropIndexesAndConstraints(db *gorm.DB) error {
	// 删除外键约束
	constraints := []string{
		"ALTER TABLE channel_groups DROP FOREIGN KEY IF EXISTS fk_channel_groups_parent",
		"ALTER TABLE alert_channels DROP FOREIGN KEY IF EXISTS fk_alert_channels_group",
		"ALTER TABLE channel_usage_stats DROP FOREIGN KEY IF EXISTS fk_channel_usage_stats_channel",
		"ALTER TABLE channel_permissions DROP FOREIGN KEY IF EXISTS fk_channel_permissions_channel",
		"ALTER TABLE rule_distributions DROP FOREIGN KEY IF EXISTS fk_rule_distributions_rule",
		"ALTER TABLE rule_distributions DROP FOREIGN KEY IF EXISTS fk_rule_distributions_cluster",
		"ALTER TABLE cluster_health_status DROP FOREIGN KEY IF EXISTS fk_cluster_health_status_cluster",
	}

	for _, constraint := range constraints {
		db.Exec(constraint) // 忽略错误
	}

	// 删除复合索引
	indexes := []string{
		"DROP INDEX IF EXISTS idx_alert_processing_records_composite ON alert_processing_records",
		"DROP INDEX IF EXISTS idx_ai_analysis_records_composite ON ai_analysis_records",
		"DROP INDEX IF EXISTS idx_automation_actions_composite ON automation_actions",
		"DROP INDEX IF EXISTS idx_alert_convergence_records_composite ON alert_convergence_records",
		"DROP INDEX IF EXISTS idx_channel_usage_stats_composite ON channel_usage_stats",
	}

	for _, index := range indexes {
		db.Exec(index) // 忽略错误
	}

	return nil
}