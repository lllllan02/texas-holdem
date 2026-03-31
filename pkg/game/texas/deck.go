package texas

import (
	"errors"
	"math/rand"
	"time"
)

// Deck 牌堆
type Deck struct {
	Cards []Card // 剩余的扑克牌列表
}

// NewDeck 初始化一副包含 52 张牌的新牌堆，按顺序排列
func NewDeck() *Deck {
	deck := &Deck{
		Cards: make([]Card, 0, 52),
	}

	// 遍历 4 种花色和 13 种点数，生成 52 张牌
	for suit := SuitSpades; suit <= SuitClubs; suit++ {
		for rank := Rank2; rank <= RankA; rank++ {
			deck.Cards = append(deck.Cards, Card{Suit: suit, Rank: rank})
		}
	}
	return deck
}

// Shuffle 使用 Fisher-Yates 算法随机打乱牌堆
func (d *Deck) Shuffle() {
	// 初始化随机数种子
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	// r.Shuffle 是 Go 1.10 引入的内置洗牌函数
	r.Shuffle(len(d.Cards), func(i, j int) {
		d.Cards[i], d.Cards[j] = d.Cards[j], d.Cards[i]
	})
}

// Draw 从牌堆顶部（切片尾部）抽出 n 张牌
// 抽出的牌会从牌堆中移除
func (d *Deck) Draw(n int) ([]Card, error) {
	if n < 0 {
		return nil, errors.New("cannot draw a negative number of cards")
	}
	if n > len(d.Cards) {
		return nil, errors.New("not enough cards in the deck")
	}

	// 从切片尾部取牌（效率更高，不需要移动前面的元素）
	drawn := make([]Card, n)
	copy(drawn, d.Cards[len(d.Cards)-n:])

	// 更新牌堆切片，移除已经抽出的牌
	d.Cards = d.Cards[:len(d.Cards)-n]

	return drawn, nil
}
