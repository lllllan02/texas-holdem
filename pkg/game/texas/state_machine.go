package texas

import (
	"fmt"
	"math/rand"
	"time"
)

// ============================================================================
// 核心状态机流转 (State Machine Flow)
// ============================================================================

// startNewHand 开始新的一局游戏
func (t *Table) startNewHand() error {
	// 1. 获取所有准备好的玩家，分配位置 (BTN, SB, BB 等)
	activeSeats := make([]int, 0)
	for _, seat := range t.Seats {
		if seat.Player != nil && seat.Player.State == PlayerStateReady {
			activeSeats = append(activeSeats, seat.SeatNumber)
			// 将玩家状态改为 Active
			seat.Player.State = PlayerStateActive
			seat.Player.CurrentBet = 0
			seat.Player.HasActedThisRound = false
			seat.Player.HoleCards = nil
		}
	}

	if len(activeSeats) < 2 {
		return fmt.Errorf("not enough players to start a hand")
	}

	// 确定庄家 (ButtonSeat)
	if t.ButtonSeat == -1 {
		// 第一局，随机选一个庄家
		t.ButtonSeat = activeSeats[rand.Intn(len(activeSeats))]
	} else {
		// 顺时针找下一个有人的座位
		found := false
		for i := 1; i <= t.MaxPlayers; i++ {
			nextSeat := (t.ButtonSeat + i) % t.MaxPlayers
			for _, s := range activeSeats {
				if s == nextSeat {
					t.ButtonSeat = nextSeat
					found = true
					break
				}
			}
			if found {
				break
			}
		}
	}

	// 2. 初始化 Hand 结构体
	t.HandCount++
	t.CurrentHand = &Hand{
		ID:                 fmt.Sprintf("hand-%d", time.Now().UnixNano()),
		Stage:              HandStagePreFlop,
		Deck:               NewDeck(),
		BoardCards:         make([]Card, 0, 5),
		Pot:                0,
		SidePots:           make([]*SidePot, 0),
		CurrentBet:         0,
		MinRaise:           t.BigBlind,
		ActionOrder:        make([]int, 0),
		CurrentPlayerIndex: -1,
	}

	// 3. 洗牌
	t.CurrentHand.Deck.Shuffle()

	// 确定行动顺序 (从 Button 左边开始)
	// 先把 activeSeats 按从 Button 开始顺时针排序
	orderedSeats := make([]int, 0, len(activeSeats))
	for i := 1; i <= t.MaxPlayers; i++ {
		seatIdx := (t.ButtonSeat + i) % t.MaxPlayers
		for _, s := range activeSeats {
			if s == seatIdx {
				orderedSeats = append(orderedSeats, seatIdx)
				break
			}
		}
	}

	var sbSeat, bbSeat int
	if len(orderedSeats) == 2 {
		// 2人桌：Button 是 SB，另一个是 BB
		sbSeat = t.ButtonSeat
		bbSeat = orderedSeats[0]
		
		// 2人桌 PreFlop 行动顺序：SB(BTN) -> BB
		t.CurrentHand.ActionOrder = []int{sbSeat, bbSeat}
	} else {
		// 多人桌：Button 左边第一个是 SB，第二个是 BB
		sbSeat = orderedSeats[0]
		bbSeat = orderedSeats[1]

		// 多人桌 PreFlop 行动顺序：UTG (BB左边第一个) 开始，最后是 BB
		for i := 2; i < len(orderedSeats); i++ {
			t.CurrentHand.ActionOrder = append(t.CurrentHand.ActionOrder, orderedSeats[i])
		}
		t.CurrentHand.ActionOrder = append(t.CurrentHand.ActionOrder, sbSeat, bbSeat)
	}

	// 4. 强制扣除小盲注和大盲注
	sbPlayer := t.Seats[sbSeat].Player
	bbPlayer := t.Seats[bbSeat].Player

	sbAmount := t.SmallBlind
	if sbPlayer.Chips < sbAmount {
		sbAmount = sbPlayer.Chips
	}
	sbPlayer.Chips -= sbAmount
	sbPlayer.CurrentBet = sbAmount
	t.CurrentHand.Pot += sbAmount

	bbAmount := t.BigBlind
	if bbPlayer.Chips < bbAmount {
		bbAmount = bbPlayer.Chips
	}
	bbPlayer.Chips -= bbAmount
	bbPlayer.CurrentBet = bbAmount
	t.CurrentHand.Pot += bbAmount

	t.CurrentHand.CurrentBet = t.BigBlind

	// 5. 给每个活跃玩家发 2 张底牌
	for _, seatIdx := range activeSeats {
		cards, err := t.CurrentHand.Deck.Draw(2)
		if err != nil {
			return fmt.Errorf("failed to draw cards: %w", err)
		}
		t.Seats[seatIdx].Player.HoleCards = cards
	}

	// 6. 确定第一个说话的玩家
	t.CurrentHand.CurrentPlayerIndex = t.CurrentHand.ActionOrder[0]

	// 7. 广播游戏开始的公共快照
	t.messenger.Broadcast(MsgTypeStartHand, "start_hand", t.BuildPublicSnapshot())

	// 8. 给每个玩家私发底牌
	for _, seatIdx := range activeSeats {
		player := t.Seats[seatIdx].Player
		payload := DealHoleCardsPayload{Cards: player.HoleCards}
		t.messenger.SendTo(player.User.ID, MsgTypeDealHoleCards, "deal_hole_cards", payload)
	}

	// 9. 通知第一个玩家行动
	t.notifyCurrentPlayer(t.ActionTimeout)

	return nil
}

