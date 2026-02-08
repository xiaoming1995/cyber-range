# 测试容器实例启动（核心业务）指南

## 🎯 测试目标

测试 `StartInstance` 方法的完整业务逻辑：
1. ✅ Docker容器成功启动
2. ✅ Flag正确生成（唯一性）
3. ✅ 端口在配置范围内（20000-40000）
4. ✅ 配额限制生效（每用户最多1个实例）
5. ✅ Redis状态正确存储
6. ✅ 数据库记录创建

---

## 🚧 当前限制

`StartInstance` 方法直接调用 Redis 静态函数（`redisRepo.StoreInstance`），**无法在单元测试中Mock**。

### 解决方案：

### 方案A：集成测试（推荐先用这个）

**需要真实环境：** Redis + MySQL + Docker

**创建文件：** `tests/integration/challenge_flow_test.go`

```go
// +build integration

package integration

import (
	"context"
	"testing"
	"time"
	
	"cyber-range/internal/infra/db"
	"cyber-range/internal/infra/docker"
	"cyber-range/internal/infra/redis"
	"cyber-range/internal/service"
	"cyber-range/pkg/config"
)

func TestStartInstance_FullIntegration(t *testing.T) {
	// 1. 加载配置
	cfg, err := config.LoadConfig("../../configs/config.yaml")
	if err != nil {
		t.Fatalf("配置加载失败: %v", err)
	}
	
	// 2. 初始化真实依赖
	ctx := context.Background()
	gormDB, _ := db.InitDB(ctx, &cfg.MySQL)
	redis.InitRedis(ctx, &cfg.Redis)
	dockerClient, _ := docker.NewDockerClient(&cfg.Docker)
	
	// 3. 创建服务
	svc := service.NewChallengeService(dockerClient, gormDB, cfg)
	
	// 4. 测试启动实例
	instance, err := svc.StartInstance(ctx, "test-user", "test-challenge-1")
	if err != nil {
		t.Fatalf("StartInstance失败: %v", err)
	}
	
	// 5. 验证结果
	if instance.Port < 20000 || instance.Port > 40000 {
		t.Errorf("端口范围错误: %d", instance.Port)
	}
	
	if !strings.HasPrefix(instance.Flag, "flag{test-user_") {
		t.Errorf("Flag格式错误: %s", instance.Flag)
	}
	
	// 6. 测试配额限制
	_, err = svc.StartInstance(ctx, "test-user", "test-challenge-1")
	if err == nil {
		t.Error("期望配额限制错误，但返回nil")
	}
	
	// 7. 清理
	svc.StopInstance(ctx, "test-user", "test-challenge-1")
}
```

**运行集成测试：**
```bash
# 需要先启动 MySQL + Redis + Docker
go test -v -tags=integration ./tests/integration
```

---

### 方案B：改进代码架构（推荐长期）

**问题根源：** Redis操作是静态函数，无法Mock

**解决方案：** 创建 `InstanceRepository` 接口

#### 步骤1：定义接口
```go
// internal/core/repository.go
type InstanceRepository interface {
    Store(ctx context.Context, instance *model.Instance) error
    GetByUser(ctx context.Context, userID string) ([]*model.Instance, error)
    Delete(ctx context.Context, instanceID, userID string) error
}
```

#### 步骤2：实现接口
```go
// internal/infra/redis/instance_repo.go
type RedisInstanceRepo struct {
    client *redis.Client
}

func (r *RedisInstanceRepo) Store(ctx context.Context, inst *model.Instance) error {
    // 原来 StoreInstance 的逻辑
}
```

#### 步骤3：注入到Service
```go
type ChallengeService struct {
    dockerClient core.ContainerEngine
    instanceRepo core.InstanceRepository  // 使用接口
    gormDB       *gorm.DB
    cfg          *config.Config
}
```

#### 步骤4：创建Mock
```go
// tests/mock/instance_repo_mock.go
type MockInstanceRepo struct {
    instances map[string]*model.Instance
}

func (m *MockInstanceRepo) Store(ctx, inst) error {
    m.instances[inst.ID] = inst
    return nil
}
```

这样就可以完全单元测试了！

---

## 🏃 快速测试方案（当前可用）

### 1. 测试部分逻辑（不依赖Redis）

```bash
go test -v ./internal/service -run TestGenerateFlag
go test -v ./internal/service -run TestGenerateID
```

### 2. 集成测试（需要环境）

```bash
# 启动依赖
docker-compose up -d mysql redis

# 运行完整流程测试
./test_core_features.sh
```

### 3. 性能测试

```bash
go test -bench=. ./internal/service
```

**预期输出：**
```
BenchmarkGenerateFlag-8      100000    10234 ns/op
BenchmarkGenerateID-8        500000     2345 ns/op
```

---

## 📊 测试覆盖情况

```bash
go test -cover ./internal/service
```

**目标：**
- 单元测试覆盖率：> 70%
- 集成测试覆盖核心流程：100%

---

## 🎓 总结

| 测试类型 | 需要的环境 | 覆盖范围 | 速度 |
|:---------|:-----------|:---------|:-----|
| **单元测试** | 无（Mock） | Flag生成、ID生成 | ⚡ 快 |
| **集成测试** | MySQL+Redis+Docker | 完整业务流程 | 🐢 慢 |
| **Bash脚本测试** | 全部+API服务器 | 端到端测试 | 🐌 很慢 |

**建议顺序：**
1. 先用 `test_core_features.sh` 快速验证完整功能
2. 补充单元测试覆盖边界情况
3. 长期：重构代码引入Repository接口，提高可测试性
