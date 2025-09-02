package redis_service

import (
	"plane_war/internal/global"
	"strconv"
	"time"
)

func SaveAccessToken(tokenString string, userID uint) error {
	key := "accessToken:" + tokenString
	return global.Redis.Set(key, userID, time.Second*time.Duration(global.Config.Auth.AccessExpire)).Err()

}
func SaveRefreshToken(tokenString string, userID uint) error {
	key := "refreshToken:" + tokenString
	return global.Redis.Set(key, userID, time.Second*time.Duration(global.Config.Auth.RefreshExpire)).Err()
}

func GetUserIDByAccessToken(tokenString string) (uint, bool) {
	key := "accessToken:" + tokenString
	val, err := global.Redis.Get(key).Result()
	if err != nil {
		return 0, false
	}
	userID, err := strconv.Atoi(val)
	if err != nil {
		return 0, false
	}

	return uint(userID), true
}

// GetUserIDByRefreshToken 根据 Refresh Token 获取用户 ID
func GetUserIDByRefreshToken(token string) (uint, bool) {
	key := "refreshToken:" + token
	val, err := global.Redis.Get(key).Result()
	if err != nil {
		return 0, false
	}
	// 将 Redis 存储的字符串 val 转换为 int，再转换为 uint
	userID, err := strconv.Atoi(val)
	if err != nil {
		return 0, false // 转换失败
	}
	return uint(userID), true
}

// RemoveAccessToken 删除 Access Token
func RemoveAccessToken(token string) error {
	key := "accessToken:" + token
	return global.Redis.Del(key).Err()
}

// RemoveRefreshToken 删除 Refresh Token
func RemoveRefreshToken(token string) error {
	key := "refreshToken:" + token
	return global.Redis.Del(key).Err()
}
