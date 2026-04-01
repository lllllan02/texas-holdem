# Texas Hold'em API & WebSocket 接口文档

本文档定义了德州扑克后端服务的所有 HTTP 接口以及 WebSocket 消息交互协议，供前端对接使用。

---

## 1. HTTP 接口 (REST API)

### 1.1 创建房间
- **URL**: `POST /api/v1/rooms`
- **Content-Type**: `application/json`
- **Request Body**:
  ```json
  {
    "owner_id": "user123", // 房主的 UserID
    "options": {           // 德州扑克引擎的房间配置 (可选)
      "player_count": 9,   // 座位数 (2~9)
      "small_blind": 10,   // 小盲注
      "big_blind": 20,     // 大盲注
      "initial_chips": 2000, // 初始筹码
      "action_timeout": 30 // 行动超时时间(秒)
    }
  }
  ```
- **Response** (200 OK):
  ```json
  {
    "room_id": "550e8400-e29b-41d4-a716-446655440000",
    "room_number": "550e84" // 6位短房号，用于加入房间
  }
  ```

### 1.2 获取用户信息 (含自动注册)
- **URL**: `GET /api/v1/users/:id`
- **说明**: 客户端在首次启动或本地生成了随机 `user_id` 后调用此接口。如果数据库中不存在该 `user_id`，服务端会自动为其注册一个默认账号（默认昵称为 `Player_xxx`，头像为空），并返回该账号信息。
- **Response** (200 OK):
  ```json
  {
    "id": "user123",
    "nickname": "Player_user123",
    "avatar": "https://example.com/avatar.png"
  }
  ```

### 1.3 更新用户信息
- **URL**: `PUT /api/v1/users/:id`
- **Content-Type**: `application/json`
- **Request Body**:
  ```json
  {
    "nickname": "TexasKing",
    "avatar": "https://example.com/new_avatar.png"
  }
  ```
- **Response** (200 OK): 返回更新后的 User 对象。

---

## 2. WebSocket 连接与消息外壳

### 2.1 建立连接
- **URL**: `ws://<host>:<port>/ws/:room_number?user_id=<your_user_id>`
- **说明**: 连接成功后，服务端会自动下发 `room.welcome` 消息，并将玩家以“旁观者”身份加入房间。

### 2.2 统一消息外壳 (Envelope)
所有的 WebSocket 消息（无论是客户端发给服务端，还是服务端发给客户端）都必须遵循以下 JSON 结构：

```json
{
  "type": "message.type",   // 消息类型，用于路由
  "reason": "trigger_reason", // 触发此消息的原因（可选，通常用于广播状态变化原因）
  "payload": {}             // 具体的业务数据结构
}
```

---

## 3. 房间级 WebSocket 消息

### 3.1 客户端 -> 服务端

| 消息类型 (`type`) | Payload 结构 | 说明 |
| :--- | :--- | :--- |
| `room.pause` | `{}` (空) | 房主请求暂停游戏 |
| `room.resume` | `{}` (空) | 房主请求恢复游戏 |

### 3.2 服务端 -> 客户端

#### `room.welcome` (欢迎加入)
连接成功后，服务端下发给该玩家的初始化信息。
```json
{
  "room_id": "uuid...",
  "room_number": "123456",
  "owner_id": "user123",
  "user": {
    "id": "user456",
    "nickname": "Alice",
    "avatar": "..."
  },
  "game_type": "texas"
}
```

#### `room.paused` (游戏暂停/恢复状态同步)
```json
{
  "is_paused": true,
  "user_id": "user123" // 触发暂停/恢复的玩家ID
}
```

#### `sys.error` (系统错误提示)
```json
{
  "error": "only owner can pause game"
}
```

---

## 4. 德州扑克专属 WebSocket 消息 (客户端 -> 服务端)

### 4.1 局外动作 (游戏未开始时可用)

#### `texas.sit_down` (请求落座)
```json
{
  "seat_number": 2 // 目标座位号 (0 ~ player_count-1)
}
```

#### `texas.stand_up` (请求站起/离开座位)
- **Payload**: `{}`

#### `texas.ready` (准备)
- **Payload**: `{}`

#### `texas.cancel` (取消准备)
- **Payload**: `{}`

### 4.2 局内动作 (游戏进行中，且轮到自己时可用)

#### `texas.action` (玩家打牌动作)
```json
{
  "action": "bet", // 动作类型: fold, check, call, bet, raise, allin
  "amount": 100    // 涉及的金额 (仅在 bet 和 raise 时有效)
}
```

