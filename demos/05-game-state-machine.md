# Demo 05: 游戏状态机与流程控制 (Game State Machine)

## 目标
实现德州扑克一局游戏的完整生命周期。控制发牌顺序、玩家下注轮转，以及验证玩家操作的合法性。

## 目录结构建议
```text
poker/
├── game/
│   ├── table.go       # 桌子状态管理
│   ├── player.go      # 玩家状态管理
│   ├── action.go      # 玩家操作定义 (Fold, Check, Call, Raise)
│   └── table_test.go  # 流程模拟测试
```

## 核心数据结构设计 (Go)
```go
// 游戏阶段枚举
type GameStage string
const (
    StageWaiting  GameStage = "WAITING"
    StagePreFlop  GameStage = "PRE_FLOP"
    StageFlop     GameStage = "FLOP"
    StageTurn     GameStage = "TURN"
    StageRiver    GameStage = "RIVER"
    StageShowdown GameStage = "SHOWDOWN"
)

// 一张游戏桌的状态
type Table struct {
    Players       []*Player
    Deck          *core.Deck
    CommunityCards []core.Card
    Stage         GameStage
    CurrentTurn   int // 当前轮到哪个玩家操作 (Index)
    Pot           int // 简化版奖池（先不考虑边池，方便跑通主流程）
    HighestBet    int // 当前轮的最高下注额
}
```

## 核心方法签名
- `func (t *Table) StartHand()`：洗牌，扣除盲注，发底牌，进入 PreFlop 阶段。
- `func (t *Table) HandleAction(playerID string, action ActionType, amount int) error`：处理玩家动作。
- `func (t *Table) NextStage()`：当一轮下注结束时，发公共牌并推进阶段。

## 具体测试用例 (table_test.go)
1. **非法操作拦截**: 
   - 还没轮到 Player A 时，A 尝试下注，返回 error。
   - 当前最高下注是 50，Player B 尝试 Call 20，返回 error (金额不足)。
2. **自动推进测试**: 模拟 3 个玩家。A 下注 10，B 跟注 10，C 跟注 10。断言 `Stage` 自动从 `PreFlop` 变为 `Flop`，并且 `CommunityCards` 增加了 3 张牌。
3. **提前结束测试**: A 下注，B 弃牌，C 弃牌。断言游戏直接结束，A 赢得奖池，不进入后续发牌阶段。