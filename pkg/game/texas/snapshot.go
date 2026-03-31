package texas

// ============================================================================
// 服务端 -> 客户端：状态快照 (Snapshot / DTO)
// ============================================================================

// StateUpdateSnapshot 牌桌全量状态快照
// 它是 Table 和 Hand 领域模型的扁平化组合，专门用于发给前端渲染 UI。
// 注意：此快照是“千人千面”的，针对不同玩家生成时，其他玩家的敏感数据（如底牌）会被抹除。
type StateUpdateSnapshot struct {
	// --- 基础与牌桌信息 ---
	HandCount  int `json:"hand_count"`  // 当前是第几局游戏
	ButtonSeat int `json:"button_seat"` // 庄家 (Button) 所在的座位号

	// --- 局内状态 (如果 stage == WAITING，这些字段通常为默认值) ---
	Stage              HandStage        `json:"stage"`                      // 当前游戏阶段 (如 PREFLOP, FLOP 等)
	Pot                int              `json:"pot"`                        // 当前总底池金额
	CurrentBet         int              `json:"current_bet"`                // 本下注圈的最高下注额（跟注需要达到的目标金额）
	MinRaise           int              `json:"min_raise"`                  // 当前合法的最小加注额
	BoardCards         []Card           `json:"board_cards"`                // 桌面上的公共牌（最多 5 张）
	CurrentPlayerIndex int              `json:"current_player_index"`       // 当前正在行动的玩家座位号 (-1 表示无人行动)
	IsPaused           bool             `json:"is_paused"`                  // 游戏是否处于暂停状态
	ActionOrder        []int            `json:"action_order"`               // 后续行动顺序队列
	SidePots           []*SidePot       `json:"side_pots"`                  // 边池列表（当有玩家 All-in 时产生）
	ShowdownSummary    *ShowdownSummary `json:"showdown_summary,omitempty"` // 摊牌结算信息（仅在 SHOWDOWN 阶段存在）
	LastAction         *ActionInfo      `json:"last_action,omitempty"`      // 刚刚发生的动作（用于前端播放下注、弃牌等动画）

	// --- 玩家列表 ---
	Players []PlayerSnapshot `json:"players"` // 玩家状态列表（包含座位上的所有玩家）
}

// DealHoleCardsPayload 私发底牌的事件载荷
// 仅在发牌阶段单独发送给对应的玩家，用于播放发牌动画
type DealHoleCardsPayload struct {
	Cards []Card `json:"cards"` // 发给该玩家的 2 张底牌
}

// PlayerSnapshot 单个玩家的状态快照
type PlayerSnapshot struct {
	ID                string       `json:"id"`                   // 玩家唯一标识
	Name              string       `json:"name"`                 // 玩家昵称
	Avatar            string       `json:"avatar"`               // 玩家头像
	Position          PositionType `json:"position,omitempty"`   // 玩家位置 (如 "SB", "BB", "UTG"，未开局时可为空)
	SeatNumber        int          `json:"seat_number"`          // 座位号 (0 ~ MaxPlayers-1)
	Chips             int          `json:"chips"`                // 玩家当前持有的筹码量
	CurrentBet        int          `json:"current_bet"`          // 玩家在本轮已下注的金额
	State             PlayerState  `json:"state"`                // 玩家状态 (如 waiting, active, folded, allin)
	HasActedThisRound bool         `json:"has_acted_this_round"` // 玩家在本轮是否已经行动过
	IsOffline         bool         `json:"is_offline"`           // 玩家是否处于断线/托管状态
	BuyInCount        int          `json:"buy_in_count"`         // 玩家买入/分配筹码的总次数
	HoleCards         []Card       `json:"hole_cards"`           // 玩家底牌（注意：如果是其他玩家且未摊牌，此字段必须为 nil）
}