// notifyCurrentPlayer 通知当前玩家行动
// timeoutSeconds: 允许玩家思考的剩余时间（秒）。如果是正常流转，传入 t.ActionTimeout；如果是暂停恢复，传入剩余时间。
func (t *Table) notifyCurrentPlayer(timeoutSeconds int) {
	if t.CurrentHand == nil || t.CurrentHand.CurrentPlayerIndex == -1 {
		return
	}

	seat := t.Seats[t.CurrentHand.CurrentPlayerIndex]
	player := seat.Player

	// 计算可用的动作和金额限制
	callAmount := t.CurrentHand.CurrentBet - player.CurrentBet
	
	validActions := []ActionType{ActionTypeFold}
	
	details := ActionDetails{}

	if callAmount == 0 {
		validActions = append(validActions, ActionTypeCheck)
	} else {
		if player.Chips > callAmount {
			validActions = append(validActions, ActionTypeCall)
			details.CallAmount = callAmount
		}
	}

	// 是否可以加注 / 下注
	// 下注/加注的最小金额
	minRaiseTotal := t.CurrentHand.CurrentBet + t.CurrentHand.MinRaise
	raiseAmount := minRaiseTotal - player.CurrentBet

	if player.Chips > callAmount {
		if t.CurrentHand.CurrentBet == 0 {
			// 没人下注，可以 Bet
			validActions = append(validActions, ActionTypeBet)
			details.MinBet = t.BigBlind
			details.MaxBet = player.Chips
		} else {
			// 有人下注，可以 Raise
			if player.Chips >= raiseAmount {
				validActions = append(validActions, ActionTypeRaise)
				details.MinRaise = minRaiseTotal
				details.MaxRaise = player.Chips + player.CurrentBet
			}
		}
		
		// 只要还有筹码，就可以 All-in
		validActions = append(validActions, ActionTypeAllIn)
		details.AllinAmount = player.Chips + player.CurrentBet
	} else if player.Chips > 0 {
		// 筹码不够 Call，只能 All-in
		validActions = append(validActions, ActionTypeAllIn)
		details.AllinAmount = player.Chips + player.CurrentBet
	}

	payload := TurnNotificationPayload{
		PlayerID:       player.User.ID,
		ValidActions:   validActions,
		ActionDetails:  details,
		TimeoutSeconds: timeoutSeconds,
	}

	t.messenger.SendTo(player.User.ID, MsgTypeTurnNotification, "turn_notification", payload)
	
	// 启动玩家行动超时定时器
	// 注意：这里不再需要手动创建 time.AfterFunc，而是交给 PausableTimer 统一管理
	timeoutDuration := time.Duration(timeoutSeconds) * time.Second
	currentPlayerID := player.User.ID
	
	t.actionTimer.Start(timeoutDuration, func() {
		// 超时自动操作 (通常是 Check 或 Fold)
		// 再次检查是否仍然是该玩家的回合
		if t.CurrentHand != nil && t.CurrentHand.CurrentPlayerIndex != -1 {
			currentSeat := t.Seats[t.CurrentHand.CurrentPlayerIndex]
			if currentSeat.Player != nil && currentSeat.Player.User.ID == currentPlayerID {
				// 如果可以 Check 就 Check，否则 Fold
				autoAction := ActionTypeFold
				if callAmount == 0 {
					autoAction = ActionTypeCheck
				}
				
				// 标记玩家为托管/断线状态
				currentSeat.Player.IsOffline = true
				
				t.processPlayerAction(currentPlayerID, autoAction, 0)
			}
		}
	})
}

