package core

// Messenger 消息发送接口，由 Handler 层实现并注入给 GameEngine，用于向客户端发送消息
type Messenger interface {
	// Broadcast 广播给游戏内的所有玩家
	Broadcast(msgType string, reason string, payload any)

	// SendTo 私发给游戏内的特定玩家
	SendTo(userID string, msgType string, reason string, payload any)
}
