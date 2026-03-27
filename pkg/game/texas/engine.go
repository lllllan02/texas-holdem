package texas

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"

	"github.com/lllllan02/texas-holdem/pkg/room"
)

// 游戏特定的 Action 定义
const (
	ActionSit   = "texas.sit"
	ActionStand = "texas.stand"
	ActionStart = "texas.start"
	ActionPause = "texas.pause"
)

// Engine 德州扑克游戏引擎，实现了 room.GameEngine 接口
type Engine struct {
	mu       sync.Mutex
	room     *room.Room
	Table    *Table
	updateCh chan struct{}

	// 引擎级控制状态
	isRunning bool // 房主是否开启了自动流转（初始默认为 false，需房主手动开始）
}

// NewEngine 创建一个新的德州扑克游戏引擎
func NewEngine() *Engine {
	return &Engine{
		updateCh:  make(chan struct{}, 10),
		isRunning: false, // 初始状态为未运行，等待房主点击开始
	}
}

// UpdateChannel 返回状态更新信号通道
func (e *Engine) UpdateChannel() <-chan struct{} {
	return e.updateCh
}

// triggerSync 触发状态同步广播
func (e *Engine) triggerSync() {
	select {
	case e.updateCh <- struct{}{}:
	default:
	}
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
	e.mu.Lock()
	defer e.mu.Unlock()

	changed := false
	switch action {
	case ActionSit:
		changed = e.handleSit(playerID, content)
	case ActionStand:
		changed = e.handleStand(playerID)
	case ActionStart:
		changed = e.handleToggleRunning(playerID, true)
	case ActionPause:
		changed = e.handleToggleRunning(playerID, false)
	default:
		// 未知或未实现的游戏动作，原样广播
		e.broadcastDefault(playerID, action, content)
	}

	if changed {
		e.triggerSync()
	}
}

func (e *Engine) handleToggleRunning(playerID string, targetState bool) bool {
	// 只有房主可以控制游戏状态
	if playerID != e.room.GetHostID() {
		actionName := "开始"
		if !targetState {
			actionName = "暂停"
		}
		e.room.SendTo(playerID, "error", fmt.Sprintf(`{"message": "只有房主可以%s游戏"}`, actionName))
		return false
	}

	if e.isRunning == targetState {
		return false // 状态未改变
	}

	if targetState {
		// 检查落座人数是否足够（至少需要2人）
		seatedCount := 0
		for _, p := range e.Table.Seats {
			if p != nil {
				seatedCount++
			}
		}

		if seatedCount < 2 {
			e.room.SendTo(playerID, "error", `{"message": "至少需要2名玩家落座才能开始游戏"}`)
			return false
		}

		log.Printf("[TexasEngine] 房间 %s 房主 %s 开始/恢复了游戏\n", e.room.GetID(), playerID)
		// TODO: 如果当前没有进行中的牌局，则调用 e.Table.StartNewHand()
	} else {
		log.Printf("[TexasEngine] 房间 %s 房主 %s 暂停了游戏\n", e.room.GetID(), playerID)
		// 注意：这里只是设置了标记，不会打断当前正在进行的牌局，
		// 而是在当前牌局结算（Showdown）后，阻止自动进入下一局。
	}

	e.isRunning = targetState
	return true
}

func (e *Engine) handleSit(playerID string, content string) bool {
	var req struct {
		Seat int `json:"seat"`
	}
	if err := json.Unmarshal([]byte(content), &req); err != nil {
		log.Printf("[TexasEngine] 玩家 %s 坐下参数解析失败: %v\n", playerID, err)
		return false
	}

	if err := e.Table.SitDown(playerID, req.Seat); err != nil {
		log.Printf("[TexasEngine] 玩家 %s 坐下失败: %v\n", playerID, err)
		// 可以发送一个错误消息给该玩家
		e.room.SendTo(playerID, "error", `{"message": "`+err.Error()+`"}`)
		return false
	}

	log.Printf("[TexasEngine] 玩家 %s 坐下座位 %d\n", playerID, req.Seat)

	// 标记状态已改变，需要广播
	return true
}

func (e *Engine) handleStand(playerID string) bool {
	if err := e.Table.StandUp(playerID); err != nil {
		log.Printf("[TexasEngine] 玩家 %s 站起失败: %v\n", playerID, err)
		return false
	}

	log.Printf("[TexasEngine] 玩家 %s 站起\n", playerID)

	// 标记状态已改变，需要广播
	return true
}

// broadcastDefault 默认的广播行为（临时占位，实际应该广播游戏状态快照）
func (e *Engine) broadcastDefault(playerID string, action string, content string) {
	// 游戏引擎现在可以直接调用 e.RoomCtx.Broadcast 或者 e.RoomCtx.SendTo
	// 而不需要自己去组装 JSON 和操作 Hub.Broadcast 通道
	e.room.Broadcast(playerID, action, content)
}

// GetState 获取当前游戏状态快照
func (e *Engine) GetState(playerID string) any {
	if e.Table == nil {
		return nil
	}

	// 获取 Table 的安全快照
	snap := e.Table.GetSnapshot(playerID)

	// 包装一层，把引擎级的状态也放进去
	return map[string]any{
		"isRunning": e.isRunning,
		"table":     snap,
	}
}
