package texas

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/lllllan02/texas-holdem/pkg/room"
)

// 游戏特定的 Action 定义
const (
	// 引擎系统通知类 Action
	ActionSysRoundEnd  = "texas.sys.round_end"  // 牌局结束
	ActionSysCountdown = "texas.sys.countdown"  // 倒计时
	ActionSysHandStart = "texas.sys.hand_start" // 新手牌开始

	// 房主控制操作
	ActionHostStart = "texas.host.start" // 房主恢复游戏
	ActionHostPause = "texas.host.pause" // 房主暂停游戏

	// 玩家基础操作
	ActionPlayerSit         = "texas.player.sit"          // 玩家坐下
	ActionPlayerStand       = "texas.player.stand"        // 玩家站起
	ActionPlayerReady       = "texas.player.ready"        // 玩家准备进入下一局
	ActionPlayerCancelReady = "texas.player.cancel_ready" // 玩家取消准备

	// 玩家打牌操作
	ActionGameBet   = "texas.game.bet"   // 下注/加注
	ActionGameFold  = "texas.game.fold"  // 弃牌
	ActionGameCheck = "texas.game.check" // 过牌
	ActionGameCall  = "texas.game.call"  // 跟注
)

// Engine 德州扑克游戏引擎，实现了 room.GameEngine 接口
type Engine struct {
	room     *room.Room
	Table    *Table
	updateCh chan struct{}

	// 引擎级控制状态
	isPaused bool // 房主是否暂停了游戏（初始默认为 false，只要大家准备好就自动开始）

	// 定时器相关
	timerCancelCh chan struct{} // 用于控制并取消后台的倒计时 Goroutine
}

// NewEngine 创建一个新的德州扑克游戏引擎
func NewEngine() *Engine {
	return &Engine{
		updateCh: make(chan struct{}, 10),
		isPaused: false, // 初始状态为未暂停，只要所有人准备就自动开始
	}
}

// UpdateChannel 返回状态更新信号通道
func (e *Engine) UpdateChannel() <-chan struct{} {
	return e.updateCh
}

// triggerSync 触发状态同步广播
func (e *Engine) triggerSync() {
	select {
	case e.updateCh <- struct{}{}:
	default:
	}
}

// OnInit 房间初始化时调用
func (e *Engine) OnInit(room *room.Room, param map[string]any) {
	e.room = room
	e.Table = NewTable(param)

	log.Printf("[TexasEngine] 房间 %s 初始化德州扑克引擎\n", room.GetID())
}

// OnDestroy 房间销毁时调用
func (e *Engine) OnDestroy() {
	log.Printf("[TexasEngine] 房间 %s 销毁，清理引擎资源\n", e.room.GetID())
	e.cancelTimer() // 清理定时器 Goroutine
}

// OnPlayerJoin 玩家加入房间时调用
func (e *Engine) OnPlayerJoin(playerID string) {
	log.Printf("[TexasEngine] 玩家 %s 加入房间 %s\n", playerID, e.room.GetID())
	// 这里可以处理旁观者逻辑，或者自动坐下逻辑
}

// OnPlayerLeave 玩家离开房间时调用
func (e *Engine) OnPlayerLeave(playerID string) {
	log.Printf("[TexasEngine] 玩家 %s 离开房间 %s\n", playerID, e.room.GetID())
	// 这里可以处理玩家离开座位、自动弃牌等逻辑
}

// HandleMessage 处理游戏特定消息
func (e *Engine) HandleMessage(playerID string, action string, content string) {
	// 路由拦截器：如果游戏处于暂停状态，拦截所有 game 级别的操作
	if e.isPaused && strings.HasPrefix(action, "texas.game.") {
		e.room.SendTo(playerID, "error", `{"message": "游戏已暂停，无法进行该操作"}`)
		return
	}

	// 动作白名单拦截器：检查玩家是否有权限执行该动作
	if !e.checkActionAllowed(playerID, action) {
		e.room.SendTo(playerID, "error", fmt.Sprintf(`{"message": "当前状态不允许执行该操作: %s"}`, action))
		return
	}

	// 1. 无条件广播玩家的动作事件，用于驱动前端表现层（日志、动画等）
	// 即使操作最终失败（比如还没轮到他说话），前端也可能需要知道他尝试了什么
	// 注意：必须在执行具体 handle 之前广播，保证事件顺序（比如先发“下注”事件，再发“游戏结束”事件）
	e.room.Broadcast(playerID, action, content)

	changed := false
	switch action {

	// 玩家操作
	case ActionPlayerSit:
		changed = e.handleSit(playerID, content)
	case ActionPlayerStand:
		changed = e.handleStand(playerID)

	// 房主操作
	case ActionHostStart:
		changed = e.handleToggleRunning(playerID, false) // 恢复游戏（取消暂停）
	case ActionHostPause:
		changed = e.handleToggleRunning(playerID, true) // 暂停游戏

	// 游戏操作
	case ActionGameBet, ActionGameCall, ActionGameCheck, ActionGameFold:
		changed = e.handlePlayerAction(playerID, action, content)
	case ActionPlayerReady:
		changed = e.handleReady(playerID)
	case ActionPlayerCancelReady:
		changed = e.handleCancelReady(playerID)

	default:
	}

	// 2. 如果操作成功导致了状态改变，触发全量状态同步
	if changed {
		e.triggerSync()
	}
}

