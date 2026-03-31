package texas

import "github.com/lllllan02/texas-holdem/pkg/core"

// Table 德州扑克牌桌 (GameEngine 的具体实现)
// 负责管理跨局的持久化状态，包括座位分配、庄家位置的流转以及对局历史。
type Table struct {
	// --- 内部依赖 ---
	messenger core.Messenger // 注入的消息发送器 (在 OnInit 时传入)

	// --- 房间规则/配置配置 (通常在游戏过程中不变) ---
	MaxPlayers    int // 最大座位数 (如 6 或 9)
	SmallBlind    int // 小盲注金额
	BigBlind      int // 大盲注金额
	InitialChips  int // 初始筹码（玩家入座后统一分配的数量）
	ActionTimeout int // 玩家行动超时时间(秒)

	// --- 牌桌运行时状态 (随着游戏进行不断变化) ---
	Seats       []*Seat            // 座位数组，长度等于 MaxPlayers
	Claimed     map[string]bool    // 标记玩家是否已经领取过初始筹码 (Key: UserID)
	ButtonSeat  int                // 当前庄家 (Dealer/Button) 所在的座位号
	HandCount   int                // 当前桌子已经进行了多少局游戏
	CurrentHand *Hand              // 当前正在进行的单局游戏实例（如果不在游戏中则为 nil）
	Histories   []*ShowdownSummary // 历史对局记录列表，用于战绩回放
}

// 确保 Table 实现了 core.GameEngine 接口
var _ core.GameEngine = (*Table)(nil)

// OnInit 引擎初始化时调用，注入消息发送器和游戏配置
func (t *Table) OnInit(messenger core.Messenger, options any) error {
	// TODO: 保存 messenger，解析 options 并初始化牌桌状态
	return nil
}

// OnDestroy 引擎被销毁时调用，用于清理资源
func (t *Table) OnDestroy() {
	// TODO: 清理定时器等资源
}

// OnPlayerJoin 玩家加入游戏时调用
func (t *Table) OnPlayerJoin(userID string) {
	// TODO: 处理玩家加入逻辑（如分配座位或加入旁观）
}

// OnPlayerLeave 玩家离开游戏/掉线时调用
func (t *Table) OnPlayerLeave(userID string) {
	// TODO: 处理玩家离开逻辑（如托管或站起）
}

// GameType 获取当前游戏引擎的类型
func (t *Table) GameType() string {
	return "texas"
}

// StartGame 尝试开始游戏
func (t *Table) StartGame() error {
	// TODO: 检查人数，初始化牌局，开始发牌
	return nil
}

// HandleMessage 处理游戏内的具体动作
func (t *Table) HandleMessage(userID string, msgType string, payload []byte) error {
	// TODO: 解析客户端发来的具体动作（如 bet, fold, check）并更新状态机
	return nil
}

// Pause 暂停游戏引擎
func (t *Table) Pause() error {
	// TODO: 暂停逻辑
	return nil
}

// Resume 恢复游戏引擎
func (t *Table) Resume() error {
	// TODO: 恢复逻辑
	return nil
}

// EndGame 强制结束或正常结束游戏
func (t *Table) EndGame() error {
	// TODO: 结算逻辑
	return nil
}
