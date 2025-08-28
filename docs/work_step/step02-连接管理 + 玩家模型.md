### 🎯 阶段二目标

给每个玩家分配 唯一 ID

将 Client 和玩家信息绑定

实现 玩家注册 / 注销 功能

支持 Hub 广播 或 单个客户端发送消息
### 新增或修改的包结构

internal/
├── model/
│   └── player.go       # 玩家模型
└── ws/
└── hub.go          # 更新后的 Hub 和 Client


## internal/model/player.go

```go
package model

import "github.com/gorilla/websocket"

// Player 玩家信息
type Player struct {
	ID   string          // 唯一玩家ID
	Name string          // 昵称
	Conn *websocket.Conn // WebSocket 连接
}
```
## internal/ws/hub.go
```go
package ws

import (
	"airplane-war/internal/model"
	"github.com/gorilla/websocket"
	"log"
)

// ---------------- Client ----------------
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

// ---------------- Hub ----------------
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
			log.Printf("新玩家连接: %s (%s)", client.Player.ID, client.Player.Name)

		case client := <-h.Unregister:
			if _, ok := h.Clients[client]; ok {
				delete(h.Clients, client)
				close(client.Send)
				log.Printf("玩家断开: %s (%s)", client.Player.ID, client.Player.Name)
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
		log.Printf("玩家 %s 发送: %s", c.Player.ID, msg)
		HubInstance.Broadcast <- msg // Echo 消息
	}
}

func (c *Client) WritePump() {
	defer c.Player.Conn.Close()
	for msg := range c.Send {
		err := c.Player.Conn.WriteMessage(websocket.TextMessage, msg)
		if err != nil {
			log.Println("write error:", err)
			break
		}
		log.Printf("发给玩家 %s: %s", c.Player.ID, msg)
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

// 升级器
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

// WsHandler WebSocket 入口
func WsHandler(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println("upgrade error:", err)
		return
	}

	// 创建 Player
	playerID := uuid.New().String()
	player := &model.Player{
		ID:   playerID,
		Name: "玩家-" + playerID[:4],
		Conn: conn,
	}

	// 创建 Client 并注册到 Hub
	client := ws.NewClientWithPlayer(player)
	go client.ReadPump()
	go client.WritePump()

	ws.HubInstance.Register <- client
}

```