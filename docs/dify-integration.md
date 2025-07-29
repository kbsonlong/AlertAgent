# Dify AI 分析集成文档

## 概述

本文档描述了 AlertAgent 系统与 Dify AI 平台的集成实现，提供异步的智能告警分析功能。

## 架构设计

### 核心组件

1. **DifyClient** - Dify API 客户端
2. **DifyAnalysisService** - 异步分析服务
3. **DifyAnalysisRepository** - 分析结果存储
4. **配置管理** - 灵活的配置系统

### 分层架构

```
┌─────────────────────────────────────┐
│           HTTP Handler              │
├─────────────────────────────────────┤
│        DifyAnalysisService          │
├─────────────────────────────────────┤
│  DifyClient  │ DifyAnalysisRepository│
├─────────────────────────────────────┤
│        Dify API │ Database           │
└─────────────────────────────────────┘
```

## 实现文件结构

### 领域层 (Domain Layer)

#### `internal/domain/analysis/dify_client.go`
- 定义 `DifyClient` 接口
- 包含聊天消息、工作流执行、健康检查等方法
- 定义请求/响应数据结构

#### `internal/domain/analysis/dify_analysis_service.go`
- 定义 `DifyAnalysisService` 接口
- 包含异步分析、结果查询、历史管理等方法
- 定义分析任务、进度、结果等数据结构

### 基础设施层 (Infrastructure Layer)

#### `internal/infrastructure/dify/dify_client_impl.go`
- 实现 `DifyClient` 接口
- 处理 HTTP 请求和响应
- 包含重试机制和错误处理

#### `internal/infrastructure/dify/config.go`
- 定义 Dify 配置结构
- 提供默认配置和验证功能
- 支持工作流映射和知识库配置

#### `internal/infrastructure/dify/wire.go`
- 依赖注入配置
- 组件初始化和绑定

#### `internal/infrastructure/repository/dify_analysis_repository_impl.go`
- 实现 `DifyAnalysisRepository` 接口
- 提供分析结果的 CRUD 操作
- 支持历史查询和趋势分析

### 应用层 (Application Layer)

#### `internal/application/analysis/dify_analysis_service_impl.go`
- 实现 `DifyAnalysisService` 接口
- 管理异步分析任务
- 协调客户端和仓储操作

## 配置说明

### 基础配置

```yaml
dify:
  base_url: "https://api.dify.ai"
  api_key: "${DIFY_API_KEY}"
  app_token: "${DIFY_APP_TOKEN}"
  user_id: "alert-agent"
  timeout: "30s"
```

### 工作流配置

```yaml
workflow:
  default_workflow_id: "default-alert-analysis"
  workflow_mapping:
    cpu_high: "cpu-analysis-workflow"
    memory_high: "memory-analysis-workflow"
    disk_full: "disk-analysis-workflow"
```

### 知识库配置

```yaml
knowledge:
  default_dataset_ids:
    - "general-troubleshooting"
    - "best-practices"
  dataset_mapping:
    cpu_high:
      - "cpu-performance"
      - "system-optimization"
```

## 使用示例

### 1. 初始化服务

```go
// 使用依赖注入
service, cleanup, err := dify.InitializeDifyAnalysisService(
    ctx, db, logger,
)
defer cleanup()
```

### 2. 提交分析任务

```go
request := &analysis.DifyAnalysisRequest{
    AlertID:      "alert-001",
    AnalysisType: "root_cause_analysis",
    Priority:     analysis.PriorityHigh,
    Options: &analysis.DifyAnalysisOptions{
        WorkflowID:        "cpu-analysis-workflow",
        KnowledgeDatasets: []string{"cpu-performance"},
        IncludeContext:    true,
    },
}

taskID, err := service.AnalyzeAlert(ctx, alert, request)
```

### 3. 查询分析进度

```go
progress, err := service.GetAnalysisProgress(ctx, taskID)
if progress.Status == analysis.StatusCompleted {
    result, err := service.GetAnalysisResult(ctx, taskID)
    // 处理分析结果
}
```

### 4. 搜索知识库

