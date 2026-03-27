package texas

// Round 表示单局游戏的状态
type Round struct {
	Stage          GameStage `json:"stage"`          // 当前处于什么阶段
	Deck           *Deck     `json:"-"`              // 牌堆，绝对不能暴露给前端
	CommunityCards []Card    `json:"communityCards"` // 公共牌

	// 盲注位置
	SmallBlindIdx int `json:"smallBlindIdx"` // 小盲注(SB)在 seats 数组中的索引
	BigBlindIdx   int `json:"bigBlindIdx"`   // 大盲注(BB)在 seats 数组中的索引

	// 筹码与下注流转
	Pots               []int `json:"pots"`               // 奖池
	HighestBet         int   `json:"highestBet"`         // 当前这一轮的最高下注额
	LastRaiseDiff      int   `json:"lastRaiseDiff"`      // 上一次的有效加注幅度
	MinRaiseAmount     int   `json:"minRaiseAmount"`     // 当前合法的最小加注总额
	ActivePlayersCount int   `json:"activePlayersCount"` // 当前未弃牌且未破产的活跃玩家数量

	// 流程控制辅助字段
	CurrentTurn   int `json:"currentTurn"`   // 当前轮到哪个玩家说话（seats 数组的索引）
	LastActionIdx int `json:"lastActionIdx"` // 记录是谁最后一次发起了有效的“主动行为”
}
