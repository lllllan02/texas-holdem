package texas

// PlayerStatus 表示玩家在牌桌上的持久状态
type PlayerStatus int

const (
	PlayerStatusNormal PlayerStatus = iota // 正常在座参与游戏
	PlayerStatusSitOut                     // 留座离桌（暂离，不参与下一局）
	PlayerStatusBroke                      // 破产（筹码为0）
)

type Player struct {
	ID     string `json:"id"`
	Name   string `json:"name"`   // 玩家昵称
	Avatar string `json:"avatar"` // 玩家头像

	// ==================================
	// 1. 持久状态（跨局保留，玩家离开房间才销毁）
	// ==================================
	Chips   int          `json:"chips"`
	Status  PlayerStatus `json:"status"`  // Normal, SitOut, Broke
	IsReady bool         `json:"isReady"` // 是否已准备好进入下一局

	// ==================================
	// 2. 单局状态（每局 StartNewHand 时重置）
	// ==================================
	Position         string `json:"position"` // 玩家位置：SB, BB, UTG, BTN 等
	HoleCards        []Card `json:"holeCards,omitempty"`
	TotalBetThisHand int    `json:"totalBetThisHand"` // 本局总下注（用于边池计算）
	IsFolded         bool   `json:"isFolded"`         // 是否已弃牌
	IsAllIn          bool   `json:"isAllIn"`          // 是否已全下

	// ==================================
	// 3. 单轮状态（每个下注圈 PreFlop, Flop, Turn, River 开始时重置）
	// ==================================
	BetThisRound int    `json:"betThisRound"` // 本轮已下注金额
	HasActed     bool   `json:"hasActed"`     // 本轮是否已经表态过（用于判断大盲过牌、一圈是否结束）
	LastAction   string `json:"lastAction"`   // 本轮上一次的动作记录 (Fold, Call, Raise, Check)
}

// NewPlayer 创建一个新玩家
func NewPlayer(id string, initialChips int) *Player {
	return &Player{
		ID:     id,
		Name:   "Player-" + id, // 默认给个名字，后续可以从业务层传入
		Chips:  initialChips,
		Status: PlayerStatusNormal,
	}
}