```go
result, err := service.SearchKnowledge(ctx, &analysis.KnowledgeSearchRequest{
    Query:      "CPU使用率过高的解决方案",
    DatasetIDs: []string{"cpu-performance"},
    Limit:      5,
})
```

## 核心功能

### 1. 异步分析处理

- **任务队列管理**: 支持并发任务处理
- **状态跟踪**: 实时监控分析进度
- **错误处理**: 自动重试和错误恢复
- **超时控制**: 防止长时间运行的任务

### 2. 告警上下文构建

- **历史数据**: 包含相关历史告警
- **系统信息**: 收集系统状态和指标
- **关联分析**: 识别相关告警和模式
- **环境上下文**: 包含部署和配置信息

### 3. 结果解析和存储

- **结构化解析**: 将 AI 响应转换为结构化数据
- **持久化存储**: 保存分析结果和元数据
- **版本管理**: 支持结果版本控制
- **索引优化**: 快速查询和检索

### 4. 分析历史和趋势

- **历史查询**: 按时间、类型、状态过滤
- **趋势分析**: 识别告警模式和趋势
- **统计报告**: 生成分析效果报告
- **性能指标**: 监控分析质量和效率

## 安全考虑

### 1. API 密钥管理

- 使用环境变量存储敏感信息
- 支持密钥轮换和更新
- 避免在日志中记录敏感数据

### 2. 数据隐私

- 敏感数据脱敏处理
- 支持数据加密存储
- 遵循数据保护法规

### 3. 访问控制

- API 调用频率限制
- 用户权限验证
- 审计日志记录

## 监控和运维

### 1. 健康检查

```go
health, err := service.HealthCheck(ctx)
if !health.Healthy {
    // 处理服务不可用情况
}
```

### 2. 指标监控

- **任务处理指标**: 成功率、处理时间、队列长度
- **API 调用指标**: 响应时间、错误率、限流状态
- **资源使用指标**: 内存、CPU、存储使用情况

### 3. 日志记录

- 结构化日志输出
- 分级日志管理
- 错误追踪和调试信息

## 性能优化

### 1. 并发控制

- 限制并发任务数量
- 任务队列管理
- 资源池复用

### 2. 缓存策略

- 分析结果缓存
- 知识库搜索缓存
- 配置信息缓存

### 3. 数据库优化

- 索引优化
- 查询性能调优
- 数据归档策略

## 故障排除

### 常见问题

1. **API 连接失败**
   - 检查网络连接
   - 验证 API 密钥
   - 确认服务端点

2. **分析任务超时**
   - 调整超时配置
   - 检查工作流复杂度
   - 监控资源使用

3. **结果解析错误**
   - 验证响应格式
   - 检查数据结构
   - 更新解析逻辑

### 调试工具

- 启用详细日志
- 使用健康检查接口
- 监控任务状态变化

## 扩展和定制

### 1. 自定义工作流

- 创建特定场景的分析工作流
- 配置工作流参数和映射
- 测试和验证工作流效果

### 2. 知识库管理

- 添加领域特定知识库
- 更新知识库内容
- 优化搜索算法

### 3. 分析算法

- 集成新的分析模型
- 调整分析参数
- 评估分析效果

## 部署指南

### 1. 环境准备

```bash
# 设置环境变量
export DIFY_API_KEY="your-api-key"
export DIFY_APP_TOKEN="your-app-token"
```

### 2. 配置文件

将 `configs/dify.yaml` 复制到部署环境并根据需要调整配置。

### 3. 数据库迁移

确保数据库包含必要的表结构用于存储分析结果。

### 4. 服务启动

```bash
# 启动服务
./alert-agent --config=configs/dify.yaml
```

## 版本兼容性

- **Go 版本**: 1.20+
- **Dify API**: v1.0+
- **数据库**: PostgreSQL 12+ / MySQL 8.0+

## 贡献指南

1. 遵循 Go 编码规范
2. 添加单元测试
3. 更新文档
4. 提交 Pull Request

## 许可证

本项目采用 MIT 许可证，详见 LICENSE 文件。