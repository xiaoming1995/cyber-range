package core

import "context"

// ContainerEngine 定义了我们对容器的操作标准
// 以后如果想换成 K8s，只需要重新实现这个接口，业务逻辑不用改
type ContainerEngine interface {
	StartContainer(ctx context.Context, image string, envVars []string, memoryLimit int64, cpuLimit float64) (string, int, error)
	StopContainer(ctx context.Context, containerID string) error
	Ping(ctx context.Context) (interface{}, error)
}
