package handler

import (
	"encoding/json"
	"log"
	"time"

	"github.com/lllllan02/texas-holdem/pkg/core"
	"github.com/lllllan02/texas-holdem/pkg/game/texas"
	"github.com/lllllan02/texas-holdem/pkg/user"
	"github.com/lllllan02/texas-holdem/pkg/wscore"
)

const maxChatHistory = 20 // 房间内最多保留的聊天记录条数

// Room 通用房间：最外层的容器，负责网络连接、聊天、用户进出
// 房间本身不关心具体玩什么游戏，它只负责挂载一个具体的 GameEngine
type Room struct {
	// 房间基础信息
	ID         string `json:"id"`          // 房间的唯一内部标识（如 UUID）
	RoomNumber string `json:"room_number"` // 供玩家输入的短房间号（如 "888888"），作为进入凭据
	OwnerID    string `json:"owner_id"`    // 房主的 UserID，用于权限校验

	// 房间状态与数据
	IsPaused bool                  `json:"is_paused"` // 房间是否处于全局暂停状态
	Users    map[string]*user.User `json:"users"`     // 当前在房间内的所有用户（包括玩家和旁观者）

	// 核心组件
	GameEngine core.GameEngine           `json:"-"` // 挂载的具体游戏引擎实例
	hub        *wscore.Hub               // WebSocket 事件循环核心
	clients    map[string]*wscore.Client // 维护 UserID 到 Client 的映射
	stopped    bool

	chatHistory []HistoryEntry // 暂存的聊天记录（最多保留 maxChatHistory 条），用于断线重连或新玩家加入时展示
	gameHistory []HistoryEntry // 暂存的本局游戏动作日志，新的一局开始时清空，用于断线重连时回放本局发生的所有事情

	manager *RoomManager
}

// NewRoom 创建一个新的房间实例
func NewRoom(id, roomNumber, ownerID string, engine core.GameEngine, options []byte, manager *RoomManager) (*Room, error) {
	r := &Room{
		ID:          id,
		RoomNumber:  roomNumber,
		OwnerID:     ownerID,
		Users:       make(map[string]*user.User),
		GameEngine:  engine,
		hub:         wscore.NewHub(),
		clients:     make(map[string]*wscore.Client),
		chatHistory: make([]HistoryEntry, 0),
		gameHistory: make([]HistoryEntry, 0),
		manager:     manager,
	}

	// 初始化游戏引擎
	if err := r.GameEngine.OnInit(r, options); err != nil {
		return nil, err
	}

	// 引擎初始化成功后，启动房间事件循环
	go r.hub.Run()

	return r, nil
}

// 确保 Room 实现了 core.Messenger 和 wscore.MessageHandler 接口
var _ core.Messenger = (*Room)(nil)

// Broadcast 广播给游戏内的所有玩家
func (r *Room) Broadcast(msgType string, reason string, payload any) {
	msg := core.Message{
		Type:    msgType,
		Reason:  reason,
		Payload: payload,
	}

	entry := HistoryEntry{
		Message: msg,
		Time:    time.Now().UnixMilli(),
	}

	// 暂存聊天记录
	if msgType == MsgTypeChat {
		r.chatHistory = append(r.chatHistory, entry)
		if len(r.chatHistory) > maxChatHistory {
			r.chatHistory = r.chatHistory[1:]
		}
	}

	// 暂存游戏日志 (针对德州扑克)
	if msgType == texas.MsgTypeStateUpdate {
		if reason == texas.ReasonDealHoleCards {
			r.gameHistory = nil // 新的一局开始，清空之前的日志
		}

		// 只要是状态更新，且不是单纯为了同步刚进房间/离开房间的快照，都记录下来
		// 这样前端在断线重连时，能通过这些日志回放本局发生的所有事情
		if reason != texas.ReasonPlayerJoined && reason != texas.ReasonPlayerLeft {
			r.gameHistory = append(r.gameHistory, entry)
		}
	}

	data, err := json.Marshal(msg)
	if err != nil {
		log.Printf("Room [%s] broadcast marshal error: %v", r.ID, err)
		return
	}
	
	// 直接遍历 clients 发送，保证与 SendTo 的顺序一致性
	for _, client := range r.clients {
		client.SendMessage(data)
	}
}

// SendTo 私发给游戏内的特定玩家
func (r *Room) SendTo(userID string, msgType string, reason string, payload any) {
	msg := core.Message{
		Type:    msgType,
		Reason:  reason,
		Payload: payload,
	}
	data, err := json.Marshal(msg)
	if err != nil {
		log.Printf("Room [%s] sendTo marshal error: %v", r.ID, err)
		return
	}

	if client, ok := r.clients[userID]; ok {
		client.SendMessage(data)
	}
}

// Execute 提交异步任务到 Hub 的主循环中执行，保证状态修改的并发安全
func (r *Room) Execute(task func()) {
	r.hub.Execute(task)
}

var _ wscore.MessageHandler = (*Room)(nil)

// Stop 停止房间并销毁游戏引擎
func (r *Room) Stop() {
	if r.stopped {
		return
	}
	r.stopped = true

	// 广播房间解散消息
	r.Broadcast(MsgTypeRoomDestroy, "owner_deleted", nil)

	r.GameEngine.OnDestroy()
	r.hub.Stop()
}

