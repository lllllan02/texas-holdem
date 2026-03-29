package texas

import (
	"fmt"

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

// IsGameRunning 判断当前是否处于游戏进行中（非等待且非结算阶段）
func (t *Table) IsGameRunning() bool {
	if t.Round == nil {
		return false
	}
	return t.Round.Stage != StageWaiting && t.Round.Stage != StageShowdown
}

// FindPlayerSeat 查找玩家的座位号，如果未找到则返回 -1
func (t *Table) FindPlayerSeat(playerID string) int {
	for i, p := range t.Seats {
		if p != nil && p.ID == playerID {
			return i
		}
	}
	return -1
}

// SitDown 玩家落座（支持未开始游戏时换座）
func (t *Table) SitDown(playerID string, seatIdx int) error {
	if seatIdx < 0 || seatIdx >= MaxTableSeats {
		return fmt.Errorf("invalid seat index")
	}
	if t.Seats[seatIdx] != nil {
		return fmt.Errorf("seat is already taken")
	}

	// 如果游戏已经开始，不允许新玩家坐下
	if t.IsGameRunning() {
		return fmt.Errorf("cannot sit down during an active game")
	}

	// 检查玩家是否已经坐下
	if p, exists := t.Players[playerID]; exists {
		// 允许换座：先清空原来的座位
		if oldSeatIdx := t.FindPlayerSeat(playerID); oldSeatIdx != -1 {
			t.Seats[oldSeatIdx] = nil
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
	_, exists := t.Players[playerID]
	if !exists {
		return fmt.Errorf("player is not seated")
	}

	// 找到玩家所在的座位并清空
	if seatIdx := t.FindPlayerSeat(playerID); seatIdx != -1 {
		t.Seats[seatIdx] = nil
	}

	delete(t.Players, playerID)
	// TODO: 如果游戏正在进行中，站起可能需要自动弃牌等逻辑

	return nil
}

// StartNewHand 开启新的一局游戏
func (t *Table) StartNewHand() error {
	// 1. 检查人数和准备状态
	activeCount := 0
	for _, p := range t.Seats {
		if p != nil {
			if !p.IsReady {
				return fmt.Errorf("player %s is not ready", p.ID)
			}
			activeCount++
		}
	}
	if activeCount < 2 {
		return fmt.Errorf("not enough players to start")
	}

	// 2. 初始化 Round
	t.Round = &Round{
		Stage:              StagePreFlop,
		ActivePlayersCount: activeCount,
		// TODO: 初始化牌堆、发底牌、扣除盲注等真实逻辑
	}

	// 开启新局时，重置所有落座玩家的弃牌和准备状态
	for _, p := range t.Seats {
		if p != nil {
			p.IsFolded = false
			p.IsReady = false // 游戏一旦开始，立刻清理准备状态
		}
	}

	// 临时：随便找个座位作为当前行动者，让前端能显示箭头
	for i, p := range t.Seats {
		if p != nil {
			t.Round.CurrentTurn = i
			t.Round.LastActionIdx = i // 初始时，LastActionIdx 设为第一个说话的人
			break
		}
	}

	return nil
}

// nextStageInternal 内部推进阶段方法，不加锁
func (t *Table) nextStageInternal() {
	if t.Round == nil {
		return
	}

	switch t.Round.Stage {
	case StagePreFlop:
		t.Round.Stage = StageFlop
		// TODO: 发3张公共牌
	case StageFlop:
		t.Round.Stage = StageTurn
		// TODO: 发第4张公共牌
	case StageTurn:
		t.Round.Stage = StageRiver
		// TODO: 发第5张公共牌
	case StageRiver:
		t.Round.Stage = StageShowdown
		// TODO: 结算比牌
	case StageShowdown:
		t.Round.Stage = StageWaiting
		// 暂时不将 t.Round 置为 nil，保留上一局的最终状态供前端展示
		// t.Round = nil // 牌局结束
		// TODO: 移动庄家按钮 (ButtonIdx)
	case StageWaiting:
		// 如果已经是 Waiting 状态，说明可能是提前强制结束（比如所有人都弃牌了），不需要再推进
		return
	}
}

// NextStage 推进游戏阶段
func (t *Table) NextStage() {
	t.nextStageInternal()
}

// AdvanceTurn 推进当前说话玩家，如果一圈结束，则调用 NextStage
func (t *Table) AdvanceTurn() {
	if t.Round == nil {
		return
	}

	// 检查当前未弃牌的玩家数量
	activeCount := 0
	lastActiveIdx := -1
	for i, p := range t.Seats {
		if p != nil && !p.IsFolded && p.Status == PlayerStatusNormal {
			activeCount++
			lastActiveIdx = i
		}
	}

	// 如果只剩一名玩家未弃牌，直接跳到结算阶段
	if activeCount <= 1 {
		// 如果还有人没说话，把说话权交给最后这个人（虽然没意义了）
		if lastActiveIdx != -1 {
			t.Round.CurrentTurn = lastActiveIdx
		}
		// 直接进入 Showdown 结算
		t.Round.Stage = StageShowdown
		t.nextStageInternal() // 这会把 Showdown 推进到 Waiting，并重置准备状态
		return
	}

	// 简单的轮转逻辑：找到下一个未弃牌的玩家
	startIdx := t.Round.CurrentTurn
	nextIdx := (startIdx + 1) % MaxTableSeats

	for nextIdx != startIdx {
		p := t.Seats[nextIdx]
		if p != nil && !p.IsFolded && p.Status == PlayerStatusNormal {
			t.Round.CurrentTurn = nextIdx
			
			// 骨架逻辑：如果转到了 LastActionIdx（一圈结束），就进入下一阶段
			// 假设 LastActionIdx 初始为庄家后面的那个人
			if nextIdx == t.Round.LastActionIdx {
				t.nextStageInternal()
				// 进入新阶段后，需要重置 LastActionIdx
				if t.Round != nil && t.Round.Stage != StageWaiting {
					t.Round.LastActionIdx = t.Round.CurrentTurn
				}
			}
			return
		}
		nextIdx = (nextIdx + 1) % MaxTableSeats
	}

	// 如果只剩一个人没弃牌，直接结束（上面的 activeCount 检查其实已经覆盖了这里，保留作为兜底）
	t.Round.Stage = StageShowdown
	t.nextStageInternal()
}

// GetSnapshot 为指定的玩家生成一份安全的数据快照
// playerID: 请求这份数据的玩家 ID。如果是旁观者，可以传空字符串 ""
func (t *Table) GetSnapshot(playerID string) *Table {
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

// CalculateAllowedActions 计算玩家当前允许的动作列表
func (t *Table) CalculateAllowedActions(playerID string, seatIdx int, isPaused bool) []string {
	actions := make([]string, 0)
	isGameRunning := t.IsGameRunning()

	// 如果未落座
	if seatIdx == -1 {
		// 检查是否有空座，且游戏未开始，如果有则允许坐下
		if !isGameRunning {
			hasEmptySeat := false
			for _, s := range t.Seats {
				if s == nil {
					hasEmptySeat = true
					break
				}
			}
			if hasEmptySeat {
				actions = append(actions, "texas.player.sit")
			}
		}
		return actions
	}

	player := t.Seats[seatIdx]

	if !isGameRunning && !player.IsReady {
		// 只有在游戏未开始，且玩家未准备的情况下，才允许站起
		actions = append(actions, "texas.player.stand")
	}

	if !isGameRunning {
		// 游戏未开始，可以准备或取消准备
		if !player.IsReady {
			actions = append(actions, "texas.player.ready")
		} else {
			actions = append(actions, "texas.player.cancel_ready")
		}
	} else {
		// 游戏进行中
		if player.IsFolded {
			// 已弃牌，只能看，不能做任何游戏动作
			return actions
		}

		// 如果游戏被房主暂停，不允许进行任何打牌操作
		if isPaused {
			return actions
		}

		// 检查是否轮到自己说话
		if t.Round.CurrentTurn == seatIdx {
			// 轮到自己，可以进行游戏操作
			actions = append(actions, "texas.game.fold")
			actions = append(actions, "texas.game.check") // 简化版：暂不判断是否真的能 check
			actions = append(actions, "texas.game.call")
			actions = append(actions, "texas.game.bet")
		}
	}

	return actions
}
