package core

// RoomContext 暴露给游戏引擎的房间操作接口
// 通过这个接口，游戏引擎只能进行广播等受限操作，无法直接修改 Room 的底层状态
type RoomContext interface {
	GetID() string
	GetHostID() string

	// 消息发送接口
	Broadcast(sender *Client, action string, content string)
	SendTo(client *Client, action string, content string)
}

// GameEngine 抽象游戏引擎接口，任何游戏都可以接入这个房间系统
type GameEngine interface {
	// OnInit 房间初始化时调用，传入受限的房间上下文
	OnInit(roomCtx RoomContext)
	// OnDestroy 房间被销毁时调用，用于清理游戏引擎内部的资源（如定时器、Goroutine）
	OnDestroy()
	// OnPlayerJoin 玩家加入房间时调用
	OnPlayerJoin(client *Client)
	// OnPlayerLeave 玩家离开/掉线时调用
	OnPlayerLeave(client *Client)
	// HandleMessage 处理客户端发来的游戏特定消息
	HandleMessage(client *Client, msg ClientMessage)
	// GetState 获取当前游戏状态快照，用于断线重连或新玩家加入时恢复状态
	GetState(client *Client) any
}
