package texas

import (
	"encoding/json"
	"fmt"

	"github.com/lllllan02/texas-holdem/pkg/core"
)

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
	IsPaused    bool               // 游戏是否处于暂停状态
}

// NewTable 创建一个新的德州扑克牌桌实例
// 注意：此时不解析业务配置，只做最基础的内存结构初始化
func NewTable() *Table {
	return &Table{
		Claimed:    make(map[string]bool),
		ButtonSeat: -1, // 游戏未开始时没有庄家，用 -1 表示
		HandCount:  0,
		Histories:  make([]*ShowdownSummary, 0),
	}
}

// 确保 Table 实现了 core.GameEngine 接口
var _ core.GameEngine = (*Table)(nil)

// GameType 获取当前游戏引擎的类型
func (t *Table) GameType() string {
	return "texas"
}

// OnInit 引擎初始化时调用，注入消息发送器和游戏配置
func (t *Table) OnInit(messenger core.Messenger, options []byte) error {
	type TableOptions struct {
		MaxPlayers    int `json:"max_players"`    // 最大座位数 (通常为 2, 6, 9)
		SmallBlind    int `json:"small_blind"`    // 小盲注金额
		BigBlind      int `json:"big_blind"`      // 大盲注金额
		InitialChips  int `json:"initial_chips"`  // 初始筹码量
		ActionTimeout int `json:"action_timeout"` // 玩家行动超时时间(秒)
	}

	t.messenger = messenger

	// 1. 解析配置
	var opts TableOptions
	if len(options) > 0 {
		if err := json.Unmarshal(options, &opts); err != nil {
			return fmt.Errorf("failed to parse table options: %w", err)
		}
	}

	// 2. 设置默认值
	if opts.MaxPlayers <= 0 {
		opts.MaxPlayers = 9
	}
	if opts.BigBlind <= 0 {
		opts.BigBlind = 20
	}
	if opts.SmallBlind <= 0 {
		opts.SmallBlind = opts.BigBlind / 2
	}
	if opts.InitialChips <= 0 {
		opts.InitialChips = opts.BigBlind * 100 // 默认带入 100 个大盲
	}
	if opts.ActionTimeout <= 0 {
		opts.ActionTimeout = 30 // 默认 30 秒思考时间
	}

	// 3. 应用配置到 Table
	t.MaxPlayers = opts.MaxPlayers
	t.SmallBlind = opts.SmallBlind
	t.BigBlind = opts.BigBlind
	t.InitialChips = opts.InitialChips
	t.ActionTimeout = opts.ActionTimeout

	// 4. 根据 MaxPlayers 初始化座位数组
	t.Seats = make([]*Seat, t.MaxPlayers)
	for i := 0; i < t.MaxPlayers; i++ {
		t.Seats[i] = &Seat{
			SeatNumber: i,
			State:      SeatEmpty,
			Player:     nil,
		}
	}

	return nil
}

// OnDestroy 引擎被销毁时调用，用于清理资源
func (t *Table) OnDestroy() {
	// TODO: 停止当前可能正在运行的倒计时定时器 (Action Timer)
	// TODO: 停止可能正在运行的延迟结算定时器 (Showdown Timer)
}

// OnPlayerJoin 玩家加入游戏时调用
// 注意：这只是建立连接进入房间，并不代表落座。玩家默认是旁观者。
func (t *Table) OnPlayerJoin(userID string) {
	// 1. 检查该玩家是否之前在座位上（断线重连）
	if seat := t.getSeatByUserID(userID); seat != nil {
		// 恢复在线状态
		seat.Player.IsOffline = false
		// 广播状态更新，通知其他人该玩家已重连
		t.messenger.Broadcast(MsgTypeStateUpdate, "player_reconnected", t.BuildPublicSnapshot())
	}

	// 2. 发送当前的全局快照给该玩家，以便前端恢复画面
	snap := t.BuildPersonalSnapshot(userID)
	t.messenger.SendTo(userID, MsgTypeStateUpdate, "player_joined", snap)
}

