# 德州扑克状态机与游戏流程 (Game Flow)

本文档描述了一局德州扑克从“房间等待”到“结算分钱”的完整生命周期，以及状态机在每个阶段需要处理的核心逻辑。

## 1. 房间等待阶段 (Stage: WAITING)

这是游戏未开始时的默认状态。

*   **允许的动作**：`ROOM.JOIN` (坐下), `ROOM.LEAVE` (站起), `ROOM.READY` (准备), `ROOM.START` (房主开始), `ROOM.DISCONNECT` (断开), `ROOM.RECONNECT` (重连)。
*   **状态特征**：
    *   玩家可以自由加入或离开座位。
    *   玩家点击准备后，`IsReady` 变为 `true`。
    *   **断线处理**：如果玩家在这个阶段断开连接（`ROOM.DISCONNECT`），通常直接将其视为 `ROOM.LEAVE` 踢出房间，或者将其 `IsReady` 设为 `false` 并标记 `IsDisconnect = true`。
*   **流转条件**：
    *   房主发送 `ROOM.START` 动作。
    *   状态机校验：至少有 2 名玩家处于 `IsReady == true` 且 `Chips > 0` 且 `IsDisconnect == false`。
    *   校验通过后，调用 `StartHand()`，进入 `PRE_FLOP` 阶段。

---

## 1.5 局内断线处理 (In-Game Disconnect)

如果玩家在游戏进行中（`PRE_FLOP` 到 `RIVER`）断开连接，状态机必须保证游戏不被卡死：

1.  **标记状态**：收到 `ROOM.DISCONNECT` 后，将玩家的 `IsDisconnect` 设为 `true`。
2.  **托管/超时机制**：
    *   如果当前轮到该玩家说话（`CurrentTurn`），系统会启动一个倒计时（比如 15 秒）。
    *   如果倒计时结束玩家仍未重连，系统自动为其执行默认动作：
        *   如果可以 `CHECK`（当前最高下注为 0，或者他已经补齐了差额），则自动 `CHECK`。
        *   否则，自动 `FOLD`。
3.  **重连恢复**：如果玩家在倒计时结束前重新连接（`ROOM.RECONNECT`），将其 `IsDisconnect` 设为 `false`，并下发当前桌子的全量状态快照，玩家继续操作。

---

## 2. 游戏初始化 (StartHand 内部逻辑)

当触发开局时，状态机瞬间完成以下“幕后工作”，不等待玩家输入：

1.  **快照参与者**：遍历所有座位，将 `IsReady == true` 的玩家设为 `IsPlaying = true`，其他人设为 `false`（旁观）。重置所有人的局内状态（如 `IsFolded`, `HasActed`, `BetThisTurn`）。
2.  **确定位置**：
    *   移动庄家按钮 (`ButtonIdx`) 到下一个有效的 `IsPlaying` 玩家。
    *   确定小盲位 (`SmallBlindIdx`) 和大盲位 (`BigBlindIdx`)。
3.  **扣除盲注**：
    *   强制从小盲玩家的筹码中扣除 `SmallBlind` 金额，记入其 `BetThisTurn`。
    *   强制从大盲玩家的筹码中扣除 `BigBlind` 金额，记入其 `BetThisTurn`。
    *   更新桌子的 `HighestBet = BigBlind`，`LastRaiseDiff = BigBlind`。
4.  **洗牌与发牌**：初始化 `Deck` 并洗牌。给每个 `IsPlaying` 的玩家发 2 张底牌。
5.  **确定第一个说话的人 (CurrentTurn)**：
    *   通常是大盲位左边的第一个人（UTG）。
    *   记录 `LastActionIdx = BigBlindIdx`（大盲是最后一个“主动”下注的人，如果一圈转下来大家都只 Call 到大盲，大盲还有权选择 Check 或 Raise）。
6.  **进入下一阶段**：Stage 变更为 `PRE_FLOP`。

---

## 3. 翻牌前下注圈 (Stage: PRE_FLOP)

*   **允许的动作**：`GAME.FOLD`, `GAME.CALL`, `GAME.RAISE`。
    *   *注意：除非你是大盲且没人加注，否则通常不能 `CHECK`。*
*   **动作校验**：
    *   必须是 `CurrentTurn` 对应的玩家。
    *   `CALL`：需要补齐的筹码 = `HighestBet - 玩家已下注金额`。如果筹码不够，则转为 All-in。
    *   `RAISE`：目标总额必须 `>= HighestBet + LastRaiseDiff`。
*   **状态更新**：
    *   玩家执行动作后，更新其 `BetThisTurn`、`Chips`、`HasActed = true`。
    *   如果是 `RAISE`，更新桌子的 `HighestBet`、`LastRaiseDiff` 和 `LastActionIdx`。
*   **寻找下一个人**：顺时针寻找下一个 `IsPlaying == true && IsFolded == false && IsAllIn == false` 的玩家。

---

## 4. 阶段流转判定 (CheckRoundOver)

每次玩家执行完动作后，状态机都会检查当前下注圈是否结束。

*   **结束条件**：
    1.  **只剩一人**：除了 1 个玩家外，其他所有 `IsPlaying` 的玩家都 `IsFolded` 了。-> **提前结束，直接进入 SHOWDOWN 结算**。
    2.  **全员表态且平账**：所有未弃牌、未 All-in 的玩家，都满足 `HasActed == true` 且 `BetThisTurn == HighestBet`。
*   **轮次结束时的处理 (NextStage)**：
    1.  将所有玩家的 `BetThisTurn` 累加到 `TotalBetThisHand`，并收集到桌子的 `Pots` 中。
    2.  清空所有人的 `BetThisTurn`，重置 `HasActed = false`。
    3.  重置桌子的 `HighestBet = 0`，`LastRaiseDiff = BigBlind`。
    4.  根据当前阶段发公共牌：
        *   `PRE_FLOP` -> 发 3 张，进入 `FLOP`。
        *   `FLOP` -> 发 1 张，进入 `TURN`。
        *   `TURN` -> 发 1 张，进入 `RIVER`。
        *   `RIVER` -> 不发牌，进入 `SHOWDOWN`。
    5.  **确定新一轮第一个说话的人**：翻牌后，永远是从小盲位（或其左边第一个存活玩家）开始说话。

---

## 5. 翻牌/转牌/河牌圈 (Stage: FLOP / TURN / RIVER)

逻辑与 `PRE_FLOP` 基本一致，区别仅在于：
*   **允许的动作**：因为初始 `HighestBet = 0`，所以第一个说话的人可以 `GAME.CHECK`（过牌）。
*   **第一个说话的人**：不再是 UTG，而是从小盲位开始顺时针找第一个存活玩家。

---

## 6. 摊牌与结算阶段 (Stage: SHOWDOWN)

当河牌圈下注结束，或者所有人都 All-in/弃牌时进入此阶段。

1.  **计算奖池**：调用 Demo 03 的逻辑，根据每个人的 `TotalBetThisHand` 计算主池和可能的边池。
2.  **比牌**：
    *   如果是因为别人都弃牌而提前结束，剩下的唯一玩家直接赢走所有奖池，不需要比牌（也不需要亮底牌）。
    *   如果是正常打到最后，调用 Demo 02 的逻辑，结合 5 张公共牌和玩家的 2 张底牌，评估牌力并比大小。
3.  **分钱**：将奖池的筹码分配给赢家（处理平局平分）。
4.  **清理战场**：
    *   重置桌子状态（清空公共牌、奖池）。
    *   Stage 变更为 `WAITING`，等待下一局开始。