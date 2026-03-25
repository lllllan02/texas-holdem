package core

import (
	"errors"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // 允许跨域，方便测试
	},
}

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
	// 注意：这里必须调用 factory() 来创建一个新的引擎实例
	// 如果直接复用一个全局对象，会导致多个房间共享同一个游戏状态（比如同一副牌、同一个奖池）
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

// JoinRoom 将玩家的 WS 连接加入到指定房间的 Hub 中
func (m *RoomManager) JoinRoom(roomID string, client *Client) error {
	m.mu.RLock()
	room, exists := m.Rooms[roomID]
	m.mu.RUnlock()

	if !exists {
		return errors.New("room not found")
	}

	return room.AddClient(client)
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
	close(room.Hub.DestroyChan)
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
	close(room.Hub.DestroyChan)
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

	// 3. 升级 HTTP 连接为 WebSocket
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println("WebSocket 升级失败:", err)
		return
	}

	// 4. 创建 Client 实例
	client := &Client{
		ID:       userID,
		Username: username,
		Conn:     conn,
		Send:     make(chan []byte, 256),
	}

	err = m.JoinRoom(roomID, client)
	if err != nil {
		log.Printf("玩家 [%s] 加入房间 [%s] 失败: %v\n", userID, roomID, err)
		conn.WriteMessage(websocket.TextMessage, []byte("Error: "+err.Error()))
		conn.Close()
		return
	}

	// 5. 启动读写 Goroutine
	go client.writePump()
	go client.readPump()

	log.Printf("玩家 [%s](%s) 成功连接到房间 [%s] 的 WebSocket\n", username, userID, roomID)
}
