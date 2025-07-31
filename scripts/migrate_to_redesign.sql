-- AlertAgent 重构数据迁移脚本
-- 将现有数据迁移到重构后的表结构

USE alert_agent;

-- 开始事务
START TRANSACTION;

-- ============================================================================
-- 1. 数据迁移前的准备工作
-- ============================================================================

-- 创建迁移日志表
CREATE TABLE IF NOT EXISTS `migration_log` (
    `id` INT AUTO_INCREMENT PRIMARY KEY,
    `migration_name` VARCHAR(255) NOT NULL,
    `status` VARCHAR(20) NOT NULL DEFAULT 'running',
    `start_time` TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    `end_time` TIMESTAMP NULL,
    `error_message` TEXT,
    `records_processed` INT DEFAULT 0,
    `records_migrated` INT DEFAULT 0,
    `records_failed` INT DEFAULT 0
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 记录迁移开始
INSERT INTO migration_log (migration_name, status) VALUES ('redesign_migration', 'running');
SET @migration_id = LAST_INSERT_ID();

-- ============================================================================
-- 2. 迁移告警规则数据 (rules -> alert_rules)
-- ============================================================================

-- 检查是否需要迁移规则数据
SET @rule_count = (SELECT COUNT(*) FROM rules WHERE deleted_at IS NULL);

-- 迁移规则数据
INSERT INTO alert_rules (
    id, 
    name, 
    expression, 
    duration, 
    severity, 
    labels, 
    annotations, 
    targets, 
    version, 
    status, 
    created_at, 
    updated_at
)
SELECT 
    UUID() as id,
    name,
    condition_expr as expression,
    '5m' as duration,  -- 默认持续时间
    CASE 
        WHEN level = 'critical' THEN 'critical'
        WHEN level = 'warning' THEN 'warning'
        WHEN level = 'high' THEN 'high'
        WHEN level = 'medium' THEN 'medium'
        WHEN level = 'low' THEN 'low'
        ELSE 'medium'
    END as severity,
    '{}' as labels,  -- 默认空标签
    JSON_OBJECT(
        'description', description,
        'notify_type', notify_type,
        'notify_group', notify_group,
        'template', template
    ) as annotations,
    JSON_ARRAY('prometheus') as targets,  -- 默认目标为prometheus
    'v1.0.0' as version,
    CASE 
        WHEN enabled = 1 THEN 'active'
        ELSE 'inactive'
    END as status,
    created_at,
    updated_at
FROM rules 
WHERE deleted_at IS NULL
ON DUPLICATE KEY UPDATE
    name = VALUES(name),
    expression = VALUES(expression),
    updated_at = VALUES(updated_at);

-- 更新迁移日志
UPDATE migration_log 
SET records_processed = @rule_count, 
    records_migrated = (SELECT COUNT(*) FROM alert_rules)
WHERE id = @migration_id;

-- ============================================================================
-- 3. 迁移告警数据，添加新字段
-- ============================================================================

-- 为现有告警添加新的分析相关字段（如果还没有的话）
ALTER TABLE alerts 
ADD COLUMN IF NOT EXISTS analysis_status VARCHAR(20) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT 'pending' COMMENT '分析状态' AFTER analysis,
ADD COLUMN IF NOT EXISTS analysis_result JSON COMMENT '分析结果' AFTER analysis_status,
ADD COLUMN IF NOT EXISTS ai_summary TEXT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci COMMENT 'AI摘要' AFTER analysis_result,
ADD COLUMN IF NOT EXISTS similar_alerts JSON COMMENT '相似告警' AFTER ai_summary,
ADD COLUMN IF NOT EXISTS resolution_suggestion TEXT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci COMMENT '解决建议' AFTER similar_alerts,
ADD COLUMN IF NOT EXISTS fingerprint VARCHAR(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci COMMENT '告警指纹' AFTER resolution_suggestion;

-- 为现有告警生成指纹
UPDATE alerts 
SET fingerprint = SHA2(CONCAT(COALESCE(name, ''), '|', COALESCE(source, ''), '|', COALESCE(content, '')), 256)
WHERE fingerprint IS NULL OR fingerprint = '';

-- 为已有分析内容的告警设置分析状态
UPDATE alerts 
SET analysis_status = CASE 
    WHEN analysis IS NOT NULL AND analysis != '' THEN 'completed'
    ELSE 'pending'
END
WHERE analysis_status = 'pending';

-- ============================================================================
-- 4. 迁移通知相关数据
-- ============================================================================

-- 迁移通知模板数据到新的插件配置格式
-- 为邮件模板创建插件配置
INSERT IGNORE INTO notification_plugins (name, display_name, version, config, enabled, priority)
SELECT 
    'email' as name,
    '邮件通知' as display_name,
    '1.0.0' as version,
    JSON_OBJECT(
        'smtp_host', 'localhost',
        'smtp_port', 587,
        'username', '',
        'password', '',
        'from_email', 'noreply@example.com',
        'to_emails', JSON_ARRAY(),
        'use_tls', true,
        'template_type', 'html',
        'template_content', content
    ) as config,
    true as enabled,
    1 as priority
FROM notify_templates 
WHERE type = 'email' 
LIMIT 1;

-- 为短信模板创建插件配置
INSERT IGNORE INTO notification_plugins (name, display_name, version, config, enabled, priority)
SELECT 
    'sms' as name,
    '短信通知' as display_name,
    '1.0.0' as version,
    JSON_OBJECT(
        'provider', 'aliyun',
        'access_key', '',
        'access_secret', '',
        'sign_name', '',
        'template_code', '',
        'template_content', content
    ) as config,
    false as enabled,
    2 as priority
FROM notify_templates 
WHERE type = 'sms' 
LIMIT 1;

-- ============================================================================
-- 5. 创建初始配置同步状态
-- ============================================================================

-- 为现有的数据源创建配置同步状态
INSERT IGNORE INTO config_sync_status (id, cluster_id, config_type, sync_status, created_at, updated_at)
SELECT 
    UUID() as id,
    CONCAT('cluster-', id) as cluster_id,
    CASE 
        WHEN type = 'prometheus' THEN 'prometheus'
        WHEN type = 'alertmanager' THEN 'alertmanager'
        WHEN type = 'victoriametrics' THEN 'vmalert'
        ELSE 'prometheus'
    END as config_type,
    CASE 
        WHEN status = 'active' THEN 'success'
        ELSE 'pending'
    END as sync_status,
    created_at,
    updated_at
FROM providers 
WHERE deleted_at IS NULL;

-- ============================================================================
-- 6. 数据一致性检查
-- ============================================================================

-- 检查迁移后的数据一致性
SET @original_rules_count = (SELECT COUNT(*) FROM rules WHERE deleted_at IS NULL);
SET @migrated_rules_count = (SELECT COUNT(*) FROM alert_rules WHERE deleted_at IS NULL);

SET @original_alerts_count = (SELECT COUNT(*) FROM alerts WHERE deleted_at IS NULL);
SET @alerts_with_fingerprint = (SELECT COUNT(*) FROM alerts WHERE fingerprint IS NOT NULL AND fingerprint != '');

-- 记录一致性检查结果
INSERT INTO migration_log (migration_name, status, records_processed, records_migrated) VALUES
('rules_consistency_check', 
 CASE WHEN @original_rules_count = @migrated_rules_count THEN 'success' ELSE 'warning' END,
 @original_rules_count, 
 @migrated_rules_count),
('alerts_fingerprint_check', 
 CASE WHEN @original_alerts_count = @alerts_with_fingerprint THEN 'success' ELSE 'warning' END,
 @original_alerts_count, 
 @alerts_with_fingerprint);

-- ============================================================================
-- 7. 创建索引优化
-- ============================================================================

-- 为新字段创建索引
ALTER TABLE alerts 
ADD INDEX IF NOT EXISTS idx_analysis_status (analysis_status),
ADD INDEX IF NOT EXISTS idx_fingerprint (fingerprint);

-- 为新表创建必要的索引（如果不存在）
ALTER TABLE alert_rules 
ADD INDEX IF NOT EXISTS idx_severity (severity),
ADD INDEX IF NOT EXISTS idx_version (version);

ALTER TABLE config_sync_status 
ADD INDEX IF NOT EXISTS idx_cluster_config_updated (cluster_id, config_type, updated_at);

-- ============================================================================
-- 8. 完成迁移
-- ============================================================================

-- 更新迁移状态
UPDATE migration_log 
SET status = 'completed', 
    end_time = NOW(),
    records_processed = @original_rules_count + @original_alerts_count,
    records_migrated = @migrated_rules_count + @alerts_with_fingerprint
WHERE id = @migration_id;

-- 提交事务
COMMIT;

-- ============================================================================
-- 9. 迁移后验证查询
-- ============================================================================

-- 验证迁移结果的查询语句（仅用于检查，不会自动执行）
/*
-- 检查规则迁移结果
SELECT 
    'Rules Migration' as check_type,
    (SELECT COUNT(*) FROM rules WHERE deleted_at IS NULL) as original_count,
    (SELECT COUNT(*) FROM alert_rules WHERE deleted_at IS NULL) as migrated_count,
    CASE 
        WHEN (SELECT COUNT(*) FROM rules WHERE deleted_at IS NULL) = 
             (SELECT COUNT(*) FROM alert_rules WHERE deleted_at IS NULL) 
        THEN 'PASS' 
        ELSE 'FAIL' 
    END as status;

-- 检查告警指纹生成结果
SELECT 
    'Alert Fingerprints' as check_type,
    (SELECT COUNT(*) FROM alerts WHERE deleted_at IS NULL) as total_alerts,
    (SELECT COUNT(*) FROM alerts WHERE fingerprint IS NOT NULL AND fingerprint != '') as with_fingerprint,
    CASE 
        WHEN (SELECT COUNT(*) FROM alerts WHERE deleted_at IS NULL) = 
             (SELECT COUNT(*) FROM alerts WHERE fingerprint IS NOT NULL AND fingerprint != '') 
        THEN 'PASS' 
        ELSE 'FAIL' 
    END as status;

-- 检查配置同步状态创建结果
SELECT 
    'Config Sync Status' as check_type,
    (SELECT COUNT(*) FROM providers WHERE deleted_at IS NULL) as providers_count,
    (SELECT COUNT(*) FROM config_sync_status WHERE deleted_at IS NULL) as sync_status_count,
    CASE 
        WHEN (SELECT COUNT(*) FROM providers WHERE deleted_at IS NULL) <= 
             (SELECT COUNT(*) FROM config_sync_status WHERE deleted_at IS NULL) 
        THEN 'PASS' 
        ELSE 'FAIL' 
    END as status;

-- 查看迁移日志
SELECT * FROM migration_log ORDER BY start_time DESC LIMIT 10;
*/

-- 显示迁移完成信息
SELECT 
    'Migration completed successfully!' as message,
    NOW() as completion_time,
    (SELECT COUNT(*) FROM alert_rules) as total_rules,
    (SELECT COUNT(*) FROM alerts WHERE fingerprint IS NOT NULL) as alerts_with_fingerprint,
    (SELECT COUNT(*) FROM config_sync_status) as config_sync_records;