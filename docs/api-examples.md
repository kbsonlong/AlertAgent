# AlertAgent API 使用示例

本文档提供了 AlertAgent API 的详细使用示例，包括各种场景下的 curl 命令和响应示例。

## 目录

- [认证](#认证)
- [健康检查](#健康检查)
- [分析管理](#分析管理)
- [通道管理](#通道管理)
- [集群管理](#集群管理)
- [插件管理](#插件管理)
- [错误处理](#错误处理)

## 认证

### 用户登录

```bash
# 用户登录获取访问令牌
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "admin",
    "password": "password123"
  }'
```

**响应示例：**
```json
{
  "status": "success",
  "message": "Login successful",
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "expires_in": 3600,
    "user": {
      "id": "user-123",
      "username": "admin",
      "email": "admin@example.com",
      "roles": ["admin"],
      "created_at": "2024-01-01T00:00:00Z",
      "last_login_at": "2024-01-15T10:30:00Z"
    }
  }
}
```

### 用户注册

```bash
# 注册新用户
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "newuser",
    "email": "newuser@example.com",
    "password": "securepassword123",
    "roles": ["user"]
  }'
```

**响应示例：**
```json
{
  "status": "success",
  "message": "User registered successfully",
  "data": {
    "user_id": "user-456",
    "username": "newuser",
    "email": "newuser@example.com",
    "created_at": "2024-01-15T10:35:00Z"
  }
}
```

## 健康检查

### 系统健康检查

```bash
# 检查系统健康状态
curl -X GET http://localhost:8080/health
```

**响应示例：**
```json
{
  "status": "ok",
  "message": "Alert Agent is running"
}
```

### 分析服务健康检查

```bash
# 检查分析服务健康状态
curl -X GET http://localhost:8080/api/v1/analysis/health
```

**响应示例：**
```json
{
  "status": "success",
  "message": "Analysis service is healthy",
  "data": {
    "service_status": "running",
    "queue_size": 5,
    "active_workers": 3,
    "last_check": "2024-01-15T10:40:00Z"
  }
}
```

## 分析管理

### 提交分析任务

```bash
# 提交告警分析任务
curl -X POST http://localhost:8080/api/v1/analysis/submit \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." \
  -d '{
    "alert_id": 12345,
    "type": "root_cause",
    "priority": 8,
    "timeout": 300,
    "options": {
      "include_historical_data": true,
      "analysis_depth": "deep"
    },
    "callback": "https://webhook.example.com/analysis-complete"
  }'
```

**响应示例：**
```json
{
  "status": "success",
  "message": "Analysis task submitted successfully",
  "data": {
    "task_id": "task-789abc",
    "status": "queued",
    "message": "Task has been queued for processing",
    "created_at": "2024-01-15T10:45:00Z"
  }
}
```

### 获取分析结果

```bash
# 获取分析任务结果
curl -X GET http://localhost:8080/api/v1/analysis/result/task-789abc \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
```

**响应示例：**
```json
{
  "status": "success",
  "message": "Analysis result retrieved successfully",
  "data": {
    "task_id": "task-789abc",
    "status": "completed",
    "result": {
      "root_cause": "Database connection timeout",
      "affected_services": ["user-service", "order-service"],
      "recommendations": [
        "Increase database connection pool size",
        "Add connection retry logic",
        "Monitor database performance metrics"
      ],
      "severity": "high",
      "impact_score": 8.5
    },
    "confidence": 0.92,
    "processing_time": 15420,
    "created_at": "2024-01-15T10:45:00Z",
    "completed_at": "2024-01-15T10:45:15Z"
  }
}
```

### 获取分析进度

```bash
# 获取分析任务进度
curl -X GET http://localhost:8080/api/v1/analysis/progress/task-789abc \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
```

**响应示例：**
```json
{
  "status": "success",
  "message": "Analysis progress retrieved successfully",
  "data": {
    "task_id": "task-789abc",
    "stage": "data_collection",
    "progress": 65.5,
    "message": "Collecting historical alert data",
    "updated_at": "2024-01-15T10:45:10Z"
  }
}
```

### 列出分析任务

```bash
# 列出分析任务（带过滤条件）
curl -X GET "http://localhost:8080/api/v1/analysis/tasks?alert_id=12345&status=completed&limit=10&offset=0" \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
```

**响应示例：**
```json
{
  "status": "success",
  "message": "Analysis tasks retrieved successfully",
  "data": [
    {
      "id": "task-789abc",
      "alert_id": 12345,
      "type": "root_cause",
      "status": "completed",
      "priority": 8,
      "created_at": "2024-01-15T10:45:00Z",
      "updated_at": "2024-01-15T10:45:15Z"
    },
    {
      "id": "task-456def",
      "alert_id": 12345,
      "type": "impact",
      "status": "completed",
      "priority": 6,
      "created_at": "2024-01-15T09:30:00Z",
      "updated_at": "2024-01-15T09:30:25Z"
    }
  ]
}
```

## 通道管理

### 获取通道列表

```bash
# 获取通道列表（带分页和过滤）
curl -X GET "http://localhost:8080/api/v1/channels?limit=10&offset=0&type=slack&status=active" \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
```

**响应示例：**
```json
{
  "status": "success",
  "message": "Channels retrieved successfully",
  "data": {
    "items": [
      {
        "id": "channel-123",
        "name": "Production Alerts",
        "type": "slack",
        "config": {
          "webhook_url": "https://hooks.slack.com/services/...",
          "channel": "#alerts",
          "username": "AlertAgent"
        },
        "status": "active",
        "health_status": "healthy",
        "created_at": "2024-01-01T00:00:00Z",
        "updated_at": "2024-01-15T10:00:00Z"
      }
    ],
    "total": 1,
    "page": 1,
    "page_size": 10,
    "has_next": false,
    "has_prev": false
  }
}
```

### 创建通道

```bash
# 创建新的Slack通道
curl -X POST http://localhost:8080/api/v1/channels \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." \
  -d '{
    "name": "Development Alerts",
    "type": "slack",
    "description": "Slack channel for development environment alerts",
    "config": {
      "webhook_url": "https://hooks.slack.com/services/T00000000/B00000000/XXXXXXXXXXXXXXXXXXXXXXXX",
      "channel": "#dev-alerts",
      "username": "AlertAgent-Dev",
      "icon_emoji": ":warning:"
    }
  }'
```

**响应示例：**
```json
{
  "status": "success",
  "message": "Channel created successfully",
  "data": {
    "id": "channel-456",
    "name": "Development Alerts",
    "type": "slack",
    "config": {
      "webhook_url": "https://hooks.slack.com/services/T00000000/B00000000/XXXXXXXXXXXXXXXXXXXXXXXX",
      "channel": "#dev-alerts",
      "username": "AlertAgent-Dev",
      "icon_emoji": ":warning:"
    },
    "status": "active",
    "health_status": "unknown",
    "created_at": "2024-01-15T11:00:00Z",
    "updated_at": "2024-01-15T11:00:00Z"
  }
}
```

### 获取通道详情

```bash
# 获取指定通道详情
curl -X GET http://localhost:8080/api/v1/channels/channel-123 \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
```

### 更新通道

```bash
# 更新通道配置
curl -X PUT http://localhost:8080/api/v1/channels/channel-123 \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." \
  -d '{
    "name": "Production Alerts - Updated",
    "description": "Updated production alerts channel",
    "config": {
      "webhook_url": "https://hooks.slack.com/services/...",
      "channel": "#prod-alerts",
      "username": "AlertAgent-Prod"
    }
  }'
```

### 测试通道连接

```bash
# 测试通道连接
curl -X POST http://localhost:8080/api/v1/channels/channel-123/test \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
```

**响应示例：**
```json
{
  "status": "success",
  "message": "Channel test completed",
  "data": {
    "success": true,
    "message": "Test message sent successfully",
    "latency": 245,
    "details": {
      "response_code": 200,
      "response_body": "ok"
    },
    "timestamp": 1705315200
  }
}
```

### 获取通道统计信息

```bash
# 获取通道统计信息
curl -X GET "http://localhost:8080/api/v1/channels/channel-123/stats?start_date=2024-01-01&end_date=2024-01-15" \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
```

**响应示例：**
```json
{
  "status": "success",
  "message": "Channel statistics retrieved successfully",
  "data": {
    "channel_id": "channel-123",
    "total_messages": 1250,
    "success_messages": 1198,
    "failed_messages": 52,
    "success_rate": 95.84,
    "avg_response_time": 187,
    "last_message_at": "2024-01-15T10:55:00Z"
  }
}
```

### 删除通道

```bash
# 删除通道
curl -X DELETE http://localhost:8080/api/v1/channels/channel-456 \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
```

**响应示例：**
```json
{
  "status": "success",
  "message": "Channel deleted successfully"
}
```

## 集群管理

### 获取集群列表

```bash
# 获取集群列表
curl -X GET "http://localhost:8080/api/v1/clusters?limit=10&offset=0&type=kubernetes" \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
```

**响应示例：**
```json
{
  "status": "success",
  "message": "Clusters retrieved successfully",
  "data": {
    "items": [
      {
        "id": "cluster-prod-001",
        "name": "Production Kubernetes Cluster",
        "type": "kubernetes",
        "status": "healthy",
        "node_count": 12,
        "version": "v1.28.4",
        "created_at": "2024-01-01T00:00:00Z",
        "updated_at": "2024-01-15T10:30:00Z"
      }
    ],
    "total": 1,
    "page": 1,
    "page_size": 10,
    "has_next": false,
    "has_prev": false
  }
}
```

## 插件管理

### 获取插件列表

```bash
# 获取可用插件列表
curl -X GET http://localhost:8080/api/v1/plugins \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
```

**响应示例：**
```json
{
  "status": "success",
  "message": "Plugins retrieved successfully",
  "data": {
    "plugins": [
      {
        "type": "slack",
        "name": "Slack Notification Plugin",
        "version": "1.2.0",
        "description": "Send notifications to Slack channels via webhooks",
        "schema": {
          "type": "object",
          "properties": {
            "webhook_url": {
              "type": "string",
              "description": "Slack webhook URL"
            },
            "channel": {
              "type": "string",
              "description": "Target channel name"
            },
            "username": {
              "type": "string",
              "description": "Bot username"
            }
          },
          "required": ["webhook_url"]
        },
        "capabilities": ["send_message", "send_attachment", "format_markdown"]
      },
      {
        "type": "email",
        "name": "Email Notification Plugin",
        "version": "1.1.0",
        "description": "Send email notifications via SMTP",
        "schema": {
          "type": "object",
          "properties": {
            "smtp_host": {
              "type": "string",
              "description": "SMTP server host"
            },
            "smtp_port": {
              "type": "integer",
              "description": "SMTP server port"
            },
            "username": {
              "type": "string",
              "description": "SMTP username"
            },
            "password": {
              "type": "string",
              "description": "SMTP password"
            },
            "from_email": {
              "type": "string",
              "description": "Sender email address"
            },
            "to_emails": {
              "type": "array",
              "items": {
                "type": "string"
              },
              "description": "Recipient email addresses"
            }
          },
          "required": ["smtp_host", "smtp_port", "from_email", "to_emails"]
        },
        "capabilities": ["send_html", "send_text", "send_attachment"]
      }
    ]
  }
}
```

## 错误处理

### 400 Bad Request

```bash
# 错误请求示例（缺少必需参数）
curl -X POST http://localhost:8080/api/v1/analysis/submit \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." \
  -d '{
    "type": "root_cause"
  }'
```

**错误响应示例：**
```json
{
  "status": "error",
  "message": "Invalid request parameters",
  "error": {
    "type": "validation_error",
    "code": "MISSING_REQUIRED_FIELD",
    "message": "Field 'alert_id' is required",
    "details": {
      "field": "alert_id",
      "expected_type": "integer"
    }
  }
}
```

### 401 Unauthorized

```bash
# 未授权访问示例（无效或过期的token）
curl -X GET http://localhost:8080/api/v1/channels \
  -H "Authorization: Bearer invalid_token"
```

**错误响应示例：**
```json
{
  "status": "error",
  "message": "Unauthorized access",
  "error": {
    "type": "authentication_error",
    "code": "INVALID_TOKEN",
    "message": "The provided token is invalid or expired"
  }
}
```

### 404 Not Found

```bash
# 资源未找到示例
curl -X GET http://localhost:8080/api/v1/channels/nonexistent-channel \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
```

**错误响应示例：**
```json
{
  "status": "error",
  "message": "Resource not found",
  "error": {
    "type": "not_found_error",
    "code": "CHANNEL_NOT_FOUND",
    "message": "Channel with ID 'nonexistent-channel' not found",
    "details": {
      "resource_type": "channel",
      "resource_id": "nonexistent-channel"
    }
  }
}
```

### 500 Internal Server Error

**错误响应示例：**
```json
{
  "status": "error",
  "message": "Internal server error",
  "error": {
    "type": "internal_error",
    "code": "DATABASE_CONNECTION_FAILED",
    "message": "Unable to connect to database",
    "details": {
      "timestamp": "2024-01-15T11:30:00Z",
      "request_id": "req-123456789"
    }
  }
}
```

## 批量操作示例

### 批量提交分析任务

```bash
# 使用脚本批量提交多个分析任务
for alert_id in 12345 12346 12347; do
  curl -X POST http://localhost:8080/api/v1/analysis/submit \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." \
    -d "{
      \"alert_id\": $alert_id,
      \"type\": \"root_cause\",
      \"priority\": 5
    }"
  echo "Submitted analysis for alert $alert_id"
done
```

### 监控分析进度

```bash
# 监控分析任务进度的脚本
task_id="task-789abc"
while true; do
  response=$(curl -s -X GET http://localhost:8080/api/v1/analysis/progress/$task_id \
    -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...")
  
  progress=$(echo $response | jq -r '.data.progress')
  stage=$(echo $response | jq -r '.data.stage')
  
  echo "Progress: $progress% - Stage: $stage"
  
  if [ "$progress" = "100" ]; then
    echo "Analysis completed!"
    break
  fi
  
  sleep 5
done
```

## 性能测试示例

### 并发请求测试

```bash
# 使用Apache Bench进行并发测试
ab -n 1000 -c 10 -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." \
   http://localhost:8080/api/v1/channels

# 使用curl进行简单的负载测试
for i in {1..100}; do
  curl -X GET http://localhost:8080/health &
done
wait
echo "All health checks completed"
```

## 环境变量配置

```bash
# 设置API基础URL和认证token
export ALERTAGENT_API_BASE="http://localhost:8080"
export ALERTAGENT_TOKEN="eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."

# 使用环境变量的API调用示例
curl -X GET "$ALERTAGENT_API_BASE/api/v1/channels" \
  -H "Authorization: Bearer $ALERTAGENT_TOKEN"
```

## 注意事项

1. **认证令牌**: 所有需要认证的API都需要在请求头中包含有效的JWT令牌
2. **速率限制**: API可能有速率限制，请注意控制请求频率
3. **错误处理**: 始终检查响应状态码和错误信息
4. **数据格式**: 所有请求和响应都使用JSON格式
5. **时区**: 所有时间戳都使用UTC时区
6. **分页**: 列表API支持分页，注意使用limit和offset参数
7. **过滤**: 大多数列表API支持过滤参数，可以减少不必要的数据传输

更多详细信息请参考 [OpenAPI 规范文档](./openapi.yaml)。