// OnConnect 客户端连接建立
func (r *Room) OnConnect(client *wscore.Client) {
	userID := client.GetID()
	r.clients[userID] = client
	if r.manager != nil {
		r.manager.OnRoomActive(r.RoomNumber)
		isOwner := userID == r.OwnerID
		r.manager.RecordUserRoom(userID, r.RoomNumber, isOwner)
	}

	u := user.GetUserByID(userID)
	r.Users[userID] = u

	r.Broadcast(MsgTypePlayerJoin, "", PlayerEventPayload{
		UserID:   userID,
		UserName: u.Nickname,
	})

	r.SendTo(userID, MsgTypeWelcome, "", WelcomePayload{
		RoomID:     r.ID,
		RoomNumber: r.RoomNumber,
		OwnerID:    r.OwnerID,
		User:       u,
		GameType:   r.GameEngine.GameType(),
	})

	// 发送历史记录
	historyMsg := core.Message{
		Type: MsgTypeHistory,
		Payload: HistoryPayload{
			ChatHistory: r.chatHistory,
			GameHistory: r.gameHistory,
		},
	}
	if historyData, err := json.Marshal(historyMsg); err == nil {
		client.SendMessage(historyData)
	}

	r.GameEngine.OnPlayerJoin(u)

	log.Printf("Room [%s] Client [%s] connected and joined", r.ID, userID)
}

// OnDisconnect 客户端断开连接
func (r *Room) OnDisconnect(client *wscore.Client) {
	userID := client.GetID()
	delete(r.clients, userID)

	// 如果用户在房间内，通知游戏引擎玩家离开
	if u, ok := r.Users[userID]; ok {
		r.Broadcast(MsgTypePlayerLeave, "", PlayerEventPayload{
			UserID:   userID,
			UserName: u.Nickname,
		})
		delete(r.Users, userID)
		r.GameEngine.OnPlayerLeave(userID)
	}
	log.Printf("Room [%s] Client [%s] disconnected", r.ID, userID)

	// 如果房间已无任何连接，启动自动回收计时器
	if len(r.clients) == 0 {
		if r.manager != nil {
			r.manager.OnRoomEmpty(r.RoomNumber)
		}
	}
}

// ClientMessage 用于接收客户端发来的消息
type ClientMessage struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload,omitempty"`
}

// OnMessage 接收并处理客户端消息
func (r *Room) OnMessage(client *wscore.Client, message []byte) {
	userID := client.GetID()

	var msg ClientMessage
	if err := json.Unmarshal(message, &msg); err != nil {
		log.Printf("Room [%s] unmarshal message error: %v", r.ID, err)
		return
	}

	// 拦截房间级消息
	switch msg.Type {
	case MsgTypePauseGame:
		r.handlePauseGame(userID)
		return
	case MsgTypeResumeGame:
		r.handleResumeGame(userID)
		return
	case MsgTypeChat:
		r.handleChat(userID, msg.Payload)
		return
	}

	// 其他消息转发给游戏引擎
	if err := r.GameEngine.HandleMessage(userID, msg.Type, msg.Payload); err != nil {
		log.Printf("Room [%s] engine handle message error: %v", r.ID, err)
		r.SendTo(userID, core.MsgTypeError, "", core.ErrorPayload{Error: err.Error()})
	}
}

func (r *Room) handlePauseGame(userID string) {
	if userID != r.OwnerID {
		r.SendTo(userID, core.MsgTypeError, "", core.ErrorPayload{Error: "only owner can pause game"})
		return
	}

	if err := r.GameEngine.Pause(); err != nil {
		r.SendTo(userID, core.MsgTypeError, "", core.ErrorPayload{Error: err.Error()})
		return
	}
	r.IsPaused = true
	r.Broadcast(MsgTypeGamePaused, "user_pause", GamePausedPayload{
		IsPaused: true,
		UserID:   userID,
	})
}

func (r *Room) handleResumeGame(userID string) {
	if userID != r.OwnerID {
		r.SendTo(userID, core.MsgTypeError, "", core.ErrorPayload{Error: "only owner can resume game"})
		return
	}

	if err := r.GameEngine.Resume(); err != nil {
		r.SendTo(userID, core.MsgTypeError, "", core.ErrorPayload{Error: err.Error()})
		return
	}
	r.IsPaused = false
	r.Broadcast(MsgTypeGamePaused, "user_resume", GamePausedPayload{
		IsPaused: false,
		UserID:   userID,
	})
}

func (r *Room) handleChat(userID string, payload []byte) {
	// 1. 解析前端传来的聊天内容
	var chatReq struct {
		Message string `json:"message"`
	}
	if err := json.Unmarshal(payload, &chatReq); err != nil {
		r.SendTo(userID, core.MsgTypeError, "", core.ErrorPayload{Error: "invalid chat payload"})
		return
	}

	if chatReq.Message == "" {
		return
	}

	// 2. 获取发送者的用户信息
	u, ok := r.Users[userID]
	if !ok {
		return
	}

	// 3. 组装并广播聊天消息给房间内的所有人
	r.Broadcast(MsgTypeChat, "", ChatPayload{
		UserID:   userID,
		UserName: u.Nickname,
		Avatar:   u.Avatar,
		Message:  chatReq.Message,
	})
}
