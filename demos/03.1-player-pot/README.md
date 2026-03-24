# Demo 03.1: Player Pot (玩家与奖池的终极融合)

这个 Demo 是 `03-pot-calculation` 的升级版，它融合了 `02.1-player-evaluator` 中的概念，提供了一个**一站式的德州扑克结算引擎**。

## 核心概念

在真实的德州扑克服务器中，我们不需要在“计算奖池”和“评估牌力”之间来回传递各种 ID 数组。相反，我们可以将所有信息封装在一个统一的 `Player` 对象中：

```go
type Player struct {
	ID        string
	HoleCards []core.Card // 玩家的2张底牌
	Bet       int         // 本局总共下注的筹码
	IsFolded  bool        // 是否已经弃牌

	// 缓存属性，避免多次比牌时重复计算
	BestHand  evaluator.HandResult
	Evaluated bool
}
```

## 提供的功能

### 1. `CalculatePots` (升级版)
直接接收 `[]*Player` 数组。它会在内部计算主池和边池，并将有资格分钱的 `*Player` 指针直接存入 `Pot` 结构体中。

### 2. `ResolveGame` (一站式结算)
这是游戏引擎在 River（河牌圈）结束时**唯一需要调用**的方法。

```go
func ResolveGame(players []*Player, publicCards []core.Card) map[string]int
```

**内部工作流：**
1. 调用 `CalculatePots` 将筹码切分为多个奖池。
2. 遍历每一个奖池。
3. 针对当前奖池内的 `Players`，调用他们的 `EvaluateHand` 方法（利用缓存避免重复计算）。
4. 比较牌力，找出赢家（处理平局）。
5. 将奖池的钱分给赢家。
6. 返回一个 `map[string]int`，告诉游戏引擎每个玩家最终赢了多少钱。

通过这个重构，上层业务逻辑变得极其简单，所有的数学计算和牌力评估都被完美地封装在了底层。
