package server

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/gorilla/websocket"
	game "github.com/lllllan02/texas-holdem/demos/05-game-state-machine"
)

const (
	ActionJoin       = "join"
	ActionLeave      = "leave"
	ActionGetPrivate = "get_private"
	ActionKick       = "kick"
	ActionChat       = "chat"
	ActionSit        = "sit"
	ActionBet        = "bet"
	ActionFold       = "fold"
	ActionRoomState  = "room_state"
)

// ClientMessage 接收前端发来的 JSON 格式消息
type ClientMessage struct {
	UserID   string `json:"userId"`
	Username string `json:"username"`
	Action   string `json:"action"`
	Content  string `json:"content"`
}

// ServerMessage 后端广播给所有人的 JSON 格式消息
type ServerMessage struct {
	Time     string   `json:"time"`
	UserID   string   `json:"userId,omitempty"`
	Username string   `json:"username,omitempty"`
	Action   string   `json:"action"`
	Content  string   `json:"content,omitempty"`
	RoomID   string   `json:"roomId,omitempty"`
	Players  []string `json:"players,omitempty"`
}

// Client 代表一个连接到房间的 WebSocket 客户端
type Client struct {
	ID       string // 对应 UserID
	Username string
	Conn     *websocket.Conn
	Room     *Room
	Send     chan []byte // 缓冲通道，用于向客户端发送消息
	closed   bool        // 标记 Send 通道是否已关闭，防止重复 close 导致 panic
}

func (c *Client) handleAction(inMsg ClientMessage) {
	switch inMsg.Action {
	case ActionJoin:
		c.handleJoin()
	case ActionGetPrivate:
		c.handleGetPrivate()
	default:
		c.handleDefaultAction(inMsg)
	}
}

func (c *Client) handleJoin() {
	outMsg := ServerMessage{
		Time:     time.Now().Format("15:04:05"),
		UserID:   c.ID,
		Username: c.Username,
		Action:   ActionJoin,
	}
	outBytes, _ := json.Marshal(outMsg)
	c.Room.Hub.Broadcast <- outBytes
}

func (c *Client) handleGetPrivate() {
	privateMsg := ServerMessage{
		Time:     time.Now().Format("15:04:05"),
		UserID:   "system",
		Username: "系统",
		Action:   ActionGetPrivate,
		Content:  "这是只有你能看到的私有消息：你的底牌是 ♠A ♥K",
	}
	privateBytes, _ := json.Marshal(privateMsg)

	select {
	case c.Send <- privateBytes:
	default:
	}
}

func (c *Client) handleDefaultAction(inMsg ClientMessage) {
	outMsg := ServerMessage{
		Time:     time.Now().Format("15:04:05"),
		UserID:   c.ID,
		Username: c.Username,
		Action:   inMsg.Action,
		Content:  inMsg.Content,
	}
	outBytes, _ := json.Marshal(outMsg)
	c.Room.Hub.Broadcast <- outBytes
}

// ReadPump 从 WebSocket 连接中读取消息
func (c *Client) ReadPump() {
	defer func() {
		// 广播离开消息
		if c.ID != "" {
			outMsg := ServerMessage{
				Time:     time.Now().Format("15:04:05"),
				UserID:   c.ID,
				Username: c.Username,
				Action:   ActionLeave,
			}
			outBytes, _ := json.Marshal(outMsg)
			c.Room.Hub.Broadcast <- outBytes
		}

		c.Room.RemoveClient(c)
		c.Conn.Close()
	}()
	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}

		var inMsg ClientMessage
		if err := json.Unmarshal(message, &inMsg); err != nil {
			log.Printf("Invalid JSON received: %v", string(message))
			continue
		}

		c.handleAction(inMsg)
	}
}

