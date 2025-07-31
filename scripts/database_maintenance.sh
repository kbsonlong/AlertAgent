#!/bin/bash

# AlertAgent 数据库维护脚本
# 用于定期执行数据库优化、清理和监控任务

set -e

# 配置参数
DB_HOST="${DB_HOST:-localhost}"
DB_PORT="${DB_PORT:-3306}"
DB_USER="${DB_USER:-root}"
DB_PASSWORD="${DB_PASSWORD:-}"
DB_NAME="${DB_NAME:-alert_agent}"
LOG_FILE="${LOG_FILE:-/var/log/alert_agent_maintenance.log}"
BACKUP_DIR="${BACKUP_DIR:-/var/backups/alert_agent}"

# 日志函数
log() {
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] $1" | tee -a "$LOG_FILE"
}

# 错误处理
error_exit() {
    log "ERROR: $1"
    exit 1
}

# 检查MySQL连接
check_mysql_connection() {
    log "Checking MySQL connection..."
    mysql -h"$DB_HOST" -P"$DB_PORT" -u"$DB_USER" -p"$DB_PASSWORD" -e "SELECT 1;" > /dev/null 2>&1 || \
        error_exit "Cannot connect to MySQL database"
    log "MySQL connection successful"
}

# 执行SQL文件
execute_sql_file() {
    local sql_file="$1"
    local description="$2"
    
    if [[ ! -f "$sql_file" ]]; then
        log "WARNING: SQL file $sql_file not found, skipping $description"
        return 0
    fi
    
    log "Executing $description..."
    mysql -h"$DB_HOST" -P"$DB_PORT" -u"$DB_USER" -p"$DB_PASSWORD" "$DB_NAME" < "$sql_file" || \
        error_exit "Failed to execute $description"
    log "$description completed successfully"
}

# 执行SQL命令
execute_sql() {
    local sql="$1"
    local description="$2"
    
    log "Executing $description..."
    mysql -h"$DB_HOST" -P"$DB_PORT" -u"$DB_USER" -p"$DB_PASSWORD" "$DB_NAME" -e "$sql" || \
        error_exit "Failed to execute $description"
    log "$description completed successfully"
}

# 创建备份
create_backup() {
    log "Creating database backup..."
    
    # 创建备份目录
    mkdir -p "$BACKUP_DIR"
    
    # 生成备份文件名
    local backup_file="$BACKUP_DIR/alert_agent_$(date +%Y%m%d_%H%M%S).sql"
    
    # 执行备份
    mysqldump -h"$DB_HOST" -P"$DB_PORT" -u"$DB_USER" -p"$DB_PASSWORD" \
        --single-transaction --routines --triggers "$DB_NAME" > "$backup_file" || \
        error_exit "Failed to create database backup"
    
    # 压缩备份文件
    gzip "$backup_file" || log "WARNING: Failed to compress backup file"
    
    log "Database backup created: ${backup_file}.gz"
    
    # 清理旧备份（保留7天）
    find "$BACKUP_DIR" -name "alert_agent_*.sql.gz" -mtime +7 -delete 2>/dev/null || true
    log "Old backups cleaned up"
}

# 数据库优化
optimize_database() {
    log "Starting database optimization..."
    
    # 执行优化脚本
    local script_dir="$(dirname "$0")"
    execute_sql_file "$script_dir/optimize_database.sql" "database optimization"
    
    # 分析表
    execute_sql "CALL AnalyzeTables();" "table analysis" 2>/dev/null || {
        log "WARNING: AnalyzeTables procedure not found, running manual analysis"
        local tables=("alerts" "alert_rules" "task_queue" "config_sync_history" "notification_records")
        for table in "${tables[@]}"; do
            execute_sql "ANALYZE TABLE $table;" "analysis of table $table"
        done
    }
    
    # 优化表
    execute_sql "CALL OptimizeTables();" "table optimization" 2>/dev/null || {
        log "WARNING: OptimizeTables procedure not found, running manual optimization"
        local tables=("alerts" "alert_rules" "config_sync_history" "task_execution_history" "notification_records")
        for table in "${tables[@]}"; do
            execute_sql "OPTIMIZE TABLE $table;" "optimization of table $table"
        done
    }
    
    log "Database optimization completed"
}

