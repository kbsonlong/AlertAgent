-- 规则版本控制相关表的数据库迁移脚本

-- 创建规则版本表
CREATE TABLE IF NOT EXISTS `rule_versions` (
    `id` VARCHAR(36) NOT NULL PRIMARY KEY COMMENT '版本记录ID',
    `rule_id` VARCHAR(36) NOT NULL COMMENT '规则ID',
    `version` VARCHAR(50) NOT NULL COMMENT '版本号',
    `name` VARCHAR(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '规则名称',
    `expression` TEXT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '规则表达式',
    `duration` VARCHAR(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '持续时间',
    `severity` VARCHAR(20) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '严重程度',
    `labels` JSON COMMENT '标签',
    `annotations` JSON COMMENT '注释',
    `targets` JSON COMMENT '目标系统',
    `status` VARCHAR(20) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '状态',
    `change_type` VARCHAR(20) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '变更类型',
    `changed_by` VARCHAR(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci COMMENT '变更人',
    `change_note` TEXT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci COMMENT '变更说明',
    `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `deleted_at` TIMESTAMP NULL DEFAULT NULL COMMENT '删除时间',
    
    INDEX `idx_rule_id` (`rule_id`),
    INDEX `idx_rule_version` (`rule_id`, `version`),
    INDEX `idx_change_type` (`change_type`),
    INDEX `idx_changed_by` (`changed_by`),
    INDEX `idx_created_at` (`created_at`),
    INDEX `idx_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='规则版本记录表';

-- 创建规则审计日志表
CREATE TABLE IF NOT EXISTS `rule_audit_logs` (
    `id` VARCHAR(36) NOT NULL PRIMARY KEY COMMENT '审计日志ID',
    `rule_id` VARCHAR(36) NOT NULL COMMENT '规则ID',
    `action` VARCHAR(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '操作类型',
    `old_version` VARCHAR(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci COMMENT '旧版本号',
    `new_version` VARCHAR(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci COMMENT '新版本号',
    `changes` JSON COMMENT '变更详情',
    `user_id` VARCHAR(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci COMMENT '用户ID',
    `user_name` VARCHAR(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci COMMENT '用户名',
    `ip_address` VARCHAR(45) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci COMMENT 'IP地址',
    `user_agent` TEXT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci COMMENT '用户代理',
    `note` TEXT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci COMMENT '备注',
    `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `deleted_at` TIMESTAMP NULL DEFAULT NULL COMMENT '删除时间',
    
    INDEX `idx_rule_id` (`rule_id`),
    INDEX `idx_action` (`action`),
    INDEX `idx_user_id` (`user_id`),
    INDEX `idx_created_at` (`created_at`),
    INDEX `idx_deleted_at` (`deleted_at`),
    INDEX `idx_rule_action` (`rule_id`, `action`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='规则审计日志表';

-- 为现有的 alert_rules 表添加外键约束（可选，根据实际需要决定是否添加）
-- ALTER TABLE `rule_versions` ADD CONSTRAINT `fk_rule_versions_rule_id` 
--     FOREIGN KEY (`rule_id`) REFERENCES `alert_rules`(`id`) ON DELETE CASCADE;

-- ALTER TABLE `rule_audit_logs` ADD CONSTRAINT `fk_rule_audit_logs_rule_id` 
--     FOREIGN KEY (`rule_id`) REFERENCES `alert_rules`(`id`) ON DELETE CASCADE;

-- 创建一些有用的视图

-- 规则版本统计视图
CREATE OR REPLACE VIEW `rule_version_stats` AS
SELECT 
    r.id as rule_id,
    r.name as rule_name,
    r.version as current_version,
    r.status as current_status,
    COUNT(rv.id) as version_count,
    MAX(rv.created_at) as last_version_time,
    rv_latest.changed_by as last_changed_by
FROM alert_rules r
LEFT JOIN rule_versions rv ON r.id = rv.rule_id
LEFT JOIN rule_versions rv_latest ON r.id = rv_latest.rule_id 
    AND rv_latest.created_at = (
        SELECT MAX(created_at) 
        FROM rule_versions 
        WHERE rule_id = r.id AND deleted_at IS NULL
    )
WHERE r.deleted_at IS NULL
GROUP BY r.id, r.name, r.version, r.status, rv_latest.changed_by;

-- 规则审计活动视图
CREATE OR REPLACE VIEW `rule_audit_activity` AS
SELECT 
    ral.id,
    ral.rule_id,
    r.name as rule_name,
    ral.action,
    ral.old_version,
    ral.new_version,
    ral.user_name,
    ral.created_at,
    ral.note
FROM rule_audit_logs ral
LEFT JOIN alert_rules r ON ral.rule_id = r.id
WHERE ral.deleted_at IS NULL
ORDER BY ral.created_at DESC;

-- 插入一些示例数据（可选，用于测试）
-- 注意：在生产环境中请删除这些示例数据

-- INSERT INTO rule_versions (id, rule_id, version, name, expression, duration, severity, labels, annotations, targets, status, change_type, changed_by, change_note)
-- VALUES 
-- ('version-1', 'rule-1', 'v1.0.0', 'Test Rule', 'up == 0', '5m', 'critical', '{}', '{}', '["prometheus"]', 'active', 'create', 'admin', 'Initial version'),
-- ('version-2', 'rule-1', 'v1.0.1', 'Test Rule Updated', 'up == 0', '3m', 'critical', '{}', '{}', '["prometheus"]', 'active', 'update', 'admin', 'Updated duration');

-- INSERT INTO rule_audit_logs (id, rule_id, action, old_version, new_version, changes, user_id, user_name, ip_address, note)
-- VALUES 
-- ('audit-1', 'rule-1', 'create', '', 'v1.0.0', '{"name": "Test Rule", "expression": "up == 0"}', 'admin', 'Administrator', '127.0.0.1', 'Rule created'),
-- ('audit-2', 'rule-1', 'update', 'v1.0.0', 'v1.0.1', '{"duration": {"old": "5m", "new": "3m"}}', 'admin', 'Administrator', '127.0.0.1', 'Updated duration');