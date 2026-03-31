package texas

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/lllllan02/texas-holdem/pkg/core"
	"github.com/lllllan02/texas-holdem/pkg/user"
	"github.com/lllllan02/texas-holdem/pkg/utils"
)

type TableOptions struct {
	MaxPlayers    int `json:"max_players"`    // 最大座位数 (通常为 2, 6, 9)
	SmallBlind    int `json:"small_blind"`    // 小盲注金额
	BigBlind      int `json:"big_blind"`      // 大盲注金额
	InitialChips  int `json:"initial_chips"`  // 初始筹码量
	ActionTimeout int `json:"action_timeout"` // 玩家行动超时时间(秒)
}

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
	Players     map[string]*Player // 记录所有参与过本桌游戏的玩家实体，用于恢复筹码和买入次数 (Key: UserID)
	ButtonSeat  int                // 当前庄家 (Dealer/Button) 所在的座位号
	HandCount   int                // 当前桌子已经进行了多少局游戏
	CurrentHand *Hand              // 当前正在进行的单局游戏实例（如果不在游戏中则为 nil）
	Histories   []*ShowdownSummary // 历史对局记录列表，用于战绩回放
	IsPaused    bool               // 游戏是否处于暂停状态

	// --- 定时器 ---
	countdownTimer *utils.PausableTimer // 开局倒计时定时器
	actionTimer    *utils.PausableTimer // 玩家行动倒计时定时器
	showdownTimer  *utils.PausableTimer // 动画缓冲定时器（用于摊牌结算、提前结束、All-in快进等场景，给前端留出播放动画的时间）
}

// NewTable 创建一个新的德州扑克牌桌实例
// 注意：此时不解析业务配置，只做最基础的内存结构初始化
func NewTable() *Table {
	return &Table{
		Players:        make(map[string]*Player),
		ButtonSeat:     -1, // 游戏未开始时没有庄家，用 -1 表示
		HandCount:      0,
		Histories:      make([]*ShowdownSummary, 0),
		countdownTimer: utils.NewPausableTimer(),
		actionTimer:    utils.NewPausableTimer(),
		showdownTimer:  utils.NewPausableTimer(),
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
			Player:     nil,
		}
	}

	return nil
}

// OnDestroy 引擎被销毁时调用，用于清理资源
func (t *Table) OnDestroy() {
	t.countdownTimer.Stop()
	t.actionTimer.Stop()
	t.showdownTimer.Stop()
}

// OnPlayerJoin 玩家加入游戏时调用
// 注意：这只是建立连接进入房间，并不代表落座。玩家默认是旁观者。
func (t *Table) OnPlayerJoin(u *user.User) {
	// 1. 将玩家信息保存或更新到 Players 映射中
	player, exists := t.Players[u.ID]
	if !exists {
		// 第一次进入房间，创建玩家实体并分配初始筹码
		player = &Player{
			User:       u,
			State:      PlayerStateWaiting,
			Chips:      t.InitialChips, // 进入房间即分配筹码，未落座对其他人不可见
			BuyInCount: 1,              // 初始带入算作第 1 次买入
		}
		t.Players[u.ID] = player
	} else {
		// 玩家之前就在房间里，更新用户信息（可能改了名字或头像）
		player.User = u
	}

	// 2. 如果该玩家之前已经落座（断线重连），恢复其在线状态并广播
	if player.IsOffline {
		player.IsOffline = false
		// 只有当玩家在座位上时，才需要向其他人广播其重连状态（旁观者重连不需要广播）
		if t.getSeatByUserID(u.ID) != nil {
			t.messenger.Broadcast(MsgTypeStateUpdate, "player_reconnected", t.BuildPublicSnapshot())
		}
	}

	// 3. 发送当前的全局快照给该玩家，以便前端恢复画面
	snap := t.BuildPersonalSnapshot(u.ID)
	t.messenger.SendTo(u.ID, MsgTypeStateUpdate, "player_joined", snap)
}

// OnPlayerLeave 玩家离开游戏/掉线时调用
func (t *Table) OnPlayerLeave(userID string) {
	// 1. 从 Players 映射中找到该玩家
	player, exists := t.Players[userID]
	if !exists {
		return
	}

	// 2. 标记为断线状态
	player.IsOffline = true

	// 3. 如果该玩家在座位上，广播其离开（断线）的消息
	// 旁观者离开不需要广播，以免打扰正在打牌的人
	if t.getSeatByUserID(userID) != nil {
		t.messenger.Broadcast(MsgTypeStateUpdate, "player_left", t.BuildPublicSnapshot())
	}
}

