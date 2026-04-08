package texas

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/lllllan02/texas-holdem/pkg/core"
	"github.com/lllllan02/texas-holdem/pkg/user"
	"github.com/lllllan02/texas-holdem/pkg/utils"
)

type TableOptions struct {
	PlayerCount    int `json:"player_count"`    // 开局必须达到的玩家数 (通常为 2, 6, 9)，同时也是座位的数量
	SmallBlind     int `json:"small_blind"`     // 小盲注金额
	BigBlind       int `json:"big_blind"`       // 大盲注金额
	InitialChips   int `json:"initial_chips"`   // 初始筹码量
	ActionTimeout  int `json:"action_timeout"`  // 玩家行动超时时间(秒)
	OfflineTimeout int `json:"offline_timeout"` // 断线玩家的行动超时时间(秒)
}

// Table 德州扑克牌桌 (GameEngine 的具体实现)
// 负责管理跨局的持久化状态，包括座位分配、庄家位置的流转以及对局历史。
type Table struct {
	// --- 内部依赖 ---
	messenger core.Messenger // 注入的消息发送器 (在 OnInit 时传入)

	// --- 房间规则/配置配置 (通常在游戏过程中不变) ---
	SmallBlind     int // 小盲注金额
	BigBlind       int // 大盲注金额
	InitialChips   int // 初始筹码（玩家入座后统一分配的数量）
	ActionTimeout  int // 玩家行动超时时间(秒)
	OfflineTimeout int // 断线玩家的行动超时时间(秒)

	// --- 牌桌运行时状态 (随着游戏进行不断变化) ---
	Seats       []*Player          // 座位数组，其长度即为该房间的 PlayerCount
	Players     map[string]*Player // 记录所有参与过本桌游戏的玩家实体，用于恢复筹码和买入次数 (Key: UserID)
	ButtonSeat  int                // 当前庄家 (Dealer/Button) 所在的座位号
	CurrentHand *Hand              // 当前正在进行的单局游戏实例（如果不在游戏中则为 nil）
	Histories   []*ShowdownSummary // 历史对局记录列表，用于战绩回放（其长度即为已完成的局数）
	IsPaused    bool               // 游戏是否处于暂停状态

	// --- 定时器 ---
	countdownTimer *utils.PausableTimer // 开局倒计时定时器
	actionTimer    *utils.PausableTimer // 玩家行动倒计时定时器
	showdownTimer  *utils.PausableTimer // 动画缓冲定时器（用于摊牌结算、提前结束、All-in快进等场景，给前端留出播放动画的时间）
}

