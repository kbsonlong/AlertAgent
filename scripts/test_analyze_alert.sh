#!/bin/bash

# 获取认证token
echo "获取认证token..."
TOKEN=$(curl -s -X POST "http://localhost:8080/api/v1/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"password"}' | jq -r '.data.access_token')

if [ "$TOKEN" = "null" ] || [ -z "$TOKEN" ]; then
  echo "认证失败"
  exit 1
fi

echo "Token获取成功: ${TOKEN:0:20}..."

# 获取alert详情
echo "\n获取Alert ID 3的详情..."
ALERT_DATA=$(curl -s -X GET "http://localhost:8080/api/v1/alerts/3" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json")

echo "Alert数据:"
echo "$ALERT_DATA" | jq .

# 提取alert数据
ALERT_JSON=$(echo "$ALERT_DATA" | jq '.data')

echo "\n调用analyze接口..."
# 调用analyze接口
ANALYZE_RESULT=$(curl -v -X POST "http://localhost:8080/api/v1/alerts/3/analyze" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d "$ALERT_JSON")

echo "\nAnalyze结果:"
echo "$ANALYZE_RESULT"

echo "\n测试完成"