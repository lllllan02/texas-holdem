package domain

// ============================================================================
// 服务端 -> 客户端：状态快照 (Snapshot / DTO)
// ============================================================================

// StateUpdateSnapshot 牌桌全量状态快照
// 它是 Table 和 Hand 的扁平化组合，专门用于发给前端渲染 UI
type StateUpdateSnapshot struct {
	// 基础与牌桌信息
	HandCount  int `json:"hand_count"`  // 第几局
	ButtonSeat int `json:"button_seat"` // 庄家座位号

	// 局内状态 (如果 stage == WAITING，这些字段可能是默认值)
	Stage              Stage            `json:"stage"`                // 当前阶段
	Pot                int              `json:"pot"`                  // 总底池
	CurrentBet         int              `json:"current_bet"`          // 本轮最高下注
	MinRaise           int              `json:"min_raise"`            // 最小加注额
	BoardCards         []Card           `json:"board_cards"`          // 公共牌
	CurrentPlayerIndex int              `json:"current_player_index"` // 当前行动玩家的座位号 (-1 表示无)
	ActionOrder        []int            `json:"action_order"`         // 后续行动顺序
	SidePots           []*SidePot       `json:"side_pots"`            // 边池
	ShowdownSummary    *ShowdownSummary `json:"showdown_summary,omitempty"` // 结算信息
	LastAction         *ActionInfo      `json:"last_action,omitempty"`      // 刚刚发生的动作（用于播动画）

	// 玩家列表 (千人千面：其他玩家的 HoleCards 会被设为 nil)
	Players []PlayerSnapshot `json:"players"`
}

// DealHoleCardsPayload 私发底牌的载荷
type DealHoleCardsPayload struct {
	Cards []Card `json:"cards"`
}

// PlayerSnapshot 玩家状态快照
type PlayerSnapshot struct {
	ID                string      `json:"id"`
	Name              string      `json:"name"`
	Position          string      `json:"position,omitempty"` // SB, BB 等 (未开局时可为空)
	SeatNumber        int         `json:"seat_number"`
	Chips             int         `json:"chips"`
	CurrentBet        int         `json:"current_bet"`
	State             PlayerState `json:"state"`                // waiting, active, folded, allin
	HasActedThisRound bool        `json:"has_acted_this_round"` // 本轮是否已行动
	HoleCards         []Card      `json:"hole_cards"`           // 底牌（如果是别人且未摊牌，这里必须是 nil）
}

// ActionInfo 记录刚刚发生的动作
type ActionInfo struct {
	PlayerID   string `json:"player_id,omitempty"`
	Action     string `json:"action,omitempty"` // "fold", "check", "call", "bet", "raise", "allin"
	Amount     int    `json:"amount,omitempty"`
	Stage      string `json:"stage,omitempty"`
	BoardCards []Card `json:"board_cards,omitempty"`
}

// ============================================================================
// 服务端 -> 客户端：行动通知
// ============================================================================

// TurnNotificationPayload 通知特定玩家行动
type TurnNotificationPayload struct {
	PlayerID       string        `json:"player_id"`
	ValidActions   []string      `json:"valid_actions"` // ["fold", "check", "call", "bet", "raise", "allin"]
	ActionDetails  ActionDetails `json:"action_details"`
	TimeoutSeconds int           `json:"timeout_seconds"`
}

// ActionDetails 玩家行动的金额限制
type ActionDetails struct {
	CallAmount  int `json:"call_amount,omitempty"`  // 跟注需要补齐的金额
	MinBet      int `json:"min_bet,omitempty"`      // 最小下注额
	MaxBet      int `json:"max_bet,omitempty"`      // 最大下注额
	MinRaise    int `json:"min_raise,omitempty"`    // 最小加注额
	MaxRaise    int `json:"max_raise,omitempty"`    // 最大加注额
	AllinAmount int `json:"allin_amount,omitempty"` // All-in 对应的金额
}

// ============================================================================
// 客户端 -> 服务端：玩家操作
// ============================================================================

// ClientActionPayload 客户端发来的操作指令
type ClientActionPayload struct {
	Action     string `json:"action"`                // "sit_down", "ready", "fold", "call", "raise" 等
	Amount     int    `json:"amount,omitempty"`      // 下注/加注金额
	SeatNumber int    `json:"seat_number,omitempty"` // 仅 sit_down 时使用
	BuyIn      int    `json:"buy_in,omitempty"`      // 仅 sit_down 时使用（带入筹码）
}
