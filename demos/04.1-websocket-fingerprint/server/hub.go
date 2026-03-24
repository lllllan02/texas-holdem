package main

import "encoding/json"

// Hub 维护了所有活动的客户端，并向它们广播消息
type Hub struct {
	// 注册的客户端 (连接对象 -> 是否存在)
	Clients map[*Client]bool

	// 用户ID到客户端的映射，用于单播和顶号逻辑
	Users map[string]*Client

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
		Users:      make(map[string]*Client),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.Register:
			// 顶号逻辑：如果该 UserID 已经有连接了，踢掉旧的
			h.kickOutOldClient(client.UserID)

			// 注册新连接
			h.Clients[client] = true
			h.Users[client.UserID] = client

		case client := <-h.Unregister:
			if _, ok := h.Clients[client]; ok {
				delete(h.Clients, client)
				// 只有当当前 Users 里的连接还是这个要注销的连接时，才删除映射
				// 防止顶号时，旧连接的注销把新连接的映射给删了
				if h.Users[client.UserID] == client {
					delete(h.Users, client.UserID)
				}
				
				if !client.closed {
					client.closed = true
					close(client.Send)
				}
			}
		case message := <-h.Broadcast:
			for client := range h.Clients {
				if client.closed {
					continue
				}
				select {
				case client.Send <- message:
				default:
					// 如果发送缓冲满了，说明客户端卡死或断开，强制注销
					if !client.closed {
						client.closed = true
						close(client.Send)
					}
					delete(h.Clients, client)
				}
			}
		}
	}
}

// kickOutOldClient 踢掉已经存在的旧连接（顶号逻辑）
func (h *Hub) kickOutOldClient(userID string) {
	if oldClient, ok := h.Users[userID]; ok {
		if oldClient.closed {
			return
		}

		// 发送被顶号的私有消息
		kickMsg := ServerMessage{
			Action:  ActionKick,
			Content: "您的账号在其他地方登录，您已被强制下线。",
		}
		kickBytes, _ := json.Marshal(kickMsg)

		// 尝试发送踢人消息，不阻塞
		select {
		case oldClient.Send <- kickBytes:
		default:
		}

		oldClient.closed = true
		close(oldClient.Send)
		delete(h.Clients, oldClient)
		// 注意：不需要 delete(h.Users, userID)，因为外面马上会用新连接覆盖它
	}
}
