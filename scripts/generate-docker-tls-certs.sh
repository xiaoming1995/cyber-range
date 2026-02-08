#!/bin/bash

#############################################################################
# Docker TLS 证书自动生成脚本
# 用途：为 Docker 远程连接生成完整的 TLS 证书（双向认证）
# 作者：Cyber Range Team
# 版本：1.0
#############################################################################

set -e  # 遇到错误立即退出

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 打印带颜色的消息
print_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

print_step() {
    echo -e "\n${GREEN}===== $1 =====${NC}\n"
}

# 显示使用说明
usage() {
    cat << EOF
使用方法:
    $0 [选项]

选项:
    -i, --ip <IP>           Docker 服务器 IP 地址 (必需)
    -d, --domain <域名>     Docker 服务器域名 (可选)
    -o, --output <目录>     证书输出目录 (默认: ./docker-certs)
    -p, --password <密码>   CA 私钥密码 (可选，不指定则交互输入)
    --no-password           生成无密码保护的 CA 私钥 (仅用于测试)
    -h, --help              显示此帮助信息

示例:
    # 基本用法（交互式输入密码和证书信息）
    $0 --ip 192.168.1.100

    # 完整用法（使用域名和指定输出目录）
    $0 --ip 192.168.1.100 --domain docker.example.com --output ~/certs

    # 测试用法（无密码保护）
    $0 --ip 192.168.1.100 --no-password

生成的文件:
    服务器证书:
      - ca.pem           (CA 证书)
      - server-cert.pem  (服务器证书)
      - server-key.pem   (服务器私钥)
    
    客户端证书:
      - ca.pem           (CA 证书)
      - cert.pem         (客户端证书)
      - key.pem          (客户端私钥)

EOF
    exit 1
}

# 默认配置
OUTPUT_DIR="./docker-certs"
SERVER_IP=""
SERVER_DOMAIN=""
CA_PASSWORD=""
USE_PASSWORD=true
CERT_DAYS=365

# 解析命令行参数
while [[ $# -gt 0 ]]; do
    case $1 in
        -i|--ip)
            SERVER_IP="$2"
            shift 2
            ;;
        -d|--domain)
            SERVER_DOMAIN="$2"
            shift 2
            ;;
        -o|--output)
            OUTPUT_DIR="$2"
            shift 2
            ;;
        -p|--password)
            CA_PASSWORD="$2"
            shift 2
            ;;
        --no-password)
            USE_PASSWORD=false
            shift
            ;;
        -h|--help)
            usage
            ;;
        *)
            print_error "未知选项: $1"
            usage
            ;;
    esac
done

# 验证必需参数
if [ -z "$SERVER_IP" ]; then
    print_error "必须指定服务器 IP 地址"
    usage
fi

# 验证 IP 格式
if ! [[ $SERVER_IP =~ ^[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}$ ]]; then
    print_error "无效的 IP 地址格式: $SERVER_IP"
    exit 1
fi

# 显示配置信息
print_step "配置信息"
print_info "服务器 IP: $SERVER_IP"
[ -n "$SERVER_DOMAIN" ] && print_info "服务器域名: $SERVER_DOMAIN"
print_info "输出目录: $OUTPUT_DIR"
print_info "证书有效期: $CERT_DAYS 天"
print_info "密码保护: $([ "$USE_PASSWORD" = true ] && echo "是" || echo "否")"

# 创建输出目录
print_step "准备工作"
if [ -d "$OUTPUT_DIR" ]; then
    print_warning "目录已存在: $OUTPUT_DIR"
    read -p "是否清空并继续? [y/N] " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        rm -rf "$OUTPUT_DIR"
        print_info "已清空目录"
    else
        print_error "操作已取消"
        exit 1
    fi
fi

mkdir -p "$OUTPUT_DIR"
cd "$OUTPUT_DIR"
print_success "创建目录: $OUTPUT_DIR"

# 检查 openssl
if ! command -v openssl &> /dev/null; then
    print_error "未找到 openssl 命令，请先安装 OpenSSL"
    exit 1
fi
print_success "OpenSSL 版本: $(openssl version)"

#############################################################################
# 第一步：生成 CA（证书颁发机构）
#############################################################################
print_step "步骤 1/4: 生成 CA 私钥和证书"

if [ "$USE_PASSWORD" = true ]; then
    print_info "生成带密码保护的 CA 私钥..."
    if [ -z "$CA_PASSWORD" ]; then
        # 交互式输入密码
        openssl genrsa -aes256 -out ca-key.pem 4096
    else
        # 使用指定的密码
        openssl genrsa -aes256 -passout pass:"$CA_PASSWORD" -out ca-key.pem 4096
    fi
else
    print_warning "生成无密码保护的 CA 私钥（仅用于测试环境）"
    openssl genrsa -out ca-key.pem 4096
fi
print_success "CA 私钥已生成: ca-key.pem"

