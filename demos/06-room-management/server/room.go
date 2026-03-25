package server

import (
	"fmt"

	"github.com/gorilla/websocket"
	game "github.com/lllllan02/texas-holdem/demos/05-game-state-machine"
)

// Client 模拟一个连接到房间的 WebSocket 客户端
type Client struct {
	ID   string
	Conn *websocket.Conn
	Room *Room
	Send chan []byte // 缓冲通道，用于向客户端发送消息
}

// ReadPump 从 WebSocket 连接中读取消息 (骨架)
func (c *Client) ReadPump() {
	defer func() {
		// TODO: 客户端断开连接时，通知 Room 移除该客户端
		c.Conn.Close()
	}()
	for {
		// TODO: 读取客户端发送的消息
		// _, message, err := c.Conn.ReadMessage()
		// TODO: 将消息转发给 Room 处理 (例如玩家下注、弃牌等操作)
		break // 骨架代码，直接退出循环
	}
}

// WritePump 将消息写入 WebSocket 连接 (骨架)
func (c *Client) WritePump() {
	defer func() {
		c.Conn.Close()
	}()
	for {
		// TODO: 监听 c.Send 通道，将消息写入 c.Conn
		break // 骨架代码，直接退出循环
	}
}

// Hub 模拟房间专属的 WebSocket 广播中心
type Hub struct {
	Clients    map[*Client]bool
	Broadcast  chan []byte
	Register   chan *Client
	Unregister chan *Client
}

// Run 启动 Hub 的事件循环 (骨架)
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.Register:
			// TODO: 将 client 添加到 h.Clients
			fmt.Printf("客户端加入 Hub: %s\n", client.ID)
		case client := <-h.Unregister:
			// TODO: 将 client 从 h.Clients 移除，并关闭 client.Send 通道
			fmt.Printf("客户端离开 Hub: %s\n", client.ID)
		case message := <-h.Broadcast:
			// TODO: 遍历 h.Clients，将 message 发送到每个 client.Send 通道
			fmt.Printf("Hub 广播消息: %s\n", string(message))
		}
	}
}

// Room 房间，包含一个游戏桌和一个广播中心
type Room struct {
	ID    string
	Table *game.Table // 游戏逻辑状态机
	Hub   *Hub        // 该房间专属的 WebSocket 广播中心
}

// NewRoom 创建一个新房间
func NewRoom(id string) *Room {
	room := &Room{
		ID:    id,
		Table: &game.Table{}, // 实际需要初始化 Table
		Hub: &Hub{
			Clients:    make(map[*Client]bool),
			Broadcast:  make(chan []byte),
			Register:   make(chan *Client),
			Unregister: make(chan *Client),
		},
	}
	// TODO: 启动 Hub 的事件循环 (go room.Hub.Run())
	return room
}

// AddClient 将客户端加入房间
func (r *Room) AddClient(client *Client) error {
	fmt.Printf("客户端 [%s] 准备加入房间 [%s]\n", client.ID, r.ID)
	// TODO: 检查房间是否已满
	// TODO: 将 client 注册到 r.Hub (r.Hub.Register <- client)
	// TODO: 将玩家信息同步到 r.Table.Players
	return nil
}

// RemoveClient 将客户端移出房间
func (r *Room) RemoveClient(client *Client) {
	fmt.Printf("客户端 [%s] 离开房间 [%s]\n", client.ID, r.ID)
	// TODO: 将 client 从 r.Hub 注销 (r.Hub.Unregister <- client)
	// TODO: 从 r.Table.Players 移除或标记为离线
}
