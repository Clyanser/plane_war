package model

import (
	"github.com/gorilla/websocket"
)

// Player 玩家信息
type Player struct {
	ID       string          `json:"id"`
	UserID   uint            `json:"user_id"`
	Name     string          `json:"name"`
	X        int             `json:"x"`        // 玩家位置 X
	Y        int             `json:"y"`        // 玩家位置 Y
	HP       int             `json:"hp"`       //玩家血量
	Position string          `json:"position"` //top or bottom
	Conn     *websocket.Conn `json:"-"`
	Ready    bool            `json:"ready"`
}
