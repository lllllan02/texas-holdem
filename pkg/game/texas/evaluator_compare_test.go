package texas

import (
	"testing"
)

func TestCompare(t *testing.T) {
	tests := []struct {
		name string
		a    []Card
		b    []Card
		want int // 1 if a > b, -1 if a < b, 0 if a == b
	}{
		{
			name: "Two Pair vs Two Pair",
			a: []Card{
				{SuitSpades, RankA},
				{SuitHearts, RankA},
				{SuitDiamonds, RankK},
				{SuitClubs, RankK},
				{SuitSpades, RankQ},
			},
			b: []Card{
				{SuitSpades, RankA},
				{SuitHearts, RankA},
				{SuitDiamonds, RankQ},
				{SuitClubs, RankQ},
				{SuitSpades, RankK},
			},
			want: 1,
		},
		{
			name: "Full House vs Full House",
			a: []Card{
				{SuitSpades, RankA},
				{SuitHearts, RankA},
				{SuitDiamonds, RankA},
				{SuitClubs, RankK},
				{SuitSpades, RankK},
			},
			b: []Card{
				{SuitSpades, RankK},
				{SuitHearts, RankK},
				{SuitDiamonds, RankK},
				{SuitClubs, RankA},
				{SuitSpades, RankA},
			},
			want: 1,
		},
		{
			name: "Flush vs Flush",
			a: []Card{
				{SuitSpades, RankA},
				{SuitSpades, RankK},
				{SuitSpades, RankQ},
				{SuitSpades, RankJ},
				{SuitSpades, Rank9},
			},
			b: []Card{
				{SuitHearts, RankA},
				{SuitHearts, RankK},
				{SuitHearts, RankQ},
				{SuitHearts, RankJ},
				{SuitHearts, Rank8},
			},
			want: 1,
		},
		{
			name: "Straight vs Straight",
			a: []Card{
				{SuitSpades, RankA},
				{SuitHearts, RankK},
				{SuitDiamonds, RankQ},
				{SuitClubs, RankJ},
				{SuitSpades, RankT},
			},
			b: []Card{
				{SuitSpades, RankK},
				{SuitHearts, RankQ},
				{SuitDiamonds, RankJ},
				{SuitClubs, RankT},
				{SuitSpades, Rank9},
			},
			want: 1,
		},
		{
			name: "Low Straight vs High Straight",
			a: []Card{
				{SuitSpades, Rank5},
				{SuitHearts, Rank4},
				{SuitDiamonds, Rank3},
				{SuitClubs, Rank2},
				{SuitSpades, RankA},
			},
			b: []Card{
				{SuitSpades, Rank6},
				{SuitHearts, Rank5},
				{SuitDiamonds, Rank4},
				{SuitClubs, Rank3},
				{SuitSpades, Rank2},
			},
			want: -1, // b wins
		},
		{
			name: "Pair vs Pair",
			a: []Card{
				{SuitSpades, RankA},
				{SuitHearts, RankA},
				{SuitDiamonds, RankK},
				{SuitClubs, RankQ},
				{SuitSpades, RankJ},
			},
			b: []Card{
				{SuitSpades, RankK},
				{SuitHearts, RankK},
				{SuitDiamonds, RankA},
				{SuitClubs, RankQ},
				{SuitSpades, RankJ},
			},
			want: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resA := Evaluate(tt.a)
			resB := Evaluate(tt.b)
			got := Compare(resA, resB)
			if got != tt.want {
				t.Errorf("expected %v, got %v", tt.want, got)
			}
		})
	}
}
