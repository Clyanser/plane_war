package router

import (
	"plane_war/internal/api"
	"plane_war/internal/middleware"
)

func (r RouterGroup) LobbyRouter() {
	r.POST("/lobby/create_room", middleware.AuthMiddleware(), api.CreatePublicRoom)
	r.GET("/lobby/room_list", middleware.AuthMiddleware(), api.GetPublicRooms)
	r.POST("/lobby/join_room", middleware.AuthMiddleware(), api.JoinPublicRoom)
	r.GET("/lobby/start_game", middleware.AuthMiddleware(), api.StartGame)
	r.GET("/lobby/dismiss_room", middleware.AuthMiddleware(), api.DismissPublicRoom)
}
