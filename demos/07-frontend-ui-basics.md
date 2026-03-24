# Demo 07: 前端 UI 基础与数据绑定 (Frontend UI & Data Binding)

## 目标
使用现代前端框架（React 或 Vue）搭建游戏界面。此时不连接后端，完全依靠写死的假数据（Mock JSON）来渲染界面，验证 UI 布局和组件复用性。

## 目录结构建议 (前端工程)
```text
frontend/
├── src/
│   ├── components/
│   │   ├── Card.tsx       # 扑克牌组件
│   │   ├── PlayerSeat.tsx # 玩家座位组件 (头像、筹码、底牌、倒计时)
│   │   ├── Board.tsx      # 桌面公共牌区域
│   │   └── Controls.tsx   # 底部操作栏 (Fold, Check, Call, Raise)
│   ├── App.tsx            # 主页面，拼装上述组件
│   └── mockState.json     # 用于驱动 UI 的假数据
```

## 核心数据结构设计 (Mock JSON)
```json
{
  "stage": "TURN",
  "pot": 1500,
  "communityCards": [
    {"suit": "Spades", "rank": "A"},
    {"suit": "Hearts", "rank": "K"},
    {"suit": "Diamonds", "rank": "5"},
    {"suit": "Clubs", "rank": "9"}
  ],
  "players": [
    {
      "id": "p1", "name": "Alice", "chips": 5000, "bet": 200, 
      "isActive": true, "isTurn": true, 
      "cards": [{"suit": "Spades", "rank": "10"}, {"suit": "Spades", "rank": "J"}]
    },
    {
      "id": "p2", "name": "Bob", "chips": 3000, "bet": 200, 
      "isActive": true, "isTurn": false, 
      "cards": [] // 别人的牌不可见
    }
  ]
}
```

## 验收流程 (手动验证)
1. 运行 `npm run dev` 启动前端。
2. **视觉验证**: 确认桌面是椭圆形的绿色背景，玩家座位均匀分布在桌子周围。
3. **数据绑定验证**: 修改 `mockState.json` 中的 `stage` 为 "RIVER"，并增加一张公共牌，保存后页面应自动热更新，显示 5 张公共牌。
4. **状态指示验证**: `isTurn: true` 的玩家头像周围应该有高亮边框或倒计时动画，提示轮到他操作了。