package texas

import (
	"log"

	game "github.com/lllllan02/texas-holdem/demos/05-game-state-machine"
	"github.com/lllllan02/texas-holdem/pkg/room"
)

// 游戏特定的 Action 定义
const (
	ActionSit  = "texas.sit"
	ActionBet  = "texas.bet"
	ActionFold = "texas.fold"
)

// TexasEngine 德州扑克游戏引擎，实现了 room.GameEngine 接口
type TexasEngine struct {
	RoomCtx  *room.Room
	Table    *game.Table
	updateCh chan struct{}
}

// NewTexasEngine 创建一个新的德州扑克游戏引擎
func NewTexasEngine() *TexasEngine {
	return &TexasEngine{
		Table:    &game.Table{},
		updateCh: make(chan struct{}, 10),
	}
}

// UpdateChannel 返回状态更新信号通道
func (e *TexasEngine) UpdateChannel() <-chan struct{} {
	return e.updateCh
}

// OnInit 房间初始化时调用
func (e *TexasEngine) OnInit(roomCtx *room.Room, param map[string]any) {
	e.RoomCtx = roomCtx
	log.Printf("[TexasEngine] 房间 %s 初始化德州扑克引擎\n", roomCtx.GetID())
}

// OnDestroy 房间销毁时调用
func (e *TexasEngine) OnDestroy() {
	log.Printf("[TexasEngine] 房间 %s 销毁，清理引擎资源\n", e.RoomCtx.GetID())
	// TODO: 这里可以清理游戏内部的定时器（比如玩家操作倒计时），防止 Goroutine 泄漏
}

// OnPlayerJoin 玩家加入房间时调用
func (e *TexasEngine) OnPlayerJoin(playerID string) {
	log.Printf("[TexasEngine] 玩家 %s 加入房间 %s\n", playerID, e.RoomCtx.GetID())
	// 这里可以处理旁观者逻辑，或者自动坐下逻辑
}

// OnPlayerLeave 玩家离开房间时调用
func (e *TexasEngine) OnPlayerLeave(playerID string) {
	log.Printf("[TexasEngine] 玩家 %s 离开房间 %s\n", playerID, e.RoomCtx.GetID())
	// 这里可以处理玩家离开座位、自动弃牌等逻辑
}

// HandleMessage 处理游戏特定消息
func (e *TexasEngine) HandleMessage(playerID string, action string, content string) {
	switch action {
	case ActionSit:
		e.handleSit(playerID, action, content)
	case ActionBet:
		e.handleBet(playerID, action, content)
	case ActionFold:
		e.handleFold(playerID, action, content)
	default:
		// 未知或未实现的游戏动作，原样广播
		e.broadcastDefault(playerID, action, content)
	}
}

func (e *TexasEngine) handleSit(playerID string, action string, content string) {
	// TODO: 结合 game.Table 的逻辑，让玩家坐下
	log.Printf("[TexasEngine] 玩家 %s 请求坐下\n", playerID)
	e.broadcastDefault(playerID, action, content)
}

func (e *TexasEngine) handleBet(playerID string, action string, content string) {
	// TODO: 结合 game.Table 的逻辑，处理下注
	log.Printf("[TexasEngine] 玩家 %s 请求下注: %s\n", playerID, content)
	e.broadcastDefault(playerID, action, content)
}

func (e *TexasEngine) handleFold(playerID string, action string, content string) {
	// TODO: 结合 game.Table 的逻辑，处理弃牌
	log.Printf("[TexasEngine] 玩家 %s 请求弃牌\n", playerID)
	e.broadcastDefault(playerID, action, content)
}

// broadcastDefault 默认的广播行为（临时占位，实际应该广播游戏状态快照）
func (e *TexasEngine) broadcastDefault(playerID string, action string, content string) {
	// 游戏引擎现在可以直接调用 e.RoomCtx.Broadcast 或者 e.RoomCtx.SendTo
	// 而不需要自己去组装 JSON 和操作 Hub.Broadcast 通道
	e.RoomCtx.Broadcast(playerID, action, content)
}

// GetState 获取当前游戏状态快照
func (e *TexasEngine) GetState(playerID string) interface{} {
	// 这里返回一个通用的状态对象，实际项目中应该返回 Table 的完整快照
	return map[string]interface{}{
		"game":    "texas",
		"stage":   "WAITING",
		"message": "这是德州扑克的恢复状态",
	}
}
