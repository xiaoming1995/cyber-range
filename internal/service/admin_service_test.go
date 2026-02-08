package service

import (
	"context"
	"cyber-range/internal/model"
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupAdminTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("创建测试数据库失败: %v", err)
	}
	db.AutoMigrate(&model.Admin{})
	return db
}

func TestAdminService_CreateAdmin(t *testing.T) {
	db := setupAdminTestDB(t)
	svc := NewAdminService(db)
	ctx := context.Background()

	tests := []struct {
		name      string
		username  string
		email     string
		password  string
		adminName string
		wantErr   bool
	}{
		{
			name:      "成功创建管理员",
			username:  "admin1",
			email:     "admin1@test.com",
			password:  "Test@1234",
			adminName: "测试管理员1",
			wantErr:   false,
		},
		{
			name:      "重复用户名",
			username:  "admin1", // 重复
			email:     "admin2@test.com",
			password:  "Test@1234",
			adminName: "测试管理员2",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			admin, err := svc.CreateAdmin(ctx, tt.username, tt.email, tt.password, tt.adminName)

			if (err != nil) != tt.wantErr {
				t.Errorf("CreateAdmin() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if admin.Username != tt.username {
					t.Errorf("Username = %v, want %v", admin.Username, tt.username)
				}
				if admin.Email != tt.email {
					t.Errorf("Email = %v, want %v", admin.Email, tt.email)
				}
				if !admin.IsActive {
					t.Error("新创建的管理员应该是激活状态")
				}
			}
		})
	}
}

func TestAdminService_Login_Success(t *testing.T) {
	db := setupAdminTestDB(t)
	svc := NewAdminService(db)
	ctx := context.Background()

	// 创建测试管理员
	password := "Test@1234"
	svc.CreateAdmin(ctx, "admin1", "admin@test.com", password, "测试管理员")

	// 测试登录
	token, admin, err := svc.Login(ctx, "admin1", password)
	if err != nil {
		t.Fatalf("登录失败: %v", err)
	}

	if token == "" {
		t.Error("Token 不应为空")
	}

	if admin.Username != "admin1" {
		t.Errorf("Username = %v, want admin1", admin.Username)
	}

	// 验证最后登录时间被更新
	if admin.LastLoginAt == nil {
		t.Error("LastLoginAt 应该被更新")
	}
}

func TestAdminService_Login_WrongPassword(t *testing.T) {
	db := setupAdminTestDB(t)
	svc := NewAdminService(db)
	ctx := context.Background()

	svc.CreateAdmin(ctx, "admin1", "admin@test.com", "Test@1234", "测试")

	_, _, err := svc.Login(ctx, "admin1", "WrongPassword")
	if err == nil {
		t.Error("错误密码应该登录失败")
	}
}

func TestAdminService_Login_NonexistentUser(t *testing.T) {
	db := setupAdminTestDB(t)
	svc := NewAdminService(db)
	ctx := context.Background()

	_, _, err := svc.Login(ctx, "nonexistent", "password")
	if err == nil {
		t.Error("不存在的用户应该登录失败")
	}
}

func TestAdminService_Login_InactiveUser(t *testing.T) {
	db := setupAdminTestDB(t)
	svc := NewAdminService(db)
	ctx := context.Background()

	// 创建管理员
	admin, _ := svc.CreateAdmin(ctx, "admin1", "admin@test.com", "Test@1234", "测试")

	// 禁用管理员
	db.Model(&model.Admin{}).Where("id = ?", admin.ID).Update("is_active", false)

	// 尝试登录
	_, _, err := svc.Login(ctx, "admin1", "Test@1234")
	if err == nil {
		t.Error("禁用的管理员不应该能登录")
	}
}

func TestAdminService_GetAdminByID(t *testing.T) {
	db := setupAdminTestDB(t)
	svc := NewAdminService(db)
	ctx := context.Background()

	// 创建管理员
	created, _ := svc.CreateAdmin(ctx, "admin1", "admin@test.com", "Test@1234", "测试")

	// 获取管理员
	admin, err := svc.GetAdminByID(ctx, created.ID)
	if err != nil {
		t.Fatalf("GetAdminByID 失败: %v", err)
	}

	if admin.ID != created.ID {
		t.Errorf("ID = %v, want %v", admin.ID, created.ID)
	}

	// 获取不存在的管理员
	_, err = svc.GetAdminByID(ctx, "nonexistent-id")
	if err == nil {
		t.Error("获取不存在的管理员应该返回错误")
	}
}

func TestAdminService_ConcurrentLogin(t *testing.T) {
	// 每个并发测试使用独立的数据库实例
	concurrency := 10
	done := make(chan error, concurrency)

	for i := 0; i < concurrency; i++ {
		go func(idx int) {
			// 每个 goroutine 使用自己的数据库实例
			db := setupAdminTestDB(t)
			svc := NewAdminService(db)
			ctx := context.Background()

			password := "Test@1234"
			_, err := svc.CreateAdmin(ctx, "admin1", "admin@test.com", password, "测试")
			if err != nil {
				done <- err
				return
			}

			_, _, err = svc.Login(ctx, "admin1", password)
			done <- err
		}(i)
	}

	// 收集结果
	errorCount := 0
	for i := 0; i < concurrency; i++ {
		if err := <-done; err != nil {
			errorCount++
			t.Logf("并发测试 #%d 出错: %v", i, err)
		}
	}

	if errorCount > 0 {
		t.Error("并发登录出现错误")
	}
}

// 性能测试
func BenchmarkAdminLogin(b *testing.B) {
	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	db.AutoMigrate(&model.Admin{})

	svc := NewAdminService(db)
	ctx := context.Background()

	password := "Test@1234"
	svc.CreateAdmin(ctx, "admin1", "admin@test.com", password, "测试")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = svc.Login(ctx, "admin1", password)
	}
}
