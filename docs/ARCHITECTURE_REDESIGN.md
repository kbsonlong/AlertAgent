# AlertAgent 架构重构文档

## 概述

本文档描述了 AlertAgent 系统的架构重构实现，将系统从独立的告警管理系统转变为智能告警管理和分发中心。

## 项目结构

### 新的目录结构

```
alert_agent/
├── cmd/                           # 应用程序入口点
│   └── server/                    # 主服务器
├── internal/                      # 私有应用代码
│   ├── application/               # 应用服务层
│   │   ├── channel/              # 渠道管理服务
│   │   ├── cluster/              # 集群管理服务
│   │   ├── gateway/              # 智能网关服务
│   │   └── analysis/             # AI分析服务
│   ├── domain/                    # 领域模型层
│   │   ├── channel/              # 渠道领域模型
│   │   ├── cluster/              # 集群领域模型
│   │   ├── gateway/              # 网关领域模型
│   │   └── analysis/             # 分析领域模型
│   ├── infrastructure/           # 基础设施层
│   │   ├── database/             # 数据库配置和模型
│   │   ├── repository/           # 数据仓储实现
│   │   └── di/                   # 依赖注入配置
│   ├── shared/                   # 共享组件
│   │   ├── errors/               # 错误处理
│   │   ├── logger/               # 日志记录
│   │   └── middleware/           # 中间件
│   └── config/                   # 配置管理
├── config/                       # 配置文件
└── docs/                        # 文档
```

## 核心组件

### 1. 渠道管理系统 (Channel Management)

**位置**: `internal/application/channel/`, `internal/domain/channel/`

**功能**:
- 统一的告警渠道管理
- 支持多种渠道类型：钉钉、企业微信、邮件、Webhook、Slack
- 插件化架构，支持扩展新的渠道类型
- 渠道健康检查和监控
- 渠道分组和标签管理

**核心接口**:
```go
type ChannelService interface {
    CreateChannel(ctx context.Context, req *CreateChannelRequest) (*Channel, error)
    SendMessage(ctx context.Context, channelID string, message *Message) error
    TestChannel(ctx context.Context, id string) (*TestResult, error)
    RegisterPlugin(plugin ChannelPlugin) error
}
```

### 2. 集群管理系统 (Cluster Management)

**位置**: `internal/application/cluster/`, `internal/domain/cluster/`

**功能**:
- 多个 Alertmanager 集群的中央管理
- 集群健康检查和监控
- 配置分发和同步
- Sidecar 模式配置同步
- 故障转移和负载均衡

**核心接口**:
```go
type ClusterService interface {
    RegisterCluster(ctx context.Context, config *ClusterConfig) (*Cluster, error)
    DistributeConfig(ctx context.Context, clusterID string, config *Config) error
    HealthCheck(ctx context.Context) (map[string]*HealthStatus, error)
}
```

### 3. 智能告警网关 (Smart Gateway)

**位置**: `internal/application/gateway/`, `internal/domain/gateway/`

**功能**:
- 告警接收和处理
- 告警收敛和抑制
- 智能路由（第二阶段）
- 处理记录和状态跟踪
- 规则管理

**核心接口**:
```go
type GatewayService interface {
    ReceiveAlert(ctx context.Context, alert *Alert) error
    ProcessAlert(ctx context.Context, alert *Alert) (*ProcessingResult, error)
    CreateSuppressionRule(ctx context.Context, rule *SuppressionRule) error
    CreateRoutingRule(ctx context.Context, rule *RoutingRule) error
}
```

### 4. AI分析系统 (Analysis System)

**位置**: `internal/application/analysis/`, `internal/domain/analysis/`

**功能**:
- 异步告警分析
- AI 驱动的根因分析
- 分析结果存储和查询
- 分析统计和监控

**核心接口**:
```go
type AnalysisService interface {
    CreateAnalysis(ctx context.Context, request *AnalysisRequest) (*AnalysisRecord, error)
    GetAnalysisStats(ctx context.Context, startTime, endTime time.Time) (*AnalysisStats, error)
}
```

## 基础设施

### 1. 依赖注入 (Dependency Injection)

使用 Google Wire 进行依赖注入管理：

```go
// ProviderSet 全局依赖注入提供者集合
var ProviderSet = wire.NewSet(
    // 基础设施提供者
    ProvideConfig,
    ProvideLogger,
    ProvideDatabase,
    ProvideRedis,
    
    // 仓储提供者
    ProvideChannelRepository,
    ProvideClusterRepository,
    ProvideAnalysisRepository,
    ProvideGatewayRepository,
    
    // 应用服务提供者
    ProvideChannelService,
    ProvideClusterService,
    ProvideAnalysisService,
    ProvideGatewayService,
)
```

