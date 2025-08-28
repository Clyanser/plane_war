package model

import (
	"sync"
	"time"
)

// Room 房间信息
type Room struct {
	ID      string       `json:"id"` //房间id
	Players []*Player    //房间内玩家
	Bullets []*Bullet    //房间内的子弹
	Lock    sync.Mutex   //房间锁，防止并发操作
	Ticker  *time.Ticker //用于房间循环
	Quit    chan bool    //用于房间循环
}

// Bullet 子弹信息
type Bullet struct {
	ID    string `json:"id"`
	X     int    `json:"x"`
	Y     int    `json:"y"`
	Owner string `json:"owner"` //玩家ID
	Speed int    `json:"speed"`
}
