#!/bin/bash

echo "🔍 Docker 主机下拉框问题排查脚本"
echo "========================================"
echo ""

# 1. 检查后端服务
echo "【1/4】检查后端服务..."
if lsof -ti:8080 > /dev/null 2>&1; then
    echo "  ✅ 后端服务运行中 (端口 8080)"
else
    echo "  ❌ 后端服务未运行！"
    echo "  💡 请运行: go run cmd/api/main.go"
    exit 1
fi

# 2. 测试 API 是否正常
echo ""
echo "【2/4】测试 Docker 主机 API..."
TOKEN=$(curl -s -X POST http://localhost:8080/api/admin/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin123"}' | \
  grep -o '"token":"[^"]*"' | cut -d'"' -f4)

if [ -z "$TOKEN" ]; then
    echo "  ❌ 获取 Token 失败！"
    exit 1
fi

RESPONSE=$(curl -s "http://localhost:8080/api/admin/docker-hosts" \
  -H "Authorization: Bearer $TOKEN")

HOST_COUNT=$(echo "$RESPONSE" | grep -o '"id":' | wc -l | tr -d ' ')

if [ "$HOST_COUNT" -gt 0 ]; then
    echo "  ✅ API 正常，返回 $HOST_COUNT 个 Docker 主机"
    echo "$RESPONSE" | jq '.data[] | {name, enabled, is_default}' 2>/dev/null || echo "$RESPONSE"
else
    echo "  ❌ API 返回数据为空！"
    echo "  响应: $RESPONSE"
    exit 1
fi

# 3. 检查前端服务
echo ""
echo "【3/4】检查前端服务..."
if lsof -ti:5173 > /dev/null 2>&1; then
    echo "  ✅ 前端服务运行中 (端口 5173)"
else
    echo "  ⚠️  前端服务未运行"
    echo "  💡 请在 web 目录运行: npm run dev"
fi

# 4. 检查浏览器能否访问
echo ""
echo "【4/4】测试前端代理..."
PROXY_RESPONSE=$(curl -s "http://localhost:5173/api/admin/docker-hosts" \
  -H "Authorization: Bearer $TOKEN" 2>&1)

if echo "$PROXY_RESPONSE" | grep -q "docker-host"; then
    echo "  ✅ 前端代理工作正常"
else
    echo "  ⚠️  前端代理可能有问题"
    echo "  响应: $PROXY_RESPONSE"
fi

echo ""
echo "========================================"
echo "📋 排查结果汇总"
echo "========================================"
echo ""
echo "如果所有检查都通过，但下拉框仍然没有数据，请："
echo ""
echo "1. 打开浏览器开发者工具 (F12)"
echo "2. 切换到 Console 标签页"
echo "3. 刷新页面或打开新建题目页面"
echo "4. 查看是否有以下错误："
echo "   - CORS 跨域错误"
echo "   - 401 认证失败"
echo "   - 网络请求失败"
echo ""
echo "5. 切换到 Network 标签页"
echo "6. 找到 'docker-hosts' 请求"
echo "7. 检查："
echo "   - Status Code (应该是 200)"
echo "   - Response (应该包含 Docker 主机数据)"
echo "   - Request Headers (应该包含 Authorization)"
echo ""
echo "💡 常见问题："
echo "   - 如果看到 401: Token 过期，请重新登录"
echo "   - 如果看到 CORS: Vite 代理可能未生效，重启前端"
echo "   - 如果看到空数组: 数据库中没有 Docker 主机，运行 go run cmd/seed/main.go"
echo ""
