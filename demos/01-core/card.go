package core

// Suit 表示扑克牌的花色
type Suit int

const (
	Spades   Suit = iota // 黑桃 ♠
	Hearts               // 红桃 ♥
	Diamonds             // 方块 ♦
	Clubs                // 梅花 ♣
)

// Rank 表示扑克牌的点数
type Rank int

const (
	Rank2 Rank = iota + 2 // 2
	Rank3                 // 3
	Rank4                 // 4
	Rank5                 // 5
	Rank6                 // 6
	Rank7                 // 7
	Rank8                 // 8
	Rank9                 // 9
	Rank10                // 10
	RankJ                 // 11 (Jack)
	RankQ                 // 12 (Queen)
	RankK                 // 13 (King)
	RankA                 // 14 (Ace)
)

// Card 表示一张扑克牌
type Card struct {
	Suit Suit
	Rank Rank
}

// String 实现了 fmt.Stringer 接口，方便打印输出（例如 "♠A"）
func (c Card) String() string {
	suits := []string{"♠", "♥", "♦", "♣"}
	ranks := []string{"", "", "2", "3", "4", "5", "6", "7", "8", "9", "10", "J", "Q", "K", "A"}
	
	return suits[c.Suit] + ranks[c.Rank]
}
