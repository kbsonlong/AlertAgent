-- AlertAgent 数据库性能优化脚本
-- 创建索引、优化查询性能、配置连接池等

USE alert_agent;

-- ============================================================================
-- 1. 创建性能优化索引
-- ============================================================================

-- 告警规则表索引优化
ALTER TABLE alert_rules 
ADD INDEX IF NOT EXISTS idx_name_status (name, status),
ADD INDEX IF NOT EXISTS idx_severity_status (severity, status),
ADD INDEX IF NOT EXISTS idx_created_updated (created_at, updated_at),
ADD INDEX IF NOT EXISTS idx_targets_gin (targets(255)),  -- 对JSON字段的前缀索引
ADD INDEX IF NOT EXISTS idx_version_status (version, status);

-- 告警表索引优化
ALTER TABLE alerts 
ADD INDEX IF NOT EXISTS idx_level_status (level, status),
ADD INDEX IF NOT EXISTS idx_source_level (source, level),
ADD INDEX IF NOT EXISTS idx_analysis_status_created (analysis_status, created_at),
ADD INDEX IF NOT EXISTS idx_fingerprint_created (fingerprint, created_at),
ADD INDEX IF NOT EXISTS idx_severity_created (severity, created_at),
ADD INDEX IF NOT EXISTS idx_rule_id_status (rule_id, status),
ADD INDEX IF NOT EXISTS idx_created_status (created_at, status);

-- 规则版本表索引优化
ALTER TABLE rule_versions 
ADD INDEX IF NOT EXISTS idx_rule_version (rule_id, version),
ADD INDEX IF NOT EXISTS idx_created_rule (created_at, rule_id),
ADD INDEX IF NOT EXISTS idx_severity_created (severity, created_at);

-- 规则分发记录表索引优化
ALTER TABLE rule_distribution_records 
ADD INDEX IF NOT EXISTS idx_rule_target_status (rule_id, target, status),
ADD INDEX IF NOT EXISTS idx_status_next_retry (status, next_retry),
ADD INDEX IF NOT EXISTS idx_target_status (target, status),
ADD INDEX IF NOT EXISTS idx_last_sync_status (last_sync, status);

-- 配置同步状态表索引优化
ALTER TABLE config_sync_status 
ADD INDEX IF NOT EXISTS idx_cluster_type_status (cluster_id, config_type, sync_status),
ADD INDEX IF NOT EXISTS idx_sync_time_status (sync_time, sync_status),
ADD INDEX IF NOT EXISTS idx_status_updated (sync_status, updated_at);

-- 配置同步历史表索引优化
ALTER TABLE config_sync_history 
ADD INDEX IF NOT EXISTS idx_cluster_type_created (cluster_id, config_type, created_at),
ADD INDEX IF NOT EXISTS idx_status_duration (sync_status, sync_duration),
ADD INDEX IF NOT EXISTS idx_hash_created (config_hash, created_at);

-- 配置同步异常表索引优化
ALTER TABLE config_sync_exceptions 
ADD INDEX IF NOT EXISTS idx_cluster_type_status (cluster_id, config_type, status),
ADD INDEX IF NOT EXISTS idx_severity_status (severity, status),
ADD INDEX IF NOT EXISTS idx_first_last_occurred (first_occurred, last_occurred),
ADD INDEX IF NOT EXISTS idx_next_retry_status (next_retry_at, status);

-- 任务队列表索引优化
ALTER TABLE task_queue 
ADD INDEX IF NOT EXISTS idx_queue_status_priority (queue_name, status, priority),
ADD INDEX IF NOT EXISTS idx_status_scheduled (status, scheduled_at),
ADD INDEX IF NOT EXISTS idx_type_status (task_type, status),
ADD INDEX IF NOT EXISTS idx_worker_status (worker_id, status),
ADD INDEX IF NOT EXISTS idx_priority_scheduled (priority, scheduled_at);

