package domain

// Card 扑克牌
type Card struct {
	Suit string `json:"suit"` // 花色: "s" (Spades), "h" (Hearts), "d" (Diamonds), "c" (Clubs)
	Rank int    `json:"rank"` // 点数绝对值: 2~14 (11=J, 12=Q, 13=K, 14=A)
}

// Deck 牌堆
type Deck struct {
	Cards []Card `json:"cards"` // 剩余的扑克牌列表
}
