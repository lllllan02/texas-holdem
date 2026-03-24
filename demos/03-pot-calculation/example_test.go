package pot_test

import (
	"fmt"

	core "github.com/lllllan02/texas-holdem/demos/01-core"
	evaluator "github.com/lllllan02/texas-holdem/demos/02-evaluator"
	pot "github.com/lllllan02/texas-holdem/demos/03-pot-calculation"
)

// Example_gameResolution 演示了如何将“奖池计算”与“牌力评估”结合起来，
// 完美处理“短码玩家 All-in 赢了主池，其他玩家继续争夺边池”的经典场景。
func Example_gameResolution() {
	// 1. 设置公共牌 (5张)
	publicCards := []core.Card{
		{Suit: core.Spades, Rank: core.Rank10},
		{Suit: core.Hearts, Rank: core.Rank9},
		{Suit: core.Diamonds, Rank: core.Rank8},
		{Suit: core.Clubs, Rank: core.Rank2},
		{Suit: core.Spades, Rank: core.Rank3},
	}

	// 2. 设置玩家手牌
	// 玩家 A (短码 All-in): 一对 A (全场最大)
	cardsA := []core.Card{
		{Suit: core.Spades, Rank: core.RankA},
		{Suit: core.Hearts, Rank: core.RankA},
	}
	// 玩家 B: 一对 K (第二大)
	cardsB := []core.Card{
		{Suit: core.Spades, Rank: core.RankK},
		{Suit: core.Hearts, Rank: core.RankK},
	}
	// 玩家 C: 一对 Q (最小)
	cardsC := []core.Card{
		{Suit: core.Spades, Rank: core.RankQ},
		{Suit: core.Hearts, Rank: core.RankQ},
	}

	playerCards := map[string][]core.Card{
		"A": cardsA,
		"B": cardsB,
		"C": cardsC,
	}

	// 3. 设置玩家下注情况
	// A 只有 100 筹码 All-in，B 和 C 各下注 200
	bets := []pot.PlayerBet{
		{PlayerID: "A", Bet: 100},
		{PlayerID: "B", Bet: 200},
		{PlayerID: "C", Bet: 200},
	}

	// 4. 计算奖池
	pots := pot.CalculatePots(bets)

	// 5. 遍历奖池，独立结算
	finalPayouts := make(map[string]int)

	for i, p := range pots {
		fmt.Printf("结算奖池 %d (金额: %d, 有资格的玩家: %v)\n", i, p.Amount, p.Players)

		var bestScore evaluator.HandResult
		var winners []string

		// 仅在有资格分该奖池的玩家中进行比牌
		for _, playerID := range p.Players {
			// 将手牌和公共牌合并 (7张牌)
			allCards := append([]core.Card{}, playerCards[playerID]...)
			allCards = append(allCards, publicCards...)

			// 评估牌力
			score := evaluator.Evaluate(allCards)

			if len(winners) == 0 {
				bestScore = score
				winners = []string{playerID}
			} else {
				cmp := evaluator.Compare(score, bestScore)
				if cmp > 0 { // 当前玩家牌更大
					bestScore = score
					winners = []string{playerID}
				} else if cmp == 0 { // 平局
					winners = append(winners, playerID)
				}
			}
		}

		fmt.Printf("  赢家: %v\n", winners)

		// 分发当前奖池
		payouts := pot.DistributePot(p, winners)
		for pid, amount := range payouts {
			finalPayouts[pid] += amount
			fmt.Printf("  玩家 %s 分得: %d\n", pid, amount)
		}
	}

	fmt.Printf("\n最终结算结果:\n")
	fmt.Printf("玩家 A: %d\n", finalPayouts["A"])
	fmt.Printf("玩家 B: %d\n", finalPayouts["B"])
	fmt.Printf("玩家 C: %d\n", finalPayouts["C"])

	// Output:
	// 结算奖池 0 (金额: 300, 有资格的玩家: [A B C])
	//   赢家: [A]
	//   玩家 A 分得: 300
	// 结算奖池 1 (金额: 200, 有资格的玩家: [B C])
	//   赢家: [B]
	//   玩家 B 分得: 200
	//
	// 最终结算结果:
	// 玩家 A: 300
	// 玩家 B: 200
	// 玩家 C: 0
}
