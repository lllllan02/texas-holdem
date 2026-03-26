package texas

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// 辅助函数：快速创建一张牌
func c(s Suit, r Rank) Card {
	return Card{Suit: s, Rank: r}
}

func TestEvaluate(t *testing.T) {
	// 1. 测试正常牌型评估
	tests := []struct {
		name     string
		cards    []Card
		expected HandRank
	}{
		{
			name: "Royal Flush",
			cards: []Card{
				c(Spades, RankA),
				c(Spades, RankK),
				c(Spades, RankQ),
				c(Spades, RankJ),
				c(Spades, Rank10),
				c(Hearts, Rank2),
				c(Clubs, Rank3),
			},
			expected: RoyalFlush,
		},
		{
			name: "Straight Flush (Low A-2-3-4-5)",
			cards: []Card{
				c(Hearts, Rank5),
				c(Hearts, Rank4),
				c(Hearts, Rank3),
				c(Hearts, Rank2),
				c(Hearts, RankA),
				c(Spades, Rank9),
				c(Clubs, Rank9),
			},
			expected: StraightFlush,
		},
		{
			name: "Four of a Kind",
			cards: []Card{
				c(Diamonds, Rank8),
				c(Spades, Rank8),
				c(Hearts, Rank8),
				c(Clubs, Rank8),
				c(Spades, RankK),
				c(Hearts, Rank2),
				c(Clubs, Rank3),
			},
			expected: FourOfAKind,
		},
		{
			name: "Full House",
			cards: []Card{
				c(Diamonds, Rank10),
				c(Spades, Rank10),
				c(Hearts, Rank10),
				c(Clubs, Rank7),
				c(Spades, Rank7),
				c(Hearts, Rank2),
				c(Clubs, Rank3),
			},
			expected: FullHouse,
		},
		{
			name: "Flush",
			cards: []Card{
				c(Clubs, Rank2),
				c(Clubs, Rank5),
				c(Clubs, Rank7),
				c(Clubs, Rank9),
				c(Clubs, RankK),
				c(Hearts, Rank2),
				c(Spades, Rank3),
			},
			expected: Flush,
		},
		{
			name: "Straight",
			cards: []Card{
				c(Clubs, Rank8),
				c(Diamonds, Rank7),
				c(Hearts, Rank6),
				c(Spades, Rank5),
				c(Clubs, Rank4),
				c(Hearts, Rank2),
				c(Spades, Rank3),
			},
			expected: Straight,
		},
		{
			name: "Three of a Kind",
			cards: []Card{
				c(Clubs, RankQ),
				c(Diamonds, RankQ),
				c(Hearts, RankQ),
				c(Spades, Rank5),
				c(Clubs, Rank4),
				c(Hearts, Rank2),
				c(Spades, Rank3),
			},
			expected: ThreeOfAKind,
		},
		{
			name: "Two Pair",
			cards: []Card{
				c(Clubs, RankJ),
				c(Diamonds, RankJ),
				c(Hearts, Rank9),
				c(Spades, Rank9),
				c(Clubs, Rank4),
				c(Hearts, Rank2),
				c(Spades, Rank3),
			},
			expected: TwoPair,
		},
		{
			name: "Pair",
			cards: []Card{
				c(Clubs, RankA),
				c(Diamonds, RankA),
				c(Hearts, Rank9),
				c(Spades, Rank7),
				c(Clubs, Rank4),
				c(Hearts, Rank2),
				c(Spades, Rank3),
			},
			expected: Pair,
		},
		{
			name: "High Card",
			cards: []Card{
				c(Clubs, RankA),
				c(Diamonds, RankQ),
				c(Hearts, Rank9),
				c(Spades, Rank7),
				c(Clubs, Rank4),
				c(Hearts, Rank2),
				c(Spades, Rank3),
			},
			expected: HighCard,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := Evaluate(tt.cards)
			assert.Equal(t, tt.expected, res.Rank)
		})
	}

	// 2. 测试少于 5 张牌时的 panic
	t.Run("Panic on less than 5 cards", func(t *testing.T) {
		assert.PanicsWithValue(t, "need at least 5 cards to evaluate", func() {
			Evaluate([]Card{
				c(Spades, Rank2),
				c(Hearts, Rank3),
			})
		})
	})
}

