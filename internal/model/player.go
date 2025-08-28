package model

import "github.com/gorilla/websocket"

// Player 玩家信息
type Player struct {
	ID   string          `json:"id"`
	Name string          `json:"name"`
	X    int             `json:"x"` // 玩家位置 X
	Y    int             `json:"y"` // 玩家位置 Y
	Conn *websocket.Conn `json:"-"`
}
