package machine

import "github.com/lllllan02/texas-holdem/demos/01-core"

// Player 表示一局游戏中的玩家状态
type Player struct {
	ID    string // 玩家唯一标识
	Name  string // 玩家昵称（可选，方便打印测试）
	Chips int    // 玩家手里剩下的总筹码 (Balance)

	// --- 房间级别的状态 ---
	IsPlaying    bool // 是否参与了当前这局游戏？(中途加入的、破产的为 false)
	IsReady      bool // 是否已准备好开始下一局 (用于房间等待阶段)
	IsDisconnect bool // 是否已断开连接 (断线重连机制)

	// --- 单局游戏内的状态 (仅当 IsPlaying == true 时有效) ---
	HoleCards        []core.Card // 底牌（2张）
	BetThisRound     int         // 当前这一轮(比如Flop圈)已经下注了多少筹码
	TotalBetThisHand int         // 本局总共下注的筹码（用于最终计算主池/边池）

	// 玩家当前的操作状态标志位
	IsFolded   bool   // 是否已经弃牌
	IsAllIn    bool   // 是否已经 All-in
	HasActed   bool   // 在当前下注轮是否已经行动过（用于判断一轮是否结束）
	LastAction string // 玩家最后一次执行的动作（如 "Check", "Raise 100"），用于前端展示
}
