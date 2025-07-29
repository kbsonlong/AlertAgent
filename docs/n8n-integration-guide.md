# n8n 集成指南

本文档介绍如何在 AlertAgent 中集成和使用 n8n 工作流自动化平台进行告警分析。

## 概述

n8n 集成允许 AlertAgent 通过工作流自动化的方式处理告警，支持：

- 异步告警分析
- 批量告警处理
- 工作流执行监控
- 回调处理
- 分析历史记录
- 执行指标统计

## 架构设计

### 核心组件

1. **N8NClient**: n8n HTTP 客户端，负责与 n8n API 通信
2. **N8NWorkflowManager**: 工作流管理器，处理工作流的触发、监控和回调
3. **N8NAnalysisService**: 分析服务，提供高级的告警分析功能
4. **N8NHandler**: HTTP 处理器，提供 REST API 接口
5. **AlertRepository**: 告警数据仓储
6. **N8NWorkflowExecutionRepository**: 工作流执行记录仓储

### 数据流

```
告警产生 -> 分析服务 -> 工作流管理器 -> n8n 客户端 -> n8n 平台
    ↓           ↑              ↑              ↑
数据库存储 <- 回调处理 <- 执行监控 <- 状态更新
```

## 快速开始

### 1. 环境准备

#### 安装 n8n

```bash
# 使用 Docker 运行 n8n
docker run -it --rm \
  --name n8n \
  -p 5678:5678 \
  -e GENERIC_TIMEZONE="Asia/Shanghai" \
  -e TZ="Asia/Shanghai" \
  n8nio/n8n
```

#### 配置环境变量

```bash
export N8N_BASE_URL="http://localhost:5678"
export N8N_API_KEY="your-n8n-api-key"
export DB_HOST="localhost"
export DB_USER="alertagent"
export DB_PASSWORD="password"
export DB_NAME="alertagent"
export PORT="8080"
```

### 2. 代码集成

#### 初始化 n8n 容器

```go
package main

import (
    "alert_agent/internal/infrastructure/di"
    "gorm.io/gorm"
    "go.uber.org/zap"
)

func main() {
    // 初始化数据库和日志
    db := initDatabase()
    logger := initLogger()
    
    // 创建 n8n 容器
    n8nContainer := di.NewN8NContainer(db, logger)
    
    // 初始化组件
    n8nContainer.Initialize(
        "http://localhost:5678", // n8n base URL
        "your-api-key",          // n8n API key
    )
    
    // 获取服务实例
    analysisService := n8nContainer.GetN8NAnalysisService()
    workflowManager := n8nContainer.GetWorkflowManager()
}
```

#### 注册路由

```go
import (
    httpHandlers "alert_agent/internal/interfaces/http"
    "github.com/gin-gonic/gin"
)

func setupRoutes(router *gin.Engine, n8nContainer *di.N8NContainer) {
    v1 := router.Group("/api/v1")
    
    // n8n 分析路由
    n8n := v1.Group("/n8n")
    httpHandlers.RegisterN8NRoutes(n8n, n8nContainer.GetN8NAnalysisService())
    
    // n8n 回调路由
    callbacks := v1.Group("/callbacks")
    httpHandlers.RegisterN8NCallbackRoutes(callbacks, n8nContainer.GetWorkflowManager())
}
```

### 3. 创建工作流模板

在 n8n 平台中创建工作流模板，用于处理告警分析：

