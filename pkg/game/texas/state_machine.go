package texas

import (
	"fmt"
	"log"
	"time"
)

// 核心状态机流转 (State Machine Flow)

// startNewHand 开始新的一局游戏
func (t *Table) startNewHand() error {
	// 1. 从庄家下一席起顺时针扫一整圈，收集本局所有 Ready 的玩家，并直接重置其本局状态。
	//    activeSeats[0] 即为新庄家（上一局 Button 的下一顺时针、仍在桌上的第一个选手）。
	//    ButtonSeat 为 -1 时等价于从 0 号位开始找第一个 Ready。
	n := len(t.Seats)
	if n < 2 {
		return fmt.Errorf("not enough seats to start a hand")
	}

	// 1. 推进庄家 (ButtonSeat)
	// 当前开局要求是严格满员，所以直接顺时针移动一位即可
	// 如果是第一局 (ButtonSeat == -1)，则 (-1 + 1) % n = 0，即 0 号位为庄家
	t.ButtonSeat = (t.ButtonSeat + 1) % n

	// 2. 初始化 Hand 结构体
	t.CurrentHand = &Hand{
		ID:                 fmt.Sprintf("hand-%d", time.Now().UnixNano()),
		Stage:              HandStagePreFlop,
		Deck:               NewDeck(),
		BoardCards:         make([]Card, 0, 5),
		Pot:                0,
		SidePots:           make([]*SidePot, 0),
		CurrentBet:         0,
		MinRaise:           t.BigBlind,
		ActionOrder:        make([]int, n),
		CurrentPlayerIndex: -1,
		HandCount:          len(t.Histories) + 1,
	}

	log.Printf("TexasEngine: === Hand #%d Started ===", t.CurrentHand.HandCount)

	// 3. 洗牌
	t.CurrentHand.Deck.Shuffle()

	// 4. 顺时针遍历所有座位（从 SB 开始，即 BTN 的下一位）
	// 在一个循环中完成：校验满员/准备状态、重置状态、发牌、排定行动顺序
	for i := 1; i <= n; i++ {
		seatIdx := (t.ButtonSeat + i) % n
		p := t.Seats[seatIdx]

		// 严格满员校验：中间发现空座或未准备，直接报错并回滚
		if p == nil || p.State != PlayerStateReady {
			t.CurrentHand = nil
			return fmt.Errorf("all seats must be occupied and ready to start (seat %d is not ready)", seatIdx)
		}

		// 初始化玩家本局状态
		p.State = PlayerStateActive
		p.ChipsBeforeHand = p.Chips
		p.CurrentBet = 0
		p.HasActedThisRound = false

		// 发牌：从小盲 (i=1) 开始顺时针发牌
		cards, err := t.CurrentHand.Deck.Draw(2)
		if err != nil {
			t.CurrentHand = nil
			return fmt.Errorf("failed to draw cards: %w", err)
		}
		p.HoleCards = cards

		// 确定翻牌前行动顺序：
		// 统一规则：BTN 后一位是 SB(i=1)，再后一位是 BB(i=2)。
		// 翻牌前从 BB 的下一位 (i=3) 开始行动。
		// 所以 i=3 对应 ActionOrder[0]，公式为 (i - 3 + n) % n
		actionIdx := (i - 3 + n) % n
		t.CurrentHand.ActionOrder[actionIdx] = seatIdx
	}

	// 5. 强制扣除小盲注和大盲注
	sbSeat := (t.ButtonSeat + 1) % n
	bbSeat := (t.ButtonSeat + 2) % n

	sbPlayer := t.Seats[sbSeat]
	bbPlayer := t.Seats[bbSeat]

	sbAmount := sbPlayer.PlaceBet(t.SmallBlind)
	t.CurrentHand.Pot += sbAmount

	bbAmount := bbPlayer.PlaceBet(t.BigBlind)
	t.CurrentHand.Pot += bbAmount

	// 当前街「封顶注」= 完整的大盲金额
	// 即使大盲玩家是短码（不够一个大盲并全下），后续玩家跟注的目标依然是完整的大盲金额
	t.CurrentHand.CurrentBet = t.BigBlind

	// 7. 确定第一个说话的玩家
	t.CurrentHand.CurrentPlayerIndex = t.CurrentHand.ActionOrder[0]

	// 8. 给每个玩家私发专属快照（包含他们自己的底牌）
	for _, seatIdx := range t.CurrentHand.ActionOrder {
		player := t.Seats[seatIdx]
		// 生成专属快照，其中会包含该玩家的 HoleCards，其他人的 HoleCards 为 nil
		snap := t.BuildPersonalSnapshot(player.User.ID)
		log.Printf("TexasEngine: Dealt hole cards to Player [%s]", player.User.ID)
		t.messenger.SendTo(player.User.ID, MsgTypeStateUpdate, ReasonDealHoleCards, snap)
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

	player := t.Seats[t.CurrentHand.CurrentPlayerIndex]

	// 如果玩家已经 All-in 或 Fold，理论上不应该被选为 CurrentPlayerIndex，
	// 因为我们在 advanceStateMachine 中已经严格跳过了他们。
	// 如果由于某种异常发生了，直接跳过他，推进状态机。
	// 这里不再广播假动作，因为这属于异常拦截，正常流程不会走到这里。
	if player.State == PlayerStateFolded || player.State == PlayerStateAllIn {
		t.advanceStateMachine()
		return
	}

	// 计算可用的动作和金额限制
	callAmount := t.CurrentHand.CurrentBet - player.CurrentBet

	var validActions []ActionType

	details := ActionDetails{}

	if callAmount == 0 {
		validActions = append(validActions, ActionTypeCheck)
	} else {
		validActions = append(validActions, ActionTypeFold)
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

	// 如果玩家已离线（托管状态），为了不让其他玩家干等，将思考时间缩短为配置的 OfflineTimeout
	actualTimeout := timeoutSeconds
	if player.IsOffline && actualTimeout > t.OfflineTimeout {
		actualTimeout = t.OfflineTimeout
	}

	payload := TurnNotificationPayload{
		PlayerID:       player.User.ID,
		ValidActions:   validActions,
		ActionDetails:  details,
		TimeoutSeconds: actualTimeout,
	}

	// 广播行动通知，告诉所有人轮到谁了，以及倒计时是多少
	t.messenger.Broadcast(MsgTypeTurnNotification, "turn_notification", payload)

	// 启动玩家行动超时定时器
	// 注意：这里不再需要手动创建 time.AfterFunc，而是交给 PausableTimer 统一管理
	timeoutDuration := time.Duration(actualTimeout) * time.Second
	currentPlayerID := player.User.ID

	t.actionTimer.Start(timeoutDuration, func() {
		t.messenger.Execute(func() {
			// 超时自动操作 (通常是 Check 或 Fold)
			// 再次检查是否仍然是该玩家的回合
			if t.CurrentHand != nil && t.CurrentHand.CurrentPlayerIndex != -1 {
				currentPlayer := t.Seats[t.CurrentHand.CurrentPlayerIndex]
				if currentPlayer != nil && currentPlayer.User.ID == currentPlayerID {
					// 如果可以 Check 就 Check，否则 Fold
					autoAction := ActionTypeFold
					if callAmount == 0 {
						autoAction = ActionTypeCheck
					}

					// 标记玩家为托管/断线状态
					currentPlayer.IsOffline = true

					t.processPlayerAction(currentPlayerID, autoAction, 0)
				}
			}
		})
	})
}

// processPlayerAction 处理玩家的具体下注/弃牌动作
func (t *Table) processPlayerAction(playerID string, action ActionType, amount int) error {
	seatIdx := t.getSeatIndexByUserID(playerID)
	if seatIdx == -1 {
		return fmt.Errorf("player not found in seat")
	}
	player := t.Seats[seatIdx]

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
		// 使用 AddBet 自动处理筹码扣除、底池增加及 All-in 状态
		t.CurrentHand.AddBet(player, callAmount)
	case ActionTypeBet:
		if t.CurrentHand.CurrentBet > 0 {
			return fmt.Errorf("cannot bet, already a bet, use raise")
		}
		if amount < t.BigBlind {
			return fmt.Errorf("bet amount must be at least big blind")
		}
		t.CurrentHand.AddBet(player, amount)
	case ActionTypeRaise:
		if t.CurrentHand.CurrentBet == 0 {
			return fmt.Errorf("cannot raise, no bet yet, use bet")
		}
		minRaiseTotal := t.CurrentHand.CurrentBet + t.CurrentHand.MinRaise
		if amount < minRaiseTotal && amount != player.Chips+player.CurrentBet {
			return fmt.Errorf("raise amount must be at least %d", minRaiseTotal)
		}
		raiseAmount := amount - player.CurrentBet
		t.CurrentHand.AddBet(player, raiseAmount)
	case ActionTypeAllIn:
		t.CurrentHand.AddBet(player, player.Chips)
	default:
		return fmt.Errorf("unknown action: %s", action)
	}

	// 动作校验通过并执行成功，此时停止该玩家的行动倒计时
	t.actionTimer.Stop()

	// 4. 标记该玩家本轮已行动
	player.HasActedThisRound = true

	// 5. 广播玩家的动作，用于前端播放动画
	actionInfo := &ActionInfo{
		PlayerID: playerID,
		Action:   action,
		Amount:   amount,
	}

	log.Printf("TexasEngine: Player [%s] acted: %s (amount: %d)", playerID, action, amount)

	snap := t.BuildPublicSnapshot()
	snap.LastAction = actionInfo
	t.messenger.Broadcast(MsgTypeStateUpdate, ReasonPlayerAction, snap)

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
		player := t.Seats[seatIdx]
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
		player := t.Seats[seatIdx]
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
			player := t.Seats[nextSeatIdx]

			// 只有未弃牌、未 All-in，且（未行动过 或 下注额不足）的玩家才需要行动
			if player.State != PlayerStateFolded && player.State != PlayerStateAllIn {
				if !player.HasActedThisRound || player.CurrentBet < t.CurrentHand.CurrentBet {
					t.CurrentHand.CurrentPlayerIndex = nextSeatIdx
					t.notifyCurrentPlayer(t.ActionTimeout)
					return
				}
			}
		}

		// 如果循环一圈都没找到需要行动的人（比如大家都在之前 All-in 了），直接进入下一阶段
		t.nextStage()
	}
}

