package domain

// Messenger 消息发送接口，由 Room 实现并注入给 GameEngine
type Messenger interface {
	// Broadcast 广播给房间内所有人
	Broadcast(msgType string, reason string, payload any)

	// SendTo 私发给房间内的特定玩家
	SendTo(userID string, msgType string, reason string, payload any)
}

// Room 通用房间：最外层的容器，负责网络连接、聊天、用户进出
// 房间本身不关心具体玩什么游戏，它只负责挂载一个具体的 GameEngine
type Room struct {
	ID         string           `json:"id"`          // 房间的唯一内部标识（如 UUID，随机生成的一串字符）
	RoomNumber string           `json:"room_number"` // 供玩家输入的短房间号（如 "888888"），作为进入凭据
	Name       string           `json:"name"`
	IsPaused   bool             `json:"is_paused"` // 房间是否处于全局暂停状态
	Users      map[string]*User `json:"users"`     // 当前在房间内的所有用户（包括玩家和旁观者）
	GameEngine GameEngine       `json:"-"`         // 挂载的具体游戏引擎实例
}
