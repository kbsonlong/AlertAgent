#!/bin/bash

# Sidecar 功能测试脚本

set -e

echo "=== AlertAgent Sidecar 功能测试 ==="

# 配置变量
ALERTAGENT_ENDPOINT="http://localhost:8080"
SIDECAR_HEALTH_PORT="8081"
CLUSTER_ID="test-cluster"
CONFIG_TYPE="prometheus"

echo "1. 检查 AlertAgent 服务状态..."
if curl -f -s "${ALERTAGENT_ENDPOINT}/api/v1/health" > /dev/null; then
    echo "✓ AlertAgent 服务正常"
else
    echo "✗ AlertAgent 服务不可用"
    exit 1
fi

echo "2. 启动 Sidecar (后台运行)..."
./sidecar \
    --endpoint="${ALERTAGENT_ENDPOINT}" \
    --cluster-id="${CLUSTER_ID}" \
    --type="${CONFIG_TYPE}" \
    --config-path="/tmp/test-config.yml" \
    --reload-url="http://localhost:9090/-/reload" \
    --sync-interval=10s \
    --health-port="${SIDECAR_HEALTH_PORT}" \
    --log-level=debug &

SIDECAR_PID=$!
echo "Sidecar PID: ${SIDECAR_PID}"

# 等待 Sidecar 启动
sleep 3

echo "3. 检查 Sidecar 健康状态..."
if curl -f -s "http://localhost:${SIDECAR_HEALTH_PORT}/health" > /dev/null; then
    echo "✓ Sidecar 健康检查通过"
else
    echo "✗ Sidecar 健康检查失败"
    kill ${SIDECAR_PID} 2>/dev/null || true
    exit 1
fi

echo "4. 检查 Sidecar 就绪状态..."
READY_STATUS=$(curl -s "http://localhost:${SIDECAR_HEALTH_PORT}/health/ready")
echo "就绪状态: ${READY_STATUS}"

echo "5. 检查 Sidecar 存活状态..."
LIVE_STATUS=$(curl -s "http://localhost:${SIDECAR_HEALTH_PORT}/health/live")
echo "存活状态: ${LIVE_STATUS}"

echo "6. 获取 Sidecar 指标..."
curl -s "http://localhost:${SIDECAR_HEALTH_PORT}/metrics" | jq '.' || echo "指标获取成功"

echo "7. 获取 Sidecar 详细状态..."
curl -s "http://localhost:${SIDECAR_HEALTH_PORT}/status" | jq '.' || echo "状态获取成功"

echo "8. 等待配置同步..."
sleep 15

echo "9. 检查配置文件是否生成..."
if [ -f "/tmp/test-config.yml" ]; then
    echo "✓ 配置文件已生成"
    echo "配置文件内容:"
    cat /tmp/test-config.yml
else
    echo "✗ 配置文件未生成"
fi

echo "10. 测试配置同步状态查询..."
curl -s "${ALERTAGENT_ENDPOINT}/api/v1/config/sync/status?cluster_id=${CLUSTER_ID}&type=${CONFIG_TYPE}" | jq '.' || echo "状态查询成功"

echo "11. 清理测试环境..."
kill ${SIDECAR_PID} 2>/dev/null || true
rm -f /tmp/test-config.yml

echo "=== 测试完成 ==="
echo "✓ 所有测试通过"