# Demo 02: 牌型评估与比大小 (Hand Evaluation)

## 目标
实现德州扑克的“裁判”系统。给定任意 7 张牌，找出最大的 5 张牌组合，并能比较两组牌的大小。

## 目录结构建议
```text
poker/
├── demos/
│   ├── 02-evaluator/
│   │   ├── evaluator.go      # 牌型评估算法
│   │   ├── evaluator_test.go # 大量的边界测试用例
```

## 核心数据结构设计 (Go)
```go
// 牌型等级枚举 (从高牌到皇家同花顺，值越大牌型越大)
type HandRank int 

const (
    HighCard HandRank = iota
    Pair
    TwoPair
    ThreeOfAKind
    Straight
    Flush
    FullHouse
    FourOfAKind
    StraightFlush
    RoyalFlush
)

// 评估结果
type HandResult struct {
    Rank       HandRank // 牌型等级
    BestCards  []Card   // 构成该牌型的最佳 5 张牌
    Kickers    []Card   // 踢脚牌（用于同级别牌型比大小）
}
```

## 核心方法签名
- `func Evaluate(cards []Card) HandResult`：输入 5-7 张牌，返回最佳牌型结果。
- `func Compare(a, b HandResult) int`：比较两组牌，返回 1 (a赢), -1 (b赢), 0 (平局)。

## 具体测试用例 (evaluator_test.go)
这是整个游戏最需要严谨测试的地方：
1. **识别测试**: 传入特定的 7 张牌，断言 `Evaluate` 能正确识别出 `FullHouse` (葫芦) 或 `Flush` (同花) 等。
2. **顺子边界测试**: 测试 A-2-3-4-5 (最小顺子) 和 10-J-Q-K-A (最大顺子) 能否被正确识别。
3. **比大小测试**: 
   - 同花 vs 顺子 -> 同花赢。
   - 两个对子比大小：对 A 赢 对 K。
   - 相同对子比踢脚牌：A-A-K-7-2 赢 A-A-Q-J-9。
4. **平局测试**: 两人底牌不同，但公共牌是皇家同花顺，断言 `Compare` 返回 0 (平局)。