func (e *Engine) handleToggleRunning(playerID string, targetPause bool) bool {
	// 只有房主可以控制游戏状态
	if playerID != e.room.GetHostID() {
		actionName := "暂停"
		if !targetPause {
			actionName = "恢复"
		}
		e.room.SendTo(playerID, "error", fmt.Sprintf(`{"message": "只有房主可以%s游戏"}`, actionName))
		return false
	}

	if e.isPaused == targetPause {
		return false // 状态未改变
	}
	e.isPaused = targetPause

	if !targetPause {
		log.Printf("[TexasEngine] 房间 %s 房主 %s 恢复了游戏\n", e.room.GetID(), playerID)
		e.checkAndStartCountdown()
	} else {
		log.Printf("[TexasEngine] 房间 %s 房主 %s 暂停了游戏\n", e.room.GetID(), playerID)

		// 注意：这里只是设置了标记，不会打断当前正在进行的牌局，
		// 而是在当前牌局结算（Showdown）后，阻止自动进入下一局。
		e.cancelTimer() // 如果刚好在等待下一局的倒计时中，取消它
	}

	return true
}

func (e *Engine) handleSit(playerID string, content string) bool {
	var req struct {
		Seat int `json:"seat"`
	}
	if err := json.Unmarshal([]byte(content), &req); err != nil {
		log.Printf("[TexasEngine] 玩家 %s 坐下参数解析失败: %v\n", playerID, err)
		return false
	}

	if err := e.Table.SitDown(playerID, req.Seat); err != nil {
		log.Printf("[TexasEngine] 玩家 %s 坐下失败: %v\n", playerID, err)
		// 可以发送一个错误消息给该玩家
		e.room.SendTo(playerID, "error", `{"message": "`+err.Error()+`"}`)
		return false
	}

	log.Printf("[TexasEngine] 玩家 %s 坐下座位 %d\n", playerID, req.Seat)

	// 标记状态已改变，需要广播
	return true
}

func (e *Engine) handleStand(playerID string) bool {
	if err := e.Table.StandUp(playerID); err != nil {
		log.Printf("[TexasEngine] 玩家 %s 站起失败: %v\n", playerID, err)
		return false
	}

	log.Printf("[TexasEngine] 玩家 %s 站起\n", playerID)

	// 检查是否因为玩家站起导致人数不足，或者剩余玩家都已准备
	e.checkAndStartCountdown()

	// 标记状态已改变，需要广播
	return true
}

func (e *Engine) handlePlayerAction(playerID string, action string, content string) bool {
	// 1. 找到玩家
	seatIdx := e.Table.FindPlayerSeat(playerID)
	if seatIdx == -1 {
		return false
	}
	player := e.Table.Seats[seatIdx]

	// 2. 根据 action 更新玩家状态
	if action == ActionGameFold {
		player.IsFolded = true
	}

	log.Printf("[TexasEngine] 玩家 %s 执行了动作 %s: %s\n", playerID, action, content)

	// 临时：模拟流程推进，现在改为推进轮次，而不是直接进入下一阶段
	if e.Table.Round != nil {
		e.Table.AdvanceTurn()

		// 骨架逻辑：如果推进后变成了 SHOWDOWN 或 WAITING（一局结束），并且当前引擎未暂停，则等待玩家准备
		if !e.Table.IsGameRunning() && !e.isPaused {
			log.Printf("[TexasEngine] 牌局结束，等待玩家确认结果\n")

			// 构造一个简单的结算结果
			result := map[string]any{
				"message": "牌局结束，请确认结果并准备下一局",
				// TODO: 这里可以加入赢家是谁、赢了多少筹码等真实结算数据
			}
			resultBytes, _ := json.Marshal(result)
			e.room.Broadcast("", ActionSysRoundEnd, string(resultBytes))
		}
	}

	return true
}

func (e *Engine) handleReady(playerID string) bool {
	// 找到玩家并设置准备状态
	seatIdx := e.Table.FindPlayerSeat(playerID)
	if seatIdx == -1 {
		return false // 旁观者不能准备
	}
	player := e.Table.Seats[seatIdx]

	if player.IsReady {
		return false // 已经准备过了
	}

	// TODO: 如果游戏正在进行中（Round != nil 且 Stage != WAITING），
	// 允许新落座的玩家点击准备，但需要标记他为“等待下一局加入”的状态，
	// 而不能让他立刻参与到当前正在进行的牌局中。

	player.IsReady = true
	log.Printf("[TexasEngine] 玩家 %s 已准备\n", playerID)

	// 检查是否所有在座玩家都已准备
	e.checkAndStartCountdown()

	return true
}

