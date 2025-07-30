-- 配置同步状态表
CREATE TABLE IF NOT EXISTS config_sync_status (
    id VARCHAR(36) PRIMARY KEY,
    cluster_id VARCHAR(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
    config_type VARCHAR(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
    config_hash VARCHAR(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci,
    sync_status VARCHAR(20) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 'pending',
    sync_time TIMESTAMP NULL,
    error_message TEXT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    
    INDEX idx_cluster_type (cluster_id, config_type),
    INDEX idx_sync_status (sync_status),
    INDEX idx_sync_time (sync_time),
    INDEX idx_deleted_at (deleted_at),
    
    UNIQUE KEY uk_cluster_config (cluster_id, config_type, deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='配置同步状态表';

-- 配置同步触发记录表
CREATE TABLE IF NOT EXISTS config_sync_triggers (
    id VARCHAR(36) PRIMARY KEY,
    cluster_id VARCHAR(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
    config_type VARCHAR(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
    trigger_by VARCHAR(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
    reason VARCHAR(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci,
    status VARCHAR(20) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 'pending',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    
    INDEX idx_cluster_type (cluster_id, config_type),
    INDEX idx_status (status),
    INDEX idx_trigger_by (trigger_by),
    INDEX idx_created_at (created_at),
    INDEX idx_deleted_at (deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='配置同步触发记录表';

-- 配置同步历史表
CREATE TABLE IF NOT EXISTS config_sync_history (
    id VARCHAR(36) PRIMARY KEY,
    cluster_id VARCHAR(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
    config_type VARCHAR(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
    config_hash VARCHAR(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
    config_size BIGINT NOT NULL DEFAULT 0,
    sync_status VARCHAR(20) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
    sync_duration BIGINT NOT NULL DEFAULT 0 COMMENT '同步耗时，毫秒',
    error_message TEXT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    
    INDEX idx_cluster_type (cluster_id, config_type),
    INDEX idx_sync_status (sync_status),
    INDEX idx_config_hash (config_hash),
    INDEX idx_created_at (created_at),
    INDEX idx_deleted_at (deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='配置同步历史表';

-- 插入一些示例数据
INSERT IGNORE INTO config_sync_status (id, cluster_id, config_type, sync_status, created_at, updated_at) VALUES
('550e8400-e29b-41d4-a716-446655440001', 'cluster-1', 'prometheus', 'pending', NOW(), NOW()),
('550e8400-e29b-41d4-a716-446655440002', 'cluster-1', 'alertmanager', 'pending', NOW(), NOW()),
('550e8400-e29b-41d4-a716-446655440003', 'cluster-2', 'prometheus', 'pending', NOW(), NOW()),
('550e8400-e29b-41d4-a716-446655440004', 'cluster-2', 'vmalert', 'pending', NOW(), NOW());