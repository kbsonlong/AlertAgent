-- 创建通知插件配置表
CREATE TABLE IF NOT EXISTS notification_plugins (
    id INT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(100) NOT NULL UNIQUE COMMENT '插件名称',
    display_name VARCHAR(200) NOT NULL COMMENT '插件显示名称',
    version VARCHAR(50) NOT NULL COMMENT '插件版本',
    config JSON NOT NULL COMMENT '插件配置JSON',
    enabled BOOLEAN DEFAULT FALSE COMMENT '是否启用',
    priority INT DEFAULT 0 COMMENT '发送优先级，数字越小优先级越高',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    deleted_at TIMESTAMP NULL COMMENT '删除时间',
    
    INDEX idx_name (name),
    INDEX idx_enabled (enabled),
    INDEX idx_priority (priority),
    INDEX idx_deleted_at (deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='通知插件配置表';

-- 插入默认插件配置示例
INSERT INTO notification_plugins (name, display_name, version, config, enabled, priority) VALUES
('dingtalk', '钉钉通知', '1.0.0', '{"webhook_url": "", "secret": "", "at_all": false, "message_type": "markdown"}', FALSE, 1),
('wechat_work', '企业微信通知', '1.0.0', '{"webhook_url": "", "message_type": "markdown"}', FALSE, 2),
('email', '邮件通知', '1.0.0', '{"smtp_host": "", "smtp_port": 587, "username": "", "password": "", "from_email": "", "to_emails": [], "use_tls": true, "template_type": "html"}', FALSE, 3)
ON DUPLICATE KEY UPDATE
display_name = VALUES(display_name),
version = VALUES(version),
updated_at = CURRENT_TIMESTAMP;