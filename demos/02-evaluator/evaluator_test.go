package evaluator

import (
	core "poker/demos/01-core"
	"testing"

	"github.com/stretchr/testify/assert"
)

// 辅助函数：快速创建一张牌
func c(s core.Suit, r core.Rank) core.Card {
	return core.Card{Suit: s, Rank: r}
}

func TestEvaluate(t *testing.T) {
	// 1. 测试正常牌型评估
	tests := []struct {
		name     string
		cards    []core.Card
		expected HandRank
	}{
		{
			name: "Royal Flush",
			cards: []core.Card{
				c(core.Spades, core.RankA),
				c(core.Spades, core.RankK),
				c(core.Spades, core.RankQ),
				c(core.Spades, core.RankJ),
				c(core.Spades, core.Rank10),
				c(core.Hearts, core.Rank2),
				c(core.Clubs, core.Rank3),
			},
			expected: RoyalFlush,
		},
		{
			name: "Straight Flush (Low A-2-3-4-5)",
			cards: []core.Card{
				c(core.Hearts, core.Rank5),
				c(core.Hearts, core.Rank4),
				c(core.Hearts, core.Rank3),
				c(core.Hearts, core.Rank2),
				c(core.Hearts, core.RankA),
				c(core.Spades, core.Rank9),
				c(core.Clubs, core.Rank9),
			},
			expected: StraightFlush,
		},
		{
			name: "Four of a Kind",
			cards: []core.Card{
				c(core.Diamonds, core.Rank8),
				c(core.Spades, core.Rank8),
				c(core.Hearts, core.Rank8),
				c(core.Clubs, core.Rank8),
				c(core.Spades, core.RankK),
				c(core.Hearts, core.Rank2),
				c(core.Clubs, core.Rank3),
			},
			expected: FourOfAKind,
		},
		{
			name: "Full House",
			cards: []core.Card{
				c(core.Diamonds, core.Rank10),
				c(core.Spades, core.Rank10),
				c(core.Hearts, core.Rank10),
				c(core.Clubs, core.Rank7),
				c(core.Spades, core.Rank7),
				c(core.Hearts, core.Rank2),
				c(core.Clubs, core.Rank3),
			},
			expected: FullHouse,
		},
		{
			name: "Flush",
			cards: []core.Card{
				c(core.Clubs, core.Rank2),
				c(core.Clubs, core.Rank5),
				c(core.Clubs, core.Rank7),
				c(core.Clubs, core.Rank9),
				c(core.Clubs, core.RankK),
				c(core.Hearts, core.Rank2),
				c(core.Spades, core.Rank3),
			},
			expected: Flush,
		},
		{
			name: "Straight",
			cards: []core.Card{
				c(core.Clubs, core.Rank8),
				c(core.Diamonds, core.Rank7),
				c(core.Hearts, core.Rank6),
				c(core.Spades, core.Rank5),
				c(core.Clubs, core.Rank4),
				c(core.Hearts, core.Rank2),
				c(core.Spades, core.Rank3),
			},
			expected: Straight,
		},
		{
			name: "Three of a Kind",
			cards: []core.Card{
				c(core.Clubs, core.RankQ),
				c(core.Diamonds, core.RankQ),
				c(core.Hearts, core.RankQ),
				c(core.Spades, core.Rank5),
				c(core.Clubs, core.Rank4),
				c(core.Hearts, core.Rank2),
				c(core.Spades, core.Rank3),
			},
			expected: ThreeOfAKind,
		},
		{
			name: "Two Pair",
			cards: []core.Card{
				c(core.Clubs, core.RankJ),
				c(core.Diamonds, core.RankJ),
				c(core.Hearts, core.Rank9),
				c(core.Spades, core.Rank9),
				c(core.Clubs, core.Rank4),
				c(core.Hearts, core.Rank2),
				c(core.Spades, core.Rank3),
			},
			expected: TwoPair,
		},
		{
			name: "Pair",
			cards: []core.Card{
				c(core.Clubs, core.RankA),
				c(core.Diamonds, core.RankA),
				c(core.Hearts, core.Rank9),
				c(core.Spades, core.Rank7),
				c(core.Clubs, core.Rank4),
				c(core.Hearts, core.Rank2),
				c(core.Spades, core.Rank3),
			},
			expected: Pair,
		},
		{
			name: "High Card",
			cards: []core.Card{
				c(core.Clubs, core.RankA),
				c(core.Diamonds, core.RankQ),
				c(core.Hearts, core.Rank9),
				c(core.Spades, core.Rank7),
				c(core.Clubs, core.Rank4),
				c(core.Hearts, core.Rank2),
				c(core.Spades, core.Rank3),
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
			Evaluate([]core.Card{
				c(core.Spades, core.Rank2),
				c(core.Hearts, core.Rank3),
			})
		})
	})
}