// processPlayerAction 处理玩家的具体下注/弃牌动作
func (t *Table) processPlayerAction(playerID string, action ActionType, amount int) error {
	// 停止当前的行动定时器
	t.actionTimer.Stop()

	seat := t.getSeatByUserID(playerID)
	if seat == nil || seat.Player == nil {
		return fmt.Errorf("player not found in seat")
	}
	player := seat.Player

	// 1. 校验动作是否合法
	callAmount := t.CurrentHand.CurrentBet - player.CurrentBet
	
	switch action {
	case ActionTypeFold:
		player.State = PlayerStateFolded
	case ActionTypeCheck:
		if callAmount > 0 {
			return fmt.Errorf("cannot check, must call %d", callAmount)
		}
	case ActionTypeCall:
		if callAmount == 0 {
			return fmt.Errorf("cannot call, no bet to call")
		}
		if player.Chips < callAmount {
			return fmt.Errorf("not enough chips to call")
		}
		player.Chips -= callAmount
		player.CurrentBet += callAmount
		t.CurrentHand.Pot += callAmount
	case ActionTypeBet:
		if t.CurrentHand.CurrentBet > 0 {
			return fmt.Errorf("cannot bet, already a bet, use raise")
		}
		if amount < t.BigBlind {
			return fmt.Errorf("bet amount must be at least big blind")
		}
		if player.Chips < amount {
			return fmt.Errorf("not enough chips to bet")
		}
		player.Chips -= amount
		player.CurrentBet += amount
		t.CurrentHand.Pot += amount
		t.CurrentHand.CurrentBet = amount
		t.CurrentHand.MinRaise = amount
	case ActionTypeRaise:
		if t.CurrentHand.CurrentBet == 0 {
			return fmt.Errorf("cannot raise, no bet yet, use bet")
		}
		minRaiseTotal := t.CurrentHand.CurrentBet + t.CurrentHand.MinRaise
		if amount < minRaiseTotal && amount != player.Chips+player.CurrentBet {
			return fmt.Errorf("raise amount must be at least %d", minRaiseTotal)
		}
		raiseAmount := amount - player.CurrentBet
		if player.Chips < raiseAmount {
			return fmt.Errorf("not enough chips to raise")
		}
		player.Chips -= raiseAmount
		player.CurrentBet += raiseAmount
		t.CurrentHand.Pot += raiseAmount
		
		// 更新 MinRaise: 新的下注额 - 旧的下注额
		t.CurrentHand.MinRaise = amount - t.CurrentHand.CurrentBet
		t.CurrentHand.CurrentBet = amount
	case ActionTypeAllIn:
		allInAmount := player.Chips
		player.Chips = 0
		player.CurrentBet += allInAmount
		t.CurrentHand.Pot += allInAmount
		player.State = PlayerStateAllIn
		
		// 如果 All-in 的金额大于当前最高下注额，更新最高下注额
		if player.CurrentBet > t.CurrentHand.CurrentBet {
			// 如果 All-in 构成了合法加注，更新 MinRaise
			raiseDiff := player.CurrentBet - t.CurrentHand.CurrentBet
			if raiseDiff >= t.CurrentHand.MinRaise {
				t.CurrentHand.MinRaise = raiseDiff
			}
			t.CurrentHand.CurrentBet = player.CurrentBet
		}
	default:
		return fmt.Errorf("unknown action: %s", action)
	}

	// 4. 标记该玩家本轮已行动
	player.HasActedThisRound = true

	// 5. 广播玩家的动作，用于前端播放动画
	actionInfo := &ActionInfo{
		PlayerID: playerID,
		Action:   action,
		Amount:   amount,
	}
	
	snap := t.BuildPublicSnapshot()
	snap.LastAction = actionInfo
	t.messenger.Broadcast(MsgTypeStateUpdate, "player_action", snap)

	// 6. 推进状态机
	t.advanceStateMachine()

	return nil
}

