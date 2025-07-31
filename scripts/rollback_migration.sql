-- AlertAgent 重构迁移回滚脚本
-- 用于回滚数据迁移操作

USE alert_agent;

-- 开始事务
START TRANSACTION;

-- ============================================================================
-- 1. 回滚前的安全检查
-- ============================================================================

-- 检查备份表是否存在
SET @backup_exists = (
    SELECT COUNT(*) 
    FROM information_schema.tables 
    WHERE table_schema = 'alert_agent' 
    AND table_name IN ('rules_backup', 'alerts_backup', 'providers_backup')
);

-- 如果备份表不存在，停止回滚
SELECT CASE 
    WHEN @backup_exists < 3 THEN 
        SIGNAL SQLSTATE '45000' SET MESSAGE_TEXT = 'Backup tables not found. Cannot proceed with rollback.'
    ELSE 
        'Backup tables found. Proceeding with rollback...'
END as rollback_status;

-- 记录回滚开始
INSERT INTO migration_log (migration_name, status) VALUES ('rollback_migration', 'running');
SET @rollback_id = LAST_INSERT_ID();

-- ============================================================================
-- 2. 回滚告警规则数据
-- ============================================================================

-- 删除迁移后创建的alert_rules记录
DELETE FROM alert_rules WHERE id NOT IN (
    SELECT CONCAT('legacy-', id) FROM rules_backup WHERE deleted_at IS NULL
);

-- 恢复原始rules表数据（如果有变更）
UPDATE rules r
JOIN rules_backup rb ON r.id = rb.id
SET 
    r.name = rb.name,
    r.description = rb.description,
    r.level = rb.level,
    r.enabled = rb.enabled,
    r.condition_expr = rb.condition_expr,
    r.notify_type = rb.notify_type,
    r.notify_group = rb.notify_group,
    r.template = rb.template,
    r.updated_at = rb.updated_at
WHERE r.updated_at > rb.updated_at;

-- ============================================================================
-- 3. 回滚告警数据的新字段
-- ============================================================================

-- 移除新添加的字段（谨慎操作）
-- 注意：这将永久删除这些字段中的数据
ALTER TABLE alerts 
DROP COLUMN IF EXISTS analysis_status,
DROP COLUMN IF EXISTS analysis_result,
DROP COLUMN IF EXISTS ai_summary,
DROP COLUMN IF EXISTS similar_alerts,
DROP COLUMN IF EXISTS resolution_suggestion,
DROP COLUMN IF EXISTS fingerprint;

-- 恢复原始alerts表数据（如果有变更）
UPDATE alerts a
JOIN alerts_backup ab ON a.id = ab.id
SET 
    a.name = ab.name,
    a.title = ab.title,
    a.level = ab.level,
    a.status = ab.status,
    a.source = ab.source,
    a.content = ab.content,
    a.labels = ab.labels,
    a.analysis = ab.analysis,
    a.updated_at = ab.updated_at
WHERE a.updated_at > ab.updated_at;

-- ============================================================================
-- 4. 清理新创建的表和数据
-- ============================================================================

-- 清空新创建的表
TRUNCATE TABLE rule_versions;
TRUNCATE TABLE rule_distribution_records;
TRUNCATE TABLE config_sync_status;
TRUNCATE TABLE config_versions;
TRUNCATE TABLE config_sync_history;
TRUNCATE TABLE config_sync_exceptions;
TRUNCATE TABLE task_queue;
TRUNCATE TABLE worker_instances;
TRUNCATE TABLE task_execution_history;
TRUNCATE TABLE user_notification_configs;
TRUNCATE TABLE notification_records;

-- 恢复原始providers表数据（如果有变更）
UPDATE providers p
JOIN providers_backup pb ON p.id = pb.id
SET 
    p.name = pb.name,
    p.type = pb.type,
    p.status = pb.status,
    p.description = pb.description,
    p.endpoint = pb.endpoint,
    p.auth_type = pb.auth_type,
    p.auth_config = pb.auth_config,
    p.labels = pb.labels,
    p.last_check = pb.last_check,
    p.last_error = pb.last_error,
    p.updated_at = pb.updated_at
WHERE p.updated_at > pb.updated_at;

-- ============================================================================
-- 5. 移除新创建的索引
-- ============================================================================

-- 移除为新字段创建的索引
ALTER TABLE alerts 
DROP INDEX IF EXISTS idx_analysis_status,
DROP INDEX IF EXISTS idx_fingerprint;

ALTER TABLE alert_rules 
DROP INDEX IF EXISTS idx_severity,
DROP INDEX IF EXISTS idx_version;