// ActionInfo 动作信息
// 用于记录游戏中发生的具体事件，方便前端展示操作提示或播放动画
type ActionInfo struct {
	PlayerID string     `json:"player_id,omitempty"` // 执行动作的玩家 ID
	Action   ActionType `json:"action,omitempty"`    // 动作类型 (如 fold, check, call, bet, raise, allin)
	Amount   int        `json:"amount,omitempty"`    // 动作涉及的金额 (如加注到多少)
}

// ============================================================================
// 服务端 -> 客户端：行动通知
// ============================================================================

// TurnNotificationPayload 轮到玩家行动的通知载荷
// 当轮到某位玩家说话时，服务器会向该玩家单独发送此消息，告知其可用的操作及限制
type TurnNotificationPayload struct {
	PlayerID       string        `json:"player_id"`       // 轮到行动的玩家 ID
	ValidActions   []ActionType  `json:"valid_actions"`   // 当前允许的动作列表 (如 [fold, check, bet])
	ActionDetails  ActionDetails `json:"action_details"`  // 具体动作的金额限制参数（用于前端渲染滑动条和按钮）
	TimeoutSeconds int           `json:"timeout_seconds"` // 剩余思考时间（秒）
}

// ActionDetails 玩家行动的金额限制明细
type ActionDetails struct {
	CallAmount  int `json:"call_amount,omitempty"`  // 跟注需要补齐的金额 (CurrentBet - Player.CurrentBet)
	MinBet      int `json:"min_bet,omitempty"`      // 最小下注额 (通常为大盲注)
	MaxBet      int `json:"max_bet,omitempty"`      // 最大下注额 (No-Limit 规则下通常等于玩家总筹码)
	MinRaise    int `json:"min_raise,omitempty"`    // 最小加注额 (当前下注额 + 最小加注幅度)
	MaxRaise    int `json:"max_raise,omitempty"`    // 最大加注额 (玩家总筹码)
	AllinAmount int `json:"allin_amount,omitempty"` // All-in 对应的总金额 (玩家当前所有筹码)
}

// CountdownPayload 倒计时通知载荷
type CountdownPayload struct {
	PlayerID string `json:"player_id,omitempty"` // 正在行动的玩家 ID（如果是开局倒计时，此字段为空）
	Seconds  int    `json:"seconds"`             // 剩余秒数
}

// ============================================================================
// 客户端 -> 服务端：玩家操作
// ============================================================================

// ClientActionPayload 客户端发给服务端的操作指令载荷
// 对应 core.Message 中的 Payload 字段 (当 MsgType 为 texas.action 时)
type ClientActionPayload struct {
	Action string `json:"action"`           // 玩家执行的动作 (如 "fold", "call", "bet", "raise", "allin")
	Amount int    `json:"amount,omitempty"` // 涉及的金额 (仅在 bet, raise 时有效)
}

// SitDownPayload 客户端请求落座的载荷
// 对应 core.Message 中的 Payload 字段 (当 MsgType 为 texas.sit_down 时)
type SitDownPayload struct {
	SeatNumber int `json:"seat_number"` // 目标座位号
}

// BuildPublicSnapshot 生成一个对所有人安全的公共快照（所有人的底牌均被隐藏，除非是摊牌阶段）
// 适用于平时游戏状态更新时的全服广播 (Broadcast)
func (t *Table) BuildPublicSnapshot() StateUpdateSnapshot {
	return t.buildSnapshotBase("") // 传入空字符串，意味着不为任何特定玩家展示底牌
}

// BuildPersonalSnapshot 为特定玩家生成专属的快照（包含该玩家自己的底牌）
// 适用于玩家刚进入房间、断线重连时的状态恢复 (SendTo)
func (t *Table) BuildPersonalSnapshot(viewerID string) StateUpdateSnapshot {
	return t.buildSnapshotBase(viewerID)
}

