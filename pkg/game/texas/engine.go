package texas

import (
	"log"

	"github.com/lllllan02/texas-holdem/pkg/room"
)

// Engine 德州扑克游戏引擎，实现了 room.GameEngine 接口
type Engine struct {
	room  *room.Room
	Table *Table
}

// NewEngine 创建一个新的德州扑克游戏引擎
func NewEngine() *Engine {
	return &Engine{}
}

// OnInit 房间初始化时调用
func (e *Engine) OnInit(room *room.Room, param map[string]any) {
	e.room = room
	e.Table = NewTable(param)

	log.Printf("[TexasEngine] 房间 %s 初始化德州扑克引擎\n", room.GetID())
}

// OnDestroy 房间销毁时调用
func (e *Engine) OnDestroy() {
	log.Printf("[TexasEngine] 房间 %s 销毁，清理引擎资源\n", e.room.GetID())
	// TODO: 这里可以清理游戏内部的定时器（比如玩家操作倒计时），防止 Goroutine 泄漏
}

// OnPlayerJoin 玩家加入房间时调用
func (e *Engine) OnPlayerJoin(playerID string) {
	log.Printf("[TexasEngine] 玩家 %s 加入房间 %s\n", playerID, e.room.GetID())
	// 这里可以处理旁观者逻辑，或者自动坐下逻辑
}

// OnPlayerLeave 玩家离开房间时调用
func (e *Engine) OnPlayerLeave(playerID string) {
	log.Printf("[TexasEngine] 玩家 %s 离开房间 %s\n", playerID, e.room.GetID())
	// 这里可以处理玩家离开座位、自动弃牌等逻辑
}

// HandleMessage 处理游戏特定消息
func (e *Engine) HandleMessage(playerID string, action string, content string) {
	switch action {
	default:
		// 未知或未实现的游戏动作，原样广播
		e.broadcastDefault(playerID, action, content)
	}
}

// broadcastDefault 默认的广播行为（临时占位，实际应该广播游戏状态快照）
func (e *Engine) broadcastDefault(playerID string, action string, content string) {
	// 游戏引擎现在可以直接调用 e.RoomCtx.Broadcast 或者 e.RoomCtx.SendTo
	// 而不需要自己去组装 JSON 和操作 Hub.Broadcast 通道
	e.room.Broadcast(playerID, action, content)
}

// GetState 获取当前游戏状态快照
func (e *Engine) GetState(playerID string) any {
	// 这里返回一个通用的状态对象，实际项目中应该返回 Table 的完整快照
	return map[string]any{
		"game":    "texas",
		"stage":   "WAITING",
		"message": "这是德州扑克的恢复状态",
	}
}
