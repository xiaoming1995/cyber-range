package docker

import (
	"context"
	"cyber-range/pkg/config"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
)

type DockerClient struct {
	cli          *client.Client
	hostID       string  // Docker 主机 ID
	portRangeMin int     // 端口范围最小值
	portRangeMax int     // 端口范围最大值
	memoryLimit  int64   // 内存限制（字节）
	cpuLimit     float64 // CPU 限制（核心数）
}

// NewDockerClient 构造Docker客户端，支持本地/远程模式配置
// 注意：此函数已被弃用，推荐使用 DockerHostManager.GetOrCreateClient
// Deprecated: 使用 DockerHostManager 替代
func NewDockerClient(cfg *config.DockerConfig) (*DockerClient, error) {
	var opts []client.Opt

	// 根据配置的模式选择主机配置（local 或 remote）
	activeHost := cfg.GetActiveHost()

	if activeHost.Host != "" {
		// 使用配置文件中指定的 Docker 主机（远程或自定义）
		opts = append(opts, client.WithHost(activeHost.Host))

		// TLS 配置
		if activeHost.TLSVerify && activeHost.CertPath != "" {
			opts = append(opts, client.WithTLSClientConfig(
				activeHost.CertPath+"/ca.pem",
				activeHost.CertPath+"/cert.pem",
				activeHost.CertPath+"/key.pem",
			))
		}
	} else {
		// 回退到环境变量 (DOCKER_HOST, DOCKER_TLS_VERIFY 等)
		opts = append(opts, client.FromEnv)
	}

	opts = append(opts, client.WithAPIVersionNegotiation())

	c, err := client.NewClientWithOpts(opts...)
	if err != nil {
		return nil, err
	}

	return &DockerClient{
		cli:          c,
		hostID:       "legacy-config",
		portRangeMin: cfg.PortRangeMin,
		portRangeMax: cfg.PortRangeMax,
		memoryLimit:  cfg.MemoryLimit,
		cpuLimit:     cfg.CPULimit,
	}, nil
}

// AllocatePort 从配置的端口范围中随机分配一个端口
func (d *DockerClient) AllocatePort() int {
	portRange := d.portRangeMax - d.portRangeMin + 1
	return d.portRangeMin + rand.Intn(portRange)
}

// Ping 实现 ContainerEngine 接口（用于健康检查）
func (d *DockerClient) Ping(ctx context.Context) (interface{}, error) {
	return d.cli.Ping(ctx)
}

// StartContainer 启动容器
func (d *DockerClient) StartContainer(ctx context.Context, imageName string, envVars []string, containerPort int, privileged bool, memoryLimit int64, cpuLimit float64) (string, int, error) {
	// 1. 确保镜像存在（优化：使用 EnsureImage）
	if err := d.EnsureImage(ctx, imageName); err != nil {
		return "", 0, fmt.Errorf("镜像准备失败: %w", err)
	}

	// 2. 分配端口
	allocatedPort := d.AllocatePort()

	// 3. 资源限制优先级: 参数传入 > Docker Host 配置 > 默认值
	effectiveMemory := d.memoryLimit
	effectiveCPU := d.cpuLimit
	if memoryLimit > 0 {
		effectiveMemory = memoryLimit
	}
	if cpuLimit > 0 {
		effectiveCPU = cpuLimit
	}

	// 4. 构建端口配置
	portStr := fmt.Sprintf("%d/tcp", containerPort)
	exposedPorts := nat.PortSet{nat.Port(portStr): struct{}{}}
	portBindings := nat.PortMap{
		nat.Port(portStr): []nat.PortBinding{{
			HostIP:   "0.0.0.0",
			HostPort: fmt.Sprintf("%d", allocatedPort),
		}},
	}

	// 5. 创建容器并设置严格的资源限制
	resp, err := d.cli.ContainerCreate(ctx,
		&container.Config{
			Image:        imageName,
			Env:          envVars,
			ExposedPorts: exposedPorts,
		},
		&container.HostConfig{
			// 关键：资源约束，防止DoS攻击
			Resources: container.Resources{
				Memory:   effectiveMemory,           // 内存限制
				NanoCPUs: int64(effectiveCPU * 1e9), // CPU 限制
			},
			PortBindings: portBindings,
			Privileged:   privileged, // 特权模式
		}, nil, nil, "")

	if err != nil {
		return "", 0, fmt.Errorf("容器创建失败: %w", err)
	}

	// 5. 启动容器
	if err := d.cli.ContainerStart(ctx, resp.ID, container.StartOptions{}); err != nil {
		return "", 0, fmt.Errorf("启动容器失败: %w", err)
	}

	return resp.ID, allocatedPort, nil
}

// EnsureImage 确保镜像存在（不存在则拉取）
func (d *DockerClient) EnsureImage(ctx context.Context, imageName string) error {
	// 1. 检查本地是否已有
	_, _, err := d.cli.ImageInspectWithRaw(ctx, imageName)
	if err == nil {
		// 镜像已存在
		return nil
	}

	// 2. 从仓库拉取
	reader, err := d.cli.ImagePull(ctx, imageName, image.PullOptions{})
	if err != nil {
		return fmt.Errorf("镜像拉取失败: %w", err)
	}
	defer reader.Close()

	// 等待拉取完成
	_, err = io.Copy(io.Discard, reader)
	if err != nil {
		return fmt.Errorf("镜像下载失败: %w", err)
	}

	return nil
}