// buildSnapshotBase 是内部基础构建逻辑
func (t *Table) buildSnapshotBase(viewerID string) StateUpdateSnapshot {
	snap := StateUpdateSnapshot{
		HandCount:  t.HandCount,
		ButtonSeat: t.ButtonSeat,
		IsPaused:   t.IsPaused,
		Players:    make([]PlayerSnapshot, 0),
	}

	// 1. 映射单局状态 (如果正在游戏中的话)
	if t.CurrentHand != nil {
		snap.Stage = t.CurrentHand.Stage
		snap.Pot = t.CurrentHand.Pot
		snap.CurrentBet = t.CurrentHand.CurrentBet
		snap.MinRaise = t.CurrentHand.MinRaise
		snap.BoardCards = t.CurrentHand.BoardCards
		snap.CurrentPlayerIndex = t.CurrentHand.CurrentPlayerIndex
		snap.ActionOrder = t.CurrentHand.ActionOrder
		snap.SidePots = t.CurrentHand.SidePots
	} else {
		// 如果没有正在进行的牌局，说明在等待阶段
		snap.Stage = "WAITING"
		snap.CurrentPlayerIndex = -1
	}

	// 2. 映射玩家列表并进行【数据脱敏】
	// 获取所有活跃玩家的位置映射 (Key: UserID, Value: PositionType)
	playerPosMap := t.getPlayerPositions()

	for _, seat := range t.Seats {
		if seat.Player == nil {
			continue
		}

		p := seat.Player
		pSnap := PlayerSnapshot{
			ID:                p.User.ID,
			Name:              p.User.Nickname,
			Avatar:            p.User.Avatar,
			Position:          playerPosMap[p.User.ID], // 从 map 中获取计算好的位置
			SeatNumber:        seat.SeatNumber,
			Chips:             p.Chips,
			CurrentBet:        p.CurrentBet,
			State:             p.State,
			HasActedThisRound: p.HasActedThisRound,
			IsOffline:         p.IsOffline,
			BuyInCount:        p.BuyInCount,
		}

		// 【核心安全逻辑】：千人千面，只发该看的底牌
		// 什么时候能看到底牌？
		// 1. 观察者就是玩家本人 (viewerID 不为空且匹配)
		// 2. 游戏进入了 SHOWDOWN 摊牌阶段，且该玩家【没有弃牌】（参与了最终比牌）
		canSeeCards := false
		if viewerID != "" && p.User.ID == viewerID {
			canSeeCards = true
		} else if snap.Stage == HandStageShowdown && p.State != PlayerStateFolded {
			canSeeCards = true
		}

		if canSeeCards {
			pSnap.HoleCards = p.HoleCards
		}

		snap.Players = append(snap.Players, pSnap)
	}

	return snap
}

// getPlayerPositions 计算并返回当前所有参与游戏玩家的位置映射
// 返回一个 map，Key 是玩家的 UserID，Value 是计算出的位置 (如 SB, BB, BTN)
func (t *Table) getPlayerPositions() map[string]PositionType {
	if len(t.Seats) == 0 {
		return nil
	}

	// 1. 从庄家的下一个座位开始，顺时针收集所有参与本局游戏的活跃玩家
	activePlayers := make([]*Player, 0, t.MaxPlayers)
	startSeat := t.ButtonSeat + 1

	for i := 0; i < t.MaxPlayers; i++ {
		seatIdx := (startSeat + i) % t.MaxPlayers
		seat := t.Seats[seatIdx]

		if seat.Player != nil && seat.Player.State != PlayerStateWaiting {
			activePlayers = append(activePlayers, seat.Player)
		}
	}

	// 2. 根据活跃玩家人数，分配位置
	playerCount := len(activePlayers)
	if playerCount < 2 || playerCount > 9 {
		return nil
	}

	positions := PositionMap[playerCount]
	posMap := make(map[string]PositionType, playerCount)

	for i, p := range activePlayers {
		posMap[p.User.ID] = positions[i]
	}

	return posMap
}
