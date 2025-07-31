-- 配置同步监控相关表结构

-- 配置版本表
CREATE TABLE IF NOT EXISTS config_versions (
    id VARCHAR(36) PRIMARY KEY,
    cluster_id VARCHAR(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
    config_type VARCHAR(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
    version VARCHAR(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
    config_hash VARCHAR(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
    config_content LONGTEXT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci,
    config_size BIGINT NOT NULL DEFAULT 0,
    description TEXT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci,
    created_by VARCHAR(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci,
    is_active BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    
    INDEX idx_cluster_config (cluster_id, config_type),
    INDEX idx_config_hash (config_hash),
    INDEX idx_is_active (is_active),
    INDEX idx_created_at (created_at),
    INDEX idx_deleted_at (deleted_at),
    
    UNIQUE KEY uk_cluster_config_version (cluster_id, config_type, version)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='配置版本表';

-- 配置同步异常表
CREATE TABLE IF NOT EXISTS config_sync_exceptions (
    id VARCHAR(36) PRIMARY KEY,
    cluster_id VARCHAR(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
    config_type VARCHAR(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
    exception_type VARCHAR(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
    error_message TEXT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci,
    severity VARCHAR(20) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 'medium',
    status VARCHAR(20) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 'open',
    first_occurred TIMESTAMP NOT NULL,
    last_occurred TIMESTAMP NOT NULL,
    occurrence_count BIGINT NOT NULL DEFAULT 1,
    auto_retry_count INT NOT NULL DEFAULT 0,
    max_auto_retry INT NOT NULL DEFAULT 3,
    next_retry_at TIMESTAMP NULL,
    resolved_at TIMESTAMP NULL,
    resolved_by VARCHAR(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    INDEX idx_cluster_config (cluster_id, config_type),
    INDEX idx_exception_type (exception_type),
    INDEX idx_severity (severity),
    INDEX idx_status (status),
    INDEX idx_first_occurred (first_occurred),
    INDEX idx_next_retry_at (next_retry_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='配置同步异常表';

-- 更新现有的配置同步历史表，添加配置大小字段
ALTER TABLE config_sync_history 
ADD COLUMN IF NOT EXISTS config_size BIGINT NOT NULL DEFAULT 0 COMMENT '配置文件大小（字节）';

-- 添加索引优化查询性能
ALTER TABLE config_sync_history 
ADD INDEX IF NOT EXISTS idx_cluster_config_created (cluster_id, config_type, created_at);

ALTER TABLE config_sync_status 
ADD INDEX IF NOT EXISTS idx_cluster_config_updated (cluster_id, config_type, updated_at);

-- 创建视图：配置同步监控概览
CREATE OR REPLACE VIEW v_config_sync_overview AS
SELECT 
    css.cluster_id,
    css.config_type,
    css.sync_status,
    css.sync_time,
    css.error_message,
    css.config_hash,
    cv.version as active_version,
    cv.created_at as version_created_at,
    CASE 
        WHEN css.sync_time IS NULL THEN -1
        ELSE TIMESTAMPDIFF(SECOND, css.sync_time, NOW())
    END as sync_delay_seconds,
    CASE 
        WHEN css.sync_status = 'success' AND 
             (css.sync_time IS NULL OR TIMESTAMPDIFF(SECOND, css.sync_time, NOW()) <= 300)
        THEN 'healthy'
        WHEN css.sync_status = 'failed' OR 
             TIMESTAMPDIFF(SECOND, css.sync_time, NOW()) > 1800
        THEN 'critical'
        ELSE 'warning'
    END as health_status
FROM config_sync_status css
LEFT JOIN config_versions cv ON css.cluster_id = cv.cluster_id 
    AND css.config_type = cv.config_type 
    AND cv.is_active = TRUE
    AND cv.deleted_at IS NULL;

-- 创建视图：异常统计
CREATE OR REPLACE VIEW v_exception_statistics AS
SELECT 
    cluster_id,
    config_type,
    exception_type,
    severity,
    status,
    COUNT(*) as exception_count,
    SUM(occurrence_count) as total_occurrences,
    MIN(first_occurred) as earliest_occurrence,
    MAX(last_occurred) as latest_occurrence,
    AVG(occurrence_count) as avg_occurrence_count
FROM config_sync_exceptions
GROUP BY cluster_id, config_type, exception_type, severity, status;

-- 插入示例数据（可选）
-- INSERT INTO config_versions (id, cluster_id, config_type, version, config_hash, config_content, config_size, description, created_by, is_active)
-- VALUES 
-- ('example-version-1', 'cluster-1', 'prometheus', 'v1.0.0', 'abc123', 'groups: []', 10, 'Initial version', 'system', TRUE);

-- 创建定期清理任务的存储过程
DELIMITER //

CREATE PROCEDURE IF NOT EXISTS CleanupOldConfigData()
BEGIN
    DECLARE EXIT HANDLER FOR SQLEXCEPTION
    BEGIN
        ROLLBACK;
        RESIGNAL;
    END;

    START TRANSACTION;
    
    -- 清理30天前的同步历史记录
    DELETE FROM config_sync_history 
    WHERE created_at < DATE_SUB(NOW(), INTERVAL 30 DAY);
    
    -- 清理已解决的异常记录（保留90天）
    DELETE FROM config_sync_exceptions 
    WHERE status = 'resolved' 
    AND resolved_at < DATE_SUB(NOW(), INTERVAL 90 DAY);
    
    -- 清理软删除的配置版本（保留7天）
    DELETE FROM config_versions 
    WHERE deleted_at IS NOT NULL 
    AND deleted_at < DATE_SUB(NOW(), INTERVAL 7 DAY);
    
    COMMIT;
END //

DELIMITER ;

-- 创建事件调度器来定期执行清理任务（需要开启事件调度器）
-- SET GLOBAL event_scheduler = ON;
-- 
-- CREATE EVENT IF NOT EXISTS evt_cleanup_config_data
-- ON SCHEDULE EVERY 1 DAY
-- STARTS CURRENT_TIMESTAMP
-- DO
--   CALL CleanupOldConfigData();

COMMIT;