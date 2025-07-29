-- AlertAgent PostgreSQL 数据库初始化脚本
-- 设置数据库配置

-- 创建扩展
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pg_trgm";
CREATE EXTENSION IF NOT EXISTS "btree_gin";

-- 设置时区
SET timezone = 'Asia/Shanghai';

-- 创建自定义类型
DO $$ BEGIN
    CREATE TYPE alert_status AS ENUM ('pending', 'firing', 'resolved', 'silenced');
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

DO $$ BEGIN
    CREATE TYPE rule_status AS ENUM ('active', 'inactive', 'draft');
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

DO $$ BEGIN
    CREATE TYPE channel_type AS ENUM ('email', 'slack', 'webhook', 'dingtalk', 'wechat', 'sms');
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

DO $$ BEGIN
    CREATE TYPE processing_status AS ENUM ('pending', 'processing', 'completed', 'failed', 'skipped');
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

-- 创建函数：更新 updated_at 字段
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- 创建函数：生成短 UUID
CREATE OR REPLACE FUNCTION generate_short_uuid()
RETURNS TEXT AS $$
BEGIN
    RETURN SUBSTRING(REPLACE(uuid_generate_v4()::TEXT, '-', ''), 1, 12);
END;
$$ language 'plpgsql';

-- 设置默认权限
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT SELECT, INSERT, UPDATE, DELETE ON TABLES TO postgres;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT USAGE, SELECT ON SEQUENCES TO postgres;

-- 创建索引函数
CREATE OR REPLACE FUNCTION create_gin_index(table_name TEXT, column_name TEXT)
RETURNS VOID AS $$
BEGIN
    EXECUTE format('CREATE INDEX IF NOT EXISTS idx_%s_%s_gin ON %s USING gin(%s)', 
                   table_name, column_name, table_name, column_name);
END;
$$ language 'plpgsql';

-- 输出初始化完成信息
DO $$
BEGIN
    RAISE NOTICE 'AlertAgent PostgreSQL 数据库初始化完成';
    RAISE NOTICE '时区设置: %', current_setting('timezone');
    RAISE NOTICE '数据库版本: %', version();
END $$;

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
CREATE TABLE IF NOT EXISTS knowledges (
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

-- 数据源表
CREATE TABLE IF NOT EXISTS providers (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    created_at DATETIME(3) NULL,
    updated_at DATETIME(3) NULL,
    deleted_at DATETIME(3) NULL,
    name VARCHAR(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
    type VARCHAR(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
    status VARCHAR(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 'active',
    description TEXT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci,
    endpoint VARCHAR(500) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
    auth_type VARCHAR(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT 'none',
    auth_config TEXT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci,
    labels TEXT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci,
    last_check DATETIME(3) NULL,
    last_error TEXT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci,
    INDEX idx_providers_deleted_at (deleted_at),
    INDEX idx_providers_type (type),
    INDEX idx_providers_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='数据源表';

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