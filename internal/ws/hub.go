package ws

import (
	"encoding/json"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"log"
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
			log.Printf("new player connected :%s", client.Player.Name)
		case client := <-h.Unregister:
			if _, ok := h.Clients[client]; ok {
				delete(h.Clients, client)
				close(client.Send)
				log.Printf("player disconnected : %s", client.Player.Name)
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
			log.Println("read error:", err)
			break
		}
		var m Message
		if err := json.Unmarshal(msg, &m); err != nil {
			log.Println("json prase error:", err)
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
					p.Conn.WriteMessage(websocket.TextMessage, data)
				}
				log.Printf("匹配成功，房间id ：%s ", room.ID)
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
			log.Println("write error:", err)
			break
		}
		log.Printf("发给玩家 %s: %s", c.Player.ID, msg)
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
