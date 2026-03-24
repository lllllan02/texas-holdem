package playerpot

import (
	"reflect"
	"testing"

	core "github.com/lllllan02/texas-holdem/demos/01-core"
)

func TestCalculatePots(t *testing.T) {
	playerA := &Player{ID: "A", Bet: 100}
	playerB := &Player{ID: "B", Bet: 200}
	playerC := &Player{ID: "C", Bet: 200}

	players := []*Player{playerA, playerB, playerC}
	pots := CalculatePots(players)

	if len(pots) != 2 {
		t.Fatalf("Expected 2 pots, got %d", len(pots))
	}

	// Pot 0: 300, Players: A, B, C
	if pots[0].Amount != 300 {
		t.Errorf("Expected Pot 0 to have 300, got %d", pots[0].Amount)
	}
	if len(pots[0].Players) != 3 {
		t.Errorf("Expected Pot 0 to have 3 players, got %d", len(pots[0].Players))
	}

	// Pot 1: 200, Players: B, C
	if pots[1].Amount != 200 {
		t.Errorf("Expected Pot 1 to have 200, got %d", pots[1].Amount)
	}
	if len(pots[1].Players) != 2 {
		t.Errorf("Expected Pot 1 to have 2 players, got %d", len(pots[1].Players))
	}
}

func TestResolveGame_ShortStackWinsMainPot(t *testing.T) {
	publicCards := []core.Card{
		{Suit: core.Spades, Rank: core.Rank10},
		{Suit: core.Hearts, Rank: core.Rank9},
		{Suit: core.Diamonds, Rank: core.Rank8},
		{Suit: core.Clubs, Rank: core.Rank2},
		{Suit: core.Spades, Rank: core.Rank3},
	}

	// 玩家 A (短码 All-in): 一对 A (全场最大)
	playerA := &Player{
		ID: "A",
		HoleCards: []core.Card{
			{Suit: core.Spades, Rank: core.RankA},
			{Suit: core.Hearts, Rank: core.RankA},
		},
		Bet: 100,
	}

	// 玩家 B: 一对 K (第二大)
	playerB := &Player{
		ID: "B",
		HoleCards: []core.Card{
			{Suit: core.Spades, Rank: core.RankK},
			{Suit: core.Hearts, Rank: core.RankK},
		},
		Bet: 200,
	}

	// 玩家 C: 一对 Q (最小)
	playerC := &Player{
		ID: "C",
		HoleCards: []core.Card{
			{Suit: core.Spades, Rank: core.RankQ},
			{Suit: core.Hearts, Rank: core.RankQ},
		},
		Bet: 200,
	}

	players := []*Player{playerA, playerB, playerC}

	payouts := ResolveGame(players, publicCards)

	expectedPayouts := map[string]int{
		"A": 300, // 赢了主池 (100 * 3)
		"B": 200, // 赢了边池 (100 * 2)
	}

	if !reflect.DeepEqual(payouts, expectedPayouts) {
		t.Errorf("Expected payouts %v, got %v", expectedPayouts, payouts)
	}
}

func TestResolveGame_TieAndFold(t *testing.T) {
	publicCards := []core.Card{
		{Suit: core.Spades, Rank: core.Rank10},
		{Suit: core.Hearts, Rank: core.Rank10},
		{Suit: core.Diamonds, Rank: core.Rank8},
		{Suit: core.Clubs, Rank: core.Rank8},
		{Suit: core.Spades, Rank: core.RankA},
	}

	// 玩家 A: 弃牌，贡献了 50 死钱
	playerA := &Player{
		ID: "A",
		HoleCards: []core.Card{
			{Suit: core.Spades, Rank: core.Rank2},
			{Suit: core.Hearts, Rank: core.Rank3},
		},
		Bet:      50,
		IsFolded: true,
	}

	// 玩家 B 和 C 牌型一样，平分奖池
	playerB := &Player{
		ID: "B",
		HoleCards: []core.Card{
			{Suit: core.Spades, Rank: core.RankK},
			{Suit: core.Clubs, Rank: core.RankK},
		},
		Bet: 100,
	}

	playerC := &Player{
		ID: "C",
		HoleCards: []core.Card{
			{Suit: core.Hearts, Rank: core.RankK},
			{Suit: core.Diamonds, Rank: core.RankK},
		},
		Bet: 100,
	}

	players := []*Player{playerA, playerB, playerC}

	payouts := ResolveGame(players, publicCards)

	// 总奖池: 50 (A) + 100 (B) + 100 (C) = 250
	// B 和 C 平分 250 -> 每人 125
	expectedPayouts := map[string]int{
		"B": 125,
		"C": 125,
	}

	if !reflect.DeepEqual(payouts, expectedPayouts) {
		t.Errorf("Expected payouts %v, got %v", expectedPayouts, payouts)
	}
}
