package texas

// ActionInfo 记录玩家的动作或发牌动作
type ActionInfo struct {
	PlayerID   string   `json:"player_id,omitempty"`
	Action     string   `json:"action,omitempty"` // "fold", "check", "call", "bet", "raise", "allin"
	Amount     int      `json:"amount,omitempty"`
	Stage      string   `json:"stage,omitempty"`
	BoardCards []string `json:"board_cards,omitempty"`
	SidePots   []Pot    `json:"side_pots,omitempty"`
}
