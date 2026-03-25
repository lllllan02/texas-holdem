package server

import (
	"errors"
	"fmt"
	"net/http"
	"sync"

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
	return &RoomManager{
		Rooms: make(map[string]*Room),
	}
}

// CreateRoom 生成一个随机的 6 位数房间号，初始化 Room 并存入 map
func (m *RoomManager) CreateRoom(roomID string) *Room {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 实际应用中需要随机生成 roomID 并确保不重复
	room := NewRoom(roomID)
	m.Rooms[roomID] = room
	
	fmt.Printf("创建房间成功: [%s]\n", roomID)
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

	// TODO: 设定房间最大人数限制拦截 (例如 8 人)
	return room.AddClient(client)
}

// RemoveRoom 当房间空了之后，清理资源
func (m *RoomManager) RemoveRoom(roomID string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// TODO: 检查房间是否真的空了
	delete(m.Rooms, roomID)
	fmt.Printf("房间已销毁: [%s]\n", roomID)
}

// ServeWS 处理 WebSocket 连接请求 (HTTP Handler 骨架)
func (m *RoomManager) ServeWS(w http.ResponseWriter, r *http.Request) {
	// 1. 获取 URL 参数中的 roomID 和 playerID
	roomID := r.URL.Query().Get("room")
	playerID := r.URL.Query().Get("player")

	if roomID == "" || playerID == "" {
		http.Error(w, "Missing room or player parameter", http.StatusBadRequest)
		return
	}

	// 2. 升级 HTTP 连接为 WebSocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("WebSocket 升级失败:", err)
		return
	}

	// 3. 创建 Client 实例
	client := &Client{
		ID:   playerID,
		Conn: conn,
		Send: make(chan []byte, 256),
	}

	// 4. 尝试加入房间
	err = m.JoinRoom(roomID, client)
	if err != nil {
		fmt.Printf("玩家 [%s] 加入房间 [%s] 失败: %v\n", playerID, roomID, err)
		conn.WriteMessage(websocket.TextMessage, []byte("Error: "+err.Error()))
		conn.Close()
		return
	}

	// 5. 启动读写 Goroutine (骨架)
	// go client.WritePump()
	// go client.ReadPump()
	
	fmt.Printf("玩家 [%s] 成功连接到房间 [%s] 的 WebSocket\n", playerID, roomID)
}
