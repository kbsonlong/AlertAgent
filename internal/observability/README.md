# 观测性模块 (Observability)

本模块提供了完整的观测性解决方案，包括指标收集、分布式追踪、结构化日志和健康检查。

## 功能特性

### 1. 指标收集 (Metrics)
- 基于 Prometheus 的指标收集
- HTTP 请求指标（请求数、响应时间、状态码分布）
- 告警处理指标（处理数量、处理时间、错误率）
- 规则分发指标（分发数量、分发时间、成功率）
- 系统资源指标（CPU、内存、磁盘使用率）
- 业务指标（活跃告警数、集群健康状态）

### 2. 分布式追踪 (Tracing)
- 自定义轻量级追踪实现
- HTTP 请求追踪
- 数据库操作追踪
- 外部服务调用追踪
- 业务操作追踪
- 上下文传播

### 3. 结构化日志 (Logging)
- 基于 Zap 的高性能日志
- 支持 JSON 和 Console 格式
- 文件轮转和压缩
- 上下文感知日志
- 多级别日志（Debug、Info、Warn、Error、Fatal）
- 业务事件和安全事件记录

### 4. 健康检查 (Health)
- 应用健康检查
- 就绪状态检查
- 数据库连接检查
- Redis 连接检查
- 外部服务检查
- 自定义健康检查器

## 快速开始

### 基本使用

```go
package main

import (
    "context"
    "log"
    
    "alert_agent/internal/observability"
    "github.com/gin-gonic/gin"
    "gorm.io/gorm"
)

func main() {
    ctx := context.Background()
    router := gin.Default()
    
    // 假设你已经有了数据库连接
    var db *gorm.DB
    
    // 快速设置观测性
    manager, err := observability.QuickSetup(ctx, router, db)
    if err != nil {
        log.Fatal("Failed to setup observability:", err)
    }
    defer manager.Shutdown(ctx)
    
    // 你的路由和业务逻辑
    router.GET("/api/test", func(c *gin.Context) {
        // 记录业务事件
        manager.LogBusinessEvent(c.Request.Context(), "test_api_called", map[string]interface{}{
            "user_id": "123",
            "action":  "test",
        })
        
        // 更新指标
        manager.UpdateActiveAlerts(42)
        
        c.JSON(200, gin.H{"message": "success"})
    })
    
    router.Run(":8080")
}
```

### 自定义配置

```go
package main

import (
    "context"
    "log"
    
    "alert_agent/internal/observability"
    "alert_agent/internal/observability/health"
    "alert_agent/internal/observability/logging"
    "alert_agent/internal/observability/metrics"
    "alert_agent/internal/observability/tracing"
    "github.com/gin-gonic/gin"
    "gorm.io/gorm"
)

func main() {
    ctx := context.Background()
    router := gin.Default()
    
    // 自定义配置
    config := &observability.Config{
        Metrics: &metrics.MetricsConfig{
            Enabled: true,
            Port:    9090,
            Path:    "/metrics",
        },
        Tracing: &tracing.TracingConfig{
            Enabled:     true,
            ServiceName: "my-service",
            SampleRate:  0.1,
            MaxSpans:    10000,
        },
        Logging: &logging.LoggingConfig{
            Level:      "info",
            Format:     "json",
            Output:     "both",
            FilePath:   "logs/app.log",
            MaxSize:    100,
            MaxBackups: 5,
            MaxAge:     30,
            Compress:   true,
        },
        Health: &health.HealthConfig{
            Version: "1.0.0",
            ExternalServices: []health.ExternalServiceConfig{
                {
                    Name:    "redis",
                    URL:     "http://redis:6379/ping",
                    Timeout: 5,
                },
            },
        },
    }
    
    // 验证配置
    if err := observability.ValidateConfig(config); err != nil {
        log.Fatal("Invalid config:", err)
    }
    
    // 初始化观测性
    var db *gorm.DB
    manager, err := observability.InitializeObservability(ctx, config, db)
    if err != nil {
        log.Fatal("Failed to initialize observability:", err)
    }
    defer manager.Shutdown(ctx)
    
    // 设置中间件
    observability.SetupObservabilityWithGin(router, manager)
    
    // 启动指标服务器
    if err := manager.StartMetricsServer(); err != nil {
        log.Printf("Failed to start metrics server: %v", err)
    }
    
    router.Run(":8080")
}
```

## 中间件使用

观测性模块提供了多个 Gin 中间件：

```go
// 创建中间件
middleware := observability.NewObservabilityMiddleware(manager)

// 单独使用中间件
router.Use(middleware.HTTPMetricsMiddleware())     // HTTP 指标
router.Use(middleware.HTTPTracingMiddleware())     // HTTP 追踪
router.Use(middleware.HTTPLoggingMiddleware())     // HTTP 日志
router.Use(middleware.ErrorHandlingMiddleware())   // 错误处理
router.Use(middleware.RecoveryMiddleware())        // 恢复中间件
router.Use(middleware.SecurityEventMiddleware())   // 安全事件

// 业务事件中间件
router.Use(middleware.BusinessEventMiddleware("user_api"))

// 或者一次性应用所有中间件
middleware.ApplyAllMiddleware(router)
```