// advanceStateMachine 推进状态机（判断当前下注圈是否结束，是否进入下一阶段）
func (t *Table) advanceStateMachine() {
	// 1. 检查是否只剩下一个未弃牌的玩家
	activePlayers := 0
	var lastActivePlayer *Player
	for _, seatIdx := range t.CurrentHand.ActionOrder {
		player := t.Seats[seatIdx].Player
		if player.State != PlayerStateFolded {
			activePlayers++
			lastActivePlayer = player
		}
	}

	if activePlayers == 1 {
		// 提前结束本局
		t.earlyFinish(lastActivePlayer)
		return
	}

	// 2. 检查当前下注圈是否所有人都已表态，且下注金额已平齐
	allMatched := true
	for _, seatIdx := range t.CurrentHand.ActionOrder {
		player := t.Seats[seatIdx].Player
		if player.State == PlayerStateFolded || player.State == PlayerStateAllIn {
			continue
		}
		if !player.HasActedThisRound || player.CurrentBet < t.CurrentHand.CurrentBet {
			allMatched = false
			break
		}
	}

	if allMatched {
		// 进入下一阶段
		t.nextStage()
	} else {
		// 寻找下一个该说话的玩家
		currIdx := -1
		for i, seatIdx := range t.CurrentHand.ActionOrder {
			if seatIdx == t.CurrentHand.CurrentPlayerIndex {
				currIdx = i
				break
			}
		}

		for i := 1; i <= len(t.CurrentHand.ActionOrder); i++ {
			nextIdx := (currIdx + i) % len(t.CurrentHand.ActionOrder)
			nextSeatIdx := t.CurrentHand.ActionOrder[nextIdx]
			player := t.Seats[nextSeatIdx].Player
			if player.State != PlayerStateFolded && player.State != PlayerStateAllIn {
				t.CurrentHand.CurrentPlayerIndex = nextSeatIdx
				t.notifyCurrentPlayer(t.ActionTimeout)
				return
			}
		}
	}
}

func (t *Table) earlyFinish(winner *Player) {
	// 将所有下注收集到底池
	for _, seatIdx := range t.CurrentHand.ActionOrder {
		player := t.Seats[seatIdx].Player
		t.CurrentHand.Pot += player.CurrentBet
		player.CurrentBet = 0
	}
	
	winner.Chips += t.CurrentHand.Pot
	
	// 广播结算快照
	summary := &ShowdownSummary{
		BoardCards: t.CurrentHand.BoardCards,
		ShowCards:  false, // 提前结束不需要亮牌
		SidePots: []*SidePot{
			{
				PotNumber: 1,
				Amount:    t.CurrentHand.Pot,
				Winners:   []string{winner.User.ID},
				Players:   []string{winner.User.ID},
				IsMainPot: true,
			},
		},
		AllHands: []HandInfo{},
	}
	
	snap := t.BuildPublicSnapshot()
	snap.Stage = HandStageShowdown
	snap.ShowdownSummary = summary
	t.messenger.Broadcast(MsgTypeStateUpdate, "early_finish", snap)
	
	// 记录对局历史
	t.Histories = append(t.Histories, summary)
	
	// 延迟清理牌桌，给前端时间展示赢家收筹码的动画
	// 提前结束虽然不需要亮牌，但依然需要时间让前端把底池的筹码飞给赢家
	t.showdownTimer.Start(3*time.Second, func() {
	for _, seat := range t.Seats {
		if seat.Player != nil {
			p := seat.Player
			if p.Chips == 0 {
					p.Chips = t.InitialChips
					p.BuyInCount++
				}
				// 重置为 Waiting，需要重新准备
				p.State = PlayerStateWaiting
				p.HoleCards = nil
				p.CurrentBet = 0
				p.HasActedThisRound = false
			}
		}
		
		t.CurrentHand = nil
		
		// 广播一局彻底结束、等待玩家重新准备的状态
		t.messenger.Broadcast(MsgTypeStateUpdate, "hand_finished", t.BuildPublicSnapshot())
	})
}

