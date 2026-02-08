# Golang 单元测试指南

## 📚 已创建的测试文件

### 1. Mock对象
**文件：** `tests/mock/docker_mock.go`

实现了Mock Docker客户端，用于在无需真实Docker环境下测试Service层逻辑。

```go
mockDocker := mock.NewMockDockerClient()
mockDocker.ShouldFailStart = true  // 模拟启动失败
```

### 2. Service层单元测试
**文件：** `internal/service/challenge_service_test.go`

包含以下测试：
- ✅ `TestGenerateFlag` - 测试Flag生成逻辑
- ✅ `TestListChallenges` - 测试获取题目列表
- ✅ `TestGetChallenge` - 测试获取单个题目
- ✅ `TestGenerateID` - 测试ID生成

---

## 🚀 如何运行测试

### 运行所有测试
```bash
go test ./...
```

### 运行指定包的测试
```bash
go test ./internal/service
```

### 详细输出
```bash
go test -v ./internal/service
```

### 运行特定测试
```bash
go test -v ./internal/service -run TestGenerateFlag
```

### 查看覆盖率
```bash
go test -cover ./internal/service
go test -coverprofile=coverage.out ./internal/service
go tool cover -html=coverage.out  # 生成HTML报告
```

---

## 📝 测试规范

### 1. 文件命名
```
service.go        # 源代码
service_test.go   # 测试文件（必须_test.go结尾）
```

### 2. 测试函数
```go
func TestXxx(t *testing.T) {
    // 测试函数必须以Test开头
    // 参数必须是 *testing.T
}
```

### 3. 表驱动测试（推荐）
```go
func TestSomething(t *testing.T) {
    tests := []struct {
        name string
        input string
        want string
    }{
        {"测试场景1", "input1", "output1"},
        {"测试场景2", "input2", "output2"},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got := YourFunction(tt.input)
            if got != tt.want {
                t.Errorf("got %v, want %v", got, tt.want)
            }
        })
    }
}
```

---

## 🎯 单元测试 vs 集成测试

### 单元测试（不需要外部依赖）
```go
// 使用Mock对象
mockDocker := mock.NewMockDockerClient()
mockDB := setupTestDB(t)  // 内存SQLite

svc := NewChallengeService(mockDocker, mockDB, cfg)
```

### 集成测试（需要真实环境）
```go
func TestIntegration(t *testing.T) {
    if testing.Short() {
        t.Skip("跳过集成测试")
    }
    
    // 使用真实MySQL + Redis + Docker
}
```

运行时排除集成测试：
```bash
go test -short ./...
```

---

## 🛠️ 常用断言模式

```go
// 1. 简单相等
if got != want {
    t.Errorf("got %v, want %v", got, want)
}

// 2. 错误检查
if err != nil {
    t.Fatalf("期望成功，但返回错误: %v", err)
}

// 3. 多条件检查
if got < 0 || got > 100 {
    t.Errorf("值超出范围: %d", got)
}

// 4. 空值检查
if result == nil {
    t.Error("结果不应为nil")
}
```

---

## 📊 测试示例输出

```
=== RUN   TestGenerateFlag
=== RUN   TestGenerateFlag/正常用户ID
=== RUN   TestGenerateFlag/空用户ID
--- PASS: TestGenerateFlag (0.00s)
    --- PASS: TestGenerateFlag/正常用户ID (0.00s)
    --- PASS: TestGenerateFlag/空用户ID (0.00s)
PASS
ok      cyber-range/internal/service    0.698s
```

---

## 🎓 最佳实践

1. **每个公开函数都应该有测试**
2. **使用表驱动测试覆盖多种场景**
3. **单元测试应该快速（<1秒）**
4. **Mock外部依赖（数据库、API、Docker）**
5. **测试名称应该描述清楚测试内容**
6. **使用 `t.Helper()` 标记辅助函数**

---

## 🔗 相关资源

- [Go Testing包文档](https://pkg.go.dev/testing)
- [表驱动测试最佳实践](https://github.com/golang/go/wiki/TableDrivenTests)
- [Testify断言库](https://github.com/stretchr/testify)（可选）
