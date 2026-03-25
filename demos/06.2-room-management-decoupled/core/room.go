package core

import (
	"log"
	"time"
)

// Room 房间，包含一个游戏桌和一个广播中心
type Room struct {
	ID      string
	HostID  string       // 房主ID
	Engine  GameEngine   // 抽象的游戏引擎
	Hub     *Hub         // 该房间专属的 WebSocket 广播中心
	Manager *RoomManager // 引用全局管理器，用于自动回收
}

// GetID 获取房间 ID
func (r *Room) GetID() string {
	return r.ID
}

// GetHostID 获取房主 ID
func (r *Room) GetHostID() string {
	return r.HostID
}

// Broadcast 向房间内所有客户端广播消息，可以指定发送者（如果是系统消息，sender 传 nil）
// 注意：此方法专供【外部 Goroutine】（如 Client.ReadPump 或 Engine）调用。
// 它通过 Channel 将消息发给 Hub，保证对 Clients map 的并发安全。
func (r *Room) Broadcast(sender *Client, action string, content string) {
	outBytes := BuildServerMessage(sender, action, content)
	r.Hub.Broadcast <- outBytes
}

// SendTo 向指定的客户端发送单播消息
// 注意：此方法专供【外部 Goroutine】调用。
// 发送失败时只会丢弃消息，不会清理连接（清理连接必须由 Hub.Run 内部执行）。
func (r *Room) SendTo(client *Client, action string, content string) {
	outBytes := BuildServerMessage(nil, action, content)

	select {
	case client.Send <- outBytes:
	default:
	}
}

// NewRoom 创建一个新房间
func NewRoom(id string, hostID string, engine GameEngine, manager *RoomManager) *Room {
	room := &Room{
		ID:      id,
		HostID:  hostID,
		Engine:  engine,
		Manager: manager,
	}

	room.Hub = &Hub{
		Room:        room,
		Clients:     make(map[*Client]bool),
		Users:       make(map[string]*Client),
		Broadcast:   make(chan []byte),
		Register:    make(chan *Client),
		Unregister:  make(chan *Client),
		DestroyChan: make(chan struct{}),
		// 初始化定时器并立即停止，等待需要时再 Reset
		roomRecycleTimer:  time.NewTimer(time.Hour * 24),
		hostTransferTimer: time.NewTimer(time.Hour * 24),
	}
	room.Hub.roomRecycleTimer.Stop()
	room.Hub.hostTransferTimer.Stop()

	if room.Engine != nil {
		room.Engine.OnInit(room)
	}

	go room.Hub.Run()

	return room
}

// AddClient 将客户端加入房间
func (r *Room) AddClient(client *Client) error {
	log.Printf("客户端 [%s] 准备加入房间 [%s]\n", client.ID, r.ID)
	client.Room = r
	r.Hub.Register <- client

	if r.Engine != nil {
		r.Engine.OnPlayerJoin(client)
	}

	return nil
}

// RemoveClient 将客户端移出房间
func (r *Room) RemoveClient(client *Client) {
	log.Printf("客户端 [%s] 离开房间 [%s]\n", client.ID, r.ID)
	r.Hub.Unregister <- client

	if r.Engine != nil {
		r.Engine.OnPlayerLeave(client)
	}
}
