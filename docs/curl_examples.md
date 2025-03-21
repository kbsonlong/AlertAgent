# 使用curl异步调用AI分析接口

本文档提供了如何使用curl命令异步调用告警分析接口的示例。

## 异步分析告警

以下命令用于提交一个异步分析任务，其中`{alert_id}`需要替换为实际的告警ID：

```bash
curl -X POST "http://localhost:8080/api/v1/alerts/{alert_id}/async-analyze" \
  -H "Content-Type: application/json" \
  -H "Accept: application/json"
```

### 响应示例

```json
{
  "code": 200,
  "msg": "分析任务已加入队列，请稍后查看结果",
  "data": {
    "task_id": 123,
    "submit_time": "2023-06-01 15:04:05"
  }
}
```

## 查询分析状态

提交异步分析任务后，可以使用以下命令查询分析状态，其中`{alert_id}`需要替换为实际的告警ID：

```bash
curl -X GET "http://localhost:8080/api/v1/alerts/{alert_id}/analysis-status" \
  -H "Accept: application/json"
```

### 响应示例 - 处理中

```json
{
  "code": 200,
  "msg": "分析任务处理中",
  "data": {
    "status": "processing"
  }
}
```

### 响应示例 - 已完成

```json
{
  "code": 200,
  "msg": "分析已完成",
  "data": {
    "status": "completed",
    "analysis": "这是AI生成的分析结果..."
  }
}
```

### 响应示例 - 分析失败

```json
{
  "code": 200,
  "msg": "分析失败",
  "data": {
    "status": "failed",
    "error": "分析过程中发生错误"
  }
}
```

## 使用示例

以下是一个完整的使用流程示例：

1. 提交告警ID为123的异步分析任务：

```bash
curl -X POST "http://localhost:8080/api/v1/alerts/123/async-analyze" \
  -H "Content-Type: application/json" \
  -H "Accept: application/json"
```

2. 查询分析状态：

```bash
curl -X GET "http://localhost:8080/api/v1/alerts/123/analysis-status" \
  -H "Accept: application/json"
```

3. 重复步骤2直到获取到完整的分析结果。

## 注意事项

- 请确保替换示例中的`{alert_id}`为实际的告警ID。
- 如果您的API服务不是运行在localhost:8080，请相应地修改URL。
- 异步分析可能需要一些时间才能完成，请耐心等待并定期查询状态。