func TestCompare(t *testing.T) {
	tests := []struct {
		name     string
		cardsA   []Card
		cardsB   []Card
		expected int // 1: A赢, -1: B赢, 0: 平局
	}{
		// 1. 不同牌型比较
		{
			name: "Different Ranks: Flush > Straight",
			cardsA: []Card{
				c(Clubs, Rank2),
				c(Clubs, Rank5),
				c(Clubs, Rank7),
				c(Clubs, Rank9),
				c(Clubs, RankK),
			},
			cardsB: []Card{
				c(Clubs, Rank8),
				c(Diamonds, Rank7),
				c(Hearts, Rank6),
				c(Spades, Rank5),
				c(Clubs, Rank4),
			},
			expected: 1,
		},
		{
			name: "Different Ranks: Full House > Flush",
			cardsA: []Card{
				c(Spades, Rank8),
				c(Hearts, Rank8),
				c(Diamonds, Rank8),
				c(Clubs, Rank2),
				c(Spades, Rank2),
			},
			cardsB: []Card{
				c(Clubs, Rank2),
				c(Clubs, Rank5),
				c(Clubs, Rank7),
				c(Clubs, Rank9),
				c(Clubs, RankK),
			},
			expected: 1,
		},

		// 2. 高牌 (High Card) 比较
		{
			name: "High Card: A-K-Q-J-9 > A-K-Q-J-8",
			cardsA: []Card{
				c(Spades, RankA),
				c(Hearts, RankK),
				c(Diamonds, RankQ),
				c(Clubs, RankJ),
				c(Spades, Rank9),
			},
			cardsB: []Card{
				c(Clubs, RankA),
				c(Diamonds, RankK),
				c(Spades, RankQ),
				c(Hearts, RankJ),
				c(Clubs, Rank8),
			},
			expected: 1,
		},

		// 3. 对子 (Pair) 比较
		{
			name: "Pair: K-K > Q-Q",
			cardsA: []Card{
				c(Spades, RankK),
				c(Hearts, RankK),
				c(Diamonds, Rank9),
				c(Clubs, Rank8),
				c(Spades, Rank7),
			},
			cardsB: []Card{
				c(Clubs, RankQ),
				c(Diamonds, RankQ),
				c(Spades, RankA),
				c(Hearts, RankJ),
				c(Clubs, Rank8),
			},
			expected: 1,
		},
		{
			name: "Pair: Same Pair, Kicker A > K",
			cardsA: []Card{
				c(Spades, Rank8),
				c(Hearts, Rank8),
				c(Diamonds, RankA),
				c(Clubs, Rank7),
				c(Spades, Rank6),
			},
			cardsB: []Card{
				c(Clubs, Rank8),
				c(Diamonds, Rank8),
				c(Spades, RankK),
				c(Hearts, RankQ),
				c(Clubs, RankJ),
			},
			expected: 1,
		},

		// 4. 两对 (Two Pair) 比较
		{
			name: "Two Pair: Top Pair A-A > K-K",
			cardsA: []Card{
				c(Spades, RankA),
				c(Hearts, RankA),
				c(Diamonds, Rank2),
				c(Clubs, Rank2),
				c(Spades, Rank7),
			},
			cardsB: []Card{
				c(Clubs, RankK),
				c(Diamonds, RankK),
				c(Spades, RankQ),
				c(Hearts, RankQ),
				c(Clubs, RankJ),
			},
			expected: 1,
		},
		{
			name: "Two Pair: Same Top Pair, Bottom Pair 9-9 > 8-8",
			cardsA: []Card{
				c(Spades, RankA),
				c(Hearts, RankA),
				c(Diamonds, Rank9),
				c(Clubs, Rank9),
				c(Spades, Rank7),
			},
			cardsB: []Card{
				c(Clubs, RankA),
				c(Diamonds, RankA),
				c(Spades, Rank8),
				c(Hearts, Rank8),
				c(Clubs, RankK),
			},
			expected: 1,
		},
		{
			name: "Two Pair: Same Pairs, Kicker K > Q",
			cardsA: []Card{
				c(Spades, RankA),
				c(Hearts, RankA),
				c(Diamonds, Rank9),
				c(Clubs, Rank9),
				c(Spades, RankK),
			},
			cardsB: []Card{
				c(Clubs, RankA),
				c(Diamonds, RankA),
				c(Spades, Rank9),
				c(Hearts, Rank9),
				c(Clubs, RankQ),
			},
			expected: 1,
		},

		// 5. 三条 (Three of a Kind) 比较
		{
			name: "Three of a Kind: 8-8-8 > 7-7-7",
			cardsA: []Card{
				c(Spades, Rank8),
				c(Hearts, Rank8),
				c(Diamonds, Rank8),
				c(Clubs, RankA),
				c(Spades, RankK),
			},
			cardsB: []Card{
				c(Clubs, Rank7),
				c(Diamonds, Rank7),
				c(Spades, Rank7),
				c(Hearts, RankA),
				c(Clubs, RankK),
			},
			expected: 1,
		},

		// 6. 顺子 (Straight) 比较
		{
			name: "Straight: 10-high > 9-high",
			cardsA: []Card{
				c(Spades, Rank10),
				c(Hearts, Rank9),
				c(Diamonds, Rank8),
				c(Clubs, Rank7),
				c(Spades, Rank6),
			},
			cardsB: []Card{
				c(Clubs, Rank9),
				c(Diamonds, Rank8),
				c(Spades, Rank7),
				c(Hearts, Rank6),
				c(Clubs, Rank5),
			},
			expected: 1,
		},
		{
			name: "Straight: 6-high > 5-high (A-2-3-4-5)",
			cardsA: []Card{
				c(Spades, Rank6),
				c(Hearts, Rank5),
				c(Diamonds, Rank4),
				c(Clubs, Rank3),
				c(Spades, Rank2),
			},
			cardsB: []Card{
				c(Clubs, Rank5),
				c(Diamonds, Rank4),
				c(Spades, Rank3),
				c(Hearts, Rank2),
				c(Clubs, RankA),
			},
			expected: 1,
		},

		// 7. 同花 (Flush) 比较
		{
			name: "Flush: A-K-Q-J-9 > A-K-Q-J-8",
			cardsA: []Card{
				c(Spades, RankA),
				c(Spades, RankK),
				c(Spades, RankQ),
				c(Spades, RankJ),
				c(Spades, Rank9),
			},
			cardsB: []Card{
				c(Hearts, RankA),
				c(Hearts, RankK),
				c(Hearts, RankQ),
				c(Hearts, RankJ),
				c(Hearts, Rank8),
			},
			expected: 1,
		},

		// 8. 葫芦 (Full House) 比较
		{
			name: "Full House: Trips 8-8-8 > 7-7-7",
			cardsA: []Card{
				c(Spades, Rank8),
				c(Hearts, Rank8),
				c(Diamonds, Rank8),
				c(Clubs, Rank2),
				c(Spades, Rank2),
			},
			cardsB: []Card{
				c(Clubs, Rank7),
				c(Diamonds, Rank7),
				c(Spades, Rank7),
				c(Hearts, RankA),
				c(Clubs, RankA),
			},
			expected: 1,
		},
		{
			name: "Full House: Same Trips, Pair A-A > K-K (Board: 8-8-8)",
			cardsA: []Card{
				c(Spades, Rank8),
				c(Hearts, Rank8),
				c(Diamonds, Rank8),
				c(Clubs, RankA),
				c(Spades, RankA),
			},
			cardsB: []Card{
				c(Spades, Rank8),
				c(Hearts, Rank8),
				c(Diamonds, Rank8),
				c(Hearts, RankK),
				c(Clubs, RankK),
			},
			expected: 1,
		},

		// 9. 四条 (Four of a Kind) 比较
		{
			name: "Four of a Kind: 8-8-8-8 > 7-7-7-7",
			cardsA: []Card{
				c(Spades, Rank8),
				c(Hearts, Rank8),
				c(Diamonds, Rank8),
				c(Clubs, Rank8),
				c(Spades, Rank2),
			},
			cardsB: []Card{
				c(Clubs, Rank7),
				c(Diamonds, Rank7),
				c(Spades, Rank7),
				c(Hearts, Rank7),
				c(Clubs, RankA),
			},
			expected: 1,
		},
		{
			name: "Four of a Kind: Same Quads, Kicker A > K (Board: 8-8-8-8)",
			cardsA: []Card{
				c(Spades, Rank8),
				c(Hearts, Rank8),
				c(Diamonds, Rank8),
				c(Clubs, Rank8),
				c(Spades, RankA),
			},
			cardsB: []Card{
				c(Spades, Rank8),
				c(Hearts, Rank8),
				c(Diamonds, Rank8),
				c(Clubs, Rank8),
				c(Hearts, RankK),
			},
			expected: 1,
		},

		// 10. 同花顺 (Straight Flush) 比较
		{
			name: "Straight Flush: J-high > 10-high",
			cardsA: []Card{
				c(Spades, RankJ),
				c(Spades, Rank10),
				c(Spades, Rank9),
				c(Spades, Rank8),
				c(Spades, Rank7),
			},
			cardsB: []Card{
				c(Hearts, Rank10),
				c(Hearts, Rank9),
				c(Hearts, Rank8),
				c(Hearts, Rank7),
				c(Hearts, Rank6),
			},
			expected: 1,
		},

		// 11. 完全平局 (Tie)
		{
			name: "Tie: Exact same hand value",
			cardsA: []Card{
				c(Spades, RankA),
				c(Hearts, RankK),
				c(Diamonds, RankQ),
				c(Clubs, RankJ),
				c(Spades, Rank9),
			},
			cardsB: []Card{
				c(Clubs, RankA),
				c(Diamonds, RankK),
				c(Spades, RankQ),
				c(Hearts, RankJ),
				c(Clubs, Rank9),
			},
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resA := Evaluate(tt.cardsA)
			resB := Evaluate(tt.cardsB)

			// 正向测试 A vs B
			assert.Equal(t, tt.expected, Compare(resA, resB), "A vs B failed")

			// 反向测试 B vs A
			expectedReverse := tt.expected * -1
			assert.Equal(t, expectedReverse, Compare(resB, resA), "B vs A failed")
		})
	}
}

