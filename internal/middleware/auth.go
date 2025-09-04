package middleware

import (
	"plane_war/internal/global"
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
			res.FailWithMsg("未携带 Authorization header", c)
			c.Abort()
			return
		}

		// 2. Bearer 格式校验
		parts := strings.Fields(authHeader)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			res.FailWithMsg("Authorization 格式错误，正确示例: Bearer <token>", c)
			c.Abort()
			return
		}
		tokenString := parts[1]

		// 3. 解析 JWT
		claims, err := jwts.ParseToken(tokenString)
		if err != nil {
			global.Log.Error(err.Error())
			res.FailWithMsg("Token 无效或已过期", c)
			c.Abort()
			return
		}
		// 4. 检查 Redis 中 token 是否存在
		userID, ok := redis_service.GetUserIDByAccessToken(tokenString)
		if !ok {
			res.FailWithMsg("Token 已失效，请重新登录", c)
			c.Abort()
			return
		}
		if userID != claims.UserID {
			res.FailWithMsg("Token 与用户信息不匹配", c)
			c.Abort()
			return
		}
		// 5. 注入用户信息到 Context
		c.Set("claims", claims)

		c.Next()
	}
}