// Pause 暂停游戏引擎（通常由房主触发）
// 注意：此方法仅负责挂起引擎内部的状态和定时器。
// TODO 调用方 (Handler) 在调用成功后，必须自行负责向客户端广播游戏暂停的消息。
func (t *Table) Pause() error {
	if t.IsPaused {
		return fmt.Errorf("game is already paused")
	}

	t.IsPaused = true

	// 挂起玩家思考倒计时
	t.actionTimer.Pause()

	// 挂起开局倒计时
	t.countdownTimer.Stop()

	// 注意：showdownTimer（动画缓冲）通常不需要挂起，因为它是为了给前端播动画，
	// 暂停游戏不应该打断正在播放的结算动画。让它自然执行完毕进入 Waiting 状态即可。

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

	// 恢复玩家思考倒计时
	if rem := t.actionTimer.Resume(); rem > 0 {
		// 定时器已成功在底层恢复，我们只需要通知前端更新 UI
		if t.CurrentHand != nil && t.CurrentHand.CurrentPlayerIndex != -1 {
			seat := t.Seats[t.CurrentHand.CurrentPlayerIndex]
			if seat.Player != nil {
				currentPlayerID := seat.Player.User.ID
				
				// 1. 广播倒计时恢复，让所有人的进度条继续走
				t.messenger.Broadcast(MsgTypeCountdown, "resume_action_timer", CountdownPayload{
					PlayerID: currentPlayerID,
					Seconds:  int(rem.Seconds()),
				})
				
				// 2. 重新向该玩家发送行动通知，以便前端重新渲染操作面板
				t.notifyCurrentPlayer(int(rem.Seconds()))
			}
		}
	}

	// 如果没有在游戏中，尝试重新触发开局检查
	if t.CurrentHand == nil {
		t.checkAndAutoStart()
	}

	return nil
}

// HandleMessage 处理游戏内的具体动作
func (t *Table) HandleMessage(userID string, msgType string, payload []byte) error {
	// 1. 如果游戏被暂停，拒绝处理任何动作
	if t.IsPaused {
		return fmt.Errorf("game is paused")
	}

	// 2. 统一的生命周期阶段校验
	// 局外动作 (SitDown, StandUp, Ready, CancelReady) 只能在游戏未开始时执行
	// 局内动作 (Action) 只能在游戏进行中执行
	isGameRunning := t.CurrentHand != nil

	switch msgType {
	case MsgTypeSitDown, MsgTypeStandUp, MsgTypeReady, MsgTypeCancel:
		if isGameRunning {
			return fmt.Errorf("cannot perform out-of-game action '%s' while game is running", msgType)
		}
	case MsgTypeAction:
		if !isGameRunning {
			return fmt.Errorf("cannot perform in-game action '%s' while game is not running", msgType)
		}
	}

	// 3. 路由到具体的处理函数
	switch msgType {
	case MsgTypeSitDown:
		return t.handleSitDown(userID, payload)
	case MsgTypeStandUp:
		return t.handleStandUp(userID)
	case MsgTypeReady:
		return t.handleReady(userID)
	case MsgTypeCancel:
		return t.handleCancelReady(userID)
	case MsgTypeAction:
		return t.handleAction(userID, payload)
	default:
		// 忽略不认识的消息
		return nil
	}
}

// ============================================================================
// 动作处理逻辑骨架 (Action Handlers)
// ============================================================================

func (t *Table) handleSitDown(userID string, payload []byte) error {
	// 1. 解析 payload 获取目标座位号 (SeatNumber)
	var sitPayload SitDownPayload
	if err := json.Unmarshal(payload, &sitPayload); err != nil {
		return fmt.Errorf("invalid sit_down payload: %w", err)
	}

	// 2. 校验座位是否合法且为空
	if sitPayload.SeatNumber < 0 || sitPayload.SeatNumber >= t.MaxPlayers {
		return fmt.Errorf("invalid seat number")
	}
	targetSeat := t.Seats[sitPayload.SeatNumber]
	if targetSeat.Player != nil {
		return fmt.Errorf("seat is already occupied")
	}

	// 3. 检查玩家是否已经在其他座位上。如果是，则执行“换座”逻辑：先清空原座位，再绑定到新座位。
	oldSeat := t.getSeatByUserID(userID)
	if oldSeat != nil {
		oldSeat.Player = nil
	}

	// 4. 从 t.Players 中查找该玩家，如果不存在则返回错误（理论上 OnPlayerJoin 已经创建了）
	player, exists := t.Players[userID]
	if !exists {
		return fmt.Errorf("player not found in room")
	}

	// 如果玩家筹码为0（比如之前破产了），重新分配初始筹码
	if player.Chips == 0 {
		player.Chips = t.InitialChips
		player.BuyInCount++
	}
	player.State = PlayerStateWaiting

	// 5. 将 Player 实体绑定到 Seat
	targetSeat.Player = player

	// 6. 广播状态更新 (MsgTypeStateUpdate)
	t.messenger.Broadcast(MsgTypeStateUpdate, "sit_down", t.BuildPublicSnapshot())
	return nil
}

