package machine

import (
	"fmt"

	core "github.com/lllllan02/texas-holdem/demos/01-core"
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

// Table 表示一张游戏桌的状态
type Table struct {
	// 1. 基础配置
	SmallBlind int    // 小盲注金额
	BigBlind   int    // 大盲注金额
	HostID     string // 房主 ID，只有房主可以手动点击开始游戏

	// 2. 玩家与座位
	Players       []*Player // 房间里的所有玩家（按座位顺序，包含旁观/破产/新加入的）
	ButtonIdx     int       // 庄家(Button)在 Players 数组中的索引
	SmallBlindIdx int       // 小盲注(SB)在 Players 数组中的索引
	BigBlindIdx   int       // 大盲注(BB)在 Players 数组中的索引

	// 3. 牌局状态
	Deck           *core.Deck
	CommunityCards []core.Card
	Stage          GameStage // 当前处于什么阶段
	CurrentTurn    int       // 当前轮到哪个玩家说话（Players 数组的索引）

	// 4. 筹码与下注流转
	Pots               []int // 奖池（如果是简化版，可以只用 Pots[0] 代表主池。如果考虑边池，这里就是多个池）
	HighestBet         int   // 当前这一轮的最高下注额
	LastRaiseDiff      int   // 上一次的有效加注幅度（用于计算下一个人 Raise 的最小合法金额）
	MinRaiseAmount     int   // 当前合法的最小加注总额 (通常 = HighestBet + LastRaiseDiff)
	ActivePlayersCount int   // 当前未弃牌且未破产的活跃玩家数量（用于判断是否提前结束）

	// 5. 流程控制辅助字段
	// 记录是谁最后一次发起了有效的“主动行为”（大盲注或 Raise）。
	// 当一轮转下来，再次轮到这个人时，说明所有人都 Call 了，这一轮结束。
	LastActionIdx int
}

// ------------------------------------------------------------------------
// 游戏主流程控制 API (供外部调用)
// ------------------------------------------------------------------------

// StartHand 开始一局新游戏
// 职责：洗牌、确定庄家/盲注位置、扣除盲注、发底牌、设置初始状态为 StagePreFlop
func (t *Table) StartHand() error {
	fmt.Println("=== 游戏开始 ===")
	// TODO: 检查玩家人数是否足够 (>=2)
	// TODO: 初始化 Deck 并洗牌

	t.moveButtonAndBlinds()
	t.postBlind(t.SmallBlindIdx, t.SmallBlind)
	t.postBlind(t.BigBlindIdx, t.BigBlind)

	t.dealCards(2, false)

	t.Stage = StagePreFlop
	fmt.Println("-> 进入阶段:", t.Stage)
	// TODO: 将 CurrentTurn 设置为 BB 的下一个人 (UTG)
	return nil
}

// HandleAction 处理玩家的操作请求
// 职责：验证操作合法性、扣除筹码、更新状态、推动状态机
func (t *Table) HandleAction(action PlayerAction) error {
	fmt.Printf("玩家 [%s] 执行动作: %s, 金额: %d\n", action.PlayerID, action.Action, action.Amount)
	// TODO: 验证当前是否轮到该玩家 (action.PlayerID == t.Players[t.CurrentTurn].ID)
	// TODO: 验证操作类型是否合法 (比如不能在有人下注时 Check)

	// TODO: 根据 action.Action 执行具体逻辑 (Fold, Check, Call, Raise)
	//       - 如果是 Fold，更新 IsFolded，减少 ActivePlayersCount
	//       - 如果是 Call/Raise，调用 t.processBet() 扣除筹码
	//       - 更新玩家状态 (HasActed, LastAction)

	// TODO: 检查是否触发了提前结束条件 (ActivePlayersCount == 1)
	//       如果是，直接调用 t.EndHandEarly()

	if t.isRoundOver() {
		t.NextStage()
	} else {
		t.CurrentTurn = t.getNextPlayerIndex(t.CurrentTurn)
	}

	return nil
}

// ------------------------------------------------------------------------
// 状态机内部流转方法 (私有)
// ------------------------------------------------------------------------

// NextStage 推进游戏到下一个阶段
// 职责：收集筹码、发公共牌、重置轮次状态
func (t *Table) NextStage() {
	t.collectPots()

	switch t.Stage {
	case StagePreFlop:
		t.Stage = StageFlop
		t.dealCards(3, true)
	case StageFlop:
		t.Stage = StageTurn
		t.dealCards(1, true)
	case StageTurn:
		t.Stage = StageRiver
		t.dealCards(1, true)
	case StageRiver:
		t.Stage = StageShowdown
		t.ResolveShowdown()
		return
	}

	fmt.Println("-> 进入阶段:", t.Stage)

	// TODO: 重置所有玩家的 BetThisRound, HasActed, LastAction
	// TODO: 将 CurrentTurn 设置为 Button 之后的第一个未弃牌玩家
}

// ResolveShowdown 正常摊牌结算
// 职责：调用 evaluator 比牌，调用 pot 分钱，结束本局
func (t *Table) ResolveShowdown() {
	fmt.Println("=== 摊牌结算 ===")
	// TODO: 找出所有未弃牌的玩家
	// TODO: 结合他们的 HoleCards 和 CommunityCards，调用 evaluator.Evaluate 算出最佳牌型
	// TODO: 遍历 t.Pots，调用 pot.DistributePot 将奖池的钱分给赢家
	// TODO: 将赢家的钱加回到他们的 Chips 中
	t.Stage = StageWaiting
	fmt.Println("-> 游戏结束，等待下一局")
}

// EndHandEarly 提前结束游戏 (只剩一人未弃牌)
// 职责：将所有奖池和当前桌上的死钱直接发给最后剩下的那个人
func (t *Table) EndHandEarly() {
	fmt.Println("=== 提前结束 (其他玩家弃牌) ===")
	// TODO: 找出唯一一个 !IsFolded 的玩家
	t.collectPots()
	// TODO: 将所有 Pots 的钱全部加给这个玩家
	t.Stage = StageWaiting
	fmt.Println("-> 游戏结束，等待下一局")
}

// ------------------------------------------------------------------------
// 状态机流转的核心辅助方法
// ------------------------------------------------------------------------

// moveButtonAndBlinds 移动庄家和盲注位置 (StartHand 时调用)
func (t *Table) moveButtonAndBlinds() {
	fmt.Println("  [系统] 移动庄家和盲注位置")
	// TODO: 找到下一个参与游戏的玩家作为 Button
	// TODO: 找到 Button 后的下一个玩家作为 SB
	// TODO: 找到 SB 后的下一个玩家作为 BB
}

// postBlind 强制扣除盲注 (StartHand 时调用)
func (t *Table) postBlind(playerIdx int, amount int) {
	fmt.Printf("  [系统] 扣除盲注: %d\n", amount)
	// TODO: 检查玩家筹码是否足够，不够则触发强制 All-in
	// TODO: 扣除筹码，更新 BetThisRound 和 TotalBetThisHand
	// TODO: 更新 HighestBet 和 MinRaiseAmount
}

// processBet 处理玩家的下注行为 (HandleAction 时调用)
// 返回实际扣除的筹码量（可能因为筹码不足变成 All-in）
func (t *Table) processBet(playerIdx int, targetAmount int) int {
	// TODO: 计算还需要补多少差额 (targetAmount - p.BetThisRound)
	// TODO: 如果差额 >= 玩家剩余筹码，将玩家标记为 IsAllIn，并全扣
	// TODO: 否则正常扣除差额
	// TODO: 更新 p.BetThisRound 和 p.TotalBetThisHand
	// TODO: 如果 targetAmount > HighestBet，更新 HighestBet 和 MinRaiseAmount
	return 0
}

// dealCards 发牌 (StartHand 和 NextStage 时调用)
func (t *Table) dealCards(count int, isCommunity bool) {
	if isCommunity {
		fmt.Printf("  [发牌] 发出 %d 张公共牌\n", count)
	} else {
		fmt.Printf("  [发牌] 给每位玩家发 %d 张底牌\n", count)
	}
	// TODO: 从 t.Deck 中抽出 count 张牌
	// TODO: 如果 isCommunity 为 true，追加到 t.CommunityCards
	// TODO: 如果 isCommunity 为 false，则给每个未弃牌的玩家发 HoleCards
}

// getNextPlayerIndex 寻找下一个该说话的玩家
// 必须跳过 IsFolded == true 和 IsAllIn == true 的玩家
func (t *Table) getNextPlayerIndex(currentIndex int) int {
	fmt.Println("  [系统] 寻找下一个说话的玩家...")
	// TODO: 实现寻找下一个玩家的逻辑
	return 0
}

// isRoundOver 判断当前下注轮是否结束
// 条件：所有未弃牌且未 All-in 的玩家都 HasActed == true，并且他们的 BetThisRound 都等于 HighestBet
func (t *Table) isRoundOver() bool {
	// 为了跑通 Demo 测试，我们用一个假逻辑：如果当前是最后一个玩家，就认为轮次结束
	// 实际逻辑需要严格判断 HasActed 和 BetThisRound
	return false // 测试代码里我们会用 mock 覆盖或手动控制
}

// collectPots 在每一轮下注结束（进入下一阶段前）收集筹码
// 将所有玩家的 BetThisRound 收集起来，调用 Demo 3 的逻辑计算/合并到主池和边池中，
// 然后将玩家的 BetThisRound 清零，HasActed 设为 false，LastAction 清空。
func (t *Table) collectPots() {
	fmt.Println("  [系统] 收集本轮筹码，计算奖池")
	// TODO: 实现收集筹码并重置玩家轮次状态的逻辑
}
