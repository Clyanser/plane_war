### 🎯 阶段三目标
1. 玩家点击“匹配” → 进入 匹配队列
2. 当匹配队列里有两名玩家 → 创建一个 房间
3. 房间里只包含这两名玩家
4. 给两名玩家发送 匹配成功通知

### 项目目录结构
game-server/
├── cmd/
│   └── server/
│       └── main.go
├── internal/
│   ├── api/
│   │   └── ws.go
│   ├── model/
│   │   ├── player.go
│   │   └── room.go
│   ├── service/
│   │   └── match/
│   │       └── match.go
│   └── ws/
│       └── hub.go
└── static/
└── html/
└── test.html

### internal/model/player.go
```go
package model

import "github.com/gorilla/websocket"

// Player 玩家信息
type Player struct {
	ID   string
	Name string
	Conn *websocket.Conn
}

```
### internal/model/room.go
```go
package model

import "sync"

// Room 房间信息
type Room struct {
	ID      string
	Players []*Player
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
### internal/ws/hub.go
```go
package ws

import (
	"airplane-war/internal/model"
	"airplane-war/internal/service/match"
	"encoding/json"
	"log"
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

// ---------------- Read/Write ----------------
type Message struct {
	Action string `json:"action"`
}

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
				for _, p := range room.Players {
					resp := map[string]string{
						"type":    "match_success",
						"room_id": room.ID,
					}
					data, _ := json.Marshal(resp)
					p.Conn.WriteMessage(1, data)
				}
				log.Printf("🎯 匹配成功，房间ID: %s", room.ID)
			}
		default:
			// Echo 消息
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
		Conn: conn,
	}

	client := ws.NewClientWithPlayer(player)
	go client.ReadPump()
	go client.WritePump()

	ws.HubInstance.Register <- client
}

```