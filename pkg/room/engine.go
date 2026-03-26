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
	// 返回值表示处理完毕后，是否需要房间广播更新后的游戏状态
	HandleMessage(playerID string, action string, content string) bool

	// GetState 获取当前游戏状态快照，用于断线重连或新玩家加入时恢复状态
	GetState(playerID string) any
}
