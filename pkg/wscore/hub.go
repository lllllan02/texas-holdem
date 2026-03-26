package wscore

import "log"

// Hub 是一个通用的 WebSocket 连接管理器。
// 它维护了一组活跃的 Client，并处理注册、注销和广播消息。
// 它是并发安全的，因为所有的操作都在一个单一的 Goroutine 中进行。
type Hub struct {
	// 活跃的客户端列表
	clients map[*Client]bool

	// 注册通道
	register chan *Client

	// 注销通道
	unregister chan *Client

	// 广播通道
	broadcast chan []byte

	// 销毁通道
	destroy chan struct{}
}

// NewHub 创建一个新的 Hub 实例
func NewHub() *Hub {
	return &Hub{
		broadcast:  make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[*Client]bool),
		destroy:    make(chan struct{}),
	}
}

// Run 启动 Hub 的事件循环。
// 它必须运行在一个独立的 Goroutine 中。
func (h *Hub) Run() {
	// 退出时确保清理所有连接
	defer func() {
		for c := range h.clients {
			h.removeClient(c)
		}
		log.Println("Hub: Destroyed")
	}()

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
			h.removeClient(c)

		case message := <-h.broadcast:
			// 广播消息给所有连接的客户端
			for c := range h.clients {
				select {
				case c.send <- message:
				default:
					// 如果发送通道已满或阻塞，说明客户端卡死，强制关闭
					h.removeClient(c)
				}
			}

		case <-h.destroy:
			// 收到销毁信号，退出事件循环，触发 defer 清理
			return
		}
	}
}

// BroadcastMessage 向 Hub 中的所有客户端广播消息。
func (h *Hub) BroadcastMessage(message []byte) {
	select {
	case h.broadcast <- message:
	case <-h.destroy:
		// 如果 Hub 已经停止，直接丢弃消息，避免阻塞
	}
}

// Stop 停止 Hub 并断开所有连接。
func (h *Hub) Stop() {
	close(h.destroy)
}

// removeClient 从 Hub 中移除客户端并清理资源。
// 注意：该方法只能在 Hub.Run 的 Goroutine 中调用，以保证并发安全。
func (h *Hub) removeClient(c *Client) {
	if _, ok := h.clients[c]; !ok {
		return // 避免重复关闭
	}

	delete(h.clients, c)
	close(c.send)
	log.Printf("Hub: Client [%s] unregistered. Total: %d", c.id, len(h.clients))
	if c.handler != nil {
		go c.handler.OnDisconnect(c)
	}
}
