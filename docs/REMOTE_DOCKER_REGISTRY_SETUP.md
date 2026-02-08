# 远程 Docker 主机配置指南

本文档说明如何配置远程 Docker 服务器以信任本地 Mac 上的私有 Registry。

## 前提条件

- 本地 Mac 已启动 Registry（端口 5000）
- 获取本地 Mac 的局域网 IP 地址

获取 Mac IP 地址：
```bash
ifconfig | grep "inet " | grep -v 127.0.0.1 | awk '{print $2}'
```

假设 IP 为 `192.168.1.100`

---

## 远程服务器配置步骤

### 1. SSH 登录远程服务器

```bash
ssh root@your-remote-server
```

### 2. 编辑 Docker 配置文件

```bash
sudo vi /etc/docker/daemon.json
```

### 3. 添加以下内容

**重要**：替换 `192.168.1.100` 为您的本地 Mac IP

```json
{
  "insecure-registries": ["192.168.1.100:5000"]
}
```

如果文件已有其他配置，合并为：

```json
{
  "insecure-registries": ["192.168.1.100:5000"],
  "其他配置": "..."
}
```

### 4. 重启 Docker

```bash
sudo systemctl restart docker
```

### 5. 验证配置

```bash
docker info | grep -A 10 "Insecure Registries"
```

应显示：
```
Insecure Registries:
  192.168.1.100:5000
```

### 6. 测试拉取镜像

```bash
# 假设已从 Mac 推送了镜像 web-xss:v1.0
docker pull 192.168.1.100:5000/web-xss:v1.0
```

如果成功则配置正确！

---

## 故障排查

### 问题 1：连接超时

**症状**：
```
Error response from daemon: Get http://192.168.1.100:5000/v2/: dial tcp 192.168.1.100:5000: connect: connection refused
```

**解决方案**：
1. 检查本地 Mac Registry 是否运行: `docker ps | grep registry`
2. 检查防火墙是否开放 5000 端口
3. 检查 IP 地址是否正确

### 问题 2：TLS 证书错误

**症状**：
```
Error response from daemon: Get https://192.168.1.100:5000/v2/: http: server gave HTTP response to HTTPS client
```

**解决方案**：
确认已将 Registry URL 添加到 `insecure-registries` 列表中

### 问题 3：镜像推送/拉取速度慢

**原因**：网络带宽限制

**优化**：
- 使用有线连接代替 Wi-Fi
- 确保局域网内无其他大流量任务

---

## 多台远程服务器配置

如有多台远程服务器，在每台服务器上重复以上 1-6 步骤

---

## 安全建议

- 私有 Registry 仅在局域网内使用
- 不要将本地 Mac IP 暴露到公网
- 生产环境建议使用 HTTPS + 认证的 Registry

---

##需要帮助？

查看官方文档：https://docs.docker.com/registry/insecure/
