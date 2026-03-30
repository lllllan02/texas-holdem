package domain

// SeatState 座位状态
type SeatState string

const (
	SeatEmpty    SeatState = "empty"    // 空座
	SeatOccupied SeatState = "occupied" // 有人
)

// Seat 座位：连接 Table 和 Player 的桥梁
type Seat struct {
	SeatNumber int       `json:"seat_number"` // 座位编号 (0 ~ MaxPlayers-1)
	Player     *Player   `json:"player"`      // 当前坐在该座位上的玩家（空座为 nil）
	State      SeatState `json:"state"`       // 座位状态
}