func (e *Engine) handleCancelReady(playerID string) bool {
	// 找到玩家并设置取消准备状态
	seatIdx := e.Table.FindPlayerSeat(playerID)
	if seatIdx == -1 {
		return false // 旁观者不能取消准备
	}
	player := e.Table.Seats[seatIdx]

	if !player.IsReady {
		return false // 还没准备过
	}

	player.IsReady = false
	log.Printf("[TexasEngine] 玩家 %s 取消准备\n", playerID)

	// 检查是否因为玩家取消准备导致不满足开局条件，取消倒计时
	e.checkAndStartCountdown()

	return true
}

// checkAndStartCountdown 检查是否满足开局条件，如果满足则开始倒计时，否则取消倒计时
func (e *Engine) checkAndStartCountdown() {
	// 只有在未暂停，且没有进行中的牌局时，才需要检查
	if e.isPaused || e.Table.IsGameRunning() {
		return
	}

	allReady := true
	seatedCount := 0
	for _, p := range e.Table.Seats {
		if p != nil {
			seatedCount++
			if !p.IsReady {
				allReady = false
			}
		}
	}

	if seatedCount < 2 {
		log.Printf("[TexasEngine] 人数不足，取消/无法开始倒计时\n")
		e.cancelTimer()
	} else if allReady {
		log.Printf("[TexasEngine] 玩家均已准备，开始倒计时\n")
		e.scheduleNextHand(3 * time.Second)
	} else {
		log.Printf("[TexasEngine] 尚有玩家未准备，等待中\n")
		e.cancelTimer() // 如果有人取消准备（虽然目前没有这个动作，但为了严谨），取消倒计时
	}
}
func (e *Engine) scheduleNextHand(delay time.Duration) {
	e.cancelTimer() // 取消可能存在的旧定时器

	cancelCh := make(chan struct{})
	e.timerCancelCh = cancelCh
	seconds := int(delay.Seconds())

	go func() {
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()

		for seconds > 0 {
			e.room.Broadcast("", ActionSysCountdown, seconds)
			select {
			case <-ticker.C:
				seconds--
			case <-cancelCh:
				return // 倒计时被取消
			}
		}

		// 倒计时结束，提交到房间的事件循环中执行开始逻辑
		e.room.GetHub().Execute(func() {
			// 再次检查是否被暂停
			if e.isPaused {
				return
			}

			if err := e.Table.StartNewHand(); err != nil {
				log.Printf("[TexasEngine] 自动开启下一局失败: %v\n", err)
				e.isPaused = true // 开启失败则暂停游戏
				e.room.Broadcast("", ActionHostPause, "人数不足，自动暂停")
			} else {
				e.room.Broadcast("", ActionSysHandStart, "新的一局开始了！")
			}

			// 触发状态同步，把发牌等新状态推给前端
			e.triggerSync()
		})
	}()
}

// cancelTimer 取消当前的定时器
func (e *Engine) cancelTimer() {
	if e.timerCancelCh != nil {
		close(e.timerCancelCh)
		e.timerCancelCh = nil

		// 广播取消倒计时的消息，让前端隐藏倒计时
		e.room.Broadcast("", ActionSysCountdown, -1)
	}
}

// GetState 获取当前游戏状态快照
func (e *Engine) GetState(playerID string) any {
	if e.Table == nil {
		return nil
	}

	// 找到玩家的座位号
	seatIdx := e.Table.FindPlayerSeat(playerID)

	// 获取允许的动作
	allowedActions := e.Table.CalculateAllowedActions(playerID, seatIdx, e.isPaused)

	// 获取 Table 的安全快照
	snap := e.Table.GetSnapshot(playerID)

	// 包装一层，把引擎级的状态也放进去
	return map[string]any{
		"isPaused": e.isPaused,
		"table":    snap,
		"myInfo": map[string]any{
			"seatIdx":        seatIdx,
			"allowedActions": allowedActions,
		},
	}
}

// checkActionAllowed 检查玩家是否有权限执行该动作
func (e *Engine) checkActionAllowed(playerID string, action string) bool {
	// 房主操作不走这个白名单，由 handleToggleRunning 内部判断
	if strings.HasPrefix(action, "texas.host.") {
		return true
	}

	// 找到玩家的座位号
	seatIdx := e.Table.FindPlayerSeat(playerID)

	// 获取允许的动作
	allowedActions := e.Table.CalculateAllowedActions(playerID, seatIdx, e.isPaused)
	for _, allowedAction := range allowedActions {
		if action == allowedAction {
			return true
		}
	}

	return false
}