1. 登录 n8n 管理界面 (http://localhost:5678)
2. 创建新的工作流
3. 添加 Webhook 节点作为触发器
4. 添加处理节点（如 HTTP Request、Code 等）
5. 配置回调节点，将结果发送回 AlertAgent
6. 保存并激活工作流

## API 接口

### 告警分析

#### 分析单个告警

```http
POST /api/v1/n8n/alerts/{id}/analyze
Content-Type: application/json

{
  "workflow_template_id": "workflow-123"
}
```

响应：
```json
{
  "execution_id": "exec-456",
  "status": "running",
  "message": "Analysis started successfully"
}
```

#### 批量分析告警

```http
POST /api/v1/n8n/alerts/batch-analyze
Content-Type: application/json

{
  "workflow_template_id": "batch-workflow-123",
  "batch_size": 10,
  "process_interval": "5s",
  "max_retries": 3,
  "timeout": "300s",
  "auto_analysis_enabled": true
}
```

### 执行管理

#### 获取执行状态

```http
GET /api/v1/n8n/executions/{execution_id}/status
```

响应：
```json
{
  "execution_id": "exec-456",
  "workflow_id": "workflow-123",
  "status": "completed",
  "started_at": "2024-01-01T10:00:00Z",
  "finished_at": "2024-01-01T10:05:00Z",
  "duration": 300,
  "input_data": {...},
  "output_data": {...}
}
```

#### 取消执行

```http
POST /api/v1/n8n/executions/{execution_id}/cancel
```

#### 重试执行

```http
POST /api/v1/n8n/executions/{execution_id}/retry
```

### 分析历史

#### 获取告警分析历史

```http
GET /api/v1/n8n/alerts/{alert_id}/analysis-history?limit=10
```

### 分析指标

#### 获取分析指标

```http
GET /api/v1/n8n/metrics?start_time=2024-01-01T00:00:00Z&end_time=2024-01-02T00:00:00Z
```

响应：
```json
{
  "total_executions": 100,
  "successful_executions": 85,
  "failed_executions": 10,
  "running_executions": 5,
  "average_execution_time": "120s",
  "time_range": {
    "start_time": "2024-01-01T00:00:00Z",
    "end_time": "2024-01-02T00:00:00Z"
  }
}
```

## 回调处理

### 工作流回调

n8n 工作流完成后，会调用以下回调接口：

```http
POST /api/v1/callbacks/n8n/workflow/{execution_id}
Content-Type: application/json

{
  "execution_id": "exec-456",
  "status": "completed",
  "data": {
    "analysis_result": "...",
    "confidence": 0.95
  },
  "error": null,
  "timestamp": "2024-01-01T10:05:00Z"
}
```

### Webhook 回调

```http
POST /api/v1/callbacks/n8n/webhook/{webhook_id}
Content-Type: application/json

{
  "webhook_id": "webhook-789",
  "event_type": "analysis_completed",
  "data": {...},
  "timestamp": "2024-01-01T10:05:00Z"
}
```

## 配置说明

### N8NAnalysisConfig

```go
type N8NAnalysisConfig struct {
    DefaultWorkflowTemplateID string        `json:"default_workflow_template_id"`
    BatchSize                 int           `json:"batch_size"`
    ProcessInterval           time.Duration `json:"process_interval"`
    MaxRetries                int           `json:"max_retries"`
    Timeout                   time.Duration `json:"timeout"`
    AutoAnalysisEnabled       bool          `json:"auto_analysis_enabled"`
}
```

### HTTPClientConfig

```go
type HTTPClientConfig struct {
    BaseURL string        `json:"base_url"`
    APIKey  string        `json:"api_key"`
    Timeout time.Duration `json:"timeout"`
}
```

### WorkflowManagerConfig

```go
type WorkflowManagerConfig struct {
    MonitorInterval   time.Duration `json:"monitor_interval"`
    MaxRetryAttempts  int           `json:"max_retry_attempts"`
    RetryDelay        time.Duration `json:"retry_delay"`
    ExecutionTimeout  time.Duration `json:"execution_timeout"`
    CallbackTimeout   time.Duration `json:"callback_timeout"`
    MaxConcurrentJobs int           `json:"max_concurrent_jobs"`
    CleanupInterval   time.Duration `json:"cleanup_interval"`
    RetentionPeriod   time.Duration `json:"retention_period"`
}
```

## 最佳实践

### 1. 工作流设计

- 使用幂等性设计，确保重试安全
- 添加适当的错误处理和重试机制
- 设置合理的超时时间
- 使用回调机制及时更新状态

### 2. 性能优化

- 合理设置批处理大小
- 使用异步处理避免阻塞
- 定期清理历史执行记录
- 监控执行指标，及时调整配置

### 3. 错误处理

- 实现重试机制
- 记录详细的错误日志
- 设置告警通知
- 提供手动干预接口

### 4. 安全考虑

- 使用 HTTPS 通信
- 配置 API 密钥认证
- 验证回调请求来源
- 限制访问权限

## 故障排查

### 常见问题

1. **连接失败**
   - 检查 n8n 服务是否正常运行
   - 验证网络连接和防火墙设置
   - 确认 API 密钥配置正确

2. **工作流执行失败**
   - 查看 n8n 执行日志
   - 检查工作流配置
   - 验证输入数据格式

3. **回调超时**
   - 检查回调 URL 配置
   - 验证网络连接
   - 调整超时设置

### 日志分析

```bash
# 查看 n8n 相关日志
grep "n8n" /var/log/alertagent/app.log

# 查看执行状态
curl -X GET "http://localhost:8080/api/v1/n8n/executions/{execution_id}/status"

# 查看分析指标
curl -X GET "http://localhost:8080/api/v1/n8n/metrics"
```

## 扩展开发

### 自定义工作流管理器

```go
type CustomWorkflowManager struct {
    *n8n.WorkflowManager
    customLogic CustomLogic
}

func (m *CustomWorkflowManager) TriggerAnalysisWorkflow(
    ctx context.Context, 
    alertID string, 
    analysisType string, 
    metadata map[string]interface{},
) (*analysis.N8NWorkflowExecution, error) {
    // 自定义逻辑
    metadata = m.customLogic.EnrichMetadata(metadata)
    
    // 调用原始方法
    return m.WorkflowManager.TriggerAnalysisWorkflow(ctx, alertID, analysisType, metadata)
}
```

### 自定义回调处理

```go
func CustomCallbackHandler(workflowManager analysis.N8NWorkflowManager) gin.HandlerFunc {
    return func(c *gin.Context) {
        // 自定义回调处理逻辑
        executionID := c.Param("execution_id")
        
        var req WorkflowCallbackRequest
        if err := c.ShouldBindJSON(&req); err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
            return
        }
        
        // 处理回调
        err := workflowManager.HandleCallback(c.Request.Context(), executionID, req.Data)
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
            return
        }
        
        c.JSON(http.StatusOK, gin.H{"message": "Callback processed successfully"})
    }
}
```

## 参考资料

- [n8n 官方文档](https://docs.n8n.io/)
- [n8n API 文档](https://docs.n8n.io/api/)
- [AlertAgent 架构设计](./alertagent-architecture-redesign.md)
- [快速开始指南](./quick-start.md)