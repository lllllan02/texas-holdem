package main

import (
	"bytes"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

const (
	// 允许向对等方写入消息的时间
	writeWait = 10 * time.Second

	// 允许读取下一个 pong 消息的时间
	pongWait = 60 * time.Second

	// 在此期间向对等方发送 ping
	pingPeriod = (pongWait * 9) / 10

	// 允许的最大消息大小
	maxMessageSize = 512
)

var (
	newline = []byte{'\n'}
	space   = []byte{' '}
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	// 允许跨域请求（测试用）
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// Client 是 websocket 连接和 hub 之间的中间人
type Client struct {
	hub *Hub

	// websocket 连接
	Conn *websocket.Conn

	// 待发送消息的缓冲通道
	Send chan []byte
}

// readPump 从 websocket 连接将消息泵送到 hub
//
// 应用程序在每个连接的 goroutine 中运行 readPump。
// 应用程序通过执行此 goroutine 中的所有读取来确保连接最多有一个读取器。
func (c *Client) readPump() {
	defer func() {
		c.hub.Unregister <- c
		c.Conn.Close()
		log.Println("Client disconnected (readPump exited)")
	}()

	c.Conn.SetReadLimit(maxMessageSize)
	c.Conn.SetReadDeadline(time.Now().Add(pongWait))
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
		message = bytes.TrimSpace(bytes.Replace(message, newline, space, -1))
		c.hub.Broadcast <- message
	}
}

// writePump 将消息从 hub 泵送到 websocket 连接
//
// 应用程序在每个连接的 goroutine 中运行 writePump。
// 应用程序通过执行此 goroutine 中的所有写入来确保连接最多有一个写入器。
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
				// hub 关闭了通道
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// 将排队的聊天消息添加到当前的 websocket 消息中
			n := len(c.Send)
			for i := 0; i < n; i++ {
				w.Write(newline)
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

// ServeWS 处理来自对等方的 websocket 请求
func ServeWS(hub *Hub, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Upgrade error:", err)
		return
	}

	client := &Client{hub: hub, Conn: conn, Send: make(chan []byte, 256)}
	client.hub.Register <- client

	log.Println("New client connected")

	// 允许在新的 goroutine 中收集调用者引用的内存
	go client.writePump()
	go client.readPump()
}
