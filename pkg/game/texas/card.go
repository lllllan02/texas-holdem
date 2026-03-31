package texas

// Card 扑克牌
type Card struct {
	Suit Suit `json:"suit"` // 花色 (s, h, d, c)
	Rank Rank `json:"rank"` // 点数 (2~14)
}

// String 返回单张牌的字符串表示 (如 "As" 代表黑桃A, "Th" 代表红桃10)
func (c Card) String() string {
	return c.Rank.String() + c.Suit.String()
}
