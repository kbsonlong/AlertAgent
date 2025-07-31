-- AlertAgent 重构迁移数据一致性验证脚本
-- 用于验证迁移后的数据完整性和一致性

USE alert_agent;

-- ============================================================================
-- 1. 基础数据统计对比
-- ============================================================================

SELECT '=== BASIC DATA STATISTICS ===' as section;

-- 规则数据对比
SELECT 
    'Rules Migration Check' as check_name,
    (SELECT COUNT(*) FROM rules WHERE deleted_at IS NULL) as original_rules,
    (SELECT COUNT(*) FROM alert_rules WHERE deleted_at IS NULL) as migrated_rules,
    CASE 
        WHEN (SELECT COUNT(*) FROM rules WHERE deleted_at IS NULL) = 
             (SELECT COUNT(*) FROM alert_rules WHERE deleted_at IS NULL) 
        THEN 'PASS' 
        ELSE 'FAIL' 
    END as status,
    CASE 
        WHEN (SELECT COUNT(*) FROM rules WHERE deleted_at IS NULL) = 
             (SELECT COUNT(*) FROM alert_rules WHERE deleted_at IS NULL) 
        THEN 'All rules migrated successfully'
        ELSE CONCAT('Missing rules: ', 
                   (SELECT COUNT(*) FROM rules WHERE deleted_at IS NULL) - 
                   (SELECT COUNT(*) FROM alert_rules WHERE deleted_at IS NULL))
    END as details;

-- 告警数据完整性检查
SELECT 
    'Alerts Integrity Check' as check_name,
    (SELECT COUNT(*) FROM alerts WHERE deleted_at IS NULL) as total_alerts,
    (SELECT COUNT(*) FROM alerts WHERE fingerprint IS NOT NULL AND fingerprint != '') as alerts_with_fingerprint,
    CASE 
        WHEN (SELECT COUNT(*) FROM alerts WHERE deleted_at IS NULL) = 
             (SELECT COUNT(*) FROM alerts WHERE fingerprint IS NOT NULL AND fingerprint != '') 
        THEN 'PASS' 
        ELSE 'PARTIAL' 
    END as status,
    CONCAT(
        ROUND(
            (SELECT COUNT(*) FROM alerts WHERE fingerprint IS NOT NULL AND fingerprint != '') * 100.0 / 
            (SELECT COUNT(*) FROM alerts WHERE deleted_at IS NULL), 2
        ), '% alerts have fingerprints'
    ) as details;

-- ============================================================================
-- 2. 数据质量检查
-- ============================================================================

SELECT '=== DATA QUALITY CHECKS ===' as section;

-- 检查重复的规则名称
SELECT 
    'Duplicate Rule Names' as check_name,
    COUNT(*) as duplicate_count,
    CASE WHEN COUNT(*) = 0 THEN 'PASS' ELSE 'FAIL' END as status,
    CASE WHEN COUNT(*) = 0 THEN 'No duplicate rule names found' 
         ELSE CONCAT(COUNT(*), ' duplicate rule names found') END as details
FROM (
    SELECT name, COUNT(*) as cnt
    FROM alert_rules 
    WHERE deleted_at IS NULL
    GROUP BY name
    HAVING cnt > 1
) duplicates;

-- 检查无效的规则表达式
SELECT 
    'Invalid Rule Expressions' as check_name,
    COUNT(*) as invalid_count,
    CASE WHEN COUNT(*) = 0 THEN 'PASS' ELSE 'WARNING' END as status,
    CASE WHEN COUNT(*) = 0 THEN 'All rule expressions are valid' 
         ELSE CONCAT(COUNT(*), ' rules have empty expressions') END as details
FROM alert_rules 
WHERE deleted_at IS NULL AND (expression IS NULL OR expression = '');

-- 检查无效的严重程度
SELECT 
    'Invalid Severity Levels' as check_name,
    COUNT(*) as invalid_count,
    CASE WHEN COUNT(*) = 0 THEN 'PASS' ELSE 'FAIL' END as status,
    CASE WHEN COUNT(*) = 0 THEN 'All severity levels are valid' 
         ELSE CONCAT(COUNT(*), ' rules have invalid severity levels') END as details
