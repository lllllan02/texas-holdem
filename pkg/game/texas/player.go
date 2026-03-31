package texas

import "github.com/lllllan02/texas-holdem/pkg/user"

// Player 玩家：代表一个坐在牌桌上的用户，包含其在游戏中的筹码和状态
// 注意：玩家站起时，其筹码数据需要被妥善保存（比如存入数据库或房间级别的缓存中），
// 当其再次落座时，应恢复其之前的筹码量，而不是重新分配初始筹码。
type Player struct {
	User              *user.User  // 关联的全局用户实体（包含 ID, Nickname, Avatar 等基础信息）
	State             PlayerState // 玩家在当前局的状态
	Chips             int         // 玩家当前携带的筹码量
	CurrentBet        int         // 本轮（如 FLOP 圈）已经下注的筹码量
	HoleCards         []Card      // 玩家的底牌（通常为 2 张）
	HasActedThisRound bool        // 在当前下注圈是否已经表态过
	IsOffline         bool        // 是否已断线（断线不代表离开座位，只是进入托管状态）
}
