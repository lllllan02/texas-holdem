package wscore

// MessageHandler 定义了处理 WebSocket 消息和生命周期事件的接口
// 业务层（如 Room、ChatServer）需要实现这个接口，并注入到 Client 中
type MessageHandler interface {
	OnConnect(client *Client)
	OnMessage(client *Client, message []byte)
	OnDisconnect(client *Client)
}
