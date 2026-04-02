/*
Hub 架构设计说明（基于 Actor 模型）

Hub 是整个 WebSocket 框架的“单线程事件循环(Event Loop)”核心，旨在极大降低上层业务（如 Room/GameEngine）处理并发的难度。

核心机制：
1. 所有的网络事件（连接、断开、收到消息）都会被转化为内部消息，投入 Hub 的 channel 中。
2. Hub.Run() 在单一 Goroutine 中串行消费这些消息，并同步调用业务层的 handler。
3. 业务层在处理 OnConnect/OnDisconnect/OnMessage 时，天然是线程安全的，无需加锁。

下游业务开发须知：
  - 客户端网络消息触发的逻辑，直接编写即可，无需考虑并发。
  - 独立的后台 Goroutine（如定时器、异步回调）若需读写房间/游戏状态，绝对不能直接操作！
    必须通过 `Hub.Execute()` 将操作打包成闭包投递回 Hub，在主循环中串行执行，以防数据竞争。
*/
package wscore

import (
	"log"
	"sync"
)

// Hub 是 WebSocket 连接管理器和事件循环核心
type Hub struct {
	clients map[*Client]bool // 活跃的客户端列表（仅在主循环中读写，无需加锁）
	clientsByID map[string]*Client // 按 clientID 索引的活跃连接（用于顶号）

	// 连接管理通道。
	// 将并发的连接建立与断开事件转化为串行消息，交由主循环统一处理。
	register   chan *Client
	unregister chan *Client

	// 广播通道。
	// 外部通过 Broadcast() 写入，主循环消费后分发给所有活跃客户端，避免阻塞调用方。
	broadcast chan []byte

	// 接收客户端消息的通道。
	// 核心设计：将各个 Client 并发读取到的网络消息统一投递到这里，
	// 让主循环串行调用业务 handler，从而实现业务层的无锁化处理。
	clientMessages chan clientMessage

	// 外部任务投递通道。
	// 允许外部 Goroutine（如定时器、API 请求）将操作打包成闭包提交到主循环执行，
	// 防止外部直接修改业务状态导致数据竞争。
	externalTasks chan func()

	destroyCh   chan struct{} // 销毁通道，关闭后触发主循环退出并清理所有连接
	destroyOnce sync.Once     // 确保 destroy 逻辑只执行一次，防止多次关闭 destroy 引发 panic
}

// clientMessage 包装客户端发来的消息
type clientMessage struct {
	client  *Client
	message []byte
}

func NewHub() *Hub {
	return &Hub{
		clients:        make(map[*Client]bool),
		clientsByID:    make(map[string]*Client),
		register:       make(chan *Client, 8),
		unregister:     make(chan *Client, 8),
		broadcast:      make(chan []byte, 64),
		clientMessages: make(chan clientMessage, 64),
		externalTasks:  make(chan func(), 64),
		destroyCh:      make(chan struct{}),
	}
}

// safeCall 包装业务 handler 调用，防止 panic 导致 Hub 崩溃
func safeCall(fn func()) {
	defer func() {
		if err := recover(); err != nil {
			log.Printf("Hub: handler panic recovered: %v", err)
		}
	}()

	fn()
}

// Run 启动 Hub 的事件循环，必须运行在独立的 Goroutine 中
func (h *Hub) Run() {
	defer func() {
		for c := range h.clients {
			h.removeClient(c)
		}
		log.Println("Hub: Destroyed")
	}()

	for {
		select {
		case c := <-h.register:
			// 顶号：同一个 clientID 只允许一个活跃连接
			if c.id != "" {
				if old := h.clientsByID[c.id]; old != nil && old != c {
					// 给旧连接一个明确的关闭原因（前端可据此提示“被顶下线”）
					old.CloseWithReason(4001, "该账号已在其他设备登录")
					h.removeClient(old)
				}
				h.clientsByID[c.id] = c
			}

			h.clients[c] = true
			log.Printf("Hub: Client [%s] registered. Total: %d", c.id, len(h.clients))

			// 通知业务层：有新连接（同步调用，保证与消息处理串行）
			safeCall(func() { c.handler.OnConnect(c) })

		case c := <-h.unregister:
			h.removeClient(c)

		case msg := <-h.clientMessages:
			// 处理客户端发来的消息
			// 此时已经在 Hub.Run 的单线程循环中，所以同步调用 handler 是绝对安全的
			safeCall(func() { msg.client.handler.OnMessage(msg.client, msg.message) })

		case message := <-h.broadcast:
			for c := range h.clients {
				select {
				case c.send <- message:
				default:
					// 如果发送通道已满或阻塞，说明客户端卡死，强制关闭
					h.removeClient(c)
				}
			}

		case task := <-h.externalTasks:
			// 执行外部提交的异步任务（如定时器回调）
			// 使用 safeCall 包装，防止单个任务的 panic 导致整个房间的 Hub 崩溃
			safeCall(task)

		case <-h.destroyCh:
			return
		}
	}
}

// Register 将客户端注册到 Hub
func (h *Hub) Register(client *Client) {
	select {
	case h.register <- client:
	case <-h.destroyCh:
	}
}

// Unregister 将客户端从 Hub 注销
func (h *Hub) Unregister(client *Client) {
	select {
	case h.unregister <- client:
	case <-h.destroyCh:
	}
}

// BroadcastMessage 向 Hub 中的所有客户端广播消息
func (h *Hub) BroadcastMessage(message []byte) {
	select {
	case h.broadcast <- message:
	case <-h.destroyCh:
	default:
		log.Println("Hub: broadcast channel full, message dropped")
	}
}

// Execute 提交一个任务到 Hub 的单线程事件循环中执行
//
// 游戏业务中经常会有“非网络触发”的异步事件（如定时器到期、跨 Goroutine 的状态推送）。
// 这些事件通常发生在独立的 Goroutine 中，如果直接修改房间或桌子状态，会和 Hub 主循环发生并发冲突。
// 下游业务必须使用 Execute 将这些异步操作“发回” Hub 的主循环中排队执行，维持整个房间状态机绝对的单线程安全。
func (h *Hub) Execute(task func()) {
	select {
	case h.externalTasks <- task:
	case <-h.destroyCh:
	default:
		log.Println("Hub: external tasks channel full, task dropped")
	}
}

// Stop 停止 Hub 并断开所有连接
func (h *Hub) Stop() {
	h.destroyOnce.Do(func() {
		close(h.destroyCh)
	})
}

// removeClient 从 Hub 中移除客户端并清理资源
// 注意：该方法只能在 Hub.Run 的 Goroutine 中调用，以保证并发安全
func (h *Hub) removeClient(c *Client) {
	if _, ok := h.clients[c]; !ok {
		return
	}

	delete(h.clients, c)
	if c.id != "" && h.clientsByID[c.id] == c {
		delete(h.clientsByID, c.id)
	}
	close(c.closeCh) // 通知 WritePump 退出
	log.Printf("Hub: Client [%s] unregistered. Total: %d", c.id, len(h.clients))

	// 通知业务层：连接断开（同步调用，保证与消息处理串行）
	safeCall(func() { c.handler.OnDisconnect(c) })
}
