package texas

import (
	"fmt"
)

// ============================================================================
// 核心状态机流转 (State Machine Flow)
// ============================================================================

// startNewHand 开始新的一局游戏
func (t *Table) startNewHand() error {
	// 1. 获取所有准备好的玩家，分配位置 (BTN, SB, BB 等)
	// 2. 初始化 Hand 结构体，设置 Stage = HandStagePreFlop
	// 3. 洗牌 (Deck.Shuffle)
	// 4. 强制扣除小盲注和大盲注
	// 5. 给每个活跃玩家发 2 张底牌
	// 6. 计算本轮的行动顺序 (ActionOrder)，通常是 UTG 第一个说话
	// 7. 广播游戏开始的公共快照 (BuildPublicSnapshot)
	// 8. 给每个玩家私发底牌 (DealHoleCardsPayload)
	// 9. 通知第一个玩家行动 (TurnNotificationPayload)

	fmt.Println("TODO: startNewHand - 初始化牌局，扣盲注，发底牌")
	return nil
}

// processPlayerAction 处理玩家的具体下注/弃牌动作
func (t *Table) processPlayerAction(playerID string, action ActionType, amount int) error {
	// 1. 校验动作是否合法 (比如余额是否足够，加注是否达到最小加注额)
	// TODO: 实现具体的动作合法性校验逻辑

	// 2. 根据动作更新玩家状态 (Chips, CurrentBet, State)
	// TODO: 根据 actionType (Fold, Check, Call, Bet, Raise, AllIn) 执行相应的状态修改

	// 3. 更新底池状态 (Pot, CurrentBet, MinRaise)
	// TODO: 累加下注金额，更新当前下注圈的最高下注额和最小加注额

	// 4. 标记该玩家本轮已行动
	// TODO: seat.Player.HasActedThisRound = true

	// 5. 广播玩家的动作，用于前端播放动画
	// TODO: t.messenger.Broadcast(MsgTypeStateUpdate, "player_action", t.BuildPublicSnapshot())

	// 6. 推进状态机
	t.advanceStateMachine()

	fmt.Printf("TODO: processPlayerAction - 玩家 %s 执行了 %s %d\n", playerID, action, amount)
	return nil
}

// advanceStateMachine 推进状态机（判断当前下注圈是否结束，是否进入下一阶段）
func (t *Table) advanceStateMachine() {
	// 1. 检查是否只剩下一个未弃牌的玩家
	//    -> 如果是，直接提前结束本局 (early finish)，将底池给该玩家

	// 2. 检查当前下注圈是否所有人都已表态，且下注金额已平齐 (All bets matched)
	//    -> 如果未平齐，找到下一个该说话的玩家，发送 TurnNotificationPayload
	//    -> 如果已平齐，进入下一阶段 (nextStage)

	fmt.Println("TODO: advanceStateMachine - 检查是否需要进入下一阶段")
}

// nextStage 进入下一个游戏阶段 (Flop, Turn, River, Showdown)
func (t *Table) nextStage() {
	// 1. 将所有玩家本轮的下注 (CurrentBet) 收集到底池 (Pot) 中
	// 2. 重置所有玩家的 CurrentBet = 0 和 HasActedThisRound = false
	// 3. 根据当前 Stage 推进到下一个 Stage：
	//    - PREFLOP -> FLOP: 发 3 张公共牌
	//    - FLOP -> TURN: 发 1 张公共牌
	//    - TURN -> RIVER: 发 1 张公共牌
	//    - RIVER -> SHOWDOWN: 进入摊牌结算
	// 4. 特殊情况：如果所有未弃牌玩家都已 All-in (或只剩一人未 All-in)
	//    -> 直接发完剩余的公共牌，快进到 SHOWDOWN
	// 5. 重新计算下一轮的行动顺序 (通常从 SB 开始)
	// 6. 广播进入新阶段的快照
	// 7. 通知第一个玩家行动 (如果还没到 SHOWDOWN)

	fmt.Println("TODO: nextStage - 发公共牌，推进阶段")
}

// handleShowdown 处理摊牌和结算
func (t *Table) handleShowdown() {
	// 1. 收集所有最终的下注到底池
	// 2. 计算边池 (Side Pots) - 极其重要且复杂的逻辑！
	// 3. 评估所有未弃牌玩家的牌型大小 (Hand Evaluator)
	// 4. 根据牌型大小分配主池和各个边池
	// 5. 构建 ShowdownSummary，包含赢家、赢的金额、亮出的底牌
	// 6. 广播结算快照
	// 7. 记录对局历史 (Histories)
	// 8. 延迟几秒后，清理牌桌，状态重置为 WAITING
	// 9. 破产清理与筹码补充：
	//    - 休闲模式：遍历所有玩家，如果 Chips == 0，则自动为其补充 t.InitialChips 的筹码，并将其 RebuyCount + 1。
	//    - TODO: 锦标赛模式 (SNG/MTT)：直接淘汰 (Eliminate) 破产玩家，不留座位。
	// 10. 尝试自动开始下一局 (调用 checkAndAutoStart)

	fmt.Println("TODO: handleShowdown - 比牌，分池，结算")
}
