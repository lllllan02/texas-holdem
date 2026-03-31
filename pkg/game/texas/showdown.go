package texas

// ShowdownSummary 结算结果摘要 (严格对齐 poker.json 的 showdown_summary)
type ShowdownSummary struct {
	BoardCards []Card     // 最终的公共牌（最多5张）
	ShowCards  bool       // 是否需要亮牌
	SidePots   []*SidePot // 奖池分配结果
	AllHands   []HandInfo // 参与比牌的玩家牌型信息
}

// HandInfo 玩家的牌型信息
type HandInfo struct {
	PlayerID string
	Cards    []Card   // 玩家的底牌
	HandRank HandRank // 牌型等级
}
