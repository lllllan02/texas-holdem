package texas

// Stage 单局游戏阶段
type Stage string

const (
	StagePreFlop  Stage = "PREFLOP"
	StageFlop     Stage = "FLOP"
	StageTurn     Stage = "TURN"
	StageRiver    Stage = "RIVER"
	StageShowdown Stage = "SHOWDOWN"
)

// Hand 单局层：存活于“发底牌”到“结算”期间，包含驱动游戏进行的所有上下文
type Hand struct {
	// --- 基础信息 ---
	ID        string `json:"id"`         // 本局唯一ID (如 "hand-12345")
	RoomID    string `json:"room_id"`    // 关联的房间ID (原 match_id)
	HandCount int    `json:"hand_count"` // 当前是第几局（对应 poker.json 的 hand_count）
	Stage     Stage  `json:"stage"`      // 当前阶段 (PREFLOP, FLOP, TURN, RIVER, SHOWDOWN)

	// --- 桌面状态 ---
	BoardCards []string `json:"board_cards"` // 公共牌，如 ["Kd", "Js", "7d"]
	Deck       *Deck    `json:"-"`           // 牌堆（内部使用，不暴露给前端）

	// --- 筹码与奖池 ---
	Pot        int   `json:"pot"`         // 当前总奖池 (对应 poker.json 的 pot)
	CurrentBet int   `json:"current_bet"` // 当前下注圈的最高下注额
	SidePots   []Pot `json:"side_pots"`   // 边池列表 (对应 poker.json 的 side_pots)

	// --- 行动流转控制 ---
	ButtonSeat            int    `json:"button_seat"`             // 庄家座位号
	CurrentPlayer         string `json:"current_player"`          // 当前说话玩家的 ID
	CurrentPlayerPosition string `json:"current_player_position"` // 当前说话玩家的位置 (如 BTN)
	CurrentPlayerIndex    int    `json:"current_player_index"`    // 当前说话玩家的座位号
	ActionOrder           []int  `json:"action_order"`            // 行动顺序队列，记录接下来该谁说话的座位号

	// --- 结算与历史 ---
	ShowdownSummary *ShowdownSummary `json:"showdown_summary,omitempty"` // 结算结果 (仅在 SHOWDOWN 阶段有值)
	LastAction      *ActionInfo      `json:"last_action,omitempty"`      // 刚刚发生的动作（用于前端播动画）
}
