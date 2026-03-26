package texas

import (
	"fmt"
	"sync"

	"github.com/spf13/cast"
)

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

type Table struct {
	mu sync.RWMutex

	MaxPlayers int `json:"maxPlayers"`

	Seats   []*Player          `json:"seats"`
	Players map[string]*Player `json:"-"` // 内部映射，不暴露给前端

	// 1. 基础配置
	SmallBlind   int `json:"smallBlind"`   // 小盲注金额
	BigBlind     int `json:"bigBlind"`     // 大盲注金额
	InitialChips int `json:"initialChips"` // 玩家初始筹码

	// 2. 跨局状态
	ButtonIdx int `json:"buttonIdx"` // 庄家(Button)在 seats 数组中的索引，每局结束后移动

	// 3. 当前牌局
	Round *Round `json:"round"` // 当前正在进行的牌局（如果没有开始则为 nil）
}

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

func NewTable(param map[string]any) *Table {
	maxPlayers := cast.ToInt(param["maxPlayers"])
	smallBlind := cast.ToInt(param["smallBlind"])
	bigBlind := cast.ToInt(param["bigBlind"])
	initialChips := cast.ToInt(param["initialChips"])

	return &Table{
		MaxPlayers:   maxPlayers,
		Seats:        make([]*Player, maxPlayers),
		Players:      make(map[string]*Player),
		SmallBlind:   smallBlind,
		BigBlind:     bigBlind,
		InitialChips: initialChips,
		ButtonIdx:    0,   // 初始庄家位置
		Round:        nil, // 初始没有进行中的牌局
	}
}

// SitDown 玩家落座（支持未开始游戏时换座）
func (t *Table) SitDown(playerID string, seatIdx int) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if seatIdx < 0 || seatIdx >= t.MaxPlayers {
		return fmt.Errorf("invalid seat index")
	}
	if t.Seats[seatIdx] != nil {
		return fmt.Errorf("seat is already taken")
	}

	// 检查玩家是否已经坐下
	if p, exists := t.Players[playerID]; exists {
		// 如果游戏已经开始，不允许换座
		if t.Round != nil && t.Round.Stage != StageWaiting {
			return fmt.Errorf("cannot switch seats during an active game")
		}

		// 允许换座：先清空原来的座位
		for i, seat := range t.Seats {
			if seat != nil && seat.ID == playerID {
				t.Seats[i] = nil
				break
			}
		}
		// 坐到新座位
		t.Seats[seatIdx] = p
		return nil
	}

	// 新落座
	p := NewPlayer(playerID, t.InitialChips)
	t.Seats[seatIdx] = p
	t.Players[playerID] = p

	return nil
}

// StandUp 玩家站起
func (t *Table) StandUp(playerID string) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	_, exists := t.Players[playerID]
	if !exists {
		return fmt.Errorf("player is not seated")
	}

	// 找到玩家所在的座位并清空
	for i, seat := range t.Seats {
		if seat != nil && seat.ID == playerID {
			t.Seats[i] = nil
			break
		}
	}

	delete(t.Players, playerID)
	// TODO: 如果游戏正在进行中，站起可能需要自动弃牌等逻辑

	return nil
}

// GetSnapshot 为指定的玩家生成一份安全的数据快照
// playerID: 请求这份数据的玩家 ID。如果是旁观者，可以传空字符串 ""
func (t *Table) GetSnapshot(playerID string) *Table {
	t.mu.RLock()
	defer t.mu.RUnlock()

	// 1. 浅拷贝 Table (不拷贝锁)
	snap := Table{
		MaxPlayers:   t.MaxPlayers,
		SmallBlind:   t.SmallBlind,
		BigBlind:     t.BigBlind,
		InitialChips: t.InitialChips,
		ButtonIdx:    t.ButtonIdx,
		Round:        t.Round, // Round 内部如果没有锁，这里浅拷贝指针是安全的，因为我们不修改 Round
	}

	// 2. 隐藏不需要暴露的内部映射（虽然加了 json:"-"，但为了严谨还是置空）
	snap.Players = nil

	// 3. 深拷贝 Seats，以防修改影响原对象
	snap.Seats = make([]*Player, len(t.Seats))
	for i, p := range t.Seats {
		if p == nil {
			continue
		}

		// 浅拷贝玩家状态
		pSnap := *p

		// 核心防作弊逻辑：底牌可见性控制
		if p.Status == PlayerStatusNormal && !p.IsFolded && t.Round != nil {
			if p.ID == playerID || t.Round.Stage == StageShowdown {
				// 1. 是自己的牌，或者是结算亮牌阶段，可以看到真实的底牌
				// pSnap.HoleCards 保持原样
			} else {
				// 2. 是别人的牌，且游戏还在进行中，只能知道他有牌，但不知道是什么
				pSnap.HoleCards = nil
			}
		} else {
			// 没有参与游戏或者是弃牌状态，底牌不可见
			pSnap.HoleCards = nil
		}

		snap.Seats[i] = &pSnap
	}

	return &snap
}
