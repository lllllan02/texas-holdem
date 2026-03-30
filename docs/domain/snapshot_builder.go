package domain

// BuildSnapshot 为特定的观察者（玩家或旁观者）生成千人千面的状态快照
func (t *Table) BuildSnapshot(viewerID string) StateUpdateSnapshot {
	snap := StateUpdateSnapshot{
		HandCount:  t.HandCount,
		ButtonSeat: t.ButtonSeat,
		Players:    make([]PlayerSnapshot, 0),
	}

	// 1. 映射单局状态 (如果正在游戏中的话)
	if t.CurrentHand != nil {
		snap.Stage = t.CurrentHand.Stage
		snap.Pot = t.CurrentHand.Pot
		snap.CurrentBet = t.CurrentHand.CurrentBet
		snap.MinRaise = t.CurrentHand.MinRaise
		snap.BoardCards = t.CurrentHand.BoardCards
		snap.CurrentPlayerIndex = t.CurrentHand.CurrentPlayerIndex
		snap.ActionOrder = t.CurrentHand.ActionOrder
		snap.SidePots = t.CurrentHand.SidePots
	} else {
		// 如果没有正在进行的牌局，说明在等待阶段
		snap.Stage = "WAITING"
		snap.CurrentPlayerIndex = -1
	}

	// 2. 映射玩家列表并进行【数据脱敏】
	for _, seat := range t.Seats {
		if seat.State == SeatEmpty || seat.Player == nil {
			continue
		}

		p := seat.Player
		pSnap := PlayerSnapshot{
			ID:                p.ID,
			Name:              p.Name,
			Position:          p.Position,
			SeatNumber:        seat.SeatNumber,
			Chips:             p.Chips,
			CurrentBet:        p.CurrentBet,
			State:             p.State,
			HasActedThisRound: p.HasActedThisRound,
		}

		// 【核心安全逻辑】：千人千面，只发该看的底牌
		// 什么时候能看到底牌？
		// 1. 观察者就是玩家本人
		// 2. 游戏进入了 SHOWDOWN 摊牌阶段（所有人都能看）
		if p.ID == viewerID || snap.Stage == StageShowdown {
			pSnap.HoleCards = p.HoleCards
		} else {
			pSnap.HoleCards = nil // 强制打码，防止透视挂
		}

		snap.Players = append(snap.Players, pSnap)
	}

	return snap
}
