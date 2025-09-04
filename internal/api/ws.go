package api

import (
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"plane_war/internal/model"
	"plane_war/internal/utils/jwts"
	"plane_war/internal/ws"
	"strconv"
)

// Upgrader webSocket升级
var Upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true //允许跨域
	},
}

func WsHandler(c *gin.Context) {
	// 从JWT中解析玩家信息
	_cliams, _ := c.Get("claims")
	claims := _cliams.(*jwts.CustomClaims)

	conn, err := Upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println("upgrader error:", err)
		return
	}
	//创建player
	player := model.Player{
		ID:   strconv.Itoa(int(claims.UserID)),
		Name: claims.Nickname,
		Conn: conn,
		HP:   100,
	}
	//创建client 并注册到hub
	client := ws.NewClientWithPlayer(&player)
	go client.ReadPump()
	go client.WritePump()

	ws.HubInstance.Register <- client
}
