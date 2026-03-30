package domain

// GameEngine 游戏引擎接口
// 任何想要接入 Room 的游戏（如德扑、UNO）都需要实现此接口
type GameEngine interface {
	// GameType 获取当前游戏引擎的类型（如 "texas", "uno"）
	GameType() string

	// StartGame 尝试开始游戏（比如检查人数是否满足要求，状态是否都为 ready）
	StartGame() error

	// HandleMessage 处理游戏内的具体动作（如德扑的 bet/fold，UNO 的出牌）
	// msg 通常是一个解析好的 JSON 载荷
	HandleMessage(userID string, msg any) error

	// Pause 暂停游戏引擎（冻结倒计时、阻止玩家行动）
	Pause() error

	// Resume 恢复游戏引擎
	Resume() error

	// EndGame 强制结束或正常结束游戏
	EndGame() error
}
