package user

// User 代表系统中的一个真实用户（不论是否在参与游戏）
type User struct {
	ID       string `json:"id"`
	Nickname string `json:"nickname"`
	Avatar   string `json:"avatar"`
}

// GetUserByID 获取用户信息（此处为 Mock 实现，实际应从 DB/Redis 获取）
func GetUserByID(id string) *User {
	// TODO: 接入真实的存储层
	nickname := "Player_" + id
	if len(id) > 6 {
		nickname = "Player_" + id[:6]
	}
	return &User{
		ID:       id,
		Nickname: nickname,
		Avatar:   "", // 默认头像
	}
}
