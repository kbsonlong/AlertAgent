# Worker 配置说明

## 概述

AlertAgent 支持通过配置文件或环境变量来控制队列工作器（Worker）的启用状态和并发数。这个功能允许您在不同的环境中灵活地控制后台任务处理能力。

## 配置选项

### 配置文件设置

在 `config/config.yaml` 中添加以下配置：

```yaml
worker:
  enabled: true      # 是否启用worker，默认为true
  concurrency: 2     # worker并发数，默认为2
```

### 环境变量设置

您也可以通过环境变量来覆盖配置文件中的设置：

- `WORKER_ENABLED`: 控制是否启用worker（true/false）
- `WORKER_CONCURRENCY`: 设置worker并发数（正整数）

## 使用示例

### 1. 启用Worker（默认配置）

```bash
# 使用默认配置启动
go run cmd/main.go
```

### 2. 禁用Worker

```bash
# 通过环境变量禁用worker
WORKER_ENABLED=false go run cmd/main.go
```

或者修改配置文件：

```yaml
worker:
  enabled: false
  concurrency: 2
```

### 3. 自定义并发数

```bash
# 设置并发数为5
WORKER_CONCURRENCY=5 go run cmd/main.go

# 同时设置启用状态和并发数
WORKER_ENABLED=true WORKER_CONCURRENCY=8 go run cmd/main.go
```

## 日志输出

### Worker启用时

```json
{"level":"INFO","msg":"Worker is enabled, starting queue worker...","concurrency":2}
{"level":"INFO","msg":"Starting worker","worker_id":"xxx","concurrency":2,"queues":["ai_analysis","notification","config_sync"]}
{"level":"INFO","msg":"Worker loop started","worker_id":"xxx-0"}
{"level":"INFO","msg":"Worker loop started","worker_id":"xxx-1"}
```

### Worker禁用时

```json
{"level":"INFO","msg":"Worker is disabled, skipping queue worker startup"}
```

## API 查看配置

您可以通过API接口查看当前的worker配置：

```bash
curl http://localhost:8080/api/v1/settings/config
```

响应示例：

```json
{
  "status": "success",
  "data": {
    "config": {
      "worker": {
        "enabled": true,
        "concurrency": 2
      }
    }
  }
}
```

## 使用场景

### 1. 开发环境

在开发环境中，您可能希望禁用worker以避免后台任务干扰调试：

```bash
WORKER_ENABLED=false go run cmd/main.go
```

### 2. 生产环境

在生产环境中，根据服务器性能调整并发数：

```yaml
# config.prod.yaml
worker:
  enabled: true
  concurrency: 8  # 高性能服务器使用更高并发数
```

### 3. 测试环境

在测试环境中，使用较低的并发数以减少资源消耗：

```yaml
# config.test.yaml
worker:
  enabled: true
  concurrency: 1
```

## 注意事项

1. **环境变量优先级**：环境变量的设置会覆盖配置文件中的设置
2. **并发数限制**：建议根据服务器CPU核心数和内存大小合理设置并发数
3. **优雅关闭**：程序会在接收到关闭信号时优雅地停止所有worker
4. **配置热重载**：修改配置文件后，程序会自动重新加载配置（但worker的启用状态需要重启程序才能生效）

## 故障排除

### Worker未启动

检查配置和日志：

1. 确认 `worker.enabled` 设置为 `true`
2. 检查环境变量 `WORKER_ENABLED` 是否被设置为 `false`
3. 查看启动日志中是否有相关错误信息

### 并发数未生效

1. 确认 `worker.concurrency` 设置正确
2. 检查环境变量 `WORKER_CONCURRENCY` 是否覆盖了配置文件设置
3. 重启程序以使配置生效