// WritePump 将消息写入 WebSocket 连接
func (c *Client) WritePump() {
	defer func() {
		c.Conn.Close()
	}()
	for message := range c.Send {
		w, err := c.Conn.NextWriter(websocket.TextMessage)
		if err != nil {
			return
		}
		w.Write(message)

		n := len(c.Send)
		for i := 0; i < n; i++ {
			w.Write([]byte{'\n'})
			w.Write(<-c.Send)
		}

		if err := w.Close(); err != nil {
			return
		}
	}
	// 当 c.Send 通道被关闭时，循环结束，发送关闭消息
	c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
}

// Hub 模拟房间专属的 WebSocket 广播中心
type Hub struct {
	Room       *Room
	Clients    map[*Client]bool
	Users      map[string]*Client // 用户ID到客户端的映射，用于单播和顶号逻辑
	Broadcast  chan []byte
	Register   chan *Client
	Unregister chan *Client
}

// Run 启动 Hub 的事件循环
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.Register:
			// 顶号逻辑：如果该 UserID 已经有连接了，踢掉旧的
			h.kickOutOldClient(client.ID)

			h.Clients[client] = true
			h.Users[client.ID] = client
			fmt.Printf("客户端加入 Hub: %s\n", client.ID)
			h.broadcastRoomState() // 广播最新房间状态

		case client := <-h.Unregister:
			if _, ok := h.Clients[client]; ok {
				delete(h.Clients, client)
				if h.Users[client.ID] == client {
					delete(h.Users, client.ID)
				}

				if !client.closed {
					client.closed = true
					close(client.Send)
				}
				fmt.Printf("客户端离开 Hub: %s\n", client.ID)
				h.broadcastRoomState() // 广播最新房间状态
			}

		case message := <-h.Broadcast:
			for client := range h.Clients {
				if client.closed {
					continue
				}
				select {
				case client.Send <- message:
				default:
					if !client.closed {
						client.closed = true
						close(client.Send)
					}
					delete(h.Clients, client)
				}
			}
		}
	}
}

// kickOutOldClient 踢掉已经存在的旧连接（顶号逻辑）
func (h *Hub) kickOutOldClient(userID string) {
	if oldClient, ok := h.Users[userID]; ok {
		if oldClient.closed {
			return
		}

		kickMsg := ServerMessage{
			Time:    time.Now().Format("15:04:05"),
			Action:  ActionKick,
			Content: "您的账号在其他地方登录，您已被强制下线。",
		}
		kickBytes, _ := json.Marshal(kickMsg)

		select {
		case oldClient.Send <- kickBytes:
		default:
		}

		oldClient.closed = true
		close(oldClient.Send)
		delete(h.Clients, oldClient)
	}
}

// broadcastRoomState 广播当前房间内的玩家列表
func (h *Hub) broadcastRoomState() {
	players := make([]string, 0, len(h.Clients))
	for client := range h.Clients {
		players = append(players, client.Username+" ("+client.ID+")")
	}

	stateMsg := ServerMessage{
		Time:    time.Now().Format("15:04:05"),
		Action:  ActionRoomState,
		RoomID:  h.Room.ID,
		Players: players,
	}

	stateBytes, err := json.Marshal(stateMsg)
	if err != nil {
		log.Println("JSON 序列化失败:", err)
		return
	}

	for client := range h.Clients {
		if client.closed {
			continue
		}
		select {
		case client.Send <- stateBytes:
		default:
			if !client.closed {
				client.closed = true
				close(client.Send)
			}
			delete(h.Clients, client)
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
	}

	room.Hub = &Hub{
		Room:       room,
		Clients:    make(map[*Client]bool),
		Users:      make(map[string]*Client),
		Broadcast:  make(chan []byte),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
	}

	go room.Hub.Run()

	return room
}

// AddClient 将客户端加入房间
func (r *Room) AddClient(client *Client) error {
	fmt.Printf("客户端 [%s] 准备加入房间 [%s]\n", client.ID, r.ID)
	client.Room = r
	r.Hub.Register <- client
	return nil
}

// RemoveClient 将客户端移出房间
func (r *Room) RemoveClient(client *Client) {
	fmt.Printf("客户端 [%s] 离开房间 [%s]\n", client.ID, r.ID)
	r.Hub.Unregister <- client
}