print_info "生成 CA 证书..."
if [ "$USE_PASSWORD" = true ]; then
    if [ -z "$CA_PASSWORD" ]; then
        openssl req -new -x509 -days $CERT_DAYS -key ca-key.pem -sha256 -out ca.pem \
            -subj "/C=CN/ST=Beijing/L=Beijing/O=Cyber Range/OU=IT/CN=Docker CA/emailAddress=admin@cyberrange.local"
    else
        openssl req -new -x509 -days $CERT_DAYS -key ca-key.pem -sha256 -out ca.pem \
            -passin pass:"$CA_PASSWORD" \
            -subj "/C=CN/ST=Beijing/L=Beijing/O=Cyber Range/OU=IT/CN=Docker CA/emailAddress=admin@cyberrange.local"
    fi
else
    openssl req -new -x509 -days $CERT_DAYS -key ca-key.pem -sha256 -out ca.pem \
        -subj "/C=CN/ST=Beijing/L=Beijing/O=Cyber Range/OU=IT/CN=Docker CA/emailAddress=admin@cyberrange.local"
fi
print_success "CA 证书已生成: ca.pem"

#############################################################################
# 第二步：生成服务器证书
#############################################################################
print_step "步骤 2/4: 生成服务器证书"

print_info "生成服务器私钥..."
openssl genrsa -out server-key.pem 4096
print_success "服务器私钥已生成: server-key.pem"

print_info "创建服务器证书签名请求..."
if [ -n "$SERVER_DOMAIN" ]; then
    COMMON_NAME="$SERVER_DOMAIN"
else
    COMMON_NAME="$SERVER_IP"
fi
openssl req -subj "/CN=$COMMON_NAME" -sha256 -new -key server-key.pem -out server.csr
print_success "CSR 已生成: server.csr"

print_info "配置 Subject Alternative Name..."
if [ -n "$SERVER_DOMAIN" ]; then
    # 包含域名和 IP
    cat > extfile.cnf << EOF
subjectAltName = DNS:$SERVER_DOMAIN,IP:$SERVER_IP,IP:127.0.0.1
EOF
    print_info "SAN: DNS:$SERVER_DOMAIN, IP:$SERVER_IP, IP:127.0.0.1"
else
    # 仅包含 IP
    cat > extfile.cnf << EOF
subjectAltName = IP:$SERVER_IP,IP:127.0.0.1
EOF
    print_info "SAN: IP:$SERVER_IP, IP:127.0.0.1"
fi

print_info "签发服务器证书..."
if [ "$USE_PASSWORD" = true ]; then
    if [ -z "$CA_PASSWORD" ]; then
        openssl x509 -req -days $CERT_DAYS -sha256 -in server.csr -CA ca.pem -CAkey ca-key.pem \
            -CAcreateserial -out server-cert.pem -extfile extfile.cnf
    else
        openssl x509 -req -days $CERT_DAYS -sha256 -in server.csr -CA ca.pem -CAkey ca-key.pem \
            -passin pass:"$CA_PASSWORD" \
            -CAcreateserial -out server-cert.pem -extfile extfile.cnf
    fi
else
    openssl x509 -req -days $CERT_DAYS -sha256 -in server.csr -CA ca.pem -CAkey ca-key.pem \
        -CAcreateserial -out server-cert.pem -extfile extfile.cnf
fi
print_success "服务器证书已生成: server-cert.pem"

#############################################################################
# 第三步：生成客户端证书
#############################################################################
print_step "步骤 3/4: 生成客户端证书"

print_info "生成客户端私钥..."
openssl genrsa -out key.pem 4096
print_success "客户端私钥已生成: key.pem"

print_info "创建客户端证书签名请求..."
openssl req -subj '/CN=client' -new -key key.pem -out client.csr
print_success "CSR 已生成: client.csr"

print_info "配置客户端证书扩展..."
echo "extendedKeyUsage = clientAuth" > extfile-client.cnf

print_info "签发客户端证书..."
if [ "$USE_PASSWORD" = true ]; then
    if [ -z "$CA_PASSWORD" ]; then
        openssl x509 -req -days $CERT_DAYS -sha256 -in client.csr -CA ca.pem -CAkey ca-key.pem \
            -CAcreateserial -out cert.pem -extfile extfile-client.cnf
    else
        openssl x509 -req -days $CERT_DAYS -sha256 -in client.csr -CA ca.pem -CAkey ca-key.pem \
            -passin pass:"$CA_PASSWORD" \
            -CAcreateserial -out cert.pem -extfile extfile-client.cnf
    fi
else
    openssl x509 -req -days $CERT_DAYS -sha256 -in client.csr -CA ca.pem -CAkey ca-key.pem \
        -CAcreateserial -out cert.pem -extfile extfile-client.cnf
fi
print_success "客户端证书已生成: cert.pem"

