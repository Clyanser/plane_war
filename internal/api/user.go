package api

import (
	"github.com/gin-gonic/gin"
	"plane_war/internal/global"
	"plane_war/internal/model"
	"plane_war/internal/model/res"
	"plane_war/internal/service/redis_service"
	"plane_war/internal/utils/jwts"
)

func Login(c *gin.Context) {
	var req struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		res.FailWithMsg("参数错误", c)
		return
	}

	var user model.User
	err := global.DB.Take(&user, "username=? and password = ?", req.Username, req.Password).Error
	if err != nil {
		global.Log.Warn("用户名不存在")
		return
	}
	//生成 ACCESS Token
	accessToken, err := jwts.GenAccessToken(jwts.JwtPayload{
		UserID:   user.ID,
		Nickname: user.Nickname,
	})
	if err != nil {
		res.FailWithMsg("生成AccessToken失败", c)
		global.Log.Warn(err)
		return
	}

	//生成 RefreshToken
	refreshToken, err := jwts.GenRefreshToken(jwts.JwtPayload{
		UserID:   user.ID,
		Nickname: user.Nickname,
	})
	if err != nil {
		res.FailWithMsg("生成Refresh Token 失败", c)
		return
	}

	//存储到 accessToken
	err = redis_service.SaveAccessToken(accessToken, user.ID)
	if err != nil {
		res.FailWithMsg("存储 accessToken 失败", c)
		return
	}
	//存储 RefreshToken
	err = redis_service.SaveRefreshToken(refreshToken, user.ID)
	if err != nil {
		res.FailWithMsg("存储 refresh token 失败", c)
		return
	}
	authData := map[string]interface{}{
		"access_token":  "Bearer" + accessToken,
		"refresh_token": "Bearer" + refreshToken,
	}
	res.OkWithData(authData, c)
}

func RefreshToken(c *gin.Context) {
	var req struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		res.FailWithMsg("参数错误", c)
		return
	}
	//获取用户ID 根据RefreshToken
	userID, ok := redis_service.GetUserIDByRefreshToken(req.RefreshToken)
	if !ok {
		res.FailWithMsg("Refresh Token 无效", c)
		return
	}

	accessToken, err := jwts.GenAccessToken(jwts.JwtPayload{
		UserID: userID,
	})
	if err != nil {
		res.FailWithMsg("生成新的Access Token 失败", c)
		return
	}
	//存储新的Access Token 到redis
	err = redis_service.SaveAccessToken(accessToken, userID)
	if err != nil {
		res.FailWithMsg("存储Access Token 失败", c)
		return
	}
	authData := map[string]interface{}{
		"access_token": "Bearer" + accessToken,
	}
	res.OkWithData(authData, c)
}
