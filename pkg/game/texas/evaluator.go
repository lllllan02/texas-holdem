package texas

import (
	"sort"
)

// HandResult 评估结果
type HandResult struct {
	Rank      HandRank
	BestCards []Card // 构成该牌型的最佳 5 张牌，按重要程度排序
}

// Compare 比较两组牌，返回 1 (a赢), -1 (b赢), 0 (平局)
func Compare(a, b HandResult) int {
	if a.Rank > b.Rank {
		return 1
	}
	if a.Rank < b.Rank {
		return -1
	}

	// 牌型相同时，依次比较 5 张牌的大小
	for i := 0; i < 5; i++ {
		if a.BestCards[i].Rank > b.BestCards[i].Rank {
			return 1
		}
		if a.BestCards[i].Rank < b.BestCards[i].Rank {
			return -1
		}
	}

	return 0
}

// Evaluate 输入 5-7 张牌，返回最佳牌型结果
func Evaluate(cards []Card) HandResult {
	if len(cards) < 5 {
		panic("need at least 5 cards to evaluate")
	}

	combos := getCombinations(cards, 5)
	var best HandResult

	for i, combo := range combos {
		res := evaluate5(combo)
		if i == 0 || Compare(res, best) > 0 {
			best = res
		}
	}
	return best
}

// evaluate5 评估恰好 5 张牌的牌型
func evaluate5(cards []Card) HandResult {
	// 按点数降序排列
	sorted := make([]Card, 5)
	copy(sorted, cards)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Rank > sorted[j].Rank
	})

	// 检查同花
	isFlush := true
	for i := 1; i < 5; i++ {
		if sorted[i].Suit != sorted[0].Suit {
			isFlush = false
			break
		}
	}

	// 检查顺子
	isStraight := true
	for i := 1; i < 5; i++ {
		if sorted[i-1].Rank-sorted[i].Rank != 1 {
			isStraight = false
			break
		}
	}

	// 特殊处理 A-2-3-4-5 顺子 (A 算作 1)
	isLowStraight := false
	if !isStraight && sorted[0].Rank == RankA && sorted[1].Rank == Rank5 &&
		sorted[2].Rank == Rank4 && sorted[3].Rank == Rank3 && sorted[4].Rank == Rank2 {
		isStraight = true
		isLowStraight = true
		// 重新排序，把 A 放到最后，因为这里 A 最小
		sorted = []Card{sorted[1], sorted[2], sorted[3], sorted[4], sorted[0]}
	}

	// 统计点数频率
	counts := make(map[Rank]int)
	for _, c := range sorted {
		counts[c.Rank]++
	}

	// 按出现次数降序排列，次数相同则按点数降序排列
	type freq struct {
		rank  Rank
		count int
	}
	var freqs []freq
	for r, c := range counts {
		freqs = append(freqs, freq{r, c})
	}
	sort.Slice(freqs, func(i, j int) bool {
		if freqs[i].count != freqs[j].count {
			return freqs[i].count > freqs[j].count
		}
		return freqs[i].rank > freqs[j].rank
	})

	// 构建 BestCards：对于对子/三条等，优先把出现次数多的牌放在前面，方便 Compare 直接逐个对比
	var bestCards []Card
	if isStraight || isFlush {
		// 顺子和同花没有对子，直接使用降序排列好的牌即可
		bestCards = sorted
	} else {
		for _, f := range freqs {
			for _, c := range sorted {
				if c.Rank == f.rank {
					bestCards = append(bestCards, c)
				}
			}
		}
	}

	// 判定牌型
	rank := HandRankHighCard
	if isFlush && isStraight {
		if bestCards[0].Rank == RankA && !isLowStraight {
			rank = HandRankRoyalFlush
		} else {
			rank = HandRankStraightFlush
		}
	} else if freqs[0].count == 4 {
		rank = HandRankFourOfAKind
	} else if freqs[0].count == 3 && freqs[1].count == 2 {
		rank = HandRankFullHouse
	} else if isFlush {
		rank = HandRankFlush
	} else if isStraight {
		rank = HandRankStraight
	} else if freqs[0].count == 3 {
		rank = HandRankThreeOfAKind
	} else if freqs[0].count == 2 && freqs[1].count == 2 {
		rank = HandRankTwoPair
	} else if freqs[0].count == 2 {
		rank = HandRankPair
	}

	return HandResult{
		Rank:      rank,
		BestCards: bestCards,
	}
}

// getCombinations 从 n 张牌中选出 k 张牌的所有组合
func getCombinations(cards []Card, k int) [][]Card {
	var result [][]Card
	var backtrack func(start int, current []Card)

	backtrack = func(start int, current []Card) {
		if len(current) == k {
			combo := make([]Card, k)
			copy(combo, current)
			result = append(result, combo)
			return
		}
		for i := start; i < len(cards); i++ {
			current = append(current, cards[i])
			backtrack(i+1, current)
			current = current[:len(current)-1]
		}
	}

	backtrack(0, []Card{})
	return result
}
