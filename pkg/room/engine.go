package room

// GameEngine 抽象游戏引擎接口，任何游戏都可以接入这个房间系统
// TODO: 未来考虑引入 RoomContext 接口，将引擎对 *Room 的强依赖解耦，以便于单元测试和更清晰的边界划分。
//
// 并发控制规范 (Concurrency Contract):
// 1. 引擎的所有方法（OnInit, OnPlayerJoin, HandleMessage 等）都保证在 wscore.Hub 的单线程事件循环中被串行调用。
// 2. 引擎内部实现无需加锁（不需要 sync.Mutex）。
// 3. 如果引擎启动了后台 Goroutine（例如倒计时、异步任务），必须通过 room.GetHub().Execute(...) 将状态修改操作提交回主循环，绝对不能在后台 Goroutine 中直接修改引擎状态或调用 Room 的非并发安全方法。
type GameEngine interface {
	// OnInit 房间初始化时调用
	OnInit(room *Room, param map[string]any) error

	// OnDestroy 房间被销毁时调用，用于清理资源
	OnDestroy()

	// OnPlayerJoin 玩家加入房间时调用（物理连接建立）
	OnPlayerJoin(playerID string)

	// OnPlayerLeave 玩家离开/掉线时调用（物理连接断开）
	OnPlayerLeave(playerID string)

	// HandleMessage 处理客户端发来的游戏特定消息
	HandleMessage(playerID string, action string, content []byte)

	// GetState 获取当前游戏状态快照，用于断线重连或新玩家加入时恢复状态
	GetState(playerID string) any
}
