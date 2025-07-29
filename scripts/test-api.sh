#!/bin/bash

# Prometheus 规则分发系统 API 测试脚本

BASE_URL="http://localhost:8080/api/v1"

echo "=== Prometheus 规则分发系统 API 测试 ==="
echo "Base URL: $BASE_URL"
echo

# 测试健康检查
echo "1. 测试健康检查"
curl -s "http://localhost:8080/health" | jq .
echo
echo

# 测试创建规则
echo "2. 测试创建规则"
curl -X POST "$BASE_URL/rules" \
  -H "Content-Type: application/json" \
  -d '{
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
  }' | jq .
echo
echo

# 测试获取规则列表
echo "3. 测试获取规则列表"
curl -s "$BASE_URL/rules?limit=10&offset=0" | jq .
echo
echo

# 测试验证规则
echo "4. 测试验证规则"
curl -X POST "$BASE_URL/rules/validate" \
  -H "Content-Type: application/json" \
  -d '{
    "expression": "cpu_usage > 80",
    "for_duration": "5m"
  }' | jq .
echo
echo

# 测试创建规则组
echo "5. 测试创建规则组"
curl -X POST "$BASE_URL/rule-groups" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "system_alerts",
    "cluster_id": "prod-cluster-1",
    "description": "System monitoring alerts",
    "interval": "30s"
  }' | jq .
echo
echo

# 测试获取规则组列表
echo "6. 测试获取规则组列表"
curl -s "$BASE_URL/rule-groups?limit=10&offset=0" | jq .
echo
echo

# 测试获取分发记录
echo "7. 测试获取分发记录"
curl -s "$BASE_URL/distributions?limit=10&offset=0" | jq .
echo
echo

# 测试检测冲突
echo "8. 测试检测冲突"
curl -s "$BASE_URL/conflicts/detect?cluster_id=prod-cluster-1" | jq .
echo
echo

# 测试获取冲突列表
echo "9. 测试获取冲突列表"
curl -s "$BASE_URL/conflicts?status=unresolved" | jq .
echo
echo

# 测试获取规则统计
echo "10. 测试获取规则统计"
curl -s "$BASE_URL/stats/rules?cluster_id=prod-cluster-1" | jq .
echo
echo

echo "=== API 测试完成 ==="