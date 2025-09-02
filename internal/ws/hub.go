package ws

import (
	"encoding/json"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"plane_war/internal/global"
	"plane_war/internal/model"
	"plane_war/internal/service/game"
	"plane_war/internal/service/match"
	"sync"
)

type Client struct {
	Player *model.Player //关联玩家信息
	Send   chan []byte   //消息通道
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
			global.Log.Printf("new player connected :%s", client.Player.Name)
		case client := <-h.Unregister:
			if _, ok := h.Clients[client]; ok {
				delete(h.Clients, client)
				close(client.Send)
				global.Log.Printf("player disconnected : %s", client.Player.Name)
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

// --------消息处理--------
type Message struct {
	Action string `json:"action"`
	X      int    `json:"x,omitempty"`
	Y      int    `json:"y,omitempty"`
}

var RoomMap = make(map[string]*model.Room)
var RoomLock sync.Mutex

func (c *Client) ReadPump() {
	defer func() {
		HubInstance.Unregister <- c
		c.Player.Conn.Close()
	}()
	for {
		_, msg, err := c.Player.Conn.ReadMessage()
		if err != nil {
			global.Log.Println("read error:", err)
			break
		}
		var m Message
		if err := json.Unmarshal(msg, &m); err != nil {
			global.Log.Println("json prase error:", err)
			continue
		}
		switch m.Action {
		case "match":
			room := match.MatchQueueInstance.AddPlayer(c.Player)
			if room != nil {
				RoomLock.Lock()
				RoomMap[room.ID] = room
				RoomLock.Unlock()

				// 设置玩家位置、血量、上下标识
				room.Lock.Lock()
				room.Players[0].X = 100
				room.Players[0].Y = 50
				room.Players[0].HP = 100
				room.Players[0].Position = "top"

				room.Players[1].X = 100
				room.Players[1].Y = 500
				room.Players[1].HP = 100
				room.Players[1].Position = "bottom"
				room.Lock.Unlock()

				// 发送匹配成功消息给双方
				state := map[string]interface{}{
					"type":    "match_success",
					"room_id": room.ID,
					"players": room.Players,
				}
				data, _ := json.Marshal(state)
				for _, p := range room.Players {
					p.Conn.WriteMessage(websocket.TextMessage, data)
				}

				global.Log.Printf("匹配成功，房间id ：%s", room.ID)
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
					ID:     uuid.New().String(),
					X:      c.Player.X + 22, //子弹从飞机中心射出
					Y:      c.Player.Y,
					Owner:  c.Player.ID,
					Damage: 10,
				}
				if c.Player.Position == "top" {
					b.Speed = -10 //向下
				} else {
					b.Speed = 10 //向上
				}
				room.Bullets = append(room.Bullets, b)
				room.Lock.Unlock()
			}

		default:
			// echo 消息
			HubInstance.Broadcast <- msg
		}
	}
}

func (c *Client) WritePump() {
	defer c.Player.Conn.Close()
	for msg := range c.Send {
		err := c.Player.Conn.WriteMessage(websocket.TextMessage, msg)
		if err != nil {
			global.Log.Println("write error:", err)
			break
		}
		global.Log.Printf("发给玩家 %s: %s", c.Player.ID, msg)
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

// 发送完整房间状态
func sendRoomState(room *model.Room) {
	room.Lock.Lock()
	defer room.Lock.Unlock()

	state := map[string]interface{}{
		"type":    "game_state",
		"players": room.Players,
		"bullets": room.Bullets,
	}

	data, _ := json.Marshal(state)
	for _, p := range room.Players {
		p.Conn.WriteMessage(websocket.TextMessage, data)
	}
}
