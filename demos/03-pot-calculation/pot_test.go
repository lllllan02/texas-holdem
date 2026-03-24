package pot

import (
	"reflect"
	"testing"
)

func TestCalculatePots(t *testing.T) {
	tests := []struct {
		name     string
		bets     []PlayerBet
		expected []Pot
	}{
		{
			name: "1. 单池场景: A, B, C 各下注 100，无人 All-in",
			bets: []PlayerBet{
				{PlayerID: "A", Bet: 100},
				{PlayerID: "B", Bet: 100},
				{PlayerID: "C", Bet: 100},
			},
			expected: []Pot{
				{Amount: 300, Players: []string{"A", "B", "C"}},
			},
		},
		{
			name: "2. 单边池场景: A 筹码 100 (All-in)，B 筹码 200，C 筹码 200",
			bets: []PlayerBet{
				{PlayerID: "A", Bet: 100},
				{PlayerID: "B", Bet: 200},
				{PlayerID: "C", Bet: 200},
			},
			expected: []Pot{
				{Amount: 300, Players: []string{"A", "B", "C"}},
				{Amount: 200, Players: []string{"B", "C"}},
			},
		},
		{
			name: "3. 多边池场景: A(10), B(20), C(30), D(40) 全部 All-in",
			bets: []PlayerBet{
				{PlayerID: "A", Bet: 10},
				{PlayerID: "B", Bet: 20},
				{PlayerID: "C", Bet: 30},
				{PlayerID: "D", Bet: 40},
			},
			expected: []Pot{
				{Amount: 40, Players: []string{"A", "B", "C", "D"}},
				{Amount: 30, Players: []string{"B", "C", "D"}},
				{Amount: 20, Players: []string{"C", "D"}},
				{Amount: 10, Players: []string{"D"}},
			},
		},
		{
			name: "4. 弃牌场景: A 下注 100 后弃牌，这 100 变成死钱进入奖池，但 A 不在 Players 中",
			bets: []PlayerBet{
				{PlayerID: "A", Bet: 100, IsFolded: true},
				{PlayerID: "B", Bet: 200},
				{PlayerID: "C", Bet: 200},
			},
			expected: []Pot{
				{Amount: 500, Players: []string{"B", "C"}},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CalculatePots(tt.bets)
			if !reflect.DeepEqual(got, tt.expected) {
				t.Errorf("CalculatePots() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestDistributePot(t *testing.T) {
	tests := []struct {
		name     string
		pot      Pot
		winners  []string
		expected map[string]int
	}{
		{
			name: "5. 分钱零头测试: 奖池 100，3 人平局",
			pot: Pot{
				Amount:  100,
				Players: []string{"A", "B", "C"},
			},
			winners: []string{"A", "B", "C"},
			expected: map[string]int{
				"A": 34,
				"B": 33,
				"C": 33,
			},
		},
		{
			name: "单人获胜",
			pot: Pot{
				Amount:  100,
				Players: []string{"A", "B", "C"},
			},
			winners: []string{"B"},
			expected: map[string]int{
				"B": 100,
			},
		},
		{
			name: "赢家不在 Players 中",
			pot: Pot{
				Amount:  100,
				Players: []string{"B", "C"},
			},
			winners: []string{"A"},
			expected: map[string]int{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DistributePot(tt.pot, tt.winners)
			if !reflect.DeepEqual(got, tt.expected) {
				t.Errorf("DistributePot() = %v, want %v", got, tt.expected)
			}
		})
	}
}
