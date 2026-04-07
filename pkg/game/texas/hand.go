package texas

// Hand 单局游戏：德州扑克最核心的状态机，存活于“发底牌”到“结算”期间
type Hand struct {
	ID                 string     // 本局唯一ID
	Stage              HandStage  // 当前游戏阶段
	Deck               *Deck      // 牌堆（包含洗好的 52 张牌，不暴露给前端）
	BoardCards         []Card     // 桌面上的公共牌（最多 5 张）
	Pot                int        // 当前总奖池大小
	SidePots           []*SidePot // 边池列表（当有玩家 All-in 时产生拆分）
	CurrentBet         int        // 本下注圈「台面封顶注」：任一玩家在本街已下筹码的最大值（跟注需凑到该额）
	MinRaise           int        // 最小加注**增量**：合法加注后的总注至少为 CurrentBet + MinRaise（全下例外）
	ActionOrder        []int      // 行动顺序队列，记录接下来该哪些座位号说话
	CurrentPlayerIndex int        // 当前正在行动的玩家座位号
	HandCount          int        // 当前是第几局
}

// AddBet 将玩家的下注增量加入底池，并自动更新当前街的封顶注 (CurrentBet) 和最小加注额 (MinRaise)
func (h *Hand) AddBet(player *Player, addAmount int) {
	actual := player.PlaceBet(addAmount)
	h.Pot += actual

	// 如果玩家的总下注超过了当前街的封顶注，更新封顶注
	if player.CurrentBet > h.CurrentBet {
		raiseDiff := player.CurrentBet - h.CurrentBet
		// 只有当加注额达到或超过当前的 MinRaise 时，才更新 MinRaise
		// （短码 All-in 可能无法达到 MinRaise，此时不更新）
		if raiseDiff >= h.MinRaise {
			h.MinRaise = raiseDiff
		}
		h.CurrentBet = player.CurrentBet
	}
}
