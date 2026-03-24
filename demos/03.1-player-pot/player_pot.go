package playerpot

import (
	core "github.com/lllllan02/texas-holdem/demos/01-core"
	evaluator "github.com/lllllan02/texas-holdem/demos/02-evaluator"
)

// Player 融合了手牌和下注概念的完整玩家对象
type Player struct {
	ID        string
	HoleCards []core.Card // 玩家的2张底牌
	Bet       int         // 本局总共下注的筹码
	IsFolded  bool        // 是否已经弃牌

	// 缓存属性，避免多次比牌时重复计算
	BestHand  evaluator.HandResult
	Evaluated bool
}

// EvaluateHand 计算并缓存玩家结合公共牌后的最佳牌型
func (p *Player) EvaluateHand(publicCards []core.Card) {
	if p.Evaluated {
		return
	}
	// 合并手牌和公共牌
	allCards := make([]core.Card, 0, len(p.HoleCards)+len(publicCards))
	allCards = append(allCards, p.HoleCards...)
	allCards = append(allCards, publicCards...)

	p.BestHand = evaluator.Evaluate(allCards)
	p.Evaluated = true
}

// Pot 一个奖池（主池或边池）
type Pot struct {
	Amount  int
	Players []*Player // 有资格分这个池子的玩家列表
}

// CalculatePots 根据所有玩家的下注情况，计算出当前的主池和所有边池
func CalculatePots(players []*Player) []Pot {
	if len(players) == 0 {
		return nil
	}

	// 内部结构，用于在计算过程中扣减下注额，而不修改原 player 的 Bet 属性
	type activeBet struct {
		player *Player
		bet    int
	}

	activeBets := make([]activeBet, len(players))
	for i, p := range players {
		activeBets[i] = activeBet{player: p, bet: p.Bet}
	}

	var pots []Pot

	for {
		// 找出当前未弃牌玩家中最小的非零下注额
		minBet := 0
		for _, b := range activeBets {
			if b.bet > 0 && !b.player.IsFolded {
				if minBet == 0 || b.bet < minBet {
					minBet = b.bet
				}
			}
		}

		// 如果所有下注都为 0，说明池子已经分完
		if minBet == 0 {
			break
		}

		currentPot := Pot{
			Amount:  0,
			Players: []*Player{},
		}

		for i, b := range activeBets {
			if b.bet > 0 {
				contribution := min(b.bet, minBet)

				currentPot.Amount += contribution
				activeBets[i].bet -= contribution

				if !b.player.IsFolded {
					currentPot.Players = append(currentPot.Players, b.player)
				}
			}
		}

		if currentPot.Amount > 0 {
			pots = append(pots, currentPot)
		}
	}

	// 处理可能残留的死钱 (Dead Money)
	if len(pots) > 0 {
		for _, b := range activeBets {
			pots[len(pots)-1].Amount += b.bet
		}
	}

	return pots
}

// ResolveGame 游戏结算的终极方法：计算奖池并根据牌力分配奖金
// 返回一个 map，记录每个 PlayerID 最终分得的筹码数量
func ResolveGame(players []*Player, publicCards []core.Card) map[string]int {
	pots := CalculatePots(players)
	finalPayouts := make(map[string]int)

	for _, pot := range pots {
		if len(pot.Players) == 0 {
			continue
		}

		var winners []*Player
		var bestResult evaluator.HandResult

		// 1. 在有资格分该奖池的玩家中进行比牌
		for i, p := range pot.Players {
			p.EvaluateHand(publicCards)

			if i == 0 {
				bestResult = p.BestHand
				winners = []*Player{p}
				continue
			}

			cmp := evaluator.Compare(p.BestHand, bestResult)
			if cmp > 0 {
				bestResult = p.BestHand
				winners = []*Player{p}
			} else if cmp == 0 {
				winners = append(winners, p)
			}
		}

		// 2. 将当前奖池的钱平分给赢家
		if len(winners) > 0 {
			baseShare := pot.Amount / len(winners)
			remainder := pot.Amount % len(winners)

			for i, w := range winners {
				share := baseShare
				if i < remainder {
					share++ // 零头给位置靠前的玩家
				}
				finalPayouts[w.ID] += share
			}
		}
	}

	return finalPayouts
}
