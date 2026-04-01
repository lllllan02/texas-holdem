package core

// Message 是客户端与服务端通信的统一消息外壳 (Envelope)
// 无论是 Handler 层还是 Game 层，都应遵守此结构进行消息的序列化和反序列化
type Message struct {
	// 消息类型，用于路由。例如: "room.join", "texas.bet", "error"
	Type string `json:"type"`

	// 触发此消息的原因/动作，通常用于广播时告诉客户端“为什么状态变了”
	// 例如 Type="texas.sync", Reason="player_bet"
	// 可选字段，如果没有原因可以为空
	Reason string `json:"reason,omitempty"`

	// 具体的业务数据，通常是一个 JSON 对象
	// 接收时：由 Handler 提取为 json.RawMessage 或 []byte 传给 Engine
	// 发送时：Engine 传入具体的 struct，由 Handler 序列化
	Payload any `json:"payload,omitempty"`
}

// ============================================================================
// 系统级消息类型 (System MsgType)
// ============================================================================
const (
	MsgTypeError = "sys.error" // 错误提示
)

// ErrorPayload 统一的错误消息体结构
type ErrorPayload struct {
	Error string `json:"error"`
}