-- ============================================================================
-- 6. 重置通知插件配置
-- ============================================================================

-- 删除迁移时创建的通知插件配置
DELETE FROM notification_plugins 
WHERE name IN ('email', 'sms') 
AND created_at >= (
    SELECT start_time 
    FROM migration_log 
    WHERE migration_name = 'redesign_migration' 
    ORDER BY start_time DESC 
    LIMIT 1
);

-- ============================================================================
-- 7. 数据一致性验证
-- ============================================================================

-- 验证回滚后的数据一致性
SET @current_rules_count = (SELECT COUNT(*) FROM rules WHERE deleted_at IS NULL);
SET @backup_rules_count = (SELECT COUNT(*) FROM rules_backup WHERE deleted_at IS NULL);

SET @current_alerts_count = (SELECT COUNT(*) FROM alerts WHERE deleted_at IS NULL);
SET @backup_alerts_count = (SELECT COUNT(*) FROM alerts_backup WHERE deleted_at IS NULL);

-- 记录一致性检查结果
INSERT INTO migration_log (migration_name, status, records_processed, records_migrated) VALUES
('rollback_rules_check', 
 CASE WHEN @current_rules_count = @backup_rules_count THEN 'success' ELSE 'warning' END,
 @backup_rules_count, 
 @current_rules_count),
('rollback_alerts_check', 
 CASE WHEN @current_alerts_count = @backup_alerts_count THEN 'success' ELSE 'warning' END,
 @backup_alerts_count, 
 @current_alerts_count);

-- ============================================================================
-- 8. 完成回滚
-- ============================================================================

-- 更新回滚状态
UPDATE migration_log 
SET status = 'completed', 
    end_time = NOW(),
    records_processed = @backup_rules_count + @backup_alerts_count,
    records_migrated = @current_rules_count + @current_alerts_count
WHERE id = @rollback_id;

-- 提交事务
COMMIT;

-- ============================================================================
-- 9. 回滚后验证查询
-- ============================================================================

-- 验证回滚结果的查询语句
SELECT 
    'Rollback Verification' as check_type,
    'Rules' as table_name,
    (SELECT COUNT(*) FROM rules_backup WHERE deleted_at IS NULL) as backup_count,
    (SELECT COUNT(*) FROM rules WHERE deleted_at IS NULL) as current_count,
    CASE 
        WHEN (SELECT COUNT(*) FROM rules_backup WHERE deleted_at IS NULL) = 
             (SELECT COUNT(*) FROM rules WHERE deleted_at IS NULL) 
        THEN 'PASS' 
        ELSE 'FAIL' 
    END as status
UNION ALL
SELECT 
    'Rollback Verification' as check_type,
    'Alerts' as table_name,
    (SELECT COUNT(*) FROM alerts_backup WHERE deleted_at IS NULL) as backup_count,
    (SELECT COUNT(*) FROM alerts WHERE deleted_at IS NULL) as current_count,
    CASE 
        WHEN (SELECT COUNT(*) FROM alerts_backup WHERE deleted_at IS NULL) = 
             (SELECT COUNT(*) FROM alerts WHERE deleted_at IS NULL) 
        THEN 'PASS' 
        ELSE 'FAIL' 
    END as status;

-- 检查新字段是否已移除
SELECT 
    'New Fields Removal' as check_type,
    COUNT(*) as remaining_new_columns
FROM information_schema.columns 
WHERE table_schema = 'alert_agent' 
AND table_name = 'alerts' 
AND column_name IN ('analysis_status', 'analysis_result', 'ai_summary', 'similar_alerts', 'resolution_suggestion', 'fingerprint');

-- 显示回滚完成信息
SELECT 
    'Rollback completed!' as message,
    NOW() as completion_time,
    (SELECT COUNT(*) FROM rules WHERE deleted_at IS NULL) as current_rules,
    (SELECT COUNT(*) FROM alerts WHERE deleted_at IS NULL) as current_alerts;

-- ============================================================================
-- 10. 清理备份表（可选，谨慎执行）
-- ============================================================================

-- 取消注释以下语句来删除备份表
-- 警告：这将永久删除备份数据，请确保回滚成功后再执行

/*
DROP TABLE IF EXISTS rules_backup;
DROP TABLE IF EXISTS alerts_backup;
DROP TABLE IF EXISTS providers_backup;

-- 清理迁移日志（可选）
DELETE FROM migration_log WHERE migration_name IN ('redesign_migration', 'rollback_migration');
*/