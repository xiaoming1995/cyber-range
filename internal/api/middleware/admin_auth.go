package middleware

import (
	"cyber-range/pkg/jwt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// AdminAuth 管理员认证中间件
func AdminAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从 Header 获取 Token
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code": 401,
				"msg":  "未提供认证token",
			})
			return
		}

		// Bearer Token 格式
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code": 401,
				"msg":  "token格式错误",
			})
			return
		}

		tokenString := parts[1]

		// 解析 Token
		claims, err := jwt.ParseAdminToken(tokenString)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code": 401,
				"msg":  "token无效或已过期",
			})
			return
		}

		// 将管理员信息存入上下文
		c.Set("admin_id", claims.AdminID)
		c.Set("admin_username", claims.Username)

		c.Next()
	}
}

// GetAdminID 从Context获取管理员ID
func GetAdminID(c *gin.Context) (string, bool) {
	adminIDVal, exists := c.Get("admin_id")
	if !exists {
		return "", false
	}
	adminID, ok := adminIDVal.(string)
	if !ok {
		return "", false
	}
	return adminID, true
}
