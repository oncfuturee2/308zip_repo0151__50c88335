package middleware

import (
	"net/http"
	"strings"

	"distribution-commission/internal/pkg/jwt"

	"github.com/gin-gonic/gin"
)

type ContextKey string

const (
	UserIDKey   ContextKey = "user_id"
	UsernameKey ContextKey = "username"
	RoleKey     ContextKey = "role"
)

func JWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "未登录或登录已过期",
			})
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if !(len(parts) == 2 && parts[0] == "Bearer") {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "认证格式错误",
			})
			c.Abort()
			return
		}

		claims, err := jwt.ParseToken(parts[1])
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "token无效或已过期",
			})
			c.Abort()
			return
		}

		c.Set(string(UserIDKey), claims.UserID)
		c.Set(string(UsernameKey), claims.Username)
		c.Set(string(RoleKey), string(claims.Role))

		c.Next()
	}
}

func RequireRoles(allowedRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get(string(RoleKey))
		if !exists {
			c.JSON(http.StatusForbidden, gin.H{
				"code":    403,
				"message": "无权限访问",
			})
			c.Abort()
			return
		}

		userRole := role.(string)
		for _, allowedRole := range allowedRoles {
			if userRole == allowedRole {
				c.Next()
				return
			}
		}

		c.JSON(http.StatusForbidden, gin.H{
			"code":    403,
			"message": "无权限访问",
		})
		c.Abort()
	}
}

func GetUserID(c *gin.Context) uint {
	if userID, exists := c.Get(string(UserIDKey)); exists {
		return userID.(uint)
	}
	return 0
}

func GetRole(c *gin.Context) string {
	if role, exists := c.Get(string(RoleKey)); exists {
		return role.(string)
	}
	return ""
}
