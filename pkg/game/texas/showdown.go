package texas

// ShowdownSummary 结算结果摘要
type ShowdownSummary struct {
	ShowCards bool       `json:"show_cards"`          // 是否需要亮牌（例如只有一人未弃牌则无需亮牌）
	AllHands  []HandInfo `json:"all_hands,omitempty"` // 所有参与比牌的玩家的牌型信息
	SidePots  []Pot      `json:"side_pots"`           // 奖池分配结果（每个奖池包含各自的赢家和赢取金额）
}

// HandInfo 玩家的牌型信息
type HandInfo struct {
	PlayerID     string   `json:"player_id"`
	Cards        []string `json:"cards"`
	HandName     string   `json:"hand_name"`
	HandStrength int64    `json:"hand_strength"`
}
