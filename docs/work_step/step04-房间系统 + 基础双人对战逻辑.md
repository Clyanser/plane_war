### 🎯 阶段四目标
1. 房间内状态管理
    * 每个房间管理自己的玩家和游戏状态（玩家位置、子弹位置等）
2. 房间循环（Game Loop）
    * 定期更新游戏状态（比如每 50ms 或 60ms）
3. 玩家动作同步
    * 玩家发消息 → 房间更新状态 → 同步给房间内另一名玩家
4. 最基础的战斗逻辑
    * 玩家飞机可以移动
    * 发射子弹（不涉及碰撞检测，可以先同步子弹位置）


### 项目目录结构
game-server/
├── cmd/server/main.go
├── internal/
│   ├── api/ws.go
│   ├── model/
│   │   ├── player.go
│   │   └── room.go
│   ├── service/
│   │   ├── match/match.go
│   │   └── game/game.go
│   └── ws/hub.go
└── static/html/test.html

### internal/model/player.go
```go
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

```
### internal/model/room.go
```go
package model

import "sync"

// Bullet 子弹信息
type Bullet struct {
	ID    string `json:"id"`
	X     int    `json:"x"`
	Y     int    `json:"y"`
	Owner string `json:"owner"`
	Speed int    `json:"speed"`
}

// Room 房间信息
type Room struct {
	ID      string
	Players []*Player
	Bullets []*Bullet
	Lock    sync.Mutex
}

```
### internal/service/match/match.go
```go
package match

import (
	"airplane-war/internal/model"
	"github.com/google/uuid"
	"sync"
)

// MatchQueue 匹配队列
type MatchQueue struct {
	queue []*model.Player
	lock  sync.Mutex
}

var MatchQueueInstance = &MatchQueue{
	queue: make([]*model.Player, 0),
}

// AddPlayer 玩家加入匹配队列
func (mq *MatchQueue) AddPlayer(p *model.Player) *model.Room {
	mq.lock.Lock()
	defer mq.lock.Unlock()

	mq.queue = append(mq.queue, p)

	if len(mq.queue) >= 2 {
		p1 := mq.queue[0]
		p2 := mq.queue[1]
		mq.queue = mq.queue[2:]

		room := &model.Room{
			ID:      uuid.New().String(),
			Players: []*model.Player{p1, p2},
		}

		return room
	}
	return nil
}

```
### internal/service/game/game.go 房间循环
```go
package game

import (
	"airplane-war/internal/model"
	"encoding/json"
	"log"
	"time"
)

// 启动房间循环
func StartRoomLoop(room *model.Room) {
	ticker := time.NewTicker(50 * time.Millisecond)
	quit := make(chan bool)

	go func() {
		for {
			select {
			case <-ticker.C:
				room.Lock.Lock()
				// 更新子弹位置
				newBullets := []*model.Bullet{}
				for _, b := range room.Bullets {
					b.Y -= b.Speed
					if b.Y >= 0 {
						newBullets = append(newBullets, b)
					}
				}
				room.Bullets = newBullets

				// 广播房间状态
				broadcastRoomState(room)
				room.Lock.Unlock()
			case <-quit:
				ticker.Stop()
				return
			}
		}
	}()
}

// 广播房间状态给玩家
func broadcastRoomState(room *model.Room) {
	state := map[string]interface{}{
		"type":    "game_state",
		"players": room.Players,
		"bullets": room.Bullets,
	}
	data, _ := json.Marshal(state)
	for _, p := range room.Players {
		p.Conn.WriteMessage(1, data)
	}
}

```
##房间循环编写思路
**使用 Ticker + select 实现** 周期性更新 + 可停止的循环
**加锁** 保证房间状态在循环和玩家动作间不会竞争
**子弹和玩家状态更新** 是循环的核心
**广播房间状态** 是让前端实时显示的关键
这种方式让每个房间独立循环，可以扩展成多房间并行运行
每个房间循环就是游戏逻辑的“引擎心跳”

