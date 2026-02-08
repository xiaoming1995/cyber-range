#!/bin/bash

echo "🐳 启动本地 Docker Registry..."

# 创建数据目录
mkdir -p ~/cyber-range-registry

# 检查是否已有Registry运行
if docker ps | grep -q "cyber-range-registry"; then
    echo "✅ Registry 已在运行"
    echo "📊 Registry URL: http://localhost:5000"
    exit 0
fi

# 启动 Registry 容器
docker run -d \
  --name cyber-range-registry \
  --restart=always \
  -p 5000:5000 \
  -v ~/cyber-range-registry:/var/lib/registry \
  registry:2

if [ $? -eq 0 ]; then
    echo "✅ Registry 已启动在 http://localhost:5000"
    echo "📊 查看镜像列表: curl http://localhost:5000/v2/_catalog"
    echo "📁 数据目录: ~/cyber-range-registry"
else
    echo "❌ Registry 启动失败"
    exit 1
fi
