package mock

import (
	"context"
	"fmt"
)

// MockDockerClient 用于单元测试的Mock Docker客户端
type MockDockerClient struct {
	// 可配置的返回值
	ShouldFailStart bool
	ShouldFailStop  bool
	NextPort        int
	NextContainerID string
}

func NewMockDockerClient() *MockDockerClient {
	return &MockDockerClient{
		NextPort:        23456,
		NextContainerID: "mock-container-id-123",
	}
}

func (m *MockDockerClient) Ping(ctx context.Context) (interface{}, error) {
	return "pong", nil
}

func (m *MockDockerClient) StartContainer(ctx context.Context, imageName string, envVars []string, memoryLimit int64, cpuLimit float64) (string, int, error) {
	if m.ShouldFailStart {
		return "", 0, fmt.Errorf("模拟的Docker启动失败")
	}
	return m.NextContainerID, m.NextPort, nil
}

func (m *MockDockerClient) StopContainer(ctx context.Context, containerID string) error {
	if m.ShouldFailStop {
		return fmt.Errorf("模拟的Docker停止失败")
	}
	return nil
}

func (m *MockDockerClient) AllocatePort() int {
	return m.NextPort
}
