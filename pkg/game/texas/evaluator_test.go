package texas

import (
	"testing"
)

func TestEvaluate(t *testing.T) {
	tests := []struct {
		name     string
		cards    []Card
		expected HandRank
	}{
		{
			name: "Royal Flush",
			cards: []Card{
				{SuitSpades, RankT},
				{SuitSpades, RankJ},
				{SuitSpades, RankQ},
				{SuitSpades, RankK},
				{SuitSpades, RankA},
			},
			expected: HandRankRoyalFlush,
		},
		{
			name: "Straight Flush Low",
			cards: []Card{
				{SuitSpades, RankA},
				{SuitSpades, Rank2},
				{SuitSpades, Rank3},
				{SuitSpades, Rank4},
				{SuitSpades, Rank5},
			},
			expected: HandRankStraightFlush,
		},
		{
			name: "Four of a Kind",
			cards: []Card{
				{SuitSpades, RankA},
				{SuitHearts, RankA},
				{SuitDiamonds, RankA},
				{SuitClubs, RankA},
				{SuitSpades, RankK},
			},
			expected: HandRankFourOfAKind,
		},
		{
			name: "Full House",
			cards: []Card{
				{SuitSpades, RankA},
				{SuitHearts, RankA},
				{SuitDiamonds, RankA},
				{SuitClubs, RankK},
				{SuitSpades, RankK},
			},
			expected: HandRankFullHouse,
		},
		{
			name: "Flush",
			cards: []Card{
				{SuitSpades, Rank2},
				{SuitSpades, Rank4},
				{SuitSpades, Rank6},
				{SuitSpades, Rank8},
				{SuitSpades, RankT},
			},
			expected: HandRankFlush,
		},
		{
			name: "Straight",
			cards: []Card{
				{SuitSpades, Rank2},
				{SuitHearts, Rank3},
				{SuitDiamonds, Rank4},
				{SuitClubs, Rank5},
				{SuitSpades, Rank6},
			},
			expected: HandRankStraight,
		},
		{
			name: "Three of a Kind",
			cards: []Card{
				{SuitSpades, RankA},
				{SuitHearts, RankA},
				{SuitDiamonds, RankA},
				{SuitClubs, RankK},
				{SuitSpades, RankQ},
			},
			expected: HandRankThreeOfAKind,
		},
		{
			name: "Two Pair",
			cards: []Card{
				{SuitSpades, RankA},
				{SuitHearts, RankA},
				{SuitDiamonds, RankK},
				{SuitClubs, RankK},
				{SuitSpades, RankQ},
			},
			expected: HandRankTwoPair,
		},
		{
			name: "One Pair",
			cards: []Card{
				{SuitSpades, RankA},
				{SuitHearts, RankA},
				{SuitDiamonds, RankK},
				{SuitClubs, RankQ},
				{SuitSpades, RankJ},
			},
			expected: HandRankPair,
		},
		{
			name: "High Card",
			cards: []Card{
				{SuitSpades, RankA},
				{SuitHearts, RankK},
				{SuitDiamonds, RankQ},
				{SuitClubs, RankJ},
				{SuitSpades, Rank9},
			},
			expected: HandRankHighCard,
		},
		{
			name: "7 Cards - Full House",
			cards: []Card{
				{SuitSpades, RankA},
				{SuitHearts, RankA},
				{SuitDiamonds, RankA},
				{SuitClubs, RankK},
				{SuitSpades, RankK},
				{SuitHearts, RankQ},
				{SuitDiamonds, RankJ},
			},
			expected: HandRankFullHouse,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := Evaluate(tt.cards)
			if res.Rank != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, res.Rank)
			}
		})
	}
}
