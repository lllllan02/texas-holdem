package domain

// Table 德州扑克牌桌 (GameEngine 的具体实现)
// 负责管理跨局的持久化状态，包括座位分配、庄家位置的流转以及对局历史。
type Table struct {
	// --- 房间规则/配置配置 (通常在游戏过程中不变) ---
	MaxPlayers    int `json:"max_players"`    // 最大座位数 (如 6 或 9)
	SmallBlind    int `json:"small_blind"`    // 小盲注金额
	BigBlind      int `json:"big_blind"`      // 大盲注金额
	InitialChips  int `json:"initial_chips"`  // 初始筹码（玩家入座后统一分配的数量）
	ActionTimeout int `json:"action_timeout"` // 玩家行动超时时间(秒)

	// --- 牌桌运行时状态 (随着游戏进行不断变化) ---
	Seats       []*Seat         `json:"seats"`        // 座位数组，长度等于 MaxPlayers
	Claimed     map[string]bool `json:"claimed"`      // 标记玩家是否已经领取过初始筹码 (Key: UserID)
	ButtonSeat  int             `json:"button_seat"`  // 当前庄家 (Dealer/Button) 所在的座位号
	HandCount   int             `json:"hand_count"`   // 当前桌子已经进行了多少局游戏
	CurrentHand *Hand              `json:"current_hand"` // 当前正在进行的单局游戏实例（如果不在游戏中则为 nil）
	Histories   []*ShowdownSummary `json:"histories"`    // 历史对局记录列表，用于战绩回放

	// --- 内部依赖 ---
	messenger Messenger `json:"-"` // 注入的消息发送器 (在 OnInit 时传入)
}

// OnInit 引擎初始化时调用
func (t *Table) OnInit(messenger Messenger, options interface{}) error {
	t.messenger = messenger
	// 解析 options 并初始化 MaxPlayers, SmallBlind 等配置...
	return nil
}

// OnDestroy 房间被销毁时调用
func (t *Table) OnDestroy() {
	// 清理定时器等资源
}

// OnPlayerJoin 玩家加入房间时调用
func (t *Table) OnPlayerJoin(userID string) {
	// 可以广播旁观者加入的消息
}

// OnPlayerLeave 玩家离开/掉线时调用
func (t *Table) OnPlayerLeave(userID string) {
	// 如果玩家在座位上，可能需要触发自动弃牌或站起逻辑
}
