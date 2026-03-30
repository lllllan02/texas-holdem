# 德州扑克 WebSocket 协议设计参考

本文档基于现有的 `poker.json` WebSocket 通信日志，整理了德州扑克游戏在实际运行中的前后端交互流程与核心数据结构。这将作为我们后续开发 `texas` 功能（特别是后端 Go 结构体定义和 WebSocket 路由分发）的重要参考。

## 1. 核心消息类型分类

整个通信协议主要围绕以下三个维度展开：

1.  **核心状态同步**：服务端向客户端广播牌桌的完整状态。
2.  **玩家行动控制**：服务端通知玩家行动，客户端提交具体操作。
3.  **游戏生命周期与结算**：房间管理、牌局开始、最终比牌与筹码结算。

---

## 2. 详细消息解析

### 2.1 核心状态同步消息 (`state_update`)

这是系统中最重要、最频繁的消息，用于向客户端同步完整的牌桌状态。客户端根据此消息渲染整个游戏界面。

*   **消息类型 (type)**: `state_update`
*   **触发原因 (reason)**:
    *   `hand_started`: 新牌局开始
    *   `player_action`: 某个玩家执行了操作（如 fold/call/raise）
    *   `board_dealt`: 发公共牌（如翻牌 Flop、转牌 Turn、河牌 River）
    *   `showdown`: 摊牌结算阶段
    *   `player_rejoined`: 玩家重新连接/回到座位

*   **状态载荷 (payload) 核心字段**:
    *   **对局基础信息**:
        *   `match_id`: 房间/对局唯一标识
        *   `game_mode`: 游戏模式（如 "SIX_MAX" 六人桌）
        *   `stage`: 当前阶段（PREFLOP, FLOP, TURN, RIVER, SHOWDOWN）
        *   `pot`: 当前总底池大小
        *   `current_bet`: 当前轮的最高下注额（用于计算 Call 需要补齐的差额）
        *   `button_seat`: 庄家 (Button) 所在的座位号
        *   `hand_count`: 当前是第几手牌
    *   **玩家列表 (`players` 数组)**:
        *   `id` / `name`: 玩家标识
        *   `position`: 玩家位置（SB, BB, UTG, MP, CO, BTN）
        *   `seat_number`: 座位号
        *   `chips`: 玩家当前剩余筹码
        *   `current_bet`: 玩家在本轮已经下注的金额
        *   `state`: 玩家状态（`active` 活跃, `folded` 已弃牌）
        *   `has_acted_this_round`: 本轮是否已经行动过
        *   `hole_cards`: 底牌（数组如 `["Qs", "3d"]`，如果不可见则为 `null`）
    *   **行动控制**:
        *   `current_player` / `current_player_index`: 当前该谁行动
        *   `board_cards`: 公共牌（如 `["2h", "Tc", "5c", "6c"]`）
        *   `action_order`: 本轮后续的行动顺序
    *   **结算与边池**:
        *   `side_pots`: 边池信息（当有人 All-in 时产生）
        *   `showdown_summary`: 摊牌比牌结果（包含赢家、赢牌牌型、赢取金额等）
    *   **最新动作 (`last_action`)**:
        *   记录导致这次状态更新的上一个动作（例如：谁做了什么操作，或者发了什么公共牌），方便客户端做动画展示。

### 2.2 玩家行动控制消息

#### 服务端下发通知 (`turn_notification`)
通知某个特定玩家开始行动，并精确告知其当前允许的操作及金额限制。

*   **消息类型**: `turn_notification`
*   **核心字段**:
    *   `player_id`: 轮到的玩家 ID
    *   `valid_actions`: 允许的操作列表（例如 `["fold", "call", "raise", "allin"]` 或 `["check", "bet", "allin"]`）
    *   `action_details`: 详细金额限制
        *   `call_amount`: 跟注需要的金额
        *   `min_bet` / `max_bet`: 下注的最小/最大值
        *   `min_raise` / `max_raise`: 加注的最小/最大值
        *   `allin_amount`: All-in 对应的全下金额
    *   `timeout_seconds`: 操作超时时间（如 30 秒）

#### 客户端发送指令 (`action`)
玩家做出的具体决策，发送给服务端。

*   **消息类型**: `action`
*   **核心字段**:
    *   `match_id`: 房间 ID
    *   `action`: 动作类型（`fold`, `check`, `call`, `bet`, `raise`, `allin`）
    *   `amount`: 附带的金额（如果是 fold/check 则通常为 0）

### 2.3 游戏生命周期与结算消息

*   **房间/连接管理**:
    *   `join_room` (Client -> Server): 客户端请求加入房间
    *   `welcome` (Server -> Client): 服务端欢迎消息，确认连接
    *   `joined` (Server -> Client): 广播有玩家加入
*   **牌局控制**:
    *   `start_hand`: 标志新一手牌的初始化
*   **历史与结算 (`game.hand.history`)**:
    *   每局结束时发送，用于战绩统计。
    *   包含玩家的盈亏 (`delta`)、累计盈亏 (`cumulative_profit`)、最终筹码 (`chips_after`)、成牌牌型 (`hand_name`) 等数据。

---

## 3. 对后端 Go 代码设计的启示

基于上述分析，我们在设计 `pkg/game/texas` 时，应该重点关注以下数据结构的映射：

1.  **Table / Match 结构体**：需要维护 `Stage`, `Pot`, `BoardCards`, `ButtonSeat`, `CurrentBet` 等全局状态。
2.  **Player 结构体**：需要维护 `Chips`, `CurrentBet` (本轮下注), `State` (Active/Fold/Allin), `Position` (位置枚举), `HoleCards` 等。
3.  **Action 验证器**：后端需要一个专门的逻辑模块，在轮到玩家时，计算出他当前合法的 `ValidActions` 和 `ActionDetails` (最小加注额的计算在德州扑克中较为复杂，需要严格遵循规则)。
4.  **Side Pot (边池) 计算**：在 `Showdown` 阶段，如果有玩家 All-in，必须能够正确拆分主池和边池，并分别计算每个池子的赢家。
5.  **事件驱动机制**：后端的每一次状态变更（玩家操作、发牌、结算），都应该生成一个对应的 `reason`，并组装出完整的 `state_update` 广播给所有客户端。