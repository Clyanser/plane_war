package api

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"plane_war/internal/global"
	"plane_war/internal/model"
	"plane_war/internal/model/res"
	"plane_war/internal/service/redis_service"
	"plane_war/internal/utils/jwts"
	"plane_war/internal/utils/pwd"
)

func Register(c *gin.Context) {
	var req struct {
		Username string `json:"username" binding:"required" msg:"请输入用户名"`
		Password string `json:"password" binding:"required" msg:"请输入密码"`
		Nickname string `json:"nickname"`
	}
	//绑定参数
	if err := c.ShouldBind(&req); err != nil {
		global.Log.Error(err.Error())
		res.FailWithMsg("参数错误", c)
		return
	}
	//检查用户是否已经存在
	var user model.User
	err := global.DB.Where("username = ?", req.Username).First(&user).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		res.FailWithMsg("用户已经存在", c)
	}
	//加密密码
	hashPassword, err := pwd.HashPassword(req.Password)
	if err != nil {
		global.Log.Error(err.Error())
		res.FailWithMsg("密码加密失败", c)
		return
	}
	//保存用户到数据库
	newUser := model.User{
		Username: req.Username,
		Password: string(hashPassword),
		Nickname: req.Nickname,
	}
	if err := global.DB.Create(&newUser).Error; err != nil {
		res.FailWithMsg("用户注册失败", c)
		return
	}
	res.OkWithMsg("注册成功", c)
}
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
	err := global.DB.Take(&user, "username=?", req.Username).Error
	if err != nil {
		global.Log.Warn("用户名不存在")
		return
	}
	//密码对比
	if !pwd.ComparePasswords(user.Password, req.Password) {
		res.FailWithMsg("密码错误", c)
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
		"access_token":  "Bearer " + accessToken,
		"refresh_token": "Bearer " + refreshToken,
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
