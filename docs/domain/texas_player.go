package domain

// PlayerState 玩家在当前局的状态
type PlayerState string

const (
	PlayerWaiting PlayerState = "waiting" // 等待下一局开始（刚坐下或刚补充筹码）
	PlayerReady   PlayerState = "ready"   // 已准备，等待开局
	PlayerActive  PlayerState = "active"  // 正常参与本局，且还有筹码
	PlayerFolded  PlayerState = "folded"  // 本局已弃牌
	PlayerAllIn   PlayerState = "allin"   // 本局已全下
)

// Player 玩家：代表一个坐在牌桌上的用户，包含其在游戏中的筹码和状态
// 注意：玩家站起时，其筹码数据需要被妥善保存（比如存入数据库或房间级别的缓存中），
// 当其再次落座时，应恢复其之前的筹码量，而不是重新分配初始筹码。
type Player struct {
	ID                string      `json:"id"`                   // 玩家唯一标识 (关联到全局 User ID)
	Name              string      `json:"name"`                 // 玩家昵称
	Chips             int         `json:"chips"`                // 玩家当前携带的筹码量
	State             PlayerState `json:"state"`                // 玩家在当前局的状态
	Position          string      `json:"position"`             // 本局的位置标识 (SB, BB, UTG, MP, CO, BTN)
	CurrentBet        int         `json:"current_bet"`          // 本轮（如 FLOP 圈）已经下注的筹码量
	HoleCards         []Card      `json:"hole_cards,omitempty"` // 玩家的底牌（通常为 2 张）
	HasActedThisRound bool        `json:"has_acted_this_round"` // 在当前下注圈是否已经表态过
}
