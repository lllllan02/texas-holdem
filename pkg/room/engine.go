package room

// GameEngine 抽象游戏引擎接口，任何游戏都可以接入这个房间系统
type GameEngine interface {
	// OnInit 房间初始化时调用
	OnInit(room *Room, param map[string]any)

	// OnDestroy 房间被销毁时调用，用于清理资源
	OnDestroy()

	// OnPlayerJoin 玩家加入房间时调用（物理连接建立）
	OnPlayerJoin(playerID string)

	// OnPlayerLeave 玩家离开/掉线时调用（物理连接断开）
	OnPlayerLeave(playerID string)

	// HandleMessage 处理客户端发来的游戏特定消息
	HandleMessage(playerID string, action string, content string)

	// GetState 获取当前游戏状态快照，用于断线重连或新玩家加入时恢复状态
	GetState(playerID string) any

	// UpdateChannel 返回一个只读 channel，当游戏状态发生变化时，引擎会向该 channel 发送信号
	UpdateChannel() <-chan struct{}
}
