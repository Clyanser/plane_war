package main

import (
	"github.com/gin-gonic/gin"
	"log"
	"plane_war/internal/api"
)

func main() {
	r := gin.Default()

	//websocket 路由
	r.GET("/ws", api.WsHandler)

	log.Println("Listening on localhost:8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatal("server run failed:", err)
	}
}
