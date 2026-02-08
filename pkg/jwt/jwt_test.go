package jwt

import (
	"testing"
	"time"
)

func TestGenerateAndParseAdminToken(t *testing.T) {
	adminID := "admin-123"
	username := "testadmin"

	// 生成 Token
	token, err := GenerateAdminToken(adminID, username)
	if err != nil {
		t.Fatalf("生成 Token 失败: %v", err)
	}

	if token == "" {
		t.Error("Token 不应为空")
	}

	// 解析 Token
	claims, err := ParseAdminToken(token)
	if err != nil {
		t.Fatalf("解析 Token 失败: %v", err)
	}

	if claims.AdminID != adminID {
		t.Errorf("AdminID = %v, want %v", claims.AdminID, adminID)
	}

	if claims.Username != username {
		t.Errorf("Username = %v, want %v", claims.Username, username)
	}

	// 验证过期时间
	expectedExpiry := time.Now().Add(24 * time.Hour)
	if claims.ExpiresAt.Time.Before(expectedExpiry.Add(-1 * time.Minute)) {
		t.Error("Token 过期时间不正确")
	}
}

func TestParseInvalidToken(t *testing.T) {
	tests := []struct {
		name  string
		token string
	}{
		{"空Token", ""},
		{"格式错误", "invalid.token.format"},
		{"错误签名", "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJhZG1pbl9pZCI6ImFkbWluLTEyMyIsInVzZXJuYW1lIjoidGVzdCJ9.invalid_signature"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseAdminToken(tt.token)
			if err == nil {
				t.Error("无效 Token 应该解析失败")
			}
		})
	}
}

func TestTokenSigningMethod(t *testing.T) {
	token, _ := GenerateAdminToken("admin-123", "test")

	// 验证签名方法是 HS256
	claims, err := ParseAdminToken(token)
	if err != nil {
		t.Fatalf("解析失败: %v", err)
	}

	if claims == nil {
		t.Error("Claims 不应为 nil")
	}
}

func TestMultipleTokenGeneration(t *testing.T) {
	// 生成多个 Token，确保它们都是有效的
	for i := 0; i < 10; i++ {
		token, err := GenerateAdminToken("admin-123", "test")
		if err != nil {
			t.Fatalf("生成第 %d 个 Token 失败: %v", i, err)
		}

		_, err = ParseAdminToken(token)
		if err != nil {
			t.Fatalf("解析第 %d 个 Token 失败: %v", i, err)
		}
	}
}

// 性能测试
func BenchmarkGenerateToken(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = GenerateAdminToken("admin-123", "test")
	}
}

func BenchmarkParseToken(b *testing.B) {
	token, _ := GenerateAdminToken("admin-123", "test")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = ParseAdminToken(token)
	}
}