# 数据清理
cleanup_data() {
    log "Starting data cleanup..."
    
    # 执行清理存储过程
    execute_sql "CALL CleanupExpiredData();" "expired data cleanup" 2>/dev/null || {
        log "WARNING: CleanupExpiredData procedure not found, running manual cleanup"
        
        # 手动清理过期数据
        execute_sql "DELETE FROM task_queue WHERE status IN ('completed', 'failed') AND completed_at < DATE_SUB(NOW(), INTERVAL 7 DAY);" "task queue cleanup"
        execute_sql "DELETE FROM worker_instances WHERE status = 'inactive' AND last_heartbeat < DATE_SUB(NOW(), INTERVAL 1 HOUR);" "worker instances cleanup"
        execute_sql "DELETE FROM config_sync_exceptions WHERE status = 'resolved' AND resolved_at < DATE_SUB(NOW(), INTERVAL 30 DAY);" "config sync exceptions cleanup"
    }
    
    log "Data cleanup completed"
}

# 数据归档
archive_data() {
    log "Starting data archiving..."
    
    # 执行归档存储过程
    execute_sql "CALL ArchiveOldData();" "data archiving" 2>/dev/null || {
        log "WARNING: ArchiveOldData procedure not found, skipping archiving"
        return 0
    }
    
    log "Data archiving completed"
}

# 监控数据库状态
monitor_database() {
    log "Monitoring database status..."
    
    # 检查表大小
    local table_sizes
    table_sizes=$(mysql -h"$DB_HOST" -P"$DB_PORT" -u"$DB_USER" -p"$DB_PASSWORD" "$DB_NAME" -e "
        SELECT 
            table_name,
            ROUND(((data_length + index_length) / 1024 / 1024), 2) as size_mb,
            table_rows
        FROM information_schema.tables 
        WHERE table_schema = '$DB_NAME'
        AND table_type = 'BASE TABLE'
        ORDER BY (data_length + index_length) DESC
        LIMIT 10;
    " 2>/dev/null) || log "WARNING: Failed to get table sizes"
    
    if [[ -n "$table_sizes" ]]; then
        log "Top 10 largest tables:"
        echo "$table_sizes" | tee -a "$LOG_FILE"
    fi
    
    # 检查连接数
    local connections
    connections=$(mysql -h"$DB_HOST" -P"$DB_PORT" -u"$DB_USER" -p"$DB_PASSWORD" -e "
        SHOW STATUS LIKE 'Threads_connected';
        SHOW STATUS LIKE 'Max_used_connections';
        SHOW VARIABLES LIKE 'max_connections';
    " 2>/dev/null) || log "WARNING: Failed to get connection status"
    
    if [[ -n "$connections" ]]; then
        log "Connection status:"
        echo "$connections" | tee -a "$LOG_FILE"
    fi
    
    # 检查慢查询
    local slow_queries
    slow_queries=$(mysql -h"$DB_HOST" -P"$DB_PORT" -u"$DB_USER" -p"$DB_PASSWORD" -e "
        SHOW STATUS LIKE 'Slow_queries';
    " 2>/dev/null) || log "WARNING: Failed to get slow query status"
    
    if [[ -n "$slow_queries" ]]; then
        log "Slow query status:"
        echo "$slow_queries" | tee -a "$LOG_FILE"
    fi
    
    log "Database monitoring completed"
}

# 验证数据完整性
validate_data() {
    log "Validating data integrity..."
    
    local script_dir="$(dirname "$0")"
    if [[ -f "$script_dir/validate_migration.sql" ]]; then
        execute_sql_file "$script_dir/validate_migration.sql" "data integrity validation"
    else
        log "WARNING: validate_migration.sql not found, skipping validation"
    fi
    
    log "Data integrity validation completed"
}

# 主函数
main() {
    local operation="${1:-full}"
    
    log "Starting database maintenance - Operation: $operation"
    
    # 检查MySQL连接
    check_mysql_connection
    
    case "$operation" in
        "backup")
            create_backup
            ;;
        "optimize")
            optimize_database
            ;;
        "cleanup")
            cleanup_data
            ;;
        "archive")
            archive_data
            ;;
        "monitor")
            monitor_database
            ;;
        "validate")
            validate_data
            ;;
        "full")
            create_backup
            cleanup_data
            archive_data
            optimize_database
            monitor_database
            validate_data
            ;;
        *)
            echo "Usage: $0 [backup|optimize|cleanup|archive|monitor|validate|full]"
            echo ""
            echo "Operations:"
            echo "  backup   - Create database backup"
            echo "  optimize - Optimize database tables and indexes"
            echo "  cleanup  - Clean up expired data"
            echo "  archive  - Archive old data"
            echo "  monitor  - Monitor database status"
            echo "  validate - Validate data integrity"
            echo "  full     - Run all operations (default)"
            exit 1
            ;;
    esac
    
    log "Database maintenance completed successfully - Operation: $operation"
}

# 执行主函数
main "$@"