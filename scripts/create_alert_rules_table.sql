-- 创建新的 alert_rules 表以支持重构后的规则管理
USE alert_agent;

-- 创建新的告警规则表
CREATE TABLE IF NOT EXISTS `alert_rules` (
    `id` VARCHAR(36) NOT NULL PRIMARY KEY COMMENT '规则ID',
    `name` VARCHAR(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '规则名称',
    `expression` TEXT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '规则表达式',
    `duration` VARCHAR(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '持续时间',
    `severity` VARCHAR(20) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '严重程度',
    `labels` JSON COMMENT '标签',
    `annotations` JSON COMMENT '注释',
    `targets` JSON COMMENT '目标系统',
    `version` VARCHAR(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 'v1.0.0' COMMENT '版本号',
    `status` VARCHAR(20) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 'pending' COMMENT '状态',
    `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    `deleted_at` TIMESTAMP NULL DEFAULT NULL COMMENT '删除时间',
    
    INDEX `idx_name` (`name`),
    INDEX `idx_status` (`status`),
    INDEX `idx_created_at` (`created_at`),
    INDEX `idx_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='告警规则表（重构版）';

-- 可选：从旧的 rules 表迁移数据到新的 alert_rules 表
-- 注意：这个迁移脚本需要根据实际的数据结构调整

-- INSERT INTO alert_rules (id, name, expression, duration, severity, labels, annotations, targets, version, status, created_at, updated_at)
-- SELECT 
--     UUID() as id,
--     name,
--     condition_expr as expression,
--     '5m' as duration,
--     CASE 
--         WHEN level = 'critical' THEN 'critical'
--         WHEN level = 'warning' THEN 'warning'
--         ELSE 'info'
--     END as severity,
--     '{}' as labels,
--     '{}' as annotations,
--     '["prometheus"]' as targets,
--     'v1.0.0' as version,
--     CASE 
--         WHEN enabled = 1 THEN 'active'
--         ELSE 'inactive'
--     END as status,
--     created_at,
--     updated_at
-- FROM rules
-- WHERE deleted_at IS NULL;