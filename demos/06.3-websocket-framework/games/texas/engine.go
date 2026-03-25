package texas

import (
	"log"

	game "github.com/lllllan02/texas-holdem/demos/05-game-state-machine"
	"github.com/lllllan02/texas-holdem/demos/06.3-websocket-framework/core"
)

// 游戏特定的 Action 定义
const (
	ActionSit  = "texas.sit"
	ActionBet  = "texas.bet"
	ActionFold = "texas.fold"
)

// TexasEngine 德州扑克游戏引擎，实现了 core.GameEngine 接口
type TexasEngine struct {
	RoomCtx core.RoomContext // 只持有受限的接口，而不是 *core.Room
	Table   *game.Table
}

// NewTexasEngine 创建一个新的德州扑克游戏引擎
func NewTexasEngine() *TexasEngine {
	return &TexasEngine{
		Table: &game.Table{},
	}
}

// OnInit 房间初始化时调用
func (e *TexasEngine) OnInit(roomCtx core.RoomContext) {
	e.RoomCtx = roomCtx
	log.Printf("[TexasEngine] 房间 %s 初始化德州扑克引擎\n", roomCtx.GetID())
}

// OnDestroy 房间销毁时调用
func (e *TexasEngine) OnDestroy() {
	log.Printf("[TexasEngine] 房间 %s 销毁，清理引擎资源\n", e.RoomCtx.GetID())
	// TODO: 这里可以清理游戏内部的定时器（比如玩家操作倒计时），防止 Goroutine 泄漏
}

// OnPlayerJoin 玩家加入房间时调用
func (e *TexasEngine) OnPlayerJoin(client *core.Client) {
	log.Printf("[TexasEngine] 玩家 %s 加入房间 %s\n", client.GetID(), e.RoomCtx.GetID())
	// 这里可以处理旁观者逻辑，或者自动坐下逻辑
}

// OnPlayerLeave 玩家离开房间时调用
func (e *TexasEngine) OnPlayerLeave(client *core.Client) {
	log.Printf("[TexasEngine] 玩家 %s 离开房间 %s\n", client.GetID(), e.RoomCtx.GetID())
	// 这里可以处理玩家离开座位、自动弃牌等逻辑
}

// HandleMessage 处理游戏特定消息
func (e *TexasEngine) HandleMessage(client *core.Client, msg core.ClientMessage) {
	switch msg.Action {
	case ActionSit:
		e.handleSit(client, msg)
	case ActionBet:
		e.handleBet(client, msg)
	case ActionFold:
		e.handleFold(client, msg)
	default:
		// 未知或未实现的游戏动作，原样广播
		e.broadcastDefault(client, msg)
	}
}

func (e *TexasEngine) handleSit(client *core.Client, msg core.ClientMessage) {
	// TODO: 结合 game.Table 的逻辑，让玩家坐下
	log.Printf("[TexasEngine] 玩家 %s 请求坐下\n", client.GetID())
	e.broadcastDefault(client, msg)
}

func (e *TexasEngine) handleBet(client *core.Client, msg core.ClientMessage) {
	// TODO: 结合 game.Table 的逻辑，处理下注
	log.Printf("[TexasEngine] 玩家 %s 请求下注: %s\n", client.GetID(), msg.Content)
	e.broadcastDefault(client, msg)
}

func (e *TexasEngine) handleFold(client *core.Client, msg core.ClientMessage) {
	// TODO: 结合 game.Table 的逻辑，处理弃牌
	log.Printf("[TexasEngine] 玩家 %s 请求弃牌\n", client.GetID())
	e.broadcastDefault(client, msg)
}

// broadcastDefault 默认的广播行为（临时占位，实际应该广播游戏状态快照）
func (e *TexasEngine) broadcastDefault(client *core.Client, msg core.ClientMessage) {
	// 游戏引擎现在可以直接调用 e.RoomCtx.Broadcast 或者 e.RoomCtx.SendTo
	// 而不需要自己去组装 JSON 和操作 Hub.Broadcast 通道
	e.RoomCtx.Broadcast(client, msg.Action, msg.Content)
}

// GetState 获取当前游戏状态快照
func (e *TexasEngine) GetState(client *core.Client) interface{} {
	// 这里返回一个通用的状态对象，实际项目中应该返回 Table 的完整快照
	return map[string]interface{}{
		"game":    "texas",
		"stage":   "WAITING",
		"message": "这是德州扑克的恢复状态",
	}
}
