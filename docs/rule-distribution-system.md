# Prometheus 规则分发系统

## 概述

Prometheus 规则分发系统是一个用于管理和分发 Prometheus 告警规则的微服务系统。它提供了规则的创建、更新、删除、验证、分发、冲突检测和版本控制等功能。

## 架构设计

系统采用 Clean Architecture 设计模式，分为以下几层：

- **Presentation Layer**: HTTP 处理器，负责处理 HTTP 请求和响应
- **Application Layer**: 业务逻辑层，实现核心业务功能
- **Domain Layer**: 领域层，定义实体、接口和业务规则
- **Infrastructure Layer**: 基础设施层，实现数据持久化和外部服务调用

## 功能特性

### 规则管理
- 创建、更新、删除 Prometheus 规则
- 规则验证和语法检查
- 规则版本控制
- 规则搜索和过滤

### 规则组管理
- 创建、更新、删除规则组
- 规则组版本控制
- 批量操作支持

### 规则分发
- 将规则分发到多个 Prometheus 集群
- 分发状态跟踪
- 失败重试机制
- 回滚支持

### 冲突检测
- 自动检测规则冲突
- 冲突解决建议
- 冲突历史记录

### 同步管理
- 集群间规则同步
- 同步状态监控
- 增量同步支持

### 统计分析
- 规则使用统计
- 分发成功率统计
- 冲突统计分析

## 快速开始

### 环境要求

- Go 1.20+
- MySQL 8.0+
- Docker (可选)

### 本地开发

1. **克隆代码**
```bash
git clone <repository-url>
cd AlertAgent
```

2. **安装依赖**
```bash
go mod download
```

3. **配置环境变量**
```bash
export DB_HOST=localhost
export DB_PORT=3306
export DB_USER=root
export DB_PASSWORD=password
export DB_NAME=alert_agent
export PORT=8080
```

4. **运行数据库迁移**
```bash
go run cmd/rule-server/main.go
```

5. **启动服务**
```bash
go run cmd/rule-server/main.go
```

### Docker 部署

1. **使用 Docker Compose**
```bash
docker-compose -f docker-compose.rule-server.yml up -d
```

2. **检查服务状态**
```bash
docker-compose -f docker-compose.rule-server.yml ps
```

3. **查看日志**
```bash
docker-compose -f docker-compose.rule-server.yml logs -f rule-server
```

## API 文档

### 规则管理 API

#### 创建规则
```http
POST /api/v1/rules
Content-Type: application/json

{
  "name": "high_cpu_usage",
  "cluster_id": "prod-cluster-1",
  "group_name": "system_alerts",
  "expression": "cpu_usage > 80",
  "for_duration": "5m",
  "severity": "warning",
  "summary": "High CPU usage detected",
  "description": "CPU usage is above 80% for more than 5 minutes",
  "labels": {
    "team": "infrastructure"
  },
  "annotations": {
    "runbook_url": "https://runbooks.example.com/cpu"
  }
}
```

#### 获取规则列表
```http
GET /api/v1/rules?limit=20&offset=0&cluster_id=prod-cluster-1&search=cpu
```

#### 更新规则
```http
PUT /api/v1/rules/{id}
Content-Type: application/json

{
  "expression": "cpu_usage > 85",
  "severity": "critical"
}
```

#### 删除规则
```http
DELETE /api/v1/rules/{id}
```

#### 验证规则
```http
POST /api/v1/rules/validate
Content-Type: application/json

{
  "expression": "cpu_usage > 80",
  "for_duration": "5m"
}
```

### 规则分发 API

#### 分发规则
```http
POST /api/v1/rules/{id}/distribute
Content-Type: application/json

{
  "cluster_ids": ["prod-cluster-1", "prod-cluster-2"]
}
```

#### 获取分发状态
```http
GET /api/v1/rules/{id}/distribution?cluster_id=prod-cluster-1
```

#### 列出分发记录
```http
GET /api/v1/distributions?rule_id=123&status=success
```