// nextStage 进入下一个游戏阶段 (Flop, Turn, River, Showdown)
func (t *Table) nextStage() {
	// 1. 将所有玩家本轮的下注 (CurrentBet) 收集到底池 (Pot) 和边池 (SidePots) 中
	t.calculateSidePots()

	// 2. 重置所有玩家的 HasActedThisRound = false (CurrentBet 已经在 calculateSidePots 中重置)
	for _, seatIdx := range t.CurrentHand.ActionOrder {
		player := t.Seats[seatIdx].Player
		if player.State != PlayerStateFolded && player.State != PlayerStateAllIn {
			player.HasActedThisRound = false
		}
	}
	t.CurrentHand.CurrentBet = 0
	t.CurrentHand.MinRaise = t.BigBlind

	// 3. 根据当前 Stage 推进到下一个 Stage：
	switch t.CurrentHand.Stage {
	case HandStagePreFlop:
		t.CurrentHand.Stage = HandStageFlop
		cards, _ := t.CurrentHand.Deck.Draw(3)
		t.CurrentHand.BoardCards = append(t.CurrentHand.BoardCards, cards...)
	case HandStageFlop:
		t.CurrentHand.Stage = HandStageTurn
		cards, _ := t.CurrentHand.Deck.Draw(1)
		t.CurrentHand.BoardCards = append(t.CurrentHand.BoardCards, cards...)
	case HandStageTurn:
		t.CurrentHand.Stage = HandStageRiver
		cards, _ := t.CurrentHand.Deck.Draw(1)
		t.CurrentHand.BoardCards = append(t.CurrentHand.BoardCards, cards...)
	case HandStageRiver:
		t.CurrentHand.Stage = HandStageShowdown
		t.handleShowdown()
		return
	}

	// 4. 特殊情况：如果所有未弃牌玩家都已 All-in (或只剩一人未 All-in)
	activePlayers := 0
	notAllInPlayers := 0
	for _, seatIdx := range t.CurrentHand.ActionOrder {
		player := t.Seats[seatIdx].Player
		if player.State != PlayerStateFolded {
			activePlayers++
			if player.State != PlayerStateAllIn {
				notAllInPlayers++
			}
		}
	}

	if notAllInPlayers <= 1 {
		// 快进到 SHOWDOWN
		for t.CurrentHand.Stage != HandStageShowdown {
			// 发牌
			switch t.CurrentHand.Stage {
			case HandStageFlop:
				t.CurrentHand.Stage = HandStageTurn
				cards, _ := t.CurrentHand.Deck.Draw(1)
				t.CurrentHand.BoardCards = append(t.CurrentHand.BoardCards, cards...)
			case HandStageTurn:
				t.CurrentHand.Stage = HandStageRiver
				cards, _ := t.CurrentHand.Deck.Draw(1)
				t.CurrentHand.BoardCards = append(t.CurrentHand.BoardCards, cards...)
			case HandStageRiver:
				t.CurrentHand.Stage = HandStageShowdown
			}
		}
		
		// 广播快进的快照
		t.messenger.Broadcast(MsgTypeStateUpdate, "fast_forward", t.BuildPublicSnapshot())
		
		// 延迟一下再结算，让前端有时间播动画
		t.showdownTimer.Start(2*time.Second, func() {
			t.handleShowdown()
		})
		return
	}

	// 5. 重新计算下一轮的行动顺序 (通常从 SB 开始)
	// 在 Flop, Turn, River 阶段，行动从 SB 开始 (即 Button 左边第一个未弃牌的玩家)
	// 我们可以直接遍历 t.CurrentHand.ActionOrder，因为它是从 SB 开始排的，但是把 UTG 放到了前面 (PreFlop时)
	// 重新排序 ActionOrder，从 SB 开始
	
	// 收集所有在座的玩家，从 Button 左边开始
	orderedSeats := make([]int, 0)
	for i := 1; i <= t.MaxPlayers; i++ {
		seatIdx := (t.ButtonSeat + i) % t.MaxPlayers
		seat := t.Seats[seatIdx]
		if seat.Player != nil && 
		   (seat.Player.State == PlayerStateActive || seat.Player.State == PlayerStateAllIn) {
			orderedSeats = append(orderedSeats, seatIdx)
		}
	}
	t.CurrentHand.ActionOrder = orderedSeats

	// 找到第一个需要说话的玩家
	t.CurrentHand.CurrentPlayerIndex = -1
	for _, seatIdx := range t.CurrentHand.ActionOrder {
		player := t.Seats[seatIdx].Player
		if player.State == PlayerStateActive {
			t.CurrentHand.CurrentPlayerIndex = seatIdx
			break
		}
	}

	// 6. 广播进入新阶段的快照
	t.messenger.Broadcast(MsgTypeStateUpdate, "next_stage", t.BuildPublicSnapshot())

	// 7. 通知第一个玩家行动 (如果还没到 SHOWDOWN)
	if t.CurrentHand.CurrentPlayerIndex != -1 {
		t.notifyCurrentPlayer(t.ActionTimeout)
	}
}

