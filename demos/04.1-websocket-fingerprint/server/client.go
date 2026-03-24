package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 1024 // 稍微调大一点以容纳 JSON
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // 允许跨域
	},
}

const (
	ActionJoin       = "join"
	ActionLeave      = "leave"
	ActionGetPrivate = "get_private"
	ActionKick       = "kick"
	ActionChat       = "chat"
	ActionSit        = "sit"
	ActionBet        = "bet"
	ActionFold       = "fold"
)

// ClientMessage 接收前端发来的 JSON 格式消息
type ClientMessage struct {
	UserID   string `json:"userId"`
	Username string `json:"username"`
	Action   string `json:"action"` // "chat", "sit", "bet", "fold"
	Content  string `json:"content"`
}

// ServerMessage 后端广播给所有人的 JSON 格式消息
type ServerMessage struct {
	Time     string `json:"time"`
	UserID   string `json:"userId"`
	Username string `json:"username"`
	Action   string `json:"action"`
	Content  string `json:"content"`
}

type Client struct {
	hub      *Hub
	Conn     *websocket.Conn
	Send     chan []byte
	UserID   string // 记录该连接绑定的用户ID
	Username string // 记录该连接绑定的用户名
	closed   bool   // 标记 Send 通道是否已关闭，防止重复 close 导致 panic
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
		UserID:   c.UserID,
		Username: c.Username,
		Action:   ActionJoin,
	}
	outBytes, _ := json.Marshal(outMsg)
	c.hub.Broadcast <- outBytes
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
		UserID:   c.UserID,
		Username: c.Username,
		Action:   inMsg.Action,
		Content:  inMsg.Content,
	}
	outBytes, _ := json.Marshal(outMsg)
	c.hub.Broadcast <- outBytes
}

func (c *Client) readPump() {
	defer func() {
		// 广播离开消息
		if c.UserID != "" {
			outMsg := ServerMessage{
				Time:     time.Now().Format("15:04:05"),
				UserID:   c.UserID,
				Username: c.Username,
				Action:   ActionLeave,
			}
			outBytes, _ := json.Marshal(outMsg)
			c.hub.Broadcast <- outBytes
		}

		c.hub.Unregister <- c
		c.Conn.Close()
		log.Println("Client disconnected (readPump exited)")
	}()
	c.Conn.SetReadLimit(maxMessageSize)
	if err := c.Conn.SetReadDeadline(time.Now().Add(pongWait)); err != nil {
		return
	}
	c.Conn.SetPongHandler(func(string) error {
		return c.Conn.SetReadDeadline(time.Now().Add(pongWait))
	})

	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}

		// 解析前端发来的 JSON
		var inMsg ClientMessage
		if err := json.Unmarshal(message, &inMsg); err != nil {
			log.Printf("Invalid JSON received: %v", string(message))
			continue
		}

		log.Printf("Received action [%s] from %s (%s)", inMsg.Action, c.Username, c.UserID)

		c.handleAction(inMsg)
	}
}

func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
		log.Println("Client disconnected (writePump exited)")
	}()
	for {
		select {
		case message, ok := <-c.Send:
			if err := c.Conn.SetWriteDeadline(time.Now().Add(writeWait)); err != nil {
				return
			}
			if !ok {
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// 将排队的消息一并发送
			n := len(c.Send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				w.Write(<-c.Send)
			}

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			if err := c.Conn.SetWriteDeadline(time.Now().Add(writeWait)); err != nil {
				return
			}
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func ServeWS(hub *Hub, w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("userId")
	username := r.URL.Query().Get("username")

	if userID == "" {
		log.Println("Connection rejected: missing userId")
		http.Error(w, "Missing userId", http.StatusBadRequest)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Upgrade error:", err)
		return
	}
	
	client := &Client{
		hub:      hub,
		Conn:     conn,
		Send:     make(chan []byte, 256),
		UserID:   userID,
		Username: username,
	}
	client.hub.Register <- client

	log.Printf("New client connected: %s (%s)", username, userID)

	go client.writePump()
	go client.readPump()
}
