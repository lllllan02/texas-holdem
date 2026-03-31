package texas

// Seat 座位：连接 Table 和 Player 的桥梁
type Seat struct {
	SeatNumber int     // 座位编号 (0 ~ MaxPlayers-1)
	Player     *Player // 当前坐在该座位上的玩家（空座为 nil）
}
