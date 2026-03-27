package texas

// PlayerStatus 表示玩家在牌桌上的持久状态
type PlayerStatus int

const (
	PlayerStatusNormal PlayerStatus = iota // 正常在座参与游戏
	PlayerStatusSitOut                     // 留座离桌（暂离，不参与下一局）
	PlayerStatusBroke                      // 破产（筹码为0）
)

type Player struct {
	ID string `json:"id"`

	// ==================================
	// 持久状态（跨局保留）
	// ==================================
	Chips    int          `json:"chips"`
	Status   PlayerStatus `json:"status"`
	IsReady  bool         `json:"isReady"` // 是否已准备好进入下一局

	// ==================================
	// 单局状态（每局 StartHand 或 NextStage 时需重置）
	// ==================================
	HoleCards        []Card `json:"holeCards,omitempty"` // 底牌，需要动态脱敏
	BetThisRound     int    `json:"betThisRound"`
	TotalBetThisHand int    `json:"totalBetThisHand"`

	HasActed bool `json:"hasActed"`
	IsFolded bool `json:"isFolded"`
	IsAllIn  bool `json:"isAllIn"`

	LastAction string `json:"lastAction"` // 上一次的动作记录
}

// NewPlayer 创建一个新玩家
func NewPlayer(id string, initialChips int) *Player {
	return &Player{
		ID:     id,
		Chips:  initialChips,
		Status: PlayerStatusNormal,
	}
}
