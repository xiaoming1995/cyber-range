package service

import (
	"cyber-range/internal/infra/db"
	"cyber-range/internal/infra/docker"
	"cyber-range/internal/model"
	"cyber-range/pkg/config"
	"strings"
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// setupTestDB 创建内存SQLite数据库用于测试
func setupTestDB(t *testing.T) *gorm.DB {
	testDB, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("创建测试数据库失败: %v", err)
	}

	// 自动迁移
	testDB.AutoMigrate(
		&model.DockerHost{},
		&model.Challenge{},
		&model.Instance{},
		&model.User{},
		&model.Submission{},
	)

	// 插入测试 Docker 主机
	testDB.Create(&model.DockerHost{
		ID:           "test-docker-host",
		Name:         "测试 Docker 主机",
		Host:         "",
		TLSVerify:    false,
		CertPath:     "",
		PortRangeMin: 20000,
		PortRangeMax: 40000,
		MemoryLimit:  134217728,
		CPULimit:     0.5,
		Enabled:      true,
		IsDefault:    true,
		Description:  "单元测试用 Docker 主机",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	})

	// 插入测试题目
	testDB.Create(&model.Challenge{
		ID:           "test-challenge-1",
		Title:        "测试题目",
		Description:  "单元测试用题目",
		Category:     "Web",
		Difficulty:   "Easy",
		Image:        "nginx:alpine",
		DockerHostID: "test-docker-host",
		Flag:         "flag{static}",
		Points:       100,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	})

	// 插入测试用户
	testDB.Create(&model.User{
		ID:           "test-user-1",
		Username:     "testuser",
		Email:        "test@test.com",
		PasswordHash: "hash123",
		Role:         "user",
		TotalPoints:  0,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	})

	return testDB
}

// setupTestConfig 创建测试配置
func setupTestConfig() *config.Config {
	return &config.Config{
		Instance: config.InstanceConfig{
			MaxPerUser: 1,
			TTLHours:   1,
		},
	}
}

// setupTestService 创建测试用的 ChallengeService
func setupTestService(t *testing.T) (*ChallengeService, *gorm.DB) {
	testDB := setupTestDB(t)
	testCfg := setupTestConfig()

	// 创建 Docker 主机管理器和 Repository
	dockerManager := docker.NewDockerHostManager()
	repository := db.NewRepository(testDB)

	svc := NewChallengeService(dockerManager, repository, testDB, testCfg)
	return svc, testDB
}

// TestGenerateFlag 测试Flag生成逻辑
func TestGenerateFlag(t *testing.T) {
	tests := []struct {
		name       string
		userID     string
		wantPrefix string
	}{
		{
			name:       "正常用户ID",
			userID:     "user_123",
			wantPrefix: "flag{user_123_",
		},
		{
			name:       "空用户ID",
			userID:     "",
			wantPrefix: "flag{_",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, _ := setupTestService(t)
			flag := svc.generateFlag(tt.userID)

			if !strings.HasPrefix(flag, tt.wantPrefix) {
				t.Errorf("generateFlag() = %v, 期望前缀 %v", flag, tt.wantPrefix)
			}

			if !strings.HasSuffix(flag, "}") {
				t.Errorf("generateFlag() = %v, 应该以 } 结尾", flag)
			}
		})
	}
}

// TestGenerateID 测试ID生成
func TestGenerateID(t *testing.T) {
	id1 := generateID()
	id2 := generateID()

	if id1 == id2 {
		t.Error("generateID() 生成了重复的ID")
	}

	if len(id1) < 10 {
		t.Errorf("generateID() = %v, ID太短", id1)
	}
}

// TestGetChallenge 测试获取题目
func TestGetChallenge(t *testing.T) {
	svc, _ := setupTestService(t)

	// 测试获取存在的题目
	challenge, err := svc.GetChallenge(nil, "test-challenge-1")
	if err != nil {
		t.Errorf("GetChallenge() error = %v", err)
	}

	if challenge.ID != "test-challenge-1" {
		t.Errorf("GetChallenge() ID = %v, want %v", challenge.ID, "test-challenge-1")
	}

	// 测试获取不存在的题目
	_, err = svc.GetChallenge(nil, "non-existent")
	if err == nil {
		t.Error("GetChallenge() 应该返回错误，但没有")
	}
}

// TestListChallenges 测试获取题目列表
func TestListChallenges(t *testing.T) {
	svc, testDB := setupTestService(t)

	// 更新题目为已发布状态
	testDB.Model(&model.Challenge{}).Where("id = ?", "test-challenge-1").Update("status", "published")

	challenges, err := svc.ListChallenges(nil)
	if err != nil {
		t.Errorf("ListChallenges() error = %v", err)
	}

	if len(challenges) != 1 {
		t.Errorf("ListChallenges() 返回 %d 个题目, 期望 1 个", len(challenges))
	}
}

// TestStartInstance_MockSuccess 测试容器启动的核心逻辑
// 注意：此测试需要Mock Redis，或者跳过Redis部分
func TestStartInstance_MockSuccess(t *testing.T) {
	t.Skip("此测试需要Redis支持，跳过。请使用集成测试或运行 test_core_features.sh")
}

// BenchmarkGenerateFlag 性能测试：Flag生成速度
func BenchmarkGenerateFlag(b *testing.B) {
	testDB, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	testCfg := setupTestConfig()
	dockerManager := docker.NewDockerHostManager()
	repository := db.NewRepository(testDB)

	svc := NewChallengeService(dockerManager, repository, testDB, testCfg)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = svc.generateFlag("user_123")
	}
}
