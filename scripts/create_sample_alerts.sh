#!/bin/bash

# 设置API基础URL
API_BASE="http://localhost:8080/api/v1"

# 创建CPU使用率过高告警
curl -X POST "${API_BASE}/alerts" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "CPU使用率告警",
    "source": "node-exporter",
    "level": "critical",
    "description": "服务器CPU使用率持续5分钟超过90%",
    "metrics": {
      "cpu_usage": 95.6,
      "duration": "5m"
    }
  }'

# 创建内存不足告警
curl -X POST "${API_BASE}/alerts" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "内存不足告警",
    "title": "内存不足告警",
    "source": "node-exporter",
    "level": "warning",
    "description": "服务器可用内存低于10%",
    "metrics": {
      "memory_available_percent": 8.5
    }
  }'

# 创建磁盘空间告警
curl -X POST "${API_BASE}/alerts" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "磁盘空间告警",
    "title": "磁盘空间告警",
    "source": "node-exporter",
    "level": "critical",
    "description": "根分区使用率超过95%",
    "metrics": {
      "disk_usage_percent": 96.8,
      "mount_point": "/"
    }
  }'

# 创建网络延迟告警
curl -X POST "${API_BASE}/alerts" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "网络延迟告警",
    "title": "网络延迟告警",
    "source": "blackbox-exporter",
    "level": "warning",
    "description": "API服务响应时间超过2秒",
    "metrics": {
      "latency": 2.5,
      "endpoint": "api.example.com"
    }
  }'

# 创建服务宕机告警
curl -X POST "${API_BASE}/alerts" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "服务宕机告警",
    "title": "服务宕机告警",
    "source": "prometheus",
    "level": "critical",
    "description": "用户认证服务无法访问",
    "metrics": {
      "service_status": "down",
      "service_name": "auth-service"
    }
  }'

# 创建数据库连接告警
curl -X POST "${API_BASE}/alerts" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "数据库连接告警",
    "title": "数据库连接告警",
    "source": "mysql-exporter",
    "level": "critical",
    "description": "数据库连接数接近最大限制",
    "metrics": {
      "max_connections": 1000,
      "current_connections": 950
    }
  }'

# 创建SSL证书过期告警
curl -X POST "${API_BASE}/alerts" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "SSL证书过期告警",
    "title": "SSL证书过期告警",
    "source": "cert-manager",
    "level": "warning",
    "description": "域名证书将在7天内过期",
    "metrics": {
      "domain": "example.com",
      "days_until_expiry": 7
    }
  }'

# 创建API错误率告警
curl -X POST "${API_BASE}/alerts" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "API错误率告警",
    "title": "API错误率告警",
    "source": "prometheus",
    "level": "critical",
    "description": "API错误率超过5%",
    "metrics": {
      "error_rate": 6.5,
      "total_requests": 10000
    }
  }'

# 创建队列积压告警
curl -X POST "${API_BASE}/alerts" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "消息队列积压告警",
    "title": "消息队列积压告警",
    "source": "rabbitmq-exporter",
    "level": "warning",
    "description": "消息队列中待处理消息数量过多",
    "metrics": {
      "queue_name": "task_queue",
      "message_count": 10000
    }
  }'

# 创建容器重启告警
curl -X POST "${API_BASE}/alerts" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "容器重启告警",
    "title": "容器重启告警",
    "source": "kubernetes",
    "level": "warning",
    "description": "容器在过去1小时内重启次数过多",
    "metrics": {
      "container_name": "web-server",
      "restart_count": 5,
      "period": "1h"
    }
  }'

echo "已创建10个示例告警记录"