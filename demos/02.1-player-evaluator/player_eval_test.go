package playereval

import (
	"reflect"
	"testing"

	core "github.com/lllllan02/texas-holdem/demos/01-core"
	evaluator "github.com/lllllan02/texas-holdem/demos/02-evaluator"
)

func TestEvaluatePlayerHand(t *testing.T) {
	publicCards := []core.Card{
		{Suit: core.Spades, Rank: core.Rank10},
		{Suit: core.Hearts, Rank: core.Rank9},
		{Suit: core.Diamonds, Rank: core.Rank8},
		{Suit: core.Clubs, Rank: core.Rank2},
		{Suit: core.Spades, Rank: core.Rank3},
	}

	playerA := &Player{
		ID: "A",
		HoleCards: []core.Card{
			{Suit: core.Spades, Rank: core.RankA},
			{Suit: core.Hearts, Rank: core.RankA},
		},
	}

	// 第一次计算
	playerA.EvaluateHand(publicCards)
	if playerA.BestHand.Rank != evaluator.Pair {
		t.Errorf("Expected Pair, got %v", playerA.BestHand.Rank)
	}
	if playerA.BestHand.BestCards[0].Rank != core.RankA {
		t.Errorf("Expected Best Card to be Ace, got %v", playerA.BestHand.BestCards[0].Rank)
	}
	if !playerA.Evaluated {
		t.Errorf("Expected player to be marked as evaluated")
	}

	// 保存第一次的结果用于对比
	res1 := playerA.BestHand

	// 第二次计算（应该直接返回缓存结果）
	playerA.EvaluateHand(publicCards)
	if !reflect.DeepEqual(res1, playerA.BestHand) {
		t.Errorf("Expected cached result to be identical to first evaluation")
	}
}

func TestFindWinners(t *testing.T) {
	publicCards := []core.Card{
		{Suit: core.Spades, Rank: core.Rank10},
		{Suit: core.Spades, Rank: core.Rank9},
		{Suit: core.Spades, Rank: core.Rank8},
		{Suit: core.Clubs, Rank: core.Rank2},
		{Suit: core.Spades, Rank: core.Rank3},
	}

	playerA := &Player{
		ID: "A",
		HoleCards: []core.Card{
			{Suit: core.Spades, Rank: core.RankA},
			{Suit: core.Spades, Rank: core.Rank2},
		},
	}

	playerB := &Player{
		ID: "B",
		HoleCards: []core.Card{
			{Suit: core.Hearts, Rank: core.RankQ},
			{Suit: core.Diamonds, Rank: core.RankJ},
		},
	}

	playerC := &Player{
		ID: "C",
		HoleCards: []core.Card{
			{Suit: core.Spades, Rank: core.RankK},
			{Suit: core.Spades, Rank: core.Rank4},
		},
	}

	players := []*Player{playerA, playerB, playerC}

	winners, bestRes := FindWinners(players, publicCards)

	if len(winners) != 1 || winners[0].ID != "A" {
		t.Errorf("Expected player A to win, got %v", winners)
	}
	if bestRes.Rank != evaluator.Flush {
		t.Errorf("Expected Flush, got %v", bestRes.Rank)
	}
}

func TestFindWinners_Tie(t *testing.T) {
	publicCards := []core.Card{
		{Suit: core.Spades, Rank: core.Rank10},
		{Suit: core.Hearts, Rank: core.Rank10},
		{Suit: core.Diamonds, Rank: core.Rank8},
		{Suit: core.Clubs, Rank: core.Rank8},
		{Suit: core.Spades, Rank: core.RankA},
	}

	playerA := &Player{
		ID: "A",
		HoleCards: []core.Card{
			{Suit: core.Spades, Rank: core.RankK},
			{Suit: core.Clubs, Rank: core.RankK},
		},
	}

	playerB := &Player{
		ID: "B",
		HoleCards: []core.Card{
			{Suit: core.Hearts, Rank: core.RankK},
			{Suit: core.Diamonds, Rank: core.RankK},
		},
	}

	players := []*Player{playerA, playerB}

	winners, bestRes := FindWinners(players, publicCards)

	if len(winners) != 2 {
		t.Errorf("Expected 2 winners for a tie, got %d", len(winners))
	}
	if bestRes.Rank != evaluator.TwoPair {
		t.Errorf("Expected TwoPair, got %v", bestRes.Rank)
	}

	expectedIDs := []string{"A", "B"}
	gotIDs := []string{winners[0].ID, winners[1].ID}
	if !reflect.DeepEqual(gotIDs, expectedIDs) {
		t.Errorf("Expected winners %v, got %v", expectedIDs, gotIDs)
	}
}