-- Worker实例表索引优化
ALTER TABLE worker_instances 
ADD INDEX IF NOT EXISTS idx_queue_status (queue_name, status),
ADD INDEX IF NOT EXISTS idx_heartbeat_status (last_heartbeat, status),
ADD INDEX IF NOT EXISTS idx_host_queue (host_name, queue_name);

-- 任务执行历史表索引优化
ALTER TABLE task_execution_history 
ADD INDEX IF NOT EXISTS idx_worker_started (worker_id, started_at),
ADD INDEX IF NOT EXISTS idx_queue_type_status (queue_name, task_type, status),
ADD INDEX IF NOT EXISTS idx_started_duration (started_at, duration_ms),
ADD INDEX IF NOT EXISTS idx_task_status (task_id, status);

-- 用户通知配置表索引优化
ALTER TABLE user_notification_configs 
ADD INDEX IF NOT EXISTS idx_user_plugin_enabled (user_id, plugin_name, enabled),
ADD INDEX IF NOT EXISTS idx_plugin_enabled (plugin_name, enabled);

-- 通知记录表索引优化
ALTER TABLE notification_records 
ADD INDEX IF NOT EXISTS idx_alert_status (alert_id, status),
ADD INDEX IF NOT EXISTS idx_user_plugin_status (user_id, plugin_name, status),
ADD INDEX IF NOT EXISTS idx_sent_status (sent_at, status),
ADD INDEX IF NOT EXISTS idx_created_status (created_at, status),
ADD INDEX IF NOT EXISTS idx_plugin_channel (plugin_name, channel(100));

-- 通知插件表索引优化
ALTER TABLE notification_plugins 
ADD INDEX IF NOT EXISTS idx_enabled_priority (enabled, priority);

-- 知识库表索引优化
ALTER TABLE knowledges 
ADD INDEX IF NOT EXISTS idx_category_created (category, created_at),
ADD INDEX IF NOT EXISTS idx_source_source_id (source, source_id),
ADD INDEX IF NOT EXISTS idx_title_category (title(100), category);

-- 数据源表索引优化
ALTER TABLE providers 
ADD INDEX IF NOT EXISTS idx_type_status_check (type, status, last_check),
ADD INDEX IF NOT EXISTS idx_endpoint_type (endpoint(100), type);

-- ============================================================================
-- 2. 创建复合索引用于常见查询模式
-- ============================================================================

-- 告警查询优化索引
ALTER TABLE alerts 
ADD INDEX IF NOT EXISTS idx_status_level_created (status, level, created_at),
ADD INDEX IF NOT EXISTS idx_analysis_severity_created (analysis_status, severity, created_at);

-- 规则查询优化索引
ALTER TABLE alert_rules 
ADD INDEX IF NOT EXISTS idx_status_severity_updated (status, severity, updated_at);

-- 任务队列查询优化索引
ALTER TABLE task_queue 
ADD INDEX IF NOT EXISTS idx_queue_status_priority_scheduled (queue_name, status, priority, scheduled_at);

-- 配置同步查询优化索引
ALTER TABLE config_sync_status 
ADD INDEX IF NOT EXISTS idx_cluster_status_sync_time (cluster_id, sync_status, sync_time);

-- ============================================================================
-- 3. 全文搜索索引（如果支持）
-- ============================================================================

-- 为告警内容创建全文索引
ALTER TABLE alerts 
ADD FULLTEXT INDEX IF NOT EXISTS ft_content_title (content, title);

-- 为知识库创建全文索引
ALTER TABLE knowledges 
ADD FULLTEXT INDEX IF NOT EXISTS ft_title_content (title, content);

-- ============================================================================
-- 4. 数据分区策略（按时间分区）
-- ============================================================================

-- 为告警表创建按月分区（需要重建表）
-- 注意：这是一个破坏性操作，需要在维护窗口执行

