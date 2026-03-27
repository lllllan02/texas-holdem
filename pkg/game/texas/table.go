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

// MaxTableSeats 德州扑克桌子的最大座位数
const MaxTableSeats = 9

type Table struct {
	mu sync.RWMutex

	Seats   [MaxTableSeats]*Player `json:"seats"` // 固定 9 人桌
	Players map[string]*Player     `json:"-"`     // 内部映射，不暴露给前端

	// 1. 基础配置
	SmallBlind   int `json:"smallBlind"`   // 小盲注金额
	BigBlind     int `json:"bigBlind"`     // 大盲注金额
	InitialChips int `json:"initialChips"` // 玩家初始筹码

	// 2. 跨局状态
	ButtonIdx int `json:"buttonIdx"` // 庄家(Button)在 seats 数组中的索引，每局结束后移动

	// 3. 当前牌局
	Round *Round `json:"round"` // 当前正在进行的牌局（如果没有开始则为 nil）
}

func NewTable(param map[string]any) *Table {
	smallBlind := cast.ToInt(param["smallBlind"])
	bigBlind := cast.ToInt(param["bigBlind"])
	initialChips := cast.ToInt(param["initialChips"])

	return &Table{
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

	if seatIdx < 0 || seatIdx >= MaxTableSeats {
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
		SmallBlind:   t.SmallBlind,
		BigBlind:     t.BigBlind,
		InitialChips: t.InitialChips,
		ButtonIdx:    t.ButtonIdx,
		Round:        t.Round, // Round 内部如果没有锁，这里浅拷贝指针是安全的，因为我们不修改 Round
	}

	// 2. 隐藏不需要暴露的内部映射（虽然加了 json:"-"，但为了严谨还是置空）
	snap.Players = nil

	// 3. 深拷贝 Seats，以防修改影响原对象
	// 注意：因为 Seats 是定长数组，我们需要显式地给每个元素赋值
	for i, p := range t.Seats {
		if p == nil {
			// 如果原座位为空，快照里的座位也必须显式置空
			snap.Seats[i] = nil
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
