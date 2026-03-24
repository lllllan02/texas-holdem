package pot

import "slices"

// PlayerBet 玩家在当前局的下注状态
type PlayerBet struct {
	PlayerID string
	Bet      int  // 本局总共下注的筹码
	IsFolded bool // 是否已经弃牌
}

// Pot 一个奖池（主池或边池）
type Pot struct {
	Amount  int      // 奖池总额
	Players []string // 有资格分这个池子的玩家 ID 列表
}

// CalculatePots 根据所有玩家的下注情况，计算出当前的主池和所有边池
func CalculatePots(bets []PlayerBet) []Pot {
	if len(bets) == 0 {
		return nil
	}

	// 复制一份数据，避免修改原数组
	activeBets := make([]PlayerBet, len(bets))
	copy(activeBets, bets)

	var pots []Pot
	for {
		// 找出当前未弃牌玩家中最小的非零下注额
		minBet := 0
		for _, b := range activeBets {
			if b.Bet > 0 && !b.IsFolded {
				if minBet == 0 || b.Bet < minBet {
					minBet = b.Bet
				}
			}
		}

		// 如果所有下注都为 0，说明池子已经分完
		if minBet == 0 {
			break
		}

		// 创建一个新的奖池
		currentPot := Pot{
			Amount:  0,
			Players: []string{},
		}

		// 收集有资格的玩家（非弃牌，且当前有下注）
		// 注意：即使玩家弃牌，他们的死钱也会进入当前池子，但他们没有资格分钱
		for i, b := range activeBets {
			if b.Bet > 0 {
				// 从玩家下注中扣除 minBet
				contribution := min(b.Bet, minBet)

				currentPot.Amount += contribution
				activeBets[i].Bet -= contribution

				// 如果玩家没有弃牌，就有资格分这个池子
				if !b.IsFolded {
					currentPot.Players = append(currentPot.Players, b.PlayerID)
				}
			}
		}

		// 如果池子有钱，加入结果列表
		if currentPot.Amount > 0 {
			pots = append(pots, currentPot)
		}
	}

	// 处理可能残留的死钱 (Dead Money)
	// 比如：P1(all-in 100), P2(bet 200 弃牌)。P1 只能 match 100，P2 多出的 100 在上述逻辑中会残留
	// 这些残留的钱应该归入最后一个产生的奖池中
	if len(pots) > 0 {
		for _, b := range activeBets {
			pots[len(pots)-1].Amount += b.Bet
		}
	}

	return pots
}

// DistributePot 将一个奖池的钱分给赢家（处理平局平分及除不尽的零头）
func DistributePot(pot Pot, winners []string) map[string]int {
	result := make(map[string]int)
	if len(winners) == 0 || pot.Amount == 0 {
		return result
	}

	// 过滤出真正在 Players 中的赢家
	eligibleWinners := []string{}
	for _, w := range winners {
		if slices.Contains(pot.Players, w) {
			eligibleWinners = append(eligibleWinners, w)
		}
	}

	if len(eligibleWinners) == 0 {
		return result
	}

	// 平分奖池
	baseShare := pot.Amount / len(eligibleWinners)
	remainder := pot.Amount % len(eligibleWinners)

	for i, w := range eligibleWinners {
		share := baseShare
		// 零头给位置靠前的玩家（这里简单按 winners 数组顺序给）
		if i < remainder {
			share++
		}
		result[w] += share
	}

	return result
}
