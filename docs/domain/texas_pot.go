package domain

// SidePot 边池
// 当有玩家 All-in 时，如果他的筹码不足以跟注当前最高下注额，奖池会发生拆分
type SidePot struct {
	PotNumber       int      `json:"pot_number"`          // 奖池编号（1通常为主池，2,3为边池）
	Amount          int      `json:"amount"`              // 该池内的筹码总额
	Winners         []string `json:"winners"`             // 赢家ID列表 (单赢家时数组长度为1，平分时长度>1)
	HandName        string   `json:"hand_name,omitempty"` // 赢的牌型（如 "One Pair"）
	EligiblePlayers []string `json:"eligible_players"`    // 有资格竞争该奖池的玩家 ID 列表
	Type            string   `json:"type"`                // "main" 或 "side"
}