/*
-- 创建分区表的示例（需要根据实际情况调整）
ALTER TABLE alerts 
PARTITION BY RANGE (YEAR(created_at) * 100 + MONTH(created_at)) (
    PARTITION p202401 VALUES LESS THAN (202402),
    PARTITION p202402 VALUES LESS THAN (202403),
    PARTITION p202403 VALUES LESS THAN (202404),
    PARTITION p202404 VALUES LESS THAN (202405),
    PARTITION p202405 VALUES LESS THAN (202406),
    PARTITION p202406 VALUES LESS THAN (202407),
    PARTITION p202407 VALUES LESS THAN (202408),
    PARTITION p202408 VALUES LESS THAN (202409),
    PARTITION p202409 VALUES LESS THAN (202410),
    PARTITION p202410 VALUES LESS THAN (202411),
    PARTITION p202411 VALUES LESS THAN (202412),
    PARTITION p202412 VALUES LESS THAN (202501),
    PARTITION p_future VALUES LESS THAN MAXVALUE
);
*/

-- ============================================================================
-- 5. 数据归档策略
-- ============================================================================

-- 创建归档表
CREATE TABLE IF NOT EXISTS alerts_archive LIKE alerts;
CREATE TABLE IF NOT EXISTS config_sync_history_archive LIKE config_sync_history;
CREATE TABLE IF NOT EXISTS task_execution_history_archive LIKE task_execution_history;
CREATE TABLE IF NOT EXISTS notification_records_archive LIKE notification_records;

-- 创建归档存储过程
DELIMITER //

CREATE PROCEDURE IF NOT EXISTS ArchiveOldData()
BEGIN
    DECLARE EXIT HANDLER FOR SQLEXCEPTION
    BEGIN
        ROLLBACK;
        RESIGNAL;
    END;

    START TRANSACTION;
    
    -- 归档6个月前的已解决告警
    INSERT INTO alerts_archive 
    SELECT * FROM alerts 
    WHERE status = 'resolved' 
    AND updated_at < DATE_SUB(NOW(), INTERVAL 6 MONTH);
    
    DELETE FROM alerts 
    WHERE status = 'resolved' 
    AND updated_at < DATE_SUB(NOW(), INTERVAL 6 MONTH);
    
    -- 归档3个月前的配置同步历史
    INSERT INTO config_sync_history_archive 
    SELECT * FROM config_sync_history 
    WHERE created_at < DATE_SUB(NOW(), INTERVAL 3 MONTH);
    
    DELETE FROM config_sync_history 
    WHERE created_at < DATE_SUB(NOW(), INTERVAL 3 MONTH);
    
    -- 归档1个月前的任务执行历史
    INSERT INTO task_execution_history_archive 
    SELECT * FROM task_execution_history 
    WHERE created_at < DATE_SUB(NOW(), INTERVAL 1 MONTH);
    
    DELETE FROM task_execution_history 
    WHERE created_at < DATE_SUB(NOW(), INTERVAL 1 MONTH);
    
    -- 归档3个月前的通知记录
    INSERT INTO notification_records_archive 
    SELECT * FROM notification_records 
    WHERE created_at < DATE_SUB(NOW(), INTERVAL 3 MONTH);
    
    DELETE FROM notification_records 
    WHERE created_at < DATE_SUB(NOW(), INTERVAL 3 MONTH);
    
    COMMIT;
    
    -- 优化表
    OPTIMIZE TABLE alerts, config_sync_history, task_execution_history, notification_records;
    
END //

DELIMITER ;

-- ============================================================================
-- 6. 数据清理存储过程
-- ============================================================================

DELIMITER //

