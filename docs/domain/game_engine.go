package domain

// GameEngine 游戏引擎接口
// 任何想要接入 Room 的游戏（如德扑、UNO）都需要实现此接口
type GameEngine interface {
	// --- 生命周期钩子 (Lifecycle Hooks) ---

	// OnInit 引擎初始化时调用，注入消息发送器和游戏配置
	OnInit(messenger Messenger, options interface{}) error

	// OnDestroy 房间被销毁时调用，用于清理资源（如停止所有定时器）
	OnDestroy()

	// OnPlayerJoin 玩家加入房间时调用（物理连接建立）
	OnPlayerJoin(userID string)

	// OnPlayerLeave 玩家离开/掉线时调用（物理连接断开）
	OnPlayerLeave(userID string)

	// --- 游戏控制与消息处理 ---

	// GameType 获取当前游戏引擎的类型（如 "texas", "uno"）
	GameType() string

	// StartGame 尝试开始游戏（比如检查人数是否满足要求，状态是否都为 ready）
	StartGame() error

	// HandleMessage 处理游戏内的具体动作（如德扑的 bet/fold，UNO 的出牌）
	// msg 通常是一个未解析的 JSON 字节流，由具体的引擎自己去反序列化
	HandleMessage(userID string, msgType string, payload []byte) error

	// Pause 暂停游戏引擎（冻结倒计时、阻止玩家行动）
	Pause() error

	// Resume 恢复游戏引擎
	Resume() error

	// EndGame 强制结束或正常结束游戏
	EndGame() error
}
