package api

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"plane_war/internal/model"
	"plane_war/internal/model/ctype"
	"plane_war/internal/model/res"
	"plane_war/internal/service/game"
	"plane_war/internal/service/redis_service"
	"plane_war/internal/utils/jwts"
	"plane_war/internal/ws"
	"strings"
	"sync"
	"time"
)

// 创建公共房间
func CreatePublicRoom(c *gin.Context) {
	// 解析 token
	_cliams, _ := c.Get("claims")
	claims := _cliams.(*jwts.CustomClaims)

	var req struct {
		Capacity int `json:"capacity" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		res.FailWithMsg("参数错误", c)
		return
	}

	player := &model.Player{
		UserID: claims.UserID,
		Name:   claims.Nickname,
	}
	fmt.Println(player)
	room := &model.PublicRoom{
		ID:       uuid.New().String(),
		OwnerID:  claims.UserID,
		Code:     uuid.New().String()[:6], // 房间码取前6位
		Players:  []*model.Player{player},
		Capacity: req.Capacity,
		Status:   ctype.Waiting,
		Created:  time.Now(),
	}

	// 保存到 Redis（自动会删除之前的房间并绑定 userID）
	if err := redis_service.SavePublicRoomToRedis(room, claims.UserID); err != nil {
		res.FailWithMsg(fmt.Sprintf("创建房间失败: %v", err), c)
		return
	}

	res.OkWithData(room, c)
}

// 获取房间列表
func GetPublicRooms(c *gin.Context) {
	rooms, err := redis_service.GetPublicRoomsList()
	if err != nil {
		res.FailWithMsg(fmt.Sprintf("获取房间列表失败: %v", err), c)
		return
	}
	res.OkWithData(rooms, c)
}

// 加入公共房间
func JoinPublicRoom(c *gin.Context) {
	// 解析 token
	_cliams, _ := c.Get("claims")
	claims := _cliams.(*jwts.CustomClaims)
	fmt.Println(claims.UserID)
	var req struct {
		RoomCode string `json:"room_code" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		res.FailWithMsg("参数错误", c)
		return
	}

	player := &model.Player{
		UserID: claims.UserID,
		Name:   claims.Nickname,
	}

	if err := redis_service.AddPlayerToRoom(req.RoomCode, player); err != nil {
		res.FailWithMsg(fmt.Sprintf("加入房间失败: %v", err), c)
		return
	}

	res.OkWithMsg("加入房间成功", c)
}

// DismissPublicRoom 解散房间
func DismissPublicRoom(c *gin.Context) {
	_cliams, _ := c.Get("claims")
	claims := _cliams.(*jwts.CustomClaims)
	fmt.Println(claims.UserID)
	// 调用解散逻辑
	if err := redis_service.DismissPublicRoom(claims.UserID); err != nil {
		res.FailWithMsg(fmt.Sprintf("解散房间失败: %v", err), c)
		return
	}

	res.OkWithMsg("房间解散成功", c)
}

// 开始游戏
func StartGame(c *gin.Context) {
	claims, _ := c.Get("claims")
	user := claims.(*jwts.CustomClaims)

	roomCode, err := redis_service.GetUserRoomCode(user.UserID)
	if err != nil {
		res.FailWithMsg("房间不存在或已解散", c)
		return
	}

	room, err := redis_service.GetPublicRoomByCode(roomCode)
	if err != nil {
		res.FailWithMsg("获取房间失败", c)
		return
	}

	if len(room.Players) != room.Capacity {
		res.FailWithMsg(fmt.Sprintf("房间人数不合法，当前人数 %d", len(room.Players)), c)
		return
	}

	var gamePlayers []*model.Player
	for _, p := range room.Players {
		client := ws.FindClientByUserID(p.UserID)
		if client == nil {
			res.FailWithMsg(fmt.Sprintf("玩家 %s 未连接", p.Name), c)
			return
		}
		p.Conn = client.Player.Conn
		gamePlayers = append(gamePlayers, p)
	}

	gameRoom := &model.Room{
		ID:      room.ID,
		Players: gamePlayers,
		Bullets: []*model.Bullet{},
		Lock:    sync.Mutex{},
		Quit:    make(chan bool),
	}

	gameRoom.Lock.Lock()
	gameRoom.Players[0].X = 100
	gameRoom.Players[0].Y = 50
	gameRoom.Players[0].HP = 100
	gameRoom.Players[0].Position = "top"

	gameRoom.Players[1].X = 100
	gameRoom.Players[1].Y = 500
	gameRoom.Players[1].HP = 100
	gameRoom.Players[1].Position = "bottom"
	gameRoom.Lock.Unlock()

	state := map[string]interface{}{
		"type":    "match_success",
		"room_id": gameRoom.ID,
		"players": gameRoom.Players,
	}
	data, _ := json.Marshal(state)
	for _, p := range gameRoom.Players {
		p.Conn.WriteMessage(websocket.TextMessage, data)
	}

	ws.RoomLock.Lock()
	ws.RoomMap[gameRoom.ID] = gameRoom
	ws.RoomLock.Unlock()

	game.StartRoomLoop(gameRoom)

	res.OkWithMsg("游戏已开始", c)
}

// parseTokenFromHeader 解析 Authorization Bearer Token
func parseTokenFromHeader(c *gin.Context) (*jwts.CustomClaims, error) {
	token := c.GetHeader("Authorization")
	if token == "" {
		return nil, fmt.Errorf("未携带 token")
	}

	parts := strings.Fields(token)
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		return nil, fmt.Errorf("Authorization 格式错误")
	}
	tokenStr := parts[1]

	claims, err := jwts.ParseToken(tokenStr)
	if err != nil {
		return nil, fmt.Errorf("无效的 token")
	}

	return claims, nil
}
