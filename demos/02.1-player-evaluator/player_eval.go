package playereval

import (
	core "github.com/lllllan02/texas-holdem/demos/01-core"
	evaluator "github.com/lllllan02/texas-holdem/demos/02-evaluator"
)

// Player 代表一个参与比牌的玩家
type Player struct {
	ID        string
	HoleCards []core.Card          // 玩家的2张底牌
	BestHand  evaluator.HandResult // 缓存的最佳牌型
	Evaluated bool                 // 是否已经计算过最佳牌型
}

// EvaluateHand 计算并缓存玩家结合公共牌后的最佳牌型
// 结果保存在结构体的 BestHand 字段中，避免多次比牌时重复计算
func (p *Player) EvaluateHand(publicCards []core.Card) {
	// 如果已经计算过，直接返回
	if p.Evaluated {
		return
	}

	// 合并手牌 (2张) 和公共牌 (5张)
	allCards := make([]core.Card, 0, len(p.HoleCards)+len(publicCards))
	allCards = append(allCards, p.HoleCards...)
	allCards = append(allCards, publicCards...)

	// 调用 demo 2 的 Evaluate 方法找出 7 选 5 的最佳牌型
	p.BestHand = evaluator.Evaluate(allCards)
	p.Evaluated = true
}

// FindWinners 在一组玩家中找出获胜者（可能有多人平局平分），并返回赢家的最佳牌型
func FindWinners(players []*Player, publicCards []core.Card) ([]*Player, evaluator.HandResult) {
	var winners []*Player
	var bestResult evaluator.HandResult

	for i, p := range players {
		// 计算当前玩家的最佳牌型 (利用了内部的缓存机制)
		p.EvaluateHand(publicCards)
		res := p.BestHand

		if i == 0 {
			bestResult = res
			winners = []*Player{p}
			continue
		}

		// 与当前已知的最大牌型进行比较
		cmp := evaluator.Compare(res, bestResult)
		if cmp > 0 {
			// 发现更大的牌型，清空之前的赢家列表，更新最大牌型
			bestResult = res
			winners = []*Player{p}
		} else if cmp == 0 {
			// 牌型完全一样，加入平局赢家列表
			winners = append(winners, p)
		}
	}

	return winners, bestResult
}
