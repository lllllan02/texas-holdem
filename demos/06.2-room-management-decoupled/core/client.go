package core

import (
	"encoding/json"
	"log"
	"time"

	"github.com/gorilla/websocket"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 5 * time.Second // 进一步缩短心跳超时时间为 5 秒，快速踢出掉线玩家
	pingPeriod     = 3 * time.Second // 每 3 秒发送一次 Ping (必须小于 pongWait)
	maxMessageSize = 512
)

// Client 代表一个连接到房间的 WebSocket 客户端
type Client struct {
	ID       string // 对应 UserID
	Username string
	Conn     *websocket.Conn
	Room     *Room
	Send     chan []byte // 缓冲通道，用于向客户端发送消息
	closed   bool        // 标记 Send 通道是否已关闭
}

func (c *Client) handleAction(inMsg ClientMessage) {
	// 拦截通用消息
	switch inMsg.Action {
	case ActionJoin:
		c.Room.Broadcast(c, ActionJoin, "")
	case ActionChat:
		// 未来扩展空间：
		// 1. 敏感词过滤：可以在这里对 inMsg.Content 进行检测和替换
		// 2. 频率限制：检查 c.LastChatTime，防止刷屏
		// 3. 聊天命令：解析以 "/" 开头的字符串，执行特定指令（如 /help, /kick）
		c.Room.Broadcast(c, ActionChat, inMsg.Content)
	default:
		// 如果不是通用消息，交给游戏引擎处理
		if c.Room.Engine != nil {
			c.Room.Engine.HandleMessage(c, inMsg)
		}
	}
}

// readPump 从 WebSocket 连接中读取消息 (内部方法)
func (c *Client) readPump() {
	log.Printf("客户端 [%s] ReadPump 开始运行\n", c.ID)
	defer func() {
		log.Printf("客户端 [%s] ReadPump 退出\n", c.ID)
		c.Room.Broadcast(c, ActionLeave, "") // 广播离开消息
		c.Room.RemoveClient(c)
		c.Conn.Close()
	}()

	// 1. 限制接收消息的最大尺寸，防止恶意客户端发送超大消息导致内存耗尽 (OOM)
	c.Conn.SetReadLimit(maxMessageSize)

	// 2. 设置初始的读取超时时间。如果在这个时间内没有收到客户端的任何消息（包括业务消息或心跳 Pong），
	// 底层的 ReadMessage() 就会返回超时错误，从而断开这个死连接。
	c.Conn.SetReadDeadline(time.Now().Add(pongWait))

	// 3. 注册 Pong 消息处理器。
	// 当后端通过 WritePump 发送 Ping 消息后，如果客户端正常存活，会自动回复一个 Pong 消息。
	// 收到 Pong 消息时，说明客户端还活着，我们就把读取超时时间往后顺延一个 pongWait 周期。
	c.Conn.SetPongHandler(func(string) error {
		return c.Conn.SetReadDeadline(time.Now().Add(pongWait))
	})

	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("客户端 [%s] ReadMessage 异常错误: %v\n", c.ID, err)
			} else {
				log.Printf("客户端 [%s] ReadMessage 退出 (可能为正常断开或超时): %v\n", c.ID, err)
			}
			break
		}

		// 收到任何消息（不仅是 Pong），都说明客户端存活，刷新读取超时时间
		c.Conn.SetReadDeadline(time.Now().Add(pongWait))

		var inMsg ClientMessage
		if err := json.Unmarshal(message, &inMsg); err != nil {
			log.Printf("客户端 [%s] 收到无效 JSON: %v\n", c.ID, string(message))
			continue
		}

		log.Printf("客户端 [%s] 收到消息: Action=%s\n", c.ID, inMsg.Action)
		c.handleAction(inMsg)
	}
}

// writePump 将消息写入 WebSocket 连接 (内部方法)
func (c *Client) writePump() {
	log.Printf("客户端 [%s] WritePump 开始运行\n", c.ID)
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		log.Printf("客户端 [%s] WritePump 退出\n", c.ID)
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			// 每次执行写操作前，都必须刷新 WriteDeadline。
			// 如果不刷新，它会一直沿用上一次设置的绝对时间，导致后续的写入全部报 timeout 错误。
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The hub closed the channel.
				log.Printf("客户端 [%s] send 通道已关闭，准备断开连接\n", c.ID)
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				log.Printf("客户端 [%s] 获取 Writer 失败: %v\n", c.ID, err)
				return
			}
			w.Write(message)

			if err := w.Close(); err != nil {
				log.Printf("客户端 [%s] 关闭 Writer 失败: %v\n", c.ID, err)
				return
			}

		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			log.Printf("客户端 [%s] 发送 Ping\n", c.ID)
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				log.Printf("客户端 [%s] 发送 Ping 失败: %v\n", c.ID, err)
				return
			}
		}
	}
}
