package texas

// ============================================================================
// 动作类型 (ActionType)
// ============================================================================

// ActionType 玩家在游戏中的动作类型
type ActionType string

const (
	ActionTypeFold  ActionType = "fold"  // 弃牌
	ActionTypeCheck ActionType = "check" // 过牌
	ActionTypeCall  ActionType = "call"  // 跟注
	ActionTypeBet   ActionType = "bet"   // 下注
	ActionTypeRaise ActionType = "raise" // 加注
	ActionTypeAllIn ActionType = "allin" // 全下
)

// ============================================================================
// 牌型等级 (HandRank)
// ============================================================================

// HandRank 牌型等级（数字越大牌型越大）
type HandRank int

const (
	HandRankHighCard      HandRank = iota // 0: 高牌 (High Card)
	HandRankPair                          // 1: 一对 (One Pair)
	HandRankTwoPair                       // 2: 两对 (Two Pair)
	HandRankThreeOfAKind                  // 3: 三条 (Three of a Kind)
	HandRankStraight                      // 4: 顺子 (Straight)
	HandRankFlush                         // 5: 同花 (Flush)
	HandRankFullHouse                     // 6: 葫芦 (Full House)
	HandRankFourOfAKind                   // 7: 四条 (Four of a Kind)
	HandRankStraightFlush                 // 8: 同花顺 (Straight Flush)
	HandRankRoyalFlush                    // 9: 皇家同花顺 (Royal Flush)
)

// String 返回牌型的英文名称
func (r HandRank) String() string {
	switch r {
	case HandRankHighCard:
		return "High Card"
	case HandRankPair:
		return "One Pair"
	case HandRankTwoPair:
		return "Two Pair"
	case HandRankThreeOfAKind:
		return "Three of a Kind"
	case HandRankStraight:
		return "Straight"
	case HandRankFlush:
		return "Flush"
	case HandRankFullHouse:
		return "Full House"
	case HandRankFourOfAKind:
		return "Four of a Kind"
	case HandRankStraightFlush:
		return "Straight Flush"
	case HandRankRoyalFlush:
		return "Royal Flush"
	default:
		return "Unknown"
	}
}

// ============================================================================
// 位置标识 (PositionType)
// ============================================================================

// PositionType 玩家在牌桌上的位置标识
// 德州扑克的位置是相对庄家(Button)而言的，不同的位置在不同的下注圈有不同的行动顺序优势。
type PositionType string

const (
	PositionBTN  PositionType = "BTN"   // 庄家 (Button)
	PositionSB   PositionType = "SB"    // 小盲注 (Small Blind)
	PositionBB   PositionType = "BB"    // 大盲注 (Big Blind)
	PositionUTG  PositionType = "UTG"   // 枪口位 (Under The Gun) - 大盲注左边第一个
	PositionUTG1 PositionType = "UTG+1" // 枪口位+1 (9人桌存在)
	PositionUTG2 PositionType = "UTG+2" // 枪口位+2 (9人桌存在)
	PositionMP   PositionType = "MP"    // 中位 (Middle Position)
	PositionMP1  PositionType = "MP+1"  // 中位+1
	PositionHJ   PositionType = "HJ"    // 劫持位 (Hijack) - 庄家右边第二个
	PositionCO   PositionType = "CO"    // 关门位 (Cutoff) - 庄家右边第一个
)

// PositionMap 预定义了 2~9 人桌的所有位置分配情况
// 索引为人数 (例如 PositionMap[6] 代表 6 人桌的位置数组)
// 数组内的元素按顺时针顺序排列，从庄家左边第一个活跃玩家（通常是 SB）开始
var PositionMap = [][]PositionType{
	nil,                                   // 0 人 (无效)
	nil,                                   // 1 人 (无效)
	{PositionSB, PositionBB},              // 2 人桌 (Heads-up): SB(BTN) -> BB
	{PositionSB, PositionBB, PositionBTN}, // 3 人桌
	{PositionSB, PositionBB, PositionUTG, PositionBTN},                                                                 // 4 人桌
	{PositionSB, PositionBB, PositionUTG, PositionCO, PositionBTN},                                                     // 5 人桌
	{PositionSB, PositionBB, PositionUTG, PositionHJ, PositionCO, PositionBTN},                                         // 6 人桌
	{PositionSB, PositionBB, PositionUTG, PositionUTG1, PositionHJ, PositionCO, PositionBTN},                           // 7 人桌
	{PositionSB, PositionBB, PositionUTG, PositionUTG1, PositionMP, PositionHJ, PositionCO, PositionBTN},               // 8 人桌
	{PositionSB, PositionBB, PositionUTG, PositionUTG1, PositionUTG2, PositionMP, PositionHJ, PositionCO, PositionBTN}, // 9 人桌
}

