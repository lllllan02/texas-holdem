package core

import (
	"encoding/json"
	"fmt"
	"log"
	"slices"
	"time"
)

// Hub 模拟房间专属的 WebSocket 广播中心
// 它是一个单 Goroutine 运行的事件循环，负责处理所有与连接相关的并发操作，
// 从而避免了在多个 Goroutine 中读写 Clients map 导致的并发冲突。
type Hub struct {
	// 关联的房间引用，用于在广播时获取房间信息或调用引擎方法
	Room *Room

	// 当前房间内所有活跃的 WebSocket 连接。
	// 使用 map 的 bool 值只是为了方便快速判断一个 Client 是否存在。
	Clients map[*Client]bool

	// 用户 ID 到 Client 的映射。
	// 作用：
	// 1. 实现单播 (SendToUser)：通过 UserID 快速找到对应的 WebSocket 连接。
	// 2. 顶号逻辑：如果同一个 UserID 再次加入，可以通过这个 map 找到旧连接并踢掉。
	Users map[string]*Client

	// 记录玩家加入房间的先后顺序（按 UserID 存储）。
	// 作用：当房主掉线时，系统需要按照“先来后到”的原则，将房主权限顺延给下一个最早加入的玩家。
	joinOrder []string

	// 广播通道。任何发往这个通道的字节流，都会被 Hub 转发给 Clients 中的所有连接。
	Broadcast chan []byte

	// 注册通道。当有新玩家连接 WebSocket 时，Client 会被发送到这个通道，Hub 会将其加入 Clients。
	Register chan *Client

	// 注销通道。当玩家断开 WebSocket 时，Client 会被发送到这个通道，Hub 会将其从 Clients 中移除。
	Unregister chan *Client

	// 销毁通道。当房间被解散或回收时，向这个通道发送信号，Hub 会结束运行并清理所有连接。
	DestroyChan chan struct{}

	// 房间空闲回收定时器。
	// 当房间内人数为 0 时启动，如果倒计时结束仍无人加入，则触发房间销毁逻辑。
	roomRecycleTimer *time.Timer

	// 房主掉线转移定时器。
	// 当房主断开连接（但房间内还有其他人）时启动。如果倒计时结束房主仍未重连，则触发房主转移逻辑。
	hostTransferTimer *time.Timer
}

// Run 启动 Hub 的事件循环
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.Register:
			h.handleRegister(client)

		case client := <-h.Unregister:
			h.handleUnregister(client)

		case <-h.roomRecycleTimer.C:
			h.handleRoomRecycleTimeout()

		case <-h.hostTransferTimer.C:
			h.handleHostTransferTimeout()

		case message := <-h.Broadcast:
			h.broadcastBytes(message)

		case <-h.DestroyChan:
			h.handleDestroy()
			return // 结束 Hub 的运行
		}
	}
}

// closeClient 安全地清理 Hub 中的客户端资源并关闭通道
// 注意：此方法必须只在 Hub.Run() 的同一个 Goroutine 中被调用，以保证对 map 的并发安全
func (h *Hub) closeClient(client *Client) {
	if !client.closed {
		client.closed = true

		// 关闭发送通道
		close(client.Send)

		// 主动关闭底层的 WebSocket 连接，触发 ReadPump 退出
		client.Conn.Close()
	}

	// 从 Clients 和 Users 中移除客户端
	delete(h.Clients, client)
	if h.Users[client.ID] == client {
		delete(h.Users, client.ID)
	}
}

func (h *Hub) handleRegister(client *Client) {
	// 顶号逻辑：如果该 UserID 已经有连接了，踢掉旧的
	h.kickOutOldClient(client.ID)

	h.Clients[client] = true
	h.Users[client.ID] = client

	// 记录加入顺序
	if !slices.Contains(h.joinOrder, client.ID) {
		h.joinOrder = append(h.joinOrder, client.ID)
	}

	// 有人加入，停止空房间回收定时器
	h.roomRecycleTimer.Stop()
	// 如果是房主重连，停止房主转移定时器
	if client.ID == h.Room.HostID {
		h.hostTransferTimer.Stop()
	}

	log.Printf("客户端加入 Hub: %s\n", client.ID)
	h.broadcastRoomState() // 广播最新房间状态

	// 触发状态同步 (断线重连/新玩家加入)
	if h.Room.Engine != nil {
		state := h.Room.Engine.GetState(client)
		stateBytes, _ := json.Marshal(state)
		h.sendTo(client, ActionSyncState, string(stateBytes))
	}
}

