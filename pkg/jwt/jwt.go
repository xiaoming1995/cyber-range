package jwt

import (
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var (
	jwtSecret []byte
)

// 初始化 JWT Secret
func init() {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		// 开发环境使用默认值，生产环境必须设置
		env := os.Getenv("APP_ENV")
		if env == "production" {
			log.Fatal("JWT_SECRET environment variable is required in production")
		}
		log.Println("Warning: Using default JWT_SECRET for development. Set JWT_SECRET env var for production.")
		secret = "cyber-range-dev-secret-change-in-production"
	}
	jwtSecret = []byte(secret)
}

// AdminClaims JWT 自定义声明
type AdminClaims struct {
	AdminID  string `json:"admin_id"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}

// GenerateAdminToken 生成管理员 JWT Token
func GenerateAdminToken(adminID, username string) (string, error) {
	claims := AdminClaims{
		AdminID:  adminID,
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

// ParseAdminToken 解析并验证管理员 JWT Token
func ParseAdminToken(tokenString string) (*AdminClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &AdminClaims{}, func(token *jwt.Token) (interface{}, error) {
		// 验证签名算法，防止算法混淆攻击
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return jwtSecret, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*AdminClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}