---

## 5. 德州扑克专属 WebSocket 消息 (服务端 -> 客户端)

### 5.1 `texas.state_update` (全量状态更新快照)
这是**最核心**的消息。当牌桌状态发生任何变化（如有人坐下、发牌、下注、结算）时，服务端都会广播此消息。
**注意**：此快照是“千人千面”的，其他玩家的底牌会被隐藏，直到摊牌阶段。

```json
{
  "hand_count": 1,
  "button_seat": 0,
  "stage": "PREFLOP", // 阶段: WAITING, PREFLOP, FLOP, TURN, RIVER, SHOWDOWN
  "pot": 150,
  "current_bet": 20,
  "min_raise": 40,
  "board_cards": [    // 公共牌
    {"suit": 0, "rank": 14} // suit: 0=黑桃,1=红桃,2=方块,3=梅花; rank: 2~14(A)
  ],
  "current_player_index": 1,
  "is_paused": false,
  "action_order": [1, 2, 0],
  "side_pots": [],
  "last_action": {    // 刚刚发生的动作，用于前端播动画
    "player_id": "user123",
    "action": "raise",
    "amount": 40
  },
  "players": [        // 座位上的玩家列表
    {
      "id": "user123",
      "name": "Alice",
      "avatar": "...",
      "position": "SB", // 位置: BTN, SB, BB, UTG, MP, CO 等
      "seat_number": 0,
      "chips": 1990,
      "current_bet": 10,
      "state": "active", // 状态: waiting, ready, active, folded, allin
      "has_acted_this_round": true,
      "is_offline": false,
      "buy_in_count": 1,
      "hole_cards": [    // 自己的底牌（或摊牌时别人的底牌）
        {"suit": 1, "rank": 13},
        {"suit": 2, "rank": 13}
      ]
    }
  ],
  "showdown_summary": null // 仅在 SHOWDOWN 阶段存在，包含赢家和牌型信息
}
```

### 5.2 `texas.turn_notification` (轮到某人行动的通知)
当轮到某位玩家说话时，服务器会向该玩家**单独**发送此消息，告知其可用的操作及限制，前端据此渲染操作按钮和滑动条。

```json
{
  "player_id": "user456",
  "valid_actions": ["fold", "call", "raise", "allin"],
  "action_details": {
    "call_amount": 10, // 跟注需要补齐的金额
    "min_bet": 0,      // 最小下注额
    "max_bet": 0,      // 最大下注额
    "min_raise": 40,   // 最小加注额 (滑动条起点)
    "max_raise": 2000, // 最大加注额 (滑动条终点)
    "allin_amount": 2000 // All-in 对应的总金额
  },
  "timeout_seconds": 30 // 剩余思考时间
}
```

### 5.3 `texas.countdown` (倒计时通知)
用于开局倒计时或玩家行动倒计时的进度同步。

```json
{
  "player_id": "user456", // 正在行动的玩家ID（开局倒计时为空）
  "seconds": 3            // 剩余秒数
}
```

---

## 6. 枚举值参考 (Enums)

### 6.1 ActionType (动作类型)
- `fold`: 弃牌
- `check`: 过牌
- `call`: 跟注
- `bet`: 下注
- `raise`: 加注
- `allin`: 全下

### 6.2 HandStage (游戏阶段)
- `WAITING`: 等待开局
- `PREFLOP`: 翻牌前
- `FLOP`: 翻牌圈 (3张公共牌)
- `TURN`: 转牌圈 (4张公共牌)
- `RIVER`: 河牌圈 (5张公共牌)
- `SHOWDOWN`: 摊牌结算

### 6.3 PlayerState (玩家状态)
- `waiting`: 等待下一局开始（刚坐下或刚补充筹码）
- `ready`: 已准备，等待开局
- `active`: 正常参与本局，且还有筹码
- `folded`: 本局已弃牌
- `allin`: 本局已全下

### 6.4 Card Suit (花色)
- `0`: 黑桃 (Spades ♠)
- `1`: 红桃 (Hearts ♥)
- `2`: 方块 (Diamonds ♦)
- `3`: 梅花 (Clubs ♣)

### 6.5 Card Rank (点数)
- `2` ~ `9`: 对应数字 2~9
- `10`: T (Ten)
- `11`: J (Jack)
- `12`: Q (Queen)
- `13`: K (King)
- `14`: A (Ace)