func (t *Table) handleStandUp(userID string) error {
	// 1. 查找玩家所在座位
	seat := t.getSeatByUserID(userID)
	if seat == nil {
		return fmt.Errorf("player not in seat")
	}

	// 2. 游戏未开始，直接清空座位 (Seat.Player = nil)
	// 3. 重置玩家状态 (Player.State = PlayerStateWaiting)
	seat.Player.State = PlayerStateWaiting
	seat.Player = nil

	// 4. 如果当前正在进行开局倒计时 (Countdown)，必须打断/取消倒计时！
	if t.countdownTimer != nil {
		t.countdownTimer.Stop()
		t.messenger.Broadcast(MsgTypeCountdown, "cancel_countdown", CountdownPayload{Seconds: 0})
	}

	// 5. 广播状态更新 (MsgTypeStateUpdate)
	t.messenger.Broadcast(MsgTypeStateUpdate, "stand_up", t.BuildPublicSnapshot())
	return nil
}

func (t *Table) handleReady(userID string) error {
	// 1. 查找玩家所在座位
	seat := t.getSeatByUserID(userID)
	if seat == nil {
		return fmt.Errorf("player not in seat")
	}

	// 2. 修改玩家状态为 PlayerStateReady
	seat.Player.State = PlayerStateReady

	// 3. 广播状态更新 (MsgTypeStateUpdate)
	t.messenger.Broadcast(MsgTypeStateUpdate, "ready", t.BuildPublicSnapshot())

	// 4. 调用 t.checkAndAutoStart() 检查是否满足开局条件
	t.checkAndAutoStart()
	return nil
}

func (t *Table) handleCancelReady(userID string) error {
	// 1. 查找玩家所在座位
	seat := t.getSeatByUserID(userID)
	if seat == nil {
		return fmt.Errorf("player not in seat")
	}

	// 2. 修改玩家状态为 PlayerStateWaiting
	seat.Player.State = PlayerStateWaiting

	// 3. 如果当前正在进行开局倒计时 (Countdown)，必须打断/取消倒计时！
	if t.countdownTimer != nil {
		t.countdownTimer.Stop()
		t.messenger.Broadcast(MsgTypeCountdown, "cancel_countdown", CountdownPayload{Seconds: 0})
	}

	// 4. 广播状态更新 (MsgTypeStateUpdate)
	t.messenger.Broadcast(MsgTypeStateUpdate, "cancel_ready", t.BuildPublicSnapshot())
	return nil
}

func (t *Table) handleAction(userID string, payload []byte) error {
	// 1. 解析 payload 获取具体的打牌动作
	var actionPayload ClientActionPayload
	if err := json.Unmarshal(payload, &actionPayload); err != nil {
		return fmt.Errorf("invalid action payload: %w", err)
	}

	actionType := ActionType(actionPayload.Action)
	amount := actionPayload.Amount

	// 2. 基础校验：动作类型是否合法
	switch actionType {
	case ActionTypeFold, ActionTypeCheck, ActionTypeCall, ActionTypeBet, ActionTypeRaise, ActionTypeAllIn:
		// 合法动作
	default:
		return fmt.Errorf("unknown action type: %s", actionType)
	}

	// 3. 基础校验：玩家是否在座位上且参与了本局
	seat := t.getSeatByUserID(userID)
	if seat == nil || seat.Player == nil {
		return fmt.Errorf("player not in seat")
	}
	if seat.Player.State == PlayerStateWaiting {
		return fmt.Errorf("player is not active in current hand")
	}

	// 4. 基础校验：是否轮到该玩家行动
	if t.CurrentHand.CurrentPlayerIndex != seat.SeatNumber {
		return fmt.Errorf("not your turn")
	}

	// 5. 将具体的业务逻辑交给状态机处理
	return t.processPlayerAction(userID, actionType, amount)
}

// getSeatByUserID 根据 UserID 查找玩家所在的座位
// 如果玩家不在座位上（比如是旁观者），返回 nil
func (t *Table) getSeatByUserID(userID string) *Seat {
	// 虽然 t.Players 可以快速找到 Player，但 Player 结构体中并没有记录 SeatNumber。
	// 因此要判断玩家是否在座位上，仍然需要遍历 t.Seats。
	// 由于 MaxPlayers 通常很小 (2~9)，这里的遍历开销可以忽略不计。
	for _, s := range t.Seats {
		if s.Player != nil && s.Player.User.ID == userID {
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

	// 2. 检查是否满员，且所有人在座玩家都已准备
	readyCount := 0
	for _, seat := range t.Seats {
		if seat.Player == nil {
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
	// 启动 3 秒倒计时
	t.countdownTimer.Start(3*time.Second, func() {
		// 倒计时结束后，再次检查条件是否仍然满足
		t.startNewHand()
	})
	
	t.messenger.Broadcast(MsgTypeCountdown, "start_countdown", CountdownPayload{Seconds: 3})
}