### 集群同步 API

#### 同步规则到集群
```http
POST /api/v1/clusters/{cluster_id}/sync
```

#### 获取同步状态
```http
GET /api/v1/clusters/{cluster_id}/sync-status
```

### 冲突管理 API

#### 检测冲突
```http
GET /api/v1/conflicts/detect?cluster_id=prod-cluster-1
```

#### 列出冲突
```http
GET /api/v1/conflicts?status=unresolved
```

#### 解决冲突
```http
PUT /api/v1/conflicts/{id}/resolve
Content-Type: application/json

{
  "resolution": "keep_newer",
  "comment": "Keep the newer version of the rule"
}
```

### 统计信息 API

#### 获取规则统计
```http
GET /api/v1/stats/rules?cluster_id=prod-cluster-1&start_date=2024-01-01&end_date=2024-01-31
```

## 数据模型

### PrometheusRule
```go
type PrometheusRule struct {
    ID          uint      `json:"id" gorm:"primaryKey"`
    Name        string    `json:"name" gorm:"uniqueIndex:idx_rule_cluster"`
    ClusterID   string    `json:"cluster_id" gorm:"uniqueIndex:idx_rule_cluster"`
    GroupName   string    `json:"group_name"`
    Expression  string    `json:"expression"`
    ForDuration string    `json:"for_duration"`
    Severity    string    `json:"severity"`
    Summary     string    `json:"summary"`
    Description string    `json:"description"`
    Labels      JSON      `json:"labels"`
    Annotations JSON      `json:"annotations"`
    Enabled     bool      `json:"enabled"`
    Version     int       `json:"version"`
    CreatedAt   time.Time `json:"created_at"`
    UpdatedAt   time.Time `json:"updated_at"`
}
```

### RuleDistribution
```go
type RuleDistribution struct {
    ID        uint      `json:"id" gorm:"primaryKey"`
    RuleID    uint      `json:"rule_id"`
    ClusterID string    `json:"cluster_id"`
    Status    string    `json:"status"`
    Message   string    `json:"message"`
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
}
```

## 配置说明

### 环境变量

| 变量名 | 描述 | 默认值 |
|--------|------|--------|
| DB_HOST | 数据库主机 | localhost |
| DB_PORT | 数据库端口 | 3306 |
| DB_USER | 数据库用户名 | root |
| DB_PASSWORD | 数据库密码 | password |
| DB_NAME | 数据库名称 | alert_agent |
| PORT | 服务端口 | 8080 |
| GIN_MODE | Gin 模式 | release |

## 监控和日志

### 健康检查
```http
GET /health
```

### 日志格式
系统使用结构化日志，包含以下字段：
- timestamp: 时间戳
- level: 日志级别
- message: 日志消息
- context: 上下文信息

## 故障排除

### 常见问题

1. **数据库连接失败**
   - 检查数据库配置
   - 确认数据库服务正在运行
   - 验证网络连接

2. **规则验证失败**
   - 检查 Prometheus 表达式语法
   - 确认指标名称正确
   - 验证时间格式

3. **分发失败**
   - 检查目标集群连接
   - 验证集群配置
   - 查看错误日志

### 日志查看
```bash
# Docker 环境
docker-compose -f docker-compose.rule-server.yml logs -f rule-server

# 本地环境
tail -f /var/log/rule-server.log
```

## 开发指南

### 代码结构
```
internal/
├── application/rule/     # 应用服务层
├── domain/rule/         # 领域层
├── handler/            # HTTP 处理器
├── infrastructure/     # 基础设施层
└── router/            # 路由配置
```

### 测试
```bash
# 运行单元测试
go test ./...

# 运行集成测试
go test -tags=integration ./...

# 生成测试覆盖率报告
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### 贡献指南

1. Fork 项目
2. 创建功能分支
3. 提交更改
4. 推送到分支
5. 创建 Pull Request

## 许可证

MIT License