package wscore

import (
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// TODO(Framework Upgrade):
// 目前为了简单，将超时和限制参数写死为全局常量。
// 如果未来要进一步提升框架的通用性，建议将这些参数抽取为 Config 结构体，
// 并挂载到 Hub 级别（例如 `hub.Config.WriteWait`），而不是 Client 级别。
// 这样既能保证同一个业务场景（Hub）下规则的一致性，又能避免每个 Client 独立存储配置带来的内存开销。
const (
	writeWait      = 10 * time.Second
	pongWait       = 5 * time.Second
	pingPeriod     = 3 * time.Second
	maxMessageSize = 512
)

// Client 包装了一个底层的 WebSocket 连接，负责处理读写循环和心跳。
// 泛型 T 代表业务层附加的上下文数据类型。
type Client struct {
	// 客户端的唯一标识（通常是 UserID）
	id string

	// 底层 WebSocket 连接
	conn *websocket.Conn

	// 所属的 Hub，用于注册和注销
	hub *Hub

	// 发送消息的缓冲通道
	send chan []byte

	// 注入的业务逻辑处理器，当收到消息或断开连接时，调用该处理器的方法
	handler MessageHandler

	// 强类型的业务上下文数据
	context any

	// 确保 Close 逻辑只执行一次
	closeOnce sync.Once
}

// GetID 获取客户端 ID
func (c *Client) GetID() string {
	return c.id
}

// GetContext 获取业务上下文
func (c *Client) GetContext() any {
	return c.context
}

// SendMessage 安全地向客户端发送消息
func (c *Client) SendMessage(message []byte) {
	select {
	case c.send <- message:
	default:
		// 如果发送失败，说明通道阻塞（客户端卡死），尝试通知 Hub 踢掉
		c.Close()
	}
}

// Close 主动断开连接。
//
// 完整的关闭流程如下：
//  1. 调用 c.Close() 向 Hub 发送注销信号 (c.hub.unregister <- c)。
//  2. Hub 收到信号后，从 clients 映射中移除该客户端，并关闭 c.send 通道。
//     - 注意：必须由 Hub 来关闭 c.send，因为 Hub 是唯一向 c.send 发送数据的生产者，
//     - Go 语言原则：永远只在发送端关闭通道，以避免 panic: send on closed channel
//  3. WritePump 监听到 c.send 被关闭，向客户端发送 WebSocket Close 帧。
//  4. WritePump 退出循环，触发 defer 关闭底层的 TCP 连接 (c.conn.Close())。
//  5. 底层连接关闭导致 ReadPump 中的 ReadMessage 阻塞返回错误。
//  6. ReadPump 退出循环，触发 defer，整个客户端生命周期安全结束。
func (c *Client) Close() {
	c.closeOnce.Do(func() {
		// 委托给 Hub 处理，保证并发安全
		select {
		case c.hub.unregister <- c:
		case <-c.hub.destroy:
			// Hub 已停止，无需注销
		}
	})
}

// NewClient 创建一个新的 WebSocket 客户端实例
func NewClient(id string, conn *websocket.Conn, hub *Hub, handler MessageHandler) *Client {
	return &Client{
		id:      id,
		conn:    conn,
		hub:     hub,
		send:    make(chan []byte, 256),
		handler: handler,
		context: *new(any), // 初始化零值，ServeWS 中会覆盖
	}
}

// ReadPump 负责从 WebSocket 连接中读取消息。
// 它必须运行在一个独立的 Goroutine 中。
func (c *Client) ReadPump() {
	defer func() {
		c.Close()
		c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("Client [%s] error: %v", c.id, err)
			}
			break
		}

		// 刷新超时时间
		c.conn.SetReadDeadline(time.Now().Add(pongWait))

		// 将消息抛给业务层处理
		if c.handler != nil {
			c.handler.OnMessage(c, message)
		}
	}
}

// WritePump 负责将消息写入 WebSocket 连接。
// 它必须运行在一个独立的 Goroutine 中。
func (c *Client) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.Close() // 确保即使是 Write 异常退出，也能通知 Hub 清理资源
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// Send 通道被 Hub 关闭，说明需要断开连接
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
