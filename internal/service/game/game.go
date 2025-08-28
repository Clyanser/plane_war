package game

import (
	"encoding/json"
	"github.com/gorilla/websocket"
	"log"
	"plane_war/internal/model"
	"time"
)

func StartRoomLoop(room *model.Room) {
	ticker := time.NewTicker(50 * time.Millisecond)
	quit := make(chan bool)

	go func() {
		for {
			select {
			case <-ticker.C:
				room.Lock.Lock()
				// 更新子弹位置,并进行碰撞检测
				newBullets := []*model.Bullet{}
				for _, bullet := range room.Bullets {
					bullet.Y -= bullet.Speed
					hit := false
					for _, p := range room.Players {
						if p.ID != bullet.Owner && checkCollision(bullet, p) {
							p.HP -= bullet.Damage
							hit = true
							break
						}
					}
					if !hit && bullet.Y >= 0 {
						newBullets = append(newBullets, bullet)
					}
				}
				room.Bullets = newBullets
				//检测玩家存活情况
				alivePlayers := []*model.Player{}
				var winner *model.Player
				for _, player := range room.Players {
					if player.HP > 0 {
						alivePlayers = append(alivePlayers, player)
					}

				}
				if len(alivePlayers) <= 1 {
					if len(alivePlayers) == 1 {
						winner = alivePlayers[0]
					}
					broadcastGameOver(room, winner)
					room.Lock.Unlock()
					ticker.Stop()
					return
				}
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

func checkCollision(b *model.Bullet, p *model.Player) bool {
	// 简单矩形碰撞
	if b.X >= p.X && b.X <= p.X+50 && b.Y >= p.Y && b.Y <= p.Y+50 {
		return true
	}
	return false
}

// 广播游戏结束状态
func broadcastGameOver(room *model.Room, winner *model.Player) {
	state := map[string]interface{}{
		"type":   "game_over",
		"winner": winner,
	}
	data, _ := json.Marshal(state)
	for _, player := range room.Players {
		player.Conn.WriteMessage(websocket.TextMessage, data)
	}
	log.Printf("房间 %s 游戏结束，胜利者: %v", room.ID, winner)
}