### 2. 错误处理

统一的错误处理机制：

```go
type AppError struct {
    Type    ErrorType `json:"type"`
    Code    string    `json:"code"`
    Message string    `json:"message"`
    Details string    `json:"details,omitempty"`
    Cause   error     `json:"-"`
}
```

支持的错误类型：
- `validation`: 验证错误
- `not_found`: 资源未找到
- `conflict`: 资源冲突
- `internal`: 内部错误
- `external`: 外部服务错误
- `timeout`: 超时错误
- `rate_limit`: 限流错误

### 3. 日志记录

基于 Zap 的结构化日志记录：

```go
// 支持上下文日志
logger := logger.WithContext(ctx)
logger.Info("Operation completed", 
    zap.String("operation", "create_channel"),
    zap.String("channel_id", channelID))
```

### 4. 配置管理

支持热重载的 YAML 配置：

```yaml
server:
  port: 8080
  mode: debug

channel:
  health_check_interval: 60
  max_retries: 3
  timeout: 30

cluster:
  default_sync_interval: 30
  max_concurrent_syncs: 5

gateway:
  enable_convergence: true
  direct_mode_enabled: true

analysis:
  enabled: true
  async_mode: true
  max_concurrent_tasks: 10
```

## 数据模型

### 核心数据表

1. **channels**: 告警渠道
2. **channel_groups**: 渠道分组
3. **clusters**: Alertmanager 集群
4. **alert_processing_records**: 告警处理记录
5. **ai_analysis_records**: AI 分析记录
6. **convergence_records**: 告警收敛记录
7. **suppression_rules**: 抑制规则
8. **routing_rules**: 路由规则
9. **config_sync_records**: 配置同步记录
10. **audit_logs**: 审计日志

### 数据库迁移

自动迁移支持：

```go
func autoMigrate(db *gorm.DB) error {
    return db.AutoMigrate(
        // 现有模型
        &model.Alert{},
        &model.Rule{},
        
        // 新增模型
        &Channel{},
        &ChannelGroup{},
        &Cluster{},
        &AlertProcessingRecord{},
        &AIAnalysisRecord{},
        // ...
    )
}
```

## 分阶段实施

### 第一阶段：基础架构和直通模式

**已完成**:
- ✅ 项目结构重构
- ✅ 依赖注入框架设置
- ✅ 统一错误处理和日志记录
- ✅ 数据库模型设计
- ✅ 基础服务接口定义
- ✅ 配置管理系统

**特点**:
- 告警直通模式：告警直接通过用户定义渠道发送
- 基础的告警收敛功能（可选开关）
- 异步 AI 分析（结果回写）
- 确保告警及时性

### 第二阶段：智能功能（待实现）

**计划功能**:
- 高级告警收敛引擎
- 智能路由引擎
- 机器学习决策逻辑
- AI 驱动的自动化操作
- 动态规则更新

## 运行和部署

### 开发环境

1. 安装依赖：
```bash
go mod download
```

2. 配置数据库和 Redis

3. 启动服务：
```bash
go run cmd/server/main.go
```

### 配置文件

主配置文件位于 `config/config.yaml`，支持热重载。

### API 端点

- `GET /health` - 健康检查
- `POST /api/v1/channels` - 创建渠道
- `GET /api/v1/channels` - 列出渠道
- `POST /api/v1/clusters` - 注册集群
- `POST /api/v1/gateway/alerts` - 接收告警
- `POST /api/v1/analysis` - 创建分析

## 技术优势

1. **职责分离**: 各组件职责明确，易于维护
2. **技术解耦**: 通过标准接口集成，降低耦合度
3. **水平扩展**: 支持多集群管理，可根据需求扩展
4. **高可用性**: 分布式架构设计，单点故障不影响整体服务
5. **可观测性**: 完整的日志、指标和链路追踪
6. **配置驱动**: 支持热重载的配置管理

## 下一步计划

1. 实现具体的 HTTP 处理器
2. 添加渠道插件实现
3. 实现配置同步 Sidecar
4. 集成 n8n 和 Dify
5. 添加监控和指标收集
6. 完善测试覆盖率
7. 部署和运维文档

## 贡献指南

1. 遵循现有的代码结构和命名约定
2. 为新功能添加相应的测试
3. 更新相关文档
4. 确保代码通过 lint 检查