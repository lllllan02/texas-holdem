package wscore

import (
	"log"
	"sync"
)

// Hub 是一个通用的 WebSocket 连接管理器，同时也是房间的“单线程事件循环(Event Loop)”核心。
// 它维护了一组活跃的 Client，并处理注册、注销、广播消息以及外部异步任务。
//
// 【架构设计说明：Actor 模型】
// 为了极大降低业务层（如 Room 和 GameEngine）处理并发的难度，Hub 采用了类似 Actor 的单线程模型：
// 1. 所有的网络事件（连接、断开、收到消息）都会被转化为内部消息，投入 Hub 的 channel 中。
// 2. Hub.Run() 在一个单一的 Goroutine 中串行消费这些消息，并同步调用业务层的 handler（OnConnect/OnDisconnect/OnMessage）。
// 3. 业务层在处理这些 handler 时，天然是线程安全的，不需要加任何锁（sync.Mutex/RWMutex）。
//
// 【下游业务层（Room/Engine）开发须知】
//   - 只要是由客户端网络消息触发的逻辑，直接写即可，无需考虑并发。
//   - 如果业务层有独立的后台 Goroutine（例如：time.AfterFunc 定时器、监听其他 channel 的循环），
//     并且这些 Goroutine 需要读取或修改房间/游戏状态，**绝对不能直接操作**，必须通过 `Hub.Execute()`
//     将操作打包成闭包函数投递回 Hub，让 Hub 在主循环中串行执行，以防数据竞争。
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

	// 接收客户端消息的通道，用于将多 Goroutine 并发的网络读取，转化为单 Goroutine 的串行处理
	incoming chan *clientMessage

	// 任务通道，允许外部（如定时器、其他系统）安全地提交任务到 Hub 的事件循环中串行执行
	tasks chan func()

	// 确保 Stop 逻辑只执行一次
	stopOnce sync.Once
}

// clientMessage 包装客户端发来的消息
type clientMessage struct {
	client  *Client
	message []byte
}

// NewHub 创建一个新的 Hub 实例
func NewHub() *Hub {
	return &Hub{
		broadcast:  make(chan []byte, 64),
		register:   make(chan *Client, 8),
		unregister: make(chan *Client, 8),
		clients:    make(map[*Client]bool),
		destroy:    make(chan struct{}),
		incoming:   make(chan *clientMessage, 64),
		tasks:      make(chan func(), 64),
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

			// 通知业务层：有新连接（同步调用，保证与消息处理串行）
			if c.handler != nil {
				c.handler.OnConnect(c)
			}

		case c := <-h.unregister:
			h.removeClient(c)

		case msg := <-h.incoming:
			// 处理客户端发来的消息。
			// 此时已经在 Hub.Run 的单线程循环中，所以同步调用 handler 是绝对安全的。
			if msg.client.handler != nil {
				msg.client.handler.OnMessage(msg.client, msg.message)
			}

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

		case task := <-h.tasks:
			// 执行外部提交的异步任务（如定时器回调）。
			// 使用闭包和 recover 包装，防止单个任务的 panic 导致整个房间的 Hub 崩溃。
			func() {
				defer func() {
					if err := recover(); err != nil {
						log.Printf("Hub: task panic recovered: %v", err)
					}
				}()
				task()
			}()

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

// Execute 提交一个任务到 Hub 的单线程事件循环中执行。
//
// 为什么需要这个方法？
// 虽然客户端发来的网络消息已经由 Hub 保证了串行调用 handler，
// 但在游戏业务中，经常会有“非网络触发”的异步事件，例如：
// 1. 定时器到期（如：发牌倒计时结束、房间空闲超时回收）。
// 2. 跨 Goroutine 的状态推送（如：Engine 状态变化通知 Room 广播）。
// 这些事件通常发生在 Go 内部独立的 Goroutine 中，如果直接在这些 Goroutine 里
// 修改房间或桌子状态，就会和 Hub 主循环发生并发冲突（Data Race）。
//
// 因此，下游业务必须使用 Execute 将这些异步操作“发回” Hub 的主循环中排队执行，
// 从而维持整个房间状态机绝对的单线程安全。
func (h *Hub) Execute(task func()) {
	select {
	case h.tasks <- task:
	case <-h.destroy:
		// 如果 Hub 已经停止，直接丢弃任务
	}
}

// Stop 停止 Hub 并断开所有连接。
func (h *Hub) Stop() {
	h.stopOnce.Do(func() {
		close(h.destroy)
	})
}

// removeClient 从 Hub 中移除客户端并清理资源。
// 注意：该方法只能在 Hub.Run 的 Goroutine 中调用，以保证并发安全。
func (h *Hub) removeClient(c *Client) {
	if _, ok := h.clients[c]; !ok {
		return // 避免重复关闭
	}

	delete(h.clients, c)
	close(c.closeCh) // 通知 WritePump 退出
	log.Printf("Hub: Client [%s] unregistered. Total: %d", c.id, len(h.clients))

	// 通知业务层：连接断开（同步调用，保证与消息处理串行）
	if c.handler != nil {
		c.handler.OnDisconnect(c)
	}
}
