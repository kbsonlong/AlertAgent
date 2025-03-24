-- 创建数据库
CREATE DATABASE IF NOT EXISTS alert_agent DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

USE alert_agent;

-- 告警规则表
CREATE TABLE IF NOT EXISTS rules (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    created_at DATETIME(3) NULL,
    updated_at DATETIME(3) NULL,
    deleted_at DATETIME(3) NULL,
    name VARCHAR(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
    description TEXT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
    level VARCHAR(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
    enabled BOOLEAN NOT NULL DEFAULT TRUE,
    condition_expr TEXT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
    notify_type VARCHAR(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
    notify_group VARCHAR(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
    template VARCHAR(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
    INDEX idx_rules_deleted_at (deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='告警规则表';

-- 告警记录表
CREATE TABLE IF NOT EXISTS alerts (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    created_at DATETIME(3) NULL,
    updated_at DATETIME(3) NULL,
    deleted_at DATETIME(3) NULL,
    name VARCHAR(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
    level VARCHAR(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
    status VARCHAR(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 'active',
    source VARCHAR(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
    content TEXT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
    handler VARCHAR(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci,
    note TEXT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci,
    analysis TEXT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci,
    rule_id BIGINT UNSIGNED NOT NULL,
    template_id BIGINT UNSIGNED,
    group_id BIGINT UNSIGNED,
    title VARCHAR(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
    labels TEXT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci,
    handle_time DATETIME,
    handle_note TEXT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci,
    notify_time DATETIME,
    notify_count INT DEFAULT 0,
    INDEX idx_alerts_deleted_at (deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='告警记录表';

-- 知识库表
CREATE TABLE IF NOT EXISTS knowledge (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    created_at DATETIME(3) NULL,
    updated_at DATETIME(3) NULL,
    deleted_at DATETIME(3) NULL,
    title VARCHAR(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
    content TEXT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
    category VARCHAR(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
    tags TEXT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci,
    source VARCHAR(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
    source_id BIGINT UNSIGNED NOT NULL,
    vector JSON,
    summary TEXT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci,
    INDEX idx_knowledge_deleted_at (deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='知识库表';

-- 通知模板表
CREATE TABLE IF NOT EXISTS notify_templates (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    created_at DATETIME(3) NULL,
    updated_at DATETIME(3) NULL,
    deleted_at DATETIME(3) NULL,
    name VARCHAR(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
    type VARCHAR(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
    content TEXT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
    INDEX idx_notify_templates_deleted_at (deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='通知模板表';

-- 通知组表
CREATE TABLE IF NOT EXISTS notify_groups (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    created_at DATETIME(3) NULL,
    updated_at DATETIME(3) NULL,
    deleted_at DATETIME(3) NULL,
    name VARCHAR(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
    description TEXT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
    contacts JSON NOT NULL,
    INDEX idx_notify_groups_deleted_at (deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='通知组表';

-- 系统设置表
CREATE TABLE IF NOT EXISTS settings (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    created_at DATETIME(3) NULL,
    updated_at DATETIME(3) NULL,
    deleted_at DATETIME(3) NULL,
    ollama_endpoint VARCHAR(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
    ollama_model VARCHAR(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
    INDEX idx_settings_deleted_at (deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='系统设置表';

-- 插入示例数据
INSERT INTO notify_templates (name, type, content) VALUES
('默认邮件模板', 'email', '告警标题：${title}\n告警级别：${level}\n告警内容：${content}\n发生时间：${time}'),
('默认短信模板', 'sms', '【告警通知】${title}，级别：${level}，时间：${time}');

INSERT INTO notify_groups (name, description, contacts) VALUES
('运维组', '运维团队', '["admin@example.com", "ops@example.com"]'),
('开发组', '开发团队', '["dev@example.com"]');

INSERT INTO rules (name, description, condition_expr, level, notify_type, notify_group, template, created_at, updated_at) VALUES
('CPU使用率告警', 'CPU使用率超过阈值', 'cpu_usage > 90', 'warning', 'email', '运维组', '默认邮件模板', NOW(), NOW()),
('内存使用率告警', '内存使用率超过阈值', 'memory_usage > 90', 'warning', 'email', '运维组', '默认邮件模板', NOW(), NOW()),
('磁盘使用率告警', '磁盘使用率超过阈值', 'disk_usage > 90', 'warning', 'email', '运维组', '默认邮件模板', NOW(), NOW());

-- 插入默认设置
INSERT INTO settings (ollama_endpoint, ollama_model, created_at, updated_at)
VALUES ('http://localhost:11434', 'llama2', NOW(), NOW());