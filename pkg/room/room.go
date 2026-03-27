package room

import (
	"context"
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/lllllan02/texas-holdem/pkg/wscore"
)

// Room 房间实体，负责管理房间内的物理连接和生命周期
type Room struct {
	// 房间的唯一标识
	id string

	// 房主的唯一标识，拥有解散房间、开始游戏等特权
	hostID string

	// 注入的游戏引擎，负责处理具体的游戏逻辑（如德州扑克、斗地主等）
	engine GameEngine

	// WebSocket 连接管理器，负责维护房间内所有玩家的长连接并处理消息广播
	hub *wscore.Hub

	// 房间管理器引用，用于在房间空闲时通知管理器销毁自己
	manager *RoomManager

	// 保护 users 和 joinOrder 等状态的互斥锁
	mu sync.RWMutex

	// 当前在房间内的在线玩家集合，Key 为玩家 ID
	users map[string]*wscore.Client

	// 记录玩家加入房间的顺序，用于在房主掉线时按顺序顺延移交房主权限
	joinOrder []string

	// 房间空闲回收定时器：当房间内没有玩家时启动，超时后自动销毁房间
	roomRecycleTimer *time.Timer

	// 房主转移定时器：当房主掉线时启动，超时后将房主权限移交给下一个玩家
	hostTransferTimer *time.Timer

	// 房间生命周期控制
	ctx    context.Context
	cancel context.CancelFunc

	// 确保 destroy 逻辑只执行一次
	destroyOnce sync.Once
}

// destroy 销毁房间，清理相关资源。
// 这是一个内部方法，只能由 RoomManager 的 RemoveRoom 调用，
// 外部业务层如果想销毁房间，应该调用 manager.RemoveRoom(roomID)。
func (r *Room) destroy() {
	r.destroyOnce.Do(func() {
		if r.cancel != nil {
			r.cancel()
		}
		r.roomRecycleTimer.Stop()
		r.hostTransferTimer.Stop()
		r.hub.Stop()

		if r.engine != nil {
			r.engine.OnDestroy()
		}
	})
}

// GetID 获取房间 ID
func (r *Room) GetID() string {
	return r.id
}

// GetHostID 获取房主 ID
func (r *Room) GetHostID() string {
	return r.hostID
}

// GetHub 获取房间的 WebSocket Hub（供 API 层接入连接使用）
func (r *Room) GetHub() *wscore.Hub {
	return r.hub
}

// Broadcast 向房间内所有客户端广播消息
func (r *Room) Broadcast(senderID string, action string, content any) {
	outBytes := BuildServerMessage(senderID, action, content)
	r.hub.BroadcastMessage(outBytes)
}

// SendTo 向指定的客户端发送单播消息
func (r *Room) SendTo(playerID string, action string, content any) {
	r.mu.RLock()
	client, ok := r.users[playerID]
	r.mu.RUnlock()

	if ok {
		outBytes := BuildServerMessage("", action, content)
		client.SendMessage(outBytes)
	}
}

// KickPlayer 踢出指定玩家（断开其连接）
func (r *Room) KickPlayer(playerID string, reason string) {
	r.mu.RLock()
	client, ok := r.users[playerID]
	r.mu.RUnlock()

	if ok {
		if reason != "" {
			outBytes := BuildServerMessage("", ActionKick, reason)
			client.SendMessage(outBytes)
		}
		client.Close()
	}
}

// NewRoom 创建一个新房间
func NewRoom(id string, hostID string, engine GameEngine, param map[string]any, manager *RoomManager) *Room {
	ctx, cancel := context.WithCancel(context.Background())
	rm := &Room{
		id:        id,
		hostID:    hostID,
		engine:    engine,
		manager:   manager,
		users:     make(map[string]*wscore.Client),
		joinOrder: make([]string, 0),
		hub:       wscore.NewHub(),
		ctx:       ctx,
		cancel:    cancel,
	}

	// 初始状态下房间为空，启动回收定时器（5 分钟内无人加入则销毁）
	rm.roomRecycleTimer = time.AfterFunc(5*time.Minute, rm.handleRoomRecycleTimeout)
	rm.hostTransferTimer = time.AfterFunc(time.Hour*24, rm.handleHostTransferTimeout)
	rm.hostTransferTimer.Stop()

	if rm.engine != nil {
		rm.engine.OnInit(rm, param)
		go rm.watchEngineUpdates()
	}

	go rm.hub.Run()
	return rm
}

// OnConnect 实现 wscore.MessageHandler 接口，处理客户端连接
func (r *Room) OnConnect(client *wscore.Client) {
	defer func() {
		if err := recover(); err != nil {
			log.Printf("Room [%s] OnConnect panic recovered: %v", r.id, err)
		}
	}()

	r.mu.Lock()

	// 顶号逻辑
	if oldClient, ok := r.users[client.GetID()]; ok {
		outBytes := BuildServerMessage("", ActionKick, "您的账号在其他地方登录，您已被强制下线。")
		oldClient.SendMessage(outBytes)
		oldClient.Close()
	}

	r.users[client.GetID()] = client

	// 记录加入顺序
	found := false
	for _, id := range r.joinOrder {
		if id == client.GetID() {
			found = true
			break
		}
	}
	if !found {
		r.joinOrder = append(r.joinOrder, client.GetID())
	}

	// 停止定时器
	r.roomRecycleTimer.Stop()
	if client.GetID() == r.hostID {
		r.hostTransferTimer.Stop()
	}

	r.mu.Unlock()

	log.Printf("玩家 [%s] 加入了房间 [%s]\n", client.GetID(), r.id)

	r.Broadcast(client.GetID(), ActionJoin, "加入了房间")

	if r.engine != nil {
		r.engine.OnPlayerJoin(client.GetID())

		state := r.getFullState(client.GetID())
		r.SendTo(client.GetID(), ActionSyncState, state)
	}
}

