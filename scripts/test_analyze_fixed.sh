#!/bin/bash

# 测试修复后的告警分析接口

echo "=== 测试告警分析接口 ==="

# 1. 获取认证token
echo "1. 获取认证token..."
TOKEN_RESPONSE=$(curl -s -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"password"}')

TOKEN=$(echo $TOKEN_RESPONSE | jq -r '.data.access_token')
echo "Token: $TOKEN"

if [ "$TOKEN" = "null" ] || [ -z "$TOKEN" ]; then
    echo "❌ 获取token失败"
    echo "Response: $TOKEN_RESPONSE"
    exit 1
fi

# 2. 获取告警ID 3的详情
echo "\n2. 获取告警详情..."
ALERT_RESPONSE=$(curl -s -X GET http://localhost:8080/api/v1/alerts/3 \
  -H "Authorization: Bearer $TOKEN")

echo "Alert details:"
echo $ALERT_RESPONSE | jq .

# 3. 测试分析接口 - 发送空的JSON对象
echo "\n3. 测试分析接口（发送空JSON）..."
ANALYZE_RESPONSE=$(curl -s -w "\nHTTP_CODE:%{http_code}\n" -X POST http://localhost:8080/api/v1/alerts/3/analyze \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{}')

echo "Analysis response:"
echo "$ANALYZE_RESPONSE"

# 4. 如果上面失败，尝试不发送请求体
echo "\n4. 测试分析接口（无请求体）..."
ANALYZE_RESPONSE2=$(curl -s -w "\nHTTP_CODE:%{http_code}\n" -X POST http://localhost:8080/api/v1/alerts/3/analyze \
  -H "Authorization: Bearer $TOKEN")

echo "Analysis response (no body):"
echo "$ANALYZE_RESPONSE2"

echo "\n=== 测试完成 ==="