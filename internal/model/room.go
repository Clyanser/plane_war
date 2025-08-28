package model

import "sync"

type Room struct {
	ID      string     `json:"id"`      //房间id
	Players []*Player  `json:"players"` //房间内玩家
	Lock    sync.Mutex //房间锁，防止并发操作
}