func TestCompare(t *testing.T) {
	tests := []struct {
		name     string
		cardsA   []core.Card
		cardsB   []core.Card
		expected int // 1: A赢, -1: B赢, 0: 平局
	}{
		// 1. 不同牌型比较
		{
			name: "Different Ranks: Flush > Straight",
			cardsA: []core.Card{
				c(core.Clubs, core.Rank2),
				c(core.Clubs, core.Rank5),
				c(core.Clubs, core.Rank7),
				c(core.Clubs, core.Rank9),
				c(core.Clubs, core.RankK),
			},
			cardsB: []core.Card{
				c(core.Clubs, core.Rank8),
				c(core.Diamonds, core.Rank7),
				c(core.Hearts, core.Rank6),
				c(core.Spades, core.Rank5),
				c(core.Clubs, core.Rank4),
			},
			expected: 1,
		},
		{
			name: "Different Ranks: Full House > Flush",
			cardsA: []core.Card{
				c(core.Spades, core.Rank8),
				c(core.Hearts, core.Rank8),
				c(core.Diamonds, core.Rank8),
				c(core.Clubs, core.Rank2),
				c(core.Spades, core.Rank2),
			},
			cardsB: []core.Card{
				c(core.Clubs, core.Rank2),
				c(core.Clubs, core.Rank5),
				c(core.Clubs, core.Rank7),
				c(core.Clubs, core.Rank9),
				c(core.Clubs, core.RankK),
			},
			expected: 1,
		},

		// 2. 高牌 (High Card) 比较
		{
			name: "High Card: A-K-Q-J-9 > A-K-Q-J-8",
			cardsA: []core.Card{
				c(core.Spades, core.RankA),
				c(core.Hearts, core.RankK),
				c(core.Diamonds, core.RankQ),
				c(core.Clubs, core.RankJ),
				c(core.Spades, core.Rank9),
			},
			cardsB: []core.Card{
				c(core.Clubs, core.RankA),
				c(core.Diamonds, core.RankK),
				c(core.Spades, core.RankQ),
				c(core.Hearts, core.RankJ),
				c(core.Clubs, core.Rank8),
			},
			expected: 1,
		},

		// 3. 对子 (Pair) 比较
		{
			name: "Pair: K-K > Q-Q",
			cardsA: []core.Card{
				c(core.Spades, core.RankK),
				c(core.Hearts, core.RankK),
				c(core.Diamonds, core.Rank9),
				c(core.Clubs, core.Rank8),
				c(core.Spades, core.Rank7),
			},
			cardsB: []core.Card{
				c(core.Clubs, core.RankQ),
				c(core.Diamonds, core.RankQ),
				c(core.Spades, core.RankA),
				c(core.Hearts, core.RankJ),
				c(core.Clubs, core.Rank8),
			},
			expected: 1,
		},
		{
			name: "Pair: Same Pair, Kicker A > K",
			cardsA: []core.Card{
				c(core.Spades, core.Rank8),
				c(core.Hearts, core.Rank8),
				c(core.Diamonds, core.RankA),
				c(core.Clubs, core.Rank7),
				c(core.Spades, core.Rank6),
			},
			cardsB: []core.Card{
				c(core.Clubs, core.Rank8),
				c(core.Diamonds, core.Rank8),
				c(core.Spades, core.RankK),
				c(core.Hearts, core.RankQ),
				c(core.Clubs, core.RankJ),
			},
			expected: 1,
		},

		// 4. 两对 (Two Pair) 比较
		{
			name: "Two Pair: Top Pair A-A > K-K",
			cardsA: []core.Card{
				c(core.Spades, core.RankA),
				c(core.Hearts, core.RankA),
				c(core.Diamonds, core.Rank2),
				c(core.Clubs, core.Rank2),
				c(core.Spades, core.Rank7),
			},
			cardsB: []core.Card{
				c(core.Clubs, core.RankK),
				c(core.Diamonds, core.RankK),
				c(core.Spades, core.RankQ),
				c(core.Hearts, core.RankQ),
				c(core.Clubs, core.RankJ),
			},
			expected: 1,
		},
		{
			name: "Two Pair: Same Top Pair, Bottom Pair 9-9 > 8-8",
			cardsA: []core.Card{
				c(core.Spades, core.RankA),
				c(core.Hearts, core.RankA),
				c(core.Diamonds, core.Rank9),
				c(core.Clubs, core.Rank9),
				c(core.Spades, core.Rank7),
			},
			cardsB: []core.Card{
				c(core.Clubs, core.RankA),
				c(core.Diamonds, core.RankA),
				c(core.Spades, core.Rank8),
				c(core.Hearts, core.Rank8),
				c(core.Clubs, core.RankK),
			},
			expected: 1,
		},
		{
			name: "Two Pair: Same Pairs, Kicker K > Q",
			cardsA: []core.Card{
				c(core.Spades, core.RankA),
				c(core.Hearts, core.RankA),
				c(core.Diamonds, core.Rank9),
				c(core.Clubs, core.Rank9),
				c(core.Spades, core.RankK),
			},
			cardsB: []core.Card{
				c(core.Clubs, core.RankA),
				c(core.Diamonds, core.RankA),
				c(core.Spades, core.Rank9),
				c(core.Hearts, core.Rank9),
				c(core.Clubs, core.RankQ),
			},
			expected: 1,
		},

		// 5. 三条 (Three of a Kind) 比较
		{
			name: "Three of a Kind: 8-8-8 > 7-7-7",
			cardsA: []core.Card{
				c(core.Spades, core.Rank8),
				c(core.Hearts, core.Rank8),
				c(core.Diamonds, core.Rank8),
				c(core.Clubs, core.RankA),
				c(core.Spades, core.RankK),
			},
			cardsB: []core.Card{
				c(core.Clubs, core.Rank7),
				c(core.Diamonds, core.Rank7),
				c(core.Spades, core.Rank7),
				c(core.Hearts, core.RankA),
				c(core.Clubs, core.RankK),
			},
			expected: 1,
		},

		// 6. 顺子 (Straight) 比较
		{
			name: "Straight: 10-high > 9-high",
			cardsA: []core.Card{
				c(core.Spades, core.Rank10),
				c(core.Hearts, core.Rank9),
				c(core.Diamonds, core.Rank8),
				c(core.Clubs, core.Rank7),
				c(core.Spades, core.Rank6),
			},
			cardsB: []core.Card{
				c(core.Clubs, core.Rank9),
				c(core.Diamonds, core.Rank8),
				c(core.Spades, core.Rank7),
				c(core.Hearts, core.Rank6),
				c(core.Clubs, core.Rank5),
			},
			expected: 1,
		},
		{
			name: "Straight: 6-high > 5-high (A-2-3-4-5)",
			cardsA: []core.Card{
				c(core.Spades, core.Rank6),
				c(core.Hearts, core.Rank5),
				c(core.Diamonds, core.Rank4),
				c(core.Clubs, core.Rank3),
				c(core.Spades, core.Rank2),
			},
			cardsB: []core.Card{
				c(core.Clubs, core.Rank5),
				c(core.Diamonds, core.Rank4),
				c(core.Spades, core.Rank3),
				c(core.Hearts, core.Rank2),
				c(core.Clubs, core.RankA),
			},
			expected: 1,
		},

		// 7. 同花 (Flush) 比较
		{
			name: "Flush: A-K-Q-J-9 > A-K-Q-J-8",
			cardsA: []core.Card{
				c(core.Spades, core.RankA),
				c(core.Spades, core.RankK),
				c(core.Spades, core.RankQ),
				c(core.Spades, core.RankJ),
				c(core.Spades, core.Rank9),
			},
			cardsB: []core.Card{
				c(core.Hearts, core.RankA),
				c(core.Hearts, core.RankK),
				c(core.Hearts, core.RankQ),
				c(core.Hearts, core.RankJ),
				c(core.Hearts, core.Rank8),
			},
			expected: 1,
		},

		// 8. 葫芦 (Full House) 比较
		{
			name: "Full House: Trips 8-8-8 > 7-7-7",
			cardsA: []core.Card{
				c(core.Spades, core.Rank8),
				c(core.Hearts, core.Rank8),
				c(core.Diamonds, core.Rank8),
				c(core.Clubs, core.Rank2),
				c(core.Spades, core.Rank2),
			},
			cardsB: []core.Card{
				c(core.Clubs, core.Rank7),
				c(core.Diamonds, core.Rank7),
				c(core.Spades, core.Rank7),
				c(core.Hearts, core.RankA),
				c(core.Clubs, core.RankA),
			},
			expected: 1,
		},
		{
			name: "Full House: Same Trips, Pair A-A > K-K (Board: 8-8-8)",
			cardsA: []core.Card{
				c(core.Spades, core.Rank8),
				c(core.Hearts, core.Rank8),
				c(core.Diamonds, core.Rank8),
				c(core.Clubs, core.RankA),
				c(core.Spades, core.RankA),
			},
			cardsB: []core.Card{
				c(core.Spades, core.Rank8),
				c(core.Hearts, core.Rank8),
				c(core.Diamonds, core.Rank8),
				c(core.Hearts, core.RankK),
				c(core.Clubs, core.RankK),
			},
			expected: 1,
		},

		// 9. 四条 (Four of a Kind) 比较
		{
			name: "Four of a Kind: 8-8-8-8 > 7-7-7-7",
			cardsA: []core.Card{
				c(core.Spades, core.Rank8),
				c(core.Hearts, core.Rank8),
				c(core.Diamonds, core.Rank8),
				c(core.Clubs, core.Rank8),
				c(core.Spades, core.Rank2),
			},
			cardsB: []core.Card{
				c(core.Clubs, core.Rank7),
				c(core.Diamonds, core.Rank7),
				c(core.Spades, core.Rank7),
				c(core.Hearts, core.Rank7),
				c(core.Clubs, core.RankA),
			},
			expected: 1,
		},
		{
			name: "Four of a Kind: Same Quads, Kicker A > K (Board: 8-8-8-8)",
			cardsA: []core.Card{
				c(core.Spades, core.Rank8),
				c(core.Hearts, core.Rank8),
				c(core.Diamonds, core.Rank8),
				c(core.Clubs, core.Rank8),
				c(core.Spades, core.RankA),
			},
			cardsB: []core.Card{
				c(core.Spades, core.Rank8),
				c(core.Hearts, core.Rank8),
				c(core.Diamonds, core.Rank8),
				c(core.Clubs, core.Rank8),
				c(core.Hearts, core.RankK),
			},
			expected: 1,
		},

		// 10. 同花顺 (Straight Flush) 比较
		{
			name: "Straight Flush: J-high > 10-high",
			cardsA: []core.Card{
				c(core.Spades, core.RankJ),
				c(core.Spades, core.Rank10),
				c(core.Spades, core.Rank9),
				c(core.Spades, core.Rank8),
				c(core.Spades, core.Rank7),
			},
			cardsB: []core.Card{
				c(core.Hearts, core.Rank10),
				c(core.Hearts, core.Rank9),
				c(core.Hearts, core.Rank8),
				c(core.Hearts, core.Rank7),
				c(core.Hearts, core.Rank6),
			},
			expected: 1,
		},

		// 11. 完全平局 (Tie)
		{
			name: "Tie: Exact same hand value",
			cardsA: []core.Card{
				c(core.Spades, core.RankA),
				c(core.Hearts, core.RankK),
				c(core.Diamonds, core.RankQ),
				c(core.Clubs, core.RankJ),
				c(core.Spades, core.Rank9),
			},
			cardsB: []core.Card{
				c(core.Clubs, core.RankA),
				c(core.Diamonds, core.RankK),
				c(core.Spades, core.RankQ),
				c(core.Hearts, core.RankJ),
				c(core.Clubs, core.Rank9),
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
		cards         []core.Card
		expectedRank  HandRank
		expectedOrder []core.Rank // 期望的 BestCards 的点数顺序
	}{
		{
			name: "Royal Flush",
			cards: []core.Card{
				c(core.Spades, core.Rank10),
				c(core.Spades, core.RankJ),
				c(core.Spades, core.RankQ),
				c(core.Spades, core.RankK),
				c(core.Spades, core.RankA),
			},
			expectedRank:  RoyalFlush,
			expectedOrder: []core.Rank{core.RankA, core.RankK, core.RankQ, core.RankJ, core.Rank10},
		},
		{
			name: "Straight Flush (Normal)",
			cards: []core.Card{
				c(core.Hearts, core.Rank9),
				c(core.Hearts, core.Rank8),
				c(core.Hearts, core.Rank7),
				c(core.Hearts, core.Rank6),
				c(core.Hearts, core.Rank5),
			},
			expectedRank:  StraightFlush,
			expectedOrder: []core.Rank{core.Rank9, core.Rank8, core.Rank7, core.Rank6, core.Rank5},
		},
		{
			name: "Straight Flush (Low A-2-3-4-5)",
			cards: []core.Card{
				c(core.Clubs, core.RankA),
				c(core.Clubs, core.Rank2),
				c(core.Clubs, core.Rank3),
				c(core.Clubs, core.Rank4),
				c(core.Clubs, core.Rank5),
			},
			expectedRank:  StraightFlush,
			expectedOrder: []core.Rank{core.Rank5, core.Rank4, core.Rank3, core.Rank2, core.RankA},
		},
		{
			name: "Four of a Kind",
			cards: []core.Card{
				c(core.Diamonds, core.Rank8),
				c(core.Spades, core.Rank8),
				c(core.Hearts, core.Rank8),
				c(core.Clubs, core.Rank8),
				c(core.Spades, core.RankK),
			},
			expectedRank:  FourOfAKind,
			expectedOrder: []core.Rank{core.Rank8, core.Rank8, core.Rank8, core.Rank8, core.RankK},
		},
		{
			name: "Full House",
			cards: []core.Card{
				c(core.Diamonds, core.Rank10),
				c(core.Spades, core.Rank10),
				c(core.Hearts, core.Rank10),
				c(core.Clubs, core.Rank7),
				c(core.Spades, core.Rank7),
			},
			expectedRank:  FullHouse,
			expectedOrder: []core.Rank{core.Rank10, core.Rank10, core.Rank10, core.Rank7, core.Rank7},
		},
		{
			name: "Flush",
			cards: []core.Card{
				c(core.Clubs, core.Rank2),
				c(core.Clubs, core.Rank5),
				c(core.Clubs, core.Rank7),
				c(core.Clubs, core.Rank9),
				c(core.Clubs, core.RankK),
			},
			expectedRank:  Flush,
			expectedOrder: []core.Rank{core.RankK, core.Rank9, core.Rank7, core.Rank5, core.Rank2},
		},
		{
			name: "Straight (Normal)",
			cards: []core.Card{
				c(core.Clubs, core.Rank8),
				c(core.Diamonds, core.Rank7),
				c(core.Hearts, core.Rank6),
				c(core.Spades, core.Rank5),
				c(core.Clubs, core.Rank4),
			},
			expectedRank:  Straight,
			expectedOrder: []core.Rank{core.Rank8, core.Rank7, core.Rank6, core.Rank5, core.Rank4},
		},
		{
			name: "Straight (Low A-2-3-4-5)",
			cards: []core.Card{
				c(core.Hearts, core.RankA),
				c(core.Clubs, core.Rank2),
				c(core.Diamonds, core.Rank3),
				c(core.Spades, core.Rank4),
				c(core.Hearts, core.Rank5),
			},
			expectedRank:  Straight,
			expectedOrder: []core.Rank{core.Rank5, core.Rank4, core.Rank3, core.Rank2, core.RankA},
		},
		{
			name: "Three of a Kind",
			cards: []core.Card{
				c(core.Clubs, core.RankQ),
				c(core.Diamonds, core.RankQ),
				c(core.Hearts, core.RankQ),
				c(core.Spades, core.Rank5),
				c(core.Clubs, core.Rank4),
			},
			expectedRank:  ThreeOfAKind,
			expectedOrder: []core.Rank{core.RankQ, core.RankQ, core.RankQ, core.Rank5, core.Rank4},
		},
		{
			name: "Two Pair",
			cards: []core.Card{
				c(core.Clubs, core.RankJ),
				c(core.Diamonds, core.RankJ),
				c(core.Hearts, core.Rank9),
				c(core.Spades, core.Rank9),
				c(core.Clubs, core.Rank4),
			},
			expectedRank:  TwoPair,
			expectedOrder: []core.Rank{core.RankJ, core.RankJ, core.Rank9, core.Rank9, core.Rank4},
		},
		{
			name: "Pair",
			cards: []core.Card{
				c(core.Hearts, core.Rank2),
				c(core.Clubs, core.Rank7),
				c(core.Diamonds, core.Rank2),
				c(core.Spades, core.RankJ),
				c(core.Hearts, core.Rank9),
			},
			expectedRank:  Pair,
			expectedOrder: []core.Rank{core.Rank2, core.Rank2, core.RankJ, core.Rank9, core.Rank7},
		},
		{
			name: "High Card",
			cards: []core.Card{
				c(core.Hearts, core.Rank2),
				c(core.Clubs, core.Rank7),
				c(core.Diamonds, core.Rank4),
				c(core.Spades, core.RankJ),
				c(core.Hearts, core.Rank9),
			},
			expectedRank:  HighCard,
			expectedOrder: []core.Rank{core.RankJ, core.Rank9, core.Rank7, core.Rank4, core.Rank2},
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
	cards := []core.Card{
		c(core.Spades, core.Rank2),
		c(core.Hearts, core.Rank3),
		c(core.Diamonds, core.Rank4),
		c(core.Clubs, core.Rank5),
		c(core.Spades, core.Rank6),
		c(core.Hearts, core.Rank7),
		c(core.Diamonds, core.Rank8),
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