CREATE PROCEDURE IF NOT EXISTS CleanupExpiredData()
BEGIN
    DECLARE EXIT HANDLER FOR SQLEXCEPTION
    BEGIN
        ROLLBACK;
        RESIGNAL;
    END;

    START TRANSACTION;
    
    -- 清理过期的任务队列记录（已完成超过7天）
    DELETE FROM task_queue 
    WHERE status IN ('completed', 'failed') 
    AND completed_at < DATE_SUB(NOW(), INTERVAL 7 DAY);
    
    -- 清理过期的Worker心跳记录（超过1小时未更新）
    DELETE FROM worker_instances 
    WHERE status = 'inactive' 
    AND last_heartbeat < DATE_SUB(NOW(), INTERVAL 1 HOUR);
    
    -- 清理已解决的配置同步异常（超过30天）
    DELETE FROM config_sync_exceptions 
    WHERE status = 'resolved' 
    AND resolved_at < DATE_SUB(NOW(), INTERVAL 30 DAY);
    
    -- 清理软删除的记录（超过30天）
    DELETE FROM alert_rules WHERE deleted_at < DATE_SUB(NOW(), INTERVAL 30 DAY);
    DELETE FROM user_notification_configs WHERE deleted_at < DATE_SUB(NOW(), INTERVAL 30 DAY);
    DELETE FROM notification_plugins WHERE deleted_at < DATE_SUB(NOW(), INTERVAL 30 DAY);
    
    COMMIT;
    
END //

DELIMITER ;

-- ============================================================================
-- 7. 性能监控视图
-- ============================================================================

-- 创建性能监控视图
CREATE OR REPLACE VIEW v_performance_metrics AS
SELECT 
    'alerts' as table_name,
    COUNT(*) as total_records,
    COUNT(CASE WHEN created_at >= DATE_SUB(NOW(), INTERVAL 1 DAY) THEN 1 END) as records_last_24h,
    COUNT(CASE WHEN created_at >= DATE_SUB(NOW(), INTERVAL 1 HOUR) THEN 1 END) as records_last_hour,
    AVG(CASE WHEN analysis_status = 'completed' THEN 1 ELSE 0 END) * 100 as analysis_completion_rate
FROM alerts WHERE deleted_at IS NULL
UNION ALL
SELECT 
    'task_queue' as table_name,
    COUNT(*) as total_records,
    COUNT(CASE WHEN created_at >= DATE_SUB(NOW(), INTERVAL 1 DAY) THEN 1 END) as records_last_24h,
    COUNT(CASE WHEN created_at >= DATE_SUB(NOW(), INTERVAL 1 HOUR) THEN 1 END) as records_last_hour,
    AVG(CASE WHEN status = 'completed' THEN 1 ELSE 0 END) * 100 as completion_rate
FROM task_queue
UNION ALL
SELECT 
    'config_sync_status' as table_name,
    COUNT(*) as total_records,
    COUNT(CASE WHEN updated_at >= DATE_SUB(NOW(), INTERVAL 1 DAY) THEN 1 END) as records_last_24h,
    COUNT(CASE WHEN updated_at >= DATE_SUB(NOW(), INTERVAL 1 HOUR) THEN 1 END) as records_last_hour,
    AVG(CASE WHEN sync_status = 'success' THEN 1 ELSE 0 END) * 100 as success_rate
FROM config_sync_status WHERE deleted_at IS NULL;

-- 创建慢查询监控视图
CREATE OR REPLACE VIEW v_slow_query_candidates AS
SELECT 
    'Large alerts without fingerprint' as issue_type,
    COUNT(*) as affected_records,
    'Consider running fingerprint generation' as recommendation
FROM alerts 
WHERE (fingerprint IS NULL OR fingerprint = '') AND deleted_at IS NULL
UNION ALL
SELECT 
    'Unanalyzed alerts older than 1 hour' as issue_type,
    COUNT(*) as affected_records,
    'Check AI analysis service' as recommendation
FROM alerts 
WHERE analysis_status = 'pending' 
AND created_at < DATE_SUB(NOW(), INTERVAL 1 HOUR) 
AND deleted_at IS NULL
UNION ALL
SELECT 
    'Failed tasks with high retry count' as issue_type,
    COUNT(*) as affected_records,
    'Review task failure patterns' as recommendation
FROM task_queue 
WHERE status = 'failed' AND retry_count >= max_retry;

-- ============================================================================
-- 8. 数据库配置优化建议
-- ============================================================================

