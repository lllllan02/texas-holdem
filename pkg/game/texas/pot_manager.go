package texas

import (
	"sort"
)

// PlayerBet 记录玩家在当前下注圈的下注额
type PlayerBet struct {
	PlayerID string
	Bet      int
	IsAllIn  bool
}

// calculateSidePots 计算并更新边池
// 在每个下注圈结束时（进入下一阶段前），将所有玩家的 CurrentBet 收集并拆分到 SidePots 中
func (t *Table) calculateSidePots() {
	if t.CurrentHand == nil {
		return
	}

	// 1. 收集所有玩家本轮的下注
	var bets []PlayerBet
	for _, p := range t.Seats {
		if p != nil {

			if p.CurrentBet > 0 {
				bets = append(bets, PlayerBet{
					PlayerID: p.User.ID,
					Bet:      p.CurrentBet,
					IsAllIn:  p.State == PlayerStateAllIn,
				})
			}
		}
	}

	if len(bets) == 0 {
		return
	}

	// 2. 按下注金额从小到大排序
	sort.Slice(bets, func(i, j int) bool {
		return bets[i].Bet < bets[j].Bet
	})

	// 3. 逐层剥离并分配到底池/边池
	for len(bets) > 0 {
		minBet := bets[0].Bet
		if minBet == 0 {
			bets = bets[1:]
			continue
		}

		// 找出当前层级有资格竞争的玩家
		var eligiblePlayers []string
		for _, b := range bets {
			// 只有未弃牌的玩家才有资格竞争
			seatIdx := t.getSeatIndexByUserID(b.PlayerID)
			if seatIdx != -1 && t.Seats[seatIdx].State != PlayerStateFolded {
				eligiblePlayers = append(eligiblePlayers, b.PlayerID)
			}
		}

		// 计算当前层级的奖池金额
		potAmount := 0
		for i := range bets {
			potAmount += minBet
			bets[i].Bet -= minBet
		}

		// 找到或创建合适的 SidePot
		// 如果最后一个 SidePot 的 eligiblePlayers 和当前一致，则合并；否则创建新的
		var targetPot *SidePot
		if len(t.CurrentHand.SidePots) > 0 {
			lastPot := t.CurrentHand.SidePots[len(t.CurrentHand.SidePots)-1]
			if isSamePlayers(lastPot.Players, eligiblePlayers) {
				targetPot = lastPot
			}
		}

		if targetPot != nil {
			targetPot.Amount += potAmount
		} else {
			isMain := len(t.CurrentHand.SidePots) == 0
			newPot := &SidePot{
				PotNumber: len(t.CurrentHand.SidePots) + 1,
				Amount:    potAmount,
				Players:   eligiblePlayers,
				IsMainPot: isMain,
			}
			t.CurrentHand.SidePots = append(t.CurrentHand.SidePots, newPot)
		}

		// 移除已经扣减完的 bet
		var nextBets []PlayerBet
		for _, b := range bets {
			if b.Bet > 0 {
				nextBets = append(nextBets, b)
			}
		}
		bets = nextBets
	}

	// 4. 重置玩家的 CurrentBet
	for _, p := range t.Seats {
		if p != nil {
			p.CurrentBet = 0
		}
	}
}

func isSamePlayers(p1, p2 []string) bool {
	if len(p1) != len(p2) {
		return false
	}
	m := make(map[string]bool)
	for _, p := range p1 {
		m[p] = true
	}
	for _, p := range p2 {
		if !m[p] {
			return false
		}
	}
	return true
}
