package machine

import "github.com/lllllan02/texas-holdem/demos/01-core"

// GameStage 表示一局游戏当前所处的阶段
type GameStage string

const (
	StageWaiting  GameStage = "WAITING"  // 等待玩家加入
	StagePreFlop  GameStage = "PRE_FLOP" // 翻牌前（发了底牌）
	StageFlop     GameStage = "FLOP"     // 翻牌圈（发3张公共牌）
	StageTurn     GameStage = "TURN"     // 转牌圈（发第4张）
	StageRiver    GameStage = "RIVER"    // 河牌圈（发第5张）
	StageShowdown GameStage = "SHOWDOWN" // 摊牌结算
)

// Table 表示一张游戏桌的状态
type Table struct {
	// 1. 基础配置
	SmallBlind int // 小盲注金额
	BigBlind   int // 大盲注金额
	HostID     string // 房主 ID，只有房主可以手动点击开始游戏

	// 2. 玩家与座位
	Players   []*Player // 房间里的所有玩家（按座位顺序，包含旁观/破产/新加入的）
	ButtonIdx int       // 庄家(Button)在 Players 数组中的索引
	SmallBlindIdx int   // 小盲注(SB)在 Players 数组中的索引
	BigBlindIdx   int   // 大盲注(BB)在 Players 数组中的索引

	// 3. 牌局状态
	Deck           *core.Deck
	CommunityCards []core.Card
	Stage          GameStage // 当前处于什么阶段
	CurrentTurn    int       // 当前轮到哪个玩家说话（Players 数组的索引）

	// 4. 筹码与下注流转
	Pots          []int // 奖池（如果是简化版，可以只用 Pots[0] 代表主池。如果考虑边池，这里就是多个池）
	HighestBet    int   // 当前这一轮的最高下注额
	LastRaiseDiff int   // 上一次的有效加注幅度（用于计算下一个人 Raise 的最小合法金额）

	// 5. 流程控制辅助字段
	// 记录是谁最后一次发起了有效的“主动行为”（大盲注或 Raise）。
	// 当一轮转下来，再次轮到这个人时，说明所有人都 Call 了，这一轮结束。
	LastActionIdx int
}
