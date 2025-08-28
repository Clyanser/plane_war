package game

import (
	"encoding/json"
	"github.com/gorilla/websocket"
	"plane_war/internal/model"
	"time"
)

func StartRoomLoop(room *model.Room) {
	ticker := time.NewTicker(500 * time.Millisecond)
	quit := make(chan bool)

	go func() {
		for {
			select {
			case <-ticker.C:
				room.Lock.Lock()
				// 更新子弹位置
				newBullets := []*model.Bullet{}
				for _, bullet := range room.Bullets {
					bullet.Y -= bullet.Speed
					if bullet.Y >= 0 {
						newBullets = append(newBullets, bullet)
					}
				}
				room.Bullets = newBullets
				// TODO: 碰撞检测

				//广播房间 状态
				broadcastRoomState(room)
				room.Lock.Unlock()
			case <-quit:
				ticker.Stop()
				return
			}
		}
	}()
}

func broadcastRoomState(room *model.Room) {
	state := map[string]interface{}{
		"type":    "game_state",
		"players": room.Players,
		"bullets": room.Bullets,
	}
	data, _ := json.Marshal(state)

	for _, player := range room.Players {
		player.Conn.WriteMessage(websocket.TextMessage, data)
	}
}
