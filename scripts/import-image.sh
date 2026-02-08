#!/bin/bash

# 使用说明
if [ -z "$1" ]; then
    echo "用法: ./import-image.sh <镜像tar文件路径> [镜像名称] [标签]"
    echo ""
    echo "示例:"
    echo "  ./import-image.sh my-challenge.tar web-xss v1.0"
    echo "  ./import-image.sh my-challenge.tar"
    exit 1
fi

TAR_FILE=$1
IMAGE_NAME=${2:-"challenge"}
IMAGE_TAG=${3:-"latest"}
REGISTRY="localhost:5000"

echo "📦 正在导入镜像..."
echo "  文件: $TAR_FILE"
echo "  目标: ${REGISTRY}/${IMAGE_NAME}:${IMAGE_TAG}"
echo ""

# 1. 加载到本地 Docker
echo "[1/4] 加载镜像到本地 Docker..."
docker load -i "$TAR_FILE"

if [ $? -ne 0 ]; then
    echo "❌ 镜像加载失败"
    exit 1
fi

# 2. 获取导入的镜像实际名称
echo "[2/4] 检测镜像名称..."
LOADED_IMAGE=$(docker images --format "{{.Repository}}:{{.Tag}}" | head -n 1)

if [ -z "$LOADED_IMAGE" ] || [ "$LOADED_IMAGE" == "<none>:<none>" ]; then
    echo "❌ 无法检测到加载的镜像"
    echo "提示: 请检查 tar 文件是否有效"
    exit 1
fi

echo "✓ 检测到镜像: $LOADED_IMAGE"

# 3. 重新打标签
echo "[3/4] 重新打标签..."
docker tag "$LOADED_IMAGE" "${REGISTRY}/${IMAGE_NAME}:${IMAGE_TAG}"

if [ $? -ne 0 ]; then
    echo "❌ 打标签失败"
    echo "  源镜像: $LOADED_IMAGE"
    echo "  目标标签: ${REGISTRY}/${IMAGE_NAME}:${IMAGE_TAG}"
    exit 1
fi

# 4. 推送到本地 Registry
echo "[4/4] 推送到 Registry..."
docker push "${REGISTRY}/${IMAGE_NAME}:${IMAGE_TAG}"

if [ $? -ne 0 ]; then
    echo "❌ 推送失败"
    echo "提示: 请确保 Registry 已启动: ./scripts/start-registry.sh"
    exit 1
fi

# 验证
curl -s "http://${REGISTRY}/v2/_catalog" | grep -q "$IMAGE_NAME"

if [ $? -eq 0 ]; then
    echo "✅ 镜像导入完成！"
    echo ""
    echo "📝 镜像信息:"
    echo "  原始镜像: $LOADED_IMAGE"
    echo "  名称: ${IMAGE_NAME}"
    echo "  标签: ${IMAGE_TAG}"
    echo "  完整路径: ${REGISTRY}/${IMAGE_NAME}:${IMAGE_TAG}"
    echo ""
    echo "📋 下一步操作:"
    echo "  1. 在管理后台注册镜像:"
    echo "    curl -X POST http://localhost:8080/api/admin/images \\"
    echo "      -H 'Authorization: Bearer YOUR_TOKEN' \\"
    echo "      -H 'Content-Type: application/json' \\"
    echo "      -d '{\"name\":\"${IMAGE_NAME}\",\"tag\":\"${IMAGE_TAG}\",\"description\":\"题目镜像\"}'"
    echo ""
    echo "  2. 在创建题目时选择该镜像"
else
    echo "⚠️  镜像已推送,但验证失败"
fi
