package server

import (
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // 允许跨域，方便测试
	},
}

// RoomManager 全局房间管理器
type RoomManager struct {
	Rooms map[string]*Room
	mu    sync.RWMutex // 保护并发读写
}

// NewRoomManager 创建一个全局房间管理器
func NewRoomManager() *RoomManager {
	rand.Seed(time.Now().UnixNano())
	return &RoomManager{
		Rooms: make(map[string]*Room),
	}
}

// CreateRoom 生成一个随机的 6 位数房间号，初始化 Room 并存入 map
func (m *RoomManager) CreateRoom() *Room {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 生成随机的 6 位数房间号，确保不重复
	var roomID string
	for {
		roomID = fmt.Sprintf("%06d", rand.Intn(1000000))
		if _, exists := m.Rooms[roomID]; !exists {
			break
		}
	}

	room := NewRoom(roomID)
	m.Rooms[roomID] = room

	fmt.Printf("创建房间成功: [%s]\n", roomID)
	return room
}

// GetOrCreateRoom 获取指定房间，如果不存在则创建（方便测试时指定房间号）
func (m *RoomManager) GetOrCreateRoom(roomID string) *Room {
	m.mu.Lock()
	defer m.mu.Unlock()

	if room, exists := m.Rooms[roomID]; exists {
		return room
	}

	room := NewRoom(roomID)
	m.Rooms[roomID] = room
	fmt.Printf("创建指定房间成功: [%s]\n", roomID)
	return room
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
	defer m.mu.Unlock()

	delete(m.Rooms, roomID)
	fmt.Printf("房间已销毁: [%s]\n", roomID)
}

// ServeWS 处理 WebSocket 连接请求
func (m *RoomManager) ServeWS(w http.ResponseWriter, r *http.Request) {
	// 1. 获取 URL 参数中的 roomID 和 playerID(userId) 以及 username
	roomID := r.URL.Query().Get("room")
	userID := r.URL.Query().Get("userId")
	username := r.URL.Query().Get("username")

	if roomID == "" || userID == "" {
		http.Error(w, "Missing room or userId parameter", http.StatusBadRequest)
		return
	}

	if username == "" {
		username = "游客_" + userID[:4]
	}

	// 2. 升级 HTTP 连接为 WebSocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("WebSocket 升级失败:", err)
		return
	}

	// 3. 创建 Client 实例
	client := &Client{
		ID:       userID,
		Username: username,
		Conn:     conn,
		Send:     make(chan []byte, 256),
	}

	// 4. 尝试加入房间 (如果房间不存在，为了测试方便我们自动创建)
	m.GetOrCreateRoom(roomID)

	err = m.JoinRoom(roomID, client)
	if err != nil {
		fmt.Printf("玩家 [%s] 加入房间 [%s] 失败: %v\n", userID, roomID, err)
		conn.WriteMessage(websocket.TextMessage, []byte("Error: "+err.Error()))
		conn.Close()
		return
	}

	// 5. 启动读写 Goroutine
	go client.WritePump()
	go client.ReadPump()

	fmt.Printf("玩家 [%s](%s) 成功连接到房间 [%s] 的 WebSocket\n", username, userID, roomID)
}
