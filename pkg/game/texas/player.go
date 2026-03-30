package texas

// PlayerState 玩家单局状态
type PlayerState string

const (
	StateActive PlayerState = "active" // 正常存活
	StateFolded PlayerState = "folded" // 已弃牌
	StateAllIn  PlayerState = "allin"  // 已全下
)

// Player 玩家层：与 poker.json 中的 players 数组元素严格对应
type Player struct {
	ID                string      `json:"id"`
	Name              string      `json:"name"`
	Position          string      `json:"position"`             // 位置标识 (SB, BB, UTG, BTN 等)
	SeatNumber        int         `json:"seat_number"`          // 座位号
	Chips             int         `json:"chips"`                // 当前携带的筹码量
	CurrentBet        int         `json:"current_bet"`          // 本轮已下注金额
	IsAI              bool        `json:"is_ai"`                // 是否为机器人
	State             PlayerState `json:"state"`                // 本局状态 (active, folded, allin)
	HasActedThisRound bool        `json:"has_acted_this_round"` // 本轮是否已经表态过
	AIProfile         string      `json:"ai_profile,omitempty"` // 机器人性格 (如 "tag", "lag", "reg")
	HoleCards         []string    `json:"hole_cards"`           // 底牌 ["Qs", "3d"]，如果是别人且未亮牌则为 null
}
