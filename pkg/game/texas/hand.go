package texas

// Hand 单局游戏：德州扑克最核心的状态机，存活于“发底牌”到“结算”期间
type Hand struct {
	ID                 string     // 本局唯一ID
	Stage              HandStage  // 当前游戏阶段
	Deck               *Deck      // 牌堆（包含洗好的 52 张牌，不暴露给前端）
	BoardCards         []Card     // 桌面上的公共牌（最多 5 张）
	Pot                int        // 当前总奖池大小
	SidePots           []*SidePot // 边池列表（当有玩家 All-in 时产生拆分）
	CurrentBet         int        // 当前下注圈的最高下注额
	MinRaise           int        // 当前合法的最小加注额 (通常 = CurrentBet + 上一次加注的差额)
	ActionOrder        []int      // 行动顺序队列，记录接下来该哪些座位号说话
	CurrentPlayerIndex int        // 当前正在行动的玩家座位号
}
