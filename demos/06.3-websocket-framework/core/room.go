package core

import (
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/lllllan02/texas-holdem/pkg/wscore"
)

type PlayerContext struct {
	Username string
}

type Client = wscore.Client

// Room 房间，包含一个游戏桌和一个广播中心
type Room struct {
	ID      string
	HostID  string       // 房主ID
	Engine  GameEngine   // 抽象的游戏引擎
	Hub     *wscore.Hub  // 该房间专属的 WebSocket 广播中心
	Manager *RoomManager // 引用全局管理器，用于自动回收

	// 房间内的玩家管理
	mu        sync.Mutex
	Users     map[string]*Client
	joinOrder []string

	// 定时器
	roomRecycleTimer  *time.Timer
	hostTransferTimer *time.Timer
}

// GetID 获取房间 ID
func (r *Room) GetID() string {
	return r.ID
}

// GetHostID 获取房主 ID
func (r *Room) GetHostID() string {
	return r.HostID
}

// Broadcast 向房间内所有客户端广播消息，可以指定发送者（如果是系统消息，sender 传 nil）
func (r *Room) Broadcast(sender *Client, action string, content string) {
	outBytes := BuildServerMessage(sender, action, content)
	r.Hub.BroadcastMessage(outBytes)
}

// SendTo 向指定的客户端发送单播消息
func (r *Room) SendTo(client *Client, action string, content string) {
	outBytes := BuildServerMessage(nil, action, content)
	client.SendMessage(outBytes)
}

// broadcastRoomState 广播当前房间内的玩家列表
func (r *Room) broadcastRoomState() {
	state := map[string]interface{}{
		"roomId":  r.ID,
		"hostId":  r.HostID,
		"players": make([]string, 0, len(r.Users)),
	}

	players := state["players"].([]string)
	for _, client := range r.Users {
		// players = append(players, client.GetContext().(*PlayerContext).Username+" ("+client.GetID()+")")
		players = append(players, client.GetID()+" ("+client.GetID()+")")
	}
	state["players"] = players

	stateBytes, _ := json.Marshal(state)
	r.Broadcast(nil, ActionRoomState, string(stateBytes))
}

// NewRoom 创建一个新房间
func NewRoom(id string, hostID string, engine GameEngine, manager *RoomManager) *Room {
	room := &Room{
		ID:        id,
		HostID:    hostID,
		Engine:    engine,
		Manager:   manager,
		Users:     make(map[string]*Client),
		joinOrder: make([]string, 0),
	}

	room.Hub = wscore.NewHub()
	room.roomRecycleTimer = time.AfterFunc(time.Hour*24, room.handleRoomRecycleTimeout)
	room.hostTransferTimer = time.AfterFunc(time.Hour*24, room.handleHostTransferTimeout)
	room.roomRecycleTimer.Stop()
	room.hostTransferTimer.Stop()

	if room.Engine != nil {
		room.Engine.OnInit(room)
	}

	go room.Hub.Run()

	return room
}

func (r *Room) handleRoomRecycleTimeout() {
	log.Printf("房间 [%s] 长期空闲，自动回收\n", r.ID)
	go r.Manager.RemoveRoom(r.ID)
}

func (r *Room) handleHostTransferTimeout() {
	r.mu.Lock()
	defer r.mu.Unlock()

	// 房主转移逻辑
	if len(r.Users) > 0 {
		var newHost string
		for _, id := range r.joinOrder {
			if _, online := r.Users[id]; online {
				newHost = id
				break
			}
		}

		if newHost != "" && newHost != r.HostID {
			r.HostID = newHost
			log.Printf("房间 [%s] 房主转移给 [%s]\n", r.ID, newHost)
			r.broadcastRoomState()

			// 广播通知
			// outBytes := BuildServerMessage(nil, ActionHostChanged, "原房主掉线，房主已自动转移给 "+r.Users[newHost].GetContext().(*PlayerContext).Username)
			outBytes := BuildServerMessage(nil, ActionHostChanged, "原房主掉线，房主已自动转移给 "+r.Users[newHost].GetID())
			r.Hub.BroadcastMessage(outBytes)
		}
	}
}

// AddClient 将客户端加入房间 (由 ServeWS 调用)
func (r *Room) AddClient(client *Client) error {
	log.Printf("客户端 [%s] 准备加入房间 [%s]\n", client.GetID(), r.ID)
	// 注册由 ServeWS 内部处理，这里其实不需要手动调用 Hub.Register
	// 但为了触发 OnPlayerJoin，我们在 OnConnect 中处理
	return nil
}

// RemoveClient 将客户端移出房间
func (r *Room) RemoveClient(client *Client) {
	log.Printf("客户端 [%s] 离开房间 [%s]\n", client.GetID(), r.ID)
	client.Close()
}

// wscore.MessageHandler 接口实现

func (r *Room) OnConnect(client *Client) {
	r.mu.Lock()

	// 顶号逻辑
	if oldClient, ok := r.Users[client.GetID()]; ok {
		// 发送踢出消息
		outBytes := BuildServerMessage(nil, ActionKick, "您的账号在其他地方登录，您已被强制下线。")
		oldClient.SendMessage(outBytes)
		oldClient.Close()
	}

	r.Users[client.GetID()] = client

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
	if client.GetID() == r.HostID {
		r.hostTransferTimer.Stop()
	}

	r.mu.Unlock()

	log.Printf("玩家 [%s] 加入了房间 [%s]\n", client.GetID(), r.ID)

	// 广播加入消息和房间状态
	r.Broadcast(client, ActionJoin, "加入了房间")
	r.broadcastRoomState()

	if r.Engine != nil {
		r.Engine.OnPlayerJoin(client)

		// 触发状态同步
		state := r.Engine.GetState(client)
		stateBytes, _ := json.Marshal(state)
		r.SendTo(client, ActionSyncState, string(stateBytes))
	}
}

func (r *Room) OnMessage(client *Client, message []byte) {
	var msg ClientMessage
	if err := json.Unmarshal(message, &msg); err != nil {
		log.Printf("消息解析失败: %v", err)
		return
	}

	// 强制覆盖为当前客户端的信息，防止伪造
	msg.UserID = client.GetID()
	// msg.Username = client.GetContext().(*PlayerContext).Username

	switch msg.Action {
	case ActionChat:
		// 房间内聊天，直接广播
		r.Broadcast(client, ActionChat, msg.Content)
	case ActionLeave:
		// 玩家主动离开
		client.Close()
	default:
		// 其他消息交给游戏引擎处理
		if r.Engine != nil {
			r.Engine.HandleMessage(client, msg)
		}
	}
}

func (r *Room) OnDisconnect(client *Client) {
	r.mu.Lock()

	// 只有当断开的连接是当前记录的连接时，才执行清理
	// 防止顶号时，旧连接断开导致新连接被误删
	if r.Users[client.GetID()] == client {
		delete(r.Users, client.GetID())

		if len(r.Users) == 0 {
			r.roomRecycleTimer.Reset(30 * time.Second)
		} else if client.GetID() == r.HostID {
			r.hostTransferTimer.Reset(15 * time.Second)
		}
	}

	r.mu.Unlock()

	log.Printf("玩家 [%s] 断开了与房间 [%s] 的连接\n", client.GetID(), r.ID)

	// 广播离开消息和房间状态
	r.Broadcast(client, ActionLeave, "离开了房间")
	r.broadcastRoomState()

	if r.Engine != nil {
		r.Engine.OnPlayerLeave(client)
	}
}
