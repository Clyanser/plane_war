package redis_service

import (
	"encoding/json"
	"fmt"
	"plane_war/internal/global"
	"plane_war/internal/model"
	"time"

	"github.com/go-redis/redis"
)

// SavePublicRoomToRedis 保存公共房间到 Redis
func SavePublicRoomToRedis(room *model.PublicRoom, userID uint) error {
	// 检查用户是否已有房间
	existingRoomCode, _ := global.Redis.Get(fmt.Sprintf("userRoom:%d", userID)).Result()
	if existingRoomCode != "" {
		// 删除之前的房间
		err := DeletePublicRoom(existingRoomCode)
		if err != nil {
			return err
		}
	}

	// 保存房间码到 ZSet
	score := float64(room.Created.Unix())
	if err := global.Redis.ZAdd("publicRooms", redis.Z{
		Score:  score,
		Member: room.Code,
	}).Err(); err != nil {
		return fmt.Errorf("保存房间失败: %v", err)
	}

	// 保存房间详细信息
	roomKey := "publicRoom:" + room.Code
	roomData, _ := json.Marshal(room)
	if err := global.Redis.Set(roomKey, roomData, 24*time.Hour).Err(); err != nil {
		return fmt.Errorf("保存房间详细信息失败: %v", err)
	}

	// 绑定用户与房间
	if err := global.Redis.Set(fmt.Sprintf("userRoom:%d", userID), room.Code, 24*time.Hour).Err(); err != nil {
		return fmt.Errorf("绑定用户房间失败: %v", err)
	}

	return nil
}

// GetPublicRoomByCode 获取房间详细信息
func GetPublicRoomByCode(code string) (*model.PublicRoom, error) {
	roomKey := "publicRoom:" + code
	data, err := global.Redis.Get(roomKey).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, fmt.Errorf("房间不存在")
		}
		return nil, fmt.Errorf("获取房间信息失败: %v", err)
	}

	var room model.PublicRoom
	if err := json.Unmarshal([]byte(data), &room); err != nil {
		return nil, fmt.Errorf("解析房间数据失败: %v", err)
	}
	return &room, nil
}

// AddPlayerToRoom 将玩家加入房间，避免重复加入
func AddPlayerToRoom(roomCode string, player *model.Player) error {
	room, err := GetPublicRoomByCode(roomCode)
	if err != nil {
		return err
	}

	// 避免重复加入
	for _, p := range room.Players {
		if p.UserID == player.UserID {
			return fmt.Errorf("玩家已在房间内")
		}
	}

	if len(room.Players) >= room.Capacity {
		return fmt.Errorf("房间已满")
	}

	room.Players = append(room.Players, player)

	// 更新房间详细信息
	roomData, _ := json.Marshal(room)
	if err := global.Redis.Set("publicRoom:"+roomCode, roomData, 24*time.Hour).Err(); err != nil {
		return fmt.Errorf("更新房间信息失败: %v", err)
	}

	return nil
}

// GetPublicRoomsList 获取所有公共房间，按创建时间排序
func GetPublicRoomsList() ([]*model.PublicRoom, error) {
	roomCodes, err := global.Redis.ZRevRange("publicRooms", 0, -1).Result()
	if err != nil {
		return nil, fmt.Errorf("获取房间列表失败: %v", err)
	}

	var rooms []*model.PublicRoom
	for _, code := range roomCodes {
		room, err := GetPublicRoomByCode(code)
		if err != nil {
			continue
		}
		rooms = append(rooms, room)
	}

	return rooms, nil
}

// DismissPublicRoom 解散房间（仅房主可操作）
func DismissPublicRoom(userID uint) error {
	// 查找用户与房间的映射关系
	roomCode, err := global.Redis.Get(fmt.Sprintf("userRoom:%d", userID)).Result()
	if err != nil {
		if err == redis.Nil {
			return fmt.Errorf("房间不存在或您不是房主")
		}
		return fmt.Errorf("获取用户房间失败: %v", err)
	}
	// 删除房间数据
	if err := DeletePublicRoom(roomCode); err != nil {
		return fmt.Errorf("解散房间失败: %v", err)
	}
	// 删除用户与房间的映射
	if err := global.Redis.Del(fmt.Sprintf("userRoom:%d", userID)).Err(); err != nil {
		return fmt.Errorf("解绑用户房间失败: %v", err)
	}
	return nil
}

// RemovePlayerFromRoom 将玩家从房间中移除
func RemovePlayerFromRoom(roomCode string, userID uint) error {
	room, err := GetPublicRoomByCode(roomCode)
	if err != nil {
		return err
	}
	// 查找并移除玩家
	found := false
	newPlayers := make([]*model.Player, 0, len(room.Players))
	for _, p := range room.Players {
		if p.UserID == userID {
			found = true
			continue
		}
		newPlayers = append(newPlayers, p)
	}
	if !found {
		return fmt.Errorf("玩家不在房间中")
	}
	room.Players = newPlayers

	// 如果房间没人了，直接删除房间
	if len(room.Players) == 0 {
		return DeletePublicRoom(roomCode)
	}

	// 更新房间信息
	roomData, _ := json.Marshal(room)
	if err := global.Redis.Set("publicRoom:"+roomCode, roomData, 24*time.Hour).Err(); err != nil {
		return fmt.Errorf("更新房间信息失败: %v", err)
	}

	// 解绑用户和房间关系
	if err := global.Redis.Del(fmt.Sprintf("userRoom:%d", userID)).Err(); err != nil {
		return fmt.Errorf("解绑用户房间失败: %v", err)
	}

	return nil
}

// DeletePublicRoom 删除公共房间
func DeletePublicRoom(code string) error {
	if err := global.Redis.ZRem("publicRooms", code).Err(); err != nil {
		return err
	}
	if err := global.Redis.Del("publicRoom:" + code).Err(); err != nil {
		return err
	}
	return nil
}

func GetUserRoomCode(userID uint) (string, error) {
	key := fmt.Sprintf("userRoom:%d", userID)
	code, err := global.Redis.Get(key).Result()
	if err != nil {
		if err == redis.Nil {
			return "", fmt.Errorf("用户没有房间")
		}
		return "", fmt.Errorf("获取用户房间失败: %v", err)
	}
	return code, nil
}
