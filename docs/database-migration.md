# AlertAgent 数据库迁移系统

本文档介绍 AlertAgent 数据库迁移系统的设计、实现和使用方法。

## 概述

AlertAgent 数据库迁移系统基于 Go 和 GORM 实现，提供了完整的数据库版本管理功能，支持：

- 自动化数据库迁移
- 版本回滚
- 迁移状态跟踪
- 错误处理和恢复
- 迁移历史管理

## 架构设计

### 核心组件

1. **Migration 记录表**：跟踪所有迁移的执行状态
2. **MigrationStep**：定义单个迁移步骤
3. **Migrator**：执行迁移的核心引擎
4. **Manager**：提供高级迁移管理功能

### 目录结构

```
internal/infrastructure/migration/
├── migration.go    # 核心迁移引擎
├── steps.go        # 迁移步骤定义
└── manager.go      # 迁移管理器

cmd/migrate/
└── main.go         # 命令行工具
```

## 迁移版本规划

### V1 到 V2 迁移内容

#### 新增表

1. **alert_channels** - 告警通道配置
2. **channel_groups** - 通道分组管理
3. **channel_templates** - 通道模板
4. **channel_usage_stats** - 通道使用统计
5. **channel_permissions** - 通道权限管理
6. **alertmanager_clusters** - Alertmanager 集群配置
7. **rule_distributions** - 规则分发记录
8. **alert_processing_records** - 告警处理记录
9. **ai_analysis_records** - AI 分析记录
10. **automation_actions** - 自动化操作记录
11. **alert_convergence_records** - 告警收敛记录
12. **cluster_health_status** - 集群健康状态

#### 扩展现有表

1. **rules** 表新增字段：
   - `cluster_id` - 关联集群
   - `channel_group_id` - 关联通道组
   - `priority` - 优先级
   - `tags` - 标签
   - `metadata` - 元数据

2. **alerts** 表新增字段：
   - `cluster_id` - 关联集群
   - `channel_id` - 关联通道
   - `processing_status` - 处理状态
   - `ai_analysis_id` - AI 分析记录ID
   - `convergence_id` - 收敛记录ID

## 使用指南

### 命令行工具

#### 构建迁移工具

```bash
# 使用 Makefile
make migrate-build

# 或直接使用 go build
go build -o bin/migrate ./cmd/migrate
```

#### 基本操作

```bash
# 执行迁移到最新版本
make migrate
# 或
./bin/migrate -action=migrate

# 查看迁移状态
make migrate-status
# 或
./bin/migrate -action=status

# 验证数据库状态
make migrate-validate
# 或
./bin/migrate -action=validate

# 显示详细迁移信息
make migrate-info
# 或
./bin/migrate -action=info
```

#### 回滚操作

```bash
# 回滚到指定版本
make migrate-rollback VERSION=v2.0.0-001
# 或
./bin/migrate -action=rollback -version=v2.0.0-001
```

#### 维护操作

```bash
# 修复失败的迁移
./bin/migrate -action=repair -version=v2.0.0-005

# 清理30天前的迁移历史
make migrate-cleanup DAYS=30
# 或
./bin/migrate -action=cleanup -keep-days=30

# 检查数据库是否为最新版本
./bin/migrate -action=check
```

### 配置参数

#### 数据库连接

```bash
# 通过命令行参数
./bin/migrate \
  -db-host=localhost \
  -db-port=5432 \
  -db-user=postgres \
  -db-password=secret \
  -db-name=alert_agent \
  -action=migrate

# 通过环境变量
export DB_HOST=localhost
export DB_PORT=5432
export DB_USER=postgres
export DB_PASSWORD=secret
export DB_NAME=alert_agent
./bin/migrate -action=migrate
```

#### 其他选项

```bash
# 设置日志级别
./bin/migrate -log-level=debug -action=migrate

# 设置超时时间
./bin/migrate -timeout=60m -action=migrate
```

## 开发指南

### 添加新的迁移步骤

