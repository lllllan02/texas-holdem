# Demo 03: 奖池与边池计算 (Pot & Side Pots)

## 目标
处理德州扑克中最复杂的金融逻辑：当玩家筹码量不同且发生 All-in 时，如何正确切分主池和边池，并在结算时正确分钱。

## 目录结构建议
```text
poker/
├── core/
│   ├── pot.go      # 奖池计算逻辑
│   ├── pot_test.go # 分钱场景测试
```

## 核心数据结构设计 (Go)
```go
// 玩家在当前局的下注状态
type PlayerBet struct {
    PlayerID string
    Bet      int  // 本局总共下注的筹码
    IsAllIn  bool // 是否已经 All-in
    IsFolded bool // 是否已经弃牌
}

// 一个奖池（主池或边池）
type Pot struct {
    Amount          int      // 奖池总额
    EligiblePlayers []string // 有资格分这个池子的玩家 ID 列表
}
```

## 核心方法签名
- `func CalculatePots(bets []PlayerBet) []Pot`：根据所有玩家的下注情况，计算出当前的主池和所有边池。
- `func DistributePot(pot Pot, winners []string) map[string]int`：将一个奖池的钱分给赢家（处理平局平分及除不尽的零头）。

## 具体测试用例 (pot_test.go)
1. **单池场景**: A, B, C 各下注 100，无人 All-in。断言生成 1 个 Pot，金额 300，资格人 [A, B, C]。
2. **单边池场景**: A 筹码 100 (All-in)，B 筹码 200，C 筹码 200。
   - 断言生成 2 个 Pot：
   - 主池：金额 300 (A:100, B:100, C:100)，资格人 [A, B, C]。
   - 边池 1：金额 200 (B:100, C:100)，资格人 [B, C]。
3. **多边池场景**: A(10), B(20), C(30), D(40) 全部 All-in。断言生成 4 个正确的池子。
4. **弃牌场景**: A 下注 100 后弃牌，这 100 变成死钱 (Dead Money) 进入奖池，但 A 不在 EligiblePlayers 中。
5. **分钱零头测试**: 奖池 100，3 人平局。断言两人分得 33，一人分得 34（零头通常给位置最靠前的玩家）。