// NewTable 创建一个新的德州扑克牌桌实例
// 注意：此时不解析业务配置，只做最基础的内存结构初始化，并赋予默认的房间配置
func NewTable() *Table {
	smallBlind := 10

	t := &Table{
		// --- 默认房间配置 ---
		SmallBlind:     smallBlind,
		BigBlind:       2 * smallBlind,
		InitialChips:   200 * smallBlind, // 默认 100 个大盲
		ActionTimeout:  30,               // 默认 30 秒思考时间
		OfflineTimeout: 5,                // 默认断线玩家 5 秒思考时间

		// --- 运行时状态初始化 ---
		Players:     make(map[string]*Player),
		ButtonSeat:  -1, // 游戏未开始时没有庄家，用 -1 表示
		Histories:   make([]*ShowdownSummary, 0),
		IsPaused:    false,
		CurrentHand: nil,
		Seats:       nil, // Seats 会在 OnInit 中根据最终的 PlayerCount 初始化
	}

	// 初始化定时器并配置生命周期回调
	t.countdownTimer = utils.NewPausableTimer(utils.WithOnStop(func() {
		// 只有在 messenger 存在时才广播，避免在初始化或销毁时报错
		if t.messenger != nil {
			t.messenger.Broadcast(MsgTypeCountdown, "cancel_countdown", CountdownPayload{Seconds: 0})
		}
	}))
	// 玩家行动倒计时定时器
	t.actionTimer = utils.NewPausableTimer(utils.WithOnResume(func(rem time.Duration) {
		// 定时器已成功在底层恢复，我们只需要通知前端更新 UI
		if t.messenger != nil && t.CurrentHand != nil && t.CurrentHand.CurrentPlayerIndex != -1 {
			player := t.Seats[t.CurrentHand.CurrentPlayerIndex]
			if player != nil {
				currentPlayerID := player.User.ID

				// 1. 广播倒计时恢复，让所有人的进度条继续走
				t.messenger.Broadcast(MsgTypeCountdown, "resume_action_timer", CountdownPayload{
					PlayerID: currentPlayerID,
					Seconds:  int(rem.Seconds()),
				})

				// 2. 重新向该玩家发送行动通知，以便前端重新渲染操作面板
				t.notifyCurrentPlayer(int(rem.Seconds()))
			}
		}
	}))
	t.showdownTimer = utils.NewPausableTimer()

	return t
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

	// 1. 解析配置并覆盖默认值
	playerCount := 9 // 默认 9 人桌
	if len(options) > 0 {
		var opts TableOptions
		if err := json.Unmarshal(options, &opts); err != nil {
			return fmt.Errorf("failed to parse table options: %w", err)
		}

		// 校验并覆盖配置，超过合理范围则忽略（使用默认值）
		if opts.PlayerCount >= 2 && opts.PlayerCount <= 9 {
			playerCount = opts.PlayerCount
		}

		if opts.BigBlind >= 2 {
			t.BigBlind = opts.BigBlind
		}

		if opts.SmallBlind >= 1 {
			t.SmallBlind = opts.SmallBlind
		} else if opts.BigBlind > 0 {
			// 如果没传小盲（或传了非法值），自动计算小盲
			t.SmallBlind = t.BigBlind / 2
		}

		if opts.InitialChips >= t.BigBlind*10 { // 至少带入 10 个大盲
			t.InitialChips = opts.InitialChips
		} else if opts.BigBlind > 0 {
			// 如果没传初始筹码（或传了非法值），默认 100 个大盲
			t.InitialChips = t.BigBlind * 100
		}

		if opts.ActionTimeout >= 5 && opts.ActionTimeout <= 120 { // 思考时间 5~120 秒
			t.ActionTimeout = opts.ActionTimeout
		}

		if opts.OfflineTimeout >= 1 && opts.OfflineTimeout <= 30 {
			t.OfflineTimeout = opts.OfflineTimeout
		}
	}

	// 2. 根据最终确定的 playerCount 初始化座位数组
	t.Seats = make([]*Player, playerCount)
	for i := 0; i < playerCount; i++ {
		t.Seats[i] = nil
	}

	return nil
}

// OnDestroy 引擎被销毁时调用，用于清理资源
func (t *Table) OnDestroy() {
	// 先置空 messenger，这样后续 Stop 定时器时就不会触发不必要的广播
	t.messenger = nil

	// 1. 停止所有正在运行的定时器，防止 Goroutine 泄漏
	t.countdownTimer.Stop()
	t.actionTimer.Stop()
	t.showdownTimer.Stop()

	// 2. 清理大对象引用，帮助 GC 更快回收内存
	t.Players = nil
	t.Seats = nil
	t.CurrentHand = nil
	t.Histories = nil
	t.messenger = nil
}

// OnPlayerJoin 玩家加入游戏时调用
// 注意：这只是建立连接进入房间，并不代表落座。玩家默认是旁观者。
func (t *Table) OnPlayerJoin(u *user.User) {
	// 1. 将玩家信息保存或更新到 Players 映射中
	player, exists := t.Players[u.ID]
	if !exists {
		// 第一次进入房间，创建玩家实体并分配初始筹码
		player = NewPlayer(u, t.InitialChips)
		t.Players[u.ID] = player
	} else {
		// 玩家之前就在房间里，更新用户信息（可能改了名字或头像）
		player.User = u
	}

	// 2. 如果该玩家之前已经落座（断线重连），恢复其在线状态并广播
	if player.IsOffline {
		player.IsOffline = false
		log.Printf("TexasEngine: Player [%s] reconnected", u.ID)
		// 只有当玩家在座位上时，才需要向其他人广播其重连状态（旁观者重连不需要广播）
		if t.getSeatIndexByUserID(u.ID) != -1 {
			t.messenger.Broadcast(MsgTypeStateUpdate, ReasonPlayerReconnected, t.BuildPublicSnapshot())
		}
	} else {
		log.Printf("TexasEngine: Player [%s] joined the room", u.ID)
	}

	// 3. 发送当前的全局快照给该玩家，以便前端恢复画面
	snap := t.BuildPersonalSnapshot(u.ID)
	t.messenger.SendTo(u.ID, MsgTypeStateUpdate, ReasonPlayerJoined, snap)

	// 4. 如果当前正在进行游戏，且正好轮到某位玩家行动，则为其补发行动通知（包含倒计时和可用操作）
	if t.CurrentHand != nil && t.CurrentHand.CurrentPlayerIndex != -1 {
		currentPlayer := t.Seats[t.CurrentHand.CurrentPlayerIndex]
		if currentPlayer != nil {
			// 计算剩余时间
			rem := int(t.actionTimer.Remaining().Seconds())
			if rem < 0 {
				rem = 0
			}
			
			payload := t.buildTurnNotificationPayload(currentPlayer, rem)
			t.messenger.SendTo(u.ID, MsgTypeTurnNotification, "turn_notification", payload)
		}
	}
}

