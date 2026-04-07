package texas

// ShowdownSummary 结算结果摘要 (严格对齐 poker.json 的 showdown_summary)
type ShowdownSummary struct {
	HandID        int                `json:"hand_id"`        // 局号 (如 Hand #3)
	TotalPot      int                `json:"total_pot"`      // 总底池大小
	BoardCards    []Card             `json:"board_cards"`    // 最终的公共牌（最多5张）
	ShowCards     bool               `json:"show_cards"`     // 是否需要亮牌
	SidePots      []*SidePot         `json:"side_pots"`      // 奖池分配结果
	PlayerResults []PlayerHandResult `json:"player_results"` // 每个玩家的最终结算结果（包含赢家和输家）
}

// PlayerHandResult 单个玩家的结算结果
type PlayerHandResult struct {
	PlayerID   string   `json:"player_id"`
	PlayerName string   `json:"player_name"` // 冗余昵称，防止玩家离桌后找不到
	NetProfit  int      `json:"net_profit"`  // 净盈亏（正数表示赢了多少，负数表示这局输了多少）
	Cards      []Card   `json:"cards"`       // 玩家底牌（如果未摊牌则为空）
	HandRank   HandRank `json:"hand_rank"`   // 最终牌型（如 4 代表顺子）
	IsWinner   bool     `json:"is_winner"`   // 是否赢得了任何奖池（主池或边池）
}

// buildShowdownSummary 统一构建结算摘要，计算每个玩家的净盈亏和赢家状态
func (t *Table) buildShowdownSummary(showCards bool) *ShowdownSummary {
	if t.CurrentHand == nil {
		return nil
	}

	summary := &ShowdownSummary{
		HandID:        len(t.Histories) + 1,
		TotalPot:      t.CurrentHand.Pot,
		BoardCards:    t.CurrentHand.BoardCards,
		ShowCards:     showCards,
		SidePots:      t.CurrentHand.SidePots,
		PlayerResults: make([]PlayerHandResult, 0),
	}

	// 找出所有赢家ID
	winnerMap := make(map[string]bool)
	for _, pot := range t.CurrentHand.SidePots {
		for _, wid := range pot.Winners {
			winnerMap[wid] = true
		}
	}

	// 遍历所有参与过本局的玩家（只要 ChipsBeforeHand > 0 或者参与了发牌）
	// 这里通过遍历所有座位，只要玩家不为空且状态不是 Waiting
	for _, p := range t.Seats {
		if p != nil && p.State != PlayerStateWaiting {
			// 计算净盈亏 = 当前筹码 - 局前筹码
			netProfit := p.Chips - p.ChipsBeforeHand

			// 获取牌型（如果未摊牌或提前结束，可能没有评估过牌型）
			rank := HandRankHighCard
			if p.BestHand != nil {
				rank = p.BestHand.Rank
			}

			// 如果不亮牌，且不是赢家，底牌可以隐藏（根据具体需求，这里统一返回底牌，前端根据 ShowCards 决定是否展示）
			var cards []Card
			// 只有在正常摊牌阶段 (showCards == true) 且玩家没有弃牌时，才展示其底牌
			// 提前结束时，所有人（包括赢家和弃牌玩家）都不亮牌
			if showCards && p.State != PlayerStateFolded {
				cards = p.HoleCards
			}

			summary.PlayerResults = append(summary.PlayerResults, PlayerHandResult{
				PlayerID:   p.User.ID,
				PlayerName: p.User.Nickname,
				NetProfit:  netProfit,
				Cards:      cards,
				HandRank:   rank,
				IsWinner:   winnerMap[p.User.ID],
			})
		}
	}

	return summary
}
