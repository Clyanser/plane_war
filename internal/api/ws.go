package api

import (
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
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
	client := ws.NewClient(conn)
	go client.ReadPump()
	go client.WritePump()

	ws.HubInstance.Register <- client
}