// HasLocalImage 检查本地是否有镜像
func (d *DockerClient) HasLocalImage(ctx context.Context, imageName string) bool {
	_, _, err := d.cli.ImageInspectWithRaw(ctx, imageName)
	return err == nil
}

// StopContainer 强制停止并删除容器
func (d *DockerClient) StopContainer(ctx context.Context, containerID string) error {
	// 强制停止（跳过优雅关闭，提高安全性）
	if err := d.cli.ContainerStop(ctx, containerID, container.StopOptions{}); err != nil {
		return fmt.Errorf("容器停止失败: %w", err)
	}

	// 删除容器以释放资源
	if err := d.cli.ContainerRemove(ctx, containerID, container.RemoveOptions{Force: true}); err != nil {
		return fmt.Errorf("容器删除失败: %w", err)
	}

	return nil
}

// ContainerStats 容器资源统计（返回给前端的精简结构）
type ContainerStats struct {
	ContainerID   string  `json:"container_id"`
	CPUPercent    float64 `json:"cpu_percent"`
	MemoryUsage   int64   `json:"memory_usage"` // 字节
	MemoryLimit   int64   `json:"memory_limit"` // 字节
	MemoryPercent float64 `json:"memory_percent"`
	NetworkRx     int64   `json:"network_rx"` // 接收字节
	NetworkTx     int64   `json:"network_tx"` // 发送字节
}

// GetContainerStats 获取容器实时资源使用情况
func (d *DockerClient) GetContainerStats(ctx context.Context, containerID string) (*ContainerStats, error) {
	// 获取容器统计信息（单次快照，不是流）
	resp, err := d.cli.ContainerStats(ctx, containerID, false)
	if err != nil {
		return nil, fmt.Errorf("获取容器统计失败: %w", err)
	}
	defer resp.Body.Close()

	// 解析 JSON 统计数据
	var stats container.StatsResponse
	if err := json.NewDecoder(resp.Body).Decode(&stats); err != nil {
		return nil, fmt.Errorf("解析统计数据失败: %w", err)
	}

	// 计算 CPU 使用率
	cpuPercent := 0.0
	cpuDelta := float64(stats.CPUStats.CPUUsage.TotalUsage - stats.PreCPUStats.CPUUsage.TotalUsage)
	systemDelta := float64(stats.CPUStats.SystemUsage - stats.PreCPUStats.SystemUsage)
	if systemDelta > 0 && cpuDelta > 0 {
		cpuPercent = (cpuDelta / systemDelta) * float64(stats.CPUStats.OnlineCPUs) * 100.0
	}

	// 计算内存使用率
	memoryPercent := 0.0
	if stats.MemoryStats.Limit > 0 {
		memoryPercent = float64(stats.MemoryStats.Usage) / float64(stats.MemoryStats.Limit) * 100.0
	}

	// 计算网络 I/O（汇总所有网络接口）
	var networkRx, networkTx int64
	for _, netStats := range stats.Networks {
		networkRx += int64(netStats.RxBytes)
		networkTx += int64(netStats.TxBytes)
	}

	return &ContainerStats{
		ContainerID:   containerID,
		CPUPercent:    cpuPercent,
		MemoryUsage:   int64(stats.MemoryStats.Usage),
		MemoryLimit:   int64(stats.MemoryStats.Limit),
		MemoryPercent: memoryPercent,
		NetworkRx:     networkRx,
		NetworkTx:     networkTx,
	}, nil
}

// GetContainerLogs 获取容器日志
func (d *DockerClient) GetContainerLogs(ctx context.Context, containerID string, tail int) (string, error) {
	// 默认获取最近 200 行
	if tail <= 0 {
		tail = 200
	}
	if tail > 5000 {
		tail = 5000 // 限制最大行数
	}

	options := container.LogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Timestamps: true,
		Tail:       fmt.Sprintf("%d", tail),
	}

	reader, err := d.cli.ContainerLogs(ctx, containerID, options)
	if err != nil {
		return "", fmt.Errorf("获取容器日志失败: %w", err)
	}
	defer reader.Close()

	// Docker 日志有 8 字节头部，需要跳过
	// 格式: [8]byte{STREAM_TYPE, 0, 0, 0, SIZE1, SIZE2, SIZE3, SIZE4}
	var logs []byte
	buf := make([]byte, 8)
	content := make([]byte, 4096)

	for {
		// 读取 8 字节头部
		_, err := reader.Read(buf)
		if err != nil {
			if err == io.EOF {
				break
			}
			break
		}

		// 计算内容长度
		size := int(buf[4])<<24 | int(buf[5])<<16 | int(buf[6])<<8 | int(buf[7])
		if size <= 0 {
			continue
		}

		// 确保缓冲区足够大
		if size > len(content) {
			content = make([]byte, size)
		}

		// 读取日志内容
		n, err := io.ReadFull(reader, content[:size])
		if err != nil && err != io.EOF && err != io.ErrUnexpectedEOF {
			break
		}
		logs = append(logs, content[:n]...)
	}

	return string(logs), nil
}