// ============================================================================
// 玩家状态 (PlayerState)
// ============================================================================

// PlayerState 玩家在当前局的状态
type PlayerState string

const (
	PlayerStateWaiting PlayerState = "waiting" // 等待下一局开始（刚坐下或刚补充筹码）
	PlayerStateReady   PlayerState = "ready"   // 已准备，等待开局
	PlayerStateActive  PlayerState = "active"  // 正常参与本局，且还有筹码
	PlayerStateFolded  PlayerState = "folded"  // 本局已弃牌
	PlayerStateAllIn   PlayerState = "allin"   // 本局已全下
)

// ============================================================================
// 游戏阶段 (HandStage)
// ============================================================================

// HandStage 单局游戏阶段
type HandStage string

const (
	HandStageWaiting  HandStage = "WAITING" // 等待开局阶段
	HandStagePreFlop  HandStage = "PREFLOP"
	HandStageFlop     HandStage = "FLOP"
	HandStageTurn     HandStage = "TURN"
	HandStageRiver    HandStage = "RIVER"
	HandStageShowdown HandStage = "SHOWDOWN"
)

// ============================================================================
// 座位状态 (SeatState)
// ============================================================================

// SeatState 座位状态
type SeatState string

const (
	SeatEmpty    SeatState = "empty"    // 空座
	SeatOccupied SeatState = "occupied" // 有人
)

// ============================================================================
// 扑克牌花色与点数 (Suit & Rank)
// ============================================================================

// Suit 扑克牌花色
type Suit int

const (
	SuitSpades   Suit = iota // 0: 黑桃 ♠
	SuitHearts               // 1: 红桃 ♥
	SuitDiamonds             // 2: 方块 ♦
	SuitClubs                // 3: 梅花 ♣
)

// String 返回花色的字符串表示 (供前端或日志使用)
func (s Suit) String() string {
	suitStrings := []string{"s", "h", "d", "c"}

	if s >= 0 && int(s) < len(suitStrings) {
		return suitStrings[s]
	}
	return "?"
}

// Rank 扑克牌点数 (2~14)
type Rank int

const (
	Rank2 Rank = 2
	Rank3 Rank = 3
	Rank4 Rank = 4
	Rank5 Rank = 5
	Rank6 Rank = 6
	Rank7 Rank = 7
	Rank8 Rank = 8
	Rank9 Rank = 9
	RankT Rank = 10 // 10 (Ten)
	RankJ Rank = 11 // Jack
	RankQ Rank = 12 // Queen
	RankK Rank = 13 // King
	RankA Rank = 14 // Ace
)

// String 返回点数的字符串表示 (如 "2", "T", "J", "A")
func (r Rank) String() string {
	rankStrings := []string{
		"?", "?", // 0, 1 不存在
		"2", "3", "4", "5", "6", "7", "8", "9", "T", "J", "Q", "K", "A",
	}

	if r >= Rank2 && r <= RankA {
		return rankStrings[r]
	}
	return "?"
}

// ============================================================================
// 德州扑克专属消息类型 (Texas MsgType)
// ============================================================================

// 客户端 -> 服务端
const (
	MsgTypeSitDown = "texas.sit_down" // 请求落座
	MsgTypeStandUp = "texas.stand_up" // 请求站起
	MsgTypeReady   = "texas.ready"    // 准备
	MsgTypeCancel  = "texas.cancel"   // 取消准备
	MsgTypeAction  = "texas.action"   // 游戏内动作 (fold, call, bet 等)
)

// 服务端 -> 客户端
const (
	MsgTypeStateUpdate      = "texas.state_update"      // 全量状态更新快照
	MsgTypeCountdown        = "texas.countdown"         // 倒计时通知（如开局倒计时）
	MsgTypeStartHand        = "texas.start_hand"        // 牌局正式开始
	MsgTypeDealHoleCards    = "texas.deal_hole_cards"   // 私发底牌
	MsgTypeTurnNotification = "texas.turn_notification" // 轮到某人行动的通知
)
