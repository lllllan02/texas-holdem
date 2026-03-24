# Demo 08: 最小可行性产品联调 (MVP Integration)

## 目标
将后端的 Go 游戏引擎、WebSocket 房间系统与前端的 React/Vue UI 彻底打通。实现一个真正可以从头玩到尾的联机德州扑克。

## 目录结构建议
```text
poker/
├── server/   # 包含 Demo 1-6 的所有 Go 代码
├── frontend/ # 包含 Demo 7 的所有前端代码
```

## 核心通信协议设计 (JSON over WebSocket)
**前端 -> 后端 (Client Action)**
```json
{
  "type": "PLAYER_ACTION",
  "payload": {
    "action": "RAISE",
    "amount": 500
  }
}
```

**后端 -> 前端 (State Sync)**
后端不发送增量更新，而是每次状态改变时，下发**对当前玩家可见**的完整桌面状态（类似 Demo 07 的 Mock JSON）。
*注意：后端在发送状态给 Player A 时，必须把 Player B 的底牌隐藏掉（清空数组），防止前端作弊（透视挂）。*

## 核心联调流程
1. **进入房间**: 前端打开 `http://localhost:5173/room/123456`，解析 URL 参数，发起 WebSocket 连接到 `ws://localhost:8080/ws/123456`。
2. **状态同步**: 连接成功后，后端立即推送当前的 `GameState`，前端根据 JSON 渲染界面。
3. **用户交互**: 轮到当前玩家时，前端底部操作栏亮起。玩家点击 "Call" 按钮，前端发送 `PLAYER_ACTION` 消息。
4. **后端处理**: Go 接收到消息，调用 `Table.HandleAction()`。状态机更新，进入下一阶段（例如发翻牌）。
5. **广播更新**: Go 将更新后的状态广播给房间内所有人。前端收到新 JSON，公共牌区域自动渲染出 3 张新牌。

## 验收标准
- 开启两个不同的浏览器（模拟两个玩家），加入同一个房间。
- 能够顺畅地完成：发底牌 -> 两人轮流下注 -> 发公共牌 -> 结算 -> 赢家筹码增加 -> 自动开始下一局。
- 没有任何卡死、状态不同步或筹码计算错误。