// OnPlayerLeave 玩家离开游戏/掉线时调用
func (t *Table) OnPlayerLeave(userID string) {
	// 1. 从 Players 映射中找到该玩家
	player, exists := t.Players[userID]
	if !exists {
		return
	}

	log.Printf("TexasEngine: Player [%s] left the room", userID)

	// 2. 标记为断线状态
	player.IsOffline = true

	// 3. 如果该玩家在座位上，广播其离开（断线）的消息
	// 旁观者离开不需要广播，以免打扰正在打牌的人
	if t.getSeatIndexByUserID(userID) != -1 {
		t.messenger.Broadcast(MsgTypeStateUpdate, ReasonPlayerLeft, t.BuildPublicSnapshot())
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

	// 暂停玩家思考倒计时（恢复时继续计时）
	t.actionTimer.Pause()

	// 取消开局倒计时（恢复时若满足条件会重新开始倒计时）
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
	t.actionTimer.Resume()

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

// 动作处理逻辑骨架 (Action Handlers)

func (t *Table) handleSitDown(userID string, payload []byte) error {
	// 1. 解析 payload 获取目标座位号 (SeatNumber)
	var sitPayload SitDownPayload
	if err := json.Unmarshal(payload, &sitPayload); err != nil {
		return fmt.Errorf("invalid sit_down payload: %w", err)
	}

	// 2. 校验座位是否合法且为空
	if sitPayload.SeatNumber < 0 || sitPayload.SeatNumber >= len(t.Seats) {
		return fmt.Errorf("invalid seat number")
	}
	if seat := t.Seats[sitPayload.SeatNumber]; seat != nil {
		return fmt.Errorf("seat is already occupied")
	}

	// 3. 检查玩家是否已经在其他座位上，如果是，则执行“换座”逻辑：先清空原座位，再绑定到新座位。
	if oldSeatIdx := t.getSeatIndexByUserID(userID); oldSeatIdx != -1 {
		t.Seats[oldSeatIdx] = nil
	}

	// 4. 从 t.Players 中查找该玩家，如果不存在则返回错误（理论上 OnPlayerJoin 已经创建了）
	player, exists := t.Players[userID]
	if !exists {
		return fmt.Errorf("player not found in room")
	}

	// 无论玩家之前是旁观者，还是从其他座位换座过来（换座前可能处于 Ready 状态），
	// 落座/换座后统一重置为 Waiting 状态，要求玩家重新手动准备，避免意外满足开局条件
	player.State = PlayerStateWaiting

	// 如果当前正在进行开局倒计时，因为玩家状态变成了 Waiting（破坏了全员准备的条件），必须打断/取消倒计时！
	t.countdownTimer.Stop()

	// 5. 将 Player 实体绑定到 Seat
	t.Seats[sitPayload.SeatNumber] = player
	log.Printf("TexasEngine: Player [%s] sat down at seat %d", userID, sitPayload.SeatNumber)

	// 6. 广播状态更新 (MsgTypeStateUpdate)
	t.messenger.Broadcast(MsgTypeStateUpdate, ReasonSitDown, t.BuildPublicSnapshot())
	return nil
}

func (t *Table) handleStandUp(userID string) error {
	// 1. 查找玩家所在座位
	seatIdx := t.getSeatIndexByUserID(userID)
	if seatIdx == -1 {
		return fmt.Errorf("player not in seat")
	}
	player := t.Seats[seatIdx]

	// 2. 游戏未开始，直接清空座位 (Seat.Player = nil)
	// 3. 重置玩家状态 (Player.State = PlayerStateWaiting)
	player.State = PlayerStateWaiting
	t.Seats[seatIdx] = nil
	log.Printf("TexasEngine: Player [%s] stood up from seat %d", userID, seatIdx)

	// 4. 如果当前正在进行开局倒计时 (Countdown)，必须打断/取消倒计时！
	t.countdownTimer.Stop()

	// 5. 广播状态更新 (MsgTypeStateUpdate)
	t.messenger.Broadcast(MsgTypeStateUpdate, ReasonStandUp, t.BuildPublicSnapshot())
	return nil
}

func (t *Table) handleReady(userID string) error {
	// 1. 查找玩家所在座位
	seatIdx := t.getSeatIndexByUserID(userID)
	if seatIdx == -1 {
		return fmt.Errorf("player not in seat")
	}
	player := t.Seats[seatIdx]

	// 2. 修改玩家状态为 PlayerStateReady
	player.State = PlayerStateReady
	
	// 准备时清空上一局的底牌，避免新的一局开局倒计时期间还显示上一局的牌
	player.HoleCards = nil
	
	log.Printf("TexasEngine: Player [%s] is ready at seat %d", userID, seatIdx)

	// 3. 广播状态更新 (MsgTypeStateUpdate)
	t.messenger.Broadcast(MsgTypeStateUpdate, ReasonReady, t.BuildPublicSnapshot())

	// 4. 调用 t.checkAndAutoStart() 检查是否满足开局条件
	t.checkAndAutoStart()
	return nil
}

func (t *Table) handleCancelReady(userID string) error {
	// 1. 查找玩家所在座位
	seatIdx := t.getSeatIndexByUserID(userID)
	if seatIdx == -1 {
		return fmt.Errorf("player not in seat")
	}
	player := t.Seats[seatIdx]

	// 2. 修改玩家状态为 PlayerStateWaiting
	player.State = PlayerStateWaiting
	log.Printf("TexasEngine: Player [%s] canceled ready at seat %d", userID, seatIdx)

	// 3. 如果当前正在进行开局倒计时 (Countdown)，必须打断/取消倒计时！
	t.countdownTimer.Stop()

	// 4. 广播状态更新 (MsgTypeStateUpdate)
	t.messenger.Broadcast(MsgTypeStateUpdate, ReasonCancelReady, t.BuildPublicSnapshot())
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
	seatIdx := t.getSeatIndexByUserID(userID)
	if seatIdx == -1 {
		return fmt.Errorf("player not in seat")
	}
	player := t.Seats[seatIdx]
	if player.State == PlayerStateWaiting {
		return fmt.Errorf("player is not active in current hand")
	}

	// 4. 基础校验：是否轮到该玩家行动
	if t.CurrentHand.CurrentPlayerIndex != seatIdx {
		return fmt.Errorf("not your turn")
	}

	// 5. 将具体的业务逻辑交给状态机处理
	return t.processPlayerAction(userID, actionType, amount)
}

// getSeatIndexByUserID 根据 UserID 查找玩家所在的座位索引
// 如果玩家不在座位上（比如是旁观者），返回 -1
func (t *Table) getSeatIndexByUserID(userID string) int {
	for i, p := range t.Seats {
		if p != nil && p.User.ID == userID {
			return i
		}
	}
	return -1
}

// checkAndAutoStart 检查是否满足开局条件，如果满足则自动开始游戏
func (t *Table) checkAndAutoStart() {
	// 1. 检查是否有正在进行的游戏
	if t.CurrentHand != nil {
		return
	}

	// 2. 检查是否满员，且所有人在座玩家都已准备
	for _, p := range t.Seats {
		if p == nil {
			return // 未满员，不开始
		}
		if p.State != PlayerStateReady {
			return // 有人未准备，不开始
		}
	}

	// 3. 所有条件满足，自动触发发牌流程
	// 启动 3 秒倒计时
	log.Printf("TexasEngine: All players ready, starting 3s countdown for new hand")
	t.countdownTimer.Start(3*time.Second, func() {
		t.messenger.Execute(func() {
			t.countdownTimer.Stop() // 倒计时结束，清理状态，否则 IsActive 会一直为 true 导致公牌被隐藏
			
			// 倒计时结束后，尝试开始新的一局
			if err := t.startNewHand(); err != nil {
				// 如果开局失败（比如有人在倒计时结束瞬间掉线导致不满员），广播错误并重置状态
				t.CurrentHand = nil
				t.messenger.Broadcast(core.MsgTypeError, "start_failed", err.Error())

				// 重新检查是否满足开局条件（可能需要重新等待玩家准备）
				t.checkAndAutoStart()
			}
		})
	})

	// 倒计时开始时，广播清空了历史数据的状态快照
	t.messenger.Broadcast(MsgTypeStateUpdate, "countdown_started", t.BuildPublicSnapshot())
	t.messenger.Broadcast(MsgTypeCountdown, "start_countdown", CountdownPayload{Seconds: 3})
}
