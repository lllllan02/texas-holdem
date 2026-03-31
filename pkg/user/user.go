package user

// User 代表系统中的一个真实用户（不论是否在参与游戏）
type User struct {
	ID       string `json:"id"`
	Nickname string `json:"nickname"`
	Avatar   string `json:"avatar"`
}
