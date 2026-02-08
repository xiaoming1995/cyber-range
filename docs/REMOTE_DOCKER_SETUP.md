# 🌐 远程 Docker 服务器配置完整指南

本文档提供 Cyber Range 靶场系统使用远程 Docker 服务器的完整配置步骤。

---

## 📋 目录

- [方案选择](#方案选择)
- [方案 A：HTTP 模式（内网测试）](#方案-a-http-模式内网测试)
- [方案 B：HTTPS + TLS 模式（生产环境）](#方案-b-https--tls-模式生产环境)
- [故障排查](#故障排查)
- [安全建议](#安全建议)

---

## 方案选择

### 对比表

| 项目 | 方案 A (HTTP) | 方案 B (HTTPS + TLS) |
|------|---------------|----------------------|
| **端口** | 2375 | 2376 |
| **加密** | ❌ 无 | ✅ TLS 1.2+ |
| **认证** | ❌ 无 | ✅ 双向证书认证 |
| **配置难度** | ⭐ 简单 | ⭐⭐⭐ 中等 |
| **安全性** | ⚠️ 低（任何人可连接） | ✅ 高 |
| **适用场景** | 内网测试环境 | 生产环境/跨网络 |
| **配置时间** | 约 5 分钟 | 约 20-30 分钟 |
| **维护成本** | 低 | 中（需管理证书） |

### 选择建议

- **开发/测试阶段**：使用方案 A，快速验证功能
- **生产部署**：使用方案 B，确保安全性
- **内网隔离环境**：可使用方案 A，但需确保网络安全
- **跨公网访问**：必须使用方案 B

---

## 方案 A: HTTP 模式（内网测试）

### ⚠️ 安全警告

**此方案无加密和认证，任何能访问 2375 端口的人都可以完全控制 Docker！**

仅在以下情况使用：
- ✅ 完全可信的内网环境
- ✅ 有防火墙保护
- ✅ 仅用于开发测试

**禁止在生产环境或公网使用！**

---

### 步骤 1：配置远程 Docker 服务器

#### 1.1 备份原配置

```bash
# SSH 登录到远程 Docker 服务器
ssh user@remote-docker-server

# 备份现有配置
sudo cp /etc/docker/daemon.json /etc/docker/daemon.json.backup.$(date +%Y%m%d)
```

#### 1.2 修改 Docker 守护进程配置

```bash
sudo vim /etc/docker/daemon.json
```

添加或修改为以下内容：

```json
{
  "hosts": ["tcp://0.0.0.0:2375", "unix:///var/run/docker.sock"]
}
```

> **说明**：
> - `tcp://0.0.0.0:2375`：监听所有网络接口的 2375 端口
> - `unix:///var/run/docker.sock`：保留本地 socket，不影响本地使用

#### 1.3 修改 systemd 服务配置

Docker 的 systemd 配置可能与 daemon.json 中的 `hosts` 冲突，需要覆盖：

```bash
# 创建 systemd 配置目录
sudo mkdir -p /etc/systemd/system/docker.service.d

# 创建覆盖配置
sudo vim /etc/systemd/system/docker.service.d/override.conf
```

添加以下内容：

```ini
[Service]
ExecStart=
ExecStart=/usr/bin/dockerd
```

> **说明**：
> - 第一行 `ExecStart=` 清空默认启动命令
> - 第二行设置新的启动命令，让 Docker 从 daemon.json 读取配置

#### 1.4 重启 Docker 服务

```bash
# 重新加载 systemd 配置
sudo systemctl daemon-reload

# 重启 Docker
sudo systemctl restart docker

# 检查 Docker 状态
sudo systemctl status docker
```

#### 1.5 验证端口监听

```bash
# 方法 1：使用 netstat
sudo netstat -tlnp | grep 2375

# 方法 2：使用 ss
sudo ss -tlnp | grep 2375

# 预期输出类似：
# tcp6  0  0  :::2375  :::*  LISTEN  1234/dockerd
```

#### 1.6 配置防火墙

根据您的防火墙类型选择：

**使用 firewalld（CentOS/RHEL/Fedora）：**

```bash
# 永久开放 2375 端口
sudo firewall-cmd --permanent --add-port=2375/tcp

# 如果只允许特定 IP 访问（推荐）
sudo firewall-cmd --permanent --add-rich-rule='rule family="ipv4" source address="应用服务器IP/32" port protocol="tcp" port="2375" accept'

# 重新加载防火墙
sudo firewall-cmd --reload

# 验证规则
sudo firewall-cmd --list-all
```

**使用 ufw（Ubuntu/Debian）：**

```bash
# 开放 2375 端口
sudo ufw allow 2375/tcp

# 如果只允许特定 IP 访问（推荐）
sudo ufw allow from 应用服务器IP to any port 2375 proto tcp

# 重新加载
sudo ufw reload

# 查看状态
sudo ufw status
```

**使用 iptables：**

```bash
# 添加规则
sudo iptables -A INPUT -p tcp --dport 2375 -j ACCEPT

# 如果只允许特定 IP（推荐）
sudo iptables -A INPUT -p tcp -s 应用服务器IP --dport 2375 -j ACCEPT

# 保存规则
sudo service iptables save
# 或
sudo iptables-save | sudo tee /etc/iptables/rules.v4
```

---

### 步骤 2：配置 Cyber Range 应用

#### 2.1 修改配置文件

在应用服务器上编辑配置文件：

```bash
vim /Users/liujiming/web/cyber-range/configs/config.yaml
```

修改 `docker` 部分：

```yaml
docker:
  # 切换到远程模式
  mode: "remote"
  
  local:
    host: ""
    tls_verify: false
    cert_path: ""
    
  remote:
    host: "tcp://远程服务器IP:2375"  # 替换为实际 IP
    tls_verify: false
    cert_path: ""
    
  # 通用配置保持不变
  port_range_min: 20000
  port_range_max: 40000
  memory_limit: 134217728  # 128MB
  cpu_limit: 0.5
```

**示例**（假设远程服务器 IP 是 192.168.1.100）：

```yaml
docker:
  mode: "remote"
  
  local:
    host: ""
    tls_verify: false
    cert_path: ""
    
  remote:
    host: "tcp://192.168.1.100:2375"
    tls_verify: false
    cert_path: ""
    
  port_range_min: 20000
  port_range_max: 40000
  memory_limit: 134217728
  cpu_limit: 0.5
```

---

### 步骤 3：验证连接

#### 3.1 使用 Docker CLI 测试

在应用服务器上执行：

```bash
# 测试连接
docker -H tcp://远程服务器IP:2375 ps

# 查看 Docker 版本
docker -H tcp://远程服务器IP:2375 version

# 查看系统信息
docker -H tcp://远程服务器IP:2375 info
```

**预期输出**：
- `ps` 命令应显示容器列表（可能为空）
- `version` 应显示远程 Docker 版本
- `info` 应显示远程系统信息

如果出现错误，请参考[故障排查](#故障排查)章节。

#### 3.2 启动应用测试

```bash
cd /Users/liujiming/web/cyber-range

# 启动应用
go run cmd/api/main.go
```

查看日志输出，确认：
- ✅ 没有 Docker 连接错误
- ✅ 应用成功启动

#### 3.3 功能测试

1. 打开前端界面
2. 选择一个挑战并点击"启动"
3. 在远程服务器上验证容器已创建：

```bash
# 在远程服务器上执行
docker ps

# 预期看到新创建的容器
```

---

## 方案 B: HTTPS + TLS 模式（生产环境）

### 🔒 安全说明

此方案使用 TLS 双向认证：
- ✅ 通信加密（防窃听）
- ✅ 服务器认证（防中间人攻击）
- ✅ 客户端认证（只有持有证书的客户端可连接）

**推荐用于生产环境和跨网络部署。**

---

### 步骤 1：生成 TLS 证书

#### 1.1 创建证书工作目录

在**远程 Docker 服务器**上执行：

```bash
# 创建目录
mkdir -p ~/docker-certs
cd ~/docker-certs

# 设置变量（替换为实际值）
export DOCKER_HOST_IP="远程服务器IP"  # 例如：192.168.1.100
export DOCKER_HOST_DOMAIN="远程服务器域名"  # 可选，例如：docker.example.com
```

#### 1.2 生成 CA（证书颁发机构）

```bash
# 生成 CA 私钥（4096 位，AES256 加密）
openssl genrsa -aes256 -out ca-key.pem 4096

# 输入密码（建议使用强密码，并妥善保管）

# 生成 CA 证书（有效期 365 天）
openssl req -new -x509 -days 365 -key ca-key.pem -sha256 -out ca.pem

# 输入 CA 私钥密码
# 填写证书信息（示例）：
# Country Name: CN
# State or Province: Beijing
# Locality Name: Beijing
# Organization Name: Your Company
# Organizational Unit Name: IT
# Common Name: Docker CA
# Email Address: admin@example.com
```

#### 1.3 生成服务器证书

```bash
# 1. 生成服务器私钥
openssl genrsa -out server-key.pem 4096

# 2. 创建证书签名请求（CSR）
# 使用 IP 地址
openssl req -subj "/CN=${DOCKER_HOST_IP}" -sha256 -new -key server-key.pem -out server.csr

# 或者使用域名
# openssl req -subj "/CN=${DOCKER_HOST_DOMAIN}" -sha256 -new -key server-key.pem -out server.csr

# 3. 配置 Subject Alternative Name (SAN)
cat > extfile.cnf <<EOF
subjectAltName = IP:${DOCKER_HOST_IP},IP:127.0.0.1
EOF

# 如果使用域名，还需添加：
# echo "subjectAltName = DNS:${DOCKER_HOST_DOMAIN},IP:${DOCKER_HOST_IP},IP:127.0.0.1" > extfile.cnf

# 4. 签发服务器证书
openssl x509 -req -days 365 -sha256 -in server.csr -CA ca.pem -CAkey ca-key.pem \
  -CAcreateserial -out server-cert.pem -extfile extfile.cnf

# 输入 CA 私钥密码
```

#### 1.4 生成客户端证书

```bash
# 1. 生成客户端私钥
openssl genrsa -out key.pem 4096

# 2. 创建客户端 CSR
openssl req -subj '/CN=client' -new -key key.pem -out client.csr

# 3. 配置客户端证书扩展
echo "extendedKeyUsage = clientAuth" > extfile-client.cnf

# 4. 签发客户端证书
openssl x509 -req -days 365 -sha256 -in client.csr -CA ca.pem -CAkey ca-key.pem \
  -CAcreateserial -out cert.pem -extfile extfile-client.cnf

# 输入 CA 私钥密码
```

#### 1.5 设置证书权限

```bash
# 移除写权限，防止意外修改
chmod 0400 ca-key.pem key.pem server-key.pem
chmod 0444 ca.pem server-cert.pem cert.pem

# 验证权限
ls -la *.pem
```

#### 1.6 清理临时文件

```bash
rm -f client.csr server.csr extfile.cnf extfile-client.cnf
```

#### 1.7 验证证书

```bash
# 查看 CA 证书信息
openssl x509 -in ca.pem -text -noout

# 查看服务器证书信息
openssl x509 -in server-cert.pem -text -noout

# 查看客户端证书信息
openssl x509 -in cert.pem -text -noout

# 验证证书链
openssl verify -CAfile ca.pem server-cert.pem
openssl verify -CAfile ca.pem cert.pem

# 预期输出：
# server-cert.pem: OK
# cert.pem: OK
```

---

### 步骤 2：配置远程 Docker 服务器

#### 2.1 复制证书到系统目录

```bash
# 创建 Docker 证书目录
sudo mkdir -p /etc/docker/certs

# 复制服务器证书
sudo cp ca.pem server-cert.pem server-key.pem /etc/docker/certs/

# 设置权限
sudo chmod 0400 /etc/docker/certs/server-key.pem
sudo chmod 0444 /etc/docker/certs/ca.pem /etc/docker/certs/server-cert.pem

# 验证
sudo ls -la /etc/docker/certs/
```

#### 2.2 备份配置

```bash
sudo cp /etc/docker/daemon.json /etc/docker/daemon.json.backup.$(date +%Y%m%d)
```

#### 2.3 修改 Docker 守护进程配置

```bash
sudo vim /etc/docker/daemon.json
```

添加或修改为：

```json
{
  "hosts": ["tcp://0.0.0.0:2376", "unix:///var/run/docker.sock"],
  "tls": true,
  "tlscert": "/etc/docker/certs/server-cert.pem",
  "tlskey": "/etc/docker/certs/server-key.pem",
  "tlscacert": "/etc/docker/certs/ca.pem",
  "tlsverify": true
}
```

> **配置说明**：
> - `tcp://0.0.0.0:2376`：TLS 使用 2376 端口（标准）
> - `tls: true`：启用 TLS
> - `tlsverify: true`：要求客户端证书认证
> - 三个证书路径指向服务器证书文件

#### 2.4 配置 systemd

```bash
sudo mkdir -p /etc/systemd/system/docker.service.d
sudo vim /etc/systemd/system/docker.service.d/override.conf
```

```ini
[Service]
ExecStart=
ExecStart=/usr/bin/dockerd
```

#### 2.5 重启 Docker

```bash
sudo systemctl daemon-reload
sudo systemctl restart docker
sudo systemctl status docker
```

#### 2.6 验证端口监听

```bash
sudo netstat -tlnp | grep 2376
# 或
sudo ss -tlnp | grep 2376

# 预期输出：
# tcp6  0  0  :::2376  :::*  LISTEN  1234/dockerd
```

#### 2.7 配置防火墙

```bash
# firewalld
sudo firewall-cmd --permanent --add-port=2376/tcp
sudo firewall-cmd --reload

# ufw
sudo ufw allow 2376/tcp

# 查看状态
sudo firewall-cmd --list-ports  # firewalld
sudo ufw status                  # ufw
```

---

### 步骤 3：配置应用服务器

#### 3.1 传输客户端证书

在**应用服务器**上执行：

```bash
# 创建证书目录
mkdir -p /Users/liujiming/web/cyber-range/certs/docker

# 使用 scp 复制证书（从远程服务器）
scp user@远程服务器IP:~/docker-certs/ca.pem /Users/liujiming/web/cyber-range/certs/docker/
scp user@远程服务器IP:~/docker-certs/cert.pem /Users/liujiming/web/cyber-range/certs/docker/
scp user@远程服务器IP:~/docker-certs/key.pem /Users/liujiming/web/cyber-range/certs/docker/

# 设置权限
chmod 0400 /Users/liujiming/web/cyber-range/certs/docker/key.pem
chmod 0444 /Users/liujiming/web/cyber-range/certs/docker/ca.pem
chmod 0444 /Users/liujiming/web/cyber-range/certs/docker/cert.pem

# 验证
ls -la /Users/liujiming/web/cyber-range/certs/docker/
```

**预期文件列表**：
```
-r--r--r-- ca.pem
-r--r--r-- cert.pem
-r-------- key.pem
```

#### 3.2 修改配置文件

```bash
vim /Users/liujiming/web/cyber-range/configs/config.yaml
```

修改 `docker` 部分：

```yaml
docker:
  mode: "remote"
  
  local:
    host: ""
    tls_verify: false
    cert_path: ""
    
  remote:
    host: "tcp://远程服务器IP:2376"  # ⚠️ 注意是 2376！
    tls_verify: true
    cert_path: "/Users/liujiming/web/cyber-range/certs/docker"
    
  port_range_min: 20000
  port_range_max: 40000
  memory_limit: 134217728
  cpu_limit: 0.5
```

**示例**（假设远程服务器 IP 是 192.168.1.100）：

```yaml
docker:
  mode: "remote"
  
  local:
    host: ""
    tls_verify: false
    cert_path: ""
    
  remote:
    host: "tcp://192.168.1.100:2376"
    tls_verify: true
    cert_path: "/Users/liujiming/web/cyber-range/certs/docker"
    
  port_range_min: 20000
  port_range_max: 40000
  memory_limit: 134217728
  cpu_limit: 0.5
```

---

### 步骤 4：验证 TLS 连接

#### 4.1 使用 Docker CLI 测试

```bash
# 设置证书路径变量
export DOCKER_CERT_PATH=/Users/liujiming/web/cyber-range/certs/docker
export DOCKER_HOST=tcp://远程服务器IP:2376
export DOCKER_TLS_VERIFY=1

# 测试连接
docker ps
docker version
docker info

# 或者每次手动指定参数
docker -H tcp://远程服务器IP:2376 \
  --tlsverify \
  --tlscacert=${DOCKER_CERT_PATH}/ca.pem \
  --tlscert=${DOCKER_CERT_PATH}/cert.pem \
  --tlskey=${DOCKER_CERT_PATH}/key.pem \
  ps
```

**预期输出**：
- ✅ 能够成功连接并列出容器
- ✅ 无 TLS 错误

#### 4.2 测试证书验证

```bash
# 尝试不带证书连接（应该失败）
docker -H tcp://远程服务器IP:2376 ps

# 预期输出：错误信息，提示需要证书
```

#### 4.3 启动应用测试

```bash
cd /Users/liujiming/web/cyber-range

# 清除环境变量（避免干扰）
unset DOCKER_HOST DOCKER_TLS_VERIFY DOCKER_CERT_PATH

# 启动应用
go run cmd/api/main.go
```

查看日志，确认：
- ✅ Docker 客户端初始化成功
- ✅ 无 TLS 错误
- ✅ 应用正常启动

#### 4.4 功能测试

1. 打开前端界面
2. 启动一个挑战
3. 在远程服务器验证：

```bash
# 在远程服务器上
docker ps

# 应该看到新创建的容器
```

---

## 故障排查

### 问题 1：连接被拒绝

**错误信息**：
```
Error response from daemon: dial tcp 192.168.1.100:2375: connect: connection refused
```

**可能原因和解决方法**：

1. **Docker 未监听对应端口**

```bash
# 在远程服务器检查
sudo netstat -tlnp | grep 2375
sudo netstat -tlnp | grep 2376

# 如果没有输出，检查 Docker 配置
sudo systemctl status docker
sudo journalctl -u docker -n 50
```

2. **防火墙阻止**

```bash
# 检查防火墙状态
sudo firewall-cmd --list-all  # firewalld
sudo ufw status verbose        # ufw

# 临时关闭防火墙测试（⚠️ 仅用于排查）
sudo systemctl stop firewalld
# 或
sudo ufw disable
```

3. **云服务器安全组未开放端口**

如果使用阿里云、腾讯云等，需要在控制台开放对应端口。

---

### 问题 2：TLS 握手失败

**错误信息**：
```
error during connect: Get "https://...": x509: certificate signed by unknown authority
```

**解决方法**：

1. **验证证书文件存在**

```bash
ls -la /Users/liujiming/web/cyber-range/certs/docker/
# 应该有 ca.pem, cert.pem, key.pem
```

2. **验证证书内容**

```bash
# 查看证书
openssl x509 -in /Users/liujiming/web/cyber-range/certs/docker/cert.pem -text -noout

# 检查 Subject Alternative Name
openssl x509 -in /Users/liujiming/web/cyber-range/certs/docker/cert.pem -text -noout | grep -A1 "Subject Alternative Name"
```

3. **验证证书链**

```bash
cd /Users/liujiming/web/cyber-range/certs/docker/
openssl verify -CAfile ca.pem cert.pem
# 应输出：cert.pem: OK
```

4. **检查 IP/域名匹配**

确保服务器证书的 SAN 包含您使用的 IP 或域名。

---

### 问题 3：权限错误

**错误信息**：
```
permission denied while trying to connect to the Docker daemon socket
```

**解决方法**：

1. **检查文件权限**

```bash
ls -la /Users/liujiming/web/cyber-range/certs/docker/

# 正确权限应该是：
# -r--r--r-- ca.pem
# -r--r--r-- cert.pem
# -r-------- key.pem
```

2. **修复权限**

```bash
chmod 0444 /Users/liujiming/web/cyber-range/certs/docker/ca.pem
chmod 0444 /Users/liujiming/web/cyber-range/certs/docker/cert.pem
chmod 0400 /Users/liujiming/web/cyber-range/certs/docker/key.pem
```

---

### 问题 4：证书过期

**错误信息**：
```
x509: certificate has expired or is not yet valid
```

**解决方法**：

1. **检查证书有效期**

```bash
openssl x509 -in /Users/liujiming/web/cyber-range/certs/docker/cert.pem -noout -dates

# 输出：
# notBefore=...
# notAfter=...
```

2. **重新生成证书**

参考[步骤 1：生成 TLS 证书](#步骤-1生成-tls-证书)重新生成。

---

### 问题 5：容器端口冲突

**错误信息**：
```
Error starting container: port is already allocated
```

**解决方法**：

1. **检查端口占用**

```bash
# 在远程服务器检查
sudo netstat -tlnp | grep <端口号>
```

2. **调整端口范围**

修改 `config.yaml` 中的 `port_range_min` 和 `port_range_max`。

---

## 安全建议

### 1. 网络安全

#### 方案 A (HTTP)
- ✅ 仅在内网使用
- ✅ 使用防火墙限制访问 IP
- ✅ 定期审计连接日志
- ❌ 禁止暴露到公网

#### 方案 B (TLS)
- ✅ 可用于跨网络部署
- ✅ 定期轮换证书（建议每年）
- ✅ 妥善保管 CA 私钥
- ✅ 监控异常连接

### 2. 证书管理

```bash
# 设置证书过期提醒
# 添加到 crontab
0 0 * * * openssl x509 -in /etc/docker/certs/server-cert.pem -checkend 2592000 \
  || echo "Docker server certificate expires in 30 days" | mail -s "Certificate Alert" admin@example.com
```

### 3. 防火墙规则

```bash
# 仅允许应用服务器访问（推荐）
sudo firewall-cmd --permanent --add-rich-rule='rule family="ipv4" source address="应用服务器IP/32" port protocol="tcp" port="2376" accept'
sudo firewall-cmd --reload
```

### 4. Docker 资源限制

已在 `config.yaml` 中配置：
- `memory_limit: 134217728` (128MB)
- `cpu_limit: 0.5` (0.5 核心)

根据实际情况调整。

### 5. 日志审计

```bash
# 在远程服务器启用 Docker 审计日志
sudo vim /etc/docker/daemon.json

# 添加：
{
  "log-driver": "json-file",
  "log-opts": {
    "max-size": "10m",
    "max-file": "3"
  }
}
```

---

## 快速参考

### 常用命令

```bash
# 查看 Docker 监听端口
sudo netstat -tlnp | grep docker

# 重启 Docker
sudo systemctl restart docker

# 查看 Docker 日志
sudo journalctl -u docker -f

# 测试远程连接（HTTP）
docker -H tcp://IP:2375 ps

# 测试远程连接（TLS）
docker -H tcp://IP:2376 --tlsverify \
  --tlscacert=ca.pem --tlscert=cert.pem --tlskey=key.pem ps

# 查看证书信息
openssl x509 -in cert.pem -text -noout

# 验证证书
openssl verify -CAfile ca.pem cert.pem
```

### 配置文件位置

| 文件 | 路径 |
|------|------|
| Docker 配置 | `/etc/docker/daemon.json` |
| systemd 覆盖 | `/etc/systemd/system/docker.service.d/override.conf` |
| 服务器证书 | `/etc/docker/certs/` |
| 客户端证书 | `/Users/liujiming/web/cyber-range/certs/docker/` |
| 应用配置 | `/Users/liujiming/web/cyber-range/configs/config.yaml` |

---

## 附录

### A. 证书目录结构

**远程服务器**：
```
/etc/docker/certs/
├── ca.pem           # CA 证书
├── server-cert.pem  # 服务器证书
└── server-key.pem   # 服务器私钥
```

**应用服务器**：
```
/Users/liujiming/web/cyber-range/certs/docker/
├── ca.pem    # CA 证书
├── cert.pem  # 客户端证书
└── key.pem   # 客户端私钥
```

### B. 端口说明

| 端口 | 协议 | 说明 |
|------|------|------|
| 2375 | HTTP | Docker API（无加密） |
| 2376 | HTTPS | Docker API（TLS 加密） |
| 20000-40000 | TCP | 容器端口映射范围 |

### C. 配置示例汇总

**方案 A 配置**：
```yaml
docker:
  mode: "remote"
  remote:
    host: "tcp://192.168.1.100:2375"
    tls_verify: false
    cert_path: ""
```

**方案 B 配置**：
```yaml
docker:
  mode: "remote"
  remote:
    host: "tcp://192.168.1.100:2376"
    tls_verify: true
    cert_path: "/Users/liujiming/web/cyber-range/certs/docker"
```

---

## 获取帮助

如遇问题，请检查：
1. 📖 本文档的[故障排查](#故障排查)章节
2. 🐛 Docker 日志：`sudo journalctl -u docker -n 100`
3. 🔍 应用日志：查看应用启动时的输出

---

**文档版本**: 1.0  
**最后更新**: 2026-01-28  
**适用版本**: Cyber Range v1.0+