-- 显示当前配置和建议
SELECT 
    'Database Configuration Recommendations' as section,
    'Current Settings' as subsection;

-- 检查关键配置参数
SELECT 
    'innodb_buffer_pool_size' as parameter,
    @@innodb_buffer_pool_size as current_value,
    'Should be 70-80% of available RAM' as recommendation
UNION ALL
SELECT 
    'innodb_log_file_size' as parameter,
    @@innodb_log_file_size as current_value,
    'Should be 25% of buffer pool size' as recommendation
UNION ALL
SELECT 
    'max_connections' as parameter,
    @@max_connections as current_value,
    'Adjust based on application connection pool size' as recommendation
UNION ALL
SELECT 
    'query_cache_size' as parameter,
    @@query_cache_size as current_value,
    'Consider disabling for high-write workloads' as recommendation;

-- ============================================================================
-- 9. 创建定期维护事件
-- ============================================================================

-- 启用事件调度器（如果未启用）
-- SET GLOBAL event_scheduler = ON;

-- 创建每日数据清理事件
/*
CREATE EVENT IF NOT EXISTS evt_daily_cleanup
ON SCHEDULE EVERY 1 DAY
STARTS CURRENT_TIMESTAMP
DO
  CALL CleanupExpiredData();
*/

-- 创建每周数据归档事件
/*
CREATE EVENT IF NOT EXISTS evt_weekly_archive
ON SCHEDULE EVERY 1 WEEK
STARTS CURRENT_TIMESTAMP
DO
  CALL ArchiveOldData();
*/

-- 创建每月表优化事件
/*
CREATE EVENT IF NOT EXISTS evt_monthly_optimize
ON SCHEDULE EVERY 1 MONTH
STARTS CURRENT_TIMESTAMP
DO
BEGIN
  OPTIMIZE TABLE alerts, alert_rules, task_queue, config_sync_history;
  ANALYZE TABLE alerts, alert_rules, task_queue, config_sync_history;
END;
*/

-- ============================================================================
-- 10. 性能测试查询
-- ============================================================================

-- 测试关键查询的性能
SELECT 'Performance Test Queries' as section;

-- 测试告警查询性能
EXPLAIN SELECT * FROM alerts 
WHERE status = 'new' 
AND level = 'critical' 
AND created_at >= DATE_SUB(NOW(), INTERVAL 1 DAY)
ORDER BY created_at DESC 
LIMIT 100;

-- 测试规则查询性能
EXPLAIN SELECT * FROM alert_rules 
WHERE status = 'active' 
AND severity IN ('critical', 'high')
ORDER BY updated_at DESC;

-- 测试任务队列查询性能
EXPLAIN SELECT * FROM task_queue 
WHERE queue_name = 'ai_analysis' 
AND status = 'pending'
ORDER BY priority DESC, scheduled_at ASC 
LIMIT 10;

-- 测试配置同步状态查询性能
EXPLAIN SELECT * FROM config_sync_status 
WHERE cluster_id = 'cluster-1' 
AND sync_status != 'success'
ORDER BY updated_at DESC;

-- ============================================================================
-- 完成优化
-- ============================================================================

-- 分析表以更新统计信息
ANALYZE TABLE 
    alerts, alert_rules, rule_versions, rule_distribution_records,
    config_sync_status, config_sync_history, config_sync_exceptions,
    task_queue, worker_instances, task_execution_history,
    user_notification_configs, notification_records, notification_plugins,
    knowledges, providers;

-- 显示优化完成信息
SELECT 
    'Database optimization completed!' as message,
    NOW() as completion_time,
    'Indexes created, procedures defined, views created' as summary;

-- 显示表大小统计
SELECT 
    table_name,
    ROUND(((data_length + index_length) / 1024 / 1024), 2) as size_mb,
    table_rows,
    ROUND((index_length / 1024 / 1024), 2) as index_size_mb
FROM information_schema.tables 
WHERE table_schema = 'alert_agent'
AND table_type = 'BASE TABLE'
ORDER BY (data_length + index_length) DESC;