1. 在 `internal/infrastructure/migration/steps.go` 中添加新的迁移函数：

```go
// createNewTableV2001 创建新表
func createNewTableV2001(db *gorm.DB) error {
    return db.Exec(`
        CREATE TABLE new_table (
            id SERIAL PRIMARY KEY,
            name VARCHAR(255) NOT NULL,
            created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
        )
    `).Error
}

// dropNewTableV2001 删除新表
func dropNewTableV2001(db *gorm.DB) error {
    return db.Exec("DROP TABLE IF EXISTS new_table").Error
}
```

2. 在 `GetMigrationSteps()` 函数中注册新步骤：

```go
steps = append(steps, migration.MigrationStep{
    Version:     "v2.0.0-013",
    Name:        "Create new table",
    Description: "创建新的业务表",
    Up:          createNewTableV2001,
    Down:        dropNewTableV2001,
    Checksum:    "checksum_value",
})
```

### 最佳实践

1. **版本命名**：使用语义化版本号，格式为 `v{major}.{minor}.{patch}-{sequence}`
2. **向后兼容**：确保迁移步骤可以安全回滚
3. **事务安全**：每个迁移步骤在事务中执行
4. **错误处理**：提供详细的错误信息和恢复建议
5. **测试验证**：在开发环境充分测试迁移脚本

### 校验和生成

```go
import (
    "crypto/sha256"
    "fmt"
)

func generateChecksum(content string) string {
    hash := sha256.Sum256([]byte(content))
    return fmt.Sprintf("%x", hash)
}
```

## 故障排除

### 常见问题

1. **迁移失败**
   - 检查数据库连接
   - 查看错误日志
   - 使用 `repair` 命令修复

2. **版本不一致**
   - 使用 `validate` 命令检查
   - 手动同步迁移记录

3. **回滚失败**
   - 检查回滚函数实现
   - 手动执行回滚SQL

### 日志分析

```bash
# 启用调试日志
./bin/migrate -log-level=debug -action=status

# 查看详细错误信息
./bin/migrate -action=info
```

### 手动修复

```sql
-- 查看迁移记录
SELECT * FROM migrations ORDER BY executed_at DESC;

-- 标记迁移为成功
UPDATE migrations SET success = true, error_msg = '' WHERE version = 'v2.0.0-005';

-- 删除失败的迁移记录
DELETE FROM migrations WHERE version = 'v2.0.0-005' AND success = false;
```

## 监控和维护

### 性能监控

- 监控迁移执行时间
- 跟踪数据库大小变化
- 记录迁移频率

### 定期维护

```bash
# 每月清理迁移历史
0 0 1 * * /path/to/migrate -action=cleanup -keep-days=90

# 每周验证数据库状态
0 0 * * 0 /path/to/migrate -action=validate
```

## 安全考虑

1. **权限控制**：确保迁移工具具有适当的数据库权限
2. **备份策略**：迁移前自动备份数据库
3. **审计日志**：记录所有迁移操作
4. **环境隔离**：生产环境迁移需要额外审批

## CI/CD 集成

### GitHub Actions

项目包含自动化的迁移测试工作流 `.github/workflows/migration-test.yml`，会在以下情况触发：

- 推送到 `main` 或 `develop` 分支
- 创建针对 `main` 或 `develop` 分支的 Pull Request
- 修改迁移相关文件

#### 测试内容

1. **基础迁移测试**：
   - 构建迁移工具
   - 执行向上迁移
   - 验证迁移状态
   - 测试回滚功能
   - 重新应用迁移

2. **Docker 迁移测试**：
   - 构建 Docker 镜像
   - 使用 Docker Compose 测试
   - 验证容器化环境

3. **代码质量检查**：
   - 运行 golangci-lint
   - 检查代码格式
   - 扫描 TODO/FIXME 注释

### 本地 CI 模拟

