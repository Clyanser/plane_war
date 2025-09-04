package model

import (
	"plane_war/internal/model/ctype"
	"time"
)

type PublicRoom struct {
	ID       string           `json:"id"`
	Code     string           `json:"code"`     // 房间码
	OwnerID  uint             `json:"owner_id"` // 房间创建者用户ID
	Players  []*Player        `json:"players"`  // 房间内玩家
	Status   ctype.RoomStatus `json:"status"`   // 状态：等待中/游戏中
	Created  time.Time        `json:"created"`  // 创建时间
	Capacity int              `json:"capacity"` // 房间最大玩家数
}
