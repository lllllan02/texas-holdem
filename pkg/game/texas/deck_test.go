package texas

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewDeck(t *testing.T) {
	deck := NewDeck()

	// 验证牌堆数量是否严格等于 52
	assert.Len(t, deck.Cards, 52, "Expected exactly 52 cards")

	// 验证是否有重复的牌
	seen := make(map[Card]bool)
	for _, card := range deck.Cards {
		assert.False(t, seen[card], "Duplicate card found: %v", card)
		seen[card] = true
	}
}

func TestShuffle(t *testing.T) {
	deck1 := NewDeck()
	deck2 := NewDeck()

	deck1.Shuffle()

	// 比较 deck1 和 deck2 的顺序。洗牌后大概率顺序是不同的。
	identical := true
	for i := range deck1.Cards {
		if deck1.Cards[i] != deck2.Cards[i] {
			identical = false
			break
		}
	}

	assert.False(t, identical, "Deck was not shuffled properly (or we hit a 1 in 52! chance)")
}

func TestDraw(t *testing.T) {
	deck := NewDeck()

	// 1. 抽 2 张牌（模拟发底牌）
	cards, err := deck.Draw(2)
	assert.NoError(t, err)
	assert.Len(t, cards, 2)
	assert.Len(t, deck.Cards, 50)

	// 2. 抽 5 张牌（模拟发公共牌）
	cards, err = deck.Draw(5)
	assert.NoError(t, err)
	assert.Len(t, cards, 5)
	assert.Len(t, deck.Cards, 45)

	// 3. 尝试从空牌堆抽牌（抽超出剩余数量的牌）
	_, err = deck.Draw(50)
	assert.Error(t, err, "Expected an error when drawing more cards than available")
}
