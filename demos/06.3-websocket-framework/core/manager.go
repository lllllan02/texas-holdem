package core

import (
	"errors"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/lllllan02/texas-holdem/pkg/wscore"
)

// EngineFactory 游戏引擎工厂函数，用于在创建房间时实例化特定的游戏逻辑
type EngineFactory func() GameEngine

// RoomManager 全局房间管理器
type RoomManager struct {
	Rooms           map[string]*Room
	mu              sync.RWMutex             // 保护并发读写
	engineFactories map[string]EngineFactory // 支持多种游戏引擎
}

// NewRoomManager 创建一个全局房间管理器
func NewRoomManager() *RoomManager {
	return &RoomManager{
		Rooms:           make(map[string]*Room),
		engineFactories: make(map[string]EngineFactory),
	}
}

// RegisterEngine 注册一种游戏引擎
func (m *RoomManager) RegisterEngine(gameType string, factory EngineFactory) {
	m.engineFactories[gameType] = factory
}

// CreateRoom 生成一个随机的 6 位数房间号，初始化 Room 并存入 map
func (m *RoomManager) CreateRoom(hostID string, gameType string) (*Room, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 检查游戏类型是否受支持
	factory, exists := m.engineFactories[gameType]
	if !exists {
		return nil, fmt.Errorf("unsupported game type: %s", gameType)
	}

	// 生成随机的 6 位数房间号，确保不重复
	var roomID string
	for {
		roomID = fmt.Sprintf("%06d", rand.Intn(1000000))
		if _, exists := m.Rooms[roomID]; !exists {
			break
		}
	}

	// 使用对应的工厂创建游戏引擎
	engine := factory()

	room := NewRoom(roomID, hostID, engine, m)
	m.Rooms[roomID] = room

	log.Printf("玩家 [%s] 创建了 [%s] 房间，房间号: [%s]\n", hostID, gameType, roomID)
	return room, nil
}

// GetRoom 获取指定房间
func (m *RoomManager) GetRoom(roomID string) *Room {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.Rooms[roomID]
}

// RemoveRoom 当房间空了之后，清理资源
func (m *RoomManager) RemoveRoom(roomID string) {
	m.mu.Lock()
	room, exists := m.Rooms[roomID]
	if !exists {
		m.mu.Unlock()
		return
	}

	// 1. 通知游戏引擎清理资源
	if room.Engine != nil {
		room.Engine.OnDestroy()
	}
	// 2. 从管理器中移除
	delete(m.Rooms, roomID)
	m.mu.Unlock()

	// 3. 关闭 Hub 广播中心 (在锁外执行，防止死锁)
	room.Hub.Stop()
	log.Printf("房间已销毁: [%s]\n", roomID)
}

// DismissRoom 房主解散房间，踢出所有玩家
func (m *RoomManager) DismissRoom(roomID string, userID string) error {
	m.mu.Lock()
	room, exists := m.Rooms[roomID]
	if !exists {
		m.mu.Unlock()
		return errors.New("room not found")
	}
	if room.HostID != userID {
		m.mu.Unlock()
		return errors.New("not the host")
	}

	// 1. 通知游戏引擎清理资源
	if room.Engine != nil {
		room.Engine.OnDestroy()
	}
	// 2. 从管理器中移除
	delete(m.Rooms, roomID)
	m.mu.Unlock()

	// 3. 触发 Hub 销毁并广播解散消息
	// 广播解散消息
	outBytes := BuildServerMessage(nil, ActionRoomDismissed, "房主已解散房间")
	room.Hub.BroadcastMessage(outBytes)
	// 关闭所有连接
	room.Hub.Stop()
	
	log.Printf("玩家 [%s] 解散了房间: [%s]\n", userID, roomID)
	return nil
}

// ServeWS 处理 WebSocket 连接请求
func (m *RoomManager) ServeWS(c *gin.Context) {
	// 1. 获取 URL 参数中的 roomID 和 playerID(userId) 以及 username
	roomID := c.Query("room")
	userID := c.Query("userId")
	username := c.Query("username")

	if roomID == "" || userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing room or userId parameter"})
		return
	}

	if username == "" {
		username = "游客_" + userID[:4]
	}

	// 2. 检查房间是否存在 (必须先通过 HTTP 接口创建)
	room := m.GetRoom(roomID)
	if room == nil {
		log.Printf("玩家 [%s] 尝试加入不存在的房间 [%s]\n", userID, roomID)
		c.JSON(http.StatusNotFound, gin.H{"error": "Room not found"})
		return
	}

	// 3. 使用 wscore 框架处理 WebSocket 连接
	ctx := &PlayerContext{
		Username: username,
	}
	wscore.ServeWS(room.Hub, room, userID, ctx, c.Writer, c.Request)

	log.Printf("玩家 [%s](%s) 成功连接到房间 [%s] 的 WebSocket\n", username, userID, roomID)
}
