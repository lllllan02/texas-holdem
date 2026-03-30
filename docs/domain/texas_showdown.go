package domain

// ShowdownSummary 结算结果摘要 (严格对齐 poker.json 的 showdown_summary)
type ShowdownSummary struct {
	BoardCards []Card     `json:"board_cards"`        // 最终的公共牌（最多5张）
	ShowCards  bool       `json:"show_cards"`         // 是否需要亮牌
	SidePots   []*SidePot `json:"side_pots"`          // 奖池分配结果
	AllHands   []HandInfo `json:"all_hands,omitempty"` // 参与比牌的玩家牌型信息
}

// HandInfo 玩家的牌型信息
type HandInfo struct {
	PlayerID string `json:"player_id"`
	Cards    []Card `json:"cards"`     // 玩家的底牌
	HandName string `json:"hand_name"` // 如 "One Pair", "Flush"
}
