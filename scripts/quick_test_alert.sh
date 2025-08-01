#!/bin/bash

# 快速测试告警生成脚本
# 基于swagger.json API接口定义

API_BASE="http://localhost:8080/api/v1"

# 获取认证token
echo "正在获取认证token..."
TOKEN=$(curl -s -X POST "${API_BASE}/auth/login" \
    -H "Content-Type: application/json" \
    -d '{"username":"admin","password":"password"}' | \
    grep -o '"access_token":"[^"]*"' | cut -d'"' -f4)

if [ -z "$TOKEN" ]; then
    echo "❌ 获取token失败，请检查服务是否运行"
    exit 1
fi

echo "✅ 成功获取token"

# 创建测试告警的函数
create_alert() {
    local name="$1"
    local title="$2"
    local content="$3"
    local level="$4"
    local severity="$5"
    local labels="$6"
    local fingerprint="$7"
    
    echo "正在创建告警: $name"
    
    response=$(curl -s -w "\n%{http_code}" -X POST "${API_BASE}/alerts" \
        -H "Content-Type: application/json" \
        -H "Authorization: Bearer $TOKEN" \
        -d "{
            \"name\": \"$name\",
            \"title\": \"$title\",
            \"content\": \"$content\",
            \"source\": \"test-generator\",
            \"level\": \"$level\",
            \"severity\": \"$severity\",
            \"status\": \"new\",
            \"labels\": \"$labels\",
            \"fingerprint\": \"$fingerprint\",
            \"rule_id\": 1,
            \"analysis_result\": \"{}\",
            \"similar_alerts\": \"[]\"
        }")
    
    http_code=$(echo "$response" | tail -n 1)
    response_body=$(echo "$response" | sed '$d')
    
    if [ "$http_code" -eq 201 ] || [ "$http_code" -eq 200 ]; then
        echo "✅ $name 创建成功"
        # 提取告警ID
        alert_id=$(echo "$response_body" | grep -o '"id":[0-9]*' | cut -d':' -f2)
        echo "   告警ID: $alert_id"
    else
        echo "❌ $name 创建失败 (HTTP: $http_code)"
        echo "   错误信息: $response_body"
    fi
    echo ""
}

echo "开始创建测试告警..."
echo ""

# 创建不同类型的测试告警
create_alert \
    "CPU使用率过高" \
    "服务器CPU使用率告警" \
    "服务器CPU使用率持续5分钟超过90%，当前使用率95.6%" \
    "critical" \
    "high" \
    "{\"instance\": \"server-01\", \"job\": \"node-exporter\", \"cpu\": \"total\"}" \
    "cpu-high-usage-$(date +%s)"

create_alert \
    "内存使用率告警" \
    "服务器内存不足告警" \
    "服务器可用内存低于10%，当前可用内存仅8.5%" \
    "high" \
    "medium" \
    "{\"instance\": \"server-02\", \"job\": \"node-exporter\", \"memory\": \"available\"}" \
    "memory-low-$(date +%s)"

create_alert \
    "磁盘空间不足" \
    "服务器磁盘空间告警" \
    "服务器/var分区磁盘使用率超过85%，当前使用率88.3%" \
    "medium" \
    "medium" \
    "{\"instance\": \"server-03\", \"job\": \"node-exporter\", \"device\": \"/dev/sda1\", \"mountpoint\": \"/var\"}" \
    "disk-space-low-$(date +%s)"

create_alert \
    "网络延迟过高" \
    "网络连接延迟告警" \
    "到目标服务器的网络延迟超过100ms，当前延迟156ms" \
    "medium" \
    "low" \
    "{\"instance\": \"server-04\", \"job\": \"blackbox-exporter\", \"target\": \"api.example.com\"}" \
    "network-latency-high-$(date +%s)"

create_alert \
    "服务不可用" \
    "关键服务下线告警" \
    "Web服务无法访问，HTTP状态码500，服务可能已下线" \
    "critical" \
    "high" \
    "{\"instance\": \"web-server-01\", \"job\": \"blackbox-exporter\", \"service\": \"web-api\"}" \
    "service-down-$(date +%s)"

echo "查看创建的告警列表:"
curl -s -H "Authorization: Bearer $TOKEN" "${API_BASE}/alerts" | jq '.data.items[] | {id: .id, name: .name, level: .level, status: .status}' 2>/dev/null || \
curl -s -H "Authorization: Bearer $TOKEN" "${API_BASE}/alerts"

echo ""
echo "=== 测试完成 ==="
echo "您可以使用以下token进行后续API调用:"
echo "export AUTH_TOKEN='$TOKEN'"
echo ""
echo "常用命令:"
echo "# 查看所有告警"
echo "curl -H \"Authorization: Bearer \$AUTH_TOKEN\" \"${API_BASE}/alerts\""
echo ""
echo "# 查看特定告警详情"
echo "curl -H \"Authorization: Bearer \$AUTH_TOKEN\" \"${API_BASE}/alerts/1\""
echo ""
echo "# 分析告警"
echo "curl -X POST -H \"Authorization: Bearer \$AUTH_TOKEN\" -H \"Content-Type: application/json\" \"${API_BASE}/alerts/1/analyze\""
echo ""
echo "# 处理告警"
echo "curl -X PUT -H \"Authorization: Bearer \$AUTH_TOKEN\" -H \"Content-Type: application/json\" \"${API_BASE}/alerts/1\" -d '{\"status\": \"acknowledged\", \"handler\": \"admin\", \"handle_note\": \"正在处理中\"}'"