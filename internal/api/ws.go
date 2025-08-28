package api

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"plane_war/internal/model"
	"plane_war/internal/ws"
)

// Upgrader webSocket升级
var Upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true //允许跨域
	},
}

func WsHandler(c *gin.Context) {
	conn, err := Upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println("upgrader error:", err)
		return
	}
	//创建player
	playerID := uuid.New().String()
	player := model.Player{
		ID:   playerID,
		Name: "玩家—" + playerID[:4],
		Conn: conn,
		HP:   100,
	}
	//创建client 并注册到hub
	client := ws.NewClientWithPlayer(&player)
	go client.ReadPump()
	go client.WritePump()

	ws.HubInstance.Register <- client
}
