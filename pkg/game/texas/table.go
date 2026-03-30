package texas

// Table 牌桌层：负责管理跨局的持久化状态
type Table struct {
	ID           string `json:"id"`            // 房间/桌子唯一ID
	MaxPlayers   int    `json:"max_players"`   // 最大玩家数上限（如 6 或 9）
	SmallBlind   int    `json:"small_blind"`   // 小盲注
	BigBlind     int    `json:"big_blind"`     // 大盲注
	InitialChips int    `json:"initial_chips"` // 初始筹码量（玩家坐下时的默认带入量）

	// 座位管理
	Seats []*Player `json:"seats"` // 长度等于 MaxPlayers，空座为 nil

	// 跨局状态
	ButtonSeat int            `json:"button_seat"` // 当前庄家座位号
	HandCount  int            `json:"hand_count"`  // 已经玩了多少局
	Histories  []*HandHistory `json:"histories"`   // 历史对局记录

	// 当前正在进行的牌局
	CurrentHand *Hand `json:"current_hand,omitempty"`
}
