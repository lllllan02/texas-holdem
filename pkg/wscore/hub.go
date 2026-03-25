package wscore

import "log"

// Hub 是一个通用的 WebSocket 连接管理器。
// 它维护了一组活跃的 Client，并处理注册、注销和广播消息。
// 它是并发安全的，因为所有的操作都在一个单一的 Goroutine 中进行。
// 泛型 T 代表业务层附加的上下文数据类型。
type Hub[T any] struct {
	// 活跃的客户端列表
	clients map[*Client[T]]bool

	// 注册通道
	register chan *Client[T]

	// 注销通道
	unregister chan *Client[T]

	// 广播通道
	broadcast chan []byte

	// 销毁通道
	destroy chan struct{}
}

// NewHub 创建一个新的 Hub 实例
func NewHub[T any]() *Hub[T] {
	return &Hub[T]{
		broadcast:  make(chan []byte),
		register:   make(chan *Client[T]),
		unregister: make(chan *Client[T]),
		clients:    make(map[*Client[T]]bool),
		destroy:    make(chan struct{}),
	}
}

// Run 启动 Hub 的事件循环。
// 它必须运行在一个独立的 Goroutine 中。
func (h *Hub[T]) Run() {
	for {
		select {
		case c := <-h.register:
			h.clients[c] = true
			log.Printf("Hub: Client [%s] registered. Total: %d", c.id, len(h.clients))

			// 通知业务层：有新连接
			if c.handler != nil {
				go c.handler.OnConnect(c)
			}

		case c := <-h.unregister:
			if _, ok := h.clients[c]; ok {
				delete(h.clients, c)
				close(c.send)
				log.Printf("Hub: Client [%s] unregistered. Total: %d", c.id, len(h.clients))

				// 通知业务层：连接断开
				if c.handler != nil {
					go c.handler.OnDisconnect(c)
				}
			}

		case message := <-h.broadcast:
			// 广播消息给所有连接的客户端
			for c := range h.clients {
				select {
				case c.send <- message:
				default:
					// 如果发送通道已满或阻塞，说明客户端卡死，强制关闭
					close(c.send)
					delete(h.clients, c)
					if c.handler != nil {
						c.handler.OnDisconnect(c)
					}
				}
			}
			
		case <-h.destroy:
			// 销毁 Hub，关闭所有连接
			for c := range h.clients {
				close(c.send)
				delete(h.clients, c)
				if c.handler != nil {
					go c.handler.OnDisconnect(c)
				}
			}
			log.Println("Hub: Destroyed")
			return
		}
	}
}

// BroadcastMessage 向 Hub 中的所有客户端广播消息。
func (h *Hub[T]) BroadcastMessage(message []byte) {
	h.broadcast <- message
}

// Stop 停止 Hub 并断开所有连接。
func (h *Hub[T]) Stop() {
	close(h.destroy)
}
