-- AlertAgent 重构数据库表设计
-- 创建重构后系统所需的所有新表和扩展现有表

USE alert_agent;

-- ============================================================================
-- 1. 告警规则管理相关表
-- ============================================================================

-- 创建新的告警规则表（如果不存在）
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
    INDEX `idx_severity` (`severity`),
    INDEX `idx_version` (`version`),
    INDEX `idx_created_at` (`created_at`),
    INDEX `idx_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='告警规则表（重构版）';

-- 规则版本历史表
CREATE TABLE IF NOT EXISTS `rule_versions` (
    `id` VARCHAR(36) NOT NULL PRIMARY KEY,
    `rule_id` VARCHAR(36) NOT NULL COMMENT '规则ID',
    `version` VARCHAR(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '版本号',
    `name` VARCHAR(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '规则名称',
    `expression` TEXT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '规则表达式',
    `duration` VARCHAR(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '持续时间',
    `severity` VARCHAR(20) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '严重程度',
    `labels` JSON COMMENT '标签',
    `annotations` JSON COMMENT '注释',
    `targets` JSON COMMENT '目标系统',
    `change_log` TEXT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci COMMENT '变更日志',
    `created_by` VARCHAR(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci COMMENT '创建者',
    `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    
    INDEX `idx_rule_id` (`rule_id`),
    INDEX `idx_version` (`version`),
    INDEX `idx_created_at` (`created_at`),
    FOREIGN KEY (`rule_id`) REFERENCES `alert_rules`(`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='规则版本历史表';

-- 规则分发记录表
CREATE TABLE IF NOT EXISTS `rule_distribution_records` (
    `id` VARCHAR(36) PRIMARY KEY,
    `rule_id` VARCHAR(36) NOT NULL COMMENT '规则ID',
    `target` VARCHAR(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '目标系统',
    `status` VARCHAR(20) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 'pending' COMMENT '分发状态',
    `version` VARCHAR(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '版本号',
    `config_hash` VARCHAR(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci COMMENT '配置哈希',
    `last_sync` TIMESTAMP NULL COMMENT '最后同步时间',
    `error` TEXT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci COMMENT '错误信息',
    `retry_count` INT DEFAULT 0 COMMENT '重试次数',
    `max_retry` INT DEFAULT 3 COMMENT '最大重试次数',
    `next_retry` TIMESTAMP NULL COMMENT '下次重试时间',
    `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    `deleted_at` TIMESTAMP NULL COMMENT '删除时间',
    
    INDEX `idx_rule_id` (`rule_id`),
    INDEX `idx_target` (`target`),
    INDEX `idx_status` (`status`),
    INDEX `idx_next_retry` (`next_retry`),
    INDEX `idx_deleted_at` (`deleted_at`),
    UNIQUE KEY `uk_rule_target` (`rule_id`, `target`, `deleted_at`),
    FOREIGN KEY (`rule_id`) REFERENCES `alert_rules`(`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='规则分发记录表';

-- ============================================================================
-- 2. 配置同步相关表
-- ============================================================================

-- 配置同步状态表
CREATE TABLE IF NOT EXISTS `config_sync_status` (
    `id` VARCHAR(36) PRIMARY KEY,
    `cluster_id` VARCHAR(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '集群ID',
    `config_type` VARCHAR(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '配置类型',
    `config_hash` VARCHAR(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci COMMENT '配置哈希',
    `sync_status` VARCHAR(20) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 'pending' COMMENT '同步状态',
    `sync_time` TIMESTAMP NULL COMMENT '同步时间',
    `error_message` TEXT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci COMMENT '错误信息',
    `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    `deleted_at` TIMESTAMP NULL COMMENT '删除时间',
    
    INDEX `idx_cluster_type` (`cluster_id`, `config_type`),
    INDEX `idx_sync_status` (`sync_status`),
    INDEX `idx_sync_time` (`sync_time`),
    INDEX `idx_deleted_at` (`deleted_at`),
    UNIQUE KEY `uk_cluster_config` (`cluster_id`, `config_type`, `deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='配置同步状态表';

-- 配置版本表
CREATE TABLE IF NOT EXISTS `config_versions` (
    `id` VARCHAR(36) PRIMARY KEY,
    `cluster_id` VARCHAR(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '集群ID',
    `config_type` VARCHAR(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '配置类型',
    `version` VARCHAR(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '版本号',
    `config_hash` VARCHAR(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '配置哈希',
    `config_content` LONGTEXT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci COMMENT '配置内容',
    `config_size` BIGINT NOT NULL DEFAULT 0 COMMENT '配置大小',
    `description` TEXT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci COMMENT '描述',
    `created_by` VARCHAR(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci COMMENT '创建者',
    `is_active` BOOLEAN NOT NULL DEFAULT FALSE COMMENT '是否激活',
    `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    `deleted_at` TIMESTAMP NULL COMMENT '删除时间',
    
    INDEX `idx_cluster_config` (`cluster_id`, `config_type`),
    INDEX `idx_config_hash` (`config_hash`),
    INDEX `idx_is_active` (`is_active`),
    INDEX `idx_created_at` (`created_at`),
    INDEX `idx_deleted_at` (`deleted_at`),
    UNIQUE KEY `uk_cluster_config_version` (`cluster_id`, `config_type`, `version`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='配置版本表';

-- 配置同步历史表
CREATE TABLE IF NOT EXISTS `config_sync_history` (
    `id` VARCHAR(36) PRIMARY KEY,
    `cluster_id` VARCHAR(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '集群ID',
    `config_type` VARCHAR(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '配置类型',
    `config_hash` VARCHAR(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '配置哈希',
    `config_size` BIGINT NOT NULL DEFAULT 0 COMMENT '配置大小',
    `sync_status` VARCHAR(20) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '同步状态',
    `sync_duration` BIGINT NOT NULL DEFAULT 0 COMMENT '同步耗时（毫秒）',
    `error_message` TEXT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci COMMENT '错误信息',
    `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `deleted_at` TIMESTAMP NULL COMMENT '删除时间',
    
    INDEX `idx_cluster_type` (`cluster_id`, `config_type`),
    INDEX `idx_sync_status` (`sync_status`),
    INDEX `idx_config_hash` (`config_hash`),
    INDEX `idx_created_at` (`created_at`),
    INDEX `idx_cluster_config_created` (`cluster_id`, `config_type`, `created_at`),
    INDEX `idx_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='配置同步历史表';

-- 配置同步异常表
CREATE TABLE IF NOT EXISTS `config_sync_exceptions` (
    `id` VARCHAR(36) PRIMARY KEY,
    `cluster_id` VARCHAR(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '集群ID',
    `config_type` VARCHAR(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '配置类型',
    `exception_type` VARCHAR(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '异常类型',
    `error_message` TEXT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci COMMENT '错误信息',
    `severity` VARCHAR(20) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 'medium' COMMENT '严重程度',
    `status` VARCHAR(20) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 'open' COMMENT '状态',
    `first_occurred` TIMESTAMP NOT NULL COMMENT '首次发生时间',
    `last_occurred` TIMESTAMP NOT NULL COMMENT '最后发生时间',
    `occurrence_count` BIGINT NOT NULL DEFAULT 1 COMMENT '发生次数',
    `auto_retry_count` INT NOT NULL DEFAULT 0 COMMENT '自动重试次数',
    `max_auto_retry` INT NOT NULL DEFAULT 3 COMMENT '最大自动重试次数',
    `next_retry_at` TIMESTAMP NULL COMMENT '下次重试时间',
    `resolved_at` TIMESTAMP NULL COMMENT '解决时间',
    `resolved_by` VARCHAR(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci COMMENT '解决者',
    `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    
    INDEX `idx_cluster_config` (`cluster_id`, `config_type`),
    INDEX `idx_exception_type` (`exception_type`),
    INDEX `idx_severity` (`severity`),
    INDEX `idx_status` (`status`),
    INDEX `idx_first_occurred` (`first_occurred`),
    INDEX `idx_next_retry_at` (`next_retry_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='配置同步异常表';

-- ============================================================================
-- 3. 异步任务队列相关表
-- ============================================================================

-- 任务队列表
CREATE TABLE IF NOT EXISTS `task_queue` (
    `id` VARCHAR(36) PRIMARY KEY,
    `queue_name` VARCHAR(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '队列名称',
    `task_type` VARCHAR(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '任务类型',
    `payload` JSON NOT NULL COMMENT '任务载荷',
    `priority` INT DEFAULT 0 COMMENT '优先级',
    `retry_count` INT DEFAULT 0 COMMENT '重试次数',
    `max_retry` INT DEFAULT 3 COMMENT '最大重试次数',
    `status` VARCHAR(20) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 'pending' COMMENT '状态',
    `scheduled_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '调度时间',
    `started_at` TIMESTAMP NULL COMMENT '开始时间',
    `completed_at` TIMESTAMP NULL COMMENT '完成时间',
    `error_message` TEXT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci COMMENT '错误信息',
    `worker_id` VARCHAR(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci COMMENT 'Worker ID',
    `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    
    INDEX `idx_queue_status` (`queue_name`, `status`),
    INDEX `idx_task_type` (`task_type`),
    INDEX `idx_priority` (`priority`),
    INDEX `idx_scheduled_at` (`scheduled_at`),
    INDEX `idx_status_scheduled` (`status`, `scheduled_at`),
    INDEX `idx_worker_id` (`worker_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='任务队列表';

-- Worker实例表
CREATE TABLE IF NOT EXISTS `worker_instances` (
    `id` VARCHAR(36) PRIMARY KEY,
    `worker_id` VARCHAR(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL UNIQUE COMMENT 'Worker ID',
    `queue_name` VARCHAR(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '队列名称',
    `host_name` VARCHAR(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '主机名',
    `process_id` INT NOT NULL COMMENT '进程ID',
    `concurrency` INT NOT NULL DEFAULT 1 COMMENT '并发数',
    `status` VARCHAR(20) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 'active' COMMENT '状态',
    `last_heartbeat` TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '最后心跳时间',
    `tasks_processed` BIGINT DEFAULT 0 COMMENT '已处理任务数',
    `tasks_failed` BIGINT DEFAULT 0 COMMENT '失败任务数',
    `started_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '启动时间',
    `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    
    INDEX `idx_worker_id` (`worker_id`),
    INDEX `idx_queue_name` (`queue_name`),
    INDEX `idx_status` (`status`),
    INDEX `idx_last_heartbeat` (`last_heartbeat`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='Worker实例表';

-- 任务执行历史表
CREATE TABLE IF NOT EXISTS `task_execution_history` (
    `id` VARCHAR(36) PRIMARY KEY,
    `task_id` VARCHAR(36) NOT NULL COMMENT '任务ID',
    `worker_id` VARCHAR(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'Worker ID',
    `queue_name` VARCHAR(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '队列名称',
    `task_type` VARCHAR(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '任务类型',
    `status` VARCHAR(20) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '执行状态',
    `started_at` TIMESTAMP NOT NULL COMMENT '开始时间',
    `completed_at` TIMESTAMP NULL COMMENT '完成时间',
    `duration_ms` BIGINT DEFAULT 0 COMMENT '执行时长（毫秒）',
    `error_message` TEXT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci COMMENT '错误信息',
    `result` JSON COMMENT '执行结果',
    `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    
    INDEX `idx_task_id` (`task_id`),
    INDEX `idx_worker_id` (`worker_id`),
    INDEX `idx_queue_name` (`queue_name`),
    INDEX `idx_task_type` (`task_type`),
    INDEX `idx_status` (`status`),
    INDEX `idx_started_at` (`started_at`),
    FOREIGN KEY (`task_id`) REFERENCES `task_queue`(`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='任务执行历史表';

-- ============================================================================
-- 4. 通知插件系统相关表
-- ============================================================================

-- 通知插件配置表（如果不存在）
CREATE TABLE IF NOT EXISTS `notification_plugins` (
    `id` INT AUTO_INCREMENT PRIMARY KEY,
    `name` VARCHAR(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL UNIQUE COMMENT '插件名称',
    `display_name` VARCHAR(200) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '插件显示名称',
    `version` VARCHAR(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '插件版本',
    `config` JSON NOT NULL COMMENT '插件配置JSON',
    `enabled` BOOLEAN DEFAULT FALSE COMMENT '是否启用',
    `priority` INT DEFAULT 0 COMMENT '发送优先级',
    `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    `deleted_at` TIMESTAMP NULL COMMENT '删除时间',
    
    INDEX `idx_name` (`name`),
    INDEX `idx_enabled` (`enabled`),
    INDEX `idx_priority` (`priority`),
    INDEX `idx_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='通知插件配置表';

-- 用户通知配置表
CREATE TABLE IF NOT EXISTS `user_notification_configs` (
    `id` VARCHAR(36) PRIMARY KEY,
    `user_id` VARCHAR(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '用户ID',
    `plugin_name` VARCHAR(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '插件名称',
    `config` JSON NOT NULL COMMENT '用户配置',
    `enabled` BOOLEAN DEFAULT TRUE COMMENT '是否启用',
    `alert_levels` JSON COMMENT '告警级别过滤',
    `time_windows` JSON COMMENT '时间窗口配置',
    `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    `deleted_at` TIMESTAMP NULL COMMENT '删除时间',
    
    INDEX `idx_user_id` (`user_id`),
    INDEX `idx_plugin_name` (`plugin_name`),
    INDEX `idx_enabled` (`enabled`),
    INDEX `idx_deleted_at` (`deleted_at`),
    UNIQUE KEY `uk_user_plugin` (`user_id`, `plugin_name`, `deleted_at`),
    FOREIGN KEY (`plugin_name`) REFERENCES `notification_plugins`(`name`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户通知配置表';

-- 通知发送记录表
CREATE TABLE IF NOT EXISTS `notification_records` (
    `id` VARCHAR(36) PRIMARY KEY,
    `alert_id` BIGINT UNSIGNED COMMENT '告警ID',
    `user_id` VARCHAR(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci COMMENT '用户ID',
    `plugin_name` VARCHAR(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '插件名称',
    `channel` VARCHAR(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '通知渠道',
    `message_title` VARCHAR(500) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci COMMENT '消息标题',
    `message_content` TEXT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci COMMENT '消息内容',
    `status` VARCHAR(20) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 'pending' COMMENT '发送状态',
    `response` TEXT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci COMMENT '响应信息',
    `retry_count` INT DEFAULT 0 COMMENT '重试次数',
    `error_message` TEXT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci COMMENT '错误信息',
    `sent_at` TIMESTAMP NULL COMMENT '发送时间',
    `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    
    INDEX `idx_alert_id` (`alert_id`),
    INDEX `idx_user_id` (`user_id`),
    INDEX `idx_plugin_name` (`plugin_name`),
    INDEX `idx_status` (`status`),
    INDEX `idx_sent_at` (`sent_at`),
    INDEX `idx_created_at` (`created_at`),
    FOREIGN KEY (`alert_id`) REFERENCES `alerts`(`id`) ON DELETE SET NULL,
    FOREIGN KEY (`plugin_name`) REFERENCES `notification_plugins`(`name`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='通知发送记录表';

-- ============================================================================
-- 5. 扩展现有告警表
-- ============================================================================

-- 为现有alerts表添加分析状态等字段
ALTER TABLE `alerts` 
ADD COLUMN IF NOT EXISTS `analysis_status` VARCHAR(20) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT 'pending' COMMENT '分析状态' AFTER `analysis`,
ADD COLUMN IF NOT EXISTS `analysis_result` JSON COMMENT '分析结果' AFTER `analysis_status`,
ADD COLUMN IF NOT EXISTS `ai_summary` TEXT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci COMMENT 'AI摘要' AFTER `analysis_result`,
ADD COLUMN IF NOT EXISTS `similar_alerts` JSON COMMENT '相似告警' AFTER `ai_summary`,
ADD COLUMN IF NOT EXISTS `resolution_suggestion` TEXT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci COMMENT '解决建议' AFTER `similar_alerts`,
ADD COLUMN IF NOT EXISTS `fingerprint` VARCHAR(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci COMMENT '告警指纹' AFTER `resolution_suggestion`;

-- 添加新的索引
ALTER TABLE `alerts` 
ADD INDEX IF NOT EXISTS `idx_analysis_status` (`analysis_status`),
ADD INDEX IF NOT EXISTS `idx_fingerprint` (`fingerprint`),
ADD INDEX IF NOT EXISTS `idx_severity` (`severity`);

-- ============================================================================
-- 6. 创建视图和存储过程
-- ============================================================================

-- 配置同步监控概览视图
CREATE OR REPLACE VIEW `v_config_sync_overview` AS
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

-- 任务队列统计视图
CREATE OR REPLACE VIEW `v_task_queue_stats` AS
SELECT 
    queue_name,
    task_type,
    status,
    COUNT(*) as task_count,
    AVG(CASE 
        WHEN completed_at IS NOT NULL AND started_at IS NOT NULL 
        THEN TIMESTAMPDIFF(MICROSECOND, started_at, completed_at) / 1000
        ELSE NULL 
    END) as avg_duration_ms,
    MIN(scheduled_at) as oldest_task,
    MAX(scheduled_at) as newest_task
FROM task_queue
GROUP BY queue_name, task_type, status;

-- Worker性能统计视图
CREATE OR REPLACE VIEW `v_worker_performance` AS
SELECT 
    wi.worker_id,
    wi.queue_name,
    wi.host_name,
    wi.concurrency,
    wi.status,
    wi.tasks_processed,
    wi.tasks_failed,
    CASE 
        WHEN wi.tasks_processed > 0 
        THEN ROUND((wi.tasks_failed / wi.tasks_processed) * 100, 2)
        ELSE 0 
    END as failure_rate_percent,
    TIMESTAMPDIFF(SECOND, wi.last_heartbeat, NOW()) as heartbeat_delay_seconds,
    TIMESTAMPDIFF(HOUR, wi.started_at, NOW()) as uptime_hours
FROM worker_instances wi;

-- ============================================================================
-- 7. 插入初始数据
-- ============================================================================

-- 插入默认通知插件配置
INSERT IGNORE INTO notification_plugins (name, display_name, version, config, enabled, priority) VALUES
('dingtalk', '钉钉通知', '1.0.0', '{"webhook_url": "", "secret": "", "at_all": false, "message_type": "markdown"}', FALSE, 1),
('wechat_work', '企业微信通知', '1.0.0', '{"webhook_url": "", "message_type": "markdown"}', FALSE, 2),
('email', '邮件通知', '1.0.0', '{"smtp_host": "", "smtp_port": 587, "username": "", "password": "", "from_email": "", "to_emails": [], "use_tls": true, "template_type": "html"}', FALSE, 3);

-- 插入示例配置同步状态
INSERT IGNORE INTO config_sync_status (id, cluster_id, config_type, sync_status, created_at, updated_at) VALUES
('550e8400-e29b-41d4-a716-446655440001', 'cluster-1', 'prometheus', 'pending', NOW(), NOW()),
('550e8400-e29b-41d4-a716-446655440002', 'cluster-1', 'alertmanager', 'pending', NOW(), NOW()),
('550e8400-e29b-41d4-a716-446655440003', 'cluster-2', 'prometheus', 'pending', NOW(), NOW()),
('550e8400-e29b-41d4-a716-446655440004', 'cluster-2', 'vmalert', 'pending', NOW(), NOW());

COMMIT;