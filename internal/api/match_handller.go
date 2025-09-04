package api

import (
	"github.com/gin-gonic/gin"
	"plane_war/internal/model"
	"plane_war/internal/utils/jwts"
)

func MatchHandler(c *gin.Context) {
	//获取用户信息
	_cliams,_ :=c.Get("cliams")
	cliam := _cliams.(*jwts.CustomClaims)

	player := &model.Player{
		UserID: cliam.UserID,
		Name: cliam.Nickname,
	}
	room :=
}