```bash
# 运行完整的迁移测试套件
make migrate-setup-clean
make migrate-setup
make migrate-verify

# 测试回滚功能
./scripts/migrate-setup.sh --rollback --target v2.0.0-001

# 重新应用迁移
make migrate-setup
```

## 快速开始

### 一键设置

```bash
# 快速设置开发环境
make migrate-setup

# 或者清理后重新设置
make migrate-setup-clean
```

### 手动设置

```bash
# 1. 启动数据库
docker-compose -f docker-compose.dev.yml up -d postgres

# 2. 构建迁移工具
make migrate-build

# 3. 执行迁移
make migrate

# 4. 验证结果
make migrate-status
make migrate-validate
```

### Docker 方式

```bash
# 构建 Docker 镜像
make migrate-docker-build

# 使用 Docker Compose 运行迁移
make migrate-docker

# 检查状态
make migrate-docker-status
```

## 生产环境部署

### 部署前检查

```bash
# 1. 备份数据库
pg_dump -h $DB_HOST -U $DB_USER -d $DB_NAME > backup_$(date +%Y%m%d_%H%M%S).sql

# 2. 验证迁移脚本
./bin/migrate -action=validate -dry-run=true

# 3. 检查迁移计划
./bin/migrate -action=info
```

### 部署步骤

```bash
# 1. 停止应用服务
sudo systemctl stop alertagent

# 2. 执行迁移
./bin/migrate -action=migrate -timeout=60m

# 3. 验证迁移结果
./bin/migrate -action=validate

# 4. 启动应用服务
sudo systemctl start alertagent

# 5. 健康检查
curl -f http://localhost:8080/health
```

### 回滚计划

```bash
# 如果迁移失败，执行回滚
./bin/migrate -action=rollback -version=$PREVIOUS_VERSION

# 恢复数据库备份（最后手段）
psql -h $DB_HOST -U $DB_USER -d $DB_NAME < backup_file.sql
```

## 高级功能

### 并行迁移

```bash
# 设置并发数（谨慎使用）
export MIGRATION_CONCURRENCY=2
./bin/migrate -action=migrate
```

### 迁移锁定

```bash
# 检查迁移锁状态
./bin/migrate -action=lock-status

# 强制释放锁（紧急情况）
./bin/migrate -action=force-unlock
```

### 自定义迁移

```bash
# 生成新的迁移文件模板
./bin/migrate -action=generate -name="add_new_feature"

# 验证迁移语法
./bin/migrate -action=validate -file="path/to/migration.sql"
```

## 监控和告警

### Prometheus 指标

迁移工具暴露以下指标：

- `migration_duration_seconds` - 迁移执行时间
- `migration_success_total` - 成功迁移计数
- `migration_failure_total` - 失败迁移计数
- `migration_rollback_total` - 回滚操作计数

### 日志监控

```bash
# 监控迁移日志
tail -f /var/log/alertagent/migration.log

# 搜索错误日志
grep "ERROR" /var/log/alertagent/migration.log
```

### 告警规则

```yaml
# Prometheus 告警规则示例
groups:
- name: migration.rules
  rules:
  - alert: MigrationFailure
    expr: increase(migration_failure_total[5m]) > 0
    for: 0m
    labels:
      severity: critical
    annotations:
      summary: "Database migration failed"
      description: "Migration failed in the last 5 minutes"

  - alert: MigrationDurationHigh
    expr: migration_duration_seconds > 300
    for: 0m
    labels:
      severity: warning
    annotations:
      summary: "Migration taking too long"
      description: "Migration duration is {{ $value }} seconds"
```

## 参考资料

- [GORM 官方文档](https://gorm.io/docs/)
- [PostgreSQL 迁移最佳实践](https://www.postgresql.org/docs/current/ddl-alter.html)
- [数据库版本控制策略](https://flywaydb.org/documentation/concepts/migrations)
- [GitHub Actions 工作流语法](https://docs.github.com/en/actions/using-workflows/workflow-syntax-for-github-actions)
- [Docker Compose 文件格式](https://docs.docker.com/compose/compose-file/)