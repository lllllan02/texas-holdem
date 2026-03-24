package machine

import "github.com/lllllan02/texas-holdem/demos/01-core"

// TableSnapshot 是发送给客户端的全局状态快照 (脱敏后)
type TableSnapshot struct {
	Stage          GameStage         `json:"stage"`
	CommunityCards []core.Card       `json:"communityCards"`
	Pots           []int             `json:"pots"`
	HighestBet     int               `json:"highestBet"`
	CurrentTurn    int               `json:"currentTurn"`
	ButtonIdx      int               `json:"buttonIdx"`
	SmallBlindIdx  int               `json:"smallBlindIdx"`
	BigBlindIdx    int               `json:"bigBlindIdx"`
	Players        []PlayerSnapshot  `json:"players"`
	
	// 注意：这里绝对不能包含 Deck (牌堆) !
}

// PlayerSnapshot 是发送给客户端的单个玩家状态快照 (脱敏后)
type PlayerSnapshot struct {
	ID               string      `json:"id"`
	Name             string      `json:"name"`
	Chips            int         `json:"chips"`
	IsPlaying        bool        `json:"isPlaying"`
	IsReady          bool        `json:"isReady"`
	IsDisconnect     bool        `json:"isDisconnect"`
	IsFolded         bool        `json:"isFolded"`
	IsAllIn          bool        `json:"isAllIn"`
	BetThisTurn      int         `json:"betThisTurn"`
	TotalBetThisHand int         `json:"totalBetThisHand"`
	
	// 核心脱敏字段：
	// 如果是自己，或者是 Showdown 阶段亮牌了，这里会有具体的牌。
	// 如果是别人且没亮牌，这里要么是 nil，要么是两张代表“背面”的假牌。
	HoleCards []core.Card `json:"holeCards,omitempty"` 
}

// GetSnapshot 为指定的玩家生成一份安全的数据快照
// requestingPlayerID: 请求这份数据的玩家 ID。如果是旁观者，可以传空字符串 ""
func (t *Table) GetSnapshot(requestingPlayerID string) TableSnapshot {
	snap := TableSnapshot{
		Stage:          t.Stage,
		CommunityCards: t.CommunityCards,
		Pots:           t.Pots,
		HighestBet:     t.HighestBet,
		CurrentTurn:    t.CurrentTurn,
		ButtonIdx:      t.ButtonIdx,
		SmallBlindIdx:  t.SmallBlindIdx,
		BigBlindIdx:    t.BigBlindIdx,
		Players:        make([]PlayerSnapshot, len(t.Players)),
	}

	for i, p := range t.Players {
		pSnap := PlayerSnapshot{
			ID:               p.ID,
			Name:             p.Name,
			Chips:            p.Chips,
			IsPlaying:        p.IsPlaying,
			IsReady:          p.IsReady,
			IsDisconnect:     p.IsDisconnect,
			IsFolded:         p.IsFolded,
			IsAllIn:          p.IsAllIn,
			BetThisTurn:      p.BetThisTurn,
			TotalBetThisHand: p.TotalBetThisHand,
		}

		// 核心防作弊逻辑：底牌可见性控制
		if p.IsPlaying && !p.IsFolded {
			if p.ID == requestingPlayerID || t.Stage == StageShowdown {
				// 1. 是自己的牌，或者是结算亮牌阶段，可以看到真实的底牌
				pSnap.HoleCards = p.HoleCards
			} else {
				// 2. 是别人的牌，且游戏还在进行中，只能知道他有牌，但不知道是什么
				// 这里我们可以返回一个空数组，或者返回两张特殊的“背面牌”。
				// 为了让前端好处理，通常返回 nil，前端根据 IsPlaying && !IsFolded 自动画两张背面牌。
				pSnap.HoleCards = nil 
			}
		}

		snap.Players[i] = pSnap
	}

	return snap
}