func (t *Table) earlyFinish(winner *Player) {
	// 1. 将所有未收集的下注收集到 SidePots 中
	t.calculateSidePots()

	// 2. 赢家赢走所有的池子 (包括主池和之前阶段产生的边池)
	totalWin := 0
	for _, pot := range t.CurrentHand.SidePots {
		totalWin += pot.Amount
		pot.Winners = []string{winner.User.ID}
		pot.HandRank = HandRankHighCard // 提前结束不需要比牌
	}
	winner.Chips += totalWin

	// 3. 广播结算快照
	summary := t.buildShowdownSummary(false)

	log.Printf("TexasEngine: Hand #%d finished early. Winner: [%s]", t.CurrentHand.HandCount, winner.User.ID)

	snap := t.BuildPublicSnapshot()
	snap.Stage = HandStageShowdown
	snap.ShowdownSummary = summary
	t.messenger.Broadcast(MsgTypeStateUpdate, ReasonEarlyFinish, snap)

	// 4. 记录对局历史
	t.Histories = append(t.Histories, summary)

	// 5. 延迟清理牌桌，给前端时间展示赢家收筹码的动画
	t.showdownTimer.Start(3*time.Second, func() {
		t.messenger.Execute(func() {
			for seatIdx, p := range t.Seats {
				if p != nil {

					// 【关键修复】：如果玩家已断线，强制其站起，腾出座位，否则会永远阻塞下一局的开始
					if p.IsOffline {
						t.Seats[seatIdx] = nil
						p.State = PlayerStateWaiting
						continue
					}

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
			log.Printf("TexasEngine: === Hand #%d Finished ===", t.CurrentHand.HandCount)
			t.messenger.Broadcast(MsgTypeStateUpdate, ReasonHandFinished, t.BuildPublicSnapshot())
		})
	})
}

// nextStage 进入下一个游戏阶段 (Flop, Turn, River, Showdown)
func (t *Table) nextStage() {
	// 1. 将所有玩家本轮的下注 (CurrentBet) 收集到底池 (Pot) 和边池 (SidePots) 中
	t.calculateSidePots()

	// 2. 重置所有玩家的 HasActedThisRound = false (CurrentBet 已经在 calculateSidePots 中重置)
	for _, seatIdx := range t.CurrentHand.ActionOrder {
		player := t.Seats[seatIdx]
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
		player := t.Seats[seatIdx]
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
		t.messenger.Broadcast(MsgTypeStateUpdate, ReasonFastForward, t.BuildPublicSnapshot())

		// 延迟一下再结算，让前端有时间播动画
		t.showdownTimer.Start(2*time.Second, func() {
			t.messenger.Execute(func() {
				t.handleShowdown()
			})
		})
		return
	}

	// 5. 重新计算下一轮的行动顺序 (通常从 SB 开始)
	// 在 Flop, Turn, River 阶段，行动从 SB 开始 (即 Button 左边第一个未弃牌的玩家)
	// 我们可以直接遍历 t.CurrentHand.ActionOrder，因为它是从 SB 开始排的，但是把 UTG 放到了前面 (PreFlop时)
	// 重新排序 ActionOrder，从 SB 开始

	// 收集所有在座的玩家，从 Button 左边开始
	orderedSeats := make([]int, 0)
	for i := 1; i <= len(t.Seats); i++ {
		seatIdx := (t.ButtonSeat + i) % len(t.Seats)
		player := t.Seats[seatIdx]
		if player != nil &&
			(player.State == PlayerStateActive || player.State == PlayerStateAllIn) {
			orderedSeats = append(orderedSeats, seatIdx)
		}
	}
	t.CurrentHand.ActionOrder = orderedSeats

	// 找到第一个需要说话的玩家
	t.CurrentHand.CurrentPlayerIndex = -1
	for _, seatIdx := range t.CurrentHand.ActionOrder {
		player := t.Seats[seatIdx]
		if player.State == PlayerStateActive {
			t.CurrentHand.CurrentPlayerIndex = seatIdx
			break
		}
	}

	// 6. 广播进入新阶段的快照
	log.Printf("TexasEngine: Advancing to stage: %s", t.CurrentHand.Stage)
	t.messenger.Broadcast(MsgTypeStateUpdate, ReasonNextStage, t.BuildPublicSnapshot())

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
	playerRanks := make(map[string]*HandResult)

	for _, p := range t.Seats {
		if p != nil {

			if p.State != PlayerStateFolded && p.State != PlayerStateWaiting {
				// 真正的牌型评估逻辑 (7选5)
				p.EvaluateHand(t.CurrentHand.BoardCards)
				playerRanks[p.User.ID] = p.BestHand
			}
		}
	}

	// 3. 根据牌型大小分配各个边池
	for _, pot := range t.CurrentHand.SidePots {
		if pot.Amount == 0 {
			continue
		}

		// 找出这个奖池中牌型最大的玩家
		var bestResult *HandResult
		var winners []string

		for _, pid := range pot.Players {
			res, ok := playerRanks[pid]
			if !ok {
				continue
			}

			if bestResult == nil {
				bestResult = res
				winners = []string{pid}
				continue
			}

			cmp := Compare(*res, *bestResult)
			if cmp > 0 {
				bestResult = res
				winners = []string{pid}
			} else if cmp == 0 {
				winners = append(winners, pid)
			}
		}

		pot.Winners = winners
		if bestResult != nil {
			pot.HandRank = bestResult.Rank
		}

		// 平分奖池
		if len(winners) > 0 {
			winAmount := pot.Amount / len(winners)
			for _, wid := range winners {
				seatIdx := t.getSeatIndexByUserID(wid)
				if seatIdx != -1 {
					t.Seats[seatIdx].Chips += winAmount
				}
			}
			// 处理余数 (通常给位置最靠前的玩家，这里简单给第一个)
			remainder := pot.Amount % len(winners)
			if remainder > 0 {
				seatIdx := t.getSeatIndexByUserID(winners[0])
				if seatIdx != -1 {
					t.Seats[seatIdx].Chips += remainder
				}
			}
		}
	}

	// 4. 构建 ShowdownSummary
	summary := t.buildShowdownSummary(true)

	// 5. 广播结算快照
	snap := t.BuildPublicSnapshot()
	snap.Stage = HandStageShowdown
	snap.ShowdownSummary = summary

	log.Printf("TexasEngine: Showdown reached for Hand #%d", t.CurrentHand.HandCount)
	t.messenger.Broadcast(MsgTypeStateUpdate, ReasonShowdown, snap)

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
		t.messenger.Execute(func() {
			// 8. 破产清理、离线清理与重置准备状态
			for seatIdx, p := range t.Seats {
				if p != nil {

					// 【关键修复】：如果玩家已断线，强制其站起，腾出座位，否则会永远阻塞下一局的开始
					if p.IsOffline {
						t.Seats[seatIdx] = nil
						p.State = PlayerStateWaiting
						continue
					}

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
			log.Printf("TexasEngine: === Hand #%d Finished ===", t.CurrentHand.HandCount)
			t.messenger.Broadcast(MsgTypeStateUpdate, ReasonHandFinished, t.BuildPublicSnapshot())
		})
	})
}
