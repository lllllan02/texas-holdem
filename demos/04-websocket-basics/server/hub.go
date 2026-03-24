package main

// Hub 维护了所有活动的客户端，并向它们广播消息
type Hub struct {
	// 注册的客户端
	Clients map[*Client]bool

	// 广播给客户端的消息
	Broadcast chan []byte

	// 注册请求
	Register chan *Client

	// 注销请求
	Unregister chan *Client
}

func NewHub() *Hub {
	return &Hub{
		Broadcast:  make(chan []byte),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
		Clients:    make(map[*Client]bool),
	}
}

func (h *Hub) Run() {
	for {
		select {

		case client := <-h.Register:
			h.Clients[client] = true

		case client := <-h.Unregister:
			if _, ok := h.Clients[client]; ok {
				delete(h.Clients, client)
				close(client.Send)
			}

		case message := <-h.Broadcast:
			for client := range h.Clients {
				select {
				case client.Send <- message:
				default:
					// 如果发送缓冲满了，说明客户端卡死或断开，强制注销
					close(client.Send)
					delete(h.Clients, client)
				}
			}
		}
	}
}
