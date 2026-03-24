# Demo 01: 核心扑克逻辑 (Core Poker Logic)

## 目标
实现德州扑克最基础的“物理”元素和发牌逻辑。我们需要用 Go 语言定义出扑克牌的实体，并保证洗牌和发牌的正确性。

## 目录结构建议
```text
poker/
├── demos/
│   ├── 01-core/
│   │   ├── card.go      # 定义花色、点数、扑克牌
│   │   ├── deck.go      # 定义牌堆、洗牌、发牌逻辑
│   │   └── deck_test.go # 单元测试
```

## 核心数据结构设计 (Go)
```go
// 花色枚举 (Spades, Hearts, Diamonds, Clubs)
type Suit int 

// 点数枚举 (2-10, J, Q, K, A)
type Rank int 

// 一张扑克牌
type Card struct {
    Suit Suit
    Rank Rank
}

// 牌堆
type Deck struct {
    Cards []Card
}
```

## 核心方法签名
- `func NewDeck() *Deck`：初始化一副包含 52 张牌的新牌堆，按顺序排列。
- `func (d *Deck) Shuffle()`：使用 `math/rand` 打乱牌堆（推荐使用 Fisher-Yates 算法）。
- `func (d *Deck) Draw(n int) ([]Card, error)`：从牌堆顶部抽出 n 张牌。如果牌不够了，返回 error。

## 具体测试用例 (deck_test.go)
1. **TestNewDeck**: 验证 `NewDeck()` 返回的牌堆长度是否严格等于 52，且没有重复的牌。
2. **TestShuffle**: 调用 `Shuffle()` 后，验证牌的顺序是否与初始状态不同（且多次 Shuffle 结果不同）。
3. **TestDraw**: 
   - 抽 2 张牌（模拟发底牌），验证返回 2 张牌，且牌堆剩余 50 张。
   - 抽 5 张牌（模拟发公共牌），验证牌堆数量正确减少。
   - 尝试从空牌堆抽牌，验证是否正确返回 error。