FROM alert_rules 
WHERE deleted_at IS NULL 
AND severity NOT IN ('critical', 'high', 'medium', 'low', 'warning', 'info');

-- ============================================================================
-- 3. JSON字段验证
-- ============================================================================

SELECT '=== JSON FIELDS VALIDATION ===' as section;

-- 验证规则的JSON字段
SELECT 
    'Rule JSON Fields' as check_name,
    SUM(CASE WHEN JSON_VALID(labels) = 0 THEN 1 ELSE 0 END) as invalid_labels,
    SUM(CASE WHEN JSON_VALID(annotations) = 0 THEN 1 ELSE 0 END) as invalid_annotations,
    SUM(CASE WHEN JSON_VALID(targets) = 0 THEN 1 ELSE 0 END) as invalid_targets,
    CASE 
        WHEN SUM(CASE WHEN JSON_VALID(labels) = 0 THEN 1 ELSE 0 END) = 0 AND
             SUM(CASE WHEN JSON_VALID(annotations) = 0 THEN 1 ELSE 0 END) = 0 AND
             SUM(CASE WHEN JSON_VALID(targets) = 0 THEN 1 ELSE 0 END) = 0
        THEN 'PASS' 
        ELSE 'FAIL' 
    END as status
FROM alert_rules 
WHERE deleted_at IS NULL;

-- 验证告警的JSON字段
SELECT 
    'Alert JSON Fields' as check_name,
    SUM(CASE WHEN analysis_result IS NOT NULL AND JSON_VALID(analysis_result) = 0 THEN 1 ELSE 0 END) as invalid_analysis_result,
    SUM(CASE WHEN similar_alerts IS NOT NULL AND JSON_VALID(similar_alerts) = 0 THEN 1 ELSE 0 END) as invalid_similar_alerts,
    CASE 
        WHEN SUM(CASE WHEN analysis_result IS NOT NULL AND JSON_VALID(analysis_result) = 0 THEN 1 ELSE 0 END) = 0 AND
             SUM(CASE WHEN similar_alerts IS NOT NULL AND JSON_VALID(similar_alerts) = 0 THEN 1 ELSE 0 END) = 0
        THEN 'PASS' 
        ELSE 'FAIL' 
    END as status
FROM alerts 
WHERE deleted_at IS NULL;

-- ============================================================================
-- 4. 外键关系验证
-- ============================================================================

SELECT '=== FOREIGN KEY RELATIONSHIPS ===' as section;

-- 检查规则版本的外键关系
SELECT 
    'Rule Version References' as check_name,
    COUNT(*) as orphaned_versions,
    CASE WHEN COUNT(*) = 0 THEN 'PASS' ELSE 'FAIL' END as status,
    CASE WHEN COUNT(*) = 0 THEN 'All rule versions have valid references' 
         ELSE CONCAT(COUNT(*), ' orphaned rule versions found') END as details
FROM rule_versions rv
LEFT JOIN alert_rules ar ON rv.rule_id = ar.id
WHERE ar.id IS NULL;

-- 检查规则分发记录的外键关系
SELECT 
    'Rule Distribution References' as check_name,
    COUNT(*) as orphaned_distributions,
    CASE WHEN COUNT(*) = 0 THEN 'PASS' ELSE 'FAIL' END as status,
    CASE WHEN COUNT(*) = 0 THEN 'All distribution records have valid references' 
         ELSE CONCAT(COUNT(*), ' orphaned distribution records found') END as details
FROM rule_distribution_records rdr
LEFT JOIN alert_rules ar ON rdr.rule_id = ar.id
WHERE ar.id IS NULL;

-- ============================================================================
-- 5. 配置同步状态验证
-- ============================================================================

SELECT '=== CONFIG SYNC STATUS VALIDATION ===' as section;

-- 检查配置同步状态的完整性
SELECT 
    'Config Sync Status' as check_name,
    COUNT(*) as total_records,
    COUNT(DISTINCT cluster_id) as unique_clusters,
    COUNT(DISTINCT config_type) as unique_config_types,
    CASE WHEN COUNT(*) > 0 THEN 'PASS' ELSE 'WARNING' END as status,
    CONCAT(COUNT(*), ' config sync records for ', COUNT(DISTINCT cluster_id), ' clusters') as details