// handleShowdown 处理摊牌和结算
func (t *Table) handleShowdown() {
	// 1. 收集所有最终的下注到底池/边池
	t.calculateSidePots()

	// 2. 评估所有未弃牌玩家的牌型大小
	allHands := make([]HandInfo, 0)
	playerRanks := make(map[string]HandRank)
	
	for _, seat := range t.Seats {
		if seat.Player != nil {
			p := seat.Player
			if p.State != PlayerStateFolded && p.State != PlayerStateWaiting {
				// TODO: 真正的牌型评估逻辑 (7选5)
				// 暂时用随机数模拟
				rank := HandRank(rand.Intn(10))
				playerRanks[p.User.ID] = rank
				
				allHands = append(allHands, HandInfo{
					PlayerID: p.User.ID,
					Cards:    p.HoleCards,
					HandRank: rank,
				})
			}
		}
	}

	// 3. 根据牌型大小分配各个边池
	for _, pot := range t.CurrentHand.SidePots {
		if pot.Amount == 0 {
			continue
		}
		
		// 找出这个奖池中牌型最大的玩家
		var maxRank HandRank = -1
		var winners []string
		
		for _, pid := range pot.Players {
			rank, ok := playerRanks[pid]
			if !ok {
				continue
			}
			if rank > maxRank {
				maxRank = rank
				winners = []string{pid}
			} else if rank == maxRank {
				winners = append(winners, pid)
			}
		}
		
		pot.Winners = winners
		pot.HandRank = maxRank
		
		// 平分奖池
		if len(winners) > 0 {
			winAmount := pot.Amount / len(winners)
			for _, wid := range winners {
				seat := t.getSeatByUserID(wid)
				if seat != nil {
					seat.Player.Chips += winAmount
				}
			}
			// 处理余数 (通常给位置最靠前的玩家，这里简单给第一个)
			remainder := pot.Amount % len(winners)
			if remainder > 0 {
				seat := t.getSeatByUserID(winners[0])
				if seat != nil {
					seat.Player.Chips += remainder
				}
			}
		}
	}

	// 4. 构建 ShowdownSummary
	summary := &ShowdownSummary{
		BoardCards: t.CurrentHand.BoardCards,
		ShowCards:  true,
		SidePots:   t.CurrentHand.SidePots,
		AllHands:   allHands,
	}

	// 5. 广播结算快照
	snap := t.BuildPublicSnapshot()
	snap.Stage = HandStageShowdown
	snap.ShowdownSummary = summary
	t.messenger.Broadcast(MsgTypeStateUpdate, "showdown", snap)

	// 6. 记录对局历史 (Histories)
	t.Histories = append(t.Histories, summary)

	// 7. 延迟几秒后，清理牌桌，状态重置为 WAITING
	// 这里的延迟是为了让前端有时间播放：
	// 1. 玩家亮出底牌的动画
	// 2. 筹码从底池飞向赢家的动画
	// 3. 弹出结算面板展示各家牌型
	// 如果没有这个延迟，前端收到 showdown 消息后瞬间又会收到 hand_finished 消息，
	// 导致牌桌瞬间被清空，玩家看不清结算结果。
	t.showdownTimer.Start(5*time.Second, func() {
		// 8. 破产清理与筹码补充，并重置准备状态
	for _, seat := range t.Seats {
		if seat.Player != nil {
			p := seat.Player
			if p.Chips == 0 {
					p.Chips = t.InitialChips
					p.BuyInCount++
				}
				// 游戏结束后，所有玩家状态重置为 Waiting，需要重新点击准备
				p.State = PlayerStateWaiting
				p.HoleCards = nil
				p.CurrentBet = 0
				p.HasActedThisRound = false
			}
		}
		
		t.CurrentHand = nil
		
		// 广播一局彻底结束、等待玩家重新准备的状态
		t.messenger.Broadcast(MsgTypeStateUpdate, "hand_finished", t.BuildPublicSnapshot())
	})
}