## 手动记录指标和事件

```go
// 记录 HTTP 请求指标
manager.RecordHTTPRequest("GET", "/api/users", 200, time.Millisecond*150)

// 记录告警处理指标
manager.RecordAlertProcessed("critical", "resolved", time.Second*5)

// 记录规则分发指标
manager.RecordRuleDistribution("cluster-1", "success", 10, time.Second*2)

// 更新业务指标
manager.UpdateActiveAlerts(25)
manager.UpdateClusterHealth("cluster-1", true)

// 记录业务事件
manager.LogBusinessEvent(ctx, "user_login", map[string]interface{}{
    "user_id": "123",
    "method":  "oauth",
})

// 记录安全事件
manager.LogSecurityEvent(ctx, "failed_login", "high", map[string]interface{}{
    "user_id":    "123",
    "ip_address": "192.168.1.100",
    "attempts":   3,
})

// 记录错误
manager.LogError(ctx, err, "database_operation", map[string]interface{}{
    "table":     "users",
    "operation": "insert",
})
```

## 分布式追踪

```go
// HTTP 请求追踪
ctx, span := manager.TraceHTTPRequest(ctx, "GET", "/api/users", "Mozilla/5.0")
defer span.Finish()

// 数据库操作追踪
ctx, span = manager.TraceDBOperation(ctx, "SELECT", "users")
defer span.Finish()

// 业务操作追踪
ctx, span = manager.TraceBusinessOperation(ctx, "process_alert", map[string]interface{}{
    "alert_id": "alert-123",
    "severity": "critical",
})
defer span.Finish()

// 在 span 中添加标签和日志
span.SetTag("user_id", "123")
span.LogFields(map[string]interface{}{
    "event": "processing_started",
    "timestamp": time.Now(),
})
```

## 健康检查

```go
// 检查应用健康状态
healthReport := manager.CheckHealth(ctx)
if healthReport.Status == "healthy" {
    fmt.Println("Application is healthy")
}

// 检查就绪状态
readinessReport := manager.CheckReadiness(ctx)
if readinessReport.Ready {
    fmt.Println("Application is ready")
}
```

## 内置端点

观测性模块自动注册以下端点：

- `GET /health` - 健康检查端点
- `GET /ready` - 就绪检查端点
- `GET /metrics` - Prometheus 指标端点（如果启用）
- `GET /metrics-info` - 指标信息端点

## 环境配置

模块提供了针对不同环境的预设配置：

```go
// 生产环境配置
config := observability.CreateProductionConfig()

// 开发环境配置
config := observability.CreateDevelopmentConfig()

// 测试环境配置
config := observability.CreateTestConfig()

// 从环境变量获取配置
config := observability.GetConfigFromEnvironment()
```

## 最佳实践

1. **指标命名**: 使用一致的命名约定，如 `alertagent_http_requests_total`
2. **日志级别**: 生产环境使用 `info` 级别，开发环境使用 `debug` 级别
3. **追踪采样**: 生产环境使用较低的采样率（如 10%），开发环境可以使用 100%
4. **健康检查**: 为所有外部依赖添加健康检查
5. **错误处理**: 始终记录错误并包含足够的上下文信息
6. **性能监控**: 监控关键业务指标和系统资源使用情况

## 故障排除

### 常见问题

1. **指标服务器启动失败**
   - 检查端口是否被占用
   - 确认防火墙设置
   - 验证配置文件

2. **日志文件无法创建**
   - 检查文件路径权限
   - 确认目录是否存在
   - 验证磁盘空间

3. **健康检查失败**
   - 检查外部服务连接
   - 验证数据库连接
   - 确认网络配置

### 调试模式

```go
// 启用调试日志
config.Logging.Level = "debug"

// 启用完整追踪
config.Tracing.SampleRate = 1.0

// 禁用指标收集（如果需要）
config.Metrics.Enabled = false
```

## 扩展

### 自定义健康检查器

```go
type CustomHealthChecker struct {
    name string
    checkFunc func(context.Context) error
}

func (c *CustomHealthChecker) Name() string {
    return c.name
}

func (c *CustomHealthChecker) Check(ctx context.Context) error {
    return c.checkFunc(ctx)
}

// 注册自定义健康检查器
customChecker := &CustomHealthChecker{
    name: "custom_service",
    checkFunc: func(ctx context.Context) error {
        // 自定义检查逻辑
        return nil
    },
}
manager.GetHealthManager().RegisterHealthChecker(customChecker)
```

### 自定义指标

```go
// 获取 Prometheus 注册表
registry := manager.GetMetricsManager().GetRegistry()

// 创建自定义指标
customCounter := prometheus.NewCounterVec(
    prometheus.CounterOpts{
        Name: "my_custom_operations_total",
        Help: "Total number of custom operations",
    },
    []string{"operation_type"},
)

// 注册指标
registry.MustRegister(customCounter)

// 使用指标
customCounter.WithLabelValues("type_a").Inc()
```