package wscore

import (
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// TODO(Framework Upgrade):
// 目前为了简单，将超时和限制参数写死为全局常量。
// 建议将这些参数抽取为 Config 结构体并挂载到 Hub 级别，
// 以保证同一业务场景下规则一致，并减少 Client 独立存储配置的内存开销。
const (
	writeWait      = 10 * time.Second
	pongWait       = 5 * time.Second
	pingPeriod     = 3 * time.Second
	maxMessageSize = 512
)

// Client 包装底层 WebSocket 连接，负责处理读写循环和心跳
type Client struct {
	id      string          // 客户端唯一标识（如 UserID）
	conn    *websocket.Conn // 底层 WebSocket 连接
	hub     *Hub            // 所属的事件循环管理器
	handler MessageHandler  // 业务逻辑处理器

	// 待发送消息的缓冲通道。
	// 用于解耦业务发送和网络 I/O：外部通过 SendMessage() 写入，WritePump 异步读取并发送。
	// 这样可以防止某个网络慢的客户端阻塞全局的消息广播。
	send chan []byte

	// 优雅关闭机制
	closeCh   chan struct{} // 保证并发错误（如心跳超时、读写错误）时，向 Hub 的注销请求只发送一次，防止死锁。
	closeOnce sync.Once     // Hub 确认注销后关闭的信号通道，WritePump 收到信号后退出循环并断开底层连接。
}

func (c *Client) GetID() string {
	return c.id
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

// Close 主动断开连接
//
// 完整的关闭流程如下：
//  1. 调用 c.Close() 向 Hub 发送注销信号
//  2. Hub 收到信号后，从 clients 映射中移除该客户端，并关闭 c.closeCh 通道
//  3. WritePump 监听到 c.closeCh 被关闭，向客户端发送 WebSocket Close 帧
//  4. WritePump 退出循环，触发 defer 关闭底层的 TCP 连接
//  5. 底层连接关闭导致 ReadPump 中的 ReadMessage 阻塞返回错误
//  6. ReadPump 退出循环，触发 defer，整个客户端生命周期安全结束
func (c *Client) Close() {
	c.closeOnce.Do(func() {
		// 委托给 Hub 处理，优先尝试非阻塞发送，失败则使用 goroutine 防止死锁
		select {
		case c.hub.unregister <- c:
		case <-c.hub.destroyCh:
		default:
			go func() {
				select {
				case c.hub.unregister <- c:
				case <-c.hub.destroyCh:
				}
			}()
		}
	})
}

// NewClient 创建 WebSocket 客户端实例
func NewClient(id string, conn *websocket.Conn, hub *Hub, handler MessageHandler) *Client {
	if handler == nil {
		panic("wscore: handler cannot be nil")
	}

	return &Client{
		id:      id,
		conn:    conn,
		hub:     hub,
		handler: handler,
		send:    make(chan []byte, 256),
		closeCh: make(chan struct{}),
	}
}

// ReadPump 从 WebSocket 连接中读取消息，必须运行在独立的 Goroutine 中
func (c *Client) ReadPump() {
	defer func() {
		if err := recover(); err != nil {
			log.Printf("Client [%s] ReadPump panic recovered: %v", c.id, err)
		}

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

		// 将消息发送给 Hub 处理
		// ReadPump 是每个连接一个独立的 Goroutine，如果直接调用 c.handler.OnMessage，
		// 多个客户端同时发消息会导致业务层的 OnMessage 并发执行，引发数据竞争。
		// 发给 Hub 的 clientMessages 通道，由 Hub.Run 进行统一串行分发，从而解放业务层。
		select {
		case c.hub.clientMessages <- clientMessage{client: c, message: message}:
		case <-c.hub.destroyCh:
			return // Hub 已停止，退出
		}
	}
}

// WritePump 将消息写入 WebSocket 连接，必须运行在独立的 Goroutine 中
func (c *Client) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		if err := recover(); err != nil {
			log.Printf("Client [%s] WritePump panic recovered: %v", c.id, err)
		}

		c.Close() // 确保即使是 Write 异常退出，也能通知 Hub 清理资源
		c.conn.Close()
		ticker.Stop()
	}()

	for {
		select {
		case message := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// 批量排空队列中积压的消息
			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-c.closeCh:
			// Hub 通知连接关闭
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			c.conn.WriteMessage(websocket.CloseMessage, []byte{})
			return

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
