package middleware

import (
	"plane_war/internal/model/res"
	"plane_war/internal/service/redis_service"
	"plane_war/internal/utils/jwts"
	"strings"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware Gin 鉴权中间件
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. 获取 Authorization Header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			res.FailWithMsg("未携带token", c)
			c.Abort()
			return
		}

		// 2. 支持 Bearer token
		parts := strings.Fields(authHeader)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			res.FailWithMsg("token 错误", c)
			c.Abort()
			return
		}
		tokenString := parts[1]

		// 3. 解析 JWT
		claims, err := jwts.ParseToken(tokenString)
		if err != nil {
			res.FailWithMsg("无效的 Token", c)
			c.Abort()
			return
		}

		// 4. 检查 Redis 中 token 是否存在
		// 判断是否在 Redis 中有效
		if userID, ok := redis_service.GetUserIDByAccessToken(tokenString); !ok || userID != claims.UserID {
			res.FailWithMsg("Token 已失效", c)
			c.Abort()
			return
		}

		// 5. 将用户信息注入 Context
		c.Set("user_id", claims.UserID)
		c.Set("nickname", claims.Nickname)

		c.Next()
	}
}
