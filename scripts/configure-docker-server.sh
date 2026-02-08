#!/bin/bash

#############################################################################
# Docker TLS 服务器端配置脚本
# 
# 使用说明：
# 1. 将服务器证书文件（ca.pem, server-cert.pem, server-key.pem）
#    上传到服务器的 /tmp/ 目录
# 2. 以 root 或 sudo 权限执行此脚本
# 3. 脚本会自动完成所有配置
#############################################################################

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

echo -e "${BLUE}===== Docker TLS 服务器配置脚本 =====${NC}\n"

# 检查是否为 root
if [ "$EUID" -ne 0 ]; then 
    echo -e "${RED}[ERROR]${NC} 请使用 root 权限或 sudo 运行此脚本"
    exit 1
fi

# 检查证书文件是否存在
echo -e "${BLUE}[1/6]${NC} 检查证书文件..."
if [ ! -f "/tmp/ca.pem" ] || [ ! -f "/tmp/server-cert.pem" ] || [ ! -f "/tmp/server-key.pem" ]; then
    echo -e "${RED}[ERROR]${NC} 证书文件不存在！"
    echo "请确保以下文件在 /tmp/ 目录："
    echo "  - ca.pem"
    echo "  - server-cert.pem"
    echo "  - server-key.pem"
    exit 1
fi
echo -e "${GREEN}[SUCCESS]${NC} 证书文件检查通过"

# 创建证书目录
echo -e "\n${BLUE}[2/6]${NC} 部署证书..."
mkdir -p /etc/docker/certs
cp /tmp/ca.pem /etc/docker/certs/
cp /tmp/server-cert.pem /etc/docker/certs/
cp /tmp/server-key.pem /etc/docker/certs/

# 设置权限
chmod 0400 /etc/docker/certs/server-key.pem
chmod 0444 /etc/docker/certs/ca.pem
chmod 0444 /etc/docker/certs/server-cert.pem

echo -e "${GREEN}[SUCCESS]${NC} 证书已部署到 /etc/docker/certs/"

# 备份原配置
echo -e "\n${BLUE}[3/6]${NC} 配置 Docker Daemon..."
if [ -f "/etc/docker/daemon.json" ]; then
    cp /etc/docker/daemon.json /etc/docker/daemon.json.backup.$(date +%Y%m%d)
    echo -e "${YELLOW}[INFO]${NC} 已备份原配置到 daemon.json.backup"
fi

# 创建新配置
cat > /etc/docker/daemon.json << 'EOF'
{
  "hosts": ["tcp://0.0.0.0:2376", "unix:///var/run/docker.sock"],
  "tls": true,
  "tlscert": "/etc/docker/certs/server-cert.pem",
  "tlskey": "/etc/docker/certs/server-key.pem",
  "tlscacert": "/etc/docker/certs/ca.pem",
  "tlsverify": true
}
EOF
echo -e "${GREEN}[SUCCESS]${NC} daemon.json 已配置"

# 配置 systemd
echo -e "\n${BLUE}[4/6]${NC} 配置 systemd..."
mkdir -p /etc/systemd/system/docker.service.d
cat > /etc/systemd/system/docker.service.d/override.conf << 'EOF'
[Service]
ExecStart=
ExecStart=/usr/bin/dockerd
EOF
echo -e "${GREEN}[SUCCESS]${NC} systemd override 已配置"

# 配置防火墙
echo -e "\n${BLUE}[5/6]${NC} 配置防火墙..."

# 检测防火墙类型
if command -v firewall-cmd &> /dev/null; then
    echo -e "${YELLOW}[INFO]${NC} 检测到 firewalld"
    firewall-cmd --permanent --add-port=2376/tcp
    firewall-cmd --reload
    echo -e "${GREEN}[SUCCESS]${NC} firewalld 规则已添加"
elif command -v ufw &> /dev/null; then
    echo -e "${YELLOW}[INFO]${NC} 检测到 ufw"
    ufw allow 2376/tcp
    echo -e "${GREEN}[SUCCESS]${NC} ufw 规则已添加"
elif command -v iptables &> /dev/null; then
    echo -e "${YELLOW}[INFO]${NC} 使用 iptables"
    iptables -A INPUT -p tcp --dport 2376 -j ACCEPT
    # 尝试保存规则
    if command -v iptables-save &> /dev/null; then
        iptables-save > /etc/iptables/rules.v4 2>/dev/null || true
    fi
    echo -e "${GREEN}[SUCCESS]${NC} iptables 规则已添加"
else
    echo -e "${YELLOW}[WARNING]${NC} 未检测到防火墙，请手动开放 2376 端口"
fi

# 重启 Docker
echo -e "\n${BLUE}[6/6]${NC} 重启 Docker 服务..."
systemctl daemon-reload
systemctl restart docker

# 等待 Docker 启动
sleep 2

# 验证配置
echo -e "\n${BLUE}===== 验证配置 =====${NC}"

if systemctl is-active --quiet docker; then
    echo -e "${GREEN}[SUCCESS]${NC} Docker 服务运行正常"
else
    echo -e "${RED}[ERROR]${NC} Docker 服务启动失败"
    systemctl status docker
    exit 1
fi

if ss -tlnp | grep -q 2376; then
    echo -e "${GREEN}[SUCCESS]${NC} Docker 正在监听 2376 端口"
    ss -tlnp | grep 2376
else
    echo -e "${RED}[ERROR]${NC} Docker 未监听 2376 端口"
    exit 1
fi

# 清理临时文件
rm -f /tmp/ca.pem /tmp/server-cert.pem /tmp/server-key.pem

# 显示摘要
echo -e "\n${GREEN}┌──────────────────────────────────────┐${NC}"
echo -e "${GREEN}│   Docker TLS 配置完成！              │${NC}"
echo -e "${GREEN}└──────────────────────────────────────┘${NC}"

echo -e "\n${YELLOW}配置摘要：${NC}"
echo "  - 监听地址: 0.0.0.0:2376"
echo "  - TLS 模式: 双向认证"
echo "  - 证书目录: /etc/docker/certs/"
echo ""
echo -e "${GREEN}下一步：${NC}"
echo "  在应用服务器上配置客户端证书并测试连接"
echo ""
echo -e "${YELLOW}测试命令（在客户端执行）：${NC}"
echo "  docker -H tcp://$(hostname -I | awk '{print $1}'):2376 \\"
echo "    --tlsverify \\"
echo "    --tlscacert=ca.pem \\"
echo "    --tlscert=cert.pem \\"
echo "    --tlskey=key.pem \\"
echo "    ps"
