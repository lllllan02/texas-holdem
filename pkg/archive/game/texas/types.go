package texas

// StateUpdatePayload 状态更新的完整载荷
type StateUpdatePayload struct {
	Stage         GameStage      `json:"stage"`
	Pot           int            `json:"pot"`
	CurrentBet    int            `json:"current_bet"`
	ButtonSeat    int            `json:"button_seat"`
	Players       []PlayerInfo   `json:"players"`
	CurrentPlayer string         `json:"current_player,omitempty"`
	CurrentPlayerIdx int         `json:"current_player_index"`
	BoardCards    []string       `json:"board_cards"`
	SidePots      []SidePotInfo  `json:"side_pots"`
	LastAction    *ActionInfo    `json:"last_action,omitempty"`
}

type PlayerInfo struct {
	ID                string   `json:"id"`
	Name              string   `json:"name"`
	Position          string   `json:"position"`
	SeatNumber        int      `json:"seat_number"`
	Chips             int      `json:"chips"`
	CurrentBet        int      `json:"current_bet"`
	State             string   `json:"state"` // "active", "folded", "allin"
	HasActedThisRound bool     `json:"has_acted_this_round"`
	HoleCards         []string `json:"hole_cards,omitempty"`
	IsReady           bool     `json:"is_ready"`
}

type SidePotInfo struct {
	PotNumber       int      `json:"pot_number"`
	Amount          int      `json:"amount"`
	EligiblePlayers []string `json:"eligible_players"`
}

type ActionInfo struct {
	PlayerID   string   `json:"player_id,omitempty"`
	Action     string   `json:"action,omitempty"`
	Amount     int      `json:"amount,omitempty"`
	Stage      string   `json:"stage,omitempty"`
	BoardCards []string `json:"board_cards,omitempty"`
}

// TurnNotificationPayload 轮到玩家行动的通知
type TurnNotificationPayload struct {
	PlayerID       string        `json:"player_id"`
	Position       string        `json:"position"`
	ValidActions   []string      `json:"valid_actions"`
	ActionDetails  ActionDetails `json:"action_details"`
	TimeoutSeconds int           `json:"timeout_seconds"`
}

type ActionDetails struct {
	CallAmount  int `json:"call_amount"`
	MinRaise    int `json:"min_raise"`
	MaxRaise    int `json:"max_raise"`
	AllinAmount int `json:"allin_amount"`
}