func (h *Hub) handleUnregister(client *Client) {
	if _, ok := h.Clients[client]; ok {
		h.closeClient(client)

		log.Printf("客户端离开 Hub: %s\n", client.ID)
		h.broadcastRoomState() // 广播最新房间状态

		// 如果房间空了，启动回收定时器 (例如 30 秒后回收)
		if len(h.Clients) == 0 {
			h.roomRecycleTimer.Reset(30 * time.Second)
		} else if client.ID == h.Room.HostID {
			// 房间还有人，但房主掉线了，启动房主转移定时器 (例如 15 秒后转移)
			h.hostTransferTimer.Reset(15 * time.Second)
		}
	}
}

func (h *Hub) handleRoomRecycleTimeout() {
	log.Printf("房间 [%s] 长期空闲，自动回收\n", h.Room.ID)
	// 因为当前已经在 Hub.Run() 的 Goroutine 中，
	// 如果直接调用 Manager.RemoveRoom，而 RemoveRoom 内部又会调用 close(DestroyChan)，
	// 这可能会导致逻辑上的循环或者不必要的信号发送。
	// 为了保持架构的统一性，我们依然通过 Manager 来走标准的销毁流程。
	// 但为了避免潜在的死锁（如果 Manager 的锁和 Hub 的通道产生交集），
	// 最好用一个异步的 Goroutine 去触发 Manager 的销毁逻辑。
	go h.Room.Manager.RemoveRoom(h.Room.ID)
}

func (h *Hub) handleHostTransferTimeout() {
	// 房主转移逻辑
	if len(h.Clients) > 0 {
		var newHost string
		// 按照加入顺序，找到最早加入且当前在线的玩家
		for _, id := range h.joinOrder {
			if _, online := h.Users[id]; online {
				newHost = id
				break
			}
		}

		if newHost != "" && newHost != h.Room.HostID {
			h.Room.HostID = newHost
			log.Printf("房间 [%s] 房主转移给 [%s]\n", h.Room.ID, newHost)
			h.broadcastRoomState()

			// 广播通知 (系统消息)
			h.broadcastSys(ActionHostChanged, fmt.Sprintf("原房主掉线，房主已自动转移给 %s", h.Users[newHost].Username))
		}
	}
}

// broadcastBytes 将字节流广播给所有客户端
// ⚠️ 警告：此方法专供【Hub.Run 内部】调用！
// 直接遍历 Clients map，不经过 Channel，避免向自身发送导致死锁。
func (h *Hub) broadcastBytes(message []byte) {
	for client := range h.Clients {
		if client.closed {
			continue
		}
		select {
		case client.Send <- message:
		default:
			h.closeClient(client)
		}
	}
}

// broadcastSys 广播系统消息
// ⚠️ 警告：专供【Hub.Run 内部】调用，避免向自身 Channel 发送导致死锁。
func (h *Hub) broadcastSys(action string, content string) {
	outBytes := BuildServerMessage(nil, action, content)
	h.broadcastBytes(outBytes)
}

// sendTo 发送单播消息
// ⚠️ 警告：专供【Hub.Run 内部】调用。
// 发送失败时会安全地调用 closeClient 清理死连接。
func (h *Hub) sendTo(client *Client, action string, content string) {
	outBytes := BuildServerMessage(nil, action, content)
	select {
	case client.Send <- outBytes:
	default:
		h.closeClient(client)
	}
}

func (h *Hub) handleDestroy() {
	// 1. 停止所有定时器，防止在销毁过程中触发
	h.roomRecycleTimer.Stop()
	h.hostTransferTimer.Stop()

	// 2. 通知并清理所有客户端
	for client := range h.Clients {
		if !client.closed {
			content := "房主已解散房间"
			if client.ID == h.Room.HostID {
				content = "您已解散房间"
			}
			// 这里是单播给特定玩家的系统消息
			h.sendTo(client, ActionRoomDismissed, content)
			h.closeClient(client)
		}
	}
}

// kickOutOldClient 踢掉已经存在的旧连接（顶号逻辑）
func (h *Hub) kickOutOldClient(userID string) {
	if oldClient, ok := h.Users[userID]; ok {
		if oldClient.closed {
			return
		}

		h.sendTo(oldClient, ActionKick, "您的账号在其他地方登录，您已被强制下线。")
		h.closeClient(oldClient)
	}
}

// broadcastRoomState 广播当前房间内的玩家列表
func (h *Hub) broadcastRoomState() {
	state := map[string]interface{}{
		"roomId":  h.Room.ID,
		"hostId":  h.Room.HostID,
		"players": make([]string, 0, len(h.Clients)),
	}

	players := state["players"].([]string)
	for client := range h.Clients {
		players = append(players, client.Username+" ("+client.ID+")")
	}
	state["players"] = players

	stateBytes, _ := json.Marshal(state)
	h.broadcastSys(ActionRoomState, string(stateBytes))
}