// OnMessage 实现 wscore.MessageHandler 接口，处理客户端消息
func (r *Room) OnMessage(client *wscore.Client, message []byte) {
	var msg ClientMessage
	if err := json.Unmarshal(message, &msg); err != nil {
		log.Printf("消息解析失败: %v", err)
		return
	}

	switch msg.Action {
	case ActionChat:
		r.Broadcast(client.GetID(), ActionChat, msg.Content)
	case ActionLeave:
		// 客户端主动请求离开房间
		r.mu.RLock()
		isHost := client.GetID() == r.hostID
		r.mu.RUnlock()

		if isHost {
			// 如果是房主主动离开，立刻触发房主转移，不等待 15 秒重连窗口
			r.handleHostTransferTimeout()
		}

		// 直接关闭其连接。
		// 底层连接关闭后，会自动触发 OnDisconnect 走统一的清理和广播流程。
		client.Close()
	default:
		if r.engine != nil {
			r.engine.HandleMessage(client.GetID(), msg.Action, msg.Content)
		}
	}
}

// OnDisconnect 实现 wscore.MessageHandler 接口，处理客户端断开
func (r *Room) OnDisconnect(client *wscore.Client) {
	defer func() {
		if err := recover(); err != nil {
			log.Printf("Room [%s] OnDisconnect panic recovered: %v", r.id, err)
		}
	}()

	r.mu.Lock()

	// 只有当断开的连接是当前记录的连接时，才执行清理（防止顶号时旧连接断开误删新连接）
	if r.users[client.GetID()] == client {
		delete(r.users, client.GetID())

		// 如果房间空了，启动回收定时器
		if len(r.users) == 0 {
			r.roomRecycleTimer.Reset(30 * time.Second)
		} else if client.GetID() == r.hostID {
			// 如果房主掉线且房间还有人，检查是否是因为主动离开触发的
			// （主动离开时 hostID 已经被修改，所以不会进这个分支，只有被动掉线才会进）
			r.hostTransferTimer.Reset(15 * time.Second)
		}
	} else {
		// 如果不是当前有效连接（比如被顶号踢掉的旧连接），直接忽略，不广播离开消息
		r.mu.Unlock()
		return
	}

	r.mu.Unlock()

	log.Printf("玩家 [%s] 断开了与房间 [%s] 的连接\n", client.GetID(), r.id)

	// 广播玩家离开的消息
	r.Broadcast(client.GetID(), ActionLeave, "离开了房间")

	// 通知游戏引擎玩家已离开
	if r.engine != nil {
		r.engine.OnPlayerLeave(client.GetID())
	}
}

func (r *Room) handleRoomRecycleTimeout() {
	log.Printf("房间 [%s] 长期空闲，自动回收\n", r.id)
	go r.manager.RemoveRoom(r.id)
}

func (r *Room) handleHostTransferTimeout() {
	r.mu.Lock()
	defer r.mu.Unlock()

	if len(r.users) > 0 {
		var newHost string
		for _, id := range r.joinOrder {
			// 找到第一个在线且不是当前房主的玩家
			if _, online := r.users[id]; online && id != r.hostID {
				newHost = id
				break
			}
		}

		if newHost != "" && newHost != r.hostID {
			r.hostID = newHost
			log.Printf("房间 [%s] 房主转移给 [%s]\n", r.id, newHost)

			r.Broadcast("", ActionHostChanged, "原房主离开或掉线，房主已自动转移给 "+newHost)
		}
	}
}

func (r *Room) watchEngineUpdates() {
	for {
		select {
		case <-r.engine.UpdateChannel():
			r.broadcastState()
		case <-r.ctx.Done():
			return
		}
	}
}

func (r *Room) broadcastState() {
	defer func() {
		if err := recover(); err != nil {
			log.Printf("Room [%s] broadcastState panic recovered: %v", r.id, err)
		}
	}()

	players := r.getPlayers()
	for _, playerID := range players {
		// 统一使用 getFullState，保证前后端状态结构一致，包含 roomId, hostId 和 gameState
		state := r.getFullState(playerID)
		r.SendTo(playerID, ActionSyncState, state)
	}
}

// getFullState 获取完整的房间和游戏状态
func (r *Room) getFullState(playerID string) map[string]any {
	state := map[string]any{
		"roomId":  r.id,
		"hostId":  r.hostID,
		"players": r.getPlayers(),
	}

	// 如果有引擎，直接把引擎的状态作为一个独立的字段放进去
	if r.engine != nil {
		// GetState 返回的可能是一个结构体指针，在转 JSON 时会变成对象
		state["gameState"] = r.engine.GetState(playerID)
	}
	return state
}

// GetPlayers 获取当前房间内所有在线玩家的 ID
func (r *Room) getPlayers() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	players := make([]string, 0, len(r.users))
	for id := range r.users {
		players = append(players, id)
	}
	return players
}
