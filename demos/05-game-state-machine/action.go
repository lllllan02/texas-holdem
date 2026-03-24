package machine

// ActionType 定义了玩家可以执行的操作类型
type ActionType string

const (
	// --- 房间级动作 (Room Actions) ---
	ActionJoin       ActionType = "ROOM.JOIN"       // 加入房间/坐下
	ActionLeave      ActionType = "ROOM.LEAVE"      // 离开房间/站起
	ActionReady      ActionType = "ROOM.READY"      // 准备开始
	ActionStart      ActionType = "ROOM.START"      // 房主手动开始游戏
	ActionDisconnect ActionType = "ROOM.DISCONNECT" // 玩家断开连接 (通常由网关/WebSocket层自动触发)
	ActionReconnect  ActionType = "ROOM.RECONNECT"  // 玩家重新连接

	// --- 局内动作 (In-Game Actions) ---
	ActionFold  ActionType = "GAME.FOLD"  // 弃牌
	ActionCheck ActionType = "GAME.CHECK" // 过牌
	ActionCall  ActionType = "GAME.CALL"  // 跟注
	ActionRaise ActionType = "GAME.RAISE" // 加注
)


// PlayerAction 表示玩家发起的一个动作请求
type PlayerAction struct {
	PlayerID string
	Action   ActionType
	Amount   int // 目标总额 (Target Total Bet)。例如当前最高下注是50，你之前下了20，想Raise到100，这里传100。
}
