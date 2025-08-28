package model

import "github.com/gorilla/websocket"

type Player struct {
	ID   string          `json:"id"`   //唯一玩家id
	Name string          `json:"name"` //玩家昵称
	Conn *websocket.Conn `json:"-"`    //websocket 连接（也可以只存client）
}
