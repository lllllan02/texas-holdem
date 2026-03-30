package texas

// HandHistory 单局历史记录，用于对局结束后保存和查看
type HandHistory struct {
	HandID          string           `json:"hand_id"`           // 本局唯一ID
	RoomID        string           `json:"room_id"`          // 关联的桌子/房间ID
	HandCount       int              `json:"hand_count"`        // 第几局
	Timestamp       int64            `json:"timestamp"`         // 结束时间戳
	BoardCards      []string         `json:"board_cards"`       // 最终公共牌
	ShowdownSummary *ShowdownSummary `json:"showdown_summary"`  // 结算结果
	PlayerDeltas    []PlayerDelta    `json:"player_deltas"`     // 每个玩家的盈亏明细
}

// PlayerDelta 玩家单局盈亏明细
type PlayerDelta struct {
	PlayerID         string   `json:"player_id"`
	Name             string   `json:"name"`
	Position         string   `json:"position"`
	HoleCards        []string `json:"hole_cards,omitempty"`
	Delta            int      `json:"delta"`              // 本局盈亏（赢的钱 - 投入的钱）
	ChipsAfter       int      `json:"chips_after"`        // 结算后的最终筹码
	CumulativeProfit int      `json:"cumulative_profit"`  // 累计盈亏（可选，视业务需求）
	Result           string   `json:"result"`             // "WIN", "LOSE", "TIE"
}
