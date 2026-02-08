package docker

import (
	"context"
	"cyber-range/internal/model"
	"fmt"
	"sync"

	"github.com/docker/docker/client"
)

// DockerHostManager 管理多个 Docker 主机客户端
type DockerHostManager struct {
	clients map[string]*DockerClient // hostID -> DockerClient
	mu      sync.RWMutex
}

// NewDockerHostManager 创建 Docker 主机管理器
func NewDockerHostManager() *DockerHostManager {
	return &DockerHostManager{
		clients: make(map[string]*DockerClient),
	}
}

// GetOrCreateClient 获取或创建指定主机的 Docker 客户端
func (m *DockerHostManager) GetOrCreateClient(ctx context.Context, host *model.DockerHost) (*DockerClient, error) {
	m.mu.RLock()
	cli, exists := m.clients[host.ID]
	m.mu.RUnlock()

	if exists {
		return cli, nil
	}

	// 创建新客户端
	m.mu.Lock()
	defer m.mu.Unlock()

	// 双重检查（避免并发创建）
	if cli, exists := m.clients[host.ID]; exists {
		return cli, nil
	}

	// 构建 Docker 客户端选项
	var opts []client.Opt
	if host.Host != "" {
		opts = append(opts, client.WithHost(host.Host))
	} else {
		opts = append(opts, client.FromEnv)
	}

	if host.TLSVerify && host.CertPath != "" {
		opts = append(opts, client.WithTLSClientConfig(
			host.CertPath+"/ca.pem",
			host.CertPath+"/cert.pem",
			host.CertPath+"/key.pem",
		))
	}

	opts = append(opts, client.WithAPIVersionNegotiation())

	dockerCli, err := client.NewClientWithOpts(opts...)
	if err != nil {
		return nil, fmt.Errorf("创建 Docker 客户端失败: %w", err)
	}

	// 包装为 DockerClient
	dc := &DockerClient{
		cli:          dockerCli,
		hostID:       host.ID,
		portRangeMin: host.PortRangeMin,
		portRangeMax: host.PortRangeMax,
		memoryLimit:  host.MemoryLimit,
		cpuLimit:     host.CPULimit,
	}

	m.clients[host.ID] = dc
	return dc, nil
}

// RemoveClient 移除指定主机的客户端（用于主机删除或更新配置）
func (m *DockerHostManager) RemoveClient(hostID string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if cli, exists := m.clients[hostID]; exists {
		cli.cli.Close()
		delete(m.clients, hostID)
	}
}

// Ping 测试指定主机的连接性
func (m *DockerHostManager) Ping(ctx context.Context, host *model.DockerHost) error {
	cli, err := m.GetOrCreateClient(ctx, host)
	if err != nil {
		return err
	}

	_, err = cli.Ping(ctx)
	return err
}