func TestEvaluate5(t *testing.T) {
	// evaluate5 是内部方法，专门处理恰好 5 张牌的情况
	// 我们需要测试所有 10 种牌型，并且重点验证 BestCards 的排序是否正确
	// （因为 Compare 方法依赖于 BestCards 已经按重要程度排好序）

	tests := []struct {
		name          string
		cards         []Card
		expectedRank  HandRank
		expectedOrder []Rank // 期望的 BestCards 的点数顺序
	}{
		{
			name: "Royal Flush",
			cards: []Card{
				c(Spades, Rank10),
				c(Spades, RankJ),
				c(Spades, RankQ),
				c(Spades, RankK),
				c(Spades, RankA),
			},
			expectedRank:  RoyalFlush,
			expectedOrder: []Rank{RankA, RankK, RankQ, RankJ, Rank10},
		},
		{
			name: "Straight Flush (Normal)",
			cards: []Card{
				c(Hearts, Rank9),
				c(Hearts, Rank8),
				c(Hearts, Rank7),
				c(Hearts, Rank6),
				c(Hearts, Rank5),
			},
			expectedRank:  StraightFlush,
			expectedOrder: []Rank{Rank9, Rank8, Rank7, Rank6, Rank5},
		},
		{
			name: "Straight Flush (Low A-2-3-4-5)",
			cards: []Card{
				c(Clubs, RankA),
				c(Clubs, Rank2),
				c(Clubs, Rank3),
				c(Clubs, Rank4),
				c(Clubs, Rank5),
			},
			expectedRank:  StraightFlush,
			expectedOrder: []Rank{Rank5, Rank4, Rank3, Rank2, RankA},
		},
		{
			name: "Four of a Kind",
			cards: []Card{
				c(Diamonds, Rank8),
				c(Spades, Rank8),
				c(Hearts, Rank8),
				c(Clubs, Rank8),
				c(Spades, RankK),
			},
			expectedRank:  FourOfAKind,
			expectedOrder: []Rank{Rank8, Rank8, Rank8, Rank8, RankK},
		},
		{
			name: "Full House",
			cards: []Card{
				c(Diamonds, Rank10),
				c(Spades, Rank10),
				c(Hearts, Rank10),
				c(Clubs, Rank7),
				c(Spades, Rank7),
			},
			expectedRank:  FullHouse,
			expectedOrder: []Rank{Rank10, Rank10, Rank10, Rank7, Rank7},
		},
		{
			name: "Flush",
			cards: []Card{
				c(Clubs, Rank2),
				c(Clubs, Rank5),
				c(Clubs, Rank7),
				c(Clubs, Rank9),
				c(Clubs, RankK),
			},
			expectedRank:  Flush,
			expectedOrder: []Rank{RankK, Rank9, Rank7, Rank5, Rank2},
		},
		{
			name: "Straight (Normal)",
			cards: []Card{
				c(Clubs, Rank8),
				c(Diamonds, Rank7),
				c(Hearts, Rank6),
				c(Spades, Rank5),
				c(Clubs, Rank4),
			},
			expectedRank:  Straight,
			expectedOrder: []Rank{Rank8, Rank7, Rank6, Rank5, Rank4},
		},
		{
			name: "Straight (Low A-2-3-4-5)",
			cards: []Card{
				c(Hearts, RankA),
				c(Clubs, Rank2),
				c(Diamonds, Rank3),
				c(Spades, Rank4),
				c(Hearts, Rank5),
			},
			expectedRank:  Straight,
			expectedOrder: []Rank{Rank5, Rank4, Rank3, Rank2, RankA},
		},
		{
			name: "Three of a Kind",
			cards: []Card{
				c(Clubs, RankQ),
				c(Diamonds, RankQ),
				c(Hearts, RankQ),
				c(Spades, Rank5),
				c(Clubs, Rank4),
			},
			expectedRank:  ThreeOfAKind,
			expectedOrder: []Rank{RankQ, RankQ, RankQ, Rank5, Rank4},
		},
		{
			name: "Two Pair",
			cards: []Card{
				c(Clubs, RankJ),
				c(Diamonds, RankJ),
				c(Hearts, Rank9),
				c(Spades, Rank9),
				c(Clubs, Rank4),
			},
			expectedRank:  TwoPair,
			expectedOrder: []Rank{RankJ, RankJ, Rank9, Rank9, Rank4},
		},
		{
			name: "Pair",
			cards: []Card{
				c(Hearts, Rank2),
				c(Clubs, Rank7),
				c(Diamonds, Rank2),
				c(Spades, RankJ),
				c(Hearts, Rank9),
			},
			expectedRank:  Pair,
			expectedOrder: []Rank{Rank2, Rank2, RankJ, Rank9, Rank7},
		},
		{
			name: "High Card",
			cards: []Card{
				c(Hearts, Rank2),
				c(Clubs, Rank7),
				c(Diamonds, Rank4),
				c(Spades, RankJ),
				c(Hearts, Rank9),
			},
			expectedRank:  HighCard,
			expectedOrder: []Rank{RankJ, Rank9, Rank7, Rank4, Rank2},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := evaluate5(tt.cards)
			assert.Equal(t, tt.expectedRank, res.Rank, "Rank mismatch")

			assert.Len(t, res.BestCards, 5)
			for i, expectedRank := range tt.expectedOrder {
				assert.Equal(t, expectedRank, res.BestCards[i].Rank, "Card %d rank mismatch in BestCards", i)
			}
		})
	}
}