FROM config_sync_status 
WHERE deleted_at IS NULL;

-- 检查配置类型的有效性
SELECT 
    'Config Type Validity' as check_name,
    COUNT(*) as invalid_types,
    CASE WHEN COUNT(*) = 0 THEN 'PASS' ELSE 'FAIL' END as status,
    CASE WHEN COUNT(*) = 0 THEN 'All config types are valid' 
         ELSE CONCAT(COUNT(*), ' invalid config types found') END as details
FROM config_sync_status 
WHERE deleted_at IS NULL 
AND config_type NOT IN ('prometheus', 'alertmanager', 'vmalert');

-- ============================================================================
-- 6. 通知系统验证
-- ============================================================================

SELECT '=== NOTIFICATION SYSTEM VALIDATION ===' as section;

-- 检查通知插件配置
SELECT 
    'Notification Plugins' as check_name,
    COUNT(*) as total_plugins,
    SUM(CASE WHEN enabled = 1 THEN 1 ELSE 0 END) as enabled_plugins,
    CASE WHEN COUNT(*) > 0 THEN 'PASS' ELSE 'WARNING' END as status,
    CONCAT(COUNT(*), ' plugins configured, ', SUM(CASE WHEN enabled = 1 THEN 1 ELSE 0 END), ' enabled') as details
FROM notification_plugins 
WHERE deleted_at IS NULL;

-- 验证插件配置的JSON格式
SELECT 
    'Plugin Config JSON' as check_name,
    SUM(CASE WHEN JSON_VALID(config) = 0 THEN 1 ELSE 0 END) as invalid_configs,
    CASE 
        WHEN SUM(CASE WHEN JSON_VALID(config) = 0 THEN 1 ELSE 0 END) = 0 
        THEN 'PASS' 
        ELSE 'FAIL' 
    END as status,
    CASE 
        WHEN SUM(CASE WHEN JSON_VALID(config) = 0 THEN 1 ELSE 0 END) = 0 
        THEN 'All plugin configs are valid JSON'
        ELSE CONCAT(SUM(CASE WHEN JSON_VALID(config) = 0 THEN 1 ELSE 0 END), ' invalid JSON configs found')
    END as details
FROM notification_plugins 
WHERE deleted_at IS NULL;

-- ============================================================================
-- 7. 索引和性能验证
-- ============================================================================

SELECT '=== INDEX AND PERFORMANCE VALIDATION ===' as section;

-- 检查关键索引是否存在
SELECT 
    'Critical Indexes' as check_name,
    COUNT(*) as missing_indexes,
    CASE WHEN COUNT(*) = 0 THEN 'PASS' ELSE 'WARNING' END as status,
    CASE WHEN COUNT(*) = 0 THEN 'All critical indexes exist' 
         ELSE CONCAT(COUNT(*), ' critical indexes missing') END as details
FROM (
    SELECT 'alert_rules.idx_status' as index_name
    WHERE NOT EXISTS (
        SELECT 1 FROM information_schema.statistics 
        WHERE table_schema = 'alert_agent' 
        AND table_name = 'alert_rules' 
        AND index_name = 'idx_status'
    )
    UNION ALL
    SELECT 'alerts.idx_analysis_status' as index_name
    WHERE NOT EXISTS (
        SELECT 1 FROM information_schema.statistics 
        WHERE table_schema = 'alert_agent' 
        AND table_name = 'alerts' 
        AND index_name = 'idx_analysis_status'
    )
    UNION ALL
    SELECT 'config_sync_status.idx_cluster_type' as index_name
    WHERE NOT EXISTS (
        SELECT 1 FROM information_schema.statistics 
        WHERE table_schema = 'alert_agent' 
        AND table_name = 'config_sync_status' 
        AND index_name = 'idx_cluster_type'
    )
) missing_idx;

-- ============================================================================
-- 8. 数据分布统计
-- ============================================================================

SELECT '=== DATA DISTRIBUTION STATISTICS ===' as section;

-- 规则严重程度分布
SELECT 
    'Rule Severity Distribution' as metric_name,
    severity,
    COUNT(*) as count,
    ROUND(COUNT(*) * 100.0 / (SELECT COUNT(*) FROM alert_rules WHERE deleted_at IS NULL), 2) as percentage
