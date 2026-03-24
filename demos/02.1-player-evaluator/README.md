# Demo 02.1: Player Evaluator (玩家牌力评估)

这个 Demo 是对 `02-evaluator` 的升级和封装。它引入了德州扑克中真实的“玩家 (Player)”和“公共牌 (Public Cards)”的概念，作为底层牌力计算器和上层游戏引擎之间的桥梁。

## 核心概念

在德州扑克中，牌力评估并不是简单地丢进去 7 张牌。真实的场景是：
- 每个玩家有 **2 张底牌 (Hole Cards)**（私有，只有自己知道或摊牌时亮出）。
- 桌面上有 **5 张公共牌 (Public Cards / Community Cards)**（所有人共享）。
- 玩家需要从这 7 张牌中选出最好的 5 张牌来比大小。

## 提供的功能

本包提供了两个核心方法，极大地方便了游戏引擎在结算阶段的调用：

### 1. `EvaluatePlayerHand`
计算单个玩家结合公共牌后的最佳牌型。
```go
func EvaluatePlayerHand(player Player, publicCards []core.Card) evaluator.HandResult
```
**内部逻辑**：自动将玩家的 2 张底牌和 5 张公共牌合并，然后调用 `02-evaluator` 中的 `Evaluate` 方法找出最佳的 5 张牌组合。

### 2. `FindWinners`
在一组玩家中，找出最终的获胜者（支持多人平局）。
```go
func FindWinners(players []Player, publicCards []core.Card) ([]Player, evaluator.HandResult)
```
**内部逻辑**：
- 遍历所有传入的玩家，计算他们各自的最佳牌型。
- 使用 `02-evaluator` 中的 `Compare` 方法两两比较。
- 维护一个“当前最大牌型”和“当前赢家列表”。
- 如果遇到更大的牌，则清空赢家列表并更新；如果遇到一样大的牌，则加入赢家列表（平分奖池）。

## 如何与 Demo 3 (Pot Calculation) 结合？

在真实的结算流程中，你会先调用 `03-pot-calculation` 算出多个奖池，然后遍历每个奖池，从 `Players` 列表中构造出对应的 `[]Player` 数组，最后调用本包的 `FindWinners` 找出该奖池的赢家并分钱。