func TestGetCombinations(t *testing.T) {
	cards := []Card{
		c(Spades, Rank2),
		c(Hearts, Rank3),
		c(Diamonds, Rank4),
		c(Clubs, Rank5),
		c(Spades, Rank6),
		c(Hearts, Rank7),
		c(Diamonds, Rank8),
	}

	t.Run("Choose 5 from 7 (Texas Hold'em case)", func(t *testing.T) {
		combos := getCombinations(cards, 5)

		// C(7,5) = 21
		assert.Len(t, combos, 21)

		// 验证没有重复的组合
		seen := make(map[string]bool)
		for _, combo := range combos {
			assert.Len(t, combo, 5)

			// 将组合转换成字符串作为唯一键
			key := ""
			for _, card := range combo {
				key += card.String() + ","
			}

			assert.False(t, seen[key], "Found duplicate combination: %s", key)
			seen[key] = true
		}
	})

	t.Run("Choose 2 from 4", func(t *testing.T) {
		combos := getCombinations(cards[:4], 2)
		assert.Len(t, combos, 6)
	})

	t.Run("Choose 4 from 4", func(t *testing.T) {
		combos := getCombinations(cards[:4], 4)
		assert.Len(t, combos, 1)
	})

	t.Run("Choose 0 from 4", func(t *testing.T) {
		combos := getCombinations(cards[:4], 0)
		assert.Len(t, combos, 1)
		assert.Len(t, combos[0], 0)
	})
}