FROM alert_rules 
WHERE deleted_at IS NULL
GROUP BY severity
ORDER BY count DESC;

-- 告警状态分布
SELECT 
    'Alert Status Distribution' as metric_name,
    status,
    COUNT(*) as count,
    ROUND(COUNT(*) * 100.0 / (SELECT COUNT(*) FROM alerts WHERE deleted_at IS NULL), 2) as percentage
FROM alerts 
WHERE deleted_at IS NULL
GROUP BY status
ORDER BY count DESC;

-- 分析状态分布
SELECT 
    'Analysis Status Distribution' as metric_name,
    analysis_status,
    COUNT(*) as count,
    ROUND(COUNT(*) * 100.0 / (SELECT COUNT(*) FROM alerts WHERE deleted_at IS NULL), 2) as percentage
FROM alerts 
WHERE deleted_at IS NULL
GROUP BY analysis_status
ORDER BY count DESC;

-- ============================================================================
-- 9. 迁移日志摘要
-- ============================================================================

SELECT '=== MIGRATION LOG SUMMARY ===' as section;

SELECT 
    migration_name,
    status,
    start_time,
    end_time,
    TIMESTAMPDIFF(SECOND, start_time, end_time) as duration_seconds,
    records_processed,
    records_migrated,
    records_failed,
    CASE 
        WHEN records_processed > 0 
        THEN ROUND(records_migrated * 100.0 / records_processed, 2)
        ELSE 0 
    END as success_rate_percent
FROM migration_log 
WHERE migration_name LIKE '%migration%'
ORDER BY start_time DESC;

-- ============================================================================
-- 10. 总体验证结果
-- ============================================================================

SELECT '=== OVERALL VALIDATION RESULT ===' as section;

-- 计算总体验证分数
SELECT 
    'Overall Migration Health' as check_name,
    CASE 
        WHEN (
            -- 规则迁移完整性
            (SELECT COUNT(*) FROM rules WHERE deleted_at IS NULL) = 
            (SELECT COUNT(*) FROM alert_rules WHERE deleted_at IS NULL)
            AND
            -- 告警指纹完整性 >= 95%
            (SELECT COUNT(*) FROM alerts WHERE fingerprint IS NOT NULL AND fingerprint != '') * 100.0 / 
            (SELECT COUNT(*) FROM alerts WHERE deleted_at IS NULL) >= 95
            AND
            -- 无重复规则名称
            (SELECT COUNT(*) FROM (
                SELECT name FROM alert_rules WHERE deleted_at IS NULL GROUP BY name HAVING COUNT(*) > 1
            ) dup) = 0
            AND
            -- JSON字段有效性
            (SELECT SUM(CASE WHEN JSON_VALID(labels) = 0 OR JSON_VALID(annotations) = 0 OR JSON_VALID(targets) = 0 THEN 1 ELSE 0 END) 
             FROM alert_rules WHERE deleted_at IS NULL) = 0
        ) THEN 'EXCELLENT'
        WHEN (
            (SELECT COUNT(*) FROM rules WHERE deleted_at IS NULL) = 
            (SELECT COUNT(*) FROM alert_rules WHERE deleted_at IS NULL)
            AND
            (SELECT COUNT(*) FROM alerts WHERE fingerprint IS NOT NULL AND fingerprint != '') * 100.0 / 
            (SELECT COUNT(*) FROM alerts WHERE deleted_at IS NULL) >= 80
        ) THEN 'GOOD'
        ELSE 'NEEDS_ATTENTION'
    END as overall_status,
    CONCAT(
        'Rules: ', (SELECT COUNT(*) FROM alert_rules WHERE deleted_at IS NULL), ', ',
        'Alerts: ', (SELECT COUNT(*) FROM alerts WHERE deleted_at IS NULL), ', ',
        'Fingerprints: ', ROUND((SELECT COUNT(*) FROM alerts WHERE fingerprint IS NOT NULL AND fingerprint != '') * 100.0 / 
                                (SELECT COUNT(*) FROM alerts WHERE deleted_at IS NULL), 1), '%'
    ) as summary;

-- 显示验证完成时间
SELECT 
    'Validation completed at:' as message,
    NOW() as timestamp;