### internal/ws/hub.go
```go
package ws

import (
	"airplane-war/internal/model"
	"airplane-war/internal/service/game"
	"airplane-war/internal/service/match"
	"encoding/json"
	"github.com/google/uuid"
	"log"
	"sync"
)

type Client struct {
	Player *model.Player
	Send   chan []byte
}

func NewClientWithPlayer(p *model.Player) *Client {
	return &Client{
		Player: p,
		Send:   make(chan []byte, 256),
	}
}

// Hub 管理客户端
type Hub struct {
	Clients    map[*Client]bool
	Broadcast  chan []byte
	Register   chan *Client
	Unregister chan *Client
}

var HubInstance = NewHub()

func NewHub() *Hub {
	h := &Hub{
		Clients:    make(map[*Client]bool),
		Broadcast:  make(chan []byte),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
	}
	go h.Run()
	return h
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.Register:
			h.Clients[client] = true
			log.Printf("✅ 新玩家连接: %s (%s)", client.Player.ID, client.Player.Name)
		case client := <-h.Unregister:
			if _, ok := h.Clients[client]; ok {
				delete(h.Clients, client)
				close(client.Send)
				log.Printf("❌ 玩家断开: %s (%s)", client.Player.ID, client.Player.Name)
			}
		case message := <-h.Broadcast:
			for client := range h.Clients {
				select {
				case client.Send <- message:
				default:
					close(client.Send)
					delete(h.Clients, client)
				}
			}
		}
	}
}

// ---------------- 消息处理 ----------------
type Message struct {
	Action string `json:"action"`
	X      int    `json:"x,omitempty"`
	Y      int    `json:"y,omitempty"`
}

// 房间管理
var RoomMap = map[string]*model.Room{}
var RoomLock sync.Mutex{}

func (c *Client) ReadPump() {
	defer func() {
		HubInstance.Unregister <- c
		c.Player.Conn.Close()
	}()

	for {
		_, msg, err := c.Player.Conn.ReadMessage()
		if err != nil {
			log.Println("read error:", err)
			break
		}

		var m Message
		if err := json.Unmarshal(msg, &m); err != nil {
			log.Println("json parse error:", err)
			continue
		}

		switch m.Action {
		case "match":
			room := match.MatchQueueInstance.AddPlayer(c.Player)
			if room != nil {
				RoomLock.Lock()
				RoomMap[room.ID] = room
				RoomLock.Unlock()

				for _, p := range room.Players {
					resp := map[string]string{
						"type":    "match_success",
						"room_id": room.ID,
					}
					data, _ := json.Marshal(resp)
					p.Conn.WriteMessage(1, data)
				}
				log.Printf("🎯 匹配成功，房间ID: %s", room.ID)

				// 启动房间循环
				game.StartRoomLoop(room)
			}

		case "move":
			room := findPlayerRoom(c.Player.ID)
			if room != nil {
				room.Lock.Lock()
				c.Player.X = m.X
				c.Player.Y = m.Y
				room.Lock.Unlock()
			}

		case "shoot":
			room := findPlayerRoom(c.Player.ID)
			if room != nil {
				room.Lock.Lock()
				b := &model.Bullet{
					ID:    uuid.New().String(),
					X:     c.Player.X,
					Y:     c.Player.Y,
					Owner: c.Player.ID,
					Speed: 10,
				}
				room.Bullets = append(room.Bullets, b)
				room.Lock.Unlock()
			}

		default:
			HubInstance.Broadcast <- msg
		}
	}
}

func (c *Client) WritePump() {
	defer c.Player.Conn.Close()
	for msg := range c.Send {
		err := c.Player.Conn.WriteMessage(1, msg)
		if err != nil {
			log.Println("write error:", err)
			break
		}
	}
}

// 根据玩家ID找到房间
func findPlayerRoom(playerID string) *model.Room {
	RoomLock.Lock()
	defer RoomLock.Unlock()
	for _, r := range RoomMap {
		for _, p := range r.Players {
			if p.ID == playerID {
				return r
			}
		}
	}
	return nil
}

```

### internal/api/ws.go
```go
package api

import (
	"airplane-war/internal/model"
	"airplane-war/internal/ws"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/google/uuid"
	"log"
	"net/http"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func WsHandler(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println("upgrade error:", err)
		return
	}

	playerID := uuid.New().String()
	player := &model.Player{
		ID:   playerID,
		Name: "玩家-" + playerID[:4],
		X:    0,
		Y:    0,
		Conn: conn,
	}

	client := ws.NewClientWithPlayer(player)
	go client.ReadPump()
	go client.WritePump()

	ws.HubInstance.Register <- client
}

```
### 前端测试
发送匹配：
```json
{"action":"match"}

```
移动玩家：
```json
{"action":"move", "x":100, "y":200}
```
房间循环每50ms 同步：
```json
{
  "type": "game_state",
  "players": [{"id":"xxx","name":"玩家-xxx","x":100,"y":200}, ...],
  "bullets": [{"id":"xxx","x":100,"y":200,"owner":"xxx","speed":10}, ...]
}

```