// OnPlayerLeave 玩家离开游戏/掉线时调用
func (t *Table) OnPlayerLeave(userID string) {
	// 1. 查找该玩家是否在座位上
	seat := t.getSeatByUserID(userID)

	// 如果只是旁观者离开，不需要处理
	if seat == nil {
		return
	}

	// 2. 如果玩家在座位上，处理断线逻辑
	// 注意：如果游戏正在进行中，我们不需要在这里做任何特殊处理（比如自动 Fold）。
	// 因为轮到他行动时，系统自然会挂起一个倒计时，倒计时结束他没操作，状态机也会自动帮他 Check/Fold。
	// 无论游戏是否开始，我们都只将他标记为“已断线/托管”，保留他的座位和筹码。
	// 真正的清座逻辑（踢人）应该由另一个专门的定时器或房主手动触发。
	seat.Player.IsOffline = true

	// 3. 广播玩家离开的消息
	t.messenger.Broadcast(MsgTypeStateUpdate, "player_left", t.BuildPublicSnapshot())
}

// Pause 暂停游戏引擎（通常由房主触发）
// 注意：此方法仅负责挂起引擎内部的状态和定时器。
// TODO 调用方 (Handler) 在调用成功后，必须自行负责向客户端广播游戏暂停的消息。
func (t *Table) Pause() error {
	if t.IsPaused {
		return fmt.Errorf("game is already paused")
	}

	t.IsPaused = true
	// TODO: 如果有正在运行的倒计时（如玩家思考时间），需要将其挂起/冻结

	return nil
}

// Resume 恢复游戏引擎
// 注意：此方法仅负责恢复引擎内部的状态和定时器。
// TODO 调用方 (Handler) 在调用成功后，必须自行负责向客户端广播游戏恢复的消息（或发送全量快照）。
func (t *Table) Resume() error {
	if !t.IsPaused {
		return fmt.Errorf("game is not paused")
	}

	t.IsPaused = false
	// TODO: 恢复之前被挂起的倒计时

	return nil
}

// HandleMessage 处理游戏内的具体动作
func (t *Table) HandleMessage(userID string, msgType string, payload []byte) error {
	switch msgType {
	case MsgTypeSitDown:
		// TODO: 处理落座
		// t.checkAndAutoStart() // 落座后如果满员且都准备了，可能触发开局
	case MsgTypeStandUp:
		// TODO: 处理站起
	case MsgTypeReady:
		// TODO: 处理准备
		// 准备后可能触发开局
		t.checkAndAutoStart()
	case MsgTypeCancel:
		// TODO: 处理取消准备
	case MsgTypeAction:
		// TODO: 解析 payload 为 ClientActionPayload，并调用 processPlayerAction
	default:
		// 忽略不认识的消息
	}
	return nil
}

// getSeatByUserID 根据 UserID 查找玩家所在的座位
// 如果玩家不在座位上（比如是旁观者），返回 nil
func (t *Table) getSeatByUserID(userID string) *Seat {
	for _, s := range t.Seats {
		if s.State == SeatOccupied && s.Player != nil && s.Player.User.ID == userID {
			return s
		}
	}
	return nil
}

// checkAndAutoStart 检查是否满足开局条件，如果满足则自动开始游戏
func (t *Table) checkAndAutoStart() {
	// 1. 检查是否有正在进行的游戏
	if t.CurrentHand != nil {
		return
	}

	// 2. 检查是否满员，且所有人都已准备
	readyCount := 0
	for _, seat := range t.Seats {
		if seat.State == SeatEmpty || seat.Player == nil {
			return // 未满员，不开始
		}
		if seat.Player.State != PlayerStateReady {
			return // 有人未准备，不开始
		}
		readyCount++
	}

	// 3. 理论上满员检查已经保证了人数，但为了严谨还是校验一下
	if readyCount < 2 {
		return
	}

	// 4. 所有条件满足，自动触发发牌流程
	// TODO: 可以在这里加一个 3 秒的倒计时，然后再 startNewHand()
	t.startNewHand()
}
