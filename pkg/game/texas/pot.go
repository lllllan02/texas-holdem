package texas

// Pot 奖池
type Pot struct {
	PotNumber       int      `json:"pot_number"`          // 编号 (1代表主池，2代表边池1，以此类推)
	Amount          int      `json:"amount"`              // 池内筹码量
	Winners         []string `json:"winners,omitempty"`   // 赢家ID列表（支持多人平分）
	HandName        string   `json:"hand_name,omitempty"` // 赢的牌型（如 "One Pair", "opponent_folded"）
	EligiblePlayers []string `json:"eligible_players"`    // 有资格分这个池的玩家ID列表
	Type            string   `json:"type"`                // "main" 或 "side"
}
