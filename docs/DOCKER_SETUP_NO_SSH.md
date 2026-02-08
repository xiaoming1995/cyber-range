# 无法 SSH 时的配置指南

**目标服务器**: 66.154.118.171  
**场景**: 本地无法 SSH 到服务器

---

## 📋 需要传输的文件

将以下文件传输到服务器的 `/tmp/` 目录：

```
docker-certs/server/
├── ca.pem
├── server-cert.pem
└── server-key.pem
```

本地路径：`/Users/liujiming/web/cyber-range/docker-certs/server/`

---

## 🚀 快速配置方案

### 方案 A：使用云控制台（推荐）

1. **登录云服务商控制台**（阿里云/腾讯云/AWS 等）
2. **找到服务器** `66.154.118.171`
3. **点击"远程连接"或"VNC"**
4. **上传证书文件到 /tmp/**
5. **执行配置脚本**

#### 上传文件方法：

**方法 1：云控制台文件上传功能**
- 大部分云控制台都有文件上传功能
- 上传 3 个证书文件到 `/tmp/`

**方法 2：使用 rz/sz 命令**
```bash
# 在云控制台终端执行
yum install -y lrzsz  # CentOS/RHEL
# 或
apt-get install -y lrzsz  # Ubuntu/Debian

# 然后使用 rz 命令接收文件
rz
# 选择本地的 3 个证书文件上传
```

**方法 3：使用 base64 编码传输**
```bash
# 在本地编码证书
cat docker-certs/server/ca.pem | base64

# 将输出复制，然后在服务器上解码
echo "粘贴的base64内容" | base64 -d > /tmp/ca.pem

# 对另外两个文件重复此操作
```

#### 执行配置：

```bash
# 在服务器上执行
cd /tmp

# 验证文件
ls -la ca.pem server-cert.pem server-key.pem

# 执行以下命令配置（手动复制粘贴）：

# 1. 部署证书
mkdir -p /etc/docker/certs
cp ca.pem server-cert.pem server-key.pem /etc/docker/certs/
chmod 0400 /etc/docker/certs/server-key.pem
chmod 0444 /etc/docker/certs/ca.pem /etc/docker/certs/server-cert.pem

# 2. 配置 Docker
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

# 3. 配置 systemd
mkdir -p /etc/systemd/system/docker.service.d
cat > /etc/systemd/system/docker.service.d/override.conf << 'EOF'
[Service]
ExecStart=
ExecStart=/usr/bin/dockerd
EOF

# 4. 开放防火墙端口
firewall-cmd --permanent --add-port=2376/tcp  # firewalld
firewall-cmd --reload
# 或
ufw allow 2376/tcp  # ufw

# 5. 重启 Docker
systemctl daemon-reload
systemctl restart docker

# 6. 验证
ss -tlnp | grep 2376
```

---

### 方案 B：使用自动配置脚本

已为您创建了自动配置脚本：`scripts/configure-docker-server.sh`

**步骤**：

1. **上传证书和脚本到服务器**
   - 证书文件 → `/tmp/` 目录
   - 脚本文件 → `/root/` 或任意目录

2. **在服务器上执行**
```bash
chmod +x configure-docker-server.sh
sudo ./configure-docker-server.sh
```

脚本会自动完成所有配置。

---

### 方案 C：通过其他途径传输

#### 1. 使用 FTP/SFTP 客户端
- FileZilla
- WinSCP
- Cyberduck

#### 2. 使用对象存储
```bash
# 在本地上传到对象存储（如阿里云 OSS、腾讯云 COS）
ossutil cp docker-certs/server/* oss://your-bucket/docker-certs/

# 在服务器上下载
ossutil cp -r oss://your-bucket/docker-certs/ /tmp/
```

#### 3. 使用 HTTP 服务器
```bash
# 在本地启动临时 HTTP 服务器
cd docker-certs/server
python3 -m http.server 8000

# 在服务器上下载（如果可以访问您的本地 IP）
wget http://your-local-ip:8000/ca.pem
wget http://your-local-ip:8000/server-cert.pem
wget http://your-local-ip:8000/server-key.pem
```

---

## ⚠️ 重要提醒

### 云服务器额外步骤

如果是云服务器，配置完成后还需要：

1. **登录云控制台**
2. **进入安全组设置**
3. **添加入站规则**：
   - 协议：TCP
   - 端口：2376
   - 来源：0.0.0.0/0（或指定应用服务器 IP）

---

## 🧪 配置完成后测试

### 在本地测试连接

```bash
# 在开发机执行
docker -H tcp://66.154.118.171:2376 \
  --tlsverify \
  --tlscacert=/Users/liujiming/web/cyber-range/certs/docker/ca.pem \
  --tlscert=/Users/liujiming/web/cyber-range/certs/docker/cert.pem \
  --tlskey=/Users/liujiming/web/cyber-range/certs/docker/key.pem \
  ps
```

**成功标志**：能够列出容器

---

## 📞 需要协助

如果以上方案都无法实施，我可以：

1. 📝 为服务器管理员创建详细的图文教程
2. 🎥 录制配置演示视频
3. 💬 提供实时指导（如果有其他通讯方式）

---

**总结**：即使无法 SSH，通过云控制台或其他传输方式也能完成配置。
