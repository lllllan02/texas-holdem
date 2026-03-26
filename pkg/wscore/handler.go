package wscore

// MessageHandler 定义了处理 WebSocket 消息和生命周期事件的接口。
// 业务层（如 Room、ChatServer）需要实现这个接口，并注入到 Client 中。
type MessageHandler interface {
	// 当客户端连接成功并注册到 Hub 后调用
	OnConnect(client *Client)

	// 当收到客户端消息时调用
	OnMessage(client *Client, message []byte)

	// 当客户端断开连接时调用
	OnDisconnect(client *Client)
}