#############################################################################
# 第四步：设置权限和清理
#############################################################################
print_step "步骤 4/4: 设置权限和清理"

print_info "设置文件权限..."
chmod 0400 ca-key.pem key.pem server-key.pem
chmod 0444 ca.pem server-cert.pem cert.pem
print_success "权限已设置"

print_info "清理临时文件..."
rm -f client.csr server.csr extfile.cnf extfile-client.cnf ca.srl
print_success "临时文件已清理"

#############################################################################
# 验证证书
#############################################################################
print_step "验证证书"

print_info "验证服务器证书..."
if openssl verify -CAfile ca.pem server-cert.pem > /dev/null 2>&1; then
    print_success "服务器证书验证通过"
else
    print_error "服务器证书验证失败"
    exit 1
fi

print_info "验证客户端证书..."
if openssl verify -CAfile ca.pem cert.pem > /dev/null 2>&1; then
    print_success "客户端证书验证通过"
else
    print_error "客户端证书验证失败"
    exit 1
fi

#############################################################################
# 生成摘要信息
#############################################################################
print_step "证书生成完成"

echo -e "${GREEN}✓ 所有证书和密钥已生成${NC}\n"

# 创建服务器证书目录
SERVER_DIR="server"
CLIENT_DIR="client"
mkdir -p "$SERVER_DIR" "$CLIENT_DIR"

# 复制服务器证书
cp ca.pem server-cert.pem server-key.pem "$SERVER_DIR/"
print_success "服务器证书已复制到: $SERVER_DIR/"

# 复制客户端证书
cp ca.pem cert.pem key.pem "$CLIENT_DIR/"
print_success "客户端证书已复制到: $CLIENT_DIR/"

# 显示文件列表
echo -e "\n${BLUE}生成的文件：${NC}"
echo "────────────────────────────────────────"
ls -lh | grep -E "\.pem$"

# 显示证书信息
echo -e "\n${BLUE}证书信息摘要：${NC}"
echo "────────────────────────────────────────"
echo -e "${YELLOW}CA 证书:${NC}"
openssl x509 -in ca.pem -noout -subject -issuer -dates

echo -e "\n${YELLOW}服务器证书:${NC}"
openssl x509 -in server-cert.pem -noout -subject -issuer -dates
echo -n "SAN: "
openssl x509 -in server-cert.pem -noout -ext subjectAltName | grep -v "X509v3" | xargs

echo -e "\n${YELLOW}客户端证书:${NC}"
openssl x509 -in cert.pem -noout -subject -issuer -dates

# 显示下一步操作指南
cat << EOF

${GREEN}┌─────────────────────────────────────────┐
│         下一步操作指南                  │
└─────────────────────────────────────────┘${NC}

${BLUE}【1】配置远程 Docker 服务器${NC}
   ${YELLOW}复制服务器证书：${NC}
   sudo mkdir -p /etc/docker/certs
   sudo cp $SERVER_DIR/* /etc/docker/certs/
   
   ${YELLOW}编辑 Docker 配置：${NC}
   sudo vim /etc/docker/daemon.json
   
   ${YELLOW}添加内容：${NC}
   {
     "hosts": ["tcp://0.0.0.0:2376", "unix:///var/run/docker.sock"],
     "tls": true,
     "tlscert": "/etc/docker/certs/server-cert.pem",
     "tlskey": "/etc/docker/certs/server-key.pem",
     "tlscacert": "/etc/docker/certs/ca.pem",
     "tlsverify": true
   }
   
   ${YELLOW}重启 Docker：${NC}
   sudo systemctl daemon-reload
   sudo systemctl restart docker

${BLUE}【2】配置应用服务器${NC}
   ${YELLOW}复制客户端证书到应用服务器：${NC}
   scp -r $CLIENT_DIR user@app-server:/path/to/cyber-range/certs/docker
   
   ${YELLOW}修改应用配置 (configs/config.yaml)：${NC}
   docker:
     mode: "remote"
     remote:
       host: "tcp://$SERVER_IP:2376"
       tls_verify: true
       cert_path: "/path/to/certs/docker"

${BLUE}【3】验证连接${NC}
   ${YELLOW}在应用服务器上测试：${NC}
   docker -H tcp://$SERVER_IP:2376 \\
     --tlsverify \\
     --tlscacert=$CLIENT_DIR/ca.pem \\
     --tlscert=$CLIENT_DIR/cert.pem \\
     --tlskey=$CLIENT_DIR/key.pem \\
     ps

${BLUE}【安全提醒】${NC}
   ⚠️  私钥文件 (ca-key.pem, server-key.pem, key.pem) 必须妥善保管
   ⚠️  不要通过不安全的渠道传输私钥
   ⚠️  建议使用 scp 或其他加密方式传输证书
   ⚠️  证书有效期为 $CERT_DAYS 天，请在到期前续期

${GREEN}证书生成成功！${NC}

EOF
