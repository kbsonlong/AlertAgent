-- 创建规则分发记录表
CREATE TABLE IF NOT EXISTS rule_distribution_records (
    id VARCHAR(36) PRIMARY KEY,
    rule_id VARCHAR(36) NOT NULL,
    target VARCHAR(255) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    version VARCHAR(50) NOT NULL,
    config_hash VARCHAR(64),
    last_sync TIMESTAMP NULL,
    error TEXT,
    retry_count INT DEFAULT 0,
    max_retry INT DEFAULT 3,
    next_retry TIMESTAMP NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    
    INDEX idx_rule_distribution_rule_id (rule_id),
    INDEX idx_rule_distribution_target (target),
    INDEX idx_rule_distribution_status (status),
    INDEX idx_rule_distribution_deleted_at (deleted_at),
    INDEX idx_rule_distribution_next_retry (next_retry),
    UNIQUE KEY uk_rule_target (rule_id, target, deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='规则分发记录表';

-- 更新现有的alert_rules表，确保有必要的字段
-- 检查并添加version列
SET @sql = (SELECT IF(
    (SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS 
     WHERE table_name = 'alert_rules' 
     AND table_schema = 'alert_agent' 
     AND column_name = 'version') > 0,
    'SELECT 1',
    'ALTER TABLE alert_rules ADD COLUMN version VARCHAR(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT ''v1.0.0'' AFTER targets'
));
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- 检查并添加status列
SET @sql = (SELECT IF(
    (SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS 
     WHERE table_name = 'alert_rules' 
     AND table_schema = 'alert_agent' 
     AND column_name = 'status') > 0,
    'SELECT 1',
    'ALTER TABLE alert_rules ADD COLUMN status VARCHAR(20) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT ''pending'' AFTER version'
));
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- 添加索引（如果不存在）
SET @sql = (SELECT IF(
    (SELECT COUNT(*) FROM INFORMATION_SCHEMA.STATISTICS 
     WHERE table_name = 'alert_rules' 
     AND table_schema = 'alert_agent' 
     AND index_name = 'idx_alert_rules_status') > 0,
    'SELECT 1',
    'ALTER TABLE alert_rules ADD INDEX idx_alert_rules_status (status)'
));
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @sql = (SELECT IF(
    (SELECT COUNT(*) FROM INFORMATION_SCHEMA.STATISTICS 
     WHERE table_name = 'alert_rules' 
     AND table_schema = 'alert_agent' 
     AND index_name = 'idx_alert_rules_version') > 0,
    'SELECT 1',
    'ALTER TABLE alert_rules ADD INDEX idx_alert_rules_version (version)'
));
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;